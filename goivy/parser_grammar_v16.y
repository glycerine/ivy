// grammar_v16.y — goyacc LALR(1) grammar for FULL Ivy file parsing, version 1.6.
// Mechanically translated from Python PLY grammar in ivy_parser.py and ivy_logic_parser.py.
//
// This is the FULL grammar: it parses complete Ivy files (top-level declarations),
// not just formulas/terms like lalr_logicparser.
//
// Precedence table (copied exactly from Python ivy_parser.py, v1.6):
//   SEMI < GLOBALLY/EVENTUALLY < ARROW/IFF < OR < AND < TILDA
//   < EQ/LE/LT/GE/GT/PTO < TILDAEQ < IF < ELSE < COLON
//   < PLUS/MINUS < TIMES/DIV < DOLLAR < OLD < DOT

%{
package goivy

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// acfg extracts the *AstConfig from the lexer for use in grammar actions.
func parser16Acfg(lex parser16Lexer) *AstConfig {
	return lex.(*parser16LexAdapter).astCfg
}

// parentObject is no longer a global. It is passed explicitly to newIvyAccum().
// See newIvyAccum(parent, parentObjName) in ivy_module.go.

// parser16GetLineno returns a Location for the current token position.
// Matches Python's get_lineno(p, n) → Location(filename, p.lineno(n)).
func parser16GetLineno(lex *parser16LexAdapter) Location {
	//xtracer.Trace("parser.get_lineno ENTER")
	return Location{
		//Filename: parser16NormalizeFilename(lex.filename),
                // xtracer.NormalizeLine handles IVY_EXAMPLES too:
                Filename: xtracer.NormalizeLine(lex.filename),
		Line:     lex.prevTok.Line,
	}
}

// parser16NormalizeFilename replaces the include directory path with <IVY_INCLUDE>
// so that canonical strings match between Go and Python regardless of install location.
// Python: stores the raw path; the golden test normalizes for display.
// For hashing, both sides must agree, so we normalize here.
func parser16NormalizeFilename(f string) string {
	stdDir := NewIvyUtilsConfig().GetStdIncludeDir()
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

// parser16NewLabel generates a unique label with the given prefix, matching Python newlabel().
func parser16NewLabel(cfg *AstConfig, pref string) *Atom {
	xtracer.Trace("parser.newlabel ENTER")
	cfg.LabelCounter++
	return cfg.NewAtom(fmt.Sprintf("%s%d", pref, cfg.LabelCounter))
}

// addLabel adds a label to a LabeledFormula if it doesn't have one.
// Matches Python addlabel() (ivy_parser.py:443-448).
func parser16AddLabel(cfg *AstConfig, lf *LabeledFormula, pref string) *LabeledFormula {
	xtracer.Trace("parser.addlabel ENTER")
	return lf
}

// mkLF wraps a node in a LabeledFormula with no label.
// Matches Python mk_lf (ivy_parser.py:1187).
func parser16MkLF(cfg *AstConfig, x Node) *LabeledFormula {
	xtracer.Trace("parser.mk_lf ENTER")
	lf := cfg.NewLabeledFormula(nil, x)
	if x != nil && x.GetLineno().Line > 0 {
		lf.SetLineno(x.GetLineno())
	}
	return lf
}

// parser16CheckNonTemporal validates that a formula doesn't contain temporal operators.
// Matches Python check_non_temporal() (ivy_parser.py:231-248).
// For LabeledFormula, recursively checks the formula part.
// Reports an error if temporal operators are found.
func parser16CheckNonTemporal(x Node) Node {
	xtracer.Trace("parser.check_non_temporal ENTER")
	if lf, ok := x.(*LabeledFormula); ok {
		parser16CheckNonTemporal(lf.Formula)
		return x
	}
	if HasTemporal(x) {
		// TODO: report_error(IvyError(x, "non-temporal formula expected"))
		fmt.Printf("warning: non-temporal formula expected\n")
	}
	return x
}

// parser16AddUnprovable marks a LabeledFormula as unprovable if cond is non-nil.
// Matches Python addunprovable() (ivy_parser.py:453-457).
func parser16AddUnprovable(lf *LabeledFormula, cond Node) *LabeledFormula {
	xtracer.Trace("parser.addunprovable ENTER")
	if cond != nil {
		lf.Unprovable = true
	}
	return lf
}

// parser16AddTemporal marks a LabeledFormula as temporal.
// Matches Python addtemporal() (ivy_parser.py:438-441).
func parser16AddTemporal(lf *LabeledFormula) *LabeledFormula {
	xtracer.Trace("parser.addtemporal ENTER")
	t := true
	lf.Temporal = &t
	return lf
}

// parser16AddExplicit marks a LabeledFormula as explicit.
// Matches Python addexplicit() (ivy_parser.py:459-462).
func parser16AddExplicit(lf *LabeledFormula) *LabeledFormula {
	xtracer.Trace("parser.addexplicit ENTER")
	lf.Explicit = true
	return lf
}

// parser16NodeLineno extracts the Location from a Node's GetLineno().
// Used when we need the line of a child node instead of lastTok (which is the lookahead).
// Emits the get_lineno trace to match Python's get_lineno(p, n) call.
func parser16NodeLineno(n Node) Location {
	//xtracer.Trace("parser.get_lineno ENTER")
	if n == nil {
		return Location{}
	}
	loc := n.GetLineno()
	loc.Filename = xtracer.NormalizeLine(loc.Filename)
	return loc
}

// atypeToString extracts the string sort name from an atype Node.
// Python atype returns a plain string; Go atype returns *Symbol or *This.
func parser16AtypeToString(n Node) string {
	switch v := n.(type) {
	case *Symbol:
		return v.Rep
	case *This:
		return "this"
	default:
		return fmt.Sprint(n)
	}
}

// parser16AtypeToAtom converts an atype Node (Symbol or This) into an Atom,
// matching Python's Atom(p[n]) where atype returns a string.
// Python atype returns a string; Go atype returns *Symbol or *This.
func parser16AtypeToAtom(cfg *AstConfig, n Node) *Atom {
	switch v := n.(type) {
	case *Symbol:
		return cfg.NewAtom(v.Rep)
	case *This:
		return cfg.NewAtom("this")
	default:
		return cfg.NewAtom(fmt.Sprint(n))
	}
}

// parser16MakeMixinName generates a unique mixin name.
// Matches Python make_mixin_name() (ivy_parser.py:2556-2562).
func parser16MakeMixinName(cfg *AstConfig, atom *Atom, suffix string) *Atom {
	xtracer.Trace("parser.make_mixin_name ENTER")
	// Python v1.6: name = atom.rep.replace(ivy_compose_character, '_') + '[' + suffix + ']'
	rep := strings.ReplaceAll(atom.Rep, ".", "_")
	return cfg.NewAtom(fmt.Sprintf("%s[%s]", rep, suffix))
}

// parser16HandleMixin declares a mixin (before/after/implement).
// Matches Python handle_mixin() (ivy_parser.py:2083-2090).
func parser16HandleMixin(cfg *AstConfig, kind string, mixer *Atom, mixee *Atom, ivy *ivyAccum) {
	xtracer.Trace("parser.handle_mixin ENTER")
	var m Node
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

// parser16HandleBeforeAfter processes before/after action declarations.
// Matches Python handle_before_after() (ivy_parser.py:2105-2115).
// parser16InferActionParams looks up the matching action definition and infers
// missing formal parameters and returns.
// Matches Python infer_action_params() (ivy_parser.py:2093-2103).
// parser16StackActionLookup searches the accumulator stack for an action definition.
// Matches Python stack_action_lookup() (ivy_parser.py:125-133).
func parser16StackActionLookup(ivy *ivyAccum, name string) (Node, int) {
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

func parser16InferActionParams(ivy *ivyAccum, actname string, formals []Node, returns []Node) ([]Node, []Node) {
	xtracer.Trace("parser.infer_action_params ENTER")
	mixee, numParams := parser16StackActionLookup(ivy, actname)
	if mixee == nil {
		return formals, returns
	}
	// Python: if ("common" in mixee.attributes) != ("common" in stack[-1].attributes):
	//             return formals, returns
	mixeeDef, ok := mixee.(*ActionDef)
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

func parser16HandleBeforeAfter(cfg *AstConfig, kind string, atom *Atom, action Node, ivy *ivyAccum, optargs []Node, optreturns []Node) {
	xtracer.Trace("parser.handle_before_after ENTER")
	mixer := parser16MakeMixinName(cfg, atom, kind)
	optargs, optreturns = parser16InferActionParams(ivy, atom.Rep, optargs, optreturns)
	df := cfg.NewActionDef(mixer, action, optargs, optreturns)
	df.SetLineno(atom.GetLineno())
	decl := cfg.NewActionDecl(df)
	ivy.declare(decl)
	parser16HandleMixin(cfg, kind, mixer, atom, ivy)
}

// parser16CreateObject processes an object declaration by expanding its body
// with prefix substitution via instMod.
// Matches Python create_object() (ivy_parser.py:678-692).
// setObjectDefined is now in inst_mod.go with proper implementation.

// parser16ParseNativequote parses a native code block, splitting on backtick-delimited references.
// Matches Python parse_nativequote() (ivy_parser.py:2457-2470).
func parser16ParseNativequote(cfg *AstConfig, raw string, lex *parser16LexAdapter) (string, []Node) {
	xtracer.Trace("parser.parse_nativequote ENTER")
	// Drop the <<< and >>> quotation marks
	s := raw
	if len(s) >= 6 && strings.HasPrefix(s, "<<<") && strings.HasSuffix(s, ">>>") {
		s = s[3 : len(s)-3]
	}
	fields := strings.Split(s, "`")
	var bqs []Node
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
	_ = parser16GetLineno(lex)
	return text, bqs
}

// parser16Methcall matches Python parser16Methcall(lhs, rhs) at ivy_parser.py:3034-3038.
// If lhs is an App or Atom with no args, compose; otherwise create MethodCall.
func parser16Methcall(cfg *AstConfig, lhs, rhs Node) Node {
	xtracer.Trace("parser.parser16Methcall ENTER")
	switch l := lhs.(type) {
	case *App:
		if len(l.Terms) == 0 {
			return ComposeAtomsGeneric(l, rhs)
		}
	case *Atom:
		if len(l.Terms) == 0 {
			return ComposeAtomsGeneric(l, rhs)
		}
	}
	return cfg.NewMethodCall(lhs, rhs)
}

// parser16FixIfPart handles the `some` condition case in if/while actions.
// Matches Python fix_if_part() (ivy_parser.py:2932-2938).
func parser16FixIfPart(cond Node, part Node) Node {
	xtracer.Trace("parser.fix_if_part ENTER")
	// Python: isinstance(cond, Some) — matches Some, SomeMin, SomeMax (all subclasses).
	// Extract params from whichever type matches.
	var params []Node
	switch s := cond.(type) {
	case *Some:
		params = s.Params
	case *SomeMin:
		params = s.Params
	case *SomeMax:
		params = s.Params
	}
	if params != nil {
		subst := make(map[string]string)
		for _, p := range params {
			// Python: subst = dict((x.rep[4:],x.rep) for x in args)
			// Params can be App, Atom, Variable, etc. — use NodeRep generically.
			rep := NodeRep(p)
			if len(rep) > 4 {
				subst[rep[4:]] = rep
			}
		}
		if len(subst) > 0 {
			part = SubstPrefixAtomsAst(part, subst, nil, nil, nil)
		}
	}
	return part
}

// parser16CreateObject processes an object declaration by expanding its body
// with prefix substitution via instMod.
// Matches Python create_object() (ivy_parser.py:678-693) EXACTLY.
func parser16CreateObject(cfg *AstConfig, top *ivyAccum, name *Atom, objectargs []Node, module *ivyAccum, lineno Location, continuation bool) {
	xtracer.Trace("parser.create_object ENTER name=%s", name.Rep)

	// Python line 680: prefargs = [Variable('V'+str(idx),pr.sort) for idx,pr in enumerate(objectargs)]
	var prefargs []Node
	for idx, pr := range objectargs {
		vname := fmt.Sprintf("V%d", idx)
		var sort string
		if a, ok := pr.(*Atom); ok && a.ASort != nil {
			if sym, ok := a.ASort.(*Symbol); ok {
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
	vsubst := make(map[string]*Variable)
	for i, pr := range objectargs {
		if i < len(prefargs) {
			prName := nodeRep(pr)
			if v, ok := prefargs[i].(*Variable); ok {
				vsubst[prName] = v
			}
		}
	}

	// Python line 688: inst_mod(top, module, pref, {}, vsubst)
	instMod(top, module, pref, map[string]string{}, vsubst, "", lineno)

	xtracer.Trace("parser.create_object EXIT name=%s", name.Rep)
}

// parser16TokLineno creates a Location from a TokenInfo, using the normalized filename
// from the lex adapter. This replaces parser16GetLineno for cases where we have direct
// access to the token's position.
func parser16TokLineno(lex *parser16LexAdapter, tok TokenInfo) Location {
        //xtracer.Trace("parser.get_lineno ENTER")
	return Location{
		Filename: xtracer.NormalizeLine(lex.filename),
		Line:     tok.Line,
	}
}

%}

// The union type for semantic values.
%union {
	node     Node
	nodes    []Node
	str      string
	tok      TokenInfo
	bval     bool
	accum    *ivyAccum
}

// Terminal tokens — identifiers and literals (carry string value + line)
%token <tok>  PARSER16_TOK_PRESYMBOL PARSER16_TOK_VARIABLE
%token <tok>  PARSER16_TOK_LABEL   // Python returns LABEL from lexer; Go handles via labelname rule instead
%token <tok>  PARSER16_TOK_NATIVEQUOTE

// Terminal tokens — punctuation (carry line number only)
%token <tok>  PARSER16_TOK_LPAREN PARSER16_TOK_RPAREN PARSER16_TOK_LB PARSER16_TOK_RB PARSER16_TOK_LCB PARSER16_TOK_RCB
%token <tok>  PARSER16_TOK_COMMA PARSER16_TOK_SEMI PARSER16_TOK_COLON PARSER16_TOK_DOT
%token <tok>  PARSER16_TOK_DOTS PARSER16_TOK_DOTDOTDOT

// Terminal tokens — operators (carry line number only)
%token <tok>  PARSER16_TOK_PLUS PARSER16_TOK_MINUS PARSER16_TOK_TIMES PARSER16_TOK_DIV
%token <tok>  PARSER16_TOK_EQ PARSER16_TOK_TILDAEQ PARSER16_TOK_TILDA PARSER16_TOK_LE PARSER16_TOK_LT PARSER16_TOK_GE PARSER16_TOK_GT
%token <tok>  PARSER16_TOK_AND PARSER16_TOK_OR PARSER16_TOK_ARROW PARSER16_TOK_IFF
%token <tok>  PARSER16_TOK_PTO PARSER16_TOK_DOLLAR PARSER16_TOK_CARET
%token <tok>  PARSER16_TOK_ASSIGN

// Terminal tokens — keywords (logic, carry line number only)
%token <tok>  PARSER16_TOK_FORALL PARSER16_TOK_EXISTS
%token <tok>  PARSER16_TOK_TRUE PARSER16_TOK_FALSE
%token <tok>  PARSER16_TOK_OLD PARSER16_TOK_THIS PARSER16_TOK_ISA
%token <tok>  PARSER16_TOK_IF PARSER16_TOK_ELSE
%token <tok>  PARSER16_TOK_GLOBALLY PARSER16_TOK_EVENTUALLY
%token <tok>  PARSER16_TOK_WHENNEXT PARSER16_TOK_WHENPREV PARSER16_TOK_WHENFIRST PARSER16_TOK_WHENLAST

// Terminal tokens — keywords (actions, carry line number only)
%token <tok>  PARSER16_TOK_ASSUME PARSER16_TOK_ASSERT PARSER16_TOK_REQUIRE PARSER16_TOK_ENSURE
%token <tok>  PARSER16_TOK_VAR PARSER16_TOK_LOCAL PARSER16_TOK_LET PARSER16_TOK_CALL
%token <tok>  PARSER16_TOK_WHILE PARSER16_TOK_FOR PARSER16_TOK_IN PARSER16_TOK_INVARIANT PARSER16_TOK_DECREASES
%token <tok>  PARSER16_TOK_RETURNS
%token <tok>  PARSER16_TOK_SOME PARSER16_TOK_MINIMIZING PARSER16_TOK_MAXIMIZING
%token <tok>  PARSER16_TOK_DEBUG PARSER16_TOK_THUNK PARSER16_TOK_UNPROVABLE PARSER16_TOK_PROOF
%token <tok>  PARSER16_TOK_INSTANTIATE

// Terminal tokens — keywords (declarations, carry line number only)
%token <tok>  PARSER16_TOK_RELATION PARSER16_TOK_INDIV PARSER16_TOK_FUNCTION PARSER16_TOK_DERIVED
%token <tok>  PARSER16_TOK_AXIOM PARSER16_TOK_CONJECTURE PARSER16_TOK_SCHEMA PARSER16_TOK_THEOREM
%token <tok>  PARSER16_TOK_PROPERTY PARSER16_TOK_DEFINITION
%token <tok>  PARSER16_TOK_TYPE PARSER16_TOK_STRUCT
%token <tok>  PARSER16_TOK_MODULE PARSER16_TOK_OBJECT PARSER16_TOK_CLASS PARSER16_TOK_SUBCLASS
%token <tok>  PARSER16_TOK_ACTION PARSER16_TOK_METHOD
%token <tok>  PARSER16_TOK_BEFORE PARSER16_TOK_AFTER PARSER16_TOK_AROUND PARSER16_TOK_MIXIN PARSER16_TOK_IMPLEMENT
%token <tok>  PARSER16_TOK_ISOLATE PARSER16_TOK_EXTRACT PARSER16_TOK_TRUSTED
%token <tok>  PARSER16_TOK_EXPORT PARSER16_TOK_IMPORT PARSER16_TOK_DELEGATE PARSER16_TOK_USING PARSER16_TOK_INCLUDE
%token <tok>  PARSER16_TOK_INTERPRET PARSER16_TOK_MACRO PARSER16_TOK_ALIAS PARSER16_TOK_ATTRIBUTE
%token <tok>  PARSER16_TOK_VARIANT PARSER16_TOK_OF
%token <tok>  PARSER16_TOK_SCENARIO
%token <tok>  PARSER16_TOK_PROGRESS PARSER16_TOK_RELY PARSER16_TOK_MIXORD
%token <tok>  PARSER16_TOK_CONCEPT PARSER16_TOK_STATE PARSER16_TOK_UPDATE PARSER16_TOK_FROM
%token <tok>  PARSER16_TOK_PARAMS PARSER16_TOK_MODIFIES PARSER16_TOK_ENSURES PARSER16_TOK_REQUIRES
%token <tok>  PARSER16_TOK_INIT PARSER16_TOK_ENTRY PARSER16_TOK_SET PARSER16_TOK_NULL PARSER16_TOK_MATCH
%token <tok>  PARSER16_TOK_FRESH PARSER16_TOK_NAMED
%token        PARSER16_TOK_TEMPORAL PARSER16_TOK_EXPLICIT
%token <tok>  PARSER16_TOK_SPECIFICATION PARSER16_TOK_IMPLEMENTATION PARSER16_TOK_PRIVATE
%token <tok>  PARSER16_TOK_GLOBAL PARSER16_TOK_COMMON
%token <tok>  PARSER16_TOK_GHOST PARSER16_TOK_FINITE
%token <tok>  PARSER16_TOK_PARAMETER
%token <tok>  PARSER16_TOK_DESTRUCTOR PARSER16_TOK_CONSTRUCTOR PARSER16_TOK_FIELD
%token <tok>  PARSER16_TOK_AUTOINSTANCE
// PARSER16_TOK_VAR_KW removed (was unused placeholder)

// Proof/tactic tokens
%token <tok>  PARSER16_TOK_TACTIC PARSER16_TOK_TRIGGER
%token <tok>  PARSER16_TOK_SHOWGOALS PARSER16_TOK_DEFERGOAL PARSER16_TOK_SPOIL
%token <tok>  PARSER16_TOK_UNFOLD PARSER16_TOK_FORGET
%token <tok>  PARSER16_TOK_APPLY
%token <tok>  PARSER16_TOK_WITH
// PARSER16_TOK_METHOD_KW, PARSER16_TOK_NULL_KW, PARSER16_TOK_SET_KW removed (were unused placeholders)

// Nonterminal types — formula/term
%type <node>  term fmla aterm var simplevar atype
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

// Precedence declarations — copied exactly from Python v1.6 precedence table.
%left         PARSER16_TOK_SEMI
%left         PARSER16_TOK_GLOBALLY PARSER16_TOK_EVENTUALLY PARSER16_TOK_WHENFIRST PARSER16_TOK_WHENLAST PARSER16_TOK_WHENNEXT PARSER16_TOK_WHENPREV
%left         PARSER16_TOK_IF
%left         PARSER16_TOK_ELSE
%left         PARSER16_TOK_OR
%left         PARSER16_TOK_AND
%left         PARSER16_TOK_TILDA
%left         PARSER16_TOK_EQ PARSER16_TOK_LE PARSER16_TOK_LT PARSER16_TOK_GE PARSER16_TOK_GT PARSER16_TOK_PTO
%left         PARSER16_TOK_TILDAEQ
%left         PARSER16_TOK_COLON
%left         PARSER16_TOK_PLUS
%left         PARSER16_TOK_MINUS
%left         PARSER16_TOK_TIMES
%left         PARSER16_TOK_DIV
%left         PARSER16_TOK_DOLLAR
// Note: Python has no ASSIGN or ISA in precedence, but goyacc needs them
// to resolve shift/reduce conflicts in action rules. These are harmless
// since ASSIGN/ISA never appear in ambiguous expression positions.
%right        PARSER16_TOK_ASSIGN

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
        lex := parser16lex.(*parser16LexAdapter)
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
    | top PARSER16_TOK_USING SYMBOLx
    {
        xtracer.Trace("parser.p_top_using_symbol ENTER (top)")
        $$ = $1
        // Python: importer(p[3]) and merge decls — deferred to post-parse
    }
    | top PARSER16_TOK_INCLUDE SYMBOLx
    {
        xtracer.Trace("parser.p_top_include_symbol ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
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
            pref := parser16Acfg(parser16lex).NewAtom(name)
            pref.SetLineno(parser16TokLineno(lex, $2))
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
    // --- Axiom (v1.6): top optexplicit opttemporal AXIOM lgprop ---
    | top optexplicit opttemporal PARSER16_TOK_AXIOM lgprop
    {
        xtracer.Trace("parser.p_top_axiom_optlabel_gprop ENTER (top)")
        $$ = $1
        lf := parser16AddLabel(parser16Acfg(parser16lex), $5.(*LabeledFormula), "axiom")
        // Python: lf = addexplicit(lf) if p[2] else lf  (explicit first)
        if $2 != nil {
            lf = parser16AddExplicit(lf)
        }
        // Python: d = AxiomDecl(addtemporal(lf) if p[3] else check_non_temporal(lf))
        if $3 != nil {
            lf = parser16AddTemporal(lf)
        } else {
            parser16CheckNonTemporal(lf)
        }
        d := parser16Acfg(parser16lex).NewAxiomDecl(lf)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $4))
        $$.declare(d)
    }
    // --- Property (v1.6): top optexplicit opttemporal PROPERTY labeledfmla optskolem optproof ---
    | top optexplicit opttemporal PARSER16_TOK_PROPERTY labeledfmla optskolem optproof
    {
        xtracer.Trace("parser.p_top_property_labeledfmla ENTER (top)")
        $$ = $1
        lf := parser16AddLabel(parser16Acfg(parser16lex), $5.(*LabeledFormula), "prop")
        // Python: lf = addtemporal(lf) if p[3] else check_non_temporal(lf)
        if $3 != nil {
            lf = parser16AddTemporal(lf)
        } else {
            parser16CheckNonTemporal(lf)
        }
        // Python: lf = addexplicit(lf) if p[2] else lf
        if $2 != nil {
            lf = parser16AddExplicit(lf)
        }
        d := parser16Acfg(parser16lex).NewPropertyDecl(lf)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $4))
        $$.declare(d)
        if $6 != nil {
            $$.declare(parser16Acfg(parser16lex).NewNamedDecl($6))
        }
        if $7 != nil {
            $$.declare(parser16Acfg(parser16lex).NewProofDecl($7))
        }
    }
    // --- Conjecture ---
    | top PARSER16_TOK_CONJECTURE labeledfmla
    {
        xtracer.Trace("parser.p_top_conjecture_labeledfmla ENTER (top)")
        $$ = $1
        lf := parser16AddLabel(parser16Acfg(parser16lex), $3.(*LabeledFormula), "conj")
        d := parser16Acfg(parser16lex).NewConjectureDecl(lf)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(d)
    }
    // --- Invariant (v1.6): top optexplicit INVARIANT labeledfmla optproof ---
    | top optexplicit PARSER16_TOK_INVARIANT labeledfmla optproof
    {
        xtracer.Trace("parser.p_top_invariant_labeledfmla ENTER (top)")
        $$ = $1
        lf := parser16AddLabel(parser16Acfg(parser16lex), $4.(*LabeledFormula), "invar")
        lf.Unprovable = false
        if $2 != nil {
            lf.Explicit = true
        }
        d := parser16Acfg(parser16lex).NewConjectureDecl(lf)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        $$.declare(d)
        if $5 != nil {
            $$.declare(parser16Acfg(parser16lex).NewProofDecl($5))
        }
    }
    // --- Unprovable Invariant ---
    | top PARSER16_TOK_UNPROVABLE PARSER16_TOK_INVARIANT labeledfmla optproof
    {
        xtracer.Trace("parser.p_top_unprovable_invariant_labeledfmla ENTER (top)")
        $$ = $1
        lf := parser16AddLabel(parser16Acfg(parser16lex), $4.(*LabeledFormula), "invar")
        lf.Unprovable = true
        lf.Explicit = true
        d := parser16Acfg(parser16lex).NewConjectureDecl(lf)
        // Python: if not lf.unprovable or check_unprovable.get(): p[0].declare(d); declare(ProofDecl)
        if !lf.Unprovable || parser16Acfg(parser16lex).CheckUnprovable {
            $$.declare(d)
            if $5 != nil {
                $$.declare(parser16Acfg(parser16lex).NewProofDecl($5))
            }
        }
    }
    // --- Module: top MODULE modulestart modcat atom optwith EQ LCB top RCB moduleend ---
    | top PARSER16_TOK_MODULE modulestart modcat atom optwith PARSER16_TOK_EQ PARSER16_TOK_LCB top PARSER16_TOK_RCB moduleend
    {
        xtracer.Trace("parser.p_top_module_atom_eq_lcb_top_rcb ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        modAccum := $9
        // Store ivyAccum directly as module body, matching Python where p[9] (Ivy instance)
        // is stored in Definition(name, ivy_instance). This preserves .objects, .defined, .static.
        d := parser16Acfg(parser16lex).NewDefinition(AppToAtom($5), modAccum)
        $$.declare(parser16Acfg(parser16lex).NewModuleDecl(d))
        // Python: if p[4] == "isolate": ... with get_lineno(p,2) on this, iso, d.args[0], d
        if $4 != nil {
            if catAtom, ok := $4.(*Atom); ok && catAtom.Rep == "isolate" {
                thisAtom := parser16Acfg(parser16lex).NewAtom("this")
                thisAtom.SetLineno(parser16TokLineno(lex, $2))
                iso := parser16Acfg(parser16lex).NewAtom("iso")
                iso.SetLineno(parser16TokLineno(lex, $2))
                isoElems := append([]Node{iso, thisAtom}, $6...)
                isoDef := parser16Acfg(parser16lex).NewIsolateDef(isoElems, len($6))
                isoDef.SetLineno(parser16TokLineno(lex, $2))
                isoDecl := parser16Acfg(parser16lex).NewIsolateDecl(isoDef)
                isoDecl.Attributes = []Node{parser16Acfg(parser16lex).NewAtom("common")}
                isoDecl.SetLineno(parser16TokLineno(lex, $2))
                modAccum.declare(isoDecl)
            }
        }
        // Python: stack.pop() — restore scope after processing module body
        lex.accum = $$
        $$.isModule = false // Python: stack[-1].is_module = False
    }
    // --- Object ---
    | top PARSER16_TOK_OBJECT objsym objectargs PARSER16_TOK_EQ PARSER16_TOK_LCB optdotdotdot top PARSER16_TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top_object_symbol_eq_lcb_top_rcb ENTER (top)")
        $$ = $1
        objAccum := $8
        pref := $3.(*Atom)
        lineno := parser16NodeLineno($3)
        parser16CreateObject(parser16Acfg(parser16lex), $$, pref, $4, objAccum, lineno, $7)
        // Python: create_object does stack.pop() at ivy_parser.py:692
        parser16lex.(*parser16LexAdapter).accum = $$
    }
    // --- Class ---
    | top PARSER16_TOK_CLASS objsym objectargs PARSER16_TOK_EQ PARSER16_TOK_LCB optdotdotdot top PARSER16_TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top__top_class_symbol_objectargs_eq_lcb_optdo ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        objAccum := $8

        // Python: scnst = Atom(This())
        // Python: scnst.lineno = get_lineno(p,2)
        scnst := parser16Acfg(parser16lex).NewAtom("this")
        scnst.SetLineno(parser16TokLineno(lex, $2))

        // Python: tdfn = TypeDef(scnst, UninterpretedSort())
        // Python: tdfn.lineno = get_lineno(p,2)
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, parser16Acfg(parser16lex).NewUninterpretedSortAST())
        tdfn.SetLineno(parser16TokLineno(lex, $2))

        // Python: p[8].declare(TypeDecl(tdfn))
        objAccum.declare(parser16Acfg(parser16lex).NewTypeDecl(tdfn))

        // Python: p[8].decls = [p[8].decls[-1]] + p[8].decls[:-1]
        if n := len(objAccum.decls); n > 1 {
            last := objAccum.decls[n-1]
            copy(objAccum.decls[1:], objAccum.decls[:n-1])
            objAccum.decls[0] = last
        }

        // Python: create_object(p[0], p[3], p[4], p[8], get_lineno(p,3), p[7])
        parser16CreateObject(parser16Acfg(parser16lex), $$, $3.(*Atom), $4, objAccum, parser16NodeLineno($3), $7)

        // Python: stack.pop() equivalent
        lex.accum = $$
    }
    // --- Subclass ---
    | top PARSER16_TOK_SUBCLASS objsym PARSER16_TOK_OF atype PARSER16_TOK_EQ PARSER16_TOK_LCB optdotdotdot top PARSER16_TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top__top_subclass_symbol_of_atype_eq_lcb_optd ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        objAccum := $9

        // Python: scnst = Atom(This())
        // Python: scnst.lineno = get_lineno(p,2)
        scnst := parser16Acfg(parser16lex).NewAtom("this")
        scnst.SetLineno(parser16TokLineno(lex, $2))

        // Python: tdfn = TypeDef(scnst, UninterpretedSort())
        // Python: tdfn.lineno = get_lineno(p,2)
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, parser16Acfg(parser16lex).NewUninterpretedSortAST())
        tdfn.SetLineno(parser16TokLineno(lex, $2))

        // Python: p[9].declare(TypeDecl(tdfn))
        objAccum.declare(parser16Acfg(parser16lex).NewTypeDecl(tdfn))

        // Python: vdfn = VariantDef(scnst, Atom(p[5]))
        // Python: p[9].declare(VariantDecl(vdfn))
        vdfn := parser16Acfg(parser16lex).NewVariantDef(scnst, parser16Acfg(parser16lex).NewAtom(NodeRep($5)))
        objAccum.declare(parser16Acfg(parser16lex).NewVariantDecl(vdfn))

        // Python: p[9].decls = p[9].decls[-2:] + p[9].decls[:-2]
        if n := len(objAccum.decls); n > 2 {
            rotated := make([]Node, n)
            copy(rotated, objAccum.decls[n-2:])
            copy(rotated[2:], objAccum.decls[:n-2])
            objAccum.decls = rotated
        }

        // Python: create_object(p[0], p[3], [], p[9], get_lineno(p,3), p[8])
        parser16CreateObject(parser16Acfg(parser16lex), $$, $3.(*Atom), []Node{}, objAccum, parser16NodeLineno($3), $8)

        // Python: stack.pop() equivalent
        lex.accum = $$
    }
    // --- Definition (v1.6): top optexplicit DEFINITION optlabel gdefn optproof ---
    | top optexplicit PARSER16_TOK_DEFINITION optlabel gdefn optproof
    {
        xtracer.Trace("parser.p_top_definition_optlabel_gdefn_optproof ENTER (top)")
        $$ = $1
        // Python: foo = p[5]
        // Python: if p[2]: foo = DefinitionSchema(*foo.args); foo.lineno = p[5].lineno
        gdefn := $5
        if $2 != nil { // optexplicit is True
            if def, ok := gdefn.(*Definition); ok {
                ds := parser16Acfg(parser16lex).NewDefinitionSchema(*def)
                ds.SetLineno(def.GetLineno())
                gdefn = ds
            }
        }
        lf := parser16Acfg(parser16lex).NewLabeledFormula($4, gdefn)
        lf.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        lf = parser16AddLabel(parser16Acfg(parser16lex), lf, "def")
        dd := parser16Acfg(parser16lex).NewDefinitionDecl(lf)
        $$.declare(dd)
        if $6 != nil {
            $$.declare(parser16Acfg(parser16lex).NewProofDecl($6))
        }
    }
    // --- Schema ---
    | top PARSER16_TOK_SCHEMA schdefn
    {
        xtracer.Trace("parser.p_top_schema_defn ENTER (top)")
        $$ = $1
        sch := parser16Acfg(parser16lex).NewSchema($3)
        sd := parser16Acfg(parser16lex).NewSchemaDecl(sch)
        $$.declare(sd)
    }
    // --- Theorem with schdefn ---
    | top PARSER16_TOK_THEOREM schdefn optproof
    {
        xtracer.Trace("parser.p_top_theorem_defn ENTER (top)")
        $$ = $1
        sch := parser16Acfg(parser16lex).NewSchema($3)
        td := parser16Acfg(parser16lex).NewTheoremDecl(sch)
        $$.declare(td)
        if $4 != nil {
            $$.declare(parser16Acfg(parser16lex).NewProofDecl($4))
        }
    }
    // --- Theorem with LABEL schdefnrhs ---
    | top PARSER16_TOK_THEOREM labelname schdefnrhs optproof
    {
        xtracer.Trace("parser.p_top_theorem_label_rhs ENTER (top)")
        $$ = $1
        label := parser16Acfg(parser16lex).NewAtom($3.Val)
        label.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        df := parser16Acfg(parser16lex).NewDefinition(label, $4)
        df.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        sch := parser16Acfg(parser16lex).NewSchema(df)
        td := parser16Acfg(parser16lex).NewTheoremDecl(sch)
        $$.declare(td)
        if $5 != nil {
            $$.declare(parser16Acfg(parser16lex).NewProofDecl($5))
        }
    }
    // --- Proof LABEL proofstep ---
    | top PARSER16_TOK_PROOF labelname proofstep
    {
        xtracer.Trace("parser.p_top_proof_label_label_proofstep ENTER (top)")
        $$ = $1
        label := parser16Acfg(parser16lex).NewAtom($3.Val)
        label.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        lf := parser16Acfg(parser16lex).NewLabeledFormula(label, $4)
        lf.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        $$.declare(parser16Acfg(parser16lex).NewProofDecl(lf))
    }
    // --- Instantiate ---
    | top PARSER16_TOK_INSTANTIATE insts
    {
        xtracer.Trace("parser.p_top_instantiate_insts ENTER (top)")
        $$ = $1
        doInsts($$, $3)
    }
    // --- Autoinstance ---
    | top PARSER16_TOK_AUTOINSTANCE insts
    {
        xtracer.Trace("parser.p_top_autoinstance_insts ENTER (top)")
        $$ = $1
        d := parser16Acfg(parser16lex).NewAutoInstanceDecl($3...)
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
    | top PARSER16_TOK_RELATION rels
    {
        xtracer.Trace("parser.p_top_relation_rels ENTER (top)")
        $$ = $1
        for _, d := range $3 {
            $$.declare(d)
        }
    }
    // --- Function ---
    | top PARSER16_TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_top_function_tapp_colon_atype ENTER (top)")
        $$ = $1
        for _, d := range $3 {
            $$.declare(d)
        }
    }
    // --- Derived ---
    | top PARSER16_TOK_DERIVED defns
    {
        xtracer.Trace("parser.p_top_derived_defns ENTER (top)")
        $$ = $1
        args := make([]Node, len($3))
        for i, x := range $3 {
            args[i] = parser16AddLabel(parser16Acfg(parser16lex), parser16MkLF(parser16Acfg(parser16lex), x), "def")
        }
        dd := parser16Acfg(parser16lex).NewDerivedDecl(args...)
        $$.declare(dd)
    }
    // --- Type (uninterpreted) ---
    | top optfinite optghost PARSER16_TOK_TYPE typesymbol
    {
        xtracer.Trace("parser.p_top_type_symbol ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        scnst := parser16Acfg(parser16lex).NewAtom($5.(*Atom).Rep)
        scnst.SetLineno(parser16NodeLineno($5))
        // Python: tdfn = (GhostTypeDef if p[3] else TypeDef)(scnst, UninterpretedSort())
        var tdfnNode Node
        if $3 { // optghost
            gt := parser16Acfg(parser16lex).NewGhostTypeDef(*parser16Acfg(parser16lex).NewTypeDef(scnst, parser16Acfg(parser16lex).NewUninterpretedSortAST()))
            if $2 { gt.Finite = true }
            gt.SetLineno(parser16TokLineno(lex, $4))
            tdfnNode = gt
        } else {
            tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, parser16Acfg(parser16lex).NewUninterpretedSortAST())
            if $2 { tdfn.Finite = true }
            tdfn.SetLineno(parser16TokLineno(lex, $4))
            tdfnNode = tdfn
        }
        td := parser16Acfg(parser16lex).NewTypeDecl(tdfnNode)
        $$.declare(td)
    }
    // --- Type with sort ---
    | top optfinite optghost PARSER16_TOK_TYPE typesymbol PARSER16_TOK_EQ sort
    {
        xtracer.Trace("parser.p_top_type_symbol_eq_sort ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        scnst := parser16Acfg(parser16lex).NewAtom($5.(*Atom).Rep)
        scnst.SetLineno(parser16NodeLineno($5))

        // Python: defsort = UninterpretedSort() if isinstance(p[7], Range) else p[7]
        sortNode := $7
        _, isRange := sortNode.(*Range)
        if isRange {
            sortNode = parser16Acfg(parser16lex).NewUninterpretedSortAST()
        }

        // Python: tdfn = (GhostTypeDef if p[3] else TypeDef)(scnst, defsort)
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, sortNode)
        if $2 { tdfn.Finite = true }
        tdfn.SetLineno(parser16TokLineno(lex, $6))
        var tdfnNode Node = tdfn
        if $3 {
            tdfnNode = parser16Acfg(parser16lex).NewGhostTypeDef(*tdfn)
        }
        td := parser16Acfg(parser16lex).NewTypeDecl(tdfnNode)
        $$.declare(td)

        // Python: if isinstance(p[7], Range): ...
        if isRange {
            imp := parser16Acfg(parser16lex).NewImplies(scnst, $7)
            imp.SetLineno(parser16TokLineno(lex, $4))
            lf := parser16MkLF(parser16Acfg(parser16lex), imp)
            lf.SetLineno(imp.GetLineno())
            labeled := parser16AddLabel(parser16Acfg(parser16lex), lf, "interp")
            thing := parser16Acfg(parser16lex).NewInterpretDecl(labeled)
            thing.SetLineno(parser16TokLineno(lex, $4))
            $$.declare(thing)
        }
    }
    // --- Progress ---
    | top PARSER16_TOK_PROGRESS defns
    {
        xtracer.Trace("parser.p_top_progress_defns ENTER (top)")
        $$ = $1
        pd := parser16Acfg(parser16lex).NewProgressDecl($3...)
        $$.declare(pd)
    }
    // --- Rely ---
    | top PARSER16_TOK_RELY atom PARSER16_TOK_ARROW atom
    {
        xtracer.Trace("parser.p_top_rely_atom_arrow_atom ENTER (top)")
        $$ = $1
        imp := parser16Acfg(parser16lex).NewImplies($3, $5)
        rd := parser16Acfg(parser16lex).NewRelyDecl(imp)
        $$.declare(rd)
    }
    | top PARSER16_TOK_RELY atom
    {
        xtracer.Trace("parser.p_top_rely_atom ENTER (top)")
        $$ = $1
        rd := parser16Acfg(parser16lex).NewRelyDecl($3)
        $$.declare(rd)
    }
    // --- Mixord ---
    | top PARSER16_TOK_MIXORD callatom PARSER16_TOK_ARROW callatom
    {
        xtracer.Trace("parser.p_top_mixord_callatom_arrow_callatom ENTER (top)")
        $$ = $1
        imp := parser16Acfg(parser16lex).NewImplies($3, $5)
        md := parser16Acfg(parser16lex).NewMixOrdDecl(imp)
        $$.declare(md)
    }
    // --- Concept ---
    | top PARSER16_TOK_CONCEPT cdefns
    {
        xtracer.Trace("parser.p_top_concept_cdefns ENTER (top)")
        $$ = $1
        cd := parser16Acfg(parser16lex).NewConceptDecl($3...)
        $$.declare(cd)
    }
    // --- Update ---
    | top PARSER16_TOK_UPDATE apps PARSER16_TOK_FROM apps upaxes
    {
        xtracer.Trace("parser.p_top_update_terms_from_terms_upaxes ENTER (top)")
        $$ = $1
        cfg := parser16Acfg(parser16lex)
        // Python (ivy_parser.py:1741-1745):
        //   dfns = [x.rep for x in p[3]]
        //   deps = [x.rep for x in p[5]]
        //   p[0].declare(UpdateDecl(PatternBasedUpdate(SymbolList(*dfns),
        //                                              SymbolList(*deps),
        //                                              UpdatePatternList(*p[6]))))
        dfns := cfg.NewSymbolList($3...)
        deps := cfg.NewSymbolList($5...)
        pats := cfg.NewUpdatePatternList($6...)
        pbu := cfg.NewPatternBasedUpdate(dfns, deps, pats)
        upd := cfg.NewUpdateDecl(pbu)
        $$.declare(upd)
    }
    // --- Macro ---
    | top PARSER16_TOK_MACRO atom PARSER16_TOK_EQ sequence
    {
        xtracer.Trace("parser.p_top_macro_atom_eq_lcb_action_rcb ENTER (top)")
        $$ = $1
        d := parser16Acfg(parser16lex).NewDefinition(AppToAtom($3), $5)
        md := parser16Acfg(parser16lex).NewMacroDecl(d)
        $$.declare(md)
    }
    // --- Action (v1.6): top optimpex actmeth SYMBOL optargs optreturns optactiondef ---
    | top optimpex actmeth SYMBOLx optargs optreturns optactiondef
    {
        xtracer.Trace("parser.p_top_optimpex_action_symbol_optargs_optreturns_eq_action ENTER (top)")
        $$ = $1
        // Python: adef = p[7]; if not hasattr(adef,'lineno'): adef.lineno = get_lineno(p,4)
        // Python almost always has lineno set, so get_lineno rarely fires.
        // Match Python by checking HasLoc on the base node.
        adef := $7
        var lineno Location
        if adef != nil {
            if b, ok := adef.(interface{ HasLocSet() bool }); ok && b.HasLocSet() {
                lineno = adef.GetLineno()
            }
        }
        if lineno == (Location{}) {
            lineno = parser16TokLineno(parser16lex.(*parser16LexAdapter), $4)
        }

        // Python: formals = p[5]
        formals := $5

        // Python: if p[3]: (actmeth is True for METHOD)
        if $3 {
            // Python: arg0 = App('self')
            // Python: arg0.sort = This()
            // Python: arg0.lineno = get_lineno(p,4)
            selfArg := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("self", nil))
            selfArg.ASort = parser16Acfg(parser16lex).NewThis()
            selfArg.SetLineno(lineno)
            // Python: formals = [arg0] + formals
            formals = append([]Node{selfArg}, formals...)
        }

        // Python: if isinstance(adef, CrashAction):
        //             adef = adef.clone([Atom(This(), formals)])
        if ca, ok := adef.(*CrashAction); ok {
            thisAtom := parser16Acfg(parser16lex).NewAtom("this", formals...)
            thisAtom.SetLineno(lineno)
            adef = ca.Clone([]Node{thisAtom})
        }

        theAtom := parser16Acfg(parser16lex).NewAtom($4.Val)
        theAtom.SetLineno(lineno)
        // Python: ActionDef(theAtom, adef, formals=formals, returns=p[6])
        actdef := parser16Acfg(parser16lex).NewActionDef(theAtom, adef, formals, $6)
        actdef.SetLineno(lineno)
        decl := parser16Acfg(parser16lex).NewActionDecl(actdef)
        decl.SetLineno(lineno)
        $$.declare(decl)
        // Python: if p[2]: declare ExportDecl/ImportDecl for this action name.
        if $2 != nil {
            switch $2.(type) {
            case *ExportDecl:
                d := parser16Acfg(parser16lex).NewExportDecl(
                    parser16Acfg(parser16lex).NewExportDef(
                        parser16Acfg(parser16lex).NewAtom($4.Val),
                        parser16Acfg(parser16lex).NewAtom(""),
                    ),
                )
                d.SetLineno(lineno)
                $$.declare(d)
            case *ImportDecl:
                d := parser16Acfg(parser16lex).NewImportDecl(
                    parser16Acfg(parser16lex).NewImportDef(
                        parser16Acfg(parser16lex).NewAtom($4.Val),
                        parser16Acfg(parser16lex).NewAtom(""),
                    ),
                )
                d.SetLineno(lineno)
                $$.declare(d)
            }
        }
    }
    // --- Mixin before ---
    | top PARSER16_TOK_MIXIN callatom PARSER16_TOK_BEFORE callatom
    {
        xtracer.Trace("parser.p_top_mixin_callatom_before_callatom ENTER (top)")
        $$ = $1
        m := parser16Acfg(parser16lex).NewMixinBeforeDef($3, $5)
        md := parser16Acfg(parser16lex).NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Mixin after ---
    | top PARSER16_TOK_MIXIN callatom PARSER16_TOK_AFTER callatom
    {
        xtracer.Trace("parser.p_top_mixin_callatom_after_callatom ENTER (top)")
        $$ = $1
        m := parser16Acfg(parser16lex).NewMixinAfterDef($3, $5)
        md := parser16Acfg(parser16lex).NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Before ---
    | top PARSER16_TOK_BEFORE atype optargs optreturns sequence
    {
        xtracer.Trace("parser.p_top_before_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := parser16Acfg(parser16lex).NewAtom($3.(*Symbol).Rep)
        atom.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        parser16HandleBeforeAfter(parser16Acfg(parser16lex), "before", atom, $6, $$, $4, $5)
    }
    // --- After ---
    | top PARSER16_TOK_AFTER atype optargs optreturns topseq
    {
        xtracer.Trace("parser.p_top_after_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := parser16Acfg(parser16lex).NewAtom($3.(*Symbol).Rep)
        atom.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        parser16HandleBeforeAfter(parser16Acfg(parser16lex), "after", atom, $6, $$, $4, $5)
    }
    // --- Around ---
    | top PARSER16_TOK_AROUND atype optargs optreturns PARSER16_TOK_LCB actseq optsemi PARSER16_TOK_DOTDOTDOT actseq optsemi PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_top_around_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := parser16Acfg(parser16lex).NewAtom($3.(*Symbol).Rep)
        aroundLoc := parser16TokLineno(parser16lex.(*parser16LexAdapter), $2)
        atom.SetLineno(aroundLoc)
        before := parser16StmtToSeq(parser16Acfg(parser16lex), $7)
        after := parser16StmtToSeq(parser16Acfg(parser16lex), $10)
        parser16HandleBeforeAfter(parser16Acfg(parser16lex), "before", atom, before, $$, $4, $5)
        parser16HandleBeforeAfter(parser16Acfg(parser16lex), "after", atom, after, $$, $4, $5)
    }
    // --- After init ---
    | top PARSER16_TOK_AFTER PARSER16_TOK_INIT optargs topseq
    {
        xtracer.Trace("parser.p_top_after_init_optargs_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := parser16Acfg(parser16lex).NewAtom("init")
        atom.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        parser16HandleBeforeAfter(parser16Acfg(parser16lex), "after", atom, $5, $$, $4, nil)
    }
    // --- Implement ---
    | top PARSER16_TOK_IMPLEMENT atype optargs optreturns topseq
    {
        xtracer.Trace("parser.p_top_implement_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := parser16Acfg(parser16lex).NewAtom($3.(*Symbol).Rep)
        atom.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        parser16HandleBeforeAfter(parser16Acfg(parser16lex), "implement", atom, $6, $$, $4, $5)
    }
    // --- Implement type ---
    | top PARSER16_TOK_IMPLEMENT PARSER16_TOK_TYPE SYMBOLx PARSER16_TOK_WITH SYMBOLx
    {
        xtracer.Trace("parser.p_top_implement_type_symbol_with_symbol ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        // Python: a1,a2 = Atom(p[4]),Atom(p[6])
        // Python: a1.lineno = get_lineno(p,4); a2.lineno = get_lineno(p,6)
        a1 := parser16Acfg(parser16lex).NewAtom($4.Val)
        a1.SetLineno(parser16TokLineno(lex, $4))
        a2 := parser16Acfg(parser16lex).NewAtom($6.Val)
        a2.SetLineno(parser16TokLineno(lex, $6))
        // Python: impl = ImplementTypeDef(a1,a2); impl.lineno = get_lineno(p,5)
        impl := parser16Acfg(parser16lex).NewImplementTypeDef([]Node{a1, a2})
        impl.SetLineno(parser16TokLineno(lex, $5))
        // Python: d = ImplementTypeDecl(mk_lf(impl)); d.lineno = get_lineno(p,2)
        d := parser16Acfg(parser16lex).NewImplementTypeDecl(parser16MkLF(parser16Acfg(parser16lex), impl))
        d.SetLineno(parser16TokLineno(lex, $2))
        $$.declare(d)
    }
    // --- Isolate ---
    | top opttrusted PARSER16_TOK_ISOLATE SYMBOLx optargs PARSER16_TOK_EQ callatoms
    {
        xtracer.Trace("parser.p_top_opttrusted_isolate_callatom_eq_callatoms ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        // Python: ty = TrustedIsolateDef if p[2] else IsolateDef
        // Python: d = IsolateDecl(ty(*([Atom(p[4],p[5])] + p[7])))
        nameAtom := parser16Acfg(parser16lex).NewAtom($4.Val, $5...)
        elems := append([]Node{nameAtom}, $7...)
        idef := parser16Acfg(parser16lex).NewIsolateDef(elems, 0)
        idef.Trusted = $2
        idef.Elems[0].SetLineno(parser16TokLineno(lex, $3))
        idef.SetLineno(parser16TokLineno(lex, $3))
        id := parser16Acfg(parser16lex).NewIsolateDecl(idef)
        $$.declare(id)
    }
    // --- Isolate with WITH ---
    | top opttrusted PARSER16_TOK_ISOLATE SYMBOLx optargs PARSER16_TOK_EQ callatoms PARSER16_TOK_WITH callatoms
    {
        xtracer.Trace("parser.p_top_opttrusted_isolate_callatom_eq_callatoms_with_callatoms ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        // Python: ty = TrustedIsolateDef if p[2] else IsolateDef
        // Python: d = IsolateDecl(ty(*([Atom(p[4],p[5])] + p[7] + p[9])))
        nameAtom := parser16Acfg(parser16lex).NewAtom($4.Val, $5...)
        elems := append(append([]Node{nameAtom}, $7...), $9...)
        idef := parser16Acfg(parser16lex).NewIsolateDef(elems, len($9))
        idef.Trusted = $2
        idef.Elems[0].SetLineno(parser16TokLineno(lex, $3))
        idef.SetLineno(parser16TokLineno(lex, $3))
        id := parser16Acfg(parser16lex).NewIsolateDecl(idef)
        $$.declare(id)
    }
    // --- Isolate with body ---
    | top opttrusted PARSER16_TOK_ISOLATE SYMBOLx optargs PARSER16_TOK_EQ PARSER16_TOK_LCB top PARSER16_TOK_RCB optwith
    {
        xtracer.Trace("parser.p_top_opttrusted_isolate_callatom_eq_lcb_top_rcb_optwith ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        objAccum := $8
        // Python: create_object(p[0],p[4],p[5],p[8],get_lineno(p,4))
        nameAtom := parser16Acfg(parser16lex).NewAtom($4.Val)
        parser16CreateObject(parser16Acfg(parser16lex), $$, nameAtom, $5, objAccum, parser16TokLineno(lex, $4), false)
        // Python: ty = TrustedIsolateDef if p[2] else IsolateDef
        // Python: df = ty(*([Atom(p[4],p[5]),Atom(p[4],p[5])]+p[10]))
        a1 := parser16Acfg(parser16lex).NewAtom($4.Val, $5...)
        a2 := parser16Acfg(parser16lex).NewAtom($4.Val, $5...)
        elems := append([]Node{a1, a2}, $10...)
        idef := parser16Acfg(parser16lex).NewIsolateDef(elems, len($10))
        idef.Trusted = $2
        idef.IsObject = true
        idef.Elems[0].SetLineno(parser16TokLineno(lex, $3))
        idef.SetLineno(parser16TokLineno(lex, $3))
        id := parser16Acfg(parser16lex).NewIsolateObjectDecl(*parser16Acfg(parser16lex).NewIsolateDecl(idef))
        $$.declare(id)
        // Python: stack.pop() equivalent
        lex.accum = $$
    }
    // --- Extract with body ---
    | top PARSER16_TOK_EXTRACT objsym objectargs PARSER16_TOK_EQ PARSER16_TOK_LCB top PARSER16_TOK_RCB optwith
    {
        xtracer.Trace("parser.p_top_opttrusted_extract_callatom_eq_lcb_top_rcb_optwith ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        objAccum := $7
        pref := $3.(*Atom)
        // Python: create_object(p[0],p[3],p[4],p[7],get_lineno(p,3))
        parser16CreateObject(parser16Acfg(parser16lex), $$, pref, $4, objAccum, parser16NodeLineno($3), false)
        // Python: ty = ProcessDef
        // Python: d = IsolateObjectDecl(ty(*([Atom(p[3],p[4]),Atom(p[3],p[4])]+p[9])))
        a1 := parser16Acfg(parser16lex).NewAtom(pref.Rep, $4...)
        a2 := parser16Acfg(parser16lex).NewAtom(pref.Rep, $4...)
        elems := append([]Node{a1, a2}, $9...)
        idefInner := parser16Acfg(parser16lex).NewIsolateDef(elems, len($9)+1)
        idefInner.IsObject = true
        edef := parser16Acfg(parser16lex).NewExtractDef(*idefInner)
        pdef := parser16Acfg(parser16lex).NewProcessDef(*edef)
        pdef.Elems[0].SetLineno(parser16TokLineno(lex, $2))
        pdef.SetLineno(parser16TokLineno(lex, $2))
        id := parser16Acfg(parser16lex).NewIsolateObjectDecl(*parser16Acfg(parser16lex).NewIsolateDecl(pdef))
        $$.declare(id)
        // Python: stack.pop() equivalent
        lex.accum = $$
    }
    // --- Extract without body ---
    | top PARSER16_TOK_EXTRACT objsym objectargs PARSER16_TOK_EQ callatoms
    {
        xtracer.Trace("parser.p_top_extract_callatom_eq_callatoms ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        // Python: stack[-1].params = []; parent_object = None
        lex.accum.params = nil
        lex.parentObjName = ""
        // Python: d = IsolateDecl(ExtractDef(*([Atom(p[3],p[4])] + p[6])))
        pref := $3.(*Atom)
        nameAtom := parser16Acfg(parser16lex).NewAtom(pref.Rep, $4...)
        elems := append([]Node{nameAtom}, $6...)
        edef := parser16Acfg(parser16lex).NewExtractDef(*parser16Acfg(parser16lex).NewIsolateDef(elems, len($6)))
        edef.Elems[0].SetLineno(parser16TokLineno(lex, $2))
        edef.SetLineno(parser16TokLineno(lex, $2))
        id := parser16Acfg(parser16lex).NewIsolateDecl(edef)
        $$.declare(id)
    }
    // --- Export ---
    | top PARSER16_TOK_EXPORT callatom
    {
        xtracer.Trace("parser.p_top_export_callatom ENTER (top)")
        $$ = $1
        ed := parser16Acfg(parser16lex).NewExportDecl(parser16Acfg(parser16lex).NewExportDef($3, parser16Acfg(parser16lex).NewAtom("")))
        ed.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(ed)
    }
    // --- Import ---
    | top PARSER16_TOK_IMPORT callatom
    {
        xtracer.Trace("parser.p_top_import_callatom ENTER (top)")
        $$ = $1
        id := parser16Acfg(parser16lex).NewImportDecl(parser16Acfg(parser16lex).NewImportDef($3, parser16Acfg(parser16lex).NewAtom("")))
        id.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(id)
    }
    // --- Private (v1.6) ---
    | top PARSER16_TOK_PRIVATE callatom
    {
        xtracer.Trace("parser.p_top_private_callatom ENTER (top)")
        $$ = $1
        pd := &PrivateDef{Elems: []Node{$3}}
        pd.Cfg = parser16Acfg(parser16lex)
        d := parser16Acfg(parser16lex).NewPrivateDecl(pd)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(d)
    }
    // --- Delegate ---
    | top PARSER16_TOK_DELEGATE callatoms optdelegee
    {
        xtracer.Trace("parser.p_top_delegate_callatom_opt ENTER (top)")
        $$ = $1
        args := make([]Node, len($3))
        for i, s := range $3 {
            if $4 != nil {
                args[i] = parser16Acfg(parser16lex).NewDelegateDef([]Node{s, $4})
            } else {
                args[i] = parser16Acfg(parser16lex).NewDelegateDef([]Node{s})
            }
        }
        dd := parser16Acfg(parser16lex).NewDelegateDecl(args...)
        $$.declare(dd)
    }
    // --- Interpret ---
    | top PARSER16_TOK_INTERPRET oper PARSER16_TOK_ARROW oper
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_symbol ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        imp := parser16Acfg(parser16lex).NewImplies($3, $5)
        imp.SetLineno(parser16TokLineno(lex, $4))
        lf := parser16AddLabel(parser16Acfg(parser16lex), parser16MkLF(parser16Acfg(parser16lex), imp), "interp")
        d := parser16Acfg(parser16lex).NewInterpretDecl(lf)
        d.SetLineno(parser16TokLineno(lex, $4))
        $$.declare(d)
    }
    // --- Interpret with range ---
    | top PARSER16_TOK_INTERPRET oper PARSER16_TOK_ARROW PARSER16_TOK_LCB term PARSER16_TOK_DOTS term PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_lcb_symbol_dots_symbol_rcb ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        rng := parser16Acfg(parser16lex).NewRange($6, $8)
        imp := parser16Acfg(parser16lex).NewImplies($3, rng)
        imp.SetLineno(parser16TokLineno(lex, $4))
        lf := parser16AddLabel(parser16Acfg(parser16lex), parser16MkLF(parser16Acfg(parser16lex), imp), "interp")
        d := parser16Acfg(parser16lex).NewInterpretDecl(lf)
        d.SetLineno(parser16TokLineno(lex, $4))
        $$.declare(d)
    }
    // --- Interpret with enum ---
    | top PARSER16_TOK_INTERPRET oper PARSER16_TOK_ARROW PARSER16_TOK_LCB SYMBOLx moresymbols PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_lcb_symbol_moresymbols_rcb ENTER (top)")
        $$ = $1
        names := append([]string{$6.Val}, func() []string {
            var r []string
            for _, n := range $7 {
                if a, ok := n.(*Atom); ok {
                    r = append(r, a.Rep)
                }
            }
            return r
        }()...)
        atoms := make([]Node, len(names))
        for i, n := range names {
            atoms[i] = parser16Acfg(parser16lex).NewAtom(n)
        }
        es := parser16Acfg(parser16lex).NewEnumeratedSort(atoms...)
        imp := parser16Acfg(parser16lex).NewImplies($3, es)
        imp.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $4))
        lf := parser16AddLabel(parser16Acfg(parser16lex), parser16MkLF(parser16Acfg(parser16lex), imp), "interp")
        d := parser16Acfg(parser16lex).NewInterpretDecl(lf)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $4))
        $$.declare(d)
    }
    // --- Alias ---
    | top PARSER16_TOK_ALIAS SYMBOLx PARSER16_TOK_EQ callatom
    {
        xtracer.Trace("parser.p_top_aliase_symbol_eq_callatom ENTER (top)")
        $$ = $1
        d := parser16Acfg(parser16lex).NewAliasDecl(parser16Acfg(parser16lex).NewDefinition(parser16Acfg(parser16lex).NewAtom($3.Val), $5))
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        $$.declare(d)
    }
    // --- Attribute ---
    | top PARSER16_TOK_ATTRIBUTE callatom PARSER16_TOK_EQ attributeval
    {
        xtracer.Trace("parser.p_top_attribute_callatom_eq_attributeval ENTER (top)")
        $$ = $1
        lex := parser16lex.(*parser16LexAdapter)
        adef := parser16Acfg(parser16lex).NewAttributeDef($3, $5)
        adef.SetLineno(parser16TokLineno(lex, $2))
        d := parser16Acfg(parser16lex).NewAttributeDecl(adef)
        d.SetLineno(parser16TokLineno(lex, $2))
        $$.declare(d)
    }
    // --- Variant ---
    | top PARSER16_TOK_VARIANT typesymbol PARSER16_TOK_OF atype
    {
        xtracer.Trace("parser.p_top_variant_symbol_of_atype ENTER (top)")
        $$ = $1
        scnst := parser16Acfg(parser16lex).NewAtom($3.(*Atom).Rep)
        scnst.SetLineno(parser16NodeLineno($3))
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, parser16Acfg(parser16lex).NewUninterpretedSortAST())
        tdfn.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $4))
        td := parser16Acfg(parser16lex).NewTypeDecl(tdfn)
        $$.declare(td)
        vdfn := parser16Acfg(parser16lex).NewVariantDef(scnst, parser16Acfg(parser16lex).NewAtom(NodeRep($5)))
        vd := parser16Acfg(parser16lex).NewVariantDecl(vdfn)
        $$.declare(vd)
    }
    // --- Variant with sort ---
    | top PARSER16_TOK_VARIANT typesymbol PARSER16_TOK_OF atype PARSER16_TOK_EQ sort
    {
        xtracer.Trace("parser.p_top_variant_symbol_of_symbol_eq_sort ENTER (top)")
        $$ = $1
        scnst := parser16Acfg(parser16lex).NewAtom($3.(*Atom).Rep)
        scnst.SetLineno(parser16NodeLineno($3))
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, $7)
        tdfn.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $4))
        td := parser16Acfg(parser16lex).NewTypeDecl(tdfn)
        $$.declare(td)
        vdfn := parser16Acfg(parser16lex).NewVariantDef(scnst, parser16Acfg(parser16lex).NewAtom(NodeRep($5)))
        vd := parser16Acfg(parser16lex).NewVariantDecl(vdfn)
        $$.declare(vd)
    }
    // --- Nativequote ---
    | top PARSER16_TOK_NATIVEQUOTE
    {
        xtracer.Trace("parser.p_top_nativequote ENTER (top)")
        $$ = $1
        // Python: text,bqs = parse_nativequote(p,2)
        // Python: defn = NativeDef(*([mk_label(None,'native')] + [text] + bqs))
        // Python: thing = NativeDecl(defn)
        text, bqs := parser16ParseNativequote(parser16Acfg(parser16lex), $2.Val, parser16lex.(*parser16LexAdapter))
        label := parser16NewLabel(parser16Acfg(parser16lex), "native")
        defnArgs := append([]Node{label, parser16Acfg(parser16lex).NewNativeCode(text)}, bqs...)
        defn := parser16Acfg(parser16lex).NewNativeDef(defnArgs)
        defn.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        thing := parser16Acfg(parser16lex).NewNativeDecl(defn)
        thing.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(thing)
    }
    // --- Scenario ---
    | top PARSER16_TOK_SCENARIO PARSER16_TOK_LCB sceninit PARSER16_TOK_SEMI scentranss PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_top_scenario_lcb_sceninit_semi_scentranss_rcb ENTER (top)")
        $$ = $1
        elems := append([]Node{$4}, $6...)
        sdef := parser16Acfg(parser16lex).NewScenarioDef(elems)
        sdef.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        sd := parser16Acfg(parser16lex).NewScenarioDecl(sdef)
        $$.declare(sd)
    }
    // --- Spec/Impl blocks ---
    | top specimpl PARSER16_TOK_LCB top PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_top_specification_lcb_top_rcb ENTER (top)")
        $$ = $1
        innerAccum := $4
        // Python: stack.pop() — restore scope after processing specimpl body
        parser16lex.(*parser16LexAdapter).accum = $$
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
    | top PARSER16_TOK_STATE SYMBOLx PARSER16_TOK_EQ state_expr
    {
        xtracer.Trace("parser.p_top_state_symbol_eq_state_expr ENTER (top)")
        $$ = $1
        sd := parser16Acfg(parser16lex).NewStateDef($3.Val, $5)
        _ = sd
    }
    // --- Init statement (v1.6) ---
    | top PARSER16_TOK_INIT labeledfmla
    {
        xtracer.Trace("parser.p_top_init_fmla ENTER (top)")
        $$ = $1
        d := parser16Acfg(parser16lex).NewInitDecl(parser16CheckNonTemporal($3))
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(d)
    }
    // --- Assert implication (v1.6) ---
    | top PARSER16_TOK_ASSERT SYMBOLx PARSER16_TOK_ARROW assert_rhs
    {
        xtracer.Trace("parser.p_top_assert_symbol_arrow_assert_rhs ENTER (top)")
        $$ = $1
        lhs := parser16Acfg(parser16lex).NewAtom($3.Val)
        thing := parser16Acfg(parser16lex).NewImplies(lhs, $5)
        thing.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $4))
        d := parser16Acfg(parser16lex).NewAssertDecl(thing)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$.declare(d)
    }
    ;

