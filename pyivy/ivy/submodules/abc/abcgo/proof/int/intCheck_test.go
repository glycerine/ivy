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

// Adding the complementary interpolant makes the union of learned state sets
// the entire state space, which is inductive even for a toggling latch.
func TestIncrementalInductionCombinesInterpolants(t *testing.T) {
	trans := aig.NewGraph()
	state := trans.NewInput()
	trans.AddOutput(state.Not())
	aig.Aig_ManSetRegNum(trans, 1)
	check := Inter_CheckStart(trans, 1)
	defer Inter_CheckStop(check)
	for i, invert := range []bool{false, true} {
		inter := aig.NewGraph()
		x := inter.NewInput()
		if invert {
			x = x.Not()
		}
		inter.AddOutput(x)
		data := cnf.Cnf_Derive(inter, 1)
		got := Inter_CheckPerform(check, data, 0)
		if got != int32(i) {
			t.Fatalf("interpolant %d: got %d want %d", i, got, i)
		}
	}
}

// Compare the k-frame SAT query in intCheck.c against explicit enumeration of
// every initial state. The predicates include constants and non-cubes.
func TestInductionMultipleFramesAgainstEnumeration(t *testing.T) {
	truth := func(g *aig.Graph, x, y aig.Edge, table int) aig.Edge {
		r := g.Const0()
		for s := 0; s < 4; s++ {
			if table&(1<<s) == 0 {
				continue
			}
			a, b := x, y
			if s&1 == 0 {
				a = a.Not()
			}
			if s&2 == 0 {
				b = b.Not()
			}
			r = g.OR(r, g.AND(a, b))
		}
		return r
	}
	for _, next := range [][2]int{{10, 12}, {5, 12}, {6, 10}, {0, 15}} {
		for k := int32(2); k <= 3; k++ {
			for predicate := 0; predicate < 16; predicate++ {
				trans := aig.NewGraph()
				x, y := trans.NewInput(), trans.NewInput()
				trans.AddOutput(truth(trans, x, y, next[0]))
				trans.AddOutput(truth(trans, x, y, next[1]))
				aig.Aig_ManSetRegNum(trans, 2)
				inter := aig.NewGraph()
				a, b := inter.NewInput(), inter.NewInput()
				inter.AddOutput(truth(inter, a, b, predicate))
				check := Inter_CheckStart(trans, k)
				data := cnf.Cnf_Derive(inter, 1)
				want := int32(1)
				for initial := 0; initial < 4; initial++ {
					s, prefix := initial, true
					for frame := int32(0); frame < k; frame++ {
						prefix = prefix && predicate&(1<<s) != 0
						s = (next[0]>>s)&1 | ((next[1]>>s)&1)<<1
					}
					if prefix && predicate&(1<<s) == 0 {
						want = 0
					}
				}
				got := Inter_CheckPerform(check, data, 0)
				Inter_CheckStop(check)
				if got != want {
					t.Fatalf("next=%v k=%d predicate=%04b got=%d want=%d", next, k, predicate, got, want)
				}
			}
		}
	}
}
