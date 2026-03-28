// grammar_v17.y — goyacc LALR(1) grammar for Ivy formula/term/action parsing, version 1.7+.
// Mechanically translated from Python PLY grammar in ivy_logic_parser.py and ivy_parser.py.
//
// In v1.7+, terms subsume formulas: comparison operators, boolean connectives,
// quantifiers, and temporal operators are all term-level constructs.
//
// This grammar also includes action productions for cross-validation of
// action body parsing (scenario mixins, before/after bodies, etc.).
//
// Precedence table (copied exactly from Python ivy_parser.py, v1.7+):
//   SEMI < GLOBALLY/EVENTUALLY < ARROW/IFF < OR < AND < TILDA
//   < EQ/LE/LT/GE/GT/PTO < TILDAEQ < IF < ELSE < COLON
//   < PLUS/MINUS < TIMES/DIV < DOLLAR < OLD < DOT

%{
package lalr_logicparser

import (
	"fmt"
	"github.com/glycerine/goivy/ast"
)

// labelCounter is a package-level counter for generating unique mixer names.
var lalrLabelCounter int

// acfg extracts the *ast.AstConfig from the lexer for use in grammar actions.
func acfg(lex v17Lexer) *ast.AstConfig {
	return lex.(*v17LexAdapter).cfg
}

// atypeToString extracts the string sort name from an atype Node.
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

%}

// The union type for semantic values.
%union {
	node     ast.Node
	nodes    []ast.Node
	str      string
}

// Terminal tokens.
%token <str>  TOK_SYMBOL TOK_VARIABLE TOK_PRESYMBOL
%token        TOK_LPAREN TOK_RPAREN TOK_LB TOK_RB TOK_LCB TOK_RCB
%token        TOK_COMMA TOK_SEMI TOK_COLON TOK_DOT
%token        TOK_PLUS TOK_MINUS TOK_TIMES TOK_DIV
%token        TOK_EQ TOK_TILDAEQ TOK_TILDA TOK_LE TOK_LT TOK_GE TOK_GT
%token        TOK_AND TOK_OR TOK_ARROW TOK_IFF
%token        TOK_PTO TOK_DOLLAR
%token        TOK_FORALL TOK_EXISTS
%token        TOK_TRUE TOK_FALSE
%token        TOK_OLD TOK_THIS TOK_ISA
%token        TOK_IF TOK_ELSE
%token        TOK_GLOBALLY TOK_EVENTUALLY
%token        TOK_WHENNEXT TOK_WHENPREV TOK_WHENFIRST TOK_WHENLAST
// Action tokens
%token        TOK_ASSUME TOK_ASSERT TOK_REQUIRE TOK_ENSURE
%token        TOK_ASSIGN
%token        TOK_VAR TOK_LOCAL TOK_LET TOK_CALL
%token        TOK_WHILE TOK_FOR TOK_IN TOK_INVARIANT TOK_DECREASES
%token        TOK_RETURNS
%token        TOK_SOME TOK_MINIMIZING TOK_MAXIMIZING
%token        TOK_DEBUG TOK_THUNK TOK_UNPROVABLE TOK_PROOF
%token        TOK_INSTANTIATE
%token <str>  TOK_LABEL
%token        TOK_CARET TOK_METHOD TOK_NULL TOK_SET
%token        TOK_WITH
// Scenario tokens
%token        TOK_SCENARIO TOK_BEFORE TOK_AFTER
// Proof/tactic tokens
%token        TOK_TACTIC TOK_DEFINITION TOK_TRIGGER
%token        TOK_SHOWGOALS TOK_DEFERGOAL TOK_SPOIL
%token        TOK_UNFOLD TOK_FORGET
%token        TOK_PROPERTY TOK_FUNCTION TOK_THEOREM
%token        TOK_APPLY

