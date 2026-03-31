// grammar_v17.y — goyacc LALR(1) grammar for FULL Ivy file parsing, version 1.7+.
// Mechanically translated from Python PLY grammar in ivy_parser.py and ivy_logic_parser.py.
//
// This is the FULL grammar: it parses complete Ivy files (top-level declarations),
// not just formulas/terms like lalr_logicparser.
//
// Precedence table (copied exactly from Python ivy_parser.py, v1.7+):
//   SEMI < GLOBALLY/EVENTUALLY < ARROW/IFF < OR < AND < TILDA
//   < EQ/LE/LT/GE/GT/PTO < TILDAEQ < IF < ELSE < COLON
//   < PLUS/MINUS < TIMES/DIV < DOLLAR < OLD < DOT

%{
package lalr_full

import (
	"fmt"
	"path/filepath"
	"strings"
	"github.com/glycerine/ivy/goivy/ast"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// acfg extracts the *ast.AstConfig from the lexer for use in grammar actions.
func acfg(lex v17Lexer) *ast.AstConfig {
	return lex.(*v17LexAdapter).astCfg
}

// parentObject is no longer a global. It is passed explicitly to newIvyAccum().
// See newIvyAccum(parent, parentObjName) in ivy_module.go.

// getLineno returns a Location for the current token position.
// Matches Python's get_lineno(p, n) → iu.Location(iu.filename, p.lineno(n)).
func getLineno(lex *v17LexAdapter) ast.Location {
	xtracer.Trace("parser.get_lineno ENTER")
	return ast.Location{
		Filename: normalizeFilename(lex.filename),
		Line:     lex.prevTok.Line,
	}
}

// normalizeFilename replaces the include directory path with <IVY_INCLUDE>
// so that canonical strings match between Go and Python regardless of install location.
// Python: stores the raw path; the golden test normalizes for display.
// For hashing, both sides must agree, so we normalize here.
func normalizeFilename(f string) string {
	stdDir := iu.NewIvyUtilsConfig().GetStdIncludeDir()
	if stdDir == "" {
		return f
	}
	// stdDir is like "/path/to/include/1.8". Get the parent: "/path/to/include/"
	baseDir := filepath.Dir(stdDir)
	if baseDir != "" && !strings.HasSuffix(baseDir, string(filepath.Separator)) {
		baseDir += string(filepath.Separator)
	}
	if strings.HasPrefix(f, baseDir) {
		return "<IVY_INCLUDE>/" + f[len(baseDir):]
	}
	return f
}

// newLabel generates a unique label with the given prefix, matching Python newlabel().
func newLabel(cfg *ast.AstConfig, pref string) *ast.Atom {
	xtracer.Trace("parser.newlabel ENTER")
	cfg.LabelCounter++
	return cfg.NewAtom(fmt.Sprintf("%s%d", pref, cfg.LabelCounter))
}

// addLabel adds a label to a LabeledFormula if it doesn't have one.
// Matches Python addlabel() (ivy_parser.py:443-448).
func addLabel(cfg *ast.AstConfig, lf *ast.LabeledFormula, pref string) *ast.LabeledFormula {
	xtracer.Trace("parser.addlabel ENTER")
	if lf.Label != nil {
		return lf
	}
	res := cfg.NewLabeledFormula(newLabel(cfg, pref), lf.Formula)
	res.Lineno = lf.Lineno
	return res
}

// mkLF wraps a node in a LabeledFormula with no label.
// Matches Python mk_lf (ivy_parser.py:1187).
func mkLF(cfg *ast.AstConfig, x ast.Node) *ast.LabeledFormula {
	xtracer.Trace("parser.mk_lf ENTER")
	lf := cfg.NewLabeledFormula(nil, x)
	return lf
}

// checkNonTemporal validates that a formula doesn't contain temporal operators.
// Matches Python check_non_temporal() (ivy_parser.py:231-248).
// For LabeledFormula, recursively checks the formula part.
// Reports an error if temporal operators are found.
func checkNonTemporal(x ast.Node) ast.Node {
	xtracer.Trace("parser.check_non_temporal ENTER")
	if lf, ok := x.(*ast.LabeledFormula); ok {
		checkNonTemporal(lf.Formula)
		return x
	}
	if ast.HasTemporal(x) {
		// TODO: report_error(IvyError(x, "non-temporal formula expected"))
		fmt.Printf("warning: non-temporal formula expected\n")
	}
	return x
}

// addUnprovable marks a LabeledFormula as unprovable if cond is non-nil.
// Matches Python addunprovable() (ivy_parser.py:453-457).
func addUnprovable(lf *ast.LabeledFormula, cond ast.Node) *ast.LabeledFormula {
	xtracer.Trace("parser.addunprovable ENTER")
	if cond != nil {
		lf.Unprovable = true
	}
	return lf
}

// addTemporal marks a LabeledFormula as temporal.
// Matches Python addtemporal() (ivy_parser.py:438-441).
func addTemporal(lf *ast.LabeledFormula) *ast.LabeledFormula {
	xtracer.Trace("parser.addtemporal ENTER")
	t := true
	lf.Temporal = &t
	return lf
}

// addExplicit marks a LabeledFormula as explicit.
// Matches Python addexplicit() (ivy_parser.py:459-462).
func addExplicit(lf *ast.LabeledFormula) *ast.LabeledFormula {
	xtracer.Trace("parser.addexplicit ENTER")
	lf.Explicit = true
	return lf
}

// nodeLineno extracts the Location from a Node's GetLineno().
// Used when we need the line of a child node instead of lastTok (which is the lookahead).
// Emits the get_lineno trace to match Python's get_lineno(p, n) call.
func nodeLineno(n ast.Node) ast.Location {
	xtracer.Trace("parser.get_lineno ENTER")
	if n == nil {
		return ast.Location{}
	}
	loc := n.GetLineno()
	loc.Filename = normalizeFilename(loc.Filename)
	return loc
}

// atypeToString extracts the string sort name from an atype Node.
// Python atype returns a plain string; Go atype returns *ast.Symbol or *ast.This.
func atypeToString(n ast.Node) string {
	switch v := n.(type) {
	case *ast.Symbol:
		return v.Rep
	case *ast.This:
		return "this"
	default:
		return fmt.Sprint(n)
	}
}

// atypeToAtom converts an atype Node (Symbol or This) into an Atom,
// matching Python's Atom(p[n]) where atype returns a string.
// Python atype returns a string; Go atype returns *ast.Symbol or *ast.This.
func atypeToAtom(cfg *ast.AstConfig, n ast.Node) *ast.Atom {
	switch v := n.(type) {
	case *ast.Symbol:
		return cfg.NewAtom(v.Rep)
	case *ast.This:
		return cfg.NewAtom("this")
	default:
		return cfg.NewAtom(fmt.Sprint(n))
	}
}

// makeMixinName generates a unique mixin name.
// Matches Python make_mixin_name() (ivy_parser.py:2556-2562).
func makeMixinName(cfg *ast.AstConfig, atom *ast.Atom, suffix string) *ast.Atom {
	xtracer.Trace("parser.make_mixin_name ENTER")
	cfg.LabelCounter++
	// Python: name = atom.rep.replace(iu.ivy_compose_character, '_') + '[' + suffix + str(label_counter) + ']'
	rep := strings.ReplaceAll(atom.Rep, ".", "_")
	return cfg.NewAtom(fmt.Sprintf("%s[%s%d]", rep, suffix, cfg.LabelCounter))
}

// handleMixin declares a mixin (before/after/implement).
// Matches Python handle_mixin() (ivy_parser.py:2083-2090).
func handleMixin(cfg *ast.AstConfig, kind string, mixer *ast.Atom, mixee *ast.Atom, ivy *ivyAccum) {
	xtracer.Trace("parser.handle_mixin ENTER")
	var m ast.Node
	switch kind {
	case "before":
		m = cfg.NewMixinBeforeDef(mixer, mixee)
	case "after":
		m = cfg.NewMixinAfterDef(mixer, mixee)
	case "implement":
		m = cfg.NewMixinImplementDef(mixer, mixee)
	default:
		m = cfg.NewMixinBeforeDef(mixer, mixee)
	}
	md := cfg.NewMixinDecl(m)
	ivy.declare(md)
}

// handleBeforeAfter processes before/after action declarations.
// Matches Python handle_before_after() (ivy_parser.py:2105-2115).
// inferActionParams looks up the matching action definition and infers
// missing formal parameters and returns.
// Matches Python infer_action_params() (ivy_parser.py:2093-2103).
// stackActionLookup searches the accumulator stack for an action definition.
// Matches Python stack_action_lookup() (ivy_parser.py:125-133).
func stackActionLookup(ivy *ivyAccum, name string) (ast.Node, int) {
	xtracer.Trace("parser.stack_action_lookup ENTER")
	params := 0
	for cur := ivy; cur != nil; cur = cur.parent {
		if cur.isModule {
			break
		}
		params += len(cur.params)
		if ad, ok := cur.actions[name]; ok {
			return ad, params
		}
	}
	return nil, 0
}

func inferActionParams(ivy *ivyAccum, actname string, formals []ast.Node, returns []ast.Node) ([]ast.Node, []ast.Node) {
	xtracer.Trace("parser.infer_action_params ENTER")
	mixee, numParams := stackActionLookup(ivy, actname)
	if mixee == nil {
		return formals, returns
	}
	// Python: if ("common" in mixee.attributes) != ("common" in stack[-1].attributes):
	//             return formals, returns
	mixeeDef, ok := mixee.(*ast.ActionDef)
	if !ok {
		return formals, returns
	}
	mixeeCommon := hasAttributeStr(mixeeDef.Attributes, "common")
	ivyCommon := hasAttributeStr(ivy.attributes, "common")
	if mixeeCommon != ivyCommon {
		return formals, returns
	}
	// Python: mformals, mreturns = mixee.formals()
	mformals, mreturns := mixeeDef.Formals()
	// Python: formals.extend(mformals[num_params+len(formals):])
	start := numParams + len(formals)
	if start < len(mformals) {
		formals = append(formals, mformals[start:]...)
	}
	// Python: returns.extend(mreturns[len(returns):])
	rstart := len(returns)
	if rstart < len(mreturns) {
		returns = append(returns, mreturns[rstart:]...)
	}
	return formals, returns
}

func handleBeforeAfter(cfg *ast.AstConfig, kind string, atom *ast.Atom, action ast.Node, ivy *ivyAccum, optargs []ast.Node, optreturns []ast.Node) {
	xtracer.Trace("parser.handle_before_after ENTER")
	mixer := makeMixinName(cfg, atom, kind)
	optargs, optreturns = inferActionParams(ivy, atom.Rep, optargs, optreturns)
	df := cfg.NewActionDef(mixer, action, optargs, optreturns)
	df.SetLineno(atom.GetLineno())
	decl := cfg.NewActionDecl(df)
	ivy.declare(decl)
	handleMixin(cfg, kind, mixer, atom, ivy)
}

// createObject processes an object declaration by expanding its body
// with prefix substitution via instMod.
// Matches Python create_object() (ivy_parser.py:678-692).
// setObjectDefined is now in inst_mod.go with proper implementation.

// parseNativequote parses a native code block, splitting on backtick-delimited references.
// Matches Python parse_nativequote() (ivy_parser.py:2457-2470).
func parseNativequote(cfg *ast.AstConfig, raw string, lex *v17LexAdapter) (string, []ast.Node) {
	xtracer.Trace("parser.parse_nativequote ENTER")
	// Drop the <<< and >>> quotation marks
	s := raw
	if len(s) >= 6 && strings.HasPrefix(s, "<<<") && strings.HasSuffix(s, ">>>") {
		s = s[3 : len(s)-3]
	}
	fields := strings.Split(s, "`")
	var bqs []ast.Node
	for idx, f := range fields {
		if idx%2 == 1 {
			if f == "this" {
				bqs = append(bqs, cfg.NewAtom("this"))
			} else {
				bqs = append(bqs, cfg.NewAtom(f))
			}
		}
	}
	var parts []string
	for idx, f := range fields {
		if idx%2 == 0 {
			parts = append(parts, f)
		} else {
			parts = append(parts, fmt.Sprintf("%d", idx/2))
		}
	}
	text := strings.Join(parts, "`")
	// Python: loc = get_lineno(p, n)
	_ = getLineno(lex)
	return text, bqs
}

// methcall matches Python methcall(lhs, rhs) at ivy_parser.py:3034-3038.
// If lhs is an App or Atom with no args, compose; otherwise create MethodCall.
func methcall(cfg *ast.AstConfig, lhs, rhs ast.Node) ast.Node {
	xtracer.Trace("parser.methcall ENTER")
	switch l := lhs.(type) {
	case *ast.App:
		if len(l.Terms) == 0 {
			return ast.ComposeAtomsGeneric(l, rhs)
		}
	case *ast.Atom:
		if len(l.Terms) == 0 {
			return ast.ComposeAtomsGeneric(l, rhs)
		}
	}
	return cfg.NewMethodCall(lhs, rhs)
}

// fixIfPart handles the `some` condition case in if/while actions.
// Matches Python fix_if_part() (ivy_parser.py:2932-2938).
func fixIfPart(cond ast.Node, part ast.Node) ast.Node {
	xtracer.Trace("parser.fix_if_part ENTER")
	// Python: isinstance(cond, Some) — matches Some, SomeMin, SomeMax (all subclasses).
	// Extract params from whichever type matches.
	var params []ast.Node
	switch s := cond.(type) {
	case *ast.Some:
		params = s.Params
	case *ast.SomeMin:
		params = s.Params
	case *ast.SomeMax:
		params = s.Params
	}
	if params != nil {
		subst := make(map[string]string)
		for _, p := range params {
			// Python: subst = dict((x.rep[4:],x.rep) for x in args)
			// Params can be App, Atom, Variable, etc. — use NodeRep generically.
			rep := ast.NodeRep(p)
			if len(rep) > 4 {
				subst[rep[4:]] = rep
			}
		}
		if len(subst) > 0 {
			part = ast.SubstPrefixAtomsAst(part, subst, nil, nil, nil)
		}
	}
	return part
}

// createObject processes an object declaration by expanding its body
// with prefix substitution via instMod.
// Matches Python create_object() (ivy_parser.py:678-693) EXACTLY.
func createObject(cfg *ast.AstConfig, top *ivyAccum, name *ast.Atom, objectargs []ast.Node, module *ivyAccum, lineno ast.Location, continuation bool) {
	xtracer.Trace("parser.create_object ENTER name=%s", name.Rep)

	// Python line 680: prefargs = [Variable('V'+str(idx),pr.sort) for idx,pr in enumerate(objectargs)]
	var prefargs []ast.Node
	for idx, pr := range objectargs {
		vname := fmt.Sprintf("V%d", idx)
		var sort string
		if a, ok := pr.(*ast.Atom); ok && a.ASort != nil {
			if sym, ok := a.ASort.(*ast.Symbol); ok {
				sort = sym.Rep
			} else {
				sort = fmt.Sprint(a.ASort)
			}
		}
		prefargs = append(prefargs, cfg.NewVariable(vname, sort))
	}

	// Python line 681: pref = Atom(name, prefargs)
	pref := cfg.NewAtom(name.Rep, prefargs...)
	pref.SetLineno(lineno)

	// Python line 684-686
	if !continuation {
		top.declare(cfg.NewObjectDecl(pref))
		// Python: top.set_object_defined(name, module.defined)
		setObjectDefined(top, name.Rep, module.defined)
	}

	// Python line 687: vsubst = dict((pr.rep,v) for pr,v in zip(objectargs,prefargs))
	vsubst := make(map[string]*ast.Variable)
	for i, pr := range objectargs {
		if i < len(prefargs) {
			prName := nodeRep(pr)
			if v, ok := prefargs[i].(*ast.Variable); ok {
				vsubst[prName] = v
			}
		}
	}

	// Python line 688: inst_mod(top, module, pref, {}, vsubst)
	instMod(top, module, pref, map[string]string{}, vsubst, "", lineno)

	xtracer.Trace("parser.create_object EXIT name=%s", name.Rep)
}

// TokenInfo carries both the string value and source location of a terminal token.
// This is the standard goyacc approach for accurate line numbers in grammar rules.
type TokenInfo struct {
	Val  string
	Line int
}

// tokLineno creates a Location from a TokenInfo, using the normalized filename
// from the lex adapter. This replaces getLineno for cases where we have direct
// access to the token's position.
func tokLineno(lex *v17LexAdapter, tok TokenInfo) ast.Location {
	xtracer.Trace("parser.get_lineno ENTER")
	return ast.Location{
		Filename: normalizeFilename(lex.filename),
		Line:     tok.Line,
	}
}

%}

// The union type for semantic values.
%union {
	node     ast.Node
	nodes    []ast.Node
	str      string
	tok      TokenInfo
	bval     bool
	accum    *ivyAccum
}

// Terminal tokens — identifiers and literals (carry string value + line)
%token <tok>  TOK_PRESYMBOL TOK_VARIABLE
%token <tok>  TOK_LABEL   // Python returns LABEL from lexer; Go handles via labelname rule instead
%token <tok>  TOK_NATIVEQUOTE

// Terminal tokens — punctuation (carry line number only)
%token <tok>  TOK_LPAREN TOK_RPAREN TOK_LB TOK_RB TOK_LCB TOK_RCB
%token <tok>  TOK_COMMA TOK_SEMI TOK_COLON TOK_DOT
%token <tok>  TOK_DOTS TOK_DOTDOTDOT

// Terminal tokens — operators (carry line number only)
%token <tok>  TOK_PLUS TOK_MINUS TOK_TIMES TOK_DIV
%token <tok>  TOK_EQ TOK_TILDAEQ TOK_TILDA TOK_LE TOK_LT TOK_GE TOK_GT
%token <tok>  TOK_AND TOK_OR TOK_ARROW TOK_IFF
%token <tok>  TOK_PTO TOK_DOLLAR TOK_CARET
%token <tok>  TOK_ASSIGN

// Terminal tokens — keywords (logic, carry line number only)
%token <tok>  TOK_FORALL TOK_EXISTS
%token <tok>  TOK_TRUE TOK_FALSE
%token <tok>  TOK_OLD TOK_THIS TOK_ISA
%token <tok>  TOK_IF TOK_ELSE
%token <tok>  TOK_GLOBALLY TOK_EVENTUALLY
%token <tok>  TOK_WHENNEXT TOK_WHENPREV TOK_WHENFIRST TOK_WHENLAST

// Terminal tokens — keywords (actions, carry line number only)
%token <tok>  TOK_ASSUME TOK_ASSERT TOK_REQUIRE TOK_ENSURE
%token <tok>  TOK_VAR TOK_LOCAL TOK_LET TOK_CALL
%token <tok>  TOK_WHILE TOK_FOR TOK_IN TOK_INVARIANT TOK_DECREASES
%token <tok>  TOK_RETURNS
%token <tok>  TOK_SOME TOK_MINIMIZING TOK_MAXIMIZING
%token <tok>  TOK_DEBUG TOK_THUNK TOK_UNPROVABLE TOK_PROOF
%token <tok>  TOK_INSTANTIATE

// Terminal tokens — keywords (declarations, carry line number only)
%token <tok>  TOK_RELATION TOK_INDIV TOK_FUNCTION TOK_DERIVED
%token <tok>  TOK_AXIOM TOK_CONJECTURE TOK_SCHEMA TOK_THEOREM
%token <tok>  TOK_PROPERTY TOK_DEFINITION
%token <tok>  TOK_TYPE TOK_STRUCT
%token <tok>  TOK_MODULE TOK_OBJECT TOK_CLASS TOK_SUBCLASS
%token <tok>  TOK_ACTION TOK_METHOD
%token <tok>  TOK_BEFORE TOK_AFTER TOK_AROUND TOK_MIXIN TOK_IMPLEMENT
%token <tok>  TOK_ISOLATE TOK_EXTRACT TOK_TRUSTED
%token <tok>  TOK_EXPORT TOK_IMPORT TOK_DELEGATE TOK_USING TOK_INCLUDE
%token <tok>  TOK_INTERPRET TOK_MACRO TOK_ALIAS TOK_ATTRIBUTE
%token <tok>  TOK_VARIANT TOK_OF
%token <tok>  TOK_SCENARIO
%token <tok>  TOK_PROGRESS TOK_RELY TOK_MIXORD
%token <tok>  TOK_CONCEPT TOK_STATE TOK_UPDATE TOK_FROM
%token <tok>  TOK_PARAMS TOK_MODIFIES TOK_ENSURES TOK_REQUIRES
%token <tok>  TOK_INIT TOK_ENTRY TOK_SET TOK_NULL TOK_MATCH
%token <tok>  TOK_FRESH TOK_NAMED
%token        TOK_TEMPORAL TOK_EXPLICIT
%token        TOK_SPECIFICATION TOK_IMPLEMENTATION TOK_PRIVATE
%token        TOK_GLOBAL TOK_COMMON
%token <tok>  TOK_GHOST TOK_FINITE
%token <tok>  TOK_PARAMETER
%token <tok>  TOK_DESTRUCTOR TOK_CONSTRUCTOR TOK_FIELD
%token <tok>  TOK_AUTOINSTANCE
// TOK_VAR_KW removed (was unused placeholder)

// Proof/tactic tokens
%token <tok>  TOK_TACTIC TOK_TRIGGER
%token <tok>  TOK_SHOWGOALS TOK_DEFERGOAL TOK_SPOIL
%token <tok>  TOK_UNFOLD TOK_FORGET
%token <tok>  TOK_APPLY
%token <tok>  TOK_WITH
// TOK_METHOD_KW, TOK_NULL_KW, TOK_SET_KW removed (were unused placeholders)

// Nonterminal types — formula/term
%type <node>  term fmla appelem aterm var simplevar atype
%type <nodes> terms vars simplevars
%type <tok>   SYMBOLx SYMsubscr labelname

// Nonterminal types — top-level
%type <accum> top

// Nonterminal types — declarations
%type <node>  labeledfmla lgprop gprop
%type <node>  opttemporal optunprovable optexplicit optlabel optskolem optinit
%type <node>  optproof
%type <node>  defn defnlhs defnrhs typeddefn gdefn schdefn schdefnrhs schconc
%type <node>  defarg somevarfmla
%type <nodes> defns defargs
%type <nodes> schdecl schdecls
%type <node>  symdecl constantdecl parameter paramval
%type <node>  tapp tterm
%type <nodes> tterms targs tsyms
%type <node>  tatom
%type <nodes> tatoms
%type <node>  rel fun
%type <nodes> rels funs
%type <node>  sort typesymbol
%type <bval>  optfinite optghost
%type <nodes> names
%type <str>   dotsym
%type <node>  optin optelse

// Nonterminal types — actions
%type <node>  action simpleact complexact sequence topseq
%type <nodes> actseq actseqrev
%type <node>  lparam
%type <nodes> lparams
%type <node>  optactiondef
%type <bval>  actmeth
%type <node>  optimpex
%type <nodes> optargs optreturns optactualreturns
%type <node>  optsemi

// Nonterminal types — callatom
%type <node>  atom callatom
%type <nodes> callatoms

// Nonterminal types — module/object
%type <node>  modulestart moduleend objectend modcat opteq objsym
%type <bval>  optdotdotdot opttrusted
%type <nodes> objectargs
%type <node>  param
%type <nodes> params
%type <nodes> optwith

// Nonterminal types — instantiate
%type <node>  inst modinst pname
%type <nodes> insts pnames

// Nonterminal types — interpret/misc
%type <node>  oper attributeval
%type <nodes> moresymbols

// Nonterminal types — specimpl
%type <str>   specimpl

// Nonterminal types — relop/infix
%type <str>   relop infix

// Nonterminal types — app
%type <node>  app
%type <nodes> apps

// Nonterminal types — lit
%type <node>  lit

// Nonterminal types — atoms
%type <nodes> atoms

// Nonterminal types — scenario
%type <node>  sceninit scenariomixin scentrans
%type <nodes> scentranss places

// Nonterminal types — proof/tactic
%type <node>  proofstep proofseq proofgroup optproofgroup opttacticwith
%type <node>  tacticwithlistchoice tacticwithelem
%type <nodes> tacticwithlist pflets
%type <node>  pflet

// Nonterminal types — somefmla/while extensions
%type <node>  somefmla
%type <nodes> bounds invariants decreases

// Nonterminal types — match/renaming
%type <node>  match renamingitem renaming optrenaming unfspec
%type <nodes> matches renaminglist renamings unfspecs

// Nonterminal types — update
%type <node>  requires ensures modifies
%type <nodes> upaxes
%type <node>  upax

// Nonterminal types — concept space
%type <node>  expr exprterm cdefn
%type <nodes> cdefns prod sum

// Nonterminal types — state
%type <node>  state_expr assert_rhs

// Nonterminal types — delegate
%type <node>  optdelegee

// Nonterminal types — eqn/let
%type <node>  eqn
%type <nodes> eqns

// Nonterminal types — termtuple
%type <node>  termtuple

// Nonterminal types — debug
%type <node>  debugarg
%type <nodes> debugargs optdebugargs

// Nonterminal types — loc
%type <str>   loc

// Nonterminal types — symbols list
%type <nodes> symbols

// Precedence declarations — copied exactly from Python v1.7+ precedence table.
%left         TOK_SEMI
%left         TOK_GLOBALLY TOK_EVENTUALLY
%left         TOK_ARROW TOK_IFF
%left         TOK_OR
%left         TOK_AND
%left         TOK_TILDA
%left         TOK_EQ TOK_LE TOK_LT TOK_GE TOK_GT TOK_PTO
%left         TOK_TILDAEQ
%left         TOK_IF
%left         TOK_ELSE
%left         TOK_COLON
%left         TOK_PLUS TOK_MINUS
%left         TOK_TIMES TOK_DIV
%left         TOK_DOLLAR
%left         TOK_OLD
%left         TOK_DOT
// Note: Python has no ASSIGN or ISA in precedence, but goyacc needs them
// to resolve shift/reduce conflicts in action rules. These are harmless
// since ASSIGN/ISA never appear in ambiguous expression positions.
%right        TOK_ASSIGN

%start        top

%%

// ============================================================
// top: The Ivy file top-level. Matches Python p_top (ivy_parser.py:361).
// The accumulator (ivyAccum) collects declarations as parsing proceeds.
// ============================================================

top:
    /* empty */
    {
        xtracer.Trace("parser.p_top ENTER (top)")
        lex := v17lex.(*v17LexAdapter)
        parent := lex.accum // nil for outermost top
        $$ = newIvyAccum(parent, lex.parentObjName)
        lex.parentObjName = "" // consumed
        $$.parent = parent
        // Ensure astCfg is set from lex adapter (for first accum where parent is nil)
        if $$.astCfg == nil {
            $$.astCfg = lex.astCfg
        }
        // Python: self.attributes = ((special_attribute,) if special_attribute else ()) +
        //                          ((global_attribute,) if global_attribute else ()) +
        //                          ((common_attribute,) if common_attribute else ())
        // Consume all three attribute slots set by specimpl rules
        if lex.specialAttribute != "" {
            $$.attributes = append($$.attributes, lex.specialAttribute)
            lex.specialAttribute = ""
        }
        if lex.globalAttribute != "" {
            $$.attributes = append($$.attributes, lex.globalAttribute)
            lex.globalAttribute = ""
        }
        if lex.commonAttribute != "" {
            $$.attributes = append($$.attributes, lex.commonAttribute)
            lex.commonAttribute = ""
        }
        // Python: if self.attributes and stack: self.attributes = stack[-1].attributes + self.attributes
        if len($$.attributes) > 0 && parent != nil {
            $$.attributes = append(append([]string{}, parent.attributes...), $$.attributes...)
        }
        lex.accum = $$
    }
    | top TOK_USING SYMBOLx
    {
        xtracer.Trace("parser.p_top_using_symbol ENTER (top)")
        $$ = $1
        // Python: importer(p[3]) and merge decls — deferred to post-parse
    }
    | top TOK_INCLUDE SYMBOLx
    {
        xtracer.Trace("parser.p_top_include_symbol ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        name := $3.Val
        // Python: if not any(p[3] in m.included for m in stack):
        // Walk the parent chain to check ALL scopes' included sets.
        alreadyIncluded := false
        for cur := lex.accum; cur != nil; cur = cur.parent {
            if cur.included[name] {
                alreadyIncluded = true
                break
            }
        }
        if !alreadyIncluded {
            $$.included[name] = true
            // Python: pref = Atom(p[3],[]); pref.lineno = get_lineno(p,2)
            pref := acfg(v17lex).NewAtom(name)
            pref.SetLineno(tokLineno(lex, $2))
            xtracer.Trace("parser.include ENTER name=%s", name)
            // Python: parent_object = "this" — in Python this affects the nested
            // parse's Ivy.__init__. In Go, imports use a separate parser invocation
            // so this global doesn't propagate to the imported parser.
            modDeclCount := 0
            if lex.importer != nil {
                mod, err := lex.importer(name)
                if err != nil {
                    xtracer.Trace("parser.include ERROR name=%s err=%v", name, err)
                } else if mod != nil {
                    modDeclCount = len(mod.Decls)
                    // Python: for decl in module.decls: p[0].declare(decl, allow_redef=True)
                    for _, d := range mod.Decls {
                        $$.declare(d)
                    }
                    // Python: p[0].included.update(module.included)
                    if $$.included == nil {
                        $$.included = make(map[string]bool)
                    }
                    for k, v := range mod.Included {
                        $$.included[k] = v
                    }
                    // Python: p[0].modules.update(module.modules)
                    for k, v := range mod.Modules {
                        $$.modules[k] = v
                    }
                }
            }
            // Python: len(module.decls) — count from the imported module, not outer accum
            xtracer.Trace("parser.include EXIT name=%s decls=%d", name, modDeclCount)
        }
    }
    // --- Axiom (v1.7+): top optexplicit opttemporal AXIOM lgprop ---
    | top optexplicit opttemporal TOK_AXIOM lgprop
    {
        xtracer.Trace("parser.p_top_axiom_optlabel_gprop ENTER (top)")
        $$ = $1
        lf := addLabel(acfg(v17lex), $5.(*ast.LabeledFormula), "axiom")
        // Python: lf = addexplicit(lf) if p[2] else lf  (explicit first)
        if $2 != nil {
            lf = addExplicit(lf)
        }
        // Python: d = AxiomDecl(addtemporal(lf) if p[3] else check_non_temporal(lf))
        if $3 != nil {
            lf = addTemporal(lf)
        } else {
            checkNonTemporal(lf)
        }
        d := acfg(v17lex).NewAxiomDecl(lf)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $4))
        $$.declare(d)
    }
    // --- Property (v1.7+): top optexplicit opttemporal PROPERTY labeledfmla optskolem optproof ---
    | top optexplicit opttemporal TOK_PROPERTY labeledfmla optskolem optproof
    {
        xtracer.Trace("parser.p_top_property_labeledfmla ENTER (top)")
        $$ = $1
        lf := addLabel(acfg(v17lex), $5.(*ast.LabeledFormula), "prop")
        // Python: lf = addtemporal(lf) if p[3] else check_non_temporal(lf)
        if $3 != nil {
            lf = addTemporal(lf)
        } else {
            checkNonTemporal(lf)
        }
        // Python: lf = addexplicit(lf) if p[2] else lf
        if $2 != nil {
            lf = addExplicit(lf)
        }
        d := acfg(v17lex).NewPropertyDecl(lf)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $4))
        $$.declare(d)
        if $6 != nil {
            $$.declare(acfg(v17lex).NewNamedDecl($6))
        }
        if $7 != nil {
            $$.declare(acfg(v17lex).NewProofDecl($7))
        }
    }
    // --- Conjecture ---
    | top TOK_CONJECTURE labeledfmla
    {
        xtracer.Trace("parser.p_top_conjecture_labeledfmla ENTER (top)")
        $$ = $1
        lf := addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "conj")
        d := acfg(v17lex).NewConjectureDecl(lf)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$.declare(d)
    }
    // --- Invariant (v1.7+): top optexplicit INVARIANT labeledfmla optproof ---
    | top optexplicit TOK_INVARIANT labeledfmla optproof
    {
        xtracer.Trace("parser.p_top_invariant_labeledfmla ENTER (top)")
        $$ = $1
        lf := addLabel(acfg(v17lex), $4.(*ast.LabeledFormula), "invar")
        lf.Unprovable = false
        if $2 != nil {
            lf.Explicit = true
        }
        d := acfg(v17lex).NewConjectureDecl(lf)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
        $$.declare(d)
        if $5 != nil {
            $$.declare(acfg(v17lex).NewProofDecl($5))
        }
    }
    // --- Unprovable Invariant ---
    | top TOK_UNPROVABLE TOK_INVARIANT labeledfmla optproof
    {
        xtracer.Trace("parser.p_top_unprovable_invariant_labeledfmla ENTER (top)")
        $$ = $1
        lf := addLabel(acfg(v17lex), $4.(*ast.LabeledFormula), "invar")
        lf.Unprovable = true
        lf.Explicit = true
        d := acfg(v17lex).NewConjectureDecl(lf)
        // Python: if not lf.unprovable or check_unprovable.get(): p[0].declare(d); declare(ProofDecl)
        if !lf.Unprovable || acfg(v17lex).CheckUnprovable {
            $$.declare(d)
            if $5 != nil {
                $$.declare(acfg(v17lex).NewProofDecl($5))
            }
        }
    }
    // --- Module: top MODULE modulestart modcat atom optwith EQ LCB top RCB moduleend ---
    | top TOK_MODULE modulestart modcat atom optwith TOK_EQ TOK_LCB top TOK_RCB moduleend
    {
        xtracer.Trace("parser.p_top_module_atom_eq_lcb_top_rcb ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        modAccum := $9
        // Store ivyAccum directly as module body, matching Python where p[9] (Ivy instance)
        // is stored in Definition(name, ivy_instance). This preserves .objects, .defined, .static.
        d := acfg(v17lex).NewDefinition(ast.AppToAtom($5), modAccum)
        $$.declare(acfg(v17lex).NewModuleDecl(d))
        // Python: if p[4] == "isolate": ... with get_lineno(p,2) on this, iso, d.args[0], d
        if $4 != nil {
            if catAtom, ok := $4.(*ast.Atom); ok && catAtom.Rep == "isolate" {
                thisAtom := acfg(v17lex).NewAtom("this")
                thisAtom.SetLineno(tokLineno(lex, $2))
                iso := acfg(v17lex).NewAtom("iso")
                iso.SetLineno(tokLineno(lex, $2))
                isoElems := append([]ast.Node{iso, thisAtom}, $6...)
                isoDef := acfg(v17lex).NewIsolateDef(isoElems, len($6))
                isoDef.SetLineno(tokLineno(lex, $2))
                isoDecl := acfg(v17lex).NewIsolateDecl(isoDef)
                isoDecl.Attributes = []ast.Node{acfg(v17lex).NewAtom("common")}
                isoDecl.SetLineno(tokLineno(lex, $2))
                modAccum.declare(isoDecl)
            }
        }
        // Python: stack.pop() — restore scope after processing module body
        lex.accum = $$
        $$.isModule = false // Python: stack[-1].is_module = False
    }
    // --- Object ---
    | top TOK_OBJECT objsym objectargs TOK_EQ TOK_LCB optdotdotdot top TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top_object_symbol_eq_lcb_top_rcb ENTER (top)")
        $$ = $1
        objAccum := $8
        pref := $3.(*ast.Atom)
        lineno := nodeLineno($3)
        createObject(acfg(v17lex), $$, pref, $4, objAccum, lineno, $7)
        // Python: create_object does stack.pop() at ivy_parser.py:692
        v17lex.(*v17LexAdapter).accum = $$
    }
    // --- Class ---
    | top TOK_CLASS objsym objectargs TOK_EQ TOK_LCB optdotdotdot top TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top__top_class_symbol_objectargs_eq_lcb_optdo ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        objAccum := $8

        // Python: scnst = Atom(This())
        // Python: scnst.lineno = get_lineno(p,2)
        scnst := acfg(v17lex).NewAtom("this")
        scnst.SetLineno(tokLineno(lex, $2))

        // Python: tdfn = TypeDef(scnst, UninterpretedSort())
        // Python: tdfn.lineno = get_lineno(p,2)
        tdfn := acfg(v17lex).NewTypeDef(scnst, acfg(v17lex).NewUninterpretedSortAST())
        tdfn.SetLineno(tokLineno(lex, $2))

        // Python: p[8].declare(TypeDecl(tdfn))
        objAccum.declare(acfg(v17lex).NewTypeDecl(tdfn))

        // Python: p[8].decls = [p[8].decls[-1]] + p[8].decls[:-1]
        if n := len(objAccum.decls); n > 1 {
            last := objAccum.decls[n-1]
            copy(objAccum.decls[1:], objAccum.decls[:n-1])
            objAccum.decls[0] = last
        }

        // Python: create_object(p[0], p[3], p[4], p[8], get_lineno(p,3), p[7])
        createObject(acfg(v17lex), $$, $3.(*ast.Atom), $4, objAccum, nodeLineno($3), $7)

        // Python: stack.pop() equivalent
        lex.accum = $$
    }
    // --- Subclass ---
    | top TOK_SUBCLASS objsym TOK_OF atype TOK_EQ TOK_LCB optdotdotdot top TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top__top_subclass_symbol_of_atype_eq_lcb_optd ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        objAccum := $9

        // Python: scnst = Atom(This())
        // Python: scnst.lineno = get_lineno(p,2)
        scnst := acfg(v17lex).NewAtom("this")
        scnst.SetLineno(tokLineno(lex, $2))

        // Python: tdfn = TypeDef(scnst, UninterpretedSort())
        // Python: tdfn.lineno = get_lineno(p,2)
        tdfn := acfg(v17lex).NewTypeDef(scnst, acfg(v17lex).NewUninterpretedSortAST())
        tdfn.SetLineno(tokLineno(lex, $2))

        // Python: p[9].declare(TypeDecl(tdfn))
        objAccum.declare(acfg(v17lex).NewTypeDecl(tdfn))

        // Python: vdfn = VariantDef(scnst, Atom(p[5]))
        // Python: p[9].declare(VariantDecl(vdfn))
        vdfn := acfg(v17lex).NewVariantDef(scnst, acfg(v17lex).NewAtom(ast.NodeRep($5)))
        objAccum.declare(acfg(v17lex).NewVariantDecl(vdfn))

        // Python: p[9].decls = p[9].decls[-2:] + p[9].decls[:-2]
        if n := len(objAccum.decls); n > 2 {
            rotated := make([]ast.Node, n)
            copy(rotated, objAccum.decls[n-2:])
            copy(rotated[2:], objAccum.decls[:n-2])
            objAccum.decls = rotated
        }

        // Python: create_object(p[0], p[3], [], p[9], get_lineno(p,3), p[8])
        createObject(acfg(v17lex), $$, $3.(*ast.Atom), []ast.Node{}, objAccum, nodeLineno($3), $8)

        // Python: stack.pop() equivalent
        lex.accum = $$
    }
    // --- Definition (v1.7+): top optexplicit DEFINITION optlabel gdefn optproof ---
    | top optexplicit TOK_DEFINITION optlabel gdefn optproof
    {
        xtracer.Trace("parser.p_top_definition_optlabel_gdefn_optproof ENTER (top)")
        $$ = $1
        // Python: foo = p[5]
        // Python: if p[2]: foo = DefinitionSchema(*foo.args); foo.lineno = p[5].lineno
        gdefn := $5
        if $2 != nil { // optexplicit is True
            if def, ok := gdefn.(*ast.Definition); ok {
                ds := acfg(v17lex).NewDefinitionSchema(*def)
                ds.SetLineno(def.GetLineno())
                gdefn = ds
            }
        }
        lf := acfg(v17lex).NewLabeledFormula($4, gdefn)
        lf.Lineno = tokLineno(v17lex.(*v17LexAdapter), $3).Line
        lf = addLabel(acfg(v17lex), lf, "def")
        dd := acfg(v17lex).NewDefinitionDecl(lf)
        $$.declare(dd)
        if $6 != nil {
            $$.declare(acfg(v17lex).NewProofDecl($6))
        }
    }
    // --- Schema ---
    | top TOK_SCHEMA schdefn
    {
        xtracer.Trace("parser.p_top_schema_defn ENTER (top)")
        $$ = $1
        sch := acfg(v17lex).NewSchema($3)
        sd := acfg(v17lex).NewSchemaDecl(sch)
        $$.declare(sd)
    }
    // --- Theorem with schdefn ---
    | top TOK_THEOREM schdefn optproof
    {
        xtracer.Trace("parser.p_top_theorem_defn ENTER (top)")
        $$ = $1
        sch := acfg(v17lex).NewSchema($3)
        td := acfg(v17lex).NewTheoremDecl(sch)
        $$.declare(td)
        if $4 != nil {
            $$.declare(acfg(v17lex).NewProofDecl($4))
        }
    }
    // --- Theorem with LABEL schdefnrhs ---
    | top TOK_THEOREM labelname schdefnrhs optproof
    {
        xtracer.Trace("parser.p_top_theorem_label_rhs ENTER (top)")
        $$ = $1
        label := acfg(v17lex).NewAtom($3.Val)
        label.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
        df := acfg(v17lex).NewDefinition(label, $4)
        df.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
        sch := acfg(v17lex).NewSchema(df)
        td := acfg(v17lex).NewTheoremDecl(sch)
        $$.declare(td)
        if $5 != nil {
            $$.declare(acfg(v17lex).NewProofDecl($5))
        }
    }
    // --- Proof LABEL proofstep ---
    | top TOK_PROOF labelname proofstep
    {
        xtracer.Trace("parser.p_top_proof_label_label_proofstep ENTER (top)")
        $$ = $1
        label := acfg(v17lex).NewAtom($3.Val)
        label.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
        lf := acfg(v17lex).NewLabeledFormula(label, $4)
        $$.declare(acfg(v17lex).NewProofDecl(lf))
    }
    // --- Instantiate ---
    | top TOK_INSTANTIATE insts
    {
        xtracer.Trace("parser.p_top_instantiate_insts ENTER (top)")
        $$ = $1
        doInsts($$, $3)
    }
    // --- Autoinstance ---
    | top TOK_AUTOINSTANCE insts
    {
        xtracer.Trace("parser.p_top_autoinstance_insts ENTER (top)")
        $$ = $1
        d := acfg(v17lex).NewAutoInstanceDecl($3...)
        $$.declare(d)
    }
    // --- symdecl ---
    | top symdecl
    {
        xtracer.Trace("parser.p_top_symdecl ENTER (top)")
        $$ = $1
        $$.declare($2)
    }
    // --- Relation ---
    | top TOK_RELATION rels
    {
        xtracer.Trace("parser.p_top_relation_rels ENTER (top)")
        $$ = $1
        for _, d := range $3 {
            $$.declare(d)
        }
    }
    // --- Function ---
    | top TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_top_function_tapp_colon_atype ENTER (top)")
        $$ = $1
        for _, d := range $3 {
            $$.declare(d)
        }
    }
    // --- Derived ---
    | top TOK_DERIVED defns
    {
        xtracer.Trace("parser.p_top_derived_defns ENTER (top)")
        $$ = $1
        args := make([]ast.Node, len($3))
        for i, x := range $3 {
            args[i] = addLabel(acfg(v17lex), mkLF(acfg(v17lex), x), "def")
        }
        dd := acfg(v17lex).NewDerivedDecl(args...)
        $$.declare(dd)
    }
    // --- Type (uninterpreted) ---
    | top optfinite optghost TOK_TYPE typesymbol
    {
        xtracer.Trace("parser.p_top_type_symbol ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        scnst := acfg(v17lex).NewAtom($5.(*ast.Atom).Rep)
        scnst.SetLineno(nodeLineno($5))
        // Python: tdfn = (GhostTypeDef if p[3] else TypeDef)(scnst, UninterpretedSort())
        var tdfnNode ast.Node
        if $3 { // optghost
            gt := acfg(v17lex).NewGhostTypeDef(*acfg(v17lex).NewTypeDef(scnst, acfg(v17lex).NewUninterpretedSortAST()))
            if $2 { gt.Finite = true }
            gt.SetLineno(tokLineno(lex, $4))
            tdfnNode = gt
        } else {
            tdfn := acfg(v17lex).NewTypeDef(scnst, acfg(v17lex).NewUninterpretedSortAST())
            if $2 { tdfn.Finite = true }
            tdfn.SetLineno(tokLineno(lex, $4))
            tdfnNode = tdfn
        }
        td := acfg(v17lex).NewTypeDecl(tdfnNode)
        $$.declare(td)
    }
    // --- Type with sort ---
    | top optfinite optghost TOK_TYPE typesymbol TOK_EQ sort
    {
        xtracer.Trace("parser.p_top_type_symbol_eq_sort ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        scnst := acfg(v17lex).NewAtom($5.(*ast.Atom).Rep)
        scnst.SetLineno(nodeLineno($5))

        // Python: defsort = UninterpretedSort() if isinstance(p[7], Range) else p[7]
        sortNode := $7
        _, isRange := sortNode.(*ast.Range)
        if isRange {
            sortNode = acfg(v17lex).NewUninterpretedSortAST()
        }

        // Python: tdfn = (GhostTypeDef if p[3] else TypeDef)(scnst, defsort)
        tdfn := acfg(v17lex).NewTypeDef(scnst, sortNode)
        if $2 { tdfn.Finite = true }
        tdfn.SetLineno(tokLineno(lex, $6))
        var tdfnNode ast.Node = tdfn
        if $3 {
            tdfnNode = acfg(v17lex).NewGhostTypeDef(*tdfn)
        }
        td := acfg(v17lex).NewTypeDecl(tdfnNode)
        $$.declare(td)

        // Python: if isinstance(p[7], Range): ...
        if isRange {
            imp := acfg(v17lex).NewImplies(scnst, $7)
            imp.SetLineno(tokLineno(lex, $4))
            lf := mkLF(acfg(v17lex), imp)
            lf.SetLineno(imp.GetLineno())
            labeled := addLabel(acfg(v17lex), lf, "interp")
            thing := acfg(v17lex).NewInterpretDecl(labeled)
            thing.SetLineno(tokLineno(lex, $4))
            $$.declare(thing)
        }
    }
    // --- Progress ---
    | top TOK_PROGRESS defns
    {
        xtracer.Trace("parser.p_top_progress_defns ENTER (top)")
        $$ = $1
        pd := acfg(v17lex).NewProgressDecl($3...)
        $$.declare(pd)
    }
    // --- Rely ---
    | top TOK_RELY atom TOK_ARROW atom
    {
        xtracer.Trace("parser.p_top_rely_atom_arrow_atom ENTER (top)")
        $$ = $1
        imp := acfg(v17lex).NewImplies($3, $5)
        rd := acfg(v17lex).NewRelyDecl(imp)
        $$.declare(rd)
    }
    | top TOK_RELY atom
    {
        xtracer.Trace("parser.p_top_rely_atom ENTER (top)")
        $$ = $1
        rd := acfg(v17lex).NewRelyDecl($3)
        $$.declare(rd)
    }
    // --- Mixord ---
    | top TOK_MIXORD callatom TOK_ARROW callatom
    {
        xtracer.Trace("parser.p_top_mixord_callatom_arrow_callatom ENTER (top)")
        $$ = $1
        imp := acfg(v17lex).NewImplies($3, $5)
        md := acfg(v17lex).NewMixOrdDecl(imp)
        $$.declare(md)
    }
    // --- Concept ---
    | top TOK_CONCEPT cdefns
    {
        xtracer.Trace("parser.p_top_concept_cdefns ENTER (top)")
        $$ = $1
        cd := acfg(v17lex).NewConceptDecl($3...)
        $$.declare(cd)
    }
    // --- Update ---
    | top TOK_UPDATE apps TOK_FROM apps upaxes
    {
        xtracer.Trace("parser.p_top_update_terms_from_terms_upaxes ENTER (top)")
        $$ = $1
        // Simplified: store as raw nodes
        _ = $3
        _ = $5
        _ = $6
    }
    // --- Macro ---
    | top TOK_MACRO atom TOK_EQ sequence
    {
        xtracer.Trace("parser.p_top_macro_atom_eq_lcb_action_rcb ENTER (top)")
        $$ = $1
        d := acfg(v17lex).NewDefinition(ast.AppToAtom($3), $5)
        md := acfg(v17lex).NewMacroDecl(d)
        $$.declare(md)
    }
    // --- Action (v1.7+): top optimpex actmeth SYMBOL optargs optreturns optactiondef ---
    | top optimpex actmeth SYMBOLx optargs optreturns optactiondef
    {
        xtracer.Trace("parser.p_top_optimpex_action_symbol_optargs_optreturns_eq_action ENTER (top)")
        $$ = $1
        // Python: adef = p[7]; if not hasattr(adef,'lineno'): adef.lineno = get_lineno(p,4)
        // Python almost always has lineno set, so get_lineno rarely fires.
        // Match Python by checking HasLoc on the base node.
        adef := $7
        var lineno ast.Location
        if adef != nil {
            if b, ok := adef.(interface{ HasLocSet() bool }); ok && b.HasLocSet() {
                lineno = adef.GetLineno()
            }
        }
        if lineno == (ast.Location{}) {
            lineno = tokLineno(v17lex.(*v17LexAdapter), $4)
        }

        // Python: formals = p[5]
        formals := $5

        // Python: if p[3]: (actmeth is True for METHOD)
        if $3 {
            // Python: arg0 = App('self')
            // Python: arg0.sort = This()
            // Python: arg0.lineno = get_lineno(p,4)
            selfArg := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("self", nil))
            selfArg.ASort = acfg(v17lex).NewThis()
            selfArg.SetLineno(lineno)
            // Python: formals = [arg0] + formals
            formals = append([]ast.Node{selfArg}, formals...)
        }

        // Python: if isinstance(adef, CrashAction):
        //             adef = adef.clone([Atom(This(), formals)])
        if ca, ok := adef.(*ast.CrashAction); ok {
            thisAtom := acfg(v17lex).NewAtom("this", formals...)
            thisAtom.SetLineno(lineno)
            adef = ca.Clone([]ast.Node{thisAtom})
        }

        theAtom := acfg(v17lex).NewAtom($4.Val)
        theAtom.SetLineno(lineno)
        // Python: ActionDef(theAtom, adef, formals=formals, returns=p[6])
        actdef := acfg(v17lex).NewActionDef(theAtom, adef, formals, $6)
        actdef.SetLineno(lineno)
        decl := acfg(v17lex).NewActionDecl(actdef)
        decl.SetLineno(lineno)
        $$.declare(decl)
        // If export/import was specified
        if $2 != nil {
            $$.declare($2)
        }
    }
    // --- Mixin before ---
    | top TOK_MIXIN callatom TOK_BEFORE callatom
    {
        xtracer.Trace("parser.p_top_mixin_callatom_before_callatom ENTER (top)")
        $$ = $1
        m := acfg(v17lex).NewMixinBeforeDef($3, $5)
        md := acfg(v17lex).NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Mixin after ---
    | top TOK_MIXIN callatom TOK_AFTER callatom
    {
        xtracer.Trace("parser.p_top_mixin_callatom_after_callatom ENTER (top)")
        $$ = $1
        m := acfg(v17lex).NewMixinAfterDef($3, $5)
        md := acfg(v17lex).NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Before ---
    | top TOK_BEFORE atype optargs optreturns sequence
    {
        xtracer.Trace("parser.p_top_before_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := acfg(v17lex).NewAtom($3.(*ast.Symbol).Rep)
        atom.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        handleBeforeAfter(acfg(v17lex), "before", atom, $6, $$, $4, $5)
    }
    // --- After ---
    | top TOK_AFTER atype optargs optreturns topseq
    {
        xtracer.Trace("parser.p_top_after_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := acfg(v17lex).NewAtom($3.(*ast.Symbol).Rep)
        atom.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        handleBeforeAfter(acfg(v17lex), "after", atom, $6, $$, $4, $5)
    }
    // --- Around ---
    | top TOK_AROUND atype optargs optreturns TOK_LCB actseq optsemi TOK_DOTDOTDOT actseq optsemi TOK_RCB
    {
        xtracer.Trace("parser.p_top_around_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := acfg(v17lex).NewAtom($3.(*ast.Symbol).Rep)
        before := lalrMakeSequence(acfg(v17lex), $7)
        after := lalrMakeSequence(acfg(v17lex), $10)
        // before mixin
        acfg(v17lex).LabelCounter++
        bmixer := acfg(v17lex).NewAtom(fmt.Sprintf("%s[before%d]", atom.Rep, acfg(v17lex).LabelCounter))
        bdf := acfg(v17lex).NewActionDef(bmixer, before, $4, $5)
        bdecl := acfg(v17lex).NewActionDecl(bdf)
        $$.declare(bdecl)
        bm := acfg(v17lex).NewMixinBeforeDef(bmixer, atom)
        bmd := acfg(v17lex).NewMixinDecl(bm)
        $$.declare(bmd)
        // after mixin
        acfg(v17lex).LabelCounter++
        amixer := acfg(v17lex).NewAtom(fmt.Sprintf("%s[after%d]", atom.Rep, acfg(v17lex).LabelCounter))
        adf := acfg(v17lex).NewActionDef(amixer, after, $4, $5)
        adecl := acfg(v17lex).NewActionDecl(adf)
        $$.declare(adecl)
        am := acfg(v17lex).NewMixinAfterDef(amixer, atom)
        amd := acfg(v17lex).NewMixinDecl(am)
        $$.declare(amd)
    }
    // --- After init ---
    | top TOK_AFTER TOK_INIT optargs topseq
    {
        xtracer.Trace("parser.p_top_after_init_optargs_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := acfg(v17lex).NewAtom("init")
        atom.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        handleBeforeAfter(acfg(v17lex), "after", atom, $5, $$, $4, nil)
    }
    // --- Implement ---
    | top TOK_IMPLEMENT atype optargs optreturns topseq
    {
        xtracer.Trace("parser.p_top_implement_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := acfg(v17lex).NewAtom($3.(*ast.Symbol).Rep)
        atom.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        handleBeforeAfter(acfg(v17lex), "implement", atom, $6, $$, $4, $5)
    }
    // --- Implement type ---
    | top TOK_IMPLEMENT TOK_TYPE SYMBOLx TOK_WITH SYMBOLx
    {
        xtracer.Trace("parser.p_top_implement_type_symbol_with_symbol ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        // Python: a1,a2 = Atom(p[4]),Atom(p[6])
        // Python: a1.lineno = get_lineno(p,4); a2.lineno = get_lineno(p,6)
        a1 := acfg(v17lex).NewAtom($4.Val)
        a1.SetLineno(tokLineno(lex, $4))
        a2 := acfg(v17lex).NewAtom($6.Val)
        a2.SetLineno(tokLineno(lex, $6))
        // Python: impl = ImplementTypeDef(a1,a2); impl.lineno = get_lineno(p,5)
        impl := acfg(v17lex).NewImplementTypeDef([]ast.Node{a1, a2})
        impl.SetLineno(tokLineno(lex, $5))
        // Python: d = ImplementTypeDecl(mk_lf(impl)); d.lineno = get_lineno(p,2)
        d := acfg(v17lex).NewImplementTypeDecl(mkLF(acfg(v17lex), impl))
        d.SetLineno(tokLineno(lex, $2))
        $$.declare(d)
    }
    // --- Isolate ---
    | top opttrusted TOK_ISOLATE SYMBOLx optargs TOK_EQ callatoms
    {
        xtracer.Trace("parser.p_top_opttrusted_isolate_callatom_eq_callatoms ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        // Python: ty = TrustedIsolateDef if p[2] else IsolateDef
        // Python: d = IsolateDecl(ty(*([Atom(p[4],p[5])] + p[7])))
        nameAtom := acfg(v17lex).NewAtom($4.Val, $5...)
        elems := append([]ast.Node{nameAtom}, $7...)
        idef := acfg(v17lex).NewIsolateDef(elems, 0)
        idef.Trusted = $2
        idef.Elems[0].SetLineno(tokLineno(lex, $3))
        idef.SetLineno(tokLineno(lex, $3))
        id := acfg(v17lex).NewIsolateDecl(idef)
        $$.declare(id)
    }
    // --- Isolate with WITH ---
    | top opttrusted TOK_ISOLATE SYMBOLx optargs TOK_EQ callatoms TOK_WITH callatoms
    {
        xtracer.Trace("parser.p_top_opttrusted_isolate_callatom_eq_callatoms_with_callatoms ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        // Python: ty = TrustedIsolateDef if p[2] else IsolateDef
        // Python: d = IsolateDecl(ty(*([Atom(p[4],p[5])] + p[7] + p[9])))
        nameAtom := acfg(v17lex).NewAtom($4.Val, $5...)
        elems := append(append([]ast.Node{nameAtom}, $7...), $9...)
        idef := acfg(v17lex).NewIsolateDef(elems, len($9))
        idef.Trusted = $2
        idef.Elems[0].SetLineno(tokLineno(lex, $3))
        idef.SetLineno(tokLineno(lex, $3))
        id := acfg(v17lex).NewIsolateDecl(idef)
        $$.declare(id)
    }
    // --- Isolate with body ---
    | top opttrusted TOK_ISOLATE SYMBOLx optargs TOK_EQ TOK_LCB top TOK_RCB optwith
    {
        xtracer.Trace("parser.p_top_opttrusted_isolate_callatom_eq_lcb_top_rcb_optwith ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        objAccum := $8
        // Python: create_object(p[0],p[4],p[5],p[8],get_lineno(p,4))
        nameAtom := acfg(v17lex).NewAtom($4.Val)
        createObject(acfg(v17lex), $$, nameAtom, $5, objAccum, tokLineno(lex, $4), false)
        // Python: ty = TrustedIsolateDef if p[2] else IsolateDef
        // Python: df = ty(*([Atom(p[4],p[5]),Atom(p[4],p[5])]+p[10]))
        a1 := acfg(v17lex).NewAtom($4.Val, $5...)
        a2 := acfg(v17lex).NewAtom($4.Val, $5...)
        elems := append([]ast.Node{a1, a2}, $10...)
        idef := acfg(v17lex).NewIsolateDef(elems, len($10))
        idef.Trusted = $2
        idef.IsObject = true
        idef.Elems[0].SetLineno(tokLineno(lex, $3))
        idef.SetLineno(tokLineno(lex, $3))
        id := acfg(v17lex).NewIsolateObjectDecl(*acfg(v17lex).NewIsolateDecl(idef))
        $$.declare(id)
        // Python: stack.pop() equivalent
        lex.accum = $$
    }
    // --- Extract with body ---
    | top TOK_EXTRACT objsym objectargs TOK_EQ TOK_LCB top TOK_RCB optwith
    {
        xtracer.Trace("parser.p_top_opttrusted_extract_callatom_eq_lcb_top_rcb_optwith ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        objAccum := $7
        pref := $3.(*ast.Atom)
        // Python: create_object(p[0],p[3],p[4],p[7],get_lineno(p,3))
        createObject(acfg(v17lex), $$, pref, $4, objAccum, nodeLineno($3), false)
        // Python: ty = ProcessDef
        // Python: d = IsolateObjectDecl(ty(*([Atom(p[3],p[4]),Atom(p[3],p[4])]+p[9])))
        a1 := acfg(v17lex).NewAtom(pref.Rep, $4...)
        a2 := acfg(v17lex).NewAtom(pref.Rep, $4...)
        elems := append([]ast.Node{a1, a2}, $9...)
        idefInner := acfg(v17lex).NewIsolateDef(elems, len($9)+1)
        idefInner.IsObject = true
        edef := acfg(v17lex).NewExtractDef(*idefInner)
        pdef := acfg(v17lex).NewProcessDef(*edef)
        pdef.Elems[0].SetLineno(tokLineno(lex, $2))
        pdef.SetLineno(tokLineno(lex, $2))
        id := acfg(v17lex).NewIsolateObjectDecl(*acfg(v17lex).NewIsolateDecl(&pdef.IsolateDef))
        $$.declare(id)
        // Python: stack.pop() equivalent
        lex.accum = $$
    }
    // --- Extract without body ---
    | top TOK_EXTRACT objsym objectargs TOK_EQ callatoms
    {
        xtracer.Trace("parser.p_top_extract_callatom_eq_callatoms ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        // Python: stack[-1].params = []; parent_object = None
        lex.accum.params = nil
        lex.parentObjName = ""
        // Python: d = IsolateDecl(ExtractDef(*([Atom(p[3],p[4])] + p[6])))
        pref := $3.(*ast.Atom)
        nameAtom := acfg(v17lex).NewAtom(pref.Rep, $4...)
        elems := append([]ast.Node{nameAtom}, $6...)
        edef := acfg(v17lex).NewExtractDef(*acfg(v17lex).NewIsolateDef(elems, len($6)))
        edef.Elems[0].SetLineno(tokLineno(lex, $2))
        edef.SetLineno(tokLineno(lex, $2))
        id := acfg(v17lex).NewIsolateDecl(&edef.IsolateDef)
        $$.declare(id)
    }
    // --- Export ---
    | top TOK_EXPORT callatom
    {
        xtracer.Trace("parser.p_top_export_callatom ENTER (top)")
        $$ = $1
        ed := acfg(v17lex).NewExportDecl(acfg(v17lex).NewExportDef($3, acfg(v17lex).NewAtom("")))
        ed.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$.declare(ed)
    }
    // --- Import ---
    | top TOK_IMPORT callatom
    {
        xtracer.Trace("parser.p_top_import_callatom ENTER (top)")
        $$ = $1
        id := acfg(v17lex).NewImportDecl(acfg(v17lex).NewImportDef($3, acfg(v17lex).NewAtom("")))
        id.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$.declare(id)
    }
    // --- Delegate ---
    | top TOK_DELEGATE callatoms optdelegee
    {
        xtracer.Trace("parser.p_top_delegate_callatom_opt ENTER (top)")
        $$ = $1
        args := make([]ast.Node, len($3))
        for i, s := range $3 {
            if $4 != nil {
                args[i] = acfg(v17lex).NewDelegateDef([]ast.Node{s, $4})
            } else {
                args[i] = acfg(v17lex).NewDelegateDef([]ast.Node{s})
            }
        }
        dd := acfg(v17lex).NewDelegateDecl(args...)
        $$.declare(dd)
    }
    // --- Interpret ---
    | top TOK_INTERPRET oper TOK_ARROW oper
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_symbol ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        imp := acfg(v17lex).NewImplies($3, $5)
        imp.SetLineno(tokLineno(lex, $4))
        lf := addLabel(acfg(v17lex), mkLF(acfg(v17lex), imp), "interp")
        d := acfg(v17lex).NewInterpretDecl(lf)
        d.SetLineno(tokLineno(lex, $4))
        $$.declare(d)
    }
    // --- Interpret with range ---
    | top TOK_INTERPRET oper TOK_ARROW TOK_LCB term TOK_DOTS term TOK_RCB
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_lcb_symbol_dots_symbol_rcb ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        rng := acfg(v17lex).NewRange($6, $8)
        imp := acfg(v17lex).NewImplies($3, rng)
        imp.SetLineno(tokLineno(lex, $4))
        lf := addLabel(acfg(v17lex), mkLF(acfg(v17lex), imp), "interp")
        d := acfg(v17lex).NewInterpretDecl(lf)
        d.SetLineno(tokLineno(lex, $4))
        $$.declare(d)
    }
    // --- Interpret with enum ---
    | top TOK_INTERPRET oper TOK_ARROW TOK_LCB SYMBOLx moresymbols TOK_RCB
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_lcb_symbol_moresymbols_rcb ENTER (top)")
        $$ = $1
        names := append([]string{$6.Val}, func() []string {
            var r []string
            for _, n := range $7 {
                if a, ok := n.(*ast.Atom); ok {
                    r = append(r, a.Rep)
                }
            }
            return r
        }()...)
        atoms := make([]ast.Node, len(names))
        for i, n := range names {
            atoms[i] = acfg(v17lex).NewAtom(n)
        }
        es := acfg(v17lex).NewEnumeratedSort(atoms...)
        imp := acfg(v17lex).NewImplies($3, es)
        imp.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $4))
        lf := addLabel(acfg(v17lex), mkLF(acfg(v17lex), imp), "interp")
        d := acfg(v17lex).NewInterpretDecl(lf)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $4))
        $$.declare(d)
    }
    // --- Alias ---
    | top TOK_ALIAS SYMBOLx TOK_EQ callatom
    {
        xtracer.Trace("parser.p_top_aliase_symbol_eq_callatom ENTER (top)")
        $$ = $1
        d := acfg(v17lex).NewAliasDecl(acfg(v17lex).NewDefinition(acfg(v17lex).NewAtom($3.Val), $5))
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
        $$.declare(d)
    }
    // --- Attribute ---
    | top TOK_ATTRIBUTE callatom TOK_EQ attributeval
    {
        xtracer.Trace("parser.p_top_attribute_callatom_eq_attributeval ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        adef := acfg(v17lex).NewAttributeDef($3, $5)
        adef.SetLineno(tokLineno(lex, $2))
        d := acfg(v17lex).NewAttributeDecl(adef)
        d.SetLineno(tokLineno(lex, $2))
        $$.declare(d)
    }
    // --- Variant ---
    | top TOK_VARIANT typesymbol TOK_OF atype
    {
        xtracer.Trace("parser.p_top_variant_symbol_of_atype ENTER (top)")
        $$ = $1
        scnst := acfg(v17lex).NewAtom($3.(*ast.Atom).Rep)
        scnst.SetLineno(nodeLineno($3))
        tdfn := acfg(v17lex).NewTypeDef(scnst, acfg(v17lex).NewUninterpretedSortAST())
        tdfn.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $4))
        td := acfg(v17lex).NewTypeDecl(tdfn)
        $$.declare(td)
        vdfn := acfg(v17lex).NewVariantDef(scnst, $5)
        vd := acfg(v17lex).NewVariantDecl(vdfn)
        $$.declare(vd)
    }
    // --- Variant with sort ---
    | top TOK_VARIANT typesymbol TOK_OF atype TOK_EQ sort
    {
        xtracer.Trace("parser.p_top_variant_symbol_of_symbol_eq_sort ENTER (top)")
        $$ = $1
        scnst := acfg(v17lex).NewAtom($3.(*ast.Atom).Rep)
        scnst.SetLineno(nodeLineno($3))
        tdfn := acfg(v17lex).NewTypeDef(scnst, $7)
        tdfn.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $4))
        td := acfg(v17lex).NewTypeDecl(tdfn)
        $$.declare(td)
        vdfn := acfg(v17lex).NewVariantDef(scnst, $5)
        vd := acfg(v17lex).NewVariantDecl(vdfn)
        $$.declare(vd)
    }
    // --- Nativequote ---
    | top TOK_NATIVEQUOTE
    {
        xtracer.Trace("parser.p_top_nativequote ENTER (top)")
        $$ = $1
        // Python: text,bqs = parse_nativequote(p,2)
        // Python: defn = NativeDef(*([mk_label(None,'native')] + [text] + bqs))
        // Python: thing = NativeDecl(defn)
        text, bqs := parseNativequote(acfg(v17lex), $2.Val, v17lex.(*v17LexAdapter))
        label := newLabel(acfg(v17lex), "native")
        defnArgs := append([]ast.Node{label, acfg(v17lex).NewNativeCode(text)}, bqs...)
        defn := acfg(v17lex).NewNativeDef(defnArgs)
        defn.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        thing := acfg(v17lex).NewNativeDecl(defn)
        thing.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$.declare(thing)
    }
    // --- Scenario ---
    | top TOK_SCENARIO TOK_LCB sceninit TOK_SEMI scentranss TOK_RCB
    {
        xtracer.Trace("parser.p_top_scenario_lcb_sceninit_semi_scentranss_rcb ENTER (top)")
        $$ = $1
        elems := append([]ast.Node{$4}, $6...)
        sdef := acfg(v17lex).NewScenarioDef(elems)
        sdef.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        sd := acfg(v17lex).NewScenarioDecl(sdef)
        $$.declare(sd)
    }
    // --- Spec/Impl blocks ---
    | top specimpl TOK_LCB top TOK_RCB
    {
        xtracer.Trace("parser.p_top_specification_lcb_top_rcb ENTER (top)")
        $$ = $1
        innerAccum := $4
        // Python: stack.pop() — restore scope after processing specimpl body
        v17lex.(*v17LexAdapter).accum = $$
        // Python: temp clear outer attributes to prevent double-applying
        // temp_attr = p[0].attributes; p[0].attributes = ()
        saveAttrs := $$.attributes
        $$.attributes = nil
        for _, d := range innerAccum.decls {
            $$.declare(d)
        }
        // Python: p[0].attributes = temp_attr
        $$.attributes = saveAttrs
    }
    // --- State ---
    | top TOK_STATE SYMBOLx TOK_EQ state_expr
    {
        xtracer.Trace("parser.p_top_state_symbol_eq_state_expr ENTER (top)")
        $$ = $1
        sd := acfg(v17lex).NewStateDef($3.Val, $5)
        _ = sd
    }
    // --- Assert (v1.6 only, kept for compat) ---
    | top TOK_ASSERT SYMBOLx TOK_ARROW assert_rhs
    {
        xtracer.Trace("parser.p_top_assert_symbol_arrow_assert_rhs ENTER (top)")
        $$ = $1
        // Python: this rule is guarded by `if iu.get_numeric_version() <= [1,6]:`
        // Since Go grammar is v1.7+, the body is intentionally empty.
        // v1.6 would do: Implies(Atom(p[3],[]),p[5]) → AssertDecl → declare
    }
    ;

// ============================================================
// --- SYMBOL handling ---
// ============================================================

SYMBOLx:
    TOK_PRESYMBOL
    {
        xtracer.Trace("parser.p_SYMBOL_PRESYMBOL ENTER (SYMBOL) val=%s", $1.Val)
        $$ = $1
    }
    | SYMBOLx TOK_LB SYMsubscr TOK_RB
    {
        xtracer.Trace("parser.p_SYMBOL_SYMBOL_LB_SYMsubscr_RB ENTER (SYMBOL)")
        $$ = TokenInfo{Val: $1.Val + "[" + $3.Val + "]", Line: $1.Line}
    }
    ;

SYMsubscr:
    SYMBOLx
    {
        xtracer.Trace("parser.p_SYMsubscr_SYMBOL ENTER (SYMsubscr)")
        $$ = $1
    }
    | TOK_THIS
    {
        xtracer.Trace("parser.p_SYMsubscr_THIS ENTER (SYMsubscr)")
        $$ = TokenInfo{Val: "this", Line: $1.Line}
    }
    | SYMsubscr TOK_DOT SYMBOLx
    {
        xtracer.Trace("parser.p_SYMsubscr_SYMsubscr_dot_symbol ENTER (SYMsubscr)")
        $$ = TokenInfo{Val: $1.Val + "." + $3.Val, Line: $1.Line}
    }
    ;

// ============================================================
// --- atype (sort names) ---
// ============================================================

atype:
    SYMBOLx
    {
        xtracer.Trace("parser.p_atype_symbol ENTER (atype) val=%s", $1.Val)
        $$ = acfg(v17lex).NewSymbol($1.Val, nil)
    }
    | atype TOK_DOT SYMBOLx
    {
        xtracer.Trace("parser.p_atype_atype_dot_symbol ENTER (atype)")
        if _, ok := $1.(*ast.This); ok {
            $$ = acfg(v17lex).NewSymbol($3.Val, nil)
        } else if sym, ok := $1.(*ast.Symbol); ok {
            $$ = acfg(v17lex).NewSymbol(sym.Rep + "." + $3.Val, nil)
        } else {
            $$ = acfg(v17lex).NewSymbol($3.Val, nil)
        }
    }
    | TOK_THIS
    {
        xtracer.Trace("parser.p_atype_this ENTER (atype)")
        t := acfg(v17lex).NewThis()
        t.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = t
    }
    ;

// ============================================================
// --- appelem ---
// ============================================================

appelem:
    SYMBOLx
    {
        xtracer.Trace("parser.p_appelem_symbol ENTER (appelem)")
        // Python: App(p[1]) — appelem produces App, not Atom.
        a := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil))
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        xtracer.Trace("parser.p_appelem_appelem_terms ENTER (appelem)")
        // Python: App(p[1], p[3])
        a := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil), $3...)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    ;

// ============================================================
// --- aterm (for concept space exprterm) ---
// Matches Python aterm : appelem | aterm DOT appelem (ivy_logic_parser.py:193-202)
// ============================================================

aterm:
    appelem
    {
        xtracer.Trace("parser.p_aterm_aappelem ENTER (aterm)")
        $$ = $1
    }
    | aterm TOK_DOT appelem
    {
        xtracer.Trace("parser.p_aterm_aterm_dot_appelem ENTER (aterm)")
        $$ = ast.ComposeAtoms($1.(*ast.Atom), $3.(*ast.Atom))
    }
    ;

// ============================================================
// --- Variables ---
// ============================================================

var:
    TOK_VARIABLE
    {
        xtracer.Trace("parser.p_var_variable ENTER (var)")
        // Python: Variable(p[1], universe) where universe = 'S'
        v := acfg(v17lex).NewVariable($1.Val, "S")
        v.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = v
    }
    | TOK_VARIABLE TOK_COLON atype
    {
        xtracer.Trace("parser.p_var_variable_colon_symbol ENTER (var)")
        // Python: Variable(p[1], p[3]) where p[3] is a string from atype
        v := acfg(v17lex).NewVariable($1.Val, atypeToString($3))
        v.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = v
    }
    ;

simplevar:
    TOK_VARIABLE
    {
        xtracer.Trace("parser.p_simplevar_variable ENTER (simplevar)")
        // Python: Variable(p[1], universe) where universe = 'S'
        v := acfg(v17lex).NewVariable($1.Val, "S")
        v.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = v
    }
    | TOK_VARIABLE TOK_COLON SYMBOLx
    {
        xtracer.Trace("parser.p_simplevar_variable_colon_symbol ENTER (simplevar)")
        // Python: Variable(p[1], p[3]) where p[3] is a string
        v := acfg(v17lex).NewVariable($1.Val, $3.Val)
        v.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = v
    }
    ;

vars:
    var
    {
        xtracer.Trace("parser.p_vars_var ENTER (vars)")
        $$ = []ast.Node{$1}
    }
    | vars TOK_COMMA var
    {
        xtracer.Trace("parser.p_vars_vars_comma_var ENTER (vars)")
        $$ = append($1, $3)
    }
    ;

simplevars:
    simplevar
    {
        xtracer.Trace("parser.p_simplevars_simplevar ENTER (simplevars)")
        $$ = []ast.Node{$1}
    }
    | simplevars TOK_COMMA simplevar
    {
        xtracer.Trace("parser.p_simplevars_simplevars_comma_simplevar ENTER (simplevars)")
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Terms list ---
// ============================================================

terms:
    /* empty */
    {
        xtracer.Trace("parser.p_terms ENTER (terms)")
        $$ = nil
    }
    | term
    {
        xtracer.Trace("parser.p_terms_term ENTER (terms)")
        $$ = []ast.Node{$1}
    }
    | terms TOK_COMMA term
    {
        xtracer.Trace("parser.p_terms_terms_term ENTER (terms)")
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Terms (v1.7+: unified with formulas) ---
// ============================================================

term:
    appelem
    {
        xtracer.Trace("parser.p_term_aappelem ENTER (term)")
        $$ = $1
    }
    | var
    {
        xtracer.Trace("parser.p_term_var ENTER (term)")
        $$ = $1
    }
    | TOK_OLD appelem
    {
        xtracer.Trace("parser.p_term_old_aappelem ENTER (term)")
        o := acfg(v17lex).NewOld($2)
        o.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = o
    }
    | term TOK_DOT appelem
    {
        xtracer.Trace("parser.p_term_dot_appelem ENTER (term)")
        lex := v17lex.(*v17LexAdapter)
        // Python: if isinstance(p[1],(Atom,App)):
        //             p[0] = compose_atoms(p[1],p[3])
        //             p[0].lineno = get_lineno(p,2)
        //         elif isinstance(p[1],Old):
        //             t = compose_atoms(p[1].args[0],p[3])
        //             t.lineno = get_lineno(p,2)
        //             p[0] = p[1]; p[0].args[0] = t
        //         else:
        //             p[0] = MethodCall(p[1],p[3])
        //             p[0].lineno = get_lineno(p,2)
        switch lhs := $1.(type) {
        case *ast.Atom:
            composed := ast.ComposeAtomsGeneric(lhs, $3)
            composed.SetLineno(tokLineno(lex, $2))
            $$ = composed
        case *ast.App:
            composed := ast.ComposeAtomsGeneric(lhs, $3)
            composed.SetLineno(tokLineno(lex, $2))
            $$ = composed
        case *ast.Old:
            t := ast.ComposeAtomsGeneric(lhs.Term, $3)
            t.SetLineno(tokLineno(lex, $2))
            lhs.Term = t
            $$ = lhs
        default:
            mc := acfg(v17lex).NewMethodCall($1, $3)
            mc.SetLineno(tokLineno(lex, $2))
            $$ = mc
        }
    }
    | TOK_LPAREN term TOK_RPAREN
    {
        xtracer.Trace("parser.p_term_lp_term_lp ENTER (term)")
        $$ = $2
    }
    // --- Arithmetic ---
    | term TOK_PLUS term
    {
        xtracer.Trace("parser.p_term_term_PLUS_term ENTER (term)")
        n := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("+", nil), $1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    | term TOK_MINUS term
    {
        xtracer.Trace("parser.p_term_term_MINUS_term ENTER (term)")
        n := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("-", nil), $1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    | term TOK_TIMES term
    {
        xtracer.Trace("parser.p_term_term_TIMES_term ENTER (term)")
        n := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("*", nil), $1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    | term TOK_DIV term
    {
        xtracer.Trace("parser.p_term_term_DIV_term ENTER (term)")
        n := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("/", nil), $1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    // --- If/else ---
    | term TOK_IF fmla TOK_ELSE term
    {
        xtracer.Trace("parser.p_term_if_fmla_else_term ENTER (term)")
        n := acfg(v17lex).NewIte($3, $1, $5)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    // --- Comparison ---
    | term TOK_EQ term
    {
        xtracer.Trace("parser.p_term_term_EQ_term ENTER (term)")
        a := acfg(v17lex).NewAtom("=", $1, $3)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = a
    }
    | term TOK_LE term
    {
        xtracer.Trace("parser.p_term_term_LE_term ENTER (term)")
        a := acfg(v17lex).NewAtom("<=", $1, $3)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = a
    }
    | term TOK_LT term
    {
        xtracer.Trace("parser.p_term_term_LT_term ENTER (term)")
        a := acfg(v17lex).NewAtom("<", $1, $3)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = a
    }
    | term TOK_GE term
    {
        xtracer.Trace("parser.p_term_term_GE_term ENTER (term)")
        a := acfg(v17lex).NewAtom(">=", $1, $3)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = a
    }
    | term TOK_GT term
    {
        xtracer.Trace("parser.p_term_term_GT_term ENTER (term)")
        a := acfg(v17lex).NewAtom(">", $1, $3)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = a
    }
    | term TOK_PTO term
    {
        xtracer.Trace("parser.p_term_term_PTO_term ENTER (term)")
        n := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("*>", nil), $1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    | term TOK_TILDAEQ term
    {
        xtracer.Trace("parser.p_term_term_tildaeq_term ENTER (term)")
        n := acfg(v17lex).NewNot(acfg(v17lex).NewAtom("=", $1, $3))
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    // --- Boolean ---
    | TOK_TRUE
    {
        xtracer.Trace("parser.p_term_true ENTER (term)")
        n := acfg(v17lex).NewAnd()
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = n
    }
    | TOK_FALSE
    {
        xtracer.Trace("parser.p_term_false ENTER (term)")
        n := acfg(v17lex).NewOr()
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = n
    }
    | TOK_TILDA term
    {
        xtracer.Trace("parser.p_term_not_term ENTER (term)")
        n := acfg(v17lex).NewNot($2)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = n
    }
    | term TOK_AND term
    {
        xtracer.Trace("parser.p_term_term_and_term ENTER (term)")
        // Python: if isinstance(p[1],And): append; else: new And with get_lineno
        if existing, ok := $1.(*ast.And); ok {
            existing.Terms = append(existing.Terms, $3)
            $$ = existing
        } else {
            n := acfg(v17lex).NewAnd($1, $3)
            n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
            $$ = n
        }
    }
    | term TOK_OR term
    {
        xtracer.Trace("parser.p_term_term_or_term ENTER (term)")
        // Python: if isinstance(p[1],Or): append; else: new Or with get_lineno
        if existing, ok := $1.(*ast.Or); ok {
            existing.Terms = append(existing.Terms, $3)
            $$ = existing
        } else {
            n := acfg(v17lex).NewOr($1, $3)
            n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
            $$ = n
        }
    }
    | term TOK_ARROW term
    {
        xtracer.Trace("parser.p_term_term_arrow_term ENTER (term)")
        n := acfg(v17lex).NewImplies($1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    | term TOK_IFF term
    {
        xtracer.Trace("parser.p_term_term_iff_term ENTER (term)")
        n := acfg(v17lex).NewIff($1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    // --- Quantifiers ---
    | TOK_FORALL simplevars TOK_DOT term    %prec TOK_SEMI
    {
        xtracer.Trace("parser.p_term_forall_simplevars_dot_term ENTER (term)")
        n := acfg(v17lex).NewForall($2, $4)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = n
    }
    | TOK_EXISTS simplevars TOK_DOT term    %prec TOK_SEMI
    {
        xtracer.Trace("parser.p_term_exists_simplevars_dot_term ENTER (term)")
        n := acfg(v17lex).NewExists($2, $4)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = n
    }
    | TOK_FORALL TOK_LPAREN vars TOK_RPAREN term
    {
        xtracer.Trace("parser.p_term_forall_lp_vars_lp_term ENTER (term)")
        n := acfg(v17lex).NewForall($3, $5)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = n
    }
    | TOK_EXISTS TOK_LPAREN vars TOK_RPAREN term
    {
        xtracer.Trace("parser.p_term_exists_lp_vars_lp_term ENTER (term)")
        n := acfg(v17lex).NewExists($3, $5)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = n
    }
    // --- Temporal ---
    | TOK_GLOBALLY term
    {
        xtracer.Trace("parser.p_term_globally_term ENTER (term)")
        n := acfg(v17lex).NewGlobally($2)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = n
    }
    | TOK_EVENTUALLY term
    {
        xtracer.Trace("parser.p_term_eventually_term ENTER (term)")
        n := acfg(v17lex).NewEventually($2)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = n
    }
    | term TOK_WHENNEXT term
    {
        xtracer.Trace("parser.p_term_term_whennext_term ENTER (term)")
        n := acfg(v17lex).NewWhenOperator("next", $1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    | term TOK_WHENPREV term
    {
        xtracer.Trace("parser.p_term_term_whenprev_term ENTER (term)")
        n := acfg(v17lex).NewWhenOperator("prev", $1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    | term TOK_WHENFIRST term
    {
        xtracer.Trace("parser.p_term_term_whenfirst_term ENTER (term)")
        n := acfg(v17lex).NewWhenOperator("first", $1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    | term TOK_WHENLAST term
    {
        xtracer.Trace("parser.p_term_term_whenlast_term ENTER (term)")
        n := acfg(v17lex).NewWhenOperator("last", $1, $3)
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    // --- ISA ---
    | term TOK_ISA atype
    {
        xtracer.Trace("parser.p_fmla_fmla_isa_atype ENTER (term)")
        // Python: tp = Atom(p[3],[]); tp.lineno = get_lineno(p,2)
        // Python: p[0] = Isa(p[1],tp); p[0].lineno = get_lineno(p,2)
        tp := atypeToAtom(acfg(v17lex), $3)
        tp.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        isa := acfg(v17lex).NewIsa($1, tp)
        isa.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = isa
    }
    // --- Sort annotation ---
    | term TOK_COLON atype
    {
        xtracer.Trace("parser.p_term_term_colon_term ENTER (term)")
        // Python: if hasattr(p[1],"sort"): raise IvyError("multiple sort annotations")
        // Python: p[1].sort = p[3]; p[0] = p[1]
        switch n := $1.(type) {
        case *ast.Variable:
            if n.VSort != "" {
                v17lex.Error(fmt.Sprintf("multiple sort annotations on %v", n))
            }
            n.VSort = atypeToString($3)
        case *ast.Atom:
            if n.ASort != nil {
                v17lex.Error(fmt.Sprintf("multiple sort annotations on %v", n))
            }
            n.ASort = $3
        case *ast.App:
            if n.ASort != nil {
                v17lex.Error(fmt.Sprintf("multiple sort annotations on %v", n))
            }
            n.ASort = $3
        }
        $$ = $1
    }
    // --- Named binders ---
    | TOK_LPAREN TOK_DOLLAR SYMBOLx simplevars TOK_DOT fmla TOK_RPAREN TOK_LPAREN terms TOK_RPAREN
    {
        xtracer.Trace("parser.p_term_namedbinder_vars_dot_term ENTER (term)")
        // Python: x = NamedBinder(p[3], p[4], p[6]); x.lineno = get_lineno(p,2)
        // Python: p[0] = App(x, p[9]); p[0].lineno = get_lineno(p,2)
        binder := acfg(v17lex).NewNamedBinder($3.Val, $4, $6)
        binder.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = acfg(v17lex).NewApp(binder, $9...)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
    }
    | TOK_DOLLAR SYMBOLx TOK_DOT fmla     %prec TOK_SEMI
    {
        xtracer.Trace("parser.p_term_namedbinder_dot_fmla ENTER (term)")
        $$ = acfg(v17lex).NewNamedBinder($2.Val, nil, $4)
    }
    | TOK_DOLLAR SYMBOLx TOK_DOLLAR fmla   %prec TOK_SEMI
    {
        xtracer.Trace("parser.p_term_namedbinder_dollar_fmla ENTER (term)")
        $$ = acfg(v17lex).NewNamedBinder($2.Val, nil, $4)
    }
    ;

// ============================================================
// --- fmla ---
// ============================================================

fmla:
    term
    {
        xtracer.Trace("parser.p_fmla_term ENTER (fmla)")
        $$ = ast.AppToAtom($1)
    }
    ;

// ============================================================
// --- labeledfmla ---
// ============================================================

labeledfmla:
    fmla
    {
        xtracer.Trace("parser.p_labeledfmla_fmla ENTER (labeledfmla)")
        lf := acfg(v17lex).NewLabeledFormula(nil, $1)
        // Python: p[0].lineno = p[1].lineno — copy formula's Location
        lf.SetLineno($1.GetLineno())
        $$ = lf
    }
    | labelname fmla
    {
        xtracer.Trace("parser.p_labeledfmla_label_fmla ENTER (labeledfmla)")
        // Python: Atom(p[1][1:-1],[]) — brackets already stripped by labelname rule
        lf := acfg(v17lex).NewLabeledFormula(acfg(v17lex).NewAtom($1.Val), $2)
        // Python: p[0].lineno = get_lineno(p,1)
        lf.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = lf
    }
    ;

// labelname matches Python's LABEL : LB SYMBOL RB (ivy_logic_parser.py:23-25).
// The Python lexer produces LABEL as a terminal, but our lexer produces
// separate LB, SYMBOL, RB tokens, so we combine them in the grammar.
labelname:
    TOK_LB SYMBOLx TOK_RB
    {
        xtracer.Trace("parser.p_LABEL_LB_SYMBOL_RB ENTER (LABEL)")
        // Python: all LABEL consumers strip brackets with [1:-1].
        // Strip here at the source so all consumers get the bare name.
        $$ = TokenInfo{Val: $2.Val, Line: $1.Line}
    }
    | TOK_LABEL
    {
        xtracer.Trace("parser.p_labelname__label ENTER (labelname)")
        // TOK_LABEL comes from lexer with brackets "[name]"; strip them
        // to match Python's p[N][1:-1] pattern applied by all consumers.
        val := $1.Val
        if len(val) >= 2 && val[0] == '[' && val[len(val)-1] == ']' {
            val = val[1 : len(val)-1]
        }
        $$ = TokenInfo{Val: val, Line: $1.Line}
    }
    ;

// ============================================================
// --- lgprop / gprop (v1.7+) ---
// ============================================================

gprop:
    fmla
    {
        xtracer.Trace("parser.p_gprop_fmla ENTER (gprop)")
        $$ = $1
    }
    | schdefnrhs
    {
        xtracer.Trace("parser.p_gprop_schdefnrhs ENTER (gprop)")
        $$ = $1
    }
    ;

lgprop:
    optlabel gprop
    {
        xtracer.Trace("parser.p_lgprop ENTER (lgprop)")
        lf := acfg(v17lex).NewLabeledFormula($1, $2)
        lf.SetLineno(nodeLineno($2))
        $$ = lf
    }
    ;

// ============================================================
// --- Optional markers ---
// ============================================================

opttemporal:
    /* empty */
    {
        xtracer.Trace("parser.p_opttemporal ENTER (opttemporal)")
        $$ = nil
    }
    | TOK_TEMPORAL
    {
        xtracer.Trace("parser.p_opttemporal_symbol ENTER (opttemporal)")
        $$ = acfg(v17lex).NewAnd() // non-nil marker
    }
    ;

optunprovable:
    /* empty */
    {
        xtracer.Trace("parser.p_optunprovable ENTER (optunprovable)")
        $$ = nil
    }
    | TOK_UNPROVABLE
    {
        xtracer.Trace("parser.p_optunprovable_symbol ENTER (optunprovable)")
        $$ = acfg(v17lex).NewAnd() // non-nil marker
    }
    ;

optexplicit:
    /* empty */
    {
        xtracer.Trace("parser.p_optexplicit ENTER (optexplicit)")
        $$ = nil
    }
    | TOK_EXPLICIT
    {
        xtracer.Trace("parser.p_optexplicit_explicit ENTER (optexplicit)")
        $$ = acfg(v17lex).NewAnd() // non-nil marker
    }
    ;

optlabel:
    /* empty */
    {
        xtracer.Trace("parser.p_optlabel ENTER (optlabel)")
        $$ = nil
    }
    | labelname
    {
        xtracer.Trace("parser.p_optlabel_label ENTER (optlabel)")
        $$ = acfg(v17lex).NewAtom($1.Val)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    ;

optskolem:
    /* empty */
    {
        xtracer.Trace("parser.p_optskolem ENTER (optskolem)")
        $$ = nil
    }
    | TOK_NAMED defnlhs
    {
        xtracer.Trace("parser.p_optskolem_symbol ENTER (optskolem)")
        $$ = $2
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    ;

optinit:
    /* empty */
    {
        xtracer.Trace("parser.p_optinit ENTER (optinit)")
        $$ = nil
    }
    | TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_optinit_assign_fmla ENTER (optinit)")
        $$ = checkNonTemporal($2)
    }
    ;

optproof:
    /* empty */
    {
        xtracer.Trace("parser.p_optproof ENTER (optproof)")
        $$ = nil
    }
    | TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_optproof_symbol ENTER (optproof)")
        $$ = $2
    }
    | TOK_PROOF labelname proofstep
    {
        xtracer.Trace("parser.p_optproof_label_proofstep ENTER (optproof)")
        label := acfg(v17lex).NewAtom($2.Val)
        label.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        lf := acfg(v17lex).NewLabeledFormula(label, $3)
        lf.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = lf
    }
    ;

// optproofgroup2 removed — use optproofgroup instead

optsemi:
    /* empty */
    {
        xtracer.Trace("parser.p_optsemi ENTER (optsemi)")
        $$ = nil
    }
    | TOK_SEMI
    {
        xtracer.Trace("parser.p_optsemi_semi ENTER (optsemi)")
        $$ = nil
    }
    ;

// ============================================================
// --- Definitions ---
// ============================================================

dotsym:
    SYMBOLx
    {
        xtracer.Trace("parser.p_dotsym_symbol ENTER (dotsym) val=%s", $1.Val)
        $$ = $1.Val
        // Python: p[0].lineno = get_lineno(p,1) — emit trace to match
        _ = tokLineno(v17lex.(*v17LexAdapter), $1)
    }
    | dotsym TOK_DOT SYMBOLx
    {
        xtracer.Trace("parser.p_dotsym_dotsym_dot_symbol ENTER (dotsym)")
        $$ = $1 + "." + $3.Val
    }
    ;

defnlhs:
    dotsym
    {
        xtracer.Trace("parser.p_defnlhs_symbol ENTER (defnlhs)")
        $$ = acfg(v17lex).NewAtom($1)
    }
    | dotsym TOK_LPAREN defargs TOK_RPAREN
    {
        xtracer.Trace("parser.p_defnlhs_symbol_lparen_defargs_rparen ENTER (defnlhs)")
        a := acfg(v17lex).NewAtom($1)
        a.Terms = $3
        $$ = a
    }
    | TOK_LPAREN defarg relop defarg TOK_RPAREN
    {
        xtracer.Trace("parser.p_defnlhs_lp_term_relop_term_rp ENTER (defnlhs)")
        a := acfg(v17lex).NewAtom($3, $2, $4)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | TOK_LPAREN defarg infix defarg TOK_RPAREN
    {
        xtracer.Trace("parser.p_defnlhs_lp_term_infix_term_rp ENTER (defnlhs)")
        a := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($3, nil), $2, $4)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    ;

defarg:
    lparam
    {
        xtracer.Trace("parser.p_defarg_lparam ENTER (defarg)")
        $$ = $1
    }
    | var
    {
        xtracer.Trace("parser.p_defarg_var ENTER (defarg)")
        $$ = $1
    }
    ;

defargs:
    defarg
    {
        xtracer.Trace("parser.p_defargs_defarg ENTER (defargs)")
        $$ = []ast.Node{$1}
    }
    | defargs TOK_COMMA defarg
    {
        xtracer.Trace("parser.p_defargs_defargs_comma_defarg ENTER (defargs)")
        $$ = append($1, $3)
    }
    ;

typeddefn:
    defnlhs
    {
        xtracer.Trace("parser.p_typeddefn_defnlhs ENTER (typeddefn)")
        $$ = $1
    }
    | defnlhs TOK_COLON atype
    {
        xtracer.Trace("parser.p_typeddefn_defnlhs_colon_atype ENTER (typeddefn)")
        // set sort on the defnlhs
        if a, ok := $1.(*ast.Atom); ok {
            a.ASort = $3
        } else if app, ok := $1.(*ast.App); ok {
            app.ASort = $3
        }
        $$ = $1
    }
    ;

defnrhs:
    fmla
    {
        xtracer.Trace("parser.p_defnrhs_fmla ENTER (defnrhs)")
        $$ = checkNonTemporal($1)
    }
    | somevarfmla
    {
        xtracer.Trace("parser.p_defnrhs_somevarfmla ENTER (defnrhs)")
        $$ = checkNonTemporal($1)
    }
    | TOK_NATIVEQUOTE
    {
        xtracer.Trace("parser.p_defnrhs_nativequote ENTER (defnrhs)")
        text, bqs := parseNativequote(acfg(v17lex), $1.Val, v17lex.(*v17LexAdapter))
        elems := append([]ast.Node{acfg(v17lex).NewAtom(text)}, bqs...)
        ne := acfg(v17lex).NewNativeExpr(elems)
        ne.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = ne
    }
    ;

defn:
    typeddefn TOK_EQ defnrhs
    {
        xtracer.Trace("parser.p_defn_atom_fmla ENTER (defn)")
        d := acfg(v17lex).NewDefinition(ast.AppToAtom($1), $3)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = d
    }
    ;

defns:
    defn
    {
        xtracer.Trace("parser.p_defns_defn ENTER (defns)")
        $$ = []ast.Node{$1}
    }
    | defns TOK_COMMA defn
    {
        xtracer.Trace("parser.p_defns_defns_comma_defn ENTER (defns)")
        $$ = append($1, $3)
    }
    ;

gdefn:
    defn
    {
        xtracer.Trace("parser.p_gdefn_defn ENTER (gdefn)")
        $$ = $1
    }
    | TOK_LCB defn TOK_RCB
    {
        xtracer.Trace("parser.p_gdefn_lcb_defn_rcb ENTER (gdefn)")
        d := $2.(*ast.Definition)
        $$ = acfg(v17lex).NewDefinitionSchema(*d)
    }
    ;

somevarfmla:
    TOK_SOME simplevar TOK_DOT fmla optin optelse
    {
        xtracer.Trace("parser.p_somevarfmla_some_simplevar_dot_fmla ENTER (somevarfmla)")
        se := acfg(v17lex).NewSomeExpr($2, $4)
        if $5 != nil { se.IfValue = $5 }
        if $6 != nil { se.ElseVal = $6 }
        se.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = se
    }
    ;

optin:
    /* empty */
    {
        xtracer.Trace("parser.p_optin ENTER (optin)")
        $$ = nil
    }
    | TOK_IN fmla
    {
        xtracer.Trace("parser.p_optin_in_fmla ENTER (optin)")
        $$ = $2
    }
    ;

optelse:
    /* empty */
    {
        xtracer.Trace("parser.p_optelse ENTER (optelse)")
        $$ = nil
    }
    | TOK_ELSE fmla
    {
        xtracer.Trace("parser.p_optelse_else_fmla ENTER (optelse)")
        $$ = $2
    }
    ;

// ============================================================
// --- Schema ---
// ============================================================

schdefnrhs:
    fmla
    {
        xtracer.Trace("parser.p_schdefnrhs_fmla ENTER (schdefnrhs)")
        $$ = checkNonTemporal($1)
    }
    | TOK_LCB schdecls schconc TOK_RCB
    {
        xtracer.Trace("parser.p_schdefnrhs_lcb_schdecls_rcb ENTER (schdefnrhs)")
        args := append($2, $3)
        $$ = acfg(v17lex).NewSchemaBody(args...)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    ;

schdecl:
    TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_schdecl_funcdecl ENTER (schdecl)")
        $$ = make([]ast.Node, len($2))
        copy($$, $2)
    }
    | TOK_FRESH TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_schdecl_fresh_funcdecl ENTER (schdecl)")
        $$ = make([]ast.Node, len($3))
        copy($$, $3)
    }
    | TOK_INDIV funs
    {
        xtracer.Trace("parser.p_schdecl_indivdecl ENTER (schdecl)")
        $$ = make([]ast.Node, len($2))
        copy($$, $2)
    }
    | TOK_FRESH TOK_INDIV funs
    {
        xtracer.Trace("parser.p_schdecl_fresh_indivdecl ENTER (schdecl)")
        $$ = make([]ast.Node, len($3))
        copy($$, $3)
    }
    | TOK_RELATION rels
    {
        xtracer.Trace("parser.p_schdecl_relation_rel ENTER (schdecl)")
        $$ = make([]ast.Node, len($2))
        copy($$, $2)
    }
    | TOK_FRESH TOK_RELATION rels
    {
        xtracer.Trace("parser.p_schdecl_fresh_relation_rel ENTER (schdecl)")
        $$ = make([]ast.Node, len($3))
        copy($$, $3)
    }
    | TOK_TYPE SYMBOLx
    {
        xtracer.Trace("parser.p_schdecl_typedecl ENTER (schdecl)")
        scnst := acfg(v17lex).NewAtom($2.Val)
        scnst.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        tdfn := acfg(v17lex).NewTypeDef(scnst, acfg(v17lex).NewUninterpretedSortAST())
        tdfn.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = []ast.Node{tdfn}
    }
    | optexplicit TOK_PROPERTY lgprop
    {
        xtracer.Trace("parser.p_schdecl_propdecl ENTER (schdecl)")
        lf := addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "prop")
        if $1 != nil {
            lf.Explicit = true
        }
        $$ = []ast.Node{checkNonTemporal(lf).(*ast.LabeledFormula)}
    }
    | TOK_THEOREM lgprop
    {
        xtracer.Trace("parser.p_schdecl_theorem_lgprop ENTER (schdecl)")
        lf := addLabel(acfg(v17lex), $2.(*ast.LabeledFormula), "prop")
        $$ = []ast.Node{checkNonTemporal(lf).(*ast.LabeledFormula)}
    }
    | schdefnrhs
    {
        xtracer.Trace("parser.p_schdecl_theorem ENTER (schdecl)")
        lf := acfg(v17lex).NewLabeledFormula(nil, $1)
        lf = addLabel(acfg(v17lex), lf, "sch")
        $$ = []ast.Node{lf}
    }
    ;

schdecls:
    /* empty */
    {
        xtracer.Trace("parser.p_schdecls ENTER (schdecls)")
        $$ = nil
    }
    | schdecls schdecl
    {
        xtracer.Trace("parser.p_schdecls_schdecls_schdecl ENTER (schdecls)")
        $$ = append($1, $2...)
    }
    ;

schconc:
    TOK_DEFINITION defn
    {
        xtracer.Trace("parser.p_schconc_defdecl ENTER (schconc)")
        $$ = $2
    }
    | optexplicit TOK_PROPERTY lgprop
    {
        xtracer.Trace("parser.p_schconc_propdecl ENTER (schconc)")
        lf := $3.(*ast.LabeledFormula)
        // Python: p[0] = check_non_temporal(fmla)
        $$ = checkNonTemporal(lf.Formula)
    }
    ;

schdefn:
    defnlhs TOK_EQ schdefnrhs
    {
        xtracer.Trace("parser.p_schdefn_atom_eq_fmla ENTER (schdefn)")
        $$ = acfg(v17lex).NewDefinition(ast.AppToAtom($1), $3)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
    }
    ;

// ============================================================
// --- Symbol/Type/Relation/Function Declarations ---
// ============================================================

symdecl:
    constantdecl
    {
        xtracer.Trace("parser.p_symdecl_constantdecl ENTER (symdecl)")
        $$ = $1
    }
    | TOK_DESTRUCTOR tterms
    {
        xtracer.Trace("parser.p_symdecl_destructor_tterms ENTER (symdecl)")
        d := acfg(v17lex).NewDestructorDecl($2...)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = d
    }
    | TOK_FIELD tterms
    {
        xtracer.Trace("parser.p_symdecl_field_tterms ENTER (symdecl)")
        // Python: arg0 = Variable('SELF',This()); arg0.lineno = get_lineno(p,1)
        // Python: Variable('SELF', This()) — This() is special; use "this" as sort string
        arg0 := acfg(v17lex).NewVariable("SELF", "this")
        arg0.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        // Python: tterms = [x.clone([arg0]+x.args) for x in p[2]]
        // Python: for x,y in zip(p[2],tterms): y.lineno = x.lineno
        cloned := make([]ast.Node, len($2))
        for i, x := range $2 {
            newArgs := append([]ast.Node{arg0}, x.Args()...)
            cloned[i] = x.Clone(newArgs)
            cloned[i].SetLineno(x.GetLineno())
        }
        d := acfg(v17lex).NewDestructorDecl(cloned...)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = d
    }
    | TOK_CONSTRUCTOR tterms
    {
        xtracer.Trace("parser.p_symdecl_constructor_tterms ENTER (symdecl)")
        // Python: for t in p[2]: if not hasattr(t,'sort'): t.sort = This()
        lex := v17lex.(*v17LexAdapter)
        for _, t := range $2 {
            if app, ok := t.(*ast.App); ok && app.ASort == nil {
                thisNode := acfg(v17lex).NewThis()
                thisNode.SetLineno(tokLineno(lex, $1))
                app.ASort = thisNode
            }
        }
        d := acfg(v17lex).NewConstructorDecl($2...)
        d.SetLineno(tokLineno(lex, $1))
        $$ = d
    }
    ;

constantdecl:
    TOK_INDIV tterms
    {
        xtracer.Trace("parser.p_constantdecl_constant_tterms ENTER (constantdecl)")
        d := acfg(v17lex).NewConstantDecl($2...)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = d
    }
    | TOK_VAR tterms
    {
        xtracer.Trace("parser.p_constantdecl_var_tterms ENTER (constantdecl)")
        d := acfg(v17lex).NewConstantDecl($2...)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = d
    }
    | TOK_PARAMETER parameter
    {
        xtracer.Trace("parser.p_constantdecl_parameter_tterm ENTER (constantdecl)")
        $$ = $2
    }
    ;

parameter:
    tterm
    {
        xtracer.Trace("parser.p_param_tterm ENTER (parameter)")
        d := acfg(v17lex).NewParameterDecl($1)
        $$ = d
    }
    | tterm TOK_EQ paramval
    {
        xtracer.Trace("parser.p_param_tterm_eq_paramval ENTER (parameter)")
        df := acfg(v17lex).NewDefinition($1, $3)
        df.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        d := acfg(v17lex).NewParameterDecl(df)
        $$ = d
    }
    ;

paramval:
    TOK_TRUE
    {
        xtracer.Trace("parser.p_paramval_true ENTER (paramval)")
        $$ = acfg(v17lex).NewAtom("true")
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | TOK_FALSE
    {
        xtracer.Trace("parser.p_paramval_false ENTER (paramval)")
        $$ = acfg(v17lex).NewAtom("false")
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | SYMBOLx
    {
        xtracer.Trace("parser.p_paramval_symbol ENTER (paramval)")
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil))
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    ;

tapp:
    SYMBOLx
    {
        xtracer.Trace("parser.p_tapp_symbol ENTER (tapp)")
        a := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil))
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    | SYMBOLx targs
    {
        xtracer.Trace("parser.p_tapp_symbol_targs ENTER (tapp)")
        args := make([]ast.Node, len($2))
        copy(args, $2)
        a := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil), args...)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    | TOK_LPAREN var infix var TOK_RPAREN
    {
        xtracer.Trace("parser.p_tapp_lp_symbol_infix_symbol_rp ENTER (tapp)")
        a := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($3, nil), $2, $4)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    ;

tterm:
    tapp
    {
        xtracer.Trace("parser.p_tterm_term ENTER (tterm)")
        $$ = $1
    }
    | tapp TOK_COLON atype
    {
        xtracer.Trace("parser.p_tterm_term_colon_symbol ENTER (tterm)")
        if app, ok := $1.(*ast.App); ok {
            app.ASort = $3
        }
        $$ = $1
    }
    ;

tterms:
    tterm
    {
        xtracer.Trace("parser.p_tterms_tterm ENTER (tterms)")
        $$ = []ast.Node{$1}
    }
    | tterms TOK_COMMA tterm
    {
        xtracer.Trace("parser.p_tterms_tterms_comma_tterm ENTER (tterms)")
        $$ = append($1, $3)
    }
    ;

targs:
    TOK_LPAREN TOK_RPAREN
    {
        xtracer.Trace("parser.p_targs_lparen_rparen ENTER (targs)")
        $$ = nil
    }
    | TOK_LPAREN tsyms TOK_RPAREN
    {
        xtracer.Trace("parser.p_targs_lparen_tsyms_rparen ENTER (targs)")
        $$ = $2
    }
    ;

tsyms:
    var
    {
        xtracer.Trace("parser.p_tsyms_tsym ENTER (tsyms)")
        $$ = []ast.Node{$1}
    }
    | tsyms TOK_COMMA var
    {
        xtracer.Trace("parser.p_tsyms_tsyms_comma_tsym ENTER (tsyms)")
        $$ = append($1, $3)
    }
    ;

tatom:
    SYMBOLx
    {
        xtracer.Trace("parser.p_tatom_symbol ENTER (tatom)")
        $$ = acfg(v17lex).NewAtom($1.Val)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | SYMBOLx targs
    {
        xtracer.Trace("parser.p_tatom_symbol_targs ENTER (tatom)")
        $$ = acfg(v17lex).NewAtom($1.Val, $2...)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | TOK_LPAREN var relop var TOK_RPAREN
    {
        xtracer.Trace("parser.p_tatom_lp_symbol_relop_symbol_rp ENTER (tatom)")
        $$ = acfg(v17lex).NewAtom($3, $2, $4)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    ;

tatoms:
    tatom
    {
        xtracer.Trace("parser.p_tatoms_tatom ENTER (tatoms)")
        $$ = []ast.Node{$1}
    }
    | tatoms TOK_COMMA tatom
    {
        xtracer.Trace("parser.p_tatoms_tatoms_comma_tatom ENTER (tatoms)")
        $$ = append($1, $3)
    }
    ;

rel:
    defnlhs
    {
        xtracer.Trace("parser.p_rel_defnlhs ENTER (rel)")
        // Python: p[1].sort = 'bool' — relation declarations have bool sort
        if a, ok := $1.(*ast.Atom); ok {
            a.ASort = acfg(v17lex).NewSymbol("bool", nil)
        } else if app, ok := $1.(*ast.App); ok {
            app.ASort = acfg(v17lex).NewSymbol("bool", nil)
        }
        d := acfg(v17lex).NewConstantDecl($1)
        $$ = d
    }
    | defn
    {
        xtracer.Trace("parser.p_rel_defn ENTER (rel)")
        lf := addLabel(acfg(v17lex), mkLF(acfg(v17lex), $1), "def")
        d := acfg(v17lex).NewDerivedDecl(lf)
        $$ = d
    }
    ;

rels:
    rel
    {
        xtracer.Trace("parser.p_rels_rel ENTER (rels)")
        $$ = []ast.Node{$1}
    }
    | rels TOK_COMMA rel
    {
        xtracer.Trace("parser.p_rels_rels_comma_rel ENTER (rels)")
        $$ = append($1, $3)
    }
    ;

fun:
    typeddefn
    {
        xtracer.Trace("parser.p_fun_defnlhs_colon_atype ENTER (fun)")
        d := acfg(v17lex).NewConstantDecl($1)
        $$ = d
    }
    | typeddefn TOK_EQ defnrhs
    {
        xtracer.Trace("parser.p_fun_defn ENTER (fun)")
        df := acfg(v17lex).NewDefinition(ast.AppToAtom($1), $3)
        df.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        lf := addLabel(acfg(v17lex), mkLF(acfg(v17lex), df), "def")
        d := acfg(v17lex).NewDerivedDecl(lf)
        $$ = d
    }
    ;

funs:
    fun
    {
        xtracer.Trace("parser.p_funs_fun ENTER (funs)")
        $$ = []ast.Node{$1}
    }
    | funs TOK_COMMA fun
    {
        xtracer.Trace("parser.p_funs_funs_comma_fun ENTER (funs)")
        $$ = append($1, $3)
    }
    ;

// --- Type-related ---

typesymbol:
    SYMBOLx
    {
        xtracer.Trace("parser.p_typesymbol_symbol ENTER (typesymbol)")
        $$ = acfg(v17lex).NewAtom($1.Val)
    }
    | TOK_THIS
    {
        xtracer.Trace("parser.p_typesymbol_this ENTER (typesymbol)")
        a := acfg(v17lex).NewAtom("this")
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    ;

optfinite:
    /* empty */
    {
        xtracer.Trace("parser.p_optfinite ENTER (optfinite)")
        $$ = false
    }
    | TOK_FINITE
    {
        xtracer.Trace("parser.p_optfinite_finite ENTER (optfinite)")
        $$ = true
    }
    ;

optghost:
    /* empty */
    {
        xtracer.Trace("parser.p_optghost ENTER (optghost)")
        $$ = false
    }
    | TOK_GHOST
    {
        xtracer.Trace("parser.p_optghost_ghost ENTER (optghost)")
        $$ = true
    }
    ;

sort:
    TOK_LCB SYMBOLx TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_symbol_rcb ENTER (sort)")
        $$ = acfg(v17lex).NewEnumeratedSort(acfg(v17lex).NewAtom($2.Val))
    }
    | TOK_LCB SYMBOLx TOK_COMMA names TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_names_rcb ENTER (sort)")
        vals := []ast.Node{acfg(v17lex).NewAtom($2.Val)}
        for _, n := range $4 {
            vals = append(vals, n)
        }
        $$ = acfg(v17lex).NewEnumeratedSort(vals...)
    }
    | TOK_LCB SYMBOLx TOK_DOTS SYMBOLx TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_symbol_dots_symbol_rcb ENTER (sort)")
        $$ = acfg(v17lex).NewRange(acfg(v17lex).NewAtom($2.Val), acfg(v17lex).NewAtom($4.Val))
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
    }
    | TOK_STRUCT TOK_LCB tterms TOK_RCB
    {
        xtracer.Trace("parser.p_sort_struct_lcb_names_rcb ENTER (sort)")
        $$ = acfg(v17lex).NewStructSort($3...)
    }
    | TOK_STRUCT TOK_LCB TOK_RCB
    {
        xtracer.Trace("parser.p_sort_struct_lcb_rcb ENTER (sort)")
        $$ = acfg(v17lex).NewStructSort()
    }
    ;

names:
    SYMBOLx
    {
        xtracer.Trace("parser.p_names_symbol ENTER (names)")
        $$ = []ast.Node{acfg(v17lex).NewAtom($1.Val)}
    }
    | names TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_names_names_comma_symbol ENTER (names)")
        $$ = append($1, acfg(v17lex).NewAtom($3.Val))
    }
    ;

// ============================================================
// --- relop / infix ---
// ============================================================

relop:
    TOK_EQ     { xtracer.Trace("parser.p_relop_eq ENTER (relop)"); $$ = "=" }
    | TOK_LE   { xtracer.Trace("parser.p_relop_le ENTER (relop)"); $$ = "<=" }
    | TOK_LT   { xtracer.Trace("parser.p_relop_lt ENTER (relop)"); $$ = "<" }
    | TOK_GE   { xtracer.Trace("parser.p_relop_ge ENTER (relop)"); $$ = ">=" }
    | TOK_GT   { xtracer.Trace("parser.p_relop_gt ENTER (relop)"); $$ = ">" }
    | TOK_PTO  { xtracer.Trace("parser.p_relop_pto ENTER (relop)"); $$ = "*>" }
    ;

infix:
    TOK_PLUS   { xtracer.Trace("parser.p_infix_plus ENTER (infix)"); $$ = "+" }
    | TOK_MINUS { xtracer.Trace("parser.p_infix_minus ENTER (infix)"); $$ = "-" }
    | TOK_TIMES { xtracer.Trace("parser.p_infix_times ENTER (infix)"); $$ = "*" }
    | TOK_DIV   { xtracer.Trace("parser.p_infix_div ENTER (infix)"); $$ = "/" }
    ;

// ============================================================
// --- atom / atoms / app / apps / lit ---
// ============================================================

atom:
    SYMBOLx
    {
        xtracer.Trace("parser.p_atom_symbol ENTER (atom)")
        a := acfg(v17lex).NewAtom($1.Val)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        xtracer.Trace("parser.p_atom_symbol_lp_terms_rp ENTER (atom)")
        a := acfg(v17lex).NewAtom($1.Val, $3...)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    ;

atoms:
    atom
    {
        xtracer.Trace("parser.p_atoms_atom ENTER (atoms)")
        $$ = []ast.Node{$1}
    }
    | atoms TOK_COMMA atom
    {
        xtracer.Trace("parser.p_atoms_atoms_atom ENTER (atoms)")
        $$ = append($1, $3)
    }
    ;

app:
    SYMBOLx
    {
        xtracer.Trace("parser.p_app_symbol ENTER (app)")
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil))
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        xtracer.Trace("parser.p_app_symbol_lp_terms_rp ENTER (app)")
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil), $3...)
    }
    | term infix term
    {
        xtracer.Trace("parser.p_app_term_infix_term ENTER (app)")
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($2, nil), $1, $3)
    }
    ;

apps:
    app
    {
        xtracer.Trace("parser.p_apps_app ENTER (apps)")
        $$ = []ast.Node{$1}
    }
    | apps TOK_COMMA app
    {
        xtracer.Trace("parser.p_apps_apps_app ENTER (apps)")
        $$ = append($1, $3)
    }
    ;

lit:
    atom
    {
        xtracer.Trace("parser.p_lit_atom ENTER (lit)")
        // Python: p[0] = Literal(1, p[1])
        $$ = acfg(v17lex).NewLiteral(1, $1)
        $$.SetLineno(nodeLineno($1))
    }
    | SYMBOLx TOK_EQ SYMBOLx
    {
        xtracer.Trace("parser.p_lit_term_eq_term ENTER (lit)")
        // Python: p[0] = Literal(1, Atom(p[2], [symbol(p[1]), symbol(p[3])]))
        a := acfg(v17lex).NewAtom("=", acfg(v17lex).NewAtom($1.Val), acfg(v17lex).NewAtom($3.Val))
        $$ = acfg(v17lex).NewLiteral(1, a)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
    }
    | SYMBOLx TOK_TILDAEQ SYMBOLx
    {
        xtracer.Trace("parser.p_lit_term_tildaeq_term ENTER (lit)")
        // Python: p[0] = Literal(0, Atom(p[2], [symbol(p[1]), symbol(p[3])]))
        a := acfg(v17lex).NewAtom("=", acfg(v17lex).NewAtom($1.Val), acfg(v17lex).NewAtom($3.Val))
        $$ = acfg(v17lex).NewLiteral(0, a)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
    }
    | TOK_TILDA lit
    {
        xtracer.Trace("parser.p_lit_tilda_atom ENTER (lit)")
        // Python: p[0] = ~p[2] — flips Literal polarity
        if lit, ok := $2.(*ast.Literal); ok {
            $$ = acfg(v17lex).NewLiteral(1 - lit.Polarity, lit.Atom)
        } else {
            $$ = acfg(v17lex).NewLiteral(0, $2)
        }
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    ;

// ============================================================
// --- callatom / callatoms ---
// ============================================================

callatom:
    atom
    {
        xtracer.Trace("parser.p_callatom_atom ENTER (callatom)")
        $$ = $1
    }
    | TOK_THIS
    {
        xtracer.Trace("parser.p_callatom_this ENTER (callatom)")
        a := acfg(v17lex).NewAtom("this")
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    | TOK_METHOD
    {
        xtracer.Trace("parser.p_callatom_method ENTER (callatom)")
        $$ = acfg(v17lex).NewAtom("method")
    }
    | callatom TOK_DOT callatom
    {
        xtracer.Trace("parser.p_callatom_callatom_dot_callatom ENTER (callatom)")
        lhs := $1.(*ast.Atom)
        rhs := $3.(*ast.Atom)
        $$ = ast.ComposeAtoms(lhs, rhs)
        $$.SetLineno(nodeLineno($1))
    }
    ;

callatoms:
    callatom
    {
        xtracer.Trace("parser.p_callatoms_callatom ENTER (callatoms)")
        $$ = []ast.Node{$1}
    }
    | callatoms TOK_COMMA callatom
    {
        xtracer.Trace("parser.p_callatoms_callatoms_callatom ENTER (callatoms)")
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Module/Object helpers ---
// ============================================================

modulestart:
    /* empty */
    {
        xtracer.Trace("parser.p_modulestart ENTER (modulestart)")
        // Python: stack[-1].is_module = True (ivy_parser.py:601)
        v17lex.(*v17LexAdapter).accum.isModule = true
        $$ = nil
    }
    ;

moduleend:
    /* empty */
    {
        xtracer.Trace("parser.p_moduleend ENTER (moduleend)")
        $$ = nil
    }
    ;

objectend:
    /* empty */
    {
        xtracer.Trace("parser.p_objectend ENTER (objectend)")
        // Python: stack[-1].is_object = False
        v17lex.(*v17LexAdapter).accum.isObject = false
        $$ = nil
    }
    ;

modcat:
    /* empty */
    {
        xtracer.Trace("parser.p_modcat ENTER (modcat)")
        $$ = nil
    }
    | TOK_OBJECT
    {
        xtracer.Trace("parser.p_modcat_object ENTER (modcat)")
        $$ = acfg(v17lex).NewAtom("object")
    }
    | TOK_ISOLATE
    {
        xtracer.Trace("parser.p_modcat_isolate ENTER (modcat)")
        $$ = acfg(v17lex).NewAtom("isolate")
    }
    ;

opteq:
    /* empty */
    {
        xtracer.Trace("parser.p_opteq ENTER (opteq)")
        $$ = nil
    }
    | TOK_EQ
    {
        xtracer.Trace("parser.p_opteq_eq ENTER (opteq)")
        $$ = nil
    }
    ;

optdotdotdot:
    /* empty */
    {
        xtracer.Trace("parser.p_optdotdotdot ENTER (optdotdotdot)")
        // Python: parent_object = None — not a continuation, clear parent
        v17lex.(*v17LexAdapter).parentObjName = ""
        $$ = false
    }
    | TOK_DOTDOTDOT
    {
        xtracer.Trace("parser.p_optdotdotdot_dotdotdot ENTER (optdotdotdot)")
        // Python: parent_object stays set — IS a continuation
        $$ = true
    }
    ;

objectargs:
    optargs
    {
        xtracer.Trace("parser.p_objectargs_optargs ENTER (objectargs)")
        $$ = $1
        // Python: stack[-1].params = p[0]
        v17lex.(*v17LexAdapter).accum.params = $1
    }
    ;

objsym:
    SYMBOLx
    {
        xtracer.Trace("parser.p_objsym ENTER (objsym)")
        lex := v17lex.(*v17LexAdapter)
        $$ = acfg(v17lex).NewAtom($1.Val)
        // Python: global parent_object; parent_object = p[0]
        lex.parentObjName = $1.Val
    }
    ;

opttrusted:
    /* empty */
    {
        xtracer.Trace("parser.p_opttrusted ENTER (opttrusted)")
        $$ = false
    }
    | TOK_TRUSTED
    {
        xtracer.Trace("parser.p_opttrusted_trusted ENTER (opttrusted)")
        $$ = true
    }
    ;

optargs:
    /* empty */
    {
        xtracer.Trace("parser.p_optargs ENTER (optargs)")
        $$ = nil
    }
    | TOK_LPAREN lparams TOK_RPAREN
    {
        xtracer.Trace("parser.p_optargs_params ENTER (optargs)")
        $$ = $2
    }
    ;

optreturns:
    /* empty */
    {
        xtracer.Trace("parser.p_optreturns ENTER (optreturns)")
        $$ = nil
    }
    | TOK_RETURNS TOK_LPAREN lparams TOK_RPAREN
    {
        xtracer.Trace("parser.p_optreturns_tsyms ENTER (optreturns)")
        $$ = $3
    }
    ;

optactualreturns:
    /* empty */
    {
        xtracer.Trace("parser.p_optactualreturns ENTER (optactualreturns)")
        $$ = nil
    }
    | callatoms TOK_ASSIGN
    {
        xtracer.Trace("parser.p_optactualreturns_callatoms_assign ENTER (optactualreturns)")
        $$ = $1
    }
    ;

param:
    SYMBOLx TOK_COLON SYMBOLx
    {
        xtracer.Trace("parser.p_param_term_colon_symbol ENTER (param)")
        a := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil))
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        a.ASort = acfg(v17lex).NewSymbol($3.Val, nil)
        $$ = a
    }
    ;

params:
    param
    {
        xtracer.Trace("parser.p_params_param ENTER (params)")
        $$ = []ast.Node{$1}
    }
    | params TOK_COMMA param
    {
        xtracer.Trace("parser.p_params_params_comma_param ENTER (params)")
        $$ = append($1, $3)
    }
    ;

optwith:
    /* empty */
    {
        xtracer.Trace("parser.p_optwith ENTER (optwith)")
        $$ = nil
    }
    | TOK_WITH callatoms
    {
        xtracer.Trace("parser.p_optwith_with_callatoms ENTER (optwith)")
        $$ = $2
    }
    ;

// ============================================================
// --- lparam / lparams ---
// ============================================================

lparam:
    SYMBOLx TOK_COLON atype
    {
        xtracer.Trace("parser.p_lparam_variable_colon_symbol ENTER (lparam)")
        // Python: p[0] = App(p[1]); p[0].sort = p[3]
        a := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil))
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        a.ASort = $3
        $$ = a
    }
    | TOK_CARET SYMBOLx TOK_COLON atype
    {
        xtracer.Trace("parser.p_lparam_caret_variable_colon_symbol ENTER (lparam)")
        // Python: p[0] = KeyArg(p[2]); p[0].sort = p[4]
        a := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($2.Val, nil))
        a.ASort = $4
        $$ = a
    }
    ;

lparams:
    lparam
    {
        xtracer.Trace("parser.p_lparams_lparam ENTER (lparams)")
        $$ = []ast.Node{$1}
    }
    | lparams TOK_COMMA lparam
    {
        xtracer.Trace("parser.p_lparams_lparams_comma_lparam ENTER (lparams)")
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Action top-level helpers ---
// ============================================================

optactiondef:
    /* empty */
    {
        xtracer.Trace("parser.p_optactiondef ENTER (optactiondef)")
        $$ = acfg(v17lex).NewSequence()
    }
    | TOK_EQ topseq
    {
        xtracer.Trace("parser.p_optactiondef_eq_topseq ENTER (optactiondef)")
        $$ = $2
    }
    | TOK_EQ TOK_TIMES
    {
        xtracer.Trace("parser.p_optactiondef_eq_symbol ENTER (optactiondef)")
        $$ = acfg(v17lex).NewCrashAction()
    }
    ;

topseq:
    sequence
    {
        xtracer.Trace("parser.p_topseq_sequence ENTER (topseq)")
        $$ = $1
    }
    | TOK_LCB TOK_NATIVEQUOTE TOK_RCB
    {
        xtracer.Trace("parser.p_topseq_lcb_nativequote_rcb ENTER (topseq)")
        // Python: NativeAction(*([text] + bqs))
        text, bqs := parseNativequote(acfg(v17lex), $2.Val, v17lex.(*v17LexAdapter))
        args := append([]ast.Node{acfg(v17lex).NewNativeCode(text)}, bqs...)
        na := acfg(v17lex).NewNativeAction(args...)
        na.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = na
    }
    ;

optimpex:
    /* empty */
    {
        xtracer.Trace("parser.p_optimpex ENTER (optimpex)")
        $$ = nil
    }
    | TOK_EXPORT
    {
        xtracer.Trace("parser.p_optimpex_export ENTER (optimpex)")
        $$ = nil // marker handled in top rule
    }
    | TOK_IMPORT
    {
        xtracer.Trace("parser.p_optimpex_import ENTER (optimpex)")
        $$ = nil // marker handled in top rule
    }
    ;

actmeth:
    TOK_ACTION
    {
        xtracer.Trace("parser.p_actmeth_action ENTER (actmeth)")
        $$ = false
    }
    | TOK_METHOD
    {
        xtracer.Trace("parser.p_actmeth_method ENTER (actmeth)")
        $$ = true
    }
    ;

// ============================================================
// --- specimpl ---
// ============================================================

specimpl:
    TOK_SPECIFICATION
    {
        xtracer.Trace("parser.p_specimpl_specification ENTER (specimpl)")
        $$ = "spec"
        // Python: global special_attribute; special_attribute = "spec"
        v17lex.(*v17LexAdapter).specialAttribute = "spec"
    }
    | TOK_IMPLEMENTATION
    {
        xtracer.Trace("parser.p_specimpl_implementation ENTER (specimpl)")
        $$ = "impl"
        // Python: global special_attribute; special_attribute = "impl"
        v17lex.(*v17LexAdapter).specialAttribute = "impl"
    }
    | TOK_PRIVATE
    {
        xtracer.Trace("parser.p_specimpl_private ENTER (specimpl)")
        $$ = "private"
        // Python: global special_attribute; special_attribute = "private"
        v17lex.(*v17LexAdapter).specialAttribute = "private"
    }
    | TOK_GLOBAL
    {
        xtracer.Trace("parser.p_specimpl_global ENTER (specimpl)")
        $$ = "global"
        // Python: global global_attribute; global_attribute = "global"
        v17lex.(*v17LexAdapter).globalAttribute = "global"
    }
    | TOK_COMMON
    {
        xtracer.Trace("parser.p_specimpl_common ENTER (specimpl)")
        $$ = "common"
        // Python: global common_attribute; common_attribute = "common"
        v17lex.(*v17LexAdapter).commonAttribute = "common"
    }
    ;

// ============================================================
// --- Instantiate ---
// ============================================================

insts:
    inst
    {
        xtracer.Trace("parser.p_insts_inst ENTER (insts)")
        $$ = []ast.Node{$1}
    }
    | insts TOK_COMMA inst
    {
        xtracer.Trace("parser.p_insts_insts_comma_inst ENTER (insts)")
        $$ = append($1, $3)
    }
    ;

inst:
    modinst
    {
        xtracer.Trace("parser.p_inst_modinst ENTER (inst)")
        n := acfg(v17lex).NewInstantiation(nil, ast.AppToAtom($1))
        n.SetLineno(nodeLineno($1))
        $$ = n
    }
    | modinst TOK_COLON modinst
    {
        xtracer.Trace("parser.p_inst_atom_colon_modinst ENTER (inst)")
        inst := acfg(v17lex).NewInstantiation(ast.AppToAtom($1), ast.AppToAtom($3))
        inst.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = inst
    }
    ;

modinst:
    dotsym
    {
        xtracer.Trace("parser.p_modinst_symbol ENTER (modinst)")
        $$ = acfg(v17lex).NewAtom($1)
    }
    | dotsym TOK_LPAREN pnames TOK_RPAREN
    {
        xtracer.Trace("parser.p_modinst_symbol_lp_pnames_rp ENTER (modinst)")
        a := acfg(v17lex).NewAtom($1)
        a.Terms = $3
        $$ = a
    }
    ;

pname:
    atype
    {
        xtracer.Trace("parser.p_pname_symbol ENTER (pname)")
        var rep string
        switch v := $1.(type) {
        case *ast.Symbol:
            rep = v.Rep
        case *ast.This:
            rep = "this"
        default:
            rep = fmt.Sprint($1)
        }
        n := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol(rep, nil))
        n.SetLineno(nodeLineno($1))
        $$ = n
    }
    | var
    {
        xtracer.Trace("parser.p_pname_var ENTER (pname)")
        $$ = $1
    }
    | infix
    {
        xtracer.Trace("parser.p_pname_infix ENTER (pname)")
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1, nil))
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | relop
    {
        xtracer.Trace("parser.p_pname_relop ENTER (pname)")
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1, nil))
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_THIS
    {
        xtracer.Trace("parser.p_pname_this ENTER (pname)")
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("this", nil))
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | TOK_TRUE
    {
        xtracer.Trace("parser.p_pname_true ENTER (pname)")
        $$ = acfg(v17lex).NewAtom("true")
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | TOK_FALSE
    {
        xtracer.Trace("parser.p_pname_false ENTER (pname)")
        $$ = acfg(v17lex).NewAtom("false")
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    ;

pnames:
    /* empty */
    {
        xtracer.Trace("parser.p_pnames ENTER (pnames)")
        $$ = nil
    }
    | pname
    {
        xtracer.Trace("parser.p_pnames_pname ENTER (pnames)")
        $$ = []ast.Node{$1}
    }
    | pnames TOK_COMMA pname
    {
        xtracer.Trace("parser.p_pnames_pnames_pname ENTER (pnames)")
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Interpret/Alias/Attribute helpers ---
// ============================================================

oper:
    atype
    {
        xtracer.Trace("parser.p_oper_symbol ENTER (oper)")
        // Python: p[0] = Atom(p[1]) — atype returns a string, oper wraps in Atom
        if sym, ok := $1.(*ast.Symbol); ok {
            $$ = acfg(v17lex).NewAtom(sym.Rep)
        } else if th, ok := $1.(*ast.This); ok {
            a := acfg(v17lex).NewAtom("this")
            a.SetLineno(th.GetLineno())
            $$ = a
        } else {
            $$ = $1
        }
    }
    | relop
    {
        xtracer.Trace("parser.p_oper_relop ENTER (oper)")
        $$ = acfg(v17lex).NewAtom($1)
    }
    | infix
    {
        xtracer.Trace("parser.p_oper_infix ENTER (oper)")
        $$ = acfg(v17lex).NewAtom($1)
    }
    | TOK_NATIVEQUOTE
    {
        xtracer.Trace("parser.p_oper_nativequote ENTER (oper)")
        text, bqs := parseNativequote(acfg(v17lex), $1.Val, v17lex.(*v17LexAdapter))
        elems := append([]ast.Node{acfg(v17lex).NewAtom(text)}, bqs...)
        nt := acfg(v17lex).NewNativeType(elems...)
        nt.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = nt
    }
    ;

attributeval:
    callatom
    {
        xtracer.Trace("parser.p_top_attributeval_callatom ENTER (attributeval)")
        $$ = $1
    }
    | TOK_TRUE
    {
        xtracer.Trace("parser.p_top_attributeval_true ENTER (attributeval)")
        $$ = acfg(v17lex).NewAtom("true")
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | TOK_FALSE
    {
        xtracer.Trace("parser.p_top_attributeval_false ENTER (attributeval)")
        $$ = acfg(v17lex).NewAtom("false")
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    ;

moresymbols:
    /* empty */
    {
        xtracer.Trace("parser.p_moresymbols ENTER (moresymbols)")
        $$ = nil
    }
    | moresymbols TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_moresymbols_more_symbols_comma_symbol ENTER (moresymbols)")
        $$ = append($1, acfg(v17lex).NewAtom($3.Val))
    }
    ;

// ============================================================
// --- Delegate ---
// ============================================================

optdelegee:
    /* empty */
    {
        xtracer.Trace("parser.p_optdelegee ENTER (optdelegee)")
        $$ = nil
    }
    | TOK_ARROW callatom
    {
        xtracer.Trace("parser.p_optdelegee_callatom ENTER (optdelegee)")
        $$ = $2
    }
    ;

// ============================================================
// --- sequence / actseq / action ---
// ============================================================

sequence:
    TOK_LCB TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_rcb ENTER (sequence)")
        $$ = acfg(v17lex).NewSequence()
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | TOK_LCB actseq TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_actseq_rcb ENTER (sequence)")
        stmts := lowerVarStmts($2)
        seq := lalrMakeSequence(acfg(v17lex), stmts)
        if s, ok := seq.(*ast.Sequence); ok {
            s.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        } else {
            // Single node — mark as having a location without calling getLineno
            // (Python always has lineno set on these nodes from their own rules)
            if b, ok := seq.(interface{ HasLocSet() bool }); ok && !b.HasLocSet() {
                loc := v17lex.(*v17LexAdapter).lastTok.Line
                seq.SetLineno(ast.Location{Line: loc})
            }
        }
        $$ = seq
    }
    | TOK_LCB actseq TOK_SEMI TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_actseq_semi_rcb ENTER (sequence)")
        // Python: p[0] = Sequence(*lower_var_stmts(p[2]))
        // Unlike p_sequence_lcb_actseq_rcb, this rule always wraps in Sequence
        // and only calls lower_var_stmts once (no len==1 shortcut).
        stmts := lowerVarStmts($2)
        seq := acfg(v17lex).NewSequence(stmts...)
        seq.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = seq
    }
    ;

// actseq and actseqrev match Python exactly (ivy_parser.py:2626-2668).
// Python uses right-recursive actseqrev, then reverses to get actseq.
// We replicate this structure faithfully to match reduction ordering.

actseq:
    actseqrev
    {
        xtracer.Trace("parser.p_actseq_actseqrev ENTER (actseq)")
        // Python: p[1].reverse()
        for i, j := 0, len($1)-1; i < j; i, j = i+1, j-1 {
            $1[i], $1[j] = $1[j], $1[i]
        }
        $$ = $1
    }
    ;

actseqrev:
    simpleact
    {
        xtracer.Trace("parser.p_actseqrev_simpleact ENTER (actseqrev)")
        $$ = []ast.Node{$1}
    }
    | complexact
    {
        xtracer.Trace("parser.p_actseqrev_complexact ENTER (actseqrev)")
        $$ = []ast.Node{$1}
    }
    | simpleact TOK_SEMI actseqrev
    {
        xtracer.Trace("parser.p_actseqrev_simpact_semi_actseqrev ENTER (actseqrev)")
        $$ = append($3, $1)
    }
    | simpleact TOK_SEMI
    {
        xtracer.Trace("parser.p_actseqrev_simpact_semi ENTER (actseqrev)")
        $$ = []ast.Node{$1}
    }
    | complexact actseqrev
    {
        xtracer.Trace("parser.p_actseqrev_complexact_actseqrev ENTER (actseqrev)")
        $$ = append($2, $1)
    }
    | complexact TOK_SEMI actseqrev
    {
        xtracer.Trace("parser.p_actseqrev_complexact_semi_actseqrev ENTER (actseqrev)")
        $$ = append($3, $1)
    }
    | complexact TOK_SEMI
    {
        xtracer.Trace("parser.p_actseqrev_complexact_semi ENTER (actseqrev)")
        $$ = []ast.Node{$1}
    }
    ;

action:
    simpleact
    {
        xtracer.Trace("parser.p_action_simpleact ENTER (action)")
        $$ = $1
    }
    | complexact
    {
        xtracer.Trace("parser.p_action_complexact ENTER (action)")
        $$ = $1
    }
    ;

// ============================================================
// --- Simple actions ---
// ============================================================

simpleact:
    TOK_ASSUME labeledfmla
    {
        xtracer.Trace("parser.p_action_assume ENTER (simpleact)")
        // Python: AssumeAction(check_non_temporal(addlabel(p[2],'asrt')))
        lf := addLabel(acfg(v17lex), $2.(*ast.LabeledFormula), "asrt")
        a := acfg(v17lex).NewAssumeAction(checkNonTemporal(lf))
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    | optunprovable TOK_ASSERT labeledfmla
    {
        xtracer.Trace("parser.p_action_assert ENTER (simpleact)")
        lf := addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := acfg(v17lex).NewAssertAction(lf)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        // Python: if p[1] and not check_unprovable.get(): p[0] = Sequence()
        if $1 != nil && !acfg(v17lex).CheckUnprovable {
            $$ = acfg(v17lex).NewSequence()
            $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        } else {
            $$ = a
        }
    }
    | optunprovable TOK_ASSERT labeledfmla TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_action_assert_proof_proofstep ENTER (simpleact)")
        // Python: AssertAction(check_non_temporal(addlabel(p[3],'asrt')),p[5])
        lf := addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := acfg(v17lex).NewAssertAction(lf, $5)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        // Python: if p[1] and not check_unprovable.get(): p[0] = Sequence()
        if $1 != nil && !acfg(v17lex).CheckUnprovable {
            $$ = acfg(v17lex).NewSequence()
            $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        } else {
            $$ = a
        }
    }
    | optunprovable TOK_REQUIRE labeledfmla
    {
        xtracer.Trace("parser.p_action_require ENTER (simpleact)")
        // Python: RequiresAction(check_non_temporal(addlabel(p[3],'asrt')))
        lf := addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := acfg(v17lex).NewRequiresAction(lf)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        // Python: if p[1] and not check_unprovable.get(): p[0] = Sequence()
        if $1 != nil && !acfg(v17lex).CheckUnprovable {
            $$ = acfg(v17lex).NewSequence()
            $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        } else {
            $$ = a
        }
    }
    | optunprovable TOK_REQUIRE labeledfmla TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_action_require_proof_proofstep ENTER (simpleact)")
        // Python: RequiresAction(check_non_temporal(addlabel(p[3],'asrt')),p[5])
        lf := addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := acfg(v17lex).NewRequiresAction(lf, $5)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        // Python: if p[1] and not check_unprovable.get(): p[0] = Sequence()
        if $1 != nil && !acfg(v17lex).CheckUnprovable {
            $$ = acfg(v17lex).NewSequence()
            $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        } else {
            $$ = a
        }
    }
    | optunprovable TOK_ENSURE labeledfmla
    {
        xtracer.Trace("parser.p_action_ensure ENTER (simpleact)")
        // Python: EnsuresAction(check_non_temporal(addlabel(p[3],'asrt')))
        lf := addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := acfg(v17lex).NewEnsuresAction(lf)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        // Python: if p[1] and not check_unprovable.get(): p[0] = Sequence()
        if $1 != nil && !acfg(v17lex).CheckUnprovable {
            $$ = acfg(v17lex).NewSequence()
            $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        } else {
            $$ = a
        }
    }
    | optunprovable TOK_ENSURE labeledfmla TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_action_ensure_proof_proofstep ENTER (simpleact)")
        // Python: EnsuresAction(check_non_temporal(addlabel(p[3],'asrt')),p[5])
        lf := addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := acfg(v17lex).NewEnsuresAction(lf, $5)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        // Python: if p[1] and not check_unprovable.get(): p[0] = Sequence()
        if $1 != nil && !acfg(v17lex).CheckUnprovable {
            $$ = acfg(v17lex).NewSequence()
            $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        } else {
            $$ = a
        }
    }
    | term TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_action_term_assign_fmla ENTER (simpleact)")
        // Python: AssignAction(p[1], check_non_temporal(p[3]))
        a := acfg(v17lex).NewAssignAction($1, checkNonTemporal($3))
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = a
    }
    | termtuple TOK_ASSIGN callatom
    {
        xtracer.Trace("parser.p_action_termtuple_assign_fmla ENTER (simpleact)")
        // Python: CallAction(*([p[3]]+list(p[1].args)))
        callArgs := append([]ast.Node{$3}, $1.Args()...)
        $$ = acfg(v17lex).NewCallAction(callArgs...)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
    }
    | term TOK_ASSIGN TOK_TIMES
    {
        xtracer.Trace("parser.p_action_term_assign_times ENTER (simpleact)")
        // Python: HavocAction(p[1])
        a := acfg(v17lex).NewHavocAction($1)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = a
    }
    | TOK_VAR tterm optinit
    {
        // Python: VarAction(p[2]) or VarAction(p[2], p[3])
        xtracer.Trace("parser.p_action_var_opttypedsym_assign_fmla ENTER (simpleact)")
        if $3 != nil {
            $$ = acfg(v17lex).NewVarAction($2, $3)
        } else {
            $$ = acfg(v17lex).NewVarAction($2)
        }
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | TOK_CALL optactualreturns callatom
    {
        xtracer.Trace("parser.p_action_call_optreturns_callatom ENTER (simpleact)")
        // Python: CallAction(*([p[3]] + p[2]))
        callArgs := append([]ast.Node{$3}, $2...)
        $$ = acfg(v17lex).NewCallAction(callArgs...)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | TOK_CALL callatom
    {
        xtracer.Trace("parser.p_action_call_callatom ENTER (simpleact)")
        // Python: CallAction(p[2])
        $$ = acfg(v17lex).NewCallAction($2)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | TOK_SET lit
    {
        xtracer.Trace("parser.p_action_set_lit ENTER (simpleact)")
        // Python: SetAction(p[2])
        a := acfg(v17lex).NewSetAction($2)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    | TOK_INSTANTIATE callatom
    {
        xtracer.Trace("parser.p_action_instantiate_atom ENTER (simpleact)")
        // Python: InstantiateAction(p[2])
        a := acfg(v17lex).NewInstantiateAction($2)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    | TOK_DEBUG SYMBOLx optdebugargs
    {
        xtracer.Trace("parser.p_simpleact_debug_symbol_optdebugargs ENTER (simpleact)")
        // Python: action = Atom(p[2],[]); action.lineno = get_lineno(p,2)
        // Python: if not p[2].startswith('"'): report_error(...)
        // Python: p[0] = DebugAction(action,*p[3]); p[0].lineno = get_lineno(p,1)
        action := acfg(v17lex).NewAtom($2.Val)
        action.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        if !strings.HasPrefix($2.Val, "\"") {
            v17lex.Error(fmt.Sprintf("expected string constant after 'debug', got %s", $2.Val))
        }
        args := append([]ast.Node{action}, $3...)
        a := acfg(v17lex).NewDebugAction(args...)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = a
    }
    | term     %prec TOK_SEMI
    {
        xtracer.Trace("parser.p_action_term ENTER (simpleact)")
        // Python: p[0] = CallAction(p[1]); p[0].lineno = p[1].lineno
        $$ = acfg(v17lex).NewCallAction($1)
        $$.SetLineno($1.GetLineno())
    }
    ;

termtuple:
    TOK_LPAREN term TOK_COMMA terms TOK_RPAREN
    {
        xtracer.Trace("parser.p_termtuple_lp_term_comma_terms_rp ENTER (termtuple)")
        args := append([]ast.Node{$2}, $4...)
        t := acfg(v17lex).NewTuple(args...)
        t.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = t
    }
    ;

debugarg:
    SYMBOLx TOK_EQ fmla
    {
        xtracer.Trace("parser.p_debugarg_symbol_equal_fmla ENTER (debugarg)")
        // Python: lhs = App(p[1]); p[0] = DebugItem(lhs, p[3])
        lhs := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil))
        lhs.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        di := acfg(v17lex).NewDebugItem(lhs, $3)
        di.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = di
    }
    ;

debugargs:
    debugarg
    {
        xtracer.Trace("parser.p_debugargs ENTER (debugargs)")
        $$ = []ast.Node{$1}
    }
    | debugargs TOK_COMMA debugarg
    {
        xtracer.Trace("parser.p_debugargs_debugarg_symbol_equal_fmla ENTER (debugargs)")
        $$ = append($1, $3)
    }
    ;

optdebugargs:
    /* empty */
    {
        xtracer.Trace("parser.p_optdebugargs ENTER (optdebugargs)")
        $$ = nil
    }
    | TOK_WITH debugargs
    {
        xtracer.Trace("parser.p_optdebugargs_with_debugargs ENTER (optdebugargs)")
        $$ = $2
    }
    ;

// ============================================================
// --- Complex actions ---
// ============================================================

complexact:
    sequence
    {
        xtracer.Trace("parser.p_action_sequence ENTER (complexact)")
        $$ = $1
    }
    | TOK_IF somefmla sequence
    {
        xtracer.Trace("parser.p_action_if_somefmla_lcb_action_rcb ENTER (complexact)")
        // Python: IfAction(cond, fix_if_part(cond, p[3]))
        cond := checkNonTemporal($2)
        body := fixIfPart(cond, $3)
        ifa := acfg(v17lex).NewIfAction(cond, body, nil)
        ifa.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = ifa
    }
    | TOK_IF somefmla sequence TOK_ELSE action
    {
        xtracer.Trace("parser.p_action_if_somefmla_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
        // Python: IfAction(cond, fix_if_part(cond, p[3]), p[5])
        cond := checkNonTemporal($2)
        body := fixIfPart(cond, $3)
        ifa := acfg(v17lex).NewIfAction(cond, body, $5)
        ifa.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = ifa
    }
    | TOK_IF TOK_TIMES sequence TOK_ELSE action
    {
        xtracer.Trace("parser.p_action_if_times_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
        // Python: ChoiceAction(p[3], p[5])
        choice := acfg(v17lex).NewChoiceAction($3, $5)
        choice.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = choice
    }
    | TOK_WHILE somefmla invariants decreases sequence
    {
        xtracer.Trace("parser.p_action_while_somefmla_invariants_decreases_lcb_action_rcb ENTER (complexact)")
        // Python: WhileAction(cond, body, *invariants, *decreases)
        cond := checkNonTemporal($2)
        body := fixIfPart($2, $5)
        args := []ast.Node{cond, body}
        args = append(args, $3...)
        args = append(args, $4...)
        w := acfg(v17lex).NewWhileAction(args...)
        w.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = w
    }
    | TOK_FOR tterm TOK_COMMA tterm TOK_IN fmla invariants decreases sequence
    {
        xtracer.Trace("parser.p_action_for_tterm_comma_tterm_in_expr_invariants_decreases_lcb_action_rcb ENTER (complexact)")

        // Python: itr,val,fmla,invars,decrs,seq = p[2],p[4],check_non_temporal(p[6]),p[7],p[8],p[9]
        itr := $2                       // *ast.App from tterm
        val := $4                       // *ast.App from tterm
        forFmla := checkNonTemporal($6) // formula
        invars := $7                    // []ast.Node from invariants
        decrs := $8                     // []ast.Node from decreases
        seq := $9                       // ast.Node from sequence

        // Python: iend = itr.rename('loc:end')
        var iend ast.Node
        if itrApp, ok := itr.(*ast.App); ok {
            iend = itrApp.Rename("loc:end")
        } else if itrAtom, ok := itr.(*ast.Atom); ok {
            iend = itrAtom.Rename("loc:end")
        } else {
            iend = itr // fallback
        }

        // Python: ln = get_lineno(p,1)
        ln := tokLineno(v17lex.(*v17LexAdapter), $1)

        // Python: didx = VarAction(itr, methcall(fmla, App('begin').sln(ln)).sln(ln)).sln(ln)
        appBegin := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("begin", nil))
        appBegin.SetLineno(ln)
        mcBegin := methcall(acfg(v17lex), forFmla, appBegin)
        mcBegin.SetLineno(ln)
        didx := acfg(v17lex).NewVarAction(itr, mcBegin)
        didx.SetLineno(ln)

        // Python: dend = VarAction(iend, methcall(fmla, App('end').sln(ln)).sln(ln)).sln(ln)
        appEnd := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("end", nil))
        appEnd.SetLineno(ln)
        mcEnd := methcall(acfg(v17lex), forFmla, appEnd)
        mcEnd.SetLineno(ln)
        dend := acfg(v17lex).NewVarAction(iend, mcEnd)
        dend.SetLineno(ln)

        // Python: dval = VarAction(val, methcall(fmla, App('value', itr).sln(ln)).sln(ln)).sln(ln)
        appValue := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("value", nil), itr)
        appValue.SetLineno(ln)
        mcValue := methcall(acfg(v17lex), forFmla, appValue)
        mcValue.SetLineno(ln)
        dval := acfg(v17lex).NewVarAction(val, mcValue)
        dval.SetLineno(ln)

        // Python: incr = AssignAction(itr, methcall(itr, App('next').sln(ln)).sln(ln)).sln(ln)
        appNext := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("next", nil))
        appNext.SetLineno(ln)
        mcNext := methcall(acfg(v17lex), itr, appNext)
        mcNext.SetLineno(ln)
        incr := acfg(v17lex).NewAssignAction(itr, mcNext)
        incr.SetLineno(ln)

        // Python: body = Sequence(*lower_var_stmts([dval, seq, incr])).sln(ln)
        bodyStmts := ast.LowerVarStatements([]ast.Node{dval, seq, incr})
        body := acfg(v17lex).NewSequence(bodyStmts...)
        body.SetLineno(ln)

        // Python: loop = WhileAction(*([App('<', itr, iend).sln(ln), body] + invars + decrs)).sln(ln)
        ltCond := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("<", nil), itr, iend)
        ltCond.SetLineno(ln)
        loopArgs := []ast.Node{ltCond, body}
        loopArgs = append(loopArgs, invars...)
        loopArgs = append(loopArgs, decrs...)
        loop := acfg(v17lex).NewWhileAction(loopArgs...)
        loop.SetLineno(ln)

        // Python: p[0] = Sequence(*lower_var_stmts([didx, dend, loop])).sln(ln)
        outerStmts := ast.LowerVarStatements([]ast.Node{didx, dend, loop})
        result := acfg(v17lex).NewSequence(outerStmts...)
        result.SetLineno(ln)
        $$ = result
    }
    | TOK_LOCAL lparams sequence
    {
        xtracer.Trace("parser.p_action_local_params_lcb_action_rcb ENTER (complexact)")
        // Python: lsyms = [s.prefix('loc:') for s in p[2]]
        // Python: subst = dict((x.rep,y.rep) for x,y in zip(p[2],lsyms))
        // Python: action = subst_prefix_atoms_ast(p[3],subst,None,None)
        // Python: p[0] = LocalAction(*(lsyms+[action]))
        bounds := $2
        lsyms := make([]ast.Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = ast.PrefixNode(s, "loc:")
            subst[ast.NodeRep(s)] = ast.NodeRep(lsyms[i])
        }
        action := ast.SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        args := append(lsyms, action)
        la := acfg(v17lex).NewLocalAction("parser.local_action_rule", args...)
        la.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = la
    }
    | TOK_LET eqns sequence
    {
        xtracer.Trace("parser.p_action_let_eqns_lcb_action_rcb ENTER (complexact)")
        // Python: LetAction(*(p[2]+[p[3]]))
        args := append($2, $3)
        la := acfg(v17lex).NewLetAction(args...)
        la.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = la
    }
    | TOK_THUNK labelname SYMBOLx optargs TOK_COLON atype TOK_ASSIGN sequence
    {
        xtracer.Trace("parser.p_action_thunk_symbol_optargs_colon_atype_assign_sequence ENTER (complexact)")
        // Python: action = Atom(p[3], p[4]); action.lineno = get_lineno(p,3)
        // Python: ThunkAction(Atom(p[2][1:-1],[]), action, Atom(p[6]), p[8])
        // Brackets already stripped by labelname rule
        label := acfg(v17lex).NewAtom($2.Val)
        label.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        actionAtom := acfg(v17lex).NewAtom($3.Val, $4...)
        actionAtom.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
        sortAtom := acfg(v17lex).NewAtom(ast.NodeRep($6))
        ta := acfg(v17lex).NewThunkAction(label, actionAtom, sortAtom, $8)
        ta.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = ta
    }
    ;

// --- somefmla ---

somefmla:
    fmla
    {
        xtracer.Trace("parser.p_somefmla_fmla ENTER (somefmla)")
        $$ = $1
    }
    | fmla TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_somefmla_fmla_assign_fmla ENTER (somefmla)")
        // Python: lsyms = [p[1].prefix('loc:')]
        // Python: lsyms[0].sort = p[1].sort
        // Python: subst = dict((x.rep,y.rep) for x,y in zip([p[1]],lsyms))
        // Python: fmla = App('*>',p[3],p[1])
        // Python: fmla = subst_prefix_atoms_ast(fmla,subst,None,None)
        // Python: p[0] = Some(*(lsyms+[fmla]))
        lhs := $1
        lsym := ast.PrefixNode(lhs, "loc:")
        if a, ok := lsym.(*ast.Atom); ok {
            if orig, ok2 := lhs.(*ast.Atom); ok2 {
                a.ASort = orig.ASort
            }
        }
        subst := map[string]string{ast.NodeRep(lhs): ast.NodeRep(lsym)}
        fmla := acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("*>", nil), $3, lhs)
        fmla.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        fmla2 := ast.SubstPrefixAtomsAst(fmla, subst, nil, nil, nil)
        some := acfg(v17lex).NewSome([]ast.Node{lsym}, fmla2)
        some.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = some
    }
    | TOK_SOME bounds fmla
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla ENTER (somefmla)")
        bounds := $2
        lsyms := make([]ast.Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = ast.PrefixNode(s, "loc:")
            subst[ast.NodeRep(s)] = ast.NodeRep(lsyms[i])
        }
        fmla := ast.SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        some := acfg(v17lex).NewSome(lsyms, fmla)
        some.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = some
    }
    | TOK_SOME bounds fmla TOK_MINIMIZING term
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla_minimizing_term ENTER (somefmla)")
        bounds := $2
        lsyms := make([]ast.Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = ast.PrefixNode(s, "loc:")
            subst[ast.NodeRep(s)] = ast.NodeRep(lsyms[i])
        }
        fmla := ast.SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        index := ast.SubstPrefixAtomsAst($5, subst, nil, nil, nil)
        smin := acfg(v17lex).NewSomeMin(lsyms, fmla, index)
        smin.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = smin
    }
    | TOK_SOME bounds fmla TOK_MAXIMIZING term
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla_maximizing_term ENTER (somefmla)")
        bounds := $2
        lsyms := make([]ast.Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = ast.PrefixNode(s, "loc:")
            subst[ast.NodeRep(s)] = ast.NodeRep(lsyms[i])
        }
        fmla := ast.SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        index := ast.SubstPrefixAtomsAst($5, subst, nil, nil, nil)
        smax := acfg(v17lex).NewSomeMax(lsyms, fmla, index)
        smax.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = smax
    }
    ;

bounds:
    params TOK_DOT
    {
        xtracer.Trace("parser.p_bounds_params_dot ENTER (bounds)")
        $$ = $1
    }
    | TOK_LPAREN lparams TOK_RPAREN
    {
        xtracer.Trace("parser.p_bounds_lparen_lparams_rparen ENTER (bounds)")
        $$ = $2
    }
    ;

invariants:
    /* empty */
    {
        xtracer.Trace("parser.p_invariants ENTER (invariants)")
        $$ = nil
    }
    | invariants TOK_INVARIANT labeledfmla
    {
        xtracer.Trace("parser.p_invariant_invariant_fmla ENTER (invariants)")
        // Python: a = AssertAction(check_non_temporal(addlabel(p[3],'asrt')))
        inv := checkNonTemporal(addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "asrt"))
        a := acfg(v17lex).NewAssertAction(inv)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = append($1, a)
    }
    | invariants TOK_INVARIANT labeledfmla TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_invariant_invariant_fmla_proof ENTER (invariants)")
        // Python: a = AssertAction(inv, p[5])
        inv := checkNonTemporal(addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "asrt"))
        a := acfg(v17lex).NewAssertAction(inv, $5)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = append($1, a)
    }
    ;

decreases:
    /* empty */
    {
        xtracer.Trace("parser.p_decreases ENTER (decreases)")
        $$ = nil
    }
    | TOK_DECREASES fmla
    {
        xtracer.Trace("parser.p_decreases_decreases_fmla ENTER (decreases)")
        // Python: rank = Ranking(check_non_temporal(p[2])); rank.lineno = get_lineno(p,1)
        fmla := checkNonTemporal($2)
        rank := acfg(v17lex).NewRanking(fmla)
        rank.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = []ast.Node{rank}
    }
    ;

// --- eqn / eqns ---

eqn:
    SYMBOLx TOK_EQ SYMBOLx
    {
        xtracer.Trace("parser.p_eqn_SYMBOL_EQ_SYMBOL ENTER (eqn)")
        // Python: Equals(App(p[1]), App(p[3])) — Equals = lg.Eq, AST equivalent is Atom("=", lhs, rhs)
        $$ = acfg(v17lex).NewAtom("=", acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1.Val, nil)), acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($3.Val, nil)))
    }
    ;

