# Option B Expr Removal Implementation Plan

Created: 2026-05-28

This plan expands `~/ivy/expr_removal_options.md` Option B into an
implementation recipe. It assumes the repository is the current flat
`goivy` package, where "ast.Node" in the design means the existing
`goivy.Node` interface.

The goal is to remove the `Expr` interface and stop maintaining a second
logic-expression node hierarchy. Parsed and compiled logical expressions
should be represented by ordinary AST nodes with a resolved sort stored on
the node, matching Python's "node plus `.sort` attribute" model.

This is a large migration. The tree is allowed to be broken while the
change is in flight. The important invariant is the final state: no
exported `Expr` interface, no `[]Expr`/`map[NodeKey]Expr` data model, no
logic mirror types for formulas/terms, and a rewritten test suite that
asserts the new AST-shaped output.

Do not use git while doing the implementation. The user owns git.

## 1. Target State

### 1.1 One tree interface

`goivy.Node` is the only tree interface for formulas, terms, actions, sort
nodes, and declaration wrappers.

Keep `Node` focused on tree behavior:

```go
type Node interface {
    Args() []Node
    Clone(args []Node) Node
    GetLineno() Location
    SetLineno(Location)
    String() string
    Canon() Canonical
    GetAstConfig() *AstConfig
}
```

Do not add `Sexp`, `Equal`, or `Children` to `Node`. Those are logical
helpers, not AST interface requirements.

Add a resolved-sort slot to `Base`:

```go
type Base struct {
    Loc    Location
    HasLoc bool
    Cfg    *AstConfig `json:"-"`

    // Resolved compile-time sort. Nil means "not compiled/sort not known".
    // This must not be emitted by Canon().
    Sort Sort `json:"-"`
}
```

Even though several AST nodes already have direct fields named `Sort`,
for parse annotations, keep the base field named `Sort` and always access
it through helpers. Never use promoted `n.Sort` in new code unless the
node's concrete type makes that unambiguous.

Add these helpers, probably in a new `goivy/node_sort.go`:

```go
func NodeSort(n Node) Sort
func SetNodeSort(n Node, s Sort) Node
func MustNodeSort(n Node) Sort
func NodeSortOrTop(n Node) Sort
func CloneNodeWithSort(n Node, args []Node, s Sort) Node
```

Rules:

- `NodeSort(nil) == nil`.
- For a semantic `Sort` node, `NodeSort(sortNode) == sortNode`.
- For normal AST nodes embedding `Base`, `NodeSort` returns `Base.Sort`.
- For action nodes, `NodeSort` may return `ActionS` while action code still
  needs that concept. Actions should not be treated as logical formulas.
- `NodeSortOrTop` is the explicit compatibility helper for places that
  intentionally use Python's unknown-sort behavior.
- `MustNodeSort` panics or returns a clear error if the sort is nil. Use it
  at solver/codegen boundaries where an unsorted node is a bug.

`Base.Canon()` and `Base.canonFields()` must continue to omit resolved
sort. Python's canonical AST output does not include the compile-time
`.sort` attribute; it only includes explicit parse annotations.

### 1.2 Structural identity is separate from Canon

Current logic code uses `Expr.Sexp()` for structural identity and map keys.
`Canon()` is not a substitute because `Canon()` must skip `Base.Sort`.

Replace method-based sexp with package-level functions:

```go
type NodeKey string

func Key(n Node) NodeKey
func Sexp(n Node) NodeKey
func SortKey(s Sort) NodeKey
func NodeEqual(a, b Node) bool
```

During the migration, make `Key(Node)` first call any existing
`interface{ Sexp() NodeKey }` implementation so old logic nodes still work
until they are deleted. The final implementation should not depend on
`Expr.Sexp()` methods.

Final `Sexp(Node)` rules:

- `Variable` emits the same logical identity old `LogicVariable.Sexp`
  emitted, using `Variable.Rep` and `NodeSort(variable)`.
- A zero-argument `Atom` used as a constant emits the same logical identity
  old `Const.Sexp` emitted: `(Symbol name:<rep> sort:<sort>)`.
- `App` emits old `Apply` identity, using `App.Rep` as the function node
  and `App.Terms` as arguments.
- Equality is represented as `Atom{Rep:"=", Terms:[lhs,rhs]}` and emits the
  old `Eq` identity.
- `And`, `Or`, `Not`, `Implies`, `Iff`, `Ite`, `Forall`, `Exists`,
  `Lambda`, `NamedBinder`, `Globally`, `Eventually`, `WhenOperator`, `Cond`,
  `Definition`, `Let`, `SomeExpr`, `Literal`, and native expression nodes
  each get explicit `Sexp` cases.
- Sort sexps include sort structure exactly as today, so `NodeKey`-based
  maps keep their semantics.

### 1.3 Tree traversal uses Args, with helper exceptions

Delete the `Children() []Expr` contract. Use:

```go
func NodeChildren(n Node) []Node
```

For most nodes this returns `n.Args()`.

Keep the existing Python-shaped behavior for binders: `Forall.Args()` and
`Exists.Args()` should return the body only; bound variables are accessed
through binder helpers, not ordinary children.

For `App`, `Args()` returns application terms only, not the function, just
like Python `Apply.args`. Code that must walk the function must explicitly
visit `App.Rep`.

Add binder helpers:

```go
func BinderVars(n Node) []*Variable
func BinderBody(n Node) Node
func CloneBinder(n Node, vars []Node, body Node) Node
func IsBinder(n Node) bool
func IsQuantifier(n Node) bool
```

### 1.4 Final concrete type mapping

Use this mapping everywhere:

| Old type/API | Final type/API |
| --- | --- |
| `Expr` | `Node` |
| `[]Expr` | `[]Node` |
| `map[NodeKey]Expr` | `map[NodeKey]Node` |
| `NodeMap`, `NodeSet` values | `Node` |
| `n.NodeSort()` | `NodeSort(n)` |
| `n.Children()` | `NodeChildren(n)` or `n.Args()` |
| `n.Equal(m)` | `NodeEqual(n, m)` |
| `n.Sexp()` | `Sexp(n)` |
| `ExprName(n)` | `NodeName(n)` |
| `LogicVariable` | `Variable` |
| `LogicVariable.Name` | `Variable.Rep` |
| `LogicVariable.VSort` | `NodeSort(variable)` |
| `Const` | zero-argument `Atom` |
| `Const.Name` | `Atom.Rep` |
| `Const.CSort` | `NodeSort(atom)` |
| `Apply` | `App` |
| `Apply.Func` | `App.Rep` |
| `Apply.Terms` | `App.Terms` as `[]Node` |
| `Apply.aSort` | `NodeSort(app)` |
| `Eq` | `Atom{Rep:"=", Terms:[]Node{lhs,rhs}}` |
| `LogicAnd` | `And` |
| `LogicOr` | `Or` |
| `LogicNot` | `Not` |
| `LogicImplies` | `Implies` |
| `LogicIff` | `Iff` |
| `LogicIte` | `Ite` |
| `ForAll` | `Forall` |
| `ForAll.Variables` | `Forall.Bounds` |
| `LogicExists` | `Exists` |
| `LogicExists.Variables` | `Exists.Bounds` |
| `Lambda` | new AST `Lambda` node with `Bounds` and `Body` |
| `LogicNamedBinder` | `NamedBinder` |
| `LogicNamedBinder.Variables` | `NamedBinder.Bounds` |
| `LogicWhenOperator` | `WhenOperator` |
| `LogicGlobally` | `Globally` |
| `LogicEventually` | `Eventually` |
| `LogicDefinition` | `Definition` |
| `LogicDefinitionSchema` | `DefinitionSchema` |
| `LogicLet` | `Let` |
| `LogicSome` | `SomeExpr` when value-producing, `Some` when formula-only |
| `LogicLiteral` | `Literal` |
| `LogicNativeExpr` | `NativeExpr` with `Base.Sort = TopS` |

