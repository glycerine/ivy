# Plan: Extract Complete Python Ivy Grammar to IVY_PYTHON_GRAMMAR.md

**Created:** 2026-03-23 19:30

## Context

Our Go LALR grammar port missed the `SYMBOL : SYMBOL LB SYMsubscr RB` rule, which
caused parsing failures for subscripted symbols like `bv[length]`. To prevent similar
omissions, we need a complete, bit-accurate reference document of every grammar rule
in the Python Ivy parser.

## Source Files

The Python Ivy grammar is defined across multiple files:

| File | Role | Rule Count |
|------|------|------------|
| `ivy_parser.py` (3654 lines) | Main parser, start='top' | ~290 p_ functions |
| `ivy_logic_parser.py` (769 lines) | Logic/formula/term rules, imported by ivy_parser.py | ~90 p_ functions |
| `ivy_logic_parser_gen.py` (40 lines) | Creates formula_parser (start='fmla') and term_parser (start='term') | 0 p_ functions |
| `ivy_lexer.py` (305 lines) | Token definitions, reserved words, lexer rules | Lexer only |
| `ivy_concept_space.py` | Separate concept space parser (not used by ivy_check) | 15 p_ functions |
| `ivy_ev_parser.py` | Separate event parser (not used by ivy_check) | 21 p_ functions |
| `ivy_dafny_grammar.py` | Separate Dafny parser (not used by ivy_check) | 60+ p_ functions |

**Critical:** `ivy_parser.py` line 3540 does `from .ivy_logic_parser import *`, merging
all logic parser rules into the main parser. The combined grammar has ~380 rules.

## Architecture

- PLY (Python Lex-Yacc) parser generator
- Version-conditional rules via `if iu.get_numeric_version() <= [1,N]:` guards
- Three version "eras": v1.0-1.2, v1.3-1.6, v1.7+ (our target)
- `parse()` function at line 3609 creates parser, calls `expand_autoinstances()` for non-nested parses
- `Ivy` class (line 255) is the accumulator (our `ivyAccum`)

## Execution Plan

When executing, create `~/goivy/IVY_PYTHON_GRAMMAR.md` with the content below.
This is a single file write operation.

---

## COMPLETE GRAMMAR DOCUMENT CONTENT

Below is the full grammar organized for reference. Each section contains the
exact productions from the Python source.

### 1. TOKENS (ivy_lexer.py)

#### Base Tokens (line 8-49)
```
COMMA  LPAREN  RPAREN  PLUS  TIMES  DIV  TILDA  AND  OR  EQ  TILDAEQ
SEMI  ASSIGN  DOT  LCB  RCB  ARROW  IFF  PRESYMBOL  VARIABLE  COLON
LE  LT  GE  GT  MINUS  DOTS  DOTDOTDOT  NATIVEQUOTE  PTO  DOLLAR  CARET
LB  RB  WHENFIRST  WHENLAST  WHENPREV  WHENNEXT  UNPROVABLE  TRIGGER
```

#### Reserved Keywords (line 51-173) — all become tokens
```
relation->RELATION  individual->INDIV  function->FUNCTION  axiom->AXIOM
conjecture->CONJECTURE  schema->SCHEMA  instantiate/instance->INSTANTIATE
derived->DERIVED  concept->CONCEPT  init->INIT  action->ACTION  method->METHOD
field->FIELD  state->STATE  assume->ASSUME  assert->ASSERT  set->SET  null->NULL
old->OLD  from->FROM  update->UPDATE  params->PARAMS  in->IN  match->MATCH
ensures->ENSURES  requires->REQUIRES  modifies->MODIFIES  true->TRUE  false->FALSE
fresh->FRESH  module/template->MODULE  object->OBJECT  class->CLASS  type->TYPE
if->IF  else->ELSE  local->LOCAL  let->LET  call->CALL  entry->ENTRY  macro->MACRO
interpret->INTERPRET  forall->FORALL  exists->EXISTS  returns->RETURNS
mixin/execute->MIXIN  before->BEFORE  after->AFTER  isolate->ISOLATE  with->WITH
export->EXPORT  delegate->DELEGATE  import->IMPORT  using->USING  include->INCLUDE
progress->PROGRESS  rely->RELY  mixord->MIXORD  extract/process->EXTRACT
destructor->DESTRUCTOR  some->SOME  maximizing->MAXIMIZING  minimizing->MINIMIZING
private->PRIVATE  implement->IMPLEMENT  property->PROPERTY  while->WHILE
invariant->INVARIANT  struct->STRUCT  definition->DEFINITION  ghost->GHOST
alias->ALIAS  trusted->TRUSTED  this->THIS  var->VAR  attribute->ATTRIBUTE
variant->VARIANT  of->OF  scenario->SCENARIO  proof->PROOF  named->NAMED
temporal->TEMPORAL  globally->GLOBALLY  eventually->EVENTUALLY  decreases->DECREASES
specification->SPECIFICATION  implementation->IMPLEMENTATION  global->GLOBAL
common->COMMON  ensure->ENSURE  require->REQUIRE  around->AROUND
parameter->PARAMETER  apply->APPLY  theorem->THEOREM  showgoals->SHOWGOALS
defergoal->DEFERGOAL  spoil->SPOIL  explicit->EXPLICIT  thunk->THUNK  isa->ISA
autoinstance->AUTOINSTANCE  constructor->CONSTRUCTOR  finite->FINITE
tactic->TACTIC  unfold->UNFOLD  forget->FORGET  debug->DEBUG  for->FOR
subclass->SUBCLASS  whenfirst->WHENFIRST  whenlast->WHENLAST  whenprev->WHENPREV
whennext->WHENNEXT  unprovable->UNPROVABLE  trigger->TRIGGER
```

Note: Some keywords share tokens: `template`->`MODULE`, `instance`->`INSTANTIATE`,
`execute`->`MIXIN`, `process`->`EXTRACT`.

