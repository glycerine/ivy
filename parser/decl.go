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
	case lexer.FINITE:
		// Python: top : top optfinite optghost TYPE typesymbol ...
		// "finite type foo" or "finite ghost type foo"
		p.advance()
		isGhost := p.match(lexer.GHOST)
		if p.at(lexer.TYPE) {
			td := p.parseTypeDecl(p.current)
			if tDecl, ok := td.(*ast.TypeDecl); ok {
				for i, arg := range tDecl.DeclArgs {
					if tDef, ok := arg.(*ast.TypeDef); ok {
						tDef.Finite = true
						if isGhost {
							ghost := &ast.GhostTypeDef{TypeDef: *tDef}
							tDecl.DeclArgs[i] = ghost
						}
					}
				}
			}
			return one(td)
		}
		p.errorf("expected 'type' after 'finite'")
		return nil
	case lexer.GHOST:
		// Python: top : top optfinite optghost TYPE typesymbol ...
		// "ghost type foo" (finite not before ghost in this case)
		p.advance()
		if p.at(lexer.TYPE) {
			td := p.parseTypeDecl(p.current)
			if tDecl, ok := td.(*ast.TypeDecl); ok {
				for i, arg := range tDecl.DeclArgs {
					if tDef, ok := arg.(*ast.TypeDef); ok {
						ghost := &ast.GhostTypeDef{TypeDef: *tDef}
						tDecl.DeclArgs[i] = ghost
					}
				}
			}
			return one(td)
		}
		p.errorf("expected 'type' after 'ghost'")
		return nil
	case lexer.RELATION:
		return p.parseRelationDeclMulti(tok)
	case lexer.INDIV:
		return one(p.parseConstantDecl(tok))
	case lexer.FUNCTION:
		return p.parseFunctionDeclMulti(tok)
	case lexer.AXIOM:
		return p.parseAxiomDeclMulti(tok)
	case lexer.PROPERTY:
		return p.parsePropertyDeclMulti(tok)
	case lexer.CONJECTURE:
		return p.parseConjectureDeclMulti(tok)
	case lexer.ACTION:
		return one(p.parseActionDecl(tok))
	case lexer.INIT:
		return one(p.parseInitDecl(tok))
	case lexer.MODULE:
		return one(p.parseModuleDecl(tok))
	case lexer.OBJECT:
		return p.parseObjectDeclMulti(tok)
	case lexer.CLASS:
		return p.parseObjectDeclMulti(tok)
	case lexer.ISOLATE:
		return p.parseIsolateDeclMulti(tok)
	case lexer.TRUSTED:
		// trusted isolate ... — consume trusted, then parse isolate
		p.advance()
		if p.at(lexer.ISOLATE) {
			return p.parseIsolateDeclMulti(p.current)
		}
		p.errorf("expected 'isolate' after 'trusted'")
		return nil
	case lexer.EXPORT:
		return p.parseExportDeclMulti(tok)
	case lexer.IMPORT:
		return p.parseImportDeclMulti(tok)
	case lexer.INSTANTIATE:
		return p.parseInstantiateDeclMulti(tok)
	case lexer.INTERPRET:
		return one(p.parseInterpretDecl(tok))
	case lexer.MIXIN:
		return one(p.parseMixinDecl(tok))
	case lexer.BEFORE:
		return p.parseMixinShorthand(tok, "before")
	case lexer.AFTER:
		return p.parseMixinShorthand(tok, "after")
	case lexer.AROUND:
		return p.parseAroundDecl(tok)
	case lexer.IMPLEMENT:
		return p.parseImplementDeclMulti(tok)
	case lexer.VARIANT:
		return p.parseVariantDeclMulti(tok)
	case lexer.DEFINITION:
		return one(p.parseDefinitionDecl(tok))
	case lexer.DESTRUCTOR:
		return one(p.parseDestructorDecl(tok))
	case lexer.CONSTRUCTOR:
		return one(p.parseConstructorDecl(tok))
	case lexer.SCHEMA:
		return one(p.parseSchemaDecl(tok))
	case lexer.THEOREM:
		return p.parseTheoremDeclMulti(tok)
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
	case lexer.MIXORD:
		return one(p.parseMixOrdDecl(tok))
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
	case lexer.DERIVED:
		return p.parseDerivedDeclMulti(tok)
	case lexer.PARAMETER:
		return one(p.parseParameterDecl(tok))
	case lexer.VAR:
		return one(p.parseVarDecl(tok))
	case lexer.MACRO:
		return one(p.parseMacroDecl(tok))
	case lexer.INVARIANT:
		return p.parseInvariantDeclMulti(tok)
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
	case lexer.UNPROVABLE:
		// Python: unprovable invariant labeledfmla — when check_unprovable is False (default),
		// the declaration is NOT emitted. We consume and discard.
		p.advance()
		if p.at(lexer.INVARIANT) {
			// Parse the invariant declaration but don't emit it
			_ = p.parseInvariantDeclMulti(p.current)
			return nil // don't emit anything
		}
		// For other unprovable constructs, parse and discard
		_ = p.parseTopLevel()
		return nil

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
	// Use parseAtomName to accept both SYMBOL and THIS
	// (Python: 'typesymbol : SYMBOL | THIS')
	nameStr, _ := p.parseAtomName()
	name := ast.NewSymbol(nameStr, nil)
	p.setLoc(name, tok)

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

// parseRelationDeclMulti parses: relation rels (comma-separated)
// Python: top : top RELATION rels → for d in rels: declare(d)
func (p *Parser) parseRelationDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	var result []ast.Node
	for {
		rel := p.parseOneRel(tok)
		result = append(result, rel)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	return result
}

// parseOneRel parses a single relation declaration: defnlhs [= expr]
func (p *Parser) parseOneRel(tok lexer.Token) ast.Node {
	lhs := p.parseDefnLhs()
	if p.match(lexer.EQ) {
		body := p.parseExpr(0)
		defn := ast.NewDefinition(lhs, body)
		lf := ast.NewLabeledFormula(nil, defn)
		return p.setLoc(ast.NewDerivedDecl(lf), tok)
	}
	// Set sort to "bool" for relation declarations without a definition.
	// Matches Python ivy_parser.py line 970: p[1].sort = 'bool'
	if atom, ok := lhs.(*ast.Atom); ok {
		atom.ASort = ast.NewSymbol("bool", nil)
	}
	return p.setLoc(ast.NewConstantDecl(lhs), tok)
}

func (p *Parser) parseRelationDecl(tok lexer.Token) ast.Node {
	// Python's 'rel' grammar:
	//   rel : defnlhs          →  ConstantDecl (plain declaration)
	//   rel : defn             →  DerivedDecl  (definition with = expr)
	p.advance()
	result := p.parseDefnLhs()
	if p.match(lexer.EQ) {
		// relation name(args) = expr → DerivedDecl (matches Python)
		body := p.parseExpr(0)
		defn := ast.NewDefinition(result, body)
		lf := ast.NewLabeledFormula(nil, defn)
		return p.setLoc(ast.NewDerivedDecl(lf), tok)
	}
	return p.setLoc(ast.NewConstantDecl(result), tok)
}

// parseDefnLhs parses a definition left-hand side.
// Matches Python's defnlhs grammar exactly:
//
//	defnlhs : dotsym                             → Atom(name)
//	        | dotsym LPAREN defargs RPAREN        → Atom(name, args...)
//	        | LPAREN defarg relop defarg RPAREN   → Atom(op, [arg1, arg2])
//	        | LPAREN defarg infix defarg RPAREN   → Atom(op, [arg1, arg2])
func (p *Parser) parseDefnLhs() ast.Node {
	if p.match(lexer.LPAREN) {
		// Infix form: (defarg op defarg)
		arg1 := p.parseDefArg()
		opTok := p.current
		var opName string
		switch opTok.Type {
		case lexer.LT:
			opName = "<"
		case lexer.GT:
			opName = ">"
		case lexer.LE:
			opName = "<="
		case lexer.GE:
			opName = ">="
		case lexer.TILDAEQ:
			opName = "~="
		case lexer.PLUS:
			opName = "+"
		case lexer.MINUS:
			opName = "-"
		case lexer.TIMES:
			opName = "*"
		case lexer.DIV:
			opName = "/"
		default:
			p.errorf("expected operator in infix declaration, got %s", opTok.Type)
			return arg1
		}
		p.advance()
		arg2 := p.parseDefArg()
		p.expect(lexer.RPAREN)
		result := ast.NewAtom(opName, arg1, arg2)
		p.setLoc(result, opTok)
		return result
	}

	// Standard form: dotsym or dotsym(defargs)
	tok := p.current
	name, _ := p.parseAtomName()
	// Handle dotted names: a.b.c
	for p.match(lexer.DOT) {
		name2, _ := p.parseAtomName()
		name = name + "." + name2
	}
	result := ast.NewAtom(name)
	p.setLoc(result, tok)
	if p.match(lexer.LPAREN) {
		// dotsym(defargs)
		args := p.parseDefArgs()
		p.expect(lexer.RPAREN)
		result = ast.NewAtom(name, args...)
		p.setLoc(result, tok)
	}
	// Check for sort annotation: name : sort
	if p.match(lexer.COLON) {
		result.ASort = p.parseAType()
	}
	return result
}