// ============================================================
// --- SYMBOL handling ---
// ============================================================

SYMBOLx:
    PARSER16_TOK_PRESYMBOL
    {
        xtracer.Trace("parser.p_SYMBOL_PRESYMBOL ENTER (SYMBOL) val=%s", $1.Val)
        $$ = $1
    }
    | SYMBOLx PARSER16_TOK_LB SYMsubscr PARSER16_TOK_RB
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
    | PARSER16_TOK_THIS
    {
        xtracer.Trace("parser.p_SYMsubscr_THIS ENTER (SYMsubscr)")
        $$ = TokenInfo{Val: "this", Line: $1.Line}
    }
    | SYMsubscr PARSER16_TOK_DOT SYMBOLx
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
        $$ = parser16Acfg(parser16lex).NewSymbol($1.Val, nil)
    }
    | atype PARSER16_TOK_DOT SYMBOLx
    {
        xtracer.Trace("parser.p_atype_atype_dot_symbol ENTER (atype)")
        if _, ok := $1.(*This); ok {
            $$ = parser16Acfg(parser16lex).NewSymbol($3.Val, nil)
        } else if sym, ok := $1.(*Symbol); ok {
            $$ = parser16Acfg(parser16lex).NewSymbol(sym.Rep + "." + $3.Val, nil)
        } else {
            $$ = parser16Acfg(parser16lex).NewSymbol($3.Val, nil)
        }
    }
    | PARSER16_TOK_THIS
    {
        xtracer.Trace("parser.p_atype_this ENTER (atype)")
        t := parser16Acfg(parser16lex).NewThis()
        t.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = t
    }
    ;

