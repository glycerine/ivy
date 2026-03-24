// grammar_v16.y — goyacc LALR(1) grammar for Ivy v1.3–v1.6.
// Key differences from v1.7+:
// - Formulas and terms are SEPARATE categories (fmla vs term)
// - Comparison operators only appear in fmla rules, not term rules
// - WHEN operators in precedence table
// - IF/ELSE bind looser than OR/AND
// - ARROW/IFF only in fmla rules (no explicit precedence entry)

%{
package v16

import (
	"fmt"
	"github.com/glycerine/goivy/ast"
)

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

%union {
	node     ast.Node
	nodes    []ast.Node
	str      string
}

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

%type <node>  top fmla term aterm var simplevar atype
%type <nodes> terms vars simplevars
%type <str>   relop SYMBOLx

// Precedence for v1.3–v1.6 (from Python ivy_parser.py):
%left         TOK_SEMI
%left         TOK_GLOBALLY TOK_EVENTUALLY TOK_WHENFIRST TOK_WHENLAST TOK_WHENNEXT TOK_WHENPREV
%left         TOK_IF
%left         TOK_ELSE
%left         TOK_OR
%left         TOK_AND
%left         TOK_TILDA
%left         TOK_EQ TOK_LE TOK_LT TOK_GE TOK_GT TOK_PTO
%left         TOK_TILDAEQ
%left         TOK_COLON
%left         TOK_PLUS
%left         TOK_MINUS
%left         TOK_TIMES
%left         TOK_DIV
%left         TOK_DOLLAR

%start        top

%%

top:
    fmla
    {
        v16lex.(*v16LexAdapter).result = $1
    }
    ;

SYMBOLx:
    TOK_PRESYMBOL
    {
        $$ = $1
    }
    ;

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

// --- aterm: v1.3–v1.6 term atoms (with DOT composition) ---

aterm:
    SYMBOLx
    {
        $$ = &ast.Atom{Rep: $1}
    }
    | aterm TOK_LPAREN terms TOK_RPAREN
    {
        a := $1.(*ast.Atom)
        a.Terms = append(a.Terms, $3...)
        $$ = a
    }
    | aterm TOK_DOT SYMBOLx
    {
        lhs := $1.(*ast.Atom)
        $$ = &ast.Atom{Rep: lhs.Rep + "." + $3, Terms: lhs.Terms}
    }
    ;

var:
    TOK_VARIABLE
    { $$ = &ast.Variable{Rep: $1, VSort: "S"} }
    | TOK_VARIABLE TOK_COLON atype
    { $$ = &ast.Variable{Rep: $1, VSort: atypeToString($3)} }
    ;

simplevar:
    TOK_VARIABLE
    { $$ = &ast.Variable{Rep: $1, VSort: "S"} }
    | TOK_VARIABLE TOK_COLON SYMBOLx
    { $$ = &ast.Variable{Rep: $1, VSort: $3} }
    ;

vars:
    var
    { $$ = []ast.Node{$1} }
    | vars TOK_COMMA var
    { $$ = append($1, $3) }
    ;

simplevars:
    simplevar
    { $$ = []ast.Node{$1} }
    | simplevars TOK_COMMA simplevar
    { $$ = append($1, $3) }
    ;

terms:
    /* empty */
    { $$ = nil }
    | term
    { $$ = []ast.Node{$1} }
    | terms TOK_COMMA term
    { $$ = append($1, $3) }
    ;

// --- term: in v1.6, terms do NOT include boolean/comparison ops ---

term:
    aterm
    { $$ = $1 }
    | var
    { $$ = $1 }
    | TOK_OLD aterm
    { $$ = &ast.Old{Term: $2} }
    | TOK_LPAREN term TOK_RPAREN
    { $$ = $2 }
    | term TOK_PLUS term
    { $$ = &ast.Atom{Rep: "+", Terms: []ast.Node{$1, $3}} }
    | term TOK_MINUS term
    { $$ = &ast.Atom{Rep: "-", Terms: []ast.Node{$1, $3}} }
    | term TOK_TIMES term
    { $$ = &ast.Atom{Rep: "*", Terms: []ast.Node{$1, $3}} }
    | term TOK_DIV term
    { $$ = &ast.Atom{Rep: "/", Terms: []ast.Node{$1, $3}} }
    | term TOK_IF fmla TOK_ELSE term
    { $$ = &ast.Ite{Cond: $3, Then: $1, Else: $5} }
    ;

relop:
    TOK_EQ  { $$ = "=" }
    | TOK_LE  { $$ = "<=" }
    | TOK_LT  { $$ = "<" }
    | TOK_GE  { $$ = ">=" }
    | TOK_GT  { $$ = ">" }
    | TOK_PTO { $$ = "*>" }
    ;

// --- fmla: in v1.6, formulas are a separate category ---

fmla:
    term
    {
        // Convert bare term to atom (app_to_atom)
        $$ = $1
    }
    | term relop term
    {
        $$ = &ast.Atom{Rep: $2, Terms: []ast.Node{$1, $3}}
    }
    | term TOK_TILDAEQ term
    {
        $$ = &ast.Not{Body: &ast.Atom{Rep: "=", Terms: []ast.Node{$1, $3}}}
    }
    | TOK_LPAREN fmla TOK_RPAREN
    { $$ = $2 }
    | TOK_TRUE
    { $$ = &ast.And{} }
    | TOK_FALSE
    { $$ = &ast.Or{} }
    | TOK_TILDA fmla
    { $$ = &ast.Not{Body: $2} }
    | fmla TOK_AND fmla
    { $$ = &ast.And{Terms: []ast.Node{$1, $3}} }
    | fmla TOK_OR fmla
    { $$ = &ast.Or{Terms: []ast.Node{$1, $3}} }
    | fmla TOK_ARROW fmla
    { $$ = &ast.Implies{T1: $1, T2: $3} }
    | fmla TOK_IFF fmla
    { $$ = &ast.Iff{T1: $1, T2: $3} }
    | TOK_FORALL simplevars TOK_DOT fmla
    { $$ = &ast.Forall{Bounds: $2, Body: $4} }
    | TOK_EXISTS simplevars TOK_DOT fmla
    { $$ = &ast.Exists{Bounds: $2, Body: $4} }
    | TOK_GLOBALLY fmla
    { $$ = &ast.Globally{Body: $2} }
    | TOK_EVENTUALLY fmla
    { $$ = &ast.Eventually{Body: $2} }
    ;

%%
