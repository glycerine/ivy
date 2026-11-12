package gia

import "testing"

func TestMoveLastInputsAndRemap(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	and := Gia_ManAppendAnd(p, b, c^1)
	Gia_ManAppendCo(p, and)
	q := Gia_ManDupMoveLast(p, 1, 1)
	if Gia_ObjValue(Gia_ManObj(p, a>>1)) != 2 || Gia_ObjValue(Gia_ManObj(p, b>>1)) != 6 || Gia_ObjValue(Gia_ManObj(p, c>>1)) != 4 {
		t.Fatal("input permutation wrong")
	}
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, err := q.Evaluate(input)
		if err != nil || out[0] != (input[2] && !input[1]) {
			t.Fatalf("input=%v out=%v err=%v", input, out, err)
		}
	}
	src := Gia_ManStart(5)
	x := Gia_ManAppendCi(src)
	Gia_ManObj(src, x>>1).Value = uint32((b>>1)<<1 | 1)
	Gia_ManCleanupRemap(src, p)
	if Gia_ObjValue(Gia_ManObj(src, x>>1)) != 7 {
		t.Fatal("complemented remap failed")
	}
}
