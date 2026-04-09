# Plan: Fix LabeledFormula divergence at i=236914 in TestOrdLive

Created: 2026-04-09 09:50

## Context

The golden conformance test `TestOrdLive` (running on `ivy-lang-examples/doc/examples/apple/ord_live.ivy`) fails with a Go-vs-Python XTRACE divergence at index `i=236914`. This is the *next* divergence after the previous fix landed (commit `bc8f93e6 atg. still golden 236914`).

The divergence is shown in `~/ivy/goivy/log.red`:

```
236913  go : XTRACE: ast.LF.__init__ id=2033 counter=2034
        py : XTRACE: ast.LF.__init__ id=2033 counter=2034

236914  go : XTRACE: ast.LF.__init__ id=2034 counter=2035
        py : XTRACE: actions.GetUpdate ENTER type=EnvAction
```

Up through `i=236913` Go and Python both create the *same* 29-LF batch (`id=2005..2033`). At `i=236914`, **Go creates one extra `LabeledFormula` (id=2034) before transitioning to `actions.GetUpdate ENTER type=EnvAction`**, while Python transitions immediately. Go is creating exactly one extra `LabeledFormula` somewhere in the call chain that leads from "guarantee report loop" to "EnvAction.update entry".

The user-visible context (per Python `~/ivy/pyivy/ivy/ivy/ivy_check.py:700-715`):

```python
for sub in guarantees:
    print("            {}guarantee".format(pretty_lineno(sub)), end=' ')
    if check and any(...):
        print_dots()
        old_checked_assert = act.checked_assert.get()
        act.checked_assert.value = sub.lineno
        for root in checked_actions:
            if root in roots:
               tried.add((root,sub.lineno))
               action = act.env_action(root)
               ag = ivy_art.AnalysisGraph()
               pre = itp.State()
               pre.clauses = get_conjs(mod)
               with itp.EvalContext(check=False):
                   post = ag.execute(action,prestate=pre)   # ← path from here to GetUpdate ENTER
```

The Go equivalent is at `check/isolate_check.go:526-585`. Both languages match exactly through the LF batch (id=2005..2033 = 29 LFs); the deviation is exactly **one extra `LabeledFormula` constructed in Go** somewhere on the path

```
ag.Execute → PostState → ArtToInterpState → interp.ApplyAction → actions.GetUpdate
```

between when the matching LF batch ends and when `actions.GetUpdate ENTER type=EnvAction` would fire. The XTRACE format `ast.LF.__init__ id=%d counter=%d` is emitted only by `(*AstConfig).NewLabeledFormula` (`ast/decl_ast.go:53`) and `(*LabeledFormula).CloneWithFreshID` (`ast/decl_ast.go:123`); all other LF allocations either go through `Clone()` (which emits `clone PRESERVE`/`clone FRESH`) or call one of these two functions transitively.

### Why both source-of-truth files matter

CLAUDE.md rule A: Python is the source of truth. Rule B.6: every Python function must exist in Go with the same name. The fix must conform Go to Python — Go must NOT create the extra LF, because Python doesn't create one here.

## Investigation summary

What I confirmed (read-only) so far:

1. `art.PostState` (Go: `art/art.go:414-433`) calls `interp.ApplyAction`. Neither `PostState`, nor `ArtToInterpState`/`artToInterpMemo`, nor `provenanceToInterpExpr` calls any LF constructor in their bodies.
2. `interp.ApplyAction` (`interp/eval.go:147-186`) constructs a `UpdateContext`, fires its own xtrace `interp.ApplyAction calling GetUpdate ...`, and calls `actions.GetUpdate(action, ctx)`. No LF construction here either.
3. `actions.GetUpdate` (`actions/update.go:2001-2007`) immediately fires `actions.GetUpdate ENTER type=...` — this is Python's `Action.update` trace point (`ivy_actions.py:228-230`).
4. `BuildEnvAction` (`actions/action.go:2006-2047`) builds `Sequence(body, ReturnAction())` per public action and wraps in `EnvAction`. No LF construction.
5. `GetConjs` (`check/check.go:383-392`) collects formulas; no LF construction.
6. `art.NewState`, `ag.Add`, `art.NewAnalysisGraph` — none construct LFs.
7. `LabeledFormula.NewLabeledFormulaFrom` (`ast/decl_ast.go:60-71`) DOES create a fresh LF and emit the trace, but is only called from `check/check.go:850` inside `ConvertPostcondsWithUpdate`, which runs *after* `GetUpdate` finishes (during postcond renaming) — NOT in this path.
8. `AlwaysCloneWithFreshID` is only set true during module instantiation in `parser/inst_mod.go:157`; it is reset to false at line 160, so by the time `check_isolate` runs the flag is false. (Worth re-verifying since the divergence is at id=2034 — see Step 3 below.)
9. The 29-LF batch immediately preceding the divergence is emitted with **no other xtraces interleaved**, which means it is a tight loop somewhere that *only* allocates LFs (no atom/app/clone/sort traces between iterations). This is the fingerprint of either `NewLabeledFormulaFrom` or `CloneWithFreshID` running in a `for` loop.

