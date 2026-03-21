
### 2.1 Explicit TODOs

| # | File:Line | Description | Status |
|---|-----------|-------------|--------|
| 8 | `end2end/conform_test.go:232` | `TODO: compile and check with Go, compare against Python` | DEFERRED — test infrastructure |

---


## 10. ivy_isolate.py → isolate/

### 10.1 MISSING

| # | Python Function | Description |
|---|-----------------|-------------|
| 1 | `check_isolate_completeness()` full impl | Only checks property labels, missing caller/callee assertion verification |
| 2 | `has_assertions`/`has_requires` | Fragile type assertion on `interface{}` instead of `isinstance(action, AssertAction)` |
| 3 | `strip_action()` full complexity | Missing interference checking (`modifies()`), `init_params`, `strip_binding` |
| 4 | `strip_labeled_fmla()` strip binding | No `get_strip_binding` call |
| 5 | `isolate_component()` create_imports | Import actions, out-calls, external stubs not implemented |
| 6 | `isolate_component()` `slv.check_compat()` | Native interpretation checking missing |
| 7 | `isolate_component()` `canonize_types` | Not called |

### 10.2 STUB

| # | Go Function | Issue |
|---|-------------|-------|
| 1 | `GetIsolateAttr` | Always returns `defaultVal` |
| 2 | `isExplicitOnly()` | Always returns false |
| 3 | `IsolateComponent()` ExtAction | Sets flag but doesn't create combined action |

### 10.3 BEHAVIORAL_DIFFERENCE

| # | Area | Issue |
|---|------|-------|
| 1 | Action classification | String-based kinds (`"assert"`) vs Python type references (`ia.AssertAction`) |
| 2 | `check_interference()` | Missing Ranking/WhileAction loop/termination checking |
| 3 | `strip_isolate()` | Missing version 1.6 `mod.params` clearing |
| 4 | `get_mixin_order()` | Different name extraction via `relNamer` interface vs `.relname` |
| 5 | `FixInitializers` | Missing `type_check_action` call |

---

## 11. ivy_module.py → module/

### 11.1 MISSING

| # | Python Function | Description |
|---|-----------------|-------------|
| 1 | `Module.__enter__`/`__exit__` | Does not save/restore `il.sig` or call `solver.clear()` |
| 2 | `Module.InitCond` initialization | Not initialized to `lu.true_clauses()` in `Clear()` |
| 3 | `CanonizeTypes` fields | Missing: `InitCond`, `ConceptSpaces`, `Progress`, `BeforeExport`, `lu.resort_sig` |
| 4 | `find_action` free function | Only exists as method, not package-level |
| 5 | `background_theory` free function | Only exists as method |
| 6 | `param_logic` parameter and `logics()` | Hardcodes `["epr"]` default |
| 7 | `sort_dependencies` NativeTypes branch | Skips native_types check |
| 8 | `CallGraph()` | Placeholder returning empty map |
| 9 | `UpdateConjs()` | Placeholder, no concept space creation |
| 10 | `copy()` | Missing many map copies: `SortDestructors`, `SortConstructors`, `Variants`, `Supertypes`, `ExtPreconds`, `ConjActions`, `IsolateProofs` |

### 11.2 BEHAVIORAL_DIFFERENCE

| # | Area | Issue |
|---|------|-------|
| 1 | `BackgroundTheory` | Package-level `theoryCache` with mutex (memory leak) instead of `self.theory` field |
| 2 | `SortCard` | `val.(string)` type assertion vs `self.attributes[attr].rep` |
| 3 | `VariantIndex` return | Go returns `-1`, Python returns `None` |
| 4 | `resort_aliases_map` | Go correctly returns result; Python has bug (no `return`) |
| 5 | `Exclusivity()` | Missing coverage axiom (disjunction that some variant must hold) |

---


7. **`check/` package largely non-functional** — `CheckProperties`, `CheckConjectures`, `CheckTemporals` are no-ops (§9)

### Medium Priority (Behavioral Fidelity)

16. `SetStringVersion` compose character mismatch (§12.2, item 1)
17. `GetStdIncludeDir` too simplistic (§12.2, item 2)
20. `PostState` missing proper havocing (§8.3, item 7)
