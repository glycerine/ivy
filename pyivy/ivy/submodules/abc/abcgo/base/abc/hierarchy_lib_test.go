package abc

import (
	"github.com/glycerine/abcgo/aig/hop"
	"testing"
)

func TestDesignModelRegistrationAndTopLevels(t *testing.T) {
	d := Abc_DesCreate("design")
	top, child, other := NewNetwork("top"), NewNetwork("child"), NewNetwork("other")
	for _, n := range []*Network{top, child, other} {
		if Abc_DesAddModel(d, n) != 1 {
			t.Fatal("registration failed")
		}
	}
	if Abc_DesAddModel(d, NewNetwork("child")) != 0 || Abc_DesFindModelByName(d, "child") != child {
		t.Fatal("model lookup or duplicate handling")
	}
	box := top.NewObject(ObjectWhitebox)
	box.Data = child
	if got := Abc_DesFindTopLevelModels(d); got != 2 || d.Tops[0] != top || d.Tops[1] != other {
		t.Fatalf("top levels: %v", d.Tops)
	}
	box.Data = nil
	if got := Abc_DesFindTopLevelModels(d); got != 3 {
		t.Fatalf("top count after disconnect: %d", got)
	}
}

func TestDesignDuplicateRemapsBoxModels(t *testing.T) {
	d := Abc_DesCreate("design")
	parent, child := NewNetwork("parent"), NewNetwork("child")
	Abc_DesAddModel(d, parent)
	Abc_DesAddModel(d, child)
	box := parent.NewObject(ObjectWhitebox)
	box.Data = child
	d.Tops = []*Network{parent}
	duplicate := Abc_DesDup(d)
	if duplicate == d || duplicate.Tops[0] == parent || duplicate.Tops[0].Boxes()[0].Data != duplicate.Module("child") {
		t.Fatal("duplicate retained original model pointers")
	}
	if duplicate.Module("child").Design != duplicate || child.Design != d {
		t.Fatal("design ownership changed")
	}
}

func TestDesignDeriveRootTransfersOwnership(t *testing.T) {
	d := Abc_DesCreate("single")
	root := NewNetwork("root")
	Abc_DesAddModel(d, root)
	d.Tops = []*Network{root}
	manager := d.FunctionManager
	if got := Abc_DesDeriveRoot(d); got != root || root.Design != nil || len(d.Modules) != 0 || d.Module("root") != nil {
		t.Fatal("root ownership was not transferred")
	}
	if root.FunctionManager != manager || d.FunctionManager != nil {
		t.Fatal("design manager was not transferred")
	}
	Abc_NtkDelete(root)
}

func TestDesignClearSharedManager(t *testing.T) {
	d := Abc_DesCreate("shared")
	manager := d.FunctionManager
	module := NewNetwork("module")
	module.FunctionManager = manager
	Abc_DesAddModel(d, module)
	Abc_DesCleanManPointer(d, manager)
	if d.FunctionManager != nil || module.FunctionManager != nil {
		t.Fatal("shared manager references remain")
	}
	hop.Hop_ManStop(manager)
	Abc_DesFree(d, nil)
}

func TestDesignDuplicateBlackboxesKeepsSavedNetwork(t *testing.T) {
	d := Abc_DesCreate("design")
	saved, blackboxModule := NewNetwork("saved"), NewNetwork("module")
	Abc_DesAddModel(d, saved)
	Abc_DesAddModel(d, blackboxModule)
	d.Tops = []*Network{saved}
	blackboxModule.Function = uint8(ABC_FUNC_BLACKBOX)
	blackboxModule.NewObject(ObjectBlackbox)
	duplicate := Abc_DesDupBlackboxes(d, saved)
	if len(duplicate.Modules) != 2 || duplicate.Modules[0] != saved || duplicate.Tops[0] != saved {
		t.Fatal("saved network not retained")
	}
	if duplicate.Modules[1] == blackboxModule || Abc_NtkBlackboxNum(duplicate.Modules[1]) != 1 {
		t.Fatal("black-box model not duplicated")
	}
	if saved.Design != d {
		t.Fatal("saved network ownership changed")
	}
}

func TestDesignFreePreservesSavedSharedManager(t *testing.T) {
	d := Abc_DesCreate("shared")
	saved, other := NewNetwork("saved"), NewNetwork("other")
	manager := hop.Hop_ManStart()
	saved.FunctionManager, other.FunctionManager = manager, manager
	Abc_DesAddModel(d, saved)
	Abc_DesAddModel(d, other)
	Abc_DesFree(d, saved)
	if saved.FunctionManager != manager || saved.Design != nil {
		t.Fatal("saved manager or ownership lost")
	}
	if other.FunctionManager != nil || len(d.Modules) != 0 {
		t.Fatal("freed design retained modules")
	}
	if hop.Hop_IthVar(manager, 0) == nil {
		t.Fatal("saved manager was stopped")
	}
	hop.Hop_ManStop(manager)
}

func TestStrashBoxModelPreservesInputAndOutputPhases(t *testing.T) {
	model := NewNetwork("xor")
	model.Type, model.Function = uint8(ABC_NTK_NETLIST), uint8(ABC_FUNC_AIG)
	manager := hop.Hop_ManStart()
	model.FunctionManager = manager
	left, right := model.NewObject(ObjectPI), model.NewObject(ObjectPI)
	leftNet, rightNet := model.NewObject(ObjectNet), model.NewObject(ObjectNet)
	Abc_ObjAddFanin(leftNet, left)
	Abc_ObjAddFanin(rightNet, right)
	node := model.NewObject(ObjectNode)
	Abc_ObjAddFanin(node, leftNet)
	Abc_ObjAddFanin(node, rightNet)
	node.Data = hop.Hop_Exor(manager, hop.Hop_IthVar(manager, 0), hop.Hop_IthVar(manager, 1))
	net := model.NewObject(ObjectNet)
	Abc_ObjAddFanin(net, node)
	output := model.NewObject(ObjectPO)
	Abc_ObjAddFanin(output, net.Signal().Not())
	parent := NewNetwork("parent")
	a, b := parent.NewObject(ObjectPI), parent.NewObject(ObjectPI)
	box, boxOutput := parent.NewObject(ObjectWhitebox), parent.NewObject(ObjectNet)
	box.Data, box.BoxInputPolarity = model, []uint32{1}
	Abc_ObjAddFanin(box, a)
	Abc_ObjAddFanin(box, b)
	Abc_ObjAddFanin(boxOutput, box)
	dst := NewNetwork("strashed")
	da, db := dst.NewObject(ObjectPI), dst.NewObject(ObjectPI)
	setCopy(a, da.Signal())
	setCopy(b, db.Signal())
	Abc_NodeStrashUsingNetwork(dst, box)
	for _, av := range []bool{false, true} {
		for _, bv := range []bool{false, true} {
			got, err := dst.Evaluate(getCopy(boxOutput), map[*Object]bool{da: av, db: bv})
			want := !((!av) != bv)
			if err != nil || got != want {
				t.Fatalf("a=%v b=%v got=%v want=%v err=%v", av, bv, got, want, err)
			}
		}
	}
	copy := Abc_NtkDupObj(NewNetwork("copy"), box, 0)
	box.BoxInputPolarity[0] = 0
	if copy.BoxInputPolarity[0] != 1 {
		t.Fatal("box polarity mask aliased the source")
	}
}
