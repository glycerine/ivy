package abc

// Cofactor mux composition from src/aig/gia/giaDup.c:Gia_ManDupMux.

// GiaMuxCofactors combines single-output combinational graphs with the same
// CI order, selecting positive when CI[index] is true and negative otherwise.
// All CIs are retained. Name/Spec come from positive as in C; source graphs
// are immutable. Constraint counts must agree and are preserved. Zero caps
// use GiaCofactorOptions defaults.
func GiaMuxCofactors(index int, positive, negative *GiaManager, options GiaCofactorOptions) (*GiaManager, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if err := positive.Check(); err != nil {
		return nil, err
	}
	if err := negative.Check(); err != nil {
		return nil, err
	}
	if positive.registers != 0 || negative.registers != 0 || positive.OutputCount() != 1 || negative.OutputCount() != 1 || positive.InputCount() != negative.InputCount() || index < 0 || index >= positive.InputCount() || positive.constraints != negative.constraints {
		return nil, ErrDimension
	}
	if _, err := checkedSize(positive.ObjectCount(), 1, negative.ObjectCount(), options.MaxMappingEntries); err != nil {
		return nil, err
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name, target.Spec = positive.Name, positive.Spec
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	inputs := make([]GiaLit, positive.InputCount())
	for i := range inputs {
		if stopped() {
			return nil, ErrCancelled
		}
		inputs[i], err = target.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	drivers := [2]GiaLit{}
	for phase, source := range []*GiaManager{negative, positive} {
		if stopped() {
			return nil, ErrCancelled
		}
		nodes, err := source.CollectConeAnds(source.cos, nil, options.MaxMappingEntries)
		if err != nil {
			return nil, err
		}
		mapping := make([]GiaLit, source.ObjectCount())
		for i := range mapping {
			mapping[i] = InvalidGiaLit
		}
		mapping[0] = GiaConst0
		for i, id := range source.cis {
			mapping[id] = inputs[i]
		}
		for _, id := range nodes {
			if stopped() {
				return nil, ErrCancelled
			}
			mapping[id], err = hashGiaCopiedGate(source, target, id, mapping)
			if err != nil {
				return nil, err
			}
		}
		driver, err := source.objects[source.cos[0]].Fanin0(source.cos[0])
		if err != nil {
			return nil, err
		}
		drivers[phase], err = mappedGiaSignal(mapping, driver)
		if err != nil {
			return nil, err
		}
	}
	result, err := target.HashMux(inputs[index], drivers[1], drivers[0])
	if err != nil {
		return nil, err
	}
	if _, err = target.AppendCO(result); err != nil {
		return nil, err
	}
	if err = target.SetConstraintCount(positive.constraints); err != nil {
		return nil, err
	}
	target.HashStop()
	if stopped() {
		return nil, ErrCancelled
	}
	if target.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	cleaned, _, err := target.Cleanup()
	if err != nil {
		return nil, err
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return cleaned, nil
}
