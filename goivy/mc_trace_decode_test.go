package goivy

import (
	"os"
	"testing"
)

// --- Encoder decode method tests ---

func TestBitsToBoolExprs(t *testing.T) {
	bits := bitsToBoolExprs("101")
	if len(bits) != 3 {
		t.Fatalf("expected 3 bits, got %d", len(bits))
	}
	if !isTrueNode(bits[0]) {
		t.Error("bits[0] should be true")
	}
	if !isFalseNode(bits[1]) {
		t.Error("bits[1] should be false")
	}
	if !isTrueNode(bits[2]) {
		t.Error("bits[2] should be true")
	}
}

func TestBinDecBool(t *testing.T) {
	trueExpr := Expr(&LogicAnd{Terms: nil})
	falseExpr := Expr(&LogicOr{Terms: nil})

	tests := []struct {
		bits []Expr
		want int
	}{
		{[]Expr{falseExpr}, 0},
		{[]Expr{trueExpr}, 1},
		{[]Expr{trueExpr, falseExpr}, 2},
		{[]Expr{falseExpr, trueExpr}, 1},
		{[]Expr{trueExpr, trueExpr}, 3},
		{[]Expr{trueExpr, falseExpr, trueExpr}, 5},
	}
	for _, tc := range tests {
		got := binDecBool(tc.bits)
		if got != tc.want {
			t.Errorf("binDecBool(%v) = %d, want %d", tc.bits, got, tc.want)
		}
	}
}

func TestDecodeValBoolean(t *testing.T) {
	enc := &Encoder{Interp: make(map[string]interface{})}
	sym := NewConst("x", Boolean)
	trueExpr := Expr(&LogicAnd{Terms: nil})
	falseExpr := Expr(&LogicOr{Terms: nil})

	val := enc.DecodeVal([]Expr{trueExpr}, sym)
	if !isTrueNode(val) {
		t.Error("expected true for boolean bit '1'")
	}
	val = enc.DecodeVal([]Expr{falseExpr}, sym)
	if !isFalseNode(val) {
		t.Error("expected false for boolean bit '0'")
	}
}

func TestDecodeValEnumerated(t *testing.T) {
	es := &LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	enc := &Encoder{Interp: make(map[string]interface{})}
	sym := NewConst("c", es)

	trueExpr := Expr(&LogicAnd{Terms: nil})
	falseExpr := Expr(&LogicOr{Terms: nil})

	val := enc.DecodeVal([]Expr{falseExpr, falseExpr}, sym)
	if c, ok := val.(*Const); !ok || c.Name != "red" {
		t.Errorf("expected red, got %v", val)
	}

	val = enc.DecodeVal([]Expr{falseExpr, trueExpr}, sym)
	if c, ok := val.(*Const); !ok || c.Name != "green" {
		t.Errorf("expected green, got %v", val)
	}

	val = enc.DecodeVal([]Expr{trueExpr, falseExpr}, sym)
	if c, ok := val.(*Const); !ok || c.Name != "blue" {
		t.Errorf("expected blue, got %v", val)
	}

	// Clamp: index 3 should clamp to last element
	val = enc.DecodeVal([]Expr{trueExpr, trueExpr}, sym)
	if c, ok := val.(*Const); !ok || c.Name != "blue" {
		t.Errorf("expected blue (clamped), got %v", val)
	}
}

func TestDecodeValRange(t *testing.T) {
	rs := &RangeSort{
		Name: "idx",
		Lb:   NumeralBound{Value: "0"},
		Ub:   NumeralBound{Value: "3"},
	}
	enc := &Encoder{Interp: map[string]interface{}{"idx": rs}}
	sym := NewConst("i", rs)

	trueExpr := Expr(&LogicAnd{Terms: nil})
	falseExpr := Expr(&LogicOr{Terms: nil})

	val := enc.DecodeVal([]Expr{trueExpr, trueExpr}, sym)
	if c, ok := val.(*Const); !ok || c.Name != "3" {
		t.Errorf("expected 3 (clamped), got %v", val)
	}

	val = enc.DecodeVal([]Expr{falseExpr, trueExpr}, sym)
	if c, ok := val.(*Const); !ok || c.Name != "1" {
		t.Errorf("expected 1, got %v", val)
	}
}

