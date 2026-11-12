# Mechanical Go porting ideas

Working backlog for starting-point, based on inspecting the current `out/` tree.
Inspected 2026-11-11. Paths and line numbers below refer to that snapshot; they
will shift after regeneration. Existing generated output was not edited.

The goal remains readable draft Go with comments and source structure intact.
Prefer direct syntax changes and known types. Leave compiler-visible mismatches
for the next pass; do not add anonymous-function adapters. No Clang dependency.

## Work included with this inspection

- [x] Read `./starting.cfg` automatically; allow `-cfg PATH`. Each setting is
  `NAME = INTEGER`. Missing default is fine; explicit missing files and malformed
  settings are errors. Example: `WIN32 = 0` disables WIN32.
- [x] Use configured choices to select decidable `#if`/`#ifdef`/`#ifndef` chains,
  including `#elif`, `#elseif`, `#else`, and nested selected branches. Keep
  unresolved choices opaque. Preserve original byte/line offsets internally.
- [x] Comment pthread.h includes even inside retained opaque conditional regions.
- [x] Convert standalone file-scope enum/typedef enum definitions to `const (...)`.
  Use `iota` for all-implicit sequences; use explicit values and `Previous + 1`
  after explicit resets. Named enums retain an integer type; comments survive.

Evidence: `out/starter.go:38` has `#ifdef WIN32`, and
`out/aig/aig/aig.go:59` has an ordinary `typedef enum` defining AIG object kinds.

## Implementation status: prototype 52

The mechanical rules below are now implemented within these explicit boundaries.
The original examples remain useful as regression targets.

- [x] Config branch selection, `-cfg`, inline `//` comments, and automatic
  `_starting.cfg` discovery when the default config is absent. The template
  lists discovered non-guard conditional symbols as `SYMBOL = 0`, with source
  directive comments. It is inactive and never overwrites an existing template.
- [x] Literal loop/if conditions, switch header cleanup, adjacent empty case
  grouping, and explicit fallthrough between C cases that reach the next case.
- [x] Scalar single-token casts (float/double/int/unsigned/char/short/long), plus
  integer literal suffixes as typed Go conversions. Pointer casts remain review.
- [x] Simple prefix increments/decrements at the start of an if condition become
  Go if initializers. Member/index operands and postfix expression uses remain.
- [x] Shared typedef/enum-tag name discovery across tree and flat-batch inputs.
- [x] One-object ABC_CALLOC and ABC_ALLOC become new(T). Array allocations,
  reallocation and freeing remain explicitly reported ownership decisions.
- [x] Ordinary bitfields expand to integer fields with width/layout/truncation
  comments; no claim of C binary-layout equivalence is made.
- [x] Pointer-member spelling is normalized in generated macro bodies as well
  as ordinary functions, while original comments/opaque source remain intact.
- [x] Discarded comma-expression statements become ordered statements. Independent
  simple loop initializers/post assignments become Go multi-assignment. Dependent
  assignments and arbitrary multi-operation post clauses remain for repair.
- [x] Braced do/while loops with no continue become bottom-tested Go loops.
  Continue-bearing bodies are retained with specific control-flow review guidance.
- [x] Known integer/pointer conditions become comparisons against zero/nil,
  including assert arguments and known function returns. Ambiguous names stay C.
- [x] Direct assert(pthread_mutex_lock/unlock(...) == 0) becomes the method call
  once, without an impossible check against a nonexistent Go return value.
- [x] Known aggregate `{0}` initialization is omitted; explicit array/aggregate
  lists become typed composite literals. Array/comma struct fields are reordered.
- [x] Recognized local-static declarations lift to uniquely named package variables
  so state persists across calls. Dynamic/lazy C++ initializers, addresses of
  automatic variables, and ambiguous shadowing remain for repair.
- [x] Ordinary file/local enum definitions, tag references and separate tag aliases
  convert; explicit resets and implicit successors retain their values. C++
  scoped enums and trailing variable declarators remain unsupported.
- [x] Expression macros return values with an int result placeholder unless a
  result type is supplied. Whole-expression parentheses are removed; statement-
  style macros use a marked return 0 for review.
- [x] Whole-object memset(p, 0, sizeof(T)) becomes *p = T{} only when p is a known
  pointer to a known struct of exactly that type. Unknown layouts remain intact.
