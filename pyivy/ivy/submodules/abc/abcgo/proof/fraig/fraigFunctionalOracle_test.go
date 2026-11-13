package fraig

import (
	"math/rand"
	"testing"
)

func TestFunctionalReductionAgainstExhaustiveTruthTables(t *testing.T) {
	r := rand.New(rand.NewSource(90210))
	for run := 0; run < 8; run++ {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FFuncRed: 1, FDoSparse: 1, FChoicing: int32(run % 2), NBTLimit: -1})
		nodes := make([]*Fraig_Node_t, 5)
		truth := make([]uint32, 5)
		for i := range nodes {
			nodes[i] = Fraig_ManReadIthVar(p, int32(i))
			for a := 0; a < 32; a++ {
				if a&(1<<uint(i)) != 0 {
					truth[i] |= 1 << uint(a)
				}
			}
		}
		for k := 0; k < 50; k++ {
			i, j := r.Intn(len(nodes)), r.Intn(len(nodes))
			left, right := nodes[i], nodes[j]
			x, y := truth[i], truth[j]
			if r.Intn(2) != 0 {
				left = Fraig_Not(left)
				x = ^x
			}
			if r.Intn(2) != 0 {
				right = Fraig_Not(right)
				y = ^y
			}
			var node *Fraig_Node_t
			var want uint32
			switch r.Intn(3) {
			case 0:
				node = Fraig_NodeAnd(p, left, right)
				want = x & y
			case 1:
				node = Fraig_NodeOr(p, left, right)
				want = x | y
			default:
				node = Fraig_NodeExor(p, left, right)
				want = x ^ y
			}
			nodes = append(nodes, node)
			truth = append(truth, want)
			for assignment := 0; assignment < 32; assignment++ {
				model := make([]int32, 5)
				for i := range model {
					model[i] = int32(assignment >> uint(i) & 1)
				}
				got := Fraig_ManSimulateBitNode(p, node, model) != 0
				expected := want&(1<<uint(assignment)) != 0
				if got != expected {
					t.Fatalf("run %d step %d assignment %d got %t want %t", run, k, assignment, got, expected)
				}
			}
		}
		Fraig_ManFree(p)
	}
}
