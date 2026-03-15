package parser

import (
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// Operator precedence levels (higher number = binds tighter).
const (
	precNone       = 0
	precSemi       = 1
	precTemporal   = 2
	precImplies    = 3
	precOr         = 4
	precAnd        = 5
	precNot        = 6  // prefix only
	precCompare    = 7
	precNotEq      = 8
	precIfElse     = 9
	precColon      = 10
	precAdd        = 11
	precMul        = 12
	precDollar     = 13
	precOld        = 14 // prefix only
	precDot        = 15
)

// infixPrec returns the precedence of a binary operator token.
func infixPrec(tt lexer.TokenType) int {
	switch tt {
	// SEMI is not an infix operator; statement sequencing is handled by parseSequence
	case lexer.GLOBALLY, lexer.EVENTUALLY,
		lexer.WHENFIRST, lexer.WHENLAST, lexer.WHENNEXT, lexer.WHENPREV:
		return precTemporal
	case lexer.ARROW:
		return precImplies
	case lexer.IFF:
		return precImplies
	case lexer.OR:
		return precOr
	case lexer.AND:
		return precAnd
	case lexer.EQ, lexer.LE, lexer.LT, lexer.GE, lexer.GT, lexer.PTO:
		return precCompare
	case lexer.TILDAEQ:
		return precNotEq
	case lexer.IF:
		return precIfElse
	case lexer.COLON:
		return precColon
	case lexer.PLUS, lexer.MINUS:
		return precAdd
	case lexer.TIMES, lexer.DIV:
		return precMul
	case lexer.DOLLAR:
		return precDollar
	case lexer.DOT:
		return precDot
	case lexer.ISA:
		return precCompare
	default:
		return precNone
	}
}

// parseExpr parses an expression using Pratt/precedence climbing.
func (p *Parser) parseExpr(minPrec int) ast.Node {
	left := p.parsePrefix()

	for {
		prec := infixPrec(p.current.Type)
		if prec <= minPrec {
			break
		}
		left = p.parseInfix(left, prec)
	}
	return left
}

// parsePrefix handles prefix expressions and primary terms.
func (p *Parser) parsePrefix() ast.Node {
	tok := p.current

	switch tok.Type {
	case lexer.SYMBOL:
		return p.parseAtomOrApp()

	case lexer.VARIABLE:
		p.advance()
		v := ast.NewVariable(tok.Value, nil)
		p.setLoc(v, tok)
		// Check for sort annotation: V:type (only in certain contexts)
		// This is handled by the COLON infix operator at precColon
		return v

	case lexer.TRUE:
		p.advance()
		return p.setLoc(ast.NewAnd(), tok)

	case lexer.FALSE:
		p.advance()
		return p.setLoc(ast.NewOr(), tok)

	case lexer.THIS:
		p.advance()
		return p.setLoc(&ast.This{}, tok)

	case lexer.OLD:
		p.advance()
		body := p.parseExpr(precOld)
		return p.setLoc(ast.NewOld(body), tok)

	case lexer.TILDA:
		p.advance()
		body := p.parseExpr(precNot)
		return p.setLoc(ast.NewNot(body), tok)

	case lexer.GLOBALLY:
		p.advance()
		body := p.parseExpr(precTemporal)
		return p.setLoc(ast.NewGlobally(body), tok)

	case lexer.EVENTUALLY:
		p.advance()
		body := p.parseExpr(precTemporal)
		return p.setLoc(ast.NewEventually(body), tok)

	case lexer.FORALL:
		p.advance()
		return p.parseQuantifier(true, tok)

	case lexer.EXISTS:
		p.advance()
		return p.parseQuantifier(false, tok)

	case lexer.SOME:
		p.advance()
		return p.parseSomeExpr(tok)

	case lexer.DOLLAR:
		p.advance()
		return p.parseNamedBinder(tok)

	case lexer.LPAREN:
		p.advance()
		if p.at(lexer.DOLLAR) {
			// ( $name vars . fmla ) (args) — applied named binder
			return p.parseAppliedNamedBinder(tok)
		}
		expr := p.parseExpr(0)
		// Check for tuple: (a, b, ...)
		if p.match(lexer.COMMA) {
			elems := []ast.Node{expr}
			elems = append(elems, p.parseExpr(0))
			for p.match(lexer.COMMA) {
				elems = append(elems, p.parseExpr(0))
			}
			p.expect(lexer.RPAREN)
			return p.setLoc(ast.NewTuple(elems...), tok)
		}
		p.expect(lexer.RPAREN)
		return expr

	case lexer.NATIVEQUOTE:
		p.advance()
		nc := ast.NewNativeCode(tok.Value)
		return p.setLoc(nc, tok)

	case lexer.CARET:
		p.advance()
		inner := p.parseAtomOrApp()
		if app, ok := inner.(*ast.App); ok {
			return &ast.KeyArg{App: app}
		}
		// Convert atom to app for keyarg
		if atom, ok := inner.(*ast.Atom); ok {
			sym := ast.NewSymbol(atom.Rep, nil)
			app := ast.NewApp(sym, atom.Terms...)
			return &ast.KeyArg{App: app}
		}
		return inner

	case lexer.MINUS:
		// Unary minus: treat as 0 - expr
		p.advance()
		right := p.parseExpr(precMul)
		zero := ast.NewSymbol("0", nil)
		return p.setLoc(ast.NewAtom("-", zero, right), tok)

	default:
		p.errorf("unexpected token %s (%q)", tok.Type, tok.Value)
		p.advance()
		return ast.NewSymbol("?error?", nil)
	}
}

// parseAtomOrApp parses a SYMBOL possibly followed by (args).
func (p *Parser) parseAtomOrApp() ast.Node {
	tok := p.advance() // consume SYMBOL
	name := tok.Value

	// Check for function application: f(x, y)
	if p.match(lexer.LPAREN) {
		args := p.parseTermList()
		p.expect(lexer.RPAREN)
		a := ast.NewAtom(name, args...)
		p.setLoc(a, tok)
		return a
	}

	// Just a symbol
	sym := ast.NewSymbol(name, nil)
	p.setLoc(sym, tok)
	return sym
}

// parseInfix handles infix operators.
func (p *Parser) parseInfix(left ast.Node, prec int) ast.Node {
	tok := p.current

	switch tok.Type {
	case lexer.AND:
		p.advance()
		right := p.parseExpr(prec)
		return p.collectNary(left, right, &ast.And{}, lexer.AND, prec, tok)

	case lexer.OR:
		p.advance()
		right := p.parseExpr(prec)
		return p.collectNary(left, right, &ast.Or{}, lexer.OR, prec, tok)

	case lexer.ARROW:
		p.advance()
		right := p.parseExpr(prec - 1) // right-associative
		return p.setLoc(ast.NewImplies(left, right), tok)

	case lexer.IFF:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewIff(left, right), tok)

	case lexer.EQ:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewAtom("=", left, right), tok)

	case lexer.TILDAEQ:
		p.advance()
		right := p.parseExpr(prec)
		eq := ast.NewAtom("=", left, right)
		return p.setLoc(ast.NewNot(eq), tok)

	case lexer.LE:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewAtom("<=", left, right), tok)

	case lexer.LT:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewAtom("<", left, right), tok)

	case lexer.GE:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewAtom(">=", left, right), tok)

	case lexer.GT:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewAtom(">", left, right), tok)

	case lexer.PTO:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewAtom("*>", left, right), tok)

	case lexer.PLUS:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewAtom("+", left, right), tok)

	case lexer.MINUS:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewAtom("-", left, right), tok)

	case lexer.TIMES:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewAtom("*", left, right), tok)

	case lexer.DIV:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewAtom("/", left, right), tok)

	case lexer.DOT:
		p.advance()
		right := p.parsePrefix()
		return p.setLoc(ast.NewDot(left, right), tok)

	case lexer.COLON:
		p.advance()
		atype := p.parseAType()
		// Sort annotation: set the sort on the left node
		if v, ok := left.(*ast.Variable); ok {
			return v.Resort(atype)
		}
		if a, ok := left.(*ast.Atom); ok {
			a.ASort = atype
			return a
		}
		// For app or symbol, wrap in a typed expression
		if sym, ok := left.(*ast.Symbol); ok {
			sym.Sort = atype
			return sym
		}
		return left

	case lexer.IF:
		// term IF cond ELSE term (ternary)
		p.advance()
		cond := p.parseExpr(0)
		p.expect(lexer.ELSE)
		else_ := p.parseExpr(prec - 1)
		return p.setLoc(ast.NewIte(cond, left, else_), tok)

	case lexer.WHENNEXT:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewWhenOperator("", left, right), tok)

	case lexer.WHENPREV:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewWhenOperator("prev", left, right), tok)

	case lexer.WHENFIRST:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewWhenOperator("first", left, right), tok)

	case lexer.WHENLAST:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewWhenOperator("last", left, right), tok)

	case lexer.ISA:
		p.advance()
		atype := p.parseAType()
		return p.setLoc(&ast.Isa{Terms: []ast.Node{left, atype}}, tok)

	case lexer.DOLLAR:
		p.advance()
		name := p.expect(lexer.SYMBOL)
		// $name. fmla or $name$ fmla
		if p.match(lexer.DOT) || p.match(lexer.DOLLAR) {
			body := p.parseExpr(0)
			return p.setLoc(ast.NewNamedBinder(name.Value, nil, body), tok)
		}
		return left

	default:
		return left
	}
}

