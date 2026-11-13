package dch

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestSweepKeepsCandidatesConnectedByExistingChoice(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	head := g.AND(g.AND(a, b), c)
	member := g.AND(a, g.AND(b, c))
	g.AddOutput(head)
	g.AddOutput(member)
	aig.Aig_ObjSetEquiv(g, head, member)
	if aig.Aig_ObjCheckTfi(g, head, member) == 0 {
		t.Fatal("fixture must exercise existing choice link")
	}
	m, err := newManager(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.computeEquivalences(); err != nil {
		t.Fatal(err)
	}
	if got := m.proven[member.NodeID()]; got != head {
		t.Fatalf("choice-linked equivalent node: representative %v, want %v", got, head)
	}
}

func TestCNFMuxOverlappingControlsAndBranches(t *testing.T) {
	for _, flip := range []bool{false, true} {
		for _, form := range []string{"control-then", "control-else", "same", "complementary"} {
			g := aig.NewGraph()
			c, a, b := g.NewInput(), g.NewInput(), g.NewInput()
			yes, no := a, b
			switch form {
			case "control-then":
				yes = c
			case "control-else":
				no = c
			case "same":
				no = a
			case "complementary":
				no = a.Not()
			}
			out := g.Mux(c, yes, no)
			g.AddOutput(out)
			pars := DefaultParams()
			pars.PolarFlip = flip
			m, err := newManager(g, &pars)
			if err != nil {
				t.Fatal(err)
			}
			m.recycleSolver()
			lit, err := m.satLiteral(out)
			if err != nil {
				t.Fatal(err)
			}
			for bits := 0; bits < 8; bits++ {
				values := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
				want, err := g.Evaluate(values)
				if err != nil {
					t.Fatal(err)
				}
				assumptions := []int32{}
				for i, in := range g.Inputs() {
					if _, used := m.variables[in.NodeID()]; !used {
						continue
					}
					v, err := m.satLiteral(in)
					if err != nil {
						t.Fatal(err)
					}
					if !values[i] {
						v ^= 1
					}
					assumptions = append(assumptions, v)
				}
				for value := 0; value < 2; value++ {
					target := lit
					if value == 0 {
						target ^= 1
					}
					status := m.solver.Solve(append(append([]int32(nil), assumptions...), target)...)
					if (status == bsat.Satisfiable) != ((value != 0) == want[0]) {
						t.Fatalf("flip=%t form=%s bits=%d output=%d status=%v", flip, form, bits, value, status)
					}
				}
			}
		}
	}
}

func TestRecycleAfterLearnedEquivalence(t *testing.T) {
	for _, flip := range []bool{false, true} {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		x := g.AND(g.AND(a, b), c)
		y := g.AND(a, g.AND(b, c))
		different := g.AND(a, b)
		for _, e := range []aig.Edge{x, y, different} {
			g.AddOutput(e)
		}
		pars := DefaultParams()
		pars.PolarFlip = flip
		pars.SatVarMax = 1
		pars.CallsRecycle = 0
		m, err := newManager(g, &pars)
		if err != nil {
			t.Fatal(err)
		}
		status, _, err := m.provePair(x, y)
		if err != nil || status != proofEquivalent {
			t.Fatalf("flip=%t proof=%v err=%v", flip, status, err)
		}
		old, recycles := m.solver, m.stats.Recycles
		status, model, err := m.provePair(x, different)
		if err != nil || status != proofDifferent || m.solver == old || m.stats.Recycles != recycles+1 {
			t.Fatalf("flip=%t post-recycle proof=%v err=%v recycles=%d", flip, status, err, m.stats.Recycles)
		}
		values, err := g.Evaluate(model)
		if err != nil || values[0] == values[2] {
			t.Fatalf("flip=%t invalid counterexample %v outputs=%v err=%v", flip, model, values, err)
		}
	}
}
