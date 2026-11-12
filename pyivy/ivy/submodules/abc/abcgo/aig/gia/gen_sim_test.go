package gia

import "testing"

func TestSimulateWordsInit(t *testing.T) {
	p := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a^1, b)
	out := Gia_ManAppendCo(p, x^1)
	in := Vec_WrdAlloc(4)
	for _, value := range []uint64{0x5555555555555555, 0xf0f0f0f0f0f0f0f0, 0x3333333333333333, 0xaaaaaaaaaaaaaaaa} {
		Vec_WrdPush(in, value)
	}
	Gia_ManSimulateWordsInit(p, in)
	if got, want := Gia_ManObjSim(p, out>>1), []uint64{^(^uint64(0x5555555555555555) & 0x3333333333333333), ^(^uint64(0xf0f0f0f0f0f0f0f0) & 0xaaaaaaaaaaaaaaaa)}; got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %x want %x", got, want)
	}
	Gia_ManSimulateWords(p, 3)
	if len(Gia_ManObjSim(p, out>>1)) != 3 {
		t.Fatal("random simulation size")
	}
}
