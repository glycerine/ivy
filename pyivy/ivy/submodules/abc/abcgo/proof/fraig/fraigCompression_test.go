package fraig

import (
	"github.com/glycerine/abcgo/sat/msat"
	"testing"
)

func TestFeedbackCompressionAndResize(t *testing.T) {
	for _, sparse := range []int32{0, 1} {
		t.Run(map[int32]string{0: "normal", 1: "sparse"}[sparse], func(t *testing.T) {
			p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
			defer Fraig_ManFree(p)
			p.fDoSparse = sparse
			a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
			root := Fraig_NodeAnd(p, a, b)
			vars := msat.Msat_IntVecAlloc(2)
			msat.Msat_IntVecPush(vars, int(a.Num))
			msat.Msat_IntVecPush(vars, int(b.Num))
			model := make([]int, len(p.vNodes.pArray))
			for i := 0; i < 4; i++ {
				model[a.Num] = int(a.Num)*2 + (i & 1)
				model[b.Num] = int(b.Num)*2 + ((i >> 1) & 1)
				Fraig_FeedBack(p, model, vars, a, b)
			}
			if p.nWordsDyna != 8 || p.iPatsPerm != 4 {
				t.Fatalf("compression: words=%d permanent=%d", p.nWordsDyna, p.iPatsPerm)
			}
			for i := int32(0); i < 4; i++ {
				va, vb := fraigBit(a.puSimD, i), fraigBit(b.puSimD, i)
				if fraigBit(root.puSimD, i) != (va && vb) {
					t.Fatalf("bad simulation in pattern %d", i)
				}
			}
		})
	}
}
