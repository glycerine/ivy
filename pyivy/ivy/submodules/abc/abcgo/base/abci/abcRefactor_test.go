package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestNodeConeTruthTracksAIGPhases(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	ab := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(ab, a)
	abc.Abc_ObjAddFanin(ab, b)
	nab := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(nab, a)
	abc.Abc_ObjAddFanin(nab, b)
	abc.Abc_ObjSetFaninC(nab, 0)
	abc.Abc_ObjSetFaninC(nab, 1)
	root := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(root, ab)
	abc.Abc_ObjAddFanin(root, nab)
	abc.Abc_ObjSetFaninC(root, 0)
	abc.Abc_ObjSetFaninC(root, 1)

	manager := Abc_NtkManRefStart(2, 2, 0, 0)
	defer Abc_NtkManRefStop(manager)
	leaves, visited := vec.Vec_PtrAlloc(2), vec.Vec_PtrAlloc(8)
	defer vec.Vec_PtrFree(leaves)
	defer vec.Vec_PtrFree(visited)
	vec.Vec_PtrPush(leaves, a)
	vec.Vec_PtrPush(leaves, b)
	root.MarkA, ab.MarkA = true, true
	truth := Abc_NodeConeTruth(manager.vVars, manager.vFuncs, int32(refTruthWords(2)), root, leaves, visited)
	if len(truth) != 1 || truth[0] != 0x6666666666666666 {
		t.Fatalf("XOR truth = %#v, want 0x6666666666666666", truth)
	}
	if !root.MarkA || !ab.MarkA {
		t.Fatal("cone truth collection changed preexisting MarkA state")
	}
}

func TestKitToDecRefactorGraphPreservesTruth(t *testing.T) {
	kitGraph := kit.Kit_TruthToGraph([]uint32{0x88888888}, 3, nil)
	defer kit.Kit_GraphFree(kitGraph)
	graph := refKitToDec(kitGraph)
	defer dec.Dec_GraphFree(graph)
	if got := dec.Dec_GraphDeriveTruth(graph); got != 0x88888888 {
		t.Fatalf("converted decomposition truth = %#08x, want %#08x", got, uint32(0x88888888))
	}
}

func TestRefactorManagerStopReleasesScratch(t *testing.T) {
	manager := Abc_NtkManRefStart(8, 12, 1, 0)
	if manager.vVars == nil || manager.vFuncs == nil || manager.vVisited == nil || manager.vCube == nil {
		t.Fatal("manager did not initialize scratch storage")
	}
	Abc_NtkManRefStop(manager)
	if manager.vVars != nil || manager.vFuncs != nil || manager.vVisited != nil || manager.vCube != nil || manager.vMemory != nil {
		t.Fatal("manager stop retained scratch storage")
	}
}

func TestRefactorLabelsOnlyDeletableNodes(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	defer abc.Abc_NtkDelete(n)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	x := abc.Abc_AigAnd(manager, a, b)
	y := abc.Abc_AigAnd(manager, x, c)
	root := abc.Abc_AigAnd(manager, y, x).Object
	abc.Abc_ObjAddFanin(abc.Abc_NtkCreatePo(n), root)
	ref := Abc_NtkManRefStart(3, 3, 0, 0)
	defer Abc_NtkManRefStop(ref)
	leaves := vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(leaves, x.Object)
	vec.Vec_PtrPush(leaves, c)
	abc.Abc_NtkIncrementTravId(n)
	abc.Abc_NodeSetTravIdCurrent(a)
	graph := Abc_NodeRefactor(ref, root, leaves, 0, 0, 0, 0)
	if graph == nil || ref.nLastGain != 1 {
		t.Fatalf("expected redundant AND to save one node, gain %d, graph %v", ref.nLastGain, graph)
	}
	defer dec.Dec_GraphFree(graph)
	if abc.Abc_NodeIsTravIdCurrent(x.Object) != 0 || abc.Abc_NodeIsTravIdCurrent(c) != 0 {
		t.Error("marked a pinned cut input as deletable")
	}
	if abc.Abc_NodeIsTravIdCurrent(root) == 0 || abc.Abc_NodeIsTravIdCurrent(y.Object) == 0 {
		t.Error("missing internal deletion-cone labels")
	}
	if abc.Abc_NodeIsTravIdPrevious(a) == 0 {
		t.Error("refactoring advanced more than one traversal epoch")
	}
	if root.MarkA || y.Object.MarkA {
		t.Error("refactoring left temporary marks on surviving nodes")
	}
	if abc.Abc_ObjFanoutNum(x.Object) != 2 || abc.Abc_ObjFanoutNum(c) != 1 {
		t.Error("boundary references changed")
	}
}

