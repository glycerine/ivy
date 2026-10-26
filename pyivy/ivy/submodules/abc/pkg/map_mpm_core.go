package abc

// Concrete subject graph, cut truth computation, and LUT mapping from
// src/map/mpm/mpm.h, mpmInt.h, mpmMig.h, mpmMig.c, mpmTruth.c, mpmMan.c,
// mpmMap.c, and mpmCore.c.

import (
	"fmt"
	"math/bits"
	"slices"
)

const (
	MPMMaximumCuts = 32
	mpmUnitEdge    = 50
	mpmUnitRefs    = 100
)

type MPMNodeKind uint8

const (
	MPMConstant MPMNodeKind = iota
	MPMInput
	MPMOutput
	MPMBuffer
	MPMAnd
	MPMXor
	MPMMux
	MPMMajority
)

type MPMSubjectLit uint32

func makeMPMSubjectLit(id int, complement bool) (MPMSubjectLit, error) {
	if id < 0 || id >= int(^uint32(0)>>1) {
		return 0, ErrInvalidLiteral
	}
	lit := MPMSubjectLit(uint32(id) << 1)
	if complement {
		lit |= 1
	}
	return lit, nil
}

func (l MPMSubjectLit) ID() int                { return int(uint32(l) >> 1) }
func (l MPMSubjectLit) Complemented() bool     { return l&1 != 0 }
func (l MPMSubjectLit) Not() MPMSubjectLit     { return l ^ 1 }
func (l MPMSubjectLit) Regular() MPMSubjectLit { return l &^ 1 }

type mpmSubjectNode struct {
	kind     MPMNodeKind
	fanins   [3]MPMSubjectLit
	faninNum uint8
	terminal int
	phase    bool
	sibling  int
}

// MPMSubjectGraph is the owned, pointer-free counterpart of Mig_Man_t. Object
// zero is constant false. Nodes and terminals retain topological object IDs.
type MPMSubjectGraph struct {
	Name        string
	nodes       []mpmSubjectNode
	inputs      []int
	outputs     []int
	registers   int
	objectLimit int
}

func NewMPMSubjectGraph(maxObjects int) (*MPMSubjectGraph, error) {
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if maxObjects < 1 || uint64(maxObjects) >= uint64(^uint32(0)>>1) {
		return nil, ErrDimension
	}
	return &MPMSubjectGraph{nodes: []mpmSubjectNode{{kind: MPMConstant, sibling: -1}}, objectLimit: maxObjects}, nil
}

func (g *MPMSubjectGraph) ObjectCount() int {
	if g == nil {
		return 0
	}
	return len(g.nodes)
}
func (g *MPMSubjectGraph) InputCount() int {
	if g == nil {
		return 0
	}
	return len(g.inputs)
}
func (g *MPMSubjectGraph) OutputCount() int {
	if g == nil {
		return 0
	}
	return len(g.outputs)
}
func (g *MPMSubjectGraph) RegisterCount() int {
	if g == nil {
		return 0
	}
	return g.registers
}
func (g *MPMSubjectGraph) NodeCount() int {
	if g == nil {
		return 0
	}
	return len(g.nodes) - len(g.inputs) - len(g.outputs) - 1
}
func (g *MPMSubjectGraph) TypeCount(kind MPMNodeKind) int {
	if g == nil {
		return 0
	}
	count := 0
	for _, node := range g.nodes {
		if node.kind == kind {
			count++
		}
	}
	return count
}
func (g *MPMSubjectGraph) AndCount() int { return g.TypeCount(MPMAnd) }
func (g *MPMSubjectGraph) XorCount() int { return g.TypeCount(MPMXor) }
func (g *MPMSubjectGraph) MuxCount() int { return g.TypeCount(MPMMux) }
func (g *MPMSubjectGraph) Inputs() []int {
	if g == nil {
		return nil
	}
	return slices.Clone(g.inputs)
}
func (g *MPMSubjectGraph) Outputs() []int {
	if g == nil {
		return nil
	}
	return slices.Clone(g.outputs)
}

