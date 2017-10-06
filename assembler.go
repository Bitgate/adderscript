package main

import (
	"fmt"
	"strconv"
	"math"
)

type Assembler struct {
	methods       []*Method
	nativeMethods []*NativeMethod
	methodIndex   int
	cpool         ConstantPool
}

type ConstantPool struct {
	ints    []int
	longs   []int64
	strings []string
}

type Method struct {
	name         string
	index        int
	instructions []Instruction
	variables    []*LocalVariable

	lvtIndex int
}

type NativeMethod struct {
	name   string
	opcode int
}

type Instruction struct {
	opcode             // Instruction code
	i             int  // Index of int
	l             int  // Index of long
	s             int  // Index of string
	labelref      int  // Reference to label ID that will later on be translated to the address, then put into the 'i'.
	labelRelative bool // Whether the address of the label needs to be treated as relative, or as absolute value
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
	default:
		panic(fmt.Sprintf("No function to walk node: %T", node))
	}
}

func (a *Assembler) assembleTrigger(n *ASTTrigger) {
	//TODO
}

func (a *Assembler) defineProc(n *ASTProc) {
	method := a.resolveMethod(n.name)
	if method != nil {
		panic(fmt.Sprintf("redefining method: %s", n.name))
	}

	method = a.defineMethod(n.name)
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

func (a *Assembler) assembleVarDecl(n *ASTVarDeclaration, m *Method) {
	var vartype = VarTypeUnresolved

	if n.varType == "int" {
		vartype = VarTypeInt
	} else if n.varType == "long" {
		vartype = VarTypeLong
	} else if n.varType == "string" {
		vartype = VarTypeString
	}

	if vartype == VarTypeUnresolved {
		panic("Unresolved variable type: " + n.varType)
	}

	// See if this variable is already defined...
	if m.resolveVariable(n.varName) != nil {
		panic("Variable redeclared: " + n.varName)
	}

	m.defineVariable(n.varName, vartype)

	// If the local var has any expression defined, assemble it
	if n.varValue != nil {
		a.assembleNode(n.varValue, m)
	}
}

func (a *Assembler) assembleMethodExpr(n *ASTMethodExpr, m *Method) {
	// See if this is a native method first. Likelihood is much greater.
	nativeMethod := a.resolveNativeMethod(n.name)
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

func (a *Assembler) resolveNativeMethod(name string) *NativeMethod {
	for _, v := range a.nativeMethods {
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
		instructions: make([]Instruction, 512)[:0],
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

func (m *Method) emit(instruction Instruction) {
	m.instructions = append(m.instructions, instruction)
	println(fmt.Sprintf("%+v", instruction))
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

func instr(op opcode, i int, l int, s int) Instruction {
	return Instruction{opcode: op, i: i, l: l, s: s}
}
