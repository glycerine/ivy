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
	"github.com/glycerine/goivy/ast"
	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/xtracer"
)

// labelCounter is a package-level counter for generating unique label/mixer names.
var lalrLabelCounter int

// parentObject matches Python's global parent_object.
// Set by objsym rule, consumed by newIvyAccum to inherit defined symbols
// for continuation objects.
var parentObject string

// getLineno returns a Location for the current token position.
// Matches Python's get_lineno(p, n) → iu.Location(iu.filename, p.lineno(n)).
func getLineno(lex *v17LexAdapter) ast.Location {
	xtracer.Trace("parser.get_lineno ENTER")
	return ast.Location{
		Filename: normalizeFilename(lex.filename),
		Line:     lex.lastTok.Line,
	}
}

// normalizeFilename replaces the include directory path with <IVY_INCLUDE>
// so that canonical strings match between Go and Python regardless of install location.
// Python: stores the raw path; the golden test normalizes for display.
// For hashing, both sides must agree, so we normalize here.
func normalizeFilename(f string) string {
	stdDir := iu.GetStdIncludeDir()
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
func newLabel(pref string) *ast.Atom {
	xtracer.Trace("parser.newlabel ENTER")
	lalrLabelCounter++
	return ast.NewAtom(fmt.Sprintf("%s%d", pref, lalrLabelCounter))
}

// addLabel adds a label to a LabeledFormula if it doesn't have one.
// Matches Python addlabel() (ivy_parser.py:443-448).
func addLabel(lf *ast.LabeledFormula, pref string) *ast.LabeledFormula {
	xtracer.Trace("parser.addlabel ENTER")
	if lf.Label != nil {
		return lf
	}
	res := ast.NewLabeledFormula(newLabel(pref), lf.Formula)
	res.Lineno = lf.Lineno
	return res
}

// mkLF wraps a node in a LabeledFormula with no label.
// Matches Python mk_lf (ivy_parser.py:1187).
func mkLF(x ast.Node) *ast.LabeledFormula {
	xtracer.Trace("parser.mk_lf ENTER")
	lf := ast.NewLabeledFormula(nil, x)
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
func atypeToAtom(n ast.Node) *ast.Atom {
	switch v := n.(type) {
	case *ast.Symbol:
		return ast.NewAtom(v.Rep)
	case *ast.This:
		return ast.NewAtom("this")
	default:
		return ast.NewAtom(fmt.Sprint(n))
	}
}

// makeMixinName generates a unique mixin name.
// Matches Python make_mixin_name() (ivy_parser.py:2556-2562).
func makeMixinName(atom *ast.Atom, suffix string) *ast.Atom {
	xtracer.Trace("parser.make_mixin_name ENTER")
	lalrLabelCounter++
	return ast.NewAtom(fmt.Sprintf("%s[%s%d]", atom.Rep, suffix, lalrLabelCounter))
}

// handleMixin declares a mixin (before/after/implement).
// Matches Python handle_mixin() (ivy_parser.py:2083-2090).
func handleMixin(kind string, mixer *ast.Atom, mixee *ast.Atom, ivy *ivyAccum) {
	xtracer.Trace("parser.handle_mixin ENTER")
	var m ast.Node
	switch kind {
	case "before":
		m = &ast.MixinBeforeDef{MixerNode: mixer, MixeeNode: mixee}
	case "after":
		m = &ast.MixinAfterDef{MixerNode: mixer, MixeeNode: mixee}
	case "implement":
		m = &ast.MixinImplementDef{MixerNode: mixer, MixeeNode: mixee}
	default:
		m = &ast.MixinBeforeDef{MixerNode: mixer, MixeeNode: mixee}
	}
	md := ast.NewMixinDecl(m)
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
	mixee, _ := stackActionLookup(ivy, actname)
	if mixee == nil {
		return formals, returns
	}
	// TODO: full param inference from matching action
	_ = mixee
	return formals, returns
}

func handleBeforeAfter(kind string, atom *ast.Atom, action ast.Node, ivy *ivyAccum, optargs []ast.Node, optreturns []ast.Node) {
	xtracer.Trace("parser.handle_before_after ENTER")
	mixer := makeMixinName(atom, kind)
	optargs, optreturns = inferActionParams(ivy, atom.Rep, optargs, optreturns)
	df := &ast.ActionDef{Name: mixer, Body: action, FormalParams: optargs, FormalReturns: optreturns}
	df.SetLineno(atom.GetLineno())
	decl := ast.NewActionDecl(df)
	ivy.declare(decl)
	handleMixin(kind, mixer, atom, ivy)
}

// createObject processes an object declaration by expanding its body
// with prefix substitution via instMod.
// Matches Python create_object() (ivy_parser.py:678-692).
// setObjectDefined matches Python set_object_defined (ivy_parser.py:354-359).
func setObjectDefined(ivy *ivyAccum, name string) {
	xtracer.Trace("parser.set_object_defined ENTER")
	// TODO: track defined names for object scope
}

// parseNativequote parses a native code block, splitting on backtick-delimited references.
// Matches Python parse_nativequote() (ivy_parser.py:2457-2470).
func parseNativequote(raw string, lex *v17LexAdapter) (string, []ast.Node) {
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
				bqs = append(bqs, ast.NewAtom("this"))
			} else {
				bqs = append(bqs, ast.NewAtom(f))
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

// fixIfPart handles the `some` condition case in if/while actions.
// Matches Python fix_if_part() (ivy_parser.py:2932-2938).
func fixIfPart(cond ast.Node, part ast.Node) ast.Node {
	xtracer.Trace("parser.fix_if_part ENTER")
	if some, ok := cond.(*ast.Some); ok {
		subst := make(map[string]string)
		for _, p := range some.Params {
			if v, ok := p.(*ast.Variable); ok && len(v.Rep) > 4 {
				subst[v.Rep[4:]] = v.Rep
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
func createObject(top *ivyAccum, name *ast.Atom, objectargs []ast.Node, module *ivyAccum, lineno ast.Location, continuation bool) {
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
		prefargs = append(prefargs, ast.NewVariable(vname, sort))
	}

	// Python line 681: pref = Atom(name, prefargs)
	pref := ast.NewAtom(name.Rep, prefargs...)
	pref.SetLineno(lineno)

	// Python line 684-686
	if !continuation {
		top.declare(ast.NewObjectDecl(pref))
		setObjectDefined(top, name.Rep)
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
	instMod(top, module.decls, pref, map[string]string{}, vsubst, "")

	xtracer.Trace("parser.create_object EXIT name=%s", name.Rep)
}

%}

// The union type for semantic values.
%union {
	node     ast.Node
	nodes    []ast.Node
	str      string
	bval     bool
	accum    *ivyAccum
}

// Terminal tokens — identifiers and literals
%token <str>  TOK_PRESYMBOL TOK_VARIABLE
%token <str>  TOK_LABEL   // Python returns LABEL from lexer; Go handles via labelname rule instead
%token <str>  TOK_NATIVEQUOTE

// Terminal tokens — punctuation
%token        TOK_LPAREN TOK_RPAREN TOK_LB TOK_RB TOK_LCB TOK_RCB
%token        TOK_COMMA TOK_SEMI TOK_COLON TOK_DOT
%token        TOK_DOTS TOK_DOTDOTDOT

// Terminal tokens — operators
%token        TOK_PLUS TOK_MINUS TOK_TIMES TOK_DIV
%token        TOK_EQ TOK_TILDAEQ TOK_TILDA TOK_LE TOK_LT TOK_GE TOK_GT
%token        TOK_AND TOK_OR TOK_ARROW TOK_IFF
%token        TOK_PTO TOK_DOLLAR TOK_CARET
%token        TOK_ASSIGN

// Terminal tokens — keywords (logic)
%token        TOK_FORALL TOK_EXISTS
%token        TOK_TRUE TOK_FALSE
%token        TOK_OLD TOK_THIS TOK_ISA
%token        TOK_IF TOK_ELSE
%token        TOK_GLOBALLY TOK_EVENTUALLY
%token        TOK_WHENNEXT TOK_WHENPREV TOK_WHENFIRST TOK_WHENLAST

// Terminal tokens — keywords (actions)
%token        TOK_ASSUME TOK_ASSERT TOK_REQUIRE TOK_ENSURE
%token        TOK_VAR TOK_LOCAL TOK_LET TOK_CALL
%token        TOK_WHILE TOK_FOR TOK_IN TOK_INVARIANT TOK_DECREASES
%token        TOK_RETURNS
%token        TOK_SOME TOK_MINIMIZING TOK_MAXIMIZING
%token        TOK_DEBUG TOK_THUNK TOK_UNPROVABLE TOK_PROOF
%token        TOK_INSTANTIATE

// Terminal tokens — keywords (declarations)
%token        TOK_RELATION TOK_INDIV TOK_FUNCTION TOK_DERIVED
%token        TOK_AXIOM TOK_CONJECTURE TOK_SCHEMA TOK_THEOREM
%token        TOK_PROPERTY TOK_DEFINITION
%token        TOK_TYPE TOK_STRUCT
%token        TOK_MODULE TOK_OBJECT TOK_CLASS TOK_SUBCLASS
%token        TOK_ACTION TOK_METHOD
%token        TOK_BEFORE TOK_AFTER TOK_AROUND TOK_MIXIN TOK_IMPLEMENT
%token        TOK_ISOLATE TOK_EXTRACT TOK_TRUSTED
%token        TOK_EXPORT TOK_IMPORT TOK_DELEGATE TOK_USING TOK_INCLUDE
%token        TOK_INTERPRET TOK_MACRO TOK_ALIAS TOK_ATTRIBUTE
%token        TOK_VARIANT TOK_OF
%token        TOK_SCENARIO
%token        TOK_PROGRESS TOK_RELY TOK_MIXORD
%token        TOK_CONCEPT TOK_STATE TOK_UPDATE TOK_FROM
%token        TOK_PARAMS TOK_MODIFIES TOK_ENSURES TOK_REQUIRES
%token        TOK_INIT TOK_ENTRY TOK_SET TOK_NULL TOK_MATCH
%token        TOK_FRESH TOK_NAMED
%token        TOK_TEMPORAL TOK_EXPLICIT
%token        TOK_SPECIFICATION TOK_IMPLEMENTATION TOK_PRIVATE
%token        TOK_GLOBAL TOK_COMMON
%token        TOK_GHOST TOK_FINITE
%token        TOK_PARAMETER
%token        TOK_DESTRUCTOR TOK_CONSTRUCTOR TOK_FIELD
%token        TOK_AUTOINSTANCE
// TOK_VAR_KW removed (was unused placeholder)

// Proof/tactic tokens
%token        TOK_TACTIC TOK_TRIGGER
%token        TOK_SHOWGOALS TOK_DEFERGOAL TOK_SPOIL
%token        TOK_UNFOLD TOK_FORGET
%token        TOK_APPLY
%token        TOK_WITH
// TOK_METHOD_KW, TOK_NULL_KW, TOK_SET_KW removed (were unused placeholders)

// Nonterminal types — formula/term
%type <node>  term fmla appelem aterm var simplevar atype
%type <nodes> terms vars simplevars
%type <str>   SYMBOLx SYMsubscr labelname

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
        parent := v17lex.(*v17LexAdapter).accum // nil for outermost top
        $$ = newIvyAccum()
        $$.parent = parent
        v17lex.(*v17LexAdapter).accum = $$
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
        name := $3
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
            pref := ast.NewAtom(name)
            pref.SetLineno(getLineno(lex))
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
        lf := addLabel($5.(*ast.LabeledFormula), "axiom")
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
        d := ast.NewAxiomDecl(lf)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$.declare(d)
    }
    // --- Property (v1.7+): top optexplicit opttemporal PROPERTY labeledfmla optskolem optproof ---
    | top optexplicit opttemporal TOK_PROPERTY labeledfmla optskolem optproof
    {
        xtracer.Trace("parser.p_top_property_labeledfmla ENTER (top)")
        $$ = $1
        lf := addLabel($5.(*ast.LabeledFormula), "prop")
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
        d := ast.NewPropertyDecl(lf)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$.declare(d)
        if $6 != nil {
            $$.declare(ast.NewNamedDecl($6))
        }
        if $7 != nil {
            $$.declare(ast.NewProofDecl($7))
        }
    }
    // --- Conjecture ---
    | top TOK_CONJECTURE labeledfmla
    {
        xtracer.Trace("parser.p_top_conjecture_labeledfmla ENTER (top)")
        $$ = $1
        lf := addLabel($3.(*ast.LabeledFormula), "conj")
        d := ast.NewConjectureDecl(lf)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$.declare(d)
    }
    // --- Invariant (v1.7+): top optexplicit INVARIANT labeledfmla optproof ---
    | top optexplicit TOK_INVARIANT labeledfmla optproof
    {
        xtracer.Trace("parser.p_top_invariant_labeledfmla ENTER (top)")
        $$ = $1
        lf := addLabel($4.(*ast.LabeledFormula), "invar")
        lf.Unprovable = false
        if $2 != nil {
            lf.Explicit = true
        }
        d := ast.NewConjectureDecl(lf)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$.declare(d)
        if $5 != nil {
            $$.declare(ast.NewProofDecl($5))
        }
    }
    // --- Unprovable Invariant ---
    | top TOK_UNPROVABLE TOK_INVARIANT labeledfmla optproof
    {
        xtracer.Trace("parser.p_top_unprovable_invariant_labeledfmla ENTER (top)")
        $$ = $1
        lf := addLabel($4.(*ast.LabeledFormula), "invar")
        lf.Unprovable = true
        lf.Explicit = true
        d := ast.NewConjectureDecl(lf)
        // Python: only declare if check_unprovable — we always declare for now
        $$.declare(d)
        if $5 != nil {
            $$.declare(ast.NewProofDecl($5))
        }
    }
    // --- Module: top MODULE modulestart modcat atom optwith EQ LCB top RCB moduleend ---
    | top TOK_MODULE modulestart modcat atom optwith TOK_EQ TOK_LCB top TOK_RCB moduleend
    {
        xtracer.Trace("parser.p_top_module_atom_eq_lcb_top_rcb ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        modAccum := $9
        body := ast.NewSequence(modAccum.decls...)
        d := ast.NewDefinition(ast.AppToAtom($5), body)
        $$.declare(ast.NewModuleDecl(d))
        // Python: if p[4] == "isolate": ... with get_lineno(p,2) on this, iso, d.args[0], d
        if $4 != nil {
            if catAtom, ok := $4.(*ast.Atom); ok && catAtom.Rep == "isolate" {
                thisAtom := ast.NewAtom("this")
                thisAtom.SetLineno(getLineno(lex))
                iso := ast.NewAtom("iso")
                iso.SetLineno(getLineno(lex))
                isoElems := append([]ast.Node{iso, thisAtom}, $6...)
                isoDef := &ast.IsolateDef{Elems: isoElems, WithArgs: len($6)}
                isoDef.SetLineno(getLineno(lex))
                isoDecl := ast.NewIsolateDecl(isoDef)
                isoDecl.Attributes = []ast.Node{ast.NewAtom("common")}
                isoDecl.SetLineno(getLineno(lex))
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
        lineno := getLineno(v17lex.(*v17LexAdapter))
        createObject($$, pref, $4, objAccum, lineno, $7)
        // Python: create_object does stack.pop() at ivy_parser.py:692
        v17lex.(*v17LexAdapter).accum = $$
    }
    // --- Class ---
    | top TOK_CLASS objsym objectargs TOK_EQ TOK_LCB optdotdotdot top TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top__top_class_symbol_objectargs_eq_lcb_optdo ENTER (top)")
        $$ = $1
        objAccum := $8
        pref := $3.(*ast.Atom)
        // Declare type
        scnst := &ast.This{}
        tdfn := &ast.TypeDef{Name: ast.NewAtom("this"), Value: ast.NewUninterpretedSortAST()}
        td := ast.NewTypeDecl(tdfn)
        _ = scnst
        
        objDecl := ast.NewObjectDecl(pref)
        $$.declare(objDecl)
        $$.declare(td)
        for _, d := range objAccum.decls {
            $$.declare(d)
        }
        // Python: stack.pop() equivalent
        v17lex.(*v17LexAdapter).accum = $$
    }
    // --- Subclass ---
    | top TOK_SUBCLASS objsym TOK_OF atype TOK_EQ TOK_LCB optdotdotdot top TOK_RCB objectend
    {
        xtracer.Trace("parser.p_top__top_subclass_symbol_of_atype_eq_lcb_optd ENTER (top)")
        $$ = $1
        objAccum := $9
        pref := $3.(*ast.Atom)
        
        objDecl := ast.NewObjectDecl(pref)
        $$.declare(objDecl)
        for _, d := range objAccum.decls {
            $$.declare(d)
        }
        // Python: stack.pop() equivalent
        v17lex.(*v17LexAdapter).accum = $$
    }
    // --- Definition (v1.7+): top optexplicit DEFINITION optlabel gdefn optproof ---
    | top optexplicit TOK_DEFINITION optlabel gdefn optproof
    {
        xtracer.Trace("parser.p_top_definition_optlabel_gdefn_optproof ENTER (top)")
        $$ = $1
        lf := ast.NewLabeledFormula($4, $5)
        lf.Lineno = getLineno(v17lex.(*v17LexAdapter)).Line
        lf = addLabel(lf, "def")
        dd := ast.NewDefinitionDecl(lf)
        $$.declare(dd)
        if $6 != nil {
            $$.declare(ast.NewProofDecl($6))
        }
    }
    // --- Schema ---
    | top TOK_SCHEMA schdefn
    {
        xtracer.Trace("parser.p_top_schema_defn ENTER (top)")
        $$ = $1
        sch := &ast.Schema{Defn: $3}
        sd := ast.NewSchemaDecl(sch)
        $$.declare(sd)
    }
    // --- Theorem with schdefn ---
    | top TOK_THEOREM schdefn optproof
    {
        xtracer.Trace("parser.p_top_theorem_defn ENTER (top)")
        $$ = $1
        sch := &ast.Schema{Defn: $3}
        td := ast.NewTheoremDecl(sch)
        $$.declare(td)
        if $4 != nil {
            $$.declare(ast.NewProofDecl($4))
        }
    }
    // --- Theorem with LABEL schdefnrhs ---
    | top TOK_THEOREM labelname schdefnrhs optproof
    {
        xtracer.Trace("parser.p_top_theorem_label_rhs ENTER (top)")
        $$ = $1
        label := ast.NewAtom($3)
        label.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        df := ast.NewDefinition(label, $4)
        df.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        sch := &ast.Schema{Defn: df}
        td := ast.NewTheoremDecl(sch)
        $$.declare(td)
        if $5 != nil {
            $$.declare(ast.NewProofDecl($5))
        }
    }
    // --- Proof LABEL proofstep ---
    | top TOK_PROOF labelname proofstep
    {
        xtracer.Trace("parser.p_top_proof_label_label_proofstep ENTER (top)")
        $$ = $1
        label := ast.NewAtom($3)
        label.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        lf := ast.NewLabeledFormula(label, $4)
        $$.declare(ast.NewProofDecl(lf))
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
        d := ast.NewAutoInstanceDecl($3...)
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
            args[i] = addLabel(mkLF(x), "def")
        }
        dd := ast.NewDerivedDecl(args...)
        $$.declare(dd)
    }
    // --- Type (uninterpreted) ---
    | top optfinite optghost TOK_TYPE typesymbol
    {
        xtracer.Trace("parser.p_top_type_symbol ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        scnst := ast.NewAtom($5.(*ast.Atom).Rep)
        scnst.SetLineno(getLineno(lex))
        tdfn := &ast.TypeDef{Name: scnst, Value: ast.NewUninterpretedSortAST()}
        if $2 { tdfn.Finite = true }
        tdfn.SetLineno(getLineno(lex))
        td := ast.NewTypeDecl(tdfn)
        $$.declare(td)
    }
    // --- Type with sort ---
    | top optfinite optghost TOK_TYPE typesymbol TOK_EQ sort
    {
        xtracer.Trace("parser.p_top_type_symbol_eq_sort ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        scnst := ast.NewAtom($5.(*ast.Atom).Rep)
        scnst.SetLineno(getLineno(lex))
        tdfn := &ast.TypeDef{Name: scnst, Value: $7}
        if $2 { tdfn.Finite = true }
        tdfn.SetLineno(getLineno(lex))
        td := ast.NewTypeDecl(tdfn)
        $$.declare(td)
    }
    // --- Progress ---
    | top TOK_PROGRESS defns
    {
        xtracer.Trace("parser.p_top_progress_defns ENTER (top)")
        $$ = $1
        pd := ast.NewProgressDecl($3...)
        $$.declare(pd)
    }
    // --- Rely ---
    | top TOK_RELY atom TOK_ARROW atom
    {
        xtracer.Trace("parser.p_top_rely_atom_arrow_atom ENTER (top)")
        $$ = $1
        imp := &ast.Implies{T1: $3, T2: $5}
        rd := ast.NewRelyDecl(imp)
        $$.declare(rd)
    }
    | top TOK_RELY atom
    {
        xtracer.Trace("parser.p_top_rely_atom ENTER (top)")
        $$ = $1
        rd := ast.NewRelyDecl($3)
        $$.declare(rd)
    }
    // --- Mixord ---
    | top TOK_MIXORD callatom TOK_ARROW callatom
    {
        xtracer.Trace("parser.p_top_mixord_callatom_arrow_callatom ENTER (top)")
        $$ = $1
        imp := &ast.Implies{T1: $3, T2: $5}
        md := ast.NewMixOrdDecl(imp)
        $$.declare(md)
    }
    // --- Concept ---
    | top TOK_CONCEPT cdefns
    {
        xtracer.Trace("parser.p_top_concept_cdefns ENTER (top)")
        $$ = $1
        cd := ast.NewConceptDecl($3...)
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
        d := ast.NewDefinition(ast.AppToAtom($3), $5)
        md := ast.NewMacroDecl(d)
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
            lineno = getLineno(v17lex.(*v17LexAdapter))
        }
        theAtom := ast.NewAtom($4)
        theAtom.SetLineno(lineno)
        actdef := &ast.ActionDef{
            Name:    theAtom,
            Body:    adef,
            FormalParams: $5,
            FormalReturns: $6,
        }
        actdef.SetLineno(lineno)
        decl := ast.NewActionDecl(actdef)
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
        m := &ast.MixinBeforeDef{MixerNode: $3, MixeeNode: $5}
        md := ast.NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Mixin after ---
    | top TOK_MIXIN callatom TOK_AFTER callatom
    {
        xtracer.Trace("parser.p_top_mixin_callatom_after_callatom ENTER (top)")
        $$ = $1
        m := &ast.MixinAfterDef{MixerNode: $3, MixeeNode: $5}
        md := ast.NewMixinDecl(m)
        $$.declare(md)
    }
    // --- Before ---
    | top TOK_BEFORE atype optargs optreturns sequence
    {
        xtracer.Trace("parser.p_top_before_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := ast.NewAtom($3.(*ast.Symbol).Rep)
        atom.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        handleBeforeAfter("before", atom, $6, $$, $4, $5)
    }
    // --- After ---
    | top TOK_AFTER atype optargs optreturns topseq
    {
        xtracer.Trace("parser.p_top_after_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := ast.NewAtom($3.(*ast.Symbol).Rep)
        atom.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        handleBeforeAfter("after", atom, $6, $$, $4, $5)
    }
    // --- Around ---
    | top TOK_AROUND atype optargs optreturns TOK_LCB actseq optsemi TOK_DOTDOTDOT actseq optsemi TOK_RCB
    {
        xtracer.Trace("parser.p_top_around_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := ast.NewAtom($3.(*ast.Symbol).Rep)
        before := lalrMakeSequence($7)
        after := lalrMakeSequence($10)
        // before mixin
        lalrLabelCounter++
        bmixer := ast.NewAtom(fmt.Sprintf("%s[before%d]", atom.Rep, lalrLabelCounter))
        bdf := &ast.ActionDef{Name: bmixer, Body: before, FormalParams: $4, FormalReturns: $5}
        bdecl := ast.NewActionDecl(bdf)
        $$.declare(bdecl)
        bm := &ast.MixinBeforeDef{MixerNode: bmixer, MixeeNode: atom}
        bmd := ast.NewMixinDecl(bm)
        $$.declare(bmd)
        // after mixin
        lalrLabelCounter++
        amixer := ast.NewAtom(fmt.Sprintf("%s[after%d]", atom.Rep, lalrLabelCounter))
        adf := &ast.ActionDef{Name: amixer, Body: after, FormalParams: $4, FormalReturns: $5}
        adecl := ast.NewActionDecl(adf)
        $$.declare(adecl)
        am := &ast.MixinAfterDef{MixerNode: amixer, MixeeNode: atom}
        amd := ast.NewMixinDecl(am)
        $$.declare(amd)
    }
    // --- After init ---
    | top TOK_AFTER TOK_INIT optargs topseq
    {
        xtracer.Trace("parser.p_top_after_init_optargs_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := ast.NewAtom("init")
        atom.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        handleBeforeAfter("after", atom, $5, $$, $4, nil)
    }
    // --- Implement ---
    | top TOK_IMPLEMENT atype optargs optreturns topseq
    {
        xtracer.Trace("parser.p_top_implement_callatom_lcb_action_rcb ENTER (top)")
        $$ = $1
        atom := ast.NewAtom($3.(*ast.Symbol).Rep)
        atom.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        handleBeforeAfter("implement", atom, $6, $$, $4, $5)
    }
    // --- Implement type ---
    | top TOK_IMPLEMENT TOK_TYPE SYMBOLx TOK_WITH SYMBOLx
    {
        xtracer.Trace("parser.p_top_implement_type_symbol_with_symbol ENTER (top)")
        $$ = $1
        a1 := ast.NewAtom($4)
        a2 := ast.NewAtom($6)
        impl := &ast.ImplementTypeDef{Elems: []ast.Node{a1, a2}}
        d := ast.NewImplementTypeDecl(mkLF(impl))
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
        nameAtom := ast.NewAtom($4, $5...)
        elems := append([]ast.Node{nameAtom}, $7...)
        idef := &ast.IsolateDef{Elems: elems, Trusted: $2}
        idef.WithArgs = 0
        idef.Elems[0].SetLineno(getLineno(lex))
        idef.SetLineno(getLineno(lex))
        id := ast.NewIsolateDecl(idef)
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
        nameAtom := ast.NewAtom($4, $5...)
        elems := append(append([]ast.Node{nameAtom}, $7...), $9...)
        idef := &ast.IsolateDef{Elems: elems, Trusted: $2}
        idef.WithArgs = len($9)
        idef.Elems[0].SetLineno(getLineno(lex))
        idef.SetLineno(getLineno(lex))
        id := ast.NewIsolateDecl(idef)
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
        nameAtom := ast.NewAtom($4)
        createObject($$, nameAtom, $5, objAccum, getLineno(lex), false)
        // Python: ty = TrustedIsolateDef if p[2] else IsolateDef
        // Python: df = ty(*([Atom(p[4],p[5]),Atom(p[4],p[5])]+p[10]))
        a1 := ast.NewAtom($4, $5...)
        a2 := ast.NewAtom($4, $5...)
        elems := append([]ast.Node{a1, a2}, $10...)
        idef := &ast.IsolateDef{Elems: elems, Trusted: $2, IsObject: true}
        idef.WithArgs = len($10)
        idef.Elems[0].SetLineno(getLineno(lex))
        idef.SetLineno(getLineno(lex))
        id := &ast.IsolateObjectDecl{IsolateDecl: *ast.NewIsolateDecl(idef)}
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
        createObject($$, pref, $4, objAccum, getLineno(lex), false)
        // Python: ty = ProcessDef
        // Python: d = IsolateObjectDecl(ty(*([Atom(p[3],p[4]),Atom(p[3],p[4])]+p[9])))
        a1 := ast.NewAtom(pref.Rep, $4...)
        a2 := ast.NewAtom(pref.Rep, $4...)
        elems := append([]ast.Node{a1, a2}, $9...)
        edef := &ast.ExtractDef{IsolateDef: ast.IsolateDef{Elems: elems, IsObject: true}}
        pdef := &ast.ProcessDef{ExtractDef: *edef}
        pdef.WithArgs = len($9) + 1
        pdef.Elems[0].SetLineno(getLineno(lex))
        pdef.SetLineno(getLineno(lex))
        id := &ast.IsolateObjectDecl{IsolateDecl: *ast.NewIsolateDecl(&pdef.IsolateDef)}
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
        parentObject = ""
        // Python: d = IsolateDecl(ExtractDef(*([Atom(p[3],p[4])] + p[6])))
        pref := $3.(*ast.Atom)
        nameAtom := ast.NewAtom(pref.Rep, $4...)
        elems := append([]ast.Node{nameAtom}, $6...)
        edef := &ast.ExtractDef{IsolateDef: ast.IsolateDef{Elems: elems}}
        edef.WithArgs = len($6)
        edef.Elems[0].SetLineno(getLineno(lex))
        edef.SetLineno(getLineno(lex))
        id := ast.NewIsolateDecl(&edef.IsolateDef)
        $$.declare(id)
    }
    // --- Export ---
    | top TOK_EXPORT callatom
    {
        xtracer.Trace("parser.p_top_export_callatom ENTER (top)")
        $$ = $1
        ed := ast.NewExportDecl(&ast.ExportDef{ExportedNode: $3, ScopeNode: ast.NewAtom("")})
        ed.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$.declare(ed)
    }
    // --- Import ---
    | top TOK_IMPORT callatom
    {
        xtracer.Trace("parser.p_top_import_callatom ENTER (top)")
        $$ = $1
        id := ast.NewImportDecl(&ast.ImportDef{Imported: $3, Scope: ast.NewAtom("")})
        id.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
                args[i] = &ast.DelegateDef{Elems: []ast.Node{s, $4}}
            } else {
                args[i] = &ast.DelegateDef{Elems: []ast.Node{s}}
            }
        }
        dd := ast.NewDelegateDecl(args...)
        $$.declare(dd)
    }
    // --- Interpret ---
    | top TOK_INTERPRET oper TOK_ARROW oper
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_symbol ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        imp := &ast.Implies{T1: $3, T2: $5}
        imp.SetLineno(getLineno(lex))
        lf := addLabel(mkLF(imp), "interp")
        d := ast.NewInterpretDecl(lf)
        d.SetLineno(getLineno(lex))
        $$.declare(d)
    }
    // --- Interpret with range ---
    | top TOK_INTERPRET oper TOK_ARROW TOK_LCB term TOK_DOTS term TOK_RCB
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_lcb_symbol_dots_symbol_rcb ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        rng := &ast.Range{Lo: $6, Hi: $8}
        imp := &ast.Implies{T1: $3, T2: rng}
        imp.SetLineno(getLineno(lex))
        lf := addLabel(mkLF(imp), "interp")
        d := ast.NewInterpretDecl(lf)
        d.SetLineno(getLineno(lex))
        $$.declare(d)
    }
    // --- Interpret with enum ---
    | top TOK_INTERPRET oper TOK_ARROW TOK_LCB SYMBOLx moresymbols TOK_RCB
    {
        xtracer.Trace("parser.p_top_interpret_symbol_arrow_lcb_symbol_moresymbols_rcb ENTER (top)")
        $$ = $1
        names := append([]string{$6}, func() []string {
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
            atoms[i] = ast.NewAtom(n)
        }
        es := &ast.EnumeratedSort{Elems: atoms}
        imp := &ast.Implies{T1: $3, T2: es}
        imp.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        lf := addLabel(mkLF(imp), "interp")
        d := ast.NewInterpretDecl(lf)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$.declare(d)
    }
    // --- Alias ---
    | top TOK_ALIAS SYMBOLx TOK_EQ callatom
    {
        xtracer.Trace("parser.p_top_aliase_symbol_eq_callatom ENTER (top)")
        $$ = $1
        d := ast.NewAliasDecl(ast.NewDefinition(ast.NewAtom($3), $5))
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$.declare(d)
    }
    // --- Attribute ---
    | top TOK_ATTRIBUTE callatom TOK_EQ attributeval
    {
        xtracer.Trace("parser.p_top_attribute_callatom_eq_attributeval ENTER (top)")
        $$ = $1
        lex := v17lex.(*v17LexAdapter)
        adef := ast.NewAttributeDef($3, $5)
        adef.SetLineno(getLineno(lex))
        d := ast.NewAttributeDecl(adef)
        d.SetLineno(getLineno(lex))
        $$.declare(d)
    }
    // --- Variant ---
    | top TOK_VARIANT typesymbol TOK_OF atype
    {
        xtracer.Trace("parser.p_top_variant_symbol_of_atype ENTER (top)")
        $$ = $1
        scnst := ast.NewAtom($3.(*ast.Atom).Rep)
        scnst.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        tdfn := &ast.TypeDef{Name: scnst, Value: ast.NewUninterpretedSortAST()}
        tdfn.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        td := ast.NewTypeDecl(tdfn)
        $$.declare(td)
        vdfn := &ast.VariantDef{Name: scnst, VSort: $5}
        vd := ast.NewVariantDecl(vdfn)
        $$.declare(vd)
    }
    // --- Variant with sort ---
    | top TOK_VARIANT typesymbol TOK_OF atype TOK_EQ sort
    {
        xtracer.Trace("parser.p_top_variant_symbol_of_symbol_eq_sort ENTER (top)")
        $$ = $1
        scnst := ast.NewAtom($3.(*ast.Atom).Rep)
        scnst.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        tdfn := &ast.TypeDef{Name: scnst, Value: $7}
        tdfn.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        td := ast.NewTypeDecl(tdfn)
        $$.declare(td)
        vdfn := &ast.VariantDef{Name: scnst, VSort: $5}
        vd := ast.NewVariantDecl(vdfn)
        $$.declare(vd)
    }
    // --- Nativequote ---
    | top TOK_NATIVEQUOTE
    {
        xtracer.Trace("parser.p_top_nativequote ENTER (top)")
        $$ = $1
        // TODO: Python creates NativeDef/NativeDecl with lineno here.
        // defn.lineno = get_lineno(p,2); thing.lineno = get_lineno(p,2)
        // Once NativeDef/NativeDecl types exist, set:
        //   defn.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        //   thing.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        parseNativequote($2, v17lex.(*v17LexAdapter))
    }
    // --- Scenario ---
    | top TOK_SCENARIO TOK_LCB sceninit TOK_SEMI scentranss TOK_RCB
    {
        xtracer.Trace("parser.p_top_scenario_lcb_sceninit_semi_scentranss_rcb ENTER (top)")
        $$ = $1
        elems := append([]ast.Node{$4}, $6...)
        sdef := &ast.ScenarioDef{Elems: elems}
        sdef.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        sd := ast.NewScenarioDecl(sdef)
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
        sd := &ast.StateDef{Name: $3, State: $5}
        _ = sd
    }
    // --- Assert (v1.6 only, kept for compat) ---
    | top TOK_ASSERT SYMBOLx TOK_ARROW assert_rhs
    {
        xtracer.Trace("parser.p_top_assert_symbol_arrow_assert_rhs ENTER (top)")
        $$ = $1
    }
    ;

// ============================================================
// --- SYMBOL handling ---
// ============================================================

SYMBOLx:
    TOK_PRESYMBOL
    {
        xtracer.Trace("parser.p_SYMBOL_PRESYMBOL ENTER (SYMBOL) val=%s", $1)
        $$ = $1
    }
    | SYMBOLx TOK_LB SYMsubscr TOK_RB
    {
        xtracer.Trace("parser.p_SYMBOL_SYMBOL_LB_SYMsubscr_RB ENTER (SYMBOL)")
        $$ = $1 + "[" + $3 + "]"
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
        $$ = "this"
    }
    | SYMsubscr TOK_DOT SYMBOLx
    {
        xtracer.Trace("parser.p_SYMsubscr_SYMsubscr_dot_symbol ENTER (SYMsubscr)")
        $$ = $1 + "." + $3
    }
    ;

// ============================================================
// --- atype (sort names) ---
// ============================================================

atype:
    SYMBOLx
    {
        xtracer.Trace("parser.p_atype_symbol ENTER (atype) val=%s", $1)
        $$ = &ast.Symbol{Rep: $1}
    }
    | atype TOK_DOT SYMBOLx
    {
        xtracer.Trace("parser.p_atype_atype_dot_symbol ENTER (atype)")
        if _, ok := $1.(*ast.This); ok {
            $$ = &ast.Symbol{Rep: $3}
        } else if sym, ok := $1.(*ast.Symbol); ok {
            $$ = &ast.Symbol{Rep: sym.Rep + "." + $3}
        } else {
            $$ = &ast.Symbol{Rep: $3}
        }
    }
    | TOK_THIS
    {
        xtracer.Trace("parser.p_atype_this ENTER (atype)")
        t := &ast.This{}
        t.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        a := ast.NewApp(&ast.Symbol{Rep: $1})
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        xtracer.Trace("parser.p_appelem_appelem_terms ENTER (appelem)")
        // Python: App(p[1], p[3])
        a := ast.NewApp(&ast.Symbol{Rep: $1}, $3...)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        v := &ast.Variable{Rep: $1, VSort: "S"}
        v.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = v
    }
    | TOK_VARIABLE TOK_COLON atype
    {
        xtracer.Trace("parser.p_var_variable_colon_symbol ENTER (var)")
        // Python: Variable(p[1], p[3]) where p[3] is a string from atype
        v := &ast.Variable{Rep: $1, VSort: atypeToString($3)}
        v.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = v
    }
    ;

simplevar:
    TOK_VARIABLE
    {
        xtracer.Trace("parser.p_simplevar_variable ENTER (simplevar)")
        // Python: Variable(p[1], universe) where universe = 'S'
        v := &ast.Variable{Rep: $1, VSort: "S"}
        v.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = v
    }
    | TOK_VARIABLE TOK_COLON SYMBOLx
    {
        xtracer.Trace("parser.p_simplevar_variable_colon_symbol ENTER (simplevar)")
        // Python: Variable(p[1], p[3]) where p[3] is a string
        v := &ast.Variable{Rep: $1, VSort: $3}
        v.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        o := &ast.Old{Term: $2}
        o.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = o
    }
    | term TOK_DOT appelem
    {
        xtracer.Trace("parser.p_term_dot_appelem ENTER (term)")
        lex := v17lex.(*v17LexAdapter)
        switch lhs := $1.(type) {
        case *ast.Atom:
            rhs := $3.(*ast.Atom)
            composed := ast.ComposeAtoms(lhs, rhs)
            composed.SetLineno(getLineno(lex))
            $$ = composed
        case *ast.Old:
            if inner, ok := lhs.Term.(*ast.Atom); ok {
                rhs := $3.(*ast.Atom)
                t := ast.ComposeAtoms(inner, rhs)
                t.SetLineno(getLineno(lex))
                lhs.Term = t
                $$ = lhs
            } else {
                $$ = &ast.MethodCall{Obj: $1, Method: $3}
            }
        default:
            $$ = &ast.MethodCall{Obj: $1, Method: $3}
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
        n := ast.NewApp(ast.NewSymbol("+", nil), $1, $3)
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | term TOK_MINUS term
    {
        xtracer.Trace("parser.p_term_term_MINUS_term ENTER (term)")
        n := ast.NewApp(ast.NewSymbol("-", nil), $1, $3)
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | term TOK_TIMES term
    {
        xtracer.Trace("parser.p_term_term_TIMES_term ENTER (term)")
        n := ast.NewApp(ast.NewSymbol("*", nil), $1, $3)
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | term TOK_DIV term
    {
        xtracer.Trace("parser.p_term_term_DIV_term ENTER (term)")
        n := ast.NewApp(ast.NewSymbol("/", nil), $1, $3)
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    // --- If/else ---
    | term TOK_IF fmla TOK_ELSE term
    {
        xtracer.Trace("parser.p_term_if_fmla_else_term ENTER (term)")
        n := &ast.Ite{Cond: $3, Then: $1, Else: $5}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    // --- Comparison ---
    | term TOK_EQ term
    {
        xtracer.Trace("parser.p_term_term_EQ_term ENTER (term)")
        a := &ast.Atom{Rep: "=", Terms: []ast.Node{$1, $3}}
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | term TOK_LE term
    {
        xtracer.Trace("parser.p_term_term_LE_term ENTER (term)")
        a := &ast.Atom{Rep: "<=", Terms: []ast.Node{$1, $3}}
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | term TOK_LT term
    {
        xtracer.Trace("parser.p_term_term_LT_term ENTER (term)")
        a := &ast.Atom{Rep: "<", Terms: []ast.Node{$1, $3}}
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | term TOK_GE term
    {
        xtracer.Trace("parser.p_term_term_GE_term ENTER (term)")
        a := &ast.Atom{Rep: ">=", Terms: []ast.Node{$1, $3}}
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | term TOK_GT term
    {
        xtracer.Trace("parser.p_term_term_GT_term ENTER (term)")
        a := &ast.Atom{Rep: ">", Terms: []ast.Node{$1, $3}}
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | term TOK_PTO term
    {
        xtracer.Trace("parser.p_term_term_PTO_term ENTER (term)")
        n := ast.NewApp(ast.NewSymbol("*>", nil), $1, $3)
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | term TOK_TILDAEQ term
    {
        xtracer.Trace("parser.p_term_term_tildaeq_term ENTER (term)")
        n := &ast.Not{Body: &ast.Atom{Rep: "=", Terms: []ast.Node{$1, $3}}}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    // --- Boolean ---
    | TOK_TRUE
    {
        xtracer.Trace("parser.p_term_true ENTER (term)")
        n := &ast.And{}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | TOK_FALSE
    {
        xtracer.Trace("parser.p_term_false ENTER (term)")
        n := &ast.Or{}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | TOK_TILDA term
    {
        xtracer.Trace("parser.p_term_not_term ENTER (term)")
        n := &ast.Not{Body: $2}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
            n := &ast.And{Terms: []ast.Node{$1, $3}}
            n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
            n := &ast.Or{Terms: []ast.Node{$1, $3}}
            n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
            $$ = n
        }
    }
    | term TOK_ARROW term
    {
        xtracer.Trace("parser.p_term_term_arrow_term ENTER (term)")
        n := &ast.Implies{T1: $1, T2: $3}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | term TOK_IFF term
    {
        xtracer.Trace("parser.p_term_term_iff_term ENTER (term)")
        n := &ast.Iff{T1: $1, T2: $3}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    // --- Quantifiers ---
    | TOK_FORALL simplevars TOK_DOT term    %prec TOK_SEMI
    {
        xtracer.Trace("parser.p_term_forall_simplevars_dot_term ENTER (term)")
        n := &ast.Forall{Bounds: $2, Body: $4}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | TOK_EXISTS simplevars TOK_DOT term    %prec TOK_SEMI
    {
        xtracer.Trace("parser.p_term_exists_simplevars_dot_term ENTER (term)")
        n := &ast.Exists{Bounds: $2, Body: $4}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | TOK_FORALL TOK_LPAREN vars TOK_RPAREN term
    {
        xtracer.Trace("parser.p_term_forall_lp_vars_lp_term ENTER (term)")
        n := &ast.Forall{Bounds: $3, Body: $5}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | TOK_EXISTS TOK_LPAREN vars TOK_RPAREN term
    {
        xtracer.Trace("parser.p_term_exists_lp_vars_lp_term ENTER (term)")
        n := &ast.Exists{Bounds: $3, Body: $5}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    // --- Temporal ---
    | TOK_GLOBALLY term
    {
        xtracer.Trace("parser.p_term_globally_term ENTER (term)")
        n := &ast.Globally{Body: $2}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | TOK_EVENTUALLY term
    {
        xtracer.Trace("parser.p_term_eventually_term ENTER (term)")
        n := &ast.Eventually{Body: $2}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | term TOK_WHENNEXT term
    {
        xtracer.Trace("parser.p_term_term_whennext_term ENTER (term)")
        n := &ast.WhenOperator{Name: "next", T1: $1, T2: $3}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | term TOK_WHENPREV term
    {
        xtracer.Trace("parser.p_term_term_whenprev_term ENTER (term)")
        n := &ast.WhenOperator{Name: "prev", T1: $1, T2: $3}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | term TOK_WHENFIRST term
    {
        xtracer.Trace("parser.p_term_term_whenfirst_term ENTER (term)")
        n := &ast.WhenOperator{Name: "first", T1: $1, T2: $3}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | term TOK_WHENLAST term
    {
        xtracer.Trace("parser.p_term_term_whenlast_term ENTER (term)")
        n := &ast.WhenOperator{Name: "last", T1: $1, T2: $3}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    // --- ISA ---
    | term TOK_ISA atype
    {
        xtracer.Trace("parser.p_fmla_fmla_isa_atype ENTER (term)")
        $$ = &ast.Isa{Terms: []ast.Node{$1, $3}}
    }
    // --- Sort annotation ---
    | term TOK_COLON atype
    {
        xtracer.Trace("parser.p_term_term_colon_term ENTER (term)")
        if v, ok := $1.(*ast.Variable); ok {
            v.VSort = atypeToString($3)
        }
        $$ = $1
    }
    // --- Named binders ---
    | TOK_LPAREN TOK_DOLLAR SYMBOLx simplevars TOK_DOT fmla TOK_RPAREN TOK_LPAREN terms TOK_RPAREN
    {
        xtracer.Trace("parser.p_term_namedbinder_vars_dot_term ENTER (term)")
        binder := &ast.NamedBinder{Name: $3, Bounds: $4, Body: $6}
        $$ = &ast.Atom{Rep: "", Terms: append([]ast.Node{binder}, $9...)}
    }
    | TOK_DOLLAR SYMBOLx TOK_DOT fmla     %prec TOK_SEMI
    {
        xtracer.Trace("parser.p_term_namedbinder_dot_fmla ENTER (term)")
        $$ = &ast.NamedBinder{Name: $2, Body: $4}
    }
    | TOK_DOLLAR SYMBOLx TOK_DOLLAR fmla   %prec TOK_SEMI
    {
        xtracer.Trace("parser.p_term_namedbinder_dollar_fmla ENTER (term)")
        $$ = &ast.NamedBinder{Name: $2, Body: $4}
    }
    ;

// ============================================================
// --- fmla ---
// ============================================================

fmla:
    term
    {
        xtracer.Trace("parser.p_fmla_term ENTER (fmla)")
        $$ = $1
    }
    ;

// ============================================================
// --- labeledfmla ---
// ============================================================

labeledfmla:
    fmla
    {
        xtracer.Trace("parser.p_labeledfmla_fmla ENTER (labeledfmla)")
        lf := ast.NewLabeledFormula(nil, $1)
        // Python: p[0].lineno = p[1].lineno — copy formula's Location
        lf.SetLineno($1.GetLineno())
        $$ = lf
    }
    | labelname fmla
    {
        xtracer.Trace("parser.p_labeledfmla_label_fmla ENTER (labeledfmla)")
        // Python: Atom(p[1][1:-1],[]) — strip surrounding brackets from label name
        name := $1
        if len(name) >= 2 && name[0] == '[' && name[len(name)-1] == ']' {
            name = name[1 : len(name)-1]
        }
        lf := ast.NewLabeledFormula(ast.NewAtom(name), $2)
        // Python: p[0].lineno = get_lineno(p,1)
        lf.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        // Python: LABEL = p[1] + p[2] + p[3] → "[sym]"
        $$ = "[" + $2 + "]"
    }
    | TOK_LABEL
    {
        xtracer.Trace("parser.p_labelname__label ENTER (labelname)")
        $$ = $1
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
        lf := ast.NewLabeledFormula($1, $2)
        lf.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = &ast.And{} // non-nil marker
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
        $$ = &ast.And{} // non-nil marker
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
        $$ = &ast.And{} // non-nil marker
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
        $$ = ast.NewAtom($1)
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
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        label := ast.NewAtom($2)
        label.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        lf := ast.NewLabeledFormula(label, $3)
        lf.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        xtracer.Trace("parser.p_dotsym_symbol ENTER (dotsym) val=%s", $1)
        _ = getLineno(v17lex.(*v17LexAdapter))
        $$ = $1
    }
    | dotsym TOK_DOT SYMBOLx
    {
        xtracer.Trace("parser.p_dotsym_dotsym_dot_symbol ENTER (dotsym)")
        $$ = $1 + "." + $3
    }
    ;

defnlhs:
    dotsym
    {
        xtracer.Trace("parser.p_defnlhs_symbol ENTER (defnlhs)")
        $$ = ast.NewAtom($1)
    }
    | dotsym TOK_LPAREN defargs TOK_RPAREN
    {
        xtracer.Trace("parser.p_defnlhs_symbol_lparen_defargs_rparen ENTER (defnlhs)")
        a := ast.NewAtom($1)
        a.Terms = $3
        $$ = a
    }
    | TOK_LPAREN defarg relop defarg TOK_RPAREN
    {
        xtracer.Trace("parser.p_defnlhs_lp_term_relop_term_rp ENTER (defnlhs)")
        a := ast.NewAtom($3, $2, $4)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | TOK_LPAREN defarg infix defarg TOK_RPAREN
    {
        xtracer.Trace("parser.p_defnlhs_lp_term_infix_term_rp ENTER (defnlhs)")
        a := ast.NewApp(ast.NewSymbol($3, nil), $2, $4)
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
        text, bqs := parseNativequote($1, v17lex.(*v17LexAdapter))
        elems := append([]ast.Node{ast.NewAtom(text)}, bqs...)
        ne := &ast.NativeExpr{Elems: elems}
        ne.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = ne
    }
    ;

defn:
    typeddefn TOK_EQ defnrhs
    {
        xtracer.Trace("parser.p_defn_atom_fmla ENTER (defn)")
        d := ast.NewDefinition(ast.AppToAtom($1), $3)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = &ast.DefinitionSchema{Definition: *d}
    }
    ;

somevarfmla:
    TOK_SOME simplevar TOK_DOT fmla optin optelse
    {
        xtracer.Trace("parser.p_somevarfmla_some_simplevar_dot_fmla ENTER (somevarfmla)")
        se := &ast.SomeExpr{Param: $2, Fmla: $4}
        if $5 != nil { se.IfValue = $5 }
        if $6 != nil { se.ElseVal = $6 }
        se.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = $1
    }
    | TOK_LCB schdecls schconc TOK_RCB
    {
        xtracer.Trace("parser.p_schdefnrhs_lcb_schdecls_rcb ENTER (schdefnrhs)")
        args := append($2, $3)
        $$ = ast.NewSchemaBody(args...)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        scnst := ast.NewAtom($2)
        scnst.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        tdfn := &ast.TypeDef{Name: scnst, Value: ast.NewUninterpretedSortAST()}
        tdfn.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = []ast.Node{tdfn}
    }
    | optexplicit TOK_PROPERTY lgprop
    {
        xtracer.Trace("parser.p_schdecl_propdecl ENTER (schdecl)")
        lf := addLabel($3.(*ast.LabeledFormula), "prop")
        if $1 != nil {
            lf.Explicit = true
        }
        $$ = []ast.Node{lf}
    }
    | TOK_THEOREM lgprop
    {
        xtracer.Trace("parser.p_schdecl_theorem_lgprop ENTER (schdecl)")
        lf := addLabel($2.(*ast.LabeledFormula), "prop")
        $$ = []ast.Node{lf}
    }
    | schdefnrhs
    {
        xtracer.Trace("parser.p_schdecl_theorem ENTER (schdecl)")
        lf := ast.NewLabeledFormula(nil, $1)
        lf = addLabel(lf, "sch")
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
        $$ = ast.NewDefinition(ast.AppToAtom($1), $3)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        d := ast.NewDestructorDecl($2...)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = d
    }
    | TOK_FIELD tterms
    {
        xtracer.Trace("parser.p_symdecl_field_tterms ENTER (symdecl)")
        // Python: arg0 = Variable('SELF',This()); arg0.lineno = get_lineno(p,1)
        // Python: Variable('SELF', This()) — This() is special; use "this" as sort string
        arg0 := ast.NewVariable("SELF", "this")
        arg0.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        // Python: tterms = [x.clone([arg0]+x.args) for x in p[2]]
        // Python: for x,y in zip(p[2],tterms): y.lineno = x.lineno
        cloned := make([]ast.Node, len($2))
        for i, x := range $2 {
            newArgs := append([]ast.Node{arg0}, x.Args()...)
            cloned[i] = x.Clone(newArgs)
            cloned[i].SetLineno(x.GetLineno())
        }
        d := ast.NewDestructorDecl(cloned...)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = d
    }
    | TOK_CONSTRUCTOR tterms
    {
        xtracer.Trace("parser.p_symdecl_constructor_tterms ENTER (symdecl)")
        // Python: for t in p[2]: if not hasattr(t,'sort'): t.sort = This()
        lex := v17lex.(*v17LexAdapter)
        for _, t := range $2 {
            if app, ok := t.(*ast.App); ok && app.ASort == nil {
                thisNode := &ast.This{}
                thisNode.SetLineno(getLineno(lex))
                app.ASort = thisNode
            }
        }
        d := ast.NewConstructorDecl($2...)
        d.SetLineno(getLineno(lex))
        $$ = d
    }
    ;

constantdecl:
    TOK_INDIV tterms
    {
        xtracer.Trace("parser.p_constantdecl_constant_tterms ENTER (constantdecl)")
        d := ast.NewConstantDecl($2...)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = d
    }
    | TOK_VAR tterms
    {
        xtracer.Trace("parser.p_constantdecl_var_tterms ENTER (constantdecl)")
        d := ast.NewConstantDecl($2...)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        d := ast.NewParameterDecl($1)
        $$ = d
    }
    | tterm TOK_EQ paramval
    {
        xtracer.Trace("parser.p_param_tterm_eq_paramval ENTER (parameter)")
        df := ast.NewDefinition($1, $3)
        df.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        d := ast.NewParameterDecl(df)
        $$ = d
    }
    ;

paramval:
    TOK_TRUE
    {
        xtracer.Trace("parser.p_paramval_true ENTER (paramval)")
        $$ = ast.NewAtom("true")
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_FALSE
    {
        xtracer.Trace("parser.p_paramval_false ENTER (paramval)")
        $$ = ast.NewAtom("false")
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | SYMBOLx
    {
        xtracer.Trace("parser.p_paramval_symbol ENTER (paramval)")
        $$ = ast.NewApp(ast.NewSymbol($1, nil))
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    ;

tapp:
    SYMBOLx
    {
        xtracer.Trace("parser.p_tapp_symbol ENTER (tapp)")
        a := ast.NewApp(ast.NewSymbol($1, nil))
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | SYMBOLx targs
    {
        xtracer.Trace("parser.p_tapp_symbol_targs ENTER (tapp)")
        args := make([]ast.Node, len($2))
        copy(args, $2)
        a := ast.NewApp(ast.NewSymbol($1, nil), args...)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | TOK_LPAREN var infix var TOK_RPAREN
    {
        xtracer.Trace("parser.p_tapp_lp_symbol_infix_symbol_rp ENTER (tapp)")
        a := ast.NewApp(ast.NewSymbol($3, nil), $2, $4)
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
        $$ = ast.NewAtom($1)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | SYMBOLx targs
    {
        xtracer.Trace("parser.p_tatom_symbol_targs ENTER (tatom)")
        $$ = &ast.Atom{Rep: $1, Terms: $2}
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_LPAREN var relop var TOK_RPAREN
    {
        xtracer.Trace("parser.p_tatom_lp_symbol_relop_symbol_rp ENTER (tatom)")
        $$ = ast.NewAtom($3, $2, $4)
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
        // relation declaration (sort = bool)
        d := ast.NewConstantDecl($1)
        $$ = d
    }
    | defn
    {
        xtracer.Trace("parser.p_rel_defn ENTER (rel)")
        lf := addLabel(mkLF($1), "def")
        d := ast.NewDerivedDecl(lf)
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
        d := ast.NewConstantDecl($1)
        $$ = d
    }
    | typeddefn TOK_EQ defnrhs
    {
        xtracer.Trace("parser.p_fun_defn ENTER (fun)")
        df := ast.NewDefinition(ast.AppToAtom($1), $3)
        df.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        lf := addLabel(mkLF(df), "def")
        d := ast.NewDerivedDecl(lf)
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
        $$ = ast.NewAtom($1)
    }
    | TOK_THIS
    {
        xtracer.Trace("parser.p_typesymbol_this ENTER (typesymbol)")
        a := ast.NewAtom("this")
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = &ast.EnumeratedSort{Elems: []ast.Node{ast.NewAtom($2)}}
    }
    | TOK_LCB SYMBOLx TOK_COMMA names TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_names_rcb ENTER (sort)")
        vals := []ast.Node{ast.NewAtom($2)}
        for _, n := range $4 {
            vals = append(vals, n)
        }
        $$ = &ast.EnumeratedSort{Elems: vals}
    }
    | TOK_LCB SYMBOLx TOK_DOTS SYMBOLx TOK_RCB
    {
        xtracer.Trace("parser.p_sort_lcb_symbol_dots_symbol_rcb ENTER (sort)")
        $$ = &ast.Range{Lo: ast.NewAtom($2), Hi: ast.NewAtom($4)}
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_STRUCT TOK_LCB tterms TOK_RCB
    {
        xtracer.Trace("parser.p_sort_struct_lcb_names_rcb ENTER (sort)")
        $$ = &ast.StructSort{Fields: $3}
    }
    | TOK_STRUCT TOK_LCB TOK_RCB
    {
        xtracer.Trace("parser.p_sort_struct_lcb_rcb ENTER (sort)")
        $$ = &ast.StructSort{}
    }
    ;

names:
    SYMBOLx
    {
        xtracer.Trace("parser.p_names_symbol ENTER (names)")
        $$ = []ast.Node{ast.NewAtom($1)}
    }
    | names TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_names_names_comma_symbol ENTER (names)")
        $$ = append($1, ast.NewAtom($3))
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
        a := ast.NewAtom($1)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        xtracer.Trace("parser.p_atom_symbol_lp_terms_rp ENTER (atom)")
        a := &ast.Atom{Rep: $1, Terms: $3}
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = ast.NewApp(ast.NewSymbol($1, nil))
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        xtracer.Trace("parser.p_app_symbol_lp_terms_rp ENTER (app)")
        $$ = ast.NewApp(ast.NewSymbol($1, nil), $3...)
    }
    | term infix term
    {
        xtracer.Trace("parser.p_app_term_infix_term ENTER (app)")
        $$ = ast.NewApp(ast.NewSymbol($2, nil), $1, $3)
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
        $$ = $1
    }
    | SYMBOLx TOK_EQ SYMBOLx
    {
        xtracer.Trace("parser.p_lit_term_eq_term ENTER (lit)")
        $$ = ast.NewAtom("=", ast.NewAtom($1), ast.NewAtom($3))
    }
    | SYMBOLx TOK_TILDAEQ SYMBOLx
    {
        xtracer.Trace("parser.p_lit_term_tildaeq_term ENTER (lit)")
        $$ = &ast.Not{Body: ast.NewAtom("=", ast.NewAtom($1), ast.NewAtom($3))}
    }
    | TOK_TILDA lit
    {
        xtracer.Trace("parser.p_lit_tilda_atom ENTER (lit)")
        $$ = &ast.Not{Body: $2}
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
        a := ast.NewAtom("this")
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | TOK_METHOD
    {
        xtracer.Trace("parser.p_callatom_method ENTER (callatom)")
        $$ = ast.NewAtom("method")
    }
    | callatom TOK_DOT callatom
    {
        xtracer.Trace("parser.p_callatom_callatom_dot_callatom ENTER (callatom)")
        lhs := $1.(*ast.Atom)
        rhs := $3.(*ast.Atom)
        $$ = ast.ComposeAtoms(lhs, rhs)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = ast.NewAtom("object")
    }
    | TOK_ISOLATE
    {
        xtracer.Trace("parser.p_modcat_isolate ENTER (modcat)")
        $$ = ast.NewAtom("isolate")
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
        parentObject = ""
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
        $$ = ast.NewAtom($1)
        // Python: global parent_object; parent_object = p[0]
        parentObject = $1
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
        a := ast.NewApp(ast.NewSymbol($1, nil))
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        a.ASort = &ast.Symbol{Rep: $3}
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
        a := &ast.Atom{Rep: $1}
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        a.ASort = $3
        $$ = a
    }
    | TOK_CARET SYMBOLx TOK_COLON atype
    {
        xtracer.Trace("parser.p_lparam_caret_variable_colon_symbol ENTER (lparam)")
        a := &ast.Atom{Rep: $2}
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
        $$ = &ast.Sequence{}
    }
    | TOK_EQ topseq
    {
        xtracer.Trace("parser.p_optactiondef_eq_topseq ENTER (optactiondef)")
        $$ = $2
    }
    | TOK_EQ TOK_TIMES
    {
        xtracer.Trace("parser.p_optactiondef_eq_symbol ENTER (optactiondef)")
        $$ = &ast.CrashAction{}
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
        parseNativequote($2, v17lex.(*v17LexAdapter))
        na := ast.NewAtom("native")
        na.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
    }
    | TOK_IMPLEMENTATION
    {
        xtracer.Trace("parser.p_specimpl_implementation ENTER (specimpl)")
        $$ = "impl"
    }
    | TOK_PRIVATE
    {
        xtracer.Trace("parser.p_specimpl_private ENTER (specimpl)")
        $$ = "private"
    }
    | TOK_GLOBAL
    {
        xtracer.Trace("parser.p_specimpl_global ENTER (specimpl)")
        $$ = "global"
    }
    | TOK_COMMON
    {
        xtracer.Trace("parser.p_specimpl_common ENTER (specimpl)")
        $$ = "common"
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
        n := &ast.Instantiation{Name: nil, Sort: ast.AppToAtom($1)}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | modinst TOK_COLON modinst
    {
        xtracer.Trace("parser.p_inst_atom_colon_modinst ENTER (inst)")
        inst := &ast.Instantiation{Name: ast.AppToAtom($1), Sort: ast.AppToAtom($3)}
        inst.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = inst
    }
    ;

modinst:
    dotsym
    {
        xtracer.Trace("parser.p_modinst_symbol ENTER (modinst)")
        $$ = ast.NewAtom($1)
    }
    | dotsym TOK_LPAREN pnames TOK_RPAREN
    {
        xtracer.Trace("parser.p_modinst_symbol_lp_pnames_rp ENTER (modinst)")
        a := ast.NewAtom($1)
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
        n := ast.NewApp(ast.NewSymbol(rep, nil))
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = ast.NewApp(ast.NewSymbol($1, nil))
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | relop
    {
        xtracer.Trace("parser.p_pname_relop ENTER (pname)")
        $$ = ast.NewApp(ast.NewSymbol($1, nil))
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_THIS
    {
        xtracer.Trace("parser.p_pname_this ENTER (pname)")
        $$ = ast.NewApp(ast.NewSymbol("this", nil))
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_TRUE
    {
        xtracer.Trace("parser.p_pname_true ENTER (pname)")
        $$ = ast.NewAtom("true")
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_FALSE
    {
        xtracer.Trace("parser.p_pname_false ENTER (pname)")
        $$ = ast.NewAtom("false")
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = $1
    }
    | relop
    {
        xtracer.Trace("parser.p_oper_relop ENTER (oper)")
        $$ = ast.NewAtom($1)
    }
    | infix
    {
        xtracer.Trace("parser.p_oper_infix ENTER (oper)")
        $$ = ast.NewAtom($1)
    }
    | TOK_NATIVEQUOTE
    {
        xtracer.Trace("parser.p_oper_nativequote ENTER (oper)")
        text, bqs := parseNativequote($1, v17lex.(*v17LexAdapter))
        elems := append([]ast.Node{ast.NewAtom(text)}, bqs...)
        nt := &ast.NativeType{Elems: elems}
        nt.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = ast.NewAtom("true")
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_FALSE
    {
        xtracer.Trace("parser.p_top_attributeval_false ENTER (attributeval)")
        $$ = ast.NewAtom("false")
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = append($1, ast.NewAtom($3))
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
        $$ = &ast.Sequence{}
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_LCB actseq TOK_RCB
    {
        xtracer.Trace("parser.p_sequence_lcb_actseq_rcb ENTER (sequence)")
        stmts := lowerVarStmts($2)
        seq := lalrMakeSequence(stmts)
        if s, ok := seq.(*ast.Sequence); ok {
            s.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        stmts := lowerVarStmts($2)
        seq := lalrMakeSequence(stmts)
        seq.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        lf := addLabel($2.(*ast.LabeledFormula), "asrt")
        $$ = ast.NewAtom("assume", checkNonTemporal(lf))
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | optunprovable TOK_ASSERT labeledfmla
    {
        xtracer.Trace("parser.p_action_assert ENTER (simpleact)")
        lf := addLabel($3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := ast.NewAtom("assert", lf)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | optunprovable TOK_ASSERT labeledfmla TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_action_assert_proof_proofstep ENTER (simpleact)")
        // Python: AssertAction(check_non_temporal(addlabel(p[3],'asrt')),p[5])
        lf := addLabel($3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := ast.NewAtom("assert", lf, $5)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | optunprovable TOK_REQUIRE labeledfmla
    {
        xtracer.Trace("parser.p_action_require ENTER (simpleact)")
        // Python: RequiresAction(check_non_temporal(addlabel(p[3],'asrt')))
        lf := addLabel($3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := ast.NewAtom("require", lf)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | optunprovable TOK_REQUIRE labeledfmla TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_action_require_proof_proofstep ENTER (simpleact)")
        // Python: RequiresAction(check_non_temporal(addlabel(p[3],'asrt')),p[5])
        lf := addLabel($3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := ast.NewAtom("require", lf, $5)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | optunprovable TOK_ENSURE labeledfmla
    {
        xtracer.Trace("parser.p_action_ensure ENTER (simpleact)")
        // Python: EnsuresAction(check_non_temporal(addlabel(p[3],'asrt')))
        lf := addLabel($3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := ast.NewAtom("ensure", lf)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | optunprovable TOK_ENSURE labeledfmla TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_action_ensure_proof_proofstep ENTER (simpleact)")
        // Python: EnsuresAction(check_non_temporal(addlabel(p[3],'asrt')),p[5])
        lf := addLabel($3.(*ast.LabeledFormula), "asrt")
        lf = checkNonTemporal(lf).(*ast.LabeledFormula)
        addUnprovable(lf, $1)
        a := ast.NewAtom("ensure", lf, $5)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | term TOK_ASSIGN fmla
    {
        xtracer.Trace("parser.p_action_term_assign_fmla ENTER (simpleact)")
        a := ast.NewAtom(":=", $1, checkNonTemporal($3))
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = a
    }
    | termtuple TOK_ASSIGN callatom
    {
        xtracer.Trace("parser.p_action_termtuple_assign_fmla ENTER (simpleact)")
        // Python: CallAction(*([p[3]]+list(p[1].args)))
        callArgs := append([]ast.Node{$3}, $1.Args()...)
        $$ = ast.NewCallAction(callArgs...)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | term TOK_ASSIGN TOK_TIMES
    {
        xtracer.Trace("parser.p_action_term_assign_times ENTER (simpleact)")
        $$ = ast.NewAtom("havoc", $1)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_VAR tterm optinit
    {
        // Python: p_action_var_opttypedsym_assign_fmla (ivy_parser.py:3183-3187)
        xtracer.Trace("parser.p_action_var_opttypedsym_assign_fmla ENTER (simpleact)")
        if $3 != nil {
            $$ = ast.NewAtom("var", $2, $3)
        } else {
            $$ = ast.NewAtom("var", $2)
        }
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_CALL optactualreturns callatom
    {
        xtracer.Trace("parser.p_action_call_optreturns_callatom ENTER (simpleact)")
        // Python: CallAction(*([p[3]] + p[2]))
        callArgs := append([]ast.Node{$3}, $2...)
        $$ = ast.NewCallAction(callArgs...)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_CALL callatom
    {
        xtracer.Trace("parser.p_action_call_callatom ENTER (simpleact)")
        // Python: CallAction(p[2])
        $$ = ast.NewCallAction($2)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_SET lit
    {
        xtracer.Trace("parser.p_action_set_lit ENTER (simpleact)")
        $$ = ast.NewAtom("set", $2)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_INSTANTIATE callatom
    {
        xtracer.Trace("parser.p_action_instantiate_atom ENTER (simpleact)")
        $$ = ast.NewAtom("instantiate", $2)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_DEBUG SYMBOLx optdebugargs
    {
        xtracer.Trace("parser.p_simpleact_debug_symbol_optdebugargs ENTER (simpleact)")
        args := append([]ast.Node{ast.NewAtom($2)}, $3...)
        $$ = ast.NewAtom("debug", args...)
    }
    | term     %prec TOK_SEMI
    {
        xtracer.Trace("parser.p_action_term ENTER (simpleact)")
        // Python: p[0] = CallAction(p[1]); p[0].lineno = p[1].lineno
        $$ = ast.NewCallAction($1)
        $$.SetLineno($1.GetLineno())
    }
    ;

termtuple:
    TOK_LPAREN term TOK_COMMA terms TOK_RPAREN
    {
        xtracer.Trace("parser.p_termtuple_lp_term_comma_terms_rp ENTER (termtuple)")
        args := append([]ast.Node{$2}, $4...)
        t := &ast.Tuple{Elems: args}
        t.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = t
    }
    ;

debugarg:
    SYMBOLx TOK_EQ fmla
    {
        xtracer.Trace("parser.p_debugarg_symbol_equal_fmla ENTER (debugarg)")
        $$ = ast.NewDefinition(ast.NewApp(ast.NewSymbol($1, nil)), $3)
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
        cond := checkNonTemporal($2)
        body := fixIfPart(cond, $3)
        ite := ast.NewIte(cond, body, &ast.Sequence{})
        ite.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = ite
    }
    | TOK_IF somefmla sequence TOK_ELSE action
    {
        xtracer.Trace("parser.p_action_if_somefmla_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
        cond := checkNonTemporal($2)
        body := fixIfPart(cond, $3)
        ite := ast.NewIte(cond, body, $5)
        ite.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = ite
    }
    | TOK_IF TOK_TIMES sequence TOK_ELSE action
    {
        xtracer.Trace("parser.p_action_if_times_lcb_action_rcb_else_LCB_action_RCB ENTER (complexact)")
        choice := ast.NewIte(ast.NewSymbol("*", nil), $3, $5)
        choice.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = choice
    }
    | TOK_WHILE somefmla invariants decreases sequence
    {
        xtracer.Trace("parser.p_action_while_somefmla_invariants_decreases_lcb_action_rcb ENTER (complexact)")
        cond := checkNonTemporal($2)
        args := []ast.Node{cond, $5}
        args = append(args, $3...)
        args = append(args, $4...)
        w := ast.NewAtom("while", args...)
        w.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = w
    }
    | TOK_FOR tterm TOK_COMMA tterm TOK_IN fmla invariants decreases sequence
    {
        xtracer.Trace("parser.p_action_for_tterm_comma_tterm_in_expr_invariants_decreases_lcb_action_rcb ENTER (complexact)")
        $$ = ast.NewAtom("for", $2, $4, $6, $9)
    }
    | TOK_LOCAL lparams sequence
    {
        xtracer.Trace("parser.p_action_local_params_lcb_action_rcb ENTER (complexact)")
        args := append($2, $3)
        la := ast.NewAtom("local", args...)
        la.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = la
    }
    | TOK_LET eqns sequence
    {
        xtracer.Trace("parser.p_action_let_eqns_lcb_action_rcb ENTER (complexact)")
        args := append($2, $3)
        $$ = ast.NewAtom("let", args...)
    }
    | TOK_THUNK labelname SYMBOLx optargs TOK_COLON atype TOK_ASSIGN sequence
    {
        xtracer.Trace("parser.p_action_thunk_symbol_optargs_colon_atype_assign_sequence ENTER (complexact)")
        $$ = ast.NewAtom("thunk", ast.NewAtom($2), ast.NewAtom($3), $8)
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
        $$ = ast.NewAtom("some_assign", $1, $3)
    }
    | TOK_SOME bounds fmla
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla ENTER (somefmla)")
        args := append($2, $3)
        sa := ast.NewAtom("some", args...)
        sa.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = sa
    }
    | TOK_SOME bounds fmla TOK_MINIMIZING term
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla_minimizing_term ENTER (somefmla)")
        args := append($2, $3, $5)
        smin := ast.NewAtom("some_min", args...)
        smin.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = smin
    }
    | TOK_SOME bounds fmla TOK_MAXIMIZING term
    {
        xtracer.Trace("parser.p_somefmla_some_bounds_fmla_maximizing_term ENTER (somefmla)")
        args := append($2, $3, $5)
        smax := ast.NewAtom("some_max", args...)
        smax.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = append($1, $3)
    }
    | invariants TOK_INVARIANT labeledfmla TOK_PROOF proofstep
    {
        xtracer.Trace("parser.p_invariant_invariant_fmla_proof ENTER (invariants)")
        $$ = append($1, $3, $5)
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
        $$ = []ast.Node{$2}
    }
    ;

// --- eqn / eqns ---

eqn:
    SYMBOLx TOK_EQ SYMBOLx
    {
        xtracer.Trace("parser.p_eqn_SYMBOL_EQ_SYMBOL ENTER (eqn)")
        $$ = ast.NewDefinition(ast.NewApp(ast.NewSymbol($1, nil)), ast.NewApp(ast.NewSymbol($3, nil)))
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
        pl := &ast.PlaceList{Elems: $2}
        pl.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = pl
    }
    ;

places:
    SYMBOLx
    {
        xtracer.Trace("parser.p_places_symbol ENTER (places)")
        a := ast.NewAtom($1)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = []ast.Node{a}
    }
    | places TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_places_places_comma_symbol ENTER (places)")
        a := ast.NewAtom($3)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        from := &ast.PlaceList{Elems: $2}
        to := &ast.PlaceList{Elems: $4}
        tr := &ast.ScenarioTransition{From: from, To: to, Action: $6}
        tr.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = append($1, tr)
    }
    | scentranss places TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentranss_scentranss_places_colon_scenariomixin ENTER (scentranss)")
        to := &ast.PlaceList{Elems: $2}
        tr := &ast.ScenarioTransition{From: nil, To: to, Action: $4}
        tr.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = append($1, tr)
    }
    ;

scentrans:
    places TOK_ARROW places TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentrans__places_arrow_places_colon_scenariomixin ENTER (scentrans)")
        from := &ast.PlaceList{Elems: $1}
        to := &ast.PlaceList{Elems: $3}
        $$ = &ast.ScenarioTransition{From: from, To: to, Action: $5}
    }
    | places TOK_COLON scenariomixin
    {
        xtracer.Trace("parser.p_scentrans__places_colon_scenariomixin ENTER (scentrans)")
        from := &ast.PlaceList{Elems: $1}
        $$ = &ast.ScenarioTransition{From: from, To: &ast.PlaceList{}, Action: $3}
    }
    ;

scenariomixin:
    TOK_BEFORE atype optargs optreturns sequence
    {
        xtracer.Trace("parser.p_scenariomixin_before_callatom_lcb_action_rcb ENTER (scenariomixin)")
        atom := ast.NewAtom($2.(*ast.Symbol).Rep)
        atom.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        lalrLabelCounter++
        mixerName := fmt.Sprintf("%s[before%d]", atom.Rep, lalrLabelCounter)
        mixer := ast.NewAtom(mixerName)
        adef := &ast.ActionDef{Name: atom, Body: $5, FormalParams: $3, FormalReturns: $4}
        sbm := &ast.ScenarioBeforeMixin{Mixer: mixer, Def: adef}
        sbm.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = sbm
    }
    | TOK_AFTER atype optargs optreturns sequence
    {
        xtracer.Trace("parser.p_scenariomixin_after_callatom_lcb_action_rcb ENTER (scenariomixin)")
        atom := ast.NewAtom($2.(*ast.Symbol).Rep)
        atom.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        lalrLabelCounter++
        mixerName := fmt.Sprintf("%s[after%d]", atom.Rep, lalrLabelCounter)
        mixer := ast.NewAtom(mixerName)
        adef := &ast.ActionDef{Name: atom, Body: $5, FormalParams: $3, FormalReturns: $4}
        sam := &ast.ScenarioAfterMixin{Mixer: mixer, Def: adef}
        sam.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        d := ast.NewDefinition($1, $3)
        d.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = addLabel($2.(*ast.LabeledFormula), "invar")
    }
    | TOK_DEFINITION typeddefn TOK_EQ fmla
    {
        xtracer.Trace("parser.p_tacticwithelem_fun_defn ENTER (tacticwithelem)")
        df := ast.NewDefinition(ast.AppToAtom($2), $4)
        df.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        lf := addLabel(mkLF(df), "def")
        $$ = ast.NewDerivedDecl(lf)
    }
    | TOK_TRIGGER atype TOK_WITH terms
    {
        xtracer.Trace("parser.p_tacticwithelem_trigger ENTER (tacticwithelem)")
        $$ = &ast.Trigger{Terms: append([]ast.Node{$2}, $4...)}
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
        $$ = &ast.TacticWith{Elems: $1}
    }
    | pflets
    {
        xtracer.Trace("parser.p_tacticwithlistchoice_pflets ENTER (tacticwithlistchoice)")
        $$ = &ast.TacticLets{Lets: $1}
    }
    ;

opttacticwith:
    /* empty */
    {
        xtracer.Trace("parser.p_opttacticwith ENTER (opttacticwith)")
        $$ = &ast.TacticWith{}
    }
    | TOK_WITH tacticwithlistchoice
    {
        xtracer.Trace("parser.p_opttacticwith_with_tacticwithlist ENTER (opttacticwith)")
        $$ = $2
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | TOK_WITH TOK_LCB tacticwithlist TOK_RCB
    {
        xtracer.Trace("parser.p_opttacticwith_with_lcb_tacticwithlist_rcb ENTER (opttacticwith)")
        tw := &ast.TacticWith{Elems: $3}
        tw.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = &ast.NullTactic{}
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = &ast.ComposeTactics{Tactics: []ast.Node{$1, $3}}
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = ast.NewDefinition($1, checkNonTemporal($3))
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = ast.NewDefinition(&ast.Variable{Rep: $3, VSort: "S"}, &ast.Variable{Rep: $1, VSort: "S"})
    }
    | SYMBOLx TOK_DIV SYMBOLx
    {
        xtracer.Trace("parser.p_renamingitem_symbol_div_symbol ENTER (renamingitem)")
        $$ = ast.NewDefinition(ast.NewAtom($3), ast.NewAtom($1))
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
        $$ = &ast.Renaming{}
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
        r := &ast.Renaming{Elems: $2}
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
        $$ = &ast.UnfoldSpec{DefName: $1, Renamings: $2}
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
        a := atypeToAtom($2)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        si := &ast.SchemaInstantiation{SchemaName: a, Ren: $3}
        si.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = si
    }
    | TOK_APPLY atype optrenaming TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_symbol_with_defns ENTER (proofstep)")
        a := atypeToAtom($2)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        si := &ast.SchemaInstantiation{SchemaName: a, Ren: $3, Matches: $5}
        si.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = si
    }
    | TOK_ASSUME atype optrenaming
    {
        xtracer.Trace("parser.p_proofstep_assume ENTER (proofstep)")
        // Python: AssumeGlobalTactic(a, p[3]); p[0].label = NoneAST()
        a := atypeToAtom($2)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        at := &ast.AssumeGlobalTactic{AssumeTactic: ast.AssumeTactic{SchemaName: a, Ren: $3}}
        at.TLabel = &ast.NoneAST{}
        at.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = at
    }
    | TOK_ASSUME atype optrenaming TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_assume_with_defns ENTER (proofstep)")
        // Python: AssumeGlobalTactic(*([a,p[3]]+p[5])); p[0].label = NoneAST()
        a := atypeToAtom($2)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        at := &ast.AssumeGlobalTactic{AssumeTactic: ast.AssumeTactic{SchemaName: a, Ren: $3, Matches: $5}}
        at.TLabel = &ast.NoneAST{}
        at.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = at
    }
    | TOK_INSTANTIATE atype optrenaming
    {
        xtracer.Trace("parser.p_proofstep_instantiate ENTER (proofstep)")
        a := atypeToAtom($2)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        at := &ast.AssumeTactic{SchemaName: a, Ren: $3}
        at.TLabel = &ast.NoneAST{}
        at.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = at
    }
    | TOK_INSTANTIATE labelname atype optrenaming
    {
        xtracer.Trace("parser.p_proofstep_instance ENTER (proofstep)")
        lex := v17lex.(*v17LexAdapter)
        a := atypeToAtom($3)
        a.SetLineno(getLineno(lex))
        label := ast.NewAtom($2)
        label.SetLineno(getLineno(lex))
        at := &ast.AssumeTactic{SchemaName: a, Ren: $4}
        at.TLabel = label
        at.SetLineno(getLineno(lex))
        $$ = at
    }
    | TOK_INSTANTIATE labelname atype optrenaming TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_instance_with_matches ENTER (proofstep)")
        lex := v17lex.(*v17LexAdapter)
        a := atypeToAtom($3)
        a.SetLineno(getLineno(lex))
        label := ast.NewAtom($2)
        label.SetLineno(getLineno(lex))
        at := &ast.AssumeTactic{SchemaName: a, Ren: $4, Matches: $6}
        at.TLabel = label
        at.SetLineno(getLineno(lex))
        $$ = at
    }
    | TOK_INSTANTIATE atype optrenaming TOK_WITH matches
    {
        xtracer.Trace("parser.p_proofstep_instantiate_with_defns ENTER (proofstep)")
        a := atypeToAtom($2)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        at := &ast.AssumeTactic{SchemaName: a, Ren: $3, Matches: $5}
        at.TLabel = &ast.NoneAST{}
        at.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = at
    }
    | TOK_INSTANTIATE TOK_WITH pflets
    {
        xtracer.Trace("parser.p_proofstep_witness_pflets ENTER (proofstep)")
        wt := &ast.WitnessTactic{Witnesses: $3}
        wt.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = wt
    }
    | TOK_SHOWGOALS
    {
        xtracer.Trace("parser.p_proofstep_showgoals ENTER (proofstep)")
        sg := &ast.ShowGoalsTactic{}
        sg.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = sg
    }
    | TOK_DEFERGOAL
    {
        xtracer.Trace("parser.p_proofstep_defergoal ENTER (proofstep)")
        dg := &ast.DeferGoalTactic{}
        dg.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = dg
    }
    | TOK_SPOIL atype
    {
        xtracer.Trace("parser.p_proofstep_spoil_atype ENTER (proofstep)")
        // Python: a = Atom(p[2]) where p[2] is a string from atype
        a := atypeToAtom($2)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        st := &ast.SpoilTactic{Target: a}
        st.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = st
    }
    | TOK_TACTIC SYMBOLx opttacticwith optproofgroup
    {
        xtracer.Trace("parser.p_proofstep_tactic ENTER (proofstep)")
        a := ast.NewAtom($2)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        proof := ast.Node($4)
        if proof == nil {
            proof = &ast.NoneAST{}
        }
        tt := &ast.TacticTactic{TName: a, Body: $3, Proof: proof}
        tt.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = tt
    }
    | opttemporal TOK_PROPERTY labeledfmla optskolem optproofgroup
    {
        xtracer.Trace("parser.p_proofstep_property ENTER (proofstep)")
        lf := addLabel($3.(*ast.LabeledFormula), "prop")
        // Python: prop = addtemporal(lf) if p[1] else check_non_temporal(lf)
        if $1 != nil {
            lf = addTemporal(lf)
        } else {
            checkNonTemporal(lf)
        }
        name := ast.Node($4)
        if name == nil { name = &ast.NoneAST{} }
        proof := ast.Node($5)
        if proof == nil { proof = &ast.NoneAST{} }
        pt := &ast.PropertyTactic{Prop: lf, PName: name, Proof: proof}
        pt.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = pt
    }
    | TOK_FUNCTION funs
    {
        xtracer.Trace("parser.p_proofstep_function ENTER (proofstep)")
        ft := &ast.FunctionTactic{Elems: $2}
        ft.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = ft
    }
    | TOK_THEOREM lgprop optproofgroup
    {
        xtracer.Trace("parser.p_proofstep_theorem ENTER (proofstep)")
        lf := addLabel($2.(*ast.LabeledFormula), "thm")
        proof := ast.Node($3)
        if proof == nil { proof = &ast.NoneAST{} }
        pt := &ast.PropertyTactic{Prop: lf, PName: &ast.NoneAST{}, Proof: proof}
        pt.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = pt
    }
    | TOK_PROOF labelname proofgroup
    {
        xtracer.Trace("parser.p_proofstep_proof ENTER (proofstep)")
        lex := v17lex.(*v17LexAdapter)
        label := ast.NewAtom($2)
        label.SetLineno(getLineno(lex))
        pt := &ast.ProofTactic{TLabel: label, Proof: $3}
        pt.SetLineno(getLineno(lex))
        $$ = pt
    }
    | TOK_LET pflets
    {
        xtracer.Trace("parser.p_proofstep_let_pflets ENTER (proofstep)")
        lt := &ast.LetTactic{Defs: $2}
        lt.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = lt
    }
    | TOK_IF fmla proofgroup TOK_ELSE proofgroup
    {
        xtracer.Trace("parser.p_proofstep_if_fmla_proofgroup_else_proofgroup ENTER (proofstep)")
        it := &ast.IfTactic{Cond: $2, Then: $3, Else: $5}
        it.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = it
    }
    | TOK_UNFOLD atype TOK_WITH unfspecs
    {
        xtracer.Trace("parser.p_proofstep_unfold_atype_with_defns ENTER (proofstep)")
        a := atypeToAtom($2)
        a.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        ut := &ast.UnfoldTactic{Premise: a, UnfSpecs: $4}
        ut.TLabel = &ast.NoneAST{}
        ut.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = ut
    }
    | TOK_UNFOLD TOK_WITH unfspecs
    {
        xtracer.Trace("parser.p_proofstep_unfold_with_defns ENTER (proofstep)")
        ut := &ast.UnfoldTactic{Premise: &ast.NoneAST{}, UnfSpecs: $3}
        ut.TLabel = &ast.NoneAST{}
        ut.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = ut
    }
    | TOK_FORGET callatoms
    {
        xtracer.Trace("parser.p_proofstep_forget_callatoms ENTER (proofstep)")
        ft := &ast.ForgetTactic{Names: $2}
        ft.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = &ast.And{}
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
        $$ = &ast.And{} // empty modifies
    }
    | TOK_MODIFIES TOK_TIMES
    {
        xtracer.Trace("parser.p_modifies_modofies_times ENTER (modifies)")
        $$ = nil
    }
    | TOK_MODIFIES atoms
    {
        xtracer.Trace("parser.p_modifies_modifies_atoms ENTER (modifies)")
        $$ = &ast.And{Terms: $2}
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
        $$ = ast.NewAtom("upax", $6, $7)
    }
    ;

assert_rhs:
    TOK_LCB requires modifies ensures TOK_RCB
    {
        xtracer.Trace("parser.p_assert_rhs_lcb_requires_modifies_ensures_rcb ENTER (assert_rhs)")
        $$ = ast.NewAtom("rme", $2, $4)
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
        $$ = &ast.And{}
    }
    | TOK_FALSE
    {
        xtracer.Trace("parser.p_state_expr_false ENTER (state_expr)")
        $$ = &ast.Or{}
    }
    | SYMBOLx
    {
        xtracer.Trace("parser.p_state_expr_symbol ENTER (state_expr)")
        $$ = ast.NewAtom($1)
    }
    | SYMBOLx TOK_LPAREN state_expr TOK_RPAREN
    {
        xtracer.Trace("parser.p_state_expr_symbol_lparen_state_expr_rparen ENTER (state_expr)")
        $$ = ast.NewAtom($1, $3)
    }
    | state_expr TOK_OR state_expr
    {
        xtracer.Trace("parser.p_state_expr_state_expr_or_state_expr ENTER (state_expr)")
        $$ = &ast.Or{Terms: []ast.Node{$1, $3}}
    }
    | TOK_LCB requires modifies ensures TOK_RCB
    {
        xtracer.Trace("parser.p_state_expr_lcb_requires_modifies_ensures_rcb ENTER (state_expr)")
        $$ = ast.NewAtom("rme", $2, $4)
    }
    | TOK_ENTRY
    {
        xtracer.Trace("parser.p_state_expr_entry ENTER (state_expr)")
        $$ = ast.NewAtom("entry")
    }
    ;

// --- Concept space ---

cdefn:
    atom TOK_EQ expr
    {
        xtracer.Trace("parser.p_cdefn_atom_expr ENTER (cdefn)")
        $$ = ast.NewDefinition(ast.AppToAtom($1), $3)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
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
        $$ = ast.NewAtom($2, $1, $3)
        $$.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
    }
    | exprterm TOK_TILDAEQ exprterm
    {
        xtracer.Trace("parser.p_expr_exprterm_tildaeq_exprterm ENTER (expr)")
        n := &ast.Not{Body: ast.NewAtom("=", $1, $3)}
        n.SetLineno(getLineno(v17lex.(*v17LexAdapter)))
        $$ = n
    }
    | TOK_TILDA expr
    {
        xtracer.Trace("parser.p_expr_tilda_atom ENTER (expr)")
        $$ = &ast.Not{Body: $2}
    }
    | TOK_LPAREN expr TOK_RPAREN
    {
        xtracer.Trace("parser.p_expr_lparen_expr_rparen ENTER (expr)")
        $$ = $2
    }
    | prod
    {
        xtracer.Trace("parser.p_expr_prod ENTER (expr)")
        $$ = ast.NewAtom("product", $1...)
    }
    | sum
    {
        xtracer.Trace("parser.p_expr_sum ENTER (expr)")
        $$ = ast.NewAtom("sum", $1...)
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
        $$ = $1
    }
    ;

// --- symbols list ---

symbols:
    SYMBOLx
    {
        xtracer.Trace("parser.p_symbols ENTER (symbols)")
        $$ = []ast.Node{ast.NewAtom($1)}
    }
    | symbols TOK_COMMA SYMBOLx
    {
        xtracer.Trace("parser.p_symbols_symbols_symbol ENTER (symbols)")
        $$ = append($1, ast.NewAtom($3))
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
func lalrMakeSequence(stmts []ast.Node) ast.Node {
	if len(stmts) == 0 {
		return &ast.Sequence{}
	}
	if len(stmts) == 1 {
		return stmts[0]
	}
	stmts = lowerVarStmts(stmts)
	return &ast.Sequence{Stmts: stmts}
}
