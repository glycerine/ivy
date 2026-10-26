package abc

// Multi-implementation output collection from src/aig/gia/giaDup.c:
// Gia_ManChoiceMiter_rec and Gia_ManChoiceMiter.

// BuildGiaChoiceMiter shares every CI across implementations and hashes their
// reachable logic in output-major, source-minor DFS order. Output k of source
// i appears at k*len(sources)+i. Sources must have equal CI/CO/register counts.
// As in C, the result retains only the first source's register count and clears
// constraint roles: its final RegisterCount collected outputs become RIs.
// This is a choice-computation working graph, not an XOR comparison query or
// a composition of independent state machines. Name/spec come from source 0.
// Per-source owned maps also support repeated manager pointers without C's
// scratch-alias output loss. Extended nodes are lowered; sources stay unchanged.
func BuildGiaChoiceMiter(sources []*GiaManager, options GiaCofactorOptions) (*GiaManager, error) {
	if len(sources) == 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if len(sources) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	total := 0
	for i, source := range sources {
		if source == nil {
			return nil, ErrDimension
		}
		if e := source.Check(); e != nil {
			return nil, e
		}
		if i > 0 && (source.InputCount() != sources[0].InputCount() || source.OutputCount() != sources[0].OutputCount() || source.registers != sources[0].registers) {
			return nil, ErrDimension
		}
		var e error
		total, e = checkedSize(total, 1, len(source.objects), options.MaxMappingEntries)
		if e != nil {
			return nil, e
		}
	}
	first := sources[0]
	base, err := checkedSize(1, 1, first.InputCount(), options.MaxObjects)
	if err != nil {
		return nil, err
	}
	if _, err = checkedSize(base, first.OutputCount(), len(sources), options.MaxObjects); err != nil {
		return nil, err
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = first.Name, first.Spec
	mappings := make([][]GiaLit, len(sources))
	for i, source := range sources {
		mappings[i] = make([]GiaLit, len(source.objects))
		for j := range mappings[i] {
			mappings[i][j] = InvalidGiaLit
		}
		mappings[i][0] = GiaConst0
	}
	for index := 0; index < first.InputCount(); index++ {
		if stopped() {
			return nil, ErrCancelled
		}
		l, e := g.AppendCI()
		if e != nil {
			return nil, e
		}
		for i, source := range sources {
			mappings[i][source.cis[index]] = l
		}
	}
	if err = g.HashStart(); err != nil {
		return nil, err
	}
	for output := 0; output < first.OutputCount(); output++ {
		for i, source := range sources {
			if stopped() {
				return nil, ErrCancelled
			}
			co := source.cos[output]
			nodes, e := source.CollectTransitiveFanins([]GiaID{co}, options.MaxMappingEntries)
			if e != nil {
				return nil, e
			}
			mapping := mappings[i]
			for _, id := range nodes {
				if mapping[id].Valid() {
					continue
				}
				if stopped() {
					return nil, ErrCancelled
				}
				mapping[id], err = hashGiaCopiedGate(source, g, id, mapping)
				if err != nil {
					return nil, err
				}
			}
			driver, e := source.objects[co].Fanin0(co)
			if e != nil {
				return nil, e
			}
			l, e := mappedGiaSignal(mapping, driver)
			if e != nil {
				return nil, e
			}
			mapping[co], err = g.AppendCO(l)
			if err != nil {
				return nil, err
			}
		}
	}
	g.HashStop()
	g.registers = first.registers
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
