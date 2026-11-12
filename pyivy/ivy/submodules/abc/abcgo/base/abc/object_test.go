package abc

import "testing"

func TestOriginalObjectLogicFunctions(t *testing.T) {
	n := NewNetwork("gates")
	a := Abc_NtkCreatePi(n)
	b := Abc_NtkCreatePi(n)
	c := Abc_NtkCreatePi(n)
	gates := []struct {
		o    *Object
		want func(bool, bool, bool) bool
	}{
		{Abc_NtkCreateNodeAnd(n, []*Object{a, b}), func(a, b, c bool) bool { return a && b }},
		{Abc_NtkCreateNodeOr(n, []*Object{a, b}), func(a, b, c bool) bool { return a || b }},
		{Abc_NtkCreateNodeExor(n, []*Object{a, b}), func(a, b, c bool) bool { return a != b }},
		{Abc_NtkCreateNodeMux(n, c, a, b), func(a, b, c bool) bool {
			if c {
				return a
			}
			return b
		}},
	}
	for _, gate := range gates {
		for i := 0; i < 8; i++ {
			av, bv, cv := i&1 != 0, i&2 != 0, i&4 != 0
			got, err := n.Evaluate(gate.o.Signal(), map[*Object]bool{a: av, b: bv, c: cv})
			if err != nil || got != gate.want(av, bv, cv) {
				t.Fatal(i, got, err)
			}
		}
	}
	zero := Abc_NtkCreateNodeConst0(n)
	one := Abc_NtkCreateNodeConst1(n)
	buf := Abc_NtkCreateNodeBuf(n, a)
	inv := Abc_NtkCreateNodeInv(n, a)
	if Abc_NodeIsConst0(zero) != 1 || Abc_NodeIsConst1(one) != 1 || Abc_NodeIsBuf(buf) != 1 || Abc_NodeIsInv(inv) != 1 {
		t.Fatal("predicates")
	}
	Abc_NodeComplement(zero)
	if Abc_NodeIsConst1(zero) != 1 {
		t.Fatal("complement constant")
	}
	Abc_NodeComplementInput(buf, a)
	if Abc_NodeIsInv(buf) != 1 {
		t.Fatal("complement variable")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestOriginalObjectNamesAndDuplication(t *testing.T) {
	n := NewNetwork("names")
	a := Abc_NtkCreatePi(n)
	a.SetName("input")
	node := Abc_NtkCreateNodeBuf(n, a)
	node.SetName("buffer")
	po := Abc_NtkCreatePo(n)
	po.SetName("output")
	Abc_ObjAddFanin(po, node)
	if Abc_NtkFindCi(n, "input") != a || Abc_NtkFindCo(n, "output") != po || Abc_NtkFindNode(n, "output") != node || Abc_NtkFindNode(n, "buffer") != node {
		t.Fatal("lookup")
	}
	net := Abc_NtkFindOrCreateNet(n, "wire")
	if Abc_NtkFindOrCreateNet(n, "wire") != net {
		t.Fatal("duplicate net")
	}
	clone := Abc_NtkCloneObj(node)
	if clone.Data != nil || Abc_ObjFanin0(clone) != a {
		t.Fatal("clone")
	}
	other := NewNetwork("copy")
	dup := Abc_NtkDupObj(other, a, 1)
	if dup.Name() != "input" || a.Copy != dup {
		t.Fatal("duplication")
	}
	box := Abc_NtkCreateLatch(n)
	bi := Abc_NtkCreateBi(n)
	bo := Abc_NtkCreateBo(n)
	Abc_ObjAddFanin(box, bi)
	Abc_ObjAddFanin(bo, box)
	copyBox := Abc_NtkDupBox(other, box, 1)
	if len(copyBox.fanins) != 1 || len(copyBox.fanouts) != 1 || copyBox.fanouts[0].Kind != ObjectBO {
		t.Fatal("box duplication")
	}
	if err := other.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestOriginalObjectDeletion(t *testing.T) {
	n := NewNetwork("delete")
	a := Abc_NtkCreatePi(n)
	node := Abc_NtkCreateNodeBuf(n, a)
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(po, node)
	id := node.ID
	Abc_NtkDeleteObj(node)
	if n.Object(id) != nil || len(po.fanins) != 0 || len(a.fanouts) != 0 {
		t.Fatal("delete connected object")
	}
	unregistered := Abc_ObjAlloc(n, ObjectNode)
	if unregistered.ID != -1 || n.contains(unregistered) {
		t.Fatal("allocation registered object")
	}
	Abc_ObjRecycle(unregistered)
	left := Abc_NtkCreateNodeBuf(n, a)
	right := Abc_NtkCreateNodeBuf(n, left)
	Abc_NtkDeleteObj_rec(right, 1)
	if n.contains(left) || n.contains(right) || !n.contains(a) {
		t.Fatal("recursive dangling deletion")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

// Unary constructors support deferred wiring, as required by AttachBottom.
func TestUnaryNodeDeferredFanin(t *testing.T) {
	for _, invert := range []bool{false, true} {
		n := NewNetwork("deferred")
		input := Abc_NtkCreatePi(n)
		var node *Object
		if invert {
			node = Abc_NtkCreateNodeInv(n, nil)
		} else {
			node = Abc_NtkCreateNodeBuf(n, nil)
		}
		if len(node.fanins) != 0 {
			t.Fatal("deferred node already has a fanin")
		}
		Abc_ObjAddFanin(node, input)
		for _, value := range []bool{false, true} {
			got, err := n.Evaluate(node.Signal(), map[*Object]bool{input: value})
			if err != nil || got != (value != invert) {
				t.Fatalf("invert=%v value=%v: got %v, %v", invert, value, got, err)
			}
		}
		if err := n.Check(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSOPConstructorStorage(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	nodes := []*Object{Abc_NtkCreateNodeConst0(n), Abc_NtkCreateNodeConst1(n), Abc_NtkCreateNodeInv(n, a), Abc_NtkCreateNodeBuf(n, a), Abc_NtkCreateNodeAnd(n, []*Object{a, b}), Abc_NtkCreateNodeOr(n, []*Object{a, b}), Abc_NtkCreateNodeExor(n, []*Object{a, b}), Abc_NtkCreateNodeMux(n, a, b, c)}
	for _, o := range nodes {
		if _, ok := o.Data.(string); !ok {
			t.Fatalf("SOP constructor stored%T", o.Data)
		}
	}
	for x := 0; x < 8; x++ {
		inputs := map[*Object]bool{a: x&1 != 0, b: x&2 != 0, c: x&4 != 0}
		want := []bool{false, true, x&1 == 0, x&1 != 0, x&3 == 3, x&3 != 0, (x&1 != 0) != (x&2 != 0), x&1 != 0 && x&2 != 0 || x&1 == 0 && x&4 != 0}
		for i, o := range nodes {
			got, err := n.Evaluate(o.Signal(), inputs)
			if err != nil || got != want[i] {
				t.Fatalf("constructor%d assignment%d", i, x)
			}
		}
	}
}
