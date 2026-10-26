package abc

// Priority counterexample care minimization from src/sat/bmc/bmcCexCare.c.

type GiaCexCarePriority struct {
	Priority int
	Value    bool
}

type CexPriorityCareOptions struct {
	CounterexampleToolOptions
	RealPrimaryInputs int
	Attempts          int
}

type CexPriorityCareResult struct {
	Candidates []*Counterexample
	Best       *Counterexample
	Verified   bool
}

func checkCexCarePlainGraph(m *GiaManager) error {
	return checkCexMin2PlainGraph(m)
}

func giaCexPriorityGate(left, right GiaCexCarePriority) GiaCexCarePriority {
	result := GiaCexCarePriority{Value: left.Value && right.Value}
	switch {
	case left.Value && right.Value:
		result.Priority = min(left.Priority, right.Priority)
	case left.Value:
		result.Priority = right.Priority
	case right.Value:
		result.Priority = left.Priority
	default:
		result.Priority = max(left.Priority, right.Priority)
	}
	return result
}

func (m *GiaManager) cexCarePriorityInputs(trace *Counterexample, realInputs, attempt int) ([]GiaCexCarePriority, error) {
	if attempt < 0 || attempt > 3 || realInputs < 0 || realInputs > m.PrimaryInputCount() {
		return nil, ErrDimension
	}
	entries := make([]GiaCexCarePriority, trace.Frames()*m.PrimaryInputCount())
	rank := 0
	assign := func(first, last, step int) {
		for frame := trace.LastFrame(); frame >= 0; frame-- {
			for input := first; input != last; input += step {
				value, _ := trace.Input(frame, input)
				entries[frame*m.PrimaryInputCount()+input] = GiaCexCarePriority{Priority: rank, Value: value}
				rank++
			}
		}
	}
	if attempt < 2 {
		assign(m.PrimaryInputCount()-1, realInputs-1, -1)
	} else {
		assign(realInputs, m.PrimaryInputCount(), 1)
	}
	if attempt&1 == 0 {
		assign(realInputs-1, -1, -1)
	} else {
		assign(0, realInputs, 1)
	}
	if rank != len(entries) {
		return nil, ErrInvariant
	}
	return entries, nil
}

func (m *GiaManager) cexCarePriorityFrame(frame int, inputPriorities, state []GiaCexCarePriority, options CounterexampleToolOptions) ([]GiaCexCarePriority, error) {
	if len(inputPriorities) != m.PrimaryInputCount() || len(state) != m.RegisterCount() {
		return nil, ErrDimension
	}
	entries := make([]GiaCexCarePriority, m.ObjectCount())
	bound := int(^uint(0) >> 1)
	entries[0] = GiaCexCarePriority{Priority: bound}
	for input, pi := range m.cis[:m.PrimaryInputCount()] {
		entries[pi] = inputPriorities[input]
	}
	for register, ro := range m.cis[m.PrimaryInputCount():] {
		entries[ro] = state[register]
	}
	read := func(literal GiaLit) GiaCexCarePriority {
		entry := entries[literal.ID()]
		entry.Value = entry.Value != literal.Complemented()
		return entry
	}
	for index, object := range m.objects {
		if cexToolStopped(options) {
			return nil, ErrCancelled
		}
		id := GiaID(index)
		if object.IsAnd() {
			first, _ := object.Fanin0(id)
			second, _ := object.Fanin1(id)
			entries[id] = giaCexPriorityGate(read(first), read(second))
		} else if object.IsCO() {
			driver, _ := object.Fanin0(id)
			entries[id] = read(driver)
		}
	}
	_ = frame
	return entries, nil
}

