package abc

// Counterexample depth targets from src/sat/bmc/bmcCexDepth.c.

import "slices"

type CexDepthOptions struct {
	CounterexampleToolOptions
	MaxCones     int
	PruneImplied bool
}

type CexDepthResult struct {
	Analysis CounterexampleAnalysis
	Starts   []int
	Cones    []*GiaManager
	Graph    *GiaManager
}

func normalizeCexDepthOptions(options CexDepthOptions) (CexDepthOptions, error) {
	normalized, err := normalizeCounterexampleToolOptions(options.CounterexampleToolOptions)
	if err != nil {
		return options, err
	}
	options.CounterexampleToolOptions = normalized
	if options.MaxCones < 0 {
		return options, ErrDimension
	}
	if options.MaxCones == 0 {
		options.MaxCones = 1 << 20
	}
	return options, nil
}

// EnlargeCounterexampleTarget creates the n-frame PO-zero target over reversed
// frame-major PIs followed by the arbitrary initial register state. PI block
// frames-1 is applied first and block zero last, preserving the source order.
func (m *GiaManager) EnlargeCounterexampleTarget(frames int, options CounterexampleToolOptions) (*GiaManager, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return nil, err
	}
	if m == nil || frames < 1 || frames > options.MaxFrames {
		if frames > options.MaxFrames {
			return nil, ErrLimit
		}
		return nil, ErrDimension
	}
	if err = m.Check(); err != nil {
		return nil, err
	}
	if m.PrimaryOutputCount() == 0 {
		return nil, ErrDimension
	}
	if m.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	inputCount, err := checkedSize(m.RegisterCount(), frames, m.PrimaryInputCount(), options.MaxMappingEntries)
	if err != nil {
		return nil, err
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name, target.Spec = m.Name, m.Spec
	inputs := make([]GiaLit, inputCount)
	for index := range inputs {
		if cexToolStopped(options) {
			return nil, ErrCancelled
		}
		inputs[index], err = target.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	state := slices.Clone(inputs[frames*m.PrimaryInputCount():])
	mapping := make([]GiaLit, m.ObjectCount())
	for step := 0; step < frames; step++ {
		for index := range mapping {
			mapping[index] = InvalidGiaLit
		}
		mapping[0] = GiaConst0
		block := (frames - 1 - step) * m.PrimaryInputCount()
		for input, id := range m.cis[:m.PrimaryInputCount()] {
			mapping[id] = inputs[block+input]
		}
		for register, id := range m.cis[m.PrimaryInputCount():] {
			mapping[id] = state[register]
		}
		for id, object := range m.objects {
			if !object.IsAnd() {
				continue
			}
			if cexToolStopped(options) {
				return nil, ErrCancelled
			}
			mapping[id], err = hashGiaCopiedGate(m, target, GiaID(id), mapping)
			if err != nil {
				return nil, err
			}
		}
		for register, ri := range m.cos[m.PrimaryOutputCount():] {
			driver, faninErr := m.objects[ri].Fanin0(ri)
			if faninErr != nil {
				return nil, faninErr
			}
			state[register], err = mappedGiaSignal(mapping, driver)
			if err != nil {
				return nil, err
			}
		}
	}
	po := m.cos[0]
	driver, err := m.objects[po].Fanin0(po)
	if err != nil {
		return nil, err
	}
	driver, err = mappedGiaSignal(mapping, driver)
	if err != nil {
		return nil, err
	}
	if _, err = target.AppendCO(driver); err != nil {
		return nil, err
	}
	target.HashStop()
	clean, _, err := target.Cleanup()
	return clean, err
}

// CounterexampleReachabilityTarget existentially removes every enlarged PI
// block, retains only state inputs, and appends the resulting predicate as an
// ordinary property of the source transition system.
func (m *GiaManager) CounterexampleReachabilityTarget(frames int, options CounterexampleToolOptions) (*GiaManager, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return nil, err
	}
	target, err := m.EnlargeCounterexampleTarget(frames, options)
	if err != nil {
		return nil, err
	}
	limit, err := checkedSize(0, frames, m.PrimaryInputCount(), options.MaxMappingEntries)
	if err != nil {
		return nil, err
	}
	for input := 0; input < limit; input++ {
		if cexToolStopped(options) {
			return nil, ErrCancelled
		}
		target, err = target.ExistentialInput(input, GiaCofactorOptions{
			MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries,
			Stop: options.Stop,
		})
		if err != nil {
			return nil, err
		}
	}
	target, err = target.RetainLastPrimaryInputs(m.RegisterCount(), GiaCofactorOptions{
		MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries,
		Stop: options.Stop,
	})
	if err != nil {
		return nil, err
	}
	return m.AppendPropertyCones([]*GiaManager{target}, true, GiaCofactorOptions{
		MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries,
		Stop: options.Stop,
	})
}

