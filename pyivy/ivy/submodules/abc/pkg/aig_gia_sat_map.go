package abc

// SAT mapping manager from src/aig/gia/giaSatMap.c.

import "slices"

type GiaSatMapSelectionKind uint8

const (
	GiaSatMapSelectCut GiaSatMapSelectionKind = iota
	GiaSatMapSelectInputInverter
)

type GiaSatMapSelection struct {
	Kind  GiaSatMapSelectionKind
	Index int
}

type GiaSatMapFaninKind uint8

const (
	GiaSatMapFaninObjectLiteral GiaSatMapFaninKind = iota
	GiaSatMapFaninInputInverter
)

// GiaSatMapFanin names one dependency of a cut. Index addresses either an
// object-literal variable or a complemented primary-input inverter.
type GiaSatMapFanin struct {
	Kind  GiaSatMapFaninKind
	Index int
}

type GiaSatMapCut struct {
	Root   int // object-literal variable implemented by this cut
	Fanins []GiaSatMapFanin
	Gate   int
	Area   uint64
}

// GiaSatMapProblem is the owned output of the source's Nf_ManExtractWindow
// dependency. Object literals and complemented-input inverters occupy the
// cardinality inputs in that order; cuts receive variables after the compiled
// cardinality network.
type GiaSatMapProblem struct {
	LogN, ObjectLiterals, Inputs int
	Roots                        []int
	Cuts                         []GiaSatMapCut
	Initial                      []GiaSatMapSelection
	InverterArea                 uint64
}

type GiaSatMapSolveResult struct {
	SAT               SatResult
	Bound             int
	SelectedObjects   []int
	SelectedInverters []int
	SelectedCuts      []int
	SelectedGates     []int
	Area              uint64
}

type GiaSatMapOptimization struct {
	InitialCount int
	Attempts     []GiaSatMapSolveResult
	Best         GiaSatMapSolveResult
}

// GiaSatMapper owns the source cardinality prefix and mapping clauses.
type GiaSatMapper struct {
	problem     GiaSatMapProblem
	cardinality GiaCardinalityCNF
	solver      *SatSolver
	clauses     [][]SatLit
	firstCut    int
	capacity    int
}

func cloneGiaSatMapProblem(problem GiaSatMapProblem) GiaSatMapProblem {
	result := problem
	result.Roots = slices.Clone(problem.Roots)
	result.Initial = slices.Clone(problem.Initial)
	result.Cuts = make([]GiaSatMapCut, len(problem.Cuts))
	for index, cut := range problem.Cuts {
		result.Cuts[index] = cut
		result.Cuts[index].Fanins = slices.Clone(cut.Fanins)
	}
	return result
}

func giaSatMapLiteral(variable int, negative bool) SatLit {
	literal, _ := MakeSatLit(SatVar(variable), negative)
	return literal
}

func validateGiaSatMapProblem(problem GiaSatMapProblem) error {
	if problem.LogN < 0 || problem.LogN > 20 || problem.ObjectLiterals < 0 || problem.ObjectLiterals&1 != 0 || problem.Inputs < 0 {
		return ErrDimension
	}
	capacity := 1 << uint(problem.LogN)
	if problem.ObjectLiterals > capacity || problem.Inputs > capacity-problem.ObjectLiterals {
		return ErrDimension
	}
	for _, root := range problem.Roots {
		if root < 0 || root >= problem.ObjectLiterals {
			return ErrDimension
		}
	}
	counts := make([]int, problem.ObjectLiterals)
	for _, cut := range problem.Cuts {
		if cut.Root < 0 || cut.Root >= problem.ObjectLiterals || cut.Gate < 0 {
			return ErrDimension
		}
		counts[cut.Root]++
		for _, fanin := range cut.Fanins {
			if fanin.Index < 0 || fanin.Kind > GiaSatMapFaninInputInverter || fanin.Kind == GiaSatMapFaninObjectLiteral && fanin.Index >= problem.ObjectLiterals || fanin.Kind == GiaSatMapFaninInputInverter && fanin.Index >= problem.Inputs {
				return ErrDimension
			}
		}
	}
	for _, count := range counts {
		if count == 0 {
			return ErrDimension
		}
	}
	seenCuts := make([]bool, len(problem.Cuts))
	seenInputs := make([]bool, problem.Inputs)
	for _, selected := range problem.Initial {
		if selected.Index < 0 || selected.Kind > GiaSatMapSelectInputInverter {
			return ErrDimension
		}
		if selected.Kind == GiaSatMapSelectCut {
			if selected.Index >= len(problem.Cuts) || seenCuts[selected.Index] {
				return ErrDimension
			}
			seenCuts[selected.Index] = true
		} else {
			if selected.Index >= problem.Inputs || seenInputs[selected.Index] {
				return ErrDimension
			}
			seenInputs[selected.Index] = true
		}
	}
	return nil
}

