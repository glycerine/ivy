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
	lrn, mark, partA, lbd, size int32
	allocWords                  int32
	lits                        []lit
}
type Sat_Mem_t struct {
	nEntries, BookMarkH, BookMarkE, iPage [2]int32
	nPageSize                             int32
	uPageMask, uLearnedMask               int32
	pPages                                []clausePage
}

// Clause headers start on even word offsets. Allocate just the portion of
// each page needed for the current stream, keeping pointers stable on growth.
type clausePage struct {
	limit int32
	slots []*clause
}

func (p *clausePage) put(offset int32, c *clause) {
	i := int(offset >> 1)
	if i >= len(p.slots) {
		p.slots = slices.Grow(p.slots, i+1-len(p.slots))[:i+1]
	}
	p.slots[i] = c
}

func (p *clausePage) truncate(limit int32) {
	end := int(limit >> 1)
	if end < len(p.slots) {
		clear(p.slots[end:])
		p.slots = p.slots[:end]
	}
	p.limit = limit
}

func Sat_MemLimit(p []int32) int32               { return p[0] }
func Sat_MemIncLimit(p []int32, n int32) int32   { p[0] += n; return p[0] }
func Sat_MemWriteLimit(p []int32, n int32)       { p[0] = n }
func Sat_MemHandPage(p *Sat_Mem_t, h cla) int32  { return int32(uint32(h) >> p.nPageSize) }
func Sat_MemHandShift(p *Sat_Mem_t, h cla) int32 { return h & p.uPageMask }
func Sat_MemIntSize(size, lrn int32) int32       { return 2 * ((size + 2 + lrn) / 2) }
func Sat_MemClauseSize(c *clause) int32          { return Sat_MemIntSize(c.size, c.lrn) }
func Sat_MemClauseSize2(c *clause) int32         { return Sat_MemIntSize(c.size, 1) }
func Sat_MemClause(p *Sat_Mem_t, i, k int32) *clause {
	return Sat_MemClauseHand(p, Sat_MemHand(p, i, k))
}
func Sat_MemClauseHand(p *Sat_Mem_t, h cla) *clause {
	if h == 0 {
		return nil
	}
	c := p.pPages[Sat_MemHandPage(p, h)].slots[Sat_MemHandShift(p, h)>>1]
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
func clause_read_lit(h cla) lit   { return h >> 1 }
func clause_learnt_h(p *Sat_Mem_t, h cla) int32 {
	if h&p.uLearnedMask != 0 {
		return 1
	}
	return 0
}
func clause_learnt(c *clause) int32     { return c.lrn }
func clause_id(c *clause) int32         { return c.lits[c.size] }
func clause_set_id(c *clause, id int32) { c.lits[c.size] = id }
func clause_size(c *clause) int32       { return c.size }
func clause_begin(c *clause) []lit      { return c.lits[:c.size] }
func clause_end(c *clause) []lit        { return c.lits[c.size:c.size] }
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
	for l := int32(0); l < 2; l++ {
		Sat_MemShrink(p, p.BookMarkH[l], l)
	}
	p.nEntries = p.BookMarkE
}
func Sat_MemFree_(p *Sat_Mem_t) { *p = Sat_Mem_t{} }
func Sat_MemFree(p *Sat_Mem_t)  { Sat_MemFree_(p) }
func (p *Sat_Mem_t) ensurePage(page int32) {
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
	c := &clause{size: size, lrn: lrn, allocWords: words, lits: make([]lit, size+1)}
	copy(c.lits[:size], data)
	if lrn|plus != 0 {
		c.lits[size] = p.nEntries[lrn]
	}
	h := Sat_MemHand(p, page, offset)
	p.pPages[page].put(offset, c)
	p.nEntries[lrn]++
	p.pPages[page].limit += words
	return h
}
func Sat_MemShrink(p *Sat_Mem_t, h, lrn int32) {
	if h == 0 || clause_learnt_h(p, h) != lrn || uint32(h) > uint32(Sat_MemHandCurrent(p, lrn)) {
		panic("invalid clause shrink")
	}
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
		dest.put(at, c)
		c.allocWords = Sat_MemClauseSize(c)
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
