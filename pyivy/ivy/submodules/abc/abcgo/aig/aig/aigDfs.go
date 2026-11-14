package aig

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
)

func Aig_ManVerifyTopoOrder(p *Graph) int32 {
	Aig_ManSetCioIds(p)
	Aig_ManIncrementTravId(p)
	for _, e := range p.LiveEdges() {
		n := p.node(e)
		for _, fanin := range nodeFanins(*n) {
			if p.node(fanin).TravID != p.travID {
				fmt.Printf("Node %d has fanin %d that is not in a topological order.\n", e.NodeID(), fanin.NodeID())
				return 0
			}
		}
		if n.Kind == Input {
			if manager, ok := p.Timing.(*tim.Tim_Man_t); ok && manager != nil {
				box := tim.Tim_ManBoxForCi(manager, int(n.CioID))
				if box >= 0 {
					first, count := tim.Tim_ManBoxInputFirst(manager, box), tim.Tim_ManBoxInputNum(manager, box)
					for k := 0; k < count; k++ {
						co := p.outputs[first+k]
						assert(tim.Tim_ManBoxForCo(manager, int(p.node(co).CioID)) == box, "box input belongs to box")
						if p.node(co).TravID != p.travID {
							fmt.Printf("Box %d has input %d that is not in a topological order.\n", box, co.NodeID())
							return 0
						}
					}
				}
			}
		}
		n.TravID = p.travID
	}
	Aig_ManCleanCioIds(p)
	return 1
}
func (p *Graph) dfs(roots []Edge, nodesOnly, preorder, choices bool) []Edge {
	seen := make(map[int]bool)
	active := make(map[int]bool)
	var result []Edge
	var visit func(Edge)
	visit = func(e Edge) {
		id := e.NodeID()
		if seen[id] {
			return
		}
		if active[id] {
			panic("aig: cyclic traversal")
		}
		active[id] = true
		n := p.node(e)
		include := !nodesOnly || n.Kind == And || n.Kind == ExclusiveOr
		if preorder && include {
			result = append(result, Aig_Regular(e))
		}
		if n.Kind == And || n.Kind == ExclusiveOr || (n.Kind == Buffer || n.Kind == Output) {
			visit(n.Fanins[0])
			if n.Kind != Buffer && n.Kind != Output {
				visit(n.Fanins[1])
			}
		}
		if choices {
			if other, ok := p.equivs[id]; ok {
				visit(other)
			}
		}
		if !preorder && include {
			result = append(result, Aig_Regular(e))
		}
		delete(active, id)
		seen[id] = true
	}
	for _, e := range roots {
		visit(e)
	}
	return result
}
func Aig_ManDfs(p *Graph, nodesOnly int32) []Edge {
	return p.dfsOutputOrder(nodesOnly != 0, false)
}
func Aig_ManDfs_rec(p *Graph, e Edge, nodes *[]Edge) {
	if e == InvalidEdge {
		return
	}
	if e.Complemented() {
		panic("aig: recursive DFS requires a regular object")
	}
	n := p.node(e)
	if n.TravID == p.travID {
		return
	}
	n.TravID = p.travID
	if alternative, ok := p.equivs[e.NodeID()]; ok {
		Aig_ManDfs_rec(p, alternative, nodes)
	}
	for _, fanin := range nodeFanins(*n) {
		Aig_ManDfs_rec(p, Aig_Regular(fanin), nodes)
	}
	*nodes = append(*nodes, e)
}

