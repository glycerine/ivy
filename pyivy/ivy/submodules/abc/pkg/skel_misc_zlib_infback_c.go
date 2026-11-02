package abc

// Hand-checked skeleton of src/misc/zlib/infback.c.

// inflateBackInit_ (ZEXPORT)
// C: int ZEXPORT inflateBackInit_(z_streamp strm, int windowBits, unsigned char FAR *window, const char *version, int stream_size)
func (strm *ZStream) InflateBackInit_(windowBits int, window []byte, version string, stream_size int) int {
	panic("TODO, stubbed")
}

// fixedtables (local; inflate.c defines a local function of the same name, so this one carries an _infback suffix)
// C: local void fixedtables(struct inflate_state FAR *state)
func (state *InflateState) fixedtables_infback() {
	panic("TODO, stubbed")
}

// inflateBack (ZEXPORT; in_func is unsigned (*)(void *, unsigned char **), out_func is int (*)(void *, unsigned char *, unsigned))
// C: int ZEXPORT inflateBack(z_streamp strm, in_func in, void FAR *in_desc, out_func out, void FAR *out_desc)
func (strm *ZStream) InflateBack(in func(any, *[]byte) uint32, in_desc any, out func(any, []byte, uint32) int, out_desc any) int {
	panic("TODO, stubbed")
}

// inflateBackEnd (ZEXPORT)
// C: int ZEXPORT inflateBackEnd(z_streamp strm)
func (strm *ZStream) InflateBackEnd() int {
	panic("TODO, stubbed")
}
