package parser

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// parseTopLevel parses one top-level declaration, returning one or more
// AST nodes. Matches Python where some constructs (e.g., "after init")
// produce multiple declarations.
func (p *Parser) parseTopLevel() []ast.Node {
	tok := p.current

	// one wraps a single node parse result as a slice.
	one := func(n ast.Node) []ast.Node {
		if n == nil {
			return nil
		}
		return []ast.Node{n}
	}

	switch tok.Type {
	case lexer.TYPE:
		return one(p.parseTypeDecl(tok))
	case lexer.RELATION:
		return one(p.parseRelationDecl(tok))
	case lexer.INDIV:
		return one(p.parseConstantDecl(tok))
	case lexer.FUNCTION:
		return one(p.parseFunctionDecl(tok))
	case lexer.AXIOM:
		return one(p.parseAxiomDecl(tok))
	case lexer.PROPERTY:
		return one(p.parsePropertyDecl(tok))
	case lexer.CONJECTURE:
		return one(p.parseConjectureDecl(tok))
	case lexer.ACTION:
		return one(p.parseActionDecl(tok))
	case lexer.INIT:
		return one(p.parseInitDecl(tok))
	case lexer.MODULE:
		return one(p.parseModuleDecl(tok))
	case lexer.OBJECT:
		return one(p.parseObjectDecl(tok))
	case lexer.CLASS:
		return one(p.parseObjectDecl(tok))
	case lexer.ISOLATE:
		return one(p.parseIsolateDecl(tok))
	case lexer.EXPORT:
		return one(p.parseExportDecl(tok))
	case lexer.IMPORT:
		return one(p.parseImportDecl(tok))
	case lexer.INSTANTIATE:
		return one(p.parseInstantiateDecl(tok))
	case lexer.INTERPRET:
		return one(p.parseInterpretDecl(tok))
	case lexer.MIXIN:
		return one(p.parseMixinDecl(tok))
	case lexer.BEFORE:
		return p.parseMixinShorthand(tok, "before")
	case lexer.AFTER:
		return p.parseMixinShorthand(tok, "after")
	case lexer.IMPLEMENT:
		return one(p.parseImplementDecl(tok))
	case lexer.VARIANT:
		return one(p.parseVariantDecl(tok))
	case lexer.DEFINITION:
		return one(p.parseDefinitionDecl(tok))
	case lexer.DESTRUCTOR:
		return one(p.parseDestructorDecl(tok))
	case lexer.CONSTRUCTOR:
		return one(p.parseConstructorDecl(tok))
	case lexer.SCHEMA:
		return one(p.parseSchemaDecl(tok))
	case lexer.THEOREM:
		return one(p.parseTheoremDecl(tok))
	case lexer.PROOF:
		return one(p.parseProofDecl(tok))
	case lexer.ATTRIBUTE:
		return one(p.parseAttributeDecl(tok))
	case lexer.PRIVATE:
		return p.parsePrivateBlock(tok)
	case lexer.ALIAS:
		return one(p.parseAliasDecl(tok))
	case lexer.DELEGATE:
		return one(p.parseDelegateDecl(tok))
	case lexer.NATIVEQUOTE:
		return one(p.parseNativeDecl(tok))
	case lexer.INCLUDE:
		return one(p.parseIncludeDecl(tok))
	case lexer.USING:
		return one(p.parseUsingDecl(tok))
	case lexer.PROGRESS:
		return one(p.parseProgressDecl(tok))
	case lexer.RELY:
		return one(p.parseRelyDecl(tok))
	case lexer.EXTRACT:
		return one(p.parseExtractDecl(tok))
	case lexer.PARAMETER:
		return one(p.parseParameterDecl(tok))
	case lexer.VAR:
		return one(p.parseVarDecl(tok))
	case lexer.MACRO:
		return one(p.parseMacroDecl(tok))
	case lexer.INVARIANT:
		return one(p.parseInvariantDecl(tok))
	case lexer.TEMPORAL:
		return one(p.parseTemporalDecl(tok))
	case lexer.EXPLICIT:
		return one(p.parseExplicitDecl(tok))
	case lexer.AUTOINSTANCE:
		return one(p.parseAutoInstanceDecl(tok))
	case lexer.SCENARIO:
		return one(p.parseScenarioDecl(tok))
	case lexer.COMMON:
		return p.parseCommonBlock(tok)
	case lexer.SPECIFICATION:
		return p.parseSpecBlock(tok)
	case lexer.IMPLEMENTATION:
		return p.parseImplBlock(tok)
	case lexer.GLOBAL:
		return one(p.parseGlobalDecl(tok))

	// v1.7+ statement-level constructs allowed at top level
	case lexer.IF:
		return one(p.parseIfAction(tok))
	case lexer.WHILE:
		return one(p.parseWhileAction(tok))
	case lexer.ASSUME:
		return one(p.parseAssumeAction(tok))
	case lexer.ASSERT:
		return one(p.parseAssertAction(tok))
	case lexer.REQUIRE:
		return one(p.parseRequireAction(tok))
	case lexer.ENSURE:
		return one(p.parseEnsureAction(tok))
	case lexer.LCB:
		return one(p.parseSequence())

	default:
		// Try to parse as expression/action
		if tok.Type == lexer.SYMBOL || tok.Type == lexer.VARIABLE || tok.Type == lexer.THIS {
			return one(p.parseExprStatement())
		}
		p.errorf("unexpected token at top level: %s (%q)", tok.Type, tok.Value)
		p.advance()
		return nil
	}
}

