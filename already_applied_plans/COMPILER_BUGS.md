# Plan: Fix compiler/ Test Failures

**Created:** 2026-03-25T12:00

## Context

Six compiler/ tests are red. Root cause analysis reveals two distinct bugs introduced during the de-globalization refactor, plus one test that needs post-fix verification.

**Failing tests:**
1. `TestCheckDefinitions_SeparatesDefsFromProps` — "expected 1 definition, got 0"
2. `TestCheckDefinitions_StaleSymbols` — "defA should be in Definitions"
3. `TestCheckDefinitions_StaleSymbolTransitive` — "defF should be in Definitions"
4. `TestInferParameters_ExtendsFormals` — "expected *ast.Atom, got *ast.App"
5. `TestInferParameters_BodyRewritten` — "expected body atom to be 'fml:p', got 'p'"
6. `TestRegression_VarInActionBody` — "cfabric.step body should assign rd_pio_fair but doesn't: {}"

---

## Bug 1: `prefixNodes` converts Atoms to Apps (causes tests 4, 5, likely 6)

**File:** `ast/decl.go:362-368`

**Root cause:** The `prefixNodes` function (used by `NewActionDef` to create `fml:`-prefixed formal params) has an incorrect comment and implementation. It converts `*Atom` → `*App`:

```go
case *Atom:
    // Python: Atom inherits from App; prefix returns same type (App).  ← WRONG
    app := &App{Rep: &Symbol{Rep: s + a.Rep}, Terms: a.Terms}
    ...
    result[i] = app
```

**Python truth:** In `ivy_ast.py:287-292`, `Atom.prefix()` calls `self.clone(self.args)` → `type(self)(self.rep, list(args))` which returns another **Atom** (type-preserving). `Atom` does NOT inherit from `App` in Python — they are sibling classes both inheriting from `Formula`.

**Impact:**
- `ActionDef.FormalParams` contains `*App` nodes instead of `*Atom` nodes
- `InferParameters` propagates these Apps to mixer actions → tests 4, 5 fail
- `CompileAction` receives Apps from `Formals()` → trace shows `origParam p = *ast.App` → may cause downstream compilation failures → test 6

**Fix:** Use the existing `Atom.Prefix()` method (`ast/ast.go:249-253`) which already returns `*Atom`:

```go
case *Atom:
    // Python: Atom.prefix() returns another Atom (type-preserving clone).
    result[i] = a.Prefix(s)
```

---

## Bug 2: Test helper ID collision (causes tests 1, 2, 3)

**File:** `compiler/batch_f_test.go:43-53`

**Root cause:** `makeLabeledDef` and `makeLabeledFormula` each create a **fresh** `ast.NewAstConfig()`. Since `NextLFID()` starts from 0 on each fresh config, ALL labeled formulas get `ID=0`.

In `CheckDefinitions` (`ivy_compile.go:1330-1343`), proofs are indexed by `prop.ID`:
```go
withProofs[pe.Formula.ID] = true  // defC.ID = 0
...
if !withProofs[prop.ID] {  // defA.ID = 0 → withProofs[0] = true → WRONG
```

Since `defA.ID == defC.ID == 0`, defA is incorrectly treated as having a proof.

**Fix:** Change helpers to accept `*ast.AstConfig` from the test's shared config:

```go
func makeLabeledDef(cfg *ast.AstConfig, label string, def *lg.Definition) *ast.LabeledFormula {
    return cfg.NewLabeledFormula(cfg.NewAtom(label), def)
}

func makeLabeledFormula(cfg *ast.AstConfig, label string, formula ast.Node) *ast.LabeledFormula {
    return cfg.NewLabeledFormula(cfg.NewAtom(label), formula)
}
```

Then update all callers (tests 1-29+) to pass `cfg` or `mod.Cfg.AstCfg`.

---

## Test 6 verification: `TestRegression_VarInActionBody`

This test parses real Ivy source through `parser.New()` → `IvyCompile()`. The `prefixNodes` bug causes action formal params to be Apps. After Fix 1, params will be correct Atoms. If the test still fails, the issue is in object expansion or `CompileActionBody` and will need separate investigation.

---

## Implementation Steps

### Step 1: Fix `prefixNodes` in ast/decl.go

**File:** `ast/decl.go:362-368`

Replace the Atom case:
```go
case *Atom:
    result[i] = a.Prefix(s)
```

This is a 1-line fix. `Atom.Prefix()` at `ast/ast.go:249` already handles cloning, Rep prefixing, and lineno preservation correctly.

### Step 2: Fix test helpers in batch_f_test.go

**File:** `compiler/batch_f_test.go:43-53`

1. Change `makeLabeledDef(label, def)` → `makeLabeledDef(cfg, label, def)`
2. Change `makeLabeledFormula(label, formula)` → `makeLabeledFormula(cfg, label, formula)`
3. Update all call sites (scan the whole file — ~29 tests use these helpers) to pass the test's `cfg` or `mod.Cfg.AstCfg`

### Step 3: Run tests, verify

```bash
go build ./...
go test ./compiler/ -v -run 'TestCheckDefinitions|TestInferParameters|TestRegression_VarInActionBody'
```

If test 6 still fails after Step 1, investigate `CompileActionBody` error path for object actions.

### Step 4: Run full test suite

```bash
go test ./compiler/
go test ./ast/
```

Ensure no regressions.

---

## Files to Modify

| File | Change |
|------|--------|
| `ast/decl.go` | Fix `prefixNodes` Atom case (line 362-368) |
| `compiler/batch_f_test.go` | Fix helpers + all callers |

## Verification

1. `go build ./...` — clean build
2. `go test ./compiler/` — all 6 previously-failing tests pass
3. `go test ./ast/` — no regressions
4. `go test ./compiler/ -count=1` — no flaky ID collisions
