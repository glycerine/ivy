package vec

import "fmt"

// Vec_Que_t is an indexed maximum-priority heap. The priority slice is accessed
// indirectly so callers can replace/reallocate it without rebuilding the queue.
// Vec_Que_t keeps heap positions and entries as int32, as vecQue.h's int
// arrays (half the cache footprint of int).
type Vec_Que_t struct {
	pHeap, pOrder []int32
	pCostsFlt     *[]float32
}

func Vec_QuePrio(p *Vec_Que_t, v int) float32 {
	if p.pCostsFlt != nil && *p.pCostsFlt != nil {
		return (*p.pCostsFlt)[v]
	}
	return float32(v)
}

// queCosts is the priority array (nil: an entry's priority is itself).
func (p *Vec_Que_t) queCosts() []float32 {
	if p.pCostsFlt == nil {
		return nil
	}
	return *p.pCostsFlt
}
func quePrio(costs []float32, v int32) float32 {
	if costs != nil {
		return costs[v]
	}
	return float32(v)
}
func Vec_QueAlloc(n int) *Vec_Que_t {
	p := new(Vec_Que_t)
	Vec_QueGrow(p, max(n, 16)+1)
	p.pHeap = p.pHeap[:1]
	p.pHeap[0] = -1
	return p
}
func Vec_QueFree(p *Vec_Que_t) {
	if p != nil {
		*p = Vec_Que_t{}
	}
}
func Vec_QueFreeP(p **Vec_Que_t) { Vec_QueFree(*p); *p = nil }
func Vec_QueSetPriority(p *Vec_Que_t, costs *[]float32) {
	// ACB reattaches the same vector after growth; its slice-header address
	// is stable in Go. Changing to another priority source still needs a rebuild.
	if p.pCostsFlt != nil && p.pCostsFlt != costs {
		panic("priority already assigned")
	}
	p.pCostsFlt = costs
}
func Vec_QueGrow(p *Vec_Que_t, n int) {
	if n <= len(p.pOrder) {
		return
	}
	if n > int(^uint32(0)>>1) {
		panic("queue size overflows int32")
	}
	heap := make([]int32, len(p.pHeap), n)
	copy(heap, p.pHeap)
	p.pHeap = heap
	order := make([]int32, n)
	copy(order, p.pOrder)
	for i := len(p.pOrder); i < n; i++ {
		order[i] = -1
	}
	p.pOrder = order
}
func Vec_QueClear(p *Vec_Que_t) {
	for _, v := range p.pHeap[1:] {
		p.pOrder[v] = -1
	}
	p.pHeap = p.pHeap[:1]
}
func Vec_QueSize(p *Vec_Que_t) int { return len(p.pHeap) - 1 }
func Vec_QueTop(p *Vec_Que_t) int {
	if len(p.pHeap) <= 1 {
		return -1
	}
	return int(p.pHeap[1])
}
func Vec_QueTopPriority(p *Vec_Que_t) float32 {
	if Vec_QueSize(p) == 0 {
		return -1000000000
	} // -ABC_INFINITY
	return Vec_QuePrio(p, Vec_QueTop(p))
}
func Vec_QueMoveUp(p *Vec_Que_t, v int) int {
	heap, order, costs := p.pHeap, p.pOrder, p.queCosts()
	i := order[v]
	if i < 1 {
		panic("queue member required")
	}
	cost := quePrio(costs, int32(v))
	moved := 0
	for i > 1 {
		parent := heap[i/2]
		if cost <= quePrio(costs, parent) {
			break
		}
		heap[i] = parent
		order[parent] = i
		i /= 2
		moved = 1
	}
	heap[i] = int32(v)
	order[v] = i
	return moved
}
func Vec_QueMoveDown(p *Vec_Que_t, v int) {
	heap, order, costs := p.pHeap, p.pOrder, p.queCosts()
	i := order[v]
	if i < 1 {
		panic("queue member required")
	}
	cost := quePrio(costs, int32(v))
	n := int32(len(heap))
	for child := 2 * i; child < n; child = 2 * i {
		if child+1 < n && quePrio(costs, heap[child]) < quePrio(costs, heap[child+1]) {
			child++
		}
		if cost >= quePrio(costs, heap[child]) {
			break
		}
		heap[i] = heap[child]
		order[heap[i]] = i
		i = child
	}
	heap[i] = int32(v)
	order[v] = i
}
func Vec_QueUpdate(p *Vec_Que_t, v int) {
	if Vec_QueMoveUp(p, v) == 0 {
		Vec_QueMoveDown(p, v)
	}
}
func Vec_QueIsMember(p *Vec_Que_t, v int) int {
	if v >= 0 && v < len(p.pOrder) && p.pOrder[v] >= 0 {
		return 1
	}
	return 0
}
func Vec_QuePush(p *Vec_Que_t, v int) {
	if v < 0 {
		panic("negative queue member")
	}
	if len(p.pHeap) >= len(p.pOrder) || v >= len(p.pOrder) {
		Vec_QueGrow(p, max(max(len(p.pHeap)+1, v+1), 2*len(p.pOrder)))
	}
	if p.pOrder[v] != -1 {
		panic("duplicate queue member")
	}
	p.pOrder[v] = int32(len(p.pHeap))
	p.pHeap = append(p.pHeap, int32(v))
	Vec_QueMoveUp(p, v)
}
func Vec_QuePop(p *Vec_Que_t) int {
	result := Vec_QueTop(p)
	p.pOrder[result] = -1
	last := p.pHeap[len(p.pHeap)-1]
	p.pHeap = p.pHeap[:len(p.pHeap)-1]
	if len(p.pHeap) > 1 {
		p.pHeap[1] = last
		p.pOrder[last] = 1
		Vec_QueMoveDown(p, int(last))
	}
	return result
}
func Vec_QuePrint(p *Vec_Que_t) {
	for start, width := 1, 1; start < len(p.pHeap); start, width = start+width, width*2 {
		for i := start; i < min(start+width, len(p.pHeap)); i++ {
			fmt.Printf("%-5d", p.pHeap[i])
		}
		fmt.Println()
		for i := start; i < min(start+width, len(p.pHeap)); i++ {
			fmt.Printf("%-5.0f", Vec_QuePrio(p, int(p.pHeap[i])))
		}
		fmt.Print("\n\n")
	}
}
func Vec_QueCheck(p *Vec_Que_t) {
	if len(p.pHeap) < 1 || len(p.pHeap) > len(p.pOrder) || p.pHeap[0] != -1 {
		panic("queue bounds")
	}
	for i := 1; i < len(p.pHeap); i++ {
		v := int(p.pHeap[i])
		if v < 0 || v >= len(p.pOrder) || int(p.pOrder[v]) != i {
			panic("queue inverse mapping")
		}
		if i > 1 && Vec_QuePrio(p, int(p.pHeap[i/2])) < Vec_QuePrio(p, v) {
			panic("queue priority order")
		}
	}
	for v, i := range p.pOrder {
		if i >= 0 && (i < 1 || int(i) >= len(p.pHeap) || int(p.pHeap[i]) != v) {
			panic("queue membership mapping")
		}
	}
}
func Vec_QueTest(costs *Vec_Flt_t) {
	p := Vec_QueAlloc(len(costs.data))
	Vec_QueSetPriority(p, Vec_FltArrayP(costs))
	for i := range len(costs.data) {
		Vec_QuePush(p, i)
	}
	Vec_QueCheck(p)
	for Vec_QueSize(p) > 0 {
		Vec_QuePop(p)
		Vec_QueCheck(p)
	}
}
