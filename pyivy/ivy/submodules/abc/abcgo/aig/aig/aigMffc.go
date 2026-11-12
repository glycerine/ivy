// Maximum fanout-free cones, ported from ABC aigMffc.c (Alan Mishchenko).
package aig

func mffcTerminal(n *Node) bool { return n.Kind == Input || n.Kind == Constant }
func mffcFanins(n *Node) []Edge {
	switch n.Kind {
	case And, ExclusiveOr:
		return n.Fanins[:]
	case Buffer, Output:
		return n.Fanins[:1]
	}
	panic("aig: unsupported MFFC object kind")
}
func mffcNodeCost(n *Node) int32 {
	if n.Kind == And || n.Kind == ExclusiveOr {
		return 1
	}
	return 0
}

// Dereference fanins until their external references or the level boundary stop
// the cone. pPower is optional, but must accompany a full probability slice.
func Aig_NodeDeref_rec(p *Graph, e Edge, levelMin uint32, power *float32, probabilities []float32) int32 {
	return mffcDeref(p, e, levelMin, power, probabilities, nil)
}
func mffcDeref(p *Graph, e Edge, levelMin uint32, power *float32, probabilities []float32, saved map[int]int) int32 {
	n := p.node(e)
	if probabilities != nil {
		if power == nil {
			panic("aig: power output missing")
		}
		*power = 0
	}
	if mffcTerminal(n) {
		return 0
	}
	var count int32
	for _, fanin := range mffcFanins(n) {
		f := p.node(fanin)
		if f.Refs <= 0 {
			panic("aig: dereferencing an unreferenced fanin")
		}
		mffcSaveReference(p, fanin, saved)
		f.Refs--
		var subtotal float32
		if f.Refs == 0 && (levelMin == 0 || f.Level > int(levelMin)) {
			count += mffcDeref(p, Aig_Regular(fanin), levelMin, &subtotal, probabilities, saved)
		}
		if probabilities != nil {
			probability := probabilities[fanin.NodeID()]
			*power += subtotal + 2*probability*(1-probability)
		}
	}
	return count + mffcNodeCost(n)
}
func Aig_NodeRef_rec(p *Graph, e Edge, levelMin uint32) int32 {
	return mffcRef(p, e, levelMin, false)
}
func Aig_NodeRefLabel_rec(p *Graph, e Edge, levelMin uint32) int32 {
	return mffcRef(p, e, levelMin, true)
}
func mffcRef(p *Graph, e Edge, levelMin uint32, label bool) int32 {
	n := p.node(e)
	if mffcTerminal(n) {
		return 0
	}
	if label {
		n.TravID = p.travID
	}
	var count int32
	for _, fanin := range mffcFanins(n) {
		f := p.node(fanin)
		old := f.Refs
		if old < 0 {
			panic("aig: referencing a negative fanout count")
		}
		f.Refs++
		if old == 0 && (levelMin == 0 || f.Level > int(levelMin)) {
			count += mffcRef(p, Aig_Regular(fanin), levelMin, label)
		}
	}
	return count + mffcNodeCost(n)
}
func Aig_NodeMffcSupp_rec(p *Graph, e Edge, levelMin uint32, support *[]Edge, topmost bool, skip Edge) {
	e = Aig_Regular(e)
	n := p.node(e)
	if n.TravID == p.travID {
		return
	}
	n.TravID = p.travID
	if !topmost && e != skip && (mffcTerminal(n) || n.Refs > 0 || n.Level <= int(levelMin)) {
		if support != nil {
			*support = append(*support, e)
		}
		return
	}
	if mffcTerminal(n) {
		return
	}
	for _, fanin := range mffcFanins(n) {
		Aig_NodeMffcSupp_rec(p, fanin, levelMin, support, false, skip)
	}
}

