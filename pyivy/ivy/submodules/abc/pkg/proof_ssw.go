package abc

// Derived from ABC src/proof/ssw.

import (
	"errors"
	"fmt"
	"math/bits"
	"slices"
	"sort"
)

var ErrSSWClosed = errors.New("SSW machine is closed")

// These sizes are the retained 64-bit Go representations used by this port.
// Keeping them explicit makes byte-limit preflights independent of allocator
// growth and includes slice headers as well as their backing elements.
const (
	sswNodeBytes      = 40
	sswSignalBytes    = 16
	sswFrameBytes     = 96
	sswSignatureBytes = 32
	sswPatternBytes   = 48
	sswClassBytes     = 56
)

type SSWNodeKind uint8

const (
	SSWNodeConstant SSWNodeKind = iota
	SSWNodeInput
	SSWNodeState
	SSWNodeAnd
	SSWNodeXor
	SSWNodeBuffer
)

type SSWSignal struct {
	Node     int
	Inverted bool
}

func (s SSWSignal) Not() SSWSignal {
	s.Inverted = !s.Inverted
	return s
}

type SSWNode struct {
	Kind SSWNodeKind
	A    SSWSignal
	B    SSWSignal
}

type SSWOptions struct {
	MaxNodes      int
	MaxEdges      int
	MaxFrames     int
	MaxPatterns   int
	MaxClasses    int
	MaxOperations int
	MaxBytes      int
	MaxDepth      int
	Stop          func() bool
}

func (o SSWOptions) normalized() (SSWOptions, error) {
	if o.MaxNodes < 0 || o.MaxEdges < 0 || o.MaxFrames < 0 || o.MaxPatterns < 0 || o.MaxClasses < 0 || o.MaxOperations < 0 || o.MaxBytes < 0 || o.MaxDepth < 0 {
		return o, ErrDimension
	}
	if o.MaxNodes == 0 {
		o.MaxNodes = 1 << 20
	}
	if o.MaxEdges == 0 {
		o.MaxEdges = 1 << 21
	}
	if o.MaxFrames == 0 {
		o.MaxFrames = 64
	}
	if o.MaxPatterns == 0 {
		o.MaxPatterns = 4096
	}
	if o.MaxClasses == 0 {
		o.MaxClasses = 1 << 20
	}
	if o.MaxOperations == 0 {
		o.MaxOperations = 1 << 26
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 1 << 28
	}
	if o.MaxDepth == 0 {
		o.MaxDepth = 1 << 16
	}
	return o, nil
}

type sswBudget struct {
	options SSWOptions
	work    int
	bytes   int
}

func newSSWBudget(options SSWOptions) (*sswBudget, error) {
	o, err := options.normalized()
	if err != nil {
		return nil, err
	}
	return &sswBudget{options: o}, nil
}

func (b *sswBudget) spend(work, memory int) error {
	if work < 0 || memory < 0 || work > b.options.MaxOperations-b.work || memory > b.options.MaxBytes-b.bytes {
		return ErrLimit
	}
	b.work += work
	b.bytes += memory
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	return nil
}

type SSWMachine struct {
	nodes   []SSWNode
	inputs  []int
	states  []int
	outputs []SSWSignal
	next    []SSWSignal
	initial []bool
	edges   int
	closed  bool
}

func sswByteSum(values ...int) (int, error) {
	total := 0
	for _, value := range values {
		if value < 0 || value > int(^uint(0)>>1)-total {
			return 0, ErrLimit
		}
		total += value
	}
	return total, nil
}

func (m *SSWMachine) retainedBytesWith(nodeCount, outputCap, nextCap int) (int, error) {
	if m == nil || nodeCount < 0 || outputCap < 0 || nextCap < 0 || nodeCount > int(^uint(0)>>1)/sswNodeBytes || cap(m.inputs) > int(^uint(0)>>1)/8 || cap(m.states) > int(^uint(0)>>1)/8 || outputCap > int(^uint(0)>>1)/sswSignalBytes || nextCap > int(^uint(0)>>1)/sswSignalBytes {
		return 0, ErrLimit
	}
	return sswByteSum(nodeCount*sswNodeBytes, cap(m.inputs)*8, cap(m.states)*8, outputCap*sswSignalBytes, nextCap*sswSignalBytes, cap(m.initial))
}

