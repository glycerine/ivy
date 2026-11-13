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

// ABC fraigFeed.c compacts permanent counterexamples, rehashes sparse nodes,
// and may resize simulation storage after each full dynamic buffer.
func TestFeedbackRepeatedCompressionAcrossWordBoundaries(t *testing.T) {
	for _, sparse := range []int32{0, 1} {
		t.Run(map[int32]string{0: "normal", 1: "sparse"}[sparse], func(t *testing.T) {
			p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1, FDoSparse: sparse})
			defer Fraig_ManFree(p)
			a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
			root := Fraig_NodeAnd(p, a, b)
			vars := msat.Msat_IntVecAlloc(2)
			msat.Msat_IntVecPush(vars, int(a.Num))
			msat.Msat_IntVecPush(vars, int(b.Num))
			model := make([]int, len(p.vNodes.pArray))
			for i := 0; i < 64; i++ {
				model[a.Num] = int(a.Num)*2 + (i & 1)
				model[b.Num] = int(b.Num)*2 + ((i >> 1) & 1)
				Fraig_FeedBack(p, model, vars, a, b)
				for w := int32(0); w < p.iWordStart; w++ {
					if root.puSimD[w] != a.puSimD[w]&b.puSimD[w] {
						t.Fatalf("after feedback %d, word %d root=%x inputs=%x,%x", i, w, root.puSimD[w], a.puSimD[w], b.puSimD[w])
					}
				}
			}
			if p.iPatsPerm <= 32 || p.iWordPerm < 2 {
				t.Fatalf("fixture did not cross a permanent word boundary: patterns=%d words=%d", p.iPatsPerm, p.iWordPerm)
			}
		})
	}
}

func TestFeedbackCompressionPreservesChoiceClassSimulation(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1, FFuncRed: 1, FDoSparse: 1, FChoicing: 1})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	left := Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, b), c)
	if Fraig_NodeAnd(p, a, Fraig_NodeAnd(p, b, c)) != left {
		t.Fatal("associative representative")
	}
	choice := Fraig_Regular(left).pNextE
	if choice == nil {
		t.Fatal("expected independent choice")
	}
	vars := msat.Msat_IntVecAlloc(3)
	for _, n := range []*Fraig_Node_t{a, b, c} {
		msat.Msat_IntVecPush(vars, int(n.Num))
	}
	model := make([]int, len(p.vNodes.pArray))
	for i := 0; i < 64; i++ {
		for j, n := range []*Fraig_Node_t{a, b, c} {
			model[n.Num] = int(n.Num)*2 + ((i >> j) & 1)
		}
		Fraig_FeedBack(p, model, vars, a, b)
		if Fraig_Regular(left).pNextE != choice || choice.pRepr != Fraig_Regular(left) {
			t.Fatal("choice links changed")
		}
		for w := int32(0); w < p.iWordStart; w++ {
			if choice.puSimD[w] != Fraig_Regular(left).puSimD[w] {
				t.Fatalf("feedback %d word %d choice simulation differs", i, w)
			}
		}
	}
	if p.iPatsPerm <= 32 {
		t.Fatalf("fixture did not cross a permanent word boundary: %d", p.iPatsPerm)
	}
}