// ============================================================
// --- aterm (v1.3-v1.6) ---
// ============================================================

aterm:
    SYMBOLx
    {
        xtracer.Trace("parser.p_aterm_symbol ENTER (aterm)")
        a := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1.Val, nil))
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | aterm PARSER16_TOK_LPAREN terms PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_aterm_aterm_terms ENTER (aterm)")
        switch a := $1.(type) {
        case *App:
            a.Terms = append(a.Terms, $3...)
            $$ = a
        case *Atom:
            a.Terms = append(a.Terms, $3...)
            $$ = a
        default:
            $$ = $1
        }
    }
    | aterm PARSER16_TOK_DOT SYMBOLx
    {
        xtracer.Trace("parser.p_term_term_dot_term ENTER (aterm)")
        rhs := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($3.Val, nil))
        rhs.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        res := ComposeAtomsGeneric($1, rhs)
        res.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = res
    }
    ;

// ============================================================
// --- Variables ---
// ============================================================

var:
    PARSER16_TOK_VARIABLE
    {
        xtracer.Trace("parser.p_var_variable ENTER (var)")
        // Python: Variable(p[1], universe) where universe = 'S'
        v := parser16Acfg(parser16lex).NewVariable($1.Val, "S")
        v.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = v
    }
    | PARSER16_TOK_VARIABLE PARSER16_TOK_COLON atype
    {
        xtracer.Trace("parser.p_var_variable_colon_symbol ENTER (var)")
        // Python: Variable(p[1], p[3]) where p[3] is a string from atype
        v := parser16Acfg(parser16lex).NewVariable($1.Val, parser16AtypeToString($3))
        v.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = v
    }
    ;