Do not rename compiled action types such as `LogicSequence` and
`LogicAssumeAction` in this migration. They are not the formula/term mirror,
and renaming them would mix a separate Python-faithfulness project into the
`Expr` removal.

### 1.5 Application representation

Use `Atom` for symbols and equality, and `App` for applications.

Final compiled forms:

- Constant or nullary symbol: `*Atom{Rep:name, Terms:nil}`, with
  `Base.Sort` set to the symbol sort.
- Function or relation application: `*App{Rep:symbolAtom, Terms:args}`,
  where `NodeSort(App.Rep)` is the full function/relation sort and
  `NodeSort(App)` is the range sort.
- Equality: `*Atom{Rep:"=", Terms:[]Node{lhs,rhs}}`, with `Base.Sort`
  set to `Boolean`.
- Boolean literals: `True` is empty `*And`, `False` is empty `*Or`.

This preserves the information old `Apply.Func.CSort` carried. Do not
represent relation applications as `Atom{Rep:"p", Terms:args}` after
compilation, because that loses the full relation sort needed by sort
inference, z3 translation, and code generation.

### 1.6 Constructors to preserve

Keep familiar package-level constructor names, but change their return
types to AST nodes:

```go
func NewConst(name string, sort Sort) *Atom
func NewVariable(name string, sort Sort) (*Variable, error)
func NewApply(fn Node, terms ...Node) (*App, error)
func MustApply(fn Node, terms ...Node) *App
func NewEq(t1, t2 Node) (*Atom, error)
func NewAnd(terms ...Node) (*And, error)
func NewOr(terms ...Node) (*Or, error)
func NewNot(body Node) (*Not, error)
func NewImplies(t1, t2 Node) (*Implies, error)
func NewIff(t1, t2 Node) (*Iff, error)
func NewIte(cond, thenNode, elseNode Node) (*Ite, error)
func NewForAll(vars []*Variable, body Node) (*Forall, error)
func NewExists(vars []*Variable, body Node) (*Exists, error)
func NewLambda(vars []*Variable, body Node) (*Lambda, error)
func NewNamedBinder(name string, vars []*Variable, environ *string, body Node) (*NamedBinder, error)
```

These constructors should enforce the same sort checks the old logic
constructors enforced. They should set `Base.Sort` on the returned node.

For AST parser constructors on `*AstConfig`, keep the current pre-compile
API (`cfg.NewVariable(rep string, sort string)`, `cfg.NewAtom`, etc.).
Those constructors leave `Base.Sort` nil unless they are deliberately used
to build already-compiled test nodes.

### 1.7 Sort hierarchy target

The immediate `Expr` removal must also remove `Sort`'s dependency on
`Expr`:

```go
type Sort interface {
    Node
    sortSeal()
}
```

Sort nodes are still ordinary `Node`s. `NodeSort(sortNode)` returns the
sort node itself.

The strict Option B end state has no `LogicFunctionSort` or
`LogicEnumeratedSort` names. Because `ast_sort.go` already has parse-time
`FunctionSort` and `EnumeratedSort`, do the rename after the `Expr` removal
build is compiling:

1. Make `FunctionSort` able to represent both parse syntax and semantic
   sorts:
   - parse fields already present: `Dom []Node`, `Rng Node`;
   - semantic field to add: `Sorts []Sort`;
   - `cfg.NewFunctionSort(dom, rng)` fills parse fields;
   - package `NewFunctionSort(sorts ...Sort)` fills semantic `Sorts`;
   - `Domain()`, `Range()`, `Arity()`, `IsFinite()`, `sortSeal()` use
     `Sorts`.
2. Merge `LogicEnumeratedSort` into `EnumeratedSort` by adding semantic
   fields such as `Name string` and `Ext []string`. Rename the existing
   method `Extension()` to `ExtensionNames()` first to avoid a field/method
   collision.
3. Keep `UninterpretedSort` as the semantic sort. Leave
   `ConstantSort`/`UninterpretedSortAST` as parse syntax until the parser
   annotation cleanup pass.
4. Keep `RangeSort` as the semantic range sort and `Range` as parse syntax
   until the same cleanup pass.

If time runs short, it is acceptable to finish the first no-`Expr` pass
with `LogicFunctionSort` and `LogicEnumeratedSort` still named that way,
provided they no longer implement or depend on `Expr`. The final grep must
make that status obvious and the remaining rename should be purely
mechanical.

## 2. Implementation Sequence

The steps below are ordered to minimize conceptual work during the actual
edit. The build may be broken for multiple phases. Use compiler errors as
the work queue, but do not change the target model.

### Phase 0: Inventory and baseline

1. Read this plan and `expr_removal_options.md`.
2. Do not run any git command.
3. Capture search inventories with `rg`:
   - `rg -n "\btype Expr\b|\bExpr\b" goivy -g '*.go'`
   - `rg -n "LogicVariable|LogicAnd|LogicOr|LogicNot|LogicImplies|LogicIff|LogicIte|LogicExists|LogicNamedBinder|LogicWhenOperator|LogicDefinition|LogicNativeExpr" goivy -g '*.go'`
   - `rg -n "\*Const|\*Apply|\*ForAll|\*Eq" goivy -g '*.go'`
   - `rg -n "NodeSort\(\)|Children\(\)|\.Equal\(|\.Sexp\(" goivy -g '*.go'`
4. Optional baseline command before editing:
   - `go test ./goivy -run 'TestLogic|TestCompiler|TestTypeInfer|TestZ3' -count=1`
   - Do not spend time fixing unrelated baseline failures.

### Phase 1: Add Node-level sort and identity helpers

