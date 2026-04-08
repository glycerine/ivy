# Fix: Enum constant cache check must precede solver_name call in translateVarOrConst

**Created:** 2026-04-08 ~18:15 UTC

## Context

The golden test (`make golden`) shows a trace divergence at line 236165:
```
236165  go : XTRACE: ivy_solver.py:65 solver_name() ENTER name=nop
        py : XTRACE: ivy_solver.py:279 functionsort() ENTER
```

Go emits a spurious `solver_name()` trace for the enum constant `nop` (of type `op_type`), while Python has already moved on to translating the next FunctionSort in the formula. The root cause is an ordering mismatch between Go and Python.

## Root Cause Analysis

### Python flow (ivy_solver.py:479–493, term_to_z3 for constants):

1. **Cache check first** (line 479): `res = z3_constants.get(str(term.rep))` — for enum constants like `nop`, this finds the entry pre-registered by `enumeratedsort()` (line 293–294: `z3_constants[str(c)] = c`)
2. If cache hit → return immediately, no `solver_name()` or `to_z3()` called
3. Only if cache miss → call `iso.to_z3()` and `solver_name(term.rep)`

### Go flow (translate.go:844–899, translateVarOrConst):

1. **solver_name first** (line 853): `z3name := t.z3Name(name, sort)` — always calls `SolverName`, emitting trace
2. **Cache check second** (line 855–857): checks `t.consts[key]` — finds enum constant registered by `TranslateSort`'s EnumeratedSort handler (line 295–297)
3. If cache hit → returns, but solver_name trace was already emitted

### Why only enum constants are affected:

Python's `enumeratedsort()` stores enum constants in `z3_constants` with **string keys** (`str(c)` = `"nop"`). The term_to_z3 lookup uses `str(term.rep)` = `"nop"` — string-to-string match works. For non-enum constants, Python stores with `z3_constants[term.rep]` (Symbol object key) but looks up with `str(term.rep)` (string key) — type mismatch means the cache never hits, so solver_name is always called for non-enum constants. Go must match this behavior.

## Plan

### Single change in `z3bridge/translate.go`, function `translateVarOrConst` (line 844)

Add an early cache check **only for EnumeratedSort constants** before the `z3Name` call at line 853. This matches Python's effective behavior where enum constants (pre-registered by `enumeratedsort()`) are found in `z3_constants` before `solver_name` is ever called.

**Before:**
```go
func (t *Translator) translateVarOrConst(name string, sort lg.Sort) (Expr, error) {
    if isNumeralName(name) {
        if result, err := t.Numeral(name, sort); result != nil {
            return *result, err
        }
    }
    z3name := t.z3Name(name, sort)    // <-- always calls solver_name
    if lg.FirstOrderSort(sort) {
        key := lg.NodeKey(name + ":" + string(sort.Sexp()))
        if cached, ok := t.consts[key]; ok {
            return cached, nil          // <-- too late, solver_name already called
        }
        ...
```

**After:**
```go
func (t *Translator) translateVarOrConst(name string, sort lg.Sort) (Expr, error) {
    if isNumeralName(name) {
        if result, err := t.Numeral(name, sort); result != nil {
            return *result, err
        }
    }

    // Python term_to_z3 line 479: z3_constants.get(str(term.rep)).
    // For enum constants pre-registered by enumeratedsort(), this cache
    // hit returns immediately without calling solver_name or to_z3.
    // Only enum constants get this early check because Python's z3_constants
    // cache only effectively works for enum constants (string-to-string key
    // match). Non-enum constants use Symbol object keys for storage but
    // string keys for lookup, so the cache never hits for them — solver_name
    // is always called. We must match that behavior.
    if _, isEnum := sort.(*lg.EnumeratedSort); isEnum {
        key := lg.NodeKey(name + ":" + string(sort.Sexp()))
        if cached, ok := t.consts[key]; ok {
            return cached, nil
        }
    }

    z3name := t.z3Name(name, sort)
    if lg.FirstOrderSort(sort) {
        ...  // rest unchanged
```

### File to modify
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/z3bridge/translate.go` — lines 844–866

No other files need changes.

## Verification

Run `cd ~/ivy/goivy && make golden` and confirm the divergence at 236165 is resolved (either the next divergence is at a later line, or the test passes).
