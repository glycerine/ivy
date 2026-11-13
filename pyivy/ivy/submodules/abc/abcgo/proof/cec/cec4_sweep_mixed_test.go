package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"math/rand"
	"testing"
)

// Native XOR/MUX and complemented edges exercise the graph-to-SAT boundary,
// rather than only checking the AND-inverter encoding in isolation.
func TestCec4SweepMixedGatesExhaustive(t *testing.T) {
	for mode := int32(0); mode < 3; mode++ {
		for seed := int64(1); seed <= 18; seed++ {
			r := rand.New(rand.NewSource(seed))
			g := gia.Gia_ManStart(80)
			g.EnableMuxes()
			lits := []int32{gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)}
			gia.Gia_ManHashAlloc(g)
			for k := 0; k < 12; k++ {
				a, b, c := lits[r.Intn(len(lits))]^int32(r.Intn(2)), lits[r.Intn(len(lits))]^int32(r.Intn(2)), lits[r.Intn(len(lits))]^int32(r.Intn(2))
				var v int32
				switch k % 3 {
				case 0:
					v = gia.Gia_ManHashAnd(g, a, b)
				case 1:
					v = gia.Gia_ManHashXorReal(g, a, b)
				case 2:
					v = gia.Gia_ManHashMuxReal(g, a, b, c)
				}
				lits = append(lits, v)
			}
			for k := 0; k < 4; k++ {
				gia.Gia_ManAppendCo(g, lits[len(lits)-1-k]^int32(r.Intn(2)))
			}
			pars := Cec_ParFra_t{jType: mode, nWords: 1, nRounds: 1, nItersMax: 10, nBTLimit: 500, nCallsRecycle: 1, nSatVarMax: 1}
			out, cex, _ := cec4Sweep(g, &pars, false, true)
			if cex != nil || out == nil {
				t.Fatalf("mode=%d seed=%d missing result", mode, seed)
			}
			for mask := 0; mask < 64; mask++ {
				in := make([]bool, 6)
				for i := range in {
					in[i] = mask&(1<<i) != 0
				}
				want, err := g.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				got, err := out.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				for i := range want {
					if got[i] != want[i] {
						t.Fatalf("mode=%d seed=%d input=%06b output=%d got=%v want=%v", mode, seed, mask, i, got[i], want[i])
					}
				}
			}
		}
	}
}
