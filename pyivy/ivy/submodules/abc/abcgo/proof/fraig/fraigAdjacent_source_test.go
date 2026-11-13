package fraig

import (
	"github.com/glycerine/abcgo/sat/msat"
	"reflect"
	"testing"
)

// The source rebuilds the selected cone's adjacency in two passes: fanins
// first, then fanouts, clearing only the entries selected by vConeVars.
func TestSourceSetupAdjacentRebuildsCone(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	left := Fraig_NodeAnd(p, a, b)
	right := Fraig_NodeAnd(p, b, c)
	fraigSatEnsureVariables(p)
	Fraig_OrderVariables(p, left, right)
	vars := append([]int(nil), msat.Msat_IntVecReadArray(p.vVarsInt)...)
	expected := make(map[int][]int)
	for _, id := range vars {
		n := p.vNodes.pArray[id]
		if Fraig_NodeIsAnd(n) != 0 {
			for _, edge := range n.vFanins.pArray {
				expected[id] = append(expected[id], int(Fraig_Regular(edge).Num))
			}
		}
	}
	for _, id := range vars {
		n := p.vNodes.pArray[id]
		if Fraig_NodeIsAnd(n) != 0 {
			for _, edge := range n.vFanins.pArray {
				child := int(Fraig_Regular(edge).Num)
				expected[child] = append(expected[child], id)
			}
		}
	}
	for _, id := range vars {
		msat.Msat_IntVecPush(p.vAdjacents[id], 9999)
	}
	Fraig_SetupAdjacent(p, p.vVarsInt)
	for _, id := range vars {
		got := msat.Msat_IntVecReadArray(p.vAdjacents[id])
		if !reflect.DeepEqual(append([]int(nil), got...), expected[id]) {
			t.Fatalf("node %d adjacency %v want %v", id, got, expected[id])
		}
	}
}
