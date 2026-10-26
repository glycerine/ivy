package abc

// Whole-network edge scheduling from aig/gia/giaSatLut.c:Gia_ManLutSat.

type GiaEdgeRemapPassResult struct {
	TimingConstrained                                     bool
	Network                                               *GiaLUTNetwork
	Functions                                             []GiaLUTFunction
	Edges                                                 *GiaEdgeAssignment
	Timing                                                GiaEdgeTiming
	DelayLimit                                            int
	Attempts                                              []GiaUnitRemapAttempt
	Improvements                                          int
	SolverBuilds, WindowResets                            int
	DiscoveryBuilds, ReferenceRefreshes, DiscoveryQueries int
}

// RemapEdgePass uses the native pivot schedule and reusable sessions. Maximum
// zero derives its limit from the supplied prior assignment when timing is
// constrained. AreaOnly leaves the maximum unenforced and preserves prior rows
// across mapping changes. Baselines never substitute fresh greedy
// selection. An unchanged result preserves owned prior rows and may exceed an
// explicitly tighter limit. Timing-constrained improvements
// return freshly selected rows for the accepted mapping; trial rows never leak.
// The graph, source network and prior assignment stay unchanged on every path.
func (n *GiaLUTNetwork) RemapEdgePass(m *GiaManager, prior *GiaEdgeAssignment, maximum int, o GiaUnitRemapPassOptions) (GiaEdgeRemapPassResult, error) {
	fail := func(e error) (GiaEdgeRemapPassResult, error) { return GiaEdgeRemapPassResult{}, e }
	if prior == nil || int64(maximum) < -2147483648 || int64(maximum) > 2147483647 {
		return fail(ErrDimension)
	}
	analysis := o.Search.Analysis
	old := analysis.Stop
	analysis.Stop = func() bool { return o.Search.Stop != nil && o.Search.Stop() || old != nil && old() }
	if _, e := n.ValidateFunctions(m, analysis); e != nil {
		return fail(e)
	}
	mapping, e := n.selectionMapping(analysis)
	if e != nil {
		return fail(e)
	}
	baseline, e := prior.EvaluateDelay(m, mapping, analysis)
	if e != nil {
		return fail(e)
	}
	if maximum == 0 && !o.Search.AreaOnly {
		maximum = baseline.Maximum
	}
	r, e := n.remapPass(m, 0, o, &giaEdgeRemapContext{maximum: maximum, prior: prior, baseline: baseline}, nil)
	if e != nil {
		return fail(e)
	}
	return r.edgeResult(maximum, !o.Search.AreaOnly), nil
}

func (r giaScheduledRemapResult) edgeResult(maximum int, constrained bool) GiaEdgeRemapPassResult {
	return GiaEdgeRemapPassResult{TimingConstrained: constrained, Network: r.Network, Functions: r.Functions, Edges: r.edges, Timing: r.edgeTiming, DelayLimit: maximum, Attempts: r.Attempts, Improvements: r.Improvements, SolverBuilds: r.SolverBuilds, WindowResets: r.WindowResets, DiscoveryBuilds: r.DiscoveryBuilds, ReferenceRefreshes: r.ReferenceRefreshes, DiscoveryQueries: r.DiscoveryQueries}
}

type GiaPackedEdgeRemapPassResult struct {
	Mapping *GiaLUTMapping
	Pass    GiaEdgeRemapPassResult
}

// RemapEdgePass converts packed storage, runs checked edge remapping and
// returns canonical packed output plus the owned accepted edge assignment.
func (mapping *GiaLUTMapping) RemapEdgePass(m *GiaManager, prior *GiaEdgeAssignment, maximum int, o GiaUnitRemapPassOptions) (GiaPackedEdgeRemapPassResult, error) {
	fail := func(e error) (GiaPackedEdgeRemapPassResult, error) { return GiaPackedEdgeRemapPassResult{}, e }
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
	pass, e := n.RemapEdgePass(m, prior, maximum, o)
	if e != nil {
		return fail(e)
	}
	packed, e := pass.Network.PackedMapping(o.MaxPackedEntries, analysis)
	if e != nil {
		return fail(e)
	}
	return GiaPackedEdgeRemapPassResult{Mapping: packed, Pass: pass}, nil
}
