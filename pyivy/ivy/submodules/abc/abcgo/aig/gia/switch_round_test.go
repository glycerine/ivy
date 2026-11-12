package gia

import "testing"

func TestSwitchSimulationRound(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b^1)
	Gia_ManAppendCo(g, x^1)
	pars := &Gia_ParSwi_t{nWords: 1, fProbTrans: 0}
	s := Gia_ManSwiCreate(g, pars)
	defer Gia_ManSwiDelete(s)
	Gia_SwiDataCi(s, 0)[0] = 0xaaaaaaaa
	Gia_SwiDataCi(s, 1)[0] = 0xcccccccc
	Gia_ManSwiSimulateRound(s, 1)
	if got := Gia_SwiDataCo(s, 0)[0]; got != ^(uint32(0xaaaaaaaa) &^ uint32(0xcccccccc)) {
		t.Fatalf("output %08x", got)
	}
	if s.pData1[x>>1] != 8 {
		t.Fatalf("ones %d", s.pData1[x>>1])
	}
}
func TestSwitchRegisterTransfer(t *testing.T) {
	g := Gia_ManStart(8)
	pi, ro := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, ro)
	Gia_ManAppendCo(g, pi)
	Gia_ManSetRegNum(g, 1)
	s := Gia_ManSwiCreate(g, &Gia_ParSwi_t{nWords: 1})
	defer Gia_ManSwiDelete(s)
	Gia_SwiDataCi(s, 0)[0] = 0x12345678
	Gia_SwiDataCi(s, 1)[0] = 0x9abcdef0
	Gia_ManSwiSimulateRound(s, 0)
	Gia_ManSwiSimInfoTransfer(s, 0)
	if got := Gia_SwiDataCi(s, 1)[0]; got != 0x12345678 {
		t.Fatalf("register %08x", got)
	}
}
