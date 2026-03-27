package compiler

// Batch E — Red (failing) tests for:
//   §6.3 #19: compile_if_action — Some/SomeMinMax existential-if handling
//   §6.2 #14: compile_thunk_action — subtype/destructor/substitution logic

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// ============================================================================
// §6.3 #19: compile_if_action — Some/SomeMinMax existential-if handling
// ============================================================================
//
// Python compile_if_action has two branches:
// 1. isinstance(self.args[0], ivy_ast.Some) — existential-if
//    - copies sig, compile_const on params, sortify fmla, clone Some with compiled args
//    - for SomeMinMax, also compiles the index expression
// 2. else — normal boolean condition (already implemented in Go)
//
// The Go CompileIf currently only handles branch 2. These tests verify branch 1.

// TestCompileIfAction_SomeCondition tests that when an if-action has a Some
// condition ("if some x:T. phi { ... }"), the compiler compiles the params
// and formula within a copied signature, and returns an IfAction whose
// condition is a compiled Some node.
//
// Python:
//   if isinstance(self.args[0], ivy_ast.Some):
//       sig = ivy_logic.sig.copy()
//       with sig:
//           ls = self.args[0].params()
//           fmla = self.args[0].fmla()
//           cls = [compile_const(v, sig) for v in ls]
//           sfmla = sortify_with_inference(fmla)
//           sargs = cls + [sfmla]
//           args = [self.args[0].clone(sargs), self.args[1].compile()]
//       args += [a.compile() for a in self.args[2:]]
//       return self.clone(args)
func TestCompileIfAction_SomeCondition(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	// Declare sort "t" in the signature
	tSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = tSort

	// Build: if some x:t. x = x { skip }
	xParam := cfg.NewAtom("x")
	xParam.ASort = cfg.NewAtom("t")
	xRef := cfg.NewAtom("x")
	fmla := cfg.NewAtom("=", xRef, xRef) // equality as atom with 2 args
	someCond := cfg.NewSome([]ast.Node{xParam}, fmla)
	thenBody := cfg.NewAtom("true") // trivial then branch

	result, err := c.CompileIf(someCond, thenBody, nil)
	if err != nil {
		t.Fatalf("CompileIf with Some condition should not error, got: %v", err)
	}

	// The result should be an IfAction whose condition is a Some-like construct
	// (compiled existential). In Python, the condition stays as a Some node
	// with compiled children.
	ifAct, ok := result.(*actions.IfAction)
	if !ok {
		// Currently this may fail because CompileIf doesn't detect Some
		t.Fatalf("expected *actions.IfAction, got %T: %v", result, result)
	}

	// The condition (first arg) of the IfAction should reflect the existential.
	condArgs := ifAct.ActionArgs()
	if len(condArgs) < 1 {
		t.Fatal("IfAction should have at least 1 arg (condition)")
	}

	// The condition should be a Some (existential) — not a plain boolean formula.
	condStr := condArgs[0].String()
	if !strings.Contains(condStr, "some") && !strings.Contains(condStr, "Some") {
		t.Errorf("condition should represent an existential (Some), got: %s", condStr)
	}
}