eqns:
    eqn
    {
        xtracer.Trace("parser.p_eqns_eqn ENTER (eqns)")
        $$ = []ast.Node{$1}
    }
    | eqns TOK_COMMA eqn
    {
        xtracer.Trace("parser.p_eqns_eqns_comma_eqn ENTER (eqns)")
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Scenario ---
// ============================================================

sceninit:
    TOK_ARROW places
    {
        xtracer.Trace("parser.p_sceninit_arrow_places ENTER (sceninit)")
        pl := acfg(v17lex).NewPlaceList($2)
        pl.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = pl
    }
    ;

places:
    SYMBOLx
    {
        xtracer.Trace("parser.p_places_symbol ENTER (places)")
        a := acfg(v17lex).NewAtom($1.Val)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = []ast.Node{a}
    }
    | places TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_places_places_comma_symbol ENTER (places)")
        a := acfg(v17lex).NewAtom($3.Val)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
        $$ = append($1, a)
    }
    ;

scentranss:
    /* empty */
    {
        xtracer.Trace("parser.p_scentranss ENTER (scentranss)")
        $$ = nil
    }
    | scentranss scentrans
    {
        xtracer.Trace("parser.p_scentranss__scentranss_scentrans ENTER (scentranss)")
        $$ = append($1, $2)
    }
    | scentranss places TOK_ARROW places TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentranss_scentranss_places_arrow_places_colon_scenariomixin ENTER (scentranss)")
        from := acfg(v17lex).NewPlaceList($2)
        to := acfg(v17lex).NewPlaceList($4)
        tr := acfg(v17lex).NewScenarioTransition(from, to, $6)
        tr.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $5))
        $$ = append($1, tr)
    }
    | scentranss places TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentranss_scentranss_places_colon_scenariomixin ENTER (scentranss)")
        to := acfg(v17lex).NewPlaceList($2)
        tr := acfg(v17lex).NewScenarioTransition(nil, to, $4)
        tr.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
        $$ = append($1, tr)
    }
    ;

