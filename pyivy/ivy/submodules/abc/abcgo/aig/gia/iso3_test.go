package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestIso3SynchronousSignatureUpdate(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b^1)
	output := Gia_ManAppendCo(g, root)
	Gia_Iso3Init(g)
	signatures := Gia_Iso3Save(g)
	if Gia_Iso3Unique(signatures) != 4 {
		t.Fatalf("initial unique count %d", Gia_Iso3Unique(signatures))
	}
	old := append([]int32(nil), vec.Vec_IntArray(signatures)...)
	Gia_Iso3Compute(g, signatures)
	expected := uint32(old[root>>1]) + uint32(old[a>>1]) + Iso_Compl[0] + Iso_Fanio[0] + uint32(old[b>>1]) + Iso_Compl[1] + Iso_Fanio[0] + uint32(old[output>>1]) + Iso_Compl[0] + Iso_Fanio[1]
	if got := Gia_ManObj(g, root>>1).Value; got != expected {
		t.Fatalf("root signature %#x want %#x", got, expected)
	}
	if !reflect.DeepEqual(vec.Vec_IntArray(signatures), old) {
		t.Fatal("signature snapshot mutated")
	}
}
func TestIso4ReverseSignaturesAreRepeatable(t *testing.T) {
	build := func() *Gia_Man_t {
		g := Gia_ManStart(8)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		root := Gia_ManAppendAnd(g, a, b^1)
		Gia_ManAppendCo(g, root)
		return g
	}
	a, b := build(), build()
	levelsA, levelsB := Gia_Iso4Gia(a), Gia_Iso4Gia(b)
	if vec.Vec_WecSize(levelsA) != vec.Vec_WecSize(levelsB) {
		t.Fatal("reverse level count differs")
	}
	for id := int32(0); id < Gia_ManObjNum(a); id++ {
		if Gia_ManObj(a, id).Value != Gia_ManObj(b, id).Value {
			t.Fatalf("signature differs at %d", id)
		}
	}
}
