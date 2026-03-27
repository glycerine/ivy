# Fix: DomainSetup.LastFact should be *ast.LabeledFormula, not lg.Expr

**Created:** 2026-03-27 (current session)

## Context

The `make golden` test diverges at line 88775:
```
go : XTRACE: ast.LF.__init__ id=1577 counter=1578
py : XTRACE: compiler.IvyDomainSetup.dispatch name=isolate
```

After processing a `proof` with `ComposeTactics`, Go creates an **extra** `LabeledFormula` that Python doesn't. This happens because Go's `DomainSetup.LastFact` is typed as `lg.Expr` (the extracted formula), while Python's `self.last_fact` is the full `LabeledFormula`. When `Proof()` needs to store into `Module.Proofs`, Go must wrap `LastFact` in a new LabeledFormula, but Python just reuses `self.last_fact` directly.

## Root Cause

**File:** `compiler/decl.go`

`DomainSetup.LastFact` is `lg.Expr` (line 24). Python's `self.last_fact` is a `LabeledFormula`. This type mismatch forces Go to create new LabeledFormulas in `Proof()` (line 1415) and `Named()` (line 1467), producing extra `ast.LF.__init__` traces that Python doesn't emit.

## Fix

### Step 1: Change `LastFact` type (`compiler/decl.go:24`)

```go
// Before:
LastFact lg.Expr

// After:
LastFact *ast.LabeledFormula
```

### Step 2: Update all setters of `LastFact`

All these currently extract `.Formula` from a `*ast.LabeledFormula` and store just the formula. Instead, store the full `*ast.LabeledFormula`:

1. **Property** (line ~549-551): `d.LastFact = clf` (remove the `if fmla, ok := clf.Formula.(lg.Expr)` wrapper)
2. **Definition** (line ~677-679): `d.LastFact = mlf`
3. **Derived** (line ~755-757): `d.LastFact = mlf`
4. **Theorem** (line ~1505): `d.LastFact = mlf` (not `compiled`)
5. **Conjecture** (line 564): `d.LastFact = nil` — no change needed (nil pointer)

### Step 3: Update all users of `LastFact`

1. **Proof** (line 1415): Replace `acfg.NewLabeledFormula(nil, d.LastFact)` with `d.LastFact` directly:
   ```go
   // Before:
   acfg := d.Compiler.Module.Cfg.AstCfg
   lastLF := acfg.NewLabeledFormula(nil, d.LastFact)
   d.Compiler.Module.Proofs = append(d.Compiler.Module.Proofs, module.ProofEntry{
       Formula: lastLF,
       Proof:   compiled,
   })

   // After:
   d.Compiler.Module.Proofs = append(d.Compiler.Module.Proofs, module.ProofEntry{
       Formula: d.LastFact,
       Proof:   compiled,
   })
   ```

2. **Named** (line 1438): Access formula via `d.LastFact.Formula`:
   ```go
   // Before:
   cond := il.DropUniversals(d.LastFact)

   // After:
   lastFormula, ok := d.LastFact.Formula.(lg.Expr)
   if !ok {
       return lg.NewIvyError(node, "named declaration without preceding property")
   }
   cond := il.DropUniversals(lastFormula)
   ```

3. **Named** (line 1467): Use `d.LastFact` directly instead of creating new LF:
   ```go
   // Before:
   acfg := d.Compiler.Module.Cfg.AstCfg
   lastLF := acfg.NewLabeledFormula(nil, d.LastFact)
   d.Compiler.Module.Named = append(d.Compiler.Module.Named, module.NamedEntry{
       Formula: lastLF,
       ...

   // After:
   d.Compiler.Module.Named = append(d.Compiler.Module.Named, module.NamedEntry{
       Formula: d.LastFact,
       ...
   ```

### Step 4: Check for any other references to `LastFact`

Grep for `LastFact` across the codebase to ensure no other code accesses this field as `lg.Expr`.

## Verification

1. `cd ~/goivy && make build` — must compile
2. `cd ~/goivy && make golden` — divergence at line 88775 should be resolved; check what the next divergence is (if any)
