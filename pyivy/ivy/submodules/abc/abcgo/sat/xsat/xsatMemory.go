// Go-owned clause-handle storage replacing xsatMemory.h's raw-word arena.
package xsat

type xSAT_Mem_t struct{ clauses []*xSAT_Clause_t }

func xSAT_MemClauseHand(p *xSAT_Mem_t, ref uint32) *xSAT_Clause_t {
	if ref == CRefUndef {
		return nil
	}
	return p.clauses[ref]
}
func xSAT_MemGrow(p *xSAT_Mem_t, capacity uint32) {
	if uint64(capacity) > uint64(cap(p.clauses)) {
		a := make([]*xSAT_Clause_t, len(p.clauses), capacity)
		copy(a, p.clauses)
		p.clauses = a
	}
}
func xSAT_MemAlloc(capacity int) *xSAT_Mem_t {
	if capacity < 0 {
		panic("negative memory capacity")
	}
	return &xSAT_Mem_t{clauses: make([]*xSAT_Clause_t, 0, capacity)}
}
func xSAT_MemRestart(p *xSAT_Mem_t) { clear(p.clauses); p.clauses = p.clauses[:0] }
func xSAT_MemFree(p *xSAT_Mem_t)    { p.clauses = nil }
func xSAT_MemAppend(p *xSAT_Mem_t, words int) uint32 {
	if words < 2 {
		panic("clause allocation smaller than its header")
	}
	if uint64(len(p.clauses)) >= uint64(CRefUndef) {
		panic("clause handle space exhausted")
	}
	ref := uint32(len(p.clauses))
	p.clauses = append(p.clauses, &xSAT_Clause_t{lits: make([]int, 0, words-2)})
	return ref
}
func xSAT_MemCRef(p *xSAT_Mem_t, c *xSAT_Clause_t) uint32 {
	for i, other := range p.clauses {
		if other == c {
			return uint32(i)
		}
	}
	return CRefUndef
}
func xSAT_MemCap(p *xSAT_Mem_t) uint32 { return uint32(cap(p.clauses)) }
func xSAT_MemWastedCap(p *xSAT_Mem_t) uint32 {
	n := uint32(0)
	for _, c := range p.clauses {
		if c.marked {
			n++
		}
	}
	return n
}