// TestCompileIfAction_SomeMinMaxCondition tests existential-if with a
// minimizing index: "if some x:t. phi minimizing idx { ... }"
//
// Python adds sargs.append(sortify_with_inference(self.args[0].index()))
// for SomeMinMax.
func TestCompileIfAction_SomeMinMaxCondition(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	tSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = tSort
	idxSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts["nat"] = idxSort
	// Python: idx must be a known symbol in the sig for sortify_with_inference to compile it.
	// Python's find_symbol raises "unknown symbol" if the name is not in sig.symbols.
	c.Sig.Symbols["idx"] = &il.SymbolEntry{Name: "idx", Sort: idxSort}

	// Build: if some x:t. x = x minimizing idx { skip }
	xParam := cfg.NewAtom("x")
	xParam.ASort = cfg.NewAtom("t")
	xRef := cfg.NewAtom("x")
	fmla := cfg.NewAtom("=", xRef, xRef)
	idxExpr := cfg.NewAtom("idx")

	// SomeMin is the AST node for "some ... minimizing ..."
	someCond := &ast.SomeMin{
		Params: []ast.Node{xParam},
		Fmla:   fmla,
		Index:  idxExpr,
	}

	thenBody := cfg.NewAtom("true")

	result, err := c.CompileIf(someCond, thenBody, nil)
	if err != nil {
		t.Fatalf("CompileIf with SomeMin condition should not error, got: %v", err)
	}

	ifAct, ok := result.(*actions.IfAction)
	if !ok {
		t.Fatalf("expected *actions.IfAction, got %T: %v", result, result)
	}

	// The compiled condition should contain the minimizing index.
	condArgs := ifAct.ActionArgs()
	if len(condArgs) < 1 {
		t.Fatal("IfAction should have at least 1 arg")
	}

	condStr := condArgs[0].String()
	// Must contain evidence of: (a) existential some, (b) minimizing index
	if !strings.Contains(condStr, "minim") && !strings.Contains(condStr, "Min") && !strings.Contains(condStr, "idx") {
		t.Errorf("SomeMin condition should include index reference, got: %s", condStr)
	}
}

// TestCompileIfAction_SomeMaxCondition tests existential-if with maximizing.
func TestCompileIfAction_SomeMaxCondition(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	tSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = tSort
	// Python: idx must be a known symbol for sortify_with_inference to compile it.
	c.Sig.Symbols["idx"] = &il.SymbolEntry{Name: "idx", Sort: tSort}

	xParam := cfg.NewAtom("x")
	xParam.ASort = cfg.NewAtom("t")
	xRef := cfg.NewAtom("x")
	fmla := cfg.NewAtom("=", xRef, xRef)
	idxExpr := cfg.NewAtom("idx")

	someCond := &ast.SomeMax{
		Params: []ast.Node{xParam},
		Fmla:   fmla,
		Index:  idxExpr,
	}

	thenBody := cfg.NewAtom("true")

	result, err := c.CompileIf(someCond, thenBody, nil)
	if err != nil {
		t.Fatalf("CompileIf with SomeMax condition should not error, got: %v", err)
	}

	ifAct, ok := result.(*actions.IfAction)
	if !ok {
		t.Fatalf("expected *actions.IfAction, got %T: %v", result, result)
	}

	condArgs := ifAct.ActionArgs()
	if len(condArgs) < 1 {
		t.Fatal("IfAction should have at least 1 arg")
	}

	condStr := condArgs[0].String()
	if !strings.Contains(condStr, "maxim") && !strings.Contains(condStr, "Max") && !strings.Contains(condStr, "idx") {
		t.Errorf("SomeMax condition should include index reference, got: %s", condStr)
	}
}

// TestCompileIfAction_SomeWithElse tests existential-if with an else branch.
// Python: args += [a.compile() for a in self.args[2:]]
func TestCompileIfAction_SomeWithElse(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	tSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = tSort

	xParam := cfg.NewAtom("x")
	xParam.ASort = cfg.NewAtom("t")
	fmla := cfg.NewAtom("=", cfg.NewAtom("x"), cfg.NewAtom("x"))
	someCond := cfg.NewSome([]ast.Node{xParam}, fmla)

	thenBody := cfg.NewAtom("true")
	elseBody := cfg.NewAtom("false")

	result, err := c.CompileIf(someCond, thenBody, elseBody)
	if err != nil {
		t.Fatalf("CompileIf with Some + else should not error, got: %v", err)
	}

	ifAct, ok := result.(*actions.IfAction)
	if !ok {
		t.Fatalf("expected *actions.IfAction, got %T: %v", result, result)
	}

	// Must have 3 args: condition, then, else
	args := ifAct.ActionArgs()
	if len(args) < 3 {
		t.Errorf("existential if-else should have 3 args (cond, then, else), got %d", len(args))
	}
}

