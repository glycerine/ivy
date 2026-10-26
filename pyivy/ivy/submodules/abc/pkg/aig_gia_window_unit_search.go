package abc

// Unit-delay SAT-window search from aig/gia/giaSatLut.c:Sbl_ManTestSat.

import "slices"

type GiaUnitWindowSearchOptions struct {
	AreaOnly      bool // native fDelay=false: optimize LUT count without timing exclusions
	LutSize       int  // zero: four; native sizes 1..4 use four, 5..6 use six
	Cuts          GiaWindowCutOptions
	CNF           GiaWindowCNFOptions
	Analysis      GiaLUTAnalysisOptions // per timing/decoding operation
	SAT           SatLimits             // per solve, matching native per-call conflict limits
	MaxIterations int                   // zero: native ten-iteration limit
	Stop          func() bool
}
type GiaUnitWindowSearchStep struct {
	Bound    int
	Status   SatStatus
	Accepted bool
	Cuts     []int
	Conflict []SatLit
}
type GiaUnitWindowSearchResult struct {
	Steps                    []GiaUnitWindowSearchStep
	BestCuts                 []int
	BestWindow               [][]GiaID
	InitialNodes, BestNodes  int
	Improved, IterationLimit bool
}

// SearchUnitWindow joins candidate generation, SAT, decoding and native unit
// timing feedback on an owned solver. It returns a proposal without changing
// the graph or mapping. The native node-path exclusion clauses are heuristic:
// exhaustion after timing rejections does not prove timing-feasible optimality.
// Current mapping closure/Boolean correctness remain caller preconditions.
// Unit timing supports native four/six-input feasibility; edge-delay search
// retains its four-input limit. All errors expose no partial proposal.
func (n *GiaLUTNetwork) SearchUnitWindow(m *GiaManager, leaves, nodes, roots []GiaID, delayStart int, o GiaUnitWindowSearchOptions) (GiaUnitWindowSearchResult, error) {
	return n.searchUnitWindow(m, leaves, nodes, roots, delayStart, o, nil)
}

func (n *GiaLUTNetwork) searchUnitWindow(m *GiaManager, leaves, nodes, roots []GiaID, delayStart int, o GiaUnitWindowSearchOptions, session *giaUnitWindowSolverSession) (GiaUnitWindowSearchResult, error) {
	return n.searchWindow(m, leaves, nodes, roots, delayStart, o, session, nil)
}

