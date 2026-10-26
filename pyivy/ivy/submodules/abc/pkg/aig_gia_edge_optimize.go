package abc

// Unit-edge improvement from aig/gia/giaEdge.c:
// Edg_ObjImprove and Edg_ManAssignEdgeNew.

import "sort"

type GiaEdgeOptimizationOptions struct {
	EdgeLimit     int // zero: two; one or two slots per object
	MaxObjects    int // zero: one million
	MaxVisits     int // aggregate construction, timing and candidate slot work; zero: one million
	MaxIterations int // zero: native 10,000 cap; smaller caps fail if exhausted
	Stop          func() bool
}
type GiaEdgeOptimization struct {
	Edges                    *GiaEdgeAssignment
	Timing                   GiaUnitEdgeTiming
	Iterations, Improvements int
}

type giaEdgeWork struct {
	used, limit int
	stop        func() bool
}

func (w *giaEdgeWork) visit() error {
	if w.stop != nil && w.stop() {
		return ErrCancelled
	}
	if w.used == w.limit {
		return ErrLimit
	}
	w.used++
	return nil
}
func (n *GiaLUTNetwork) improveEdgeNode(m *GiaManager, e *GiaEdgeAssignment, t GiaUnitEdgeTiming, root GiaID, limit int, w *giaEdgeWork) (bool, error) {
	fanins, fanouts := make([]GiaID, 0, 4), make([]GiaID, 0, limit)
	inputEC, outputEC, edges, outputCritical := 0, 0, 0, 0
	inspect := func(neighbor GiaID, opposite int, input bool) error {
		if err := w.visit(); err != nil {
			return err
		}
		if input && !m.objects[neighbor].IsAnd() {
			return nil
		}
		if !input && !m.objects[neighbor].IsAnd() {
			return ErrInvariant
		}
		marked := e.rows[root][0] == neighbor || e.rows[root][1] == neighbor
		local := t.Forward[root]
		if input {
			local = t.Reverse[root]
		}
		cost := 1
		if marked {
			cost = 0
			edges++
		}
		if local > t.Maximum || opposite > t.Maximum-local || cost > t.Maximum-local-opposite {
			return ErrInvariant
		}
		if local+opposite+cost == t.Maximum {
			if marked {
				if input {
					inputEC++
				} else {
					outputEC++
				}
			} else if input {
				fanins = append(fanins, neighbor)
			} else {
				if outputCritical < limit {
					fanouts = append(fanouts, neighbor)
				}
				outputCritical++
			}
		}
		return nil
	}
	for _, id := range n.fanins[root] {
		if err := inspect(id, t.Forward[id], true); err != nil {
			return false, err
		}
	}
	for _, id := range n.fanouts[root] {
		if err := inspect(id, t.Reverse[id], false); err != nil {
			return false, err
		}
	}
	if edges > limit {
		return false, ErrInvariant
	}
	if edges == limit {
		return false, nil
	}
	add := func(neighbors []GiaID) (bool, error) {
		for _, id := range neighbors {
			count, _ := e.Count(id)
			if count == limit {
				return false, nil
			}
		}
		for _, id := range neighbors {
			e.add(root, id)
			e.add(id, root)
		}
		return true, nil
	}
	if inputEC == 0 && len(fanins) > 0 && len(fanins) <= limit-edges {
		if ok, err := add(fanins); ok || err != nil {
			return ok, err
		}
	}
	if outputEC == 0 && outputCritical > 0 && outputCritical <= limit-edges {
		return add(fanouts)
	}
	return false, nil
}

// OptimizeEdges implements native plain-mapping unit-edge improvement. It
// starts empty, scans critical levels by increasing population, and tries
// fanins before fanouts. Equal-sized levels retain original level order for
// deterministic ties (matching the pinned reference). Only strict delay
// improvements save a backup; plateau insertions are discarded on return.
// Native assertion preconditions reject safely. Inputs remain unchanged.
func (mapping *GiaLUTMapping) OptimizeEdges(m *GiaManager, options GiaEdgeOptimizationOptions) (GiaEdgeOptimization, error) {
	fail := func(err error) (GiaEdgeOptimization, error) { return GiaEdgeOptimization{}, err }
	if mapping == nil || m == nil || options.EdgeLimit < 0 || options.EdgeLimit > 2 || options.MaxObjects < 0 || options.MaxVisits < 0 || options.MaxIterations < 0 {
		return fail(ErrDimension)
	}
	if options.EdgeLimit == 0 {
		options.EdgeLimit = 2
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 10000
	}
	if len(m.objects) > options.MaxObjects {
		return fail(ErrLimit)
	}
	if err := mapping.checkGraph(m); err != nil {
		return fail(err)
	}
	n, err := mapping.Network(GiaLUTAnalysisOptions{MaxObjects: options.MaxObjects, MaxVisits: options.MaxVisits, Stop: options.Stop})
	if err != nil {
		return fail(err)
	}
	w := giaEdgeWork{limit: options.MaxVisits, stop: options.Stop}
	evaluationWork := 0
	for _, id := range n.Roots() {
		w.used += len(n.fanins[id])
		evaluationWork += len(n.fanins[id]) + len(n.fanouts[id])
	}
	evaluate := func(e *GiaEdgeAssignment) (GiaUnitEdgeTiming, error) {
		if evaluationWork > w.limit-w.used {
			return GiaUnitEdgeTiming{}, ErrLimit
		}
		r, err := n.EvaluateEdges(e, GiaLUTAnalysisOptions{MaxObjects: options.MaxObjects, MaxVisits: max(1, evaluationWork), Stop: options.Stop})
		if err == nil {
			w.used += evaluationWork
		}
		return r, err
	}
	current := &GiaEdgeAssignment{rows: make([][2]GiaID, len(m.objects))}
	var best *GiaEdgeAssignment
	previous, lastImprovement := int(^uint(0)>>1), -1
	result := GiaEdgeOptimization{}
	for iteration := 0; iteration < 10000; iteration++ {
		if iteration > lastImprovement+50 {
			break
		}
		if iteration == options.MaxIterations {
			return fail(ErrLimit)
		}
		timing, err := evaluate(current)
		if err != nil {
			return fail(err)
		}
		if timing.Maximum != timing.ReverseMaximum+1 {
			return fail(ErrInvariant)
		}
		result.Iterations++
		if timing.Maximum < previous {
			best = &GiaEdgeAssignment{rows: current.Rows(), skipped: current.skipped}
			previous = timing.Maximum
			lastImprovement = iteration
		}
		levels, err := n.CriticalLevels(timing)
		if err != nil {
			return fail(err)
		}
		sort.SliceStable(levels, func(i, j int) bool { return len(levels[i]) < len(levels[j]) })
		changed := false
		for _, level := range levels {
			for _, root := range level {
				ok, err := n.improveEdgeNode(m, current, timing, root, options.EdgeLimit, &w)
				if err != nil {
					return fail(err)
				}
				if ok {
					changed = true
					break
				}
			}
			if changed {
				break
			}
		}
		if !changed {
			break
		}
		result.Improvements++
	}
	if best == nil {
		return fail(ErrInvariant)
	}
	result.Edges = best
	result.Timing, err = evaluate(best)
	if err != nil {
		return fail(err)
	}
	if options.Stop != nil && options.Stop() {
		return fail(ErrCancelled)
	}
	return result, nil
}
