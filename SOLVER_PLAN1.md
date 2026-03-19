# Plan: Fix Solver Bugs 7.1, 7.2, 7.3

## Context

The Go port's `solver/` package has numerous incomplete or incorrect functions compared to Python's `ivy_solver.py`. These are catalogued in AUDIT18MARCH.md sections 7.1 (MISSING), 7.2 (STUB), and 7.3 (BEHAVIORAL_DIFFERENCE). Many are correctness-critical for verification — quantifier constraints, model extraction, binary encoding, and range-sort handling all affect whether Ivy proofs succeed or fail.

The work is organized into 8 phases by dependency and criticality.

## Files to Modify

| File | Changes |
|---|---|
| `z3bridge/translate.go` | Phase 1: quantifier constraint callback; Phase 8: variable naming |
| `solver/z3convert.go` | Phase 2: MyEq; Phase 3: Gebin; Phase 5: NumeralToZ3 clamping; Phase 6: bfeToZ3; Phase 8: SolverName |
| `solver/encoding.go` | Phase 4: EncodeTerm/EncodeEquality Z3-level; Phase 8: RangeSortBounds |
| `solver/herbrand.go` | Phase 7: clauseModelSimp early-return; Phase 8: SortCard, ClausesCase UnitRes |
| `solver/clauses.go` | Phase 7: ClausesModelToDiagram full rewrite |
| `solver/model.go` | Phase 7: ClausesModelToClausesWithModel use ModelFacts |
| `solver/compat.go` | Phase 8: CheckNativeCompatSym, ModelIfNone, QuantConstraints |
| `solver/solver.go` | Phase 1: wire QuantConstraintsFn; Phase 8: CheckSequence reporter |

## Python Reference

All references: `/Users/jaten/pyivy/ivy/ivy/ivy_solver.py`

## Implementation Phases

---

### Phase 1: Quantifier Bound Constraints (7.3.1) — P0

**Problem:** `ForAll`/`Exists` over nat/range-sorted variables are missing `0 <= x` / `lb <= x <= ub` constraints inside the quantifier body. Go's `QuantConstraints` (compat.go:303) only handles EnumeratedSort membership, not nat or RangeSort bounds.

**Python (lines 509-533):** `quant_constraints(vs, z3_vs)` checks `sig.interp[v.sort.name]`: if `'nat'` → add `0 <= z3_v`; if `RangeSort` and `handle_range_sorts` → add `lb <= z3_v` and `z3_v <= ub`. Then `forall()` wraps body in `Implies(And(constraints), body)` and `exists()` wraps in `And(constraints + [body])`.

**Changes:**

1. **`z3bridge/translate.go`**: Add callback field to `Translator`:
   ```go
   QuantConstraintsFn func(varSort logic.Sort, z3Var Expr) []Expr
   ```

2. **`z3bridge/translate.go:translateQuantifier`** (line 375): After creating `bound[i]` and before translating body, collect constraints. After `zBody` is computed, wrap:
   ```go
   var allConstraints []Expr
   for i, v := range variables {
       if t.QuantConstraintsFn != nil {
           cs := t.QuantConstraintsFn(v.VSort, bound[i])
           allConstraints = append(allConstraints, cs...)
       }
   }
   if len(allConstraints) > 0 {
       if isForall {
           zBody = t.Ctx.Implies(t.Ctx.And(allConstraints...), zBody)
       } else {
           zBody = t.Ctx.And(append(allConstraints, zBody)...)
       }
   }
   ```

3. **`solver/solver.go`**: In solver initialization (wherever `s.tr` is configured), wire:
   ```go
   s.tr.QuantConstraintsFn = func(varSort lg.Sort, z3Var z3bridge.Expr) []z3bridge.Expr {
       sortName := il.SortName(varSort)
       itp, ok := s.sig.Interp[sortName]
       // if itp == "nat": return []Expr{ctx.Le(ctx.IntVal(0), z3Var)}
       // if itp is *lg.RangeSort && HandleRangeSorts: return lb <= z3Var, z3Var <= ub
   }
   ```

---

### Phase 2: MyEq True/False Short-Circuit (7.3.7) — P0

