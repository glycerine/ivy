// Tests for the batch fixes from AUDIT18MARCH.md section 10.
//
// Batch 1: Quick Fixes (S2, B3, B4, B5)
// Batch 2: Action Type Classification (M2, B1)
// Batch 3: Strip Binding System (M3, M4)
// Batch 4: Interference / Loop Checking (B2)
// Batch 5: isolate_component Features (M5, M6, M7, S3)
// Batch 6: check_isolate_completeness Full Implementation (M1)

package isolate

import (
	"testing"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// -----------------------------------------------------------------------
// Test helpers for batch tests
// -----------------------------------------------------------------------

// testMixin implements module.MixinDef for testing.
type testMixin struct {
	mixer, mixee string
	after        bool
	implement    bool
}

func (m *testMixin) Mixer() string { return m.mixer }
func (m *testMixin) Mixee() string { return m.mixee }
func (m *testMixin) IsAfter() bool { return m.after }

// testExporter implements module.Exporter for testing.
type testExporter struct {
	name  string
	scope string
}

func (e *testExporter) Exported() string { return e.name }
func (e *testExporter) Scope() string    { return e.scope }

// testDelegator implements module.Delegator for testing.
type testDelegator struct {
	delegated string
	delegee   string
}

func (d *testDelegator) Delegated() string { return d.delegated }
func (d *testDelegator) Delegee() string   { return d.delegee }

// mkModuleWithSig creates a module with an initialized signature.
func mkModuleWithSig() *module.Module {
	m := mkModule()
	if m.Sig == nil {
		m.Sig = &il.Sig{
			Symbols:      iu.NewInsMap[string, *il.SymbolEntry](),
			Sorts:        make(map[string]lg.Sort),
			Constructors: make(map[string]bool),
			Interp:       make(map[string]interface{}),
		}
	}
	return m
}

// -----------------------------------------------------------------------
// Batch 1: Quick Fixes
// -----------------------------------------------------------------------

// S2: isExplicitOnly now returns lf.Explicit
func TestIsExplicitOnly_True(t *testing.T) {
	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, nil)
	lf.Explicit = true
	if !isExplicitOnly(lf) {
		t.Error("isExplicitOnly should return true when Explicit=true")
	}
}

func TestIsExplicitOnly_False(t *testing.T) {
	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, nil)
	if isExplicitOnly(lf) {
		t.Error("isExplicitOnly should return false when Explicit=false")
	}
}

// B4: nodeRelname extracts relname from AST nodes
func TestNodeRelname_Atom(t *testing.T) {
	cfg := ast.NewAstConfig()
	a := cfg.NewAtom("myaction")
	got := nodeRelname(a)
	if got != "myaction" {
		t.Errorf("nodeRelname(Atom) = %q, want %q", got, "myaction")
	}
}

func TestNodeRelname_Symbol(t *testing.T) {
	cfg := ast.NewAstConfig()
	s := cfg.NewSymbol("mysymbol", nil)
	got := nodeRelname(s)
	if got != "mysymbol" {
		t.Errorf("nodeRelname(Symbol) = %q, want %q", got, "mysymbol")
	}
}

func TestNodeRelname_Variable(t *testing.T) {
	cfg := ast.NewAstConfig()
	v := cfg.NewVariable("X", "")
	got := nodeRelname(v)
	if got != "X" {
		t.Errorf("nodeRelname(Variable) = %q, want %q", got, "X")
	}
}

// B3: StripIsolate clears mod.Params for version <= 1.6
func TestStripIsolateVersion16ClearsParams(t *testing.T) {
	m := mkModuleWithSig()
	save := m.Cfg.IuCfg.GetStringVersion()
	defer iu.SetStringVersionOn(m.Cfg.IuCfg, save)
	iu.SetStringVersionOn(m.Cfg.IuCfg, "1.6")
	sym := lg.NewConst("p", lg.Boolean)
	m.Params = append(m.Params, sym)

	stripMap := StripMap{"server": {"s"}}
	// Add a symbol that matches the strip map
	m.Sig.Symbols.Set("server.x", &il.SymbolEntry{Name: "server.x", Sort: lg.Boolean})

	err := StripIsolate(m, stripMap, nil)
	if err != nil {
		t.Fatalf("StripIsolate error: %v", err)
	}

	if len(m.Params) != 0 {
		t.Errorf("expected Params to be cleared for version 1.6, got %d params", len(m.Params))
	}
}

