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
	if !p.valid(e) {
		panic("aig: invalid level update edge")
	}
	p.recomputeLevels()
}
func Aig_ManUpdateReverseLevel(p *Graph, e Edge) {
	if p.reverseLevels == nil || !p.valid(e) {
		panic("aig: invalid reverse-level update")
	}
	p.recomputeReverseLevels()
}
func (p *Graph) recomputeReverseLevels() {
	p.reverseLevels = make([]int32, len(p.nodes))
	for _, e := range Aig_ManDfsReverse(p) {
		Aig_ObjSetReverseLevel(p, e, Aig_ObjReverseLevelNew(p, e))
	}
}
func Aig_ManVerifyLevel(p *Graph) error {
	for id, n := range p.nodes {
		if p.deleted[id] || n.Kind == Input || n.Kind == Constant {
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
