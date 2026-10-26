package abc

// Lazy directional CNF loading from src/sat/bmc/bmcLoad.c.

import "slices"

type GiaLazyCNFOptions struct {
	MaxVariables, MaxClauses, MaxLiterals, MaxIterations int
	Stop                                                 func() bool
}

type GiaLazyCNFStats struct {
	Variables, Clauses, Literals int
	LoadRequests, Loadings       int
}

type GiaLazyCNFSolveResult struct {
	Result     SatResult
	Inputs     []bool
	Output     int
	Iterations int
	Stats      GiaLazyCNFStats
}

// GiaLazyCNF is a concrete, owned directional encoder. SAT variable zero is
// reserved exactly as in Bmc_LoadStart; the constant-zero object receives the
// next variable and a negative unit clause.
type GiaLazyCNF struct {
	graph                          *GiaManager
	solver                         *SatSolver
	objectToVariable               []int
	variableToObject               []GiaID
	positiveLoaded, negativeLoaded []bool
	clauses                        [][]SatLit
	options                        GiaLazyCNFOptions
	loadRequests, loadings         int
	literalCount                   int
	closed                         bool
}

func normalizeGiaLazyCNFOptions(options GiaLazyCNFOptions, objects int) (GiaLazyCNFOptions, error) {
	if options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 || options.MaxIterations < 0 {
		return options, ErrDimension
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
		options.MaxIterations = 2*objects + 1
	}
	if options.MaxVariables < 2 || options.MaxClauses < 1 || options.MaxLiterals < 1 || options.MaxIterations < 1 {
		return options, ErrLimit
	}
	return options, nil
}

func NewGiaLazyCNF(graph *GiaManager, options GiaLazyCNFOptions) (*GiaLazyCNF, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	options, err := normalizeGiaLazyCNFOptions(options, len(graph.objects))
	if err != nil {
		return nil, err
	}
	for id, object := range graph.objects {
		if object.IsAnd() && (object.IsBuffer() || object.IsXor()) {
			return nil, ErrUnsupported
		}
		if _, mux := graph.MuxControl(GiaID(id)); mux {
			return nil, ErrUnsupported
		}
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return nil, err
	}
	if err = solver.SetVariableCount(1); err != nil {
		return nil, err
	}
	manager := &GiaLazyCNF{
		graph: graph.Clone(), solver: solver, options: options,
		objectToVariable: make([]int, len(graph.objects)),
		variableToObject: []GiaID{InvalidGiaID},
		positiveLoaded:   make([]bool, len(graph.objects)),
		negativeLoaded:   make([]bool, len(graph.objects)),
	}
	for index := range manager.objectToVariable {
		manager.objectToVariable[index] = -1
	}
	constant, err := manager.variableFor(0)
	if err != nil {
		return nil, err
	}
	constantFalse, err := MakeSatLit(constant, true)
	if err != nil {
		return nil, err
	}
	if err = manager.addClause([]SatLit{constantFalse}); err != nil {
		return nil, err
	}
	return manager, nil
}

func (manager *GiaLazyCNF) usable() bool {
	return manager != nil && !manager.closed && manager.graph != nil && manager.solver != nil
}

func (manager *GiaLazyCNF) clone() *GiaLazyCNF {
	copyManager := *manager
	copyManager.objectToVariable = slices.Clone(manager.objectToVariable)
	copyManager.variableToObject = slices.Clone(manager.variableToObject)
	copyManager.positiveLoaded = slices.Clone(manager.positiveLoaded)
	copyManager.negativeLoaded = slices.Clone(manager.negativeLoaded)
	copyManager.clauses = make([][]SatLit, len(manager.clauses))
	for index := range manager.clauses {
		copyManager.clauses[index] = slices.Clone(manager.clauses[index])
	}
	copySolver := manager.solver.cloneState()
	copyManager.solver = &copySolver
	return &copyManager
}

func (manager *GiaLazyCNF) stopped() bool {
	return manager.options.Stop != nil && manager.options.Stop()
}

func (manager *GiaLazyCNF) variableFor(id GiaID) (SatVar, error) {
	if !manager.usable() || uint64(id) >= uint64(len(manager.objectToVariable)) {
		return 0, ErrInvalidLiteral
	}
	if variable := manager.objectToVariable[id]; variable >= 0 {
		return SatVar(variable), nil
	}
	if len(manager.variableToObject) == manager.options.MaxVariables {
		return 0, ErrLimit
	}
	variable := SatVar(len(manager.variableToObject))
	if err := manager.solver.SetVariableCount(int(variable) + 1); err != nil {
		return 0, err
	}
	manager.objectToVariable[id] = int(variable)
	manager.variableToObject = append(manager.variableToObject, id)
	return variable, nil
}