#### Version-Based Reserved Word Removal (LexerVersion, line 253-304)
- v<=1.0: remove `state`, `local`
- v<=1.1: remove `returns mixin before after isolate with export delegate import include`
- v>1.1: remove `state set null match`
- v<=1.4: remove `function class object method execute destructor some maximizing minimizing private implement using property while invariant struct definition ghost alias trusted this var attribute scenario proof named fresh`
- v<=1.5: remove `variant of globally eventually temporal`
- v<=1.6: remove `decreases specification implementation require ensure around parameter apply theorem showgoals spoil explicit thunk isa autoinstance constructor tactic finite unfold forget`
- v<=1.7: remove `global common debug field for process subclass template whenfirst whenlast whennext whenprev unprovable trigger`
- v>1.7: remove `requires ensures`

#### Lexer Rules
```
t_TILDA    = r'\~'
t_COMMA    = r'\,'
t_PLUS     = r'\+'
t_TIMES    = r'\*'
t_DIV      = r'\/'
t_MINUS    = r'\-'
t_LT       = r'\<'
t_LE       = r'\<='
t_GT       = r'\>'
t_GE       = r'\>='
t_PTO      = r'\*>'
t_LPAREN   = r'\('
t_RPAREN   = r'\)'
t_OR       = r'\|'
t_AND      = r'\&'
t_EQ       = r'\='
t_TILDAEQ  = r'\~='
t_SEMI     = r'\;'
t_ASSIGN   = r'\:='
t_DOT      = r'\.'
t_LCB      = r'\{'
t_RCB      = r'\}'
t_ARROW    = r'\->'
t_IFF      = r'\<->'
t_COLON    = r':'
t_DOTS     = r'\.\.'
t_DOTDOTDOT = r'\.\.\.'
t_DOLLAR   = r'\$'
t_CARET    = r'\^'
t_LB       = r'\['
t_RB       = r'\]'

PRESYMBOL  : r'[_a-z0-9][_a-zA-Z0-9]*|".*?"'   (checks reserved)
VARIABLE   : r'[A-Z][_a-zA-Z0-9]*(\[[ab-zA-Z_0-9]*\])*'  (checks reserved)
NATIVEQUOTE: r'<<<[\s\S]*?>>>'
GLOBALLY   : '\u25A1'
EVENTUALLY : '\u25C7'
```

### 2. PRECEDENCE (ivy_parser.py)

#### v1.7+ (line 20-37)
```
('left', 'SEMI')
('left', 'GLOBALLY', 'EVENTUALLY')
('left', 'ARROW', 'IFF')
('left', 'OR')
('left', 'AND')
('left', 'TILDA')
('left', 'EQ','LE','LT','GE','GT','PTO')
('left', 'TILDAEQ')
('left', 'IF')
('left', 'ELSE')
('left', 'COLON')
('left', 'PLUS','MINUS')
('left', 'TIMES','DIV')
('left', 'DOLLAR')
('left', 'OLD')
('left', 'DOT')
```

#### v1.3-1.6 (line 40-56)
```
('left', 'SEMI')
('left', 'GLOBALLY', 'EVENTUALLY','WHENFIRST','WHENLAST','WHENNEXT','WHENPREV')
('left', 'IF')
('left', 'ELSE')
('left', 'OR')
('left', 'AND')
('left', 'TILDA')
('left', 'EQ','LE','LT','GE','GT','PTO')
('left', 'TILDAEQ')
('left', 'COLON')
('left', 'PLUS')
('left', 'MINUS')
('left', 'TIMES')
('left', 'DIV')
('left', 'DOLLAR')
```

#### v1.0-1.2 (line 63-76)
```
('left', 'SEMI')
('left', 'IF')
('left', 'ELSE')
('left', 'OR')
('left', 'AND')
('left', 'PLUS')
('left', 'TIMES')
('left', 'DIV')
('left', 'TILDA')
('left', 'EQ','LE','LT','GE','GT')
('left', 'TILDAEQ')
('left', 'COLON')
```

### 3. SYMBOL AND SUBSCRIPT RULES (ivy_logic_parser.py, lines 18-46)

```
SYMBOL     : PRESYMBOL                              # line 18
SYMBOL     : SYMBOL LB SYMsubscr RB                 # line 23
LABEL      : LB SYMBOL RB                           # line 28
SYMsubscr  : SYMBOL                                 # line 33
SYMsubscr  : THIS                                   # line 38
SYMsubscr  : SYMsubscr DOT SYMBOL                   # line 43
```

### 4. TYPE REFERENCES (ivy_logic_parser.py, lines 49-66)

```
atype      : SYMBOL                                 # line 49, all versions
```
**v1.3+:**
```
atype      : atype DOT SYMBOL                       # line 55
atype      : THIS                                   # line 62
```

### 5. TERM AND FORMULA RULES — v1.7+ (ivy_logic_parser.py)

These are the rules active for version > 1.6 (our target version):

