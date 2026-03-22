// Package parser implements a hand-written recursive descent parser for Ivy.
// It converts a token stream from the lexer into an AST.
package parser

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// ParseError represents a parser error with location info.
type ParseError struct {
	Message string
	Line    int
	Column  int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at %d:%d: %s", e.Line, e.Column, e.Message)
}

// Parser converts tokens to AST nodes.
type Parser struct {
	lex          *lexer.Lexer
	current      lexer.Token
	version      lexer.Version
	errors       []ParseError
	labelCounter int // auto-label counter, matches Python's label_counter
	// modules maps module names to their definitions for instantiation.
	// Matches Python's stack_lookup for module expansion.
	modules map[string]*ast.ModuleDecl
}

// New creates a parser for the given input and language version.
func New(input string, version lexer.Version) *Parser {
	p := &Parser{
		lex:     lexer.New(input, version),
		version: version,
		modules: make(map[string]*ast.ModuleDecl),
	}
	p.advance() // prime the current token
	return p
}

// Parse parses the input and returns the top-level AST.
// Matches Python behavior: each grammar rule can produce one or more
// declarations (e.g., "after init" produces both ActionDecl and MixinDecl).
// After parsing, expand_autoinstances is run (matching Python's parse()).
func (p *Parser) Parse() ([]ast.Node, error) {
	var decls []ast.Node
	for p.current.Type != lexer.EOF {
		ds := p.parseTopLevel()
		decls = append(decls, ds...)
		// Python continues parsing past errors — don't stop on first error.
	}
	// Post-parse: expand autoinstances (matches Python's expand_autoinstances)
	decls = p.expandAutoInstances(decls)
	var err error
	if len(p.errors) > 0 {
		err = &p.errors[0]
	}
	return decls, err
}

// expandAutoInstances implements Python's expand_autoinstances.
// AutoInstanceDecl nodes are collected and removed from the decl list.
// When a type referenced by an autoinstance is encountered in another
// declaration, the autoinstance is expanded inline via do_insts.
func (p *Parser) expandAutoInstances(decls []ast.Node) []ast.Node {
	type autoKey struct {
		prefix string
		nparams int
	}
	autos := make(map[autoKey][]*ast.Instantiation)
	trefs := make(map[string]bool)
	var result []ast.Node

	for _, decl := range decls {
		if aid, ok := decl.(*ast.AutoInstanceDecl); ok {
			for _, arg := range aid.Args() {
				inst, ok := arg.(*ast.Instantiation)
				if !ok || inst == nil {
					continue
				}
				// inst.Name is the prefix pattern, inst.Sort is the module call
				if inst.Name != nil {
					var nameStr string
					if a, ok := inst.Name.(*ast.Atom); ok {
						nameStr = a.Rep
					} else {
						nameStr = fmt.Sprint(inst.Name)
					}
					pref, parms := extractParametersName(nameStr)
					key := autoKey{pref, len(parms)}
					autos[key] = append(autos[key], inst)
				}
			}
		} else {
			// Collect type names from this declaration
			typeNames := getTypeNames(decl)
			for _, tname := range typeNames {
				if trefs[tname] {
					continue
				}
				trefs[tname] = true
				pref, refparms := extractParametersName(tname)
				key := autoKey{pref, len(refparms)}
				for _, inst := range autos[key] {
					var instNameStr string
					if a, ok := inst.Name.(*ast.Atom); ok {
						instNameStr = a.Rep
					}
					_, parms := extractParametersName(instNameStr)
					// Build substitution: formal parms → actual refparms
					subst := make(map[string]string)
					for i := 0; i < len(parms) && i < len(refparms); i++ {
						subst[parms[i]] = refparms[i]
					}
					// Clone the RHS with substituted args
					lhs := ast.NewAtom(tname)
					var rhsArgs []ast.Node
					if sortAtom, ok := inst.Sort.(*ast.Atom); ok {
						for _, a := range sortAtom.Terms {
							if sa, ok := a.(*ast.Atom); ok {
								rep := sa.Rep
								if repl, ok := subst[rep]; ok {
									rep = repl
								}
								rhsArgs = append(rhsArgs, ast.NewAtom(rep))
							} else {
								rhsArgs = append(rhsArgs, a)
							}
						}
					}
					var rhs ast.Node
					if sortAtom, ok := inst.Sort.(*ast.Atom); ok {
						rhs = ast.NewAtom(sortAtom.Rep, rhsArgs...)
					} else {
						rhs = inst.Sort
					}
					newInst := ast.NewInstantiation(lhs, rhs)
					// Expand via the module registry (do_insts equivalent)
					expanded := p.expandInstantiation(newInst)
					result = append(result, expanded...)
				}
			}
			result = append(result, decl)
		}
	}
	return result
}

