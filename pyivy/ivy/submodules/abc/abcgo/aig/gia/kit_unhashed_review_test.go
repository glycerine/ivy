package gia

import (
	"github.com/glycerine/abcgo/bool/kit"
	"testing"
)

func TestKitUnhashedSynthesisRepeatedLeaves(t *testing.T) {
	p := Gia_ManStart(16)
	defer Gia_ManStop(p)
	a := Gia_ManAppendCi(p)
	var scratch []uint32
	root := kit.Kit_TruthToGia(p, []uint32{0x88888888}, 2, &scratch, []int32{a, a}, 0)
	if root != a || p.nObjs != 2 {
		t.Fatalf("duplicate leaves root%d objs%d want root%d objs2", root, p.nObjs, a)
	}
}
func TestKitUnhashedAndUsesSourceLocalSimplification(t *testing.T) {
	for _, tc := range []struct {
		name    string
		b, want int32
	}{{"same", 2, 2}, {"opposite", 3, 0}, {"zero", 0, 0}, {"one", 1, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			p := Gia_ManStart(16)
			defer Gia_ManStop(p)
			a := Gia_ManAppendCi(p)
			if got := p.AndLiteral(a, tc.b, false); got != tc.want || p.nObjs != 2 {
				t.Fatalf("root%d objs%d want root%d objs2", got, p.nObjs, tc.want)
			}
		})
	}
}
func TestKitUnhashedAndRetainsDistinctGates(t *testing.T) {
	p := Gia_ManStart(16)
	defer Gia_ManStop(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	first, second := p.AndLiteral(a, b, false), p.AndLiteral(a, b, false)
	if first == second || Gia_ManAndNum(p) != 2 {
		t.Fatal("nonhashed synthesis unexpectedly merged gates")
	}
	p.fGiaSimple = 1
	before := p.nObjs
	p.AndLiteral(a, a, false)
	if p.nObjs != before+1 {
		t.Fatal("simple mode lost source raw append behavior")
	}
}