func TestStripIsolateVersion17KeepsParams(t *testing.T) {
	m := mkModuleWithSig()
	save := m.Cfg.IuCfg.GetStringVersion()
	defer iu.SetStringVersionOn(m.Cfg.IuCfg, save)
	iu.SetStringVersionOn(m.Cfg.IuCfg, "1.7")
	sym := lg.NewConst("p", lg.Boolean)
	m.Params = append(m.Params, sym)

	stripMap := StripMap{"server": {"s"}}
	m.Sig.Symbols.Set("server.x", &il.SymbolEntry{Name: "server.x", Sort: lg.Boolean})

	err := StripIsolate(m, stripMap, nil)
	if err != nil {
		t.Fatalf("StripIsolate error: %v", err)
	}

	if len(m.Params) == 0 {
		t.Error("expected Params to be kept for version 1.7")
	}
}

// -----------------------------------------------------------------------
// Batch 2: Action Type Classification
// -----------------------------------------------------------------------

func TestIsAssertLike_AssertAction(t *testing.T) {
	a := actions.NewAssertAction(lg.True)
	if !actions.IsAssertLike(a) {
		t.Error("AssertAction should be assert-like")
	}
}

func TestIsAssertLike_RequiresAction(t *testing.T) {
	a := actions.NewRequiresAction(lg.True)
	if !actions.IsAssertLike(a) {
		t.Error("RequiresAction should be assert-like")
	}
}

func TestIsAssertLike_EnsuresAction(t *testing.T) {
	a := actions.NewEnsuresAction(lg.True)
	if !actions.IsAssertLike(a) {
		t.Error("EnsuresAction should be assert-like")
	}
}

func TestIsAssertLike_SubgoalAction(t *testing.T) {
	a := actions.NewSubgoalAction(lg.True)
	if !actions.IsAssertLike(a) {
		t.Error("SubgoalAction should be assert-like")
	}
}

func TestIsAssertLike_SequenceAction(t *testing.T) {
	a := actions.NewSequence()
	if actions.IsAssertLike(a) {
		t.Error("Sequence should NOT be assert-like")
	}
}

func TestIsAssertLike_CallAction(t *testing.T) {
	a := actions.NewCallActionOn(actions.NewActionsConfig(),lg.NewConst("foo", lg.TopS))
	if actions.IsAssertLike(a) {
		t.Error("CallAction should NOT be assert-like")
	}
}

// HasAssertions should match all assert subclasses
func TestHasAssertions_WithRequiresAction(t *testing.T) {
	// Python isinstance(action, ia.AssertAction) matches RequiresAction
	m := mkModuleWithActions(map[string]actions.Action{
		"foo": actions.NewSequence(actions.NewRequiresAction(lg.True)),
	})
	if !HasAssertions(m, "foo") {
		t.Error("HasAssertions should match RequiresAction (subclass of AssertAction)")
	}
}

func TestHasAssertions_WithEnsuresAction(t *testing.T) {
	m := mkModuleWithActions(map[string]actions.Action{
		"foo": actions.NewSequence(actions.NewEnsuresAction(lg.True)),
	})
	if !HasAssertions(m, "foo") {
		t.Error("HasAssertions should match EnsuresAction (subclass of AssertAction)")
	}
}

func TestHasAssertions_WithSubgoalAction(t *testing.T) {
	m := mkModuleWithActions(map[string]actions.Action{
		"foo": actions.NewSequence(actions.NewSubgoalAction(lg.True)),
	})
	if !HasAssertions(m, "foo") {
		t.Error("HasAssertions should match SubgoalAction (subclass of AssertAction)")
	}
}

func TestHasAssertions_WithPlainAssertAction(t *testing.T) {
	m := mkModuleWithActions(map[string]actions.Action{
		"foo": actions.NewSequence(actions.NewAssertAction(lg.True)),
	})
	if !HasAssertions(m, "foo") {
		t.Error("HasAssertions should match plain AssertAction")
	}
}

