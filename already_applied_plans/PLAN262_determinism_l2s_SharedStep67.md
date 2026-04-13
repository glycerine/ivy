# Fix Non-Deterministic Iteration Ordering in l2s SharedStep6 and SharedStep7

Created: 2026-04-13 ~19:30 UTC

## Context

`TestOrdLive` compares Go and Python xtraces line-by-line. At trace line 596105, they diverge in `l2s.SharedStep6 toG[0]`:

- **Go toG[0]**: `(Not body:(Eq t1:(Apply func:(Symbol name:ref.evs.l_req ...` — an Eq node
- **Python toG[0]**: `(Not body:(NamedBinder name:l2s_g ...` — a NamedBinder node

Both sides have the **same set** of elements in `toG`, but in **different order**. The root cause is that both sides iterate unordered collections (Go `map`, Python `set`) without applying a matching deterministic sort. The same class of bug exists in multiple places in SharedStep6 and SharedStep7.

## All Non-Deterministic Iteration Sites Found

| # | Location | Go collection | Python collection | Affects |
|---|----------|--------------|-------------------|---------|
| A | Step6: `toG` build | `map[string]L2sGTriple` sorted by `fmt.Sprint(Body)` | `set()` unsorted | `AssumeGAxioms` order + xtrace `toG[i]` |
| B | Step6: `assume_when_axioms` | `map[string]*NamedBinder` unsorted | `set()` unsorted | `AssumeWhenAxioms` order |
| C | Step7: `symprops` build | iterates `cfg.L2sGs` map | iterates `l2s_gs` set | `symprops[sym]` list ordering |
| D | Step7: `symwhens` build | iterates `cfg.L2sWhensSet` map | iterates `l2s_whens` set | `symwhens[sym]` list ordering |
| E | Step7: `propEventsFunc` | iterates `map[string]*NamedBinder` | iterates `set()` | `pre`/`post` action list order |
| F | Step7: `whenEventsFunc` | iterates `map[string]*NamedBinder` | iterates `set()` | `pre`/`post` action list order |
| G | Step7: `waitEventsFunc` | iterates `map[string]*NamedBinder` | iterates `set()` | `res` action list order |

Sites C and D feed into E/F/G (via `symprops`→`eventProps`→`propEventsFunc`), but since E/F/G receive maps/sets (which lose ordering anyway), fixing E/F/G is sufficient for action ordering. However, C and D should also be fixed for belt-and-suspenders determinism and to ensure `symprops[sym]` list order matches between Go and Python.

## Fix Strategy

Both sides must sort by the **canonical s-expression** (`Canon()`/`canon()`) since this is the cross-language standard already used in xtrace comparison. This replaces Go's current `fmt.Sprint(Body)` sort key (which uses Go-specific formatting) and Python's current lack of sorting.

---

## Detailed Changes

### Fix A — Step6 `toG` ordering (IMMEDIATE DIVERGENCE)

**Go** — `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s_shared.go` lines 272-274

Change:
```go
sort.Slice(toG, func(i, j int) bool {
    return fmt.Sprint(toG[i].Body) < fmt.Sprint(toG[j].Body)
})
```
To:
```go
sort.Slice(toG, func(i, j int) bool {
    return string(toG[i].Body.Canon()) < string(toG[j].Body.Canon())
})
```

**Python** — `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_l2s.py` after line 1030

After `to_g = list(dict.fromkeys(to_g))`, add:
```python
to_g.sort(key=lambda x: x[1].canon())
```

### Fix B — Step6 `L2sWhensSet` / `l2s_whens` ordering

**Go** — `l2s_shared.go` lines 294-306

Replace:
```go
cfg.AssumeWhenAxioms = nil
for _, when := range cfg.L2sWhensSet {
```
With:
```go
cfg.AssumeWhenAxioms = nil
sortedWhens := sortNamedBinderMap(cfg.L2sWhensSet)
for _, when := range sortedWhens {
```

**Add helper function** in `l2s_shared.go` (near top, after imports):
```go
// sortNamedBinderMap extracts values from a map[string]*lg.NamedBinder and
// returns them sorted by Canon() for deterministic cross-language ordering.
func sortNamedBinderMap(m map[string]*lg.NamedBinder) []*lg.NamedBinder {
    sorted := make([]*lg.NamedBinder, 0, len(m))
    for _, v := range m {
        sorted = append(sorted, v)
    }
    sort.Slice(sorted, func(i, j int) bool {
        return string(sorted[i].Canon()) < string(sorted[j].Canon())
    })
    return sorted
}
```

**Python** — `ivy_l2s.py` lines 1045-1048

Replace:
```python
assume_when_axioms = [
    AssumeAction(forall(when.variables, lg.Implies(when.body.t1,lg.Eq(when(*when.variables),when.body.t2))))
    for when in l2s_whens
]
```
With:
```python
sorted_whens = sorted(l2s_whens, key=lambda w: w.canon())
assume_when_axioms = [
    AssumeAction(forall(when.variables, lg.Implies(when.body.t1,lg.Eq(when(*when.variables),when.body.t2))))
    for when in sorted_whens
]
```

### Fix C — Step7 `symprops` build loop (deterministic list ordering)

**Go** — `l2s_shared.go` lines 336-341

Replace:
```go
for _, triple := range cfg.L2sGs {
    prop := l2sG(triple.Vars, triple.Body, triple.Environ)
    for _, sym := range il.SymbolsAst(triple.Body) {
        symprops[lg.Key(sym)] = append(symprops[lg.Key(sym)], prop)
    }
}
```
With:
```go
sortedTriples := sortL2sGTriples(cfg.L2sGs)
for _, triple := range sortedTriples {
    prop := l2sG(triple.Vars, triple.Body, triple.Environ)
    for _, sym := range il.SymbolsAst(triple.Body) {
        symprops[lg.Key(sym)] = append(symprops[lg.Key(sym)], prop)
    }
}
```

