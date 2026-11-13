package fraig

import (
	"github.com/glycerine/abcgo/sat/msat"
	"math/rand"
	"testing"
)

// Exercise multiple resizes and partial permanent words on a mixed DAG. Each
// stored simulation bit is checked against independent scalar AIG evaluation.
func TestFeedbackCompressionMixedConeStress(t *testing.T) {
	for _, sparse := range []int32{0, 1} {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FDoSparse: sparse})
		inputs := make([]*Fraig_Node_t, 8)
		vars := msat.Msat_IntVecAlloc(8)
		for i := range inputs {
			inputs[i] = Fraig_ManReadIthVar(p, int32(i))
			msat.Msat_IntVecPush(vars, int(inputs[i].Num))
		}
		a := Fraig_NodeAnd(p, inputs[0], Fraig_Not(inputs[1]))
		b := Fraig_NodeOr(p, inputs[2], inputs[3])
		c := Fraig_NodeExor(p, inputs[4], inputs[5])
		root := Fraig_NodeMux(p, inputs[6], Fraig_NodeAnd(p, a, b), Fraig_NodeOr(p, c, inputs[7]))
		nodes := []*Fraig_Node_t{a, b, c, root}
		model := make([]int, len(p.vNodes.pArray))
		random := rand.New(rand.NewSource(6701))
		for iter := 0; iter < 160; iter++ {
			for _, n := range inputs {
				model[n.Num] = int(n.Num)*2 + random.Intn(2)
			}
			Fraig_FeedBack(p, model, vars, inputs[0], inputs[1])
			for bit := int32(0); bit < p.iPatsPerm; bit++ {
				scalar := make([]int32, len(inputs))
				for i, n := range inputs {
					if fraigBit(n.puSimD, bit) {
						scalar[i] = 1
					}
				}
				for _, edge := range nodes {
					n := Fraig_Regular(edge)
					got := fraigBit(n.puSimD, bit)
					if n.fInv != 0 {
						got = !got
					}
					if Fraig_IsComplement(edge) {
						got = !got
					}
					want := fraigTestSimulateBitNode(p, edge, scalar) != 0
					if got != want {
						t.Fatalf("sparse=%d iteration=%d bit=%d node=%d got=%v want=%v", sparse, iter, bit, n.Num, got, want)
					}
				}
			}
		}
		if p.iPatsPerm <= 64 {
			t.Fatalf("sparse=%d did not reach multiple permanent words: %d", sparse, p.iPatsPerm)
		}
		Fraig_ManFree(p)
	}
}
