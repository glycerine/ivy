package gia

import "testing"

func TestRsb2ManagerCareAndLevels(t *testing.T) {
	// (a & b) XOR c, followed by an output.
	raw := []int32{0, 0, 0, 0, 0, 0, 0, 0, 2, 4, 8, 6, 10, 10}
	p := Gia_Rsb2ManAlloc()
	Gia_Rsb2ManStart(p, raw, 7, 20, 0, 1, 0, 0, 0)
	if got := Gia_Rsb2ManLevel(p); got != 2 || p.levels[4] != 1 || p.levels[5] != 2 {
		t.Fatalf("levels %v", p.levels)
	}
	care := Gia_Rsb2ManOdcs(p, 4)
	if care != ^word(0) {
		t.Fatalf("care %x", care)
	}
	if p.sims[8] != word(0xaaaaaaaaaaaaaaaa)&word(0xcccccccccccccccc) || p.sims[10] != p.sims[8]^word(0xf0f0f0f0f0f0f0f0) {
		t.Fatal("baseline simulation not restored")
	}
	if Gia_Rsb2ManDivs(p, 4) < 5 || p.nMffc != 1 {
		t.Fatalf("divisors %v mffc %d", p.divisorIDs, p.nMffc)
	}
	Gia_Rsb2ManFree(p)
}