**Add helper function** in `l2s_shared.go`:
```go
// sortL2sGTriples extracts values from a map[string]L2sGTriple and
// returns them sorted by Body.Canon() for deterministic cross-language ordering.
func sortL2sGTriples(m map[string]L2sGTriple) []L2sGTriple {
    sorted := make([]L2sGTriple, 0, len(m))
    for _, v := range m {
        sorted = append(sorted, v)
    }
    sort.Slice(sorted, func(i, j int) bool {
        return string(sorted[i].Body.Canon()) < string(sorted[j].Body.Canon())
    })
    return sorted
}
```

**Python** — `ivy_l2s.py` lines 1173-1177

Replace:
```python
for vs, t, env in l2s_gs:
    prop = l2s_g(vs,t,env)
    envprops[env].append(prop)
    for sym in ilu.symbols_ilu_ast(t):
        symprops[sym].append(prop)
```
With:
```python
for vs, t, env in sorted(l2s_gs, key=lambda x: x[1].canon()):
    prop = l2s_g(vs,t,env)
    envprops[env].append(prop)
    for sym in ilu.symbols_ilu_ast(t):
        symprops[sym].append(prop)
```

### Fix D — Step7 `symwhens` build loop (deterministic list ordering)

**Go** — `l2s_shared.go` lines 342-345

Replace:
```go
for _, when := range cfg.L2sWhensSet {
    for _, sym := range il.SymbolsAst(when.Body) {
        symwhens[lg.Key(sym)] = append(symwhens[lg.Key(sym)], when)
    }
}
```
With:
```go
sortedWhens7 := sortNamedBinderMap(cfg.L2sWhensSet)
for _, when := range sortedWhens7 {
    for _, sym := range il.SymbolsAst(when.Body) {
        symwhens[lg.Key(sym)] = append(symwhens[lg.Key(sym)], when)
    }
}
```

**Python** — `ivy_l2s.py` lines 1178-1180

Replace:
```python
for when in l2s_whens:
    for sym in ilu.symbols_ilu_ast(when.body):
        symwhens[sym].append(when)
```
With:
```python
for when in sorted(l2s_whens, key=lambda w: w.canon()):
    for sym in ilu.symbols_ilu_ast(when.body):
        symwhens[sym].append(when)
```

### Fix E — Step7 `propEventsFunc` (sort map/set at entry)

**Go** — `l2s_shared.go` lines 356-394

Change the function to sort at entry. Replace:
```go
propEventsFunc := func(gprops map[string]*lg.NamedBinder) ([]actions.Action, []actions.Action) {
    var pre, post []actions.Action
    for _, gprop := range gprops {
```
With:
```go
propEventsFunc := func(gprops map[string]*lg.NamedBinder) ([]actions.Action, []actions.Action) {
    var pre, post []actions.Action
    sortedProps := sortNamedBinderMap(gprops)
    for _, gprop := range sortedProps {
```
And change the second loop (line 370):
```go
    for _, gprop := range gprops {
```
To:
```go
    for _, gprop := range sortedProps {
```

**Python** — `ivy_l2s.py` lines 1093-1108

Add sort at function entry. After `def prop_events(gprops):`, before line 1094:
```python
    gprops = sorted(gprops, key=lambda p: p.canon())
```
(This replaces the set argument with a sorted list; the rest of the function iterates `gprops` unchanged.)

### Fix F — Step7 `whenEventsFunc` (sort map/set at entry)

**Go** — `l2s_shared.go` lines 399-435

Replace:
```go
whenEventsFunc := func(whens map[string]*lg.NamedBinder) ([]actions.Action, []actions.Action) {
    var pre, post []actions.Action
    for _, when := range whens {
```
With:
```go
whenEventsFunc := func(whens map[string]*lg.NamedBinder) ([]actions.Action, []actions.Action) {
    var pre, post []actions.Action
    sortedWhens := sortNamedBinderMap(whens)
    for _, when := range sortedWhens {
```
And change the second loop (line 423):
```go
    for _, when := range whens {
```
To:
```go
    for _, when := range sortedWhens {
```

**Python** — `ivy_l2s.py` lines 1110-1131

Add sort at function entry. After `def when_events(whens):`, before line 1111:
```python
    whens = sorted(whens, key=lambda w: w.canon())
```

### Fix G — Step7 `waitEventsFunc` (sort map/set at entry)

**Go** — `l2s_shared.go` lines 438-453

Replace:
```go
waitEventsFunc := func(waits map[string]*lg.NamedBinder) []actions.Action {
    var res []actions.Action
    for _, wait := range waits {
```
With:
```go
waitEventsFunc := func(waits map[string]*lg.NamedBinder) []actions.Action {
    var res []actions.Action
    sortedWaits := sortNamedBinderMap(waits)
    for _, wait := range sortedWaits {
```

**Python** — `ivy_l2s.py` lines 1138-1155

Add sort at function entry. After `def wait_events(waits):`, before line 1139:
```python
    waits = sorted(waits, key=lambda w: w.canon())
```

---

## Files to Modify

| File | Fixes |
|------|-------|
| `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s_shared.go` | A, B, C, D, E, F, G + two helper functions |
| `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_l2s.py` | A, B, C, D, E, F, G |

## Verification

1. Run `TestOrdLive` — it should pass the divergence point at trace 596105 and either pass fully or reveal a later divergence from a different source
2. If the test passes, also run the broader golden test suite to check for regressions
3. The xtrace comparison is the definitive check — if Go and Python produce identical traces, the ordering is correct
