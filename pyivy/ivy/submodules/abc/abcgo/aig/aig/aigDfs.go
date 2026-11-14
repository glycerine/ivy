package aig

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
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
	Aig_ManIncrementTravId(p)
	for _, e := range p.outputs {
		p.node(e).TravID = p.travID
	}
	var nodes []Edge
	active := make(map[int]bool)
	for id, n := range p.nodes {
		if !p.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			p.reverseDfsRec(Edge(id)<<1, &nodes, active)
		}
	}
	return nodes
}
func Aig_ManDfsReverse_rec(p *Graph, e Edge, nodes *[]Edge) {
	p.reverseDfsRec(e, nodes, make(map[int]bool))
}
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
func Aig_ManLevelNum(p *Graph) int32          { return Aig_ManLevels(p) }
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
	p.updateConeMarks(e, true, false)
}
func Aig_ConeCleanAndMark_rec(p *Graph, e Edge) {
	p.updateConeMarks(e, true, true)
}
func Aig_ConeCountAndMark_rec(p *Graph, e Edge) int32 {
	return p.updateConeMarks(e, true, false)
}
func Aig_ConeUnmark_rec(p *Graph, e Edge) {
	p.updateConeMarks(e, false, false)
}

// All four C cone routines stop at non-logic objects and gates already in
// the requested mark state. Update in postorder, optionally clearing pData
// (represented by the copy map and auxiliary payload map in Go).
func (p *Graph) updateConeMarks(root Edge, marked, clean bool) int32 {
	active := make(map[int]bool)
	var visit func(Edge) int32
	visit = func(e Edge) int32 {
		if e.Complemented() {
			panic("aig: cone marking requires a regular object")
		}
		n := p.node(e)
		if n.Kind != And && n.Kind != ExclusiveOr || n.MarkA == marked {
			return 0
		}
		if active[e.NodeID()] {
			panic("aig: cyclic cone")
		}
		active[e.NodeID()] = true
		count := 1 + visit(Aig_Regular(n.Fanins[0])) + visit(Aig_Regular(n.Fanins[1]))
		delete(active, e.NodeID())
		n.MarkA = marked
		if clean {
			delete(p.copies, e.NodeID())
			delete(p.objectData, e.NodeID())
		}
		return count
	}
	return visit(root)
}
func Aig_DagSize(p *Graph, e Edge) int32 {
	root := Aig_Regular(e)
	count := Aig_ConeCountAndMark_rec(p, root)
	Aig_ConeUnmark_rec(p, root)
	return count
}
func Aig_Support(p *Graph, e Edge) []Edge {
	if e.Complemented() || p.node(e).Kind == Output {
		panic("aig: support requires a regular non-output root")
	}
	Aig_ManIncrementTravId(p)
	var result []Edge
	Aig_Support_rec(p, e, &result)
	return result
}
func Aig_SupportSize(p *Graph, e Edge) int32 {
	if e.Complemented() || p.node(e).Kind == Output {
		panic("aig: support size requires a regular non-output root")
	}
	Aig_ManIncrementTravId(p)
	var count int32
	Aig_SupportSize_rec(p, e, &count)
	return count
}
func Aig_SupportSize_rec(p *Graph, e Edge, count *int32) {
	n := p.node(e)
	if n.TravID == p.travID {
		return
	}
	n.TravID = p.travID
	if n.Kind == Input {
		*count++
		return
	}
	if n.Kind != And && n.Kind != ExclusiveOr && n.Kind != Buffer {
		panic("aig: support size recursion requires a CI, node or buffer")
	}
	for _, fanin := range nodeFanins(*n) {
		Aig_SupportSize_rec(p, Aig_Regular(fanin), count)
	}
}
func Aig_SupportSizeTest(p *Graph) int32 {
	started := util.Abc_Clock()
	var total int32
	for id, node := range p.nodes {
		if !p.deleted[id] && (node.Kind == And || node.Kind == ExclusiveOr) && Aig_SupportSize(p, Edge(id)<<1) <= 16 {
			total++
		}
	}
	fmt.Printf("Nodes with small support %d (out of %d)\n", total, Aig_ManNodeNum(p))
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-started)
	return total
}
func Aig_Support_rec(p *Graph, e Edge, support *[]Edge) {
	n := p.node(e)
	if n.TravID == p.travID {
		return
	}
	n.TravID = p.travID
	if n.Kind == Constant {
		return
	}
	if n.Kind == Input {
		*support = append(*support, Aig_Regular(e))
		return
	}
	if n.Kind != And && n.Kind != ExclusiveOr && n.Kind != Buffer {
		panic("aig: support recursion requires a CI, node or buffer")
	}
	for _, fanin := range nodeFanins(*n) {
		Aig_Support_rec(p, Aig_Regular(fanin), support)
	}
}
func Aig_SupportNodes(p *Graph, roots []Edge, n int32, support *[]Edge) {
	if n < 0 || int(n) > len(roots) {
		panic("aig: invalid support range")
	}
	*support = (*support)[:0]
	Aig_ManIncrementTravId(p)
	p.node(p.Const1()).TravID = p.travID
	for _, e := range roots[:n] {
		if e.Complemented() {
			panic("aig: support roots must be regular objects")
		}
		if p.node(e).Kind == Output {
			e = p.outputDriver(e)
		}
		Aig_Support_rec(p, Aig_Regular(e), support)
	}
}
func Aig_Transfer_rec(source, destination *Graph, root Edge) {
	source.transferCone(destination, root, make(map[int]bool))
}
func (source *Graph) transferCone(destination *Graph, root Edge, active map[int]bool) {
	if root.Complemented() {
		panic("aig: recursive transfer requires a regular object")
	}
	n := *source.node(root)
	if n.Kind != And && n.Kind != ExclusiveOr || n.MarkA {
		return
	}
	if active[root.NodeID()] {
		panic("aig: cyclic transfer cone")
	}
	active[root.NodeID()] = true
	source.transferCone(destination, Aig_Regular(n.Fanins[0]), active)
	source.transferCone(destination, Aig_Regular(n.Fanins[1]), active)
	a, b := Aig_ObjRealCopy(source, n.Fanins[0]), Aig_ObjRealCopy(source, n.Fanins[1])
	var copy Edge
	if n.Kind == ExclusiveOr {
		// Retain the existing Go graph extension for explicit XOR nodes.
		copy = destination.NewXor(a, b)
	} else {
		copy = destination.AND(a, b)
	}
	Aig_ObjSetCopy(source, root, copy)
	source.node(root).MarkA = true
	delete(active, root.NodeID())
}
func Aig_Transfer(source, destination *Graph, root Edge, nVars int32) Edge {
	if source == destination {
		return root
	}
	if root.NodeID() == 0 {
		return dupPhase(destination.Const1(), root.Complemented())
	}
	for i, input := range source.inputs {
		if int32(i) == nVars {
			break
		}
		Aig_ObjSetCopy(source, input, Aig_IthVar(destination, int32(i)))
	}
	Aig_Transfer_rec(source, destination, Aig_Regular(root))
	Aig_ConeUnmark_rec(source, Aig_Regular(root))
	return Aig_ObjRealCopy(source, root)
}
func Aig_Compose_rec(p *Graph, root, function, variable Edge) {
	p.composeCone(root, function, variable, make(map[int]bool))
}
func (p *Graph) composeCone(root, function, variable Edge, active map[int]bool) {
	if root.Complemented() {
		panic("aig: recursive composition requires a regular object")
	}
	// Copy the node value: building the replacement may grow p.nodes.
	n := *p.node(root)
	if n.MarkA {
		return
	}
	if n.Kind == Constant || n.Kind == Input {
		copy := root
		if root == variable {
			copy = function
		}
		Aig_ObjSetCopy(p, root, copy)
		return
	}
	if n.Kind != And {
		panic("aig: composition requires an AND-inverter cone")
	}
	if active[root.NodeID()] {
		panic("aig: cyclic composition cone")
	}
	active[root.NodeID()] = true
	p.composeCone(Aig_Regular(n.Fanins[0]), function, variable, active)
	p.composeCone(Aig_Regular(n.Fanins[1]), function, variable, active)
	copy := p.AND(Aig_ObjRealCopy(p, n.Fanins[0]), Aig_ObjRealCopy(p, n.Fanins[1]))
	Aig_ObjSetCopy(p, root, copy)
	p.node(root).MarkA = true
	delete(active, root.NodeID())
}
func Aig_Compose(p *Graph, root, function Edge, index int32) Edge {
	if int(index) >= len(p.inputs) {
		fmt.Printf("Aig_Compose(): The PI variable %d is not defined.\n", index)
		return InvalidEdge
	}
	if index < 0 {
		panic("aig: invalid composition input")
	}
	Aig_Compose_rec(p, Aig_Regular(root), function, p.inputs[index])
	Aig_ConeUnmark_rec(p, Aig_Regular(root))
	return Aig_ObjRealCopy(p, root)
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

func (p *Graph) reverseDfsRec(e Edge, nodes *[]Edge, active map[int]bool) {
	if e.Complemented() {
		panic("aig: reverse DFS requires a regular object")
	}
	n := p.node(e)
	if n.TravID == p.travID {
		return
	}
	if n.Kind != And && n.Kind != ExclusiveOr && n.Kind != Buffer {
		panic("aig: reverse DFS requires a node, buffer or premarked CO")
	}
	if active[e.NodeID()] {
		panic("aig: cyclic reverse traversal")
	}
	active[e.NodeID()] = true
	for i, fan := 0, Aig_ObjFanout0Int(p, int32(e.NodeID())); i < n.Refs; i, fan = i+1, Aig_ObjFanoutNext(p, fan) {
		p.reverseDfsRec(Aig_ManObj(p, fan>>1), nodes, active)
	}
	delete(active, e.NodeID())
	n.TravID = p.travID
	*nodes = append(*nodes, e)
}