scentrans:
    places TOK_ARROW places TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentrans__places_arrow_places_colon_scenariomixin ENTER (scentrans)")
        from := acfg(v17lex).NewPlaceList($1)
        to := acfg(v17lex).NewPlaceList($3)
        $$ = acfg(v17lex).NewScenarioTransition(from, to, $5)
    }
    | places TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentrans__places_colon_scenariomixin ENTER (scentrans)")
        from := acfg(v17lex).NewPlaceList($1)
        $$ = acfg(v17lex).NewScenarioTransition(from, acfg(v17lex).NewPlaceList(nil), $3)
    }
    ;

scenariomixin:
    TOK_BEFORE atype optargs optreturns sequence
    {
        xtracer.Trace("parser.p_scenariomixin_before_callatom_lcb_action_rcb ENTER (scenariomixin)")
        atom := acfg(v17lex).NewAtom($2.(*ast.Symbol).Rep)
        atom.SetLineno(nodeLineno($2))
        mixer := makeMixinName(acfg(v17lex), atom, "before")
        // Python: optargs, optreturns = infer_action_params(atom.rep, p[3], p[4])
        formals, returns := inferActionParams(v17lex.(*v17LexAdapter).accum, atom.Rep, $3, $4)
        adef := acfg(v17lex).NewActionDef(atom, $5, formals, returns)
        sbm := acfg(v17lex).NewScenarioBeforeMixin(mixer, adef)
        sbm.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = sbm
    }
    | TOK_AFTER atype optargs optreturns sequence
    {
        xtracer.Trace("parser.p_scenariomixin_after_callatom_lcb_action_rcb ENTER (scenariomixin)")
        atom := acfg(v17lex).NewAtom($2.(*ast.Symbol).Rep)
        atom.SetLineno(nodeLineno($2))
        mixer := makeMixinName(acfg(v17lex), atom, "after")
        // Python: optargs, optreturns = infer_action_params(atom.rep, p[3], p[4])
        formals, returns := inferActionParams(v17lex.(*v17LexAdapter).accum, atom.Rep, $3, $4)
        adef := acfg(v17lex).NewActionDef(atom, $5, formals, returns)
        sam := acfg(v17lex).NewScenarioAfterMixin(mixer, adef)
        sam.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = sam
    }
    ;

