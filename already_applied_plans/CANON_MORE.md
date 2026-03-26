# Plan: Add sexp() to Python logic IR + Canon comparison across full pipeline

**Created:** 2026-03-26T23:30

## Context

The Canon() s-expression comparison system covers AST nodes but stops at the parser boundary. To verify the Go port produces correct compiled state, we need to extend Canon through the full pipeline: compilation → transition relations → Z3. The critical gap: Python's logic IR types (`ivy_logic.py`) have no `sexp()` methods. Go has `Sexp()` on all 27 logic types in `logic/sexp.go`.

## Phase A: Add sexp() to Python logic IR types

### New file: `~/pyivy/ivy/ivy/logic_sexp.py`

Monkey-patch `sexp()` onto all logic IR classes, matching Go's `logic/sexp.go` format exactly. Uses an `install()` function (same pattern as `canon_ast.py`).

**27 types to cover**, with exact Go format strings:

#### Sort types (6):
```python
def _uninterpreted_sort_sexp(self):
    return '(UninterpretedSort name:%s)' % self.name

def _boolean_sort_sexp(self):
    return '(BooleanSort)'

def _function_sort_sexp(self):
    parts = ' '.join(s.sexp() for s in self.sorts)
    return '(FunctionSort sorts:[%s])' % parts

def _enumerated_sort_sexp(self):
    # NOTE: comma-separated, not space-separated (matches Go)
    return '(EnumeratedSort name:%s ext:[%s])' % (self.name, ','.join(self.extension))

def _range_sort_sexp(self):
    return '(RangeSort name:%s lb:%s ub:%s)' % (self.name, self.lb, self.ub)

def _top_sort_sexp(self):
    return '(TopSort name:%s)' % self.name
```

#### Term types (3):
```python
def _var_sexp(self):
    return '(Variable name:%s sort:%s)' % (self.name, self.sort.sexp())

def _const_sexp(self):
    # Go Symbol has pre-computed sexp; Python must compute on the fly
    return '(Symbol name:%s sort:%s)' % (self.name, self.sort.sexp())

def _apply_sexp(self):
    parts = ' '.join(t.sexp() for t in self.terms)
    return '(Apply func:%s terms:[%s])' % (self.func.sexp(), parts)
```

#### Formula types (15):
```python
def _eq_sexp(self):     return '(Eq t1:%s t2:%s)' % (self.t1.sexp(), self.t2.sexp())
def _not_sexp(self):    return '(Not body:%s)' % self.body.sexp()
def _and_sexp(self):    return '(And terms:[%s])' % ' '.join(t.sexp() for t in self.terms)
def _or_sexp(self):     return '(Or terms:[%s])' % ' '.join(t.sexp() for t in self.terms)
def _implies_sexp(self): return '(Implies t1:%s t2:%s)' % (self.t1.sexp(), self.t2.sexp())
def _iff_sexp(self):    return '(Iff t1:%s t2:%s)' % (self.t1.sexp(), self.t2.sexp())
def _ite_sexp(self):    return '(Ite cond:%s then:%s else:%s)' % (self.cond.sexp(), self.t_then.sexp(), self.t_else.sexp())
def _globally_sexp(self):
    env = 'nil' if self.environ is None else self.environ
    return '(Globally environ:%s body:%s)' % (env, self.body.sexp())
def _eventually_sexp(self):
    env = 'nil' if self.environ is None else self.environ
    return '(Eventually environ:%s body:%s)' % (env, self.body.sexp())
def _when_operator_sexp(self):
    return '(WhenOperator name:%s t1:%s t2:%s)' % (self.name, self.t1.sexp(), self.t2.sexp())
def _cond_sexp(self):   return '(Cond t1:%s t2:%s)' % (self.t1.sexp(), self.t2.sexp())
```

#### Quantifier types — critical detail: `vars:` key, sorted variables
Go uses `vars:` (not `variables:`). Python `ForAll`/`Exists` store variables as `frozenset`, so must sort deterministically for sexp. Sort by `(name, sort.sexp())`:
```python
def _vars_sexp(variables):
    # Sort frozenset deterministically by (name, sort_sexp)
    sorted_vars = sorted(variables, key=lambda v: (v.name, v.sort.sexp()))
    return '[%s]' % ' '.join(v.sexp() for v in sorted_vars)

def _forall_sexp(self): return '(ForAll vars:%s body:%s)' % (_vars_sexp(self.variables), self.body.sexp())
def _exists_sexp(self): return '(Exists vars:%s body:%s)' % (_vars_sexp(self.variables), self.body.sexp())
def _lambda_sexp(self): return '(Lambda vars:%s body:%s)' % (_vars_sexp(self.variables), self.body.sexp())
def _named_binder_sexp(self):
    env = 'nil' if self.environ is None else self.environ
    return '(NamedBinder name:%s environ:%s vars:%s body:%s)' % (self.name, env, _vars_sexp(self.variables), self.body.sexp())
```

#### Definition types (2):
```python
def _definition_sexp(self): return '(Def lhs:%s rhs:%s)' % (self.lhs.sexp(), self.rhs.sexp())
def _definition_schema_sexp(self): return '(DefSchema lhs:%s rhs:%s)' % (self.lhs.sexp(), self.rhs.sexp())
```

**Go-side check**: Go `ForAll`/`Exists` store `Variables` as `[]*Variable` (ordered slice). If Go populates them from Python's frozenset during compilation, we need to verify Go also sorts them the same way. Check `logic/quantifier.go` or wherever ForAll is constructed.

