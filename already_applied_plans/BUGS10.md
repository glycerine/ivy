# Plan: Fix createObject and instMod to match Python — line 1926 divergence

**Created:** 2026-03-23 18:00

## Context

`make golden` compares xtrace output line-by-line between Python `ivy_check`
and Go `goivy_check_xtrace` on `ord_live.ivy` with `isolate=cf_live`. Previous
fixes advanced matching from line 192 to line 1926. At line 1926, Go and Python
diverge in declaration stream position — Go reduces `p_atype_symbol` while
Python sees `p_SYMBOL_PRESYMBOL`. The root cause is `createObject` producing
structurally different prefix atoms and missing variable substitutions, causing
`instMod`/`SubstPrefixAtomsAst` to expand object bodies differently.

## Root Cause Analysis

Comparing Python `create_object` (ivy_parser.py:678-693) vs Go `createObject`
(grammar_v17.y:168-180):

### BUG 1 (CRITICAL): createObject ignores objectargs — no Variable prefix args

**Python:**
```python
prefargs = [Variable('V'+str(idx), pr.sort) for idx,pr in enumerate(objectargs)]
pref = Atom(name, prefargs)
vsubst = dict((pr.rep, v) for pr, v in zip(objectargs, prefargs))
inst_mod(top, module, pref, {}, vsubst)
```

**Go:**
```go
pref := ast.NewAtom(name.Rep)  // NO args!
instMod(top, module.decls, pref, map[string]string{}, map[string]*ast.Variable{}, "")  // empty vsubst!
```

Go creates a bare prefix atom with no Variable arguments. Python creates fresh
Variables (V0, V1, ...) mapped to the object's parameter sorts, and builds a
vsubst mapping parameter names → those fresh Variables. This means:
- Prefix compositions via `ComposeAtoms` differ (args vs no args)
- Variable substitution path in `instMod` is never taken in Go

### BUG 2: objectargs rule doesn't set params on accum

**Python** (ivy_parser.py:670): `stack[-1].params = p[0]`
**Go** (grammar_v17.y:2651-2657): Only returns `$1`, doesn't set `params` on accum.

### BUG 3: instMod missing "common" attribute handling

**Python** (line 161): `dpref = pref.clone([]) if "common" in decl.attributes else pref`
**Go** (line 153): `dpref := pref` — always uses pref as-is.

### BUG 4: instMod missing AttributeDecl special handling

**Python** (lines 163-171): Uses `compose_atoms` directly for AttributeDecl.
**Go**: Treats AttributeDecl like any other decl through generic SubstPrefixAtomsAst.

### BUG 5: ivyAccum missing attributes field

**Python** saves/restores `ivy.attributes` in inst_mod, filters to "common" only.
**Go** has no `attributes` field on ivyAccum at all.

### BUG 6: instMod doesn't propagate attributes for nested InstantiateDecl

**Python** (lines 193-196): `ivy.attributes = ivy.attributes + idecl.attributes`
**Go**: Just calls `doInsts()` without attribute adjustment.

### BUG 7: instMod doesn't copy attributes/common to idecl

**Python** (lines 180-188): Sets `idecl.common = ...` and `idecl.attributes = decl.attributes`.
**Go**: No attribute or common field handling on rewritten declarations.

### BUG 8: createObject doesn't pop stack

