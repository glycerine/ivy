# Ivy Python -> Go Port Plan

## Context

Ivy is a formal analysis language (~45K lines of Python) using Z3 for model checking. We're porting it to Go incrementally, bottom-up, with comprehensive unit tests and fuzz tests per chunk. The Go code will live at `goivy/` in the repo root. Parser will be hand-written recursive descent (no yacc/lex).

## Go Package Structure

```
goivy/
  go.mod                    # module github.com/kenmcmil/ivy/goivy
  logic/                    # Chunks 1-3: sorts, terms, formulas, errors
    sort.go                 # Sort interface + concrete sort types
    sort_test.go
    term.go                 # Var, Const, Apply
    term_test.go
    formula.go              # Eq, Not, And, Or, Implies, Iff, ForAll, Exists, etc.
    formula_test.go
    node.go                 # Node interface
    error.go                # IvyError, SortError
  typeinfer/                # Chunk 4: union-find type inference
    sortvar.go
    unify.go
    infer.go
    typeinfer_test.go
  logicutil/                # Chunk 5: free_variables, substitution, etc.
  ivyutils/                 # Chunk 6: parameters, naming, utilities
  parser/                   # Chunk 7+: hand-written recursive descent
    lexer.go
    parser.go
```

## Chunk 1: Sort Types + Error Types

**Source**: `logic.py` lines 14-101, `general.py`

**Files to create**: `goivy/go.mod`, `goivy/logic/error.go`, `goivy/logic/sort.go`, `goivy/logic/sort_test.go`

### error.go
- `IvyError` struct implementing `error` — wraps a message
- `SortError` struct implementing `error` — distinct type for `errors.As`

### sort.go
- `Sort` interface: `String() string`, `Equal(Sort) bool`, unexported `sortSeal()` method
- Concrete types:
  - `UninterpretedSort{Name string}`
  - `BooleanSort struct{}` — singleton `var Boolean = BooleanSort{}`
  - `FunctionSort{Sorts []Sort}` — last element is range, rest is domain
    - Methods: `Domain() []Sort`, `Range() Sort`, `Arity() int`
    - Constructor `NewFunctionSort(sorts ...Sort) (*FunctionSort, error)` — validates ≥1 sort, all first-order
  - `EnumeratedSort{Name string, Extension []string}`
    - Methods: `Card() int`
  - `RangeSort{Name string, Lb, Ub string}`
  - `TopSort{Name string}` — default `var TopS = TopSort{Name: "TopSort"}`
    - Method: `IsSortVariable() bool` (true when Name != "TopSort")
- Functions: `FirstOrderSort(s Sort) bool`

### Tests & Fuzz
- Table-driven tests for each sort constructor, `String()`, `Equal()`
- `FunctionSort` validation: zero args → error, higher-order args → error
- `FirstOrderSort` returns false only for `FunctionSort`
- Fuzz: `FuzzNewFunctionSort` with random sort slices — no panics, valid error or valid object
- Fuzz: `FuzzUninterpretedSort` with random name strings — `Equal` reflexivity

---

## Chunk 2: Terms (Var, Const, Apply)

**Source**: `logic.py` lines 115-184

**Files to create**: `goivy/logic/node.go`, `goivy/logic/term.go`, `goivy/logic/term_test.go`

### node.go
- `Node` interface: `NodeSort() Sort`, `Children() []Node`, `String() string`, `Equal(Node) bool`
- All Sort types get `NodeSort()` returning themselves, `Children()` returning nil

### term.go
- `Var{Name string, VSort Sort}` — constructor validates name starts with uppercase
- `Const{Name string, CSort Sort}` — no name restriction (commented out in Python)
- `Apply{Func Node, Terms []Node}` — constructor validates:
  - Func sort is FunctionSort or TopSort
  - Arity matches
  - Argument sorts match domain sorts (unless TopSort involved)
  - `NodeSort()` returns range of func sort (or TopS if func is TopSort)
- `Var` and `Const` implement `func (v *Var) Call(terms ...Node) (Node, error)` returning Apply or self