```
# Application elements (v1.7+, line 104-115)
appelem    : SYMBOL                                  # line 104
appelem    : SYMBOL LPAREN terms RPAREN              # line 110

# Variables (unconditional, line 118-140)
var        : VARIABLE                                # line 118
var        : VARIABLE COLON atype                    # line 124
simplevar  : VARIABLE                                # line 130
simplevar  : VARIABLE COLON SYMBOL                   # line 136

# Term base forms (v1.7+, line 167-191)
term       : appelem                                 # line 167
term       : OLD appelem                             # line 172
term       : term DOT appelem                        # line 178

aterm      : appelem                                 # line 193
aterm      : aterm DOT appelem                       # line 198

# Term with sort annotation (v1.7+ only, line 205)
term       : term COLON atype                        # line 205

# Variable as term (unconditional, line 213)
term       : var                                     # line 213

# Arithmetic operators (v1.3+, line 220-242)
term       : term PLUS term                          # line 220
term       : term MINUS term                         # line 226
term       : term TIMES term                         # line 232
term       : term DIV term                           # line 238

# Conditional term (v1.3+, line 244)
term       : term IF fmla ELSE term                  # line 244

# Term lists (unconditional, line 273-287)
terms      : /* empty */                             # line 273
terms      : term                                    # line 278
terms      : terms COMMA term                        # line 283

# Parenthesized term (unconditional, line 290)
term       : LPAREN term RPAREN                      # line 290

# Variable lists (unconditional, line 296-316)
vars       : var                                     # line 296
vars       : vars COMMA var                          # line 301
simplevars : simplevar                               # line 307
simplevars : simplevars COMMA simplevar              # line 312

# Application (unconditional, line 320-336)
app        : SYMBOL                                  # line 320
app        : SYMBOL LPAREN terms RPAREN              # line 326
app        : term infix term                         # line 332
apps       : app                                     # line 339
apps       : apps COMMA app                          # line 344

# Atoms (unconditional, line 352-373)
atom       : SYMBOL                                  # line 352
atom       : SYMBOL LPAREN terms RPAREN              # line 358
atoms      : atom                                    # line 364
atoms      : atoms COMMA atom                        # line 369

# Literals (unconditional, line 377-399)
lit        : atom                                    # line 377
lit        : SYMBOL EQ SYMBOL                        # line 383
lit        : SYMBOL TILDAEQ SYMBOL                   # line 389
lit        : TILDA lit                               # line 395

# Relational operators (unconditional, line 401-429)
relop      : EQ                                      # line 401
relop      : LE                                      # line 406
relop      : LT                                      # line 411
relop      : GE                                      # line 416
relop      : GT                                      # line 421
relop      : PTO                                     # line 426

# Infix operators (unconditional, line 431-449)
infix      : PLUS                                    # line 431
infix      : MINUS                                   # line 436
infix      : TIMES                                   # line 441
infix      : DIV                                     # line 446

# Formula (unconditional, line 453)
fmla       : term                                    # line 453
```

#### v1.7+ FORMULAS (term-based, ivy_logic_parser.py line 584+)

In v1.7+, formulas are terms (no separate `fmla` productions for relops/booleans).
Instead, these are `term` rules:

```
term       : term EQ term                            # line 586
term       : term LE term                            # line 592
term       : term LT term                            # line 598
term       : term GE term                            # line 604
term       : term GT term                            # line 610
term       : term PTO term                           # line 616
term       : term TILDAEQ term                       # line 622
term       : TRUE                                    # line 628
term       : FALSE                                   # line 634
term       : TILDA term                              # line 640
term       : term AND term                           # line 646
term       : term OR term                            # line 656
term       : term ARROW term                         # line 666
term       : term IFF term                           # line 674
term       : FORALL simplevars DOT term %prec SEMI   # line 680
term       : EXISTS simplevars DOT term %prec SEMI   # line 686
term       : FORALL LPAREN vars RPAREN term          # line 692
term       : EXISTS LPAREN vars RPAREN term          # line 698
term       : GLOBALLY term                           # line 704
term       : EVENTUALLY term                         # line 710
term       : term WHENNEXT term                      # line 716
term       : term WHENPREV term                      # line 722
term       : term WHENFIRST term                     # line 728
term       : term WHENLAST term                      # line 734
```

#### v1.6 FORMULAS (fmla-based, ivy_logic_parser.py line 460+)

In v1.3-1.6, formulas are separate from terms:

```
fmla       : term relop term                         # line 462
fmla       : term TILDAEQ term                       # line 468
fmla       : LPAREN fmla RPAREN                      # line 474
fmla       : TRUE                                    # line 480
fmla       : FALSE                                   # line 486
fmla       : TILDA fmla                              # line 492
fmla       : fmla AND fmla                           # line 498
fmla       : fmla OR fmla                            # line 508
fmla       : fmla ARROW fmla                         # line 520 (v>1.0)
fmla       : fmla IFF fmla                           # line 526
fmla       : FORALL simplevars DOT fmla              # line 534 (v<=1.6)
fmla       : EXISTS simplevars DOT fmla              # line 540 (v<=1.6)
fmla       : GLOBALLY fmla                           # line 572
fmla       : EVENTUALLY fmla                         # line 578
```

#### NAMED BINDERS AND ISA (unconditional, ivy_logic_parser.py line 741+)

```
term       : LPAREN DOLLAR SYMBOL simplevars DOT fmla RPAREN LPAREN terms RPAREN  # line 741
term       : DOLLAR SYMBOL DOT fmla %prec SEMI       # line 749
term       : DOLLAR SYMBOL DOLLAR fmla %prec SEMI     # line 755
```

**v1.7+:**
```
term       : term ISA atype                           # line 762
```

### 6. TOP-LEVEL DECLARATIONS (ivy_parser.py)

```
# Start rule (line 380)
top        : /* empty */                             # line 380

# Include/Using (line 386-414)
top        : top USING SYMBOL                        # line 386
top        : top INCLUDE SYMBOL                      # line 396

# Labeled formulas (line 416-426)
labeledfmla : fmla                                   # line 416
labeledfmla : LABEL fmla                             # line 422

# Optional temporal/unprovable/explicit (line 428-462)
opttemporal : /* empty */                            # line 428
opttemporal : TEMPORAL                               # line 433
optunprovable : /* empty */                          # line 443
optunprovable : UNPROVABLE                           # line 448
optexplicit : /* empty */                            # line 555
optexplicit : EXPLICIT                               # line 560
```

#### AXIOM, PROPERTY, CONJECTURE

**v1.7+ (line 493-519):**
```
gprop      : fmla                                    # line 494
gprop      : schdefnrhs                              # line 499
lgprop     : optlabel gprop                          # line 504
top        : top optexplicit opttemporal AXIOM lgprop  # line 511
```

**v1.6 (line 485-492):**
```
top        : top opttemporal AXIOM labeledfmla       # line 485
```

**Property (unconditional, line 532-545):**
```
top        : top optexplicit opttemporal PROPERTY labeledfmla optskolem optproof  # line 532
```

**Conjecture (unconditional, line 547-553):**
```
top        : top CONJECTURE labeledfmla              # line 547
```

**Invariant (v1.7+, line 569-596):**
```
top        : top optexplicit INVARIANT labeledfmla optproof         # line 569
top        : top UNPROVABLE INVARIANT labeledfmla optproof          # line 584
```

#### MODULE, OBJECT, CLASS

