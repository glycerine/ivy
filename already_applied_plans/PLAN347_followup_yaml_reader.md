# Deep Conformance Audit: goivy/check vs pyivy/ivy_check.py

**Created: 2026-04-29 ~UTC**

## Context

The Go `check/` package is a mechanical port of Python `ivy_check.py` — the main verification orchestrator for the Ivy formal methods tool. Python is the source of truth. After ~1030 divergences already fixed, this audit found **7 confirmed behavioral divergences** through line-by-line comparison of `check_isolate`, `check_subgoals`, `check_module`, `MCIsolate`, and helper functions.

---

## Confirmed Divergences

### D1. `check_subgoals`: `return` vs `continue` on SKIPPED

**Files:** `isolate_check.go:829-834` vs `ivy_check.py:836-838`

**Python:** `if act.check_unprovable.get(): print("SKIPPED\n"); return` — exits the entire `check_subgoals` function, stopping all remaining goals.

**Go:** `fmt.Println("SKIPPED"); ... continue` — continues to the next goal.

**Sub-issue:** Go enters `TheoryContext()` BEFORE the unprovable check (line 827-829). Python checks BEFORE entering theory context (line 836 vs 839). Go does unnecessary work and potential state side-effects.

**Fix:**
1. Change `continue` to `return nil` at line 834.
2. Move the `OnlyCheckUnprovable` check BEFORE `cleanup := withLocalMod.TheoryContext()`.
3. Apply the same fix in BOTH the temporal branch (line 827) and the non-temporal branch (line 923).

---

### D2. `check_isolate`: Guarantee/assumption phase entirely gated on `check`

**Files:** `isolate_check.go:442` vs `ivy_check.py:667-760`

**Python structure:**
```
callgraph = ...                          # ALWAYS built
for actname, action in mod.actions:      # ALWAYS iterated
    assumptions = ...                    # ALWAYS printed
for actname, action in mod.actions:      # ALWAYS iterated  
    guarantees = ...
    if guarantees and not no_check_guarantees:
        print header                     # ALWAYS printed
        for sub in guarantees:
            if check and any(untried):   # ONLY verification gated on check
                ...verify...
            else:
                print("")
```

**Go:** `if !mod.Cfg.NoCheckGuarantees && check {` at line 442 wraps ALL of: callgraph construction, assumption printing, guarantee printing, AND verification.

**Fix:** Restructure to match Python:
1. Move callgraph construction (lines 444-451) OUTSIDE the `check` gate.
2. Move assumption printing loop (lines 464-505) OUTSIDE both gates.
3. Keep guarantee loop always running; only gate the SMT verification on `check`.
4. Gate guarantee header/checking on `!NoCheckGuarantees` (matching Python line 718), but NOT on `check`.

---

### D3. `check_isolate`: Initializer guarantee header gated on `check`

**Files:** `isolate_check.go:338` vs `ivy_check.py:634-639`

**Python:** `if guarantees and not unprovable:` prints header always. `if check:` only gates verification.

**Go:** `if len(guarantees) > 0 && !mod.Cfg.OnlyCheckUnprovable && check {` — header gated on `check` too.

**Fix:** Split into two conditions:
```go
if len(guarantees) > 0 && !mod.Cfg.OnlyCheckUnprovable {
    fmt.Print("\n    Any assertions in initializers must be checked ")
    if check {
        // ... actual checking ...
    }
}
```

---

### D4. `MCIsolate` per-assertion loop: Theory context on wrong module

**Files:** `isolate_check.go:1219-1226` vs `ivy_check.py:936-946`

**Python:** `with im.module.copy():` makes the copy the current module. `meth()` runs on the copy, inside the copy's `theory_context()`.

**Go:** `modCopy := mod.Copy()` creates a copy, `cleanup := modCopy.TheoryContext()` enters theory context on the copy, but `method()` (a closure capturing the original `isoMod`) runs on the ORIGINAL — without the theory context.

**Fix:** The method closure needs to operate on the copy. Two approaches:
- **Option A (preferred):** Pass the module to MCIsolate's method as a parameter: `method func(m *module.Module) error`. The MCIsolate loop calls `method(modCopy)`. Each call site (mc, vmt, bmc) uses the passed module.
- **Option B:** Set `mod.Cfg.CheckLineno` on `modCopy.Cfg` instead of `mod.Cfg`, and ensure the method closure reads from the module it's given.

Also fix: `mod.Cfg.CheckLineno` should be set on `modCopy.Cfg`, not `mod.Cfg`.

---

### D5. `check_module`: ACL file loading unimplemented

**Files:** `acl/acl.go`, `isolate_check.go:1039,1068-1069`, `check.go:302-306` vs `ivy_check.py:982-983`