The key insight: Go is one LF *ahead* of Python at exactly the moment Python first traces `actions.GetUpdate ENTER` for an `EnvAction`. The divergence is **between the end of the per-iteration LF batch and the entry to `GetUpdate`**, on a path that has *no* other xtraces, so a single targeted diagnostic will pin it precisely.

## Plan

We will not attempt to guess the LF site by reading more code. Instead we add **permanent matched-pair xtraces** along the call path — one trace on the Go side, one matching trace on the Python source-of-truth side, for each probe — so the next test run pinpoints the extra allocation, then apply the minimal fix that brings Go into line with Python. The matched traces are kept forever: today's debugging session accretes permanent regression coverage for this code path. **We only accrete xtrace coverage; we never delete it.**

### Step 1 — Add surgical diagnostic xtraces (PERMANENT — matched pairs on both sides)

**Policy: these traces are not temporary.** Per the user's direction we only *accrete* xtrace coverage; we never delete it. The debugging we do once today becomes a permanent conformance check that prevents future regressions in this area, at essentially zero runtime cost (xtraces compile out under the xtracer_off build tag on the Go side and under `python3 -O` on the Python side, matching CLAUDE.md's existing xtracer contract).

**Critical constraint**: every xtrace we add must be a *matched pair*. Each Go `xtracer.Trace(...)` must have a Python `xtracer.trace(...)` at the equivalent code location emitting the exact same string. If we only added traces on the Go side the two streams would diverge immediately at the first new trace. CLAUDE.md section F.5 documents this contract for canonical s-exps; it applies equally to xtrace lines.

Go source and the Python source of truth must be edited in lock-step. Python paths are under `/Users/jaten/ivy/pyivy/ivy/ivy/`.

Add these MATCHED PAIRS of traces:

1. **`check/isolate_check.go`** — just before `BuildEnvAction(...)` (around current line 552):
   `xtracer.Trace("check.guarantee_loop pre BuildEnvAction root=%s", root)`
   **`pyivy/ivy/ivy_check.py`** — inside the `for root in checked_actions:` loop (line ~708-709), just before `action = act.env_action(root)`:
   `if __debug__: xtracer.trace("check.guarantee_loop pre BuildEnvAction root=%s" % root)`

2. **`check/isolate_check.go`** — just after `BuildEnvAction(...)`:
   `xtracer.Trace("check.guarantee_loop post BuildEnvAction")`
   **`ivy_check.py`** — just after `action = act.env_action(root)`:
   `if __debug__: xtracer.trace("check.guarantee_loop post BuildEnvAction")`

3. **`check/isolate_check.go`** — after `art.NewAnalysisGraph(mod)`:
   `xtracer.Trace("check.guarantee_loop post NewAnalysisGraph")`
   **`ivy_check.py`** — after `ag = ivy_art.AnalysisGraph()`:
   `if __debug__: xtracer.trace("check.guarantee_loop post NewAnalysisGraph")`

4. **`check/isolate_check.go`** — after `art.NewState(mod, GetConjs(mod))`:
   `xtracer.Trace("check.guarantee_loop post NewState+GetConjs")`
   **`ivy_check.py`** — after `pre = itp.State(); pre.clauses = get_conjs(mod)`:
   `if __debug__: xtracer.trace("check.guarantee_loop post NewState+GetConjs")`

