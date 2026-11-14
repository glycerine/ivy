package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestBuiltInPatternPackingAndGrowth(t *testing.T) {
	p := Gia_ManStart(16)
	for i := 0; i < 7; i++ {
		Gia_ManAppendCi(p)
	}
	Gia_ManBuiltInSimStart(p, 1, 16)
	pattern := vec.Vec_IntAlloc(7)
	for value := 0; value < 65; value++ {
		vec.Vec_IntClear(pattern)
		for i := 0; i < 7; i++ {
			lit := int32(2 * i)
			if value&(1<<i) == 0 {
				lit++
			}
			vec.Vec_IntPush(pattern, lit)
		}
		Gia_ManBuiltInSimAddPat(p, pattern)
	}
	if p.NSimWords != 2 || p.iPatsPi != 65 {
		t.Fatalf("words=%d patterns=%d", p.NSimWords, p.iPatsPi)
	}
	if Gia_ManBuiltInSimPack(p, pattern) != 64 {
		t.Fatal("last pattern not retained")
	}
	for i := 0; i < 7; i++ {
		if giaSimBit(Gia_ManBuiltInDataPi(p, int32(i)), 64) != 1 || giaSimBit(Gia_ManBuiltInData(p, int32(i+1)), 64) != uint64((64>>i)&1) {
			t.Fatalf("incorrect input %d", i)
		}
	}
	Gia_ManBuiltInSimAddPat(p, pattern)
	if p.iPatsPi != 65 {
		t.Fatal("compatible pattern was not reused")
	}
}

func TestBuiltInPatternReplacement(t *testing.T) {
	p := Gia_ManStart(16)
	for i := 0; i < 7; i++ {
		Gia_ManAppendCi(p)
		vec.Vec_IntPush(&p.vSuppVars, int32(i))
	}
	Gia_ManBuiltInSimStart(p, 1, 16)
	p.nSimWordsMax = 1
	pattern := vec.Vec_IntAlloc(7)
	for value := 0; value < 65; value++ {
		vec.Vec_IntClear(pattern)
		for i := 0; i < 7; i++ {
			lit := int32(2 * i)
			if value&(1<<i) == 0 {
				lit++
			}
			vec.Vec_IntPush(pattern, lit)
		}
		Gia_ManBuiltInSimAddPat(p, pattern)
	}
	if p.iPatsPi != 64 || p.iPastPiMax != 1 || Gia_ManBuiltInSimPack(p, pattern) != 0 {
		t.Fatal("full pattern buffer was not replaced")
	}
}
