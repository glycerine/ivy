package gia

import "testing"

func TestTernarySimulationFindsToggleLoop(t *testing.T) {
	p := Gia_ManStart(8)
	r := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, r)
	Gia_ManAppendCo(p, r^1)
	Gia_ManSetRegNum(p, 1)
	s := Gia_ManTerSimulate(p, 0)
	defer Gia_ManTerDelete(s)
	if p.nTerStates != 3 || p.nTerLoop != 0 {
		t.Fatalf("states=%d loop=%d", p.nTerStates, p.nTerLoop)
	}
	want := []int32{GIA_ZER, GIA_ONE, GIA_ZER}
	for i, value := range want {
		state := Vec_PtrEntry(s.vStates, i).([]uint32)
		if got := Gia_ManTerSimInfoGet(state, 0); got != value {
			t.Fatalf("state %d=%d", i, got)
		}
	}
	if s.pCount0[0] != 2 || s.pCountX[0] != 0 {
		t.Fatal("state counters")
	}
}
