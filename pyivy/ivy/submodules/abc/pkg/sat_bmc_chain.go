package abc

// Counterexample-chain construction from src/sat/bmc/bmcChain.c.

import (
	"errors"
	"slices"
)

// ChainStopReason explains why ChainBMC returned. A finite BMC bound never
// proves the remaining outputs safe.
type ChainStopReason uint8

const (
	ChainAllFailed ChainStopReason = iota
	ChainBoundReached
	ChainIterationLimit
	ChainSATLimit
	ChainCancelled
	ChainResourceLimit
)

// ChainOptions bounds every owned phase of counterexample-chain construction.
// Frames is the number of frames checked by each BMC segment. SAT limits apply
// independently to each BMC/output query, matching the source routines.
type ChainOptions struct {
	Frames, MaxIterations                       int
	MaxObjects, MaxMappingEntries, MaxVariables int
	SAT                                         SatLimits
}

// ChainOutputSearch contains the outputs satisfiable in the graph's current
// zero state. Counterexamples are one-frame witnesses in that graph's current
// (possibly phase-recoded) coordinates and correspond positionally to Outputs.
type ChainOutputSearch struct {
	Status          SatStatus
	Reason          SatStopReason
	Outputs         []int
	Counterexamples []*Counterexample
	Stats           SatStats
}

// ChainSegment keeps coordinate changes explicit. Path starts from zero in the
// graph before this segment. Reached is that graph's register state at the
// start of Path's failing frame. StateCounterexamples start from zero in the
// graph after recoding Reached to zero. They are deliberately not flattened:
// successive segments use different state coordinates.
type ChainSegment struct {
	Path                 *Counterexample
	Reached              []bool
	FailedOutputs        []int
	StateCounterexamples []*Counterexample
}

// ChainResult owns the final phase-recoded graph and all witnesses. Failed is
// the number of syntactically constant-zero primary outputs in FinalGraph.
type ChainResult struct {
	Reason     ChainStopReason
	SATReason  SatStopReason
	Segments   []ChainSegment
	Depth      int
	Failed     int
	Undecided  int
	FinalGraph *GiaManager
	Stats      SatStats
}

