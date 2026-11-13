package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestTimeframeSimplify(t *testing.T) {
	for _, init := range []int32{0, 1} {
		g := aig.NewGraph()
		pi, lo := g.NewInput(), g.NewInput()
		g.AddOutput(lo)
		g.AddOutput(g.XOR(pi, lo))
		aig.Aig_ManSetRegNum(g, 1)
		q := Saig_ManTimeframeSimplify(g, 2, 3, init, 0, func(frames *aig.Graph, limit, prove int32) *aig.Graph {
			if limit != 1000 || prove != 0 {
				t.Fatalf("unexpected FRA options %d/%d", limit, prove)
			}
			aig.Aig_ManReprStart(frames, int32(len(frames.Nodes())))
			return frames
		})
		if q == nil {
			t.Fatal("missing result")
		}
		if err := q.Check(); err != nil {
			t.Fatal(err)
		}
		if Saig_ManHasMap1(g) != 0 || Saig_ManHasMap2(g) != 0 {
			t.Fatal("mapping not cleaned up")
		}
		for bits := 0; bits < 1<<q.InputCount(); bits++ {
			in := make([]bool, q.InputCount())
			for i := range in {
				in[i] = bits&(1<<i) != 0
			}
			out, err := q.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			state := false
			if init == 0 {
				state = in[3]
			}
			for f := 0; f < 3; f++ {
				if out[f] != state {
					t.Fatalf("init=%d input=%v output=%v", init, in, out)
				}
				state = state != in[f]
			}
			if init == 0 && out[3] != state {
				t.Fatal("incorrect final state")
			}
		}
	}
}

func TestTimeframeSimplifyRejectsMissingRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state)
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	defer func() {
		if recover() == nil {
			t.Fatal("missing source FRA representative precondition")
		}
	}()
	Saig_ManTimeframeSimplifyWithFraig(g, 1, 1, 0, 0, func(frames *aig.Graph, _, _ int32) *aig.Graph { return frames })
}

func TestTimeframeSimplifyFrameBoundAfterFRA(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state)
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	called := false
	defer func() {
		if recover() == nil || !called {
			t.Fatal("frame bound must be checked after FRA mapping")
		}
	}()
	Saig_ManTimeframeSimplifyWithFraig(g, 2, 1, 0, 0, func(frames *aig.Graph, _, _ int32) *aig.Graph {
		called = true
		aig.Aig_ManReprStart(frames, int32(len(frames.Nodes())))
		return frames
	})
}

func TestTimeframeSimplifyRejectsNoRegisters(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.NewInput())
	defer func() {
		if recover() == nil {
			t.Fatal("expected source register assertion")
		}
	}()
	Saig_ManTimeframeSimplifyWithFraig(g, 1, 1, 0, 0, func(g *aig.Graph, _, _ int32) *aig.Graph { return g })
}