5. **`check/isolate_check.go`** — after `ag.Add(pre, nil)`:
   `xtracer.Trace("check.guarantee_loop post ag.Add")`
   **`ivy_check.py`** — NOTE: Python does NOT call `ag.add` on the pre-state at this point; it relies on `ag.execute` to add the post-state later. To keep the matched pair, emit the Python-side trace at the *same logical location* (just after the pre-state setup) so both sides emit this trace once per iteration with no observable semantic effect:
   `if __debug__: xtracer.trace("check.guarantee_loop post ag.Add")`
   *(The Python trace is a pure sentinel at this point; the Go call is the real `ag.Add`. The trace itself imposes no new semantics.)*

6. **`check/isolate_check.go`** — just before `ag.Execute(...)`:
   `xtracer.Trace("check.guarantee_loop pre ag.Execute")`
   **`ivy_check.py`** — just before `post = ag.execute(action, prestate=pre)`:
   `if __debug__: xtracer.trace("check.guarantee_loop pre ag.Execute")`

7. **`art/art.go`** — first line of `Execute`:
   `xtracer.Trace("art.Execute ENTER label=%s", label)`
   **`pyivy/ivy/ivy_art.py`** — first line of `def execute(self, op, prestate=None, abstractor=None, label=None):`:
   `if __debug__: xtracer.trace("art.Execute ENTER label=%s" % (label if label is not None else ''))`

8. **`art/art.go`** — first line of `PostState`, BEFORE `ArtToInterpState`:
   `xtracer.Trace("art.PostState ENTER opName=%s", op.Name())`
   **`ivy_art.py`** — first line of `def post_state(self, op, pre_state, abstractor):`, BEFORE the existing `art.PostState calling GetUpdate` trace. IMPORTANT: the existing Python trace `art.PostState calling GetUpdate type=%s` must be preserved; we add this new trace immediately before it:
   `if __debug__: xtracer.trace("art.PostState ENTER opName=%s" % (op.name() if hasattr(op, 'name') else type(op).__name__))`

9. **`art/art.go`** — after `ArtToInterpState(preState)` returns:
   `xtracer.Trace("art.PostState post ArtToInterpState")`
   **`ivy_art.py`** — Python does not have an `ArtToInterpState` conversion step (the states are already the right type). Emit the matched Python sentinel at the same logical point (just after the existing `art.PostState calling GetUpdate` trace, before `op.update(...)`) so the streams stay aligned:
   `if __debug__: xtracer.trace("art.PostState post ArtToInterpState")`

10. **`interp/eval.go`** — very first line of `ApplyAction`:
    `xtracer.Trace("interp.ApplyAction ENTER actionName=%s", actionName)`
    **`ivy_art.py`** — Python doesn't have an `ApplyAction` wrapper; the `post_state` method directly calls `op.update(...)`. Add the matched Python sentinel right before `s = concrete_post(op.update(...), pre_state)`:
    `if __debug__: xtracer.trace("interp.ApplyAction ENTER actionName=%s" % (op.name() if hasattr(op, 'name') else type(op).__name__))`

11. **`interp/eval.go`** — between the `UpdateContext` build and the existing `interp.ApplyAction calling GetUpdate` trace:
    `xtracer.Trace("interp.ApplyAction post UpdateContext build")`
    **`ivy_art.py`** — matched sentinel at the same logical point in `post_state` (right after probe 10):
    `if __debug__: xtracer.trace("interp.ApplyAction post UpdateContext build")`

All Python traces use `if __debug__:` so they compile out under `python3 -O`, mirroring the `xtracer_off` build tag on the Go side.

After these probes are in place, run the failing test once and look at the trailing traces in `log.red` around the new divergence index (it will be larger than 236914 since both sides now emit additional matched xtraces before the LF batch). The probe that *Python* fires but Go doesn't, or that Go fires but Python doesn't, is the one that brackets the offending LF allocation. Use the probe sub-regions to pin the file/function to fix:

