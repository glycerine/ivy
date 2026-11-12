// Sparse binary matrices, ported from Espresso matrix.c.
package espresso

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func sm_alloc() *sm_matrix { return &sm_matrix{} }
func sm_alloc_size(row, col int) *sm_matrix {
	matrix := sm_alloc()
	sm_resize(matrix, row, col)
	return matrix
}
func sm_free(matrix *sm_matrix) {
	if matrix == nil {
		return
	}
	for row := matrix.first_row; row != nil; {
		next := row.next_row
		for element := row.first_col; element != nil; {
			following := element.next_col
			*element = sm_element{}
			element = following
		}
		*row = sm_row{}
		row = next
	}
	for column := matrix.first_col; column != nil; {
		next := column.next_col
		*column = sm_col{}
		column = next
	}
	*matrix = sm_matrix{}
}
func sm_dup(matrix *sm_matrix) *sm_matrix {
	result := sm_alloc()
	if matrix.last_row != nil {
		sm_resize(result, matrix.last_row.row_num, matrix.last_col.col_num)
	}
	for row := matrix.first_row; row != nil; row = row.next_row {
		sm_copy_row(result, row.row_num, row)
	}
	return result
}
func sm_resize(matrix *sm_matrix, row, col int) {
	grow := func(size, index int) int {
		if index < size {
			return size
		}
		maxInt := int(^uint(0) >> 1)
		if index == maxInt {
			panic("espresso: sparse index overflows allocation size")
		}
		wanted := index + 1
		if size <= maxInt/2 && 2*size > wanted {
			wanted = 2 * size
		}
		return wanted
	}
	if size := grow(len(matrix.rows), row); size > len(matrix.rows) {
		matrix.rows = append(matrix.rows, make([]*sm_row, size-len(matrix.rows))...)
	}
	if size := grow(len(matrix.cols), col); size > len(matrix.cols) {
		matrix.cols = append(matrix.cols, make([]*sm_col, size-len(matrix.cols))...)
	}
}
func sm_add_row_header(matrix *sm_matrix, row *sm_row) {
	next := matrix.first_row
	if matrix.last_row != nil && matrix.last_row.row_num < row.row_num {
		next = nil
	} else {
		for next != nil && next.row_num < row.row_num {
			next = next.next_row
		}
	}
	previous := matrix.last_row
	if next != nil {
		previous = next.prev_row
	}
	row.prev_row, row.next_row = previous, next
	if previous == nil {
		matrix.first_row = row
	} else {
		previous.next_row = row
	}
	if next == nil {
		matrix.last_row = row
	} else {
		next.prev_row = row
	}
	matrix.nrows++
}
func sm_add_col_header(matrix *sm_matrix, column *sm_col) {
	next := matrix.first_col
	if matrix.last_col != nil && matrix.last_col.col_num < column.col_num {
		next = nil
	} else {
		for next != nil && next.col_num < column.col_num {
			next = next.next_col
		}
	}
	previous := matrix.last_col
	if next != nil {
		previous = next.prev_col
	}
	column.prev_col, column.next_col = previous, next
	if previous == nil {
		matrix.first_col = column
	} else {
		previous.next_col = column
	}
	if next == nil {
		matrix.last_col = column
	} else {
		next.prev_col = column
	}
	matrix.ncols++
}
func sm_drop_row_header(matrix *sm_matrix, row *sm_row) {
	if row.prev_row == nil {
		matrix.first_row = row.next_row
	} else {
		row.prev_row.next_row = row.next_row
	}
	if row.next_row == nil {
		matrix.last_row = row.prev_row
	} else {
		row.next_row.prev_row = row.prev_row
	}
	matrix.rows[row.row_num] = nil
	matrix.nrows--
	*row = sm_row{}
}
func sm_drop_col_header(matrix *sm_matrix, column *sm_col) {
	if column.prev_col == nil {
		matrix.first_col = column.next_col
	} else {
		column.prev_col.next_col = column.next_col
	}
	if column.next_col == nil {
		matrix.last_col = column.prev_col
	} else {
		column.next_col.prev_col = column.prev_col
	}
	matrix.cols[column.col_num] = nil
	matrix.ncols--
	*column = sm_col{}
}
func sm_insert(matrix *sm_matrix, rowIndex, colIndex int) *sm_element {
	if rowIndex < 0 || colIndex < 0 {
		panic("espresso: negative matrix index")
	}
	if existing := sm_find(matrix, rowIndex, colIndex); existing != nil {
		return existing
	}
	sm_resize(matrix, rowIndex, colIndex)
	row := matrix.rows[rowIndex]
	if row == nil {
		row = sm_row_alloc()
		row.row_num = rowIndex
		matrix.rows[rowIndex] = row
		sm_add_row_header(matrix, row)
	}
	column := matrix.cols[colIndex]
	if column == nil {
		column = sm_col_alloc()
		column.col_num = colIndex
		matrix.cols[colIndex] = column
		sm_add_col_header(matrix, column)
	}
	element := &sm_element{row_num: rowIndex, col_num: colIndex, matrix: matrix}
	sm_link_row_element(row, element)
	sm_link_col_element(column, element)
	return element
}
func sm_find(matrix *sm_matrix, rowIndex, colIndex int) *sm_element {
	row, column := sm_get_row(matrix, rowIndex), sm_get_col(matrix, colIndex)
	if row == nil || column == nil {
		return nil
	}
	if row.length < column.length {
		return sm_row_find(row, colIndex)
	}
	return sm_col_find(column, rowIndex)
}
func sm_remove(matrix *sm_matrix, row, col int) { sm_remove_element(matrix, sm_find(matrix, row, col)) }
func sm_remove_element(matrix *sm_matrix, element *sm_element) {
	// Matrix identity makes detached/foreign removal safe without an extra lookup.
	if element == nil || element.matrix != matrix {
		return
	}
	row, column := matrix.rows[element.row_num], matrix.cols[element.col_num]
	sm_unlink_row_element(row, element)
	sm_unlink_col_element(column, element)
	if row.length == 0 {
		sm_drop_row_header(matrix, row)
	}
	if column.length == 0 {
		sm_drop_col_header(matrix, column)
	}
	*element = sm_element{}
}
func sm_delrow(matrix *sm_matrix, index int) {
	row := sm_get_row(matrix, index)
	if row == nil {
		return
	}
	for row.length > 0 {
		sm_remove_element(matrix, row.first_col)
		if matrix.rows[index] == nil {
			return
		}
	}
}
func sm_delcol(matrix *sm_matrix, index int) {
	column := sm_get_col(matrix, index)
	if column == nil {
		return
	}
	for column.length > 0 {
		sm_remove_element(matrix, column.first_row)
		if matrix.cols[index] == nil {
			return
		}
	}
}
func sm_copy_row(destination *sm_matrix, index int, row *sm_row) {
	for element := row.first_col; element != nil; element = element.next_col {
		sm_insert(destination, index, element.col_num)
	}
}
func sm_copy_col(destination *sm_matrix, index int, column *sm_col) {
	for element := column.first_row; element != nil; element = element.next_row {
		sm_insert(destination, element.row_num, index)
	}
}
func sm_longest_row(matrix *sm_matrix) *sm_row {
	var best *sm_row
	for row := matrix.first_row; row != nil; row = row.next_row {
		if best == nil || row.length > best.length {
			best = row
		}
	}
	return best
}
func sm_longest_col(matrix *sm_matrix) *sm_col {
	var best *sm_col
	for column := matrix.first_col; column != nil; column = column.next_col {
		if best == nil || column.length > best.length {
			best = column
		}
	}
	return best
}
func sm_num_elements(matrix *sm_matrix) int {
	count := 0
	for row := matrix.first_row; row != nil; row = row.next_row {
		count += row.length
	}
	return count
}

