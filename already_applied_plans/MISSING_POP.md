# Plan: Fix LALR Parser Scope Management — 29 vs 39 Decl Divergence

**Created:** 2026-03-23 23:45

## Context

After applying all 17 grammar variance fixes, `make golden` shows a divergence:
```
4872  go : XTRACE: parser.Parse EXIT decls=29
      py : XTRACE: parser.Parse EXIT decls=39
```

The Go LALR parser produces 10 fewer declarations than Python for the order.ivy
include file (or possibly another nested parse). Root cause analysis reveals a
critical scope management bug: **Go's `v17lex.accum` is never restored after
exiting nested scopes**, which is Python's `stack.pop()`. This corrupts the
parent chain used for module lookups during instantiation expansion.

## Root Cause: Missing `stack.pop()` Equivalent

### Python's Scope Management

Python maintains a global `stack` (list of `Ivy` objects, ivy_parser.py:106).
Each nested scope (module body, object body) pushes a new `Ivy` onto the stack
via the `top :` empty production (line 384: `stack.append(p[0])`). After
processing, `stack.pop()` restores the parent scope:

- **Module rule** (line 651): `stack.pop()` + `stack[-1].is_module = False`
- **Object rule** via `create_object` (line 692): `stack.pop()`
- **Include rule** (line 409): `stack.pop()` (removes the nested parse's Ivy)

### Go's Broken Equivalent

Go uses `v17lex.accum` as a pointer to the current scope, with a `parent` chain
replacing the Python stack. The `top :` empty production (grammar_v17.y:507-515)
creates a new `ivyAccum` and sets `v17lex.accum = $$`:

```go
parent := v17lex.(*v17LexAdapter).accum
$$ = newIvyAccum()
$$.parent = parent
v17lex.(*v17LexAdapter).accum = $$
```

But **no rule ever restores `v17lex.accum` to the parent**. After a module or
object body is parsed, `v17lex.accum` still points to the INNER accumulator.
When the NEXT sibling scope starts, its `top :` fires with `parent = v17lex.accum`
pointing to the stale inner accumulator instead of the outer scope.

### How This Corrupts Module Lookups

Parsing order.ivy:
```
module totally_ordered(t) = { ... }        # body → accum B
module totally_ordered_with_zero(t) = {    # body → accum C
    instantiate totally_ordered(t)         # doInsts looks up via C's parent chain
}
```

Without fix: C.parent = B (sibling!) instead of A (outer). Lookups walk
C → B → A, adding an extra hop through a sibling's body. While this USUALLY
still finds modules (because B.parent = A), the chain is incorrect and may
cause subtle failures in more complex cases where sibling scopes have modules
or defined names that interfere with lookups.

## Critical Files to Modify

1. `~/goivy/lalr_full/grammar_v17.y` — restore `v17lex.accum` after nested scopes
2. `~/goivy/lalr_full/ivy_module.go` — add `is_module` tracking
3. Regenerate `~/goivy/lalr_full/grammar_v17.go` — via `go generate`

## Fixes

### Fix A: Restore `v17lex.accum` in Module Rule (CRITICAL)

**File:** `grammar_v17.y` line ~632-641

**Current:**
```go
| top TOK_MODULE modulestart modcat atom optwith TOK_EQ TOK_LCB top TOK_RCB moduleend
{
    $$ = $1
    modAccum := $9
    body := ast.NewSequence(modAccum.decls...)
    d := ast.NewDefinition(ast.AppToAtom($5), body)
    $$.declare(ast.NewModuleDecl(d))
}
```

**Fix:** Add `v17lex.accum` restoration (Python's `stack.pop()`):
```go
| top TOK_MODULE modulestart modcat atom optwith TOK_EQ TOK_LCB top TOK_RCB moduleend
{
    $$ = $1
    modAccum := $9
    body := ast.NewSequence(modAccum.decls...)
    d := ast.NewDefinition(ast.AppToAtom($5), body)
    $$.declare(ast.NewModuleDecl(d))
    // Python: stack.pop() — restore scope after processing module body
    v17lex.(*v17LexAdapter).accum = $$
}
```

### Fix B: Restore `v17lex.accum` in Object Rule

**File:** `grammar_v17.y` line ~643-651

**Current:**
```go
| top TOK_OBJECT objsym objectargs TOK_EQ TOK_LCB optdotdotdot top TOK_RCB objectend
{
    $$ = $1
    objAccum := $8
    pref := $3.(*ast.Atom)
    lineno := getLineno(v17lex.(*v17LexAdapter))
    createObject($$, pref, $4, objAccum, lineno, $7)
}
```

**Fix:** Add restoration:
```go
    createObject($$, pref, $4, objAccum, lineno, $7)
    // Python: create_object does stack.pop() at line 692
    v17lex.(*v17LexAdapter).accum = $$
```

### Fix C: Restore `v17lex.accum` in Class Rule

**File:** `grammar_v17.y` line ~653-671

Add at end of semantic action:
```go
    v17lex.(*v17LexAdapter).accum = $$
```

### Fix D: Restore `v17lex.accum` in Subclass Rule

**File:** `grammar_v17.y` line ~673-685

Add at end of semantic action:
```go
    v17lex.(*v17LexAdapter).accum = $$
```

### Fix E: Restore `v17lex.accum` in Extract Rule

**File:** `grammar_v17.y` line ~1050 (extract objsym rule with LCB/RCB)

Add at end of semantic action:
```go
    v17lex.(*v17LexAdapter).accum = $$
```

### Fix F: Add `is_module` Tracking in `modulestart`

**File:** `grammar_v17.y` line ~2712

**Current:**
```go
modulestart:
    /* empty */
    {
        xtracer.Trace("parser.p_modulestart ENTER (modulestart)")
        $$ = nil
    }
```

**Fix:** Match Python `stack[-1].is_module = True`:
```go
modulestart:
    /* empty */
    {
        xtracer.Trace("parser.p_modulestart ENTER (modulestart)")
        v17lex.(*v17LexAdapter).accum.isModule = true
        $$ = nil
    }
```

And in the module rule, after `v17lex.accum = $$`:
```go
    $$.isModule = false  // Python: stack[-1].is_module = False
```

### Fix G: Store Module Body as `ivyAccum` Reference (Not Just Sequence)

**File:** `grammar_v17.y` line ~638

**Current:** Module body stored as `ast.NewSequence(modAccum.decls...)` — loses
the inner scope's `modules`, `defined`, `static`, and `objects` maps.

**Python:** Passes the full `Ivy` object as the module body: `d = Definition(app_to_atom(p[5]), p[9])`

**Fix:** Store the `ivyAccum` pointer alongside the declaration sequence so
`doInsts`/`instMod` can access inner modules and defined names directly.

Option 1 (preferred): Add `ModuleAccum` field to the grammar's `ModuleDecl`:
```go
modAccum := $9
body := ast.NewSequence(modAccum.decls...)
d := ast.NewDefinition(ast.AppToAtom($5), body)
md := ast.NewModuleDecl(d)
// Store the parsed module's accum for later inst_mod access
md.ParsedAccum = modAccum  // New field on ModuleDecl
$$.declare(md)
```

Then in `doInsts`, when looking up the module body:
```go
if md.ParsedAccum != nil {
    // Use the stored accum's decls and inner modules directly
    bodyDecls = md.ParsedAccum.decls
    // Can also access md.ParsedAccum.modules for nested module lookup
}
```

**Alternative (simpler):** Store the inner `modules` map in `ModuleDecl`:
```go
md.InnerModules = modAccum.modules
```

This preserves Python's behavior where `module.modules` is accessible during
`do_insts` when expanding nested instantiations. order.ivy has modules like
`unbounded_sequence` that contain `object spec` and `object impl` — the inner
module map tracks these sub-objects.

### Fix H: Propagate `decls` count change in `doInsts` trace

**File:** `~/goivy/lalr_full/inst_mod.go` line ~129

**Current:** `xtracer.Trace("parser.do_insts EXIT decls=%d", len(others))`

The Go trace reports `len(others)` (unresolved count) matching Python. Ensure
the trace format stays identical after fixes.

## Execution Order

1. Apply Fixes A-F to `grammar_v17.y` (scope restoration + is_module tracking)
2. Regenerate: `cd ~/goivy/lalr_full && go generate`
3. Build: `cd ~/goivy && go build ./...`
4. Test: `cd ~/goivy && make golden`
5. If the 29 vs 39 divergence persists, apply Fix G (module body storage)
6. If Fix G is needed, also update `doInsts` and `instMod` to use stored accum
7. Re-test after each fix

## Verification

```bash
cd ~/goivy && make golden
```

**Success criteria:** The golden test advances past line 4872 (the
`parser.Parse EXIT decls=29 vs 39` mismatch). Ideally both report `decls=39`.

**Secondary check:** `go build ./...` compiles clean. Conflict count unchanged.

## Summary

| Fix | What | Python Equivalent |
|-----|------|-------------------|
| A | Restore v17lex.accum after module rule | `stack.pop()` at line 651 |
| B | Restore v17lex.accum after object rule | `stack.pop()` in create_object line 692 |
| C | Restore v17lex.accum after class rule | `stack.pop()` in module rule (class variant) |
| D | Restore v17lex.accum after subclass rule | `stack.pop()` in module rule |
| E | Restore v17lex.accum after extract rule | `stack.pop()` in module rule |
| F | Set is_module in modulestart | `stack[-1].is_module = True` line 601 |
| G | Store module accum for inst_mod | Python passes full Ivy object as module body |
| H | Verify doInsts trace format | Keep `decls=%d` format matching Python |
