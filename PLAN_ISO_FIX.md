# Plan: Fix Bugs and Divergences in goivy/isolate vs Python ivy_isolate.py

## Context

The Go `isolate` package (`~/goivy/isolate/`) is a mechanical port of Python's `ivy_isolate.py` (`~/pyivy/ivy/ivy/ivy_isolate.py`). A thorough code review has identified **36+ divergences** across 8 Go files. These range from critical logic bugs to missing functionality. This plan organizes fixes by severity and file.

---

## CRITICAL Bugs (Break correctness)

### C1. `specAncestors` uses `continue` instead of `break`
- **File**: `helpers.go:440`
- **Python**: `ivy_isolate.py:655` — `break` stops ascending the hierarchy when "spec" child is found
- **Go**: Uses `continue`, which skips "spec" but keeps ascending — yields extra ancestors
- **Fix**: Change `continue` to `break` on line 441

### C2. `GetCalloutsAction` uses original `head` instead of mutated `h`
- **File**: `helpers.go:845`
- **Python**: `ivy_isolate.py:536-539` — mutates `head` then uses the mutated value in index
- **Go**: Computes `h` (mutated) but uses `head` (original) in the index. Line 858 `_ = h` confirms `h` is discarded
- **Fix**: Replace `head` with `h` at line 845 and remove `_ = h`

### C3. `StripSort` — relational sort with empty domain returns range instead of FunctionSort
- **File**: `strip.go:454-458`
- **Python**: `ivy_isolate.py:238-240` — when domain is empty but sort is relational, returns `FunctionSort(rng)` (a zero-arg relational sort)
- **Go**: Both `isBool` branches return `fs.Range()`, making the check dead code
- **Fix**: When `isBool` is true, return `NewFunctionSort(fs.Range())` instead of `fs.Range()`

