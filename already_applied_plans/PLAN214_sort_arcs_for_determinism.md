# PLAN213: Fix arc ordering divergence due to set/map iteration order

**Created:** 2026-04-07 ~05:30 UTC

## Context

The golden test (`make golden`) diverges at raw line 226078, `arc[684]`:

```
Go:  from_id=1764(root=1309) to_id=1772(root=1772)
Py:  from_id=51(root=1746)   to_id=1772(root=1772)
```

Both `to_id` and `fmla` match. Only `from_id` (and its root) differ.

**Root cause:** The `from` nodes of arcs come from iterating over `uvs` — a `map[*uf.UFNode]bool` in Go and a `set()` in Python. When `uvs` contains multiple elements, **Go iterates by pointer hash (memory address)** while **Python iterates by `__hash__` = `self.id` (integer hash)**. These orderings differ, causing arcs from the same formula to appear at different indices in the arc list.

Arcs 0–683 all matched because every `uvs` set encountered up to that point had ≤1 element (single-element sets have no ordering issue). At arc[684], the outer equality `rfn.abs.queue(M_d0, Y_a) = (ref.evs.req(T_a2) = read)` produces a `uvs` with 2+ elements, exposing the iteration-order difference.

Both nodes (1764 and 51) exist on both sides — they are simply iterated in different order, placing the arcs at different list positions.

## Approach: Sort `uvs` nodes by ID before creating arcs

Sort the `uvs` set/map elements by UFNode ID before iterating, in both the equality and application arc-creation loops. This ensures deterministic arc ordering regardless of Go vs Python hash function differences.

Arc order doesn't affect the cycle detection algorithm — it operates on the graph structure, not list position.

## File changes

### 1. Go: `~/ivy/goivy/fragment/fragment.go` — sort uvs before arc creation

**Add helper** (near top of file, after arc struct):

```go
// sortedUFNodes returns the nodes in a map sorted by ID for deterministic ordering.
func sortedUFNodes(m map[*uf.UFNode]bool) []*uf.UFNode {
    nodes := make([]*uf.UFNode, 0, len(m))
    for n := range m {
        nodes = append(nodes, n)
    }
    sort.Slice(nodes, func(i, j int) bool {
        return nodes[i].ID < nodes[j].ID
    })
    return nodes
}
```

**Line 258–260** — Equality case:

```go
// Before:
for v := range reses[i].uvs {
    c.arcs = append(c.arcs, arc{from: v, to: sSigma, fmla: fmla, lineno: lineno, argIdx: -1})
}

// After:
for _, v := range sortedUFNodes(reses[i].uvs) {
    c.arcs = append(c.arcs, arc{from: v, to: sSigma, fmla: fmla, lineno: lineno, argIdx: -1})
}
```

**Line 311–312** — Application case:

```go
// Before:
for v := range reses[i].uvs {
    c.arcs = append(c.arcs, arc{from: v, to: anode, fmla: fmla, lineno: lineno, argIdx: i, hasIdx: true})
}

// After:
for _, v := range sortedUFNodes(reses[i].uvs) {
    c.arcs = append(c.arcs, arc{from: v, to: anode, fmla: fmla, lineno: lineno, argIdx: i, hasIdx: true})
}
```

### 2. Python: `~/ivy/pyivy/ivy/ivy/ivy_fragment.py` — sort uvs before arc creation

**Line 120** — Equality case:

```python
# Before:
arcs.extend((v,S_sigma,fmla,lineno) for v in uv)

# After:
arcs.extend((v,S_sigma,fmla,lineno) for v in sorted(uv, key=lambda n: n.id))
```

**Line 148** — Application case:

```python
# Before:
arcs.extend((v,anode,fmla,lineno,idx) for v in uvs[idx])

# After:
arcs.extend((v,anode,fmla,lineno,idx) for v in sorted(uvs[idx], key=lambda n: n.id))
```

## Verification

```bash
# Golden test (should advance past line 226078)
cd ~/ivy/goivy && make golden
```

## Notes

- The cycle detection algorithm (`iu.cycle` / `c.findCycle`) operates on graph topology, not arc list order, so sorting arcs by creation order has no effect on correctness.
- Future divergences of the same kind (set iteration order) are prevented by this fix.
- The `all_uvs` return value from `mapFmla` is itself a set/map, but it's consumed by the caller's arc creation loop — which is also fixed by these changes.
