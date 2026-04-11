package check

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/temporal"
)

// rankingTestAstCfg is the test fixture for ranking tests. Renamed from
// testAstCfg during the l2s/+ranking/ → check/ merge to avoid colliding
// with check/check_port_test.go's testAstCfg.
var rankingTestAstCfg = ast.NewAstConfig()

// --- helpers ---

func boolVar(name string) *lg.Variable {
	v, _ := lg.NewVariable(name, lg.Boolean)
	return v
}

func boolConst(name string) *lg.Const {
	return lg.NewConst(name, lg.Boolean)
}

func unintSort(name string) lg.Sort {
	return &lg.UninterpretedSort{Name: name}
}

// --- ForAll / Exists ---

func TestForAllEmpty(t *testing.T) {
	body := lg.True
	result := ForAll(nil, body)
	if result != body {
		t.Error("ForAll with nil vars should return body unchanged")
	}
}

func TestForAllWithVars(t *testing.T) {
	v := boolVar("X")
	body := lg.True
	result := ForAll([]*lg.Variable{v}, body)
	fa, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("expected ForAll, got %T", result)
	}
	if len(fa.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(fa.Variables))
	}
}

func TestExistsEmpty(t *testing.T) {
	body := lg.True
	result := Exists(nil, body)
	if result != body {
		t.Error("Exists with nil vars should return body unchanged")
	}
}

func TestExistsWithVars(t *testing.T) {
	v := boolVar("X")
	body := lg.True
	result := Exists([]*lg.Variable{v}, body)
	ex, ok := result.(*lg.Exists)
	if !ok {
		t.Fatalf("expected Exists, got %T", result)
	}
	if len(ex.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(ex.Variables))
	}
}

// --- OldOf ---

func TestOldOfConst(t *testing.T) {
	c := boolConst("foo")
	result := OldOf(c)
	rc, ok := result.(*lg.Const)
	if !ok {
		t.Fatalf("expected Const, got %T", result)
	}
	if rc.Name != "old_foo" {
		t.Errorf("expected 'old_foo', got %q", rc.Name)
	}
}

func TestOldOfNot(t *testing.T) {
	c := boolConst("bar")
	n := &lg.Not{Body: c}
	result := OldOf(n)
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

func TestOldOfAnd(t *testing.T) {
	a := boolConst("A")
	b := boolConst("B")
	and := &lg.And{Terms: []lg.Expr{a, b}}
	result := OldOf(and)
	ra, ok := result.(*lg.And)
	if !ok {
		t.Fatalf("expected And, got %T", result)
	}
	if len(ra.Terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(ra.Terms))
	}
}

func TestOldOfOr(t *testing.T) {
	a := boolConst("A")
	b := boolConst("B")
	or := &lg.Or{Terms: []lg.Expr{a, b}}
	result := OldOf(or)
	_, ok := result.(*lg.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", result)
	}
}

func TestOldOfImplies(t *testing.T) {
	a := boolConst("A")
	b := boolConst("B")
	imp := &lg.Implies{T1: a, T2: b}
	result := OldOf(imp)
	ri, ok := result.(*lg.Implies)
	if !ok {
		t.Fatalf("expected Implies, got %T", result)
	}
	if ri.T1.(*lg.Const).Name != "old_A" {
		t.Error("T1 not renamed")
	}
}

func TestOldOfForAll(t *testing.T) {
	v := boolVar("X")
	body := boolConst("P")
	fa := &lg.ForAll{Variables: []*lg.Variable{v}, Body: body}
	result := OldOf(fa)
	rfa, ok := result.(*lg.ForAll)
	if !ok {
		t.Fatalf("expected ForAll, got %T", result)
	}
	if rfa.Body.(*lg.Const).Name != "old_P" {
		t.Error("body not renamed")
	}
}

func TestOldOfNil(t *testing.T) {
	result := OldOf(nil)
	if result != nil {
		t.Error("OldOf(nil) should return nil")
	}
}

func TestOldOfEq(t *testing.T) {
	a := boolConst("A")
	b := boolConst("B")
	eq := &lg.Eq{T1: a, T2: b}
	result := OldOf(eq)
	req, ok := result.(*lg.Eq)
	if !ok {
		t.Fatalf("expected Eq, got %T", result)
	}
	if req.T1.(*lg.Const).Name != "old_A" {
		t.Error("T1 not renamed")
	}
}

func TestOldOfIff(t *testing.T) {
	a := boolConst("A")
	b := boolConst("B")
	iff := &lg.Iff{T1: a, T2: b}
	result := OldOf(iff)
	_, ok := result.(*lg.Iff)
	if !ok {
		t.Fatalf("expected Iff, got %T", result)
	}
}

func TestOldOfExists(t *testing.T) {
	v := boolVar("X")
	body := boolConst("P")
	ex := &lg.Exists{Variables: []*lg.Variable{v}, Body: body}
	result := OldOf(ex)
	rex, ok := result.(*lg.Exists)
	if !ok {
		t.Fatalf("expected Exists, got %T", result)
	}
	if rex.Body.(*lg.Const).Name != "old_P" {
		t.Error("body not renamed")
	}
}

