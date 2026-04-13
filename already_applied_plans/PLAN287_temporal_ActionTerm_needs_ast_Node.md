# Fix: ActionTerm must implement ast.Node for Python-matching modPass behavior

**Created:** 2026-04-13 14:15 UTC

## Context

At xtrace line 330541 (golden 330541, test `TestOrdLive`), Go and Python diverge:

```
330540  go : l2s.modPass clone binding[0] ENTER name=ext:lclock.next
        py : l2s.modPass clone binding[0] ENTER name=ext:lclock.next

330541  go : transformAction ENTER type=Sequence
        py : ilu.replaceTemporalsRec ENTER type=ActionTerm HASH canon=(actionTerm)
```

**Root cause:** In Python, `ActionTerm` extends `ia.AST` (is an AST node). Python's `mod_pass` calls `transform(b.action)` directly, passing the ActionTerm to `replace_temporals_rec`, which recursively processes `ActionTerm.args = [stmt]` and clones.

In Go, `ActionTerm` is a plain struct -- NOT an `ast.Node`. Go's `modPass` works around this by extracting `b.Action.Stmt` and passing it to `transformAction` -- a Go-only function that doesn't exist in Python. This:
1. Skips the ActionTerm level entirely (trace divergence)
2. Produces `transformAction ENTER` traces instead of `ilu.replaceTemporalsRec ENTER` traces for all action nodes

Python's `replace_temporals_rec` handles the ENTIRE tree (formulas AND actions) uniformly via `.args`/`.clone()`. Go already has actions implementing `ast.Node` (via `action_expr.go`), so `replaceTemporalsRec` CAN process them -- we just need to make ActionTerm an ast.Node too and stop using `transformAction`.

## Plan

### Step 1: Make `ActionTerm` implement `ast.Node` with field-aware Canon/Sexp

**File:** `temporal/temporal.go`

1a. Rename existing `Clone(stmt actions.Action) *ActionTerm` to `CloneStmt(stmt actions.Action) *ActionTerm`

1b. Add `ast.Node` methods and field-aware `Canon()` to `ActionTerm`:

```go
// --- ast.Node interface for ActionTerm ---
// Python: ActionTerm(ia.AST) with args=[self.stmt]

func (at *ActionTerm) Args() []ast.Node {
    return []ast.Node{at.Stmt}
}

func (at *ActionTerm) Clone(args []ast.Node) ast.Node {
    return &ActionTerm{
        Inputs:  at.Inputs,
        Outputs: at.Outputs,
        Labels:  at.Labels,
        Stmt:    args[0].(actions.Action),
    }
}

func (at *ActionTerm) GetLineno() ast.Location     { return ast.Location{} }
func (at *ActionTerm) SetLineno(l ast.Location)     {}
func (at *ActionTerm) GetAstConfig() *ast.AstConfig { return nil }

func (at *ActionTerm) Canon() iu.Canonical {
    return iu.Canonical(fmt.Sprintf("(actionTerm inputs:%s outputs:%s labels:%s stmt:%s)",
        constSliceCanon(at.Inputs), constSliceCanon(at.Outputs),
        stringSliceCanon(at.Labels), at.Stmt.Canon()))
}
```

Currently Python's `ActionTerm.canon()` uses the base `_ast_canon` fallback which produces `(actionTerm)` with no fields. This is insufficient -- both sides need field-aware canon. The Go Canon above includes all four fields (inputs, outputs, labels, stmt) following the established pattern in `canon_ast.py` for action types.

**Field mapping (Python attr -> Go field -> canon key):**
- `self.inputs` -> `Inputs []*lg.Const` -> `inputs:`
- `self.outputs` -> `Outputs []*lg.Const` -> `outputs:`
- `self.labels` -> `Labels []string` -> `labels:`
- `self.stmt` -> `Stmt actions.Action` -> `stmt:`

`String()` already exists on ActionTerm.

1c. Add local canon helpers in `temporal/temporal.go`:

```go
// constSliceCanon returns canonical form for []*lg.Const.
func constSliceCanon(cs []*lg.Const) string {
    if len(cs) == 0 {
        return "[]"
    }
    parts := make([]string, len(cs))
    for i, c := range cs {
        parts[i] = string(c.Canon())
    }
    return "[" + strings.Join(parts, " ") + "]"
}

// stringSliceCanon returns canonical form for []string.
func stringSliceCanon(ss []string) string {
    if len(ss) == 0 {
        return "[]"
    }
    parts := make([]string, len(ss))
    for i, s := range ss {
        parts[i] = fmt.Sprintf("%q", s)
    }
    return "[" + strings.Join(parts, " ") + "]"
}
```

