package framgr

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// Exhaustively compare small mixed AIGs with their swept versions. This checks
// all outputs together, including complemented edges and expanded XOR nodes.
func TestSweepPreservesMixedGraphTruthTables(t *testing.T) {
	r := rand.New(rand.NewSource(20261113))
	for graphNo := 0; graphNo < 40; graphNo++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.Const1()}
		for i := 0; i < 4; i++ {
			edges = append(edges, g.NewInput())
		}
		pick := func() aig.Edge {
			e := edges[r.Intn(len(edges))]
			if r.Intn(2) == 0 {
				e = e.Not()
			}
			return e
		}
		for i := 0; i < 30; i++ {
			a, b := pick(), pick()
			switch r.Intn(3) {
			case 0:
				edges = append(edges, g.AND(a, b))
			case 1:
				edges = append(edges, g.XOR(a, b))
			case 2:
				edges = append(edges, g.XOR(a, b))
			}
		}
		for i := 0; i < 5; i++ {
			g.AddOutput(pick())
		}
		result, _ := Fra_FraigEquivenceMapped(g, 1000, 0, nil)
		for bits := 0; bits < 16; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
			want, err := g.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			got, err := result.Evaluate(in)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("graph %d inputs %v: got %v, want %v (error %v)", graphNo, in, got, want, err)
			}
		}
	}
}
