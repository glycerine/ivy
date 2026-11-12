package gia

import "testing"

func TestZeroUndcAcrossFrames(t *testing.T) {
	p := Gia_ManStart(12)
	a, r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, r0)
	Gia_ManAppendCo(p, r1)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, r0)
	Gia_ManSetRegNum(p, 2)
	out := Gia_ManDupZeroUndc(p, "1x", 0, 0, 0)
	if Gia_ManPiNum(out) != 2 || Gia_ManRegNum(out) != 3 {
		t.Fatal("wrong undc interface")
	}
	for bits := 0; bits < 4; bits++ {
		a, u := bits&1 != 0, bits&2 != 0
		frame, e := out.Evaluate([]bool{a, u, false, false, false})
		if e != nil || !frame[0] || frame[1] != u || frame[2] == a || !frame[3] || !frame[4] {
			t.Fatalf("frame 0 %v %v", frame, e)
		}
		next, e := out.Evaluate([]bool{false, !u, frame[2], frame[3], frame[4]})
		if e != nil || next[0] != a || !next[1] {
			t.Fatalf("frame 1 %v %v", next, e)
		}
	}
}
func TestMiter2SharedAndIndependentInputs(t *testing.T) {
	p := sequentialFixture()
	for _, label := range []string{"x0", "X0"} {
		out := Gia_ManMiter2(p, label, 0)
		if Gia_ManRegNum(out) != 2 {
			t.Fatal("miter registers")
		}
		pis := Gia_ManPiNum(out)
		wantPis := int32(1)
		if label[0] == 'X' {
			wantPis = 2
		}
		if pis != wantPis {
			t.Fatalf("input count %d", pis)
		}
		first := make([]bool, pis+2)
		first[0] = true
		if pis == 2 {
			first[1] = false
		}
		f0, e := out.Evaluate(first)
		if e != nil || f0[0] {
			t.Fatalf("frame 0 %v %v", f0, e)
		}
		next := make([]bool, pis+2)
		next[0] = true
		if pis == 2 {
			next[1] = true
		}
		copy(next[pis:], f0[1:])
		f1, e := out.Evaluate(next)
		if e != nil || f1[0] != (label[0] == 'X') {
			t.Fatalf("frame 1 %s %v %v", label, f1, e)
		}
	}
}