func NewSSWMachine(inputCount, stateCount int, initial []bool, options SSWOptions) (*SSWMachine, error) {
	budget, err := newSSWBudget(options)
	if err != nil || inputCount < 0 || stateCount < 0 || len(initial) != stateCount {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if inputCount > budget.options.MaxNodes-stateCount-1 {
		return nil, ErrLimit
	}
	count := 1 + inputCount + stateCount
	if count > budget.options.MaxNodes || count > int(^uint(0)>>1)/sswNodeBytes {
		return nil, ErrLimit
	}
	if err = budget.spend(count, count*sswNodeBytes+stateCount*2+inputCount*8+stateCount*8); err != nil {
		return nil, err
	}
	m := &SSWMachine{nodes: make([]SSWNode, 1, count), initial: slices.Clone(initial)}
	m.nodes[0] = SSWNode{Kind: SSWNodeConstant, A: SSWSignal{Node: -1}, B: SSWSignal{Node: -1}}
	for i := 0; i < inputCount; i++ {
		m.inputs = append(m.inputs, len(m.nodes))
		m.nodes = append(m.nodes, SSWNode{Kind: SSWNodeInput, A: SSWSignal{Node: -1}, B: SSWSignal{Node: -1}})
	}
	for i := 0; i < stateCount; i++ {
		m.states = append(m.states, len(m.nodes))
		m.nodes = append(m.nodes, SSWNode{Kind: SSWNodeState, A: SSWSignal{Node: -1}, B: SSWSignal{Node: -1}})
	}
	m.next = make([]SSWSignal, stateCount)
	for i, id := range m.states {
		m.next[i] = SSWSignal{Node: id}
	}
	retained, retainedErr := m.retainedBytesWith(len(m.nodes), cap(m.outputs), cap(m.next))
	if retainedErr != nil || retained > budget.options.MaxBytes || stateCount > budget.options.MaxEdges {
		return nil, ErrLimit
	}
	return m, nil
}

func (m *SSWMachine) checkOpen() error {
	if m == nil || m.closed {
		return ErrSSWClosed
	}
	return nil
}

func (m *SSWMachine) InputCount() int {
	if m == nil || m.closed {
		return 0
	}
	return len(m.inputs)
}
func (m *SSWMachine) StateCount() int {
	if m == nil || m.closed {
		return 0
	}
	return len(m.states)
}
func (m *SSWMachine) NodeCount() int {
	if m == nil || m.closed {
		return 0
	}
	return len(m.nodes)
}
func (m *SSWMachine) OutputCount() int {
	if m == nil || m.closed {
		return 0
	}
	return len(m.outputs)
}

func (m *SSWMachine) Constant(value bool) SSWSignal { return SSWSignal{Node: 0, Inverted: value} }
func (m *SSWMachine) Input(index int) (SSWSignal, error) {
	if err := m.checkOpen(); err != nil {
		return SSWSignal{}, err
	}
	if index < 0 || index >= len(m.inputs) {
		return SSWSignal{}, ErrDimension
	}
	return SSWSignal{Node: m.inputs[index]}, nil
}
func (m *SSWMachine) State(index int) (SSWSignal, error) {
	if err := m.checkOpen(); err != nil {
		return SSWSignal{}, err
	}
	if index < 0 || index >= len(m.states) {
		return SSWSignal{}, ErrDimension
	}
	return SSWSignal{Node: m.states[index]}, nil
}

func (m *SSWMachine) validSignal(signal SSWSignal) bool {
	return signal.Node >= 0 && signal.Node < len(m.nodes)
}

func (m *SSWMachine) addNode(node SSWNode, options SSWOptions) (SSWSignal, error) {
	if err := m.checkOpen(); err != nil {
		return SSWSignal{}, err
	}
	budget, err := newSSWBudget(options)
	if err != nil {
		return SSWSignal{}, err
	}
	edges := 1
	if node.Kind == SSWNodeAnd || node.Kind == SSWNodeXor {
		edges = 2
	}
	if len(m.nodes) >= budget.options.MaxNodes || m.edges+len(m.outputs)+len(m.next) > budget.options.MaxEdges-edges || len(m.nodes) >= int(^uint(0)>>1)/sswNodeBytes {
		return SSWSignal{}, ErrLimit
	}
	retained, retainedErr := m.retainedBytesWith(len(m.nodes)+1, cap(m.outputs), cap(m.next))
	if retainedErr != nil || retained > budget.options.MaxBytes {
		return SSWSignal{}, ErrLimit
	}
	if !m.validSignal(node.A) || node.A.Node >= len(m.nodes) || (edges == 2 && (!m.validSignal(node.B) || node.B.Node >= len(m.nodes))) {
		return SSWSignal{}, ErrInvalidLiteral
	}
	if err = budget.spend(1+edges, sswNodeBytes); err != nil {
		return SSWSignal{}, err
	}
	id := len(m.nodes)
	// Keep the backing store exact so MaxBytes describes retained memory rather
	// than depending on the implementation's slice growth heuristic.
	nodes := make([]SSWNode, id+1)
	copy(nodes, m.nodes)
	nodes[id] = node
	m.nodes = nodes
	m.edges += edges
	return SSWSignal{Node: id}, nil
}

func (m *SSWMachine) AddAnd(a, b SSWSignal, options SSWOptions) (SSWSignal, error) {
	return m.addNode(SSWNode{Kind: SSWNodeAnd, A: a, B: b}, options)
}
func (m *SSWMachine) AddXor(a, b SSWSignal, options SSWOptions) (SSWSignal, error) {
	return m.addNode(SSWNode{Kind: SSWNodeXor, A: a, B: b}, options)
}
func (m *SSWMachine) AddBuffer(a SSWSignal, options SSWOptions) (SSWSignal, error) {
	return m.addNode(SSWNode{Kind: SSWNodeBuffer, A: a, B: SSWSignal{Node: -1}}, options)
}

func (m *SSWMachine) SetOutputs(outputs []SSWSignal, options SSWOptions) error {
	if err := m.checkOpen(); err != nil {
		return err
	}
	budget, err := newSSWBudget(options)
	if err != nil {
		return err
	}
	if len(outputs) > budget.options.MaxEdges || len(outputs) > int(^uint(0)>>1)/16 {
		return ErrLimit
	}
	if err = budget.spend(0, len(outputs)*sswSignalBytes); err != nil {
		return err
	}
	for _, output := range outputs {
		if err = budget.spend(1, 0); err != nil {
			return err
		}
		if !m.validSignal(output) {
			return ErrInvalidLiteral
		}
	}
	copyOutputs := slices.Clone(outputs)
	retained, retainedErr := m.retainedBytesWith(len(m.nodes), cap(copyOutputs), cap(m.next))
	if retainedErr != nil || retained > budget.options.MaxBytes || m.edges+len(copyOutputs)+len(m.next) > budget.options.MaxEdges {
		return ErrLimit
	}
	m.outputs = copyOutputs
	return nil
}

func (m *SSWMachine) SetNext(next []SSWSignal, options SSWOptions) error {
	if err := m.checkOpen(); err != nil {
		return err
	}
	if len(next) != len(m.states) {
		return ErrDimension
	}
	budget, err := newSSWBudget(options)
	if err != nil {
		return err
	}
	if len(next) > int(^uint(0)>>1)/sswSignalBytes {
		return ErrLimit
	}
	if err = budget.spend(0, len(next)*sswSignalBytes); err != nil {
		return err
	}
	for _, signal := range next {
		if err = budget.spend(1, 0); err != nil {
			return err
		}
		if !m.validSignal(signal) {
			return ErrInvalidLiteral
		}
	}
	copyNext := slices.Clone(next)
	retained, retainedErr := m.retainedBytesWith(len(m.nodes), cap(m.outputs), cap(copyNext))
	if retainedErr != nil || retained > budget.options.MaxBytes || m.edges+len(m.outputs)+len(copyNext) > budget.options.MaxEdges {
		return ErrLimit
	}
	m.next = copyNext
	return nil
}

func (m *SSWMachine) Node(id int) (SSWNode, error) {
	if err := m.checkOpen(); err != nil {
		return SSWNode{}, err
	}
	if id < 0 || id >= len(m.nodes) {
		return SSWNode{}, ErrDimension
	}
	return m.nodes[id], nil
}

func (m *SSWMachine) Validate(options SSWOptions) error {
	budget, err := newSSWBudget(options)
	if err != nil {
		return err
	}
	return m.validateWithBudget(budget)
}

func (m *SSWMachine) validateWithBudget(budget *sswBudget) error {
	if err := m.checkOpen(); err != nil {
		return err
	}
	retained, retainedErr := m.retainedBytesWith(len(m.nodes), cap(m.outputs), cap(m.next))
	if retainedErr != nil || len(m.nodes) > budget.options.MaxNodes || m.edges+len(m.outputs)+len(m.next) > budget.options.MaxEdges || retained > budget.options.MaxBytes {
		return ErrLimit
	}
	if err := budget.spend(0, retained); err != nil {
		return err
	}
	edges := 0
	depth := make([]int, len(m.nodes))
	if err := budget.spend(0, len(depth)*9); err != nil {
		return err
	}
	for id, node := range m.nodes {
		if err := budget.spend(1, 0); err != nil {
			return err
		}
		switch node.Kind {
		case SSWNodeConstant:
			if id != 0 || node.A.Node != -1 || node.B.Node != -1 {
				return ErrInvariant
			}
		case SSWNodeInput, SSWNodeState:
			if node.A.Node != -1 || node.B.Node != -1 {
				return ErrInvariant
			}
		case SSWNodeBuffer:
			if node.A.Node < 0 || node.A.Node >= id || node.B.Node != -1 {
				return ErrInvariant
			}
			edges++
			depth[id] = depth[node.A.Node] + 1
		case SSWNodeAnd, SSWNodeXor:
			if node.A.Node < 0 || node.A.Node >= id || node.B.Node < 0 || node.B.Node >= id {
				return ErrInvariant
			}
			edges += 2
			depth[id] = max(depth[node.A.Node], depth[node.B.Node]) + 1
		default:
			return ErrInvariant
		}
		if depth[id] > budget.options.MaxDepth {
			return ErrLimit
		}
	}
	if edges != m.edges || len(m.inputs)+len(m.states)+1 > len(m.nodes) {
		return ErrInvariant
	}
	partition := make([]SSWNodeKind, len(m.nodes))
	if err := budget.spend(0, len(partition)); err != nil {
		return err
	}
	for _, id := range m.inputs {
		if err := budget.spend(1, 0); err != nil {
			return err
		}
		if id <= 0 || id >= len(m.nodes) || partition[id] != SSWNodeConstant || m.nodes[id].Kind != SSWNodeInput {
			return ErrInvariant
		}
		partition[id] = SSWNodeInput
	}
	for _, id := range m.states {
		if err := budget.spend(1, 0); err != nil {
			return err
		}
		if id <= 0 || id >= len(m.nodes) || partition[id] != SSWNodeConstant || m.nodes[id].Kind != SSWNodeState {
			return ErrInvariant
		}
		partition[id] = SSWNodeState
	}
	for id := 1; id < len(m.nodes); id++ {
		if err := budget.spend(1, 0); err != nil {
			return err
		}
		if (m.nodes[id].Kind == SSWNodeInput || m.nodes[id].Kind == SSWNodeState) && partition[id] != m.nodes[id].Kind {
			return ErrInvariant
		}
	}
	for _, signal := range m.outputs {
		if err := budget.spend(1, 0); err != nil {
			return err
		}
		if !m.validSignal(signal) {
			return ErrInvariant
		}
	}
	for _, signal := range m.next {
		if err := budget.spend(1, 0); err != nil {
			return err
		}
		if !m.validSignal(signal) {
			return ErrInvariant
		}
	}
	if len(m.initial) != len(m.states) || len(m.next) != len(m.states) {
		return ErrInvariant
	}
	return nil
}

type SSWFrame struct {
	Inputs  []bool
	State   []bool
	Outputs []bool
	Next    []bool
}

type SSWTrace struct {
	Frames []SSWFrame
}

func (m *SSWMachine) evaluate(state, inputs []bool, budget *sswBudget) ([]bool, []bool, error) {
	if len(state) != len(m.states) || len(inputs) != len(m.inputs) {
		return nil, nil, ErrDimension
	}
	if err := budget.spend(len(m.nodes), len(m.nodes)+len(m.outputs)+len(m.next)); err != nil {
		return nil, nil, err
	}
	values := make([]bool, len(m.nodes))
	for i, id := range m.inputs {
		values[id] = inputs[i]
	}
	for i, id := range m.states {
		values[id] = state[i]
	}
	read := func(signal SSWSignal) bool {
		value := values[signal.Node]
		if signal.Inverted {
			return !value
		}
		return value
	}
	for id := 1; id < len(m.nodes); id++ {
		node := m.nodes[id]
		switch node.Kind {
		case SSWNodeAnd:
			values[id] = read(node.A) && read(node.B)
		case SSWNodeXor:
			values[id] = read(node.A) != read(node.B)
		case SSWNodeBuffer:
			values[id] = read(node.A)
		}
	}
	outputs := make([]bool, len(m.outputs))
	for i, signal := range m.outputs {
		outputs[i] = read(signal)
	}
	next := make([]bool, len(m.next))
	for i, signal := range m.next {
		next[i] = read(signal)
	}
	return outputs, next, nil
}

func (m *SSWMachine) Simulate(initial []bool, inputFrames [][]bool, options SSWOptions) (SSWTrace, error) {
	budget, err := newSSWBudget(options)
	if err != nil {
		return SSWTrace{}, err
	}
	if err = m.validateWithBudget(budget); err != nil {
		return SSWTrace{}, err
	}
	return m.simulateValidated(initial, inputFrames, budget)
}

func (m *SSWMachine) simulateValidated(initial []bool, inputFrames [][]bool, budget *sswBudget) (SSWTrace, error) {
	if len(initial) != len(m.states) {
		return SSWTrace{}, ErrDimension
	}
	if len(inputFrames) > budget.options.MaxFrames || len(inputFrames) > int(^uint(0)>>1)/64 {
		return SSWTrace{}, ErrLimit
	}
	for _, input := range inputFrames {
		if len(input) != len(m.inputs) {
			return SSWTrace{}, ErrDimension
		}
	}
	perFrame, sizeErr := sswByteSum(sswFrameBytes, len(m.inputs), len(m.states), len(m.outputs), len(m.next))
	if sizeErr != nil || len(inputFrames) != 0 && perFrame > budget.options.MaxBytes/len(inputFrames) {
		return SSWTrace{}, ErrLimit
	}
	if err := budget.spend(0, perFrame*len(inputFrames)); err != nil {
		return SSWTrace{}, err
	}
	if err := budget.spend(0, len(initial)); err != nil {
		return SSWTrace{}, err
	}
	state := slices.Clone(initial)
	trace := SSWTrace{Frames: make([]SSWFrame, 0, len(inputFrames))}
	for _, input := range inputFrames {
		outputs, next, evalErr := m.evaluate(state, input, budget)
		if evalErr != nil {
			return SSWTrace{}, evalErr
		}
		trace.Frames = append(trace.Frames, SSWFrame{Inputs: slices.Clone(input), State: slices.Clone(state), Outputs: outputs, Next: next})
		state = next
	}
	return trace, nil
}

func (m *SSWMachine) InitialState() []bool {
	if m == nil || m.closed {
		return nil
	}
	return slices.Clone(m.initial)
}

func (m *SSWMachine) Snapshot(options SSWOptions) ([]SSWNode, []SSWSignal, []SSWSignal, error) {
	budget, err := newSSWBudget(options)
	if err != nil {
		return nil, nil, nil, err
	}
	if err = m.validateWithBudget(budget); err != nil {
		return nil, nil, nil, err
	}
	memory := len(m.nodes)*sswNodeBytes + (len(m.outputs)+len(m.next))*sswSignalBytes
	if err = budget.spend(len(m.nodes)+len(m.outputs)+len(m.next), memory); err != nil {
		return nil, nil, nil, err
	}
	return slices.Clone(m.nodes), slices.Clone(m.outputs), slices.Clone(m.next), nil
}

func (m *SSWMachine) Close() {
	if m == nil || m.closed {
		return
	}
	m.nodes = nil
	m.inputs = nil
	m.states = nil
	m.outputs = nil
	m.next = nil
	m.initial = nil
	m.edges = 0
	m.closed = true
}

type SSWPattern struct {
	State  []bool
	Inputs []bool
}

type SSWSignature struct {
	Node  int
	Words []uint64
}

func (m *SSWMachine) Signatures(patterns []SSWPattern, options SSWOptions) ([]SSWSignature, error) {
	budget, err := newSSWBudget(options)
	if err != nil {
		return nil, err
	}
	if err = m.validateWithBudget(budget); err != nil {
		return nil, err
	}
	return m.signaturesWithBudget(patterns, budget)
}

func (m *SSWMachine) signaturesWithBudget(patterns []SSWPattern, budget *sswBudget) ([]SSWSignature, error) {
	if len(patterns) > budget.options.MaxPatterns || len(patterns) > int(^uint(0)>>1)-63 {
		return nil, ErrLimit
	}
	words := (len(patterns) + 63) / 64
	if words > int(^uint(0)>>1)/8 {
		return nil, ErrLimit
	}
	perNode, perNodeErr := sswByteSum(words*8, sswSignatureBytes)
	if perNodeErr != nil || perNode != 0 && len(m.nodes) > budget.options.MaxBytes/perNode {
		return nil, ErrLimit
	}
	if len(patterns) != 0 && len(m.nodes) > int(^uint(0)>>1)/len(patterns) {
		return nil, ErrLimit
	}
	if err := budget.spend(len(patterns)*len(m.nodes), len(m.nodes)*perNode); err != nil {
		return nil, err
	}
	result := make([]SSWSignature, len(m.nodes))
	for id := range result {
		result[id] = SSWSignature{Node: id, Words: make([]uint64, words)}
	}
	for p, pattern := range patterns {
		values, evalErr := m.valuesForPattern(pattern, budget)
		if evalErr != nil {
			return nil, evalErr
		}
		for id, value := range values {
			if value {
				result[id].Words[p>>6] |= uint64(1) << uint(p&63)
			}
		}
	}
	return result, nil
}

func (m *SSWMachine) valuesForPattern(pattern SSWPattern, budget *sswBudget) ([]bool, error) {
	if len(pattern.State) != len(m.states) || len(pattern.Inputs) != len(m.inputs) {
		return nil, ErrDimension
	}
	if err := budget.spend(len(m.nodes), len(m.nodes)); err != nil {
		return nil, err
	}
	values := make([]bool, len(m.nodes))
	for i, id := range m.inputs {
		values[id] = pattern.Inputs[i]
	}
	for i, id := range m.states {
		values[id] = pattern.State[i]
	}
	read := func(s SSWSignal) bool {
		v := values[s.Node]
		if s.Inverted {
			return !v
		}
		return v
	}
	for id := 1; id < len(m.nodes); id++ {
		n := m.nodes[id]
		switch n.Kind {
		case SSWNodeAnd:
			values[id] = read(n.A) && read(n.B)
		case SSWNodeXor:
			values[id] = read(n.A) != read(n.B)
		case SSWNodeBuffer:
			values[id] = read(n.A)
		}
	}
	return values, nil
}

type SSWClass struct {
	Representative int
	Members        []int
	Inverted       []bool
}

func (m *SSWMachine) CandidateClasses(patterns []SSWPattern, options SSWOptions) ([]SSWClass, error) {
	budget, err := newSSWBudget(options)
	if err != nil {
		return nil, err
	}
	if err = m.validateWithBudget(budget); err != nil {
		return nil, err
	}
	return m.candidateClassesWithBudget(patterns, budget)
}

func (m *SSWMachine) candidateClassesWithBudget(patterns []SSWPattern, budget *sswBudget) ([]SSWClass, error) {
	signatures, err := m.signaturesWithBudget(patterns, budget)
	if err != nil {
		return nil, err
	}
	maskLast := ^uint64(0)
	if len(patterns)&63 != 0 {
		maskLast = (uint64(1) << uint(len(patterns)&63)) - 1
	}
	type entry struct {
		id       int
		inverted bool
	}
	groups := make(map[string][]entry)
	keys := make([]string, 0)
	words := 0
	if len(signatures) != 0 {
		words = len(signatures[0].Words)
	}
	// Each node may retain one canonical word copy, textual key, map entry,
	// key-slice entry, and class member/phase pair. Preflight the maximum so
	// no grouping allocation can escape the cumulative byte limit.
	perSignature, sizeErr := sswByteSum(96, words*24)
	if sizeErr != nil || perSignature != 0 && len(signatures) > budget.options.MaxBytes/perSignature {
		return nil, ErrLimit
	}
	classMemory, sizeErr := sswByteSum(len(signatures)*perSignature, len(signatures)*9, len(signatures)*sswClassBytes)
	if sizeErr != nil {
		return nil, sizeErr
	}
	if err = budget.spend(len(signatures), classMemory); err != nil {
		return nil, err
	}
	for _, sig := range signatures {
		if sig.Node == 0 {
			continue
		}
		words := slices.Clone(sig.Words)
		inverted := false
		if len(words) != 0 && words[0]&1 != 0 {
			inverted = true
			for i := range words {
				words[i] = ^words[i]
			}
			words[len(words)-1] &= maskLast
		}
		key := fmt.Sprint(words)
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], entry{id: sig.Node, inverted: inverted})
	}
	sort.Strings(keys)
	classes := make([]SSWClass, 0)
	for _, key := range keys {
		entries := groups[key]
		if len(entries) < 2 {
			continue
		}
		if len(classes) >= budget.options.MaxClasses {
			return nil, ErrLimit
		}
		class := SSWClass{Representative: entries[0].id, Members: make([]int, len(entries)), Inverted: make([]bool, len(entries))}
		for i, e := range entries {
			class.Members[i] = e.id
			class.Inverted[i] = e.inverted != entries[0].inverted
		}
		classes = append(classes, class)
	}
	return classes, nil
}

