package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestFraigSatOutputModesAndWitness(t *testing.T) {
	for _, newSolver := range []int32{0, 1} {
		for _, flipBits := range []int32{0, 1} {
			for _, andOutputs := range []int32{0, 1} {
				g := aig.NewGraph()
				x := g.NewInput()
				g.AddOutput(x)
				g.AddOutput(x.Not())
				status := Fra_FraigSat(g, 0, 0, 0, 0, 0, flipBits, andOutputs, newSolver, 0)
				if andOutputs != 0 {
					if status != 1 || g.Data != nil {
						t.Fatalf("solver=%d flip=%d AND: status=%d data=%v", newSolver, flipBits, status, g.Data)
					}
					continue
				}
				if status != 0 {
					t.Fatalf("solver=%d flip=%d OR: status=%d", newSolver, flipBits, status)
				}
				model, ok := g.Data.([]int32)
				if !ok || len(model) != 1 {
					t.Fatalf("invalid witness %v", g.Data)
				}
				outputs, err := g.Evaluate([]bool{model[0] != 0})
				if err != nil || !outputs[0] && !outputs[1] {
					t.Fatalf("witness %v gives outputs %v: %v", model, outputs, err)
				}
			}
		}
	}
}

func TestAigManCountXors(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.XOR(x, y))
	if got := Aig_ManCountXors(g); got != 1 {
		t.Fatalf("XOR count = %d, want 1", got)
	}
}

func TestFraigSatAllTwoInputFunctions(t *testing.T) {
	for table := 0; table < 16; table++ {
		for solver := int32(0); solver < 2; solver++ {
			for flip := int32(0); flip < 2; flip++ {
				g := aig.NewGraph()
				x, y := g.NewInput(), g.NewInput()
				root := g.Const0()
				for mask := 0; mask < 4; mask++ {
					if table&(1<<mask) == 0 {
						continue
					}
					a, b := x, y
					if mask&1 == 0 {
						a = a.Not()
					}
					if mask&2 == 0 {
						b = b.Not()
					}
					root = g.OR(root, g.AND(a, b))
				}
				g.AddOutput(root)
				got := Fra_FraigSat(g, 0, 0, 0, 0, 0, flip, 0, solver, 0)
				if (got == 0) != (table != 0) {
					t.Fatalf("table=%04b solver=%d flip=%d got=%d", table, solver, flip, got)
				}
				if got == 0 {
					model := g.Data.([]int32)
					out, err := g.Evaluate([]bool{model[0] != 0, model[1] != 0})
					if err != nil || len(out) != 1 || !out[0] {
						t.Fatalf("table=%04b model=%v output=%v err=%v", table, model, out, err)
					}
				}
			}
		}
	}
}
