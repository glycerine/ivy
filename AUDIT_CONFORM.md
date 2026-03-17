# AUDIT_CONFORM.md — Python↔Go Conformance Audit

Generated: 2026-03-17

This document catalogs every structural and semantic divergence between
Python Ivy (`~/pyivy/ivy/ivy/`) and Go goivy (`~/go/src/github.com/glycerine/goivy/`).
Each finding includes a description paragraph explaining how to bring
the Go into conformance with the Python.

---

## 1. `logic.py` vs `logic/sort.go`, `logic/term.go`, `logic/formula.go`

### 1.1 `recstruct` immutability vs Go mutable structs

**Python**: All logic types use `recstruct`, which produces **immutable, hashable, structurally-equal** tuple-like objects. A `Const("f", S)` can be a dict key, set member, or frozenset element. Two instances with the same fields ARE the same object for purposes of `==` and `hash`.

**Go**: All logic types are mutable pointer structs. Equality is via explicit `.Equal()` method, not structural. Hashing is via pointer identity (or custom). `*Const` cannot be a map key by value.

**Impact**: Any Python code that puts logic nodes into sets, frozensets, or dict keys relies on structural equality. Go equivalents must use explicit map-by-name or sorted-slice workarounds. The `ForAll.Variables` and `Exists.Variables` fields are particularly affected (see §1.2).

**How to conform**: This is a fundamental type-system difference that cannot be changed wholesale. Instead, audit every place where Python uses `set()`, `frozenset()`, or `dict()` with logic nodes as keys, and verify the Go equivalent uses appropriate key types (usually string names, or custom comparison).

---

### 1.2 `ForAll.Variables` / `Exists.Variables`: `frozenset` vs `[]*Var` slice

**Python** (logic.py:380): `ForAll._preprocess_` returns `frozenset(variables), body`. Variables are stored as a **frozenset** — unordered, deduplicated, hashable. This means `ForAll([X, Y], body)` and `ForAll([Y, X], body)` produce the **same object** (same hash, same equality).

**Go** (formula.go:339): `NewForAll` copies variables into a `[]*Var` **slice** — ordered, may contain duplicates. `ForAll([X, Y], body)` and `ForAll([Y, X], body)` are **different** because slice order matters.

**Impact**: Any code that compares ForAll/Exists nodes, or that relies on variable order independence, will behave differently. The `Equal()` method compares slice-by-element, so reordered variables will produce inequality. The `String()` method sorts by name (`varSortList`), so string output matches, but semantic equality does not.

**How to conform**: Change `NewForAll` and `NewExists` to sort the variables slice by name during construction. This ensures that `[X, Y]` and `[Y, X]` produce the same internal representation. Also, deduplicate by name. Alternatively, change `Equal()` to compare as sets.

---

### 1.3 `EnumeratedSort.__str__`: `{red,green,blue}` vs `self.name`

**Python** (logic.py:58): `EnumeratedSort.__str__` returns `'{' + ','.join(self.extension) + '}'` — the extension elements.

