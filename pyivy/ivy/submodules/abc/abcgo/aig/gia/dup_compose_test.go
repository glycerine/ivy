package gia

import "testing"

func TestDupAppendNewSequential(t *testing.T) {
	first := Gia_ManStart(12)
	a, r := Gia_ManAppendCi(first), Gia_ManAppendCi(first)
	Gia_ManAppendCo(first, Gia_ManAppendAnd(first, a, r))
	Gia_ManAppendCo(first, a)
	Gia_ManSetRegNum(first, 1)
	second := Gia_ManStart(12)
	b, s := Gia_ManAppendCi(second), Gia_ManAppendCi(second)
	Gia_ManAppendCo(second, Gia_ManAppendAnd(second, b^1, s))
	Gia_ManAppendCo(second, b^1)
	Gia_ManSetRegNum(second, 1)
	out := Gia_ManDupAppendNew(first, second)
	if Gia_ManPiNum(out) != 1 || Gia_ManRegNum(out) != 2 || Gia_ManPoNum(out) != 2 {
		t.Fatal("wrong combined interface")
	}
	for bits := 0; bits < 8; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, e := out.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		x, e := first.Evaluate([]bool{input[0], input[1]})
		if e != nil {
			t.Fatal(e)
		}
		y, e := second.Evaluate([]bool{input[0], input[2]})
		if e != nil || got[0] != x[0] || got[1] != y[0] || got[2] != x[1] || got[3] != y[1] {
			t.Fatalf("composition %v %v %v %v", input, got, x, y)
		}
	}
}
func TestDupAppendAndShare(t *testing.T) {
	first := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(first), Gia_ManAppendCi(first)
	Gia_ManAppendCo(first, Gia_ManAppendAnd(first, a, b))
	second := Gia_ManStart(12)
	x, y := Gia_ManAppendCi(second), Gia_ManAppendCi(second)
	Gia_ManAppendCo(second, Gia_ManAppendAnd(second, x, y^1))
	both := Gia_ManDup2(first, second)
	if Gia_ManCiNum(both) != 2 || Gia_ManCoNum(both) != 2 {
		t.Fatal("shared interface")
	}
	shared := Gia_ManDup(first)
	Gia_ManDupAppendShare(shared, second)
	separate := Gia_ManDup(first)
	Gia_ManDupAppend(separate, second)
	for bits := 0; bits < 4; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0}
		want, e := both.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		got, e := shared.Evaluate(input)
		if e != nil || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("shared %v %v", got, e)
		}
		got, e = separate.Evaluate([]bool{input[0], input[1], input[0], input[1]})
		if e != nil || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("append %v %v", got, e)
		}
	}
}

func TestDupOntopAndNewPo(t *testing.T) {
	p := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b^1)
	q := Gia_ManStart(10)
	x, y := Gia_ManAppendCi(q), Gia_ManAppendCi(q)
	Gia_ManAppendCo(q, Gia_ManAppendAnd(q, x, y))
	cascade := Gia_ManDupOntop(p, q)
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, e := cascade.Evaluate(in)
		if e != nil || got[0] != (in[0] && !in[1]) {
			t.Fatalf("cascade %v %v", got, e)
		}
	}
	sequential := sequentialFixture()
	prop := Gia_ManStart(5)
	r := Gia_ManAppendCi(prop)
	Gia_ManAppendCo(prop, r^1)
	replaced := Gia_ManDupWithNewPo(sequential, prop)
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, e := replaced.Evaluate(in)
		if e != nil || got[0] == in[1] || got[1] != in[0] {
			t.Fatalf("property %v %v", got, e)
		}
	}
}

func TestDupWithConstraints(t *testing.T) {
	p := Gia_ManStart(10)
	a, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, r)
	Gia_ManAppendCo(p, a^1)
	Gia_ManSetRegNum(p, 1)
	out := Gia_ManDupWithConstraints(p, permutation(1, 0))
	if out.nConstrs != 1 || Gia_ManRegNum(out) != 1 {
		t.Fatal("constraint interface")
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != in[1] || got[1] == in[0] || got[2] == in[0] {
			t.Fatalf("constraint %v %v", got, e)
		}
	}
}
