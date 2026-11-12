package abc

import (
	"reflect"
	"testing"
)

func TestNodeConeCollect(t *testing.T) {
	n := NewNetwork("cone")
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	shared := n.AndEdges(a.Signal(), b.Signal()).Object
	left := n.AndEdges(shared.Signal(), c.Signal()).Object
	root := n.AndEdges(shared.Signal(), left.Signal()).Object
	shared.MarkA = true
	if got, want := Abc_NodeConeCollect([]*Object{root}, []*Object{shared, c}, false), []*Object{left, root}; !reflect.DeepEqual(got, want) {
		t.Fatalf("cone: got %v, want %v", got, want)
	}
	if got, want := Abc_NodeConeCollect([]*Object{root}, []*Object{shared, c}, true), []*Object{shared, c, left, root}; !reflect.DeepEqual(got, want) {
		t.Fatalf("cone with leaves: got %v, want %v", got, want)
	}
	if !shared.MarkA {
		t.Fatal("cone traversal changed caller's marks")
	}
}
