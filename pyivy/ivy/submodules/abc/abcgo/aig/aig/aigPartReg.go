// Register partitioning, ported from ABC aigPartReg.c (Alan Mishchenko).
package aig

import (
	"fmt"
	"sort"
)

// Aig_ManPre_t owns the support-cost heuristic's working set. Slice lengths
// carry all dynamic counts; bit sets distinguish global and current membership.
type Aig_ManPre_t struct {
	maxSize            int
	supports           [][]int32
	used, inPart       []bool
	regs, unique, free []int32
}

// Cache each combinational cone once; shared next-state logic must not be
// traversed independently for every register. Returned supports are immutable.
func registerSupportMatrix(p *Graph) [][]int32 {
	n := int(p.nRegs)
	loIndex := map[int]int32{}
	for i, e := range p.inputs[len(p.inputs)-n:] {
		loIndex[e.NodeID()] = int32(i)
	}
	cached, active := map[int][]int32{}, map[int]bool{}
	var support func(Edge) []int32
	support = func(e Edge) []int32 {
		if !p.valid(e) {
			panic("aig: invalid register support edge")
		}
		id := e.NodeID()
		if result, ok := cached[id]; ok {
			return result
		}
		if active[id] {
			panic("aig: cyclic register logic")
		}
		if reg, ok := loIndex[id]; ok {
			cached[id] = []int32{reg}
			return cached[id]
		}
		active[id] = true
		found := map[int32]bool{}
		for _, fanin := range nodeFanins(*p.node(e)) {
			for _, reg := range support(fanin) {
				found[reg] = true
			}
		}
		var result []int32
		for reg := range found {
			result = append(result, reg)
		}
		sort.Slice(result, func(a, b int) bool { return result[a] < result[b] })
		delete(active, id)
		cached[id] = result
		return result
	}
	result := make([][]int32, n)
	for i, e := range p.outputs[len(p.outputs)-n:] {
		result[i] = support(e)
	}
	return result
}
func Aig_ManRegManStart(p *Graph, size int32) *Aig_ManPre_t {
	if size <= 0 {
		panic("aig: partition size must be positive")
	}
	if p.nRegs < 0 || int(p.nRegs) > len(p.inputs) || int(p.nRegs) > len(p.outputs) {
		panic("aig: invalid register interface")
	}
	return &Aig_ManPre_t{maxSize: int(size), supports: registerSupportMatrix(p), used: make([]bool, p.nRegs), inPart: make([]bool, p.nRegs)}
}
func Aig_ManRegManStop(p *Aig_ManPre_t) { *p = Aig_ManPre_t{} }
func Aig_ManRegFindSeed(p *Aig_ManPre_t) int32 {
	best, count := int32(-1), -1
	for i, support := range p.supports {
		if p.used[i] {
			continue
		}
		n := 0
		for _, reg := range support {
			if !p.used[reg] {
				n++
			}
		}
		if n > count {
			best, count = int32(i), n
		}
	}
	return best
}
func Aig_ManRegFindBestVar(p *Aig_ManPre_t) int32 {
	best, bestCost := int32(-1), int(^uint(0)>>1)
	for _, candidate := range p.free {
		cost := 0
		for _, reg := range p.supports[candidate] {
			if !p.inPart[reg] {
				cost++
				if p.used[reg] {
					cost += 3
				}
			}
		}
		if cost == 0 {
			return candidate
		}
		if cost < bestCost {
			best, bestCost = candidate, cost
		}
	}
	return best
}
func Aig_ManRegPartitionAdd(p *Aig_ManPre_t, reg int32) {
	if reg < 0 || int(reg) >= len(p.supports) {
		panic("aig: invalid register partition member")
	}
	for _, old := range p.regs {
		if old == reg {
			panic("aig: duplicate partition member")
		}
	}
	if !p.used[reg] {
		p.used[reg] = true
		p.unique = append(p.unique, reg)
	}
	if len(p.free) > 0 {
		found := -1
		for i, candidate := range p.free {
			if candidate == reg {
				found = i
				break
			}
		}
		if found < 0 {
			panic("aig: register is not a frontier variable")
		}
		p.free = append(p.free[:found], p.free[found+1:]...)
	} else if p.inPart[reg] {
		panic("aig: repeated partition seed")
	}
	p.inPart[reg] = true
	p.regs = append(p.regs, reg)
	for _, dep := range p.supports[reg] {
		if !p.inPart[dep] {
			p.inPart[dep] = true
			p.free = append(p.free, dep)
		}
	}
}
func Aig_ManRegPartitionSmart(p *Graph, size int32) [][]int32 {
	manager := Aig_ManRegManStart(p, size)
	defer Aig_ManRegManStop(manager)
	var result [][]int32
	for seed := Aig_ManRegFindSeed(manager); seed >= 0; seed = Aig_ManRegFindSeed(manager) {
		manager.regs = nil
		manager.unique = nil
		manager.free = nil
		clear(manager.inPart)
		Aig_ManRegPartitionAdd(manager, seed)
		for len(manager.regs) < manager.maxSize {
			candidate := Aig_ManRegFindBestVar(manager)
			if candidate < 0 {
				break
			}
			Aig_ManRegPartitionAdd(manager, candidate)
			if len(manager.free) == 0 {
				break
			}
		}
		result = append(result, append([]int32(nil), manager.regs...))
		fmt.Printf("Part %3d SUMMARY: Free = %4d. Total = %4d. Ratio = %6.2f. Unique = %4d.\n", len(result)-1, len(manager.free), len(manager.regs), float64(len(manager.free))/float64(len(manager.regs)), len(manager.unique))
	}
	return result
}
func partitionDomain(domain []int32, size, overlap int32) [][]int32 {
	if size <= 0 || overlap < 0 {
		panic("aig: invalid partition size or overlap")
	}
	if overlap >= size {
		fmt.Printf("Overlap size (%d) is more or equal than the partition size (%d).\nAdjusting it to be equal to half of the partition size.\n", overlap, size)
		overlap = size / 2
	}
	var result [][]int32
	for start := 0; start < len(domain); start += int(size - overlap) {
		end := min(start+int(size), len(domain))
		if len(result) == 0 || end-start > int(overlap) {
			result = append(result, append([]int32(nil), domain[start:end]...))
		}
	}
	return result
}
func Aig_ManRegPartitionSimple(p *Graph, size, overlap int32) [][]int32 {
	domain := make([]int32, p.nRegs)
	for i := range domain {
		domain[i] = int32(i)
	}
	return partitionDomain(domain, size, overlap)
}
func Aig_ManPartDivide(result *[][]int32, domain []int32, size, overlap int32) {
	if size <= 0 || len(domain) <= int(size) {
		panic("aig: domain must exceed partition size")
	}
	*result = append(*result, partitionDomain(domain, size, overlap)...)
}

