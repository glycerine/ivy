package abc

// Satoko-style GIA SAT integration from src/aig/gia/giaSatoko.c.

import (
	"bufio"
	"bytes"
	"slices"
	"strconv"
	"strings"
)

type GiaSatokoMode uint8

const (
	GiaSatokoAggregate GiaSatokoMode = iota
	GiaSatokoSplit
	GiaSatokoIncremental
)

type GiaSatokoOptions struct {
	MaxObjects, MaxMappingEntries         int
	MaxVariables, MaxClauses, MaxLiterals int
	SAT                                   SatLimits
	Stop                                  func() bool
}

type GiaSatokoReport struct {
	Output                                int
	Status                                SatStatus
	Reason                                SatStopReason
	Model                                 []bool
	Statistics                            SatStats
	Conflicts                             uint64
	ConeVariables                         []SatVar
	Objects, Variables, Clauses, Literals int
}

type GiaSatokoResult struct {
	Mode    GiaSatokoMode
	Reports []GiaSatokoReport
}

// GiaSatokoSession owns one concrete Go solver. It preserves the wrapper's
// Boolean relation, lifecycle and reporting roles; it does not claim Satoko's
// internal branching or cone-scheduling heuristics.
type GiaSatokoSession struct {
	solver             *SatSolver
	inputs, outputs    []SatLit
	mapping            []int
	baseUnsatisfiable  bool
	objects, variables int
	clauses, literals  int
	limits             SatLimits
	stop               func() bool
}

func checkGiaSatokoOptions(options GiaSatokoOptions) (GiaSatokoOptions, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
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
		options.MaxClauses = 1 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 4 << 20
	}
	configured := options.SAT.Stop
	options.SAT.Stop = func() bool {
		return options.Stop != nil && options.Stop() || configured != nil && configured()
	}
	return options, nil
}

// CollectCNFMappingVariables iteratively expands packed records of the form
// [fanin-count, fanin-var...] and returns unique variables in DFS postorder.
// The first variableCount entries map variables to record offsets; zero marks
// a leaf. Unlike the C assertion, a leaf root is accepted and returned.
func CollectCNFMappingVariables(root SatVar, mapping []int, variableCount, maxVisits int, stop func() bool) ([]SatVar, error) {
	if variableCount < 0 || maxVisits < 0 || len(mapping) < variableCount || uint64(root) >= uint64(variableCount) {
		return nil, ErrDimension
	}
	type frame struct {
		variable SatVar
		exit     bool
	}
	state := make([]uint8, variableCount)
	stack := []frame{{variable: root}}
	result := make([]SatVar, 0)
	visits := 0
	for len(stack) > 0 {
		if stop != nil && stop() {
			return nil, ErrCancelled
		}
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		variable := int(current.variable)
		if current.exit {
			state[variable] = 2
			result = append(result, current.variable)
			continue
		}
		if state[variable] == 2 {
			continue
		}
		if state[variable] == 1 {
			return nil, ErrInvariant
		}
		visits++
		if maxVisits > 0 && visits > maxVisits {
			return nil, ErrLimit
		}
		state[variable] = 1
		stack = append(stack, frame{variable: current.variable, exit: true})
		offset := mapping[variable]
		if offset < 0 || offset >= len(mapping) {
			return nil, ErrDimension
		}
		if offset == 0 {
			continue
		}
		count := mapping[offset]
		if count < 0 || count > len(mapping)-offset-1 {
			return nil, ErrDimension
		}
		for index := count; index >= 1; index-- {
			child := mapping[offset+index]
			if child < 0 || child >= variableCount {
				return nil, ErrDimension
			}
			if state[child] == 1 {
				return nil, ErrInvariant
			}
			if state[child] == 0 {
				stack = append(stack, frame{variable: SatVar(child)})
			}
		}
	}
	return result, nil
}

func giaSatokoMapping(aig *AigManager, relation *AigCNF, maximum int) ([]int, error) {
	if relation.VariableCount() > maximum {
		return nil, ErrLimit
	}
	mapping := make([]int, relation.VariableCount())
	for id, object := range aig.objects {
		if object.Type != AigAND {
			continue
		}
		root, err := relation.ObjectLiteral(AigID(id))
		if err != nil {
			return nil, err
		}
		left, err := relation.ObjectLiteral(object.Fanin0.ID())
		if err != nil {
			return nil, err
		}
		right, err := relation.ObjectLiteral(object.Fanin1.ID())
		if err != nil {
			return nil, err
		}
		if len(mapping) > maximum-3 {
			return nil, ErrLimit
		}
		mapping[root.Var()] = len(mapping)
		mapping = append(mapping, 2, int(left.Var()), int(right.Var()))
	}
	return mapping, nil
}

