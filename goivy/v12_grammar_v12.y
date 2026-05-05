// grammar_v12.y — goyacc LALR(1) grammar for Ivy v1.2 and earlier.
// Key differences:
// - No arithmetic operators (+, -, *, /)
// - No temporal operators (globally, eventually)
// - Composition via COLON, not DOT
// - TILDA binds LOOSER than comparison (opposite of v1.7+!)
// - No ARROW (implies)
// - Simpler precedence table (12 levels)

%{
package goivy

import (
	"fmt"
)

func lalr12AtypeToString(n Node) string {
	switch v := n.(type) {
	case *Symbol:
		return v.Rep
	case *This:
		return "this"
	default:
		return fmt.Sprint(n)
	}
}

func lalr12Acfg(lex lalr12Lexer) *AstConfig {
	return lex.(*lalr12LexAdapter).cfg
}

%}

%union {
	node     Node
	nodes    []Node
	str      string
}

%token <str>  LALR12_TOK_SYMBOL LALR12_TOK_VARIABLE LALR12_TOK_PRESYMBOL
%token        LALR12_TOK_LPAREN LALR12_TOK_RPAREN LALR12_TOK_LB LALR12_TOK_RB LALR12_TOK_LCB LALR12_TOK_RCB
%token        LALR12_TOK_COMMA LALR12_TOK_SEMI LALR12_TOK_COLON LALR12_TOK_DOT
%token        LALR12_TOK_PLUS LALR12_TOK_MINUS LALR12_TOK_TIMES LALR12_TOK_DIV
%token        LALR12_TOK_EQ LALR12_TOK_TILDAEQ LALR12_TOK_TILDA LALR12_TOK_LE LALR12_TOK_LT LALR12_TOK_GE LALR12_TOK_GT
%token        LALR12_TOK_AND LALR12_TOK_OR LALR12_TOK_ARROW LALR12_TOK_IFF
%token        LALR12_TOK_PTO LALR12_TOK_DOLLAR
%token        LALR12_TOK_FORALL LALR12_TOK_EXISTS
%token        LALR12_TOK_TRUE LALR12_TOK_FALSE
%token        LALR12_TOK_OLD LALR12_TOK_THIS LALR12_TOK_ISA
%token        LALR12_TOK_IF LALR12_TOK_ELSE
%token        LALR12_TOK_GLOBALLY LALR12_TOK_EVENTUALLY
%token        LALR12_TOK_WHENNEXT LALR12_TOK_WHENPREV LALR12_TOK_WHENFIRST LALR12_TOK_WHENLAST

%type <node>  top fmla term aterm var simplevar atype
%type <nodes> terms vars simplevars
%type <str>   relop SYMBOLx

// Precedence for v1.2 and earlier (from Python ivy_parser.py):
// NOTE: TILDA is BELOW comparison operators here!
%left         LALR12_TOK_SEMI
%left         LALR12_TOK_IF
%left         LALR12_TOK_ELSE
%left         LALR12_TOK_OR
%left         LALR12_TOK_AND
%left         LALR12_TOK_PLUS
%left         LALR12_TOK_TIMES
%left         LALR12_TOK_DIV
%left         LALR12_TOK_TILDA
%left         LALR12_TOK_EQ LALR12_TOK_LE LALR12_TOK_LT LALR12_TOK_GE LALR12_TOK_GT
%left         LALR12_TOK_TILDAEQ
%left         LALR12_TOK_COLON

%start        top

%%

top:
    fmla
    {
        lalr12lex.(*lalr12LexAdapter).result = $1
    }
    ;

SYMBOLx:
    LALR12_TOK_PRESYMBOL
    { $$ = $1 }
    ;

atype:
    SYMBOLx
    { $$ = lalr12Acfg(lalr12lex).NewSymbol($1, nil) }
    ;

// --- aterm: v1.2 uses COLON for composition ---

