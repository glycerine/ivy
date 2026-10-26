package abc

// SAT-solver3 GIA integration from src/aig/gia/giaSat3.c.

import "slices"

type GiaSat3Options struct {
	MaxObjects, MaxMappingEntries         int
	MaxVariables, MaxClauses, MaxLiterals int
	SAT                                   SatSolver3Limits
	Stop                                  func() bool
}

type GiaSat3Report struct {
	Output                                int
	Status                                SatStatus
	Reason                                SatStopReason
	Model                                 []bool
	Statistics                            SatStats
	Conflicts                             uint64
	Objects, Variables, Clauses, Literals int
}

type GiaSat3Result struct {
	Split   bool
	Reports []GiaSat3Report
}

// GiaSat3Session owns the concrete solver and its GIA-CI model projection.
type GiaSat3Session struct {
	solver             *SatSolver3
	inputs             []SatLit
	baseUnsatisfiable  bool
	objects, variables int
	clauses, literals  int
	limits             SatSolver3Limits
}

func checkGiaSat3Options(options GiaSat3Options) (GiaSat3Options, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
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
	configured := options.SAT.Stop
	options.SAT.Stop = func() bool {
		return options.Stop != nil && options.Stop() || configured != nil && configured()
	}
	return options, nil
}

func giaSat3SelectedOutputs(manager *GiaManager, selected []int) ([]int, error) {
	if selected == nil {
		selected = make([]int, manager.OutputCount())
		for i := range selected {
			selected[i] = i
		}
	} else {
		selected = slices.Clone(selected)
	}
	if len(selected) == 0 {
		return nil, ErrDimension
	}
	seen := make([]bool, manager.OutputCount())
	for _, output := range selected {
		if output < 0 || output >= manager.OutputCount() || seen[output] {
			return nil, ErrDimension
		}
		seen[output] = true
	}
	return selected, nil
}

// NewGiaSat3Session creates the source's asserted-disjunction CNF session.
// Nil selects every CO; an explicit nonempty slice selects those CO ordinals.
// The CNF is a native stand-alone reference encoding rather than the still
// separate eight-input Mf mapping implementation, but has the same asserted
// Boolean relation and uses the concrete solver3 port.
func NewGiaSat3Session(manager *GiaManager, selected []int, options GiaSat3Options) (*GiaSat3Session, error) {
	if manager == nil {
		return nil, ErrDimension
	}
	options, err := checkGiaSat3Options(options)
	if err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	if err = manager.Check(); err != nil {
		return nil, err
	}
	if manager.ObjectCount() > options.MaxObjects || manager.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	selected, err = giaSat3SelectedOutputs(manager, selected)
	if err != nil {
		return nil, err
	}
	aig, _, err := manager.ToAig()
	if err != nil {
		return nil, err
	}
	if aig.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	relation, err := EncodeAigCNF(aig)
	if err != nil {
		return nil, err
	}
	clauses := relation.Clauses()
	outputs := relation.OutputLiterals()
	inputs := relation.InputLiterals()
	literals := 0
	for _, clause := range clauses {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if len(clause) > options.MaxLiterals-literals {
			return nil, ErrLimit
		}
		literals += len(clause)
	}
	if relation.VariableCount() > options.MaxVariables || len(clauses)+1 > options.MaxClauses || len(selected) > options.MaxLiterals-literals {
		return nil, ErrLimit
	}
	solver, err := NewSatSolver3(SatSolver3Options{MaxVariables: options.MaxVariables})
	if err != nil {
		return nil, err
	}
	fail := func(err error) (*GiaSat3Session, error) {
		solver.Stop()
		return nil, err
	}
	if err = solver.SetVariableCount(relation.VariableCount()); err != nil {
		return fail(err)
	}
	session := &GiaSat3Session{
		solver: solver, inputs: slices.Clone(inputs), objects: aig.ObjectCount(),
		variables: relation.VariableCount(), clauses: len(clauses) + 1,
		literals: literals + len(selected), limits: options.SAT,
	}
	ok, err := solver.AddClauses(clauses)
	if err != nil {
		return fail(err)
	}
	if !ok {
		session.baseUnsatisfiable = true
		return session, nil
	}
	orClause := make([]SatLit, len(selected))
	for i, output := range selected {
		orClause[i] = outputs[output]
	}
	ok, err = solver.AddClause(orClause)
	if err != nil {
		return fail(err)
	}
	if !ok {
		session.baseUnsatisfiable = true
		return session, nil
	}
	ok, err = solver.Simplify()
	if err != nil {
		return fail(err)
	}
	session.baseUnsatisfiable = !ok
	return session, nil
}

