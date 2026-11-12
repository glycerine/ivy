// Sparse matrix types, ported from Espresso sparse.h. Each matrix element is
// shared by its sorted row and column lists; header arrays provide direct lookup.
package espresso

type sm_element struct {
	row_num, col_num   int
	next_row, prev_row *sm_element // neighbors in its column
	next_col, prev_col *sm_element // neighbors in its row
	user_word          any
	matrix             *sm_matrix
}
type sm_row struct {
	row_num, length, flag int
	first_col, last_col   *sm_element
	next_row, prev_row    *sm_row
	user_word             any
}
type sm_col struct {
	col_num, length, flag int
	first_row, last_row   *sm_element
	next_col, prev_col    *sm_col
	user_word             any
}
type sm_matrix struct {
	rows                []*sm_row
	cols                []*sm_col
	first_row, last_row *sm_row
	first_col, last_col *sm_col
	nrows, ncols        int
	user_word           any
}

// Public aliases expose the matrix family without duplicating its storage.
type SparseMatrix = sm_matrix
type SparseRow = sm_row
type SparseColumn = sm_col
type SparseElement = sm_element

func sm_get_row(matrix *sm_matrix, index int) *sm_row {
	if index < 0 || index >= len(matrix.rows) {
		return nil
	}
	return matrix.rows[index]
}
func sm_get_col(matrix *sm_matrix, index int) *sm_col {
	if index < 0 || index >= len(matrix.cols) {
		return nil
	}
	return matrix.cols[index]
}

// NewSparseMatrix constructs an empty matrix for SmMinimumCover and other
// native Go callers. Row and column numbers are nonnegative integer IDs.
func NewSparseMatrix() *SparseMatrix { return sm_alloc() }
func (matrix *sm_matrix) Insert(row, column int) *SparseElement {
	return sm_insert(matrix, row, column)
}
func (matrix *sm_matrix) Find(row, column int) *SparseElement { return sm_find(matrix, row, column) }
func (matrix *sm_matrix) Remove(row, column int)              { sm_remove(matrix, row, column) }
func (matrix *sm_matrix) RemoveRow(row int)                   { sm_delrow(matrix, row) }
func (matrix *sm_matrix) RemoveColumn(column int)             { sm_delcol(matrix, column) }
func (matrix *sm_matrix) Row(index int) *SparseRow            { return sm_get_row(matrix, index) }
func (matrix *sm_matrix) Column(index int) *SparseColumn      { return sm_get_col(matrix, index) }
func (matrix *sm_matrix) RowCount() int                       { return matrix.nrows }
func (matrix *sm_matrix) ColumnCount() int                    { return matrix.ncols }
func (matrix *sm_matrix) ElementCount() int                   { return sm_num_elements(matrix) }

// Columns returns a copy, so callers cannot mutate a matrix-owned vector.
func (row *sm_row) Columns() []int {
	var result []int
	if row != nil {
		for element := row.first_col; element != nil; element = element.next_col {
			result = append(result, element.col_num)
		}
	}
	return result
}
func (column *sm_col) Rows() []int {
	var result []int
	if column != nil {
		for element := column.first_row; element != nil; element = element.next_row {
			result = append(result, element.row_num)
		}
	}
	return result
}
func (element *sm_element) Row() int    { return element.row_num }
func (element *sm_element) Column() int { return element.col_num }
