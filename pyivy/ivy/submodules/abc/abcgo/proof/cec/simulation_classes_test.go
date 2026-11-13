package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec/core"
	"testing"
)

func cecTestClassGraph() (*Gia_Man_t, []int32) {
	p := gia.Gia_ManStart(30)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	c := gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManAppendAnd(p, a, b)
	dup := gia.Gia_ManAppendAnd(p, a, b)
	other := gia.Gia_ManAppendAnd(p, a, c)
	neg := gia.Gia_ManAppendAnd(p, a, b^1)
	zero := gia.Gia_ManAppendAnd(p, x, neg)
	for _, l := range []int32{x, dup, other, zero} {
		gia.Gia_ManAppendCo(p, l)
	}
	return p, []int32{x >> 1, dup >> 1, other >> 1, zero >> 1}
}
func TestSimulationClassRefinementAndPool(t *testing.T) {
	aig, ids := cecTestClassGraph()
	pars := &Cec_ParSim_t{nWords: 2}
	p := Cec_ManSimStart(aig, pars)
	aig.StartRepresentatives()
	cecSimRefs(aig)
	for i := int32(1); i < gia.Gia_ManObjNum(aig); i++ {
		if gia.Gia_ObjIsAnd(gia.Gia_ManObj(aig, i)) != 0 {
			gia.Gia_ObjSetRepr(aig, i, 0)
		}
	}
	for w := 0; w < 2; w++ {
		for i, x := range []uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0} {
			vec.Vec_PtrEntry(p.vCiSimInfo, i).([]uint32)[w] = x
		}
	}
	if Cec_ManSimSimulateRound(p, p.vCiSimInfo, p.vCoSimInfo) != 0 {
		t.Fatal("unexpected miter check")
	}
	if !cecHead(aig, ids[0]) || gia.Gia_ObjRepr(aig, ids[1]) != ids[0] || gia.Gia_ObjRepr(aig, ids[2]) != gia.GIA_VOID || gia.Gia_ObjRepr(aig, ids[3]) != 0 {
		t.Fatal("wrong equivalence classes")
	}
	if p.nMems != 1 {
		t.Fatal("memory leak")
	}
	if Cec_ManSimClassRemoveOne(p, ids[1]) != 1 || cecClass(aig, ids[0]) || cecClass(aig, ids[1]) {
		t.Fatal("class removal")
	}
	p.nWords = 1
	Cec_ManSimSimulateRound(p, p.vCiSimInfo, p.vCoSimInfo)
	p.nWords = 2
	Cec_ManSimSimulateRound(p, p.vCiSimInfo, p.vCoSimInfo)
	// Force several growth operations while many records remain live.
	aig = gia.Gia_ManStart(200)
	a := gia.Gia_ManAppendCi(aig)
	b := gia.Gia_ManAppendCi(aig)
	var lits []int32
	for i := 0; i < 80; i++ {
		lits = append(lits, gia.Gia_ManAppendAnd(aig, a, b))
	}
	for _, l := range lits {
		gia.Gia_ManAppendCo(aig, l)
	}
	p = Cec_ManSimStart(aig, &Cec_ParSim_t{nWords: 1})
	aig.StartRepresentatives()
	cecSimRefs(aig)
	for _, l := range lits {
		gia.Gia_ObjSetRepr(aig, l>>1, 0)
	}
	p.pMems = make([]uint32, 8)
	p.nWordsOld = 1
	vec.Vec_PtrEntry(p.vCiSimInfo, 0).([]uint32)[0] = 0xaaaaaaaa
	vec.Vec_PtrEntry(p.vCiSimInfo, 1).([]uint32)[0] = 0xcccccccc
	Cec_ManSimSimulateRound(p, p.vCiSimInfo, nil)
	if p.nMems != 1 || len(cecClassMembers(aig, lits[0]>>1)) != 80 {
		t.Fatal("live pool growth lost equivalences")
	}
}
func TestSimulationCounterexampleAndPrepare(t *testing.T) {
	aig, _ := cecTestClassGraph()
	p := Cec_ManSimStart(aig, &Cec_ParSim_t{nWords: 1, nRounds: 1, fCheckMiter: 1})
	for i, x := range []uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0} {
		vec.Vec_PtrEntry(p.vCiSimInfo, i).([]uint32)[0] = x
	}
	aig.StartRepresentatives()
	cecSimRefs(aig)
	if Cec_ManSimSimulateRound(p, p.vCiSimInfo, p.vCoSimInfo) != 1 || p.pCexComb == nil {
		t.Fatal("missing counterexample")
	}
	in := make([]bool, 3)
	for i := range in {
		in[i] = cecInfoBit(p.pCexComb.PData, int32(i))
	}
	out, err := aig.Evaluate(in)
	if err != nil || !out[p.pCexComb.IPo] {
		t.Fatal("invalid reported counterexample", in, out)
	}
	aig, _ = cecTestClassGraph()
	p = Cec_ManSimStart(aig, &Cec_ParSim_t{nWords: 1, nRounds: 1, fVerbose: 1})
	Cec_ManSimClassesPrepare(p, -1)
	Cec_ManSimClassesRefine(p) // rounds<5 must not divide by zero
}
func TestBackwardGreedyWitnessAndImplication(t *testing.T) {
	aig, _ := cecTestClassGraph()
	cecSimSetPhases(aig)
	for id := int32(1); id < gia.Gia_ManObjNum(aig); id++ {
		o := gia.Gia_ManObj(aig, id)
		if gia.Gia_ObjIsAnd(o) == 0 {
			continue
		}
		for value := int32(0); value < 2; value++ {
			gia.Gia_ManIncrementTravId(aig)
			if Cec_ManCheckSat_rec(aig, o, value) == 0 {
				continue
			}
			in := make([]bool, gia.Gia_ManCiNum(aig))
			for i := range in {
				ci := gia.Gia_ManCi(aig, int32(i))
				in[i] = gia.Gia_ObjIsTravIdCurrent(aig, ci) != 0 && gia.Gia_ObjMark0(ci) != 0
			}
			m, err := core.NewModel(aig)
			if err != nil {
				t.Fatal(err)
			}
			vals, err := m.Values(in)
			if err != nil || vals[id] != (value != 0) {
				t.Fatal("invalid greedy witness")
			}
		}
	}
	if Cec_ManCheckSat2(aig, gia.Gia_ManConst0(aig), 1, nil) != 0 {
		t.Fatal("constant zero justified as one")
	}
	for _, words := range []int32{1, 2, 4} {
		gia.Gia_ManRandomW(1)
		p := Cec_ManSStart(aig, words)
		if Cec_ManSRunSim(p, 4, 5) != 1 {
			t.Fatal("identical gates not disproven as different")
		}
		if Cec_ManSRunSim(p, 4, 6) != 0 {
			t.Fatal("different gates eliminated")
		}
		for id := int32(1); id < gia.Gia_ManObjNum(aig); id++ {
			for _, v := range []int32{0, 1} {
				for _, w := range Cec_ManSSim(p, id, v) {
					if w != 0 {
						t.Fatal("stale implication signature")
					}
				}
			}
		}
	}
}