### Tests & Fuzz
- Var: valid names, invalid names (lowercase → error), `String()`, `Equal()`
- Const: construction, `String()`, `Equal()`
- Apply: valid application, arity mismatch → error, sort mismatch → error, TopSort passthrough
- Reproduce Python's `__main__` examples: `leq(X,Y)` etc.
- Fuzz: random Var names, random Apply constructions

---

## Chunk 3: Formulas (Eq through NamedBinder)

**Source**: `logic.py` lines 186-448

**Files to create**: `goivy/logic/formula.go`, `goivy/logic/formula_test.go`

### formula.go

All formula types — constructors validate sort constraints:

| Type | Fields | Validation |
|------|--------|------------|
| `Eq{T1, T2 Node}` | two terms | same sort, first-order |
| `Ite{ISort Sort, Cond, Then, Else Node}` | sort from Then | cond is Boolean |
| `Not{Body Node}` | one sub | body is Boolean |
| `Globally{Environ *string, Body Node}` | | body is Boolean |
| `Eventually{Environ *string, Body Node}` | | body is Boolean |
| `WhenOperator{WSort Sort, Name string, T1, T2 Node}` | sort from T1 | T2 is Boolean |
| `Cond{CSort Sort, T1, T2 Node}` | sort from T2 | |
| `And{Terms []Node}` | variadic | all Boolean |
| `Or{Terms []Node}` | variadic | all Boolean |
| `Implies{T1, T2 Node}` | | both Boolean |
| `Iff{T1, T2 Node}` | | both Boolean |
| `ForAll{Variables []*Var, Body Node}` | | ≥1 var, body Boolean |
| `Exists{Variables []*Var, Body Node}` | | ≥1 var, body Boolean |
| `Lambda{Variables []*Var, Body Node}` | | all must be Var |
| `NamedBinder{Name string, Variables []*Var, Environ *string, Body Node}` | | sort is FunctionSort of var sorts → body sort |

Package-level:
```go
var True Node = &And{}    // And with no terms
var False Node = &Or{}    // Or with no terms
```

### Tests & Fuzz
- Each formula type: valid construction, sort validation errors
- `Children()` returns only sub-structures (not meta fields like Variables, Environ)
- `String()` matches Python output format
- Reproduce `__main__` examples: transitive, antisymmetric formulas, NamedBinder
- `ContainsTopSort(n Node) bool` and `IsPolymorphic(n Node) bool` — test with Python's assertions
- Fuzz: build random well-typed formula trees, verify `NodeSort()` invariants

---

## Chunk 4: Type Inference (Union-Find + Sort Unification)

**Source**: `type_inference.py` (~305 lines)

**Files to create**: `goivy/typeinfer/sortvar.go`, `goivy/typeinfer/unify.go`, `goivy/typeinfer/infer.go`, `goivy/typeinfer/typeinfer_test.go`

### sortvar.go
- `SortVar struct{ Instance SortOrVar }` — mutable, used in union-find
- `SortOrVar` interface implemented by both `logic.Sort` and `*SortVar`

### unify.go
- `Find(x SortOrVar) SortOrVar` — path-compressing find
- `Unify(s1, s2 SortOrVar) error` — unification with occurs check
- `OccursIn(s1, s2 SortOrVar) bool`

### infer.go
- `InferSorts(node logic.Node, env map[string]SortOrVar) error`
- `ConcretizeSorts(node logic.Node, sort logic.Sort) (logic.Node, error)`

### Tests & Fuzz
- Union-find: chains, path compression, cycle detection
- Unify: same sorts, TopSort wildcards, SortVar assignment, FunctionSort element-wise, incompatible → error
- Reproduce `type_inference.py` `__main__` block examples
- Fuzz: random sort pairs for `Unify` — no panics

---

## Verification Plan

After each chunk:
1. `go build ./...` — compiles
2. `go test ./... -v` — all unit tests pass
3. `go test ./... -fuzz=. -fuzztime=30s` — fuzz tests find no crashes
4. `go vet ./...` — no warnings
5. Cross-reference with Python: run Python `logic.py __main__` block, verify Go produces identical `String()` output

## Future Chunks (brief)

- **Chunk 5**: `logicutil` — `FreeVariables`, `UsedVariables`, `Substitute`, etc. from `logic_util.py`
- **Chunk 6**: `ivyutils` — `Parameter`, `UniqueRenamer`, `TopologicalSort`, name utilities from `ivy_utils.py`
- **Chunk 7+**: `parser` — hand-written recursive descent lexer+parser (replaces PLY lex/yacc)

