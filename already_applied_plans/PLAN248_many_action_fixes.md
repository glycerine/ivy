# Conformance Audit + Fixes: applyActuals oldOfOld, Hide/SubstAction/BindOldsUpdate, state_to_action wiring, compile_action_body, NamedUpdate / DerivedUpdate / PatternBasedUpdate dead code, and instantiateMacro

*Created: 2026-04-10 09:30*

## Context

The previous plan implemented fixes for nine `int_update` overrides (Sites 1-9) and explicitly listed five follow-up audit areas as out-of-scope. The user has now requested all five follow-ups be folded into a single plan and executed sequentially:

1. `applyActuals` deeper than the silent-NullUpdate paths — specifically the `oldOfOld` substitution at update.go:1918-1919 and the broader question of whether it matches Python's `subst[old(s)] = old(t)` for compound `old` symbols.
2. `Hide`, `SubstAction`, `BindOldsUpdate` — both the **annotation** pipelines AND the helpers themselves (modified set, formula transformation, ExistQuantClauses vs Python `exist_quant`).
3. Porting `state_to_action(v.value)` for non-Action `Domain.Actions` entries.
4. `compile_action_body` callback wiring used by `InstantiateAction`.
5. `instantiateMacro` audit + audit of `NamedUpdate`, `DerivedUpdate`, and `PatternBasedUpdate.dependencies.symbols` lookup.

All audits below were performed sequentially against the Python source of truth (`/Users/jaten/ivy/pyivy/ivy/ivy/`) and the Go port (`/Users/jaten/go/src/github.com/glycerine/ivy/goivy/`), reading the actual code line-by-line.

**Headline result**: Multiple real bugs found across follow-ups 1, 3, 4, and 5. Follow-up 2 (helpers + annotations) is verified conformant. Follow-up 5 surfaces a major dead-code situation: Go's `PatternBasedUpdate` and `NamedUpdate` are defined but never constructed by the parser/compiler.

## Audit results

### Area 1 — `applyActuals` oldOfOld (CRITICAL CONFORMANCE BUG)

**Python (`ivy_actions.py:1363-1365`):**
```python
subst = distinct_obj_renaming(v.formal_params+v.formal_returns,vocab)
for s,t in list(subst.items()):
    subst[old(s)] = old(t)
```

`old(sym)` is `ivy_transrel.py:58-62`:
```python
def old(sym):
    return sym.prefix('old_')
```

So Python adds `subst[old_x] = old_y` entries — keys named `old_x` that DO appear in callee bodies (from prior `BindOldsAction` wrappers or `old(x)` references inside the callee body).

**Go (`/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/update.go:1914-1920`):**
```go
substMap := make(map[lg.NodeKey]lg.Expr)
for oldSym, newSym := range renaming {
    substMap[lg.Key(oldSym)] = newSym
    // Also map old(s) → old(t) for pre-state symbols
    oldOfOld := lg.NewConst("old("+oldSym.Name+")", oldSym.CSort)
    substMap[lg.Key(oldOfOld)] = lg.NewConst("old("+newSym.Name+")", newSym.CSort)
}
```

**Bug**: Go constructs symbol names with literal parentheses — `"old(x)"`, `"old(y)"` — but no symbol with that name exists in any callee body. Pre-state symbols use the `old_` prefix (per the existing `Old()` helper at `transrel.go:58-60`):
```go
func Old(name string) string { return "old_" + name }
```

The substMap entries built by the buggy code key on names that nothing matches, so the entire Python `for s,t in subst.items(): subst[old(s)] = old(t)` loop is silently a no-op in Go. When a callee body contains a real `old_x` symbol, Go fails to rename it for capture avoidance, while Python correctly renames it to `old_y`.

**Severity**: Real conformance failure affecting any call site whose callee references `old(...)` of a parameter.

### Area 2 — `Hide` / `SubstAction` / `BindOldsUpdate`: helpers AND annotations (NO BUGS)

**Verified by reading both Python (`ivy_transrel.py:236-243, 397-413, 437-445`) and Go (`actions/transrel.go:945-1015, 1602-1627, 1632-1669`) line-by-line:**

| Helper | Modified set | Formula transformation | Annotation | Status |
|--------|--------------|------------------------|------------|--------|
| Hide | `[s for s in u[0] if s not in syms]` matches Go's `if !symNames[s.Name]` filter | Both call `exist_quant`/`ExistQuantClauses` for TR and Pre using `__`-prefix `UniqueRenamer` | `RenameClauses` correctly invokes `Annot.Rename(exprSubs)` (ops.go:650-658) | ✓ |
| SubstAction | `[subst.get(s,s) for s in u[0]]` matches Go's name-keyed loop | Both build extended subst with `(new(s), new(syms[s]))` for each modified symbol; both rename TR and Pre via `rename_clauses`/`RenameClauses` | Same RenameClauses → Annot.Rename pipeline | ✓ |
| BindOldsUpdate | passed through unchanged | Both rename `old_X` → `X` for every `old_`-prefixed used symbol in TR and Pre | Same pipeline | ✓ |

