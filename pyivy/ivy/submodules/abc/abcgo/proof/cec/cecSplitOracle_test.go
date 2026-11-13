package cec

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// Exercise real SAT solving after a forced split, comparing both the answer and
// any reconstructed original-input witness with exhaustive enumeration.
func TestSplitAgainstExhaustiveTruthTables(t *testing.T) {
	random := rand.New(rand.NewSource(73021))
	for trial := 0; trial < 32; trial++ {
		g := gia.Gia_ManStart(100)
		gia.Gia_ManHashAlloc(g)

		lits := make([]int32, 4)
		for i := range lits {
			lits[i] = gia.Gia_ManAppendCi(g)
		}
		for j := 0; j < 18; j++ {
			a, b, c := lits[random.Intn(len(lits))]^int32(random.Intn(2)), lits[random.Intn(len(lits))]^int32(random.Intn(2)), lits[random.Intn(len(lits))]^int32(random.Intn(2))
			if a>>1 == b>>1 || a>>1 == c>>1 || b>>1 == c>>1 {
				a, b, c = lits[0]^int32(random.Intn(2)), lits[1]^int32(random.Intn(2)), lits[2]^int32(random.Intn(2))
			}
			_ = c
			out := gia.Gia_ManHashAnd(g, a, b)
			lits = append(lits, out)
		}
		out := lits[len(lits)-1]
		if trial%4 == 0 {
			tmp := gia.Gia_ManHashAnd(g, lits[0], lits[1])
			out = gia.Gia_ManHashAnd(g, tmp, lits[0]^1)
		}
		gia.Gia_ManAppendCo(g, out)
		sat := false
		for bits := 0; bits < 16; bits++ {
			input := make([]bool, 4)
			for i := range input {
				input[i] = bits&(1<<i) != 0
			}
			result, err := g.Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			sat = sat || result[0]
		}
		result := splitSolve(splitPart{graph: g}, 0)
		if (result.status == 0) != sat || result.status == -1 {
			t.Fatalf("trial %d status %d sat %v", trial, result.status, sat)
		}
		if sat {
			outputs, err := g.Evaluate(result.model)
			if err != nil || !outputs[0] {
				t.Fatalf("trial %d witness %v", trial, result.model)
			}
		}
	}
}
