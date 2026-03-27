# Fix FixConstructors trace ordering in DomainSetup
**Created:** 2026-03-27 (current session)

## Context

Golden test diverges at line 102471. Go emits `compiler.DomainSetup EXIT` while Python emits `compiler.FixConstructors ENTER`. The root cause is a trace ordering mismatch: in Python, `fix_constructors(mod)` is called **inside** the DomainSetup ENTER/EXIT bracket, but in Go it's called **after** the EXIT trace.

**Python** (`ivy_compiler.py:2479-2482`):
```python
xtracer.trace("compiler.DomainSetup ENTER")
IvyDomainSetup(mod)(decls)
fix_constructors(mod)                          # inside bracket
xtracer.trace("compiler.DomainSetup EXIT")
```

**Go** (`compiler/ivy_compile.go:122-131`):
```go
xtracer.Trace("compiler.DomainSetup ENTER")
domainInterp.ProcessDecls(decls)
xtracer.Trace("compiler.DomainSetup EXIT")     // EXIT too early
FixConstructors(mod)                            // outside bracket
mod.CanonSnapshot("after-domain-setup")
```

## Plan

**File:** `compiler/ivy_compile.go` (lines 122-131)

Move `FixConstructors(mod)` and `mod.CanonSnapshot(...)` before the `DomainSetup EXIT` trace, matching Python's ordering:

```go
xtracer.Trace("compiler.DomainSetup ENTER")
domainInterp := NewDomainSetup(c)
if err := domainInterp.ProcessDecls(decls); err != nil {
    return fmt.Errorf("domain setup: %w", err)
}
// fix_constructors: ensure constructor sorts are properly set
FixConstructors(mod)
mod.CanonSnapshot("after-domain-setup")
xtracer.Trace("compiler.DomainSetup EXIT")
```

## Verification

Run `cd ~/goivy && make golden` and confirm the test passes line 102471 (FixConstructors ENTER now matches).
