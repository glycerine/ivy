package aig

// orderLinks represents ABC's circular object-order list. Node zero is its
// sentinel; a missing object uses negative links instead of an unsigned mask.
type orderLinks struct{ prev, next int32 }

func Aig_ManOrderStart(p *Graph) {
	if p.order != nil {
		panic("aig: order already active")
	}
	if Aig_ManBufNum(p) != 0 {
		panic("aig: order requires propagated buffers")
	}
	p.order = make([]orderLinks, len(p.nodes))
	for i := range p.order {
		p.order[i] = orderLinks{-1, -1}
	}
	p.order[0] = orderLinks{0, 0}
	p.iPrev, p.iNext, p.nAndTotal, p.nAndPrev = 0, 0, 0, 0
	for id, n := range p.nodes {
		if !p.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr) {
			Aig_ObjOrderInsert(p, int32(id))
		}
	}
}
func Aig_ManOrderStop(p *Graph) {
	if p.order == nil {
		panic("aig: order not active")
	}
	p.order = nil
	p.iPrev, p.iNext, p.nAndTotal, p.nAndPrev = 0, 0, 0, 0
}
func Aig_ObjOrderInsert(p *Graph, id int32) {
	if p.order == nil || id <= 0 || !p.valid(Edge(id)<<1) || !Aig_ObjIsNode(p, Edge(id)<<1) {
		panic("aig: invalid order insertion")
	}
	for int(id) >= len(p.order) {
		p.order = append(p.order, orderLinks{-1, -1})
	}
	if p.order[id] != (orderLinks{-1, -1}) {
		panic("aig: duplicate order insertion")
	}
	previous := p.order[p.iNext].prev
	if p.order[previous].next != p.iNext {
		panic("aig: corrupt order list")
	}
	p.order[id] = orderLinks{previous, p.iNext}
	p.order[previous].next = id
	p.order[p.iNext].prev = id
	p.nAndTotal++
}
func Aig_ObjOrderRemove(p *Graph, id int32) {
	if p.order == nil || id <= 0 || int(id) >= len(p.order) {
		panic("aig: invalid order removal")
	}
	links := p.order[id]
	if links.prev < 0 || links.next < 0 {
		panic("aig: missing order entry")
	}
	p.order[links.prev].next = links.next
	p.order[links.next].prev = links.prev
	p.order[id] = orderLinks{-1, -1}
	if p.iPrev == id {
		p.nAndPrev--
		p.iPrev = links.prev
	}
	if p.iNext == id {
		p.iNext = links.next
	}
	p.nAndTotal--
}
func Aig_ObjOrderAdvance(p *Graph) {
	if p.order == nil || p.order[p.iPrev].next != p.iNext {
		panic("aig: inconsistent order cursor")
	}
	p.iPrev = p.iNext
	p.nAndPrev++
}

// OrderedEdges snapshots the active circular order without exposing its links.
func (p *Graph) OrderedEdges() []Edge {
	if p.order == nil {
		return nil
	}
	result := make([]Edge, 0, p.nAndTotal)
	for id := p.order[0].next; id != 0; id = p.order[id].next {
		if len(result) >= int(p.nAndTotal) {
			panic("aig: corrupt order cycle")
		}
		result = append(result, Edge(id)<<1)
	}
	if len(result) != int(p.nAndTotal) {
		panic("aig: inconsistent order count")
	}
	return result
}
