package fraig

import (
	"math/rand"
	"testing"
)

func TestSATAgainstRandomSmallTruthTables(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	nodes := []*Fraig_Node_t{Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2), Fraig_ManReadIthVar(p, 3)}
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 36; i++ {
		a, b := nodes[r.Intn(len(nodes))], nodes[r.Intn(len(nodes))]
		if r.Intn(2) == 0 {
			a = Fraig_Not(a)
		}
		if r.Intn(2) == 0 {
			b = Fraig_Not(b)
		}
		switch r.Intn(3) {
		case 0:
			nodes = append(nodes, Fraig_NodeAnd(p, a, b))
		case 1:
			nodes = append(nodes, Fraig_NodeOr(p, a, b))
		default:
			nodes = append(nodes, Fraig_NodeExor(p, a, b))
		}
	}
	signatures := make([]uint16, len(nodes))
	for i, n := range nodes {
		for assignment := 0; assignment < 16; assignment++ {
			model := []int32{int32(assignment & 1), int32(assignment >> 1 & 1), int32(assignment >> 2 & 1), int32(assignment >> 3 & 1)}
			if Fraig_ManSimulateBitNode(p, n, model) != 0 {
				signatures[i] |= 1 << uint(assignment)
			}
		}
	}
	for i := 0; i < 120; i++ {
		a, b := r.Intn(len(nodes)), r.Intn(len(nodes))
		got := Fraig_NodesAreEqual(p, nodes[a], nodes[b], -1, 5) != 0
		want := signatures[a] == signatures[b]
		if got != want {
			t.Fatalf("pair %d,%d: got %t truth %04x %04x", a, b, got, signatures[a], signatures[b])
		}
	}
}
