# Plan: Fix `stackLookup` to walk parent accum chain (like Python's `stack`)

**Created:** 2026-03-23 (during research session)

## Context

`make golden` fails at trace line 192. When parsing `order.ivy`, the module
`totally_ordered_with_zero` contains `instantiate totally_ordered(t)`.
Python finds `totally_ordered` because `stack_lookup()` walks a global
`stack` list (outermost→innermost). The Go `stackLookup()` only checks the
**current** (innermost) accum and returns nil, so the instantiation falls
through to `others` and is declared as an unresolved `InstantiateDecl`.

### Root cause

Python `ivy_parser.py`:
- Line 383: `p_top` rule appends new `Ivy()` to global `stack`
- Line 651: `p_moduleend` pops from `stack`
- Line 117-122: `stack_lookup` iterates `reversed(stack)` — walks parents

Go `lalr_full/`:
- `newIvyAccum()` has no parent pointer
- `stackLookup()` only checks `ivy.modules[name]` — flat, no parent walk

## Fix

### 1. Add `parent` field to `ivyAccum` (`ivy_module.go`)

```go
type ivyAccum struct {
    parent   *ivyAccum  // <-- NEW: link to enclosing scope
    decls    []ast.Node
    modules  map[string]*ast.ModuleDecl
    actions  map[string]*ast.ActionDecl
    defined  map[string]bool
    ...
}
```

### 2. Set parent when creating inner `top` accums (`grammar_v17.y`)

The `top: /* empty */` rule creates a new accum. When it fires inside a
module body (e.g. `top MODULE modulestart modcat atom optwith = { top }`),
the inner `top` needs its parent set.

The cleanest approach matching Python's `stack.append / stack.pop`:
maintain a package-level `accumStack []*ivyAccum`. On every `top: /* empty */`
reduction, push the new accum. The parent is set to the previous stack top.

In `grammar_v17.y`, the `top: /* empty */` rule:
```go
top:
    /* empty */
    {
        xtracer.Trace("parser.p_top ENTER (top)")
        parent := v17lex.(*v17LexAdapter).accum  // current accum is parent
        $$ = newIvyAccum()
        $$.parent = parent
        v17lex.(*v17LexAdapter).accum = $$
    }
```

This works because `lex.accum` always points to the most recently created
accum. When a nested `top` reduces, the previous `lex.accum` becomes the
parent of the new accum.

### 3. Fix `stackLookup` to walk parent chain (`inst_mod.go`)

```go
func stackLookup(ivy *ivyAccum, name string) *ast.ModuleDecl {
    xtracer.Trace("parser.stack_lookup ENTER")
    for cur := ivy; cur != nil; cur = cur.parent {
        if md, ok := cur.modules[name]; ok {
            return md
        }
    }
    return nil
}
```

### 4. Also fix `stack_action_lookup` if present

Check if Go has an equivalent of Python's `stack_action_lookup` (line 125-133)
and apply the same parent-walk pattern.

## Files to modify

1. **`/Users/jaten/goivy/lalr_full/ivy_module.go`** — add `parent *ivyAccum` field to struct
2. **`/Users/jaten/goivy/lalr_full/grammar_v17.y`** — set `parent` in `top: /* empty */` rule
3. **`/Users/jaten/goivy/lalr_full/inst_mod.go`** — fix `stackLookup` to walk parent chain

## Verification

```bash
cd ~/goivy && make golden
```

Should advance past line 192 (the `totally_ordered` instantiation). The next
divergence (if any) will be a different issue.
