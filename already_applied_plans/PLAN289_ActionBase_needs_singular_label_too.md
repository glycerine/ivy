# Fix ActionTerm labels divergence: `label` (singular) vs `labels` (plural)

Created: 2026-04-13 ~16:00 UTC

## Context

The `TestOrdLive` golden test produces an s-expression diff on `ActionTerm.Labels`:

```
-   labels:["ext:cfabric.step"]   (Go)
+   labels:["cf_live"]            (Python)
```

**Root cause:** Python actions have TWO independent attributes:
- `action.label` (singular, a `str`) -- a display/identification label, set by `CreateIsolate` for public actions and by `env_action` for branches
- `action.labels` (plural, a `list[str]`) -- isolate membership list, set by `handle_temporals`

Go's `ActionBase` has only ONE field: `Labels []string`. Both operations write to the same field. Since `HandleTemporals` runs **before** `CreateIsolate`, the correct isolate-membership labels `["cf_live"]` get **overwritten** by `CreateIsolate`'s `SetLabels([]string{"ext:cfabric.step"})`.

## Execution order (both languages)

```
HandleTemporals(mod)       // sets action.labels = sorted(imap[actname])  -> ["cf_live"]
CreateIsolate("this", mod) // Python: action.label = name  (different field!)
                           // Go:     SetLabels([]string{name})  (OVERWRITES Labels!)
```

Then `NormalProgramFromModule` -> `OldActionToNew` reads `act.labels`/`act.Labels` to populate `ActionTerm.Labels`.

## Plan

### Step 1: Add `Label` (singular) field to `ActionBase`

**File:** `module/action.go`

- Add `Label string` field to `ActionBase` struct (after `Labels []string`)
- Add methods:
  ```go
  func (b *ActionBase) SetLabel(label string) { b.Label = label }
  func (b *ActionBase) GetLabel() string       { return b.Label }
  ```
- Do NOT copy `Label` in `CopyFormalsTo` (matches Python's `copy_formals` which only copies `labels` plural)

### Step 2: Fix `CreateIsolate` public action labeling

**File:** `isolate/create.go` (lines 290-301)

Change from:
```go
lb.SetLabels([]string{name})
```
To:
```go
lb.SetLabel(name)
```

This is the **primary fix** for the divergence. It matches Python line 1920: `mod.actions[name].label = name`

### Step 3: Fix `BuildEnvAction` branch labeling

**File:** `actions/action.go`

Line 2064: change `seq.Labels = []string{lbl}` to `seq.Label = lbl`
Line 2073: change `env.Labels = []string{label}` to `env.Label = label`

Matches Python `ivy_actions.py` lines 1836 and 1842 which both set `.label` (singular).

### Step 4: Fix `temporal.EnvAction` branch labeling

**File:** `temporal/temporal.go` (line 391)

Change `ract.SetLabels([]string{name})` to `ract.SetLabel(name)`

Matches Python `ivy_temporal.py` line 238: `ract.label = name[4:] if name.startswith('ext:') else name`

### Step 5: Fix `vmt.addLabel`

**File:** `vmt/vmt.go` (lines 731-733)

Change `ab.SetLabels([]string{name})` to use `SetLabel(name)` instead. This corresponds to Python's `add_label()` method which sets `.label` (singular).

## Files to modify

1. `module/action.go` -- add `Label string` field, `SetLabel`, `GetLabel`
2. `isolate/create.go` -- fix public action labeling (primary fix)
3. `actions/action.go` -- fix `BuildEnvAction` branch/env labeling
4. `temporal/temporal.go` -- fix `EnvAction` branch labeling
5. `vmt/vmt.go` -- fix `addLabel`

## Verification

1. Run `go build ./...` to confirm compilation
2. Run `TestOrdLive` to verify the `labels:["cf_live"]` match:
   ```
   cd /Users/jaten/ivy/goivy && go test -run TestOrdLive -v -count=1
   ```
3. Run full test suite to check for regressions
