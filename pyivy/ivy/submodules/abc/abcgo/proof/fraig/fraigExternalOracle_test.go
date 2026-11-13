package fraig

import (
	"math/rand"
	"testing"
)

func TestExternalClausesAgainstExhaustiveTruthTables(t *testing.T) {
	r := rand.New(rand.NewSource(19017))
	for trial := 0; trial < 30; trial++ {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
		a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
		nodes := []*Fraig_Node_t{a, b, c, Fraig_NodeAnd(p, a, b), Fraig_NodeExor(p, b, c)}
		truth := []uint8{0xaa, 0xcc, 0xf0, 0x88, 0x3c}
		var clause []*Fraig_Node_t
		var allowed uint8
		for allowed == 0 {
			clause = nil
			for k := 0; k < 2+r.Intn(2); k++ {
				i := r.Intn(len(nodes))
				edge, mask := nodes[i], truth[i]
				if r.Intn(2) != 0 {
					edge = Fraig_Not(edge)
					mask = ^mask
				}
				clause = append(clause, edge)
				allowed |= mask
			}
		}
		Fraig_ManAddClause(p, clause, int32(len(clause)))
		for q := 0; q < 12; q++ {
			i, j := r.Intn(len(nodes)), r.Intn(len(nodes))
			if Fraig_Regular(nodes[i]) == Fraig_Regular(nodes[j]) {
				continue
			}
			x, y := nodes[i], nodes[j]
			mx, my := truth[i], truth[j]
			if r.Intn(2) != 0 {
				x = Fraig_Not(x)
				mx = ^mx
			}
			if r.Intn(2) != 0 {
				y = Fraig_Not(y)
				my = ^my
			}
			want := allowed & ^(mx|my) == 0
			got := Fraig_ManCheckClauseUsingSat(p, x, y, -1) != 0
			if got != want {
				t.Fatalf("trial %d query %d: SAT=%t want %t, allowed=%08b clause=%08b", trial, q, got, want, allowed, mx|my)
			}
		}
		Fraig_ManFree(p)
	}
}