// Nonterminal types — formula/term
%type <node>  top term fmla appelem var simplevar atype
%type <nodes> terms vars simplevars
%type <str>   SYMBOLx SYMsubscr
// Nonterminal types — actions
%type <node>  action simpleact complexact sequence labeledfmla
%type <node>  tterm
%type <nodes> tterms actseq
%type <nodes> lparams
%type <node>  lparam
// Nonterminal types — scenario
%type <node>  scenario sceninit scenariomixin scentrans
%type <nodes> scentranss places
// Nonterminal types — proof/tactic
%type <node>  proofstep proofseq proofgroup optproofgroup opttacticwith
%type <node>  tacticwithlistchoice tacticwithelem
%type <nodes> tacticwithlist pflets
%type <node>  pflet

// Precedence declarations — copied exactly from Python v1.7+ precedence table.
%left         TOK_SEMI
%left         TOK_GLOBALLY TOK_EVENTUALLY
%left         TOK_ARROW TOK_IFF
%left         TOK_OR
%left         TOK_AND
%left         TOK_TILDA
%left         TOK_EQ TOK_LE TOK_LT TOK_GE TOK_GT TOK_PTO TOK_ISA
%left         TOK_TILDAEQ
%left         TOK_IF
%left         TOK_ELSE
%left         TOK_COLON
%left         TOK_PLUS TOK_MINUS
%left         TOK_TIMES TOK_DIV
%left         TOK_DOLLAR
%left         TOK_OLD
%left         TOK_DOT
%right        TOK_ASSIGN

%start        top

%%

top:
    fmla
    {
        v17lex.(*v17LexAdapter).result = $1
    }
    | sequence
    {
        v17lex.(*v17LexAdapter).result = $1
    }
    | scenario
    {
        v17lex.(*v17LexAdapter).result = $1
    }
    // Proof/tactic entry: tactic SYMBOL opttacticwith optproofgroup
    | TOK_TACTIC atype opttacticwith optproofgroup
    {
        v17lex.(*v17LexAdapter).result = acfg(v17lex).NewTacticTactic($2, $3, $4)
    }
    // Proof entry: proof [label] { proofseq }
    | TOK_PROOF TOK_LABEL proofgroup
    {
        v17lex.(*v17LexAdapter).result = acfg(v17lex).NewProofTactic(acfg(v17lex).NewAtom($2), $3)
    }
    ;

// --- SYMBOL handling ---

SYMBOLx:
    TOK_PRESYMBOL
    {
        $$ = $1
    }
    ;

SYMsubscr:
    SYMBOLx
    {
        $$ = $1
    }
    | TOK_THIS
    {
        $$ = "this"
    }
    | SYMsubscr TOK_DOT SYMBOLx
    {
        $$ = $1 + "." + $3
    }
    ;

// --- atype (sort names) ---

atype:
    SYMBOLx
    {
        $$ = acfg(v17lex).NewSymbol($1, nil)
    }
    | atype TOK_DOT SYMBOLx
    {
        if _, ok := $1.(*ast.This); ok {
            $$ = acfg(v17lex).NewSymbol($3, nil)
        } else if sym, ok := $1.(*ast.Symbol); ok {
            $$ = acfg(v17lex).NewSymbol(sym.Rep + "." + $3, nil)
        } else {
            $$ = acfg(v17lex).NewSymbol($3, nil)
        }
    }
    | TOK_THIS
    {
        $$ = acfg(v17lex).NewThis()
    }
    ;

// --- appelem (v1.7+: application elements) ---
// Bare symbol or symbol(terms...) — produces *ast.Atom for the LALR parser
// (matching what the hand-written parser produces).

appelem:
    SYMBOLx
    {
        // Python: App(p[1]) — appelem produces App, not Atom.
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1, nil))
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        // Python: App(p[1], p[3])
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol($1, nil), $3...)
    }
    ;

// --- Variables ---

