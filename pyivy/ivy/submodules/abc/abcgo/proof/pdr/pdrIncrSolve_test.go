package pdr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestIncrementalPdrRestartAndWitness(t *testing.T) {
	g := aig.NewGraph()
	x, q0, q1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(q1)
	g.AddOutput(x)
	g.AddOutput(q0)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.fSilent, pars.nFrameMax = 1, 1
	result := IPdr_ManSolve(g, &pars)
	if result != 0 {
		t.Fatalf("got %d, want counterexample", result)
	}
	if pars.nFrameMax <= 1 {
		t.Fatal("incremental driver did not restart")
	}
	if g.SequentialCounterexample() == nil || saig.Saig_ManVerifyCex(g, g.SequentialCounterexample()) != 1 {
		t.Fatal("missing or invalid counterexample")
	}
}

func TestIncrementalPdrSafeAndUnsafe(t *testing.T) {
	for _, bad := range []bool{false, true} {
		g := aig.NewGraph()
		x, q := g.NewInput(), g.NewInput()
		g.AddOutput(q)
		if bad {
			g.AddOutput(x)
		} else {
			g.AddOutput(g.Const0())
		}
		aig.Aig_ManSetRegNum(g, 1)
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fSilent = 1
		result := IPdr_ManSolve(g, &pars)
		want := int32(1)
		if bad {
			want = 0
		}
		if result != want {
			t.Fatalf("bad=%v got %d, want %d", bad, result, want)
		}
		if bad && (g.SequentialCounterexample() == nil || saig.Saig_ManVerifyCex(g, g.SequentialCounterexample()) != 1) {
			t.Fatal("missing or invalid counterexample")
		}
	}
}

func TestIncrementalPdrRestartPreservesSolvedOutputs(t *testing.T) {
	g := aig.NewGraph()
	x, q0, q1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(x)
	g.AddOutput(q1)
	g.AddOutput(x)
	g.AddOutput(q0)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.fSilent, pars.fNotVerbose, pars.fSolveAll, pars.fStoreCex, pars.nFrameMax = 1, 1, 1, 1, 1
	result := IPdr_ManSolve(g, &pars)
	if result != 0 || pars.nFailOuts != 2 {
		t.Fatalf("result %d failed outputs %d", result, pars.nFailOuts)
	}
	if pars.nFrameMax <= 1 {
		t.Fatal("expected restart")
	}
	witnesses := g.SequentialCounterexamples()
	if witnesses == nil {
		t.Fatal("missing output witnesses")
	}
	for i := 0; i < 2; i++ {
		cex, ok := vec.Vec_PtrEntry(witnesses, i).(*util.Abc_Cex_t)
		if !ok || cex == nil || saig.Saig_ManVerifyCex(g, cex) != 1 {
			t.Fatalf("output %d missing or invalid witness", i)
		}
	}
}

func TestIncrementalPdrRestartDoesNotRecountProvedOutputs(t *testing.T) {
	g := aig.NewGraph()
	x, q0, q1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(q1)
	g.AddOutput(x)
	g.AddOutput(q0)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.fSilent, pars.fNotVerbose, pars.fSolveAll, pars.nFrameMax = 1, 1, 1, 1
	result := IPdr_ManSolve(g, &pars)
	if result != 0 || pars.nFailOuts != 1 || pars.nProveOuts != 1 {
		t.Fatalf("result %d failed %d proved %d", result, pars.nFailOuts, pars.nProveOuts)
	}
}

func TestIncrementalPdrAgainstExhaustiveReachability(t *testing.T) {
	for trans := 0; trans < 16; trans++ {
		for prop := 0; prop < 16; prop++ {
			g := aig.NewGraph()
			in := g.NewInput()
			latch := g.NewInput()
			truth := func(table int) aig.Edge {
				result := g.Const0()
				for row := 0; row < 4; row++ {
					if table&(1<<row) == 0 {
						continue
					}
					x, y := in, latch
					if row&1 == 0 {
						x = x.Not()
					}
					if row&2 == 0 {
						y = y.Not()
					}
					term := g.AND(x, y)
					result = g.AND(result.Not(), term.Not()).Not()
				}
				return result
			}
			g.AddOutput(truth(prop))
			g.AddOutput(truth(trans))
			aig.Aig_ManSetRegNum(g, 1)
			reached := [2]bool{true, false}
			bad := false
			for changed := true; changed; {
				changed = false
				for state := 0; state < 2; state++ {
					if !reached[state] {
						continue
					}
					for input := 0; input < 2; input++ {
						row := input + 2*state
						if prop&(1<<row) != 0 {
							bad = true
						}
						next := (trans >> row) & 1
						if !reached[next] {
							reached[next] = true
							changed = true
						}
					}
				}
			}
			var pars Pdr_Par_t
			Pdr_ManSetDefaultParams(&pars)
			pars.fSilent = 1
			pars.nFrameMax = 1
			got := IPdr_ManSolve(g, &pars)
			want := int32(1)
			if bad {
				want = 0
			}
			if got != want {
				t.Fatalf("transition %#x property %#x: got %d want %d", trans, prop, got, want)
			}
			if got == 0 && (g.SequentialCounterexample() == nil || saig.Saig_ManVerifyCex(g, g.SequentialCounterexample()) != 1) {
				t.Fatalf("transition %#x property %#x: invalid counterexample", trans, prop)
			}
		}
	}
}
