package cec

// Cec4 uses the same phase-normalized simulation hash as Cec3.
func Cec4_ManSimHashKey(sim []uint64, nSims, tableSize int32) int32 {
	return Cec3_ManSimHashKey(sim, nSims, tableSize)
}