func newGiaSatokoRelation(manager *GiaManager, assertOutputs []int, options GiaSatokoOptions) (*GiaSatokoSession, error) {
	if manager == nil {
		return nil, ErrDimension
	}
	var err error
	options, err = checkGiaSatokoOptions(options)
	if err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	if err = manager.Check(); err != nil {
		return nil, err
	}
	if manager.ObjectCount() > options.MaxObjects {
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
	clauses := relation.Clauses()
	outputs := relation.OutputLiterals()
	if relation.VariableCount() > options.MaxVariables || len(clauses) > options.MaxClauses {
		return nil, ErrLimit
	}
	literalCount := 0
	for _, clause := range clauses {
		if len(clause) > options.MaxLiterals-literalCount {
			return nil, ErrLimit
		}
		literalCount += len(clause)
	}
	if assertOutputs != nil {
		if len(assertOutputs) == 0 {
			return nil, ErrDimension
		}
		if len(clauses) == options.MaxClauses || len(assertOutputs) > options.MaxLiterals-literalCount {
			return nil, ErrLimit
		}
		seen := make([]bool, len(outputs))
		clause := make([]SatLit, len(assertOutputs))
		for index, output := range assertOutputs {
			if output < 0 || output >= len(outputs) || seen[output] {
				return nil, ErrDimension
			}
			seen[output] = true
			clause[index] = outputs[output]
		}
		clauses = append(clauses, clause)
		literalCount += len(clause)
	}
	mapping, err := giaSatokoMapping(aig, relation, options.MaxMappingEntries)
	if err != nil {
		return nil, err
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return nil, err
	}
	if err = solver.SetVariableCount(relation.VariableCount()); err != nil {
		return nil, err
	}
	session := &GiaSatokoSession{
		solver: solver, inputs: relation.InputLiterals(), outputs: outputs, mapping: mapping,
		objects: aig.ObjectCount(), variables: relation.VariableCount(), clauses: len(clauses),
		literals: literalCount, limits: options.SAT, stop: options.Stop,
	}
	for _, clause := range clauses {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		ok, addErr := solver.AddClause(clause)
		if addErr != nil {
			return nil, addErr
		}
		if !ok {
			session.baseUnsatisfiable = true
			break
		}
	}
	if !session.baseUnsatisfiable {
		ok, simplifyErr := solver.Simplify()
		if simplifyErr != nil {
			return nil, simplifyErr
		}
		session.baseUnsatisfiable = !ok
	}
	return session, nil
}

// NewGiaSatokoSession creates the asserted disjunction used by the source's
// create/call-one path. Nil selects every output.
func NewGiaSatokoSession(manager *GiaManager, selected []int, options GiaSatokoOptions) (*GiaSatokoSession, error) {
	if manager == nil || manager.OutputCount() == 0 {
		return nil, ErrDimension
	}
	if selected == nil {
		selected = make([]int, manager.OutputCount())
		for index := range selected {
			selected[index] = index
		}
	} else {
		selected = slices.Clone(selected)
	}
	return newGiaSatokoRelation(manager, selected, options)
}

func (session *GiaSatokoSession) solve(output int, assumption *SatLit, cone []SatVar) (GiaSatokoReport, error) {
	if session == nil || session.solver == nil {
		return GiaSatokoReport{}, ErrDimension
	}
	report := GiaSatokoReport{Output: output, ConeVariables: slices.Clone(cone), Objects: session.objects, Variables: session.variables, Clauses: session.clauses, Literals: session.literals}
	if session.baseUnsatisfiable {
		report.Status, report.Reason = SatUnsatisfiable, SatFinished
		return report, nil
	}
	var assumptions []SatLit
	if assumption != nil {
		assumptions = []SatLit{*assumption}
	}
	answer, err := session.solver.Solve(assumptions, session.limits)
	if err != nil {
		return GiaSatokoReport{}, err
	}
	report.Status, report.Reason, report.Statistics = answer.Status, answer.Reason, answer.Stats
	report.Conflicts = answer.Stats.Conflicts
	if answer.Status == SatSatisfiable {
		report.Model = make([]bool, len(session.inputs))
		for index, literal := range session.inputs {
			if uint64(literal.Var()) >= uint64(len(answer.Model)) {
				return GiaSatokoReport{}, ErrInvariant
			}
			report.Model[index] = answer.Model[literal.Var()] != literal.Negative()
		}
	}
	return report, nil
}

func (session *GiaSatokoSession) Solve() (GiaSatokoReport, error) {
	return session.solve(-1, nil, nil)
}

func (session *GiaSatokoSession) Close() {
	if session == nil {
		return
	}
	session.solver, session.inputs, session.outputs, session.mapping = nil, nil, nil, nil
}

// SolveGiaSatoko runs aggregate, split-cone or shared incremental mode.
func (manager *GiaManager) SolveGiaSatoko(mode GiaSatokoMode, options GiaSatokoOptions) (GiaSatokoResult, error) {
	result := GiaSatokoResult{Mode: mode}
	if manager == nil || manager.OutputCount() == 0 || mode > GiaSatokoIncremental {
		return GiaSatokoResult{}, ErrDimension
	}
	options, err := checkGiaSatokoOptions(options)
	if err != nil {
		return GiaSatokoResult{}, err
	}
	if mode == GiaSatokoAggregate {
		session, createErr := NewGiaSatokoSession(manager, nil, options)
		if createErr != nil {
			return GiaSatokoResult{}, createErr
		}
		defer session.Close()
		report, solveErr := session.Solve()
		if solveErr != nil {
			return GiaSatokoResult{}, solveErr
		}
		result.Reports = []GiaSatokoReport{report}
		return result, nil
	}
	if mode == GiaSatokoSplit {
		inputs := make([]int, manager.InputCount())
		for index := range inputs {
			inputs[index] = index
		}
		for output, co := range manager.cos {
			if options.Stop != nil && options.Stop() {
				return GiaSatokoResult{}, ErrCancelled
			}
			driver, driverErr := manager.objects[co].Fanin0(co)
			if driverErr != nil {
				return GiaSatokoResult{}, driverErr
			}
			cone, coneErr := manager.ExtractSignalWithInputs(driver, inputs, GiaCofactorOptions{MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, Stop: options.Stop})
			if coneErr != nil {
				return GiaSatokoResult{}, coneErr
			}
			session, createErr := NewGiaSatokoSession(cone, nil, options)
			if createErr != nil {
				return GiaSatokoResult{}, createErr
			}
			report, solveErr := session.solve(output, nil, nil)
			session.Close()
			if solveErr != nil {
				return GiaSatokoResult{}, solveErr
			}
			result.Reports = append(result.Reports, report)
		}
		return result, nil
	}
	session, err := newGiaSatokoRelation(manager, nil, options)
	if err != nil {
		return GiaSatokoResult{}, err
	}
	defer session.Close()
	for output, literal := range session.outputs {
		cone, collectErr := CollectCNFMappingVariables(literal.Var(), session.mapping, session.variables, session.variables, options.Stop)
		if collectErr != nil {
			return GiaSatokoResult{}, collectErr
		}
		report, solveErr := session.solve(output, &literal, cone)
		if solveErr != nil {
			return GiaSatokoResult{}, solveErr
		}
		result.Reports = append(result.Reports, report)
	}
	return result, nil
}

// ParseGiaSatokoDIMACS builds a stand-alone session from an in-memory DIMACS
// dataset. Comments and one optional "p cnf" header are accepted.
func ParseGiaSatokoDIMACS(data []byte, options GiaSatokoOptions) (*GiaSatokoSession, error) {
	options, err := checkGiaSatokoOptions(options)
	if err != nil {
		return nil, err
	}
	var clauses [][]SatLit
	var current []SatLit
	declaredVariables, declaredClauses := -1, -1
	maximumVariable, literalCount := 0, 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024), options.MaxLiterals*24+1024)
	for scanner.Scan() {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || fields[0] == "c" {
			continue
		}
		if fields[0] == "p" {
			if len(fields) != 4 || fields[1] != "cnf" || declaredVariables >= 0 || len(current) != 0 || len(clauses) != 0 {
				return nil, ErrDimension
			}
			declaredVariables, err = strconv.Atoi(fields[2])
			if err != nil || declaredVariables < 0 || declaredVariables > options.MaxVariables {
				return nil, ErrDimension
			}
			declaredClauses, err = strconv.Atoi(fields[3])
			if err != nil || declaredClauses < 0 || declaredClauses > options.MaxClauses {
				return nil, ErrDimension
			}
			continue
		}
		for _, field := range fields {
			value, parseErr := strconv.Atoi(field)
			if parseErr != nil {
				return nil, ErrInvalidLiteral
			}
			if value == 0 {
				if len(clauses) == options.MaxClauses {
					return nil, ErrLimit
				}
				clauses = append(clauses, current)
				current = nil
				continue
			}
			literal, literalErr := ParseSatDIMACSLiteral(value)
			if literalErr != nil {
				return nil, literalErr
			}
			literalCount++
			if literalCount > options.MaxLiterals {
				return nil, ErrLimit
			}
			if int(literal.Var())+1 > maximumVariable {
				maximumVariable = int(literal.Var()) + 1
			}
			current = append(current, literal)
		}
	}
	if err = scanner.Err(); err != nil || len(current) != 0 {
		return nil, ErrDimension
	}
	if declaredVariables >= 0 && (maximumVariable > declaredVariables || len(clauses) != declaredClauses) {
		return nil, ErrDimension
	}
	variables := maximumVariable
	if declaredVariables >= 0 {
		variables = declaredVariables
	}
	solver, err := NewSatSolverWithLimit(options.MaxVariables)
	if err != nil {
		return nil, err
	}
	if err = solver.SetVariableCount(variables); err != nil {
		return nil, err
	}
	inputs := make([]SatLit, variables)
	for variable := range inputs {
		inputs[variable], _ = MakeSatLit(SatVar(variable), false)
	}
	session := &GiaSatokoSession{solver: solver, inputs: inputs, variables: variables, clauses: len(clauses), literals: literalCount, limits: options.SAT, stop: options.Stop}
	for _, clause := range clauses {
		if len(clause) == 0 {
			session.baseUnsatisfiable = true
			break
		}
		ok, addErr := solver.AddClause(clause)
		if addErr != nil {
			return nil, addErr
		}
		if !ok {
			session.baseUnsatisfiable = true
			break
		}
	}
	return session, nil
}