// parseDefArg parses a definition argument.
// Matches Python's defarg grammar:
//
//	defarg : lparam     →  SYMBOL COLON atype
//	       | var        →  VARIABLE [COLON atype]
func (p *Parser) parseDefArg() ast.Node {
	tok := p.current
	// Python: lparam : CARET SYMBOL COLON atype → KeyArg(symbol, sort)
	if tok.Type == lexer.CARET {
		p.advance()
		nameTok := p.expect(lexer.SYMBOL)
		a := ast.NewAtom(nameTok.Value)
		p.setLoc(a, nameTok)
		if p.match(lexer.COLON) {
			a.ASort = p.parseAType()
		}
		ka := &ast.KeyArg{App: ast.NewApp(ast.NewSymbol(nameTok.Value, nil))}
		if a.ASort != nil {
			ka.App.ASort = a.ASort
		}
		p.setLoc(ka, tok)
		return ka
	}
	if tok.Type == lexer.VARIABLE {
		// var: VARIABLE or VARIABLE COLON atype
		p.advance()
		name := tok.Value
		var sort ast.Node
		if p.match(lexer.COLON) {
			sort = p.parseAType()
		}
		v := ast.NewVariable(name, sort)
		p.setLoc(v, tok)
		return v
	}
	// lparam: SYMBOL COLON atype
	name, nameTok := p.parseAtomName()
	a := ast.NewAtom(name)
	p.setLoc(a, nameTok)
	if p.match(lexer.COLON) {
		a.ASort = p.parseAType()
	}
	return a
}

// parseDefArgs parses comma-separated definition arguments.
func (p *Parser) parseDefArgs() []ast.Node {
	var args []ast.Node
	args = append(args, p.parseDefArg())
	for p.match(lexer.COMMA) {
		args = append(args, p.parseDefArg())
	}
	return args
}

func (p *Parser) parseConstantDecl(tok lexer.Token) ast.Node {
	p.advance()
	// Python: constantdecl : INDIV tterms → ConstantDecl(*tterms)
	terms := p.parseTTermList()
	return p.setLoc(ast.NewConstantDecl(terms...), tok)
}

// parseFunctionDeclMulti parses: function funs (comma-separated)
// Python: top : top FUNCTION funs → for d in funs: declare(d)
func (p *Parser) parseFunctionDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	var result []ast.Node
	for {
		fun := p.parseOneFun(tok)
		result = append(result, fun)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	return result
}

// parseOneFun parses a single function declaration: typeddefn [= defnrhs]
func (p *Parser) parseOneFun(tok lexer.Token) ast.Node {
	result := p.parseDefnLhs()
	// Check for ": return_type" suffix
	if p.match(lexer.COLON) {
		retSort := p.parseAType()
		if a, ok := result.(*ast.Atom); ok {
			a.ASort = retSort
		}
	}
	// Check for "= definition" → DerivedDecl
	if p.match(lexer.EQ) {
		body := p.parseExpr(0)
		defn := ast.NewDefinition(result, body)
		p.setLoc(defn, tok)
		lf := ast.NewLabeledFormula(nil, defn)
		return p.setLoc(ast.NewDerivedDecl(lf), tok)
	}
	return p.setLoc(ast.NewConstantDecl(result), tok)
}

func (p *Parser) parseFunctionDecl(tok lexer.Token) ast.Node {
	p.advance()
	result := p.parseDefnLhs()
	// Check for ": return_type" suffix
	if p.match(lexer.COLON) {
		retSort := p.parseAType()
		if a, ok := result.(*ast.Atom); ok {
			a.ASort = retSort
		}
	}
	// Check for "= definition" → DerivedDecl (matches Python: fun : typeddefn EQ defnrhs)
	if p.match(lexer.EQ) {
		body := p.parseExpr(0)
		defn := ast.NewDefinition(result, body)
		p.setLoc(defn, tok)
		lf := ast.NewLabeledFormula(nil, defn)
		return p.setLoc(ast.NewDerivedDecl(lf), tok)
	}
	return p.setLoc(ast.NewConstantDecl(result), tok)
}

func (p *Parser) parseAxiomDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewAxiomDecl(lf), tok)
}

func (p *Parser) parseAxiomDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	result := []ast.Node{p.setLoc(ast.NewAxiomDecl(lf), tok)}
	if p.match(lexer.PROOF) {
		proofBody := p.parseProofBody()
		result = append(result, p.setLoc(ast.NewProofDecl(proofBody), tok))
	}
	return result
}

func (p *Parser) parsePropertyDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	if p.match(lexer.NAMED) {
		_ = p.parseDefnLhs()
	}
	if p.match(lexer.PROOF) {
		_ = p.parseProofBody()
	}
	return p.setLoc(ast.NewPropertyDecl(lf), tok)
}

// parsePropertyDeclMulti returns PropertyDecl + optional NamedDecl + optional ProofDecl,
// matching Python which emits all as separate top-level declarations.
func (p *Parser) parsePropertyDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	result := []ast.Node{p.setLoc(ast.NewPropertyDecl(lf), tok)}
	// Optional "named" skolemization → emit separate NamedDecl
	if p.match(lexer.NAMED) {
		skolemName := p.parseDefnLhs()
		result = append(result, p.setLoc(ast.NewNamedDecl(skolemName), tok))
	}
	if p.match(lexer.PROOF) {
		proofBody := p.parseProofBody()
		result = append(result, p.setLoc(ast.NewProofDecl(proofBody), tok))
	}
	return result
}

func (p *Parser) parseConjectureDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	return p.setLoc(ast.NewConjectureDecl(lf), tok)
}

func (p *Parser) parseConjectureDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	result := []ast.Node{p.setLoc(ast.NewConjectureDecl(lf), tok)}
	if p.match(lexer.PROOF) {
		proofBody := p.parseProofBody()
		result = append(result, p.setLoc(ast.NewProofDecl(proofBody), tok))
	}
	return result
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
		if p.match(lexer.TIMES) {
			// Python: optactiondef : EQ TIMES → CrashAction()
			body = ast.NewCrashAction(ast.NewAtom(name, params...))
			p.setLoc(body, tok)
		} else {
			body = p.parseActionBody()
		}
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
	p.expect(lexer.EQ)
	p.expect(lexer.LCB)
	body, _ := p.parseBlock()
	p.expect(lexer.RCB)

	// Build the name node with params as args
	nameWithParams := name
	if a, ok := name.(*ast.Atom); ok && len(params) > 0 {
		nameWithParams = ast.NewAtom(a.Rep, params...)
	}
	defn := ast.NewDefinition(nameWithParams, blockToNode(body))
	md := ast.NewModuleDecl(defn)
	p.setLoc(md, tok)

	// Register for later instantiation (matches Python's stack)
	var modName string
	if a, ok := name.(*ast.Atom); ok {
		modName = a.Rep
	}
	if modName != "" {
		md.FormalParams = params
		md.BodyDecls = body
		p.modules[modName] = md
	}
	return md
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

// parseObjectDeclMulti parses "object name = { decls }" and produces:
// 1. An ObjectDecl node
// 2. All inner declarations with names prefixed by the object name
// This matches Python's create_object + inst_mod behavior.
func (p *Parser) parseObjectDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	name := p.parseCallatom()
	var objectArgs []ast.Node
	if p.match(lexer.LPAREN) {
		objectArgs = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}
	_ = objectArgs
	p.expect(lexer.EQ)
	p.expect(lexer.LCB)
	// Check for continuation: object name = { ... body }
	// Python: optdotdotdot : DOTDOTDOT → continuation=True
	continuation := p.match(lexer.DOTDOTDOT)
	innerDecls, _ := p.parseBlock()
	p.expect(lexer.RCB)

	var nameStr string
	if a, ok := name.(*ast.Atom); ok {
		nameStr = a.Rep
	} else {
		nameStr = fmt.Sprint(name)
	}

	var result []ast.Node

	// Python: if not continuation: top.declare(ObjectDecl(pref))
	if !continuation {
		pref := ast.NewAtom(nameStr)
		p.setLoc(pref, tok)
		objDecl := ast.NewObjectDecl(pref)
		p.setLoc(objDecl, tok)
		result = append(result, objDecl)
	}

	// Python: inst_mod(top, module, pref, {}, vsubst)
	// Use ast.AstRewrite with AstRewriteSubstPrefix — exactly matching Python.
	pref := ast.NewAtom(nameStr)
	p.setLoc(pref, tok)
	defined := collectDefinedNames(innerDecls)
	for _, decl := range innerDecls {
		idecl := ast.SubstPrefixAtomsAst(decl, nil, pref, defined, nil)
		result = append(result, idecl)
	}

	return result
}

