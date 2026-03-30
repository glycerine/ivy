# Plan: Full-Content Data Tracing for Symbol Collection Conformance

**Created**: 2026-03-30 15:30

## Context

The allSyms collection has multiple stages (formulas, formals, natives, normalize, action refs, proof defs). When a divergence appears in the final `allSyms_pre_follow` dump, we can't tell WHICH stage introduced it. Similarly, `allSyms2` and `interfSyms` currently have count-only traces in Go (Python has comma-separated lists). We need per-symbol full-content dumps at every stage so the golden test pinpoints the exact stage and exact symbol where Go/Python first diverge.

## Design

### Trace Format

Use the proven **per-symbol** format (one trace line per symbol, sorted, with count at the end) — the same pattern as the existing `allSyms_pre_follow.sym` traces. This enables the golden test to stop at the exact first differing symbol.

```
isolate.<label>.sym <display_name>
isolate.<label>.sym <display_name>
...
isolate.<label> n=<count>
```

### Go Helper Function

Extract the existing dump-allSyms pattern into a reusable helper to avoid code duplication:

```go
// traceSymSet dumps the full sorted contents of a symbol set.
// Matches Python: for x in sorted(s, key=lambda x: str(x)): xtracer.trace("label.sym %s" % str(x))
func traceSymSet(label string, syms map[lg.NodeKey]lg.Expr) {
    displayNames := make([]string, 0, len(syms))
    for _, v := range syms {
        if c, ok := v.(*lg.Const); ok {
            displayNames = append(displayNames, actions.ConstSymDisplay(c))
        }
    }
    sort.Strings(displayNames)
    for _, s := range displayNames {
        xtracer.Trace("%s.sym %s", label, s)
    }
    xtracer.Trace("%s n=%d", label, len(syms))
}
```

### Python Helper Function

```python
def _trace_sym_set(label, syms):
    """Dump full sorted contents of a symbol set for golden test comparison."""
    if __debug__:
        for x in sorted(syms, key=lambda x: str(x)):
            xtracer.trace("%s.sym %s" % (label, str(x)))
        xtracer.trace("%s n=%d" % (label, len(syms)))
```

## Checkpoint Plan

### Group A: allSyms Collection Stages

| Checkpoint | Label | Go location | Python location |
|---|---|---|---|
| A1 | `isolate.allSyms_post_normalize` | After normalization (line ~941) | After `all_syms = set(map(normalize,...))` (line ~1228) |
| A2 | `isolate.allSyms_post_action_refs` | After action refs loop (line ~946) | After `action.get_references` loop (line ~1230) |
| A3 | `isolate.allSyms_pre_follow` | Existing (line ~993) | Existing (line ~1246) |

**A1** is the first Go/Python comparison point — it covers formulas+formals+natives+normalize combined. If A1 diverges, Go has 3 sub-stage dumps (formulas-only, +formals, +natives) to narrow it down further. Python doesn't need sub-stage dumps since it combines them in one `used_symbols_asts` call.

| Sub-checkpoint | Label | Go only |
|---|---|---|
| A0a | `isolate.allSyms_post_formulas` | After formula collection |
| A0b | `isolate.allSyms_post_formals` | After formal params |
| A0c | `isolate.allSyms_post_natives` | After natives |

### Group B: allSyms2 Traces

| Checkpoint | Label | Go location | Python location |
|---|---|---|---|
| B1 | `isolate.allSyms2` | Replace count-only trace (line ~1279) | Replace comma-list trace (line ~1347) |

Both Go and Python change to per-symbol format.

### Group C: interfSyms Traces

| Checkpoint | Label | Go location | Python location |
|---|---|---|---|
| C1 | `isolate.interfSyms_pre_follow` | Replace count-only (line ~1325) | Replace comma-list (line ~1377) |
| C2 | `isolate.interfSyms_after_follow` | Replace count-only (line ~1327) | Replace comma-list (line ~1379) |

Both change to per-symbol format.

## Files to Modify

