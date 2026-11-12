package gia

import "testing"

func TestTrimConstantOutputsAndUnusedInputs(t *testing.T) {
	p := Gia_ManStart(12)
	a := Gia_ManAppendCi(p)
	Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, 0)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, 1)
	out := Gia_ManDupTrimmed(p, 1, 1, 0, 0)
	if Gia_ManPiNum(out) != 1 || Gia_ManPoNum(out) != 2 {
		t.Fatalf("trim interface %d %d", Gia_ManPiNum(out), Gia_ManPoNum(out))
	}
	for _, v := range []bool{false, true} {
		got, e := out.Evaluate([]bool{v})
		if e != nil || got[0] != v || !got[1] {
			t.Fatalf("trim %v %v", got, e)
		}
	}
	only := Gia_ManDupTrimmed2(p)
	if Gia_ManPiNum(only) != 1 || Gia_ManPoNum(only) != 1 {
		t.Fatal("dummy interface")
	}
	got, e := only.Evaluate([]bool{true})
	if e != nil || got[0] {
		t.Fatalf("dummy output %v %v", got, e)
	}
}
func TestTrimDualConstantPair(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b)
	out := Gia_ManDupTrimmed(p, 0, 1, 1, 0)
	if Gia_ManPoNum(out) != 2 {
		t.Fatalf("expected one pair, got %d", Gia_ManPoNum(out))
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != in[0] || got[1] != in[1] {
			t.Fatalf("trim dual %v %v", got, e)
		}
	}
}
