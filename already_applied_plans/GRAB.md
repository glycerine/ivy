# Fix All Remaining LALR_AUDIT.md Items: C7, C8, M1-M10, L1-L3

**Created:** 2026-03-25T02:00

## Context

This plan covers fixing all remaining discrepancies between the Go LALR grammar (`grammar_v17.y`) and the Python source of truth (`ivy_parser.py`). Items C1-C6, C9, and C2 were already fixed in prior sessions. The remaining 15 items are: C7, C8, M1-M10, L1-L3.

---

## C7. `IF *` creates wrong type (Ite instead of ChoiceAction)

**File:** `lalr_full/grammar_v17.y:4069-4075`

**Current Go:**
```go
choice := ast.NewIte(ast.NewSymbol("*", nil), $3, $5)
```

**Python:** `ChoiceAction(p[3], p[5])` — creates a ChoiceAction with unique_id counter.

**Problem:** Go creates `Ite` (wrong AST type). Python creates `ChoiceAction` with a `unique_id` auto-increment counter.

**Fix — two parts:**

### Part A: Add ChoiceAction to `ast/ast.go`
`ChoiceAction` exists in `actions/action.go` (semantic layer) but NOT in `ast/` (parse layer). Need to add it to `ast/ast.go` matching the Python `ChoiceAction(AST)` class. It takes variadic args (2 branches: then, else) and has a `unique_id` counter.

```go
var choiceActionCounter int64

type ChoiceAction struct {
    Base
    Branches []Node
    UniqueID int64
}

func NewChoiceAction(branches ...Node) *ChoiceAction {
    choiceActionCounter++
    return &ChoiceAction{Branches: branches, UniqueID: choiceActionCounter}
}
func (c *ChoiceAction) Args() []Node { return c.Branches }
func (c *ChoiceAction) Clone(args []Node) Node {
    choiceActionCounter++
    return &ChoiceAction{Base: c.Base, Branches: args, UniqueID: choiceActionCounter}
}
func (c *ChoiceAction) Canon() iu.Canonical { ... }
```

### Part B: Fix grammar rule
```go
| TOK_IF TOK_TIMES sequence TOK_ELSE action
{
    choice := ast.NewChoiceAction($3, $5)
    choice.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    $$ = choice
}
```

### Tests: `lalr_full/choice_action_test.go`
- `TestIfStarCreatesChoiceAction` — parse `if * { skip } else { skip }`, verify result is `*ast.ChoiceAction`
- `TestChoiceActionHasUniqueID` — verify UniqueID is set and auto-increments
- `TestChoiceActionCanon` — verify canon output contains "choiceAction"
- `TestChoiceActionBranches` — verify 2 branches are stored

---

## C8. DEFINITION missing DefinitionSchema for explicit

**File:** `lalr_full/grammar_v17.y:940-952`

**Current Go:**
```go
lf := ast.NewLabeledFormula($4, $5)
// ... never checks $2 (optexplicit)
```

**Python:**
```python
foo = p[5]
if p[2]:  # optexplicit is True
    foo = DefinitionSchema(*foo.args)
    foo.lineno = p[5].lineno
lf = LabeledFormula(p[4], foo)
```

**Fix:** `DefinitionSchema` already exists at `ast/formula.go:316`. Add the check:

```go
| top optexplicit TOK_DEFINITION optlabel gdefn optproof
{
    $$ = $1
    gdefn := $5
    if $2 != nil {  // optexplicit is True
        if def, ok := gdefn.(*ast.Definition); ok {
            ds := &ast.DefinitionSchema{Definition: *def}
            ds.SetLineno(def.GetLineno())
            gdefn = ds
        }
    }
    lf := ast.NewLabeledFormula($4, gdefn)
    lf.Lineno = tokLineno(v17lex.(*v17LexAdapter), $3).Line
    lf = addLabel(lf, "def")
    dd := ast.NewDefinitionDecl(lf)
    $$.declare(dd)
    if $6 != nil {
        $$.declare(ast.NewProofDecl($6))
    }
}
```

