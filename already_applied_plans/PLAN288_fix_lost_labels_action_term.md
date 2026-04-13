# Fix Go ActionTerm Missing Labels

Created: 2026-04-13 ~15:30 UTC

## Context

At golden test trace line 330541, Go ActionTerm for `ext:lclock.next` has `labels:[]` while Python has `labels:["cf_live" "cf_live" "rfn" "rfn" "tprops" "tprops" ...]`. The labels are completely lost in Go.

The Python duplicate labels (each appearing twice) are **expected behavior** — `get_isolate_map(verified=True, present=True)` iterates both verified and present lists, and components in both get their isolate name appended twice. Go's `IterIsolate` follows the same pattern, so duplicates will naturally match once labels flow through.

## Root Cause

`ApplyMixin` in `actions/helpers.go:198-200` has **dead code** instead of copying labels:

```go
// GO BUG — dead code:
if ab, ok := action2.(interface{ SetLabels([]string) }); ok {
    _ = ab // labels handled by CopyFormalsTo   <-- NEVER CALLED
}
```

```python
# PYTHON — actually copies labels:
if hasattr(action2,'labels'):
    res.labels = action2.labels
```

**How labels are lost:**
1. `HandleTemporals` (compiler/ivy_compile.go:233) sets labels on actions in `mod.Actions`
2. `CreateIsolate` (compiler/ivy_compile.go:242) builds new actions via `AddMixinsExt` → `ApplyMixin`
3. `ApplyMixin` creates a fresh `*Sequence` via `ConcatActions` (Labels=nil), sets FormalParams/FormalReturns, but **never copies labels**
4. `mod.Actions` is replaced with the label-less new actions (isolate/isolate.go:1344)
5. `NormalProgramFromModule` reads empty labels

## Fixes (all in one file)

**File: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/helpers.go`**

### Fix 1: `ApplyMixin` (line 198-200) — primary fix

Replace the dead code with actual label copying, matching Python `ivy_actions.py:1513-1514`:

```go
// BEFORE (dead code):
if ab, ok := action2.(interface{ SetLabels([]string) }); ok {
    _ = ab // labels handled by CopyFormalsTo
}

// AFTER:
if labeler, ok := action2.(interface{ GetLabels() []string }); ok {
    if labels := labeler.GetLabels(); labels != nil {
        res.SetLabels(labels)
    }
}
```

### Fix 2: `AppendToAction` (line 211-222) — secondary fix

Add label copying from `action1`, matching Python `ivy_actions.py:1525-1526`. After the FormalReturns copy (line 221), add:

```go
if labeler, ok := action1.(interface{ GetLabels() []string }); ok {
    if labels := labeler.GetLabels(); labels != nil {
        res.SetLabels(labels)
    }
}
```

### Fix 3: Standalone `CopyFormalsTo` wrapper (line 228-235) — tertiary fix

The wrapper's comment says "copies formal parameters, returns, and labels" but doesn't copy labels. The `ActionBase.CopyFormalsTo` method at `module/action.go:67-78` does copy labels. Make the wrapper match. Add after line 234:

```go
if labeler, ok := src.(interface{ GetLabels() []string }); ok {
    if labels := labeler.GetLabels(); labels != nil {
        if setter, ok2 := dst.(interface{ SetLabels([]string) }); ok2 {
            setter.SetLabels(labels)
        }
    }
}
```

This also fixes `EmptyClone` (isolate/isolate.go:174-178), `PrefixAction`, and `PostfixAction` which all use this wrapper.

## Note: `SummarizeAction` is already correct

`isolate/isolate.go:133-149` has the same dead-code pattern at lines 138-140 but then has a SEPARATE working block at lines 141-149 that correctly copies labels. No change needed.

## Verification

Run the golden test that currently fails at trace line 330541:
```
cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test -run TestOrdLive -v ./parser/ -timeout 300s
```

The test should advance past line 330541 (labels now populated and matching Python's duplicated format).