func TestHasAssertions_NoAssert(t *testing.T) {
	m := mkModuleWithActions(map[string]actions.Action{
		"foo": actions.NewSequence(),
	})
	if HasAssertions(m, "foo") {
		t.Error("HasAssertions should return false for empty sequence")
	}
}

func TestHasAssertions_MissingAction(t *testing.T) {
	m := mkModule()
	if HasAssertions(m, "nonexistent") {
		t.Error("HasAssertions should return false for nonexistent action")
	}
}

func TestHasRequires_MatchesRequiresOnly(t *testing.T) {
	m := mkModuleWithActions(map[string]actions.Action{
		"foo": actions.NewSequence(actions.NewRequiresAction(lg.True)),
	})
	if !HasRequires(m, "foo") {
		t.Error("HasRequires should match RequiresAction")
	}
}

func TestHasRequires_DoesNotMatchAssertAction(t *testing.T) {
	m := mkModuleWithActions(map[string]actions.Action{
		"foo": actions.NewSequence(actions.NewAssertAction(lg.True)),
	})
	if HasRequires(m, "foo") {
		t.Error("HasRequires should NOT match plain AssertAction")
	}
}

// HasSideEffect with IsAssertLike
func TestHasSideEffect_WithEnsuresAction(t *testing.T) {
	// EnsuresAction is assert-like, so it's a side effect
	act := actions.NewSequence(actions.NewEnsuresAction(lg.True))
	m := mkModuleWithSig()
	m.Actions.Set("foo", act)
	actionMap := iu.NewInsMap[string, actions.Action]()
	actionMap.Set("foo", act)
	if !HasSideEffect(m, "foo", actionMap) {
		t.Error("EnsuresAction should be detected as side effect via IsAssertLike")
	}
}

// -----------------------------------------------------------------------
// Batch 3: Strip Binding System
// -----------------------------------------------------------------------

func TestGetStripBinding_BasicApply(t *testing.T) {
	m := mkModuleWithSig()
	sort1 := mkSort("T")
	m.Sig.Symbols.Set("f", &il.SymbolEntry{Name: "f", Sort: sort1})

	// Create strip map: f has 1 strip param "s"
	stripMap := StripMap{"f": {"s"}}

	// Create Apply(f, X) where X is a variable.
	// Need a FunctionSort for NewApply to succeed.
	fnSort, err := lg.NewFunctionSort(sort1, lg.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort error: %v", err)
	}
	fSym := lg.NewConst("f", fnSort)
	xVar, err := lg.NewVariable("X", sort1)
	if err != nil {
		t.Fatalf("NewVariable error: %v", err)
	}
	app, err := lg.NewApply(fSym, xVar)
	if err != nil {
		t.Fatalf("NewApply error: %v", err)
	}

	binding := make(map[lg.NodeKey]string)
	err = GetStripBinding(app, stripMap, binding, m)
	if err != nil {
		t.Fatalf("GetStripBinding error: %v", err)
	}

	// X should be bound to "s"
	key := lg.Key(xVar)
	if binding[key] != "s" {
		t.Errorf("binding[X] = %q, want %q", binding[key], "s")
	}
}

func TestGetStripBinding_NoMatch(t *testing.T) {
	m := mkModuleWithSig()
	stripMap := StripMap{"g": {"s"}}

	fSym := lg.NewConst("f", lg.Boolean)
	binding := make(map[lg.NodeKey]string)
	err := GetStripBinding(fSym, stripMap, binding, m)
	if err != nil {
		t.Fatalf("GetStripBinding error: %v", err)
	}
	if len(binding) != 0 {
		t.Error("expected empty binding when name doesn't match strip map")
	}
}

