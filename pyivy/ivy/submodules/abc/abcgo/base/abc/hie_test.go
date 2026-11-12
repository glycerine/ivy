package abc

import (
	"strings"
	"testing"
)

func TestHierarchyFlattenNestedAndReusedModels(t *testing.T) {
	for _, suffix := range []bool{false, true} {
		root := testHierarchicalDesign(t, true)
		var flat *Network
		if suffix {
			flat = Abc_NtkFlattenLogicHierarchy2(root)
		} else {
			flat = Abc_NtkFlattenLogicHierarchy(root)
		}
		if flat == nil {
			t.Fatal("flatten failed", suffix)
		}
		if len(flat.ObjectsOfKind(ObjectWhitebox)) != 0 {
			t.Fatal("logic boxes retained")
		}
		assertBufferTruth(t, flat, true)
		if flat.Design == root.Design || flat.Design.Module(flat.Name) != flat {
			t.Fatal("design ownership not separated")
		}
	}
	root := testHierarchicalDesign(t, false)
	leaf := root.Design.Module("leaf")
	inputs := root.ObjectsOfKind(ObjectPI)
	left, right := inputs[0].fanouts[0], inputs[1].fanouts[0]
	second := testInstance(root, leaf, "second", right, left)
	testOutputNet(root, "second_result", second.Signal())
	flat := Abc_NtkFlattenLogicHierarchy2(root)
	if flat == nil {
		t.Fatal("reused module rejected")
	}
	for assignment := 0; assignment < 4; assignment++ {
		x, y := assignment&1 != 0, assignment&2 != 0
		values := map[*Object]bool{}
		for i, pi := range flat.ObjectsOfKind(ObjectPI) {
			values[pi] = []bool{x, y}[i]
		}
		for i, po := range flat.ObjectsOfKind(ObjectPO) {
			got, err := flat.Evaluate(po.Signal(), values)
			want := x && !y
			if i == 1 {
				want = y && !x
			}
			if err != nil || got != want {
				t.Fatal(assignment, i, got, want, err)
			}
		}
	}
}
func TestHierarchyFlattenWireAliasAndProperty(t *testing.T) {
	wire := testModule("wire")
	net := testInputNet(wire, "x")
	po := wire.NewObject(ObjectPO)
	po.SetName("y")
	Abc_ObjAddFanin(po, net)
	root := testModule("top")
	input := testInputNet(root, "a")
	output := testInstance(root, wire, "wire_inst", input)
	testOutputNet(root, "z", output.Signal())
	flat := Abc_NtkFlattenLogicHierarchy(root)
	if flat == nil {
		t.Fatal("wire alias flatten failed")
	}
	for _, value := range []bool{false, true} {
		got, err := flat.Evaluate(flat.ObjectsOfKind(ObjectPO)[0].Signal(), map[*Object]bool{flat.ObjectsOfKind(ObjectPI)[0]: value})
		if err != nil || got != value {
			t.Fatal(got, value, err)
		}
	}
	original := testHierarchicalDesign(t, false)
	leaf := original.Design.Module("leaf")
	leaf.ObjectsOfKind(ObjectPO)[0].fanins[0].Object.SetName("abc_property_leaf")
	oldOutput := original.ObjectsOfKind(ObjectPO)[0]
	driver := oldOutput.fanins[0].Object
	inverter := Abc_NtkCreateNodeInv(original, driver)
	newNet := netForDriver(original, inverter.Signal())
	newNet.SetName("inverted")
	if err := original.PatchFanin(oldOutput, 0, newNet.Signal()); err != nil {
		t.Fatal(err)
	}
	flat = Abc_NtkFlattenLogicHierarchy(original)
	if flat == nil || len(flat.ObjectsOfKind(ObjectPO)) != 2 {
		t.Fatal("nested property not exported", flat)
	}
	found := false
	for _, net := range flat.ObjectsOfKind(ObjectNet) {
		found = found || strings.Contains(net.Name(), "abc_property_leaf")
	}
	if !found {
		t.Fatal("property name lost")
	}
}
func testBlackboxHierarchy(t *testing.T) *Network {
	t.Helper()
	model := testModule("opaque")
	model.Function = uint8(ABC_FUNC_BLACKBOX)
	input := testInputNet(model, "u")
	body := model.NewObject(ObjectBlackbox)
	body.SetName("body")
	bi, bo := model.NewObject(ObjectBI), model.NewObject(ObjectBO)
	Abc_ObjAddFanin(bi, input)
	Abc_ObjAddFanin(body, bi)
	Abc_ObjAddFanin(bo, body)
	testOutputNet(model, "v", bo.Signal())
	root := testModule("host")
	a, b := testInputNet(root, "a"), testInputNet(root, "b")
	logic := createLogicNode(root, []*Object{a, b}, func(v []bool) bool { return v[0] != v[1] })
	drive := netForDriver(root, logic.Signal())
	drive.SetName("drive")
	box := root.NewObject(ObjectBlackbox)
	box.Data = model
	box.SetName("opaque_inst")
	in, out := root.NewObject(ObjectBI), root.NewObject(ObjectBO)
	Abc_ObjAddFanin(in, drive)
	Abc_ObjAddFanin(box, in)
	Abc_ObjAddFanin(out, box)
	outNet := netForDriver(root, out.Signal())
	outNet.SetName("opaque_out")
	combined := createLogicNode(root, []*Object{drive, outNet}, func(v []bool) bool { return v[0] && v[1] })
	testOutputNet(root, "z", combined.Signal())
	observed := root.NewObject(ObjectPO)
	observed.SetName("observed")
	Abc_ObjAddFanin(observed, drive)
	design := NewDesign("blackboxes")
	if err := design.AddModule(root); err != nil {
		t.Fatal(err)
	}
	if err := design.AddModule(model); err != nil {
		t.Fatal(err)
	}
	design.Tops = []*Network{root}
	return root
}
func TestHierarchyBlackboxExtractReinsertAndSharedOutputNet(t *testing.T) {
	root := testBlackboxHierarchy(t)
	flat := Abc_NtkFlattenLogicHierarchy2(root)
	if flat == nil || len(flat.ObjectsOfKind(ObjectBlackbox)) != 1 {
		t.Fatal("blackbox flatten failed")
	}
	if flat.ObjectsOfKind(ObjectBlackbox)[0].Data == root.Design.Module("opaque") {
		t.Fatal("blackbox definition not owned by new design")
	}
	logic := Abc_NtkConvertBlackboxes(flat)
	if logic == nil || len(logic.ObjectsOfKind(ObjectBlackbox)) != 0 {
		t.Fatal("blackbox extraction failed")
	}
	if len(logic.ObjectsOfKind(ObjectPI)) != 3 || len(logic.ObjectsOfKind(ObjectPO)) != 2 {
		t.Fatal("wrong cut interface")
	}
	design := flat.Design
	restored := Abc_NtkInsertNewLogic(flat, logic)
	if restored == nil {
		t.Fatal("blackbox insertion failed")
	}
	if flat.Design != nil || restored.Design != design || design.Module(restored.Name) != restored || design.Tops[0] != restored {
		t.Fatal("design ownership transfer failed")
	}
	for assignment := 0; assignment < 8; assignment++ {
		a, b, opaque := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		values := map[*Object]bool{}
		for _, ci := range restored.Inputs() {
			switch {
			case ci.Kind == ObjectBO:
				values[ci] = opaque
			case ci.Name() == "a":
				values[ci] = a
			default:
				values[ci] = b
			}
		}
		for _, po := range restored.ObjectsOfKind(ObjectPO) {
			got, err := restored.Evaluate(po.Signal(), values)
			want := a != b
			if po.Name() == "z" {
				want = want && opaque
			}
			if err != nil || got != want {
				t.Fatal(assignment, po.Name(), got, want, err)
			}
		}
	}
}
func TestHierarchyFlattenBDDManagers(t *testing.T) {
	root := testHierarchicalDesign(t, true)
	for _, model := range root.Design.Modules {
		if Abc_NtkSopToBdd(model) != 1 {
			t.Fatal("module BDD conversion failed")
		}
	}
	flat := Abc_NtkFlattenLogicHierarchy2(root)
	if flat == nil {
		t.Fatal("BDD hierarchy flatten failed")
	}
	assertBufferTruth(t, flat, true)
}
func TestHierarchyRejectsCycleBeforeConstruction(t *testing.T) {
	root := testHierarchicalDesign(t, false)
	leaf := root.Design.Module("leaf")
	leaf.Boxes()
	testInstance(leaf, root, "recursive", leaf.ObjectsOfKind(ObjectPI)[0].fanouts[0], leaf.ObjectsOfKind(ObjectPI)[1].fanouts[0])
	if flat := Abc_NtkFlattenLogicHierarchy(root); flat != nil {
		t.Fatal("recursive hierarchy flattened")
	}
}
func TestHierarchyMVWireAliasRetainsParentValueNames(t *testing.T) {
	wire := testModule("mvwire")
	wire.Function = uint8(ABC_FUNC_BLIFMV)
	net := testInputNet(wire, "x")
	Abc_NtkSetMvVarValues(net, 3)
	Abc_ObjMvVar(net).Names = []string{"child0", "child1", "child2"}
	po := wire.NewObject(ObjectPO)
	po.SetName("y")
	Abc_ObjAddFanin(po, net)
	root := testModule("mvtop")
	root.Function = uint8(ABC_FUNC_BLIFMV)
	input := testInputNet(root, "a")
	Abc_NtkSetMvVarValues(input, 3)
	Abc_ObjMvVar(input).Names = []string{"red", "green", "blue"}
	output := testInstance(root, wire, "inst", input)
	Abc_NtkSetMvVarValues(output, 3)
	testOutputNet(root, "z", output.Signal())
	flat := Abc_NtkFlattenLogicHierarchy2(root)
	if flat == nil {
		t.Fatal("MV wire flatten failed")
	}
	pi := flat.ObjectsOfKind(ObjectPI)[0]
	values := Abc_ObjMvVar(pi.fanouts[0])
	if values == nil || values.Names[0] != "red" {
		t.Fatal("child overwrote parent values", values)
	}
	for input := 0; input < 3; input++ {
		got, err := flat.EvaluateMV(flat.ObjectsOfKind(ObjectPO)[0], map[*Object]int{pi: input})
		if err != nil || len(got) != 1 || got[0] != input {
			t.Fatal(input, got, err)
		}
	}
}