var:
    TOK_VARIABLE
    {
        $$ = acfg(v17lex).NewVariable($1, "")
    }
    | TOK_VARIABLE TOK_COLON atype
    {
        v := acfg(v17lex).NewVariable($1, atypeToString($3))
        $$ = v
    }
    ;

simplevar:
    TOK_VARIABLE
    {
        $$ = acfg(v17lex).NewVariable($1, "S")
    }
    | TOK_VARIABLE TOK_COLON SYMBOLx
    {
        $$ = acfg(v17lex).NewVariable($1, $3)
    }
    ;

vars:
    var
    {
        $$ = []ast.Node{$1}
    }
    | vars TOK_COMMA var
    {
        $$ = append($1, $3)
    }
    ;

simplevars:
    simplevar
    {
        $$ = []ast.Node{$1}
    }
    | simplevars TOK_COMMA simplevar
    {
        $$ = append($1, $3)
    }
    ;

// --- Terms list ---

terms:
    /* empty */
    {
        $$ = nil
    }
    | term
    {
        $$ = []ast.Node{$1}
    }
    | terms TOK_COMMA term
    {
        $$ = append($1, $3)
    }
    ;

// --- Terms (v1.7+: unified with formulas) ---

term:
    appelem
    {
        $$ = $1
    }
    | var
    {
        $$ = $1
    }
    | TOK_OLD appelem
    {
        $$ = acfg(v17lex).NewOld($2)
    }
    | term TOK_DOT appelem
    {
        switch lhs := $1.(type) {
        case *ast.Atom:
            rhs := $3.(*ast.Atom)
            newRep := lhs.Rep + "." + rhs.Rep
            newTerms := make([]ast.Node, 0, len(lhs.Terms)+len(rhs.Terms))
            newTerms = append(newTerms, lhs.Terms...)
            newTerms = append(newTerms, rhs.Terms...)
            $$ = acfg(v17lex).NewAtom(newRep, newTerms...)
        case *ast.Old:
            if inner, ok := lhs.Term.(*ast.Atom); ok {
                rhs := $3.(*ast.Atom)
                newRep := inner.Rep + "." + rhs.Rep
                newTerms := make([]ast.Node, 0, len(inner.Terms)+len(rhs.Terms))
                newTerms = append(newTerms, inner.Terms...)
                newTerms = append(newTerms, rhs.Terms...)
                lhs.Term = acfg(v17lex).NewAtom(newRep, newTerms...)
                $$ = lhs
            } else {
                $$ = acfg(v17lex).NewMethodCall($1, $3)
            }
        default:
            $$ = acfg(v17lex).NewMethodCall($1, $3)
        }
    }
    | TOK_LPAREN term TOK_RPAREN
    {
        $$ = $2
    }
    // --- Arithmetic (App, not Atom — matches Python's App for term-level ops) ---
    | term TOK_PLUS term
    {
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("+", nil), $1, $3)
    }
    | term TOK_MINUS term
    {
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("-", nil), $1, $3)
    }
    | term TOK_TIMES term
    {
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("*", nil), $1, $3)
    }
    | term TOK_DIV term
    {
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("/", nil), $1, $3)
    }
    // --- If/else ---
    | term TOK_IF fmla TOK_ELSE term
    {
        $$ = acfg(v17lex).NewIte($3, $1, $5)
    }
    // --- Comparison (v1.7+: term-level) ---
    | term TOK_EQ term
    {
        $$ = acfg(v17lex).NewAtom("=", $1, $3)
    }
    | term TOK_LE term
    {
        $$ = acfg(v17lex).NewAtom("<=", $1, $3)
    }
    | term TOK_LT term
    {
        $$ = acfg(v17lex).NewAtom("<", $1, $3)
    }
    | term TOK_GE term
    {
        $$ = acfg(v17lex).NewAtom(">=", $1, $3)
    }
    | term TOK_GT term
    {
        $$ = acfg(v17lex).NewAtom(">", $1, $3)
    }
    | term TOK_PTO term
    {
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("*>", nil), $1, $3)
    }
    | term TOK_TILDAEQ term
    {
        $$ = acfg(v17lex).NewNot(acfg(v17lex).NewAtom("=", $1, $3))
    }
    // --- Boolean ---
    | TOK_TRUE
    {
        $$ = acfg(v17lex).NewAnd()
    }
    | TOK_FALSE
    {
        $$ = acfg(v17lex).NewOr()
    }
    | TOK_TILDA term
    {
        $$ = acfg(v17lex).NewNot($2)
    }
    | term TOK_AND term
    {
        // Python: if isinstance(p[1], And): p[0] = p[1]; p[0].args.append(p[3])
        //         else: p[0] = And(p[1], p[3])
        // This flattens left-associative chains and absorbs true (And{}) identity.
        if a, ok := $1.(*ast.And); ok {
            a.Terms = append(a.Terms, $3)
            $$ = a
        } else {
            $$ = acfg(v17lex).NewAnd($1, $3)
        }
    }
    | term TOK_OR term
    {
        // Python: if isinstance(p[1], Or): p[0] = p[1]; p[0].args.append(p[3])
        //         else: p[0] = Or(p[1], p[3])
        if o, ok := $1.(*ast.Or); ok {
            o.Terms = append(o.Terms, $3)
            $$ = o
        } else {
            $$ = acfg(v17lex).NewOr($1, $3)
        }
    }
    | term TOK_ARROW term
    {
        $$ = acfg(v17lex).NewImplies($1, $3)
    }
    | term TOK_IFF term
    {
        $$ = acfg(v17lex).NewIff($1, $3)
    }
    // --- Quantifiers ---
    | TOK_FORALL simplevars TOK_DOT term    %prec TOK_SEMI
    {
        $$ = acfg(v17lex).NewForall($2, $4)
    }
    | TOK_EXISTS simplevars TOK_DOT term    %prec TOK_SEMI
    {
        $$ = acfg(v17lex).NewExists($2, $4)
    }
    | TOK_FORALL TOK_LPAREN vars TOK_RPAREN term
    {
        $$ = acfg(v17lex).NewForall($3, $5)
    }
    | TOK_EXISTS TOK_LPAREN vars TOK_RPAREN term
    {
        $$ = acfg(v17lex).NewExists($3, $5)
    }
    // --- Temporal ---
    | TOK_GLOBALLY term
    {
        $$ = acfg(v17lex).NewGlobally($2)
    }
    | TOK_EVENTUALLY term
    {
        $$ = acfg(v17lex).NewEventually($2)
    }
    | term TOK_WHENNEXT term
    {
        $$ = acfg(v17lex).NewWhenOperator("next", $1, $3)
    }
    | term TOK_WHENPREV term
    {
        $$ = acfg(v17lex).NewWhenOperator("prev", $1, $3)
    }
    | term TOK_WHENFIRST term
    {
        $$ = acfg(v17lex).NewWhenOperator("first", $1, $3)
    }
    | term TOK_WHENLAST term
    {
        $$ = acfg(v17lex).NewWhenOperator("last", $1, $3)
    }
    // --- ISA ---
    | term TOK_ISA atype
    {
        $$ = acfg(v17lex).NewIsa($1, $3)
    }
    // --- Sort annotation ---
    | term TOK_COLON atype
    {
        if v, ok := $1.(*ast.Variable); ok {
            v.VSort = atypeToString($3)
        }
        $$ = $1
    }
    // --- Named binders ---
    | TOK_LPAREN TOK_DOLLAR SYMBOLx simplevars TOK_DOT fmla TOK_RPAREN TOK_LPAREN terms TOK_RPAREN
    {
        binder := acfg(v17lex).NewNamedBinder($3, $4, $6)
        $$ = acfg(v17lex).NewAtom("", append([]ast.Node{binder}, $9...)...)
    }
    | TOK_DOLLAR SYMBOLx TOK_DOT fmla     %prec TOK_SEMI
    {
        $$ = acfg(v17lex).NewNamedBinder($2, nil, $4)
    }
    | TOK_DOLLAR SYMBOLx TOK_DOLLAR fmla   %prec TOK_SEMI
    {
        $$ = acfg(v17lex).NewNamedBinder($2, nil, $4)
    }
    ;