// ============================================================
// --- Proof / tactic ---
// ============================================================

pflet:
    var TOK_EQ fmla
    {
        xtracer.Trace("parser.p_pflet_var_eq_fmla ENTER (pflet)")
        d := acfg(v17lex).NewDefinition($1, $3)
        d.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = d
    }
    ;

pflets:
    pflet
    {
        xtracer.Trace("parser.p_pflets_pflet ENTER (pflets)")
        $$ = []ast.Node{$1}
    }
    | pflets TOK_COMMA pflet
    {
        xtracer.Trace("parser.p_pflets_pflets_pflet ENTER (pflets)")
        $$ = append($1, $3)
    }
    ;

tacticwithelem:
    TOK_INVARIANT labeledfmla
    {
        xtracer.Trace("parser.p_tacticwithelem_invariant ENTER (tacticwithelem)")
        // Python: p[0] = addlabel(p[2], 'invar')
        $$ = addLabel(acfg(v17lex), $2.(*ast.LabeledFormula), "invar")
    }
    | TOK_DEFINITION typeddefn TOK_EQ fmla
    {
        xtracer.Trace("parser.p_tacticwithelem_fun_defn ENTER (tacticwithelem)")
        df := acfg(v17lex).NewDefinition(ast.AppToAtom($2), $4)
        df.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
        lf := addLabel(acfg(v17lex), mkLF(acfg(v17lex), df), "def")
        $$ = acfg(v17lex).NewDerivedDecl(lf)
    }
    | TOK_TRIGGER atype TOK_WITH terms
    {
        xtracer.Trace("parser.p_tacticwithelem_trigger ENTER (tacticwithelem)")
        // Python: p[0] = Trigger(*([Atom(p[2])]+p[4])); p[0].lineno = get_lineno(p,3)
        trigAtom := atypeToAtom(acfg(v17lex), $2)
        trig := acfg(v17lex).NewTrigger(trigAtom, $4...)
        trig.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
        $$ = trig
    }
    ;