func (m *GiaManager) minimizeCounterexamplePriorityAttempt(trace *Counterexample, priorities []GiaCexCarePriority, options CounterexampleToolOptions) (*Counterexample, error) {
	bound := m.PrimaryInputCount()*trace.Frames() + 1
	state := make([]GiaCexCarePriority, m.RegisterCount())
	for register := range state {
		state[register].Priority = bound
		state[register].Value, _ = trace.Initial(register)
	}
	stateRows := make([]GiaCexCarePriority, trace.Frames()*m.RegisterCount())
	for frame := 0; frame < trace.Frames(); frame++ {
		copy(stateRows[frame*m.RegisterCount():], state)
		entries, err := m.cexCarePriorityFrame(frame, priorities[frame*m.PrimaryInputCount():(frame+1)*m.PrimaryInputCount()], state, options)
		if err != nil {
			return nil, err
		}
		for register, ri := range m.cos[m.PrimaryOutputCount():] {
			state[register] = entries[ri]
		}
		if frame == trace.LastFrame() && !entries[m.cos[trace.Property]].Value {
			return nil, ErrInvariant
		}
	}
	care, err := NewCounterexample(0, m.PrimaryInputCount(), trace.Frames())
	if err != nil {
		return nil, err
	}
	care.Property = trace.Property
	carry := make([]bool, m.ObjectCount())
	isPI := make([]bool, m.ObjectCount())
	for _, pi := range m.cis[:m.PrimaryInputCount()] {
		isPI[pi] = true
	}
	for frame := trace.LastFrame(); frame >= 0; frame-- {
		state = stateRows[frame*m.RegisterCount() : (frame+1)*m.RegisterCount()]
		entries, frameErr := m.cexCarePriorityFrame(frame, priorities[frame*m.PrimaryInputCount():(frame+1)*m.PrimaryInputCount()], state, options)
		if frameErr != nil {
			return nil, frameErr
		}
		marked := make([]bool, m.ObjectCount())
		copy(marked, carry)
		if frame == trace.LastFrame() {
			marked[m.cos[trace.Property]] = true
		}
		for _, co := range m.cos {
			if marked[co] {
				driver, _ := m.objects[co].Fanin0(co)
				marked[driver.ID()] = true
			}
		}
		read := func(literal GiaLit) GiaCexCarePriority {
			entry := entries[literal.ID()]
			entry.Value = entry.Value != literal.Complemented()
			return entry
		}
		for index := len(m.objects) - 1; index > 0; index-- {
			if cexToolStopped(options) {
				return nil, ErrCancelled
			}
			id := GiaID(index)
			object := m.objects[id]
			if !object.IsAnd() || !marked[id] {
				continue
			}
			first, _ := object.Fanin0(id)
			second, _ := object.Fanin1(id)
			left, right := read(first), read(second)
			switch {
			case left.Value && right.Value:
				marked[first.ID()], marked[second.ID()] = true, true
			case left.Value:
				marked[second.ID()] = true
			case right.Value:
				marked[first.ID()] = true
			case marked[first.ID()] || marked[second.ID()]:
			case isPI[first.ID()]:
				marked[first.ID()] = true
			case isPI[second.ID()]:
				marked[second.ID()] = true
			case left.Priority > right.Priority:
				marked[first.ID()] = true
			default:
				marked[second.ID()] = true
			}
		}
		for input, pi := range m.cis[:m.PrimaryInputCount()] {
			if marked[pi] {
				_ = care.SetInput(frame, input, true)
			}
		}
		for index := range carry {
			carry[index] = false
		}
		if frame > 0 {
			for register, ro := range m.cis[m.PrimaryInputCount():] {
				if marked[ro] {
					carry[m.cos[m.PrimaryOutputCount()+register]] = true
				}
			}
		}
	}
	return care, nil
}

