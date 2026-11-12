package gia

import "testing"

func TestSweeperExtractAndGraft(t *testing.T) {
	p := Gia_SweeperStart(nil)
	defer Gia_SweeperStop(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManHashAnd(p, a, b^1)
	Gia_ManObj(p, x>>1).Value = 99
	probe := Gia_SweeperProbeCreate(p, x^1)
	ids := Vec_IntAlloc(1)
	Vec_IntPush(ids, probe)
	q := Gia_SweeperExtractUserLogic(p, ids, nil, nil)
	if Gia_ManObj(p, x>>1).Value != 99 {
		t.Fatal("source node value not restored")
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != !(in[0] && !in[1]) {
			t.Fatalf("input=%v output=%v err=%v", in, out, err)
		}
	}
	out := Gia_SweeperGraft(p, nil, q)
	if Vec_IntSize(out) != 1 || Vec_IntEntry(out, 0) != x^1 {
		t.Fatalf("graft=%v", Vec_IntArray(out))
	}
}
