package abc

// Hand-checked skeleton of src/misc/zlib/deflate.c.

// BlockState is the block_state enum.
type BlockState int

const (
	need_more      BlockState = iota // block not completed, need more input or more output
	block_done                       // block flush performed
	finish_started                   // finish started, need only more output at next deflate
	finish_done                      // finish done, accept no more input or output
)

// Config is config (struct config_s). configuration_table has 10 entries
// (FASTEST is not defined).
type Config struct {
	good_length uint16                              // ush : reduce lazy search above this match length
	max_lazy    uint16                              // ush : do not perform lazy search above this match length
	nice_length uint16                              // ush : quit search above this match length
	max_chain   uint16                              // ush
	fn          func(*DeflateState, int) BlockState // compress_func func
}

// deflate.c also declares `struct static_tree_desc_s {int dummy;}` under
// #ifndef NO_DUMMY_DECL; that placeholder is never used and is not represented
// (see StaticTreeDesc in trees.c).
// The FASTEST variant of longest_match and the DEBUG-only check_match are not
// compiled and are not represented.

// deflateInit_ (ZEXPORT)
// C: int ZEXPORT deflateInit_(z_streamp strm, int level, const char *version, int stream_size)
func (strm *ZStream) DeflateInit_(level int, version string, stream_size int) int {
	panic("TODO, stubbed")
}

// deflateInit2_ (ZEXPORT)
// C: int ZEXPORT deflateInit2_(z_streamp strm, int level, int method, int windowBits, int memLevel, int strategy, const char *version, int stream_size)
func (strm *ZStream) DeflateInit2_(level int, method int, windowBits int, memLevel int, strategy int, version string, stream_size int) int {
	panic("TODO, stubbed")
}

// deflateSetDictionary (ZEXPORT)
// C: int ZEXPORT deflateSetDictionary(z_streamp strm, const Bytef *dictionary, uInt dictLength)
func (strm *ZStream) DeflateSetDictionary(dictionary []byte, dictLength uint32) int {
	panic("TODO, stubbed")
}

// deflateReset (ZEXPORT)
// C: int ZEXPORT deflateReset(z_streamp strm)
func (strm *ZStream) DeflateReset() int {
	panic("TODO, stubbed")
}

// deflateSetHeader (ZEXPORT)
// C: int ZEXPORT deflateSetHeader(z_streamp strm, gz_headerp head)
func (strm *ZStream) DeflateSetHeader(head *GzHeader) int {
	panic("TODO, stubbed")
}

// deflatePrime (ZEXPORT)
// C: int ZEXPORT deflatePrime(z_streamp strm, int bits, int value)
func (strm *ZStream) DeflatePrime(bits int, value int) int {
	panic("TODO, stubbed")
}

// deflateParams (ZEXPORT)
// C: int ZEXPORT deflateParams(z_streamp strm, int level, int strategy)
func (strm *ZStream) DeflateParams(level int, strategy int) int {
	panic("TODO, stubbed")
}

// deflateTune (ZEXPORT)
// C: int ZEXPORT deflateTune(z_streamp strm, int good_length, int max_lazy, int nice_length, int max_chain)
func (strm *ZStream) DeflateTune(good_length int, max_lazy int, nice_length int, max_chain int) int {
	panic("TODO, stubbed")
}

// deflateBound (ZEXPORT)
// C: uLong ZEXPORT deflateBound(z_streamp strm, uLong sourceLen)
func (strm *ZStream) DeflateBound(sourceLen uint64) uint64 {
	panic("TODO, stubbed")
}

// putShortMSB (local)
// C: local void putShortMSB(deflate_state *s, uInt b)
func (s *DeflateState) putShortMSB(b uint32) {
	panic("TODO, stubbed")
}

// flush_pending (local)
// C: local void flush_pending(z_streamp strm)
func (strm *ZStream) flush_pending() {
	panic("TODO, stubbed")
}

// deflate (ZEXPORT)
// C: int ZEXPORT deflate(z_streamp strm, int flush)
func (strm *ZStream) Deflate(flush int) int {
	panic("TODO, stubbed")
}

// deflateEnd (ZEXPORT)
// C: int ZEXPORT deflateEnd(z_streamp strm)
func (strm *ZStream) DeflateEnd() int {
	panic("TODO, stubbed")
}

// deflateCopy (ZEXPORT)
// C: int ZEXPORT deflateCopy(z_streamp dest, z_streamp source)
func (dest *ZStream) DeflateCopy(source *ZStream) int {
	panic("TODO, stubbed")
}

// read_buf (local)
// C: local int read_buf(z_streamp strm, Bytef *buf, unsigned size)
func (strm *ZStream) read_buf(buf []byte, size uint32) int {
	panic("TODO, stubbed")
}

// lm_init (local)
// C: local void lm_init(deflate_state *s)
func (s *DeflateState) lm_init() {
	panic("TODO, stubbed")
}

// longest_match (local; the live variant (FASTEST is undefined))
// C: local uInt longest_match(deflate_state *s, IPos cur_match)
func (s *DeflateState) longest_match(cur_match uint32) uint32 {
	panic("TODO, stubbed")
}

// fill_window (local)
// C: local void fill_window(deflate_state *s)
func (s *DeflateState) fill_window() {
	panic("TODO, stubbed")
}

// deflate_stored (local)
// C: local block_state deflate_stored(deflate_state *s, int flush)
func (s *DeflateState) deflate_stored(flush int) BlockState {
	panic("TODO, stubbed")
}

// deflate_fast (local)
// C: local block_state deflate_fast(deflate_state *s, int flush)
func (s *DeflateState) deflate_fast(flush int) BlockState {
	panic("TODO, stubbed")
}

// deflate_slow (local)
// C: local block_state deflate_slow(deflate_state *s, int flush)
func (s *DeflateState) deflate_slow(flush int) BlockState {
	panic("TODO, stubbed")
}

// deflate_rle (local)
// C: local block_state deflate_rle(deflate_state *s, int flush)
func (s *DeflateState) deflate_rle(flush int) BlockState {
	panic("TODO, stubbed")
}

// deflate_huff (local)
// C: local block_state deflate_huff(deflate_state *s, int flush)
func (s *DeflateState) deflate_huff(flush int) BlockState {
	panic("TODO, stubbed")
}
