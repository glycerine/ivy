package abc

// Implication checking from src/opt/cgt/cgtSat.c.

import "slices"

// ClockGateImplicationOptions are the owned manager parameters used by
// Cgt_CheckImplication. Zero limits are unlimited. Stop is synchronous and
// must not reenter either the session or its solver.
type ClockGateImplicationOptions struct {
	ConflictLimit uint64
	MaxCalls      uint64
	Stop          func() bool
}

// ClockGateImplicationResult reports whether gate implies the negation of
// miter. SAT is the underlying check of gate AND miter. On an UNSAT answer the
// source installs the entailed binary clause (!gate OR !miter); ClauseInstalled
// says that the concrete solver needed to retain a new copy of that clause.
type ClockGateImplicationResult struct {
	Implication     bool
	ClauseInstalled bool
	SAT             SatResult
}

// ClockGateImplicationStats replaces the source manager's counters. Solver is
// the sum of the per-call SAT work, not the lifetime statistics of the shared
// solver.
type ClockGateImplicationStats struct {
	Calls, Satisfiable, Unsatisfiable, Undecided uint64
	ClausesInstalled                             uint64
	Solver                                       SatStats
}

// ClockGateImplicationSession is the concrete Cgt manager slice needed by
// cgtSat.c. SAT literals replace the source's AIG-object-to-CNF-variable table.
// The caller retains ownership of Solver; Close only releases this view.
type ClockGateImplicationSession struct {
	solver  *SatSolver3
	options ClockGateImplicationOptions
	stats   ClockGateImplicationStats
	closed  bool
}

func NewClockGateImplicationSession(solver *SatSolver3, options ClockGateImplicationOptions) (*ClockGateImplicationSession, error) {
	if solver == nil || !solver.usable() {
		return nil, ErrDimension
	}
	return &ClockGateImplicationSession{solver: solver, options: options}, nil
}

func (session *ClockGateImplicationSession) check() error {
	if session == nil || session.closed || session.solver == nil || !session.solver.usable() {
		return ErrInvariant
	}
	return nil
}

func clockGateAddSatStats(first *SatStats, second SatStats) {
	first.Conflicts += second.Conflicts
	first.Propagations += second.Propagations
	first.Decisions += second.Decisions
	first.Inspections += second.Inspections
	first.Restarts += second.Restarts
	first.Learnt += second.Learnt
	first.Deleted += second.Deleted
}

// Check tests gate => !miter by solving gate AND miter. Miter must be a
// positive literal, and gate must name a different variable, matching the
// source assertions that the miter is regular and the two AIG objects differ.
// Unknown results preserve the source's undecided outcome. Models, conflicts
// and clauses in the result are owned by the caller.
func (session *ClockGateImplicationSession) Check(gate, miter SatLit) (ClockGateImplicationResult, error) {
	var answer ClockGateImplicationResult
	if err := session.check(); err != nil {
		return answer, err
	}
	variables := session.solver.VariableCount()
	if !gate.ValidForVariables(variables) || !miter.ValidForVariables(variables) {
		return answer, ErrInvalidLiteral
	}
	if miter.Negative() || gate.Var() == miter.Var() {
		return answer, ErrDimension
	}
	if session.options.MaxCalls != 0 && session.stats.Calls >= session.options.MaxCalls {
		return answer, ErrLimit
	}

	result, err := session.solver.Solve([]SatLit{gate, miter}, SatSolver3Limits{
		Conflicts: session.options.ConflictLimit,
		Stop:      session.options.Stop,
	})
	if err != nil {
		return answer, err
	}
	result.Model = slices.Clone(result.Model)
	result.Conflict = slices.Clone(result.Conflict)
	answer.SAT = result

	// Commit counters only after the solver call succeeds. This keeps rejected
	// inputs and API errors invisible to the manager statistics.
	session.stats.Calls++
	clockGateAddSatStats(&session.stats.Solver, result.Stats)
	switch result.Status {
	case SatUnsatisfiable:
		clause := []SatLit{gate.Not(), miter.Not()}
		beforeClauses := session.solver.OriginalClauseCount()
		consistent, addErr := session.solver.AddClause(clause)
		if addErr != nil {
			return ClockGateImplicationResult{}, addErr
		}
		installed := session.solver.OriginalClauseCount() > beforeClauses
		if consistent {
			if _, compressErr := session.solver.Compress(); compressErr != nil {
				return ClockGateImplicationResult{}, compressErr
			}
		}
		if installed {
			session.stats.ClausesInstalled++
		}
		session.stats.Unsatisfiable++
		answer.Implication = true
		answer.ClauseInstalled = installed
	case SatSatisfiable:
		session.stats.Satisfiable++
	case SatUnknown:
		session.stats.Undecided++
	default:
		return ClockGateImplicationResult{}, ErrInvariant
	}
	return answer, nil
}

func (session *ClockGateImplicationSession) Stats() (ClockGateImplicationStats, error) {
	if err := session.check(); err != nil {
		return ClockGateImplicationStats{}, err
	}
	return session.stats, nil
}

func (session *ClockGateImplicationSession) Close() error {
	if err := session.check(); err != nil {
		return err
	}
	session.solver = nil
	session.options.Stop = nil
	session.closed = true
	return nil
}
