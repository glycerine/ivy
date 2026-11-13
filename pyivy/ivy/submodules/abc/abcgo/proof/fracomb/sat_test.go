package fracomb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSATTwoOutputTruthTables(t *testing.T) {
	for aTable := 0; aTable < 16; aTable++ {
		for bTable := 0; bTable < 16; bTable++ {
			for _, andOutputs := range []int32{0, 1} {
				for _, solver := range []int32{0, 1} {
					g := aig.NewGraph()
					x, y := g.NewInput(), g.NewInput()
					mk := func(table int) aig.Edge {
						out := g.Const0()
						for assignment := 0; assignment < 4; assignment++ {
							if table&(1<<assignment) == 0 {
								continue
							}
							a, b := x, y
							if assignment&1 == 0 {
								a = a.Not()
							}
							if assignment&2 == 0 {
								b = b.Not()
							}
							out = g.OR(out, g.AND(a, b))
						}
						return out
					}
					g.AddOutput(mk(aTable))
					g.AddOutput(mk(bTable))
					table := aTable | bTable
					if andOutputs != 0 {
						table = aTable & bTable
					}
					got := Fra_FraigSat(g, 0, 0, 0, 0, 0, 0, andOutputs, solver, 0)
					if (got == 0) != (table != 0) {
						t.Fatalf("tables %x,%x mode %d solver %d result %d", aTable, bTable, andOutputs, solver, got)
					}
					if got == 0 {
						model, ok := g.Data.([]int32)
						if !ok || len(model) != 2 {
							t.Fatalf("invalid model %v", g.Data)
						}
						outputs, err := g.Evaluate([]bool{model[0] != 0, model[1] != 0})
						if err != nil {
							t.Fatal(err)
						}
						valid := outputs[0] || outputs[1]
						if andOutputs != 0 {
							valid = outputs[0] && outputs[1]
						}
						if !valid {
							t.Fatalf("invalid model %v outputs %v", model, outputs)
						}
					}
				}
			}
		}
	}
}
