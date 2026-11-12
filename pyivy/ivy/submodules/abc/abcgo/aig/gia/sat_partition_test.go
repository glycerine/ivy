package gia

import "testing"

func TestSatPartitionCounts(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	for _, l := range []int32{a, b, c} {
		Gia_ManObj(p, l>>1).fMark0 = 1
	}
	x := Gia_ManAppendAnd(p, a, b)
	root := Gia_ManAppendAnd(p, x^1, c)
	obj := Gia_ManObj(p, root>>1)
	var leaves, nodes, on, off int32
	level := Gia_ManSatPartCount(p, obj, &leaves, &nodes)
	Gia_ManSatPartCountClauses(p, obj, &on, &off)
	if leaves != 3 || nodes != 2 || level != 1 || on != 2 || off != 2 || Gia_ManSatPartCountNodes(p, obj) != 2 {
		t.Fatalf("leaves=%d nodes=%d level=%d onset=%d offset=%d", leaves, nodes, level, on, off)
	}
	var n int32
	lits := make([]int32, 8)
	Gia_ManSatPartCollectSuper(p, obj, lits, &n)
	if n != 2 || !((lits[0] == x^1 && lits[1] == c) || (lits[1] == x^1 && lits[0] == c)) {
		t.Fatalf("supergate = %v", lits[:n])
	}
}