// prefixDeclNames prefixes all declaration names in a node with "prefix.".
// This matches Python's subst_prefix_atoms_ast behavior.
// IMPORTANT: This clones the declaration to avoid mutating shared module body data.
func prefixDeclNames(decl ast.Node, prefix string) []ast.Node {
	// Clone the declaration first so we don't mutate shared data
	// (modules can be instantiated multiple times).
	decl = decl.Clone(decl.Args())

	pname := func(name string) string {
		return prefix + "." + name
	}

	switch n := decl.(type) {
	case *ast.ConstantDecl:
		for i, arg := range n.DeclArgs {
			if a, ok := arg.(*ast.Atom); ok {
				pa := a.Prefix(prefix + ".")
				n.DeclArgs[i] = pa
			}
		}
		return []ast.Node{n}

	case *ast.ActionDecl:
		for i, arg := range n.DeclArgs {
			if ad, ok := arg.(*ast.ActionDef); ok {
				// Deep clone the ActionDef to avoid mutating shared data
				newAd := *ad
				if a, ok := ad.Name.(*ast.Atom); ok {
					newAd.Name = ast.NewAtom(pname(a.Rep))
				}
				n.DeclArgs[i] = &newAd
			}
		}
		return []ast.Node{n}

	case *ast.MixinDecl:
		for i, arg := range n.DeclArgs {
			switch m := arg.(type) {
			case *ast.MixinAfterDef:
				newM := *m
				if a, ok := m.MixerNode.(*ast.Atom); ok {
					newM.MixerNode = ast.NewAtom(pname(a.Rep))
				}
				n.DeclArgs[i] = &newM
			case *ast.MixinBeforeDef:
				newM := *m
				if a, ok := m.MixerNode.(*ast.Atom); ok {
					newM.MixerNode = ast.NewAtom(pname(a.Rep))
				}
				n.DeclArgs[i] = &newM
			case *ast.MixinImplementDef:
				newM := *m
				if a, ok := m.MixerNode.(*ast.Atom); ok {
					newM.MixerNode = ast.NewAtom(pname(a.Rep))
				}
				n.DeclArgs[i] = &newM
			}
		}
		return []ast.Node{n}

	case *ast.TypeDecl:
		for i, arg := range n.DeclArgs {
			if td, ok := arg.(*ast.TypeDef); ok {
				// Deep clone TypeDef to avoid mutating shared data
				newTd := *td
				// Python: compose_atoms(prefix, atom) — if atom.rep is This, use just prefix
				if sym, ok := td.Name.(*ast.Symbol); ok {
					if sym.Rep == "this" {
						newTd.Name = ast.NewSymbol(prefix, sym.Sort)
					} else {
						newTd.Name = ast.NewSymbol(pname(sym.Rep), sym.Sort)
					}
				} else if a, ok := td.Name.(*ast.Atom); ok {
					if a.Rep == "this" {
						newTd.Name = ast.NewAtom(prefix)
					} else {
						newTd.Name = ast.NewAtom(pname(a.Rep))
					}
				}
				n.DeclArgs[i] = &newTd
			}
		}
		return []ast.Node{n}

	case *ast.ObjectDecl:
		if len(n.DeclArgs) > 0 {
			if a, ok := n.DeclArgs[0].(*ast.Atom); ok {
				n.DeclArgs[0] = ast.NewAtom(pname(a.Rep))
			}
		}
		return []ast.Node{n}

	case *ast.IsolateObjectDecl:
		return []ast.Node{n}

	case *ast.TheoremDecl:
		// Prefix theorem label if present
		return []ast.Node{n}

	case *ast.VariantDecl:
		// Prefix variant name (this → prefix)
		for i, arg := range n.DeclArgs {
			if vd, ok := arg.(*ast.VariantDef); ok {
				newVd := *vd
				if a, ok := vd.Name.(*ast.Atom); ok {
					if a.Rep == "this" {
						newVd.Name = ast.NewAtom(prefix)
					} else {
						newVd.Name = ast.NewAtom(pname(a.Rep))
					}
				}
				n.DeclArgs[i] = &newVd
			}
		}
		return []ast.Node{n}

	case *ast.DefinitionDecl:
		return []ast.Node{n}

	case *ast.InterpretDecl:
		return []ast.Node{n}

	case *ast.ProofDecl:
		return []ast.Node{n}

	case *ast.ConjectureDecl, *ast.PropertyDecl, *ast.AxiomDecl:
		return []ast.Node{decl}

	case *ast.InitDecl:
		return []ast.Node{decl}

	case *ast.InstantiateDecl:
		return []ast.Node{decl}

	default:
		return []ast.Node{decl}
	}
}

// parseIsolateDeclMulti parses "isolate name = { decls } [with args]" and
// produces the same AST as Python:
//  1. ObjectDecl(name)
//  2. ...inlined inner declarations with prefixed names...
//  3. IsolateObjectDecl
//
// This matches Python's create_object + IsolateObjectDecl behavior.
func (p *Parser) parseIsolateDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	ca := p.parseCallatom()

	var nameStr string
	if a, ok := ca.(*ast.Atom); ok {
		nameStr = a.Rep
	} else {
		nameStr = fmt.Sprint(ca)
	}

	if p.match(lexer.EQ) {
		if p.at(lexer.LCB) {
			// Rule 3: isolate NAME = { body } optwith → object inlining + IsolateObjectDecl
			p.advance() // consume LCB
			innerDecls, _ := p.parseBlock()
			p.expect(lexer.RCB)

			// Parse optional "with" clause
			var withElems []ast.Node
			if p.match(lexer.WITH) {
				for {
					withElems = append(withElems, p.parseCallatom())
					if !p.match(lexer.COMMA) {
						break
					}
				}
			}

			// 1. ObjectDecl (Python emits ObjectDecl first, not IsolateDecl)
			pref := ast.NewAtom(nameStr)
			p.setLoc(pref, tok)
			objDecl := ast.NewObjectDecl(pref)
			p.setLoc(objDecl, tok)
			result := []ast.Node{objDecl}

			// 2. Inline inner declarations with prefixed names
			// Python: inst_mod → subst_prefix_atoms_ast
			defined := collectDefinedNames(innerDecls)
			for _, decl := range innerDecls {
				idecl := ast.SubstPrefixAtomsAst(decl, nil, pref, defined, nil)
				result = append(result, idecl)
			}

			// 3. IsolateObjectDecl at the end (Python: IsolateObjectDecl)
			isoElems := []ast.Node{ca, ca}
			if len(withElems) > 0 {
				isoElems = append(isoElems, withElems...)
			}
			baseIso := ast.NewIsolateDecl(&ast.IsolateDef{
				Elems:    isoElems,
				WithArgs: len(withElems),
			})
			isoDecl := &ast.IsolateObjectDecl{IsolateDecl: *baseIso}
			p.setLoc(isoDecl, tok)
			result = append(result, isoDecl)

			return result
		}

		// Rule 1/2: isolate NAME = callatoms [WITH callatoms] → simple IsolateDecl
		elems := []ast.Node{ca}
		// Parse the callatoms after =
		for {
			elems = append(elems, p.parseCallatom())
			if !p.match(lexer.COMMA) {
				break
			}
		}
		withArgs := 0
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
		return []ast.Node{p.setLoc(ast.NewIsolateDecl(idef), tok)}
	}

	// "isolate name with a, b, c" (no body)
	elems := []ast.Node{ca}
	withArgs := 0
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
	return []ast.Node{p.setLoc(ast.NewIsolateDecl(idef), tok)}
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
		// Check for "with" clause after closing brace: "} with node, id, trans"
		elems = append(elems, blockToNode(body))
		if p.match(lexer.WITH) {
			for {
				elems = append(elems, p.parseCallatom())
				withArgs++
				if !p.match(lexer.COMMA) {
					break
				}
			}
		}
		return p.setLoc(ast.NewIsolateDecl(&ast.IsolateDef{Elems: elems, WithArgs: withArgs}), tok)
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
	return p.setLoc(ast.NewExportDecl(&ast.ExportDef{ExportedNode: ca, ScopeNode: &ast.NoneAST{}}), tok)
}