- [x] Union, pointer-cast, sizeof, allocation, thread and unsupported control-flow
  cases have targeted review metadata instead of invented representations.
- [x] Function metadata uses the actual emitted function text, including renamed
  static variables, to prevent per-function/file-wide rewrite disagreement.

Validation uses **only** the library and CLI packages:
`go -C starting-point test . ./cmd/starting-point`. Experiments are excluded.
Runtime regressions combine rules, including a configured multi-file translation
with header typedefs, enums, bitfields, arrays, allocation, assertions, static
state, switch fallthrough, and comma loop clauses. Boundary tests protect comments,
opaque branches, shadowed names, nested breaks, increment operands, and filenames.

## Implemented small improvements and their boundaries

| Priority | Pattern and observed example | Proposed rule | Boundary / verification |
|---|---|---|---|
| 1 | `for 1 {` in `out/starter.go:150` and `:173` | Emit `for {` for literal true loop conditions; translate literal numeric if conditions to boolean constants. | Keep expression side effects; test zero and nonzero literals separately. |
| 2 | `switch ( nAlgo ) {` in `out/aig/aig/aigPartSat.go:494` | Reuse header-parenthesis removal and brace placement for switch headers. | This is syntax cleanup only; do not silently erase C fallthrough. |
| 3 | `(int)` casts in iterator definitions at `out/aig/aig/aig.go:227` | Extend the float-cast operand parser to known scalar C types: `int32(expr)`, `uint32(expr)`, etc. | Preserve unary/binary precedence and integer widths. Pointer casts stay separate. |
| 4 | `if ++nLines == MAX_COMM_NUM` at `out/starter.go:169` | Lift a simple prefix increment into a preceding statement, then test the variable. | Only standalone if conditions initially; `else if`, short-circuit conditions and postfix increments need extra care. |
| 5 | `abctime clk;` at `out/aig/aig/aigPartSat.go:490` | Build a lightweight shared typedef/name index from the input inventory and reuse it for declaration recognition. | Prevent lowercase typedef names from being confused with multiplication. No include expansion needed for an initial name-only index. |
| 6 | `p = ABC_CALLOC(Xyz_ManTry_t, 1)` at `out/template.go:61` | Translate this narrowly identifiable single-object allocation to `p = new(Xyz_ManTry_t)`. | Do not map arbitrary `ABC_ALLOC(T, n)` to a slice until uses establish the representation. |
| 7 | `unsigned int Type : 3` at `out/aig/aig/aig.go:78` | Expand bitfields to ordinary integer fields and retain the bit width in a comment. | This loses C packing and truncation semantics; clearly mark these as representation-review sites. Avoid booleanizing 1-bit integer fields automatically. |
| 8 | `p->nCutsMax` remains inside a generated iterator function at `out/aig/aig/aig.go:221` | Apply token-level member-access normalization consistently to generated macro replacement bodies. | Do not alter saved original-macro comments, strings, or opaque conditional text. |

## Structural rules and remaining boundaries

### Comma expressions and loop clauses

`out/aig/aig/aig.go:221` has both `i = 0, pCut = ...` and
`i++, pCut = ...` inside one for header. Pure assignment lists can sometimes
become Go multi-assignment; dependent assignments must remain ordered statements.
A general comma-expression statement can become multiple statements without a
closure. A loop post clause with multiple operations needs careful rewriting of
`continue`, or a small loop-shape transformation. Keep iteration order visible.

Start with comma-separated standalone calls (such as ABC_PRT), then simple
independent assignments. Do not rewrite all commas indiscriminately: calls,
declarators and composite literals use them differently.

### Do/while loops

`out/aig/aig/aigTsim.go:71` contains a simple rejection-sampling loop:

```c
do { RetValue = Aig_ManRandom(0) & 3; } while (RetValue == 0);
```

For a body with no continue targeting this loop, emit a `for { ...; if !cond {
break } }`. A continue in C still evaluates the condition; blindly moving the
condition to the bottom of a Go loop would skip it. Restrict the first rule or
introduce an explicit condition label, without changing nested-loop control flow.

### Numeric and pointer truth values

`out/aig/aig/aigPartSat.go:513` uses `if fVerbose`; the parameter is an integer.
Use known declaration/signature types to produce `fVerbose != 0`, and pointers
as `p != nil`. Apply to simple negations and assert arguments as well. Avoid
heuristics that turn a boolean-returning call into `call() != 0`. A type-name
index plus a modest function-return index would make this much less ambiguous.

