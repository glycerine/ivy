package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSATBinaryQueriesAgainstTruthTables(t *testing.T) {
	for _, clause := range []bool{false, true} {
		for negLeft := int32(0); negLeft < 2; negLeft++ {
			for negRight := int32(0); negRight < 2; negRight++ {
				g := aig.NewGraph()
				a, b := g.NewInput(), g.NewInput()
				right := g.AND(a, b)
				g.AddOutput(a)
				g.AddOutput(right)
				p := satTestManager(g)
				var result int32
				if clause {
					result = Fra_NodesAreClause(p, a, right, negLeft, negRight)
				} else {
					result = Fra_NodesAreImp(p, a, right, negLeft, negRight)
				}
				valid := true
				for mask := 0; mask < 4; mask++ {
					leftValue := (mask&1 != 0) != (negLeft != 0)
					rightValue := (mask&1 != 0 && mask&2 != 0) != (negRight != 0)
					holds := !leftValue || rightValue
					if clause {
						holds = leftValue || rightValue
					}
					if !holds {
						valid = false
					}
				}
				if (result == 1) != valid {
					t.Fatalf("clause=%v signs=%d/%d got=%d valid=%v", clause, negLeft, negRight, result, valid)
				}
				if result == 0 {
					bits := p.PPatWords[0]
					av, bv := bits&1 != 0, bits&2 != 0
					leftValue := av != (negLeft != 0)
					rightValue := (av && bv) != (negRight != 0)
					violated := leftValue && !rightValue
					if clause {
						violated = !leftValue && !rightValue
					}
					if !violated {
						t.Fatalf("clause=%v signs=%d/%d invalid witness %b", clause, negLeft, negRight, bits)
					}
				}
			}
		}
	}
}
