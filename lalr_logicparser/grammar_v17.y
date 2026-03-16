// grammar_v17.y — goyacc LALR(1) grammar for Ivy formula/term parsing, version 1.7+.
// Mechanically translated from Python PLY grammar in ivy_logic_parser.py.
//
// In v1.7+, terms subsume formulas: comparison operators, boolean connectives,
// quantifiers, and temporal operators are all term-level constructs.
//
// Precedence table (copied exactly from Python ivy_parser.py, v1.7+):
//   SEMI < GLOBALLY/EVENTUALLY < ARROW/IFF < OR < AND < TILDA
//   < EQ/LE/LT/GE/GT/PTO < TILDAEQ < IF < ELSE < COLON
//   < PLUS/MINUS < TIMES/DIV < DOLLAR < OLD < DOT

%{
package lalr_logicparser

import (
	"github.com/glycerine/goivy/ast"
)

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

// Nonterminal types
%type <node>  top term fmla appelem var simplevar atype
%type <nodes> terms vars simplevars
%type <str>   SYMBOLx SYMsubscr

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

%start        top

%%

top:
    fmla
    {
        v17lex.(*v17LexAdapter).result = $1
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
        v := &ast.Variable{Rep: $1}
        v.VSort = $3
        $$ = v
    }
    ;

simplevar:
    TOK_VARIABLE
    {
        $$ = &ast.Variable{Rep: $1}
    }
    | TOK_VARIABLE TOK_COLON SYMBOLx
    {
        v := &ast.Variable{Rep: $1}
        v.VSort = &ast.Symbol{Rep: $3}
        $$ = v
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
    // --- Arithmetic ---
    | term TOK_PLUS term
    {
        $$ = &ast.Atom{Rep: "+", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_MINUS term
    {
        $$ = &ast.Atom{Rep: "-", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_TIMES term
    {
        $$ = &ast.Atom{Rep: "*", Terms: []ast.Node{$1, $3}}
    }
    | term TOK_DIV term
    {
        $$ = &ast.Atom{Rep: "/", Terms: []ast.Node{$1, $3}}
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
        $$ = &ast.Atom{Rep: "*>", Terms: []ast.Node{$1, $3}}
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
            v.VSort = $3
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

%%