Files: `goivy/ast.go`, new `goivy/node_sort.go`, new
`goivy/node_sexp.go`, possibly `goivy/ivyutils_typename.go`.

1. Add `Sort Sort` to `Base`.
2. Add helper methods on `Base`:
   - `func (b *Base) GetResolvedSort() Sort`
   - `func (b *Base) SetResolvedSort(s Sort)`
   These are optional convenience methods; keep package-level `NodeSort`
   as the normal public API.
3. Keep all `Canon()` output unchanged. Do not include `Base.Sort` in
   `Base.Canon`, `canonFields`, or any AST node `Canon`.
4. Change `Key` in `logic_sexp.go` or a replacement file to accept `Node`
   instead of `Expr`. Because `Expr` currently embeds `Node`, this widening
   should not break old callers.
5. Add transitional `Sexp(Node)`:
   - if `n` has old `Sexp() NodeKey`, call it;
   - otherwise dispatch on AST nodes;
   - fallback to `NodeKey(n.Canon())` only for non-logical wrappers where
     sort is irrelevant.
6. Add `NodeEqual(a,b Node) bool { return Key(a) == Key(b) }`.
7. Add `NodeChildren(n Node) []Node`:
   - if old `Children() []Expr` exists, convert to `[]Node`;
   - otherwise return `n.Args()`.
8. Add `NodeName(n Node) string` with cases for `Atom`, `Variable`,
   `App`, sort types, `WhenOperator`, and `NamedBinder`.
9. Add test helpers in a new `_test.go` file or an existing common test
   helper:
   - `mustNodeSort(t, n Node) Sort`
   - `assertNodeSort(t, n Node, want Sort)`
   - `asAtom`, `asApp`, `asAnd`, `asOr`, `asNot`, `asForall`,
     `asExists`, `asVariable`
   - `assertPretty(t, n Node, want string)`
   These helpers allow tests to be rewritten before every production call
   site is final.

### Phase 2: Break Sort free from Expr

Files: `goivy/logic_sort.go`, `goivy/logic_node.go`,
`goivy/logic_action_sort.go`, `goivy/typeinfer_sortvar.go`,
`goivy/ivylogic.go`, `goivy/logic_sexp.go`.

1. Change `type Sort interface` so it no longer embeds `Expr`.
2. Make `Sort` embed `Node` directly, or require `Node`-equivalent methods
   explicitly if an action-sort singleton needs a custom implementation.
3. Replace `SortEqual`'s use of `Equal(Expr)` with `SortKey` or a new
   `sortEqualConcrete(a,b Sort) bool` switch.
4. Update sort types so they satisfy `Node` without relying on
   `logic_ast_compat.go`.
5. Change `CompiledBound.Expr Expr` to `CompiledBound.Node Node`.
6. Change `ContainsTopSort` and `IsPolymorphic` to accept either `Node` or
   `Sort` through helpers:
   - `ContainsTopSortNode(n Node) bool`
   - `ContainsTopSortSort(s Sort) bool`
   - keep old names if practical, but signatures should not mention `Expr`.
7. Update `IvySortName`, `SortDomain`, `SortRange`, `IsRelationalSort`,
   `LogicRelationSort`, `FuncConstSort`, and `TopFunctionSort` to use the
   new sort interface.
8. Keep `LogicFunctionSort`/`LogicEnumeratedSort` names temporarily if that
   reduces breakage. The key requirement in this phase is "Sort no longer
   depends on Expr".

### Phase 3: Add missing AST node shapes

Files: `goivy/ast_formula.go`, `goivy/ast.go`,
`goivy/ast_decl_ast.go` if `NativeExpr` lives there.

1. Add `Lambda`:
   - fields: `Base`, `Bounds []Node`, `Body Node`;
   - `Args()` returns `[]Node{Body}`;
   - `Clone(args)` preserves `Bounds` and replaces body;
   - `String`, `Canon`, and `Sexp` match current logic behavior.
2. Add `Cond`:
   - fields: `Base`, `T1 Node`, `T2 Node`;
   - sort is `NodeSort(T2)`;
   - `Args()` returns `{T1,T2}`.
3. Add `Environ *string` to `Globally` and `Eventually`.
   - Parser-created nodes leave it nil.
   - Compiler-created nodes set it the same way current logic compile does.
   - Pretty printing currently ignores environ, but `Sexp` should include it.
4. Add `Environ *string` to `NamedBinder`.
   - Parser-created nodes leave it nil.
   - `NewNamedBinder` accepts an environ argument.
5. Make `NativeExpr` usable as a compiled node:
   - compiled children stay in the existing `Args` representation;
   - set `Base.Sort = TopS` after compilation;
   - delete `LogicNativeExpr` later.
6. Ensure all new/changed nodes' `Clone` methods preserve `Base`, including
   resolved sort.

### Phase 4: Introduce final constructors while old logic nodes still exist

Files: `goivy/logic_term.go`, `goivy/logic_formula.go`,
`goivy/ivylogic_constructors.go`, or a new `goivy/node_constructors.go`.

This is the first intentionally breaking phase if constructors are changed
in place.

1. Replace package-level `NewConst` so it returns `*Atom`.
   - `Atom.Rep = name`
   - `Atom.Terms = nil`
   - `Base.Sort = sort`
   - precompute no cached sexp; use `Sexp(atom)`.
2. Replace package-level `NewVariable` so it returns `*Variable`.
   - reject bad lowercase names as today;
   - set `Rep = name`;
   - set `VSort = IvySortName(sort)` only when preserving an explicit
     annotation is useful for printing; `Base.Sort` is the source of truth;
   - set `Base.Sort = sort`.
3. Replace `NewApply`, `MustApply`, `TryApply`, and `NewApplyUnchecked` so
   they return `*App`.
   - Validate `NodeSort(fn)` exactly as old `NewApply` validated
     `fn.NodeSort()`.
   - Store `fn` in `App.Rep`.
   - Store copied `terms` in `App.Terms`.
   - Set `Base.Sort` to function range or `TopS`.
4. Replace `NewEq` and `NewEqualsNode` so equality is an `*Atom`:
   - `Rep = "="`;
   - `Terms = []Node{lhs,rhs}`;
   - `Base.Sort = Boolean`;
   - preserve old sort validation.
5. Replace formula constructors to return AST nodes with sort set:
   - `NewAnd` -> `*And`
   - `NewOr` -> `*Or`
   - `NewNot` -> `*Not`
   - `NewImplies` -> `*Implies`
   - `NewIff` -> `*Iff`
   - `NewIte` -> `*Ite`
   - `NewForAll` -> `*Forall`
   - `NewExists` -> `*Exists`
   - `NewLambda` -> `*Lambda`
   - `NewNamedBinder` -> `*NamedBinder`