// --- fmla: in v1.7+ this converts term to atom form ---

fmla:
    term
    {
        // Python: app_to_atom(p[1]) — convert top-level App to Atom in formula position.
        $$ = ast.AppToAtom($1)
    }
    ;

// --- labeledfmla: formula with optional label ---
// Matches Python's labeledfmla: fmla | LABEL fmla
// The hand-written parser wraps in LabeledFormula.

labeledfmla:
    fmla
    {
        $$ = acfg(v17lex).NewLabeledFormula(nil, $1)
    }
    ;

// ============================================================
// --- Action grammar (from Python ivy_parser.py) ---
// ============================================================

// --- tterm: typed term (symbol with optional sort annotation) ---

tterm:
    SYMBOLx
    {
        a := acfg(v17lex).NewAtom($1)
        $$ = a
    }
    | SYMBOLx TOK_COLON atype
    {
        a := acfg(v17lex).NewAtom($1)
        a.ASort = $3
        $$ = a
    }
    | TOK_CARET SYMBOLx TOK_COLON atype
    {
        // Ghost parameter: ^name : type
        a := acfg(v17lex).NewAtom($2)
        a.ASort = $4
        $$ = a
    }
    ;

tterms:
    tterm
    {
        $$ = []ast.Node{$1}
    }
    | tterms TOK_COMMA tterm
    {
        $$ = append($1, $3)
    }
    ;