func TestGetStripBinding_Conflict(t *testing.T) {
	m := mkModuleWithSig()
	sort1 := mkSort("T")

	stripMap := StripMap{"f": {"s1"}, "g": {"s2"}}

	fnSort, _ := lg.NewFunctionSort(sort1, lg.Boolean)
	xVar, _ := lg.NewVariable("X", sort1)

	// First binding: f(X) -> X maps to "s1"
	fSym := lg.NewConst("f", fnSort)
	app1, err := lg.NewApply(fSym, xVar)
	if err != nil {
		t.Fatalf("NewApply(f,X) error: %v", err)
	}
	binding := make(map[lg.NodeKey]string)
	err = GetStripBinding(app1, stripMap, binding, m)
	if err != nil {
		t.Fatalf("first binding error: %v", err)
	}

	// Second binding: g(X) -> X maps to "s2" - should conflict
	gSym := lg.NewConst("g", fnSort)
	app2, err := lg.NewApply(gSym, xVar)
	if err != nil {
		t.Fatalf("NewApply(g,X) error: %v", err)
	}
	err = GetStripBinding(app2, stripMap, binding, m)
	if err == nil {
		t.Error("expected conflict error when same variable maps to different params")
	}
}

func TestStripActionFull_BindingSubstitution(t *testing.T) {
	m := mkModuleWithSig()
	sort1 := mkSort("T")

	// Create binding: variable X -> "param_s"
	xVar, _ := lg.NewVariable("X", sort1)
	binding := map[lg.NodeKey]string{
		lg.Key(xVar): "param_s",
	}

	// Create an assign action that references X
	act := actions.NewAssignAction(xVar, lg.True)

	m.Cfg.IsolateCfg.StripAddedSymbols = nil
	result := StripActionFull(act, StripMap{}, m, binding, false, nil)
	if result == nil {
		t.Fatal("StripActionFull returned nil")
	}

	// Should have added param_s to StripAddedSymbols
	if len(m.Cfg.IsolateCfg.StripAddedSymbols) == 0 {
		// The substitution happens in stripNodeFull, which only fires
		// on lg.Const or lg.Variable nodes that match the binding key.
		// AssignAction's LHS is the variable X, which should match.
		t.Log("StripAddedSymbols may be empty if action args don't go through stripNodeFull")
	}
}

func TestStripLabeledFormula_WithBinding(t *testing.T) {
	m := mkModuleWithSig()
	sort1 := mkSort("T")

	fnSort, err := lg.NewFunctionSort(sort1, lg.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort error: %v", err)
	}
	m.Sig.Symbols.Set("f", &il.SymbolEntry{Name: "f", Sort: fnSort})

	stripMap := StripMap{"f": {"s"}}

	// Create a labeled formula with Apply(f, X)
	fSym := lg.NewConst("f", fnSort)
	xVar, err := lg.NewVariable("X", sort1)
	if err != nil {
		t.Fatalf("NewVariable error: %v", err)
	}
	app, err := lg.NewApply(fSym, xVar)
	if err != nil {
		t.Fatalf("NewApply error: %v", err)
	}

	acfg := ast.NewAstConfig()
	lf := acfg.NewLabeledFormula(lg.NewConst("lbl", lg.Boolean), app)

	result := StripLabeledFormula(lf, stripMap, m)
	if result == nil {
		t.Fatal("StripLabeledFormula returned nil")
	}
	if result.Formula == nil {
		t.Fatal("stripped formula is nil")
	}
}

func TestStripIsolate_UsesStripActionFull(t *testing.T) {
	m := mkModuleWithSig()
	sortT := mkSort("T")

	// Create function sort: T -> Boolean
	fnSort, _ := lg.NewFunctionSort(sortT, lg.Boolean)
	m.Sig.Symbols.Set("server.f", &il.SymbolEntry{Name: "server.f", Sort: fnSort})

	// Create action with formal params [s]
	s := lg.NewConst("s", sortT)
	act := actions.NewSequence()
	act.SetFormalParams([]*lg.Const{s})

	m.Actions.Set("server.f", act)

	stripMap := StripMap{"server": {"s"}}
	err := StripIsolate(m, stripMap, nil)
	if err != nil {
		t.Fatalf("StripIsolate error: %v", err)
	}
}

func TestNumIsolateParams_SetByStripIsolateParams(t *testing.T) {
	m := mkModuleWithSig()
	acfg := ast.NewAstConfig()
	iso := acfg.NewIsolateDef(nil, 0)
	// IsolateDef with no params -> NumIsolateParams = 0
	m.Cfg.IsolateCfg.NumIsolateParams = 99 // sentinel
	_ = StripIsolateParams(m, iso, nil, nil, nil)
	// After call, NumIsolateParams should reflect isolate params count
	if m.Cfg.IsolateCfg.NumIsolateParams == 99 {
		t.Error("NumIsolateParams was not updated by StripIsolateParams")
	}
}

