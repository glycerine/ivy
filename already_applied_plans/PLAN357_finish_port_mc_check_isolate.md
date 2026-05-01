# Full Port of check_isolate() Trace Reconstruction

**Created**: 2026-05-01 ~UTC  
**Scope**: mc/checker.go CheckIsolate, mc/trace_decode.go AigerMatchHandler2 + AigerWitnessToIvyTrace2, mc/encoder.go decoder methods

## Context

Python's `check_isolate()` (ivy_mc.py:1716-1802) has a complete counterexample trace reconstruction pipeline: when the ABC model checker finds a counterexample, it calls `aiger_witness_to_ivy_trace2()` which simulates the AIGER circuit with witness inputs, walks the action/annotation tree via `match_annotation`, and builds a full `TraceBase`-based analysis graph. The Go port stopped short — `CheckIsolate` returns the raw `WitnessTrace` with a comment "For now, return the raw trace" and never calls `AigerWitnessToIvyTrace2`. The Go `AigerMatchHandler2` has wrong method signatures, doesn't implement `AnnotationHandler`, and its `NewState`/`FinalState` are oversimplified stubs.

## Files to Modify

| File | What |
|------|------|
| `mc/encoder.go` | Add DecodeVal, GetSym, GetNextSym, GetState |
| `mc/toaiger.go:~499` | Set Encoder.Interp field during construction |
| `mc/trace_decode.go` | Rewrite AigerMatchHandler2 + AigerWitnessToIvyTrace2 |
| `mc/checker.go` | Update CheckResult, wire CheckIsolate to decoded trace |
| `mc/trace_decode_test.go` (new) | Comprehensive tests |

## Step 1: Encoder Decode Methods — `mc/encoder.go`

**Python**: ivy_mc.py:490-530 (Encoder class methods)

### 1a. Add `Interp` field to Encoder struct

```go
type Encoder struct {
    // ... existing fields ...
    Interp  map[string]interface{} // sort interpretations for DecodeVal
}
```

### 1b. DecodeVal(bits []lg.Expr, v *lg.Const) lg.Expr

Python: `decode_val(bits, v)` (line 490-508). Converts multi-bit simulation values to an Ivy expression based on the sort's theory.

- Use `theory.GetSortTheory(v.CSort, e.Interp)` to determine sort type
- Enumerated sort: `num = BinDec(intBits)`, index into `sort.Extension`, clamp to `len-1`
- Range sort: `num = BinDec(intBits)`, clamp to upper bound
- BitVector: `num = BinDec(intBits)`, return as string symbol
- Boolean: return `bits[0]` directly (already an `lg.And`=true or `lg.Or`=false)

### 1c. GetSym(v *lg.Const) lg.Expr

Python: `get_sym(v)` (line 520-524). Reads current simulation value.

```
enc := e.Encoding[lg.Key(v)]          // []string of sub-bit names
abits := e.Sub.SymVals(enc)           // string of '0'/'1' chars
bits := bitsToBoolExprs(abits)        // convert to []lg.Expr
return e.DecodeVal(bits, v)
```

### 1d. GetNextSym(v *lg.Const) lg.Expr

Same as GetSym but uses `e.Sub.SymNextVals(enc)`.

### 1e. GetState(post string) map[lg.NodeKey]lg.Expr

Python: `get_state(post)` (line 511-518). Decodes all latch values from a post-state string.

```
subres := e.Sub.GetState(post)        // map[string]byte
for each latch v:
    enc := e.Encoding[lg.Key(v)]
    bits := [subres[s] for s in enc]  // look up each sub-bit
    val := e.DecodeVal(bits, v)
    res[lg.Key(v)] = val
```

Helper: `bitsToBoolExprs(s string) []lg.Expr` — converts '0'→`&lg.Or{}`, '1'→`&lg.And{}`.

## Step 2: Set Encoder.Interp in ToAiger — `mc/toaiger.go`

After `NewEncoder(inputs, stVars, outputs)` (~line 499), add:
```go
aiger.Interp = mod.Sig.Interp  // or wherever the interpretation map lives
```

Check: Python has `thy.get_sort_theory(v.sort)` which uses a global registry. In Go, the interp map is on the module. Verify `mod.Sig.Interp` exists and has the right type.

## Step 3: Rewrite AigerMatchHandler2 — `mc/trace_decode.go`

**Python**: ivy_mc.py:1527-1599 (class AigerMatchHandler2, extends TraceBase)

### 3a. New struct definition

