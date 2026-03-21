# Plan: Fix Section 10 (ivy_isolate.py → isolate/) Audit Issues

## Context

The AUDIT18MARCH.md section 10 identifies 15 issues across the Go `isolate/` package — 7 MISSING, 3 STUB, and 5 BEHAVIORAL_DIFFERENCE items. These issues represent gaps between the Python `ivy_isolate.py` source of truth and the Go port. The goal is to close all gaps so the Go port faithfully reproduces the Python behavior.

---

## Issue Inventory (15 total)

### 10.1 MISSING
| # | Issue | Batch |
|---|-------|-------|
| M1 | `check_isolate_completeness()` missing caller/callee assertion verification | 6 |
| M2 | `has_assertions`/`has_requires` fragile string-based type check | 2 |
| M3 | `strip_action()` missing `strip_binding`, `init_params`, `modifies()` checks | 3 |
| M4 | `strip_labeled_fmla()` missing `get_strip_binding` call | 3 |
| M5 | `isolate_component()` `create_imports` logic not implemented | 5 |
| M6 | `isolate_component()` `slv.check_compat()` not called | 5 |
| M7 | `isolate_component()` `canonize_types` not called | 5 |

### 10.2 STUB
| # | Issue | Batch |
|---|-------|-------|
| S1 | `GetIsolateAttr` always returns defaultVal | 1 |
| S2 | `isExplicitOnly()` always returns false | 1 |
| S3 | `IsolateComponent()` ExtAction — sets flag but doesn't create combined action | 5 |

### 10.3 BEHAVIORAL_DIFFERENCE
| # | Issue | Batch |
|---|-------|-------|
| B1 | Action classification: string `.Name()` vs Python `isinstance()` | 2 |
| B2 | `check_interference()` missing Ranking/WhileAction loop/termination checking | 4 |
| B3 | `strip_isolate()` missing version 1.6 `mod.params` clearing | 1 |
| B4 | `get_mixin_order()` name extraction via `relNamer` interface vs `.relname` | 1 |
| B5 | `FixInitializers` missing `type_check_action` call | 1 |

---

## Batches (ordered by recommended execution order)

### Batch 1: Quick Fixes (5 issues: S1, S2, B3, B4, B5)
**Complexity: Low** — Each is a small, isolated change.

#### B5: `FixInitializers` missing `type_check_action`
- **Discovery**: Python's `type_check_action()` at `ivy_actions.py:1329` is **disabled** — it immediately returns with no effect. The function body is commented out.
- **Action**: Add a no-op `TypeCheckAction()` call in Go's `FixInitializers` for structural parity (with a comment noting it's disabled in Python too). Or simply add a comment documenting this. Minimal risk.
- **File**: `isolate/create.go`

#### B3: `strip_isolate()` version 1.6 `mod.params` clearing
- Python (line ~439): `if iu.version_le(iu.get_string_version(), "1.6"): del mod.params[:]`
- **Action**: Add version check in `StripIsolate()` or `StripIsolateParams()` — if version ≤ 1.6, clear `mod.Params` slice.
- **File**: `isolate/strip.go`
- **Depends on**: `ivyutils.VersionLe()` and `ivyutils.GetStringVersion()` existing in Go (verify first).

#### B4: `get_mixin_order()` name extraction
- Python uses `rdf.args[0].relname` to get mixin ordering arc names.
- Go uses `fmt.Sprint(args[0])` via a `relNamer` interface — may produce different string representation.
- **Action**: Ensure Go AST nodes for MixOrd entries expose a `Relname()` method (or equivalent). Update `GetMixinOrder` to call it instead of `fmt.Sprint()`.
- **File**: `isolate/create.go` (lines ~402-484)

#### S1: `GetIsolateAttr` stub
- Currently always returns `defaultVal`.
- **Action**: Port Python's attribute lookup — check `mod.Attributes` for the isolate name and return the attribute value if found.
- **File**: Find where `GetIsolateAttr` is defined (likely `isolate/isolate.go` or `isolate/helpers.go`), implement real lookup.

#### S2: `isExplicitOnly()` stub
- Currently always returns `false`.
- **Action**: Port Python logic — check if the isolate definition has the "explicit" attribute/flag set.
- **File**: Same location as `isExplicitOnly`.

---

### Batch 2: Action Type Classification (2 issues: M2, B1)
**Complexity: Medium** — Touches multiple call sites but is a consistent pattern.

The core problem: Python uses `isinstance(action, ia.AssertAction)` which matches AssertAction AND all subclasses (RequiresAction, EnsuresAction, SubgoalAction). Go uses `.Name() == "assert"` which only matches AssertAction exactly.

#### Fix Strategy
1. Add a helper function `IsAssertLike(action Action) bool` that checks if the action is any of `*AssertAction`, `*RequiresAction`, `*EnsuresAction`, `*SubgoalAction` (since they all embed AssertAction in Go).
2. Update `HasAssertions()` in `iter.go:274-285` to use the helper instead of `sub.Name() == "assert"`.
3. Update `HasRequires()` in `iter.go:289-300` — this one is actually correct for its specific purpose (only matching RequiresAction), but verify against Python.
4. Update `FindSomeAssertion()` in `phase7.go:299-310` — currently uses `sub.(*actions.AssertAction)` which misses subclasses.
5. Update `HasSideEffectRec()` in `deps.go:148` — same issue.
6. Audit all `.Name() ==` checks across `isolate/` for correctness.

**Files**: `isolate/iter.go`, `isolate/phase7.go`, `isolate/deps.go`, `actions/action.go` (for helper)

---

### Batch 3: Strip Binding System (2 issues: M3, M4)
**Complexity: High** — Requires implementing a new subsystem (`strip_binding` dictionary mechanism).