# PROGRESS UPDATE 1

# Ivy Python → Go Port: Progress & Next Steps

## Context

Ivy is a formal analysis language (~45K lines of Python) using Z3 for model checking. We're porting it to Go incrementally, bottom-up, with comprehensive unit tests and fuzz tests per chunk. The Go code lives at `goivy/` in the repo root. Module path: `github.com/glycerine/goivy`. Branch: `goport`.

---

## PART 1: COMPLETED WORK

### Package Structure (current)

```
goivy/                          # 6,495 lines across 24 files
├── go.mod                      # module github.com/glycerine/goivy, go 1.21
├── go.sum
├── logic/                      # Chunks 1-3: sorts, terms, formulas (1,361 lines)
│   ├── error.go                # IvyError, SortError
│   ├── node.go                 # Node interface + Sort→Node implementations
│   ├── sort.go                 # Sort interface + 6 concrete types
│   ├── term.go                 # Var, Const, Apply
│   ├── formula.go              # 15 formula types + True/False
│   ├── sort_test.go            # 12 tests + 2 fuzz
│   ├── term_test.go            # 15 tests + 1 fuzz
│   └── formula_test.go         # 25 tests + 1 fuzz
├── typeinfer/                  # Chunk 4: union-find type inference (912 lines)
│   ├── sortvar.go              # SortOrVar interface, SortWrapper, SortVar
│   ├── unify.go                # Find, Unify, OccursIn, ConvertFrom/ToSortVars
│   ├── infer.go                # InferSorts, ConcretizeSorts (all 15+ node types)
│   └── typeinfer_test.go       # 14 tests + 1 fuzz
├── logicutil/                  # Chunk 5: logic utilities (1,124 lines)
│   ├── logicutil.go            # FreeVariables, Substitute, EqualModAlpha, etc.
│   └── logicutil_test.go       # 22 tests + 1 fuzz
├── ivyutils/                   # Chunk 6: general utilities (647 lines)
│   ├── renamer.go              # UniqueRenamer, VariableGenerator, ConstantNameGenerator
│   ├── parameter.go            # Parameter, BooleanParameter, EnumeratedParameter, Registry
│   ├── graph.go                # TopologicalSort[T,K], Reachable[T,K], FindCycle[K]
│   ├── names.go                # ComposeNames, SplitName, BaseName, PolymorphicSymbols
│   └── ivyutils_test.go        # 27 tests + 2 fuzz
└── z3bridge/                   # Chunk 7: Z3 integration (1,233 lines)
    ├── quantifier.go           # Self-contained Z3 CGo wrapper (Context, Sort, Expr, FuncDecl, Solver, Model, ForAll, Exists)
    ├── translate.go            # Translator: Ivy logic.Node → Z3 Expr, Implies(), IsSat()
    └── z3bridge_test.go        # 22 tests (incl. Python z3_utils.py __main__ reproduction)
```

### Chunk Details

**Chunk 1-3: `logic/`** — Ported from `logic.py` (488 lines)
- `Sort` interface (embeds `Node`) with seal method; 6 concrete types: `UninterpretedSort`, `BooleanSort`, `FunctionSort`, `EnumeratedSort`, `RangeSort`, `TopSort`
- `Node` interface: `NodeSort()`, `Children()`, `String()`, `Equal(Node)`
- Terms: `Var` (uppercase name validation), `Const`, `Apply` (arity/sort validation)
- 15 formula types: `Eq`, `Ite`, `Not`, `Globally`, `Eventually`, `WhenOperator`, `Cond`, `And`, `Or`, `Implies`, `Iff`, `ForAll`, `Exists`, `Lambda`, `NamedBinder`
- `True = &And{}`, `False = &Or{}`
- Helpers: `SortEqual`, `FirstOrderSort`, `ContainsTopSort`, `IsPolymorphic`, `IsBooleanOrTop`

