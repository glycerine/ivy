// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package utils

func stk_uint_max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}
func stk_uint_compare(a, b *uint32) int32 {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