// CheckGiaSatMapSolution verifies the exact closure checked by Sbm_ManCheckSol
// and returns the native preferred-true variable order.
func CheckGiaSatMapSolution(problem GiaSatMapProblem, selections []GiaSatMapSelection, firstCut int) ([]SatVar, error) {
	if err := validateGiaSatMapProblem(problem); err != nil || firstCut < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if _, err := checkedSize(firstCut, 1, len(problem.Cuts), int(MaxLiteralID)+1); err != nil {
		return nil, err
	}
	usedObjects := make([]bool, problem.ObjectLiterals)
	usedInputs := make([]bool, problem.Inputs)
	seenCuts := make([]bool, len(problem.Cuts))
	seenInputs := make([]bool, problem.Inputs)
	preferred := make([]SatVar, 0, 2*len(selections))
	for _, selected := range selections {
		if selected.Index < 0 || selected.Kind > GiaSatMapSelectInputInverter {
			return nil, ErrDimension
		}
		if selected.Kind == GiaSatMapSelectInputInverter {
			if selected.Index >= problem.Inputs || seenInputs[selected.Index] {
				return nil, ErrDimension
			}
			seenInputs[selected.Index], usedInputs[selected.Index] = true, true
			preferred = append(preferred, SatVar(problem.ObjectLiterals+selected.Index))
			continue
		}
		if selected.Index >= len(problem.Cuts) || seenCuts[selected.Index] {
			return nil, ErrDimension
		}
		seenCuts[selected.Index] = true
		cut := problem.Cuts[selected.Index]
		preferred = append(preferred, SatVar(firstCut+selected.Index))
		if !usedObjects[cut.Root] {
			usedObjects[cut.Root] = true
			preferred = append(preferred, SatVar(cut.Root))
		}
	}
	for _, root := range problem.Roots {
		if !usedObjects[root] {
			return nil, ErrInvariant
		}
	}
	for cutIndex, selected := range seenCuts {
		if !selected {
			continue
		}
		for _, fanin := range problem.Cuts[cutIndex].Fanins {
			if fanin.Kind == GiaSatMapFaninObjectLiteral && !usedObjects[fanin.Index] || fanin.Kind == GiaSatMapFaninInputInverter && !usedInputs[fanin.Index] {
				return nil, ErrInvariant
			}
		}
		preferred = append(preferred, SatVar(firstCut+cutIndex))
	}
	return preferred, nil
}

func NewGiaSatMapper(problem GiaSatMapProblem, options GiaWindowCNFOptions) (*GiaSatMapper, error) {
	if err := validateGiaSatMapProblem(problem); err != nil || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
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
	cardinality, err := BuildGiaCardinalityCNF(problem.LogN, options)
	if err != nil {
		return nil, err
	}
	firstCut := cardinality.Variables
	variables, err := checkedSize(firstCut, 1, len(problem.Cuts), min(options.MaxVariables, int(MaxLiteralID)+1))
	if err != nil {
		return nil, err
	}
	preferred, err := CheckGiaSatMapSolution(problem, problem.Initial, firstCut)
	if err != nil {
		return nil, err
	}
	clauses := make([][]SatLit, 0, len(cardinality.Clauses)+problem.ObjectLiterals+len(problem.Cuts)*2)
	literals := 0
	add := func(clause []SatLit) error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if len(clauses) == options.MaxClauses || len(clause) > options.MaxLiterals-literals {
			return ErrLimit
		}
		literals += len(clause)
		clauses = append(clauses, clause)
		return nil
	}
	for _, clause := range cardinality.Clauses {
		if err = add(slices.Clone(clause)); err != nil {
			return nil, err
		}
	}
	byRoot := make([][]int, problem.ObjectLiterals)
	for index, cut := range problem.Cuts {
		byRoot[cut.Root] = append(byRoot[cut.Root], index)
	}
	last := make([]int, problem.ObjectLiterals)
	for object, indexes := range byRoot {
		clause := make([]SatLit, 1, len(indexes)+1)
		clause[0] = giaSatMapLiteral(object, true)
		for _, cut := range indexes {
			clause = append(clause, giaSatMapLiteral(firstCut+cut, false))
		}
		last[object] = indexes[len(indexes)-1]
		if err = add(clause); err != nil {
			return nil, err
		}
		if object&1 != 0 {
			if err = add([]SatLit{giaSatMapLiteral(firstCut+last[object], true), giaSatMapLiteral(firstCut+last[object-1], true)}); err != nil {
				return nil, err
			}
		}
	}
	for cutIndex, cut := range problem.Cuts {
		if err = add([]SatLit{giaSatMapLiteral(cut.Root, false), giaSatMapLiteral(firstCut+cutIndex, true)}); err != nil {
			return nil, err
		}
		for _, fanin := range cut.Fanins {
			variable := fanin.Index
			if fanin.Kind == GiaSatMapFaninInputInverter {
				variable += problem.ObjectLiterals
			}
			if err = add([]SatLit{giaSatMapLiteral(variable, false), giaSatMapLiteral(firstCut+cutIndex, true)}); err != nil {
				return nil, err
			}
		}
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return nil, err
	}
	if err = solver.SetVariableCount(variables); err != nil {
		return nil, err
	}
	for _, clause := range clauses {
		if _, err = solver.AddClause(clause); err != nil {
			return nil, err
		}
	}
	if err = solver.SetPreferredVariables(preferred); err != nil {
		return nil, err
	}
	return &GiaSatMapper{problem: cloneGiaSatMapProblem(problem), cardinality: cardinality, solver: solver, clauses: clauses, firstCut: firstCut, capacity: 1 << uint(problem.LogN)}, nil
}

