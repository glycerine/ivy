package abc

// Edge-delay feedback from aig/gia/giaSatLut.c:Sbl_ManEvaluateMappingEdge.

type GiaEdgeWindowEvaluation struct {
	Accepted bool
	Edges    *GiaEdgeAssignment
	Timing   GiaEdgeTiming
	Clause   []SatLit // negative selection variables of all nonempty proposal entries
}

// EvaluateEdgeWindow recomputes native greedy two-edge timing for the complete
// proposed network. Late proposals exclude every mapped window node together,
// in window order, as C does. This heuristic exclusion is not an optimality
// proof. The signed maximum is explicit; zero is not an automatic baseline.
// Boolean/closure validation is separate. Source state stays unchanged.
func (n *GiaLUTNetwork) EvaluateEdgeWindow(m *GiaManager, nodes []GiaID, window [][]GiaID, maximum int, o GiaLUTAnalysisOptions) (GiaEdgeWindowEvaluation, error) {
	r, e := n.evaluateEdgeWindow(m, nodes, window, maximum, o, nil)
	return r.Evaluation, e
}

func (n *GiaLUTNetwork) evaluateEdgeWindow(m *GiaManager, nodes []GiaID, window [][]GiaID, maximum int, o GiaLUTAnalysisOptions, boxes *TimManager) (GiaBoxEdgeWindowEvaluation, error) {
	fail := func(e error) (GiaBoxEdgeWindowEvaluation, error) { return GiaBoxEdgeWindowEvaluation{}, e }
	if n == nil || m == nil || len(nodes) != len(window) || int64(maximum) < -2147483648 || int64(maximum) > 2147483647 || o.MaxObjects < 0 || o.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	capObjects := o.MaxObjects
	if capObjects == 0 {
		capObjects = 1 << 20
	}
	if len(nodes) > capObjects || len(m.objects) > capObjects {
		return fail(ErrLimit)
	}
	for _, id := range nodes {
		if o.Stop != nil && o.Stop() {
			return fail(ErrCancelled)
		}
		if int(id) >= len(m.objects) || !m.objects[id].IsAnd() {
			return fail(ErrDimension)
		}
	}
	var edges *GiaEdgeAssignment
	var timing GiaEdgeTiming
	var manager *TimManager
	var order []GiaID
	if boxes == nil {
		var e error
		edges, timing, e = n.EvaluateWindow(m, nodes, window, true, o)
		if e != nil {
			return fail(e)
		}
	} else {
		proposal, e := n.WithWindow(nodes, window, o)
		if e != nil {
			return fail(e)
		}
		selected, e := proposal.SelectEdgesWithBoxes(m, boxes, true, o)
		if e != nil {
			return fail(e)
		}
		edges, timing, manager, order = selected.Edges, selected.Timing, selected.Manager, selected.Order
	}
	r := GiaEdgeWindowEvaluation{Accepted: timing.Maximum <= maximum, Edges: edges, Timing: timing}
	if !r.Accepted {
		for i, fanins := range window {
			if o.Stop != nil && o.Stop() {
				return fail(ErrCancelled)
			}
			if len(fanins) > 0 {
				l, e := MakeSatLit(SatVar(i), true)
				if e != nil {
					return fail(e)
				}
				r.Clause = append(r.Clause, l)
			}
		}
	}
	if o.Stop != nil && o.Stop() {
		return fail(ErrCancelled)
	}
	return GiaBoxEdgeWindowEvaluation{Evaluation: r, Manager: manager, Order: order}, nil
}

type GiaEdgeWindowSearchResult GiaUnitWindowSearchResult

// SearchEdgeWindow runs the native structural SAT search with explicit global
// edge-delay feedback. Timing boxes and automatic baseline selection are
// separate entry points; the maximum is used verbatim, including zero.
func (n *GiaLUTNetwork) SearchEdgeWindow(m *GiaManager, leaves, nodes, roots []GiaID, maximum int, o GiaUnitWindowSearchOptions) (GiaEdgeWindowSearchResult, error) {
	if int64(maximum) < -2147483648 || int64(maximum) > 2147483647 {
		return GiaEdgeWindowSearchResult{}, ErrDimension
	}
	r, e := n.searchWindow(m, leaves, nodes, roots, 0, o, nil, &giaWindowSearchTiming{maximum: maximum})
	return GiaEdgeWindowSearchResult(r), e
}