**Chunk 4: `typeinfer/`** — Ported from `type_inference.py` (419 lines)
- `SortOrVar` interface with `SortWrapper` (wraps `logic.Sort`) and `SortVar` (mutable union-find node)
- `Find()` with path compression, `Unify()` with occurs check, `OccursIn()`
- `ConvertFromSortVars()`, `ConvertToSortVars()`, `InsertSortVars()`
- `InferSorts()` handles all 15+ node types via type switch, returns `InferResult{Sort, Concretize}`
- `ConcretizeSorts()` — replaces TopSorts with concrete sorts
- Note: `collectNames()` is a simplified stand-in; should eventually use `logicutil.FreeVariables`

**Chunk 5: `logicutil/`** — Ported from `logic_util.py` (335 lines)
- `FreeVariables(t)` — by pointer identity; `FreeVariablesByName(t)` — by name
- `UsedVariables(t)`, `BoundVariables(t)`, `UsedConstants(t)`
- `Substitute(t, subs)` — simultaneous substitution with `CaptureError` detection
- `IsTautologyEquality(t)` — detects `Eq(x,x)`
- `EqualModAlpha(t, u)` — alpha-equivalence using `pushableMap`
- Convenience: `FreeVariablesList()`, `UsedConstantsList()`

**Chunk 6: `ivyutils/`** — Ported from `ivy_utils.py` (773 lines)
- `UniqueRenamer` — generates unique names with suffix collision avoidance
- `VariableGenerator` — generates A,B,...,Z,AA,BB,... (skipping O)
- `ConstantNameGenerator()` — yields a-z, a0-z0, a1-z1,...
- `Parameter`, `BooleanParameter`, `EnumeratedParameter` with `Registry` and `Parameterize`
- `TopologicalSort[T,K]`, `Reachable[T,K]`, `FindCycle[K]` (generics)
- `ComposeNames`, `SplitName`, `BaseName`, `ParentChildName`, `ExtractParametersName`
- `Distinct[T]`, `PolymorphicSymbols`

**Chunk 7: `z3bridge/`** — Ported from `z3_utils.py` (197 lines) + `ivy_solver.py` (1,716 lines)
- Self-contained Z3 CGo wrapper (NOT using go-z3, which lacks quantifier support)
- `Context` with thread-safe mutex, `Sort`, `Expr`, `FuncDecl` with proper ref counting via `runtime.SetFinalizer`
- Boolean ops: `Not`, `And`, `Or`, `Implies`, `Iff`, `Eq`, `Ite`
- **Quantifiers**: `ForAll`, `Exists` via `Z3_mk_forall_const`/`Z3_mk_exists_const`
- `Solver` with `Assert`, `Check`, `Push`, `Pop`, `Model`
- `Translator` — converts all Ivy `logic.Node` types to Z3 `Expr`
- `Implies(f1, f2)`, `IsSat(f)` convenience functions
- CGo flags: `-I/usr/local/opt/z3/include -L/usr/local/opt/z3/lib -lz3`

### Test Summary

**138 tests total, all passing, `go vet` clean:**
| Package | Tests | Fuzz | Lines |
|---------|-------|------|-------|
| logic | 52 | 4 | 934 |
| typeinfer | 14 | 1 | 250 |
| logicutil | 22 | 1 | 453 |
| ivyutils | 27 | 2 | 368 |
| z3bridge | 22 | 0 | 419 |
| **Total** | **137** | **8** | **2,424** |

### Key Design Decisions Made
1. `Sort` embeds `Node` (solves `Equal(Sort)` vs `Equal(Node)` conflict); `SortEqual()` helper
2. Python `recstruct.__iter__` → Go `Children()` returns only sub-fields, not meta fields
3. Python `frozenset(variables)` → Go sorted display in `varSortList()` for deterministic `String()`
4. Python `__call__` on Var/Const/NamedBinder → Go `Call()` methods
5. Z3 integration uses direct CGo (not go-z3) because go-z3 lacks ForAll/Exists
6. `newExpr`/`newSort`/`newFuncDecl` must be called with lock held (avoids deadlock from nested `ctx.do()`)

---

## PART 2: REMAINING PYTHON CODEBASE

### Dependency Graph (what depends on what)

