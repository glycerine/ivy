package abc

// Complete fault preparation and campaign workflow from src/sat/bmc/bmcFault.c.

import "slices"

type GiaFaultPreparationOptions struct {
	Parameters  GiaFaultParameters
	Active      []bool
	Patterns    []bool
	WarmUp      bool
	MaxPatterns int
	Transform   GiaFaultTransformOptions
	Formula     FaultFormulaOptions
	CNF         GiaFaultCNFOptions
	Cardinality SatCardinalityOptions
	SAT         SatLimits
	Stop        func() bool
}

type GiaFaultPreparation struct {
	Miter              *GiaManager
	Relation           *GiaFaultCNF
	Solver             *SatSolver
	FunctionalInputs   int
	ParameterVariables []SatVar
	Active             []bool
	Patterns           []bool
	Ready              bool
	Status             SatStatus
	Reason             SatStopReason
}

type GiaFaultNewOptions struct {
	MaxAssignments int
	Transform      GiaFaultTransformOptions
	CNF            GiaFaultCNFOptions
	SAT            SatLimits
	Stop           func() bool
}

type GiaFaultNewResult struct {
	Assignments [][]int
	Iterations  int
	Status      SatStatus
	Reason      SatStopReason
}

type GiaFaultCampaignOptions struct {
	GiaFaultPreparationOptions
	MaxIterations  int
	MaxAssignments int
}

type GiaFaultCampaignResult struct {
	Patterns              []bool
	PatternData           []byte
	Active                []bool
	Generated             int
	Iterations            int
	Status                SatStatus
	Reason                SatStopReason
	IterationLimitReached bool
	UntestableStatus      SatStatus
	UntestableReason      SatStopReason
	Untestable            [][]int
	NewFaults             [][]int
	Statistics            GiaFaultRunStatistics
}

func faultCombinedStop(first, second func() bool) func() bool {
	if first == nil {
		return second
	}
	if second == nil {
		return first
	}
	return func() bool { return first() || second() }
}

func faultPreparationLimits(options *GiaFaultPreparationOptions) error {
	if options.MaxPatterns < 0 || options.Parameters.Timeout < 0 || options.Parameters.CheckIterations < 0 || options.Parameters.CardinalityConstraint < 0 {
		return ErrDimension
	}
	if options.MaxPatterns == 0 {
		options.MaxPatterns = 1 << 20
	}
	options.Transform.Stop = faultCombinedStop(options.Stop, options.Transform.Stop)
	options.Formula.Stop = faultCombinedStop(options.Stop, options.Formula.Stop)
	options.CNF.Stop = faultCombinedStop(options.Stop, options.CNF.Stop)
	options.Cardinality.Stop = faultCombinedStop(options.Stop, options.Cardinality.Stop)
	options.SAT.Stop = faultCombinedStop(options.Stop, options.SAT.Stop)
	return nil
}

func faultActiveMap(count int, supplied []bool) ([]bool, error) {
	if count < 0 || supplied != nil && len(supplied) != count {
		return nil, ErrDimension
	}
	if supplied == nil {
		result := make([]bool, count)
		for i := range result {
			result[i] = true
		}
		return result, nil
	}
	return slices.Clone(supplied), nil
}