aterm:
    SYMBOLx
    { $$ = lalr12Acfg(lalr12lex).NewAtom($1) }
    | aterm LALR12_TOK_LPAREN terms LALR12_TOK_RPAREN
    {
        a := $1.(*Atom)
        a.Terms = append(a.Terms, $3...)
        $$ = a
    }
    | aterm LALR12_TOK_COLON SYMBOLx
    {
        lhs := $1.(*Atom)
        $$ = lalr12Acfg(lalr12lex).NewAtom(lhs.Rep + ":" + $3, lhs.Terms...)
    }
    ;

var:
    LALR12_TOK_VARIABLE
    { $$ = lalr12Acfg(lalr12lex).NewVariable($1, "S") }
    | LALR12_TOK_VARIABLE LALR12_TOK_COLON atype
    { $$ = lalr12Acfg(lalr12lex).NewVariable($1, lalr12AtypeToString($3)) }
    ;

simplevar:
    LALR12_TOK_VARIABLE
    { $$ = lalr12Acfg(lalr12lex).NewVariable($1, "S") }
    | LALR12_TOK_VARIABLE LALR12_TOK_COLON SYMBOLx
    { $$ = lalr12Acfg(lalr12lex).NewVariable($1, $3) }
    ;

vars:
    var
    { $$ = []Node{$1} }
    | vars LALR12_TOK_COMMA var
    { $$ = append($1, $3) }
    ;

simplevars:
    simplevar
    { $$ = []Node{$1} }
    | simplevars LALR12_TOK_COMMA simplevar
    { $$ = append($1, $3) }
    ;

terms:
    /* empty */
    { $$ = nil }
    | term
    { $$ = []Node{$1} }
    | terms LALR12_TOK_COMMA term
    { $$ = append($1, $3) }
    ;

// --- term: v1.2 has NO arithmetic ops ---

term:
    aterm
    { $$ = $1 }
    | var
    { $$ = $1 }
    | LALR12_TOK_OLD aterm
    { $$ = lalr12Acfg(lalr12lex).NewOld($2) }
    | LALR12_TOK_LPAREN term LALR12_TOK_RPAREN
    { $$ = $2 }
    ;

relop:
    LALR12_TOK_EQ  { $$ = "=" }
    | LALR12_TOK_LE  { $$ = "<=" }
    | LALR12_TOK_LT  { $$ = "<" }
    | LALR12_TOK_GE  { $$ = ">=" }
    | LALR12_TOK_GT  { $$ = ">" }
    ;

// --- fmla: v1.2 formulas (no ARROW, no temporal) ---

fmla:
    term
    { $$ = $1 }
    | term relop term
    { $$ = lalr12Acfg(lalr12lex).NewAtom($2, $1, $3) }
    | term LALR12_TOK_TILDAEQ term
    { $$ = lalr12Acfg(lalr12lex).NewNot(lalr12Acfg(lalr12lex).NewAtom("=", $1, $3)) }
    | LALR12_TOK_LPAREN fmla LALR12_TOK_RPAREN
    { $$ = $2 }
    | LALR12_TOK_TRUE
    { $$ = lalr12Acfg(lalr12lex).NewAnd() }
    | LALR12_TOK_FALSE
    { $$ = lalr12Acfg(lalr12lex).NewOr() }
    | LALR12_TOK_TILDA fmla
    { $$ = lalr12Acfg(lalr12lex).NewNot($2) }
    | fmla LALR12_TOK_AND fmla
    { $$ = lalr12Acfg(lalr12lex).NewAnd($1, $3) }
    | fmla LALR12_TOK_OR fmla
    { $$ = lalr12Acfg(lalr12lex).NewOr($1, $3) }
    | fmla LALR12_TOK_IFF fmla
    { $$ = lalr12Acfg(lalr12lex).NewIff($1, $3) }
    | LALR12_TOK_FORALL simplevars LALR12_TOK_DOT fmla
    { $$ = lalr12Acfg(lalr12lex).NewForall($2, $4) }
    | LALR12_TOK_EXISTS simplevars LALR12_TOK_DOT fmla
    { $$ = lalr12Acfg(lalr12lex).NewExists($2, $4) }
    ;

%%