type GiaValidatedEdgeWindow struct {
	Network   *GiaLUTNetwork
	Functions []GiaLUTFunction
	Edges     *GiaEdgeAssignment
	Timing    GiaEdgeTiming
}

// WithValidatedEdgeWindow checks structural functions/closure and applies the
// global edge limit. It imposes no additional unit-delay condition.
func (n *GiaLUTNetwork) WithValidatedEdgeWindow(m *GiaManager, nodes []GiaID, window [][]GiaID, maximum int, o GiaLUTAnalysisOptions) (GiaValidatedEdgeWindow, error) {
	return n.withValidatedEdgeWindow(m, nodes, window, maximum, o, true)
}

func (n *GiaLUTNetwork) withValidatedEdgeWindow(m *GiaManager, nodes []GiaID, window [][]GiaID, maximum int, o GiaLUTAnalysisOptions, enforceTiming bool) (GiaValidatedEdgeWindow, error) {
	fail := func(e error) (GiaValidatedEdgeWindow, error) { return GiaValidatedEdgeWindow{}, e }
	if _, e := n.ValidateFunctions(m, o); e != nil {
		return fail(e)
	}
	proposal, e := n.WithWindow(nodes, window, o)
	if e != nil {
		return fail(e)
	}
	functions, e := proposal.ValidateFunctions(m, o)
	if e != nil {
		return fail(e)
	}
	evaluation, e := n.EvaluateEdgeWindow(m, nodes, window, maximum, o)
	if e != nil {
		return fail(e)
	}
	if enforceTiming && !evaluation.Accepted {
		return fail(ErrDimension)
	}
	return GiaValidatedEdgeWindow{Network: proposal, Functions: functions, Edges: evaluation.Edges, Timing: evaluation.Timing}, nil
}

type GiaEdgeWindowRemapResult struct {
	Search      GiaEdgeWindowSearchResult
	Replacement GiaValidatedEdgeWindow
}

// RemapEdgeWindow joins explicit-window edge search to checked replacement
// construction. Network is nil when no improvement was found. Every error
// returns no partial proposal; graph, network and prior edge rows are untouched.
func (n *GiaLUTNetwork) RemapEdgeWindow(m *GiaManager, leaves, nodes, roots []GiaID, maximum int, o GiaUnitWindowSearchOptions) (GiaEdgeWindowRemapResult, error) {
	return n.remapEdgeWindow(m, leaves, nodes, roots, maximum, o, nil)
}

func (n *GiaLUTNetwork) remapEdgeWindow(m *GiaManager, leaves, nodes, roots []GiaID, maximum int, o GiaUnitWindowSearchOptions, session *giaUnitWindowSolverSession) (GiaEdgeWindowRemapResult, error) {
	if int64(maximum) < -2147483648 || int64(maximum) > 2147483647 {
		return GiaEdgeWindowRemapResult{}, ErrDimension
	}
	raw, e := n.searchWindow(m, leaves, nodes, roots, 0, o, session, &giaWindowSearchTiming{maximum: maximum})
	search := GiaEdgeWindowSearchResult(raw)
	if e != nil {
		return GiaEdgeWindowRemapResult{}, e
	}
	r := GiaEdgeWindowRemapResult{Search: search}
	if !search.Improved {
		return r, nil
	}
	analysis := o.Analysis
	prior := analysis.Stop
	analysis.Stop = func() bool { return o.Stop != nil && o.Stop() || prior != nil && prior() }
	r.Replacement, e = n.withValidatedEdgeWindow(m, nodes, search.BestWindow, maximum, analysis, !o.AreaOnly)
	if e != nil {
		return GiaEdgeWindowRemapResult{}, e
	}
	return r, nil
}
