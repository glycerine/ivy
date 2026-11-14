// Sequential remapping and cleanup, ported from ABC aigScl.c.
package aig

import (
	"fmt"
	"io"
	"os"
)

func sclLayout(p *Graph) (registers, pis, pos int) {
	registers = int(p.nRegs)
	if registers < 0 || registers > len(p.inputs) || registers > len(p.outputs) {
		panic("aig: invalid sequential terminal counts")
	}
	if p.FlopNumbers != nil && len(p.FlopNumbers) != registers {
		panic("aig: flop-number metadata must match register count")
	}
	for _, domain := range p.ClockDomains {
		for _, reg := range domain {
			if reg < 0 || int(reg) >= registers {
				panic("aig: clock-domain register out of bounds")
			}
		}
	}
	return registers, len(p.inputs) - registers, len(p.outputs) - registers
}

// Aig_ManRemap applies CI replacements simultaneously. Targets are original
// inputs or constants, not partially rewritten CI mappings. Dangling nodes are
// retained, as in ABC; sequential cleanup is a separate operation.
func Aig_ManRemap(p *Graph, mapping []Edge) *Graph {
	registers, pis, _ := sclLayout(p)
	if len(mapping) != len(p.inputs) {
		panic("aig: wrong sequential mapping width")
	}
	index := make(map[int]int, len(p.inputs))
	for i, input := range p.inputs {
		index[input.NodeID()] = i
	}
	if p.FlopReprs != nil && len(p.FlopNumbers) != registers {
		panic("aig: representative tracking requires flop numbers")
	}
	for i, target := range mapping {
		if !p.valid(target) || (p.node(target).Kind != Input && p.node(target).Kind != Constant) {
			panic("aig: sequential mapping must target an input or constant")
		}
		if p.FlopReprs == nil || i < pis || p.inputs[i] == target {
			continue
		}
		from := p.FlopNumbers[i-pis]
		if from < 0 {
			panic("aig: negative original flop number")
		}
		if p.node(target).Kind == Constant {
			if target != p.Const0() {
				panic("aig: tracked constant register must preserve zero initialization")
			}
		} else {
			targetIndex := index[target.NodeID()]
			if target.Complemented() || targetIndex < pis {
				panic("aig: tracked register mapping must target a regular register")
			}
			to := p.FlopNumbers[targetIndex-pis]
			if to < 0 || to == from {
				panic("aig: invalid original flop representative")
			}
		}
	}
	q := NewGraph()
	dupMetadata(q, p)
	original := map[int]Edge{0: q.Const1()}
	for _, input := range p.inputs {
		copy := q.NewInput()
		q.node(copy).Level = p.node(input).Level
		original[input.NodeID()] = copy
	}
	p.copies = map[int]Edge{0: q.Const1()}
	for i, input := range p.inputs {
		target := mapping[i]
		p.copies[input.NodeID()] = dupPhase(original[target.NodeID()], target.Complemented())
		if p.FlopReprs == nil || i < pis || input == target {
			continue
		}
		from, to := p.FlopNumbers[i-pis], int32(-1)
		if p.node(target).Kind != Constant {
			targetIndex := index[target.NodeID()]
			to = p.FlopNumbers[targetIndex-pis]
		}
		q.FlopReprs = append(q.FlopReprs, from, to)
	}
	active := make(map[int]bool)
	for id, n := range p.nodes {
		if !p.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			dupVisit(q, p, Edge(id)<<1, false, false, active)
		}
	}
	for _, output := range p.outputs {
		q.AddOutput(dupVisit(q, p, p.outputDriver(output), false, false, active))
	}
	Aig_ManSetRegNum(q, p.nRegs)
	if err := q.Check(); err != nil {
		panic(fmt.Sprintf("aig: invalid sequential remap: %v", err))
	}
	return q
}

