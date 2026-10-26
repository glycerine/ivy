package abc

// Port of the standalone abstraction and refinement semantics represented by
// src/proof/abs. Native GIA/AIG pointer managers, solver proof objects, threads,
// command adapters, and process I/O are outside this concrete owned model.

import (
	"fmt"
	"slices"
	"sort"
)

// AbsParameters mirrors the public defaults in Abs_ParSetDefaults.
type AbsParameters struct {
	FramesMax, FramesStart, FramesPast                          int
	ConflictLimit, LearnedMax, LearnedStart, LearnedDelta       int
	LearnedPercent, Timeout, RatioMin, RatioMinRefine, RatioMax int
	UseTerminalVariables, UseRollback, PropagateFanout          bool
	AddLayer, NewRefinement, UseSkip, UseSimple, SkipHash       bool
	UseFullProof, DumpVABS, DumpMABS, CallProver, Simplify      bool
	Frame, FrameProved, FramesNoChange, FramesNoChangeLimit     int
}

func DefaultAbsParameters() AbsParameters {
	return AbsParameters{
		FramesPast: 4, LearnedMax: 1000, LearnedStart: 1000,
		LearnedDelta: 200, LearnedPercent: 70, RatioMax: 30,
		PropagateFanout: true, Frame: -1, FrameProved: -1,
		FramesNoChangeLimit: 2,
	}
}

type AbsOptions struct {
	MaxStateBits, MaxInputBits, MaxConcreteStates int
	MaxAbstractStates, MaxTransitions, MaxFrames  int
	MaxRefinements, MaxTraceStates, MaxDepth      int
	MaxNodes, MaxClauses                          int
	MaxBytes, MaxWork                             uint64
	Stop                                          func() bool
}

func (o AbsOptions) normalized() (AbsOptions, error) {
	if o.MaxStateBits < 0 || o.MaxInputBits < 0 || o.MaxConcreteStates < 0 ||
		o.MaxAbstractStates < 0 || o.MaxTransitions < 0 || o.MaxFrames < 0 ||
		o.MaxRefinements < 0 || o.MaxTraceStates < 0 || o.MaxDepth < 0 ||
		o.MaxNodes < 0 || o.MaxClauses < 0 ||
		o.MaxBytes > uint64(^uint(0)>>1) {
		return o, ErrDimension
	}
	if o.MaxStateBits == 0 {
		o.MaxStateBits = 20
	}
	if o.MaxInputBits == 0 {
		o.MaxInputBits = 12
	}
	if o.MaxConcreteStates == 0 {
		o.MaxConcreteStates = 1 << 20
	}
	if o.MaxAbstractStates == 0 {
		o.MaxAbstractStates = 1 << 20
	}
	if o.MaxTransitions == 0 {
		o.MaxTransitions = 1 << 22
	}
	if o.MaxFrames == 0 {
		o.MaxFrames = 256
	}
	if o.MaxRefinements == 0 {
		o.MaxRefinements = o.MaxStateBits + 1
	}
	if o.MaxTraceStates == 0 {
		o.MaxTraceStates = o.MaxFrames + 1
	}
	if o.MaxDepth == 0 {
		o.MaxDepth = o.MaxFrames + 1
	}
	if o.MaxNodes == 0 {
		o.MaxNodes = o.MaxConcreteStates
	}
	if o.MaxClauses == 0 {
		o.MaxClauses = o.MaxTransitions
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 256 << 20
	}
	if o.MaxWork == 0 {
		o.MaxWork = 1 << 30
	}
	return o, nil
}

type AbsSystem struct {
	StateBits, InputBits int
	Initial              []uint64
	Next                 []uint64
	Bad                  []bool
}

