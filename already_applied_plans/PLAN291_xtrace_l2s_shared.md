# Add Comprehensive xtracer Tracing to l2s_shared.go and ivy_l2s.py

Created: 2026-04-13 ~19:00 UTC

## Context

At xtrace line 596035, Go and Python diverge in `replaceTemporalsRec` — the ASTs passed in are already structurally different. Go has `Not(Globally(Not(And(Apply(cf_live.issued_memc(...)), ...))))` while Python has `Not(Globally(Not(Symbol(cfabric.rd_fair))))`. This means `cfg.ToWait` is constructed differently between Go and Python. The existing tracing in `l2s_shared.go` is minimal (no xtracer calls at all), making it impossible to pinpoint where the divergence originates. We need comprehensive tracing in both Go and Python to identify the root cause.

## Files to Modify

1. **Go**: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s_shared.go`
2. **Python**: `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_l2s.py`

## Trace Points to Add

All traces use format: `xtracer.Trace("l2s.<function> <description> <key>=<value> HASH canon=<canon>")` in Go and `xtracer.trace(...)` in Python. Every Go trace must have a matching Python trace with identical format so they can be diffed.

### 1. SharedStep1_ConvertTemporals (Go:76-115 / Python:811-827)

Already traced: ENTER/EXIT with counts (in Python, Go currently has none in l2s_shared.go but l2s.go emits them).

**Add in Go** (after line 109, after `cfg.NotLf` computation):
```
l2s.SharedStep1 notLf HASH canon=<cfg.NotLf.Canon()>
```

**Add in Python** (after line 815):
```
l2s.SharedStep1 notLf HASH canon=<not_lf.canon()>
```

### 2. SharedStep3_CollectNamedBinders (Go:123-202 / Python:848-895)

**Add after collecting from sources** (Go ~line 143, Python ~line 865):
```
l2s.SharedStep3 namedBindersConjs key=<k> nEntries=<len(v)>
```
For each key in named_binders_conjs, trace key and count.

**Add for each to_wait entry** (Go ~line 201, Python ~line 892):
```
l2s.SharedStep3 toWait[<i>] nVars=<len(vb.Vars)> HASH canon=<vb.Body.Canon()>
```

**Add for each to_save entry** (Go ~line 202, Python ~line 894):
```
l2s.SharedStep3 toSave[<i>] nVars=<len(vb.Vars)> HASH canon=<vb.Body.Canon()>
```

### 3. SharedBuildSaveAndWait (Go:207-238 / Python:905-927)

**Add for each save_state entry** (Go ~line 212, Python ~line 908):
```
l2s.SharedBuildSaveAndWait saveState[<i>] nVars=<len(vb.Vars)> HASH canon=<vb.Body.Canon()>
```

**Add for each done_waiting entry** (Go ~line 219, Python ~line 912):
```
l2s.SharedBuildSaveAndWait doneWaiting[<i>] nVars=<len(vb.Vars)> HASH canon=<inner.Canon()>
```

**Add for each reset_w iteration** (Go ~line 224, Python ~line 921):
```
l2s.SharedBuildSaveAndWait resetW[<i>] nVars=<len(vb.Vars)> body HASH canon=<vb.Body.Canon()>
```

**Add before ReplaceTemporals call** (Go ~line 233, Python ~line 919):
```
l2s.SharedBuildSaveAndWait resetW[<i>] preReplace HASH canon=<inputFormula.Canon()>
```

**Add after ReplaceTemporals call** (Go ~line 234, Python ~line 919):
```
l2s.SharedBuildSaveAndWait resetW[<i>] postReplace HASH canon=<negGlob.Canon()>
```

**Add EXIT** (Go ~line 238, Python line 927):
```
l2s.SharedBuildSaveAndWait EXIT nSaveState=<len> nDoneWaiting=<len> nResetW=<len>
```

### 4. SharedStep6_BuildTableau (Go:242-297 / Python:1025-1049)

**Add for each to_g entry** (Go ~line 253, Python ~line 1028):
```
l2s.SharedStep6 toG[<i>] nVars=<len(vars)> HASH canon=<body.Canon()>
```

**Add counts at EXIT** (after building all axiom lists):
```
l2s.SharedStep6 EXIT nAssumeG=<len> nAssumeWhen=<len> nAssumeInit=<len> nAssumeW=<len>
```

### 5. SharedStep7_InstrumentActions (Go:301-512 / Python:1172-1241)

Already traced ENTER/EXIT in Python. Add matching in Go.

### 6. SharedStep8_PatchExports (Go:517-562 / Python:1245-1280)

Already traced ENTER/EXIT in Python. Add matching in Go.

### 7. SharedStep11_ReplaceNamedBinders (Go:564-599 / Python:1314-1344)

Already traced ENTER/EXIT in Python. **Add in both**:
```
l2s.SharedStep11 namedBinders key=<k> count=<len(binders)>
```

For each subs entry:
```
l2s.SharedStep11 sub freshName=<freshName> binderKey=<binder.String()>
```

## Implementation Details

### Go: l2s_shared.go

Add import:
```go
"github.com/glycerine/ivy/goivy/xtracer"
```

Use `xtracer.Trace(...)` with `%s` for Canon() values (Canon() returns `iu.Canonical` which is a string type).

For logic expressions: `expr.Canon()` → canonical s-expression string.
For actions: `act.Canon()` → canonical s-expression string.
For VarBodyPair: trace `vb.Body.Canon()` for body and `len(vb.Vars)` for var count.

### Python: ivy_l2s.py

Use `if __debug__: xtracer.trace(...)` pattern (matching existing traces).

For logic nodes: `node.canon()` → canonical s-expression string.
For VarBodyPair-like tuples `(vs, t)`: trace `t.canon()` for body and `len(vs)` for var count.

Note: Python's list comprehensions (e.g. `save_state = [... for vs, t in to_save]`, `reset_w = [... for vs, t in to_wait]`) need to be converted to explicit loops to add per-iteration traces.

## Verification

After adding traces, run the golden test:
```
cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy/parser && go test -run TestOrdLive -v -count=1 -timeout 600s
```

The divergence should now show up at a more specific trace point within SharedStep3 or SharedBuildSaveAndWait, pinpointing exactly which `to_wait` entry or `reset_w` input differs.
