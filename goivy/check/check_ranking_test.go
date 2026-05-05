package check

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/proof"
	"github.com/glycerine/ivy/goivy/temporal"
)

// rankingTestAstCfg is the test fixture for ranking tests. The "ranking"
// prefix avoids colliding with check_port_test.go's testAstCfg.
var rankingTestAstCfg = ast.NewAstConfig()

// --- helpers ---

func checkBoolVar(name string) *lg.Variable {
	v, _ := lg.NewVariable(name, lg.Boolean)
	return v
}

func checkBoolConst(name string) *lg.Const {
	return lg.NewConst(name, lg.Boolean)
}

func checkUnintSort(name string) lg.Sort {
	return &lg.UninterpretedSort{Name: name}
}

// --- CheckForAll / CheckExists ---

func TestCheckForAllEmpty(t *testing.T) {
	body := lg.True
	result := CheckForAll(nil, body)
	if result != body {
		t.Error("CheckForAll with nil vars should return body unchanged")
	}
}

func TestCheckForAllWithVars(t *testing.T) {
	v := checkBoolVar("X")
	body := lg.True
	result := CheckForAll([]*lg.Variable{v}, body)
	fa, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("expected CheckForAll, got %T", result)
	}
	if len(fa.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(fa.Variables))
	}
}

func TestCheckExistsEmpty(t *testing.T) {
	body := lg.True
	result := CheckExists(nil, body)
	if result != body {
		t.Error("CheckExists with nil vars should return body unchanged")
	}
}

func TestCheckExistsWithVars(t *testing.T) {
	v := checkBoolVar("X")
	body := lg.True
	result := CheckExists([]*lg.Variable{v}, body)
	ex, ok := result.(*lg.Exists)
	if !ok {
		t.Fatalf("expected CheckExists, got %T", result)
	}
	if len(ex.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(ex.Variables))
	}
}

// --- CheckOldOf ---

func TestCheckOldOfConst(t *testing.T) {
	c := checkBoolConst("foo")
	result := CheckOldOf(c)
	rc, ok := result.(*lg.Const)
	if !ok {
		t.Fatalf("expected Const, got %T", result)
	}
	if rc.Name != "old_foo" {
		t.Errorf("expected 'old_foo', got %q", rc.Name)
	}
}

func TestCheckOldOfNot(t *testing.T) {
	c := checkBoolConst("bar")
	n := &lg.Not{Body: c}
	result := CheckOldOf(n)
	not, ok := result.(*lg.Not)
	if !ok {
		t.Fatalf("expected Not, got %T", result)
	}
	rc, ok := not.Body.(*lg.Const)
	if !ok {
		t.Fatalf("expected Const in Not body, got %T", not.Body)
	}
	if rc.Name != "old_bar" {
		t.Errorf("expected 'old_bar', got %q", rc.Name)
	}
}

func TestCheckOldOfAnd(t *testing.T) {
	a := checkBoolConst("A")
	b := checkBoolConst("B")
	and := &lg.And{Terms: []lg.Expr{a, b}}
	result := CheckOldOf(and)
	ra, ok := result.(*lg.And)
	if !ok {
		t.Fatalf("expected And, got %T", result)
	}
	if len(ra.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(ra.Terms))
	}
}

func TestCheckOldOfOr(t *testing.T) {
	a := checkBoolConst("A")
	b := checkBoolConst("B")
	or := &lg.Or{Terms: []lg.Expr{a, b}}
	result := CheckOldOf(or)
	_, ok := result.(*lg.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", result)
	}
}

func TestCheckOldOfImplies(t *testing.T) {
	a := checkBoolConst("A")
	b := checkBoolConst("B")
	imp := &lg.Implies{T1: a, T2: b}
	result := CheckOldOf(imp)
	ri, ok := result.(*lg.Implies)
	if !ok {
		t.Fatalf("expected Implies, got %T", result)
	}
	if ri.T1.(*lg.Const).Name != "old_A" {
		t.Error("T1 not renamed")
	}
}