Must implement `actions.AnnotationHandler` interface:
```go
type AnnotationHandler interface {
    Eval(cond lg.Expr) bool
    Handle(action Action, env map[lg.NodeKey]lg.Expr)
    DoReturn(action Action, env map[lg.NodeKey]lg.Expr)
    Fail()
}
```

The Go `trace.TraceBase.Handle` takes `(actions.Action, map[string]string)` which is the wrong signature for `AnnotationHandler`. Following the pattern from `check/helpers.go MatchHandler`, the MC handler should implement `AnnotationHandler` directly and replicate the needed TraceBase sub-trace delegation logic.

```go
type AigerMatchHandler2 struct {
    Aiger    *Encoder
    Decoder  map[string]lg.Expr    // symbol name → original expression
    Consts   map[string]bool
    StVarSet map[string]bool
    Current  map[string]lg.Expr
    Mod      *module.Module

    // Trace-building state (replicate TraceBase pattern)
    AG          *art.AnalysisGraph
    LastAction  actions.Action
    Sub         *AigerMatchHandler2
    Returned    *AigerMatchHandler2
    IsFullTrace bool
    States      [][]lg.Expr          // collected state equations per step
}
```

### 3b. Eval(cond lg.Expr) bool — Python lines 1533-1543

```
if isFalse(cond): return false
if isTrue(cond):  return true
if Not:           return !Eval(body)
default:          return isTrueNode(e.Aiger.GetSym(cond.(*lg.Const)))
```

### 3c. Handle(action actions.Action, env map[lg.NodeKey]lg.Expr)

Replicates Python TraceBase.handle (ivy_trace.py:201-211):
```
if sub != nil:
    sub.Handle(action, env)
elif lastAction is CallAction/EnvAction and returned is nil:
    sub = clone()
    sub.Handle(action, env)
else:
    if action lineno != "nowhere":
        NewState(env)
        lastAction = action
```

### 3d. NewState(env map[lg.NodeKey]lg.Expr) — Python lines 1554-1585

This is the most complex method. Must faithfully port:

1. Build `invEnv` (map[lg.NodeKey]lg.Expr): invert env, skipping skolems and is_new symbols
2. For each input `v` in decoder: `showSym(v, decoder[v.Name], aiger.GetSym(v), &eqns)`
3. Build `rn` map: `{x → new(x)}` for each `x` in stvarset (rename to `new_` prefix)
4. For each latch `v` in decoder:
   - `showSym(v, decoder[v.Name], aiger.GetSym(v), &eqns)`
   - Compute `nextDecd = RenameASTByName(decd, rn)` and `curDecd = RenameASTByName(decd, envNameMap)`
   - If `nextDecd == curDecd`: also `showSym(v, nextDecd, aiger.GetNextSym(v), &eqns)`
5. Call `AddState(eqns)`

**showSym(v, decd, val lg.Expr, eqns *[]lg.Expr)** — Python lines 1559-1568:
- If decd is Apply with `__new_` prefix: strip to `new_` prefix
- Check all used symbols in decd are either in invEnv, or (not skolem, not new, not in env)
- Rename decd with invEnv
- Skip if result is Apply and is_new(rep)
- Skip if constant and is constructor
- Append `Eq{expr, val}` to eqns

### 3e. FinalState() — Python lines 1587-1599

```
aiger.Sub.Advance()
post := aiger.Sub.LatchVals()
stmap := aiger.GetState(post)
for each latch v not named "__init":
    if val := stmap[lg.Key(v)]; val != nil:
        stvals = append(stvals, Eq{decoder[v.Name], val})
AddState(stvals)
```

### 3f. End() — Python TraceBase.end (ivy_trace.py:231-236)

```
if sub != nil:
    sub.End()
    returned = sub
    sub = nil
FinalState()
```

### 3g. Clone(), DoReturn(), Fail(), AddState()

- Clone: returns new handler with same aiger/decoder/consts/stvarset/current, fresh trace state
- DoReturn: no-op (Python line 1551-1552)
- Fail: wrap LastAction in FailAction
- AddState: create `module.Clauses` from equations, create `art.State`, add to AG

## Step 4: Rewrite AigerWitnessToIvyTrace2 — `mc/trace_decode.go`

**Python**: ivy_mc.py:1662-1701

### New signature

```go
func AigerWitnessToIvyTrace2(
    result *ToAigerResult,
    witnessFilename string,
    mod *module.Module,
) (*AigerMatchHandler2, error)
```

