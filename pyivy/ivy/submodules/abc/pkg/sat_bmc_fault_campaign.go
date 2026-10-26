package abc

// Reusable SAT campaign operations from src/sat/bmc/bmcFault.c.

import "slices"

type FaultParameterAnalysisOptions struct {
	SAT  SatLimits
	Stop func() bool
}

type FaultParameterAnalysis struct {
	Active                []bool
	Runs, Unsatisfiable   int
	Unknown, ModelCovered int
}

func AnalyzeFaultParameters(solver *SatSolver, parameters []SatVar, active []bool, options FaultParameterAnalysisOptions) (FaultParameterAnalysis, error) {
	result := FaultParameterAnalysis{}
	if solver == nil || len(parameters) != len(active) {
		return result, ErrDimension
	}
	seen := make([]bool, solver.VariableCount())
	for _, variable := range parameters {
		if uint64(variable) >= uint64(len(seen)) || seen[variable] {
			return result, ErrDimension
		}
		seen[variable] = true
	}
	result.Active = slices.Clone(active)
	work := slices.Clone(active)
	for index, variable := range parameters {
		if !work[index] {
			continue
		}
		if options.Stop != nil && options.Stop() {
			return FaultParameterAnalysis{}, ErrCancelled
		}
		result.Runs++
		query, err := solver.Solve([]SatLit{satCardinalityLit(variable, false)}, options.SAT)
		if err != nil {
			return FaultParameterAnalysis{}, err
		}
		switch query.Status {
		case SatUnknown:
			result.Unknown++
		case SatUnsatisfiable:
			result.Unsatisfiable++
			result.Active[index] = false
		case SatSatisfiable:
			for i, candidate := range parameters {
				if work[i] && query.Model[candidate] {
					work[i] = false
					result.ModelCovered++
				}
			}
		}
	}
	return result, nil
}

type FaultAssignmentEnumerationOptions struct {
	MaxIterations int
	SAT           SatLimits
	Stop          func() bool
}

type FaultAssignmentEnumeration struct {
	Assignments [][]int
	Iterations  int
	Status      SatStatus
	Reason      SatStopReason
}

func EnumerateFaultAssignments(solver *SatSolver, parameters []SatVar, options FaultAssignmentEnumerationOptions) (FaultAssignmentEnumeration, error) {
	result := FaultAssignmentEnumeration{Status: SatUnknown}
	if solver == nil || options.MaxIterations < 0 {
		return result, ErrDimension
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 10000
	}
	seen := make([]bool, solver.VariableCount())
	for _, variable := range parameters {
		if uint64(variable) >= uint64(len(seen)) || seen[variable] {
			return result, ErrDimension
		}
		seen[variable] = true
	}
	trial := solver.cloneState()
	for result.Iterations < options.MaxIterations {
		if options.Stop != nil && options.Stop() {
			return FaultAssignmentEnumeration{}, ErrCancelled
		}
		query, err := trial.Solve(nil, options.SAT)
		if err != nil {
			return FaultAssignmentEnumeration{}, err
		}
		result.Status, result.Reason = query.Status, query.Reason
		if query.Status != SatSatisfiable {
			break
		}
		block := make([]SatLit, len(parameters))
		var enabled []int
		for i, variable := range parameters {
			value := query.Model[variable]
			block[i] = satCardinalityLit(variable, value)
			if value {
				enabled = append(enabled, i)
			}
		}
		if len(enabled) > 0 {
			result.Assignments = append(result.Assignments, enabled)
		}
		result.Iterations++
		ok, err := trial.AddClause(block)
		if err != nil {
			return FaultAssignmentEnumeration{}, err
		}
		if !ok {
			result.Status = SatUnsatisfiable
			break
		}
	}
	return result, nil
}

type AddFaultPatternOptions struct {
	CNF  GiaFaultCNFOptions
	Stop func() bool
}

type AddFaultPatternResult struct {
	FirstVariable, Variables, Clauses int
	Inconsistent                      bool
	InputLiterals, OutputLiterals     []SatLit
}

// AddFaultPatternInstance adds one cofactored miter relation atomically.
// detect=true requires at least one PO; false requires every PO to be false.
// sharedParameters connects the non-functional PIs to an existing base copy.
func AddFaultPatternInstance(solver *SatSolver, miter *GiaManager, pattern []bool, functionalInputs int, detect bool, sharedParameters []SatVar, options AddFaultPatternOptions) (AddFaultPatternResult, error) {
	result := AddFaultPatternResult{}
	if solver == nil || miter == nil || functionalInputs < 0 || functionalInputs > miter.PrimaryInputCount() || len(pattern) != functionalInputs || len(sharedParameters) != miter.PrimaryInputCount()-functionalInputs {
		return result, ErrDimension
	}
	seen := make([]bool, solver.VariableCount())
	for _, variable := range sharedParameters {
		if uint64(variable) >= uint64(len(seen)) || seen[variable] {
			return result, ErrDimension
		}
		seen[variable] = true
	}
	if options.Stop != nil && options.Stop() {
		return result, ErrCancelled
	}
	cofactor, err := CofactorGiaFaultInputs(miter, pattern, GiaFaultTransformOptions{
		MaxObjects: options.CNF.MaxObjects, MaxMappingEntries: options.CNF.MaxObjects, Stop: options.Stop,
	})
	if err != nil {
		return result, err
	}
	relation, err := DeriveGiaFaultCNF(cofactor, options.CNF)
	if err != nil {
		return result, err
	}
	base := solver.VariableCount()
	lifted, err := relation.Lift(base, options.CNF)
	if err != nil {
		return result, err
	}
	trial := solver.cloneState()
	trial.bookmark = solver.bookmark
	ok, err := lifted.LoadInto(&trial)
	if err != nil {
		return result, err
	}
	inconsistent := !ok
	outputs := lifted.OutputLiterals()[:cofactor.PrimaryOutputCount()]
	if len(outputs) == 0 {
		return result, ErrDimension
	}
	add := func(clause []SatLit) error {
		if inconsistent {
			return nil
		}
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		accepted, err := trial.AddClause(clause)
		if err != nil {
			return err
		}
		inconsistent = !accepted
		return nil
	}
	if detect {
		if err = add(outputs); err != nil {
			return result, err
		}
	} else {
		for _, output := range outputs {
			if err = add([]SatLit{output.Not()}); err != nil {
				return result, err
			}
		}
	}
	inputs := lifted.InputLiterals()
	for i, variable := range sharedParameters {
		instance := inputs[functionalInputs+i]
		shared := satCardinalityLit(variable, false)
		if err = add([]SatLit{shared.Not(), instance}); err != nil {
			return result, err
		}
		if err = add([]SatLit{shared, instance.Not()}); err != nil {
			return result, err
		}
	}
	*solver = trial
	result = AddFaultPatternResult{
		FirstVariable: base, Variables: trial.VariableCount() - base,
		Clauses: len(lifted.clauses), Inconsistent: inconsistent,
		InputLiterals: inputs, OutputLiterals: slices.Clone(outputs),
	}
	return result, nil
}
