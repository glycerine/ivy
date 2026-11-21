package mf

import (
	"fmt"
	"math/bits"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Network follows GIA literals (id<<1|complement, zero=false). Gate kinds are
// 0 constant, 1 input, 2 AND, 3 XOR, 4 MUX(control,then,else), 5 buffer, 6 CO.
type Network interface {
	ObjectCount() int
	InputLiterals() []int32
	OutputLiterals() []int32
	Gate(id int32) (a, b, c int32, kind int)
}
type Options struct {
	LUTSize, CutLimit, Rounds, ExactRounds int
	ObjectIDs                              bool
}

func DefaultOptions() Options { return Options{LUTSize: 8, CutLimit: 8, Rounds: 2, ExactRounds: 1} }

type node struct {
	flow, flowRefs float32
	delay          int32
	refs           int32
}

// Mapping owns its cut/truth arrays. Cuts[id][0] is the selected source MF cut.
// The network must remain unchanged throughout generation and CNF derivation.
type Mapping struct {
	Network      Network
	Options      Options
	Cuts         [][]Cut
	Truths       [][]uint64
	ClauseSizes  []int
	Refs         []int32
	nodes        []node
	truthStore   *vec.Vec_Mem_t
	truthResults []truthResult
	phases       []bool
}

func (m *Mapping) intern(table []uint64, nLeaves int) int32 {
	id := int32(vec.Vec_MemHashInsert(m.truthStore, table))
	if int(id) < len(m.ClauseSizes) {
		return id
	}
	var cost int32
	switch {
	case nLeaves <= 6:
		cost = util.Abc_Tt6CnfSize(table[0], int32(nLeaves))
	case len(table) == 2:
		// The seven-input store uses two words. The wide CNF entry point takes four.
		padded := [4]uint64{table[0], table[1], table[0], table[1]}
		cost = util.Abc_Tt8CnfSize(padded[:], int32(nLeaves))
	default:
		cost = util.Abc_Tt8CnfSize(table, int32(nLeaves))
	}
	m.ClauseSizes = append(m.ClauseSizes, int(cost))
	return id
}
func (m *Mapping) area(cut Cut) int {
	if len(cut.Leaves) < 2 {
		return 0
	}
	return m.ClauseSizes[cut.Function>>1]
}
func (m *Mapping) prepare(id int32, unit bool, scratch []Cut, unitLeaf *[1]int32) []Cut {
	cuts := m.Cuts[id]
	if len(cuts) == 0 || unit && len(cuts[0].Leaves) > 1 {
		unitLeaf[0] = id
		cuts = append(append(scratch, cuts...), Cut{Leaves: unitLeaf[:], Function: 2})
	}
	return cuts
}
func (m *Mapping) parameters(cut *Cut, refs float32) {
	var delay int32
	var flow float32
	for _, leaf := range cut.Leaves {
		delay = max(delay, m.nodes[leaf].delay)
		flow += m.nodes[leaf].flow
	}
	if len(cut.Leaves) > 1 {
		delay++
	}
	cut.Delay = delay
	cut.Flow = (flow + float32(m.area(*cut))) / refs
}

// A truth operation depends on each signed function and on the positions its
// variables occupy in the merged cut. The numeric leaf IDs can differ on reuse.
type truthOperation struct {
	operands  [3]int32
	positions [3]uint8
	kind      uint8
}

type truthResult struct {
	operation truthOperation
	function  int32
	retained  uint8
}

func (m *Mapping) computeTruth(cuts []Cut, complements []int32, kind int, leaves []int32) Cut {
	operation := truthOperation{kind: uint8(kind)}
	hash := uint32(kind)
	for i, cut := range cuts {
		literal := cut.Function ^ complements[i]
		positions := leafPositions(cut.Leaves, leaves)
		operation.operands[i] = literal
		operation.positions[i] = positions
		hash = bits.RotateLeft32(hash, 7) ^ (uint32(literal)*0x9e3779b1 ^ uint32(positions)*0x85ebca6b)
	}
	entry := &m.truthResults[hash&uint32(len(m.truthResults)-1)]
	if entry.operation == operation {
		if entry.retained != 0xff {
			n := 0
			for positions := entry.retained; positions != 0; positions &= positions - 1 {
				leaves[n] = leaves[bits.TrailingZeros8(positions)]
				n++
			}
			leaves = leaves[:n]
		}
		return Cut{Leaves: leaves, Function: entry.function}
	}

	var original [maxLeaves]int32
	count := copy(original[:], leaves)
	var tables [3][4]uint64
	words := 1 << max(0, m.Options.LUTSize-6)
	for i, cut := range cuts {
		literal := operation.operands[i]
		source := vec.Vec_MemReadEntry(m.truthStore, int(literal>>1))
		expandTruth(tables[i][:words], source, len(cut.Leaves), operation.positions[i], m.Options.LUTSize, literal&1 != 0)
	}
	truth := tables[0][:words]
	for i := range truth {
		switch kind {
		case 3:
			truth[i] = tables[0][i] ^ tables[1][i]
		case 4:
			truth[i] = tables[2][i]&tables[1][i] | ^tables[2][i]&tables[0][i]
		default:
			truth[i] = tables[0][i] & tables[1][i]
		}
	}
	complement := int32(truth[0] & 1)
	if complement != 0 {
		for i := range truth {
			truth[i] = ^truth[i]
		}
	}
	// The >6-variable source path minimizes support even in object-ID mode.
	if m.Options.LUTSize > 6 || !m.Options.ObjectIDs {
		truth, leaves = minimizeSupport(truth, leaves, m.Options.LUTSize)
	}
	function := m.intern(truth, len(leaves))<<1 | complement
	retained := uint8(0xff)
	if len(leaves) != count {
		retained = leafPositions(leaves, original[:count])
	}
	*entry = truthResult{operation: operation, function: function, retained: retained}
	return Cut{Leaves: leaves, Function: function}
}
func (m *Mapping) merge(id int32, result *cutSet) error {
	a, b, c, kind := m.Network.Gate(id)
	if kind == 5 {
		b = a
		kind = 2
	}
	// The cut limit is at most 16, including the optional unit cut.
	var buffers [3][maxCuts]Cut
	var control []Cut
	if kind == 4 { // normalize control,then,else to source order.
		a, c = c, a
		control = m.prepare(c>>1, true, buffers[2][:0], &result.unitLeaves[2])
	}
	first := m.prepare(a>>1, true, buffers[0][:0], &result.unitLeaves[0])
	second := m.prepare(b>>1, true, buffers[1][:0], &result.unitLeaves[1])
	result.count = 0
	if choices, ok := m.Network.(interface{ Sibling(int32) int32 }); ok {
		if sibling := choices.Sibling(id); sibling != 0 {
			for _, candidate := range m.prepare(sibling, false, nil, &result.unitLeaves[3]) {
				if m.phases[id] != m.phases[sibling] {
					candidate.Function ^= 1
				}
				m.parameters(&candidate, m.nodes[id].flowRefs)
				result.add(candidate, signature(candidate.Leaves))
			}
		}
	}
	var pair, triple [maxLeaves]int32
	offer := func(parts []Cut, complements []int32, leaves []int32, sign uint64) {
		if result.contains(leaves, sign) {
			return
		}
		cut := m.computeTruth(parts, complements, kind, leaves)
		if len(cut.Leaves) < len(leaves) {
			sign = signature(cut.Leaves)
		}
		m.parameters(&cut, m.nodes[id].flowRefs)
		result.add(cut, sign)
	}
	var signs [3][maxCuts]uint64
	for part, fanin := range [3][]Cut{first, second, control} {
		for i, cut := range fanin {
			signs[part][i] = signature(cut.Leaves)
		}
	}
	for i, left := range first {
		for j, right := range second {
			sign := signs[0][i] | signs[1][j]
			if bits.OnesCount64(sign) > m.Options.LUTSize {
				continue
			}
			data, ok := mergeLeaves(&pair, left.Leaves, right.Leaves, m.Options.LUTSize)
			if !ok {
				continue
			}
			if kind != 4 {
				offer([]Cut{left, right}, []int32{a & 1, b & 1}, data, sign)
				continue
			}
			for k, ctrl := range control {
				all := sign | signs[2][k]
				if bits.OnesCount64(all) > m.Options.LUTSize {
					continue
				}
				// Truth minimization may edit this candidate's leaves. Keep the
				// shared data pair in its own buffer across control candidates.
				leaves, ok := mergeLeaves(&triple, data, ctrl.Leaves, m.Options.LUTSize)
				if ok {
					offer([]Cut{left, right, ctrl}, []int32{a & 1, b & 1, c & 1}, leaves, all)
				}
			}
		}
	}
	if result.count == 0 {
		return fmt.Errorf("mf: no cuts for object %d", id)
	}
	cuts := result.save()
	m.Cuts[id] = cuts
	m.nodes[id].flow = cuts[0].Flow
	m.nodes[id].delay = cuts[0].Delay
	return nil
}

func gateNode(kind int) bool { return kind >= 2 && kind <= 5 }
func (m *Mapping) setMapRefs(iter int, exact bool) {
	if !exact {
		for _, literal := range m.Network.OutputLiterals() {
			m.nodes[literal>>1].refs++
		}
	}
	for i := len(m.nodes) - 1; i >= 0; i-- {
		_, _, _, kind := m.Network.Gate(int32(i))
		if !gateNode(kind) || m.nodes[i].refs == 0 {
			continue
		}
		if !exact {
			for _, leaf := range m.Cuts[i][0].Leaves {
				m.nodes[leaf].refs++
			}
		}
	}
	coefficient := float32(1) / float32(1+(iter+1)*(iter+1))
	for i := range m.nodes {
		m.nodes[i].flowRefs = coefficient*m.nodes[i].flowRefs + (1-coefficient)*float32(max(1, m.nodes[i].refs))
	}
}
func (m *Mapping) changeRefs(cut Cut, delta int32, limit int, touched *[]int32) int {
	count := m.area(cut)
	if limit == 0 {
		return count
	}
	for _, leaf := range cut.Leaves {
		if touched != nil {
			*touched = append(*touched, leaf)
		}
		old := m.nodes[leaf].refs
		m.nodes[leaf].refs += delta
		recurse := delta > 0 && old == 0 || delta < 0 && m.nodes[leaf].refs == 0
		if recurse && len(m.Cuts[leaf]) != 0 {
			count += m.changeRefs(m.Cuts[leaf][0], delta, limit-1, touched)
		}
	}
	return count
}
func (m *Mapping) cutFlow(cut Cut) (float32, int32) {
	var flow float32
	var delay int32
	for _, leaf := range cut.Leaves {
		flow += m.nodes[leaf].flow
		delay = max(delay, m.nodes[leaf].delay)
	}
	return flow + float32(m.area(cut)), delay + 1
}
func (m *Mapping) bestCut(id int32, exact bool, touched []int32) []int32 {
	best := &m.nodes[id]
	if exact && best.refs != 0 {
		m.changeRefs(m.Cuts[id][0], -1, 1<<30, nil)
	}
	selected := -1
	var bestFlow float32
	var bestTime int32
	for i, cut := range m.Cuts[id] {
		var flow float32
		var delay int32
		if exact {
			touched = touched[:0]
			flow = float32(m.changeRefs(cut, 1, 8, &touched))
			for _, leaf := range touched {
				m.nodes[leaf].refs--
			}
		} else {
			flow, delay = m.cutFlow(cut)
		}
		if selected < 0 || bestFlow > flow+epsilon || bestFlow > flow-epsilon && bestTime > delay {
			selected = i
			bestFlow = flow
			bestTime = delay
		}
	}
	if exact && best.refs != 0 {
		m.changeRefs(m.Cuts[id][selected], 1, 1<<30, nil)
	} else {
		best.refs = 0
	}
	if exact {
		_, bestTime = m.cutFlow(m.Cuts[id][selected])
	}
	best.delay = bestTime
	best.flow = bestFlow / best.flowRefs
	// Preserve relative ordering of all remaining cuts, matching memmove in C.
	chosen := m.Cuts[id][selected]
	copy(m.Cuts[id][1:selected+1], m.Cuts[id][:selected])
	m.Cuts[id][0] = chosen
	return touched[:0]
}

// GenerateMapping performs the source CNF-cost cut enumeration, area-flow round
// and exact-area round. Coarsening is a distinct preceding source operation;
// callers must pass its resulting network when reproducing GenerateCnf defaults.
func GenerateMapping(network Network, options Options) (*Mapping, error) {
	if network == nil {
		return nil, fmt.Errorf("mf: nil network")
	}
	if options.LUTSize < 3 || options.LUTSize > 8 || options.CutLimit < 2 || options.CutLimit > maxCuts || options.Rounds < 1 || options.ExactRounds < 0 {
		return nil, fmt.Errorf("mf: invalid mapping options")
	}
	count := network.ObjectCount()
	if count < 1 {
		return nil, fmt.Errorf("mf: empty network")
	}
	words := 1 << max(0, options.LUTSize-6)
	// Bound the initial reservation. Large networks can still have very few
	// distinct functions, and each eight-input truth occupies four words.
	hint := max(2, min(count, 8192))
	m := &Mapping{
		Network: network, Options: options, Cuts: make([][]Cut, count), nodes: make([]node, count),
		ClauseSizes:  make([]int, 0, hint),
		truthResults: make([]truthResult, 1<<min(12, bits.Len(uint(count-1)))),
		truthStore:   vec.Vec_MemAlloc(words, min(7, bits.Len(uint(hint))-1)),
	}
	vec.Vec_MemHashAlloc(m.truthStore, hint)
	m.intern(make([]uint64, words), 0)
	variable := make([]uint64, words)
	for i := range variable {
		variable[i] = 0xaaaaaaaaaaaaaaaa
	}
	m.intern(variable, 1)
	refs := make([]int32, count)
	m.phases = make([]bool, count)
	add := func(lit int32, owner int) error {
		if lit < 0 || int(lit>>1) >= owner {
			return fmt.Errorf("mf: non-topological fanin %d at object %d", lit, owner)
		}
		_, _, _, kind := network.Gate(lit >> 1)
		if kind == 6 || kind < 0 {
			return fmt.Errorf("mf: invalid fanin object %d", lit>>1)
		}
		if gateNode(kind) {
			refs[lit>>1]++
		}
		return nil
	}
	for i := 0; i < count; i++ {
		a, b, c, kind := network.Gate(int32(i))
		if i == 0 && kind != 0 {
			return nil, fmt.Errorf("mf: missing constant zero")
		}
		var fanins []int32
		switch kind {
		case 0:
			if i != 0 {
				return nil, fmt.Errorf("mf: noncanonical constant")
			}
		case 1:
		case 2, 3:
			fanins = []int32{a, b}
		case 4:
			fanins = []int32{c, b, a}
		case 5:
			fanins = []int32{a, a}
		case 6:
			if err := add(a, i); err != nil {
				return nil, err
			}
			continue
		default:
			return nil, fmt.Errorf("mf: unsupported object kind %d", kind)
		}
		for _, lit := range fanins {
			if err := add(lit, i); err != nil {
				return nil, err
			}
		}
		value := func(lit int32) bool { return m.phases[lit>>1] != (lit&1 != 0) }
		switch kind {
		case 2:
			m.phases[i] = value(a) && value(b)
		case 3:
			m.phases[i] = value(a) != value(b)
		case 4:
			if value(a) {
				m.phases[i] = value(b)
			} else {
				m.phases[i] = value(c)
			}
		case 5:
			m.phases[i] = value(a)
		}
		if choices, ok := network.(interface{ Sibling(int32) int32 }); ok {
			sibling := choices.Sibling(int32(i))
			if sibling < 0 || int(sibling) >= i && sibling != 0 {
				return nil, fmt.Errorf("mf: invalid sibling %d at %d", sibling, i)
			}
		}
	}
	// CO references were counted above. If outputs are direct driver literals
	// without explicit CO nodes, count them once here instead.
	explicitOutputs := 0
	for i := 0; i < count; i++ {
		_, _, _, kind := network.Gate(int32(i))
		if kind == 6 {
			explicitOutputs++
		}
	}
	outputs := network.OutputLiterals()
	for _, lit := range outputs {
		if lit < 0 || int(lit>>1) >= count {
			return nil, fmt.Errorf("mf: invalid output literal %d", lit)
		}
	}
	listed := make(map[int32]bool)
	for _, lit := range network.InputLiterals() {
		if lit < 0 || lit&1 != 0 || int(lit>>1) >= count || listed[lit] {
			return nil, fmt.Errorf("mf: invalid input literal %d", lit)
		}
		_, _, _, kind := network.Gate(lit >> 1)
		if kind != 1 {
			return nil, fmt.Errorf("mf: interface input is not a CI")
		}
		listed[lit] = true
	}
	for i := 0; i < count; i++ {
		_, _, _, kind := network.Gate(int32(i))
		if kind == 1 && !listed[int32(i)<<1] {
			return nil, fmt.Errorf("mf: unlisted CI %d", i)
		}
	}
	if explicitOutputs == 0 {
		for _, lit := range outputs {
			if err := add(lit, count); err != nil {
				return nil, err
			}
		}
	} else if explicitOutputs != len(outputs) {
		return nil, fmt.Errorf("mf: output interface does not match CO objects")
	}
	discountFlowReferences(network, refs)
	for i := range m.nodes {
		m.nodes[i].flowRefs = float32(max(1, refs[i]))
	}
	candidates := newCutSet(options.CutLimit)
	for i := 0; i < count; i++ {
		_, _, _, kind := network.Gate(int32(i))
		if gateNode(kind) {
			if err := m.merge(int32(i), candidates); err != nil {
				return nil, err
			}
		}
	}
	m.truthResults = nil
	m.setMapRefs(0, false)
	var touched []int32
	for iter := 1; iter < options.Rounds+options.ExactRounds; iter++ {
		exact := iter >= options.Rounds
		for i := 0; i < count; i++ {
			_, _, _, kind := network.Gate(int32(i))
			if gateNode(kind) {
				touched = m.bestCut(int32(i), exact, touched)
			}
		}
		m.setMapRefs(iter, exact)
	}
	m.Refs = make([]int32, count)
	for i := range m.nodes {
		m.Refs[i] = m.nodes[i].refs
	}
	m.nodes = nil
	// The exported views keep the pages live after the hash index is discarded.
	m.Truths = make([][]uint64, len(m.ClauseSizes))
	for i := range m.Truths {
		m.Truths[i] = vec.Vec_MemReadEntry(m.truthStore, i)
	}
	m.truthStore = nil
	m.phases = nil
	return m, nil
}
