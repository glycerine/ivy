package abc

// Constraint-aware input sensitivity from src/base/abci/abcSense.c.

import "slices"

type ABCSensitivityOptions struct {
	MaxInputs, MaxObjects, MaxWork        int
	MaxVariables, MaxClauses, MaxLiterals int
	SAT                                   SatLimits
	Stop                                  func() bool
}

type ABCSensitivityInput struct {
	Index  int
	Status SatStatus
	Reason SatStopReason
	Model  []bool
	Stats  SatStats
}

type ABCSensitivityResult struct {
	Sensitive           []int
	Inputs              []ABCSensitivityInput
	SAT, UNSAT, Unknown int
}

type abcSensitivityBudget struct {
	options ABCSensitivityOptions
	work    int
}

func normalizeABCSensitivityOptions(options ABCSensitivityOptions) (ABCSensitivityOptions, error) {
	if options.MaxInputs < 0 || options.MaxObjects < 0 || options.MaxWork < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return options, ErrDimension
	}
	if options.MaxInputs == 0 {
		options.MaxInputs = 1 << 20
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 8 << 20
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

func (budget *abcSensitivityBudget) step(count int) error {
	if count < 0 || budget.work > budget.options.MaxWork-count {
		return ErrLimit
	}
	budget.work += count
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (budget *abcSensitivityBudget) checkObjects(manager *AigManager) error {
	if manager.ObjectCount() > budget.options.MaxObjects {
		return ErrLimit
	}
	return nil
}

func abcSensitivityMapLiteral(mapping []AigLit, literal AigLit) (AigLit, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) || !mapping[literal.ID()].Valid() {
		return InvalidAigLit, ErrInvariant
	}
	return mapping[literal.ID()].NotIf(literal.Complemented()), nil
}

// ABCSensitivityMiter builds the two cofactors of every affected output. The
// last source output is the constraint and must hold in both cofactors; every
// preceding output contributes its cofactor XOR to the sensitivity OR.
func ABCSensitivityMiter(source *AigManager, variable int, options ABCSensitivityOptions) (*AigManager, error) {
	options, err := normalizeABCSensitivityOptions(options)
	if err != nil {
		return nil, err
	}
	if source == nil || variable < 0 {
		return nil, ErrDimension
	}
	if err = source.Check(); err != nil {
		return nil, err
	}
	if source.RegisterCount() != 0 || source.InputCount() == 0 || source.OutputCount() == 0 || variable >= source.InputCount() {
		return nil, ErrDimension
	}
	if source.InputCount() > options.MaxInputs || source.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	budget := &abcSensitivityBudget{options: options}
	if err = budget.step(1); err != nil {
		return nil, err
	}
	result := NewAigManager()
	result.Name, result.Spec = source.Name, source.Spec
	zero := make([]AigLit, source.ObjectCount())
	one := make([]AigLit, source.ObjectCount())
	for index := range zero {
		zero[index], one[index] = InvalidAigLit, InvalidAigLit
	}
	zero[0], one[0] = AigConst1, AigConst1
	for index, id := range source.cis {
		if err = budget.step(1); err != nil {
			return nil, err
		}
		input, inputErr := result.CreateCI()
		if inputErr != nil {
			return nil, inputErr
		}
		if index == variable {
			zero[id], one[id] = AigConst0, AigConst1
		} else {
			zero[id], one[id] = input, input
		}
		if err = budget.checkObjects(result); err != nil {
			return nil, err
		}
	}
	for id, object := range source.objects {
		if err = budget.step(1); err != nil {
			return nil, err
		}
		switch object.Type {
		case AigNone, AigConstant1, AigCI, AigCO:
			continue
		case AigBuffer:
			zero[id], err = abcSensitivityMapLiteral(zero, object.Fanin0)
			if err == nil {
				one[id], err = abcSensitivityMapLiteral(one, object.Fanin0)
			}
		case AigAND:
			var z0, z1, o0, o1 AigLit
			z0, err = abcSensitivityMapLiteral(zero, object.Fanin0)
			if err == nil {
				z1, err = abcSensitivityMapLiteral(zero, object.Fanin1)
			}
			if err == nil {
				o0, err = abcSensitivityMapLiteral(one, object.Fanin0)
			}
			if err == nil {
				o1, err = abcSensitivityMapLiteral(one, object.Fanin1)
			}
			if err == nil {
				zero[id], err = result.And(z0, z1)
			}
			if err == nil {
				one[id], err = result.And(o0, o1)
			}
		default:
			err = ErrUnsupported
		}
		if err != nil {
			return nil, err
		}
		if err = budget.checkObjects(result); err != nil {
			return nil, err
		}
	}
	difference := AigConst0
	for index, outputID := range source.cos {
		if err = budget.step(1); err != nil {
			return nil, err
		}
		output := source.objects[outputID]
		low, lowErr := abcSensitivityMapLiteral(zero, output.Fanin0)
		if lowErr != nil {
			return nil, lowErr
		}
		high, highErr := abcSensitivityMapLiteral(one, output.Fanin0)
		if highErr != nil {
			return nil, highErr
		}
		if index == len(source.cos)-1 {
			difference, err = result.And(difference, low)
			if err == nil {
				difference, err = result.And(difference, high)
			}
		} else {
			var changed AigLit
			changed, err = result.Xor(low, high)
			if err == nil {
				difference, err = result.Or(difference, changed)
			}
		}
		if err != nil {
			return nil, err
		}
		if err = budget.checkObjects(result); err != nil {
			return nil, err
		}
	}
	if _, err = result.CreateCO(difference); err != nil {
		return nil, err
	}
	if err = budget.checkObjects(result); err != nil {
		return nil, err
	}
	result.Cleanup()
	if err = result.Check(); err != nil {
		return nil, err
	}
	return result, nil
}

func abcSensitivityLimits(options ABCSensitivityOptions) SatLimits {
	limits := options.SAT
	satStop := limits.Stop
	limits.Stop = func() bool {
		return options.Stop != nil && options.Stop() || satStop != nil && satStop()
	}
	return limits
}

// ABCSensitivity classifies every input with a complete native SAT query.
// SAT and resource-limited unknown inputs are included in Sensitive, matching
// Abc_NtkSensitivity. SAT models contain source-CI values and are replayed on
// the generated miter before they are returned.
func ABCSensitivity(source *AigManager, options ABCSensitivityOptions) (ABCSensitivityResult, error) {
	options, err := normalizeABCSensitivityOptions(options)
	if err != nil {
		return ABCSensitivityResult{}, err
	}
	if source == nil {
		return ABCSensitivityResult{}, ErrDimension
	}
	if err = source.Check(); err != nil {
		return ABCSensitivityResult{}, err
	}
	if source.RegisterCount() != 0 || source.InputCount() == 0 || source.OutputCount() == 0 {
		return ABCSensitivityResult{}, ErrDimension
	}
	if source.InputCount() > options.MaxInputs || source.ObjectCount() > options.MaxObjects {
		return ABCSensitivityResult{}, ErrLimit
	}
	result := ABCSensitivityResult{Inputs: make([]ABCSensitivityInput, source.InputCount())}
	for variable := 0; variable < source.InputCount(); variable++ {
		if options.Stop != nil && options.Stop() {
			return ABCSensitivityResult{}, ErrCancelled
		}
		miter, miterErr := ABCSensitivityMiter(source, variable, options)
		if miterErr != nil {
			return ABCSensitivityResult{}, miterErr
		}
		relation, relationErr := EncodeAigCNF(miter)
		if relationErr != nil {
			return ABCSensitivityResult{}, relationErr
		}
		if relation.VariableCount() > options.MaxVariables || len(relation.clauses) > options.MaxClauses {
			return ABCSensitivityResult{}, ErrLimit
		}
		literals := 0
		for _, clause := range relation.clauses {
			if literals > options.MaxLiterals-len(clause) {
				return ABCSensitivityResult{}, ErrLimit
			}
			literals += len(clause)
		}
		solver, solverErr := NewSatSolverWithLimit(options.MaxVariables)
		if solverErr != nil {
			return ABCSensitivityResult{}, solverErr
		}
		if ok, loadErr := relation.loadIntoWithStop(solver, 0, options.Stop); loadErr != nil || !ok {
			if loadErr != nil {
				return ABCSensitivityResult{}, loadErr
			}
			return ABCSensitivityResult{}, ErrInvariant
		}
		outputs := relation.OutputLiterals()
		if len(outputs) != 1 {
			return ABCSensitivityResult{}, ErrInvariant
		}
		solved, solveErr := solver.Solve([]SatLit{outputs[0]}, abcSensitivityLimits(options))
		if solveErr != nil {
			return ABCSensitivityResult{}, solveErr
		}
		entry := ABCSensitivityInput{Index: variable, Status: solved.Status, Reason: solved.Reason, Stats: solved.Stats}
		switch solved.Status {
		case SatSatisfiable:
			entry.Model = make([]bool, source.InputCount())
			for index, literal := range relation.InputLiterals() {
				if int(literal.Var()) >= len(solved.Model) {
					return ABCSensitivityResult{}, ErrInvariant
				}
				entry.Model[index] = solved.Model[literal.Var()] != literal.Negative()
			}
			outputs, replayErr := miter.Evaluate(entry.Model)
			if replayErr != nil || len(outputs) != 1 || !outputs[0] {
				return ABCSensitivityResult{}, ErrInvariant
			}
			result.Sensitive = append(result.Sensitive, variable)
			result.SAT++
		case SatUnsatisfiable:
			result.UNSAT++
		case SatUnknown:
			result.Sensitive = append(result.Sensitive, variable)
			result.Unknown++
		default:
			return ABCSensitivityResult{}, ErrInvariant
		}
		result.Inputs[variable] = entry
	}
	result.Sensitive = slices.Clone(result.Sensitive)
	return result, nil
}