### Go: `isolate/isolate.go`

1. **Add `traceSymSet` helper** (near existing trace code)

2. **Add sub-stage dumps** (Go-only, for debugging). All calls guarded by `if xtracer.Enabled {}` so the compiler eliminates them when disabled:
   ```go
   // After formula collection (line ~917):
   if xtracer.Enabled { traceSymSet("isolate.allSyms_post_formulas", allSyms) }

   // After formal params (line ~926):
   if xtracer.Enabled { traceSymSet("isolate.allSyms_post_formals", allSyms) }

   // After natives (line ~935):
   if xtracer.Enabled { traceSymSet("isolate.allSyms_post_natives", allSyms) }
   ```

3. **Add matched checkpoints**:
   ```go
   // After normalize (line ~941):
   if xtracer.Enabled { traceSymSet("isolate.allSyms_post_normalize", allSyms) }

   // After action refs (line ~946):
   if xtracer.Enabled { traceSymSet("isolate.allSyms_post_action_refs", allSyms) }
   ```

4. **Refactor existing allSyms_pre_follow** to use `traceSymSet`:
   ```go
   if xtracer.Enabled { traceSymSet("isolate.allSyms_pre_follow", allSyms) }
   ```
   (Replaces the inline block at lines 984-997)

5. **Upgrade allSyms2 trace** (line ~1279):
   ```go
   if xtracer.Enabled { traceSymSet("isolate.allSyms2", allSyms2) }
   ```

6. **Upgrade interfSyms traces** (lines ~1325, ~1327):
   ```go
   if xtracer.Enabled { traceSymSet("isolate.interfSyms_pre_follow", interfSyms) }
   // ... follow definitions ...
   if xtracer.Enabled { traceSymSet("isolate.interfSyms_after_follow", interfSyms) }
   ```

### Python: `ivy_isolate.py`

1. **Add `_trace_sym_set` helper** (near existing trace code)

2. **Add matched checkpoints**:
   ```python
   # After line 1228 (all_syms = set(map(normalize,...))):
   _trace_sym_set("isolate.allSyms_post_normalize", all_syms)

   # After line 1230 (action.get_references loop):
   _trace_sym_set("isolate.allSyms_post_action_refs", all_syms)
   ```

3. **Refactor existing allSyms_pre_follow** to use `_trace_sym_set`:
   ```python
   _trace_sym_set("isolate.allSyms_pre_follow", all_syms)
   ```
   (Replaces lines 1245-1248)

4. **Upgrade allSyms2 trace** (line ~1347):
   ```python
   _trace_sym_set("isolate.allSyms2", all_syms)
   ```

5. **Upgrade interfSyms traces** (lines ~1377, ~1379):
   ```python
   _trace_sym_set("isolate.interfSyms_pre_follow", interf_syms)
   # ... follow definitions ...
   _trace_sym_set("isolate.interfSyms_after_follow", interf_syms)
   ```

## Trace Output Structure (Example)

For the `cf_live` isolate, each checkpoint produces ~159 symbol traces. The golden test sees:

```
isolate.allSyms_post_normalize.sym 0:index
isolate.allSyms_post_normalize.sym 0:lclock
...
isolate.allSyms_post_normalize n=155
isolate.allSyms_post_action_refs.sym 0:index
isolate.allSyms_post_action_refs.sym 0:lclock
...
isolate.allSyms_post_action_refs n=159
isolate.allSyms_pre_follow.sym 0:index     ← existing traces
...
```

When a divergence occurs, the golden test stops at the FIRST differing line, immediately identifying both the stage and the specific symbol.

## Verification

1. `go build ./...` — ensure compilation
2. `cd ~/goivy && make golden` — run golden test
3. Check `~/goivy/log.red` — if the new intermediate checkpoints match, the divergence narrows to the remaining `allSyms_pre_follow` issue (currently `cfabric.rd_fair`)
4. The sub-stage Go-only dumps (A0a-A0c) help debug formula/formal/native issues even when Python can't split them