func checkDepthPlainGraph(m *GiaManager) error {
	if m == nil {
		return ErrDimension
	}
	if err := m.Check(); err != nil {
		return err
	}
	for id, object := range m.objects {
		if !object.IsAnd() {
			continue
		}
		if object.IsBuffer() || object.IsXor() {
			return ErrUnsupported
		}
		if _, mux := m.MuxControl(GiaID(id)); mux {
			return ErrUnsupported
		}
	}
	return nil
}

// BuildCounterexampleDepthCone builds the unate register-state justification
// from start through the failing frame. states is the expanded PI/RO trace
// returned by CounterexampleInnerStates. pruneImplied consolidates the first C
// implementation; false consolidates Bmc_CexBuildNetwork2_. Both have the same
// Boolean contract, while pruning may reduce intermediate construction.
func (m *GiaManager) BuildCounterexampleDepthCone(states *Counterexample, start int, pruneImplied bool, options CounterexampleToolOptions) (*GiaManager, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return nil, err
	}
	if err = checkDepthPlainGraph(m); err != nil {
		return nil, err
	}
	if states == nil || states.Registers() != 0 || states.Inputs() != m.InputCount() || states.Property < 0 || states.Property >= m.PrimaryOutputCount() || start < 0 || start >= states.Frames() {
		return nil, ErrDimension
	}
	if err = states.Validate(); err != nil {
		return nil, err
	}
	if states.Frames() > options.MaxFrames || states.BitCount() > options.MaxBits || m.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name = "unate"
	stateFormula := make([]GiaLit, m.RegisterCount())
	stateValues := make([]bool, m.RegisterCount())
	stateFixed := make([]bool, m.RegisterCount())
	for register := range stateFormula {
		stateValues[register], _ = states.Input(start, m.PrimaryInputCount()+register)
		stateFormula[register], err = target.AppendCI()
		if err != nil {
			return nil, err
		}
		stateFormula[register] = stateFormula[register].NotIf(!stateValues[register])
	}
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	ciValues := make([]bool, m.InputCount())
	ciFixed := make([]bool, m.InputCount())
	formula := make([]GiaLit, m.ObjectCount())
	for frame := start; frame < states.Frames(); frame++ {
		if cexToolStopped(options) {
			return nil, ErrCancelled
		}
		cexReadFrame(states, frame, ciValues)
		for input, id := range m.cis[:m.PrimaryInputCount()] {
			ciFixed[input] = true
			formula[id] = GiaConst1
		}
		for register, id := range m.cis[m.PrimaryInputCount():] {
			ciValues[m.PrimaryInputCount()+register] = stateValues[register]
			ciFixed[m.PrimaryInputCount()+register] = stateFixed[register]
			formula[id] = stateFormula[register]
		}
		values, fixed, evalErr := cexFrameValuesAndImplications(m, ciValues, ciFixed, options)
		if evalErr != nil {
			return nil, evalErr
		}
		formula[0] = GiaConst1
		for id, object := range m.objects {
			if !object.IsAnd() {
				continue
			}
			if pruneImplied && fixed[id] {
				formula[id] = GiaConst1
				continue
			}
			first, _ := object.Fanin0(GiaID(id))
			second, _ := object.Fanin1(GiaID(id))
			left, right := formula[first.ID()], formula[second.ID()]
			leftValue := values[first.ID()] != first.Complemented()
			rightValue := values[second.ID()] != second.Complemented()
			switch {
			case leftValue && rightValue:
				formula[id], err = target.HashAnd(left, right)
			case !leftValue && !rightValue:
				formula[id], err = target.HashOr(left, right)
			case !leftValue:
				formula[id] = left
			default:
				formula[id] = right
			}
			if err != nil {
				return nil, err
			}
		}
		for _, co := range m.cos {
			driver, faninErr := m.objects[co].Fanin0(co)
			if faninErr != nil {
				return nil, faninErr
			}
			formula[co] = formula[driver.ID()]
		}
		for register, ri := range m.cos[m.PrimaryOutputCount():] {
			stateValues[register] = values[ri]
			stateFixed[register] = fixed[ri]
			stateFormula[register] = formula[ri]
		}
	}
	po := m.cos[states.Property]
	frameValues := make([]bool, m.InputCount())
	cexReadFrame(states, states.LastFrame(), frameValues)
	outputs, evalErr := m.Evaluate(frameValues)
	if evalErr != nil || !outputs[states.Property] || formula[po] == GiaConst0 {
		if evalErr != nil {
			return nil, evalErr
		}
		return nil, ErrInvariant
	}
	if _, err = target.AppendCO(formula[po]); err != nil {
		return nil, err
	}
	target.HashStop()
	clean, _, err := target.Cleanup()
	return clean, err
}

