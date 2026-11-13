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

func TestCecDirectInputProducesValidWitness(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	if got := Fra_FraigCec(&g, 1000, 0, nil); got != 0 {
		t.Fatalf("status %d", got)
	}
	model, ok := g.Data.([]int32)
	if !ok || len(model) != 1 || model[0] != 1 {
		t.Fatalf("invalid witness %v", g.Data)
	}
}
func TestCecPreservesExistingWitness(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	g.Data = []int32{1}
	if Fra_FraigCec(&g, 1000, 0, nil) != 0 {
		t.Fatal("lost known failing model")
	}
	model, ok := g.Data.([]int32)
	if !ok || len(model) != 1 || model[0] != 1 {
		t.Fatalf("invalid witness %v", g.Data)
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
func TestCecRejectsInvalidExistingWitness(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	g.Data = []int32{0}
	if got := Fra_FraigCec(&g, 1000, 0, nil); got != 0 {
		t.Fatalf("status %d", got)
	}
	model, ok := g.Data.([]int32)
	if !ok || len(model) != 1 || model[0] != 1 {
		t.Fatalf("invalid witness %v", g.Data)
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

func TestCecRejectsSequentialGraphs(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(g.Const1())
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	defer func() {
		if recover() == nil {
			t.Fatal("accepted a sequential miter")
		}
	}()
	Fra_FraigCec(&g, 1000, 0, nil)
}
