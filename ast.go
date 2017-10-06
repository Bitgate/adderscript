package main

import "fmt"

type ASTType int

const (
	TypeTrigger     ASTType = iota
	TypeMethodCall
	TypeProc
	TypeBlockStmt
	TypeVarDecl
	TypeExprStmt
	TypeLiteral
	TypeIfStmt
	TypeLogicalExpr
	TypeIdentifierExpr // Can be either a var or a method ref
)

type ASTNode interface {
	Type() ASTType
}

func (t ASTType) Type() ASTType {
	return t
}

type ASTTrigger struct {
	ASTType
	trigger   string
	value     string
	statement ASTNode
}

func (t ASTTrigger) String() string {
	return fmt.Sprintf("ASTTrigger{on=%s, id=%s, statement=...}", t.trigger, t.value)
}

func newTrigger(trigger string, value string, statement ASTNode) *ASTTrigger {
	return &ASTTrigger{
		trigger:   trigger,
		value:     value,
		ASTType:   TypeTrigger,
		statement: statement,
	}
}

type ASTMethodExpr struct {
	ASTType
	name       string
	parameters []ASTNode
}

func (m ASTMethodExpr) String() string {
	return fmt.Sprintf("ASTMethodExpr{name=%s, params=%s}", m.name, m.parameters)
}

func newMethodExpr(name string, parameters ...ASTNode) *ASTMethodExpr {
	return &ASTMethodExpr{
		ASTType:    TypeMethodCall,
		name:       name,
		parameters: parameters,
	}
}

type ASTProc struct {
	ASTType
	name      string
	arguments []ProcArgument
	body      ASTNode
}

type ProcArgument struct {
	name    string
	argtype string
}

func (p ASTProc) String() string {
	return fmt.Sprintf("ASTProc{name=%s, args=%+v}", p.name, p.arguments)
}

func newProc(name string, body ASTNode, arguments ...ProcArgument) *ASTProc {
	return &ASTProc{
		ASTType:   TypeProc,
		name:      name,
		body:      body,
		arguments: arguments,
	}
}

type ASTExprStatement struct {
	ASTType
	expression ASTNode
}

func newStmt(expr ASTNode) *ASTExprStatement {
	return &ASTExprStatement{
		ASTType:    TypeExprStmt,
		expression: expr,
	}
}

type LiteralType int

const (
	LiteralNumber  LiteralType = iota
	LiteralString
	LiteralBoolean
)

type ASTLiteralExpr struct {
	ASTType
	literalType LiteralType
	value       string
}

func newLiteral(t LiteralType, value string) *ASTLiteralExpr {
	return &ASTLiteralExpr{
		ASTType:     TypeLiteral,
		literalType: t,
		value:       value,
	}
}

type ASTBlockStatement struct {
	ASTType
	statements []ASTNode
}

func newBlock(statements ...ASTNode) *ASTBlockStatement {
	return &ASTBlockStatement{
		ASTType:    TypeBlockStmt,
		statements: statements,
	}
}

type ASTVarDeclaration struct {
	ASTType
	varType  string
	varName  string
	varValue ASTNode // Optional. If non-nil, becomes an assign instruction too.
}

func newAssignment(varType, varName string, varValue ASTNode) *ASTVarDeclaration {
	return &ASTVarDeclaration{
		ASTType:  TypeVarDecl,
		varType:  varType,
		varName:  varName,
		varValue: varValue,
	}
}

type ASTIfStmt struct {
	ASTType
	condition ASTNode
	ifTrue    ASTNode
	ifFalse   ASTNode
}

func newIfStmt(condition ASTNode, ifTrue ASTNode, ifFalse ASTNode) *ASTIfStmt {
	return &ASTIfStmt{
		ASTType:   TypeIfStmt,
		condition: condition,
		ifTrue:    ifTrue,
		ifFalse:   ifFalse,
	}
}

type ASTLogicalExpr struct {
	ASTType
	left       ASTNode
	comparator tokenType
	right      ASTNode
}

func newLogicalExpr(left ASTNode, comparator tokenType, right ASTNode) *ASTLogicalExpr {
	return &ASTLogicalExpr{
		ASTType:    TypeLogicalExpr,
		left:       left,
		comparator: comparator,
		right:      right,
	}
}

type ASTIdentifierExpr struct {
	ASTType
	identifier string
}

func newIdentifier(identifier string) *ASTIdentifierExpr {
	return &ASTIdentifierExpr{
		ASTType:    TypeIdentifierExpr,
		identifier: identifier,
	}
}