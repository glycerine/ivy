package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"testing"
)

func TestLatchCorrespondencePreservesEquivalentLatches(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.XOR(a, b))
	next := g.AND(x, y)
	g.AddOutput(next)
	g.AddOutput(next)
	aig.Aig_ManSetRegNum(g, 2)
	var iterations int32
	reduced := Fra_FraigLatchCorrespondence(g, 0, 1000, 0, 0, &iterations, 0)
	if reduced == nil || iterations == 0 {
		t.Fatalf("no reduction or iterations %d", iterations)
	}
	regs := int(reduced.RegisterCount())
	if regs > 1 {
		t.Fatalf("increased latch count: %d", regs)
	}
	for mask := 0; mask < 16; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		for i := 0; i < regs; i++ {
			inputs = append(inputs, false)
		}
		for frame := 0; frame < 3; frame++ {
			out, err := reduced.Evaluate(inputs)
			if err != nil || out[0] {
				t.Fatalf("mask=%d frame=%d out=%v err=%v", mask, frame, out, err)
			}
			inputs = append(inputs[:2], out[len(out)-regs:]...)
		}
	}
}

func TestLatchCorrespondenceKeepsDistinctLatches(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.XOR(a, b))
	g.AddOutput(g.AND(x, y))
	g.AddOutput(g.XOR(x, y))
	aig.Aig_ManSetRegNum(g, 2)
	var it int32
	reduced := Fra_FraigLatchCorrespondence(g, 0, 1000, 0, 0, &it, 0)
	if reduced == nil || reduced.RegisterCount() != 2 {
		t.Fatalf("distinct latches lost: %+v", reduced)
	}
	for mask := 0; mask < 4; mask++ {
		xx, yy := mask&1 != 0, mask&2 != 0
		out, err := reduced.Evaluate([]bool{xx, yy, false, false})
		if err != nil {
			t.Fatal(err)
		}
		out, err = reduced.Evaluate([]bool{false, false, out[1], out[2]})
		if err != nil || out[0] != ((xx && yy) != (xx != yy)) {
			t.Fatalf("mask %d out=%v err=%v", mask, out, err)
		}
	}
}

func TestLatchCorrespondenceProducesCounterexample(t *testing.T) {
	g := aig.NewGraph()
	x, a := g.NewInput(), g.NewInput()
	g.NewInput()
	g.AddOutput(g.AND(x, a))
	g.AddOutput(x)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 2)
	if out := Fra_FraigLatchCorrespondence(g, 0, 10000, 1, 0, nil, 0); out != nil {
		t.Fatal("expected failing miter to stop correspondence")
	}
	cex := g.SequentialCounterexample()
	if cex == nil || saig.Saig_ManVerifyCex(g, cex) == 0 {
		t.Fatalf("invalid counterexample: %+v", cex)
	}
}

func TestLatchCorrespondenceExpiredTimeBudget(t *testing.T) {
	g := aig.NewGraph()
	x, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.XOR(a, b))
	g.AddOutput(x)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 2)
	iterations := int32(-1)
	if out := Fra_FraigLatchCorrespondence(g, 0, 10000, 0, 0, &iterations, 1e-12); out != nil {
		t.Fatal("expired budget returned a reduction")
	}
	if iterations != 0 {
		t.Fatalf("iterations=%d", iterations)
	}
}

func TestLatchPartitionTreatsNativeExorAsAnd(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.NewInput() // latch output
	g.AddOutput(g.NewXor(x, y))
	aig.Aig_ManSetRegNum(g, 1)
	p := Lcr_ManAlloc(g)
	p.classes = Fra_ClassesStart(g)
	p.marked[0] = true
	part := Fra_LcrCreatePart(p, []int32{0})
	for mask := 0; mask < 4; mask++ {
		values := make([]bool, part.InputCount())
		values[len(values)-2] = mask&1 != 0
		values[len(values)-1] = mask&2 != 0
		got, err := part.Evaluate(values)
		if err != nil || got[0] != (mask == 3) {
			t.Fatalf("mask=%d output=%v err=%v", mask, got, err)
		}
	}
}