func normalizeChainOptions(options ChainOptions) (ChainOptions, error) {
	if options.Frames < 1 || options.MaxIterations < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 {
		return options, ErrDimension
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 10000
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	return options, nil
}

func chainStopped(options ChainOptions) bool {
	return options.SAT.Stop != nil && options.SAT.Stop()
}

// ChainFailOneOutput performs one bounded search and returns the earliest
// failing property in frame/property order. Its witness is replayed by the
// native BMC implementation before return.
func (m *GiaManager) ChainFailOneOutput(options ChainOptions) (BmcResult, error) {
	var empty BmcResult
	options, err := normalizeChainOptions(options)
	if err != nil {
		return empty, err
	}
	if m == nil {
		return empty, ErrDimension
	}
	if err = m.Check(); err != nil {
		return empty, err
	}
	aig, _, err := m.ToAig()
	if err != nil {
		return empty, err
	}
	model, err := NewAigerModel(aig, m.PrimaryOutputCount()-m.ConstraintCount(), 0, m.ConstraintCount(), nil)
	if err != nil {
		return empty, err
	}
	return model.BoundedCheck(BmcOptions{
		Frames: options.Frames, MaxObjects: options.MaxObjects,
		MaxMappingEntries: options.MaxMappingEntries, MaxVariables: options.MaxVariables,
		SAT: options.SAT,
	})
}

// MoveChainCounterexample verifies path and recodes the state at the start of
// its failing frame to zero. It is the owned, scratch-free equivalent of
// Gia_ManVerifyCexAndMove plus Gia_ManDupWithInit.
func (m *GiaManager) MoveChainCounterexample(path *Counterexample, options ChainOptions) (*GiaManager, []bool, error) {
	options, err := normalizeChainOptions(options)
	if err != nil {
		return nil, nil, err
	}
	if m == nil || path == nil {
		return nil, nil, ErrDimension
	}
	if err = m.Check(); err != nil {
		return nil, nil, err
	}
	if err = path.Validate(); err != nil {
		return nil, nil, err
	}
	if path.Registers() != m.RegisterCount() || path.Inputs() != m.PrimaryInputCount() || path.Property < 0 || path.Property >= m.PrimaryOutputCount() {
		return nil, nil, ErrDimension
	}
	if len(m.objects) > options.MaxObjects || len(m.objects) > options.MaxMappingEntries {
		return nil, nil, ErrLimit
	}
	state := make([]bool, m.RegisterCount())
	for register := range state {
		state[register], _ = path.Initial(register)
	}
	inputs := make([]bool, m.PrimaryInputCount())
	for frame := 0; frame < path.Frames(); frame++ {
		if chainStopped(options) {
			return nil, nil, ErrCancelled
		}
		for input := range inputs {
			inputs[input], _ = path.Input(frame, input)
		}
		outputs, next, stepErr := m.Step(inputs, state)
		if stepErr != nil {
			return nil, nil, stepErr
		}
		if frame == path.LastFrame() {
			if !outputs[path.Property] {
				return nil, nil, ErrInvariant
			}
			break
		}
		state = next
	}
	reached := slices.Clone(state)
	moved, err := m.FlipRegisters(reached, GiaCofactorOptions{
		MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries,
		Stop: options.SAT.Stop,
	})
	if err != nil {
		return nil, nil, err
	}
	return moved, reached, nil
}

// PropagateChainInitialState substitutes zero for every register output,
// retains only primary outputs, structurally hashes their combinational logic,
// and cleans unreachable nodes.
func (m *GiaManager) PropagateChainInitialState(options ChainOptions) (*GiaManager, error) {
	options, err := normalizeChainOptions(options)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrDimension
	}
	if err = m.Check(); err != nil {
		return nil, err
	}
	if len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = m.Name, m.Spec
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	primaryInputs := m.PrimaryInputCount()
	primaryOutputs := m.PrimaryOutputCount()
	for id, object := range m.objects {
		if id == 0 {
			continue
		}
		if chainStopped(options) {
			return nil, ErrCancelled
		}
		switch {
		case object.IsCI():
			index, _ := object.TerminalIndex()
			if index < primaryInputs {
				mapping[id], err = g.AppendCI()
			} else {
				mapping[id] = GiaConst0
			}
		case object.IsAnd():
			mapping[id], err = hashGiaCopiedGate(m, g, GiaID(id), mapping)
		case object.IsCO():
			index, _ := object.TerminalIndex()
			if index < primaryOutputs {
				var driver GiaLit
				driver, err = object.Fanin0(GiaID(id))
				if err == nil {
					driver, err = mappedGiaSignal(mapping, driver)
				}
				if err == nil {
					mapping[id], err = g.AppendCO(driver)
				}
			}
		}
		if err != nil {
			return nil, err
		}
	}
	g.HashStop()
	g.constraints = m.constraints
	clean, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return clean, nil
}

