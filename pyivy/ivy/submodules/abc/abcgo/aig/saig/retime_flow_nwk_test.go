package saig_test

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/nwk"
	"math/rand"
	"testing"
)

func TestRetimeMinCutMatchesNWK(t *testing.T) {
	r := rand.New(rand.NewSource(8327))
	for trial := 0; trial < 200; trial++ {
		g := aig.NewGraph()
		edges := make([]aig.Edge, 5)
		for i := range edges {
			edges[i] = g.NewInput()
		}
		for i := 0; i < 9; i++ {
			a, b := edges[r.Intn(len(edges))], edges[r.Intn(len(edges))]
			if r.Intn(2) == 0 {
				a = a.Not()
			}
			if r.Intn(2) == 0 {
				b = b.Not()
			}
			edges = append(edges, g.AND(a, b))
		}
		for i := 0; i < 5; i++ {
			g.AddOutput(edges[r.Intn(len(edges))])
		}
		aig.Aig_ManSetRegNum(g, 3)
		aig.Aig_ManCleanup(g)
		for _, forward := range []bool{false, true} {
			wantFlag := int32(0)
			if forward {
				wantFlag = 1
			}
			got := saig.Saig_ManDeriveRetimingCut(g, forward)
			want := nwk.Nwk_ManDeriveRetimingCut(g, wantFlag, 0)
			if len(got) != vec.Vec_PtrSize(want) {
				t.Fatalf("trial %d forward=%v got %v want %v nodes=%+v inputs=%v outputs=%v", trial, forward, got, vec.Vec_PtrArray(want), g.Nodes(), g.Inputs(), g.Outputs())
			}
		}
	}
}
