package espresso

// Greedily select pairwise-disjoint rows in the row-intersection graph. The
// sum of each selected row's cheapest column is an admissible covering bound.
func sm_maximal_independent_set(matrix *sm_matrix, weight []int32) *solution_t {
	independent := solution_alloc()
	intersections := build_intersection_matrix(matrix)
	for intersections.nrows > 0 {
		best := intersections.first_row
		for row := best.next_row; row != nil; row = row.next_row {
			if row.length < best.length {
				best = row
			}
		}
		least := int32(1)
		if weight != nil {
			row := sm_get_row(matrix, best.row_num)
			least = weight[row.first_col.col_num]
			for element := row.first_col.next_col; element != nil; element = element.next_col {
				least = min(least, weight[element.col_num])
			}
		}
		independent.cost += least
		sm_row_insert(independent.row, best.row_num)
		saved := sm_row_dup(best)
		for element := saved.first_col; element != nil; element = element.next_col {
			sm_delrow(intersections, element.col_num)
			sm_delcol(intersections, element.col_num)
		}
		sm_row_free(saved)
	}
	sm_free(intersections)
	return independent
}
func build_intersection_matrix(matrix *sm_matrix) *sm_matrix {
	result := sm_alloc()
	for row := matrix.first_row; row != nil; row = row.next_row {
		seen := map[int]bool{}
		for element := row.first_col; element != nil; element = element.next_col {
			column := sm_get_col(matrix, element.col_num)
			for other := column.first_row; other != nil; other = other.next_row {
				if !seen[other.row_num] {
					seen[other.row_num] = true
					sm_insert(result, row.row_num, other.row_num)
				}
			}
		}
	}
	return result
}
