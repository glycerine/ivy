package abc

// Hand-checked skeleton of src/misc/bzlib/compress.c.

// BZ2_bsInitWrite
// C: void BZ2_bsInitWrite( EState* s )
func (s *EState) BZ2_bsInitWrite() {
	panic("TODO, stubbed")
}

// bsFinishWrite (static)
// C: void bsFinishWrite( EState* s )
func (s *EState) bsFinishWrite() {
	panic("TODO, stubbed")
}

// bsW (static inline)
// C: void bsW( EState* s, Int32 n, UInt32 v )
func (s *EState) bsW(n int32, v uint32) {
	panic("TODO, stubbed")
}

// bsPutUInt32 (static)
// C: void bsPutUInt32( EState* s, UInt32 u )
func (s *EState) bsPutUInt32(u uint32) {
	panic("TODO, stubbed")
}

// bsPutUChar (static)
// C: void bsPutUChar( EState* s, UChar c )
func (s *EState) bsPutUChar(c uint8) {
	panic("TODO, stubbed")
}

// makeMaps_e (static)
// C: void makeMaps_e( EState* s )
func (s *EState) makeMaps_e() {
	panic("TODO, stubbed")
}

// generateMTFValues (static)
// C: void generateMTFValues( EState* s )
func (s *EState) generateMTFValues() {
	panic("TODO, stubbed")
}

// sendMTFValues (static)
// C: void sendMTFValues( EState* s )
func (s *EState) sendMTFValues() {
	panic("TODO, stubbed")
}

// BZ2_compressBlock
// C: void BZ2_compressBlock( EState* s, Bool is_last_block )
func (s *EState) BZ2_compressBlock(is_last_block bool) {
	panic("TODO, stubbed")
}