func (p *Parser) parseTypeDecl(tok lexer.Token) ast.Node {
	p.advance() // consume TYPE
	td := p.parseTypeDef()
	return p.setLoc(ast.NewTypeDecl(td), tok)
}

func (p *Parser) parseTypeDef() ast.Node {
	tok := p.current
	name := p.parseSymbol()

	if p.match(lexer.EQ) {
		sort := p.parseSort()
		td := ast.NewTypeDef(name, sort)
		p.setLoc(td, tok)
		return td
	}

	// Just "type t" with no definition
	td := ast.NewTypeDef(name, ast.NewConstantSort())
	p.setLoc(td, tok)
	return td
}

func (p *Parser) parseRelationDecl(tok lexer.Token) ast.Node {
	// In Python Ivy, "relation r(X:t)" produces ConstantDecl with sort=bool,
	// NOT a separate RelationDecl. Relations are just constants with Boolean range.
	p.advance()
	terms := p.parseTTermList()
	return p.setLoc(ast.NewConstantDecl(terms...), tok)
}

func (p *Parser) parseConstantDecl(tok lexer.Token) ast.Node {
	p.advance()
	terms := p.parseTTermList()
	return p.setLoc(ast.NewConstantDecl(terms...), tok)
}

func (p *Parser) parseFunctionDecl(tok lexer.Token) ast.Node {
	p.advance()
	terms := p.parseTTermList()
	return p.setLoc(ast.NewConstantDecl(terms...), tok) // function maps to individual
}

func (p *Parser) parseAxiomDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewAxiomDecl(lf), tok)
}

func (p *Parser) parsePropertyDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	// Optional proof
	if p.match(lexer.PROOF) {
		_ = p.parseProofBody()
	}
	return p.setLoc(ast.NewPropertyDecl(lf), tok)
}

func (p *Parser) parseConjectureDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewConjectureDecl(lf), tok)
}

func (p *Parser) parseActionDecl(tok lexer.Token) ast.Node {
	p.advance()
	adef := p.parseActionDef()
	return p.setLoc(ast.NewActionDecl(adef), tok)
}

func (p *Parser) parseActionDef() ast.Node {
	tok := p.current
	// Parse just the action name (not including params).
	// Matches Python grammar: ACTION atype optargs optreturns EQ sequence
	name, nameTok := p.parseAtomName()
	ca := ast.NewAtom(name)
	p.setLoc(ca, nameTok)

	// Handle dot chaining: a.b.c
	for p.match(lexer.DOT) {
		name2, tok2 := p.parseAtomName()
		right := ast.NewAtom(name2)
		p.setLoc(right, tok2)
		name = name + "." + name2
		ca = ast.NewAtom(name)
		p.setLoc(ca, tok2)
	}

	// Parse formal params: (x:client, y:server)
	var params []ast.Node
	if p.match(lexer.LPAREN) {
		params = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}

	// Parse returns: returns (r:type)
	var returns []ast.Node
	if p.match(lexer.RETURNS) {
		p.expect(lexer.LPAREN)
		returns = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}

	var body ast.Node
	if p.match(lexer.EQ) {
		body = p.parseActionBody()
	} else {
		body = ast.NewAnd() // empty body = true
	}

	ad := ast.NewActionDef(ca, body, params, returns)
	p.setLoc(ad, tok)
	return ad
}