func NewAbsSystem(stateBits, inputBits int, initial, next []uint64, bad []bool, options AbsOptions) (*AbsSystem, error) {
	o, err := options.normalized()
	if err != nil {
		return nil, err
	}
	if stateBits < 0 || inputBits < 0 || stateBits > o.MaxStateBits || inputBits > o.MaxInputBits || stateBits >= 63 || inputBits >= 63 || stateBits+inputBits >= 63 {
		return nil, ErrDimension
	}
	states, inputs := uint64(1)<<stateBits, uint64(1)<<inputBits
	if states > uint64(o.MaxConcreteStates) || states > uint64(o.MaxNodes) || states > uint64(^uint(0)>>1) || inputs > uint64(^uint(0)>>1) || states > ^uint64(0)/inputs {
		return nil, ErrLimit
	}
	entries := states * inputs
	if entries > uint64(o.MaxClauses) {
		return nil, ErrLimit
	}
	if entries > uint64(^uint(0)>>1) || len(next) != int(entries) || len(bad) != int(states) || len(initial) == 0 {
		return nil, ErrDimension
	}
	// At peak, construction owns the cloned transition/bad tables together
	// with a state bitmap and the canonical initial-state slice.
	temporary, ok := absMulAdd(states, 10, 0)
	if !ok {
		return nil, ErrLimit
	}
	bytes, ok := absMulAdd(entries, 8, temporary)
	if !ok || bytes > o.MaxBytes {
		return nil, ErrLimit
	}
	work, ok := absAdd(uint64(len(initial)), uint64(len(next)))
	if !ok || work > ^uint64(0)-states || work+states > o.MaxWork {
		return nil, ErrLimit
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	seen := make([]bool, int(states))
	for _, state := range initial {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		if state >= states {
			return nil, ErrInvariant
		}
		seen[state] = true
	}
	for _, state := range next {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		if state >= states {
			return nil, ErrInvariant
		}
	}
	// Initial may contain arbitrarily many duplicates. Cap the canonical form by
	// the concrete state count so caller-owned duplication cannot amplify owned
	// memory beyond the construction preflight.
	unique := make([]uint64, 0, int(states))
	for state, present := range seen {
		if present {
			unique = append(unique, uint64(state))
		}
	}
	return &AbsSystem{StateBits: stateBits, InputBits: inputBits, Initial: unique, Next: slices.Clone(next), Bad: slices.Clone(bad)}, nil
}

func (s *AbsSystem) Clone() *AbsSystem {
	if s == nil {
		return nil
	}
	return &AbsSystem{StateBits: s.StateBits, InputBits: s.InputBits, Initial: slices.Clone(s.Initial), Next: slices.Clone(s.Next), Bad: slices.Clone(s.Bad)}
}

func (s *AbsSystem) NextState(state, input uint64) (uint64, error) {
	if s == nil || s.StateBits < 0 || s.InputBits < 0 || s.StateBits >= 63 || s.InputBits >= 63 {
		return 0, ErrInvariant
	}
	states, inputs := uint64(1)<<s.StateBits, uint64(1)<<s.InputBits
	if state >= states || input >= inputs || states > ^uint64(0)/inputs {
		return 0, ErrDimension
	}
	index := state*inputs + input
	if index >= uint64(len(s.Next)) {
		return 0, ErrInvariant
	}
	return s.Next[index], nil
}

type AbsStatus uint8

const (
	AbsUnknown AbsStatus = iota
	AbsProved
	AbsCounterexample
)

type AbsStopReason uint8

const (
	AbsFixedPoint AbsStopReason = iota
	AbsRealCounterexample
	AbsFrameLimit
	AbsRefinementLimit
	AbsCancelled
	AbsResourceLimit
)

type AbsTrace struct {
	States         []uint64
	AbstractStates []uint64
	Inputs         []uint64
}

func (t AbsTrace) Clone() AbsTrace {
	return AbsTrace{States: slices.Clone(t.States), AbstractStates: slices.Clone(t.AbstractStates), Inputs: slices.Clone(t.Inputs)}
}

type AbsStats struct {
	Refinements, Frames, AbstractStates, Transitions int
	Nodes, Clauses, States, Depth                    int
	ReplayAttempts, SpuriousTraces                   int
	Work, PeakBytes                                  uint64
}

type AbsPhase uint8

const (
	AbsReady AbsPhase = iota
	AbsSearching
	AbsFinished
	AbsClosed
)

type AbsResult struct {
	Status      AbsStatus
	Reason      AbsStopReason
	VisibleBits []int
	Trace       *AbsTrace
	Stats       AbsStats
}

type AbsManager struct {
	system   *AbsSystem
	visible  []bool
	options  AbsOptions
	retained uint64
	closed   bool
	phase    AbsPhase
	stats    AbsStats
}

func NewAbsManager(system *AbsSystem, visibleBits []int, options AbsOptions) (*AbsManager, error) {
	o, err := options.normalized()
	if err != nil {
		return nil, err
	}
	if system == nil {
		return nil, ErrInvariant
	}
	copySystem, err := NewAbsSystem(system.StateBits, system.InputBits, system.Initial, system.Next, system.Bad, o)
	if err != nil {
		return nil, err
	}
	visible := make([]bool, system.StateBits)
	for _, bit := range visibleBits {
		if bit < 0 || bit >= system.StateBits || visible[bit] {
			return nil, ErrDimension
		}
		visible[bit] = true
	}
	words, ok := absAdd(uint64(len(copySystem.Initial)), uint64(len(copySystem.Next)))
	if !ok {
		return nil, ErrLimit
	}
	tail, ok := absAdd(uint64(len(copySystem.Bad)), uint64(len(visible)))
	if !ok {
		return nil, ErrLimit
	}
	retained, ok := absMulAdd(words, 8, tail)
	if !ok || retained > o.MaxBytes {
		return nil, ErrLimit
	}
	return &AbsManager{
		system: copySystem, visible: visible, options: o, phase: AbsReady,
		retained: retained,
		stats:    AbsStats{Nodes: len(copySystem.Bad), Clauses: len(copySystem.Next), PeakBytes: retained},
	}, nil
}

func (m *AbsManager) Close() {
	if m == nil || m.closed {
		return
	}
	m.closed = true
	m.phase = AbsClosed
	m.system = nil
	clear(m.visible)
	m.visible = nil
	m.retained = 0
}

func (m *AbsManager) Phase() AbsPhase {
	if m == nil {
		return AbsClosed
	}
	return m.phase
}

func (m *AbsManager) VisibleBits() []int {
	if m == nil || m.closed {
		return nil
	}
	return absVisibleList(m.visible)
}

func (m *AbsManager) Stats() AbsStats {
	if m == nil {
		return AbsStats{}
	}
	return m.stats
}

func (m *AbsManager) check() error {
	if m == nil || m.closed || m.system == nil || len(m.visible) != m.system.StateBits {
		return ErrInvariant
	}
	if m.options.Stop != nil && m.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (m *AbsManager) work(amount uint64) error {
	if m.options.Stop != nil && m.options.Stop() {
		return ErrCancelled
	}
	if amount > m.options.MaxWork-m.stats.Work {
		return ErrLimit
	}
	m.stats.Work += amount
	return nil
}

func (m *AbsManager) bytes(amount uint64) error {
	if amount > ^uint64(0)-m.retained || m.retained+amount > m.options.MaxBytes {
		return ErrLimit
	}
	total := m.retained + amount
	if total > m.stats.PeakBytes {
		m.stats.PeakBytes = total
	}
	return nil
}

func (m *AbsManager) project(state uint64) uint64 {
	var result uint64
	position := uint(0)
	for bit, visible := range m.visible {
		if visible {
			result |= ((state >> uint(bit)) & 1) << position
			position++
		}
	}
	return result
}

func (m *AbsManager) Refine(bits []int) error {
	if err := m.check(); err != nil {
		return err
	}
	if m.stats.Refinements >= m.options.MaxRefinements {
		return ErrLimit
	}
	if err := m.bytes(uint64(len(m.visible))); err != nil {
		return err
	}
	candidate := slices.Clone(m.visible)
	changed := false
	for _, bit := range bits {
		if err := m.work(1); err != nil {
			return err
		}
		if bit < 0 || bit >= len(candidate) {
			return ErrDimension
		}
		if !candidate[bit] {
			candidate[bit] = true
			changed = true
		}
	}
	if !changed {
		return ErrInvariant
	}
	m.visible = candidate
	m.stats.Refinements++
	return nil
}

type absParent struct {
	state uint64
	input uint64
}

func (m *AbsManager) abstractSearch() (AbsStatus, AbsStopReason, *AbsTrace, error) {
	states := uint64(1) << m.system.StateBits
	inputs := uint64(1) << m.system.InputBits
	// Conservatively bound group payloads, maps, BFS state, parent records,
	// and the largest per-state target set before allocating those structures.
	temporary, ok := absMulAdd(states, 112, 0)
	if !ok || m.bytes(temporary) != nil {
		return AbsUnknown, AbsResourceLimit, nil, ErrLimit
	}
	groups := make(map[uint64][]uint64)
	for state := uint64(0); state < states; state++ {
		if err := m.work(1); err != nil {
			return AbsUnknown, AbsResourceLimit, nil, err
		}
		a := m.project(state)
		groups[a] = append(groups[a], state)
	}
	if len(groups) > m.options.MaxAbstractStates {
		return AbsUnknown, AbsResourceLimit, nil, ErrLimit
	}
	frontier := make([]uint64, 0, len(m.system.Initial))
	parents := make(map[uint64]absParent)
	depth := make(map[uint64]int)
	for _, initial := range m.system.Initial {
		a := m.project(initial)
		if _, seen := depth[a]; !seen {
			depth[a] = 0
			frontier = append(frontier, a)
		}
	}
	m.stats.AbstractStates = len(depth)
	if len(depth) > m.stats.States {
		m.stats.States = len(depth)
	}
	sort.Slice(frontier, func(i, j int) bool { return frontier[i] < frontier[j] })
	for len(frontier) > 0 {
		current := frontier[0]
		frontier = frontier[1:]
		frame := depth[current]
		if frame+1 > m.stats.Frames {
			m.stats.Frames = frame + 1
		}
		if frame > m.stats.Depth {
			m.stats.Depth = frame
		}
		for _, concrete := range groups[current] {
			if m.system.Bad[concrete] {
				if frame+1 > m.options.MaxTraceStates {
					return AbsUnknown, AbsResourceLimit, nil, ErrLimit
				}
				trace := m.abstractTrace(current, parents)
				return AbsCounterexample, AbsRealCounterexample, trace, nil
			}
		}
		if frame >= m.options.MaxFrames {
			return AbsUnknown, AbsFrameLimit, nil, nil
		}
		if frame >= m.options.MaxDepth {
			return AbsUnknown, AbsResourceLimit, nil, ErrLimit
		}
		targets := make(map[uint64]uint64)
		for input := uint64(0); input < inputs; input++ {
			for _, concrete := range groups[current] {
				if err := m.work(1); err != nil {
					return AbsUnknown, AbsResourceLimit, nil, err
				}
				next := m.system.Next[concrete*inputs+input]
				target := m.project(next)
				if prior, exists := targets[target]; !exists || input < prior {
					targets[target] = input
				}
			}
		}
		ordered := make([]uint64, 0, len(targets))
		for target := range targets {
			ordered = append(ordered, target)
		}
		sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
		for _, target := range ordered {
			m.stats.Transitions++
			if m.stats.Transitions > m.options.MaxTransitions {
				return AbsUnknown, AbsResourceLimit, nil, ErrLimit
			}
			if _, seen := depth[target]; seen {
				continue
			}
			depth[target] = frame + 1
			parents[target] = absParent{state: current, input: targets[target]}
			frontier = append(frontier, target)
			if len(depth) > m.options.MaxAbstractStates {
				return AbsUnknown, AbsResourceLimit, nil, ErrLimit
			}
		}
		m.stats.AbstractStates = len(depth)
		if len(depth) > m.stats.States {
			m.stats.States = len(depth)
		}
	}
	return AbsProved, AbsFixedPoint, nil, nil
}

func (m *AbsManager) abstractTrace(last uint64, parents map[uint64]absParent) *AbsTrace {
	states := []uint64{last}
	inputs := make([]uint64, 0)
	for {
		parent, exists := parents[last]
		if !exists {
			break
		}
		inputs = append(inputs, parent.input)
		last = parent.state
		states = append(states, last)
	}
	slices.Reverse(states)
	slices.Reverse(inputs)
	return &AbsTrace{AbstractStates: states, Inputs: inputs}
}

// Replay checks a trace against the concrete transition relation and returns a
// deterministic concrete witness when it reaches a bad state.
func (m *AbsManager) Replay(trace AbsTrace) (AbsTrace, bool, error) {
	if err := m.check(); err != nil {
		return AbsTrace{}, false, err
	}
	m.stats.ReplayAttempts++
	if len(trace.AbstractStates) == 0 || len(trace.Inputs)+1 != len(trace.AbstractStates) || len(trace.AbstractStates) > m.options.MaxTraceStates {
		return AbsTrace{}, false, ErrDimension
	}
	visibleCount := len(absVisibleList(m.visible))
	abstractStates := uint64(1) << visibleCount
	for _, state := range trace.AbstractStates {
		if state >= abstractStates {
			return AbsTrace{}, false, ErrDimension
		}
	}
	states := uint64(1) << m.system.StateBits
	perState, ok := absMulAdd(uint64(len(trace.AbstractStates)), 8, 64)
	if !ok {
		return AbsTrace{}, false, ErrLimit
	}
	temporary, ok := absMulAdd(states, perState, 0)
	if !ok || m.bytes(temporary) != nil {
		return AbsTrace{}, false, ErrLimit
	}
	current := make(map[uint64][]uint64)
	for _, state := range m.system.Initial {
		if m.project(state) == trace.AbstractStates[0] {
			current[state] = []uint64{state}
		}
	}
	inputs := uint64(1) << m.system.InputBits
	for frame, input := range trace.Inputs {
		if input >= inputs || frame+1 > m.options.MaxDepth {
			return AbsTrace{}, false, ErrDimension
		}
		nextSet := make(map[uint64][]uint64)
		ordered := make([]uint64, 0, len(current))
		for state := range current {
			ordered = append(ordered, state)
		}
		sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
		for _, state := range ordered {
			if err := m.work(1); err != nil {
				return AbsTrace{}, false, err
			}
			next := m.system.Next[state*inputs+input]
			if m.project(next) != trace.AbstractStates[frame+1] {
				continue
			}
			if _, seen := nextSet[next]; !seen {
				path := slices.Clone(current[state])
				path = append(path, next)
				nextSet[next] = path
			}
		}
		current = nextSet
		if len(current) == 0 {
			return AbsTrace{}, false, nil
		}
	}
	ordered := make([]uint64, 0, len(current))
	for state := range current {
		ordered = append(ordered, state)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	for _, state := range ordered {
		if m.system.Bad[state] {
			return AbsTrace{States: slices.Clone(current[state]), AbstractStates: slices.Clone(trace.AbstractStates), Inputs: slices.Clone(trace.Inputs)}, true, nil
		}
	}
	return AbsTrace{}, false, nil
}

func (m *AbsManager) Analyze() (AbsResult, error) {
	if err := m.check(); err != nil {
		return AbsResult{}, err
	}
	m.phase = AbsSearching
	defer func() {
		if !m.closed {
			m.phase = AbsFinished
		}
	}()
	for refinement := 0; refinement <= m.options.MaxRefinements; refinement++ {
		status, reason, abstractTrace, err := m.abstractSearch()
		if err != nil {
			if err == ErrCancelled {
				reason = AbsCancelled
			}
			return AbsResult{Status: AbsUnknown, Reason: reason, VisibleBits: m.VisibleBits(), Stats: m.stats}, err
		}
		if status == AbsProved || status == AbsUnknown {
			return AbsResult{Status: status, Reason: reason, VisibleBits: m.VisibleBits(), Stats: m.stats}, nil
		}
		concrete, real, err := m.Replay(*abstractTrace)
		if err != nil {
			return AbsResult{Status: AbsUnknown, Reason: AbsResourceLimit, VisibleBits: m.VisibleBits(), Stats: m.stats}, err
		}
		if real {
			return AbsResult{Status: AbsCounterexample, Reason: AbsRealCounterexample, VisibleBits: m.VisibleBits(), Trace: &concrete, Stats: m.stats}, nil
		}
		m.stats.SpuriousTraces++
		bit := -1
		for i, visible := range m.visible {
			if !visible {
				bit = i
				break
			}
		}
		if bit < 0 || refinement == m.options.MaxRefinements {
			return AbsResult{Status: AbsUnknown, Reason: AbsRefinementLimit, VisibleBits: m.VisibleBits(), Stats: m.stats}, nil
		}
		if err := m.Refine([]int{bit}); err != nil {
			reason := AbsResourceLimit
			if err == ErrCancelled {
				reason = AbsCancelled
			}
			return AbsResult{Status: AbsUnknown, Reason: reason, VisibleBits: m.VisibleBits(), Stats: m.stats}, err
		}
	}
	return AbsResult{Status: AbsUnknown, Reason: AbsRefinementLimit, VisibleBits: m.VisibleBits(), Stats: m.stats}, nil
}

// Verify independently checks a returned proof or counterexample against the
// owned concrete transition system. Proof checking exhaustively computes the
// reachable fixed point; counterexample checking replays the supplied trace.
func (m *AbsManager) Verify(result AbsResult) (bool, error) {
	if err := m.check(); err != nil {
		return false, err
	}
	if result.Status == AbsCounterexample {
		if result.Trace == nil || len(result.Trace.States) == 0 {
			return false, ErrDimension
		}
		return m.verifyConcreteTrace(result.Trace.Clone())
	}
	if result.Status != AbsProved || result.Reason != AbsFixedPoint {
		return false, nil
	}
	states := uint64(1) << m.system.StateBits
	inputs := uint64(1) << m.system.InputBits
	temporary, ok := absMulAdd(states, 9, 0)
	if !ok {
		return false, ErrLimit
	}
	if err := m.bytes(temporary); err != nil {
		return false, err
	}
	seen := make([]bool, int(states))
	queue := make([]uint64, 0, int(states))
	queue = append(queue, m.system.Initial...)
	for _, state := range queue {
		seen[state] = true
	}
	for head := 0; head < len(queue); head++ {
		state := queue[head]
		if err := m.work(1); err != nil {
			return false, err
		}
		if m.system.Bad[state] {
			return false, nil
		}
		for input := uint64(0); input < inputs; input++ {
			if err := m.work(1); err != nil {
				return false, err
			}
			next := m.system.Next[state*inputs+input]
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return true, nil
}

func (m *AbsManager) verifyConcreteTrace(trace AbsTrace) (bool, error) {
	if len(trace.States) == 0 || len(trace.States) != len(trace.Inputs)+1 || len(trace.States) > m.options.MaxTraceStates {
		return false, ErrDimension
	}
	if len(trace.AbstractStates) != 0 && len(trace.AbstractStates) != len(trace.States) {
		return false, ErrDimension
	}
	initial := false
	for _, state := range m.system.Initial {
		if state == trace.States[0] {
			initial = true
			break
		}
	}
	if !initial {
		return false, nil
	}
	states := uint64(1) << m.system.StateBits
	inputs := uint64(1) << m.system.InputBits
	for frame, state := range trace.States {
		if err := m.work(1); err != nil {
			return false, err
		}
		if state >= states || (len(trace.AbstractStates) != 0 && m.project(state) != trace.AbstractStates[frame]) {
			return false, nil
		}
		if frame == len(trace.Inputs) {
			continue
		}
		input := trace.Inputs[frame]
		if input >= inputs || m.system.Next[state*inputs+input] != trace.States[frame+1] {
			return false, nil
		}
	}
	return m.system.Bad[trace.States[len(trace.States)-1]], nil
}

func absVisibleList(visible []bool) []int {
	result := make([]int, 0, len(visible))
	for bit, set := range visible {
		if set {
			result = append(result, bit)
		}
	}
	return result
}

func absMulAdd(a, b, c uint64) (uint64, bool) {
	if b != 0 && a > ^uint64(0)/b {
		return 0, false
	}
	value := a * b
	if value > ^uint64(0)-c {
		return 0, false
	}
	return value + c, true
}

func absAdd(a, b uint64) (uint64, bool) {
	if a > ^uint64(0)-b {
		return 0, false
	}
	return a + b, true
}

// AbsFramesToCounts converts a VTA-style list of per-frame object IDs into the
// GLA-style count for each object. It is the owned form of the conversion in
// absUtil.c.
func AbsFramesToCounts(frames [][]int, objects int, options AbsOptions) ([]int, error) {
	o, err := options.normalized()
	if err != nil {
		return nil, err
	}
	if objects < 0 || objects > o.MaxConcreteStates || len(frames) > o.MaxFrames || len(frames) > o.MaxDepth {
		return nil, ErrDimension
	}
	if uint64(objects)*8 > o.MaxBytes {
		return nil, ErrLimit
	}
	counts := make([]int, objects)
	work := uint64(0)
	for _, frame := range frames {
		work++
		if work > o.MaxWork {
			return nil, ErrLimit
		}
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		for _, object := range frame {
			work++
			if work > o.MaxWork {
				return nil, ErrLimit
			}
			if o.Stop != nil && o.Stop() {
				return nil, ErrCancelled
			}
			if object < 0 || object >= objects || counts[object] == int(^uint(0)>>1) {
				return nil, ErrDimension
			}
			counts[object]++
		}
	}
	return counts, nil
}

func AbsCountsToFrames(counts []int, options AbsOptions) ([][]int, error) {
	o, err := options.normalized()
	if err != nil {
		return nil, err
	}
	if len(counts) > o.MaxConcreteStates {
		return nil, ErrLimit
	}
	maximum, entries, work := 0, uint64(0), uint64(0)
	for _, count := range counts {
		work++
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		if work > o.MaxWork {
			return nil, ErrLimit
		}
		if count < 0 {
			return nil, ErrDimension
		}
		if count > maximum {
			maximum = count
		}
		if uint64(count) > ^uint64(0)-entries {
			return nil, ErrLimit
		}
		entries += uint64(count)
		if entries > o.MaxWork-work {
			return nil, ErrLimit
		}
	}
	if maximum > o.MaxFrames || maximum > o.MaxDepth {
		return nil, ErrLimit
	}
	remaining, ok := absAdd(entries, uint64(maximum))
	if !ok || remaining > o.MaxWork-work {
		return nil, ErrLimit
	}
	entryBytes, ok := absMulAdd(entries, 8, 0)
	if !ok {
		return nil, ErrLimit
	}
	// The temporary size-difference array lets every frame receive its exact
	// capacity, avoiding append growth beyond the preflight.
	bytes, ok := absMulAdd(uint64(maximum), 32, entryBytes)
	if !ok || bytes > o.MaxBytes {
		return nil, ErrLimit
	}
	frames := make([][]int, maximum)
	sizes := make([]int, maximum)
	for _, count := range counts {
		if count != 0 {
			sizes[0]++
			if count < maximum {
				sizes[count]--
			}
		}
	}
	active := 0
	for frame := range frames {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		active += sizes[frame]
		frames[frame] = make([]int, 0, active)
	}
	for object, count := range counts {
		for frame := 0; frame < count; frame++ {
			if o.Stop != nil && o.Stop() {
				return nil, ErrCancelled
			}
			frames[frame] = append(frames[frame], object)
		}
	}
	return frames, nil
}

var AbsBuildInventory = []string{
	"absDup.c", "absGla.c", "absGlaOld.c", "absIter.c", "absOldCex.c",
	"absOldRef.c", "absOldSat.c", "absOldSim.c", "absOut.c", "absPth.c",
	"absRef.c", "absRefSelect.c", "absRpm.c", "absRpmOld.c", "absVta.c", "absUtil.c",
}

// AbsSourceInventory reports every source, public header, and build manifest
// boundary in src/proof/abs, including abs.c's empty translation unit.
func AbsSourceInventory() []string {
	return []string{
		"abs.c", "abs.h", "absDup.c", "absGla.c", "absGlaOld.c", "absIter.c",
		"absOldCex.c", "absOldRef.c", "absOldSat.c", "absOldSim.c", "absOut.c",
		"absPth.c", "absRef.c", "absRef.h", "absRefJ.c", "absRefJ.h",
		"absRefSelect.c", "absRpm.c", "absRpmOld.c", "absUtil.c", "absVta.c",
		"module.make",
	}
}

func AbsInventoryReport() string {
	return fmt.Sprintf("proof/abs: %d boundaries, %d default-build units; owned explicit abstraction/refinement core", len(AbsSourceInventory()), len(AbsBuildInventory))
}
