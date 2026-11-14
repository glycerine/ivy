package gia

import "testing"

func TestSimulatorFrontierWords(t *testing.T) {
	g := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ro := Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b^1)
	y := Gia_ManAppendAnd(g, x, ro)
	Gia_ManAppendCo(g, y^1)
	Gia_ManAppendCo(g, x)
	Gia_ManSetRegNum(g, 1)
	var pars Gia_ParSim_t
	Gia_ManSimSetDefaultParams(&pars)
	pars.NWords = 2
	sim := Gia_ManSimCreate(g, &pars)
	defer Gia_ManSimDelete(sim)
	copy(Gia_SimDataCi(sim, 0), []uint32{0xaaaaaaaa, 0xffffffff})
	copy(Gia_SimDataCi(sim, 1), []uint32{0xcccccccc, 0})
	copy(Gia_SimDataCi(sim, 2), []uint32{0xffffffff, 0})
	Gia_ManSimulateRound(sim)
	want := []uint32{^(uint32(0xaaaaaaaa) &^ uint32(0xcccccccc)), ^uint32(0)}
	for i, v := range Gia_SimDataCo(sim, 0) {
		if v != want[i] {
			t.Fatalf("word %d got %08x want %08x", i, v, want[i])
		}
	}
	next := append([]uint32(nil), Gia_SimDataCo(sim, 1)...)
	Gia_ManSimInfoTransfer(sim)
	for i, v := range Gia_SimDataCi(sim, 2) {
		if v != next[i] {
			t.Fatal("register state was not transferred")
		}
	}
	var po, pat int32
	if Gia_ManCheckPos(sim, &po, &pat) != 1 || po != 0 || pat != 0 {
		t.Fatalf("failed output %d pattern %d", po, pat)
	}
}