func Aig_ManRegCreatePart(p *Graph, part []int32, countPis, countRegs *int32, mapBack *[]int32) *Graph {
	nRegs := int(p.nRegs)
	ciOffset, coOffset := len(p.inputs)-nRegs, len(p.outputs)-nRegs
	selected := map[int]bool{}
	for _, index := range part {
		if index < 0 || int(index) >= nRegs || selected[int(index)] {
			panic("aig: invalid or duplicate partition register")
		}
		selected[int(index)] = true
	}
	roots := make([]Edge, len(part))
	for i, index := range part {
		roots[i] = p.outputDriver(p.outputs[coOffset+int(index)])
	}
	cone := Aig_ManDfsNodes(p, roots, int32(len(roots)))
	for _, index := range part {
		Aig_ObjSetTravIdPrevious(p, p.inputs[ciOffset+int(index)])
	}
	result := NewGraph()
	result.Name, result.Spec = p.Name, p.Spec
	p.copies = map[int]Edge{0: result.Const1()}
	pure, external := int32(0), int32(0)
	// External support is placed before selected register outputs. Fake outputs
	// observe every selected LO, preserving it during sequential cleanup.
	for i, input := range p.inputs {
		if p.node(input).TravID != p.travID {
			continue
		}
		p.copies[input.NodeID()] = result.NewInput()
		p.node(input).TravID = p.travID
		if i < ciOffset {
			pure++
		} else {
			external++
		}
	}
	for _, index := range part {
		input := p.inputs[ciOffset+int(index)]
		mapped := result.NewInput()
		p.copies[input.NodeID()] = mapped
		p.node(input).TravID = p.travID
		result.AddOutput(mapped)
	}
	for _, e := range cone {
		switch p.node(e).Kind {
		case And, ExclusiveOr, Buffer:
			dupVisit(result, p, e, false, false, map[int]bool{})
		}
	}
	for _, index := range part {
		root := p.outputDriver(p.outputs[coOffset+int(index)])
		result.AddOutput(dupVisit(result, p, root, false, false, map[int]bool{}))
	}
	result.nRegs = int32(len(part))
	if countPis != nil {
		*countPis = pure
	}
	if countRegs != nil {
		*countRegs = external
	}
	if mapBack != nil {
		mapping := make([]int32, len(result.nodes))
		for i := range mapping {
			mapping[i] = -1
		}
		mapping[0] = 0
		for _, e := range cone {
			mapping[p.copies[e.NodeID()].NodeID()] = int32(e.NodeID())
		}
		for _, index := range part {
			e := p.inputs[ciOffset+int(index)]
			mapping[p.copies[e.NodeID()].NodeID()] = int32(e.NodeID())
		}
		*mapBack = mapping
	}
	return result
}
func Aig_ManRegProjectOnehots(p, part *Graph, groups [][]int32, verbose int32) [][]int32 {
	inputIndex := map[int]int32{}
	for i, e := range part.inputs {
		inputIndex[e.NodeID()] = int32(i)
	}
	base := len(p.inputs) - int(p.nRegs)
	var result [][]int32
	for _, group := range groups {
		var projected []int32
		for _, reg := range group {
			if reg < 0 || reg >= p.nRegs {
				panic("aig: invalid onehot register")
			}
			source := p.inputs[base+int(reg)]
			if p.node(source).TravID != p.travID {
				continue
			}
			if mapped, ok := p.copies[source.NodeID()]; ok {
				if index, ok := inputIndex[mapped.NodeID()]; ok {
					projected = append(projected, index)
				}
			}
		}
		if len(projected) > 1 {
			result = append(result, projected)
		}
	}
	if verbose != 0 && len(result) > 0 {
		fmt.Printf("Partition contains %d groups of 1-hot registers\n", len(result))
	}
	return result
}
func Aig_ManRegPartitionTraverse_rec(p *Graph, edge Edge, registers *[]Edge) {
	if !p.valid(edge) {
		panic("aig: invalid traversal edge")
	}
	node := p.node(edge)
	if node.TravID == p.travID {
		return
	}
	node.TravID = p.travID
	if node.Kind == Input {
		for _, lo := range p.inputs[len(p.inputs)-int(p.nRegs):] {
			if lo.NodeID() == edge.NodeID() {
				*registers = append(*registers, edge&^1)
				break
			}
		}
		return
	}
	for _, fanin := range nodeFanins(*node) {
		Aig_ManRegPartitionTraverse_rec(p, fanin, registers)
	}
}
func Aig_ManRegPartitionTraverse(p *Graph) []Edge {
	Aig_ManIncrementTravId(p)
	var result []Edge
	for _, li := range p.outputs[len(p.outputs)-int(p.nRegs):] {
		Aig_ManRegPartitionTraverse_rec(p, li, &result)
	}
	return result
}

// Aig_ManRegPartitionLinear is unfinished in C: it traverses the registers,
// ignores the requested size and returns NULL rather than forming partitions.
func Aig_ManRegPartitionLinear(p *Graph, size int32) [][]int32 {
	Aig_ManRegPartitionTraverse(p)
	return nil
}
