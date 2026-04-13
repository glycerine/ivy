# Fix quantifier sort-annotation placement in dropAnnotations
Created: 2026-04-13 (current session)

## Context

When printing formulas, Python produces `forall T. (T:lclock <= _T & ...)` while Go produces `forall T:lclock. (T <= _T & ...)`. The sort annotation appears on the quantifier binding in Go but on the first body occurrence in Python. Python is the source of truth.

This is a **true AST-level divergence** in the `dropAnnotations` pass (not just cosmetic). The root cause is that Go processes bound variables **before** the body, while Python processes the body **first**.

## Root Cause

`dropAnnotations` uses a shared mutable `annotatedVars` set/map. The processing order determines which occurrence keeps the annotation:

**Python** (`ivy_logic.py:1434-1436`):
```python
def quant_drop_annotations(self, inferred_sort, annotated_vars):
    body = self.body.drop_annotations(True, annotated_vars)   # body FIRST
    return type(self)([v.drop_annotations(False, annotated_vars) for v in self.variables], body)
```

**Go** (`logic/pretty.go:287-298`):
```go
case *ForAll:
    // variables FIRST (WRONG)
    vars := ...
    for i, v := range t.Variables {
        dv := dropAnnotations(v, false, annotatedVars)
        ...
    }
    // body SECOND (WRONG)
    body := dropAnnotations(t.Body, true, annotatedVars)
    return &ForAll{Variables: vars, Body: body}
```

### Why the order matters (trace for `forall T:lclock. (T:lclock <= _T)`)

**Python (body first):**
1. Body: `<=` is polymorphic, returns Boolean → `inferredSort && !Boolean` = `False` → first `T:lclock` keeps annotation, adds T to annotatedVars
2. Bound var T: already in annotatedVars → annotation stripped → `T`
3. **Result:** `forall T. (T:lclock <= _T & ...)`

**Go (vars first):**
1. Bound var T: `inferredSort=false`, T not in annotatedVars, sort not TopSort → keeps annotation, adds T to annotatedVars
2. Body: T already in annotatedVars → all occurrences stripped
3. **Result:** `forall T:lclock. (T <= _T & ...)`

## Fix

**File:** `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/pretty.go`

For `ForAll`, `Exists`, and `Lambda` cases (lines 287-324): swap the order to process **body first**, then bound variables. This matches Python's `quant_drop_annotations`.

**Do NOT change `NamedBinder`** (lines 326-337) — Python's NamedBinder lambda evaluates vars before body (due to left-to-right argument evaluation), so Go's current order is already correct there.

### Before (ForAll, lines 287-298):
```go
case *ForAll:
    vars := make([]*Variable, len(t.Variables))
    for i, v := range t.Variables {
        dv := dropAnnotations(v, false, annotatedVars)
        if vv, ok := dv.(*Variable); ok { vars[i] = vv } else { vars[i] = v }
    }
    body := dropAnnotations(t.Body, true, annotatedVars)
    return &ForAll{Variables: vars, Body: body}
```

### After:
```go
case *ForAll:
    body := dropAnnotations(t.Body, true, annotatedVars)
    vars := make([]*Variable, len(t.Variables))
    for i, v := range t.Variables {
        dv := dropAnnotations(v, false, annotatedVars)
        if vv, ok := dv.(*Variable); ok { vars[i] = vv } else { vars[i] = v }
    }
    return &ForAll{Variables: vars, Body: body}
```

Apply the same swap to `Exists` (lines 300-311) and `Lambda` (lines 313-324).

## Unit Test

**File:** `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/formula_test.go`

Add `TestDropAnnotationsQuantifierOrder` that thoroughly enforces the Python behavior:
body is processed before bound variables, so the annotation moves from the quantifier
to the first body occurrence.

### Test cases:

1. **ForAll with polymorphic `<=` (the exact bug)**:
   Build `forall T:lclock. (T:lclock <= _T)` and verify String() produces
   `(forall T. T:lclock <= _T)` — annotation on body T, NOT on quantifier T.
   Uses `<=` as a polymorphic symbol returning Boolean, which causes `inferredSort && !Boolean = False`,
   keeping the annotation on the first body occurrence.

2. **Exists with polymorphic `<=`**: Same pattern but with Exists.
   Verify `(exists T. T:lclock <= _T)`.

3. **Lambda with polymorphic `<=`**: Same pattern but with Lambda.
   Verify `(lambda T. T:lclock <= _T)`.

4. **ForAll with non-polymorphic function (control case)**:
   Build `forall X:S. f(X:S)` where f is a non-polymorphic function.
   In this case, body args get `inferredSort=true`, so annotation is stripped from body
   and stays on the quantifier variable. Verify: `(forall X:S. f(X))`.

5. **NamedBinder (vars-before-body order preserved)**:
   Build a NamedBinder with a polymorphic body. Since NamedBinder processes vars first
   (matching Python), annotation should stay on the quantifier variable.

### Construction pattern (for case 1):

```go
lclock := &UninterpretedSort{Name: "lclock"}
T, _ := NewVariable("T", lclock)
_T, _ := NewVariable("_T", lclock)
// <= is polymorphic, returns Boolean
leSort := mustFuncSort(t, lclock, lclock, Boolean)
le := NewConst("<=", leSort)
// T <= _T  (Apply of <= to T, _T; aSort = Boolean)
body, _ := NewApply(le, T, _T)
fa, _ := NewForAll([]*Variable{T}, body)
got := fa.String()
want := "(forall T. T:lclock <= _T)"
```

The polymorphic `<=` is the key: `isPolymorphicSymbolName("<=")` returns true,
so `dropAnnotations` on the Apply passes `inferredSort && !SortEqual(aSort, Boolean)`
= `False` to arg0, which means the first occurrence of T in the body keeps its annotation.

## Verification

1. Run `go build ./...` to check compilation
2. Run existing tests: `go test ./logic/...`
3. Run the new test: `go test ./logic/ -run TestDropAnnotationsQuantifierOrder -v`
4. Re-run the comparison that produced `ivy.log` vs `goivy.log.full` and verify the `forall T.` / `forall T:lclock.` divergence is gone
