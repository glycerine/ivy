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
	"github.com/glycerine/ivy/goivy/ast"
)

// labelCounter is a package-level counter for generating unique mixer names.
var lalrLabelCounter int

// acfg extracts the *ast.AstConfig from the lexer for use in grammar actions.
func lalr17Acfg(lex lalr17Lexer) *ast.AstConfig {
	return lex.(*lalr17LexAdapter).cfg
}

// atypeToString extracts the string sort name from an atype Node.
func lalr17AtypeToString(n ast.Node) string {
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
%token <str>  LALR17_TOK_SYMBOL LALR17_TOK_VARIABLE LALR17_TOK_PRESYMBOL
%token        LALR17_TOK_LPAREN LALR17_TOK_RPAREN LALR17_TOK_LB LALR17_TOK_RB LALR17_TOK_LCB LALR17_TOK_RCB
%token        LALR17_TOK_COMMA LALR17_TOK_SEMI LALR17_TOK_COLON LALR17_TOK_DOT
%token        LALR17_TOK_PLUS LALR17_TOK_MINUS LALR17_TOK_TIMES LALR17_TOK_DIV
%token        LALR17_TOK_EQ LALR17_TOK_TILDAEQ LALR17_TOK_TILDA LALR17_TOK_LE LALR17_TOK_LT LALR17_TOK_GE LALR17_TOK_GT
%token        LALR17_TOK_AND LALR17_TOK_OR LALR17_TOK_ARROW LALR17_TOK_IFF
%token        LALR17_TOK_PTO LALR17_TOK_DOLLAR
%token        LALR17_TOK_FORALL LALR17_TOK_EXISTS
%token        LALR17_TOK_TRUE LALR17_TOK_FALSE
%token        LALR17_TOK_OLD LALR17_TOK_THIS LALR17_TOK_ISA
%token        LALR17_TOK_IF LALR17_TOK_ELSE
%token        LALR17_TOK_GLOBALLY LALR17_TOK_EVENTUALLY
%token        LALR17_TOK_WHENNEXT LALR17_TOK_WHENPREV LALR17_TOK_WHENFIRST LALR17_TOK_WHENLAST
// Action tokens
%token        LALR17_TOK_ASSUME LALR17_TOK_ASSERT LALR17_TOK_REQUIRE LALR17_TOK_ENSURE
%token        LALR17_TOK_ASSIGN
%token        LALR17_TOK_VAR LALR17_TOK_LOCAL LALR17_TOK_LET LALR17_TOK_CALL
%token        LALR17_TOK_WHILE LALR17_TOK_FOR LALR17_TOK_IN LALR17_TOK_INVARIANT LALR17_TOK_DECREASES
%token        LALR17_TOK_RETURNS
%token        LALR17_TOK_SOME LALR17_TOK_MINIMIZING LALR17_TOK_MAXIMIZING
%token        LALR17_TOK_DEBUG LALR17_TOK_THUNK LALR17_TOK_UNPROVABLE LALR17_TOK_PROOF
%token        LALR17_TOK_INSTANTIATE
%token <str>  LALR17_TOK_LABEL
%token        LALR17_TOK_CARET LALR17_TOK_METHOD LALR17_TOK_NULL LALR17_TOK_SET
%token        LALR17_TOK_WITH
// Scenario tokens
%token        LALR17_TOK_SCENARIO LALR17_TOK_BEFORE LALR17_TOK_AFTER
// Proof/tactic tokens
%token        LALR17_TOK_TACTIC LALR17_TOK_DEFINITION LALR17_TOK_TRIGGER
%token        LALR17_TOK_SHOWGOALS LALR17_TOK_DEFERGOAL LALR17_TOK_SPOIL
%token        LALR17_TOK_UNFOLD LALR17_TOK_FORGET
%token        LALR17_TOK_PROPERTY LALR17_TOK_FUNCTION LALR17_TOK_THEOREM
%token        LALR17_TOK_APPLY

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
%left         LALR17_TOK_SEMI
%left         LALR17_TOK_GLOBALLY LALR17_TOK_EVENTUALLY
%left         LALR17_TOK_ARROW LALR17_TOK_IFF
%left         LALR17_TOK_OR
%left         LALR17_TOK_AND
%left         LALR17_TOK_TILDA
%left         LALR17_TOK_EQ LALR17_TOK_LE LALR17_TOK_LT LALR17_TOK_GE LALR17_TOK_GT LALR17_TOK_PTO LALR17_TOK_ISA
%left         LALR17_TOK_TILDAEQ
%left         LALR17_TOK_IF
%left         LALR17_TOK_ELSE
%left         LALR17_TOK_COLON
%left         LALR17_TOK_PLUS LALR17_TOK_MINUS
%left         LALR17_TOK_TIMES LALR17_TOK_DIV
%left         LALR17_TOK_DOLLAR
%left         LALR17_TOK_OLD
%left         LALR17_TOK_DOT
%right        LALR17_TOK_ASSIGN

%start        top

%%

top:
    fmla
    {
        lalr17lex.(*lalr17LexAdapter).result = $1
    }
    | sequence
    {
        lalr17lex.(*lalr17LexAdapter).result = $1
    }
    | scenario
    {
        lalr17lex.(*lalr17LexAdapter).result = $1
    }
    // Proof/tactic entry: tactic SYMBOL opttacticwith optproofgroup
    | LALR17_TOK_TACTIC atype opttacticwith optproofgroup
    {
        lalr17lex.(*lalr17LexAdapter).result = lalr17Acfg(lalr17lex).NewTacticTactic($2, $3, $4)
    }
    // Proof entry: proof [label] { proofseq }
    | LALR17_TOK_PROOF LALR17_TOK_LABEL proofgroup
    {
        lalr17lex.(*lalr17LexAdapter).result = lalr17Acfg(lalr17lex).NewProofTactic(lalr17Acfg(lalr17lex).NewAtom($2), $3)
    }
    ;

// --- SYMBOL handling ---

SYMBOLx:
    LALR17_TOK_PRESYMBOL
    {
        $$ = $1
    }
    ;

SYMsubscr:
    SYMBOLx
    {
        $$ = $1
    }
    | LALR17_TOK_THIS
    {
        $$ = "this"
    }
    | SYMsubscr LALR17_TOK_DOT SYMBOLx
    {
        $$ = $1 + "." + $3
    }
    ;

// --- atype (sort names) ---

atype:
    SYMBOLx
    {
        $$ = lalr17Acfg(lalr17lex).NewSymbol($1, nil)
    }
    | atype LALR17_TOK_DOT SYMBOLx
    {
        if _, ok := $1.(*ast.This); ok {
            $$ = lalr17Acfg(lalr17lex).NewSymbol($3, nil)
        } else if sym, ok := $1.(*ast.Symbol); ok {
            $$ = lalr17Acfg(lalr17lex).NewSymbol(sym.Rep + "." + $3, nil)
        } else {
            $$ = lalr17Acfg(lalr17lex).NewSymbol($3, nil)
        }
    }
    | LALR17_TOK_THIS
    {
        $$ = lalr17Acfg(lalr17lex).NewThis()
    }
    ;

// --- appelem (v1.7+: application elements) ---
// Bare symbol or symbol(terms...) — produces *ast.Atom for the LALR parser
// (matching what the hand-written parser produces).

appelem:
    SYMBOLx
    {
        // Python: App(p[1]) — appelem produces App, not Atom.
        $$ = lalr17Acfg(lalr17lex).NewApp(lalr17Acfg(lalr17lex).NewSymbol($1, nil))
    }
    | SYMBOLx LALR17_TOK_LPAREN terms LALR17_TOK_RPAREN
    {
        // Python: App(p[1], p[3])
        $$ = lalr17Acfg(lalr17lex).NewApp(lalr17Acfg(lalr17lex).NewSymbol($1, nil), $3...)
    }
    ;

// --- Variables ---

var:
    LALR17_TOK_VARIABLE
    {
        $$ = lalr17Acfg(lalr17lex).NewVariable($1, "")
    }
    | LALR17_TOK_VARIABLE LALR17_TOK_COLON atype
    {
        v := lalr17Acfg(lalr17lex).NewVariable($1, lalr17AtypeToString($3))
        $$ = v
    }
    ;

simplevar:
    LALR17_TOK_VARIABLE
    {
        $$ = lalr17Acfg(lalr17lex).NewVariable($1, "S")
    }
    | LALR17_TOK_VARIABLE LALR17_TOK_COLON SYMBOLx
    {
        $$ = lalr17Acfg(lalr17lex).NewVariable($1, $3)
    }
    ;

vars:
    var
    {
        $$ = []ast.Node{$1}
    }
    | vars LALR17_TOK_COMMA var
    {
        $$ = append($1, $3)
    }
    ;

simplevars:
    simplevar
    {
        $$ = []ast.Node{$1}
    }
    | simplevars LALR17_TOK_COMMA simplevar
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
    | terms LALR17_TOK_COMMA term
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
    | LALR17_TOK_OLD appelem
    {
        $$ = lalr17Acfg(lalr17lex).NewOld($2)
    }
    | term LALR17_TOK_DOT appelem
    {
        switch lhs := $1.(type) {
        case *ast.Atom:
            rhs := $3.(*ast.Atom)
            newRep := lhs.Rep + "." + rhs.Rep
            newTerms := make([]ast.Node, 0, len(lhs.Terms)+len(rhs.Terms))
            newTerms = append(newTerms, lhs.Terms...)
            newTerms = append(newTerms, rhs.Terms...)
            $$ = lalr17Acfg(lalr17lex).NewAtom(newRep, newTerms...)
        case *ast.AstOld:
            if inner, ok := lhs.Term.(*ast.Atom); ok {
                rhs := $3.(*ast.Atom)
                newRep := inner.Rep + "." + rhs.Rep
                newTerms := make([]ast.Node, 0, len(inner.Terms)+len(rhs.Terms))
                newTerms = append(newTerms, inner.Terms...)
                newTerms = append(newTerms, rhs.Terms...)
                lhs.Term = lalr17Acfg(lalr17lex).NewAtom(newRep, newTerms...)
                $$ = lhs
            } else {
                $$ = lalr17Acfg(lalr17lex).NewMethodCall($1, $3)
            }
        default:
            $$ = lalr17Acfg(lalr17lex).NewMethodCall($1, $3)
        }
    }
    | LALR17_TOK_LPAREN term LALR17_TOK_RPAREN
    {
        $$ = $2
    }
    // --- Arithmetic (App, not Atom — matches Python's App for term-level ops) ---
    | term LALR17_TOK_PLUS term
    {
        $$ = lalr17Acfg(lalr17lex).NewApp(lalr17Acfg(lalr17lex).NewSymbol("+", nil), $1, $3)
    }
    | term LALR17_TOK_MINUS term
    {
        $$ = lalr17Acfg(lalr17lex).NewApp(lalr17Acfg(lalr17lex).NewSymbol("-", nil), $1, $3)
    }
    | term LALR17_TOK_TIMES term
    {
        $$ = lalr17Acfg(lalr17lex).NewApp(lalr17Acfg(lalr17lex).NewSymbol("*", nil), $1, $3)
    }
    | term LALR17_TOK_DIV term
    {
        $$ = lalr17Acfg(lalr17lex).NewApp(lalr17Acfg(lalr17lex).NewSymbol("/", nil), $1, $3)
    }
    // --- If/else ---
    | term LALR17_TOK_IF fmla LALR17_TOK_ELSE term
    {
        $$ = lalr17Acfg(lalr17lex).NewIte($3, $1, $5)
    }
    // --- Comparison (v1.7+: term-level) ---
    | term LALR17_TOK_EQ term
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("=", $1, $3)
    }
    | term LALR17_TOK_LE term
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("<=", $1, $3)
    }
    | term LALR17_TOK_LT term
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("<", $1, $3)
    }
    | term LALR17_TOK_GE term
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom(">=", $1, $3)
    }
    | term LALR17_TOK_GT term
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom(">", $1, $3)
    }
    | term LALR17_TOK_PTO term
    {
        $$ = lalr17Acfg(lalr17lex).NewApp(lalr17Acfg(lalr17lex).NewSymbol("*>", nil), $1, $3)
    }
    | term LALR17_TOK_TILDAEQ term
    {
        $$ = lalr17Acfg(lalr17lex).NewNot(lalr17Acfg(lalr17lex).NewAtom("=", $1, $3))
    }
    // --- Boolean ---
    | LALR17_TOK_TRUE
    {
        $$ = lalr17Acfg(lalr17lex).NewAnd()
    }
    | LALR17_TOK_FALSE
    {
        $$ = lalr17Acfg(lalr17lex).NewOr()
    }
    | LALR17_TOK_TILDA term
    {
        $$ = lalr17Acfg(lalr17lex).NewNot($2)
    }
    | term LALR17_TOK_AND term
    {
        // Python: if isinstance(p[1], And): p[0] = p[1]; p[0].args.append(p[3])
        //         else: p[0] = And(p[1], p[3])
        // This flattens left-associative chains and absorbs true (And{}) identity.
        if a, ok := $1.(*ast.AstAnd); ok {
            a.Terms = append(a.Terms, $3)
            $$ = a
        } else {
            $$ = lalr17Acfg(lalr17lex).NewAnd($1, $3)
        }
    }
    | term LALR17_TOK_OR term
    {
        // Python: if isinstance(p[1], Or): p[0] = p[1]; p[0].args.append(p[3])
        //         else: p[0] = Or(p[1], p[3])
        if o, ok := $1.(*ast.AstOr); ok {
            o.Terms = append(o.Terms, $3)
            $$ = o
        } else {
            $$ = lalr17Acfg(lalr17lex).NewOr($1, $3)
        }
    }
    | term LALR17_TOK_ARROW term
    {
        $$ = lalr17Acfg(lalr17lex).NewImplies($1, $3)
    }
    | term LALR17_TOK_IFF term
    {
        $$ = lalr17Acfg(lalr17lex).NewIff($1, $3)
    }
    // --- Quantifiers ---
    | LALR17_TOK_FORALL simplevars LALR17_TOK_DOT term    %prec LALR17_TOK_SEMI
    {
        $$ = lalr17Acfg(lalr17lex).NewForall($2, $4)
    }
    | LALR17_TOK_EXISTS simplevars LALR17_TOK_DOT term    %prec LALR17_TOK_SEMI
    {
        $$ = lalr17Acfg(lalr17lex).NewExists($2, $4)
    }
    | LALR17_TOK_FORALL LALR17_TOK_LPAREN vars LALR17_TOK_RPAREN term
    {
        $$ = lalr17Acfg(lalr17lex).NewForall($3, $5)
    }
    | LALR17_TOK_EXISTS LALR17_TOK_LPAREN vars LALR17_TOK_RPAREN term
    {
        $$ = lalr17Acfg(lalr17lex).NewExists($3, $5)
    }
    // --- Temporal ---
    | LALR17_TOK_GLOBALLY term
    {
        $$ = lalr17Acfg(lalr17lex).NewGlobally($2)
    }
    | LALR17_TOK_EVENTUALLY term
    {
        $$ = lalr17Acfg(lalr17lex).NewEventually($2)
    }
    | term LALR17_TOK_WHENNEXT term
    {
        $$ = lalr17Acfg(lalr17lex).NewWhenOperator("next", $1, $3)
    }
    | term LALR17_TOK_WHENPREV term
    {
        $$ = lalr17Acfg(lalr17lex).NewWhenOperator("prev", $1, $3)
    }
    | term LALR17_TOK_WHENFIRST term
    {
        $$ = lalr17Acfg(lalr17lex).NewWhenOperator("first", $1, $3)
    }
    | term LALR17_TOK_WHENLAST term
    {
        $$ = lalr17Acfg(lalr17lex).NewWhenOperator("last", $1, $3)
    }
    // --- ISA ---
    | term LALR17_TOK_ISA atype
    {
        $$ = lalr17Acfg(lalr17lex).NewIsa($1, $3)
    }
    // --- Sort annotation ---
    | term LALR17_TOK_COLON atype
    {
        if v, ok := $1.(*ast.Variable); ok {
            v.VSort = lalr17AtypeToString($3)
        }
        $$ = $1
    }
    // --- Named binders ---
    | LALR17_TOK_LPAREN LALR17_TOK_DOLLAR SYMBOLx simplevars LALR17_TOK_DOT fmla LALR17_TOK_RPAREN LALR17_TOK_LPAREN terms LALR17_TOK_RPAREN
    {
        binder := lalr17Acfg(lalr17lex).NewNamedBinder($3, $4, $6)
        //binder.SetLineno(getLineno(lalr17lex))
        $$ = lalr17Acfg(lalr17lex).NewApp(binder, $9...)
        //$$.(*ast.App).SetLineno(getLineno(lalr17lex))
    }
    | LALR17_TOK_DOLLAR SYMBOLx LALR17_TOK_DOT fmla     %prec LALR17_TOK_SEMI
    {
        $$ = lalr17Acfg(lalr17lex).NewNamedBinder($2, nil, $4)
    }
    | LALR17_TOK_DOLLAR SYMBOLx LALR17_TOK_DOLLAR fmla   %prec LALR17_TOK_SEMI
    {
        $$ = lalr17Acfg(lalr17lex).NewNamedBinder($2, nil, $4)
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
        $$ = lalr17Acfg(lalr17lex).NewLabeledFormula(nil, $1)
    }
    ;

