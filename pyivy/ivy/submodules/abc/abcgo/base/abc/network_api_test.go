package abc

import (
	"reflect"
	"testing"
)

func TestLatchInitializationAPIs(t *testing.T) {
	n := NewNetwork("latches")
	latch := n.NewObject(ObjectLatch)
	if Abc_LatchIsInitNone(latch) != 1 {
		t.Fatal("fresh latch is not uninitialized")
	}
	setters := []func(*Object){Abc_LatchSetInitNone, Abc_LatchSetInit0, Abc_LatchSetInit1, Abc_LatchSetInitDc}
	predicates := []func(*Object) int32{Abc_LatchIsInitNone, Abc_LatchIsInit0, Abc_LatchIsInit1, Abc_LatchIsInitDc}
	for i, set := range setters {
		set(latch)
		for j, predicate := range predicates {
			if (predicate(latch) != 0) != (i == j) {
				t.Fatal(i, j, latch.Data)
			}
		}
	}
	if Abc_NtkHasOnlyLatchBoxes(n) != 1 || Abc_NtkIsComb(n) != 0 {
		t.Fatal("latch network classification")
	}
	n.NewObject(ObjectBlackbox)
	if Abc_NtkHasOnlyLatchBoxes(n) != 0 {
		t.Fatal("mixed boxes accepted")
	}
}
func TestBoxPermutationUpdatesAllTerminalOrders(t *testing.T) {
	n := NewNetwork("box order")
	pi := n.NewObject(ObjectPI)
	bi1, box1, bo1 := newBarrier(n, ABC_INIT_ZERO)
	bi2, box2, bo2 := newBarrier(n, ABC_INIT_ONE)
	Abc_ObjAddFanin(bi1, pi)
	Abc_ObjAddFanin(bi2, pi)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, bo2)
	if err := n.SetBoxOrder([]int{box2.ID, box1.ID}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(n.Inputs(), []*Object{pi, bo2, bo1}) || !reflect.DeepEqual(n.Outputs(), []*Object{po, bi2, bi1}) || !reflect.DeepEqual(n.ObjectsOfKind(ObjectBO), []*Object{bo2, bo1}) {
		t.Fatal("box and terminal order diverged")
	}
	before := n.Boxes()
	if err := n.SetBoxOrder([]int{box2.ID, box2.ID}); err == nil || !reflect.DeepEqual(before, n.Boxes()) {
		t.Fatal("invalid permutation was not atomic")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestSATModelSnapshots(t *testing.T) {
	n := NewNetwork("model")
	values := []int32{1, 0, -1}
	Abc_NtkSetModel(n, values)
	values[0] = 0
	copy := Abc_NtkReadModel(n)
	copy[1] = 1
	if !reflect.DeepEqual(n.Model, []int32{1, 0, -1}) {
		t.Fatal("SAT model aliased caller storage")
	}
}