// Ordinary and preorder DFS share C's setup and completeness assertions. Their
// recursive helpers differ in when they append the object; both traverse a
// choice alternative in ordinary postorder before visiting ordinary fanins.
func (p *Graph) dfsOutputOrder(nodesOnly, preorder bool) []Edge {
	Aig_ManIncrementTravId(p)
	p.node(p.Const1()).TravID = p.travID
	var nodes []Edge
	if nodesOnly {
		for _, e := range p.inputs {
			p.node(e).TravID = p.travID
		}
	} else {
		nodes = append(nodes, p.Const1())
	}
	for _, e := range p.outputs {
		if nodesOnly {
			e = p.outputDriver(e)
		}
		if preorder {
			Aig_ManDfsPreorder_rec(p, Aig_Regular(e), &nodes)
		} else {
			Aig_ManDfs_rec(p, Aig_Regular(e), &nodes)
		}
	}
	expected := Aig_ManObjNum(p)
	if nodesOnly {
		expected = Aig_ManNodeNum(p)
	}
	if int32(len(nodes)) != expected {
		panic("aig: DFS object count differs from manager count")
	}
	return nodes
}
func Aig_ManDfsAll_rec(p *Graph, e Edge, nodes *[]Edge) {
	n := p.node(e)
	if n.TravID == p.travID {
		return
	}
	n.TravID = p.travID
	switch n.Kind {
	case Input:
	case Output:
		Aig_ManDfsAll_rec(p, n.Fanins[0], nodes)
	case And, ExclusiveOr:
		Aig_ManDfsAll_rec(p, n.Fanins[0], nodes)
		Aig_ManDfsAll_rec(p, n.Fanins[1], nodes)
	default:
		panic("aig: all-object DFS requires a CI, CO or logic node")
	}
	*nodes = append(*nodes, Aig_Regular(e))
}
func Aig_ManDfsArray(p *Graph, nodes []Edge, n int32) []Edge {
	if n < 0 || int(n) > len(nodes) {
		panic("aig: invalid DFS range")
	}
	Aig_ManIncrementTravId(p)
	p.node(p.Const1()).TravID = p.travID
	result := []Edge{p.Const1()}
	for _, e := range nodes[:n] {
		Aig_ManDfsAll_rec(p, e, &result)
	}
	return result
}
func Aig_ManDfsAll(p *Graph) []Edge {
	nodes := Aig_ManDfsArray(p, p.outputs, int32(len(p.outputs)))
	for _, e := range p.inputs {
		if p.node(e).TravID != p.travID {
			nodes = append(nodes, e)
		}
	}
	if int32(len(nodes)) != Aig_ManObjNum(p) {
		panic("aig: DFS all object count differs from manager count")
	}
	return nodes
}
func Aig_ManDfsPreorder_rec(p *Graph, e Edge, nodes *[]Edge) {
	if e == InvalidEdge {
		return
	}
	if e.Complemented() {
		panic("aig: recursive preorder DFS requires a regular object")
	}
	n := p.node(e)
	if n.TravID == p.travID {
		return
	}
	n.TravID = p.travID
	*nodes = append(*nodes, e)
	if alternative, ok := p.equivs[e.NodeID()]; ok {
		Aig_ManDfs_rec(p, alternative, nodes)
	}
	for _, fanin := range nodeFanins(*n) {
		Aig_ManDfsPreorder_rec(p, Aig_Regular(fanin), nodes)
	}
}
func Aig_ManDfsPreorder(p *Graph, nodesOnly int32) []Edge {
	return p.dfsOutputOrder(nodesOnly != 0, true)
}
func Aig_ManDfsNodes(p *Graph, nodes []Edge, n int32) []Edge {
	if n < 0 || int(n) > len(nodes) {
		panic("aig: invalid DFS range")
	}
	Aig_ManIncrementTravId(p)
	p.node(p.Const1()).TravID = p.travID
	var result []Edge
	for _, e := range nodes[:n] {
		if p.node(e).Kind == Output {
			e = p.outputDriver(e)
		}
		Aig_ManDfs_rec(p, Aig_Regular(e), &result)
	}
	return result
}
func Aig_ManDfsChoices_rec(p *Graph, e Edge, nodes *[]Edge) {
	p.dfsChoices(e, nodes, make(map[int]bool))
}
func (p *Graph) dfsChoices(e Edge, nodes *[]Edge, active map[int]bool) {
	if e == InvalidEdge {
		return
	}
	n := p.node(e)
	if n.TravID == p.travID {
		return
	}
	if n.Kind != And && n.Kind != ExclusiveOr {
		panic("aig: choice DFS requires logic nodes or premarked terminals")
	}
	if active[e.NodeID()] {
		panic("aig: cyclic choice traversal")
	}
	active[e.NodeID()] = true
	p.dfsChoices(n.Fanins[0], nodes, active)
	p.dfsChoices(n.Fanins[1], nodes, active)
	if alternative, ok := p.equivs[e.NodeID()]; ok {
		p.dfsChoices(alternative, nodes, active)
	}
	delete(active, e.NodeID())
	n.TravID = p.travID
	*nodes = append(*nodes, Aig_Regular(e))
}
func Aig_ManDfsChoices(p *Graph) []Edge {
	if p.equivs == nil {
		panic("aig: choice DFS requires equivalence storage")
	}
	Aig_ManIncrementTravId(p)
	p.node(p.Const1()).TravID = p.travID
	for _, e := range p.inputs {
		p.node(e).TravID = p.travID
	}
	var nodes []Edge
	active := make(map[int]bool)
	for _, e := range p.outputs {
		p.dfsChoices(p.outputDriver(e), &nodes, active)
	}
	return nodes
}
func Aig_ManDfsReverse(p *Graph) []Edge {
	roots := make([]Edge, 0, len(p.nodes))
	for id, n := range p.nodes {
		if !p.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			roots = append(roots, Edge(id)<<1)
		}
	}
	return p.reverseDfs(roots, nil)
}
func Aig_ManDfsReverse_rec(p *Graph, e Edge, nodes *[]Edge) { *nodes = p.reverseDfs([]Edge{e}, *nodes) }
func Aig_ManLevelize(p *Graph) [][]Edge {
	maxLevel := int(Aig_ManLevelNum(p))
	levels := make([][]Edge, maxLevel+1)
	for id, n := range p.nodes {
		if p.deleted[id] {
			continue
		}
		if n.Level < 0 || n.Level > maxLevel {
			panic("aig: object level exceeds manager level")
		}
		levels[n.Level] = append(levels[n.Level], Edge(id)<<1)
	}
	return levels
}
func Aig_ManLevelNum(p *Graph) int32 {
	max := 0
	for _, e := range p.outputs {
		if level := p.node(e).Level; level > max {
			max = level
		}
	}
	return int32(max)
}
func Aig_ManChoiceLevel_rec(p *Graph, e Edge) { p.choiceLevels([]Edge{e}) }
func Aig_ManChoiceLevel(p *Graph) int32 {
	Aig_ManSetCioIds(p)
	defer Aig_ManCleanCioIds(p)
	for i := range p.nodes {
		p.nodes[i].Level = 0
	}
	roots := append(append([]Edge(nil), p.outputs...), p.inputs...)
	p.choiceLevels(roots)
	max := int32(0)
	for _, e := range roots {
		if level := int32(p.node(e).Level); level > max {
			max = level
		}
	}
	return max
}
func (p *Graph) choiceLevels(roots []Edge) {
	state := make([]uint8, len(p.nodes))
	var visit func(Edge)
	visit = func(e Edge) {
		id := e.NodeID()
		if state[id] == 2 {
			return
		}
		if state[id] == 1 {
			panic("aig: cyclic choice/box dependency")
		}
		state[id] = 1
		n := p.nodes[id]
		level := 0
		if n.Kind == Input {
			if manager, ok := p.Timing.(*tim.Tim_Man_t); ok {
				box := tim.Tim_ManBoxForCi(manager, int(Aig_ObjCioId(p, e)))
				if box >= 0 {
					first, count := tim.Tim_ManBoxInputFirst(manager, box), tim.Tim_ManBoxInputNum(manager, box)
					level = 0
					for i := 0; i < count; i++ {
						if first+i >= len(p.outputs) {
							panic("aig: timing box terminal outside graph")
						}
						co := p.outputs[first+i]
						visit(co)
						if l := p.node(co).Level; l > level {
							level = l
						}
					}
					level++
				}
			}
		} else {
			for _, child := range nodeFanins(n) {
				visit(child)
				if l := p.node(child).Level; l > level {
					level = l
				}
			}
			if n.Kind == And || n.Kind == ExclusiveOr {
				level++
			}
		}
		if other, ok := p.equivs[id]; ok {
			visit(other)
			if l := p.node(other).Level; l > level {
				level = l
			}
		}
		p.nodes[id].Level = level
		state[id] = 2
	}
	for _, e := range roots {
		visit(e)
	}
}
func Aig_ConeMark_rec(p *Graph, e Edge) {
	for _, x := range p.dfs([]Edge{e}, false, false, false) {
		p.node(x).MarkA = true
	}
}
func Aig_ConeCleanAndMark_rec(p *Graph, e Edge) {
	for _, x := range p.dfs([]Edge{e}, false, false, false) {
		p.node(x).MarkA = true
		delete(p.copies, x.NodeID())
	}
}
func Aig_ConeCountAndMark_rec(p *Graph, e Edge) int32 {
	var n int32
	for _, x := range p.dfs([]Edge{e}, true, false, false) {
		if !p.node(x).MarkA {
			n++
			p.node(x).MarkA = true
		}
	}
	return n
}
func Aig_ConeUnmark_rec(p *Graph, e Edge) {
	for _, x := range p.dfs([]Edge{e}, false, false, false) {
		p.node(x).MarkA = false
	}
}
func Aig_DagSize(p *Graph, e Edge) int32 { return int32(len(p.dfs([]Edge{e}, true, false, false))) }
func Aig_Support(p *Graph, e Edge) []Edge {
	var result []Edge
	for _, x := range p.dfs([]Edge{e}, false, false, false) {
		if p.node(x).Kind == Input {
			result = append(result, x)
		}
	}
	return result
}
func Aig_SupportSize(p *Graph, e Edge) int32 { return int32(len(Aig_Support(p, e))) }
func Aig_SupportSize_rec(p *Graph, e Edge, count *int32) {
	for _, x := range Aig_Support(p, e) {
		if p.node(x).TravID != p.travID {
			p.node(x).TravID = p.travID
			*count++
		}
	}
}
func Aig_SupportSizeTest(p *Graph) int32 {
	var total int32
	for _, e := range p.outputs {
		total += Aig_SupportSize(p, e)
	}
	return total
}
func Aig_Support_rec(p *Graph, e Edge, support *[]Edge) { Aig_SupportNodes(p, []Edge{e}, 1, support) }
func Aig_SupportNodes(p *Graph, roots []Edge, n int32, support *[]Edge) {
	if n < 0 || int(n) > len(roots) {
		panic("aig: invalid support range")
	}
	seen := make(map[int]bool)
	for _, e := range *support {
		seen[e.NodeID()] = true
	}
	for _, e := range p.dfs(roots[:n], false, false, false) {
		if p.node(e).Kind == Input && !seen[e.NodeID()] {
			*support = append(*support, e)
			seen[e.NodeID()] = true
		}
	}
}
func Aig_Transfer_rec(source, destination *Graph, root Edge) {
	Aig_ManDup_rec(destination, source, root)
}
func Aig_Transfer(source, destination *Graph, root Edge, nVars int32) Edge {
	if source == destination {
		return root
	}
	if nVars < 0 || int(nVars) > len(source.inputs) {
		panic("aig: invalid transfer variable count")
	}
	source.copies = map[int]Edge{0: 0}
	for i := int32(0); i < nVars; i++ {
		source.copies[source.inputs[i].NodeID()] = Aig_IthVar(destination, i)
	}
	return Aig_ManDup_rec(destination, source, root)
}
func Aig_Compose_rec(p *Graph, root, function, variable Edge) {
	index := -1
	for i, e := range p.inputs {
		if e == variable {
			index = i
			break
		}
	}
	if index < 0 {
		panic("aig: composition variable is not input")
	}
	result := Aig_Compose(p, root, function, int32(index))
	Aig_ObjSetCopy(p, root, result)
}
func Aig_Compose(p *Graph, root, function Edge, index int32) Edge {
	if index < 0 || int(index) >= len(p.inputs) {
		panic("aig: invalid composition input")
	}
	variable := p.inputs[index]
	memo := map[int]Edge{0: 0, variable.NodeID(): function}
	var visit func(Edge) Edge
	visit = func(e Edge) Edge {
		id := e.NodeID()
		value, ok := memo[id]
		if !ok {
			n := p.nodes[id]
			if n.Kind == Input {
				value = Edge(id) << 1
			} else if n.Kind == And {
				value = p.AND(visit(n.Fanins[0]), visit(n.Fanins[1]))
			} else {
				panic("aig: unsupported composition node")
			}
			memo[id] = value
		}
		if e.Complemented() {
			value = value.Not()
		}
		return value
	}
	return visit(root)
}
func Aig_ObjCollectCut_rec(p *Graph, root Edge, nodes *[]Edge) {
	if p.node(root).MarkA {
		return
	}
	p.node(root).MarkA = true
	n := p.nodes[root.NodeID()]
	if n.Kind != And && n.Kind != ExclusiveOr {
		panic("aig: incomplete cut")
	}
	Aig_ObjCollectCut_rec(p, n.Fanins[0], nodes)
	Aig_ObjCollectCut_rec(p, n.Fanins[1], nodes)
	*nodes = append(*nodes, Aig_Regular(root))
}
func Aig_ObjCollectCut(p *Graph, root Edge, leaves []Edge, nodes *[]Edge) {
	*nodes = (*nodes)[:0]
	for _, e := range leaves {
		if p.node(e).MarkA {
			panic("aig: marked cut leaf")
		}
		p.node(e).MarkA = true
	}
	defer func() {
		for _, e := range leaves {
			p.node(e).MarkA = false
		}
		for _, e := range *nodes {
			p.node(e).MarkA = false
		}
	}()
	Aig_ObjCollectCut_rec(p, root, nodes)
}
func Aig_ObjCollectSuper_rec(p *Graph, root, e Edge, super *[]Edge) int32 {
	for _, old := range *super {
		if old == e {
			return 1
		}
		if old == e.Not() {
			return -1
		}
	}
	n := p.node(e)
	if e != root && (e.Complemented() || n.Kind != p.node(root).Kind || n.Refs > 1) {
		*super = append(*super, e)
		return 0
	}
	if n.Kind != And && n.Kind != ExclusiveOr {
		panic(fmt.Sprintf("aig: invalid supergate node %d", e))
	}
	a := Aig_ObjCollectSuper_rec(p, root, p.real(n.Fanins[0]), super)
	b := Aig_ObjCollectSuper_rec(p, root, p.real(n.Fanins[1]), super)
	if a < 0 || b < 0 {
		return -1
	}
	if a != 0 || b != 0 {
		return 1
	}
	return 0
}
func Aig_ObjCollectSuper(p *Graph, root Edge, super *[]Edge) int32 {
	if root.Complemented() {
		panic("aig: complemented supergate root")
	}
	*super = (*super)[:0]
	result := Aig_ObjCollectSuper_rec(p, root, root, super)
	if result < 0 {
		*super = (*super)[:0]
	}
	return result
}

func (p *Graph) reverseDfs(roots, prior []Edge) []Edge {
	fanouts := make([][]Edge, len(p.nodes))
	for id, n := range p.nodes {
		if p.deleted[id] {
			continue
		}
		for _, e := range nodeFanins(n) {
			fanouts[e.NodeID()] = append(fanouts[e.NodeID()], Edge(id)<<1)
		}
	}
	result := append([]Edge(nil), prior...)
	state := make([]uint8, len(p.nodes))
	for _, e := range prior {
		state[e.NodeID()] = 2
	}
	var visit func(Edge)
	visit = func(e Edge) {
		id := e.NodeID()
		if state[id] == 2 || p.node(e).Kind == Output {
			return
		}
		if state[id] == 1 {
			panic("aig: cyclic reverse traversal")
		}
		state[id] = 1
		for _, fan := range fanouts[id] {
			visit(fan)
		}
		state[id] = 2
		result = append(result, Aig_Regular(e))
	}
	for _, e := range roots {
		visit(e)
	}
	return result
}