func TestForwardSimulationSpecialGates(t *testing.T) {
	g := gia.Gia_ManStart(30)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	c := gia.Gia_ManAppendCi(g)
	buf := gia.Gia_ManAppendBuf(g, a^1)
	xor := gia.Gia_ManAppendXorReal(g, a, b)
	g.EnableMuxes()
	mux := gia.Gia_ManAppendMuxReal(g, c, xor, buf)
	for _, l := range []int32{buf, xor, mux, mux ^ 1} {
		gia.Gia_ManAppendCo(g, l)
	}
	g.StartRepresentatives()
	cecSimRefs(g)
	p := Cec_ManSimStart(g, &Cec_ParSim_t{nWords: 2})
	p.vCoSimInfo = vec.Vec_PtrAllocSimInfo(4, 2)
	rows := [][2]uint32{{0xaaaaaaaa, 0x57491086}, {0xcccccccc, 0x17432901}, {0xf0f0f0f0, 0x27453012}}
	for i, row := range rows {
		copy(vec.Vec_PtrEntry(p.vCiSimInfo, i).([]uint32), row[:])
	}
	Cec_ManSimSimulateRound(p, p.vCiSimInfo, p.vCoSimInfo)
	for bit := int32(0); bit < 64; bit++ {
		in := make([]bool, 3)
		for i := range in {
			in[i] = cecInfoBit(vec.Vec_PtrEntry(p.vCiSimInfo, i).([]uint32), bit)
		}
		want, err := g.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		for i, v := range want {
			got := cecInfoBit(vec.Vec_PtrEntry(p.vCoSimInfo, i).([]uint32), bit)
			if got != v {
				t.Fatalf("bit%d output%d actual=%t expected=%t", bit, i, got, v)
			}
		}
	}
	if p.nMems != 1 {
		t.Fatal("special-gate record leak")
	}
}

// cecSimBack.c leaves Cec_ManCheckSat2 as a zero-returning stub, while its
// separate recursive heuristic treats a fresh constant as an assignable leaf.
func TestBackwardCheckSourceStubAndConstantBehavior(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	obj := gia.Gia_ManObj(g, x>>1)
	gia.Gia_ManIncrementTravId(g)
	if Cec_ManCheckSat2(g, obj, 1, nil) != 0 || gia.Gia_ObjIsTravIdCurrent(g, obj) != 0 {
		t.Fatal("stub changed traversal state")
	}
	zero := gia.Gia_ManConst0(g)
	if Cec_ManCheckSat_rec(g, zero, 1) != 1 || gia.Gia_ObjMark0(zero) != 1 {
		t.Fatal("source heuristic must assign fresh constant like a leaf")
	}
	if Cec_ManCheckSat_rec(g, zero, 0) != 0 {
		t.Fatal("source heuristic must reject conflicting assigned value")
	}
}

func TestBackwardWordSourceStoresOnlyMarkBit(t *testing.T) {
	g := gia.Gia_ManStart(4)
	id := gia.Gia_ManAppendCi(g)
	o := gia.Gia_ManObj(g, id>>1)
	gia.Gia_ManIncrementTravId(g)
	if Cec_ManCheckSat2_rec(g, o, 2, nil) != 1 || gia.Gia_ObjMark0(o) != 0 {
		t.Fatal("source mark bit truncation")
	}
	if Cec_ManCheckSat2_rec(g, o, 2, nil) != 0 {
		t.Fatal("source compares full word with stored bit")
	}
}
