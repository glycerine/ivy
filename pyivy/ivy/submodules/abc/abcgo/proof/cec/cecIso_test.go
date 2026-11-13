package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestIsomorphismMatchesUniqueColoredNodes(t *testing.T) {
	p := gia.Gia_ManStart(12)
	x, y := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	a := gia.Gia_ManAppendAnd(p, x, y)
	b := gia.Gia_ManAppendAnd(p, x, y)
	c := gia.Gia_ManAppendAnd(p, x^1, y)
	gia.Gia_ManAppendCo(p, a)
	p.StartRepresentatives()
	for _, id := range []int32{0, x >> 1, y >> 1} {
		gia.Gia_ObjSetColors(p, id)
	}
	gia.Gia_ObjSetColor(p, a>>1, 0)
	gia.Gia_ObjSetColor(p, b>>1, 1)
	gia.Gia_ObjSetColor(p, c>>1, 1)
	pairs := Cec_ManDetectIsomorphism(p)
	if pairs[a>>1] != b>>1 || pairs[b>>1] != a>>1 || pairs[c>>1] != 0 {
		t.Fatalf("unexpected pairing %v", pairs)
	}
	p.SetIsomorphism(pairs)
	pairs[a>>1] = 0
	Cec_ManTransformClasses(p)
	if got := gia.Gia_ObjRepr(p, b>>1); got != a>>1 {
		t.Fatalf("representative = %d, want %d", got, a>>1)
	}
	if got := gia.Gia_ObjNext(p, a>>1); got != b>>1 {
		t.Fatalf("next = %d, want %d", got, b>>1)
	}
}

func TestIsomorphismSimulationComplements(t *testing.T) {
	p := gia.Gia_ManStart(8)
	x, y := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gate := gia.Gia_ManAppendAnd(p, x^1, y)
	store := make([]uint32, int(gia.Gia_ManObjNum(p))*2)
	copy(Cec_ManIsoInfo(store, 2, x>>1), []uint32{0xaaaaaaaa, 0xffff0000})
	copy(Cec_ManIsoInfo(store, 2, y>>1), []uint32{0xcccccccc, 0xff00ff00})
	Gia_ManIsoSimulate(gia.Gia_ManObj(p, gate>>1), gate>>1, store, 2)
	got := Cec_ManIsoInfo(store, 2, gate>>1)
	if got[0] != 0x44444444 || got[1] != 0x0000ff00 {
		t.Fatalf("simulation %x", got)
	}
}

// cecIso.c requires both existing representative arrays and pIso.
func TestTransformClassesRequiresAllocatedClasses(t *testing.T) {
	p := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(p)
	p.SetIsomorphism(make([]int32, gia.Gia_ManObjNum(p)))
	defer func() {
		if recover() == nil {
			t.Fatal("missing class arrays accepted")
		}
	}()
	Cec_ManTransformClasses(p)
}