// collectNary collects a chain of the same n-ary operator.
func (p *Parser) collectNary(left, right ast.Node, proto ast.Node, opTok lexer.TokenType, prec int, tok lexer.Token) ast.Node {
	terms := []ast.Node{left, right}
	for p.at(opTok) {
		p.advance()
		terms = append(terms, p.parseExpr(prec))
	}
	switch proto.(type) {
	case *ast.And:
		return p.setLoc(ast.NewAnd(terms...), tok)
	case *ast.Or:
		return p.setLoc(ast.NewOr(terms...), tok)
	}
	return left
}

// parseQuantifier parses forall/exists after the keyword.
func (p *Parser) parseQuantifier(isForall bool, tok lexer.Token) ast.Node {
	var bounds []ast.Node

	if p.match(lexer.LPAREN) {
		// forall (X:t, Y:u) body
		bounds = p.parseSimpleVars()
		p.expect(lexer.RPAREN)
	} else {
		// forall X:t, Y:u . body
		bounds = p.parseSimpleVars()
		p.expect(lexer.DOT)
	}

	body := p.parseExpr(0)

	if isForall {
		return p.setLoc(ast.NewForall(bounds, body), tok)
	}
	return p.setLoc(ast.NewExists(bounds, body), tok)
}

// parseSomeExpr parses "some X:t. phi" expressions.
func (p *Parser) parseSomeExpr(tok lexer.Token) ast.Node {
	// Python's bounds grammar:
	//   bounds : params DOT        →  X:t, Y:u .
	//          | LPAREN lparams RPAREN  →  (x:t, y:u)
	var bounds []ast.Node
	if p.match(lexer.LPAREN) {
		// Parenthesized form: some(x:t, y:u) fmla
		bounds = p.parseDefArgs()
		p.expect(lexer.RPAREN)
	} else {
		// DOT-delimited form: some X:t, Y:u . fmla
		bounds = p.parseSimpleVars()
		p.expect(lexer.DOT)
	}
	fmla := p.parseExpr(0)

	if p.match(lexer.MINIMIZING) {
		idx := p.parseExpr(0)
		return p.setLoc(&ast.SomeMin{Params: bounds, Fmla: fmla, Index: idx}, tok)
	}
	if p.match(lexer.MAXIMIZING) {
		idx := p.parseExpr(0)
		return p.setLoc(&ast.SomeMax{Params: bounds, Fmla: fmla, Index: idx}, tok)
	}

	if len(bounds) == 1 {
		se := &ast.SomeExpr{Param: bounds[0], Fmla: fmla}
		if p.match(lexer.IN) {
			se.IfValue = p.parseExpr(0)
			if p.match(lexer.ELSE) {
				se.ElseVal = p.parseExpr(0)
			}
		}
		return p.setLoc(se, tok)
	}

	return p.setLoc(ast.NewSome(bounds, fmla), tok)
}

