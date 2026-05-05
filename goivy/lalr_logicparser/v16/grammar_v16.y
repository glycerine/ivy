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
	"github.com/glycerine/ivy/goivy/ast"
)

func lalr16AtypeToString(n ast.Node) string {
	switch v := n.(type) {
	case *ast.Symbol:
		return v.Rep
	case *ast.This:
		return "this"
	default:
		return fmt.Sprint(n)
	}
}

func lalr16Acfg(lex lalr16Lexer) *ast.AstConfig {
	return lex.(*lalr16LexAdapter).cfg
}

%}

%union {
	node     ast.Node
	nodes    []ast.Node
	str      string
}

%token <str>  LALR16_TOK_SYMBOL LALR16_TOK_VARIABLE LALR16_TOK_PRESYMBOL
%token        LALR16_TOK_LPAREN LALR16_TOK_RPAREN LALR16_TOK_LB LALR16_TOK_RB LALR16_TOK_LCB LALR16_TOK_RCB
%token        LALR16_TOK_COMMA LALR16_TOK_SEMI LALR16_TOK_COLON LALR16_TOK_DOT
%token        LALR16_TOK_PLUS LALR16_TOK_MINUS LALR16_TOK_TIMES LALR16_TOK_DIV
%token        LALR16_TOK_EQ LALR16_TOK_TILDAEQ LALR16_TOK_TILDA LALR16_TOK_LE LALR16_TOK_LT LALR16_TOK_GE LALR16_TOK_GT
%token        LALR16_TOK_AND LALR16_TOK_OR LALR16_TOK_ARROW LALR16_TOK_IFF
%token        LALR16_TOK_PTO LALR16_TOK_DOLLAR
%token        LALR16_TOK_FORALL LALR16_TOK_EXISTS
%token        LALR16_TOK_TRUE LALR16_TOK_FALSE
%token        LALR16_TOK_OLD LALR16_TOK_THIS LALR16_TOK_ISA
%token        LALR16_TOK_IF LALR16_TOK_ELSE
%token        LALR16_TOK_GLOBALLY LALR16_TOK_EVENTUALLY
%token        LALR16_TOK_WHENNEXT LALR16_TOK_WHENPREV LALR16_TOK_WHENFIRST LALR16_TOK_WHENLAST

%type <node>  top fmla term aterm var simplevar atype
%type <nodes> terms vars simplevars
%type <str>   relop SYMBOLx

// Precedence for v1.3–v1.6 (from Python ivy_parser.py):
%left         LALR16_TOK_SEMI
%left         LALR16_TOK_GLOBALLY LALR16_TOK_EVENTUALLY LALR16_TOK_WHENFIRST LALR16_TOK_WHENLAST LALR16_TOK_WHENNEXT LALR16_TOK_WHENPREV
%left         LALR16_TOK_IF
%left         LALR16_TOK_ELSE
%left         LALR16_TOK_OR
%left         LALR16_TOK_AND
%left         LALR16_TOK_TILDA
%left         LALR16_TOK_EQ LALR16_TOK_LE LALR16_TOK_LT LALR16_TOK_GE LALR16_TOK_GT LALR16_TOK_PTO
%left         LALR16_TOK_TILDAEQ
%left         LALR16_TOK_COLON
%left         LALR16_TOK_PLUS
%left         LALR16_TOK_MINUS
%left         LALR16_TOK_TIMES
%left         LALR16_TOK_DIV
%left         LALR16_TOK_DOLLAR

%start        top

%%

top:
    fmla
    {
        lalr16lex.(*lalr16LexAdapter).result = $1
    }
    ;

SYMBOLx:
    LALR16_TOK_PRESYMBOL
    {
        $$ = $1
    }
    ;

atype:
    SYMBOLx
    {
        $$ = lalr16Acfg(lalr16lex).NewSymbol($1, nil)
    }
    | atype LALR16_TOK_DOT SYMBOLx
    {
        if _, ok := $1.(*ast.This); ok {
            $$ = lalr16Acfg(lalr16lex).NewSymbol($3, nil)
        } else if sym, ok := $1.(*ast.Symbol); ok {
            $$ = lalr16Acfg(lalr16lex).NewSymbol(sym.Rep + "." + $3, nil)
        } else {
            $$ = lalr16Acfg(lalr16lex).NewSymbol($3, nil)
        }
    }
    | LALR16_TOK_THIS
    {
        $$ = lalr16Acfg(lalr16lex).NewThis()
    }
    ;

// --- aterm: v1.3–v1.6 term atoms (with DOT composition) ---

aterm:
    SYMBOLx
    {
        $$ = lalr16Acfg(lalr16lex).NewAtom($1)
    }
    | aterm LALR16_TOK_LPAREN terms LALR16_TOK_RPAREN
    {
        a := $1.(*ast.Atom)
        a.Terms = append(a.Terms, $3...)
        $$ = a
    }
    | aterm LALR16_TOK_DOT SYMBOLx
    {
        lhs := $1.(*ast.Atom)
        $$ = lalr16Acfg(lalr16lex).NewAtom(lhs.Rep + "." + $3, lhs.Terms...)
    }
    ;