simplevar:
    PARSER16_TOK_VARIABLE
    {
        xtracer.Trace("parser.p_simplevar_variable ENTER (simplevar)")
        // Python: Variable(p[1], universe) where universe = 'S'
        v := parser16Acfg(parser16lex).NewVariable($1.Val, "S")
        v.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = v
    }
    | PARSER16_TOK_VARIABLE PARSER16_TOK_COLON SYMBOLx
    {
        xtracer.Trace("parser.p_simplevar_variable_colon_symbol ENTER (simplevar)")
        // Python: Variable(p[1], p[3]) where p[3] is a string
        v := parser16Acfg(parser16lex).NewVariable($1.Val, $3.Val)
        v.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = v
    }
    ;

vars:
    var
    {
        xtracer.Trace("parser.p_vars_var ENTER (vars)")
        $$ = []Node{$1}
    }
    | vars PARSER16_TOK_COMMA var
    {
        xtracer.Trace("parser.p_vars_vars_comma_var ENTER (vars)")
        $$ = append($1, $3)
    }
    ;

simplevars:
    simplevar
    {
        xtracer.Trace("parser.p_simplevars_simplevar ENTER (simplevars)")
        $$ = []Node{$1}
    }
    | simplevars PARSER16_TOK_COMMA simplevar
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
        $$ = []Node{$1}
    }
    | terms PARSER16_TOK_COMMA term
    {
        xtracer.Trace("parser.p_terms_terms_term ENTER (terms)")
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Terms (v1.6) ---
// ============================================================

term:
    aterm
    {
        xtracer.Trace("parser.p_term_aterm ENTER (term)")
        $$ = $1
    }
    | var
    {
        xtracer.Trace("parser.p_term_var ENTER (term)")
        $$ = $1
    }
    | PARSER16_TOK_OLD aterm
    {
        xtracer.Trace("parser.p_aterm_old_symbol ENTER (term)")
        o := parser16Acfg(parser16lex).NewOld($2)
        o.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = o
    }
    | PARSER16_TOK_LPAREN term PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_term_lp_term_lp ENTER (term)")
        $$ = $2
    }
    // --- Arithmetic ---
    | term PARSER16_TOK_PLUS term
    {
        xtracer.Trace("parser.p_term_term_PLUS_term ENTER (term)")
        n := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("+", nil), $1, $3)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = n
    }
    | term PARSER16_TOK_MINUS term
    {
        xtracer.Trace("parser.p_term_term_MINUS_term ENTER (term)")
        n := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("-", nil), $1, $3)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = n
    }
    | term PARSER16_TOK_TIMES term
    {
        xtracer.Trace("parser.p_term_term_TIMES_term ENTER (term)")
        n := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("*", nil), $1, $3)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = n
    }
    | term PARSER16_TOK_DIV term
    {
        xtracer.Trace("parser.p_term_term_DIV_term ENTER (term)")
        n := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("/", nil), $1, $3)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = n
    }
    // --- If/else ---
    | term PARSER16_TOK_IF fmla PARSER16_TOK_ELSE term
    {
        xtracer.Trace("parser.p_term_if_fmla_else_term ENTER (term)")
        n := parser16Acfg(parser16lex).NewIte($3, $1, $5)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = n
    }
    // --- Named binders ---
    | PARSER16_TOK_LPAREN PARSER16_TOK_DOLLAR SYMBOLx simplevars PARSER16_TOK_DOT fmla PARSER16_TOK_RPAREN PARSER16_TOK_LPAREN terms PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_term_namedbinder_vars_dot_term ENTER (term)")
        // Python: x = NamedBinder(p[3], p[4], p[6]); x.lineno = get_lineno(p,2)
        // Python: p[0] = App(x, p[9]); p[0].lineno = get_lineno(p,2)
        binder := parser16Acfg(parser16lex).NewNamedBinder($3.Val, $4, $6)
        binder.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = parser16Acfg(parser16lex).NewApp(binder, $9...)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | PARSER16_TOK_DOLLAR SYMBOLx PARSER16_TOK_DOT fmla     %prec PARSER16_TOK_SEMI
    {
        xtracer.Trace("parser.p_term_namedbinder_dot_fmla ENTER (term)")
        $$ = parser16Acfg(parser16lex).NewNamedBinder($2.Val, nil, $4)
    }
    | PARSER16_TOK_DOLLAR SYMBOLx PARSER16_TOK_DOLLAR fmla   %prec PARSER16_TOK_SEMI
    {
        xtracer.Trace("parser.p_term_namedbinder_dollar_fmla ENTER (term)")
        $$ = parser16Acfg(parser16lex).NewNamedBinder($2.Val, nil, $4)
    }
    ;

// ============================================================
// --- fmla ---
// ============================================================

fmla:
    term
    {
        xtracer.Trace("parser.p_fmla_term ENTER (fmla)")
        $$ = AppToAtom($1)
    }
    | term relop term
    {
        xtracer.Trace("parser.p_fmla_term_relop_term ENTER (fmla)")
        n := parser16Acfg(parser16lex).NewAtom($2, $1, $3)
        n.SetLineno(parser16GetLineno(parser16lex.(*parser16LexAdapter)))
        $$ = n
    }
    | term PARSER16_TOK_TILDAEQ term
    {
        xtracer.Trace("parser.p_fmla_term_tildaeq_term ENTER (fmla)")
        n := parser16Acfg(parser16lex).NewNot(parser16Acfg(parser16lex).NewAtom("=", $1, $3))
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = n
    }
    | PARSER16_TOK_LPAREN fmla PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_fmla_lparen_fmla_rparen ENTER (fmla)")
        $$ = $2
    }
    | PARSER16_TOK_TRUE
    {
        xtracer.Trace("parser.p_fmla_true ENTER (fmla)")
        $$ = parser16Acfg(parser16lex).NewAnd()
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_FALSE
    {
        xtracer.Trace("parser.p_fmla_false ENTER (fmla)")
        $$ = parser16Acfg(parser16lex).NewOr()
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_TILDA fmla
    {
        xtracer.Trace("parser.p_fmla_not_fmla ENTER (fmla)")
        n := parser16Acfg(parser16lex).NewNot($2)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = n
    }
    | fmla PARSER16_TOK_AND fmla
    {
        xtracer.Trace("parser.p_fmla_fmla_and_fmla ENTER (fmla)")
        if existing, ok := $1.(*And); ok {
            existing.Terms = append(existing.Terms, $3)
            $$ = existing
        } else {
            n := parser16Acfg(parser16lex).NewAnd($1, $3)
            n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
            $$ = n
        }
    }
    | fmla PARSER16_TOK_OR fmla
    {
        xtracer.Trace("parser.p_fmla_fmla_or_fmla ENTER (fmla)")
        if existing, ok := $1.(*Or); ok {
            existing.Terms = append(existing.Terms, $3)
            $$ = existing
        } else {
            n := parser16Acfg(parser16lex).NewOr($1, $3)
            n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
            $$ = n
        }
    }
    | fmla PARSER16_TOK_ARROW fmla
    {
        xtracer.Trace("parser.p_fmla_fmla_arrow_fmla ENTER (fmla)")
        n := parser16Acfg(parser16lex).NewImplies($1, $3)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = n
    }
    | fmla PARSER16_TOK_IFF fmla
    {
        xtracer.Trace("parser.p_fmla_fmla_iff_fmla ENTER (fmla)")
        n := parser16Acfg(parser16lex).NewIff($1, $3)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = n
    }
    | PARSER16_TOK_FORALL simplevars PARSER16_TOK_DOT fmla
    {
        xtracer.Trace("parser.p_fmla_forall_vars_dot_fmla ENTER (fmla)")
        n := parser16Acfg(parser16lex).NewForall($2, $4)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = n
    }
    | PARSER16_TOK_EXISTS simplevars PARSER16_TOK_DOT fmla
    {
        xtracer.Trace("parser.p_fmla_exists_vars_dot_fmla ENTER (fmla)")
        n := parser16Acfg(parser16lex).NewExists($2, $4)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = n
    }
    | PARSER16_TOK_GLOBALLY fmla
    {
        xtracer.Trace("parser.p_fmla_globally_fmla ENTER (fmla)")
        n := parser16Acfg(parser16lex).NewGlobally($2)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = n
    }
    | PARSER16_TOK_EVENTUALLY fmla
    {
        xtracer.Trace("parser.p_fmla_eventually_fmla ENTER (fmla)")
        n := parser16Acfg(parser16lex).NewEventually($2)
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = n
    }
    ;

// ============================================================
// --- labeledfmla ---
// ============================================================

labeledfmla:
    fmla
    {
        xtracer.Trace("parser.p_labeledfmla_fmla ENTER (labeledfmla)")
        lf := parser16Acfg(parser16lex).NewLabeledFormula(nil, $1)
        // Python: p[0].lineno = p[1].lineno — copy formula's Location
        lf.SetLineno($1.GetLineno())
        $$ = lf
    }
    | labelname fmla
    {
        xtracer.Trace("parser.p_labeledfmla_label_fmla ENTER (labeledfmla)")
        // Python: Atom(p[1][1:-1],[]) — brackets already stripped by labelname rule
        lf := parser16Acfg(parser16lex).NewLabeledFormula(parser16Acfg(parser16lex).NewAtom($1.Val), $2)
        // Python: p[0].lineno = get_lineno(p,1)
        lf.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = lf
    }
    ;

// labelname matches Python's LABEL : LB SYMBOL RB (ivy_logic_parser.py:23-25).
// The Python lexer produces LABEL as a terminal, but our lexer produces
// separate LB, SYMBOL, RB tokens, so we combine them in the grammar.
labelname:
    PARSER16_TOK_LB SYMBOLx PARSER16_TOK_RB
    {
        xtracer.Trace("parser.p_LABEL_LB_SYMBOL_RB ENTER (LABEL)")
        // Python: all LABEL consumers strip brackets with [1:-1].
        // Strip here at the source so all consumers get the bare name.
        $$ = TokenInfo{Val: $2.Val, Line: $1.Line}
    }
    | PARSER16_TOK_LABEL
    {
        xtracer.Trace("parser.p_labelname__label ENTER (labelname)")
        // PARSER16_TOK_LABEL comes from lexer with brackets "[name]"; strip them
        // to match Python's p[N][1:-1] pattern applied by all consumers.
        val := $1.Val
        if len(val) >= 2 && val[0] == '[' && val[len(val)-1] == ']' {
            val = val[1 : len(val)-1]
        }
        $$ = TokenInfo{Val: val, Line: $1.Line}
    }
    ;

// ============================================================
// --- lgprop / gprop (v1.6) ---
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
        lf := parser16Acfg(parser16lex).NewLabeledFormula($1, $2)
        lf.SetLineno(parser16NodeLineno($2))
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
    | PARSER16_TOK_TEMPORAL
    {
        xtracer.Trace("parser.p_opttemporal_symbol ENTER (opttemporal)")
        $$ = parser16Acfg(parser16lex).NewAnd() // non-nil marker
    }
    ;

optunprovable:
    /* empty */
    {
        xtracer.Trace("parser.p_optunprovable ENTER (optunprovable)")
        $$ = nil
    }
    | PARSER16_TOK_UNPROVABLE
    {
        xtracer.Trace("parser.p_optunprovable_symbol ENTER (optunprovable)")
        $$ = parser16Acfg(parser16lex).NewAnd() // non-nil marker
    }
    ;

optexplicit:
    /* empty */
    {
        xtracer.Trace("parser.p_optexplicit ENTER (optexplicit)")
        $$ = nil
    }
    | PARSER16_TOK_EXPLICIT
    {
        xtracer.Trace("parser.p_optexplicit_explicit ENTER (optexplicit)")
        $$ = parser16Acfg(parser16lex).NewAnd() // non-nil marker
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
        $$ = parser16Acfg(parser16lex).NewAtom($1.Val)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

optskolem:
    /* empty */
    {
        xtracer.Trace("parser.p_optskolem ENTER (optskolem)")
        $$ = nil
    }
    | PARSER16_TOK_NAMED defnlhs
    {
        xtracer.Trace("parser.p_optskolem_symbol ENTER (optskolem)")
        $$ = $2
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

optinit:
    /* empty */
    {
        xtracer.Trace("parser.p_optinit ENTER (optinit)")
        $$ = nil
    }
    | PARSER16_TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_optinit_assign_fmla ENTER (optinit)")
        $$ = parser16CheckNonTemporal($2)
    }
    ;

optproof:
    /* empty */
    {
        xtracer.Trace("parser.p_optproof ENTER (optproof)")
        $$ = nil
    }
    | PARSER16_TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_optproof_symbol ENTER (optproof)")
        $$ = $2
    }
    | PARSER16_TOK_PROOF labelname proofstep
    {
        xtracer.Trace("parser.p_optproof_label_proofstep ENTER (optproof)")
        label := parser16Acfg(parser16lex).NewAtom($2.Val)
        label.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        lf := parser16Acfg(parser16lex).NewLabeledFormula(label, $3)
        lf.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
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
    | PARSER16_TOK_SEMI
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
        _ = parser16TokLineno(parser16lex.(*parser16LexAdapter), $1)
    }
    | dotsym PARSER16_TOK_DOT SYMBOLx
    {
        xtracer.Trace("parser.p_dotsym_dotsym_dot_symbol ENTER (dotsym)")
        $$ = $1 + "." + $3.Val
    }
    ;

defnlhs:
    dotsym
    {
        xtracer.Trace("parser.p_defnlhs_symbol ENTER (defnlhs)")
        $$ = parser16Acfg(parser16lex).NewAtom($1)
    }
    | dotsym PARSER16_TOK_LPAREN defargs PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_defnlhs_symbol_lparen_defargs_rparen ENTER (defnlhs)")
        a := parser16Acfg(parser16lex).NewAtom($1)
        a.Terms = $3
        $$ = a
    }
    | PARSER16_TOK_LPAREN defarg relop defarg PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_defnlhs_lp_term_relop_term_rp ENTER (defnlhs)")
        a := parser16Acfg(parser16lex).NewAtom($3, $2, $4)
        a.SetLineno(parser16GetLineno(parser16lex.(*parser16LexAdapter)))
        $$ = a
    }
    | PARSER16_TOK_LPAREN defarg infix defarg PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_defnlhs_lp_term_infix_term_rp ENTER (defnlhs)")
        a := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($3, nil), $2, $4)
        a.SetLineno(parser16GetLineno(parser16lex.(*parser16LexAdapter)))
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
        $$ = []Node{$1}
    }
    | defargs PARSER16_TOK_COMMA defarg
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
    | defnlhs PARSER16_TOK_COLON atype
    {
        xtracer.Trace("parser.p_typeddefn_defnlhs_colon_atype ENTER (typeddefn)")
        // set sort on the defnlhs
        if a, ok := $1.(*Atom); ok {
            a.ASort = $3
        } else if app, ok := $1.(*App); ok {
            app.ASort = $3
        }
        $$ = $1
    }
    ;

defnrhs:
    fmla
    {
        xtracer.Trace("parser.p_defnrhs_fmla ENTER (defnrhs)")
        $$ = parser16CheckNonTemporal($1)
    }
    | somevarfmla
    {
        xtracer.Trace("parser.p_defnrhs_somevarfmla ENTER (defnrhs)")
        $$ = parser16CheckNonTemporal($1)
    }
    | PARSER16_TOK_NATIVEQUOTE
    {
        xtracer.Trace("parser.p_defnrhs_nativequote ENTER (defnrhs)")
        text, bqs := parser16ParseNativequote(parser16Acfg(parser16lex), $1.Val, parser16lex.(*parser16LexAdapter))
        elems := append([]Node{parser16Acfg(parser16lex).NewAtom(text)}, bqs...)
        ne := parser16Acfg(parser16lex).NewNativeExpr(elems)
        ne.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = ne
    }
    ;

