package abc

// Box-aware plain-mapping greedy selection from aig/gia/giaEdge.c:
// Gia_ManComputeEdgeDelay timing-box branch.

type GiaBoxEdgeSelection struct {
	Edges   *GiaEdgeAssignment
	Timing  GiaEdgeTiming
	Manager *TimManager
	Order   []GiaID
}

// SelectEdgesWithBoxes starts with empty rows and performs native staged
// greedy selection, using checked signed CI arrivals and cloned box timing.
// C's nonnegative-candidate assertion rejects safely for sufficiently negative
// arrivals. Shared boxed traversal/work limits and node-level tie/capacity
// rules preserve native behavior. No-box fallback ignores terminal overrides,
// evaluates every mapped LUT and leaves timing epochs unchanged, as in C.
// Inputs remain unchanged; failures/cancellation expose no partial result.
func (mapping *GiaLUTMapping) SelectEdgesWithBoxes(m *GiaManager, timing *TimManager, useTwo bool, options GiaLUTAnalysisOptions) (GiaBoxEdgeSelection, error) {
	if mapping == nil || m == nil || timing == nil || options.MaxVisits < 0 {
		return GiaBoxEdgeSelection{}, ErrDimension
	}
	edges, _, err := newGiaEdges(len(m.objects), GiaEdgeOptions{MaxObjects: options.MaxObjects, Stop: options.Stop})
	if err != nil {
		return GiaBoxEdgeSelection{}, err
	}
	result, err := edges.walkBoxEdges(m, mapping, timing, options, true, useTwo)
	if err != nil {
		return GiaBoxEdgeSelection{}, err
	}
	return GiaBoxEdgeSelection{Edges: edges, Timing: result.Timing, Manager: result.Manager, Order: result.Order}, nil
}
