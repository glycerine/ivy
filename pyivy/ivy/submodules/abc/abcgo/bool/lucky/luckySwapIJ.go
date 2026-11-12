package lucky

// Arbitrary swaps use adjacent transpositions and restore intermediate slots.
// The operation remains word-parallel for low/word-boundary/high variables.
func swap_ij(table []uint64, vars, i, j int32) {
	if i < 0 || j < 0 || i >= vars || j >= vars {
		panic("swap variable outside truth support")
	}
	if i > j {
		i, j = j, i
	}
	for k := i; k < j; k++ {
		Kit_TruthSwapAdjacentVars_64bit(table, vars, k)
	}
	for k := j - 2; k >= i; k-- {
		Kit_TruthSwapAdjacentVars_64bit(table, vars, k)
	}
}
func swap_ij_case1(table []uint64, vars, i, j int32) { swap_ij(table, vars, i, j) }
func swap_ij_case2(table []uint64, vars, i, j int32) { swap_ij(table, vars, i, j) }
func swap_ij_case3(table []uint64, vars, i, j int32) { swap_ij(table, vars, i, j) }
func SwapVariables(table []uint64, vars, i, j int32) { swap_ij(table, vars, i, j) }