defn:
    typeddefn PARSER16_TOK_EQ defnrhs
    {
        xtracer.Trace("parser.p_defn_atom_fmla ENTER (defn)")
        d := parser16Acfg(parser16lex).NewDefinition(AppToAtom($1), $3)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = d
    }
    ;

defns:
    defn
    {
        xtracer.Trace("parser.p_defns_defn ENTER (defns)")
        $$ = []Node{$1}
    }
    | defns PARSER16_TOK_COMMA defn
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
    | PARSER16_TOK_LCB defn PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_gdefn_lcb_defn_rcb ENTER (gdefn)")
        d := $2.(*Definition)
        $$ = parser16Acfg(parser16lex).NewDefinitionSchema(*d)
    }
    ;

somevarfmla:
    PARSER16_TOK_SOME simplevar PARSER16_TOK_DOT fmla optin optelse
    {
        xtracer.Trace("parser.p_somevarfmla_some_simplevar_dot_fmla ENTER (somevarfmla)")
        se := parser16Acfg(parser16lex).NewSomeExpr($2, $4)
        if $5 != nil { se.IfValue = $5 }
        if $6 != nil { se.ElseVal = $6 }
        se.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = se
    }
    ;

optin:
    /* empty */
    {
        xtracer.Trace("parser.p_optin ENTER (optin)")
        $$ = nil
    }
    | PARSER16_TOK_IN fmla
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
    | PARSER16_TOK_ELSE fmla
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
        $$ = parser16CheckNonTemporal($1)
    }
    | PARSER16_TOK_LCB schdecls schconc PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_schdefnrhs_lcb_schdecls_rcb ENTER (schdefnrhs)")
        args := append($2, $3)
        $$ = parser16Acfg(parser16lex).NewSchemaBody(args...)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

schdecl:
    PARSER16_TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_schdecl_funcdecl ENTER (schdecl)")
        $$ = make([]Node, len($2))
        copy($$, $2)
    }
    | PARSER16_TOK_FRESH PARSER16_TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_schdecl_fresh_funcdecl ENTER (schdecl)")
        $$ = make([]Node, len($3))
        copy($$, $3)
    }
    | PARSER16_TOK_INDIV funs
    {
        xtracer.Trace("parser.p_schdecl_indivdecl ENTER (schdecl)")
        $$ = make([]Node, len($2))
        copy($$, $2)
    }
    | PARSER16_TOK_FRESH PARSER16_TOK_INDIV funs
    {
        xtracer.Trace("parser.p_schdecl_fresh_indivdecl ENTER (schdecl)")
        $$ = make([]Node, len($3))
        copy($$, $3)
    }
    | PARSER16_TOK_RELATION rels
    {
        xtracer.Trace("parser.p_schdecl_relation_rel ENTER (schdecl)")
        $$ = make([]Node, len($2))
        copy($$, $2)
    }
    | PARSER16_TOK_FRESH PARSER16_TOK_RELATION rels
    {
        xtracer.Trace("parser.p_schdecl_fresh_relation_rel ENTER (schdecl)")
        $$ = make([]Node, len($3))
        copy($$, $3)
    }
    | PARSER16_TOK_TYPE SYMBOLx
    {
        xtracer.Trace("parser.p_schdecl_typedecl ENTER (schdecl)")
        scnst := parser16Acfg(parser16lex).NewAtom($2.Val)
        scnst.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        tdfn := parser16Acfg(parser16lex).NewTypeDef(scnst, parser16Acfg(parser16lex).NewUninterpretedSortAST())
        tdfn.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = []Node{tdfn}
    }
    | optexplicit PARSER16_TOK_PROPERTY lgprop
    {
        xtracer.Trace("parser.p_schdecl_propdecl ENTER (schdecl)")
        lf := parser16AddLabel(parser16Acfg(parser16lex), $3.(*LabeledFormula), "prop")
        if $1 != nil {
            lf.Explicit = true
        }
        $$ = []Node{parser16CheckNonTemporal(lf).(*LabeledFormula)}
    }
    | PARSER16_TOK_THEOREM lgprop
    {
        xtracer.Trace("parser.p_schdecl_theorem_lgprop ENTER (schdecl)")
        lf := parser16AddLabel(parser16Acfg(parser16lex), $2.(*LabeledFormula), "prop")
        $$ = []Node{parser16CheckNonTemporal(lf).(*LabeledFormula)}
    }
    | schdefnrhs
    {
        xtracer.Trace("parser.p_schdecl_theorem ENTER (schdecl)")
        lf := parser16Acfg(parser16lex).NewLabeledFormula(nil, $1)
        if $1 != nil && $1.GetLineno().Line > 0 {
            lf.SetLineno($1.GetLineno())
        }
        lf = parser16AddLabel(parser16Acfg(parser16lex), lf, "sch")
        $$ = []Node{lf}
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
    PARSER16_TOK_DEFINITION defn
    {
        xtracer.Trace("parser.p_schconc_defdecl ENTER (schconc)")
        $$ = $2
    }
    | optexplicit PARSER16_TOK_PROPERTY lgprop
    {
        xtracer.Trace("parser.p_schconc_propdecl ENTER (schconc)")
        lf := $3.(*LabeledFormula)
        // Python: p[0] = check_non_temporal(fmla)
        $$ = parser16CheckNonTemporal(lf.Formula)
    }
    ;

schdefn:
    defnlhs PARSER16_TOK_EQ schdefnrhs
    {
        xtracer.Trace("parser.p_schdefn_atom_eq_fmla ENTER (schdefn)")
        $$ = parser16Acfg(parser16lex).NewDefinition(AppToAtom($1), $3)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
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
    | PARSER16_TOK_DESTRUCTOR tterms
    {
        xtracer.Trace("parser.p_symdecl_destructor_tterms ENTER (symdecl)")
        d := parser16Acfg(parser16lex).NewDestructorDecl($2...)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = d
    }
    | PARSER16_TOK_FIELD tterms
    {
        xtracer.Trace("parser.p_symdecl_field_tterms ENTER (symdecl)")
        // Python: arg0 = Variable('SELF',This()); arg0.lineno = get_lineno(p,1)
        // Python: Variable('SELF', This()) — This() is special; use "this" as sort string
        arg0 := parser16Acfg(parser16lex).NewVariable("SELF", "this")
        arg0.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        // Python: tterms = [x.clone([arg0]+x.args) for x in p[2]]
        // Python: for x,y in zip(p[2],tterms): y.lineno = x.lineno
        cloned := make([]Node, len($2))
        for i, x := range $2 {
            newArgs := append([]Node{arg0}, x.Args()...)
            cloned[i] = x.Clone(newArgs)
            cloned[i].SetLineno(x.GetLineno())
        }
        d := parser16Acfg(parser16lex).NewDestructorDecl(cloned...)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = d
    }
    | PARSER16_TOK_CONSTRUCTOR tterms
    {
        xtracer.Trace("parser.p_symdecl_constructor_tterms ENTER (symdecl)")
        // Python: for t in p[2]: if not hasattr(t,'sort'): t.sort = This()
        lex := parser16lex.(*parser16LexAdapter)
        for _, t := range $2 {
            if app, ok := t.(*App); ok && app.ASort == nil {
                thisNode := parser16Acfg(parser16lex).NewThis()
                thisNode.SetLineno(parser16TokLineno(lex, $1))
                app.ASort = thisNode
            }
        }
        d := parser16Acfg(parser16lex).NewConstructorDecl($2...)
        d.SetLineno(parser16TokLineno(lex, $1))
        $$ = d
    }
    ;

constantdecl:
    PARSER16_TOK_INDIV tterms
    {
        xtracer.Trace("parser.p_constantdecl_constant_tterms ENTER (constantdecl)")
        d := parser16Acfg(parser16lex).NewConstantDecl($2...)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = d
    }
    | PARSER16_TOK_VAR tterms
    {
        xtracer.Trace("parser.p_constantdecl_var_tterms ENTER (constantdecl)")
        d := parser16Acfg(parser16lex).NewConstantDecl($2...)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = d
    }
    | PARSER16_TOK_PARAMETER parameter
    {
        xtracer.Trace("parser.p_constantdecl_parameter_tterm ENTER (constantdecl)")
        $$ = $2
    }
    ;

parameter:
    tterm
    {
        xtracer.Trace("parser.p_param_tterm ENTER (parameter)")
        d := parser16Acfg(parser16lex).NewParameterDecl($1)
        $$ = d
    }
    | tterm PARSER16_TOK_EQ paramval
    {
        xtracer.Trace("parser.p_param_tterm_eq_paramval ENTER (parameter)")
        df := parser16Acfg(parser16lex).NewDefinition($1, $3)
        df.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        d := parser16Acfg(parser16lex).NewParameterDecl(df)
        $$ = d
    }
    ;

paramval:
    PARSER16_TOK_TRUE
    {
        xtracer.Trace("parser.p_paramval_true ENTER (paramval)")
        $$ = parser16Acfg(parser16lex).NewAtom("true")
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_FALSE
    {
        xtracer.Trace("parser.p_paramval_false ENTER (paramval)")
        $$ = parser16Acfg(parser16lex).NewAtom("false")
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | SYMBOLx
    {
        xtracer.Trace("parser.p_paramval_symbol ENTER (paramval)")
        $$ = parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1.Val, nil))
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

tapp:
    SYMBOLx
    {
        xtracer.Trace("parser.p_tapp_symbol ENTER (tapp)")
        a := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1.Val, nil))
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | SYMBOLx targs
    {
        xtracer.Trace("parser.p_tapp_symbol_targs ENTER (tapp)")
        args := make([]Node, len($2))
        copy(args, $2)
        a := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1.Val, nil), args...)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_LPAREN var infix var PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_tapp_lp_symbol_infix_symbol_rp ENTER (tapp)")
        a := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($3, nil), $2, $4)
        a.SetLineno(parser16GetLineno(parser16lex.(*parser16LexAdapter)))
        $$ = a
    }
    ;

tterm:
    tapp
    {
        xtracer.Trace("parser.p_tterm_term ENTER (tterm)")
        $$ = $1
    }
    | tapp PARSER16_TOK_COLON atype
    {
        xtracer.Trace("parser.p_tterm_term_colon_symbol ENTER (tterm)")
        if app, ok := $1.(*App); ok {
            app.ASort = $3
        }
        $$ = $1
    }
    ;

tterms:
    tterm
    {
        xtracer.Trace("parser.p_tterms_tterm ENTER (tterms)")
        $$ = []Node{$1}
    }
    | tterms PARSER16_TOK_COMMA tterm
    {
        xtracer.Trace("parser.p_tterms_tterms_comma_tterm ENTER (tterms)")
        $$ = append($1, $3)
    }
    ;

targs:
    PARSER16_TOK_LPAREN PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_targs_lparen_rparen ENTER (targs)")
        $$ = nil
    }
    | PARSER16_TOK_LPAREN tsyms PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_targs_lparen_tsyms_rparen ENTER (targs)")
        $$ = $2
    }
    ;

tsyms:
    var
    {
        xtracer.Trace("parser.p_tsyms_tsym ENTER (tsyms)")
        $$ = []Node{$1}
    }
    | tsyms PARSER16_TOK_COMMA var
    {
        xtracer.Trace("parser.p_tsyms_tsyms_comma_tsym ENTER (tsyms)")
        $$ = append($1, $3)
    }
    ;

tatom:
    SYMBOLx
    {
        xtracer.Trace("parser.p_tatom_symbol ENTER (tatom)")
        $$ = parser16Acfg(parser16lex).NewAtom($1.Val)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | SYMBOLx targs
    {
        xtracer.Trace("parser.p_tatom_symbol_targs ENTER (tatom)")
        $$ = parser16Acfg(parser16lex).NewAtom($1.Val, $2...)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_LPAREN var relop var PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_tatom_lp_symbol_relop_symbol_rp ENTER (tatom)")
        $$ = parser16Acfg(parser16lex).NewAtom($3, $2, $4)
        $$.SetLineno(parser16GetLineno(parser16lex.(*parser16LexAdapter)))
    }
    ;

tatoms:
    tatom
    {
        xtracer.Trace("parser.p_tatoms_tatom ENTER (tatoms)")
        $$ = []Node{$1}
    }
    | tatoms PARSER16_TOK_COMMA tatom
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
        if a, ok := $1.(*Atom); ok {
            a.ASort = parser16Acfg(parser16lex).NewSymbol("bool", nil)
        } else if app, ok := $1.(*App); ok {
            app.ASort = parser16Acfg(parser16lex).NewSymbol("bool", nil)
        }
        d := parser16Acfg(parser16lex).NewConstantDecl($1)
        $$ = d
    }
    | defn
    {
        xtracer.Trace("parser.p_rel_defn ENTER (rel)")
        lf := parser16AddLabel(parser16Acfg(parser16lex), parser16MkLF(parser16Acfg(parser16lex), $1), "def")
        d := parser16Acfg(parser16lex).NewDerivedDecl(lf)
        $$ = d
    }
    ;

rels:
    rel
    {
        xtracer.Trace("parser.p_rels_rel ENTER (rels)")
        $$ = []Node{$1}
    }
    | rels PARSER16_TOK_COMMA rel
    {
        xtracer.Trace("parser.p_rels_rels_comma_rel ENTER (rels)")
        $$ = append($1, $3)
    }
    ;

fun:
    typeddefn
    {
        xtracer.Trace("parser.p_fun_defnlhs_colon_atype ENTER (fun)")
        d := parser16Acfg(parser16lex).NewConstantDecl($1)
        $$ = d
    }
    | typeddefn PARSER16_TOK_EQ defnrhs
    {
        xtracer.Trace("parser.p_fun_defn ENTER (fun)")
        df := parser16Acfg(parser16lex).NewDefinition(AppToAtom($1), $3)
        df.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        lf := parser16AddLabel(parser16Acfg(parser16lex), parser16MkLF(parser16Acfg(parser16lex), df), "def")
        d := parser16Acfg(parser16lex).NewDerivedDecl(lf)
        $$ = d
    }
    ;

funs:
    fun
    {
        xtracer.Trace("parser.p_funs_fun ENTER (funs)")
        $$ = []Node{$1}
    }
    | funs PARSER16_TOK_COMMA fun
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
        $$ = parser16Acfg(parser16lex).NewAtom($1.Val)
    }
    | PARSER16_TOK_THIS
    {
        xtracer.Trace("parser.p_typesymbol_this ENTER (typesymbol)")
        a := parser16Acfg(parser16lex).NewAtom("this")
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    ;

optfinite:
    /* empty */
    {
        xtracer.Trace("parser.p_optfinite ENTER (optfinite)")
        $$ = false
    }
    | PARSER16_TOK_FINITE
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
    | PARSER16_TOK_GHOST
    {
        xtracer.Trace("parser.p_optghost_ghost ENTER (optghost)")
        $$ = true
    }
    ;

sort:
    PARSER16_TOK_LCB SYMBOLx PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_symbol_rcb ENTER (sort)")
        $$ = parser16Acfg(parser16lex).NewEnumeratedSort(parser16Acfg(parser16lex).NewAtom($2.Val))
    }
    | PARSER16_TOK_LCB SYMBOLx PARSER16_TOK_COMMA names PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_names_rcb ENTER (sort)")
        vals := []Node{parser16Acfg(parser16lex).NewAtom($2.Val)}
        for _, n := range $4 {
            vals = append(vals, n)
        }
        $$ = parser16Acfg(parser16lex).NewEnumeratedSort(vals...)
    }
    | PARSER16_TOK_LCB SYMBOLx PARSER16_TOK_DOTS SYMBOLx PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_symbol_dots_symbol_rcb ENTER (sort)")
        $$ = parser16Acfg(parser16lex).NewRange(parser16Acfg(parser16lex).NewAtom($2.Val), parser16Acfg(parser16lex).NewAtom($4.Val))
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | PARSER16_TOK_STRUCT PARSER16_TOK_LCB tterms PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sort_struct_lcb_names_rcb ENTER (sort)")
        $$ = parser16Acfg(parser16lex).NewStructSort($3...)
    }
    | PARSER16_TOK_STRUCT PARSER16_TOK_LCB PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sort_struct_lcb_rcb ENTER (sort)")
        $$ = parser16Acfg(parser16lex).NewStructSort()
    }
    ;

names:
    SYMBOLx
    {
        xtracer.Trace("parser.p_names_symbol ENTER (names)")
        $$ = []Node{parser16Acfg(parser16lex).NewAtom($1.Val)}
    }
    | names PARSER16_TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_names_names_comma_symbol ENTER (names)")
        $$ = append($1, parser16Acfg(parser16lex).NewAtom($3.Val))
    }
    ;

// ============================================================
// --- relop / infix ---
// ============================================================

relop:
    PARSER16_TOK_EQ     { xtracer.Trace("parser.p_relop_eq ENTER (relop)"); $$ = "=" }
    | PARSER16_TOK_LE   { xtracer.Trace("parser.p_relop_le ENTER (relop)"); $$ = "<=" }
    | PARSER16_TOK_LT   { xtracer.Trace("parser.p_relop_lt ENTER (relop)"); $$ = "<" }
    | PARSER16_TOK_GE   { xtracer.Trace("parser.p_relop_ge ENTER (relop)"); $$ = ">=" }
    | PARSER16_TOK_GT   { xtracer.Trace("parser.p_relop_gt ENTER (relop)"); $$ = ">" }
    | PARSER16_TOK_PTO  { xtracer.Trace("parser.p_relop_pto ENTER (relop)"); $$ = "*>" }
    ;

infix:
    PARSER16_TOK_PLUS   { xtracer.Trace("parser.p_infix_plus ENTER (infix)"); $$ = "+" }
    | PARSER16_TOK_MINUS { xtracer.Trace("parser.p_infix_minus ENTER (infix)"); $$ = "-" }
    | PARSER16_TOK_TIMES { xtracer.Trace("parser.p_infix_times ENTER (infix)"); $$ = "*" }
    | PARSER16_TOK_DIV   { xtracer.Trace("parser.p_infix_div ENTER (infix)"); $$ = "/" }
    ;

// ============================================================
// --- atom / atoms / app / apps / lit ---
// ============================================================

atom:
    SYMBOLx
    {
        xtracer.Trace("parser.p_atom_symbol ENTER (atom)")
        a := parser16Acfg(parser16lex).NewAtom($1.Val)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | SYMBOLx PARSER16_TOK_LPAREN terms PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_atom_symbol_lp_terms_rp ENTER (atom)")
        a := parser16Acfg(parser16lex).NewAtom($1.Val, $3...)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    ;

atoms:
    atom
    {
        xtracer.Trace("parser.p_atoms_atom ENTER (atoms)")
        $$ = []Node{$1}
    }
    | atoms PARSER16_TOK_COMMA atom
    {
        xtracer.Trace("parser.p_atoms_atoms_atom ENTER (atoms)")
        $$ = append($1, $3)
    }
    ;

app:
    SYMBOLx
    {
        xtracer.Trace("parser.p_app_symbol ENTER (app)")
        $$ = parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1.Val, nil))
    }
    | SYMBOLx PARSER16_TOK_LPAREN terms PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_app_symbol_lp_terms_rp ENTER (app)")
        $$ = parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1.Val, nil), $3...)
    }
    | term infix term
    {
        xtracer.Trace("parser.p_app_term_infix_term ENTER (app)")
        $$ = parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($2, nil), $1, $3)
    }
    ;

