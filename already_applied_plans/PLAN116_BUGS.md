# Code Review: compiler/, actions/, logic/ Post-Refactoring

**Created:** 2026-03-27 18:45

## Context

Steps 1-7 of the lg.Expr refactoring are COMPLETE. All 30+ action types now implement lg.Expr directly. WrapAction/UnwrapAction eliminated. Golden test advanced from 124567 to 124642. This plan captures all bugs and recommended refactors found during code review.

---

## BUGS

### BUG 1: Duplicate type assertions — dead else branches (phase6.go)

**Severity:** HIGH — dead code, likely copy-paste errors hiding a real intent

Three locations have identical if/else-if chains that check the same type twice:

**File:** `compiler/phase6.go`

**Location A — lines 1939-1943** (WhileAction invariants):
```go
if subAct, ok := inv.(actions.Action); ok {
    r = recur(subAct)
} else if subAct, ok := inv.(actions.Action); ok {  // DEAD — same check
    r = recur(subAct)
}
```

**Location B — lines 1958-1962** (WhileAction body):
```go
if bodyAct, ok := w.Body.(actions.Action); ok {
    newBody = recur(bodyAct)
} else if bodyAct, ok := w.Body.(actions.Action); ok {  // DEAD
    newBody = recur(bodyAct)
}
```

**Location C — lines 1980-1983** (LocalAction args):
```go
if subAct, ok := arg.(actions.Action); ok {
    newArgs[i] = recur(subAct)
} else if subAct, ok := arg.(actions.Action); ok {  // DEAD
    newArgs[i] = recur(subAct)
}
```

**Fix:** These were likely introduced during the WrapAction removal. The else-if was probably meant to handle the old `ActionNodeWrapper` case. Now that actions implement lg.Expr directly, the else-if branches are unreachable. Remove all three dead else-if blocks. Check the Python source to confirm there's no second dispatch path being missed.

---

### BUG 2: Nil pointer dereference in CompileCall (action.go:736)

**Severity:** HIGH — panic if TopCtx is nil

**File:** `compiler/action.go`, line 736

```go
// Line 691: if c.TopCtx != nil && name != "" {
//   ... returns early if name NOT in Actions ...
// }
// Line 736: falls through when TopCtx is nil!
info := c.TopCtx.Actions[name]  // PANIC: nil dereference
```

The guard at line 691 skips the "name not found" fallback when TopCtx is nil, but line 736 unconditionally dereferences TopCtx.

**Fix:** Add `if c.TopCtx == nil { return nil, fmt.Errorf("no top context for call to %q", name) }` before line 736. Check Python's `compile_call` to see what it does when `top_context` is None (likely it would also fail, so an error return is correct).

---

### BUG 3: Unsafe type assertion — prop.Formula (phase6.go:2156)

**Severity:** MEDIUM — panic if Formula is not lg.Expr

**File:** `compiler/phase6.go`, line 2156

```go
if isSchemaBody(prop.Formula.(lg.Expr)) {
```

Bare assertion without comma-ok. If `prop.Formula` holds an ast.Node that doesn't implement lg.Expr, this panics.

**Fix:** Use `if expr, ok := prop.Formula.(lg.Expr); ok && isSchemaBody(expr) {`

---

### BUG 4: Unsafe type assertion — cn.Node (phase6.go:825)

**Severity:** MEDIUM — panic if CompiledNode.Node is not lg.Expr

**File:** `compiler/phase6.go`, line 825

```go
withExprs = append(withExprs, cn.Node.(lg.Expr))
```

**Fix:** Use `if expr, ok := cn.Node.(lg.Expr); ok { withExprs = append(withExprs, expr) }`

---

### BUG 5: Unsafe type assertions on array elements (ivy_compile.go:967, 990)

**Severity:** MEDIUM — panic if choices[0] or afters[0] is not actions.Action

**File:** `compiler/ivy_compile.go`

Line 967:
```go
act.SetLineno(choices[0].(actions.Action).GetLineno())
```

Line 990:
```go
seqAct.SetLineno(afters[0].(actions.Action).GetLineno())
```

**Fix:** Add comma-ok type assertions before accessing GetLineno().

---

### BUG 6: Unsafe type assertion — mlf.Formula (actions/action.go:2152)

