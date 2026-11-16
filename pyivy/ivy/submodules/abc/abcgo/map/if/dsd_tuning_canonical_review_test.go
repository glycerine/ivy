package if_

import (
	"github.com/glycerine/abcgo/opt/dau"
	"testing"
)

func TestDsdCellTuningVisitsCanonicalBaseNodes(t *testing.T) {
	p := If_DsdManAlloc(2, 0)
	defer If_DsdManFree(p, 0)
	Id_DsdManTuneStr1(p, "{ab}", 0, 0)
	for _, lit := range []int32{0, 2} {
		if If_DsdManReadMark(p, lit) != 0 {
			t.Errorf("canonical base literal%d rejected by universal 2-LUT", lit)
		}
	}
}

func TestDsdCellTuningVisitsCanonicalInternalNodes(t *testing.T) {
	p := If_DsdManAlloc(4, 0)
	defer If_DsdManFree(p, 0)
	If_DsdManCompute(p, dau.Dau_DsdToTruth("((ab)[cd])", 4), 4, make([]byte, 4), "")
	Id_DsdManTuneStr1(p, "{abcd}", 0, 0)
	for id := int32(0); id < If_DsdManObjNum(p); id++ {
		if If_DsdManReadMark(p, id<<1) != 0 {
			t.Errorf("canonical node%d rejected by universal 4-LUT", id)
		}
	}
}

func TestDsdTwoLutTuningUsesInternalCanonicalTruth(t *testing.T) {
	p := If_DsdManAlloc(11, 0)
	defer If_DsdManFree(p, 0)
	// AND ten inputs, then XOR the eleventh: two 6-LUTs suffice.
	// Its ten-input AND is an interned internal node, never a Compute root.
	If_DsdManCompute(p, dau.Dau_DsdToTruth("[(abcdefghij)k]", 11), 11, make([]byte, 11), "")
	If_DsdManTune(p, 6, 0, 0, 0, 0)
	for id := int32(0); id < If_DsdManObjNum(p); id++ {
		if If_DsdManReadMark(p, id<<1) != 0 {
			t.Errorf("canonical node%d should fit two 6-LUTs", id)
		}
	}
}
