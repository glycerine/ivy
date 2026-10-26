package abc

// Derived from ABC src/proof/fraig/fraig.h, fraigInt.h, fraigVec.c,
// fraigMem.c, fraigTable.c, fraigCanon.c, fraigNode.c, and fraigFanout.c.

import (
	"math/bits"
	"sort"
)

// FraigID identifies an owned node. IDs are stable and never reused.
type FraigID uint32

// FraigEdge stores a node ID and a complement bit, mirroring tagged FRAIG
// pointers without relying on unsafe pointer arithmetic.
type FraigEdge uint32

func makeFraigEdge(id FraigID, complement bool) FraigEdge {
	e := FraigEdge(id << 1)
	if complement {
		e++
	}
	return e
}

func (e FraigEdge) ID() FraigID        { return FraigID(e >> 1) }
func (e FraigEdge) Complemented() bool { return e&1 != 0 }
func (e FraigEdge) Not() FraigEdge     { return e ^ 1 }

type FraigNodeKind uint8

const (
	FraigNodeConstant FraigNodeKind = iota
	FraigNodeInput
	FraigNodeAnd
)

type FraigOptions struct {
	MaxNodes         int
	MaxInputs        int
	MaxOutputs       int
	MaxTruthInputs   int
	FunctionalReduce bool
	Stop             func() bool
}

func (o FraigOptions) normalized() (FraigOptions, error) {
	if o.MaxNodes < 0 || o.MaxInputs < 0 || o.MaxOutputs < 0 || o.MaxTruthInputs < 0 {
		return o, ErrDimension
	}
	if o.MaxNodes == 0 {
		o.MaxNodes = 1 << 20
	}
	if o.MaxInputs == 0 {
		o.MaxInputs = 1 << 16
	}
	if o.MaxOutputs == 0 {
		o.MaxOutputs = 1 << 20
	}
	if o.MaxTruthInputs == 0 {
		o.MaxTruthInputs = 6
	}
	if o.MaxTruthInputs > 6 {
		o.MaxTruthInputs = 6
	}
	return o, nil
}

type fraigNode struct {
	kind        FraigNodeKind
	fanin       [2]FraigEdge
	input       int
	level       int
	refs        int
	fanouts     []FraigID
	output      bool
	truth       uint64
	truthKnown  bool
	randomSim   []uint32
	dynamicSim  []uint32
	simInverted bool
}

type fraigPair struct{ a, b FraigEdge }
type fraigFunction struct {
	inputs uint8
	truth  uint64
}

// FraigNodeSnapshot is an owned observation of one node.
type FraigNodeSnapshot struct {
	ID         FraigID
	Kind       FraigNodeKind
	Fanins     [2]FraigEdge
	Input      int
	Level      int
	References int
	Fanouts    []FraigID
	Output     bool
}

type FraigStats struct {
	Nodes             int
	Inputs            int
	Outputs           int
	StructuralEntries int
	FunctionalEntries int
}

// FraigManager is a concrete, owned AND/inverter graph foundation.
type FraigManager struct {
	opts       FraigOptions
	nodes      []fraigNode
	inputs     []FraigEdge
	outputs    []FraigEdge
	structural map[fraigPair]FraigEdge
	functional map[fraigFunction]FraigEdge
	closed     bool
}