tacticwithlist:
    tacticwithelem
    {
        xtracer.Trace("parser.p_tactwithlist_tacticwithelem ENTER (tacticwithlist)")
        $$ = []ast.Node{$1}
    }
    | tacticwithlist tacticwithelem
    {
        xtracer.Trace("parser.p_tactwithlist_tactwithlist_tacticwithelem ENTER (tacticwithlist)")
        $$ = append($1, $2)
    }
    ;

tacticwithlistchoice:
    tacticwithlist
    {
        xtracer.Trace("parser.p_tacticwithlistchoice_tactwithlist ENTER (tacticwithlistchoice)")
        $$ = acfg(v17lex).NewTacticWith($1)
    }
    | pflets
    {
        xtracer.Trace("parser.p_tacticwithlistchoice_pflets ENTER (tacticwithlistchoice)")
        $$ = acfg(v17lex).NewTacticLets($1)
    }
    ;

opttacticwith:
    /* empty */
    {
        xtracer.Trace("parser.p_opttacticwith ENTER (opttacticwith)")
        $$ = acfg(v17lex).NewTacticWith(nil)
    }
    | TOK_WITH tacticwithlistchoice
    {
        xtracer.Trace("parser.p_opttacticwith_with_tacticwithlist ENTER (opttacticwith)")
        $$ = $2
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    | TOK_WITH TOK_LCB tacticwithlist TOK_RCB
    {
        xtracer.Trace("parser.p_opttacticwith_with_lcb_tacticwithlist_rcb ENTER (opttacticwith)")
        tw := acfg(v17lex).NewTacticWith($3)
        tw.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = tw
    }
    ;

proofgroup:
    TOK_LCB proofseq TOK_RCB
    {
        xtracer.Trace("parser.p_proofgroup_lcb_proofseq_rcb ENTER (proofgroup)")
        $$ = $2
    }
    | TOK_LCB TOK_RCB
    {
        xtracer.Trace("parser.p_proofgroup_lcb_rcb ENTER (proofgroup)")
        $$ = acfg(v17lex).NewNullTactic()
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    }
    ;

optproofgroup:
    /* empty */
    {
        xtracer.Trace("parser.p_optproofgroup ENTER (optproofgroup)")
        $$ = nil
    }
    | TOK_PROOF proofgroup
    {
        xtracer.Trace("parser.p_optproofgroup_symbol ENTER (optproofgroup)")
        $$ = $2
    }
    ;

proofseq:
    proofstep
    {
        xtracer.Trace("parser.p_proofseq_proofstep ENTER (proofseq)")
        $$ = $1
    }
    | proofseq optsemi proofstep
    {
        xtracer.Trace("parser.p_proofseq_proofseq_semi_proofstep ENTER (proofseq)")
        $$ = acfg(v17lex).NewComposeTactics([]ast.Node{$1, $3})
        $$.SetLineno(nodeLineno($2))
    }
    ;

// --- match/renaming ---

match:
    defn
    {
        xtracer.Trace("parser.p_match_defn ENTER (match)")
        $$ = $1
    }
    | var TOK_EQ fmla
    {
        xtracer.Trace("parser.p_match_var_eq_fmla ENTER (match)")
        // Python: Definition(p[1], check_non_temporal(p[3]))
        $$ = acfg(v17lex).NewDefinition($1, checkNonTemporal($3))
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
    }
    ;

matches:
    match
    {
        xtracer.Trace("parser.p_matches ENTER (matches)")
        $$ = []ast.Node{$1}
    }
    | matches TOK_COMMA match
    {
        xtracer.Trace("parser.p_matches_matches_comma_match ENTER (matches)")
        $$ = append($1, $3)
    }
    ;

renamingitem:
    TOK_VARIABLE TOK_DIV TOK_VARIABLE
    {
        xtracer.Trace("parser.p_renamingitem_variable_div_variable ENTER (renamingitem)")
        // Python: Definition(Variable(p[3],universe),Variable(p[1],universe))
        $$ = acfg(v17lex).NewDefinition(acfg(v17lex).NewVariable($3.Val, "S"), acfg(v17lex).NewVariable($1.Val, "S"))
    }
    | SYMBOLx TOK_DIV SYMBOLx
    {
        xtracer.Trace("parser.p_renamingitem_symbol_div_symbol ENTER (renamingitem)")
        $$ = acfg(v17lex).NewDefinition(acfg(v17lex).NewAtom($3.Val), acfg(v17lex).NewAtom($1.Val))
    }
    ;

renaminglist:
    renamingitem
    {
        xtracer.Trace("parser.p_renaminglist_renamingitem ENTER (renaminglist)")
        $$ = []ast.Node{$1}
    }
    | renaminglist TOK_COMMA renamingitem
    {
        xtracer.Trace("parser.p_renaminglist_renaminglist_comma_renamingitem ENTER (renaminglist)")
        $$ = append($1, $3)
    }
    ;

optrenaming:
    /* empty */
    {
        xtracer.Trace("parser.p_renaming ENTER (optrenaming)")
        $$ = acfg(v17lex).NewRenaming(nil)
    }
    | renaming
    {
        xtracer.Trace("parser.p_optrenaming_renaming ENTER (optrenaming)")
        $$ = $1
    }
    ;

renaming:
    TOK_LT renaminglist TOK_GT
    {
        xtracer.Trace("parser.p_renaming_lt_renaminglist_gt ENTER (renaming)")
        r := acfg(v17lex).NewRenaming($2)
        $$ = r
    }
    ;

// --- renamings (zero or more renamings, for unfold specs) ---
// Matches Python renamings : /* empty */ | renamings renaming (ivy_parser.py:1414-1423)

renamings:
    /* empty */
    {
        xtracer.Trace("parser.p_renamings ENTER (renamings)")
        $$ = nil
    }
    | renamings renaming
    {
        xtracer.Trace("parser.p_renamings_renamings_renaming ENTER (renamings)")
        $$ = append($1, $2)
    }
    ;

// --- unfold spec / unfold specs ---
// Matches Python unfspec : callatom renamings (ivy_parser.py:1425-1440)

unfspec:
    callatom renamings
    {
        xtracer.Trace("parser.p_unfspec_callatom_renamings ENTER (unfspec)")
        $$ = acfg(v17lex).NewUnfoldSpec($1, $2)
    }
    ;

unfspecs:
    unfspec
    {
        xtracer.Trace("parser.p_unfspecs_unfspec ENTER (unfspecs)")
        $$ = []ast.Node{$1}
    }
    | unfspecs TOK_COMMA unfspec
    {
        xtracer.Trace("parser.p_unfspecs_unfspecs_unfspec ENTER (unfspecs)")
        $$ = append($1, $3)
    }
    ;

// --- proofstep ---

proofstep:
    TOK_APPLY atype optrenaming
    {
        xtracer.Trace("parser.p_proofstep_symbol ENTER (proofstep)")
        // Python: a = Atom(p[2]); a.lineno = get_lineno(p,2)
        // Python: p[0] = SchemaInstantiation(a, p[3]); p[0].lineno = get_lineno(p,1)
        a := atypeToAtom(acfg(v17lex), $2)
        a.SetLineno(nodeLineno($2))
        si := acfg(v17lex).NewSchemaInstantiation(a, $3)
        si.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = si
    }
    | TOK_APPLY atype optrenaming TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_symbol_with_defns ENTER (proofstep)")
        a := atypeToAtom(acfg(v17lex), $2)
        a.SetLineno(nodeLineno($2))
        si := acfg(v17lex).NewSchemaInstantiationWithMatches(a, $3, $5)
        si.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = si
    }
    | TOK_ASSUME atype optrenaming
    {
        xtracer.Trace("parser.p_proofstep_assume ENTER (proofstep)")
        // Python: AssumeGlobalTactic(a, p[3]); p[0].label = NoneAST()
        a := atypeToAtom(acfg(v17lex), $2)
        a.SetLineno(nodeLineno($2))
        at := acfg(v17lex).NewAssumeGlobalTactic(a, $3)
        at.TLabel = acfg(v17lex).NewNoneAST()
        at.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = at
    }
    | TOK_ASSUME atype optrenaming TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_assume_with_defns ENTER (proofstep)")
        // Python: AssumeGlobalTactic(*([a,p[3]]+p[5])); p[0].label = NoneAST()
        a := atypeToAtom(acfg(v17lex), $2)
        a.SetLineno(nodeLineno($2))
        at := acfg(v17lex).NewAssumeGlobalTacticWithMatches(a, $3, $5)
        at.TLabel = acfg(v17lex).NewNoneAST()
        at.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = at
    }
    | TOK_INSTANTIATE atype optrenaming
    {
        xtracer.Trace("parser.p_proofstep_instantiate ENTER (proofstep)")
        a := atypeToAtom(acfg(v17lex), $2)
        a.SetLineno(nodeLineno($2))
        at := acfg(v17lex).NewAssumeTactic(a, $3)
        at.TLabel = acfg(v17lex).NewNoneAST()
        at.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = at
    }
    | TOK_INSTANTIATE labelname atype optrenaming
    {
        xtracer.Trace("parser.p_proofstep_instance ENTER (proofstep)")
        lex := v17lex.(*v17LexAdapter)
        a := atypeToAtom(acfg(v17lex), $3)
        a.SetLineno(tokLineno(lex, $2))
        label := acfg(v17lex).NewAtom($2.Val)
        label.SetLineno(tokLineno(lex, $2))
        at := acfg(v17lex).NewAssumeTactic(a, $4)
        at.TLabel = label
        at.SetLineno(tokLineno(lex, $1))
        $$ = at
    }
    | TOK_INSTANTIATE labelname atype optrenaming TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_instance_with_matches ENTER (proofstep)")
        lex := v17lex.(*v17LexAdapter)
        a := atypeToAtom(acfg(v17lex), $3)
        a.SetLineno(tokLineno(lex, $2))
        label := acfg(v17lex).NewAtom($2.Val)
        label.SetLineno(tokLineno(lex, $2))
        at := acfg(v17lex).NewAssumeTacticWithMatches(a, $4, $6)
        at.TLabel = label
        at.SetLineno(tokLineno(lex, $1))
        $$ = at
    }
    | TOK_INSTANTIATE atype optrenaming TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_instantiate_with_defns ENTER (proofstep)")
        a := atypeToAtom(acfg(v17lex), $2)
        a.SetLineno(nodeLineno($2))
        at := acfg(v17lex).NewAssumeTacticWithMatches(a, $3, $5)
        at.TLabel = acfg(v17lex).NewNoneAST()
        at.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = at
    }
    | TOK_INSTANTIATE TOK_WITH pflets
    {
        xtracer.Trace("parser.p_proofstep_witness_pflets ENTER (proofstep)")
        wt := acfg(v17lex).NewWitnessTactic($3)
        wt.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = wt
    }
    | TOK_SHOWGOALS
    {
        xtracer.Trace("parser.p_proofstep_showgoals ENTER (proofstep)")
        sg := acfg(v17lex).NewShowGoalsTactic()
        sg.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = sg
    }
    | TOK_DEFERGOAL
    {
        xtracer.Trace("parser.p_proofstep_defergoal ENTER (proofstep)")
        dg := acfg(v17lex).NewDeferGoalTactic()
        dg.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = dg
    }
    | TOK_SPOIL atype
    {
        xtracer.Trace("parser.p_proofstep_spoil_atype ENTER (proofstep)")
        // Python: a = Atom(p[2]) where p[2] is a string from atype
        a := atypeToAtom(acfg(v17lex), $2)
        a.SetLineno(nodeLineno($2))
        st := acfg(v17lex).NewSpoilTactic(a)
        st.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = st
    }
    | TOK_TACTIC SYMBOLx opttacticwith optproofgroup
    {
        xtracer.Trace("parser.p_proofstep_tactic ENTER (proofstep)")
        a := acfg(v17lex).NewAtom($2.Val)
        a.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        proof := ast.Node($4)
        if proof == nil {
            proof = acfg(v17lex).NewNoneAST()
        }
        tt := acfg(v17lex).NewTacticTactic(a, $3, proof)
        tt.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = tt
    }
    | opttemporal TOK_PROPERTY labeledfmla optskolem optproofgroup
    {
        xtracer.Trace("parser.p_proofstep_property ENTER (proofstep)")
        lf := addLabel(acfg(v17lex), $3.(*ast.LabeledFormula), "prop")
        // Python: prop = addtemporal(lf) if p[1] else check_non_temporal(lf)
        if $1 != nil {
            lf = addTemporal(lf)
        } else {
            checkNonTemporal(lf)
        }
        name := ast.Node($4)
        if name == nil { name = acfg(v17lex).NewNoneAST() }
        proof := ast.Node($5)
        if proof == nil { proof = acfg(v17lex).NewNoneAST() }
        pt := acfg(v17lex).NewPropertyTactic(lf, name, proof)
        pt.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = pt
    }
    | TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_proofstep_function ENTER (proofstep)")
        ft := acfg(v17lex).NewFunctionTactic($2)
        ft.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = ft
    }
    | TOK_THEOREM lgprop optproofgroup
    {
        xtracer.Trace("parser.p_proofstep_theorem ENTER (proofstep)")
        lf := addLabel(acfg(v17lex), $2.(*ast.LabeledFormula), "thm")
        proof := ast.Node($3)
        if proof == nil { proof = acfg(v17lex).NewNoneAST() }
        pt := acfg(v17lex).NewPropertyTactic(lf, acfg(v17lex).NewNoneAST(), proof)
        pt.SetLineno(nodeLineno($2))
        $$ = pt
    }
    | TOK_PROOF labelname proofgroup
    {
        xtracer.Trace("parser.p_proofstep_proof ENTER (proofstep)")
        lex := v17lex.(*v17LexAdapter)
        label := acfg(v17lex).NewAtom($2.Val)
        label.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        pt := acfg(v17lex).NewProofTactic(label, $3)
        pt.SetLineno(tokLineno(lex, $1))
        $$ = pt
    }
    | TOK_LET pflets
    {
        xtracer.Trace("parser.p_proofstep_let_pflets ENTER (proofstep)")
        lt := acfg(v17lex).NewLetTactic($2)
        lt.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = lt
    }
    | TOK_IF fmla proofgroup TOK_ELSE proofgroup
    {
        xtracer.Trace("parser.p_proofstep_if_fmla_proofgroup_else_proofgroup ENTER (proofstep)")
        it := acfg(v17lex).NewIfTactic($2, $3, $5)
        it.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = it
    }
    | TOK_UNFOLD atype TOK_WITH unfspecs
    {
        xtracer.Trace("parser.p_proofstep_unfold_atype_with_defns ENTER (proofstep)")
        a := atypeToAtom(acfg(v17lex), $2)
        a.SetLineno(nodeLineno($2))
        ut := acfg(v17lex).NewUnfoldTactic(a, $4)
        ut.TLabel = acfg(v17lex).NewNoneAST()
        ut.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = ut
    }
    | TOK_UNFOLD TOK_WITH unfspecs
    {
        xtracer.Trace("parser.p_proofstep_unfold_with_defns ENTER (proofstep)")
        ut := acfg(v17lex).NewUnfoldTactic(acfg(v17lex).NewNoneAST(), $3)
        ut.TLabel = acfg(v17lex).NewNoneAST()
        ut.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = ut
    }
    | TOK_FORGET callatoms
    {
        xtracer.Trace("parser.p_proofstep_forget_callatoms ENTER (proofstep)")
        ft := acfg(v17lex).NewForgetTactic($2)
        ft.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = ft
    }
    | proofgroup
    {
        xtracer.Trace("parser.p_proofstep_proofgroup ENTER (proofstep)")
        $$ = $1
    }
    ;