### Clauses.sexp() — in same file or `ivy_logic_utils.py`
```python
def _clauses_sexp(self):
    fmlas = ' '.join(f.sexp() for f in self.fmlas)
    defs = ' '.join(d.sexp() for d in self.defs)
    return '(clauses fmlas:[%s] defs:[%s])' % (fmlas, defs)
```

### Install function
```python
def install():
    """Monkey-patch sexp() onto all logic IR types. Call once at startup."""
    from . import logic as lg
    from . import ivy_logic_utils as lut
    lg.UninterpretedSort.sexp = _uninterpreted_sort_sexp
    lg.Boolean.sexp = _boolean_sort_sexp
    # ... all 27 types ...
    lut.Clauses.sexp = _clauses_sexp
```

Call `install()` from wherever `canon_ast.install()` is called (likely `ivy_check.py` or test setup).

## Phase B: Module.canon_snapshot() on both sides

### Go side — already written: `module/canon.go`
- `CanonSnapshot(label)` emits xtracer traces for key module fields
- Uses `Clauses.Canon()` (already written: `clauseops/canon.go`)

### Python side — add to `ivy_module.py`
```python
def canon_snapshot(self, label):
    xtracer.trace("module.CanonSnapshot ENTER label=%s" % label)
    xtracer.trace("module.CanonSnapshot %s labeledAxioms=%s" % (label, _canon_lf_slice(self.labeled_axioms)))
    xtracer.trace("module.CanonSnapshot %s definitions=%s" % (label, _canon_lf_slice(self.definitions)))
    # ... same fields as Go ...
    xtracer.trace("module.CanonSnapshot EXIT label=%s" % label)
```

### Emit snapshots after each compilation pass

**Go** `compiler/ivy_compile.go` — after each pass:
```go
mod.CanonSnapshot("after-domain-setup")
mod.CanonSnapshot("after-conj-setup")
mod.CanonSnapshot("after-arg-setup")
```

**Python** `ivy_compiler.py` — after each `IvyDomainSetup`, `IvyConjectureSetup`, `IvyARGSetup`:
```python
mod.canon_snapshot("after-domain-setup")
mod.canon_snapshot("after-conj-setup")
mod.canon_snapshot("after-arg-setup")
```

## Phase C: Z3 sexpr choke point

### Both sides: instrument solver calls

**Python** `ivy_solver.py` or `z3_utils.py`:
```python
# Before solver.check():
for assertion in solver.assertions():
    normalized = normalize_z3_varnames(assertion.sexpr())
    xtracer.trace("z3.assert sexpr=%s" % normalized)
```

**Go** `z3bridge/translate.go` or `solver/solver.go`:
```go
// Before solver.Check():
for _, assertion := range solver.Assertions() {
    normalized := normalizeZ3VarNames(assertion.Sexpr())
    xtracer.Trace("z3.assert sexpr=%s", normalized)
}
```

### Variable name normalization
Z3 uses internal names like `!k!0`, `!k!1` etc. These may differ between Go (cgo) and Python bindings. Normalize by replacing `!k!N` patterns with deterministic names based on declaration order within each assertion.

```python
import re
def normalize_z3_varnames(sexpr):
    counter = [0]
    seen = {}
    def replace(m):
        name = m.group(0)
        if name not in seen:
            seen[name] = '!v!%d' % counter[0]
            counter[0] += 1
        return seen[name]
    return re.sub(r'!k!\d+', replace, sexpr)
```

## Phase D: Action/transition Canon

### Action update tuples
Python: `(modified_symbols, transition_clauses, precondition)`
Go: equivalent in actions package

```python
def _update_sexp(modified, clauses, precond):
    mod_str = '[%s]' % ' '.join(sorted(str(s) for s in modified))
    return '(update modified:%s clauses:%s precond:%s)' % (mod_str, clauses.sexp(), precond.sexp())
```

### History objects
```python
def _history_sexp(self):
    return '(history post:%s steps:%d)' % (self.post.sexp(), len(self.maps))
```

## Files to modify

| File | Changes |
|---|---|
| **New**: `~/pyivy/ivy/ivy/logic_sexp.py` | sexp() for all 27 logic IR types + Clauses + install() |
| `~/pyivy/ivy/ivy/ivy_module.py` | Add `canon_snapshot()` method |
| `~/pyivy/ivy/ivy/ivy_compiler.py` | Emit canon_snapshot after each pass |
| `~/pyivy/ivy/ivy/ivy_solver.py` | Emit Z3 sexpr via xtracer |
| **Exists**: `clauseops/canon.go` | Clauses.Canon() — already written |
| **Exists**: `module/canon.go` | Module.CanonSnapshot() — already written |
| `compiler/ivy_compile.go` | Emit CanonSnapshot after each pass |
| `z3bridge/translate.go` or `solver/solver.go` | Emit Z3 sexpr via xtracer |
| Sync `~/pyivy` → venv after Python changes | |

## Key risks

1. **ForAll/Exists variable ordering**: Python uses `frozenset`, Go uses ordered `[]*Variable`. Must sort identically. Go may need to sort its Variables slice too if not already deterministic.
2. **Symbol sexp caching**: Go pre-computes `Symbol.sexp` at creation time. Python must compute on-the-fly or cache similarly.
3. **Clauses.canon LabeledFormula**: Go `Module.CanonSnapshot` calls `lf.Canon()` on `*ast.LabeledFormula`. Python's LabeledFormula already has canon() from `canon_ast.py`. These are AST-level, not logic-level — they should already match.

## Verification

After each phase, build both sides and run golden test:
```bash
go build ./...
XTRACE_OFF=1 go test ./... -count=1 -timeout 120s
cd ~/goivy && make golden
```

Canon snapshots will immediately reveal compilation state divergences. Each divergence is a bug to fix in the Go port.
