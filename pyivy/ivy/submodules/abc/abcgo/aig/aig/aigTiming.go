package aig

import "fmt"

func Aig_ObjReverseLevel(p *Graph, e Edge) int32 {
	if p.reverseLevels == nil {
		panic("aig: reverse levels not active")
	}
	if !p.valid(e) {
		panic("aig: invalid reverse-level edge")
	}
	for len(p.reverseLevels) <= e.NodeID() {
		p.reverseLevels = append(p.reverseLevels, 0)
	}
	return p.reverseLevels[e.NodeID()]
}
func Aig_ObjSetReverseLevel(p *Graph, e Edge, level int32) {
	if level < 0 {
		panic("aig: negative reverse level")
	}
	Aig_ObjReverseLevel(p, e)
	p.reverseLevels[e.NodeID()] = level
}
func Aig_ObjClearReverseLevel(p *Graph, e Edge) { Aig_ObjSetReverseLevel(p, e, 0) }
func Aig_ObjRequiredLevel(p *Graph, e Edge) int32 {
	return p.maxRequiredLevel + 1 - Aig_ObjReverseLevel(p, e)
}
func Aig_ObjReverseLevelNew(p *Graph, e Edge) int32 {
	max := int32(0)
	for i, fan := int32(0), Aig_ObjFanout0Int(p, int32(e.NodeID())); i < Aig_ObjRefs(p, e); i, fan = i+1, Aig_ObjFanoutNext(p, fan) {
		if level := Aig_ObjReverseLevel(p, Aig_ManObj(p, fan>>1)); level > max {
			max = level
		}
	}
	return max + 1
}
func Aig_ManStartReverseLevels(p *Graph, increase int32) {
	if p.fanouts == nil {
		panic("aig: reverse levels require an active fanout index")
	}
	if p.reverseLevels != nil {
		panic("aig: reverse levels already active")
	}
	if increase < 0 {
		panic("aig: negative level allowance")
	}
	p.maxRequiredLevel = Aig_ManLevels(p) + increase
	p.reverseLevels = make([]int32, len(p.nodes))
	for _, e := range Aig_ManDfsReverse(p) {
		if p.node(e).MarkA {
			panic("aig: reverse level setup requires clear node marks")
		}
		Aig_ObjSetReverseLevel(p, e, Aig_ObjReverseLevelNew(p, e))
	}
}
func Aig_ManStopReverseLevels(p *Graph) {
	if p.reverseLevels == nil {
		panic("aig: reverse levels not active")
	}
	p.reverseLevels = nil
	p.maxRequiredLevel = 0
}
func Aig_ManUpdateLevel(p *Graph, e Edge) {
	if p.fanouts == nil || e.Complemented() || !p.valid(e) || !Aig_ObjIsNode(p, e) {
		panic("aig: incremental levels require fanouts and a regular logic node")
	}
	old := p.node(e).Level
	if old == int(Aig_ObjLevelNew(p, e)) {
		return
	}
	var queue timingQueue
	queue.add(p, e, old)
	for level := old; level < len(queue); level++ {
		for i := 0; i < len(queue[level]); i++ {
			current := queue[level][i]
			n := p.node(current)
			n.MarkA = false
			if n.Level != level {
				panic("aig: forward timing queue has inconsistent old level")
			}
			n.Level = int(Aig_ObjLevelNew(p, current))
			if n.Level == level {
				continue
			}
			for k, fan := 0, Aig_ObjFanout0Int(p, int32(current.NodeID())); k < n.Refs; k, fan = k+1, Aig_ObjFanoutNext(p, fan) {
				next := Aig_ManObj(p, fan>>1)
				if Aig_ObjIsNode(p, next) && !p.node(next).MarkA {
					if p.node(next).Level < level {
						panic("aig: fanout old level precedes timing queue")
					}
					queue.add(p, next, p.node(next).Level)
				}
			}
		}
	}
}
func Aig_ManUpdateReverseLevel(p *Graph, e Edge) {
	if p.reverseLevels == nil || p.fanouts == nil || e.Complemented() || !p.valid(e) || !Aig_ObjIsNode(p, e) {
		panic("aig: incremental reverse levels require timing, fanouts and a regular logic node")
	}
	old := int(Aig_ObjReverseLevel(p, e))
	if old == int(Aig_ObjReverseLevelNew(p, e)) {
		return
	}
	var queue timingQueue
	queue.add(p, e, old)
	for level := old; level < len(queue); level++ {
		// Appending a newly allocated fanin to this same old-level bucket is
		// allowed. Read the live bucket length, as C's Vec_Vec iterator does.
		for i := 0; i < len(queue[level]); i++ {
			current := queue[level][i]
			p.node(current).MarkA = false
			if int(Aig_ObjReverseLevel(p, current)) != level {
				panic("aig: reverse timing queue has inconsistent old level")
			}
			nextLevel := Aig_ObjReverseLevelNew(p, current)
			Aig_ObjSetReverseLevel(p, current, nextLevel)
			if int(nextLevel) == level {
				continue
			}
			for _, fanin := range nodeFanins(*p.node(current)) {
				fanin = Aig_Regular(fanin)
				if Aig_ObjIsNode(p, fanin) && !p.node(fanin).MarkA {
					faninLevel := int(Aig_ObjReverseLevel(p, fanin))
					if faninLevel < level {
						panic("aig: fanin old reverse level precedes timing queue")
					}
					queue.add(p, fanin, faninLevel)
				}
			}
		}
	}
}