**Problem:** `MyEq` doesn't simplify `x = True` → `x` or `x = False` → `Not(x)`.

**Python (lines 88-95):** `if z3.is_true(y): return x; if z3.is_false(y): return z3.Not(x)`

**File:** `solver/z3convert.go:963-968`

**Change:** Add before existing sort dispatch:
```go
if y.IsTrue() { return x }
if y.IsFalse() { return ctx.Not(x) }
```

z3bridge already has `Expr.IsTrue()` and `Expr.IsFalse()` (quantifier.go:1062,1072).

---

### Phase 3: Gebin Correct Semantics (7.3.8) — P0

**Problem:** Go's `Gebin(ctx, x, bound)` takes a single Z3 int expr. Python's `gebin(bits, n)` takes a `[]Bool` bit list (MSB first) and recursively encodes "bits >= n".

**Python (lines 1570-1578):** Recursive MSB-first: if `n==0` → true; if `n >= 2^len(bits)` → false; `hval = 2^(len-1)`; if `hval <= n` → `And(bits[0], gebin(rest, n-hval))`; else → `Or(bits[0], gebin(rest, n))`.

**File:** `solver/z3convert.go:977-987`

**Change:** Rewrite signature and body:
```go
func Gebin(ctx *z3bridge.Context, bits []z3bridge.Expr, n int) z3bridge.Expr {
    if n == 0 { return ctx.BoolVal(true) }
    if len(bits) == 0 || n >= (1<<len(bits)) { return ctx.BoolVal(false) }
    hval := 1 << (len(bits) - 1)
    if hval <= n {
        return ctx.And(bits[0], Gebin(ctx, bits[1:], n-hval))
    }
    return ctx.Or(bits[0], Gebin(ctx, bits[1:], n))
}
```

Update all callers (currently only EncodeEquality, which is rewritten in Phase 4).

---

### Phase 4: EncodeTerm & EncodeEquality (7.2.2, 7.2.3) — P1

**Problem:** Both produce wrong results. `EncodeTerm` only handles one case; `EncodeEquality` returns plain `lg.Eq` instead of binary-encoded Z3 equality.

**Python (lines 1587-1627):** `encode_term` returns `[z3Bool]*n` (bit-vector as Bool list). `encode_equality` zips encoded terms, conjuncts bit-equalities, adds `gebin` overflow guard.

**File:** `solver/encoding.go`

**Changes:**

1. Add Z3-level `(s *Solver) encodeTerm(t lg.Expr, n int, sort *lg.EnumeratedSort) ([]z3bridge.Expr, error)`:
   - `*lg.Ite`: translate cond, recurse then/else, element-wise `ctx.Ite`
   - Constructor (`sig.Constructors[sym.Name]`): find index in `sort.Extension`, `BinEnc(m,n)` → `ctx.BoolVal` list
   - `*lg.Variable`: create `n` Bool constants named `rep:sort:bit_index`
   - General function: create `n` Z3 Bool functions named `sym:bit_index`, apply to translated args

2. Add Z3-level `(s *Solver) encodeEquality(t1, t2 lg.Expr, sort *lg.EnumeratedSort) (z3bridge.Expr, error)`:
   - `n = sort.Card()`, `bits = CeilLog2(n)`
   - Encode both terms, zip bit-equalities, compute `Gebin(ctx, eterms[i], n-1)` for overflow
   - Return `ctx.Or(eqs, alt)`

3. Hook into translator: add `EnumEqualityFn func(t1, t2 lg.Expr, sort *lg.EnumeratedSort) (z3bridge.Expr, error)` callback on `Translator`. Wire in solver init when `!UseZ3Enums`.

---

### Phase 5: NumeralToZ3 Range Clamping (7.3.5) — P0

**Problem:** Numerals in range sorts aren't clamped to `[lb, ub]`.

**Python (lines 399-404):** `val = z3.If(val < lb, lb, z3.If(ub < val, ub, val))`

**File:** `solver/z3convert.go` — update `NumeralToZ3` (currently just delegates to `s.tr.Translate`).

