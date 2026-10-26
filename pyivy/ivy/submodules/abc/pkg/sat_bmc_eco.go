package abc

// ECO miter and patch-cover synthesis from src/sat/bmc/bmcEco.c.

import "slices"

type EcoStatus int8

const (
	EcoUnknown    EcoStatus = -1
	EcoImpossible EcoStatus = 0
	EcoComputed   EcoStatus = 1
)

type EcoLiteral struct {
	Divisor int
	Value   bool
}

type EcoCube []EcoLiteral

type EcoOptions struct {
	MaxObjects, MaxMappingEntries            int
	MaxVariables, MaxClauses, MaxLiterals    int
	MaxIterations, MaxCubes, MaxCubeLiterals int
	SAT                                      SatLimits
	Stop                                     func() bool
}

type EcoPatchResult struct {
	Status EcoStatus
	Reason SatStopReason
	Cubes  []EcoCube
	Stats  SatStats
}

func normalizeEcoOptions(options EcoOptions) (EcoOptions, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 || options.MaxIterations < 0 || options.MaxCubes < 0 || options.MaxCubeLiterals < 0 {
		return options, ErrDimension
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
	if options.MaxClauses == 0 {
		options.MaxClauses = 1 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 4 << 20
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 1 << 20
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxCubeLiterals == 0 {
		options.MaxCubeLiterals = 1 << 20
	}
	if options.SAT.Conflicts == 0 {
		options.SAT.Conflicts = 1_000_000
	}
	return options, nil
}

func ecoStopped(options EcoOptions) bool {
	return options.Stop != nil && options.Stop() || options.SAT.Stop != nil && options.SAT.Stop()
}

func checkEcoPlainGraph(graph *GiaManager) error {
	if graph == nil {
		return ErrDimension
	}
	if err := graph.Check(); err != nil {
		return err
	}
	for id, object := range graph.objects {
		if !object.IsAnd() {
			continue
		}
		if object.IsBuffer() || object.IsXor() {
			return ErrUnsupported
		}
		if _, mux := graph.MuxControl(GiaID(id)); mux {
			return ErrUnsupported
		}
	}
	return nil
}

func ecoMappedSignal(mapping []GiaLit, literal GiaLit) (GiaLit, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) || !mapping[literal.ID()].Valid() {
		return InvalidGiaLit, ErrInvariant
	}
	return mapping[literal.ID()].NotIf(literal.Complemented()), nil
}

func ecoCopyOrdinaryGates(source, target *GiaManager, mapping []GiaLit, root GiaID, replacement GiaLit, options EcoOptions) error {
	for id, object := range source.objects {
		if ecoStopped(options) {
			return ErrCancelled
		}
		if !object.IsAnd() {
			continue
		}
		if GiaID(id) == root {
			mapping[id] = replacement
			continue
		}
		left, _ := object.Fanin0(GiaID(id))
		right, _ := object.Fanin1(GiaID(id))
		left, err := ecoMappedSignal(mapping, left)
		if err != nil {
			return err
		}
		right, err = ecoMappedSignal(mapping, right)
		if err != nil {
			return err
		}
		mapping[id], err = target.HashAnd(left, right)
		if err != nil {
			return err
		}
	}
	return nil
}

func ecoOutputSignals(source *GiaManager, mapping []GiaLit) ([]GiaLit, error) {
	outputs := make([]GiaLit, source.OutputCount())
	for index, id := range source.cos {
		driver, err := source.objects[id].Fanin0(id)
		if err != nil {
			return nil, err
		}
		outputs[index], err = ecoMappedSignal(mapping, driver)
		if err != nil {
			return nil, err
		}
		mapping[id] = outputs[index]
	}
	return outputs, nil
}

// BuildEcoMiter implements Bmc_EcoMiter. The last output of old identifies the
// replacement root. The returned outputs are per-golden-output onset
// differences, one offset difference OR, then the requested divisor signals.
func BuildEcoMiter(golden, old *GiaManager, fans []GiaID, options EcoOptions) (*GiaManager, error) {
	options, err := normalizeEcoOptions(options)
	if err != nil {
		return nil, err
	}
	if err = checkEcoPlainGraph(golden); err != nil {
		return nil, err
	}
	if err = checkEcoPlainGraph(old); err != nil {
		return nil, err
	}
	if golden.InputCount() != old.InputCount() || golden.OutputCount()+1 != old.OutputCount() || old.PrimaryOutputCount() < 1 {
		return nil, ErrDimension
	}
	if golden.ObjectCount() > options.MaxMappingEntries || old.ObjectCount() > options.MaxMappingEntries || len(fans) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	rootCO := old.cos[old.PrimaryOutputCount()-1]
	rootLiteral, err := old.objects[rootCO].Fanin0(rootCO)
	if err != nil || !old.objects[rootLiteral.ID()].IsAnd() {
		return nil, ErrDimension
	}
	root := rootLiteral.ID()
	for _, fan := range fans {
		if uint64(fan) >= uint64(old.ObjectCount()) {
			return nil, ErrInvalidLiteral
		}
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name, target.Spec = golden.Name, golden.Spec
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	goldMap := make([]GiaLit, golden.ObjectCount())
	oldMap := make([]GiaLit, old.ObjectCount())
	for index := range goldMap {
		goldMap[index] = InvalidGiaLit
	}
	for index := range oldMap {
		oldMap[index] = InvalidGiaLit
	}
	goldMap[0], oldMap[0] = GiaConst0, GiaConst0
	for index, id := range golden.cis {
		if ecoStopped(options) {
			return nil, ErrCancelled
		}
		goldMap[id], err = target.AppendCI()
		if err != nil {
			return nil, err
		}
		oldMap[old.cis[index]] = goldMap[id]
	}
	newPI, err := target.AppendCI()
	if err != nil {
		return nil, err
	}
	if err = ecoCopyOrdinaryGates(golden, target, goldMap, InvalidGiaID, InvalidGiaLit, options); err != nil {
		return nil, err
	}
	goldOutputs, err := ecoOutputSignals(golden, goldMap)
	if err != nil {
		return nil, err
	}
	if err = ecoCopyOrdinaryGates(old, target, oldMap, root, newPI, options); err != nil {
		return nil, err
	}
	onsetOutputs, err := ecoOutputSignals(old, oldMap)
	if err != nil {
		return nil, err
	}
	for index, goldenOutput := range goldOutputs {
		difference, err := target.HashXor(goldenOutput, onsetOutputs[index])
		if err != nil {
			return nil, err
		}
		if _, err = target.AppendCO(difference); err != nil {
			return nil, err
		}
	}
	for index := range oldMap {
		if !old.objects[index].IsCI() && index != 0 {
			oldMap[index] = InvalidGiaLit
		}
	}
	if err = ecoCopyOrdinaryGates(old, target, oldMap, root, newPI.Not(), options); err != nil {
		return nil, err
	}
	offsetOutputs, err := ecoOutputSignals(old, oldMap)
	if err != nil {
		return nil, err
	}
	miter := GiaConst0
	for index, goldenOutput := range goldOutputs {
		difference, err := target.HashXor(goldenOutput, offsetOutputs[index])
		if err != nil {
			return nil, err
		}
		miter, err = target.HashOr(miter, difference)
		if err != nil {
			return nil, err
		}
	}
	if _, err = target.AppendCO(miter); err != nil {
		return nil, err
	}
	for _, fan := range fans {
		if !oldMap[fan].Valid() {
			return nil, ErrInvalidLiteral
		}
		if _, err = target.AppendCO(oldMap[fan]); err != nil {
			return nil, err
		}
	}
	cleaned, _, err := target.Cleanup()
	if err != nil {
		return nil, err
	}
	if cleaned.InputCount() != golden.InputCount()+1 || cleaned.OutputCount() != golden.OutputCount()+1+len(fans) {
		return nil, ErrInvariant
	}
	return cleaned, nil
}

func cloneEcoCubes(cubes []EcoCube) []EcoCube {
	result := make([]EcoCube, len(cubes))
	for index := range cubes {
		result[index] = slices.Clone(cubes[index])
	}
	return result
}

// SolveEcoCover implements Bmc_EcoSolve over concrete Go SAT literals. Root is
// the replacement-function input and divisors are the observable fan signals.
func SolveEcoCover(solver *SatSolver, root SatLit, divisors []SatLit, options EcoOptions) (EcoPatchResult, error) {
	var result EcoPatchResult
	options, err := normalizeEcoOptions(options)
	if err != nil {
		return result, err
	}
	if solver == nil || uint64(root.Var()) >= uint64(solver.VariableCount()) {
		return result, ErrDimension
	}
	for _, divisor := range divisors {
		if uint64(divisor.Var()) >= uint64(solver.VariableCount()) {
			return result, ErrInvalidLiteral
		}
	}
	originalLiterals := 0
	for id := 1; id < len(solver.clauses); id++ {
		clause := solver.clauses[id]
		if clause.deleted || clause.learnt {
			continue
		}
		if len(clause.lits) > options.MaxLiterals-originalLiterals {
			return result, ErrLimit
		}
		originalLiterals += len(clause.lits)
	}
	if solver.OriginalClauseCount() > options.MaxClauses {
		return result, ErrLimit
	}
	if solver.VariableCount() >= options.MaxVariables {
		return result, ErrLimit
	}
	selector := SatLit(solver.VariableCount()) << 1
	if err = solver.SetVariableCount(solver.VariableCount() + 1); err != nil {
		return result, err
	}
	finish := func(status EcoStatus, reason SatStopReason) (EcoPatchResult, error) {
		result.Status, result.Reason = status, reason
		result.Stats = solver.Statistics()
		result.Cubes = cloneEcoCubes(result.Cubes)
		return result, nil
	}
	for iteration := 0; iteration < options.MaxIterations; iteration++ {
		if ecoStopped(options) {
			return finish(EcoUnknown, SatCancelled)
		}
		onset, err := solver.Solve([]SatLit{root, selector}, options.SAT)
		if err != nil {
			return result, err
		}
		if onset.Status == SatUnknown {
			return finish(EcoUnknown, onset.Reason)
		}
		if onset.Status == SatUnsatisfiable {
			return finish(EcoComputed, SatFinished)
		}
		assumptions := make([]SatLit, 1, len(divisors)+1)
		assumptions[0] = root.Not()
		values := make([]bool, len(divisors))
		for index, divisor := range divisors {
			values[index] = onset.Model[divisor.Var()] != divisor.Negative()
			if values[index] {
				assumptions = append(assumptions, divisor)
			} else {
				assumptions = append(assumptions, divisor.Not())
			}
		}
		offset, err := solver.Solve(assumptions, options.SAT)
		if err != nil {
			return result, err
		}
		if offset.Status == SatUnknown {
			return finish(EcoUnknown, offset.Reason)
		}
		if offset.Status == SatSatisfiable {
			return finish(EcoImpossible, SatFinished)
		}
		clause := []SatLit{selector.Not()}
		cube := make(EcoCube, 0, len(offset.Conflict))
		used := make([]bool, len(divisors))
		for _, literal := range offset.Conflict {
			if literal == root {
				continue
			}
			found := -1
			for index, divisor := range divisors {
				if !used[index] && divisor.Var() == literal.Var() {
					found = index
					break
				}
			}
			if found < 0 {
				return result, ErrInvariant
			}
			used[found] = true
			clause = append(clause, literal)
			cube = append(cube, EcoLiteral{Divisor: found, Value: values[found]})
		}
		if len(result.Cubes) >= options.MaxCubes || len(cube) > options.MaxCubeLiterals || solver.OriginalClauseCount() >= options.MaxClauses || len(clause) > options.MaxLiterals-originalLiterals {
			return EcoPatchResult{}, ErrLimit
		}
		ok, err := solver.AddClause(clause)
		if err != nil {
			return EcoPatchResult{}, err
		}
		if !ok {
			return EcoPatchResult{}, ErrInvariant
		}
		originalLiterals += len(clause)
		result.Cubes = append(result.Cubes, cube)
	}
	return finish(EcoUnknown, SatConflictLimit)
}

// ComputeEcoPatch implements Bmc_EcoPatch for a miter produced by
// BuildEcoMiter. Output layout is nOuts differences, one offset OR, nIns
// divisor signals; the final PI is the replacement-function root.
func ComputeEcoPatch(miter *GiaManager, nIns, nOuts int, options EcoOptions) (EcoPatchResult, error) {
	var result EcoPatchResult
	options, err := normalizeEcoOptions(options)
	if err != nil {
		return result, err
	}
	if err = checkEcoPlainGraph(miter); err != nil {
		return result, err
	}
	if nIns < 0 || nOuts < 0 || miter.PrimaryOutputCount() != nOuts+1+nIns || miter.PrimaryInputCount() < 1 {
		return result, ErrDimension
	}
	if miter.ObjectCount() > options.MaxMappingEntries {
		return result, ErrLimit
	}
	aig, mapping, err := miter.ToAig()
	if err != nil {
		return result, err
	}
	cnf, err := EncodeAigCNF(aig)
	if err != nil {
		return result, err
	}
	clauses, literals := cnf.Clauses(), 0
	for _, clause := range clauses {
		if len(clause) > options.MaxLiterals-literals {
			return result, ErrLimit
		}
		literals += len(clause)
	}
	if cnf.VariableCount()+1 > options.MaxVariables || len(clauses)+nOuts+1 > options.MaxClauses {
		return result, ErrLimit
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return result, err
	}
	ok, err := cnf.loadIntoWithStop(solver, 0, options.Stop)
	if err != nil {
		return result, err
	}
	if !ok {
		return result, ErrInvariant
	}
	outputs := cnf.OutputLiterals()
	for index := 0; index < nOuts; index++ {
		ok, err = solver.AddClause([]SatLit{outputs[index].Not()})
		if err != nil {
			return result, err
		}
		if !ok {
			return result, ErrInvariant
		}
	}
	ok, err = solver.AddClause([]SatLit{outputs[nOuts]})
	if err != nil {
		return result, err
	}
	if !ok {
		return result, ErrInvariant
	}
	divisors := slices.Clone(outputs[nOuts+1:])
	rootGia := miter.cis[miter.PrimaryInputCount()-1]
	rootAig := mapping[rootGia]
	root, err := cnf.ObjectLiteral(rootAig.ID())
	if err != nil {
		return result, err
	}
	root ^= SatLit(rootAig & 1)
	return SolveEcoCover(solver, root, divisors, options)
}

func (cube EcoCube) Matches(values []bool) (bool, error) {
	for _, literal := range cube {
		if literal.Divisor < 0 || literal.Divisor >= len(values) {
			return false, ErrDimension
		}
		if values[literal.Divisor] != literal.Value {
			return false, nil
		}
	}
	return true, nil
}

func (result EcoPatchResult) Evaluate(values []bool) (bool, error) {
	if result.Status != EcoComputed {
		return false, ErrInvariant
	}
	for _, cube := range result.Cubes {
		match, err := cube.Matches(values)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}
