package aig

// Aig_NodeGetLeafCostOne scores the number of new frontier leaves needed to
// expand a marked node. High-fanout nodes expand only when frontier size cannot
// increase, matching ABC's reconvergence-driven one-level heuristic.
func Aig_NodeGetLeafCostOne(p *Graph, e Edge, fanoutLimit int32) int32 {
	n := p.node(e)
	if !n.MarkA {
		panic("aig: unmarked cut frontier")
	}
	if n.Kind == Input || n.Kind == Constant {
		return 999
	}
	var cost int32
	for _, fanin := range nodeFanins(*n) {
		if !p.node(fanin).MarkA {
			cost++
		}
	}
	if cost < 2 {
		return cost
	}
	if n.Refs > int(fanoutLimit) {
		return 999
	}
	return cost
}
func Aig_ManFindCut_int(p *Graph, front, visited *[]Edge, sizeLimit, fanoutLimit int32) bool {
	best := -1
	costBest := int32(100)
	for i, e := range *front {
		cost := Aig_NodeGetLeafCostOne(p, e, fanoutLimit)
		if cost < costBest || (cost == costBest && best >= 0 && p.node(e).Level > p.node((*front)[best]).Level) {
			best = i
			costBest = cost
		}
		if costBest == 0 {
			break
		}
	}
	if best < 0 || len(*front)-1+int(costBest) > int(sizeLimit) {
		return false
	}
	e := (*front)[best]
	*front = append((*front)[:best], (*front)[best+1:]...)
	for _, fanin := range nodeFanins(*p.node(e)) {
		fanin = Aig_Regular(fanin)
		if !p.node(fanin).MarkA {
			p.node(fanin).MarkA = true
			*front = append(*front, fanin)
			*visited = append(*visited, fanin)
		}
	}
	return true
}
func Aig_ManFindCut(p *Graph, root Edge, front, visited *[]Edge, sizeLimit, fanoutLimit int32) {
	if root.Complemented() || !Aig_ObjIsNode(p, root) {
		panic("aig: cut root must be a regular internal node")
	}
	if sizeLimit < 2 {
		panic("aig: cut limit is smaller than root frontier")
	}
	n := p.node(root)
	a, b := Aig_Regular(n.Fanins[0]), Aig_Regular(n.Fanins[1])
	if p.node(root).MarkA || p.node(a).MarkA || p.node(b).MarkA {
		panic("aig: cut marks were not cleared")
	}
	*front = append((*front)[:0], a, b)
	*visited = append((*visited)[:0], root, a, b)
	for _, e := range *visited {
		p.node(e).MarkA = true
	}
	defer func() {
		for _, e := range *visited {
			p.node(e).MarkA = false
		}
	}()
	for Aig_ManFindCut_int(p, front, visited, sizeLimit, fanoutLimit) {
	}
}