// expandInstantiation expands a single Instantiation node using the module registry.
func (p *Parser) expandInstantiation(inst *ast.Instantiation) []ast.Node {
	ca := inst.Sort
	var modName string
	var actualArgs []ast.Node
	if a, ok := ca.(*ast.Atom); ok {
		modName = a.Rep
		actualArgs = a.Terms
	}
	modDef, found := p.modules[modName]
	if !found || modDef == nil {
		return []ast.Node{ast.NewInstantiateDecl(inst)}
	}
	formalParams := modDef.FormalParams
	subst := make(map[string]string)
	for i := 0; i < len(formalParams) && i < len(actualArgs); i++ {
		var formalName string
		if a, ok := formalParams[i].(*ast.Atom); ok {
			formalName = a.Rep
		} else if s, ok := formalParams[i].(*ast.Symbol); ok {
			formalName = s.Rep
		}
		var actualName string
		if a, ok := actualArgs[i].(*ast.Atom); ok {
			actualName = a.Rep
		} else if s, ok := actualArgs[i].(*ast.Symbol); ok {
			actualName = s.Rep
		} else {
			actualName = fmt.Sprint(actualArgs[i])
		}
		if formalName != "" {
			subst[formalName] = actualName
		}
	}
	prefix := ""
	if inst.Name != nil {
		if a, ok := inst.Name.(*ast.Atom); ok {
			prefix = a.Rep
		}
	}
	var result []ast.Node
	if prefix != "" {
		result = append(result, ast.NewObjectDecl(ast.NewAtom(prefix)))
	}
	for _, bodyDecl := range modDef.BodyDecls {
		expanded := substituteNamesInDecl(bodyDecl, subst)
		if prefix != "" {
			result = append(result, prefixDeclNames(expanded, prefix)...)
		} else {
			result = append(result, expanded)
		}
	}
	return result
}

// extractParametersName extracts the prefix and bracket parameters from a name.
// "foo[bar][baz]" → ("foo", ["bar", "baz"])
// Matches Python's ivy_utils.extract_parameters_name.
func extractParametersName(name string) (string, []string) {
	var parms []string
	pos := len(name) - 1
	for pos >= 0 && name[pos] == ']' {
		end := pos
		pos--
		count := 1
		for pos >= 0 && count > 0 {
			if name[pos] == '[' {
				count--
			} else if name[pos] == ']' {
				count++
			}
			pos--
		}
		if pos >= 0 {
			parms = append(parms, name[pos+2:end])
		}
	}
	if pos >= 0 {
		// Reverse parms
		for i, j := 0, len(parms)-1; i < j; i, j = i+1, j-1 {
			parms[i], parms[j] = parms[j], parms[i]
		}
		return name[:pos+1], parms
	}
	return name, nil
}

// getTypeNames collects type names referenced by a declaration.
// Matches Python's TypeNames / get_type_names / tterm_type_names.
func getTypeNames(decl ast.Node) []string {
	var names []string
	seen := make(map[string]bool)
	var addName func(string)
	addName = func(tname string) {
		if seen[tname] {
			return
		}
		seen[tname] = true
		// Recurse into parameter names
		_, refparms := extractParametersName(tname)
		for _, rp := range refparms {
			addName(rp)
		}
		names = append(names, tname)
	}
	// Walk the declaration looking for sort annotations
	walkTypeNames(decl, addName)
	return names
}

// walkTypeNames recursively walks an AST node collecting sort/type names.
func walkTypeNames(node ast.Node, addName func(string)) {
	if node == nil {
		return
	}
	// Check for sort annotations
	switch n := node.(type) {
	case *ast.Atom:
		if n.ASort != nil {
			if sym, ok := n.ASort.(*ast.Symbol); ok {
				addName(sym.Rep)
			}
		}
	case *ast.Variable:
		if n.VSort != nil {
			if sym, ok := n.VSort.(*ast.Symbol); ok {
				addName(sym.Rep)
			}
		}
	case *ast.Symbol:
		if n.Sort != nil {
			if sym, ok := n.Sort.(*ast.Symbol); ok {
				addName(sym.Rep)
			}
		}
	case *ast.TypeDef:
		if n.Name != nil {
			if sym, ok := n.Name.(*ast.Symbol); ok {
				addName(sym.Rep)
			}
		}
	}
	for _, child := range node.Args() {
		walkTypeNames(child, addName)
	}
}

