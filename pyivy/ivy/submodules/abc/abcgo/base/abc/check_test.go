package abc

import (
	"testing"

	"github.com/glycerine/abcgo/misc/nm"
)

func validCheckedLogic() *Network {
	n := NewNetwork("checked")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_SOP)
	pi := n.NewObject(ObjectPI)
	pi.SetName("input")
	node := Abc_NtkCreateNodeBuf(n, pi)
	po := n.NewObject(ObjectPO)
	po.SetName("output")
	Abc_ObjAddFanin(po, node)
	return n
}
func TestOriginalNetworkChecksPreserveCopies(t *testing.T) {
	n := validCheckedLogic()
	sentinel := n.NewObject(ObjectNode)
	sentinel.Data = func([]bool) bool { return true }
	pi := n.ObjectsOfKind(ObjectPI)[0]
	pi.Copy = sentinel
	if Abc_NtkDoCheck(n) != 1 || Abc_NtkCheckRead(n) != 1 || pi.Copy != sentinel {
		t.Fatal("valid graph or scratch preservation")
	}
	pi.Data = "illegal PI data"
	if Abc_NtkCheckPis(n) != 0 {
		t.Fatal("PI function accepted")
	}
	pi.Data = nil
	po := n.ObjectsOfKind(ObjectPO)[0]
	po.SetName("input")
	if Abc_NtkCheckNames(n) != 0 {
		t.Fatal("nontrivial shared CI/CO name accepted")
	}
	po.SetName("output")
	n.cis = append(n.cis, pi)
	if Abc_NtkDoCheck(n) != 0 {
		t.Fatal("duplicate CI accepted")
	}
}
func TestNetlistAndMVChecks(t *testing.T) {
	n := makeMVPermutation(t)
	if Abc_NtkDoCheck(n) != 1 {
		t.Fatal("valid MV netlist rejected")
	}
	net := n.ObjectsOfKind(ObjectNet)[0]
	if Abc_NtkCheckNet(n, net) != 1 {
		t.Fatal("driven net rejected")
	}
	Abc_ObjRemoveFanins(net)
	if Abc_NtkCheckNet(n, net) != 0 {
		t.Fatal("undriven net accepted")
	}
}
func TestCheckedLatchesAndNamedComparison(t *testing.T) {
	n := validCheckedLogic()
	pi := n.ObjectsOfKind(ObjectPI)[0]
	input, latch, output := newBarrier(n, ABC_INIT_ZERO)
	input.SetName("next")
	output.SetName("state")
	Abc_ObjAddFanin(input, pi)
	if Abc_NtkCheckLatch(n, latch) != 1 || Abc_NtkDoCheck(n) != 1 {
		t.Fatal("valid latch rejected")
	}
	latch.Data = ABC_INIT_NONE
	if Abc_NtkCheckLatch(n, latch) != 0 {
		t.Fatal("unknown reset accepted")
	}
	a, b := validCheckedLogic(), validCheckedLogic()
	if Abc_NtkCompareSignals(a, b, 0, 0) != 1 {
		t.Fatal("equivalent interface comparison")
	}
	b.ObjectsOfKind(ObjectPI)[0].SetName("different")
	if Abc_NtkCompareSignals(a, b, 1, 0) != 0 {
		t.Fatal("different input names accepted")
	}
}
func TestChecksRejectMalformedOrderedListsWithoutPanic(t *testing.T) {
	n := validCheckedLogic()
	n.cis = append(n.cis, nil)
	if Abc_NtkDoCheck(n) != 0 {
		t.Fatal("nil interface object accepted")
	}
}
func TestCheckAcceptsSpecialXorSop(t *testing.T) {
	n := validCheckedLogic()
	node := n.ObjectsOfKind(ObjectNode)[0]
	node.Data = "1 x\n"
	if Abc_NtkCheckNode(n, node) != 1 {
		t.Fatal("ABC special XOR SOP rejected")
	}
}
func TestHierarchyCycleCheckAllowsRepeatedAcyclicModels(t *testing.T) {
	root := testHierarchicalDesign(t, false)
	leaf := root.Design.Module("leaf")
	pis := root.ObjectsOfKind(ObjectPI)
	testInstance(root, leaf, "other", pis[0].fanouts[0], pis[1].fanouts[0])
	if Abc_NtkIsAcyclicHierarchy(root) != 1 {
		t.Fatal("DAG reuse mistaken for recursion")
	}
	testInstance(leaf, root, "cycle", leaf.ObjectsOfKind(ObjectPI)[0].fanouts[0], leaf.ObjectsOfKind(ObjectPI)[1].fanouts[0])
	if Abc_NtkIsAcyclicHierarchy(root) != 0 {
		t.Fatal("recursive model accepted")
	}
}