6. Update `True` and `False`:
   - `var True Node = &And{Base: Base{Sort: Boolean}}`
   - `var False Node = &Or{Base: Base{Sort: Boolean}}`
   - `IsTrue(Node)` checks empty `*And`.
   - `IsFalse(Node)` checks empty `*Or`.
7. Keep the old logic type definitions physically present until all
   references are gone. Their constructors will no longer produce them.

### Phase 5: Change core APIs from Expr to Node

This is the broadest mechanical phase.

1. In public data structures, change fields:
   - `Clauses.Fmlas []Expr` -> `[]Node`
   - `IvyDefinition.Lhs/Rhs Expr` -> `Node`
   - `LogicSome` users -> `SomeExpr`/`Some` as Node
   - `Module.AllRelations []Expr` -> `[]Node`
   - `Module.Rely []Expr` -> `[]Node`
   - `Module.ExtPreconds map[string]Expr` -> `map[string]Node`
   - `Module.Instantiator func([]Expr) *Clauses` -> `func([]Node) *Clauses`
   - `Vocab.Relations`, `Vocab.Functions`, `Vocab.Variables`,
     `Vocab.Sorts` should use `Node` for symbols/variables and `Sort` for
     sorts.
2. In action APIs:
   - `type Action interface { Node; ... }`
   - `ActionClone(args []Expr)` -> `ActionClone(args []Node)`
   - `ActionArgs() []Expr` -> `ActionArgs() []Node`
   - action formula/proof/local/actual fields become `Node`;
   - formal parameter/return slices become `[]*Atom` if they are always
     symbols, otherwise `[]Node`.
3. Delete `actionArgsToNodes` and `actionsNodesToExprs`. They should become
   no-ops or disappear.
4. In `compiler.go`:
   - `Cmpl(Node) (Node,error)`
   - `CompileNode(Node) (Node,error)`
   - `compileNodeCore(Node,bool) (Node,error)`
   - `SortifyWithInference(Node) (Node,error)`
   - `SortInfer(Node) (Node,error)`
   - `compileArgs(Node) ([]Node,error)`
   - `CompileWithSortInference(Node) (Node,error)` remains.
5. In `compiler_action.go` and `compiler_decl.go`, apply the same
   signature replacement. Do not keep local `Expr` aliases.
6. In proof/matching:
   - `map[NodeKey]Expr` -> `map[NodeKey]Node`
   - `InsMap[NodeKey, Expr]` -> `InsMap[NodeKey, Node]`
   - matching functions accept and return `Node`.
7. In logic utilities:
   - `UsedVariables(Node) map[NodeKey]Node`
   - `FreeVariables(Node) *Omap[NodeKey, Node]`
   - `Substitute(Node, map[NodeKey]Node) (Node,error)`
   - `RenameVarsNoClash([]Node, []Node) []Node`
   - binder helpers return `[]*Variable`.
8. In z3, codegen, model checking, webui, and concept code, change only the
   signatures and type switches needed for compile. Keep behavioral cleanup
   for later phases.

Mechanical replacements to apply carefully:

| Replace | With |
| --- | --- |
| `Expr` | `Node` |
| `[]Expr` | `[]Node` |
| `map[NodeKey]Expr` | `map[NodeKey]Node` |
| `InsMap[NodeKey, Expr]` | `InsMap[NodeKey, Node]` |
| `NodeMap` value type `Expr` | `Node` |
| `NodeSet` value type `Expr` | `Node` |
| `.NodeSort()` | `NodeSort(...)` |
| `.Children()` | `NodeChildren(...)` or `.Args()` |
| `.Equal(x)` | `NodeEqual(..., x)` |
| `.Sexp()` | `Sexp(...)` |

Do not blindly replace `Expr` inside names such as `ExprContext`,
`NativeExpr`, `SomeExpr`, `compileExprVocab`, or external solver types.

### Phase 6: Rewrite compiler construction to clone AST nodes

Files: `goivy/compiler.go`, `goivy/compiler_action.go`,
`goivy/compiler_decl.go`, `goivy/proof_goal.go`,
`goivy/proof_phase5_matching.go`.

1. `compileAnd`:
   - compile child args;
   - clone to `*And`;
   - set sort `Boolean`;
   - return `Node`.
2. `compileOr`: same, sort `Boolean`.
3. `compileNot`:
   - compile body;
   - return `*Not` with sort `Boolean`.
4. `compileImplies`, `compileIff`: clone with compiled children, sort
   `Boolean`.
5. `compileIte`:
   - compile condition, then, else;
   - set sort to `NodeSort(then)`.
6. `compileGlobally` and `compileEventually`:
   - compile body;
   - clone AST node;
   - set `Environ` as current logic code does;
   - set sort `Boolean`.
7. `compileWhenOperator`:
   - compile both args;
   - set sort to first argument's sort.
8. `CompileQuantifier`:
   - compile bounds into `*Variable` nodes with `Base.Sort` set;
   - extend `VarCtx` with `bound.Rep -> NodeSort(bound)`;
   - compile body;
   - return `*Forall` or `*Exists` with `Bounds` and `Body`, sorted/deduped
     by `Key` where the old constructor did that.
9. `compileNamedBinder`:
   - compile bounds into `*Variable`;
   - compile body;
   - return `*NamedBinder` with `Bounds`, `Body`, `Environ`;
   - set sort to function sort over bound variable sorts plus body sort.
10. `compileSymbol`:
    - if bound variable, return `NewVariable(name, sort)`;
    - if signature symbol, return `NewConst(name, entry.Sort)`;
    - if uppercase unresolved, return `NewVariable(name, TopS or annotated sort)`;
    - otherwise return `NewConst(name, TopS)`.
11. `CompileVariable`:
    - resolve declared or contextual sort;
    - return `*Variable` with `Base.Sort` set.
12. `CompileApp(*Atom)`:
    - resolve alias;
    - `"true"` and `"false"` return `True`/`False`;
    - compile args with `ReturnCtx` nil;
    - `"="` returns equality `*Atom`;
    - find polymorphic or signature symbol as a zero-term `*Atom` with the
      full symbol sort;
    - if no args, return the symbol atom;
    - if args, return `NewApply(symbolAtom, args...)`;
    - field-reference fallback returns `Node`.
13. `compileAppNode(*App)`:
    - if `Rep` is a parse `*Symbol`, convert to `CompileApp` as today;
    - if `Rep` is `*NamedBinder` or another node, compile terms first, then
      compile `Rep` with `Cmpl`, then `NewApply(repNode,args...)`.
14. `CompileLF`:
    - stop asserting `compiled.Formula.(Expr)`;
    - `Formula` remains a `Node`;
    - clone labels and formula exactly as Python does.
15. `compileNativeExpr`:
    - clone native expression with compiled children;
    - set `Base.Sort = TopS`;
    - return `*NativeExpr`.
16. `compileTrigger`:
    - compile pattern and terms as `Node`;
    - return `*And` or the chosen trigger representation with sort `Boolean`.
