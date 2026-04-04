# PLAN199: Codebase Duplication & Redundancy Audit

**Created:** 2026-04-04 ~23:30 UTC
**Status:** AUDIT ONLY (no fixes — catalog for future cleanup plans)

## Context

During the Python-to-Go port, code was duplicated across packages due to: (a) type
differences requiring parallel implementations (ast.Node vs lg.Expr), (b) private/public
copies within the same package, and (c) wrapper functions that delegate without adding
logic. This audit catalogs all found duplication for future cleanup.

**Exclusions:** `vprint.go` (13 copies) and `omap.go` (2 copies) are intentional and excluded.

---

## HIGH — Exact Duplicates Within/Across Packages

### D4: `GetCone()` / `getCone()` — 2 copies in isolate/

- `isolate/helpers.go:738-763` — private `getCone()`
- `isolate/phase7.go:225-250` — public `GetCone()`

Same recursive cone-of-influence algorithm with minor structural differences.

### D5: `GetPropsProvedInIsolateOrig()` — 2 copies in isolate/

- `isolate/helpers.go:662-681` — private `getPropsProvedInIsolateOrig()`
- `isolate/phase7.go:149-172` — public `GetPropsProvedInIsolateOrig()`

Identical bodies.

### D6: `makeAnd()` / `makeAndH()` — 2 copies in isolate/

- `isolate/helpers.go:830-842` — `makeAndH()`
- `isolate/isolate.go:1708-1720` — `makeAnd()`

Identical conjunction-creation logic.

---

## MEDIUM — Similar Logic, Different Types or Signatures

### D9: `DistinctVariableRenaming()` — 2 packages

- `ast/rewrite.go:987-1006` — operates on `[]*Variable`, returns `map[string]Node`
- `module/batch16.go:134-174` — operates on `map[lg.NodeKey]lg.Expr`, returns `map[string]lg.Expr`

~85% identical algorithm (add prime suffixes until distinct). Differs in type signatures.

### D10: Substitute/Rename function families — ast/ vs module/

- `ast/rewrite.go:828-953` — `SubstituteAst()`, `SubstituteConstantsAst()`, `SubstituteConstantsAst2()`
- `module/astutil.go:113-265` — `SubstituteConstantsAST()`, `RenameAST()`, `RenameASTByName()`

Similar algorithms ported independently for `ast.Node` vs `lg.Expr` types.

### D11: `usedSymbolNames()` — 2 variants

- `actions/transrel.go:316-323` — returns `map[string]bool`
- `isolate/helpers.go:391-404` — returns `[]string`

Same concept (collect used symbol names), different return types.

### D12: `ConjToAssume()` — 2 variants in isolate/

- `isolate/create.go:859-866` — private, does NOT set Lineno
- `isolate/phase7.go:295-303` — public, DOES set Lineno

Near-identical except location tracking.

### D13: `IsTrue()` / `IsFalse()` — 5 packages

- `logic/formula.go:19-21` — canonical (checks empty And/Or)
- `logicutil/logic_utils.go:1229-1231` — same logic
- `ast/ast.go:1829-1831` — same logic (on ast.Node)
- `module/clauses.go:408-410` — delegates to `il.IsTrue()`
- `ivylogic/ivylogic.go:312-314` — delegates to `lg.IsTrue()`

Three implement the same check; two delegate.


---

## LOW — Trivial Wrappers / Delegation Chains

### D17: `SubstituteConstantsAction()` — trivial type-cast wrapper

- `actions/helpers.go:215-217` — wraps `mod.SubstituteConstantsAST()` with `.(Action)` cast

### D18: `formulaToClauses()` — hardcoded-nil wrapper

- `isolate/isolate.go:1766-1768` — wraps `module.FormulaToClauses(fmla, nil)`

### D19: `SubstituteConstantsExpr()` — trivial type-cast wrapper

- `module/astutil.go:139-141` — wraps `SubstituteConstantsAST()` with `.(lg.Expr)` cast

### D20: Map copy helpers — 10 identical-pattern functions

- `module/module.go:663-743` — `copyNodeSlice()`, `copyLFSlice()`, `copyMapLF()`,
  `copyMapSort()`, `copyMapIface()`, `copyMapAction()`, `copyMapNode()`,
  `copyMapNativeType()`, `copyMapStr()`, `copyMapBool()`

All follow the same `make` + `range` + `copy` pattern. Could use a single generic
`CopyMap[K,V]()` function with Go 1.18+ generics.

---

## Summary

| Severity | ID | Description | Lines Duplicated |
|----------|----|-------------|-----------------|
| HIGH | D3 | sortStrings() ×2 | ~14 |
| HIGH | D4 | GetCone() ×2 | ~50 |
| HIGH | D5 | GetPropsProvedInIsolateOrig() ×2 | ~40 |
| HIGH | D6 | makeAnd()/makeAndH() ×2 | ~26 |
| MEDIUM | D9 | DistinctVariableRenaming ×2 | ~60 |
| MEDIUM | D10 | Substitute/Rename families | ~250 |
| MEDIUM | D11 | usedSymbolNames() ×2 | ~16 |
| MEDIUM | D12 | ConjToAssume() ×2 | ~16 |
| MEDIUM | D13 | IsTrue()/IsFalse() ×5 | ~30 |
| LOW | D16 | ShortTypeName() chain | ~12 |
| LOW | D17 | SubstituteConstantsAction() | ~3 |
| LOW | D18 | formulaToClauses() | ~3 |
| LOW | D19 | SubstituteConstantsExpr() | ~3 |
| LOW | D20 | Map copy helpers ×10 | ~70 |
| | | **TOTAL** | **~764** |

## Notes

- D14 (IsSkolem) may be intentionally different — needs Python comparison before consolidating.
- D10 (Substitute/Rename) duplication is driven by type differences (ast.Node vs lg.Expr)
  and may require interface unification to fully resolve.
- D17-D19 wrappers may exist for API convenience; removal requires caller audit.
- This is an audit document. No fixes should be made from this plan alone.
