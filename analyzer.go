package main

import (
	"strings"
	"fmt"
	"strconv"
)

type AnalyzedProgram struct {
	Nodes []ASTNode
	methods      []*Method
	triggers     []*Trigger
	runtime      *AdderRuntime
	methodIndex  int
	triggerIndex int
}

type Trigger struct {
	name       string
	definition *RuntimeListener
	label      *Instruction
	values      []interface{}
}

type Method struct {
	name         string
	index        int
	instructions []*Instruction
	variables    []*LocalVariable
	arguments    []*LocalVariable
	entry        *Instruction

	lvtIndex int
	labelPtr int
}

type LocalVariable struct {
	index int
	name  string
	typ VariableType
}

type VariableType struct {
	builtin bool
	keyword string
	native string
}

var (
	VarTypeInt = VariableType{builtin: true, keyword: "int"}
	VarTypeLong = VariableType{builtin: true, keyword: "long"}
	VarTypeString = VariableType{builtin: true, keyword: "string"}
	VarTypeBool = VariableType{builtin: true, keyword: "bool"}
	VarTypeVoid = VariableType{builtin: true, keyword: "void"}    // Not a variable type, but defined to be used with method types

	VarTypeUnresolved = VariableType{builtin: true, keyword: "MISSING_TYPE"}
)

func ProcessAndAnalyzeProgram(runtime *AdderRuntime, rootNodes []ASTNode) AnalyzedProgram {
	program := AnalyzedProgram{runtime: runtime, Nodes: rootNodes}

	// Hoist proc declarations
	for _, v := range rootNodes {
		if v.Type() == TypeProc {
			program.defineProc(v.(*ASTProc))
		}
	}

	for _, v := range rootNodes {
		program.analyzeNode(v, nil)
	}

	return program
}

func (p *AnalyzedProgram) analyzeNode(node ASTNode, method *Method) {
	switch n := node.(type) {
	case *ASTTrigger:
		p.analyzeTrigger(n)
	case *ASTProc:
		p.analyzeProc(n)
	case *ASTBlockStatement:
		p.analyzeBlock(n, method)
	case *ASTVarDeclaration:
		p.analyzeVarDecl(n, method)
	case *ASTMethodExpr:
		p.analyzeMethodExpr(n, method)
	case *ASTLiteralExpr:
		p.analyzeLiteralExpr(n, method)
	case *ASTIfStmt:
		p.analyzeIfStatement(n, method)
	case *ASTLogicalExpr:
		p.analyzeLogicalExpr(n, method)
	case *ASTIdentifierExpr:
		p.analyzeIdentifierExpr(n, method)
	default:
		panic(fmt.Sprintf("No function to walk node: %T", node))
	}
}

func (a *AnalyzedProgram) analyzeTrigger(n *ASTTrigger) {
	var trigger Trigger
	trigger.name = n.trigger

	// Resolve the trigger uid
	listener := a.runtime.FindListener(trigger.name)
	if listener == nil {
		panic(fmt.Errorf("unknown trigger %s, not defined in runtime", trigger.name))
	}

	trigger.definition = listener
	n.entry = &trigger

	// Verify that the filter value is a valid value.
	// For now, values are longs only. This is subject to change.
	parsed, err := strconv.ParseUint(n.value, 10, 64)
	if err != nil {
		panic(fmt.Errorf("cannot parse trigger value into long: %s", n.value))
	}

	n.method = a.defineMethod("@" + n.trigger + "@" + n.value + "@" + strconv.Itoa(a.triggerIndex))
	a.triggerIndex++

	trigger.values = []interface{} {int64(parsed)} // TODO All value types here.

	a.triggers = append(a.triggers, &trigger)

	// Assemble the code belonging to this call
	a.analyzeNode(n.statement, n.method)
}

func (a *AnalyzedProgram) analyzeProc(n *ASTProc) {
	m := a.resolveMethod(n.name)
	a.analyzeNode(n.body, m)
}

func (a *AnalyzedProgram) analyzeBlock(n *ASTBlockStatement, m *Method) {
	for _, v := range n.statements {
		a.analyzeNode(v, m)
	}
}

func (a *AnalyzedProgram) analyzeIfStatement(n *ASTIfStmt, m *Method) {
	a.analyzeNode(n.condition, m)

	a.analyzeNode(n.ifTrue, m)

	if n.ifFalse != nil {
		a.analyzeNode(n.ifFalse, m)
	}
}

func (a *AnalyzedProgram) analyzeVarDecl(n *ASTVarDeclaration, m *Method) {
	var vartype = ResolveVarType(n.varType)

	if vartype == VarTypeUnresolved {
		panic("unresolved variable type: " + n.varType)
	}

	// See if this variable is already defined...
	if m.resolveVariable(n.varName) != nil {
		panic("variable redeclared: " + n.varName)
	}

	n.variable = m.defineVariable(n.varName, vartype)

	// If the local var has any expression defined, assemble it
	if n.varValue != nil {
		// Assemble the node first so it resolves the type
		a.analyzeNode(n.varValue, m)

		// Now verify that type against the variable type
		exprType := m.TypeOfNode(n.varValue)

		// Verify types
		if exprType != vartype {
			panic("cannot assign value of type '" + exprType.String() + "' to '" + n.varType + " " + n.varName + "'")
		}
	}
}