// ============================================================
// --- Action grammar (from Python ivy_parser.py) ---
// ============================================================

// --- tterm: typed term (symbol with optional sort annotation) ---

tterm:
    SYMBOLx
    {
        a := lalr17Acfg(lalr17lex).NewAtom($1)
        $$ = a
    }
    | SYMBOLx LALR17_TOK_COLON atype
    {
        a := lalr17Acfg(lalr17lex).NewAtom($1)
        a.ASort = $3
        $$ = a
    }
    | LALR17_TOK_CARET SYMBOLx LALR17_TOK_COLON atype
    {
        // Ghost parameter: ^name : type
        a := lalr17Acfg(lalr17lex).NewAtom($2)
        a.ASort = $4
        $$ = a
    }
    ;

tterms:
    tterm
    {
        $$ = []ast.Node{$1}
    }
    | tterms LALR17_TOK_COMMA tterm
    {
        $$ = append($1, $3)
    }
    ;

// --- lparam / lparams: local params for local/let/optargs ---

lparam:
    SYMBOLx LALR17_TOK_COLON atype
    {
        a := lalr17Acfg(lalr17lex).NewAtom($1)
        a.ASort = $3
        $$ = a
    }
    | LALR17_TOK_CARET SYMBOLx LALR17_TOK_COLON atype
    {
        a := lalr17Acfg(lalr17lex).NewAtom($2)
        a.ASort = $4
        $$ = a
    }
    ;