func buildGiaFaultSides(design, golden *GiaManager, options GiaFaultPreparationOptions) (left, right *GiaManager, functional int, active []bool, err error) {
	if design == nil || golden == nil || options.Parameters.Algorithm < 0 || options.Parameters.Algorithm > 4 {
		return nil, nil, 0, nil, ErrDimension
	}
	if err = design.Check(); err != nil {
		return nil, nil, 0, nil, err
	}
	if err = golden.Check(); err != nil {
		return nil, nil, 0, nil, err
	}
	var parameterCount int
	switch options.Parameters.Algorithm {
	case 0:
		var unfolded *GiaFaultFormulaUnfold
		unfolded, err = UnfoldGiaFaultFormula(design, options.Parameters.Formula, options.Parameters.FlopsOnly, options.Formula)
		if err == nil {
			right = unfolded.Graph
			for _, row := range unfolded.ParameterInputs {
				parameterCount += len(row)
			}
		}
		functional = design.InputCount()
	case 1:
		if design.RegisterCount() == 0 || golden.RegisterCount() == 0 {
			return nil, nil, 0, nil, ErrDimension
		}
		var nominal, faulty *GiaTemporalFaultUnfold
		nominal, err = UnfoldGiaTemporalFault(golden, false, options.Parameters.FlopsOnly, options.Transform)
		if err == nil {
			faulty, err = UnfoldGiaTemporalFault(design, true, options.Parameters.FlopsOnly, options.Transform)
		}
		if err == nil {
			left, right = nominal.Graph, faulty.Graph
			parameterCount = len(faulty.Controls)
		}
		functional = design.RegisterCount() + 2*design.PrimaryInputCount()
	case 2:
		parameterCount = 2 * design.AndCount()
		active, err = faultActiveMap(parameterCount, options.Active)
		if err == nil {
			var unfolded *GiaFaultTransform
			unfolded, err = UnfoldGiaStuckAt(design, active, options.Transform)
			if err == nil {
				right = unfolded.Graph
			}
		}
		functional = design.InputCount()
	case 3:
		parameterCount = design.AndCount()
		active, err = faultActiveMap(parameterCount, options.Active)
		if err == nil {
			var unfolded *GiaFaultTransform
			unfolded, err = UnfoldGiaFlip(design, active, options.Transform)
			if err == nil {
				right = unfolded.Graph
			}
		}
		functional = design.InputCount()
	case 4:
		parameterCount = 4 * design.AndCount()
		active, err = faultActiveMap(parameterCount, options.Active)
		if err == nil {
			var unfolded *GiaFaultTransform
			unfolded, err = UnfoldGiaFourOutputFunction(design, active, options.Transform)
			if err == nil {
				right = unfolded.Graph
			}
		}
		functional = design.InputCount()
	}
	if err != nil {
		return nil, nil, 0, nil, err
	}
	if active == nil {
		active, err = faultActiveMap(parameterCount, options.Active)
		if err != nil {
			return nil, nil, 0, nil, err
		}
	}
	if left == nil {
		extra := right.InputCount() - golden.InputCount()
		if extra < 0 {
			return nil, nil, 0, nil, ErrDimension
		}
		left, _, err = DuplicateGiaWithPrimaryInputs(golden, extra, GiaCofactorOptions{
			MaxObjects: options.Transform.MaxObjects, MaxMappingEntries: options.Transform.MaxMappingEntries,
			Stop: options.Transform.Stop,
		})
	}
	if err != nil || left.InputCount() != right.InputCount() || left.OutputCount() != right.OutputCount() || right.InputCount() != functional+parameterCount {
		if err != nil {
			return nil, nil, 0, nil, err
		}
		return nil, nil, 0, nil, ErrDimension
	}
	return left, right, functional, active, nil
}

func faultVariables(literals []SatLit) ([]SatVar, error) {
	result := make([]SatVar, len(literals))
	for i, literal := range literals {
		if !literal.Valid() || literal.Negative() {
			return nil, ErrInvalidLiteral
		}
		result[i] = literal.Var()
	}
	return result, nil
}

func addFaultBaseConstraints(solver *SatSolver, relation *GiaFaultCNF, functional int, active []bool, parameters GiaFaultParameters, cardinality SatCardinalityOptions) ([]SatVar, error) {
	if solver == nil || relation == nil || functional < 0 {
		return nil, ErrDimension
	}
	inputs, outputs := relation.InputLiterals(), relation.OutputLiterals()
	if functional > len(inputs) || len(outputs) == 0 || len(active) != len(inputs)-functional {
		return nil, ErrDimension
	}
	variables, err := faultVariables(inputs[functional:])
	if err != nil {
		return nil, err
	}
	for i, enabled := range active {
		if enabled {
			continue
		}
		ok, clauseErr := solver.AddClause([]SatLit{satCardinalityLit(variables[i], true)})
		if clauseErr != nil {
			return nil, clauseErr
		}
		if !ok {
			return variables, nil
		}
	}
	if parameters.Basic && len(variables) > 0 {
		_, err = AddSatCardinalityConstraint(solver, variables, 1, false, SatCardinalityAtMostOneTree, cardinality)
	} else if parameters.CardinalityConstraint != 0 {
		if parameters.CardinalityConstraint < 1 || parameters.CardinalityConstraint >= len(variables) {
			return nil, ErrDimension
		}
		_, err = AddSatCardinalityConstraint(solver, variables, parameters.CardinalityConstraint, !parameters.NonStrict, SatCardinalityGeneral, cardinality)
	}
	return variables, err
}

