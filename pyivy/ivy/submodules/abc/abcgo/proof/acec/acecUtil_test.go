package acec

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCollectLastXorStopsAtSharedNodes(t *testing.T) {
	for _, shared := range []bool{false, true} {
		p := gia.Gia_ManStart(32)
		a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		first := gia.Gia_ManAppendXorReal(p, a, b)
		last := gia.Gia_ManAppendXorReal(p, first, c)
		if shared {
			gia.Gia_ManAppendCo(p, first)
		}
		gia.Gia_ManAppendCo(p, last)
		got := vec.Vec_IntArray(Gia_PolynCollectLastXor(p, 0))
		want := []int32{last >> 1}
		if !shared {
			want = append(want, first>>1)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("shared=%v got %v want %v", shared, got, want)
		}
		if p.Refs != nil {
			t.Fatal("temporary reference counts retained")
		}
	}
}
