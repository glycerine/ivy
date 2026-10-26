package abc

// Boxed candidate unit feedback from giaSatLut.c:Sbl_ManEvaluateMapping.

type GiaBoxUnitWindowEvaluation struct {
	Evaluation GiaUnitWindowEvaluation
	Manager    *TimManager // owned ORIGINAL arrival/required propagation, not applied proposal state
	Order      []GiaID
}

// EvaluateUnitWindowWithBoxes uses corrected boxed original timing, then
// native local candidate arrivals and first-critical-path feedback. Candidate
// recomputation keeps original box-output CI arrivals, as C does; checked
// application separately recomputes global staged timing. Limits apply to the
// original boxed timing and local evaluation separately. Source stays unchanged.
func (n *GiaLUTNetwork) EvaluateUnitWindowWithBoxes(m *GiaManager, boxes *TimManager, nodes []GiaID, window [][]GiaID, roots []GiaID, delayStart int, o GiaLUTAnalysisOptions) (GiaBoxUnitWindowEvaluation, error) {
	original, e := n.UnitTimingWithBoxes(m, boxes, delayStart, o)
	if e != nil {
		return GiaBoxUnitWindowEvaluation{}, e
	}
	evaluation, e := n.evaluateUnitWindow(m, nodes, window, roots, delayStart, o, &original.Timing)
	if e != nil {
		return GiaBoxUnitWindowEvaluation{}, e
	}
	return GiaBoxUnitWindowEvaluation{Evaluation: evaluation, Manager: original.Manager, Order: original.Order}, nil
}
func (n *GiaLUTNetwork) SearchUnitWindowWithBoxes(m *GiaManager, boxes *TimManager, leaves, nodes, roots []GiaID, delayStart int, o GiaUnitWindowSearchOptions) (GiaUnitWindowSearchResult, error) {
	if boxes == nil || m == nil || len(boxes.cis) != len(m.cis) || len(boxes.cos) != len(m.cos) || int64(delayStart) < -2147483648 || int64(delayStart) > 2147483647 {
		return GiaUnitWindowSearchResult{}, ErrDimension
	}
	return n.searchWindow(m, leaves, nodes, roots, delayStart, o, nil, &giaWindowSearchTiming{unit: true, boxes: boxes})
}

type GiaValidatedBoxUnitWindow struct {
	Network   *GiaLUTNetwork
	Functions []GiaLUTFunction
	Timing    GiaBoxLUTUnitTiming
}

// WithValidatedUnitWindowWithBoxes checks complete functions/closure and full
// staged original/proposal maxima. The global check catches changes propagated
// through boxes even when local candidate CI arrivals or root lists omit them.
func (n *GiaLUTNetwork) WithValidatedUnitWindowWithBoxes(m *GiaManager, boxes *TimManager, nodes []GiaID, window [][]GiaID, delayStart int, o GiaLUTAnalysisOptions) (GiaValidatedBoxUnitWindow, error) {
	return n.withValidatedUnitWindowWithBoxes(m, boxes, nodes, window, delayStart, o, true)
}

func (n *GiaLUTNetwork) withValidatedUnitWindowWithBoxes(m *GiaManager, boxes *TimManager, nodes []GiaID, window [][]GiaID, delayStart int, o GiaLUTAnalysisOptions, enforceTiming bool) (GiaValidatedBoxUnitWindow, error) {
	fail := func(e error) (GiaValidatedBoxUnitWindow, error) { return GiaValidatedBoxUnitWindow{}, e }
	if _, e := n.ValidateFunctions(m, o); e != nil {
		return fail(e)
	}
	original, e := n.UnitTimingWithBoxes(m, boxes, delayStart, o)
	if e != nil {
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
	timing, e := proposal.UnitTimingWithBoxes(m, boxes, delayStart, o)
	if e != nil {
		return fail(e)
	}
	if enforceTiming && timing.Timing.Maximum > original.Timing.Maximum {
		return fail(ErrDimension)
	}
	return GiaValidatedBoxUnitWindow{Network: proposal, Functions: functions, Timing: timing}, nil
}

type GiaBoxUnitWindowRemapResult struct {
	Search      GiaUnitWindowSearchResult
	Replacement GiaValidatedBoxUnitWindow
}

func (n *GiaLUTNetwork) RemapUnitWindowWithBoxes(m *GiaManager, boxes *TimManager, leaves, nodes, roots []GiaID, delayStart int, o GiaUnitWindowSearchOptions) (GiaBoxUnitWindowRemapResult, error) {
	return n.remapUnitWindowWithBoxes(m, boxes, leaves, nodes, roots, delayStart, o, nil)
}

func (n *GiaLUTNetwork) remapUnitWindowWithBoxes(m *GiaManager, boxes *TimManager, leaves, nodes, roots []GiaID, delayStart int, o GiaUnitWindowSearchOptions, session *giaUnitWindowSolverSession) (GiaBoxUnitWindowRemapResult, error) {
	if boxes == nil || m == nil || len(boxes.cis) != len(m.cis) || len(boxes.cos) != len(m.cos) || int64(delayStart) < -2147483648 || int64(delayStart) > 2147483647 {
		return GiaBoxUnitWindowRemapResult{}, ErrDimension
	}
	raw, e := n.searchWindow(m, leaves, nodes, roots, delayStart, o, session, &giaWindowSearchTiming{unit: true, boxes: boxes})
	search := raw
	if e != nil {
		return GiaBoxUnitWindowRemapResult{}, e
	}
	r := GiaBoxUnitWindowRemapResult{Search: search}
	if !search.Improved {
		return r, nil
	}
	analysis := o.Analysis
	prior := analysis.Stop
	analysis.Stop = func() bool { return o.Stop != nil && o.Stop() || prior != nil && prior() }
	r.Replacement, e = n.withValidatedUnitWindowWithBoxes(m, boxes, nodes, search.BestWindow, delayStart, analysis, !o.AreaOnly)
	if e != nil {
		return GiaBoxUnitWindowRemapResult{}, e
	}
	return r, nil
}
