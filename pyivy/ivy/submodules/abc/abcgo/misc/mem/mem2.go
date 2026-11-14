package mem

type Mmr_Flex_t struct {
	nPageBase, PageMask, nEntries, nEntriesMax, iNext int32
	pages                                             [][]uint64
}
type Mmr_Fixed_t struct {
	nPageBase, PageMask, nEntryWords, nEntries, nEntriesMax int32
	pages                                                   [][]uint64
	frees                                                   []int32
}
type Mmr_Step_t struct {
	nBits, uMask, nEntries, nEntriesMax, nEntriesAll int32
	mems                                             []Mmr_Fixed_t
}

func Mmr_FlexStart(base int32) *Mmr_Flex_t {
	if base < 2 || base > 28 {
		panic("invalid page size")
	}
	return &Mmr_Flex_t{nPageBase: base, PageMask: (1 << base) - 1, iNext: 1 << base}
}
func Mmr_FlexStop(p *Mmr_Flex_t) { *p = Mmr_Flex_t{} }
func Mmr_FlexEntry(p *Mmr_Flex_t, h int32) []uint64 {
	page, off := h>>p.nPageBase, h&p.PageMask
	if h <= 0 || int(page) >= len(p.pages) || p.pages[page] == nil {
		panic("invalid handle")
	}
	return p.pages[page][off:]
}
func Mmr_FlexFetch(p *Mmr_Flex_t, n int32) int32 {
	if n <= 0 || n >= p.PageMask {
		panic("invalid entry size")
	}
	if p.iNext+n >= p.PageMask {
		p.pages = append(p.pages, filledWords(p.PageMask+1))
		p.iNext = 1
	}
	h := int32(len(p.pages)-1)<<p.nPageBase | p.iNext
	p.iNext += n
	p.nEntries++
	if p.nEntries > p.nEntriesMax {
		p.nEntriesMax = p.nEntries
	}
	return h
}
func Mmr_FlexRelease(p *Mmr_Flex_t, h int32) {
	page := h >> p.nPageBase
	if h <= 0 || int(page) >= len(p.pages) {
		panic("invalid handle")
	}
	if page > 0 {
		p.pages[page-1] = nil
	}
}
func filledWords(n int32) []uint64 {
	b := make([]uint64, n)
	for i := range b {
		b[i] = ^uint64(0)
	}
	return b
}
func Mmr_FixedCreate(p *Mmr_Fixed_t, base, n int32) {
	if base < 1 || base > 28 || n <= 0 || n >= 1<<base {
		panic("invalid fixed arena")
	}
	*p = Mmr_Fixed_t{nPageBase: base, PageMask: (1 << base) - 1, nEntryWords: n}
}
func Mmr_FixedStart(base, n int32) *Mmr_Fixed_t {
	p := new(Mmr_Fixed_t)
	Mmr_FixedCreate(p, base, n)
	return p
}
func Mmr_FixedStop(p *Mmr_Fixed_t, freeLast int32) { *p = Mmr_Fixed_t{} }
func Mmr_FixedEntry(p *Mmr_Fixed_t, h int32) []uint64 {
	page, off := h>>p.nPageBase, h&p.PageMask
	if h <= 0 || int(page) >= len(p.pages) || off+p.nEntryWords > p.PageMask+1 {
		panic("invalid handle")
	}
	return p.pages[page][off : off+p.nEntryWords : off+p.nEntryWords]
}
func Mmr_FixedFetch(p *Mmr_Fixed_t) int32 {
	if len(p.frees) == 0 {
		base := int32(len(p.pages)) << p.nPageBase
		p.pages = append(p.pages, filledWords(p.PageMask+1))
		var handles []int32
		for i := int32(1); i+p.nEntryWords <= p.PageMask; i += p.nEntryWords {
			handles = append(handles, base|i)
		}
		// A maximum-size entry uses every word except reserved offset zero.
		if len(handles) == 0 {
			handles = append(handles, base|1)
		}
		for i := len(handles) - 1; i >= 0; i-- {
			p.frees = append(p.frees, handles[i])
		}
	}
	i := len(p.frees) - 1
	h := p.frees[i]
	p.frees = p.frees[:i]
	p.nEntries++
	if p.nEntries > p.nEntriesMax {
		p.nEntriesMax = p.nEntries
	}
	return h
}
func Mmr_FixedRecycle(p *Mmr_Fixed_t, h int32) {
	entry := Mmr_FixedEntry(p, h)
	for i := range entry {
		entry[i] = ^uint64(0)
	}
	p.nEntries--
	p.frees = append(p.frees, h)
}
func Mmr_FixedMemory(p *Mmr_Fixed_t) int32 { return int32(len(p.pages)) * (p.PageMask + 1) }
func Mmr_StepStart(base, bits int32) *Mmr_Step_t {
	if bits < 1 || bits > 16 {
		panic("invalid step bits")
	}
	p := &Mmr_Step_t{nBits: bits, uMask: (1 << bits) - 1, mems: make([]Mmr_Fixed_t, 1<<bits)}
	for i := int32(1); i <= p.uMask; i++ {
		Mmr_FixedCreate(&p.mems[i], base, i)
	}
	return p
}
func Mmr_StepStop(p *Mmr_Step_t) { *p = Mmr_Step_t{} }
func Mmr_StepEntry(p *Mmr_Step_t, h int32) []uint64 {
	n := h & p.uMask
	if n == 0 {
		panic("invalid handle")
	}
	return Mmr_FixedEntry(&p.mems[n], h>>p.nBits)
}
func Mmr_StepFetch(p *Mmr_Step_t, n int32) int32 {
	if n <= 0 || n > p.uMask {
		panic("invalid entry size")
	}
	p.nEntries++
	p.nEntriesAll++
	if p.nEntries > p.nEntriesMax {
		p.nEntriesMax = p.nEntries
	}
	return Mmr_FixedFetch(&p.mems[n])<<p.nBits | n
}
func Mmr_StepRecycle(p *Mmr_Step_t, h int32) {
	n := h & p.uMask
	if n == 0 {
		panic("invalid handle")
	}
	p.nEntries--
	Mmr_FixedRecycle(&p.mems[n], h>>p.nBits)
}

// Mmr_StepEntryNum returns the current number of live entries, matching
// Mmr_Step_t.nEntries.
func Mmr_StepEntryNum(p *Mmr_Step_t) int32 { return p.nEntries }

// Mmr_StepEntryNumMax returns the high-water live-entry count, matching
// Mmr_Step_t.nEntriesMax. It is not the number of entries ever fetched.
func Mmr_StepEntryNumMax(p *Mmr_Step_t) int32 { return p.nEntriesMax }

func Mmr_StepMemory(p *Mmr_Step_t) int32 {
	var n int32
	for i := int32(1); i <= p.uMask; i++ {
		n += Mmr_FixedMemory(&p.mems[i])
	}
	return n
}
