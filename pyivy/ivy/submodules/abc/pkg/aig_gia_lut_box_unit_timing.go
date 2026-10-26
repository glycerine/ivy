package abc

// Boxed unit timing from aig/gia/giaSatLut.c:Sbl_ManCreateTiming, with
// corrected reverse loop indexing and corrected timing-table arc guards.

type GiaBoxLUTUnitTiming struct {
	Timing  GiaLUTUnitTiming
	Manager *TimManager
	Order   []GiaID
}

// UnitTimingWithBoxes propagates arrivals forward and required times backward
// on an owned timing clone. It retains native delayStart/global visited-LUT
// maxima, CO-driver required assignment, integer truncation and zero/eternity
// CO array entries. Separate iteration variables correct C's reverse traversal
// counter clobbering. Box required propagation uses corrected per-arc guards.
// MaxVisits combines copied table values, both order passes, mapped fanin
// slots and evaluated box arcs. Missing LUTs retain initial graph values.
func (n *GiaLUTNetwork) UnitTimingWithBoxes(m *GiaManager, timing *TimManager, delayStart int, options GiaLUTAnalysisOptions) (GiaBoxLUTUnitTiming, error) {
	fail := func(err error) (GiaBoxLUTUnitTiming, error) { return GiaBoxLUTUnitTiming{}, err }
	if n == nil || m == nil || timing == nil || len(n.fanins) != len(m.objects) || len(timing.cis) != len(m.cis) || len(timing.cos) != len(m.cos) || options.MaxObjects < 0 || options.MaxVisits < 0 || int64(delayStart) < -2147483648 || int64(delayStart) > 2147483647 {
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
	if err := m.Check(); err != nil {
		return fail(err)
	}
	for _, id := range n.Roots() {
		if !m.objects[id].IsAnd() {
			return fail(ErrDimension)
		}
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	if timing.BoxCount() == 0 {
		used := timing.tableValues
		for _, id := range n.Roots() {
			var err error
			used, err = checkedSize(used, 2, len(n.fanins[id]), options.MaxVisits)
			if err != nil {
				return fail(err)
			}
		}
		child := options
		child.MaxVisits = max(1, options.MaxVisits-timing.tableValues)
		r, err := n.UnitTiming(m, delayStart, child)
		if err != nil {
			return fail(err)
		}
		return GiaBoxLUTUnitTiming{Timing: r, Manager: timing.Clone()}, nil
	}
	order, err := m.OrderWithBoxes(timing, options)
	if err != nil {
		return fail(err)
	}
	work, err := checkedSize(timing.tableValues, 2, len(order), options.MaxVisits)
	if err != nil {
		return fail(err)
	}
	charge := func(count int) error {
		if stopped() {
			return ErrCancelled
		}
		next, err := checkedSize(work, 1, count, options.MaxVisits)
		if err == nil {
			work = next
		}
		return err
	}
	r := GiaBoxLUTUnitTiming{Timing: GiaLUTUnitTiming{Arrivals: make([]int, len(m.objects)), Required: make([]int, len(m.objects)), Maximum: delayStart}, Manager: timing.Clone(), Order: order}
	manager := r.Manager
	if err = manager.IncrementTraversal(); err != nil {
		return fail(err)
	}
	for _, id := range order {
		if stopped() {
			return fail(ErrCancelled)
		}
		obj := m.objects[id]
		if obj.IsAnd() {
			if len(n.fanins[id]) > 0 {
				for _, fanin := range n.fanins[id] {
					if err = charge(1); err != nil {
						return fail(err)
					}
					value, err := giaEdgeDelayAdd(r.Timing.Arrivals[fanin], 1)
					if err != nil {
						return fail(err)
					}
					r.Timing.Arrivals[id] = max(r.Timing.Arrivals[id], value)
				}
				r.Timing.Maximum = max(r.Timing.Maximum, r.Timing.Arrivals[id])
			}
		} else if obj.IsCI() {
			index := int(obj.Diff1())
			p := manager.cis[index]
			if p.Box >= 0 && (!manager.useTraversal || p.TravID != manager.epoch) {
				table, err := manager.boxTable(p.Box)
				if err != nil {
					return fail(err)
				}
				if err = charge(len(table.Delays)); err != nil {
					return fail(err)
				}
			}
			value, err := manager.CIArrival(index)
			if err != nil {
				return fail(err)
			}
			arrival, err := timArrivalInteger(value)
			if err != nil {
				return fail(err)
			}
			r.Timing.Arrivals[id] = int(arrival)
		} else if obj.IsCO() {
			driver, err := obj.Fanin0(id)
			if err != nil {
				return fail(err)
			}
			if err = manager.SetCOArrival(int(obj.Diff1()), float32(r.Timing.Arrivals[driver.ID()])); err != nil {
				return fail(err)
			}
		}
	}
	for i := range r.Timing.Required {
		r.Timing.Required[i] = 1000000000
	}
	for _, id := range m.cos {
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return fail(err)
		}
		r.Timing.Required[driver.ID()] = min(r.Timing.Required[driver.ID()], r.Timing.Maximum)
	}
	if err = manager.IncrementTraversal(); err != nil {
		return fail(err)
	}
	if err = manager.InitPORequiredAll(float32(r.Timing.Maximum)); err != nil {
		return fail(err)
	}
	for position := len(order) - 1; position >= 0; position-- {
		if stopped() {
			return fail(ErrCancelled)
		}
		id := order[position]
		obj := m.objects[id]
		if obj.IsAnd() {
			if len(n.fanins[id]) > 0 {
				value, err := giaEdgeDelayAdd(r.Timing.Required[id], -1)
				if err != nil {
					return fail(err)
				}
				for _, fanin := range n.fanins[id] {
					if err = charge(1); err != nil {
						return fail(err)
					}
					r.Timing.Required[fanin] = min(r.Timing.Required[fanin], value)
				}
			}
		} else if obj.IsCI() {
			if err = manager.SetCIRequired(int(obj.Diff1()), float32(r.Timing.Required[id])); err != nil {
				return fail(err)
			}
		} else if obj.IsCO() {
			index := int(obj.Diff1())
			p := manager.cos[index]
			if p.Box >= 0 && (!manager.useTraversal || p.TravID != manager.epoch) {
				table, err := manager.boxTable(p.Box)
				if err != nil {
					return fail(err)
				}
				if err = charge(len(table.Delays)); err != nil {
					return fail(err)
				}
			}
			value, err := manager.CORequired(index)
			if err != nil {
				return fail(err)
			}
			required, err := timArrivalInteger(value)
			if err != nil {
				return fail(err)
			}
			driver, err := obj.Fanin0(id)
			if err != nil {
				return fail(err)
			}
			r.Timing.Required[driver.ID()] = int(required)
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return r, nil
}