// TestCompileIfAction_SomeParamsCompiledWithSigCopy tests that the Some
// parameters are compiled within a COPIED signature, so they don't leak
// into the outer scope.
//
// Python:
//   sig = ivy_logic.sig.copy()
//   with sig:
//       cls = [compile_const(v, sig) for v in ls]
func TestCompileIfAction_SomeParamsCompiledWithSigCopy(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	tSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = tSort

	xParam := cfg.NewAtom("x")
	xParam.ASort = cfg.NewAtom("t")
	fmla := cfg.NewAtom("=", cfg.NewAtom("x"), cfg.NewAtom("x"))
	someCond := cfg.NewSome([]ast.Node{xParam}, fmla)

	thenBody := cfg.NewAtom("true")

	// Record symbols before compilation
	symsBefore := len(c.Sig.Symbols)

	_, err := c.CompileIf(someCond, thenBody, nil)
	if err != nil {
		t.Fatalf("CompileIf error: %v", err)
	}

	// The bound variable "x" should NOT have leaked into the outer signature.
	// Python uses sig.copy() to prevent this.
	symsAfter := len(c.Sig.Symbols)
	if _, found := c.Sig.Symbols["x"]; found {
		t.Errorf("bound variable 'x' leaked into outer signature (before=%d, after=%d)", symsBefore, symsAfter)
	}
}

// TestCompileIfAction_SomeMultipleParams tests existential with multiple
// bound variables: "if some x:t, y:t. phi { ... }"
func TestCompileIfAction_SomeMultipleParams(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	tSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = tSort

	xParam := cfg.NewAtom("x")
	xParam.ASort = cfg.NewAtom("t")
	yParam := cfg.NewAtom("y")
	yParam.ASort = cfg.NewAtom("t")
	fmla := cfg.NewAtom("=", cfg.NewAtom("x"), cfg.NewAtom("y"))
	someCond := cfg.NewSome([]ast.Node{xParam, yParam}, fmla)

	thenBody := cfg.NewAtom("true")

	result, err := c.CompileIf(someCond, thenBody, nil)
	if err != nil {
		t.Fatalf("CompileIf with multi-param Some should not error, got: %v", err)
	}

	ifAct, ok := result.(*actions.IfAction)
	if !ok {
		t.Fatalf("expected *actions.IfAction, got %T: %v", result, result)
	}

	// Condition should be an existential with both x and y compiled
	condStr := ifAct.ActionArgs()[0].String()
	if !strings.Contains(condStr, "some") && !strings.Contains(condStr, "Some") {
		t.Errorf("multi-param existential condition not preserved, got: %s", condStr)
	}
}

// TestCompileIfAction_SomeThenBranchUsesExistentialVar tests Bug 1:
// The then-branch must be compiled INSIDE the sig scope so that
// existentially bound variables are visible during then-branch compilation.
//
// Python line 622: self.args[1].compile() is inside `with sig:`
func TestCompileIfAction_SomeThenBranchUsesExistentialVar(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	tSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = tSort

	// Build: if some x:t. x = x { assert x = x }
	// The then-branch references "x" which is only visible inside the sig copy.
	xParam := cfg.NewAtom("x")
	xParam.ASort = cfg.NewAtom("t")
	fmla := cfg.NewAtom("=", cfg.NewAtom("x"), cfg.NewAtom("x"))
	someCond := cfg.NewSome([]ast.Node{xParam}, fmla)

	// Then-branch references "x" — must succeed because x is in scope
	thenBody := cfg.NewAtom("=", cfg.NewAtom("x"), cfg.NewAtom("x"))

	result, err := c.CompileIf(someCond, thenBody, nil)
	if err != nil {
		t.Fatalf("CompileIf should succeed when then-branch uses existential var, got: %v", err)
	}

	ifAct, ok := result.(*actions.IfAction)
	if !ok {
		t.Fatalf("expected *actions.IfAction, got %T", result)
	}

	// Verify the then-branch compiled and is present
	if ifAct.ThenBody == nil {
		t.Fatal("then-branch should not be nil")
	}

	// "x" should NOT leak into the outer sig
	if _, found := c.Sig.Symbols["x"]; found {
		t.Errorf("bound variable 'x' leaked into outer signature")
	}
}

