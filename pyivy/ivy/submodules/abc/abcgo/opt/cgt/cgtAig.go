package cgt

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"slices"
	"sort"
)

func Cgt_ManDetectCandidates(g *aig.Graph, useful []bool, root aig.Edge, maxLevel int) []aig.Edge {
	kind := g.Node(root).Kind
	if kind != aig.And && kind != aig.ExclusiveOr {
		return nil
	}
	seen := make(map[int]bool)
	var candidates []aig.Edge
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		id := e.NodeID()
		if seen[id] {
			return
		}
		seen[id] = true
		for _, f := range fanins(g, e) {
			visit(f)
		}
		if g.Node(e).Level <= maxLevel && (useful == nil || id < len(useful) && useful[id]) {
			candidates = append(candidates, regular(e))
		}
	}
	visit(root)
	return candidates
}
func Cgt_ManDetectFanout(g *aig.Graph, root aig.Edge, maxLevel int) []aig.Edge {
	index := fanoutIndex(g)
	seen := make(map[int]bool)
	var visited []aig.Edge
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		n := g.Node(e)
		if n.Kind == aig.Output || n.Level > maxLevel || seen[e.NodeID()] {
			return
		}
		seen[e.NodeID()] = true
		visited = append(visited, regular(e))
		for _, out := range index[e.NodeID()] {
			visit(out)
		}
	}
	visit(root)
	var boundary []aig.Edge
	for _, e := range visited {
		for _, out := range index[e.NodeID()] {
			if !seen[out.NodeID()] {
				boundary = append(boundary, e)
				break
			}
		}
	}
	sort.Slice(boundary, func(i, j int) bool { return boundary[i] < boundary[j] })
	return boundary
}
func Cgt_ManCollectVisited(g *aig.Graph, roots []aig.Edge) []aig.Edge {
	seen := make(map[int]bool)
	var result []aig.Edge
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		if seen[e.NodeID()] {
			return
		}
		seen[e.NodeID()] = true
		n := g.Node(e)
		if n.Kind == aig.Input || n.Kind == aig.Constant {
			return
		}
		for _, f := range fanins(g, e) {
			visit(f)
		}
		result = append(result, regular(e))
	}
	for _, e := range roots {
		visit(e)
	}
	return result
}
func Cgt_ManConstructCareCondition(p *Manager, dst *aig.Graph, lo aig.Edge, copies0, copies1 map[int]aig.Edge) aig.Edge {
	boundary := Cgt_ManDetectFanout(p.Graph, lo, p.Params.ODCMax)
	result := dst.Const0()
	newInput := func(e aig.Edge) aig.Edge {
		if copy, ok := p.futureInputs[e.NodeID()]; ok {
			return copy
		}
		copy := dst.NewInput()
		p.futureInputs[e.NodeID()] = copy
		return copy
	}
	for _, e := range boundary {
		left := cloneEdge(dst, p.Graph, e, copies0, newInput)
		right := cloneEdge(dst, p.Graph, e, copies1, newInput)
		result = dst.OR(result, dst.XOR(left, right))
	}
	return result
}
func Cgt_ManDeriveAigForGating(p *Manager) *aig.Graph {
	src := p.Graph
	p.futureInputs = make(map[int]aig.Edge)
	dst := aig.NewGraph()
	dst.Name = "CG_miter"
	copies := map[int]aig.Edge{0: dst.Const1()}
	for _, e := range src.Inputs() {
		copies[e.NodeID()] = dst.NewInput()
	}
	newInput := func(e aig.Edge) aig.Edge { panic("cgt: first-frame input missing") }
	for _, e := range src.LiveEdges() {
		if src.Node(e).Kind != aig.Output {
			cloneEdge(dst, src, e, copies, newInput)
		}
	}
	regs, pis, pos := layout(src)
	inputs, outputs := src.Inputs(), src.Outputs()
	for r := 0; r < regs; r++ {
		lo := inputs[pis+r]
		next := cloneEdge(dst, src, driver(src, outputs[pos+r]), copies, newInput)
		miter := dst.XOR(copies[lo.NodeID()], next)
		if p.Params.ODCMax > 0 {
			a, b := map[int]aig.Edge{0: dst.Const1()}, map[int]aig.Edge{0: dst.Const1()}
			for j := 0; j < regs; j++ {
				other := inputs[pis+j]
				value := cloneEdge(dst, src, driver(src, outputs[pos+j]), copies, newInput)
				a[other.NodeID()], b[other.NodeID()] = value, value
			}
			a[lo.NodeID()], b[lo.NodeID()] = dst.Const0(), dst.Const1()
			care := Cgt_ManConstructCareCondition(p, dst, lo, a, b)
			miter = dst.AND(miter, care)
		}
		co := dst.AddOutput(miter)
		copies[outputs[pos+r].NodeID()] = co
	}
	p.frameCopies = copies
	p.Frame = dst
	return dst
}
func Cgt_ManConstructCare(dst, care *aig.Graph, sourceInputs []aig.Edge, inputCopies map[int]aig.Edge) int {
	if care == nil {
		return 0
	}
	mapping := map[int]aig.Edge{0: dst.Const1()}
	careInputs := care.Inputs()
	for i, source := range sourceInputs {
		if i >= len(careInputs) {
			break
		}
		if copy, ok := inputCopies[source.NodeID()]; ok {
			mapping[careInputs[i].NodeID()] = copy
		}
	}
	count := 0
	for _, out := range care.Outputs() {
		complete := true
		seen := make(map[int]bool)
		var support func(aig.Edge)
		support = func(e aig.Edge) {
			if seen[e.NodeID()] {
				return
			}
			seen[e.NodeID()] = true
			if care.Node(e).Kind == aig.Input {
				if _, ok := mapping[e.NodeID()]; !ok {
					complete = false
				}
			}
			for _, f := range fanins(care, e) {
				support(f)
			}
		}
		support(out)
		if !complete {
			continue
		}
		copy := cloneEdge(dst, care, driver(care, out), mapping, func(aig.Edge) aig.Edge { panic("cgt: missing care support") })
		if copy == dst.Const1() {
			continue
		}
		dst.AddOutput(copy)
		count++
	}
	return count
}
func cgtPartition(frame *aig.Graph, varsMin, flopsMin, start int, care *aig.Graph) (*aig.Graph, map[int]aig.Edge, int, int) {
	dst := aig.NewGraph()
	dst.Name = "partition"
	copies := map[int]aig.Edge{0: dst.Const1()}
	newInput := func(e aig.Edge) aig.Edge { return dst.NewInput() }
	var roots, sourceOutputs []aig.Edge
	outs := frame.Outputs()
	for i := start; i < len(outs) && (len(roots) < flopsMin || dst.NodeCount() < varsMin); i++ {
		roots = append(roots, cloneEdge(dst, frame, driver(frame, outs[i]), copies, newInput))
		sourceOutputs = append(sourceOutputs, outs[i])
	}
	constraints := Cgt_ManConstructCare(dst, care, frame.Inputs(), copies)
	for i, root := range roots {
		copies[sourceOutputs[i].NodeID()] = dst.AddOutput(root)
	}
	return dst, copies, len(roots), constraints
}
func Cgt_ManDupPartition(frame *aig.Graph, varsMin, flopsMin, start int, care *aig.Graph) (*aig.Graph, int, error) {
	if frame == nil || aig.Aig_ManRegNum(frame) != 0 || varsMin <= 0 || flopsMin <= 0 || start < 0 || start >= frame.OutputCount() {
		return nil, 0, fmt.Errorf("cgt: invalid partition request")
	}
	g, copies, n, _ := cgtPartition(frame, varsMin, flopsMin, start, care)
	for id, copy := range copies {
		aig.Aig_ObjSetCopy(frame, aig.Edge(id*2), copy)
	}
	return g, n, nil
}
func copyMetadata(dst, src *aig.Graph) {
	dst.Name, dst.Spec = src.Name, src.Spec
	dst.Assertions, dst.BarrierBuffers = src.Assertions, src.BarrierBuffers
	dst.FlopNumbers = slices.Clone(src.FlopNumbers)
	dst.FlopReprs = slices.Clone(src.FlopReprs)
	dst.ClockDomains = slices.Clone(src.ClockDomains)
	for i := range dst.ClockDomains {
		dst.ClockDomains[i] = slices.Clone(src.ClockDomains[i])
	}
	dst.PartitionInputs = slices.Clone(src.PartitionInputs)
	dst.PartitionOutputs = slices.Clone(src.PartitionOutputs)
	dst.RetimingChanged = src.RetimingChanged
	if src.Timing != nil {
		if clone, ok := src.Timing.(interface{ CloneTiming() any }); ok {
			dst.Timing = clone.CloneTiming()
		} else {
			panic("cgt: timing metadata is not cloneable")
		}
	}
	if src.History != nil {
		if clone, ok := src.History.(interface{ CloneHistory() any }); ok {
			dst.History = clone.CloneHistory()
		} else {
			panic("cgt: history metadata is not cloneable")
		}
	}
}
func Cgt_ManBuildClockGate(dst, src *aig.Graph, gates []aig.Edge, copies, preserved map[int]aig.Edge) aig.Edge {
	total := dst.Const0()
	for _, e := range gates {
		copy, ok := preserved[e.NodeID()]
		if !src.ValidEdge(e) {
			panic("cgt: invalid source gate")
		}
		if !ok {
			copy, ok = copies[e.NodeID()]
			if !ok {
				panic("cgt: source gate has no destination copy")
			}
		}
		total = dst.OR(total, copy^aig.Edge(e&1))
	}
	return total
}
func Cgt_ManDeriveGatedAig(src *aig.Graph, gates Gates, reduce bool) (*aig.Graph, int, error) {
	regs, pis, pos := layout(src)
	if len(gates) != regs {
		return nil, 0, fmt.Errorf("cgt: gate/register count differs")
	}
	dst := aig.NewGraph()
	copyMetadata(dst, src)
	copies := map[int]aig.Edge{0: dst.Const1()}
	preserved := make(map[int]aig.Edge)
	marked := make(map[int]uint8)
	for _, set := range gates {
		for _, e := range set {
			if !src.ValidEdge(e) {
				return nil, 0, fmt.Errorf("cgt: invalid gate literal")
			}
			marked[e.NodeID()] |= 1 << uint(e&1)
		}
	}
	for _, e := range src.Inputs() {
		copies[e.NodeID()] = dst.NewInput()
	}
	var clone func(aig.Edge) aig.Edge
	active := make(map[int]bool)
	clone = func(e aig.Edge) aig.Edge {
		if copy, ok := copies[e.NodeID()]; ok {
			return copy ^ aig.Edge(e&1)
		}
		if active[e.NodeID()] {
			panic("cgt: cyclic graph")
		}
		active[e.NodeID()] = true
		n := src.Node(e)
		var value aig.Edge
		switch n.Kind {
		case aig.And:
			value = dst.AND(clone(n.Fanins[0]), clone(n.Fanins[1]))
		case aig.ExclusiveOr:
			value = dst.NewXor(clone(n.Fanins[0]), clone(n.Fanins[1]))
		case aig.Buffer, aig.Output:
			value = clone(n.Fanins[0])
		default:
			panic("cgt: invalid copied gate")
		}
		delete(active, e.NodeID())
		copy := value
		if reduce && (n.Kind == aig.And || n.Kind == aig.ExclusiveOr) {
			if marked[e.NodeID()] == 3 {
				panic("cgt: area reduction selected both gate phases")
			}
			if marked[e.NodeID()] != 0 {
				preserved[e.NodeID()] = value
				copy = dst.Const0()
				if marked[e.NodeID()] == 2 {
					copy = dst.Const1()
				}
			}
		}
		copies[e.NodeID()] = copy
		return copy ^ aig.Edge(e&1)
	}
	for _, e := range src.LiveEdges() {
		clone(e)
	}
	used := dst.NodeCount() - dst.InputCount() - 1
	outputs, inputs := src.Outputs(), src.Inputs()
	for i := 0; i < pos; i++ {
		dst.AddOutput(clone(driver(src, outputs[i])))
	}
	for r := 0; r < regs; r++ {
		next := clone(driver(src, outputs[pos+r]))
		if len(gates[r]) > 0 {
			gate := Cgt_ManBuildClockGate(dst, src, gates[r], copies, preserved)
			next = dst.Mux(gate, copies[inputs[pis+r].NodeID()], next)
		}
		dst.AddOutput(next)
	}
	aig.Aig_ManSetRegNum(dst, int32(regs))
	aig.Aig_ManSetConstrNum(dst, aig.Aig_ManConstrNum(src))
	aig.Aig_ManCleanup(dst)
	return dst, used, nil
}
