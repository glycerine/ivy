# Compiler Pass-Correctness & TODO Completion Plan

## Context

The previous implementation pass (Fixes 1–16) fixed many Python divergences in the compiler but left behind TODOs in the code and identified additional double-storage bugs that were deferred. This plan completes all remaining work: every TODO becomes an implementation, every double-storage bug becomes a no-op, and the pass 1/3 boundary matches Python exactly.

---

## Fix A: DomainSetup double-storage — `Assert`, `Private`, `Isolate`, `Attribute`, `Init`

**Problem:** Python's `IvyDomainSetup` (pass 1) does NOT have `_assert`, `private`, `isolate`, `attribute`, or `init`. These only exist in `IvyARGSetup` (pass 3). The Go `DomainSetup` currently has working implementations that store data, causing double-storage since `ARGSetup.ProcessDecls` also handles them.

**Evidence:**
- Python `IvyDomainSetup` class (ivy_compiler.py:1024–1343) — none of these methods exist
- Python `IvyARGSetup` class (ivy_compiler.py:1399–1530) — all five are defined there

**Files:** `compiler/decl.go`

**Fix:** Make each a no-op in DomainSetup, matching the pattern already applied for Export/Import/Delegate:

1. `DomainSetup.Assert()` (line ~1388) → `return nil` (ARGSetup handles at ivy_compile.go:503)
2. `DomainSetup.Private()` (line ~1170) → `return nil` (ARGSetup handles at ivy_compile.go:559)
3. `DomainSetup.Isolate()` (line ~871) → `return nil` (ARGSetup handles at ivy_compile.go:522)
4. `DomainSetup.Attribute()` (line ~1091) → `return nil` (ARGSetup handles at ivy_compile.go:584)
5. `DomainSetup.Init()` (line ~768) → already skipped via `case *ast.InitDecl: // skip` in ProcessDecl, but the Init method body still does work; confirm it's truly unreachable, or make it `return nil` for safety

Note: `Init` is already guarded by the `case *ast.InitDecl: // skip` comment in `ProcessDecl` (decl.go:99–103), so `DomainSetup.Init()` is never called in pass 1. The body can stay or be made a no-op — no functional difference. To be clean, make it a no-op with a comment.

---

## Fix B: TypeDecl EnumeratedSort — missing `Functions` and `Constructors` registration

**Problem:** Python's `IvyDomainSetup.type()` (ivy_compiler.py:1243–1247) registers each enum constructor in both `self.domain.functions[sym] = 0` and `self.domain.sig.constructors.add(sym)`. Go's `TypeDecl()` calls `AddSymbol` but does NOT register in `mod.Functions` or `sig.Constructors`.

**File:** `compiler/decl.go` — the `case *ast.EnumeratedSort:` block (line ~425)

**Fix:** After `AddSymbol`, add:
```go
mod.Functions[elemName] = sort
sig.Constructors[elemName] = true
```

---

## Fix C: Interpret() range bounds — proper compilation (Fix 5 TODO)

**Problem:** Both `TypeDecl()` (line ~442) and `Interpret()` (line ~962) use `fmt.Sprint(rng.Lo)` / `fmt.Sprint(rng.Hi)` instead of compiling bounds. The `RangeSort.Lb`/`Ub` fields are `string` type.

**Python** (ivy_compiler.py:1295–1306):
```python
def compile_bound(b):
    if not ivy_logic.is_numeral_name(b.rep):
        b.sort = lhs
        self.parameter(b)
    with top_sort_as_default():
        res = b.compile()
    with ASTContext(thing):
        res = sort_infer(res, sort)
    return res
```

**Fix — two parts:**

### Part 1: Change `RangeSort.Lb`/`Ub` from `string` to typed interface
**File:** `logic/sort.go:111-116`

