package abc

// Incremental initial-state equivalence from src/sat/bmc/bmcBmci.c.

import "slices"

type GiaBmciInitial uint8

const (
	GiaBmciZero GiaBmciInitial = iota
	GiaBmciOne
	GiaBmciVariable
)

func (value GiaBmciInitial) valid() bool { return value <= GiaBmciVariable }

type GiaBmciOptions struct {
	Frames, MaxObjects, MaxMappingEntries int
	MaxVariables, MaxClauses, MaxLiterals int
	SAT                                   SatLimits
}

type GiaBmciFrame struct {
	Frame, Miters, SAT, Merged int
}

type GiaBmciStatus uint8

const (
	GiaBmciComplete GiaBmciStatus = iota
	GiaBmciUnknown
)

type GiaBmciResult struct {
	Status       GiaBmciStatus
	Reason       SatStopReason
	FixedPoint   bool
	Frames       []GiaBmciFrame
	Left, Right  []GiaLit
	Graph        *GiaManager
	SATChecks    int
	MergedStates int
	Stats        SatStats
}

func normalizeGiaBmciOptions(options GiaBmciOptions) (GiaBmciOptions, error) {
	if options.Frames < 1 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
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
		options.MaxLiterals = 16 << 20
	}
	return options, nil
}

func checkBmciPlainGraph(manager *GiaManager) error {
	if manager == nil || manager.RegisterCount() == 0 {
		return ErrDimension
	}
	if err := manager.Check(); err != nil {
		return err
	}
	for id, object := range manager.objects {
		if !object.IsAnd() {
			continue
		}
		if object.IsBuffer() || object.IsXor() {
			return ErrUnsupported
		}
		if _, mux := manager.MuxControl(GiaID(id)); mux {
			return ErrUnsupported
		}
	}
	return nil
}

// UnfoldBmciFrame appends one transition frame. Reused inputs must be the
// shared primary-input block previously appended for the other trajectory.
func (manager *GiaManager) UnfoldBmciFrame(target *GiaManager, state []GiaLit, reusedInputs []GiaLit, options GiaBmciOptions) ([]GiaLit, []GiaLit, error) {
	if err := checkBmciPlainGraph(manager); err != nil {
		return nil, nil, err
	}
	if target == nil || len(state) != manager.RegisterCount() || (reusedInputs != nil && len(reusedInputs) != manager.PrimaryInputCount()) {
		return nil, nil, ErrDimension
	}
	mapping := make([]GiaLit, manager.ObjectCount())
	for index := range mapping {
		mapping[index] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	inputs := slices.Clone(reusedInputs)
	if inputs == nil {
		inputs = make([]GiaLit, manager.PrimaryInputCount())
		for input := range inputs {
			if options.SAT.Stop != nil && options.SAT.Stop() {
				return nil, nil, ErrCancelled
			}
			var err error
			inputs[input], err = target.AppendCI()
			if err != nil {
				return nil, nil, err
			}
		}
	}
	for input, pi := range manager.cis[:manager.PrimaryInputCount()] {
		mapping[pi] = inputs[input]
	}
	for register, ro := range manager.cis[manager.PrimaryInputCount():] {
		mapping[ro] = state[register]
	}
	for index, object := range manager.objects {
		if !object.IsAnd() {
			continue
		}
		if options.SAT.Stop != nil && options.SAT.Stop() {
			return nil, nil, ErrCancelled
		}
		var err error
		mapping[index], err = hashGiaCopiedGate(manager, target, GiaID(index), mapping)
		if err != nil {
			return nil, nil, err
		}
	}
	next := make([]GiaLit, manager.RegisterCount())
	for register, ri := range manager.cos[manager.PrimaryOutputCount():] {
		driver, _ := manager.objects[ri].Fanin0(ri)
		var err error
		next[register], err = mappedGiaSignal(mapping, driver)
		if err != nil {
			return nil, nil, err
		}
	}
	return next, inputs, nil
}

func solveBmciLiteral(graph *GiaManager, literal GiaLit, options GiaBmciOptions) (SatResult, error) {
	aig, mapping, err := graph.ToAig()
	if err != nil {
		return SatResult{}, err
	}
	cnf, err := EncodeAigCNF(aig)
	if err != nil {
		return SatResult{}, err
	}
	clauses, literals := cnf.Clauses(), 0
	for _, clause := range clauses {
		literals += len(clause)
	}
	if cnf.VariableCount() > options.MaxVariables || len(clauses) > options.MaxClauses || literals > options.MaxLiterals {
		return SatResult{}, ErrLimit
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return SatResult{}, err
	}
	ok, err := cnf.loadIntoWithStop(solver, 0, options.SAT.Stop)
	if err != nil {
		return SatResult{}, err
	}
	if !ok {
		return SatResult{Status: SatUnsatisfiable, Reason: SatFinished}, nil
	}
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) {
		return SatResult{}, ErrInvalidLiteral
	}
	aigLiteral := mapping[literal.ID()].NotIf(literal.Complemented())
	assumption, err := cnf.ObjectLiteral(aigLiteral.ID())
	if err != nil {
		return SatResult{}, err
	}
	assumption ^= SatLit(aigLiteral & 1)
	return solver.Solve([]SatLit{assumption}, options.SAT)
}

