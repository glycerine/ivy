package abc

// Box-aware plain-mapping evaluation from aig/gia/giaEdge.c:
// Gia_ManEvalEdgeDelay timing-box branch.

type GiaBoxEdgeTiming struct {
	Timing  GiaEdgeTiming
	Manager *TimManager // owned arrival/traversal updates, committed only on success
	Order   []GiaID
}

// EvaluateDelayWithBoxes evaluates staged GIA/LUT/box timing on an owned
// timing-manager clone. CI arrivals use native checked int32 truncation; CO
// driver delays update timing terminals while graph CO delay entries stay
// zero. Missing/unmapped LUTs retain zero graph delays, as in C. Without boxes,
// native no-box evaluation ignores terminal arrival overrides and epochs.
// MaxVisits bounds copied table values, returned order entries, LUT fanin
// slots and box arcs actually evaluated, combined. MaxObjects also bounds
// sparse table slots. Inputs stay unchanged on success, errors or cancellation.
func (e *GiaEdgeAssignment) EvaluateDelayWithBoxes(m *GiaManager, mapping *GiaLUTMapping, timing *TimManager, options GiaLUTAnalysisOptions) (GiaBoxEdgeTiming, error) {
	return e.walkBoxEdges(m, mapping, timing, options, false, false)
}

func (e *GiaEdgeAssignment) walkBoxEdges(m *GiaManager, mapping *GiaLUTMapping, timing *TimManager, options GiaLUTAnalysisOptions, selectEdges, useTwo bool) (GiaBoxEdgeTiming, error) {
	fail := func(err error) (GiaBoxEdgeTiming, error) { return GiaBoxEdgeTiming{}, err }
	if e == nil || m == nil || mapping == nil || timing == nil || len(e.rows) != len(m.objects) || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if len(m.objects) > options.MaxObjects || len(timing.tables) > options.MaxObjects || timing.tableValues > options.MaxVisits {
		return fail(ErrLimit)
	}
	if len(timing.cis) != len(m.cis) || len(timing.cos) != len(m.cos) {
		return fail(ErrDimension)
	}
	if err := mapping.checkGraph(m); err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	if timing.BoxCount() == 0 {
		used := timing.tableValues
		for _, id := range mapping.Roots() {
			var err error
			used, err = checkedSize(used, 1, mapping.packed[mapping.packed[id]], options.MaxVisits)
			if err != nil {
				return fail(err)
			}
		}
		childOptions := options
		childOptions.MaxVisits = max(1, options.MaxVisits-timing.tableValues)
		var r GiaEdgeTiming
		var err error
		if selectEdges {
			var selected *GiaEdgeAssignment
			selected, r, err = mapping.SelectEdges(m, useTwo, childOptions)
			if err == nil {
				e.rows = selected.rows
				e.skipped = selected.skipped
			}
		} else {
			r, err = e.EvaluateDelay(m, mapping, childOptions)
		}
		if err != nil {
			return fail(err)
		}
		return GiaBoxEdgeTiming{Timing: r, Manager: timing.Clone()}, nil
	}
	order, err := m.OrderWithBoxes(timing, options)
	if err != nil {
		return fail(err)
	}
	visits, err := checkedSize(timing.tableValues, 1, len(order), options.MaxVisits)
	if err != nil {
		return fail(err)
	}
	result := GiaBoxEdgeTiming{Timing: GiaEdgeTiming{Delays: make([]int, len(m.objects))}, Manager: timing.Clone(), Order: order}
	manager := result.Manager
	if err = manager.IncrementTraversal(); err != nil {
		return fail(err)
	}
	for _, id := range order {
		if stopped() {
			return fail(ErrCancelled)
		}
		obj := m.objects[id]
		if obj.IsAnd() {
			if mapping.packed[id] != 0 {
				if selectEdges {
					err = mapping.selectEdgeNode(id, e, &result.Timing, useTwo, &visits, options.MaxVisits, options.Stop)
				} else {
					err = e.evaluateEdgeNode(mapping, id, &result.Timing, &visits, options.MaxVisits, options.Stop)
				}
				if err != nil {
					return fail(err)
				}
			}
		} else if obj.IsCI() {
			index := int(obj.Diff1())
			terminal := manager.cis[index]
			if terminal.Box >= 0 && (!manager.useTraversal || terminal.TravID != manager.epoch) {
				table, err := manager.boxTable(terminal.Box)
				if err != nil {
					return fail(err)
				}
				visits, err = checkedSize(visits, 1, len(table.Delays), options.MaxVisits)
				if err != nil {
					return fail(err)
				}
			}
			arrival, err := manager.CIArrival(index)
			if err != nil {
				return fail(err)
			}
			value, err := timArrivalInteger(arrival)
			if err != nil {
				return fail(err)
			}
			result.Timing.Delays[id] = int(value)
		} else if obj.IsCO() {
			driver, err := obj.Fanin0(id)
			if err != nil {
				return fail(err)
			}
			if err = manager.SetCOArrival(int(obj.Diff1()), float32(result.Timing.Delays[driver.ID()])); err != nil {
				return fail(err)
			}
		} else if id != 0 {
			return fail(ErrInvariant)
		}
	}
	for _, id := range m.cos {
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return fail(err)
		}
		result.Timing.Maximum = max(result.Timing.Maximum, result.Timing.Delays[driver.ID()])
	}
	positive := 0
	for _, row := range e.rows {
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
