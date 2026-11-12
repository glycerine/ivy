package abc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestClassRetimeIntegration(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 0)
	pi := Abc_NtkCreatePi(n)
	pi.SetName("in")
	lo := Abc_NtkAddLatch(n, nil, ABC_INIT_ZERO)
	li := Abc_ObjFanin0(Abc_ObjFanin0(lo))
	Abc_ObjAddFanin(li, pi)
	po := Abc_NtkCreatePo(n)
	po.SetName("out")
	Abc_ObjAddFanin(po, lo)
	Abc_NtkOrderCisCos(n)
	classes := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(classes, -1)
	var out *vec.Vec_Int_t
	result := Abc_NtkRetimeWithClassesNtk(n, classes, &out, 0)
	if result == nil || Abc_NtkLatchNum(result) != 1 || vec.Vec_IntSize(out) != 1 || Abc_NtkCheck(result) == 0 {
		t.Fatal("retiming produced invalid network or classes")
	}
}