// TestCompileWhile_SomeCondition tests Bug 2:
// CompileWhile with a Some condition should compile like the if-Some path
// plus invariants.
//
// Python:
//   if isinstance(self.args[0], ivy_ast.Some):
//       res = compile_if_action(self.clone(self.args[:2]))
//       invars = list(map(sortify_with_inference, self.args[2:]))
//       return res.clone(res.args + invars)
func TestCompileWhile_SomeCondition(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	tSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = tSort

	// Build: while some x:t. x = x { skip } invariant true
	xParam := cfg.NewAtom("x")
	xParam.ASort = cfg.NewAtom("t")
	fmla := cfg.NewAtom("=", cfg.NewAtom("x"), cfg.NewAtom("x"))
	someCond := cfg.NewSome([]ast.Node{xParam}, fmla)

	bodyNode := cfg.NewAtom("true")
	invNode := cfg.NewAtom("true") // invariant

	result, err := c.CompileWhile(someCond, bodyNode, []ast.Node{invNode})
	if err != nil {
		t.Fatalf("CompileWhile with Some condition should not error, got: %v", err)
	}

	// Result should be a WhileAction (not IfAction)
	whileAct, ok := result.(*actions.WhileAction)
	if !ok {
		t.Fatalf("expected *actions.WhileAction, got %T: %v", result, result)
	}

	// The condition should be a SomeCondition
	if whileAct.Cond == nil {
		t.Fatal("WhileAction condition should not be nil")
	}
	_, isSome := whileAct.Cond.(*actions.SomeCondition)
	if !isSome {
		t.Errorf("WhileAction condition should be *actions.SomeCondition, got %T", whileAct.Cond)
	}

	// Should have 1 invariant
	if len(whileAct.Invariants) != 1 {
		t.Errorf("expected 1 invariant, got %d", len(whileAct.Invariants))
	}
}

// TestCompileWhile_SomeMinCondition tests Bug 2 for SomeMin variant.
func TestCompileWhile_SomeMinCondition(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	tSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = tSort
	c.Sig.Symbols["idx"] = &il.SymbolEntry{Name: "idx", Sort: tSort}

	xParam := cfg.NewAtom("x")
	xParam.ASort = cfg.NewAtom("t")
	fmla := cfg.NewAtom("=", cfg.NewAtom("x"), cfg.NewAtom("x"))
	someCond := &ast.SomeMin{
		Params: []ast.Node{xParam},
		Fmla:   fmla,
		Index:  cfg.NewAtom("idx"),
	}

	bodyNode := cfg.NewAtom("true")

	result, err := c.CompileWhile(someCond, bodyNode, nil)
	if err != nil {
		t.Fatalf("CompileWhile with SomeMin condition should not error, got: %v", err)
	}

	whileAct, ok := result.(*actions.WhileAction)
	if !ok {
		t.Fatalf("expected *actions.WhileAction, got %T: %v", result, result)
	}

	// The condition should be a SomeCondition with kind "some_min"
	someCd, ok := whileAct.Cond.(*actions.SomeCondition)
	if !ok {
		t.Fatalf("condition should be *actions.SomeCondition, got %T", whileAct.Cond)
	}
	if someCd.Kind != "some_min" {
		t.Errorf("expected kind 'some_min', got %q", someCd.Kind)
	}
}

// ============================================================================
// §6.2 #14: compile_thunk_action — subtype/destructor/substitution logic
// ============================================================================
//
// Python compile_thunk_action does 9 steps:
// 1. Copy sig, compile formals from args[0] + args[1]
// 2. Compile body via sortify, set formal_params/formal_returns
// 3. Collect fml:/loc: symbols from body that are in ivy_logic.sig.symbols
// 4. Get subtypename from args[0].relname, find subsort
// 5. Create $self parameter of subsort
// 6. Create destructor symbols (FunctionSort with subsort as first domain)
// 7. Register destructors in module.destructor_sorts / module.sort_destructors
// 8. Substitute captured vars → destructor($self) in body, register run action
// 9. Build LocalAction wrapping assignments + continuation
//
// Current Go stub skips steps 3-9.

// thunkNodeWith5Args creates a wrapper AST node that provides 5 args to
// CompileThunkAction (matching Python's ThunkAction which has args[0..4]):
//   args[0] = label (subtypename)
//   args[1] = action name
//   args[2] = sort
//   args[3] = body
//   args[4] = continuation
type thunkWith5Args struct {
	ast.Base
	inner       *ast.ThunkAction
	continuation ast.Node
}

