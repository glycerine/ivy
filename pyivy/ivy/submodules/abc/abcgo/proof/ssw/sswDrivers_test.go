package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func driverGraph(invert bool) (*aig.Graph, aig.Edge) {
	g := aig.NewGraph()
	pi, reg := g.NewInput(), g.NewInput()
	out := g.AND(pi, reg)
	if invert {
		out = aig.Aig_Not(out)
	}
	g.AddOutput(out)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	return g, out
}
func TestSequentialDriversEquivalentAndInequivalent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		invert bool
		want   int32
	}{{"equivalent", false, 1}, {"inequivalent", true, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			for _, driver := range []string{"general", "pairs", "similarity", "similarity-pairs", "miter"} {
				t.Run(driver, func(t *testing.T) {
					a, x := driverGraph(false)
					b, y := driverGraph(tc.invert)
					ids1, ids2 := vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
					vec.Vec_IntPush(ids1, int32(aig.Aig_Regular(x).NodeID()))
					vec.Vec_IntPush(ids2, int32(aig.Aig_Regular(y).NodeID()))
					var got int32
					switch driver {
					case "general":
						got = Ssw_SecGeneral(a, b, nil)
					case "pairs":
						got = Ssw_SecWithPairs(a, b, ids1, ids2, nil)
					case "similarity":
						got = Ssw_SecWithSimilarity(a, b, nil)
					case "similarity-pairs":
						got = Ssw_SecWithSimilarityPairs(a, b, saig.Saig_StrSimPerformMatching(a, b, 0, 0, nil), nil)
					case "miter":
						got = Ssw_SecGeneralMiter(saig.Saig_ManCreateMiter(a, b, 0), nil)
					}
					if got != tc.want {
						t.Fatalf("got %d, want %d", got, tc.want)
					}
				})
			}
		})
	}
}
