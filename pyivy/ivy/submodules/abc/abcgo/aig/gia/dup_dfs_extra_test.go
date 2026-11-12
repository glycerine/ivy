package gia

import "testing"

func TestDupDfsVariants(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, r^1)
	Gia_ManAppendCo(p, a)
	Gia_ManSetRegNum(p, 1)
	dfs := Gia_ManDupDfs(p)
	dfs2 := Gia_ManDupDfs2(p)
	one := Gia_ManDupDfsOnePo(p, 1)
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, e := p.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		for _, out := range []*Gia_Man_t{dfs, dfs2} {
			got, e := out.Evaluate(in)
			if e != nil || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
				t.Fatalf("dfs %v %v", got, e)
			}
		}
		got, e := one.Evaluate(in)
		if e != nil || got[0] != want[1] || got[1] != want[2] {
			t.Fatalf("one %v %v", got, e)
		}
	}
	p.pObjs[Gia_ManPo(p, 0).id].fMark1 = 1
	skip := Gia_ManDupDfsSkip(p)
	if Gia_ManPoNum(skip) != 1 || Gia_ManRegNum(skip) != 1 {
		t.Fatal("incorrect skipped outputs")
	}
	lits := permutation(x^1, a)
	out := Gia_ManDupDfsLitArray(p, lits)
	if Gia_ManCiNum(out) != 3 {
		t.Fatal("input count changed")
	}
	got, e := out.Evaluate([]bool{true, true, false})
	if e != nil || got[0] || !got[1] {
		t.Fatalf("literals %v %v", got, e)
	}
}
func TestDupDfsOldRepresentatives(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, y)
	p.pReprsOld = make([]int32, p.nObjs)
	for i := range p.pReprsOld {
		p.pReprsOld[i] = -1
	}
	p.pReprsOld[y>>1] = x >> 1
	out := Gia_ManDupDfs2(p)
	if Gia_ManAndNum(out) != 1 {
		t.Fatal("representative was not reused")
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != (in[0] && in[1]) {
			t.Fatalf("repr %v %v", got, e)
		}
	}
}

func TestDupDfsCiMap(t *testing.T) {
	p := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	out := Gia_ManDupDfsCiMap(p, []int32{-1, a ^ 1}, nil)
	for bits := 0; bits < 4; bits++ {
		got, e := out.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
		if e != nil || got[0] {
			t.Fatalf("substitution %v %v", got, e)
		}
	}
}
