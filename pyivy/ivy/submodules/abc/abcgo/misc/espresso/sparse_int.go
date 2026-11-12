// Typed list operations replace the sparse_int.h field-substitution macros.
package espresso

import "math/bits"

func sm_link_row_element(row *sm_row, element *sm_element) {
	next := row.first_col
	if row.last_col != nil && row.last_col.col_num < element.col_num {
		next = nil
	} else {
		for next != nil && next.col_num < element.col_num {
			next = next.next_col
		}
	}
	previous := row.last_col
	if next != nil {
		previous = next.prev_col
	}
	element.prev_col, element.next_col = previous, next
	if previous == nil {
		row.first_col = element
	} else {
		previous.next_col = element
	}
	if next == nil {
		row.last_col = element
	} else {
		next.prev_col = element
	}
	row.length++
}
func sm_unlink_row_element(row *sm_row, element *sm_element) {
	if element.prev_col == nil {
		row.first_col = element.next_col
	} else {
		element.prev_col.next_col = element.next_col
	}
	if element.next_col == nil {
		row.last_col = element.prev_col
	} else {
		element.next_col.prev_col = element.prev_col
	}
	row.length--
	element.next_col = nil
	element.prev_col = nil
}
func sm_link_col_element(column *sm_col, element *sm_element) {
	next := column.first_row
	if column.last_row != nil && column.last_row.row_num < element.row_num {
		next = nil
	} else {
		for next != nil && next.row_num < element.row_num {
			next = next.next_row
		}
	}
	previous := column.last_row
	if next != nil {
		previous = next.prev_row
	}
	element.prev_row, element.next_row = previous, next
	if previous == nil {
		column.first_row = element
	} else {
		previous.next_row = element
	}
	if next == nil {
		column.last_row = element
	} else {
		next.prev_row = element
	}
	column.length++
}
func sm_unlink_col_element(column *sm_col, element *sm_element) {
	if element.prev_row == nil {
		column.first_row = element.next_row
	} else {
		element.prev_row.next_row = element.next_row
	}
	if element.next_row == nil {
		column.last_row = element.prev_row
	} else {
		element.next_row.prev_row = element.prev_row
	}
	column.length--
	element.next_row = nil
	element.prev_row = nil
}

// Compute (sum*17+value) mod modulus without signed or unsigned overflow. Matrix
// indices are nonnegative; normalize negative standalone-vector entries too.
func sm_hash_step(sum, value, modulus int) int {
	if modulus <= 0 {
		panic("espresso: nonpositive sparse hash modulus")
	}
	value %= modulus
	if value < 0 {
		value += modulus
	}
	hi, lo := bits.Mul64(uint64(sum), 17)
	lo, carry := bits.Add64(lo, uint64(value), 0)
	hi += carry
	_, remainder := bits.Div64(hi, lo, uint64(modulus))
	return int(remainder)
}

// The default C allocator has no free lists. Go garbage collection provides the
// same cleanup policy, so there is no package-owned pool to drain.
func sm_cleanup() {}
