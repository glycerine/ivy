# Audit-and-fix all Python canon fall-throughs vs. Go's explicit Canon()

Creation date/time: 2026-04-21 (UTC)

## Context

After the `LabelForTrace()` fix advanced the test by ~40K lines, a NEW
divergence surfaced at index 323279: Python's `TemporalModels.canon()`
falls through to the default empty `(temporalModels)` while Go emits
`(temporalModels model:… fmla:…)`. This is one of many such gaps.

The user's directive: **audit and fix them all now** — add the missing
Python `canon()` methods (and the one missing Go `Canon()` method) up
front so `make golden-2hr` doesn't have to surface each one across
multiple 2-hour runs. Output must be **deterministic** (sort map keys
before emission) so divergences reflect real data differences.

## Audit results

I walked every Go type with `Canon()` (from `ast/ast.go`, `ast/formula.go`,
`ast/sort.go`, `ast/tactic.go`, `ast/canon_decl.go`, `temporal/temporal.go`)
and cross-referenced every registration in `canon_ast.py::install()` plus
the conditional decls loop at lines 521-559.

### Gaps — Python `canon()` missing while Go emits structural fields

**Term/formula types** (`ivy_ast.py`):

| Class | Python file:line | Go file:line | Go format |
|---|---|---|---|
| `Dot` | ivy_ast.py:1525 | ast.go:701 | `(dot{lf} left:{} right:{})` |
| `Bracket` | ivy_ast.py:1535 | ast.go:723 | `(bracket{lf} left:{} right:{})` |
| `KeyArg` | ivy_ast.py:1968 | ast.go:912 | `(keyArg app:{})` (no lineno) |
| `DebugItem` | ivy_ast.py:1490 | ast.go:928 | `(debugItem{lf} name:{} value:{})` |

**Def types — empty bodies** (`ivy_ast.py`, all emit `(defName{lf})`):

`AttributeDef` (1341 / canon_decl.go:303), `DelegateDef` (1257 / 283),
`ImportDef` (1221 / 279), `ExportDef` (1207 / 275),
`Instantiation` (958 / 307), `MixinAfterDef` (? / 259),
`MixinBeforeDef` (1142 / 251), `MixinImplementDef` (1147 / 255),
`NativeCode` (1287 / 287), `NativeType` (1301 / 291),
`NativeExpr` (1306 / 295), `NativeDef` (1325 / 299),
`PlaceList` (1443 / 320), `ScenarioTransition` (1459 / 324),
`StateDef` (1428 / 311).

**Def types — have extra fields**:

| Class | Python file:line | Go format |
|---|---|---|
| `VariantDef` | ivy_ast.py:1379 | `(variantDef{lf} name:{} vSort:{})` |
| `ScenarioDef` | ivy_ast.py:1463 | `(scenarioDef{lf} elems:{})` |
| `ScenarioBeforeMixin` | ivy_ast.py:1451 | `(scenarioBeforeMixin{lf} mixer:{} def:{})` |
| `ScenarioAfterMixin` | ivy_ast.py:1455 | `(scenarioAfterMixin{lf} mixer:{} def:{})` |
| `PrivateDef` | ivy_ast.py:1235 | `(privateDef{lf} elems:{})` |
| `ImplementTypeDef` | ivy_ast.py:1277 | `(implementTypeDef{lf} elems:{})` |

**Temporal types**:

| Class | Python file:line | Go format |
|---|---|---|
| `TemporalModels` | ivy_ast.py:1972 | ast.go:1683 → `(temporalModels{lf} model:{} fmla:{})` |
| `NormalProgram` | ivy_temporal.py:172 | temporal.go — *no `Canon()` on Go side either* |

### Gap — Go `Canon()` missing on `NormalProgram`

Go's `NormalProgram` (temporal/temporal.go:206) has no explicit `Canon()`;
it falls back to the embedded `ast.Base.Canon()` producing
`(base hasLoc:false loc:)`. Both sides need a structural canon for
`NormalProgram` to get real visibility.