func (p *Parser) parseInitDecl(tok lexer.Token) ast.Node {
	p.advance()
	body := p.parseActionBody()
	return p.setLoc(ast.NewInitDecl(body), tok)
}

func (p *Parser) parseModuleDecl(tok lexer.Token) ast.Node {
	p.advance()
	name := p.parseCallatom()
	var params []ast.Node
	if p.match(lexer.LPAREN) {
		params = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}
	_ = params
	p.expect(lexer.EQ)
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	defn := ast.NewDefinition(name, blockToNode(body))
	return p.setLoc(ast.NewModuleDecl(defn), tok)
}

func (p *Parser) parseObjectDecl(tok lexer.Token) ast.Node {
	p.advance()
	name := p.parseCallatom()
	p.expect(lexer.EQ)
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	defn := ast.NewDefinition(name, blockToNode(body))
	return p.setLoc(ast.NewObjectDecl(defn), tok)
}

func (p *Parser) parseIsolateDecl(tok lexer.Token) ast.Node {
	p.advance()
	ca := p.parseCallatom()

	elems := []ast.Node{ca}
	withArgs := 0
	if p.match(lexer.EQ) {
		p.expect(lexer.LCB)
		body, _ := p.parseBlock()
		p.expect(lexer.RCB)
		return p.setLoc(ast.NewIsolateDecl(&ast.IsolateDef{Elems: []ast.Node{ca, blockToNode(body)}, WithArgs: 0}), tok)
	}
	if p.match(lexer.WITH) {
		for {
			elems = append(elems, p.parseCallatom())
			withArgs++
			if !p.match(lexer.COMMA) {
				break
			}
		}
	}
	idef := &ast.IsolateDef{Elems: elems, WithArgs: withArgs}
	return p.setLoc(ast.NewIsolateDecl(idef), tok)
}

func (p *Parser) parseExportDecl(tok lexer.Token) ast.Node {
	p.advance()
	ca := p.parseCallatom()
	return p.setLoc(ast.NewExportDecl(&ast.ExportDef{Exported: ca, Scope: &ast.NoneAST{}}), tok)
}

func (p *Parser) parseImportDecl(tok lexer.Token) ast.Node {
	p.advance()
	ca := p.parseCallatom()
	return p.setLoc(ast.NewImportDecl(&ast.ImportDef{Imported: ca, Scope: &ast.NoneAST{}}), tok)
}

