# PLAN214: Fix cycle detection trace divergence and union-find Unify mismatch

**Created:** 2026-04-07 ~08:00 UTC

## Context

The golden test (`make golden`) now diverges at line 236013. At line 236012, the checker canon fully matched (arcs, stratMap, stratInfo, macroValueMap all agree). The divergence is:

```
Go:  XTRACE: fragment CheckFEU EXIT
Py:  XTRACE: fragment/checker.reportCycle ENTER
```

## Root Cause Analysis

### Issue 1: Trace structure mismatch (immediate cause)

**Python** (`ivy_fragment.py:517,540`) always calls `report_cycle()` at line 517, which unconditionally traces "reportCycle ENTER" (line 424) and "reportCycle EXIT" (line 429), then at line 540 traces "CheckFEU EXIT".

So Python's trace sequence is always:
```
fragment/checker.reportCycle ENTER
[fragment/checker.reportCycle report cycle error]  # only if cycle found
fragment/checker.reportCycle EXIT
fragment CheckFEU EXIT
```

**Go** (`fragment.go:838-845`) only calls `reportCycle` when a cycle IS found:
```go
cycle := c.findCycle()
if len(cycle) > 0 {
    return c.reportCycle(cycle)
}
return nil  // triggers deferred "CheckFEU EXIT"
```

So Go's no-cycle trace is just `fragment CheckFEU EXIT` — missing the reportCycle ENTER/EXIT pair.

### Issue 2: Union-find `Unify` self-merge behavior (latent bug)

**Python** (`ivy_union_find2.py:50-66`) does NOT check if x and y are already in the same set:
```python
def unify(x, y):
    x = find(x)
    y = find(y)
    # NO check: if x == y: return
    if x.rank < y.rank:
        x, y = y, x
    y.parent = x           # no-op when x == y
    if x.rank == y.rank:   # always true when x == y
        x.rank = x.rank + 1  # spurious rank increment!
```

**Go** (`unionfind.go:57-58`) has an early return:
```go
if x.ID == y.ID {
    return       // Python doesn't do this
}
```

When `unify(x,y)` is called and both are already in the same set, Python increments the root's rank but Go skips entirely. Over many calls, Python's ranks become inflated. Since rank determines merge direction (higher-rank node becomes root), this can change which node is the root after future merges, altering `find()` results and the effective graph topology for cycle detection.

Currently the arc root dumps match (the HASH canon= lines before line 236012 all agreed), so this hasn't caused a visible difference yet. But it's a correctness issue per the mechanical port rules — Go must match Python's behavior exactly.

## Approach

Fix both issues:

1. **Always call `reportCycle` in Go** regardless of whether a cycle was found, matching Python's unconditional call pattern.
2. **Remove the `x.ID == y.ID` early return in Go's `Unify`** to match Python's behavior.

## File Changes

### 1. Go: `fragment/fragment.go` — always call reportCycle

**Lines 838-845** — Change:

```go
// Before:
cycle := c.findCycle()
if len(cycle) > 0 {
    return c.reportCycle(cycle)
}
return nil

// After:
cycle := c.findCycle()
return c.reportCycle(cycle)
```

The existing `reportCycle` (lines 688-702) already handles the empty case:
```go
func (c *checker) reportCycle(cycle []arc) error {
    xtracer.Trace("fragment/checker.reportCycle ENTER")
    defer xtracer.Trace("fragment/checker.reportCycle EXIT")
    if len(cycle) == 0 {
        return nil
    }
    // ... report error ...
}
```

### 2. Go: `unionfind/unionfind.go` — remove self-merge optimization

**Lines 57-59** — Remove early return to match Python:

```go
// Before:
x = Find(x)
y = Find(y)
if x.ID == y.ID {
    return
}

// After:
x = Find(x)
y = Find(y)
```

After removal, when x == y (same root node):
- `y.parent = x` → no-op (root already points to itself)
- `x.rank == y.rank` → true (same node), so `x.rank++`
- This matches Python's behavior exactly.

## Verification

```bash
cd ~/ivy/goivy && make golden
```

Expected: The test advances past line 236013. If Python truly found no cycle, the next 3 trace lines should now match:
```
line 236013: fragment/checker.reportCycle ENTER
line 236014: fragment/checker.reportCycle EXIT
line 236015: fragment CheckFEU EXIT
```

If Python DID find a cycle (trace contains "report cycle error"), the union-find fix (Issue 2) should resolve the graph topology difference so Go also finds the cycle.

## Notes

- `reportCycle` already traces ENTER/EXIT and handles nil/empty gracefully — no changes needed to that function.
- The union-find rank inflation in Python is technically a bug (standard union-find should skip same-root merges), but per mechanical port rules, Go must match Python's behavior since Python is the source of truth.
- The `findCycle` algorithm itself is correct and equivalent between Go and Python — the issue was purely about when `reportCycle` is called and whether the underlying union-find trees match.
