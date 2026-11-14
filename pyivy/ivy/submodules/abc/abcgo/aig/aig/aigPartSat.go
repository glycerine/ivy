package aig

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/glycerine/abcgo/sat/satoko"
)

func satPartitionNode(kind NodeKind) bool {
	return kind == And || kind == ExclusiveOr || kind == Buffer
}
func Aig_ManPartitionMonolithic(g *Graph) []int32 { return make([]int32, len(g.nodes)) }
func Aig_ManPartitionLevelized(g *Graph, size int32) []int32 {
	if size <= 0 {
		panic("aig: partition size must be positive")
	}
	result := make([]int32, len(g.nodes))
	// C enumerates every live object in reverse level/ID order. Terminals
	// consume positions too, even though only gates are split into partitions.
	var ids []int
	for id := range g.nodes {
		if !g.deleted[id] {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := ids[i], ids[j]
		if g.nodes[a].Level != g.nodes[b].Level {
			return g.nodes[a].Level > g.nodes[b].Level
		}
		return a > b
	})
	for i, id := range ids {
		result[id] = int32(i) / size
	}
	return result
}
func Aig_ManPartitionDfs(g *Graph, size, preorder int32) []int32 {
	if size <= 0 {
		panic("aig: partition size must be positive")
	}
	// Use the source traversal setup and recursive helpers: a choice is visited
	// in postorder before ordinary fanins even during preorder partitioning.
	// Unlike the strict whole-manager DFS entry points, permit dangling gates
	// and unary buffers in this typed-graph extension.
	Aig_ManIncrementTravId(g)
	g.node(g.Const1()).TravID = g.travID
	for _, e := range g.inputs {
		g.node(e).TravID = g.travID
	}
	var visited, nodes []Edge
	for _, e := range g.outputs {
		root := Aig_Regular(g.outputDriver(e))
		if preorder != 0 {
			Aig_ManDfsPreorder_rec(g, root, &visited)
		} else {
			Aig_ManDfs_rec(g, root, &visited)
		}
	}
	for _, e := range visited {
		if satPartitionNode(g.node(e).Kind) {
			nodes = append(nodes, e)
		}
	}
	if preorder == 0 {
		for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}
	}
	result := make([]int32, len(g.nodes))
	for i, e := range nodes {
		result[e.NodeID()] = int32(i) / size
	}
	return result
}
func Aig_ManPartSplitOne_rec(destination, source *Graph, node Edge, pio *[]int32) {
	node = Aig_Regular(node)
	n := source.node(node)
	if n.Kind == Output {
		Aig_ManPartSplitOne_rec(destination, source, n.Fanins[0], pio)
		source.copies[node.NodeID()] = Aig_ObjRealCopy(source, n.Fanins[0])
		return
	}
	if n.TravID != source.travID {
		n.TravID = source.travID
		copy := destination.NewInput()
		destination.node(copy).MarkA = n.MarkA
		destination.node(copy).MarkB = n.MarkB
		source.copies[node.NodeID()] = copy
		*pio = append(*pio, int32(node.NodeID()))
		return
	}
	if _, exists := source.copies[node.NodeID()]; exists {
		return
	}
	Aig_ManPartSplitOne_rec(destination, source, n.Fanins[0], pio)
	a := Aig_ObjRealCopy(source, n.Fanins[0])
	var copy Edge
	if n.Kind == Buffer {
		copy = a
	} else {
		Aig_ManPartSplitOne_rec(destination, source, n.Fanins[1], pio)
		b := Aig_ObjRealCopy(source, n.Fanins[1])
		if n.Kind == ExclusiveOr {
			copy = destination.NewXor(a, b)
		} else {
			copy = destination.AND(a, b)
		}
	}
	source.copies[node.NodeID()] = copy
}
func Aig_ManPartSplitOne(source *Graph, nodes []Edge) (*Graph, []int32) {
	Aig_ManIncrementTravId(source)
	source.copies = map[int]Edge{}
	for _, node := range nodes {
		if node.Complemented() || !satPartitionNode(source.node(node).Kind) {
			panic("aig: invalid partition node")
		}
		if source.node(node).TravID == source.travID {
			panic("aig: duplicate partition node")
		}
		source.node(node).TravID = source.travID
	}
	destination := NewGraph()
	var pio []int32
	for _, node := range nodes {
		Aig_ManPartSplitOne_rec(destination, source, node, &pio)
	}
	for _, node := range nodes {
		copy := source.copies[node.NodeID()]
		if destination.node(copy).Refs < source.node(node).Refs {
			Aig_ObjCreateCo(destination, copy)
			pio = append(pio, int32(node.NodeID()))
		}
	}
	return destination, pio
}