func (p *Parser) parseInstantiateDecl(tok lexer.Token) ast.Node {
	p.advance()
	var insts []ast.Node
	for {
		var label ast.Node
		ca := p.parseCallatom()
		if p.match(lexer.COLON) {
			label = ca
			ca = p.parseCallatom()
		}
		inst := ast.NewInstantiation(label, ca)
		p.setLoc(inst, tok)
		insts = append(insts, inst)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	return p.setLoc(ast.NewInstantiateDecl(insts...), tok)
}

func (p *Parser) parseInterpretDecl(tok lexer.Token) ast.Node {
	p.advance()
	// Parse LHS as a simple name (not a full expression, to avoid consuming ->)
	lhs := p.parseAType()
	p.expect(lexer.ARROW)
	rhs := p.parseSort()
	def := ast.NewDefinition(lhs, rhs)
	lf := ast.NewLabeledFormula(lhs, def)
	return p.setLoc(ast.NewInterpretDecl(lf), tok)
}

func (p *Parser) parseMixinDecl(tok lexer.Token) ast.Node {
	p.advance()
	ca := p.parseCallatom()
	if p.match(lexer.BEFORE) {
		mixee := p.parseCallatom()
		return p.setLoc(ast.NewMixinDecl(&ast.MixinBeforeDef{Mixer: ca, Mixee: mixee}), tok)
	}
	if p.match(lexer.AFTER) {
		mixee := p.parseCallatom()
		return p.setLoc(ast.NewMixinDecl(&ast.MixinAfterDef{Mixer: ca, Mixee: mixee}), tok)
	}
	return p.setLoc(ast.NewMixinDecl(ca), tok)
}

// parseMixinShorthand parses "before name { ... }" or "after name { ... }".
// Matches Python's handle_before_after: produces BOTH an ActionDecl and a MixinDecl.
func (p *Parser) parseMixinShorthand(tok lexer.Token, kind string) []ast.Node {
	p.advance()
	// Special case: "after init { ... }" — init is a keyword, not a symbol.
	var ca ast.Node
	if p.at(lexer.INIT) {
		initTok := p.current
		p.advance()
		ca = ast.NewAtom("init")
		p.setLoc(ca, initTok)
	} else {
		ca = p.parseCallatom()
	}

	// Parse optional params
	var params []ast.Node
	if p.match(lexer.LPAREN) {
		params = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}

	// Parse optional returns
	var returns []ast.Node
	if p.match(lexer.RETURNS) {
		p.expect(lexer.LPAREN)
		returns = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}

	body := p.parseActionBody()

	// Generate mixer name: "init[after1]" matching Python's make_mixin_name.
	mixerName := ca.String()
	p.labelCounter++
	mixerName = mixerName + "[" + kind + fmt.Sprintf("%d", p.labelCounter) + "]"
	mixer := ast.NewAtom(mixerName)
	p.setLoc(mixer, tok)

	// 1. ActionDecl — the action with the generated mixer name
	adef := ast.NewActionDef(mixer, body, params, returns)
	p.setLoc(adef, tok)
	actionDecl := ast.NewActionDecl(adef)
	p.setLoc(actionDecl, tok)

	// 2. MixinDecl — connects the mixer to the mixee
	var mdef ast.Node
	switch kind {
	case "before":
		mdef = &ast.MixinBeforeDef{Mixer: mixer, Mixee: ca}
	case "after":
		mdef = &ast.MixinAfterDef{Mixer: mixer, Mixee: ca}
	default:
		mdef = &ast.MixinImplementDef{Mixer: mixer, Mixee: ca}
	}
	mixinDecl := ast.NewMixinDecl(mdef)
	p.setLoc(mixinDecl, tok)

	return []ast.Node{actionDecl, mixinDecl}
}

func (p *Parser) parseImplementDecl(tok lexer.Token) ast.Node {
	p.advance()
	ca := p.parseCallatom()

	// "implement action_name { ... }"
	if p.at(lexer.LCB) || p.at(lexer.LPAREN) {
		var params []ast.Node
		if p.match(lexer.LPAREN) {
			params = p.parseTTermList()
			p.expect(lexer.RPAREN)
		}
		body := p.parseActionBody()
		adef := ast.NewActionDef(ca, body, params, nil)
		mdef := &ast.MixinImplementDef{Mixer: adef, Mixee: ca}
		return p.setLoc(ast.NewMixinDecl(mdef), tok)
	}

	return p.setLoc(ast.NewMixinDecl(&ast.MixinImplementDef{Mixer: ca, Mixee: ca}), tok)
}

func (p *Parser) parseVariantDecl(tok lexer.Token) ast.Node {
	p.advance()
	name := p.parseExpr(0)
	p.expect(lexer.OF)
	base := p.parseAType()
	vd := ast.NewVariantDef(name, base)
	return p.setLoc(ast.NewVariantDecl(vd), tok)
}

func (p *Parser) parseDefinitionDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewDefinitionDecl(lf), tok)
}

func (p *Parser) parseDestructorDecl(tok lexer.Token) ast.Node {
	p.advance()
	terms := p.parseTTermList()
	return p.setLoc(ast.NewDestructorDecl(terms...), tok)
}

func (p *Parser) parseConstructorDecl(tok lexer.Token) ast.Node {
	p.advance()
	terms := p.parseTTermList()
	return p.setLoc(ast.NewConstructorDecl(terms...), tok)
}

func (p *Parser) parseSchemaDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewSchemaDecl(lf), tok)
}

func (p *Parser) parseTheoremDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	if p.match(lexer.PROOF) {
		_ = p.parseProofBody()
	}
	return p.setLoc(ast.NewTheoremDecl(lf), tok)
}

func (p *Parser) parseProofDecl(tok lexer.Token) ast.Node {
	p.advance()
	body := p.parseProofBody()
	return p.setLoc(ast.NewProofDecl(body), tok)
}

func (p *Parser) parseAttributeDecl(tok lexer.Token) ast.Node {
	p.advance()
	name := p.parseCallatom()
	p.expect(lexer.EQ)
	val := p.parseExpr(0)
	ad := ast.NewAttributeDef(name, val)
	return p.setLoc(ast.NewAttributeDecl(ad), tok)
}

func (p *Parser) parsePrivateDecl(tok lexer.Token) ast.Node {
	p.advance()
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	return p.setLoc(ast.NewPrivateDecl(body...), tok)
}