// parseExportDeclMulti handles "export name" and "export action name(...) = {...}"
// The latter produces ActionDecl + ExportDecl (matching Python).
func (p *Parser) parseExportDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	if p.at(lexer.ACTION) {
		// "export action name(...) = {...}" — parse the action, then add export
		actionDecls := p.parseTopLevel() // re-enter parseTopLevel for "action ..."
		var result []ast.Node
		result = append(result, actionDecls...)
		// Add ExportDecl for the action name
		for _, d := range actionDecls {
			if ad, ok := d.(*ast.ActionDecl); ok {
				if len(ad.Args()) > 0 {
					if adef, ok := ad.Args()[0].(*ast.ActionDef); ok {
						ca := ast.NewAtom(adef.Defines())
						result = append(result, p.setLoc(ast.NewExportDecl(&ast.ExportDef{ExportedNode: ca, ScopeNode: &ast.NoneAST{}}), tok))
					}
				}
			}
		}
		return result
	}
	ca := p.parseCallatom()
	return []ast.Node{p.setLoc(ast.NewExportDecl(&ast.ExportDef{ExportedNode: ca, ScopeNode: &ast.NoneAST{}}), tok)}
}

func (p *Parser) parseImportDecl(tok lexer.Token) ast.Node {
	p.advance()
	ca := p.parseCallatom()
	return p.setLoc(ast.NewImportDecl(&ast.ImportDef{Imported: ca, Scope: &ast.NoneAST{}}), tok)
}

// parseImportDeclMulti handles "import name" and "import action name(...)"
// The latter produces ActionDecl + ImportDecl (matching Python).
func (p *Parser) parseImportDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	if p.at(lexer.ACTION) {
		// "import action name(...)" — parse the action, then add import
		actionDecls := p.parseTopLevel() // re-enter parseTopLevel for "action ..."
		var result []ast.Node
		result = append(result, actionDecls...)
		// Add ImportDecl for the action name
		for _, d := range actionDecls {
			if ad, ok := d.(*ast.ActionDecl); ok {
				if len(ad.Args()) > 0 {
					if adef, ok := ad.Args()[0].(*ast.ActionDef); ok {
						ca := ast.NewAtom(adef.Defines())
						result = append(result, p.setLoc(ast.NewImportDecl(&ast.ImportDef{Imported: ca, Scope: &ast.NoneAST{}}), tok))
					}
				}
			}
		}
		return result
	}
	ca := p.parseCallatom()
	return []ast.Node{p.setLoc(ast.NewImportDecl(&ast.ImportDef{Imported: ca, Scope: &ast.NoneAST{}}), tok)}
}

// parseInstantiateDeclMulti parses "instantiate modname(args)" and expands
// known module definitions inline, matching Python's do_insts/inst_mod.
// If the module is not found, falls back to emitting InstantiateDecl.
func (p *Parser) parseInstantiateDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	var insts []ast.Node
	for {
		var label ast.Node
		ca := p.parseModInst()
		if p.match(lexer.COLON) {
			label = ca
			ca = p.parseModInst()
		}
		inst := ast.NewInstantiation(label, ca)
		p.setLoc(inst, tok)
		insts = append(insts, inst)
		if !p.match(lexer.COMMA) {
			break
		}
	}

	// Try to expand each instantiation
	var result []ast.Node
	var unexpanded []ast.Node

	for _, inst := range insts {
		instNode, ok := inst.(*ast.Instantiation)
		if !ok {
			unexpanded = append(unexpanded, inst)
			continue
		}
		// Get the callatom (module name + args)
		// Instantiation.Sort is the module call, Instantiation.Name is the prefix/label
		ca := instNode.Sort
		var modName string
		var actualArgs []ast.Node
		if a, ok := ca.(*ast.Atom); ok {
			modName = a.Rep
			actualArgs = a.Terms
		}

		modDef, found := p.modules[modName]
		if !found || modDef == nil {
			unexpanded = append(unexpanded, inst)
			continue
		}

		// Build substitutions: formal param name → actual arg name
		// Python separates into subst (constant args) and vsubst (variable args)
		formalParams := modDef.FormalParams
		subst := make(map[string]string)
		vsubst := make(map[string]*ast.Variable)
		for i := 0; i < len(formalParams) && i < len(actualArgs); i++ {
			var formalName string
			if a, ok := formalParams[i].(*ast.Atom); ok {
				formalName = a.Rep
			} else if s, ok := formalParams[i].(*ast.Symbol); ok {
				formalName = s.Rep
			}
			if formalName == "" {
				continue
			}
			// Check if actual arg is a variable (goes to vsubst)
			if v, ok := actualArgs[i].(*ast.Variable); ok {
				vsubst[formalName] = v
			} else {
				var actualName string
				if a, ok := actualArgs[i].(*ast.Atom); ok {
					actualName = a.Rep
				} else if s, ok := actualArgs[i].(*ast.Symbol); ok {
					actualName = s.Rep
				} else {
					actualName = fmt.Sprint(actualArgs[i])
				}
				subst[formalName] = actualName
			}
		}

		// Get prefix from label (e.g., "instance foo : bar(t)" → prefix "foo")
		prefix := ""
		if instNode.Name != nil {
			if a, ok := instNode.Name.(*ast.Atom); ok {
				prefix = a.Rep
			}
		}

		// If there's a prefix, emit ObjectDecl first (Python: ivy.declare(ObjectDecl(pref)))
		if prefix != "" {
			pref := ast.NewAtom(prefix)
			p.setLoc(pref, tok)
			result = append(result, ast.NewObjectDecl(pref))
		}

		// Python: inst_mod → subst_prefix_atoms_ast(decl, subst, pref, module.defined, static)
		var pref *ast.Atom
		if prefix != "" {
			pref = ast.NewAtom(prefix)
			p.setLoc(pref, tok)
		}
		defined := collectDefinedNames(modDef.BodyDecls)

		// Build static set: names defined as types or destructors
		static := make(map[string]bool)
		for name := range defined {
			// Mark type and destructor definitions as static
			static[name] = true
		}

		for _, bodyDecl := range modDef.BodyDecls {
			idecl := ast.SubstPrefixAtomsAst(bodyDecl, subst, pref, defined, static)
			// Apply variable substitution if any
			if len(vsubst) > 0 {
				vsub := make(map[string]ast.Node)
				for k, v := range vsubst {
					vsub[k] = v
				}
				idecl = ast.SubstituteConstantsAst(idecl, vsub)
			}
			result = append(result, idecl)
		}
	}

	if len(unexpanded) > 0 {
		result = append(result, p.setLoc(ast.NewInstantiateDecl(unexpanded...), tok))
	}
	return result
}

// collectDefinedNames collects the names defined by a list of declarations.
// This matches Python's module.defined — a set of names that the module defines,
// used by AstRewriteSubstPrefix to know which names to prefix.
func collectDefinedNames(decls []ast.Node) map[string]bool {
	defined := make(map[string]bool)
	for _, d := range decls {
		switch n := d.(type) {
		case *ast.ActionDecl:
			for _, arg := range n.DeclArgs {
				if ad, ok := arg.(*ast.ActionDef); ok {
					if a, ok := ad.Name.(*ast.Atom); ok {
						defined[a.Rep] = true
					}
				}
			}
		case *ast.ConstantDecl:
			for _, arg := range n.DeclArgs {
				if a, ok := arg.(*ast.Atom); ok {
					defined[a.Rep] = true
				}
			}
		case *ast.TypeDecl:
			for _, arg := range n.DeclArgs {
				if td, ok := arg.(*ast.TypeDef); ok {
					if sym, ok := td.Name.(*ast.Symbol); ok {
						defined[sym.Rep] = true
					} else if a, ok := td.Name.(*ast.Atom); ok {
						defined[a.Rep] = true
					}
				}
			}
		case *ast.ObjectDecl:
			for _, arg := range n.DeclArgs {
				if a, ok := arg.(*ast.Atom); ok {
					defined[a.Rep] = true
				}
			}
		case *ast.MixinDecl:
			for _, arg := range n.DeclArgs {
				switch m := arg.(type) {
				case *ast.MixinAfterDef:
					if a, ok := m.MixerNode.(*ast.Atom); ok {
						defined[a.Rep] = true
					}
				case *ast.MixinBeforeDef:
					if a, ok := m.MixerNode.(*ast.Atom); ok {
						defined[a.Rep] = true
					}
				case *ast.MixinImplementDef:
					if a, ok := m.MixerNode.(*ast.Atom); ok {
						defined[a.Rep] = true
					}
				}
			}
		case *ast.DerivedDecl:
			for _, arg := range n.DeclArgs {
				if lf, ok := arg.(*ast.LabeledFormula); ok {
					if def, ok := lf.Formula.(*ast.Definition); ok {
						if a, ok := def.Lhs.(*ast.Atom); ok {
							defined[a.Rep] = true
						}
					}
				}
			}
		case *ast.ConjectureDecl, *ast.PropertyDecl, *ast.AxiomDecl:
			// labeled formulas — labels define names
		case *ast.IsolateDecl, *ast.IsolateObjectDecl:
			// isolate names
		}
	}
	// Add "this" — always in scope
	defined["this"] = true
	return defined
}

