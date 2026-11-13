package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func oldTruth2(g *aig.Graph, x, y aig.Edge, table int) aig.Edge {
	out := g.Const0()
	for bits := 0; bits < 4; bits++ {
		if table&(1<<bits) == 0 {
			continue
		}
		a, b := x, y
		if bits&1 == 0 {
			a = a.Not()
		}
		if bits&2 == 0 {
			b = b.Not()
		}
		out = g.OR(out, g.AND(a, b))
	}
	return out
}
func TestOldSatCareAgainstTwoInputOracle(t *testing.T) {
	for truth := 1; truth < 16; truth++ {
		for assignment := 0; assignment < 4; assignment++ {
			if truth&(1<<assignment) == 0 {
				continue
			}
			g := aig.NewGraph()
			x, y := g.NewInput(), g.NewInput()
			g.AddOutput(oldTruth2(g, x, y, truth))
			cex := util.Abc_CexAlloc(0, 2, 1)
			for bit := 0; bit < 2; bit++ {
				if assignment&(1<<bit) != 0 {
					util.Abc_InfoSetBit(cex.PData, int32(bit))
				}
			}
			p := Saig_RefManStart(g, cex, 0, 0)
			care := Saig_RefManRunSat(p, 0)
			if care == nil {
				t.Fatalf("truth=%04b assignment=%02b: no care set", truth, assignment)
			}
			for other := 0; other < 4; other++ {
				match := true
				for bit := 0; bit < 2; bit++ {
					if util.Abc_InfoHasBit(care.PData, int32(bit)) && (assignment>>bit)&1 != (other>>bit)&1 {
						match = false
					}
				}
				if match && truth&(1<<other) == 0 {
					t.Fatalf("truth=%04b assignment=%02b care=%v allows %02b", truth, assignment, care.PData, other)
				}
			}
		}
	}
}
