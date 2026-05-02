# Plan: Z3 Enumerated Types as Native Sorts (todo.md item 8)

**Created:** 2026-05-02 07:12 UTC

## Context

The todo says Go's `enumerated_to_numeral` is unimplemented while Python converts
enums to Z3 native sorts. Investigation reveals **both** Python (ivy_solver.py:478-480)
and Go (translate.go:1135-1142) raise the same error for interpreted enum sorts:
`"Cannot interpret enumerated type as a native sort (not yet supported)"`.

The normal enum → Z3 `EnumSort` path **already works** in both languages. The stub
is only for the "interpreted" case (`interpret color -> int`).

However, two real Go-vs-Python divergences exist in `solver_herbrand.go` where Go
is missing `not use_z3_enums` guards. Additionally, a dead `EnumeratedToNumeral`
function exists in `solver_encoding.go`.

**Scope:** Fix the two divergences + implement the interpreted-enum feature behind
a `const EnableInterpretedEnums = false` flag so it doesn't affect xtrace conformance
testing until the port is verified complete.

---

## Part 1: Fix Go-vs-Python Divergences

### 1A. Add `!UseZ3Enums` guard to `getModelConstant`

**File:** `z3bridge/solver_herbrand.go:363`

**Current (line 363):**
```go
if es, ok := sort.(*lg.EnumeratedSort); ok {
```

**Python (ivy_solver.py:1055):**
```python
if isinstance(s,ivy_logic.EnumeratedSort) and not use_z3_enums:
```

**Change to:**
```go
if es, ok := sort.(*lg.EnumeratedSort); ok && (h.tr.s == nil || !h.tr.s.opts.UseZ3Enums) {
```

The nil guard handles test scenarios where no high-level solver is attached.

When `UseZ3Enums=true` (the default), Go currently wastefully iterates all enum
values checking equality. Python falls through to the general `constant_from_z3`
path which does a single `m.eval(term_to_z3(t), model_completion=True)` — for
Z3 native enums, this returns the enum constant directly (e.g., `"red"`), and
`constantFromZ3` creates the correct `lg.Const` from its string representation.

### 1B. Add `!UseZ3Enums` guard to `FunctionModelToClauses`

**File:** `z3bridge/solver_herbrand.go:689`

**Current (line 689):**
```go
if es, ok := rng.(*lg.EnumeratedSort); ok {
```

**Python (ivy_solver.py:1800):**
```python
if isinstance(rng,ivy_logic.EnumeratedSort) and not use_z3_enums:
```

**Change to:**
```go
if es, ok := rng.(*lg.EnumeratedSort); ok && (h.tr.s == nil || !h.tr.s.opts.UseZ3Enums) {
```

Same rationale: when native Z3 enums are active, the general path (line 699-702)
works correctly because Z3 knows the enum sort natively.

### 1C. Remove dead `EnumeratedToNumeral` function

**File:** `z3bridge/solver_encoding.go:435-447`

`EnumeratedToNumeral()` is defined but never called anywhere (confirmed by grep).
Remove it. Part 2 introduces a different implementation behind the feature flag.

---

## Part 2: Implement Interpreted Enum Feature (Flag-Gated)

### 2A. Add feature flag constant

**File:** `z3bridge/translate.go` (top of file, after imports)

```go
// EnableInterpretedEnums gates the enumerated_to_numeral feature: converting
// enum constants to integer ordinals when the enum sort has a numeric
// interpretation (e.g., "interpret color -> int"). This is OFF during
// xtrace conformance testing against Python (which also has this stubbed).
// Flip to true once the Go port is verified faithful.
const EnableInterpretedEnums = false
```

### 2B. Implement `enumerated_to_numeral` in `translateVarOrConst`

**File:** `z3bridge/translate.go:1135-1142`

Replace the current error stub with:

```go
} else if es, ok := sort.(*lg.EnumeratedSort); ok && t.s != nil && t.s.sig != nil {
    if _, interped := t.s.sig.Interp[es.Name]; interped {
        xtracer.Trace("ivy_solver.py:441 enumerated_to_numeral() ENTER")
        if !EnableInterpretedEnums {
            return Expr{}, fmt.Errorf("cannot interpret enumerated type %q as a native sort (not yet supported)", es.Name)
        }
        // Convert enum constant name to its ordinal index
        ordinal := -1
        for i, eName := range es.Extension {
            if eName == name {
                ordinal = i
                break
            }
        }
        if ordinal < 0 {
            return Expr{}, fmt.Errorf("enum constant %q not found in sort %q extension", name, es.Name)
        }
        // Translate as numeral in the interpreted sort's Z3 representation
        return t.enumeratedToNumeralZ3(ordinal, es)
    }
}
```

### 2C. Add `enumeratedToNumeralZ3` helper

**File:** `z3bridge/translate.go` (new method near the stub site)

