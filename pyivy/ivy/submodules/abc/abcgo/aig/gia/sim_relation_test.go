package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSimulationRelation(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	p.VSimsPi = vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(p.VSimsPi, 0xaaaaaaaaaaaaaaaa)
	vec.Vec_WrdPush(p.VSimsPi, 0xcccccccccccccccc)
	values := Gia_ManSimPatValues(p)
	selected := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(selected, x>>1)
	rel := Gia_ManSimRel(p, selected, values)
	if rel == nil {
		t.Fatal("missing relation")
	}
	data := vec.Vec_WrdArray(rel)
	expected := vec.Vec_WrdEntry(values, 0)
	for pat := int32(0); pat < 64; pat++ {
		for mint := int32(0); mint < 2; mint++ {
			got := giaSimBit(data, 2*pat+mint)
			want := uint64(0)
			if (expected>>uint(pat))&1 == uint64(mint) {
				want = 1
			}
			if got != want {
				t.Fatalf("pattern %d mint %d: %d", pat, mint, got)
			}
		}
	}
	if Gia_ManObj(p, x>>1).fPhase != 0 {
		t.Fatal("temporary phase retained")
	}
}