**Note:** Need to check what type `optexplicit` returns — it's `%type <node>` in the grammar. Python returns `True`/`False` (bool). Go likely returns a node or nil. Check actual grammar definition for `optexplicit`.

### Tests: `lalr_full/definition_schema_test.go`
- `TestExplicitDefinitionCreatesSchema` — parse `explicit definition ...`, verify DefinitionSchema wraps the definition
- `TestNonExplicitDefinitionNoSchema` — parse `definition ...`, verify raw Definition (not wrapped)

---

## M1. THUNK action is a stub

**File:** `lalr_full/grammar_v17.y:4192-4196`

**Current Go:**
```go
$$ = ast.NewAtom("thunk", ast.NewAtom($2.Val), ast.NewAtom($3.Val), $8)
```

**Python:**
```python
p[0] = ThunkAction(Atom(p[2][1:-1],[]), action, Atom(p[6]), p[8])
```

**Fix:** `ThunkAction` already exists at `ast/ast.go:693` with constructor `NewThunkAction(label, action, sort, body)`.

```go
| TOK_THUNK labelname SYMBOLx optargs TOK_COLON atype TOK_ASSIGN sequence
{
    // Python: ThunkAction(Atom(p[2][1:-1],[]), action, Atom(p[6]), p[8])
    // p[2] is label (with brackets), p[3] is SYMBOL, p[4] is optargs, p[6] is atype, p[8] is body
    labelStr := strings.Trim($2.Val, "[]")
    label := ast.NewAtom(labelStr)
    label.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
    // Python: action = ActionDef(Atom(p[3],[]), p[8], formals=p[4])
    // Actually Python line 3232: action = ActionDef(Atom(p[3]),p[8],formals=p[4])
    actionAtom := ast.NewAtom($3.Val)
    actionAtom.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))
    sortAtom := ast.NewAtom($6.(*ast.Symbol).Rep)
    ta := ast.NewThunkAction(label, actionAtom, sortAtom, $8)
    ta.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    $$ = ta
}
```

**Caveat:** Need to re-read Python lines 3229-3235 carefully to get the exact parameter mapping. The `action` variable in Python is an `ActionDef`, not just the atom. May need to construct ActionDef here.

### Tests: `lalr_full/thunk_let_test.go`
- `TestThunkActionCreated` — parse `thunk [lab] foo(x:t) : t := { skip }`, verify `*ast.ThunkAction`
- `TestThunkActionFields` — verify label, action, sort, body fields set correctly

---

## M2. LET action is a stub

**File:** `lalr_full/grammar_v17.y:4186-4191`

**Current Go:**
```go
$$ = ast.NewAtom("let", args...)
```

**Python:**
```python
p[0] = LetAction(*(p[2]+[p[3]]))
```

**Fix:** `LetAction` does NOT exist in `ast/` package. Only in `actions/action.go:776` (semantic layer). Need to add `LetAction` to `ast/ast.go`.

### Part A: Add LetAction to `ast/ast.go`
Python `LetAction(Action)` takes variadic args where all-but-last are equation bindings and last is body.

```go
type LetAction struct {
    Base
    Bindings []Node // equation bindings (Equals/Definition nodes)
    Body     Node   // the body action
}

func NewLetAction(args ...Node) *LetAction {
    if len(args) == 0 {
        return &LetAction{}
    }
    return &LetAction{Bindings: args[:len(args)-1], Body: args[len(args)-1]}
}
func (l *LetAction) Args() []Node { return append(l.Bindings, l.Body) }
func (l *LetAction) Clone(args []Node) Node { ... }
func (l *LetAction) Canon() iu.Canonical { ... }
```

### Part B: Fix grammar rule
```go
| TOK_LET eqns sequence
{
    args := append($2, $3)
    la := ast.NewLetAction(args...)
    la.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    $$ = la
}
```

### Tests:
- `TestLetActionCreated` — parse `let x = y { skip }`, verify `*ast.LetAction`
- `TestLetActionBindingsAndBody` — verify bindings/body split

---

## M3. DECREASES missing Ranking wrapper and checkNonTemporal

