package ivy

import (
	"slices"
	"sort"
	"testing"
)

func TestIvyCutGetTruthTracksLatchDepthAndPolarity(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	latch := Ivy_Latch(p, a, IVY_INIT_0)
	root := Ivy_And(p, latch, Ivy_Not(b))
	leaves := []int32{Ivy_LeafCreate(a.Id, 1), Ivy_LeafCreate(b.Id, 0)}
	truth := Ivy_CutGetTruth(p, Ivy_Regular(root), leaves, int32(len(leaves)))
	want := ivySeqTruthMasks[0] & ^ivySeqTruthMasks[1]
	if truth != want {
		t.Fatalf("truth = %#08x, want %#08x", truth, want)
	}
}

func TestIvyCutGetTruthRejectsMissingLeaf(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	latch := Ivy_Latch(p, a, IVY_INIT_0)
	root := Ivy_And(p, latch, b)
	defer func() {
		if recover() == nil {
			t.Fatal("missing latch-depth leaf did not fail")
		}
	}()
	Ivy_CutGetTruth(p, root, []int32{Ivy_LeafCreate(a.Id, 0), Ivy_LeafCreate(b.Id, 0)}, 2)
}

func ivySeqEvaluate(p *Ivy_Man_t, piBits, latchBits int) ([]bool, []bool) {
	values := map[*Ivy_Obj_t]bool{p.pConst1: true}
	latches := make([]*Ivy_Obj_t, 0)
	for _, n := range p.vObjs {
		if n != nil && Ivy_ObjIsLatch(n) != 0 {
			latches = append(latches, n)
		}
	}
	for i, n := range p.vPis {
		values[n] = piBits&(1<<i) != 0
	}
	for i, n := range latches {
		values[n] = latchBits&(1<<i) != 0
	}
	var eval func(*Ivy_Obj_t) bool
	eval = func(edge *Ivy_Obj_t) bool {
		n := Ivy_Regular(edge)
		value, known := values[n]
		if !known {
			a := eval(Ivy_ObjChild0(n))
			switch n.Type {
			case IVY_AND:
				value = a && eval(Ivy_ObjChild1(n))
			case IVY_EXOR:
				value = a != eval(Ivy_ObjChild1(n))
			case IVY_BUF, IVY_PO:
				value = a
			default:
				panic("unsupported sequential evaluation node")
			}
			values[n] = value
		}
		if Ivy_IsComplement(edge) != 0 {
			return !value
		}
		return value
	}
	outputs := make([]bool, len(p.vPos))
	for i, n := range p.vPos {
		outputs[i] = eval(n)
	}
	next := make([]bool, len(latches))
	for i, n := range latches {
		next[i] = eval(Ivy_ObjChild0(n))
	}
	return outputs, next
}

func TestIvySequentialRewritePreservesSourceNetworkBehavior(t *testing.T) {
	p := Ivy_ManStart()
	for i := 0; i < 4; i++ {
		Ivy_ObjCreatePi(p)
	}
	var latched []*Ivy_Obj_t
	for _, pi := range p.vPis {
		latched = append(latched, Ivy_Latch(p, pi, IVY_INIT_DC))
	}
	left := Ivy_And(p, latched[0], p.vPis[1])
	right := Ivy_And(p, latched[2], p.vPis[3])
	root := Ivy_And(p, left, right)
	Ivy_ObjCreatePo(p, root)
	cuts := Ivy_CutComputeForNode(p, root, 5)
	found := false
	for i := int32(1); i < cuts.nCuts; i++ {
		cut := &cuts.pCuts[i]
		if cut.nSize != 4 {
			continue
		}
		depths := make([]int32, 0, 4)
		for _, leaf := range cut.pArray[:cut.nSize] {
			depths = append(depths, Ivy_LeafLat(leaf))
		}
		sort.Slice(depths, func(i, j int) bool { return depths[i] < depths[j] })
		if depths[0] == 0 && depths[1] == 0 && depths[2] == 1 && depths[3] == 1 {
			found = true
			if got, want := Ivy_CutGetTruth(p, root, cut.pArray[:cut.nSize], 4), ivySeqTruthMasks[0]&ivySeqTruthMasks[1]&ivySeqTruthMasks[2]&ivySeqTruthMasks[3]; got != want {
				t.Fatalf("four-latch cut truth %#x, want %#x", got, want)
			}
			break
		}
	}
	if !found {
		t.Fatal("sequential cut enumeration did not preserve four one-latch leaves")
	}
	q := Ivy_ManDup(p)
	defer Ivy_ManStop(p)
	defer Ivy_ManStop(q)
	createdBeforeRewrite := q.nCreated
	if Ivy_ManRewriteSeq(q, 1, 0) == 0 {
		t.Fatal("sequential rewrite did not start")
	}
	if q.nCreated <= createdBeforeRewrite {
		t.Fatal("sequential rewrite did not build/update a latch-aware candidate")
	}
	if Ivy_ManCheck(q) == 0 || Ivy_ManLatchNum(p) != Ivy_ManLatchNum(q) {
		t.Fatal("sequential rewrite changed graph invariants or latch count")
	}
	for pis := 0; pis < 16; pis++ {
		for states := 0; states < 4; states++ {
			beforeOut, beforeNext := ivySeqEvaluate(p, pis, states)
			afterOut, afterNext := ivySeqEvaluate(q, pis, states)
			if !slices.Equal(beforeOut, afterOut) || !slices.Equal(beforeNext, afterNext) {
				t.Fatalf("source network behavior differs for PI=%04b state=%04b", pis, states)
			}
		}
	}
}