// FindChainFailedOutputs proves which primary outputs are satisfiable in the
// current zero state. Unknown queries stop the scan and are never classified as
// failures. Each returned SAT witness is independently replayed on m.
func (m *GiaManager) FindChainFailedOutputs(options ChainOptions) (ChainOutputSearch, error) {
	var result ChainOutputSearch
	options, err := normalizeChainOptions(options)
	if err != nil {
		return result, err
	}
	if m == nil {
		return result, ErrDimension
	}
	initial, err := m.PropagateChainInitialState(options)
	if err != nil {
		return result, err
	}
	aig, _, err := initial.ToAig()
	if err != nil {
		return result, err
	}
	relation, err := EncodeAigCNF(aig)
	if err != nil {
		return result, err
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return result, err
	}
	ok, err := relation.LoadInto(solver, 0)
	if err != nil {
		return result, err
	}
	if !ok {
		return result, ErrInvariant
	}
	inputs := relation.InputLiterals()
	outputs := relation.OutputLiterals()
	result.Status, result.Reason = SatSatisfiable, SatFinished
	for output, query := range outputs {
		if chainStopped(options) {
			result.Status, result.Reason = SatUnknown, SatCancelled
			break
		}
		if query == 1 { // syntactic/encoded constant zero
			continue
		}
		answer, solveErr := solver.Solve([]SatLit{query}, options.SAT)
		if solveErr != nil {
			return result, solveErr
		}
		addSatStats(&result.Stats, answer.Stats)
		if answer.Status == SatUnknown {
			result.Status, result.Reason = SatUnknown, answer.Reason
			break
		}
		if answer.Status == SatUnsatisfiable {
			continue
		}
		trace, makeErr := NewCounterexample(m.RegisterCount(), m.PrimaryInputCount(), 1)
		if makeErr != nil {
			return result, makeErr
		}
		trace.Property = output
		for input, literal := range inputs {
			if uint64(literal.Var()) >= uint64(len(answer.Model)) {
				return result, ErrInvariant
			}
			value := answer.Model[literal.Var()] != literal.Negative()
			_ = trace.SetInput(0, input, value)
		}
		primary := make([]bool, m.PrimaryInputCount())
		for input := range primary {
			primary[input], _ = trace.Input(0, input)
		}
		values, _, replayErr := m.Step(primary, make([]bool, m.RegisterCount()))
		if replayErr != nil || output >= len(values) || !values[output] {
			return result, ErrInvariant
		}
		result.Outputs = append(result.Outputs, output)
		result.Counterexamples = append(result.Counterexamples, trace)
	}
	return result, nil
}

// CountChainNonzeroOutputs returns the number of primary outputs whose driver
// is not syntactically constant zero.
func (m *GiaManager) CountChainNonzeroOutputs() (int, error) {
	if m == nil {
		return 0, ErrDimension
	}
	if err := m.Check(); err != nil {
		return 0, err
	}
	count := 0
	for _, id := range m.cos[:m.PrimaryOutputCount()] {
		driver, _ := m.objects[id].Fanin0(id)
		if driver != GiaConst0 {
			count++
		}
	}
	return count, nil
}