type SSWCounterexample struct {
	Initial []bool
	Inputs  [][]bool
	Frame   int
	Output  int
	Left    bool
	Right   bool
}

type SSWEquivalenceResult struct {
	Equivalent     bool
	Counterexample *SSWCounterexample
	CheckedCases   uint64
}

func CheckSSWEquivalent(left, right *SSWMachine, frames int, options SSWOptions) (SSWEquivalenceResult, error) {
	if left == nil || right == nil || frames < 0 {
		return SSWEquivalenceResult{}, ErrDimension
	}
	budget, err := newSSWBudget(options)
	if err != nil {
		return SSWEquivalenceResult{}, err
	}
	if err = left.validateWithBudget(budget); err != nil {
		return SSWEquivalenceResult{}, err
	}
	if err = right.validateWithBudget(budget); err != nil {
		return SSWEquivalenceResult{}, err
	}
	if left.InputCount() != right.InputCount() || left.StateCount() != right.StateCount() || left.OutputCount() != right.OutputCount() {
		return SSWEquivalenceResult{}, ErrDimension
	}
	if frames > budget.options.MaxFrames {
		return SSWEquivalenceResult{}, ErrLimit
	}
	maxInt := int(^uint(0) >> 1)
	if frames > maxInt/24 || left.InputCount() != 0 && frames > (maxInt-left.StateCount())/left.InputCount() {
		return SSWEquivalenceResult{}, ErrLimit
	}
	bitsTotal := left.StateCount() + left.InputCount()*frames
	if bitsTotal > 24 || bitsTotal >= bits.UintSize {
		return SSWEquivalenceResult{}, ErrLimit
	}
	cases := uint64(1) << uint(bitsTotal)
	result := SSWEquivalenceResult{Equivalent: true}
	for assignment := uint64(0); assignment < cases; assignment++ {
		caseBytes, sizeErr := sswByteSum(left.StateCount(), frames*24, frames*left.InputCount())
		if sizeErr != nil {
			return SSWEquivalenceResult{}, sizeErr
		}
		if err = budget.spend(1, caseBytes); err != nil {
			return SSWEquivalenceResult{}, err
		}
		initial := make([]bool, left.StateCount())
		for i := range initial {
			initial[i] = assignment&(uint64(1)<<uint(i)) != 0
		}
		inputs := make([][]bool, frames)
		offset := left.StateCount()
		for f := 0; f < frames; f++ {
			inputs[f] = make([]bool, left.InputCount())
			for i := range inputs[f] {
				inputs[f][i] = assignment&(uint64(1)<<uint(offset+f*left.InputCount()+i)) != 0
			}
		}
		lt, e := left.simulateValidated(initial, inputs, budget)
		if e != nil {
			return SSWEquivalenceResult{}, e
		}
		rt, e := right.simulateValidated(initial, inputs, budget)
		if e != nil {
			return SSWEquivalenceResult{}, e
		}
		result.CheckedCases++
		for f := range lt.Frames {
			for o := range lt.Frames[f].Outputs {
				if lt.Frames[f].Outputs[o] != rt.Frames[f].Outputs[o] {
					result.Equivalent = false
					result.Counterexample = &SSWCounterexample{Initial: initial, Inputs: inputs, Frame: f, Output: o, Left: lt.Frames[f].Outputs[o], Right: rt.Frames[f].Outputs[o]}
					return result, nil
				}
			}
		}
	}
	return result, nil
}

