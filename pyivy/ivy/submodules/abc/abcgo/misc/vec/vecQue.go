package vec

import "fmt"

// Vec_Que_t is an indexed maximum-priority heap. The priority slice is accessed
// indirectly so callers can replace/reallocate it without rebuilding the queue.
type Vec_Que_t struct {
	nSize, nCap   int
	pHeap, pOrder []int
	pCostsFlt     *[]float32
}

func Vec_QuePrio(p *Vec_Que_t, v int) float32 {
	if p.pCostsFlt != nil && *p.pCostsFlt != nil {
		return (*p.pCostsFlt)[v]
	}
	return float32(v)
}
func Vec_QueAlloc(n int) *Vec_Que_t {
	n = max(n, 16) + 1
	p := &Vec_Que_t{nSize: 1}
	Vec_QueGrow(p, n)
	return p
}
func Vec_QueFree(p *Vec_Que_t) {
	if p != nil {
		*p = Vec_Que_t{}
	}
}
func Vec_QueFreeP(p **Vec_Que_t) { Vec_QueFree(*p); *p = nil }
func Vec_QueSetPriority(p *Vec_Que_t, costs *[]float32) {
	if p.pCostsFlt != nil {
		panic("priority already assigned")
	}
	p.pCostsFlt = costs
}
func Vec_QueGrow(p *Vec_Que_t, n int) {
	if n <= p.nCap {
		return
	}
	for i := p.nCap; i < n; i++ {
		p.pHeap = append(p.pHeap, -1)
		p.pOrder = append(p.pOrder, -1)
	}
	p.nCap = n
}
func Vec_QueClear(p *Vec_Que_t) {
	for i := 1; i < p.nSize; i++ {
		p.pOrder[p.pHeap[i]] = -1
		p.pHeap[i] = -1
	}
	p.nSize = 1
}
func Vec_QueSize(p *Vec_Que_t) int { return p.nSize - 1 }
func Vec_QueTop(p *Vec_Que_t) int {
	if p.nSize <= 1 {
		panic("empty queue")
	}
	return p.pHeap[1]
}
func Vec_QueTopPriority(p *Vec_Que_t) float32 { return Vec_QuePrio(p, Vec_QueTop(p)) }
func Vec_QueMoveUp(p *Vec_Que_t, v int) int {
	i := p.pOrder[v]
	if i < 1 {
		panic("queue member required")
	}
	cost := Vec_QuePrio(p, v)
	moved := 0
	for i > 1 && cost > Vec_QuePrio(p, p.pHeap[i/2]) {
		p.pHeap[i] = p.pHeap[i/2]
		p.pOrder[p.pHeap[i]] = i
		i /= 2
		moved = 1
	}
	p.pHeap[i] = v
	p.pOrder[v] = i
	return moved
}
func Vec_QueMoveDown(p *Vec_Que_t, v int) {
	i := p.pOrder[v]
	if i < 1 {
		panic("queue member required")
	}
	cost := Vec_QuePrio(p, v)
	for child := 2 * i; child < p.nSize; child = 2 * i {
		if child+1 < p.nSize && Vec_QuePrio(p, p.pHeap[child]) < Vec_QuePrio(p, p.pHeap[child+1]) {
			child++
		}
		if cost >= Vec_QuePrio(p, p.pHeap[child]) {
			break
		}
		p.pHeap[i] = p.pHeap[child]
		p.pOrder[p.pHeap[i]] = i
		i = child
	}
	p.pHeap[i] = v
	p.pOrder[v] = i
}
func Vec_QueUpdate(p *Vec_Que_t, v int) {
	if Vec_QueMoveUp(p, v) == 0 {
		Vec_QueMoveDown(p, v)
	}
}
func Vec_QueIsMember(p *Vec_Que_t, v int) int {
	if v >= 0 && v < p.nCap && p.pOrder[v] >= 0 {
		return 1
	}
	return 0
}
func Vec_QuePush(p *Vec_Que_t, v int) {
	if v < 0 {
		panic("negative queue member")
	}
	if p.nSize >= p.nCap || v >= p.nCap {
		Vec_QueGrow(p, max(max(p.nSize+1, v+1), 2*p.nCap))
	}
	if p.pOrder[v] != -1 {
		panic("duplicate queue member")
	}
	p.pOrder[v] = p.nSize
	p.pHeap[p.nSize] = v
	p.nSize++
	Vec_QueMoveUp(p, v)
}
func Vec_QuePop(p *Vec_Que_t) int {
	result := Vec_QueTop(p)
	p.pOrder[result] = -1
	p.nSize--
	if p.nSize == 1 {
		p.pHeap[1] = -1
		return result
	}
	v := p.pHeap[p.nSize]
	p.pHeap[p.nSize] = -1
	p.pHeap[1] = v
	p.pOrder[v] = 1
	Vec_QueMoveDown(p, v)
	return result
}
func Vec_QuePrint(p *Vec_Que_t) {
	for start, width := 1, 1; start < p.nSize; start, width = start+width, width*2 {
		for i := start; i < min(start+width, p.nSize); i++ {
			fmt.Printf("%-5d", p.pHeap[i])
		}
		fmt.Println()
		for i := start; i < min(start+width, p.nSize); i++ {
			fmt.Printf("%-5.0f", Vec_QuePrio(p, p.pHeap[i]))
		}
		fmt.Print("\n\n")
	}
}
func Vec_QueCheck(p *Vec_Que_t) {
	if p.nSize < 1 || p.nSize > p.nCap || p.pHeap[0] != -1 {
		panic("queue bounds")
	}
	for i := 1; i < p.nSize; i++ {
		v := p.pHeap[i]
		if v < 0 || v >= p.nCap || p.pOrder[v] != i {
			panic("queue inverse mapping")
		}
		if i > 1 && Vec_QuePrio(p, p.pHeap[i/2]) < Vec_QuePrio(p, v) {
			panic("queue priority order")
		}
	}
	for v, i := range p.pOrder {
		if i >= 0 && (i < 1 || i >= p.nSize || p.pHeap[i] != v) {
			panic("queue membership mapping")
		}
	}
}
func Vec_QueTest(costs *Vec_Flt_t) {
	p := Vec_QueAlloc(costs.nSize)
	Vec_QueSetPriority(p, Vec_FltArrayP(costs))
	for i := range costs.nSize {
		Vec_QuePush(p, i)
	}
	Vec_QueCheck(p)
	for Vec_QueSize(p) > 0 {
		Vec_QuePop(p)
		Vec_QueCheck(p)
	}
}