// Aig_ManSeqCleanup_rec traverses combinational fanins and queues the next-state
// output whenever a register output is reached. A register boundary is crossed
// through the worklist, not by recursively following clock-cycle feedback.
func Aig_ManSeqCleanup_rec(p *Graph, e Edge, pending *[]Edge) {
	registers, pis, pos := sclLayout(p)
	regByID := make(map[int]int, registers)
	for i, input := range p.inputs[pis:] {
		regByID[input.NodeID()] = i
	}
	sclReach(p, e, pending, regByID, pos)
}
func sclReach(p *Graph, e Edge, pending *[]Edge, registers map[int]int, pos int) {
	n := p.node(e)
	if n.TravID == p.travID {
		return
	}
	n.TravID = p.travID
	if n.Kind == Constant {
		return
	}
	if n.Kind == Input {
		if reg, ok := registers[e.NodeID()]; ok {
			*pending = append(*pending, p.outputs[pos+reg])
		}
		return
	}
	for _, fanin := range nodeFanins(*n) {
		sclReach(p, fanin, pending, registers, pos)
	}
}
func sclLiveRegisters(p *Graph) []bool {
	registers, pis, pos := sclLayout(p)
	Aig_ManIncrementTravId(p)
	p.nodes[0].TravID = p.travID
	for _, input := range p.inputs[:pis] {
		p.node(input).TravID = p.travID
	}
	index := make(map[int]int, registers)
	for i, input := range p.inputs[pis:] {
		index[input.NodeID()] = i
	}
	pending := append([]Edge(nil), p.outputs[:pos]...)
	for i := 0; i < len(pending); i++ {
		sclReach(p, pending[i], &pending, index, pos)
	}
	live := make([]bool, registers)
	for i, input := range p.inputs[pis:] {
		live[i] = p.node(input).TravID == p.travID
		// C clears the LO's pNext/CioId union after using it for the LI link.
		p.node(input).CioID = 0
		delete(p.next, input.NodeID())
	}
	return live
}
func Aig_ManSeqCleanup(p *Graph) int32 {
	_, pis, pos := sclLayout(p)
	live := sclLiveRegisters(p)
	var removedInputs []Edge
	inputs := append([]Edge(nil), p.inputs[:pis]...)
	outputs := append([]Edge(nil), p.outputs[:pos]...)
	oldToNew := make([]int32, len(live))
	for i := range oldToNew {
		oldToNew[i] = -1
	}
	var numbers []int32
	if p.FlopNumbers != nil {
		numbers = make([]int32, 0, len(live))
	}
	for i, keep := range live {
		input, output := p.inputs[pis+i], p.outputs[pos+i]
		if keep {
			oldToNew[i] = int32(len(inputs) - pis)
			inputs = append(inputs, input)
			outputs = append(outputs, output)
			if p.FlopNumbers != nil {
				numbers = append(numbers, p.FlopNumbers[i])
			}
			continue
		}
		// Keep the LO addressable until recursive cleanup releases its pins.
		// C can dereference its arena object even after clearing its vObjs slot.
		removedInputs = append(removedInputs, input)
		if p.node(output).Kind == Output {
			Aig_ObjDisconnect(p, Aig_Regular(output))
			p.deleted[output.NodeID()] = true
		} else {
			p.node(output).Refs-- // Legacy implicit CO consumer.
		}
	}
	if p.ClockDomains != nil {
		domains := make([][]int32, len(p.ClockDomains))
		for i, domain := range p.ClockDomains {
			for _, reg := range domain {
				if mapped := oldToNew[reg]; mapped >= 0 {
					domains[i] = append(domains[i], mapped)
				}
			}
		}
		p.ClockDomains = domains
	}
	if p.OneHotGroups != nil {
		groups := make([][]int32, 0, len(p.OneHotGroups))
		for _, group := range p.OneHotGroups {
			var mapped []int32
			for _, reg := range group {
				if reg < 0 || int(reg) >= len(oldToNew) {
					panic("aig: invalid one-hot register")
				}
				if next := oldToNew[reg]; next >= 0 {
					mapped = append(mapped, next)
				}
			}
			if len(mapped) > 1 {
				groups = append(groups, mapped)
			}
		}
		p.OneHotGroups = groups
	}
	if p.OneHotInputGroups != nil {
		groups := make([][]int32, 0, len(p.OneHotInputGroups))
		for _, group := range p.OneHotInputGroups {
			var mapped []int32
			for _, input := range group {
				if input < 0 || int(input) >= pis+len(oldToNew) {
					panic("aig: invalid one-hot input")
				}
				if int(input) < pis {
					mapped = append(mapped, input)
					continue
				}
				if next := oldToNew[int(input)-pis]; next >= 0 {
					mapped = append(mapped, int32(pis)+next)
				}
			}
			if len(mapped) > 1 {
				groups = append(groups, mapped)
			}
		}
		p.OneHotInputGroups = groups
	}
	p.inputs, p.outputs = inputs, outputs
	p.nRegs = int32(len(inputs) - pis)
	p.nTruePis = int32(pis)
	p.nTruePos = int32(len(outputs)) - p.nRegs
	p.FlopNumbers = numbers
	Aig_ManSetCioIds(p)
	before := Aig_ManNodeNum(p)
	Aig_ManCleanup(p)
	sclReleaseRemovedInputBuffers(p, removedInputs)
	for _, input := range removedInputs {
		p.deleted[input.NodeID()] = true
	}
	return before - Aig_ManNodeNum(p)
}