17. `compileDefnImpl`:
    - equality is `Atom("=")`;
    - extracting lhs/rhs from compiled equality means reading
      `eq.Terms[0]` and `eq.Terms[1]`;
    - definitions are AST `Definition`/`DefinitionSchema`.
18. Every former `return &LogicAnd{...}` fallback becomes `NewAnd(...)` or a
    direct `*And` with `Base.Sort = Boolean`.

### Phase 7: Rewrite type inference around Node

Files: `goivy/typeinfer_infer.go`, `goivy/typeinfer_unify.go`,
`goivy/typeinfer_sortvar.go`, `goivy/ivylogic_sortinfer.go`.

1. Change:
   - `InferResult.Concretize func() (Node,error)`
   - `InferSorts(t Node, env map[string]SortOrVar) (*InferResult,error)`
   - `ConcretizeSorts(t Node, s Sort) (Node,error)`
   - `ConcretizeTerms(terms []Node, sorts []Sort) ([]Node,error)`
   - `SortInfer(term Node, sort Sort) (Node,error)`
   - `SortInferList(terms []Node, sorts []Sort, unsorted map[string]bool) ([]Node,error)`
   - `CheckConcretelySorted(term Node, unsorted map[string]bool) error`
2. Specialized cases:
   - `*Variable`: key by `Rep`; declared sort is `NodeSort(n)` unless nil,
     then parse `VSort`, then `TopS`; concretize by cloning/creating a
     `Variable` and setting `Base.Sort`.
   - zero-term `*Atom`: constant symbol; key by `Rep`; declared sort is
     `NodeSort(n)` or `TopS`; concretize to `NewConst(rep, concreteSort)`.
   - equality `*Atom` with `Rep == "="`: infer both terms, unify their
     sorts, concretize to equality atom with `Boolean` sort.
   - `*App`: infer `Rep` explicitly, infer terms, unify function sort with
     `FunctionSort(termSorts..., resultSort)`, concretize to `NewApply`.
   - `*Ite`, `*Cond`, `*Not`, `*And`, `*Or`, `*Implies`, `*Iff`,
     `*Globally`, `*Eventually`, `*WhenOperator`: same sort checks as old
     logic cases, but return AST constructors.
   - `*Forall`, `*Exists`, `*Lambda`, `*NamedBinder`: use `Bounds`; create
     fresh sort variables for bound variables; concretize bounds and body;
     set binder sort.
   - `*Definition`, `*Let`, `*Some`, `*SomeExpr`, `*Literal`, `*NativeExpr`:
     use generic fallback unless a special sort rule is required.
3. `collectNames` must visit:
   - variables;
   - zero-term atom symbols;
   - `App.Rep`;
   - `App.Terms`;
   - binder bounds and bodies.
4. `CheckConcretelySorted` must use `UsedVariables` and `UsedSymbolsAst`
   rewritten for Node.
5. Keep Python's permissive TopSort behavior exactly where old code did.
   Use `NodeSortOrTop` only in those places; otherwise nil sorts should
   expose incomplete compile paths.

### Phase 8: Rewrite pretty printing

Files: `goivy/logic_pretty.go`, any `String()` methods that call it.

1. Change `PrettyFmla(n Expr)` to `PrettyFmla(n Node)`.
2. Change `PrettyFmlaAmbiguous(n Node)`.
3. Change `ugly(n Node, prec int)`.
4. Update cases:
   - `*Variable` uses `Rep` and `NodeSort`.
   - zero-term `*Atom` uses `Rep` and `NodeSort` for numeral annotations.
   - equality `*Atom` uses `naryUgly("="...)`.
   - `*App` uses `App.Rep` for function name and `App.Terms` for args.
   - formula nodes use AST names.
5. `dropAnnotations` now returns `Node`.
6. When dropping a variable annotation, construct a `*Variable` with
   `Base.Sort = TopS`; do not mutate the original.
7. When dropping a numeral constant annotation, construct a zero-term
   `*Atom` with `Base.Sort = TopS`.
8. Preserve all existing pretty output. The expected pretty strings should
   not change.

### Phase 9: Rewrite logic utilities, clauses, and matching

Files: `goivy/logicutil.go`, `goivy/logicutil_logic_utils.go`,
`goivy/module_clauses.go`, `goivy/module_litclause.go`,
`goivy/proof_goal.go`, `goivy/proof_matching.go`,
`goivy/proof_phase5_matching.go`, `goivy/proof_match.go`.

1. `UsedVariables`:
   - variables are `*Variable`;
   - constants are zero-term `*Atom`;
   - explicitly visit `App.Rep`;
   - use `Key(Node)`.
2. `FreeVariables`:
   - bound set keys use `Key(*Variable)`;
   - binder variables come from `BinderVars`.
3. `Substitute`:
   - map type is `map[NodeKey]Node`;
   - variable substitution keys are `Key(variable)`;
   - constant substitution keys are `Key(zeroTermAtom)`;
   - `App` substitutes both `Rep` and terms when appropriate;
   - constructors preserve resolved sort.
4. `EqualModAlpha`:
   - binders use `Bounds`;
   - variables use `Rep`;
   - structural fallback uses `Sexp`.
5. `Clauses`:
   - formulas are `[]Node`;
   - `ToOpenFormula` and `ToFormula` return `Node`;
   - `FormulaToClauses(Node, annot)`;
   - `dropUniversals` and `dropExistentials` use `*Forall` and `*Exists`.
6. `LogicLiteral` replacement:
   - use `Literal` with `Atom Node`;
   - if code needs a compiled literal, ensure `Atom` is a Node formula with
     `Boolean` sort.
7. Matching:
   - all match maps become `map[NodeKey]Node`;
   - `CompileOneMatch(lhs,rhs Node, ...)`;
   - `ApplyMatchAlt` returns `Node`;
   - `ApplyFun(fun Node,args []Node)`;
   - `ApplyMatchSort` still operates on `Sort`.
8. Labeled formula helpers:
   - remove `GoalConcExpr` or make it return `Node`;
   - no `Formula.(Expr)` assertions remain.

### Phase 10: Rewrite solver and code generators

Files: `goivy/z3bridge_translate.go`, `goivy/z3bridge_solver*.go`,
`goivy/mc_*.go`, `goivy/ivy2go/*.go`, `goivy/ivy2cpp/*.go`,
`goivy/gogen/*.go`, webui display paths.

1. Z3 translator:
   - `Translate(Node)`;
   - `translateCore(Node, caller string)`;
   - `TermToZ3(Node)`;
   - variable case is `*Variable`;
   - constant/symbol case is zero-term `*Atom`;
   - application case is `*App`;
   - equality case is `*Atom` with `Rep == "="`;
   - quantifier variables are `[]*Variable`;
   - enumerated sorts use the merged or temporary sort type.