**Change:** After translation, check if `num.CSort` is interpreted as a `*lg.RangeSort` via `s.sig.Interp`. If so and `HandleRangeSorts`:
```go
lb, ub := s.RangeSortBoundsToZ3(rs)
val = ctx.Ite(ctx.Lt(val, lb), lb, ctx.Ite(ctx.Lt(ub, val), ub, val))
```

Note: `RangeSortBoundsToZ3` already exists in z3convert.go:463-477.

---

### Phase 6: bfeToZ3 IntSort Input (7.2.7) — P0

**Problem:** `bfeToZ3` (z3convert.go:818-849) just does `ctx.Extract(hi, lo, args[0])` without handling IntSort inputs (needs `Int2Bv`), BV size overflow, zero-width, or zero-extension.

**Python (lines 174-209):** Full decision tree: check `insort == IntSort()`, handle BV size clamping, output-sort matching.

**File:** `solver/z3convert.go:818-849`

**Change:** Rewrite to match Python's decision tree:
1. Extract domain/range sorts from `sym.CSort.(*lg.FunctionSort)`
2. Translate to Z3 sorts
3. Check if domain is IntSort → use `ctx.Int2Bv(hi+1, x)` before Extract
4. Clamp `hi` if `insort.size() <= hi` for BV inputs
5. Handle output: IntSort → `ctx.Bv2Int(Extract(...), false)`, BV → check zero-extension, zero-width
6. z3bridge has `Int2Bv`, `Bv2Int`, `BvSortSize` (quantifier.go:660,674,742)

---

### Phase 7: Model Extraction (7.3.2, 7.2.4, 7.2.5) — P0/P1

#### 7a: clauseModelSimp early-return (7.3.2)

**Problem:** Go keeps all true/unknown literals. Python returns `[l]` (just the first true ground literal), causing early exit.

**Python (lines 1062-1076):** `if z3.is_true(v): return [l]` — single-literal early return.

**File:** `solver/herbrand.go:741-776`

**Change:** At line 762, when `val.IsTrue()` (use `IsTrue()` instead of string comparison), return the single literal immediately:
```go
if val.IsTrue() {
    return lit  // early return: clause satisfied by this literal
}
```

#### 7b: ClausesModelToDiagram (7.2.4)

**Problem:** Creates `sym = sym` (tautologies). Should use `ModelFacts` + `NumeralAssign` + `SubstituteConstantsClauses` + `FilterRedundantFacts`.

**Python (lines 1448-1516):** `model_if_none` → `model_facts` → `numeral_assign` → `substitute_constants_clauses` → `filter_redundant_facts` → optional weakening.

**File:** `solver/clauses.go:145-194`

**Change:** Replace the body with:
1. `h := s.ModelIfNone(AndClauses(clauses, axioms), implied, nil)` — get HerbrandModel
2. `res := ModelFacts(h, ignore, clauses, true)` — extract facts (already handles constants, relations, functions)
3. `reps := NumeralAssignWithClauses(h, res)` — assign numeral names
4. Build substitution map, call `clauseops.SubstituteConstantsClauses(res, subs)` — need to check this exists
5. Filter skolem-defined symbols (Python line 1500)
6. `res, _ = s.FilterRedundantFacts(res, axioms)` — already implemented
7. Return res

`ModelFacts` is at herbrand.go:419. `NumeralAssignWithClauses` is at compat.go:180. `SubstituteConstantsClauses` needs to be verified — search for it in clauseops.

#### 7c: ClausesModelToClausesWithModel (7.2.5)

**Problem:** Only evaluates 0-arity symbols directly. Should use `ModelFacts` for the full extraction, then do numeral/prefix renaming and substitution.

**Python (lines 1373-1395):** `model_facts(h, ignore, clauses)` → if numerals: `numeral_assign(res, h)` else prefix `__` → `substitute_constants_clauses(res, m)`.

**File:** `solver/model.go:336-401`

**Change:** Replace the symbol-loop (lines 357-401) with:
1. Build `HerbrandModel` from the `ModelResult`
2. Call `ModelFacts(h, ignore, clauses, false)`
3. Apply renaming (numeral or `__` prefix)
4. Substitute constants

---

### Phase 8: P1/P2 Remaining Items

