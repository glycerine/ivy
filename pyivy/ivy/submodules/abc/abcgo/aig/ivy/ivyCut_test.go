package ivy

import (
	"fmt"
	"slices"
	"testing"
)

func TestCutEnumerationTreeAndFiltering(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	ab := Ivy_And(p, a, b)
	root := Ivy_And(p, ab, c)
	Ivy_ObjCreatePo(p, root)
	cuts := Ivy_NodeFindCutsAll(p, root, 3)
	want := map[string]bool{fmt.Sprint([]int32{root.Id}): true, fmt.Sprint([]int32{c.Id, ab.Id}): true, fmt.Sprint([]int32{a.Id, b.Id, c.Id}): true}
	if int(cuts.nCuts) != len(want) {
		t.Fatalf("got %d cuts", cuts.nCuts)
	}
	for _, cut := range cuts.pCuts[:cuts.nCuts] {
		ids := cut.pArray[:cut.nSize]
		if !want[fmt.Sprint(ids)] {
			t.Fatalf("unexpected cut %v", ids)
		}
	}
	small := Ivy_NodeFindCutsAll(p, root, 2)
	if small.nCuts != 2 || cuts.nCuts != 3 {
		t.Fatal("cut limit or independent result ownership failed")
	}
}
func TestSequentialCutCrossesLatch(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	latch := Ivy_Latch(p, a, IVY_INIT_0)
	root := Ivy_And(p, latch, b)
	var front, inside []int32
	Ivy_ManSeqFindCut(p, root, &front, &inside, 2)
	if len(front) != 2 || !slices.Contains(front, Ivy_LeafCreate(a.Id, 1)) || !slices.Contains(front, Ivy_LeafCreate(b.Id, 0)) {
		t.Fatalf("sequential frontier %v", front)
	}
}
func TestBooleanCutMarksAndReconvergence(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c, d := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	tree := Ivy_And(p, Ivy_And(p, a, b), Ivy_And(p, c, d))
	root := Ivy_And(p, Ivy_And(p, a, b), Ivy_And(p, a, c))
	var front, volume, leaves []*Ivy_Obj_t
	if Ivy_ManFindBoolCut(p, tree, &front, &volume, &leaves) != 0 {
		t.Fatal("tree has no reconvergence pivot")
	}
	for _, n := range p.vObjs {
		if n != nil && (n.fMarkA != 0 || n.fMarkB != 0) {
			t.Fatal("failed Boolean cut leaked marks")
		}
	}
	if Ivy_ManFindBoolCut(p, root, &front, &volume, &leaves) != 1 {
		t.Fatal("reconvergence cut not found")
	}
	if len(leaves) != 3 || !slices.Contains(leaves, a) || !slices.Contains(leaves, b) || !slices.Contains(leaves, c) {
		t.Fatal("wrong Boolean cut leaves")
	}
	for _, n := range p.vObjs {
		if n != nil && (n.fMarkA != 0 || n.fMarkB != 0) {
			t.Fatal("successful Boolean cut leaked marks")
		}
	}
}
func TestCutExpansionHandlesBufferAndDuplicateFanin(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	buf := Ivy_ObjCreate(p, Ivy_ObjCreateGhost(p, a, nil, IVY_BUF, IVY_INIT_NONE))
	root := Ivy_And(p, buf, b)
	cuts := Ivy_NodeFindCutsAll(p, root, 2)
	if cuts.nCuts < 2 {
		t.Fatal("buffered cut not expanded")
	}
	c := Ivy_Cut_t{nSize: 2, nSizeMax: 2}
	c.pArray[0], c.pArray[1] = a.Id, b.Id
	var next Ivy_Cut_t
	if Ivy_NodeCutDeriveNew(&c, &next, b.Id, a.Id, a.Id) == 0 || next.nSize != 1 || next.pArray[0] != a.Id {
		t.Fatal("duplicate replacement leaves not merged")
	}
}
func TestMuxRecognitionAllPolarities(t *testing.T) {
	for polarity := 0; polarity < 8; polarity++ {
		p := Ivy_ManStart()
		a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
		root := Ivy_Regular(Ivy_Mux(p, Ivy_NotCond(a, int32(polarity&1)), Ivy_NotCond(b, int32(polarity&2)), Ivy_NotCond(c, int32(polarity&4))))
		if Ivy_ObjIsMuxType(root) == 0 {
			t.Fatal("structural MUX not detected")
		}
		var then, otherwise *Ivy_Obj_t
		control := Ivy_ObjRecognizeMux(root, &then, &otherwise)
		if Ivy_IsComplement(control) != 0 {
			t.Fatal("complemented control")
		}
		rebuilt := Ivy_Mux(p, control, then, otherwise)
		Ivy_ObjCreatePo(p, root)
		Ivy_ObjCreatePo(p, rebuilt)
		for bits := 0; bits < 8; bits++ {
			v := ivyBalanceEvaluate(p, bits)
			if v[0] != v[1] {
				t.Fatalf("MUX polarity %d assignment %d", polarity, bits)
			}
		}
	}
}
