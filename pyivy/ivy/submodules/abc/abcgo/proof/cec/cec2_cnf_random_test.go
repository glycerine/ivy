package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
	"math/rand"
	"testing"
)

// Compare incremental SAT queries against exhaustive evaluation for random
// shared AIGs, including complemented edges and structural MUX nodes.
func TestCec2CNFAgainstExhaustiveGraphs(t *testing.T) {
	r := rand.New(rand.NewSource(101))
	for trial := 0; trial < 60; trial++ {
		g := gia.Gia_ManStart(64)
		lits := []int32{gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)}
		for n := 0; n < 10; n++ {
			a, b := lits[r.Intn(len(lits))], lits[r.Intn(len(lits))]
			if a>>1 == b>>1 {
				b = lits[(int(b>>1))%len(lits)]
			}
			a ^= int32(r.Intn(2))
			b ^= int32(r.Intn(2))
			lits = append(lits, gia.Gia_ManAppendAnd(g, a, b))
		}
		for _, lit := range lits {
			gia.Gia_ManAppendCo(g, lit)
		}
		gia.Gia_ManCreateRefs(g)
		for _, lit := range lits[4:] {
			obj := gia.Gia_ManObj(g, lit>>1)
			if gia.Gia_ObjIsMuxType(obj) != 0 {
				gia.Gia_ObjSetMark0(obj, 1)
			}
		}
		sim := make([][]bool, 16)
		for mask := 0; mask < 16; mask++ {
			in := make([]bool, 4)
			for i := range in {
				in[i] = mask&(1<<i) != 0
			}
			var err error
			sim[mask], err = g.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
		}
		c := newCec2CNF(g, satoko.New())
		for q := 0; q < 15; q++ {
			i, j := r.Intn(len(lits)), r.Intn(len(lits))
			if i == j {
				continue
			}
			phase := r.Intn(2)
			want := false
			for mask := range sim {
				if (sim[mask][i] != sim[mask][j]) != (phase != 0) {
					want = true
					break
				}
			}
			status, _ := c.solveTwoWithCone(lits[i]>>1, lits[j]>>1, phase, 0, true)
			if (status == satoko.SATOKO_SAT) != want {
				t.Fatalf("trial %d query %d/%d phase %d got %d want SAT %v", trial, i, j, phase, status, want)
			}
		}
	}
}