// Read retains the source output-pointer/boolean result contract. Successful
// prefixes remain available on malformed input, matching fscanf-based reading.
func sm_scan_decimal(reader io.Reader) (int, error) {
	var token string
	if _, err := fmt.Fscan(reader, &token); err != nil {
		return 0, err
	}
	return strconv.Atoi(token)
}
func sm_read(reader io.Reader, result **sm_matrix) bool {
	matrix := sm_alloc()
	*result = matrix
	input := bufio.NewReader(reader)
	for {
		row, err := sm_scan_decimal(input)
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}
		col, err := sm_scan_decimal(input)
		maxInt := int(^uint(0) >> 1)
		if err != nil || row < 0 || col < 0 || row == maxInt || col == maxInt {
			return false
		}
		sm_insert(matrix, row, col)
	}
}
func sm_read_compressed(reader io.Reader, result **sm_matrix) bool {
	matrix := sm_alloc()
	*result = matrix
	input := bufio.NewReader(reader)
	rows, rowErr := sm_scan_decimal(input)
	cols, colErr := sm_scan_decimal(input)
	if rowErr != nil || colErr != nil || rows < 0 || cols < 0 {
		return false
	}
	// Capacity hints do not create row/column entries. Grow only for actual
	// nonzeros, so a truncated large header cannot force an allocation first.
	readHex := func() (uint64, bool) {
		var token string
		if _, err := fmt.Fscan(input, &token); err != nil {
			return 0, false
		}
		token = strings.TrimPrefix(token, "+")
		if strings.HasPrefix(token, "0x") || strings.HasPrefix(token, "0X") {
			token = token[2:]
		}
		value, err := strconv.ParseUint(token, 16, 64)
		return value, err == nil
	}
	for row := 0; row < rows; row++ {
		if _, ok := readHex(); !ok {
			return false
		} // source row label/count word
		for col := 0; col < cols; {
			word, ok := readHex()
			if !ok || word>>32 != 0 {
				return false
			}
			remaining := min(32, cols-col)
			if remaining < 32 && word>>uint(remaining) != 0 {
				return false
			}
			for bit := 0; word != 0; word, bit = word>>1, bit+1 {
				if word&1 != 0 {
					sm_insert(matrix, row, col+bit)
				}
			}
			col += remaining
		}
	}
	return true
}
func sm_write(writer io.Writer, matrix *sm_matrix) {
	for row := matrix.first_row; row != nil; row = row.next_row {
		for element := row.first_col; element != nil; element = element.next_col {
			fmt.Fprintf(writer, "%d %d\n", element.row_num, element.col_num)
		}
	}
}
func sm_print(writer io.Writer, matrix *sm_matrix) {
	if matrix.last_col == nil {
		return
	}
	for _, scale := range []int{100, 10, 1} {
		if scale > 1 && matrix.last_col.col_num < scale {
			continue
		}
		fmt.Fprint(writer, "    ")
		for column := matrix.first_col; column != nil; column = column.next_col {
			fmt.Fprint(writer, (column.col_num/scale)%10)
		}
		fmt.Fprintln(writer)
	}
	fmt.Fprint(writer, "    ")
	for column := matrix.first_col; column != nil; column = column.next_col {
		fmt.Fprint(writer, "-")
	}
	fmt.Fprintln(writer)
	for row := matrix.first_row; row != nil; row = row.next_row {
		fmt.Fprintf(writer, "%3d:", row.row_num)
		for column := matrix.first_col; column != nil; column = column.next_col {
			value := '.'
			if sm_row_find(row, column.col_num) != nil {
				value = '1'
			}
			fmt.Fprintf(writer, "%c", value)
		}
		fmt.Fprintln(writer)
	}
}
func sm_dump(matrix *sm_matrix, label string, maximum int) {
	fmt.Fprintf(os.Stdout, "%s %d rows by %d cols\n", label, matrix.nrows, matrix.ncols)
	if matrix.nrows < maximum {
		sm_print(os.Stdout, matrix)
	}
}
