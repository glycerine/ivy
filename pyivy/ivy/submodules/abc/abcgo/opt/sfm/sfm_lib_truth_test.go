package sfm

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestFindComplementInputGate(t *testing.T) {
	original := uint64(0x8888888888888888)
	flipped := util.Abc_Tt6Flip(original, 0)
	swapped := util.Abc_Tt6SwapAdjacent(flipped, 0)
	var pin int
	if got := Sfm_LibFindComplInputGate([]uint64{original, flipped}, 0, 2, 0, &pin); got != 1 || pin != 0 {
		t.Fatalf("direct %d pin %d", got, pin)
	}
	if got := Sfm_LibFindComplInputGate([]uint64{original, swapped}, 0, 2, 0, &pin); got != 1 || pin != 1 {
		t.Fatalf("swapped %d pin %d", got, pin)
	}
	if got := Sfm_LibFindComplInputGate([]uint64{original}, 0, 2, 0, &pin); got != -1 || pin != -1 {
		t.Fatalf("missing %d pin %d", got, pin)
	}
}