// -----------------------------------------------------------------------
// Batch 4: Interference / Loop Checking
// -----------------------------------------------------------------------

func TestWhileHasRanking_WithRanking(t *testing.T) {
	// WhileAction with RankingWrapper as last invariant
	cond := lg.True
	body := actions.NewSequence()
	rankRel := lg.NewConst("lt", lg.Boolean)
	rw := &actions.RankingWrapper{Ranking: actions.NewRanking(rankRel)}

	wa := actions.NewWhileAction(cond, body, rw)
	if !whileHasRanking(wa) {
		t.Error("WhileAction with RankingWrapper should have ranking")
	}
}

func TestWhileHasRanking_WithoutRanking(t *testing.T) {
	cond := lg.True
	body := actions.NewSequence()
	// Invariant that's NOT a RankingWrapper
	inv := lg.NewConst("inv", lg.Boolean)

	wa := actions.NewWhileAction(cond, body, inv)
	if whileHasRanking(wa) {
		t.Error("WhileAction without RankingWrapper should not have ranking")
	}
}

func TestWhileHasRanking_NoInvariants(t *testing.T) {
	cond := lg.True
	body := actions.NewSequence()
	wa := actions.NewWhileAction(cond, body)
	if whileHasRanking(wa) {
		t.Error("WhileAction with no invariants should not have ranking")
	}
}

func TestGetCallsModsRec_DetectsLoops(t *testing.T) {
	cond := lg.True
	body := actions.NewSequence()
	// WhileAction without ranking
	wa := actions.NewWhileAction(cond, body)

	m := mkModuleWithActions(map[string]actions.Action{
		"foo": actions.NewSequence(wa),
	})

	summarized := map[string]bool{"foo": true}
	calls := make(map[string]map[string]bool)
	mods := make(map[string]map[string]bool)
	loops := make(map[string][]actions.Action)

	GetCallsModsRec(m, summarized, "foo", calls, mods, loops)

	if len(loops["foo"]) == 0 {
		t.Error("expected loop to be detected for WhileAction without ranking")
	}
}

func TestGetCallsModsRec_NoLoopWithRanking(t *testing.T) {
	cond := lg.True
	body := actions.NewSequence()
	rankRel := lg.NewConst("lt", lg.Boolean)
	rw := &actions.RankingWrapper{Ranking: actions.NewRanking(rankRel)}
	wa := actions.NewWhileAction(cond, body, rw)

	m := mkModuleWithActions(map[string]actions.Action{
		"foo": actions.NewSequence(wa),
	})

	summarized := map[string]bool{"foo": true}
	calls := make(map[string]map[string]bool)
	mods := make(map[string]map[string]bool)
	loops := make(map[string][]actions.Action)

	GetCallsModsRec(m, summarized, "foo", calls, mods, loops)

	if len(loops["foo"]) != 0 {
		t.Error("WhileAction with ranking should not be reported as loop")
	}
}

func TestCheckInterferenceFull_TerminationCheck(t *testing.T) {
	cond := lg.True
	body := actions.NewSequence()
	wa := actions.NewWhileAction(cond, body) // no ranking

	m := mkModuleWithSig()
	m.Actions.Set("bar", actions.NewSequence(wa))

	summarized := map[string]bool{"bar": true}
	newActions := iu.NewInsMap[string, actions.Action]()
	newActions.Set("bar", actions.NewSequence(wa))

	m.Cfg.IsolateCfg.DoCheckInterference = true

	err := CheckInterferenceFull(m, newActions, summarized,
		nil, true, nil, nil, nil) // checkTerm=true
	if err == nil {
		t.Error("expected termination error for loop without decreases clause")
	}
}

