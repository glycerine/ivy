package abc

// Incremental packed-mapping timing from aig/gia/giaEdge.c:
// Gia_ManEvalWindowInc, using the shared Gia_ObjComputeEdgeDelay kernel.

import "slices"

type GiaIncrementalEdges struct {
	Edges           *GiaEdgeAssignment
	Timing          GiaEdgeTiming // global CO-driver maximum and positive-slot edge count
	AffectedMaximum int           // native incremental return value: maximum touched LUT delay
	Order           []GiaID       // native reversed structural fanout postorder, including unmapped gates
}

// RecomputeEdges ports the incremental packed-mapping recomputation phase.
// It copies prior rows/delays, visits the structural TFO of roots in native
// order, skips unmapped gates, and clears/reselects each visited LUT row.
// Old reciprocal entries in other rows are retained, matching C; results may
// be asymmetric. This is a recomputation of the supplied packed mapping, not
// an implicit mapping2 proposal swap (C prioritizes packed mapping here).
// MaxObjects bounds the arena/TFO roots and MaxVisits the recomputed fanin
// slots. CI/constant roots and extended gates are safe traversal extensions.
// Errors/cancellation expose no partial result and preserve every input.
func (mapping *GiaLUTMapping) RecomputeEdges(m *GiaManager, roots []GiaID, priorEdges *GiaEdgeAssignment, prior GiaEdgeTiming, useTwo bool, options GiaLUTAnalysisOptions) (GiaIncrementalEdges, error) {
	fail := func(err error) (GiaIncrementalEdges, error) { return GiaIncrementalEdges{}, err }
	if mapping == nil || m == nil || priorEdges == nil || len(priorEdges.rows) != len(m.objects) || len(prior.Delays) != len(m.objects) || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if len(m.objects) > options.MaxObjects || len(roots) > options.MaxObjects {
		return fail(ErrLimit)
	}
	if err := mapping.checkGraph(m); err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	for _, delay := range prior.Delays {
		if delay < 0 {
			return fail(ErrDimension)
		}
	}
	maxEdges, err := checkedSize(0, 3, len(m.objects), int(^uint(0)>>1))
	if err != nil {
		return fail(err)
	}
	order, err := m.CollectTransitiveFanouts(roots, options.MaxObjects, maxEdges)
	if err != nil {
		return fail(err)
	}
	slices.Reverse(order)
	result := GiaIncrementalEdges{Edges: &GiaEdgeAssignment{rows: priorEdges.Rows(), skipped: priorEdges.skipped}, Timing: GiaEdgeTiming{Delays: slices.Clone(prior.Delays)}, Order: order}
	visits := 0
	for _, id := range order {
		if stopped() {
			return fail(ErrCancelled)
		}
		if mapping.packed[id] == 0 {
			continue
		}
		if err := mapping.selectEdgeNode(id, result.Edges, &result.Timing, useTwo, &visits, options.MaxVisits, options.Stop); err != nil {
			return fail(err)
		}
		result.AffectedMaximum = max(result.AffectedMaximum, result.Timing.Delays[id])
	}
	for _, id := range m.cos {
		if stopped() {
			return fail(ErrCancelled)
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return fail(err)
		}
		result.Timing.Maximum = max(result.Timing.Maximum, result.Timing.Delays[driver.ID()])
	}
	positive := 0
	for _, row := range result.Edges.rows {
		for _, id := range row {
			if id > 0 {
				positive++
			}
		}
	}
	result.Timing.Edges = positive / 2
	if stopped() {
		return fail(ErrCancelled)
	}
	return result, nil
}
