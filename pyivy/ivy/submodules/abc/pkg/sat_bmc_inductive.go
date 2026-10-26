package abc

// M-inductiveness checking and register-set minimization from
// src/sat/bmc/bmcICheck.c.

import "slices"

type GiaMInductiveStatus uint8

const (
	GiaMInductiveUnknown GiaMInductiveStatus = iota
	GiaMInductive
	GiaMNotInductive
)

// GiaMInductiveOptions bounds construction and SAT work. Frames is the exact
// M value for CheckMInductiveSet and MinimizeMInductiveCore, and the inclusive
// maximum M for SearchMInductiveRegisters. Empty selects the source core
// check's all-disabled starting set. Reverse and BackwardTopological select
// the greedy search order.
type GiaMInductiveOptions struct {
	Frames, MaxObjects, MaxMappingEntries int
	MaxVariables, MaxClauses, MaxLiterals int
	Empty, Reverse, BackwardTopological   bool
	SAT                                   SatLimits
}

type GiaMInductiveResult struct {
	Status        GiaMInductiveStatus
	Reason        SatStopReason
	Selected      []bool
	Frame         int
	CheckedFrames int
	SATChecks     int
	Stats         SatStats
}

type giaMInductiveProblem struct {
	cnf       *AigCNF
	selectors []SatLit
	root      SatLit
	options   GiaMInductiveOptions
}