// parsePrivateBlock parses "private { decls }" and returns the inner
// declarations directly (not wrapped), matching Python behavior.
func (p *Parser) parsePrivateBlock(tok lexer.Token) []ast.Node {
	p.advance()
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	return body
}

func (p *Parser) parseAliasDecl(tok lexer.Token) ast.Node {
	p.advance()
	name := p.parseSymbol()
	p.expect(lexer.EQ)
	target := p.parseAType()
	def := ast.NewDefinition(name, target)
	return p.setLoc(ast.NewAliasDecl(def), tok)
}

func (p *Parser) parseDelegateDecl(tok lexer.Token) ast.Node {
	p.advance()
	ca := p.parseCallatom()
	return p.setLoc(ast.NewDelegateDecl(&ast.DelegateDef{Elems: []ast.Node{ca}}), tok)
}

func (p *Parser) parseNativeDecl(tok lexer.Token) ast.Node {
	nq := p.advance() // consume NATIVEQUOTE
	nc := ast.NewNativeCode(nq.Value)
	return p.setLoc(ast.NewNativeDecl(nc), tok)
}

func (p *Parser) parseIncludeDecl(tok lexer.Token) ast.Node {
	p.advance()
	name := p.expect(lexer.SYMBOL)
	return p.setLoc(ast.NewNamedDecl(ast.NewSymbol(name.Value, nil)), tok)
}

func (p *Parser) parseUsingDecl(tok lexer.Token) ast.Node {
	p.advance()
	name := p.expect(lexer.SYMBOL)
	return p.setLoc(ast.NewNamedDecl(ast.NewSymbol(name.Value, nil)), tok)
}

func (p *Parser) parseProgressDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewProgressDecl(lf), tok)
}

func (p *Parser) parseRelyDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewRelyDecl(lf), tok)
}

func (p *Parser) parseExtractDecl(tok lexer.Token) ast.Node {
	p.advance()
	ca := p.parseCallatom()
	elems := []ast.Node{ca}
	withArgs := 0
	if p.match(lexer.EQ) {
		p.expect(lexer.LCB)
		body, _ := p.parseBlock()
		p.expect(lexer.RCB)
		elems = append(elems, blockToNode(body))
	}
	if p.match(lexer.WITH) {
		for {
			elems = append(elems, p.parseCallatom())
			withArgs++
			if !p.match(lexer.COMMA) {
				break
			}
		}
	}
	edef := &ast.ExtractDef{IsolateDef: ast.IsolateDef{Elems: elems, WithArgs: withArgs}}
	return p.setLoc(ast.NewIsolateDecl(edef), tok)
}

func (p *Parser) parseParameterDecl(tok lexer.Token) ast.Node {
	p.advance()
	terms := p.parseTTermList()
	return p.setLoc(ast.NewParameterDecl(terms...), tok)
}

func (p *Parser) parseVarDecl(tok lexer.Token) ast.Node {
	p.advance()
	terms := p.parseTTermList()
	return p.setLoc(ast.NewConstantDecl(terms...), tok)
}

func (p *Parser) parseMacroDecl(tok lexer.Token) ast.Node {
	p.advance()
	ca := p.parseCallatom()
	p.expect(lexer.EQ)
	body := p.parseActionBody()
	def := ast.NewDefinition(ca, body)
	return p.setLoc(ast.NewMacroDecl(def), tok)
}

func (p *Parser) parseInvariantDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	if p.match(lexer.PROOF) {
		_ = p.parseProofBody()
	}
	return p.setLoc(ast.NewConjectureDecl(lf), tok) // invariant → conjecture (matches Python)
}

func (p *Parser) parseTemporalDecl(tok lexer.Token) ast.Node {
	p.advance()
	// temporal property ...
	if p.match(lexer.PROPERTY) {
		lf := p.parseLabeledFmla()
		lf.Temporal = ast.NewSymbol("temporal", nil)
		return p.setLoc(ast.NewPropertyDecl(lf), tok)
	}
	p.errorf("expected 'property' after 'temporal'")
	return nil
}

func (p *Parser) parseExplicitDecl(tok lexer.Token) ast.Node {
	p.advance()
	// explicit property ...
	if p.match(lexer.PROPERTY) {
		lf := p.parseLabeledFmla()
		lf.Explicit = true
		return p.setLoc(ast.NewPropertyDecl(lf), tok)
	}
	p.errorf("expected 'property' after 'explicit'")
	return nil
}

