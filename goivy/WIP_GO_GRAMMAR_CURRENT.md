# Complete Go Ivy Grammar Reference (Current WIP)

**Source:** `~/goivy/parser/grammar_v17.y` (4,296 lines)
**Lexer:** `~/goivy/lexer/token.go`, `~/goivy/lexer/lexer.go`
**Date extracted:** 2026-03-23
**Nonterminals:** 138 rules
**Start symbol:** `top`

---

## 1. TOKENS (grammar_v17.y lines 282-348)

### Identifiers and literals
```
%token <str>  TOK_PRESYMBOL TOK_VARIABLE          # line 282
%token <str>  TOK_LABEL                             # line 283
%token <str>  TOK_NATIVEQUOTE                       # line 284
```

### Punctuation
```
%token  TOK_LPAREN TOK_RPAREN TOK_LB TOK_RB TOK_LCB TOK_RCB   # line 287
%token  TOK_COMMA TOK_SEMI TOK_COLON TOK_DOT                    # line 288
%token  TOK_DOTS TOK_DOTDOTDOT                                   # line 289
```

### Operators
```
%token  TOK_PLUS TOK_MINUS TOK_TIMES TOK_DIV                    # line 292
%token  TOK_EQ TOK_TILDAEQ TOK_TILDA TOK_LE TOK_LT TOK_GE TOK_GT  # line 293
%token  TOK_AND TOK_OR TOK_ARROW TOK_IFF                        # line 294
%token  TOK_PTO TOK_DOLLAR TOK_CARET                             # line 295
%token  TOK_ASSIGN                                               # line 296
```

### Logic keywords
```
%token  TOK_FORALL TOK_EXISTS                                    # line 299
%token  TOK_TRUE TOK_FALSE                                       # line 300
%token  TOK_OLD TOK_THIS TOK_ISA                                 # line 301
%token  TOK_IF TOK_ELSE                                          # line 302
%token  TOK_GLOBALLY TOK_EVENTUALLY                              # line 303
%token  TOK_WHENNEXT TOK_WHENPREV TOK_WHENFIRST TOK_WHENLAST    # line 304
```

### Action keywords
```
%token  TOK_ASSUME TOK_ASSERT TOK_REQUIRE TOK_ENSURE            # line 307
%token  TOK_VAR TOK_LOCAL TOK_LET TOK_CALL                      # line 308
%token  TOK_WHILE TOK_FOR TOK_IN TOK_INVARIANT TOK_DECREASES    # line 309
%token  TOK_RETURNS                                              # line 310
%token  TOK_SOME TOK_MINIMIZING TOK_MAXIMIZING                  # line 311
%token  TOK_DEBUG TOK_THUNK TOK_UNPROVABLE TOK_PROOF            # line 312
%token  TOK_INSTANTIATE                                          # line 313
```

### Declaration keywords
```
%token  TOK_RELATION TOK_INDIV TOK_FUNCTION TOK_DERIVED         # line 316
%token  TOK_AXIOM TOK_CONJECTURE TOK_SCHEMA TOK_THEOREM         # line 317
%token  TOK_PROPERTY TOK_DEFINITION                              # line 318
%token  TOK_TYPE TOK_STRUCT                                      # line 319
%token  TOK_MODULE TOK_OBJECT TOK_CLASS TOK_SUBCLASS             # line 320
%token  TOK_ACTION TOK_METHOD                                    # line 321
%token  TOK_BEFORE TOK_AFTER TOK_AROUND TOK_MIXIN TOK_IMPLEMENT # line 322
%token  TOK_ISOLATE TOK_EXTRACT TOK_TRUSTED                     # line 323
%token  TOK_EXPORT TOK_IMPORT TOK_DELEGATE TOK_USING TOK_INCLUDE # line 324
%token  TOK_INTERPRET TOK_MACRO TOK_ALIAS TOK_ATTRIBUTE          # line 325
%token  TOK_VARIANT TOK_OF                                       # line 326
%token  TOK_SCENARIO                                             # line 327
%token  TOK_PROGRESS TOK_RELY TOK_MIXORD                         # line 328
%token  TOK_CONCEPT TOK_STATE TOK_UPDATE TOK_FROM                # line 329
%token  TOK_PARAMS TOK_MODIFIES TOK_ENSURES TOK_REQUIRES         # line 330
%token  TOK_INIT TOK_ENTRY TOK_SET TOK_NULL TOK_MATCH            # line 331
%token  TOK_FRESH TOK_NAMED                                      # line 332
%token  TOK_TEMPORAL TOK_EXPLICIT                                # line 333
%token  TOK_SPECIFICATION TOK_IMPLEMENTATION TOK_PRIVATE         # line 334
%token  TOK_GLOBAL TOK_COMMON                                    # line 335
%token  TOK_GHOST TOK_FINITE                                     # line 336
%token  TOK_PARAMETER                                            # line 337
%token  TOK_DESTRUCTOR TOK_CONSTRUCTOR TOK_FIELD                 # line 338
%token  TOK_AUTOINSTANCE                                         # line 339
%token  TOK_VAR_KW                                               # line 340 (placeholder)
```

### Proof/tactic keywords
```
%token  TOK_TACTIC TOK_TRIGGER                                   # line 343
%token  TOK_SHOWGOALS TOK_DEFERGOAL TOK_SPOIL                   # line 344
%token  TOK_UNFOLD TOK_FORGET                                    # line 345
%token  TOK_APPLY                                                # line 346
%token  TOK_WITH                                                 # line 347
%token  TOK_METHOD_KW TOK_NULL_KW TOK_SET_KW                    # line 348 (placeholders)
```

## 2. PRECEDENCE (lines 477-493)

