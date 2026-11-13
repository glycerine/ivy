// Go-owned chunk allocators following msatMem.c.
package msat

import "github.com/glycerine/abcgo/misc/mem"

type Msat_MmFixed_t struct {
	entrySize int
	arena     *mem.Mem_Fixed_t
}

func Msat_MmFixedStart(size int) *Msat_MmFixed_t {
	return &Msat_MmFixed_t{entrySize: size, arena: mem.Mem_FixedStart(int32(size))}
}
func Msat_MmFixedStop(p *Msat_MmFixed_t, verbose int) {
	if p != nil {
		mem.Mem_FixedStop(p.arena, int32(verbose))
	}
}
func Msat_MmFixedEntryFetch(p *Msat_MmFixed_t) []byte      { return mem.Mem_FixedEntryFetch(p.arena) }
func Msat_MmFixedEntryRecycle(p *Msat_MmFixed_t, b []byte) { mem.Mem_FixedEntryRecycle(p.arena, b) }
func Msat_MmFixedRestart(p *Msat_MmFixed_t) {
	assert(mem.Mem_FixedReadMemUsage(p.arena) != 0, "fixed arena has first chunk")
	mem.Mem_FixedRestart(p.arena)
}
func Msat_MmFixedReadMemUsage(p *Msat_MmFixed_t) int { return int(mem.Mem_FixedReadMemUsage(p.arena)) }

type Msat_MmFlex_t struct {
	chunks                       [][]byte
	chunkSize, offset, allocated int
}

func Msat_MmFlexStart() *Msat_MmFlex_t { return &Msat_MmFlex_t{chunkSize: 1 << 12} }
func Msat_MmFlexStop(p *Msat_MmFlex_t, verbose int) {
	if p != nil {
		p.chunks = nil
		p.allocated = 0
	}
}
func Msat_MmFlexEntryFetch(p *Msat_MmFlex_t, n int) []byte {
	assert(n >= 0, "nonnegative allocation")
	if len(p.chunks) == 0 || p.offset+n > p.chunkSize {
		if n > p.chunkSize {
			p.chunkSize = 2 * n
		}
		p.chunks = append(p.chunks, make([]byte, p.chunkSize))
		p.allocated += p.chunkSize
		p.offset = 0
	}
	chunk := p.chunks[len(p.chunks)-1]
	start := p.offset
	p.offset += n
	return chunk[start:p.offset:p.offset]
}
func Msat_MmFlexReadMemUsage(p *Msat_MmFlex_t) int { return p.allocated }

type Msat_MmStep_t struct{ pools []*Msat_MmFixed_t }

func Msat_MmStepStart(steps int) *Msat_MmStep_t {
	assert(steps > 0 && steps <= 30, "valid pool steps")
	p := &Msat_MmStep_t{}
	for i := 0; i < steps; i++ {
		p.pools = append(p.pools, Msat_MmFixedStart(8<<i))
	}
	return p
}
func Msat_MmStepStop(p *Msat_MmStep_t, verbose int) {
	if p == nil {
		return
	}
	for _, pool := range p.pools {
		Msat_MmFixedStop(pool, 0)
	}
	p.pools = nil
}
func (p *Msat_MmStep_t) pool(n int) *Msat_MmFixed_t {
	for _, pool := range p.pools {
		if pool.entrySize >= n {
			return pool
		}
	}
	return nil
}
func Msat_MmStepEntryFetch(p *Msat_MmStep_t, n int) []byte {
	if n == 0 {
		return nil
	}
	assert(n > 0, "nonnegative allocation")
	if pool := p.pool(n); pool != nil {
		return Msat_MmFixedEntryFetch(pool)[:n]
	}
	return make([]byte, n)
}
func Msat_MmStepEntryRecycle(p *Msat_MmStep_t, b []byte, n int) {
	if n == 0 {
		return
	}
	assert(n > 0, "nonnegative allocation")
	if pool := p.pool(n); pool != nil {
		Msat_MmFixedEntryRecycle(pool, b[:pool.entrySize])
	}
}
func Msat_MmStepReadMemUsage(p *Msat_MmStep_t) int {
	n := 0
	for _, pool := range p.pools {
		n += Msat_MmFixedReadMemUsage(pool)
	}
	return n
}
