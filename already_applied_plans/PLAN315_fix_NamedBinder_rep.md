# PLAN: Fix compile_app trace divergence for App-with-NamedBinder-rep (xtrace 775993)

Created: 2026-04-16 (late afternoon, sixth pass)

## Context

After the previous fix (concept-space label variable ordering at xtrace 406841), the user re-ran `make tlb` and the trace now matches all the way to the L2S tactic's `compileInvar[10]` pre-compile of `invar380`. The new divergence is at xtrace event **775993** (`/Users/jaten/ivy/goivy/log.tlb:73267-73268`):

```
775990  go : XTRACE: compiler.Thing ENTER type=App
        py : XTRACE: compiler.Thing ENTER type=App

775991  go : XTRACE: compiler.CompileNode ENTER type=App
        py : XTRACE: compiler.CompileNode ENTER type=App

775992  go : XTRACE: compiler.CompileNode return case=App
        py : XTRACE: compiler.CompileNode return case=App

775993  go : XTRACE: compiler.Thing ENTER type=NamedBinder       <-- DIVERGES
        py : XTRACE: compiler.compile_app ENTER old=False
```

Both sides agree the AST node is an `*ast.App` (not `*ast.Atom`) and both dispatch to the App handler. Then they diverge: Python emits `compile_app ENTER`; Go skips that trace entirely and instead recurses into the App's `Rep` (a NamedBinder) via the `Thing` wrapper.

This is the formula `invar380`, which contains the term `(app rep:(namedBinder name:"l2s_w" bounds:[(variable rep:"P" vSort:...)]) ...)` (see log line 73154). So the App has `Rep = *ast.NamedBinder`.

### Root cause (NamedBinder-rep branch only)

**Python** (`~/ivy/pyivy/ivy/ivy/ivy_compiler.py:382-408, 457`) uses ONE function for both Atom and App:

```python
def compile_app(self,old=False):
    xtracer.trace("compiler.CompileNode return case=%s" % type(self).__name__)
    xtracer.trace("compiler.compile_app ENTER old=%s" % old)              # <-- always emitted
    rep = resolve_alias(self.rep) if isinstance(self.rep,str) else self.rep
    if rep == "true" or rep == "false": ...
    xtracer.trace("compiler.CompileApp args loop nTerms=%d" % (len(self.args)))
    with ReturnContext(None):
        args = [a.compile() for a in self.args]                            # <-- args FIRST
    ...
    sym = rep.cmpl() if isinstance(rep,ivy_ast.NamedBinder) else ...       # <-- rep AFTER args, via cmpl() not thing()
    ...
    return (sym)(*args)

ivy_ast.App.cmpl = ivy_ast.Atom.cmpl = compile_app                         # one function for both
```

So Python: any `compile_app` invocation emits `"compile_app ENTER"` + args-loop trace, compiles args first, then calls `rep.cmpl()` (no `thing()` wrapper, no `Thing ENTER` trace) when rep is a NamedBinder.

**Go** (`/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go:712-846`) splits the responsibility across two functions:

- `CompileApp(*ast.Atom, bool)` (line 712) — handles Atom (string rep). Already emits `"compile_app ENTER old=%v"` at line 713 and the args-loop trace at line 739. ✓
- `compileAppNode(*ast.App)` (line 814) — handles App (Node rep). The Symbol-rep branch (lines 815-822) routes to `CompileApp` and so emits the correct traces via `CompileApp`. The **NamedBinder-rep branch** (lines 823-845) has three concrete divergences from Python:

```go
// CURRENT: NamedBinder-rep branch (lines 823-845, the only branch this plan touches)
repNode, err := c.Thing(n.Rep)         // line 824 — emits "Thing ENTER type=NamedBinder" ✗
if err != nil { return nil, err }
saved := c.ReturnCtx
c.ReturnCtx = nil
args := make([]lg.Expr, len(n.Terms))
for i, a := range n.Terms {
    r, err := c.Thing(a)               // args compiled AFTER rep ✗
    ...
}
c.ReturnCtx = saved
if len(args) == 0 { return repNode, nil }
return lg.NewApply(repNode, args...)
```