func TestConstantRefactorCountsFullMFFC(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	defer abc.Abc_NtkDelete(n)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	m := n.FunctionManager.(*abc.Abc_Aig_t)
	x := abc.Abc_AigAnd(m, a, b)
	y := abc.Abc_AigAnd(m, x, c)
	root := abc.Abc_AigAnd(m, y, x.Not()).Object
	abc.Abc_ObjAddFanin(abc.Abc_NtkCreatePo(n), root)
	leaves := vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(leaves, x.Object)
	vec.Vec_PtrPush(leaves, c)
	ref := Abc_NtkManRefStart(2, 2, 0, 0)
	defer Abc_NtkManRefStop(ref)
	abc.Abc_NtkIncrementTravId(n)
	abc.Abc_NodeSetTravIdCurrent(a)
	graph := Abc_NodeRefactor(ref, root, leaves, 0, 0, 0, 0)
	if graph == nil || dec.Dec_GraphIsConst0(graph) == 0 {
		t.Fatal("expected constant zero")
	}
	defer dec.Dec_GraphFree(graph)
	if ref.nLastGain != 3 || ref.nNodesGained != 3 {
		t.Errorf("constant substitution gained %d nodes, total %d, want all three ANDs", ref.nLastGain, ref.nNodesGained)
	}
	if abc.Abc_NodeIsTravIdCurrent(a) == 0 {
		t.Error("constant substitution changed the traversal epoch")
	}
}

func BenchmarkRefactorTruthSmallCone(b *testing.B) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	defer abc.Abc_NtkDelete(n)
	a, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	for range 4096 {
		abc.Abc_NtkCreatePi(n)
	}
	m := n.FunctionManager.(*abc.Abc_Aig_t)
	left := abc.Abc_AigAnd(m, a, c)
	right := abc.Abc_AigAnd(m, a.Signal().Not(), c.Signal().Not())
	root := abc.Abc_AigAnd(m, left.Not(), right.Not())
	abc.Abc_ObjAddFanin(abc.Abc_NtkCreatePo(n), root)
	ref := Abc_NtkManRefStart(2, 2, 0, 0)
	defer Abc_NtkManRefStop(ref)
	leaves := vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(leaves, a)
	vec.Vec_PtrPush(leaves, c)
	b.ReportAllocs()
	for b.Loop() {
		truth := Abc_NodeConeTruth(ref.vVars, ref.vFuncs, 1, root.Object, leaves, ref.vVisited)
		if len(truth) != 1 || truth[0] != 0x6666666666666666 || vec.Vec_PtrSize(ref.vVisited) != 3 {
			b.Fatal("invalid cone truth")
		}
	}
}

