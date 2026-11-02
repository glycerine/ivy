package abc

// Hand-checked skeleton of src/misc/zlib/trees.c.

// StaticTreeDesc is static_tree_desc (struct static_tree_desc_s as defined in
// trees.c). The send_bits function is a macro unless DEBUG is defined and
// gen_trees_header is only compiled with GEN_TREES_H, so neither is
// represented. tr_static_init is defined in both configurations (its body is
// conditional).
type StaticTreeDesc struct {
	static_tree []CtData // const ct_data * : static tree or NULL
	extra_bits  []int    // const intf * : extra bits for each code or NULL
	extra_base  int      // base index for extra_bits
	elems       int      // max number of elements in the tree
	max_length  int      // max bit length for the codes
}

// tr_static_init (local)
// C: local void tr_static_init()
func tr_static_init() {
	panic("TODO, stubbed")
}

// _tr_init (ZLIB_INTERNAL)
// C: void ZLIB_INTERNAL _tr_init(deflate_state *s)
func (s *DeflateState) _tr_init() {
	panic("TODO, stubbed")
}

// init_block (local)
// C: local void init_block(deflate_state *s)
func (s *DeflateState) init_block() {
	panic("TODO, stubbed")
}

// pqdownheap (local)
// C: local void pqdownheap(deflate_state *s, ct_data *tree, int k)
func (s *DeflateState) pqdownheap(tree []CtData, k int) {
	panic("TODO, stubbed")
}

// gen_bitlen (local)
// C: local void gen_bitlen(deflate_state *s, tree_desc *desc)
func (s *DeflateState) gen_bitlen(desc *TreeDesc) {
	panic("TODO, stubbed")
}

// gen_codes (local)
// C: local void gen_codes(ct_data *tree, int max_code, ushf *bl_count)
func gen_codes(tree []CtData, max_code int, bl_count []uint16) {
	panic("TODO, stubbed")
}

// build_tree (local)
// C: local void build_tree(deflate_state *s, tree_desc *desc)
func (s *DeflateState) build_tree(desc *TreeDesc) {
	panic("TODO, stubbed")
}

// scan_tree (local)
// C: local void scan_tree(deflate_state *s, ct_data *tree, int max_code)
func (s *DeflateState) scan_tree(tree []CtData, max_code int) {
	panic("TODO, stubbed")
}

// send_tree (local)
// C: local void send_tree(deflate_state *s, ct_data *tree, int max_code)
func (s *DeflateState) send_tree(tree []CtData, max_code int) {
	panic("TODO, stubbed")
}

// build_bl_tree (local)
// C: local int build_bl_tree(deflate_state *s)
func (s *DeflateState) build_bl_tree() int {
	panic("TODO, stubbed")
}

// send_all_trees (local)
// C: local void send_all_trees(deflate_state *s, int lcodes, int dcodes, int blcodes)
func (s *DeflateState) send_all_trees(lcodes int, dcodes int, blcodes int) {
	panic("TODO, stubbed")
}

// _tr_stored_block (ZLIB_INTERNAL)
// C: void ZLIB_INTERNAL _tr_stored_block(deflate_state *s, charf *buf, ulg stored_len, int last)
func (s *DeflateState) _tr_stored_block(buf []byte, stored_len uint64, last int) {
	panic("TODO, stubbed")
}

// _tr_align (ZLIB_INTERNAL)
// C: void ZLIB_INTERNAL _tr_align(deflate_state *s)
func (s *DeflateState) _tr_align() {
	panic("TODO, stubbed")
}

// _tr_flush_block (ZLIB_INTERNAL)
// C: void ZLIB_INTERNAL _tr_flush_block(deflate_state *s, charf *buf, ulg stored_len, int last)
func (s *DeflateState) _tr_flush_block(buf []byte, stored_len uint64, last int) {
	panic("TODO, stubbed")
}

// _tr_tally (ZLIB_INTERNAL)
// C: int ZLIB_INTERNAL _tr_tally (deflate_state *s, unsigned dist, unsigned lc)
func (s *DeflateState) _tr_tally(dist uint32, lc uint32) int {
	panic("TODO, stubbed")
}

// compress_block (local)
// C: local void compress_block(deflate_state *s, ct_data *ltree, ct_data *dtree)
func (s *DeflateState) compress_block(ltree []CtData, dtree []CtData) {
	panic("TODO, stubbed")
}

// detect_data_type (local)
// C: local int detect_data_type(deflate_state *s)
func (s *DeflateState) detect_data_type() int {
	panic("TODO, stubbed")
}

// bi_reverse (local)
// C: local unsigned bi_reverse(unsigned code, int len)
func bi_reverse(code uint32, len_ int) uint32 {
	panic("TODO, stubbed")
}

// bi_flush (local)
// C: local void bi_flush(deflate_state *s)
func (s *DeflateState) bi_flush() {
	panic("TODO, stubbed")
}

// bi_windup (local)
// C: local void bi_windup(deflate_state *s)
func (s *DeflateState) bi_windup() {
	panic("TODO, stubbed")
}

// copy_block (local)
// C: local void copy_block(deflate_state *s, charf    *buf, unsigned len, int      header)
func (s *DeflateState) copy_block(buf []byte, len_ uint32, header int) {
	panic("TODO, stubbed")
}