**File:** `lalr_full/grammar_v17.y:4325-4329`

**Current Go:**
```go
$$ = []ast.Node{$2}
```

**Python:**
```python
rank = Ranking(check_non_temporal(p[2]))
rank.lineno = get_lineno(p,1)
p[0] = [rank]
```

**Fix:** `Ranking` does NOT exist in `ast/` package. Only in `actions/extra_actions.go:169`. Need to add to `ast/ast.go`.

### Part A: Add Ranking to `ast/ast.go`
Python `Ranking(Action)` just wraps a single formula. `name()` returns `'decreases'`.

```go
type Ranking struct {
    Base
    Fmla Node
}

func NewRanking(fmla Node) *Ranking {
    return &Ranking{Fmla: fmla}
}
func (r *Ranking) Args() []Node { return []Node{r.Fmla} }
func (r *Ranking) Clone(args []Node) Node { return &Ranking{Base: r.Base, Fmla: args[0]} }
func (r *Ranking) Canon() iu.Canonical { ... }
```

### Part B: Fix grammar rule
```go
| TOK_DECREASES fmla
{
    fmla := checkNonTemporal($2)
    rank := ast.NewRanking(fmla)
    rank.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
    $$ = []ast.Node{rank}
}
```

### Tests:
- `TestDecreasesCreatesRanking` — verify Ranking wrapper and checkNonTemporal applied

---

## M4. TYPE with GHOST not using GhostTypeDef

**File:** `lalr_full/grammar_v17.y:1053-1067` (uninterpreted type rule)

**Current Go:** Always uses `TypeDef`. Never checks `$3` (optghost).

**Python:** `tdfn = (GhostTypeDef if p[3] else TypeDef)(scnst, defsort)`

**Fix:** `GhostTypeDef` exists at `ast/decl.go:597`. Check both TYPE rules (uninterpreted at ~line 1053 and with-sort at ~line 1067):

```go
// For uninterpreted type:
var tdfn ast.Node
if $3 {  // optghost
    tdfn = &ast.GhostTypeDef{TypeDef: ast.TypeDef{Name: scnst, Value: ast.NewUninterpretedSortAST()}}
} else {
    tdfn = &ast.TypeDef{Name: scnst, Value: ast.NewUninterpretedSortAST()}
}
```

Same pattern for the TYPE = sort rule. (C9 already fixed the Range/InterpretDecl part; M4 adds the ghost check.)

### Tests:
- `TestGhostTypeUsesGhostTypeDef` — parse `ghost type t`, verify `*ast.GhostTypeDef`
- `TestNonGhostTypeUsesTypeDef` — parse `type t`, verify `*ast.TypeDef` (not ghost)

---

## M5. INVARIANT missing check_unprovable gate

**File:** `lalr_full/grammar_v17.y:~800`

**Current Go:** Always declares invariant regardless of unprovable flag.

**Python:**
```python
if not lf.unprovable or check_unprovable.get():
    p[0].declare(d)
```

**Fix:** `check_unprovable` is a thread-local flag from `ivy_actions.py`. In Go, add a global `checkUnprovable` bool (or use a function). When `lf.Unprovable` is true and `checkUnprovable` is false, skip the declaration.

For now, since `check_unprovable` defaults to `False` in Python (only set to `True` by `ivy_compiler.py` during certain verification modes), the behavior difference is: Python silently drops unprovable invariants by default; Go always declares them. This may not affect golden test output if no unprovable invariants appear.

**Approach:** Add `checkUnprovable` global var to grammar_v17.go. Default `false`. Add the gate:
```go
if !lf.Unprovable || checkUnprovable {
    $$.declare(d)
}
```

### Tests:
- `TestUnprovableInvariantSkippedByDefault` — parse invariant with `[unprovable]`, verify NOT declared when checkUnprovable=false
- `TestUnprovableInvariantDeclaredWhenEnabled` — verify declared when checkUnprovable=true

---

## M6. ASSERT/REQUIRE/ENSURE actions missing unprovable→Sequence conversion

**File:** `lalr_full/grammar_v17.y:3744-3808` (multiple action rules)