type SSWSweepResult struct {
	Representative []int
	Inverted       []bool
	Classes        []SSWClass
	VerifiedPairs  int
}

func (m *SSWMachine) Sweep(patterns []SSWPattern, options SSWOptions) (SSWSweepResult, error) {
	budget, err := newSSWBudget(options)
	if err != nil {
		return SSWSweepResult{}, err
	}
	if err = m.validateWithBudget(budget); err != nil {
		return SSWSweepResult{}, err
	}
	classes, err := m.candidateClassesWithBudget(patterns, budget)
	if err != nil {
		return SSWSweepResult{}, err
	}
	variables := len(m.inputs) + len(m.states)
	if variables > 20 || variables >= bits.UintSize {
		return SSWSweepResult{}, ErrLimit
	}
	count := uint64(1) << uint(variables)
	classMembers := 0
	for _, class := range classes {
		if classMembers > int(^uint(0)>>1)-len(class.Members) {
			return SSWSweepResult{}, ErrLimit
		}
		classMembers += len(class.Members)
	}
	sweepMemory, sizeErr := sswByteSum(len(m.nodes)*9, len(classes)*sswClassBytes, classMembers*9, len(m.inputs), len(m.states))
	if sizeErr != nil {
		return SSWSweepResult{}, sizeErr
	}
	if err = budget.spend(0, sweepMemory); err != nil {
		return SSWSweepResult{}, err
	}
	representative := make([]int, len(m.nodes))
	inverted := make([]bool, len(m.nodes))
	for i := range representative {
		representative[i] = i
	}
	verified := make([]SSWClass, 0, len(classes))
	pattern := SSWPattern{State: make([]bool, len(m.states)), Inputs: make([]bool, len(m.inputs))}
	for _, class := range classes {
		kept := SSWClass{Representative: class.Representative, Members: []int{class.Representative}, Inverted: []bool{false}}
		for i := 1; i < len(class.Members); i++ {
			member := class.Members[i]
			phase := class.Inverted[i]
			equal := true
			for assignment := uint64(0); assignment < count; assignment++ {
				if err = budget.spend(1, 0); err != nil {
					return SSWSweepResult{}, err
				}
				for j := range pattern.Inputs {
					pattern.Inputs[j] = assignment&(uint64(1)<<uint(j)) != 0
				}
				for j := range pattern.State {
					pattern.State[j] = assignment&(uint64(1)<<uint(len(m.inputs)+j)) != 0
				}
				values, e := m.valuesForPattern(pattern, budget)
				if e != nil {
					return SSWSweepResult{}, e
				}
				if values[member] != (values[class.Representative] != phase) {
					equal = false
					break
				}
			}
			if equal {
				representative[member] = class.Representative
				inverted[member] = phase
				kept.Members = append(kept.Members, member)
				kept.Inverted = append(kept.Inverted, phase)
			}
		}
		if len(kept.Members) > 1 {
			verified = append(verified, kept)
		}
	}
	return SSWSweepResult{Representative: representative, Inverted: inverted, Classes: verified, VerifiedPairs: len(m.nodes) - countIdentity(representative)}, nil
}