```
ALREADY PORTED:
  logic.py → logic/
  type_inference.py → typeinfer/
  logic_util.py → logicutil/
  ivy_utils.py → ivyutils/
  z3_utils.py + ivy_solver.py → z3bridge/

NEXT TO PORT:
  ivy_lexer.py (304 lines) — PLY lexer, 47 tokens + 173 reserved keywords
       ↓
  ivy_ast.py (1,965 lines) — 157 AST node classes
       ↓
  ivy_logic_parser.py (655 lines) — formula sub-parser
       ↓
  ivy_parser.py (3,161 lines) — 279 grammar rules (PLY yacc)
       ↓
  ivy_logic.py (1,774 lines) — higher-level IR, Symbol/Sig management
       ↓
  ivy_actions.py (1,687 lines) — 49 action classes (imperative semantics)
       ↓
  ivy_transrel.py (669 lines) — transition relation composition
       ↓
  ivy_module.py (412 lines) — module context/definitions
       ↓
  ivy_compiler.py (2,320 lines) — AST → logic IR compilation
       ↓
  ivy_isolate.py (2,022 lines) — modular verification
       ↓
  ivy_check.py (1,041 lines) — top-level verification
       ↓
  ivy_proof.py (1,653 lines) — proof state/tactics
       ↓
  ivy_mc.py (1,772 lines) — model checking
       ↓
  Code generators (ivy_to_cpp.py 6,715 lines, etc.)
       ↓
  UI/Visualization (ivy_ui*.py, tk_*.py, cy_*.py — ~7,000 lines)
```

### File Inventory (unported, by priority)

**Tier 1 — Language Frontend (~6,085 lines):**
| File | Lines | Role |
|------|-------|------|
| `ivy_lexer.py` | 304 | Tokenizer (47 tokens, 173 keywords, version-gated) |
| `ivy_ast.py` | 1,965 | 157 AST node classes (Formula, Decl, Action, Sort hierarchies) |
| `ivy_logic_parser.py` | 655 | Formula sub-parser (terms, atoms, quantifiers) |
| `ivy_parser.py` | 3,161 | Main parser (279 grammar rules, modules/actions/declarations) |

**Tier 2 — Core Semantics (~4,542 lines):**
| File | Lines | Role |
|------|-------|------|
| `ivy_logic.py` | 1,774 | Higher-level logic IR, Symbol/Sig, sort inference |
| `ivy_actions.py` | 1,687 | 49 action classes (Assume, Assert, Assign, If, While, Call, etc.) |
| `ivy_transrel.py` | 669 | Transition relations, state versioning, interpolation |
| `ivy_module.py` | 412 | Module context storage |

**Tier 3 — Compiler & Verification (~7,036 lines):**
| File | Lines | Role |
|------|-------|------|
| `ivy_compiler.py` | 2,320 | AST → logic IR transformation |
| `ivy_isolate.py` | 2,022 | Modular verification strategies |
| `ivy_check.py` | 1,041 | Top-level verification checker |
| `ivy_proof.py` | 1,653 | Proof state management and tactics |

**Tier 4 — Advanced Analysis (~3,219 lines):**
| File | Lines | Role |
|------|-------|------|
| `ivy_mc.py` | 1,772 | Model checking engine |
| `ivy_ranking.py` | 1,195 | Termination/ranking analysis |
| `ivy_theory.py` | 181 | Built-in theories (int arithmetic) |
| `ivy_interp.py` | 642 | Symbolic interpreter |
| `ivy_art.py` | 523 | Abstraction Refinement Tree (CEGAR) |

**Tier 5 — Code Generation (~8,375 lines):**
| File | Lines | Role |
|------|-------|------|
| `ivy_to_cpp.py` | 6,715 | C++ code generation |
| `ivy_cpp_types.py` | 525 | C++ type system |
| `ivy_cpp.py` | 414 | C++ compilation utilities |
| `ivy_dafny_compiler.py` | 478 | Dafny output |
| Others | ~243 | Lean, Markdown outputs |

**Tier 6 — UI/Visualization (~7,000+ lines):** Not a porting priority.

---

## PART 3: NEXT CHUNKS — DETAILED PLANS

### Chunk 8: `ast/` — AST Node Types

**Source**: `ivy_ast.py` (1,965 lines, 157 classes)

**Files to create**: `goivy/ast/ast.go`, `goivy/ast/formula.go`, `goivy/ast/decl.go`, `goivy/ast/action.go`, `goivy/ast/sort.go`, `goivy/ast/ast_test.go`