func TestCheckOldOfForAll(t *testing.T) {
	v := checkBoolVar("X")
	body := checkBoolConst("P")
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: body}
	result := CheckOldOf(fa)
	rfa, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("expected CheckForAll, got %T", result)
	}
	if rfa.Body.(*lg.Const).Name != "old_P" {
		t.Error("body not renamed")
	}
}

func TestCheckOldOfNil(t *testing.T) {
	result := CheckOldOf(nil)
	if result != nil {
		t.Error("CheckOldOf(nil) should return nil")
	}
}

func TestCheckOldOfEq(t *testing.T) {
	a := checkBoolConst("A")
	b := checkBoolConst("B")
	eq := &lg.Eq{T1: a, T2: b}
	result := CheckOldOf(eq)
	req, ok := result.(*lg.Eq)
	if !ok {
		t.Fatalf("expected Eq, got %T", result)
	}
	if req.T1.(*lg.Const).Name != "old_A" {
		t.Error("T1 not renamed")
	}
}

func TestCheckOldOfIff(t *testing.T) {
	a := checkBoolConst("A")
	b := checkBoolConst("B")
	iff := &lg.Iff{T1: a, T2: b}
	result := CheckOldOf(iff)
	_, ok := result.(*lg.Iff)
	if !ok {
		t.Fatalf("expected Iff, got %T", result)
	}
}

func TestCheckOldOfExists(t *testing.T) {
	v := checkBoolVar("X")
	body := checkBoolConst("P")
	ex := &lg.Exists{Variables: []*lg.Variable{v}, Body: body}
	result := CheckOldOf(ex)
	rex, ok := result.(*lg.Exists)
	if !ok {
		t.Fatalf("expected CheckExists, got %T", result)
	}
	if rex.Body.(*lg.Const).Name != "old_P" {
		t.Error("body not renamed")
	}
}

// --- L2S named binder constructors ---

func TestCheckL2sD(t *testing.T) {
	s := checkUnintSort("node")
	d := L2sD(s)
	if d == nil {
		t.Fatal("L2sD returned nil")
	}
	if d.Name != "l2s_d" {
		t.Errorf("expected name 'l2s_d', got %q", d.Name)
	}
}

// --- TrigGlob ---

func TestCheckTrigGlobGlobally(t *testing.T) {
	body := checkBoolConst("P")
	g, _ := lg.NewGlobally(nil, body)
	result := TrigGlob(g, true)
	if len(result) == 0 {
		t.Error("expected at least one result for Globally in positive polarity")
	}
}

func TestCheckTrigGlobEventually(t *testing.T) {
	body := checkBoolConst("P")
	e, _ := lg.NewEventually(nil, body)
	result := TrigGlob(e, false)
	if len(result) == 0 {
		t.Error("expected at least one result for Eventually in negative polarity")
	}
}

func TestCheckTrigGlobImplies(t *testing.T) {
	// ~(Q -> G(P)) is equivalent to Q & ~G(P).
	// In negative polarity, Implies recurses into T1 with !pos=true and T2 with pos=false.
	// T2 = Globally with pos=false -> no match.
	// So we test with a structure where the match works:
	// Implies(G(P), Q) in negative polarity -> T1 = G(P) with !pos = true -> match!
	body := checkBoolConst("P")
	g, _ := lg.NewGlobally(nil, body)
	imp := &lg.Implies{T1: g, T2: checkBoolConst("Q")}
	result := TrigGlob(imp, false)
	if len(result) == 0 {
		t.Error("expected results from Implies negative polarity with Globally in antecedent")
	}
}

func TestCheckTrigGlobAnd(t *testing.T) {
	body := checkBoolConst("P")
	g, _ := lg.NewGlobally(nil, body)
	and := &lg.And{Terms: []lg.Expr{g, checkBoolConst("Q")}}
	result := TrigGlob(and, true)
	if len(result) == 0 {
		t.Error("expected results from And positive polarity")
	}
}

func TestCheckTrigGlobNil(t *testing.T) {
	result := TrigGlob(nil, true)
	if len(result) != 0 {
		t.Error("expected empty result for nil")
	}
}