// MinimizeCounterexampleCarePriority runs the source's one-to-four priority
// orders and retains the first sparsest verified mask.
func (m *GiaManager) MinimizeCounterexampleCarePriority(trace *Counterexample, options CexPriorityCareOptions) (CexPriorityCareResult, error) {
	var result CexPriorityCareResult
	normalized, err := normalizeCounterexampleToolOptions(options.CounterexampleToolOptions)
	if err != nil {
		return result, err
	}
	if err = checkCexCarePlainGraph(m); err != nil {
		return result, err
	}
	if err = checkCounterexampleForGraph(m, trace, normalized); err != nil {
		return result, err
	}
	if options.RealPrimaryInputs < 0 || options.RealPrimaryInputs > m.PrimaryInputCount() || options.Attempts < 1 || options.Attempts > 4 {
		return result, ErrDimension
	}
	inputEntries, err := checkedSize(0, trace.Frames(), m.PrimaryInputCount(), normalized.MaxMappingEntries)
	if err != nil {
		return result, err
	}
	stateEntries, err := checkedSize(0, trace.Frames(), m.RegisterCount(), normalized.MaxMappingEntries)
	if err != nil {
		return result, err
	}
	if _, err = checkedSize(inputEntries, 1, stateEntries, normalized.MaxMappingEntries); err != nil {
		return result, err
	}
	bestCount := int(^uint(0) >> 1)
	for attempt := 0; attempt < options.Attempts; attempt++ {
		priorities, priorityErr := m.cexCarePriorityInputs(trace, options.RealPrimaryInputs, attempt)
		if priorityErr != nil {
			return CexPriorityCareResult{}, priorityErr
		}
		candidate, candidateErr := m.minimizeCounterexamplePriorityAttempt(trace, priorities, normalized)
		if candidateErr != nil {
			return CexPriorityCareResult{}, candidateErr
		}
		verified, verifyErr := m.VerifyCounterexampleCare(trace, candidate, normalized)
		if verifyErr != nil || !verified {
			if verifyErr != nil {
				return CexPriorityCareResult{}, verifyErr
			}
			return CexPriorityCareResult{}, ErrInvariant
		}
		result.Candidates = append(result.Candidates, candidate)
		if count := candidate.CountOnes(); count < bestCount {
			bestCount, result.Best = count, candidate
		}
	}
	result.Verified = result.Best != nil
	return result, nil
}

// IntersectCounterexampleCare returns the bitwise intersection of compatible
// owned care masks.
func IntersectCounterexampleCare(cares ...*Counterexample) (*Counterexample, error) {
	if len(cares) == 0 || cares[0] == nil {
		return nil, ErrDimension
	}
	first := cares[0]
	if err := first.Validate(); err != nil {
		return nil, err
	}
	result := first.Clone()
	for _, care := range cares[1:] {
		if care == nil || care.Registers() != first.Registers() || care.Inputs() != first.Inputs() || care.Frames() != first.Frames() || care.Property != first.Property {
			return nil, ErrDimension
		}
		if err := care.Validate(); err != nil {
			return nil, err
		}
		for bit := 0; bit < result.BitCount(); bit++ {
			value, _ := result.Bit(bit)
			other, _ := care.Bit(bit)
			_ = result.SetBit(bit, value && other)
		}
	}
	return result, nil
}