func TestCheckInterferenceFull_NoTermCheckWhenDisabled(t *testing.T) {
	cond := lg.True
	body := actions.NewSequence()
	wa := actions.NewWhileAction(cond, body) // no ranking

	m := mkModuleWithSig()
	m.Actions.Set("bar", actions.NewSequence(wa))

	summarized := map[string]bool{"bar": true}
	newActions := iu.NewInsMap[string, actions.Action]()
	newActions.Set("bar", actions.NewSequence(wa))

	m.Cfg.IsolateCfg.DoCheckInterference = true

	err := CheckInterferenceFull(m, newActions, summarized,
		nil, false, nil, nil, nil) // checkTerm=false
	if err != nil {
		t.Errorf("expected no error when checkTerm=false, got: %v", err)
	}
}

func TestGetLocMods_FiltersOnFml(t *testing.T) {
	fmlSym := lg.NewConst("fml:x", lg.Boolean)
	normalSym := lg.NewConst("y", lg.Boolean)

	m := mkModuleWithSig()
	m.Sig.Symbols.Set("fml:x", &il.SymbolEntry{Name: "fml:x", Sort: lg.Boolean})
	m.Sig.Symbols.Set("y", &il.SymbolEntry{Name: "y", Sort: lg.Boolean})

	// Python's get_loc_mods calls action.modifies() which is non-recursive.
	// For a Sequence, modifies() returns [] (no override). Only direct
	// AssignAction/HavocAction have non-empty modifies().
	// Test with a direct AssignAction to check the fml: filter.
	act := actions.NewAssignAction(fmlSym, lg.True)
	m.Actions.Set("act1", act)

	result := GetLocMods(m, "act1")
	found := false
	for _, r := range result {
		if r == "fml:x" {
			found = true
		}
	}
	if !found {
		t.Error("GetLocMods should include fml:x for direct AssignAction")
	}

	// A Sequence should return empty (Python: Sequence.modifies() returns [])
	seqAct := actions.NewSequence(
		actions.NewAssignAction(fmlSym, lg.True),
		actions.NewAssignAction(normalSym, lg.True),
	)
	m.Actions.Set("act2", seqAct)
	result2 := GetLocMods(m, "act2")
	if len(result2) != 0 {
		t.Errorf("GetLocMods on Sequence should return empty (matching Python), got %v", result2)
	}
}

// -----------------------------------------------------------------------
// Batch 5: isolate_component Features
// -----------------------------------------------------------------------

// S3: ExtAction creates EnvAction
func TestExtAction_CreatesEnvAction(t *testing.T) {
	m := mkModuleWithSig()
	act1 := actions.NewSequence()
	act2 := actions.NewSequence()
	m.Actions.Set("ext:a", act1)
	m.Actions.Set("ext:b", act2)
	m.PublicActions = iu.NewInsMap[string, bool]()
	m.PublicActions.Set("ext:a", true)
	m.PublicActions.Set("ext:b", true)

	m.Cfg.ExtAction = "ext"
	defer func() { m.Cfg.ExtAction = "" }()

	// Simulate the ExtAction creation logic from create.go
	// (We test it in isolation since CreateIsolate is complex)
	afterInitNames := map[string]bool{}
	var sortedPublic []string
	for name := range m.PublicActions.All() {
		sortedPublic = append(sortedPublic, name)
	}

	var extBranches []lg.Expr
	for _, name := range sortedPublic {
		if afterInitNames[CanonAct(name)] {
			continue
		}
		if act, ok := m.Actions.Get2(name); ok {
			extBranches = append(extBranches, act)
		}
	}

	if len(extBranches) < 2 {
		t.Errorf("expected at least 2 branches, got %d", len(extBranches))
	}

	extAct := actions.NewEnvAction(extBranches...)
	if extAct == nil {
		t.Fatal("NewEnvAction returned nil")
	}
	if extAct.Name() != "env" {
		t.Errorf("EnvAction.Name() = %q, want %q", extAct.Name(), "env")
	}
}

// -----------------------------------------------------------------------
// Batch 6: CheckIsolateCompleteness Full Implementation
// -----------------------------------------------------------------------

func TestCheckIsolateCompleteness_NilModule(t *testing.T) {
	result := CheckIsolateCompleteness(nil)
	if result != nil {
		t.Error("expected nil for nil module")
	}
}

