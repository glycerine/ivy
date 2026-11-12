package abc

import (
	"math"
	"reflect"
	"testing"
)

func dfsFixture() (*Network, *Object, *Object, *Object, *Object, *Object) {
	n := NewNetwork("dfs")
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	x := n.CreateAndEdges(a.Signal(), b.Signal()).Object
	y := n.CreateAndEdges(x.Signal(), c.Signal()).Object
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, y)
	return n, a, b, c, x, y
}
func TestDFSRecursiveIterativeAndDangling(t *testing.T) {
	n, a, b, _, x, y := dfsFixture()
	dangling := n.CreateAndEdges(a.Signal(), b.Signal()).Object
	want := []*Object{x, y}
	if got := Abc_NtkDfs(n, 0); !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	if got := Abc_NtkDfsIter(n, 0); !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	want = append(want, dangling)
	if got := Abc_NtkDfs(n, 1); !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	if got := Abc_NtkDfsIter(n, 1); !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	if got := Abc_NtkDfsReverse(n); !reflect.DeepEqual(got, []*Object{y, x, dangling}) {
		t.Fatal(got)
	}
}
func TestDFSLevelsSupportAndInitialArrival(t *testing.T) {
	n, a, b, c, x, y := dfsFixture()
	if level := Abc_NtkLevel(n); level != 2 || x.Level != 1 || y.Level != 2 {
		t.Fatal(level, x.Level, y.Level)
	}
	if got := Abc_NtkNodeSupportInt(n, 0); !reflect.DeepEqual(got, []int32{0, 1, 2}) {
		t.Fatal(got)
	}
	if got := Abc_NtkNodeSupport(n, []*Object{y}, 1); !reflect.DeepEqual(got, []*Object{a, b, c}) {
		t.Fatal(got)
	}
	n.CILevel = func(o *Object) int {
		if o == a {
			return 3
		}
		return 0
	}
	if level := Abc_NtkLevel(n); level != 5 {
		t.Fatal(level)
	}
	if level := Abc_NtkLevelReverse(n); level != 2 || x.Level != 2 || y.Level != 1 {
		t.Fatal(level, x.Level, y.Level)
	}
	n.CILevel = nil
	Abc_NtkLevel(n)
	levels := Abc_DfsLevelized(a, 0)
	if len(levels) != 3 || !reflect.DeepEqual(levels[1], []*Object{x}) || !reflect.DeepEqual(levels[2], []*Object{y}) {
		t.Fatal(levels)
	}
}
func TestDFSContainedSupport(t *testing.T) {
	n, a, b, c, x, y := dfsFixture()
	levels := Abc_NtkDfsReverseNodesContained(n, []*Object{a, b}, 2)
	if levels[1] != x || levels[2] != nil {
		t.Fatal(levels)
	}
	levels = Abc_NtkDfsReverseNodesContained(n, []*Object{a, b, c}, 3)
	if levels[1] != x || levels[2] != y || levels[3] != n.cos[0] {
		t.Fatal(levels)
	}
}
func TestDFSContainedZeroDelayBarrier(t *testing.T) {
	n := NewNetwork("zero-delay")
	a := n.NewObject(ObjectPI)
	barrier := n.NewObject(ObjectNode)
	barrier.Barrier = true
	Abc_ObjAddFanin(barrier, a)
	node := n.CreateAndEdges(barrier.Signal(), barrier.Signal()).Object
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, node)
	levels := Abc_NtkDfsReverseNodesContained(n, []*Object{a}, 1)
	if levels[1] != node || levels[2] != po {
		t.Fatal("zero-level barrier stopped propagation", levels)
	}
}
func TestDFSCyclesChoicesAndEpochWrap(t *testing.T) {
	n, _, _, _, x, y := dfsFixture()
	if Abc_NtkIsAcyclic(n) != 1 {
		t.Fatal("acyclic graph rejected")
	}
	Abc_ObjAddFanin(x, y)
	n.travID = math.MaxUint64 - 1
	for _, o := range n.Objects() {
		o.travID = 0
	}
	if Abc_NtkIsAcyclic(n) != 0 {
		t.Fatal("cycle missed after epoch wrap")
	}
	n2, _, _, _, x2, y2 := dfsFixture()
	x2.Choice = y2
	if Abc_NtkIsAcyclic(n2) != 0 {
		t.Fatal("choice cycle missed")
	}
	if got := Abc_AigDfs(n2, 0, 0); len(got) != 2 {
		t.Fatal("choice cycle traversal repeated nodes", got)
	}
}
func TestDFSSequentialFeedback(t *testing.T) {
	n := NewNetwork("sequential")
	pi := n.NewObject(ObjectPI)
	bi, latch, bo := n.NewObject(ObjectBI), n.NewObject(ObjectLatch), n.NewObject(ObjectBO)
	node := n.CreateAndEdges(pi.Signal(), bo.Signal()).Object
	Abc_ObjAddFanin(bi, node)
	Abc_ObjAddFanin(latch, bi)
	Abc_ObjAddFanin(bo, latch)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, node)
	if Abc_NtkIsAcyclic(n) != 1 || Abc_NtkIsAcyclicWithBoxes(n) != 1 {
		t.Fatal("latch feedback treated as combinational cycle")
	}
	nodes := Abc_NtkDfsSeq(n)
	seen := map[*Object]bool{}
	for _, o := range nodes {
		if seen[o] {
			t.Fatal("repeated object")
		}
		seen[o] = true
	}
	if len(nodes) != 6 || !seen[latch] || !seen[bo] || !seen[bi] {
		t.Fatal(nodes)
	}
}
func TestDFSNetlistTraversal(t *testing.T) {
	n := NewNetwork("netlist")
	n.Type = uint8(ABC_NTK_NETLIST)
	pi := n.NewObject(ObjectPI)
	input := n.NewObject(ObjectNet)
	Abc_ObjAddFanin(input, pi)
	node := n.NewObject(ObjectNode)
	Abc_ObjAddFanin(node, input)
	output := n.NewObject(ObjectNet)
	Abc_ObjAddFanin(output, node)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, output)
	if got := Abc_NtkDfs(n, 0); !reflect.DeepEqual(got, []*Object{node}) {
		t.Fatal(got)
	}
	if got := Abc_NtkDfsReverse(n); !reflect.DeepEqual(got, []*Object{output}) {
		t.Fatal(got)
	}
	if got := Abc_NtkDfsNets(n); !reflect.DeepEqual(got, []*Object{input, output}) {
		t.Fatal(got)
	}
}
func TestDFSFunctionalIsoCommonInputs(t *testing.T) {
	n := NewNetwork("iso")
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	left := n.AndEdges(a.Signal(), b.Signal().Not())
	right := n.AndEdges(b.Signal(), c.Signal().Not())
	for _, edge := range []Edge{left, right} {
		po := n.NewObject(ObjectPO)
		Abc_ObjAddFanin(po, edge)
	}
	if Abc_NtkFunctionalIso(n, 0, 1, 0) != 1 {
		t.Fatal("renamed functions should be isomorphic")
	}
	if Abc_NtkFunctionalIso(n, 0, 1, 1) != 0 {
		t.Fatal("shared input identity ignored")
	}
	if Abc_NtkFunctionalIso(n, 0, 99, 0) != 0 {
		t.Fatal("invalid output accepted")
	}
}
func TestDFSFunctionalIsoSOPAndConstant(t *testing.T) {
	n := NewNetwork("sop")
	n.Type = uint8(ABC_NTK_LOGIC)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	x := n.NewObject(ObjectNode)
	x.Data = "11 1\n"
	Abc_ObjAddFanin(x, a)
	Abc_ObjAddFanin(x, b)
	y := Abc_NtkCreateNodeAnd(n, []*Object{a, b})
	for _, o := range []*Object{x, y} {
		po := n.NewObject(ObjectPO)
		Abc_ObjAddFanin(po, o)
	}
	if Abc_NtkFunctionalIso(n, 0, 1, 1) != 1 {
		t.Fatal("SOP/function equivalence failed")
	}
	constant := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(constant, n.One())
	if got := Abc_NtkSupport(n); len(got) != 2 {
		t.Fatal("global constant is not CI support", got)
	}
}
func TestDFSChoiceOrderClearsStaleScratch(t *testing.T) {
	n, a, b, _, x, y := dfsFixture()
	dangling := n.CreateAndEdges(a.Signal(), b.Signal()).Object
	for _, o := range n.Objects() {
		o.Temp = 999
	}
	order := Abc_AigGetLevelizedOrder(n, 0)
	if !reflect.DeepEqual(order, []*Object{x, y}) {
		t.Fatal(order)
	}
	if dangling.Temp != 0 {
		t.Fatal("stale dangling level", dangling.Temp)
	}
	order = Abc_AigGetLevelizedOrder(n, 1)
	if len(order) != len(n.cis)+3 {
		t.Fatal(order)
	}
	for _, o := range order {
		if o == nil {
			t.Fatal("nil prefix in levelized order")
		}
	}
}