```
%left   TOK_SEMI                                                 # line 477
%left   TOK_GLOBALLY TOK_EVENTUALLY                              # line 478
%left   TOK_ARROW TOK_IFF                                        # line 479
%left   TOK_OR                                                   # line 480
%left   TOK_AND                                                  # line 481
%left   TOK_TILDA                                                # line 482
%left   TOK_EQ TOK_LE TOK_LT TOK_GE TOK_GT TOK_PTO TOK_ISA     # line 483
%left   TOK_TILDAEQ                                              # line 484
%left   TOK_IF                                                   # line 485
%left   TOK_ELSE                                                 # line 486
%left   TOK_COLON                                                # line 487
%left   TOK_PLUS TOK_MINUS                                       # line 488
%left   TOK_TIMES TOK_DIV                                        # line 489
%left   TOK_DOLLAR                                               # line 490
%left   TOK_OLD                                                  # line 491
%left   TOK_DOT                                                  # line 492
%right  TOK_ASSIGN                                               # line 493
```

NOTE: Python does NOT have `%right TOK_ASSIGN` or `TOK_ISA` in the precedence table.
Python has `TOK_ISA` nowhere in precedence; Go adds it at the relop level.
Python has no ASSIGN in precedence at all.

## 3. LEXER (lexer/lexer.go, lexer/token.go)

### Token types (token.go)
Same set as grammar tokens, plus: `EOF`, `ERROR`, `SYMBOL` (=PRESYMBOL after reserved check).

### Reserved keyword map (lexer.go allReserved)
Matches Python exactly, including aliases:
- `template` → MODULE, `instance` → INSTANTIATE, `execute` → MIXIN, `process` → EXTRACT

### Version-based keyword gating (lexer.go buildReserved)
Matches Python's LexerVersion class:
- v≤1.0: remove `state`, `local`
- v≤1.1: remove `returns mixin before after isolate with export delegate import include`; v>1.1: remove `state set null match`
- v≤1.4: remove 28 keywords (function, class, object, etc.)
- v≤1.5: remove `variant of globally eventually temporal`
- v≤1.6: remove 21 keywords (decreases, specification, etc.)
- v≤1.7: remove 16 keywords (global, common, debug, etc.); v>1.7: remove `requires ensures`

### Lexer character matching (lexer.go scan())
```
<<<...>>>  → NATIVEQUOTE
...        → DOTDOTDOT
..         → DOTS
.          → DOT
->         → ARROW
-          → MINUS
<->        → IFF
<=         → LE
<          → LT
>=         → GE
>          → GT
~=         → TILDAEQ
~          → TILDA
*>         → PTO
*          → TIMES
:=         → ASSIGN
:          → COLON
,→COMMA  (→LPAREN  )→RPAREN  {→LCB  }→RCB  [→LB  ]→RB
+→PLUS  /→DIV  &→AND  |→OR  =→EQ  ;→SEMI  $→DOLLAR  ^→CARET
□ (U+25A1) → GLOBALLY
◇ (U+25C7) → EVENTUALLY
"..."      → SYMBOL (quoted)
[A-Z]...   → VARIABLE (with optional [subscript])
[_a-z0-9]... → SYMBOL or reserved keyword
```

### Lex() adapter (parser/lalr_parser.go lines 96-422)
Maps lexer tokens to TOK_* constants. Unmapped tokens default to TOK_PRESYMBOL.

## 4. GRAMMAR PRODUCTIONS

### 4.1 top (line 504) — 53 alternatives