**Current Go:** Never replaces with empty Sequence.

**Python:**
```python
if p[1] and not check_unprovable.get():
    p[0] = Sequence()  # replace action with empty sequence
```

**Fix:** Same dependency on `checkUnprovable` as M5. When `optunprovable` is true and `checkUnprovable` is false, replace the action with `ast.NewSequence()`.

Apply to all 6 affected rules: ASSERT, ASSUME, REQUIRE, ENSURE, ASSERT...PROOF, ENSURE...PROOF.

### Tests:
- `TestUnprovableAssertBecomesSequence` — verify assertion replaced with empty Sequence
- `TestProvableAssertKept` — verify normal assertion preserved

---

## M7. makeMixinName format difference (dot→underscore)

**File:** `lalr_full/grammar_v17.y:175-179` (grammar_v17.go in generated code)

**Current Go:**
```go
name := fmt.Sprintf("%s[%s%d]", atom.Rep, suffix, lalrLabelCounter)
```

**Python:**
```python
name = atom.rep.replace(iu.ivy_compose_character, '_') + '[' + suffix + str(label_counter) + ']'
```

**Problem:** Python replaces `.` (ivy_compose_character) with `_` before composing. Go does not.

**Fix:**
```go
rep := strings.ReplaceAll(atom.Rep, ".", "_")
name := fmt.Sprintf("%s[%s%d]", rep, suffix, lalrLabelCounter)
```

### Tests:
- `TestMakeMixinNameReplacesDots` — call `makeMixinName` with atom rep `"foo.bar"`, verify output contains `"foo_bar["`

---

## M8. scenariomixin missing infer_action_params

**File:** `lalr_full/grammar_v17.y:4432-4458`

**Current Go:** Directly uses `$3` (optargs) and `$4` (optreturns) without inference.

**Python:**
```python
optargs, optreturns = infer_action_params(atom.rep, p[3], p[4])
```

**Fix:** `inferActionParams` exists at `grammar_v17.go:217`. Call it before creating the ActionDef:

```go
// scenariomixin BEFORE:
formals, returns := inferActionParams(v17lex.(*v17LexAdapter).accum, atom.Rep, $3, $4)
adef := &ast.ActionDef{Name: atom, Body: $5, FormalParams: formals, FormalReturns: returns}
```

Same for AFTER scenariomixin.

### Tests:
- `TestScenarioMixinInfersParams` — parse scenario with before/after mixin, verify params inferred from action definition

---

## M9. specimpl: GLOBAL and COMMON share single slot

**File:** `lalr_full/lalr_parser.go:96`, `lalr_full/grammar_v17.y:3475-3511`

**Current Go:** Single `specialAttribute string` field holds "spec", "impl", "private", "global", or "common".

**Python:** THREE separate globals: `special_attribute`, `global_attribute`, `common_attribute`. All three are combined into `self.attributes` tuple in `p_top_empty`.

**Fix:** Add two new fields to `v17LexAdapter`:
```go
type v17LexAdapter struct {
    specialAttribute string  // for "spec", "impl", "private"
    globalAttribute  string  // for "global"
    commonAttribute  string  // for "common"
}
```

Update specimpl rules:
- SPECIFICATION/IMPLEMENTATION/PRIVATE → set `specialAttribute`
- GLOBAL → set `globalAttribute`
- COMMON → set `commonAttribute`

Update the `top: /* empty */` rule (where attributes are consumed) to combine all three:
```go
attrs := []string{}
if lex.specialAttribute != "" { attrs = append(attrs, lex.specialAttribute) }
if lex.globalAttribute != "" { attrs = append(attrs, lex.globalAttribute) }
if lex.commonAttribute != "" { attrs = append(attrs, lex.commonAttribute) }
ia.attributes = attrs
lex.specialAttribute = ""
lex.globalAttribute = ""
lex.commonAttribute = ""
```

### Tests:
- `TestGlobalAndCommonCombined` — verify `global common { ... }` produces attributes tuple with both

---

## M10. instMod missing lineno parameter