func TestSignalComparisonOrdersLatchesByOutputNames(t *testing.T) {
	makeNetwork := func(reverse bool) *Network {
		n := validCheckedLogic()
		pi := n.ObjectsOfKind(ObjectPI)[0]
		names := []string{"alpha", "beta"}
		if reverse {
			names = []string{"beta", "alpha"}
		}
		for i, name := range names {
			in, box, out := newBarrier(n, ABC_INIT_ZERO)
			in.SetName(name + "_next")
			out.SetName(name)
			if i == 0 {
				box.SetName("box_a")
			} else {
				box.SetName("box_z")
			}
			Abc_ObjAddFanin(in, pi)
		}
		return n
	}
	if Abc_NtkCompareSignals(makeNetwork(false), makeNetwork(true), 0, 1) != 1 {
		t.Fatal("latches sorted by box names instead of output names")
	}
}

func TestFullNetworkCheckUsesOwningHashManager(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 0)
	defer Abc_NtkDelete(n)
	m := n.FunctionManager.(*Abc_Aig_t)
	a, b := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	a.SetName("a")
	b.SetName("b")
	gate := Abc_AigAnd(m, a, b)
	po := Abc_NtkCreatePo(n)
	po.SetName("y")
	Abc_ObjAddFanin(po, gate)
	if Abc_NtkDoCheck(n) != 1 {
		t.Fatal("valid strashed network failed")
	}
	Abc_AigAndDelete(m, gate.Object)
	if Abc_NtkDoCheck(n) != 0 {
		t.Error("whole-network check accepted AND removed from owning hash table")
	}
}

func TestNetworkCheckRejectsDuplicateObjectSlots(t *testing.T) {
	n := NewNetwork("duplicate slot")
	node := n.NewObject(ObjectPI)
	n.objects = append(n.objects, node)
	if err := n.Check(); err == nil {
		t.Error("same object accepted at two IDs")
	}
}

func TestAigCheckRejectsMissingAndDuplicateGateEntries(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 0)
	defer Abc_NtkDelete(n)
	m := n.FunctionManager.(*Abc_Aig_t)
	a, b, c := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	first := Abc_AigAnd(m, a, b)
	second := Abc_AigAnd(m, first, c)
	original := append([]*Object(nil), n.kindOrder[ObjectNode]...)
	defer func() { n.kindOrder[ObjectNode] = original }()
	for _, set := range [][]*Object{{first.Object}, {first.Object, first.Object}, {first.Object, second.Object, first.Object}} {
		n.kindOrder[ObjectNode] = set
		if Abc_AigCheck(m) != 0 {
			t.Errorf("accepted gate inventory %v", set)
		}
	}
}