func TestGetSymReturnsDecodedValue(t *testing.T) {
	boolSym := NewConst("x", Boolean)
	enc := NewEncoder([]*Const{boolSym}, nil, nil)
	enc.Interp = make(map[string]interface{})

	enc.Sub.Reset()
	enc.Sub.Step("1" + "0") // x=1, bogus=0

	val := enc.GetSym(boolSym)
	if !isTrueNode(val) {
		t.Error("expected true for input x=1")
	}
}

func TestGetNextSymReturnsNextState(t *testing.T) {
	boolSym := NewConst("s", Boolean)
	enc := NewEncoder(nil, []*Const{boolSym}, nil)
	enc.Interp = make(map[string]interface{})

	// Set next-state of latch to true (literal 1)
	subBits := enc.Encoding[Key(boolSym)]
	enc.Sub.Set(subBits[0], enc.Sub.True())

	enc.Sub.Reset()
	enc.Sub.Step("0") // bogus=0

	val := enc.GetNextSym(boolSym)
	if !isTrueNode(val) {
		t.Error("expected true for next-state of s")
	}
}

func TestGetEncoderState(t *testing.T) {
	s0 := NewConst("s0", Boolean)
	s1 := NewConst("s1", Boolean)
	enc := NewEncoder(nil, []*Const{s0, s1}, nil)
	enc.Interp = make(map[string]interface{})

	stmap := enc.GetEncoderState("10")
	v0, ok0 := stmap[Key(s0)]
	v1, ok1 := stmap[Key(s1)]
	if !ok0 || !ok1 {
		t.Fatal("expected both latches in state map")
	}
	if !isTrueNode(v0) {
		t.Error("s0 should be true")
	}
	if !isFalseNode(v1) {
		t.Error("s1 should be false")
	}
}

// --- AigerMatchHandler2 tests ---

func TestAigerMatchHandler2ImplementsAnnotationHandler(t *testing.T) {
	var _ AnnotationHandler = (*AigerMatchHandler2)(nil)
}

func TestAigerMatchHandler2EvalFalseTrue(t *testing.T) {
	h := NewAigerMatchHandler2(nil, nil, nil, nil, nil)

	if h.Eval(&LogicOr{Terms: nil}) {
		t.Error("expected false for LogicOr{}")
	}
	if !h.Eval(&LogicAnd{Terms: nil}) {
		t.Error("expected true for LogicAnd{}")
	}

	notTrue := &LogicNot{Body: &LogicAnd{Terms: nil}}
	if h.Eval(notTrue) {
		t.Error("expected false for Not(true)")
	}
}

func TestAigerMatchHandler2Clone(t *testing.T) {
	decoder := map[string]Expr{"x": NewConst("x", Boolean)}
	h := NewAigerMatchHandler2(nil, decoder, nil, nil, nil)
	h.States = append(h.States, []Expr{NewConst("a", Boolean)})

	clone := h.Clone()
	if clone == h {
		t.Error("clone should be a different object")
	}
	if len(clone.States) != 0 {
		t.Error("clone should start with empty states")
	}
	if clone.Decoder["x"] == nil {
		t.Error("clone should share decoder")
	}
}

func TestAigerMatchHandler2AddState(t *testing.T) {
	h := NewAigerMatchHandler2(nil, nil, nil, nil, nil)
	eqns := []Expr{
		&Eq{T1: NewConst("x", Boolean), T2: &LogicAnd{Terms: nil}},
	}
	h.AddState(eqns)
	if len(h.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(h.States))
	}
	if len(h.States[0]) != 1 {
		t.Fatalf("expected 1 equation, got %d", len(h.States[0]))
	}
}

