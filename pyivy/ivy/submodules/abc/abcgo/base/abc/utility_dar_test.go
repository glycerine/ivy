package abc

import "testing"

func TestNtkToDarInitialOne(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 0)
	a := Abc_NtkCreatePi(n)
	lo := Abc_NtkAddLatch(n, nil, ABC_INIT_ONE)
	latch := Abc_ObjFanin0(lo)
	li := Abc_ObjFanin0(latch)
	Abc_ObjAddFanin(li, a)
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(po, lo)
	Abc_NtkOrderCisCos(n)
	g := Abc_NtkToDar(n, 0, 1)
	if g.RegisterCount() != 1 || len(g.FlopNumbers) != 1 {
		t.Fatal("lost register metadata")
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		out, err := g.Evaluate(in)
		if err != nil || out[0] != !in[1] || out[1] != !in[0] {
			t.Fatalf("input %v output %v err %v", in, out, err)
		}
	}
}
