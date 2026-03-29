# Plan: Faithful Port of `apply_assert_proofs.recur` — Full Audit and Fix

**Created**: 2026-03-29 19:30

## Context

The Go `ApplyAssertProofsWithProver` in `compiler/phase6.go` contains a `recur` function that must be a mechanical port of Python's `apply_assert_proofs.recur` in `ivy_compiler.py:2210-2232`. The golden test diverges at the first clone inside `recur` because Go's implementation has multiple unfaithful divergences from Python's logic.

Module action bodies are confirmed identical going INTO `apply_assert_proofs` (the `before-apply-assert-proofs` CanonSnapshot matches). The divergence is created by `recur` itself.

## Side-by-Side Audit

### Python `recur` (ivy_compiler.py:2210-2232)

```python
def recur(self):                                          # P1
    if not isinstance(self,Action):                       # P2
        return self                                       # P3
    if isinstance(self,AssertAction):                     # P4 — catches Assert, Requires, Ensures, Subgoal
        if len(self.args) > 1:                            # P5 — "has proof" = args has 2 elements
            if option_verifying:                          # P6
                return apply_assert_proof(prover,self,self.args[1])  # P7
            return self.clone(self.args[:1])              # P8 — clone with proof stripped
        return self                                       # P9 — no proof, return unchanged
    if isinstance(self,WhileAction):                      # P10
        if len(self.args) > 2:                            # P11 — has invariants
            new_invars = []                               # P12
            for a in self.args[2:]:                       # P13
                r = recur(a)                              # P14
                if isinstance(r,Sequence):                # P15
                    new_invars.extend(r.args)             # P16
                else:                                     # P17
                    new_invars.append(r)                  # P18
            return self.clone(list(map(recur,self.args[0:2])) + new_invars)  # P19
    # NOTE: WhileAction WITHOUT invariants FALLS THROUGH to P20/P22
    if isinstance(self,LocalAction):                      # P20
        with ivy_logic.WithSymbols(self.args[0:-1]):      # P21
            return self.clone(list(map(recur,self.args))) # P22
    return self.clone(list(map(recur,self.args)))         # P23 — ALWAYS clones
```

### Go `recur` — Current Bugs Identified

**Bug 1: AssertAction "has proof" check uses `a.Proof != nil` instead of `len(ActionArgs()) > 1`**

Go (line 1929): `if a.Proof != nil`
Python (line P5): `if len(self.args) > 1`

These are equivalent because Go's `ActionArgs()` returns `[Formula, Proof]` when Proof is non-nil, and `[Formula]` when nil. So `a.Proof != nil` ↔ `len(ActionArgs()) > 1`. **Not a bug** — functionally equivalent.

**Bug 2: AssertAction non-verifying path creates wrong type for subclasses**

Go (line 1933): `return actions.NewAssertAction(a.Formula)` — always returns `*AssertAction`
Python (line P8): `return self.clone(self.args[:1])` — clones preserving the concrete type

When `self` is a `RequiresAction`, Python returns a `RequiresAction` (via `self.clone`). Go returns a plain `AssertAction`. Similarly for `EnsuresAction` and `SubgoalAction`.

Go lines 1942, 1951, 1960 attempt to fix this with type-specific constructors (`NewRequiresAction`, `NewEnsuresAction`, `NewSubgoalAction`), which IS functionally correct. **Not a bug** for type preservation, but see Bug 3.

**Bug 3: `apply_assert_proof` receives different info for subclasses**

Go passes `&a.AssertAction` (line 1940, 1949, 1958) to `applyAssertProofAction`, losing the concrete type. Python passes `self` which is the concrete subclass. In Python, `apply_assert_proof(prover, self, self.args[1])` receives the full `RequiresAction`/`EnsuresAction`/`SubgoalAction` object. Python's `apply_assert_proof` uses `type(self)` to set `sga.kind`.

Go uses `kindName` string instead, which was the explicit design choice. The `a.Name()` call on the concrete type should return the correct kind. **Functionally equivalent if Name() is correct.**

**Bug 4 (CRITICAL): WhileAction without invariants does NOT fall through to generic clone**

Python (lines P10-P19): When `isinstance(self, WhileAction)` is True but `len(self.args) > 2` is False (no invariants), execution falls through to P20 (LocalAction check) and then P23 (generic clone). The WhileAction gets cloned via the generic path.

Go (lines 1965-1993): When `act.(*actions.WhileAction)` matches but `len(w.Invariants) > 0` is False, the `if` block is skipped but execution continues to the LocalAction check and then the generic path. **This is actually correct in the current code** — it falls through. Not a bug.

**Bug 5 (CRITICAL, FIXED): Generic path had `changed` optimization**

Go previously had:
```go
if !changed { return act }
return act.ActionClone(newArgs)
```
Python (line P23): `return self.clone(list(map(recur, self.args)))` — ALWAYS clones.

This was already fixed in the previous edit to always clone. **Verify the fix is in place.**

**Bug 6: WhileAction invariant processing calls `recur` only on Action args**

Go (lines 1968-1982):
```go
for _, inv := range w.Invariants {
    var r actions.Action
    if subAct, ok := inv.(actions.Action); ok {
        r = recur(subAct)
    }
    if r == nil {
        newInvars = append(newInvars, inv)
        continue
    }
```
Python (lines P13-P18):
```python
for a in self.args[2:]:
    r = recur(a)
```