```
# Module infrastructure (line 598-632)
modulestart : /* empty */                            # line 598
moduleend   : /* empty */                            # line 604
modcat      : /* empty */                            # line 609
modcat      : OBJECT                                 # line 614
modcat      : ISOLATE                                # line 619
opteq       : /* empty */                            # line 624
opteq       : EQ                                     # line 629

# Module definition (line 634)
top        : top MODULE modulestart modcat atom optwith EQ LCB top RCB moduleend  # line 634

# Object/class continuation (line 654-664)
optdotdotdot : /* empty */                           # line 654
optdotdotdot : DOTDOTDOT                             # line 661

# Object args (line 666-676)
objectargs : optargs                                 # line 666  (sets stack[-1].params = p[0])
objectend  : /* empty */                             # line 672

# Object symbol (line 695-700)
objsym     : SYMBOL                                  # line 695  (sets parent_object = p[0])

# Object (line 702)
top        : top OBJECT objsym objectargs EQ LCB optdotdotdot top RCB objectend  # line 702

# Class (line 708)
top        : top CLASS objsym objectargs EQ LCB optdotdotdot top RCB objectend   # line 708

# Subclass (v1.8+, line 720)
top        : top SUBCLASS objsym OF atype EQ LCB optdotdotdot top RCB objectend  # line 720

# Optional semicolon (line 734-742)
optsemi    : /* empty */                             # line 734
optsemi    : SEMI                                    # line 739

# Macro (line 744)
top        : top MACRO atom EQ sequence              # line 744
```

#### SCHEMA, THEOREM, PROOF

```
# Schema definition (line 751-874)
schdefnrhs : fmla                                   # line 751
schdecl    : FUNCTION funs                           # line 756
schdecl    : FRESH FUNCTION funs                     # line 761
schdecl    : INDIV funs                              # line 766
schdecl    : FRESH INDIV funs                        # line 771
schdecl    : RELATION rels                           # line 776
schdecl    : FRESH RELATION rels                     # line 781
schdecl    : TYPE SYMBOL                             # line 786
schdecl    : schdefnrhs                              # line 840
schdecls   : /* empty */                             # line 847
schdecls   : schdecls schdecl                        # line 852
schdefnrhs : LCB schdecls schconc RCB               # line 858
schdefn    : defnlhs EQ schdefnrhs                   # line 864
schconc    : DEFINITION defn                         # line 817
```

**v1.7+ (line 804-815):**
```
schdecl    : optexplicit PROPERTY lgprop             # line 804
schdecl    : THEOREM lgprop                          # line 812
schconc    : optexplicit PROPERTY lgprop             # line 831
```

**v1.6 (line 797-827):**
```
schdecl    : PROPERTY labeledfmla                    # line 797
schconc    : PROPERTY fmla                           # line 824
```

```
top        : top SCHEMA schdefn                      # line 870
top        : top THEOREM schdefn optproof            # line 876
top        : top THEOREM LABEL schdefnrhs optproof   # line 884
top        : top PROOF LABEL proofstep               # line 896
```

#### OPTSKOLEM (line 521-529)

```
optskolem  : /* empty */                             # line 521
optskolem  : NAMED defnlhs                           # line 526
```

#### INSTANTIATE (line 904-914)

```
top        : top INSTANTIATE insts                   # line 904
top        : top AUTOINSTANCE insts                  # line 910
insts      : inst                                    # line 916
insts      : insts COMMA inst                        # line 921
```

#### PNAME, MODINST, INST (line 927-1005)

```
pname      : atype                                   # line 927
pname      : var                                     # line 933
pname      : infix                                   # line 938
pname      : relop                                   # line 944
pname      : THIS                                    # line 950
pname      : TRUE                                    # line 956
pname      : FALSE                                   # line 962
pnames     : /* empty */                              # line 968
pnames     : pname                                    # line 973
pnames     : pnames COMMA pname                       # line 978
modinst    : dotsym                                   # line 984
modinst    : dotsym LPAREN pnames RPAREN              # line 989
inst       : modinst                                  # line 995
inst       : modinst COLON modinst                    # line 1001
```

#### SYMBOL DECLARATIONS (line 1007-1096)

```
top        : top symdecl                             # line 1007
symdecl    : constantdecl                            # line 1013
symdecl    : DESTRUCTOR tterms                       # line 1018
symdecl    : FIELD tterms                            # line 1024
constantdecl : INDIV tterms                          # line 1047
constantdecl : VAR tterms                            # line 1053
constantdecl : PARAMETER parameter                   # line 1092
```

**v1.7+:**
```
symdecl    : CONSTRUCTOR tterms                      # line 1036
```

#### PARAMETER (line 1059-1091)

```
parameter  : tterm                                   # line 1059
paramval   : TRUE                                    # line 1066
paramval   : FALSE                                   # line 1072
paramval   : SYMBOL                                  # line 1078
parameter  : tterm EQ paramval                       # line 1084
```

#### RELATION, FUNCTION, DERIVED (line 1097-1197)

```
rel        : defnlhs                                 # line 1097
rel        : defn                                    # line 1103
rels       : rel                                     # line 1108
rels       : rels COMMA rel                          # line 1113
top        : top RELATION rels                       # line 1119

tatoms     : tatom                                   # line 1126
tatoms     : tatoms COMMA tatom                      # line 1131
tatom      : SYMBOL                                  # line 1137
tatom      : SYMBOL targs                            # line 1143
tatom      : LPAREN var relop var RPAREN             # line 1149

fun        : typeddefn                               # line 1155
fun        : typeddefn EQ defnrhs                    # line 1162
funs       : fun                                     # line 1169
funs       : funs COMMA fun                          # line 1174
top        : top FUNCTION funs                       # line 1180
top        : top DERIVED defns                       # line 1193
```

#### PROOF STEPS (line 1199-1498)

