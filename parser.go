package main

import (
	"strings"
	"fmt"
	"strconv"
	"math"
)

var eofToken = token{tokenType: tokenEOF}

type parser struct {
	source string
	tokens []token
	pos    int
}

func (p *parser) peek(ahead int) token {
	if p.pos+ahead >= len(p.tokens) {
		return eofToken
	}

	return p.tokens[p.pos+ahead]
}

func (p *parser) next() token {
	if p.pos >= len(p.tokens) {
		return eofToken
	}

	t := p.tokens[p.pos]
	p.pos++
	return t
}

func (p *parser) rewind() {
	if p.pos > 0 {
		p.pos--
	}
}

func (p *parser) unexpected(t token, expected ...string) {
	panic(fmt.Sprintf("unexpected '%s' (%d), expected one of: %s\n\n%s", t.value, t.tokenType, strings.Join(expected, ", "), p.generateErrorIndicator(t)))
}

func (p *parser) generateErrorIndicator(t token) string {
	sourcelen := len(p.source)
	lineStart := 0
	lineEnd := sourcelen
	lineNumber := 1

	// Find line start..
	for i := 0; i < sourcelen && i < t.from; i++ {
		if p.source[i] == '\n' && i != sourcelen-1 {
			lineStart = i + 1
			lineNumber++
		}
	}

	// Find line end..
	for i := lineStart + 1; i < sourcelen; i++ {
		if p.source[i] == '\n' {
			lineEnd = i - 1
			break
		}
	}

	// Create ^^-indicator..
	col := t.from - lineStart
	lineNumberStr := strconv.Itoa(lineNumber) + ": "
	indicator := strings.Repeat(" ", len(lineNumberStr)+col) + strings.Repeat("^", t.to-t.from)

	return fmt.Sprintf("%s%s\n%s", lineNumberStr, p.source[lineStart:lineEnd], indicator)
}

func (p *parser) unexpect(t tokenType, name string, expected ...string) {
	if p.peek(0).tokenType == t {
		panic(fmt.Sprintf("unexpected %s, expected one of: %s\n\n%s", name, strings.Join(expected, ", "), p.generateErrorIndicator(p.peek(0))))
	}
}

func (p *parser) expect(t tokenType, name string) {
	if p.peek(0).tokenType != t {
		panic(fmt.Sprintf("unexpected %s, expected %s\n\n%s", p.peek(0).value, name, p.generateErrorIndicator(p.peek(0))))
		panic("unexpected '" + p.peek(0).value + "', expected '" + name + "'")
	}
}

func (p *parser) expectConsume(t tokenType, name string) token {
	p.expect(t, name)
	return p.next()
}

func (p *parser) parseTopLevelDecl() ASTNode {
	t := p.next()
	if t.tokenType == tokenOn {
		p.rewind()
		return p.parseTrigger()
	} else if t.tokenType == tokenFunc {
		p.rewind()
		return p.parseFunc()
	} else {
		p.unexpected(t, "on", "func")
	}

	return nil
}

func (p *parser) parseTrigger() ASTNode {
	p.expectConsume(tokenOn, "on")
	identifier := p.expectConsume(tokenIdentifier, "identifier")
	p.expectConsume(tokenLParen, "(")
	value := p.expectConsume(tokenInteger, "integer")
	p.expectConsume(tokenRParen, ")")
	stmt := p.parseStatement()
	return newTrigger(identifier.value, value.value, stmt)
}

func (p *parser) parseFunc() ASTNode {
	p.expectConsume(tokenFunc, "func")
	name := p.expectConsume(tokenIdentifier, "function name")
	p.expectConsume(tokenLParen, "'('")

	// Parse argument list..
	arguments := []FuncArgument{}
	needsComma := false
	for {
		peek := p.peek(0)
		if peek.tokenType == tokenRParen {
			break
		} else if peek.tokenType == tokenComma {
			if needsComma {
				p.expectConsume(tokenComma, "','")
				needsComma = false
			} else {
				p.unexpect(tokenComma, "','")
			}
		} else {
			argType := p.expectConsume(tokenIdentifier, "argument type")
			argName := p.expectConsume(tokenIdentifier, "argument name")

			arguments = append(arguments, FuncArgument{argName.value, argType.value})
			needsComma = true
		}
	}

	p.expectConsume(tokenRParen, "')")
	body := p.parseStatement()
	return newFunc(name.value, body, arguments...)
}

func (p *parser) parseStatement() ASTNode {
	switch p.peek(0).tokenType {
	case tokenIdentifier:
		peek := p.peek(1)
		if peek.tokenType == tokenLParen {
			return p.parseMethodCall()
		} else if peek.tokenType == tokenAssign {
			return p.parseVarAssign()
		} else {
			return p.parseVarDecl()
		}
	case tokenLBrack:
		return p.parseBlockStatement()
	case tokenIf:
		return p.parseIfStmt()
	default:
		p.unexpected(p.peek(0), "method call", "variable declaration")
	}

	return nil
}

func (p *parser) parseMethodCall() ASTNode {
	method := p.parseMethodExpr()
	p.expectConsume(tokenSemicolon, "';'")
	return method
}

func (p *parser) parseMethodExpr() ASTNode {
	identifier := p.expectConsume(tokenIdentifier, "method identifier")
	p.expectConsume(tokenLParen, "(")

	arguments := []ASTNode{}
	expectComma := false

	for {
		peek := p.peek(0)

		if peek.tokenType == tokenRParen {
			break
		} else if peek.tokenType == tokenComma {
			if expectComma {
				p.expectConsume(tokenComma, ",")
				expectComma = false
			} else {
				p.unexpected(peek, "expression", "')'")
			}
		} else {
			expr := p.parseExpression()
			arguments = append(arguments, expr)
			expectComma = true
		}
	}

	p.expectConsume(tokenRParen, ")")
	return newMethodExpr(identifier.value, arguments...)
}