func (n *GiaLUTNetwork) searchWindow(m *GiaManager, leaves, nodes, roots []GiaID, delayStart int, o GiaUnitWindowSearchOptions, session *giaUnitWindowSolverSession, edgeTiming *giaWindowSearchTiming) (GiaUnitWindowSearchResult, error) {
	fail := func(e error) (GiaUnitWindowSearchResult, error) { return GiaUnitWindowSearchResult{}, e }
	if o.LutSize < 0 || o.LutSize > 6 {
		return fail(ErrDimension)
	}
	if o.LutSize == 0 {
		o.LutSize = 4
	}
	if edgeTiming != nil && !edgeTiming.unit && o.LutSize > 4 {
		return fail(ErrUnsupported)
	}
	if o.MaxIterations < 0 {
		return fail(ErrDimension)
	}
	if o.MaxIterations == 0 {
		o.MaxIterations = 10
	}
	if o.MaxIterations > 10000 {
		return fail(ErrLimit)
	}
	stopped := func() bool { return o.Stop != nil && o.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	combine := func(old func() bool) func() bool { return func() bool { return stopped() || old != nil && old() } }
	o.Cuts.Stop = combine(o.Cuts.Stop)
	o.CNF.Stop = combine(o.CNF.Stop)
	o.Analysis.Stop = combine(o.Analysis.Stop)
	o.SAT.Stop = combine(o.SAT.Stop)
	cuts, e := m.GenerateWindowCuts(leaves, nodes, o.LutSize, o.Cuts)
	if e != nil {
		return fail(e)
	}
	state, e := n.WindowMappingState(m, leaves, nodes, roots, cuts, o.Analysis)
	if e != nil {
		return fail(e)
	}
	r := GiaUnitWindowSearchResult{InitialNodes: len(state.InitialNodes), BestNodes: len(state.InitialNodes)}
	if len(state.RootVariables) == 0 || r.InitialNodes == 0 {
		return r, nil
	}
	log := 0
	for 1<<uint(log) < len(nodes) {
		log++
	}
	var solver *GiaWindowSolver
	if session == nil {
		solver, e = NewGiaWindowSolver(log, len(nodes), cuts.Cuts, cuts.Ranges, state.RootVariables, o.CNF)
	} else {
		solver, e = session.prepare(len(nodes), cuts.Cuts, cuts.Ranges, state.RootVariables, o.CNF)
	}
	if e != nil {
		return fail(e)
	}
	if e = solver.SetMappingPreferences(state); e != nil {
		return fail(e)
	}
	timingRoots := make([]GiaID, len(state.RootVariables))
	for i, v := range state.RootVariables {
		timingRoots[i] = nodes[v]
	}
	return runGiaWindowSearch(n, m, leaves, nodes, timingRoots, cuts, solver, r.InitialNodes, delayStart, o, edgeTiming)
}

func runGiaUnitWindowSearch(n *GiaLUTNetwork, m *GiaManager, leaves, nodes, timingRoots []GiaID, cuts GiaWindowCuts, solver *GiaWindowSolver, initialNodes, delayStart int, o GiaUnitWindowSearchOptions) (GiaUnitWindowSearchResult, error) {
	return runGiaWindowSearch(n, m, leaves, nodes, timingRoots, cuts, solver, initialNodes, delayStart, o, nil)
}

func runGiaWindowSearch(n *GiaLUTNetwork, m *GiaManager, leaves, nodes, timingRoots []GiaID, cuts GiaWindowCuts, solver *GiaWindowSolver, initialNodes, delayStart int, o GiaUnitWindowSearchOptions, edgeTiming *giaWindowSearchTiming) (GiaUnitWindowSearchResult, error) {
	fail := func(e error) (GiaUnitWindowSearchResult, error) { return GiaUnitWindowSearchResult{}, e }
	stopped := func() bool { return o.Stop != nil && o.Stop() }
	r := GiaUnitWindowSearchResult{InitialNodes: initialNodes, BestNodes: initialNodes}
	bound := r.InitialNodes
	for bound > 0 && len(r.Steps) < o.MaxIterations {
		if stopped() {
			return fail(ErrCancelled)
		}
		answer, e := solver.Solve(bound, o.SAT)
		if e != nil {
			return fail(e)
		}
		step := GiaUnitWindowSearchStep{Bound: bound, Status: answer.SAT.Status}
		if answer.SAT.Status != SatSatisfiable {
			r.Steps = append(r.Steps, step)
			break
		}
		step.Cuts = slices.Clone(answer.SelectedCuts)
		window, e := DecodeGiaWindowCuts(leaves, nodes, cuts.Cuts, answer.SelectedCuts, o.Analysis)
		if e != nil {
			return fail(e)
		}
		var accepted bool
		var clause []SatLit
		if o.AreaOnly {
			accepted = true
		} else if edgeTiming == nil {
			timing, e := n.EvaluateUnitWindow(m, nodes, window, timingRoots, delayStart, o.Analysis)
			if e != nil {
				return fail(e)
			}
			accepted = timing.Accepted
			clause = timing.Clause
		} else if edgeTiming.unit {
			timing, e := n.EvaluateUnitWindowWithBoxes(m, edgeTiming.boxes, nodes, window, timingRoots, delayStart, o.Analysis)
			if e != nil {
				return fail(e)
			}
			accepted = timing.Evaluation.Accepted
			clause = timing.Evaluation.Clause
		} else {
			timing, e := n.evaluateEdgeWindow(m, nodes, window, edgeTiming.maximum, o.Analysis, edgeTiming.boxes)
			if e != nil {
				return fail(e)
			}
			accepted = timing.Evaluation.Accepted
			clause = timing.Evaluation.Clause
		}
		step.Accepted = accepted
		if accepted {
			r.BestCuts = slices.Clone(answer.SelectedCuts)
			r.BestWindow = window
			r.BestNodes = len(answer.SelectedNodes)
			r.Improved = len(r.BestCuts) < r.InitialNodes // native acceptance counts selected cuts
			bound--
		} else {
			if len(clause) == 0 {
				return fail(ErrInvariant)
			}
			for _, l := range clause {
				if int(l.Var()) >= len(nodes) || !l.Negative() || !answer.SAT.Model[l.Var()] {
					return fail(ErrInvariant)
				}
			}
			step.Conflict = slices.Clone(clause)
			if _, e = solver.solver.AddClause(clause); e != nil {
				return fail(e)
			}
		}
		r.Steps = append(r.Steps, step)
	}
	r.IterationLimit = bound > 0 && len(r.Steps) == o.MaxIterations && r.Steps[len(r.Steps)-1].Status == SatSatisfiable
	if stopped() {
		return fail(ErrCancelled)
	}
	return r, nil
}
