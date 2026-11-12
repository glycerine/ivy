package bmc

// packed2Get reads one two-bit field from a word containing 16 fields.
func packed2Get(word uint32, objectID uint32) uint32 {
	shift := (objectID & 15) << 1
	return (word >> shift) & 3
}

// packed2Set replaces one two-bit field without disturbing its neighbours.
func packed2Set(word uint32, objectID uint32, value uint32) uint32 {
	shift := (objectID & 15) << 1
	mask := uint32(3) << shift
	return (word &^ mask) | ((value & 3) << shift)
}

// packed2Add ORs a two-bit value into one field, matching Abc_InfoAdd2Bits.
func packed2Add(word uint32, objectID uint32, value uint32) uint32 {
	shift := (objectID & 15) << 1
	return word | ((value & 3) << shift)
}
