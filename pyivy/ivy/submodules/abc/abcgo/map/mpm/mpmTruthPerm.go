package mpm

func Extra_Truth6SwapAdjacent(t word, iVar int32) word {
	masks := [5][3]word{
		{0x9999999999999999, 0x2222222222222222, 0x4444444444444444},
		{0xC3C3C3C3C3C3C3C3, 0x0C0C0C0C0C0C0C0C, 0x3030303030303030},
		{0xF00FF00FF00FF00F, 0x00F000F000F000F0, 0x0F000F000F000F00},
		{0xFF0000FFFF0000FF, 0x0000FF000000FF00, 0x00FF000000FF0000},
		{0xFFFF00000000FFFF, 0x00000000FFFF0000, 0x0000FFFF00000000},
	}
	assert(iVar >= 0 && iVar < 5, "valid adjacent variable")
	shift := uint(1 << iVar)
	return t&masks[iVar][0] | (t&masks[iVar][1])<<shift | (t&masks[iVar][2])>>shift
}
func Extra_Truth6ChangePhase(t word, iVar int32) word {
	masks := [6]word{0xAAAAAAAAAAAAAAAA, 0xCCCCCCCCCCCCCCCC, 0xF0F0F0F0F0F0F0F0, 0xFF00FF00FF00FF00, 0xFFFF0000FFFF0000, 0xFFFFFFFF00000000}
	assert(iVar >= 0 && iVar < 6, "valid phase variable")
	shift := uint(1 << iVar)
	return (t&^masks[iVar])<<shift | (t&masks[iVar])>>shift
}