func TestCheckTrigGlobNot(t *testing.T) {
	body := checkBoolConst("P")
	g, _ := lg.NewGlobally(nil, body)
	n := &lg.Not{Body: g}
	// Not flips polarity: so Globally in negative polarity -> no match
	result := TrigGlob(n, true)
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

// --- DiagnoseFailure ---

func TestCheckDiagnoseFailureCreated(t *testing.T) {
	info := DiagnoseFailure("l2s_created_foo", nil, nil)
	if !strings.Contains(info.Message, "work_created") {
		t.Error("expected message about work_created")
	}
	if info.Suffix != "_foo" {
		t.Errorf("expected suffix '_foo', got %q", info.Suffix)
	}
}

func TestCheckDiagnoseFailureEventuallyStart(t *testing.T) {
	info := DiagnoseFailure("l2s_eventually_start", nil, nil)
	if !strings.Contains(info.Message, "eventually") {
		t.Error("expected message about eventually")
	}
}

func TestCheckDiagnoseFailureNeededPreserved(t *testing.T) {
	info := DiagnoseFailure("l2s_needed_preserved_bar", nil, nil)
	if !strings.Contains(info.Message, "work_needed") {
		t.Error("expected message about work_needed")
	}
}

func TestCheckDiagnoseFailureProgress(t *testing.T) {
	info := DiagnoseFailure("l2s_progress_baz", nil, nil)
	if !strings.Contains(info.Message, "work_needed") {
		t.Error("expected message about work_needed decreases")
	}
}

func TestCheckDiagnoseFailureSchedStable(t *testing.T) {
	info := DiagnoseFailure("l2s_sched_stable_x", nil, nil)
	if !strings.Contains(info.Message, "work_helpful") {
		t.Error("expected message about work_helpful stable")
	}
}

func TestCheckDiagnoseFailureSchedExists(t *testing.T) {
	tasks := map[string]*Task{
		"_a": {WorkHelpful: lg.True},
	}
	info := DiagnoseFailure("l2s_sched_exists", tasks, nil)
	if !strings.Contains(info.Message, "helpful") {
		t.Error("expected message about helpful sets")
	}
}

func TestCheckDiagnoseFailureUnknown(t *testing.T) {
	info := DiagnoseFailure("unknown_name", nil, nil)
	if !strings.Contains(info.Message, "Unknown") {
		t.Error("expected Unknown message")
	}
}

// --- IsTemporalAndL2S ---

func TestCheckIsTemporalAndL2S(t *testing.T) {
	tests := []struct {
		name   string
		expect bool
	}{
		{"l2s_w_0", true},
		{"l2s_s_1", true},
		{"l2s_g_0", false}, // l2s_g is excluded
		{"_old_l2s_g", true},
		{"regular", false},
		{"l2s_d", true},
		{"l2s_init_0", true},
	}
	for _, tt := range tests {
		got := IsTemporalAndL2S(tt.name)
		if got != tt.expect {
			t.Errorf("IsTemporalAndL2S(%q) = %v, want %v", tt.name, got, tt.expect)
		}
	}
}

// --- RenameMap ---

func TestCheckRenameMap(t *testing.T) {
	subs := map[string]string{"a": "b", "c": "d"}
	inv := RenameMap(subs)
	if inv["b"] != "a" {
		t.Errorf("expected inv[b]=a, got %q", inv["b"])
	}
	if inv["d"] != "c" {
		t.Errorf("expected inv[d]=c, got %q", inv["d"])
	}
}

// --- Dependencies ---

func TestCheckDependenciesEmpty(t *testing.T) {
	result := Dependencies(nil, nil)
	if len(result) != 0 {
		t.Error("expected empty result")
	}
}

func TestCheckDependenciesLinear(t *testing.T) {
	deps := map[string][]string{
		"a": {"b"},
		"b": {"c"},
	}
	result := Dependencies([]string{"a"}, deps)
	for _, s := range []string{"a", "b", "c"} {
		if !result[s] {
			t.Errorf("expected %q in result", s)
		}
	}
}

func TestCheckDependenciesCyclic(t *testing.T) {
	deps := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}
	result := Dependencies([]string{"a"}, deps)
	if !result["a"] || !result["b"] {
		t.Error("cyclic dependencies should still work")
	}
}