### Explicit non-gaps (already covered, do NOT touch)

- Every Decl type: registered via `_make_decl_canon` loop at canon_ast.py:521-559 or individually 292-301.
- All formulas/binders/tactics/sorts: explicit entries.
- Action types: explicit or covered by `_action_generic_canon` loop (lines 638-646).
- `ActionTerm`, `ActionTermBinding`: registered at canon_ast.py:729, 735.
- Go-only types (`SomeAssignAction`, `CompiledNode`, `NamedSpace`,
  `ProductSpace`, `SumSpace`, `UninterpretedSortAST`): no Python equivalent
  so they never appear in Python canon — no divergence possible.
- Python-only types (`MixinDef` abstract, `ScenarioMixin` abstract,
  `TacticWithMatch`): no Go equivalent — never appear in Go canon.

## Plan

### Change 1 — Go: add `NormalProgram.Canon()` with sorted `Postconds`

File: `/Users/jaten/ivy/goivy/temporal/temporal.go`, alongside
`ActionTerm.Canon()` (line 92) / `ActionTermBinding.Canon()` (line 150).

```go
import "sort"  // add if not already imported

func (np *NormalProgram) Canon() iu.Canonical {
    var lf string
    if np.HasLoc {
        lf = fmt.Sprintf(" lineno:%d", np.Loc.Line)
    }
    bindings := nodeSliceCanonTemp(np.Bindings)   // []*ActionTermBinding
    invars   := nodeSliceCanonTemp(np.Invars)     // []*ast.LabeledFormula
    asms     := nodeSliceCanonTemp(np.Asms)       // []*ast.LabeledFormula
    calls    := stringSliceCanon(np.Calls)        // []string
    init := "nil"
    if np.Init != nil { init = string(np.Init.Canon()) }
    postconds := postcondsHashCanon(np.Postconds) // map — sorted by key
    return iu.Canonical(fmt.Sprintf(
        "(normalProgram%s bindings:%s init:%s invars:%s asms:%s calls:%s postconds:%s)",
        lf, bindings, init, invars, asms, calls, postconds))
}

// postcondsHashCanon emits "(hash "k":[...] ...)" deterministically.
func postcondsHashCanon(m map[string][]*ast.LabeledFormula) string {
    if len(m) == 0 { return "(hash)" }
    keys := make([]string, 0, len(m))
    for k := range m { keys = append(keys, k) }
    sort.Strings(keys)
    parts := make([]string, 0, len(keys))
    for _, k := range keys {
        parts = append(parts, fmt.Sprintf("%q:%s", k, nodeSliceCanonTemp(m[k])))
    }
    return "(hash " + strings.Join(parts, " ") + ")"
}

// nodeSliceCanonTemp is analogous to constSliceCanon (already at line 161)
// but for any slice whose element type implements ast.Node. Returns "[]"
// if empty, else "[elem1 elem2 ...]" with each Canon() inlined.
func nodeSliceCanonTemp[T ast.Node](xs []T) string {
    if len(xs) == 0 { return "[]" }
    parts := make([]string, len(xs))
    for i, x := range xs {
        parts[i] = string(x.Canon())
    }
    return "[" + strings.Join(parts, " ") + "]"
}
```

(If generics are undesirable in this package, three non-generic helpers
— `bindingSliceCanon`, `lfSliceCanon` — work equally well. Prefer reusing
any existing helper; add minimal code.)

### Change 2 — Python: add canon() registrations in `canon_ast.py`

File: `/Users/jaten/ivy/pyivy/ivy/ivy/canon_ast.py`, inside `install()`.

**Imports at top of file:**
```python
from . import ivy_temporal as temp   # already used deeper in install(); move to top
```

