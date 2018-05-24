package main

import (
	"fmt"
	"strconv"
)

type Assembler struct {
	program AnalyzedProgram
	cpool        ConstantPool
}

type Instruction struct {
	// Opcode is the operation code of this instruction
	Opcode

	// Constant pool entry index
	cpoolIndex int

	// address is the computed address (offset) of this instruction
	address int

	// labelFunc is a function that will be called once the address of this instruction is defined. Used in labels.
	labelFunc func(address int)
}

func (a *Assembler) AssembleProgram() {
	for _, v := range a.program.Nodes {
		a.assembleNode(v, nil)
	}

	// Compute addresses
	address := 0
	for _, method := range a.program.methods {
		for _, instr := range method.instructions {
			instr.address = address

			// Execute instruction function if any
			if instr.labelFunc != nil {
				instr.labelFunc(address)
			}

			// Labels do not alter the address. They're filtered out during encoding.
			if instr.Opcode != op_label {
				address++
			}
		}
	}
}

func (a *Assembler) assembleNode(node ASTNode, method *Method) {
	switch n := node.(type) {
	case *ASTTrigger:
		a.assembleTrigger(n)
	case *ASTFunc:
		a.assembleFunc(n)
	case *ASTBlockStatement:
		a.assembleBlock(n, method)
	case *ASTVarDeclaration:
		a.assembleVarDecl(n, method)
	case *ASTVarAssign:
		a.assembleVarAssign(n, method)
	case *ASTMethodExpr:
		a.assembleMethodExpr(n, method)
	case *ASTLiteralExpr:
		a.assembleLiteralExpr(n, method)
	case *ASTIfStmt:
		a.assembleIfStmt(n, method)
	case *ASTLogicalExpr:
		a.assembleLogicalExpr(n, method)
	case *ASTIdentifierExpr:
		a.assembleIdentifierExpr(n, method)
	default:
		panic(fmt.Sprintf("No function to walk node: %T", node))
	}
}

func (a *Assembler) assembleTrigger(n *ASTTrigger) {
	// Assemble the code belonging to this call
	a.assembleNode(n.statement, n.method)

	// Drop a return statement
	n.method.emitOp(op_return)
}

func (a *Assembler) assembleFunc(n *ASTFunc) {
	m := a.program.resolveMethod(n.name)

	// Assemble the parameters (take values from stack and assign to locals)
	for _, v := range m.arguments {
		m.emit(instr(op_setlocal, v.index))
	}

	a.assembleNode(n.body, m)
	m.emitOp(op_return)
}

func (a *Assembler) assembleBlock(n *ASTBlockStatement, m *Method) {
	for _, v := range n.statements {
		a.assembleNode(v, m)
	}
}

func (a *Assembler) assembleIfStmt(n *ASTIfStmt, m *Method) {
	a.assembleNode(n.condition, m)
	lblFalse := m.newLabel()
	lblEnd := m.newLabel()

	// JZ to lblFalse - absolute
	jzToLblfalse := m.emit(instr(op_jz, 0)) // Jump if false (0) to the false block
	lblFalse.labelFunc = func(address int) {
		jzToLblfalse.cpoolIndex = address
	}

	// Encode true block (jz jumps over this if expression is false)
	a.assembleNode(n.ifTrue, m)
	jmpToLblend := m.emit(instr(op_jmp, 0))
	lblEnd.labelFunc = func(address int) {
		jmpToLblend.cpoolIndex = address
	}

	// Encode false block (the true block jumps over this)
	m.emit(lblFalse)

	if n.ifFalse != nil {
		a.assembleNode(n.ifFalse, m)
	}

	// Add the end node - the true block jumps to this to avoid executing the false block.
	m.emit(lblEnd)
}

func (a *Assembler) assembleVarDecl(n *ASTVarDeclaration, m *Method) {
	// If the local var has any expression defined, assemble it
	if n.varValue != nil {
		a.assembleNode(n.varValue, m)
		m.emit(instr(op_setlocal, n.variable.index))
	}
}

func (a *Assembler) assembleVarAssign(n *ASTVarAssign, m *Method) {
	local := m.resolveVariable(n.varName)

	if local == nil {
		panic("undeclared variable " + n.varName)
	}

	// Assemble the node first so it resolves the type
	a.assembleNode(n.varValue, m)

	// Now verify that type against the variable type
	exprType := m.TypeOfNode(n.varValue)

	// Verify types
	if exprType != local.typ {
		panic("assigning wrong type to '" + local.typ.String() + " " + n.varName + "' (passed: " + exprType.String() + ")")
	}

	m.emit(instr(op_setlocal, local.index))
}

func (a *Assembler) assembleMethodExpr(n *ASTMethodExpr, m *Method) {
	// Assemble method parameters
	for i := range n.parameters {
		a.assembleNode(n.parameters[len(n.parameters) - i - 1], m)
	}

	if n.native != nil {
		m.emit(instr(op_nativecall, n.native.InternalId))
	} else {
		m.emit(instr(op_call, n.local.index))
	}
}

func (a *Assembler) assembleLogicalExpr(n *ASTLogicalExpr, m *Method) {
	a.assembleNode(n.left, m)
	a.assembleNode(n.right, m)

	switch n.comparator {
	case tokenEqual:
		m.emitOp(op_eq)
	case tokenPlus:
		m.emitOp(op_add)
	case tokenMinus:
		m.emitOp(op_sub)
	case tokenMultiply:
		m.emitOp(op_mul)
	case tokenDivide:
		m.emitOp(op_div)
	case tokenModulo:
		m.emitOp(op_mod)
	default:
		panic("unknown comparator node! " + strconv.Itoa(int(n.comparator)))
	}
}

func (a *Assembler) assembleIdentifierExpr(n *ASTIdentifierExpr, m *Method) {
	m.emit(instr(op_getlocal, n.resolved.index))
}

func (a *Assembler) assembleLiteralExpr(n *ASTLiteralExpr, m *Method) {
	if n.literalType == LiteralString {
		m.emit(instr(op_pushconst, a.cpool.getString(n.value.(string))))
	} else if n.literalType == LiteralInteger {
		m.emit(instr(op_pushconst, a.cpool.getInt(int(n.value.(int)))))
	} else if n.literalType == LiteralLong {
		m.emit(instr(op_pushconst, a.cpool.getLong(n.value.(int64))))
	} else if n.literalType == LiteralBoolean {
		if n.value == "true" {
			m.emit(instr(op_pushconst, a.cpool.getInt(int(1))))
		} else {
			m.emit(instr(op_pushconst, a.cpool.getInt(int(0))))
		}
	} else {
		panic(fmt.Sprintf("unknown literal type %d (value %s)", n.literalType, n.value))
	}
}

func (m *Method) emit(instruction *Instruction) *Instruction {
	m.instructions = append(m.instructions, instruction)
	return instruction
}

func (m *Method) emitOp(op Opcode) {
	instruction := instr(op, 0)
	m.instructions = append(m.instructions, instruction)
}

func instr(op Opcode, i int) *Instruction {
	return &Instruction{Opcode: op, cpoolIndex: i}
}

func (m *Method) newLabel() *Instruction {
	m.labelPtr++

	return &Instruction{
		Opcode: op_label,
	}
}