### Calls whose C status no longer exists

`out/starter.go:151` contains `assert(mutex.Lock() == 0, ...)` after the current
mutex conversion. A narrowly recognized original `assert(pthread_mutex_lock(...)
== 0)` can become just the Lock call: Go's method has no status return.
Likewise for Unlock. This is a structural rule around a known method, not a
closure or a general deletion of assertions. Evaluate the receiver exactly once.

### Brace initializers and local static storage

`out/aig/aig/aigCanon.go:101` contains
`static Bdc_Par_t Pars = {0}, * pPars = &Pars;`, and line 152 has a static array.
These combine multiple missing rules:

- Recognize `{0}` as zero initialization for a known aggregate.
- Convert C array initializer lists to Go composite literals once the type is known.
- Split mixed pointer/nonpointer declarators while preserving initialization order.
- Preserve local-static lifetime, possibly with a uniquely named package variable.

Simply deleting `static` inside a function would reset storage on each call and
change behavior. Implement zero/composite initialization separately first.

### Further enum cases

The new pass handles ordinary file-scope definitions. Follow-ups include local
enums, separate `typedef enum Tag Alias;`, `enum Tag` references in signatures,
C++ scoped enums, and enums with trailing variable declarators. Normalize integer
literal suffixes in enum expressions. Review values outside int32 before fixing
an underlying type; C enum representation is not universally signed int32.

### Switch fallthrough and grouped cases

After header cleanup, combine adjacent empty C case labels into Go case lists.
Where a C case actually falls through, insert Go `fallthrough` only when it is
legal and matches control flow. Leave explicit breaks initially; many are legal
Go and deleting them adds little value compared with preserving behavior.

### Common macro returns

Many function-like macros are expression macros, but their generated functions
have no result. Recover obvious returns for a single expression when a supplied
or inferred result type is available; otherwise add a useful repair marker.
Iterator macros are a different problem: their caller-provided body cannot be
recovered merely by returning a value. Recognize iterator call sites separately.

## Representation-heavy work to keep explicit

- **Allocation and memory:** `out/aig/aig/aig.go:370` uses
  `memset(pObj, 0, sizeof(Aig_Obj_t))`. If a destination is known to point to one
  complete struct, `*pObj = Aig_Obj_t{}` is a plausible rule. Arbitrary byte counts,
  unions and partial-object clearing are not equivalent.
- **Unions:** `out/aig/aig/aig.go:73` overlays a pointer and an integer. Expanding
  these into independent fields makes syntax easier but changes aliasing and
  layout. Offer a documented policy/annotation rather than claiming equivalence.
- **Pointer-tagged AIG edges:** keep a representation decision explicit. Do not
  translate pointer-to-integer casts into ordinary numeric conversions.
- **Char pointers:** `(char *)Buffer` at `out/starter.go:162` may become a string
  conversion only after deciding NUL termination and byte-buffer ownership.
- **Threads:** `pthread_create` at `out/starter.go:168` suggests goroutines, but
  argument lifetime, termination, join behavior and error handling need a
  coordinated change. Mutex mapping alone does not establish these choices.
- **Unused locals/imports:** the compiler can identify these after syntax repair.
  Avoid deleting side-effecting initializers just to quiet a diagnostic.

## Scope indicators from the inspected output

These are approximate matching **line counts**, not independent defects or
promises of automatic conversion. The scan skipped lines beginning with `//`,
but can still include block comments, strings, macro drafts, and opaque C.

| Candidate pattern | Matching lines |
|---|---:|
| C-style switch headers | 764 |
| do blocks | 369 |
| sizeof calls | 2,219 |
| selected C scalar/pointer cast spellings | 8,206 |
| ternary-looking expressions | 4,176 |
| ForEach-style iterator names | 14,203 |
| ABC allocation/free macro calls | 5,865 |
| enum keyword | 132 |
| union keyword | 105 |

Next work requires choosing representations or expanding the documented rule
boundaries: continue-aware do/while, dependent comma loop clauses, callback/iterator
macro expansion, scoped C++ enums, unions, pointer tagging, array ownership, and
thread lifecycle. These are not silently marked as mechanically solved. Keep
unused-local cleanup compiler-guided and retain side-effecting initializers.
