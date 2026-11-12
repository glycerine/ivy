package gia

import "testing"

func TestConditionalSweeperReducesOnlyUnderCare(t *testing.T) {
	p := Gia_SweeperStart(nil)
	defer Gia_SweeperStop(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManHashAnd(p, a, b)
	probe := Gia_SweeperProbeCreate(p, x)
	cond := Gia_SweeperProbeCreate(p, a)
	outputs := Vec_IntAlloc(1)
	Vec_IntPush(outputs, probe)
	Gia_SweeperCondPush(p, cond)
	q := Gia_SweeperSweep(p, outputs, 1, 1000, 1, 0)
	if Gia_ManAndNum(q) != 0 {
		t.Fatal("conditional constant was not reduced")
	}
	for _, bv := range []bool{false, true} {
		out, err := q.Evaluate([]bool{false, bv})
		if err != nil || out[0] {
			t.Fatalf("care result %v %v", out, err)
		}
	}
	Gia_SweeperCondPop(p)
	q = Gia_SweeperSweep(p, outputs, 1, 1000, 1, 0)
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] && in[1]) {
			t.Fatalf("unconstrained input %v output %v err %v", in, out, err)
		}
	}
}
