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

### 1.7 `EnumeratedSort.String()` name vs extension — VERIFIED CONFORMANT

**Python** (logic.py:57-58): Returns `'{' + ','.join(self.extension) + '}'`. No monkey-patching override in ivy_logic.py (verified by running Python and by grep). `pretty_fmla` system does not patch EnumeratedSort.

**Go** (sort.go:81-82): Returns `"{" + strings.Join(s.Extension, ",") + "}"`. Matches Python.

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

### 1.13 `String()` methods: Go diverges from Python `pretty_fmla` / `ugly` — PARTIALLY FIXED

Python's `ivy_logic.py:1428-1434` monkey-patches `__str__` on ALL formula types to use `pretty_fmla → ugly` which produces **infix** notation with operator precedence. Go's `String()` methods still use prefix notation for And, Or, Iff, etc.

**FIXED**: Implemented `PrettyFmla(n Node) string` in `logic/pretty.go` that replicates Python's full `pretty_fmla → drop_annotations → ugly` system with correct operator precedence, infix notation, and sort annotation handling. This is now used in `webui/session.go` for conformance-critical output. The existing `String()` methods are retained for debug/internal use; `PrettyFmla` should be used for all user-facing or conformance-critical formula display.

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

### 2.1 `ConvertToSortVars`: SortVar cannot be stored in FunctionSort — FIXED

**Python** (type_inference.py:123-126): `convert_to_sortvars` for FunctionSort returns `FunctionSort(*(convert_to_sortvars(x) for x in s))`. Python's `FunctionSort` accepts `SortVar` objects in its sorts list because `SortVar` is duck-typed — it acts like a Sort.

**Go** (typeinfer/unify.go): Previously, `ConvertToSortVars` for FunctionSort fell back to `logic.NewTopSort()` as a placeholder when a sub-sort converted to a `SortVar`, losing the sort variable linkage.

**FIXED**: Introduced `FunctionSortVar` type in `typeinfer/sortvar.go` — a parallel to `FunctionSort` that holds `[]SortOrVar` instead of `[]logic.Sort`, mirroring Python's ability to store `SortVar` objects inside `FunctionSort`. Updated `ConvertToSortVars`, `InsertSortVars`, `ConvertFromSortVars`, `Unify`, `OccursIn`, and the Apply case in `InferSorts` to use `FunctionSortVar`.

---

### 2.2 `infer_sorts` Apply case: Python unifies func sort with constructed FunctionSort — FIXED

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

**FIXED**: The Go Apply case now builds `FunctionSortVar(termSorts..., resultSortVar)` and unifies with the function's sort — directly mirroring Python's `unify(func_s, FunctionSort(*sorts))`. The old two-branch approach (element-by-element for resolved sorts, TopSort-placeholder for unresolved) has been replaced with a single unified path using `FunctionSortVar`.

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

### 4.1 `substitute_ast` walks `.args` (not `.Children()`)

**Python** (ivy_logic_utils.py:160-170): `substitute_ast` iterates `ast.args` for recursive substitution. For Apply nodes, `.args` returns ONLY terms (not func), so the function head is never substituted — only its arguments are.

**Go**: Go's substitution functions must use the equivalent of `.args` (Terms only), NOT `Children()` (which includes Func). If Go substitution walks `Children()`, it will incorrectly attempt to substitute inside the function symbol of Apply nodes.

**Impact**: CRITICAL. Substitution is one of the most frequently used operations. If Go substitutes inside `Apply.Func`, it could change function symbols in ways Python never does.

**How to conform**: Audit all Go substitution functions to ensure they skip `Apply.Func` and only process `Apply.Terms`. The correct Go pattern is:
```go
case *logic.Apply:
    // substitute in terms only, not func
    newTerms := make([]logic.Node, len(a.Terms))
    for i, t := range a.Terms {
        newTerms[i] = substitute(t, subs)
    }
    return logic.NewApply(a.Func, newTerms...)
```

---

### 4.2 `constants_ast` walks `.args` → excludes Apply.Func from constant collection

**Python** (ivy_logic_utils.py:501-507): `constants_ast` yields `ast.rep` if `is_constant(ast)`, then recurses into `ast.args`. For Apply, `.args` excludes func, so function symbols are NOT collected as constants. The function symbol is accessed separately via `ast.rep` in `symbols_ast` (line 534-545).

