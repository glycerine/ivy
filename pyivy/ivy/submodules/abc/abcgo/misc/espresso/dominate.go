// Sparse covering-matrix dominance, ported from Espresso dominate.c.
package espresso

func sm_row_dominance(matrix *sm_matrix) int {
	before := matrix.nrows
	for row := matrix.first_row; row != nil; row = row.next_row {
		least := sm_get_col(matrix, row.first_col.col_num)
		for element := row.first_col.next_col; element != nil; element = element.next_col {
			column := sm_get_col(matrix, element.col_num)
			if column.length < least.length {
				least = column
			}
		}
		for element := least.first_row; element != nil; {
			next := element.next_row
			candidate := sm_get_row(matrix, element.row_num)
			if candidate.length > row.length || (candidate.length == row.length && candidate.row_num > row.row_num) {
				if sm_row_contains(row, candidate) {
					sm_delrow(matrix, candidate.row_num)
				}
			}
			element = next
		}
	}
	return before - matrix.nrows
}
func sm_col_dominance(matrix *sm_matrix, weights []int32) int {
	if weights != nil {
		for column := matrix.first_col; column != nil; column = column.next_col {
			if column.col_num >= len(weights) {
				panic("espresso: sparse column weight missing")
			}
		}
	}
	before := matrix.ncols
	for column := matrix.first_col; column != nil; {
		next := column.next_col
		least := sm_get_row(matrix, column.first_row.row_num)
		for element := column.first_row.next_row; element != nil; element = element.next_row {
			row := sm_get_row(matrix, element.row_num)
			if row.length < least.length {
				least = row
			}
		}
		for element := least.first_col; element != nil; element = element.next_col {
			candidate := sm_get_col(matrix, element.col_num)
			if weights != nil && weights[candidate.col_num] > weights[column.col_num] {
				continue
			}
			if candidate.length > column.length || (candidate.length == column.length && candidate.col_num > column.col_num) {
				if sm_col_contains(column, candidate) {
					sm_delcol(matrix, column.col_num)
					break
				}
			}
		}
		column = next
	}
	return before - matrix.ncols
}
