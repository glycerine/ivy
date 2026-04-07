// cs_grammar.y — goyacc LALR(1) grammar for concept space parsing.
// Mechanically translated from Python PLY grammar in ivy_concept_space.py.
//
// Grammar:
//   expr  : lit | '(' prod ')' | '(' sum ')'
//   term  : SYMBOL    (uppercase → Variable, else Constant)
//   terms : /* empty */ | term | terms ',' term
//   atom  : SYMBOL '(' terms ')'
//   lit   : atom | '~' atom
//   prod  : expr '*' expr | prod '*' expr
//   sum   : expr '+' expr | sum '+' expr

%{
package conceptspace

import (
    "unicode"

    il "github.com/glycerine/ivy/goivy/ivylogic"
    lg "github.com/glycerine/ivy/goivy/logic"
)
%}

%union {
    space   Space
    spaces  []Space
    lit     *il.Literal
    atom    lg.Node
    term    lg.Node
    terms   []lg.Node
    str     string
}

%token <str>  CS_SYMBOL
%token        CS_COMMA CS_LPAREN CS_RPAREN CS_LBR CS_RBR
%token        CS_PLUS CS_TIMES CS_TILDA

%type <space>   top expr
%type <spaces>  prod sum
%type <lit>     lit
%type <atom>    atom
%type <term>    term
%type <terms>   terms

%start top

%%

top:
    expr
    {
        cslex.(*csLexAdapter).result = $1
    }
    ;

expr:
    lit
    {
        $$ = &NamedSpace{Lit: $1}
    }
    | CS_LPAREN prod CS_RPAREN
    {
        $$ = &ProductSpace{Spaces: $2}
    }
    | CS_LPAREN sum CS_RPAREN
    {
        $$ = &SumSpace{Spaces: $2}
    }
    ;

term:
    CS_SYMBOL
    {
        name := $1
        if len(name) > 0 && unicode.IsUpper(rune(name[0])) {
            v, _ := lg.NewVar(name, lg.TopS)
            $$ = v
        } else {
            $$ = lg.NewConst(name, lg.TopS)
        }
    }
    ;

terms:
    /* empty */
    {
        $$ = nil
    }
    | term
    {
        $$ = []lg.Node{$1}
    }
    | terms CS_COMMA term
    {
        $$ = append($1, $3)
    }
    ;

atom:
    CS_SYMBOL CS_LPAREN terms CS_RPAREN
    {
        sym := lg.NewConst($1, lg.TopS)
        $$ = lg.MustApply(sym, $3...)
    }
    ;

lit:
    atom
    {
        $$ = il.NewLiteral(1, $1)
    }
    | CS_TILDA atom
    {
        $$ = il.NewLiteral(0, $2)
    }
    ;

prod:
    expr CS_TIMES expr
    {
        $$ = []Space{$1, $3}
    }
    | prod CS_TIMES expr
    {
        $$ = append($1, $3)
    }
    ;

sum:
    expr CS_PLUS expr
    {
        $$ = []Space{$1, $3}
    }
    | sum CS_PLUS expr
    {
        $$ = append($1, $3)
    }
    ;

%%