// PIO mappings contain original object IDs. Output groups contain original
// driver literals, preserving output complement phases without tagged pointers.
func Aig_ManPartSplit(source *Graph, node2part []int32) (parts []*Graph, pio [][]int32, partOutputs [][]Edge) {
	if len(node2part) < len(source.nodes) {
		panic("aig: incomplete partition map")
	}
	groups := map[int32][]Edge{}
	for id, n := range source.nodes {
		if source.deleted[id] || !satPartitionNode(n.Kind) {
			continue
		}
		label := node2part[id]
		if label < 0 {
			panic("aig: negative partition label")
		}
		groups[label] = append(groups[label], Edge(id)<<1)
	}
	if len(groups) == 0 {
		part := NewGraph()
		source.copies = map[int]Edge{0: 0}
		for _, input := range source.inputs {
			source.copies[input.NodeID()] = part.NewInput()
		}
		var mapping []int32
		for _, input := range source.inputs {
			mapping = append(mapping, int32(input.NodeID()))
		}
		for _, output := range source.outputs {
			driver := satOutputDriver(source, output)
			part.AddOutput(dupEdge(part, source, Aig_Regular(driver)))
			mapping = append(mapping, int32(driver.NodeID()))
		}
		return []*Graph{part}, [][]int32{mapping}, [][]Edge{append([]Edge(nil), outputDrivers(source)...)}
	}
	// Retain pre-existing client marks rather than using persistent graph flags
	// as temporary partitioning state.
	type marks struct{ a, b bool }
	saved := make([]marks, len(source.nodes))
	for id := range source.nodes {
		n := &source.nodes[id]
		saved[id] = marks{n.MarkA, n.MarkB}
		n.MarkA, n.MarkB = false, false
	}
	defer func() {
		for id, m := range saved {
			source.nodes[id].MarkA, source.nodes[id].MarkB = m.a, m.b
		}
	}()
	for _, output := range source.outputs {
		driver := satOutputDriver(source, output)
		n := source.node(driver)
		if n.Kind != Input {
			continue
		}
		if driver.Complemented() {
			n.MarkA = true
		} else {
			n.MarkB = true
		}
	}
	labels := make([]int32, 0, len(groups))
	for label := range groups {
		labels = append(labels, label)
	}
	sort.Slice(labels, func(i, j int) bool { return labels[i] < labels[j] })
	compact := map[int32]int{}
	for _, label := range labels {
		part, mapping := Aig_ManPartSplitOne(source, groups[label])
		compact[label] = len(parts)
		parts = append(parts, part)
		pio = append(pio, mapping)
		partOutputs = append(partOutputs, nil)
	}
	for _, output := range source.outputs {
		driver := satOutputDriver(source, output)
		if !satPartitionNode(source.node(driver).Kind) {
			continue
		}
		part := compact[node2part[driver.NodeID()]]
		partOutputs[part] = append(partOutputs[part], driver)
	}
	return
}