func NewFraigManager(options FraigOptions) (*FraigManager, error) {
	o, err := options.normalized()
	if err != nil {
		return nil, err
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	if o.MaxNodes < 1 {
		return nil, ErrLimit
	}
	m := &FraigManager{opts: o, structural: make(map[fraigPair]FraigEdge), functional: make(map[fraigFunction]FraigEdge)}
	mask := fraigTruthMask(0)
	m.nodes = append(m.nodes, fraigNode{kind: FraigNodeConstant, input: -1, refs: 1, truth: mask, truthKnown: true})
	if o.FunctionalReduce {
		m.functional[fraigFunction{truth: mask}] = makeFraigEdge(0, false)
	}
	return m, nil
}

func fraigTruthMask(inputs int) uint64 {
	if inputs >= 6 {
		return ^uint64(0)
	}
	return uint64(1)<<(1<<inputs) - 1
}

func (m *FraigManager) check() error {
	if m == nil || m.closed {
		return ErrInvariant
	}
	if m.opts.Stop != nil && m.opts.Stop() {
		return ErrCancelled
	}
	return nil
}

func (m *FraigManager) ConstantOne() FraigEdge  { return makeFraigEdge(0, false) }
func (m *FraigManager) ConstantZero() FraigEdge { return makeFraigEdge(0, true) }

func (m *FraigManager) validEdge(e FraigEdge) bool { return m != nil && int(e.ID()) < len(m.nodes) }

func (m *FraigManager) NewInput() (FraigEdge, error) {
	if err := m.check(); err != nil {
		return 0, err
	}
	if len(m.inputs) >= m.opts.MaxInputs || len(m.nodes) >= m.opts.MaxNodes {
		return 0, ErrLimit
	}
	oldInputs := len(m.inputs)
	if oldInputs < m.opts.MaxTruthInputs {
		shift := uint(1 << oldInputs)
		for i := range m.nodes {
			if m.nodes[i].truthKnown {
				m.nodes[i].truth |= m.nodes[i].truth << shift
			}
		}
	}
	id := FraigID(len(m.nodes))
	index := len(m.inputs)
	n := fraigNode{kind: FraigNodeInput, input: index, level: 0, refs: 1}
	if len(m.inputs)+1 <= m.opts.MaxTruthInputs {
		var truth uint64
		for assignment := 0; assignment < 1<<(index+1); assignment++ {
			if assignment&(1<<index) != 0 {
				truth |= uint64(1) << assignment
			}
		}
		n.truth, n.truthKnown = truth, true
	}
	m.nodes = append(m.nodes, n)
	edge := makeFraigEdge(id, false)
	m.inputs = append(m.inputs, edge)
	m.rebuildFunctionalTable()
	return edge, nil
}

func (m *FraigManager) rebuildFunctionalTable() {
	m.functional = make(map[fraigFunction]FraigEdge)
	if !m.opts.FunctionalReduce || len(m.inputs) > m.opts.MaxTruthInputs {
		return
	}
	for id := range m.nodes {
		n := m.nodes[id]
		if !n.truthKnown {
			continue
		}
		canon, inv := canonicalFraigTruth(n.truth, len(m.inputs))
		e := makeFraigEdge(FraigID(id), false)
		if inv {
			e = e.Not()
		}
		key := fraigFunction{uint8(len(m.inputs)), canon}
		if _, exists := m.functional[key]; !exists {
			m.functional[key] = e
		}
	}
}

func orderedFraigPair(a, b FraigEdge) fraigPair {
	if a > b {
		a, b = b, a
	}
	return fraigPair{a, b}
}

func (m *FraigManager) edgeTruth(e FraigEdge) (uint64, bool) {
	n := &m.nodes[e.ID()]
	if !n.truthKnown {
		return 0, false
	}
	t := n.truth
	if e.Complemented() {
		t = ^t
	}
	return t & fraigTruthMask(len(m.inputs)), true
}

func canonicalFraigTruth(t uint64, inputs int) (uint64, bool) {
	mask := fraigTruthMask(inputs)
	t &= mask
	inv := (^t) & mask
	if inv < t {
		return inv, true
	}
	return t, false
}

// And performs trivial reduction, structural hashing, and optional exact
// functional reduction for managers within the configured truth-table width.
func (m *FraigManager) And(a, b FraigEdge) (FraigEdge, error) {
	if err := m.check(); err != nil {
		return 0, err
	}
	if !m.validEdge(a) || !m.validEdge(b) {
		return 0, ErrDimension
	}
	if a == b {
		return a, nil
	}
	if a == b.Not() {
		return m.ConstantZero(), nil
	}
	if a.ID() == 0 {
		if a.Complemented() {
			return m.ConstantZero(), nil
		}
		return b, nil
	}
	if b.ID() == 0 {
		if b.Complemented() {
			return m.ConstantZero(), nil
		}
		return a, nil
	}
	key := orderedFraigPair(a, b)
	if e, ok := m.structural[key]; ok {
		return e, nil
	}
	if len(m.nodes) >= m.opts.MaxNodes {
		return 0, ErrLimit
	}
	level := m.nodes[a.ID()].level
	if m.nodes[b.ID()].level > level {
		level = m.nodes[b.ID()].level
	}
	level++
	n := fraigNode{kind: FraigNodeAnd, fanin: [2]FraigEdge{key.a, key.b}, input: -1, level: level}
	if ta, oka := m.edgeTruth(a); oka {
		if tb, okb := m.edgeTruth(b); okb {
			n.truth = ta & tb
			n.truthKnown = true
		}
	}
	if m.opts.FunctionalReduce && n.truthKnown {
		canon, inv := canonicalFraigTruth(n.truth, len(m.inputs))
		fk := fraigFunction{uint8(len(m.inputs)), canon}
		if old, ok := m.functional[fk]; ok {
			if inv {
				old = old.Not()
			}
			return old, nil
		}
	}
	id := FraigID(len(m.nodes))
	edge := makeFraigEdge(id, false)
	m.nodes = append(m.nodes, n)
	m.structural[key] = edge
	for _, f := range n.fanin {
		node := &m.nodes[f.ID()]
		node.refs++
		node.fanouts = append(node.fanouts, id)
	}
	if m.opts.FunctionalReduce && n.truthKnown {
		canon, inv := canonicalFraigTruth(n.truth, len(m.inputs))
		stored := edge
		if inv {
			stored = stored.Not()
		}
		m.functional[fraigFunction{uint8(len(m.inputs)), canon}] = stored
	}
	return edge, nil
}

func (m *FraigManager) AddOutput(edge FraigEdge) error {
	if err := m.check(); err != nil {
		return err
	}
	if !m.validEdge(edge) {
		return ErrDimension
	}
	if len(m.outputs) >= m.opts.MaxOutputs {
		return ErrLimit
	}
	m.outputs = append(m.outputs, edge)
	n := &m.nodes[edge.ID()]
	n.refs++
	n.output = true
	return nil
}

func (m *FraigManager) RemoveLastOutput() (FraigEdge, error) {
	if err := m.check(); err != nil {
		return 0, err
	}
	if len(m.outputs) == 0 {
		return 0, ErrDimension
	}
	i := len(m.outputs) - 1
	e := m.outputs[i]
	m.outputs = m.outputs[:i]
	n := &m.nodes[e.ID()]
	n.refs--
	n.output = false
	for _, o := range m.outputs {
		if o.ID() == e.ID() {
			n.output = true
			break
		}
	}
	return e, nil
}

func (m *FraigManager) Snapshot(id FraigID) (FraigNodeSnapshot, error) {
	if err := m.check(); err != nil {
		return FraigNodeSnapshot{}, err
	}
	if int(id) >= len(m.nodes) {
		return FraigNodeSnapshot{}, ErrDimension
	}
	n := m.nodes[id]
	return FraigNodeSnapshot{ID: id, Kind: n.kind, Fanins: n.fanin, Input: n.input, Level: n.level, References: n.refs, Fanouts: append([]FraigID(nil), n.fanouts...), Output: n.output}, nil
}

func (m *FraigManager) Inputs() []FraigEdge {
	if m == nil {
		return nil
	}
	return append([]FraigEdge(nil), m.inputs...)
}
func (m *FraigManager) Outputs() []FraigEdge {
	if m == nil {
		return nil
	}
	return append([]FraigEdge(nil), m.outputs...)
}
func (m *FraigManager) Stats() FraigStats {
	if m == nil {
		return FraigStats{}
	}
	return FraigStats{len(m.nodes), len(m.inputs), len(m.outputs), len(m.structural), len(m.functional)}
}

func (m *FraigManager) Evaluate(edge FraigEdge, inputValues []bool) (bool, error) {
	if err := m.check(); err != nil {
		return false, err
	}
	if !m.validEdge(edge) || len(inputValues) != len(m.inputs) {
		return false, ErrDimension
	}
	values := make([]bool, len(m.nodes))
	values[0] = true
	for i := 1; i < len(m.nodes); i++ {
		n := m.nodes[i]
		switch n.kind {
		case FraigNodeInput:
			values[i] = inputValues[n.input]
		case FraigNodeAnd:
			a := values[n.fanin[0].ID()] != n.fanin[0].Complemented()
			b := values[n.fanin[1].ID()] != n.fanin[1].Complemented()
			values[i] = a && b
		}
	}
	return values[edge.ID()] != edge.Complemented(), nil
}

func removeFraigID(ids []FraigID, id FraigID) ([]FraigID, bool) {
	for i, v := range ids {
		if v == id {
			copy(ids[i:], ids[i+1:])
			return ids[:len(ids)-1], true
		}
	}
	return ids, false
}

// TransferFanouts replaces every use of from with to while retaining edge
// polarity. It is transactional and rejects cycles or a target with fanouts,
// matching the source precondition.
func (m *FraigManager) TransferFanouts(from, to FraigID) error {
	if err := m.check(); err != nil {
		return err
	}
	if int(from) >= len(m.nodes) || int(to) >= len(m.nodes) || from == to {
		return ErrDimension
	}
	if len(m.nodes[to].fanouts) != 0 {
		return ErrInvariant
	}
	if m.nodes[from].truthKnown != m.nodes[to].truthKnown || m.nodes[from].truthKnown && m.nodes[from].truth != m.nodes[to].truth {
		return ErrInvariant
	}
	for _, fo := range m.nodes[from].fanouts {
		if fo == to || fo < to {
			return ErrInvariant
		}
	}
	nodes := make([]fraigNode, len(m.nodes))
	copy(nodes, m.nodes)
	for i := range nodes {
		nodes[i].fanouts = append([]FraigID(nil), m.nodes[i].fanouts...)
	}
	structural := make(map[fraigPair]FraigEdge, len(m.structural))
	for _, fo := range nodes[from].fanouts {
		if m.opts.Stop != nil && m.opts.Stop() {
			return ErrCancelled
		}
		n := &nodes[fo]
		for j, e := range n.fanin {
			if e.ID() == from {
				n.fanin[j] = makeFraigEdge(to, e.Complemented())
				nodes[from].refs--
				nodes[to].refs++
			}
		}
		p := orderedFraigPair(n.fanin[0], n.fanin[1])
		n.fanin = [2]FraigEdge{p.a, p.b}
	}
	nodes[to].fanouts = append([]FraigID(nil), nodes[from].fanouts...)
	nodes[from].fanouts = nil
	for id := range nodes {
		if m.opts.Stop != nil && m.opts.Stop() {
			return ErrCancelled
		}
		if nodes[id].kind != FraigNodeAnd {
			continue
		}
		k := orderedFraigPair(nodes[id].fanin[0], nodes[id].fanin[1])
		if _, exists := structural[k]; exists {
			return ErrInvariant
		}
		structural[k] = makeFraigEdge(FraigID(id), false)
	}
	oldNodes, oldStructural := m.nodes, m.structural
	m.nodes, m.structural = nodes, structural
	if err := m.Validate(); err != nil {
		m.nodes, m.structural = oldNodes, oldStructural
		return err
	}
	return nil
}

func (m *FraigManager) SetSimulation(id FraigID, random, dynamic []uint32, inverted bool) error {
	if err := m.check(); err != nil {
		return err
	}
	if int(id) >= len(m.nodes) {
		return ErrDimension
	}
	n := &m.nodes[id]
	n.randomSim = append([]uint32(nil), random...)
	n.dynamicSim = append([]uint32(nil), dynamic...)
	n.simInverted = inverted
	return nil
}

func (m *FraigManager) simulation(id FraigID, random bool) ([]uint32, bool) {
	n := &m.nodes[id]
	if random {
		return n.randomSim, n.simInverted
	}
	return n.dynamicSim, n.simInverted
}

func (m *FraigManager) CompareSimulation(a, b FraigID, random bool, words int) (bool, error) {
	if err := m.check(); err != nil {
		return false, err
	}
	if int(a) >= len(m.nodes) || int(b) >= len(m.nodes) || words < 0 {
		return false, ErrDimension
	}
	x, _ := m.simulation(a, random)
	y, _ := m.simulation(b, random)
	if words > len(x) || words > len(y) {
		return false, ErrDimension
	}
	for i := 0; i < words; i++ {
		if x[i] != y[i] {
			return false, nil
		}
	}
	return true, nil
}

// CompareSimulationUnderMask implements Fraig_CompareSimInfoUnderMask using
// raw, phase-normalized simulation words, as the source table does.
func (m *FraigManager) CompareSimulationUnderMask(a, b FraigID, random bool, mask []uint32) (bool, error) {
	if err := m.check(); err != nil {
		return false, err
	}
	if int(a) >= len(m.nodes) || int(b) >= len(m.nodes) {
		return false, ErrDimension
	}
	x, _ := m.simulation(a, random)
	y, _ := m.simulation(b, random)
	if len(mask) > len(x) || len(mask) > len(y) {
		return false, ErrDimension
	}
	for i, wordMask := range mask {
		if x[i]&wordMask != y[i]&wordMask {
			return false, nil
		}
	}
	return true, nil
}

func (m *FraigManager) FirstSimulationDifference(a, b FraigID, complement, random bool, words int) (int, error) {
	if err := m.check(); err != nil {
		return -1, err
	}
	if int(a) >= len(m.nodes) || int(b) >= len(m.nodes) || words < 0 {
		return -1, ErrDimension
	}
	x, xi := m.simulation(a, random)
	y, yi := m.simulation(b, random)
	if words > len(x) || words > len(y) {
		return -1, ErrDimension
	}
	complement = complement != xi
	complement = complement != yi
	for i := 0; i < words; i++ {
		d := x[i] ^ y[i]
		if complement {
			d = ^d
		}
		if d != 0 {
			return i*32 + bits.TrailingZeros32(d), nil
		}
	}
	return -1, nil
}

func (m *FraigManager) SimulationXOR(a, b FraigID, random bool, words int) ([]uint32, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	if int(a) >= len(m.nodes) || int(b) >= len(m.nodes) || words < 0 {
		return nil, ErrDimension
	}
	x, _ := m.simulation(a, random)
	y, _ := m.simulation(b, random)
	if words > len(x) || words > len(y) {
		return nil, ErrDimension
	}
	r := make([]uint32, words)
	for i := range r {
		r[i] = x[i] ^ y[i]
	}
	return r, nil
}

func (m *FraigManager) Validate() error {
	if err := m.check(); err != nil {
		return err
	}
	if len(m.nodes) == 0 || m.nodes[0].kind != FraigNodeConstant {
		return ErrInvariant
	}
	refs := make([]int, len(m.nodes))
	refs[0] = 1
	fans := make([][]FraigID, len(m.nodes))
	for _, e := range m.inputs {
		refs[e.ID()]++
	}
	for _, e := range m.outputs {
		refs[e.ID()]++
	}
	seen := make(map[fraigPair]FraigID)
	for id := 1; id < len(m.nodes); id++ {
		n := m.nodes[id]
		if n.kind == FraigNodeInput {
			continue
		}
		if n.kind != FraigNodeAnd {
			return ErrInvariant
		}
		for _, e := range n.fanin {
			if int(e.ID()) >= id {
				return ErrInvariant
			}
			refs[e.ID()]++
			fans[e.ID()] = append(fans[e.ID()], FraigID(id))
		}
		k := orderedFraigPair(n.fanin[0], n.fanin[1])
		if _, ok := seen[k]; ok {
			return ErrInvariant
		}
		seen[k] = FraigID(id)
	}
	for i := range m.nodes {
		if m.nodes[i].refs != refs[i] || len(m.nodes[i].fanouts) != len(fans[i]) {
			return ErrInvariant
		}
		for j := range fans[i] {
			if fans[i][j] != m.nodes[i].fanouts[j] {
				return ErrInvariant
			}
		}
	}
	return nil
}

func (m *FraigManager) Close() {
	if m == nil || m.closed {
		return
	}
	m.closed = true
	m.nodes = nil
	m.inputs = nil
	m.outputs = nil
	m.structural = nil
	m.functional = nil
}

// FraigNodeVector is the owned replacement for Fraig_NodeVec_t.
type FraigNodeVector struct {
	entries []FraigEdge
	max     int
	closed  bool
}

func NewFraigNodeVector(capacity, max int) (*FraigNodeVector, error) {
	if capacity < 0 || max < 0 || max > 0 && capacity > max {
		return nil, ErrDimension
	}
	if capacity > 0 && capacity < 8 {
		capacity = 8
	}
	return &FraigNodeVector{entries: make([]FraigEdge, 0, capacity), max: max}, nil
}
func (v *FraigNodeVector) check() error {
	if v == nil || v.closed {
		return ErrInvariant
	}
	return nil
}
func (v *FraigNodeVector) Push(e FraigEdge) error {
	if err := v.check(); err != nil {
		return err
	}
	if v.max > 0 && len(v.entries) >= v.max {
		return ErrLimit
	}
	v.entries = append(v.entries, e)
	return nil
}
func (v *FraigNodeVector) PushUnique(e FraigEdge) (bool, error) {
	if err := v.check(); err != nil {
		return false, err
	}
	for _, x := range v.entries {
		if x == e {
			return true, nil
		}
	}
	return false, v.Push(e)
}
func (v *FraigNodeVector) PushOrderedNumber(e FraigEdge) error {
	if err := v.Push(e); err != nil {
		return err
	}
	i := len(v.entries) - 1
	for i > 0 && v.entries[i].ID() < v.entries[i-1].ID() {
		v.entries[i], v.entries[i-1] = v.entries[i-1], v.entries[i]
		i--
	}
	return nil
}
func (v *FraigNodeVector) PushUniqueOrderedNumber(e FraigEdge) (bool, error) {
	if err := v.check(); err != nil {
		return false, err
	}
	for _, old := range v.entries {
		if old == e {
			return true, nil
		}
	}
	return false, v.PushOrderedNumber(e)
}
func (v *FraigNodeVector) PushOrderedLevel(m *FraigManager, e FraigEdge) error {
	if err := m.check(); err != nil {
		return err
	}
	if !m.validEdge(e) {
		return ErrDimension
	}
	if err := v.Push(e); err != nil {
		return err
	}
	i := len(v.entries) - 1
	for i > 0 && m.nodes[v.entries[i].ID()].level > m.nodes[v.entries[i-1].ID()].level {
		v.entries[i], v.entries[i-1] = v.entries[i-1], v.entries[i]
		i--
	}
	return nil
}
func (v *FraigNodeVector) PushUniqueOrderedLevel(m *FraigManager, e FraigEdge) (bool, error) {
	if err := v.check(); err != nil {
		return false, err
	}
	for _, old := range v.entries {
		if old == e {
			return true, nil
		}
	}
	return false, v.PushOrderedLevel(m, e)
}
func (v *FraigNodeVector) Pop() (FraigEdge, error) {
	if err := v.check(); err != nil {
		return 0, err
	}
	if len(v.entries) == 0 {
		return 0, ErrDimension
	}
	i := len(v.entries) - 1
	e := v.entries[i]
	v.entries = v.entries[:i]
	return e, nil
}
func (v *FraigNodeVector) Remove(e FraigEdge) error {
	if err := v.check(); err != nil {
		return err
	}
	for i, x := range v.entries {
		if x == e {
			copy(v.entries[i:], v.entries[i+1:])
			v.entries = v.entries[:len(v.entries)-1]
			return nil
		}
	}
	return ErrDimension
}
func (v *FraigNodeVector) At(i int) (FraigEdge, error) {
	if err := v.check(); err != nil {
		return 0, err
	}
	if i < 0 || i >= len(v.entries) {
		return 0, ErrDimension
	}
	return v.entries[i], nil
}
func (v *FraigNodeVector) Write(i int, e FraigEdge) error {
	if err := v.check(); err != nil {
		return err
	}
	if i < 0 || i >= len(v.entries) {
		return ErrDimension
	}
	v.entries[i] = e
	return nil
}
func (v *FraigNodeVector) Shrink(size int) error {
	if err := v.check(); err != nil {
		return err
	}
	if size < 0 || size > len(v.entries) {
		return ErrDimension
	}
	v.entries = v.entries[:size]
	return nil
}
func (v *FraigNodeVector) Clear() {
	if v != nil && !v.closed {
		v.entries = v.entries[:0]
	}
}
func (v *FraigNodeVector) Snapshot() []FraigEdge {
	if v == nil || v.closed {
		return nil
	}
	return append([]FraigEdge(nil), v.entries...)
}
func (v *FraigNodeVector) Clone() (*FraigNodeVector, error) {
	if err := v.check(); err != nil {
		return nil, err
	}
	return &FraigNodeVector{entries: append([]FraigEdge(nil), v.entries...), max: v.max}, nil
}
func (v *FraigNodeVector) SortByNumber() {
	if v == nil || v.closed {
		return
	}
	sort.SliceStable(v.entries, func(i, j int) bool { return v.entries[i].ID() < v.entries[j].ID() })
}
func (v *FraigNodeVector) SortByLevel(m *FraigManager, increasing bool) error {
	if err := v.check(); err != nil {
		return err
	}
	if err := m.check(); err != nil {
		return err
	}
	for _, e := range v.entries {
		if !m.validEdge(e) {
			return ErrDimension
		}
	}
	sort.SliceStable(v.entries, func(i, j int) bool {
		a, b := m.nodes[v.entries[i].ID()].level, m.nodes[v.entries[j].ID()].level
		if a == b {
			return v.entries[i].ID() < v.entries[j].ID()
		}
		if increasing {
			return a < b
		}
		return a > b
	})
	return nil
}
func (v *FraigNodeVector) SortByReferences(m *FraigManager) error {
	if err := v.check(); err != nil {
		return err
	}
	if err := m.check(); err != nil {
		return err
	}
	sort.SliceStable(v.entries, func(i, j int) bool {
		a, b := m.nodes[v.entries[i].ID()], m.nodes[v.entries[j].ID()]
		if a.refs != b.refs {
			return a.refs < b.refs
		}
		if a.level != b.level {
			return a.level < b.level
		}
		return v.entries[i].ID() < v.entries[j].ID()
	})
	return nil
}
func (v *FraigNodeVector) Close() {
	if v != nil {
		v.closed = true
		v.entries = nil
	}
}

// FraigFixedPool is a checked handle pool replacing raw fixed-entry chunks.
type FraigFixedPool struct {
	entrySize, chunkSize, max int
	used, peak                int
	allocated                 int
	free                      []int
	live                      []bool
	closed                    bool
}
type FraigPoolStats struct{ EntrySize, ChunkSize, Allocated, Used, Peak, Memory int }

func NewFraigFixedPool(entrySize, max int) (*FraigFixedPool, error) {
	if entrySize < 4 || max < 0 {
		return nil, ErrDimension
	}
	maxInt := int(^uint(0) >> 1)
	if entrySize > maxInt/8 {
		return nil, ErrLimit
	}
	chunk := 1024
	if entrySize >= (1<<16)/chunk {
		chunk = (1 << 16) / entrySize
	}
	if chunk < 8 {
		chunk = 8
	}
	return &FraigFixedPool{entrySize: entrySize, chunkSize: chunk, max: max}, nil
}
func (p *FraigFixedPool) Fetch() (int, error) {
	if p == nil || p.closed {
		return 0, ErrInvariant
	}
	if p.max > 0 && p.used >= p.max {
		return 0, ErrLimit
	}
	if len(p.free) == 0 {
		start := len(p.live)
		grow := p.chunkSize
		if p.max > 0 && start+grow > p.max {
			grow = p.max - start
		}
		if grow <= 0 {
			return 0, ErrLimit
		}
		p.live = append(p.live, make([]bool, grow)...)
		p.allocated = len(p.live)
		for i := len(p.live) - 1; i >= start; i-- {
			p.free = append(p.free, i)
		}
	}
	n := len(p.free)
	h := p.free[n-1]
	p.free = p.free[:n-1]
	p.live[h] = true
	p.used++
	if p.used > p.peak {
		p.peak = p.used
	}
	return h, nil
}
func (p *FraigFixedPool) Recycle(h int) error {
	if p == nil || p.closed {
		return ErrInvariant
	}
	if h < 0 || h >= len(p.live) || !p.live[h] {
		return ErrDimension
	}
	p.live[h] = false
	p.free = append(p.free, h)
	p.used--
	return nil
}
func (p *FraigFixedPool) Restart() error {
	if p == nil || p.closed {
		return ErrInvariant
	}
	keep := p.chunkSize
	if keep > len(p.live) {
		keep = len(p.live)
	}
	p.live = make([]bool, keep)
	p.allocated = keep
	p.used = 0
	p.free = p.free[:0]
	for i := keep - 1; i >= 0; i-- {
		p.free = append(p.free, i)
	}
	return nil
}
func (p *FraigFixedPool) Stats() FraigPoolStats {
	if p == nil {
		return FraigPoolStats{}
	}
	maxInt := int(^uint(0) >> 1)
	memory := maxInt
	if p.allocated <= maxInt/p.entrySize {
		memory = p.allocated * p.entrySize
	}
	return FraigPoolStats{p.entrySize, p.chunkSize, p.allocated, p.used, p.peak, memory}
}
func (p *FraigFixedPool) Close() {
	if p != nil {
		p.closed = true
		p.live = nil
		p.free = nil
	}
}