**Add a small helper near the top of `install()`:**
```python
def _make_empty_def_canon(name):
    def _canon(self):
        return '({}{})'.format(name, lineno_fields(self))
    return _canon
```

**Term/formula canons** (add in the "core" block near existing `_dot`/`_bracket`
placeholder around line 89):
```python
def _dot_canon(self):
    return '(dot{} left:{} right:{})'.format(
        lineno_fields(self), node_canon(self.args[0]), node_canon(self.args[1]))
ast.Dot.canon = _dot_canon

def _bracket_canon(self):
    return '(bracket{} left:{} right:{})'.format(
        lineno_fields(self), node_canon(self.args[0]), node_canon(self.args[1]))
ast.Bracket.canon = _bracket_canon

def _keyarg_canon(self):
    # Go: "(keyArg app:{})" — NO lineno_fields (Go emits it without).
    return '(keyArg app:{})'.format(ast.App.canon(self))
ast.KeyArg.canon = _keyarg_canon

def _debugitem_canon(self):
    return '(debugItem{} name:{} value:{})'.format(
        lineno_fields(self),
        node_canon(self.args[0]),
        node_canon(self.args[1]))
ast.DebugItem.canon = _debugitem_canon
```

**Empty-body Def canons** (register in a loop, similar to the Decl loop
at line 521-559):
```python
for name, cls_name in [
    ('attributeDef', 'AttributeDef'),
    ('delegateDef', 'DelegateDef'),
    ('importDef', 'ImportDef'),
    ('exportDef', 'ExportDef'),
    ('instantiation', 'Instantiation'),
    ('mixinAfterDef', 'MixinAfterDef'),
    ('mixinBeforeDef', 'MixinBeforeDef'),
    ('mixinImplementDef', 'MixinImplementDef'),
    ('nativeCode', 'NativeCode'),
    ('nativeType', 'NativeType'),
    ('nativeExpr', 'NativeExpr'),
    ('nativeDef', 'NativeDef'),
    ('placeList', 'PlaceList'),
    ('scenarioTransition', 'ScenarioTransition'),
    ('stateDef', 'StateDef'),
]:
    cls = getattr(ast, cls_name, None)
    if cls is not None:
        cls.canon = _make_empty_def_canon(name)
```

**Def canons with extra fields:**
```python
def _variantdef_canon(self):
    # Go: "(variantDef{lf} name:{} vSort:{})"
    # Python VariantDef stores name/vsort in self.args or attrs — verify
    # at implementation time (read ivy_ast.py:1379).
    name = self.args[0] if len(self.args) > 0 else None
    vsort = self.args[1] if len(self.args) > 1 else None
    return '(variantDef{} name:{} vSort:{})'.format(
        lineno_fields(self), node_canon(name), node_canon(vsort))
ast.VariantDef.canon = _variantdef_canon

def _scenariodef_canon(self):
    return '(scenarioDef{} elems:{})'.format(
        lineno_fields(self), slice_canon(list(self.args)))
ast.ScenarioDef.canon = _scenariodef_canon

def _scenariobeforemixin_canon(self):
    # Go fields: mixer, def — verify attribute names in ivy_ast.py:1451
    mixer = self.args[0] if len(self.args) > 0 else None
    defn  = self.args[1] if len(self.args) > 1 else None
    return '(scenarioBeforeMixin{} mixer:{} def:{})'.format(
        lineno_fields(self), node_canon(mixer), node_canon(defn))
ast.ScenarioBeforeMixin.canon = _scenariobeforemixin_canon

def _scenarioaftermixin_canon(self):
    mixer = self.args[0] if len(self.args) > 0 else None
    defn  = self.args[1] if len(self.args) > 1 else None
    return '(scenarioAfterMixin{} mixer:{} def:{})'.format(
        lineno_fields(self), node_canon(mixer), node_canon(defn))
ast.ScenarioAfterMixin.canon = _scenarioaftermixin_canon

def _privatedef_canon(self):
    return '(privateDef{} elems:{})'.format(
        lineno_fields(self), slice_canon(list(self.args)))
ast.PrivateDef.canon = _privatedef_canon

def _implementtypedef_canon(self):
    return '(implementTypeDef{} elems:{})'.format(
        lineno_fields(self), slice_canon(list(self.args)))
ast.ImplementTypeDef.canon = _implementtypedef_canon
```