Specifically verified:
- Hide's `syms.update(new(s) for s in update[0] if s in syms)` is exactly mirrored by Go's `for _, s := range u.Modified { if symNames[s.Name] { syms = append(syms, NewConst(s)); ... } }` (transrel.go:964-972). NewConst returns a Const with the same sort but `new_` prefix.
- ExistQuantClauses uses `iu.NewUniqueRenamer("__", ...)` matching Python's `UniqueRenamer('__', used)` (transrel.go:1025).
- BindOldsClausesClauses iterates `module.UsedSymbolsClauses(clauses)` filtering by `IsOld(s.Name)` and renames via `lg.NewConst(OldOf(s.Name), s.CSort)` — exact mirror of Python's `dict((s, old_of(s)) for s in used_symbols_clauses(clauses) if is_old(s))` (transrel.go:1612-1626).
- All annotation types (`EmptyAnnotation`, `ConjAnnotation`, `ComposeAnnotation`, `RenameAnnotation`, `IteAnnotation`) implement `Rename` by wrapping in a `RenameAnnotation`, matching Python's `rename_annot` (ivy_actions.py:1659-1661).

**Caveat (not a bug)**: Go uses name-based hashing where Python uses Symbol identity. For canonical symbols this is equivalent. There's also dead code: `BindOldsClauses(node lg.Expr) lg.Expr` (transrel.go:1586) is a formula-level helper used only by tests; its existence parallel to `BindOldsClausesClauses` is awkward but harmless. **No fix needed for Area 2.**

### Area 3 — `state_to_action` porting (PORTED, BUT TWO CALL SITES MISSING; non-Action callee path is unreachable in Go's typed model)

**Status of `StateToAction` itself**: fully ported at `actions/transrel.go:1115-1134`, with all dependencies (`New`/`NewConst`, `IsOld`, `OldOf`, `module.UsedSymbolsClauses`, `module.RenameClauses`) present.

**Missing call site 3a — `PatternBasedUpdate.GetUpdateAxioms`** (action.go:1881-1929):

Python (ivy_actions.py:152-164):
```python
def get_update_axioms(self,updated,action):
    for x in updated:
        if x in self.dependencies.symbols:
            updated = updated + [y for y in self.defines.symbols if y not in updated]
            try:
                precond,postcond = next(y for y in (p.match(action) for p in self.patterns.args) if y != None)
            except StopIteration:
                raise IvyError(action,'No matching update axiom for ' + str(x))
            postcond = state_to_action((updated,postcond,precond))      # <-- MISSING IN GO
            return (updated,postcond[1],precond)
    return (updated,true_clauses(),false_clauses())
```

Go returns `transrel` raw without the `state_to_action` rename, AND silently returns `(updated, TrueClauses, FalseClauses)` if no pattern matches (Python panics with `IvyError`). Both are bugs.

Note: even though Go's `PatternBasedUpdate` is currently dead code (see Area 5, follow-up 5b below), we still want this method correct so the dead-code wake-up actually works.

**Call site 3b — `CallAction.IntUpdate` non-Action callee path** (update.go:1849-1868):

Python (ivy_actions.py:1341-1354):
```python
def int_update(self,domain,pvars):
    v = self.get_callee()
    if not isinstance(v,tuple):
        if isinstance(v,Action):
            v = self.apply_actuals(domain,pvars,v)
        else:
            v = state_to_action(v.value)   # <-- non-Action with .value attribute
    return v
```

`context.get(name)` (ivy_actions.py:66-67) calls `ivy_module.find_action(symbol)` which can return tuples or non-Action objects with `.value`.

**Verification**: Go's `module.Module.Actions` is typed `*iu.InsMap[string, Action]` (module.go:38). The Go type system **prevents** non-Action values from ever being stored. Therefore the `tuple` and `state_to_action(v.value)` branches are unreachable in Go's typed model.

**Decision**: keep the panic at update.go:1854-1859. Update the comment to cite the `InsMap[string, Action]` typing as the reason. **Do not port** the `state_to_action(v.value)` path. If a future change ever loosens the type, the panic will surface immediately and the port can happen then.

### Area 4 — `compile_action_body` callback wiring (SILENT ERROR + MISSING CALLBACK FALLTHROUGH)

