package abc

// Logic-sharing divisor discovery from src/sat/bmc/bmcFx.c.

import (
	"slices"
	"sort"
)

type FXDivisorType uint8

const (
	FXDivisorConstant FXDivisorType = iota
	FXDivisorAND
	FXDivisorXOR
	FXDivisorMux
	FXDivisorNone
)

type FXDivisor struct {
	Type   FXDivisorType
	Inputs [3]SatLit // cover-column literals; InvalidSatLit marks an unused input
	Cost   int
}

type FXDivisorOptions struct {
	MaxDivisors, MaxCubes, MaxCubeEntries, MaxPairs int
	Stop                                            func() bool
}

type fxDivisorKey struct {
	type_   FXDivisorType
	a, b, c SatLit
}

func normalizeFXDivisorOptions(options FXDivisorOptions) (FXDivisorOptions, error) {
	if options.MaxDivisors < 0 || options.MaxCubes < 0 || options.MaxCubeEntries < 0 || options.MaxPairs < 0 {
		return options, ErrDimension
	}
	if options.MaxDivisors == 0 {
		options.MaxDivisors = 1 << 20
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxCubeEntries == 0 {
		options.MaxCubeEntries = 16 << 20
	}
	if options.MaxPairs == 0 {
		options.MaxPairs = 16 << 20
	}
	if options.MaxDivisors < 1 || options.MaxCubes < 1 || options.MaxCubeEntries < 1 || options.MaxPairs < 1 {
		return options, ErrLimit
	}
	return options, nil
}

func fxCubeLiterals(cube string) ([]SatLit, error) {
	result := make([]SatLit, 0, len(cube))
	for variable, value := range []byte(cube) {
		if value == '-' {
			continue
		}
		if value != '0' && value != '1' {
			return nil, ErrDimension
		}
		literal, err := MakeSatLit(SatVar(variable), value == '0')
		if err != nil {
			return nil, err
		}
		result = append(result, literal)
	}
	return result, nil
}

func fxAddResidual(target *[2]SatLit, literal SatLit) bool {
	if target[0] == InvalidSatLit {
		target[0] = literal
		return true
	}
	if target[1] == InvalidSatLit {
		target[1] = literal
		return true
	}
	return false
}

func fxCubeDifference(left, right []SatLit) (shared int, leftOnly, rightOnly [2]SatLit, ok bool) {
	leftOnly = [2]SatLit{InvalidSatLit, InvalidSatLit}
	rightOnly = [2]SatLit{InvalidSatLit, InvalidSatLit}
	for l, r := 0, 0; l < len(left) || r < len(right); {
		switch {
		case l < len(left) && r < len(right) && left[l] == right[r]:
			shared++
			l++
			r++
		case r == len(right) || l < len(left) && left[l] < right[r]:
			if !fxAddResidual(&leftOnly, left[l]) {
				return 0, leftOnly, rightOnly, false
			}
			l++
		default:
			if !fxAddResidual(&rightOnly, right[r]) {
				return 0, leftOnly, rightOnly, false
			}
			r++
		}
	}
	return shared, leftOnly, rightOnly, true
}

func fxDivisorOR(left, right SatLit) (FXDivisorType, [3]SatLit) {
	inputs := [3]SatLit{InvalidSatLit, InvalidSatLit, InvalidSatLit}
	if left.Var() == right.Var() {
		return FXDivisorConstant, inputs
	}
	if left > right {
		left, right = right, left
	}
	inputs[0], inputs[1] = left.Not(), right.Not()
	return FXDivisorAND, inputs
}

func fxDivisorXOR(left, right SatLit) (FXDivisorType, [3]SatLit) {
	return FXDivisorXOR, [3]SatLit{left &^ 1, right &^ 1, InvalidSatLit}
}

func fxDivisorMux(control, invertedControl, whenTrue, whenFalse SatLit) (FXDivisorType, [3]SatLit) {
	if control.Negative() {
		control = control &^ 1
		whenTrue, whenFalse = whenFalse, whenTrue
	}
	if whenTrue.Negative() {
		whenTrue, whenFalse = whenTrue.Not(), whenFalse.Not()
	}
	_ = invertedControl
	return FXDivisorMux, [3]SatLit{control, whenTrue, whenFalse}
}

func fxClassifyDivisor(left, right [2]SatLit) (FXDivisorType, [3]SatLit) {
	unused := [3]SatLit{InvalidSatLit, InvalidSatLit, InvalidSatLit}
	if left[0] == InvalidSatLit || right[0] == InvalidSatLit {
		return FXDivisorConstant, unused
	}
	if left[1] == InvalidSatLit && right[1] == InvalidSatLit {
		return fxDivisorOR(left[0], right[0])
	}
	if left[1] == InvalidSatLit || right[1] == InvalidSatLit {
		if left[1] == InvalidSatLit {
			left, right = right, left
		}
		if right[0].Var() == left[0].Var() {
			return fxDivisorOR(right[0], left[1])
		}
		if right[0].Var() == left[1].Var() {
			return fxDivisorOR(right[0], left[0])
		}
		return FXDivisorNone, unused
	}
	if left[0].Var() == right[0].Var() && left[1].Var() == right[1].Var() {
		return fxDivisorXOR(left[0], right[1])
	}
	if left[0].Var() == right[0].Var() {
		return fxDivisorMux(left[0], right[0], left[1], right[1])
	}
	if left[0].Var() == right[1].Var() {
		return fxDivisorMux(left[0], right[1], left[1], right[0])
	}
	if left[1].Var() == right[0].Var() {
		return fxDivisorMux(left[1], right[0], left[0], right[1])
	}
	if left[1].Var() == right[1].Var() {
		return fxDivisorMux(left[1], right[1], left[0], right[0])
	}
	return FXDivisorNone, unused
}

// FindFXDivisors ranks reusable two-literal ANDs and cube-pair AND/XOR/mux
// divisors. Entries retain first-discovery order for equal costs; the lookup
// map is never iterated.
func FindFXDivisors(cover SOPCover, options FXDivisorOptions) ([]FXDivisor, error) {
	checked, err := normalizeFXDivisorOptions(options)
	if err != nil {
		return nil, err
	}
	if cover.Variables < 0 || len(cover.Cubes) == 0 || len(cover.Cubes) > checked.MaxCubes {
		if len(cover.Cubes) > checked.MaxCubes {
			return nil, ErrLimit
		}
		return nil, ErrDimension
	}
	if _, err = checkedSize(0, len(cover.Cubes), cover.Variables, checked.MaxCubeEntries); err != nil {
		return nil, err
	}
	cubes := make([][]SatLit, len(cover.Cubes))
	for index, cube := range cover.Cubes {
		if len(cube) != cover.Variables {
			return nil, ErrDimension
		}
		cubes[index], err = fxCubeLiterals(cube)
		if err != nil {
			return nil, err
		}
	}
	entries := make([]FXDivisor, 0)
	lookup := make(map[fxDivisorKey]int)
	add := func(type_ FXDivisorType, inputs [3]SatLit, cost int) error {
		if type_ < FXDivisorAND || type_ > FXDivisorMux || cost <= 0 {
			return nil
		}
		key := fxDivisorKey{type_: type_, a: inputs[0], b: inputs[1], c: inputs[2]}
		if index, found := lookup[key]; found {
			entries[index].Cost += cost
			return nil
		}
		if len(entries) >= checked.MaxPairs {
			return ErrLimit
		}
		lookup[key] = len(entries)
		entries = append(entries, FXDivisor{Type: type_, Inputs: inputs, Cost: cost})
		return nil
	}
	pairs := 0
	for left, cube := range cubes {
		if checked.Stop != nil && checked.Stop() {
			return nil, ErrCancelled
		}
		for first := 0; first < len(cube); first++ {
			for second := first + 1; second < len(cube); second++ {
				pairs++
				if pairs > checked.MaxPairs {
					return nil, ErrLimit
				}
				inputs := [3]SatLit{cube[first], cube[second], InvalidSatLit}
				if err = add(FXDivisorAND, inputs, 1); err != nil {
					return nil, err
				}
			}
		}
		for right := left + 1; right < len(cubes); right++ {
			pairs++
			if pairs > checked.MaxPairs {
				return nil, ErrLimit
			}
			shared, leftOnly, rightOnly, ok := fxCubeDifference(cube, cubes[right])
			if !ok {
				continue
			}
			type_, inputs := fxClassifyDivisor(leftOnly, rightOnly)
			if err = add(type_, inputs, shared); err != nil {
				return nil, err
			}
		}
	}
	sort.SliceStable(entries, func(left, right int) bool { return entries[left].Cost > entries[right].Cost })
	count := 0
	for count < len(entries) && count < checked.MaxDivisors && entries[count].Cost > 1 {
		count++
	}
	return slices.Clone(entries[:count]), nil
}

type FXExtractionOptions struct {
	Output, Iterations, DivisorsPerIteration          int
	MaxCandidates, MaxCubes, MaxCubeEntries, MaxPairs int
	MaxObjects, MaxMappingEntries                     int
	MaxVariables, MaxClauses, MaxLiterals             int
	SAT                                               SatLimits
	Stop                                              func() bool
}

type FXCoverResult struct {
	Cover                     SOPCover
	Status                    SatStatus
	Reason                    SatStopReason
	Stats                     SatStats
	RemovedLiterals, Literals int
}

type FXIteration struct {
	Cover      SOPCover
	Added      []FXDivisor
	Candidates int
	Stats      SatStats
}

type FXExtractionResult struct {
	Status     SatStatus
	Reason     SatStopReason
	BaseInputs int
	Divisors   []FXDivisor // append order; inputs refer earlier candidate ordinals
	Iterations []FXIteration
	Stats      SatStats
}

type FXOutputAnalysis struct {
	Output        int
	Onset, Offset FXCoverResult
	RepeatedCubes [2][2]int
}

func normalizeFXExtractionOptions(options FXExtractionOptions) (FXExtractionOptions, error) {
	if options.Output < 0 || options.Iterations < 1 || options.DivisorsPerIteration < 0 || options.MaxCandidates < 0 || options.MaxCubes < 0 || options.MaxCubeEntries < 0 || options.MaxPairs < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return options, ErrDimension
	}
	if options.MaxCandidates == 0 {
		options.MaxCandidates = 1 << 20
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxCubeEntries == 0 {
		options.MaxCubeEntries = 16 << 20
	}
	if options.MaxPairs == 0 {
		options.MaxPairs = 16 << 20
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
	if options.MaxCandidates < 1 || options.MaxCubes < 1 || options.MaxCubeEntries < 1 || options.MaxPairs < 1 || options.MaxObjects < 1 || options.MaxMappingEntries < 1 || options.MaxVariables < 1 || options.MaxClauses < 1 || options.MaxLiterals < 1 {
		return options, ErrLimit
	}
	if options.SAT.Stop == nil {
		options.SAT.Stop = options.Stop
	}
	return options, nil
}

func fxCoverAssumptions(cube []byte, candidates []SatLit) []SatLit {
	result := make([]SatLit, 0, len(cube))
	for index, value := range cube {
		if value == '-' {
			continue
		}
		literal := candidates[index]
		if value == '0' {
			literal = literal.Not()
		}
		result = append(result, literal)
	}
	return result
}

func fxSolve(solver *SatSolver, assumptions []SatLit, options FXExtractionOptions, stats *SatStats) (SatResult, error) {
	if options.Stop != nil && options.Stop() {
		return SatResult{}, ErrCancelled
	}
	answer, err := solver.Solve(assumptions, options.SAT)
	if err == nil {
		addSatStats(stats, answer.Stats)
	}
	return answer, err
}

// fxSolveCover is the concrete Bmc_FxSolve operation. target denotes the
// function phase whose onset is enumerated; candidates may be correlated.
func fxSolveCover(relation *AigCNF, target SatLit, candidates []SatLit, options FXExtractionOptions) (FXCoverResult, error) {
	var result FXCoverResult
	if relation == nil || int(target.Var()) >= relation.variables || len(candidates) > options.MaxCandidates {
		return result, ErrDimension
	}
	for _, candidate := range candidates {
		if int(candidate.Var()) >= relation.variables {
			return result, ErrInvalidLiteral
		}
	}
	enumeration, onset, err := giaCollapseNewSolver(relation, target, true, GiaCollapseOptions{MaxVariables: options.MaxVariables, Stop: options.Stop, SAT: options.SAT})
	if err != nil {
		return result, err
	}
	clean, offset, err := giaCollapseNewSolver(relation, target, false, GiaCollapseOptions{MaxVariables: options.MaxVariables, Stop: options.Stop, SAT: options.SAT})
	if err != nil {
		return result, err
	}
	if !onset {
		result.Cover = SOPCover{Variables: 0, Cubes: []string{""}, Output: '0'}
		result.Status, result.Reason = SatSatisfiable, SatFinished
		return result, nil
	}
	if !offset {
		result.Cover = SOPCover{Variables: 0, Cubes: []string{""}, Output: '1'}
		result.Status, result.Reason = SatSatisfiable, SatFinished
		return result, nil
	}
	cubes := make([]string, 0)
	for {
		answer, solveErr := fxSolve(enumeration, nil, options, &result.Stats)
		if solveErr != nil {
			return FXCoverResult{}, solveErr
		}
		if answer.Status == SatUnknown {
			result.Status, result.Reason = answer.Status, answer.Reason
			return result, nil
		}
		if answer.Status == SatUnsatisfiable {
			break
		}
		if len(cubes) >= options.MaxCubes {
			return FXCoverResult{}, ErrLimit
		}
		if _, sizeErr := checkedSize(0, len(cubes)+1, len(candidates), options.MaxCubeEntries); sizeErr != nil {
			return FXCoverResult{}, sizeErr
		}
		cube := make([]byte, len(candidates))
		for index, candidate := range candidates {
			if int(candidate.Var()) >= len(answer.Model) {
				return FXCoverResult{}, ErrInvariant
			}
			if answer.Model[candidate.Var()] != candidate.Negative() {
				cube[index] = '1'
			} else {
				cube[index] = '0'
			}
		}
		before := len(candidates)
		for index := range cube {
			saved := cube[index]
			cube[index] = '-'
			trial, trialErr := fxSolve(clean, fxCoverAssumptions(cube, candidates), options, &result.Stats)
			if trialErr != nil {
				return FXCoverResult{}, trialErr
			}
			if trial.Status == SatUnknown {
				result.Status, result.Reason = trial.Status, trial.Reason
				return result, nil
			}
			if trial.Status == SatSatisfiable {
				cube[index] = saved
			}
		}
		assumptions := fxCoverAssumptions(cube, candidates)
		core, coreErr := fxSolve(clean, assumptions, options, &result.Stats)
		if coreErr != nil {
			return FXCoverResult{}, coreErr
		}
		if core.Status == SatUnknown {
			result.Status, result.Reason = core.Status, core.Reason
			return result, nil
		}
		if core.Status != SatUnsatisfiable {
			return FXCoverResult{}, ErrInvariant
		}
		if len(core.Conflict) < len(assumptions) {
			for index, value := range cube {
				if value == '-' {
					continue
				}
				literal := candidates[index]
				if value == '0' {
					literal = literal.Not()
				}
				if !slices.Contains(core.Conflict, literal.Not()) {
					cube[index] = '-'
				}
			}
		}
		literals := 0
		for _, value := range cube {
			if value != '-' {
				literals++
			}
		}
		result.RemovedLiterals += before - literals
		result.Literals += literals
		cubes = append(cubes, string(cube))
		block := fxCoverAssumptions(cube, candidates)
		for index := range block {
			block[index] = block[index].Not()
		}
		ok, addErr := enumeration.AddClause(block)
		if addErr != nil {
			return FXCoverResult{}, addErr
		}
		if !ok {
			break
		}
	}
	result.Cover = SOPCover{Variables: len(candidates), Cubes: cubes, Output: '1'}
	if len(cubes) == 0 {
		return FXCoverResult{}, ErrInvariant
	}
	result.Status, result.Reason = SatSatisfiable, SatFinished
	return result, nil
}

func fxRelationUsage(relation *AigCNF) (int, int) {
	literals := 0
	for _, clause := range relation.clauses {
		literals += len(clause)
	}
	return len(relation.clauses), literals
}

func fxAppendGate(relation *AigCNF, type_ FXDivisorType, inputs [3]SatLit, variable SatVar) error {
	z := SatLit(variable) << 1
	appendClause := func(literals ...SatLit) { relation.clauses = append(relation.clauses, literals) }
	switch type_ {
	case FXDivisorAND:
		a, b := inputs[0], inputs[1]
		appendClause(z.Not(), a)
		appendClause(z.Not(), b)
		appendClause(z, a.Not(), b.Not())
	case FXDivisorXOR:
		a, b := inputs[0], inputs[1]
		appendClause(a, b, z.Not())
		appendClause(a, b.Not(), z)
		appendClause(a.Not(), b, z)
		appendClause(a.Not(), b.Not(), z.Not())
	case FXDivisorMux:
		control, whenTrue, whenFalse := inputs[0], inputs[1], inputs[2]
		appendClause(control.Not(), whenTrue.Not(), z)
		appendClause(control.Not(), whenTrue, z.Not())
		appendClause(control, whenFalse.Not(), z)
		appendClause(control, whenFalse, z.Not())
	default:
		return ErrDimension
	}
	return nil
}

// fxAddClauses is Bmc_FxAddClauses over owned CNF. Selected divisors receive
// variables in reverse selection order, matching iVarStart+nDivs-1-i.
func fxAddClauses(relation *AigCNF, candidates []SatLit, selected []FXDivisor, options FXExtractionOptions) ([]SatLit, []FXDivisor, error) {
	if relation == nil {
		return nil, nil, ErrDimension
	}
	if len(candidates)+len(selected) > options.MaxCandidates || relation.variables+len(selected) > options.MaxVariables {
		return nil, nil, ErrLimit
	}
	base := relation.variables
	appended := make([]FXDivisor, len(selected))
	resolvedBatch := make([]FXDivisor, len(selected))
	additionalClauses, additionalLiterals := 0, 0
	for selection, divisor := range selected {
		resolved := divisor
		if divisor.Type < FXDivisorAND || divisor.Type > FXDivisorMux {
			return nil, nil, ErrDimension
		}
		for input := range resolved.Inputs {
			literal := resolved.Inputs[input]
			if literal == InvalidSatLit {
				if input < 2 || divisor.Type == FXDivisorMux {
					return nil, nil, ErrInvalidLiteral
				}
				continue
			}
			if int(literal.Var()) >= len(candidates) {
				return nil, nil, ErrInvalidLiteral
			}
			resolved.Inputs[input] = candidates[literal.Var()]
			if literal.Negative() {
				resolved.Inputs[input] = resolved.Inputs[input].Not()
			}
		}
		resolvedBatch[selection] = resolved
		if divisor.Type == FXDivisorAND {
			additionalClauses += 3
			additionalLiterals += 7
		} else {
			additionalClauses += 4
			additionalLiterals += 12
		}
		appended[len(selected)-1-selection] = divisor
	}
	clauses, literals := fxRelationUsage(relation)
	if clauses > options.MaxClauses-additionalClauses || literals > options.MaxLiterals-additionalLiterals {
		return nil, nil, ErrLimit
	}
	for selection, resolved := range resolvedBatch {
		variable := SatVar(base + len(selected) - 1 - selection)
		if err := fxAppendGate(relation, resolved.Type, resolved.Inputs, variable); err != nil {
			return nil, nil, err
		}
	}
	relation.variables += len(selected)
	result := slices.Clone(candidates)
	for index := range selected {
		result = append(result, SatLit(SatVar(base+index))<<1)
	}
	return result, appended, nil
}

// ExtractFXSharing performs Bmc_FxComputeOne's iterative cover/divisor loop.
func (manager *GiaManager) ExtractFXSharing(options FXExtractionOptions) (FXExtractionResult, error) {
	var result FXExtractionResult
	if manager == nil {
		return result, ErrDimension
	}
	checked, err := normalizeFXExtractionOptions(options)
	if err != nil {
		return result, err
	}
	if err = manager.Check(); err != nil {
		return result, err
	}
	if checked.Output >= len(manager.cos) {
		return result, ErrDimension
	}
	if len(manager.objects) > checked.MaxObjects || len(manager.objects) > checked.MaxMappingEntries {
		return result, ErrLimit
	}
	aig, _, err := manager.ToAig()
	if err != nil {
		return result, err
	}
	relation, err := EncodeAigCNF(aig)
	if err != nil {
		return result, err
	}
	clauses, literals := fxRelationUsage(relation)
	if relation.variables > checked.MaxVariables || clauses > checked.MaxClauses || literals > checked.MaxLiterals {
		return result, ErrLimit
	}
	candidates := relation.InputLiterals()
	if len(candidates) > checked.MaxCandidates {
		return result, ErrLimit
	}
	result.BaseInputs = len(candidates)
	output := relation.OutputLiterals()[checked.Output]
	for iteration := 0; iteration < checked.Iterations; iteration++ {
		cover, solveErr := fxSolveCover(relation, output, candidates, checked)
		addSatStats(&result.Stats, cover.Stats)
		if solveErr != nil {
			return FXExtractionResult{}, solveErr
		}
		if cover.Status == SatUnknown {
			result.Status, result.Reason = cover.Status, cover.Reason
			return result, nil
		}
		step := FXIteration{Cover: cloneSOPCover(cover.Cover), Candidates: len(candidates), Stats: cover.Stats}
		if cover.Cover.Variables > 0 && checked.DivisorsPerIteration > 0 {
			selected, findErr := FindFXDivisors(cover.Cover, FXDivisorOptions{MaxDivisors: checked.DivisorsPerIteration, MaxCubes: checked.MaxCubes, MaxCubeEntries: checked.MaxCubeEntries, MaxPairs: checked.MaxPairs, Stop: checked.Stop})
			if findErr != nil {
				return FXExtractionResult{}, findErr
			}
			candidates, step.Added, findErr = fxAddClauses(relation, candidates, selected, checked)
			if findErr != nil {
				return FXExtractionResult{}, findErr
			}
			result.Divisors = append(result.Divisors, step.Added...)
		}
		result.Iterations = append(result.Iterations, step)
	}
	result.Status, result.Reason = SatSatisfiable, SatFinished
	return result, nil
}

func fxLiteralValue(values []bool, literal SatLit) (bool, error) {
	if int(literal.Var()) >= len(values) {
		return false, ErrInvalidLiteral
	}
	value := values[literal.Var()]
	if literal.Negative() {
		value = !value
	}
	return value, nil
}

// EvaluateCandidates reconstructs original and derived candidate values.
func (result FXExtractionResult) EvaluateCandidates(inputs []bool) ([]bool, error) {
	if len(inputs) != result.BaseInputs {
		return nil, ErrDimension
	}
	values := slices.Clone(inputs)
	for _, divisor := range result.Divisors {
		left, err := fxLiteralValue(values, divisor.Inputs[0])
		if err != nil {
			return nil, err
		}
		right, err := fxLiteralValue(values, divisor.Inputs[1])
		if err != nil {
			return nil, err
		}
		value := false
		switch divisor.Type {
		case FXDivisorAND:
			value = left && right
		case FXDivisorXOR:
			value = left != right
		case FXDivisorMux:
			whenFalse, readErr := fxLiteralValue(values, divisor.Inputs[2])
			if readErr != nil {
				return nil, readErr
			}
			if left {
				value = right
			} else {
				value = whenFalse
			}
		default:
			return nil, ErrDimension
		}
		values = append(values, value)
	}
	return values, nil
}

func (result FXExtractionResult) EvaluateIteration(iteration int, inputs []bool) (bool, error) {
	if iteration < 0 || iteration >= len(result.Iterations) {
		return false, ErrDimension
	}
	values, err := result.EvaluateCandidates(inputs)
	if err != nil {
		return false, err
	}
	step := result.Iterations[iteration]
	if step.Cover.Variables == 0 {
		return step.Cover.Evaluate(nil)
	}
	if step.Candidates > len(values) || step.Cover.Variables != step.Candidates {
		return false, ErrInvariant
	}
	return step.Cover.Evaluate(values[:step.Candidates])
}

// AnalyzeFXOutputs replaces Bmc_FxCompute's duplicated print-only runs with
// owned onset/offset covers and the exact repeated count table.
func (manager *GiaManager) AnalyzeFXOutputs(options FXExtractionOptions) ([]FXOutputAnalysis, error) {
	if manager == nil {
		return nil, ErrDimension
	}
	checked, err := normalizeFXExtractionOptions(options)
	if err != nil {
		return nil, err
	}
	if err = manager.Check(); err != nil {
		return nil, err
	}
	if len(manager.objects) > checked.MaxObjects || len(manager.objects) > checked.MaxMappingEntries {
		return nil, ErrLimit
	}
	aig, _, err := manager.ToAig()
	if err != nil {
		return nil, err
	}
	relation, err := EncodeAigCNF(aig)
	if err != nil {
		return nil, err
	}
	clauses, literals := fxRelationUsage(relation)
	if relation.variables > checked.MaxVariables || clauses > checked.MaxClauses || literals > checked.MaxLiterals {
		return nil, ErrLimit
	}
	candidates := relation.InputLiterals()
	if len(candidates) > checked.MaxCandidates {
		return nil, ErrLimit
	}
	outputs := relation.OutputLiterals()
	result := make([]FXOutputAnalysis, len(outputs))
	for output, literal := range outputs {
		onset, solveErr := fxSolveCover(relation, literal, candidates, checked)
		if solveErr != nil {
			return nil, solveErr
		}
		offset, solveErr := fxSolveCover(relation, literal.Not(), candidates, checked)
		if solveErr != nil {
			return nil, solveErr
		}
		result[output] = FXOutputAnalysis{Output: output, Onset: onset, Offset: offset}
		onCount, offCount := len(onset.Cover.Cubes), len(offset.Cover.Cubes)
		result[output].RepeatedCubes = [2][2]int{{onCount, offCount}, {onCount, offCount}}
	}
	return result, nil
}
