package gia

import "testing"

func TestCecSimulationStorage(t *testing.T) {
	g := Gia_ManStart(8)
	Gia_ManAppendCi(g)
	Gia_ManCecSimAlloc(g, 2)
	if len(Gia_ManCecSimWords(g)) != 4 || g.SimulationWordCount() != 2 {
		t.Fatal("wrong dimensions")
	}
	Gia_ManCecSimWords(g)[2] = 17
	Gia_ManCecSetSimPattern(g, 65)
	Gia_ManCecSaveCIWord(g, 17)
	if Gia_ManCecSimPattern(g) != 65 || g.SimulationInputs()[0] != 17 {
		t.Fatal("simulation state was not retained")
	}
	Gia_ManCecSimFree(g)
	if g.SimulationWordCount() != 0 {
		t.Fatal("simulation not released")
	}
}
