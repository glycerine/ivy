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
        v17lex.(*v17LexAdapter).result = &ast.TacticTactic{TName: $2, Body: $3, Proof: $4}
    }
    // Proof entry: proof [label] { proofseq }
    | TOK_PROOF TOK_LABEL proofgroup
    {
        v17lex.(*v17LexAdapter).result = &ast.ProofTactic{TLabel: acfg(v17lex).NewAtom($2), Proof: $3}
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
        $$ = &ast.Symbol{Rep: $1}
    }
    | atype TOK_DOT SYMBOLx
    {
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
        $$ = &ast.This{}
    }
    ;

// --- appelem (v1.7+: application elements) ---
// Bare symbol or symbol(terms...) — produces *ast.Atom for the LALR parser
// (matching what the hand-written parser produces).

appelem:
    SYMBOLx
    {
        $$ = &ast.Atom{Rep: $1}
    }
    | SYMBOLx TOK_LPAREN terms TOK_RPAREN
    {
        $$ = &ast.Atom{Rep: $1, Terms: $3}
    }
    ;

// --- Variables ---

var:
    TOK_VARIABLE
    {
        $$ = &ast.Variable{Rep: $1}
    }
    | TOK_VARIABLE TOK_COLON atype
    {
        v := &ast.Variable{Rep: $1, VSort: atypeToString($3)}
        $$ = v
    }
    ;