Add a new interface that constrains the bound to its two legal forms:
```go
// NumeralOrCompiledBound represents a range bound that is either a
// numeral string (e.g. "0", "255") or a compiled logic expression
// (e.g. a parameter symbol after sort inference).
type NumeralOrCompiledBound interface {
    BoundString() string // human-readable representation
    IsNumeral() bool     // true if this is a literal numeral string
}

// NumeralBound is a literal numeral string like "0" or "255".
type NumeralBound struct {
    Value string
}
func (b NumeralBound) BoundString() string { return b.Value }
func (b NumeralBound) IsNumeral() bool     { return true }

// CompiledBound is a compiled logic expression (e.g. a parameter symbol).
type CompiledBound struct {
    Expr Expr  // the compiled lg.Expr
}
func (b CompiledBound) BoundString() string { return fmt.Sprint(b.Expr) }
func (b CompiledBound) IsNumeral() bool     { return false }
```

Then change:
```go
type RangeSort struct {
    ast.Base
    Name string
    Lb   NumeralOrCompiledBound
    Ub   NumeralOrCompiledBound
}
```

Update `String()` to call `Lb.BoundString()` / `Ub.BoundString()`. Audit all readers of Lb/Ub (theory compilation, sort printing, etc.) to use the interface methods or type-switch when they need the underlying value.

### Part 2: Implement `compileBound` helper
**File:** `compiler/decl.go`

Add a `compileBound` method on `DomainSetup` that:
1. Extracts `b.Rep` from the bound node
2. If `!il.IsNumeralName(b.Rep)`: sets `b.Sort = lhs`, calls `d.Parameter(b)` to register it
3. Compiles with `TopSortAsDefault` context: `res = d.Compiler.CompileNode(b)` (or equivalent)
4. Runs `SortInfer(res, sort)`
5. Returns the compiled expression

Apply in both `TypeDecl` range case and `Interpret` range case.

### Audit points for Lb/Ub readers:
- `logic/sort.go` — `RangeSort.String()`
- `theory/` — `get_theory_schemata` equivalent
- `solver/` — any encoding that reads range bounds
- `cppgen/`, `gogen/` — code generators that emit range info

---

## Fix D: Interpret() solver validation (Fix 9 TODO)

**Problem:** Go has `// TODO: validate via slv.is_solver_sort(rhs)` and `// TODO: validate via slv.is_solver_op(rhs)`. Both `IsSolverSort` and `IsSolverOp` exist in `solver/encoding.go:322-349`.

**Import concern:** `solver` does NOT import `compiler`, so `compiler → solver` is safe (no cycle).

**File:** `compiler/decl.go` — the `Interpret()` function, final sort/symbol interpretation block (line ~1005)

**Fix:**
1. Add `slv "github.com/glycerine/goivy/solver"` to imports in `decl.go`
2. Replace TODO comments:
```go
if inSorts {
    if !slv.IsSolverSort(rhsName) {
        return lg.NewIvyError(node, fmt.Sprintf("%s not a native sort", rhsName))
    }
    ...
}
if inSymbols {
    if !slv.IsSolverOp(rhsName) {
        return lg.NewIvyError(node, fmt.Sprintf("%s not a native symbol", rhsName))
    }
    ...
}
```

---

## Fix E: CheckDefinitions — call AdmitDefinition (Fix 13 TODO)

**Problem:** `compiler/ivy_compile.go:1438` has `// TODO: call prover.AdmitDefinition(d, pmap[d.ID]) when ported`. The method `proof.ProofChecker.AdmitDefinition` is fully implemented at `proof/checker.go:331-382`.

**Import concern:** `proof/phase5_matching.go` imports `compiler` → `compiler → proof` creates a cycle.

**Fix — callback on Module:**

1. Add to `module/module.go`:
   ```go
   // AdmitDefinitionFn is injected by the driver to call proof.ProofChecker.AdmitDefinition
   // without creating a compiler→proof import cycle. Python: prover.admit_definition(d, pmap[d.id])
   AdmitDefinitionFn func(defn *ast.LabeledFormula, proof interface{}) error
   ```

