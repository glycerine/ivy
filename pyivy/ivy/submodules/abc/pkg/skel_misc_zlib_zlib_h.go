package abc

// Hand-written skeleton of the types in src/misc/zlib/zlib.h. uInt is
// unsigned, uLong is unsigned long, Bytef is a byte, voidpf is void *.

// ZStream is z_stream. The `state` member is `struct internal_state *`, which
// deflate.c defines as deflate_state and inflate.c reinterprets as
// inflate_state; it is kept as the declared deflate_state.
type ZStream struct {
	next_in   []byte                                          // Bytef * : next input byte
	avail_in  uint32                                          // uInt : number of bytes available at next_in
	total_in  uint64                                          // uLong : total nb of input bytes read so far
	next_out  []byte                                          // Bytef * : next output byte should be put there
	avail_out uint32                                          // uInt : remaining free space at next_out
	total_out uint64                                          // uLong : total nb of bytes output so far
	msg       string                                          // char * : last error message, NULL if no error
	state     *DeflateState                                   // struct internal_state * : not visible by applications
	zalloc    func(opaque any, items uint32, size uint32) any // alloc_func : used to allocate the internal state
	zfree     func(opaque any, address any)                   // free_func : used to free the internal state
	opaque    any                                             // voidpf : private data object passed to zalloc and zfree
	data_type int                                             // best guess about the data type: binary or text
	adler     uint64                                          // uLong : adler32 value of the uncompressed data
	reserved  uint64                                          // uLong : reserved for future use
}

// GzHeader is gz_header.
type GzHeader struct {
	text      int    // true if compressed data believed to be text
	time      uint64 // uLong : modification time
	xflags    int    // extra flags (not used when writing a gzip file)
	os        int    // operating system
	extra     []byte // Bytef * : pointer to extra field or Z_NULL if none
	extra_len uint32 // uInt : extra field length (valid if extra != Z_NULL)
	extra_max uint32 // uInt : space at extra (only when reading header)
	name      []byte // Bytef * : pointer to zero-terminated file name or Z_NULL
	name_max  uint32 // uInt : space at name (only when reading header)
	comment   []byte // Bytef * : pointer to zero-terminated comment or Z_NULL
	comm_max  uint32 // uInt : space at comment (only when reading header)
	hcrc      int    // true if there was or will be a header crc
	done      int    // true when done reading gzip header (not used when writing a gzip file)
}