- Divergence lands between probes (1) and (2) → `BuildEnvAction` is the offender. Compare to Python `env_action` in `ivy_actions.py:1797-1816`.
- Between (2) and (3) → `NewAnalysisGraph` is the offender. Compare to Python `ivy_art.AnalysisGraph.__init__`.
- Between (3) and (4) → `NewState` or (more likely) `GetConjs`/`module.NewClauses(...)` is the offender. Compare to Python `get_conjs` (`ivy_check.py:453-455`) and `lut.Clauses(fmlas, annot=...)`.
- Between (4) and (5) → `ag.Add` is the offender (Go-only call; the Python side doesn't call `ag.add` here at all — this alone may be the root cause).
- Between (5) and (6) → stray LF allocation in test plumbing between `Add` and `Execute`.
- Between (6) and (7) → `Execute`'s argument handling is allocating.
- Between (7) and (8) → `PostState`'s entry path is allocating.
- Between (8) and (9) → `ArtToInterpState`/`artToInterpMemo`/`provenanceToInterpExpr` is allocating an LF.
- Between (9) and (10) → state hand-off into `interp.ApplyAction`.
- Between (10) and (11) → `UpdateContext` build, the `GetAction` closure, or `state.Domain.Cfg.ActCfg` access is allocating.
- Between (11) and the existing `interp.ApplyAction calling GetUpdate` xtrace → `ActionTypeName(action)` formatting or the `GetUpdate` call setup is allocating.

### Step 2 — Run the test and read the diff

```
cd /Users/jaten/ivy/goivy && go test -tags xtracer -run TestOrdLive ./parser/ 2>&1 | tee log.red
```

The probes are silent up to `i=236913`; the next bracketing probe identifies the location. (If multiple probes fire before `i=236914`, the divergence point will simply move forward by the number of new probes — that's expected; what matters is which probe is *immediately* before `LF.__init__ id=2034` in the new run.)

### Step 3 — Compare Python and Go side-by-side and fix

Once the offending function is identified, do a strict mechanical-port diff against the Python source of truth:

- If the Python version uses `clone()` and Go uses `NewLabeledFormulaFrom`, switch Go to `(*LabeledFormula).Clone(args)` so that the `lf_counter` is decremented and the original ID is restored (Python's `LabeledFormula.clone` in `ivy_ast.py:649-666` decrements `lf_counter` and restores `res.id = self.id` when `always_clone_with_fresh_id` is False — that is the contract).
- If the Python version doesn't construct an LF at all (e.g. it just rebinds a reference), remove the construction from Go entirely.
- If `AlwaysCloneWithFreshID` has accidentally remained `true` past instantiation, find the missing `cfg.SetAlwaysCloneWithFreshID(false)` and add it. Verify by inspection of `parser/inst_mod.go` lines around 154-160 and any other call sites.
- If Go's path adds a "wrapping" step (e.g. wrapping the EnvAction body in an extra labeled wrapper) that Python doesn't have, delete that wrapping step. Per CLAUDE.md rule B.7 we do not add Go-only abstractions.

The fix MUST be applied in the Go file that holds the offending allocation. Do NOT change the LF tracing or the `xtracer.Trace` format — those are the contract that lets Go and Python be compared.

### Step 4 — Keep the diagnostic xtraces (they are permanent regression coverage)

**Do NOT delete the probes added in Step 1.** Per policy they are retained forever: this region of code was hard to debug today, and the matched traces permanently pin the shape of the `guarantee → env_action → execute → post_state → apply_action → get_update` path. Future divergences in the same area will now be caught at the nearest probe instead of drifting undetected. Cost is zero when `xtracer_off`/`python3 -O` is in effect.

The only cleanup in this step is to verify the matched pairs really are matched: for every trace added to Go in Step 1, confirm the Python side has the identical string at the logically-equivalent code point. Mismatched pairs will cause a fresh divergence.

### Step 5 — Re-run the conformance test to confirm

```
cd /Users/jaten/ivy/goivy && go test -tags xtracer -run TestOrdLive ./parser/ 2>&1 | tail -200
```

Expected outcome: `TestOrdLive` either passes outright or advances to a new (later) divergence index strictly greater than 236914. If it passes, also run the broader conformance suite to make sure no neighbouring test regressed:

```
cd /Users/jaten/ivy/goivy && go test -tags xtracer -run 'TestOrd' ./parser/ 2>&1 | tail -50
```

## Critical files

- `/Users/jaten/ivy/goivy/log.red` — current divergence log (read-only reference)
- `/Users/jaten/ivy/goivy/check/isolate_check.go` — Go guarantee loop (lines 462-588)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_check.py` — source of truth (lines 681-720)
- `/Users/jaten/ivy/goivy/art/art.go` — `Execute` (line 372), `PostState` (line 414), `ArtToInterpState` (line 1594), `NewAnalysisGraph` (line 277), `NewState` (line 48), `Add` (line 307)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_art.py` — Python `AnalysisGraph.execute` (line 188), `post_state` (line 159)
- `/Users/jaten/ivy/goivy/interp/eval.go` — `ApplyAction` (line 147)
- `/Users/jaten/ivy/goivy/actions/action.go` — `BuildEnvAction` (line 2006)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_actions.py` — `env_action` (line 1797), `Action.update` (line 228), `EnvAction.int_update` (line 910)
- `/Users/jaten/ivy/goivy/actions/update.go` — `GetUpdate` (line 2001), `EnvAction.IntUpdateEnv` (~line 1335)
- `/Users/jaten/ivy/goivy/check/check.go` — `GetConjs` (line 383), `ConvertPostcondsWithUpdate` (line 813, NOT in the failing path but useful as a reference for the right vs wrong way to recreate an LF)
- `/Users/jaten/ivy/goivy/ast/decl_ast.go` — `NewLabeledFormula` (line 45), `NewLabeledFormulaFrom` (line 60), `Clone` (line 74), `cloneInternal` (line 97), `CloneWithFreshID` (line 115)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_ast.py` — `LabeledFormula.__init__`/`clone`/`clone_with_fresh_id` (lines 614-679)
- `/Users/jaten/ivy/goivy/ast/config.go` — `LfCounter`, `NextLFID`, `AlwaysCloneWithFreshID`, `SetAlwaysCloneWithFreshID` (lines 25-87)
- `/Users/jaten/ivy/goivy/parser/inst_mod.go` — only place `AlwaysCloneWithFreshID` is toggled (line 153-160); must verify it is reliably reset to false
- `/Users/jaten/ivy/goivy/parser/golden_test.go` — the test harness, line-matching loop and divergence display (lines 510-680)

## Verification

End-to-end:

1. After Step 4 (probes removed) and the fix in place, run:
   ```
   cd /Users/jaten/ivy/goivy && go test -tags xtracer -run TestOrdLive ./parser/ 2>&1 | tee log.red
   ```
2. Confirm the test no longer fails at `i=236914`. Either the test passes outright, or it advances to a strictly larger index (in which case there is *another* downstream divergence to chase in a follow-up plan, but THIS divergence is fixed).
3. Run the focused related tests as a regression check:
   ```
   cd /Users/jaten/ivy/goivy && go test -tags xtracer -run 'TestOrd' ./parser/
   cd /Users/jaten/ivy/goivy && go test -tags xtracer -run 'TestApple' ./parser/
   ```
4. Run unit tests for the touched packages (likely `check`, `art`, `interp`, `actions`, or `ast` depending on which one held the offending allocation):
   ```
   cd /Users/jaten/ivy/goivy && go test ./check/... ./art/... ./interp/... ./actions/... ./ast/...
   ```
5. **Sanity-check that every Step-1 probe exists on BOTH sides** (the probes are permanent, so we *want* matches — one from Go, one from Python, for each probe string):
   ```
   for s in 'check.guarantee_loop pre BuildEnvAction' \
            'check.guarantee_loop post BuildEnvAction' \
            'check.guarantee_loop post NewAnalysisGraph' \
            'check.guarantee_loop post NewState+GetConjs' \
            'check.guarantee_loop post ag.Add' \
            'check.guarantee_loop pre ag.Execute' \
            'art.Execute ENTER label=' \
            'art.PostState ENTER opName=' \
            'art.PostState post ArtToInterpState' \
            'interp.ApplyAction ENTER actionName=' \
            'interp.ApplyAction post UpdateContext build'; do
       echo "== $s =="
       grep -rn -F "$s" /Users/jaten/ivy/goivy
       grep -rn -F "$s" /Users/jaten/ivy/pyivy
   done
   ```
   Expected: each probe string has exactly one Go match (in the file it was added to) AND one Python match (in `ivy_check.py` or `ivy_art.py`).
