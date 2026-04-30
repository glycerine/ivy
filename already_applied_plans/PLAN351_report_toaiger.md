# Re-port ToAiger to match Python to_aiger

Created: 2026-04-30 ~06:55 UTC

## Context

Two bugs found in `mc/toaiger.go:ToAiger` so far (AddErrFlag using `Not` instead of `DualFormula`, now a missing ProofChecker). Rather than patching one at a time, this plan catalogs **all divergences** between Go's `ToAiger` (mc/toaiger.go:30-554) and Python's `to_aiger` (ivy_mc.py:1118-1427) and fixes them in one pass.

The first fix (DualFormula in AddErrFlag) is already applied and tested. This plan covers everything remaining.

## Divergence catalog

### D1. Missing ProofChecker creation and proof-tactic processing [XTRACE 212999]

**Python** (ivy_mc.py:1136-1150):
```python
pc = ivy_proof.ProofChecker(mod.labeled_axioms, mod.definitions, mod.schemata)
pmap = dict((lf.id, p) for lf, p in mod.proofs)
conjs = []
for lf in mod.labeled_conjs:
    if not checked(lf):
        continue
    if verbose: print(...)
    if lf.id in pmap:
        proof = pmap[lf.id]
        subgoals = pc.admit_proposition(lf, proof)
        conjs.extend(subgoals)
    else:
        conjs.append(lf)
```

**Go** (toaiger.go:90-95) — just copies all LabeledConjs directly:
```go
var conjs []*ast.LabeledFormula
for _, lf := range mod.LabeledConjs {
    conjs = append(conjs, lf)
}
```

**Fix**: Replace Go lines 90-95 with:
1. Create `proof.NewProofChecker(mod.Cfg.ProofCfg, mod, mod.LabeledAxioms, mod.Definitions, mod.Schemata, mod.Cfg.AstCfg)`
2. Build `pmap` from `mod.Proofs` (a `[]module.ProofEntry` where `.Formula.LabelName()` → ID)
3. Loop `mod.LabeledConjs` with `checked` filter
4. If proof exists in pmap, call `pc.AdmitProposition(lf, proof)`; extend conjs with subgoals
5. Otherwise append lf directly

**Existing functions**: `proof.NewProofChecker` (proof/checker.go:39), `pc.AdmitProposition` (proof/checker.go:544), `mc.Checked` (mc/phase7.go:269 — needs adaptation for LabeledFormula)

### D2. Skolemization uses wrong variable collection

**Python** (ivy_mc.py:1154-1157):
```python
skolemizer = lambda v: ilu.var_to_skolem('__', il.Variable(v.rep, v.sort))
vs = ilu.used_variables_in_order_ast(invariant)
sksubs = dict((v.rep, skolemizer(v)) for v in vs)
invariant = ilu.substitute_ast(invariant, sksubs)
```

**Go** (toaiger.go:109-117) uses `lu.FreeVariablesList` and creates Consts directly with `lg.NewConst("__"+v.Name, v.VSort)`.

**Fix**: Use `module.UsedVariablesOrdered` (wrapping in Clauses like `DualFormula` does at module/skolem.go:57) and `module.VarToSkolem("__", v)` (module/skolem.go:41) to match Python exactly. The conditional `if len(freeVars) > 0` should also be removed — Python applies substitution unconditionally.

### D3. `indhyps` uses bare ForAll instead of CloseFormula

**Python** (ivy_mc.py:1185):
```python
indhyps = [il.close_formula(il.Implies(init_var, lf.formula))
           for lf in mod.labeled_conjs + mod.assumed_invariants]
```

**Go** (toaiger.go:174-185) always wraps in `&lg.ForAll{Body: ...}`, creating empty ForAll when no free vars.

**Fix**: Use `il.CloseFormula(...)` (ivylogic/util.go:256) instead of manually constructing ForAll.

### D4. `funs` collection includes wrong symbols

**Python** (ivy_mc.py:1200-1207) collects from def **RHS only** (`df.args[1]`), each fmla, and invariant, then filters for function sort.

**Go** (toaiger.go:196-208) uses `module.SymbolsClauses(trans)` which collects from ALL parts of definitions (LHS and RHS).

**Fix**: Match Python's approach — iterate `trans.Defs` collecting from `df.Rhs` only, iterate `trans.Fmlas` collecting symbols, add invariant symbols, then filter for function sort.

### D5. Missing `from_asserts` conjunction in MineConstants

**Python** (ivy_mc.py:1233-1237):
```python
from_asserts = il.And(*[il.Equals(x,x) for x in ilu.used_symbols_ast(il.And(*errconds))
                        if tr.is_skolem(x) and not il.is_function_sort(x.sort)])
invar_syms.update(ilu.used_symbols_ast(from_asserts))
sort_constants = mine_constants(mod, trans, il.And(invariant, from_asserts))
```

**Go** (toaiger.go:246-256) correctly updates `invarSyms` but passes only `invariant` to `MineConstants`, missing the `from_asserts` conjunction.

**Fix**: Build `fromAsserts` as `And(Equals(x,x) for each skolem non-function sym)`, then pass `And(invariant, fromAsserts)` to `MineConstants`. The `invarSyms` update (lines 247-254) already partially does this but the `MineConstants` call at line 256 is wrong.

### D6. Missing `error` clauses rename

**Python** (ivy_mc.py:1175): `error = ilu.rename_clauses(error, rn)` — renames the Pre (error/precondition) clauses alongside the transition relation.

**Go** (toaiger.go:132-160) processes `trans` but ignores `updWithAxioms.Pre` (the error/precondition clauses).

**Fix**: Add `module.RenameClauses(updWithAxioms.Pre, rn)`. The renamed `error` is unused downstream in Python too, but renaming it may produce XTRACE output.

### D7. `defsyms` built from wrong source

**Python** (ivy_mc.py:1172): `defsyms = set(x.defines() for x in bgt.defs)` — gets def symbols from **background theory defs**.

**Go** (toaiger.go:144-149): `defSymsByName` is built from `trans.Defs` — gets def symbols from **the merged trans clauses**.

**Fix**: Build `defSymsByName` from `bgt.Defs` (not `trans.Defs`) to match Python. `trans` at this point already includes `bgt.defs` merged in, so Go's version may have extra defs that shouldn't be in the rename map.

## Files to modify

- `mc/toaiger.go` — all fixes in `ToAiger` function

## Existing functions to reuse

- `proof.NewProofChecker` — proof/checker.go:39
- `proof.ProofChecker.AdmitProposition` — proof/checker.go:544
- `il.CloseFormula` — ivylogic/util.go:256
- `module.VarToSkolem` — module/skolem.go:41
- `module.UsedVariablesOrdered` — module/ops.go:998 (wrap expr in Clauses)
- `module.UsedSymbolsAST` — module/ops.go (already used)

## Verification

`cd ~/ivy/goivy && make test` then `make rfn` — the test should advance past XTRACE 212999.
