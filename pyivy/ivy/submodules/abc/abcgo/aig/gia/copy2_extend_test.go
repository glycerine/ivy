package gia

import "testing"

func TestExtendCopy2ArrayKeepsExistingValues(t *testing.T) {
	g := Gia_ManStart(8)
	x := Gia_ManAppendCi(g)
	Gia_ManExtendCopy2Array(g)
	Gia_ObjSetCopy2Array(g, x>>1, 17)
	y := Gia_ManAppendCi(g)
	Gia_ManExtendCopy2Array(g)
	if Gia_ObjCopy2Array(g, x>>1) != 17 || Gia_ObjCopy2Array(g, y>>1) != -1 {
		t.Fatal("extended copy array lost values")
	}
}
