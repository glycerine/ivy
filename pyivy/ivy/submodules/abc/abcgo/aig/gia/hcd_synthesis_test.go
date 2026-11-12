package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestHcdChoiceSynthesis(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.AND(a, b), c))
	snapshots := Hcd_ChoiceSynthesis(aig.Aig_ManDupDfs(g), 1, 1, 0, 0)
	if Vec_PtrSize(snapshots) != 3 {
		t.Fatal("missing synthesis snapshots")
	}
	for i := 0; i < Vec_PtrSize(snapshots); i++ {
		p := Vec_PtrEntry(snapshots, i).(*Gia_Man_t)
		for bits := 0; bits < 8; bits++ {
			got, err := p.Evaluate([]bool{bits&1 != 0, bits&2 != 0, bits&4 != 0})
			if err != nil || got[0] != (bits == 7) {
				t.Fatalf("snapshot=%d bits=%d got=%v err=%v", i, bits, got, err)
			}
		}
	}
}

func TestHcdComputeChoices(t *testing.T) {
	for _, synthesis := range []int32{0, 1} {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(g.AND(a, b), c))
		q := Hcd_ComputeChoices(g, 1000, synthesis, 0, 0)
		for bits := 0; bits < 8; bits++ {
			got, err := q.Evaluate([]bool{bits&1 != 0, bits&2 != 0, bits&4 != 0})
			if err != nil || len(got) != 1 || got[0] != (bits == 7) {
				t.Fatalf("synthesis=%d bits=%d got=%v err=%v", synthesis, bits, got, err)
			}
		}
	}
}