2. Codegen:
   - `emitExpr(Node)`;
   - replace `*Const` with zero-term `*Atom`;
   - replace `*Apply` with `*App`;
   - replace `*ForAll`/`*LogicExists` with `*Forall`/`*Exists`;
   - generate constructor calls rather than struct literals wherever
     possible. This makes future field cleanup easier.
3. Web UI and ART display:
   - use `PrettyFmla(Node)`;
   - use `NodeName`, `NodeSort`, and `NodeChildren`.
4. Model checker and encoders:
   - state/input/output symbol slices become `[]*Atom` or `[]Node`;
   - if a slice is semantically "symbols only", prefer `[]*Atom`.

### Phase 11: Delete the old logic mirror

Do this only after `rg` shows no production references to the old concrete
types.

Delete or reduce these files:

- `goivy/logic_node.go`: delete `Expr`; move any surviving helpers to
  `node_sort.go` or `node_sexp.go`.
- `goivy/logic_term.go`: delete `LogicVariable`, `Const`, `Apply`; keep
  only generic helpers that now operate on Node.
- `goivy/logic_formula.go`: delete `Eq`, `LogicIte`, `LogicNot`,
  `LogicAnd`, `LogicOr`, `LogicImplies`, `LogicIff`, `ForAll`,
  `LogicExists`, `Lambda` if replaced, `LogicNamedBinder`,
  `LogicWhenOperator`, `Cond` if replaced.
- `goivy/logic_definition.go`: delete after `Definition`/`DefinitionSchema`
  are used everywhere.
- `goivy/logic_nativeexpr.go`: delete after `NativeExpr` carries compiled
  children and sort.
- `goivy/logic_ast_compat.go`: delete entirely.
- `goivy/logic_canon.go`: delete or replace with `Sexp(Node)` helpers.
- `goivy/logic_sexp.go`: keep `NodeKey`, `Key`, `Sexp`, `NodeMap`,
  `NodeSet`, but no Expr methods.
- `goivy/actions_action_expr.go`: rewrite into an action Node/canon file
  only. Remove `Children`, `NodeSort`, `Equal`, and `Sexp` methods unless
  actions still need explicit `Sexp` through package-level dispatch.

After deletion, run:

- `rg -n "\btype Expr\b|\bExpr interface\b|goivy\.Expr|\[\]Expr|map\[NodeKey\]Expr|InsMap\[NodeKey, Expr\]" goivy -g '*.go'`
- `rg -n "LogicVariable|LogicAnd|LogicOr|LogicNot|LogicImplies|LogicIff|LogicIte|LogicExists|LogicNamedBinder|LogicWhenOperator|LogicDefinition|LogicNativeExpr" goivy -g '*.go'`
- `rg -n "\*Const|\*Apply|\*ForAll|\*Eq" goivy -g '*.go'`

Allowed leftovers:

- `ExprContext`, `NativeExpr`, `SomeExpr`, and function names like
  `CompileExprVocab` may remain.
- Action types named `LogicSequence`, `LogicAssumeAction`, etc. may remain.
- External solver packages may have their own unrelated `Expr` names.

### Phase 12: Merge/rename sort types

Only start this after the no-`Expr` build compiles.

1. Change `Sort` helper code to prefer final names:
   - `LogicFunctionSort` -> `FunctionSort`;
   - `LogicEnumeratedSort` -> `EnumeratedSort`.
2. Merge `FunctionSort` as described in section 1.7.
3. Merge `EnumeratedSort` as described in section 1.7.
4. Update all type switches and tests.
5. Keep parse-only `ConstantSort`, `UninterpretedSortAST`, and `Range`
   temporarily if parser cleanup is not part of this change.
6. Final sort grep:
   - `rg -n "LogicFunctionSort|LogicEnumeratedSort" goivy -g '*.go'`

### Phase 13: Format and compile

1. Run `gofmt` on touched Go files.
2. Start with narrow compile checks:
   - `go test ./goivy -run '^$' -count=0`
   - `go test ./goivy/webui -run '^$' -count=0`
   - `go test ./goivy/ivy2go -run '^$' -count=0`
   - `go test ./goivy/ivy2cpp -run '^$' -count=0`
3. Then run focused tests by subsystem after the corresponding tests are
   rewritten.
4. Finish with broader runs:
   - `go test ./goivy -count=1`
   - `go test ./... -count=1` if practical.

## 3. Test Rewrite Plan

Rewrite tests alongside production code. Do not try to preserve old test
assertions that name `LogicX` types; those tests are supposed to change.

### 3.1 Add shared test helpers first

Create a helper file in the main package tests, for example
`goivy/expr_removal_helpers_test.go`.

Recommended helpers:

```go
func requireSort(t *testing.T, n Node, want Sort)
func requireSortName(t *testing.T, n Node, want string)
func requirePretty(t *testing.T, n Node, want string)
func requireKey(t *testing.T, n Node, want NodeKey)

func requireAtom(t *testing.T, n Node) *Atom
func requireConstAtom(t *testing.T, n Node, name string) *Atom
func requireEqAtom(t *testing.T, n Node) *Atom
func requireApp(t *testing.T, n Node) *App
func requireVariable(t *testing.T, n Node, name string) *Variable
func requireAnd(t *testing.T, n Node) *And
func requireOr(t *testing.T, n Node) *Or
func requireNot(t *testing.T, n Node) *Not
func requireForall(t *testing.T, n Node) *Forall
func requireExists(t *testing.T, n Node) *Exists
```

Use helpers instead of repeating type assertions. This keeps future cleanup
mechanical and makes failures read like the new model.

### 3.2 Mechanical assertion rewrites

Use this mapping in tests:

| Old assertion | New assertion |
| --- | --- |
| `result.(*LogicAnd)` | `requireAnd(t, result)` |
| `result.(*LogicOr)` | `requireOr(t, result)` |
| `result.(*LogicNot)` | `requireNot(t, result)` |
| `result.(*LogicImplies)` | `result.(*Implies)` or helper |
| `result.(*LogicIff)` | `result.(*Iff)` or helper |
| `result.(*LogicIte)` | `result.(*Ite)` or helper |
| `result.(*ForAll)` | `requireForall(t, result)` |
| `result.(*LogicExists)` | `requireExists(t, result)` |
| `result.(*LogicVariable)` | `requireVariable(t, result, "X")` |
| `result.(*Const)` | `requireConstAtom(t, result, "x")` |
| `result.(*Apply)` | `requireApp(t, result)` |
| `result.(*Eq)` | `requireEqAtom(t, result)` |
| `x.NodeSort()` | `NodeSort(x)` |
| `x.Name` on variable | `x.Rep` |
| `x.Name` on const | `x.Rep` |
| `x.VSort` on logic variable | `NodeSort(x)` |
| `x.CSort` | `NodeSort(x)` |
| `app.Func` | `app.Rep` |
| `fa.Variables` | `fa.Bounds` |
| `ex.Variables` | `ex.Bounds` |
| `expr.Children()` | `NodeChildren(expr)` or concrete fields |

