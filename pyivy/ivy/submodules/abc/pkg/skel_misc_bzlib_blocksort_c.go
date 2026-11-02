package abc

// Hand-checked skeleton of src/misc/bzlib/blocksort.c.

// fallbackSimpleSort (static)
// C: void fallbackSimpleSort( UInt32* fmap, UInt32* eclass, Int32 lo, Int32 hi )
func fallbackSimpleSort(fmap []uint32, eclass []uint32, lo int32, hi int32) {
	panic("TODO, stubbed")
}

// fallbackQSort3 (static)
// C: void fallbackQSort3( UInt32* fmap, UInt32* eclass, Int32 loSt, Int32 hiSt )
func fallbackQSort3(fmap []uint32, eclass []uint32, loSt int32, hiSt int32) {
	panic("TODO, stubbed")
}

// fallbackSort (static)
// C: void fallbackSort( UInt32* fmap, UInt32* eclass, UInt32* bhtab, Int32 nblock, Int32 verb )
func fallbackSort(fmap []uint32, eclass []uint32, bhtab []uint32, nblock int32, verb int32) {
	panic("TODO, stubbed")
}

// mainGtU (static inline)
// C: Bool mainGtU( UInt32 i1, UInt32 i2, UChar* block, UInt16* quadrant, UInt32 nblock, Int32* budget )
func mainGtU(i1 uint32, i2 uint32, block []uint8, quadrant []uint16, nblock uint32, budget *int32) bool {
	panic("TODO, stubbed")
}

// mainSimpleSort (static)
// C: void mainSimpleSort( UInt32* ptr, UChar* block, UInt16* quadrant, Int32 nblock, Int32 lo, Int32 hi, Int32 d, Int32* budget )
func mainSimpleSort(ptr []uint32, block []uint8, quadrant []uint16, nblock int32, lo int32, hi int32, d int32, budget *int32) {
	panic("TODO, stubbed")
}

// mmed3 (static inline)
// C: UChar mmed3( UChar a, UChar b, UChar c )
func mmed3(a uint8, b uint8, c uint8) uint8 {
	panic("TODO, stubbed")
}

// mainQSort3 (static)
// C: void mainQSort3( UInt32* ptr, UChar* block, UInt16* quadrant, Int32 nblock, Int32 loSt, Int32 hiSt, Int32 dSt, Int32* budget )
func mainQSort3(ptr []uint32, block []uint8, quadrant []uint16, nblock int32, loSt int32, hiSt int32, dSt int32, budget *int32) {
	panic("TODO, stubbed")
}

// mainSort (static)
// C: void mainSort( UInt32* ptr, UChar* block, UInt16* quadrant, UInt32* ftab, Int32 nblock, Int32 verb, Int32* budget )
func mainSort(ptr []uint32, block []uint8, quadrant []uint16, ftab []uint32, nblock int32, verb int32, budget *int32) {
	panic("TODO, stubbed")
}

// BZ2_blockSort
// C: void BZ2_blockSort( EState* s )
func (s *EState) BZ2_blockSort() {
	panic("TODO, stubbed")
}
