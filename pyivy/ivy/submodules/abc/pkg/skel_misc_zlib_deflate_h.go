package abc

// Hand-written skeleton of the types in src/misc/zlib/deflate.h. LENGTH_CODES
// is 29, LITERALS 256, L_CODES 286, D_CODES 30, BL_CODES 19, HEAP_SIZE 573 and
// MAX_BITS 15. The DEBUG-only members compressed_len and bits_sent are not
// compiled by default and are omitted. The macro max_insert_length is an alias
// of max_lazy_match.

// CtData is ct_data (struct ct_data_s). The two anonymous unions are written
// as both members each.
type CtData struct {
	fcFreq uint16 // union fc : ush freq, frequency count
	fcCode uint16 // union fc : ush code, bit string
	dlDad  uint16 // union dl : ush dad, father node in Huffman tree
	dlLen  uint16 // union dl : ush len, length of bit string
}

// TreeDesc is tree_desc (struct tree_desc_s).
type TreeDesc struct {
	dyn_tree  []CtData        // ct_data * : the dynamic tree
	max_code  int             // largest code with non zero frequency
	stat_desc *StaticTreeDesc // the corresponding static tree
}

// DeflateState is deflate_state (struct internal_state).
type DeflateState struct {
	strm             *ZStream  // pointer back to this zlib stream
	status           int       // as the name implies
	pending_buf      []byte    // Bytef * : output still pending
	pending_buf_size uint64    // ulg : size of pending_buf
	pending_out      []byte    // Bytef * : next pending byte to output to the stream
	pending          uint32    // uInt : nb of bytes in the pending buffer
	wrap             int       // bit 0 true for zlib, bit 1 true for gzip
	gzhead           *GzHeader // gzip header information to write
	gzindex          uint32    // uInt : where in extra, name, or comment
	method           uint8     // Byte : STORED (for zip only) or DEFLATED
	last_flush       int       // value of flush param for previous deflate call
	// used by deflate.c:
	w_size           uint32   // uInt : LZ77 window size (32K by default)
	w_bits           uint32   // uInt : log2(w_size)  (8..16)
	w_mask           uint32   // uInt : w_size - 1
	window           []byte   // Bytef * : sliding window
	window_size      uint64   // ulg : actual size of window: 2*wSize, except when the user input buffer is directly used as sliding window
	prev             []uint16 // Posf * : link to older string with same hash index
	head             []uint16 // Posf * : heads of the hash chains or NIL
	ins_h            uint32   // uInt : hash index of string to be inserted
	hash_size        uint32   // uInt : number of elements in hash table
	hash_bits        uint32   // uInt : log2(hash_size)
	hash_mask        uint32   // uInt : hash_size-1
	hash_shift       uint32   // uInt : number of bits by which ins_h must be shifted at each input step
	block_start      int64    // long : window position at the beginning of the current output block
	match_length     uint32   // uInt : length of best match
	prev_match       uint32   // IPos : previous match
	match_available  int      // set if previous match exists
	strstart         uint32   // uInt : start of string to insert
	match_start      uint32   // uInt : start of matching string
	lookahead        uint32   // uInt : number of valid bytes ahead in window
	prev_length      uint32   // uInt : length of the best match at previous step
	max_chain_length uint32   // uInt : hash chains are never searched beyond this length
	max_lazy_match   uint32   // uInt : attempt to find a better match only when the current match is strictly smaller than this value
	level            int      // compression level (1..9)
	strategy         int      // favor or force Huffman coding
	good_match       uint32   // uInt : use a faster search when the previous match is longer than this
	nice_match       int      // stop searching when current match exceeds this
	// used by trees.c:
	dyn_ltree    [573]CtData // struct ct_data_s dyn_ltree[HEAP_SIZE] : literal and length tree
	dyn_dtree    [61]CtData  // struct ct_data_s dyn_dtree[2*D_CODES+1] : distance tree
	bl_tree      [39]CtData  // struct ct_data_s bl_tree[2*BL_CODES+1] : Huffman tree for bit lengths
	l_desc       TreeDesc    // desc. for literal tree
	d_desc       TreeDesc    // desc. for distance tree
	bl_desc      TreeDesc    // desc. for bit length tree
	bl_count     [16]uint16  // ush bl_count[MAX_BITS+1] : number of codes at each bit length for an optimal tree
	heap         [573]int    // heap used to build the Huffman trees
	heap_len     int         // number of elements in the heap
	heap_max     int         // element of largest frequency
	depth        [573]uint8  // uch depth[2*L_CODES+1] : depth of each subtree used as tie breaker for trees of equal frequency
	l_buf        []uint8     // uchf * : buffer for literals or lengths
	lit_bufsize  uint32      // uInt : size of match buffer for literals/lengths
	last_lit     uint32      // uInt : running index in l_buf
	d_buf        []uint16    // ushf * : buffer for distances
	opt_len      uint64      // ulg : bit length of current block with optimal trees
	static_len   uint64      // ulg : bit length of current block with static trees
	matches      uint32      // uInt : number of string matches in current block
	last_eob_len int         // bit length of EOB code for last block
	bi_buf       uint16      // ush : output buffer; bits are inserted starting at the bottom
	bi_valid     int         // number of valid bits in bi_buf
	high_water   uint64      // ulg : high water mark offset in window for initialized bytes
}