func TestAigerMatchHandler2EndCallsFinalState(t *testing.T) {
	boolSym := NewConst("s", Boolean)
	enc := NewEncoder(nil, []*Const{boolSym}, nil)
	enc.Interp = make(map[string]interface{})

	subBits := enc.Encoding[Key(boolSym)]
	enc.Sub.Set(subBits[0], enc.Sub.True())

	enc.Sub.Reset()
	enc.Sub.Step("0") // bogus=0

	decoder := map[string]Expr{boolSym.Name: boolSym}
	h := NewAigerMatchHandler2(enc, decoder, nil, nil, nil)

	h.End()
	if len(h.States) < 1 {
		t.Error("End() should have added a final state")
	}
}

func TestAigerMatchHandler2NewStateUsesStructuralEnvRenameForNextLatch(t *testing.T) {
	x := NewConst("x", Boolean)
	newX := NewConst("new_x", Boolean)
	enc := NewEncoder(nil, []*Const{x}, nil)
	enc.Interp = make(map[string]interface{})
	enc.SetSym(x, []int{enc.Sub.True()})

	enc.Sub.Reset()
	enc.Sub.Step("0") // bogus=0

	h := NewAigerMatchHandler2(
		enc,
		map[string]Expr{x.Name: x},
		map[string]bool{},
		map[string]bool{x.Name: true},
		nil,
	)
	h.NewState(map[NodeKey]Expr{Key(x): newX})

	if len(h.States) != 1 {
		t.Fatalf("expected one decoded state, got %d", len(h.States))
	}
	if len(h.States[0]) != 1 {
		t.Fatalf("expected env-renamed next latch equation, got %d: %#v", len(h.States[0]), h.States[0])
	}
	eq, ok := h.States[0][0].(*Eq)
	if !ok {
		t.Fatalf("expected Eq, got %T", h.States[0][0])
	}
	if !eq.T1.Equal(x) || !IsTrue(eq.T2) {
		t.Fatalf("expected x = true from env-renamed next latch, got %s", eq)
	}
}

func TestAigerMatchHandler2ShowSymSkipsEnvFormalLikePython(t *testing.T) {
	client := &UninterpretedSort{Name: "client"}
	message1 := &LogicEnumeratedSort{Name: "message1", Extension: []string{"empty1", "reqshared", "reqexclusive"}}
	channel1 := NewConst("s.channel1", &LogicFunctionSort{Sorts: []Sort{client, message1}})
	formal := NewConst("__fml:cl", client)
	decd := MustApply(channel1, formal)
	val := NewConst("reqshared", message1)

	handler := NewAigerMatchHandler2(nil, nil, nil, nil, New())
	handler.Depth = 1

	var eqns []Expr
	handler.showSym2(
		"__abs[0]",
		decd,
		val,
		map[string]string{},
		map[string]bool{"__fml:cl": true},
		&eqns,
	)
	if len(eqns) != 0 {
		t.Fatalf("showSym2 emitted %d eqns for env-bound __fml formal, want 0: %v", len(eqns), eqns)
	}
}

