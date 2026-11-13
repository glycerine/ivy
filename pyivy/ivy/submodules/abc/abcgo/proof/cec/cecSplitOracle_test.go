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
		g.EnableMuxes()
		lits := make([]int32, 4)
		for i := range lits {
			lits[i] = gia.Gia_ManAppendCi(g)
		}
		for j := 0; j < 18; j++ {
			a, b, c := lits[random.Intn(len(lits))]^int32(random.Intn(2)), lits[random.Intn(len(lits))]^int32(random.Intn(2)), lits[random.Intn(len(lits))]^int32(random.Intn(2))
			if a>>1 == b>>1 || a>>1 == c>>1 || b>>1 == c>>1 {
				a, b, c = lits[0]^int32(random.Intn(2)), lits[1]^int32(random.Intn(2)), lits[2]^int32(random.Intn(2))
			}
			var out int32
			switch j % 3 {
			case 0:
				out = gia.Gia_ManAppendAnd(g, a, b)
			case 1:
				out = gia.Gia_ManAppendXorReal(g, a, b)
			case 2:
				out = gia.Gia_ManAppendMuxReal(g, a, b, c)
			}
			lits = append(lits, out)
		}
		out := lits[len(lits)-1]
		if trial%4 == 0 {
			tmp := gia.Gia_ManAppendAnd(g, lits[0], lits[1])
			out = gia.Gia_ManAppendAnd(g, tmp, lits[0]^1)
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
		for _, workers := range []int32{1, 3} {
			solve := func(part splitPart, timeout int32) splitResult {
				if len(part.fixes) == 0 {
					return splitResult{status: -1}
				}
				return splitSolve(part, timeout)
			}
			status, model := splitRun(g, workers, 0, 0, 2, 0, 0, 1, solve)
			// Constant outputs have no split variable, so use the direct solver.
			if status == -1 {
				status, model = splitRun(g, workers, 0, 0, 2, 0, 0, 1, splitSolve)
			}
			if (status == 0) != sat || status == -1 {
				t.Fatalf("trial %d workers %d status %d sat %t", trial, workers, status, sat)
			}
			if sat {
				outputs, err := g.Evaluate(model)
				if err != nil || !outputs[0] {
					t.Fatalf("trial %d workers %d invalid witness %v outputs %v err %v", trial, workers, model, outputs, err)
				}
			}
		}
	}
}
