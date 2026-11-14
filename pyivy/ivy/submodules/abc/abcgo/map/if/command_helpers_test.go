package if_

import (
	"github.com/glycerine/abcgo/opt/dau"
	"testing"
)

func TestDsdTuneResetRetryAndStructuralDecomposition(t *testing.T) {
	for _, fast := range []int32{0, 1} {
		p := If_DsdManAlloc(4, 2)
		literals := make([]int32, 0, 3)
		for _, text := range []string{"(abc)", "<abc>", "[abcd]"} {
			n := int32(3)
			if text == "[abcd]" {
				n = 4
			}
			literals = append(literals, If_DsdManCompute(p, dau.Dau_DsdToTruth(text, n), n, make([]byte, n), ""))
		}
		If_DsdManTune(p, 2, fast, 0, 0, 0)
		if If_DsdManReadMark(p, literals[0]) != 0 {
			t.Errorf("fast=%d: three-input AND fits two 2-LUTs", fast)
		}
		if If_DsdManReadMark(p, literals[2]) != 1 {
			t.Errorf("fast=%d: four-variable parity cannot fit two 2-LUTs", fast)
		}
		If_DsdManTune(p, 3, fast, 1, 0, 0)
		if If_DsdManReadMark(p, literals[2]) != 0 {
			t.Errorf("fast=%d: four-variable parity fits two 3-LUTs", fast)
		}
		for _, lit := range literals {
			dsdObject(p, lit).fMark = 1
		}
		If_DsdManTune(p, 0, fast, 1, 0, 0)
		for id := int32(0); id < If_DsdManObjNum(p); id++ {
			if If_DsdVecObj(&p.vObjs, id).fMark != 0 {
				t.Errorf("reset left id %d marked", id)
			}
		}
		If_DsdManFree(p, 0)
	}
}

func TestLutWireDelayAdjustsAndRestoresAllPins(t *testing.T) {
	lib := &If_LibLut_t{LutMax: 4}
	for size := int32(0); size <= lib.LutMax; size++ {
		for pin := int32(0); pin <= size; pin++ {
			lib.pLutDelays[size][pin] = float32(10*size + pin)
		}
	}
	before := *lib
	If_LibLutAddWireDelay(lib, 0.5)
	for size := int32(0); size <= lib.LutMax; size++ {
		for pin := int32(0); pin <= size; pin++ {
			if got := lib.pLutDelays[size][pin]; got != before.pLutDelays[size][pin]+0.5 {
				t.Errorf("size=%d pin=%d got=%g", size, pin, got)
			}
		}
	}
	If_LibLutAddWireDelay(lib, -0.5)
	if lib.pLutDelays != before.pLutDelays {
		t.Fatal("wire delay failed to restore the library")
	}
}