1. Missing `compile_app ENTER old=False` trace.
2. Missing `CompileApp args loop nTerms=N` trace.
3. Wrong order: rep compiled BEFORE args. Python compiles args first.
4. `Thing(rep)` instead of `Cmpl(rep)`: Python's `rep.cmpl()` is a direct cmpl() call without the `thing()` wrapper, so it must NOT emit `Thing ENTER`.

### Why `Cmpl` is the right helper

Go already has the matching helper:

```go
// Cmpl compiles an AST node via the type-specific handler, matching
// Python's cmpl() dispatch. Unlike CompileNode, it does NOT emit the
// "CompileNode ENTER" trace — only the type-specific "CompileNode return case=X"
// trace fires. Use Cmpl where Python calls a.cmpl() (e.g., callee args in
// compile_call's "in actions" path).
func (c *Compiler) Cmpl(node ast.Node) (lg.Expr, error) {
    return c.compileNodeCore(node, false)
}
```

This is exactly what Python's `rep.cmpl()` does in `compile_app:399`. The dispatcher case for `*ast.NamedBinder` (compiler.go:302-304) emits `"CompileNode return case=NamedBinder"` and then `compileNamedBinder` emits `"CompileNamedBinder ENTER"` — matching Python's `_named_binder_cmpl` (ivy_compiler.py:507-516) line-for-line.

## Fix

Single-file edit: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go`, function `compileAppNode` lines 823-845 (the NamedBinder-rep branch only).

**The Symbol-rep branch at lines 815-822 is left entirely unchanged.** This plan does NOT touch the Symbol→Atom conversion pattern that exists there (and elsewhere in the file at lines 893-902 for compileOld and in phase6.go:401-411 for CompileOld). Those are pre-existing and out of scope for fixing xtrace 775993.

Replace the NamedBinder-rep branch (current lines 823-845) with:

```go
// Match Python's compile_app trace order for the NamedBinder-rep case.
// Python ivy_compiler.py:382-408, 457 has App.cmpl == Atom.cmpl ==
// compile_app, so any rep type that reaches compile_app emits the same
// "compile_app ENTER" + args-loop trace sequence.
xtracer.Trace("compiler.compile_app ENTER old=%v", false)
xtracer.Trace("compiler.CompileApp args loop nTerms=%d", len(n.Terms))

// Args FIRST (Python ivy_compiler.py:391-394: with ReturnContext(None):
// args = [a.compile() for a in self.args]).
saved := c.ReturnCtx
c.ReturnCtx = nil
args := make([]lg.Expr, len(n.Terms))
for i, a := range n.Terms {
    r, err := c.Thing(a)
    if err != nil {
        c.ReturnCtx = saved
        return nil, err
    }
    args[i] = r
}
c.ReturnCtx = saved

// Then compile the rep — must be NamedBinder here (Python ivy_compiler.py:399:
// sym = rep.cmpl() if isinstance(rep,ivy_ast.NamedBinder) else ...).
// Use Cmpl (not Thing) to match Python's direct rep.cmpl() call without
// the thing()/Thing() wrapper trace.
repNode, err := c.Cmpl(n.Rep)
if err != nil {
    return nil, err
}

