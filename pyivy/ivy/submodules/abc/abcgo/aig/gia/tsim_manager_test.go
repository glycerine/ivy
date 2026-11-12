package gia

import "testing"

func TestTernaryPackedValues(t *testing.T) {
	words := make([]uint32, 3)
	for i := int32(0); i < 40; i++ {
		Gia_ManTerSimInfoSet(words, i, 1+i%3)
	}
	for i := int32(0); i < 40; i++ {
		if got := Gia_ManTerSimInfoGet(words, i); got != 1+i%3 {
			t.Fatalf("index %d value %d", i, got)
		}
	}
}
func TestTernaryFrameTransfer(t *testing.T) {
	g := Gia_ManStart(8)
	in, reg := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, reg)
	Gia_ManAppendCo(g, in)
	Gia_ManSetRegNum(g, 1)
	p := Gia_ManTerCreate(g)
	defer Gia_ManTerDelete(p)
	Gia_ManTerSimInfoInit(p)
	Gia_ManTerSimulateRound(p)
	if Gia_ManTerSimInfoGet(p.pDataSimCos, 0) != GIA_ZER || Gia_ManTerSimInfoGet(p.pDataSimCos, 1) != GIA_UND {
		t.Fatal("initial ternary frame")
	}
	Gia_ManTerSimInfoTransfer(p)
	Gia_ManTerSimulateRound(p)
	if Gia_ManTerSimInfoGet(p.pDataSimCos, 0) != GIA_UND {
		t.Fatal("unknown input did not propagate to register")
	}
}