**Temporal types** (verify/add near the existing `ActionTerm` block at
line 725):
```python
def _temporalmodels_canon(self):
    return '(temporalModels{} model:{} fmla:{})'.format(
        lineno_fields(self), node_canon(self.model), node_canon(self.fmla))
ast.TemporalModels.canon = _temporalmodels_canon

def _normalprogram_canon(self):
    # postconds is a dict — sort by key for determinism.
    postconds_sexp = '(hash)'
    if hasattr(self, 'postconds') and self.postconds:
        parts = []
        for k in sorted(self.postconds.keys()):
            parts.append('{}:{}'.format(
                string_canon(k), slice_canon(list(self.postconds[k]))))
        postconds_sexp = '(hash ' + ' '.join(parts) + ')'
    return '(normalProgram{} bindings:{} init:{} invars:{} asms:{} calls:{} postconds:{})'.format(
        lineno_fields(self),
        slice_canon(list(self.bindings)),
        node_canon(self.init),
        slice_canon(list(self.invars)),
        slice_canon(list(self.asms)),
        string_slice_canon(list(self.calls)),
        postconds_sexp)
temp.NormalProgram.canon = _normalprogram_canon
```

### Implementation discipline

- For each Python class touched, read its `__init__` / `@property args`
  first (to confirm field names / ordering). The agent-provided table uses
  `self.args[0]` / `self.args[1]` as a best guess — verify these match the
  Go struct's `Left`/`Right`/etc. When in doubt, use the attribute name
  directly (e.g., `self.model`, `self.mixer`, `self.name`) rather than
  `self.args[i]`.
- Do NOT change Python's `_ast_canon` default — other classes still rely
  on the empty-form fallback.
- Do NOT remove/relax Go's existing `Canon()` methods — visibility is the
  goal.

### Determinism

- Only one map appears in the new canons: Go `NormalProgram.Postconds` and
  the Python mirror. Both MUST `sort` keys before emission.
- No other collections require sorting (all slices preserve insertion
  order on both sides).

## Critical paths

- `/Users/jaten/ivy/goivy/temporal/temporal.go` — add `NormalProgram.Canon()` and `postcondsHashCanon`.
- `/Users/jaten/ivy/pyivy/ivy/ivy/canon_ast.py` — add the ~27 new canon registrations.
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_temporal.py:172` — reference only; verify `.bindings/.init/.invars/.asms/.calls/.postconds` attrs exist.
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_ast.py` — reference only for each class's field names.
- `/Users/jaten/ivy/goivy/ast/ast.go`, `ast/canon_decl.go` — reference only for matching Go canon format.

## Verification

1. `cd /Users/jaten/ivy/goivy && go build ./...` — clean build.
2. `cd /Users/jaten/ivy/goivy && make golden-2hr` — background run; monitor
   for divergence / PASS. Expected: divergence at 323279 is gone; test
   either PASSes or surfaces a genuinely new, distinct divergence
   (one that's rooted in real data mismatch, not canon-format gap).
3. Quick Python sanity: `cd /Users/jaten/ivy/pyivy && python -c "from ivy.ivy.canon_ast import install; install(); print('ok')"` — confirms imports / no typos.

## Explicitly out of scope

- Touching Decl types (all covered).
- Touching action canons (all either explicitly registered or covered by
  the generic `elems:` loop).
- Adding canons for Go-only or Python-only types.
- Retroactively shrinking Go's canon output to match Python's sparser form
  (user rejected that direction — visibility is the goal).
