package res

import (
	"github.com/glycerine/abcgo/base/abc"
	"testing"
)

func TestSourceResynthesisAfterSweepCreatesObjectHoles(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	a.SetName("a")
	b.SetName("b")
	inv := abc.Abc_NtkCreateNodeInv(n, a)
	root := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{inv, b})
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, root)
	po.SetName("out")
	abc.Abc_NtkCreateNodeBuf(n, a)
	params := &Res_Par_t{nWindow: 10, nSimWords: 1, nCands: 1}
	if got := Abc_NtkResynthesize(n, params); got != 1 {
		t.Fatalf("source resynthesis returned %d", got)
	}
	if abc.Abc_NtkCheck(n) == 0 {
		t.Fatal("invalid network after source resynthesis")
	}
}
