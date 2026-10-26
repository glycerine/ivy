package abc

// Shared-PI composition from src/aig/gia/giaDup.c:Gia_ManDupAppendNew.

// GiaComposition maps each source regular object to a target literal. CO
// entries point to target CO objects. Register order is first then second.
type GiaComposition struct {
	Graph         *GiaManager
	First, Second []GiaLit
}

// ComposeGiaSharedInputs combines equal-PI-count graphs, sharing PIs by index
// and retaining distinct register banks. Ordinary POs are first's then second's,
// followed by both constraint suffixes and both RI banks. This preserves Go
// constraint roles, which C drops. Name/spec come from first. All logic is
// hashed (extended nodes are lowered), and source scratch remains untouched.
func ComposeGiaSharedInputs(first, second *GiaManager, options GiaCofactorOptions) (GiaComposition, error) {
	fail := func(e error) (GiaComposition, error) { return GiaComposition{}, e }
	if first == nil || second == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	for _, m := range []*GiaManager{first, second} {
		if e := m.Check(); e != nil {
			return fail(e)
		}
	}
	if first.PrimaryInputCount() != second.PrimaryInputCount() {
		return fail(ErrDimension)
	}
	if _, e := checkedSize(len(first.objects), 1, len(second.objects), options.MaxMappingEntries); e != nil {
		return fail(e)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	g.Name, g.Spec = first.Name, first.Spec
	if err = g.HashStart(); err != nil {
		return fail(err)
	}
	maps := [2][]GiaLit{}
	for source, m := range []*GiaManager{first, second} {
		maps[source] = make([]GiaLit, len(m.objects))
		mapping := maps[source]
		for i := range mapping {
			mapping[i] = InvalidGiaLit
		}
		mapping[0] = GiaConst0
		for i, obj := range m.objects {
			if stopped() {
				return fail(ErrCancelled)
			}
			if obj.IsCI() {
				index, e := obj.TerminalIndex()
				if e != nil {
					return fail(e)
				}
				if source == 1 && index < m.PrimaryInputCount() {
					mapping[i] = maps[0][first.cis[index]]
				} else {
					mapping[i], err = g.AppendCI()
				}
			} else if obj.IsAnd() {
				mapping[i], err = hashGiaCopiedGate(m, g, GiaID(i), mapping)
			}
			if err != nil {
				return fail(err)
			}
		}
	}
	g.HashStop()
	// Three role groups preserve a single constraint suffix and trailing RIs.
	for role := 0; role < 3; role++ {
		for source, m := range []*GiaManager{first, second} {
			ordinary := m.PrimaryOutputCount() - m.constraints
			lo, hi := 0, ordinary
			if role == 1 {
				lo, hi = ordinary, m.PrimaryOutputCount()
			} else if role == 2 {
				lo, hi = m.PrimaryOutputCount(), len(m.cos)
			}
			for _, id := range m.cos[lo:hi] {
				if stopped() {
					return fail(ErrCancelled)
				}
				driver, e := m.objects[id].Fanin0(id)
				if e != nil {
					return fail(e)
				}
				copied, e := mappedGiaSignal(maps[source], driver)
				if e != nil {
					return fail(e)
				}
				maps[source][id], err = g.AppendCO(copied)
				if err != nil {
					return fail(err)
				}
			}
		}
	}
	g.registers = first.registers + second.registers
	g.constraints = first.constraints + second.constraints
	if err = g.Check(); err != nil {
		return fail(err)
	}
	return GiaComposition{g, maps[0], maps[1]}, nil
}
