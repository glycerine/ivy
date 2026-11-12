package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func structuralConstraintFixture() (*aig.Graph, aig.Edge, aig.Edge, aig.Edge) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	lo := g.NewInput()
	property := g.AND(g.AND(lo.Not(), x), y)
	g.AddOutput(property)
	// The next-state function is !x OR lo. Its complement is an AND
	// supergate containing the current latch output and complemented x term.
	g.AddOutput(g.AND(lo.Not(), x).Not())
	aig.Aig_ManSetRegNum(g, 1)
	return g, x, y, lo
}

func TestDetectStructuralConstraintFactors(t *testing.T) {
	g, x, y, lo := structuralConstraintFixture()
	split, ok, err := DetectStructuralConstraint(g, 0)
	if err != nil || !ok {
		t.Fatalf("detect: ok=%t err=%v", ok, err)
	}
	if split.ConstantFalse || len(split.PropertyNegations) != 1 || split.PropertyNegations[0] != y.Not() {
		t.Fatalf("property factors: %+v", split)
	}
	if len(split.ConstraintNegations) != 1 || split.ConstraintNegations[0] != x.Not() {
		t.Fatalf("constraint factors: %+v", split)
	}
	if split.PropertyNegations[0] == lo || split.ConstraintNegations[0] == lo {
		t.Fatal("latch factor leaked into returned properties/constraints")
	}
}

func TestSourceNameDetectorStatus(t *testing.T) {
	g, _, _, _ := structuralConstraintFixture()
	_, _, status := Saig_ManDetectConstr(g, 0)
	if status != 1 {
		t.Fatalf("matched constraint count = %d, want 1", status)
	}
	plain := aig.NewGraph()
	plain.NewInput()
	plain.AddOutput(plain.Const0())
	plain.AddOutput(plain.Const0())
	aig.Aig_ManSetRegNum(plain, 1)
	_, _, status = Saig_ManDetectConstr(plain, 0)
	if status != -1 {
		t.Fatalf("constant-false status = %d, want -1", status)
	}
}

func TestDetectConstantFalseOutput(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 1)
	split, ok, err := DetectStructuralConstraint(g, 0)
	if err != nil || !ok || !split.ConstantFalse || len(split.PropertyNegations) != 1 || split.PropertyNegations[0] != g.Const1() {
		t.Fatalf("constant false detection: split=%+v ok=%t err=%v", split, ok, err)
	}
}

func TestUnfoldStructuralConstraints(t *testing.T) {
	g, x, y, lo := structuralConstraintFixture()
	unfolded, changed, err := UnfoldStructuralConstraints(g)
	if err != nil || !changed {
		t.Fatalf("unfold: changed=%t err=%v", changed, err)
	}
	if aig.Aig_ManRegNum(unfolded) != 1 || aig.Aig_ManConstrNum(unfolded) != 1 || unfolded.OutputCount() != 3 {
		t.Fatalf("wrong unfolded interface: regs=%d constraints=%d outputs=%d", aig.Aig_ManRegNum(unfolded), aig.Aig_ManConstrNum(unfolded), unfolded.OutputCount())
	}
	for bits := 0; bits < 8; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		values, evalErr := unfolded.Evaluate(inputs)
		if evalErr != nil {
			t.Fatal(evalErr)
		}
		if values[0] != inputs[1] || values[1] != !inputs[0] || values[2] != (inputs[2] || !inputs[0]) {
			t.Fatalf("wrong unfolded outputs for x=%t y=%t lo=%t: %v", inputs[0], inputs[1], inputs[2], values)
		}
	}
	_ = x
	_ = y
	_ = lo
}

