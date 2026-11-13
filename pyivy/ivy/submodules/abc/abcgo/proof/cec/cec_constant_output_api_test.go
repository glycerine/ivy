package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec2AndCec3DirectOutputSimulationMatchesSourcePhase(t *testing.T) {
	for _, engine := range []int{2, 3} {
		g := gia.Gia_ManStart(4)
		gia.Gia_ManAppendCo(g, 1)
		Cec2_ManSimAlloc(g, 1)
		var result int32
		if engine == 2 {
			result = Cec2_ManSimulateCos(g)
		} else {
			result = Cec3_ManSimulateCos(g)
		}
		if result != 1 || g.SequentialCounterexample() != nil {
			t.Fatalf("engine %d did not preserve source phase normalization", engine)
		}
	}
}