// Output objects are transparent in the Boolean network, including COs that
// appear as fanins after graph transformations.
func satOutputDriver(g *Graph, edge Edge) Edge {
	for g.node(edge).Kind == Output {
		edge = dupPhase(g.node(edge).Fanins[0], edge.Complemented())
	}
	return edge
}
func outputDrivers(g *Graph) []Edge {
	result := make([]Edge, len(g.outputs))
	for i, output := range g.outputs {
		result[i] = satOutputDriver(g, output)
	}
	return result
}
func Aig_ManPartResetNodePolarity(part *Graph) {
	for id := range part.nodes {
		part.nodes[id].Phase = false
	}
}
func Aig_ManPartSetNodePolarity(source, part *Graph, pio []int32) {
	if len(pio) != len(part.inputs)+len(part.outputs) {
		panic("aig: invalid PIO mapping")
	}
	part.nodes[0].Phase = true
	for i, input := range part.inputs {
		part.node(input).Phase = source.nodes[pio[i]].Phase
	}
	state := make([]uint8, len(part.nodes))
	state[0] = 2
	for _, input := range part.inputs {
		state[input.NodeID()] = 2
	}
	var phase func(Edge) bool
	phase = func(edge Edge) bool {
		id := edge.NodeID()
		if state[id] == 1 {
			panic("aig: cyclic polarity graph")
		}
		if state[id] != 2 {
			state[id] = 1
			n := part.node(edge)
			switch n.Kind {
			case And:
				n.Phase = phase(n.Fanins[0]) && phase(n.Fanins[1])
			case ExclusiveOr:
				n.Phase = phase(n.Fanins[0]) != phase(n.Fanins[1])
			case Buffer, Output:
				n.Phase = phase(n.Fanins[0])
			default:
				panic("aig: unknown polarity node")
			}
			state[id] = 2
		}
		return part.nodes[id].Phase != edge.Complemented()
	}
	for id := range part.nodes {
		if !part.deleted[id] {
			phase(Edge(id) << 1)
		}
	}
	for i, output := range part.outputs {
		if phase(output) != source.nodes[pio[len(part.inputs)+i]].Phase {
			panic("aig: partition output phase mismatch")
		}
	}
}

// NodeVariables records the phase of each shared SAT variable as well as its
// number. ID zero means unassigned; SAT variable zero is deliberately unused.
type NodeVariables struct {
	IDs    []int32
	Phases []bool
}