lparams:
    lparam
    {
        $$ = []ast.Node{$1}
    }
    | lparams LALR17_TOK_COMMA lparam
    {
        $$ = append($1, $3)
    }
    ;

// --- sequence: { action; action; ... } ---

sequence:
    LALR17_TOK_LCB LALR17_TOK_RCB
    {
        $$ = lalr17Acfg(lalr17lex).NewAnd()
    }
    | LALR17_TOK_LCB actseq LALR17_TOK_RCB
    {
        $$ = lalr17MakeSequence(lalr17Acfg(lalr17lex), $2)
    }
    | LALR17_TOK_LCB actseq LALR17_TOK_SEMI LALR17_TOK_RCB
    {
        $$ = lalr17MakeSequence(lalr17Acfg(lalr17lex), $2)
    }
    ;

actseq:
    action
    {
        $$ = []ast.Node{$1}
    }
    | actseq LALR17_TOK_SEMI action
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
    LALR17_TOK_ASSUME labeledfmla
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("assume", $2)
    }
    | LALR17_TOK_ASSERT labeledfmla
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("assert", $2)
    }
    | LALR17_TOK_REQUIRE labeledfmla
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("require", $2)
    }
    | LALR17_TOK_ENSURE labeledfmla
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("ensure", $2)
    }
    | term LALR17_TOK_ASSIGN fmla
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom(":=", $1, $3)
    }
    | term LALR17_TOK_ASSIGN LALR17_TOK_TIMES
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("havoc", $1)
    }
    | LALR17_TOK_VAR tterm
    {
        $$ = lalr17Acfg(lalr17lex).NewVarAction($2)
    }
    | LALR17_TOK_VAR tterm LALR17_TOK_ASSIGN fmla
    {
        $$ = lalr17Acfg(lalr17lex).NewVarAction($2, $4)
    }
    | LALR17_TOK_CALL term
    {
        // Simple call: call f(x)
        $$ = $2
    }
    | LALR17_TOK_INSTANTIATE term
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("instantiate", $2)
    }
    | LALR17_TOK_UNPROVABLE simpleact
    {
        // When check_unprovable is False (default), unprovable statements are no-ops
        $$ = lalr17Acfg(lalr17lex).NewAnd()
    }
    | term     %prec LALR17_TOK_SEMI
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
    | LALR17_TOK_IF fmla sequence
    {
        $$ = lalr17Acfg(lalr17lex).NewIte($2, $3, lalr17Acfg(lalr17lex).NewAnd())
    }
    | LALR17_TOK_IF fmla sequence LALR17_TOK_ELSE action
    {
        $$ = lalr17Acfg(lalr17lex).NewIte($2, $3, $5)
    }
    | LALR17_TOK_IF LALR17_TOK_TIMES sequence LALR17_TOK_ELSE action
    {
        // ChoiceAction: if * { ... } else { ... }
        $$ = lalr17Acfg(lalr17lex).NewIte(lalr17Acfg(lalr17lex).NewSymbol("*", nil), $3, $5)
    }
    | LALR17_TOK_WHILE fmla sequence
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("while", $2, $3)
    }
    | LALR17_TOK_WHILE fmla LALR17_TOK_INVARIANT fmla sequence
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("while", $2, $5)
    }
    | LALR17_TOK_FOR tterm LALR17_TOK_COMMA tterm LALR17_TOK_IN fmla sequence
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("for", $2, $4, $6, $7)
    }
    | LALR17_TOK_LOCAL lparams sequence
    {
        args := append($2, $3)
        $$ = lalr17Acfg(lalr17lex).NewAtom("local", args...)
    }
    | LALR17_TOK_LET fmla sequence
    {
        $$ = lalr17Acfg(lalr17lex).NewAtom("let", $2, $3)
    }
    ;

