package bsat

import "fmt"

/**CFile****************************************************************

  FileName    [satTruth.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Truth table computation package.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satTruth.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

// Even handles identify canonical tables with bit zero clear. The low bit is
// a complement tag; handles zero and one are the constant functions.
type Tru_Man_t_ struct {
	nVars, nWords, nEntrySize, nTableSize int32
	pTable                                []int32
	entries                               []*Tru_One_t
	pZero                                 []word
	hIthVars                              [16]int32
	nTableLookups                         int32
}
type Tru_One_t struct {
	Handle, Next int32
	pTruth       []word
}

func Tru_ManReadOne(p *Tru_Man_t, h int32) *Tru_One_t {
	if h == 0 {
		return nil
	}
	if h < 0 || h&1 != 0 || int(h/2) >= len(p.entries) {
		panic("invalid truth handle")
	}
	return p.entries[h/2]
}

var truthPrimes = []int32{1291, 1699, 2357, 4177, 5147, 5647, 6343, 7103, 7873, 8147}

func Tru_ManHash(truth []word, n, bins int32, primes []int32) uint32 {
	if bins <= 0 || len(primes) < 8 {
		panic("invalid truth hash dimensions")
	}
	var hash uint32
	for i, v := range truth[:n] {
		hash ^= uint32(v * uint64(uint32(primes[i&7])))
	}
	return hash % uint32(bins)
}
func Tru_ManLookup(p *Tru_Man_t, truth []word) *int32 {
	if truth[0]&1 != 0 {
		panic("noncanonical truth lookup")
	}
	slot := &p.pTable[Tru_ManHash(truth, p.nWords, p.nTableSize, truthPrimes)]
	for *slot != 0 {
		entry := Tru_ManReadOne(p, *slot)
		if Tru_ManEqual(entry.pTruth, truth, p.nWords) != 0 {
			return slot
		}
		slot = &entry.Next
	}
	return slot
}
func Tru_ManResize(p *Tru_Man_t) {
	old := p.pTable
	p.nTableSize = 2*p.nTableSize + 1
	p.pTable = make([]int32, p.nTableSize)
	count := 0
	for _, head := range old {
		for head != 0 {
			entry := Tru_ManReadOne(p, head)
			next := entry.Next
			entry.Next = 0
			slot := Tru_ManLookup(p, entry.pTruth)
			if *slot != 0 {
				panic("duplicate truth during rehash")
			}
			*slot = entry.Handle
			count++
			head = next
		}
	}
	if count != len(p.entries)-1 {
		panic("missing truth during rehash")
	}
}
func Tru_ManInsert(p *Tru_Man_t, truth []word) int32 {
	if Tru_ManEqual0(truth, p.nWords) != 0 {
		return 0
	}
	if Tru_ManEqual1(truth, p.nWords) != 0 {
		return 1
	}
	p.nTableLookups++
	if len(p.entries)-1 > 2*int(p.nTableSize) {
		Tru_ManResize(p)
	}
	canonical := append([]word(nil), truth[:p.nWords]...)
	complement := int32(canonical[0] & 1)
	if complement != 0 {
		Tru_ManNot(canonical, p.nWords)
	}
	slot := Tru_ManLookup(p, canonical)
	if *slot == 0 {
		if len(p.entries) > 0x3fffffff {
			panic("truth handle overflow")
		}
		h := int32(len(p.entries)) * 2
		entry := &Tru_One_t{Handle: h, pTruth: canonical}
		p.entries = append(p.entries, entry)
		*slot = h
	}
	return *slot ^ complement
}
func Tru_ManAlloc(vars int32) *Tru_Man_t {
	if vars <= 0 || vars > 16 {
		panic("truth variable count must be 1..16")
	}
	words := int32(1)
	if vars >= 6 {
		words = 1 << uint(vars-6)
	}
	p := &Tru_Man_t{nVars: vars, nWords: words, nEntrySize: 2 + 2*words, nTableSize: 8147, pTable: make([]int32, 8147), entries: []*Tru_One_t{nil}, pZero: make([]word, words)}
	masks := []word{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	for v := int32(0); v < vars; v++ {
		for w := int32(0); w < words; w++ {
			value := word(0)
			if v < 6 {
				value = masks[v]
			} else if w&(1<<uint(v-6)) != 0 {
				value = ^word(0)
			}
			p.pZero[w] = value
		}
		p.hIthVars[v] = Tru_ManInsert(p, p.pZero)
	}
	clear(p.pZero)
	return p
}
func Tru_ManFree(p *Tru_Man_t) {
	if p == nil {
		return
	}
	fmt.Printf("Lookups = %d. Entries = %d.\n", p.nTableLookups, len(p.entries)-1)
	*p = Tru_Man_t{}
}

// Returned table slices are read-only views; use Copy before truth operations.
func Tru_ManVar(p *Tru_Man_t, v int32) []word {
	if v < 0 || v >= p.nVars {
		panic("truth variable outside manager")
	}
	return Tru_ManReadOne(p, p.hIthVars[v]).pTruth
}
func Tru_ManFunc(p *Tru_Man_t, h int32) []word {
	if h&1 != 0 {
		panic("strip complement tag before reading truth")
	}
	if h == 0 {
		return p.pZero
	}
	return Tru_ManReadOne(p, h).pTruth
}
