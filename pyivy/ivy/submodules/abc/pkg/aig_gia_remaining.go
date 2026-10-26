package abc

// Standalone trace, counterexample, simulation-statistics, and cone-partition
// semantics from src/aig/gia/giaCex.c, giaSim.c, giaSwitch.c, and giaCone.c.

import (
	"fmt"
	"math"
	"strconv"
)

type GiaTraceOptions struct {
	MaxFrames, MaxObjects, MaxWork, MaxBytes int
	Stop                                     func() bool
}

type GiaTrace struct {
	Outputs [][]bool
	States  [][]bool
}

type GiaCounterexample struct {
	InitialState []bool
	Inputs       [][]bool
	Output       int
	Frame        int
	DualOutput   bool
}

type GiaFailure struct {
	Output int
	Frame  int
}

type GiaSimulationEntry struct {
	Ones, Samples, Transitions, TransitionSamples int
	ProbabilityOne, SwitchingProbability          float64
	TransitionProbability                         float64
}

type GiaSimulationReport struct {
	Objects []GiaSimulationEntry
}

type GiaOutputConePartition struct {
	Groups [][]int
}

type giaRemainingBudget struct {
	options         GiaTraceOptions
	work            int
	validationBytes int
}

func (b *giaRemainingBudget) take(n int) error {
	if n < 0 || b.work > math.MaxInt-n {
		return ErrLimit
	}
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	b.work += n
	if b.options.MaxWork != 0 && b.work > b.options.MaxWork {
		return ErrLimit
	}
	return nil
}