These match `ast.SliceCanon`/`ast.stringSliceCanon` patterns and Python's `slice_canon`/`string_slice_canon` from `canon.py`.

### Step 2: Add field-aware canon for ActionTerm in Python

**File:** `~/ivy/pyivy/ivy/ivy/canon_ast.py` (inside `install()`)

Add after the existing action type canons (around line 724, before `install_fragment()`):

```python
# --- Temporal types (ivy_temporal.py) ---
from . import ivy_temporal as itm

def _actionterm_canon(self):
    return '(actionTerm inputs:{} outputs:{} labels:{} stmt:{})'.format(
        slice_canon(self.inputs), slice_canon(self.outputs),
        string_slice_canon(self.labels), node_canon(self.stmt))
itm.ActionTerm.canon = _actionterm_canon
itm.ActionTerm.sexp = _actionterm_canon
```

This overrides the base `_ast_canon` fallback so both Go and Python produce the same field-aware representation.

### Step 3: Update callers of old `Clone` -> `CloneStmt`

These callers construct a new Stmt manually and need the old signature:

- `check/l2s_shared.go:510`: `b.Action.Clone(newStmt)` -> `b.Action.CloneStmt(newStmt)`
- `check/l2s_shared.go:556`: `b.Action.Clone(newStmt)` -> `b.Action.CloneStmt(newStmt)`
- `temporal/temporal.go:588`: `b.Action.Clone(newStmt)` -> `b.Action.CloneStmt(newStmt)`

### Step 4: Change `modPass` in `check/l2s.go` (lines 544-553)

**Bindings** (lines 544-548): Change from extracting Stmt + transformAction to calling transform on ActionTerm directly:

```go
// Before:
newStmt := transformAction(b.Action.Stmt, transform)
model.Bindings[i] = b.Clone(b.Action.Clone(newStmt))

// After (matches Python: model.bindings[i] = b.clone([transform(b.action)])):
newAction := transform(b.Action).(*temporal.ActionTerm)
model.Bindings[i] = b.Clone(newAction)
```

**Init** (lines 550-553): Change from transformAction to calling transform directly:

```go
// Before:
model.Init = transformAction(model.Init, transform)

// After (matches Python: model.init = transform(model.init)):
model.Init = transform(model.Init).(actions.Action)
```

### Step 5: Change `modPass` in `check/ranking.go` (lines 403-412)

Same changes as Step 4:
- Bindings: `transform(b.Action).(*temporal.ActionTerm)` + `b.Clone(newAction)`
- Init: `transform(model.Init).(actions.Action)`

### Step 6: Leave `transformAction` in place

`transformAction` (l2s.go:888-909) is tested in `l2s_test.go` and may be useful in the future. Leave it but it's no longer called from modPass.

## Files to modify

1. `temporal/temporal.go` -- ActionTerm: rename Clone->CloneStmt, add ast.Node methods + Canon + helpers, update line 588
2. `~/ivy/pyivy/ivy/ivy/canon_ast.py` -- add field-aware ActionTerm canon/sexp override
3. `check/l2s.go` -- modPass binding/init handling (lines 544-553)
4. `check/ranking.go` -- modPass binding/init handling (lines 403-412)
5. `check/l2s_shared.go` -- update Clone->CloneStmt calls (lines 510, 556)

## Imports needed

`temporal/temporal.go` will need to add:
- `iu "github.com/glycerine/ivy/goivy/ivyutils"` (for `iu.Canonical`; not currently imported)
- `"github.com/glycerine/ivy/goivy/ast"` already imported

## Why this works

- Go's actions already implement `ast.Node` (via `action_expr.go`). `replaceTemporalsRec` handles them correctly through `Args()/Clone()`.
- Making ActionTerm an ast.Node lets `replaceTemporalsRec` process it like Python does: enter ActionTerm -> recurse into `Args()=[Stmt]` -> recurse into Stmt's children -> clone back up.
- The trace will now show `ilu.replaceTemporalsRec ENTER type=ActionTerm HASH canon=(actionTerm ...)` matching Python.
- All action sub-trees (Sequence, AssumeAction, etc.) will also be processed by `replaceTemporalsRec` instead of `transformAction`, producing matching traces.

## Verification

1. `cd ~/ivy/goivy && go build ./...` -- verify compilation
2. Run the failing test: `go test -run TestOrdLive -count=1 ./parser/ -timeout 600s`
3. Check that line 330541 now matches between Go and Python