// deepCloneNode recursively deep-clones an AST node to ensure complete isolation.
func deepCloneNode(node ast.Node) ast.Node {
	if node == nil {
		return nil
	}
	args := node.Args()
	if len(args) == 0 {
		// Leaf node — clone just the node itself
		return node.Clone(nil)
	}
	newArgs := make([]ast.Node, len(args))
	for i, a := range args {
		newArgs[i] = deepCloneNode(a)
	}
	return node.Clone(newArgs)
}

// substituteNamesInDecl replaces formal parameter names with actual argument
// names in a declaration. This is a simplified version of Python's
// subst_prefix_atoms_ast for module instantiation.
func substituteNamesInDecl(decl ast.Node, subst map[string]string) ast.Node {
	// Always clone to avoid mutating shared module body data.
	// Even when subst is empty, we clone because prefixDeclNames
	// will mutate the result.
	return substNamesAST(decl, subst)
}

// substNamesAST recursively substitutes names in an AST node.
func substNamesAST(node ast.Node, subst map[string]string) ast.Node {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.Symbol:
		if repl, ok := subst[n.Rep]; ok {
			return ast.NewSymbol(repl, n.Sort)
		}
		return n
	case *ast.Atom:
		newTerms := make([]ast.Node, len(n.Terms))
		for i, t := range n.Terms {
			newTerms[i] = substNamesAST(t, subst)
		}
		rep := n.Rep
		if repl, ok := subst[rep]; ok {
			rep = repl
		}
		result := ast.NewAtom(rep, newTerms...)
		result.ASort = n.ASort
		return result
	case *ast.Variable:
		rep := n.Rep
		if repl, ok := subst[rep]; ok {
			rep = repl
		}
		sort := n.VSort
		if sort != nil {
			sort = substNamesAST(sort, subst)
		}
		return ast.NewVariable(rep, sort)
	default:
		// For complex nodes, clone with substituted children
		args := node.Args()
		if len(args) == 0 {
			return node
		}
		newArgs := make([]ast.Node, len(args))
		for i, a := range args {
			newArgs[i] = substNamesAST(a, subst)
		}
		return node.Clone(newArgs)
	}
}

// parseModInst parses a module instantiation: dotsym or dotsym(pnames).
// Matches Python's modinst grammar.
func (p *Parser) parseModInst() ast.Node {
	tok := p.current
	name, _ := p.parseAtomName()
	for p.match(lexer.DOT) {
		name2, _ := p.parseAtomName()
		name = name + "." + name2
	}
	result := ast.NewAtom(name)
	p.setLoc(result, tok)
	if p.match(lexer.LPAREN) {
		args := p.parsePNames()
		p.expect(lexer.RPAREN)
		result = ast.NewAtom(name, args...)
		p.setLoc(result, tok)
	}
	return result
}

// parsePNames parses comma-separated pname arguments (possibly empty).
// Matches Python's pnames grammar: pname accepts atype, var, infix, relop, THIS, TRUE, FALSE.
func (p *Parser) parsePNames() []ast.Node {
	if p.at(lexer.RPAREN) {
		return nil // empty args
	}
	var args []ast.Node
	args = append(args, p.parsePName())
	for p.match(lexer.COMMA) {
		args = append(args, p.parsePName())
	}
	return args
}

// parsePName parses a single module parameter name.
// Matches Python's pname grammar:
//
//	pname : atype | var | infix | relop | THIS | TRUE | FALSE
func (p *Parser) parsePName() ast.Node {
	tok := p.current
	switch tok.Type {
	case lexer.VARIABLE:
		// var: VARIABLE [COLON atype]
		p.advance()
		var sort ast.Node
		if p.match(lexer.COLON) {
			sort = p.parseAType()
		}
		v := ast.NewVariable(tok.Value, sort)
		p.setLoc(v, tok)
		return v
	case lexer.THIS:
		p.advance()
		a := ast.NewAtom("this")
		p.setLoc(a, tok)
		return a
	// Relops: <, <=, >, >=, ~=
	case lexer.LT:
		p.advance()
		return ast.NewAtom("<")
	case lexer.LE:
		p.advance()
		return ast.NewAtom("<=")
	case lexer.GT:
		p.advance()
		return ast.NewAtom(">")
	case lexer.GE:
		p.advance()
		return ast.NewAtom(">=")
	case lexer.TILDAEQ:
		p.advance()
		return ast.NewAtom("~=")
	// Infix: +, -, *, /
	case lexer.PLUS:
		p.advance()
		return ast.NewAtom("+")
	case lexer.MINUS:
		p.advance()
		return ast.NewAtom("-")
	case lexer.TIMES:
		p.advance()
		return ast.NewAtom("*")
	case lexer.DIV:
		p.advance()
		return ast.NewAtom("/")
	default:
		// atype: SYMBOL [. SYMBOL]* [LB subscr RB]
		return p.parseAType()
	}
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
		return p.setLoc(ast.NewMixinDecl(&ast.MixinBeforeDef{MixerNode: ca, MixeeNode: mixee}), tok)
	}
	if p.match(lexer.AFTER) {
		mixee := p.parseCallatom()
		return p.setLoc(ast.NewMixinDecl(&ast.MixinAfterDef{MixerNode: ca, MixeeNode: mixee}), tok)
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
		mdef = &ast.MixinBeforeDef{MixerNode: mixer, MixeeNode: ca}
	case "after":
		mdef = &ast.MixinAfterDef{MixerNode: mixer, MixeeNode: ca}
	default:
		mdef = &ast.MixinImplementDef{MixerNode: mixer, MixeeNode: ca}
	}
	mixinDecl := ast.NewMixinDecl(mdef)
	p.setLoc(mixinDecl, tok)

	return []ast.Node{actionDecl, mixinDecl}
}

// parseAroundDecl parses: around atype optargs optreturns { actseq ... actseq }
// Python: top : top AROUND atype optargs optreturns LCB actseq optsemi DOTDOTDOT actseq optsemi RCB
// Produces two mixin pairs: before + after, each with ActionDecl + MixinDecl.
func (p *Parser) parseAroundDecl(tok lexer.Token) []ast.Node {
	p.advance()
	ca := p.parseCallatom()

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

	// Parse body: { before_actions ... after_actions }
	p.expect(lexer.LCB)
	beforeBody := p.parseActionSeq()
	p.expect(lexer.DOTDOTDOT)
	afterBody := p.parseActionSeq()
	p.expect(lexer.RCB)

	var result []ast.Node
	mixerBase := ca.String()

	// Before mixin: handle_before_after("before", atom, before, ...)
	p.labelCounter++
	beforeMixerName := mixerBase + "[before" + fmt.Sprintf("%d", p.labelCounter) + "]"
	beforeMixer := ast.NewAtom(beforeMixerName)
	p.setLoc(beforeMixer, tok)
	beforeAdef := ast.NewActionDef(beforeMixer, beforeBody, params, returns)
	p.setLoc(beforeAdef, tok)
	beforeActionDecl := ast.NewActionDecl(beforeAdef)
	p.setLoc(beforeActionDecl, tok)
	beforeMdef := &ast.MixinBeforeDef{MixerNode: beforeMixer, MixeeNode: ca}
	beforeMixinDecl := ast.NewMixinDecl(beforeMdef)
	p.setLoc(beforeMixinDecl, tok)
	result = append(result, beforeActionDecl, beforeMixinDecl)

	// After mixin: handle_before_after("after", atom, after, ...)
	p.labelCounter++
	afterMixerName := mixerBase + "[after" + fmt.Sprintf("%d", p.labelCounter) + "]"
	afterMixer := ast.NewAtom(afterMixerName)
	p.setLoc(afterMixer, tok)
	afterAdef := ast.NewActionDef(afterMixer, afterBody, params, returns)
	p.setLoc(afterAdef, tok)
	afterActionDecl := ast.NewActionDecl(afterAdef)
	p.setLoc(afterActionDecl, tok)
	afterMdef := &ast.MixinAfterDef{MixerNode: afterMixer, MixeeNode: ca}
	afterMixinDecl := ast.NewMixinDecl(afterMdef)
	p.setLoc(afterMixinDecl, tok)
	result = append(result, afterActionDecl, afterMixinDecl)

	return result
}

func (p *Parser) parseImplementDecl(tok lexer.Token) ast.Node {
	p.advance()
	ca := p.parseCallatom()
	if p.at(lexer.LCB) || p.at(lexer.LPAREN) {
		var params []ast.Node
		if p.match(lexer.LPAREN) {
			params = p.parseTTermList()
			p.expect(lexer.RPAREN)
		}
		body := p.parseActionBody()
		adef := ast.NewActionDef(ca, body, params, nil)
		mdef := &ast.MixinImplementDef{MixerNode: adef, MixeeNode: ca}
		return p.setLoc(ast.NewMixinDecl(mdef), tok)
	}
	return p.setLoc(ast.NewMixinDecl(&ast.MixinImplementDef{MixerNode: ca, MixeeNode: ca}), tok)
}

