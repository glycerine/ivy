package core

import (
	"context"
	"errors"
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"testing"
)

func TestRewriteReleasesReverseLevelsOnCancellation(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := DefaultRewriteParams()
	p.UpdateLevel, p.Context = true, ctx
	if err := Dar_ManRewrite(g, &p); !errors.Is(err, context.Canceled) {
		t.Fatalf("rewrite error %v, want cancellation", err)
	}
	// Starting timing again panics if the rewrite leaked reverse-level state.
	if aig.Aig_ManHasFanout(g) {
		t.Fatal("rewrite leaked its fanout index")
	}
	aig.Aig_ManFanoutStart(g)
	defer aig.Aig_ManFanoutStop(g)
	aig.Aig_ManStartReverseLevels(g, 0)
	aig.Aig_ManStopReverseLevels(g)
}

// darCore.c uses each root's required level, allowing area-reducing rewrites
// to use timing slack as long as the graph's original maximum level is kept.
func TestRewriteUsesRequiredLevelSlack(t *testing.T) {
	rng := rand.New(rand.NewSource(606))
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 12)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	nodes := append([]aig.Edge(nil), inputs[:4]...)
	for i := 0; i < 12; i++ {
		a, b := nodes[rng.Intn(len(nodes))], nodes[rng.Intn(len(nodes))]
		if rng.Intn(2) == 0 {
			a = a.Not()
		}
		if rng.Intn(2) == 0 {
			b = b.Not()
		}
		nodes = append(nodes, g.AND(a, b))
	}
	g.AddOutput(nodes[len(nodes)-1])
	deep := inputs[4]
	for i := 5; i < len(inputs); i++ {
		deep = g.AND(deep, inputs[i])
	}
	g.AddOutput(deep)
	original := aig.Aig_ManDupDfs(g)
	initialLevel := aig.Aig_ManLevelNum(original)
	p := DefaultRewriteParams()
	p.UpdateLevel = true
	p.UseZeros = true
	if err := Dar_ManRewrite(g, &p); err != nil {
		t.Fatal(err)
	}
	compact := aig.Aig_ManDupDfs(g)
	if nodes := aig.Aig_ManNodeNum(compact); nodes > 12 {
		t.Fatalf("rewrite failed to use available slack: %d nodes, want at most 12", nodes)
	}
	if level := aig.Aig_ManLevelNum(compact); level > initialLevel {
		t.Fatalf("rewrite increased maximum level %d -> %d", initialLevel, level)
	}
	for bits := 0; bits < 4096; bits++ {
		in := make([]bool, 12)
		for i := range in {
			in[i] = bits&(1<<i) != 0
		}
		want, e := original.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		got, e := compact.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		for i := range want {
			if want[i] != got[i] {
				t.Fatalf("input %v got %v want %v", in, got, want)
			}
		}
	}
}
