# Add Matching xtracer Landmark Traces to Python's create_isolate

**Created**: 2026-04-01 (updated 2026-04-01 17:40)

## Context

The golden trace diverges at line 156567. Go emits `check.CreateIsolate after_isolate_component` while Python emits `ast.LF.clone PRESERVE origid=207 counter=1822`. The block ordering already matches. The problem is **19 xtracer landmark traces** in Go's `create.go` that Python's `create_isolate` does not emit. These extra Go traces misalign all subsequent trace positions.

## Change

**File**: `~/ivy/pyivy/ivy/ivy/ivy_isolate.py` — function `create_isolate` (lines ~1858–1954)

Add matching `xtracer.trace()` calls to Python at the same logical positions as Go. All should use `if __debug__: xtracer.trace(...)` pattern.

### Trace insertions with precise locations

All use `if __debug__: xtracer.trace(...)` pattern.

**Format alignment rules** (Go → Python):
- `%q` (Go quoted string `""`) → `'"%s"'` in Python (NOT `%r` which uses single quotes)
- `%v` on boolean (Go `true`/`false`) → `str(val).lower()` in Python (NOT `%s` which gives `True`/`False`)
- `%d`, `%s` → same

---

**1. Line 1873** — after `mixed = ia.apply_mixin(...)`, before `mod.actions[mixed_name] = mixed`:
```python
                    if __debug__: xtracer.trace("isolate.create_no_iso mixer=%s mixee=%s" % (mixin.mixer(), mixin.mixee()))
```

**2. Line 1895** — after end of `if iso: ... else: ...` block, before the `# Create one big external action` comment:
```python
        if __debug__: xtracer.trace("check.CreateIsolate after_isolate_component")
```

**3. Line 1899** — before `for name in mod.public_actions:`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate before_label_public n_public=%d" % len(mod.public_actions))
```

**4. Line 1901** — after label loop, before `ext = kwargs[...`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate after_label_public")
```

**5. Line 1901** — after computing `ext`, before `if ext is not None:`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate before_ext_action ext=\"%s\"" % (ext if ext is not None else ""))
```

**6. Line 1908** — after the ext action if-block:
```python
        if __debug__: xtracer.trace("check.CreateIsolate after_ext_action")
```

**7. Line 1910** — before `slv.check_compat()`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate before_check_compat")
```

**8. Line 1912** — after `slv.check_compat()`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate after_check_compat")
```

**9. Line 1914** — before `mod.update_conjs()`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate before_update_conjs n_conjs=%d" % len(mod.labeled_conjs))
```

**10. Line 1916** — after `mod.update_conjs()`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate after_update_conjs")
```

**11. Line 1918** — before `cone = get_mod_cone(mod)`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate before_cone_of_influence cone=%s" % str(cone_of_influence.get()).lower())
```

**12. Line 1937** — after the cone/pedantic if-else block:
```python
        if __debug__: xtracer.trace("check.CreateIsolate after_cone_of_influence")
```

**13. Line 1938** — before `fix_initializers(mod,after_inits)`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate before_fix_initializers n_afterInits=%d" % len(after_inits))
```

**14. Line 1940** — after `fix_initializers(...)`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate after_fix_initializers")
```

**15. Line 1941** — before `mod.canonize_types()`. Compute sort_refs count to match Go:
```python
        _n_sort_refs = len(list(ivy_logic.sort_refinement()))
        if __debug__: xtracer.trace("check.CreateIsolate before_canonize_types n_sortRefs=%d" % _n_sort_refs)
```
Note: `ivy_logic` is imported at ivy_isolate.py line 5. `ivy_logic.sort_refinement()` (line 1471 of ivy_logic.py) is the Python equivalent of Go's `module.ComputeSortRefinements(mod.Sig)`. Python's `canonize_types()` recomputes it internally via `with self.sig:`, but we can query it here since the module sig context is already active.

**16. Line 1942** — after `mod.canonize_types()`:
```python
        if __debug__: xtracer.trace("check.CreateIsolate after_canonize_types")
```

**17. Line 1943** — before the bracket actions block:
```python
        if __debug__: xtracer.trace("check.CreateIsolate before_bracket_actions n_brackets=%d" % len(brackets))
```

**18. Line 1945** — inside bracket loop, before `bracket_action(...)`:
```python
                if __debug__: xtracer.trace("check.CreateIsolate bracket_action actname=%s n_before=%d n_after=%d" % (actname, len(before), len(after)))
```

**19. Line 1947** — after bracket actions block (after the if):
```python
        if __debug__: xtracer.trace("check.CreateIsolate after_bracket_actions")`
```

## Verification

```bash
cd ~/ivy/goivy && make golden
```

Check that the trace divergence point advances past line 156567.
