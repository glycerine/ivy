// grammar_v12.y — goyacc LALR(1) grammar for Ivy v1.2 and earlier.
// Key differences:
// - No arithmetic operators (+, -, *, /)
// - No temporal operators (globally, eventually)
// - Composition via COLON, not DOT
// - TILDA binds LOOSER than comparison (opposite of v1.7+!)
// - No ARROW (implies)
// - Simpler precedence table (12 levels)

%{
package v12

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

func acfg(lex v12Lexer) *ast.AstConfig {
	return lex.(*v12LexAdapter).cfg
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

// Precedence for v1.2 and earlier (from Python ivy_parser.py):
// NOTE: TILDA is BELOW comparison operators here!
%left         TOK_SEMI
%left         TOK_IF
%left         TOK_ELSE
%left         TOK_OR
%left         TOK_AND
%left         TOK_PLUS
%left         TOK_TIMES
%left         TOK_DIV
%left         TOK_TILDA
%left         TOK_EQ TOK_LE TOK_LT TOK_GE TOK_GT
%left         TOK_TILDAEQ
%left         TOK_COLON

%start        top

%%

top:
    fmla
    {
        v12lex.(*v12LexAdapter).result = $1
    }
    ;

SYMBOLx:
    TOK_PRESYMBOL
    { $$ = $1 }
    ;

atype:
    SYMBOLx
    { $$ = acfg(v12lex).NewSymbol($1, nil) }
    ;

// --- aterm: v1.2 uses COLON for composition ---

aterm:
    SYMBOLx
    { $$ = acfg(v12lex).NewAtom($1) }
    | aterm TOK_LPAREN terms TOK_RPAREN
    {
        a := $1.(*ast.Atom)
        a.Terms = append(a.Terms, $3...)
        $$ = a
    }
    | aterm TOK_COLON SYMBOLx
    {
        lhs := $1.(*ast.Atom)
        $$ = acfg(v12lex).NewAtom(lhs.Rep + ":" + $3, lhs.Terms...)
    }
    ;

var:
    TOK_VARIABLE
    { $$ = acfg(v12lex).NewVariable($1, "S") }
    | TOK_VARIABLE TOK_COLON atype
    { $$ = acfg(v12lex).NewVariable($1, atypeToString($3)) }
    ;

simplevar:
    TOK_VARIABLE
    { $$ = acfg(v12lex).NewVariable($1, "S") }
    | TOK_VARIABLE TOK_COLON SYMBOLx
    { $$ = acfg(v12lex).NewVariable($1, $3) }
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

// --- term: v1.2 has NO arithmetic ops ---

term:
    aterm
    { $$ = $1 }
    | var
    { $$ = $1 }
    | TOK_OLD aterm
    { $$ = acfg(v12lex).NewOld($2) }
    | TOK_LPAREN term TOK_RPAREN
    { $$ = $2 }
    ;

relop:
    TOK_EQ  { $$ = "=" }
    | TOK_LE  { $$ = "<=" }
    | TOK_LT  { $$ = "<" }
    | TOK_GE  { $$ = ">=" }
    | TOK_GT  { $$ = ">" }
    ;

// --- fmla: v1.2 formulas (no ARROW, no temporal) ---

fmla:
    term
    { $$ = $1 }
    | term relop term
    { $$ = acfg(v12lex).NewAtom($2, $1, $3) }
    | term TOK_TILDAEQ term
    { $$ = acfg(v12lex).NewNot(acfg(v12lex).NewAtom("=", $1, $3)) }
    | TOK_LPAREN fmla TOK_RPAREN
    { $$ = $2 }
    | TOK_TRUE
    { $$ = acfg(v12lex).NewAnd() }
    | TOK_FALSE
    { $$ = acfg(v12lex).NewOr() }
    | TOK_TILDA fmla
    { $$ = acfg(v12lex).NewNot($2) }
    | fmla TOK_AND fmla
    { $$ = acfg(v12lex).NewAnd($1, $3) }
    | fmla TOK_OR fmla
    { $$ = acfg(v12lex).NewOr($1, $3) }
    | fmla TOK_IFF fmla
    { $$ = acfg(v12lex).NewIff($1, $3) }
    | TOK_FORALL simplevars TOK_DOT fmla
    { $$ = acfg(v12lex).NewForall($2, $4) }
    | TOK_EXISTS simplevars TOK_DOT fmla
    { $$ = acfg(v12lex).NewExists($2, $4) }
    ;

%%
