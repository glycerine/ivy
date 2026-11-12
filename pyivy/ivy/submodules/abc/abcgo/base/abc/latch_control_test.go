package abc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestClassifyLatchControls(t *testing.T) {
	n := NewNetwork("controls")
	enable := n.NewObject(ObjectPI)
	data := n.NewObject(ObjectPI)
	makeLatch := func(opposite bool) (*Object, *Object) {
		output := Abc_NtkAddLatch(n, nil, ABC_INIT_ZERO)
		latch := Abc_ObjFanin0(output)
		input := Abc_ObjFanin0(latch)
		mux := n.NewObject(ObjectNode)
		Abc_ObjAddFanin(mux, enable)
		if opposite {
			Abc_ObjAddFanin(mux, output)
			Abc_ObjAddFanin(mux, data)
		} else {
			Abc_ObjAddFanin(mux, data)
			Abc_ObjAddFanin(mux, output)
		}
		Abc_ObjAddFanin(input, mux)
		return input, mux
	}
	input, _ := makeLatch(false)
	other, otherMux := makeLatch(true)
	controls, classes, counts := abcClassifyLatchControls(n)
	if counts.enabled != 1 || counts.opposite != 1 || vec.Vec_PtrSize(controls) != 1 || vec.Vec_IntEntry(classes, 0) != 0 || vec.Vec_IntEntry(classes, 1) != -1 {
		t.Fatalf("counts %+v classes %v", counts, vec.Vec_IntArray(classes))
	}
	if Abc_ObjFanin0(input) != data || Abc_ObjFanin0(other) != otherMux {
		t.Fatal("incorrect mux rewrite")
	}
}
