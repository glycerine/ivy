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

func TestSATPolarityTransformPreservesInputWitness(t *testing.T) {
	for _, solver := range []int32{0, 1} {
		g := aig.NewGraph()
		x := g.NewInput()
		g.AddOutput(x)
		aig.Aig_ObjSetPhase(g, x, true)
		if got := Fra_FraigSat(g, 0, 0, 0, 0, 0, 1, 0, solver, 0); got != 0 {
			t.Fatalf("solver %d got %d", solver, got)
		}
		model, ok := g.Data.([]int32)
		if !ok || len(model) != 1 || model[0] != 1 {
			t.Fatalf("solver %d invalid model %v", solver, g.Data)
		}
	}
}

func TestSATEmptyOutputCombination(t *testing.T) {
	for _, andOutputs := range []int32{0, 1} {
		for _, solver := range []int32{0, 1} {
			g := aig.NewGraph()
			g.NewInput()
			got := Fra_FraigSat(g, 0, 0, 0, 0, 0, 1, andOutputs, solver, 0)
			want := int32(1)
			if andOutputs != 0 {
				want = 0
			}
			if got != want {
				t.Fatalf("and=%d solver=%d got=%d want=%d", andOutputs, solver, got, want)
			}
		}
	}
}

func TestSATAllTruthFunctionsWithInputPhases(t *testing.T) {
	for table := 0; table < 16; table++ {
		for phases := 0; phases < 4; phases++ {
			for _, solver := range []int32{0, 1} {
				for _, flip := range []int32{0, 1} {
					g := aig.NewGraph()
					x, y := g.NewInput(), g.NewInput()
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
					g.AddOutput(out)
					aig.Aig_ObjSetPhase(g, x, phases&1 != 0)
					aig.Aig_ObjSetPhase(g, y, phases&2 != 0)
					status := Fra_FraigSat(g, 0, 0, 0, 0, 0, flip, 0, solver, 0)
					if (status == 0) != (table != 0) {
						t.Fatalf("table=%x phases=%d solver=%d flip=%d status=%d", table, phases, solver, flip, status)
					}
					if status == 0 {
						m, ok := g.Data.([]int32)
						if !ok || len(m) != 2 {
							t.Fatalf("bad model %v", g.Data)
						}
						o, err := g.Evaluate([]bool{m[0] != 0, m[1] != 0})
						if err != nil || !o[0] {
							t.Fatalf("table=%x phases=%d solver=%d flip=%d model=%v", table, phases, solver, flip, m)
						}
					}
				}
			}
		}
	}
}