func normalizeGiaMInductiveOptions(options GiaMInductiveOptions) (GiaMInductiveOptions, error) {
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

// unfoldMInductiveFrame appends one transition relation and returns its
// primary outputs and next state. Inputs can be reused by the second copy.
func (manager *GiaManager) unfoldMInductiveFrame(target *GiaManager, state, reusedInputs []GiaLit, options GiaMInductiveOptions) ([]GiaLit, []GiaLit, []GiaLit, error) {
	if target == nil || len(state) != manager.RegisterCount() || (reusedInputs != nil && len(reusedInputs) != manager.PrimaryInputCount()) {
		return nil, nil, nil, ErrDimension
	}
	if manager.ObjectCount() > options.MaxMappingEntries {
		return nil, nil, nil, ErrLimit
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
				return nil, nil, nil, ErrCancelled
			}
			var err error
			inputs[input], err = target.AppendCI()
			if err != nil {
				return nil, nil, nil, err
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
			return nil, nil, nil, ErrCancelled
		}
		var err error
		mapping[index], err = hashGiaCopiedGate(manager, target, GiaID(index), mapping)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	outputs := make([]GiaLit, manager.PrimaryOutputCount())
	for output, co := range manager.cos[:manager.PrimaryOutputCount()] {
		driver, _ := manager.objects[co].Fanin0(co)
		var err error
		outputs[output], err = mappedGiaSignal(mapping, driver)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	next := make([]GiaLit, manager.RegisterCount())
	for register, ri := range manager.cos[manager.PrimaryOutputCount():] {
		driver, _ := manager.objects[ri].Fanin0(ri)
		var err error
		next[register], err = mappedGiaSignal(mapping, driver)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return outputs, next, inputs, nil
}

func (manager *GiaManager) deriveMInductiveProblem(options GiaMInductiveOptions) (*giaMInductiveProblem, error) {
	if err := checkBmciPlainGraph(manager); err != nil {
		return nil, err
	}
	var err error
	options, err = normalizeGiaMInductiveOptions(options)
	if err != nil {
		return nil, err
	}
	if manager.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	graph, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	if err = graph.HashStart(); err != nil {
		return nil, err
	}
	selectors := make([]GiaLit, manager.RegisterCount())
	for register := range selectors {
		selectors[register], err = graph.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	left, right := make([]GiaLit, manager.RegisterCount()), make([]GiaLit, manager.RegisterCount())
	for register := range left {
		left[register], err = graph.AppendCI()
		if err == nil {
			right[register], err = graph.AppendCI()
		}
		if err != nil {
			return nil, err
		}
	}
	priorEqual, finalDifference := GiaConst1, GiaConst0
	for frame := 0; frame <= options.Frames; frame++ {
		if options.SAT.Stop != nil && options.SAT.Stop() {
			return nil, ErrCancelled
		}
		var leftOutputs, rightOutputs, inputs []GiaLit
		leftOutputs, left, inputs, err = manager.unfoldMInductiveFrame(graph, left, nil, options)
		if err == nil {
			rightOutputs, right, _, err = manager.unfoldMInductiveFrame(graph, right, inputs, options)
		}
		if err != nil {
			return nil, err
		}
		for output := range leftOutputs {
			difference, hashErr := graph.HashXor(leftOutputs[output], rightOutputs[output])
			if hashErr != nil {
				return nil, hashErr
			}
			if frame == options.Frames {
				finalDifference, err = graph.HashOr(finalDifference, difference)
			} else {
				priorEqual, err = graph.HashAnd(priorEqual, difference.Not())
			}
			if err != nil {
				return nil, err
			}
		}
		for register := range left {
			difference, hashErr := graph.HashXor(left[register], right[register])
			if hashErr != nil {
				return nil, hashErr
			}
			enabled, hashErr := graph.HashAnd(selectors[register], difference)
			if hashErr != nil {
				return nil, hashErr
			}
			if frame == options.Frames {
				finalDifference, err = graph.HashOr(finalDifference, enabled)
			} else {
				priorEqual, err = graph.HashAnd(priorEqual, enabled.Not())
			}
			if err != nil {
				return nil, err
			}
		}
	}
	root, err := graph.HashAnd(priorEqual, finalDifference)
	if err != nil {
		return nil, err
	}
	aig, mapping, err := graph.ToAig()
	if err != nil {
		return nil, err
	}
	cnf, err := EncodeAigCNF(aig)
	if err != nil {
		return nil, err
	}
	clauses, literalCount := cnf.Clauses(), 0
	for _, clause := range clauses {
		literalCount += len(clause)
	}
	if cnf.VariableCount() > options.MaxVariables || len(clauses)+1 > options.MaxClauses || literalCount+1 > options.MaxLiterals {
		return nil, ErrLimit
	}
	toSAT := func(literal GiaLit) (SatLit, error) {
		if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) {
			return InvalidSatLit, ErrInvariant
		}
		aigLiteral := mapping[literal.ID()].NotIf(literal.Complemented())
		satLiteral, literalErr := cnf.ObjectLiteral(aigLiteral.ID())
		return satLiteral ^ SatLit(aigLiteral&1), literalErr
	}
	satSelectors := make([]SatLit, len(selectors))
	for register, selector := range selectors {
		satSelectors[register], err = toSAT(selector)
		if err != nil {
			return nil, err
		}
	}
	satRoot, err := toSAT(root)
	if err != nil {
		return nil, err
	}
	return &giaMInductiveProblem{cnf: cnf, selectors: satSelectors, root: satRoot, options: options}, nil
}

func (problem *giaMInductiveProblem) newSolver() (*SatSolver, error) {
	solver, err := NewSatSolverWithLimit(problem.options.MaxVariables)
	if err != nil {
		return nil, err
	}
	ok, err := problem.cnf.loadIntoWithStop(solver, 0, problem.options.SAT.Stop)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvariant
	}
	ok, err = solver.AddClause([]SatLit{problem.root})
	if err != nil {
		return nil, err
	}
	if !ok {
		return solver, nil
	}
	return solver, nil
}

func (problem *giaMInductiveProblem) assumptions(selected []bool) ([]SatLit, error) {
	if len(selected) != len(problem.selectors) {
		return nil, ErrDimension
	}
	assumptions := make([]SatLit, len(selected))
	for register, enabled := range selected {
		assumptions[register] = problem.selectors[register]
		if !enabled {
			assumptions[register] = assumptions[register].Not()
		}
	}
	return assumptions, nil
}

func solveMInductive(solver *SatSolver, problem *giaMInductiveProblem, selected []bool) (SatResult, error) {
	assumptions, err := problem.assumptions(selected)
	if err != nil {
		return SatResult{}, err
	}
	return solver.Solve(assumptions, problem.options.SAT)
}

func mInductiveResult(answer SatResult, selected []bool, frame int) GiaMInductiveResult {
	result := GiaMInductiveResult{Reason: answer.Reason, Selected: slices.Clone(selected), Frame: frame, SATChecks: 1, Stats: answer.Stats}
	switch answer.Status {
	case SatUnsatisfiable:
		result.Status = GiaMInductive
	case SatSatisfiable:
		result.Status = GiaMNotInductive
	default:
		result.Status = GiaMInductiveUnknown
	}
	return result
}

// CheckMInductiveSet checks one exact M and one selected register set.
func (manager *GiaManager) CheckMInductiveSet(selected []bool, options GiaMInductiveOptions) (GiaMInductiveResult, error) {
	problem, err := manager.deriveMInductiveProblem(options)
	if err != nil {
		return GiaMInductiveResult{}, err
	}
	solver, err := problem.newSolver()
	if err != nil {
		return GiaMInductiveResult{}, err
	}
	answer, err := solveMInductive(solver, problem, selected)
	if err != nil {
		return GiaMInductiveResult{}, err
	}
	result := mInductiveResult(answer, selected, options.Frames)
	if result.Status == GiaMInductive {
		result.CheckedFrames = options.Frames
	}
	return result, nil
}

// MinimizeMInductiveCore reproduces Bmc_PerformICheck's iterative UNSAT-core
// shrinking. Every removed register is absent from a solver conflict and is
// therefore redundant for the checked M-inductiveness result.
func (manager *GiaManager) MinimizeMInductiveCore(options GiaMInductiveOptions) (GiaMInductiveResult, error) {
	problem, err := manager.deriveMInductiveProblem(options)
	if err != nil {
		return GiaMInductiveResult{}, err
	}
	selected := make([]bool, manager.RegisterCount())
	if !options.Empty {
		for register := range selected {
			selected[register] = true
		}
	}
	result := GiaMInductiveResult{Selected: slices.Clone(selected), Frame: options.Frames}
	for {
		solver, solveErr := problem.newSolver()
		if solveErr != nil {
			return GiaMInductiveResult{}, solveErr
		}
		answer, solveErr := solveMInductive(solver, problem, selected)
		if solveErr != nil {
			return GiaMInductiveResult{}, solveErr
		}
		result.SATChecks++
		addSatStats(&result.Stats, answer.Stats)
		result.Reason = answer.Reason
		if answer.Status == SatSatisfiable {
			result.Status, result.Selected = GiaMNotInductive, slices.Clone(selected)
			return result, nil
		}
		if answer.Status == SatUnknown {
			result.Status, result.Selected = GiaMInductiveUnknown, slices.Clone(selected)
			return result, nil
		}
		used := make([]bool, len(problem.selectors))
		for _, literal := range answer.Conflict {
			for register, selector := range problem.selectors {
				if literal.Var() == selector.Var() {
					used[register] = true
					break
				}
			}
		}
		changed := false
		for register := range selected {
			if selected[register] && !used[register] {
				selected[register], changed = false, true
			}
		}
		if !changed || options.Empty {
			result.Status, result.Reason = GiaMInductive, SatFinished
			result.Selected = slices.Clone(selected)
			result.CheckedFrames = options.Frames
			return result, nil
		}
	}
}

// BackwardRegisterOrder returns the source's DFS order: registers reached from
// primary outputs and recursively from their next-state cones, then dangling
// registers in natural order.
func (manager *GiaManager) BackwardRegisterOrder(maxObjects int) ([]int, error) {
	return manager.backwardRegisterOrder(maxObjects, nil)
}

func (manager *GiaManager) backwardRegisterOrder(maxObjects int, stop func() bool) ([]int, error) {
	if err := checkBmciPlainGraph(manager); err != nil {
		return nil, err
	}
	if maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects > 0 && manager.ObjectCount() > maxObjects {
		return nil, ErrLimit
	}
	visited := make([]bool, manager.ObjectCount())
	visited[0] = true
	queued := make([]bool, manager.RegisterCount())
	order := make([]int, 0, manager.RegisterCount())
	roots := make([]GiaLit, 0, manager.PrimaryOutputCount()+manager.RegisterCount())
	for _, co := range manager.cos[:manager.PrimaryOutputCount()] {
		driver, _ := manager.objects[co].Fanin0(co)
		roots = append(roots, driver)
	}
	for root := 0; root < len(roots); root++ {
		stack := []GiaLit{roots[root]}
		for len(stack) > 0 {
			if stop != nil && stop() {
				return nil, ErrCancelled
			}
			last := len(stack) - 1
			literal := stack[last]
			stack = stack[:last]
			id := literal.ID()
			if visited[id] {
				continue
			}
			visited[id] = true
			object := manager.objects[id]
			if object.IsCI() {
				terminal, _ := object.TerminalIndex()
				register := terminal - manager.PrimaryInputCount()
				if register >= 0 && !queued[register] {
					queued[register] = true
					order = append(order, register)
					ri := manager.cos[manager.PrimaryOutputCount()+register]
					driver, _ := manager.objects[ri].Fanin0(ri)
					roots = append(roots, driver)
				}
				continue
			}
			if !object.IsAnd() {
				return nil, ErrInvariant
			}
			left, _ := object.Fanin0(id)
			right, _ := object.Fanin1(id)
			stack = append(stack, right, left)
		}
	}
	for register := range queued {
		if !queued[register] {
			order = append(order, register)
		}
	}
	return order, nil
}

// SearchMInductiveRegisters reproduces Bmc_PerformISearch: it carries a
// selected set through M=1..Frames and greedily removes registers in the
// requested natural or backward-topological order.
func (manager *GiaManager) SearchMInductiveRegisters(options GiaMInductiveOptions) (GiaMInductiveResult, error) {
	var err error
	options, err = normalizeGiaMInductiveOptions(options)
	if err != nil {
		return GiaMInductiveResult{}, err
	}
	if err := checkBmciPlainGraph(manager); err != nil {
		return GiaMInductiveResult{}, err
	}
	selected := make([]bool, manager.RegisterCount())
	for register := range selected {
		selected[register] = true
	}
	order := make([]int, manager.RegisterCount())
	for register := range order {
		order[register] = register
	}
	if options.BackwardTopological {
		order, err = manager.backwardRegisterOrder(options.MaxMappingEntries, options.SAT.Stop)
		if err != nil {
			return GiaMInductiveResult{}, err
		}
	}
	if options.Reverse {
		slices.Reverse(order)
	}
	result := GiaMInductiveResult{Selected: slices.Clone(selected)}
	for frame := 1; frame <= options.Frames; frame++ {
		frameOptions := options
		frameOptions.Frames = frame
		problem, err := manager.deriveMInductiveProblem(frameOptions)
		if err != nil {
			return GiaMInductiveResult{}, err
		}
		solver, err := problem.newSolver()
		if err != nil {
			return GiaMInductiveResult{}, err
		}
		answer, err := solveMInductive(solver, problem, selected)
		if err != nil {
			return GiaMInductiveResult{}, err
		}
		result.Frame, result.SATChecks = frame, result.SATChecks+1
		result.Reason = answer.Reason
		addSatStats(&result.Stats, answer.Stats)
		if answer.Status == SatSatisfiable {
			result.Status, result.Selected = GiaMNotInductive, slices.Clone(selected)
			return result, nil
		}
		if answer.Status == SatUnknown {
			result.Status, result.Selected = GiaMInductiveUnknown, slices.Clone(selected)
			return result, nil
		}
		for _, register := range order {
			if !selected[register] {
				continue
			}
			selected[register] = false
			answer, err = solveMInductive(solver, problem, selected)
			if err != nil {
				return GiaMInductiveResult{}, err
			}
			result.SATChecks++
			result.Reason = answer.Reason
			addSatStats(&result.Stats, answer.Stats)
			if answer.Status == SatSatisfiable {
				selected[register] = true
			} else if answer.Status == SatUnknown {
				selected[register] = true
				result.Status, result.Selected = GiaMInductiveUnknown, slices.Clone(selected)
				return result, nil
			}
		}
		result.CheckedFrames = frame
	}
	result.Status, result.Reason = GiaMInductive, SatFinished
	result.Selected = slices.Clone(selected)
	return result, nil
}