func (g *MPMSubjectGraph) validSignal(l MPMSubjectLit) bool {
	return g != nil && l.ID() >= 0 && l.ID() < len(g.nodes) && g.nodes[l.ID()].kind != MPMOutput
}
func (g *MPMSubjectGraph) appendNode(n mpmSubjectNode) (MPMSubjectLit, error) {
	if g == nil || len(g.nodes) == 0 {
		return 0, ErrInvariant
	}
	if len(g.nodes) >= g.objectLimit {
		return 0, ErrLimit
	}
	n.sibling = -1
	g.nodes = append(g.nodes, n)
	return makeMPMSubjectLit(len(g.nodes)-1, false)
}
func (g *MPMSubjectGraph) AppendInput() (MPMSubjectLit, error) {
	l, err := g.appendNode(mpmSubjectNode{kind: MPMInput, terminal: len(g.inputs)})
	if err == nil {
		g.inputs = append(g.inputs, l.ID())
	}
	return l, err
}
func (g *MPMSubjectGraph) AppendOutput(driver MPMSubjectLit) (MPMSubjectLit, error) {
	if !g.validSignal(driver) {
		return 0, ErrInvalidLiteral
	}
	l, err := g.appendNode(mpmSubjectNode{kind: MPMOutput, fanins: [3]MPMSubjectLit{driver}, faninNum: 1, terminal: len(g.outputs)})
	if err == nil {
		g.outputs = append(g.outputs, l.ID())
	}
	return l, err
}
func (g *MPMSubjectGraph) appendGate(kind MPMNodeKind, fanins ...MPMSubjectLit) (MPMSubjectLit, error) {
	want := 2
	if kind == MPMBuffer {
		want = 1
	}
	if kind == MPMMux || kind == MPMMajority {
		want = 3
	}
	if len(fanins) != want {
		return 0, ErrDimension
	}
	var n mpmSubjectNode
	n.kind, n.faninNum = kind, uint8(want)
	for i, fanin := range fanins {
		if !g.validSignal(fanin) {
			return 0, ErrInvalidLiteral
		}
		n.fanins[i] = fanin
	}
	return g.appendNode(n)
}
func (g *MPMSubjectGraph) AppendBuffer(a MPMSubjectLit) (MPMSubjectLit, error) {
	return g.appendGate(MPMBuffer, a)
}
func (g *MPMSubjectGraph) AppendAnd(a, b MPMSubjectLit) (MPMSubjectLit, error) {
	return g.appendGate(MPMAnd, a, b)
}
func (g *MPMSubjectGraph) AppendXor(a, b MPMSubjectLit) (MPMSubjectLit, error) {
	return g.appendGate(MPMXor, a, b)
}
func (g *MPMSubjectGraph) AppendMux(low, high, control MPMSubjectLit) (MPMSubjectLit, error) {
	return g.appendGate(MPMMux, low, high, control)
}
func (g *MPMSubjectGraph) AppendMajority(a, b, c MPMSubjectLit) (MPMSubjectLit, error) {
	return g.appendGate(MPMMajority, a, b, c)
}

func (g *MPMSubjectGraph) SetRegisterCount(count int) error {
	if g == nil || count < 0 || count > len(g.inputs) || count > len(g.outputs) {
		return ErrDimension
	}
	g.registers = count
	return nil
}
func (g *MPMSubjectGraph) SetChoice(root, sibling int, oppositePhase bool) error {
	if g == nil || root <= sibling || sibling <= 0 || root >= len(g.nodes) || g.nodes[root].kind < MPMAnd || g.nodes[sibling].kind < MPMAnd {
		return ErrDimension
	}
	g.nodes[root].sibling = sibling
	g.nodes[root].phase = oppositePhase
	return nil
}

func (g *MPMSubjectGraph) Check() error {
	if g == nil || len(g.nodes) == 0 || g.nodes[0].kind != MPMConstant {
		return ErrInvariant
	}
	seenInputs, seenOutputs := 0, 0
	for id, node := range g.nodes {
		want := 0
		switch node.kind {
		case MPMConstant:
			if id != 0 {
				return ErrInvariant
			}
		case MPMInput:
			want = 0
			if node.terminal != seenInputs {
				return ErrInvariant
			}
			seenInputs++
		case MPMOutput, MPMBuffer:
			want = 1
			if node.kind == MPMOutput {
				if node.terminal != seenOutputs {
					return ErrInvariant
				}
				seenOutputs++
			}
		case MPMAnd, MPMXor:
			want = 2
		case MPMMux, MPMMajority:
			want = 3
		default:
			return ErrInvariant
		}
		if int(node.faninNum) != want {
			return ErrInvariant
		}
		for i := 0; i < want; i++ {
			if node.fanins[i].ID() >= id || node.fanins[i].ID() < 0 {
				return ErrInvariant
			}
		}
		if node.sibling >= id || node.sibling == 0 {
			return ErrInvariant
		}
	}
	if seenInputs != len(g.inputs) || seenOutputs != len(g.outputs) || g.registers < 0 || g.registers > min(seenInputs, seenOutputs) {
		return ErrInvariant
	}
	return nil
}

