package abc

// SAT-guided SOP collapse from src/sat/bmc/bmcClp.c.

import "slices"

type GiaOutputSensitivity struct {
	Value       bool
	Differences []bool
}

type GiaSensitivityStats struct {
	Samples            int
	PatternCounts      [][2][2]int
	DifferenceCounts   []int
	CoDifferenceCounts [2][][]int
}

// OutputSensitivity evaluates one pattern and every single-CI perturbation.
// It is the owned, width-independent counterpart of Bmc_ComputeSimDiff.
func (manager *GiaManager) OutputSensitivity(output int, pattern []bool, maxEvaluations int, stop func() bool) (GiaOutputSensitivity, error) {
	var result GiaOutputSensitivity
	if manager == nil || output < 0 || output >= len(manager.cos) || len(pattern) != len(manager.cis) || maxEvaluations < 0 {
		return result, ErrDimension
	}
	if maxEvaluations == 0 {
		maxEvaluations = 1 << 20
	}
	if len(pattern)+1 > maxEvaluations {
		return result, ErrLimit
	}
	if stop != nil && stop() {
		return result, ErrCancelled
	}
	values, err := manager.Evaluate(pattern)
	if err != nil {
		return result, err
	}
	result.Value = values[output]
	result.Differences = make([]bool, len(pattern))
	changed := slices.Clone(pattern)
	for variable := range pattern {
		if stop != nil && stop() {
			return GiaOutputSensitivity{}, ErrCancelled
		}
		changed[variable] = !changed[variable]
		values, err = manager.Evaluate(changed)
		changed[variable] = !changed[variable]
		if err != nil {
			return GiaOutputSensitivity{}, err
		}
		result.Differences[variable] = values[output] != result.Value
	}
	return result, nil
}

// AnalyzeOutputSensitivity replaces Bmc_ComputeSimTest's ambient rand stream
// and stdout tables with caller-supplied patterns and owned counters.
func (manager *GiaManager) AnalyzeOutputSensitivity(output int, patterns [][]bool, maxSamples, maxEvaluations int, stop func() bool) (GiaSensitivityStats, error) {
	var result GiaSensitivityStats
	if manager == nil || maxSamples < 0 || len(patterns) == 0 {
		return result, ErrDimension
	}
	if maxSamples == 0 {
		maxSamples = 1 << 20
	}
	if len(patterns) > maxSamples {
		return result, ErrLimit
	}
	inputs := len(manager.cis)
	result.PatternCounts = make([][2][2]int, inputs)
	result.DifferenceCounts = make([]int, inputs)
	for phase := range result.CoDifferenceCounts {
		result.CoDifferenceCounts[phase] = make([][]int, inputs)
		for variable := range result.CoDifferenceCounts[phase] {
			result.CoDifferenceCounts[phase][variable] = make([]int, inputs)
		}
	}
	for _, pattern := range patterns {
		if len(pattern) != inputs {
			return GiaSensitivityStats{}, ErrDimension
		}
		sample, err := manager.OutputSensitivity(output, pattern, maxEvaluations, stop)
		if err != nil {
			return GiaSensitivityStats{}, err
		}
		phase := 0
		if sample.Value {
			phase = 1
		}
		for variable, value := range pattern {
			bit := 0
			if value {
				bit = 1
			}
			result.PatternCounts[variable][phase][bit]++
			if sample.Differences[variable] {
				result.DifferenceCounts[variable]++
				for other := variable + 1; other < inputs; other++ {
					if sample.Differences[other] {
						result.CoDifferenceCounts[phase][variable][other]++
					}
				}
			}
		}
		result.Samples++
	}
	return result, nil
}

type GiaCollapseOptions struct {
	Output                                          int
	Canonical, Reverse                              bool
	MaxCubes, MaxVariables, MaxClauses, MaxLiterals int
	MaxObjects, MaxMappingEntries                   int
	SAT                                             SatLimits
	Stop                                            func() bool
}

type GiaCollapseResult struct {
	Cover            SOPCover
	Status           SatStatus
	Reason           SatStopReason
	Stats            SatStats
	EnumeratedOn     int
	EnumeratedOff    int
	RemovedRedundant int
}

type giaCollapsePhase struct {
	value bool
	enum  *SatSolver
	clean *SatSolver
	cubes []string
	done  bool
}

func normalizeGiaCollapseOptions(options GiaCollapseOptions) (GiaCollapseOptions, error) {
	if options.Output < 0 || options.MaxCubes < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
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
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	if options.MaxCubes < 1 || options.MaxVariables < 1 || options.MaxClauses < 1 || options.MaxLiterals < 1 || options.MaxObjects < 1 || options.MaxMappingEntries < 1 {
		return options, ErrLimit
	}
	if options.SAT.Stop == nil {
		options.SAT.Stop = options.Stop
	}
	return options, nil
}

func giaCollapseStopped(options GiaCollapseOptions) bool {
	return options.Stop != nil && options.Stop()
}