**v1.7+ (line 1207-1498):**
```
proofstep  : APPLY atype optrenaming                 # line 1208
proofstep  : ASSUME atype optrenaming                # line 1216
proofstep  : INSTANTIATE atype optrenaming           # line 1225
proofstep  : INSTANTIATE LABEL atype optrenaming     # line 1234
proofstep  : SHOWGOALS                               # line 1245
proofstep  : DEFERGOAL                               # line 1251
proofstep  : SPOIL atype                             # line 1257
proofstep  : TACTIC SYMBOL opttacticwith optproofgroup  # line 1265
proofstep  : opttemporal PROPERTY labeledfmla optskolem optproofgroup  # line 1274
proofstep  : FUNCTION funs                           # line 1284
proofstep  : THEOREM lgprop optproofgroup            # line 1290
proofstep  : PROOF LABEL proofgroup                  # line 1298
```

#### MATCH (line 1306-1326)

```
match      : defn                                    # line 1306
match      : var EQ fmla                             # line 1311
matches    : match                                   # line 1317
matches    : matches COMMA match                     # line 1322
```

#### RENAMING (v1.7+, line 1337-1497)

```
renamingitem : VARIABLE DIV VARIABLE                 # line 1338
renamingitem : SYMBOL DIV SYMBOL                     # line 1344
renaminglist : renamingitem                           # line 1350
renaminglist : renaminglist COMMA renamingitem        # line 1355
optrenaming  : /* empty */                            # line 1361
renaming     : LT renaminglist GT                     # line 1366
optrenaming  : renaming                               # line 1372

proofstep  : APPLY atype optrenaming WITH matches    # line 1377
proofstep  : ASSUME atype optrenaming WITH matches   # line 1385
proofstep  : INSTANTIATE atype optrenaming WITH matches  # line 1394
proofstep  : INSTANTIATE LABEL atype optrenaming WITH matches  # line 1403
renamings  : /* empty */                             # line 1414
renamings  : renamings renaming                      # line 1419
unfspec    : callatom renamings                      # line 1425
unfspecs   : unfspec                                 # line 1431
unfspecs   : unfspecs COMMA unfspec                  # line 1436
proofstep  : UNFOLD atype WITH unfspecs              # line 1442
proofstep  : UNFOLD WITH unfspecs                    # line 1451
proofstep  : FORGET callatoms                        # line 1458
pflet      : var EQ fmla                             # line 1464
pflets     : pflet                                   # line 1470
pflets     : pflets COMMA pflet                      # line 1475
proofstep  : LET pflets                              # line 1481
proofstep  : INSTANTIATE WITH pflets                 # line 1487
proofstep  : IF fmla proofgroup ELSE proofgroup      # line 1493

opttacticwith : /* empty */                          # line 1499
opttacticwith : WITH tacticwithlistchoice            # line 1504
tacticwithlistchoice : tacticwithlist                # line 1510
tacticwithlistchoice : pflets                        # line 1515
tacticwithelem : INVARIANT labeledfmla               # line 1520
tacticwithelem : DEFINITION typeddefn EQ fmla        # line 1525
tacticwithelem : TRIGGER atype WITH terms            # line 1532
tacticwithlist : tacticwithelem                      # line 1538
tacticwithlist : tacticwithlist tacticwithelem        # line 1543
opttacticwith : WITH LCB tacticwithlist RCB          # line 1549
```

**v1.6 (line 1328-1335):**
```
proofstep  : SYMBOL                                  # line 1200
proofstep  : SYMBOL WITH matches                     # line 1329
```

#### PROOF SEQUENCES (unconditional, line 1555-1608)

```
proofseq   : proofstep                               # line 1555
proofseq   : proofseq optsemi proofstep              # line 1560
proofstep  : proofgroup                              # line 1566
proofgroup : LCB proofseq RCB                        # line 1571
proofgroup : LCB RCB                                 # line 1576
optproof   : /* empty */                             # line 1582
optproof   : PROOF proofstep                         # line 1587
optproof   : PROOF LABEL proofstep                   # line 1592
optproofgroup : /* empty */                          # line 1600
optproofgroup : PROOF proofgroup                     # line 1605
```

#### DEFINITION (line 1610-1655)

**v1.7+ (line 1618-1654):**
```
optlabel   : LABEL                                   # line 1620
optlabel   : /* empty */                             # line 1626
gdefn      : defn                                    # line 1631
gdefn      : LCB defn RCB                            # line 1636
top        : top optexplicit DEFINITION optlabel gdefn optproof  # line 1642
```

**v1.6 (line 1611-1617):**
```
top        : top DEFINITION defns optproof           # line 1611
```

#### PROGRESS, RELY, CONCEPT, UPDATE (line 1657-1706)

```
top        : top PROGRESS defns                      # line 1657
top        : top RELY atom ARROW atom                # line 1663
top        : top MIXORD callatom ARROW callatom      # line 1669
top        : top RELY atom                           # line 1675
top        : top CONCEPT cdefns                      # line 1681
```

**v1.6:**
```
top        : top INIT labeledfmla                    # line 1689
```

```
top        : top UPDATE apps FROM apps upaxes        # line 1698
```

#### TYPE DECLARATIONS (line 1708-1924)

```
optfinite  : /* empty */                             # line 1708
optfinite  : FINITE                                  # line 1713
optghost   : /* empty */                             # line 1718
optghost   : GHOST                                   # line 1723
typesymbol : SYMBOL                                  # line 1728
typesymbol : THIS                                    # line 1733
top        : top optfinite optghost TYPE typesymbol  # line 1739
top        : top optfinite optghost TYPE typesymbol EQ sort  # line 1751

tsyms      : var                                     # line 1771
tsyms      : tsyms COMMA var                         # line 1776
targs      : LPAREN RPAREN                           # line 1782
targs      : LPAREN tsyms RPAREN                     # line 1787

param      : SYMBOL COLON SYMBOL                     # line 1799
params     : param                                   # line 1807
params     : params COMMA param                      # line 1812
optargs    : /* empty */                             # line 1818
optargs    : LPAREN lparams RPAREN                   # line 1823
optreturns : /* empty */                             # line 1828
optreturns : RETURNS LPAREN lparams RPAREN           # line 1833
optactualreturns : /* empty */                       # line 1838
optactualreturns : callatoms ASSIGN                  # line 1843
```

#### TAPP, TTERM, TTERMS (line 1848-1887)

