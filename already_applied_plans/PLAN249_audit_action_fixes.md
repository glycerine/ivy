# Conformance Audit + Fixes: subst_both_clauses NodeKey API, ast_rewrite Node-typed return, compileUpdatePattern sig.copy(), Hide/SubstAction/BindOldsUpdate callsites, add_definition NativeExpr routing

*Created: 2026-04-10 14:00*

## Context

The previous plan (witty-churning-balloon, completed 2026-04-10) implemented fixes for the original 9 `int_update` overrides and 5 follow-up areas (applyActuals oldOfOld, PatternBasedUpdate.GetUpdateAxioms, InstantiateAction macro silent error, instantiateMacro fallthroughs, NamedUpdate dead code, PatternBasedUpdate parser/compiler wire-up). It explicitly listed 5 NEW follow-up audits as out-of-scope. The user has now requested all 5 of those follow-ups be folded into a new plan and executed sequentially.

The 5 audit areas in this plan:

1. **Hide/SubstAction/BindOldsUpdate callsites** (not the helpers themselves — those were verified conformant in the previous plan)
2. **`subst_both_clauses`** (used by `UpdatePattern.match` in Python) and its Go counterpart
3. **`ast_rewrite` and `AstRewriteSubstConstantsParams`** consistency between Python and Go (used by `instantiateMacro`)
4. **`upaxes` parser rule production** cross-check (the previous plan's Group C parser fix assumed `$6` is a slice of compiled UpdatePattern nodes)
5. **`add_definition` callback chain** in Python `IvyDomainSetup` versus Go `DomainSetup.DefinitionDecl`

All audits were performed sequentially against the Python source of truth (`/Users/jaten/ivy/pyivy/ivy/ivy/`) and the Go port (`/Users/jaten/go/src/github.com/glycerine/ivy/goivy/`), reading the actual code line-by-line. **No parallel agents used** per prior user instruction.

**Headline result**: Bugs found in 4 of 5 areas. Area 1 (Hide/SubstAction/BindOldsUpdate callsites) is verified clean. Area 2 has a sort-handling bug fixable by adopting the codebase's `map[lg.NodeKey]lg.Expr` convention. Area 3 has a structural restriction in the AstRewriter interface. Area 4 has TWO bugs identified: 4-i (parser-side SymbolList element type) is **DEFERRED** for deeper investigation per user direction, and 4-ii (compileUpdatePattern missing `sig.copy()` scoping) is fixed in this plan. Area 5 has two bugs: missing NativeExpr→NativeDefinitions routing and duplicated logic that should be a shared `AddDefinition` helper.

## Audit results

### Area 1 — Hide/SubstAction/BindOldsUpdate callsites (CONFORMANT)

**Python callsites verified:**

| Callsite | Python file:line | Go file:line | Status |
|----------|------------------|--------------|--------|
| `Action.hide_formals` (formals + returns) | ivy_actions.py:231-239 | actions/update.go:2147-2161 (`hideFormals`) | ✓ |
| `LocalAction.int_update` (locals) | ivy_actions.py:1173 | actions/update.go:1775 | ✓ |
| `CallAction.apply_actuals` (formals after binding) | ivy_actions.py:1393 | actions/update.go:2009 | ✓ |
| `LetAction.int_update` (subst_action) | ivy_actions.py:1199 | actions/update.go:1809 | ✓ |
| `BindOldsAction.int_update` (bind_olds_action) | ivy_actions.py:1313 | actions/update.go:1826 | ✓ |
| `Action.update` (outer wrap) | ivy_actions.py:230 | actions/update.go:2139-2144 (`GetUpdate`) | ✓ |
| `EnvAction.int_update` (per-branch update) | ivy_actions.py:923 | actions/update.go:1443 | ✓ |

The wrap order in both is identical: `int_update → bind_olds_action → hide_formals` for the outer `update`/`GetUpdate` entrypoints. Branch-level wrapping in EnvAction calls `update`/`GetUpdate` per branch, also matching.

The empty-list guards differ slightly but not semantically:
- Python `hide_formals`: `if to_hide: update = hide(to_hide, update)`
- Go `hideFormals`: `if len(toHide) > 0 { update = Hide(toHide, update) }`

These are equivalent (Python's truthiness check on a list is `len > 0`).

**No bugs found in Area 1.** All helper callsites match Python.

### Area 2 — `subst_both_clauses` (BUG: sort-mismatched lookup key)

**Python (`ivy_logic_utils.py:1017-1019`):**
```python
def subst_both_clauses(clauses,subst):
    """ apply substitution to both variables and constants in a clause """
    return substitute_constants_clauses(substitute_clauses(clauses,subst),subst)
```

`substitute_constants_ast` (line 183-193):
```python
def substitute_constants_ast(ast,subs):
    if is_constant(ast):  # is_constant → isinstance(ast, lg.Const)
        return subs.get(ast.rep,ast)
```

For Python's Const: `Symbol.rep = property(lambda self: self)` (ivy_logic.py:129) — so `ast.rep` IS the Const itself. The dict lookup `subs.get(Const_obj, default)` uses Python's tuple-based hash/eq from recstruct (utils/recstruct_object.py:62-86): `__eq__` compares `(name, sort)` tuples, `__hash__` is hash of the tuple.

**So Python's lookup is sort-sensitive**: a Const with name "x" and sort A is not equal to a Const with name "x" and sort B.

**Caller — Python `UpdatePattern.match` (ivy_actions.py:122-130):**
```python
def match(self,action):
    subst = dict()
    if self.pattern.match(action,self.placeholders.args,subst):
        axioms_as_clause_sets = (formula_to_clauses(x) for x in (Not(self.precond),self.transrel))
        return (subst_both_clauses(x,subst) for x in axioms_as_clause_sets)
    return None
```

`ast_match` (ivy_logic.py:1220-1242) populates `subst[y] = x` where `y` is the placeholder Const (placeholder's sort) and `x` is the matched term. The dict thus has Const keys with placeholder sorts.

When `substitute_constants_ast` then traverses the formula, the formula's Const has the placeholder's original sort (the formula was built from the placeholder Const). The lookup `subs.get(formula_const, default)` succeeds because `formula_const == placeholder_const` (same name AND sort).

**Go (`module/ops.go:1022-1044`):**
```go
func SubstBothClauses(clauses *Clauses, subs map[string]lg.Expr) *Clauses {
    if clauses == nil || len(subs) == 0 {
        return clauses
    }
    varSubs := make(map[lg.NodeKey]lg.Expr)
    for name, val := range subs {
        v, err := lg.NewVariable(name, val.NodeSort())  // ← uses val.NodeSort()
        if err == nil {
            varSubs[lg.Key(v)] = val
        }
    }
    result := SubstituteClauses(clauses, varSubs)
    constSubs := make(map[lg.NodeKey]lg.Expr, len(subs))
    for name, val := range subs {
        sym := lg.NewConst(name, val.NodeSort())  // ← uses val.NodeSort()
        constSubs[lg.Key(sym)] = val
    }
    result = SubstituteConstantsClauses(result, constSubs)
    return result
}
```

**Caller — Go `UpdatePattern.Match` (actions/action.go:1731-1749):**
```go
subst := make(map[string]lg.Expr)
if !actionMatch(action, p.Pattern, p.Placeholders, subst) {
    return nil, nil
}
precondClauses = module.SubstBothClauses(precondClauses, subst)
```

And `nodeMatch` populates: `subst[pc.Name] = actual` (action.go:1789) — keyed by string name only, throwing away the placeholder's sort.

**Bug 2-i (sort mismatch in lookup key construction)**:

The Go function builds the constant lookup key as `NewConst(name, val.NodeSort())` — using the **substituted value's** sort. But the formula contains the **placeholder's** Const with the **placeholder's** sort. These differ in general (the value can be any expression with any sort). When they differ, `lg.Key` produces a different NodeKey, and `SubstituteConstantsClauses` silently fails to substitute.

Python's lookup-key comes from the placeholder Const stored as the dict key (whose `(name, sort)` tuple-equality matches the formula's Const). Go's lookup-key uses the value's sort, which is wrong.

**Bug 2-ii (string-keyed API discards structural identity)**:

The map type `map[string]lg.Expr` discards the placeholder's structural identity (name + sort). The Go convention for "places where Python uses structural equivalence (which accounts for Sort)" is `map[lg.NodeKey]lg.Expr` where the NodeKey is produced by `lg.Key(node)` (which calls `node.Sexp()`). Examples already in the codebase: `compiler/phase6.go:691,2804,2814,2832,2865`, `l2s/shared.go:385,390,396`, `compiler/ivy_compile.go:1481,1497`.

The downstream functions `SubstituteClauses` and `SubstituteConstantsClauses` (module/ops.go:983, 666) **already** take `map[lg.NodeKey]lg.Expr`. The bug is purely in `SubstBothClauses`'s wrapper signature `map[string]lg.Expr` and its broken key reconstruction logic. The fix is to remove the wrapper's reconstruction step entirely and accept the canonical `map[lg.NodeKey]lg.Expr` directly, letting the caller build keys via `lg.Key(placeholderConst)`.

Note that Python's `subst_both_clauses` runs BOTH substitute_clauses (variables) and substitute_constants_clauses (constants) on the same subst dict. The dict can carry a polymorphic mix of variable-keyed and constant-keyed entries; each pass picks up only its own type. The Go `map[lg.NodeKey]lg.Expr` naturally supports this because `lg.Key(*Variable)` produces `(Variable name:... sort:...)` while `lg.Key(*Const)` produces a Const-prefixed sexp — different structural keys, no collision. Each pass sees only the keys whose construction it can match.

**Caveat**: This entire path is currently only exercised once Area 5c of the previous plan (PatternBasedUpdate parser/compiler wire-up) is actually triggered by a real `update X from Y { ... }` declaration in production. It's still required for correctness when activated.

### Area 3 — `ast_rewrite` and `AstRewriteSubstConstantsParams` (BUG: Atom-only return restricts substitution)

**Python (`ivy_ast.py:1637-1645`):**
```python
class AstRewriteSubstConstantsParams(object):
    def __init__(self,subst,psubst):
        self.subst = subst
        self.psubst = psubst
    def rewrite_name(self,name):
        return subst_subscripts(name,self.psubst)
    def rewrite_atom(self,atom):
        subst = self.subst
        return subst[atom.rep] if not atom.args and atom.rep in subst else atom
```

`rewrite_atom` returns `subst[atom.rep]` which is **whatever Node type is in the dict** — could be App, Atom, Variable, NamedBinder, etc. The Python type system doesn't constrain this.

`ast_rewrite` (ivy_ast.py:1697-1755) handles Atom/App at lines 1706-1717:
```python
if isinstance(x,Atom) or isinstance(x,App):
    if isinstance(x.rep, NamedBinder):
        atom = type(x)(ast_rewrite(x.rep,rewrite),ast_rewrite(x.args,rewrite))
    else:
        atom = type(x)(rewrite.rewrite_name(x.rep),ast_rewrite(x.args,rewrite))
    copy_attributes_ast_ref(x,atom)
    if hasattr(x,'sort'):
        atom.sort = rewrite_sort(rewrite,x.sort)
    if isinstance(x.rep, NamedBinder) or base_name_differs(x.rep,atom.rep):
        return atom
    return rewrite.rewrite_atom(atom)
```

The `return rewrite.rewrite_atom(atom)` returns **whatever the rewriter returns** — possibly a totally different Node type than the original Atom/App.

**Caller — Python `instantiate_macro` (ivy_actions.py:783-795):**
```python
subst = dict((x.rep,y) for x,y in zip(fparams,aparams))
psubst = dict((x.rep,y.rep) for x,y in zip(fparams,aparams)
              if (isinstance(y,ivy_ast.App) or isinstance(y,ivy_ast.Atom)) and
              len(y.args) == 0)
return ivy_ast.ast_rewrite(defn.args[1],ivy_ast.AstRewriteSubstConstantsParams(subst,psubst))
```

The `subst` values are `aparams` — the actual macro call arguments, which can be Apps, Atoms, NamedBinders, etc.

**Go (`ast/rewrite.go:355-379`):**
```go
type AstRewriteSubstConstantsParams struct {
    Subst  map[string]Node
    PSubst map[string]string
}

func (r *AstRewriteSubstConstantsParams) RewriteAtom(atom *Atom, always bool) *Atom {  // ← returns *Atom
    if len(atom.Terms) == 0 {
        if repl, ok := r.Subst[atom.Rep]; ok {
            if a, ok := repl.(*Atom); ok {  // ← TYPE ASSERTION restricts to Atom
                return a
            }
        }
    }
    return atom
}
```

**Bug 3-i (interface restricts return type to *Atom)**:

The `AstRewriter` interface's `RewriteAtom` returns `*Atom`. This means even if a rewriter wanted to substitute with an App or Variable, it cannot: the type system forbids it.

**Bug 3-ii (type assertion silently drops non-Atom substitutions)**:

`AstRewriteSubstConstantsParams.RewriteAtom` does `if a, ok := repl.(*Atom); ok` and only returns the replacement when the cast succeeds. For App, Variable, NamedBinder substitutions, the cast fails silently and the original placeholder atom is returned unchanged. **The macro expansion silently fails to substitute.**

`AstRewriteSubstConstants.RewriteAtom` (rewrite.go:344-353) has the **same bug**.

**Bug 3-iii (AstRewrite ignores polymorphic return)**:

The `case *Atom` and `case *App` in `AstRewrite` (rewrite.go:578-660) pass the result of `RewriteAtom` through as if it were the same type. They cannot handle a polymorphic Node return because the interface signature forbids it.

**Fix scope**: Change the `AstRewriter` interface so `RewriteAtom` returns `Node` instead of `*Atom`. Update all callers (the Atom/App branches of `AstRewrite`, the App-internal-to-Atom-conversion path, and the per-rewriter implementations: `AstRewriteSubstConstants`, `AstRewriteSubstConstantsParams`, `AstRewriteSubstPrefix`, `AstRewritePostfix`, `AstRewriteAddParams`).

The implementations that already return Atom can wrap their return in the polymorphic `Node` interface (no-op cast). Only the two `Subst*` rewriters need to actually return non-Atom nodes.

### Area 4 — `upaxes` parser rule production (BUGS: SymbolList element type + missing sig.copy())

**Python parser (`ivy_parser.py:1737-1745`):**
```python
def p_top_update_terms_from_terms_upaxes(p):
    'top : top UPDATE apps FROM apps upaxes'
    p[0] = p[1]
    dfns = [x.rep for x in p[3]]   # ← extract .rep STRINGS from each App
    deps = [x.rep for x in p[5]]   # ← extract .rep STRINGS
    p[0].declare(UpdateDecl(PatternBasedUpdate(SymbolList(*dfns),
                                               SymbolList(*deps),
                                               UpdatePatternList(*p[6]))))
```

`p[3]` and `p[5]` are `apps` — lists of `App` nodes (from `ivy_logic_parser.py:354-363`). The list comprehension `[x.rep for x in p[3]]` extracts **string names** (Python `App.rep` is a string assigned in `__init__`).

`SymbolList.__init__` (ivy_actions.py:99-102):
```python
def __init__(self,*symbols):
    assert all(isinstance(a,str) or isinstance(a,Symbol) for a in symbols)
    self.symbols = symbols
    self.args = symbols
```

Asserts strings or Symbols. At parse time, **strings** are passed.

`SymbolList.cmpl` (ivy_compiler.py:491):
```python
SymbolList.cmpl = lambda self: self.clone([find_symbol(s) for s in self.symbols])
```

The compiler walks the strings and resolves each via `find_symbol(s)` (ivy_logic.py:346-353), returning Symbols with proper sorts.

**Go parser (`parser/grammar_v17.y:1196-1213`):**
```go
| top TOK_UPDATE apps TOK_FROM apps upaxes
{
    xtracer.Trace("parser.p_top_update_terms_from_terms_upaxes ENTER (top)")
    $$ = $1
    cfg := acfg(v17lex)
    dfns := cfg.NewSymbolList($3...)  // ← passes []ast.Node (Apps), NOT strings
    deps := cfg.NewSymbolList($5...)
    pats := cfg.NewUpdatePatternList($6...)
    pbu := cfg.NewPatternBasedUpdate(dfns, deps, pats)
    upd := cfg.NewUpdateDecl(pbu)
    $$.declare(upd)
}
```

`cfg.NewSymbolList(elems ...Node)` (ast/decl_ast.go:2323):
```go
func (cfg *AstConfig) NewSymbolList(elems ...Node) *SymbolList {
    d := &SymbolList{Elems: elems}
    d.Cfg = cfg
    return d
}
```

**Bug 4-i (SymbolList stores Node objects, not strings) — DEFERRED, NOT ADDRESSED IN THIS PLAN**:

Go's parser passes `[]ast.Node` (App nodes) directly into `SymbolList.Elems`. Python stores strings. This diverges:
- Python: `SymbolList.symbols = ("foo", "bar")`
- Go: `SymbolList.Elems = []ast.Node{*App{Rep:"foo",...}, *App{Rep:"bar",...}}`

This is a Canon-level divergence (Python's `_symbollist_canon` shows `elems:["foo" "bar"]` while Go would show `elems:[(app rep:"foo" ...) (app rep:"bar" ...)]`).

The runtime compile path in Go (`compiler/compiler.go:411-419`) extracts `nodeRepStr(elem)` to get the name string and then calls `lookupOrCreateConst(name)`. This works at runtime but is an indirect workaround for storing the wrong data type at parse-time.

**Decision**: Bug 4-i is **deferred for later investigation**. Changing the parser to drop App nodes and store only string-equivalent Symbol nodes would lose potentially important type information (sort annotations on the App, lineno, attribute references) that may be needed by downstream paths not yet audited. A naive parser-side fix risks breaking unidentified downstream consumers.

**This plan does NOT touch the parser for Bug 4-i.** The parser stays as-is, passing App nodes to NewSymbolList. The compile-time `nodeRepStr` workaround continues to operate. Bug 4-i is added to the "Out-of-scope / follow-ups" section below pending a proper deep audit of every consumer of `SymbolList.Elems`.

**Bug 4-ii (compileUpdatePattern doesn't wrap placeholders in sig.copy())**:

**Python (`ivy_compiler.py:528-532`):**
```python
def UpdatePattern_cmpl(self):
    with ivy_logic.sig.copy():
        return ivy_ast.AST.cmpl(self)

UpdatePattern.cmpl = UpdatePattern_cmpl
```

The `with ivy_logic.sig.copy()` context manager creates a temporary signature copy for the duration of compilation. Any new symbols added by `ConstantDecl_cmpl` (via `compile_const → add_symbol`) live only in the copy, then get discarded when the context exits. This prevents the temporary placeholders from polluting the global signature.

**Go (`compiler/compiler.go:456-505`):**
```go
func (c *Compiler) compileUpdatePattern(up *ast.UpdatePattern) (*actions.UpdatePattern, error) {
    var placeholders []lg.Expr
    if up.Params != nil {
        for _, p := range up.Params.Args() {
            compiled, err := c.Thing(p)  // ← compiles into c.Sig directly, no copy/restore
            if err != nil {
                return nil, err
            }
            placeholders = append(placeholders, compiled)
        }
    }
    // ... compile pattern action, requires, ensures (all using the polluted Sig)
}
```

There's no `c.Sig = c.Sig.Copy()` / restore. Placeholder symbols from `Params` permanently pollute `c.Sig.Symbols`, leaking into the rest of compilation.

`compiler/compiler.go:1313-1315` (`compileDefnImpl`) shows the existing pattern for sig.copy():
```go
sigCopy := c.Sig.Copy()
savedSig := c.Sig
c.Sig = sigCopy
// ... compile children using sigCopy ...
c.Sig = savedSig  // restore (typically via defer)
```

`compileUpdatePattern` should use the same pattern to scope the placeholder declarations to the UpdatePattern compile.

### Area 5 — `add_definition` callback chain (BUGS: missing NativeExpr routing + duplicated logic)

**Python (`ivy_compiler.py:1317-1327`):**
```python
def add_definition(self,ldf):
    defs = self.domain.native_definitions if isinstance(ldf.formula.args[1],ivy_ast.NativeExpr) else self.domain.labeled_props
    lhsvs = list(lu.variables_ast(ldf.formula.args[0]))
    for idx,v in enumerate(lhsvs):
        if v in lhsvs[idx+1:]:
            raise IvyError(ldf,"Variable {} occurs twice on left-hand side of definition".format(v))
    for v in lu.used_variables_ast(ldf.formula.args[1]):
        if v not in lhsvs:
            raise IvyError(ldf,"Variable {} occurs free on right-hand side of definition".format(v))
    defs.append(ldf)
    self.last_fact = ldf
```

Called from:
- `derived(self, ldf)` at ivy_compiler.py:1342: `self.add_definition(ldf.clone([label,df]))`
- `definition(self, ldf)` at ivy_compiler.py:1353: `self.add_definition(ldf.clone([label,df]))`

The shared `add_definition` method:
1. **Routes** based on RHS type: NativeExpr → `domain.native_definitions`, else → `domain.labeled_props`
2. Validates LHS no-duplicates
3. Validates RHS no-free-vars
4. Appends to chosen list
5. Sets `last_fact`

**Go (`compiler/decl.go:612-696` Derived, `:701-775` DefinitionDecl):**

Both `Derived` (line 682) and `DefinitionDecl` (line 758) hardcode:
```go
d.Compiler.Module.LabeledProps = append(d.Compiler.Module.LabeledProps, mlf)
d.LastFact = mlf
```

**Bug 5-i (missing NativeExpr→NativeDefinitions routing)**:

Neither `Derived` nor `DefinitionDecl` checks if the RHS is a `*ast.NativeExpr`. Both unconditionally append to `LabeledProps`. Python would route NativeExpr-RHS definitions to `native_definitions` instead.

The Go `Module.NativeDefinitions []*ast.LabeledFormula` field exists (module/module.go:78) and is read by downstream code (compiler/ivy_compile.go:1493, isolate/isolate.go:1303), so the routing simply isn't happening at the population side. NativeExpr-RHS definitions are lost (or rather, misfiled into LabeledProps where they trigger downstream behavior matching the wrong type of definition).

**Bug 5-ii (duplicated logic instead of shared `AddDefinition` method)**:

Python has ONE `add_definition` method shared by `derived()` and `definition()`. Go duplicates the validation + append logic in both `Derived` and `DefinitionDecl`. This violates the mechanical-port rule "If a Python function exists, a Go function must exist with the same name." The fix is to factor out a shared `AddDefinition(ldf *ast.LabeledFormula)` method on `*DomainSetup` that mirrors Python's `add_definition` exactly, including the NativeExpr routing.

The `addDefinitionChecks` helper at decl.go:324-342 already covers the variable validation (steps 2-3 of Python's `add_definition`). The new `AddDefinition` method should call it, then handle the routing + append + LastFact assignment.

## Summary of bugs to fix

| # | Area | Bug | Severity |
|---|------|-----|----------|
| 2-i | subst_both_clauses | Lookup key uses `val.NodeSort()` instead of placeholder's sort | Real (silently drops substitutions) |
| 2-ii | subst_both_clauses | API `map[string]lg.Expr` discards structural identity; should be `map[lg.NodeKey]lg.Expr` (codebase convention) | Structural |
| 3-i | ast_rewrite | `RewriteAtom` interface returns `*Atom`, restricting substitution | Real (silently drops non-Atom substitutions) |
| 3-ii | AstRewriteSubstConstantsParams | Type assertion `repl.(*Atom)` silently drops App/Variable replacements | Real |
| 3-iii | AstRewriteSubstConstants | Same type-assertion bug as 3-ii | Real |
| ~~4-i~~ | ~~upaxes parser~~ | ~~SymbolList stores Node objects~~ | **DEFERRED** — see Out-of-scope |
| 4-ii | compileUpdatePattern | Missing `sig.copy()` scope around placeholder declarations | Real (pollutes global Sig) |
| 5-i | Derived/DefinitionDecl | Missing NativeExpr→NativeDefinitions routing | Real |
| 5-ii | Derived/DefinitionDecl | Duplicated logic instead of shared `AddDefinition` method | Non-faithful port |

Area 1 (Hide/SubstAction/BindOldsUpdate callsites) had no bugs.

## Files to modify

### Group A — `subst_both_clauses` and `UpdatePattern.Match` (Area 2)

#### A1. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/module/ops.go`

Change `SubstBothClauses` to use the codebase's standard `map[lg.NodeKey]lg.Expr` convention (matching `SubstituteClauses` and `SubstituteConstantsClauses` at module/ops.go:983 and :666):

```go
// SubstBothClauses applies substitution to both variables and constants in clauses.
// Python: subst_both_clauses (ivy_logic_utils.py:1017-1019):
//   substitute_constants_clauses(substitute_clauses(clauses, subst), subst)
//
// Both passes share the same subs map. Python's dict can hold a polymorphic
// mix of variable-keyed (Variable.rep is a string name) and constant-keyed
// (Const.rep is the Const itself, with sort-sensitive __eq__) entries; each
// pass picks up only its own type via different .rep semantics.
//
// The Go convention for "places where Python uses structural equivalence
// (which accounts for Sort)" is map[lg.NodeKey]lg.Expr keyed by lg.Key(node),
// which calls node.Sexp(). Variable.Sexp() and Const.Sexp() produce
// type-prefixed sexps that naturally segregate the two pass populations:
// the variable pass only matches keys built from Variables, the constant pass
// only matches keys built from Consts.
//
// The caller (UpdatePattern.Match) builds the map with lg.Key(placeholderConst)
// keys — preserving the placeholder's actual sort (matching Python's
// "subst[y] = x" where y is the placeholder Symbol with its sort).
func SubstBothClauses(clauses *Clauses, subs map[lg.NodeKey]lg.Expr) *Clauses {
    if clauses == nil || len(subs) == 0 {
        return clauses
    }
    // Both passes share the same subs map. SubstituteClauses iterates Variables
    // in the formula and looks them up by lg.Key(var); SubstituteConstantsClauses
    // iterates Consts and looks them up by lg.Key(const). The two key namespaces
    // don't collide because Variable.Sexp() and Const.Sexp() have different
    // type prefixes.
    result := SubstituteClauses(clauses, subs)
    result = SubstituteConstantsClauses(result, subs)
    return result
}
```

The function shrinks dramatically because the wrapping/key-reconstruction logic was the bug source — the underlying functions already do the right thing.

#### A2. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/action.go`

Change `UpdatePattern.Match`, `actionMatch`, and `nodeMatch` to use `map[lg.NodeKey]lg.Expr` keyed by the placeholder's structural identity:

```go
func (p *UpdatePattern) Match(action Action) (*module.Clauses, *module.Clauses) {
    if p.Pattern == nil {
        return nil, nil
    }
    // Python (ivy_actions.py:122-130): subst dict is populated by ast_match,
    // which sets subst[placeholder_symbol] = matched_term. The key is the
    // placeholder Symbol object whose __eq__/__hash__ compare (name, sort).
    // Go convention: map[lg.NodeKey]lg.Expr where the NodeKey is built from
    // the placeholder Const via lg.Key (which calls node.Sexp()).
    subst := make(map[lg.NodeKey]lg.Expr)
    if !actionMatch(action, p.Pattern, p.Placeholders, subst) {
        return nil, nil
    }
    precondFmla := &lg.Not{Body: p.Precond}
    precondClauses := module.FormulaToClauses(precondFmla, nil)
    precondClauses = module.SubstBothClauses(precondClauses, subst)
    transrelClauses := module.FormulaToClauses(p.TransRel, nil)
    transrelClauses = module.SubstBothClauses(transrelClauses, subst)
    return precondClauses, transrelClauses
}
```

Update `actionMatch` and `nodeMatch` signatures from `map[string]lg.Expr` to `map[lg.NodeKey]lg.Expr`. In `nodeMatch`, the placeholder-binding step becomes:
```go
if pc, ok := pattern.(*lg.Const); ok {
    for _, ph := range placeholders {
        if phc, ok := ph.(*lg.Const); ok && phc.Name == pc.Name {
            // It's a placeholder — bind it. Use lg.Key(phc) so the
            // placeholder's full structural identity (name + sort) is the
            // map key. This matches Python's subst[y] = x where y is the
            // placeholder Symbol.
            phKey := lg.Key(phc)
            if existing, found := subst[phKey]; found {
                return actual.Equal(existing)
            }
            subst[phKey] = actual
            return true
        }
    }
}
```

The check `phc.Name == pc.Name` (matching the formula occurrence by name) stays as-is since the placeholder list is being scanned for membership; once a match is found, the structural key is built from the placeholder Const itself, not from the pattern Const that was found inside the formula.

### Group B — `ast_rewrite` interface refactor (Area 3)

#### B1. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/ast/rewrite.go`

Change the `AstRewriter` interface so `RewriteAtom` returns `Node` instead of `*Atom`:

```go
type AstRewriter interface {
    RewriteName(name string) string
    RewriteAtom(atom *Atom, always bool) Node  // ← was *Atom
}
```

Update all five rewriter implementations:

**`AstRewriteSubstConstants.RewriteAtom`** — return Node directly without type-asserting:
```go
func (r *AstRewriteSubstConstants) RewriteAtom(atom *Atom, always bool) Node {
    if len(atom.Terms) == 0 {
        if repl, ok := r.Subst[atom.Rep]; ok {
            return repl  // any Node type
        }
    }
    return atom
}
```

**`AstRewriteSubstConstantsParams.RewriteAtom`** — same change:
```go
func (r *AstRewriteSubstConstantsParams) RewriteAtom(atom *Atom, always bool) Node {
    if len(atom.Terms) == 0 {
        if repl, ok := r.Subst[atom.Rep]; ok {
            return repl
        }
    }
    return atom
}
```

**`AstRewriteSubstPrefix.RewriteAtom`**, **`AstRewritePostfix.RewriteAtom`**, **`AstRewriteAddParams.RewriteAtom`** — change return type from `*Atom` to `Node`. The body can stay the same since `*Atom` satisfies the `Node` interface; just rename the return type. Verify each callsite isn't depending on the concrete `*Atom` return.

#### B2. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/ast/rewrite.go` (`AstRewrite` function)

Update the `case *Atom` branch (around line 578-597) to handle a `Node` return from `RewriteAtom`:
```go
case *Atom:
    newRep := rewrite.RewriteName(n.Rep)
    newArgs := AstRewriteSlice(n.Terms, rewrite)
    newAtom := &Atom{Rep: newRep, Terms: newArgs}
    newAtom.Cfg = n.Cfg
    CopyAttributesAstRef(n, newAtom)
    if n.ASort != nil {
        sortStr := fmt.Sprint(n.ASort)
        newSortStr := RewriteSort(rewrite, sortStr, n.Cfg)
        ss := &Symbol{Rep: newSortStr}
        ss.Cfg = n.Cfg
        newAtom.ASort = ss
    }
    if BaseNameDiffers(n.Rep, newAtom.Rep) {
        return newAtom
    }
    return rewrite.RewriteAtom(newAtom, false)  // returns Node now
```

Same for `case *App` (line 599-660): the `appAtom` conversion path needs to handle `Node` return:
```go
appAtom := &Atom{Rep: newRep, Terms: newApp.Terms}
appAtom.Base = newApp.Base
appAtom.ASort = newApp.ASort
rewritten := rewrite.RewriteAtom(appAtom, false)
// rewritten is Node now — could be any type
if rAtom, ok := rewritten.(*Atom); ok {
    // converted-to-Atom case (PrefixStr, etc.)
    if rAtom.Rep != newRep {
        // wrap as App-with-Symbol-rep, same as before
        rSym := &Symbol{Rep: rAtom.Rep}
        rSym.Cfg = n.Cfg
        // ... existing logic ...
    } else {
        return newApp
    }
}
// non-Atom replacement (App, Variable, etc.) — return as-is
return rewritten
```

The exact App-branch structure needs reading the existing 80 lines around line 600-680 to preserve the existing semantics (Atom-to-App conversion, sort handling, etc.) while permitting non-Atom returns.

#### B3. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/ast/rewrite.go` — `RewriteSort` and other internal users

`RewriteSort` (around line 690 in the Python equivalent at ivy_ast.py:1690-1695) calls `rewrite.rewrite_atom(Atom(sort)).rep`. The Go equivalent calls `rewrite.RewriteAtom(...)` and accesses `.Rep`. With the new `Node` return type, that field access needs a type switch:
```go
func RewriteSort(rewrite AstRewriter, origSort string, cfg *AstConfig) string {
    sort := rewrite.RewriteName(origSort)
    if BaseNameDiffers(sort, origSort) {
        return sort
    }
    a := &Atom{Rep: sort}
    a.Cfg = cfg
    rewritten := rewrite.RewriteAtom(a, false)
    if rAtom, ok := rewritten.(*Atom); ok {
        return rAtom.Rep
    }
    // Non-Atom rewrite of a sort name is not meaningful — return original.
    return sort
}
```

### Group C — compileUpdatePattern sig.copy() (Area 4 — Bug 4-ii only)

**Note**: Bug 4-i (parser-side SymbolList element type) is deferred. The parser stays as-is. Group C only addresses Bug 4-ii. There is no `parser/grammar_v17.y` change in this group.

#### C1. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go` (compileUpdatePattern)

Add `sig.copy()` scoping around the entire UpdatePattern compile:

```go
func (c *Compiler) compileUpdatePattern(up *ast.UpdatePattern) (*actions.UpdatePattern, error) {
    // Python (ivy_compiler.py:528-532):
    //   def UpdatePattern_cmpl(self):
    //       with ivy_logic.sig.copy():
    //           return ivy_ast.AST.cmpl(self)
    // The placeholder symbols compiled below are temporary; they must NOT
    // pollute the global signature for the rest of compilation.
    sigCopy := c.Sig.Copy()
    savedSig := c.Sig
    c.Sig = sigCopy
    defer func() { c.Sig = savedSig }()

    // Compile placeholders (ConstantDecl → []lg.Expr of *lg.Const)
    var placeholders []lg.Expr
    if up.Params != nil {
        for _, p := range up.Params.Args() {
            compiled, err := c.Thing(p)
            if err != nil {
                return nil, err
            }
            placeholders = append(placeholders, compiled)
        }
    }

    // ... existing compile of pattern action / requires / ensures ...
    // (all using c.Sig which is now sigCopy)

    return &actions.UpdatePattern{
        Placeholders: placeholders,
        Pattern:      patternAction,
        Precond:      precond,
        TransRel:     transrel,
    }, nil
}
```

Note: the returned `placeholders`, `patternAction`, `precond`, `transrel` are all `lg.Expr` / `actions.Action` pointers. They reference the temporary symbols. The temporary symbols themselves are gone from `c.Sig` after the defer, but the returned objects continue to hold references to them. This matches Python: the returned UpdatePattern carries Symbol objects independently of `ivy_logic.sig`.

#### C2. `compilePatternBasedUpdate` — no change required

The previous plan's `compilePatternBasedUpdate` and `nodeRepStr` already handle the App nodes that the parser produces (since Bug 4-i is deferred). No change needed in the compile path.

### Group D — `add_definition` shared method + NativeExpr routing (Area 5)

#### D1. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/decl.go`

Add a shared `AddDefinition` method on `*DomainSetup` mirroring Python's `add_definition`:

```go
// AddDefinition validates and routes a labeled definition to the appropriate
// module list. Corresponds to Python IvyDomainSetup.add_definition
// (ivy_compiler.py:1317-1327).
//
// Routing: if the RHS (defn.Rhs / formula.args[1]) is a *ast.NativeExpr, the
// definition is appended to NativeDefinitions; otherwise to LabeledProps.
//
// Validation matches addDefinitionChecks (no duplicate LHS variables, no free
// RHS variables).
func (d *DomainSetup) AddDefinition(ldf *ast.LabeledFormula) error {
    // Extract the Definition from the LabeledFormula's formula field.
    var defNode *ast.Definition
    switch f := ldf.Formula.(type) {
    case *ast.Definition:
        defNode = f
    case *ast.DefinitionSchema:
        defNode = &f.Definition
    default:
        // Python would AttributeError on ldf.formula.args[1] for non-Definition
        // formulas. Faithful port panics.
        panic(fmt.Sprintf("AddDefinition: ldf.Formula is not a Definition: %T", ldf.Formula))
    }

    // Validate (Python: variable checks at lines 1319-1325)
    if err := addDefinitionChecks(defNode); err != nil {
        return err
    }

    // Python (ivy_compiler.py:1318):
    //   defs = self.domain.native_definitions
    //          if isinstance(ldf.formula.args[1], ivy_ast.NativeExpr)
    //          else self.domain.labeled_props
    if _, isNative := defNode.Rhs.(*ast.NativeExpr); isNative {
        d.Compiler.Module.NativeDefinitions = append(
            d.Compiler.Module.NativeDefinitions, ldf)
    } else {
        d.Compiler.Module.LabeledProps = append(
            d.Compiler.Module.LabeledProps, ldf)
    }
    d.LastFact = ldf
    return nil
}
```

#### D2. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/decl.go` (Derived)

Replace the inline validation + LabeledProps append (lines 634, 681-683) with a call to `AddDefinition`:

```go
// (existing code that compiles the Definition into `compiled` ...)

// Python: self.add_definition(ldf.clone([label, df]))
mlf := lf.Clone([]ast.Node{lf.Label, compiled}).(*ast.LabeledFormula)
if err := d.AddDefinition(mlf); err != nil {
    return err
}
// (existing code: SymbolOrder, AllRelations, Relations, Updates append ...)
```

#### D3. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/decl.go` (DefinitionDecl)

Same change for `DefinitionDecl` (lines 723, 757-759):

```go
// (existing code that compiles the Definition into `compiled` ...)

// Python: self.add_definition(ldf.clone([label, df]))
mlf := lf.Clone([]ast.Node{lf.Label, compiled}).(*ast.LabeledFormula)
if err := d.AddDefinition(mlf); err != nil {
    return err
}
// (existing code: SymbolOrder, Updates append ...)
```

The `addDefinitionChecks` helper at line 324-342 stays — it's still used by `AddDefinition`. The duplicated `addDefinitionChecks` call at line 634 (Derived) and line 723 (DefinitionDecl) can be removed since `AddDefinition` now does it.

### Group E — Verification

After all fixes, run in this order:

1. **Build:**
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go build ./...
   ```
   Group B (ast_rewrite interface refactor) is the highest-risk change — it touches the `AstRewriter` interface used by all rewriters. A clean build verifies no callsites broke.

2. **AST package tests** (Group B):
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./ast/
   ```

3. **Module package tests** (Group A):
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./module/
   ```

4. **Actions package tests** (Group A + macro tests for Group B):
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./actions/
   ```
   The existing `TestSubstAction`, `TestBindOldsClauses`, `TestBindOldsUpdate`, `TestInstantiateMacroNonNodeInst`, and the SubstBoth-related impl tests should still pass.

5. **Parser package tests** — not needed for this plan (no parser changes).

6. **Compiler package tests** (Groups C + D):
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./compiler/
   ```
   The `TestCompileNativeDef_FieldBasedDecision` and existing definition tests should still pass. Watch for new failures around `LabeledProps` vs `NativeDefinitions` count assertions — those would indicate pre-existing tests that depended on the (broken) routing.

7. **Broader regression:**
   ```
   cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go test ./...
   ```

8. **Critical panic checks** — any of these indicates a previously-hidden upstream construction bug surfaced by the new panics. Investigate (do not revert):
   - `AddDefinition: ldf.Formula is not a Definition: ...` (Group D)
   - Any sig.copy() restoration failure in compileUpdatePattern (Group C)

## Risks / considerations

- **Group A (subst_both_clauses API change)**: Changes the `SubstBothClauses` signature from `map[string]lg.Expr` to the codebase's standard `map[lg.NodeKey]lg.Expr` (matching `SubstituteClauses` and `SubstituteConstantsClauses` already in module/ops.go). The key construction `lg.Key(placeholderConst)` invokes `node.Sexp()` which produces a structural key carrying both name AND sort, matching Python's recstruct tuple-based equality. There are only 2 callsites (both in `actions/action.go`'s `UpdatePattern.Match`). Tests `TestSubstBothClauses` (if any) will need to construct keys via `lg.Key(...)` instead of using bare strings.

- **Group B (RewriteAtom interface change)**: Changes the AstRewriter interface signature. Ripple effect through every rewriter implementation (5 types) plus `AstRewrite` and `RewriteSort` themselves. This is the most invasive single change in the plan. Risk: missing a callsite and breaking the build. Mitigation: build compilation will catch all callsites the interface change affects.

- **Group C (sig.copy() in compileUpdatePattern)**: This change isolates placeholder symbols from the global signature. If any test was previously relying on placeholders leaking into `c.Sig` after UpdatePattern compile (extremely unlikely but possible), it would break. The fix is the right behavior — leaking placeholders is itself a bug. Investigate any new test failures.

- **Bug 4-i deferral**: The parser-side SymbolList element type mismatch is **not addressed** in this plan per explicit user direction. A naive fix risks losing type information from the App nodes that may be needed by downstream consumers. This is added to follow-ups as a deeper investigation.

- **Group D (NativeExpr routing)**: This change moves NativeExpr-RHS definitions from LabeledProps to NativeDefinitions. Any test that asserts on `len(LabeledProps)` after a `definition foo = <native ...>` declaration will see a different count. Same for `len(NativeDefinitions)`. Investigate; the new behavior matches Python.

- **No parser changes / no goyacc regen**: Bug 4-i is deferred, so neither `parser/grammar_v17.y` nor the generated `parser/grammar_v17.go` is touched in this plan.

- **Order of execution**: Groups can be implemented in any order, but Group B's interface change is the riskiest, so doing it first lets failures surface early.

## Critical files to be modified

- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/module/ops.go` (Group A1 — SubstBothClauses signature: `map[string]lg.Expr` → `map[lg.NodeKey]lg.Expr`)
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/action.go` (Group A2 — UpdatePattern.Match, actionMatch, nodeMatch use `map[lg.NodeKey]lg.Expr` and `lg.Key(placeholder)`)
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/ast/rewrite.go` (Group B — AstRewriter interface + 5 implementations + AstRewrite + RewriteSort)
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go` (Group C — compileUpdatePattern sig.copy())
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/decl.go` (Group D — AddDefinition method, Derived, DefinitionDecl)

**NOT modified**: `parser/grammar_v17.y`, `parser/grammar_v17.go` (Bug 4-i deferred per user direction).

## Existing helpers to reuse

- `addDefinitionChecks` (compiler/decl.go:324-342) — reused by the new `AddDefinition` method
- `nodeRepStr` (compiler/compiler.go:518-530) — already handles Symbol nodes for Group C
- `Compiler.Sig.Copy` / restore pattern (compiler/compiler.go:1313-1315) — reused for Group C2's compileUpdatePattern
- `lookupOrCreateConst` (compiler/compiler.go:509-514) — reused unchanged in Group C2
- `lg.Key`, `lg.NewConst`, `lg.NewVariable` — used by Group A's reworked SubstBothClauses

## Out-of-scope / follow-ups (NOT in this plan)

- **Bug 4-i (deferred): SymbolList stores Apps instead of strings/Symbols at parse time.** This affects the `apps` rule used by `update X from Y { ... }` declarations. The Go parser currently passes `[]ast.Node` (App nodes) where Python passes string `.rep` extracts. The naive fix (parser-side extraction to Symbols) would lose type information from the App nodes — sort annotations, lineno, attribute references — that may be needed by downstream consumers not yet audited. A proper fix requires:
  1. Audit of every `SymbolList.Elems` consumer in the codebase to confirm what data they actually need.
  2. Investigation of what type information lives on the App nodes that would be lost by the naive fix.
  3. Decision: either (a) introduce a parallel `Names []string` field on SymbolList and populate both, (b) have the parser extract a richer Symbol-with-attributes node, or (c) leave the indirect runtime workaround in place and document it.
  This audit must be performed deeply before any parser change.
- Audit of other rewriters used outside `instantiateMacro` (e.g., `subst_prefix_atoms_ast` callers in compile chain) — beyond Area 3's interface fix.
- Investigation of whether `ast_match` in Go (`nodeMatch`/`actionMatch`) handles Variable placeholders. Python's `ast_match` accepts both `is_variable(y) or is_constant(y)` placeholders. Go's nodeMatch only handles Const. May need a follow-up audit.
- Audit of the `compile_native_def` chain in Python (ivy_compiler.py:1445-1446) versus Go ARGSetup native handling — adjacent to Area 5 but covers a different pathway (top-level native declarations, not native-RHS definitions).
- Verification of `subst_subscripts` in Go vs Python — used by `RewriteName` in `AstRewriteSubstConstantsParams`.
- Audit of `BindOldsClauses(node lg.Expr) lg.Expr` (transrel.go:1586) which was noted in the previous plan as dead code parallel to `BindOldsClausesClauses` — separate cleanup.