**Python** (logic.py:58, but actually see ivy_logic.py's monkey-patching): **Wait** — logic.py:58 shows `'{' + ','.join(...)` but `ivy_logic.py` may override this. Let me check.

Actually, looking at logic.py:57-58 more carefully: the `__str__` IS the extension format. But note the commented-out line at logic.py:58: `# return self.name`. The live code returns extensions.

**Go** (sort.go:81-82): `EnumeratedSort.String()` returns `"{" + strings.Join(s.Extension, ",") + "}"` — matches the Python live code.

**Status**: CONFORMANT. No action needed.

---

### 1.4 `Apply.__str__`: comma-space separator

**Python** (logic.py:179): `', '.join(str(t) for t in self.terms)` — uses `", "` (comma + space).

**Go** (term.go:140): `strings.Join(parts, ",")` — uses `","` (comma, no space).

**However**: Python's `ivy_logic.py:1434` monkey-patches `__str__` to `pretty_fmla` which calls `self.ugly(0)`. The `ugly` method for Apply is NOT defined in the same list (line 1303-1319 defines ugly for Var, Const, Eq, And, Or, Not, Implies, Iff, ForAll, Exists, Lambda, NamedBinder — but NOT Apply).

For Apply, `ugly` falls back to the `__str__` of `logic.py:177` via `pretty_fmla → drop_annotations → ugly`. But `Apply` doesn't have an `ugly` method defined, so when `pretty_fmla` calls `d.ugly(0)`, it falls through to `Apply.__str__` which uses `', '.join(...)` — comma+space.

**Impact**: The Go Apply.String() uses no-space commas but the active Python uses comma+space. This will cause string mismatches in conformance testing.

**How to conform**: Change Go `Apply.String()` back to `", "` separator, matching Python's actual Apply formatting. The change to `","` we made earlier was wrong — it happened to work for the test case because `ivy_logic.py`'s `pretty_fmla` path produces `", "` for Apply too. **UPDATE**: Actually, need to verify by running Python and checking exact output.

---

### 1.5 `Apply.sort` property: `TopS` for TopSort func vs cached `aSort`

**Python** (logic.py:182-183): `sort = property(lambda self: TopS if isinstance(self.func.sort, TopSort) else self.func.sort.range)` — computed dynamically each time. If `func.sort` changes (possible since Python recstructs are immutable, but the reference could be replaced), the sort updates.

**Go** (term.go:73,84,113): `aSort Sort` is cached at construction time. If TopSort func → `aSort = TopS`. If FunctionSort func → `aSort = fs.Range()`.

**Impact**: In practice, immutable Python objects mean the property never changes. The Go cache is equivalent. **But**: if any code creates an Apply with a bare Func pointer and later expects `NodeSort()` to reflect changes to the Func's sort, it won't work in Go.

**Status**: Effectively conformant for correct usage. No action needed.

---

### 1.6 `is_polymorphic()`: Const name check divergence

**Python** (logic.py:108): `type(x) == Const and not x.name[0].islower()` — a Const is polymorphic if its first character is NOT lowercase. This means uppercase, underscore, digit, or special characters are all "polymorphic". Note: this uses `islower()` which returns False for non-alpha characters.

**Go** (sort.go:164): `len(c.Name) > 0 && !isLower(c.Name[0])` where `isLower(b) = b >= 'a' && b <= 'z'`. This is equivalent — non-lowercase-ASCII is treated as polymorphic.

**Impact**: Characters like `_`, `0-9`, Unicode letters: Python's `islower()` returns False for `_` and digits (matches Go). For Unicode, Python returns True for lowercase Unicode letters but Go only checks ASCII. This could differ for non-ASCII symbol names.

**How to conform**: If Ivy symbol names are always ASCII (which they appear to be), this is conformant. If Unicode names are possible, use `unicode.IsLower()` in Go.

---

### 1.7 `EnumeratedSort.String()` name vs extension

**Python** (logic.py:57-58): Returns `'{' + ','.join(self.extension) + '}'` — the extensions.
But **ivy_logic.py** monkey-patches: at line 228, `EnumeratedSort.__str__` is overridden to `return self.name`.

**Go** (sort.go:81-82): Returns the extension format `{red,green,blue}`.

**Impact**: When `ivy_logic.py` is loaded (which is always the case in practice), Python's `str(EnumeratedSort)` returns the **name** (e.g., `"color"`), not the extension. Go returns the extension. This causes string mismatches everywhere enumerated sorts appear in printed output.

**How to conform**: Change Go `EnumeratedSort.String()` to return `s.Name` (matching the monkey-patched Python behavior). The extension format is only used in the base `logic.py` which is never actually used standalone.

---

### 1.8 `Lambda.sort`: computed FunctionSort vs hardcoded Boolean

**Python** (logic.py:407): `Lambda.sort = Boolean` — hardcoded as Boolean.

**Go** (formula.go:410): `func (l *Lambda) NodeSort() Sort { return Boolean }` — also hardcoded as Boolean.

**Status**: CONFORMANT. Both return Boolean. (Note: this seems semantically wrong — a Lambda should have a FunctionSort from var sorts to body sort — but both Python and Go agree on this behavior.)

---

### 1.9 `NamedBinder.sort`: computed FunctionSort — Go matches Python

**Python** (logic.py:437-442): Dynamic property computing `FunctionSort(v.sort for v in vars, body.sort)` if vars non-empty, else `body.sort`.

**Go** (formula.go:442-457): Same computation via `NewFunctionSort(sorts...)`.

**Status**: CONFORMANT.

---

### 1.10 `Ite` sort field: `sort` metadata field vs `ISort`

**Python** (logic.py:206): `Ite(recstruct('Ite', ['sort'], ['cond', 't_then', 't_else']))` — `sort` is a metadata field (non-child). The `__init__` sets it to `t_then.sort`. Being a metadata field means `sort` is NOT iterated by `for x in self` (which iterates only children).

**Go** (formula.go:48): `ISort Sort` field, set to `then_.NodeSort()` in `NewIte`. `Children()` returns `[Cond, Then, Else]` — excluding `ISort`. Matches Python's separation of metadata from children.

**Status**: CONFORMANT.

---

### 1.11 `WhenOperator.sort`: metadata field preservation

**Python** (logic.py:268-271): `WhenOperator.__init__(self, name, t1, t2)` calls `super().__init__(t1.sort, name, t1, t2)`. The `sort` metadata field is set to `t1.sort`.

**Go** (formula.go:177): `WSort: t1.NodeSort()`. Matches.

**Status**: CONFORMANT.

---

### 1.12 `Cond` sort validation: Python has dead-code validation

**Python** (logic.py:287-288): `bad_sorts = [i for i, t in enumerate([t1]) if i == 1 and t.sort not in (Boolean, TopS)]` — This iterates over `[t1]` (single element) with `i == 1` which is never true for a single-element list. So the validation is effectively dead code.

**Go** (formula.go:200-201): `NewCond` has no sort validation at all.

**Status**: Both effectively skip validation. CONFORMANT (both have the same bug/non-behavior).

---

### 1.13 `String()` methods: Go diverges from Python `pretty_fmla` / `ugly`

Python's `ivy_logic.py:1428-1434` monkey-patches `__str__` on ALL formula types to use `pretty_fmla → ugly` which produces **infix** notation with operator precedence. Go uses **prefix** notation (except for `Implies` and `Not` which we recently changed).

| Type | Python `ugly` output | Go `String()` output | Match? |
|------|---------------------|---------------------|--------|
| `Eq` | `(a = b)` | `(a == b)` | NO — `=` vs `==` |
| `Not(Eq)` | `(a ~= b)` | `(a != b)` | NO — `~=` vs `!=` |
| `Not(x)` | `~x` | `~x` | YES |
| `And(a,b)` | `(a & b)` | `And(a, b)` | NO |
| `Or(a,b)` | `(a \| b)` | `Or(a, b)` | NO |
| `Implies(a,b)` | `(a -> b)` | `(a -> b)` | YES |
| `Iff(a,b)` | `(a <-> b)` | `Iff(a, b)` | NO |
| `ForAll` | `(forall X:S. body)` | `(ForAll X:S. body)` | NO — case |
| `Exists` | `(exists X:S. body)` | `(Exists X:S. body)` | NO — case |
| `Lambda` | `(lambda X:S. body)` | `(Lambda X:S. body)` | NO — case |
| `Apply(f,x,y)` | `f(x,y)` | `f(x,y)` | YES |
| `Var` | `X` (or `X:S` with annotation) | `X` | YES |
| `Const` | `c` (or `c:S` with annotation) | `c` | YES |
| `Globally` | `□ body` (unicode) | `globally(body)` | NO |
| `Eventually` | `◆ body` (unicode) | `eventually(body)` | NO |

**Impact**: String conformance is critical for randomized conformance testing. Every formula printed for comparison will differ.

**How to conform**: Implement a `PrettyFmla(n Node) string` function in Go that replicates Python's `ugly` formatting with operator precedence. This should be used for all user-facing output. The existing `String()` methods can remain as debug/internal format, or be replaced entirely. Key rules:
- `Eq`: `(a = b)` not `(a == b)`
- `Not(Eq(a,b))`: `(a ~= b)` not `(a != b)`
- `And`: infix `&` with parens when precedence requires
- `Or`: infix `|`
- `Iff`: infix `<->`
- `ForAll`/`Exists`/`Lambda`: lowercase keywords
- `Globally`/`Eventually`: Unicode symbols or lowercase keywords

---

### 1.14 `EnumeratedSort.constructors` property missing from Go

**Python** (logic.py:60-61): `EnumeratedSort.constructors` returns `[Const(n, self) for n in self.extension]` — creates Const nodes for each extension element with the sort set to the EnumeratedSort itself.

**Go**: No equivalent method on `EnumeratedSort`.

**Impact**: Any code that calls `sort.constructors` to get the constructor constants will need a different approach in Go.

**How to conform**: Add a `Constructors() []*Const` method to `EnumeratedSort` that returns `[NewConst(name, self) for each extension element]`.

---

### 1.15 `Var.__call__` / `Const.__call__` / `NamedBinder.__call__`: Python callable objects

**Python**: `Var.__call__`, `Const.__call__`, and `NamedBinder.__call__` allow `f(x, y)` syntax to create `Apply(f, x, y)`. This is used extensively in Python code like `leq(X, Y)`.

**Go**: Has `Call(terms ...Node) (Node, error)` methods. Callers use `v.Call(x, y)` instead of `v(x, y)`.

**Status**: Structural difference, conformant in behavior. The Go API is idiomatic. No action needed.

---

## 2. `type_inference.py` vs `typeinfer/`

### 2.1 `ConvertToSortVars`: SortVar cannot be stored in FunctionSort

**Python** (type_inference.py:123-126): `convert_to_sortvars` for FunctionSort returns `FunctionSort(*(convert_to_sortvars(x) for x in s))`. Python's `FunctionSort` accepts `SortVar` objects in its sorts list because `SortVar` is duck-typed — it acts like a Sort.

**Go** (typeinfer/unify.go:114-127): `ConvertToSortVars` for FunctionSort converts each sub-sort, but when the result is a `SortVar` (not a concrete sort), it falls back to `logic.NewTopSort()` as a placeholder (line 123). This **loses the sort variable linkage**.

**Impact**: This is the root cause of the TopSort leak we fixed in the parser. While the parser fix addresses the immediate symptom, this `ConvertToSortVars` limitation means that any FunctionSort with TopSort elements will lose sort variable information during type inference. The inference can still work via the Apply case's direct unification, but it's fragile.

**How to conform**: The proper fix requires either: (a) making `logic.FunctionSort.Sorts` accept a `SortOrVar` interface instead of `logic.Sort`, or (b) tracking a parallel `[]SortOrVar` alongside the `[]Sort` in the type inference context. Option (b) is less invasive — maintain a mapping from FunctionSort identity to `[]SortOrVar` in the inference environment.

---

### 2.2 `infer_sorts` Apply case: Python unifies func sort with constructed FunctionSort

**Python** (type_inference.py:175-183):
```python
func_s, func_t = infer_sorts(t.func, env)
xys = [infer_sorts(tt, env) for tt in t.terms]
terms_s = [x for x, y in xys]
sorts = terms_s + [SortVar()]
unify(func_s, FunctionSort(*sorts))
return sorts[-1], ...
```
Always builds a `FunctionSort` from term sort vars + a fresh result sort var, and unifies with the function's sort. This works because Python's `FunctionSort` accepts `SortVar` objects.

**Go** (typeinfer/infer.go:78-147): Has two branches:
- If func's sort resolves to a `SortWrapper` wrapping a `FunctionSort`: unifies element-by-element (correct).
- If func's sort is still a `SortVar`: builds a FunctionSort with TopSort placeholders and unifies (lossy, see §2.1).

**Impact**: The `else` branch (lines 114-129) is structurally different from Python. When a function's sort hasn't been resolved yet (still a SortVar), Go creates a FunctionSort with TopSort elements, losing the sort variable linkage. Python creates a FunctionSort with actual SortVars.

**How to conform**: Restructure the Apply case to always unify element-by-element, regardless of whether the func sort is already resolved. Create fresh SortVars for term sorts, unify each term sort with the corresponding function domain sort, and unify the result sort with the function range sort. This avoids the need to create a FunctionSort with SortVars.

---

## 3. `ivy_logic.py` vs `ivylogic/`

### 3.1 `clone()` method: Python monkey-patches all logic types

**Python** (ivy_logic.py:260-284): Adds `.clone(args)` to all logic types via monkey-patching. The semantics differ by type:
- `lg.Apply.clone = lambda self,args: type(self)(self.func, *args)` — clone replaces **terms only**, preserving func.
- `lg.ForAll/Exists/Lambda.clone = lambda self,args: type(self)(self.variables, *args)` — preserves variables.
- `lg.Globally/Eventually.clone = lambda self,args: type(self)(self.environ, *args)` — preserves environ.
- `lg.NamedBinder.clone = lambda self,args: NamedBinder(self.name, self.variables, self.environ, *args)` — preserves name, variables, environ.
- `Symbol.clone = lambda self,args: self` — constants are immutable.
- `Variable.clone = lambda self,args: self` — variables are immutable.

**Go**: The `Clone()` method signature and behavior should match exactly. Check: does Go's `Apply.Clone(args)` preserve `Func` while replacing `Terms`? Does `ForAll.Clone(args)` preserve `Variables`?

**Impact**: Any code that uses `clone()` for tree rewriting (substitution, renaming) depends on these exact semantics. If Go's clone replaces the wrong fields, substitution will produce incorrect results.

**How to conform**: Audit every `Clone()` implementation in Go against the Python monkey-patched version. The most common error is Go's `Clone` replacing ALL children (including metadata fields) instead of just the `args`.

---

### 3.2 `.args` property: Python exposes children-only view

**Python** (ivy_logic.py:260-284):
- `lg.Apply.args = property(lambda self: self.terms)` — returns ONLY terms, NOT func.
- For all other formula types: `cls.args = property(lambda self: [a for a in self])` — iterates the `recstruct` which yields only children (not metadata fields).
- `Symbol.args = property(lambda self: [])` — constants have no args.
- `Variable.args = property(lambda self: [])` — variables have no args.

**Go** `Children()` method:
- `Apply.Children()` returns `[Func] + Terms` — includes Func!
- Other types return their children.

**Impact**: CRITICAL. Python's `args` for `Apply` does NOT include `func`, but Go's `Children()` DOES include `Func`. Any code that iterates children to walk the formula tree will process `func` twice in Go (once as a child, once explicitly) or will apply transformations to `func` when it shouldn't.

This affects substitution (`substitute`), variable collection (`free_variables`, `used_variables`, `used_constants`), printing, and tree comparison. Functions that walk `args` in Python skip the function head of Apply nodes, while Go's `Children()` includes it.

**How to conform**: Either:
(a) Change `Apply.Children()` to return only `Terms` (not `Func`), matching Python's `args`. This is the simplest fix but requires auditing all Go code that explicitly accesses `Func` after calling `Children()`.
(b) Add a separate `Args() []Node` method matching Python's semantics and use it wherever Python uses `.args`.

This is the most architecturally significant divergence found so far.

---

### 3.3 `.rep` property: Python exposes the "representative" of a term

**Python** (ivy_logic.py:128-129,282-283,295):
- `Symbol.rep = property(lambda self: self)` — a constant's rep is itself.
- `lg.Apply.rep = property(lambda self: self.func)` — an Apply's rep is its func.
- `lg.Eq.rep = property(lambda self: Symbol('=', RelationSort(...)))` — Eq's rep is the = symbol.
- `Variable.rep = property(lambda self: self.name)` — a variable's rep is its name (string!).

**Go**: No `.Rep()` method on logic types. The Go code accesses `Func` directly on Apply, `Name` on Const/Var.

**Impact**: Any Python code that uses `x.rep` polymorphically (e.g., to get the "head symbol" of a term regardless of whether it's Apply, Const, or Eq) will need case-by-case handling in Go.

**How to conform**: Add a `Rep() Node` method to the logic types, or ensure all callsites use the appropriate field directly.

---

### 3.4 `Symbol.__call__`: conditional Apply creation

**Python** (ivy_logic.py:153): `Symbol.__call__ = lambda self,*args: App(self,*args) if len(args) > 0 or isinstance(self.sort, FunctionSort) else self`

Note the `or isinstance(self.sort, FunctionSort)` clause. If a FunctionSort constant is called with zero args, it STILL creates `Apply(self)` (a nullary application). This is different from Go's `Call()` which returns self for zero args.

**Go** (term.go:62-67): `func (c *Const) Call(terms ...Node) (Node, error) { if len(terms) == 0 { return c, nil } ... }`

**Impact**: In Python, calling a FunctionSort constant with no args creates `Apply(const)`, while in Go it returns `const`. This matters for 0-arity functions where the distinction between a function symbol and its application is semantically important.

**How to conform**: Change Go's `Const.Call()` to check if `c.CSort` is a `FunctionSort` and if so, create `Apply(c)` even with zero terms.

---

### 3.5 `Variable.__call__`: Python only applies if sort is FunctionSort

**Python** (ivy_logic.py:712): `Variable.__call__ = lambda self,*args: App(self,*args) if isinstance(self.sort, FunctionSort) else self`

Note: this ignores the number of args! If `self.sort` is not FunctionSort, it returns `self` regardless of args. This differs from Go's `Var.Call()` which tries `NewApply` for any non-zero args.

**Impact**: If a variable with a non-function sort is accidentally called with args, Python silently returns the variable, while Go creates an Apply (which may fail at construction due to sort mismatch).

**How to conform**: Minor — Python's behavior is arguably buggy (silently drops args). The Go behavior of failing with a sort error is more correct. Document but don't change.

---

### 3.6 `Sig.__init__`: Python initializes `sorts["bool"]` to `RelationSort([])`

**Python** (ivy_logic.py:882): `self.sorts["bool"] = RelationSort([])` — which is just `Boolean` (since `RelationSort([])` returns `Boolean` when domain is empty).

**Go**: `ivylogic.NewSig()` — need to check if it initializes `Sorts["bool"]`.

**Impact**: If Go doesn't add "bool" to the sort map, any code that looks up `sig.sorts["bool"]` will fail or return nil.

**How to conform**: Verify Go's `NewSig()` adds `Sorts["bool"] = Boolean`. If not, add it.

---

### 3.7 `Sig.add_symbol`: polymorphic handling via `UnionSort`

**Python** (ivy_logic.py:910-924): When `ivy_have_polymorphism` is true and the symbol name is in `polymorphic_symbols`, it stores a `Symbol(name, UnionSort())` and appends sorts to the union. For non-polymorphic symbols, it checks for redefinition.

**Go** (`ivylogic/sig.go:91-128`): Has `IsPolymorphicName()` check and `UnionSort` handling. Need to verify the polymorphic symbol list matches Python's `polymorphic_symbols_list`.

**Impact**: If the polymorphic symbol lists differ, arithmetic and comparison operators will be handled differently.

**How to conform**: Compare Go's `polymorphicSymbolsList` (or equivalent) with Python's `polymorphic_symbols_list` at ivy_logic.py:1043-1066. Ensure all entries match.

---

### 3.8 `PolySymsDict`: dynamic `bfe[lo:hi]` pattern matching

**Python** (ivy_logic.py:1079-1085): `PolySymsDict` overrides `__contains__` and `__getitem__` to dynamically create entries for `bfe[...]` patterns.

**Go**: Need to check if `FindPolymorphicSymbol()` handles the `bfe[` prefix dynamically.

**Impact**: Without this, bit-field extract operations won't be found as polymorphic symbols.

**How to conform**: Verify Go's polymorphic symbol lookup handles `bfe[` patterns.

---

### 3.9 `polymorphic_macros_map`: `<=`, `>`, `>=` expand to `<`

**Python** (ivy_logic.py:1090-1094): `<=` maps to `<`, `>` maps to `<`, `>=` maps to `<`. These are expanded via `macros_expansions` at line 1096-1100.

**Go**: Need to check if Go has equivalent macro expansion for comparison operators.

**Impact**: If Go doesn't expand `<=` to `<` + `=`, the Z3 encoding will differ.

**How to conform**: Verify Go's macro expansion matches Python's. Look for `polymorphic_macros_map` equivalent in Go.

---

### 3.10 `EnumeratedSort.__str__` monkey-patched to `self.name`

**Python** (ivy_logic.py:777-783): After monkey-patching, `EnumeratedSort.defines()`, `.is_relational()`, `.dom`, `.rng`, `.is_finite`, `.rep` are added. The `__str__` from `logic.py` (which returns extensions) is NOT explicitly overridden here, but the `pretty_fmla` system replaces `__str__` for display (§1.13). For sorts specifically, `str(sort)` uses `logic.py`'s original `__str__`, not `ugly`.

Actually wait — line 1434 only patches `[Eq, Not, And, Or, Implies, Iff, Ite, ForAll, Exists, Apply, Var, Const, Lambda, NamedBinder]`. It does NOT patch `EnumeratedSort`, `UninterpretedSort`, etc. So `str(EnumeratedSort)` still uses logic.py's `'{' + ','.join(...)}`... unless it was further patched elsewhere.

Checking logic.py:57-59: `def __str__(self): return '{' + ','.join(self.extension) + '}'`. But this is commented out at line 58 with `# return self.name` BELOW it... Actually no, the live line IS the extension format, and the name return is in a comment.

**Status**: Need to actually run Python and check what `str(EnumeratedSort("color", ["red","green","blue"]))` returns. Based on code reading, it returns `{red,green,blue}` — matching Go.

---

### 3.11 `pretty_fmla` / `ugly`: Complete specification

The `ugly` system uses precedence-based formatting. Here is the complete spec:

| Prec | Operator |
|------|----------|
| 1 | default (function application) |
| 2 | temporal (globally, eventually, when) |
| 3 | `->`, `<->` |
| 4 | `\|` |
| 5 | `&` |
| 6 | `~` (negation) |
| 7 | `=`, `~=` |
| 8 | (used in Not(Eq) for ~=) |
| 9 | Ite/Cond interior |
| 12-15 | arithmetic (`+`, `-`, `*`, `/`) |

The `nary_ugly(op, args, myprec, prec)` function:
- Joins args with ` op `
- Wraps in parens if `len(args) > 1 AND myprec <= prec`

The `nary_paren(op, args, myprec, prec)` function (used only for `And`):
- Joins args with ` op `
- ALWAYS wraps in parens (regardless of precedence)

`Apply.ugly` (`app_ugly`):
- Infix symbols (`<`,`<=`,`>`,`>=`,`+`,`-`,`*`,`/`): uses ` op ` join with precedence
- Non-infix: `name(arg1,arg2,...)` — NOTE: comma without space between args

Quantifier `ugly` (`quant_ugly`):
- `forall`/`exists`/`lambda`/`$name` (lowercase)
- Variables formatted with `v.ugly(1)` (which may include `:sort` annotation)
- Body formatted with `body.ugly(1)`
- Wrapped in parens if `prec >= 1`

`Var.ugly`:
- If `show_variable_sorts` and sort is NOT TopSort or SortVar: `name:sort_name`
- Otherwise: `name`

`Const.ugly`:
- If `show_numeral_sorts` and `is_numeral()` and sort is NOT TopSort: `name:sort_name`
- Otherwise: `name`

**How to conform**: Implement a `PrettyFmla(n Node) string` function in Go that replicates this exact precedence/formatting system. Use it for all user-facing formula display.

---

## 4. `ivy_logic_utils.py` vs `clauseops/`, `logicutil/`

(To be continued)

---

## 5. `ivy_solver.py` vs `solver/`

(To be continued)

---

## 6. `ivy_transrel.py` vs `transrel/`

(To be continued)

---

## 7. `ivy_actions.py` vs `actions/`

(To be continued)

---

## 8. `ivy_compiler.py` vs `compiler/`

(To be continued)

---

## 9. `ivy_isolate.py` vs `isolate/`

(To be continued)

---

## 10. `ivy_check.py` vs `check/`

(To be continued)

---

## 11. `ivy_art.py` / `ivy_interp.py` vs `art/`, `interp/`

### 11.1 `concrete_post` stores `update` on state — FIXED

**Python** (ivy_interp.py:206): `res.update = update` — stores the transition relation Update on the post-state for later use by `get_history`.

**Go** (art/art.go): `PostState()` was not storing `s.Update = update`. **Now fixed** in this session.

**Status**: FIXED.

---

### 11.2 `get_history` / `history_forward_step` axioms parameter

**Python** (ivy_interp.py:591): `history.forward_step(state.pred.domain.background_theory(state.pred.in_scope), state.update, action)`

**Go** (art/art.go): `GetHistory` was passing `lg.True` for axioms. **Now fixed** to pass `state.Pred.Domain.BackgroundTheory(...)`.

**Status**: FIXED.

---

### 11.3 `check_final_cond` uses `get_history` — FIXED

**Python** (ivy_trace.py:326-328): `history = ag.get_history(post)` then `clauses = history.post`.

**Go** (trace/trace.go): `CheckFinalCond` was using `post.Clauses` directly. **Now fixed** to use `ag.GetHistory(post, nil)`.

**Status**: FIXED.

---

## Summary of Required Fixes (by priority)

### Critical (affects verification correctness)
1. §2.1 / §2.2 — `ConvertToSortVars` TopSort placeholder in FunctionSort. The parser fix addresses the immediate symptom, but the underlying type inference limitation remains.

### High (affects conformance testing)
2. §1.13 — String formatting divergence (`ugly`/`pretty_fmla`). Need to implement Go `PrettyFmla` matching Python's infix notation with operator precedence.
3. §1.7 — `EnumeratedSort.String()` returns extension format instead of name.
4. §1.2 — ForAll/Exists variable ordering (frozenset vs slice).

### Medium (could cause subtle bugs)
5. §1.14 — Missing `EnumeratedSort.Constructors()` method.
6. §1.1 — Immutability / hashability differences (audit all set/dict usage with logic nodes).

### Low (unlikely to cause issues)
7. §1.4 — Apply comma separator (verify actual Python output).
8. §1.6 — Unicode in symbol names (unlikely in practice).
