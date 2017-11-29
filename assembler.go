package main

import (
	"fmt"
	"strconv"
	"math"
)

type Assembler struct {
	methods       []*Method
	triggers      []*Trigger
	nativeMethods map[string]*NativeMethod
	methodIndex   int
	triggerIndex int
	cpool         ConstantPool
}

type ConstantPool struct {
	ints    []int
	longs   []int64
	strings []string
}

type Trigger struct {
	name  string
	label *Instruction
	value uint64
}

type Method struct {
	name         string
	index        int
	instructions []*Instruction
	variables    []*LocalVariable
	arguments    []*LocalVariable

	lvtIndex int
	labelPtr int
}

type NativeMethod struct {
	name   string
	opcode int
}

type Instruction struct {
	// opcode is the operation code of this instruction
	opcode

	// i is the index of the int in the constant pool
	i int

	// l is the index of the long in the constant pool
	l int

	// s is the index of the string in the constant pool
	s int

	// address is the computed address (offset) of this instruction
	address int

	// labelFunc is a function that will be called once the address of this instruction is defined. Used in labels.
	labelFunc func(address int)
}

type LocalVariable struct {
	index int
	name  string
}

type VariableType int

const (
	VarTypeInt    VariableType = iota
	VarTypeLong
	VarTypeString

	VarTypeUnresolved VariableType = -1
)

func (a *Assembler) AssembleProgram(rootNodes []ASTNode) {
	// Hoist proc declarations
	for _, v := range rootNodes {
		if v.Type() == TypeProc {
			a.defineProc(v.(*ASTProc))
		}
	}

	for _, v := range rootNodes {
		a.assembleNode(v, nil)
	}

	// Compute addresses
	address := 0
	for _, method := range a.methods {
		for _, instr := range method.instructions {
			instr.address = address

			// Execute instruction function if any
			if instr.labelFunc != nil {
				instr.labelFunc(address)
			}

			// Labels do not alter the address. They're filtered out during encoding.
			if instr.opcode != op_label {
				address++
			}
		}
	}
}

func (a *Assembler) assembleNode(node ASTNode, method *Method) {
	switch n := node.(type) {
	case *ASTTrigger:
		a.assembleTrigger(n)
	case *ASTProc:
		a.assembleProc(n)
	case *ASTBlockStatement:
		a.assembleBlock(n, method)
	case *ASTVarDeclaration:
		a.assembleVarDecl(n, method)
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
	var trigger Trigger
	trigger.name = n.trigger

	// Verify that the filter value is a valid value.
	// For now, values are longs only. This is subject to change.
	parsed, err := strconv.ParseUint(n.value, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("cannot parse trigger value into long: %s", n.value))
	}

	method := a.defineMethod("@" + n.trigger + "@" + n.value + "@" + strconv.Itoa(a.triggerIndex))
	a.triggerIndex++

	label := method.newLabel()
	method.emit(label)
	trigger.value = parsed
	trigger.label = label

	a.triggers = append(a.triggers, &trigger)

	// Assemble the code belonging to this call
	a.assembleNode(n.statement, method)
}

func (a *Assembler) defineProc(n *ASTProc) {
	method := a.resolveMethod(n.name)
	if method != nil {
		panic(fmt.Sprintf("redefining method: %s", n.name))
	}

	method = a.defineMethod(n.name)

	// Define method parameters as local variables
	for _, arg := range n.arguments {
		var vt = resolveVartype(arg.argtype)

		if vt == VarTypeUnresolved {
			panic(fmt.Sprintf("unresolved variable type %s", arg.argtype))
		}

		lv := method.defineVariable(arg.name, vt)
		method.arguments = append(method.arguments, lv)
	}
}