// --- L2S named binder constructors ---

func TestL2sD(t *testing.T) {
	s := unintSort("node")
	d := L2sD(s)
	if d == nil {
		t.Fatal("L2sD returned nil")
	}
	if d.Name != "l2s_d" {
		t.Errorf("expected name 'l2s_d', got %q", d.Name)
	}
}

func TestL2sW(t *testing.T) {
	nb := L2sW(nil, lg.True, "env")
	if nb == nil {
		t.Fatal("L2sW returned nil")
	}
	if nb.Name != "l2s_w" {
		t.Errorf("expected name 'l2s_w', got %q", nb.Name)
	}
}

func TestL2sG(t *testing.T) {
	v := boolVar("X")
	nb := L2sG([]*lg.Variable{v}, lg.True, "env")
	if nb == nil {
		t.Fatal("L2sG returned nil")
	}
	if nb.Name != "l2s_g" {
		t.Errorf("expected name 'l2s_g', got %q", nb.Name)
	}
	if len(nb.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(nb.Variables))
	}
}

func TestOldL2sG(t *testing.T) {
	nb := OldL2sG(nil, lg.True, "env")
	if nb == nil {
		t.Fatal("OldL2sG returned nil")
	}
	if nb.Name != "_old_l2s_g" {
		t.Errorf("expected name '_old_l2s_g', got %q", nb.Name)
	}
}

func TestL2sInit(t *testing.T) {
	nb := L2sInit(nil, lg.True, "label")
	if nb == nil {
		t.Fatal("L2sInit returned nil")
	}
	if nb.Name != "l2s_init" {
		t.Errorf("expected name 'l2s_init', got %q", nb.Name)
	}
}

func TestL2sWhen(t *testing.T) {
	nb := L2sWhen("next", nil, lg.True, "label")
	if nb == nil {
		t.Fatal("L2sWhen returned nil")
	}
	if nb.Name != "l2s_whennext" {
		t.Errorf("expected name 'l2s_whennext', got %q", nb.Name)
	}
}

func TestL2sOld(t *testing.T) {
	nb := L2sOld(nil, lg.True, "label")
	if nb == nil {
		t.Fatal("L2sOld returned nil")
	}
	if nb.Name != "l2s_old" {
		t.Errorf("expected name 'l2s_old', got %q", nb.Name)
	}
}

func TestL2sS(t *testing.T) {
	nb := L2sS(nil, lg.True, "label")
	if nb == nil {
		t.Fatal("L2sS returned nil")
	}
	if nb.Name != "l2s_s" {
		t.Errorf("expected name 'l2s_s', got %q", nb.Name)
	}
}

// --- TrigGlob ---

func TestTrigGlobGlobally(t *testing.T) {
	body := boolConst("P")
	g, _ := lg.NewGlobally(nil, body)
	result := TrigGlob(g, true)
	if len(result) == 0 {
		t.Error("expected at least one result for Globally in positive polarity")
	}
}

func TestTrigGlobEventually(t *testing.T) {
	body := boolConst("P")
	e, _ := lg.NewEventually(nil, body)
	result := TrigGlob(e, false)
	if len(result) == 0 {
		t.Error("expected at least one result for Eventually in negative polarity")
	}
}

func TestTrigGlobImplies(t *testing.T) {
	// ~(Q -> G(P)) is equivalent to Q & ~G(P).
	// In negative polarity, Implies recurses into T1 with !pos=true and T2 with pos=false.
	// T2 = Globally with pos=false -> no match.
	// So we test with a structure where the match works:
	// Implies(G(P), Q) in negative polarity -> T1 = G(P) with !pos = true -> match!
	body := boolConst("P")
	g, _ := lg.NewGlobally(nil, body)
	imp := &lg.Implies{T1: g, T2: boolConst("Q")}
	result := TrigGlob(imp, false)
	if len(result) == 0 {
		t.Error("expected results from Implies negative polarity with Globally in antecedent")
	}
}

func TestTrigGlobAnd(t *testing.T) {
	body := boolConst("P")
	g, _ := lg.NewGlobally(nil, body)
	and := &lg.And{Terms: []lg.Expr{g, boolConst("Q")}}
	result := TrigGlob(and, true)
	if len(result) == 0 {
		t.Error("expected results from And positive polarity")
	}
}

func TestTrigGlobNil(t *testing.T) {
	result := TrigGlob(nil, true)
	if len(result) != 0 {
		t.Error("expected empty result for nil")
	}
}

