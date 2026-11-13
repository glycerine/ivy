// Go port of msatQueue.c by Alan Mishchenko, UC Berkeley (2004).
package msat

type Msat_Queue_t struct {
	data []int
	head int
}

func Msat_QueueAlloc(n int) *Msat_Queue_t {
	if n < 0 {
		panic("negative queue capacity")
	}
	return &Msat_Queue_t{data: make([]int, 0, n)}
}
func Msat_QueueFree(p *Msat_Queue_t)         { p.data = nil; p.head = 0 }
func Msat_QueueReadSize(p *Msat_Queue_t) int { return len(p.data) - p.head }
func Msat_QueueInsert(p *Msat_Queue_t, lit int) {
	assert(len(p.data) < cap(p.data), "queue has space")
	p.data = append(p.data, lit)
}
func Msat_QueueExtract(p *Msat_Queue_t) int {
	if p.head == len(p.data) {
		return -1
	}
	lit := p.data[p.head]
	p.head++
	return lit
}
func Msat_QueueClear(p *Msat_Queue_t) { p.data = p.data[:0]; p.head = 0 }
