# Plan: Add xtracer.Trace instrumentation to mc/ and proof/ (Go) + ivy_mc.py (Python)

**Created:** 2026-04-30 18:30 UTC

## Context

The `make rfn` golden test compares Go and Python xtrace output line-by-line. The mc/ package (toaiger.go, checker.go) has only 1 trace (`mc.ActionToTR calling GetUpdate`). Python's ivy_mc.py also has only 2 traces. Both sides need matching xtracer.Trace() calls at key pipeline stages so that divergences are caught early — especially in the ToAiger pipeline which has ~7 major transformation steps with no visibility today.

The proof/ package is already well-instrumented (~150+ traces); no changes needed there. The check/ package also has good coverage. The gap is specifically in the mc (model-checking) pipeline.

## Trace message format convention

All new traces use the `"HASH canon="` style so the golden test's `DiffSexp()` can produce structured diffs:

```
xtracer.Trace("mc.ToAiger step4d EXIT nAxioms=%d HASH canon= trans=%s", len(axs), trans.Canon())
```

Python equivalent:
```python
if __debug__: xtracer.trace("mc.ToAiger step4d EXIT nAxioms=%d HASH canon= trans=%s" % (len(axs), ilu.clauses_canon(trans)))
```

## Traces to add

### A. mc/toaiger.go (Go) + ivy_mc.py (Python)

Each trace must appear in both files with matching messages. Line numbers are approximate.

#### A1. After skolemization of invariant (Go ~line 118, Py ~line 1157)
```
mc.ToAiger postSkolemize HASH canon= invariant=<invariant.Sexp()>
```

#### A2. After GetUpdate + AddPostAxioms (Go ~line 134, Py ~line 1176)
```
mc.ToAiger postAddPostAxioms nStVars=<N> nTRfmlas=<N> nTRdefs=<N> HASH canon= trans=<trans.Canon()>
```

#### A3. After Qelim (Go ~line 294, Py ~line 1242)
```
mc.ToAiger postQelim nStVars=<N> nTRfmlas=<N> nTRdefs=<N> HASH canon= trans=<trans.Canon()>
```

#### A4. After axiom instantiation (Go ~line 307, Py ~line 1261)
```
mc.ToAiger postAxiomInst nAxioms=<N> nTRfmlas=<N> nTRdefs=<N> HASH canon= trans=<trans.Canon()>
```

#### A5. After propositional abstraction (Go ~line 375, Py ~line 1343)
```
mc.ToAiger postPropAbs nStVars=<N> nNewStVars=<N> nTRfmlas=<N> nTRdefs=<N> HASH canon= trans=<trans.Canon()>
```

#### A6. After Step 5 rename + extra defs (Go ~line 433, Py ~line 1378)
```
mc.ToAiger postRename nStVars=<N> nTRfmlas=<N> nTRdefs=<N> HASH canon= trans=<trans.Canon()>
```

#### A7. After Step 6 — final trans before encoder (Go ~line 454, Py ~line 1387)
```
mc.ToAiger finalTrans nStVars=<N> nInputs=<N> nTRdefs=<N> HASH canon= trans=<trans.Canon()>
```

#### A8. After DefList + latch setting + miter — AIGER result (Go ~line 550, Py ~line 1418)
```
mc.ToAiger aigerDone nInputs=<N> nLatches=<N> nOutputs=<N> nGates=<N>
```
(Gate count from `aiger.Sub.NumGates()` / `aiger.sub.num_gates()` — no canon needed, just metrics.)

#### A9. Invariant HASH (Go ~line 378, Py ~line 1348)
```
mc.ToAiger invariant HASH canon= <invariant.Sexp()>
```

### B. mc/checker.go (Go) + ivy_mc.py check_isolate (Python)

#### B1. CheckIsolate ENTER (Go ~line 132, Py ~line 1687)
```
mc.CheckIsolate ENTER method=<method>
```

#### B2. CheckIsolate after ToAiger (Go ~line 144, Py ~line 1708)
```
mc.CheckIsolate postToAiger aigerLen=<len(aigerStr)>
```

#### B3. CheckIsolate EXIT (Go ~line 158-168, Py ~line 1767)
```
mc.CheckIsolate EXIT proved=<bool> err=<err>
```

### C. check/check.go MCTactic (Go) + ivy_check.py mc_tactic (Python)

#### C1. MCTactic ENTER (Go ~line 1073, Py ~line 859)
```
check.MCTactic ENTER nGoals=<N>
```

#### C2. MCTactic after temporal tactic chain (Go ~line 1077, Py ~line 868)
```
check.MCTactic postTacticChain nGoals=<N>
```

#### C3. MCTactic EXIT (Go ~line 1093, Py ~line 871)
```
check.MCTactic EXIT nRemainingGoals=<N> err=<err>
```

## Python-side Canon helpers

Python doesn't have `Clauses.canon()` built-in. Need to use `ilu.clauses_canon(trans)` or add a matching helper. Check if `canon_ast.py` already provides this, or add:

```python
def clauses_canon(clauses):
    fmlas = ' '.join(f.sexp() if hasattr(f,'sexp') else str(f) for f in clauses.fmlas)
    defs = ' '.join(d.sexp() if hasattr(d,'sexp') else str(d) for d in clauses.defs)
    return '(clauses fmlas:[%s] defs:[%s])' % (fmlas, defs)
```

This must match Go's `module.Clauses.Canon()` output format exactly.

## Files to modify

| File | Side | Changes |
|------|------|---------|
| `mc/toaiger.go` | Go | Add traces A1-A9 (~9 traces) |
| `mc/checker.go` | Go | Add traces B1-B3 (~3 traces) |
| `check/check.go` | Go | Add traces C1-C3 (~3 traces) |
| `ivy_mc.py` | Py | Add matching traces A1-A9, B1-B3 |
| `ivy_check.py` | Py | Add matching traces C1-C3 |
| `ivy_logic_utils.py` or `canon_ast.py` | Py | Add `clauses_canon()` if not present |

## Verification

1. `cd ~/ivy/goivy && make test` — all existing tests pass
2. `cd ~/ivy/goivy && make rfn` — run golden comparison. New traces should appear on both sides, matching line-for-line. The test should get at least as far as before (line 225115) or further.