// ExtendCounterexampleCareToObjects ternary-simulates a PI care mask and
// returns one column per source object indicating whether its value is known.
// Initial registers retain the concrete trace value; unlike the source's
// zero-only scratch initialization, nonzero initial states are supported.
func (m *GiaManager) ExtendCounterexampleCareToObjects(trace, care *Counterexample, options CounterexampleToolOptions) (*Counterexample, error) {
	options, err := normalizeCounterexampleToolOptions(options)
	if err != nil {
		return nil, err
	}
	if err = checkCexCarePlainGraph(m); err != nil {
		return nil, err
	}
	if err = checkCounterexampleForGraph(m, trace, options); err != nil {
		return nil, err
	}
	if care == nil || care.Registers() != 0 || care.Inputs() != m.PrimaryInputCount() || care.Frames() != trace.Frames() || care.Property != trace.Property {
		return nil, ErrDimension
	}
	if err = care.Validate(); err != nil {
		return nil, err
	}
	if _, err = checkedSize(0, trace.Frames(), m.ObjectCount(), options.MaxMappingEntries); err != nil {
		return nil, err
	}
	if _, err = checkedSize(0, trace.Frames(), m.ObjectCount(), options.MaxBits); err != nil {
		return nil, err
	}
	result, err := NewCounterexample(0, m.ObjectCount(), trace.Frames())
	if err != nil {
		return nil, err
	}
	result.Property = trace.Property
	state := make([]GiaTernary, m.RegisterCount())
	for register := range state {
		value, _ := trace.Initial(register)
		state[register] = GiaTernaryZero
		if value {
			state[register] = GiaTernaryOne
		}
	}
	values := make([]GiaTernary, m.ObjectCount())
	read := func(literal GiaLit) GiaTernary {
		value := values[literal.ID()]
		if literal.Complemented() {
			return giaTernaryNot(value)
		}
		return value
	}
	for frame := 0; frame < trace.Frames(); frame++ {
		values[0] = GiaTernaryZero
		for input, pi := range m.cis[:m.PrimaryInputCount()] {
			known, _ := care.Input(frame, input)
			values[pi] = GiaTernaryUnknown
			if known {
				concrete, _ := trace.Input(frame, input)
				values[pi] = GiaTernaryZero
				if concrete {
					values[pi] = GiaTernaryOne
				}
			}
		}
		for register, ro := range m.cis[m.PrimaryInputCount():] {
			values[ro] = state[register]
		}
		for index, object := range m.objects {
			if cexToolStopped(options) {
				return nil, ErrCancelled
			}
			id := GiaID(index)
			if object.IsAnd() {
				first, _ := object.Fanin0(id)
				second, _ := object.Fanin1(id)
				values[id] = giaTernaryAnd(read(first), read(second))
			} else if object.IsCO() {
				driver, _ := object.Fanin0(id)
				values[id] = read(driver)
			}
		}
		for id, value := range values {
			if value != GiaTernaryUnknown {
				_ = result.SetInput(frame, id, true)
			}
		}
		for register, ri := range m.cos[m.PrimaryOutputCount():] {
			state[register] = values[ri]
		}
	}
	if values[m.cos[trace.Property]] != GiaTernaryOne {
		return nil, ErrInvariant
	}
	return result, nil
}

// MinimizeCounterexampleCarePriority converts an AIG through the concrete GIA
// representation and applies the four source priority policies.
func (m *AigManager) MinimizeCounterexampleCarePriority(trace *Counterexample, options CexPriorityCareOptions) (CexPriorityCareResult, error) {
	gia, _, err := GiaFromAig(m)
	if err != nil {
		return CexPriorityCareResult{}, err
	}
	return gia.MinimizeCounterexampleCarePriority(trace, options)
}

// VerifyCounterexampleCare converts an AIG and returns the source verifier's
// useful Boolean result instead of printing it.
func (m *AigManager) VerifyCounterexampleCare(trace, care *Counterexample, options CounterexampleToolOptions) (bool, error) {
	gia, _, err := GiaFromAig(m)
	if err != nil {
		return false, err
	}
	return gia.VerifyCounterexampleCare(trace, care, options)
}

// MinimizeCounterexampleCareSAT replaces the source wrapper's unavailable
// external Bmc_CexCareSatBasedMinimizeAig dependency. Low effort performs one
// deterministic implication-aware pass; high effort uses the verified
// care/essential/minimized pipeline. Both results are independently verified.
func (m *AigManager) MinimizeCounterexampleCareSAT(trace *Counterexample, realPrimaryInputs int, highEffort bool, options CounterexampleToolOptions) (*Counterexample, error) {
	gia, _, err := GiaFromAig(m)
	if err != nil {
		return nil, err
	}
	if realPrimaryInputs < 0 || realPrimaryInputs > gia.PrimaryInputCount() {
		return nil, ErrDimension
	}
	if highEffort {
		analysis, analysisErr := gia.AnalyzeCounterexample(trace, options)
		if analysisErr != nil {
			return nil, analysisErr
		}
		return analysis.Minimized.Clone(), nil
	}
	states, implications, err := gia.CounterexampleInnerStates(trace, options)
	if err != nil {
		return nil, err
	}
	care, err := gia.CounterexampleCareBits(states, implications, nil, false, options)
	if err != nil {
		return nil, err
	}
	verified, err := gia.VerifyCounterexampleCare(trace, care, options)
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, ErrInvariant
	}
	return care, nil
}