```
top : /* empty */                                                    # line 505
    | top TOK_USING SYMBOLx                                         # line 513
    | top TOK_INCLUDE SYMBOLx                                       # line 519
    | top optexplicit opttemporal TOK_AXIOM lgprop                   # line 549
    | top optexplicit opttemporal TOK_PROPERTY labeledfmla optskolem optproof  # line 565
    | top TOK_CONJECTURE labeledfmla                                 # line 590
    | top optexplicit TOK_INVARIANT labeledfmla optproof             # line 599
    | top TOK_UNPROVABLE TOK_INVARIANT labeledfmla optproof          # line 615
    | top TOK_MODULE modulestart modcat atom optwith TOK_EQ TOK_LCB top TOK_RCB moduleend  # line 630
    | top TOK_OBJECT objsym objectargs TOK_EQ TOK_LCB optdotdotdot top TOK_RCB objectend  # line 640
    | top TOK_CLASS objsym objectargs TOK_EQ TOK_LCB optdotdotdot top TOK_RCB objectend   # line 650
    | top TOK_SUBCLASS objsym TOK_OF atype TOK_EQ TOK_LCB optdotdotdot top TOK_RCB objectend  # line 670
    | top optexplicit TOK_DEFINITION optlabel gdefn optproof         # line 684
    | top TOK_SCHEMA schdefn                                         # line 698
    | top TOK_THEOREM schdefn optproof                               # line 707
    | top TOK_THEOREM labelname schdefnrhs optproof                  # line 719
    | top TOK_PROOF labelname proofstep                              # line 733
    | top TOK_INSTANTIATE insts                                      # line 742
    | top TOK_AUTOINSTANCE insts                                     # line 749
    | top symdecl                                                    # line 757
    | top TOK_RELATION rels                                          # line 764
    | top TOK_FUNCTION funs                                          # line 773
    | top TOK_DERIVED defns                                          # line 782
    | top optfinite optghost TOK_TYPE typesymbol                     # line 794
    | top optfinite optghost TOK_TYPE typesymbol TOK_EQ sort         # line 808
    | top TOK_PROGRESS defns                                         # line 822
    | top TOK_RELY atom TOK_ARROW atom                               # line 830
    | top TOK_RELY atom                                              # line 838
    | top TOK_MIXORD callatom TOK_ARROW callatom                     # line 846
    | top TOK_CONCEPT cdefns                                         # line 855
    | top TOK_UPDATE apps TOK_FROM apps upaxes                       # line 863
    | top TOK_MACRO atom TOK_EQ sequence                             # line 873
    | top optimpex actmeth SYMBOLx optargs optreturns optactiondef   # line 882
    | top TOK_MIXIN callatom TOK_BEFORE callatom                     # line 906
    | top TOK_MIXIN callatom TOK_AFTER callatom                      # line 915
    | top TOK_BEFORE atype optargs optreturns sequence               # line 924
    | top TOK_AFTER atype optargs optreturns topseq                  # line 933
    | top TOK_AROUND atype optargs optreturns TOK_LCB actseq optsemi TOK_DOTDOTDOT actseq optsemi TOK_RCB  # line 942
    | top TOK_AFTER TOK_INIT optargs topseq                         # line 969
    | top TOK_IMPLEMENT atype optargs optreturns topseq              # line 984
    | top TOK_IMPLEMENT TOK_TYPE SYMBOLx TOK_WITH SYMBOLx           # line 993
    | top opttrusted TOK_ISOLATE SYMBOLx optargs TOK_EQ callatoms   # line 1004
    | top opttrusted TOK_ISOLATE SYMBOLx optargs TOK_EQ callatoms TOK_WITH callatoms  # line 1016
    | top opttrusted TOK_ISOLATE SYMBOLx optargs TOK_EQ TOK_LCB top TOK_RCB optwith  # line 1028
    | top TOK_EXTRACT objsym objectargs TOK_EQ TOK_LCB top TOK_RCB optwith  # line 1047
    | top TOK_EXTRACT objsym objectargs TOK_EQ callatoms             # line 1061
    | top TOK_EXPORT callatom                                        # line 1068
    | top TOK_IMPORT callatom                                        # line 1076
    | top TOK_DELEGATE callatoms optdelegee                          # line 1084
    | top TOK_INTERPRET oper TOK_ARROW oper                          # line 1100
    | top TOK_INTERPRET oper TOK_ARROW TOK_LCB term TOK_DOTS term TOK_RCB  # line 1113
    | top TOK_INTERPRET oper TOK_ARROW TOK_LCB SYMBOLx moresymbols TOK_RCB  # line 1127
    | top TOK_ALIAS SYMBOLx TOK_EQ callatom                         # line 1151
    | top TOK_ATTRIBUTE callatom TOK_EQ attributeval                  # line 1160
    | top TOK_VARIANT typesymbol TOK_OF atype                        # line 1172
    | top TOK_VARIANT typesymbol TOK_OF atype TOK_EQ sort            # line 1185
    | top TOK_NATIVEQUOTE                                            # line 1198
    | top TOK_SCENARIO TOK_LCB sceninit TOK_SEMI scentranss TOK_RCB  # line 1205
    | top specimpl TOK_LCB top TOK_RCB                               # line 1215
    | top TOK_STATE SYMBOLx TOK_EQ state_expr                       # line 1225
    | top TOK_ASSERT SYMBOLx TOK_ARROW assert_rhs                   # line 1233
    ;
```

### 4.2 SYMBOL and subscripts (lines 1244-1273)

```
SYMBOLx    : TOK_PRESYMBOL                          # line 1244
           | SYMBOLx TOK_LB SYMsubscr TOK_RB        # line 1250
           ;

SYMsubscr  : SYMBOLx                                # line 1257
           | TOK_THIS                                # line 1263
           | SYMsubscr TOK_DOT SYMBOLx               # line 1268
           ;
```

### 4.3 atype (lines 1279-1303)

```
atype      : SYMBOLx                                # line 1279
           | atype TOK_DOT SYMBOLx                   # line 1285
           | TOK_THIS                                # line 1296
           ;
```

### 4.4 appelem (lines 1309-1323)

```
appelem    : SYMBOLx                                # line 1309
           | SYMBOLx TOK_LPAREN terms TOK_RPAREN     # line 1317
           ;
```

### 4.5 Variables (lines 1330-1389)

```
var        : TOK_VARIABLE                            # line 1330
           | TOK_VARIABLE TOK_COLON atype             # line 1338
           ;

simplevar  : TOK_VARIABLE                            # line 1348
           | TOK_VARIABLE TOK_COLON SYMBOLx           # line 1356
           ;

vars       : var                                     # line 1366
           | vars TOK_COMMA var                       # line 1372
           ;

simplevars : simplevar                               # line 1379
           | simplevars TOK_COMMA simplevar           # line 1385
           ;
```

### 4.6 terms (lines 1396-1411)

```
terms      : /* empty */                             # line 1396
           | term                                    # line 1402
           | terms TOK_COMMA term                     # line 1407
           ;
```

### 4.7 term (lines 1418-1709) — 38 alternatives