// BuildCounterexampleDepthTargets builds the last count nonzero-start cones
// and appends them as register-only properties. This preserves the source's
// exclusion of frame zero from the diagnostic cone list.
func (m *GiaManager) BuildCounterexampleDepthTargets(states *Counterexample, count int, options CexDepthOptions) (CexDepthResult, error) {
	var result CexDepthResult
	options, err := normalizeCexDepthOptions(options)
	if err != nil {
		return result, err
	}
	if states == nil || count < 0 {
		return result, ErrDimension
	}
	if err = checkDepthPlainGraph(m); err != nil {
		return result, err
	}
	if err = states.Validate(); err != nil {
		return result, err
	}
	if states.Registers() != 0 || states.Inputs() != m.InputCount() || states.Property < 0 || states.Property >= m.PrimaryOutputCount() {
		return result, ErrDimension
	}
	if states.Frames() > options.MaxFrames || states.BitCount() > options.MaxBits || m.ObjectCount() > options.MaxMappingEntries {
		return result, ErrLimit
	}
	count = min(count, states.LastFrame())
	if count > options.MaxCones {
		return result, ErrLimit
	}
	for start := states.LastFrame(); start > states.LastFrame()-count; start-- {
		cone, buildErr := m.BuildCounterexampleDepthCone(states, start, options.PruneImplied, options.CounterexampleToolOptions)
		if buildErr != nil {
			return CexDepthResult{}, buildErr
		}
		result.Starts = append(result.Starts, start)
		result.Cones = append(result.Cones, cone)
	}
	result.Graph, err = m.AppendPropertyCones(result.Cones, true, GiaCofactorOptions{
		MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries,
		Stop: options.Stop,
	})
	return result, err
}

// AnalyzeCounterexampleDepth runs the shared care analysis, verifies it, then
// returns the selected depth cones and appended transition graph.
func (m *GiaManager) AnalyzeCounterexampleDepth(trace *Counterexample, count int, options CexDepthOptions) (CexDepthResult, error) {
	var result CexDepthResult
	options, err := normalizeCexDepthOptions(options)
	if err != nil {
		return result, err
	}
	result.Analysis, err = m.AnalyzeCounterexample(trace, options.CounterexampleToolOptions)
	if err != nil {
		return result, err
	}
	targets, err := m.BuildCounterexampleDepthTargets(result.Analysis.States, count, options)
	if err != nil {
		return CexDepthResult{}, err
	}
	targets.Analysis = result.Analysis
	return targets, nil
}
