// Standalone row vectors, ported from Espresso rows.c.
package espresso

import (
	"fmt"
	"io"
)

func sm_row_alloc() *sm_row { return &sm_row{} }
func sm_row_free(vector *sm_row) {
	if vector == nil {
		return
	}
	if vector.first_col != nil && vector.first_col.matrix != nil {
		panic("espresso: matrix-owned vectors require matrix operations")
	}
	for element := vector.first_col; element != nil; {
		next := element.next_col
		*element = sm_element{}
		element = next
	}
	*vector = sm_row{}
}
func sm_row_dup(vector *sm_row) *sm_row {
	result := sm_row_alloc()
	for element := vector.first_col; element != nil; element = element.next_col {
		sm_row_insert(result, element.col_num)
	}
	return result
}
func sm_row_insert(vector *sm_row, index int) *sm_element {
	if vector.first_col != nil && vector.first_col.matrix != nil {
		panic("espresso: matrix-owned vectors require matrix operations")
	}
	if element := sm_row_find(vector, index); element != nil {
		return element
	}
	element := &sm_element{col_num: index}
	sm_link_row_element(vector, element)
	return element
}
func sm_row_find(vector *sm_row, index int) *sm_element {
	for element := vector.first_col; element != nil && element.col_num <= index; element = element.next_col {
		if element.col_num == index {
			return element
		}
	}
	return nil
}
func sm_row_remove(vector *sm_row, index int) {
	if vector.first_col != nil && vector.first_col.matrix != nil {
		panic("espresso: matrix-owned vectors require matrix operations")
	}
	if element := sm_row_find(vector, index); element != nil {
		sm_unlink_row_element(vector, element)
		*element = sm_element{}
	}
}
func sm_row_remove_element(vector *sm_row, element *sm_element) {
	if element != nil && element.matrix != nil {
		panic("espresso: matrix-owned vectors require matrix operations")
	}
	if element == nil {
		return
	}
	if sm_row_find(vector, element.col_num) != element {
		panic("espresso: foreign sparse row element")
	}
	sm_unlink_row_element(vector, element)
	*element = sm_element{}
}

// Contains follows the source argument order: second contains first.
func sm_row_contains(first, second *sm_row) bool {
	a, b := first.first_col, second.first_col
	for a != nil {
		if b == nil || a.col_num < b.col_num {
			return false
		}
		if a.col_num == b.col_num {
			a = a.next_col
		}
		b = b.next_col
	}
	return true
}
func sm_row_intersects(first, second *sm_row) bool {
	a, b := first.first_col, second.first_col
	for a != nil && b != nil {
		if a.col_num == b.col_num {
			return true
		}
		if a.col_num < b.col_num {
			a = a.next_col
		} else {
			b = b.next_col
		}
	}
	return false
}
func sm_row_compare(first, second *sm_row) int {
	a, b := first.first_col, second.first_col
	for a != nil && b != nil {
		if a.col_num != b.col_num {
			if a.col_num < b.col_num {
				return -1
			}
			return 1
		}
		a = a.next_col
		b = b.next_col
	}
	if a != nil {
		return 1
	}
	if b != nil {
		return -1
	}
	return 0
}
func sm_row_and(first, second *sm_row) *sm_row {
	result := sm_row_alloc()
	a, b := first.first_col, second.first_col
	for a != nil && b != nil {
		if a.col_num == b.col_num {
			sm_row_insert(result, a.col_num)
			a = a.next_col
			b = b.next_col
		} else if a.col_num < b.col_num {
			a = a.next_col
		} else {
			b = b.next_col
		}
	}
	return result
}
func sm_row_hash(vector *sm_row, modulus int) int {
	if modulus <= 0 {
		panic("espresso: nonpositive sparse hash modulus")
	}
	sum := 0
	for element := vector.first_col; element != nil; element = element.next_col {
		sum = sm_hash_step(sum, element.col_num, modulus)
	}
	return sum
}
func sm_row_print(w io.Writer, vector *sm_row) {
	for element := vector.first_col; element != nil; element = element.next_col {
		fmt.Fprintf(w, " %d", element.col_num)
	}
}
