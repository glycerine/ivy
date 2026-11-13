package fraig

import "testing"

func TestSATMatchesExhaustiveSmallCircuits(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	edges := []*Fraig_Node_t{a, b, c, Fraig_NodeAnd(p, a, b), Fraig_NodeAnd(p, a, c), Fraig_NodeOr(p, a, b), Fraig_NodeExor(p, a, b), Fraig_NodeMux(p, c, a, b), Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, b), c), Fraig_NodeAnd(p, a, Fraig_NodeAnd(p, b, c))}
	signatures := make([]uint8, len(edges))
	for i, n := range edges {
		for bits := 0; bits < 8; bits++ {
			model := []int32{int32(bits & 1), int32(bits >> 1 & 1), int32(bits >> 2 & 1)}
			if fraigTestSimulateBitNode(p, n, model) != 0 {
				signatures[i] |= 1 << uint(bits)
			}
		}
	}
	for i, a := range edges {
		for j := 0; j < i; j++ {
			b := edges[j]
			got := Fraig_NodesAreEqual(p, a, b, -1, 5) != 0
			want := signatures[i] == signatures[j]
			if got != want {
				t.Fatalf("edges %d and %d: SAT equality %t, truth tables %02x %02x", i, j, got, signatures[i], signatures[j])
			}
		}
	}
}
