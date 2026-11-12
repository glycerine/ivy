package espresso

// Eliminate a two-column primary row when one column occurs in exactly two
// rows. Merge the secondary row into the other column's rows, solve one unit
// cheaper, then restore whichever primary column covers the secondary row.
func gimpel_reduce(matrix *sm_matrix, selected *solution_t, weight []int32, lower, bound int32, depth int, stats *stats_t, best **solution_t) bool {
	var c1, c2 *sm_col
	primary, secondary := -1, -1
	for row := matrix.first_row; row != nil; row = row.next_row {
		if row.length != 2 {
			continue
		}
		a, b := sm_get_col(matrix, row.first_col.col_num), sm_get_col(matrix, row.last_col.col_num)
		if a.length != 2 {
			if b.length != 2 {
				continue
			}
			a, b = b, a
		}
		c1, c2 = a, b
		primary = row.row_num
		secondary = c1.first_row.row_num
		if secondary == primary {
			secondary = c1.last_row.row_num
		}
		break
	}
	if primary < 0 {
		return false
	}
	first, second := c1.col_num, c2.col_num
	saved := sm_row_dup(sm_get_row(matrix, secondary))
	sm_row_remove(saved, first)
	for element := c2.first_row; element != nil; element = element.next_row {
		if element.row_num == primary {
			continue
		}
		for other := saved.first_col; other != nil; other = other.next_col {
			sm_insert(matrix, element.row_num, other.col_num)
		}
	}
	sm_delcol(matrix, first)
	sm_delcol(matrix, second)
	sm_delrow(matrix, primary)
	sm_delrow(matrix, secondary)
	stats.gimpel_count++
	stats.gimpel++
	*best = sm_mincov(matrix, selected, weight, lower-1, bound-1, depth, stats)
	stats.gimpel--
	if *best != nil {
		if sm_row_intersects(saved, (*best).row) {
			solution_add(*best, weight, second)
		} else {
			solution_add(*best, weight, first)
		}
	}
	sm_row_free(saved)
	return true
}