**Python** (line 692): `stack.pop()` — removes the module accumulator from the stack.
**Go**: No equivalent (uses parent chain instead, but parent isn't reset).

### BUG 9: SubstPrefixAtomsAst missing variables_distinct_ast

**Python** (ivy_ast.py:1749): `po = variables_distinct_ast(pref,ast) if pref else pref`
**Go** (rewrite.go:675): `po := pref` — no variable renaming to avoid capture.

### BUG 10: instMod spaa ModuleDecl substitution augmentation

**Python** (lines 148-151): When modname is set and decl is ModuleDecl, adds
`child_name(modname) → pref.rep` to subst. Go doesn't do this.

## Fix Plan — Ordered by Impact

### Fix 1: createObject — add Variable prefix args and vsubst (CRITICAL)

**File:** `lalr_full/grammar_v17.y` — `createObject` function (line 168)

Change `createObject` to match Python exactly:

```go
func createObject(top *ivyAccum, name *ast.Atom, objectargs []ast.Node,
    module *ivyAccum, lineno ast.Location, continuation bool) {
    xtracer.Trace("parser.create_object ENTER name=%s", name.Rep)

    // Python: prefargs = [Variable('V'+str(idx),pr.sort) for idx,pr in enumerate(objectargs)]
    var prefargs []ast.Node
    for idx, pr := range objectargs {
        vname := fmt.Sprintf("V%d", idx)
        var sort ast.Node
        if a, ok := pr.(*ast.Atom); ok {
            sort = a.Sort  // pr.sort in Python
        }
        prefargs = append(prefargs, ast.NewVariable(vname, sort))
    }

    // Python: pref = Atom(name, prefargs)
    pref := ast.NewAtom(name.Rep, prefargs...)  // or set Terms after
    pref.SetLineno(lineno)

    if !continuation {
        top.declare(ast.NewObjectDecl(pref))
        setObjectDefined(top, name.Rep)
    }

    // Python: vsubst = dict((pr.rep,v) for pr,v in zip(objectargs,prefargs))
    vsubst := make(map[string]*ast.Variable)
    for i, pr := range objectargs {
        if i < len(prefargs) {
            prName := nodeRep(pr)
            if v, ok := prefargs[i].(*ast.Variable); ok {
                vsubst[prName] = v
            }
        }
    }

    instMod(top, module.decls, pref, map[string]string{}, vsubst, "")
    xtracer.Trace("parser.create_object EXIT name=%s", name.Rep)
}
```

**Dependencies:** Need to verify `ast.NewAtom` can accept Terms args, and
`ast.NewVariable(name, sort)` exists. Check `ast/node.go` or `ast/logic.go`.

### Fix 2: objectargs rule — set params on accum

**File:** `lalr_full/grammar_v17.y` — `objectargs` rule (line 2651)

Add `v17lex.(*v17LexAdapter).accum.params = $1` matching Python's
`stack[-1].params = p[0]`.

### Fix 3: instMod — add attributes field to ivyAccum and handle common

**File:** `lalr_full/ivy_module.go` — ivyAccum struct

Add `attributes []string` field. Initialize to empty in `newIvyAccum`.

**File:** `lalr_full/inst_mod.go` — instMod function

Add at entry:
```go
save := ivy.attributes
ivy.attributes = filterCommon(ivy.attributes) // keep only "common"
defer func() { ivy.attributes = save }()
```

In the declaration loop, handle common:
```go
dpref := pref
if pref != nil && hasAttribute(decl, "common") {
    dpref = pref.Clone(nil) // clone with empty args
}
```

### Fix 4: instMod — AttributeDecl special handling

**File:** `lalr_full/inst_mod.go` — instMod declaration loop

Add AttributeDecl check before the generic path:
```go
if attrDecl, ok := decl.(*ast.AttributeDecl); ok {
    // Python: compose_atoms(dpref, x.args[0]) for each attribute arg
    // Special handling with compose_atoms instead of SubstPrefixAtomsAst
    ...
}
```

### Fix 5: instMod — nested InstantiateDecl attribute propagation

**File:** `lalr_full/inst_mod.go` — declareInstDecl or instMod loop

For InstantiateDecl case:
```go
oldAttrs := ivy.attributes
ivy.attributes = append(ivy.attributes, idecl.Attributes()...)
doInsts(ivy, instDecl.Args())
ivy.attributes = oldAttrs
```

### Fix 6: SubstPrefixAtomsAst — variables_distinct_ast

**File:** `ast/rewrite.go` — SubstPrefixAtomsAst function (line 671)

Implement `VariablesDistinctAst(pref, node)` to rename variables in pref
that also appear in node. This prevents variable capture.

## Critical Files to Modify

1. `~/goivy/lalr_full/grammar_v17.y` — createObject, objectargs rule
2. `~/goivy/lalr_full/inst_mod.go` — instMod, declareInstDecl
3. `~/goivy/lalr_full/ivy_module.go` — ivyAccum (attributes field)
4. `~/goivy/ast/rewrite.go` — SubstPrefixAtomsAst (variables_distinct_ast)

## Execution Order

1. **Fix 1 first** (createObject prefix args + vsubst) — this is the most likely
   root cause of the line 1926 divergence and requires no other changes
2. Run `make golden` to see if we advance past 1926
3. If still stuck, apply Fix 2 (objectargs params)
4. Then Fixes 3-5 (attributes/common handling) as golden test hits them
5. Fix 6 (variables_distinct_ast) only when needed for parameterized objects

## Verification

```bash
cd ~/goivy && make golden
```

Compare the Go xtrace with Python xtrace — should advance past line 1926.
If Fix 1 alone doesn't resolve it, apply subsequent fixes incrementally.