```
term       : appelem                                 # line 1418
           | var                                     # line 1424
           | TOK_OLD appelem                          # line 1429
           | term TOK_DOT appelem                     # line 1436
           | TOK_LPAREN term TOK_RPAREN               # line 1464
           | term TOK_PLUS term                       # line 1478
           | term TOK_MINUS term                      # line 1486
           | term TOK_TIMES term                      # line 1494
           | term TOK_DIV term                        # line 1502
           | term TOK_IF fmla TOK_ELSE term           # line 1510
           | term TOK_EQ term                         # line 1524
           | term TOK_LE term                         # line 1533
           | term TOK_LT term                         # line 1542
           | term TOK_GE term                         # line 1551
           | term TOK_GT term                         # line 1560
           | term TOK_PTO term                        # line 1569
           | term TOK_TILDAEQ term                    # line 1578
           | TOK_TRUE                                 # line 1587
           | TOK_FALSE                                # line 1594
           | TOK_TILDA term                           # line 1601
           | term TOK_AND term                        # line 1610
           | term TOK_OR term                         # line 1622
           | term TOK_ARROW term                      # line 1634
           | term TOK_IFF term                        # line 1643
           | TOK_FORALL simplevars TOK_DOT term %prec TOK_SEMI   # line 1652
           | TOK_EXISTS simplevars TOK_DOT term %prec TOK_SEMI   # line 1660
           | TOK_FORALL TOK_LPAREN vars TOK_RPAREN term          # line 1668
           | TOK_EXISTS TOK_LPAREN vars TOK_RPAREN term          # line 1676
           | TOK_GLOBALLY term                        # line 1684
           | TOK_EVENTUALLY term                      # line 1690
           | term TOK_WHENNEXT term                   # line 1696
           | term TOK_WHENPREV term                   # line 1700
           | term TOK_WHENFIRST term                  # line 1704
           | term TOK_WHENLAST term                   # line 1708
           | term TOK_ISA atype                       # line  (after WHENLAST)
           | term TOK_COLON atype                     # line  (sort annotation)
           | TOK_LPAREN TOK_DOLLAR SYMBOLx simplevars TOK_DOT fmla TOK_RPAREN TOK_LPAREN terms TOK_RPAREN  # named binder
           | TOK_DOLLAR SYMBOLx TOK_DOT fmla %prec TOK_SEMI     # named binder
           | TOK_DOLLAR SYMBOLx TOK_DOLLAR fmla %prec TOK_SEMI  # named binder
           ;
```

### 4.8 fmla (line 1712)

```
fmla       : term                                    # line 1712
           ;
```

### 4.9 labeledfmla, labelname (lines 1724-1757)

```
labeledfmla : fmla                                   # line 1724
            | labelname fmla                          # line 1731
            ;

labelname  : TOK_LB SYMBOLx TOK_RB                  # line 1742
           | TOK_LABEL                                # line 1750
           ;
```

### 4.10 gprop, lgprop (lines 1759-1781)

```
gprop      : fmla                                    # line 1759
           | schdefnrhs                              # line 1765
           ;

lgprop     : optlabel gprop                          # line 1772
           ;
```

### 4.11 Optional markers (lines 1785-1868)

```
opttemporal   : /* empty */                          # line 1785
              | TOK_TEMPORAL                          # line 1791
              ;
optunprovable : /* empty */                          # line 1798
              | TOK_UNPROVABLE                        # line 1804
              ;
optexplicit   : /* empty */                          # line 1811
              | TOK_EXPLICIT                          # line 1817
              ;
optlabel      : /* empty */                          # line 1824
              | labelname                             # line 1830
              ;
optskolem     : /* empty */                          # line 1837
              | TOK_NAMED defnlhs                     # line 1843
              ;
optproof      : /* empty */                          # line 1850
              | TOK_PROOF proofstep                   # line 1856
              | TOK_PROOF labelname proofstep         # line 1862
              ;
optsemi       : /* empty */                          # line 1871
              | TOK_SEMI                              # line 1877
              ;
```

### 4.12 Definitions (lines 1888-2068)

```
dotsym     : SYMBOLx                                # line 1888
           | dotsym TOK_DOT SYMBOLx                   # line 1895
           ;

defnlhs    : dotsym                                  # line 1902
           | dotsym TOK_LPAREN defargs TOK_RPAREN     # line 1908
           | TOK_LPAREN defarg relop defarg TOK_RPAREN  # line 1917
           | TOK_LPAREN defarg infix defarg TOK_RPAREN  # line 1924
           ;

defarg     : lparam                                  # line 1931
           | var                                     # line 1937
           ;

defargs    : defarg                                  # line 1944
           | defargs TOK_COMMA defarg                 # line 1950
           ;

typeddefn  : defnlhs                                 # line 1957
           | defnlhs TOK_COLON atype                  # line 1963
           ;

defnrhs    : fmla                                    # line 1976
           | somevarfmla                             # line 1982
           | TOK_NATIVEQUOTE                          # line 1988
           ;

defn       : typeddefn TOK_EQ defnrhs                # line 1995
           ;

defns      : defn                                    # line 2005
           | defns TOK_COMMA defn                     # line 2011
           ;

gdefn      : defn                                    # line 2018
           | TOK_LCB defn TOK_RCB                     # line 2024
           ;

somevarfmla : TOK_SOME simplevar TOK_DOT fmla optin optelse  # line 2032
            ;

optin      : /* empty */                             # line 2043
           | TOK_IN fmla                              # line 2049
           ;

optelse    : /* empty */                             # line 2056
           | TOK_ELSE fmla                            # line 2062
           ;
```

### 4.13 Schema (lines 2073-2190)

```
schdefnrhs : fmla                                    # line 2073
           | TOK_LCB schdecls schconc TOK_RCB         # line 2079
           ;

schdecl    : TOK_FUNCTION funs                        # line 2087
           | TOK_FRESH TOK_FUNCTION funs               # line 2095
           | TOK_INDIV funs                            # line 2103
           | TOK_FRESH TOK_INDIV funs                  # line 2111
           | TOK_RELATION rels                         # line 2119
           | TOK_FRESH TOK_RELATION rels               # line 2127
           | TOK_TYPE SYMBOLx                          # line 2135
           | optexplicit TOK_PROPERTY lgprop           # line 2141
           | TOK_THEOREM lgprop                        # line 2147
           | schdefnrhs                               # line 2151
           ;

schdecls   : /* empty */                             # line 2155
           | schdecls schdecl                         # line 2161
           ;

schconc    : TOK_DEFINITION defn                      # line 2168
           | optexplicit TOK_PROPERTY lgprop           # line 2174
           ;

schdefn    : defnlhs TOK_EQ schdefnrhs               # line 2182
           ;
```

### 4.14 Symbol declarations (lines 2194-2268)

