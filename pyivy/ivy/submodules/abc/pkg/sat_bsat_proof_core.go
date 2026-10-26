package abc

// Port of src/sat/bsat/satInterP.c.

import (
	"fmt"
	"slices"
	"strings"
)

type SatProofCoreLimits struct {
	MaxVariables, MaxClauses, MaxLiterals       int
	MaxResolutionSteps, MaxAntecedentReferences int
	Stop                                        func() bool
}

type SatProofCoreStats struct {
	Variables, Roots, Clauses, RootAssignments int
	ResolutionSteps, AntecedentReferences      int
	CoreSize                                   int
	Stopped                                    bool
}

type SatBMCVariable struct {
	Object, Frame int
}

type satProofCoreClause struct {
	literals []SatLit
	root     bool
}

type satProofCoreRun struct {
	manager          *SatProofCoreManager
	clauses          []satProofCoreClause
	assigns          []SatLit
	assigned         []bool
	reasons          []int
	seen             []bool
	trail            []SatLit
	rootSize         int
	watches          [][]int
	proofNumbers     []int
	antecedents      [][]int
	antecedentSet    []bool
	interpolation    *satTruthInterpolationRun
	aigInterpolation *satAigInterpolationRun
	counter          int
	empty            int
	work             int
}

type SatProofCoreManager struct {
	limits  SatProofCoreLimits
	stats   SatProofCoreStats
	busy    bool
	stopped bool
}

func checkSatProofCoreLimits(limits SatProofCoreLimits) (SatProofCoreLimits, error) {
	if limits.MaxVariables < 0 || limits.MaxClauses < 0 || limits.MaxLiterals < 0 || limits.MaxResolutionSteps < 0 || limits.MaxAntecedentReferences < 0 {
		return limits, ErrDimension
	}
	if limits.MaxVariables == 0 {
		limits.MaxVariables = 1 << 20
	}
	if limits.MaxClauses == 0 {
		limits.MaxClauses = 16 << 20
	}
	if limits.MaxLiterals == 0 {
		limits.MaxLiterals = 64 << 20
	}
	if limits.MaxResolutionSteps == 0 {
		limits.MaxResolutionSteps = 64 << 20
	}
	if limits.MaxAntecedentReferences == 0 {
		limits.MaxAntecedentReferences = 64 << 20
	}
	return limits, nil
}

func NewSatProofCoreManager(limits SatProofCoreLimits) (*SatProofCoreManager, error) {
	checked, err := checkSatProofCoreLimits(limits)
	if err != nil {
		return nil, err
	}
	return &SatProofCoreManager{limits: checked}, nil
}