func (a *AnalyzedProgram) analyzeMethodExpr(n *ASTMethodExpr, m *Method) {
	// Form list of argument types
	var types []VariableType
	for _, v := range n.parameters {
		types = append(types, m.TypeOfNode(v))
	}

	// See if this is a native method first. Likelihood is much greater.
	nativeMethod := a.runtime.FindFunctionWithArguments(n.name, types...)
	var localMethod *Method

	if nativeMethod == nil {
		localMethod = a.resolveMethod(n.name)

		// Still not found? Panic.
		if localMethod == nil {
			params := "(" + TypeListToString(", ", types...) + ")"
			panic("Cannot resolve local or native method: " + n.name + params)
		}
	}

	// Analyze method parameters
	for i := range n.parameters {
		a.analyzeNode(n.parameters[len(n.parameters) - i - 1], m)
	}

	if nativeMethod != nil {
		n.native = nativeMethod
	} else {
		n.local = localMethod
	}
}

func (a *AnalyzedProgram) analyzeLogicalExpr(n *ASTLogicalExpr, m *Method) {
	a.analyzeNode(n.left, m)
	a.analyzeNode(n.right, m)
}

func (a *AnalyzedProgram) analyzeIdentifierExpr(n *ASTIdentifierExpr, m *Method) {
	//TODO type checks
	n.resolved = m.resolveVariable(n.identifier)
	if n.resolved == nil {
		panic("undefined variable: " + n.identifier)
	}
}

func (a *AnalyzedProgram) analyzeLiteralExpr(n *ASTLiteralExpr, m *Method) {
	if n.literalType == LiteralString {

	} else if n.literalType == LiteralInteger {

	} else if n.literalType == LiteralLong {

	} else if n.literalType == LiteralBoolean {

	} else {
		panic(fmt.Sprintf("unknown literal type %d (value %s)", n.literalType, n.value))
	}
}

func (a AnalyzedProgram) resolveMethod(name string) *Method {
	for _, v := range a.methods {
		if v.name == name {
			return v
		}
	}

	return nil
}

func (p *AnalyzedProgram) defineMethod(name string) *Method {
	index := p.methodIndex
	p.methodIndex++

	method := &Method{
		name:         name,
		index:        index,
		instructions: make([]*Instruction, 512)[:0],
		variables:    make([]*LocalVariable, 4)[:0],
	}

	// Define entry point, drop a label.
	label := method.newLabel()
	method.emit(label)
	method.entry = label

	p.methods = append(p.methods, method)
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
		typ: t,
	}

	a.variables = append(a.variables, local)
	return local
}


func (p *AnalyzedProgram) defineProc(n *ASTProc) {
	method := p.resolveMethod(n.name)
	if method != nil {
		panic(fmt.Sprintf("redefining method: %s", n.name))
	}

	method = p.defineMethod(n.name)

	// Define method parameters as local variables
	for _, arg := range n.arguments {
		var vt = ResolveVarType(arg.argtype)

		if vt == VarTypeUnresolved {
			panic(fmt.Sprintf("unresolved variable type %s", arg.argtype))
		}

		lv := method.defineVariable(arg.name, vt)
		method.arguments = append(method.arguments, lv)
	}
}

func ResolveVarType(varType string) VariableType {
	switch varType {
	case "int":
		return VarTypeInt
	case "long":
		return VarTypeLong
	case "string":
		return VarTypeString
	case "bool":
		return VarTypeBool
	default:
		if strings.HasPrefix(varType, "native<") && strings.HasSuffix(varType, ">") {
			contents := strings.Replace(strings.Replace(varType, ">", "", -1), "native<", "", -1)
			return VariableType{builtin: false, keyword:"native", native:contents}
		}
		return VarTypeUnresolved
	}
}

func ResolveType(typ string) VariableType {
	switch typ {
	case "void":
		return VarTypeVoid
	default:
		return ResolveVarType(typ)
	}
}

func (m *Method) TypeOfNode(node ASTNode) VariableType {
	switch t := node.(type) {
	case *ASTLiteralExpr:
		if t.literalType == LiteralInteger {
			return VarTypeInt
		} else if t.literalType == LiteralLong {
			return VarTypeLong
		} else if t.literalType == LiteralString {
			return VarTypeString
		} else if t.literalType == LiteralBoolean {
			return VarTypeBool
		} else {
			panic(fmt.Errorf("cannot resolve type from LiteralExpr, unknown literal type %d", t.literalType))
		}
	case *ASTMethodExpr: {
		if t.local != nil {
			panic("local method return types not supported")
		} else if t.native != nil {
			return t.native.ReturnType
		} else {
			panic("no resolved local/native func: " + t.name)
		}
	}
	case *ASTIdentifierExpr: {
		// TODO do this a bit nicer
		resolved := m.resolveVariable(t.identifier)
		if resolved == nil {
			panic("cannot resolve variable " + t.identifier)
		}
		return resolved.typ
	}
	}

	panic(fmt.Sprintf("cannot resolve type of node: %T", node))
}
