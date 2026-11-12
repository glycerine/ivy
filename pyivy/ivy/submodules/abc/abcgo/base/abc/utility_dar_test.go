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

func TestDarSequentialRoundTrip(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 0)
	a := Abc_NtkCreatePi(n)
	lo := Abc_NtkAddLatch(n, nil, ABC_INIT_ZERO)
	li := Abc_ObjFanin0(Abc_ObjFanin0(lo))
	Abc_ObjAddFanin(li, a)
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(po, lo)
	Abc_NtkOrderCisCos(n)
	graph := Abc_NtkToDar(n, 0, 1)
	result := Abc_NtkFromDarSeqSweep(n, graph)
	if Abc_NtkLatchNum(result) != 1 || Abc_NtkPiNum(result) != 1 || Abc_NtkPoNum(result) != 1 {
		t.Fatal("invalid sequential interface")
	}
	again := Abc_NtkToDar(result, 0, 1)
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		before, e1 := graph.Evaluate(in)
		after, e2 := again.Evaluate(in)
		if e1 != nil || e2 != nil || before[0] != after[0] || before[1] != after[1] {
			t.Fatalf("input %v before %v after %v", in, before, after)
		}
	}
}
