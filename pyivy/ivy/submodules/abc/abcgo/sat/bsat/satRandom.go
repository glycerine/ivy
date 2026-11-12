package bsat

// drand preserves MiniSat's reproducible floating-point recurrence. This is a
// heuristic generator, not a cryptographic source. Callers retain their seed.
func drand(seed *float64) float64 {
	*seed *= 1389796
	q := int32(*seed / 2147483647)
	*seed -= float64(q) * 2147483647
	return *seed / 2147483647
}
func irand(seed *float64, size int32) int32 { return int32(drand(seed) * float64(size)) }
