package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestGlaClassHelpers(t *testing.T) {
	p := gia.Gia_ManStart(10)
	gia.Gia_ManAppendCi(p)
	obj := gia.Gia_ManObj(p, 0)
	p.SetGateClass(0, 1)
	if Gia_ObjIsInGla(p, obj) != 1 {
		t.Fatal("initial class")
	}
	Gia_ObjRemFromGla(p, obj)
	if Gia_ObjIsInGla(p, obj) != 0 {
		t.Fatal("remove")
	}
	Gia_ObjAddToGla(p, obj)
	if Gia_ObjIsInGla(p, obj) != 1 {
		t.Fatal("restore")
	}
}
