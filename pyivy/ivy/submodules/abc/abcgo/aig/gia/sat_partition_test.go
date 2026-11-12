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

func TestSatPartitionStore(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	for _, l := range []int32{a, b, c} {
		Gia_ManObj(p, l>>1).fMark0 = 1
	}
	x := Gia_ManAppendAnd(p, a, b)
	root := Gia_ManAppendAnd(p, x^1, c)
	store := make([]uint32, 16)
	n := Gia_ManSatPartCreate(p, Gia_ManObj(p, root>>1), store)
	if n != 5 || store[0]&0xff != 4 || (store[0]>>8)&0xff != 1 {
		t.Fatalf("root record %v", store[:n])
	}
	seen := map[int32]bool{}
	var walk func(int)
	walk = func(i int) {
		v := store[i]
		if v&1 != 0 {
			seen[int32(v>>1)] = true
			return
		}
		for j := 0; j < int((v&0xff)>>1); j++ {
			walk(i + int((v>>8)&0xff) + j)
		}
	}
	walk(0)
	if len(seen) != 3 || !seen[a] || !seen[b] || !seen[c] {
		t.Fatalf("terminal literals %v", seen)
	}
}
