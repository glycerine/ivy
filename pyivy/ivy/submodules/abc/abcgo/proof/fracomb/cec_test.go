package fracomb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestCecDriverTrivialProofAndCounterexample(t *testing.T) {
	for _, one := range []bool{false, true} {
		g := aig.NewGraph()
		g.NewInput()
		if one {
			g.AddOutput(g.Const1())
		} else {
			g.AddOutput(g.Const0())
		}
		calls := 0
		sweep := func(g *aig.Graph, _, _ int32) *aig.Graph { calls++; return g }
		status := Fra_FraigCec(&g, 10000, 0, sweep)
		want := int32(1)
		if one {
			want = 0
		}
		if status != want {
			t.Fatalf("constant %t status %d", one, status)
		}
		if calls != 0 {
			t.Fatal("unneeded FRAIG sweep")
		}
		if one {
			model, ok := g.Data.([]int32)
			if !ok || len(model) != 1 {
				t.Fatalf("bad counterexample %v", g.Data)
			}
			out, err := g.Evaluate([]bool{model[0] != 0})
			if err != nil || !out[0] {
				t.Fatalf("invalid model %v", model)
			}
		}
	}
}

func TestCecDirectInputUsesSourceZeroWitness(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	if got := Fra_FraigCec(&g, 1000, 0, nil); got != 0 {
		t.Fatalf("status %d", got)
	}
	model, ok := g.Data.([]int32)
	if !ok || len(model) != 1 || model[0] != 0 {
		t.Fatalf("source zero witness %v", g.Data)
	}
}
func TestCecReplacesExistingWitnessWithSourceZeroWitness(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	g.Data = []int32{1}
	if Fra_FraigCec(&g, 1000, 0, nil) != 0 {
		t.Fatal("lost known failing model")
	}
	model, ok := g.Data.([]int32)
	if !ok || len(model) != 1 || model[0] != 0 {
		t.Fatalf("source zero witness %v", g.Data)
	}
}

func TestCountXorsRecognizesNativeAndStructuralNodes(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.NewXor(x, y)
	g.XOR(x, y)
	if got := Aig_ManCountXors(g); got != 2 {
		t.Fatalf("got %d XORs, want 2", got)
	}
}
func TestCecReplacesInvalidExistingWitness(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	g.Data = []int32{0}
	if got := Fra_FraigCec(&g, 1000, 0, nil); got != 0 {
		t.Fatalf("status %d", got)
	}
	model, ok := g.Data.([]int32)
	if !ok || len(model) != 1 || model[0] != 0 {
		t.Fatalf("source zero witness %v", g.Data)
	}
}

func cecPigeonholeMiter() *aig.Graph {
	const pigeons, holes = 8, 7
	g := aig.NewGraph()
	var inputs [pigeons][holes]aig.Edge
	for i := range inputs {
		for j := range inputs[i] {
			inputs[i][j] = g.NewInput()
		}
	}
	formula := g.Const1()
	for i := range inputs {
		clause := g.Const0()
		for _, v := range inputs[i] {
			clause = g.OR(clause, v)
		}
		formula = g.AND(formula, clause)
	}
	for h := 0; h < holes; h++ {
		for i := 0; i < pigeons; i++ {
			for j := 0; j < i; j++ {
				formula = g.AND(formula, g.OR(inputs[i][h].Not(), inputs[j][h].Not()))
			}
		}
	}
	g.AddOutput(formula)
	return g
}

func TestCecSourceDoesNotRevalidateSweepModel(t *testing.T) {
	// The initial 600-conflict SAT call cannot solve this pigeonhole miter.
	// A sweep callback supplies a deliberately invalid model to characterize
	// ABC's handling of a later status-0 result.
	g := cecPigeonholeMiter()
	calls := 0
	status := Fra_FraigCec(&g, 100000, 0, func(graph *aig.Graph, _, _ int32) *aig.Graph {
		calls++
		graph.Data = make([]int32, graph.InputCount())
		return graph
	})
	if calls == 0 || status != 0 {
		t.Fatalf("sweep calls=%d status=%d", calls, status)
	}
	model, ok := g.Data.([]int32)
	if !ok || len(model) != g.InputCount() {
		t.Fatalf("source model %v", g.Data)
	}
	outputs, err := g.Evaluate(make([]bool, g.InputCount()))
	if err != nil || outputs[0] {
		t.Fatalf("fixture unexpectedly validates: outputs=%v err=%v", outputs, err)
	}
}

func TestCecSourceSweepLimitsAndFinalSAT(t *testing.T) {
	for _, tc := range []struct{ limit, want int32 }{{1, -1}, {100000, 1}} {
		g := cecPigeonholeMiter()
		calls := 0
		status := Fra_FraigCec(&g, tc.limit, 0, func(graph *aig.Graph, nodeLimit, miterLimit int32) *aig.Graph {
			wantNode := int32(2) << uint(3*calls)
			wantMiter := int32(300) << uint(calls)
			if nodeLimit != wantNode || miterLimit != wantMiter {
				t.Fatalf("sweep %d limits=(%d,%d), want (%d,%d)", calls, nodeLimit, miterLimit, wantNode, wantMiter)
			}
			calls++
			return graph
		})
		if calls != 6 || status != tc.want {
			t.Fatalf("limit=%d calls=%d status=%d want=%d", tc.limit, calls, status, tc.want)
		}
	}
}

func TestMiterStatusAndAssertedOutput(t *testing.T) {
	cases := []struct {
		name             string
		output           func(*aig.Graph) aig.Edge
		status, asserted int32
	}{
		{"zero", func(g *aig.Graph) aig.Edge { return g.Const0() }, 1, -1},
		{"one", func(g *aig.Graph) aig.Edge { return g.Const1() }, 0, 0},
		{"input", func(g *aig.Graph) aig.Edge { return g.NewInput() }, 0, -1},
		{"inverted input", func(g *aig.Graph) aig.Edge { return g.NewInput().Not() }, 0, 0},
		{"unknown", func(g *aig.Graph) aig.Edge { return g.AND(g.NewInput(), g.NewInput()) }, -1, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			g.AddOutput(tc.output(g))
			if got := Fra_FraigMiterStatus(g); got != tc.status {
				t.Fatalf("status=%d", got)
			}
			if got := Fra_FraigMiterAssertedOutput(g); got != tc.asserted {
				t.Fatalf("asserted=%d", got)
			}
		})
	}
}

func TestCecSourceShortcutPrecedesCombinationalCheck(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(g.Const1())
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	if got := Fra_FraigCec(&g, 1000, 0, nil); got != 0 {
		t.Fatalf("source shortcut status %d", got)
	}
}