func giaCollapseNewSolver(relation *AigCNF, output SatLit, value bool, options GiaCollapseOptions) (*SatSolver, bool, error) {
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return nil, false, err
	}
	ok, err := relation.loadIntoWithStop(solver, 0, options.Stop)
	if err != nil || !ok {
		if err == nil {
			err = ErrInvariant
		}
		return nil, false, err
	}
	literal := output
	if !value {
		literal = literal.Not()
	}
	ok, err = solver.AddClause([]SatLit{literal})
	return solver, ok, err
}

func giaCollapseSolve(solver *SatSolver, assumptions []SatLit, options GiaCollapseOptions, stats *SatStats) (SatResult, error) {
	if giaCollapseStopped(options) {
		return SatResult{}, ErrCancelled
	}
	answer, err := solver.Solve(assumptions, options.SAT)
	if err == nil {
		addSatStats(stats, answer.Stats)
	}
	return answer, err
}

func giaCollapseOrderedInputs(count int, reverse bool) []int {
	result := make([]int, count)
	for index := range result {
		if reverse {
			result[index] = count - index - 1
		} else {
			result[index] = index
		}
	}
	return result
}

func giaCollapseCanonicalModel(phase *giaCollapsePhase, inputs []SatLit, order []int, options GiaCollapseOptions, stats *SatStats) (SatResult, error) {
	prefix := make([]SatLit, 0, len(order))
	for _, input := range order {
		trial := append(slices.Clone(prefix), inputs[input].Not())
		answer, err := giaCollapseSolve(phase.enum, trial, options, stats)
		if err != nil || answer.Status == SatUnknown {
			return answer, err
		}
		if answer.Status == SatSatisfiable {
			prefix = trial
		} else {
			prefix = append(prefix, inputs[input])
		}
	}
	return giaCollapseSolve(phase.enum, prefix, options, stats)
}

func giaCollapseCubeAssumptions(cube []byte, inputs []SatLit, flip int) []SatLit {
	result := make([]SatLit, 0, len(cube))
	for variable, value := range cube {
		if value == '-' {
			continue
		}
		literal := inputs[variable]
		if value == '0' {
			literal = literal.Not()
		}
		if variable == flip {
			literal = literal.Not()
		}
		result = append(result, literal)
	}
	return result
}

func giaCollapseExpandCube(phase *giaCollapsePhase, cube []byte, inputs []SatLit, order []int, options GiaCollapseOptions, stats *SatStats) (SatStatus, SatStopReason, error) {
	for position := len(order) - 1; position >= 0; position-- {
		variable := order[position]
		if cube[variable] == '-' {
			continue
		}
		if options.Canonical {
			answer, err := giaCollapseSolve(phase.enum, giaCollapseCubeAssumptions(cube, inputs, variable), options, stats)
			if err != nil {
				return SatUnknown, SatCancelled, err
			}
			if answer.Status == SatUnknown {
				return SatUnknown, answer.Reason, nil
			}
			if answer.Status == SatUnsatisfiable {
				continue
			}
		}
		saved := cube[variable]
		cube[variable] = '-'
		answer, err := giaCollapseSolve(phase.clean, giaCollapseCubeAssumptions(cube, inputs, -1), options, stats)
		if err != nil {
			return SatUnknown, SatCancelled, err
		}
		if answer.Status == SatUnknown {
			return SatUnknown, answer.Reason, nil
		}
		if answer.Status == SatSatisfiable {
			cube[variable] = saved
		}
	}
	return SatSatisfiable, SatFinished, nil
}

func giaCollapseStep(phase *giaCollapsePhase, inputs []SatLit, order []int, options GiaCollapseOptions, stats *SatStats) (SatStatus, SatStopReason, error) {
	var answer SatResult
	var err error
	if options.Canonical {
		answer, err = giaCollapseCanonicalModel(phase, inputs, order, options, stats)
	} else {
		answer, err = giaCollapseSolve(phase.enum, nil, options, stats)
	}
	if err != nil {
		return SatUnknown, SatCancelled, err
	}
	if answer.Status != SatSatisfiable {
		if answer.Status == SatUnsatisfiable {
			phase.done = true
		}
		return answer.Status, answer.Reason, nil
	}
	if len(phase.cubes) >= options.MaxCubes {
		return SatUnknown, SatFinished, ErrLimit
	}
	cube := make([]byte, len(inputs))
	for variable, input := range inputs {
		if int(input.Var()) >= len(answer.Model) {
			return SatUnknown, SatFinished, ErrInvariant
		}
		if answer.Model[input.Var()] != input.Negative() {
			cube[variable] = '1'
		} else {
			cube[variable] = '0'
		}
	}
	status, reason, err := giaCollapseExpandCube(phase, cube, inputs, order, options, stats)
	if err != nil || status == SatUnknown {
		return status, reason, err
	}
	phase.cubes = append(phase.cubes, string(cube))
	block := giaCollapseCubeAssumptions(cube, inputs, -1)
	for index := range block {
		block[index] = block[index].Not()
	}
	ok, err := phase.enum.AddClause(block)
	if err != nil {
		return SatUnknown, SatFinished, err
	}
	if !ok {
		phase.done = true
	}
	return SatSatisfiable, SatFinished, nil
}