#### 8a: SortCard BV/range (7.3.11)
**File:** `solver/herbrand.go:623-630`
**Change:** Add BV case (`2^size`), RangeSort with numeral ub case (`int(ub)+1`). Needs `s *Solver` receiver to access `sig.Interp` and translator for sort→Z3.

#### 8b: ClausesCase UnitRes (7.3.3)
**File:** `solver/herbrand.go:693-734`
**Change:** After model simplification, convert clauses to `[][]*unitres.Literal`, call `unitres.NewUnitRes(...)`, `r.Propagate(nil)`, collect `r.UsedUnitLiterals()` + remaining clauses, convert back to `[]lg.Expr`. The `unitres` package exists with full `UnitRes`, `Propagate`, etc.

#### 8c: ModelIfNone simultaneous sort search (7.3.4)
**File:** `solver/compat.go:96-129`
**Change:** Currently delegates to `GetSmallModel` which minimizes each sort independently (model.go:199-213 loops per sort). Python searches all sorts at the same size N simultaneously. Change `ModelIfNone` to do its own loop: for N=1,2,3... push, add `SortSizeConstraint(sort, N)` for ALL uninterpreted sorts, check, pop.

#### 8d: RangeSortBounds stub (7.2.1)
**File:** `solver/encoding.go:248-256`
**Change:** Parse `rs.Lb`/`rs.Ub` as integers (like `RangeSortBoundsToZ3` in z3convert.go already does), instead of returning `(0, MaxInt32)`.

#### 8e: CheckNativeCompatSym invoke (7.2.6)
**File:** `solver/compat.go:16-33`
**Change:** Make it `(s *Solver) CheckNativeCompatSym`. After sort checks pass, create dummy Z3 args, invoke `LookupNative`, check returned value's sort matches declared range sort.

#### 8f: SolverName z3_builtins error (7.2.8)
**File:** `solver/z3convert.go:898-901`
**Change:** Expand `bit0`/`bit1` check to the full Python `z3_builtins` set. Return error instead of empty string for collisions.

#### 8g: Variable naming :sort_name (7.3.9)
**File:** `z3bridge/translate.go:389`
**Change:** When creating Z3 const for a `*logic.Variable`, use `v.Name + ":" + sortName` as the Z3 name (not just `v.Name`). Python: `sksym = term.rep + ':' + term.sort.name`.

#### 8h: CheckSequence reporter (7.3.10)
**File:** `solver/solver.go:711-738`
**Change:** Add `Reporter` interface with `Start(isAssert bool, doc string)` and `End(result bool, doc string) bool` (returns false to abort). Add optional reporter parameter to `CheckSequence`.

#### 8i: filter_redundant_facts activation literals (7.3.6)
**Assessment:** Go uses push/pop, Python uses activation literals. Both are correct. Low priority, leave as-is unless performance is an issue.

## Existing Functions to Reuse

- `z3bridge.Expr.IsTrue()` / `.IsFalse()` — quantifier.go:1062,1072
- `z3bridge.Context.Int2Bv(width, e)` — quantifier.go:674
- `z3bridge.Context.Bv2Int(e, signed)` — quantifier.go:660
- `z3bridge.Context.BvSortSize(s)` — quantifier.go:742
- `ModelFacts(h, ignore, clauses, upclose)` — herbrand.go:419
- `NumeralAssignWithClauses(model, clauses)` — compat.go:180
- `RangeSortBoundsToZ3` — z3convert.go:463
- `unitres.NewUnitRes(clauses)` / `.Propagate()` — unitres/unitres.go:766,1200
- `clauseops.SubstituteConstantsClauses` — needs verification

## Verification

After each phase: `cd /Users/jaten/go/src/github.com/glycerine/goivy && make test`

Key test file: `solver/solver_test.go` (1359 lines of existing tests).

Add targeted unit tests for each phase before implementing (TDD):
- Phase 1: Test that `ForAll(X:nat, P(X))` translates to `ForAll(X, Implies(0<=X, P(X)))`
- Phase 2: Test `MyEq(ctx, x, ctx.BoolVal(true))` returns `x`
- Phase 3: Test `Gebin(ctx, [a,b], 2)` returns `And(a, ...)`
- Phase 7: Test `clauseModelSimp` returns single literal when true in model
