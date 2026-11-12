// The standalone variable-heap API declared in msat.h, using the indexed heap
// algorithm in msatOrderH.c. Activity storage remains caller-owned.
package msat

import (
	"fmt"
	"io"
)

type Msat_VarHeap_t struct{ order *Msat_Order_t }

func Msat_VarHeapAlloc() *Msat_VarHeap_t {
	return &Msat_VarHeap_t{order: &Msat_Order_t{pSat: &Msat_Solver_t{}}}
}
func Msat_VarHeapSetActivity(p *Msat_VarHeap_t, activity []float64) {
	p.order.pSat.pdActivity = activity
}
func Msat_VarHeapGrow(p *Msat_VarHeap_t, n int) {
	if n < 0 {
		panic("negative heap size")
	}
	for len(p.order.index) < n {
		p.order.index = append(p.order.index, -1)
	}
}
func Msat_VarHeapStart(p *Msat_VarHeap_t, vars []int, n, allocated int) {
	if n < 0 || n > len(vars) || allocated < 0 {
		panic("invalid heap bounds")
	}
	p.order.heap = p.order.heap[:0]
	p.order.index = p.order.index[:0]
	Msat_VarHeapGrow(p, allocated)
	for _, v := range vars[:n] {
		Msat_VarHeapInsert(p, v)
	}
}
func Msat_VarHeapStop(p *Msat_VarHeap_t) {
	p.order.heap = nil
	p.order.index = nil
	p.order.pSat.pdActivity = nil
}
func Msat_VarHeapContainsVar(p *Msat_VarHeap_t, v int) int {
	return boolInt(v >= 0 && v < len(p.order.index) && p.order.index[v] >= 0)
}
func Msat_VarHeapInsert(p *Msat_VarHeap_t, v int) {
	if v < 0 || v >= len(p.order.pSat.pdActivity) {
		panic("heap activity index out of range")
	}
	Msat_VarHeapGrow(p, v+1)
	if Msat_VarHeapContainsVar(p, v) != 0 {
		panic("duplicate heap entry")
	}
	p.order.insert(v)
}
func Msat_VarHeapUpdate(p *Msat_VarHeap_t, v int) {
	if Msat_VarHeapContainsVar(p, v) == 0 {
		panic("absent heap entry")
	}
	p.order.up(p.order.index[v])
	p.order.down(p.order.index[v])
}
func Msat_VarHeapDelete(p *Msat_VarHeap_t, v int) {
	if Msat_VarHeapContainsVar(p, v) == 0 {
		panic("absent heap entry")
	}
	i := p.order.index[v]
	last := p.order.heap[len(p.order.heap)-1]
	p.order.heap = p.order.heap[:len(p.order.heap)-1]
	p.order.index[v] = -1
	if i < len(p.order.heap) {
		p.order.heap[i] = last
		p.order.index[last] = i
		p.order.up(i)
		p.order.down(p.order.index[last])
	}
}
func Msat_VarHeapReadMax(p *Msat_VarHeap_t) int {
	if len(p.order.heap) == 0 {
		panic("empty heap")
	}
	return p.order.heap[0]
}
func Msat_VarHeapGetMax(p *Msat_VarHeap_t) int { v := Msat_VarHeapReadMax(p); p.order.pop(); return v }
func Msat_VarHeapReadMaxWeight(p *Msat_VarHeap_t) float64 {
	return p.order.pSat.pdActivity[Msat_VarHeapReadMax(p)]
}
func Msat_VarHeapCountNodes(p *Msat_VarHeap_t, limit float64) int {
	n := 0
	for _, v := range p.order.heap {
		if p.order.pSat.pdActivity[v] >= limit {
			n++
		}
	}
	return n
}
func Msat_VarHeapCheck(p *Msat_VarHeap_t) {
	if Msat_OrderCheck(p.order) == 0 {
		panic("invalid indexed heap")
	}
}
func Msat_VarHeapCheckOne(p *Msat_VarHeap_t, v int) {
	if Msat_VarHeapContainsVar(p, v) == 0 {
		return
	}
	i := p.order.index[v]
	if p.order.heap[i] != v {
		panic("heap inverse mapping corrupted")
	}
	if i > 0 && p.order.better(v, p.order.heap[(i-1)/2]) {
		panic("heap ordering corrupted")
	}
	for _, j := range []int{2*i + 1, 2*i + 2} {
		if j < len(p.order.heap) && p.order.better(p.order.heap[j], v) {
			panic("heap ordering corrupted")
		}
	}
}
func Msat_VarHeapPrint(w io.Writer, p *Msat_VarHeap_t) {
	for _, v := range p.order.heap {
		fmt.Fprintf(w, "%d:%g ", v, p.order.pSat.pdActivity[v])
	}
	fmt.Fprintln(w)
}