if len(args) == 0 {
    return repNode, nil
}
return lg.NewApply(repNode, args...)
```

The full updated `compileAppNode` after the edit:

```go
func (c *Compiler) compileAppNode(n *ast.App) (lg.Expr, error) {
    // Symbol-rep branch UNCHANGED.
    if sym, ok := n.Rep.(*ast.Symbol); ok {
        cfg := c.Module.Cfg.AstCfg
        atom := cfg.NewAtom(sym.Rep, n.Terms...)
        atom.SetLineno(n.GetLineno())
        atom.ASort = n.ASort
        return c.CompileApp(atom, false)
    }

    // NamedBinder-rep branch: emit traces, compile args first, then Cmpl(rep).
    xtracer.Trace("compiler.compile_app ENTER old=%v", false)
    xtracer.Trace("compiler.CompileApp args loop nTerms=%d", len(n.Terms))

    saved := c.ReturnCtx
    c.ReturnCtx = nil
    args := make([]lg.Expr, len(n.Terms))
    for i, a := range n.Terms {
        r, err := c.Thing(a)
        if err != nil {
            c.ReturnCtx = saved
            return nil, err
        }
        args[i] = r
    }
    c.ReturnCtx = saved

    repNode, err := c.Cmpl(n.Rep)
    if err != nil {
        return nil, err
    }
    if len(args) == 0 {
        return repNode, nil
    }
    return lg.NewApply(repNode, args...)
}
```

### Trace sequence after fix (App-with-NamedBinder-rep case)

```
Thing ENTER type=App                       (from outer Thing wrapper, unchanged)
CompileNode ENTER type=App                 (from CompileNode dispatcher)
CompileNode return case=App                (from compiler.go:284, unchanged)
compile_app ENTER old=false                (NEW — added by fix)
CompileApp args loop nTerms=N              (NEW — added by fix)
Thing ENTER type=...                       (per arg — unchanged)
... per-arg compile traces ...
Thing return type=...                      (per arg)
CompileNode return case=NamedBinder        (from dispatcher via Cmpl, no Thing ENTER)
CompileNamedBinder ENTER                   (from compileNamedBinder — already correct)
... bound vars and body compile traces ...
Thing return type=App                      (from outer Thing wrapper)
```

This matches Python's `compile_app(App)` → `[a.compile() for a in self.args]` → `rep.cmpl()` sequence exactly.

## What we are NOT changing

- The Symbol-rep branch in `compileAppNode` (lines 815-822) — pre-existing, out of scope.
- `compileOld` (compiler.go:889-906) and `CompileOld` (phase6.go:395-414) — both contain the same Symbol→Atom pattern; pre-existing, out of scope.
- `CompileApp` (the Atom-rep path) — already emits `compile_app ENTER` correctly.
- `compileNamedBinder` (compiler.go:988-…) — already correct; its traces match Python's `_named_binder_cmpl`.
- The CompileNode dispatcher case for `*ast.App` (line 285) — emits `CompileNode return case=App` correctly.
- `Cmpl` and `Thing` helpers — both correct; we just need to use `Cmpl` (not `Thing`) for the rep.

## Verification

**Do NOT run `make tlb` from this agent — the user runs it on another machine.**

Local checks this agent will run after the user approves:

1. `go build ./compiler/...` — confirm the change compiles.
2. `go test ./compiler/...` — confirm existing compiler tests still pass. If a test was relying on the old wrong trace order, that test was wrong; investigate before changing it.
3. Optional: `go vet ./compiler/...` — confirm no new vet warnings.

After the user re-runs `make tlb` on the other machine, expected outcomes:

A. **Best case**: xtrace 775993 now matches. Trace continues into `CompileApp args loop` for nTerms=1 (the NamedBinder's argument), then into per-arg compile, then into `CompileNode return case=NamedBinder` + `CompileNamedBinder ENTER`. Next divergence (if any) appears further along.

B. **Possible case**: trace matches up through `CompileApp args loop` and per-arg compile but diverges at `CompileNode return case=NamedBinder` because Go is emitting a different trace. That would mean a stale `c.Thing(n.Rep)` survives somewhere; re-grep `compileAppNode` for accidental Thing-on-rep calls.

C. **Unexpected case**: Go emits `compile_app ENTER` but at a different point than Python. That would mean the dispatcher in compiler.go:284 needs reordering — but this should not happen since the fix only restructures `compileAppNode`'s NamedBinder branch.

## Critical files

- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go` — Edit `compileAppNode` NamedBinder-rep branch (lines 823-845)

## Reference (source of truth)

