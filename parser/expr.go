package parser

import (
	"github.com/glycerine/goivy/ast"
	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/lexer"
)

// composeAtomsExpr flattens dotted name expressions, matching Python's
// compose_atoms(pr, atom) in ivy_ast.py:1884-1891.
//
// When both sides are Atom/Symbol (named nodes), it joins them:
//
//	Atom("ref",[]) DOT Atom("evs",[T]) → Atom("ref.evs", [T])
//
// When the left side is Old, it composes inside the Old:
//
//	Old(Atom("x")) DOT Atom("y") → Old(Atom("x.y"))
//
// Otherwise (e.g., Variable on left), falls back to a Dot node
// (Python creates MethodCall in this case).
func composeAtomsExpr(left, right ast.Node) ast.Node {
	// Python: elif isinstance(p[1], Old):
	//             t = compose_atoms(p[1].args[0], p[3]); p[0] = Old(t)
	if old, ok := left.(*ast.Old); ok {
		inner := composeAtomsExpr(old.Body, right)
		return ast.NewOld(inner)
	}

	leftName, leftArgs := extractAtomNameAndArgs(left)
	rightName, rightArgs := extractAtomNameAndArgs(right)

	// Python: if isinstance(p[1], (Atom, App)):
	//             p[0] = compose_atoms(p[1], p[3])
	if leftName != "" && rightName != "" {
		composedName := iu.ComposeNames(leftName, rightName)
		allArgs := make([]ast.Node, 0, len(leftArgs)+len(rightArgs))
		allArgs = append(allArgs, leftArgs...)
		allArgs = append(allArgs, rightArgs...)
		return ast.NewAtom(composedName, allArgs...)
	}

	// Fallback: MethodCall in Python; Dot in Go
	return ast.NewDot(left, right)
}

// extractAtomNameAndArgs extracts the name string and argument list from
// an Atom or Symbol node. Returns ("", nil) for other node types.
func extractAtomNameAndArgs(n ast.Node) (string, []ast.Node) {
	switch v := n.(type) {
	case *ast.Atom:
		return v.Rep, v.Terms
	case *ast.Symbol:
		return v.Rep, nil
	default:
		return "", nil
	}
}

// Operator precedence levels (higher number = binds tighter).
const (
	precNone     = 0
	precSemi     = 1
	precTemporal = 2
	precImplies  = 3
	precOr       = 4
	precAnd      = 5
	precNot      = 6 // prefix only
	precCompare  = 7
	precNotEq    = 8
	precIfElse   = 9
	precColon    = 10
	precAdd      = 11
	precMul      = 12
	precDollar   = 13
	precOld      = 14 // prefix only
	precDot      = 15
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

// ParseExpr parses a single expression from the current token stream using
// Pratt/precedence climbing. This is the public entry point for parsing
// formula/term strings. Pass minPrec=0 to parse a complete expression.
func (p *Parser) ParseExpr(minPrec int) ast.Node {
	return p.parseExpr(minPrec)
}

// parseExpr parses an expression using Pratt/precedence climbing.
func (p *Parser) parseExpr(minPrec int) ast.Node {
	left := p.parsePrefix()

	for {
		prec := infixPrec(p.current.Type)
		if prec <= minPrec {
			break
		}
		// Guard against infinite loops: if parseInfix doesn't consume
		// any tokens, break to avoid looping forever.
		savedPos := p.current
		left = p.parseInfix(left, prec)
		if p.current == savedPos {
			break
		}
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
		return p.setLoc(ast.NewApp(ast.NewSymbol("-", nil), zero, right), tok)

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
		right := p.parseExpr(prec) // left-associative (matches Python PLY: ('left', 'ARROW'))
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
		return p.setLoc(ast.NewApp(ast.NewSymbol("*>", nil), left, right), tok)

	case lexer.PLUS:
		p.advance()
		right := p.parseExpr(prec)
		// Python: App('+', left, right) — term-level operation
		return p.setLoc(ast.NewApp(ast.NewSymbol("+", nil), left, right), tok)

	case lexer.MINUS:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewApp(ast.NewSymbol("-", nil), left, right), tok)

	case lexer.TIMES:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewApp(ast.NewSymbol("*", nil), left, right), tok)

	case lexer.DIV:
		p.advance()
		right := p.parseExpr(prec)
		return p.setLoc(ast.NewApp(ast.NewSymbol("/", nil), left, right), tok)

	case lexer.DOT:
		p.advance()
		right := p.parsePrefix()
		// Python: compose_atoms(p[1], p[3]) — flatten dotted names into
		// single qualified atoms instead of creating Dot wrapper nodes.
		// ivy_logic_parser.py line 151-163:
		//   if isinstance(p[1], (Atom, App)):
		//       p[0] = compose_atoms(p[1], p[3])
		//   elif isinstance(p[1], Old):
		//       t = compose_atoms(p[1].args[0], p[3])
		//       p[0] = Old(t)
		//   else:
		//       p[0] = MethodCall(p[1], p[3])
		return p.setLoc(composeAtomsExpr(left, right), tok)

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
		// DOT-delimited form: some x:t, y:u . fmla
		// Python: bounds : params DOT (params are SYMBOL:SYMBOL)
		bounds = p.parseSomeParams()
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