func (p *Parser) parseAutoInstanceDecl(tok lexer.Token) ast.Node {
	p.advance()
	var insts []ast.Node
	for {
		ca := p.parseCallatom()
		inst := ast.NewInstantiation(nil, ca)
		insts = append(insts, inst)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	return p.setLoc(ast.NewAutoInstanceDecl(insts...), tok)
}

func (p *Parser) parseScenarioDecl(tok lexer.Token) ast.Node {
	p.advance()
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	return p.setLoc(ast.NewScenarioDecl(body...), tok)
}

// parseCommonBlock, parseSpecBlock, parseImplBlock inline their contents
// to match Python behavior where specification/implementation/common blocks
// are scope modifiers, not wrapper nodes.
func (p *Parser) parseCommonBlock(tok lexer.Token) []ast.Node {
	p.advance()
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	return body
}

func (p *Parser) parseSpecBlock(tok lexer.Token) []ast.Node {
	p.advance()
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	return body
}

func (p *Parser) parseImplBlock(tok lexer.Token) []ast.Node {
	p.advance()
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	return body
}

func (p *Parser) parseGlobalDecl(tok lexer.Token) ast.Node {
	p.advance()
	return nil // global is a scope modifier; handled by the next declaration
}

// parseBlock parses declarations until RCB or EOF.
func (p *Parser) parseBlock() ([]ast.Node, error) {
	// Skip optional ...
	p.match(lexer.DOTDOTDOT)

	var decls []ast.Node
	for !p.at(lexer.RCB) && !p.at(lexer.EOF) {
		ds := p.parseTopLevel()
		decls = append(decls, ds...)
		if len(p.errors) > 0 {
			return decls, &p.errors[0]
		}
	}
	return decls, nil
}

// blockToNode wraps a list of declarations in a single node.
func blockToNode(decls []ast.Node) ast.Node {
	if len(decls) == 0 {
		return ast.NewAnd() // empty = true
	}
	if len(decls) == 1 {
		return decls[0]
	}
	return ast.NewAnd(decls...)
}

// parseProofBody parses a proof body (simplified).
func (p *Parser) parseProofBody() ast.Node {
	tok := p.current
	if p.match(lexer.LCB) {
		var steps []ast.Node
		for !p.at(lexer.RCB) && !p.at(lexer.EOF) {
			step := p.parseProofStep()
			if step != nil {
				steps = append(steps, step)
			}
			p.match(lexer.SEMI)
		}
		p.expect(lexer.RCB)
		if len(steps) == 0 {
			return &ast.NullTactic{}
		}
		if len(steps) == 1 {
			return steps[0]
		}
		return &ast.ComposeTactics{Tactics: steps}
	}
	// Single proof step
	return p.setLoc(p.parseProofStep(), tok)
}

func (p *Parser) parseProofStep() ast.Node {
	tok := p.current
	switch tok.Type {
	case lexer.APPLY:
		p.advance()
		schema := p.parseCallatom()
		return p.setLoc(&ast.SchemaInstantiation{SchemaName: schema, Ren: &ast.NoneAST{}}, tok)
	case lexer.SHOWGOALS:
		p.advance()
		return p.setLoc(&ast.ShowGoalsTactic{}, tok)
	case lexer.UNFOLD:
		p.advance()
		name := p.parseCallatom()
		return p.setLoc(&ast.UnfoldTactic{Premise: &ast.NoneAST{}, UnfSpecs: []ast.Node{&ast.UnfoldSpec{DefName: name}}}, tok)
	case lexer.TACTIC:
		p.advance()
		name := p.parseCallatom()
		return p.setLoc(&ast.TacticTactic{TName: name, Body: &ast.NoneAST{}}, tok)
	case lexer.LET:
		p.advance()
		var defs []ast.Node
		for {
			d := p.parseExpr(0)
			defs = append(defs, d)
			if !p.match(lexer.COMMA) {
				break
			}
		}
		return p.setLoc(&ast.LetTactic{Defs: defs}, tok)
	case lexer.PROPERTY:
		p.advance()
		lf := p.parseLabeledFmla()
		return p.setLoc(&ast.PropertyTactic{Prop: lf, PName: &ast.NoneAST{}, Proof: &ast.NoneAST{}}, tok)
	case lexer.LCB:
		return p.parseProofBody()
	default:
		// Fallback: try to parse as expression
		return p.parseExpr(0)
	}
}
