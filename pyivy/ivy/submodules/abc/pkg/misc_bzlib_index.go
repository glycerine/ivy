package abc

// Cumulative-frequency lookup from src/misc/bzlib/bzlib.c.

// BZip2IndexIntoF returns the native 0-through-255 bucket for index. The table
// must contain 257 nondecreasing cumulative boundaries.
func BZip2IndexIntoF(index int32, cumulative []int32) (int, error) {
	if len(cumulative) != 257 {
		return 0, ErrDimension
	}
	for position := 1; position < len(cumulative); position++ {
		if cumulative[position] < cumulative[position-1] {
			return 0, ErrInvariant
		}
	}
	below, above := 0, 256
	for above-below != 1 {
		middle := (below + above) >> 1
		if index >= cumulative[middle] {
			below = middle
		} else {
			above = middle
		}
	}
	return below, nil
}
