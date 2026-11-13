package int

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/cnf"
)

func TestUnrollFrameOutputs(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(x)
	g.AddOutput(g.XOR(x, state))
	aig.Aig_ManSetRegNum(g, 1)
	frames := Inter_ManUnrollFrames(g, 2)
	for mask := 0; mask < 8; mask++ {
		initial, x0, x1 := mask&1 != 0, mask&2 != 0, mask&4 != 0
		got, err := frames.Evaluate([]bool{initial, x0, x1})
		want := []bool{initial != x0, (initial != x0) != x1}
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("mask %d: %v, %v; want %v", mask, got, err, want)
		}
	}
}

func TestIncrementalInduction(t *testing.T) {
	for _, toggle := range []bool{false, true} {
		trans := aig.NewGraph()
		state := trans.NewInput()
		if toggle {
			state = state.Not()
		}
		trans.AddOutput(state)
		aig.Aig_ManSetRegNum(trans, 1)
		check := Inter_CheckStart(trans, 1)
		if check == nil {
			t.Fatal("failed to start checker")
		}
		inter := aig.NewGraph()
		inter.AddOutput(inter.NewInput())
		data := cnf.Cnf_Derive(inter, 1)
		vars := append([]int32(nil), data.VarNums...)
		clauses := make([][]int32, len(data.Clauses))
		for i, clause := range data.Clauses {
			clauses[i] = append([]int32(nil), clause...)
		}
		want := int32(1)
		if toggle {
			want = 0
		}
		for attempt := 0; attempt < 2; attempt++ {
			if got := Inter_CheckPerform(check, data, 0); got != want {
				t.Fatalf("toggle=%v attempt=%d: got %d want %d", toggle, attempt, got, want)
			}
			if !reflect.DeepEqual(vars, data.VarNums) || !reflect.DeepEqual(clauses, data.Clauses) {
				t.Fatal("interpolant CNF was not restored")
			}
		}
		Inter_CheckStop(check)
	}
}
