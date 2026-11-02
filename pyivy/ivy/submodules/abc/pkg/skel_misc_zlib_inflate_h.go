package abc

// Hand-written skeleton of the types in src/misc/zlib/inflate.h. ENOUGH is
// ENOUGH_LENS + ENOUGH_DISTS = 852 + 592 = 1444.

// InflateMode is the inflate_mode enum.
type InflateMode int

const (
	HEAD     InflateMode = iota // i: waiting for magic header
	FLAGS                       // i: waiting for method and flags (gzip)
	TIME                        // i: waiting for modification time (gzip)
	OS                          // i: waiting for extra flags and operating system (gzip)
	EXLEN                       // i: waiting for extra length (gzip)
	EXTRA                       // i: waiting for extra bytes (gzip)
	NAME                        // i: waiting for end of file name (gzip)
	COMMENT                     // i: waiting for end of comment (gzip)
	HCRC                        // i: waiting for header crc (gzip)
	DICTID                      // i: waiting for dictionary check value
	DICT                        // waiting for inflateSetDictionary() call
	TYPE                        // i: waiting for type bits, including last-flag bit
	TYPEDO                      // i: same, but skip check to exit inflate on new block
	STORED                      // i: waiting for stored size (length and complement)
	COPY_                       // i/o: same as COPY below, but only first time in
	COPY                        // i/o: waiting for input or output to copy stored block
	TABLE                       // i: waiting for dynamic block table lengths
	LENLENS                     // i: waiting for code length code lengths
	CODELENS                    // i: waiting for length/lit and distance code lengths
	LEN_                        // i: same as LEN below, but only first time in
	LEN                         // i: waiting for length/lit/eob code
	LENEXT                      // i: waiting for length extra bits
	DIST                        // i: waiting for distance code
	DISTEXT                     // i: waiting for distance extra bits
	MATCH                       // o: waiting for output space to copy string
	LIT                         // o: waiting for output space to write literal
	CHECK                       // i: waiting for 32-bit check value
	LENGTH                      // i: waiting for 32-bit length (gzip)
	DONE                        // finished check, done -- remain here until reset
	BAD                         // got a data error -- remain here until reset
	MEM                         // got an inflate() memory error -- remain here until reset
	SYNC                        // looking for synchronization bytes to restart inflate()
)

// InflateState is struct inflate_state.
type InflateState struct {
	mode     InflateMode // current inflate mode
	last     int         // true if processing last block
	wrap     int         // bit 0 true for zlib, bit 1 true for gzip
	havedict int         // true if dictionary provided
	flags    int         // gzip header method and flags (0 if zlib)
	dmax     uint32      // unsigned : zlib header max distance (INFLATE_STRICT)
	check    uint64      // unsigned long : protected copy of check value
	total    uint64      // unsigned long : protected copy of output count
	head     *GzHeader   // gz_headerp : where to save gzip header information
	wbits    uint32      // unsigned : log base 2 of requested window size
	wsize    uint32      // unsigned : window size or zero if not using window
	whave    uint32      // unsigned : valid bytes in the window
	wnext    uint32      // unsigned : window write index
	window   []byte      // unsigned char * : allocated sliding window, if needed
	hold     uint64      // unsigned long : input bit accumulator
	bits     uint32      // unsigned : number of bits in "in"
	length   uint32      // unsigned : literal or length of data to copy
	offset   uint32      // unsigned : distance back to copy string from
	extra    uint32      // unsigned : extra bits needed
	lencode  []Code      // code const * : starting table for length/literal codes
	distcode []Code      // code const * : starting table for distance codes
	lenbits  uint32      // unsigned : index bits for lencode
	distbits uint32      // unsigned : index bits for distcode
	ncode    uint32      // unsigned : number of code length code lengths
	nlen     uint32      // unsigned : number of length code lengths
	ndist    uint32      // unsigned : number of distance code lengths
	have     uint32      // unsigned : number of code lengths in lens[]
	next     []Code      // code * : next available space in codes[]
	lens     [320]uint16 // unsigned short : temporary storage for code lengths
	work     [288]uint16 // unsigned short : work area for code table building
	codes    [1444]Code  // code codes[ENOUGH] : space for code tables
	sane     int         // if false, allow invalid distance too far
	back     int         // bits back of last unprocessed length/lit
	was      uint32      // unsigned : initial length of match
}
