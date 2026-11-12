// Go-owned pool port of msatMem.c by Alan Mishchenko, UC Berkeley (2004).
package msat

type Msat_MmFixed_t struct {
	entrySize int
	free, all [][]byte
	live      map[*byte]bool
}

func Msat_MmFixedStart(size int) *Msat_MmFixed_t {
	if size <= 0 {
		panic("invalid pool entry size")
	}
	return &Msat_MmFixed_t{entrySize: size, live: map[*byte]bool{}}
}
func Msat_MmFixedStop(p *Msat_MmFixed_t, verbose int) { p.all = nil; p.free = nil; p.live = nil }
func Msat_MmFixedEntryFetch(p *Msat_MmFixed_t) []byte {
	var b []byte
	if len(p.free) > 0 {
		i := len(p.free) - 1
		b = p.free[i]
		p.free = p.free[:i]
	} else {
		b = make([]byte, p.entrySize)
		p.all = append(p.all, b)
	}
	p.live[&b[0]] = true
	return b
}
func Msat_MmFixedEntryRecycle(p *Msat_MmFixed_t, b []byte) {
	if len(b) != p.entrySize || !p.live[&b[0]] {
		panic("foreign or recycled pool entry")
	}
	delete(p.live, &b[0])
	p.free = append(p.free, b)
}
func Msat_MmFixedRestart(p *Msat_MmFixed_t)          { p.free = append(p.free[:0], p.all...); clear(p.live) }
func Msat_MmFixedReadMemUsage(p *Msat_MmFixed_t) int { return len(p.all) * p.entrySize }

type Msat_MmFlex_t struct {
	chunks [][]byte
	used   int
}

func Msat_MmFlexStart() *Msat_MmFlex_t              { return &Msat_MmFlex_t{} }
func Msat_MmFlexStop(p *Msat_MmFlex_t, verbose int) { p.chunks = nil; p.used = 0 }
func Msat_MmFlexEntryFetch(p *Msat_MmFlex_t, n int) []byte {
	if n < 0 {
		panic("negative allocation")
	}
	b := make([]byte, n)
	p.chunks = append(p.chunks, b)
	p.used += n
	return b
}
func Msat_MmFlexReadMemUsage(p *Msat_MmFlex_t) int { return p.used }

type Msat_MmStep_t struct {
	pools []*Msat_MmFixed_t
	large map[*byte][]byte
}

func Msat_MmStepStart(steps int) *Msat_MmStep_t {
	if steps < 0 || steps > 30 {
		panic("invalid pool steps")
	}
	p := &Msat_MmStep_t{large: map[*byte][]byte{}}
	for i := 0; i < steps; i++ {
		p.pools = append(p.pools, Msat_MmFixedStart(8<<i))
	}
	return p
}
func Msat_MmStepStop(p *Msat_MmStep_t, verbose int) {
	for _, pool := range p.pools {
		Msat_MmFixedStop(pool, 0)
	}
	p.pools = nil
	p.large = nil
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
	if n <= 0 {
		panic("invalid allocation size")
	}
	if pool := p.pool(n); pool != nil {
		return Msat_MmFixedEntryFetch(pool)[:n]
	}
	b := make([]byte, n)
	p.large[&b[0]] = b
	return b
}
func Msat_MmStepEntryRecycle(p *Msat_MmStep_t, b []byte, n int) {
	if n <= 0 || len(b) < n {
		panic("invalid recycle size")
	}
	if pool := p.pool(n); pool != nil {
		Msat_MmFixedEntryRecycle(pool, b[:pool.entrySize])
		return
	}
	if _, ok := p.large[&b[0]]; !ok {
		panic("foreign large allocation")
	}
	delete(p.large, &b[0])
}
func Msat_MmStepReadMemUsage(p *Msat_MmStep_t) int {
	n := 0
	for _, pool := range p.pools {
		n += Msat_MmFixedReadMemUsage(pool)
	}
	for _, b := range p.large {
		n += len(b)
	}
	return n
}
