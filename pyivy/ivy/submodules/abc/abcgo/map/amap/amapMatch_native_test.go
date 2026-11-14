package amap

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func matchFixture() (*Amap_Man_t, *Amap_Obj_t, *Amap_Obj_t, *Amap_Obj_t) {
	lib := Amap_LibAlloc()
	vec.Vec_PtrPush(lib.vGates, &Amap_Gat_t{dArea: 3, nPins: 2, Pins: make([]Amap_Pin_t, 2)})
	set := &Amap_Set_t{nIns: 2}
	set.Ins[0] = 0
	set.Ins[1] = 2
	lib.pNodes = make([]Amap_Nod_t, 2)
	lib.pNodes[1].pSets = set
	p := &Amap_Man_t{pLib: lib, pPars: &Amap_Par_t{FEpsilon: 0.001, FADratio: 1}, fAreaInv: 0.5, fEpsilonInternal: 0.001, vObjs: vec.Vec_PtrAlloc(8), vPis: vec.Vec_PtrAlloc(2), vPos: vec.Vec_PtrAlloc(2)}
	a := &Amap_Obj_t{Type: AMAP_OBJ_PI, Id: 0, EstRefs: 1}
	b := &Amap_Obj_t{Type: AMAP_OBJ_PI, Id: 1, EstRefs: 1}
	cut := &Amap_Cut_t{iMat: 1, nFans: 2, Fans: []int32{0, 2}}
	n := &Amap_Obj_t{Type: AMAP_OBJ_AND, Id: 2, nRefs: 2, pData: cut, Best: Amap_Mat_t{pCut: cut, pSet: set, Area: 3, Delay: 1}}
	for _, o := range []*Amap_Obj_t{a, b, n} {
		vec.Vec_PtrPush(p.vObjs, o)
	}
	vec.Vec_PtrPush(p.vPis, a)
	vec.Vec_PtrPush(p.vPis, b)
	for i, lit := range []int32{4, 5} {
		po := &Amap_Obj_t{Type: AMAP_OBJ_PO, Id: uint32(3 + i), Fan: [3]int32{lit}}
		vec.Vec_PtrPush(p.vObjs, po)
		vec.Vec_PtrPush(p.vPos, po)
	}
	return p, a, b, n
}

func TestMatchMappingSharedPolarities(t *testing.T) {
	p, a, b, n := matchFixture()
	if area := Amap_ManComputeMapping(p); area != 3 {
		t.Fatal(area)
	}
	if n.nFouts != [2]int32{1, 1} || a.nFouts != [2]int32{1, 0} || b.nFouts != [2]int32{1, 0} {
		t.Fatal("shared logic counted twice", n.nFouts, a.nFouts, b.nFouts)
	}
	if Amap_ManCountInverters(p) != 1 {
		t.Fatal("missing opposite-polarity inverter")
	}
	n.Best.Delay = 1.75
	if Amap_ManMaxDelay(p) != 1.75 {
		t.Fatal("fractional delay was truncated or complemented fanin not regularized")
	}
}

func TestMatchExactAreaRollbackAndPolarity(t *testing.T) {
	p, a, b, n := matchFixture()
	n.Best.pSet.Ins[0] = 1
	if got := Amap_CutAreaDerefed2(p, n, &n.Best); got != 3.5 {
		t.Fatal(got)
	}
	if a.nFouts != [2]int32{} || b.nFouts != [2]int32{} {
		t.Fatal("temporary references leaked")
	}
	if got := Amap_CutAreaRef(p, &n.Best); got != 3.5 {
		t.Fatal(got)
	}
	if a.nFouts != [2]int32{0, 1} {
		t.Fatal(a.nFouts)
	}
	if got := Amap_CutAreaDeref(p, &n.Best); got != 3.5 {
		t.Fatal(got)
	}
	if a.nFouts != [2]int32{} || b.nFouts != [2]int32{} {
		t.Fatal("reference/dereference did not balance")
	}
	n.nFouts[1] = 1
	if got := Amap_CutAreaDerefed2(p, n, &n.Best); got != 4 {
		t.Fatal("output inverter not included", got)
	}
}

func TestMatchNodeFlowsAndIndependentBestCut(t *testing.T) {
	p, a, b, n := matchFixture()
	a.Best.Delay = 0.25
	b.Best.Delay = 1.25
	a.Best.Area = 2
	b.Best.Area = 4
	a.nFouts[0] = 1
	b.nFouts[0] = 2
	a.EstRefs = 2
	b.EstRefs = 4
	original := n.pData.(*Amap_Cut_t)
	Amap_ManMatchNode(p, n, 1, 0)
	if n.Best.Delay != 2.25 || n.Best.Area != 5 || n.Best.AveFan != 1.5 {
		t.Fatal(n.Best)
	}
	original.Fans[0] = 99
	if n.Best.pCut.Fans[0] != 0 {
		t.Fatal("best cut aliases scratch cut")
	}
}

func TestMatchBoundedAreaRestoresRepeatedCounter(t *testing.T) {
	p, a, _, n := matchFixture()
	n.Best.pCut.Fans[1] = 0
	if got := Amap_CutAreaDerefed2(p, n, &n.Best); got != 3 {
		t.Fatal(got)
	}
	if a.nFouts != [2]int32{} {
		t.Fatal("duplicate fanin references leaked", a.nFouts)
	}
}

func TestMatchExactAreaEightLevelBoundary(t *testing.T) {
	p, a, b, n := matchFixture()
	// Replace the output objects with a deep chain: bounded recovery counts the
	// root plus eight recursive levels, then stops before referencing that cut.
	p.vObjs = vec.Vec_PtrAlloc(20)
	vec.Vec_PtrPush(p.vObjs, a)
	vec.Vec_PtrPush(p.vObjs, b)
	previous := a
	for i := 0; i < 12; i++ {
		cut := &Amap_Cut_t{iMat: 1, nFans: 2, Fans: []int32{int32(previous.Id << 1), 2}}
		node := &Amap_Obj_t{Type: AMAP_OBJ_AND, Id: uint32(i + 2), Best: Amap_Mat_t{pCut: cut, pSet: n.Best.pSet}}
		vec.Vec_PtrPush(p.vObjs, node)
		previous = node
	}
	if got := Amap_CutAreaDerefed2(p, previous, &previous.Best); got != 27 {
		t.Fatal("wrong source recursion boundary", got)
	}
	for _, o := range amapMatchObjects(p) {
		if o.nFouts != [2]int32{} {
			t.Fatal("deep temporary reference leaked", o.Id, o.nFouts)
		}
	}
	if got := Amap_CutAreaDerefed(p, previous, &previous.Best); got != 36 {
		t.Fatal("full exact area", got)
	}
}

func TestMatchComparisonPriorities(t *testing.T) {
	p, _, _, _ := matchFixture()
	fast := Amap_Mat_t{Delay: 1, Area: 10, AveFan: 1}
	small := Amap_Mat_t{Delay: 2, Area: 5, AveFan: 1}
	if Amap_CutCompareDelay(p, &fast, &small) != -1 || Amap_CutCompareArea(p, &fast, &small) != 1 {
		t.Fatal("primary priorities reversed")
	}
	tied := fast
	tied.Delay = 1.0001
	if Amap_CutCompareDelay(p, &fast, &tied) != -1 {
		t.Fatal("source near-tie preference changed")
	}
	tied = fast
	tied.AveFan = 3
	if Amap_CutCompareArea(p, &fast, &tied) != 1 {
		t.Fatal("larger fanout should win tied area")
	}
}