// CollapseGiaOutputToSOP computes an irredundant SOP for one combinational
// output. It alternates on-set and off-set construction and returns whichever
// finishes first, matching Bmc_CollapseOne. Returned cubes always use original
// CI order; Reverse changes SAT enumeration order only.
func (manager *GiaManager) CollapseGiaOutputToSOP(options GiaCollapseOptions) (GiaCollapseResult, error) {
	var result GiaCollapseResult
	if manager == nil {
		return result, ErrDimension
	}
	checked, err := normalizeGiaCollapseOptions(options)
	if err != nil {
		return result, err
	}
	if err = manager.Check(); err != nil {
		return result, err
	}
	if checked.Output >= len(manager.cos) || len(manager.cis) > checked.MaxVariables || len(manager.objects) > checked.MaxObjects {
		if checked.Output >= len(manager.cos) {
			return result, ErrDimension
		}
		return result, ErrLimit
	}
	if len(manager.objects) > checked.MaxMappingEntries {
		return result, ErrLimit
	}
	if giaCollapseStopped(checked) {
		return result, ErrCancelled
	}
	aig, _, err := manager.ToAig()
	if err != nil {
		return result, err
	}
	relation, err := EncodeAigCNF(aig)
	if err != nil {
		return result, err
	}
	if len(relation.clauses) > checked.MaxClauses {
		return result, ErrLimit
	}
	literals := 0
	for _, clause := range relation.clauses {
		literals += len(clause)
		if literals > checked.MaxLiterals {
			return result, ErrLimit
		}
	}
	inputs := relation.InputLiterals()
	outputs := relation.OutputLiterals()
	output := outputs[checked.Output]
	phases := [2]giaCollapsePhase{{value: true}, {value: false}}
	for index := range phases {
		phases[index].enum, phases[index].done, err = giaCollapseNewSolver(relation, output, phases[index].value, checked)
		phases[index].done = !phases[index].done
		if err != nil {
			return result, err
		}
		phases[index].clean, _, err = giaCollapseNewSolver(relation, output, !phases[index].value, checked)
		if err != nil {
			return result, err
		}
	}
	if phases[0].done && phases[1].done {
		return result, ErrInvariant
	}
	if phases[0].done || phases[1].done {
		value := phases[1].done
		outputCode := byte('0')
		if value {
			outputCode = '1'
		}
		result.Cover = SOPCover{Variables: 0, Cubes: []string{""}, Output: outputCode}
		result.Status, result.Reason = SatSatisfiable, SatFinished
		return result, nil
	}
	order := giaCollapseOrderedInputs(len(inputs), checked.Reverse)
	var winner *giaCollapsePhase
	for winner == nil {
		for index := range phases {
			status, reason, stepErr := giaCollapseStep(&phases[index], inputs, order, checked, &result.Stats)
			if stepErr != nil {
				return GiaCollapseResult{}, stepErr
			}
			if status == SatUnknown {
				result.Status, result.Reason = status, reason
				return result, nil
			}
			if phases[index].done {
				winner = &phases[index]
				break
			}
		}
	}
	result.EnumeratedOn, result.EnumeratedOff = len(phases[0].cubes), len(phases[1].cubes)
	outputCode := byte('0')
	if winner.value {
		outputCode = '1'
	}
	cover := SOPCover{Variables: len(inputs), Cubes: slices.Clone(winner.cubes), Output: outputCode}
	if len(cover.Cubes) == 0 {
		return GiaCollapseResult{}, ErrInvariant
	}
	before := len(cover.Cubes)
	entries, err := checkedSize(0, len(cover.Cubes), max(1, len(inputs)), int(^uint(0)>>1))
	if err != nil {
		return GiaCollapseResult{}, err
	}
	expansion := SOPExpansionOptions{
		MaxVariables: checked.MaxVariables, MaxClauses: checked.MaxClauses,
		MaxLiterals: checked.MaxLiterals, MaxCubes: checked.MaxCubes,
		MaxCubeEntries: entries, SAT: checked.SAT, Stop: checked.Stop,
	}
	cover, irredundantStats, status, reason, err := sopCoverIrredundant(cover, expansion)
	addSatStats(&result.Stats, irredundantStats)
	if err != nil {
		return GiaCollapseResult{}, err
	}
	if status == SatUnknown {
		result.Status, result.Reason = status, reason
		return result, nil
	}
	result.Cover = cover
	result.RemovedRedundant = before - len(cover.Cubes)
	result.Status, result.Reason = SatSatisfiable, SatFinished
	return result, nil
}