Callback registration at `compiler/ivy_compile.go:265-270` is correct (sets `mod.CompileActionBodyFn` during `IvyCompile`). The callback type and AST→Action dispatch via `Compiler.CompileActionBody` is functionally equivalent to Python's monkey-patched `im.compile()` chain.

**Bugs in `InstantiateAction.IntUpdate` (action.go:2154-2179)**:

```go
if compileFn != nil {
    compiled, err := compileFn(rewritten)
    if err == nil && compiled != nil {
        return IntUpdate(compiled, ctx)
    }
    // SILENT FALLTHROUGH on err != nil or compiled == nil
}
```

- (4-i) Silent error swallowing: a `compileFn` failure on a successfully matched macro falls through to the schemata check, which then panics with the misleading "instantiation of undefined" message. Python would propagate the exception immediately.
- (4-ii) `compileFn == nil` (when neither `ctx.CompileActionBody` nor `ctx.Domain.CompileActionBodyFn` is set) also falls through silently. Python's `.compile()` is monkey-patched at startup so it's always available; the Go callback should always be registered. If it's missing when a macro matched, that's a programming error in callback registration and should panic.

### Area 5 — Update protocol (DerivedUpdate / NamedUpdate / PatternBasedUpdate) and `instantiateMacro`

#### 5a — DerivedUpdate.GetUpdateAxioms (CONFORMANT)

Python (`ivy_actions.py:166-176`):
```python
class DerivedUpdate(object):
    def __init__(self,defn):
        self.defn = defn
        self.dependencies = used_symbols_ast(defn.args[1])  # RHS only
    def get_update_axioms(self,updated,action):
        defines = self.defn.args[0].rep
        if defines not in updated and any(x in self.dependencies for x in updated):
            updated.append(defines)
        return (updated,None,None)
```

Go (`module/derivedupdate.go:45-70`): collects deps from `a.Defn` (the entire Definition, not just RHS). This includes the LHS symbol in deps. **Functionally equivalent**: the GetUpdateAxioms guard `if !updatedSet[defSym.Name]` ensures we only enter the body when the defined symbol is NOT in `updated`, so the spurious LHS-in-deps entry can't trigger a self-add. **No fix needed for 5a.**

#### 5b — NamedUpdate dead code + sort loss in GetUpdateAxioms (BUGS)

Python (`ivy_actions.py:178-186`):
```python
class NamedUpdate(object):
    def __init__(self,sym,fmla):
        self.sym = sym                           # Symbol with sort
        self.dependencies = used_symbols_ast(fmla)
    def get_update_axioms(self,updated,action):
        defines = self.sym
        if defines not in updated and any(x in self.dependencies for x in updated):
            updated.append(defines)              # appends Symbol with sort
        return (updated,None,None)
```

Constructed at `ivy_compiler.py:1237` inside `def named`:
```python
sym = self.domain.sig.add_symbol(lhs.rep,ivy_logic.FuncConstSort(*(dom+[rng])))
self.domain.named.append((self.last_fact,sym(*targs) if targs else sym))
self.domain.updates.append(NamedUpdate(sym,cond))    # <-- KEY: sym carries sort, cond is the formula
```

Go (`actions/action.go:1934-1991`):
```go
type NamedUpdate struct {
    ActionBase
    UpdateName string                            // <-- string, no sort
    Body       lg.Expr                           // <-- the formula
}

func NewNamedUpdate(name string, body lg.Expr) *NamedUpdate {  // <-- never called from anywhere
    return &NamedUpdate{UpdateName: name, Body: body}
}

func (a *NamedUpdate) GetUpdateAxioms(updated []*lg.Const, action Action) ([]*lg.Const, *module.Clauses, *module.Clauses) {
    defines := a.UpdateName
    if defines == "" {
        return updated, nil, nil                 // <-- silent fallback
    }
    deps := make(map[string]bool)
    module.CollectSymNames(a.Body, deps)
    updatedSet := make(map[string]bool)
    for _, u := range updated {
        updatedSet[u.Name] = true
    }
    if !updatedSet[defines] {
        for _, u := range updated {
            if deps[u.Name] {
                sym := lg.NewConst(defines, lg.TopS)             // <-- WRONG SORT (TopS fallback)
                if c, ok := a.Body.(*lg.Const); ok {
                    sym = lg.NewConst(defines, c.CSort)          // <-- still wrong: a.Body is the FORMULA, not the symbol
                }
                updated = append(updated, sym)
                break
            }
        }
    }
    return updated, nil, nil
}
```