func (manager *GiaLazyCNF) satLiteral(literal GiaLit) (SatLit, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(manager.graph.objects)) || manager.graph.objects[literal.ID()].IsCO() {
		return 0, ErrInvalidLiteral
	}
	variable, err := manager.variableFor(literal.ID())
	if err != nil {
		return 0, err
	}
	return MakeSatLit(variable, literal.Complemented())
}

func (manager *GiaLazyCNF) addClause(clause []SatLit) error {
	if manager.stopped() {
		return ErrCancelled
	}
	if len(manager.clauses) == manager.options.MaxClauses || manager.literalCount > manager.options.MaxLiterals-len(clause) {
		return ErrLimit
	}
	ok, err := manager.solver.AddClause(clause)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvariant
	}
	manager.clauses = append(manager.clauses, slices.Clone(clause))
	manager.literalCount += len(clause)
	return nil
}

func (manager *GiaLazyCNF) loadLiteral(literal GiaLit) (bool, error) {
	if manager.stopped() {
		return false, ErrCancelled
	}
	encoded, err := manager.satLiteral(literal)
	if err != nil {
		return false, err
	}
	manager.loadRequests++
	id := literal.ID()
	object := manager.graph.objects[id]
	if object.IsCI() || object.IsConstant0() {
		return false, nil
	}
	if !object.IsAnd() || object.IsBuffer() || object.IsXor() {
		return false, ErrUnsupported
	}
	if _, mux := manager.graph.MuxControl(id); mux {
		return false, ErrUnsupported
	}
	loaded := manager.positiveLoaded
	if literal.Complemented() {
		loaded = manager.negativeLoaded
	}
	if loaded[id] {
		return false, nil
	}
	left, err := object.Fanin0(id)
	if err != nil {
		return false, err
	}
	right, err := object.Fanin1(id)
	if err != nil {
		return false, err
	}
	leftSAT, err := manager.satLiteral(left)
	if err != nil {
		return false, err
	}
	rightSAT, err := manager.satLiteral(right)
	if err != nil {
		return false, err
	}
	if literal.Complemented() {
		if err = manager.addClause([]SatLit{encoded.Not(), leftSAT.Not(), rightSAT.Not()}); err != nil {
			return false, err
		}
		manager.negativeLoaded[id] = true
	} else {
		if err = manager.addClause([]SatLit{encoded.Not(), leftSAT}); err != nil {
			return false, err
		}
		if err = manager.addClause([]SatLit{encoded.Not(), rightSAT}); err != nil {
			return false, err
		}
		manager.positiveLoaded[id] = true
	}
	manager.loadings++
	return true, nil
}

// LoadLiteral adds the one-way clauses needed when the given graph literal is
// assigned true. The update is atomic; false means the direction was already
// loaded or the literal is a CI/constant.
func (manager *GiaLazyCNF) LoadLiteral(literal GiaLit) (bool, error) {
	if !manager.usable() {
		return false, ErrInvariant
	}
	trial := manager.clone()
	loaded, err := trial.loadLiteral(literal)
	if err != nil {
		return false, err
	}
	*manager = *trial
	return loaded, nil
}

func (manager *GiaLazyCNF) loadCone(root GiaLit) error {
	if _, err := manager.satLiteral(root); err != nil {
		return err
	}
	seen := make([]bool, len(manager.graph.objects))
	stack := []GiaID{root.ID()}
	for len(stack) != 0 {
		if manager.stopped() {
			return ErrCancelled
		}
		last := len(stack) - 1
		id := stack[last]
		stack = stack[:last]
		if seen[id] {
			continue
		}
		seen[id] = true
		object := manager.graph.objects[id]
		if !object.IsAnd() {
			continue
		}
		plain, _ := MakeGiaLit(id, false)
		if _, err := manager.loadLiteral(plain); err != nil {
			return err
		}
		if _, err := manager.loadLiteral(plain.Not()); err != nil {
			return err
		}
		left, _ := object.Fanin0(id)
		right, _ := object.Fanin1(id)
		stack = append(stack, right.ID(), left.ID())
	}
	return nil
}

// LoadCone eagerly loads both directions of every AND in a root cone.
func (manager *GiaLazyCNF) LoadCone(root GiaLit) error {
	if !manager.usable() {
		return ErrInvariant
	}
	trial := manager.clone()
	if err := trial.loadCone(root); err != nil {
		return err
	}
	*manager = *trial
	return nil
}

func (manager *GiaLazyCNF) Stats() GiaLazyCNFStats {
	if !manager.usable() {
		return GiaLazyCNFStats{}
	}
	return GiaLazyCNFStats{Variables: len(manager.variableToObject), Clauses: len(manager.clauses), Literals: manager.literalCount, LoadRequests: manager.loadRequests, Loadings: manager.loadings}
}

func (manager *GiaLazyCNF) Clauses() [][]SatLit {
	if !manager.usable() {
		return nil
	}
	result := make([][]SatLit, len(manager.clauses))
	for index := range manager.clauses {
		result[index] = slices.Clone(manager.clauses[index])
	}
	return result
}

