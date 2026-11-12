package abc

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestReverseTopoContainsFanoutCone(t *testing.T) {
	n := NewNetwork("topology")
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	root := n.AndEdges(a.Signal(), b.Signal()).Object
	left := n.AndEdges(root.Signal(), c.Signal()).Object
	right := n.AndEdges(root.Signal(), c.Signal().Not()).Object
	for _, node := range []*Object{left, right} {
		Abc_ObjAddFanin(n.NewObject(ObjectPO), node)
	}
	Abc_NtkReverseTopoOrder(n)
	begin, end := Abc_NtkTopoHasBeg(root), Abc_NtkTopoHasEnd(root)
	if begin == 0 || end <= begin || int(end) > len(n.Topo) {
		t.Fatalf("invalid root interval %d:%d", begin, end)
	}
	found := make(map[int32]bool)
	for _, id := range n.Topo[begin:end] {
		found[id] = true
	}
	for _, node := range []*Object{left, right} {
		if !found[int32(node.ID)] {
			t.Fatalf("fanout %d missing from interval", node.ID)
		}
	}
	visited := vec.Vec_IntAlloc(8)
	Abc_NtkIncrementTravId(n)
	Abc_NtkReverse_rec(root, visited)
	if got := vec.Vec_IntArray(visited); len(got) != 5 || got[len(got)-1] != int32(root.ID) {
		t.Fatalf("reverse traversal = %v", got)
	}
}