func PrepareGiaFaultCampaign(design, golden *GiaManager, options GiaFaultPreparationOptions) (*GiaFaultPreparation, error) {
	if err := faultPreparationLimits(&options); err != nil {
		return nil, err
	}
	left, right, functional, active, err := buildGiaFaultSides(design, golden, options)
	if err != nil {
		return nil, err
	}
	miter, err := BuildGiaDesignMiter(left, right, GiaMiterOptions{
		GiaCofactorOptions: GiaCofactorOptions{MaxObjects: options.Transform.MaxObjects, MaxMappingEntries: options.Transform.MaxMappingEntries, Stop: options.Transform.Stop},
		Mode:               GiaMiterXor,
	})
	if err != nil {
		return nil, err
	}
	relation, err := DeriveGiaFaultCNF(miter.Graph, options.CNF)
	if err != nil {
		return nil, err
	}
	solver := NewSatSolver()
	ok, err := relation.LoadInto(solver)
	if err != nil {
		return nil, err
	}
	result := &GiaFaultPreparation{Miter: miter.Graph, Relation: relation, Solver: solver, FunctionalInputs: functional, Active: active, Status: SatUnknown}
	if !ok {
		result.Status = SatUnsatisfiable
		return result, nil
	}
	outputs := relation.OutputLiterals()
	if len(outputs) == 0 {
		return nil, ErrDimension
	}
	if ok, err = solver.AddClause(outputs); err != nil {
		return nil, err
	} else if !ok {
		result.Status = SatUnsatisfiable
		return result, nil
	}
	result.ParameterVariables, err = addFaultBaseConstraints(solver, relation, functional, active, options.Parameters, options.Cardinality)
	if err != nil {
		return nil, err
	}
	patterns := slices.Clone(options.Patterns)
	if functional <= 0 || len(patterns)%functional != 0 || len(patterns)/functional > options.MaxPatterns {
		return nil, ErrDimension
	}
	generatedStart := len(patterns) == 0 && options.Parameters.StartPatterns
	if generatedStart {
		patterns = make([]bool, 2*functional)
		for i := functional; i < 2*functional; i++ {
			patterns[i] = true
		}
	}
	for row := 0; row < len(patterns)/functional; row++ {
		if options.WarmUp || generatedStart {
			answer, solveErr := solver.Solve(nil, options.SAT)
			if solveErr != nil {
				return nil, solveErr
			}
			if answer.Status != SatSatisfiable {
				result.Status, result.Reason = answer.Status, answer.Reason
				result.Patterns = slices.Clone(patterns[:row*functional])
				return result, nil
			}
		}
		pattern := patterns[row*functional : (row+1)*functional]
		added, addErr := AddFaultPatternInstance(solver, miter.Graph, pattern, functional, false, result.ParameterVariables, AddFaultPatternOptions{CNF: options.CNF, Stop: options.Stop})
		if addErr != nil {
			return nil, addErr
		}
		if added.Inconsistent {
			result.Status = SatUnsatisfiable
			return result, nil
		}
		result.Patterns = append(result.Patterns, pattern...)
	}
	result.Ready = true
	return result, nil
}

func FindGiaFaultNewAssignments(miter *GiaManager, functional int, prior, newest []bool, options GiaFaultNewOptions) (GiaFaultNewResult, error) {
	result := GiaFaultNewResult{Status: SatUnknown}
	if miter == nil || functional <= 0 || len(newest) != functional || len(prior)%functional != 0 || options.MaxAssignments < 0 {
		return result, ErrDimension
	}
	if options.MaxAssignments == 0 {
		options.MaxAssignments = 10000
	}
	options.Transform.Stop = faultCombinedStop(options.Stop, options.Transform.Stop)
	options.CNF.Stop = faultCombinedStop(options.Stop, options.CNF.Stop)
	options.SAT.Stop = faultCombinedStop(options.Stop, options.SAT.Stop)
	cofactor, err := CofactorGiaFaultInputs(miter, newest, options.Transform)
	if err != nil {
		return result, err
	}
	relation, err := DeriveGiaFaultCNF(cofactor, options.CNF)
	if err != nil {
		return result, err
	}
	solver := NewSatSolver()
	ok, err := relation.LoadInto(solver)
	if err != nil {
		return result, err
	}
	if !ok {
		result.Status = SatUnsatisfiable
		return result, nil
	}
	outputs := relation.OutputLiterals()
	inputs := relation.InputLiterals()
	if functional > len(inputs) || len(outputs) == 0 {
		return result, ErrDimension
	}
	if ok, err = solver.AddClause(outputs); err != nil {
		return result, err
	} else if !ok {
		result.Status = SatUnsatisfiable
		return result, nil
	}
	parameters, err := faultVariables(inputs[functional:])
	if err != nil {
		return result, err
	}
	for row := 0; row < len(prior)/functional; row++ {
		pattern := prior[row*functional : (row+1)*functional]
		added, addErr := AddFaultPatternInstance(solver, miter, pattern, functional, false, parameters, AddFaultPatternOptions{CNF: options.CNF, Stop: options.Stop})
		if addErr != nil {
			return result, addErr
		}
		if added.Inconsistent {
			result.Status = SatUnsatisfiable
			return result, nil
		}
	}
	enumeration, err := EnumerateFaultAssignments(solver, parameters, FaultAssignmentEnumerationOptions{MaxIterations: options.MaxAssignments, SAT: options.SAT, Stop: options.Stop})
	if err != nil {
		return result, err
	}
	result.Assignments, result.Iterations = enumeration.Assignments, enumeration.Iterations
	result.Status, result.Reason = enumeration.Status, enumeration.Reason
	return result, nil
}