func (manager *GiaLazyCNF) ObjectVariable(id GiaID) (SatVar, bool) {
	if !manager.usable() || uint64(id) >= uint64(len(manager.objectToVariable)) || manager.objectToVariable[id] < 0 {
		return 0, false
	}
	return SatVar(manager.objectToVariable[id]), true
}

func (manager *GiaLazyCNF) firstModelMismatch(model []bool) (GiaLit, bool, error) {
	for variable := 1; variable < len(manager.variableToObject); variable++ {
		id := manager.variableToObject[variable]
		object := manager.graph.objects[id]
		if !object.IsAnd() {
			continue
		}
		left, _ := object.Fanin0(id)
		right, _ := object.Fanin1(id)
		leftVariable, leftFound := manager.ObjectVariable(left.ID())
		rightVariable, rightFound := manager.ObjectVariable(right.ID())
		if variable >= len(model) {
			return InvalidGiaLit, false, ErrInvariant
		}
		if !leftFound || !rightFound {
			literal, _ := MakeGiaLit(id, model[variable] == false)
			return literal, true, nil
		}
		if int(leftVariable) >= len(model) || int(rightVariable) >= len(model) {
			return InvalidGiaLit, false, ErrInvariant
		}
		leftValue := model[leftVariable] != left.Complemented()
		rightValue := model[rightVariable] != right.Complemented()
		if model[variable] != (leftValue && rightValue) {
			literal, _ := MakeGiaLit(id, model[variable] == false)
			return literal, true, nil
		}
	}
	return InvalidGiaLit, false, nil
}

// SolveOutput solves one primary output. Eager loads the complete cone first.
// Lazy mode starts with only the output direction and refines each spurious SAT
// model until direct graph semantics agree. UNSAT from any directional subset
// is already sound because every emitted clause is implied by the full gate.
func (manager *GiaLazyCNF) SolveOutput(output int, eager bool, limits SatLimits) (GiaLazyCNFSolveResult, error) {
	result := GiaLazyCNFSolveResult{Output: output}
	if !manager.usable() {
		return result, ErrDimension
	}
	trial := manager.clone()
	result, err := trial.solveOutput(output, eager, limits)
	if err != nil {
		return result, err
	}
	*manager = *trial
	return result, nil
}

func (manager *GiaLazyCNF) solveOutput(output int, eager bool, limits SatLimits) (GiaLazyCNFSolveResult, error) {
	result := GiaLazyCNFSolveResult{Output: output}
	if !manager.usable() || output < 0 || output >= manager.graph.PrimaryOutputCount() {
		return result, ErrDimension
	}
	co := manager.graph.objects[manager.graph.cos[output]]
	driver, err := co.Fanin0(manager.graph.cos[output])
	if err != nil {
		return result, err
	}
	if eager {
		if err = manager.loadCone(driver); err != nil {
			return result, err
		}
	} else if _, err = manager.loadLiteral(driver); err != nil {
		return result, err
	}
	assumption, err := manager.satLiteral(driver)
	if err != nil {
		return result, err
	}
	for result.Iterations < manager.options.MaxIterations {
		if manager.stopped() {
			return result, ErrCancelled
		}
		result.Iterations++
		answer, solveErr := manager.solver.Solve([]SatLit{assumption}, limits)
		if solveErr != nil {
			return result, solveErr
		}
		result.Result = answer
		if answer.Status != SatSatisfiable {
			result.Stats = manager.Stats()
			return result, nil
		}
		mismatch, found, mismatchErr := manager.firstModelMismatch(answer.Model)
		if mismatchErr != nil {
			return result, mismatchErr
		}
		if found {
			loaded, loadErr := manager.loadLiteral(mismatch)
			if loadErr != nil {
				return result, loadErr
			}
			if !loaded {
				return result, ErrInvariant
			}
			continue
		}
		result.Inputs = make([]bool, len(manager.graph.cis))
		for index, id := range manager.graph.cis {
			if variable, found := manager.ObjectVariable(id); found {
				if int(variable) >= len(answer.Model) {
					return result, ErrInvariant
				}
				result.Inputs[index] = answer.Model[variable]
			}
		}
		outputs, evaluateErr := manager.graph.Evaluate(result.Inputs)
		if evaluateErr != nil || !outputs[output] {
			return result, ErrInvariant
		}
		result.Stats = manager.Stats()
		return result, nil
	}
	return result, ErrLimit
}

func (manager *GiaLazyCNF) Close() {
	if manager == nil || manager.closed {
		return
	}
	manager.graph = nil
	manager.solver = nil
	manager.objectToVariable = nil
	manager.variableToObject = nil
	manager.positiveLoaded = nil
	manager.negativeLoaded = nil
	manager.clauses = nil
	manager.literalCount = 0
	manager.closed = true
}
