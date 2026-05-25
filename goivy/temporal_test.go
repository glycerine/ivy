package goivy

import (
	"strings"
	"testing"
)

// helpers for tests
func temporalBoolConst(name string) *Const {
	return NewConst(name, Boolean)
}

func makeActionTerm(inputs, outputs []*Const, labels []string, stmt ActionsAction) *ActionTerm {
	return &ActionTerm{
		Inputs:  inputs,
		Outputs: outputs,
		Labels:  labels,
		Stmt:    stmt,
	}
}

// --- ActionTerm tests ---

func TestActionTermString_NoParams(t *testing.T) {
	stmt := NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	s := at.String()
	if !strings.HasPrefix(s, "action") {
		t.Errorf("expected ActionTerm.String() to start with 'action', got %q", s)
	}
	if !strings.Contains(s, "{") {
		t.Errorf("expected ActionTerm.String() to contain '{', got %q", s)
	}
}

func TestActionTermString_WithInputs(t *testing.T) {
	x := temporalBoolConst("x")
	stmt := NewSequence()
	at := makeActionTerm([]*Const{x}, nil, nil, stmt)
	s := at.String()
	if !strings.Contains(s, "x") {
		t.Errorf("expected ActionTerm.String() to mention input 'x', got %q", s)
	}
}

func TestActionTermString_WithOutputs(t *testing.T) {
	y := temporalBoolConst("y")
	stmt := NewSequence()
	at := makeActionTerm(nil, []*Const{y}, nil, stmt)
	s := at.String()
	if !strings.Contains(s, "returns") {
		t.Errorf("expected ActionTerm.String() to contain 'returns', got %q", s)
	}
}

func TestActionTermClone(t *testing.T) {
	stmt1 := NewSequence()
	stmt2 := NewAssumeAction(True)
	at := makeActionTerm(nil, nil, []string{"env1"}, stmt1)
	cloned := at.CloneStmt(stmt2)
	if cloned.Stmt != stmt2 {
		t.Error("CloneStmt should replace the statement")
	}
	if len(cloned.Labels) != 1 || cloned.Labels[0] != "env1" {
		t.Error("CloneStmt should preserve labels")
	}
}

// --- ActionTermBinding tests ---

func TestActionTermBindingString(t *testing.T) {
	stmt := NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	b := &ActionTermBinding{Name: "myaction", Action: at}
	s := b.String()
	if !strings.Contains(s, "myaction") {
		t.Errorf("expected binding string to contain name, got %q", s)
	}
	if !strings.Contains(s, "=") {
		t.Errorf("expected binding string to contain '=', got %q", s)
	}
}

func TestActionTermBindingClone(t *testing.T) {
	stmt := NewSequence()
	at1 := makeActionTerm(nil, nil, nil, stmt)
	at2 := makeActionTerm(nil, nil, []string{"x"}, stmt)
	b := &ActionTermBinding{Name: "test", Action: at1}
	cloned := b.CloneAction(at2)
	if cloned.Name != "test" {
		t.Error("CloneAction should preserve name")
	}
	if cloned.Action != at2 {
		t.Error("CloneAction should replace action")
	}
}

// --- NormalProgram tests ---

func TestNormalProgramString(t *testing.T) {
	stmt := NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	binding := &ActionTermBinding{Name: "act1", Action: at}
	np := &NormalProgram{
		Bindings: []*ActionTermBinding{binding},
		Init:     NewSequence(),
		Invars:   nil,
		Asms:     nil,
		Calls:    []string{"act1", "act2"},
	}
	s := np.String()
	if !strings.Contains(s, "let") {
		t.Error("expected NormalProgram string to contain 'let'")
	}
	if !strings.Contains(s, "while *") {
		t.Error("expected NormalProgram string to contain 'while *'")
	}
	if !strings.Contains(s, "diverge") {
		t.Error("expected NormalProgram string to contain 'diverge'")
	}
	if !strings.Contains(s, "act1") {
		t.Error("expected NormalProgram string to contain 'act1'")
	}
}

