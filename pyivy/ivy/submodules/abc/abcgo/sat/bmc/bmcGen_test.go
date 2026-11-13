package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestMoSimulation(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gate := gia.Gia_ManAppendAnd(p, a^1, b)
	gia.Gia_ManAppendCo(p, gate^1)
	out := gia.Gia_ObjId(p, gia.Gia_ManCo(p, 0))
	Gia_ManMoFindSimulate(p, 2)
	av, bv, gv, ov := Gia_ManMoObj(p, a>>1), Gia_ManMoObj(p, b>>1), Gia_ManMoObj(p, gate>>1), Gia_ManMoObj(p, out)
	for i := range gv {
		if gv[i] != ^av[i]&bv[i] || ov[i] != ^gv[i] {
			t.Fatalf("word %d: a=%x b=%x gate=%x out=%x", i, av[i], bv[i], gv[i], ov[i])
		}
	}
	before := append([]uint64(nil), av...)
	Gia_ManMoFindSimulate(p, 2)
	for i, v := range Gia_ManMoObj(p, a>>1) {
		if v != before[i] {
			t.Fatalf("non-deterministic simulation word %d", i)
		}
	}
}
func TestSatEnumeration(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a)
	if Gia_ManTestSatEnum(p) != 1 {
		t.Fatal("SAT enumeration failed")
	}
	if Gia_ManTestSatEnum(gia.Gia_ManStart(1)) != 0 {
		t.Fatal("missing output accepted")
	}
}