func buildGiaUntestableSolver(prepared *GiaFaultPreparation, parameters GiaFaultParameters, patterns []bool, options GiaFaultPreparationOptions) (*SatSolver, error) {
	if prepared == nil || prepared.Relation == nil || prepared.Miter == nil || len(patterns)%prepared.FunctionalInputs != 0 {
		return nil, ErrDimension
	}
	solver := NewSatSolver()
	ok, err := prepared.Relation.LoadInto(solver)
	if err != nil || !ok {
		if err != nil {
			return nil, err
		}
		return solver, nil
	}
	variables, err := faultVariables(prepared.Relation.InputLiterals()[prepared.FunctionalInputs:])
	if err != nil {
		return nil, err
	}
	for i, enabled := range prepared.Active {
		if !enabled {
			if ok, err = solver.AddClause([]SatLit{satCardinalityLit(variables[i], true)}); err != nil || !ok {
				return solver, err
			}
		}
	}
	if len(variables) == 0 {
		return nil, ErrDimension
	}
	positive := make([]SatLit, len(variables))
	for i, variable := range variables {
		positive[i] = satCardinalityLit(variable, false)
	}
	if ok, err = solver.AddClause(positive); err != nil || !ok {
		return solver, err
	}
	if parameters.Basic {
		if _, err = AddSatCardinalityConstraint(solver, variables, 1, false, SatCardinalityAtMostOneTree, options.Cardinality); err != nil {
			return nil, err
		}
	}
	for _, output := range prepared.Relation.OutputLiterals() {
		if ok, err = solver.AddClause([]SatLit{output.Not()}); err != nil || !ok {
			return solver, err
		}
	}
	for row := 0; row < len(patterns)/prepared.FunctionalInputs; row++ {
		pattern := patterns[row*prepared.FunctionalInputs : (row+1)*prepared.FunctionalInputs]
		added, addErr := AddFaultPatternInstance(solver, prepared.Miter, pattern, prepared.FunctionalInputs, false, variables, AddFaultPatternOptions{CNF: options.CNF, Stop: options.Stop})
		if addErr != nil {
			return nil, addErr
		}
		if added.Inconsistent {
			break
		}
	}
	return solver, nil
}

