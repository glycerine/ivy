package abc

// Port of src/proof/pdr/pdrInv.c:Pdr_ManDeriveMarkNonInductive and the
// repeated pruning loop in Pdr_ManDeriveInfinityClauses.

import "errors"

type PdrPruneOptions struct {
	MaxVariables, MaxClauses int
	MaxQueries               uint64
	SAT                      SatLimits
}
type PdrPruneResult struct {
	// Complete means transition closure only; initiation and safety must still
	// be checked with Verify. Interrupted runs return no candidate invariant.
	Complete        bool
	Invariant       *PdrInvariant
	Removed, Passes int
	Queries         uint64
	QueryLimit      bool
	Reason          SatStopReason
	Stats           SatStats
}

// PruneNonInductive repeatedly removes clauses violated by a successor of a
// state satisfying all clauses retained at the start of that pass. Repeating
// is essential: removing a clause can invalidate another clause's induction.
// Positive environment constraints guard transitions. Source is immutable.
func (invariant *PdrInvariant) PruneNonInductive(model *AigerModel, options PdrPruneOptions) (result PdrPruneResult, err error) {
	if invariant == nil || options.MaxVariables < 0 || options.MaxClauses < 0 {
		return result, ErrDimension
	}
	if err = model.Validate(); err != nil {
		return result, err
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 1 << 20
	}
	if options.MaxQueries == 0 {
		options.MaxQueries = 100000
	}
	if len(invariant.clauses) > options.MaxClauses {
		return result, ErrLimit
	}
	for _, cube := range invariant.clauses {
		if err = cube.Check(); err != nil {
			return result, err
		}
		for _, l := range cube.state {
			if uint64(l.Index()) >= uint64(model.graph.RegisterCount()) {
				return result, ErrDimension
			}
		}
	}
	stopped := func() bool { return options.SAT.Stop != nil && options.SAT.Stop() }
	if stopped() {
		result.Reason = SatCancelled
		return result, nil
	}
	settings, err := (PdrOptions{MaxVariables: options.MaxVariables, MaxFrameVariables: options.MaxVariables, SAT: options.SAT}).normalized()
	if err != nil {
		return result, err
	}
	p, err := newPdrEngine(model, settings)
	if err != nil {
		return result, err
	}
	remaining := append([]*PdrCube(nil), invariant.clauses...)
	for {
		if stopped() {
			result.Reason = SatCancelled
			return result, nil
		}
		p.frames = nil
		if err = p.newFrame(false); err != nil {
			if errors.Is(err, ErrCancelled) || errors.Is(err, errPdrStopped) {
				result.Reason = SatCancelled
				return result, nil
			}
			return result, err
		}
		solver := p.frames[0].solver
		// Each pass retains its original premise even after a clause is marked.
		pass := func() (kept []*PdrCube, complete bool, cause error) {
			defer func() { result.Stats = pdrAddSatStats(result.Stats, solver.Statistics()) }()
			for _, cube := range remaining {
				if stopped() {
					result.Reason = SatCancelled
					return nil, false, nil
				}
				clause, err := p.cubeLiterals(cube, false, true)
				if err != nil {
					return nil, false, err
				}
				if _, err = solver.AddClause(clause); err != nil {
					return nil, false, err
				}
			}
			result.Passes++
			kept = make([]*PdrCube, 0, len(remaining))
			for _, cube := range remaining {
				if stopped() {
					result.Reason = SatCancelled
					return nil, false, nil
				}
				if result.Queries >= options.MaxQueries {
					result.QueryLimit = true
					return nil, false, nil
				}
				literals, err := p.cubeLiterals(cube, true, false)
				if err != nil {
					return nil, false, err
				}
				result.Queries++
				answer, err := solver.Solve(literals, options.SAT)
				if err != nil {
					return nil, false, err
				}
				if answer.Status == SatUnknown {
					result.Reason = answer.Reason
					return nil, false, nil
				}
				if answer.Status == SatUnsatisfiable {
					kept = append(kept, cube)
				}
			}
			return kept, true, nil
		}
		kept, complete, cause := pass()
		if cause != nil {
			return result, cause
		}
		if !complete {
			return result, nil
		}
		if len(kept) == len(remaining) {
			result.Invariant, err = NewPdrInvariant(kept)
			if err != nil {
				return result, err
			}
			result.Complete = true
			return result, nil
		}
		result.Removed += len(remaining) - len(kept)
		remaining = kept
	}
}
