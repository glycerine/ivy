package abc

// Initial-state maximization from src/sat/bmc/bmcMaxi.c.

import "slices"

type GiaInitialClass uint8

const (
	GiaInitialKnownZero GiaInitialClass = iota
	GiaInitialKnownOne
	GiaInitialRelaxedZero
	GiaInitialRelaxedOne
	GiaInitialFixedZero
	GiaInitialFixedOne
)

func (value GiaInitialClass) valid() bool { return value <= GiaInitialFixedOne }

type GiaMaxUnfoldOptions struct {
	Frames, MaxObjects, MaxMappingEntries int
	UseVariables                          bool
	Stop                                  func() bool
}

type GiaMaxUnfolding struct {
	graph                       *GiaManager
	controls, data, frameInputs []GiaLit
	frames, inputs              int
}

func (unfolding *GiaMaxUnfolding) Graph() *GiaManager {
	if unfolding == nil {
		return nil
	}
	return unfolding.graph
}
func (unfolding *GiaMaxUnfolding) Controls() []GiaLit {
	if unfolding == nil {
		return nil
	}
	return slices.Clone(unfolding.controls)
}
func (unfolding *GiaMaxUnfolding) Data() []GiaLit {
	if unfolding == nil {
		return nil
	}
	return slices.Clone(unfolding.data)
}
func (unfolding *GiaMaxUnfolding) FrameInputs(frame int) ([]GiaLit, error) {
	if unfolding == nil || frame < 0 || frame >= unfolding.frames {
		return nil, ErrDimension
	}
	begin := frame * unfolding.inputs
	return slices.Clone(unfolding.frameInputs[begin : begin+unfolding.inputs]), nil
}

func normalizeGiaMaxUnfoldOptions(options GiaMaxUnfoldOptions) (GiaMaxUnfoldOptions, error) {
	if options.Frames < 1 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	return options, nil
}

