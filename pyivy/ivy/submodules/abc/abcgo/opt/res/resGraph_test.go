package res

import (
	"context"
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestResubstituteGraphFindsLowerCostEquivalentDivisors(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	ab := g.AND(a, b)
	ac := g.AND(a, c)
	root := g.AND(ab, ac)
	bc := g.AND(b, c) // Existing divisor exposes a lower-cost factorization.
	g.AddOutput(root)
	g.AddOutput(bc)
	before := graphTruthOverAllInputs(t, g)
	beforeAnds := countGraphAnds(g)

	stats, err := ResubstituteGraph(g, GraphParams{MaxSupport: 4, MaxDivisors: 32})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Replacements == 0 {
		t.Fatalf("expected a profitable resubstitution, stats=%+v", stats)
	}
	if got := countGraphAnds(g); got >= beforeAnds {
		t.Fatalf("AND count did not decrease: before=%d after=%d, stats=%+v", beforeAnds, got, stats)
	}
	after := graphTruthOverAllInputs(t, g)
	if !equalTruthOutputs(before, after) {
		t.Fatal("resubstitution changed the output truth table")
	}
}

func TestResubstituteGraphPreservesRandomNetworkTruth(t *testing.T) {
	rng := rand.New(rand.NewSource(0xabc))
	for trial := 0; trial < 80; trial++ {
		g := aig.NewGraph()
		inputs := make([]aig.Edge, 4)
		for i := range inputs {
			inputs[i] = g.NewInput()
		}
		pool := append([]aig.Edge(nil), inputs...)
		for i := 0; i < 20; i++ {
			a := pool[rng.Intn(len(pool))]
			b := pool[rng.Intn(len(pool))]
			if rng.Intn(2) != 0 {
				a = a.Not()
			}
			if rng.Intn(2) != 0 {
				b = b.Not()
			}
			pool = append(pool, g.AND(a, b))
		}
		for i := 0; i < 3; i++ {
			g.AddOutput(pool[rng.Intn(len(pool))])
		}
		before := graphTruthOverAllInputs(t, g)
		beforeAnds := countGraphAnds(g)
		if _, err := ResubstituteGraph(g, GraphParams{MaxSupport: 4, MaxDivisors: 20}); err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		after := graphTruthOverAllInputs(t, g)
		if !equalTruthOutputs(before, after) {
			t.Fatalf("trial %d changed output truth", trial)
		}
		if got := countGraphAnds(g); got > beforeAnds {
			t.Fatalf("trial %d increased AND count from %d to %d", trial, beforeAnds, got)
		}
	}
}

func TestResubstituteGraphUsesLocalCutForWidePrimarySupport(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 8)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	ab := g.AND(inputs[0], inputs[1])
	cd := g.AND(inputs[2], inputs[3])
	ef := g.AND(inputs[4], inputs[5])
	gh := g.AND(inputs[6], inputs[7])
	left := g.AND(ab, cd)
	right := g.AND(ef, gh)
	root := g.AND(left, right)
	altLeft := g.AND(ab, ef)
	altRight := g.AND(cd, gh)
	alternate := g.AND(altLeft, altRight)
	g.AddOutput(root)
	g.AddOutput(alternate)
	before := graphTruthOverAllInputs(t, g)

	stats, err := ResubstituteGraph(g, GraphParams{MaxSupport: 6, MaxDivisors: 64})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Replacements == 0 {
		t.Fatalf("expected cut-local replacement for 8-input cone, stats=%+v", stats)
	}
	if after := graphTruthOverAllInputs(t, g); !equalTruthOutputs(before, after) {
		t.Fatal("wide-support cut replacement changed graph outputs")
	}
}

func TestResubstituteGraphRejectsInvalidBoundsAndCancellation(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	if _, err := ResubstituteGraph(g, GraphParams{MaxSupport: 7}); err == nil {
		t.Fatal("expected unsupported cut bound error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before := graphTruthOverAllInputs(t, g)
	if _, err := ResubstituteGraph(g, GraphParams{Context: ctx}); err == nil {
		t.Fatal("expected cancellation")
	}
	if after := graphTruthOverAllInputs(t, g); !equalTruthOutputs(before, after) {
		t.Fatal("cancelled call changed graph")
	}
}

func graphTruthOverAllInputs(t *testing.T, g *aig.Graph) [][]bool {
	t.Helper()
	truth := make([][]bool, 1<<g.InputCount())
	for assignment := range truth {
		inputs := make([]bool, g.InputCount())
		for i := range inputs {
			inputs[i] = assignment&(1<<i) != 0
		}
		outputs, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatalf("evaluate %04b: %v", assignment, err)
		}
		truth[assignment] = outputs
	}
	return truth
}

func equalTruthOutputs(a, b [][]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