**File:** `lalr_full/inst_mod.go:135`

**Current Go signature:**
```go
func instMod(ivy *ivyAccum, bodyDecls []ast.Node, pref *ast.Atom, subst map[string]string, vsubst map[string]*ast.Variable, modname string)
```

**Python signature:**
```python
def inst_mod(ivy, module, pref, subst, vsubst, modname=None, lineno=None):
```

**Python uses lineno for:** `set_reference_lineno(lineno)` before substitution, `set_reference_lineno(None)` after.

**Fix:** Add `lineno ast.Location` parameter. Update all callers (doInsts, createObject).

```go
func instMod(ivy *ivyAccum, bodyDecls []ast.Node, pref *ast.Atom, subst map[string]string, vsubst map[string]*ast.Variable, modname string, lineno ast.Location) {
    // ... existing code ...
    // In the spaa inner function, before subst_prefix:
    if lineno != (ast.Location{}) {
        ast.SetReferenceLineno(lineno)
    }
    // ... do substitution ...
    if lineno != (ast.Location{}) {
        ast.SetReferenceLineno(ast.Location{})
    }
}
```

Need to check if `SetReferenceLineno` exists. If not, add it.

### Tests:
- `TestInstModLinenoParameter` — verify instMod accepts and uses lineno parameter

---

## L1. eqn rule uses Definition instead of Equals

**File:** `lalr_full/grammar_v17.y:4334-4339`

**Current Go:**
```go
$$ = ast.NewDefinition(ast.NewApp(...), ast.NewApp(...))
```

**Python:**
```python
p[0] = Equals(App(p[1]), App(p[3]))
```

`Equals(x,y)` in Python = `lg.Eq(x,y)` which is `logic.Eq` — a logic-layer type. But at parse time, the result is passed to `LetAction` as bindings. In the AST layer, the equivalent is `Atom("=", lhs, rhs)`.

**Fix:** Use `ast.NewAtom("=", lhs, rhs)` instead of `ast.NewDefinition`:
```go
eqn:
    SYMBOLx TOK_EQ SYMBOLx
    {
        lhs := ast.NewApp(ast.NewSymbol($1.Val, nil))
        rhs := ast.NewApp(ast.NewSymbol($3.Val, nil))
        $$ = ast.NewAtom("=", lhs, rhs)
    }
```

### Tests:
- `TestEqnCreatesEqualsAtom` — verify eqn produces `Atom("=", ...)` not `Definition`

---

## L2. debugarg rule uses Definition instead of DebugItem

**File:** `lalr_full/grammar_v17.y:4005-4011`

**Current Go:**
```go
$$ = ast.NewDefinition(ast.NewApp(...), $3)
```

**Python:**
```python
lhs = App(p[1])
p[0] = DebugItem(lhs, p[3])
```

**Fix:** `DebugItem` exists at `ast/ast.go:675`. Use it:
```go
debugarg:
    SYMBOLx TOK_EQ fmla
    {
        lhs := ast.NewApp(ast.NewSymbol($1.Val, nil))
        lhs.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $1))
        $$ = &ast.DebugItem{Name: lhs, Value: $3}
        $$.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $2))
    }
```

### Tests:
- `TestDebugArgCreatesDebugItem` — verify debugarg produces `*ast.DebugItem` not `Definition`

---

## L3. doInsts missing object tracking

**File:** `lalr_full/inst_mod.go:122-123`

**Current Go:** Comment says "object tracking — deferred for now".

**Python:**
```python
if pref is None:
    ivy.objects.update(module.objects)
```

**Fix:** The `objects` map already exists on `ivyAccum` (ivy_module.go:44). Need to:
1. Track objects during module parsing (when an ObjectDecl is declared, add to `objects` map)
2. In `doInsts`, when `pref == nil`, merge module's objects into ivy's objects

```go
// In doInsts, after instMod call:
if pref == nil {
    // merge module objects
    if modAccum, ok := lookupModuleAccum(modName); ok {
        for k, v := range modAccum.objects {
            ivy.objects[k] = v
        }
    }
}
```

