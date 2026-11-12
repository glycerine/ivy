package gia

import (
	"reflect"
	"testing"
)

func TestSwitchManagerWords(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	var pars Gia_ParSwi_t
	Gia_ManSetDefaultParamsSwi(&pars)
	pars.nWords = 2
	p := Gia_ManSwiCreate(g, &pars)
	defer Gia_ManSwiDelete(p)
	xs := Gia_SwiDataCi(p, 0)
	ys := Gia_SwiDataCi(p, 1)
	Gia_ManSwiSimInfoOne(p, xs)
	Gia_ManSwiSimInfoZero(p, ys)
	Gia_ManSwiSimInfoCopy(p, ys, xs)
	if !reflect.DeepEqual(ys, []uint32{^uint32(0), ^uint32(0)}) {
		t.Fatal("copy")
	}
	Gia_ManSwiSimInfoZero(p, ys)
	Gia_ManSwiSimInfoCopyShift(p, ys, xs)
	if !reflect.DeepEqual(ys, []uint32{0xffff, 0xffff}) {
		t.Fatal("shifted copy")
	}
	Gia_ManRandom(1)
	Gia_ManSwiSimInfoRandom(p, ys, 1)
	if ys[0] != ys[1] {
		t.Fatal("expected shared transition mask")
	}
}