**Bugs**:
- (5b-i) **Dead code**: `NewNamedUpdate` is defined but never called. `compiler/decl.go:1424-1478` (`DomainSetup.Named`) does NOT append a `NamedUpdate` to `mod.Updates`, while Python's `def named` (ivy_compiler.py:1237) does.
- (5b-ii) **Sort loss**: Go stores only `UpdateName` (string), losing the `FuncConstSort` that Python keeps via `self.sym`. The reconstruction in `GetUpdateAxioms` uses `lg.TopS` or `a.Body.(*lg.Const).CSort` (which is wrong because `a.Body` is the formula `cond`, not the symbol).
- (5b-iii) **Silent fallback** on empty `UpdateName` should panic (Python would AttributeError on missing `self.sym`).

#### 5c — PatternBasedUpdate dead code at parser AND compiler (BUGS)

Python parser (`ivy_parser.py:1737-1745`):
```python
def p_top_update_terms_from_terms_upaxes(p):
    'top : top UPDATE apps FROM apps upaxes'
    p[0] = p[1]
    dfns = [x.rep for x in p[3]]
    deps = [x.rep for x in p[5]]
    p[0].declare(UpdateDecl(PatternBasedUpdate(SymbolList(*dfns),
                                               SymbolList(*deps),
                                               UpdatePatternList(*p[6]))))
```

Go parser (`parser/grammar_v17.y:1196-1204`):
```go
| top TOK_UPDATE apps TOK_FROM apps upaxes
{
    xtracer.Trace("parser.p_top_update_terms_from_terms_upaxes ENTER (top)")
    $$ = $1
    // Simplified: store as raw nodes
    _ = $3
    _ = $5
    _ = $6
}
```

The Go parser drops `$3`, `$5`, `$6` entirely. No `UpdateDecl` is constructed. So `update X from Y { ... }` declarations parse silently as no-ops.

**Compiler path**: even if the parser were fixed, Go's `compiler/phase6.go` `Thing → CompileNode` has no case for `*ast.PatternBasedUpdate`. `DomainSetup.Update` (decl.go:1658) calls `Compiler.Thing(node)` which falls through to default handling.

**Bugs**:
- (5c-i) **Parser stub**: builds no AST node for `update X from Y { ... }` declarations.
- (5c-ii) **Compiler missing case**: no `CompileNode` case to convert `*ast.PatternBasedUpdate` → `*actions.PatternBasedUpdate` (which would build the symbol-typed `Defines`/`Dependencies` slices and the compiled `UpdatePatternList`).
- (5c-iii) Already documented in Area 3: `GetUpdateAxioms` is missing `state_to_action` and panics on no-match.

#### 5d — `instantiateMacro` (BUGS — silent fallthroughs on type mismatches)

Python (`ivy_actions.py:783-796`):
```python
def instantiate_macro(inst,defns):
    if inst.relname in defns:
        defn = defns[inst.relname]
        aparams = inst.args
        fparams = defn.args[0].args
        if len(aparams) != len(fparams):
            raise IvyError(inst,"wrong number of parameters");
        subst = dict((x.rep,y) for x,y in zip(fparams,aparams))
        psubst = dict((x.rep,y.rep) for x,y in zip(fparams,aparams)
                      if (isinstance(y,ivy_ast.App) or isinstance(y,ivy_ast.Atom)) and
                      len(y.args) == 0)
        return ivy_ast.ast_rewrite(defn.args[1],ivy_ast.AstRewriteSubstConstantsParams(subst,psubst))
    return None
```

Go (`actions/action.go:2246-2317`) silently:
- (5d-i) Returns nil on non-Atom/Symbol `astInst` (lines 2250-2259). Python would AttributeError on `inst.relname` / `inst.args`.
- (5d-ii) Leaves `fparams` as nil if `defn.Lhs` is not `*ast.Atom` (lines 2267-2270). Python uses `defn.args[0].args` which would AttributeError. The downstream count check then either passes (if aparams is also empty) or panics with the wrong message.
- (5d-iii) Skips entries in the subst-building loop if `fp` is neither `*ast.Atom` nor `*ast.Symbol` (lines 2278-2285). Python iterates without type checking.

These should panic with descriptive messages so any malformed macro definition or call site surfaces immediately.

#### 5e — PatternBasedUpdate.dependencies lookup (CONFORMANT)

The Go check at `action.go:1887-1902` uses string-based name comparison via `depSet[u.Name]`, while Python uses `x in self.dependencies.symbols` which compares Symbol objects. For canonical symbols these are functionally equivalent. **No fix needed.**

## Files to modify

### Group A — Direct conformance fixes (Areas 1, 3a, 4, 5d)

#### 1. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/update.go`