This is a separate AST from `logic/` — it represents the *parsed syntax tree* before compilation to the logic IR. The `logic/` package represents the *semantic* IR.

**Key base types:**
```go
// AST is the base for all syntax tree nodes (has line number info)
type AST struct {
    Lineno  Location    // source location
    Args    []ASTNode   // sub-expressions (recstruct pattern)
}

type ASTNode interface {
    GetLineno() Location
    GetArgs() []ASTNode
    String() string
}
```

**Formula hierarchy** (AST-level, NOT logic-level):
- `Symbol{Name string}` — identifier
- `Atom{Relname string, Terms []ASTNode}` — atomic formula/predicate
- `App{Func ASTNode, Args []ASTNode}` — function application
- `Variable{Name string, Sort ASTNode}` — variable with sort annotation
- `And`, `Or`, `Not`, `Implies`, `Iff`, `Ite` — logical connectives
- `Forall`, `Exists` — quantifiers with bounds
- `Globally`, `Eventually` — temporal operators
- `NamedBinder{Name string, Params []ASTNode, Body ASTNode}`
- `Old{Term ASTNode}` — temporal "old" operator
- `This{}` — self-reference

**Declaration hierarchy:**
- `ModuleDecl{Name, Args, Body}` — module definition
- `ObjectDecl{Name, Body}` — object definition
- `ActionDecl{Name, Params, Returns, Body}`
- `TypeDecl{Name, Sort}`, `VariantDecl`
- `RelationDecl{Name, Params}`, `ConstantDecl{Name, Sort}`
- `AxiomDecl`, `PropertyDecl`, `ConjectureDecl` — labeled formulas
- `InitDecl`, `ExportDecl`, `ImportDecl`, `IsolateDecl`
- `NativeDecl{Code string}`, `AttributeDecl`
- `MixinDecl{Kind string}` — before/after/implement mixins
- `InterpretDecl`, `DefinitionDecl`

**Action AST nodes** (syntax level, distinct from `ivy_actions.py` semantic actions):
- `AssignAction{Lhs, Rhs ASTNode}`
- `CallAction{Callee ASTNode, Args []ASTNode}`
- `IfAction{Cond, Then, Else ASTNode}`
- `WhileAction{Cond, Body ASTNode, Invariant ASTNode}`
- `Sequence{Actions []ASTNode}`
- `AssumeAction{Formula ASTNode}`, `AssertAction{Formula ASTNode}`
- `LocalAction{Vars []ASTNode, Body ASTNode}`

**AST Sort nodes:**
- `ConstantSort{Name string}` — uninterpreted
- `EnumeratedSort{Name string, Values []string}`
- `StructSort{Name string, Fields []ASTNode}`
- `FunctionSort{Domain []ASTNode, Range ASTNode}`
- `RelationSort{Domain []ASTNode}`

**Root node:**
- `Ivy{Body []ASTNode}` — top-level program, contains all declarations

**Tests**: Construction of each node type, String() output, Children/Args traversal.

### Chunk 9: `lexer/` — Hand-Written Lexer

**Source**: `ivy_lexer.py` (304 lines)

**Files to create**: `goivy/lexer/token.go`, `goivy/lexer/lexer.go`, `goivy/lexer/lexer_test.go`

**Design**: Hand-written lexer (no PLY equivalent in Go). Simple `Lexer` struct with `NextToken()`.

**token.go:**
```go
type TokenType int
const (
    // Punctuation
    LPAREN TokenType = iota; RPAREN; LCB; RCB; LB; RB
    COMMA; SEMI; COLON; DOT; DOTS; DOTDOTDOT; ARROW; ASSIGN
    // Operators
    PLUS; MINUS; TIMES; DIV; EQ; TILDAEQ; TILDA; LE; LT; GE; GT
    AND; OR; IFF; DOLLAR; CARET; PTO
    // Literals/identifiers
    SYMBOL      // lowercase identifier or quoted string
    VARIABLE    // uppercase identifier
    NATIVEQUOTE // <<<...>>>
    // Special
    EOF; NEWLINE; ERROR
)

type Token struct {
    Type   TokenType
    Value  string
    Line   int
    Column int
}
```