// Public cone queries restore references even if malformed input panics. This
// also preserves the original reference values when checking cone symmetry.
func mffcSaveReference(p *Graph, e Edge, saved map[int]int) {
	if saved == nil {
		return
	}
	if _, exists := saved[e.NodeID()]; !exists {
		saved[e.NodeID()] = p.node(e).Refs
	}
}
func mffcRestoreReferences(p *Graph, saved map[int]int) {
	for id, refs := range saved {
		p.nodes[id].Refs = refs
	}
}
func mffcRequireRegular(p *Graph, e Edge) {
	if !p.valid(e) || e.Complemented() {
		panic("aig: MFFC root must be a regular graph edge")
	}
}
func Aig_NodeMffcSupp(p *Graph, e Edge, levelMin int32, support *[]Edge) int32 {
	mffcRequireRegular(p, e)
	if levelMin < 0 {
		panic("aig: negative MFFC level")
	}
	if support != nil {
		*support = (*support)[:0]
	}
	if n := p.node(e); mffcTerminal(n) {
		if n.Kind == Input && support != nil {
			*support = append(*support, e)
		}
		return 0
	}
	saved := make(map[int]int)
	defer mffcRestoreReferences(p, saved)
	mffcIncrementTravID(p)
	size := mffcDeref(p, e, uint32(levelMin), nil, nil, saved)
	Aig_NodeMffcSupp_rec(p, e, uint32(levelMin), support, true, InvalidEdge)
	if restored := Aig_NodeRef_rec(p, e, uint32(levelMin)); restored != size {
		panic("aig: asymmetric MFFC references")
	}
	return size
}
func Aig_NodeMffcLabel(p *Graph, e Edge, power *float32, probabilities []float32) int32 {
	mffcRequireRegular(p, e)
	if mffcTerminal(p.node(e)) {
		panic("aig: MFFC label requires an internal node")
	}
	if (power != nil) != (probabilities != nil) {
		panic("aig: power and probabilities must be supplied together")
	}
	if probabilities != nil && len(probabilities) < len(p.nodes) {
		panic("aig: incomplete MFFC probabilities")
	}
	saved := make(map[int]int)
	defer mffcRestoreReferences(p, saved)
	mffcIncrementTravID(p)
	size := mffcDeref(p, e, 0, power, probabilities, saved)
	if restored := Aig_NodeRefLabel_rec(p, e, 0); restored != size {
		panic("aig: asymmetric MFFC references")
	}
	return size
}
func Aig_NodeMffcLabelCut(p *Graph, e Edge, leaves []Edge) int32 {
	mffcRequireRegular(p, e)
	if mffcTerminal(p.node(e)) {
		panic("aig: MFFC cut requires an internal node")
	}
	saved := make(map[int]int)
	defer mffcRestoreReferences(p, saved)
	mffcIncrementTravID(p)
	for _, leaf := range leaves {
		mffcSaveReference(p, leaf, saved)
		p.node(leaf).Refs++
	}
	size := mffcDeref(p, e, 0, nil, nil, saved)
	if restored := Aig_NodeRefLabel_rec(p, e, 0); restored != size {
		panic("aig: asymmetric cut references")
	}
	for _, leaf := range leaves {
		p.node(leaf).Refs--
	}
	return size
}
func Aig_NodeMffcExtendCut(p *Graph, e Edge, leaves []Edge, result *[]Edge) bool {
	mffcRequireRegular(p, e)
	if result == nil {
		panic("aig: missing extended cut output")
	}
	maxLevel := 0
	for _, leaf := range leaves {
		if level := p.node(leaf).Level; level > maxLevel {
			maxLevel = level
		}
	}
	if maxLevel == 0 {
		return false
	}
	saved := make(map[int]int)
	defer mffcRestoreReferences(p, saved)
	size := mffcDeref(p, e, 0, nil, nil, saved)
	best := InvalidEdge
	bestCount := int32(1<<31 - 1)
	for _, leaf := range leaves {
		if p.node(leaf).Level != maxLevel {
			continue
		}
		count := mffcDeref(p, Aig_Regular(leaf), 0, nil, nil, saved)
		if count < bestCount {
			bestCount = count
			best = Aig_Regular(leaf)
		}
		if restored := Aig_NodeRef_rec(p, Aig_Regular(leaf), 0); restored != count {
			panic("aig: asymmetric candidate references")
		}
	}
	if best == InvalidEdge || mffcTerminal(p.node(best)) {
		panic("aig: no internal cut leaf to expand")
	}
	extra := mffcDeref(p, best, 0, nil, nil, saved)
	*result = (*result)[:0]
	mffcIncrementTravID(p)
	Aig_NodeMffcSupp_rec(p, e, 0, result, true, best)
	if restored := Aig_NodeRef_rec(p, best, 0); restored != extra {
		panic("aig: asymmetric extended cut references")
	}
	if restored := Aig_NodeRef_rec(p, e, 0); restored != size {
		panic("aig: asymmetric MFFC references")
	}
	return true
}

// Advance traversal identity, clearing old marks before the counter wraps.
func mffcIncrementTravID(p *Graph) {
	if p.travID >= (1<<30)-1 {
		for i := range p.nodes {
			p.nodes[i].TravID = 0
		}
		p.travID = 0
	}
	p.travID++
}
