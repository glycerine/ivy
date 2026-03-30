# Plan: Investigate and Fix Golden Test Divergence at Line 147682

**Created**: 2026-03-30 03:15

## Context

The previous plan (fix Args()/Clone() on 5 LF-bearing actions, eliminate NodeArgs/NodeClone) is COMPLETE. The golden test advanced from 147239 to 147682. Now we need to fix the next divergence:

```
147682  go : XTRACE: LocalAction.__init__ uniqueID=719 caller=ast.LocalAction.clone
        py : XTRACE: CallAction.__init__ uniqueID=381 counter=382
```

Go's action tree has a `LocalAction` node where Python has a `CallAction` node. During a recursive traversal (almost certainly `PrefixCallsFunc` in `actions/transforms.go`), Go clones a LocalAction while Python creates a new CallAction. The preceding 12 traces (147670-147681) match perfectly, so the tree structures are identical except at this one position.

## Analysis

### What we know

1. **The traversal is PrefixCallsFunc**: LocalAction goes through the default case (clone via `action.Clone(newArgs)` -> `NewLocalActionOn("ast.LocalAction.clone")` -> trace). CallAction goes through the special `case *CallAction:` (creates new via `NewCallActionOn` -> trace). The trace pattern matches this exactly.

2. **The tree structure differs at exactly one node**: The tree being traversed by PrefixCallsFunc has a LocalAction in Go where Python has a CallAction.

3. **Where does this tree come from?** The traversal happens after `apply_mixin EXIT` (line 147669). In the isolation flow (`isolate/isolate.go:488-545`), actions are processed through:
   - `AssertToAssume` (preserves node types)
   - `PrefixCalls` (preserves node types except CallAction gets new callee)
   - `AddMixinsExt` (applies mixins via ApplyMixin, which uses SubstituteConstantsAction)

   None of these transforms change a CallAction into a LocalAction or vice versa. So the difference originates from the **compiled action tree** stored in `mod.Actions`.

4. **LocalAction wraps CallAction**: Go's `SplitReturns` (actions/action.go:719) wraps a CallAction with temp returns in a LocalAction. Python's `split_returns` (ivy_actions.py:1372) does the same. But `SplitReturns`/`split_returns` is only called from `ivy_ranking.py` and `ivy_l2s.py`, NOT from the isolation flow. It's also called from `action_update` paths.

5. **Missing `copy_formals` in Go transforms**: Go's `PrefixCallsFunc`, `assertToAssumeChildren`, `DropInvariants`, and `UnrollLoops` default cases do NOT call `CopyFormalsTo` after cloning. Python always calls `self.copy_formals(res)`. This is a metadata bug but unlikely to cause the type difference. It SHOULD be fixed anyway.

### What we don't know

- **Which specific node** in the compiled action tree differs (we know it's one node among many in a nested tree)
- **Which action definition** is being traversed (which `actname` in the isolation loop)
- **Whether the compiler or some isolation-specific helper** (like `hide_action_params`) created the extra LocalAction

## Plan

### Step 1: Add diagnostic tracing to identify the divergent node

Add ENTER tracing to `PrefixCallsFunc` (Go) and `prefix_calls` (Python) to show the node type being processed. This will reveal exactly where the tree types diverge.

**Go — `actions/transforms.go` PrefixCallsFunc:**
```go
func PrefixCallsFunc(action Action, renamer func(string) string) Action {
    if action == nil || renamer == nil {
        return action
    }
    xtracer.Trace("actions.prefix_calls ENTER type=%s", shortTypeName(action))
    switch a := action.(type) {
    ...
```

**Python — `ivy_actions.py` Action.prefix_calls (line 256):**
```python
def prefix_calls(self,pref):
    if __debug__: xtracer.trace("actions.prefix_calls ENTER type=%s" % type(self).__name__)
    args = [a.prefix_calls(pref) if isinstance(a,Action) else a for a in self.args]
    ...
```

**Python — `ivy_actions.py` CallAction.prefix_calls (line 1334):**
```python
def prefix_calls(self,pref):
    if __debug__: xtracer.trace("actions.prefix_calls ENTER type=%s" % type(self).__name__)
    res = CallAction(*([self.args[0].prefix(pref) ...
    ...
```

### Step 2: Run golden test to identify the divergence point

`cd ~/goivy && make golden` — the new ENTER traces will show exactly which node type differs at the point of divergence, giving us the tree path context.

### Step 3: Trace back to root cause

Based on Step 2 results, add tracing to the action creation site (compiler/action.go, compiler/compiler.go, isolate/helpers.go, etc.) to understand why Go creates a LocalAction where Python creates a CallAction.

### Step 4: Fix the root cause

Fix the compilation or isolation code to match Python's behavior.

### Step 5: Fix missing `copy_formals` in transform default cases

While investigating, also fix the missing `CopyFormalsTo` calls in the default cases of:
- `PrefixCallsFunc` (actions/transforms.go:372)
- `assertToAssumeChildren` (actions/transforms.go)
- `DropInvariants` (actions/transforms.go)
- `UnrollLoops` (actions/transforms.go)

Python's default implementations all call `self.copy_formals(res)` after cloning.

## Files to modify

1. **`actions/transforms.go`** — Add ENTER trace to PrefixCallsFunc; add CopyFormalsTo in default cases
2. **`~/pyivy/ivy/ivy/ivy_actions.py`** — Add matching ENTER traces to Action.prefix_calls and CallAction.prefix_calls
3. **TBD after Step 2** — The file containing the root cause of the LocalAction/CallAction tree difference

## Verification

`go build ./...` then `cd ~/goivy && make golden`
