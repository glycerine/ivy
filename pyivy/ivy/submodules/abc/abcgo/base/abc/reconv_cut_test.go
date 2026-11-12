package abc

import (
	"reflect"
	"testing"
)

func TestReconvergenceCutAndContainingCut(t *testing.T) {
	n := NewNetwork("cut")
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	shared := n.AndEdges(a.Signal(), b.Signal()).Object
	left := n.AndEdges(shared.Signal(), c.Signal()).Object
	right := n.AndEdges(shared.Signal(), c.Signal().Not()).Object
	root := n.AndEdges(left.Signal(), right.Signal()).Object
	root.MarkB = true
	m := Abc_NtkManCutStart(2, 4, 2, 100)
	cut := Abc_NodeFindCut(m, root, 1)
	if !reflect.DeepEqual(cut, []*Object{left, right}) {
		t.Fatalf("small cut = %v", cut)
	}
	large := Abc_NtkManCutReadCutLarge(m)
	if !reflect.DeepEqual(large, []*Object{c, a, b}) {
		t.Fatalf("large cut = %v", large)
	}
	if !root.MarkB || left.MarkB || right.MarkB || shared.MarkB {
		t.Fatal("cut computation changed object marks")
	}
	if got := Abc_NodeConeCollect([]*Object{root}, large, false); len(got) != 4 || got[len(got)-1] != root {
		t.Fatalf("wrong containing cone: %v", got)
	}
}

func TestCutTfoCandidatesExcludeRootCone(t *testing.T) {
	n := NewNetwork("tfo")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	root := n.AndEdges(a.Signal(), b.Signal()).Object
	candidate := n.AndEdges(a.Signal().Not(), b.Signal()).Object
	Abc_ObjAddFanin(n.NewObject(ObjectPO), root)
	Abc_ObjAddFanin(n.NewObject(ObjectPO), candidate)
	m := Abc_NtkManCutStart(2, 3, 2, 100)
	got := Abc_NodeCollectTfoCands(m, root, []*Object{a, b}, 10)
	if !reflect.DeepEqual(got, []*Object{candidate}) {
		t.Fatalf("TFO candidates = %v", got)
	}
	if got := Abc_NodeCollectTfoCands(m, root, []*Object{a, b}, 0); len(got) != 0 {
		t.Fatalf("level limit should exclude candidate: %v", got)
	}
}