var:
    LALR16_TOK_VARIABLE
    { $$ = lalr16Acfg(lalr16lex).NewVariable($1, "S") }
    | LALR16_TOK_VARIABLE LALR16_TOK_COLON atype
    { $$ = lalr16Acfg(lalr16lex).NewVariable($1, lalr16AtypeToString($3)) }
    ;

simplevar:
    LALR16_TOK_VARIABLE
    { $$ = lalr16Acfg(lalr16lex).NewVariable($1, "S") }
    | LALR16_TOK_VARIABLE LALR16_TOK_COLON SYMBOLx
    { $$ = lalr16Acfg(lalr16lex).NewVariable($1, $3) }
    ;

vars:
    var
    { $$ = []ast.Node{$1} }
    | vars LALR16_TOK_COMMA var
    { $$ = append($1, $3) }
    ;

simplevars:
    simplevar
    { $$ = []ast.Node{$1} }
    | simplevars LALR16_TOK_COMMA simplevar
    { $$ = append($1, $3) }
    ;

terms:
    /* empty */
    { $$ = nil }
    | term
    { $$ = []ast.Node{$1} }
    | terms LALR16_TOK_COMMA term
    { $$ = append($1, $3) }
    ;

// --- term: in v1.6, terms do NOT include boolean/comparison ops ---

term:
    aterm
    { $$ = $1 }
    | var
    { $$ = $1 }
    | LALR16_TOK_OLD aterm
    { $$ = lalr16Acfg(lalr16lex).NewOld($2) }
    | LALR16_TOK_LPAREN term LALR16_TOK_RPAREN
    { $$ = $2 }
    | term LALR16_TOK_PLUS term
    { $$ = lalr16Acfg(lalr16lex).NewAtom("+", $1, $3) }
    | term LALR16_TOK_MINUS term
    { $$ = lalr16Acfg(lalr16lex).NewAtom("-", $1, $3) }
    | term LALR16_TOK_TIMES term
    { $$ = lalr16Acfg(lalr16lex).NewAtom("*", $1, $3) }
    | term LALR16_TOK_DIV term
    { $$ = lalr16Acfg(lalr16lex).NewAtom("/", $1, $3) }
    | term LALR16_TOK_IF fmla LALR16_TOK_ELSE term
    { $$ = lalr16Acfg(lalr16lex).NewIte($3, $1, $5) }
    ;

relop:
    LALR16_TOK_EQ  { $$ = "=" }
    | LALR16_TOK_LE  { $$ = "<=" }
    | LALR16_TOK_LT  { $$ = "<" }
    | LALR16_TOK_GE  { $$ = ">=" }
    | LALR16_TOK_GT  { $$ = ">" }
    | LALR16_TOK_PTO { $$ = "*>" }
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
        $$ = lalr16Acfg(lalr16lex).NewAtom($2, $1, $3)
    }
    | term LALR16_TOK_TILDAEQ term
    {
        $$ = lalr16Acfg(lalr16lex).NewNot(lalr16Acfg(lalr16lex).NewAtom("=", $1, $3))
    }
    | LALR16_TOK_LPAREN fmla LALR16_TOK_RPAREN
    { $$ = $2 }
    | LALR16_TOK_TRUE
    { $$ = lalr16Acfg(lalr16lex).NewAnd() }
    | LALR16_TOK_FALSE
    { $$ = lalr16Acfg(lalr16lex).NewOr() }
    | LALR16_TOK_TILDA fmla
    { $$ = lalr16Acfg(lalr16lex).NewNot($2) }
    | fmla LALR16_TOK_AND fmla
    { $$ = lalr16Acfg(lalr16lex).NewAnd($1, $3) }
    | fmla LALR16_TOK_OR fmla
    { $$ = lalr16Acfg(lalr16lex).NewOr($1, $3) }
    | fmla LALR16_TOK_ARROW fmla
    { $$ = lalr16Acfg(lalr16lex).NewImplies($1, $3) }
    | fmla LALR16_TOK_IFF fmla
    { $$ = lalr16Acfg(lalr16lex).NewIff($1, $3) }
    | LALR16_TOK_FORALL simplevars LALR16_TOK_DOT fmla
    { $$ = lalr16Acfg(lalr16lex).NewForall($2, $4) }
    | LALR16_TOK_EXISTS simplevars LALR16_TOK_DOT fmla
    { $$ = lalr16Acfg(lalr16lex).NewExists($2, $4) }
    | LALR16_TOK_GLOBALLY fmla
    { $$ = lalr16Acfg(lalr16lex).NewGlobally($2) }
    | LALR16_TOK_EVENTUALLY fmla
    { $$ = lalr16Acfg(lalr16lex).NewEventually($2) }
    ;

%%
