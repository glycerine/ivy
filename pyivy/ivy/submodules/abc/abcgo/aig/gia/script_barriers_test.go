package gia

import "testing"

func TestScriptBarrierRoundTrip(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	z := Gia_ManAppendBuf(p, x)
	Gia_ManAppendCo(p, z)
	split := Gia_ManDupFromBarBufs(p)
	if Gia_ManCiNum(split) != 3 || Gia_ManCoNum(split) != 2 {
		t.Fatal("barrier not split")
	}
	split.pSibls = make([]int32, split.nObjs+1)
	restored := Gia_ManDupToBarBufs(split, 1)
	if Gia_ManBufNum(restored) != 1 || Gia_ManCiNum(restored) != 2 || Gia_ManCoNum(restored) != 1 {
		t.Fatal("barrier not restored")
	}
	for _, in := range [][]bool{{false, false}, {false, true}, {true, false}, {true, true}} {
		aout, e := p.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		bout, e := restored.Evaluate(in)
		if e != nil || aout[0] != bout[0] {
			t.Fatalf("input %v outputs %v %v", in, aout, bout)
		}
	}
}
