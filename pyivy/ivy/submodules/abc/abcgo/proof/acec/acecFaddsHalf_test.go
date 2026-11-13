package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDetectHalfAdderRepresentations(t *testing.T) {
	for _, real := range []bool{false, true} {
		p := gia.Gia_ManStart(32)
		a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		var x, carry int32
		if real {
			x = gia.Gia_ManAppendXorReal(p, a, b)
			carry = gia.Gia_ManAppendAnd(p, a, b)
		} else {
			gia.Gia_ManHashStart(p)
			x = gia.Gia_ManHashXor(p, a, b)
			carry = gia.Gia_ManHashAnd(p, a, b)
			gia.Gia_ManHashStop(p)
		}
		got := Gia_ManDetectHalfAdders(p, 0)
		found := false
		for i := 0; i+1 < vec.Vec_IntSize(got); i += 2 {
			if vec.Vec_IntEntry(got, i) == x>>1 && vec.Vec_IntEntry(got, i+1) == carry>>1 {
				found = true
			}
		}
		if !found {
			t.Fatalf("real=%v XOR=%d carry=%d pairs=%v", real, x>>1, carry>>1, vec.Vec_IntArray(got))
		}
	}
}