For equality tests, assert:

```go
eq := requireEqAtom(t, result)
if got := len(eq.Terms); got != 2 { ... }
```

For relation/function application tests, assert:

```go
app := requireApp(t, result)
fn := requireConstAtom(t, app.Rep, "p")
requireSort(t, fn, expectedFunctionSort)
requireSort(t, app, Boolean)
```

### 3.3 Unit test files by category

#### `logic_formula_test.go`

Rename mentally to "compiled AST formula tests"; physical file rename is
optional.

Rewrite expectations:

- Constructors still named `NewAnd`, `NewOr`, etc., but return AST types.
- `Forall` and `Exists` use `Bounds`.
- `Lambda` uses `Bounds`.
- `NamedBinder` uses `Bounds` and optional `Environ`.
- `Children()` expectations become `Args()` or `NodeChildren()`:
  - Forall/Exists/Lambda/NamedBinder should still expose body only through
    `Args()`/`NodeChildren()`.
- Pretty output should stay byte-for-byte the same.
- Bad-sort tests should stay, but call `NodeSort` in diagnostics.

#### `logic_term_test.go`

Rewrite:

- `NewVariable` returns `*Variable`.
- `NewConst` returns zero-term `*Atom`.
- `NewApply` returns `*App`.
- `Const.Name` assertions become `Atom.Rep`.
- `Const.CSort` assertions become `NodeSort(atom)`.
- `Apply.Func` assertions become `App.Rep`.
- `Apply.aSort` checks become `NodeSort(app)`.

#### `logic_sort_test.go`

Rewrite after Phase 2 and again after Phase 12 if sort names are merged.

- `Sort` no longer satisfies `Expr`; tests should assert it satisfies `Node`
  and `Sort`.
- Use `SortKey` for structural identity.
- `FunctionSort`/`LogicFunctionSort` expectations depend on whether Phase
  12 has been completed. If Phase 12 is postponed, document the temporary
  name in the test.

#### `logic_sexp_test.go`, `logic_sexp_cross_test.go`, `python_logic_sexp_test.go`

These tests are important. Keep the logical sexp strings as stable as
possible even though the concrete Go types changed.

Rewrite all direct `.Sexp()` calls to `Sexp(node)` and all `Key(expr)` calls
to `Key(node)`.

Expected strings should generally remain:

- variables still look like `(Variable name:X sort:...)`;
- constants still look like `(Symbol name:c sort:...)`;
- applications still look like `(Apply func:... terms:[...])`;
- equality still looks like `(Eq t1:... t2:...)`.

Do not change expected strings merely because the concrete Go type is now
`Atom`; sexp is semantic identity, not concrete type reflection.

#### `compiler_test.go`, `compiler_*_test.go`

Rewrite result-shape assertions:

- AST `And` compiles to `*And`, not `*LogicAnd`.
- AST relation application compiles to `*App` whose `Rep` is a zero-term
  symbol atom with function/relation sort.
- Equality compiles to equality `*Atom`.
- Variables compile to `*Variable` with `Base.Sort` set.
- `LabeledFormula.Formula` remains `Node`; remove all `.(Expr)` assertions.

For tests checking Python parity, prefer pretty strings, `Sexp`, and
resolved sort names over concrete Go type names unless the type shape is
the point of the test.

#### `logicutil_test.go`, `logicutil_div_conformance_test.go`

Rewrite around Node maps:

- substitution maps are `map[NodeKey]Node`;
- bound/free variable helpers return `*Variable`;
- lambda/binder tests use `Bounds`;
- `SubstituteApply` works on `*App`.

Capture-avoidance tests should keep the same semantic assertions but use
`Variable.Rep` and `NodeSort(variable)`.

#### `module_*_test.go`, `actions_*_test.go`

Rewrite:

- `Clauses.Fmlas` is `[]Node`;
- `FormulaToClauses` accepts `Node`;
- transition relation pieces use `*And`, `*Or`, `*Not`, equality atoms, and
  apps;
- action fields `Formula`, `Proof`, `Lhs`, `Rhs`, `Args`, `Locals`,
  `Modified` use `Node` or `*Atom` as appropriate.

Where tests currently inspect `*Const` in formals/modified sets, inspect
zero-term `*Atom` and use `Rep` plus `NodeSort`.

#### `proof_*_test.go`

Rewrite:

- match maps use `map[NodeKey]Node`;
- `InsMap[NodeKey, Node]`;
- schemas and goals keep `LabeledFormula` wrappers with `Formula Node`;
- no `unwrapLogicNode` assertions;
- variable matching uses `*Variable`;
- sort matching still uses `Sort`.

Keep trace-order tests intact unless the trace text contains concrete Go
type names. If trace text says `LogicVariable` or `LogicAnd`, update the
trace formatter to emit Python names (`Var`, `And`) rather than updating
goldens to Go-only names.

#### `z3bridge_*_test.go`

Rewrite translator construction helpers:

- variables are `*Variable`;
- constants are zero-term `*Atom`;
- apps are `*App`;
- equality is equality `*Atom`;
- quantifier variable slices are `[]*Variable`.

Expected SMT/Z3 output should not change except for debug-only type names.

#### `ivy2go`, `ivy2cpp`, `gogen` tests

Prefer emitted constructor calls over struct literals.

Update expectations:

- `goivy.NewConst(...)` may remain in emitted code if the constructor name
  is preserved.
- Any emitted `goivy.ForAll{Variables: []*goivy.LogicVariable{...}}` must
  become constructor-based code, e.g. `mustForall(...)`, or a new AST literal
  with `Bounds`.
- Runtime helper signatures such as `mustApply(fn goivy.Expr, ...)` become
  `mustApply(fn goivy.Node, ...)`.
- Template/native helper args become `[]goivy.Node`.

### 3.4 Golden/canon expectations

Canon tests should mostly not change. If an AST canon golden changes because
resolved `Base.Sort` appeared, fix the canon implementation instead of
updating the golden.

Sexp tests may change only if the semantic identity truly changes. The plan
is to preserve old logic sexp strings wherever possible.

Pretty-printing tests should not change. If they do, compare against Python
before accepting the change.

### 3.5 Test execution order after rewrite

After production compiles:

1. Constructor and sort basics:
   - `go test ./goivy -run 'TestLogic|TestSort|TestSexp' -count=1`
2. Compiler basics:
   - `go test ./goivy -run 'TestCompiler|TestExpr6|TestCompile' -count=1`
3. Type inference and logic utilities:
   - `go test ./goivy -run 'TestTypeInfer|TestLogicUtil|TestDIV' -count=1`
