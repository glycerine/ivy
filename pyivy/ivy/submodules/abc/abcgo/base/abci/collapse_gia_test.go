package abci

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
)

func TestSourceCollapseGiaPreservesConesAndScratch(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	n.Name, n.Spec = "named", "design.blif"
	x, y := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	active := abc.Abc_AigAnd(manager, x, y.Signal().Not())
	dead := abc.Abc_AigAnd(manager, x, y)
	p0, p1 := abc.Abc_NtkCreatePo(n), abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(p0, active)
	abc.Abc_ObjAddFanin(p1, active.Not())
	dst := Abc_NtkClpGia(n)
	name, spec := dst.Identity()
	if name != "named" || spec != "design.blif" || gia.Gia_ManCiNum(dst) != 2 || gia.Gia_ManCoNum(dst) != 2 || gia.Gia_ManAndNum(dst) != 1 {
		t.Fatalf("identity=%s/%s CI=%d CO=%d AND=%d", name, spec, gia.Gia_ManCiNum(dst), gia.Gia_ManCoNum(dst), gia.Gia_ManAndNum(dst))
	}
	if dead.Object.Temp != -1 || p0.Temp != -1 || p1.Temp != -1 || n.One().Object.Temp != 1 {
		t.Fatal("source scratch values differ")
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		got, err := dst.Evaluate(in)
		want := in[0] && !in[1]
		if err != nil || got[0] != want || got[1] == want {
			t.Fatalf("inputs=%v outputs=%v error=%v", in, got, err)
		}
	}
}
