// Disconnected covering-matrix blocks, ported from Espresso part.c.
package espresso

// sm_block_partition splits the bipartite row/column graph at the component
// containing its first row. Neither matrix entries nor user payloads are moved;
// returned blocks contain independent elements, as in the source copy routine.
func sm_block_partition(matrix *sm_matrix, left, right **sm_matrix) bool {
	if matrix.nrows == 0 {
		return false
	}
	for row := matrix.first_row; row != nil; row = row.next_row {
		row.flag = 0
	}
	for column := matrix.first_col; column != nil; column = column.next_col {
		column.flag = 0
	}
	pending := []*sm_row{matrix.first_row}
	matrix.first_row.flag = 1
	for i := 0; i < len(pending); i++ {
		for element := pending[i].first_col; element != nil; element = element.next_col {
			column := sm_get_col(matrix, element.col_num)
			if column.flag != 0 {
				continue
			}
			column.flag = 1
			for entry := column.first_row; entry != nil; entry = entry.next_row {
				row := sm_get_row(matrix, entry.row_num)
				if row.flag == 0 {
					row.flag = 1
					pending = append(pending, row)
				}
			}
		}
	}
	if len(pending) == matrix.nrows {
		return false
	}
	if left == nil || right == nil {
		panic("espresso: missing sparse partition output")
	}
	*left, *right = sm_alloc(), sm_alloc()
	for row := matrix.first_row; row != nil; row = row.next_row {
		destination := *right
		if row.flag != 0 {
			destination = *left
		}
		sm_copy_row(destination, row.row_num, row)
	}
	return true
}