// CompareInitialStateEvolution proves or refutes register equivalence for two
// symbolic initial states under shared primary inputs, merging every UNSAT
// register miter before the next frame.
func (manager *GiaManager) CompareInitialStateEvolution(leftInitial, rightInitial []GiaBmciInitial, options GiaBmciOptions) (GiaBmciResult, error) {
	var result GiaBmciResult
	if err := checkBmciPlainGraph(manager); err != nil {
		return result, err
	}
	options, err := normalizeGiaBmciOptions(options)
	if err != nil {
		return result, err
	}
	if len(leftInitial) != manager.RegisterCount() || len(rightInitial) != manager.RegisterCount() || manager.ObjectCount() > options.MaxMappingEntries {
		return result, ErrDimension
	}
	for _, values := range [][]GiaBmciInitial{leftInitial, rightInitial} {
		for _, value := range values {
			if !value.valid() {
				return result, ErrDimension
			}
		}
	}
	graph, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return result, err
	}
	graph.Name = manager.Name
	if err = graph.HashStart(); err != nil {
		return result, err
	}
	makeState := func(initial []GiaBmciInitial) ([]GiaLit, error) {
		state := make([]GiaLit, len(initial))
		for register, value := range initial {
			switch value {
			case GiaBmciZero:
				state[register] = GiaConst0
			case GiaBmciOne:
				state[register] = GiaConst1
			case GiaBmciVariable:
				state[register], err = graph.AppendCI()
			}
			if err != nil {
				return nil, err
			}
		}
		return state, nil
	}
	left, err := makeState(leftInitial)
	if err != nil {
		return result, err
	}
	right, err := makeState(rightInitial)
	if err != nil {
		return result, err
	}
	result.Status, result.Reason = GiaBmciComplete, SatFinished
	for frame := 0; frame < options.Frames; frame++ {
		if options.SAT.Stop != nil && options.SAT.Stop() {
			return GiaBmciResult{}, ErrCancelled
		}
		var inputs []GiaLit
		var unfoldErr error
		left, inputs, unfoldErr = manager.UnfoldBmciFrame(graph, left, nil, options)
		if unfoldErr != nil {
			return GiaBmciResult{}, unfoldErr
		}
		right, _, unfoldErr = manager.UnfoldBmciFrame(graph, right, inputs, options)
		if unfoldErr != nil {
			return GiaBmciResult{}, unfoldErr
		}
		frameResult := GiaBmciFrame{Frame: frame + 1}
		for register := range left {
			if left[register] == right[register] || (left[register] < 2 && right[register] < 2) {
				continue
			}
			miter, hashErr := graph.HashXor(left[register], right[register])
			if hashErr != nil {
				return GiaBmciResult{}, hashErr
			}
			if miter == GiaConst0 {
				continue
			}
			frameResult.Miters++
			answer, solveErr := solveBmciLiteral(graph, miter, options)
			if solveErr != nil {
				return GiaBmciResult{}, solveErr
			}
			result.SATChecks++
			result.Stats.Conflicts += answer.Stats.Conflicts
			result.Stats.Propagations += answer.Stats.Propagations
			result.Stats.Decisions += answer.Stats.Decisions
			result.Stats.Inspections += answer.Stats.Inspections
			result.Stats.Restarts += answer.Stats.Restarts
			result.Stats.Learnt += answer.Stats.Learnt
			result.Stats.Deleted += answer.Stats.Deleted
			switch answer.Status {
			case SatSatisfiable:
				frameResult.SAT++
			case SatUnknown:
				result.Status, result.Reason = GiaBmciUnknown, answer.Reason
				result.Left, result.Right, result.Graph = slices.Clone(left), slices.Clone(right), graph
				result.Frames = append(result.Frames, frameResult)
				return result, nil
			case SatUnsatisfiable:
				if right[register] >= 2 {
					right[register] = left[register]
				} else {
					left[register] = right[register]
				}
				frameResult.Merged++
				result.MergedStates++
			}
		}
		result.Frames = append(result.Frames, frameResult)
		if frameResult.SAT == 0 {
			result.FixedPoint = true
			break
		}
	}
	result.Left, result.Right, result.Graph = slices.Clone(left), slices.Clone(right), graph
	return result, nil
}

// CompareInitialStateWithZero is the source wrapper's useful operation.
func (manager *GiaManager) CompareInitialStateWithZero(initial []GiaBmciInitial, options GiaBmciOptions) (GiaBmciResult, error) {
	return manager.CompareInitialStateEvolution(initial, make([]GiaBmciInitial, manager.RegisterCount()), options)
}