```
symdecl    : constantdecl                            # line 2194
           | TOK_DESTRUCTOR tterms                    # line 2200
           | TOK_FIELD tterms                         # line 2206
           | TOK_CONSTRUCTOR tterms                   # line 2213
           ;

constantdecl : TOK_INDIV tterms                      # line 2220
             | TOK_VAR tterms                         # line 2226
             | TOK_PARAMETER parameter                # line 2233
             ;

parameter  : tterm                                   # line 2240
           | tterm TOK_EQ paramval                    # line 2247
           ;

paramval   : TOK_TRUE                                # line 2256
           | TOK_FALSE                                # line 2261
           | SYMBOLx                                  # line 2266
           ;
```

### 4.15 tapp, tterm, tterms, targs, tsyms (lines 2274-2351)

```
tapp       : SYMBOLx                                 # line 2274
           | SYMBOLx targs                            # line 2282
           | TOK_LPAREN var infix var TOK_RPAREN       # line 2290
           ;

tterm      : tapp                                    # line 2300
           | tapp TOK_COLON atype                     # line 2306
           ;

tterms     : tterm                                   # line 2316
           | tterms TOK_COMMA tterm                   # line 2322
           ;

targs      : TOK_LPAREN TOK_RPAREN                    # line 2329
           | TOK_LPAREN tsyms TOK_RPAREN               # line 2335
           ;

tsyms      : var                                     # line 2342
           | tsyms TOK_COMMA var                       # line 2348
           ;
```

### 4.16 tatom, tatoms (lines 2355-2381)

```
tatom      : SYMBOLx                                 # line 2355
           | SYMBOLx targs                            # line 2361
           | TOK_LPAREN var relop var TOK_RPAREN       # line 2367
           ;

tatoms     : tatom                                   # line 2373
           | tatoms TOK_COMMA tatom                   # line 2379
           ;
```

### 4.17 rel, rels, fun, funs (lines 2386-2446)

```
rel        : defnlhs                                 # line 2386
           | defn                                    # line 2394
           ;

rels       : rel                                     # line 2403
           | rels TOK_COMMA rel                       # line 2409
           ;

fun        : typeddefn                               # line 2416
           | typeddefn TOK_EQ defnrhs                 # line 2424
           ;

funs       : fun                                     # line 2434
           | funs TOK_COMMA fun                       # line 2440
           ;
```

### 4.18 Type (lines 2449-2535)

```
typesymbol : SYMBOLx                                 # line 2449
           | TOK_THIS                                 # line 2455
           ;

optfinite  : /* empty */                             # line 2464
           | TOK_FINITE                               # line 2470
           ;

optghost   : /* empty */                             # line 2477
           | TOK_GHOST                                # line 2483
           ;

sort       : TOK_LCB SYMBOLx TOK_RCB                  # line 2490
           | TOK_LCB SYMBOLx TOK_COMMA names TOK_RCB  # line 2497
           | TOK_LCB SYMBOLx TOK_DOTS SYMBOLx TOK_RCB  # line 2504
           | TOK_STRUCT TOK_LCB tterms TOK_RCB         # line 2511
           | TOK_STRUCT TOK_LCB TOK_RCB                # line 2517
           ;

names      : SYMBOLx                                 # line 2522
           | names TOK_COMMA SYMBOLx                   # line 2528
           ;
```

### 4.19 relop, infix (lines 2539-2555)

```
relop      : TOK_EQ                                  # line 2539
           | TOK_LE                                   #
           | TOK_LT                                   #
           | TOK_GE                                   #
           | TOK_GT                                   #
           | TOK_PTO                                  #
           ;

infix      : TOK_PLUS                                # line 2548
           | TOK_MINUS                                #
           | TOK_TIMES                                #
           | TOK_DIV                                  #
           ;
```

### 4.20 atom, atoms, app, apps, lit (lines 2559-2633)

```
atom       : SYMBOLx                                 # line 2559
           | SYMBOLx TOK_LPAREN terms TOK_RPAREN      # line 2567
           ;

atoms      : atom                                    # line 2576
           | atoms TOK_COMMA atom                     # line 2582
           ;

app        : SYMBOLx                                 # line 2589
           | SYMBOLx TOK_LPAREN terms TOK_RPAREN      # line 2595
           | term infix term                          # line 2601
           ;

apps       : app                                     # line 2607
           | apps TOK_COMMA app                       # line 2613
           ;

lit        : atom                                    # line 2620
           | TOK_TILDA lit                            # line 2628
           ;
```

NOTE: Python `lit` also has `SYMBOL EQ SYMBOL` and `SYMBOL TILDAEQ SYMBOL`.
Go `lit` is missing those two alternatives.

### 4.21 callatom, callatoms (lines 2637-2676)

```
callatom   : atom                                    # line 2637
           | TOK_THIS                                 # line 2643
           | TOK_METHOD                               # line 2649
           | callatom TOK_DOT callatom                # line 2655
           ;

callatoms  : callatom                                # line 2664
           | callatoms TOK_COMMA callatom             # line 2670
           ;
```

### 4.22 Module/Object infrastructure (lines 2681-2792)

```
modulestart : /* empty */                            # line 2681
            ;
moduleend   : /* empty */                            # line 2689
            ;
objectend   : /* empty */                            # line 2697
            ;

modcat     : /* empty */                             # line 2705
           | TOK_OBJECT                               # line 2711
           | TOK_ISOLATE                              # line 2717
           ;

opteq      : /* empty */                             # line 2723
           | TOK_EQ                                   # line 2729
           ;

optdotdotdot : /* empty */                           # line 2736
             | TOK_DOTDOTDOT                           # line 2743
             ;

objectargs : optargs                                 # line 2752
           ;

objsym     : SYMBOLx                                 # line 2762
           ;

opttrusted : /* empty */                             # line 2772
           | TOK_TRUSTED                              # line 2778
           ;

optargs    : /* empty */                             # line 2785
           | TOK_LPAREN lparams TOK_RPAREN             # line 2791
           ;

optreturns : /* empty */                             # line 2798
           | TOK_RETURNS TOK_LPAREN lparams TOK_RPAREN  # line 2804
           ;

optactualreturns : /* empty */                       # line 2811
                 | callatoms TOK_ASSIGN               # line 2817
                 ;

param      : SYMBOLx TOK_COLON SYMBOLx               # line 2824
           ;

params     : param                                   # line 2834
           | params TOK_COMMA param                   # line 2840
           ;

optwith    : /* empty */                             # line 2847
           | TOK_WITH callatoms                       # line 2853
           ;
```

