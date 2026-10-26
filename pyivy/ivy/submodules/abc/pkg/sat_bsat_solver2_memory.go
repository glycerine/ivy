package abc

// Owned memory accounting for src/sat/bsat/satSolver2.c.

import "unsafe"

type SatSolverMemoryUsage struct {
	SolverStruct, ClauseArena, ClauseLiterals uintptr
	WatchHeaders, WatchEntries                uintptr
	VariableArrays, TrailArrays, ResultArrays uintptr
	Bookmark, Proof                           uintptr
	FacadeStruct, FacadeArena, FacadeLiterals uintptr
	FacadeVariables, FacadeBookmark           uintptr
	Total                                     uintptr
}

func satCapacityBytes[T any](values []T) uintptr {
	var value T
	return uintptr(cap(values)) * unsafe.Sizeof(value)
}

func (usage *SatSolverMemoryUsage) setTotal() {
	usage.Total = usage.SolverStruct + usage.ClauseArena + usage.ClauseLiterals +
		usage.WatchHeaders + usage.WatchEntries + usage.VariableArrays +
		usage.TrailArrays + usage.ResultArrays + usage.Bookmark + usage.Proof +
		usage.FacadeStruct + usage.FacadeArena + usage.FacadeLiterals +
		usage.FacadeVariables + usage.FacadeBookmark
}

func satSolverMemoryUsage(solver *SatSolver, includeWatchEntries, includeBookmark bool) SatSolverMemoryUsage {
	usage := SatSolverMemoryUsage{
		SolverStruct:   unsafe.Sizeof(*solver),
		ClauseArena:    satCapacityBytes(solver.clauses),
		WatchHeaders:   satCapacityBytes(solver.watches),
		VariableArrays: satCapacityBytes(solver.assigns) + satCapacityBytes(solver.levels) + satCapacityBytes(solver.reasons) + satCapacityBytes(solver.polarity) + satCapacityBytes(solver.used) + satCapacityBytes(solver.activity) + satCapacityBytes(solver.activityFactors) + satCapacityBytes(solver.activityFactorVariables) + satCapacityBytes(solver.globalActivityVariables) + satCapacityBytes(solver.order) + satCapacityBytes(solver.orderpos) + satCapacityBytes(solver.seen),
		TrailArrays:    satCapacityBytes(solver.units) + satCapacityBytes(solver.trail) + satCapacityBytes(solver.trailLim),
		ResultArrays:   satCapacityBytes(solver.model) + satCapacityBytes(solver.lastConflict),
	}
	for _, clause := range solver.clauses {
		usage.ClauseLiterals += satCapacityBytes(clause.lits)
	}
	usage.ResultArrays += satCapacityBytes(solver.proofRootConflict)
	if solver.proof != nil && solver.proof.manager != nil {
		usage.Proof = uintptr(solver.proof.manager.Stats().OwnedBytes) + unsafe.Sizeof(*solver.proof) + satCapacityBytes(solver.proof.lastAssumptions)
	}
	if solver.resolutionProof != nil && solver.resolutionProof.dag != nil {
		proof := solver.resolutionProof
		usage.Proof += unsafe.Sizeof(*proof) + unsafe.Sizeof(*proof.dag) + satCapacityBytes(proof.dag.nodes) + satCapacityBytes(proof.variablePart) + satCapacityBytes(proof.rootPartA) + satCapacityBytes(proof.lastAssumptions) + satCapacityBytes(proof.roots)
		for _, root := range proof.roots {
			usage.Proof += satCapacityBytes(root)
		}
		for _, node := range proof.dag.nodes {
			usage.Proof += satCapacityBytes(node.Antecedents)
		}
	}
	if includeWatchEntries {
		for _, watches := range solver.watches {
			usage.WatchEntries += satCapacityBytes(watches)
		}
	}
	if includeBookmark && solver.bookmark != nil {
		bookmark := satSolverMemoryUsage(&solver.bookmark.state, includeWatchEntries, false)
		usage.Bookmark = unsafe.Sizeof(*solver.bookmark) + bookmark.Total - bookmark.SolverStruct
	}
	usage.setTotal()
	return usage
}

// MemoryUsage reports allocated backing capacities. When includeWatchEntries
// is false, watch-list headers remain counted but their individual integer
// arrays are omitted, matching sat_solver2_memory's fAll switch. Runtime and
// allocator metadata outside owned Go objects are intentionally excluded.
func (s *SatSolver) MemoryUsage(includeWatchEntries bool) (SatSolverMemoryUsage, error) {
	if s == nil {
		return SatSolverMemoryUsage{}, ErrDimension
	}
	if len(s.clauses) == 0 {
		return SatSolverMemoryUsage{}, ErrInvariant
	}
	return satSolverMemoryUsage(s, includeWatchEntries, true), nil
}

func satSolver2FacadeBacking(clauses []SatSolver2Clause, variablePart []bool) (arena, literals, variables uintptr) {
	arena = satCapacityBytes(clauses)
	for _, clause := range clauses {
		literals += satCapacityBytes(clause.Literals)
	}
	variables = satCapacityBytes(variablePart)
	return
}

func (s *SatSolver2) MemoryUsage(includeWatchEntries bool) (SatSolverMemoryUsage, error) {
	if !s.usable() {
		return SatSolverMemoryUsage{}, ErrInvariant
	}
	usage, err := s.core.MemoryUsage(includeWatchEntries)
	if err != nil {
		return SatSolverMemoryUsage{}, err
	}
	usage.FacadeStruct = unsafe.Sizeof(*s)
	usage.FacadeArena, usage.FacadeLiterals, usage.FacadeVariables = satSolver2FacadeBacking(s.clauses, s.variablePart)
	if s.bookmark != nil {
		arena, literals, variables := satSolver2FacadeBacking(s.bookmark.clauses, s.bookmark.variablePart)
		usage.FacadeBookmark = unsafe.Sizeof(*s.bookmark) + arena + literals + variables
	}
	usage.setTotal()
	return usage, nil
}
