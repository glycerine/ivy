// ev_grammar.y — goyacc LALR(1) grammar for event trace parsing.
// Mechanically translated from Python PLY grammar in ivy_ev_parser.py.
//
// Grammar:
//   events : /* empty */ | events event
//   event  : optdir SYMBOL optargs optsubs
//   optdir : /* empty */ | GT | LT
//   optsubs: /* empty */ | SEMI | LCB events RCB
//   optargs: /* empty */ | LPAREN list RPAREN
//   value  : SYMBOL | SYMBOL LPAREN list RPAREN | LBR RBR | LBR list RBR | LCB RCB | LCB dict RCB
//   list   : value | list COMMA value
//   dict   : SYMBOL COLON value | dict COMMA SYMBOL COLON value

%{
package evparser
%}

%union {
    events   Events
    event    *Event
    dir      EventDir
    values   []Value
    value    Value
    dict     *DictValue
    str      string
}

%token <str>  EV_SYMBOL
%token        EV_COMMA EV_LPAREN EV_RPAREN EV_LBR EV_RBR EV_LCB EV_RCB
%token        EV_SEMI EV_COLON EV_GT EV_LT

%type <events>  top events
%type <event>   event
%type <dir>     optdir
%type <events>  optsubs
%type <values>  optargs list
%type <value>   value
%type <dict>    dict

%start top

%%

top:
    events
    {
        evlex.(*evLexAdapter).result = $1
    }
    ;

events:
    /* empty */
    {
        $$ = nil
    }
    | events event
    {
        $$ = append($1, $2)
    }
    ;

event:
    optdir EV_SYMBOL optargs optsubs
    {
        $$ = &Event{Dir: $1, Rep: $2, Args: $3, Children: $4}
    }
    ;

optdir:
    /* empty */
    {
        $$ = DirNone
    }
    | EV_GT
    {
        $$ = DirIn
    }
    | EV_LT
    {
        $$ = DirOut
    }
    ;

optsubs:
    /* empty */
    {
        $$ = nil
    }
    | EV_SEMI
    {
        $$ = nil
    }
    | EV_LCB events EV_RCB
    {
        $$ = $2
    }
    ;

optargs:
    /* empty */
    {
        $$ = nil
    }
    | EV_LPAREN list EV_RPAREN
    {
        $$ = $2
    }
    ;

value:
    EV_SYMBOL
    {
        $$ = &Symbol{Name: $1}
    }
    | EV_SYMBOL EV_LPAREN list EV_RPAREN
    {
        $$ = &App{Rep: $1, Args: $3}
    }
    | EV_LBR EV_RBR
    {
        $$ = &ListValue{}
    }
    | EV_LBR list EV_RBR
    {
        $$ = &ListValue{Items: $2}
    }
    | EV_LCB EV_RCB
    {
        $$ = NewDictValue()
    }
    | EV_LCB dict EV_RCB
    {
        $$ = $2
    }
    ;

list:
    value
    {
        $$ = []Value{$1}
    }
    | list EV_COMMA value
    {
        $$ = append($1, $3)
    }
    ;

dict:
    EV_SYMBOL EV_COLON value
    {
        d := NewDictValue()
        d.Set($1, $3)
        $$ = d
    }
    | dict EV_COMMA EV_SYMBOL EV_COLON value
    {
        $1.Set($3, $5)
        $$ = $1
    }
    ;

%%
