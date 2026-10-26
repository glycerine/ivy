package abc

// Greedy fast-edge selection from aig/gia/giaEdge.c:
// Gia_ObjComputeEdgeDelay and Gia_ManComputeEdgeDelay (plain/no-box path).

// SelectEdges computes native greedy fast-edge rows and corresponding delays.
// First-level LUTs retain delay 10. Later LUTs select one critical fanin, or
// two tied slots when useTwo is true, only when each fanin has capacity.
// Fanin order, duplicate slots, native runner-up tracking and independent
// directed insertion are preserved. LUT widths must be one through four.
// This is the plain-mapping/no-box path; source graph/mapping stay unchanged.
func (mapping *GiaLUTMapping) SelectEdges(m *GiaManager, useTwo bool, options GiaLUTAnalysisOptions) (*GiaEdgeAssignment, GiaEdgeTiming, error) {
	fail := func(err error) (*GiaEdgeAssignment, GiaEdgeTiming, error) { return nil, GiaEdgeTiming{}, err }
	if mapping == nil || m == nil || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if len(m.objects) > options.MaxObjects {
		return fail(ErrLimit)
	}
	if err := mapping.checkGraph(m); err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	edges := &GiaEdgeAssignment{rows: make([][2]GiaID, len(m.objects))}
	timing := GiaEdgeTiming{Delays: make([]int, len(m.objects))}
	visits := 0
	for _, root := range mapping.Roots() {
		if stopped() {
			return fail(ErrCancelled)
		}
		if err := mapping.selectEdgeNode(root, edges, &timing, useTwo, &visits, options.MaxVisits, options.Stop); err != nil {
			return fail(err)
		}
	}
	for _, id := range m.cos {
		if stopped() {
			return fail(ErrCancelled)
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return fail(err)
		}
		timing.Maximum = max(timing.Maximum, timing.Delays[driver.ID()])
	}
	positive := 0
	for _, row := range edges.rows {
		for _, id := range row {
			if id > 0 {
				positive++
			}
		}
	}
	timing.Edges = positive / 2
	if stopped() {
		return fail(ErrCancelled)
	}
	return edges, timing, nil
}

func (mapping *GiaLUTMapping) selectEdgeNode(root GiaID, edges *GiaEdgeAssignment, timing *GiaEdgeTiming, useTwo bool, visits *int, maxVisits int, stop func() bool) error {
	edges.rows[root] = [2]GiaID{}
	offset := mapping.packed[root]
	count := mapping.packed[offset]
	if count < 1 || count > 4 {
		return ErrUnsupported
	}
	maximum, second, ties := 0, 0, 0
	var firstFan, secondFan GiaID
	for _, id := range mapping.packed[offset+1 : offset+1+count] {
		if stop != nil && stop() {
			return ErrCancelled
		}
		if *visits == maxVisits {
			return ErrLimit
		}
		*visits++
		delay, err := giaEdgeDelayAdd(timing.Delays[id], 10)
		if err != nil {
			return err
		}
		if delay > maximum {
			second = maximum
			maximum = delay
			firstFan = GiaID(id)
			ties = 1
		} else if delay == maximum {
			secondFan = GiaID(id)
			ties++
			if !useTwo {
				second = maximum
			}
		} else {
			second = max(second, delay)
		}
	}
	if ties == 0 {
		return ErrInvariant
	}
	if maximum > 10 {
		capacity := func(id GiaID) bool { count, _ := edges.Count(id); return count <= 1 }
		if ties == 1 && capacity(firstFan) {
			edges.add(firstFan, root)
			edges.add(root, firstFan)
			maximum = max(second, maximum-8)
		} else if useTwo && ties == 2 && capacity(firstFan) && capacity(secondFan) {
			edges.add(firstFan, root)
			edges.add(secondFan, root)
			edges.add(root, firstFan)
			edges.add(root, secondFan)
			maximum = max(second, maximum-8)
		}
	}
	timing.Delays[root] = maximum
	return nil
}
