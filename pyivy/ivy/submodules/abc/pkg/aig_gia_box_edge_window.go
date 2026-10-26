package abc

// Box-aware candidate feedback from giaSatLut.c:Sbl_ManEvaluateMappingEdge.

type giaWindowSearchTiming struct {
	unit    bool
	maximum int
	boxes   *TimManager
}
type GiaBoxEdgeWindowEvaluation struct {
	Evaluation GiaEdgeWindowEvaluation
	Manager    *TimManager
	Order      []GiaID
}

// EvaluateEdgeWindowWithBoxes evaluates staged global greedy two-edge timing
// on an owned timing clone and emits the native all-nonempty-node exclusion.
// The maximum is explicit, including zero. No-box behavior ignores terminal
// overrides and preserves epochs. This low-level operation does not prove
// Boolean/closure validity. Each storage/selection phase obeys the limits.
func (n *GiaLUTNetwork) EvaluateEdgeWindowWithBoxes(m *GiaManager, boxes *TimManager, nodes []GiaID, window [][]GiaID, maximum int, o GiaLUTAnalysisOptions) (GiaBoxEdgeWindowEvaluation, error) {
	if boxes == nil {
		return GiaBoxEdgeWindowEvaluation{}, ErrDimension
	}
	return n.evaluateEdgeWindow(m, nodes, window, maximum, o, boxes)
}

func (n *GiaLUTNetwork) SearchEdgeWindowWithBoxes(m *GiaManager, boxes *TimManager, leaves, nodes, roots []GiaID, maximum int, o GiaUnitWindowSearchOptions) (GiaEdgeWindowSearchResult, error) {
	if boxes == nil || m == nil || len(boxes.cis) != len(m.cis) || len(boxes.cos) != len(m.cos) || int64(maximum) < -2147483648 || int64(maximum) > 2147483647 {
		return GiaEdgeWindowSearchResult{}, ErrDimension
	}
	r, e := n.searchWindow(m, leaves, nodes, roots, 0, o, nil, &giaWindowSearchTiming{maximum: maximum, boxes: boxes})
	return GiaEdgeWindowSearchResult(r), e
}

type GiaValidatedBoxEdgeWindow struct {
	Network   *GiaLUTNetwork
	Functions []GiaLUTFunction
	Edges     *GiaEdgeAssignment
	Timing    GiaEdgeTiming
	Manager   *TimManager
	Order     []GiaID
}

// WithValidatedEdgeWindowWithBoxes checks complete structural functions and
// dependency closure, then the staged global edge limit, before exposing owned
// mapping/edge/timing updates. No source state changes on any path.
func (n *GiaLUTNetwork) WithValidatedEdgeWindowWithBoxes(m *GiaManager, boxes *TimManager, nodes []GiaID, window [][]GiaID, maximum int, o GiaLUTAnalysisOptions) (GiaValidatedBoxEdgeWindow, error) {
	return n.withValidatedEdgeWindowWithBoxes(m, boxes, nodes, window, maximum, o, true)
}

func (n *GiaLUTNetwork) withValidatedEdgeWindowWithBoxes(m *GiaManager, boxes *TimManager, nodes []GiaID, window [][]GiaID, maximum int, o GiaLUTAnalysisOptions, enforceTiming bool) (GiaValidatedBoxEdgeWindow, error) {
	fail := func(e error) (GiaValidatedBoxEdgeWindow, error) { return GiaValidatedBoxEdgeWindow{}, e }
	if boxes == nil {
		return fail(ErrDimension)
	}
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
	evaluation, e := n.EvaluateEdgeWindowWithBoxes(m, boxes, nodes, window, maximum, o)
	if e != nil {
		return fail(e)
	}
	if enforceTiming && !evaluation.Evaluation.Accepted {
		return fail(ErrDimension)
	}
	return GiaValidatedBoxEdgeWindow{Network: proposal, Functions: functions, Edges: evaluation.Evaluation.Edges, Timing: evaluation.Evaluation.Timing, Manager: evaluation.Manager, Order: evaluation.Order}, nil
}

type GiaBoxEdgeWindowRemapResult struct {
	Search      GiaEdgeWindowSearchResult
	Replacement GiaValidatedBoxEdgeWindow
}

func (n *GiaLUTNetwork) RemapEdgeWindowWithBoxes(m *GiaManager, boxes *TimManager, leaves, nodes, roots []GiaID, maximum int, o GiaUnitWindowSearchOptions) (GiaBoxEdgeWindowRemapResult, error) {
	return n.remapEdgeWindowWithBoxes(m, boxes, leaves, nodes, roots, maximum, o, nil)
}

func (n *GiaLUTNetwork) remapEdgeWindowWithBoxes(m *GiaManager, boxes *TimManager, leaves, nodes, roots []GiaID, maximum int, o GiaUnitWindowSearchOptions, session *giaUnitWindowSolverSession) (GiaBoxEdgeWindowRemapResult, error) {
	if boxes == nil || m == nil || len(boxes.cis) != len(m.cis) || len(boxes.cos) != len(m.cos) || int64(maximum) < -2147483648 || int64(maximum) > 2147483647 {
		return GiaBoxEdgeWindowRemapResult{}, ErrDimension
	}
	raw, e := n.searchWindow(m, leaves, nodes, roots, 0, o, session, &giaWindowSearchTiming{maximum: maximum, boxes: boxes})
	search := GiaEdgeWindowSearchResult(raw)
	if e != nil {
		return GiaBoxEdgeWindowRemapResult{}, e
	}
	r := GiaBoxEdgeWindowRemapResult{Search: search}
	if !search.Improved {
		return r, nil
	}
	analysis := o.Analysis
	prior := analysis.Stop
	analysis.Stop = func() bool { return o.Stop != nil && o.Stop() || prior != nil && prior() }
	r.Replacement, e = n.withValidatedEdgeWindowWithBoxes(m, boxes, nodes, search.BestWindow, maximum, analysis, !o.AreaOnly)
	if e != nil {
		return GiaBoxEdgeWindowRemapResult{}, e
	}
	return r, nil
}
