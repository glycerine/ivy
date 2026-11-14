package aig

import (
	"reflect"
	"testing"
)

func TestSupportCollectorRespectsCurrentFrontier(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	x := p.AND(a, b)
	root := p.AND(x, c)
	Aig_ManIncrementTravId(p)
	for _, e := range []Edge{a, x, root} {
		Aig_ObjSetTravIdCurrent(p, e)
	}
	got := Aig_SuppMinCollectSupport(p, []Edge{x, root})
	if !reflect.DeepEqual(got, []Edge{b, c}) {
		t.Fatalf("support = %v, want unmarked frontier %v", got, []Edge{b, c})
	}
	for _, e := range got {
		if !Aig_ObjIsTravIdCurrent(p, e) {
			t.Error("collected support missing source traversal mark")
		}
	}
}

func TestSupportReconstructionLeavesSourceCopies(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	x := p.AND(a, b)
	root := p.AND(x, c)
	unrelated := p.NewInput()
	Aig_ObjSetCopy(p, unrelated, c.Not())
	result := Aig_SuppMinReconstruct(p, []Edge{a}, []Edge{x, root}, []Edge{a, b, c})
	for e, want := range map[Edge]Edge{a: p.Const1(), b: b, c: c, x: b, root: result, unrelated: c.Not()} {
		if got := Aig_ObjCopy(p, e); got != want {
			t.Errorf("copy of %d = %d, want %d", e, got, want)
		}
	}
	if result != p.AND(b, c) {
		t.Fatal("incorrect reconstructed cone")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