**Fix 1 (Area 1) — applyActuals oldOfOld at lines 1914-1920:**
```go
substMap := make(map[lg.NodeKey]lg.Expr)
for oldSym, newSym := range renaming {
    substMap[lg.Key(oldSym)] = newSym
    // Python (ivy_actions.py:1364-1365): subst[old(s)] = old(t)
    //   where old(sym) = sym.prefix('old_'). Use the Old() helper here so
    //   the substMap key actually matches a real pre-state symbol in the
    //   callee body (e.g., from a prior BindOldsAction wrapper).
    oldOfOldSym := lg.NewConst(Old(oldSym.Name), oldSym.CSort)
    oldOfNewSym := lg.NewConst(Old(newSym.Name), newSym.CSort)
    substMap[lg.Key(oldOfOldSym)] = oldOfNewSym
}
```

**Fix 3b (Area 3) — comment update only at lines 1854-1859:**

Replace the existing panic comment with a citation of the `InsMap[string, Action]` typing:
```go
if !isAct {
    // Python (ivy_actions.py:1350): v = state_to_action(v.value)
    //   for non-Action context entries (state-style updates with .value).
    // Go's module.Module.Actions is typed *iu.InsMap[string, Action]
    // (module.go:38), so the type system prevents non-Action values from
    // ever being stored. The state_to_action(v.value) branch is unreachable
    // in Go's typed model. If this panic ever fires, the upstream
    // construction site is violating the type contract — investigate that,
    // do not port state_to_action(v.value).
    panic(fmt.Sprintf("CallAction.IntUpdate: callee %s resolved to non-Action %T (Domain.Actions is typed Action; non-Action storage is a programming error)", calleeName, v))
}
```

#### 2. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/action.go`

**Fix 3a (Area 3) — PatternBasedUpdate.GetUpdateAxioms at lines 1916-1928:**
```go
if a.Patterns != nil {
    for _, pat := range a.Patterns.Patterns {
        precond, transrel := pat.Match(action)
        if precond != nil && transrel != nil {
            // Python (ivy_actions.py:161):
            //   postcond = state_to_action((updated, postcond, precond))
            //   return (updated, postcond[1], precond)
            stateUpdate := &Update{
                Modified: updated,
                TR:       transrel,
                Pre:      precond,
            }
            actionUpdate := StateToAction(stateUpdate)
            return updated, actionUpdate.TR, precond
        }
    }
    // Python (ivy_actions.py:159-160):
    //   raise IvyError(action, 'No matching update axiom for ' + str(x))
    panic(fmt.Sprintf("PatternBasedUpdate.GetUpdateAxioms: no matching update axiom for %v", a.Defines))
}
// patterns is nil but a dep was hit — also a construction-site error.
panic(fmt.Sprintf("PatternBasedUpdate.GetUpdateAxioms: dependency hit but no patterns: %v", a.Defines))
```