// Errors returns all accumulated parse errors.
func (p *Parser) Errors() []ParseError {
	return p.errors
}

// AtEOF returns true if the parser has consumed all input.
func (p *Parser) AtEOF() bool {
	return p.current.Type == lexer.EOF
}

// CurrentTokenType returns the type of the current (lookahead) token.
func (p *Parser) CurrentTokenType() lexer.TokenType {
	return p.current.Type
}

// --- Token management ---

func (p *Parser) advance() lexer.Token {
	prev := p.current
	p.current = p.lex.NextToken()
	return prev
}

func (p *Parser) peek() lexer.Token {
	return p.current
}

func (p *Parser) at(tt lexer.TokenType) bool {
	return p.current.Type == tt
}

func (p *Parser) match(tt lexer.TokenType) bool {
	if p.current.Type == tt {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) expect(tt lexer.TokenType) lexer.Token {
	if p.current.Type == tt {
		return p.advance()
	}
	p.errorf("expected %s, got %s (%q)", tt, p.current.Type, p.current.Value)
	return p.current
}

func (p *Parser) errorf(format string, args ...interface{}) {
	p.errors = append(p.errors, ParseError{
		Message: fmt.Sprintf(format, args...),
		Line:    p.current.Line,
		Column:  p.current.Column,
	})
}

func (p *Parser) loc() ast.Location {
	return ast.Location{Line: p.current.Line}
}

func (p *Parser) setLoc(n ast.Node, tok lexer.Token) ast.Node {
	n.SetLineno(ast.Location{Line: tok.Line})
	return n
}

// vle checks if version <= (major, minor).
func (p *Parser) vle(major, minor int) bool {
	return p.version[0] < major || (p.version[0] == major && p.version[1] <= minor)
}

// --- Helper parsers ---

// parseSymbol parses a SYMBOL token and returns an ast.Symbol.
func (p *Parser) parseSymbol() *ast.Symbol {
	tok := p.expect(lexer.SYMBOL)
	return ast.NewSymbol(tok.Value, nil)
}

// parseAtomName parses a symbol name (SYMBOL or THIS).
// parseAtomName parses a symbol name, absorbing bracket subscripts.
// Matches Python's grammar: SYMBOL : SYMBOL LB SYMsubscr RB
// So "bv[64]" → "bv[64]", "map[key][value]" → "map[key][value]"
func (p *Parser) parseAtomName() (string, lexer.Token) {
	tok := p.current
	switch tok.Type {
	case lexer.SYMBOL:
		p.advance()
		name := tok.Value
		// Absorb bracket subscripts: name[sub1][sub2]...
		for p.at(lexer.LB) {
			name += p.absorbSubscript()
		}
		return name, tok
	case lexer.THIS:
		p.advance()
		return "this", tok
	default:
		p.expect(lexer.SYMBOL) // will error
		return "", tok
	}
}

// parseCallatom parses a callable atom (may have dots).
// Python (line 2662): callatom : METHOD → Atom('method')
func (p *Parser) parseCallatom() ast.Node {
	// Item 16: METHOD as a callatom
	if p.at(lexer.METHOD) {
		tok := p.advance()
		result := ast.Node(ast.NewAtom("method"))
		p.setLoc(result, tok)
		return result
	}
	name, tok := p.parseAtomName()
	var args []ast.Node
	if p.match(lexer.LPAREN) {
		args = p.parseTermList()
		p.expect(lexer.RPAREN)
	}
	result := ast.Node(ast.NewAtom(name, args...))
	p.setLoc(result, tok)

	// Handle dot chaining: a.b.c(...)
	for p.match(lexer.DOT) {
		name2, tok2 := p.parseAtomName()
		var args2 []ast.Node
		if p.match(lexer.LPAREN) {
			args2 = p.parseTermList()
			p.expect(lexer.RPAREN)
		}
		right := ast.NewAtom(name2, args2...)
		p.setLoc(right, tok2)
		result = ast.NewDot(result, right)
	}
	return result
}

// parseTermList parses comma-separated terms.
func (p *Parser) parseTermList() []ast.Node {
	if p.at(lexer.RPAREN) || p.at(lexer.RB) {
		return nil
	}
	var terms []ast.Node
	terms = append(terms, p.parseExpr(0))
	for p.match(lexer.COMMA) {
		terms = append(terms, p.parseExpr(0))
	}
	return terms
}

// parseAType parses a type annotation (just a name, possibly dotted).
// Uses parseAtomName which handles bracket subscripts.
func (p *Parser) parseAType() ast.Node {
	tok := p.current

	switch tok.Type {
	case lexer.SYMBOL, lexer.THIS:
		// parseAtomName handles SYMBOL (with bracket absorption) and THIS
		name, nameTok := p.parseAtomName()
		result := ast.Node(ast.NewSymbol(name, nil))
		p.setLoc(result, nameTok)
		// Handle dotted types: mod.type
		for p.match(lexer.DOT) {
			name2, _ := p.parseAtomName()
			right := ast.NewSymbol(name2, nil)
			result = ast.NewDot(result, right)
		}
		return result
	default:
		p.errorf("expected type name, got %s", tok.Type)
		return ast.NewSymbol("?", nil)
	}
}

// absorbSubscript consumes [content] and returns it as a string (e.g., "[64]").
// Matches Python: SYMBOL : SYMBOL LB SYMsubscr RB
func (p *Parser) absorbSubscript() string {
	if !p.match(lexer.LB) {
		return ""
	}
	content := p.current.Value
	p.advance() // consume the subscript content
	p.expect(lexer.RB)
	return "[" + content + "]"
}

// parseTTerm parses a typed term: name or name : type or name(params) : type
// Also handles ^name:type (KeyArg) for action parameters.
func (p *Parser) parseTTerm() ast.Node {
	tok := p.current
	// Python: lparam : CARET SYMBOL COLON atype → KeyArg
	if p.at(lexer.CARET) {
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
	ca := p.parseCallatom()
	if p.match(lexer.COLON) {
		atype := p.parseAType()
		if a, ok := ca.(*ast.Atom); ok {
			a.ASort = atype
			return a
		}
	}
	p.setLoc(ca, tok)
	return ca
}

// parseTTermList parses comma-separated typed terms.
func (p *Parser) parseTTermList() []ast.Node {
	var terms []ast.Node
	terms = append(terms, p.parseTTerm())
	for p.match(lexer.COMMA) {
		terms = append(terms, p.parseTTerm())
	}
	return terms
}

// parseLabel parses an optional [label] prefix. Returns nil if no label.
func (p *Parser) parseLabel() ast.Node {
	if p.at(lexer.LB) {
		p.advance()
		name, tok := p.parseAtomName()
		var args []ast.Node
		// Label can have arguments: [label(x,y)]
		if p.match(lexer.LPAREN) {
			args = p.parseTermList()
			p.expect(lexer.RPAREN)
		}
		p.expect(lexer.RB)
		a := ast.NewAtom(name, args...)
		p.setLoc(a, tok)
		return a
	}
	return nil
}

// parseLabeledFmla parses an optional label followed by a formula.
func (p *Parser) parseLabeledFmla() *ast.LabeledFormula {
	tok := p.current
	label := p.parseLabel()
	// Check for schema body: { decls ; conclusion }
	// Matches Python's schdefnrhs : LCB schdecls schconc RCB
	var fmla ast.Node
	if p.at(lexer.LCB) {
		fmla = p.parseSchemaBody()
	} else {
		fmla = p.parseExpr(0)
	}
	lf := ast.NewLabeledFormula(label, fmla)
	p.setLoc(lf, tok)
	return lf
}

// parseSchemaBody parses "{ schdecls schconc }".
// Matches Python's schdefnrhs : LCB schdecls schconc RCB
// schdecl can be type/relation/function/individual/property declarations.
// The LAST property/axiom is the conclusion; all others are premises.
func (p *Parser) parseSchemaBody() ast.Node {
	tok := p.current
	p.expect(lexer.LCB)
	var elems []ast.Node
	prevTok := p.current
	firstIter := true
	for !p.at(lexer.RCB) && !p.at(lexer.EOF) {
		// Guard against infinite loops
		if !firstIter && p.current == prevTok {
			break
		}
		firstIter = false
		prevTok = p.current
		if p.at(lexer.LCB) {
			// Nested schema body: schdecl : schdefnrhs
			// Python wraps in LabeledFormula(None, SchemaBody(...)) with label 'sch'
			nested := p.parseSchemaBody()
			lf := ast.NewLabeledFormula(nil, nested)
			p.setLoc(lf, tok)
			elems = append(elems, lf)
		} else if p.at(lexer.FRESH) {
			// Python (lines 684-702): FRESH FUNCTION funs / FRESH INDIV funs / FRESH RELATION rels
			// Replace ConstantDecl with FreshConstantDecl
			p.advance() // consume FRESH
			ds := p.parseTopLevel()
			for _, d := range ds {
				if cd, ok := d.(*ast.ConstantDecl); ok {
					// Convert ConstantDecl to FreshConstantDecl
					fcd := &ast.FreshConstantDecl{ConstantDecl: *cd}
					elems = append(elems, fcd)
				} else {
					elems = append(elems, d)
				}
			}
		} else {
			ds := p.parseTopLevel()
			elems = append(elems, ds...)
		}
	}
	p.expect(lexer.RCB)
	sb := &ast.SchemaBody{Elems: elems}
	p.setLoc(sb, tok)
	return sb
}

// parseSimpleVars parses comma-separated simple variables: X:S, Y:T
// Uses simple type names (no dotted types) to avoid consuming the DOT
// that terminates "forall X:t. body".
// parseSomeParams parses parameters for "some" expressions: SYMBOL:SYMBOL or VARIABLE:SYMBOL.
// Python: params : param (comma-separated), param : SYMBOL COLON SYMBOL
func (p *Parser) parseSomeParams() []ast.Node {
	var params []ast.Node
	for p.at(lexer.SYMBOL) || p.at(lexer.VARIABLE) {
		tok := p.advance()
		name := tok.Value
		var sort ast.Node
		if p.match(lexer.COLON) {
			stok := p.expect(lexer.SYMBOL)
			sort = ast.NewSymbol(stok.Value, nil)
		}
		// Python creates App(name) with sort, which maps to our Atom
		a := ast.NewAtom(name)
		if sort != nil {
			a.ASort = sort
		}
		p.setLoc(a, tok)
		params = append(params, a)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	return params
}

func (p *Parser) parseSimpleVars() []ast.Node {
	var vars []ast.Node
	for p.at(lexer.VARIABLE) {
		tok := p.advance()
		var sort ast.Node
		if p.match(lexer.COLON) {
			// Use simple type name (SYMBOL only, no dots)
			stok := p.expect(lexer.SYMBOL)
			sort = ast.NewSymbol(stok.Value, nil)
		}
		v := ast.NewVariable(tok.Value, sort)
		p.setLoc(v, tok)
		vars = append(vars, v)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	return vars
}

// parseParams parses parenthesized or bare parameter lists.
func (p *Parser) parseParams() []ast.Node {
	if p.match(lexer.LPAREN) {
		params := p.parseTTermList()
		p.expect(lexer.RPAREN)
		return params
	}
	return nil
}

// parseSort parses a sort definition: {a,b,c} | struct{...} | uninterpreted | range
func (p *Parser) parseSort() ast.Node {
	tok := p.current
	switch tok.Type {
	case lexer.LCB:
		p.advance()
		// Could be {a,b,c} enum or {lo..hi} range
		first := p.parseExpr(0)
		if p.match(lexer.DOTS) {
			// Range: {lo..hi}
			hi := p.parseExpr(0)
			p.expect(lexer.RCB)
			return p.setLoc(ast.NewRange(first, hi), tok)
		}
		// Enum: {a, b, c}
		elems := []ast.Node{first}
		for p.match(lexer.COMMA) {
			elems = append(elems, p.parseExpr(0))
		}
		p.expect(lexer.RCB)
		return p.setLoc(ast.NewEnumeratedSort(elems...), tok)

	case lexer.STRUCT:
		p.advance()
		p.expect(lexer.LCB)
		var fields []ast.Node
		if !p.at(lexer.RCB) {
			fields = p.parseTTermList()
		}
		p.expect(lexer.RCB)
		return p.setLoc(ast.NewStructSort(fields...), tok)

	default:
		// Just a type name
		return p.parseAType()
	}
}
