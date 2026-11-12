package espresso

// SmMinimumCover selects a minimum-cost column set. Nil weights means unit
// costs; heuristic!=0 accepts the first branch instead of proving optimality.
// Matrix structure and caller weight storage are preserved.
func SmMinimumCover(matrix *SparseMatrix, weight []int32, heuristic, debugLevel int) *SparseRow {
	return sm_minimum_cover(matrix, weight, heuristic, debugLevel)
}
