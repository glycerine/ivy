package sfm

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/util"
)

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

// Sfm_LibTruth6Two evaluates a bottom cell connected to one top-cell pin.
func Sfm_LibTruth6Two(bottom, top *mio.Mio_Cell2_t, pin int32) uint64 {
	if pin < 0 || pin >= top.Fanins() || bottom.Fanins()+top.Fanins()-1 > 6 {
		panic("sfm: invalid composed cell")
	}
	bot := mio.Exp_Truth6(bottom.Fanins(), bottom.Expression(), nil)
	fanins := make([]uint64, top.Fanins())
	next := bottom.Fanins()
	for i := range fanins {
		if int32(i) == pin {
			fanins[i] = bot
		} else {
			for bit := uint(0); bit < 64; bit++ {
				if bit&(1<<uint(next)) != 0 {
					fanins[i] |= 1 << bit
				}
			}
			next++
		}
	}
	return mio.Exp_Truth6(top.Fanins(), top.Expression(), fanins)
}
func Sfm_LibTruth8Two(bottom, top *mio.Mio_Cell2_t, pin int32, result []uint64) {
	if pin < 0 || pin >= top.Fanins() || bottom.Fanins()+top.Fanins()-1 > 8 || len(result) < 4 {
		panic("sfm: invalid composed cell")
	}
	bot := make([]uint64, 4)
	mio.Exp_Truth8(bottom.Fanins(), bottom.Expression(), nil, bot)
	fanins := make([][]uint64, top.Fanins())
	next := bottom.Fanins()
	for i := range fanins {
		if int32(i) == pin {
			fanins[i] = bot
		} else {
			fanins[i] = make([]uint64, 4)
			for a := uint(0); a < 256; a++ {
				if a&(1<<uint(next)) != 0 {
					fanins[i][a>>6] |= 1 << (a & 63)
				}
			}
			next++
		}
	}
	mio.Exp_Truth8(top.Fanins(), top.Expression(), fanins, result)
}
func Sfm_LibCellProfile(bottom, top *mio.Mio_Cell2_t, permutation []int32) []int32 {
	result := make([]int32, len(permutation))
	for i, pin := range permutation {
		result[i] = 1
		if top != nil && pin < bottom.Fanins() {
			result[i]++
		}
	}
	return result
}
func Sfm_LibNewIsContained(oldArea int32, oldProfile []int32, newArea int32, newProfile []int32) bool {
	if newArea < oldArea {
		return false
	}
	for i, v := range oldProfile {
		if newProfile[i] < v {
			return false
		}
	}
	return true
}
func Sfm_LibNewContains(oldArea int32, oldProfile []int32, newArea int32, newProfile []int32) bool {
	if newArea > oldArea {
		return false
	}
	for i, v := range oldProfile {
		if newProfile[i] > v {
			return false
		}
	}
	return true
}
