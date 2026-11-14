package ver

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestParseConnectBoxNamedAndPositional(t *testing.T) {
	for _, named := range []bool{false, true} {
		t.Run(map[bool]string{false: "positional", true: "named"}[named], func(t *testing.T) {
			parent := abc.NewNetwork("parent")
			model := abc.NewNetwork("child")
			Ver_ParseCreatePi(model, "din[3]")
			Ver_ParseCreatePo(model, "dout")
			box := abc.Abc_NtkCreateBlackbox(parent)
			box.Data = model
			in := abc.Abc_NtkFindOrCreateNet(parent, "actual_in")
			out := abc.Abc_NtkFindOrCreateNet(parent, "actual_out")
			bundles := vec.Vec_PtrAlloc(2)
			add := func(formal string, actual *abc.Object) {
				b := &Ver_Bundle_t{pNameFormal: formal, vNetsActual: vec.Vec_PtrAlloc(1)}
				vec.Vec_PtrPush(b.vNetsActual, actual)
				vec.Vec_PtrPush(bundles, b)
			}
			if named {
				add("din", in)
				add("dout", out)
			} else {
				add("", in)
				add("", out)
			}
			man := &Ver_Man_t_{boxBundles: map[*abc.Object]*vec.Vec_Ptr_t{box: bundles}}
			if Ver_ParseConnectBox(man, box) == 0 {
				t.Fatalf("connect failed: %s", string(man.sError[:]))
			}
			if abc.Abc_ObjFaninNum(box) != 1 || abc.Abc_ObjFanoutNum(box) != 1 {
				t.Fatalf("unexpected box terminals: %d inputs, %d outputs", abc.Abc_ObjFaninNum(box), abc.Abc_ObjFanoutNum(box))
			}
			bi := abc.Abc_ObjFanin(box, 0)
			bo := abc.Abc_ObjFanout(box, 0)
			if abc.Abc_ObjFanin0(bi) != in || abc.Abc_ObjFanout0(bo) != out {
				t.Fatal("box terminal connected to wrong actual net")
			}
			if man.boxBundles[box] != nil {
				t.Fatal("bundle ownership was not released")
			}
		})
	}
}

func TestParseCollectUndefBoxesGroupsBlackboxInstances(t *testing.T) {
	design := abc.NewDesign("undef grouping")
	top := abc.NewNetwork("top")
	model := abc.NewNetwork("missing")
	for _, module := range []*abc.Network{top, model} {
		if err := design.AddModule(module); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		box := abc.Abc_NtkCreateBlackbox(top)
		box.Data = model
		abc.Abc_ObjAssignName(box, []string{"u0", "u1"}[i], "")
	}
	got := Ver_ParseCollectUndefBoxes(&Ver_Man_t_{pDesign: design})
	defer vec.Vec_PtrFree(got)
	if vec.Vec_PtrSize(got) != 1 || vec.Vec_PtrEntry(got, 0) != model {
		t.Fatalf("undefined models = %d, want one missing model", vec.Vec_PtrSize(got))
	}
	instances, ok := model.Data.(*vec.Vec_Ptr_t)
	if !ok || vec.Vec_PtrSize(instances) != 2 {
		t.Fatalf("grouped instances = %v, want two", instances)
	}
	vec.Vec_PtrFree(instances)
	model.Data = nil
}