```
tapp       : SYMBOL                                  # line 1848
tapp       : SYMBOL targs                            # line 1854
tapp       : LPAREN var infix var RPAREN             # line 1860
tterm      : tapp                                    # line 1867
tterm      : tapp COLON atype                        # line 1872
tterms     : tterm                                   # line 1878
tterms     : tterms COMMA tterm                      # line 1883
```

#### SORT (line 1889-1913)

```
sort       : LCB SYMBOL RCB                          # line 1889
sort       : LCB SYMBOL COMMA names RCB              # line 1894
sort       : LCB SYMBOL DOTS SYMBOL RCB              # line 1899
sort       : STRUCT LCB tterms RCB                   # line 1905
sort       : STRUCT LCB RCB                          # line 1910
names      : SYMBOL                                  # line 1915
names      : names COMMA SYMBOL                      # line 1920
```

#### UPDATE AXES (line 1926-1977)

```
upaxes     : /* empty */                             # line 1926
upaxes     : upaxes upax                             # line 1931
upax       : PARAMS tterms IN action ARROW requires ensures  # line 1938
requires   : /* empty */                             # line 1949
requires   : REQUIRES fmla                           # line 1954
modifies   : /* empty */                             # line 1959
modifies   : MODIFIES LCB RCB                        # line 1964
modifies   : MODIFIES TIMES                          # line 1969
modifies   : MODIFIES atoms                          # line 1974
ensures    : ENSURES fmla                            # line 1983
```

#### ACTION DECLARATIONS (line 1988-2303)

**v<=1.1 (line 1989-1993):**
```
top        : top ACTION SYMBOL loc EQ sequence loc   # line 1989
```

**v1.2+ (line 1994-2080):**
```
optactiondef : /* empty */                           # line 1996
topseq     : sequence                                # line 2001
topseq     : LCB NATIVEQUOTE RCB                     # line 2006
optactiondef : EQ topseq                             # line 2013
optactiondef : EQ TIMES                              # line 2018
optimpex   : /* empty */                             # line 2024
optimpex   : EXPORT                                  # line 2029
optimpex   : IMPORT                                  # line 2034
actmeth    : ACTION                                  # line 2039
actmeth    : METHOD                                  # line 2044
top        : top optimpex actmeth SYMBOL optargs optreturns optactiondef  # line 2049
```

#### MIXIN, BEFORE, AFTER, IMPLEMENT (v1.2+, line 2117-2303)

```
top        : top MIXIN callatom BEFORE callatom      # line 2118
top        : top MIXIN callatom AFTER callatom       # line 2123
top        : top BEFORE atype optargs optreturns sequence  # line 2128
top        : top AFTER atype optargs optreturns topseq   # line 2135
top        : top AFTER INIT optargs topseq           # line 2164
top        : top IMPLEMENT atype optargs optreturns topseq  # line 2171
top        : top IMPLEMENT TYPE SYMBOL WITH SYMBOL   # line 2178
```

**v1.7+:**
```
top        : top AROUND atype optargs optreturns LCB actseq optsemi DOTDOTDOT actseq optsemi RCB  # line 2153
```

#### TRUSTED, ISOLATE, EXTRACT (v1.2+, line 2190-2261)

```
opttrusted : /* empty */                             # line 2190
opttrusted : TRUSTED                                 # line 2194
top        : top opttrusted ISOLATE SYMBOL optargs EQ callatoms  # line 2198
top        : top opttrusted ISOLATE SYMBOL optargs EQ callatoms WITH callatoms  # line 2208
optwith    : /* empty */                             # line 2218
optwith    : WITH callatoms                          # line 2222
top        : top opttrusted ISOLATE SYMBOL optargs EQ LCB top RCB optwith  # line 2226
top        : top EXTRACT objsym objectargs EQ LCB top RCB optwith  # line 2239
top        : top EXTRACT objsym objectargs EQ callatoms  # line 2250
```

#### EXPORT, IMPORT, DELEGATE (v1.2+, line 2262-2303)

```
top        : top EXPORT callatom                     # line 2262
top        : top IMPORT callatom                     # line 2270
optdelegee : /* empty */                             # line 2286
optdelegee : ARROW callatom                          # line 2290
top        : top DELEGATE callatoms optdelegee       # line 2294
```

**v1.6:**
```
top        : top PRIVATE callatom                    # line 2279
```

#### SPECIFICATION/IMPLEMENTATION/PRIVATE/GLOBAL/COMMON (v1.7+, line 2305-2352)

```
specimpl   : SPECIFICATION                           # line 2307
specimpl   : IMPLEMENTATION                          # line 2314
specimpl   : PRIVATE                                 # line 2321
specimpl   : GLOBAL                                  # line 2328
specimpl   : COMMON                                  # line 2335
top        : top specimpl LCB top RCB                # line 2342
```

#### ALIAS, STATE, ASSERT (line 2361-2392)

```
top        : top ALIAS SYMBOL EQ callatom            # line 2361
top        : top STATE SYMBOL EQ state_expr          # line 2369
assert_rhs : LCB requires modifies ensures RCB      # line 2375
assert_rhs : fmla                                   # line 2380
```

**v1.6:**
```
top        : top ASSERT SYMBOL ARROW assert_rhs      # line 2386
```

#### INTERPRET (line 2394-2455)

```
oper       : atype                                   # line 2394
oper       : relop                                   # line 2399
oper       : infix                                   # line 2404
oper       : NATIVEQUOTE                             # line 2409
top        : top INTERPRET oper ARROW oper           # line 2416
top        : top INTERPRET oper ARROW LCB term DOTS term RCB  # line 2426
moresymbols : /* empty */                            # line 2436
moresymbols : moresymbols COMMA SYMBOL              # line 2441
top        : top INTERPRET oper ARROW LCB SYMBOL moresymbols RCB  # line 2447
```

#### NATIVEQUOTE (line 2475-2484)

```
top        : top NATIVEQUOTE                         # line 2475
```

#### ATTRIBUTE (line 2486-2511)

```
attributeval : callatom                              # line 2487
attributeval : TRUE                                  # line 2492
attributeval : FALSE                                 # line 2498
top        : top ATTRIBUTE callatom EQ attributeval  # line 2503
```