### 4.23 lparam, lparams (lines 2864-2895)

```
lparam     : SYMBOLx TOK_COLON atype                 # line 2864
           | TOK_CARET SYMBOLx TOK_COLON atype        # line 2872
           ;

lparams    : lparam                                  # line 2882
           | lparams TOK_COMMA lparam                 # line 2888
           ;
```

### 4.24 Action definitions (lines 2899-2961)

```
optactiondef : /* empty */                           # line 2899
             | TOK_EQ topseq                          # line 2905
             | TOK_EQ TOK_TIMES                        # line 2910
             ;

topseq     : sequence                                # line 2917
           | TOK_LCB TOK_NATIVEQUOTE TOK_RCB           # line 2923
           ;

optimpex   : /* empty */                             # line 2933
           | TOK_EXPORT                               # line 2939
           | TOK_IMPORT                               # line 2944
           ;

actmeth    : TOK_ACTION                              # line 2951
           | TOK_METHOD                               # line 2957
           ;
```

### 4.25 specimpl (lines 2968-2993)

```
specimpl   : TOK_SPECIFICATION                       # line 2968
           | TOK_IMPLEMENTATION                       # line 2974
           | TOK_PRIVATE                              # line 2979
           | TOK_GLOBAL                               # line 2984
           | TOK_COMMON                               # line 2989
           ;
```

### 4.26 Instantiate (lines 3000-3110)

```
insts      : inst                                    # line 3000
           | insts TOK_COMMA inst                     # line 3006
           ;

inst       : modinst                                 # line 3013
           | modinst TOK_COLON modinst                # line 3021
           ;

modinst    : dotsym                                  # line 3030
           | dotsym TOK_LPAREN pnames TOK_RPAREN       # line 3036
           ;

pname      : atype                                   # line 3045
           | var                                     # line 3062
           | infix                                   # line 3067
           | relop                                   # line 3072
           | TOK_THIS                                 # line 3077
           | TOK_TRUE                                 # line 3082
           | TOK_FALSE                                # line 3087
           ;

pnames     : /* empty */                             # line 3094
           | pname                                   # line 3100
           | pnames TOK_COMMA pname                   # line 3105
           ;
```

### 4.27 oper, attributeval, moresymbols, optdelegee (lines 3116-3185)

```
oper       : atype                                   # line 3116
           | relop                                   # line 3122
           | infix                                   # line 3127
           | TOK_NATIVEQUOTE                          # line 3132
           ;

attributeval : callatom                              # line 3140
             | TOK_TRUE                               # line 3146
             | TOK_FALSE                              # line 3152
             ;

moresymbols : /* empty */                            # line 3158
            | moresymbols TOK_COMMA SYMBOLx            # line 3164
            ;

optdelegee : /* empty */                             # line 3175
           | TOK_ARROW callatom                       # line 3181
           ;
```

### 4.28 Sequences and actions (lines 3192-3505)

```
sequence   : TOK_LCB TOK_RCB                          # line 3192
           | TOK_LCB actseq TOK_RCB                    # line 3198
           | TOK_LCB actseq TOK_SEMI TOK_RCB           # line 3208
           ;

actseq     : actseqrev                               # line 3220 (reverses list)
           ;

actseqrev  : simpleact                               # line 3232
           | complexact                              # line 3238
           | simpleact TOK_SEMI actseqrev              # line 3243
           | simpleact TOK_SEMI                        # line 3248
           | complexact actseqrev                     # line 3253
           | complexact TOK_SEMI actseqrev             # line 3258
           | complexact TOK_SEMI                       # line 3263
           ;

action     : simpleact                               # line 3270
           | complexact                              # line 3276
           ;
```

### 4.29 simpleact (lines 3287-3392) — 18 alternatives

```
simpleact  : TOK_ASSUME labeledfmla                                    # line 3287
           | optunprovable TOK_ASSERT labeledfmla                      # line 3293
           | optunprovable TOK_ASSERT labeledfmla TOK_PROOF proofstep  # line 3303
           | optunprovable TOK_REQUIRE labeledfmla                     # line 3308
           | optunprovable TOK_REQUIRE labeledfmla TOK_PROOF proofstep # line 3313
           | optunprovable TOK_ENSURE labeledfmla                      # line 3318
           | optunprovable TOK_ENSURE labeledfmla TOK_PROOF proofstep  # line 3323
           | term TOK_ASSIGN fmla                                      # line 3328
           | termtuple TOK_ASSIGN callatom                             # line 3335
           | term TOK_ASSIGN TOK_TIMES                                 # line 3340
           | TOK_VAR tterm                                             # line 3345
           | TOK_VAR tterm TOK_ASSIGN fmla                             # line 3350
           | TOK_CALL optactualreturns callatom                        # line 3355
           | TOK_CALL callatom                                         # line 3361
           | TOK_SET lit                                               # line 3366
           | TOK_INSTANTIATE callatom                                  # line 3371
           | TOK_UNPROVABLE simpleact                                  # line 3376
           | TOK_DEBUG SYMBOLx optdebugargs                            # line 3381
           | term     %prec TOK_SEMI                                   # line 3387
           ;
```