func TestNetworkCheckMatchesEveryReciprocalPin(t *testing.T) {
	n := NewNetwork("mixed degree")
	src := make([]*Object, 96)
	for i := range src {
		src[i] = n.NewObject(ObjectPI)
	}
	wide := n.NewObject(ObjectNode)
	for i := range 192 {
		if err := n.AddFanin(wide, src[i%len(src)].Signal()); err != nil {
			t.Fatal(err)
		}
	}
	second := n.NewObject(ObjectNode)
	for _, i := range []int{0, 1, 0} {
		if err := n.AddFanin(second, src[i].Signal()); err != nil {
			t.Fatal(err)
		}
	}
	hole := n.NewObject(ObjectNode)
	if err := n.DeleteObject(hole); err != nil {
		t.Fatal(err)
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
	// Keep both overall numbers of edges and both destination multiplicities.
	src[3].fanouts[0], src[0].fanouts[2] = src[0].fanouts[2], src[3].fanouts[0]
	if err := n.Check(); err == nil {
		t.Error("accepted exchanged reciprocal edges")
	}
	src[3].fanouts[0], src[0].fanouts[2] = src[0].fanouts[2], src[3].fanouts[0]
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
	src[0].fanouts = append(src[0].fanouts, wide)
	if err := n.Check(); err == nil {
		t.Error("accepted excess duplicate fanout")
	}
}

func TestNetworkCheckAuditsAllObjectInventories(t *testing.T) {
	for _, tc := range []struct {
		name    string
		corrupt func(*Network)
	}{
		{"extra box", func(n *Network) { n.boxes = append(n.boxes, n.cis[0]) }},
		{"missing logic node", func(n *Network) { n.kindOrder[ObjectNode] = nil }},
		{"wrong logic node", func(n *Network) { n.kindOrder[ObjectNode] = []*Object{n.cis[0]} }},
		{"nil box", func(n *Network) { n.boxes = append(n.boxes, nil) }},
		{"foreign box", func(n *Network) { n.boxes = append(n.boxes, NewNetwork("other").NewObject(ObjectWhitebox)) }},
		{"missing whitebox", func(n *Network) { n.NewObject(ObjectWhitebox); n.boxes = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := validCheckedLogic()
			tc.corrupt(n)
			if Abc_NtkDoCheck(n) != 0 {
				t.Error("invalid inventory accepted")
			}
		})
	}
}
func TestCheckLatchRejectsMissingObject(t *testing.T) {
	if Abc_NtkCheckLatch(validCheckedLogic(), nil) != 0 {
		t.Fatal("accepted missing latch")
	}
}
func TestCheckNamesTracksBothTerminalKinds(t *testing.T) {
	n := NewNetwork("names")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	a.SetName("shared")
	b.SetName("second")
	y := n.NewObject(ObjectPO)
	y.SetName("shared")
	Abc_ObjAddFanin(y, a)
	if Abc_NtkCheckNames(n) != 1 {
		t.Fatal("direct same-name output rejected")
	}
	b.SetName("shared")
	if Abc_NtkCheckNames(n) != 0 {
		t.Fatal("duplicate input name accepted")
	}
	b.SetName("second")
	z := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(z, a)
	if Abc_NtkCheckNames(n) != 0 {
		t.Fatal("unnamed output accepted")
	}
	z.SetName("shared")
	if Abc_NtkCheckNames(n) != 0 {
		t.Fatal("duplicate output name accepted")
	}
	z.SetName("second")
	if Abc_NtkCheckNames(n) != 0 {
		t.Fatal("misconnected namesake accepted")
	}
	z.SetName("out")
	nm.Nm_ManStoreIdName(n.names, 4711, int32(ObjectPO), "stale", "")
	if Abc_NtkCheckNames(n) != 0 {
		t.Fatal("deleted name ID accepted")
	}
}
func TestCheckAigRejectsCachedTrivialPairs(t *testing.T) {
	n := NewNetwork("reductions")
	m := Abc_AigAlloc(n)
	a := n.NewObject(ObjectPI).Signal()
	for _, b := range []Edge{a, a.Not(), n.One(), n.One().Not()} {
		gate := Abc_AigAndCreate(m, a, b)
		if Abc_AigCheck(m) != 0 {
			t.Errorf("accepted reducible AND with second operand %+v", b)
		}
		Abc_AigDeleteNode(m, gate.Object)
	}
}
