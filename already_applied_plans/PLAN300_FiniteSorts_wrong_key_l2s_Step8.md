# Plan: Fix `FiniteSorts` lookup using wrong sort key in l2s Step 8

Created: 2026-04-14 ~22:15 UTC

## Context

**Divergence**: `make golden` diverges at xtrace line 736175 (in `TestOrdLive`, up from 734726 after the PrefixAction fix).

Both Go and Python enter `replaceNamedBindersAst` with an ActionTerm for binding `ext:cfabric.step` (binding[7]). The ActionTerm's `stmt` field differs at ENTER time — the first child of the outermost Sequence differs:

- **Go**: `assignAction(l2s_d(ph_type -> bool)(fml:ph), true)` — an l2s "done" variable assignment
- **Python**: `assumeAction(Implies(NamedBinder(l2s_g, ...), ...))` — an l2s "guarantee" assume

This means Go's `addParamsToD` list in Step 8 (PatchExports) is non-empty, while Python's `add_params_to_d` is empty for this binding.

**Root cause**: In `l2s_shared.go:703`, the `FiniteSorts` lookup uses `p.CSort.String()`:
```go
if p.CSort != nil && !cfg.FiniteSorts[p.CSort.String()] {
```

But `EnumeratedSort.String()` returns the extension list format `"{nop_ph,wr_ph,rd_ph,cpl_ph}"` (from `logic/sort.go:106-108`), while `cfg.FiniteSorts` is keyed by sort **names** like `"ph_type"` (populated from `m.Sig.Sorts` in `l2s.go:463`).

Result: `cfg.FiniteSorts["{nop_ph,wr_ph,rd_ph,cpl_ph}"]` is false (key not found), so Go incorrectly considers `ph_type` non-finite and adds `assignAction(l2s_d(ph_type)(fml:ph), true)` to `addParamsToD`.

Python uses `p.sort.name` (ivy_l2s.py:1357) which correctly returns the sort name `"ph_type"`, matching the key in `finite_sorts`. So Python correctly excludes it.

**Same bug also at line 301**: `cfg.FiniteSorts[v.VSort.String()]` in the reset_w logic has the identical issue.

## Fix

### 1. Export `sortName` as `SortName` in `logic/pretty.go`

The function already exists at `logic/pretty.go:396-410` and correctly handles all sort types:
```go
func sortName(s Sort) string {
    switch st := s.(type) {
    case *UninterpretedSort: return st.Name
    case *BooleanSort:       return "bool"
    case *EnumeratedSort:    return st.Name
    case *TopSort:           return st.Name
    default:                 return s.String()
    }
}
```

Rename to `SortName` (exported). Update the two internal callers at lines 100 and 111 in the same file.

### 2. Fix `l2s_shared.go:703` — Step 8 PatchExports

Change:
```go
if p.CSort != nil && !cfg.FiniteSorts[p.CSort.String()] {
```
To:
```go
if p.CSort != nil && !cfg.FiniteSorts[lg.SortName(p.CSort)] {
```

### 3. Fix `l2s_shared.go:301` — reset_w logic

Change:
```go
if !cfg.FiniteSorts[v.VSort.String()] {
```
To:
```go
if !cfg.FiniteSorts[lg.SortName(v.VSort)] {
```

## Critical files to modify

1. `logic/pretty.go` — export `sortName` → `SortName` (rename + update 2 internal callers)
2. `check/l2s_shared.go` — fix FiniteSorts lookups at lines 301 and 703

## Verification

1. `go build ./...` compiles clean
2. `make golden` — divergence advances past line 736175