func TestRefactorTruthMultiwordSharedCone(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	defer abc.Abc_NtkDelete(n)
	ref := Abc_NtkManRefStart(7, 7, 0, 0)
	defer Abc_NtkManRefStop(ref)
	m := n.FunctionManager.(*abc.Abc_Aig_t)
	leaves := vec.Vec_PtrAlloc(7)
	var inputs [7]abc.Edge
	for i := range inputs {
		inputs[i] = abc.Abc_NtkCreatePi(n).Signal()
		vec.Vec_PtrPush(leaves, inputs[i].Object)
	}
	shared := abc.Abc_AigAnd(m, inputs[0].Not(), inputs[6])
	left := abc.Abc_AigAnd(m, shared, inputs[2])
	right := abc.Abc_AigAnd(m, shared, inputs[3].Not())
	root := abc.Abc_AigAnd(m, left.Not(), right.Not())
	inputs[6].Object.MarkA = true
	defer func() { inputs[6].Object.MarkA = false }()
	vec.Vec_PtrPush(ref.vVisited, inputs[4].Object)
	truth := Abc_NodeConeTruth(ref.vVars, ref.vFuncs, 2, root.Object, leaves, ref.vVisited)
	if len(truth) != 2 || vec.Vec_PtrSize(ref.vVisited) != 4 {
		t.Fatalf("got truth %v and %d nodes", truth, vec.Vec_PtrSize(ref.vVisited))
	}
	for i, e := range []abc.Edge{shared, left, right, root} {
		if got := vec.Vec_PtrEntry(ref.vVisited, i); got != e.Object {
			t.Errorf("visited node %d got %v, want %v", i, got, e.Object)
		}
	}
	for row := range 128 {
		v := row&1 == 0 && row&64 != 0
		want := !(v && row&4 != 0) && !(v && row&8 == 0)
		if got := truth[row/64]&(uint64(1)<<uint(row%64)) != 0; got != want {
			t.Fatalf("assignment %d got %t, want %t", row, got, want)
		}
	}
	if !inputs[6].Object.MarkA {
		t.Error("lost cut input mark")
	}
}

func TestRefactorConversionPreservesORAnnotationsAndSelectedRoot(t *testing.T) {
	g := kit.Kit_GraphCreate(3)
	a, b, c := kit.Kit_EdgeCreate(0, 1), kit.Kit_EdgeCreate(1, 0), kit.Kit_EdgeCreate(2, 1)
	or := kit.Kit_GraphAddNodeOr(g, a, b)
	last := kit.Kit_GraphAddNodeAnd(g, or, c)
	for _, root := range []kit.Kit_Edge_t{last, or, c} {
		kit.Kit_GraphSetRoot(g, root)
		out := refKitToDec(g)
		if got := dec.Dec_GraphNodeNum(out); got != 2 {
			t.Errorf("root %d: converted %d gates, want both original gates", root.Node, got)
			continue
		}
		if got := dec.Dec_GraphNode(out, int32(or.Node)).PrintingFlags(); got != [3]uint32{1, 1, 0} {
			t.Errorf("OR with complemented first operand has printing flags %v", got)
		}
		if got := dec.Dec_GraphNode(out, int32(last.Node)).PrintingFlags(); got != [3]uint32{0, 1, 1} {
			t.Errorf("AND with two complemented operands has printing flags %v", got)
		}
		if got := dec.Dec_EdgeToInt(dec.Dec_GraphRootEdge(out)); got != kit.Kit_EdgeToInt(root) {
			t.Errorf("selected root word %d became %d", kit.Kit_EdgeToInt(root), got)
		}
		if got, want := dec.Dec_GraphDeriveTruth(out), kit.Kit_GraphToTruth(g); got != want {
			t.Errorf("root %d truth %08x, want %08x", root.Node, got, want)
		}
	}
}

func BenchmarkRefactorGraphConversion(b *testing.B) {
	g := kit.Kit_GraphCreate(4)
	root := kit.Kit_EdgeCreate(0, 0)
	for i := range 7 {
		root = kit.Kit_GraphAddNodeOr(g, root, kit.Kit_EdgeCreate(int32(1+i%3), int32(i%2)))
	}
	kit.Kit_GraphSetRoot(g, root)
	b.ReportAllocs()
	for b.Loop() {
		out := refKitToDec(g)
		if dec.Dec_GraphNodeNum(out) != 7 {
			b.Fatal("expected seven gates")
		}
		dec.Dec_GraphFree(out)
	}
}
