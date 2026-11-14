package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func permutation(values ...int32) *Vec_Int_t {
	v := vec.Vec_IntAlloc(len(values))
	for _, x := range values {
		vec.Vec_IntPush(v, x)
	}
	return v
}
func TestDupInputAndFlopPermutation(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, r0))
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, b, r1))
	Gia_ManAppendCo(p, a^1)
	Gia_ManAppendCo(p, b)
	Gia_ManSetRegNum(p, 2)
	pi := Gia_ManDupPerm(p, permutation(1, 0))
	regs := Gia_ManDupPermFlop(p, permutation(1, 0))
	gap := Gia_ManDupPermFlopGap(p, permutation(1, -1, 0))
	if Gia_ManRegNum(gap) != 3 {
		t.Fatal("gap lost register")
	}
	for bits := 0; bits < 16; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
		want, e := p.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		got, e := pi.Evaluate([]bool{in[1], in[0], in[2], in[3]})
		if e != nil {
			t.Fatal(e)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatal("PI permutation", in)
			}
		}
		got, e = regs.Evaluate([]bool{in[0], in[1], in[3], in[2]})
		if e != nil || got[0] != want[0] || got[1] != want[1] || got[2] != want[3] || got[3] != want[2] {
			t.Fatalf("register permutation %v %v", got, e)
		}
		got, e = gap.Evaluate([]bool{in[0], in[1], in[3], true, in[2]})
		if e != nil || got[0] != want[0] || got[1] != want[1] || got[2] != want[3] || got[3] || got[4] != want[2] {
			t.Fatalf("register gap %v %v", got, e)
		}
	}
	zero := Gia_ManDupZero(p)
	got, e := zero.Evaluate([]bool{true, true, true, true})
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range got {
		if v {
			t.Fatal("zero graph")
		}
	}
}
func TestDupDropUnusedInputs(t *testing.T) {
	p := Gia_ManStart(10)
	Gia_ManAppendCi(p)
	b := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, b)
	out := Gia_ManDupLastPis(p, 1)
	got, e := out.Evaluate([]bool{true})
	if e != nil || !got[0] {
		t.Fatalf("last pi %v %v", got, e)
	}
	q := Gia_ManStart(10)
	a := Gia_ManAppendCi(q)
	Gia_ManAppendCi(q)
	Gia_ManAppendCo(q, a)
	out = Gia_ManDupRemovePis(q, 1)
	got, e = out.Evaluate([]bool{true})
	if e != nil || !got[0] {
		t.Fatalf("removed pi %v %v", got, e)
	}
}

func TestDupOnsetOffsetAndAttributes(t *testing.T) {
	p := Gia_ManStart(10)
	a := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	p.VConfigs = permutation(10, 20)
	p.VMapping = permutation(0, 1, 2)
	p.VPacking = permutation(1, 2)
	p.PCellStr = "cell"
	p.nAnd2Delay = 5
	copy := Gia_ManDupWithAttributes(p)
	vec.Vec_IntWriteEntry(copy.VConfigs, 0, 99)
	vec.Vec_IntWriteEntry(copy.VMapping, 0, 99)
	vec.Vec_IntWriteEntry(copy.VPacking, 0, 99)
	if vec.Vec_IntEntry(p.VConfigs, 0) != 10 || vec.Vec_IntEntry(p.VMapping, 0) != 0 || vec.Vec_IntEntry(p.VPacking, 0) != 1 || copy.PCellStr != "cell" || copy.nAnd2Delay != 5 {
		t.Fatal("attributes not independently copied")
	}
	dual := Gia_ManDupOnsetOffset(p)
	for _, value := range []bool{false, true} {
		got, e := dual.Evaluate([]bool{value})
		if e != nil || len(got) != 2 || got[0] != value || got[1] == value {
			t.Fatalf("dual outputs %v %v", got, e)
		}
	}
}

func TestDupPiPermDeterministic(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
	one, two := Gia_ManDupPiPerm(p), Gia_ManDupPiPerm(p)
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		a, e := one.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		b, e := two.Evaluate(in)
		if e != nil || a[0] != b[0] {
			t.Fatalf("nondeterministic mutation %v %v", a, b)
		}
	}
}