**Go**: If Go's constant collection uses `Children()`, it will include `Apply.Func` as a constant, producing a superset of the Python result.

**Impact**: Affects `used_constants`, `used_symbols`, and any function that collects symbols from formulas. Overcounting could affect solver interaction, cone-of-influence filtering, and isolate extraction.

**How to conform**: Same fix as §4.1 — audit all tree-walking functions in Go.

---

### 4.3 `symbols_ast` explicitly accesses `ast.rep` for Apply head

**Python** (ivy_logic_utils.py:534-545):
```python
def symbols_ast(ast):
    if is_app(ast):
        if is_binder(ast.rep):
            for x in symbols_ast(ast.rep.body): yield x
        else:
            yield ast.rep   # <-- explicit access to func head
    for arg in ast.args:    # <-- iterates only terms (not func)
        for x in symbols_ast(arg): yield x
```

This explicitly yields `ast.rep` (the function symbol) for non-binder Apply nodes, then recurses into `ast.args` (terms only). The function head is handled once, explicitly.

**Go**: If Go uses `Children()` which includes Func, and also has explicit Func handling, the function symbol will be processed twice.

**How to conform**: Go's symbol collection must follow the same pattern: explicitly handle `Apply.Func`, then recurse into only `Apply.Terms`.

---

### 4.4 `Clauses.__init__` flattens And via `collect_and_list`

**Python** (ivy_logic_utils.py:45): `self.fmlas = list(collect_and_list([coerce_clause_to_formula(c) for c in fmlas]))` — this flattens nested And nodes. If a formula is `And(a, And(b, c))`, `collect_and_list` produces `[a, b, c]`.

**Go** (`clauseops/clauses.go`): Check whether `NewClauses` or `FormulaToClauses` flattens And nodes similarly.

**Impact**: If Go doesn't flatten, a clause set with `And(a, And(b, c))` will have 1 formula instead of 3, potentially affecting solver interaction.

**How to conform**: Verify Go flattens And in Clauses construction. If not, add `collectAndList` helper.

---

### 4.5 `Clauses.copy()` loses annotation

**Python** (ivy_logic_utils.py:68): `def copy(self): return Clauses(list(self.fmlas), list(self.defs))` — does NOT copy `annot`. The annotation is lost.

**Go**: Check if `Clauses.Copy()` copies the annotation.

**Impact**: If Go copies the annotation but Python doesn't, composed clause sets will have different annotation state.

**How to conform**: Verify Go's `Copy()` matches Python (either both copy annot, or both drop it).

---

### 4.6 `close_epr` wraps in `ForAll` for free variables

**Python** (ivy_logic_utils.py:107-120): `close_epr` wraps formula in `ForAll(variables, fmla)` where variables are the free variables. If no free variables, returns as-is. Note: uses `used_variables_ast` (not `free_variables`) so bound variables are excluded.

**Go** (`clauseops.CloseEPR` or equivalent): Check implementation matches.

**Impact**: Affects `Clauses.to_formula()` which calls `close_epr`. If Go wraps differently, Z3 receives different quantifier structure.

---

## 5. `ivy_solver.py` vs `solver/`

### 5.1 `solver_name`: polymorphic symbol naming includes domain sorts

**Python** (ivy_solver.py:60-78): For polymorphic symbols, the solver name is composed as `name + ':' + domain_sort_names`. E.g., `+:int:int` for integer addition. For symbols in `sig.interp` (interpreted sorts), returns `None` (handled natively by Z3).

**Go**: Check if Go's Z3 bridge uses the same naming convention for polymorphic symbols. The `makeFuncDecl` in `z3bridge/translate.go` uses `name + ":" + fs.String()` as the cache key.

**Impact**: If naming differs, the same polymorphic symbol at different sorts could collide or create spurious duplicates in Z3.

**How to conform**: Verify Go's polymorphic symbol naming matches Python's `solver_name` convention exactly.

---

### 5.2 Z3 sort translation: `int`, `nat`, `bv[N]`, `strbv[N]`, `intbv[N]`