#### VARIANT (line 2513-2535)

```
top        : top VARIANT typesymbol OF atype         # line 2513
top        : top VARIANT typesymbol OF atype EQ sort # line 2525
```

#### SCENARIO (line 2537-2614)

```
places     : SYMBOL                                  # line 2537
places     : places COMMA SYMBOL                     # line 2543
sceninit   : ARROW places                            # line 2550
scenariomixin : BEFORE atype optargs optreturns sequence  # line 2564
scenariomixin : AFTER atype optargs optreturns sequence   # line 2575
scentranss : /* empty */                             # line 2587
scentranss : scentranss places ARROW places COLON scenariomixin  # line 2592
scentranss : scentranss places COLON scenariomixin   # line 2600
top        : top SCENARIO LCB sceninit SEMI scentranss RCB  # line 2608
```

#### LOCATION (line 2616-2624)

```
loc        : /* empty */                             # line 2616
loc        : SYMBOL                                  # line 2621
```

### 7. ACTION SEQUENCES (ivy_parser.py, line 2626-2719)

```
actseqrev  : simpleact                               # line 2626
actseqrev  : complexact                              # line 2631
actseqrev  : simpleact SEMI actseqrev                # line 2636
actseqrev  : simpleact SEMI                          # line 2642
actseqrev  : complexact actseqrev                    # line 2647
actseqrev  : complexact SEMI actseqrev               # line 2653
actseqrev  : complexact SEMI                         # line 2659
actseq     : actseqrev                               # line 2664

sequence   : LCB RCB                                 # line 2699
sequence   : LCB actseq RCB                          # line 2705
sequence   : LCB actseq SEMI RCB                     # line 2715

complexact : sequence                                # line 2721
action     : simpleact                               # line 2726
action     : complexact                              # line 2731
```

### 8. SIMPLE ACTIONS (ivy_parser.py, line 2736-2856)

```
simpleact  : ASSUME labeledfmla                      # line 2737
simpleact  : SET lit                                 # line 2821
simpleact  : term ASSIGN fmla                        # line 2827
simpleact  : termtuple ASSIGN callatom               # line 2839
simpleact  : term ASSIGN TIMES                       # line 2845
simpleact  : term                                    # line 2851
simpleact  : INSTANTIATE callatom                    # line 3048
simpleact  : CALL optactualreturns callatom          # line 3102
simpleact  : CALL callatom                           # line 3108
```

**v1.7+:**
```
simpleact  : optunprovable ASSERT labeledfmla        # line 2756
simpleact  : optunprovable ASSERT labeledfmla PROOF proofstep  # line 2765
simpleact  : optunprovable ENSURE labeledfmla        # line 2774
simpleact  : optunprovable ENSURE labeledfmla PROOF proofstep  # line 2783
simpleact  : optunprovable REQUIRE labeledfmla       # line 2792
simpleact  : optunprovable REQUIRE labeledfmla PROOF proofstep  # line 2801
```

**v1.6:**
```
simpleact  : ASSERT labeledfmla                      # line 2744
simpleact  : ENSURES labeledfmla                     # line 2750
```

**v1.2 field assignment:**
```
simpleact  : term DOT SYMBOL ASSIGN term             # line 3023
simpleact  : term DOT SYMBOL ASSIGN NULL             # line 3030
simpleact  : term DOT SYMBOL ASSIGN term DOT SYMBOL  # line 3036
simpleact  : term DOT SYMBOL ASSIGN FALSE            # line 3042
```

```
termtuple  : LPAREN term COMMA terms RPAREN          # line 2833
```

### 9. COMPLEX ACTIONS (ivy_parser.py, line 2858-3019)

**v<=1.4:**
```
complexact : IF fmla sequence                        # line 2860
complexact : IF fmla sequence ELSE action            # line 2866
```

**v1.5+ (line 2872-2993):**
```
somefmla   : fmla                                    # line 2874
somefmla   : fmla ASSIGN fmla                        # line 2879
bounds     : params DOT                              # line 2893
bounds     : LPAREN lparams RPAREN                   # line 2898
somefmla   : SOME bounds fmla                        # line 2903
somefmla   : SOME bounds fmla MINIMIZING term        # line 2912
somefmla   : SOME bounds fmla MAXIMIZING term        # line 2922

complexact : IF somefmla sequence                    # line 2940
complexact : IF somefmla sequence ELSE action        # line 2947

invariants : /* empty */                             # line 2954
invariants : invariants INVARIANT labeledfmla        # line 2959
invariants : invariants INVARIANT labeledfmla PROOF proofstep  # line 2968
decreases  : DECREASES fmla                          # line 2977
decreases  : /* empty */                             # line 2984

complexact : WHILE somefmla invariants decreases sequence  # line 2989

complexact : FOR tterm COMMA tterm IN fmla invariants decreases sequence  # line 3001
```

**Unconditional:**
```
complexact : IF TIMES sequence ELSE action           # line 3015
complexact : LOCAL lparams sequence                  # line 3148
complexact : LET eqns sequence                       # line 3257
```

**v1.6+:**
```
simpleact  : VAR tterm optinit                       # line 3183
optinit    : /* empty */                             # line 3173
optinit    : ASSIGN fmla                             # line 3178
opttypedsym : SYMBOL                                 # line 3159
opttypedsym : SYMBOL COLON atype                     # line 3166
```

**v1.7+:**
```
complexact : THUNK LABEL SYMBOL optargs COLON atype ASSIGN sequence  # line 3190
simpleact  : DEBUG SYMBOL optdebugargs               # line 3230
debugarg   : SYMBOL EQ fmla                          # line 3201
debugargs  : debugarg                                # line 3209
debugargs  : debugargs COMMA debugarg                # line 3214
optdebugargs : /* empty */                           # line 3220
optdebugargs : WITH debugargs                        # line 3225
```

### 10. CALLATOM (line 3055-3100)

```
callatom   : atom                                    # line 3055
callatoms  : callatom                                # line 3091
callatoms  : callatoms COMMA callatom                # line 3096
```

**v1.6+:**
```
callatom   : THIS                                    # line 3061
callatom   : METHOD                                  # line 3067
```