- `~/ivy/pyivy/ivy/ivy/ivy_compiler.py:382-408` — `compile_app` (handles both Atom and App)
- `~/ivy/pyivy/ivy/ivy/ivy_compiler.py:457` — `ivy_ast.App.cmpl = ivy_ast.Atom.cmpl = compile_app`
- `~/ivy/pyivy/ivy/ivy/ivy_compiler.py:507-516` — `_named_binder_cmpl` (assigned to `ivy_ast.NamedBinder.cmpl`)
- `~/ivy/pyivy/ivy/ivy/ivy_compiler.py:88-94` — `thing()` (the wrapper that emits `Thing ENTER`)
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go:216-218` — `Cmpl` (Python's `cmpl()` direct dispatch, no Thing wrapper)
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/phase6.go:50-55` — `Thing` (the wrapper that emits `Thing ENTER`)
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go:712-811` — `CompileApp` (the Atom-rep path, already correct)
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go:813-846` — `compileAppNode` (the function to edit; only NamedBinder branch lines 823-845)
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go:283-285` — dispatcher case `*ast.App`
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go:302-304` — dispatcher case `*ast.NamedBinder`
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go:988-…` — `compileNamedBinder`
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/lalr_logicparser/grammar_v17.go:1287-1289` — confirms parser produces `App{Rep: *ast.NamedBinder}` (grammar case 62)

## Thing↔Cmpl audit (full pass over all Python `.cmpl()` sites)

Audit done across `~/ivy/pyivy/ivy/ivy/*.py` to catch every site where Python calls `X.cmpl()` directly (skipping the `thing()`/`Thing` wrapper) and confirm Go uses `Cmpl` (not `Thing`) at each.

| # | Python site | Context | Go site | Go usage | Status |
|---|---|---|---|---|---|
| 1 | `ivy_compiler.py:92` | `result = self.cmpl()` inside `thing()` itself (the dispatch path) | `compiler/phase6.go:53` — `c.CompileNode(node)` inside `Thing()` | dispatch, not a wrapper-skip case | ✓ correct by design |
| 2 | `ivy_compiler.py:399` | `rep.cmpl()` in `compile_app` (NamedBinder rep branch) | `compiler/compiler.go:824` — `c.Thing(n.Rep)` | **MISMATCH** | ✗ **fixed by this plan** |
| 3 | `ivy_compiler.py:660` | commented-out line `# return c.cmpl()` | — | — | N/A |
| 4 | `ivy_compiler.py:710` | `[a.cmpl() for a in self.args[1:]]` in `compile_call` non-action / field-reference returns | `compiler/action.go:847` — `c.Cmpl(r)` | matches | ✓ correct |
| 5 | `ivy_compiler.py:720` | `args = [a.cmpl() for a in self.args[0].args]` in `compile_call` action-call callee args | `compiler/action.go:895` — `c.Cmpl(a)` | matches | ✓ correct |
| 6 | `ivy_compiler.py:729` | `[a.cmpl() for a in self.args[1:]]` in `compile_call` action-call returns into `CallAction` | `compiler/action.go:949` — `c.Cmpl(r)` | matches | ✓ correct |
| 7 | `ivy_actions.py:823` | dead code: appears after `return self` on line 822 in `InstantiateAction.cmpl` | — | — | N/A (unreachable) |

**Result**: the audit identified exactly ONE live mismatch — the one this plan fixes. All three other live sites (`action.go:847`, `:895`, `:949` in `CompileCall`) already use `Cmpl` correctly. The two N/A sites are commented-out / dead code.

A complementary reverse-direction audit was also performed: `grep c.Cmpl(` across all of `goivy/` returns exactly the three `action.go` call sites listed above — confirming there are no stray Go `Cmpl` calls without a corresponding Python `.cmpl()`.

After the edit in this plan, every live Python `.cmpl()` site has a matching Go `Cmpl` site, and no Go `Cmpl` call is unmotivated.

## What we are NOT investigating in this plan

- Cleaning up the Symbol→Atom conversion pattern repeated in `compileAppNode`, `compileOld`, and `CompileOld`. The user confirmed this is out of scope for fixing xtrace 775993; address as a separate refactor.
- The original `invar386` TopSort panic. Each fix in this sequence moves the trace divergence further along; the panic, if still present, will surface once the trace is fully aligned.
