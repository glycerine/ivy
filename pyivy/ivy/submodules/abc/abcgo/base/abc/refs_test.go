package abc

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestTemporaryFanoutAdjustmentPreservesGraph(t *testing.T) {
	n := NewNetwork("fanout-adjustment")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	root := n.CreateAndEdges(a.Signal(), b.Signal())
	if got := Abc_ObjFanoutNum(a); got != 1 {
		t.Fatal(got)
	}
	Abc_ObjAdjustFanoutCount(a, -1)
	if got := Abc_ObjFanoutNum(a); got != 0 {
		t.Fatalf("adjusted fanout count = %d", got)
	}
	if got := vec.Vec_IntEntry(Abc_NtkFanoutCounts(n), int(a.ID)); got != 0 {
		t.Fatalf("fanout vector count = %d", got)
	}
	if len(a.fanouts) != 1 || a.fanouts[0] != root.Object {
		t.Fatal("temporary adjustment changed graph adjacency")
	}
	Abc_ObjAdjustFanoutCount(a, 1)
	Abc_NtkFinishFanoutCountAdjustments(n)
	if n.refs != nil || Abc_ObjFanoutNum(a) != 1 {
		t.Fatal("temporary counts were not restored")
	}
}

func TestMffcReferenceAccounting(t *testing.T) {
	n := NewNetwork("mffc")
	a, b, c, d := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	shared := n.CreateAndEdges(a.Signal(), b.Signal())
	left := n.CreateAndEdges(shared, c.Signal())
	right := n.CreateAndEdges(shared, d.Signal())
	root := n.CreateAndEdges(left, right)
	if got := Abc_NodeMffcSize(root.Object); got != 4 {
		t.Fatal(got)
	}
	if n.refs != nil {
		t.Fatal("references not restored")
	}
	before := len(shared.Object.fanouts)
	if got := Abc_NodeDeref_rec(root.Object); got != 4 {
		t.Fatal(got)
	}
	if Abc_ObjFanoutNum(shared.Object) != 0 {
		t.Fatal("virtual fanout count not exposed")
	}
	if len(shared.Object.fanouts) != before {
		t.Fatal("graph topology modified")
	}
	var cone, support []*Object
	Abc_NodeMffcConeSupp(root.Object, &cone, &support)
	if len(cone) != 4 || len(support) != 4 {
		t.Fatal(len(cone), len(support))
	}
	if got := Abc_NodeRef_rec(root.Object); got != 4 {
		t.Fatal(got)
	}
	if n.refs != nil {
		t.Fatal("scratch refs retained")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestMffcExternalFanoutAndCut(t *testing.T) {
	n := NewNetwork("cut")
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	child := n.CreateAndEdges(a.Signal(), b.Signal())
	root := n.CreateAndEdges(child, c.Signal())
	var inside []*Object
	if count := Abc_NodeMffcInside(root.Object, []*Object{child.Object}, &inside); count != 1 || len(inside) != 1 || inside[0] != root.Object {
		t.Fatal(count, inside)
	}
	if count := Abc_NodeMffcSize(root.Object); count != 2 {
		t.Fatal(count)
	}
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, child)
	if count := Abc_NodeMffcSize(root.Object); count != 1 {
		t.Fatal(count)
	}
	if n.refs != nil {
		t.Fatal("cut refs not restored")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestMffcComplementStopAndLabels(t *testing.T) {
	n := NewNetwork("stop")
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	child := n.CreateAndEdges(a.Signal(), b.Signal())
	root := n.CreateAndEdges(child.Not(), c.Signal())
	if got := Abc_NodeMffcSizeStop(root.Object); got != 1 {
		t.Fatal(got)
	}
	if got := Abc_NodeMffcSize(root.Object); got != 2 {
		t.Fatal(got)
	}
	var nodes []*Object
	if got := Abc_NodeMffcLabel(root.Object, &nodes); got != 2 || len(nodes) != 2 {
		t.Fatal(got, nodes)
	}
	if Abc_NodeIsTravIdCurrent(root.Object) == 0 || Abc_NodeIsTravIdCurrent(child.Object) == 0 || Abc_NodeIsTravIdCurrent(a) != 0 {
		t.Fatal("label scope")
	}
	Abc_NtkIncrementTravId(n)
	if got := Abc_NodeMffcLabelAig(root.Object); got != 2 || Abc_NodeIsTravIdCurrent(a) == 0 {
		t.Fatal("dereference label semantics", got)
	}
}
func TestMffcDuplicateInputsAndConstants(t *testing.T) {
	n := NewNetwork("duplicates")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	child := n.CreateAndEdges(a.Signal(), b.Signal())
	root := n.CreateAndEdges(child, child)
	if got := Abc_NodeMffcSize(root.Object); got != 2 {
		t.Fatal(got)
	}
	if n.refs != nil {
		t.Fatal("duplicate refs unbalanced")
	}
	root2 := n.CreateAndEdges(n.One(), a.Signal())
	if got := Abc_NodeMffcSize(root2.Object); got != 1 {
		t.Fatal(got)
	}
	constant := Abc_NtkCreateNodeConst0(n)
	if got := Abc_NodeMffcSize(constant); got != 0 {
		t.Fatal(got)
	}
}

func TestMffcForeignLeafDoesNotChangeReferences(t *testing.T) {
	n := NewNetwork("invalid-cut")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	root := n.CreateAndEdges(a.Signal(), b.Signal())
	foreign := NewNetwork("foreign").NewObject(ObjectPI)
	defer func() {
		if recover() == nil {
			t.Error("accepted foreign cut leaf")
		}
		if n.refs != nil {
			t.Error("invalid cut changed reference state")
		}
	}()
	var inside []*Object
	Abc_NodeMffcInside(root.Object, []*Object{a, foreign}, &inside)
}

func TestTraversalInitializationAndWrap(t *testing.T) {
	n := NewNetwork("trav")
	a := n.NewObject(ObjectPI)
	if Abc_NodeIsTravIdCurrent(a) != 0 {
		t.Fatal("unvisited node initially current")
	}
	Abc_NodeSetTravIdCurrent(a)
	if Abc_NodeIsTravIdCurrent(a) == 0 {
		t.Fatal("mark")
	}
	n.travID = ^uint64(0)
	a.travID = 1
	Abc_NtkIncrementTravId(n)
	if n.travID != 1 || Abc_NodeIsTravIdCurrent(a) != 0 {
		t.Fatal("overflow reused stale marker")
	}
}
func TestPreviousTraversalEpochAPI(t *testing.T) {
	n := NewNetwork("previous")
	o := n.NewObject(ObjectPI)
	Abc_NtkIncrementTravId(n)
	if Abc_NodeIsTravIdPrevious(o) != 0 {
		t.Fatal("unvisited epoch mistaken for previous")
	}
	Abc_NodeSetTravIdPrevious(o)
	if Abc_NodeIsTravIdPrevious(o) != 1 || Abc_NodeIsTravIdCurrent(o) != 0 {
		t.Fatal("previous marking failed")
	}
	Abc_NtkIncrementTravId(n)
	if Abc_NodeIsTravIdPrevious(o) != 0 {
		t.Fatal("stale previous epoch retained")
	}
	Abc_NodeSetTravIdCurrent(o)
	if Abc_NodeIsTravIdCurrent(o) != 1 || Abc_NodeIsTravIdPrevious(o) != 0 {
		t.Fatal("current/previous distinction failed")
	}
}
