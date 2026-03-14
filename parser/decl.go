package parser

import (
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// parseTopLevel parses one top-level declaration.
func (p *Parser) parseTopLevel() ast.Node {
	tok := p.current

	switch tok.Type {
	case lexer.TYPE:
		return p.parseTypeDecl(tok)
	case lexer.RELATION:
		return p.parseRelationDecl(tok)
	case lexer.INDIV:
		return p.parseConstantDecl(tok)
	case lexer.FUNCTION:
		return p.parseFunctionDecl(tok)
	case lexer.AXIOM:
		return p.parseAxiomDecl(tok)
	case lexer.PROPERTY:
		return p.parsePropertyDecl(tok)
	case lexer.CONJECTURE:
		return p.parseConjectureDecl(tok)
	case lexer.ACTION:
		return p.parseActionDecl(tok)
	case lexer.INIT:
		return p.parseInitDecl(tok)
	case lexer.MODULE:
		return p.parseModuleDecl(tok)
	case lexer.OBJECT:
		return p.parseObjectDecl(tok)
	case lexer.CLASS:
		return p.parseObjectDecl(tok) // same as object
	case lexer.ISOLATE:
		return p.parseIsolateDecl(tok)
	case lexer.EXPORT:
		return p.parseExportDecl(tok)
	case lexer.IMPORT:
		return p.parseImportDecl(tok)
	case lexer.INSTANTIATE:
		return p.parseInstantiateDecl(tok)
	case lexer.INTERPRET:
		return p.parseInterpretDecl(tok)
	case lexer.MIXIN:
		return p.parseMixinDecl(tok)
	case lexer.BEFORE:
		return p.parseMixinShorthand(tok, "before")
	case lexer.AFTER:
		return p.parseMixinShorthand(tok, "after")
	case lexer.IMPLEMENT:
		return p.parseImplementDecl(tok)
	case lexer.VARIANT:
		return p.parseVariantDecl(tok)
	case lexer.DEFINITION:
		return p.parseDefinitionDecl(tok)
	case lexer.DESTRUCTOR:
		return p.parseDestructorDecl(tok)
	case lexer.CONSTRUCTOR:
		return p.parseConstructorDecl(tok)
	case lexer.SCHEMA:
		return p.parseSchemaDecl(tok)
	case lexer.THEOREM:
		return p.parseTheoremDecl(tok)
	case lexer.PROOF:
		return p.parseProofDecl(tok)
	case lexer.ATTRIBUTE:
		return p.parseAttributeDecl(tok)
	case lexer.PRIVATE:
		return p.parsePrivateDecl(tok)
	case lexer.ALIAS:
		return p.parseAliasDecl(tok)
	case lexer.DELEGATE:
		return p.parseDelegateDecl(tok)
	case lexer.NATIVEQUOTE:
		return p.parseNativeDecl(tok)
	case lexer.INCLUDE:
		return p.parseIncludeDecl(tok)
	case lexer.USING:
		return p.parseUsingDecl(tok)
	case lexer.PROGRESS:
		return p.parseProgressDecl(tok)
	case lexer.RELY:
		return p.parseRelyDecl(tok)
	case lexer.EXTRACT:
		return p.parseExtractDecl(tok)
	case lexer.PARAMETER:
		return p.parseParameterDecl(tok)
	case lexer.VAR:
		return p.parseVarDecl(tok)
	case lexer.MACRO:
		return p.parseMacroDecl(tok)
	case lexer.INVARIANT:
		return p.parseInvariantDecl(tok)
	case lexer.TEMPORAL:
		return p.parseTemporalDecl(tok)
	case lexer.EXPLICIT:
		return p.parseExplicitDecl(tok)
	case lexer.AUTOINSTANCE:
		return p.parseAutoInstanceDecl(tok)
	case lexer.SCENARIO:
		return p.parseScenarioDecl(tok)
	case lexer.COMMON:
		return p.parseCommonBlock(tok)
	case lexer.SPECIFICATION:
		return p.parseSpecBlock(tok)
	case lexer.IMPLEMENTATION:
		return p.parseImplBlock(tok)
	case lexer.GLOBAL:
		return p.parseGlobalDecl(tok)

	// v1.7+ statement-level constructs allowed at top level
	case lexer.IF:
		return p.parseIfAction(tok)
	case lexer.WHILE:
		return p.parseWhileAction(tok)
	case lexer.ASSUME:
		return p.parseAssumeAction(tok)
	case lexer.ASSERT:
		return p.parseAssertAction(tok)
	case lexer.REQUIRE:
		return p.parseRequireAction(tok)
	case lexer.ENSURE:
		return p.parseEnsureAction(tok)
	case lexer.LCB:
		return p.parseSequence()

	default:
		// Try to parse as expression/action
		if tok.Type == lexer.SYMBOL || tok.Type == lexer.VARIABLE || tok.Type == lexer.THIS {
			return p.parseExprStatement()
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
	p.advance()
	terms := p.parseTTermList()
	return p.setLoc(ast.NewRelationDecl(terms...), tok)
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
	ca := p.parseCallatom()

	// Parse formal params
	var params []ast.Node
	if p.match(lexer.LPAREN) {
		params = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}

	// Parse returns
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
	lhs := p.parseExpr(0)
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

func (p *Parser) parseMixinShorthand(tok lexer.Token, kind string) ast.Node {
	p.advance()
	ca := p.parseCallatom()

	// Parse optional params
	var params []ast.Node
	if p.match(lexer.LPAREN) {
		params = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}
	_ = params

	body := p.parseActionBody()
	adef := ast.NewActionDef(ca, body, params, nil)

	var mdef ast.Node
	switch kind {
	case "before":
		mdef = &ast.MixinBeforeDef{Mixer: adef, Mixee: ca}
	case "after":
		mdef = &ast.MixinAfterDef{Mixer: adef, Mixee: ca}
	default:
		mdef = &ast.MixinImplementDef{Mixer: adef, Mixee: ca}
	}
	return p.setLoc(ast.NewMixinDecl(mdef), tok)
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
	return p.setLoc(ast.NewPropertyDecl(lf), tok) // invariant ≈ property
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

func (p *Parser) parseCommonBlock(tok lexer.Token) ast.Node {
	p.advance()
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	// Wrap in a module-like structure
	return p.setLoc(ast.NewObjectDecl(body...), tok)
}

func (p *Parser) parseSpecBlock(tok lexer.Token) ast.Node {
	p.advance()
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	return p.setLoc(ast.NewObjectDecl(body...), tok)
}

func (p *Parser) parseImplBlock(tok lexer.Token) ast.Node {
	p.advance()
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)
	return p.setLoc(ast.NewObjectDecl(body...), tok)
}

func (p *Parser) parseGlobalDecl(tok lexer.Token) ast.Node {
	p.advance()
	// global can prefix other declarations
	return p.parseTopLevel()
}

// parseBlock parses declarations until RCB or EOF.
func (p *Parser) parseBlock() ([]ast.Node, error) {
	// Skip optional ...
	p.match(lexer.DOTDOTDOT)

	var decls []ast.Node
	for !p.at(lexer.RCB) && !p.at(lexer.EOF) {
		d := p.parseTopLevel()
		if d != nil {
			decls = append(decls, d)
		}
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
