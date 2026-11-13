package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"math/rand"
	"testing"
)

func TestBmcAgainstExhaustiveSequentialTraces(t *testing.T) {
	rng := rand.New(rand.NewSource(10389))
	for trial := 0; trial < 80; trial++ {
		g := aig.NewGraph()
		pi, lo := g.NewInput(), g.NewInput()
		nodes := []aig.Edge{pi, lo, g.Const1(), g.Const0()}
		for i := 0; i < 8; i++ {
			a, b := nodes[rng.Intn(len(nodes))], nodes[rng.Intn(len(nodes))]
			if rng.Intn(2) == 0 {
				a = a.Not()
			}
			if rng.Intn(2) == 0 {
				b = b.Not()
			}
			nodes = append(nodes, g.AND(a, b))
		}
		output, next := nodes[4+rng.Intn(len(nodes)-4)], nodes[4+rng.Intn(len(nodes)-4)]
		if rng.Intn(2) == 0 {
			output = output.Not()
		}
		if rng.Intn(2) == 0 {
			next = next.Not()
		}
		g.AddOutput(output)
		g.AddOutput(next)
		aig.Aig_ManSetRegNum(g, 1)
		reachable := false
		for sequence := 0; sequence < 16; sequence++ {
			state := false
			for frame := 0; frame < 4; frame++ {
				out, err := g.Evaluate([]bool{sequence&(1<<frame) != 0, state})
				if err != nil {
					t.Fatal(err)
				}
				if out[0] {
					reachable = true
				}
				state = out[1]
			}
		}
		var last int32
		got := Ssw_BmcDynamic(g, 4, 0, 0, &last)
		if reachable {
			if got != 0 {
				t.Fatalf("trial %d: reachable output returned %d", trial, got)
			}
			cex := g.SequentialCounterexample()
			if cex == nil || saig.Saig_ManVerifyCex(g, cex) == 0 {
				t.Fatalf("trial %d: invalid witness", trial)
			}
		} else if got != 1 {
			t.Fatalf("trial %d: bounded-safe output returned %d", trial, got)
		}
	}
}

func TestBmcHandlesRewiredNodeOrder(t *testing.T) {
	g := aig.NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	early, late := g.AND(a, q), g.AND(b, q)
	aig.Aig_ObjDisconnect(g, early)
	aig.Aig_ObjConnect(g, early, late, b)
	g.AddOutput(early)
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	if got := Ssw_BmcDynamic(g, 3, 0, 0, nil); got != 0 {
		t.Fatalf("reachable output returned %d", got)
	}
	if cex := g.SequentialCounterexample(); cex == nil || saig.Saig_ManVerifyCex(g, cex) == 0 {
		t.Fatal("missing or invalid BMC witness")
	}
}
