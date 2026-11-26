package util

import (
	"math/rand"
	"testing"
)

// ttMapReference is the bit-level definition: out[i] = in[mapping(i)].
func ttMapReference(t []word, n int32, mapping func(int) int) []word {
	out := make([]word, n)
	for i := 0; i < int(n)*64; i++ {
		ttPut(out, i, ttBit(t, mapping(i)))
	}
	return out
}

func TestTtWordOpsMatchBitDefinition(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	for it := 0; it < 3000; it++ {
		nVars := int32(1 + r.Intn(10))
		n := int32(1)
		if nVars > 6 {
			n = 1 << (nVars - 6)
		}
		in := make([]word, n)
		for i := range in {
			in[i] = r.Uint64()
		}
		v := int32(r.Intn(int(max(nVars, 6))))
		check := func(name string, op func([]word), mapping func(int) int) {
			got := append([]word(nil), in...)
			op(got)
			want := ttMapReference(in, n, mapping)
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("%s nVars %d v %d word %d: got %016x want %016x", name, nVars, v, i, got[i], want[i])
				}
			}
		}
		check("cof0", func(x []word) { Abc_TtCofactor0(x, n, v) }, func(i int) int { return i &^ (1 << v) })
		check("cof1", func(x []word) { Abc_TtCofactor1(x, n, v) }, func(i int) int { return i | (1 << v) })
		check("flip", func(x []word) { Abc_TtFlip(x, n, v) }, func(i int) int { return i ^ (1 << v) })
		if v+1 < max(nVars, 6) {
			check("swap", func(x []word) { Abc_TtSwapAdjacent(x, n, v) }, func(k int) int {
				if (k>>v)&1 != (k>>(v+1))&1 {
					k ^= 1<<v | 1<<(v+1)
				}
				return k
			})
		}
	}
}
