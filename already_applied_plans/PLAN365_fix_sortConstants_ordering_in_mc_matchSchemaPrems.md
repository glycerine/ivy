# Fix sortConstants ordering divergence in mc.matchSchemaPrems

**Created:** 2026-05-02 17:45 UTC

## Context

The golden test `Test2hrOrdLive` fails at XTRACE line 28624662. Go reports `nCands=2` where Python reports `nCands=1` for the same `matchSchemaPrems` non-function candidate lookup with `sortKey=t inMap=True isBound=True`.

The root cause is in `mc/transforms.go` `matchSchemaPrems`, the "allSorts" code path (lines 490-497). Go alphabetically sorts the `sortConstants` map keys, while Python iterates `sort_constants.values()` in dict insertion order. This produces different candidate ordering, so the first successful unification maps `t` to different concrete sorts, yielding different candidate counts on the recursive call.

Both sides have 2 sort buckets with 3 total constants (distribution {2,1}). Go's alphabetical order tries the sort with 2 constants first; Python's insertion order tries the sort with 1 constant first.

## Fix

Change `sortConstants` from `map[string][]*lg.Const` to `*iu.InsMap[string, []*lg.Const]` throughout the mc package. InsMap preserves insertion order (matching Python's dict), so the allSorts iteration matches Python exactly.

## File Changes

### 1. `mc/mine.go`

- Add import `iu "github.com/glycerine/ivy/goivy/ivyutils"`
- **MineConstants** (line 15): return type `map[string][]*lg.Const` -> `*iu.InsMap[string, []*lg.Const]`
  - `make(map[string][]*lg.Const)` -> `iu.NewInsMap[string, []*lg.Const]()`
  - `res[sortKey] = append(res[sortKey], sym)` -> `res.Set(sortKey, append(res.Get(sortKey), sym))`
- **MineConstants2** (line 58): same pattern as MineConstants
- **PrevExpr** (line 118): parameter `sortConstants map[string][]*lg.Const` -> `*iu.InsMap[string, []*lg.Const]`
  - `sortConstants[sortKey]` -> `sortConstants.Get(sortKey)` (nil is safe for range)

### 2. `mc/transforms.go` (contains the actual bug fix)

- **ExpandSchemata** (line 347): parameter type change; `len(sortConstants)` -> `sortConstants.Len()`
- **matchSchemaPrems** (line 410): parameter type change
  - Line 487: `sortConstants[lookupKey]` -> `sortConstants.Get(lookupKey)`
  - **Lines 490-497 (THE FIX):** Replace alphabetical sort block:
    ```go
    // OLD:
    scKeys := make([]string, 0, len(sortConstants))
    for k := range sortConstants { scKeys = append(scKeys, k) }
    sort.Strings(scKeys)
    for _, k := range scKeys { cands = append(cands, sortConstants[k]...) }
    
    // NEW:
    for _, v := range sortConstants.All() { cands = append(cands, v...) }
    ```
- **InstantiateAxioms** (line 619): parameter type change

### 3. `mc/qelim.go`

- **Qelim struct** (lines 27-28): both `SortConstants` and `SortConstants2` fields -> `*iu.InsMap[string, []*lg.Const]`
- **NewQelim** (line 34): both parameters -> `*iu.InsMap[string, []*lg.Const]`
- **GetConsts** (line 54): parameter type; `sortConstants[name]` -> `sortConstants.Get2(name)`
- **QE** (line 93): parameter type
- **qeQuantifier** (line 118): parameter type
- Apply (line 178): no code change needed, inferred types flow from struct fields

### 4. `mc/propabs.go`

- **PropAbs.SortConstants** (line 44): field type -> `*iu.InsMap[string, []*lg.Const]`
- **NewPropAbs** (line 50): parameter type

### 5. `mc/toaiger.go`

- Lines 292-293: no change needed (type inferred from MineConstants/MineConstants2 return)
- Lines 295, 303, 324: no change needed (parameter types now match)
- **Line 576**: `for _, consts := range sortConstants` -> `for _, consts := range sortConstants.All()`

## NOT changed

- `mc/schema.go` / `mc/mc_test.go` - use `map[string][]string` (separate simplified test type)
- `autoinst/schemata.go` - same class of bug but different package/code path; follow-up

## Verification

Run `cd ~/ivy/goivy && make test` - the Test2hrOrdLive test should now show Go matching Python's `nCands=1` at XTRACE line 28624662.
