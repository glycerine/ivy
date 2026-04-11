# Plan: Add Fine-Grained Callsite Instrumentation to TranslateSort / uninterpretedsort

Created: 2026-04-10 23:30 UTC

## Context

`make golden` now diverges at `~/ivy/goivy/log.red` line 247715:

```
247714  go : XTRACE: ivy_solver.py:95 my_eq() ENTER
        py : XTRACE: ivy_solver.py:95 my_eq() ENTER

247715  go : XTRACE: ivy_solver.py:263 uninterpretedsort() ENTER name=lclock
        py : XTRACE: ivy_solver.py:95 my_eq() ENTER
```

Both sides agree on `my_eq() ENTER` at 247714. Then Go performs an extra `TranslateSort(lclock)` call that Python does not, while Python proceeds directly to the next `my_eq() ENTER`.

The earlier work in this session (Z3SessionCache) **fixed** the original divergence at line 247689 (Go's `sorts=[]` vs Python's 11 cached sorts). Both sides now show identical sort caches at 247689, 247694, 247699, 247708. The new divergence at 247715 is a **separate downstream bug** — the previous plan's success exposed it.

We have **stopped guessing**. We are spending too long staring at the trace trying to deduce which call site emitted the extra trace. Instead, we will instrument every TranslateSort / `.to_z3()` call site in both Go and Python with a unique `callsite=<name>` tag. After re-running `make golden`, the divergent line will name the responsible call site directly.

## Goal

Make every Z3 sort-translation trace self-identify its caller. After this change, an `uninterpretedsort() ENTER name=lclock` trace becomes:

```
XTRACE: ivy_solver.py:263 uninterpretedsort() ENTER name=lclock callsite=translateQuantifier
```

The `callsite=` value is a stable, semantic name shared by Go and Python at matching call sites. When the trace diverges, the divergent line names exactly which Go path is firing that Python skips.

## Approach

### Strategy: emit a callsite trace IMMEDIATELY BEFORE each `.to_z3()` / `TranslateSort` call

We do **not** change the signature of `TranslateSort` (Go) or `uninterpretedsort` (Python). Instead, we add a single `xtracer` line right before each call site. The format is identical in both languages:

```
XTRACE: TranslateSort_call callsite=<name>
```

This produces TWO traces per sort lookup (the new callsite line + the existing `uninterpretedsort ENTER` line). When the divergence appears, both lines tell us:
1. Which logical call site is firing (`callsite=<name>`)
2. What sort it's looking up (`name=<sortname>`)

### Why a separate trace, not a function-signature change

- `Sort.to_z3()` in Python is monkey-patched. Adding a parameter would require changing every monkey-patched method (`UninterpretedSort.to_z3`, `FunctionSort.to_z3`, `EnumeratedSort.to_z3`, `RangeSort.to_z3`, `BooleanSort.to_z3`) AND every call site.
- A thread-local "current callsite" variable would be a package global (forbidden by CLAUDE.md section C).
- A wrapper function `to_z3_with_callsite(sort, name)` would still need every call site updated, AND would need a wrapper helper.

A bare trace at each call site has zero coupling, is mechanical, and is trivially reversible if we want to remove it later.

### Naming convention

Each Go/Python call-site pair uses the **same** callsite name. Names are lowercase, underscored, semantically descriptive — they read as "the function name + which sub-call". The shared vocabulary makes diffs trivial.

