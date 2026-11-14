package gia

import "testing"

func sequentialFixture() *Gia_Man_t {
	p := Gia_ManStart(12)
	a, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, r))
	Gia_ManAppendCo(p, a)
	Gia_ManSetRegNum(p, 1)
	return p
}
func TestDupSelfAndInvariantMiter(t *testing.T) {
	p := sequentialFixture()
	self := Gia_ManDupSelf(p)
	if Gia_ManPiNum(self) != 2 || Gia_ManRegNum(self) != 1 {
		t.Fatal("self loop interface")
	}
	for bits := 0; bits < 8; bits++ {
		ctrl, a, r := bits&1 != 0, bits&2 != 0, bits&4 != 0
		got, e := self.Evaluate([]bool{ctrl, a, r})
		next := r
		if ctrl {
			next = a
		}
		if e != nil || got[0] != (a && r) || got[1] != next {
			t.Fatalf("self loop %v %v", got, e)
		}
	}
	inv := Gia_ManStart(8)
	x := Gia_ManAppendCi(inv)
	Gia_ManAppendCo(inv, x)
	miter := Gia_ManDupInvMiter(p, inv)
	for bits := 0; bits < 4; bits++ {
		a, r := bits&1 != 0, bits&2 != 0
		got, e := miter.Evaluate([]bool{a, r})
		if e != nil || got[0] != (r && !a) {
			t.Fatalf("miter %v %v", got, e)
		}
	}
}
func TestDupAppendConesAndTimes(t *testing.T) {
	p := sequentialFixture()
	cone := Gia_ManStart(8)
	state := Gia_ManAppendCi(cone)
	Gia_ManAppendCo(cone, state^1)
	out := Gia_ManDupAppendCones(p, []*Gia_Man_t{cone}, 1, 1)
	if Gia_ManRegNum(out) != 1 || Gia_ManPoNum(out) != 2 {
		t.Fatal("appended cone interface")
	}
	for bits := 0; bits < 4; bits++ {
		a, r := bits&1 != 0, bits&2 != 0
		got, e := out.Evaluate([]bool{a, r})
		if e != nil || got[0] != (a && r) || got[1] == r || got[2] != a {
			t.Fatalf("appended cone %v %v", got, e)
		}
	}
	copies := Gia_ManDupTimes(p, 2)
	if Gia_ManPiNum(copies) != 2 || Gia_ManRegNum(copies) != 2 {
		t.Fatal("parallel interface")
	}
	for bits := 0; bits < 16; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
		got, e := copies.Evaluate(in)
		if e != nil || got[0] != (in[0] && in[2]) || got[1] != (in[1] && in[3]) || got[2] != in[0] || got[3] != in[1] {
			t.Fatalf("parallel %v %v", got, e)
		}
	}
}
func TestDupFlopClass(t *testing.T) {
	p := Gia_ManStart(16)
	a, r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, r0, r1))
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, r0)
	Gia_ManSetRegNum(p, 2)
	p.VFlopClasses = permutation(3, 5)
	out := Gia_ManDupFlopClass(p, 5)
	if Gia_ManPiNum(out) != 2 || Gia_ManRegNum(out) != 1 || Gia_ManPoNum(out) != 2 {
		t.Fatal("class interface")
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != (in[1] && in[2]) || got[1] != in[0] || got[2] != in[1] {
			t.Fatalf("class outputs %v %v", got, e)
		}
	}
}