// UnfoldMaxInitial implements Gia_ManMaxiUnfold. The first register-count CIs
// are controls, the next group is data, and the remaining CIs are frame-major
// primary inputs. Outputs are the final register-input functions.
func (manager *GiaManager) UnfoldMaxInitial(initial []GiaInitialClass, options GiaMaxUnfoldOptions) (*GiaMaxUnfolding, error) {
	if manager == nil {
		return nil, ErrDimension
	}
	if err := manager.Check(); err != nil {
		return nil, err
	}
	options, err := normalizeGiaMaxUnfoldOptions(options)
	if err != nil {
		return nil, err
	}
	if manager.registers == 0 || len(initial) != manager.registers || len(manager.objects) > options.MaxMappingEntries {
		return nil, ErrDimension
	}
	for id, object := range manager.objects {
		if object.IsAnd() && (object.IsBuffer() || object.IsXor()) {
			return nil, ErrUnsupported
		}
		if _, mux := manager.MuxControl(GiaID(id)); mux {
			return nil, ErrUnsupported
		}
	}
	for _, value := range initial {
		if !value.valid() {
			return nil, ErrDimension
		}
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name = manager.Name
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	result := &GiaMaxUnfolding{graph: target, frames: options.Frames, inputs: manager.PrimaryInputCount()}
	for index := 0; index < 2*manager.registers; index++ {
		if stopped() {
			return nil, ErrCancelled
		}
		literal, appendErr := target.AppendCI()
		if appendErr != nil {
			return nil, appendErr
		}
		if index < manager.registers {
			result.controls = append(result.controls, literal)
		} else {
			result.data = append(result.data, literal)
		}
	}
	state := make([]GiaLit, manager.registers)
	for index, value := range initial {
		control, data := result.controls[index], result.data[index]
		switch value {
		case GiaInitialKnownZero:
			if options.UseVariables {
				state[index], err = target.HashAnd(control, data)
			} else {
				state[index] = GiaConst0
			}
		case GiaInitialKnownOne:
			if options.UseVariables {
				state[index], err = target.HashOr(control.Not(), data)
			} else {
				state[index] = GiaConst1
			}
		case GiaInitialRelaxedZero:
			state[index], err = target.HashAnd(control, data)
		case GiaInitialRelaxedOne:
			state[index], err = target.HashOr(control.Not(), data)
		case GiaInitialFixedZero:
			state[index] = GiaConst0
		case GiaInitialFixedOne:
			state[index] = GiaConst1
		}
		if err != nil {
			return nil, err
		}
	}
	row := make([]GiaLit, len(manager.objects))
	for frame := 0; frame < options.Frames; frame++ {
		if stopped() {
			return nil, ErrCancelled
		}
		for index := range row {
			row[index] = InvalidGiaLit
		}
		row[0] = GiaConst0
		for _, id := range manager.cis[:manager.PrimaryInputCount()] {
			literal, appendErr := target.AppendCI()
			if appendErr != nil {
				return nil, appendErr
			}
			row[id] = literal
			result.frameInputs = append(result.frameInputs, literal)
		}
		for index, id := range manager.cis[manager.PrimaryInputCount():] {
			row[id] = state[index]
		}
		copyLiteral := func(literal GiaLit) (GiaLit, error) {
			mapped := row[literal.ID()]
			if mapped == InvalidGiaLit {
				return InvalidGiaLit, ErrInvariant
			}
			return mapped.NotIf(literal.Complemented()), nil
		}
		for id, object := range manager.objects {
			if !object.IsAnd() {
				continue
			}
			if stopped() {
				return nil, ErrCancelled
			}
			left, _ := object.Fanin0(GiaID(id))
			right, _ := object.Fanin1(GiaID(id))
			mappedLeft, mapErr := copyLiteral(left)
			if mapErr != nil {
				return nil, mapErr
			}
			mappedRight, mapErr := copyLiteral(right)
			if mapErr != nil {
				return nil, mapErr
			}
			row[id], err = target.HashAnd(mappedLeft, mappedRight)
			if err != nil {
				return nil, err
			}
		}
		next := make([]GiaLit, manager.registers)
		for index, id := range manager.cos[manager.PrimaryOutputCount():] {
			driver, faninErr := manager.objects[id].Fanin0(id)
			if faninErr != nil {
				return nil, faninErr
			}
			next[index], faninErr = copyLiteral(driver)
			if faninErr != nil {
				return nil, faninErr
			}
		}
		state = next
	}
	for _, literal := range state {
		if stopped() {
			return nil, ErrCancelled
		}
		if _, err = target.AppendCO(literal); err != nil {
			return nil, err
		}
	}
	target.HashStop()
	cleaned, remap, err := target.Cleanup()
	if err != nil {
		return nil, err
	}
	remapLiteral := func(literal GiaLit) GiaLit { return remap[literal.ID()].NotIf(literal.Complemented()) }
	for index, literal := range result.controls {
		result.controls[index] = remapLiteral(literal)
	}
	for index, literal := range result.data {
		result.data[index] = remapLiteral(literal)
	}
	for index, literal := range result.frameInputs {
		result.frameInputs[index] = remapLiteral(literal)
	}
	result.graph = cleaned
	return result, nil
}

type GiaMaxInitialOptions struct {
	Frames, MaxObjects, MaxMappingEntries int
	MaxVariables, MaxClauses, MaxLiterals int
	MaxIterations                         int
	SAT                                   SatLimits
	Stop                                  func() bool
}

type GiaMaxInitialResult struct {
	Classes       []GiaInitialClass
	CoreRegisters []int
	Status        SatStatus
	Reason        SatStopReason
	Iterations    int
	FixedPoint    bool
	Stats         SatStats
}

func normalizeGiaMaxInitialOptions(options GiaMaxInitialOptions) (GiaMaxInitialOptions, error) {
	if options.Frames < 1 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 || options.MaxIterations < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 4 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 12 << 20
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 1 << 20
	}
	return options, nil
}

