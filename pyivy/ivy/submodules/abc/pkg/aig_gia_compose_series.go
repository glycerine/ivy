package abc

// Serial combinational composition from src/aig/gia/giaDup.c:Gia_ManDupOntop.

// ComposeGiaSeries feeds bottom COs into top CIs by terminal index. Both
// graphs must be combinational, with bottom output count equal to top input
// count. Bottom constraints are rejected because consuming assumption outputs
// as data needs caller policy; top constraint roles are preserved. Result CIs
// follow bottom, COs follow top, and name/spec come from bottom. Extended gates
// are lowered through hashing; owned scratch and cleanup leave sources intact.
func ComposeGiaSeries(bottom, top *GiaManager, options GiaCofactorOptions) (*GiaManager, error) {
	if bottom == nil || top == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	for _, m := range []*GiaManager{bottom, top} {
		if e := m.Check(); e != nil {
			return nil, e
		}
	}
	if bottom.registers != 0 || top.registers != 0 || bottom.constraints != 0 || len(bottom.cos) != len(top.cis) {
		return nil, ErrDimension
	}
	if _, e := checkedSize(len(bottom.objects), 1, len(top.objects), options.MaxMappingEntries); e != nil {
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
	g.Name, g.Spec = bottom.Name, bottom.Spec
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	lower := make([]GiaLit, len(bottom.objects))
	upper := make([]GiaLit, len(top.objects))
	for _, mapping := range [][]GiaLit{lower, upper} {
		for i := range mapping {
			mapping[i] = InvalidGiaLit
		}
		mapping[0] = GiaConst0
	}
	for _, id := range bottom.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		lower[id], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	for i, obj := range bottom.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		lower[i], err = hashGiaCopiedGate(bottom, g, GiaID(i), lower)
		if err != nil {
			return nil, err
		}
	}
	for i, id := range bottom.cos {
		driver, e := bottom.objects[id].Fanin0(id)
		if e != nil {
			return nil, e
		}
		upper[top.cis[i]], err = mappedGiaSignal(lower, driver)
		if err != nil {
			return nil, err
		}
	}
	for i, obj := range top.objects {
		if !obj.IsAnd() {
			continue
		}
		if stopped() {
			return nil, ErrCancelled
		}
		upper[i], err = hashGiaCopiedGate(top, g, GiaID(i), upper)
		if err != nil {
			return nil, err
		}
	}
	for _, id := range top.cos {
		if stopped() {
			return nil, ErrCancelled
		}
		driver, e := top.objects[id].Fanin0(id)
		if e != nil {
			return nil, e
		}
		l, e := mappedGiaSignal(upper, driver)
		if e != nil {
			return nil, e
		}
		if _, err = g.AppendCO(l); err != nil {
			return nil, err
		}
	}
	g.HashStop()
	g.constraints = top.constraints
	if stopped() {
		return nil, ErrCancelled
	}
	result, _, err := g.Cleanup()
	if err != nil {
		return nil, err
	}
	return result, nil
}