apps:
    app
    {
        xtracer.Trace("parser.p_apps_app ENTER (apps)")
        $$ = []Node{$1}
    }
    | apps PARSER16_TOK_COMMA app
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
        $$ = parser16Acfg(parser16lex).NewLiteral(1, $1)
        $$.SetLineno(parser16NodeLineno($1))
    }
    | SYMBOLx PARSER16_TOK_EQ SYMBOLx
    {
        xtracer.Trace("parser.p_lit_term_eq_term ENTER (lit)")
        // Python: p[0] = Literal(1, Atom(p[2], [symbol(p[1]), symbol(p[3])]))
        a := parser16Acfg(parser16lex).NewAtom("=", parser16Acfg(parser16lex).NewAtom($1.Val), parser16Acfg(parser16lex).NewAtom($3.Val))
        $$ = parser16Acfg(parser16lex).NewLiteral(1, a)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | SYMBOLx PARSER16_TOK_TILDAEQ SYMBOLx
    {
        xtracer.Trace("parser.p_lit_term_tildaeq_term ENTER (lit)")
        // Python: p[0] = Literal(0, Atom(p[2], [symbol(p[1]), symbol(p[3])]))
        a := parser16Acfg(parser16lex).NewAtom("=", parser16Acfg(parser16lex).NewAtom($1.Val), parser16Acfg(parser16lex).NewAtom($3.Val))
        $$ = parser16Acfg(parser16lex).NewLiteral(0, a)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | PARSER16_TOK_TILDA lit
    {
        xtracer.Trace("parser.p_lit_tilda_atom ENTER (lit)")
        // Python: p[0] = ~p[2] — flips Literal polarity
        if lit, ok := $2.(*Literal); ok {
            $$ = parser16Acfg(parser16lex).NewLiteral(1 - lit.Polarity, lit.Atom)
        } else {
            $$ = parser16Acfg(parser16lex).NewLiteral(0, $2)
        }
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
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
    | PARSER16_TOK_THIS
    {
        xtracer.Trace("parser.p_callatom_this ENTER (callatom)")
        a := parser16Acfg(parser16lex).NewAtom("this")
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_METHOD
    {
        xtracer.Trace("parser.p_callatom_method ENTER (callatom)")
        $$ = parser16Acfg(parser16lex).NewAtom("method")
    }
    | callatom PARSER16_TOK_DOT callatom
    {
        xtracer.Trace("parser.p_callatom_callatom_dot_callatom ENTER (callatom)")
        lhs := $1.(*Atom)
        rhs := $3.(*Atom)
        $$ = ComposeAtoms(lhs, rhs)
        $$.SetLineno(parser16NodeLineno($1))
    }
    ;

callatoms:
    callatom
    {
        xtracer.Trace("parser.p_callatoms_callatom ENTER (callatoms)")
        $$ = []Node{$1}
    }
    | callatoms PARSER16_TOK_COMMA callatom
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
        parser16lex.(*parser16LexAdapter).accum.isModule = true
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
        parser16lex.(*parser16LexAdapter).accum.isObject = false
        $$ = nil
    }
    ;

modcat:
    /* empty */
    {
        xtracer.Trace("parser.p_modcat ENTER (modcat)")
        $$ = nil
    }
    | PARSER16_TOK_OBJECT
    {
        xtracer.Trace("parser.p_modcat_object ENTER (modcat)")
        $$ = parser16Acfg(parser16lex).NewAtom("object")
    }
    | PARSER16_TOK_ISOLATE
    {
        xtracer.Trace("parser.p_modcat_isolate ENTER (modcat)")
        $$ = parser16Acfg(parser16lex).NewAtom("isolate")
    }
    ;

opteq:
    /* empty */
    {
        xtracer.Trace("parser.p_opteq ENTER (opteq)")
        $$ = nil
    }
    | PARSER16_TOK_EQ
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
        parser16lex.(*parser16LexAdapter).parentObjName = ""
        $$ = false
    }
    | PARSER16_TOK_DOTDOTDOT
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
        parser16lex.(*parser16LexAdapter).accum.params = $1
    }
    ;

objsym:
    SYMBOLx
    {
        xtracer.Trace("parser.p_objsym ENTER (objsym)")
        lex := parser16lex.(*parser16LexAdapter)
        $$ = parser16Acfg(parser16lex).NewAtom($1.Val)
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
    | PARSER16_TOK_TRUSTED
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
    | PARSER16_TOK_LPAREN lparams PARSER16_TOK_RPAREN
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
    | PARSER16_TOK_RETURNS PARSER16_TOK_LPAREN lparams PARSER16_TOK_RPAREN
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
    | callatoms PARSER16_TOK_ASSIGN
    {
        xtracer.Trace("parser.p_optactualreturns_callatoms_assign ENTER (optactualreturns)")
        $$ = $1
    }
    ;

param:
    SYMBOLx PARSER16_TOK_COLON SYMBOLx
    {
        xtracer.Trace("parser.p_param_term_colon_symbol ENTER (param)")
        a := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1.Val, nil))
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        a.ASort = parser16Acfg(parser16lex).NewSymbol($3.Val, nil)
        $$ = a
    }
    ;

params:
    param
    {
        xtracer.Trace("parser.p_params_param ENTER (params)")
        $$ = []Node{$1}
    }
    | params PARSER16_TOK_COMMA param
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
    | PARSER16_TOK_WITH callatoms
    {
        xtracer.Trace("parser.p_optwith_with_callatoms ENTER (optwith)")
        $$ = $2
    }
    ;

// ============================================================
// --- lparam / lparams ---
// ============================================================

lparam:
    SYMBOLx PARSER16_TOK_COLON atype
    {
        xtracer.Trace("parser.p_lparam_variable_colon_symbol ENTER (lparam)")
        // Python: p[0] = App(p[1]); p[0].sort = p[3]
        a := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1.Val, nil))
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        a.ASort = $3
        $$ = a
    }
    | PARSER16_TOK_CARET SYMBOLx PARSER16_TOK_COLON atype
    {
        xtracer.Trace("parser.p_lparam_caret_variable_colon_symbol ENTER (lparam)")
        // Python: p[0] = KeyArg(p[2]); p[0].sort = p[4]
        a := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($2.Val, nil))
        a.ASort = $4
        $$ = a
    }
    ;

lparams:
    lparam
    {
        xtracer.Trace("parser.p_lparams_lparam ENTER (lparams)")
        $$ = []Node{$1}
    }
    | lparams PARSER16_TOK_COMMA lparam
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
        $$ = parser16Acfg(parser16lex).NewSequence()
    }
    | PARSER16_TOK_EQ topseq
    {
        xtracer.Trace("parser.p_optactiondef_eq_topseq ENTER (optactiondef)")
        $$ = $2
    }
    | PARSER16_TOK_EQ PARSER16_TOK_TIMES
    {
        xtracer.Trace("parser.p_optactiondef_eq_symbol ENTER (optactiondef)")
        $$ = parser16Acfg(parser16lex).NewCrashAction()
    }
    ;

topseq:
    sequence
    {
        xtracer.Trace("parser.p_topseq_sequence ENTER (topseq)")
        $$ = $1
    }
    | PARSER16_TOK_LCB PARSER16_TOK_NATIVEQUOTE PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_topseq_lcb_nativequote_rcb ENTER (topseq)")
        // Python: NativeAction(*([text] + bqs))
        text, bqs := parser16ParseNativequote(parser16Acfg(parser16lex), $2.Val, parser16lex.(*parser16LexAdapter))
        args := append([]Node{parser16Acfg(parser16lex).NewNativeCode(text)}, bqs...)
        na := parser16Acfg(parser16lex).NewNativeAction(args...)
        na.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = na
    }
    ;

optimpex:
    /* empty */
    {
        xtracer.Trace("parser.p_optimpex ENTER (optimpex)")
        $$ = nil
    }
    | PARSER16_TOK_EXPORT
    {
        xtracer.Trace("parser.p_optimpex_export ENTER (optimpex)")
        $$ = parser16Acfg(parser16lex).NewExportDecl()
    }
    | PARSER16_TOK_IMPORT
    {
        xtracer.Trace("parser.p_optimpex_import ENTER (optimpex)")
        $$ = parser16Acfg(parser16lex).NewImportDecl()
    }
    ;

actmeth:
    PARSER16_TOK_ACTION
    {
        xtracer.Trace("parser.p_actmeth_action ENTER (actmeth)")
        $$ = false
    }
    | PARSER16_TOK_METHOD
    {
        xtracer.Trace("parser.p_actmeth_method ENTER (actmeth)")
        $$ = true
    }
    ;

// ============================================================
// --- specimpl ---
// ============================================================

specimpl:
    PARSER16_TOK_SPECIFICATION
    {
        xtracer.Trace("parser.p_specimpl_specification ENTER (specimpl)")
        $$ = "spec"
        // Python: global special_attribute; special_attribute = "spec"
        parser16lex.(*parser16LexAdapter).specialAttribute = "spec"
    }
    | PARSER16_TOK_IMPLEMENTATION
    {
        xtracer.Trace("parser.p_specimpl_implementation ENTER (specimpl)")
        $$ = "impl"
        // Python: global special_attribute; special_attribute = "impl"
        parser16lex.(*parser16LexAdapter).specialAttribute = "impl"
    }
    | PARSER16_TOK_PRIVATE
    {
        xtracer.Trace("parser.p_specimpl_private ENTER (specimpl)")
        $$ = "private"
        // Python: global special_attribute; special_attribute = "private"
        parser16lex.(*parser16LexAdapter).specialAttribute = "private"
    }
    | PARSER16_TOK_GLOBAL
    {
        xtracer.Trace("parser.p_specimpl_global ENTER (specimpl)")
        $$ = "global"
        // Python: global global_attribute; global_attribute = "global"
        parser16lex.(*parser16LexAdapter).globalAttribute = "global"
    }
    | PARSER16_TOK_COMMON
    {
        xtracer.Trace("parser.p_specimpl_common ENTER (specimpl)")
        $$ = "common"
        // Python: global common_attribute; common_attribute = "common"
        parser16lex.(*parser16LexAdapter).commonAttribute = "common"
    }
    ;

// ============================================================
// --- Instantiate ---
// ============================================================

insts:
    inst
    {
        xtracer.Trace("parser.p_insts_inst ENTER (insts)")
        $$ = []Node{$1}
    }
    | insts PARSER16_TOK_COMMA inst
    {
        xtracer.Trace("parser.p_insts_insts_comma_inst ENTER (insts)")
        $$ = append($1, $3)
    }
    ;

inst:
    modinst
    {
        xtracer.Trace("parser.p_inst_modinst ENTER (inst)")
        n := parser16Acfg(parser16lex).NewInstantiation(nil, AppToAtom($1))
        n.SetLineno(parser16NodeLineno($1))
        $$ = n
    }
    | modinst PARSER16_TOK_COLON modinst
    {
        xtracer.Trace("parser.p_inst_atom_colon_modinst ENTER (inst)")
        inst := parser16Acfg(parser16lex).NewInstantiation(AppToAtom($1), AppToAtom($3))
        inst.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = inst
    }
    ;

modinst:
    dotsym
    {
        xtracer.Trace("parser.p_modinst_symbol ENTER (modinst)")
        $$ = parser16Acfg(parser16lex).NewAtom($1)
    }
    | dotsym PARSER16_TOK_LPAREN pnames PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_modinst_symbol_lp_pnames_rp ENTER (modinst)")
        a := parser16Acfg(parser16lex).NewAtom($1)
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
        case *Symbol:
            rep = v.Rep
        case *This:
            rep = "this"
        default:
            rep = fmt.Sprint($1)
        }
        n := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol(rep, nil))
        n.SetLineno(parser16NodeLineno($1))
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
        $$ = parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1, nil))
        $$.SetLineno(parser16GetLineno(parser16lex.(*parser16LexAdapter)))
    }
    | relop
    {
        xtracer.Trace("parser.p_pname_relop ENTER (pname)")
        $$ = parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1, nil))
        $$.SetLineno(parser16GetLineno(parser16lex.(*parser16LexAdapter)))
    }
    | PARSER16_TOK_THIS
    {
        xtracer.Trace("parser.p_pname_this ENTER (pname)")
        $$ = parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("this", nil))
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_TRUE
    {
        xtracer.Trace("parser.p_pname_true ENTER (pname)")
        $$ = parser16Acfg(parser16lex).NewAtom("true")
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_FALSE
    {
        xtracer.Trace("parser.p_pname_false ENTER (pname)")
        $$ = parser16Acfg(parser16lex).NewAtom("false")
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
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
        $$ = []Node{$1}
    }
    | pnames PARSER16_TOK_COMMA pname
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
        if sym, ok := $1.(*Symbol); ok {
            $$ = parser16Acfg(parser16lex).NewAtom(sym.Rep)
        } else if th, ok := $1.(*This); ok {
            a := parser16Acfg(parser16lex).NewAtom("this")
            a.SetLineno(th.GetLineno())
            $$ = a
        } else {
            $$ = $1
        }
    }
    | relop
    {
        xtracer.Trace("parser.p_oper_relop ENTER (oper)")
        $$ = parser16Acfg(parser16lex).NewAtom($1)
    }
    | infix
    {
        xtracer.Trace("parser.p_oper_infix ENTER (oper)")
        $$ = parser16Acfg(parser16lex).NewAtom($1)
    }
    | PARSER16_TOK_NATIVEQUOTE
    {
        xtracer.Trace("parser.p_oper_nativequote ENTER (oper)")
        text, bqs := parser16ParseNativequote(parser16Acfg(parser16lex), $1.Val, parser16lex.(*parser16LexAdapter))
        elems := append([]Node{parser16Acfg(parser16lex).NewAtom(text)}, bqs...)
        nt := parser16Acfg(parser16lex).NewNativeType(elems...)
        nt.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = nt
    }
    ;

