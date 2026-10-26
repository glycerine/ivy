package abc

// Alternate-mapping boxed paths from aig/gia/giaEdge.c:
// Gia_ManEvalEdgeDelay and Gia_ManComputeEdgeDelay mapping2 branches.

// EvaluateDelayWithBoxes evaluates this alternate mapping, preserving its
// empty-list non-LUT semantics. The private packed adapter shares checked
// node kernels and boxed traversal; no caller mapping or timing state changes.
func (n *GiaLUTNetwork) EvaluateDelayWithBoxes(m *GiaManager, edges *GiaEdgeAssignment, timing *TimManager, options GiaLUTAnalysisOptions) (GiaBoxEdgeTiming, error) {
	if n == nil || m == nil || timing == nil || edges == nil || len(n.fanins) != len(m.objects) {
		return GiaBoxEdgeTiming{}, ErrDimension
	}
	mapping, err := n.selectionMapping(options)
	if err != nil {
		return GiaBoxEdgeTiming{}, err
	}
	return edges.EvaluateDelayWithBoxes(m, mapping, timing, options)
}

// SelectEdgesWithBoxes performs native alternate-mapping boxed selection
// through the shared engine. Adapter storage and each evaluation pass obey
// the supplied bounds. Empty fanin lists are omitted from LUT iteration.
func (n *GiaLUTNetwork) SelectEdgesWithBoxes(m *GiaManager, timing *TimManager, useTwo bool, options GiaLUTAnalysisOptions) (GiaBoxEdgeSelection, error) {
	if n == nil || m == nil || timing == nil || len(n.fanins) != len(m.objects) {
		return GiaBoxEdgeSelection{}, ErrDimension
	}
	mapping, err := n.selectionMapping(options)
	if err != nil {
		return GiaBoxEdgeSelection{}, err
	}
	return mapping.SelectEdgesWithBoxes(m, timing, useTwo, options)
}