// CleanupChainOutputs atomically replaces selected primary outputs by false
// and cleans unreachable logic. The source graph remains unchanged.
func (m *GiaManager) CleanupChainOutputs(outputs []int, options ChainOptions) (*GiaManager, error) {
	options, err := normalizeChainOptions(options)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrDimension
	}
	if err = m.Check(); err != nil {
		return nil, err
	}
	selected := make([]bool, m.PrimaryOutputCount())
	for _, output := range outputs {
		if output < 0 || output >= len(selected) || selected[output] {
			return nil, ErrDimension
		}
		id := m.cos[output]
		driver, _ := m.objects[id].Fanin0(id)
		if driver == GiaConst0 {
			return nil, ErrDimension
		}
		selected[output] = true
	}
	if len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for id, object := range m.objects {
		if id == 0 {
			continue
		}
		if chainStopped(options) {
			return nil, ErrCancelled
		}
		switch {
		case object.IsCI():
			mapping[id], err = g.AppendCI()
		case object.IsAnd():
			mapping[id], err = appendGiaCopiedGate(m, g, GiaID(id), mapping)
		case object.IsCO():
			index, _ := object.TerminalIndex()
			driver := GiaConst0
			if index >= len(selected) || !selected[index] {
				driver, err = object.Fanin0(GiaID(id))
				if err == nil {
					driver, err = mappedGiaSignal(mapping, driver)
				}
			}
			if err == nil {
				mapping[id], err = g.AppendCO(driver)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	g.registers, g.constraints = m.registers, m.constraints
	clean, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return clean, nil
}

// ChainBMC repeatedly moves to the first bounded failure state, enumerates all
// outputs failing in that state, and replaces them by false. Constraints are
// rejected because the source routine's second SAT phase treats every PO as a
// property and therefore cannot preserve their environment semantics.
func (m *GiaManager) ChainBMC(options ChainOptions) (ChainResult, error) {
	var result ChainResult
	options, err := normalizeChainOptions(options)
	if err != nil {
		return result, err
	}
	if m == nil {
		return result, ErrDimension
	}
	if err = m.Check(); err != nil {
		return result, err
	}
	if m.ConstraintCount() != 0 {
		return result, ErrUnsupported
	}
	current := m.Clone()
	result.Reason = ChainIterationLimit
	for iteration := 0; iteration < options.MaxIterations; iteration++ {
		remaining, countErr := current.CountChainNonzeroOutputs()
		if countErr != nil {
			return result, countErr
		}
		if remaining == 0 {
			result.Reason = ChainAllFailed
			break
		}
		bmc, bmcErr := current.ChainFailOneOutput(options)
		if bmcErr != nil {
			return result, bmcErr
		}
		addSatStats(&result.Stats, bmc.Stats)
		if bmc.Status != BmcCounterexample {
			result.SATReason = bmc.SatReason
			switch bmc.Reason {
			case BmcCancelled:
				result.Reason = ChainCancelled
			case BmcSatLimit:
				result.Reason = ChainSATLimit
			case BmcResourceLimit:
				result.Reason = ChainResourceLimit
			default:
				result.Reason = ChainBoundReached
			}
			break
		}
		moved, reached, moveErr := current.MoveChainCounterexample(bmc.Counterexample, options)
		if moveErr != nil {
			if errors.Is(moveErr, ErrCancelled) {
				result.Reason, result.SATReason = ChainCancelled, SatCancelled
				break
			}
			if errors.Is(moveErr, ErrLimit) {
				result.Reason = ChainResourceLimit
				break
			}
			return result, moveErr
		}
		search, searchErr := moved.FindChainFailedOutputs(options)
		if searchErr != nil {
			if errors.Is(searchErr, ErrCancelled) {
				result.Reason, result.SATReason = ChainCancelled, SatCancelled
				break
			}
			if errors.Is(searchErr, ErrLimit) {
				result.Reason = ChainResourceLimit
				break
			}
			return result, searchErr
		}
		addSatStats(&result.Stats, search.Stats)
		if search.Status == SatUnknown {
			result.SATReason = search.Reason
			if search.Reason == SatCancelled {
				result.Reason = ChainCancelled
			} else {
				result.Reason = ChainSATLimit
			}
			break
		}
		if !slices.Contains(search.Outputs, bmc.Counterexample.Property) {
			return result, ErrInvariant
		}
		segment := ChainSegment{
			Path: bmc.Counterexample, Reached: reached,
			FailedOutputs:        slices.Clone(search.Outputs),
			StateCounterexamples: slices.Clone(search.Counterexamples),
		}
		result.Segments = append(result.Segments, segment)
		result.Depth += bmc.Counterexample.LastFrame()
		current, err = moved.CleanupChainOutputs(search.Outputs, options)
		if err != nil {
			if errors.Is(err, ErrCancelled) {
				result.Reason, result.SATReason = ChainCancelled, SatCancelled
				break
			}
			if errors.Is(err, ErrLimit) {
				result.Reason = ChainResourceLimit
				break
			}
			return result, err
		}
	}
	remaining, err := current.CountChainNonzeroOutputs()
	if err != nil {
		return result, err
	}
	result.FinalGraph = current
	result.Undecided = remaining
	result.Failed = current.PrimaryOutputCount() - remaining
	if remaining == 0 {
		result.Reason = ChainAllFailed
	}
	return result, nil
}
