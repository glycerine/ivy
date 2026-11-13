package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestOneHotComputeAndConstraints(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(b)
	aig.Aig_ManSetRegNum(g, 2)
	sim := newFraSimulation(g, 0, 1, 1)
	sim.NodeWords(a.NodeID())[0] = 0x55555555
	sim.NodeWords(b.NodeID())[0] = 0xaaaaaaaa
	p := &Fra_Man_t{PManAig: g, PManFraig: g, PPars: &Fra_Par_t{NFramesK: 1}, PSml: sim, PSat: bsat.NewSolver(), PMemSatNums: make([]int32, len(g.Nodes()))}
	clauses := Fra_OneHotCompute(p, sim)
	if Fra_OneHotCount(p, clauses) != 1 {
		t.Fatal("expected one candidate")
	}
	exdc := Fra_OneHotCreateExdc(p, clauses)
	for mask := 0; mask < 4; mask++ {
		values := []bool{mask&1 != 0, mask&2 != 0}
		got, err := exdc.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		if got[0] != (mask != 3) {
			t.Fatalf("mask %d output %v", mask, got)
		}
	}
	p.PMemSatNums[a.NodeID()] = 1
	p.PMemSatNums[b.NodeID()] = 2
	Fra_OneHotAssume(p, clauses)
	if p.PSat.Solve(bsat.Positive(1), bsat.Positive(2)) != bsat.Unsatisfiable {
		t.Fatal("one-hot constraint not enforced")
	}
	if p.PSat.Solve(bsat.Positive(1), bsat.Negative(2)) != bsat.Satisfiable {
		t.Fatal("valid one-hot assignment rejected")
	}
}

// The C routine tests three polarity combinations in a fixed order and keeps
// at most one clause for each nonconstant, nonequivalent register pair.
func TestOneHotCandidatePolarities(t *testing.T) {
	for _, tc := range []struct {
		name         string
		a, b         uint32
		signA, signB int32
	}{
		{"not a or not b", 0x33333333, 0x44444444, 1, 1},
		{"a or not b", 0x77777777, 0x11111111, 0, 1},
		{"not a or b", 0x11111111, 0x77777777, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			g.AddOutput(a)
			g.AddOutput(b)
			aig.Aig_ManSetRegNum(g, 2)
			sim := newFraSimulation(g, 0, 1, 1)
			sim.NodeWords(a.NodeID())[0], sim.NodeWords(b.NodeID())[0] = tc.a, tc.b
			p := &Fra_Man_t{PManAig: g, PManFraig: g, PPars: &Fra_Par_t{NFramesK: 1}, PSml: sim, PSat: bsat.NewSolver(), PMemSatNums: make([]int32, len(g.Nodes()))}
			clauses := Fra_OneHotCompute(p, sim)
			if vec.Vec_IntSize(clauses) != 2 || vec.Vec_IntEntry(clauses, 0) != Fra_RegToLit(0, tc.signA) || vec.Vec_IntEntry(clauses, 1) != Fra_RegToLit(1, tc.signB) {
				t.Fatalf("unexpected candidates: %v", vec.Vec_IntArray(clauses))
			}
			p.PMemSatNums[a.NodeID()], p.PMemSatNums[b.NodeID()] = 1, 2
			Fra_OneHotAssume(p, clauses)
			exdc := Fra_OneHotCreateExdc(p, clauses)
			for mask := 0; mask < 4; mask++ {
				av, bv := mask&1 != 0, mask&2 != 0
				want := (av != (tc.signA != 0)) || (bv != (tc.signB != 0))
				got, err := exdc.Evaluate([]bool{av, bv})
				if err != nil || len(got) != 1 || got[0] != want {
					t.Fatalf("mask=%d output=%v error=%v want=%v", mask, got, err, want)
				}
				la, lb := bsat.Positive(1), bsat.Positive(2)
				if !av {
					la = bsat.Negative(1)
				}
				if !bv {
					lb = bsat.Negative(2)
				}
				if (p.PSat.Solve(la, lb) == bsat.Satisfiable) != want {
					t.Fatalf("mask=%d SAT differs from clause", mask)
				}
			}
		})
	}
}