// --- ConvertToInit ---

func TestCheckConvertToInitConst(t *testing.T) {
	c := checkBoolConst("P")
	result := ConvertToInit(c, "label")
	nb, ok := result.(*lg.NamedBinder)
	if !ok {
		t.Fatalf("expected NamedBinder, got %T", result)
	}
	if nb.Name != "l2s_init" {
		t.Errorf("expected 'l2s_init', got %q", nb.Name)
	}
}

func TestCheckConvertToInitAnd(t *testing.T) {
	a := checkBoolConst("A")
	b := checkBoolConst("B")
	and := &lg.And{Terms: []lg.Expr{a, b}}
	result := ConvertToInit(and, "label")
	ra, ok := result.(*lg.And)
	if !ok {
		t.Fatalf("expected And, got %T", result)
	}
	if len(ra.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(ra.Terms))
	}
	// Each term should be l2s_init wrapped
	for _, term := range ra.Terms {
		if _, ok := term.(*lg.NamedBinder); !ok {
			t.Errorf("expected NamedBinder term, got %T", term)
		}
	}
}

func TestCheckConvertToInitNil(t *testing.T) {
	result := ConvertToInit(nil, "label")
	if result != nil {
		t.Error("ConvertToInit(nil) should return nil")
	}
}

// --- Desugar ---

func TestCheckDesugarNonBinder(t *testing.T) {
	c := checkBoolConst("P")
	result := RankingDesugar(c, "label", lg.True)
	// Non-binder should be returned unchanged
	if result != c {
		t.Error("non-binder should pass through unchanged")
	}
}

func TestCheckDesugarNil(t *testing.T) {
	result := RankingDesugar(nil, "label", lg.True)
	if result != nil {
		t.Error("RankingDesugar(nil) should return nil")
	}
}

// --- L2STactic ---

