package abc

// Hand-written skeleton of the types in src/misc/bzlib/bzlib_private.h.
// BZ_MAX_ALPHA_SIZE is 258, BZ_N_GROUPS is 6, BZ_G_SIZE is 50, so
// BZ_MAX_SELECTORS is 2 + 900000/50 = 18002; MTFA_SIZE is 4096 and MTFL_SIZE
// is 16. The BZ_RAND_DECLS macro expands to `Int32 rNToGo; Int32 rTPos`.

// EState is the compression-side state (typedef of an anonymous struct).
type EState struct {
	// pointer back to the struct bz_stream
	strm *BzStream
	// mode this stream is in, and whether inputting or outputting data
	mode  int32 // Int32
	state int32 // Int32
	// remembers avail_in when flush/finish requested
	avail_in_expect uint32 // UInt32
	// for doing the block sorting
	arr1    []uint32 // UInt32 *
	arr2    []uint32 // UInt32 *
	ftab    []uint32 // UInt32 *
	origPtr int32    // Int32
	// aliases for arr1 and arr2
	ptr   []uint32 // UInt32 *
	block []uint8  // UChar *
	mtfv  []uint16 // UInt16 *
	zbits []uint8  // UChar *
	// for deciding when to use the fallback sorting algorithm
	workFactor int32 // Int32
	// run-length-encoding of the input
	state_in_ch  uint32 // UInt32
	state_in_len int32  // Int32
	rNToGo       int32  // BZ_RAND_DECLS : Int32 rNToGo
	rTPos        int32  // BZ_RAND_DECLS : Int32 rTPos
	// input and output limits and current posns
	nblock        int32 // Int32
	nblockMAX     int32 // Int32
	numZ          int32 // Int32
	state_out_pos int32 // Int32
	// map of bytes used in block
	nInUse     int32      // Int32
	inUse      [256]bool  // Bool inUse[256]
	unseqToSeq [256]uint8 // UChar unseqToSeq[256]
	// the buffer for bit stream creation
	bsBuff uint32 // UInt32
	bsLive int32  // Int32
	// block and combined CRCs
	blockCRC    uint32 // UInt32
	combinedCRC uint32 // UInt32
	// misc administratium
	verbosity     int32 // Int32
	blockNo       int32 // Int32
	blockSize100k int32 // Int32
	// stuff for coding the MTF values
	nMTF        int32         // Int32
	mtfFreq     [258]int32    // Int32 mtfFreq[BZ_MAX_ALPHA_SIZE]
	selector    [18002]uint8  // UChar selector[BZ_MAX_SELECTORS]
	selectorMtf [18002]uint8  // UChar selectorMtf[BZ_MAX_SELECTORS]
	len         [6][258]uint8 // UChar len[BZ_N_GROUPS][BZ_MAX_ALPHA_SIZE]
	code        [6][258]int32 // Int32 code[BZ_N_GROUPS][BZ_MAX_ALPHA_SIZE]
	rfreq       [6][258]int32 // Int32 rfreq[BZ_N_GROUPS][BZ_MAX_ALPHA_SIZE]
	// second dimension: only 3 needed; 4 makes index calculations faster
	len_pack [258][4]uint32 // UInt32 len_pack[BZ_MAX_ALPHA_SIZE][4]
}

// DState is the decompression-side state (typedef of an anonymous struct).
type DState struct {
	// pointer back to the struct bz_stream
	strm *BzStream
	// state indicator for this stream
	state int32 // Int32
	// for doing the final run-length decoding
	state_out_ch    uint8 // UChar
	state_out_len   int32 // Int32
	blockRandomised bool  // Bool
	rNToGo          int32 // BZ_RAND_DECLS : Int32 rNToGo
	rTPos           int32 // BZ_RAND_DECLS : Int32 rTPos
	// the buffer for bit stream reading
	bsBuff uint32 // UInt32
	bsLive int32  // Int32
	// misc administratium
	blockSize100k   int32 // Int32
	smallDecompress bool  // Bool
	currBlockNo     int32 // Int32
	verbosity       int32 // Int32
	// for undoing the Burrows-Wheeler transform
	origPtr     int32      // Int32
	tPos        uint32     // UInt32
	k0          int32      // Int32
	unzftab     [256]int32 // Int32 unzftab[256]
	nblock_used int32      // Int32
	cftab       [257]int32 // Int32 cftab[257]
	cftabCopy   [257]int32 // Int32 cftabCopy[257]
	// for undoing the Burrows-Wheeler transform (FAST)
	tt []uint32 // UInt32 *
	// for undoing the Burrows-Wheeler transform (SMALL)
	ll16 []uint16 // UInt16 *
	ll4  []uint8  // UChar *
	// stored and calculated CRCs
	storedBlockCRC        uint32 // UInt32
	storedCombinedCRC     uint32 // UInt32
	calculatedBlockCRC    uint32 // UInt32
	calculatedCombinedCRC uint32 // UInt32
	// map of bytes used in block
	nInUse     int32      // Int32
	inUse      [256]bool  // Bool inUse[256]
	inUse16    [16]bool   // Bool inUse16[16]
	seqToUnseq [256]uint8 // UChar seqToUnseq[256]
	// for decoding the MTF values
	mtfa        [4096]uint8     // UChar mtfa[MTFA_SIZE]
	mtfbase     [256 / 16]int32 // Int32 mtfbase[256 / MTFL_SIZE]
	selector    [18002]uint8    // UChar selector[BZ_MAX_SELECTORS]
	selectorMtf [18002]uint8    // UChar selectorMtf[BZ_MAX_SELECTORS]
	len         [6][258]uint8   // UChar len[BZ_N_GROUPS][BZ_MAX_ALPHA_SIZE]
	limit       [6][258]int32   // Int32 limit[BZ_N_GROUPS][BZ_MAX_ALPHA_SIZE]
	base        [6][258]int32   // Int32 base[BZ_N_GROUPS][BZ_MAX_ALPHA_SIZE]
	perm        [6][258]int32   // Int32 perm[BZ_N_GROUPS][BZ_MAX_ALPHA_SIZE]
	minLens     [6]int32        // Int32 minLens[BZ_N_GROUPS]
	// save area for scalars in the main decompress code
	save_i          int32
	save_j          int32
	save_t          int32
	save_alphaSize  int32
	save_nGroups    int32
	save_nSelectors int32
	save_EOB        int32
	save_groupNo    int32
	save_groupPos   int32
	save_nextSym    int32
	save_nblockMAX  int32
	save_nblock     int32
	save_es         int32
	save_N          int32
	save_curr       int32
	save_zt         int32
	save_zn         int32
	save_zvec       int32
	save_zj         int32
	save_gSel       int32
	save_gMinlen    int32
	save_gLimit     []int32 // Int32 * : points into limit[]
	save_gBase      []int32 // Int32 * : points into base[]
	save_gPerm      []int32 // Int32 * : points into perm[]
}