func NewNodeVariables(nodes int) *NodeVariables {
	return &NodeVariables{make([]int32, nodes), make([]bool, nodes)}
}
func newPartVariable(solver *satoko.Solver) int32 {
	if solver.VariableCount() == 0 {
		solver.SetVariables(1)
	}
	return int32(solver.AddVariable(false))
}
func sharedLiteral(variables *NodeVariables, id int) int {
	lit := int(variables.IDs[id]) << 1
	if variables.Phases[id] {
		lit ^= 1
	}
	return lit
}
func establishConstant(solver *satoko.Solver, variables *NodeVariables) bool {
	if len(variables.IDs) == 0 || len(variables.Phases) != len(variables.IDs) {
		panic("aig: invalid shared variable storage")
	}
	if variables.IDs[0] == 0 {
		variables.IDs[0] = newPartVariable(solver)
	}
	return solver.AddClause([]int{sharedLiteral(variables, 0)})
}
func Aig_ManDeriveCounterExample(g *Graph, variables *NodeVariables, solver *satoko.Solver) {
	model := make([]int32, len(g.inputs))
	for i, input := range g.inputs {
		id := input.NodeID()
		if variables.IDs[id] == 0 {
			continue
		}
		value := solver.ModelValue(int(variables.IDs[id]))
		if value < 0 {
			value = 0
		}
		if variables.Phases[id] {
			value ^= 1
		}
		model[i] = int32(value)
	}
	g.Data = model
}
func Aig_ManAddNewCnfToSolver(solver *satoko.Solver, part *Graph, variables *NodeVariables, pio []int32, outputs []Edge, _ int32) int32 {
	if len(pio) != len(part.inputs)+len(part.outputs) {
		panic("aig: incomplete CNF terminal mapping")
	}
	if !establishConstant(solver, variables) {
		return 1
	}
	local := make([]int32, len(part.nodes))
	for id := range part.nodes {
		if !part.deleted[id] {
			local[id] = newPartVariable(solver)
		}
	}
	literal := func(edge Edge) int {
		lit := int(local[edge.NodeID()]) << 1
		if edge.Complemented() != part.node(edge).Phase {
			lit ^= 1
		}
		return lit
	}
	add := func(clause ...int) bool { return solver.AddClause(clause) }
	if !add(literal(part.Const1())) {
		return 1
	}
	for id, n := range part.nodes {
		if part.deleted[id] {
			continue
		}
		z := literal(Edge(id) << 1)
		switch n.Kind {
		case And:
			a, b := literal(n.Fanins[0]), literal(n.Fanins[1])
			if !add(z^1, a) || !add(z^1, b) || !add(z, a^1, b^1) {
				return 1
			}
		case ExclusiveOr:
			a, b := literal(n.Fanins[0]), literal(n.Fanins[1])
			if !add(a^1, b^1, z^1) || !add(a, b, z^1) || !add(a, b^1, z) || !add(a^1, b, z) {
				return 1
			}
		case Buffer, Output:
			a := literal(n.Fanins[0])
			if !add(z^1, a) || !add(z, a^1) {
				return 1
			}
		}
	}
	terminals := append(append([]Edge(nil), part.inputs...), part.outputs...)
	for i, terminal := range terminals {
		id := int(pio[i])
		if id < 0 || id >= len(variables.IDs) {
			panic("aig: terminal mapping outside source")
		}
		if variables.IDs[id] == 0 {
			variables.IDs[id] = local[terminal.NodeID()]
			variables.Phases[id] = part.node(terminal).Phase
		} else {
			a, b := sharedLiteral(variables, id), literal(terminal)
			if !add(a^1, b) || !add(a, b^1) {
				return 1
			}
		}
	}
	for _, driver := range outputs {
		id := driver.NodeID()
		if id >= len(variables.IDs) || variables.IDs[id] == 0 {
			panic("aig: output has no shared SAT variable")
		}
		lit := sharedLiteral(variables, id)
		if driver.Complemented() {
			lit ^= 1
		}
		if !add(lit) {
			return 1
		}
	}
	for i, input := range part.inputs {
		n := part.node(input)
		id := int(pio[i])
		lit := sharedLiteral(variables, id)
		if n.MarkA && !add(lit^1) {
			return 1
		}
		if n.MarkB && !add(lit) {
			return 1
		}
		n.MarkA, n.MarkB = false, false
	}
	return 0
}

// PartSATSynthesizer is supplied from a higher layer (for example opt/dar),
// avoiding an aig-to-DAR-to-aig import cycle. Terminal order must be preserved.
type PartSATSynthesizer interface{ SynthesizePart(*Graph) (*Graph, error) }
type PartSATOptions struct {
	Algorithm                             int32
	PartitionSize                         int32
	PartConflictLimit, TotalConflictLimit int64
	AlignPolarity                         bool
	Synthesize                            bool
	Synthesizer                           PartSATSynthesizer
	Verbose                               bool
}

