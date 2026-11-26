package bsat

import (
	"fmt"
	"iter"
	"slices"
)

/**CFile****************************************************************

  FileName    [satMem.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Memory management.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: satMem.h,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

// Clauses use typed storage; handles retain ABC's alternating-page encoding.
const (
	LEARNT_MAX_START_DEFAULT = 10000
	LEARNT_MAX_INCRE_DEFAULT = 1000
	LEARNT_MAX_RATIO_DEFAULT = 50
)

type clause struct {
	lits                  []lit
	allocWords, id        int32
	lrn, mark, partA, lbd uint8
}

// Clause headers are allocated in blocks and never overwritten while a caller
// may retain a clause pointer. Literals live in their page's arena.
type clauseBlock struct {
	headers    []clause
	nextHeader int
}

func (b *clauseBlock) take() *clause {
	if b.nextHeader == len(b.headers) {
		b.headers = make([]clause, min(128, max(1, 2*len(b.headers))))
		b.nextHeader = 0
	}
	c := &b.headers[b.nextHeader]
	b.nextHeader++
	return c
}

type Sat_Mem_t struct {
	nEntries, BookMarkH, BookMarkE, iPage [2]int32
	nPageSize                             int32
	uPageMask, uLearnedMask               int32
	pPages                                []clausePage
	headers                               [2]clauseBlock
}

// The native page offset counts literal words. Only blocks containing a clause
// header need an index. Most solvers use just a few entries in the first block.
const clauseIndexBlockSize = 256

// Like ABC, each page stores its clauses contiguously at their handle
// offsets: data[offset] holds the literal count and the literals follow.
// clause.lits aliases this arena, so propagation reaches a clause's literals
// from its handle without visiting the header. Growing the arena re-points
// the page's clauses at the new storage.
type clausePage struct {
	limit  int32
	blocks [][]*clause
	data   []lit
}

// reserve makes data hold at least end words, never more than pageWords.
func (p *clausePage) reserve(end int32, pageWords int64) {
	if int(end) <= len(p.data) {
		return
	}
	grown := make([]lit, min(pageWords, int64(max(int(end), 2*len(p.data), 64))))
	copy(grown, p.data)
	p.data = grown
	for blockID, block := range p.blocks {
		for slot, c := range block {
			if c != nil {
				p.attach(int32(2*(blockID*clauseIndexBlockSize+slot)), c, int32(len(c.lits)))
			}
		}
	}
}

// attach points c at size literals stored after the header word at offset.
func (p *clausePage) attach(offset int32, c *clause, size int32) {
	p.data[offset] = size
	c.lits = p.data[offset+1 : offset+1+size : offset+1+size]
}

func (p *clausePage) put(offset int32, c *clause) {
	index := int(offset >> 1)
	blockID, slot := index/clauseIndexBlockSize, index%clauseIndexBlockSize
	if blockID >= len(p.blocks) {
		p.blocks = slices.Grow(p.blocks, blockID+1-len(p.blocks))[:blockID+1]
	}
	block := p.blocks[blockID]
	if slot >= len(block) {
		entries := clauseIndexBlockSize
		if blockID == 0 {
			entries = min(entries, max(slot+1, 2*len(block)))
		}
		grown := make([]*clause, entries)
		copy(grown, block)
		p.blocks[blockID] = grown
		block = grown
	}
	block[slot] = c
}

func (p *clausePage) truncate(limit int32) {
	p.detach(limit)
	end := int(limit >> 1)
	blockID, slot := end/clauseIndexBlockSize, end%clauseIndexBlockSize
	if blockID < len(p.blocks) {
		if slot != 0 {
			if block := p.blocks[blockID]; slot < len(block) {
				clear(block[slot:])
			}
			blockID++
		}
		clear(p.blocks[blockID:])
		p.blocks = p.blocks[:blockID]
	}
	p.limit = limit
}

// detach moves the literals of clauses at or beyond limit out of the arena
// before it is reused, so retained clause pointers keep their contents.
// Dropped clauses are contiguous, so one copy serves them all.
func (p *clausePage) detach(limit int32) {
	if limit >= p.limit {
		return
	}
	saved := slices.Clone(p.data[limit:p.limit])
	for blockID := int(limit>>1) / clauseIndexBlockSize; blockID < len(p.blocks); blockID++ {
		for slot, c := range p.blocks[blockID] {
			offset := int32(2 * (blockID*clauseIndexBlockSize + slot))
			if c != nil && offset >= limit {
				start := offset + 1 - limit
				c.lits = saved[start : start+int32(len(c.lits)) : start+int32(len(c.lits))]
			}
		}
	}
}

func Sat_MemLimit(p []int32) int32               { return p[0] }
func Sat_MemIncLimit(p []int32, n int32) int32   { p[0] += n; return p[0] }
func Sat_MemWriteLimit(p []int32, n int32)       { p[0] = n }
func Sat_MemHandPage(p *Sat_Mem_t, h cla) int32  { return int32(uint32(h) >> p.nPageSize) }
func Sat_MemHandShift(p *Sat_Mem_t, h cla) int32 { return h & p.uPageMask }
func Sat_MemIntSize(size, lrn int32) int32 {
	words := (int64(size) + 2 + int64(lrn)) &^ 1
	if words > 1<<31-1 {
		panic("clause word count exceeds handle range")
	}
	return int32(words)
}
func Sat_MemClauseSize(c *clause) int32  { return Sat_MemIntSize(clause_size(c), int32(c.lrn)) }
func Sat_MemClauseSize2(c *clause) int32 { return Sat_MemIntSize(clause_size(c), 1) }
func Sat_MemClause(p *Sat_Mem_t, i, k int32) *clause {
	return Sat_MemClauseHand(p, Sat_MemHand(p, i, k))
}
func (p *Sat_Mem_t) clauseAt(h cla) *clause {
	index := Sat_MemHandShift(p, h) >> 1
	return p.pPages[Sat_MemHandPage(p, h)].blocks[index/clauseIndexBlockSize][index%clauseIndexBlockSize]
}

func Sat_MemClauseHand(p *Sat_Mem_t, h cla) *clause {
	if h == 0 {
		return nil
	}
	c := p.clauseAt(h)
	if h&1 != 0 || c == nil {
		panic("invalid clause handle")
	}
	return c
}
func Sat_MemEntryNum(p *Sat_Mem_t, lrn int32) int32 { return p.nEntries[lrn] }
func Sat_MemHand(p *Sat_Mem_t, i, k int32) cla      { return i<<p.nPageSize | k }
func Sat_MemHandCurrent(p *Sat_Mem_t, lrn int32) cla {
	return Sat_MemHand(p, p.iPage[lrn], p.pPages[p.iPage[lrn]].limit)
}
func Sat_MemClauseUsed(p *Sat_Mem_t, h cla) int32 {
	lrn := clause_learnt_h(p, h)
	if uint32(h) < uint32(p.BookMarkH[lrn]) {
		return 1
	}
	return 0
}
func Sat_MemMemoryHand(p *Sat_Mem_t, h cla) float64 {
	return float64((Sat_MemHandPage(p, h)+2)/2)*float64(uint64(1)<<uint(p.nPageSize+2)) + float64(Sat_MemHandShift(p, h))*4
}
func Sat_MemMemoryUsed(p *Sat_Mem_t, lrn int32) float64 {
	return Sat_MemMemoryHand(p, Sat_MemHandCurrent(p, lrn))
}
func Sat_MemMemoryAllUsed(p *Sat_Mem_t) float64 {
	return Sat_MemMemoryUsed(p, 0) + Sat_MemMemoryUsed(p, 1)
}
func Sat_MemMemoryAll(p *Sat_Mem_t) float64 {
	return float64(p.iPage[0]+p.iPage[1]+2) * float64(uint64(1)<<uint(p.nPageSize+2))
}
func clause_from_lit(l lit) int32 { return l + l + 1 }
func clause_is_lit(h cla) int32   { return h & 1 }
func clause_read_lit(h cla) lit   { return lit(uint32(h) >> 1) }
func clause_learnt_h(p *Sat_Mem_t, h cla) int32 {
	if h&p.uLearnedMask != 0 {
		return 1
	}
	return 0
}
func clause_learnt(c *clause) int32     { return int32(c.lrn) }
func clause_id(c *clause) int32         { return c.id }
func clause_set_id(c *clause, id int32) { c.id = id }
func clause_size(c *clause) int32       { return int32(len(c.lits)) }
func clause_begin(c *clause) []lit      { return c.lits }
func clause_end(c *clause) []lit        { return c.lits[len(c.lits):] }
func clause_print_(c *clause) {
	fmt.Print("{ ")
	for _, l := range clause_begin(c) {
		n := l >> 1
		if l&1 != 0 {
			n = -n
		}
		fmt.Printf("%d ", n)
	}
	fmt.Println("}")
}
func Sat_MemCountL(p *Sat_Mem_t) int32 {
	var count int32
	for range p.clauses(1) {
		count++
	}
	return count
}
func Sat_MemAlloc_(p *Sat_Mem_t, n int32) {
	if n <= 8 || n >= 32 {
		panic("invalid SAT page size")
	}
	*p = Sat_Mem_t{nPageSize: n, uPageMask: int32((uint64(1) << uint(n)) - 1), uLearnedMask: int32(uint64(1) << uint(n)), pPages: make([]clausePage, 2)}
	p.iPage = [2]int32{0, 1}
	p.pPages[0].limit = 2
	p.pPages[1].limit = 2
}
func Sat_MemAlloc(n int32) *Sat_Mem_t { p := new(Sat_Mem_t); Sat_MemAlloc_(p, n); return p }
func Sat_MemRestart(p *Sat_Mem_t) {
	p.headers = [2]clauseBlock{}
	p.nEntries = [2]int32{}
	p.iPage = [2]int32{0, 1}
	for i := range p.pPages {
		p.pPages[i].truncate(2)
	}
}
func Sat_MemBookMark(p *Sat_Mem_t) {
	p.BookMarkE = p.nEntries
	for l := int32(0); l < 2; l++ {
		p.BookMarkH[l] = Sat_MemHandCurrent(p, l)
	}
}
func Sat_MemRollBack(p *Sat_Mem_t) {
	for stream, bookmark := range p.BookMarkH {
		p.checkShrink(bookmark, int32(stream))
	}
	for stream, bookmark := range p.BookMarkH {
		p.shrink(bookmark, int32(stream))
	}
	p.nEntries = p.BookMarkE
}
func Sat_MemFree_(p *Sat_Mem_t) { *p = Sat_Mem_t{} }
func Sat_MemFree(p *Sat_Mem_t)  { Sat_MemFree_(p) }
func (p *Sat_Mem_t) ensurePage(page int32) {
	if uint32(page) > ^uint32(0)>>uint(p.nPageSize) {
		panic("SAT clause page exceeds handle range")
	}
	if int(page) >= len(p.pPages) {
		p.pPages = slices.Grow(p.pPages, int(page)+1-len(p.pPages))[:page+1]
	}
}
func Sat_MemAppend(p *Sat_Mem_t, data []int32, size, lrn, plus int32) int32 {
	if size < 0 || lrn < 0 || lrn > 1 || plus < 0 || plus > 1 || data != nil && int32(len(data)) < size {
		panic("invalid clause append")
	}
	words := Sat_MemIntSize(size, lrn|plus)
	limit := int64(1) << uint(p.nPageSize)
	if int64(words)+3 >= limit {
		panic("clause exceeds page")
	}
	page := p.iPage[lrn]
	offset := p.pPages[page].limit
	if int64(offset)+int64(words)+2 >= limit {
		page += 2
		p.ensurePage(page)
		p.pPages[page].limit = 2
		p.iPage[lrn] = page
		offset = 2
	}
	pg := &p.pPages[page]
	pg.reserve(offset+words, limit)
	c := p.headers[lrn].take()
	pg.attach(offset, c, size)
	if data == nil {
		clear(c.lits)
	}
	c.lrn, c.allocWords = uint8(lrn), words
	copy(c.lits, data)
	if lrn|plus != 0 {
		c.id = p.nEntries[lrn]
	}
	h := Sat_MemHand(p, page, offset)
	p.pPages[page].put(offset, c)
	p.nEntries[lrn]++
	p.pPages[page].limit += words
	return h
}
func Sat_MemShrink(p *Sat_Mem_t, h, lrn int32) {
	p.checkShrink(h, lrn)
	p.shrink(h, lrn)
}

func (p *Sat_Mem_t) checkShrink(h cla, stream int32) {
	if h == 0 || h&1 != 0 || clause_learnt_h(p, h) != stream || uint32(h) > uint32(Sat_MemHandCurrent(p, stream)) {
		panic("invalid clause shrink")
	}
	page := &p.pPages[Sat_MemHandPage(p, h)]
	offset := Sat_MemHandShift(p, h)
	if offset == page.limit {
		return
	}
	index := int(offset >> 1)
	blockID, slot := index/clauseIndexBlockSize, index%clauseIndexBlockSize
	if offset < 2 || blockID >= len(page.blocks) || slot >= len(page.blocks[blockID]) || page.blocks[blockID][slot] == nil {
		panic("clause shrink is not at a header boundary")
	}
}

func (p *Sat_Mem_t) shrink(h cla, lrn int32) {
	page := Sat_MemHandPage(p, h)
	for i := page + 2; i <= p.iPage[lrn]; i += 2 {
		p.pPages[i].truncate(2)
	}
	p.iPage[lrn] = page
	p.pPages[page].truncate(Sat_MemHandShift(p, h))
}
func (p *Sat_Mem_t) clauses(lrn int32) iter.Seq2[cla, *clause] {
	return func(yield func(cla, *clause) bool) {
		for page := lrn; page <= p.iPage[lrn]; page += 2 {
			for offset := int32(2); offset < p.pPages[page].limit; {
				h := Sat_MemHand(p, page, offset)
				c := Sat_MemClauseHand(p, h)
				if !yield(h, c) {
					return
				}
				offset += c.allocWords
			}
		}
	}
}

func (p *Sat_Mem_t) handles(lrn int32) []cla {
	var handles []cla
	for h := range p.clauses(lrn) {
		handles = append(handles, h)
	}
	return handles
}

func Sat_MemCompactLearned(p *Sat_Mem_t, move int32) int32 {
	page, offset, count := int32(1), int32(2), int32(0)
	limit := int64(1) << uint(p.nPageSize)
	bookmark := p.BookMarkH[1]
	var bookmarkNew, bookmarkCount int32
	var survivors []*clause
	if move != 0 {
		survivors = make([]*clause, 0, p.nEntries[1])
	}
	for h, c := range p.clauses(1) {
		if c.mark != 0 {
			continue
		}
		words := Sat_MemClauseSize(c)
		if int64(offset)+int64(words) >= limit {
			page += 2
			offset = 2
		}
		newH := Sat_MemHand(p, page, offset)
		if bookmark != 0 && bookmarkNew == 0 && uint32(h) >= uint32(bookmark) {
			bookmarkNew, bookmarkCount = newH, count
		}
		if move == 0 {
			clause_set_id(c, newH)
		} else {
			if clause_id(c) != newH {
				panic("clause compaction requires dry run")
			}
			survivors = append(survivors, c)
		}
		offset += words
		count++
	}
	if move == 0 {
		return count
	}
	// Every relocation is valid. Keep the typed clauses and rebuild only
	// the learned-page index, so pointers held by clients remain stable.
	for i := int32(1); i <= p.iPage[1]; i += 2 {
		p.pPages[i].truncate(2)
	}
	for id, c := range survivors {
		h := clause_id(c)
		dest := &p.pPages[Sat_MemHandPage(p, h)]
		at := Sat_MemHandShift(p, h)
		c.allocWords = Sat_MemClauseSize(c)
		// Destinations never pass a later survivor's source, and an
		// overlapping move within one array is a memmove.
		dest.reserve(at+c.allocWords, limit)
		size := int32(len(c.lits))
		copy(dest.data[at+1:at+1+size], c.lits)
		dest.attach(at, c, size)
		dest.put(at, c)
		dest.limit = at + c.allocWords
		clause_set_id(c, int32(id))
	}
	p.iPage[1], p.nEntries[1] = page, count
	if bookmark != 0 {
		if bookmarkNew == 0 {
			bookmarkNew, bookmarkCount = Sat_MemHand(p, page, offset), count
		}
		p.BookMarkH[1], p.BookMarkE[1] = bookmarkNew, bookmarkCount
	}
	return count
}