2. In `CheckDefinitions` (compiler/ivy_compile.go), replace the TODO at line ~1438:
   ```go
   if mod.AdmitDefinitionFn != nil {
       if err := mod.AdmitDefinitionFn(d, pmap[d.ID]); err != nil {
           return err
       }
   }
   ```

3. Set the callback from the top-level driver code that CAN import both `compiler` and `proof`. Find where `IvyCompile` is called — likely in `end2end/`, `ivyshell/`, or `main.go` — and wire:
   ```go
   mod.AdmitDefinitionFn = func(defn *ast.LabeledFormula, pf interface{}) error {
       prover := proof.NewProofChecker(mod.LabeledAxioms, nil, mod.Schemata)
       _, err := prover.AdmitDefinition(defn, pf.(ast.Node))
       return err
   }
   ```

4. Audit callers: grep for `IvyCompile(` and `compiler.IvyCompile(` to find all call sites that need the callback wiring.

---

## Fix F: Parameter default — store AST node instead of string (Fix 14)

**Problem:** `Parameter()` stores `dflt = fmt.Sprint(def.Rhs)` but Python stores the raw AST node.

**Downstream usage:** `cppgen/repl.go:334-335` reads ParamDefaults as strings and injects into C++ code. The string approach works for C++ codegen but loses AST structure.

**File:** `module/module.go` (type change), `compiler/decl.go` (storage)

**Fix:**
1. Change `ParamDefaults []string` to `ParamDefaults []interface{}` in `module/module.go:95`
2. In `compiler/decl.go Parameter()`: store `def.Rhs` directly instead of `fmt.Sprint(def.Rhs)`
3. Store `nil` for no-default case (instead of `""`)
4. Update `cppgen/repl.go` to call `fmt.Sprint()` at consumption time
5. Update `isolate/strip.go:338` to append `nil` instead of `""`
6. Update `module/module.go` Copy/Reset methods for the new type
7. Update any test assertions about ParamDefaults

---

## Fix G: DomainSetup.Action — thunk scanning (Fix 10 TODO)

**Problem:** `DomainSetup.Action()` (decl.go:757-766) has a lazy `// TODO: scan for ThunkAction instances (rare feature, deferred)` with a wrong claim that this is "rare/deferred". It must match the Python faithfully.

**Python** (ivy_compiler.py:1345–1370): Walks `a.args[1].iter_subactions()` looking for `ThunkAction`. For each:
1. `subtype = action.args[0]` — the label atom
2. `suptype = action.args[2]` — the sort/type atom
3. Creates `TypeDef(subtype, UninterpretedSort())` → calls `self.type(tdef)`
4. Creates `VariantDef(subtype, suptype)` → calls `self.variant(vdef)`
5. Computes `actname = compose_names(subtype.relname, 'run')`
6. Creates a self-parameter `Atom('fml:$self')` with `.sort = subtype.relname`
7. `orig_args = action.args[0].args + action.args[1].args`
8. Registers `top_context.actions[actname] = (orig_args + [selfparam], [], len(orig_args))`

**All Go infrastructure exists:**
- `ast.ThunkAction` struct: `ast/ast.go:505` — fields: `Label`, `Action`, `Sort`, `Body`
- `actions.ThunkAction`: `actions/action.go:984` — the compiled form (not needed here)
- `ast.TypeDef`: `ast/decl.go:428` — `NewTypeDef(name, value)`
- `ast.ConstantSort`: `ast/sort.go:12` — Go name for Python's `UninterpretedSort()`
- `ast.VariantDef`: `ast/decl.go:489` — `NewVariantDef(name, sort)`
- `iu.ComposeNames`: `ivyutils`
- `TopContext.Actions`: `compiler/compiler.go:92-93`
- `ActionInfo`: `compiler/compiler.go:26-32`