**Reserved keywords** (173 total, version-gated):
- Core: `relation`, `individual`, `function`, `axiom`, `conjecture`, `action`, `init`, `module`, `object`, ...
- Version 1.4+: `struct`, `ghost`, `alias`, `variant`, `of`, `constructor`
- Version 1.6+: `specification`, `implementation`, `parameter`, `tactic`
- Version 1.7+: `debug`, `for`, `process`, `subclass`, `template`

**lexer.go:**
```go
type Lexer struct {
    input   string
    pos     int
    line    int
    col     int
    version [2]int  // e.g., [1, 7]
}

func NewLexer(input string, version string) *Lexer
func (l *Lexer) NextToken() Token
func (l *Lexer) Peek() Token
```

Key behaviors:
- Identifiers starting with uppercase → `VARIABLE`
- Identifiers starting with lowercase → check reserved words → `SYMBOL` or keyword token
- Quoted strings `"..."` → `SYMBOL`
- Native code `<<<...>>>` → `NATIVEQUOTE`
- Comments `#...` → skip to end of line
- Multi-character operators: `->`, `<->`, `~=`, `*>`, `..`, `...`
- Version-gated keywords (e.g., `struct` is only a keyword in version ≥ 1.4)

**Tests**: Token-by-token verification for sample inputs, version-gated keyword tests, edge cases (empty input, unterminated strings, native quotes).

### Chunk 10: `parser/` — Hand-Written Recursive Descent Parser

**Source**: `ivy_parser.py` (3,161 lines, 279 grammar rules) + `ivy_logic_parser.py` (655 lines)

**Files to create**: `goivy/parser/parser.go`, `goivy/parser/expr.go`, `goivy/parser/decl.go`, `goivy/parser/action.go`, `goivy/parser/parser_test.go`

**Design**: Hand-written recursive descent with Pratt parsing for expressions. No PLY/yacc.

**parser.go:**
```go
type Parser struct {
    lexer   *lexer.Lexer
    current lexer.Token
    peek    lexer.Token
    errors  []ParseError
    version [2]int
}

func NewParser(input string, version string) *Parser
func (p *Parser) Parse() (*ast.Ivy, error)
```

**expr.go** — Expression/formula parsing (Pratt/precedence climbing):
```
Precedence (low to high):
1. SEMI (sequence in actions)
2. GLOBALLY, EVENTUALLY, temporal operators
3. IF/ELSE
4. OR (|)
5. AND (&)
6. TILDA (~, NOT)
7. Comparisons (=, <, <=, >, >=, ~=, *>)
8. COLON (sort annotation)
9. PLUS, MINUS
10. TIMES, DIV
11. DOLLAR ($, named binders)
12. DOT (field access)
13. Primary (atoms, parenthesized, quantifiers)
```

Key expression rules:
- `parseExpr(minPrec)` — Pratt parser core
- `parsePrimary()` — atoms, variables, parenthesized, quantifiers
- `parseQuantifier()` — `forall`/`exists` with variable lists
- `parseApp()` — function application `f(x, y)`
- `parseSortAnnotation()` — `X:Sort`

**decl.go** — Top-level declarations:
- `parseTopLevel()` — dispatch on keyword
- `parseModule()`, `parseObject()`, `parseClass()`
- `parseAction()`, `parseInit()`
- `parseRelation()`, `parseFunction()`, `parseIndividual()`
- `parseAxiom()`, `parseProperty()`, `parseConjecture()`
- `parseType()`, `parseVariant()`
- `parseIsolate()`, `parseExport()`, `parseImport()`
- `parseInterpret()`, `parseDefinition()`
- `parseNative()`, `parseAttribute()`

**action.go** — Action body parsing:
- `parseActionBody()` — sequence of statements
- `parseStatement()` — dispatch on keyword/pattern
- `parseIf()`, `parseWhile()`
- `parseAssign()` — `lhs := rhs`
- `parseCall()` — `call action(args)`
- `parseLocal()`, `parseLet()`
- `parseAssume()`, `parseAssert()`

**Tests**: Parse sample Ivy programs, verify AST structure. Key test cases:
- Simple declarations: `type t`, `relation r(X:t, Y:t)`
- Axioms: `axiom forall X:t. r(X, X)`
- Actions: `action a(x:t) = { ... }`
- Modules: `module m(t) = { ... }`
- Full programs from Ivy test suite