**Severity:** MEDIUM — panic in actions package

**File:** `actions/action.go`, line 2152

```go
clauses := co.FormulaToClauses(mlf.Formula.(lg.Expr), nil)
```

**Fix:** Use comma-ok: `if fmla, ok := mlf.Formula.(lg.Expr); ok { clauses := co.FormulaToClauses(fmla, nil); ... }`

---

### BUG 7: Panic instead of error return (ivy_compile.go:1805)

**Severity:** LOW — panic for invariant violation should be error

**File:** `compiler/ivy_compile.go`, line 1805

```go
if _, isDef := conc.(*lg.Definition); isDef {
    panic(fmt.Sprintf("definitional subgoal must be discharged"))
}
```

**Fix:** Replace with `return nil, fmt.Errorf("definitional subgoal must be discharged")`

---

## INCONSISTENCIES

### INC 1: Comment-code mismatch — InstantiateDecl routing

**File:** `compiler/compiler.go`, line 289

Comment says "SetAction, HavocAction, InstantiateDecl use other_thing (default)" but `*ast.InstantiateDecl` IS handled in `CompileActionBody` (action.go:366). The comment is accurate about CompileNode routing (InstantiateDecl is NOT in the case list), but misleading about CompileActionBody.

**Fix:** Clarify comment: "SetAction, HavocAction use other_thing. InstantiateDecl is not routed from CompileNode but IS handled in CompileActionBody when reached from other callers."

---

### INC 2: Trace messages say "CompileNode" inside CompileActionBody

**File:** `compiler/action.go`, lines 185, 393, 413, etc.

The traces are intentionally named this way to match Python's trace output (Python emits "CompileNode return case=..." from within the `.cmpl` handlers which are called by `thing()` which is the Python equivalent of CompileNode). This is NOT a bug — it's correct for golden test conformance. No fix needed.

---

## RECOMMENDED REFACTORS (Python-conformance oriented)

### REF 1: Verify Clone() methods use safe assertions (logic/ast_compat.go)

**Files:** `logic/ast_compat.go` — multiple Clone() methods

All Clone() methods use bare `arg.(Expr)` and `arg.(Sort)` assertions. While these are internal-only calls and should always receive correct types, adding comma-ok checks would prevent obscure panics during development.

**Priority:** LOW — these match Python's pattern where clone just copies args. Only fix if a panic is observed.

---

### REF 2: Add ActionS handling to pretty.go functions (logic/pretty.go)

**File:** `logic/pretty.go`

Functions `ugly()` (line 40), `dropAnnotations()` (line 226), and `sortName()` (line 400) have type switches that don't explicitly handle action nodes or ActionSort.

- `ugly()`: Action nodes fall to default → `fmt.Sprint(n)` — acceptable but may not match Python's display
- `dropAnnotations()`: Action nodes fall to default → returned unchanged — likely correct since Python annotations are only on formulas
- `sortName()`: ActionSort falls to `s.String()` → "ActionSort" — acceptable

**Priority:** LOW — these paths are unlikely to be hit during normal compilation. Only fix if golden test divergence is observed in pretty-printing.

---

### REF 3: Type ActionContext.Domain as *module.Module (actions/action.go)

**File:** `actions/action.go`, line 1083

`ActionContext.Domain` is typed `interface{}` but always holds `*module.Module`. Typing it properly would eliminate type assertions in `Get()` and other methods.

**Priority:** LOW — this is a quality-of-life improvement. Check if Python's equivalent uses a typed field or generic one before changing.

---

## FILES WITH BUGS

| File | Bug IDs | Lines |
|------|---------|-------|
| `compiler/phase6.go` | BUG 1 (A,B,C), BUG 3, BUG 4 | 825, 1939-43, 1958-62, 1980-83, 2156 |
| `compiler/action.go` | BUG 2 | 736 |
| `compiler/ivy_compile.go` | BUG 5, BUG 7 | 967, 990, 1805 |
| `actions/action.go` | BUG 6 | 2152 |

## VERIFICATION

After fixing:
```bash
go build ./...
go test ./compiler/
go test ./actions/
go test ./logic/
cd ~/goivy && make golden   # should not regress from line 124642
```
