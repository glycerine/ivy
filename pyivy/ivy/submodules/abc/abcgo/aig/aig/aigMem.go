package aig

import "fmt"

// Go slices replace the C byte-pointer arenas. Pools retain entry ownership so
// recycling foreign storage or recycling twice is detected rather than corrupting
// an intrusive free list. Reused entries retain their bytes, as in ABC.
type Aig_MmFixed_t struct {
	size               int
	free               [][]byte
	live               map[*byte][]byte
	allocated, maximum int
}
type Aig_MmFixed_t_ = Aig_MmFixed_t
type Aig_MmFlex_t struct {
	entries [][]byte
	bytes   int
}
type Aig_MmFlex_t_ = Aig_MmFlex_t
type Aig_MmStep_t struct {
	pools   []*Aig_MmFixed_t
	large   map[*byte][]byte
	retired map[*byte]bool
}
type Aig_MmStep_t_ = Aig_MmStep_t

func Aig_MmFixedStart(size, maximum int32) *Aig_MmFixed_t {
	if size <= 0 || maximum < 0 {
		panic("aig: invalid fixed pool dimensions")
	}
	return &Aig_MmFixed_t{size: int(size), live: make(map[*byte][]byte)}
}
func Aig_MmFixedStop(p *Aig_MmFixed_t, verbose int32) {
	if p == nil {
		return
	}
	if verbose != 0 {
		fmt.Printf("Fixed pool: %d bytes, peak %d entries\n", Aig_MmFixedReadMemUsage(p), p.maximum)
	}
	*p = Aig_MmFixed_t{}
}
func Aig_MmFixedEntryFetch(p *Aig_MmFixed_t) []byte {
	if p.size <= 0 {
		panic("aig: stopped fixed pool")
	}
	var entry []byte
	if n := len(p.free); n > 0 {
		entry = p.free[n-1]
		p.free = p.free[:n-1]
	} else {
		entry = make([]byte, p.size)
		p.allocated++
	}
	p.live[&entry[0]] = entry
	if len(p.live) > p.maximum {
		p.maximum = len(p.live)
	}
	return entry
}
func Aig_MmFixedEntryRecycle(p *Aig_MmFixed_t, entry []byte) {
	if len(entry) != p.size || len(entry) == 0 {
		panic("aig: incorrect fixed entry size")
	}
	original, ok := p.live[&entry[0]]
	if !ok {
		panic("aig: foreign or recycled fixed entry")
	}
	delete(p.live, &entry[0])
	p.free = append(p.free, original)
}
func Aig_MmFixedRestart(p *Aig_MmFixed_t) {
	for _, entry := range p.live {
		p.free = append(p.free, entry)
	}
	clear(p.live)
}
func Aig_MmFixedReadMemUsage(p *Aig_MmFixed_t) int32       { return int32(p.allocated * p.size) }
func Aig_MmFixedReadMaxEntriesUsed(p *Aig_MmFixed_t) int32 { return int32(p.maximum) }
func Aig_MmFlexStart() *Aig_MmFlex_t                       { return &Aig_MmFlex_t{} }
func Aig_MmFlexStop(p *Aig_MmFlex_t, verbose int32) {
	if p == nil {
		return
	}
	if verbose != 0 {
		fmt.Printf("Flexible pool: %d bytes\n", p.bytes)
	}
	*p = Aig_MmFlex_t{}
}
func Aig_MmFlexEntryFetch(p *Aig_MmFlex_t, size int32) []byte {
	if size < 0 {
		panic("aig: negative flexible allocation")
	}
	entry := make([]byte, size)
	p.entries = append(p.entries, entry)
	p.bytes += int(size)
	return entry
}
func Aig_MmFlexRestart(p *Aig_MmFlex_t)            { p.entries = nil; p.bytes = 0 }
func Aig_MmFlexReadMemUsage(p *Aig_MmFlex_t) int32 { return int32(p.bytes) }
func Aig_MmStepStart(steps int32) *Aig_MmStep_t {
	if steps < 0 || steps > 30 {
		panic("aig: invalid stepped pool dimensions")
	}
	p := &Aig_MmStep_t{large: make(map[*byte][]byte), retired: make(map[*byte]bool)}
	for i := int32(0); i < steps; i++ {
		p.pools = append(p.pools, Aig_MmFixedStart(8<<uint(i), 0))
	}
	return p
}
func Aig_MmStepStop(p *Aig_MmStep_t, verbose int32) {
	if p == nil {
		return
	}
	if verbose != 0 {
		fmt.Printf("Stepped pool: %d bytes\n", Aig_MmStepReadMemUsage(p))
	}
	for _, pool := range p.pools {
		Aig_MmFixedStop(pool, 0)
	}
	*p = Aig_MmStep_t{}
}
func Aig_MmStepEntryFetch(p *Aig_MmStep_t, size int32) []byte {
	if size == 0 {
		return nil
	}
	if size < 0 {
		panic("aig: invalid stepped allocation")
	}
	for _, pool := range p.pools {
		if int(size) <= pool.size {
			return Aig_MmFixedEntryFetch(pool)[:size]
		}
	}
	entry := make([]byte, size)
	p.large[&entry[0]] = entry
	return entry
}
func Aig_MmStepEntryRecycle(p *Aig_MmStep_t, entry []byte, size int32) {
	if size == 0 {
		return
	}
	if size < 0 || len(entry) != int(size) {
		panic("aig: incorrect stepped entry size")
	}
	for _, pool := range p.pools {
		if int(size) <= pool.size {
			if cap(entry) < pool.size {
				panic("aig: foreign stepped entry")
			}
			Aig_MmFixedEntryRecycle(pool, entry[:pool.size])
			return
		}
	}
	original, ok := p.large[&entry[0]]
	if !ok || p.retired[&entry[0]] || len(original) != int(size) {
		panic("aig: foreign or recycled stepped entry")
	}
	p.retired[&entry[0]] = true
}
func Aig_MmStepReadMemUsage(p *Aig_MmStep_t) int32 {
	var result int32
	for _, pool := range p.pools {
		result += Aig_MmFixedReadMemUsage(pool)
	}
	for _, entry := range p.large {
		result += int32(len(entry))
	}
	return result
}