// --- lparam / lparams: local params for local/let/optargs ---

lparam:
    SYMBOLx TOK_COLON atype
    {
        a := acfg(v17lex).NewAtom($1)
        a.ASort = $3
        $$ = a
    }
    | TOK_CARET SYMBOLx TOK_COLON atype
    {
        a := acfg(v17lex).NewAtom($2)
        a.ASort = $4
        $$ = a
    }
    ;

lparams:
    lparam
    {
        $$ = []ast.Node{$1}
    }
    | lparams TOK_COMMA lparam
    {
        $$ = append($1, $3)
    }
    ;

// --- sequence: { action; action; ... } ---

sequence:
    TOK_LCB TOK_RCB
    {
        $$ = acfg(v17lex).NewAnd()
    }
    | TOK_LCB actseq TOK_RCB
    {
        $$ = lalrMakeSequence(acfg(v17lex), $2)
    }
    | TOK_LCB actseq TOK_SEMI TOK_RCB
    {
        $$ = lalrMakeSequence(acfg(v17lex), $2)
    }
    ;

actseq:
    action
    {
        $$ = []ast.Node{$1}
    }
    | actseq TOK_SEMI action
    {
        $$ = append($1, $3)
    }
    | actseq complexact
    {
        // complexact after complexact (no semicolon needed)
        $$ = append($1, $2)
    }
    ;

action:
    simpleact
    {
        $$ = $1
    }
    | complexact
    {
        $$ = $1
    }
    ;

// --- Simple actions (from Python ivy_parser.py:2374-2478) ---

