# Plan: Investigate EraseUnrefed Trace Order Divergence (CallAction vs LocalAction)

**Created**: 2026-03-30 06:50

## Context

After fixing EraseUnrefed to always clone (matching Python), the golden test diverges at line 150264:
```
150264  go : XTRACE: CallAction.__init__ uniqueID=606 counter=607
        py : XTRACE: LocalAction.__init__ uniqueID=1064 caller=ast.LocalAction.clone
```

The always-clone fix IS working — Go now emits action __init__ traces during EraseUnrefed (previously it exited CreateIsolate here). But the FIRST action type cloned differs: Go hits a CallAction first, Python hits a LocalAction first.

Both traces come from the default case of erase_unrefed cloning actions. The divergence means either:
1. The actions dict is iterated in different order
2. The action trees have different internal structure (from earlier processing)
3. Different AssignActions/HavocActions get erased (due to different `allSyms`), changing which non-leaf actions see "changed" children

## Investigation needed

The root cause is unclear — multiple possible causes. Diagnostic tracing will identify the exact source.

### Step 1: Add diagnostic trace to EraseUnrefed

**File**: `actions/transforms.go`, at the entry of EraseUnrefed

Add a trace showing what action type and name is being processed:

```go
func EraseUnrefed(action Action, syms map[string]bool, names map[string]bool) Action {
    if action == nil { return nil }
    xtracer.Trace("actions.erase_unrefed ENTER type=%s", actionTypeName(action))
    ...
```

And add corresponding trace in Python's `Action.erase_unrefed`:
```python
def erase_unrefed(self,refs,ref_names):
    if __debug__: xtracer.trace("actions.erase_unrefed ENTER type=%s" % type(self).__name__)
    ...
```

### Step 2: Add trace at the erase_unrefed call site showing action names

**File**: `isolate/isolate.go`, line 949-951

```go
for actname, act := range newActions.All() {
    xtracer.Trace("isolate.erase_unrefed_loop actname=%s type=%s", actname, actionTypeName(act))
    newActions.Set(actname, actions.EraseUnrefed(act, allSyms, allNames))
}
```

And in Python `ivy_isolate.py` around line 1217:
```python
def _erase_with_trace(actname, action, all_syms, all_names):
    if __debug__: xtracer.trace("isolate.erase_unrefed_loop actname=%s type=%s" % (actname, type(action).__name__))
    return action.erase_unrefed(all_syms, all_names)
new_actions = dict((actname, _erase_with_trace(actname, action, all_syms, all_names)) for actname,action in new_actions.items())
```

### Step 3: Run test and analyze divergence

Run `cd ~/goivy && make golden` and examine where the traces first diverge. This will tell us:
- If the action iteration order differs → fix InsMap ordering
- If the same action has different internal structure → find where the tree was built differently
- If different actions get erased → investigate `allSyms`/`allNames` computation

### Step 4: Fix based on findings

The fix depends on what Step 3 reveals.

## Files to modify (diagnostic phase)

1. **`actions/transforms.go`** — Add trace to EraseUnrefed entry
2. **`isolate/isolate.go`** — Add trace to erase_unrefed loop
3. **`~/pyivy/ivy/ivy/ivy_actions.py`** — Add trace to Action.erase_unrefed
4. **`~/pyivy/ivy/ivy/ivy_isolate.py`** — Add trace to erase_unrefed loop

## Verification

`go build ./...` then `cd ~/goivy && make golden` — examine the new traces to identify the root cause.