// parseNamedBinder parses "$name . body" or "$name $ body".
func (p *Parser) parseNamedBinder(tok lexer.Token) ast.Node {
	name := p.expect(lexer.SYMBOL)
	if p.match(lexer.DOT) || p.match(lexer.DOLLAR) {
		body := p.parseExpr(0)
		return p.setLoc(ast.NewNamedBinder(name.Value, nil, body), tok)
	}
	// Just a dollar + symbol without binding
	return p.setLoc(ast.NewSymbol("$"+name.Value, nil), tok)
}

// parseAppliedNamedBinder parses "($name vars . fmla)(args)" after the opening paren.
func (p *Parser) parseAppliedNamedBinder(tok lexer.Token) ast.Node {
	p.advance() // consume $
	name := p.expect(lexer.SYMBOL)
	bounds := p.parseSimpleVars()
	p.expect(lexer.DOT)
	body := p.parseExpr(0)
	p.expect(lexer.RPAREN)

	nb := ast.NewNamedBinder(name.Value, bounds, body)

	// Check for application
	if p.match(lexer.LPAREN) {
		args := p.parseTermList()
		p.expect(lexer.RPAREN)
		sym := ast.NewSymbol(name.Value, nil)
		app := ast.NewApp(sym, args...)
		_ = nb // The named binder defines the function
		return p.setLoc(app, tok)
	}

	return p.setLoc(nb, tok)
}
