package fraig

import (
	"math/rand"
	"testing"
)

func TestMultipleExternalClausesAgainstTruthTables(t *testing.T) {
	r := rand.New(rand.NewSource(511))
	for trial := 0; trial < 40; trial++ {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
		a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
		nodes := []*Fraig_Node_t{a, b, c, Fraig_NodeAnd(p, a, b)}
		truth := []uint8{0xaa, 0xcc, 0xf0, 0x88}
		allowed := uint8(0xff)
		for k := 0; k < 3; k++ {
			var clause []*Fraig_Node_t
			var mask uint8
			for {
				clause = nil
				mask = 0
				count := 1 + r.Intn(2)
				for z := 0; z < count; z++ {
					i := r.Intn(len(nodes))
					edge, bits := nodes[i], truth[i]
					if r.Intn(2) != 0 {
						edge = Fraig_Not(edge)
						bits = ^bits
					}
					clause = append(clause, edge)
					mask |= bits
				}
				if allowed&mask != 0 {
					break
				}
			}
			Fraig_ManAddClause(p, clause, int32(len(clause)))
			allowed &= mask
		}
		for i := 0; i < 3; i++ {
			for j := i + 1; j < 3; j++ {
				for phase := 0; phase < 4; phase++ {
					x, y := nodes[i], nodes[j]
					mx, my := truth[i], truth[j]
					if phase&1 != 0 {
						x = Fraig_Not(x)
						mx = ^mx
					}
					if phase&2 != 0 {
						y = Fraig_Not(y)
						my = ^my
					}
					want := allowed & ^(mx|my) == 0
					got := Fraig_ManCheckClauseUsingSat(p, x, y, -1) != 0
					if got != want {
						t.Fatalf("trial %d %d,%d phase %d got %t want %t", trial, i, j, phase, got, want)
					}
				}
			}
		}
		Fraig_ManFree(p)
	}
}
