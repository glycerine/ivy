// Go port of the indexed activity heap in msatOrderH.c (the configured ABC order).
package msat

type Msat_Order_t struct {
	pSat        *Msat_Solver_t
	heap, index []int
}

func Msat_OrderAlloc(p *Msat_Solver_t) *Msat_Order_t { return &Msat_Order_t{pSat: p} }
func Msat_OrderSetBounds(p *Msat_Order_t, n int) {
	p.index = make([]int, n)
	for i := range p.index {
		p.index[i] = -1
	}
	p.heap = p.heap[:0]
}
func (p *Msat_Order_t) better(a, b int) bool { return p.pSat.pdActivity[a] > p.pSat.pdActivity[b] }
func (p *Msat_Order_t) up(i int) {
	x := p.heap[i]
	for i > 0 {
		parent := (i - 1) / 2
		if !p.better(x, p.heap[parent]) {
			break
		}
		p.heap[i] = p.heap[parent]
		p.index[p.heap[i]] = i
		i = parent
	}
	p.heap[i] = x
	p.index[x] = i
}
func (p *Msat_Order_t) down(i int) {
	x := p.heap[i]
	for 2*i+1 < len(p.heap) {
		j := 2*i + 1
		if j+1 < len(p.heap) && p.better(p.heap[j+1], p.heap[j]) {
			j++
		}
		if !p.better(p.heap[j], x) {
			break
		}
		p.heap[i] = p.heap[j]
		p.index[p.heap[i]] = i
		i = j
	}
	p.heap[i] = x
	p.index[x] = i
}
func (p *Msat_Order_t) insert(v int) {
	if p.index[v] >= 0 {
		return
	}
	p.index[v] = len(p.heap)
	p.heap = append(p.heap, v)
	p.up(len(p.heap) - 1)
}
func (p *Msat_Order_t) pop() int {
	v := p.heap[0]
	last := p.heap[len(p.heap)-1]
	p.heap = p.heap[:len(p.heap)-1]
	p.index[v] = -1
	if len(p.heap) > 0 {
		p.heap[0] = last
		p.index[last] = 0
		p.down(0)
	}
	return v
}
func Msat_OrderClean(p *Msat_Order_t, cone *Msat_IntVec_t) {
	Msat_OrderSetBounds(p, len(p.pSat.pAssigns))
	for _, v := range cone.data {
		p.index[v] = len(p.heap)
		p.heap = append(p.heap, v)
	}
}
func Msat_OrderFree(p *Msat_Order_t) { p.heap = nil; p.index = nil; p.pSat = nil }
func Msat_OrderVarSelect(p *Msat_Order_t) int {
	for len(p.heap) > 0 {
		v := p.pop()
		if p.pSat.pAssigns[v] < 0 {
			return v
		}
	}
	return MSAT_ORDER_UNKNOWN
}
func Msat_OrderVarAssigned(p *Msat_Order_t, v int) { /* assigned entries are skipped lazily by selection */
}
func Msat_OrderVarUnassigned(p *Msat_Order_t, v int) {
	p.insert(v)
}
func Msat_OrderUpdate(p *Msat_Order_t, v int) {
	if v < len(p.index) && p.index[v] >= 0 {
		p.up(p.index[v])
	}
}
func Msat_OrderCheck(p *Msat_Order_t) int {
	for i, v := range p.heap {
		if p.index[v] != i {
			return 0
		}
		if i > 0 && p.better(v, p.heap[(i-1)/2]) {
			return 0
		}
	}
	return 1
}