### 4.30 termtuple, debugarg/args (lines 3394-3434)

```
termtuple  : TOK_LPAREN term TOK_COMMA terms TOK_RPAREN  # line 3394
           ;

debugarg   : SYMBOLx TOK_EQ fmla                     # line 3403
           ;

debugargs  : debugarg                                # line 3411
           | debugargs TOK_COMMA debugarg              # line 3417
           ;

optdebugargs : /* empty */                           # line 3424
             | TOK_WITH debugargs                     # line 3430
             ;
```

### 4.31 complexact (lines 3441-3505) — 10 alternatives

```
complexact : sequence                                # line 3441
           | TOK_IF somefmla sequence                 # line 3447
           | TOK_IF somefmla sequence TOK_ELSE action  # line 3456
           | TOK_IF TOK_TIMES sequence TOK_ELSE action  # line 3465
           | TOK_WHILE somefmla invariants decreases sequence  # line 3472
           | TOK_FOR tterm TOK_COMMA tterm TOK_IN fmla invariants decreases sequence  # line 3483
           | TOK_LOCAL lparams sequence               # line 3488
           | TOK_LET eqns sequence                    # line 3494
           | TOK_THUNK labelname SYMBOLx optargs TOK_COLON atype TOK_ASSIGN sequence  # line 3500
           ;
```

### 4.32 somefmla, bounds (lines 3509-3551)

```
somefmla   : fmla                                    # line 3509
           | fmla TOK_ASSIGN fmla                     # line 3515
           | TOK_SOME bounds fmla                     # line 3520
           | TOK_SOME bounds fmla TOK_MINIMIZING term  # line 3526
           | TOK_SOME bounds fmla TOK_MAXIMIZING term  # line 3532
           ;

bounds     : params TOK_DOT                           # line 3540
           | TOK_LPAREN lparams TOK_RPAREN             # line 3546
           ;
```

### 4.33 invariants, decreases (lines 3553-3582)

```
invariants : /* empty */                             # line 3553
           | invariants TOK_INVARIANT labeledfmla     # line 3559
           | invariants TOK_INVARIANT labeledfmla TOK_PROOF proofstep  # line 3564
           ;

decreases  : /* empty */                             # line 3571
           | TOK_DECREASES fmla                       # line 3577
           ;
```

### 4.34 eqn, eqns (lines 3586-3604)

```
eqn        : SYMBOLx TOK_EQ SYMBOLx                   # line 3586
           ;

eqns       : eqn                                    # line 3594
           | eqns TOK_COMMA eqn                       # line 3600
           ;
```

### 4.35 Scenario (lines 3611-3681)

```
sceninit   : TOK_ARROW places                        # line 3611
           ;

places     : SYMBOLx                                 # line 3619
           | places TOK_COMMA SYMBOLx                  # line 3625
           ;

scentranss : /* empty */                             # line 3632
           | scentranss scentrans                     # line 3638
           ;

scentrans  : places TOK_ARROW places TOK_COLON scenariomixin   # line 3645
           | places TOK_COLON scenariomixin            # line 3653
           ;

scenariomixin : TOK_BEFORE atype optargs optreturns sequence   # line 3661
              | TOK_AFTER atype optargs optreturns sequence     # line 3672
              ;
```

### 4.36 Proof/tactic (lines 3688-4014) — 24 proofstep alternatives

```
pflet      : var TOK_EQ fmla                          # line 3688
           ;

pflets     : pflet                                   # line 3696
           | pflets TOK_COMMA pflet                   # line 3702
           ;

tacticwithelem : TOK_INVARIANT labeledfmla            # line 3709
               | TOK_DEFINITION atype TOK_EQ fmla     # line 3715
               | TOK_TRIGGER atype TOK_WITH terms      # line 3720
               ;

tacticwithlist : tacticwithelem                       # line 3727
               | tacticwithlist tacticwithelem         # line 3733
               ;

tacticwithlistchoice : tacticwithlist                 # line 3740
                     | pflets                         # line 3746
                     ;

opttacticwith : /* empty */                          # line 3753
              | TOK_WITH tacticwithlistchoice          # line 3759
              | TOK_WITH TOK_LCB tacticwithlist TOK_RCB  # line 3764
              ;

proofgroup : TOK_LCB proofseq TOK_RCB                 # line 3771
           | TOK_LCB TOK_RCB                           # line 3777
           ;

optproofgroup : /* empty */                          # line 3784
              | proofgroup                            # line 3790
              ;

proofseq   : proofstep                               # line 3797
           | proofseq TOK_SEMI proofstep              # line 3803
           | proofseq proofstep                       # line 3808
           ;

proofstep  : TOK_APPLY atype optrenaming                                  # line 3893
           | TOK_APPLY atype optrenaming TOK_WITH matches                 # line 3899
           | TOK_ASSUME atype optrenaming                                 # line 3904
           | TOK_ASSUME atype optrenaming TOK_WITH matches                # line 3909
           | TOK_INSTANTIATE atype optrenaming                            # line 3914
           | TOK_INSTANTIATE labelname atype optrenaming                  # line 3919
           | TOK_INSTANTIATE atype optrenaming TOK_WITH matches           # line 3924
           | TOK_INSTANTIATE TOK_WITH pflets                              # line 3929
           | TOK_SHOWGOALS                                                # line 3934
           | TOK_DEFERGOAL                                                # line 3939
           | TOK_SPOIL atype                                              # line 3944
           | TOK_TACTIC atype opttacticwith optproofgroup                 # line 3949
           | opttemporal TOK_PROPERTY labeledfmla optskolem optproofgroup # line 3954
           | TOK_FUNCTION funs                                            # line 3967
           | TOK_THEOREM lgprop optproofgroup                             # line 3972
           | TOK_PROOF labelname proofgroup                               # line 3978
           | TOK_LET pflets                                               # line 3983
           | TOK_IF fmla proofgroup TOK_ELSE proofgroup                   # line 3988
           | TOK_UNFOLD atype TOK_WITH callatoms                          # line 3993
           | TOK_UNFOLD TOK_WITH callatoms                                # line 4000
           | TOK_FORGET callatoms                                         # line 4005
           | proofgroup                                                   # line 4010
           ;
```

