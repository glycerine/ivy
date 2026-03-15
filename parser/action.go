package parser

import (
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// parseActionBody parses an action body (sequence or single expression).
func (p *Parser) parseActionBody() ast.Node {
	if p.at(lexer.LCB) {
		return p.parseSequence()
	}
	return p.parseExpr(0)
}

// parseActionSeq parses a sequence of actions without braces,
// stopping at DOTDOTDOT, RCB, or EOF. Used by around { before ... after }.
// Python: actseq grammar
func (p *Parser) parseActionSeq() ast.Node {
	tok := p.current
	var stmts []ast.Node
	for !p.at(lexer.DOTDOTDOT) && !p.at(lexer.RCB) && !p.at(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		for p.match(lexer.SEMI) {
		}
	}
	if len(stmts) == 0 {
		return p.setLoc(ast.NewAnd(), tok)
	}
	if len(stmts) == 1 {
		return stmts[0]
	}
	return p.setLoc(ast.NewAnd(stmts...), tok)
}

// parseSequence parses { action; action; ... }.
func (p *Parser) parseSequence() ast.Node {
	tok := p.current
	p.expect(lexer.LCB)

	if p.match(lexer.RCB) {
		return p.setLoc(ast.NewAnd(), tok) // empty sequence = skip
	}

	var stmts []ast.Node
	for !p.at(lexer.RCB) && !p.at(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		// Consume optional semicolons between statements
		for p.match(lexer.SEMI) {
		}
		if len(p.errors) > 0 {
			break
		}
	}
	p.expect(lexer.RCB)

	if len(stmts) == 0 {
		return p.setLoc(ast.NewAnd(), tok)
	}
	if len(stmts) == 1 {
		return stmts[0]
	}
	return p.setLoc(ast.NewAnd(stmts...), tok)
}

// parseStatement parses a single statement in an action body.
func (p *Parser) parseStatement() ast.Node {
	tok := p.current

	switch tok.Type {
	case lexer.ASSUME:
		return p.parseAssumeAction(tok)
	case lexer.ASSERT:
		return p.parseAssertAction(tok)
	case lexer.REQUIRE:
		return p.parseRequireAction(tok)
	case lexer.ENSURE:
		return p.parseEnsureAction(tok)
	case lexer.IF:
		return p.parseIfAction(tok)
	case lexer.WHILE:
		return p.parseWhileAction(tok)
	case lexer.FOR:
		return p.parseForAction(tok)
	case lexer.LOCAL:
		return p.parseLocalAction(tok)
	case lexer.LET:
		return p.parseLetAction(tok)
	case lexer.VAR:
		return p.parseVarAction(tok)
	case lexer.CALL:
		return p.parseCallAction(tok)
	case lexer.LCB:
		return p.parseSequence()
	case lexer.DEBUG:
		return p.parseDebugAction(tok)
	case lexer.INSTANTIATE:
		return p.parseInstantiateDecl(tok)
	case lexer.THUNK:
		return p.parseThunkAction(tok)
	default:
		return p.parseExprStatement()
	}
}

// parseExprStatement parses an expression that may be an assignment or call.
func (p *Parser) parseExprStatement() ast.Node {
	tok := p.current
	lhs := p.parseExpr(0)

	// Check for assignment: lhs := rhs
	if p.match(lexer.ASSIGN) {
		if p.match(lexer.TIMES) {
			// Havoc: x := *
			return p.setLoc(ast.NewAtom("havoc", lhs), tok)
		}
		rhs := p.parseExpr(0)
		return p.setLoc(ast.NewAtom(":=", lhs, rhs), tok)
	}

	// Standalone expression (procedure call)
	return lhs
}

func (p *Parser) parseAssumeAction(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewAtom("assume", lf), tok)
}

func (p *Parser) parseAssertAction(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	if p.match(lexer.PROOF) {
		_ = p.parseProofBody()
	}
	return p.setLoc(ast.NewAtom("assert", lf), tok)
}

func (p *Parser) parseRequireAction(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewAtom("require", lf), tok)
}

func (p *Parser) parseEnsureAction(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewAtom("ensure", lf), tok)
}

func (p *Parser) parseIfAction(tok lexer.Token) ast.Node {
	p.advance()

	// Parse condition (may include "some")
	var cond ast.Node
	if p.match(lexer.TIMES) {
		cond = ast.NewSymbol("*", nil) // non-deterministic
	} else if p.at(lexer.SOME) {
		someTok := p.current
		p.advance()
		cond = p.parseSomeExpr(someTok)
	} else {
		cond = p.parseExpr(0)
	}

	thenBranch := p.parseSequence()

	var elseBranch ast.Node
	if p.match(lexer.ELSE) {
		if p.at(lexer.IF) {
			elseBranch = p.parseIfAction(p.current)
		} else {
			elseBranch = p.parseActionBody()
		}
	}

	if elseBranch == nil {
		return p.setLoc(ast.NewIte(cond, thenBranch, ast.NewAnd()), tok)
	}
	return p.setLoc(ast.NewIte(cond, thenBranch, elseBranch), tok)
}

