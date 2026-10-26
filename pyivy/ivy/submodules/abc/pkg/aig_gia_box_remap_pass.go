package abc

// Box-aware whole-network scheduling from giaSatLut.c:Gia_ManLutSat.

type GiaBoxUnitRemapPassResult struct {
	Pass    GiaUnitRemapPassResult
	Manager *TimManager
	Order   []GiaID
}

// RemapUnitPassWithBoxes reuses native pivot/SAT/discovery scheduling and
// carries only accepted owned timing updates between windows. Corrected boxed
// unit propagation and a final global comparison protect the original delay
// unless AreaOnly explicitly permits a timing tradeoff.
func (n *GiaLUTNetwork) RemapUnitPassWithBoxes(m *GiaManager, boxes *TimManager, delayStart int, o GiaUnitRemapPassOptions) (GiaBoxUnitRemapPassResult, error) {
	fail := func(e error) (GiaBoxUnitRemapPassResult, error) { return GiaBoxUnitRemapPassResult{}, e }
	analysis := o.Search.Analysis
	old := analysis.Stop
	analysis.Stop = func() bool { return o.Search.Stop != nil && o.Search.Stop() || old != nil && old() }
	original, e := n.UnitTimingWithBoxes(m, boxes, delayStart, analysis)
	if e != nil {
		return fail(e)
	}
	r, e := n.remapPass(m, delayStart, o, nil, boxes)
	if e != nil {
		return fail(e)
	}
	if !o.Search.AreaOnly && r.Timing.Maximum > original.Timing.Maximum {
		return fail(ErrInvariant)
	}
	return GiaBoxUnitRemapPassResult{Pass: r.GiaUnitRemapPassResult, Manager: r.manager, Order: r.order}, nil
}

type GiaBoxEdgeRemapPassResult struct {
	Pass    GiaEdgeRemapPassResult
	Manager *TimManager
	Order   []GiaID
}

// RemapEdgePassWithBoxes derives an automatic zero limit from the supplied
// prior assignment with staged boxes when timing is constrained. AreaOnly
// retains prior rows across mapping changes and does not enforce a delay limit.
// Trials stay private; accepted mapping,
// rows and timing move together. An unchanged result retains owned prior rows.
func (n *GiaLUTNetwork) RemapEdgePassWithBoxes(m *GiaManager, boxes *TimManager, prior *GiaEdgeAssignment, maximum int, o GiaUnitRemapPassOptions) (GiaBoxEdgeRemapPassResult, error) {
	fail := func(e error) (GiaBoxEdgeRemapPassResult, error) { return GiaBoxEdgeRemapPassResult{}, e }
	if boxes == nil || prior == nil || int64(maximum) < -2147483648 || int64(maximum) > 2147483647 {
		return fail(ErrDimension)
	}
	analysis := o.Search.Analysis
	old := analysis.Stop
	analysis.Stop = func() bool { return o.Search.Stop != nil && o.Search.Stop() || old != nil && old() }
	if _, e := n.ValidateFunctions(m, analysis); e != nil {
		return fail(e)
	}
	baseline, e := n.EvaluateDelayWithBoxes(m, prior, boxes, analysis)
	if e != nil {
		return fail(e)
	}
	if maximum == 0 && !o.Search.AreaOnly {
		maximum = baseline.Timing.Maximum
	}
	r, e := n.remapPass(m, 0, o, &giaEdgeRemapContext{maximum: maximum, prior: prior, baseline: baseline.Timing}, baseline.Manager)
	if e != nil {
		return fail(e)
	}
	return GiaBoxEdgeRemapPassResult{Pass: r.edgeResult(maximum, !o.Search.AreaOnly), Manager: r.manager, Order: r.order}, nil
}

type GiaPackedBoxUnitRemapPassResult struct {
	Mapping *GiaLUTMapping
	Pass    GiaBoxUnitRemapPassResult
}

func (mapping *GiaLUTMapping) RemapUnitPassWithBoxes(m *GiaManager, boxes *TimManager, delayStart int, o GiaUnitRemapPassOptions) (GiaPackedBoxUnitRemapPassResult, error) {
	fail := func(e error) (GiaPackedBoxUnitRemapPassResult, error) { return GiaPackedBoxUnitRemapPassResult{}, e }
	if o.MaxPackedEntries < 0 {
		return fail(ErrDimension)
	}
	analysis := o.Search.Analysis
	old := analysis.Stop
	analysis.Stop = func() bool { return o.Search.Stop != nil && o.Search.Stop() || old != nil && old() }
	n, e := mapping.AlternateNetwork(analysis)
	if e != nil {
		return fail(e)
	}
	pass, e := n.RemapUnitPassWithBoxes(m, boxes, delayStart, o)
	if e != nil {
		return fail(e)
	}
	packed, e := pass.Pass.Network.PackedMapping(o.MaxPackedEntries, analysis)
	if e != nil {
		return fail(e)
	}
	return GiaPackedBoxUnitRemapPassResult{Mapping: packed, Pass: pass}, nil
}

type GiaPackedBoxEdgeRemapPassResult struct {
	Mapping *GiaLUTMapping
	Pass    GiaBoxEdgeRemapPassResult
}

func (mapping *GiaLUTMapping) RemapEdgePassWithBoxes(m *GiaManager, boxes *TimManager, prior *GiaEdgeAssignment, maximum int, o GiaUnitRemapPassOptions) (GiaPackedBoxEdgeRemapPassResult, error) {
	fail := func(e error) (GiaPackedBoxEdgeRemapPassResult, error) { return GiaPackedBoxEdgeRemapPassResult{}, e }
	if o.MaxPackedEntries < 0 {
		return fail(ErrDimension)
	}
	analysis := o.Search.Analysis
	old := analysis.Stop
	analysis.Stop = func() bool { return o.Search.Stop != nil && o.Search.Stop() || old != nil && old() }
	n, e := mapping.Network(analysis)
	if e != nil {
		return fail(e)
	}
	pass, e := n.RemapEdgePassWithBoxes(m, boxes, prior, maximum, o)
	if e != nil {
		return fail(e)
	}
	packed, e := pass.Pass.Network.PackedMapping(o.MaxPackedEntries, analysis)
	if e != nil {
		return fail(e)
	}
	return GiaPackedBoxEdgeRemapPassResult{Mapping: packed, Pass: pass}, nil
}