// MPMSubjectGraphFromGia performs the independently useful mpmMig conversion.
// Object IDs are preserved, including native buffers, XORs, and muxes.
func MPMSubjectGraphFromGia(source *GiaManager, maxObjects int) (*MPMSubjectGraph, error) {
	if source == nil {
		return nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	if maxObjects == 0 {
		maxObjects = source.ObjectCount()
	}
	if maxObjects < source.ObjectCount() {
		return nil, ErrLimit
	}
	g, err := NewMPMSubjectGraph(maxObjects)
	if err != nil {
		return nil, err
	}
	g.Name = source.Name
	for id := 1; id < source.ObjectCount(); id++ {
		o := source.objects[id]
		if o.IsCI() {
			_, err = g.AppendInput()
		} else if o.IsCO() {
			a, e := o.Fanin0(GiaID(id))
			if e != nil {
				return nil, e
			}
			_, err = g.AppendOutput(MPMSubjectLit(a))
		} else {
			a, e := o.Fanin0(GiaID(id))
			if e != nil {
				return nil, e
			}
			if o.IsBuffer() {
				_, err = g.AppendBuffer(MPMSubjectLit(a))
			} else {
				b, e := o.Fanin1(GiaID(id))
				if e != nil {
					return nil, e
				}
				if c, mux := source.MuxControl(GiaID(id)); mux {
					_, err = g.AppendMux(MPMSubjectLit(a), MPMSubjectLit(b), MPMSubjectLit(c))
				} else if o.IsXor() {
					_, err = g.AppendXor(MPMSubjectLit(a), MPMSubjectLit(b))
				} else {
					_, err = g.AppendAnd(MPMSubjectLit(a), MPMSubjectLit(b))
				}
			}
		}
		if err != nil {
			return nil, err
		}
	}
	if err = g.SetRegisterCount(source.RegisterCount()); err != nil {
		return nil, err
	}
	return g, g.Check()
}

type MPMParameters struct {
	Library         *MPMLUTLibrary
	CutLimit        int
	DelayTarget     int
	UseTruth        bool
	MinimizeSupport bool
	OneRound        bool
	MaxObjects      int
	MaxCuts         int
	MaxCombinations int
	Stop            func() bool
}

func DefaultMPMParameters() MPMParameters {
	return MPMParameters{CutLimit: 8, DelayTarget: -1, UseTruth: true, MaxObjects: 1 << 20, MaxCuts: 1 << 20, MaxCombinations: 1 << 22}
}

type MPMCut struct {
	Leaves      []int
	Truth       []uint64
	Complement  bool
	Arrival     int
	Area        int
	Edges       int
	AverageRefs int
}

func (c MPMCut) clone() MPMCut {
	c.Leaves = slices.Clone(c.Leaves)
	c.Truth = slices.Clone(c.Truth)
	return c
}
func (c MPMCut) Value(assignment uint64) bool {
	if len(c.Truth) == 0 {
		return c.Complement
	}
	word, bit := assignment>>6, assignment&63
	v := word < uint64(len(c.Truth)) && c.Truth[word]&(uint64(1)<<bit) != 0
	return v != c.Complement
}

type MPMMappingRecord struct {
	Root       int
	Leaves     []int
	Truth      []uint64
	Complement bool
	Arrival    int
	Area       int
	Edges      int
}

type MPMMappingStatistics struct {
	Rounds, Combinations, Merged, Retained, SupportReductions int
}

type MPMMappingResult struct {
	Records    []MPMMappingRecord
	Cuts       [][]MPMCut
	Arrival    []int
	Required   []int
	References []int
	Delay      int
	Area       int
	Edges      int
	Statistics MPMMappingStatistics
}

func (r *MPMMappingResult) GiaMapping(maxEntries int) (*GiaLUTMapping, error) {
	if r == nil || len(r.Cuts) == 0 {
		return nil, ErrDimension
	}
	records := make([]GiaLUTRecord, len(r.Records))
	for i, record := range r.Records {
		fanins := make([]GiaID, len(record.Leaves))
		for k, leaf := range record.Leaves {
			if leaf < 0 || leaf >= record.Root {
				return nil, ErrInvariant
			}
			fanins[k] = GiaID(leaf)
		}
		records[i] = GiaLUTRecord{Root: GiaID(record.Root), Fanins: fanins}
	}
	return NewGiaLUTMapping(len(r.Cuts), records, maxEntries)
}

type mpmRoundMode uint8

const (
	mpmDelay mpmRoundMode = iota
	mpmDelayArea
	mpmArea
	mpmAreaEdges
)

type mpmMapper struct {
	g                              *MPMSubjectGraph
	p                              MPMParameters
	refs, mapRefs, estRefs         []int
	arrival, required, area, edges []int
	cuts                           [][]MPMCut
	best                           []MPMCut
	stats                          MPMMappingStatistics
}

func normalizeMPMParameters(g *MPMSubjectGraph, p MPMParameters) (MPMParameters, error) {
	if p.Library == nil {
		return p, fmt.Errorf("%w: missing LUT library", ErrDimension)
	}
	if p.CutLimit == 0 {
		p.CutLimit = 8
	}
	if p.MaxObjects == 0 {
		p.MaxObjects = 1 << 20
	}
	if p.MaxCuts == 0 {
		p.MaxCuts = 1 << 20
	}
	if p.MaxCombinations == 0 {
		p.MaxCombinations = 1 << 22
	}
	if p.DelayTarget == 0 {
		p.DelayTarget = -1
	}
	if p.CutLimit < 2 || p.CutLimit > MPMMaximumCuts || p.Library.MaximumLUTSize < 1 || p.Library.MaximumLUTSize > MPMMaximumVariables || p.MaxObjects < 1 || p.MaxCuts < 1 || p.MaxCombinations < 1 {
		return p, ErrDimension
	}
	if g.ObjectCount() > p.MaxObjects {
		return p, ErrLimit
	}
	return p, nil
}

func (g *MPMSubjectGraph) MapLUTs(parameters MPMParameters) (*MPMMappingResult, error) {
	if err := g.Check(); err != nil {
		return nil, err
	}
	p, err := normalizeMPMParameters(g, parameters)
	if err != nil {
		return nil, err
	}
	if p.Stop != nil && p.Stop() {
		return nil, ErrCancelled
	}
	n := len(g.nodes)
	m := &mpmMapper{g: g, p: p, refs: make([]int, n), mapRefs: make([]int, n), estRefs: make([]int, n), arrival: make([]int, n), required: make([]int, n), area: make([]int, n), edges: make([]int, n), cuts: make([][]MPMCut, n), best: make([]MPMCut, n)}
	for _, node := range g.nodes {
		for i := 0; i < int(node.faninNum); i++ {
			m.refs[node.fanins[i].ID()]++
		}
		if node.sibling > 0 {
			m.refs[node.sibling]++
		}
	}
	for i := range m.estRefs {
		m.estRefs[i] = mpmUnitRefs * m.refs[i]
	}
	modes := []mpmRoundMode{mpmDelay}
	if !p.OneRound {
		modes = []mpmRoundMode{mpmDelay, mpmDelayArea, mpmArea, mpmArea, mpmAreaEdges}
	}
	for round, mode := range modes {
		if p.Stop != nil && p.Stop() {
			return nil, ErrCancelled
		}
		if round >= 3 {
			for i := range m.estRefs {
				m.estRefs[i] = (m.estRefs[i] + mpmUnitRefs*m.mapRefs[i]) / 2
			}
		}
		if err = m.performRound(mode, round >= 3); err != nil {
			return nil, err
		}
		m.stats.Rounds++
	}
	result := m.result()
	return result, nil
}

func mpmTruthWords(vars int) int {
	if vars <= 6 {
		return 1
	}
	return 1 << uint(vars-6)
}
func mpmTruthBit(words []uint64, assignment uint64) bool {
	return words[assignment>>6]&(uint64(1)<<(assignment&63)) != 0
}
func mpmSetTruthBit(words []uint64, assignment uint64) {
	words[assignment>>6] |= uint64(1) << (assignment & 63)
}

func mpmUnitCut(id int) MPMCut { return MPMCut{Leaves: []int{id}, Truth: []uint64{2}} }
func mpmConstantCut() MPMCut   { return MPMCut{Truth: []uint64{0}} }

func (m *mpmMapper) performRound(mode mpmRoundMode, main bool) error {
	n := len(m.g.nodes)
	m.cuts = make([][]MPMCut, n)
	m.arrival, m.area, m.edges = make([]int, n), make([]int, n), make([]int, n)
	m.cuts[0] = []MPMCut{mpmConstantCut()}
	for _, id := range m.g.inputs {
		m.cuts[id] = []MPMCut{mpmUnitCut(id)}
	}
	stored := len(m.g.inputs) + 1
	for id, node := range m.g.nodes {
		if m.p.Stop != nil && m.p.Stop() {
			return ErrCancelled
		}
		if node.kind == MPMBuffer {
			driver := node.fanins[0]
			if len(m.cuts[driver.ID()]) == 0 {
				return ErrInvariant
			}
			m.cuts[id] = make([]MPMCut, len(m.cuts[driver.ID()]))
			for i := range m.cuts[id] {
				m.cuts[id][i] = m.cuts[driver.ID()][i].clone()
				m.cuts[id][i].Complement = m.cuts[id][i].Complement != driver.Complemented()
			}
			m.best[id] = m.cuts[id][0].clone()
			m.arrival[id], m.area[id], m.edges[id] = m.arrival[driver.ID()], m.area[driver.ID()], m.edges[driver.ID()]
			continue
		}
		if node.kind < MPMAnd {
			continue
		}
		candidates, err := m.nodeCuts(id, node, mode, main)
		if err != nil {
			return err
		}
		stored += len(candidates)
		if stored > m.p.MaxCuts {
			return ErrLimit
		}
		m.cuts[id] = append(candidates, mpmUnitCut(id))
		m.best[id] = candidates[0].clone()
		m.arrival[id], m.area[id], m.edges[id] = candidates[0].Arrival, candidates[0].Area, candidates[0].Edges
	}
	return m.finalizeRound()
}

func (m *mpmMapper) nodeCuts(id int, node mpmSubjectNode, mode mpmRoundMode, main bool) ([]MPMCut, error) {
	lists := make([][]MPMCut, node.faninNum)
	for i := range lists {
		lists[i] = m.cuts[node.fanins[i].ID()]
		if len(lists[i]) == 0 {
			return nil, ErrInvariant
		}
	}
	kept := make([]MPMCut, 0, m.p.CutLimit-1)
	indices := make([]int, len(lists))
	for {
		m.stats.Combinations++
		if m.stats.Combinations > m.p.MaxCombinations {
			return nil, ErrLimit
		}
		parts := make([]MPMCut, len(lists))
		for i := range lists {
			parts[i] = lists[i][indices[i]]
		}
		cut, ok, err := m.compose(node, parts, main)
		if err != nil {
			return nil, err
		}
		if ok {
			m.stats.Merged++
			kept = m.insertCut(kept, cut, mode)
		}
		at := len(indices) - 1
		for at >= 0 {
			indices[at]++
			if indices[at] < len(lists[at]) {
				break
			}
			indices[at] = 0
			at--
		}
		if at < 0 {
			break
		}
	}
	if node.sibling > 0 {
		for _, old := range m.cuts[node.sibling] {
			if len(old.Leaves) == 1 && old.Leaves[0] == node.sibling {
				continue
			}
			old = old.clone()
			old.Complement = old.Complement != node.phase
			old.Arrival, old.Area, old.Edges, old.AverageRefs = m.cutMetrics(old, main)
			kept = m.insertCut(kept, old, mode)
		}
	}
	if len(kept) == 0 {
		return nil, fmt.Errorf("%w: no feasible cut for object %d", ErrInvariant, id)
	}
	m.stats.Retained += len(kept)
	return kept, nil
}

func mergeMPMLeaves(parts []MPMCut, limit int) ([]int, bool) {
	seen := make(map[int]bool)
	for _, part := range parts {
		for _, leaf := range part.Leaves {
			seen[leaf] = true
		}
	}
	if len(seen) > limit {
		return nil, false
	}
	leaves := make([]int, 0, len(seen))
	for leaf := range seen {
		leaves = append(leaves, leaf)
	}
	slices.Sort(leaves)
	return leaves, true
}
func assignmentForMPMPart(part MPMCut, positions map[int]int, assignment uint64) uint64 {
	var local uint64
	for i, leaf := range part.Leaves {
		if assignment&(uint64(1)<<uint(positions[leaf])) != 0 {
			local |= uint64(1) << uint(i)
		}
	}
	return local
}
func (m *mpmMapper) compose(node mpmSubjectNode, parts []MPMCut, main bool) (MPMCut, bool, error) {
	leaves, ok := mergeMPMLeaves(parts, m.p.Library.MaximumLUTSize)
	if !ok {
		return MPMCut{}, false, nil
	}
	positions := make(map[int]int, len(leaves))
	for i, leaf := range leaves {
		positions[leaf] = i
	}
	truth := make([]uint64, mpmTruthWords(len(leaves)))
	assignments := uint64(1) << uint(len(leaves))
	for a := uint64(0); a < assignments; a++ {
		v := [3]bool{}
		for i, part := range parts {
			v[i] = part.Value(assignmentForMPMPart(part, positions, a)) != node.fanins[i].Complemented()
		}
		value := false
		switch node.kind {
		case MPMAnd:
			value = v[0] && v[1]
		case MPMXor:
			value = v[0] != v[1]
		case MPMMux:
			value = v[1]
			if !v[2] {
				value = v[0]
			}
		case MPMMajority:
			value = v[0] && v[1] || v[0] && v[2] || v[1] && v[2]
		default:
			return MPMCut{}, false, ErrInvariant
		}
		if value {
			mpmSetTruthBit(truth, a)
		}
	}
	cut := MPMCut{Leaves: leaves, Truth: truth}
	if truth[0]&1 != 0 {
		for i := range truth {
			truth[i] = ^truth[i]
		}
		cut.Complement = true
		maskMPMTruth(truth, len(leaves))
	}
	if m.p.MinimizeSupport {
		reduced := minimizeMPMCut(cut)
		if len(reduced.Leaves) < len(cut.Leaves) {
			m.stats.SupportReductions++
			cut = reduced
		}
	}
	cut.Arrival, cut.Area, cut.Edges, cut.AverageRefs = m.cutMetrics(cut, main)
	return cut, true, nil
}
func maskMPMTruth(words []uint64, vars int) {
	if vars < 6 {
		words[0] &= (uint64(1) << (uint64(1) << uint(vars))) - 1
	}
}
func minimizeMPMCut(cut MPMCut) MPMCut {
	for variable := 0; variable < len(cut.Leaves); {
		independent := true
		for a := uint64(0); a < uint64(1)<<uint(len(cut.Leaves)); a++ {
			if a&(uint64(1)<<uint(variable)) == 0 && cut.Value(a) != cut.Value(a|uint64(1)<<uint(variable)) {
				independent = false
				break
			}
		}
		if !independent {
			variable++
			continue
		}
		newLeaves := append(slices.Clone(cut.Leaves[:variable]), cut.Leaves[variable+1:]...)
		newTruth := make([]uint64, mpmTruthWords(len(newLeaves)))
		for a := uint64(0); a < uint64(1)<<uint(len(newLeaves)); a++ {
			low := a & ((uint64(1) << uint(variable)) - 1)
			high := a >> uint(variable)
			old := low | high<<uint(variable+1)
			if cut.Value(old) != cut.Complement {
				mpmSetTruthBit(newTruth, a)
			}
		}
		cut.Leaves, cut.Truth = newLeaves, newTruth
	}
	return cut
}

func (m *mpmMapper) cutMetrics(cut MPMCut, main bool) (arrival, area, edges, average int) {
	area = m.p.Library.Areas[len(cut.Leaves)]
	edges = mpmUnitEdge * len(cut.Leaves)
	for pin, leaf := range cut.Leaves {
		arrival = max(arrival, m.arrival[leaf]+m.p.Library.Delays[len(cut.Leaves)][pin])
		if main && m.mapRefs[leaf] == 0 {
			area += m.area[leaf]
			edges += m.edges[leaf]
		} else {
			refs := m.estRefs[leaf]
			if refs <= 0 {
				refs = mpmUnitRefs
			}
			area += mpmUnitRefs * m.area[leaf] / refs
			edges += mpmUnitRefs * m.edges[leaf] / refs
			if main {
				average += m.mapRefs[leaf]
			} else {
				average += m.refs[leaf]
			}
		}
	}
	if len(cut.Leaves) != 0 {
		average = average * mpmUnitEdge / len(cut.Leaves)
	}
	return
}
func compareMPMCuts(a, b MPMCut, mode mpmRoundMode) int {
	values := [][2]int{}
	switch mode {
	case mpmDelay:
		values = [][2]int{{a.Arrival, b.Arrival}, {len(a.Leaves), len(b.Leaves)}, {a.Area, b.Area}, {a.Edges, b.Edges}}
	case mpmDelayArea:
		values = [][2]int{{a.Arrival, b.Arrival}, {a.Area, b.Area}, {a.Edges, b.Edges}, {len(a.Leaves), len(b.Leaves)}}
	case mpmArea:
		values = [][2]int{{a.Area, b.Area}, {len(a.Leaves), len(b.Leaves)}, {a.Edges, b.Edges}, {a.AverageRefs, b.AverageRefs}, {a.Arrival, b.Arrival}}
	case mpmAreaEdges:
		values = [][2]int{{a.Area, b.Area}, {a.Edges, b.Edges}, {a.AverageRefs, b.AverageRefs}, {len(a.Leaves), len(b.Leaves)}, {a.Arrival, b.Arrival}}
	}
	for _, pair := range values {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	return slices.Compare(a.Leaves, b.Leaves)
}
func mpmContains(base, subset []int) bool {
	i := 0
	for _, x := range subset {
		for i < len(base) && base[i] < x {
			i++
		}
		if i == len(base) || base[i] != x {
			return false
		}
		i++
	}
	return true
}
func (m *mpmMapper) insertCut(cuts []MPMCut, cut MPMCut, mode mpmRoundMode) []MPMCut {
	sign := func(c MPMCut) uint64 {
		var s uint64
		for _, x := range c.Leaves {
			s |= uint64(1) << uint(x&63)
		}
		return s
	}
	cutSign := sign(cut)
	for _, old := range cuts {
		if len(old.Leaves) <= len(cut.Leaves) && sign(old)&cutSign == sign(old) && mpmContains(cut.Leaves, old.Leaves) {
			return cuts
		}
	}
	kept := cuts[:0]
	for _, old := range cuts {
		if len(cut.Leaves) <= len(old.Leaves) && cutSign&sign(old) == cutSign && mpmContains(old.Leaves, cut.Leaves) {
			continue
		}
		kept = append(kept, old)
	}
	kept = append(kept, cut)
	slices.SortStableFunc(kept, func(a, b MPMCut) int { return compareMPMCuts(a, b, mode) })
	if len(kept) >= m.p.CutLimit {
		kept = kept[:m.p.CutLimit-1]
	}
	return kept
}

func (m *mpmMapper) finalizeRound() error {
	const infinity = int(^uint(0) >> 2)
	for i := range m.mapRefs {
		m.mapRefs[i] = 0
		m.required[i] = infinity
	}
	delay := 0
	for _, output := range m.g.outputs {
		driver := m.g.nodes[output].fanins[0].ID()
		delay = max(delay, m.arrival[driver])
	}
	if m.p.DelayTarget >= 0 {
		delay = max(delay, m.p.DelayTarget)
	}
	for _, output := range m.g.outputs {
		driver := m.g.nodes[output].fanins[0].ID()
		m.mapRefs[driver]++
		m.required[driver] = delay
	}
	for id := len(m.g.nodes) - 1; id > 0; id-- {
		if m.mapRefs[id] == 0 {
			continue
		}
		if m.g.nodes[id].kind == MPMBuffer {
			driver := m.g.nodes[id].fanins[0].ID()
			m.mapRefs[driver] += m.mapRefs[id]
			m.required[driver] = min(m.required[driver], m.required[id])
			continue
		}
		if m.g.nodes[id].kind < MPMAnd {
			continue
		}
		cut := m.best[id]
		for pin, leaf := range cut.Leaves {
			m.mapRefs[leaf]++
			m.required[leaf] = min(m.required[leaf], m.required[id]-m.p.Library.Delays[len(cut.Leaves)][pin])
		}
	}
	return nil
}
func (m *mpmMapper) result() *MPMMappingResult {
	r := &MPMMappingResult{Cuts: make([][]MPMCut, len(m.cuts)), Arrival: slices.Clone(m.arrival), Required: slices.Clone(m.required), References: slices.Clone(m.mapRefs), Statistics: m.stats}
	for i := range m.cuts {
		r.Cuts[i] = make([]MPMCut, len(m.cuts[i]))
		for k := range m.cuts[i] {
			r.Cuts[i][k] = m.cuts[i][k].clone()
		}
	}
	for _, output := range m.g.outputs {
		r.Delay = max(r.Delay, m.arrival[m.g.nodes[output].fanins[0].ID()])
	}
	for id := 1; id < len(m.g.nodes); id++ {
		if m.mapRefs[id] == 0 || m.g.nodes[id].kind < MPMAnd {
			continue
		}
		c := m.best[id]
		r.Area += m.p.Library.Areas[len(c.Leaves)]
		r.Edges += len(c.Leaves)
		r.Records = append(r.Records, MPMMappingRecord{Root: id, Leaves: slices.Clone(c.Leaves), Truth: slices.Clone(c.Truth), Complement: c.Complement, Arrival: c.Arrival, Area: c.Area, Edges: c.Edges})
	}
	if MPMUnitArea != 0 {
		r.Area /= MPMUnitArea
	}
	return r
}

func (g *MPMSubjectGraph) SupportSize(root int, maximum int, stop func() bool) (int, error) {
	if err := g.Check(); err != nil {
		return 0, err
	}
	if root < 0 || root >= len(g.nodes) || maximum < 0 {
		return 0, ErrDimension
	}
	seen := make([]bool, len(g.nodes))
	stack := []int{root}
	count := 0
	for len(stack) != 0 {
		if stop != nil && stop() {
			return 0, ErrCancelled
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[id] {
			continue
		}
		seen[id] = true
		n := g.nodes[id]
		if n.kind == MPMInput {
			count++
			if maximum > 0 && count > maximum {
				return 0, ErrLimit
			}
			continue
		}
		for i := 0; i < int(n.faninNum); i++ {
			stack = append(stack, n.fanins[i].ID())
		}
	}
	return count, nil
}

func (g *MPMSubjectGraph) SmallSupportNodeCount(maximum int, stop func() bool) (int, error) {
	if maximum < 0 {
		return 0, ErrDimension
	}
	if err := g.Check(); err != nil {
		return 0, err
	}
	count := 0
	for id, node := range g.nodes {
		if node.kind < MPMAnd {
			continue
		}
		if stop != nil && stop() {
			return 0, ErrCancelled
		}
		size, err := g.SupportSize(id, 0, stop)
		if err != nil {
			return 0, err
		}
		if size <= maximum {
			count++
		}
	}
	return count, nil
}

func (r *MPMMappingResult) Validate(g *MPMSubjectGraph, p MPMParameters) error {
	if r == nil || g == nil || p.Library == nil || len(r.Cuts) != g.ObjectCount() || len(r.Arrival) != g.ObjectCount() || len(r.Required) != g.ObjectCount() || len(r.References) != g.ObjectCount() {
		return ErrInvariant
	}
	for _, rec := range r.Records {
		if rec.Root <= 0 || rec.Root >= g.ObjectCount() || len(rec.Leaves) > p.Library.MaximumLUTSize || !slices.IsSorted(rec.Leaves) {
			return ErrInvariant
		}
		for _, leaf := range rec.Leaves {
			if leaf < 0 || leaf >= rec.Root {
				return ErrInvariant
			}
		}
		if bits.Len(uint(len(rec.Truth))) > MPMMaximumVariables {
			return ErrInvariant
		}
	}
	return nil
}