func (a *Assembler) assembleProc(n *ASTProc) {
	m := a.resolveMethod(n.name)
	a.assembleNode(n.body, m)
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
	jzToLblfalse := m.emit(instr(op_jz, 0, 0, 0)) // Jump if false (0) to the false block
	lblFalse.labelFunc = func(address int) {
		jzToLblfalse.i = address
	}

	// Encode true block (jz jumps over this if expression is false)
	a.assembleNode(n.ifTrue, m)
	jmpToLblend := m.emit(instr(op_jmp, 0, 0, 0))
	lblEnd.labelFunc = func(address int) {
		jmpToLblend.i = address
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
	var vartype = resolveVartype(n.varType)

	if vartype == VarTypeUnresolved {
		panic("Unresolved variable type: " + n.varType)
	}

	// See if this variable is already defined...
	if m.resolveVariable(n.varName) != nil {
		panic("Variable redeclared: " + n.varName)
	}

	local := m.defineVariable(n.varName, vartype)

	// If the local var has any expression defined, assemble it
	if n.varValue != nil {
		a.assembleNode(n.varValue, m)
		// todo all types
		m.emit(instr(op_setivar, local.index, 0, 0))
	}
}

func resolveVartype(vartype string) VariableType {
	switch vartype {
	case "int":
		return VarTypeInt;
	case "long":
		return VarTypeLong;
	case "string":
		return VarTypeString;
	default:
		return VarTypeUnresolved
	}
}

func (a *Assembler) assembleMethodExpr(n *ASTMethodExpr, m *Method) {
	// See if this is a native method first. Likelihood is much greater.
	nativeMethod := a.nativeMethods[n.name]
	var localMethod *Method

	if nativeMethod == nil {
		localMethod = a.resolveMethod(n.name)

		// Still not found? Panic.
		if localMethod == nil {
			panic("Cannot resolve local or native method: " + n.name)
		}
	}

	// Assemble method parameters
	for _, v := range n.parameters {
		a.assembleNode(v, m)
	}

	if nativeMethod != nil {
		m.emit(instr(op_nativecall, nativeMethod.opcode, 0, 0))
	} else {
		m.emit(instr(op_call, localMethod.index, 0, 0))
	}
}

func (a *Assembler) assembleLogicalExpr(n *ASTLogicalExpr, m *Method) {
	a.assembleNode(n.left, m)
	a.assembleNode(n.right, m)

	switch n.comparator {
	case tokenEqual:
		m.emitOp(op_eq)
	default:
		panic("unknown comparator node! " + strconv.Itoa(int(n.comparator)))
	}
}

func (a *Assembler) assembleIdentifierExpr(n *ASTIdentifierExpr, m *Method) {
	//TODO type checks
	local := m.resolveVariable(n.identifier)
	if local == nil {
		panic("undefined variable: " + n.identifier)
	}

	m.emit(instr(op_getivar, local.index, 0, 0))
}

func (a *Assembler) assembleLiteralExpr(n *ASTLiteralExpr, m *Method) {
	if n.literalType == LiteralString {
		v, _ := strconv.Unquote(n.value)
		m.emit(instr(op_strpush, 0, 0, a.cpool.getString(v)))
	} else if n.literalType == LiteralNumber {
		v, e := strconv.ParseInt(n.value, 10, 64)
		if e != nil {
			panic("error parsing int value from string: " + n.value)
		}

		if v > math.MaxInt32 || v < math.MinInt32 {
			m.emit(instr(op_lpush, 0, a.cpool.getLong(v), 0))
		} else {
			m.emit(instr(op_ipush, a.cpool.getInt(int(v)), 0, 0))
		}
	} else if n.literalType == LiteralBoolean {
		if n.value == "true" {
			m.emit(instr(op_ipush, a.cpool.getInt(int(1)), 0, 0))
		} else {
			m.emit(instr(op_ipush, a.cpool.getInt(int(0)), 0, 0))
		}
	} else {
		panic(fmt.Sprintf("unknown literal type %d (value %s)", n.literalType, n.value))
	}
}

func (a *Assembler) resolveMethod(name string) *Method {
	for _, v := range a.methods {
		if v.name == name {
			return v
		}
	}

	return nil
}

func (a *Assembler) defineMethod(name string) *Method {
	index := a.methodIndex
	a.methodIndex++

	method := &Method{
		name:         name,
		index:        index,
		instructions: make([]*Instruction, 512)[:0],
		variables:    make([]*LocalVariable, 4)[:0],
	}

	a.methods = append(a.methods, method)
	return method
}

func (m *Method) resolveVariable(name string) *LocalVariable {
	for _, v := range m.variables {
		if v.name == name {
			return v
		}
	}

	return nil
}

func (a *Method) defineVariable(name string, t VariableType) *LocalVariable {
	index := a.lvtIndex
	a.lvtIndex++

	local := &LocalVariable{
		name:  name,
		index: index,
	}

	a.variables = append(a.variables, local)
	return local
}

func (m *Method) emit(instruction *Instruction) *Instruction {
	m.instructions = append(m.instructions, instruction)
	return instruction
}

func (m *Method) emitOp(op opcode) {
	instruction := instr(op, 0, 0, 0)
	m.instructions = append(m.instructions, instruction)
}

func (c *ConstantPool) getInt(i int) int {
	for k, v := range c.ints {
		if v == i {
			return k
		}
	}

	c.ints = append(c.ints, i)
	return len(c.ints) - 1
}

func (c *ConstantPool) getLong(i int64) int {
	for k, v := range c.longs {
		if v == i {
			return k
		}
	}

	c.longs = append(c.longs, i)
	return len(c.longs) - 1
}

func (c *ConstantPool) getString(s string) int {
	for k, v := range c.strings {
		if v == s {
			return k
		}
	}

	c.strings = append(c.strings, s)
	return len(c.strings) - 1
}

func instr(op opcode, i int, l int, s int) *Instruction {
	return &Instruction{opcode: op, i: i, l: l, s: s}
}

func (m *Method) newLabel() *Instruction {
	m.labelPtr++

	return &Instruction{
		opcode: op_label,
	}
}

func (a *Assembler) resolveNativeMethod(opcode int) *NativeMethod {
	for _, v := range a.nativeMethods {
		if v.opcode == opcode {
			return v
		}
	}

	return nil
}