The challenge is that the module body is stored as `[]ast.Node` (decls), not as an `ivyAccum`. Need to either:
- Track objects when ObjectDecl is declared (in `ivyAccum.declare()`)
- Or extract objects from the decl list

### Tests:
- `TestDoInstsObjectTracking` — verify objects map is populated after module instantiation

---

## New AST Types Summary

| Type | File | Status |
|---|---|---|
| `ChoiceAction` | `ast/ast.go` | **NEW** — add struct + constructor + Canon |
| `LetAction` | `ast/ast.go` | **NEW** — add struct + constructor + Canon |
| `Ranking` | `ast/ast.go` | **NEW** — add struct + constructor + Canon |
| `DebugItem` | `ast/ast.go:675` | EXISTS — already has Canon |
| `ThunkAction` | `ast/ast.go:693` | EXISTS — already has Canon |
| `DefinitionSchema` | `ast/formula.go:316` | EXISTS — already has Canon |
| `GhostTypeDef` | `ast/decl.go:597` | EXISTS — needs Canon method |

---

## Files to Modify

| File | Changes |
|---|---|
| `ast/ast.go` | Add ChoiceAction, LetAction, Ranking types (~90 lines) |
| `ast/decl.go` | Add GhostTypeDef.Canon() if missing |
| `lalr_full/grammar_v17.y` | Fix rules C7, C8, M1-M4, M7-M8, L1-L2, and M5-M6 gates (~15 rule changes) |
| `lalr_full/lalr_parser.go` | M9: Add globalAttribute/commonAttribute fields |
| `lalr_full/inst_mod.go` | M10: Add lineno param; L3: Add object tracking |
| `lalr_full/grammar_v17.go` | M7: Fix makeMixinName dot replacement |
| **Test files (NEW):** | |
| `lalr_full/choice_action_test.go` | C7 tests |
| `lalr_full/definition_schema_test.go` | C8 tests |
| `lalr_full/thunk_let_test.go` | M1, M2 tests |
| `lalr_full/decreases_ranking_test.go` | M3 tests |
| `lalr_full/ghost_type_test.go` | M4 tests |
| `lalr_full/unprovable_test.go` | M5, M6 tests |
| `lalr_full/mixin_name_test.go` | M7 tests |
| `lalr_full/scenario_infer_test.go` | M8 tests |
| `lalr_full/specimpl_test.go` | M9 tests |
| `lalr_full/instmod_lineno_test.go` | M10 tests |
| `lalr_full/eqn_debug_test.go` | L1, L2 tests |
| `lalr_full/object_tracking_test.go` | L3 tests |

---

## Recommended Implementation Order

1. **C7** — ChoiceAction (new AST type + grammar fix) — blocks correct `if *` parsing
2. **C8** — DefinitionSchema (type exists, wire it) — simple
3. **L1** — eqn Equals fix (simple type swap)
4. **L2** — debugarg DebugItem fix (simple type swap)
5. **M1** — ThunkAction (type exists, fix grammar)
6. **M2** — LetAction (new AST type + grammar fix)
7. **M3** — Ranking (new AST type + grammar fix)
8. **M4** — GhostTypeDef (type exists, add check)
9. **M7** — makeMixinName dot replacement (one-line fix)
10. **M8** — scenariomixin inferActionParams (function exists, call it)
11. **M5** — check_unprovable gate for invariant
12. **M6** — unprovable→Sequence for assert/require/ensure
13. **M9** — specimpl separate attribute slots
14. **M10** — instMod lineno parameter
15. **L3** — doInsts object tracking

---

## Verification

1. `cd lalr_full && go generate` — regenerate parser after grammar changes
2. `go build ./...` — full build
3. `go test ./lalr_full/ -v` — all tests pass (existing + new)
4. `go test ./ast/ -v` — ast package tests pass
5. `make golden` — verify no regression in golden test output
6. Each item gets its own test file with focused tests

---

## LALR_AUDIT.md Update Rules

When marking items as fixed, ONLY prepend "FIXED. DONE." status text. NEVER delete original Python/Go code snippets or implementation details.