**Python:** `ivy_acl.register_from_file(opt_unchecked_properties.get())` (line 982-983):
1. Opens the YAML file, reads all lines, parses with `yaml.load()`
2. Expects optional keys `'ignores'` and `'assumes'`, each a list of strings
3. Strings can be literal names or `regex(PATTERN)` for regex matching
4. Calls `register_ignores()` and `register_assumes()` to populate globals
5. Called ONCE before the isolate loop

**Go:** `acl.NewConfig()` creates empty config. The `RegisterIgnores`/`RegisterAssumes` methods already handle literal names and `regex(...)` patterns correctly. Only `RegisterFromFile` is missing.

**Fix — 3 changes:**

**Change 0: `acl/acl.go` — fix regex matching semantics**

Python `is_ignored`/`is_assumed` use `re.match()` which only matches at the START of the string. Go uses `regexp.MatchString()` which matches ANYWHERE. Fix: anchor combined regex with `^(?:...)` in `RegisterIgnores`/`RegisterAssumes`:
```go
combined := "^(?:" + strings.Join(regexParts, "|") + ")"
```

**Change 1: `acl/acl.go` — add `RegisterFromFile`**
```go
func RegisterFromFile(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("acl: reading %s: %w", path, err)
    }
    var doc struct {
        Ignores []string `yaml:"ignores"`
        Assumes []string `yaml:"assumes"`
    }
    if err := yaml.Unmarshal(data, &doc); err != nil {
        return nil, fmt.Errorf("acl: parsing %s: %w", path, err)
    }
    cfg := NewConfig()
    if err := cfg.Register(doc.Ignores, doc.Assumes); err != nil {
        return nil, fmt.Errorf("acl: %s: %w", path, err)
    }
    return cfg, nil
}
```

**Change 2: `isolate_check.go` CheckModule — load ACL once before the loop**

Move the `opt_unchecked_properties` handling from inside the per-isolate loop to BEFORE the loop (matching Python line 982-983). Store the loaded `*acl.Config` and pass it into `PreprocessAssumedIgnoredProperties` for each isolate.

Current code (inside the loop):
```go
if mod.Cfg.OptUncheckedProps != "" {
    PreprocessAssumedIgnoredProperties(isoMod, acl.NewConfig())
}
```

New code (before the loop):
```go
var aclCfg *acl.Config
if mod.Cfg.OptUncheckedProps != "" {
    var err error
    aclCfg, err = acl.RegisterFromFile(mod.Cfg.OptUncheckedProps)
    if err != nil {
        return err
    }
}
```

Then inside the loop:
```go
if aclCfg != nil {
    PreprocessAssumedIgnoredProperties(isoMod, aclCfg)
}
```

**Change 3: `check.go` CheckTemporals — use loaded ACL config**

Currently creates a fresh empty config (line 304). Instead, accept an `*acl.Config` parameter OR read it from `mod.Cfg`. The simplest approach: store the loaded ACL config on `mod.Cfg` so `CheckTemporals` can access it without a signature change.

Since `module` doesn't import `acl` and we want to avoid adding that dependency, pass the `*acl.Config` through the call chain inside `check/`:

- `CheckTemporals(mod)` → `CheckTemporals(mod, aclCfg *acl.Config)` (add parameter; `nil` means no ACL filtering)
- `CheckIsolate` already calls `CheckTemporals` at the bottom. `CheckIsolate` doesn't have `aclCfg` in scope. Solution: store `aclCfg` on the `Module` as `interface{}` (already used for `TraceHook`) or pass through `CheckIsolate` → `CheckTemporals`. The simplest: add an `AclCfg interface{}` field to `module.Module` (not `Config`) — matching the existing `TraceHook interface{}` pattern. Then `CheckModule` sets `isoMod.AclCfg = aclCfg` before calling `CheckIsolate`, and `CheckTemporals` reads it.

Critical files:
- `acl/acl.go` — changes 0 and 1
- `check/isolate_check.go` — change 2 (CheckModule ACL load + isoMod.AclCfg assignment)
- `check/check.go` — change 3 (CheckTemporals reads mod.AclCfg)
- `module/module.go` — add `AclCfg interface{}` field to Module struct

---

### D6. `CheckLineno` format mismatch (three incompatible formats)

**Files:** `isolate_check.go:604,1222,329,538,1324`, `helpers.go:409-410`, `isolate_check.go:1446`

Three different formats are used in Go, none of which match each other:

| Location | Format | Example |
|----------|--------|---------|
| Guarantee inner loop (line 604) | `"%s:%d"` | `"foo.ivy:42"` |
| MCIsolate per-assertion (line 1222) | `":%d"` | `":42"` |
| FilterCheckers comparison (line 409-410) | `"line %d"` or `"%d"` | `"line 42"` or `"42"` |
| Inline guarantee filter (lines 329, 538) | `"%d"` | `"42"` |
| CheckConjsInStateWithAG (line 1446) | `"%d"` | `"42"` |
| AllAssertLinenos parse (line 1326) | `fmt.Sscanf("%d")` | extracts bare int |