func (session *GiaSat3Session) solve(output int) (GiaSat3Report, error) {
	if session == nil || session.solver == nil {
		return GiaSat3Report{}, ErrDimension
	}
	report := GiaSat3Report{
		Output: output, Objects: session.objects, Variables: session.variables,
		Clauses: session.clauses, Literals: session.literals,
	}
	if session.baseUnsatisfiable {
		report.Status, report.Reason = SatUnsatisfiable, SatFinished
		report.Statistics = session.solver.Statistics()
		report.Conflicts = report.Statistics.Conflicts
		return report, nil
	}
	answer, err := session.solver.Solve(nil, session.limits)
	if err != nil {
		return GiaSat3Report{}, err
	}
	report.Status, report.Reason = answer.Status, answer.Reason
	report.Statistics = session.solver.Statistics()
	report.Conflicts = report.Statistics.Conflicts
	if answer.Status == SatSatisfiable {
		report.Model = make([]bool, len(session.inputs))
		for i, literal := range session.inputs {
			if int(literal.Var()) >= len(answer.Model) {
				return GiaSat3Report{}, ErrInvariant
			}
			report.Model[i] = answer.Model[literal.Var()] != literal.Negative()
		}
	}
	return report, nil
}

// Solve runs the owned session once and returns a typed replacement for the
// source's status/timing report.
func (session *GiaSat3Session) Solve() (GiaSat3Report, error) {
	return session.solve(-1)
}

func (session *GiaSat3Session) Close() {
	if session == nil || session.solver == nil {
		return
	}
	session.solver.Stop()
	session.solver = nil
	session.inputs = nil
}

// SolveGiaSat3 checks whether any CO is satisfiable, or checks every CO cone
// independently in source order when split is true.
func (manager *GiaManager) SolveGiaSat3(split bool, options GiaSat3Options) (GiaSat3Result, error) {
	result := GiaSat3Result{Split: split}
	if manager == nil {
		return result, ErrDimension
	}
	options, err := checkGiaSat3Options(options)
	if err != nil {
		return GiaSat3Result{}, err
	}
	if err = manager.Check(); err != nil {
		return GiaSat3Result{}, err
	}
	if manager.OutputCount() == 0 {
		return GiaSat3Result{}, ErrDimension
	}
	if !split {
		session, createErr := NewGiaSat3Session(manager, nil, options)
		if createErr != nil {
			return GiaSat3Result{}, createErr
		}
		defer session.Close()
		report, solveErr := session.solve(-1)
		if solveErr != nil {
			return GiaSat3Result{}, solveErr
		}
		result.Reports = []GiaSat3Report{report}
		return result, nil
	}
	inputs := make([]int, manager.InputCount())
	for i := range inputs {
		inputs[i] = i
	}
	for output, co := range manager.cos {
		if options.Stop != nil && options.Stop() {
			return GiaSat3Result{}, ErrCancelled
		}
		driver, driverErr := manager.objects[co].Fanin0(co)
		if driverErr != nil {
			return GiaSat3Result{}, driverErr
		}
		cone, coneErr := manager.ExtractSignalWithInputs(driver, inputs, GiaCofactorOptions{
			MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, Stop: options.Stop,
		})
		if coneErr != nil {
			return GiaSat3Result{}, coneErr
		}
		session, createErr := NewGiaSat3Session(cone, nil, options)
		if createErr != nil {
			return GiaSat3Result{}, createErr
		}
		report, solveErr := session.solve(output)
		session.Close()
		if solveErr != nil {
			return GiaSat3Result{}, solveErr
		}
		result.Reports = append(result.Reports, report)
	}
	return result, nil
}
