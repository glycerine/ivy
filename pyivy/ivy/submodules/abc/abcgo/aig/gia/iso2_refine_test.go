package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestIso2PrepareAndUniqueSignatures(t *testing.T) {
	g := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, a, c^1)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, y)
	p := Gia_Iso2ManStart(g)
	defer Gia_Iso2ManStop(p)
	Gia_Iso2ManPrepare(g)
	if Gia_ManObj(g, a>>1).Value != s_256Primes[ISO_MASK-1] {
		t.Fatal("primary input signature")
	}
	Gia_Iso2ManPropagate(g)
	if Gia_ManObj(g, x>>1).Value == Gia_ManObj(g, y>>1).Value {
		t.Fatal("fanin phases not reflected in signature")
	}
	if got := Gia_Iso2ManUniqify(p); got != 2 {
		t.Fatalf("singletons %d want 2", got)
	}
	if vec.Vec_IntSize(p.vTied) != 3 || p.nUniques != 3 {
		t.Fatalf("ties %v unique %d", vec.Vec_IntArray(p.vTied), p.nUniques)
	}
	for _, id := range vec.Vec_IntArray(p.vTied) {
		if Gia_ManObj(g, id).fMark0 != 0 {
			t.Fatal("tie mark leaked")
		}
	}
}
func TestIso2PerformGroupsEquivalentOutputs(t *testing.T) {
	g := Gia_ManStart(14)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x, y, z := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, c, d), Gia_ManAppendAnd(g, a, b^1)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, y)
	Gia_ManAppendCo(g, z)
	classes := Gia_Iso2ManPerform(g, 0)
	if vec.Vec_WecSize(classes) != 2 {
		t.Fatalf("class count %d", vec.Vec_WecSize(classes))
	}
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(classes, 0)); !reflect.DeepEqual(got, []int32{0, 1}) {
		t.Fatalf("equivalent class %v", got)
	}
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(classes, 1)); !reflect.DeepEqual(got, []int32{2}) {
		t.Fatalf("distinct class %v", got)
	}
}