func TestNormalProgramBindingMap(t *testing.T) {
	stmt := NewAssumeAction(True)
	at := makeActionTerm(nil, nil, nil, stmt)
	binding := &ActionTermBinding{Name: "myact", Action: at}
	np := &NormalProgram{
		Bindings: []*ActionTermBinding{binding},
		Init:     NewSequence(),
		Calls:    nil,
	}
	bm := np.BindingMap()
	if _, ok := bm["myact"]; !ok {
		t.Error("BindingMap should contain 'myact'")
	}
}

func TestNormalProgramFormulas(t *testing.T) {
	stmt := NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	binding := &ActionTermBinding{Name: "a", Action: at}
	acfg := NewAstConfig()
	lf := acfg.NewLabeledFormula(nil, True)
	np := &NormalProgram{
		Bindings: []*ActionTermBinding{binding},
		Init:     NewSequence(),
		Invars:   []*LabeledFormula{lf},
		Asms:     nil,
		Calls:    nil,
	}
	fmlas := np.Formulas()
	// Should have binding + init + invar = 3
	if len(fmlas) != 3 {
		t.Errorf("expected 3 formulas, got %d", len(fmlas))
	}
}

// --- OldActionToNew / NewActionToOld ---

func TestOldActionToNewRoundTrip(t *testing.T) {
	act := NewAssumeAction(True)
	act.SetFormalParams([]*Const{temporalBoolConst("p")})
	act.SetFormalReturns([]*Const{temporalBoolConst("r")})
	at := OldActionToNew(act)
	if len(at.Inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(at.Inputs))
	}
	if len(at.Outputs) != 1 {
		t.Errorf("expected 1 output, got %d", len(at.Outputs))
	}
	back := NewActionToOld(at)
	if back != act {
		t.Error("roundtrip should return the original action")
	}
}

// --- NormalProgramFromModule ---

func TestNormalProgramFromModule_Empty(t *testing.T) {
	mod := New()
	np := NormalProgramFromModule(mod)
	if len(np.Bindings) != 0 {
		t.Error("empty module should have no bindings")
	}
	if len(np.Calls) != 0 {
		t.Error("empty module should have no calls")
	}
}

func TestNormalProgramFromModule_WithActions(t *testing.T) {
	mod := New()
	act := NewSequence()
	mod.Actions.Set("ext:myact", act)
	mod.PublicActions.Set("ext:myact", true)
	np := NormalProgramFromModule(mod)
	if len(np.Bindings) != 1 {
		t.Errorf("expected 1 binding, got %d", len(np.Bindings))
	}
	if len(np.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(np.Calls))
	}
}

// --- EnvAction ---

func TestEnvAction_Basic(t *testing.T) {
	stmt := NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	b1 := &ActionTermBinding{Name: "ext:foo", Action: at}
	b2 := &ActionTermBinding{Name: "bar", Action: at}
	env := TemporalEnvAction(NewActionsConfig(), []*ActionTermBinding{b1, b2})
	if env == nil {
		t.Fatal("EnvAction should not be nil")
	}
	if len(env.Branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(env.Branches))
	}
}

// --- PropEvent ---

func TestPropEvent_Eventually(t *testing.T) {
	body := temporalBoolConst("phi")
	notBody, _ := NewNot(body)
	ev, _ := NewEventually(nil, notBody)
	loc := Location{Line: 42}
	event := TemporalPropEvent(ev, loc)
	if _, ok := event.(*LogicAssertAction); !ok {
		t.Error("PropEvent for Eventually should produce AssertAction")
	}
	if event.GetLineno().Line != 42 {
		t.Errorf("expected lineno 42, got %d", event.GetLineno().Line)
	}
}

func TestPropEvent_Globally(t *testing.T) {
	body := temporalBoolConst("psi")
	g, _ := NewGlobally(nil, body)
	loc := Location{Line: 10}
	event := TemporalPropEvent(g, loc)
	if _, ok := event.(*LogicAssumeAction); !ok {
		t.Error("PropEvent for Globally should produce AssumeAction")
	}
}

// --- IsGloballyFormula / IsEventuallyFormula / IsTemporalFormula ---

