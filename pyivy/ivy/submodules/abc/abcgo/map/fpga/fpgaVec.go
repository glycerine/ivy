package fpga

import "sort"

func Fpga_NodeVecAlloc(capacity int32) *Fpga_NodeVec_t {
	return &Fpga_NodeVec_t{Nodes: make([]*Fpga_Node_t, 0, capacity)}
}
func Fpga_NodeVecFree(p *Fpga_NodeVec_t) {
	if p != nil {
		p.Nodes = nil
	}
}
func Fpga_NodeVecReadArray(p *Fpga_NodeVec_t) []*Fpga_Node_t { return p.Nodes }
func Fpga_NodeVecReadSize(p *Fpga_NodeVec_t) int32           { return int32(len(p.Nodes)) }
func Fpga_NodeVecGrow(p *Fpga_NodeVec_t, n int32) {
	if int(n) > cap(p.Nodes) {
		a := make([]*Fpga_Node_t, len(p.Nodes), n)
		copy(a, p.Nodes)
		p.Nodes = a
	}
}
func Fpga_NodeVecShrink(p *Fpga_NodeVec_t, n int32)      { p.Nodes = p.Nodes[:n] }
func Fpga_NodeVecClear(p *Fpga_NodeVec_t)                { clear(p.Nodes); p.Nodes = p.Nodes[:0] }
func Fpga_NodeVecPush(p *Fpga_NodeVec_t, n *Fpga_Node_t) { p.Nodes = append(p.Nodes, n) }
func Fpga_NodeVecPushUnique(p *Fpga_NodeVec_t, n *Fpga_Node_t) int32 {
	for _, v := range p.Nodes {
		if v == n {
			return 1
		}
	}
	Fpga_NodeVecPush(p, n)
	return 0
}
func Fpga_NodeVecPop(p *Fpga_NodeVec_t) *Fpga_Node_t {
	i := len(p.Nodes) - 1
	n := p.Nodes[i]
	p.Nodes[i] = nil
	p.Nodes = p.Nodes[:i]
	return n
}
func Fpga_NodeVecWriteEntry(p *Fpga_NodeVec_t, i int32, n *Fpga_Node_t) { p.Nodes[i] = n }
func Fpga_NodeVecReadEntry(p *Fpga_NodeVec_t, i int32) *Fpga_Node_t     { return p.Nodes[i] }
func Fpga_NodeVecCompareLevels(a, b **Fpga_Node_t) int32 {
	x, y := Fpga_Regular(*a), Fpga_Regular(*b)
	if x.Level < y.Level || x.Level == y.Level && x.Num < y.Num {
		return -1
	}
	if x.Level == y.Level && x.Num == y.Num {
		return 0
	}
	return 1
}
func Fpga_NodeVecSortByLevel(p *Fpga_NodeVec_t) {
	sort.SliceStable(p.Nodes, func(i, j int) bool { return Fpga_NodeVecCompareLevels(&p.Nodes[i], &p.Nodes[j]) < 0 })
}
func Fpga_NodeVecCompareArrivals(a, b **Fpga_Node_t) int32 {
	x, y := Fpga_Regular(*a).Best.Arrival, Fpga_Regular(*b).Best.Arrival
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}
func Fpga_SortNodesByArrivalTimes(p *Fpga_NodeVec_t) {
	sort.SliceStable(p.Nodes, func(i, j int) bool { return Fpga_NodeVecCompareArrivals(&p.Nodes[i], &p.Nodes[j]) < 0 })
}

// Union preserves the source routine's concatenation semantics (including duplicates).
func Fpga_NodeVecUnion(p, a, b *Fpga_NodeVec_t) {
	p.Nodes = append(append([]*Fpga_Node_t{}, a.Nodes...), b.Nodes...)
}
func Fpga_NodeVecPushOrder(p *Fpga_NodeVec_t, n *Fpga_Node_t, increasing int32) {
	p.Nodes = append(p.Nodes, n)
	sort.SliceStable(p.Nodes, func(i, j int) bool {
		a, b := Fpga_Regular(p.Nodes[i]).Best.Arrival, Fpga_Regular(p.Nodes[j]).Best.Arrival
		if increasing != 0 {
			return a < b
		}
		return a > b
	})
}
func Fpga_NodeVecReverse(p *Fpga_NodeVec_t) {
	for i, j := 0, len(p.Nodes)-1; i < j; i, j = i+1, j-1 {
		p.Nodes[i], p.Nodes[j] = p.Nodes[j], p.Nodes[i]
	}
}