**Python** (ivy_solver.py:111-135): Maps sort names to Z3 sorts:
- `int` → `z3.IntSort()`
- `nat` → `z3.IntSort()` (same as int, with non-negativity constraints)
- `bv[N]` → `z3.BitVecSort(N)`
- `strbv[N]` → `z3.BitVecSort(N)`
- `intbv[N]` → `z3.BitVecSort(N)`
- `real` → `z3.RealSort()`
- `strlit` → `z3.StringSort()`
- `arr[dom][rng]` → `z3.ArraySort(dom_z3, rng_z3)`

**Go**: Check `TranslateSort` in `z3bridge/translate.go` handles all these cases.

**Impact**: Missing sort translations will cause Z3 errors for programs using these types.

---

### 5.3 `relations_dict` and `functions_dict`: BV-aware comparison

**Python** (ivy_solver.py:152-157): Comparison operators check `z3.is_bv(x)` and dispatch to unsigned BV comparisons (`z3.ULT`, `z3.ULE`, `z3.UGT`, `z3.UGE`) for bitvector sorts, falling back to integer comparison otherwise.

**Go**: Check if Go's `translateBuiltinOp` in z3bridge handles the BV vs integer dispatch.

**Impact**: Without BV-aware comparisons, bitvector programs will get integer semantics for `<`, `<=`, etc.

---

### 5.4 Z3 enum encoding: `use_z3_enums` flag

**Python** (ivy_solver.py:33): `use_z3_enums = True` — uses Z3 native enumeration sorts. This affects how EnumeratedSort is translated to Z3.

**Go**: Check if Go uses Z3 native enums or a manual encoding.

**Impact**: Different encodings may produce different model structure and potentially different SAT/UNSAT results.

---

## 6. `ivy_transrel.py` vs `transrel/`

### 6.1 Update representation: tuple of Clauses vs struct of lg.Node — VERIFIED, MINOR FIX

**Python**: An update is a triple `(modified, clauses, pre)` where `clauses` and `pre` are `Clauses` objects (with `fmlas`, `defs`, and `annot`).

**Go**: `transrel.Update` has `Modified []string`, `TR lg.Node`, `Pre lg.Node`, `Annot interface{}`. TR and Pre are plain Node, not Clauses.

**Verified**: Go's approach of inlining definitions as formulas is semantically correct for the current verification pipeline. Python's `Definition.to_constraint()` produces `Eq(lhs, rhs)` for individuals and `Iff(lhs, rhs)` for boolean relations — Go now uses the same. When fed to Z3, both produce equivalent constraints. The structural difference (separate `defs` list vs inlined formulas) only matters for Python's `not_clauses_to_z3` which separates Skolem definitions during negation — a path Go doesn't use.

**Bug fixed**: Go's `mkAssignClauses` was using a manual CNF encoding `And(Or(a, ~b), Or(~a, b))` for boolean definitions instead of `Iff(a, b)`. Changed to use `Iff` to match Python's `Definition.to_constraint()` exactly.

---

### 6.2 `forward_image_map`: existential quantification of modified symbols

**Python** (ivy_transrel.py:417-426): `forward_image_map` conjoins pre-state with transition relation, then existentially quantifies out the modified symbols (via `exist_quant_map`), then renames `new_x → x`.

**Go** (`transrel.ForwardImageMap`): Check implementation matches.

**Impact**: Incorrect forward image computation will produce wrong post-states, leading to false verification results.

---

### 6.3 `compose_state_action` checks precondition

**Python** (ivy_transrel.py:464-488): `compose_state_action` with `check=True` checks the action's precondition against the state. If satisfied (model found), raises `ActionFailed` with counterexample. This is how `require` violations are detected.

**Go**: Check if Go's compose function performs this precondition check.

**Impact**: Without precondition checking, `require` statement violations will not be detected during verification.

---

## 7. `ivy_actions.py` vs `actions/`

### 7.1 `Action.int_update` applies update axioms from `domain.updates` — VERIFIED CORRECT

**Python** (ivy_actions.py:201-217): After computing `action_update`, iterates `domain.updates` and calls `u.get_update_axioms(updated, self)` for each.

**Go** (`actions/update.go:853`): `intUpdateFromActionUpdate` calls `applyUpdateAxioms` which iterates `ctx.Domain.Updates` and calls `GetUpdateAxioms`. Matches Python.

---

### 7.2 `Action.update` applies `bind_olds` and `hide_formals` — FIXED

