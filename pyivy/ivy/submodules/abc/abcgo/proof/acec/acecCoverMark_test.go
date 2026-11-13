package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMarkFullAdderCone(t *testing.T) {
	p := gia.Gia_ManStart(64)
	in := []int32{gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)}
	out := make([]int32, 2)
	Acec_InsertFadd(p, in, out)
	sigs := []int32{in[0] >> 1, in[1] >> 1, in[2] >> 1, out[0] >> 1, out[1] >> 1}
	Gia_AcecMarkFadd(p, sigs)
	for _, id := range sigs[:3] {
		if gia.Gia_ObjMark0(gia.Gia_ManObj(p, id)) == 0 {
			t.Errorf("input %d unmarked", id)
		}
	}
	for _, id := range sigs[3:] {
		if gia.Gia_ObjMark1(gia.Gia_ManObj(p, id)) == 0 {
			t.Errorf("output %d unmarked", id)
		}
	}
}
func TestCollectRealXorCone(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManAppendXorReal(p, a, b)
	y := gia.Gia_ManAppendXorReal(p, x, c)
	gia.Gia_ManAppendCo(p, y)
	got := Gia_AcecCollectXors(p, nil)
	if vec.Vec_IntSize(got) != 2 || vec.Vec_IntEntry(got, 0) != y>>1 || vec.Vec_IntEntry(got, 1) != x>>1 {
		t.Fatalf("xors %v", vec.Vec_IntArray(got))
	}
}