func (p *Parser) parseWhileAction(tok lexer.Token) ast.Node {
	p.advance()
	cond := p.parseExpr(0)

	// Parse optional invariants
	var invs []ast.Node
	for p.match(lexer.INVARIANT) {
		lf := p.parseLabeledFmla()
		invs = append(invs, lf)
	}
	_ = invs

	// Parse optional decreases
	if p.match(lexer.DECREASES) {
		_ = p.parseExpr(0) // ranking function
	}

	body := p.parseSequence()
	return p.setLoc(ast.NewAtom("while", cond, body), tok)
}

func (p *Parser) parseForAction(tok lexer.Token) ast.Node {
	p.advance()
	iter := p.parseTTerm()
	p.expect(lexer.COMMA)
	idx := p.parseTTerm()
	p.expect(lexer.IN)
	collection := p.parseExpr(0)

	// Optional invariants
	for p.match(lexer.INVARIANT) {
		_ = p.parseLabeledFmla()
	}
	if p.match(lexer.DECREASES) {
		_ = p.parseExpr(0)
	}

	body := p.parseSequence()
	return p.setLoc(ast.NewAtom("for", iter, idx, collection, body), tok)
}

func (p *Parser) parseLocalAction(tok lexer.Token) ast.Node {
	p.advance()
	params := p.parseTTermList()
	body := p.parseSequence()
	args := append(params, body)
	return p.setLoc(ast.NewAtom("local", args...), tok)
}

func (p *Parser) parseLetAction(tok lexer.Token) ast.Node {
	p.advance()
	var defs []ast.Node
	for {
		d := p.parseExpr(0)
		defs = append(defs, d)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	body := p.parseSequence()
	allArgs := append(defs, body)
	return p.setLoc(ast.NewAtom("let", allArgs...), tok)
}

func (p *Parser) parseVarAction(tok lexer.Token) ast.Node {
	p.advance()
	tt := p.parseTTerm()
	var init ast.Node
	if p.match(lexer.ASSIGN) {
		init = p.parseExpr(0)
	}
	if init != nil {
		return p.setLoc(ast.NewAtom("var", tt, init), tok)
	}
	return p.setLoc(ast.NewAtom("var", tt), tok)
}

func (p *Parser) parseCallAction(tok lexer.Token) ast.Node {
	p.advance()
	// Python: CALL optactualreturns callatom
	// optactualreturns : callatoms ASSIGN | empty
	// Parse first callatom, then check for ASSIGN
	var returns []ast.Node
	first := p.parseCallatom()
	for p.match(lexer.COMMA) {
		returns = append(returns, first)
		first = p.parseCallatom()
	}
	if p.match(lexer.ASSIGN) {
		// Everything parsed so far was return variables
		returns = append(returns, first)
		// Now parse the actual call target
		ca := p.parseCallatom()
		args := append([]ast.Node{ca}, returns...)
		callNode := ast.NewAtom("call", args...)
		p.setLoc(callNode, tok)
		return callNode
	}
	// No ASSIGN — simple call
	if len(returns) > 0 {
		// We consumed commas but no ASSIGN — this shouldn't normally happen
		// Just return the last callatom
		return first
	}
	return first
}

func (p *Parser) parseDebugAction(tok lexer.Token) ast.Node {
	p.advance()
	name := p.expect(lexer.SYMBOL)
	var items []ast.Node
	if p.match(lexer.LPAREN) {
		for !p.at(lexer.RPAREN) && !p.at(lexer.EOF) {
			key := p.parseExpr(0)
			p.expect(lexer.EQ)
			val := p.parseExpr(0)
			items = append(items, &ast.DebugItem{Name: key, Value: val})
			if !p.match(lexer.COMMA) {
				break
			}
		}
		p.expect(lexer.RPAREN)
	}
	args := []ast.Node{ast.NewSymbol(name.Value, nil)}
	args = append(args, items...)
	return p.setLoc(ast.NewAtom("debug", args...), tok)
}

// parseThunkAction parses: thunk [label] name(args) : type := { body }
// Python: complexact : THUNK LABEL SYMBOL optargs COLON atype ASSIGN sequence
// Result: ThunkAction(Atom(label), Atom(name, args), Atom(type), body)
func (p *Parser) parseThunkAction(tok lexer.Token) ast.Node {
	p.advance() // consume THUNK
	// Parse [label] — Python: LABEL : LB SYMBOL RB
	p.expect(lexer.LB)
	labelTok := p.expect(lexer.SYMBOL)
	p.expect(lexer.RB)
	label := ast.NewAtom(labelTok.Value)
	p.setLoc(label, labelTok)

	// Parse name with optional args
	nameTok := p.expect(lexer.SYMBOL)
	var nameArgs []ast.Node
	if p.match(lexer.LPAREN) {
		nameArgs = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}
	action := ast.NewAtom(nameTok.Value, nameArgs...)
	p.setLoc(action, nameTok)

	// Parse : type
	p.expect(lexer.COLON)
	sortNode := p.parseAType()

	// Parse := body
	p.expect(lexer.ASSIGN)
	body := p.parseSequence()

	return p.setLoc(ast.NewThunkAction(label, action, sortNode, body), tok)
}
