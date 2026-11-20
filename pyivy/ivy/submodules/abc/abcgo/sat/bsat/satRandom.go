package bsat

// drand preserves MiniSat's reproducible floating-point recurrence. This is a
// heuristic generator, not a cryptographic source. Callers retain their seed.
func drand(seed *float64) float64 {
	const modulus = 1<<31 - 1
	// For a positive integral seed, the floating-point product is exact.
	// Since 2^31 = 1 modulo the divisor, one fold and subtraction suffice.
	if n := int32(*seed); n > 0 && float64(n) == *seed {
		product := uint64(n) * 1389796
		remainder := (product & modulus) + (product >> 31)
		if remainder >= modulus {
			remainder -= modulus
		}
		*seed = float64(uint32(remainder))
		return *seed / modulus
	}
	*seed *= 1389796
	q := int32(*seed / modulus)
	*seed -= float64(q) * modulus
	return *seed / modulus
}
func irand(seed *float64, size int32) int32 { return int32(drand(seed) * float64(size)) }