attributeval:
    callatom
    {
        xtracer.Trace("parser.p_top_attributeval_callatom ENTER (attributeval)")
        $$ = $1
    }
    | PARSER16_TOK_TRUE
    {
        xtracer.Trace("parser.p_top_attributeval_true ENTER (attributeval)")
        $$ = parser16Acfg(parser16lex).NewAtom("true")
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_FALSE
    {
        xtracer.Trace("parser.p_top_attributeval_false ENTER (attributeval)")
        $$ = parser16Acfg(parser16lex).NewAtom("false")
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

moresymbols:
    /* empty */
    {
        xtracer.Trace("parser.p_moresymbols ENTER (moresymbols)")
        $$ = nil
    }
    | moresymbols PARSER16_TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_moresymbols_more_symbols_comma_symbol ENTER (moresymbols)")
        $$ = append($1, parser16Acfg(parser16lex).NewAtom($3.Val))
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
    | PARSER16_TOK_ARROW callatom
    {
        xtracer.Trace("parser.p_optdelegee_callatom ENTER (optdelegee)")
        $$ = $2
    }
    ;

// ============================================================
// --- sequence / actseq / action ---
// ============================================================

sequence:
    PARSER16_TOK_LCB PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_rcb ENTER (sequence)")
        $$ = parser16Acfg(parser16lex).NewSequence()
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_LCB actseq PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_actseq_rcb ENTER (sequence)")
        stmts := parser16LowerVarStmts($2)
        seq := parser16MakeSequence(parser16Acfg(parser16lex), stmts)
        if s, ok := seq.(*Sequence); ok {
            s.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        } else {
            // Single node — mark as having a location without calling parser16GetLineno
            // (Python always has lineno set on these nodes from their own rules)
            if b, ok := seq.(interface{ HasLocSet() bool }); ok && !b.HasLocSet() {
                loc := parser16lex.(*parser16LexAdapter).lastTok.Line
                seq.SetLineno(Location{Line: loc})
            }
        }
        $$ = seq
    }
    | PARSER16_TOK_LCB actseq PARSER16_TOK_SEMI PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_actseq_semi_rcb ENTER (sequence)")
        // Python: p[0] = Sequence(*lower_var_stmts(p[2]))
        // Unlike p_sequence_lcb_actseq_rcb, this rule always wraps in Sequence
        // and only calls lower_var_stmts once (no len==1 shortcut).
        stmts := parser16LowerVarStmts($2)
        seq := parser16Acfg(parser16lex).NewSequence(stmts...)
        seq.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
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
        $$ = []Node{$1}
    }
    | complexact
    {
        xtracer.Trace("parser.p_actseqrev_complexact ENTER (actseqrev)")
        $$ = []Node{$1}
    }
    | simpleact PARSER16_TOK_SEMI actseqrev
    {
        xtracer.Trace("parser.p_actseqrev_simpact_semi_actseqrev ENTER (actseqrev)")
        $$ = append($3, $1)
    }
    | simpleact PARSER16_TOK_SEMI
    {
        xtracer.Trace("parser.p_actseqrev_simpact_semi ENTER (actseqrev)")
        $$ = []Node{$1}
    }
    | complexact actseqrev
    {
        xtracer.Trace("parser.p_actseqrev_complexact_actseqrev ENTER (actseqrev)")
        $$ = append($2, $1)
    }
    | complexact PARSER16_TOK_SEMI actseqrev
    {
        xtracer.Trace("parser.p_actseqrev_complexact_semi_actseqrev ENTER (actseqrev)")
        $$ = append($3, $1)
    }
    | complexact PARSER16_TOK_SEMI
    {
        xtracer.Trace("parser.p_actseqrev_complexact_semi ENTER (actseqrev)")
        $$ = []Node{$1}
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
    PARSER16_TOK_ASSUME labeledfmla
    {
        xtracer.Trace("parser.p_action_assume ENTER (simpleact)")
        // Python: AssumeAction(check_non_temporal(addlabel(p[2],'asrt')))
        lf := parser16AddLabel(parser16Acfg(parser16lex), $2.(*LabeledFormula), "asrt")
        a := parser16Acfg(parser16lex).NewAssumeAction(parser16CheckNonTemporal(lf))
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_ASSERT labeledfmla
    {
        xtracer.Trace("parser.p_action_assert ENTER (simpleact)")
        lf := parser16AddLabel(parser16Acfg(parser16lex), $2.(*LabeledFormula), "asrt")
        lf = parser16CheckNonTemporal(lf).(*LabeledFormula)
        a := parser16Acfg(parser16lex).NewAssertAction(lf)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_ENSURES labeledfmla
    {
        xtracer.Trace("parser.p_action_ensures ENTER (simpleact)")
        lf := parser16AddLabel(parser16Acfg(parser16lex), $2.(*LabeledFormula), "asrt")
        lf = parser16CheckNonTemporal(lf).(*LabeledFormula)
        a := parser16Acfg(parser16lex).NewEnsuresAction(lf)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | term PARSER16_TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_action_term_assign_fmla ENTER (simpleact)")
        // Python: AssignAction(p[1], check_non_temporal(p[3]))
        a := parser16Acfg(parser16lex).NewAssignAction($1, parser16CheckNonTemporal($3))
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = a
    }
    | termtuple PARSER16_TOK_ASSIGN callatom
    {
        xtracer.Trace("parser.p_action_termtuple_assign_fmla ENTER (simpleact)")
        // Python: CallAction(*([p[3]]+list(p[1].args)))
        callArgs := append([]Node{$3}, $1.Args()...)
        $$ = parser16Acfg(parser16lex).NewCallAction(callArgs...)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
    }
    | term PARSER16_TOK_ASSIGN PARSER16_TOK_TIMES
    {
        xtracer.Trace("parser.p_action_term_assign_times ENTER (simpleact)")
        // Python: HavocAction(p[1])
        a := parser16Acfg(parser16lex).NewHavocAction($1)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = a
    }
    | PARSER16_TOK_VAR tterm optinit
    {
        // Python: VarAction(p[2]) or VarAction(p[2], p[3])
        xtracer.Trace("parser.p_action_var_opttypedsym_assign_fmla ENTER (simpleact)")
        if $3 != nil {
            $$ = parser16Acfg(parser16lex).NewVarAction($2, $3)
        } else {
            $$ = parser16Acfg(parser16lex).NewVarAction($2)
        }
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_CALL optactualreturns callatom
    {
        xtracer.Trace("parser.p_action_call_optreturns_callatom ENTER (simpleact)")
        // Python: CallAction(*([p[3]] + p[2]))
        callArgs := append([]Node{$3}, $2...)
        $$ = parser16Acfg(parser16lex).NewCallAction(callArgs...)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_CALL callatom
    {
        xtracer.Trace("parser.p_action_call_callatom ENTER (simpleact)")
        // Python: CallAction(p[2])
        $$ = parser16Acfg(parser16lex).NewCallAction($2)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_SET lit
    {
        xtracer.Trace("parser.p_action_set_lit ENTER (simpleact)")
        // Python: SetAction(p[2])
        a := parser16Acfg(parser16lex).NewSetAction($2)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_INSTANTIATE callatom
    {
        xtracer.Trace("parser.p_action_instantiate_atom ENTER (simpleact)")
        // Python: InstantiateAction(p[2])
        a := parser16Acfg(parser16lex).NewInstantiateAction($2)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | PARSER16_TOK_DEBUG SYMBOLx optdebugargs
    {
        xtracer.Trace("parser.p_simpleact_debug_symbol_optdebugargs ENTER (simpleact)")
        // Python: action = Atom(p[2],[]); action.lineno = get_lineno(p,2)
        // Python: if not p[2].startswith('"'): report_error(...)
        // Python: p[0] = DebugAction(action,*p[3]); p[0].lineno = get_lineno(p,1)
        action := parser16Acfg(parser16lex).NewAtom($2.Val)
        action.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        if !strings.HasPrefix($2.Val, "\"") {
            parser16lex.Error(fmt.Sprintf("expected string constant after 'debug', got %s", $2.Val))
        }
        args := append([]Node{action}, $3...)
        a := parser16Acfg(parser16lex).NewDebugAction(args...)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = a
    }
    | term     %prec PARSER16_TOK_SEMI
    {
        xtracer.Trace("parser.p_action_term ENTER (simpleact)")
        // Python: p[0] = CallAction(p[1]); p[0].lineno = p[1].lineno
        $$ = parser16Acfg(parser16lex).NewCallAction($1)
        $$.SetLineno($1.GetLineno())
    }
    ;

termtuple:
    PARSER16_TOK_LPAREN term PARSER16_TOK_COMMA terms PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_termtuple_lp_term_comma_terms_rp ENTER (termtuple)")
        args := append([]Node{$2}, $4...)
        t := parser16Acfg(parser16lex).NewTuple(args...)
        t.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = t
    }
    ;

debugarg:
    SYMBOLx PARSER16_TOK_EQ fmla
    {
        xtracer.Trace("parser.p_debugarg_symbol_equal_fmla ENTER (debugarg)")
        // Python: lhs = App(p[1]); p[0] = DebugItem(lhs, p[3])
        lhs := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1.Val, nil))
        lhs.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        di := parser16Acfg(parser16lex).NewDebugItem(lhs, $3)
        di.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = di
    }
    ;

debugargs:
    debugarg
    {
        xtracer.Trace("parser.p_debugargs ENTER (debugargs)")
        $$ = []Node{$1}
    }
    | debugargs PARSER16_TOK_COMMA debugarg
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
    | PARSER16_TOK_WITH debugargs
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
    | PARSER16_TOK_IF somefmla sequence
    {
        xtracer.Trace("parser.p_action_if_somefmla_lcb_action_rcb ENTER (complexact)")
        // Python: IfAction(cond, fix_if_part(cond, p[3]))
        cond := parser16CheckNonTemporal($2)
        body := parser16FixIfPart(cond, $3)
        ifa := parser16Acfg(parser16lex).NewIfAction(cond, body, nil)
        ifa.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = ifa
    }
    | PARSER16_TOK_IF somefmla sequence PARSER16_TOK_ELSE action
    {
        xtracer.Trace("parser.p_action_if_somefmla_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
        // Python: IfAction(cond, fix_if_part(cond, p[3]), p[5])
        cond := parser16CheckNonTemporal($2)
        body := parser16FixIfPart(cond, $3)
        ifa := parser16Acfg(parser16lex).NewIfAction(cond, body, $5)
        ifa.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = ifa
    }
    | PARSER16_TOK_IF PARSER16_TOK_TIMES sequence PARSER16_TOK_ELSE action
    {
        xtracer.Trace("parser.p_action_if_times_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
        // Python: ChoiceAction(p[3], p[5])
        choice := parser16Acfg(parser16lex).NewChoiceAction($3, $5)
        choice.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = choice
    }
    | PARSER16_TOK_WHILE somefmla invariants decreases sequence
    {
        xtracer.Trace("parser.p_action_while_somefmla_invariants_decreases_lcb_action_rcb ENTER (complexact)")
        // Python: WhileAction(cond, body, *invariants, *decreases)
        cond := parser16CheckNonTemporal($2)
        body := parser16FixIfPart($2, $5)
        args := []Node{cond, body}
        args = append(args, $3...)
        args = append(args, $4...)
        w := parser16Acfg(parser16lex).NewWhileAction(args...)
        w.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = w
    }
    | PARSER16_TOK_FOR tterm PARSER16_TOK_COMMA tterm PARSER16_TOK_IN fmla invariants decreases sequence
    {
        xtracer.Trace("parser.p_action_for_tterm_comma_tterm_in_expr_invariants_decreases_lcb_action_rcb ENTER (complexact)")

        // Python: itr,val,fmla,invars,decrs,seq = p[2],p[4],check_non_temporal(p[6]),p[7],p[8],p[9]
        itr := $2                       // *App from tterm
        val := $4                       // *App from tterm
        forFmla := parser16CheckNonTemporal($6) // formula
        invars := $7                    // []Node from invariants
        decrs := $8                     // []Node from decreases
        seq := $9                       // Node from sequence

        // Python: iend = itr.rename('loc:end')
        var iend Node
        if itrApp, ok := itr.(*App); ok {
            iend = itrApp.Rename("loc:end")
        } else if itrAtom, ok := itr.(*Atom); ok {
            iend = itrAtom.Rename("loc:end")
        } else {
            iend = itr // fallback
        }

        // Python: ln = get_lineno(p,1)
        ln := parser16TokLineno(parser16lex.(*parser16LexAdapter), $1)

        // Python: didx = VarAction(itr, parser16Methcall(fmla, App('begin').sln(ln)).sln(ln)).sln(ln)
        appBegin := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("begin", nil))
        appBegin.SetLineno(ln)
        mcBegin := parser16Methcall(parser16Acfg(parser16lex), forFmla, appBegin)
        mcBegin.SetLineno(ln)
        didx := parser16Acfg(parser16lex).NewVarAction(itr, mcBegin)
        didx.SetLineno(ln)

        // Python: dend = VarAction(iend, parser16Methcall(fmla, App('end').sln(ln)).sln(ln)).sln(ln)
        appEnd := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("end", nil))
        appEnd.SetLineno(ln)
        mcEnd := parser16Methcall(parser16Acfg(parser16lex), forFmla, appEnd)
        mcEnd.SetLineno(ln)
        dend := parser16Acfg(parser16lex).NewVarAction(iend, mcEnd)
        dend.SetLineno(ln)

        // Python: dval = VarAction(val, parser16Methcall(fmla, App('value', itr).sln(ln)).sln(ln)).sln(ln)
        appValue := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("value", nil), itr)
        appValue.SetLineno(ln)
        mcValue := parser16Methcall(parser16Acfg(parser16lex), forFmla, appValue)
        mcValue.SetLineno(ln)
        dval := parser16Acfg(parser16lex).NewVarAction(val, mcValue)
        dval.SetLineno(ln)

        // Python: incr = AssignAction(itr, parser16Methcall(itr, App('next').sln(ln)).sln(ln)).sln(ln)
        appNext := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("next", nil))
        appNext.SetLineno(ln)
        mcNext := parser16Methcall(parser16Acfg(parser16lex), itr, appNext)
        mcNext.SetLineno(ln)
        incr := parser16Acfg(parser16lex).NewAssignAction(itr, mcNext)
        incr.SetLineno(ln)

        // Python: body = Sequence(*lower_var_stmts([dval, seq, incr])).sln(ln)
        bodyStmts := LowerVarStatements([]Node{dval, seq, incr})
        body := parser16Acfg(parser16lex).NewSequence(bodyStmts...)
        body.SetLineno(ln)

        // Python: loop = WhileAction(*([App('<', itr, iend).sln(ln), body] + invars + decrs)).sln(ln)
        ltCond := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("<", nil), itr, iend)
        ltCond.SetLineno(ln)
        loopArgs := []Node{ltCond, body}
        loopArgs = append(loopArgs, invars...)
        loopArgs = append(loopArgs, decrs...)
        loop := parser16Acfg(parser16lex).NewWhileAction(loopArgs...)
        loop.SetLineno(ln)

        // Python: p[0] = Sequence(*lower_var_stmts([didx, dend, loop])).sln(ln)
        outerStmts := LowerVarStatements([]Node{didx, dend, loop})
        result := parser16Acfg(parser16lex).NewSequence(outerStmts...)
        result.SetLineno(ln)
        $$ = result
    }
    | PARSER16_TOK_LOCAL lparams sequence
    {
        xtracer.Trace("parser.p_action_local_params_lcb_action_rcb ENTER (complexact)")
        // Python: lsyms = [s.prefix('loc:') for s in p[2]]
        // Python: subst = dict((x.rep,y.rep) for x,y in zip(p[2],lsyms))
        // Python: action = subst_prefix_atoms_ast(p[3],subst,None,None)
        // Python: p[0] = LocalAction(*(lsyms+[action]))
        bounds := $2
        lsyms := make([]Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = PrefixNode(s, "loc:")
            subst[NodeRep(s)] = NodeRep(lsyms[i])
        }
        action := SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        args := append(lsyms, action)
        la := parser16Acfg(parser16lex).NewLocalAction("parser.local_action_rule", args...)
        la.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = la
    }
    | PARSER16_TOK_LET eqns sequence
    {
        xtracer.Trace("parser.p_action_let_eqns_lcb_action_rcb ENTER (complexact)")
        // Python: LetAction(*(p[2]+[p[3]]))
        args := append($2, $3)
        la := parser16Acfg(parser16lex).NewLetAction(args...)
        la.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = la
    }
    | PARSER16_TOK_THUNK labelname SYMBOLx optargs PARSER16_TOK_COLON atype PARSER16_TOK_ASSIGN sequence
    {
        xtracer.Trace("parser.p_action_thunk_symbol_optargs_colon_atype_assign_sequence ENTER (complexact)")
        // Python: action = Atom(p[3], p[4]); action.lineno = get_lineno(p,3)
        // Python: ThunkAction(Atom(p[2][1:-1],[]), action, Atom(p[6]), p[8])
        // Brackets already stripped by labelname rule
        label := parser16Acfg(parser16lex).NewAtom($2.Val)
        label.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        actionAtom := parser16Acfg(parser16lex).NewAtom($3.Val, $4...)
        actionAtom.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        sortAtom := parser16Acfg(parser16lex).NewAtom(NodeRep($6))
        ta := parser16Acfg(parser16lex).NewThunkAction(label, actionAtom, sortAtom, $8)
        ta.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
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
    | fmla PARSER16_TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_somefmla_fmla_assign_fmla ENTER (somefmla)")
        // Python: lsyms = [p[1].prefix('loc:')]
        // Python: lsyms[0].sort = p[1].sort
        // Python: subst = dict((x.rep,y.rep) for x,y in zip([p[1]],lsyms))
        // Python: fmla = App('*>',p[3],p[1])
        // Python: fmla = subst_prefix_atoms_ast(fmla,subst,None,None)
        // Python: p[0] = Some(*(lsyms+[fmla]))
        lhs := $1
        lsym := PrefixNode(lhs, "loc:")
        if a, ok := lsym.(*Atom); ok {
            if orig, ok2 := lhs.(*Atom); ok2 {
                a.ASort = orig.ASort
            }
        }
        subst := map[string]string{NodeRep(lhs): NodeRep(lsym)}
        fmla := parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol("*>", nil), $3, lhs)
        fmla.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        fmla2 := SubstPrefixAtomsAst(fmla, subst, nil, nil, nil)
        some := parser16Acfg(parser16lex).NewSome([]Node{lsym}, fmla2)
        some.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = some
    }
    | PARSER16_TOK_SOME bounds fmla
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla ENTER (somefmla)")
        bounds := $2
        lsyms := make([]Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = PrefixNode(s, "loc:")
            subst[NodeRep(s)] = NodeRep(lsyms[i])
        }
        fmla := SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        some := parser16Acfg(parser16lex).NewSome(lsyms, fmla)
        some.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = some
    }
    | PARSER16_TOK_SOME bounds fmla PARSER16_TOK_MINIMIZING term
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla_minimizing_term ENTER (somefmla)")
        bounds := $2
        lsyms := make([]Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = PrefixNode(s, "loc:")
            subst[NodeRep(s)] = NodeRep(lsyms[i])
        }
        fmla := SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        index := SubstPrefixAtomsAst($5, subst, nil, nil, nil)
        smin := parser16Acfg(parser16lex).NewSomeMin(lsyms, fmla, index)
        smin.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = smin
    }
    | PARSER16_TOK_SOME bounds fmla PARSER16_TOK_MAXIMIZING term
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla_maximizing_term ENTER (somefmla)")
        bounds := $2
        lsyms := make([]Node, len(bounds))
        subst := make(map[string]string)
        for i, s := range bounds {
            lsyms[i] = PrefixNode(s, "loc:")
            subst[NodeRep(s)] = NodeRep(lsyms[i])
        }
        fmla := SubstPrefixAtomsAst($3, subst, nil, nil, nil)
        index := SubstPrefixAtomsAst($5, subst, nil, nil, nil)
        smax := parser16Acfg(parser16lex).NewSomeMax(lsyms, fmla, index)
        smax.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = smax
    }
    ;

bounds:
    params PARSER16_TOK_DOT
    {
        xtracer.Trace("parser.p_bounds_params_dot ENTER (bounds)")
        $$ = $1
    }
    | PARSER16_TOK_LPAREN lparams PARSER16_TOK_RPAREN
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
    | invariants PARSER16_TOK_INVARIANT labeledfmla
    {
        xtracer.Trace("parser.p_invariant_invariant_fmla ENTER (invariants)")
        // Python: a = AssertAction(check_non_temporal(addlabel(p[3],'asrt')))
        inv := parser16CheckNonTemporal(parser16AddLabel(parser16Acfg(parser16lex), $3.(*LabeledFormula), "asrt"))
        a := parser16Acfg(parser16lex).NewAssertAction(inv)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = append($1, a)
    }
    | invariants PARSER16_TOK_INVARIANT labeledfmla PARSER16_TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_invariant_invariant_fmla_proof ENTER (invariants)")
        // Python: a = AssertAction(inv, p[5])
        inv := parser16CheckNonTemporal(parser16AddLabel(parser16Acfg(parser16lex), $3.(*LabeledFormula), "asrt"))
        a := parser16Acfg(parser16lex).NewAssertAction(inv, $5)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = append($1, a)
    }
    ;

decreases:
    /* empty */
    {
        xtracer.Trace("parser.p_decreases ENTER (decreases)")
        $$ = nil
    }
    | PARSER16_TOK_DECREASES fmla
    {
        xtracer.Trace("parser.p_decreases_decreases_fmla ENTER (decreases)")
        // Python: rank = Ranking(check_non_temporal(p[2])); rank.lineno = get_lineno(p,1)
        fmla := parser16CheckNonTemporal($2)
        rank := parser16Acfg(parser16lex).NewRanking(fmla)
        rank.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = []Node{rank}
    }
    ;

// --- eqn / eqns ---

eqn:
    SYMBOLx PARSER16_TOK_EQ SYMBOLx
    {
        xtracer.Trace("parser.p_eqn_SYMBOL_EQ_SYMBOL ENTER (eqn)")
        // Python: Equals(App(p[1]), App(p[3])) — Equals = lg.Eq, AST equivalent is Atom("=", lhs, rhs)
        $$ = parser16Acfg(parser16lex).NewAtom("=", parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($1.Val, nil)), parser16Acfg(parser16lex).NewApp(parser16Acfg(parser16lex).NewSymbol($3.Val, nil)))
    }
    ;

eqns:
    eqn
    {
        xtracer.Trace("parser.p_eqns_eqn ENTER (eqns)")
        $$ = []Node{$1}
    }
    | eqns PARSER16_TOK_COMMA eqn
    {
        xtracer.Trace("parser.p_eqns_eqns_comma_eqn ENTER (eqns)")
        $$ = append($1, $3)
    }
    ;

// ============================================================
// --- Scenario ---
// ============================================================

sceninit:
    PARSER16_TOK_ARROW places
    {
        xtracer.Trace("parser.p_sceninit_arrow_places ENTER (sceninit)")
        pl := parser16Acfg(parser16lex).NewPlaceList($2)
        pl.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = pl
    }
    ;

places:
    SYMBOLx
    {
        xtracer.Trace("parser.p_places_symbol ENTER (places)")
        a := parser16Acfg(parser16lex).NewAtom($1.Val)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = []Node{a}
    }
    | places PARSER16_TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_places_places_comma_symbol ENTER (places)")
        a := parser16Acfg(parser16lex).NewAtom($3.Val)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
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
    | scentranss places PARSER16_TOK_ARROW places PARSER16_TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentranss_scentranss_places_arrow_places_colon_scenariomixin ENTER (scentranss)")
        from := parser16Acfg(parser16lex).NewPlaceList($2)
        to := parser16Acfg(parser16lex).NewPlaceList($4)
        tr := parser16Acfg(parser16lex).NewScenarioTransition(from, to, $6)
        tr.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $5))
        $$ = append($1, tr)
    }
    | scentranss places PARSER16_TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentranss_scentranss_places_colon_scenariomixin ENTER (scentranss)")
        to := parser16Acfg(parser16lex).NewPlaceList($2)
        tr := parser16Acfg(parser16lex).NewScenarioTransition(nil, to, $4)
        tr.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        $$ = append($1, tr)
    }
    ;

scentrans:
    places PARSER16_TOK_ARROW places PARSER16_TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentrans__places_arrow_places_colon_scenariomixin ENTER (scentrans)")
        from := parser16Acfg(parser16lex).NewPlaceList($1)
        to := parser16Acfg(parser16lex).NewPlaceList($3)
        $$ = parser16Acfg(parser16lex).NewScenarioTransition(from, to, $5)
    }
    | places PARSER16_TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentrans__places_colon_scenariomixin ENTER (scentrans)")
        from := parser16Acfg(parser16lex).NewPlaceList($1)
        $$ = parser16Acfg(parser16lex).NewScenarioTransition(from, parser16Acfg(parser16lex).NewPlaceList(nil), $3)
    }
    ;

scenariomixin:
    PARSER16_TOK_BEFORE atype optargs optreturns sequence
    {
        xtracer.Trace("parser.p_scenariomixin_before_callatom_lcb_action_rcb ENTER (scenariomixin)")
        atom := parser16Acfg(parser16lex).NewAtom($2.(*Symbol).Rep)
        atom.SetLineno(parser16NodeLineno($2))
        mixer := parser16MakeMixinName(parser16Acfg(parser16lex), atom, "before")
        // Python: optargs, optreturns = infer_action_params(atom.rep, p[3], p[4])
        formals, returns := parser16InferActionParams(parser16lex.(*parser16LexAdapter).accum, atom.Rep, $3, $4)
        adef := parser16Acfg(parser16lex).NewActionDef(atom, $5, formals, returns)
        sbm := parser16Acfg(parser16lex).NewScenarioBeforeMixin(mixer, adef)
        sbm.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = sbm
    }
    | PARSER16_TOK_AFTER atype optargs optreturns sequence
    {
        xtracer.Trace("parser.p_scenariomixin_after_callatom_lcb_action_rcb ENTER (scenariomixin)")
        atom := parser16Acfg(parser16lex).NewAtom($2.(*Symbol).Rep)
        atom.SetLineno(parser16NodeLineno($2))
        mixer := parser16MakeMixinName(parser16Acfg(parser16lex), atom, "after")
        // Python: optargs, optreturns = infer_action_params(atom.rep, p[3], p[4])
        formals, returns := parser16InferActionParams(parser16lex.(*parser16LexAdapter).accum, atom.Rep, $3, $4)
        adef := parser16Acfg(parser16lex).NewActionDef(atom, $5, formals, returns)
        sam := parser16Acfg(parser16lex).NewScenarioAfterMixin(mixer, adef)
        sam.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = sam
    }
    ;

// ============================================================
// --- Proof / tactic ---
// ============================================================

pflet:
    var PARSER16_TOK_EQ fmla
    {
        xtracer.Trace("parser.p_pflet_var_eq_fmla ENTER (pflet)")
        d := parser16Acfg(parser16lex).NewDefinition($1, $3)
        d.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = d
    }
    ;

pflets:
    pflet
    {
        xtracer.Trace("parser.p_pflets_pflet ENTER (pflets)")
        $$ = []Node{$1}
    }
    | pflets PARSER16_TOK_COMMA pflet
    {
        xtracer.Trace("parser.p_pflets_pflets_pflet ENTER (pflets)")
        $$ = append($1, $3)
    }
    ;

tacticwithelem:
    PARSER16_TOK_INVARIANT labeledfmla
    {
        xtracer.Trace("parser.p_tacticwithelem_invariant ENTER (tacticwithelem)")
        // Python: p[0] = addlabel(p[2], 'invar')
        $$ = parser16AddLabel(parser16Acfg(parser16lex), $2.(*LabeledFormula), "invar")
    }
    | PARSER16_TOK_DEFINITION typeddefn PARSER16_TOK_EQ fmla
    {
        xtracer.Trace("parser.p_tacticwithelem_fun_defn ENTER (tacticwithelem)")
        df := parser16Acfg(parser16lex).NewDefinition(AppToAtom($2), $4)
        df.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        lf := parser16AddLabel(parser16Acfg(parser16lex), parser16MkLF(parser16Acfg(parser16lex), df), "def")
        $$ = parser16Acfg(parser16lex).NewDerivedDecl(lf)
    }
    | PARSER16_TOK_TRIGGER atype PARSER16_TOK_WITH terms
    {
        xtracer.Trace("parser.p_tacticwithelem_trigger ENTER (tacticwithelem)")
        // Python: p[0] = Trigger(*([Atom(p[2])]+p[4])); p[0].lineno = get_lineno(p,3)
        trigAtom := parser16AtypeToAtom(parser16Acfg(parser16lex), $2)
        trig := parser16Acfg(parser16lex).NewTrigger(trigAtom, $4...)
        trig.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $3))
        $$ = trig
    }
    ;

tacticwithlist:
    tacticwithelem
    {
        xtracer.Trace("parser.p_tactwithlist_tacticwithelem ENTER (tacticwithlist)")
        $$ = []Node{$1}
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
        $$ = parser16Acfg(parser16lex).NewTacticWith($1)
    }
    | pflets
    {
        xtracer.Trace("parser.p_tacticwithlistchoice_pflets ENTER (tacticwithlistchoice)")
        $$ = parser16Acfg(parser16lex).NewTacticLets($1)
    }
    ;

opttacticwith:
    /* empty */
    {
        xtracer.Trace("parser.p_opttacticwith ENTER (opttacticwith)")
        $$ = parser16Acfg(parser16lex).NewTacticWith(nil)
    }
    | PARSER16_TOK_WITH tacticwithlistchoice
    {
        xtracer.Trace("parser.p_opttacticwith_with_tacticwithlist ENTER (opttacticwith)")
        $$ = $2
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    | PARSER16_TOK_WITH PARSER16_TOK_LCB tacticwithlist PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_opttacticwith_with_lcb_tacticwithlist_rcb ENTER (opttacticwith)")
        tw := parser16Acfg(parser16lex).NewTacticWith($3)
        tw.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = tw
    }
    ;

proofgroup:
    PARSER16_TOK_LCB proofseq PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_proofgroup_lcb_proofseq_rcb ENTER (proofgroup)")
        $$ = $2
    }
    | PARSER16_TOK_LCB PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_proofgroup_lcb_rcb ENTER (proofgroup)")
        $$ = parser16Acfg(parser16lex).NewNullTactic()
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
    }
    ;

optproofgroup:
    /* empty */
    {
        xtracer.Trace("parser.p_optproofgroup ENTER (optproofgroup)")
        $$ = nil
    }
    | PARSER16_TOK_PROOF proofgroup
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
        $$ = parser16Acfg(parser16lex).NewComposeTactics([]Node{$1, $3})
        $$.SetLineno(parser16NodeLineno($2))
    }
    ;