The **actions layer** (`update.go:374-377`) already uses `"file:line"` format correctly — `fmt.Sprintf("%s:%d", loc.Filename, loc.Line)`. Tests confirm this (`audit51_test.go:614`, `div_conformance_test.go:237`). All other sites must align to this format.

**Fix:** Standardize ALL sites on `"file:line"` format for maximum readability:

1. **Guarantee inner loop** (`isolate_check.go:604`): Already correct — `fmt.Sprintf("%s:%d", lineno.Filename, lineno.Line)`. No change needed.

2. **MCIsolate per-assertion** (`isolate_check.go:1222`): Change `fmt.Sprintf(":%d", lineno)` to produce `"file:line"`. `AllAssertLinenos` must return `[]ast.Location` (not `[]int`) so the filename is available.

3. **AllAssertLinenos** (`isolate_check.go:1292-1335`): Change return type from `([]int, error)` to `([]ast.Location, error)`. Capture full `sub.GetLineno()` Location instead of just `.Line`. The `CheckLineno` filter at line 1324 must compare full `"file:line"` strings.

4. **FilterCheckers** (`helpers.go:409-410`): Replace the `"line %d"` / `"%d"` comparisons with `"file:line"`:
   ```go
   loc := cc.LF.GetLineno()
   locStr := fmt.Sprintf("%s:%d", loc.Filename, loc.Line)
   if locStr == checkLineno { result = append(result, fc) }
   ```

5. **Inline guarantee filter** (`isolate_check.go:329, 538`): Replace `fmt.Sprintf("%d", sub.GetLineno().Line) == mod.Cfg.CheckLineno` with:
   ```go
   loc := sub.GetLineno()
   fmt.Sprintf("%s:%d", loc.Filename, loc.Line) == mod.Cfg.CheckLineno
   ```

6. **CheckConjsInStateWithAG** (`isolate_check.go:1446`): Replace `fmt.Sprintf("%d", c.Lineno())` with:
   ```go
   loc := c.GetLineno()
   fmt.Sprintf("%s:%d", loc.Filename, loc.Line)
   ```

7. **AllAssertLinenos parse** (`isolate_check.go:1324-1326`): Replace `fmt.Sscanf(mod.Cfg.CheckLineno, "%d", &checkLine)` with full `"file:line"` string comparison against the locations in `seen`.

8. **check.go:1236, 1324** (`"none.ivy:0"` sentinel): These already use `"file:line"` format. No change needed.

9. **CheckConjsInState** (`check.go:743-751`): Same fix as item 6 — use `GetLineno()` to produce `"file:line"` for comparison.

---

### D7. `check_isolate`: Missing EvalContext(check=False) — RESOLVED (no fix needed)

**Files:** `isolate_check.go:269-274` vs `ivy_check.py:620-623`

**Python:** `with itp.EvalContext(check=False):` wraps BOTH `AnalysisGraph(initializer=...)` and `check_conjs_in_state(...)`.

**Analysis:** The Python `EvalContext.check` flag is ONLY read in `concrete_post` (ivy_interp.py:204): `cons = compose_state_action(..., check=context.check)`. This is called during action execution (`ag.execute`), NOT during conjecture checking (`check_fcs_in_state` → `history.satisfy`). The `History.satisfy` method (ivy_transrel.py:649) does not read the EvalContext at all.

Go's `AddInitialState` already passes `checkPrecondFalse` to `interp.ConcretePost` (art.go:1582), which matches the Python `EvalContext(check=False)` for action execution. The `CheckConjsInStateWithAG` call only does solver-based checking (not action execution), so the EvalContext is irrelevant there.

**Verdict:** No fix needed — Go already handles the check=False semantics correctly in `AddInitialState`.

---

## Verification Plan

After each fix, run:
```bash
cd ~/ivy/goivy && make test
```

For D5, add a unit test in `acl/acl_test.go` that creates a temp YAML file, calls `RegisterFromFile`, and checks that `IsIgnored`/`IsAssumed` return the correct values for both literal and regex entries.

---

## Implementation Order

1. **D6** (CheckLineno format) — DONE
2. **D1** (return vs continue) — DONE
3. **D3** (initializer header) — DONE
4. **D2** (guarantee phase restructure) — DONE
5. **D4** (MCIsolate theory context) — DONE
6. **D7** (EvalContext) — RESOLVED: no fix needed (Go already correct)
7. **D5** (ACL file loading) — remaining work