func TestAigerWitnessToIvyTrace2CallsMatchAnnotation(t *testing.T) {
	// Build a minimal circuit: one boolean input, one boolean latch, one output
	input := NewConst("inp", Boolean)
	latch := NewConst("st", Boolean)
	output := NewConst("out", Boolean)

	enc := NewEncoder([]*Const{input}, []*Const{latch}, []*Const{output})
	enc.Interp = make(map[string]interface{})

	// Set latch next-state = input (simple pass-through)
	inpBits := enc.Encoding[Key(input)]
	latchBits := enc.Encoding[Key(latch)]
	inpLit, _ := enc.Sub.Lit(inpBits[0])
	enc.Sub.Set(latchBits[0], inpLit)

	// Set output = latch
	outBits := enc.Encoding[Key(output)]
	latchLit, _ := enc.Sub.Lit(latchBits[0])
	enc.Sub.Set(outBits[0], latchLit)

	decoder := map[string]Expr{
		input.Name: input,
		latch.Name: latch,
	}

	// Create a minimal annotation: EmptyAnnotation (simplest valid annotation)
	annot := EmptyAnnotation{}

	// Create a Sequence() action (empty sequence matches EmptyAnnotation)
	action := NewSequence()

	result := &ToAigerResult{
		Aiger:    enc,
		Decoder:  decoder,
		Annot:    annot,
		Consts:   make(map[string]bool),
		Action:   action,
		StVarSet: make(map[string]bool),
	}

	// Write a witness file with one step: "0 10 0 0"
	// (pre=0, inp=10 [input=1, bogus=0], out=0, post=0)
	tmpFile, err := os.CreateTemp("", "test_witness_*.out")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("1\n0 10 0 0\n")
	tmpFile.Close()

	mod := New()

	handler, err := AigerWitnessToIvyTrace2(result, tmpFile.Name(), mod)
	if err != nil {
		t.Fatalf("AigerWitnessToIvyTrace2 failed: %v", err)
	}
	if handler == nil {
		t.Fatal("handler should not be nil")
	}
	// The handler should have collected at least one state from End()->FinalState()
	if len(handler.States) == 0 {
		t.Error("handler should have collected at least one state (from End/FinalState)")
	}
}

func TestAigerWitnessToIvyTrace2RejectsBadWitness(t *testing.T) {
	result := &ToAigerResult{
		Aiger:    NewEncoder(nil, nil, nil),
		Decoder:  make(map[string]Expr),
		Consts:   make(map[string]bool),
		StVarSet: make(map[string]bool),
	}

	// Write a witness file with wrong header
	tmpFile, err := os.CreateTemp("", "test_badwit_*.out")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("0\n")
	tmpFile.Close()

	_, err = AigerWitnessToIvyTrace2(result, tmpFile.Name(), nil)
	if err == nil {
		t.Error("expected error for bad witness header")
	}
}

func TestAigerMatchHandler2HandleDelegation(t *testing.T) {
	h := NewAigerMatchHandler2(nil, nil, nil, nil, nil)

	// Create a mock Sequence with a non-"nowhere" lineno
	callAction := NewSequence()
	callAction.SetLineno(Location{Line: 1, Filename: "test.ivy"})

	// First call: should set LastAction
	emptyEnv := make(map[NodeKey]Expr)
	h.Handle(callAction, emptyEnv)
	if h.LastAction == nil {
		t.Error("LastAction should be set after Handle")
	}
}

func TestAigerMatchHandler2Fail(t *testing.T) {
	h := NewAigerMatchHandler2(nil, nil, nil, nil, nil)
	mockAction := NewSequence()
	h.LastAction = mockAction

	h.Fail()
	if _, ok := h.LastAction.(*FailAction); !ok {
		t.Error("Fail() should wrap LastAction in FailAction")
	}
}

func TestIsFalseNode(t *testing.T) {
	if !isFalseNode(&LogicOr{Terms: nil}) {
		t.Error("LogicOr{} should be false")
	}
	if !isFalseNode(NewConst("false", Boolean)) {
		t.Error("Const 'false' should be false")
	}
	if isFalseNode(&LogicAnd{Terms: nil}) {
		t.Error("LogicAnd{} should not be false")
	}
}

func TestIsTrueNode(t *testing.T) {
	if !isTrueNode(&LogicAnd{Terms: nil}) {
		t.Error("LogicAnd{} should be true")
	}
	if !isTrueNode(NewConst("true", Boolean)) {
		t.Error("Const 'true' should be true")
	}
	if isTrueNode(&LogicOr{Terms: nil}) {
		t.Error("LogicOr{} should not be true")
	}
}