func countIdentity(values []int) int {
	n := 0
	for i, v := range values {
		if i == v {
			n++
		}
	}
	return n
}

func ReplaySSWCounterexample(left, right *SSWMachine, cex *SSWCounterexample, options SSWOptions) (bool, error) {
	if left == nil || right == nil || cex == nil || cex.Frame < 0 || cex.Output < 0 || cex.Frame >= len(cex.Inputs) {
		return false, ErrDimension
	}
	budget, err := newSSWBudget(options)
	if err != nil {
		return false, err
	}
	if err = left.validateWithBudget(budget); err != nil {
		return false, err
	}
	if err = right.validateWithBudget(budget); err != nil {
		return false, err
	}
	lt, err := left.simulateValidated(cex.Initial, cex.Inputs, budget)
	if err != nil {
		return false, err
	}
	rt, err := right.simulateValidated(cex.Initial, cex.Inputs, budget)
	if err != nil {
		return false, err
	}
	if cex.Frame >= len(lt.Frames) || cex.Frame >= len(rt.Frames) || cex.Output >= len(lt.Frames[cex.Frame].Outputs) || cex.Output >= len(rt.Frames[cex.Frame].Outputs) {
		return false, ErrDimension
	}
	lv := lt.Frames[cex.Frame].Outputs[cex.Output]
	rv := rt.Frames[cex.Frame].Outputs[cex.Output]
	return lv == cex.Left && rv == cex.Right && lv != rv, nil
}