### Chunk 11: `ivylogic/` — Higher-Level Logic IR

**Source**: `ivy_logic.py` (1,774 lines)

**Files to create**: `goivy/ivylogic/sig.go`, `goivy/ivylogic/symbol.go`, `goivy/ivylogic/ivylogic.go`, `goivy/ivylogic/ivylogic_test.go`

This wraps `logic/` with:
- `Sig` — signature maintaining all symbols, sorts, and their relationships
- `Symbol` — higher-level constant/relation with name mangling
- Sort inference and resolution
- Formula classification: `IsQF()`, `IsPrenexUniversal()`, etc.
- `WithSymbols`, `WithSorts` — context-like sort/symbol scoping
- `VariableUniqifier` — unique variable renaming
- `PolySymsDict` — polymorphic symbol resolution

### Chunk 12: `actions/` — Action Semantics

**Source**: `ivy_actions.py` (1,687 lines, 49 classes)

Action class hierarchy (semantic level, post-compilation):
- `Action` base with `UpdateAction`, `AssumeAction`, `AssertAction`
- `AssignAction`, `HavocAction`, `SetAction`
- `IfAction`, `WhileAction`, `ChoiceAction`, `Sequence`
- `CallAction`, `LocalAction`, `LetAction`
- `NativeAction`, `CrashAction`, `ThunkAction`
- Type checking: `TypeCheckAction()`, `HasCode()`

### Chunk 13: `transrel/` — Transition Relations

**Source**: `ivy_transrel.py` (669 lines)

- State variable versioning (`new()`, `old()`, `rename()`)
- `ComposeUpdates()`, `JoinAction()`, `IteAction()`
- Forward interpolation and interpolant computation
- State-action composition

### Chunk 14: `compiler/` — AST → Logic IR

**Source**: `ivy_compiler.py` (2,320 lines)

- AST node → logic node compilation
- Type inference during compilation
- Field reference resolution
- Module instantiation with parameter substitution
- Context managers for scoping

---

## PART 4: IMPLEMENTATION STRATEGY

### Recommended Order

1. **Chunk 8: `ast/`** — AST node types (no parsing yet, just data structures). ~800 lines Go.
2. **Chunk 9: `lexer/`** — Hand-written tokenizer. ~400 lines Go.
3. **Chunk 10: `parser/`** — Recursive descent parser producing AST. ~1,500 lines Go.
4. **Chunk 11: `ivylogic/`** — Higher-level IR wrapping `logic/`. ~800 lines Go.
5. **Chunk 12: `actions/`** — Action semantics. ~700 lines Go.
6. **Chunk 13: `transrel/`** — Transition relations. ~400 lines Go.
7. **Chunk 14: `compiler/`** — AST → logic compilation. ~1,200 lines Go.

### Estimated Total: ~5,800 lines of Go (excluding tests)

### Verification After Each Chunk
1. `go build ./...` — compiles
2. `go test ./... -v` — all unit tests pass
3. `go test ./... -fuzz=. -fuzztime=30s` — fuzz tests find no crashes
4. `go vet ./...` — no warnings
5. Cross-reference with Python where applicable

### Key Python Source Files (for reference during implementation)
- `/Users/jaten/pyivy/ivy/ivy/ivy_ast.py` — AST node definitions
- `/Users/jaten/pyivy/ivy/ivy/ivy_lexer.py` — token definitions and lexer
- `/Users/jaten/pyivy/ivy/ivy/ivy_parser.py` — 279 grammar rules
- `/Users/jaten/pyivy/ivy/ivy/ivy_logic_parser.py` — formula sub-parser
- `/Users/jaten/pyivy/ivy/ivy/ivy_logic.py` — higher-level logic IR
- `/Users/jaten/pyivy/ivy/ivy/ivy_actions.py` — action semantics
- `/Users/jaten/pyivy/ivy/ivy/ivy_transrel.py` — transition relations
- `/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py` — AST → logic compilation
- `/Users/jaten/pyivy/ivy/ivy/ivy_module.py` — module context

# PROGRESS UPDATE 2