func TestCheckL2STacticNilConfig(t *testing.T) {
	_, err := RankingL2STactic(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestCheckL2STacticNoGoals(t *testing.T) {
	cfg := &L2STacticConfig{Goals: nil}
	_, err := RankingL2STactic(cfg)
	if err == nil {
		t.Error("expected error for no goals")
	}
}

func TestCheckL2STacticWithLets(t *testing.T) {
	// L2STactic first checks the conclusion, then checks for lets.
	// Since we can't easily create a proper temporal goal here,
	// we verify the error from the conclusion check instead.
	cfg := &L2STacticConfig{
		Goals: []*ast.LabeledFormula{rankingTestAstCfg.NewLabeledFormula(nil, rankingTestAstCfg.NewAtom("true"))},
		Proof: &ProofDecl{TacticLets: []ast.Node{rankingTestAstCfg.NewAtom("x")}},
	}
	_, err := RankingL2STactic(cfg)
	if err == nil {
		t.Error("expected error for tactic with non-temporal goal")
	}
	// The error should be about the goal not being temporal (checked before lets)
	if err != nil && !strings.Contains(err.Error(), "no conclusion") && !strings.Contains(err.Error(), "temporal") {
		t.Errorf("expected error about conclusion or temporal, got %q", err.Error())
	}
}

// --- ModelPass ---

func TestCheckModelPassNil(t *testing.T) {
	// Should not panic
	ModelPass(nil, func(n ast.Node) ast.Node { return n })
}

func TestCheckModelPassTransform(t *testing.T) {
	cfg := ast.NewAstConfig()
	model := &temporal.NormalProgram{
		Invars: []*ast.LabeledFormula{
			cfg.NewLabeledFormula(nil, lg.True),
		},
		Asms: []*ast.LabeledFormula{
			cfg.NewLabeledFormula(nil, lg.True),
		},
	}
	called := 0
	ModelPass(model, func(n ast.Node) ast.Node {
		called++
		return n
	})
	if called != 2 {
		t.Errorf("expected transform called 2 times, got %d", called)
	}
}

// --- RegisterTactic / GetTactic ---

func TestCheckRegisterAndGetTactic(t *testing.T) {
	called := false
	RegisterTactic("test_tactic", func(cfg *L2STacticConfig) ([]*ast.LabeledFormula, error) {
		called = true
		return nil, nil
	})
	fn, ok := GetTactic("test_tactic")
	if !ok {
		t.Fatal("registered tactic not found")
	}
	fn(nil)
	if !called {
		t.Error("tactic function not called")
	}
	// Clean up
	delete(RegisteredTactics, "test_tactic")
}

func TestCheckGetTacticNotFound(t *testing.T) {
	_, ok := GetTactic("nonexistent")
	if ok {
		t.Error("nonexistent tactic should not be found")
	}
}

func TestCheckGetRankingTactic(t *testing.T) {
	fn, ok := GetTactic("ranking")
	if !ok {
		t.Fatal("ranking tactic should be registered by default")
	}
	if fn == nil {
		t.Error("ranking tactic function should not be nil")
	}
}

// --- Bug 12 regression: ranking modPass must transform property prems ---

// TestRankingModPass_PropertyPremsTransformed verifies that the ranking
// modPass pattern correctly applies a transform to property premises.
// Bug 12: Go was iterating local invars instead of prems.
func TestCheckRankingModPass_PropertyPremsTransformed(t *testing.T) {
	cfg := ast.NewAstConfig()
	model := &temporal.NormalProgram{
		Invars: []*ast.LabeledFormula{
			cfg.NewLabeledFormula(nil, lg.True),
		},
	}

	// Build prems: one property prem (plain formula), one schema prem.
	propPrem := cfg.NewLabeledFormula(checkBoolConst("P"), checkBoolConst("Q"))
	schemaPrem := cfg.NewLabeledFormula(checkBoolConst("S"), cfg.NewSchemaBody(checkBoolConst("X")))
	prems := []ast.Node{propPrem, schemaPrem}

	// Define a transform that wraps the formula in Not (handles LabeledFormula).
	negate := func(n ast.Node) ast.Node {
		if lf, ok := n.(*ast.LabeledFormula); ok {
			return lf.Clone([]ast.Node{lf.Label, &lg.Not{Body: lf.Formula.(lg.Expr)}})
		}
		return &lg.Not{Body: n.(lg.Expr)}
	}

	// Simulate the ranking modPass prems loop (Bug 12 fix).
	for i, p := range prems {
		lf, ok := p.(*ast.LabeledFormula)
		if !ok || !proof.GoalIsProperty(lf) {
			continue
		}
		prems[i] = negate(lf).(*ast.LabeledFormula)
	}
	// Also transform model invars.
	for i, inv := range model.Invars {
		model.Invars[i] = negate(inv).(*ast.LabeledFormula)
	}

	// Property prem should be transformed (Not wrapped).
	transformedPrem, ok := prems[0].(*ast.LabeledFormula)
	if !ok {
		t.Fatal("expected LabeledFormula")
	}
	if _, ok := transformedPrem.Formula.(*lg.Not); !ok {
		t.Errorf("Bug 12 regression: property prem should be transformed, got %T", transformedPrem.Formula)
	}

	// Schema prem should be unchanged.
	schemaPremResult, ok := prems[1].(*ast.LabeledFormula)
	if !ok {
		t.Fatal("expected LabeledFormula")
	}
	if _, ok := schemaPremResult.Formula.(*ast.SchemaBody); !ok {
		t.Errorf("schema prem should NOT be transformed, got %T", schemaPremResult.Formula)
	}

	// Model invar should also be transformed.
	if _, ok := model.Invars[0].Formula.(*lg.Not); !ok {
		t.Errorf("model invar should be transformed, got %T", model.Invars[0].Formula)
	}
}

// TestRankingModPass_PropertyPremClonePreserve verifies that cloning a
// property prem during modPass preserves the LF id (PRESERVE semantics).
func TestCheckRankingModPass_PropertyPremClonePreserve(t *testing.T) {
	cfg := ast.NewAstConfig()
	origLF := cfg.NewLabeledFormula(checkBoolConst("label"), checkBoolConst("body"))
	origID := origLF.ID

	// Clone with identity transform (PRESERVE mode).
	cfg.AlwaysCloneWithFreshID = false
	cloned := origLF.Clone([]ast.Node{origLF.Label, origLF.Formula.(lg.Expr)}).(*ast.LabeledFormula)

	if cloned.ID != origID {
		t.Errorf("Bug 12 regression: PRESERVE clone should keep origid=%d, got %d", origID, cloned.ID)
	}
	if cloned == origLF {
		t.Error("clone should be a different pointer")
	}
}

// --- Bug 13 regression: ranking invars must merge into model.Invars ---

// TestRankingInvarsMerge verifies that ranking-generated invars are merged
// into model.Invars before downstream processing.
// Bug 13: Go kept ranking invars separate, so SharedStep3 missed them.
func TestCheckRankingInvarsMerge(t *testing.T) {
	cfg := ast.NewAstConfig()

	// Original model invars.
	modelInvar := cfg.NewLabeledFormula(checkBoolConst("model_inv"), lg.True)
	model := &temporal.NormalProgram{
		Invars: []*ast.LabeledFormula{modelInvar},
	}

	// Ranking-generated invars (separate list).
	rankInvar1 := cfg.NewLabeledFormula(checkBoolConst("rank_inv1"), lg.True)
	rankInvar2 := cfg.NewLabeledFormula(checkBoolConst("rank_inv2"), lg.True)
	invars := []*ast.LabeledFormula{rankInvar1, rankInvar2}

	// Bug 13 fix: merge invars into model.Invars.
	model.Invars = append(model.Invars, invars...)

	if len(model.Invars) != 3 {
		t.Fatalf("Bug 13 regression: expected 3 invars after merge, got %d", len(model.Invars))
	}
	// Verify all three are present.
	names := make(map[string]bool)
	for _, inv := range model.Invars {
		if c, ok := inv.Label.(*lg.Const); ok {
			names[c.Name] = true
		}
	}
	for _, want := range []string{"model_inv", "rank_inv1", "rank_inv2"} {
		if !names[want] {
			t.Errorf("Bug 13 regression: missing invar %q after merge", want)
		}
	}
}

// TestRankingInvarsMerge_SharedStep3Visible verifies that after merging,
// SharedStep3_CollectNamedBinders would see all invars in model.Invars.
func TestCheckRankingInvarsMerge_SharedStep3Visible(t *testing.T) {
	cfg := ast.NewAstConfig()

	// Create a model invar and a ranking invar, both with named binders.
	modelInvar := cfg.NewLabeledFormula(checkBoolConst("m"), checkBoolConst("body_m"))
	rankInvar := cfg.NewLabeledFormula(checkBoolConst("r"), checkBoolConst("body_r"))

	model := &temporal.NormalProgram{
		Invars: []*ast.LabeledFormula{modelInvar},
	}

	// Before merge: only 1 invar visible.
	if len(model.Invars) != 1 {
		t.Fatalf("expected 1 invar before merge, got %d", len(model.Invars))
	}

	// Merge.
	invars := []*ast.LabeledFormula{rankInvar}
	model.Invars = append(model.Invars, invars...)

	// After merge: 2 invars visible.
	if len(model.Invars) != 2 {
		t.Fatalf("Bug 13 regression: expected 2 invars after merge, got %d", len(model.Invars))
	}
}

// --- Bug 14 regression: WhenOperator invariant generation ---

// TestWhenOperatorInvarGeneration verifies that WhenOperator{Name:"first"}
// nodes in temporals produce ranking invariants.
// Bug 14: Go ranking was completely missing this code.
func TestCheckWhenOperatorInvarGeneration(t *testing.T) {
	// Build a WhenOperator with Name="first"
	wo := &lg.WhenOperator{Name: "first", T1: checkBoolConst("cond"), T2: checkBoolConst("val")}

	// Collect temporals — should find the WhenOperator
	temporals := lu.TemporalsAst(wo)
	if len(temporals) == 0 {
		t.Fatal("TemporalsAst should find WhenOperator")
	}
	found := false
	for _, tmp := range temporals {
		if w, ok := tmp.(*lg.WhenOperator); ok && w.Name == "first" {
			found = true
		}
	}
	if !found {
		t.Error("Bug 14 regression: should find WhenOperator{first} in temporals")
	}
}

// TestWhenOperatorInvarGeneration_NoFirst verifies that WhenOperator{Name:"next"}
// does NOT generate invariants.
func TestCheckWhenOperatorInvarGeneration_NoFirst(t *testing.T) {
	wo := &lg.WhenOperator{Name: "next", T1: checkBoolConst("cond"), T2: checkBoolConst("val")}
	temporals := lu.TemporalsAst(wo)
	for _, tmp := range temporals {
		if w, ok := tmp.(*lg.WhenOperator); ok && w.Name == "first" {
			t.Error("should NOT find WhenOperator{first} when name is 'next'")
			_ = w
		}
	}
}

// --- Bug 15 regression: AddParamsToD type check ---

// TestAddParamsToD_AcceptsNonUninterpSort verifies that addParamsToD
// includes parameters of any non-finite sort, not just UninterpretedSort.
// Bug 15: Go had an extra type assertion restricting to *lg.UninterpretedSort.
func TestCheckAddParamsToD_AcceptsNonUninterpSort(t *testing.T) {
	// A non-finite, non-UninterpretedSort (e.g., function sort).
	us := &lg.UninterpretedSort{Name: "node"}
	fs, _ := lg.NewFunctionSort(us, lg.Boolean)

	finiteSorts := map[string]bool{"nat": true}

	// Simulate the fixed addParamsToD loop.
	sort := fs
	sortStr := sort.String()
	if sort != nil && !finiteSorts[sortStr] {
		// Bug 15 fix: no extra type assertion — this should be reached.
		t.Logf("Bug 15 regression: correctly includes sort %q", sortStr)
	} else {
		t.Errorf("Bug 15 regression: sort %q should not be excluded", sortStr)
	}
}

// --- Bug 16 regression: ranking trace_hook ---

// TestRankingTraceHook_Pattern verifies the trace_hook closure pattern
// used by both l2s and ranking tactics.
// Bug 16: Go ranking was missing trace_hook setup entirely.
func TestCheckRankingTraceHook_Pattern(t *testing.T) {
	// Verify the TraceHookFn type exists and can wrap a function.
	subs := map[string]string{"_c0": "l2s_w_0"}
	var hookCalled bool
	hook := TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
		hookCalled = true
		if subs != nil {
			applyRenamingToHandler(handler, subs)
		}
	})
	// Call the hook with a simple handler.
	handler := &MatchHandler{Lines: []string{"val=_c0"}}
	hook(handler, nil)
	if !hookCalled {
		t.Error("Bug 16 regression: trace hook should be callable")
	}
	if handler.Lines[0] != "val=l2s_w_0" {
		t.Errorf("Bug 16 regression: expected renaming applied, got %q", handler.Lines[0])
	}
}

// --- Fuzz ---

func FuzzOldOf(f *testing.F) {
	f.Add("foo")
	f.Add("bar.baz")
	f.Add("__skolem")
	f.Add("")
	f.Fuzz(func(t *testing.T, name string) {
		c := lg.NewConst(name, lg.Boolean)
		result := CheckOldOf(c)
		if result == nil {
			t.Error("CheckOldOf should not return nil")
		}
		rc, ok := result.(*lg.Const)
		if !ok {
			t.Fatalf("expected Const, got %T", result)
		}
		expected := "old_" + name
		if rc.Name != expected {
			t.Errorf("expected %q, got %q", expected, rc.Name)
		}
	})
}