```go
// enumeratedToNumeralZ3 converts an enum ordinal to a Z3 expression in the
// sort that the enum is interpreted as.
// Matches Python's enumerated_to_numeral (when it is eventually implemented).
func (t *Translator) enumeratedToNumeralZ3(ordinal int, es *lg.EnumeratedSort) (Expr, error) {
    itp := t.s.sig.Interp[es.Name]
    switch v := itp.(type) {
    case string:
        switch v {
        case "int", "nat":
            return t.Ctx.IntVal(ordinal), nil
        default:
            if strings.HasPrefix(v, "bv[") {
                // Extract bit width from "bv[N]"
                widthStr := v[3 : len(v)-1]
                width, err := strconv.Atoi(widthStr)
                if err != nil {
                    return Expr{}, fmt.Errorf("bad bv width in interpretation of %q: %v", es.Name, err)
                }
                if ordinal >= (1 << width) {
                    return Expr{}, fmt.Errorf("enum ordinal %d exceeds bv[%d] capacity", ordinal, width)
                }
                return t.Ctx.BitVecVal(ordinal, width), nil
            }
            return Expr{}, fmt.Errorf("cannot interpret enum %q as native sort %q", es.Name, v)
        }
    case *lg.RangeSort:
        // RangeSort is backed by Z3 IntSort; ordinal maps directly.
        return t.Ctx.IntVal(ordinal), nil
    default:
        return Expr{}, fmt.Errorf("cannot interpret enum %q: unsupported interpretation type %T", es.Name, itp)
    }
}
```

### 2D. Add reverse mapping for model extraction (flag-gated)

**File:** `z3bridge/solver_herbrand.go` — in `getModelConstant`, after the general case

After line 392 (`return constantFromZ3(sort, val)`), add a flag-gated check:

```go
// General case
zt, err := h.tr.Translate(c)
if err != nil {
    return lg.NewConst("?", sort)
}
val, ok := h.model.Eval(zt, true)
if !ok {
    return lg.NewConst("?", sort)
}
result := constantFromZ3(sort, val)

// When EnableInterpretedEnums is active and this is an interpreted enum,
// the model value is a numeral — map it back to the enum constant name.
if EnableInterpretedEnums {
    if es, ok2 := sort.(*lg.EnumeratedSort); ok2 && h.sig != nil {
        if _, interped := h.sig.Interp[es.Name]; interped {
            if idx, err := strconv.Atoi(result.Name); err == nil && idx >= 0 && idx < len(es.Extension) {
                return lg.NewConst(es.Extension[idx], sort)
            }
        }
    }
}
return result
```

---

## Part 3: Tests

All tests go in `z3bridge/solver_test.go` (appended) or `z3bridge/solver2_test.go`.

### Test 3A: `TestGetModelConstant_NativeEnum`
- Create enum sort `color = {red, green, blue}`
- Create solver with `UseZ3Enums=true`
- Add constraint `x = green` where `x` is of sort color
- Check SAT, extract model
- Call `getModelConstant(x)`, verify it returns `green`
- This tests the fix from 1A: the general path (not the iteration path)

### Test 3B: `TestGetModelConstant_BinaryEncoding`
- Same setup but `UseZ3Enums=false`
- Verify the iteration path works correctly
- This is a regression test for the existing iteration code

### Test 3C: `TestFunctionModelToClauses_NativeEnum`
- Create function `f: node -> color` where color is enum
- Constrain `f(n1) = red`
- Check SAT, extract model
- Call `FunctionModelToClauses`, verify correct clauses with `UseZ3Enums=true`

### Test 3D: `TestFunctionModelToClauses_BinaryEncoding`
- Same with `UseZ3Enums=false`

### Test 3E: `TestEnumeratedToNumeral_FlagOff`
- With `EnableInterpretedEnums=false` (the default)
- Create `type color = {red, green, blue}` with `interpret color -> int`
- Attempt to translate `red` — verify the original error is returned
- Regression test ensuring flag-off behavior matches Python

### Test 3F: `TestEnumeratedToNumeral_FlagOn`
- Since the flag is a `const`, this test uses a separate helper that
  takes a boolean parameter simulating the flag
- OR: test the `enumeratedToNumeralZ3` method directly (it doesn't check the flag)
- Verify `red` → `IntVal(0)`, `green` → `IntVal(1)`, `blue` → `IntVal(2)`
- Verify with `interpret color -> bv[2]`: `red` → `BitVecVal(0,2)`
- Verify overflow: 5-element enum with `bv[2]` errors for ordinal >= 4

### Test 3G: `TestEnumeratedToNumeral_RangeSort`
- Test with `interpret color -> {0..5}` (RangeSort interpretation)
- Verify `red` → `IntVal(0)`, `green` �� `IntVal(1)`, `blue` → `IntVal(2)`

---

## File Summary

| File | Changes |
|------|---------|
| `z3bridge/translate.go` | Add `EnableInterpretedEnums` const; replace stub with flag-gated implementation; add `enumeratedToNumeralZ3` helper |
| `z3bridge/solver_herbrand.go` | Fix `getModelConstant` guard (line 363); fix `FunctionModelToClauses` guard (line 689); add flag-gated reverse mapping in `getModelConstant` |
| `z3bridge/solver_encoding.go` | Remove dead `EnumeratedToNumeral` function (lines 435-447) |
| `z3bridge/solver_test.go` | Add tests 3A-3G |

## Verification

Run: `cd ~/ivy/goivy && make test`

All existing xtrace conformance tests must continue to pass unchanged (the flag is off).
New tests verify: (a) divergence fixes are correct, (b) the feature works when tested directly, (c) the error stub still fires when the flag is off.
