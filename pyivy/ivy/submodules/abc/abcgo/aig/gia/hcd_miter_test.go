package gia

import "testing"

func TestHcdChoiceMiterAndTfi(t *testing.T) {
	graphs := Vec_PtrAlloc(2)
	for i := 0; i < 2; i++ {
		g := Gia_ManStart(8)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		Gia_ManAppendCo(g, x^int32(i))
		Vec_PtrPush(graphs, g)
	}
	p := Hcd_ManChoiceMiter(graphs)
	for bits := 0; bits < 4; bits++ {
		got, err := p.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
		if err != nil || got[0] != (bits == 3) || got[1] == got[0] {
			t.Fatalf("input=%d got=%v err=%v", bits, got, err)
		}
	}
	g := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	if Hcd_ObjCheckTfi(g, Gia_ManObj(g, x>>1), Gia_ManObj(g, y>>1)) != 1 || Hcd_ObjCheckTfi(g, Gia_ManObj(g, y>>1), Gia_ManObj(g, x>>1)) != 0 {
		t.Fatal("incorrect TFI")
	}
	for _, obj := range g.pObjs {
		if obj.fMark0 != 0 {
			t.Fatal("TFI left marks")
		}
	}
}