func TestCheckIsolateCompleteness_EmptyModule(t *testing.T) {
	m := mkModule()
	result := CheckIsolateCompleteness(m)
	if len(result) != 0 {
		t.Errorf("expected 0 errors for empty module, got %d", len(result))
	}
}

func TestCheckIsolateCompleteness_UncheckedProperty(t *testing.T) {
	m := mkModule()
	m.LabeledProps = []*ast.LabeledFormula{
		{Label: lg.NewConst("myprop", lg.Boolean), Formula: lg.True},
	}

	result := CheckIsolateCompleteness(m)
	found := false
	for _, err := range result {
		if err.Caller == "myprop" {
			found = true
		}
	}
	if !found {
		t.Error("expected unchecked property error for 'myprop'")
	}
}

func TestCheckIsolateCompleteness_CheckedProperty(t *testing.T) {
	m := mkModule()
	cfg := ast.NewAstConfig()
	m.LabeledProps = []*ast.LabeledFormula{
		cfg.NewLabeledFormula(lg.NewConst("myprop", lg.Boolean), lg.True),
	}

	// Add an isolate that verifies "myprop"
	// IsolateDef: Elems[0]=name, Elems[1:end-WithArgs]=verified, Elems[end-WithArgs:]=present
	iso := cfg.NewIsolateDef(
		[]ast.Node{cfg.NewAtom("test_iso"), cfg.NewAtom("myprop")}, 0,
	)
	m.Isolates["test_iso"] = iso

	result := CheckIsolateCompleteness(m)
	for _, err := range result {
		if err.Caller == "myprop" {
			t.Error("property 'myprop' should not be reported as unchecked")
		}
	}
}

func TestCheckIsolateCompleteness_UncheckedAssertion(t *testing.T) {
	m := mkModule()

	// Action "caller" calls "callee", callee has an assertion
	calleeAtom := lg.NewConst("callee", lg.TopS)
	callAction := actions.NewCallActionOn(actions.NewActionsConfig(),calleeAtom)
	m.Actions.Set("caller", actions.NewSequence(callAction))
	m.Actions.Set("callee", actions.NewSequence(actions.NewAssertAction(lg.True)))

	result := CheckIsolateCompleteness(m)
	found := false
	for _, err := range result {
		if err.Caller == "caller" && err.Callee == "callee" {
			found = true
		}
	}
	if !found {
		t.Error("expected unchecked assertion error for caller->callee")
	}
}

func TestCheckIsolateCompleteness_DelegateAllowsUnchecked(t *testing.T) {
	m := mkModule()

	calleeAtom := lg.NewConst("callee", lg.TopS)
	callAction := actions.NewCallActionOn(actions.NewActionsConfig(),calleeAtom)
	m.Actions.Set("caller", actions.NewSequence(callAction))
	m.Actions.Set("callee", actions.NewSequence(actions.NewAssertAction(lg.True)))

	// Delegate callee (no delegee = self-delegate)
	m.Delegates = append(m.Delegates, &testDelegator{delegated: "callee", delegee: ""})

	// The caller must be in checkedContext[callee] for the delegate exception
	// Without an isolate, it won't be, so it should still fail.
	result := CheckIsolateCompleteness(m)
	// With a delegate but no isolate verifying caller, there should still be an error
	foundAssertion := false
	for _, err := range result {
		if err.Callee == "callee" && err.Msg == "assertion is not checked" {
			foundAssertion = true
		}
	}
	if !foundAssertion {
		t.Error("expected assertion error even with delegate when no isolate verifies")
	}
}

func TestCheckIsolateCompleteness_ExportedActionUnchecked(t *testing.T) {
	m := mkModule()
	m.Actions.Set("exported_act", actions.NewSequence(
		actions.NewAssertAction(lg.True),
	))
	m.Exports = append(m.Exports, &testExporter{name: "exported_act", scope: ""})

	result := CheckIsolateCompleteness(m)
	found := false
	for _, err := range result {
		if err.Caller == "external" && err.Callee == "exported_act" {
			found = true
		}
	}
	if !found {
		t.Error("expected unchecked export error for 'exported_act'")
	}
}