simpleact:
    TOK_ASSUME labeledfmla
    {
        $$ = acfg(v17lex).NewAtom("assume", $2)
    }
    | TOK_ASSERT labeledfmla
    {
        $$ = acfg(v17lex).NewAtom("assert", $2)
    }
    | TOK_REQUIRE labeledfmla
    {
        $$ = acfg(v17lex).NewAtom("require", $2)
    }
    | TOK_ENSURE labeledfmla
    {
        $$ = acfg(v17lex).NewAtom("ensure", $2)
    }
    | term TOK_ASSIGN fmla
    {
        $$ = acfg(v17lex).NewAtom(":=", $1, $3)
    }
    | term TOK_ASSIGN TOK_TIMES
    {
        $$ = acfg(v17lex).NewAtom("havoc", $1)
    }
    | TOK_VAR tterm
    {
        $$ = acfg(v17lex).NewVarAction($2)
    }
    | TOK_VAR tterm TOK_ASSIGN fmla
    {
        $$ = acfg(v17lex).NewVarAction($2, $4)
    }
    | TOK_CALL term
    {
        // Simple call: call f(x)
        $$ = $2
    }
    | TOK_INSTANTIATE term
    {
        $$ = acfg(v17lex).NewAtom("instantiate", $2)
    }
    | TOK_UNPROVABLE simpleact
    {
        // When check_unprovable is False (default), unprovable statements are no-ops
        $$ = acfg(v17lex).NewAnd()
    }
    | term     %prec TOK_SEMI
    {
        // Bare expression (procedure call)
        $$ = $1
    }
    ;

// --- Complex actions (from Python ivy_parser.py:2553-2827) ---

complexact:
    sequence
    {
        $$ = $1
    }
    | TOK_IF fmla sequence
    {
        $$ = acfg(v17lex).NewIte($2, $3, acfg(v17lex).NewAnd())
    }
    | TOK_IF fmla sequence TOK_ELSE action
    {
        $$ = acfg(v17lex).NewIte($2, $3, $5)
    }
    | TOK_IF TOK_TIMES sequence TOK_ELSE action
    {
        // ChoiceAction: if * { ... } else { ... }
        $$ = acfg(v17lex).NewIte(acfg(v17lex).NewSymbol("*", nil), $3, $5)
    }
    | TOK_WHILE fmla sequence
    {
        $$ = acfg(v17lex).NewAtom("while", $2, $3)
    }
    | TOK_WHILE fmla TOK_INVARIANT fmla sequence
    {
        $$ = acfg(v17lex).NewAtom("while", $2, $5)
    }
    | TOK_FOR tterm TOK_COMMA tterm TOK_IN fmla sequence
    {
        $$ = acfg(v17lex).NewAtom("for", $2, $4, $6, $7)
    }
    | TOK_LOCAL lparams sequence
    {
        args := append($2, $3)
        $$ = acfg(v17lex).NewAtom("local", args...)
    }
    | TOK_LET fmla sequence
    {
        $$ = acfg(v17lex).NewAtom("let", $2, $3)
    }
    ;

// ============================================================
// --- Scenario grammar (from Python ivy_parser.py:2202-2269) ---
// ============================================================

scenario:
    TOK_SCENARIO TOK_LCB sceninit TOK_SEMI scentranss TOK_RCB
    {
        elems := append([]ast.Node{$3}, $5...)
        sdef := acfg(v17lex).NewScenarioDef(elems)
        $$ = acfg(v17lex).NewScenarioDecl(sdef)
    }
    ;

sceninit:
    TOK_ARROW places
    {
        $$ = acfg(v17lex).NewPlaceList($2)
    }
    ;

places:
    TOK_PRESYMBOL
    {
        $$ = []ast.Node{acfg(v17lex).NewAtom($1)}
    }
    | places TOK_COMMA TOK_PRESYMBOL
    {
        $$ = append($1, acfg(v17lex).NewAtom($3))
    }
    ;

scentranss:
    /* empty */
    {
        $$ = nil
    }
    | scentranss scentrans
    {
        $$ = append($1, $2)
    }
    ;

scentrans:
    places TOK_ARROW places TOK_COLON scenariomixin
    {
        $$ = acfg(v17lex).NewScenarioTransition(acfg(v17lex).NewPlaceList($1), acfg(v17lex).NewPlaceList($3), $5)
    }
    | places TOK_COLON scenariomixin
    {
        $$ = acfg(v17lex).NewScenarioTransition(acfg(v17lex).NewPlaceList($1), acfg(v17lex).NewPlaceList(nil), $3)
    }
    ;