**Python** (ivy_actions.py:218-219): `def update(self, domain, in_scope): return self.hide_formals(bind_olds_action(self.int_update(domain, in_scope)))`.

**Go** (`actions/update.go:1500`): `GetUpdate` correctly calls `IntUpdate` → `BindOldsAction` → `hideFormals`. This chain matches Python exactly.

**FIXED**: `art.PostState` was using a `Updater` interface that no action implemented (dead code). Changed to call `actions.GetUpdateForArt(op, domain, inScope)` directly, which invokes the full `GetUpdate` chain (IntUpdate + BindOlds + HideFormals). Previously, `PostState` fell through to the "carry pre-state forward" fallback, never computing the actual transition relation through the action semantics pipeline.

---

### 7.3 `AssignAction.action_update`: destructor and variant assignment

**Python** (ivy_actions.py:493-573): `AssignAction.action_update` handles three cases:
1. Hierarchical assignment (field assignment via `domain.hierarchy`)
2. Destructor assignment (`n.name in module.destructor_sorts`)
3. Variant assignment (`domain.is_variant(lhs.sort, rhs.sort)`)
4. Normal assignment (`mk_assign_clauses`)

Each produces different clause structure. The normal case creates a `Definition(new_n(placeholders), Ite(eqs, rhs, n(placeholders)))` for partial assignments.

**Go** (`actions/update.go:488`): `AssignAction.ActionUpdate` — check all four cases.

**Impact**: Incorrect assignment update computation will produce wrong transition relations.

---

### 7.4 `mk_assign_clauses`: partial assignment with ITE — VERIFIED CORRECT

**Python** (ivy_actions.py:578-590): For a partial assignment like `a(x) := v`, creates `new_a(V0) = Ite(V0 = x, v, a(V0))`.

**Go** (`actions/update.go:354`): Faithfully replicates: placeholder variables, equality conditions, variable substitution, ITE for partial assignment, Iff/Eq for definition encoding. Verified step-by-step against Python.

---

## 8. `ivy_compiler.py` vs `compiler/`

### 8.1 Three-pass compilation: IvyDomainSetup, IvyConjectureSetup, IvyARGSetup

**Python** (ivy_compiler.py:2190-2254): Runs three separate declaration interpreter passes:
1. `IvyDomainSetup`: types, relations, functions, axioms, definitions
2. `IvyConjectureSetup`: conjectures
3. `IvyARGSetup`: exports, delegates, actions, initializers

Each pass uses `TopContext(collect_actions(decls))` which pre-collects action signatures for forward reference resolution.

**Go**: Check if Go uses three passes or a single pass.

**Impact**: Single-pass compilation can't handle forward references (action A calling action B before B is declared).

---

### 8.2 Post-processing passes

**Python** (ivy_compiler.py:2210-2253): After the three compilation passes:
- `create_sort_order` (Tarjan SCC)
- `create_constructor_schemata`
- `fix_constructors`
- `check_definitions` (cycle detection)
- `attach_proofs`
- `check_properties`
- `apply_assert_proofs`
- `create_conj_actions`
- `handle_temporals`

**Go**: MISSING.md marks these as [x] done, but verify each is wired into the compilation pipeline.

---

## 9. `ivy_isolate.py` vs `isolate/`

### 9.1 `create_isolate` modifies `im.module` in-place

**Python**: `create_isolate` operates on `im.module` (the global module singleton), modifying it in-place. The `with im.module.copy()` context manager saves/restores the original.

**Go**: Check if Go's `CreateIsolate` modifies the module in-place or creates a copy.

**Impact**: If Go creates a copy but callers expect in-place modification (or vice versa), the module state will be wrong for subsequent checking.

---

## 10. `ivy_check.py` vs `check/`

### 10.1 `check_isolate` initialization check: `initializer=lambda x:None`

**Python** (ivy_check.py:603): `ag = ivy_art.AnalysisGraph(initializer=lambda x:None)` — creates an AG with a no-op initializer. This produces an initial state with `True` clauses (no initialization constraints).

**Go**: Check if Go's init check uses the same approach.

**Impact**: If Go's init state includes initialization constraints, the invariant check will be checking a stronger condition than Python.

---

### 10.2 `check_conjs_in_state` uses `ConjSubgoals` if available