| Callsite name | Semantics | Python (`ivy_solver.py`) | Go (`goivy/z3bridge/`) |
|---|---|---|---|
| `term_to_z3_variable` | Variable case in `term_to_z3` | line 470 | `translate.go:961` (translateVariable) |
| `term_to_z3_const` | Individual constant case | line 491 | `translate.go:1033` (translateVarOrConst first-order) |
| `term_to_z3_func` | Function-application sig build | line 505 | (n/a — Go uses makeFuncDecl) |
| `functionsort_dom` | Domain sort in `functionsort` | line 283 | `translate.go:1117` |
| `functionsort_rng` | Range sort in `functionsort` | line 284 | `translate.go:1134` |
| `lt_pred` | `<` predicate sig build | line 517 | (n/a — Go uses functionSort directly inside ltPred) |
| `atom_to_z3_relation` | Relation sig build in `atom_to_z3` | line 541 | (n/a — Go atomToZ3 calls functionSort directly) |
| `symbol_to_z3_const` | 0-ary symbol in `symbol_to_z3` | line 302 | `translate.go:1045` (translateVarOrConst 0-ary fn) |
| `symbol_to_z3_func` | n-ary symbol in `symbol_to_z3` | line 302 | `translate.go:1071` (translateVarOrConst higher-order) |
| `lookup_native_enum_or_range` | Enum/Range sort dispatch | line 357 | `solver_z3convert.go:651, 657` |
| `lookup_native_arrcst` | `arrcst` polymorphic | line 334 | (Go-specific path) |
| `bfe_to_z3_dom` | bfe domain sort | line 200 | `solver_z3convert.go:935` |
| `bfe_to_z3_rng` | bfe range sort | line 201 | `solver_z3convert.go:936` |
| `sort_name_to_z3` | Top-level sort lookup | line 123 | `solver_z3convert.go:1197` |
| `sort_card` | `sort_card()` lookup | line 399 | (n/a unless Go has equivalent) |
| `native_symbol_const` | `native_symbol` 0-ary | line 414 | (n/a unless Go has equivalent) |
| `native_symbol_func` | `native_symbol` n-ary | line 415 | (n/a unless Go has equivalent) |
| `numeral_to_z3` | Uninterpreted numeral fallback | line 435 | (Go has Numeral but goes through different path) |
| `encode_term_relation_sort` | RelationSort.to_z3 in encode_term | line 1780 | `solver_encoding.go:143` |
| `z3_function_dom` | Z3Function domain build | (n/a — Python builds inline) | `solver_encoding.go:211` |
| `z3_function_rng` | Z3Function range build | (n/a) | `solver_encoding.go:217` |
| `check_native_compat_dom` | Compat check domain | (no sub-call) | `solver_compat.go:86` |
| `check_native_compat_rng` | Compat check range | (no sub-call) | `solver_compat.go:95` |
| **`translate_quantifier_var`** | **Bound-var sort in quantifier** | **(n/a — Python doesn't TranslateSort here)** | **`translate.go:1186` ← suspected redundant call** |

The "`(n/a)`" entries on the Python side are call sites that exist in Go but have no Python counterpart. **These are the prime suspects** — if Go fires `callsite=translate_quantifier_var` at the divergence point and Python emits no corresponding trace, that's the bug.

## Architecture

### Modify: `goivy/z3bridge/translate.go` (Go side)

For each `t.TranslateSort(...)` call, add an `xtracer.Trace` line immediately before it. Example for `translateQuantifier:1186`:

```go
for i, v := range variables {
    xtracer.Trace("TranslateSort_call callsite=translate_quantifier_var")
    zs, err := t.TranslateSort(v.VSort)
    if err != nil {
        return Expr{}, err
    }
    ...
}
```

Apply the same edit at every TranslateSort call site listed below. Use `xtracer.Trace`, not a custom helper. The trace string is **exactly** `"TranslateSort_call callsite=<name>"` — the format is identical to Python's.

Sites in `translate.go`:
- line 961 (translateVariable): `callsite=term_to_z3_variable`
- line 1033 (translateVarOrConst first-order): `callsite=term_to_z3_const`
- line 1045 (translateVarOrConst 0-ary fn): `callsite=symbol_to_z3_const`
- line 1071 (translateVarOrConst higher-order range): `callsite=symbol_to_z3_func`
- line 1117 (functionSort dom): `callsite=functionsort_dom`
- line 1134 (functionSort rng): `callsite=functionsort_rng`
- line 1186 (translateQuantifier): `callsite=translate_quantifier_var`

Sites in `solver_z3convert.go`:
- line 651 (lookupNative enum): `callsite=lookup_native_enum_or_range`
- line 657 (lookupNative range): `callsite=lookup_native_enum_or_range`
- line 935 (bfe dom): `callsite=bfe_to_z3_dom`
- line 936 (bfe rng): `callsite=bfe_to_z3_rng`
- line 1197 (SortNameToZ3): `callsite=sort_name_to_z3`

Sites in `solver_encoding.go`:
- line 143 (EncodeTermZ3 dom): `callsite=encode_term_relation_sort`
- line 211 (Z3Function dom): `callsite=z3_function_dom`
- line 217 (Z3Function rng): `callsite=z3_function_rng`

Sites in `solver_compat.go`:
- line 86 (CheckNativeCompatSym dom): `callsite=check_native_compat_dom`
- line 95 (CheckNativeCompatSym rng): `callsite=check_native_compat_rng`

### Modify: `pyivy/ivy/ivy/ivy_solver.py` (Python side)

For each `.to_z3()` call on a sort, add an `xtracer.trace` line immediately before. Example for `term_to_z3` line 491:

```python
else:
    iso = term.rep.sort
    if __debug__: xtracer.trace("TranslateSort_call callsite=term_to_z3_const")
    sig = iso.to_z3() if iso is not None else S
    res = z3.Const(solver_name(term.rep), sig)
```

Apply the edit at every `.to_z3()` call site that operates on a sort:
- line 123 (sort_name_to_z3 → `sort.to_z3()`): `callsite=sort_name_to_z3`
- line 200 (bfe_to_z3 dom): `callsite=bfe_to_z3_dom`
- line 201 (bfe_to_z3 rng): `callsite=bfe_to_z3_rng`
- line 283 (functionsort dom): `callsite=functionsort_dom`
- line 284 (functionsort rng): `callsite=functionsort_rng`
- line 302 (symbol_to_z3 dispatch — emits BOTH callsite and the .to_z3 call):
  - `if s.sort.dom == []`: `callsite=symbol_to_z3_const`
  - `else`: `callsite=symbol_to_z3_func`
- line 334 (lookup_native arrcst): `callsite=lookup_native_arrcst`
- line 357 (lookup_native enum/range): `callsite=lookup_native_enum_or_range`
- line 399 (sort_card): `callsite=sort_card`
- line 414 (native_symbol const): `callsite=native_symbol_const`
- line 415 (native_symbol func): `callsite=native_symbol_func`
- line 435 (numeral_to_z3): `callsite=numeral_to_z3`
- line 470 (term_to_z3 variable): `callsite=term_to_z3_variable`
- line 491 (term_to_z3 individual const): `callsite=term_to_z3_const`
- line 505 (term_to_z3 function): `callsite=term_to_z3_func`
- line 517 (lt_pred): `callsite=lt_pred`
- line 541 (atom_to_z3 relation): `callsite=atom_to_z3_relation`
- line 1780 (encode_term RelationSort): `callsite=encode_term_relation_sort`

### Format requirement

Both Go and Python emit the **exact same** trace string format:

```
XTRACE: TranslateSort_call callsite=<name>
```

No file:line prefix, no caller info — just the literal string above. This is critical so the diff aligns. Diff sensitivity: every byte matters.

## What we expect to see

Around the current divergence point (line 247714 my_eq → 247715 Go-only uninterpretedsort), the new instrumented trace should look like:

```
247714  go : XTRACE: ivy_solver.py:95 my_eq() ENTER
        py : XTRACE: ivy_solver.py:95 my_eq() ENTER

247715  go : XTRACE: TranslateSort_call callsite=<X>          ← TELLS US WHICH GO CALL SITE
        py : XTRACE: ivy_solver.py:95 my_eq() ENTER

247716  go : XTRACE: ivy_solver.py:263 uninterpretedsort() ENTER name=lclock
        py : ...
```

Whatever `<X>` turns out to be, **that** is the bug location. We then read the surrounding code, compare with the matching Python path, and remove or repair the redundant call. The hypothesis (validated or refuted by `<X>`) is `translate_quantifier_var`.

## Files affected

### Modified files (no new files)
- `goivy/z3bridge/translate.go` — 7 callsite traces
- `goivy/z3bridge/solver_z3convert.go` — 5 callsite traces
- `goivy/z3bridge/solver_encoding.go` — 3 callsite traces
- `goivy/z3bridge/solver_compat.go` — 2 callsite traces
- `pyivy/ivy/ivy/ivy_solver.py` — ~17 callsite traces

Total: ~34 single-line additions across 5 files. Pure trace additions, no logic changes.

### Critical files to read before editing
- `goivy/z3bridge/translate.go:955-1200` — Translator helpers and TranslateSort call sites
- `goivy/z3bridge/solver_z3convert.go:560-1200` — secondary call sites
- `goivy/z3bridge/solver_encoding.go:130-225` — encode_term and Z3Function
- `goivy/z3bridge/solver_compat.go:80-100` — compat checks
- `pyivy/ivy/ivy/ivy_solver.py:120-545` — main term/atom/sort code paths
- `pyivy/ivy/ivy/ivy_solver.py:1770-1790` — encode_term

## Verification

### Step 1: Compile and run the unit tests
```
cd ~/ivy/goivy && go build ./... && go test ./z3bridge/...
```
Trace-only additions cannot break anything. If they do, the format string is wrong.

### Step 2: Re-run `make golden`
```
cd ~/ivy/goivy && make golden 2>&1 | tee log.parser.red
```

Inspect `~/ivy/goivy/log.red` around line 247715 (which may have shifted by ~30 lines due to the new traces inserted earlier in the run).

### Step 3: Read the divergent callsite name

The diverging Go-side trace will be:
```
XTRACE: TranslateSort_call callsite=<NAME>
```

Look up `<NAME>` in the table above. That row points at the file:line in Go where the redundant call lives. Compare with the Python side:
- If Python has the same `callsite=<NAME>` in its trace immediately before, the call is shared but the underlying cache state diverges.
- If Python emits NOTHING for `<NAME>` (the next Python trace is something else), the call site is **Go-only** — the bug is that Go is calling a code path Python doesn't.

### Step 4: Fix the underlying bug

Once we know `<NAME>`, the fix is local:
- If `<NAME>` = `translate_quantifier_var` (the leading hypothesis): remove the unconditional `t.TranslateSort(v.VSort)` at `translate.go:1186` (the `_ = zs` is a tell that the result is unused). `translateVariable(v)` already calls `TranslateSort` from inside its cache-miss branch (`translate.go:961`), so the cache is still populated correctly when actually needed.
- If `<NAME>` = some other name: read the corresponding Go function, identify what extra work it does compared to its Python counterpart, and align them.

### Step 5: Remove the instrumentation (optional)

After the underlying divergence is fixed, the callsite traces can either:
(a) Stay forever, as additional regression coverage (per the user's "Accrete xtraces, never delete" rule).
(b) Be removed in a follow-up commit if they create too much noise.

**Default: keep them.** The user has been explicit that xtraces accrete in this codebase. They'll catch the next divergence even faster.

## Risks and rollback

- **Risk: trace format mismatch**. If Go and Python use even slightly different format strings (e.g., extra space, different quoting), the lines won't match and the diff will explode. Mitigation: copy-paste the exact format string `"TranslateSort_call callsite=<name>"` from this plan into both files. No interpolation, no f-strings on variable parts of the format — only the `<name>` is variable.
- **Risk: missed call sites**. If we instrument 17 of 18 sites and the divergence is on the missed one, we'll see a Go trace with NO matching Python callsite, AND no Go callsite either — just the bare `uninterpretedsort ENTER`. That tells us we have a missing call site to instrument. Add it and re-run.
- **Risk: trace volume**. The 17 extra traces per Z3 translation add ~17× to the log size. For a single divergence run, this is fine; for long-term retention, consider option (b) of step 5.
- **Rollback**: pure trace additions are trivially revertible — `git checkout` the 5 files.

## Out of scope

- **Fixing the divergence itself.** This plan only adds visibility. The actual fix is a follow-up step once we know which call site is responsible (Step 4 of Verification).
- **Refactoring TranslateSort to take a callsite parameter.** Considered and rejected — see "Why a separate trace, not a function-signature change" above.
- **Removing other unrelated `_ = zs` patterns.** Even though the `translate.go:1186` `_ = zs` is the leading hypothesis, we don't fix it preemptively. We wait for the trace to confirm.
- **Instrumenting `enumeratedsort` / `BooleanSort` / `RangeSort` traces.** The current divergence is on `uninterpretedsort lclock`. If a later divergence implicates a different sort kind, expand the instrumentation then.