#### What's Missing
1. **`get_strip_binding()` function** (Python lines 291-301): Walks the AST extracting parameter-to-argument mappings into a `strip_binding` dict. This is completely absent in Go.
2. **`strip_action()` enhancements** (Python lines 242-289):
   - Add `stripBinding map[lg.NodeKey]string` parameter
   - Add `isInit bool` parameter
   - Add `initParams []lg.Expr` parameter
   - Add `modifies()` interference checking (lines 250-259)
   - Add constant/variable substitution from strip_binding (lines 268-273)
   - Add Atom node handling (lines 283-288)
3. **`strip_labeled_fmla()` enhancement**: Call `get_strip_binding()` before `strip_action()`.

#### Implementation Steps
1. Implement `GetStripBinding(action actions.Action, stripMap StripMap, stripBinding map[lg.NodeKey]string)` in `strip.go`.
2. Add `Modifies()` method to the Action interface (or specific action types) if not present.
3. Extend `StripAction` / `stripActionRec` signatures with the new parameters.
4. Update `StripLabeledFormula` to call `GetStripBinding` first.
5. Update all callers of these functions.

**Files**: `isolate/strip.go`, `actions/action.go` (Modifies method), `isolate/phase7.go` (GetStripParams)

---

### Batch 4: Interference / Loop Checking (1 issue: B2)
**Complexity: High** — Requires WhileAction/Ranking detection infrastructure.

#### What's Missing
Python's `check_interference()` (lines 577-642) and `get_calls_mods()` (lines 483-520):
- Detects `WhileAction` nodes without `Ranking` clauses (line 509)
- Tracks loops per action in a `loops` dict
- Checks `check_term` flag for loop termination validation (lines 611-615)
- Uses `get_loc_mods()` (lines 560-563) to find local formula modifications (`fml:` prefixed)

#### Implementation Steps
1. Implement `GetLocMods()` function in `deps.go` — extract symbols starting with `fml:` from `action.Modifies()`.
2. Extend `GetCallsModsRec()` to return loops (WhileAction without Ranking) — add a `loops` parameter.
3. Add WhileAction/Ranking detection in the recursive walk (check `*actions.WhileAction` type, check if last arg is `*actions.Ranking`).
4. Extend `CheckInterferenceFull()` to use loops and `checkTerm` for termination validation.
5. Add `pre_refed` symbol tracking from before-mixins.

**Files**: `isolate/deps.go`, possibly `actions/action.go` (WhileAction/Ranking types)

---

### Batch 5: `isolate_component()` Features (4 issues: M5, M6, M7, S3)
**Complexity: High** — These are features within the ~500-line `isolate_component()` function.

#### M5: `create_imports` logic
- Python creates import actions for out-calls and external stubs when `create_imports` flag is set.
- This is gated by a parameter flag, so it may not be exercised in normal verification mode.
- **Action**: Port the `create_imports` branch from Python `isolate_component()` (~50 lines).

#### M6: `slv.check_compat()` — native interpretation checking
- Python calls `slv.check_compat()` to verify that native type interpretations are compatible.
- **Action**: Implement the call in Go's `IsolateComponent()`. Requires `solver.CheckCompat()` to exist (may need porting in solver package first).
- **Dependency**: May depend on solver package having `CheckCompat`.

#### M7: `canonize_types` not called
- Python calls `mod.canonize_types()` near the end of `isolate_component()`.
- **Action**: Add `mod.CanonizeTypes()` call in Go. Verify `CanonizeTypes` exists on Module (audit item 11.1.3 notes it's incomplete too — may need partial fix there first).

#### S3: ExtAction combined action creation
- Python creates a combined external action from all exported actions.
- Go sets the `ExtAction` flag but doesn't build the combined action.
- **Action**: Port the action combination logic from Python.

**Files**: `isolate/isolate.go` or `isolate/helpers.go` (IsolateComponent), `solver/` package (CheckCompat), `module/` package (CanonizeTypes)

---

### Batch 6: `check_isolate_completeness()` Full Implementation (1 issue: M1)
**Complexity: High** — The Go version is dramatically simplified vs Python.

#### What's Missing
Python (lines 1804-1916) has multi-level verification:
- Tracks `checked`, `checked_props`, `checked_context`, `verified_context` maps
- Verifies caller/callee assertion relationships
- Handles delegate-specific rules
- Checks mixin verification against mixed actions
- Verifies requires assertions
- Produces detailed multi-line error messages

Go (lines 310-364) only checks that properties appear somewhere.

#### Implementation Steps
1. Study Python lines 1804-1916 in detail (read full function).
2. Add `checkedContext` and `verifiedContext` maps.
3. Port the caller/callee assertion iteration loop.
4. Port delegate handling logic.
5. Port mixin verification checks.
6. Update `IsolateError` struct if needed for richer error context.
7. Add tests for each verification scenario.

**Files**: `isolate/iter.go` (CheckIsolateCompleteness)

---

## Verification Plan

For each batch:
1. Run existing tests: `cd /Users/jaten/go/src/github.com/glycerine/goivy && go test ./isolate/...`
2. Add new tests for ported functionality where Python behavior is well-defined.
3. Cross-reference against Python by running equivalent test cases in both.
4. Run full test suite: `go test ./...`

## Recommended Execution Order

1. **Batch 1** (Quick Fixes) — unblocks nothing but easy wins
2. **Batch 2** (Action Type Classification) — foundational fix needed by Batches 4, 5, 6
3. **Batch 3** (Strip Binding) — independent subsystem
4. **Batch 4** (Interference/Loop) — depends on Batch 2
5. **Batch 5** (isolate_component features) — depends on Batches 2, 3
6. **Batch 6** (completeness check) — depends on Batch 2
