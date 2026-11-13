package core_test

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec/core"
)

// Exercise the public GIA adapter, rather than the core test circuit, against
// GIA's independent evaluator for every assignment of each small network.
func TestGIAGateQueriesAgainstExhaustiveEvaluation(t *testing.T) {
	r := rand.New(rand.NewSource(8493))
	for trial := 0; trial < 30; trial++ {
		g := gia.Gia_ManStart(80)
		g.EnableMuxes()
		lits := []int32{0}
		for i := 0; i < 4; i++ {
			lits = append(lits, gia.Gia_ManAppendCi(g))
		}
		pick := func() int32 { return lits[r.Intn(len(lits))] ^ int32(r.Intn(2)) }
		for i := 0; i < 20; i++ {
			a, b, c := pick(), pick(), pick()
			for b>>1 == a>>1 {
				b = pick()
			}
			for c>>1 == a>>1 || c>>1 == b>>1 {
				c = pick()
			}
			switch r.Intn(4) {
			case 0:
				lits = append(lits, gia.Gia_ManAppendAnd(g, a, b))
			case 1:
				lits = append(lits, gia.Gia_ManAppendXorReal(g, a, b))
			case 2:
				lits = append(lits, gia.Gia_ManAppendMuxReal(g, c, a, b))
			case 3:
				lits = append(lits, gia.Gia_ManAppendBuf(g, a))
			}
		}
		for _, lit := range lits {
			gia.Gia_ManAppendCo(g, lit)
		}
		for _, flip := range []bool{false, true} {
			m, err := core.NewManager(g, core.Options{PolarFlip: flip, MaxSATVariables: 10, RecycleQueries: 1})
			if err != nil {
				t.Fatal(err)
			}
			for query := 0; query < 15; query++ {
				i, j := r.Intn(len(lits)), r.Intn(len(lits))
				a, b := lits[i]^int32(r.Intn(2)), lits[j]^int32(r.Intn(2))
				equal := true
				for bits := 0; bits < 16; bits++ {
					in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
					out, err := g.Evaluate(in)
					if err != nil {
						t.Fatal(err)
					}
					if (out[i] != (a != lits[i])) != (out[j] != (b != lits[j])) {
						equal = false
						break
					}
				}
				result, err := m.CheckPair(a, b)
				if err != nil || result.Status == core.Undecided || (result.Status == core.Equivalent) != equal {
					t.Fatalf("trial=%d flip=%v pair=%d,%d equal=%v result=%+v error=%v", trial, flip, a, b, equal, result, err)
				}
				if result.Status == core.Different {
					out, err := g.Evaluate(result.Counterexample.Inputs)
					if err != nil || (out[i] != (a != lits[i])) == (out[j] != (b != lits[j])) {
						t.Fatalf("invalid witness: %+v %v", result, err)
					}
				}
			}
		}
	}
}
