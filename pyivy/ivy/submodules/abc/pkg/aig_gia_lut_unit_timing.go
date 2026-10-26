package abc

// Unboxed unit arrival/required timing from aig/gia/giaSatLut.c:
// Sbl_ManComputeDelay and Sbl_ManCreateTiming.

type GiaLUTUnitTiming struct {
	Arrivals, Required []int
	Maximum            int
}

// UnitTiming computes native unboxed unit timing. Maximum starts at delayStart
// and includes every mapped LUT, including dangling logic. Required times start
// at 1e9, CO drivers are downdated with Maximum, and every mapped LUT propagates
// required-1 backward, even on unobservable cones. CIs start at zero arrival;
// CO array entries retain their initial values. MaxVisits bounds combined
// forward/reverse fanin slots; MaxObjects bounds the arena. Checked arithmetic
// and cancellation expose no partial result or source changes.
func (n *GiaLUTNetwork) UnitTiming(m *GiaManager, delayStart int, options GiaLUTAnalysisOptions) (GiaLUTUnitTiming, error) {
	fail := func(err error) (GiaLUTUnitTiming, error) { return GiaLUTUnitTiming{}, err }
	if n == nil || m == nil || len(n.fanins) != len(m.objects) || options.MaxObjects < 0 || options.MaxVisits < 0 || int64(delayStart) < -2147483648 || int64(delayStart) > 2147483647 {
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
	if err := m.Check(); err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	roots := n.Roots()
	for _, id := range roots {
		if !m.objects[id].IsAnd() {
			return fail(ErrDimension)
		}
	}
	r := GiaLUTUnitTiming{Arrivals: make([]int, len(m.objects)), Required: make([]int, len(m.objects)), Maximum: delayStart}
	for i := range r.Required {
		r.Required[i] = 1000000000
	}
	visits := 0
	for _, id := range roots {
		if stopped() {
			return fail(ErrCancelled)
		}
		for _, fanin := range n.fanins[id] {
			if stopped() {
				return fail(ErrCancelled)
			}
			if visits == options.MaxVisits {
				return fail(ErrLimit)
			}
			visits++
			value, err := giaEdgeDelayAdd(r.Arrivals[fanin], 1)
			if err != nil {
				return fail(err)
			}
			r.Arrivals[id] = max(r.Arrivals[id], value)
		}
		r.Maximum = max(r.Maximum, r.Arrivals[id])
	}
	for _, id := range m.cos {
		if stopped() {
			return fail(ErrCancelled)
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return fail(err)
		}
		r.Required[driver.ID()] = min(r.Required[driver.ID()], r.Maximum)
	}
	for i := len(roots) - 1; i >= 0; i-- {
		if stopped() {
			return fail(ErrCancelled)
		}
		id := roots[i]
		value, err := giaEdgeDelayAdd(r.Required[id], -1)
		if err != nil {
			return fail(err)
		}
		for _, fanin := range n.fanins[id] {
			if stopped() {
				return fail(ErrCancelled)
			}
			if visits == options.MaxVisits {
				return fail(ErrLimit)
			}
			visits++
			r.Required[fanin] = min(r.Required[fanin], value)
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return r, nil
}
