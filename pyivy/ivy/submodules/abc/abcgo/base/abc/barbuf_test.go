package abc

import "testing"

func testInputNet(n *Network, name string) *Object {
	pi := n.NewObject(ObjectPI)
	pi.SetName(name)
	net := netForDriver(n, pi.Signal())
	net.SetName(name)
	return net
}
func testOutputNet(n *Network, name string, driver Edge) *Object {
	net := netForDriver(n, driver)
	net.SetName(name)
	po := n.NewObject(ObjectPO)
	po.SetName(name)
	Abc_ObjAddFanin(po, net)
	return po
}
func testModule(name string) *Network {
	n := NewNetwork(name)
	n.Type = uint8(ABC_NTK_NETLIST)
	n.Function = uint8(ABC_FUNC_SOP)
	return n
}
func testInstance(n, model *Network, name string, inputs ...*Object) *Object {
	box := n.NewObject(ObjectWhitebox)
	box.Data = model
	box.SetName(name)
	for _, net := range inputs {
		bi := n.NewObject(ObjectBI)
		Abc_ObjAddFanin(bi, net)
		Abc_ObjAddFanin(box, bi)
	}
	bo := n.NewObject(ObjectBO)
	Abc_ObjAddFanin(bo, box)
	net := netForDriver(n, bo.Signal())
	net.SetName(name + "_out")
	return net
}
func testHierarchicalDesign(t *testing.T, nested bool) *Network {
	t.Helper()
	leaf := testModule("leaf")
	a, b := testInputNet(leaf, "a"), testInputNet(leaf, "b")
	node := createLogicNode(leaf, []*Object{a, b}, func(v []bool) bool { return v[0] && !v[1] })
	testOutputNet(leaf, "result", node.Signal())
	model := leaf
	var middle *Network
	if nested {
		middle = testModule("middle")
		x, y := testInputNet(middle, "x"), testInputNet(middle, "y")
		out := testInstance(middle, leaf, "inner", x, y)
		inv := Abc_NtkCreateNodeInv(middle, out)
		testOutputNet(middle, "middle_result", inv.Signal())
		model = middle
	}
	root := testModule("root")
	x, y := testInputNet(root, "x"), testInputNet(root, "y")
	out := testInstance(root, model, "instance", x, y)
	testOutputNet(root, "out", out.Signal())
	design := NewDesign("design")
	for _, n := range []*Network{root, leaf} {
		if err := design.AddModule(n); err != nil {
			t.Fatal(err)
		}
	}
	if middle != nil {
		if err := design.AddModule(middle); err != nil {
			t.Fatal(err)
		}
	}
	design.Tops = []*Network{root}
	return root
}
func assertBufferTruth(t *testing.T, n *Network, negate bool) {
	t.Helper()
	for i := 0; i < 4; i++ {
		x, y := i&1 != 0, i&2 != 0
		inputs := map[*Object]bool{}
		for _, pi := range n.ObjectsOfKind(ObjectPI) {
			if pi.Name() == "x" {
				inputs[pi] = x
			} else {
				inputs[pi] = y
			}
		}
		want := x && !y
		if negate {
			want = !want
		}
		got, err := n.Evaluate(n.ObjectsOfKind(ObjectPO)[0].Signal(), inputs)
		if err != nil || got != want {
			t.Fatalf("input %d: got %v %v want %v", i, got, err, want)
		}
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestHierarchyBarrierRoundtrip(t *testing.T) {
	for _, nested := range []bool{false, true} {
		root := testHierarchicalDesign(t, nested)
		if Abc_NtkCheckSingleInstance(root) != 1 {
			t.Fatal("single-instance design rejected")
		}
		var li, lo []*Object
		boxes := Abc_NtkCollectPiPos(root, &li, &lo)
		wantBoxes, wantBarriers := int32(1), 3
		if nested {
			wantBoxes, wantBarriers = 2, 6
		}
		if boxes != wantBoxes || len(li) != wantBarriers || len(lo) != wantBarriers {
			t.Fatal(boxes, len(li), len(lo))
		}
		flat := Abc_NtkToBarBufs(root)
		if flat == nil || flat.Barriers != wantBarriers {
			t.Fatal("flattening")
		}
		buffers := Abc_NtkBarBufsToBuffers(flat)
		if int(Abc_NtkCountBarBufs(buffers)) != wantBarriers {
			t.Fatal("buffer count")
		}
		assertBufferTruth(t, buffers, nested)
		restoredFlat := Abc_NtkBarBufsFromBuffers(flat, buffers)
		if restoredFlat.Barriers != wantBarriers {
			t.Fatal("lost barrier metadata")
		}
		assertBufferTruth(t, Abc_NtkBarBufsToBuffers(restoredFlat), nested)
		restored := Abc_NtkFromBarBufs(root, restoredFlat)
		if restored.Design == nil || len(restored.Design.Modules) != int(wantBoxes)+1 || restored.Design.Module("leaf") == nil {
			t.Fatal("lost hierarchy")
		}
		reflat := Abc_NtkToBarBufs(restored)
		assertBufferTruth(t, Abc_NtkBarBufsToBuffers(reflat), nested)
	}
}
func TestOrderedTerminalIDs(t *testing.T) {
	n := NewNetwork("ordered")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	beforeA, beforeB := a.ID, b.ID
	if err := n.SetTerminalOrder(ObjectPI, []int{b.ID, a.ID}); err != nil {
		t.Fatal(err)
	}
	if n.Inputs()[0] != b || n.ObjectsOfKind(ObjectPI)[0] != b || a.ID != beforeA || b.ID != beforeB {
		t.Fatal("order changed stable IDs")
	}
	if err := n.SetTerminalOrder(ObjectPI, []int{a.ID, a.ID}); err == nil {
		t.Fatal("accepted repeated ID")
	}
	if err := n.SetTerminalOrder(ObjectPO, []int{a.ID}); err == nil {
		t.Fatal("accepted foreign kind")
	}
	if err := n.DeleteObject(b); err != nil {
		t.Fatal(err)
	}
	if len(n.Inputs()) != 1 || n.Inputs()[0] != a {
		t.Fatal("deleted interface remains ordered")
	}
}
func TestHierarchyRejectsSharedAndRecursiveModels(t *testing.T) {
	root := testHierarchicalDesign(t, false)
	leaf := root.Design.Module("leaf")
	pis := root.ObjectsOfKind(ObjectPI)
	testInstance(root, leaf, "second", pis[0].fanouts[0], pis[1].fanouts[0])
	if Abc_NtkCheckSingleInstance(root) != 0 || Abc_NtkToBarBufs(root) != nil {
		t.Fatal("accepted shared model")
	}
	root = testHierarchicalDesign(t, false)
	leaf = root.Design.Module("leaf")
	testInstance(leaf, root, "cycle", leaf.ObjectsOfKind(ObjectPI)[0].fanouts[0], leaf.ObjectsOfKind(ObjectPI)[1].fanouts[0])
	if Abc_NtkCheckSingleInstance(root) != 0 {
		t.Fatal("accepted recursive hierarchy")
	}
}

func TestHierarchyUsesModelTerminalOrder(t *testing.T) {
	root := testHierarchicalDesign(t, false)
	leaf := root.Design.Module("leaf")
	pis := leaf.ObjectsOfKind(ObjectPI)
	if err := leaf.SetTerminalOrder(ObjectPI, []int{pis[1].ID, pis[0].ID}); err != nil {
		t.Fatal(err)
	}
	buffers := Abc_NtkBarBufsToBuffers(Abc_NtkToBarBufs(root))
	for i := 0; i < 4; i++ {
		x, y := i&1 != 0, i&2 != 0
		values := map[*Object]bool{}
		for _, pi := range buffers.ObjectsOfKind(ObjectPI) {
			if pi.Name() == "x" {
				values[pi] = x
			} else {
				values[pi] = y
			}
		}
		got, err := buffers.Evaluate(buffers.ObjectsOfKind(ObjectPO)[0].Signal(), values)
		if err != nil || got != (y && !x) {
			t.Fatal("instance ignored declared model PI order", i, got, err)
		}
	}
}
func TestHierarchyUnusedPrimaryInputAndComplementEdges(t *testing.T) {
	root := testHierarchicalDesign(t, false)
	unused := root.NewObject(ObjectPI)
	unused.SetName("unused")
	po := root.ObjectsOfKind(ObjectPO)[0]
	po.fanins[0].Complement = true
	flat := Abc_NtkToBarBufs(root)
	buffers := Abc_NtkBarBufsToBuffers(flat)
	for i := 0; i < 8; i++ {
		values := map[*Object]bool{}
		for _, pi := range buffers.ObjectsOfKind(ObjectPI) {
			switch pi.Name() {
			case "x":
				values[pi] = i&1 != 0
			case "y":
				values[pi] = i&2 != 0
			case "unused":
				values[pi] = i&4 != 0
			}
		}
		got, err := buffers.Evaluate(buffers.ObjectsOfKind(ObjectPO)[0].Signal(), values)
		want := !((i&1 != 0) && !(i&2 != 0))
		if err != nil || got != want {
			t.Fatal(i, got, err, want)
		}
	}
	restored := Abc_NtkFromBarBufs(root, Abc_NtkBarBufsFromBuffers(flat, buffers))
	if len(restored.ObjectsOfKind(ObjectPI)) != 3 {
		t.Fatal("unused input lost")
	}
}

func TestHierarchyRoundtripPreservesUnusedModules(t *testing.T) {
	root := testHierarchicalDesign(t, false)
	unused := testModule("unused")
	a, b := testInputNet(unused, "a"), testInputNet(unused, "b")
	node := Abc_NtkCreateNodeAnd(unused, []*Object{a, b})
	testOutputNet(unused, "z", node.Signal())
	if err := root.Design.AddModule(unused); err != nil {
		t.Fatal(err)
	}
	root.Design.Tops = append(root.Design.Tops, unused)
	flat := Abc_NtkToBarBufs(root)
	restored := Abc_NtkFromBarBufs(root, flat)
	kept := restored.Design.Module("unused")
	if kept == nil || kept == unused || len(kept.ObjectsOfKind(ObjectNode)) != 1 || len(restored.Design.Tops) != 2 {
		t.Fatal("unused module/top lost")
	}
	if err := kept.Check(); err != nil {
		t.Fatal(err)
	}
	values := map[*Object]bool{}
	for _, pi := range kept.ObjectsOfKind(ObjectPI) {
		values[pi] = true
	}
	got, err := kept.Evaluate(kept.ObjectsOfKind(ObjectPO)[0].Signal(), values)
	if err != nil || !got {
		t.Fatal(got, err)
	}
}

func TestBarrierMismatchRejectedBeforeChangingCopies(t *testing.T) {
	root := testHierarchicalDesign(t, false)
	flat := Abc_NtkToBarBufs(root)
	buffers := Abc_NtkBarBufsToBuffers(flat)
	for _, node := range buffers.ObjectsOfKind(ObjectNode) {
		if isBarrierBuffer(node) {
			node.Barrier = false
			node.Data = func(v []bool) bool { return v[0] }
			break
		}
	}
	sentinel := NewNetwork("sentinel")
	flat.Copy = sentinel
	defer func() {
		if recover() == nil {
			t.Error("accepted missing barrier")
		}
		if flat.Copy != sentinel {
			t.Error("failed restoration modified source copy mapping")
		}
	}()
	Abc_NtkBarBufsFromBuffers(flat, buffers)
}
func TestDesignRejectsSameModuleIdentityTwice(t *testing.T) {
	d := NewDesign("unique")
	n := NewNetwork("first")
	if err := d.AddModule(n); err != nil {
		t.Fatal(err)
	}
	n.Name = "renamed"
	if err := d.AddModule(n); err == nil {
		t.Fatal("module pointer was registered twice")
	}
}

func TestHierarchyUsesOrderedOutputMaps(t *testing.T) {
	model := testModule("two-output")
	a, b := testInputNet(model, "a"), testInputNet(model, "b")
	pa := testOutputNet(model, "aout", a.Signal())
	pb := testOutputNet(model, "bout", b.Signal())
	if err := model.SetTerminalOrder(ObjectPO, []int{pb.ID, pa.ID}); err != nil {
		t.Fatal(err)
	}
	root := testModule("root-two")
	x, y := testInputNet(root, "x"), testInputNet(root, "y")
	box := root.NewObject(ObjectWhitebox)
	box.Data = model
	for _, net := range []*Object{x, y} {
		bi := root.NewObject(ObjectBI)
		Abc_ObjAddFanin(bi, net)
		Abc_ObjAddFanin(box, bi)
	}
	var pos []*Object
	for _, name := range []string{"zero", "one"} {
		bo := root.NewObject(ObjectBO)
		Abc_ObjAddFanin(bo, box)
		net := netForDriver(root, bo.Signal())
		pos = append(pos, testOutputNet(root, name, net.Signal()))
	}
	if err := root.SetTerminalOrder(ObjectPO, []int{pos[1].ID, pos[0].ID}); err != nil {
		t.Fatal(err)
	}
	design := NewDesign("ordered-output")
	for _, module := range []*Network{root, model} {
		if err := design.AddModule(module); err != nil {
			t.Fatal(err)
		}
	}
	design.Tops = []*Network{root}
	flat := Abc_NtkToBarBufs(root)
	buffers := Abc_NtkBarBufsToBuffers(flat)
	for pattern := 0; pattern < 4; pattern++ {
		xv, yv := pattern&1 != 0, pattern&2 != 0
		values := map[*Object]bool{}
		for _, pi := range buffers.ObjectsOfKind(ObjectPI) {
			if pi.Name() == "x" {
				values[pi] = xv
			} else {
				values[pi] = yv
			}
		}
		for _, po := range buffers.ObjectsOfKind(ObjectPO) {
			want := xv
			if po.Name() == "zero" {
				want = yv
			}
			got, err := buffers.Evaluate(po.Signal(), values)
			if err != nil || got != want {
				t.Fatal(pattern, po.Name(), got, err, want)
			}
		}
	}
	restored := Abc_NtkFromBarBufs(root, flat)
	if restored.ObjectsOfKind(ObjectPO)[0].Name() != "one" || restored.Design.Module("two-output").ObjectsOfKind(ObjectPO)[0].Name() != "bout" {
		t.Fatal("ordered output interface lost")
	}
}