// ============================================================
// --- Update / State / Concept (less common) ---
// ============================================================

requires:
    /* empty */
    {
        xtracer.Trace("parser.p_requires ENTER (requires)")
        $$ = acfg(v17lex).NewAnd()
    }
    | TOK_REQUIRES fmla
    {
        xtracer.Trace("parser.p_requires_requires_fmla ENTER (requires)")
        $$ = $2
    }
    ;

ensures:
    TOK_ENSURES fmla
    {
        xtracer.Trace("parser.p_ensures_ensures_fmla ENTER (ensures)")
        $$ = $2
    }
    ;

modifies:
    /* empty */
    {
        xtracer.Trace("parser.p_modifies ENTER (modifies)")
        $$ = nil
    }
    | TOK_MODIFIES TOK_LCB TOK_RCB
    {
        xtracer.Trace("parser.p_modifies_modifies_lcb_rcb ENTER (modifies)")
        $$ = acfg(v17lex).NewAnd() // empty modifies
    }
    | TOK_MODIFIES TOK_TIMES
    {
        xtracer.Trace("parser.p_modifies_modofies_times ENTER (modifies)")
        $$ = nil
    }
    | TOK_MODIFIES atoms
    {
        xtracer.Trace("parser.p_modifies_modifies_atoms ENTER (modifies)")
        $$ = acfg(v17lex).NewAnd($2...)
    }
    ;