Python calls `recur(a)` on EVERY invariant arg. If the arg is not an Action, Python's `recur` returns it unchanged (line P2-P3). Go skips non-Action args entirely without calling `recur`. **Functionally equivalent** since recur on non-Action returns identity.

BUT: Go's `r == nil` check is wrong. If `inv` is not an Action, `r` remains nil, and Go appends `inv` unchanged. But if `inv` IS an Action and recur returns non-nil, Go checks `isinstance(r, Sequence)`. This matches Python. **OK.**

**Bug 7: WhileAction clone does not recurse cond**

Go (lines 1984-1989):
```go
newCond := w.Cond
newBody := w.Body
if bodyAct, ok := w.Body.(actions.Action); ok {
    newBody = recur(bodyAct)
}
```
Python (line P19): `self.clone(list(map(recur, self.args[0:2])) + new_invars)`

Python applies `map(recur, self.args[0:2])` which calls `recur` on BOTH `args[0]` (cond) AND `args[1]` (body). Go only recurses the body, not the cond.

In Python, `recur(args[0])` on the condition (a formula, not an Action) returns it unchanged (P2-P3). So this is **functionally equivalent** — the cond is a formula, not an action.

**Bug 8 (CRITICAL): LocalAction path only recurses Action children**

Go (lines 2004-2012):
```go
allArgs := la.ActionArgs()
newArgs := make([]lg.Expr, len(allArgs))
for i, arg := range allArgs {
    if subAct, ok := arg.(actions.Action); ok {
        newArgs[i] = recur(subAct)
    } else {
        newArgs[i] = arg
    }
}
```
Python (line P22): `self.clone(list(map(recur, self.args)))`

Python calls `recur` on ALL args including local symbol declarations (which are not Actions). For non-Actions, `recur` returns identity (P2-P3). Go skips non-Action args. **Functionally equivalent** — non-Action args are passed through unchanged in both cases.

**Bug 9: Generic path only recurses Action children**

Same pattern as Bug 8. Go checks `if subAct, ok := arg.(actions.Action)` before calling recur. Python calls `map(recur, self.args)` on everything. **Functionally equivalent** for the same reason.

### Summary of Real Bugs

Only **Bug 5** is a real behavioral divergence (the `changed` optimization). It was already fixed. Let me verify the fix is in the current code.

## Plan

### Step 1: Verify Bug 5 fix is in place

Read the current Go `recur` generic path and confirm it always clones (no `changed` check).

### Step 2: Add point-by-point Python mapping comments

Replace the Go `recur` function with a version that has clear comments mapping each block to the Python line numbers (P1-P23 from the audit above). This makes future audits trivial.

**File**: `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/phase6.go`

The rewritten `recur` should have this structure:

```go
recur = func(act actions.Action) actions.Action {
    // P1: def recur(self):
    if act == nil {
        return nil
    }
    // P2-P3: if not isinstance(self, Action): return self
    // [Go: all args to recur are actions.Action by type, so this is implicit]

    // P4-P9: if isinstance(self, AssertAction): ...
    // Python isinstance catches all subclasses: AssertAction, RequiresAction,
    // EnsuresAction, SubgoalAction. Go must check each concrete type.
    if aa := getAssertAction(act); aa != nil {
        // P5: if len(self.args) > 1:  (has proof)
        if aa.Proof != nil {
            // P6-P7: if option_verifying: return apply_assert_proof(prover, self, self.args[1])
            if getModVerifying(mod) {
                return applyAssertProofAction(mod, aa, actionName(act), prover)
            }
            // P8: return self.clone(self.args[:1])  — clone with proof stripped
            return cloneAssertWithoutProof(act)
        }
        // P9: return self
        return act
    }

    // P10-P19: if isinstance(self, WhileAction): ...
    if w, ok := act.(*actions.WhileAction); ok {
        // P11: if len(self.args) > 2:  (has invariants)
        if len(w.Invariants) > 0 {
            // P12-P18: process invariants
            ...
            // P19: return self.clone(list(map(recur, self.args[0:2])) + new_invars)
            ...
            return ...
        }
        // WhileAction without invariants falls through to P20/P23
    }

    // P20-P22: if isinstance(self, LocalAction): ...
    if la, ok := act.(*actions.LocalAction); ok {
        // P21: with ivy_logic.WithSymbols(self.args[0:-1]):
        ...
        // P22: return self.clone(list(map(recur, self.args)))
        return la.ActionClone(newArgs)
    }

    // P23: return self.clone(list(map(recur, self.args)))
    // ALWAYS clones — no "changed" optimization.
    ...
    return act.ActionClone(newArgs)
}
```

### Step 3: Add helper functions for cleaner assert handling

Add `getAssertAction(act) *AssertAction` that extracts the embedded AssertAction from any assert-like type, and `cloneAssertWithoutProof(act)` that clones preserving the concrete type with proof stripped. This avoids the repetitive 4-way type switch.

### Step 4: Run `make golden`

Verify the divergence is resolved.

## Files to Modify

1. `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/phase6.go` — Rewrite `recur` with point-by-point Python mapping comments, verify Bug 5 fix

## Verification

1. `cd ~/goivy && make golden` — the divergence at line 146905 should be resolved
