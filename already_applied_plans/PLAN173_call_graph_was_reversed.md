# Fix ActionCallGraph Direction: Reverse Graph to Match Python

**Created**: 2026-04-01 20:45

## Context

Golden test diverges at line 156791:

```
156791  go : XTRACE: check.CreateIsolate before_bracket_actions n_brackets=17
        py : XTRACE: check.CreateIsolate before_bracket_actions n_brackets=11
```

Go produces 17 bracket actions while Python produces 11. Bracket actions wrap exported actions with assume(conjecture) statements for invariant checking.

## Root Cause

**Go's `ActionCallGraph` builds the graph in the OPPOSITE direction from Python's `call_graph()`.**

### Python (`ivy_module.py:282-287`) — REVERSE graph: callee → [callers]
```python
def call_graph(self):
    callgraph = defaultdict(list)
    for actname, action in self.actions.items():
        for called_name in action.iter_calls():
            callgraph[called_name].append(actname)  # callee → caller
    return callgraph
```

### Go (`isolate/deps.go:874-893`) — FORWARD graph: caller → [callees]
```go
func ActionCallGraph(mod *module.Module) map[string][]string {
    graph := make(map[string][]string)
    for name, act := range mod.Actions.All() {
        // ...collects callees...
        graph[name] = calls  // caller → callees (WRONG DIRECTION)
    }
    return graph
}
```

### How this causes the bracket count divergence

`GetIsolateExports` / `get_isolate_exports` uses the call graph to determine which actions are exports:

**Python** (`ivy_isolate.py:2204-2209`): checks `cg[act]` = callers of `act`. If any caller is outside the isolate → act is exported (called from outside = boundary crossing inward).

**Go** (`isolate/iter.go:224-247`): checks `callGraph[act]` = callees of `act`. If any callee is outside the isolate → act is exported (calls outside = boundary crossing outward).

These are **opposite semantics**. Python correctly identifies actions called FROM outside. Go incorrectly identifies actions that CALL outside. This produces different export sets → different bracket counts.

## Changes

### 1. Fix `ActionCallGraph` to build reverse graph (callee → callers)

**File**: `~/ivy/goivy/isolate/deps.go` (lines 874-893)

Change from forward graph to reverse graph matching Python:

```go
func ActionCallGraph(mod *module.Module) map[string][]string {
    graph := make(map[string][]string)
    for actname, act := range mod.Actions.All() {
        for _, sub := range act.IterSubactions() {
            if ca, ok := sub.(*actions.CallAction); ok {
                calledName := CanonAct(ca.CalleeName())
                graph[calledName] = append(graph[calledName], actname)
            }
        }
    }
    // Sort each caller list for determinism
    for k := range graph {
        sortStrings(graph[k])
    }
    return graph
}
```

### 2. Update comment in `GetIsolateExports`

**File**: `~/ivy/goivy/isolate/iter.go` (line 236)

Change comment from "Check if any callee is outside the isolate" to "Check if any caller is outside the isolate" and rename variable:

```go
// Check if any caller of this action is outside the isolate
if callers, ok := callGraph[act]; ok {
    for _, caller := range callers {
        if !isoActions[caller] {
            exports[act] = true
            break
        }
    }
}
```

### 3. Update `TestActionCallGraph` test

**File**: `~/ivy/goivy/isolate/isolate_test.go` (around line 490)

Update test expectations to reflect reverse graph semantics (callee → callers instead of caller → callees).

### Files to modify
- `~/ivy/goivy/isolate/deps.go` — lines 874-893 (reverse graph direction)
- `~/ivy/goivy/isolate/iter.go` — lines 236-243 (update comments/variable names)
- `~/ivy/goivy/isolate/isolate_test.go` — around line 490 (update test expectations)

## Verification

```bash
cd ~/ivy/goivy && go test ./isolate/ && go build ./... && make golden
```

Check that trace divergence advances past 156791.
