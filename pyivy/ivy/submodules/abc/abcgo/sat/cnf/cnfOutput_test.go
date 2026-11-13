package cnf

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestOutputClausesUseOutputVariables(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	o1, o2 := g.AddOutput(x), g.AddOutput(y)
	data, err := EncodeAIG(g, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		add  func(*bsat.Solver, *Cnf_Dat_t) int32
		want func(bool, bool) bool
	}{
		{"or", Cnf_DataWriteOrClause, func(a, b bool) bool { return a || b }},
		{"or2", Cnf_DataWriteOrClause2, func(a, b bool) bool { return a || b }},
		{"and", Cnf_DataWriteAndClauses, func(a, b bool) bool { return a && b }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := bsat.NewSolver()
			if tc.add(s, data) != 1 {
				t.Fatal("could not add output constraint")
			}
			for mask := 0; mask < 4; mask++ {
				assumptions := []bsat.Literal{bsat.Positive(data.VarNums[o1.NodeID()]), bsat.Positive(data.VarNums[o2.NodeID()])}
				for i := range assumptions {
					if mask&(1<<i) == 0 {
						assumptions[i] ^= 1
					}
				}
				got := s.Solve(assumptions...) == bsat.Satisfiable
				if want := tc.want(mask&1 != 0, mask&2 != 0); got != want {
					t.Fatalf("mask=%d got=%v want=%v", mask, got, want)
				}
			}
		})
	}
}

func TestTransformPolarityUsesNodePhases(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	gate := g.AND(x.Not(), y.Not())
	out := g.AddOutput(gate)
	for _, transformOutputs := range []int32{0, 1} {
		data, err := EncodeAIG(g, 0)
		if err != nil {
			t.Fatal(err)
		}
		gateLit, outLit, inputLit := data.VarNums[gate.NodeID()]<<1, data.VarNums[out.NodeID()]<<1, data.VarNums[x.NodeID()]<<1
		data.Clauses = [][]int32{{gateLit, outLit, inputLit}}
		Cnf_DataTranformPolarity(data, transformOutputs)
		want := [][]int32{{gateLit ^ 1, outLit ^ transformOutputs, inputLit}}
		if !reflect.DeepEqual(data.Clauses, want) {
			t.Fatalf("outputs=%d got %v want %v", transformOutputs, data.Clauses, want)
		}
	}
}

func TestSolverLoadingFramesAndInitialization(t *testing.T) {
	g := aig.NewGraph()
	lo := g.NewInput()
	g.AddOutput(lo.Not())
	aig.Aig_ManSetRegNum(g, 1)
	data, err := EncodeAIG(g, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, frames := range []int32{1, 2} {
		s := Cnf_DataWriteIntoSolver(data, frames, 1)
		if s == nil {
			t.Fatalf("frames=%d: failed to load", frames)
		}
		if s.Solve(bsat.Positive(data.VarNums[lo.NodeID()])) != bsat.Unsatisfiable {
			t.Fatalf("frames=%d: initial latch is not constrained to zero", frames)
		}
		if frames == 2 {
			next := data.VarNums[lo.NodeID()] + int32(data.Variables)
			if s.Solve(bsat.Positive(next)) != bsat.Satisfiable || s.Solve(bsat.Negative(next)) != bsat.Unsatisfiable {
				t.Fatal("second frame does not follow the inverted latch input")
			}
		}
	}
	combinational := aig.NewGraph()
	x := combinational.NewInput()
	combinational.AddOutput(x)
	plain, err := EncodeAIG(combinational, 0)
	if err != nil {
		t.Fatal(err)
	}
	if Cnf_DataWriteIntoSolver(plain, 2, 0) == nil {
		t.Fatal("combinational frames rejected")
	}
}