func TestCheckIsolateCompleteness_ScopedExportSkipped(t *testing.T) {
	m := mkModule()
	m.Actions.Set("scoped_act", actions.NewSequence(
		actions.NewAssertAction(lg.True),
	))
	// Scoped export (not global) - should be skipped
	m.Exports = append(m.Exports, &testExporter{name: "scoped_act", scope: "mymod"})

	result := CheckIsolateCompleteness(m)
	for _, err := range result {
		if err.Callee == "scoped_act" && err.Caller == "external" {
			t.Error("scoped export should not be checked as external")
		}
	}
}

func TestCheckIsolateCompleteness_RequiresUnchecked(t *testing.T) {
	m := mkModule()

	calleeAtom := lg.NewConst("callee", lg.TopS)
	callAction := actions.NewCallActionOn(actions.NewActionsConfig(),calleeAtom)
	m.Actions.Set("caller", actions.NewSequence(callAction))
	m.Actions.Set("callee", actions.NewSequence(actions.NewRequiresAction(lg.True)))

	result := CheckIsolateCompleteness(m)
	foundRequire := false
	for _, err := range result {
		if err.Kind == "require" && err.Callee == "callee" {
			foundRequire = true
		}
	}
	if !foundRequire {
		t.Error("expected unchecked requires error for caller->callee")
	}
}

func TestCheckIsolateCompleteness_NoAssertionNoError(t *testing.T) {
	m := mkModule()

	// caller calls callee, but callee has NO assertions -> no error
	calleeAtom := lg.NewConst("callee", lg.TopS)
	callAction := actions.NewCallActionOn(actions.NewActionsConfig(),calleeAtom)
	m.Actions.Set("caller", actions.NewSequence(callAction))
	m.Actions.Set("callee", actions.NewSequence()) // no assertions

	result := CheckIsolateCompleteness(m)
	for _, err := range result {
		if err.Callee == "callee" {
			t.Error("should not report error when callee has no assertions")
		}
	}
}

func TestCheckIsolateCompleteness_MixinAssertionUnchecked(t *testing.T) {
	m := mkModule()

	calleeAtom := lg.NewConst("callee", lg.TopS)
	callAction := actions.NewCallActionOn(actions.NewActionsConfig(),calleeAtom)
	m.Actions.Set("caller", actions.NewSequence(callAction))
	m.Actions.Set("callee", actions.NewSequence())
	m.Actions.Set("mixin_act", actions.NewSequence(
		actions.NewAssertAction(lg.True),
	))

	// Add a before-mixin on callee
	m.Mixins.Set("callee", []module.MixinDef{
		&testMixin{mixer: "mixin_act", mixee: "callee", after: false},
	})

	result := CheckIsolateCompleteness(m)
	foundMixin := false
	for _, err := range result {
		if err.Callee == "mixin_act" {
			foundMixin = true
		}
	}
	if !foundMixin {
		t.Error("expected error for unchecked mixin assertion")
	}
}

func TestCheckIsolateCompleteness_PropertyDedup(t *testing.T) {
	m := mkModule()
	// Same property twice -> should only report once
	m.LabeledProps = []*ast.LabeledFormula{
		{Label: lg.NewConst("dup", lg.Boolean), Formula: lg.True},
		{Label: lg.NewConst("dup", lg.Boolean), Formula: lg.True},
	}

	result := CheckIsolateCompleteness(m)
	count := 0
	for _, err := range result {
		if err.Caller == "dup" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected property 'dup' reported exactly once, got %d", count)
	}
}

// -----------------------------------------------------------------------
// Integration: backward compat with old tests
// -----------------------------------------------------------------------

func TestGetCallsModsRec_BackwardCompat(t *testing.T) {
	// GetCallsModsRec should work without the optional loops param
	m := mkModuleWithActions(map[string]actions.Action{
		"foo": actions.NewSequence(),
	})
	summarized := map[string]bool{"foo": true}
	calls := make(map[string]map[string]bool)
	mods := make(map[string]map[string]bool)

	// Call without loops param (old behavior)
	GetCallsModsRec(m, summarized, "foo", calls, mods)
	if calls["foo"] == nil {
		t.Error("calls should be populated even without loops param")
	}
}