func giaRemainingLimits(m *GiaManager, frames int, options GiaTraceOptions) (*giaRemainingBudget, error) {
	if m == nil || frames < 0 || options.MaxFrames < 0 || options.MaxObjects < 0 || options.MaxWork < 0 || options.MaxBytes < 0 {
		return nil, ErrDimension
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 1 << 20
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 26
	}
	validationBytes := 0
	if m.hash != nil {
		var err error
		validationBytes, err = giaRemainingMul(len(m.objects), 64)
		if err != nil {
			return nil, err
		}
		if options.MaxBytes != 0 && validationBytes > options.MaxBytes {
			return nil, ErrLimit
		}
	}
	budget := &giaRemainingBudget{options: options, validationBytes: validationBytes}
	if err := budget.take(0); err != nil {
		return nil, err
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if frames > options.MaxFrames || len(m.objects) > options.MaxObjects {
		return nil, ErrLimit
	}
	return budget, nil
}

func giaRemainingAdd(a, b int) (int, error) {
	if a < 0 || b < 0 || a > math.MaxInt-b {
		return 0, ErrLimit
	}
	return a + b, nil
}

func giaRemainingMul(a, b int) (int, error) {
	if a < 0 || b < 0 || a != 0 && b > math.MaxInt/a {
		return 0, ErrLimit
	}
	return a * b, nil
}

func giaRemainingBytes(options GiaTraceOptions, terms ...int) error {
	total := 0
	for _, term := range terms {
		var err error
		total, err = giaRemainingAdd(total, term)
		if err != nil {
			return err
		}
	}
	if options.MaxBytes != 0 && total > options.MaxBytes {
		return ErrLimit
	}
	return nil
}

func giaRemainingEvaluate(m *GiaManager, inputs []bool, budget *giaRemainingBudget) ([]bool, error) {
	if len(inputs) != len(m.cis) {
		return nil, ErrDimension
	}
	values := make([]bool, len(m.objects))
	for i, id := range m.cis {
		if err := budget.take(1); err != nil {
			return nil, err
		}
		values[id] = inputs[i]
	}
	read := func(l GiaLit) bool { return values[l.ID()] != l.Complemented() }
	for i, obj := range m.objects {
		if err := budget.take(1); err != nil {
			return nil, err
		}
		if !obj.IsAnd() && !obj.IsCO() {
			continue
		}
		a, err := obj.Fanin0(GiaID(i))
		if err != nil {
			return nil, err
		}
		if obj.IsCO() || obj.IsBuffer() {
			values[i] = read(a)
			continue
		}
		b, err := obj.Fanin1(GiaID(i))
		if err != nil {
			return nil, err
		}
		if control, yes := m.MuxControl(GiaID(i)); yes {
			if read(control) {
				values[i] = read(b)
			} else {
				values[i] = read(a)
			}
		} else if obj.IsXor() {
			values[i] = read(a) != read(b)
		} else {
			values[i] = read(a) && read(b)
		}
	}
	return values, nil
}

// SimulateTrace performs exact sequential replay on a frozen manager. States
// contains the initial state followed by one state per input frame. Outputs
// contains primary-output values for each frame. Results never alias caller
// input. Concurrent calls are safe only while no caller mutates the manager.
func (m *GiaManager) SimulateTrace(initial []bool, inputs [][]bool, options GiaTraceOptions) (GiaTrace, error) {
	budget, err := giaRemainingLimits(m, len(inputs), options)
	if err != nil {
		return GiaTrace{}, err
	}
	if len(initial) != m.registers {
		return GiaTrace{}, ErrDimension
	}
	for i := range inputs {
		if err := budget.take(1); err != nil {
			return GiaTrace{}, err
		}
		if len(inputs[i]) != m.PrimaryInputCount() {
			return GiaTrace{}, ErrDimension
		}
	}
	frames, objects, cis := len(inputs), len(m.objects), len(m.cis)
	stateRows, err := giaRemainingAdd(frames, 1)
	if err != nil {
		return GiaTrace{}, err
	}
	outerElements, err := giaRemainingAdd(frames, stateRows)
	if err != nil {
		return GiaTrace{}, err
	}
	outer, err := giaRemainingMul(outerElements, 3*strconv.IntSize/8)
	if err != nil {
		return GiaTrace{}, err
	}
	perFrame, err := giaRemainingAdd(cis, objects)
	if err != nil {
		return GiaTrace{}, err
	}
	perFrame, err = giaRemainingAdd(perFrame, m.PrimaryOutputCount())
	if err != nil {
		return GiaTrace{}, err
	}
	perFrame, err = giaRemainingAdd(perFrame, m.registers)
	if err != nil {
		return GiaTrace{}, err
	}
	inner, err := giaRemainingMul(frames, perFrame)
	if err != nil {
		return GiaTrace{}, err
	}
	initialBytes, err := giaRemainingMul(2, m.registers)
	if err != nil {
		return GiaTrace{}, err
	}
	if err := giaRemainingBytes(options, budget.validationBytes, outer, inner, initialBytes); err != nil {
		return GiaTrace{}, err
	}
	state := append([]bool(nil), initial...)
	result := GiaTrace{
		Outputs: make([][]bool, 0, len(inputs)),
		States:  make([][]bool, 1, len(inputs)+1),
	}
	result.States[0] = append([]bool(nil), state...)
	for i := range inputs {
		ciValues := make([]bool, len(m.cis))
		copy(ciValues, inputs[i])
		copy(ciValues[m.PrimaryInputCount():], state)
		values, err := giaRemainingEvaluate(m, ciValues, budget)
		if err != nil {
			return GiaTrace{}, err
		}
		output := make([]bool, m.PrimaryOutputCount())
		for j := range output {
			if err := budget.take(1); err != nil {
				return GiaTrace{}, err
			}
			output[j] = values[m.cos[j]]
		}
		next := make([]bool, m.registers)
		for j := range next {
			if err := budget.take(1); err != nil {
				return GiaTrace{}, err
			}
			next[j] = values[m.cos[m.PrimaryOutputCount()+j]]
		}
		result.Outputs = append(result.Outputs, output)
		state = next
		result.States = append(result.States, state)
	}
	return result, nil
}

func validateGiaCounterexample(m *GiaManager, c GiaCounterexample) error {
	if c.Frame < 0 || c.Frame >= len(c.Inputs) || len(c.Inputs) != c.Frame+1 {
		return ErrDimension
	}
	outputs := m.PrimaryOutputCount()
	if c.Output < 0 {
		return ErrDimension
	}
	if c.DualOutput {
		if c.Output > (math.MaxInt-1)/2 || 2*c.Output+1 >= outputs {
			return ErrDimension
		}
	} else if c.Output >= outputs {
		return ErrDimension
	}
	return nil
}

// VerifyCounterexample exactly replays a supplied trace and checks its selected
// output in the final frame. Dual mode checks that output pair 2*i,2*i+1 differs.
func (m *GiaManager) VerifyCounterexample(c GiaCounterexample, options GiaTraceOptions) (bool, error) {
	if m == nil {
		return false, ErrDimension
	}
	if err := validateGiaCounterexample(m, c); err != nil {
		return false, err
	}
	trace, err := m.SimulateTrace(c.InitialState, c.Inputs, options)
	if err != nil {
		return false, err
	}
	values := trace.Outputs[c.Frame]
	if c.DualOutput {
		return values[2*c.Output] != values[2*c.Output+1], nil
	}
	return values[c.Output], nil
}

// FindFirstFailure returns the earliest asserted primary output, breaking ties
// by output order. This makes the standalone behavior deterministic; the two
// native giaCex.c discovery helpers use different final/latest-frame scans.
func (m *GiaManager) FindFirstFailure(initial []bool, inputs [][]bool, firstOutput int, options GiaTraceOptions) (GiaFailure, bool, error) {
	if m == nil || firstOutput < 0 || firstOutput > m.PrimaryOutputCount() {
		return GiaFailure{}, false, ErrDimension
	}
	trace, err := m.SimulateTrace(initial, inputs, options)
	if err != nil {
		return GiaFailure{}, false, err
	}
	for frame := range trace.Outputs {
		for output := firstOutput; output < len(trace.Outputs[frame]); output++ {
			if trace.Outputs[frame][output] {
				return GiaFailure{Output: output, Frame: frame}, true, nil
			}
		}
	}
	return GiaFailure{}, false, nil
}

// SimulatePatterns computes exact per-object one and adjacent-pattern switching
// counts on a frozen manager. Each pattern supplies all CIs in terminal order.
// This is the deterministic explicit-pattern counterpart of giaSwitch.c; its
// random packed simulation, warm-up frames and transition generator are absent.
func (m *GiaManager) SimulatePatterns(patterns [][]bool, options GiaTraceOptions) (GiaSimulationReport, error) {
	budget, err := giaRemainingLimits(m, len(patterns), options)
	if err != nil {
		return GiaSimulationReport{}, err
	}
	for i := range patterns {
		if err := budget.take(1); err != nil {
			return GiaSimulationReport{}, err
		}
		if len(patterns[i]) != len(m.cis) {
			return GiaSimulationReport{}, ErrDimension
		}
	}
	entryBytes, err := giaRemainingMul(len(m.objects), 4*strconv.IntSize/8+3*8)
	if err != nil {
		return GiaSimulationReport{}, err
	}
	valueBytes, err := giaRemainingMul(len(patterns), len(m.objects))
	if err != nil {
		return GiaSimulationReport{}, err
	}
	if err := giaRemainingBytes(options, budget.validationBytes, entryBytes, valueBytes); err != nil {
		return GiaSimulationReport{}, err
	}
	entries := make([]GiaSimulationEntry, len(m.objects))
	var previous []bool
	for i := range patterns {
		values, err := giaRemainingEvaluate(m, patterns[i], budget)
		if err != nil {
			return GiaSimulationReport{}, err
		}
		for id, value := range values {
			if err := budget.take(1); err != nil {
				return GiaSimulationReport{}, err
			}
			entries[id].Samples++
			if value {
				entries[id].Ones++
			}
			if previous != nil {
				entries[id].TransitionSamples++
				if value != previous[id] {
					entries[id].Transitions++
				}
			}
		}
		previous = values
	}
	for i := range entries {
		if err := budget.take(1); err != nil {
			return GiaSimulationReport{}, err
		}
		if entries[i].Samples != 0 {
			entries[i].ProbabilityOne = float64(entries[i].Ones) / float64(entries[i].Samples)
			entries[i].SwitchingProbability = 2 * entries[i].ProbabilityOne * (1 - entries[i].ProbabilityOne)
		}
		if entries[i].TransitionSamples != 0 {
			entries[i].TransitionProbability = float64(entries[i].Transitions) / float64(entries[i].TransitionSamples)
		}
	}
	return GiaSimulationReport{Objects: entries}, nil
}

type giaRemainingUnion struct{ parent []int }

func newGiaRemainingUnion(n int, budget *giaRemainingBudget) (giaRemainingUnion, error) {
	p := make([]int, n)
	for i := range p {
		if err := budget.take(1); err != nil {
			return giaRemainingUnion{}, err
		}
		p[i] = i
	}
	return giaRemainingUnion{parent: p}, nil
}
func (u *giaRemainingUnion) find(x int) int {
	for x != u.parent[x] {
		u.parent[x] = u.parent[u.parent[x]]
		x = u.parent[x]
	}
	return x
}
func (u *giaRemainingUnion) merge(a, b int) {
	a, b = u.find(a), u.find(b)
	if a == b {
		return
	}
	if a > b {
		a, b = b, a
	}
	u.parent[b] = a
}

// PartitionOutputCones groups primary outputs connected through a shared
// nonconstant transitive fanin on a frozen manager. Register outputs close
// through their matching register inputs. This computes the deterministic
// connected-component result of giaCone.c's mutable frontier merger.
func (m *GiaManager) PartitionOutputCones(options GiaTraceOptions) (GiaOutputConePartition, error) {
	if m == nil {
		return GiaOutputConePartition{}, ErrDimension
	}
	outputs := m.PrimaryOutputCount()
	budget, err := giaRemainingLimits(m, 0, options)
	if err != nil {
		return GiaOutputConePartition{}, err
	}
	if outputs == 0 {
		return GiaOutputConePartition{Groups: [][]int{}}, nil
	}
	intBytes := strconv.IntSize / 8
	objectBytes, err := giaRemainingMul(len(m.objects), 2*intBytes+3*4)
	if err != nil {
		return GiaOutputConePartition{}, ErrLimit
	}
	outputBytes, err := giaRemainingMul(outputs, 4*intBytes+3*intBytes)
	if err != nil {
		return GiaOutputConePartition{}, ErrLimit
	}
	if err := giaRemainingBytes(options, budget.validationBytes, objectBytes, outputBytes); err != nil {
		return GiaOutputConePartition{}, err
	}
	owner := make([]int, len(m.objects))
	for i := range owner {
		if err := budget.take(1); err != nil {
			return GiaOutputConePartition{}, err
		}
		owner[i] = -1
	}
	union, err := newGiaRemainingUnion(outputs, budget)
	if err != nil {
		return GiaOutputConePartition{}, err
	}
	seen := make([]int, len(m.objects))
	stackCapacity, err := giaRemainingMul(len(m.objects), 3)
	if err != nil {
		return GiaOutputConePartition{}, err
	}
	stack := make([]GiaID, 0, stackCapacity)
	for output := 0; output < outputs; output++ {
		stack = append(stack[:0], m.cos[output])
		for len(stack) > 0 {
			if err := budget.take(1); err != nil {
				return GiaOutputConePartition{}, err
			}
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if id == 0 || seen[id] == output+1 {
				continue
			}
			seen[id] = output + 1
			if owner[id] < 0 {
				owner[id] = output
			} else {
				union.merge(output, owner[id])
			}
			obj := m.objects[id]
			if obj.IsCI() {
				index := int(obj.Diff1())
				if index >= m.PrimaryInputCount() {
					register := index - m.PrimaryInputCount()
					ri := m.cos[m.PrimaryOutputCount()+register]
					stack = append(stack, ri)
				}
				continue
			}
			if !obj.IsAnd() && !obj.IsCO() {
				return GiaOutputConePartition{}, fmt.Errorf("%w: cone object", ErrInvariant)
			}
			first, err := obj.Fanin0(id)
			if err != nil {
				return GiaOutputConePartition{}, err
			}
			stack = append(stack, first.ID())
			if obj.IsAnd() && !obj.IsBuffer() {
				second, err := obj.Fanin1(id)
				if err != nil {
					return GiaOutputConePartition{}, err
				}
				stack = append(stack, second.ID())
				if control, yes := m.MuxControl(id); yes {
					stack = append(stack, control.ID())
				}
			}
		}
	}
	counts := make([]int, outputs)
	for i := 0; i < outputs; i++ {
		if err := budget.take(1); err != nil {
			return GiaOutputConePartition{}, err
		}
		root := union.find(i)
		counts[root]++
	}
	groups := make([][]int, 0, outputs)
	rootToGroup := make([]int, outputs)
	for i := range rootToGroup {
		rootToGroup[i] = -1
		if counts[i] != 0 {
			rootToGroup[i] = len(groups)
			groups = append(groups, make([]int, 0, counts[i]))
		}
	}
	for i := 0; i < outputs; i++ {
		if err := budget.take(1); err != nil {
			return GiaOutputConePartition{}, err
		}
		root := union.find(i)
		groups[rootToGroup[root]] = append(groups[rootToGroup[root]], i)
	}
	return GiaOutputConePartition{Groups: groups}, nil
}
