package ver

import (
	"github.com/glycerine/abcgo/base/abc"
	"testing"
)

func TestCreateTerminalsLatchAndInverter(t *testing.T) {
	n := abc.NewNetwork("module")
	pi := Ver_ParseCreatePi(n, "in")
	in := abc.Abc_NtkFindNet(n, "in")
	if in == nil || abc.Abc_ObjFanin0(in) != pi {
		t.Fatal("PI to net connectivity")
	}
	po := Ver_ParseCreatePo(n, "out")
	out := abc.Abc_NtkFindNet(n, "out")
	if out == nil || abc.Abc_ObjFanin0(po) != out {
		t.Fatal("net to PO connectivity")
	}
	latch := Ver_ParseCreateLatch(n, in, out)
	bi := abc.Abc_ObjFanin0(latch)
	bo := abc.Abc_ObjFanin0(out)
	if abc.Abc_ObjFanin0(bi) != in || abc.Abc_ObjFanin0(bo) != latch || abc.Abc_LatchIsInitDc(latch) != 1 || abc.Abc_ObjName(latch) != "outL" {
		t.Fatal("latch chain, name or initialization")
	}
	inverse := Ver_ParseCreateInv(n, in)
	node := abc.Abc_ObjFanin0(inverse)
	if node == nil || abc.Abc_ObjFanin0(node) != in {
		t.Fatal("inverter connectivity")
	}
}
