package abc

// Edge timing from aig/gia/giaEdge.c: Gia_ObjEvalEdgeDelay,
// Gia_ManEvalEdgeDelay (plain mapping, no timing boxes), Gia_ManEvalEdgeCount.

type GiaEdgeTiming struct {
	Delays  []int // native object array; CO entries stay zero without timing boxes
	Maximum int   // maximum CO-driver delay, excluding dangling LUTs
	Edges   int   // positive directed slots / 2, truncated as in C
}

// EvaluateDelay applies native fast/ordinary delays 2/10 to mapped fanins.
// Lookup uses the LUT root's row, including asymmetric and zero-sentinel
// behavior. At most four fanin slots per LUT are supported, as in C. Unmapped
// objects start at zero; no mapping completeness is presumed. This is the
// native no-box mapping path; timing boxes and mapping2 remain separate work.
// Returned scratch is owned. Errors/cancellation return no partial result.
func (e *GiaEdgeAssignment) EvaluateDelay(m *GiaManager, mapping *GiaLUTMapping, options GiaLUTAnalysisOptions) (GiaEdgeTiming, error) {
	fail := func(err error) (GiaEdgeTiming, error) { return GiaEdgeTiming{}, err }
	if e == nil || m == nil || mapping == nil || len(e.rows) != len(m.objects) || options.MaxObjects < 0 || options.MaxVisits < 0 {
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
	r := GiaEdgeTiming{Delays: make([]int, len(m.objects))}
	visits, positive := 0, 0
	for _, row := range e.rows {
		for _, id := range row {
			if id > 0 {
				positive++
			}
		}
	}
	r.Edges = positive / 2
	for _, root := range mapping.Roots() {
		if stopped() {
			return fail(ErrCancelled)
		}
		if err := e.evaluateEdgeNode(mapping, root, &r, &visits, options.MaxVisits, options.Stop); err != nil {
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
		r.Maximum = max(r.Maximum, r.Delays[driver.ID()])
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return r, nil
}

func giaEdgeDelayAdd(value, step int) (int, error) {
	sum := int64(value) + int64(step)
	if sum < -2147483648 || sum > 2147483647 {
		return 0, ErrLimit
	}
	return int(sum), nil
}
func (e *GiaEdgeAssignment) evaluateEdgeNode(mapping *GiaLUTMapping, root GiaID, r *GiaEdgeTiming, visits *int, maxVisits int, stop func() bool) error {
	offset := mapping.packed[root]
	count := mapping.packed[offset]
	if count > 4 {
		return ErrUnsupported
	}
	for _, fanin := range mapping.packed[offset+1 : offset+1+count] {
		if stop != nil && stop() {
			return ErrCancelled
		}
		if *visits == maxVisits {
			return ErrLimit
		}
		*visits++
		step := 10
		if e.rows[root][0] == GiaID(fanin) || e.rows[root][1] == GiaID(fanin) {
			step = 2
		}
		delay, err := giaEdgeDelayAdd(r.Delays[fanin], step)
		if err != nil {
			return err
		}
		r.Delays[root] = max(r.Delays[root], delay)
	}
	return nil
}