4. Clauses/actions:
   - `go test ./goivy -run 'Test.*Clause|Test.*Action|Test.*Update' -count=1`
5. Proof/matching:
   - `go test ./goivy -run 'TestProof|TestApplyMatch|TestGoal|TestTactic' -count=1`
6. Solver/codegen:
   - `go test ./goivy -run 'TestZ3|TestTranslate|TestBMC|TestMC' -count=1`
   - `go test ./goivy/ivy2go ./goivy/ivy2cpp ./goivy/gogen -count=1`
7. Broader:
   - `go test ./goivy -count=1`
   - `go test ./... -count=1`

## 4. File-by-file Touch List

Heavy-touch files:

- `goivy/ast.go`
- `goivy/ast_formula.go`
- `goivy/ast_sort.go`
- `goivy/logic_node.go`
- `goivy/logic_term.go`
- `goivy/logic_formula.go`
- `goivy/logic_sort.go`
- `goivy/logic_definition.go`
- `goivy/logic_nativeexpr.go`
- `goivy/logic_ast_compat.go`
- `goivy/logic_sexp.go`
- `goivy/logic_canon.go`
- `goivy/logic_pretty.go`
- `goivy/ivylogic.go`
- `goivy/ivylogic_constructors.go`
- `goivy/ivylogic_formula.go`
- `goivy/ivylogic_sortinfer.go`
- `goivy/ivylogic_util.go`
- `goivy/ivylogic_symbols.go`
- `goivy/ivylogic_sig.go`
- `goivy/typeinfer_infer.go`
- `goivy/typeinfer_unify.go`
- `goivy/typeinfer_sortvar.go`
- `goivy/compiler.go`
- `goivy/compiler_action.go`
- `goivy/compiler_decl.go`
- `goivy/module.go`
- `goivy/module_action.go`
- `goivy/actions_action.go`
- `goivy/actions_action_expr.go`
- `goivy/actions_update.go`
- `goivy/actions_transforms.go`
- `goivy/module_clauses.go`
- `goivy/module_litclause.go`
- `goivy/logicutil.go`
- `goivy/logicutil_logic_utils.go`
- `goivy/proof_goal.go`
- `goivy/proof_matching.go`
- `goivy/proof_phase5_matching.go`
- `goivy/proof_match.go`
- `goivy/z3bridge_translate.go`
- `goivy/z3bridge_solver*.go`
- `goivy/mc_*.go`
- `goivy/ivy2go/*.go`
- `goivy/ivy2cpp/*.go`
- `goivy/gogen/*.go`
- webui display/session files that call `PrettyFmla` or inspect symbols.

Likely delete or mostly empty:

- `goivy/logic_node.go`
- `goivy/logic_ast_compat.go`
- `goivy/logic_canon.go`
- `goivy/logic_definition.go`
- `goivy/logic_nativeexpr.go`

Keep but rewrite:

- `goivy/logic_sexp.go` as node sexp/key support.
- `goivy/logic_pretty.go` as node pretty support, or rename later.
- `goivy/logic_sort.go` as unified sort support, later renamed if desired.

## 5. Completion Checklist

The migration is complete when all of these are true:

1. `goivy` has no `type Expr interface`.
2. There is no exported `goivy.Expr` type or alias.
3. Production code has no `[]Expr`, `map[NodeKey]Expr`, or
   `InsMap[NodeKey, Expr]`.
4. `Sort` does not embed `Expr`.
5. `PrettyFmla` accepts `Node`.
6. `Compiler.CompileNode`, `Cmpl`, and `SortifyWithInference` return `Node`.
7. `LabeledFormula.Formula` is never asserted to `Expr`.
8. `logic_ast_compat.go` is gone.
9. Old mirror concrete types are gone or reduced to temporary sort-name
   leftovers explicitly documented in Phase 12:
   - no `LogicVariable`;
   - no `Const` type;
   - no `Apply` type;
   - no `Eq` type;
   - no `LogicAnd`, `LogicOr`, `LogicNot`, `LogicImplies`, `LogicIff`,
     `LogicIte`, `LogicExists`, `LogicNamedBinder`, `LogicWhenOperator`,
     `LogicDefinition`, `LogicNativeExpr`.
10. `Canon()` output does not include resolved sort.
11. `Sexp(Node)` includes resolved sort where old logic sexp did.
12. Pretty output for existing tests is unchanged unless Python says
    otherwise.
13. The test suite has been rewritten to assert AST nodes and `NodeSort`.

## 6. Common Pitfalls

1. Do not replace `Canon()` with `Sexp()`. Canon is for parser/module
   parity; sexp is for logical identity.
2. Do not represent compiled relation applications as `Atom` with terms.
   Use `App` with a zero-term symbol atom in `Rep`, or the function sort is
   lost.
3. Do not put resolved sort into parse annotation fields (`VSort`, `ASort`,
   `Symbol.Sort`) during the first migration. Those fields are parse syntax;
   `Base.Sort` is compile state.
4. Do not keep `Formula.(Expr)` assertions "just for now". They hide exactly
   the polymorphism this migration removes.
5. Do not update canon goldens to include `Base.Sort`.
6. Do not let tests assert Go-only `LogicX` type names after the migration.
7. Do not use a permanent `type Expr = Node` alias. A temporary alias can
   help during a local compile checkpoint, but the final tree must not have
   it.
8. Do not forget `App.Rep` traversal. Old `Apply.Children()` intentionally
   skipped the function; many algorithms compensated manually.
9. Do not rename action types in this migration. Keep the blast radius on
   formula/term/sort `Expr` removal.
10. Do not use git commands.

## 7. Suggested Final Greps

Run these before declaring the migration done:

```sh
rg -n "\btype Expr\b|\bExpr interface\b|goivy\.Expr|\[\]Expr|map\[NodeKey\]Expr|InsMap\[NodeKey, Expr\]" goivy -g '*.go'
rg -n "LogicVariable|LogicAnd|LogicOr|LogicNot|LogicImplies|LogicIff|LogicIte|LogicExists|LogicNamedBinder|LogicWhenOperator|LogicDefinition|LogicNativeExpr" goivy -g '*.go'
rg -n "\*Const|\*Apply|\*ForAll|\*Eq" goivy -g '*.go'
rg -n "\.NodeSort\(\)|\.Children\(\)|\.Equal\(|\.Sexp\(" goivy -g '*.go'
rg -n "Formula\.\(Expr\)|\.\(Expr\)" goivy -g '*.go'
```

Expected allowed hits:

- `ExprContext`, `NativeExpr`, `SomeExpr`, and function names containing
  `Expr` as historical names.
- Action type names beginning with `Logic`.
- External packages or solver-local expression names unrelated to
  `goivy.Expr`.

Everything else should be either removed or justified in a short comment.