func TestTrigGlobNot(t *testing.T) {
	body := boolConst("P")
	g, _ := lg.NewGlobally(nil, body)
	n := &lg.Not{Body: g}
	// Not flips polarity: so Globally in negative polarity -> no match
	result := TrigGlob(n, true)
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

// --- DiagnoseFailure ---

func TestDiagnoseFailureCreated(t *testing.T) {
	info := DiagnoseFailure("l2s_created_foo", nil, nil)
	if !strings.Contains(info.Message, "work_created") {
		t.Error("expected message about work_created")
	}
	if info.Suffix != "_foo" {
		t.Errorf("expected suffix '_foo', got %q", info.Suffix)
	}
}

func TestDiagnoseFailureEventuallyStart(t *testing.T) {
	info := DiagnoseFailure("l2s_eventually_start", nil, nil)
	if !strings.Contains(info.Message, "eventually") {
		t.Error("expected message about eventually")
	}
}

func TestDiagnoseFailureNeededPreserved(t *testing.T) {
	info := DiagnoseFailure("l2s_needed_preserved_bar", nil, nil)
	if !strings.Contains(info.Message, "work_needed") {
		t.Error("expected message about work_needed")
	}
}

func TestDiagnoseFailureProgress(t *testing.T) {
	info := DiagnoseFailure("l2s_progress_baz", nil, nil)
	if !strings.Contains(info.Message, "work_needed") {
		t.Error("expected message about work_needed decreases")
	}
}

func TestDiagnoseFailureSchedStable(t *testing.T) {
	info := DiagnoseFailure("l2s_sched_stable_x", nil, nil)
	if !strings.Contains(info.Message, "work_helpful") {
		t.Error("expected message about work_helpful stable")
	}
}

func TestDiagnoseFailureSchedExists(t *testing.T) {
	tasks := map[string]*Task{
		"_a": {WorkHelpful: lg.True},
	}
	info := DiagnoseFailure("l2s_sched_exists", tasks, nil)
	if !strings.Contains(info.Message, "helpful") {
		t.Error("expected message about helpful sets")
	}
}

func TestDiagnoseFailureUnknown(t *testing.T) {
	info := DiagnoseFailure("unknown_name", nil, nil)
	if !strings.Contains(info.Message, "Unknown") {
		t.Error("expected Unknown message")
	}
}

// --- IsTemporalAndL2S ---

func TestIsTemporalAndL2S(t *testing.T) {
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

func TestRenameMap(t *testing.T) {
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

func TestDependenciesEmpty(t *testing.T) {
	result := Dependencies(nil, nil)
	if len(result) != 0 {
		t.Error("expected empty result")
	}
}

func TestDependenciesLinear(t *testing.T) {
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

func TestDependenciesCyclic(t *testing.T) {
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

func TestConvertToInitConst(t *testing.T) {
	c := boolConst("P")
	result := ConvertToInit(c, "label")
	nb, ok := result.(*lg.NamedBinder)
	if !ok {
		t.Fatalf("expected NamedBinder, got %T", result)
	}
	if nb.Name != "l2s_init" {
		t.Errorf("expected 'l2s_init', got %q", nb.Name)
	}
}

func TestConvertToInitAnd(t *testing.T) {
	a := boolConst("A")
	b := boolConst("B")
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

func TestConvertToInitNil(t *testing.T) {
	result := ConvertToInit(nil, "label")
	if result != nil {
		t.Error("ConvertToInit(nil) should return nil")
	}
}

// --- Desugar ---

func TestDesugarNonBinder(t *testing.T) {
	c := boolConst("P")
	result := RankingDesugar(c, "label", lg.True)
	// Non-binder should be returned unchanged
	if result != c {
		t.Error("non-binder should pass through unchanged")
	}
}

func TestDesugarNil(t *testing.T) {
	result := RankingDesugar(nil, "label", lg.True)
	if result != nil {
		t.Error("RankingDesugar(nil) should return nil")
	}
}

// --- L2STactic ---

func TestL2STacticNilConfig(t *testing.T) {
	_, err := RankingL2STactic(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestL2STacticNoGoals(t *testing.T) {
	cfg := &L2STacticConfig{Goals: nil}
	_, err := RankingL2STactic(cfg)
	if err == nil {
		t.Error("expected error for no goals")
	}
}

func TestL2STacticWithLets(t *testing.T) {
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

func TestModelPassNil(t *testing.T) {
	// Should not panic
	ModelPass(nil, func(n lg.Expr) lg.Expr { return n })
}

func TestModelPassTransform(t *testing.T) {
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
	ModelPass(model, func(n lg.Expr) lg.Expr {
		called++
		return n
	})
	if called != 2 {
		t.Errorf("expected transform called 2 times, got %d", called)
	}
}

// --- RegisterTactic / GetTactic ---

func TestRegisterAndGetTactic(t *testing.T) {
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

func TestGetTacticNotFound(t *testing.T) {
	_, ok := GetTactic("nonexistent")
	if ok {
		t.Error("nonexistent tactic should not be found")
	}
}

func TestGetRankingTactic(t *testing.T) {
	fn, ok := GetTactic("ranking")
	if !ok {
		t.Fatal("ranking tactic should be registered by default")
	}
	if fn == nil {
		t.Error("ranking tactic function should not be nil")
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
		result := OldOf(c)
		if result == nil {
			t.Error("OldOf should not return nil")
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