upaxes:
    /* empty */
    {
        xtracer.Trace("parser.p_upaxes ENTER (upaxes)")
        $$ = nil
    }
    | upaxes upax
    {
        xtracer.Trace("parser.p_upaxes_upaxes_upax ENTER (upaxes)")
        $$ = append($1, $2)
    }
    ;

upax:
    TOK_PARAMS tterms TOK_IN action TOK_ARROW requires ensures
    {
        xtracer.Trace("parser.p_upax_params_apps_in_action_arrow_ensures_fmla ENTER (upax)")
        $$ = acfg(v17lex).NewAtom("upax", $6, $7)
    }
    ;

assert_rhs:
    TOK_LCB requires modifies ensures TOK_RCB
    {
        xtracer.Trace("parser.p_assert_rhs_lcb_requires_modifies_ensures_rcb ENTER (assert_rhs)")
        $$ = acfg(v17lex).NewAtom("rme", $2, $4)
    }
    | fmla
    {
        xtracer.Trace("parser.p_assert_rhs_fmla ENTER (assert_rhs)")
        $$ = $1
    }
    ;

state_expr:
    TOK_TRUE
    {
        xtracer.Trace("parser.p_state_expr_true ENTER (state_expr)")
        $$ = acfg(v17lex).NewAnd()
    }
    | TOK_FALSE
    {
        xtracer.Trace("parser.p_state_expr_false ENTER (state_expr)")
        $$ = acfg(v17lex).NewOr()
    }
    | SYMBOLx
    {
        xtracer.Trace("parser.p_state_expr_symbol ENTER (state_expr)")
        $$ = acfg(v17lex).NewAtom($1.Val)
    }
    | SYMBOLx TOK_LPAREN state_expr TOK_RPAREN
    {
        xtracer.Trace("parser.p_state_expr_symbol_lparen_state_expr_rparen ENTER (state_expr)")
        $$ = acfg(v17lex).NewAtom($1.Val, $3)
    }
    | state_expr TOK_OR state_expr
    {
        xtracer.Trace("parser.p_state_expr_state_expr_or_state_expr ENTER (state_expr)")
        $$ = acfg(v17lex).NewOr($1, $3)
    }
    | TOK_LCB requires modifies ensures TOK_RCB
    {
        xtracer.Trace("parser.p_state_expr_lcb_requires_modifies_ensures_rcb ENTER (state_expr)")
        $$ = acfg(v17lex).NewAtom("rme", $2, $4)
    }
    | TOK_ENTRY
    {
        xtracer.Trace("parser.p_state_expr_entry ENTER (state_expr)")
        $$ = acfg(v17lex).NewAtom("entry")
    }
    ;