func SSWDeterministicPatterns(inputCount, stateCount, count int, seed uint64, options SSWOptions) ([]SSWPattern, error) {
	budget, err := newSSWBudget(options)
	if err != nil {
		return nil, err
	}
	if inputCount < 0 || stateCount < 0 || count < 0 {
		return nil, ErrDimension
	}
	perPattern, sizeErr := sswByteSum(sswPatternBytes, inputCount, stateCount)
	if sizeErr != nil || count > budget.options.MaxPatterns || count != 0 && perPattern > budget.options.MaxBytes/count {
		return nil, ErrLimit
	}
	if err = budget.spend(0, count*perPattern); err != nil {
		return nil, err
	}
	if seed == 0 {
		seed = 0x9e3779b97f4a7c15
	}
	patterns := make([]SSWPattern, count)
	for p := range patterns {
		patterns[p] = SSWPattern{Inputs: make([]bool, inputCount), State: make([]bool, stateCount)}
		for i := 0; i < inputCount+stateCount; i++ {
			seed ^= seed << 7
			seed ^= seed >> 9
			seed ^= seed << 8
			if i < inputCount {
				patterns[p].Inputs[i] = seed&1 != 0
			} else {
				patterns[p].State[i-inputCount] = seed&1 != 0
			}
		}
		if err = budget.spend(inputCount+stateCount, 0); err != nil {
			return nil, err
		}
	}
	return patterns, nil
}

func SSWBuildInventory() []string {
	return []string{"sswAig.c", "sswBmc.c", "sswClass.c", "sswCnf.c", "sswConstr.c", "sswCore.c", "sswDyn.c", "sswFilter.c", "sswIslands.c", "sswLcorr.c", "sswMan.c", "sswPart.c", "sswPairs.c", "sswRarity.c", "sswSat.c", "sswSemi.c", "sswSim.c", "sswSimSat.c", "sswSweep.c", "sswUnique.c"}
}

func SSWTrackedInventory() []string {
	return []string{"module.make", "ssw.h", "sswAig.c", "sswBmc.c", "sswClass.c", "sswCnf.c", "sswConstr.c", "sswCore.c", "sswDyn.c", "sswFilter.c", "sswInt.h", "sswIslands.c", "sswLcorr.c", "sswMan.c", "sswPairs.c", "sswPart.c", "sswRarity.c", "sswRarity2.c", "sswSat.c", "sswSemi.c", "sswSim.c", "sswSimSat.c", "sswSweep.c", "sswUnique.c"}
}
