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
