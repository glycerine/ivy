// Standalone col vectors, ported from Espresso cols.c.
package espresso

import (
	"fmt"
	"io"
)

func sm_col_alloc() *sm_col { return &sm_col{} }
func sm_col_free(vector *sm_col) {
	if vector == nil {
		return
	}
	if vector.first_row != nil && vector.first_row.matrix != nil {
		panic("espresso: matrix-owned vectors require matrix operations")
	}
	for element := vector.first_row; element != nil; {
		next := element.next_row
		*element = sm_element{}
		element = next
	}
	*vector = sm_col{}
}
func sm_col_dup(vector *sm_col) *sm_col {
	result := sm_col_alloc()
	for element := vector.first_row; element != nil; element = element.next_row {
		sm_col_insert(result, element.row_num)
	}
	return result
}
func sm_col_insert(vector *sm_col, index int) *sm_element {
	if vector.first_row != nil && vector.first_row.matrix != nil {
		panic("espresso: matrix-owned vectors require matrix operations")
	}
	if element := sm_col_find(vector, index); element != nil {
		return element
	}
	element := &sm_element{row_num: index}
	sm_link_col_element(vector, element)
	return element
}
func sm_col_find(vector *sm_col, index int) *sm_element {
	for element := vector.first_row; element != nil && element.row_num <= index; element = element.next_row {
		if element.row_num == index {
			return element
		}
	}
	return nil
}
func sm_col_remove(vector *sm_col, index int) {
	if vector.first_row != nil && vector.first_row.matrix != nil {
		panic("espresso: matrix-owned vectors require matrix operations")
	}
	if element := sm_col_find(vector, index); element != nil {
		sm_unlink_col_element(vector, element)
		*element = sm_element{}
	}
}
func sm_col_remove_element(vector *sm_col, element *sm_element) {
	if element != nil && element.matrix != nil {
		panic("espresso: matrix-owned vectors require matrix operations")
	}
	if element == nil {
		return
	}
	if sm_col_find(vector, element.row_num) != element {
		panic("espresso: foreign sparse col element")
	}
	sm_unlink_col_element(vector, element)
	*element = sm_element{}
}

// Contains follows the source argument order: second contains first.
func sm_col_contains(first, second *sm_col) bool {
	a, b := first.first_row, second.first_row
	for a != nil {
		if b == nil || a.row_num < b.row_num {
			return false
		}
		if a.row_num == b.row_num {
			a = a.next_row
		}
		b = b.next_row
	}
	return true
}
func sm_col_intersects(first, second *sm_col) bool {
	a, b := first.first_row, second.first_row
	for a != nil && b != nil {
		if a.row_num == b.row_num {
			return true
		}
		if a.row_num < b.row_num {
			a = a.next_row
		} else {
			b = b.next_row
		}
	}
	return false
}
func sm_col_compare(first, second *sm_col) int {
	a, b := first.first_row, second.first_row
	for a != nil && b != nil {
		if a.row_num != b.row_num {
			if a.row_num < b.row_num {
				return -1
			}
			return 1
		}
		a = a.next_row
		b = b.next_row
	}
	if a != nil {
		return 1
	}
	if b != nil {
		return -1
	}
	return 0
}
func sm_col_and(first, second *sm_col) *sm_col {
	result := sm_col_alloc()
	a, b := first.first_row, second.first_row
	for a != nil && b != nil {
		if a.row_num == b.row_num {
			sm_col_insert(result, a.row_num)
			a = a.next_row
			b = b.next_row
		} else if a.row_num < b.row_num {
			a = a.next_row
		} else {
			b = b.next_row
		}
	}
	return result
}
func sm_col_hash(vector *sm_col, modulus int) int {
	if modulus <= 0 {
		panic("espresso: nonpositive sparse hash modulus")
	}
	sum := 0
	for element := vector.first_row; element != nil; element = element.next_row {
		sum = sm_hash_step(sum, element.row_num, modulus)
	}
	return sum
}
func sm_col_print(w io.Writer, vector *sm_col) {
	for element := vector.first_row; element != nil; element = element.next_row {
		fmt.Fprintf(w, " %d", element.row_num)
	}
}
