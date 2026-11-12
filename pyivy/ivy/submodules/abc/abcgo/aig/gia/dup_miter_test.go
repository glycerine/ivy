package gia

import "testing"

func TestMiterAndTransforms(t *testing.T) {
	a := Gia_ManStart(12)
	x, y := Gia_ManAppendCi(a), Gia_ManAppendCi(a)
	Gia_ManAppendCo(a, Gia_ManAppendAnd(a, x, y))
	b := Gia_ManStart(12)
	u, v := Gia_ManAppendCi(b), Gia_ManAppendCi(b)
	Gia_ManAppendCo(b, Gia_ManAppendAnd(b, u, v^1))
	normal := Gia_ManMiter(a, b, 0, 0, 0, 0, 0)
	dual := Gia_ManMiter(a, b, 0, 1, 0, 0, 0)
	implic := Gia_ManMiter(a, b, 0, 0, 0, 1, 0)
	converted := Gia_ManTransformMiter(dual)
	reduced := Gia_ManDupAndOr(dual, 2, 1, 0)
	for bits := 0; bits < 4; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0}
		lv, rv := input[0] && input[1], input[0] && !input[1]
		for _, out := range []*Gia_Man_t{normal, converted} {
			got, e := out.Evaluate(input)
			if e != nil || got[0] != (lv != rv) {
				t.Fatalf("miter %v %v", got, e)
			}
		}
		got, e := dual.Evaluate(input)
		if e != nil || got[0] != lv || got[1] != rv {
			t.Fatalf("dual %v %v", got, e)
		}
		got, e = implic.Evaluate(input)
		if e != nil || got[0] != (lv && !rv) {
			t.Fatalf("implication %v %v", got, e)
		}
		got, e = reduced.Evaluate(input)
		if e != nil || got[0] != (lv || rv) || got[1] != got[0] {
			t.Fatalf("reduce %v %v", got, e)
		}
	}
}
func TestSequentialMiter(t *testing.T) {
	left, right := sequentialFixture(), sequentialFixture()
	out := Gia_ManMiter(left, right, 0, 0, 1, 0, 0)
	if out == nil || Gia_ManPiNum(out) != 1 || Gia_ManRegNum(out) != 2 || Gia_ManPoNum(out) != 1 {
		t.Fatal("miter interface")
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != ((in[0] && in[1]) != (in[0] && in[2])) || got[1] != in[0] || got[2] != in[0] {
			t.Fatalf("sequential %v %v", got, e)
		}
	}
}
func TestTwoWordAndDualTransforms(t *testing.T) {
	p := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b)
	Gia_ManAppendCo(p, a^1)
	Gia_ManAppendCo(p, b^1)
	m := Gia_ManTransformMiter2(p)
	d := Gia_ManTransformTwoWord2DualOutput(p)
	z := Gia_ManTransformToDual(m)
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, e := m.Evaluate(in)
		if e != nil || !got[0] || !got[1] {
			t.Fatalf("miter2 %v %v", got, e)
		}
		got, e = d.Evaluate(in)
		if e != nil || got[0] != in[0] || got[1] == in[0] || got[2] != in[1] || got[3] == in[1] {
			t.Fatalf("interleaved %v %v", got, e)
		}
		got, e = z.Evaluate(in)
		if e != nil || !got[0] || got[1] || !got[2] || got[3] {
			t.Fatalf("dual %v %v", got, e)
		}
	}
}

func TestDualOutputReorder(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, a)
	out := Gia_ManTransformDualOutput(p)
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != in[0] || got[1] != (in[0] && in[1] && in[2]) {
			t.Fatalf("swap %v %v", got, e)
		}
	}
}
