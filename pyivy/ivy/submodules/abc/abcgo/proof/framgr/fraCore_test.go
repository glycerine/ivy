package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestFraMiterStatusAndAssertedOutput(t *testing.T) {
	cases := []struct {
		name             string
		makeOutput       func(*aig.Graph) aig.Edge
		status, asserted int32
	}{
		{"zero", func(g *aig.Graph) aig.Edge { return g.Const0() }, 1, -1},
		{"one", func(g *aig.Graph) aig.Edge { return g.Const1() }, 0, 0},
		// New managers start with nTruePis == 0 in ABC.
		{"input", func(g *aig.Graph) aig.Edge { return g.NewInput() }, -1, -1},
		{"unknown", func(g *aig.Graph) aig.Edge { return g.AND(g.NewInput(), g.NewInput()) }, -1, -1},
		{"phase one", func(g *aig.Graph) aig.Edge { return g.AND(g.NewInput().Not(), g.NewInput().Not()) }, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			g.AddOutput(tc.makeOutput(g))
			if got := Fra_FraigMiterStatus(g); got != tc.status {
				t.Fatalf("status %d want %d", got, tc.status)
			}
			if got := Fra_FraigMiterAssertedOutput(g); got != tc.asserted {
				t.Fatalf("asserted %d want %d", got, tc.asserted)
			}
			g.Data = []int32{1}
			if Fra_FraigMiterStatus(g) != 0 {
				t.Fatal("model did not mark miter as failed")
			}
		})
	}
}

func TestFraFraigMappedCallbackAndTruth(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	n := g.AND(a, b)
	g.AddOutput(n)
	called := false
	reduced, mapping := Fra_FraigEquivenceMapped(g, 100, 0, func(m *Fra_Man_t) {
		called = true
		if m.PManAig != g || m.PManFraig == nil || Fra_ObjFraig(m, n, 0) == aig.InvalidEdge {
			t.Fatal("callback ran without mapping")
		}
	})
	if !called || len(mapping) != len(g.Nodes()) || mapping[n.NodeID()] == aig.InvalidEdge {
		t.Fatalf("callback %v mapping %v", called, mapping)
	}
	for i := 0; i < 4; i++ {
		input := []bool{i&1 != 0, i&2 != 0}
		got, err := reduced.Evaluate(input)
		if err != nil || got[0] != (input[0] && input[1]) {
			t.Fatalf("input %v: output %v error %v", input, got, err)
		}
	}
}

func TestFraFraigMappedNoInternalNodes(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x.Not())
	called := false
	result, mapping := Fra_FraigEquivenceMapped(g, 100, 0, func(*Fra_Man_t) { called = true })
	if called {
		t.Fatal("ABC's no-node fast path returns before the callback")
	}
	if mapping[x.NodeID()] == aig.InvalidEdge {
		t.Fatal("input was not mapped")
	}
	for _, v := range []bool{false, true} {
		got, err := result.Evaluate([]bool{v})
		if err != nil || got[0] == v {
			t.Fatalf("input %v got %v error %v", v, got, err)
		}
	}
}

func TestFraFraigNativeXorFollowsSourceAnd(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	structural := g.XOR(x, y)
	explicit := g.NewXor(x, y)
	g.AddOutput(structural)
	g.AddOutput(explicit)
	reduced, mapping := Fra_FraigEquivenceMapped(g, 1000, 0, nil)
	if aig.Aig_Regular(mapping[structural.NodeID()]) == aig.Aig_Regular(mapping[explicit.NodeID()]) {
		t.Fatalf("native XOR unexpectedly merged with expanded XOR: %v", mapping)
	}
	for i := 0; i < 4; i++ {
		input := []bool{i&1 != 0, i&2 != 0}
		got, err := reduced.Evaluate(input)
		if err != nil || got[0] != (input[0] != input[1]) || got[1] != (input[0] && input[1]) {
			t.Fatalf("inputs %v outputs %v error %v", input, got, err)
		}
	}
}

func TestFraFraigChoicePreservesTruth(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.XOR(x, y))
	g.AddOutput(g.NewXor(x, y))
	got := Fra_FraigChoice(g, 1000, 0)
	for i := 0; i < 4; i++ {
		in := []bool{i&1 != 0, i&2 != 0}
		out, err := got.Evaluate(in)
		if err != nil || out[0] != (in[0] != in[1]) || out[1] != out[0] {
			t.Fatalf("inputs %v outputs %v error %v", in, out, err)
		}
	}
}

func TestFraFraigProveReturnsSimulationWitness(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, y))
	reduced, _ := Fra_FraigEquivenceMapped(g, 1000, 1, nil)
	model, ok := reduced.Data.([]int32)
	if !ok || len(model) != 3 {
		t.Fatalf("missing simulation model: %v", reduced.Data)
	}
	got, err := g.Evaluate([]bool{model[0] != 0, model[1] != 0})
	if err != nil || !got[0] {
		t.Fatalf("invalid witness %v, output %v, error %v", model, got, err)
	}
}
