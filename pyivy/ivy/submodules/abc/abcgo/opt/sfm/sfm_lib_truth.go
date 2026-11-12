package sfm

import "github.com/glycerine/abcgo/misc/util"

// Sfm_LibFindComplInputGate finds a gate after complementing one input,
// allowing one adjacent input swap. pinNew receives the resulting pin index.
func Sfm_LibFindComplInputGate(functions []uint64, gate, fanins, pin int, pinNew *int) int {
	if pin < 0 || pin >= fanins {
		panic("sfm: invalid input pin")
	}
	flipped := util.Abc_Tt6Flip(functions[gate], int32(pin))
	find := func(truth uint64) int {
		for i, v := range functions {
			if v == truth {
				return i
			}
		}
		return -1
	}
	if pinNew != nil {
		*pinNew = pin
	}
	if i := find(flipped); i >= 0 {
		return i
	}
	if pin > 0 {
		if pinNew != nil {
			*pinNew = pin - 1
		}
		if i := find(util.Abc_Tt6SwapAdjacent(flipped, int32(pin-1))); i >= 0 {
			return i
		}
	}
	if pin+1 < fanins {
		if pinNew != nil {
			*pinNew = pin + 1
		}
		if i := find(util.Abc_Tt6SwapAdjacent(flipped, int32(pin))); i >= 0 {
			return i
		}
	}
	if pinNew != nil {
		*pinNew = -1
	}
	return -1
}
