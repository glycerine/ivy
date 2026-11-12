package gia

import (
	"math/bits"
	"testing"
)

func TestOneHotViolations(t *testing.T) {
	for n := int32(0); n <= 5; n++ {
		p := Gia_ManOneHot(1, n)
		for mask := 0; mask < 1<<n; mask++ {
			in := make([]bool, n+1)
			for i := int32(0); i < n; i++ {
				in[i+1] = mask&(1<<i) != 0
			}
			got, e := p.Evaluate(in)
			if e != nil {
				t.Fatal(e)
			}
			violation := false
			for _, v := range got {
				violation = violation || v
			}
			if violation != (bits.OnesCount(uint(mask)) != 1) {
				t.Fatalf("size %d mask %d outputs %v", n, mask, got)
			}
		}
	}
}
func TestDupLimitedCones(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	for _, out := range []*Gia_Man_t{Gia_ManDupAndConesLimit(p, []int32{y >> 1}, 1, 2), Gia_ManDupAndConesLimit2(p, []int32{y >> 1}, 1, 1)} {
		if Gia_ManCiNum(out) != 2 {
			t.Fatal("limited frontier")
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			got, e := out.Evaluate(in)
			if e != nil || got[0] != (in[0] && in[1]) {
				t.Fatalf("limited cone %v %v", got, e)
			}
		}
	}
}

func TestDupOneHotKeepsRegisterInputsLast(t *testing.T) {
	p := Gia_ManStart(15)
	a, r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, r0)
	Gia_ManAppendCo(p, r1)
	Gia_ManSetRegNum(p, 2)
	out := Gia_ManDupOneHot(p)
	if Gia_ManRegNum(out) != 2 || Gia_ManPoNum(out) < 2 {
		t.Fatal("one-hot interface")
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		got, e := out.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		if got[0] != in[0] || got[len(got)-2] != in[1] || got[len(got)-1] != in[2] {
			t.Fatalf("RI order %v", got)
		}
		violation := false
		for _, v := range got[1 : len(got)-2] {
			violation = violation || v
		}
		if violation != (bits.OnesCount(uint(mask>>1)) != 1) {
			t.Fatalf("constraints %v", got)
		}
	}
}