**Python** (ivy_check.py): Checks `im.module.conj_subgoals` first, falling back to `im.module.labeled_conjs`. Subgoals come from proof tactic application.

**Go**: Check if Go uses `mod.ConjSubgoals` before `mod.LabeledConjs`.

---

### 10.3 Action preservation: `get_conjs` as pre-state, execute, check post

**Python** (ivy_check.py:636-642):
```python
ag = ivy_art.AnalysisGraph()
pre = itp.State()
pre.clauses = get_conjs(mod)
with itp.EvalContext(check=False):
    post = ag.execute(action, pre)
check_conjs_in_state(mod, ag, post, indent=12, pcs=...)
```

**Go** (check/isolate_check.go:239-247): Similar flow. But verify:
1. `get_conjs` excludes explicit and unprovable conjectures (matching Python)
2. The `EvalContext(check=False)` disables precondition checking during execution
3. `post` state is created with correct clauses

---

## 11. `ivy_art.py` / `ivy_interp.py` vs `art/`, `interp/`

(Findings from debugging session documented above — §11.1-11.3 are FIXED)

### 11.4 `AnalysisGraph.__init__`: initializer parameter

**Python** (ivy_art.py): `AnalysisGraph.__init__` accepts an `initializer` parameter. If `None`, runs the default initialization (computing init_cond from module). If provided as `lambda x: None`, skips initialization.

**Go**: Check if Go's `art.NewAnalysisGraph` has equivalent initializer handling.

---

### 11.5 `State.value` vs `State.Clauses`

**Python**: States use `.value` to access their clauses (a tuple `(updated, clauses, pre)`). The `value` is a full transrel-style state, not just clauses.

**Go**: States use `.Clauses` which is `*clauseops.Clauses`. This is a different representation — Go stores only the clauses, not the full update triple.

**Impact**: This affects how states are composed, how histories are built, and how forward images are computed. If Go stores only clauses but Python stores the full update triple, state composition will differ.

**How to conform**: Verify that Go's state composition in `art.PostState` properly computes the full transition relation (not just carrying clauses forward).

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

1. §7.1/7.2 — Update axioms, bind_olds, hide_formals in action updates. **FIXED**: `art.PostState` was using a dead-code `Updater` interface. Now calls `actions.GetUpdateForArt` directly, which invokes the full pipeline: `IntUpdate` (with update axioms) → `BindOldsAction` → `hideFormals`.
2. §6.1 — Update representation (Clauses with defs vs bare Node). **FIXED**: Refactored `transrel.Update` to use `*co.Clauses` for TR and Pre (matching Python's Clauses tuple), and `[]*lg.Const` for Modified (matching Python's Symbol list). Assignment updates now store `Definition` objects in `Clauses.Defs` (matching Python's `Clauses([], [Definition(...)], annot)` pattern). Fixed `collectAndList` to consume empty `And()` matching Python. Fixed `renameASTRec` to preserve original sort when replacement has TopSort. Updated all callers across transrel, actions, art, interp, bmc, mc, vmt, fragment, check, module.
3. §7.4 — `mk_assign_clauses` partial assignment ITE structure. **VERIFIED** correct — now also stores Definition in Clauses.Defs matching Python exactly (part of §6.1 refactor).

### High (affects conformance testing)
4. §1.7 — `EnumeratedSort.String()` — **VERIFIED** conformant (both return extension format `{ext1,ext2,...}`).
5. §1.2 — ForAll/Exists variable ordering (frozenset vs slice).
6. §3.11 — Complete `PrettyFmla` implementation for string conformance.
7. §4.4 — Clauses And-flattening.

### Medium (could cause subtle bugs)
8. §1.14 — Missing `EnumeratedSort.Constructors()` method.
9. §1.1 — Immutability / hashability differences (audit all set/dict usage with logic nodes).
10. §5.1 — Polymorphic symbol naming in Z3.
11. §5.3 — BV-aware comparison dispatch.
12. §8.1 — Three-pass compilation with forward references.
13. §10.1 — Initialization check with no-op initializer.

### Low
14. §5.4 — Z3 enum encoding.
15. §3.4 — `Symbol.__call__` zero-arg FunctionSort Apply creation.
16. §1.4 — Apply comma separator (verify actual Python output).
17. §1.6 — Unicode in symbol names (unlikely in practice).
