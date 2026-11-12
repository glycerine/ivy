package abc

// Abc_ManCut_t holds the small and containing reconvergence-driven cuts.
type Abc_ManCut_t struct {
	nodeLimit, coneLimit       int
	nodeFanStop, coneFanStop   int
	small, large, visited, tfo []*Object
}

func Abc_NtkManCutStart(nodeLimit, coneLimit, nodeFanStop, coneFanStop int32) *Abc_ManCut_t {
	if nodeLimit < 2 || coneLimit < nodeLimit {
		panic("abc: invalid cut size limits")
	}
	return &Abc_ManCut_t{nodeLimit: int(nodeLimit), coneLimit: int(coneLimit), nodeFanStop: int(nodeFanStop), coneFanStop: int(coneFanStop)}
}

// Abc_NtkManCutStop releases the manager's retained object references.
func Abc_NtkManCutStop(m *Abc_ManCut_t)                   { m.small, m.large, m.visited, m.tfo = nil, nil, nil, nil }
func Abc_NtkManCutReadCutSmall(m *Abc_ManCut_t) []*Object { return m.small }
func Abc_NtkManCutReadCutLarge(m *Abc_ManCut_t) []*Object { return m.large }
func Abc_NtkManCutReadVisited(m *Abc_ManCut_t) []*Object  { return m.visited }

// Abc_NodeFindCut greedily expands leaves with the lowest boundary cost.
func Abc_NodeFindCut(m *Abc_ManCut_t, root *Object, contain int32) []*Object {
	if root == nil || root.Kind != ObjectNode || len(root.fanins) != 2 {
		panic("abc: cut root must be a two-input node")
	}
	if contain != 0 && m.coneLimit <= m.nodeLimit {
		panic("abc: containing cut limit must exceed node limit")
	}
	m.visited = append(m.visited[:0], root, root.fanins[0].Object, root.fanins[1].Object)
	m.small = append(m.small[:0], root.fanins[0].Object, root.fanins[1].Object)
	m.large = m.large[:0]
	seen := make(map[*Object]bool)
	for _, node := range m.visited {
		seen[node] = true
	}
	expand := func(leaves []*Object, limit, fanStop int) []*Object {
		for {
			best, bestCost := -1, 100
			for i, node := range leaves {
				if node.Kind != ObjectNode || len(node.fanins) != 2 {
					continue
				}
				cost := 0
				for _, edge := range node.fanins {
					if !seen[edge.Object] {
						cost++
					}
				}
				if cost >= 2 && len(node.fanouts) > fanStop {
					continue
				}
				if cost < bestCost || cost == bestCost && node.Level > leaves[best].Level {
					best, bestCost = i, cost
				}
				if bestCost == 0 {
					break
				}
			}
			if best < 0 || len(leaves)-1+bestCost > limit {
				return leaves
			}
			node := leaves[best]
			leaves = append(leaves[:best], leaves[best+1:]...)
			for _, edge := range node.fanins {
				if !seen[edge.Object] {
					seen[edge.Object] = true
					leaves = append(leaves, edge.Object)
					m.visited = append(m.visited, edge.Object)
				}
			}
		}
	}
	m.small = expand(m.small, m.nodeLimit, m.nodeFanStop)
	if contain != 0 {
		m.large = append(m.large, m.small...)
		m.large = expand(m.large, m.coneLimit, m.coneFanStop)
	}
	return m.small
}

// Abc_NodeCollectTfoCands finds nodes whose fanins are covered by the cut,
// excluding nodes in the root's maximum fanout-free cone.
func Abc_NodeCollectTfoCands(m *Abc_ManCut_t, root *Object, leaves []*Object, levelMax int32) []*Object {
	if root == nil || root.network == nil || Abc_NtkIsStrash(root.network) == 0 {
		panic("abc: cut TFO requires a strashed network root")
	}
	n := root.network
	Abc_NtkIncrementTravId(n)
	levels := make(map[int][]*Object)
	queued := make(map[*Object]bool)
	levelMin := -1
	for _, leaf := range leaves {
		if leaf == nil || leaf.network != n {
			panic("abc: invalid cut leaf")
		}
		if leaf.Level > int(levelMax) {
			continue
		}
		Abc_NodeSetTravIdCurrent(leaf)
		if !queued[leaf] {
			levels[leaf.Level] = append(levels[leaf.Level], leaf)
			queued[leaf] = true
		}
		levelMin = max(levelMin, leaf.Level)
	}
	if levelMin < 0 {
		panic("abc: cut has no leaves within level limit")
	}
	Abc_NodeMffcLabelAig(root)
	m.tfo = m.tfo[:0]
	for {
		level := -1
		for candidate := range levels {
			if candidate >= levelMin && candidate <= int(levelMax) && (level < 0 || candidate < level) {
				level = candidate
			}
		}
		if level < 0 {
			break
		}
		current := levels[level]
		delete(levels, level)
		for _, node := range current {
			if Abc_NodeIsTravIdCurrent(node) == 0 {
				if len(node.fanins) != 2 || Abc_NodeIsTravIdCurrent(node.fanins[0].Object) == 0 || Abc_NodeIsTravIdCurrent(node.fanins[1].Object) == 0 {
					continue
				}
				m.tfo = append(m.tfo, node)
				Abc_NodeSetTravIdCurrent(node)
			}
			for _, fanout := range node.fanouts {
				if fanout.IsCO() || fanout.Level > int(levelMax) || Abc_NodeIsTravIdCurrent(fanout) != 0 || queued[fanout] {
					continue
				}
				levels[fanout.Level] = append(levels[fanout.Level], fanout)
				queued[fanout] = true
			}
		}
		levelMin = level
	}
	return m.tfo
}