func newThunkWith5Args(label, action, sort, body, cont ast.Node) *thunkWith5Args {
	cfg := ast.NewAstConfig()
	return &thunkWith5Args{
		inner:        cfg.NewThunkAction(label, action, sort, body),
		continuation: cont,
	}
}

func (t *thunkWith5Args) Args() []ast.Node {
	return []ast.Node{t.inner.Label, t.inner.Action, t.inner.Sort, t.inner.Body, t.continuation}
}

func (t *thunkWith5Args) Clone(args []ast.Node) ast.Node {
	cfg := t.Cfg
	return &thunkWith5Args{
		Base:         t.Base,
		inner:        cfg.NewThunkAction(args[0], args[1], args[2], args[3]),
		continuation: args[4],
	}
}

func (t *thunkWith5Args) String() string {
	return t.inner.String() + " ; " + fmt.Sprint(t.continuation)
}

// TestCompileThunkAction_RegistersRunAction tests that compile_thunk_action
// registers a "<subtypename>.run" action on the module.
//
// Python:
//   subtyperun = iu.compose_names(subtypename, 'run')
//   im.module.actions[subtyperun] = body
func TestCompileThunkAction_RegistersRunAction(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	// Declare thunk sort
	thunkSort := &lg.UninterpretedSort{Name: "handler"}
	c.Sig.Sorts["handler"] = thunkSort

	// Build ThunkAction with 5 args (including continuation)
	node := newThunkWith5Args(
		cfg.NewAtom("handler"),  // label/subtypename
		cfg.NewAtom("callback"), // action name
		cfg.NewAtom("handler"),  // sort
		cfg.NewAtom("true"),     // body
		cfg.NewAtom("true"),     // continuation
	)

	_, err := c.CompileThunkAction(node)
	if err != nil {
		t.Fatalf("CompileThunkAction error: %v", err)
	}

	// Check that "handler.run" was registered as an action
	runName := "handler.run"
	if _, found := c.Module.Actions[runName]; !found {
		t.Errorf("expected action %q to be registered on module, but it wasn't", runName)
		t.Logf("module actions: %v", mapKeys(c.Module.Actions))
	}
}

// TestCompileThunkAction_CreatesDestructors tests that destructors are
// created for captured fml:/loc: variables.
//
// Python:
//   for sym in syms:
//       dsort = FunctionSort(*([subsort] + sym.sort.dom + [sym.sort.rng]))
//       dsym = Symbol(compose_names(subtypename, sym.name[4:]), dsort)
//       module.destructor_sorts[dsym.name] = subsort
//       module.sort_destructors[subsort.name].append(dsym)
func TestCompileThunkAction_CreatesDestructors(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	thunkSort := &lg.UninterpretedSort{Name: "handler"}
	c.Sig.Sorts["handler"] = thunkSort

	// Add a captured variable "fml:x" to the signature
	xSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = xSort
	c.Sig.Symbols["fml:x"] = &il.SymbolEntry{Name: "fml:x", Sort: xSort}

	node := newThunkWith5Args(
		cfg.NewAtom("handler"),
		cfg.NewAtom("callback"),
		cfg.NewAtom("handler"),
		cfg.NewAtom("fml:x"), // body references captured var
		cfg.NewAtom("true"),  // continuation
	)

	_, err := c.CompileThunkAction(node)
	if err != nil {
		t.Fatalf("CompileThunkAction error: %v", err)
	}

	// Check destructor was created: "handler.x" should be in DestructorSorts
	destrName := "handler.x"
	if _, found := c.Module.DestructorSorts[destrName]; !found {
		t.Errorf("expected destructor %q in module.DestructorSorts, not found", destrName)
		t.Logf("DestructorSorts: %v", c.Module.DestructorSorts)
	}

	// Check sort_destructors for the handler sort
	if destrs, found := c.Module.SortDestructors["handler"]; !found || len(destrs) == 0 {
		t.Errorf("expected sort_destructors[%q] to contain at least 1 destructor", "handler")
	}
}

