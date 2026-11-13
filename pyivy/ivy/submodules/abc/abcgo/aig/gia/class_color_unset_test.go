package gia

import "testing"

func TestUnsetRepresentativeColor(t *testing.T) {
	g := Gia_ManStart(4)
	x := Gia_ManAppendCi(g)
	g.StartRepresentatives()
	Gia_ObjSetColors(g, x>>1)
	Gia_ObjUnsetColor(g, x>>1, 0)
	if Gia_ObjColor(g, x>>1, 0) != 0 || Gia_ObjColor(g, x>>1, 1) != 1 {
		t.Fatal("wrong color cleared")
	}
}
