package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestBoothPartialProductMarking(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	product := gia.Gia_ManAppendAnd(p, a, b)
	marks := Acec_MultMarkPPs(p)
	if vec.Vec_BitEntry(marks, int(product>>1)) != 1 {
		t.Fatal("simple AND partial product not marked")
	}
	ignore := Acec_BoothFindPPG(p)
	if vec.Vec_BitEntry(ignore, int(product>>1)) != 1 {
		t.Fatal("partial product cone not marked")
	}
	for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
		if gia.Gia_ObjMark0(gia.Gia_ManObj(p, id)) != 0 {
			t.Fatalf("mark leaked on %d", id)
		}
	}
}
