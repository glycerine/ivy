package espresso

import (
	"bytes"
	"math/big"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

type sparseModel map[[2]int]bool

func sparseSnapshot(matrix *sm_matrix) sparseModel {
	result := make(sparseModel)
	for row := matrix.first_row; row != nil; row = row.next_row {
		for element := row.first_col; element != nil; element = element.next_col {
			result[[2]int{element.row_num, element.col_num}] = true
		}
	}
	return result
}
func sparseCheck(t *testing.T, matrix *sm_matrix, model sparseModel) {
	t.Helper()
	seen := make(sparseModel)
	rows, cols := 0, 0
	var previousRow *sm_row
	for row := matrix.first_row; row != nil; row = row.next_row {
		if row.prev_row != previousRow || previousRow != nil && previousRow.row_num >= row.row_num || sm_get_row(matrix, row.row_num) != row || row.length <= 0 {
			t.Fatal("row header invariant")
		}
		count := 0
		var previous *sm_element
		for element := row.first_col; element != nil; element = element.next_col {
			if element.prev_col != previous || previous != nil && previous.col_num >= element.col_num || element.row_num != row.row_num || element.matrix != matrix || sm_find(matrix, element.row_num, element.col_num) != element {
				t.Fatal("row element invariant")
			}
			key := [2]int{element.row_num, element.col_num}
			if !model[key] || seen[key] {
				t.Fatal("unexpected/duplicate element", key)
			}
			seen[key] = true
			previous = element
			count++
		}
		if count != row.length || previous != row.last_col {
			t.Fatal("row length/tail invariant")
		}
		rows++
		previousRow = row
	}
	if rows != matrix.nrows || previousRow != matrix.last_row || len(seen) != len(model) {
		t.Fatal("row count/model invariant", rows, matrix.nrows, len(seen), len(model))
	}
	var previousCol *sm_col
	columnElements := make(map[*sm_element]bool)
	for column := matrix.first_col; column != nil; column = column.next_col {
		if column.prev_col != previousCol || previousCol != nil && previousCol.col_num >= column.col_num || sm_get_col(matrix, column.col_num) != column || column.length <= 0 {
			t.Fatal("column header invariant")
		}
		count := 0
		var previous *sm_element
		for element := column.first_row; element != nil; element = element.next_row {
			if element.prev_row != previous || previous != nil && previous.row_num >= element.row_num || element.col_num != column.col_num || sm_find(matrix, element.row_num, element.col_num) != element {
				t.Fatal("column element invariant")
			}
			columnElements[element] = true
			previous = element
			count++
		}
		if count != column.length || previous != column.last_row {
			t.Fatal("column length/tail invariant")
		}
		cols++
		previousCol = column
	}
	if cols != matrix.ncols || previousCol != matrix.last_col || len(columnElements) != len(model) {
		t.Fatal("column count invariant")
	}
	for i, row := range matrix.rows {
		if row != nil && (row.row_num != i || row.length <= 0) {
			t.Fatal("row index invariant")
		}
	}
	for i, col := range matrix.cols {
		if col != nil && (col.col_num != i || col.length <= 0) {
			t.Fatal("column index invariant")
		}
	}
	if sm_num_elements(matrix) != len(model) {
		t.Fatal("element count")
	}
}
func TestSparseMatrixRandomOperations(t *testing.T) {
	random := rand.New(rand.NewPCG(51, 99))
	matrix := sm_alloc()
	model := make(sparseModel)
	for operation := 0; operation < 6000; operation++ {
		row, col := random.IntN(20), random.IntN(25)
		key := [2]int{row, col}
		switch random.IntN(6) {
		case 0, 1, 2:
			old := sm_find(matrix, row, col)
			element := sm_insert(matrix, row, col)
			if old != nil && old != element {
				t.Fatal("duplicate replaced element")
			}
			model[key] = true
		case 3:
			element := sm_find(matrix, row, col)
			sm_remove_element(matrix, element)
			sm_remove(matrix, row, col)
			delete(model, key)
		case 4:
			sm_delrow(matrix, row)
			for key := range model {
				if key[0] == row {
					delete(model, key)
				}
			}
		case 5:
			sm_delcol(matrix, col)
			for key := range model {
				if key[1] == col {
					delete(model, key)
				}
			}
		}
		sparseCheck(t, matrix, model)
	}
	duplicate := sm_dup(matrix)
	sparseCheck(t, duplicate, model)
	sm_free(matrix)
	sparseCheck(t, matrix, make(sparseModel))
	sparseCheck(t, duplicate, model)
	sm_free(duplicate)
}
func TestSparseCopyColumnAndPayloadOwnership(t *testing.T) {
	source := sm_alloc()
	first := sm_insert(source, 2, 4)
	sm_insert(source, 5, 4)
	first.user_word = map[string]int{"payload": 7}
	destination := sm_alloc()
	sm_copy_col(destination, 9, sm_get_col(source, 4))
	sparseCheck(t, destination, sparseModel{{2, 9}: true, {5, 9}: true})
	sm_copy_row(destination, 8, sm_get_row(source, 2))
	if sm_find(destination, 8, 4) == nil {
		t.Fatal("row copy")
	}
	duplicate := sm_dup(source)
	if sm_find(duplicate, 2, 4).user_word != nil {
		t.Fatal("duplicate copied user payload")
	}
	// Foreign and detached elements cannot corrupt another matrix.
	sm_remove_element(destination, first)
	if sm_find(source, 2, 4) != first {
		t.Fatal("foreign removal changed source")
	}
	sm_remove_element(source, first)
	sm_remove_element(source, first)
	if first.matrix != nil || first.user_word != nil {
		t.Fatal("element not detached")
	}
	sparseCheck(t, source, sparseModel{{5, 4}: true})
}
func rowValues(row *sm_row) []int {
	var result []int
	for element := row.first_col; element != nil; element = element.next_col {
		result = append(result, element.col_num)
	}
	return result
}
func colValues(column *sm_col) []int {
	var result []int
	for element := column.first_row; element != nil; element = element.next_row {
		result = append(result, element.row_num)
	}
	return result
}
func TestSparseRowAndColumnSetOperations(t *testing.T) {
	first, second := sm_row_alloc(), sm_row_alloc()
	a, b := sm_col_alloc(), sm_col_alloc()
	for _, value := range []int{9, 1, 4, 1, 7} {
		sm_row_insert(first, value)
		sm_col_insert(a, value)
	}
	for _, value := range []int{1, 4, 7, 12} {
		sm_row_insert(second, value)
		sm_col_insert(b, value)
	}
	if !slices.Equal(rowValues(first), []int{1, 4, 7, 9}) || !slices.Equal(colValues(a), []int{1, 4, 7, 9}) {
		t.Fatal("sorted uniqueness")
	}
	intersection := sm_row_and(first, second)
	common := sm_col_and(a, b)
	if !slices.Equal(rowValues(intersection), []int{1, 4, 7}) || !slices.Equal(colValues(common), []int{1, 4, 7}) {
		t.Fatal("intersection")
	}
	if !sm_row_contains(intersection, first) || sm_row_contains(first, intersection) || !sm_row_intersects(first, second) || !sm_col_contains(common, a) || !sm_col_intersects(a, b) {
		t.Fatal("set predicates")
	}
	if sm_row_compare(first, second) >= 0 || sm_col_compare(a, b) >= 0 || sm_row_compare(first, sm_row_dup(first)) != 0 {
		t.Fatal("lexical ordering")
	}
	sm_row_remove(first, 4)
	sm_col_remove(a, 4)
	sm_row_remove_element(first, sm_row_find(first, 9))
	sm_col_remove_element(a, sm_col_find(a, 9))
	if !slices.Equal(rowValues(first), []int{1, 7}) || !slices.Equal(colValues(a), []int{1, 7}) {
		t.Fatal("vector removal")
	}
	sm_row_free(first)
	sm_col_free(a)
	if first.length != 0 || a.length != 0 {
		t.Fatal("vector free")
	}
}
func TestSparseHashHandlesLargeNativeIndices(t *testing.T) {
	max := int(^uint(0) >> 1)
	values := []int{-7, 3, max - 1, max}
	row, column := sm_row_alloc(), sm_col_alloc()
	for _, value := range values {
		sm_row_insert(row, value)
		sm_col_insert(column, value)
	}
	modulus := max - 2
	expected := new(big.Int)
	m := big.NewInt(int64(modulus))
	for _, value := range rowValues(row) {
		expected.Mul(expected, big.NewInt(17))
		expected.Add(expected, big.NewInt(int64(value)))
		expected.Mod(expected, m)
	}
	if sm_row_hash(row, modulus) != int(expected.Int64()) || sm_col_hash(column, modulus) != int(expected.Int64()) {
		t.Fatal("overflowed sparse hash")
	}
}
func TestSparseIOAndEmptyPrint(t *testing.T) {
	var matrix *sm_matrix
	if !sm_read(strings.NewReader("4 8\n0 3\n4 8\n2 1\n"), &matrix) {
		t.Fatal("read")
	}
	var output bytes.Buffer
	sm_write(&output, matrix)
	if output.String() != "0 3\n2 1\n4 8\n" {
		t.Fatal(output.String())
	}
	var reread *sm_matrix
	if !sm_read(strings.NewReader(output.String()), &reread) {
		t.Fatal("reread")
	}
	sparseCheck(t, reread, sparseSnapshot(matrix))
	output.Reset()
	sm_print(&output, sm_alloc())
	if output.Len() != 0 {
		t.Fatal(output.String())
	}
	if sm_read(strings.NewReader("0 1\n2"), &matrix) || sm_read(strings.NewReader("-1 2"), &matrix) {
		t.Fatal("malformed read accepted")
	}
	compressed := "2 35\n0 5 4\n0 0 1\n"
	if !sm_read_compressed(strings.NewReader(compressed), &matrix) {
		t.Fatal("compressed read")
	}
	sparseCheck(t, matrix, sparseModel{{0, 0}: true, {0, 2}: true, {0, 34}: true, {1, 32}: true})
}