// --- Concept space ---

cdefn:
    atom TOK_EQ expr
    {
        xtracer.Trace("parser.p_cdefn_atom_expr ENTER (cdefn)")
        $$ = acfg(v17lex).NewDefinition(ast.AppToAtom($1), $3)
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
    }
    ;

cdefns:
    cdefn
    {
        xtracer.Trace("parser.p_cdefns_cdefn ENTER (cdefns)")
        $$ = []ast.Node{$1}
    }
    | cdefns TOK_COMMA cdefn
    {
        xtracer.Trace("parser.p_cdefns_cdefns_comma_cdefn ENTER (cdefns)")
        $$ = append($1, $3)
    }
    ;

expr:
    TOK_LCB fmla TOK_RCB
    {
        xtracer.Trace("parser.p_expr_fmla ENTER (expr)")
        $$ = $2
    }
    | exprterm
    {
        xtracer.Trace("parser.p_expr_exprterm ENTER (expr)")
        $$ = $1
    }
    | exprterm relop exprterm
    {
        xtracer.Trace("parser.p_expr_exprterm_relop_exprterm ENTER (expr)")
        $$ = acfg(v17lex).NewAtom($2, $1, $3)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | exprterm TOK_TILDAEQ exprterm
    {
        xtracer.Trace("parser.p_expr_exprterm_tildaeq_exprterm ENTER (expr)")
        n := acfg(v17lex).NewNot(acfg(v17lex).NewAtom("=", $1, $3))
        n.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
        $$ = n
    }
    | TOK_TILDA expr
    {
        xtracer.Trace("parser.p_expr_tilda_atom ENTER (expr)")
        $$ = acfg(v17lex).NewNot($2)
    }
    | TOK_LPAREN expr TOK_RPAREN
    {
        xtracer.Trace("parser.p_expr_lparen_expr_rparen ENTER (expr)")
        $$ = $2
    }
    | prod
    {
        xtracer.Trace("parser.p_expr_prod ENTER (expr)")
        $$ = acfg(v17lex).NewAtom("product", $1...)
    }
    | sum
    {
        xtracer.Trace("parser.p_expr_sum ENTER (expr)")
        $$ = acfg(v17lex).NewAtom("sum", $1...)
    }
    ;

exprterm:
    aterm
    {
        xtracer.Trace("parser.p_exprterm_aterm ENTER (exprterm)")
        $$ = $1
    }
    | var
    {
        xtracer.Trace("parser.p_exprterm_var ENTER (exprterm)")
        $$ = $1
    }
    ;

prod:
    expr TOK_TIMES expr
    {
        xtracer.Trace("parser.p_prod_expr_expr ENTER (prod)")
        $$ = []ast.Node{$1, $3}
    }
    | prod TOK_TIMES expr
    {
        xtracer.Trace("parser.p_prod_prod_expr ENTER (prod)")
        $$ = append($1, $3)
    }
    ;

sum:
    expr TOK_PLUS expr
    {
        xtracer.Trace("parser.p_sum_expr_expr ENTER (sum)")
        $$ = []ast.Node{$1, $3}
    }
    | sum TOK_PLUS expr
    {
        xtracer.Trace("parser.p_sum_sum_expr ENTER (sum)")
        $$ = append($1, $3)
    }
    ;

// --- loc (for old v1.1 compat) ---

loc:
    /* empty */
    {
        xtracer.Trace("parser.p_loc ENTER (loc)")
        $$ = ""
    }
    | SYMBOLx
    {
        xtracer.Trace("parser.p_loc_symbol ENTER (loc)")
        $$ = $1.Val
    }
    ;

// --- symbols list ---

symbols:
    SYMBOLx
    {
        xtracer.Trace("parser.p_symbols ENTER (symbols)")
        $$ = []ast.Node{acfg(v17lex).NewAtom($1.Val)}
    }
    | symbols TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_symbols_symbols_symbol ENTER (symbols)")
        $$ = append($1, acfg(v17lex).NewAtom($3.Val))
    }
    ;

%%

// lalrMakeSequence wraps a list of action nodes into a single sequence node.
// lowerVarStmts transforms VarAction (local variable declarations) into
// LocalAction with proper scoping. Matches Python lower_var_stmts (ivy_parser.py:2670-2697).
func lowerVarStmts(stmts []ast.Node) []ast.Node {
	return ast.LowerVarStatements(stmts)
}

// lalrMakeSequence matches Python p_sequence_lcb_actseq_rcb (ivy_parser.py:2705-2713).
// Python: stmts = lower_var_stmts(p[2]); if len(stmts)==1: return stmts[0]
//         else: Sequence(*lower_var_stmts(stmts))
// So lower_var_stmts is called ONCE always, and a SECOND time only if len > 1.
func lalrMakeSequence(cfg *ast.AstConfig, stmts []ast.Node) ast.Node {
	if len(stmts) == 0 {
		return cfg.NewSequence()
	}
	if len(stmts) == 1 {
		return stmts[0]
	}
	stmts = lowerVarStmts(stmts)
	return cfg.NewSequence(stmts...)
}