// TestCompileThunkAction_SelfParam tests that a $self parameter is inserted
// into the body's formal_params.
//
// Python:
//   selfparam = ivy_logic.Symbol('$self', subsort)
//   body.formal_params.insert(len(body.formal_params), selfparam)
func TestCompileThunkAction_SelfParam(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	thunkSort := &lg.UninterpretedSort{Name: "handler"}
	c.Sig.Sorts["handler"] = thunkSort

	node := newThunkWith5Args(
		cfg.NewAtom("handler"),
		cfg.NewAtom("callback"),
		cfg.NewAtom("handler"),
		cfg.NewAtom("true"),
		cfg.NewAtom("true"),
	)

	_, err := c.CompileThunkAction(node)
	if err != nil {
		t.Fatalf("CompileThunkAction error: %v", err)
	}

	// The registered run action should have $self in its formal_params
	// Python: body.formal_params.insert(len(body.formal_params), selfparam)
	runName := "handler.run"
	runAction, found := c.Module.Actions[runName]
	if !found {
		t.Fatalf("action %q not registered", runName)
	}

	// Check that the run action has $self in its formal params
	act, ok := runAction.(actions.Action)
	if !ok {
		t.Fatalf("run action should be an Action, got %T", runAction)
	}
	fp := act.GetFormalParams()
	hasSelf := false
	for _, p := range fp {
		if p.Name == "$self" {
			hasSelf = true
			break
		}
	}
	if !hasSelf {
		t.Errorf("expected $self in formal params, got: %v", fp)
	}
}

// TestCompileThunkAction_SubstitutionInBody tests that captured variables
// in the body are substituted with destructor applications.
//
// Python:
//   subs[sym] = dsym(selfparam)
//   new_body = lu.substitute_constants_ast(body, subs)
func TestCompileThunkAction_SubstitutionInBody(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	thunkSort := &lg.UninterpretedSort{Name: "handler"}
	c.Sig.Sorts["handler"] = thunkSort
	xSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = xSort
	c.Sig.Symbols["fml:x"] = &il.SymbolEntry{Name: "fml:x", Sort: xSort}

	node := newThunkWith5Args(
		cfg.NewAtom("handler"),
		cfg.NewAtom("callback"),
		cfg.NewAtom("handler"),
		cfg.NewAtom("fml:x"), // references captured var
		cfg.NewAtom("true"),
	)

	_, err := c.CompileThunkAction(node)
	if err != nil {
		t.Fatalf("CompileThunkAction error: %v", err)
	}

	// The registered run action's body should reference handler.x($self)
	// instead of the raw fml:x symbol.
	runName := "handler.run"
	runAction, found := c.Module.Actions[runName]
	if !found {
		t.Fatalf("action %q not registered", runName)
	}

	bodyStr := fmt.Sprintf("%v", runAction)
	if strings.Contains(bodyStr, "fml:x") {
		t.Errorf("body should have fml:x substituted with destructor, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "handler.x") && !strings.Contains(bodyStr, "$self") {
		t.Errorf("body should reference handler.x($self), got: %s", bodyStr)
	}
}

// TestCompileThunkAction_ReturnsLocalAction tests that the result is a
// LocalAction wrapping destructor assignments + continuation.
//
// Python:
//   lsym = add_symbol('loc:' + self.args[1].relname, subsort)
//   asgns = [AssignAction(dsym(lsym), sym) for sym, dsym in zip(syms, dsyms)]
//   res = LocalAction(lsym, Sequence(*(asgns + [cont])))
func TestCompileThunkAction_ReturnsLocalAction(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	thunkSort := &lg.UninterpretedSort{Name: "handler"}
	c.Sig.Sorts["handler"] = thunkSort
	xSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = xSort
	c.Sig.Symbols["fml:x"] = &il.SymbolEntry{Name: "fml:x", Sort: xSort}

	node := newThunkWith5Args(
		cfg.NewAtom("handler"),
		cfg.NewAtom("callback"),
		cfg.NewAtom("handler"),
		cfg.NewAtom("fml:x"),
		cfg.NewAtom("true"),
	)

	result, err := c.CompileThunkAction(node)
	if err != nil {
		t.Fatalf("CompileThunkAction error: %v", err)
	}

	// Result should be a LocalAction (or unwrapped from one)
	act, _ := result.(actions.Action)
	if act == nil {
		t.Fatalf("result should be an action, got: %T", result)
	}

	localAct, ok := act.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T: %v", act, act)
	}

	// The LocalAction should contain local var + sequence body
	subacts := localAct.ActionArgs()
	if len(subacts) < 2 {
		t.Errorf("LocalAction should have local var + sequence body, got %d args", len(subacts))
	}
}