**File:** `compiler/decl.go`

**Fix — replace the TODO body with:**
```go
func (d *DomainSetup) Action(node ast.Node) error {
    actDef, ok := node.(*ast.ActionDef)
    if !ok {
        return nil
    }
    // Python: for action in a.args[1].iter_subactions():
    //             if isinstance(action, ThunkAction): ...
    iterASTSubactions(actDef.Body, func(sub ast.Node) error {
        thunk, ok := sub.(*ast.ThunkAction)
        if !ok {
            return nil
        }
        subtype := thunk.Label  // Python: action.args[0]
        suptype := thunk.Sort   // Python: action.args[2]

        // Python: tdef = TypeDef(subtype, UninterpretedSort()); self.type(tdef)
        tdef := ast.NewTypeDef(subtype, ast.NewConstantSort())
        if err := d.TypeDecl(tdef); err != nil {
            return err
        }

        // Python: vdef = VariantDef(subtype, suptype); self.variant(vdef)
        vdef := ast.NewVariantDef(subtype, suptype)
        if err := d.Variant(vdef); err != nil {
            return err
        }

        // Python: actname = compose_names(subtype.relname, 'run')
        subtypeAtom, ok := subtype.(*ast.Atom)
        if !ok {
            return nil
        }
        actname := iu.ComposeNames(subtypeAtom.Relname(), "run")

        // Python: selfparam = Atom('fml:$self', []); selfparam.sort = subtype.relname
        selfparam := ast.NewAtom("fml:$self")
        selfparam.ASort = ast.NewAtom(subtypeAtom.Relname())

        // Python: orig_args = action.args[0].args + action.args[1].args
        var origArgs []ast.Node
        origArgs = append(origArgs, thunk.Label.Args()...)
        origArgs = append(origArgs, thunk.Action.Args()...)

        // Python: top_context.actions[actname] = (orig_args + [selfparam], [], len(orig_args))
        allFormals := append(origArgs, selfparam)
        if d.Compiler.TopCtx != nil {
            d.Compiler.TopCtx.Actions[actname] = &ActionInfo{
                FormalAST: allFormals,
                KeyPos:    len(origArgs),
            }
        }
        return nil
    })
    return nil
}
```

**Also add** `iterASTSubactions` helper (recursive AST walker):
```go
// iterASTSubactions walks an AST action tree, calling fn for each node.
// This is the AST-level equivalent of Python's Action.iter_subactions().
func iterASTSubactions(node ast.Node, fn func(ast.Node) error) error {
    if node == nil {
        return nil
    }
    if err := fn(node); err != nil {
        return err
    }
    for _, child := range node.Args() {
        if err := iterASTSubactions(child, fn); err != nil {
            return err
        }
    }
    return nil
}
```

---

## Files to Modify

| File | Fixes |
|------|-------|
| `compiler/decl.go` | A (5 no-ops), B (enum constructors), C (compileBound), D (solver validation), G (thunk scan) |
| `compiler/ivy_compile.go` | E (AdmitDefinition callback) |
| `logic/sort.go` | C (RangeSort field types) |
| `module/module.go` | E (AdmitDefinitionFn field), F (ParamDefaults type) |
| `cppgen/repl.go` | F (ParamDefaults consumption) |
| `isolate/strip.go` | F (ParamDefaults append) |

---

## Verification

1. `make test` — no regressions after each fix
2. Verify DomainSetup no-ops: existing end2end tests (TestParseCompile_*) exercise the full 3-pass pipeline — if double-storage was hiding, removing it may surface issues
3. For Fix C (range bounds): write a test with a non-numeral range bound (e.g., `interpret t -> {n..m}` where n,m are parameters) to verify bounds are compiled, not stringified
4. For Fix D (solver validation): write a test that interprets a sort as a non-solver name and expects an error
5. For Fix F (ParamDefaults): verify cppgen tests still pass after type change