func RunGiaFaultCampaign(design, golden *GiaManager, options GiaFaultCampaignOptions) (GiaFaultCampaignResult, error) {
	result := GiaFaultCampaignResult{Status: SatUnknown, UntestableStatus: SatUnknown}
	if options.MaxIterations < 0 || options.MaxAssignments < 0 {
		return result, ErrDimension
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 1000000
	}
	if options.MaxAssignments == 0 {
		options.MaxAssignments = 10000
	}
	if options.Parameters.FileName != "" {
		if options.Patterns != nil {
			return result, ErrDimension
		}
		patterns, readErr := ReadFaultPatterns(options.Parameters.FileName, FaultPatternOptions{Stop: options.Stop})
		if readErr != nil {
			return result, readErr
		}
		options.Patterns = patterns
		options.Parameters.FileName = ""
	}
	options.WarmUp = true
	prepared, err := PrepareGiaFaultCampaign(design, golden, options.GiaFaultPreparationOptions)
	if err != nil {
		return result, err
	}
	result.Patterns, result.Active = slices.Clone(prepared.Patterns), slices.Clone(prepared.Active)
	if !prepared.Ready {
		result.Status, result.Reason = prepared.Status, prepared.Reason
		return result, nil
	}
	initialRows := len(result.Patterns) / prepared.FunctionalInputs
	for result.Iterations = initialRows; result.Iterations < options.MaxIterations; result.Iterations++ {
		if options.Stop != nil && options.Stop() {
			return GiaFaultCampaignResult{}, ErrCancelled
		}
		if options.Parameters.CheckIterations > 0 && result.Iterations > 0 && result.Iterations%options.Parameters.CheckIterations == 0 {
			analysis, analysisErr := AnalyzeFaultParameters(prepared.Solver, prepared.ParameterVariables, result.Active, FaultParameterAnalysisOptions{SAT: options.SAT, Stop: options.Stop})
			if analysisErr != nil {
				return GiaFaultCampaignResult{}, analysisErr
			}
			result.Active = analysis.Active
			rebuild := options.GiaFaultPreparationOptions
			rebuild.Active, rebuild.Patterns, rebuild.WarmUp = result.Active, result.Patterns, false
			rebuild.Parameters.StartPatterns = false
			prepared, err = PrepareGiaFaultCampaign(design, golden, rebuild)
			if err != nil {
				return GiaFaultCampaignResult{}, err
			}
			if !prepared.Ready {
				result.Status, result.Reason = prepared.Status, prepared.Reason
				break
			}
		}
		answer, solveErr := prepared.Solver.Solve(nil, options.SAT)
		if solveErr != nil {
			return GiaFaultCampaignResult{}, solveErr
		}
		result.Status, result.Reason = answer.Status, answer.Reason
		if answer.Status != SatSatisfiable {
			break
		}
		inputs := prepared.Relation.InputLiterals()
		pattern := make([]bool, prepared.FunctionalInputs)
		for i, literal := range inputs[:prepared.FunctionalInputs] {
			pattern[i] = answer.Model[literal.Var()] != literal.Negative()
		}
		if options.Parameters.DumpNewFaults {
			newResult, newErr := FindGiaFaultNewAssignments(prepared.Miter, prepared.FunctionalInputs, result.Patterns, pattern, GiaFaultNewOptions{
				MaxAssignments: options.MaxAssignments, Transform: options.Transform, CNF: options.CNF, SAT: options.SAT, Stop: options.Stop,
			})
			if newErr != nil {
				return GiaFaultCampaignResult{}, newErr
			}
			result.NewFaults = newResult.Assignments
			break
		}
		result.Patterns = append(result.Patterns, pattern...)
		result.Generated++
		added, addErr := AddFaultPatternInstance(prepared.Solver, prepared.Miter, pattern, prepared.FunctionalInputs, false, prepared.ParameterVariables, AddFaultPatternOptions{CNF: options.CNF, Stop: options.Stop})
		if addErr != nil {
			return GiaFaultCampaignResult{}, addErr
		}
		if added.Inconsistent {
			result.Status, result.Reason = SatUnsatisfiable, SatFinished
			result.Iterations++
			break
		}
	}
	if result.Iterations >= options.MaxIterations && result.Status == SatSatisfiable {
		result.IterationLimitReached = true
		result.Status = SatUnknown
	}
	checkUntestable := options.Parameters.CheckUntestable || options.Parameters.DumpUntestable || design != golden
	if len(result.Patterns) > 0 && checkUntestable {
		untestable, buildErr := buildGiaUntestableSolver(prepared, options.Parameters, result.Patterns, options.GiaFaultPreparationOptions)
		if buildErr != nil {
			return GiaFaultCampaignResult{}, buildErr
		}
		answer, solveErr := untestable.Solve(nil, options.SAT)
		if solveErr != nil {
			return GiaFaultCampaignResult{}, solveErr
		}
		result.UntestableStatus, result.UntestableReason = answer.Status, answer.Reason
		if options.Parameters.DumpUntestable && answer.Status == SatSatisfiable {
			enumeration, enumErr := EnumerateFaultAssignments(untestable, prepared.ParameterVariables, FaultAssignmentEnumerationOptions{MaxIterations: options.MaxAssignments, SAT: options.SAT, Stop: options.Stop})
			if enumErr != nil {
				return GiaFaultCampaignResult{}, enumErr
			}
			result.Untestable = enumeration.Assignments
		}
	}
	if options.Parameters.Dump {
		if options.Parameters.DumpDelay && options.Parameters.Algorithm == 1 {
			result.PatternData, err = MarshalFaultDelayPatterns(design, result.Patterns, FaultPatternOptions{Stop: options.Stop})
		} else {
			result.PatternData, err = MarshalFaultPatterns(result.Patterns, prepared.FunctionalInputs, FaultPatternOptions{Stop: options.Stop})
		}
		if err != nil {
			return GiaFaultCampaignResult{}, err
		}
	}
	result.Statistics, err = FaultRunStatistics(prepared.Miter, prepared.Solver, len(result.Patterns)/prepared.FunctionalInputs)
	return result, err
}