simplevar:
    TOK_VARIABLE
    {
        $$ = &ast.Variable{Rep: $1, VSort: "S"}
    }
    | TOK_VARIABLE TOK_COLON SYMBOLx
    {
        $$ = &ast.Variable{Rep: $1, VSort: $3}
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
        $$ = &ast.Old{Term: $2}
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
            $$ = &ast.Atom{Rep: newRep, Terms: newTerms}
        case *ast.Old:
            if inner, ok := lhs.Term.(*ast.Atom); ok {
                rhs := $3.(*ast.Atom)
                newRep := inner.Rep + "." + rhs.Rep
                newTerms := make([]ast.Node, 0, len(inner.Terms)+len(rhs.Terms))
                newTerms = append(newTerms, inner.Terms...)
                newTerms = append(newTerms, rhs.Terms...)
                lhs.Term = &ast.Atom{Rep: newRep, Terms: newTerms}
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
        $$ = &ast.Ite{Cond: $3, Then: $1, Else: $5}
    }
    // --- Comparison (v1.7+: term-level) ---
    | term TOK_EQ term
    {
        $$ = &ast.Atom{Rep: "=", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_LE term
    {
        $$ = &ast.Atom{Rep: "<=", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_LT term
    {
        $$ = &ast.Atom{Rep: "<", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_GE term
    {
        $$ = &ast.Atom{Rep: ">=", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_GT term
    {
        $$ = &ast.Atom{Rep: ">", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_PTO term
    {
        $$ = acfg(v17lex).NewApp(acfg(v17lex).NewSymbol("*>", nil), $1, $3)
    }
    | term TOK_TILDAEQ term
    {
        $$ = &ast.Not{Body: &ast.Atom{Rep: "=", Terms: []ast.Node{$1, $3}}}
    }
    // --- Boolean ---
    | TOK_TRUE
    {
        $$ = &ast.And{}
    }
    | TOK_FALSE
    {
        $$ = &ast.Or{}
    }
    | TOK_TILDA term
    {
        $$ = &ast.Not{Body: $2}
    }
    | term TOK_AND term
    {
        $$ = &ast.And{Terms: []ast.Node{$1, $3}}
    }
    | term TOK_OR term
    {
        $$ = &ast.Or{Terms: []ast.Node{$1, $3}}
    }
    | term TOK_ARROW term
    {
        $$ = &ast.Implies{T1: $1, T2: $3}
    }
    | term TOK_IFF term
    {
        $$ = &ast.Iff{T1: $1, T2: $3}
    }
    // --- Quantifiers ---
    | TOK_FORALL simplevars TOK_DOT term    %prec TOK_SEMI
    {
        $$ = &ast.Forall{Bounds: $2, Body: $4}
    }
    | TOK_EXISTS simplevars TOK_DOT term    %prec TOK_SEMI
    {
        $$ = &ast.Exists{Bounds: $2, Body: $4}
    }
    | TOK_FORALL TOK_LPAREN vars TOK_RPAREN term
    {
        $$ = &ast.Forall{Bounds: $3, Body: $5}
    }
    | TOK_EXISTS TOK_LPAREN vars TOK_RPAREN term
    {
        $$ = &ast.Exists{Bounds: $3, Body: $5}
    }
    // --- Temporal ---
    | TOK_GLOBALLY term
    {
        $$ = &ast.Globally{Body: $2}
    }
    | TOK_EVENTUALLY term
    {
        $$ = &ast.Eventually{Body: $2}
    }
    | term TOK_WHENNEXT term
    {
        $$ = &ast.WhenOperator{Name: "next", T1: $1, T2: $3}
    }
    | term TOK_WHENPREV term
    {
        $$ = &ast.WhenOperator{Name: "prev", T1: $1, T2: $3}
    }
    | term TOK_WHENFIRST term
    {
        $$ = &ast.WhenOperator{Name: "first", T1: $1, T2: $3}
    }
    | term TOK_WHENLAST term
    {
        $$ = &ast.WhenOperator{Name: "last", T1: $1, T2: $3}
    }
    // --- ISA ---
    | term TOK_ISA atype
    {
        $$ = &ast.Isa{Terms: []ast.Node{$1, $3}}
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
        binder := &ast.NamedBinder{Name: $3, Bounds: $4, Body: $6}
        $$ = &ast.Atom{Rep: "", Terms: append([]ast.Node{binder}, $9...)}
    }
    | TOK_DOLLAR SYMBOLx TOK_DOT fmla     %prec TOK_SEMI
    {
        $$ = &ast.NamedBinder{Name: $2, Body: $4}
    }
    | TOK_DOLLAR SYMBOLx TOK_DOLLAR fmla   %prec TOK_SEMI
    {
        $$ = &ast.NamedBinder{Name: $2, Body: $4}
    }
    ;

// --- fmla: in v1.7+ this converts term to atom form ---

fmla:
    term
    {
        // Convert App to Atom if needed (matches Python's app_to_atom)
        $$ = $1
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
        a := &ast.Atom{Rep: $1}
        $$ = a
    }
    | SYMBOLx TOK_COLON atype
    {
        a := &ast.Atom{Rep: $1}
        a.ASort = $3
        $$ = a
    }
    | TOK_CARET SYMBOLx TOK_COLON atype
    {
        // Ghost parameter: ^name : type
        a := &ast.Atom{Rep: $2}
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
        a := &ast.Atom{Rep: $1}
        a.ASort = $3
        $$ = a
    }
    | TOK_CARET SYMBOLx TOK_COLON atype
    {
        a := &ast.Atom{Rep: $2}
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
        $$ = &ast.And{}
    }
    | TOK_LCB actseq TOK_RCB
    {
        $$ = lalrMakeSequence($2)
    }
    | TOK_LCB actseq TOK_SEMI TOK_RCB
    {
        $$ = lalrMakeSequence($2)
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
        $$ = acfg(v17lex).NewAtom("var", $2)
    }
    | TOK_VAR tterm TOK_ASSIGN fmla
    {
        $$ = acfg(v17lex).NewAtom("var", $2, $4)
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
        $$ = &ast.And{}
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
        $$ = acfg(v17lex).NewIte($2, $3, &ast.And{})
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
        sdef := &ast.ScenarioDef{Elems: elems}
        $$ = acfg(v17lex).NewScenarioDecl(sdef)
    }
    ;

sceninit:
    TOK_ARROW places
    {
        $$ = &ast.PlaceList{Elems: $2}
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
        $$ = &ast.ScenarioTransition{
            From:   &ast.PlaceList{Elems: $1},
            To:     &ast.PlaceList{Elems: $3},
            Action: $5,
        }
    }
    | places TOK_COLON scenariomixin
    {
        $$ = &ast.ScenarioTransition{
            From:   &ast.PlaceList{Elems: $1},
            To:     &ast.PlaceList{},
            Action: $3,
        }
    }
    ;

scenariomixin:
    TOK_BEFORE atype sequence
    {
        atom := acfg(v17lex).NewAtom($2.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixerName := fmt.Sprintf("%s[before%d]", atom.Rep, lalrLabelCounter)
        mixer := acfg(v17lex).NewAtom(mixerName)
        adef := &ast.ActionDef{Name: atom, Body: $3}
        $$ = &ast.ScenarioBeforeMixin{Mixer: mixer, Def: adef}
    }
    | TOK_AFTER atype sequence
    {
        atom := acfg(v17lex).NewAtom($2.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixerName := fmt.Sprintf("%s[after%d]", atom.Rep, lalrLabelCounter)
        mixer := acfg(v17lex).NewAtom(mixerName)
        adef := &ast.ActionDef{Name: atom, Body: $3}
        $$ = &ast.ScenarioAfterMixin{Mixer: mixer, Def: adef}
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
        $$ = &ast.Trigger{Terms: append([]ast.Node{$2}, $4...)}
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
        $$ = &ast.TacticWith{Elems: $1}
    }
    | pflets
    {
        $$ = &ast.TacticLets{Lets: $1}
    }
    ;

// opttacticwith : (empty) | WITH tacticwithlistchoice | WITH LCB tacticwithlist RCB
opttacticwith:
    /* empty */
    {
        $$ = &ast.TacticWith{}
    }
    | TOK_WITH tacticwithlistchoice
    {
        $$ = $2
    }
    | TOK_WITH TOK_LCB tacticwithlist TOK_RCB
    {
        $$ = &ast.TacticWith{Elems: $3}
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
        $$ = &ast.NullTactic{}
    }
    ;

// optproofgroup : (empty) | proofgroup
optproofgroup:
    /* empty */
    {
        $$ = &ast.NoneAST{}
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
        $$ = &ast.ComposeTactics{Tactics: []ast.Node{$1, $3}}
    }
    | proofseq proofstep
    {
        $$ = &ast.ComposeTactics{Tactics: []ast.Node{$1, $2}}
    }
    ;

// proofstep — all the various proof step forms
proofstep:
    // proofstep : APPLY atype
    TOK_APPLY atype
    {
        $$ = &ast.SchemaInstantiation{SchemaName: $2, Ren: &ast.NoneAST{}}
    }
    // proofstep : ASSUME atype
    | TOK_ASSUME atype
    {
        $$ = &ast.AssumeTactic{SchemaName: $2, Ren: &ast.NoneAST{}}
    }
    // proofstep : SHOWGOALS
    | TOK_SHOWGOALS
    {
        $$ = &ast.ShowGoalsTactic{}
    }
    // proofstep : DEFERGOAL
    | TOK_DEFERGOAL
    {
        $$ = &ast.DeferGoalTactic{}
    }
    // proofstep : SPOIL atype
    | TOK_SPOIL atype
    {
        $$ = &ast.SpoilTactic{Target: $2}
    }
    // proofstep : TACTIC SYMBOL opttacticwith optproofgroup
    | TOK_TACTIC atype opttacticwith optproofgroup
    {
        $$ = &ast.TacticTactic{TName: $2, Body: $3, Proof: $4}
    }
    // proofstep : PROPERTY labeledfmla optproofgroup
    | TOK_PROPERTY labeledfmla optproofgroup
    {
        $$ = &ast.PropertyTactic{Prop: $2, PName: &ast.NoneAST{}, Proof: $3}
    }
    // proofstep : FUNCTION atype
    | TOK_FUNCTION atype
    {
        $$ = &ast.FunctionTactic{Elems: []ast.Node{$2}}
    }
    // proofstep : PROOF LABEL proofgroup
    | TOK_PROOF TOK_LABEL proofgroup
    {
        $$ = &ast.ProofTactic{TLabel: acfg(v17lex).NewAtom($2), Proof: $3}
    }
    // proofstep : LET pflets
    | TOK_LET pflets
    {
        $$ = &ast.LetTactic{Defs: $2}
    }
    // proofstep : INSTANTIATE WITH pflets (witness tactic)
    | TOK_INSTANTIATE TOK_WITH pflets
    {
        $$ = &ast.WitnessTactic{Witnesses: $3}
    }
    // proofstep : IF fmla proofgroup ELSE proofgroup
    | TOK_IF fmla proofgroup TOK_ELSE proofgroup
    {
        $$ = &ast.IfTactic{Cond: $2, Then: $3, Else: $5}
    }
    // proofstep : UNFOLD WITH atype (simplified)
    | TOK_UNFOLD TOK_WITH atype
    {
        $$ = &ast.UnfoldTactic{Premise: &ast.NoneAST{}, UnfSpecs: []ast.Node{$3}}
    }
    // proofstep : FORGET atype
    | TOK_FORGET atype
    {
        $$ = &ast.ForgetTactic{Names: []ast.Node{$2}}
    }
    // proofstep : proofgroup (nested braces)
    | proofgroup
    {
        $$ = $1
    }
    ;

%%

// lalrMakeSequence wraps a list of action nodes into a single And node (sequence).
func lalrMakeSequence(stmts []ast.Node) ast.Node {
	if len(stmts) == 0 {
		return &ast.And{}
	}
	if len(stmts) == 1 {
		return stmts[0]
	}
	return &ast.And{Terms: stmts}
}
