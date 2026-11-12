// Package hash implements ABC's chained hash tables using Go-owned storage.
package hash

// Hash_DefaultHashFunc preserves the original 32-bit C integer arithmetic.
func Hash_DefaultHashFunc(key, nBins int32) int32 {
	if nBins <= 0 {
		panic("hash: nonpositive bin count")
	}
	h := ((key+11)*key*7 + 3) % nBins
	if h < 0 {
		h = -h
	}
	return h
}