func (m *GiaSatMapper) Clauses() [][]SatLit {
	if m == nil || m.solver == nil {
		return nil
	}
	result := make([][]SatLit, len(m.clauses))
	for index := range result {
		result[index] = slices.Clone(m.clauses[index])
	}
	return result
}

func (m *GiaSatMapper) Assumptions(bound int) ([]SatLit, error) {
	if m == nil || m.solver == nil || bound < 0 || bound > m.capacity {
		return nil, ErrDimension
	}
	result := make([]SatLit, 0, 1+m.capacity-m.problem.ObjectLiterals-m.problem.Inputs+len(m.problem.Roots))
	if bound < m.capacity {
		result = append(result, giaSatMapLiteral(int(m.cardinality.Outputs[bound]), true))
	}
	for variable := m.problem.ObjectLiterals + m.problem.Inputs; variable < m.capacity; variable++ {
		result = append(result, giaSatMapLiteral(variable, true))
	}
	for _, root := range m.problem.Roots {
		result = append(result, giaSatMapLiteral(root, false))
	}
	return result, nil
}

func (m *GiaSatMapper) Solve(bound int, limits SatLimits) (GiaSatMapSolveResult, error) {
	assumptions, err := m.Assumptions(bound)
	if err != nil {
		return GiaSatMapSolveResult{}, err
	}
	answer, err := m.solver.Solve(assumptions, limits)
	if err != nil {
		return GiaSatMapSolveResult{}, err
	}
	result := GiaSatMapSolveResult{SAT: answer, Bound: bound}
	if answer.Status != SatSatisfiable {
		return result, nil
	}
	for object := 0; object < m.problem.ObjectLiterals; object++ {
		if answer.Model[object] {
			result.SelectedObjects = append(result.SelectedObjects, object)
		}
	}
	for input := 0; input < m.problem.Inputs; input++ {
		if answer.Model[m.problem.ObjectLiterals+input] {
			result.SelectedInverters = append(result.SelectedInverters, input)
			if result.Area > ^uint64(0)-m.problem.InverterArea {
				return GiaSatMapSolveResult{}, ErrLimit
			}
			result.Area += m.problem.InverterArea
		}
	}
	for cutIndex, cut := range m.problem.Cuts {
		if answer.Model[m.firstCut+cutIndex] {
			result.SelectedCuts = append(result.SelectedCuts, cutIndex)
			result.SelectedGates = append(result.SelectedGates, cut.Gate)
			if result.Area > ^uint64(0)-cut.Area {
				return GiaSatMapSolveResult{}, ErrLimit
			}
			result.Area += cut.Area
		}
	}
	return result, nil
}

// Optimize follows Sbm_ManTestSat: try one fewer selected cardinality input at
// a time, retain the last satisfiable mapping, and stop at UNSAT or UNKNOWN.
func (m *GiaSatMapper) Optimize(limits SatLimits) (GiaSatMapOptimization, error) {
	if m == nil || m.solver == nil {
		return GiaSatMapOptimization{}, ErrDimension
	}
	result := GiaSatMapOptimization{InitialCount: len(m.problem.Initial)}
	for bound := result.InitialCount - 1; bound > 0; bound-- {
		attempt, err := m.Solve(bound, limits)
		if err != nil {
			return GiaSatMapOptimization{}, err
		}
		result.Attempts = append(result.Attempts, attempt)
		if attempt.SAT.Status == SatSatisfiable {
			result.Best = attempt
			continue
		}
		break
	}
	return result, nil
}

func (m *GiaSatMapper) Close() {
	if m == nil || m.solver == nil {
		return
	}
	m.solver = nil
	m.problem = GiaSatMapProblem{}
	m.cardinality = GiaCardinalityCNF{}
	m.clauses = nil
}