### 4.37 match, renaming (lines 3817-3888)

```
match      : defn                                    # line 3817
           | var TOK_EQ fmla                          # line 3823
           ;

matches    : match                                   # line 3830
           | matches TOK_COMMA match                  # line 3836
           ;

renamingitem : TOK_VARIABLE TOK_DIV TOK_VARIABLE      # line 3843
             | SYMBOLx TOK_DIV SYMBOLx                 # line 3849
             ;

renaminglist : renamingitem                           # line 3856
             | renaminglist TOK_COMMA renamingitem      # line 3862
             ;

optrenaming  : /* empty */                            # line 3869
             | renaming                               # line 3875
             ;

renaming     : TOK_LT renaminglist TOK_GT              # line 3882
             ;
```

### 4.38 Update/state/concept (lines 4021-4269)

```
requires   : /* empty */                             # line 4021
           | TOK_REQUIRES fmla                        # line 4027
           ;

ensures    : TOK_ENSURES fmla                        # line 4034
           ;

modifies   : /* empty */                             # line 4042
           | TOK_MODIFIES TOK_LCB TOK_RCB              # line 4048
           | TOK_MODIFIES TOK_TIMES                    # line 4053
           | TOK_MODIFIES atoms                       # line 4058
           ;

upaxes     : /* empty */                             # line 4065
           | upaxes upax                              # line 4071
           ;

upax       : TOK_PARAMS tterms TOK_IN action TOK_ARROW requires ensures  # line 4078
           ;

assert_rhs : TOK_LCB requires modifies ensures TOK_RCB  # line 4086
           | fmla                                     # line 4092
           ;

state_expr : TOK_TRUE                                # line 4099
           | TOK_FALSE                                # line 4105
           | SYMBOLx                                  # line 4110
           | SYMBOLx TOK_LPAREN state_expr TOK_RPAREN  # line 4115
           | state_expr TOK_OR state_expr              # line 4120
           | TOK_LCB requires modifies ensures TOK_RCB  # line 4125
           | TOK_ENTRY                                 # line 4130
           ;

cdefn      : atom TOK_EQ expr                        # line 4139
           ;

cdefns     : cdefn                                   # line 4147
           | cdefns TOK_COMMA cdefn                   # line 4153
           ;

expr       : TOK_LCB fmla TOK_RCB                     # line 4160
           | exprterm                                 # line 4166
           | exprterm relop exprterm                   # line 4171
           | exprterm TOK_TILDAEQ exprterm             # line 4176
           | TOK_TILDA expr                            # line 4181
           | TOK_LPAREN expr TOK_RPAREN                # line 4186
           | prod                                     # line 4191
           | sum                                      # line 4196
           ;

exprterm   : appelem                                 # line 4203
           | var                                     # line 4209
           ;

prod       : expr TOK_TIMES expr                      # line 4216
           | prod TOK_TIMES expr                       # line 4222
           ;

sum        : expr TOK_PLUS expr                       # line 4229
           | sum TOK_PLUS expr                         # line 4235
           ;

loc        : /* empty */                             # line 4244
           | SYMBOLx                                  # line 4250
           ;

symbols    : SYMBOLx                                 # line 4259
           | symbols TOK_COMMA SYMBOLx                 # line 4265
           ;
```

## 5. HELPER FUNCTIONS (grammar_v17.y)

### In %{ %} section (lines 12-270)
```
getLineno(lex)                     # line 32  — returns ast.Location
newLabel(pref)                     # line 41  — generates unique label
addLabel(lf, pref)                 # line 49  — adds label if missing
mkLF(x)                           # line 61  — wraps in LabeledFormula
checkNonTemporal(x)                # line 71  — validates no temporal ops
addUnprovable(lf, cond)            # line 86  — marks unprovable
makeMixinName(atom, suffix)        # line 96  — generates mixin name
handleMixin(kind, mixer, mixee, ivy)  # line 104
stackActionLookup(ivy, name)       # line 128
inferActionParams(ivy, actname, formals, returns)  # line 143
handleBeforeAfter(kind, atom, action, ivy, optargs, optreturns)  # line 154
setObjectDefined(ivy, name)        # line 169
parseNativequote(raw, lex)         # line 176
fixIfPart(cond, part)              # line 210
createObject(top, name, objectargs, module, lineno, continuation)  # line 229
```

### After %% (lines 4273-4296)
```
lowerVarStmts(stmts)               # line 4277 — TODO: stub
lalrMakeSequence(stmts)             # line 4287 — wraps in Sequence
```

## 6. KNOWN DIFFERENCES FROM PYTHON (noted during extraction)

1. **Precedence:** Go adds `%right TOK_ASSIGN` and `TOK_ISA` in `%left` relop group. Python has neither.
2. **lit rule:** Go missing `SYMBOL EQ SYMBOL` and `SYMBOL TILDAEQ SYMBOL` alternatives.
3. **proofseq:** Go adds `proofseq proofstep` (no separator). Python only has `proofseq optsemi proofstep`.
4. **lowerVarStmts:** Go is a stub (TODO), Python has full implementation.
5. **LABEL token:** Go has `TOK_LABEL` in token declarations but the lexer adapter never returns it. `labelname` rule handles `TOK_LB SYMBOLx TOK_RB` instead.
6. **Unused tokens:** `TOK_VAR_KW`, `TOK_METHOD_KW`, `TOK_NULL_KW`, `TOK_SET_KW` declared but never used.

