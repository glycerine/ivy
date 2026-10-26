package abc

// Clock-gating construction and selection from src/opt/cgt/cgt.h,
// src/opt/cgt/cgtInt.h, cgtAig.c, cgtCore.c, cgtDecide.c, and cgtMan.c.

import "slices"

// ClockGatingParameters is the owned Cgt_Par_t representation.
type ClockGatingParameters struct {
	LevelMax, CandidateMax, ODCMax int
	ConflictMax, VariablesMin      int
	FlopsMin                       int
	AreaOnly, Verbose, VeryVerbose bool
}

func DefaultClockGatingParameters() ClockGatingParameters {
	return ClockGatingParameters{LevelMax: 25, CandidateMax: 1000, ConflictMax: 10, VariablesMin: 1000, FlopsMin: 10}
}

// ClockGatingLimits bounds all retained state. Zero selects a finite default.
type ClockGatingLimits struct {
	MaxObjects, MaxCandidates, MaxSATVariables, MaxCalls int
	Stop                                                 func() bool
}

func (l ClockGatingLimits) normalized() (ClockGatingLimits, error) {
	if l.MaxObjects < 0 || l.MaxCandidates < 0 || l.MaxSATVariables < 0 || l.MaxCalls < 0 {
		return l, ErrDimension
	}
	if l.MaxObjects == 0 {
		l.MaxObjects = 1 << 20
	}
	if l.MaxCandidates == 0 {
		l.MaxCandidates = 1 << 20
	}
	if l.MaxSATVariables == 0 {
		l.MaxSATVariables = 1 << 20
	}
	if l.MaxCalls == 0 {
		l.MaxCalls = 1 << 20
	}
	return l, nil
}

// ClockGatingStats is the non-timing portion of Cgt_Man_t statistics.
type ClockGatingStats struct {
	Recycles, Calls, Satisfiable, Unsatisfiable, Undecided, Filtered int
}

// ClockGatingManager is the concrete owned Cgt_Man_t state. Candidate and
// selected gate slices are indexed by register number and preserve source order.
type ClockGatingManager struct {
	parameters ClockGatingParameters
	limits     ClockGatingLimits
	graph      *AigManager
	care       *AigManager
	useful     []bool
	candidates [][]AigLit
	gates      [][]AigLit
	stats      ClockGatingStats
	closed     bool
}