// MaximizeInitialState implements Gia_ManMaxiPerform/Gia_ManMaxiTest. A nil
// initial vector selects candidate known-zero values. Final UNSAT conflicts
// repeatedly shrink the controls that must remain false; omitted known values
// become their relaxed counterparts.
func (manager *GiaManager) MaximizeInitialState(initial []GiaInitialClass, options GiaMaxInitialOptions) (GiaMaxInitialResult, error) {
	fail := func(err error) (GiaMaxInitialResult, error) { return GiaMaxInitialResult{}, err }
	if manager == nil {
		return fail(ErrDimension)
	}
	if err := manager.Check(); err != nil {
		return fail(err)
	}
	options, err := normalizeGiaMaxInitialOptions(options)
	if err != nil {
		return fail(err)
	}
	if manager.registers == 0 {
		return fail(ErrDimension)
	}
	if initial == nil {
		initial = make([]GiaInitialClass, manager.registers)
	} else {
		initial = slices.Clone(initial)
	}
	if len(initial) != manager.registers {
		return fail(ErrDimension)
	}
	baseOptions := GiaMaxUnfoldOptions{Frames: options.Frames, MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, Stop: options.Stop}
	baseline, err := manager.UnfoldMaxInitial(initial, baseOptions)
	if err != nil {
		return fail(err)
	}
	baseOptions.UseVariables = true
	variable, err := manager.UnfoldMaxInitial(initial, baseOptions)
	if err != nil {
		return fail(err)
	}
	miter, err := BuildGiaDesignMiter(baseline.Graph(), variable.Graph(), GiaMiterOptions{GiaCofactorOptions: GiaCofactorOptions{MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, Stop: options.Stop}})
	if err != nil {
		return fail(err)
	}
	aig, _, err := miter.Graph.ToAig()
	if err != nil {
		return fail(err)
	}
	cnf, err := EncodeAigCNF(aig)
	if err != nil {
		return fail(err)
	}
	literalCount := 0
	for _, clause := range cnf.clauses {
		literalCount += len(clause)
	}
	if cnf.variables > options.MaxVariables || len(cnf.clauses)+1 > options.MaxClauses || literalCount+len(cnf.outputs) > options.MaxLiterals {
		return fail(ErrLimit)
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return fail(err)
	}
	if _, err = cnf.LoadInto(solver, 0); err != nil {
		return fail(err)
	}
	if _, err = solver.AddClause(cnf.outputs); err != nil {
		return fail(err)
	}
	inputs := cnf.InputLiterals()
	if len(inputs) < manager.registers {
		return fail(ErrInvariant)
	}
	assumptions := make([]SatLit, 0, manager.registers)
	for index, value := range initial {
		if !value.valid() {
			return fail(ErrDimension)
		}
		if value == GiaInitialKnownZero || value == GiaInitialKnownOne {
			assumptions = append(assumptions, inputs[index].Not())
		}
	}
	stop := options.Stop
	callerStop := options.SAT.Stop
	limits := options.SAT
	limits.Stop = func() bool { return stop != nil && stop() || callerStop != nil && callerStop() }
	result := GiaMaxInitialResult{Classes: slices.Clone(initial), Status: SatUnknown}
	for result.Iterations < options.MaxIterations {
		if stop != nil && stop() {
			return fail(ErrCancelled)
		}
		result.Iterations++
		answer, solveErr := solver.Solve(assumptions, limits)
		if solveErr != nil {
			return fail(solveErr)
		}
		result.Status, result.Reason, result.Stats = answer.Status, answer.Reason, answer.Stats
		if answer.Status != SatUnsatisfiable {
			break
		}
		if len(answer.Conflict) == len(assumptions) {
			result.FixedPoint = true
			break
		}
		next := make([]SatLit, len(answer.Conflict))
		for index, literal := range answer.Conflict {
			next[index] = literal.Not()
		}
		assumptions = next
	}
	if result.Status == SatUnsatisfiable && !result.FixedPoint && result.Iterations == options.MaxIterations {
		return fail(ErrLimit)
	}
	retained := make([]bool, cnf.variables)
	for _, literal := range assumptions {
		retained[literal.Var()] = true
	}
	for index, value := range result.Classes {
		if (value == GiaInitialKnownZero || value == GiaInitialKnownOne) && !retained[inputs[index].Var()] {
			result.Classes[index] = value + 2
		}
		if value == GiaInitialKnownZero || value == GiaInitialKnownOne {
			if retained[inputs[index].Var()] {
				result.CoreRegisters = append(result.CoreRegisters, index)
			}
		}
	}
	return result, nil
}