func TestUnfoldFallsBackWithoutChangingBehavior(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	copy, changed, err := UnfoldStructuralConstraints(g)
	if err != nil || changed {
		t.Fatalf("fallback: changed=%t err=%v", changed, err)
	}
	for bits := 0; bits < 4; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0}
		want, wantErr := g.Evaluate(inputs)
		got, gotErr := copy.Evaluate(inputs)
		if wantErr != nil || gotErr != nil || len(want) != len(got) || want[0] != got[0] || want[1] != got[1] {
			t.Fatalf("fallback changed behavior for %v: want=%v got=%v (%v, %v)", inputs, want, got, wantErr, gotErr)
		}
	}
}

func TestUnfoldFallsBackWhenOutputsDisagreeOnConstraints(t *testing.T) {
	g := aig.NewGraph()
	x, y, a, b := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	lo0, lo1 := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.AND(lo0.Not(), x), a))
	g.AddOutput(g.AND(g.AND(lo1.Not(), y), b))
	g.AddOutput(g.AND(lo0.Not(), x).Not())
	g.AddOutput(g.AND(lo1.Not(), y).Not())
	aig.Aig_ManSetRegNum(g, 2)

	copy, changed, err := UnfoldStructuralConstraints(g)
	if err != nil || changed {
		t.Fatalf("incompatible constraints should fall back: changed=%t err=%v", changed, err)
	}
	for bits := 0; bits < 64; bits++ {
		inputs := make([]bool, 6)
		for i := range inputs {
			inputs[i] = bits&(1<<i) != 0
		}
		want, wantErr := g.Evaluate(inputs)
		got, gotErr := copy.Evaluate(inputs)
		if wantErr != nil || gotErr != nil || len(want) != len(got) {
			t.Fatalf("eval assignment %06b: want err=%v got err=%v", bits, wantErr, gotErr)
		}
		for i := range want {
			if want[i] != got[i] {
				t.Fatalf("fallback changed assignment %06b output %d: %v != %v", bits, i, want, got)
			}
		}
	}
}

func TestUnfoldPreservesConstantFalseProperties(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 1)

	unfolded, changed, err := UnfoldStructuralConstraints(g)
	if err != nil || !changed {
		t.Fatalf("constant-false unfold: changed=%t err=%v", changed, err)
	}
	if unfolded.OutputCount() != 2 || aig.Aig_ManConstrNum(unfolded) != 0 {
		t.Fatalf("unexpected constant-false result: outputs=%d constraints=%d", unfolded.OutputCount(), aig.Aig_ManConstrNum(unfolded))
	}
	for _, value := range []bool{false, true} {
		got, evalErr := unfolded.Evaluate([]bool{value})
		if evalErr != nil || got[0] {
			t.Fatalf("constant-false property became true for state=%t: %v (%v)", value, got, evalErr)
		}
	}
}

func TestFoldStructuralConstraintsAllAssignments(t *testing.T) {
	g := aig.NewGraph()
	x, y, lo := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(x)
	g.AddOutput(y)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)

	folded, err := FoldStructuralConstraints(g, []int{1})
	if err != nil {
		t.Fatal(err)
	}
	if aig.Aig_ManRegNum(folded) != 2 || aig.Aig_ManConstrNum(folded) != 1 || folded.OutputCount() != 4 {
		t.Fatalf("wrong folded interface: regs=%d constraints=%d outputs=%d", aig.Aig_ManRegNum(folded), aig.Aig_ManConstrNum(folded), folded.OutputCount())
	}
	for bits := 0; bits < 16; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
		got, evalErr := folded.Evaluate(inputs)
		if evalErr != nil {
			t.Fatal(evalErr)
		}
		violation := inputs[1] || inputs[3]
		want := []bool{inputs[0] && !violation, inputs[1] && !violation, inputs[2], violation}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("assignment %04b output %d: got %t, want %t", bits, i, got[i], want[i])
			}
		}
	}
}

func TestFoldRejectsInvalidConstraintOutput(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	for _, index := range []int{-1, 0, 1} {
		if _, err := FoldStructuralConstraints(g, []int{index}); err == nil {
			t.Fatalf("accepted invalid/reserved output index %d", index)
		}
	}
}