// TestCompileThunkAction_DestructorSort tests that each destructor has the
// correct FunctionSort: domain starts with the subsort, followed by the
// captured symbol's domain, with the captured symbol's range.
//
// Python:
//   dsort = FunctionSort(*([subsort] + sym.sort.dom + [sym.sort.rng]))
func TestCompileThunkAction_DestructorSort(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	thunkSort := &lg.UninterpretedSort{Name: "handler"}
	c.Sig.Sorts["handler"] = thunkSort
	xSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = xSort
	c.Sig.Symbols["fml:x"] = &il.SymbolEntry{Name: "fml:x", Sort: xSort}

	node := newThunkWith5Args(
		cfg.NewAtom("handler"),
		cfg.NewAtom("callback"),
		cfg.NewAtom("handler"),
		cfg.NewAtom("fml:x"),
		cfg.NewAtom("true"),
	)

	_, err := c.CompileThunkAction(node)
	if err != nil {
		t.Fatalf("CompileThunkAction error: %v", err)
	}

	// The destructor "handler.x" should have sort: handler -> t
	destrs := c.Module.SortDestructors["handler"]
	if len(destrs) == 0 {
		t.Fatal("no destructors found for handler sort")
	}

	var found *lg.Symbol
	for _, d := range destrs {
		if d.Name == "handler.x" {
			found = d
			break
		}
	}
	if found == nil {
		t.Fatal("destructor handler.x not found in sort_destructors")
	}

	fs, ok := found.NodeSort().(*lg.FunctionSort)
	if !ok {
		t.Fatalf("destructor sort should be FunctionSort, got %T", found.NodeSort())
	}
	dom := fs.Domain()
	if len(dom) != 1 {
		t.Errorf("destructor domain should have 1 element (handler sort), got %d", len(dom))
	} else if dom[0].String() != "handler" {
		t.Errorf("destructor domain[0] should be handler, got %s", dom[0])
	}
	if fs.Range().String() != "t" {
		t.Errorf("destructor range should be t, got %s", fs.Range())
	}
}

// TestCompileThunkAction_LocalVarHasCorrectSort tests that the local variable
// created for the thunk result has the subsort type.
//
// Python:
//   lsym = add_symbol('loc:' + self.args[1].relname, subsort)
func TestCompileThunkAction_LocalVarHasCorrectSort(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	thunkSort := &lg.UninterpretedSort{Name: "handler"}
	c.Sig.Sorts["handler"] = thunkSort

	node := newThunkWith5Args(
		cfg.NewAtom("handler"),
		cfg.NewAtom("callback"),
		cfg.NewAtom("handler"),
		cfg.NewAtom("true"),
		cfg.NewAtom("true"),
	)

	result, err := c.CompileThunkAction(node)
	if err != nil {
		t.Fatalf("CompileThunkAction error: %v", err)
	}

	act, _ := result.(actions.Action)
	localAct, ok := act.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", act)
	}

	// First arg of LocalAction is the local variable
	args := localAct.ActionArgs()
	if len(args) == 0 {
		t.Fatal("LocalAction has no args")
	}

	sym, ok := args[0].(*lg.Symbol)
	if !ok {
		t.Fatalf("first arg of LocalAction should be a Symbol, got %T", args[0])
	}

	// The local variable should have sort "handler" (the subsort)
	if sym.NodeSort().String() != "handler" {
		t.Errorf("local var sort should be 'handler', got %s", sym.NodeSort())
	}

	// The local variable should be named "loc:callback"
	if sym.Name != "loc:callback" {
		t.Errorf("local var should be named 'loc:callback', got %s", sym.Name)
	}
}

