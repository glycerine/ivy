package abc

import "testing"

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