scenariomixin:
    TOK_BEFORE atype sequence
    {
        atom := acfg(v17lex).NewAtom($2.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixerName := fmt.Sprintf("%s[before%d]", atom.Rep, lalrLabelCounter)
        mixer := acfg(v17lex).NewAtom(mixerName)
        adef := acfg(v17lex).NewActionDef(atom, $3, nil, nil)
        $$ = acfg(v17lex).NewScenarioBeforeMixin(mixer, adef)
    }
    | TOK_AFTER atype sequence
    {
        atom := acfg(v17lex).NewAtom($2.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixerName := fmt.Sprintf("%s[after%d]", atom.Rep, lalrLabelCounter)
        mixer := acfg(v17lex).NewAtom(mixerName)
        adef := acfg(v17lex).NewActionDef(atom, $3, nil, nil)
        $$ = acfg(v17lex).NewScenarioAfterMixin(mixer, adef)
    }
    ;

// ========================================================================
// Proof / tactic productions — ported from Python ivy_parser.py.
// These define the grammar for proof scripts that invoke tactics like
// tempind, skolemizenp, vcgen, sorry, etc.
// ========================================================================

// pflet : var EQ fmla
pflet:
    var TOK_EQ fmla
    {
        $$ = acfg(v17lex).NewDefinition($1, $3)
    }
    ;

// pflets : pflet | pflets COMMA pflet
pflets:
    pflet
    {
        $$ = []ast.Node{$1}
    }
    | pflets TOK_COMMA pflet
    {
        $$ = append($1, $3)
    }
    ;

// tacticwithelem : INVARIANT labeledfmla | DEFINITION atype EQ fmla | TRIGGER atype WITH terms
tacticwithelem:
    TOK_INVARIANT labeledfmla
    {
        $$ = $2
    }
    | TOK_DEFINITION atype TOK_EQ fmla
    {
        $$ = acfg(v17lex).NewDefinition($2, $4)
    }
    | TOK_TRIGGER atype TOK_WITH terms
    {
        $$ = acfg(v17lex).NewTrigger(nil, append([]ast.Node{$2}, $4...)...)
    }
    ;

// tacticwithlist : tacticwithelem | tacticwithlist tacticwithelem
tacticwithlist:
    tacticwithelem
    {
        $$ = []ast.Node{$1}
    }
    | tacticwithlist tacticwithelem
    {
        $$ = append($1, $2)
    }
    ;

// tacticwithlistchoice : tacticwithlist | pflets
tacticwithlistchoice:
    tacticwithlist
    {
        $$ = acfg(v17lex).NewTacticWith($1)
    }
    | pflets
    {
        $$ = acfg(v17lex).NewTacticLets($1)
    }
    ;

// opttacticwith : (empty) | WITH tacticwithlistchoice | WITH LCB tacticwithlist RCB
opttacticwith:
    /* empty */
    {
        $$ = acfg(v17lex).NewTacticWith(nil)
    }
    | TOK_WITH tacticwithlistchoice
    {
        $$ = $2
    }
    | TOK_WITH TOK_LCB tacticwithlist TOK_RCB
    {
        $$ = acfg(v17lex).NewTacticWith($3)
    }
    ;

// proofgroup : LCB proofseq RCB | LCB RCB
proofgroup:
    TOK_LCB proofseq TOK_RCB
    {
        $$ = $2
    }
    | TOK_LCB TOK_RCB
    {
        $$ = acfg(v17lex).NewNullTactic()
    }
    ;

// optproofgroup : (empty) | proofgroup
optproofgroup:
    /* empty */
    {
        $$ = acfg(v17lex).NewNoneAST()
    }
    | proofgroup
    {
        $$ = $1
    }
    ;

// proofseq : proofstep | proofseq optsemi proofstep
proofseq:
    proofstep
    {
        $$ = $1
    }
    | proofseq TOK_SEMI proofstep
    {
        $$ = acfg(v17lex).NewComposeTactics([]ast.Node{$1, $3})
    }
    | proofseq proofstep
    {
        $$ = acfg(v17lex).NewComposeTactics([]ast.Node{$1, $2})
    }
    ;

// proofstep — all the various proof step forms
proofstep:
    // proofstep : APPLY atype
    TOK_APPLY atype
    {
        $$ = acfg(v17lex).NewSchemaInstantiation($2, acfg(v17lex).NewNoneAST())
    }
    // proofstep : ASSUME atype
    | TOK_ASSUME atype
    {
        $$ = acfg(v17lex).NewAssumeTactic($2, acfg(v17lex).NewNoneAST())
    }
    // proofstep : SHOWGOALS
    | TOK_SHOWGOALS
    {
        $$ = acfg(v17lex).NewShowGoalsTactic()
    }
    // proofstep : DEFERGOAL
    | TOK_DEFERGOAL
    {
        $$ = acfg(v17lex).NewDeferGoalTactic()
    }
    // proofstep : SPOIL atype
    | TOK_SPOIL atype
    {
        $$ = acfg(v17lex).NewSpoilTactic($2)
    }
    // proofstep : TACTIC SYMBOL opttacticwith optproofgroup
    | TOK_TACTIC atype opttacticwith optproofgroup
    {
        $$ = acfg(v17lex).NewTacticTactic($2, $3, $4)
    }
    // proofstep : PROPERTY labeledfmla optproofgroup
    | TOK_PROPERTY labeledfmla optproofgroup
    {
        $$ = acfg(v17lex).NewPropertyTactic($2, acfg(v17lex).NewNoneAST(), $3)
    }
    // proofstep : FUNCTION atype
    | TOK_FUNCTION atype
    {
        $$ = acfg(v17lex).NewFunctionTactic([]ast.Node{$2})
    }
    // proofstep : PROOF LABEL proofgroup
    | TOK_PROOF TOK_LABEL proofgroup
    {
        $$ = acfg(v17lex).NewProofTactic(acfg(v17lex).NewAtom($2), $3)
    }
    // proofstep : LET pflets
    | TOK_LET pflets
    {
        $$ = acfg(v17lex).NewLetTactic($2)
    }
    // proofstep : INSTANTIATE WITH pflets (witness tactic)
    | TOK_INSTANTIATE TOK_WITH pflets
    {
        $$ = acfg(v17lex).NewWitnessTactic($3)
    }
    // proofstep : IF fmla proofgroup ELSE proofgroup
    | TOK_IF fmla proofgroup TOK_ELSE proofgroup
    {
        $$ = acfg(v17lex).NewIfTactic($2, $3, $5)
    }
    // proofstep : UNFOLD WITH atype (simplified)
    | TOK_UNFOLD TOK_WITH atype
    {
        $$ = acfg(v17lex).NewUnfoldTactic(acfg(v17lex).NewNoneAST(), []ast.Node{$3})
    }
    // proofstep : FORGET atype
    | TOK_FORGET atype
    {
        $$ = acfg(v17lex).NewForgetTactic([]ast.Node{$2})
    }
    // proofstep : proofgroup (nested braces)
    | proofgroup
    {
        $$ = $1
    }
    ;

%%

// lalrMakeSequence wraps a list of action nodes into a single And node (sequence).
func lalrMakeSequence(cfg *ast.AstConfig, stmts []ast.Node) ast.Node {
	// Lower var declarations into nested local scopes, matching Python/HW parser.
	stmts = ast.LowerVarStatements(stmts)
	if len(stmts) == 0 {
		return cfg.NewAnd()
	}
	if len(stmts) == 1 {
		return stmts[0]
	}
	return cfg.NewAnd(stmts...)
}