// --- match/renaming ---

match:
    defn
    {
        xtracer.Trace("parser.p_match_defn ENTER (match)")
        $$ = $1
    }
    | var PARSER16_TOK_EQ fmla
    {
        xtracer.Trace("parser.p_match_var_eq_fmla ENTER (match)")
        // Python: Definition(p[1], check_non_temporal(p[3]))
        $$ = parser16Acfg(parser16lex).NewDefinition($1, parser16CheckNonTemporal($3))
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
    }
    ;

matches:
    match
    {
        xtracer.Trace("parser.p_matches ENTER (matches)")
        $$ = []Node{$1}
    }
    | matches PARSER16_TOK_COMMA match
    {
        xtracer.Trace("parser.p_matches_matches_comma_match ENTER (matches)")
        $$ = append($1, $3)
    }
    ;

renamingitem:
    PARSER16_TOK_VARIABLE PARSER16_TOK_DIV PARSER16_TOK_VARIABLE
    {
        xtracer.Trace("parser.p_renamingitem_variable_div_variable ENTER (renamingitem)")
        // Python: Definition(Variable(p[3],universe),Variable(p[1],universe))
        $$ = parser16Acfg(parser16lex).NewDefinition(parser16Acfg(parser16lex).NewVariable($3.Val, "S"), parser16Acfg(parser16lex).NewVariable($1.Val, "S"))
    }
    | SYMBOLx PARSER16_TOK_DIV SYMBOLx
    {
        xtracer.Trace("parser.p_renamingitem_symbol_div_symbol ENTER (renamingitem)")
        $$ = parser16Acfg(parser16lex).NewDefinition(parser16Acfg(parser16lex).NewAtom($3.Val), parser16Acfg(parser16lex).NewAtom($1.Val))
    }
    ;

renaminglist:
    renamingitem
    {
        xtracer.Trace("parser.p_renaminglist_renamingitem ENTER (renaminglist)")
        $$ = []Node{$1}
    }
    | renaminglist PARSER16_TOK_COMMA renamingitem
    {
        xtracer.Trace("parser.p_renaminglist_renaminglist_comma_renamingitem ENTER (renaminglist)")
        $$ = append($1, $3)
    }
    ;

optrenaming:
    /* empty */
    {
        xtracer.Trace("parser.p_renaming ENTER (optrenaming)")
        $$ = parser16Acfg(parser16lex).NewRenaming(nil)
    }
    | renaming
    {
        xtracer.Trace("parser.p_optrenaming_renaming ENTER (optrenaming)")
        $$ = $1
    }
    ;

renaming:
    PARSER16_TOK_LT renaminglist PARSER16_TOK_GT
    {
        xtracer.Trace("parser.p_renaming_lt_renaminglist_gt ENTER (renaming)")
        r := parser16Acfg(parser16lex).NewRenaming($2)
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
        $$ = parser16Acfg(parser16lex).NewUnfoldSpec($1, $2)
    }
    ;

unfspecs:
    unfspec
    {
        xtracer.Trace("parser.p_unfspecs_unfspec ENTER (unfspecs)")
        $$ = []Node{$1}
    }
    | unfspecs PARSER16_TOK_COMMA unfspec
    {
        xtracer.Trace("parser.p_unfspecs_unfspecs_unfspec ENTER (unfspecs)")
        $$ = append($1, $3)
    }
    ;

// --- proofstep ---

proofstep:
    SYMBOLx
    {
        xtracer.Trace("parser.p_proofstep_symbol ENTER (proofstep)")
        a := parser16Acfg(parser16lex).NewAtom($1.Val)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        si := parser16Acfg(parser16lex).NewSchemaInstantiation(a, parser16Acfg(parser16lex).NewRenaming(nil))
        si.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = si
    }
    | SYMBOLx PARSER16_TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_symbol_with_defns ENTER (proofstep)")
        a := parser16Acfg(parser16lex).NewAtom($1.Val)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        si := parser16Acfg(parser16lex).NewSchemaInstantiationWithMatches(a, parser16Acfg(parser16lex).NewRenaming(nil), $3)
        si.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = si
    }
    | PARSER16_TOK_ASSUME atype optrenaming
    {
        xtracer.Trace("parser.p_proofstep_assume ENTER (proofstep)")
        // Python: AssumeGlobalTactic(a, p[3]); p[0].label = NoneAST()
        a := parser16AtypeToAtom(parser16Acfg(parser16lex), $2)
        a.SetLineno(parser16NodeLineno($2))
        at := parser16Acfg(parser16lex).NewAssumeGlobalTactic(a, $3)
        at.TLabel = parser16Acfg(parser16lex).NewNoneAST()
        at.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = at
    }
    | PARSER16_TOK_ASSUME atype optrenaming PARSER16_TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_assume_with_defns ENTER (proofstep)")
        // Python: AssumeGlobalTactic(*([a,p[3]]+p[5])); p[0].label = NoneAST()
        a := parser16AtypeToAtom(parser16Acfg(parser16lex), $2)
        a.SetLineno(parser16NodeLineno($2))
        at := parser16Acfg(parser16lex).NewAssumeGlobalTacticWithMatches(a, $3, $5)
        at.TLabel = parser16Acfg(parser16lex).NewNoneAST()
        at.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = at
    }
    | PARSER16_TOK_INSTANTIATE atype optrenaming
    {
        xtracer.Trace("parser.p_proofstep_instantiate ENTER (proofstep)")
        a := parser16AtypeToAtom(parser16Acfg(parser16lex), $2)
        a.SetLineno(parser16NodeLineno($2))
        at := parser16Acfg(parser16lex).NewAssumeTactic(a, $3)
        at.TLabel = parser16Acfg(parser16lex).NewNoneAST()
        at.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = at
    }
    | PARSER16_TOK_INSTANTIATE labelname atype optrenaming
    {
        xtracer.Trace("parser.p_proofstep_instance ENTER (proofstep)")
        lex := parser16lex.(*parser16LexAdapter)
        a := parser16AtypeToAtom(parser16Acfg(parser16lex), $3)
        a.SetLineno(parser16TokLineno(lex, $2))
        label := parser16Acfg(parser16lex).NewAtom($2.Val)
        label.SetLineno(parser16TokLineno(lex, $2))
        at := parser16Acfg(parser16lex).NewAssumeTactic(a, $4)
        at.TLabel = label
        at.SetLineno(parser16TokLineno(lex, $1))
        $$ = at
    }
    | PARSER16_TOK_INSTANTIATE labelname atype optrenaming PARSER16_TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_instance_with_matches ENTER (proofstep)")
        lex := parser16lex.(*parser16LexAdapter)
        a := parser16AtypeToAtom(parser16Acfg(parser16lex), $3)
        a.SetLineno(parser16TokLineno(lex, $2))
        label := parser16Acfg(parser16lex).NewAtom($2.Val)
        label.SetLineno(parser16TokLineno(lex, $2))
        at := parser16Acfg(parser16lex).NewAssumeTacticWithMatches(a, $4, $6)
        at.TLabel = label
        at.SetLineno(parser16TokLineno(lex, $1))
        $$ = at
    }
    | PARSER16_TOK_INSTANTIATE atype optrenaming PARSER16_TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_instantiate_with_defns ENTER (proofstep)")
        a := parser16AtypeToAtom(parser16Acfg(parser16lex), $2)
        a.SetLineno(parser16NodeLineno($2))
        at := parser16Acfg(parser16lex).NewAssumeTacticWithMatches(a, $3, $5)
        at.TLabel = parser16Acfg(parser16lex).NewNoneAST()
        at.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = at
    }
    | PARSER16_TOK_INSTANTIATE PARSER16_TOK_WITH pflets
    {
        xtracer.Trace("parser.p_proofstep_witness_pflets ENTER (proofstep)")
        wt := parser16Acfg(parser16lex).NewWitnessTactic($3)
        wt.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = wt
    }
    | PARSER16_TOK_SHOWGOALS
    {
        xtracer.Trace("parser.p_proofstep_showgoals ENTER (proofstep)")
        sg := parser16Acfg(parser16lex).NewShowGoalsTactic()
        sg.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = sg
    }
    | PARSER16_TOK_DEFERGOAL
    {
        xtracer.Trace("parser.p_proofstep_defergoal ENTER (proofstep)")
        dg := parser16Acfg(parser16lex).NewDeferGoalTactic()
        dg.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = dg
    }
    | PARSER16_TOK_SPOIL atype
    {
        xtracer.Trace("parser.p_proofstep_spoil_atype ENTER (proofstep)")
        // Python: a = Atom(p[2]) where p[2] is a string from atype
        a := parser16AtypeToAtom(parser16Acfg(parser16lex), $2)
        a.SetLineno(parser16NodeLineno($2))
        st := parser16Acfg(parser16lex).NewSpoilTactic(a)
        st.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = st
    }
    | PARSER16_TOK_TACTIC SYMBOLx opttacticwith optproofgroup
    {
        xtracer.Trace("parser.p_proofstep_tactic ENTER (proofstep)")
        a := parser16Acfg(parser16lex).NewAtom($2.Val)
        a.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        proof := Node($4)
        if proof == nil {
            proof = parser16Acfg(parser16lex).NewNoneAST()
        }
        tt := parser16Acfg(parser16lex).NewTacticTactic(a, $3, proof)
        tt.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = tt
    }
    | opttemporal PARSER16_TOK_PROPERTY labeledfmla optskolem optproofgroup
    {
        xtracer.Trace("parser.p_proofstep_property ENTER (proofstep)")
        lf := parser16AddLabel(parser16Acfg(parser16lex), $3.(*LabeledFormula), "prop")
        // Python: prop = addtemporal(lf) if p[1] else check_non_temporal(lf)
        if $1 != nil {
            lf = parser16AddTemporal(lf)
        } else {
            parser16CheckNonTemporal(lf)
        }
        name := Node($4)
        if name == nil { name = parser16Acfg(parser16lex).NewNoneAST() }
        proof := Node($5)
        if proof == nil { proof = parser16Acfg(parser16lex).NewNoneAST() }
        pt := parser16Acfg(parser16lex).NewPropertyTactic(lf, name, proof)
        pt.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = pt
    }
    | PARSER16_TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_proofstep_function ENTER (proofstep)")
        ft := parser16Acfg(parser16lex).NewFunctionTactic($2)
        ft.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = ft
    }
    | PARSER16_TOK_THEOREM lgprop optproofgroup
    {
        xtracer.Trace("parser.p_proofstep_theorem ENTER (proofstep)")
        lf := parser16AddLabel(parser16Acfg(parser16lex), $2.(*LabeledFormula), "thm")
        proof := Node($3)
        if proof == nil { proof = parser16Acfg(parser16lex).NewNoneAST() }
        pt := parser16Acfg(parser16lex).NewPropertyTactic(lf, parser16Acfg(parser16lex).NewNoneAST(), proof)
        pt.SetLineno(parser16NodeLineno($2))
        $$ = pt
    }
    | PARSER16_TOK_PROOF labelname proofgroup
    {
        xtracer.Trace("parser.p_proofstep_proof ENTER (proofstep)")
        lex := parser16lex.(*parser16LexAdapter)
        label := parser16Acfg(parser16lex).NewAtom($2.Val)
        label.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        pt := parser16Acfg(parser16lex).NewProofTactic(label, $3)
        pt.SetLineno(parser16TokLineno(lex, $1))
        $$ = pt
    }
    | PARSER16_TOK_LET pflets
    {
        xtracer.Trace("parser.p_proofstep_let_pflets ENTER (proofstep)")
        lt := parser16Acfg(parser16lex).NewLetTactic($2)
        lt.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = lt
    }
    | PARSER16_TOK_IF fmla proofgroup PARSER16_TOK_ELSE proofgroup
    {
        xtracer.Trace("parser.p_proofstep_if_fmla_proofgroup_else_proofgroup ENTER (proofstep)")
        it := parser16Acfg(parser16lex).NewIfTactic($2, $3, $5)
        it.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = it
    }
    | PARSER16_TOK_UNFOLD atype PARSER16_TOK_WITH unfspecs
    {
        xtracer.Trace("parser.p_proofstep_unfold_atype_with_defns ENTER (proofstep)")
        a := parser16AtypeToAtom(parser16Acfg(parser16lex), $2)
        a.SetLineno(parser16NodeLineno($2))
        ut := parser16Acfg(parser16lex).NewUnfoldTactic(a, $4)
        ut.TLabel = parser16Acfg(parser16lex).NewNoneAST()
        ut.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = ut
    }
    | PARSER16_TOK_UNFOLD PARSER16_TOK_WITH unfspecs
    {
        xtracer.Trace("parser.p_proofstep_unfold_with_defns ENTER (proofstep)")
        ut := parser16Acfg(parser16lex).NewUnfoldTactic(parser16Acfg(parser16lex).NewNoneAST(), $3)
        ut.TLabel = parser16Acfg(parser16lex).NewNoneAST()
        ut.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
        $$ = ut
    }
    | PARSER16_TOK_FORGET callatoms
    {
        xtracer.Trace("parser.p_proofstep_forget_callatoms ENTER (proofstep)")
        ft := parser16Acfg(parser16lex).NewForgetTactic($2)
        ft.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $1))
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
        $$ = parser16Acfg(parser16lex).NewAnd()
    }
    | PARSER16_TOK_REQUIRES fmla
    {
        xtracer.Trace("parser.p_requires_requires_fmla ENTER (requires)")
        $$ = $2
    }
    ;

ensures:
    PARSER16_TOK_ENSURES fmla
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
    | PARSER16_TOK_MODIFIES PARSER16_TOK_LCB PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_modifies_modifies_lcb_rcb ENTER (modifies)")
        $$ = parser16Acfg(parser16lex).NewAnd() // empty modifies
    }
    | PARSER16_TOK_MODIFIES PARSER16_TOK_TIMES
    {
        xtracer.Trace("parser.p_modifies_modofies_times ENTER (modifies)")
        $$ = nil
    }
    | PARSER16_TOK_MODIFIES atoms
    {
        xtracer.Trace("parser.p_modifies_modifies_atoms ENTER (modifies)")
        $$ = parser16Acfg(parser16lex).NewAnd($2...)
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
    PARSER16_TOK_PARAMS tterms PARSER16_TOK_IN action PARSER16_TOK_ARROW requires ensures
    {
        xtracer.Trace("parser.p_upax_params_apps_in_action_arrow_ensures_fmla ENTER (upax)")
        cfg := parser16Acfg(parser16lex)
        // Python (ivy_parser.py:1980):
        //   p[0] = UpdatePattern(ConstantDecl(*p[2]), p[4], p[6], p[7])
        params := cfg.NewConstantDecl($2...)
        $$ = cfg.NewUpdatePattern(params, $4, $6, $7)
    }
    ;

assert_rhs:
    PARSER16_TOK_LCB requires modifies ensures PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_assert_rhs_lcb_requires_modifies_ensures_rcb ENTER (assert_rhs)")
        $$ = parser16Acfg(parser16lex).NewAtom("rme", $2, $4)
    }
    | fmla
    {
        xtracer.Trace("parser.p_assert_rhs_fmla ENTER (assert_rhs)")
        $$ = $1
    }
    ;

state_expr:
    PARSER16_TOK_TRUE
    {
        xtracer.Trace("parser.p_state_expr_true ENTER (state_expr)")
        $$ = parser16Acfg(parser16lex).NewAnd()
    }
    | PARSER16_TOK_FALSE
    {
        xtracer.Trace("parser.p_state_expr_false ENTER (state_expr)")
        $$ = parser16Acfg(parser16lex).NewOr()
    }
    | SYMBOLx
    {
        xtracer.Trace("parser.p_state_expr_symbol ENTER (state_expr)")
        $$ = parser16Acfg(parser16lex).NewAtom($1.Val)
    }
    | SYMBOLx PARSER16_TOK_LPAREN state_expr PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_state_expr_symbol_lparen_state_expr_rparen ENTER (state_expr)")
        $$ = parser16Acfg(parser16lex).NewAtom($1.Val, $3)
    }
    | state_expr PARSER16_TOK_OR state_expr
    {
        xtracer.Trace("parser.p_state_expr_state_expr_or_state_expr ENTER (state_expr)")
        $$ = parser16Acfg(parser16lex).NewOr($1, $3)
    }
    | PARSER16_TOK_LCB requires modifies ensures PARSER16_TOK_RCB
    {
        xtracer.Trace("parser.p_state_expr_lcb_requires_modifies_ensures_rcb ENTER (state_expr)")
        $$ = parser16Acfg(parser16lex).NewAtom("rme", $2, $4)
    }
    | PARSER16_TOK_ENTRY
    {
        xtracer.Trace("parser.p_state_expr_entry ENTER (state_expr)")
        $$ = parser16Acfg(parser16lex).NewAtom("entry")
    }
    ;

// --- Concept space ---

cdefn:
    atom PARSER16_TOK_EQ expr
    {
        xtracer.Trace("parser.p_cdefn_atom_expr ENTER (cdefn)")
        $$ = parser16Acfg(parser16lex).NewDefinition(AppToAtom($1), $3)
        $$.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
    }
    ;

cdefns:
    cdefn
    {
        xtracer.Trace("parser.p_cdefns_cdefn ENTER (cdefns)")
        $$ = []Node{$1}
    }
    | cdefns PARSER16_TOK_COMMA cdefn
    {
        xtracer.Trace("parser.p_cdefns_cdefns_comma_cdefn ENTER (cdefns)")
        $$ = append($1, $3)
    }
    ;

expr:
    PARSER16_TOK_LCB fmla PARSER16_TOK_RCB
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
        $$ = parser16Acfg(parser16lex).NewAtom($2, $1, $3)
        $$.SetLineno(parser16GetLineno(parser16lex.(*parser16LexAdapter)))
    }
    | exprterm PARSER16_TOK_TILDAEQ exprterm
    {
        xtracer.Trace("parser.p_expr_exprterm_tildaeq_exprterm ENTER (expr)")
        n := parser16Acfg(parser16lex).NewNot(parser16Acfg(parser16lex).NewAtom("=", $1, $3))
        n.SetLineno(parser16TokLineno(parser16lex.(*parser16LexAdapter), $2))
        $$ = n
    }
    | PARSER16_TOK_TILDA expr
    {
        xtracer.Trace("parser.p_expr_tilda_atom ENTER (expr)")
        $$ = parser16Acfg(parser16lex).NewNot($2)
    }
    | PARSER16_TOK_LPAREN expr PARSER16_TOK_RPAREN
    {
        xtracer.Trace("parser.p_expr_lparen_expr_rparen ENTER (expr)")
        $$ = $2
    }
    | prod
    {
        xtracer.Trace("parser.p_expr_prod ENTER (expr)")
        $$ = parser16Acfg(parser16lex).NewAtom("product", $1...)
    }
    | sum
    {
        xtracer.Trace("parser.p_expr_sum ENTER (expr)")
        $$ = parser16Acfg(parser16lex).NewAtom("sum", $1...)
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
    expr PARSER16_TOK_TIMES expr
    {
        xtracer.Trace("parser.p_prod_expr_expr ENTER (prod)")
        $$ = []Node{$1, $3}
    }
    | prod PARSER16_TOK_TIMES expr
    {
        xtracer.Trace("parser.p_prod_prod_expr ENTER (prod)")
        $$ = append($1, $3)
    }
    ;

sum:
    expr PARSER16_TOK_PLUS expr
    {
        xtracer.Trace("parser.p_sum_expr_expr ENTER (sum)")
        $$ = []Node{$1, $3}
    }
    | sum PARSER16_TOK_PLUS expr
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
        $$ = []Node{parser16Acfg(parser16lex).NewAtom($1.Val)}
    }
    | symbols PARSER16_TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_symbols_symbols_symbol ENTER (symbols)")
        $$ = append($1, parser16Acfg(parser16lex).NewAtom($3.Val))
    }
    ;

%%

// lalrMakeSequence wraps a list of action nodes into a single sequence node.
// parser16LowerVarStmts transforms VarAction (local variable declarations) into
// LocalAction with proper scoping. Matches Python lower_var_stmts (ivy_parser.py:2670-2697).
func parser16LowerVarStmts(stmts []Node) []Node {
	return LowerVarStatements(stmts)
}

// lalrMakeSequence matches Python p_sequence_lcb_actseq_rcb (ivy_parser.py:2705-2713).
// Python: stmts = lower_var_stmts(p[2]); if len(stmts)==1: return stmts[0]
//         else: Sequence(*lower_var_stmts(stmts))
// So lower_var_stmts is called ONCE always, and a SECOND time only if len > 1.
func parser16MakeSequence(cfg *AstConfig, stmts []Node) Node {
	if len(stmts) == 0 {
		return cfg.NewSequence()
	}
	if len(stmts) == 1 {
		return stmts[0]
	}
	stmts = parser16LowerVarStmts(stmts)
	return cfg.NewSequence(stmts...)
}

// parser16StmtToSeq matches Python stmt_to_seq() used by around declarations
// in Ivy 1.7. It deliberately has its own trace boundary before lowering.
func parser16StmtToSeq(cfg *AstConfig, stmts []Node) Node {
	xtracer.Trace("parser.stmt_to_seq ENTER")
	stmts = parser16LowerVarStmts(stmts)
	if len(stmts) == 1 {
		return stmts[0]
	}
	res := cfg.NewSequence(stmts...)
	if len(stmts) > 0 {
		res.SetLineno(stmts[0].GetLineno())
	}
	return res
}