// ============================================================
// --- Scenario grammar (from Python ivy_parser.py:2202-2269) ---
// ============================================================

scenario:
    LALR17_TOK_SCENARIO LALR17_TOK_LCB sceninit LALR17_TOK_SEMI scentranss LALR17_TOK_RCB
    {
        elems := append([]ast.Node{$3}, $5...)
        sdef := lalr17Acfg(lalr17lex).NewScenarioDef(elems)
        $$ = lalr17Acfg(lalr17lex).NewScenarioDecl(sdef)
    }
    ;

sceninit:
    LALR17_TOK_ARROW places
    {
        $$ = lalr17Acfg(lalr17lex).NewPlaceList($2)
    }
    ;

places:
    LALR17_TOK_PRESYMBOL
    {
        $$ = []ast.Node{lalr17Acfg(lalr17lex).NewAtom($1)}
    }
    | places LALR17_TOK_COMMA LALR17_TOK_PRESYMBOL
    {
        $$ = append($1, lalr17Acfg(lalr17lex).NewAtom($3))
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
    places LALR17_TOK_ARROW places LALR17_TOK_COLON scenariomixin
    {
        $$ = lalr17Acfg(lalr17lex).NewScenarioTransition(lalr17Acfg(lalr17lex).NewPlaceList($1), lalr17Acfg(lalr17lex).NewPlaceList($3), $5)
    }
    | places LALR17_TOK_COLON scenariomixin
    {
        $$ = lalr17Acfg(lalr17lex).NewScenarioTransition(lalr17Acfg(lalr17lex).NewPlaceList($1), lalr17Acfg(lalr17lex).NewPlaceList(nil), $3)
    }
    ;

scenariomixin:
    LALR17_TOK_BEFORE atype sequence
    {
        atom := lalr17Acfg(lalr17lex).NewAtom($2.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixerName := fmt.Sprintf("%s[before%d]", atom.Rep, lalrLabelCounter)
        mixer := lalr17Acfg(lalr17lex).NewAtom(mixerName)
        adef := lalr17Acfg(lalr17lex).NewActionDef(atom, $3, nil, nil)
        $$ = lalr17Acfg(lalr17lex).NewScenarioBeforeMixin(mixer, adef)
    }
    | LALR17_TOK_AFTER atype sequence
    {
        atom := lalr17Acfg(lalr17lex).NewAtom($2.(*ast.Symbol).Rep)
        lalrLabelCounter++
        mixerName := fmt.Sprintf("%s[after%d]", atom.Rep, lalrLabelCounter)
        mixer := lalr17Acfg(lalr17lex).NewAtom(mixerName)
        adef := lalr17Acfg(lalr17lex).NewActionDef(atom, $3, nil, nil)
        $$ = lalr17Acfg(lalr17lex).NewScenarioAfterMixin(mixer, adef)
    }
    ;

// ========================================================================
// Proof / tactic productions — ported from Python ivy_parser.py.
// These define the grammar for proof scripts that invoke tactics like
// tempind, skolemizenp, vcgen, sorry, etc.
// ========================================================================

// pflet : var EQ fmla
pflet:
    var LALR17_TOK_EQ fmla
    {
        $$ = lalr17Acfg(lalr17lex).NewDefinition($1, $3)
    }
    ;

// pflets : pflet | pflets COMMA pflet
pflets:
    pflet
    {
        $$ = []ast.Node{$1}
    }
    | pflets LALR17_TOK_COMMA pflet
    {
        $$ = append($1, $3)
    }
    ;

// tacticwithelem : INVARIANT labeledfmla | DEFINITION atype EQ fmla | TRIGGER atype WITH terms
tacticwithelem:
    LALR17_TOK_INVARIANT labeledfmla
    {
        $$ = $2
    }
    | LALR17_TOK_DEFINITION atype LALR17_TOK_EQ fmla
    {
        $$ = lalr17Acfg(lalr17lex).NewDefinition($2, $4)
    }
    | LALR17_TOK_TRIGGER atype LALR17_TOK_WITH terms
    {
        $$ = lalr17Acfg(lalr17lex).NewTrigger(nil, append([]ast.Node{$2}, $4...)...)
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
        $$ = lalr17Acfg(lalr17lex).NewTacticWith($1)
    }
    | pflets
    {
        $$ = lalr17Acfg(lalr17lex).NewTacticLets($1)
    }
    ;

// opttacticwith : (empty) | WITH tacticwithlistchoice | WITH LCB tacticwithlist RCB
opttacticwith:
    /* empty */
    {
        $$ = lalr17Acfg(lalr17lex).NewTacticWith(nil)
    }
    | LALR17_TOK_WITH tacticwithlistchoice
    {
        $$ = $2
    }
    | LALR17_TOK_WITH LALR17_TOK_LCB tacticwithlist LALR17_TOK_RCB
    {
        $$ = lalr17Acfg(lalr17lex).NewTacticWith($3)
    }
    ;

// proofgroup : LCB proofseq RCB | LCB RCB
proofgroup:
    LALR17_TOK_LCB proofseq LALR17_TOK_RCB
    {
        $$ = $2
    }
    | LALR17_TOK_LCB LALR17_TOK_RCB
    {
        $$ = lalr17Acfg(lalr17lex).NewNullTactic()
    }
    ;

// optproofgroup : (empty) | proofgroup
optproofgroup:
    /* empty */
    {
        $$ = lalr17Acfg(lalr17lex).NewNoneAST()
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
    | proofseq LALR17_TOK_SEMI proofstep
    {
        $$ = lalr17Acfg(lalr17lex).NewComposeTactics([]ast.Node{$1, $3})
    }
    | proofseq proofstep
    {
        $$ = lalr17Acfg(lalr17lex).NewComposeTactics([]ast.Node{$1, $2})
    }
    ;

// proofstep — all the various proof step forms
proofstep:
    // proofstep : APPLY atype
    LALR17_TOK_APPLY atype
    {
        $$ = lalr17Acfg(lalr17lex).NewSchemaInstantiation($2, lalr17Acfg(lalr17lex).NewNoneAST())
    }
    // proofstep : ASSUME atype
    | LALR17_TOK_ASSUME atype
    {
        $$ = lalr17Acfg(lalr17lex).NewAssumeTactic($2, lalr17Acfg(lalr17lex).NewNoneAST())
    }
    // proofstep : SHOWGOALS
    | LALR17_TOK_SHOWGOALS
    {
        $$ = lalr17Acfg(lalr17lex).NewShowGoalsTactic()
    }
    // proofstep : DEFERGOAL
    | LALR17_TOK_DEFERGOAL
    {
        $$ = lalr17Acfg(lalr17lex).NewDeferGoalTactic()
    }
    // proofstep : SPOIL atype
    | LALR17_TOK_SPOIL atype
    {
        $$ = lalr17Acfg(lalr17lex).NewSpoilTactic($2)
    }
    // proofstep : TACTIC SYMBOL opttacticwith optproofgroup
    | LALR17_TOK_TACTIC atype opttacticwith optproofgroup
    {
        $$ = lalr17Acfg(lalr17lex).NewTacticTactic($2, $3, $4)
    }
    // proofstep : PROPERTY labeledfmla optproofgroup
    | LALR17_TOK_PROPERTY labeledfmla optproofgroup
    {
        $$ = lalr17Acfg(lalr17lex).NewPropertyTactic($2, lalr17Acfg(lalr17lex).NewNoneAST(), $3)
    }
    // proofstep : FUNCTION atype
    | LALR17_TOK_FUNCTION atype
    {
        $$ = lalr17Acfg(lalr17lex).NewFunctionTactic([]ast.Node{$2})
    }
    // proofstep : PROOF LABEL proofgroup
    | LALR17_TOK_PROOF LALR17_TOK_LABEL proofgroup
    {
        $$ = lalr17Acfg(lalr17lex).NewProofTactic(lalr17Acfg(lalr17lex).NewAtom($2), $3)
    }
    // proofstep : LET pflets
    | LALR17_TOK_LET pflets
    {
        $$ = lalr17Acfg(lalr17lex).NewLetTactic($2)
    }
    // proofstep : INSTANTIATE WITH pflets (witness tactic)
    | LALR17_TOK_INSTANTIATE LALR17_TOK_WITH pflets
    {
        $$ = lalr17Acfg(lalr17lex).NewWitnessTactic($3)
    }
    // proofstep : IF fmla proofgroup ELSE proofgroup
    | LALR17_TOK_IF fmla proofgroup LALR17_TOK_ELSE proofgroup
    {
        $$ = lalr17Acfg(lalr17lex).NewIfTactic($2, $3, $5)
    }
    // proofstep : UNFOLD WITH atype (simplified)
    | LALR17_TOK_UNFOLD LALR17_TOK_WITH atype
    {
        $$ = lalr17Acfg(lalr17lex).NewUnfoldTactic(lalr17Acfg(lalr17lex).NewNoneAST(), []ast.Node{$3})
    }
    // proofstep : FORGET atype
    | LALR17_TOK_FORGET atype
    {
        $$ = lalr17Acfg(lalr17lex).NewForgetTactic([]ast.Node{$2})
    }
    // proofstep : proofgroup (nested braces)
    | proofgroup
    {
        $$ = $1
    }
    ;

%%

// lalrMakeSequence wraps a list of action nodes into a single And node (sequence).
func lalr17MakeSequence(cfg *ast.AstConfig, stmts []ast.Node) ast.Node {
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