// parseImplementDeclMulti parses "implement name { body }" and produces
// both ActionDecl and MixinDecl, matching Python's handle_before_after
// for the "implement" kind.
func (p *Parser) parseImplementDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	ca := p.parseCallatom()

	// "implement type T with S" — different construct, single node
	if p.at(lexer.TYPE) {
		return []ast.Node{p.setLoc(ast.NewMixinDecl(&ast.MixinImplementDef{MixerNode: ca, MixeeNode: ca}), tok)}
	}

	if p.at(lexer.LCB) || p.at(lexer.LPAREN) {
		var params []ast.Node
		if p.match(lexer.LPAREN) {
			params = p.parseTTermList()
			p.expect(lexer.RPAREN)
		}
		var returns []ast.Node
		if p.match(lexer.RETURNS) {
			p.expect(lexer.LPAREN)
			returns = p.parseTTermList()
			p.expect(lexer.RPAREN)
		}
		body := p.parseActionBody()

		// Generate mixer name matching Python: name[implementN]
		mixerName := ca.String()
		p.labelCounter++
		mixerName = mixerName + "[implement" + fmt.Sprintf("%d", p.labelCounter) + "]"
		mixer := ast.NewAtom(mixerName)
		p.setLoc(mixer, tok)

		// 1. ActionDecl
		adef := ast.NewActionDef(mixer, body, params, returns)
		p.setLoc(adef, tok)
		actionDecl := ast.NewActionDecl(adef)
		p.setLoc(actionDecl, tok)

		// 2. MixinDecl
		mdef := &ast.MixinImplementDef{MixerNode: mixer, MixeeNode: ca}
		mixinDecl := ast.NewMixinDecl(mdef)
		p.setLoc(mixinDecl, tok)

		return []ast.Node{actionDecl, mixinDecl}
	}

	// Bare "implement name" — just a mixin
	return []ast.Node{p.setLoc(ast.NewMixinDecl(&ast.MixinImplementDef{MixerNode: ca, MixeeNode: ca}), tok)}
}

// parseVariantDeclMulti parses: variant NAME of BASE [= SORT]
// Python produces TypeDecl + VariantDecl.
// For "variant NAME of BASE": TypeDecl(TypeDef(NAME, UninterpretedSort())) + VariantDecl(VariantDef(NAME, BASE))
// For "variant NAME of BASE = SORT": TypeDecl(TypeDef(NAME, SORT)) + VariantDecl(VariantDef(NAME, BASE))
func (p *Parser) parseVariantDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	name, _ := p.parseAtomName()
	nameAtom := ast.NewAtom(name)
	p.setLoc(nameAtom, tok)

	p.expect(lexer.OF)
	base := p.parseAType()

	var sort ast.Node
	if p.match(lexer.EQ) {
		sort = p.parseSort()
	} else {
		sort = &ast.ConstantSort{}
	}

	tdfn := ast.NewTypeDef(nameAtom, sort)
	p.setLoc(tdfn, tok)
	typeDecl := ast.NewTypeDecl(tdfn)
	p.setLoc(typeDecl, tok)

	vd := ast.NewVariantDef(nameAtom, ast.NewAtom(fmt.Sprint(base)))
	variantDecl := ast.NewVariantDecl(vd)
	p.setLoc(variantDecl, tok)

	return []ast.Node{typeDecl, variantDecl}
}

// parseVariantDecl remains for backward compat (single return)
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
	// Python: schema schdefn → schdefn : defnlhs EQ schdefnrhs
	// SchemaDecl(Schema(Definition(lhs, rhs)))
	// schdefnrhs can be { schdecls schconc } or a formula
	lhs := p.parseDefnLhs()
	p.expect(lexer.EQ)
	var rhs ast.Node
	if p.at(lexer.LCB) {
		rhs = p.parseSchemaBody()
	} else {
		rhs = p.parseExpr(0)
	}
	defn := ast.NewDefinition(lhs, rhs)
	schema := ast.NewSchema(defn)
	return p.setLoc(ast.NewSchemaDecl(schema), tok)
}

func (p *Parser) parseTheoremDecl(tok lexer.Token) ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	if p.match(lexer.PROOF) {
		_ = p.parseProofBody()
	}
	return p.setLoc(ast.NewTheoremDecl(lf), tok)
}

func (p *Parser) parseTheoremDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	result := []ast.Node{p.setLoc(ast.NewTheoremDecl(lf), tok)}
	if p.match(lexer.PROOF) {
		proofBody := p.parseProofBody()
		result = append(result, p.setLoc(ast.NewProofDecl(proofBody), tok))
	}
	return result
}

func (p *Parser) parseProofDecl(tok lexer.Token) ast.Node {
	p.advance()
	_ = p.parseLabel() // optional label: proof [name] { ... }
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
	// delegate callatoms [-> callatom]
	// Python: DelegateDecl(*[DelegateDef(s, target) for s in callatoms])
	var callatoms []ast.Node
	callatoms = append(callatoms, p.parseCallatom())
	for p.match(lexer.COMMA) {
		callatoms = append(callatoms, p.parseCallatom())
	}
	var target ast.Node
	if p.match(lexer.ARROW) {
		target = p.parseCallatom()
	}
	var defs []ast.Node
	for _, ca := range callatoms {
		elems := []ast.Node{ca}
		if target != nil {
			elems = append(elems, target)
		}
		defs = append(defs, &ast.DelegateDef{Elems: elems})
	}
	return p.setLoc(ast.NewDelegateDecl(defs...), tok)
}

func (p *Parser) parseMixOrdDecl(tok lexer.Token) ast.Node {
	p.advance()
	// mixord callatom -> callatom → MixOrdDecl(Implies(a, b))
	left := p.parseCallatom()
	p.expect(lexer.ARROW)
	right := p.parseCallatom()
	impl := ast.NewImplies(left, right)
	return p.setLoc(ast.NewMixOrdDecl(impl), tok)
}

