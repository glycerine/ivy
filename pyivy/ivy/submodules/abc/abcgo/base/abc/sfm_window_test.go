package abc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSfmWindowTraversal(t *testing.T) {
	n := NewNetwork("test")
	a := n.NewObject(ObjectPI)
	b := n.NewObject(ObjectPI)
	x := n.NewObject(ObjectNode)
	Abc_ObjAddFanin(x, a)
	Abc_ObjAddFanin(x, b)
	x.Level = 1
	y := n.NewObject(ObjectNode)
	Abc_ObjAddFanin(y, x)
	y.Level = 2
	out := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(out, y)
	Abc_NtkIncrementTravId(n)
	v := vec.Vec_IntAlloc(0)
	calls := 0
	label := Abc_NtkDfsOne(y, v, 0, 1, func(*Object) { calls++ })
	if label != 1 || vec.Vec_IntSize(v) != 4 || calls != 2 {
		t.Fatalf("label %d count %d calls %d", label, vec.Vec_IntSize(v), calls)
	}
	Abc_NtkIncrementTravId(n)
	if Abc_NtkDfsCheck(y, x) {
		t.Fatal("missed pivot dependency")
	}
	Abc_NtkIncrementTravId(n)
	vec.Vec_IntClear(v)
	Abc_NtkDfsReverseOne(x, v, 3, 3)
	if vec.Vec_IntSize(v) != 2 {
		t.Fatalf("fanout cone %v", vec.Vec_IntArray(v))
	}
}