The early `if !found { return updated, TrueClauses, FalseClauses }` at lines 1904-1906 IS conformant (matches Python's `return (updated,true_clauses(),false_clauses())` at line 164) — do not touch it.

**Fix 4 (Area 4) — InstantiateAction.IntUpdate macro path at lines 2154-2179:**
```go
if rewritten := instantiateMacro(a.AstInst, ctx.Domain.Macros); rewritten != nil {
    // Python (ivy_actions.py:812): res = im.compile().int_update(domain, pvars)
    // The monkey-patched .compile() is always available; the Go equivalent
    // is ctx.CompileActionBody (test override) or ctx.Domain.CompileActionBodyFn
    // (registered by ivy_compile.go:265 during IvyCompile).
    compileFn := ctx.CompileActionBody
    if compileFn == nil && ctx.Domain.CompileActionBodyFn != nil {
        moduleFn := ctx.Domain.CompileActionBodyFn
        compileFn = func(node ast.Node) (Action, error) {
            result, err := moduleFn(node)
            if err != nil {
                return nil, err
            }
            if act, ok := result.(Action); ok {
                return act, nil
            }
            return nil, fmt.Errorf("CompileActionBodyFn returned non-Action type %T", result)
        }
    }
    if compileFn == nil {
        // Python: .compile() is monkey-patched at startup and always present.
        // If neither hook is set, callback registration is broken upstream.
        panic("InstantiateAction.IntUpdate: macro matched but no CompileActionBody hook is registered")
    }
    compiled, err := compileFn(rewritten)
    if err != nil {
        // Python would propagate the exception from im.compile().
        panic(fmt.Sprintf("InstantiateAction.IntUpdate: compile failed for macro expansion: %v", err))
    }
    if compiled == nil {
        panic("InstantiateAction.IntUpdate: compileFn returned (nil, nil) for macro expansion")
    }
    return IntUpdate(compiled, ctx)
}
```

The schemata check at lines 2199-2216 stays unchanged — only reached when `instantiateMacro` returned nil (no matching macro).

**Fix 5d (Area 5d) — instantiateMacro at lines 2246-2317:**

Replace the silent fallthroughs with descriptive panics:

```go
func instantiateMacro(astInst ast.Node, macros map[string]*ast.Definition) ast.Node {
    // Python (ivy_actions.py:783-784): if inst.relname in defns
    //   would AttributeError on .relname / .args if inst is not Atom-like.
    var name string
    var aparams []ast.Node
    switch n := astInst.(type) {
    case *ast.Atom:
        name = n.Rep
        aparams = n.Terms
    case *ast.Symbol:
        name = n.Rep
        aparams = nil
    default:
        panic(fmt.Sprintf("instantiateMacro: instantiation node is not Atom/Symbol: %T", astInst))
    }

    defn, ok := macros[name]
    if !ok || defn == nil {
        return nil // matches Python "if inst.relname in defns: ... else None"
    }

    // Python (ivy_actions.py:787): fparams = defn.args[0].args
    //   would AttributeError if defn.args[0] is not Atom-like.
    lhs, ok := defn.Lhs.(*ast.Atom)
    if !ok {
        panic(fmt.Sprintf("instantiateMacro: macro %s definition LHS is not Atom: %T", name, defn.Lhs))
    }
    fparams := lhs.Terms

    if len(aparams) != len(fparams) {
        panic(fmt.Sprintf("wrong number of parameters for macro %s", name))
    }

    // Python: subst = dict((x.rep, y) for x, y in zip(fparams, aparams))
    //   x.rep would AttributeError if x is not Atom-like.
    subst := make(map[string]ast.Node)
    for i, fp := range fparams {
        var fpName string
        switch s := fp.(type) {
        case *ast.Atom:
            fpName = s.Rep
        case *ast.Symbol:
            fpName = s.Rep
        default:
            panic(fmt.Sprintf("instantiateMacro: macro %s formal param %d is not Atom/Symbol: %T", name, i, fp))
        }
        subst[fpName] = aparams[i]
    }

    // psubst building (lines 2290-2312) — keep as-is, the type-switch fallthrough
    // is a structural test (only Atom-with-zero-Terms or Symbol qualify), not a
    // silent skip of malformed input. The fparams loop already validated types.
    psubst := make(map[string]string)
    for i, fp := range fparams {
        var fpName string
        switch s := fp.(type) {
        case *ast.Atom:
            fpName = s.Rep
        case *ast.Symbol:
            fpName = s.Rep
        }
        // fpName is guaranteed non-empty by the subst loop's type validation.
        ap := aparams[i]
        switch a := ap.(type) {
        case *ast.Atom:
            if len(a.Terms) == 0 {
                psubst[fpName] = a.Rep
            }
        case *ast.Symbol:
            psubst[fpName] = a.Rep
        }
    }

    rewriter := ast.NewAstRewriteSubstConstantsParams(subst, psubst)
    return ast.AstRewrite(defn.Rhs, rewriter)
}
```

### Group B — NamedUpdate (Area 5b): wire up + sort fix

#### 3. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/action.go`

**Fix 5b-ii (sort loss) — extend NamedUpdate struct at lines 1934-1957:**
```go
type NamedUpdate struct {
    ActionBase
    UpdateName string
    Sym        *lg.Const   // NEW: the named symbol with its actual sort (Python: self.sym)
    Body       lg.Expr     // the formula `cond` (Python: fmla)
}

func NewNamedUpdate(sym *lg.Const, body lg.Expr) *NamedUpdate {
    return &NamedUpdate{UpdateName: sym.Name, Sym: sym, Body: body}
}
```

Update `ActionClone` (line 1946) to copy `Sym`:
```go
func (a *NamedUpdate) ActionClone(args []lg.Expr) Action {
    r := &NamedUpdate{ActionBase: a.ActionBase, UpdateName: a.UpdateName, Sym: a.Sym}
    if len(args) >= 1 {
        r.Body = args[0]
    }
    return r
}
```

**Fix 5b-ii (continued) — NamedUpdate.GetUpdateAxioms at lines 1962-1991:**
```go
func (a *NamedUpdate) GetUpdateAxioms(updated []*lg.Const, action Action) ([]*lg.Const, *module.Clauses, *module.Clauses) {
    if a.Sym == nil {
        // Python: would AttributeError on self.sym. Faithful port panics.
        panic("NamedUpdate.GetUpdateAxioms: Sym is nil")
    }
    defines := a.Sym

    // Python: self.dependencies = used_symbols_ast(fmla) — computed at __init__.
    // Go re-computes each call (acceptable).
    deps := make(map[string]bool)
    module.CollectSymNames(a.Body, deps)

    updatedSet := make(map[string]bool)
    for _, u := range updated {
        updatedSet[u.Name] = true
    }
    if !updatedSet[defines.Name] {
        for _, u := range updated {
            if deps[u.Name] {
                updated = append(updated, defines)
                break
            }
        }
    }
    return updated, nil, nil
}
```

Note: removing `if defines == "" { return updated, nil, nil }` because the new precondition `if a.Sym == nil` covers the construction-error case with a panic.

#### 4. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/decl.go`

**Fix 5b-i (dead code) — DomainSetup.Named at lines 1424-1478:**

Add the missing `NamedUpdate` construction after appending to `mod.Named`. Before the new code, ensure we have access to the `cond` formula (already extracted at line 1444 as `cond := il.DropUniversals(lastFormula)`). After the existing append at line 1473-1476, add:
```go
// Python (ivy_compiler.py:1237): self.domain.updates.append(NamedUpdate(sym, cond))
d.Compiler.Module.Updates = append(d.Compiler.Module.Updates,
    actions.NewNamedUpdate(sym, cond))
```

This requires importing the `actions` package in `compiler/decl.go` (verify it's already imported; if not, add it).

### Group C — PatternBasedUpdate parser + compiler (Area 5c)

#### 5. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/parser/grammar_v17.y`

**Fix 5c-i — UPDATE/FROM rule at lines 1196-1204:**

Replace the stub with the actual AST construction:
```go
| top TOK_UPDATE apps TOK_FROM apps upaxes
{
    xtracer.Trace("parser.p_top_update_terms_from_terms_upaxes ENTER (top)")
    $$ = $1
    cfg := acfg(v17lex)
    // Python (ivy_parser.py:1741-1745):
    //   dfns = [x.rep for x in p[3]]
    //   deps = [x.rep for x in p[5]]
    //   p[0].declare(UpdateDecl(PatternBasedUpdate(SymbolList(*dfns),
    //                                              SymbolList(*deps),
    //                                              UpdatePatternList(*p[6]))))
    dfns := cfg.NewSymbolList(extractAppReps($3)...)
    deps := cfg.NewSymbolList(extractAppReps($5)...)
    pats := cfg.NewUpdatePatternList($6...)
    pbu := cfg.NewPatternBasedUpdate(dfns, deps, pats)
    upd := cfg.NewUpdateDecl(pbu)
    $$.declare(upd)
}
```

Side tasks for this fix:
- Verify the existence of `cfg.NewSymbolList`, `cfg.NewUpdatePatternList`, `cfg.NewUpdateDecl`. The constructors `cfg.NewPatternBasedUpdate` and `*ast.PatternBasedUpdate` already exist (decl_ast.go:2262). Add any missing constructors.
- Add a small Go-side helper `extractAppReps(nodes []ast.Node) []ast.Node` (or similar) that mirrors Python's `[x.rep for x in p[3]]`. Verify the upaxes type is `[]ast.Node` of `*ast.UpdatePattern` matching Python.
- Regenerate `parser/grammar_v17.go` from the `.y` file via the project's standard goyacc invocation. The generated file is committed.

#### 6. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/phase6.go` (or wherever `CompileNode` lives)

**Fix 5c-ii — add CompileNode case for `*ast.PatternBasedUpdate`:**

Add a new case in `CompileNode`'s type switch that:
1. Compiles each symbol in `Dfns.Symbols` to a `*lg.Const` (looking up the symbol in `Compiler.Sig.Symbols` to recover the proper sort).
2. Compiles each symbol in `Deps.Symbols` similarly.
3. Compiles the `UpdatePatternList`'s patterns (via `Compiler.Thing` on each).
4. Constructs and returns `actions.NewPatternBasedUpdate(definesConsts, depsConsts, compiledPatternList)`.

The actions-side constructor at action.go:1862 is:
```go
func NewPatternBasedUpdate(defines, deps []*lg.Const, patterns *UpdatePatternList) *PatternBasedUpdate
```

Implementation note: Python's `.compile()` chain on `PatternBasedUpdate` likely walks all three args; Go's `CompileNode` case should do the same (compile children, then construct the actions-level object).

Verify `actions.UpdatePatternList` has the constructor needed; if not, add one matching the Python protocol.

### Group D — Verification

After all fixes:

1. Build:
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go build ./...
   ```

2. Run actions package tests:
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./actions/
   ```

3. Run compiler package tests (Group B/C may surface upstream construction bugs):
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./compiler/
   ```

4. Run parser package tests (Group C requires goyacc regen):
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./parser/
   ```

5. Run the failing golden test:
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy
   go test ./parser/ -run TestOrdLive -v 2>&1 | tee ~/ivy/goivy/log.red.new
   ```
   Confirm divergence behavior:
   - If shifts further into trace: progress.
   - If remains at i=233622 unchanged: fixes don't touch the divergence path; still valid conformance work.
   - If new panics surface: investigate upstream construction sites.

6. Broader regression:
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./...
   ```

7. **Critical panic checks**: any of the following indicates a previously-hidden upstream bug to investigate (do not revert):
   - `PatternBasedUpdate.GetUpdateAxioms: no matching update axiom for ...`
   - `PatternBasedUpdate.GetUpdateAxioms: dependency hit but no patterns: ...`
   - `InstantiateAction.IntUpdate: macro matched but no CompileActionBody hook is registered`
   - `InstantiateAction.IntUpdate: compile failed for macro expansion: ...`
   - `InstantiateAction.IntUpdate: compileFn returned (nil, nil) for macro expansion`
   - `instantiateMacro: instantiation node is not Atom/Symbol: ...`
   - `instantiateMacro: macro X definition LHS is not Atom: ...`
   - `instantiateMacro: macro X formal param N is not Atom/Symbol: ...`
   - `NamedUpdate.GetUpdateAxioms: Sym is nil`
   - `CallAction.IntUpdate: callee X resolved to non-Action ...`

## Risks / considerations

- **Area 1 (oldOfOld)** is the only one that changes runtime semantics for code that currently compiles silently. Until now, any callee body referencing `old(...)` symbols of formal parameters was silently passed through without renaming. Fixing this may shift TestOrdLive past the current divergence OR cause a *new* divergence (if any test was inadvertently relying on the broken renaming). Both outcomes are real signal — investigate, do not revert.

- **Area 3a (PatternBasedUpdate.GetUpdateAxioms state_to_action wrap)** will only have effect once Area 5c (parser + compiler) wires up PatternBasedUpdate. Until then, it's correctness-on-paper for dead code. That's fine — the fix is still required for the eventual wakeup.

- **Area 4 (compile_action_body)** silent error swallowing is the fix most likely to surface a previously-hidden bug, because it converts a silent fallthrough into a loud panic. If the panic fires in tests, the upstream macro definition or callback registration is at fault — investigate that, do not revert.

- **Area 5b (NamedUpdate dead code)** wire-up may break any test that currently expects `mod.Updates` to NOT contain a NamedUpdate after a `named` declaration. Since `mod.Updates` is currently empty for that path, any test asserting on its emptiness will fail. Search for tests that assert on `mod.Updates` after a `named` declaration before implementing.

- **Area 5c (PatternBasedUpdate parser + compiler)** is the largest change. The parser stub means `update X from Y { ... }` declarations are currently no-ops. Fixing the parser will activate runtime processing for any test that uses these declarations. Risk: golden test trace divergences in tests that exercise these declarations. Search ord_test_live.ivy for `update ... from` to see if TestOrdLive triggers it directly. If not, the impact is limited to other tests.

- **Area 5d (instantiateMacro)** silent-fallthrough fixes may surface upstream parser bugs that produce malformed AST (non-Atom LHS in macro definition, etc.). The new panics will pinpoint the bug clearly.

- **NewNamedUpdate signature change** (from `(name string, body lg.Expr)` to `(sym *lg.Const, body lg.Expr)`) is a breaking change. There are no callers (NewNamedUpdate is dead code), but search for `NewNamedUpdate(` one more time before changing to confirm.

- **Group C generated code**: `parser/grammar_v17.go` is generated from `grammar_v17.y` via goyacc. After editing the `.y`, the `.go` must be regenerated with the project's standard `go generate` (or whatever goyacc invocation the project uses). Verify this is part of the build process before committing.

## Out-of-scope / follow-ups (NOT in this plan)

- Audit of Hide/SubstAction/BindOldsUpdate **callsites** (not the helpers themselves). The audits above verified the helpers are conformant; whether all call sites pass the right arguments is a separate exercise.
- Audit of `subst_both_clauses` (used by `UpdatePattern.match` in Python) and its Go counterpart (likely needed for Group C compile).
- Audit of `ast_rewrite` and `AstRewriteSubstConstantsParams` consistency between Python and Go (used by `instantiateMacro` to rewrite the macro body).
- Cross-check of the `upaxes` parser rule production (the Group C parser fix assumes `$6` is a slice of compiled UpdatePattern nodes; verify against the Python equivalent).
- Audit of `add_definition` callback chain in Python `IvyDomainSetup` versus Go `DomainSetup.DefinitionDecl`.
