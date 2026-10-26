package abc

// Inductive-step miter from src/aig/gia/giaDup.c:Gia_ManDupInvMiter.

// GiaInvariantMiter builds a combinational output I(state) AND !I(nextState).
// Invariant inputs follow transition register order; target inputs follow all
// transition CIs (PIs then state). Like C, source property/constraint outputs
// are ignored. This checks unconstrained induction only, not initiation or
// property implication; UNSAT alone is not a complete safety certificate.
// All CIs and name/spec survive cleanup. Source scratch is untouched, extended
// gates are lowered, and constant invariant roots are explicitly supported.
func GiaInvariantMiter(transition, invariant *GiaManager, options GiaCofactorOptions) (*GiaManager, error) {
	if transition == nil || invariant == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	for _, m := range []*GiaManager{transition, invariant} {
		if e := m.Check(); e != nil {
			return nil, e
		}
	}
	if transition.registers < 1 || invariant.registers != 0 || len(invariant.cos) != 1 || invariant.constraints != 0 || len(invariant.cis) != transition.registers {
		return nil, ErrDimension
	}
	if _, e := checkedSize(len(transition.objects), 2, len(invariant.objects), options.MaxMappingEntries); e != nil {
		return nil, e
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = transition.Name, transition.Spec
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	mapping := make([]GiaLit, len(transition.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for i, obj := range transition.objects {
		if stopped() {
			return nil, ErrCancelled
		}
		if obj.IsCI() {
			mapping[i], err = g.AppendCI()
		} else if obj.IsAnd() {
			mapping[i], err = hashGiaCopiedGate(transition, g, GiaID(i), mapping)
		} else if obj.IsCO() {
			driver, e := obj.Fanin0(GiaID(i))
			if e != nil {
				return nil, e
			}
			mapping[i], err = mappedGiaSignal(mapping, driver)
		}
		if err != nil {
			return nil, err
		}
	}
	values := [2]GiaLit{}
	for frame := 0; frame < 2; frame++ {
		copies := make([]GiaLit, len(invariant.objects))
		for i := range copies {
			copies[i] = InvalidGiaLit
		}
		copies[0] = GiaConst0
		for i, id := range invariant.cis {
			if frame == 0 {
				copies[id] = mapping[transition.cis[transition.PrimaryInputCount()+i]]
			} else {
				copies[id] = mapping[transition.cos[transition.PrimaryOutputCount()+i]]
			}
		}
		for i, obj := range invariant.objects {
			if !obj.IsAnd() {
				continue
			}
			if stopped() {
				return nil, ErrCancelled
			}
			copies[i], err = hashGiaCopiedGate(invariant, g, GiaID(i), copies)
			if err != nil {
				return nil, err
			}
		}
		co := invariant.cos[0]
		driver, e := invariant.objects[co].Fanin0(co)
		if e != nil {
			return nil, e
		}
		values[frame], err = mappedGiaSignal(copies, driver)
		if err != nil {
			return nil, err
		}
	}
	if stopped() {
		return nil, ErrCancelled
	}
	violation, err := g.HashAnd(values[0], values[1].Not())
	if err != nil {
		return nil, err
	}
	if _, err = g.AppendCO(violation); err != nil {
		return nil, err
	}
	g.HashStop()
	result, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return result, nil
}