// TestCompileThunkAction_AssignmentsInResult tests that the result Sequence
// contains assignment actions for each captured destructor.
//
// Python:
//   asgns = [AssignAction(dsym(lsym), sym) for sym, dsym in zip(syms, dsyms)]
//   res = LocalAction(lsym, Sequence(*(asgns + [cont])))
func TestCompileThunkAction_AssignmentsInResult(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	thunkSort := &lg.UninterpretedSort{Name: "handler"}
	c.Sig.Sorts["handler"] = thunkSort
	xSort := &lg.UninterpretedSort{Name: "t"}
	c.Sig.Sorts["t"] = xSort
	c.Sig.Symbols["fml:x"] = &il.SymbolEntry{Name: "fml:x", Sort: xSort}

	node := newThunkWith5Args(
		cfg.NewAtom("handler"),
		cfg.NewAtom("callback"),
		cfg.NewAtom("handler"),
		cfg.NewAtom("fml:x"),
		cfg.NewAtom("true"),
	)

	result, err := c.CompileThunkAction(node)
	if err != nil {
		t.Fatalf("CompileThunkAction error: %v", err)
	}

	act, _ := result.(actions.Action)
	localAct, ok := act.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", act)
	}

	// The body of LocalAction should be a Sequence containing assignments
	args := localAct.ActionArgs()
	if len(args) < 2 {
		t.Fatal("LocalAction should have local var + sequence")
	}

	seqExpr := args[1]
	seqAct, _ := seqExpr.(actions.Action)
	seq, ok := seqAct.(*actions.Sequence)
	if !ok {
		t.Fatalf("expected Sequence in LocalAction body, got %T", seqAct)
	}

	// Should have at least 1 assignment (for fml:x) + 1 continuation
	seqArgs := seq.ActionArgs()
	if len(seqArgs) < 2 {
		t.Errorf("Sequence should have >= 2 elements (assignments + continuation), got %d", len(seqArgs))
	}

	// First element should be an AssignAction
	if len(seqArgs) > 0 {
		firstAct, _ := seqArgs[0].(actions.Action)
		if _, ok := firstAct.(*actions.AssignAction); !ok {
			t.Errorf("first Sequence element should be AssignAction, got %T", firstAct)
		}
	}
}

// TestCompileThunkAction_LocSymbolNotLeaked tests that "loc:" symbols created
// by the thunk compilation don't pollute the outer signature.
func TestCompileThunkAction_LocSymbolNotLeaked(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	thunkSort := &lg.UninterpretedSort{Name: "handler"}
	c.Sig.Sorts["handler"] = thunkSort

	node := newThunkWith5Args(
		cfg.NewAtom("handler"),
		cfg.NewAtom("callback"),
		cfg.NewAtom("handler"),
		cfg.NewAtom("true"),
		cfg.NewAtom("true"),
	)

	_, err := c.CompileThunkAction(node)
	if err != nil {
		t.Fatalf("CompileThunkAction error: %v", err)
	}

	// Python adds loc:callback to a sig copy for continuation compilation.
	// The outer sig should not have loc:callback.
	if _, found := c.Sig.Symbols["loc:callback"]; found {
		t.Errorf("loc:callback leaked into outer signature")
	}
}

// TestCompileThunkAction_PreservesLineno tests that the result LocalAction
// preserves the lineno from the thunk node.
//
// Python: res.lineno = self.lineno
func TestCompileThunkAction_PreservesLineno(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	thunkSort := &lg.UninterpretedSort{Name: "handler"}
	c.Sig.Sorts["handler"] = thunkSort

	node := newThunkWith5Args(
		cfg.NewAtom("handler"),
		cfg.NewAtom("callback"),
		cfg.NewAtom("handler"),
		cfg.NewAtom("true"),
		cfg.NewAtom("true"),
	)
	node.SetLineno(ast.Location{Line: 42})

	result, err := c.CompileThunkAction(node)
	if err != nil {
		t.Fatalf("CompileThunkAction error: %v", err)
	}

	act, _ := result.(actions.Action)
	localAct, ok := act.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", act)
	}

	if localAct.GetLineno().Line != 42 {
		t.Errorf("expected lineno 42, got %d", localAct.GetLineno().Line)
	}
}

// ============================================================================
// Helpers
// ============================================================================

func mapKeys(m map[string]module.Action) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