func (m *SatProofCoreManager) tick() error {
	if m.limits.Stop != nil && m.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

func (m *SatProofCoreManager) newRun(store *SatProofStore) (*satProofCoreRun, error) {
	if m == nil || m.stopped || m.busy || store == nil || store.stopped {
		return nil, ErrInvariant
	}
	stats := store.Stats()
	if stats.Variables <= 0 || stats.Clauses <= 0 || stats.Roots < 0 || stats.Roots > stats.Clauses ||
		stats.Variables > m.limits.MaxVariables || stats.Clauses > m.limits.MaxClauses || stats.Literals > m.limits.MaxLiterals {
		if stats.Variables > m.limits.MaxVariables || stats.Clauses > m.limits.MaxClauses || stats.Literals > m.limits.MaxLiterals {
			return nil, ErrLimit
		}
		return nil, ErrInvariant
	}
	if stats.EmptyClause != stats.Clauses-1 || stats.EmptyClause < stats.Roots {
		return nil, ErrInvariant
	}
	run := &satProofCoreRun{
		manager:       m,
		clauses:       make([]satProofCoreClause, stats.Clauses),
		assigns:       make([]SatLit, stats.Variables),
		assigned:      make([]bool, stats.Variables),
		reasons:       make([]int, stats.Variables),
		seen:          make([]bool, stats.Variables),
		trail:         make([]SatLit, 0, stats.Variables),
		watches:       make([][]int, 2*stats.Variables),
		proofNumbers:  make([]int, stats.Clauses),
		antecedents:   make([][]int, stats.Clauses-stats.Roots),
		antecedentSet: make([]bool, stats.Clauses-stats.Roots),
		empty:         stats.EmptyClause,
	}
	for i := range run.reasons {
		run.reasons[i] = -1
	}
	for i, clause := range store.clauses {
		if err := m.tick(); err != nil {
			return nil, err
		}
		if clause.ID != i || clause.Root != (i < stats.Roots) || clause.PartA != (i < stats.ClausesA) {
			return nil, ErrInvariant
		}
		if i != run.empty && len(clause.Literals) == 0 {
			return nil, ErrInvariant
		}
		for j, literal := range clause.Literals {
			if !literal.ValidForVariables(stats.Variables) || (j > 0 && clause.Literals[j-1] >= literal) {
				return nil, ErrInvariant
			}
		}
		run.clauses[i] = satProofCoreClause{slices.Clone(clause.Literals), clause.Root}
	}
	m.stats = SatProofCoreStats{Variables: stats.Variables, Roots: stats.Roots, Clauses: stats.Clauses}
	return run, nil
}

func (r *satProofCoreRun) assignedTrue(literal SatLit) bool {
	variable := int(literal.Var())
	return r.assigned[variable] && r.assigns[variable] == literal
}

func (r *satProofCoreRun) assignedFalse(literal SatLit) bool {
	variable := int(literal.Var())
	return r.assigned[variable] && r.assigns[variable] == literal.Not()
}

func (r *satProofCoreRun) enqueue(literal SatLit, reason int) bool {
	variable := int(literal.Var())
	if r.assigned[variable] {
		return r.assigns[variable] == literal
	}
	r.assigned[variable] = true
	r.assigns[variable] = literal
	r.reasons[variable] = reason
	r.trail = append(r.trail, literal)
	return true
}

func (r *satProofCoreRun) cancelUntil(level int) {
	for i := len(r.trail) - 1; i >= level; i-- {
		variable := int(r.trail[i].Var())
		r.assigned[variable] = false
		r.reasons[variable] = -1
	}
	r.trail = r.trail[:level]
}

func (r *satProofCoreRun) watchClause(clauseID int, literal SatLit) error {
	clause := &r.clauses[clauseID]
	if len(clause.literals) < 2 || (clause.literals[0] != literal && clause.literals[1] != literal) {
		return ErrInvariant
	}
	trigger := int(literal.Not())
	r.watches[trigger] = append(r.watches[trigger], clauseID)
	return nil
}

func (r *satProofCoreRun) propagateOne(literal SatLit) (int, error) {
	trigger := int(literal)
	falseLiteral := literal.Not()
	for index := 0; index < len(r.watches[trigger]); {
		if err := r.manager.tick(); err != nil {
			return -1, err
		}
		clauseID := r.watches[trigger][index]
		clause := &r.clauses[clauseID]
		if len(clause.literals) < 2 {
			return -1, ErrInvariant
		}
		if clause.literals[0] == falseLiteral {
			clause.literals[0], clause.literals[1] = clause.literals[1], clause.literals[0]
		}
		if clause.literals[1] != falseLiteral {
			return -1, ErrInvariant
		}
		if r.assignedTrue(clause.literals[0]) {
			index++
			continue
		}
		moved := false
		for i := 2; i < len(clause.literals); i++ {
			if r.assignedFalse(clause.literals[i]) {
				continue
			}
			clause.literals[1], clause.literals[i] = clause.literals[i], clause.literals[1]
			copy(r.watches[trigger][index:], r.watches[trigger][index+1:])
			r.watches[trigger] = r.watches[trigger][:len(r.watches[trigger])-1]
			if err := r.watchClause(clauseID, clause.literals[1]); err != nil {
				return -1, err
			}
			moved = true
			break
		}
		if moved {
			continue
		}
		if !r.enqueue(clause.literals[0], clauseID) {
			return clauseID, nil
		}
		index++
	}
	return -1, nil
}

func (r *satProofCoreRun) propagate(start int) (int, error) {
	for i := start; i < len(r.trail); i++ {
		conflict, err := r.propagateOne(r.trail[i])
		if err != nil || conflict >= 0 {
			return conflict, err
		}
	}
	return -1, nil
}

func (r *satProofCoreRun) setAntecedents(clauseID int, antecedents []int) error {
	index := clauseID - r.manager.stats.Roots
	if index < 0 || index >= len(r.antecedents) || r.antecedentSet[index] {
		return ErrInvariant
	}
	if len(antecedents) > r.manager.limits.MaxAntecedentReferences-r.manager.stats.AntecedentReferences {
		return ErrLimit
	}
	r.antecedents[index] = slices.Clone(antecedents)
	r.antecedentSet[index] = true
	r.manager.stats.AntecedentReferences += len(antecedents)
	return nil
}

func (r *satProofCoreRun) trace(conflictID, finalID int) error {
	if conflictID < 0 || conflictID >= len(r.clauses) || finalID < r.manager.stats.Roots || finalID >= len(r.clauses) {
		return ErrInvariant
	}
	conflict := &r.clauses[conflictID]
	final := &r.clauses[finalID]
	if r.interpolation != nil {
		r.interpolation.copy(finalID, conflictID)
	}
	if r.aigInterpolation != nil {
		r.aigInterpolation.copy(finalID, conflictID)
	}
	resolvent := slices.Clone(conflict.literals)
	for _, literal := range conflict.literals {
		r.seen[literal.Var()] = true
	}
	antecedents := []int{conflictID}
	if r.proofNumbers[conflictID] <= 0 {
		return ErrInvariant
	}
	for i := len(r.trail) - 1; i >= 0; i-- {
		if err := r.manager.tick(); err != nil {
			return err
		}
		variable := int(r.trail[i].Var())
		if !r.seen[variable] {
			continue
		}
		r.seen[variable] = false
		reasonID := r.reasons[variable]
		if reasonID < 0 {
			continue
		}
		reason := &r.clauses[reasonID]
		if len(reason.literals) == 0 || reason.literals[0] != r.trail[i] || r.proofNumbers[reasonID] <= 0 {
			return ErrInvariant
		}
		for _, literal := range reason.literals[1:] {
			r.seen[literal.Var()] = true
		}
		if r.interpolation != nil {
			r.interpolation.resolve(finalID, reasonID, variable)
		}
		if r.aigInterpolation != nil {
			if err := r.aigInterpolation.resolve(finalID, reasonID, variable, reason.literals); err != nil {
				return err
			}
		}
		if r.manager.stats.ResolutionSteps == r.manager.limits.MaxResolutionSteps {
			return ErrLimit
		}
		r.counter++
		r.manager.stats.ResolutionSteps++
		pivot := -1
		for j, literal := range resolvent {
			if int(literal.Var()) == variable {
				pivot = j
				break
			}
		}
		if pivot < 0 || resolvent[pivot] != reason.literals[0].Not() {
			return ErrInvariant
		}
		resolvent = append(resolvent[:pivot], resolvent[pivot+1:]...)
		for _, literal := range reason.literals[1:] {
			found := -1
			for j, current := range resolvent {
				if current.Var() == literal.Var() {
					found = j
					break
				}
			}
			if found < 0 {
				resolvent = append(resolvent, literal)
			} else if resolvent[found] != literal {
				return ErrInvariant
			}
		}
		antecedents = append(antecedents, reasonID)
	}
	for _, literal := range resolvent {
		if !slices.Contains(final.literals, literal) {
			return ErrInvariant
		}
	}
	kept := final.literals[:0]
	for _, literal := range final.literals {
		if slices.Contains(resolvent, literal) {
			kept = append(kept, literal)
		}
	}
	final.literals = kept
	if len(final.literals) != len(resolvent) {
		return ErrInvariant
	}
	if err := r.setAntecedents(finalID, antecedents); err != nil {
		return err
	}
	r.proofNumbers[finalID] = r.counter
	return nil
}

func (r *satProofCoreRun) processRoots() (bool, error) {
	for clauseID := 0; clauseID < r.manager.stats.Roots; clauseID++ {
		clause := &r.clauses[clauseID]
		if len(clause.literals) > 1 {
			if err := r.watchClause(clauseID, clause.literals[0]); err != nil {
				return false, err
			}
			if err := r.watchClause(clauseID, clause.literals[1]); err != nil {
				return false, err
			}
		}
		if len(clause.literals) == 1 && !r.enqueue(clause.literals[0], clauseID) {
			if err := r.trace(clauseID, r.empty); err != nil {
				return false, err
			}
			return false, nil
		}
	}
	conflict, err := r.propagate(0)
	if err != nil {
		return false, err
	}
	if conflict >= 0 {
		if err := r.trace(conflict, r.empty); err != nil {
			return false, err
		}
		return false, nil
	}
	r.rootSize = len(r.trail)
	r.manager.stats.RootAssignments = r.rootSize
	return true, nil
}

func (r *satProofCoreRun) recordClause(clauseID int) (bool, error) {
	clause := &r.clauses[clauseID]
	if clause.root || len(clause.literals) == 0 || len(r.trail) != r.rootSize {
		return false, ErrInvariant
	}
	for _, literal := range clause.literals {
		if r.assignedTrue(literal) {
			return true, r.setAntecedents(clauseID, nil)
		}
	}
	for _, literal := range clause.literals {
		if !r.enqueue(literal.Not(), -1) {
			return false, ErrInvariant
		}
	}
	conflict, err := r.propagate(r.rootSize)
	if err != nil {
		return false, err
	}
	if conflict < 0 {
		return false, ErrInvariant
	}
	if len(clause.literals) >= len(r.clauses[conflict].literals) {
		subset := true
		for _, literal := range r.clauses[conflict].literals {
			if !slices.Contains(clause.literals, literal) {
				subset = false
				break
			}
		}
		if subset {
			r.cancelUntil(r.rootSize)
			return true, r.setAntecedents(clauseID, nil)
		}
	}
	if err := r.trace(conflict, clauseID); err != nil {
		return false, err
	}
	r.cancelUntil(r.rootSize)
	if len(clause.literals) > 1 {
		if err := r.watchClause(clauseID, clause.literals[0]); err != nil {
			return false, err
		}
		if err := r.watchClause(clauseID, clause.literals[1]); err != nil {
			return false, err
		}
		return true, nil
	}
	if len(clause.literals) != 1 || !r.enqueue(clause.literals[0], clauseID) {
		return false, ErrInvariant
	}
	conflict, err = r.propagate(r.rootSize)
	if err != nil {
		return false, err
	}
	if conflict >= 0 {
		if err := r.trace(conflict, r.empty); err != nil {
			return false, err
		}
		return false, nil
	}
	r.rootSize = len(r.trail)
	r.manager.stats.RootAssignments = r.rootSize
	return true, nil
}

func (r *satProofCoreRun) derive() error {
	for clauseID := 0; clauseID < r.manager.stats.Roots; clauseID++ {
		r.counter++
		r.proofNumbers[clauseID] = r.counter
	}
	keepGoing, err := r.processRoots()
	if err != nil {
		return err
	}
	for clauseID := r.manager.stats.Roots; keepGoing && clauseID < r.empty; clauseID++ {
		keepGoing, err = r.recordClause(clauseID)
		if err != nil {
			return err
		}
	}
	if !r.antecedentSet[r.empty-r.manager.stats.Roots] {
		return ErrInvariant
	}
	return nil
}

type satProofCoreFrame struct {
	clause, next int
	entered      bool
}

func (r *satProofCoreRun) core(includeLearned bool) ([]int, error) {
	visited := make([]bool, len(r.clauses))
	stack := []satProofCoreFrame{{clause: r.empty}}
	core := make([]int, 0)
	for len(stack) > 0 {
		if err := r.manager.tick(); err != nil {
			return nil, err
		}
		frame := &stack[len(stack)-1]
		if !frame.entered {
			if frame.clause < 0 || frame.clause >= len(r.clauses) {
				return nil, ErrInvariant
			}
			if visited[frame.clause] {
				stack = stack[:len(stack)-1]
				continue
			}
			visited[frame.clause] = true
			if frame.clause < r.manager.stats.Roots {
				if !includeLearned {
					core = append(core, frame.clause)
				}
				stack = stack[:len(stack)-1]
				continue
			}
			index := frame.clause - r.manager.stats.Roots
			if !r.antecedentSet[index] {
				return nil, ErrInvariant
			}
			frame.entered = true
		}
		antecedents := r.antecedents[frame.clause-r.manager.stats.Roots]
		if frame.next < len(antecedents) {
			child := antecedents[frame.next]
			frame.next++
			stack = append(stack, satProofCoreFrame{clause: child})
			continue
		}
		if includeLearned {
			core = append(core, frame.clause)
		}
		stack = stack[:len(stack)-1]
	}
	return core, nil
}

func (m *SatProofCoreManager) UnsatCore(store *SatProofStore, includeLearned bool) ([]int, error) {
	run, err := m.newRun(store)
	if err != nil {
		return nil, err
	}
	m.busy = true
	defer func() { m.busy = false }()
	if err = run.derive(); err != nil {
		return nil, err
	}
	core, err := run.core(includeLearned)
	if err != nil {
		return nil, err
	}
	m.stats.CoreSize = len(core)
	return core, nil
}

func VerifySatProofRootCore(store *SatProofStore, core []int) error {
	if store == nil || store.stopped {
		return ErrInvariant
	}
	stats := store.Stats()
	solver, err := NewSatSolverWithLimit(stats.Variables)
	if err != nil {
		return err
	}
	for _, clauseID := range core {
		if clauseID < 0 || clauseID >= stats.Roots {
			return ErrInvalidLiteral
		}
		clause := store.clauses[clauseID]
		ok, addErr := solver.AddClause(clause.Literals)
		if addErr != nil {
			return addErr
		}
		if !ok {
			return nil
		}
	}
	result, err := solver.Solve(nil, SatLimits{})
	if err != nil {
		return err
	}
	if result.Status != SatUnsatisfiable {
		return ErrInvariant
	}
	return nil
}

func FormatSatProofCoreClause(clause SatProofStoreClause, proofNumber int) (string, error) {
	if clause.ID < 0 || proofNumber < 0 {
		return "", ErrDimension
	}
	var result strings.Builder
	fmt.Fprintf(&result, "Clause ID = %d. Proof = %d. {", clause.ID, proofNumber)
	for _, literal := range clause.Literals {
		if !literal.Valid() {
			return "", ErrInvalidLiteral
		}
		fmt.Fprintf(&result, " %d", literal)
	}
	result.WriteString(" }\n")
	return result.String(), nil
}

func FormatSatProofCoreResolvent(literals []SatLit) (string, error) {
	var result strings.Builder
	result.WriteString("Resolvent: {")
	for _, literal := range literals {
		if !literal.Valid() {
			return "", ErrInvalidLiteral
		}
		fmt.Fprintf(&result, " %d", literal)
	}
	result.WriteString(" }\n")
	return result.String(), nil
}

func FormatSatProofCoreInterpolantClause(clauseID int) (string, error) {
	if clauseID < 0 {
		return "", ErrDimension
	}
	return fmt.Sprintf("Clause %2d :  \n", clauseID), nil
}

func FormatSatProofLearnedCoreForBMC(store *SatProofStore, core []int, variables []SatBMCVariable) (string, error) {
	if store == nil || store.stopped || len(variables) < store.variables {
		return "", ErrInvariant
	}
	var result strings.Builder
	fmt.Fprintf(&result, "UNSAT contains %d learned clauses:\n", len(core))
	for i, clauseID := range core {
		if clauseID < 0 || clauseID >= len(store.clauses) {
			return "", ErrInvalidLiteral
		}
		clause := store.clauses[clauseID]
		fmt.Fprintf(&result, "%6d : %6d : ", i, clauseID-store.roots)
		for _, literal := range clause.Literals {
			mapped := variables[literal.Var()]
			if literal.Negative() {
				result.WriteByte('!')
			}
			fmt.Fprintf(&result, "%d(%d) ", mapped.Object, mapped.Frame)
		}
		if len(clause.Literals) == 0 {
			result.WriteString("Empty")
		}
		result.WriteByte('\n')
	}
	return result.String(), nil
}

func (m *SatProofCoreManager) Stats() SatProofCoreStats {
	if m == nil {
		return SatProofCoreStats{Stopped: true}
	}
	stats := m.stats
	stats.Stopped = m.stopped
	return stats
}

func (m *SatProofCoreManager) Stop() {
	if m == nil || m.stopped || m.busy {
		return
	}
	m.stats = SatProofCoreStats{Stopped: true}
	m.stopped = true
}
