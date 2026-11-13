package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMatchBoxesSharedAndUnique(t *testing.T) {
	p := gia.Gia_ManStart(8)
	for i := 0; i < 3; i++ {
		gia.Gia_ManAppendCi(p)
	}
	a := &Acec_Box_t{pGia: p, vLeafLits: vec.Vec_WecStart(1), vRootLits: vec.Vec_WecStart(1)}
	b := &Acec_Box_t{pGia: p, vLeafLits: vec.Vec_WecStart(1), vRootLits: vec.Vec_WecStart(1)}
	for _, lit := range []int32{2, 4} {
		vec.Vec_WecPush(a.vLeafLits, 0, lit)
	}
	for _, lit := range []int32{2, 6} {
		vec.Vec_WecPush(b.vLeafLits, 0, lit)
	}
	if got := Acec_MatchBoxes(a, b); got != 1 {
		t.Fatalf("matched %d", got)
	}
	for _, box := range []*Acec_Box_t{a, b} {
		if vec.Vec_IntEntry(vec.Vec_WecEntry(box.vShared, 0), 0) != 2 {
			t.Fatal("wrong shared input")
		}
	}
	if vec.Vec_IntEntry(vec.Vec_WecEntry(a.vUnique, 0), 0) != 4 || vec.Vec_IntEntry(vec.Vec_WecEntry(b.vUnique, 0), 0) != 6 {
		t.Fatal("wrong unique inputs")
	}
}