// PartitionedSAT returns 0 for SAT, 1 for UNSAT and -1 for an incomplete solve.
func PartitionedSAT(g *Graph, options PartSATOptions) (int32, error) {
	if g == nil {
		return -1, errors.New("aig: nil SAT graph")
	}
	if err := g.Check(); err != nil {
		return -1, err
	}
	if options.Synthesize && options.Synthesizer == nil {
		return -1, errors.New("aig: partition synthesis requested without a PartSATSynthesizer")
	}
	if options.PartConflictLimit < 0 || options.TotalConflictLimit < 0 {
		return -1, errors.New("aig: negative SAT conflict limit")
	}
	var labels []int32
	switch options.Algorithm {
	case 0:
		labels = Aig_ManPartitionMonolithic(g)
	case 1, 2, 3:
		if options.PartitionSize <= 0 {
			return -1, errors.New("aig: partition size must be positive")
		}
		if options.Algorithm == 1 {
			labels = Aig_ManPartitionLevelized(g, options.PartitionSize)
		} else {
			pre := int32(0)
			if options.Algorithm == 3 {
				pre = 1
			}
			labels = Aig_ManPartitionDfs(g, options.PartitionSize, pre)
		}
	default:
		return -1, errors.New("aig: unknown partitioning algorithm")
	}
	started := time.Now()
	parts, pio, outputs := Aig_ManPartSplit(g, labels)
	if options.Synthesize {
		for i, part := range parts {
			inputs, outputs := len(part.inputs), len(part.outputs)
			rewritten, err := options.Synthesizer.SynthesizePart(part)
			if err != nil {
				return -1, err
			}
			if rewritten == nil || len(rewritten.inputs) != inputs || len(rewritten.outputs) != outputs {
				return -1, errors.New("aig: synthesis changed partition terminal interface")
			}
			if err := rewritten.Check(); err != nil {
				return -1, err
			}
			parts[i] = rewritten
		}
	}
	solver := satoko.New()
	variables := NewNodeVariables(len(g.nodes))
	if !establishConstant(solver, variables) {
		return 1, nil
	}
	// Pre-map original CIs so isolated PI outputs are constrained and included in
	// the final counterexample even when no AND cone uses them.
	for _, input := range g.inputs {
		variables.IDs[input.NodeID()] = newPartVariable(solver)
		variables.Phases[input.NodeID()] = options.AlignPolarity && g.node(input).Phase
	}
	for _, output := range g.outputs {
		driver := satOutputDriver(g, output)
		if g.node(driver).Kind != Input && driver.NodeID() != 0 {
			continue
		}
		lit := sharedLiteral(variables, driver.NodeID())
		if driver.Complemented() {
			lit ^= 1
		}
		if !solver.AddClause([]int{lit}) {
			return 1, nil
		}
	}
	remaining := options.TotalConflictLimit
	finalStatus := int32(-1)
	for i, part := range parts {
		if options.AlignPolarity {
			Aig_ManPartSetNodePolarity(g, part, pio[i])
		} else {
			Aig_ManPartResetNodePolarity(part)
		}
		if Aig_ManAddNewCnfToSolver(solver, part, variables, pio[i], outputs[i], 0) != 0 {
			return 1, nil
		}
		budget := options.PartConflictLimit
		if options.TotalConflictLimit > 0 && (budget == 0 || remaining < budget) {
			budget = remaining
		}
		before := solver.Stats().AllConflicts
		status := solver.SolveAssumptionsLimit(nil, budget)
		used := solver.Stats().AllConflicts - before
		if options.Verbose {
			fmt.Printf("Partition %d/%d: vars=%d conflicts=%d elapsed=%s\n", i+1, len(parts), solver.VariableCount(), used, time.Since(started))
		}
		if status == satoko.SATOKO_UNSAT {
			return 1, nil
		}
		if status == satoko.SATOKO_SAT {
			finalStatus = 0
		} else {
			finalStatus = -1
		}
		if options.TotalConflictLimit > 0 {
			remaining -= used
			if remaining <= 0 && i+1 < len(parts) {
				return -1, nil
			}
		}
	}
	if finalStatus == 0 {
		Aig_ManDeriveCounterExample(g, variables, solver)
	}
	return finalStatus, nil
}
func Aig_ManPartitionedSat(g *Graph, algorithm, size, partLimit, totalLimit, align, synthesize, verbose int32) int32 {
	status, err := PartitionedSAT(g, PartSATOptions{Algorithm: algorithm, PartitionSize: size, PartConflictLimit: int64(partLimit), TotalConflictLimit: int64(totalLimit), AlignPolarity: align != 0, Synthesize: synthesize != 0, Verbose: verbose != 0})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return status
}