// timingQueue holds nodes by their old levels. MarkA prevents duplicate work
// while a node is queued; each processing step clears its own queue mark.
type timingQueue [][]Edge

func (q *timingQueue) add(p *Graph, e Edge, level int) {
	if level < 0 {
		panic("aig: negative timing queue level")
	}
	if level >= len(*q) {
		*q = append(*q, make([][]Edge, level-len(*q)+1)...)
	}
	(*q)[level] = append((*q)[level], e)
	p.node(e).MarkA = true
}

// RefreshReverseLevels recomputes all reverse levels for Go graph rewrites
// that alter several nodes or preserve replacement IDs through buffers.
// Source Aig_ManUpdateReverseLevel instead propagates from one changed node.
func (p *Graph) RefreshReverseLevels() {
	if p.reverseLevels == nil || p.fanouts == nil {
		panic("aig: reverse level refresh requires active timing and fanouts")
	}
	p.reverseLevels = make([]int32, len(p.nodes))
	for _, e := range Aig_ManDfsReverse(p) {
		Aig_ObjSetReverseLevel(p, e, Aig_ObjReverseLevelNew(p, e))
	}
}
func Aig_ManVerifyLevel(p *Graph) error {
	if p.fanouts == nil {
		return fmt.Errorf("aig: level verification requires active fanouts")
	}
	for id, n := range p.nodes {
		if p.deleted[id] || n.Kind != And && n.Kind != ExclusiveOr {
			continue
		}
		expected := Aig_ObjLevelNew(p, Edge(id)<<1)
		if int32(n.Level) != expected {
			return fmt.Errorf("aig: node %d level %d, expected %d", id, n.Level, expected)
		}
	}
	return nil
}
func Aig_ManVerifyReverseLevel(p *Graph) error {
	if p.reverseLevels == nil {
		return fmt.Errorf("aig: reverse levels not active")
	}
	for id, n := range p.nodes {
		if p.deleted[id] || (n.Kind != And && n.Kind != ExclusiveOr && n.Kind != Buffer) {
			continue
		}
		edge := Edge(id) << 1
		expected := Aig_ObjReverseLevelNew(p, edge)
		if actual := Aig_ObjReverseLevel(p, edge); actual != expected {
			return fmt.Errorf("aig: node %d reverse level %d, expected %d", id, actual, expected)
		}
	}
	return nil
}