### C4. `extModMixin` uses unconditional `"ext:"` prefix instead of conditional `prefixCallExt`
- **File**: `isolate.go:501`
- **Python**: `ivy_isolate.py:936-937` — `prefix_call_ext(name)` only prefixes names in `verified`; `mod_mixin` calls `m.prefix_calls(prefix_call_ext)` with a function, not a string
- **Go**: `actions.PrefixCalls(m, "ext:")` unconditionally prefixes ALL calls
- **Fix**: Create a `prefixCallExt` closure that checks `StartsWithSome(name, verified, mod, implMap)` before prefixing, and use `actions.PrefixCallsFunc(m, prefixCallExt)` (may need to add `PrefixCallsFunc` to the actions package if it doesn't exist)

### C5. `VStartsWithEqSome` and other functions called with `nil` implMap throughout `IsolateComponent`
- **File**: `isolate.go:497,517,518` and many others
- **Python**: `ivy_isolate.py:178-184` — always uses global `implementation_map`
- **Go**: Passes `nil` for the implementation map parameter, skipping the lookup
- **Fix**: Pass the local `implementationMap` variable at every call site within `IsolateComponent`

### C6. `GetCallsModsRec` missing mixin dependency tracking
- **File**: `deps.go:76-149`
- **Python**: `ivy_isolate.py:483-521` — tracks 4 maps (calls, mods, mixins, loops); follows mixin call chains at lines 511-520
- **Go**: Only tracks calls, mods, loops; no `mixins` map; no mixin processing loop
- **Fix**: Add `mixins map[string]map[string]bool` parameter; add loop iterating `mod.Mixins[actname]` to follow mixin deps; merge `mixins[calledname]` into `acalls` (Python line 506)

### C7. `CheckInterferenceFull` missing `locmods` and `pre_refed` checks
- **File**: `deps.go:254-421`
- **Python**: `ivy_isolate.py:586,594-605` — computes `locmods` (formal-parameter modifications) and `pre_refed` (symbols in unsummarized before-mixins); checks their intersection
- **Go**: Neither `locmods` nor `pre_refed` are computed or checked
- **Fix**: Add `locmods` computation via `GetLocMods`; compute `pre_refed` from unsummarized before-mixins; add intersection check per Python lines 603-605

---

## HIGH Bugs (Missing functionality)

### H1. `stripNatives` is a stub
- **File**: `strip.go:679-688`
- **Python**: `ivy_isolate.py:323-336` — builds `strip_binding` from `native.args[2:]`, strips formulas
- **Fix**: Implement full native stripping logic matching Python

### H2. `StripIsolateParams` missing variable parameter substitution
- **File**: `strip.go:492-496`
- **Python**: `ivy_isolate.py:345-352` — substitutes `Variable` isolate params with `iso:` prefix
- **Fix**: Add variable detection and `SubstituteAst` call for variable params

### H3. `StripIsolateParams` missing parameter validation
- **File**: `strip.go:498-503`
- **Python**: `ivy_isolate.py:355-369` — validates unbound params, checks atom args are simple Apps, checks no symbol redefinition
- **Fix**: Add three validation checks matching Python lines 355-369

### H4. `StripLabeledFormulas` missing SchemaBody check
- **File**: `strip.go:427-436`
- **Python**: `ivy_isolate.py:317-318` — raises error "cannot strip parameter from a theorem"
- **Fix**: Add type check for SchemaBody and error

### H5. `IsolateComponent` missing `enforce_axioms` action implementation check
- **File**: `isolate.go:976-993`
- **Python**: `ivy_isolate.py:1225-1239` — checks present actions calling non-present non-`imp__` actions have implementations; checks definitions referenced in `all_syms` are present
- **Fix**: Add both checks matching Python

### H6. `IsolateComponent` missing DefinitionSchema-to-Definition conversion
- **File**: `isolate.go` (missing entirely)
- **Python**: `ivy_isolate.py:1249-1256` — converts DefinitionSchema to Definition for exact_present names
- **Fix**: Add conversion pass after definition filtering

### H7. `IsolateComponent` missing `determined` set filtering in `enforce_axioms`
- **File**: `isolate.go:976-993`
- **Python**: `ivy_isolate.py:1216-1220` — builds `determined` set from deterministic defs and mod.params; skips those symbols
- **Fix**: Build `determined` set and filter before checking dropped axioms

### H8. `IsolateComponent` nil isolate creates nil instead of default
- **File**: `isolate.go:301-311`
- **Python**: `ivy_isolate.py:891-893` — creates `IsolateDef(Atom('iso'), Atom('this'))` when `isolate_name is None`
- **Fix**: Create default isolate with "this" when isolate name is empty

### H9. `CreateIsolate` passes `nil` for `extra_with`, `extra_strip`, `after_inits`
- **File**: `create.go:124`
- **Python**: `ivy_isolate.py:1687-1688` — passes computed values
- **Fix**: Pass the computed `extraWith`, `extraStrip`, and `afterInits` values

### H10. `CreateIsolate` missing `create_imports` logic
- **File**: `create.go:189-252`
- **Python**: `ivy_isolate.py:1609-1673` — complex outcall detection, `imp__` action creation, attribute copying, `fixit()` variable handling
- **Fix**: Implement full create_imports block

---

## MEDIUM Bugs (Subtle correctness issues)

### M1. `HasSideEffectRec` missing `Ranking` check
- **File**: `phase7.go:101-139` and `deps.go:156-219`
- **Python**: `ivy_isolate.py:472-473` — `isinstance(sub, ia.Ranking)` returns True
- **Fix**: Add Ranking action type check

### M2. `IsolateComponent` `all_syms` not normalized via `normalize_symbol`
- **File**: `isolate.go:907-935`
- **Python**: `ivy_isolate.py:1187` — calls `ivy_logic.normalize_symbol`
- **Fix**: Add normalization step

### M3. `IsolateComponent` SchemaBody formulas not excluded from first symbol pass
- **File**: `isolate.go:908-916`
- **Python**: `ivy_isolate.py:1179` — filters out SchemaBody
- **Fix**: Add SchemaBody type check in symbol collection loop

### M4. `StripActionFull` assignment interference is warning instead of error
- **File**: `strip.go:148-163`
- **Python**: `ivy_isolate.py:259` — `raise iu.IvyError`
- **Fix**: Change `fmt.Printf` warning to return error

### M5. `StripActionFull` AssignAction does not validate extra bindings are variables
- **File**: `strip.go:107-119`
- **Python**: `ivy_isolate.py:263-264` — checks `is_variable(x)` for each extra binding
- **Fix**: Add variable check and error

### M6. `stripNode` does not skip constructors
- **File**: `strip.go:291-330`
- **Python**: `ivy_isolate.py:275` — `ast not in im.module.sig.constructors`
- **Fix**: Add constructor check before stripping Apply nodes

### M7. `IsolateComponent` untrusted native check too broad
- **File**: `isolate.go:1202`
- **Python**: `ivy_isolate.py:1365` — exact type check `type(isolate) == ivy_ast.IsolateDef`
- **Fix**: Check for exact IsolateDef type, not just "not extract"

### M8. `IsolateComponent` missing definition NativeExpr check in untrusted isolate
- **File**: `isolate.go` (missing)
- **Python**: `ivy_isolate.py:1369-1371`
- **Fix**: Add definition NativeExpr check

### M9. `IsolateComponent` `init_cond` not set when no inits
- **File**: `isolate.go:1214-1225`
- **Python**: `ivy_isolate.py:1388` — always creates And (even if empty = true)
- **Fix**: Set `mod.InitCond` to empty And when no labeled inits

### M10. `StripSort` missing error for domain length mismatch
- **File**: `strip.go:447-448`
- **Python**: `ivy_isolate.py:416-417` — raises error
- **Fix**: Return error instead of silently returning original sort

### M11. `StripIsolateParams` wrong lookup key for param sorts
- **File**: `strip.go:530-546`
- **Python**: `ivy_isolate.py:449` — uses `s.sort` (the type name), not param name
- **Fix**: Use the sort name from the parameter, not the parameter name itself

### M12. `CreateIsolate` missing delegate validations
- **File**: `create.go:74-79`
- **Python**: `ivy_isolate.py:1593` — checks delegee is in hierarchy
- **Fix**: Add hierarchy membership check for delegees

### M13. `CreateIsolate` missing mixee action validation
- **File**: `create.go:66-72`
- **Python**: `ivy_isolate.py:1587` — validates both mixer AND mixee
- **Fix**: Add LookupAction check for mixee

### M14. `getPrivateFromAttributes` missing error for invalid values
- **File**: `helpers.go:189-203`
- **Python**: `ivy_isolate.py:700-701` — raises error for values not in `['spec','impl','priv']`
- **Fix**: Add default error case in switch

### M15. `CreateIsolate` ordering: initializer extraction before with-parameter/conjecture checks
- **File**: `create.go:57-104`
- **Python**: `ivy_isolate.py:1570-1578` — checks with-parameters first
- **Fix**: Reorder to match Python sequence

---

## LOW Priority (Minor, cosmetic, or edge-case only)

### L1. `CreateIsolate` missing `show_compiled` feature (`create.go`)
### L2. `CreateIsolate` missing `pedantic` warning logic (`create.go:327-334`)
### L3. `CreateIsolate` `slv.check_compat()` ordering divergence (`create.go:271`)
### L4. `CreateIsolate` no-isolate path doesn't record isolate_info implementations/monitors (`create.go:133-183`)
### L5. `find_references` uses string names vs Python Symbol objects (`helpers.go:938-975`)
### L6. `IsolateComponent` natives filter uses `lf.Label` vs Python `c.args[0]` (`isolate.go:836-844`)
### L7. `IsolateComponent` second `all_syms` pass missing natives `args[2:]` (`isolate.go:1066-1070`)
### L8. `GetIsolateLFs` label key format may differ (`iter.go:164`)
### L9. `stripNode` no separate Atom handling (`strip.go:291-330`)

---

## Recommended Fix Order

**Phase 1 — Critical bugs (C1-C7)**: These break core logic and are relatively contained fixes.
1. C1 `specAncestors` break vs continue (1 line)
2. C2 `GetCalloutsAction` head vs h (2 lines)
3. C3 `StripSort` relational sort (3 lines)
4. C4 `extModMixin` conditional prefix (10-15 lines + possible new function in actions pkg)
5. C5 nil implMap at call sites (change ~5 call sites)
6. C6 `GetCallsModsRec` mixin tracking (~30 lines)
7. C7 `CheckInterferenceFull` locmods/pre_refed (~40 lines)

**Phase 2 — High bugs (H1-H10)**: Missing functionality blocks.

**Phase 3 — Medium bugs (M1-M15)**: Subtle correctness and validation gaps.

**Phase 4 — Low priority (L1-L9)**: Edge cases and cosmetic alignment.

---

## Verification

After each phase:
1. Run `go build ./isolate/...` to verify compilation
2. Run `go test ./isolate/...` to verify existing tests pass
3. For each fix, consider adding a targeted test that exercises the corrected behavior
4. Cross-reference fixed Go code against corresponding Python lines to confirm faithful port

---

## Key Files to Modify

| File | Issues |
|------|--------|
| `helpers.go` | C1, C2, M14 |
| `strip.go` | C3, H1, H2, H3, H4, M4, M5, M6, M10, M11 |
| `isolate.go` | C4, C5, H5, H6, H7, H8, M2, M3, M7, M8, M9 |
| `deps.go` | C6, C7, M1 |
| `create.go` | H9, H10, M12, M13, M15 |
| `phase7.go` | M1 |
| `iter.go` | (minor) L8 |
