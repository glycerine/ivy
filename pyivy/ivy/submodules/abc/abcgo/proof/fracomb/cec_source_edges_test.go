package fracomb

import (
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
)

// Fra_FraigMiterStatus only treats true PIs as immediately nonconstant and
// ignores register-input COs. ABC's Aig_ManSetRegNum establishes that split.
func TestSourceMiterStatusSequentialInterface(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	ro := g.NewInput()
	g.AddOutput(ro)
	g.AddOutput(g.Const1()) // next state, not a property output
	aig.Aig_ManSetRegNum(g, 1)
	if got := Fra_FraigMiterStatus(g); got != -1 {
		t.Fatalf("register output status=%d", got)
	}
	if got := Fra_FraigMiterAssertedOutput(g); got != -1 {
		t.Fatalf("register output asserted=%d", got)
	}
	// A fresh graph is needed because POs precede RIs in ABC's interface.
	h := aig.NewGraph()
	input, state := h.NewInput(), h.NewInput()
	h.AddOutput(input)
	h.AddOutput(state)
	aig.Aig_ManSetRegNum(h, 1)
	if got := Fra_FraigMiterStatus(h); got != 0 {
		t.Fatalf("true PI status=%d", got)
	}
}

func TestSourceMiterStatusMixedOutputs(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(g.AND(x, y)) // undecided under ABC's zero-input phase
	g.AddOutput(x)           // nonconstant, but not asserted by zero input
	aig.Aig_ManSetRegNum(g, 0)
	if got := Fra_FraigMiterStatus(g); got != 0 {
		t.Fatalf("mixed status=%d", got)
	}
	if got := Fra_FraigMiterAssertedOutput(g); got != -1 {
		t.Fatalf("zero-pattern asserted output=%d", got)
	}
	g.AddOutput(g.Const1())
	if got := Fra_FraigMiterAssertedOutput(g); got != 3 {
		t.Fatalf("first asserted output=%d", got)
	}
}

func TestSourceCountXorsAllInputPolarities(t *testing.T) {
	for phase := 0; phase < 4; phase++ {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		if phase&1 != 0 {
			a = a.Not()
		}
		if phase&2 != 0 {
			b = b.Not()
		}
		g.AddOutput(g.XOR(a, b))
		if got := Aig_ManCountXors(g); got != 1 {
			t.Fatalf("phase=%d XOR count=%d", phase, got)
		}
	}
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.Mux(a, b, c))
	if got := Aig_ManCountXors(g); got != 0 {
		t.Fatalf("ordinary MUX count=%d", got)
	}
}

func TestSourceCecNoOutputsAsserts(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	defer func() {
		if recover() == nil {
			t.Fatal("C reaches empty OR clause assertion")
		}
	}()
	Fra_FraigCec(&g, 10, 0, nil)
}

// With an expired Time2Quit and an initially undecided SAT query, ABC's
// Dar_ManRwsat returns NULL and Fra_FraigCec dereferences it in the loop.
func TestSourceCecExpiredRewriteFails(t *testing.T) {
	g := cecPigeonholeMiter()
	g.Time2Quit = time.Now().Add(-time.Second)
	defer func() {
		if recover() == nil {
			t.Fatal("C dereferences the failed rewrite")
		}
	}()
	Fra_FraigCec(&g, 1000, 0, func(graph *aig.Graph, _, _ int32) *aig.Graph { return graph })
}