// C can retain a standalone buffer whose fanin is a removed LO. Reclaim only
// those dead buffer cones so the Go graph has no edges into deleted terminals;
// independent standalone buffers retain the source cleanup behavior.
func sclReleaseRemovedInputBuffers(p *Graph, inputs []Edge) {
	removed := make(map[int]bool, len(inputs))
	for _, input := range inputs {
		if p.node(input).Refs != 0 {
			removed[input.NodeID()] = true
		}
	}
	if len(removed) == 0 {
		return
	}
	depends := map[int]bool{}
	var visit func(Edge) bool
	visit = func(e Edge) bool {
		id := e.NodeID()
		if removed[id] {
			return true
		}
		if value, known := depends[id]; known {
			return value
		}
		for _, fanin := range nodeFanins(*p.node(e)) {
			if visit(fanin) {
				depends[id] = true
				return true
			}
		}
		depends[id] = false
		return false
	}
	for id, node := range p.nodes {
		if !p.deleted[id] && node.Kind == Buffer && node.Refs == 0 && visit(Edge(id)<<1) {
			Aig_ObjDelete_rec(p, Edge(id)<<1, 1)
		}
	}
}
func Aig_ManSeqCleanupBasic(p *Graph) int32 {
	_, _, pos := sclLayout(p)
	live := sclLiveRegisters(p)
	for reg, keep := range live {
		if keep {
			continue
		}
		output := p.outputs[pos+reg]
		if p.node(output).Kind == Output {
			Aig_ObjPatchFanin0(p, Aig_Regular(output), dupPhase(p.Const0(), output.Complemented()))
		} else {
			p.node(output).Refs--
			p.outputs[pos+reg] = p.Const0()
			p.node(p.Const0()).Refs++
			if p.node(output).Refs == 0 && p.node(output).Kind != Input && p.node(output).Kind != Constant {
				Aig_ObjDelete_rec(p, Aig_Regular(output), 1)
			}
		}
	}
	return Aig_ManCleanup(p)
}
func Aig_ManCountMergeRegs(p *Graph) int32 {
	writeMergeRegisterStats(os.Stdout, p)
	return 0
}
func writeMergeRegisterStats(w io.Writer, p *Graph) {
	_, _, pos := sclLayout(p)
	drivers := make(map[int]bool)
	zero, one := 0, 0
	Aig_ManIncrementTravId(p)
	for _, output := range p.outputs[pos:] {
		driver := p.outputDriver(output)
		drivers[driver.NodeID()] = true
		p.node(driver).TravID = p.travID
		if driver == p.Const0() {
			zero++
		}
		if driver == p.Const1() {
			one++
		}
	}
	fmt.Fprintf(w, "Regs = %d. Fanins = %d. Const0 = %d. Const1 = %d.\n", p.nRegs, len(drivers), zero, one)
}
func Aig_ManReduceLachesCount(p *Graph) int32 {
	registers, _, pos := sclLayout(p)
	if registers == 0 {
		panic("aig: register reduction requires registers")
	}
	seen := make(map[Edge]bool)
	var count int32
	for _, output := range p.outputs[pos:] {
		driver := p.outputDriver(output)
		if seen[driver] {
			count++
		}
		seen[driver] = true
	}
	return count
}
func Aig_ManReduceLachesOnce(p *Graph) []Edge {
	_, pis, pos := sclLayout(p)
	mapping := append([]Edge(nil), p.inputs[:pis]...)
	representative := make(map[Edge]Edge)
	for i, output := range p.outputs[pos:] {
		driver := p.outputDriver(output)
		input := p.inputs[pis+i]
		if first, ok := representative[driver]; ok {
			mapping = append(mapping, first)
		} else {
			representative[driver] = input
			mapping = append(mapping, input)
		}
	}
	return mapping
}
func Aig_ManReduceLaches(p *Graph, verbose int32) *Graph {
	if verbose != 0 {
		fmt.Println("Performing combinational register sweep:")
	}
	for p.nRegs > 0 && Aig_ManReduceLachesCount(p) > 0 {
		next := Aig_ManRemap(p, Aig_ManReduceLachesOnce(p))
		Aig_ManSeqCleanup(next)
		if verbose != 0 {
			Aig_ManReportImprovement(p, next)
		}
		if next.nRegs >= p.nRegs {
			panic("aig: register equality sweep made no progress")
		}
		Aig_ManStop(p)
		p = next
	}
	return p
}