// parseDerivedDeclMulti parses: derived defn [, defn]*
// Python: DerivedDecl(*[addlabel(mk_lf(x),'def') for x in defns])
// Each defn is: typeddefn = rhs → Definition(lhs, rhs)
// parseDerivedDeclMulti parses: derived defn [, defn]*
// Python: DerivedDecl(*[addlabel(mk_lf(x),'def') for x in defns])
// Each defn is: typeddefn = rhs → Definition(lhs, rhs), wrapped in LabeledFormula
func (p *Parser) parseDerivedDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	var lfs []ast.Node
	for {
		lhs := p.parseDefnLhs()
		p.expect(lexer.EQ)
		rhs := p.parseExpr(0)
		defn := ast.NewDefinition(lhs, rhs)
		lf := ast.NewLabeledFormula(nil, defn)
		lfs = append(lfs, lf)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	dd := ast.NewDerivedDecl(lfs...)
	p.setLoc(dd, tok)
	return []ast.Node{dd}
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
	// Python: parameter : tterm EQ paramval | tterm
	// Parse the typed term (e.g., "max : foo")
	tterm := p.parseTTerm()
	if p.match(lexer.EQ) {
		// parameter tterm = value → ParameterDecl(Definition(tterm, value))
		val := p.parseExpr(0)
		defn := ast.NewDefinition(tterm, val)
		p.setLoc(defn, tok)
		return p.setLoc(ast.NewParameterDecl(defn), tok)
	}
	return p.setLoc(ast.NewParameterDecl(tterm), tok)
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

func (p *Parser) parseInvariantDeclMulti(tok lexer.Token) []ast.Node {
	p.advance()
	lf := p.parseLabeledFmla()
	result := []ast.Node{p.setLoc(ast.NewConjectureDecl(lf), tok)}
	if p.match(lexer.PROOF) {
		proofBody := p.parseProofBody()
		result = append(result, p.setLoc(ast.NewProofDecl(proofBody), tok))
	}
	return result
}

func (p *Parser) parseTemporalDecl(tok lexer.Token) ast.Node {
	p.advance()
	// temporal property ...
	if p.match(lexer.PROPERTY) {
		lf := p.parseLabeledFmla()
		lf.Temporal = ast.BoolPtr(true)
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

// parseAutoInstanceDecl parses "autoinstance pattern : modinst, ..."
// Matches Python: 'top : top AUTOINSTANCE insts' where
// inst : modinst | modinst COLON modinst
func (p *Parser) parseAutoInstanceDecl(tok lexer.Token) ast.Node {
	p.advance()
	var insts []ast.Node
	for {
		ca := p.parseModInst()
		var label ast.Node
		if p.match(lexer.COLON) {
			label = ca
			ca = p.parseModInst()
		}
		inst := ast.NewInstantiation(label, ca)
		p.setLoc(inst, tok)
		insts = append(insts, inst)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	return p.setLoc(ast.NewAutoInstanceDecl(insts...), tok)
}

// parseScenarioDecl parses: SCENARIO LCB sceninit SEMI scentranss RCB
// Python: p_top_scenario_lcb_sceninit_semi_scentranss_rcb (ivy_parser.py:2264-2269)
func (p *Parser) parseScenarioDecl(tok lexer.Token) ast.Node {
	p.advance() // consume SCENARIO
	p.expect(lexer.LCB)

	// sceninit: ARROW places
	initPlaces := p.parseScenInit()

	p.expect(lexer.SEMI)

	// scentranss: list of transitions
	var transs []ast.Node
	for !p.at(lexer.RCB) && !p.at(lexer.EOF) {
		tr := p.parseScenarioTransition()
		if tr != nil {
			transs = append(transs, tr)
		}
	}

	p.expect(lexer.RCB)

	// ScenarioDef(initPlaces, transs...)
	elems := append([]ast.Node{initPlaces}, transs...)
	sdef := &ast.ScenarioDef{Elems: elems}
	p.setLoc(sdef, tok)
	return p.setLoc(ast.NewScenarioDecl(sdef), tok)
}

// parseScenInit parses: ARROW places
// Python: p_sceninit_arrow_places (ivy_parser.py:2213-2216)
func (p *Parser) parseScenInit() *ast.PlaceList {
	arrowTok := p.current
	p.expect(lexer.ARROW)
	places := p.parsePlaces()
	pl := &ast.PlaceList{Elems: places}
	p.setLoc(pl, arrowTok)
	return pl
}

// parsePlaces parses: SYMBOL | places COMMA SYMBOL
// Python: p_places_symbol, p_places_places_comma_symbol (ivy_parser.py:2202-2211)
func (p *Parser) parsePlaces() []ast.Node {
	tok := p.expect(lexer.SYMBOL)
	atom := ast.NewAtom(tok.Value)
	p.setLoc(atom, tok)
	places := []ast.Node{atom}
	for p.match(lexer.COMMA) {
		tok = p.expect(lexer.SYMBOL)
		atom = ast.NewAtom(tok.Value)
		p.setLoc(atom, tok)
		places = append(places, atom)
	}
	return places
}

// parseScenarioTransition parses one transition:
//   places ARROW places COLON scenariomixin
//   places COLON scenariomixin  (no target)
// Python: p_scentranss_scentranss_places_arrow_places_colon_scenariomixin (ivy_parser.py:2250-2262)
func (p *Parser) parseScenarioTransition() *ast.ScenarioTransition {
	fromPlaces := p.parsePlaces()
	from := &ast.PlaceList{Elems: fromPlaces}

	var to *ast.PlaceList
	if p.match(lexer.ARROW) {
		toPlaces := p.parsePlaces()
		to = &ast.PlaceList{Elems: toPlaces}
	} else {
		to = &ast.PlaceList{} // empty PlaceList
	}

	colonTok := p.current
	p.expect(lexer.COLON)

	mixin := p.parseScenarioMixin()

	tr := &ast.ScenarioTransition{From: from, To: to, Action: mixin}
	p.setLoc(tr, colonTok)
	return tr
}

// parseScenarioMixin parses:
//   BEFORE atype optargs optreturns sequence
//   AFTER atype optargs optreturns sequence
// Python: p_scenariomixin_before/after (ivy_parser.py:2225-2244)
func (p *Parser) parseScenarioMixin() ast.Node {
	tok := p.current
	var kind string
	switch tok.Type {
	case lexer.BEFORE:
		kind = "before"
	case lexer.AFTER:
		kind = "after"
	default:
		p.errorf("expected 'before' or 'after' in scenario mixin, got %s", tok.Type)
		return nil
	}
	p.advance()

	// atype — the action name
	actionName := p.parseAType()
	atom := ast.NewAtom(actionName.String())
	p.setLoc(atom, tok)

	// optargs
	var params []ast.Node
	if p.match(lexer.LPAREN) {
		params = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}

	// optreturns
	var returns []ast.Node
	if p.match(lexer.RETURNS) {
		p.expect(lexer.LPAREN)
		returns = p.parseTTermList()
		p.expect(lexer.RPAREN)
	}

	// sequence (action body)
	body := p.parseActionBody()

	// Generate mixer name matching Python's make_mixin_name (ivy_parser.py:2218-2223).
	// v1.6: "a[before]"  (no counter)
	// v1.7+: "a[before1]" (with counter)
	mixerName := atom.String()
	if p.vle(1, 6) {
		mixerName = mixerName + "[" + kind + "]"
	} else {
		p.labelCounter++
		mixerName = mixerName + "[" + kind + fmt.Sprintf("%d", p.labelCounter) + "]"
	}
	mixer := ast.NewAtom(mixerName)
	p.setLoc(mixer, tok)

	// ActionDef for the mixin body
	adef := ast.NewActionDef(atom, body, params, returns)
	p.setLoc(adef, tok)

	switch kind {
	case "before":
		m := &ast.ScenarioBeforeMixin{Mixer: mixer, Def: adef}
		p.setLoc(m, tok)
		return m
	default:
		m := &ast.ScenarioAfterMixin{Mixer: mixer, Def: adef}
		p.setLoc(m, tok)
		return m
	}
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
		savedTok := p.current
		ds := p.parseTopLevel()
		decls = append(decls, ds...)
		// Guard against infinite loops: if no tokens consumed, skip one
		if p.current == savedTok {
			p.advance()
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
// parseUnfoldSpecs parses comma-separated unfold specifications.
// Matches Python: unfspecs : unfspec | unfspecs COMMA unfspec
// where unfspec : callatom renamings
func (p *Parser) parseUnfoldSpecs() []ast.Node {
	var specs []ast.Node
	for {
		name := p.parseCallatom()
		// Optional renamings after each unfold spec
		for p.match(lexer.LT) {
			p.advance() // lhs variable
			p.expect(lexer.DIV)
			p.advance() // rhs variable
			if !p.match(lexer.COMMA) {
				// could be more renamings in the same < >
			}
			p.expect(lexer.GT)
		}
		specs = append(specs, &ast.UnfoldSpec{DefName: name})
		if !p.match(lexer.COMMA) {
			break
		}
	}
	return specs
}

// parseOptRenaming consumes an optional <V1/V2, ...> renaming after a schema name.
// Returns true if a renaming was consumed.
func (p *Parser) parseOptRenaming() bool {
	if !p.match(lexer.LT) {
		return false
	}
	for {
		p.advance() // lhs
		p.expect(lexer.DIV)
		p.advance() // rhs
		if !p.match(lexer.COMMA) {
			break
		}
	}
	p.expect(lexer.GT)
	return true
}

// parseTacticWithList parses the "with" clause of a tactic:
//
//	tacticwithlist : tacticwithelem+
//	tacticwithelem : INVARIANT labeledfmla
//	               | DEFINITION typeddefn EQ fmla
//	               | TRIGGER atype WITH terms
//
// The list continues as long as we see INVARIANT/DEFINITION/TRIGGER tokens.
// Matches Python's tacticwithlist grammar.
func (p *Parser) parseTacticWithList() ast.Node {
	// Optional braces: WITH { list } or WITH list
	braced := p.match(lexer.LCB)

	// Python grammar: tacticwithlistchoice : tacticwithlist | pflets
	// If the first token is INVARIANT/DEFINITION/TRIGGER, parse as TacticWith.
	// Otherwise, parse as TacticLets (pflets: var = expr, var = expr, ...).
	switch p.current.Type {
	case lexer.INVARIANT, lexer.DEFINITION, lexer.TRIGGER:
		// tacticwithlist path
		var elems []ast.Node
		for {
			switch p.current.Type {
			case lexer.INVARIANT:
				p.advance()
				lf := p.parseLabeledFmla()
				elems = append(elems, lf)
			case lexer.DEFINITION:
				p.advance()
				defn := p.parseDefnLhs()
				p.expect(lexer.EQ)
				body := p.parseExpr(0)
				elems = append(elems, ast.NewDefinition(defn, body))
			case lexer.TRIGGER:
				p.advance()
				atype := p.parseAType()
				p.expect(lexer.WITH)
				var terms []ast.Node
				terms = append(terms, p.parseExpr(0))
				for p.match(lexer.COMMA) {
					terms = append(terms, p.parseExpr(0))
				}
				trigger := &ast.Trigger{Terms: append([]ast.Node{atype}, terms...)}
				elems = append(elems, trigger)
			default:
				goto tacticWithDone
			}
		}
	tacticWithDone:
		if braced {
			p.expect(lexer.RCB)
		}
		return &ast.TacticWith{Elems: elems}
	default:
		// pflets path: var = expr [, var = expr]*
		// Corresponds to Python: pflet : var EQ fmla
		var lets []ast.Node
		for {
			lhs := p.parseExpr(0)
			p.expect(lexer.EQ)
			rhs := p.parseExpr(0)
			lets = append(lets, ast.NewAtom("=", lhs, rhs))
			if !p.match(lexer.COMMA) {
				break
			}
		}
		if braced {
			p.expect(lexer.RCB)
		}
		return &ast.TacticLets{Lets: lets}
	}
}

func (p *Parser) parseProofBody() ast.Node {
	tok := p.current
	// Optional label before body: proof [name] { ... }
	_ = p.parseLabel()
	if p.match(lexer.LCB) {
		var steps []ast.Node
		for !p.at(lexer.RCB) && !p.at(lexer.EOF) {
			savedTok := p.current
			step := p.parseProofStep()
			if step != nil {
				steps = append(steps, step)
			}
			p.match(lexer.SEMI)
			// Guard against infinite loops: if no tokens were consumed, skip one
			if p.current == savedTok {
				p.advance()
			}
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
		// Optional "with" clause: apply schema with X=val, Y=val, ...
		// Matches Python: 'proofstep : APPLY atype optrenaming'
		// + 'proofstep : APPLY atype optrenaming WITH matches'
		var ren ast.Node = &ast.NoneAST{}
		if p.match(lexer.WITH) {
			var matches []ast.Node
			for {
				m := p.parseExpr(0) // parse X=Z or f(X)=X:t+1
				matches = append(matches, m)
				if !p.match(lexer.COMMA) {
					break
				}
			}
			if len(matches) > 0 {
				ren = ast.NewAnd(matches...)
			}
		}
		return p.setLoc(&ast.SchemaInstantiation{SchemaName: schema, Ren: ren}, tok)
	case lexer.SHOWGOALS:
		p.advance()
		return p.setLoc(&ast.ShowGoalsTactic{}, tok)
	case lexer.DEFERGOAL:
		// Matches Python: 'proofstep : DEFERGOAL'
		p.advance()
		return p.setLoc(&ast.DeferGoalTactic{}, tok)
	case lexer.UNFOLD:
		// Matches Python: 'proofstep : UNFOLD atype WITH unfspecs'
		//                  'proofstep : UNFOLD WITH unfspecs'
		p.advance()
		var specs []ast.Node
		if p.match(lexer.WITH) {
			// UNFOLD WITH unfspecs (no target)
			specs = p.parseUnfoldSpecs()
		} else {
			name := p.parseCallatom()
			if p.match(lexer.WITH) {
				specs = p.parseUnfoldSpecs()
			} else {
				specs = []ast.Node{&ast.UnfoldSpec{DefName: name}}
			}
		}
		return p.setLoc(&ast.UnfoldTactic{Premise: &ast.NoneAST{}, UnfSpecs: specs}, tok)
	case lexer.IF:
		// Matches Python: 'proofstep : IF fmla proofgroup ELSE proofgroup'
		p.advance()
		cond := p.parseExpr(0)
		thenBranch := p.parseProofBody()
		var elseBranch ast.Node = &ast.NoneAST{}
		if p.match(lexer.ELSE) {
			elseBranch = p.parseProofBody()
		}
		return p.setLoc(&ast.IfTactic{Cond: cond, Then: thenBranch, Else: elseBranch}, tok)
	case lexer.TACTIC:
		// tactic SYMBOL opttacticwith optproofgroup
		// Matches Python: 'proofstep : TACTIC SYMBOL opttacticwith optproofgroup'
		p.advance()
		name := p.parseCallatom()
		// Parse optional "with" clause containing invariants/definitions/triggers
		var withElems ast.Node = &ast.NoneAST{}
		if p.match(lexer.WITH) {
			withElems = p.parseTacticWithList()
		}
		// Parse optional proof group
		var proof ast.Node = &ast.NoneAST{}
		if p.at(lexer.LCB) || p.at(lexer.PROOF) {
			proof = p.parseProofBody()
		}
		return p.setLoc(&ast.TacticTactic{TName: name, Body: withElems, Proof: proof}, tok)
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
		var proof ast.Node = &ast.NoneAST{}
		if p.match(lexer.PROOF) {
			proof = p.parseProofBody()
		}
		return p.setLoc(&ast.PropertyTactic{Prop: lf, PName: &ast.NoneAST{}, Proof: proof}, tok)
	case lexer.THEOREM:
		// theorem [name] { schema_body } [proof { ... }]
		p.advance()
		lf := p.parseLabeledFmla()
		var proof ast.Node = &ast.NoneAST{}
		if p.match(lexer.PROOF) {
			proof = p.parseProofBody()
		}
		return p.setLoc(&ast.PropertyTactic{Prop: lf, PName: &ast.NoneAST{}, Proof: proof}, tok)
	case lexer.INSTANTIATE:
		p.advance()
		// Three forms:
		//   instantiate schema [with matches]
		//   instantiate LABEL schema [with matches]
		//   instantiate with matches  (no schema)
		if p.at(lexer.WITH) {
			// "instantiate with Z = expr, ..."
			p.advance()
			var matches []ast.Node
			for {
				m := p.parseExpr(0)
				matches = append(matches, m)
				if !p.match(lexer.COMMA) {
					break
				}
			}
			ren := ast.Node(ast.NewAnd(matches...))
			return p.setLoc(&ast.SchemaInstantiation{SchemaName: &ast.NoneAST{}, Ren: ren}, tok)
		}
		// Check for optional label: instantiate [label] schema ...
		var label ast.Node
		if p.at(lexer.LB) {
			label = p.parseLabel()
		}
		_ = label
		schema := p.parseCallatom()
		// Optional renaming: schema<V1/V2, V3/V4>
		// Python: optrenaming : LT renamings GT
		// renamings : VARIABLE DIV VARIABLE
		if p.match(lexer.LT) {
			for {
				p.advance() // consume variable name
				p.expect(lexer.DIV)
				p.advance() // consume replacement name
				if !p.match(lexer.COMMA) {
					break
				}
			}
			p.expect(lexer.GT)
		}
		var ren ast.Node = &ast.NoneAST{}
		if p.match(lexer.WITH) {
			var matches []ast.Node
			for {
				m := p.parseExpr(0)
				matches = append(matches, m)
				if !p.match(lexer.COMMA) {
					break
				}
			}
			if len(matches) > 0 {
				ren = ast.NewAnd(matches...)
			}
		}
		return p.setLoc(&ast.SchemaInstantiation{SchemaName: schema, Ren: ren}, tok)
	case lexer.ASSUME:
		p.advance()
		schema := p.parseCallatom()
		// Optional renaming: assume schema<V1/V2>
		p.parseOptRenaming()
		var ren ast.Node = &ast.NoneAST{}
		if p.match(lexer.WITH) {
			var matches []ast.Node
			for {
				m := p.parseExpr(0)
				matches = append(matches, m)
				if !p.match(lexer.COMMA) {
					break
				}
			}
			if len(matches) > 0 {
				ren = ast.NewAnd(matches...)
			}
		}
		return p.setLoc(&ast.AssumeTactic{SchemaName: schema, Ren: ren}, tok)
	case lexer.FORGET:
		p.advance()
		var targets []ast.Node
		for {
			targets = append(targets, p.parseCallatom())
			if !p.match(lexer.COMMA) {
				break
			}
		}
		return p.setLoc(&ast.ForgetTactic{Names: targets}, tok)
	case lexer.SPOIL:
		p.advance()
		target := p.parseCallatom()
		return p.setLoc(&ast.SpoilTactic{Target: target}, tok)
	case lexer.PROOF:
		// Nested proof: "proof [label] { ... }"
		p.advance()
		return p.parseProofBody()
	case lexer.FUNCTION:
		// function definition inside proof: "function f(X) = expr"
		return p.parseFunctionDecl(tok)
	case lexer.INDIV:
		// individual declaration inside proof
		return p.parseConstantDecl(tok)
	case lexer.TYPE:
		// type declaration inside proof/schema
		return p.parseTypeDecl(tok)
	case lexer.RELATION:
		// relation declaration inside proof/schema
		return p.parseRelationDecl(tok)
	case lexer.AXIOM:
		return p.parseAxiomDecl(tok)
	case lexer.DEFINITION:
		return p.parseDefinitionDecl(tok)
	case lexer.LCB:
		return p.parseProofBody()
	default:
		// Matches Python: 'proofstep : SYMBOL' and 'proofstep : SYMBOL WITH matches'
		// Try to parse as expression; if followed by WITH, consume matches.
		expr := p.parseExpr(0)
		if p.match(lexer.WITH) {
			var matches []ast.Node
			for {
				m := p.parseExpr(0)
				matches = append(matches, m)
				if !p.match(lexer.COMMA) {
					break
				}
			}
			ren := ast.Node(ast.NewAnd(matches...))
			return p.setLoc(&ast.SchemaInstantiation{SchemaName: expr, Ren: ren}, tok)
		}
		return expr
	}
}
