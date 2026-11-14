package gia

import "testing"

func TestCecSimulationReleaseAll(t *testing.T) {
	g := Gia_ManStart(4)
	Gia_ManAppendCi(g)
	Gia_ManCecSimAlloc(g, 1)
	Gia_ManCecSaveCIWord(g, 1)
	Gia_ManCecSimFreeAll(g)
	if g.vSims != nil || g.VSimsPi != nil || g.NSimWords != 0 {
		t.Fatal("CEC buffers not released")
	}
}
