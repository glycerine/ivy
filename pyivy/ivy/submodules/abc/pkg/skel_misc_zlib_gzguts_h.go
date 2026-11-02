package abc

// Hand-written skeleton of the gz_state type in src/misc/zlib/gzguts.h.
// z_off64_t is a 64-bit offset.

// GzState is gz_state.
type GzState struct {
	mode     int     // see gzip modes above
	fd       int     // file descriptor
	path     string  // char * : path or fd for error messages
	pos      int64   // z_off64_t : current position in uncompressed data
	size     uint32  // unsigned : buffer size, zero if not allocated yet
	want     uint32  // unsigned : requested buffer size, default is GZBUFSIZE
	in       []byte  // unsigned char * : input buffer
	out      []byte  // unsigned char * : output buffer (double-sized when reading)
	next     []byte  // unsigned char * : next output data to deliver or write
	have     uint32  // unsigned : amount of output data unused at next
	eof      int     // true if end of input file reached
	start    int64   // z_off64_t : where the gzip data started, for rewinding
	raw      int64   // z_off64_t : where the raw data started, for seeking
	how      int     // 0: get header, 1: copy, 2: decompress
	direct   int     // true if last read direct, false if gzip
	level    int     // compression level
	strategy int     // compression strategy
	skip     int64   // z_off64_t : amount to skip (already rewound if backwards)
	seek     int     // true if seek request pending
	err      int     // error code
	msg      string  // char * : error message
	strm     ZStream // z_stream : stream structure in-place (not a pointer)
}
