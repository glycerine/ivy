package espresso

// A set preserves Espresso's 32-bit packed layout: word zero contains loop,
// flags and optional cardinality metadata; data words start at index one.
type pset []uint32
type pcube = pset
type set_family_t struct {
	wsize, sf_size, capacity, count, active_count int32
	data                                          pset
	next                                          *set_family_t
}
type pset_family = *set_family_t
type pcover = pset_family

const (
	BPI      int32  = 32
	LOGBPI          = 5
	PRIME    uint32 = 0x8000
	NONESSEN uint32 = 0x4000
	ACTIVE   uint32 = 0x2000
	REDUND   uint32 = 0x1000
	COVERED  uint32 = 0x0800
	RELESSEN uint32 = 0x0400
)

func WHICH_WORD(element int32) int32 { return (element >> LOGBPI) + 1 }
func WHICH_BIT(element int32) int32  { return element & (BPI - 1) }
func LOOP(set pset) int32            { return int32(set[0] & 0x03ff) }
func LOOPCOPY(set pset) int32        { return LOOP(set) }
func PUTLOOP(set pset, n int32)      { set[0] = (set[0] &^ 0x03ff) | uint32(n) }
func SIZE(set pset) int32            { return int32(set[0] >> 16) }
func PUTSIZE(set pset, size int32)   { set[0] = (set[0] & 0xffff) | (uint32(size) << 16) }
func NELEM(set pset) int32           { return BPI * LOOP(set) }
func LOOPINIT(size int32) int32 {
	if size <= BPI {
		return 1
	}
	return WHICH_WORD(size - 1)
}
func SET_SIZE(size int32) int32 {
	if size < 0 || size > BPI*1023 {
		panic("espresso: set size outside packed metadata range")
	}
	return LOOPINIT(size) + 1
}
func SET(set pset, flag uint32)        { set[0] |= flag }
func RESET(set pset, flag uint32)      { set[0] &^= flag }
func TESTP(set pset, flag uint32) bool { return set[0]&flag != 0 }
func GETSET(family pset_family, index int32) pset {
	if index < 0 || index >= family.capacity {
		panic("espresso: set-family row out of bounds")
	}
	start := family.wsize * index
	return family.data[start : start+family.wsize : start+family.wsize]
}
func set_new(size int32) pset  { return set_clear(make(pset, SET_SIZE(size)), size) }
func set_full(size int32) pset { return set_fill(make(pset, SET_SIZE(size)), size) }
func set_save(set pset) pset   { return set_copy(make(pset, LOOPCOPY(set)+1), set) }
func set_free(pset)            {} // Sets are Go-owned slices; no manual allocator is retained.
func is_in_set(set pset, element int32) bool {
	return set[WHICH_WORD(element)]&(uint32(1)<<WHICH_BIT(element)) != 0
}
func set_remove(set pset, element int32) {
	set[WHICH_WORD(element)] &^= uint32(1) << WHICH_BIT(element)
}
func set_insert(set pset, element int32) { set[WHICH_WORD(element)] |= uint32(1) << WHICH_BIT(element) }

// Set and Family expose the same storage to package users without exporting
// the original C bookkeeping fields.
type Set = pset
type Family = set_family_t

func NewSet(size int32) Set                  { return set_new(size) }
func NewFamily(capacity, size int32) *Family { return sf_new(capacity, size) }
func (s pset) Contains(element int32) bool   { return is_in_set(s, element) }
func (s pset) Insert(element int32)          { set_insert(s, element) }
func (s pset) Remove(element int32)          { set_remove(s, element) }
func (s pset) Cardinality() int32            { return set_ord(s) }
func (f *set_family_t) Count() int32         { return f.count }
func (f *set_family_t) SetSize() int32       { return f.sf_size }
func (f *set_family_t) Row(index int32) Set {
	if index < 0 || index >= f.count {
		panic("espresso: family row out of bounds")
	}
	return GETSET(f, index)
}
func (f *set_family_t) Add(set Set) { sf_addset(f, set) }