func TestIsGloballyFormula(t *testing.T) {
	body := temporalBoolConst("x")
	g, _ := NewGlobally(nil, body)
	if !IsGloballyFormula(g) {
		t.Error("should be true for Globally")
	}
	if IsGloballyFormula(body) {
		t.Error("should be false for non-Globally")
	}
}

func TestIsEventuallyFormula(t *testing.T) {
	body := temporalBoolConst("x")
	e, _ := NewEventually(nil, body)
	if !IsEventuallyFormula(e) {
		t.Error("should be true for Eventually")
	}
}

func TestIsTemporalFormula(t *testing.T) {
	body := temporalBoolConst("x")
	g, _ := NewGlobally(nil, body)
	e, _ := NewEventually(nil, body)
	if !IsTemporalFormula(g) {
		t.Error("Globally should be temporal")
	}
	if !IsTemporalFormula(e) {
		t.Error("Eventually should be temporal")
	}
	if IsTemporalFormula(body) {
		t.Error("Const should not be temporal")
	}
}

func TestHasTemporalOperator(t *testing.T) {
	body := temporalBoolConst("x")
	g, _ := NewGlobally(nil, body)
	if !HasTemporalOperator(g) {
		t.Error("Globally should have temporal operator")
	}
	if HasTemporalOperator(body) {
		t.Error("plain const should not have temporal operator")
	}
}

func TestIsGprop(t *testing.T) {
	body := temporalBoolConst("x")
	g, _ := NewGlobally(nil, body)
	if !TemporalIsGprop(g) {
		t.Error("G(non-temporal) should be Gprop")
	}

	inner, _ := NewGlobally(nil, body)
	outer, _ := NewGlobally(nil, inner)
	if TemporalIsGprop(outer) {
		t.Error("G(G(x)) should not be Gprop (inner is temporal)")
	}
}

func TestInvarianceTacticCompilesAuxInvariantsBeforeCloningMainInvariant(t *testing.T) {
	cfg := NewAstConfig()
	mod := New()
	if _, err := mod.Sig.AddSymbol("b", Boolean); err != nil {
		t.Fatalf("AddSymbol(b): %v", err)
	}
	pc := NewProofChecker(nil, mod, nil, nil, nil, cfg)

	body := temporalBoolConst("b")
	globally, err := NewGlobally(nil, body)
	if err != nil {
		t.Fatalf("NewGlobally: %v", err)
	}
	model := &NormalProgram{Init: NewSequence()}
	goal := cfg.NewLabeledFormula(
		cfg.NewAtom("myprop"),
		cfg.NewTemporalModels(model, globally),
	)
	aux := cfg.NewLabeledFormula(cfg.NewAtom("inv1"), cfg.NewAtom("b"))
	proof := cfg.NewTacticTactic(
		cfg.NewAtom("invariance"),
		cfg.NewTacticWith([]Node{aux}),
		cfg.NewNoneAST(),
	)

	var result []*LabeledFormula
	out := captureActionUpdateStdout(t, func() {
		result, err = InvarianceTactic(pc, []*LabeledFormula{goal}, proof)
	})
	if err != nil {
		t.Fatalf("InvarianceTactic: %v\n%s", err, out)
	}

	compileIdx := strings.Index(out, "XTRACE: compiler.Thing ENTER type=LabeledFormula\n")
	cloneIdx := strings.Index(out, "XTRACE: ast.LF.__init__")
	if compileIdx < 0 {
		t.Fatalf("auxiliary invariant was not compiled:\n%s", out)
	}
	cloneGoalIdx := strings.Index(out, "XTRACE: proof.CloneGoal ENTER label=myprop nprems=0 concType=Const\n")
	if cloneGoalIdx < 0 {
		t.Fatalf("main invariant did not go through proof.CloneGoal:\n%s", out)
	}
	if cloneIdx < 0 {
		t.Fatalf("main invariant clone trace missing:\n%s", out)
	}
	if compileIdx > cloneGoalIdx || cloneGoalIdx > cloneIdx {
		t.Fatalf("expected auxiliary compile, then proof.CloneGoal, then LF allocation:\n%s", out)
	}

	if len(result) != 1 {
		t.Fatalf("got %d goals, want 1", len(result))
	}
	tm, ok := GoalConc(result[0]).(*TemporalModels)
	if !ok {
		t.Fatalf("result conclusion is %T, want *TemporalModels", GoalConc(result[0]))
	}
	np, ok := tm.Model.(*NormalProgram)
	if !ok {
		t.Fatalf("result model is %T, want *NormalProgram", tm.Model)
	}
	if len(np.Invars) != 2 {
		t.Fatalf("model invariant count = %d, want auxiliary plus main", len(np.Invars))
	}
}