func (p *parser) parseIfStmt() ASTNode {
	p.expectConsume(tokenIf, "if")
	p.expectConsume(tokenLParen, "'('")
	condition := p.parseExpression()
	p.expectConsume(tokenRParen, "')'")

	ifTrue := p.parseStatement()

	var ifFalse ASTNode

	if p.peek(0).tokenType == tokenElse {
		p.expectConsume(tokenElse, "else")
		ifFalse = p.parseStatement()
	}

	return newIfStmt(condition, ifTrue, ifFalse)
}

func (p *parser) parseVarDecl() ASTNode {
	varType := p.expectConsume(tokenIdentifier, "variable type")
	varName := p.expectConsume(tokenIdentifier, "variable name")

	var varValue ASTNode
	if p.peek(0).tokenType == tokenAssign {
		p.expectConsume(tokenAssign, "=")
		varValue = p.parseExpression()
	}

	p.expectConsume(tokenSemicolon, "';'")
	return newAssignment(varType.value, varName.value, varValue)
}

func (p *parser) parseVarAssign() ASTNode {
	varName := p.expectConsume(tokenIdentifier, "variable name")
	p.expectConsume(tokenAssign, "'='")
	varValue := p.parseExpression()
	p.expectConsume(tokenSemicolon, "';'")

	return newVarAssign(varName.value, varValue)
}

func (p *parser) parseBlockStatement() ASTNode {
	p.expectConsume(tokenLBrack, "'{'")

	statements := []ASTNode{}
	for {
		peek := p.peek(0)
		if peek.tokenType == tokenRBrack {
			break
		}

		statements = append(statements, p.parseStatement())
	}

	p.expectConsume(tokenRBrack, "'}'")
	return newBlock(statements...)
}

func (p *parser) parseExpression() ASTNode {
	return p.parseLogicalExpression()
}

func (p *parser) parseLogicalExpression() ASTNode {
	left := p.parseEquality()
	for isLogicalOperator(p.peek(0).tokenType) {
		operator := p.next()
		right := p.parseEquality()
		left = newLogicalExpr(left, operator.tokenType, right)
	}

	return left
}

func (p *parser) parseEquality() ASTNode {
	left := p.parseRelational()
	for isEqualityOperator(p.peek(0).tokenType) {
		operator := p.next()
		right := p.parseRelational()
		left = newLogicalExpr(left, operator.tokenType, right)
	}

	return left
}

func (p *parser) parseRelational() ASTNode {
	left := p.parseAddSubtract()
	for isRelationalOperator(p.peek(0).tokenType) {
		operator := p.next()
		right := p.parseAddSubtract()
		left = newLogicalExpr(left, operator.tokenType, right)
	}

	return left
}

func (p *parser) parseAddSubtract() ASTNode {
	left := p.parseMulDivide()

	for isAddOrSubtract(p.peek(0).tokenType) {
		operator := p.next()
		right := p.parseMulDivide()
		left = newLogicalExpr(left, operator.tokenType, right)
	}

	return left
}

func (p *parser) parseMulDivide() ASTNode {
	left := p.parseTerminalExpression()

	for isMultiplyOrDivide(p.peek(0).tokenType) {
		operator := p.next()
		right := p.parseTerminalExpression()
		left = newLogicalExpr(left, operator.tokenType, right)
	}

	return left
}

func isLogicalOperator(t tokenType) bool {
	return false
}

func isEqualityOperator(t tokenType) bool {
	return t == tokenEqual || t == tokenNotEqual
}

func isRelationalOperator(t tokenType) bool {
	return t == tokenLessThan || t == tokenLessOrEqual || t == tokenGreaterThan || t == tokenGreaterOrEqual
}

func isAddOrSubtract(t tokenType) bool {
	return t == tokenPlus || t == tokenMinus
}

func isMultiplyOrDivide(t tokenType) bool {
	return t == tokenMultiply || t == tokenDivide
}

func (p *parser) parseTerminalExpression() ASTNode {
	peek := p.peek(0)
	switch peek.tokenType {
	case tokenInteger:
		tok := p.next()
		v, e := strconv.ParseInt(tok.value, 10, 64)
		if e != nil {
			panic("error parsing int value from string: " + tok.value)
		}

		if v > math.MaxInt32 || v < math.MinInt32 {
			return newLiteral(LiteralLong, int64(v))
		} else {
			return newLiteral(LiteralInteger, int(v))
		}
	case tokenString:
		v, _ := strconv.Unquote(p.next().value)
		return newLiteral(LiteralString, v)
	case tokenBool:
		value := p.next().value == "true"
		return newLiteral(LiteralBoolean, value)
	case tokenIdentifier:
		if p.peek(1).tokenType == tokenLParen {
			return p.parseMethodExpr()
		} else {
			return newIdentifier(p.next().value)
		}
	case tokenLParen: // Parenthesized expression
		p.expectConsume(tokenLParen, "'('")
		expr := p.parseExpression()
		p.expectConsume(tokenRParen, "')'")
		return expr
	default:
		p.unexpected(peek, "integer", "string", "'('")
	}

	return nil
}

func (p *parser) run() []ASTNode {
	nodes := []ASTNode{}

	for {
		peek := p.peek(0)
		if peek.tokenType == tokenEOF {
			break
		}

		node := p.parseTopLevelDecl()
		nodes = append(nodes, node)
	}

	return nodes
}

func Parse(source string, tokens []token) []ASTNode {
	p := parser{
		tokens: tokens,
		source: source,
	}

	return p.run()
}
