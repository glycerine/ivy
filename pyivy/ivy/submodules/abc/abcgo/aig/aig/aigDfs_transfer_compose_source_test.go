package aig

import "testing"

func TestTransferConstantPreservesSourceCopyState(t *testing.T) {
	src, dst := NewGraph(), NewGraph()
	a := src.NewInput()
	Aig_ObjSetCopy(src, a, dst.Const0())
	if got := Aig_Transfer(src, dst, src.Const0(), 0); got != dst.Const0() {
		t.Errorf("constant transfer = %d, want const0", got)
	}
	if Aig_ObjCopy(src, a) != dst.Const0() || dst.InputCount() != 0 {
		t.Fatal("constant fast path changed source copies or destination inputs")
	}
}

func TestTransferSourceVariableBoundMayExceedInputCount(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("C maps all available CIs for a larger variable bound: %v", r)
		}
	}()
	src, dst := NewGraph(), NewGraph()
	a, b := src.NewInput(), src.NewInput()
	root := src.AND(a, b)
	got := Aig_Transfer(src, dst, root, 5)
	if dst.InputCount() != 2 || got != dst.AND(dst.inputs[0], dst.inputs[1]) {
		t.Fatal("transfer did not map all available source inputs")
	}
}

func TestTransferRecHonorsMarksAndRefreshesUnmarkedCopies(t *testing.T) {
	src, _, _, barrier, root := coneSourceFixture()
	dst := NewGraph()
	for range src.inputs {
		dst.NewInput()
	}
	for i, e := range src.inputs {
		Aig_ObjSetCopy(src, e, dst.inputs[i])
	}
	src.node(barrier).MarkA = true
	Aig_ObjSetCopy(src, barrier, dst.inputs[0].Not())
	Aig_ObjSetCopy(src, root, dst.Const0())
	Aig_Transfer_rec(src, dst, root)
	want := dst.AND(dst.inputs[0].Not(), dst.inputs[3])
	if Aig_ObjCopy(src, root) != want || !src.node(root).MarkA || !src.node(barrier).MarkA {
		t.Fatal("recursive transfer ignored a marked cone or trusted an unmarked stale copy")
	}
}

func TestTransferWrapperPreservesSeededAndUnrelatedCopies(t *testing.T) {
	src, _, _, barrier, root := coneSourceFixture()
	other := src.AND(src.inputs[0], src.inputs[3])
	dst := NewGraph()
	for range src.inputs {
		dst.NewInput()
	}
	src.node(barrier).MarkA = true
	Aig_ObjSetCopy(src, barrier, dst.inputs[0].Not())
	Aig_ObjSetCopy(src, other, dst.Const0())
	got := Aig_Transfer(src, dst, root.Not(), int32(len(src.inputs)))
	want := dst.AND(dst.inputs[0].Not(), dst.inputs[3]).Not()
	if got != want || Aig_ObjCopy(src, other) != dst.Const0() {
		t.Error("wrapper discarded seeded or unrelated source copies")
	}
	if src.node(root).MarkA || src.node(barrier).MarkA {
		t.Error("wrapper did not unmark the transferred cone")
	}
}

func TestComposeUndefinedVariableReturnsSourceNull(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("C returns NULL for an undefined PI index: %v", r)
		}
	}()
	g := NewGraph()
	a := g.NewInput()
	if got := Aig_Compose(g, a, g.Const0(), 1); got != InvalidEdge {
		t.Fatalf("undefined-variable result = %d, want InvalidEdge", got)
	}
}

func TestComposeRecHonorsMarkedCopiesAndLeavesMarks(t *testing.T) {
	g, a, _, barrier, root := coneSourceFixture()
	g.node(barrier).MarkA = true
	Aig_ObjSetCopy(g, barrier, a.Not())
	Aig_ObjSetCopy(g, root, g.Const0())
	Aig_Compose_rec(g, root, g.inputs[2], a)
	want := g.AND(a.Not(), g.inputs[3])
	if Aig_ObjCopy(g, root) != want || !g.node(root).MarkA || !g.node(barrier).MarkA {
		t.Error("recursive composition ignored seeded copies or failed to leave source marks")
	}
	if Aig_ObjCopy(g, g.inputs[3]) != g.inputs[3] {
		t.Error("recursive composition did not record its visited terminal copy")
	}
}

func TestComposeRecRecordsCopiesAcrossNodeSliceGrowth(t *testing.T) {
	g := NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(a, c)
	function := g.OR(b, d)
	// Any allocation must move the node slice; references into its old backing
	// array must not be used for the subsequent source mark update.
	g.nodes = append([]Node(nil), g.nodes...)
	g.nodes = g.nodes[:len(g.nodes):len(g.nodes)]
	before := len(g.nodes)
	Aig_Compose_rec(g, root, function, a)
	if len(g.nodes) == before || !g.node(root).MarkA {
		t.Error("composition did not preserve its source mark across graph growth")
	}
	if Aig_ObjCopy(g, a) != function || Aig_ObjCopy(g, c) != c {
		t.Error("composition did not retain leaf copy mappings")
	}
}