// --- GetEnviron / EnvironStr ---

func TestGetEnviron(t *testing.T) {
	body := temporalBoolConst("x")
	g, _ := NewGlobally(nil, body)
	if GetEnviron(g) != nil {
		t.Error("nil environ should return nil")
	}

	env := "myenv"
	g2, _ := NewGlobally(&env, body)
	got := GetEnviron(g2)
	if got == nil || *got != "myenv" {
		t.Error("should return the environment")
	}
}

func TestEnvironStr(t *testing.T) {
	body := temporalBoolConst("x")
	g, _ := NewGlobally(nil, body)
	if EnvironStr(g) != "" {
		t.Error("nil environ should give empty string")
	}
	env := "e1"
	g2, _ := NewGlobally(&env, body)
	if EnvironStr(g2) != "e1" {
		t.Errorf("expected 'e1', got %q", EnvironStr(g2))
	}
}

// --- NormalProgramClone ---

func TestNormalProgramClone(t *testing.T) {
	stmt := NewSequence()
	at := makeActionTerm(nil, nil, nil, stmt)
	binding := &ActionTermBinding{Name: "a", Action: at}
	np := &NormalProgram{
		Bindings: []*ActionTermBinding{binding},
		Init:     NewSequence(),
		Invars:   nil,
		Asms:     nil,
		Calls:    []string{"a"},
	}
	cloned := NormalProgramClone(np)

	// Mirror Python's NormalProgram.clone (ivy_temporal.py:179-183): the
	// clone's slice fields are the SAME slice headers as the original.
	// Verify by checking that the slice base pointers (via reflect or
	// element identity through length-preserving operations) match.
	if &cloned.Calls[0] != &np.Calls[0] {
		t.Error("Calls: clone must share slice with original")
	}
	if &cloned.Bindings[0] != &np.Bindings[0] {
		t.Error("Bindings: clone must share slice with original")
	}
}

// --- PrefixActionTerm ---

func TestPrefixActionTerm(t *testing.T) {
	body := NewSequence()
	at := makeActionTerm(nil, nil, nil, body)
	prefix := NewAssumeAction(True)
	result := PrefixActionTerm(at, []ActionsAction{prefix})
	if result.Stmt == at.Stmt {
		t.Error("PrefixActionTerm should create a new statement")
	}
}

// --- Fuzz tests ---

func FuzzActionTermString(f *testing.F) {
	f.Add("act1", "x", "y")
	f.Fuzz(func(t *testing.T, name, inp, out string) {
		x := NewConst(inp, Boolean)
		y := NewConst(out, Boolean)
		stmt := NewSequence()
		at := makeActionTerm([]*Const{x}, []*Const{y}, nil, stmt)
		b := &ActionTermBinding{Name: name, Action: at}
		// Just verify no panic
		_ = b.String()
		_ = at.String()
	})
}

func FuzzNormalProgramString(f *testing.F) {
	f.Add("act1", "act2")
	f.Fuzz(func(t *testing.T, call1, call2 string) {
		stmt := NewSequence()
		at := makeActionTerm(nil, nil, nil, stmt)
		binding := &ActionTermBinding{Name: call1, Action: at}
		np := &NormalProgram{
			Bindings: []*ActionTermBinding{binding},
			Init:     NewSequence(),
			Calls:    []string{call1, call2},
		}
		// Should not panic
		_ = np.String()
	})
}