**v<=1.2:**
```
callatom   : callatom COLON callatom                 # line 3076
```

**v1.3+:**
```
callatom   : callatom DOT callatom                   # line 3084
```

### 11. LPARAM (line 3120-3145)

```
lparam     : SYMBOL COLON atype                      # line 3120
lparams    : lparam                                  # line 3136
lparams    : lparams COMMA lparam                    # line 3141
```

**v1.7+:**
```
lparam     : CARET SYMBOL COLON atype                # line 3129
```

### 12. EQUATIONS (line 3241-3271)

```
eqn        : SYMBOL EQ SYMBOL                        # line 3241
eqns       : eqn                                     # line 3246
eqns       : eqns COMMA eqn                          # line 3251
symbols    : SYMBOL                                  # line 3262
symbols    : symbols COMMA SYMBOL                    # line 3267
```

### 13. CONCEPT DEFINITIONS (line 3273-3288)

```
cdefns     : cdefn                                   # line 3273
cdefns     : cdefns COMMA cdefn                      # line 3278
cdefn      : atom EQ expr                            # line 3284
```

### 14. DEFINITIONS (line 3290-3395)

```
defns      : defn                                    # line 3290
defns      : defns COMMA defn                        # line 3295
dotsym     : SYMBOL                                  # line 3301
dotsym     : dotsym DOT SYMBOL                       # line 3307
defnlhs    : dotsym                                  # line 3313
defnlhs    : dotsym LPAREN defargs RPAREN            # line 3318
defargs    : defarg                                  # line 3324
defargs    : defargs COMMA defarg                    # line 3329
defarg     : lparam                                  # line 3335
defarg     : var                                     # line 3340
defnlhs    : LPAREN defarg relop defarg RPAREN       # line 3345
defnlhs    : LPAREN defarg infix defarg RPAREN       # line 3351
typeddefn  : defnlhs                                 # line 3357
typeddefn  : defnlhs COLON atype                     # line 3362
defnrhs    : fmla                                    # line 3368
defnrhs    : somevarfmla                             # line 3373
defnrhs    : NATIVEQUOTE                             # line 3378
defn       : typeddefn EQ defnrhs                    # line 3385
```

### 15. SOME EXPRESSION (line 3397-3421)

```
optin      : /* empty */                             # line 3397
optin      : IN fmla                                 # line 3402
optelse    : /* empty */                             # line 3407
optelse    : ELSE fmla                               # line 3412
somevarfmla : SOME simplevar DOT fmla optin optelse  # line 3417
```

### 16. CONCEPT SPACE EXPRESSIONS (line 3423-3499)

```
expr       : LCB fmla RCB                            # line 3423
exprterm   : aterm                                   # line 3428
exprterm   : var                                     # line 3433
expr       : exprterm                                # line 3438
expr       : exprterm relop exprterm                  # line 3443
expr       : exprterm TILDAEQ exprterm               # line 3449
expr       : TILDA expr                              # line 3455
expr       : LPAREN expr RPAREN                      # line 3464
expr       : prod                                    # line 3469
expr       : sum                                     # line 3474
prod       : expr TIMES expr                         # line 3479
prod       : prod TIMES expr                         # line 3484
sum        : expr PLUS expr                          # line 3490
sum        : sum PLUS expr                           # line 3495
```

### 17. STATE EXPRESSIONS (line 3501-3538)

```
state_expr : TRUE                                    # line 3501
state_expr : FALSE                                   # line 3506
state_expr : SYMBOL                                  # line 3511
state_expr : SYMBOL LPAREN state_expr RPAREN         # line 3516
state_expr : state_expr OR state_expr                # line 3521
state_expr : LCB requires modifies ensures RCB      # line 3530
state_expr : ENTRY                                   # line 3535
```

### 18. PARSER CONSTRUCTION (line 3540-3555)

```python
from .ivy_logic_parser import *   # line 3540: merges all logic rules

parser = yacc.yacc(start='top', tabmodule='ivy_parsetab', ...)  # line 3555
```

### 19. expand_autoinstances (line 3574-3607)

Called by `parse()` for non-nested parses. Expands `AutoInstanceDecl` nodes:
- Collects autoinstance declarations (keyed by prefix+param count)
- For each other declaration, collects type names via `get_type_names()`
- For matching type references, creates new `Instantiation` and calls `do_insts()`

### 20. HELPER FUNCTIONS

```python
create_object(top, name, objectargs, module, lineno, continuation)  # line 678
  # prefargs = [Variable('V'+str(idx), pr.sort) for idx,pr in enumerate(objectargs)]
  # pref = Atom(name, prefargs)
  # vsubst = dict((pr.rep,v) for pr,v in zip(objectargs,prefargs))
  # inst_mod(top, module, pref, {}, vsubst)
  # stack.pop()

inst_mod(ivy, module, pref, subst, vsubst, modname, lineno)  # line 135
do_insts(ivy, insts)  # line 203
parse_nativequote(p, n)  # line 2457
fix_if_part(cond, part)  # line 2932
check_non_temporal(x)  # line 239
```

### 21. SECONDARY PARSERS (not used by ivy_check)

#### ivy_concept_space.py (line 122-186)
Separate parser for concept space expressions. Tokens: SYMBOL, COMMA, LPAREN,
RPAREN, LBR, RBR, PLUS, TIMES, TILDA.

#### ivy_ev_parser.py (line 290-419)
Separate parser for event parsing. Tokens: SYMBOL, COMMA, LPAREN, RPAREN, LBR,
RBR, LCB, RCB, SEMI, COLON, GT, LT.

#### ivy_dafny_grammar.py (line 21-327)
Separate parser for Dafny code (60+ rules). Not connected to main parser.

---

## Verification

After creating IVY_PYTHON_GRAMMAR.md, cross-check against the Go grammar:

```bash
cd ~/goivy
# Count grammar rules in Go vs this document
grep -c "^[a-z_]" IVY_PYTHON_GRAMMAR.md
grep -c '^[a-z_]' lalr_full/grammar_v17.y
```

Then systematically check each production in the document against `grammar_v17.y`
to find any missing rules in the Go port.