// RegisterComponents matches the source's undirected register-dependency
// grouping. Despite its historical SCC name, it traverses both forward and
// reverse edges and therefore computes weakly connected components.
func RegisterComponents(p *Graph) [][]int32 {
	registers, _, _ := sclLayout(p)
	neighbors := make([][]int, registers)
	for output, support := range Aig_ManSupportsRegisters(p) {
		for _, input := range support {
			source := int(input)
			neighbors[output] = append(neighbors[output], source)
			neighbors[source] = append(neighbors[source], output)
		}
	}
	seen := make([]bool, registers)
	var result [][]int32
	for start := 0; start < registers; start++ {
		if seen[start] {
			continue
		}
		seen[start] = true
		component := []int32{int32(start)}
		for i := 0; i < len(component); i++ {
			for _, next := range neighbors[component[i]] {
				if !seen[next] {
					seen[next] = true
					component = append(component, int32(next))
				}
			}
		}
		result = append(result, component)
	}
	return result
}
func Aig_ManComputeSccs(p *Graph) { writeRegisterComponents(os.Stdout, p) }
func writeRegisterComponents(w io.Writer, p *Graph) {
	if p.nRegs == 0 {
		fmt.Fprintln(w, "The network is combinational.")
		return
	}
	components := RegisterComponents(p)
	for i, component := range components {
		if len(component) == int(p.nRegs) {
			fmt.Fprintln(w, "There is only one SCC of registers in this network.")
			break
		}
		fmt.Fprintf(w, "SCC #%d contains %5d registers.\n", i+1, len(component))
	}
}

// Aig_ManSclPart sweeps clock-domain partitions and transfers discovered classes
// to the original graph before rebuilding it through those representatives.
func Aig_ManSclPart(p *Graph, latchConst, latchEqual, verbose int32) *Graph {
	sclLayout(p)
	var partitions [][]int32
	if len(p.ClockDomains) > 0 {
		for _, domain := range p.ClockDomains {
			partitions = append(partitions, append([]int32(nil), domain...))
		}
	} else if p.nRegs > 0 {
		// The C caller passed a zero partition size, contradicting the partitioner's
		// positive-size contract. Its intended default is one complete partition.
		partitions = Aig_ManRegPartitionSimple(p, p.nRegs, 0)
	}
	Aig_ManReprStart(p, int32(len(p.nodes)))
	for i, partition := range partitions {
		if len(partition) == 0 {
			continue
		}
		var countPis, countRegs int32
		var mapBack []int32
		part := Aig_ManRegCreatePart(p, partition, &countPis, &countRegs, &mapBack)
		if countPis > 0 {
			reduced := Aig_ManScl(part, latchConst, latchEqual, 0, -1, -1, verbose, 0)
			classes := Aig_TransferMappedClasses(p, part, mapBack)
			if verbose != 0 {
				fmt.Printf("%3d : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d. It = %3d. Cl = %5d\n", i, len(partition), len(part.inputs)-len(partition), countPis, countRegs, Aig_ManNodeNum(part), 0, classes)
			}
			Aig_ManStop(reduced)
		}
		Aig_ManStop(part)
	}
	result := Aig_ManDupRepr(p, 0)
	Aig_ManSeqCleanup(result)
	return result
}

// Aig_ManScl preserves zero-initialized sequential behavior while removing
// unreachable registers, proven constants and equal next-state functions. The
// optional simulator overrides the built-in multivalued engine for callers
// supplying a more specialized implementation.
func Aig_ManScl(initial *Graph, latchConst, latchEqual, useMV, framesSymbolic, framesSatur, verbose, veryVerbose int32, engines ...MappingSimulator) *Graph {
	registers, pis, _ := sclLayout(initial)
	if len(initial.ClockDomains) > 0 {
		return Aig_ManSclPart(initial, latchConst, latchEqual, verbose)
	}
	if initial.FlopNumbers != nil {
		panic("aig: sequential sweep requires untracked original flop numbers")
	}
	p := Aig_ManDupSimple(initial)
	p.FlopNumbers = make([]int32, registers)
	for i := range p.FlopNumbers {
		p.FlopNumbers[i] = int32(i)
	}
	p.FlopReprs = make([]int32, 0)
	Aig_ManSeqCleanup(p)
	if latchConst != 0 && p.nRegs > 0 {
		p = Aig_ManConstReduce(p, useMV, framesSymbolic, framesSatur, verbose, veryVerbose, engines...)
	}
	if latchEqual != 0 && p.nRegs > 0 {
		p = Aig_ManReduceLaches(p, verbose)
	}
	Aig_ManReprStart(initial, int32(len(initial.nodes)))
	if len(p.FlopReprs)%2 != 0 {
		panic("aig: malformed flop representative pairs")
	}
	for i := 0; i < len(p.FlopReprs); i += 2 {
		from, to := p.FlopReprs[i], p.FlopReprs[i+1]
		if from < 0 || int(from) >= registers || to < -1 || int(to) >= registers {
			panic("aig: flop representative outside original register range")
		}
		a := initial.inputs[pis+int(from)]
		b := initial.Const1()
		if to != -1 {
			b = initial.inputs[pis+int(to)]
		}
		if a == b {
			panic("aig: self-representing register merge")
		}
		Aig_ObjSetRepr_(initial, a, b)
	}
	Aig_ManStop(p)
	result := Aig_ManDupRepr(initial, 0)
	Aig_ManSeqCleanup(result)
	return result
}
