package abc

// Hand-checked skeleton of src/misc/zlib/inflate.c.

// The makefixed function is compiled only with MAKEFIXED and is not
// represented.

// inflateReset (ZEXPORT)
// C: int ZEXPORT inflateReset(z_streamp strm)
func (strm *ZStream) InflateReset() int {
	panic("TODO, stubbed")
}

// inflateReset2 (ZEXPORT)
// C: int ZEXPORT inflateReset2(z_streamp strm, int windowBits)
func (strm *ZStream) InflateReset2(windowBits int) int {
	panic("TODO, stubbed")
}

// inflateInit2_ (ZEXPORT)
// C: int ZEXPORT inflateInit2_(z_streamp strm, int windowBits, const char *version, int stream_size)
func (strm *ZStream) InflateInit2_(windowBits int, version string, stream_size int) int {
	panic("TODO, stubbed")
}

// inflateInit_ (ZEXPORT)
// C: int ZEXPORT inflateInit_(z_streamp strm, const char *version, int stream_size)
func (strm *ZStream) InflateInit_(version string, stream_size int) int {
	panic("TODO, stubbed")
}

// inflatePrime (ZEXPORT)
// C: int ZEXPORT inflatePrime(z_streamp strm, int bits, int value)
func (strm *ZStream) InflatePrime(bits int, value int) int {
	panic("TODO, stubbed")
}

// fixedtables (local)
// C: local void fixedtables(struct inflate_state FAR *state)
func (state *InflateState) fixedtables() {
	panic("TODO, stubbed")
}

// updatewindow (local)
// C: local int updatewindow(z_streamp strm, unsigned out)
func (strm *ZStream) updatewindow(out uint32) int {
	panic("TODO, stubbed")
}

// inflate (ZEXPORT)
// C: int ZEXPORT inflate(z_streamp strm, int flush)
func (strm *ZStream) Inflate(flush int) int {
	panic("TODO, stubbed")
}

// inflateEnd (ZEXPORT)
// C: int ZEXPORT inflateEnd(z_streamp strm)
func (strm *ZStream) InflateEnd() int {
	panic("TODO, stubbed")
}

// inflateSetDictionary (ZEXPORT)
// C: int ZEXPORT inflateSetDictionary(z_streamp strm, const Bytef *dictionary, uInt dictLength)
func (strm *ZStream) InflateSetDictionary(dictionary []byte, dictLength uint32) int {
	panic("TODO, stubbed")
}

// inflateGetHeader (ZEXPORT)
// C: int ZEXPORT inflateGetHeader(z_streamp strm, gz_headerp head)
func (strm *ZStream) InflateGetHeader(head *GzHeader) int {
	panic("TODO, stubbed")
}

// syncsearch (local; have is an in/out scalar)
// C: local unsigned syncsearch(unsigned FAR *have, unsigned char FAR *buf, unsigned len)
func syncsearch(have *uint32, buf []byte, len_ uint32) uint32 {
	panic("TODO, stubbed")
}

// inflateSync (ZEXPORT)
// C: int ZEXPORT inflateSync(z_streamp strm)
func (strm *ZStream) InflateSync() int {
	panic("TODO, stubbed")
}

// inflateSyncPoint (ZEXPORT)
// C: int ZEXPORT inflateSyncPoint(z_streamp strm)
func (strm *ZStream) InflateSyncPoint() int {
	panic("TODO, stubbed")
}

// inflateCopy (ZEXPORT)
// C: int ZEXPORT inflateCopy(z_streamp dest, z_streamp source)
func (dest *ZStream) InflateCopy(source *ZStream) int {
	panic("TODO, stubbed")
}

// inflateUndermine (ZEXPORT)
// C: int ZEXPORT inflateUndermine(z_streamp strm, int subvert)
func (strm *ZStream) InflateUndermine(subvert int) int {
	panic("TODO, stubbed")
}

// inflateMark (ZEXPORT)
// C: long ZEXPORT inflateMark(z_streamp strm)
func (strm *ZStream) InflateMark() int64 {
	panic("TODO, stubbed")
}