### Logic (matching Python)

1. Open witness file, read first line (must be "1")
2. Collect remaining non-empty lines
3. `result.Aiger.Sub.Reset()`
4. Create handler: `NewAigerMatchHandler2(result.Aiger, result.Decoder, result.Consts, result.StVarSet, mod)`
5. For each line (count from 1):
   a. Split into 4 columns (pre, inp, out, post)
   b. `result.Aiger.Sub.Step(inp)`
   c. If `count == len(lines)`: check `invar__fail` via `result.Aiger.GetSym()` — if true, break
   d. **`actions.MatchAnnotation(result.Action, result.Annot.(actions.Annotation), handler, mod)`** — THIS IS THE KEY MISSING CALL
   e. **`handler.End()`** — THIS IS THE OTHER KEY MISSING CALL
6. Return handler

**Note**: `result.Annot` is `interface{}`; type-assert to `actions.Annotation`.

## Step 5: Update CheckIsolate — `mc/checker.go`

### 5a. Update CheckResult

```go
type CheckResult struct {
    Proved       bool
    Trace        *WitnessTrace
    DecodedTrace *AigerMatchHandler2  // decoded Ivy trace (new)
    Error        error
}
```

### 5b. Refactor temp file management

Currently `RunABC` manages temp files internally and deletes them. But `AigerWitnessToIvyTrace2` needs the witness file path. Two options:

**Option A** (preferred — matches Python structure): Move temp file management into `CheckIsolate`. Have `RunABC` accept explicit file paths. This is how Python does it — `check_isolate` creates the temp files and passes `outfilename` directly.

**Option B** (simpler): Have `RunABC` return the witness file path in `CheckResult` and defer cleanup to caller.

Go with Option A for faithful porting.

### 5c. Wire up trace reconstruction

Replace lines 172-175 ("For now, return the raw trace"):

```go
decodedTrace, err := AigerWitnessToIvyTrace2(result, outName, mod)
if err != nil {
    return &CheckResult{Trace: checkResult.Trace, Error: fmt.Errorf("trace decode: %w", err)}, nil
}
return &CheckResult{Trace: checkResult.Trace, DecodedTrace: decodedTrace}, nil
```

### 5d. Add logfile support

Python opens `ivy_mc.log` as a global. Go should open `goivy_mc.log`. Per CLAUDE.md rules (no globals), add a `MCLogFile *os.File` or `MCLogFileName string` field to the mc Config or module.Config, and open/write in CheckIsolate. The logfile name is **`goivy_mc.log`** to avoid conflicting with the Python side during golden side-by-side testing.

## Step 6: Comprehensive Tests — `mc/trace_decode_test.go`

Tests that would have caught the missing functionality:

| Test | What it verifies |
|------|------------------|
| TestDecodeValBoolean | DecodeVal returns correct bool expression |
| TestDecodeValEnumerated | DecodeVal maps binary to correct enum symbol |
| TestDecodeValRange | DecodeVal clamps to range upper bound |
| TestGetSymReturnsDecodedValue | GetSym reads simulation state correctly |
| TestGetNextSymReturnsNextState | GetNextSym reads next-state correctly |
| TestGetStateDecodesAllLatches | GetState decodes full latch map |
| TestAigerMatchHandler2ImplementsAnnotationHandler | Compile-time interface check |
| TestAigerMatchHandler2EvalFalseTrue | Eval handles false/true/Not |
| TestAigerMatchHandler2HandleBuildsState | Handle triggers NewState, creates trace |
| TestAigerMatchHandler2EndCallsFinalState | End finalizes trace |
| TestAigerMatchHandler2Clone | Clone produces independent handler |
| TestAigerWitnessToIvyTrace2CallsMatchAnnotation | **Key test**: verify handler has trace states after witness processing (catches the current bug) |
| TestCheckIsolateReturnsDecodedTrace | Integration: ABC counterexample → decoded trace |

## Dependency Order

```
Step 1 (Encoder methods)        ← no deps
Step 2 (toaiger Interp field)   ← no deps
Step 3 (AigerMatchHandler2)     ← depends on Step 1
Step 4 (AigerWitnessToIvyTrace2) ← depends on Step 3
Step 5 (CheckIsolate)           ← depends on Step 4
Step 6 (Tests)                  ← alongside each step
```

## Verification

Run `cd ~/ivy/goivy && make test` after each step to verify no regressions. The golden/rfn tests exercise the mc path end-to-end.