func NewClockGatingManager(graph, care *AigManager, parameters ClockGatingParameters, limits ClockGatingLimits) (*ClockGatingManager, error) {
	limits, err := limits.normalized()
	if err != nil || graph == nil {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if parameters == (ClockGatingParameters{}) {
		parameters = DefaultClockGatingParameters()
	}
	if parameters.LevelMax < 0 || parameters.CandidateMax < 0 || parameters.ODCMax < 0 || parameters.ConflictMax < 0 || parameters.VariablesMin < 0 || parameters.FlopsMin < 0 {
		return nil, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return nil, err
	}
	if graph.RegisterCount() == 0 {
		return nil, ErrDimension
	}
	if graph.ObjectCount() > limits.MaxObjects {
		return nil, ErrLimit
	}
	if care != nil {
		if err = care.Check(); err != nil {
			return nil, err
		}
		if care.InputCount() != graph.InputCount() {
			care = nil
		}
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &ClockGatingManager{
		parameters: parameters, limits: limits, graph: graph, care: care,
		candidates: make([][]AigLit, graph.RegisterCount()), gates: make([][]AigLit, graph.RegisterCount()),
	}, nil
}

func (m *ClockGatingManager) check() error {
	if m == nil || m.closed || m.graph == nil {
		return ErrInvariant
	}
	return nil
}

func (m *ClockGatingManager) stopped() bool { return m.limits.Stop != nil && m.limits.Stop() }

func (m *ClockGatingManager) SetUseful(useful []bool) error {
	if err := m.check(); err != nil {
		return err
	}
	if useful != nil && len(useful) != m.graph.ObjectCount() {
		return ErrDimension
	}
	if m.stopped() {
		return ErrCancelled
	}
	m.useful = slices.Clone(useful)
	return nil
}

func (m *ClockGatingManager) Candidates() ([][]AigLit, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	return cloneClockGates(m.candidates), nil
}

func (m *ClockGatingManager) Gates() ([][]AigLit, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	return cloneClockGates(m.gates), nil
}

func (m *ClockGatingManager) Stats() (ClockGatingStats, error) {
	if err := m.check(); err != nil {
		return ClockGatingStats{}, err
	}
	return m.stats, nil
}

func (m *ClockGatingManager) Clean() error {
	if err := m.check(); err != nil {
		return err
	}
	if m.stopped() {
		return ErrCancelled
	}
	// Cgt_ManClean releases only partition-local CNF, SAT, and simulation
	// storage. This concrete manager does not retain those between calls; the
	// accumulated candidates, selected gates, and counters remain intact.
	return nil
}

func (m *ClockGatingManager) Close() error {
	if err := m.check(); err != nil {
		return err
	}
	m.graph, m.care, m.useful, m.candidates, m.gates = nil, nil, nil, nil, nil
	m.limits.Stop = nil
	m.closed = true
	return nil
}

func cloneClockGates(gates [][]AigLit) [][]AigLit {
	result := make([][]AigLit, len(gates))
	for i := range gates {
		result[i] = slices.Clone(gates[i])
	}
	return result
}

func cgtSignal(graph *AigManager, id AigID) bool {
	return uint64(id) < uint64(len(graph.objects)) && graph.objects[id].Type != AigNone && graph.objects[id].Type != AigCO
}

// DetectClockGateCandidates implements Cgt_ManDetectCandidates in recursive
// fanin-zero-first postorder.
func DetectClockGateCandidates(graph *AigManager, root AigLit, useful []bool, levelMax, maxCandidates int, stop func() bool) ([]AigLit, error) {
	if graph == nil || !graph.validSignal(root) || levelMax < 0 || maxCandidates < 0 || useful != nil && len(useful) != len(graph.objects) {
		return nil, ErrDimension
	}
	if maxCandidates == 0 {
		maxCandidates = 1 << 20
	}
	if graph.objects[root.ID()].Type != AigAND {
		return []AigLit{}, nil
	}
	seen := make([]bool, len(graph.objects))
	result := make([]AigLit, 0)
	var visit func(AigID) error
	visit = func(id AigID) error {
		if stop != nil && stop() {
			return ErrCancelled
		}
		if seen[id] {
			return nil
		}
		seen[id] = true
		object := graph.objects[id]
		if object.Type == AigAND {
			if err := visit(object.Fanin0.ID()); err != nil {
				return err
			}
			if err := visit(object.Fanin1.ID()); err != nil {
				return err
			}
		}
		if int(object.Level) <= levelMax && (useful == nil || useful[id]) {
			if len(result) == maxCandidates {
				return ErrLimit
			}
			result = append(result, AigLit(id)<<1)
		}
		return nil
	}
	if err := visit(root.ID()); err != nil {
		return nil, err
	}
	return result, nil
}

// DetectClockGateFanout implements the bounded fanout frontier used for ODCs.
func DetectClockGateFanout(graph *AigManager, root AigID, levelMax int, stop func() bool) ([]AigID, error) {
	if graph == nil || !cgtSignal(graph, root) || levelMax < 0 {
		return nil, ErrDimension
	}
	index, err := BuildAigFanoutIndex(graph, AigTraversalOptions{Stop: stop})
	if err != nil {
		return nil, err
	}
	seen := make([]bool, len(graph.objects))
	var included []AigID
	var visit func(AigID) error
	visit = func(id AigID) error {
		if stop != nil && stop() {
			return ErrCancelled
		}
		if seen[id] || graph.objects[id].Type == AigCO || int(graph.objects[id].Level) > levelMax {
			return nil
		}
		seen[id] = true
		included = append(included, id)
		fanouts, e := index.Fanouts(id)
		if e != nil {
			return e
		}
		for _, fanout := range fanouts {
			if e = visit(fanout.Object); e != nil {
				return e
			}
		}
		return nil
	}
	if err = visit(root); err != nil {
		return nil, err
	}
	frontier := make([]AigID, 0, len(included))
	for _, id := range included {
		fanouts, e := index.Fanouts(id)
		if e != nil {
			return nil, e
		}
		all := len(fanouts) > 0
		for _, fanout := range fanouts {
			if !seen[fanout.Object] {
				all = false
				break
			}
		}
		if !all {
			frontier = append(frontier, id)
		}
	}
	slices.Sort(frontier)
	return frontier, nil
}

// CollectClockGateVisited returns the AND fanin cone in topological order.
func CollectClockGateVisited(graph *AigManager, roots []AigID, stop func() bool) ([]AigID, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	seen := make([]bool, len(graph.objects))
	result := make([]AigID, 0)
	var visit func(AigID) error
	visit = func(id AigID) error {
		if stop != nil && stop() {
			return ErrCancelled
		}
		if uint64(id) >= uint64(len(graph.objects)) || graph.objects[id].Type == AigNone {
			return ErrInvalidLiteral
		}
		if seen[id] || graph.objects[id].Type == AigCI || id == 0 {
			return nil
		}
		object := graph.objects[id]
		if object.Type != AigAND {
			return ErrDimension
		}
		seen[id] = true
		if err := visit(object.Fanin0.ID()); err != nil {
			return err
		}
		if err := visit(object.Fanin1.ID()); err != nil {
			return err
		}
		result = append(result, id)
		return nil
	}
	for _, root := range roots {
		if err := visit(root); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func cgtCopyLogic(source, target *AigManager, inputMap []AigLit, maxObjects int, stop func() bool) ([]AigLit, error) {
	if len(inputMap) != len(source.cis) {
		return nil, ErrDimension
	}
	mapping := make([]AigLit, len(source.objects))
	for i := range mapping {
		mapping[i] = InvalidAigLit
	}
	mapping[0] = AigConst1
	for i, id := range source.cis {
		mapping[id] = inputMap[i]
	}
	for id, object := range source.objects {
		if stop != nil && stop() {
			return nil, ErrCancelled
		}
		var result AigLit
		var err error
		read := func(l AigLit) AigLit { return mapping[l.ID()].NotIf(l.Complemented()) }
		switch object.Type {
		case AigAND:
			result, err = target.And(read(object.Fanin0), read(object.Fanin1))
		case AigBuffer:
			result = read(object.Fanin0)
		default:
			continue
		}
		if err != nil {
			return nil, err
		}
		mapping[id] = result
		if target.ObjectCount() > maxObjects {
			return nil, ErrLimit
		}
	}
	return mapping, nil
}

// DeriveClockGateMiterAIG implements Cgt_ManDeriveAigForGating without using
// borrowed pData scratch fields. Outputs are one hold-vs-update XOR per latch.
func DeriveClockGateMiterAIG(source *AigManager, limits ClockGatingLimits) (*AigManager, error) {
	limits, err := limits.normalized()
	if err != nil || source == nil || source.RegisterCount() == 0 {
		return nil, ErrDimension
	}
	if err = source.Check(); err != nil {
		return nil, err
	}
	target := NewAigManager()
	inputs := make([]AigLit, source.InputCount())
	for i := range inputs {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		inputs[i], err = target.CreateCI()
		if err != nil {
			return nil, err
		}
	}
	mapping, err := cgtCopyLogic(source, target, inputs, limits.MaxObjects, limits.Stop)
	if err != nil {
		return nil, err
	}
	pi, po, registers := source.PrimaryInputCount(), source.PrimaryOutputCount(), source.RegisterCount()
	for i := 0; i < registers; i++ {
		lo := inputs[pi+i]
		co := source.objects[source.cos[po+i]]
		li := mapping[co.Fanin0.ID()].NotIf(co.Fanin0.Complemented())
		miter, e := target.Xor(lo, li)
		if e != nil {
			return nil, e
		}
		if _, e = target.CreateCO(miter); e != nil {
			return nil, e
		}
		if target.ObjectCount() > limits.MaxObjects {
			return nil, ErrLimit
		}
	}
	target.Name = "CG_miter"
	return target, nil
}

func cgtSolver(graph *AigManager, limits ClockGatingLimits) (*SatSolver3, *AigCNF, error) {
	cnf, err := EncodeAigCNF(graph)
	if err != nil {
		return nil, nil, err
	}
	if cnf.VariableCount() > limits.MaxSATVariables {
		return nil, nil, ErrLimit
	}
	solver, err := NewSatSolver3(SatSolver3Options{MaxVariables: limits.MaxSATVariables})
	if err != nil {
		return nil, nil, err
	}
	if err = solver.SetVariableCount(cnf.VariableCount()); err != nil {
		return nil, nil, err
	}
	if ok, addErr := solver.AddClauses(cnf.Clauses()); addErr != nil || !ok {
		if addErr != nil {
			return nil, nil, addErr
		}
		return nil, nil, ErrInvariant
	}
	return solver, cnf, nil
}

func cgtODCCare(work, source *AigManager, latchOutput AigID, levelMax int, limits ClockGatingLimits) (AigLit, error) {
	frontier, err := DetectClockGateFanout(source, latchOutput, levelMax, limits.Stop)
	if err != nil {
		return InvalidAigLit, err
	}
	zero := make([]AigLit, len(source.objects))
	one := make([]AigLit, len(source.objects))
	for i := range zero {
		zero[i], one[i] = InvalidAigLit, InvalidAigLit
	}
	zero[0], one[0] = AigConst1, AigConst1
	pi, po, registers := source.PrimaryInputCount(), source.PrimaryOutputCount(), source.RegisterCount()
	// The ODC cone is the source's second frame: primary inputs are fresh and
	// non-target latch outputs are driven by their first-frame next-state logic.
	for i := 0; i < registers; i++ {
		lo := source.cis[pi+i]
		co := source.objects[source.cos[po+i]]
		literal := co.Fanin0
		zero[lo], one[lo] = literal, literal
	}
	zero[latchOutput], one[latchOutput] = AigConst0, AigConst1
	visited, err := CollectClockGateVisited(source, frontier, limits.Stop)
	if err != nil {
		return InvalidAigLit, err
	}
	isPrimary := make([]bool, len(source.objects))
	for i := 0; i < pi; i++ {
		isPrimary[source.cis[i]] = true
	}
	for _, id := range visited {
		object := source.objects[id]
		for _, fanin := range []AigLit{object.Fanin0, object.Fanin1} {
			if !isPrimary[fanin.ID()] || zero[fanin.ID()] != InvalidAigLit {
				continue
			}
			literal, createErr := work.CreateCI()
			if createErr != nil {
				return InvalidAigLit, createErr
			}
			zero[fanin.ID()], one[fanin.ID()] = literal, literal
		}
	}
	read := func(mapping []AigLit, literal AigLit) AigLit {
		return mapping[literal.ID()].NotIf(literal.Complemented())
	}
	for _, id := range visited {
		if limits.Stop != nil && limits.Stop() {
			return InvalidAigLit, ErrCancelled
		}
		object := source.objects[id]
		zero[id], err = work.And(read(zero, object.Fanin0), read(zero, object.Fanin1))
		if err == nil {
			one[id], err = work.And(read(one, object.Fanin0), read(one, object.Fanin1))
		}
		if err != nil {
			return InvalidAigLit, err
		}
		if work.ObjectCount() > limits.MaxObjects {
			return InvalidAigLit, ErrLimit
		}
	}
	care := AigConst0
	for _, id := range frontier {
		difference, xorErr := work.Xor(zero[id], one[id])
		if xorErr != nil {
			return InvalidAigLit, xorErr
		}
		care, err = work.Or(care, difference)
		if err != nil {
			return InvalidAigLit, err
		}
		if work.ObjectCount() > limits.MaxObjects {
			return InvalidAigLit, ErrLimit
		}
	}
	return care, nil
}

func (m *ClockGatingManager) deriveCandidatesForRegister(register int) ([]AigLit, error) {
	graph := m.graph
	pi, po := graph.PrimaryInputCount(), graph.PrimaryOutputCount()
	co := graph.objects[graph.cos[po+register]]
	driver := co.Fanin0
	// nCandMax is present in Cgt_Par_t but is not consulted by the source's
	// active candidate traversal. MaxCandidates is the Go resource bound.
	candidates, err := DetectClockGateCandidates(graph, driver, m.useful, m.parameters.LevelMax, m.limits.MaxCandidates, m.limits.Stop)
	if err != nil {
		return nil, err
	}
	work := graph.Clone()
	lo := AigLit(graph.cis[pi+register]) << 1
	miter, err := work.Xor(lo, driver)
	if err != nil {
		return nil, err
	}
	if m.parameters.ODCMax > 0 {
		care, careErr := cgtODCCare(work, graph, graph.cis[pi+register], m.parameters.ODCMax, m.limits)
		if careErr != nil {
			return nil, careErr
		}
		miter, err = work.And(miter, care)
		if err != nil {
			return nil, err
		}
	}
	var careLiterals []AigLit
	if m.care != nil {
		sharedInputs := make([]AigLit, len(m.care.cis))
		for i, id := range work.cis[:len(m.care.cis)] {
			sharedInputs[i] = AigLit(id) << 1
		}
		careMap, copyErr := cgtCopyLogic(m.care, work, sharedInputs, m.limits.MaxObjects, m.limits.Stop)
		if copyErr != nil {
			return nil, copyErr
		}
		careLiterals = make([]AigLit, len(m.care.cos))
		for i, id := range m.care.cos {
			output := m.care.objects[id].Fanin0
			careLiterals[i] = careMap[output.ID()].NotIf(output.Complemented())
		}
	}
	solver, cnf, err := cgtSolver(work, m.limits)
	if err != nil {
		return nil, err
	}
	defer solver.Stop()
	careUnsatisfiable := false
	for _, careLiteral := range careLiterals {
		literal, literalErr := cnf.ObjectLiteral(careLiteral.ID())
		if literalErr != nil {
			return nil, literalErr
		}
		if careLiteral.Complemented() {
			literal = literal.Not()
		}
		if ok, addErr := solver.AddClause([]SatLit{literal}); addErr != nil || !ok {
			if addErr != nil {
				return nil, addErr
			}
			careUnsatisfiable = true
			break
		}
	}
	if careUnsatisfiable {
		valid := make([]AigLit, 0, len(candidates))
		for _, candidate := range candidates {
			if m.stats.Calls >= m.limits.MaxCalls {
				return nil, ErrLimit
			}
			// The source tries direct polarity first and continues to the next
			// candidate immediately after an UNSAT result.
			valid = append(valid, candidate)
			m.stats.Calls++
			m.stats.Unsatisfiable++
		}
		return valid, nil
	}
	miterSAT, err := cnf.ObjectLiteral(miter.ID())
	if err != nil {
		return nil, err
	}
	if miter.Complemented() {
		miterSAT = miterSAT.Not()
	}
	// The source implication checker receives a regular partition-output
	// variable. The simple encoder aliases COs, so materialize that variable
	// when the XOR is represented by a complemented AIG edge.
	if miterSAT.Negative() {
		variable, addErr := solver.AddVariable()
		if addErr != nil {
			return nil, addErr
		}
		regular := SatLit(variable) << 1
		if ok, addErr := solver.AddClauses([][]SatLit{{regular.Not(), miterSAT}, {regular, miterSAT.Not()}}); addErr != nil || !ok {
			if addErr != nil {
				return nil, addErr
			}
			return nil, ErrInvariant
		}
		miterSAT = regular
	}
	remainingCalls := m.limits.MaxCalls - m.stats.Calls
	if remainingCalls <= 0 {
		return nil, ErrLimit
	}
	session, err := NewClockGateImplicationSession(solver, ClockGateImplicationOptions{ConflictLimit: uint64(m.parameters.ConflictMax), MaxCalls: uint64(remainingCalls), Stop: m.limits.Stop})
	if err != nil {
		return nil, err
	}
	defer session.Close()
	valid := make([]AigLit, 0)
candidateLoop:
	for _, candidate := range candidates {
		for polarity := 0; polarity < 2; polarity++ {
			gate := candidate.NotIf(polarity != 0)
			gateSAT, e := cnf.ObjectLiteral(gate.ID())
			if e != nil {
				return nil, e
			}
			if gate.Complemented() {
				gateSAT = gateSAT.Not()
			}
			answer, e := session.Check(gateSAT, miterSAT)
			if e != nil {
				return nil, e
			}
			m.stats.Calls++
			switch answer.SAT.Status {
			case SatUnsatisfiable:
				m.stats.Unsatisfiable++
				valid = append(valid, gate)
				if polarity == 0 {
					// Exact Cgt_ClockGatingRangeCheck ordering: a successful
					// direct polarity suppresses the reverse-polarity query.
					continue candidateLoop
				}
			case SatSatisfiable:
				m.stats.Satisfiable++
			case SatUnknown:
				m.stats.Undecided++
			}
		}
	}
	return valid, nil
}

// FindCandidates implements the complete Cgt_ClockGatingCandidates flow.
func (m *ClockGatingManager) FindCandidates() ([][]AigLit, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	staged := make([][]AigLit, m.graph.RegisterCount())
	total := 0
	for register := range staged {
		if m.stopped() {
			return nil, ErrCancelled
		}
		one, err := m.deriveCandidatesForRegister(register)
		if err != nil {
			return nil, err
		}
		total += len(one)
		if total > m.limits.MaxCandidates {
			return nil, ErrLimit
		}
		staged[register] = one
	}
	m.candidates = staged
	m.stats.Recycles++
	return cloneClockGates(staged), nil
}

func cgtValues(graph *AigManager, inputs []bool) ([]bool, error) {
	if len(inputs) != len(graph.cis) {
		return nil, ErrDimension
	}
	values := make([]bool, len(graph.objects))
	values[0] = true
	for i, id := range graph.cis {
		values[id] = inputs[i]
	}
	read := func(l AigLit) bool { return values[l.ID()] != l.Complemented() }
	for id, object := range graph.objects {
		switch object.Type {
		case AigAND:
			values[id] = read(object.Fanin0) && read(object.Fanin1)
		case AigBuffer, AigCO:
			values[id] = read(object.Fanin0)
		}
	}
	return values, nil
}

func cgtGateScore(graph *AigManager, gate AigLit, stop func() bool) (int, error) {
	variables := graph.InputCount()
	samples := 1 << min(variables, 12)
	inputs := make([]bool, variables)
	score := 0
	state := uint64(0x9e3779b97f4a7c15)
	for assignment := 0; assignment < samples; assignment++ {
		if stop != nil && stop() {
			return 0, ErrCancelled
		}
		for i := range inputs {
			if variables <= 12 {
				inputs[i] = assignment&(1<<i) != 0
			} else {
				state ^= state << 7
				state ^= state >> 9
				inputs[i] = state&1 != 0
			}
		}
		values, err := cgtValues(graph, inputs)
		if err != nil {
			return 0, err
		}
		if values[gate.ID()] != gate.Complemented() {
			score++
		}
	}
	return score, nil
}

// DecideSimple selects the candidate with greatest deterministic activity.
func (m *ClockGatingManager) DecideSimple() ([][]AigLit, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	selected := make([][]AigLit, len(m.candidates))
	for register, candidates := range m.candidates {
		bestScore := 0
		var best AigLit
		found := false
		for _, candidate := range candidates {
			score, err := cgtGateScore(m.graph, candidate, m.limits.Stop)
			if err != nil {
				return nil, err
			}
			if score > bestScore {
				bestScore, best, found = score, candidate, true
			}
		}
		if found {
			selected[register] = []AigLit{best}
		}
	}
	m.gates = selected
	return cloneClockGates(selected), nil
}

// CompleteClockGates implements Cgt_ManCompleteGates. A signed candidate is
// complete when every reachable CO is a latch input containing that same
// signed candidate and no reachable primary output exists.
func CompleteClockGates(graph *AigManager, candidates [][]AigLit, stop func() bool) ([]AigLit, error) {
	if graph == nil || len(candidates) != graph.RegisterCount() {
		return nil, ErrDimension
	}
	index, err := BuildAigFanoutIndex(graph, AigTraversalOptions{Stop: stop})
	if err != nil {
		return nil, err
	}
	unique := make(map[AigID]bool)
	result := make([]AigLit, 0)
	po := graph.PrimaryOutputCount()
	for _, one := range candidates {
		for _, gate := range one {
			if unique[gate.ID()] {
				continue
			}
			unique[gate.ID()] = true
			seen := make([]bool, len(graph.objects))
			stack := []AigID{gate.ID()}
			complete := true
			for len(stack) != 0 && complete {
				if stop != nil && stop() {
					return nil, ErrCancelled
				}
				id := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if seen[id] {
					continue
				}
				seen[id] = true
				fanouts, fanoutErr := index.Fanouts(id)
				if fanoutErr != nil {
					return nil, fanoutErr
				}
				for _, fanout := range fanouts {
					object := graph.objects[fanout.Object]
					if object.Type != AigCO {
						stack = append(stack, fanout.Object)
						continue
					}
					coIndex := slices.Index(graph.cos, fanout.Object)
					if coIndex < po || coIndex < 0 || !slices.Contains(candidates[coIndex-po], gate) {
						complete = false
						break
					}
				}
			}
			if complete {
				result = append(result, gate)
			}
		}
	}
	return result, nil
}

// DecideArea retains gates common to every latch-input fanout they reach.
func (m *ClockGatingManager) DecideArea() ([][]AigLit, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	complete, err := CompleteClockGates(m.graph, m.candidates, m.limits.Stop)
	if err != nil {
		return nil, err
	}
	selected := make([][]AigLit, len(m.candidates))
	for register, one := range m.candidates {
		for _, gate := range one {
			if slices.Contains(complete, gate) {
				selected[register] = append(selected[register], gate)
			}
		}
	}
	m.gates = selected
	return cloneClockGates(selected), nil
}

// ClockGateCoverage returns the percentage of sampled gate-high transitions.
func ClockGateCoverage(graph *AigManager, gates [][]AigLit, stop func() bool) (float64, error) {
	if graph == nil || len(gates) != graph.RegisterCount() || len(gates) == 0 {
		return 0, ErrDimension
	}
	total := 0
	for _, one := range gates {
		for _, gate := range one {
			score, err := cgtGateScore(graph, gate, stop)
			if err != nil {
				return 0, err
			}
			total += score
		}
	}
	samples := 1 << min(graph.InputCount(), 12)
	return 100 * float64(total) / float64(samples*len(gates)), nil
}

// BuildClockGate implements Cgt_ManBuildClockGate as an ordered OR reduction.
func BuildClockGate(graph *AigManager, gates []AigLit, maxObjects int, stop func() bool) (AigLit, error) {
	if graph == nil || len(gates) == 0 || maxObjects < 0 {
		return InvalidAigLit, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	total := AigConst0
	for _, gate := range gates {
		if stop != nil && stop() {
			return InvalidAigLit, ErrCancelled
		}
		if !graph.validSignal(gate) {
			return InvalidAigLit, ErrInvalidLiteral
		}
		var err error
		total, err = graph.Or(total, gate)
		if err != nil {
			return InvalidAigLit, err
		}
		if graph.ObjectCount() > maxObjects {
			return InvalidAigLit, ErrLimit
		}
	}
	return total, nil
}

func cgtCopyReducedLogic(source, target *AigManager, inputs []AigLit, polarity []uint8, limits ClockGatingLimits) ([]AigLit, []AigLit, error) {
	mapping := make([]AigLit, len(source.objects))
	original := make([]AigLit, len(source.objects))
	for i := range mapping {
		mapping[i], original[i] = InvalidAigLit, InvalidAigLit
	}
	mapping[0], original[0] = AigConst1, AigConst1
	for i, id := range source.cis {
		mapping[id], original[id] = inputs[i], inputs[i]
	}
	for id, object := range source.objects {
		if limits.Stop != nil && limits.Stop() {
			return nil, nil, ErrCancelled
		}
		read := func(literal AigLit) AigLit { return mapping[literal.ID()].NotIf(literal.Complemented()) }
		var literal AigLit
		var err error
		switch object.Type {
		case AigAND:
			literal, err = target.And(read(object.Fanin0), read(object.Fanin1))
		case AigBuffer:
			literal = read(object.Fanin0)
		default:
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		original[id] = literal
		switch polarity[id] {
		case 1:
			mapping[id] = AigConst0
		case 2:
			mapping[id] = AigConst1
		default:
			mapping[id] = literal
		}
		if target.ObjectCount() > limits.MaxObjects {
			return nil, nil, ErrLimit
		}
	}
	return mapping, original, nil
}

// DeriveClockGatedAIG implements the non-reducing Cgt_ManDeriveGatedAig path.
// Construction is transactional: errors discard the new graph and never
// mutate source.
func DeriveClockGatedAIG(source *AigManager, gates [][]AigLit, limits ClockGatingLimits) (*AigManager, int, error) {
	return deriveClockGatedAIG(source, gates, false, limits)
}

// DeriveClockGatedAIGReduced implements fReduce: selected positive gate nodes
// are replaced by zero in ordinary logic, selected negative gate nodes by one,
// while the clock-gate OR uses their unreduced copies.
func DeriveClockGatedAIGReduced(source *AigManager, gates [][]AigLit, limits ClockGatingLimits) (*AigManager, int, error) {
	return deriveClockGatedAIG(source, gates, true, limits)
}

func deriveClockGatedAIG(source *AigManager, gates [][]AigLit, reduce bool, limits ClockGatingLimits) (*AigManager, int, error) {
	limits, err := limits.normalized()
	if err != nil || source == nil || len(gates) != source.RegisterCount() {
		return nil, 0, ErrDimension
	}
	if err = source.Check(); err != nil {
		return nil, 0, err
	}
	target := NewAigManager()
	target.Name, target.Spec = source.Name, source.Spec
	inputs := make([]AigLit, source.InputCount())
	for i := range inputs {
		if limits.Stop != nil && limits.Stop() {
			return nil, 0, ErrCancelled
		}
		inputs[i], err = target.CreateCI()
		if err != nil {
			return nil, 0, err
		}
	}
	polarity := make([]uint8, len(source.objects))
	if reduce {
		for _, one := range gates {
			for _, gate := range one {
				if !source.validSignal(gate) {
					return nil, 0, ErrInvalidLiteral
				}
				value := uint8(1)
				if gate.Complemented() {
					value = 2
				}
				if polarity[gate.ID()] != 0 && polarity[gate.ID()] != value {
					return nil, 0, ErrInvariant
				}
				polarity[gate.ID()] = value
			}
		}
	}
	var mapping, original []AigLit
	if reduce {
		mapping, original, err = cgtCopyReducedLogic(source, target, inputs, polarity, limits)
	} else {
		mapping, err = cgtCopyLogic(source, target, inputs, limits.MaxObjects, limits.Stop)
		original = mapping
	}
	if err != nil {
		return nil, 0, err
	}
	usedNodes := target.AndCount()
	po, pi, registers := source.PrimaryOutputCount(), source.PrimaryInputCount(), source.RegisterCount()
	for i := 0; i < po; i++ {
		co := source.objects[source.cos[i]]
		driver := mapping[co.Fanin0.ID()].NotIf(co.Fanin0.Complemented())
		if _, err = target.CreateCO(driver); err != nil {
			return nil, 0, err
		}
	}
	for i := 0; i < registers; i++ {
		co := source.objects[source.cos[po+i]]
		driver := mapping[co.Fanin0.ID()].NotIf(co.Fanin0.Complemented())
		if len(gates[i]) > 0 {
			mappedGates := make([]AigLit, len(gates[i]))
			for j, gate := range gates[i] {
				if !source.validSignal(gate) {
					return nil, 0, ErrInvalidLiteral
				}
				base := mapping[gate.ID()]
				if reduce && polarity[gate.ID()] != 0 {
					base = original[gate.ID()]
				}
				mappedGates[j] = base.NotIf(gate.Complemented())
			}
			gate, e := BuildClockGate(target, mappedGates, limits.MaxObjects, limits.Stop)
			if e != nil {
				return nil, 0, e
			}
			driver, err = target.Mux(gate, inputs[pi+i], driver)
			if err != nil {
				return nil, 0, err
			}
		}
		if _, err = target.CreateCO(driver); err != nil {
			return nil, 0, err
		}
		if target.ObjectCount() > limits.MaxObjects {
			return nil, 0, ErrLimit
		}
	}
	if err = target.SetRegisterCount(registers); err != nil {
		return nil, 0, err
	}
	if err = target.SetConstraintCount(source.ConstraintCount()); err != nil {
		return nil, 0, err
	}
	target.Cleanup()
	if err = target.Check(); err != nil {
		return nil, 0, err
	}
	return target, usedNodes, nil
}

// ClockGateAIG runs candidate generation, decision, and gated-AIG derivation.
func ClockGateAIG(source, care *AigManager, parameters ClockGatingParameters, useful []bool, limits ClockGatingLimits) (*AigManager, [][]AigLit, ClockGatingStats, error) {
	manager, err := NewClockGatingManager(source, care, parameters, limits)
	if err != nil {
		return nil, nil, ClockGatingStats{}, err
	}
	defer manager.Close()
	if err = manager.SetUseful(useful); err != nil {
		return nil, nil, ClockGatingStats{}, err
	}
	if _, err = manager.FindCandidates(); err != nil {
		return nil, nil, ClockGatingStats{}, err
	}
	var gates [][]AigLit
	if manager.parameters.AreaOnly {
		gates, err = manager.DecideArea()
	} else {
		gates, err = manager.DecideSimple()
	}
	if err != nil {
		return nil, nil, ClockGatingStats{}, err
	}
	gated, _, err := deriveClockGatedAIG(source, gates, manager.parameters.AreaOnly, limits)
	return gated, gates, manager.stats, err
}
