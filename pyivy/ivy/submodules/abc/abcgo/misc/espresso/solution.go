package espresso

func solution_alloc() *solution_t { return &solution_t{row: sm_row_alloc()} }
func solution_free(solution *solution_t) {
	if solution != nil {
		sm_row_free(solution.row)
		solution.row = nil
	}
}
func solution_dup(solution *solution_t) *solution_t {
	return &solution_t{row: sm_row_dup(solution.row), cost: solution.cost}
}
func solution_add(solution *solution_t, weight []int32, column int) {
	if sm_row_find(solution.row, column) != nil {
		return
	}
	sm_row_insert(solution.row, column)
	solution.cost += mincovWeight(weight, column)
}
func solution_accept(solution *solution_t, matrix *sm_matrix, weight []int32, column int) {
	solution_add(solution, weight, column)
	col := sm_get_col(matrix, column)
	if col == nil {
		return
	}
	for element := col.first_row; element != nil; {
		next := element.next_row
		sm_delrow(matrix, element.row_num)
		element = next
	}
}
func solution_reject(solution *solution_t, matrix *sm_matrix, weight []int32, column int) {
	sm_delcol(matrix, column)
}
func solution_choose_best(a, b *solution_t) *solution_t {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.cost <= b.cost {
		solution_free(b)
		return a
	}
	solution_free(a)
	return b
}
