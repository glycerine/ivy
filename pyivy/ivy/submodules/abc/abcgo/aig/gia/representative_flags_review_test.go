package gia

import (
	"testing"
	"unsafe"
)

func TestRepresentativeFlagsRemainIndependent(t *testing.T) {
	t.Logf("representative record: %d bytes", unsafe.Sizeof(Gia_Rpr_t{}))
	g := Gia_ManStart(1)
	Gia_ManAppendCi(g)
	Gia_ManAppendCi(g)
	g.StartRepresentatives()
	for flags := int32(0); flags < 16; flags++ {
		g.PReprs[2] = Gia_Rpr_t{iRepr: GIA_VOID}
		Gia_ObjSetRepr(g, 2, 1)
		if flags&1 != 0 {
			Gia_ObjSetProved(g, 2)
		}
		if flags&2 != 0 {
			Gia_ObjSetFailed(g, 2)
		}
		if flags&4 != 0 {
			Gia_ObjSetColor(g, 2, 0)
		}
		if flags&8 != 0 {
			Gia_ObjSetColor(g, 2, 1)
		}
		if Gia_ObjProved(g, 2) != flags&1 || Gia_ObjFailed(g, 2) != (flags>>1)&1 || Gia_ObjColors(g, 2) != flags>>2 || Gia_ObjRepr(g, 2) != 1 {
			t.Fatalf("flag combination %d was not retained", flags)
		}
		before := g.PReprs[2]
		Gia_ObjUnsetRepr(g, 2)
		Gia_ObjUnsetProved(g, 2)
		Gia_ObjUnsetColor(g, 2, 0)
		if Gia_ObjRepr(g, 2) != GIA_VOID || Gia_ObjProved(g, 2) != 0 || Gia_ObjColor(g, 2, 0) != 0 || Gia_ObjFailed(g, 2) != (flags>>1)&1 || Gia_ObjColor(g, 2, 1) != (flags>>3)&1 {
			t.Fatalf("clearing fields changed other flags in combination %d", flags)
		}
		g.PReprs[2] = before
		if Gia_ObjProved(g, 2) != flags&1 || Gia_ObjColors(g, 2) != flags>>2 || Gia_ObjRepr(g, 2) != 1 {
			t.Fatal("value copy did not preserve representative metadata")
		}
		if g.PReprs[1] != (Gia_Rpr_t{iRepr: GIA_VOID}) {
			t.Fatal("flags leaked into another record")
		}
	}
}
