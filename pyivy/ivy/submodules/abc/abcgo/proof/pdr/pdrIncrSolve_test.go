package pdr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestIncrementalFirstRestartAssertsOnNilSavedClauses(t *testing.T) {
	g := aig.NewGraph()
	x, q0, q1 := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(q1)
	g.AddOutput(x)
	g.AddOutput(q0)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.fSilent, pars.nFrameMax = 1, 1
	defer func() {
		if recover() == nil {
			t.Fatal("expected C's nil saved-clause assertion")
		}
	}()
	IPdr_ManSolve(g, &pars)
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

func TestIncrementalRestartRecreatesOutputStateAsInC(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const0())
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.fSolveAll, pars.nTimeOutOne = 1, 50
	first := Pdr_ManStart(g, &pars, nil)
	vec.Vec_PtrWriteEntry(first.vCexes, 0, "solved")
	vec.Vec_IntWriteEntry(pars.vOutMap, 0, 0)
	first.pTime4Outs[0] = 0
	pars.nFailOuts = 1
	Pdr_ManStop(first)
	second := Pdr_ManStart(g, &pars, nil)
	defer Pdr_ManStop(second)
	if vec.Vec_PtrEntry(second.vCexes, 0) != nil || vec.Vec_IntEntry(pars.vOutMap, 0) != -2 || second.pTime4Outs[0] == 0 || pars.nFailOuts != 1 {
		t.Fatalf("restart state: witness=%v map=%d budget=%d failed=%d", vec.Vec_PtrEntry(second.vCexes, 0), vec.Vec_IntEntry(pars.vOutMap, 0), second.pTime4Outs[0], pars.nFailOuts)
	}
}

func TestIncrementalFreshManagerRecountsTrivialOutputsAsInC(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const1())
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.fSolveAll, pars.fNotVerbose, pars.fSilent = 1, 1, 1
	for run := 0; run < 2; run++ {
		manager := Pdr_ManStart(g, &pars, nil)
		if got := IPdr_ManSolveInt(manager, 0, 0); got != 0 {
			t.Fatalf("run %d: got result %d", run, got)
		}
		Pdr_ManStop(manager)
	}
	if pars.nProveOuts != 2 || pars.nFailOuts != 2 {
		t.Fatalf("source counters after fresh managers: proved=%d failed=%d", pars.nProveOuts, pars.nFailOuts)
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
