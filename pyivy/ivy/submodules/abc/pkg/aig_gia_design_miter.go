package abc

// General two-design miter from src/aig/gia/giaDup.c:Gia_ManMiter_rec/Gia_ManMiter.

type GiaMiterMode uint8

const (
	GiaMiterXor         GiaMiterMode = iota
	GiaMiterImplication              // first AND NOT second: a counterexample to first=>second.
	GiaMiterDual                     // adjacent uncombined first/second output values.
)

// GiaMiterOptions controls shared-input and state treatment. IndependentSuffix
// duplicates that many final PIs (sequential) or CIs (combinational) for side 1.
type GiaMiterOptions struct {
	GiaCofactorOptions
	Sequential        bool
	IndependentSuffix int
	Mode              GiaMiterMode
}

// GiaDesignMiter maps each source CI ordinal to a target CI ordinal. Sequential
// state banks are first then second, so reset lists concatenate in that order.
// Combinational mode reinterprets all source CIs/COs, including ROs/RIs, as data.
type GiaDesignMiter struct {
	Graph                           *GiaManager
	FirstInputs, SecondInputs       []int
	FirstRegisters, SecondRegisters int
}

// BuildGiaDesignMiter compares two graphs, sharing an input prefix. Sequential
// mode requires equal PI/PO counts and nonempty register banks on both sides;
// combinational mode requires equal CI/CO counts. XOR, implication violation
// and dual-output modes match C. All source POs participate, including source
// constraints; no assumption roles are inherited. Extended nodes are lowered.
// Result name is "miter" and Spec is empty as in C; sources stay unchanged.
// This constructs a query and makes no equivalence/safety assertion.
func BuildGiaDesignMiter(first, second *GiaManager, options GiaMiterOptions) (GiaDesignMiter, error) {
	fail := func(e error) (GiaDesignMiter, error) { return GiaDesignMiter{}, e }
	if first == nil || second == nil || options.IndependentSuffix < 0 || options.Mode > GiaMiterDual || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	for _, source := range []*GiaManager{first, second} {
		if e := source.Check(); e != nil {
			return fail(e)
		}
	}
	inputCount, outputCount := first.InputCount(), first.OutputCount()
	if options.Sequential {
		inputCount, outputCount = first.PrimaryInputCount(), first.PrimaryOutputCount()
		if first.registers == 0 || second.registers == 0 || inputCount != second.PrimaryInputCount() || outputCount != second.PrimaryOutputCount() {
			return fail(ErrDimension)
		}
	} else if inputCount != second.InputCount() || outputCount != second.OutputCount() {
		return fail(ErrDimension)
	}
	if options.IndependentSuffix > inputCount {
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
	g.Name = "miter"
	sources := [2]*GiaManager{first, second}
	maps := [2][]GiaLit{}
	indices := [2][]int{}
	for side, source := range sources {
		maps[side] = make([]GiaLit, len(source.objects))
		for i := range maps[side] {
			maps[side][i] = InvalidGiaLit
		}
		maps[side][0] = GiaConst0
		indices[side] = make([]int, source.InputCount())
	}
	for i := 0; i < inputCount; i++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		indices[0][i] = g.InputCount()
		maps[0][first.cis[i]], err = g.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	for i := 0; i < inputCount; i++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		if i < inputCount-options.IndependentSuffix {
			indices[1][i] = indices[0][i]
			maps[1][second.cis[i]] = maps[0][first.cis[i]]
		} else {
			indices[1][i] = g.InputCount()
			maps[1][second.cis[i]], err = g.AppendCI()
			if err != nil {
				return fail(err)
			}
		}
	}
	if options.Sequential {
		for side, source := range sources {
			for i := inputCount; i < source.InputCount(); i++ {
				if stopped() {
					return fail(ErrCancelled)
				}
				indices[side][i] = g.InputCount()
				maps[side][source.cis[i]], err = g.AppendCI()
				if err != nil {
					return fail(err)
				}
			}
		}
	}
	if err = g.HashStart(); err != nil {
		return fail(err)
	}
	copyRoot := func(side, index int) (GiaLit, error) {
		source := sources[side]
		co := source.cos[index]
		nodes, e := source.CollectTransitiveFanins([]GiaID{co}, options.MaxMappingEntries)
		if e != nil {
			return InvalidGiaLit, e
		}
		mapping := maps[side]
		for _, id := range nodes {
			if mapping[id].Valid() {
				continue
			}
			if stopped() {
				return InvalidGiaLit, ErrCancelled
			}
			mapping[id], e = hashGiaCopiedGate(source, g, id, mapping)
			if e != nil {
				return InvalidGiaLit, e
			}
		}
		driver, e := source.objects[co].Fanin0(co)
		if e != nil {
			return InvalidGiaLit, e
		}
		return mappedGiaSignal(mapping, driver)
	}
	for i := 0; i < outputCount; i++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		a, e := copyRoot(0, i)
		if e != nil {
			return fail(e)
		}
		b, e := copyRoot(1, i)
		if e != nil {
			return fail(e)
		}
		if options.Mode == GiaMiterDual {
			if _, err = g.AppendCO(a); err != nil {
				return fail(err)
			}
			if _, err = g.AppendCO(b); err != nil {
				return fail(err)
			}
		} else {
			var driver GiaLit
			if options.Mode == GiaMiterImplication {
				driver, err = g.HashAnd(a, b.Not())
			} else {
				driver, err = g.HashXor(a, b)
			}
			if err != nil {
				return fail(err)
			}
			if _, err = g.AppendCO(driver); err != nil {
				return fail(err)
			}
		}
	}
	firstRegisters, secondRegisters := 0, 0
	if options.Sequential {
		firstRegisters, secondRegisters = first.registers, second.registers
		for side, source := range sources {
			for i := source.PrimaryOutputCount(); i < source.OutputCount(); i++ {
				if stopped() {
					return fail(ErrCancelled)
				}
				driver, e := copyRoot(side, i)
				if e != nil {
					return fail(e)
				}
				if _, err = g.AppendCO(driver); err != nil {
					return fail(err)
				}
			}
		}
		g.registers = firstRegisters + secondRegisters
	}
	g.HashStop()
	if stopped() {
		return fail(ErrCancelled)
	}
	cleaned, _, err := g.Cleanup()
	if err != nil {
		return fail(err)
	}
	// C normalizes physical terminal placement after cleanup. Inputs were created
	// first here; retain physical gate order while collecting COs at the end.
	nodes, err := cleaned.AndObjects(options.MaxMappingEntries)
	if err != nil {
		return fail(err)
	}
	normalized, _, err := cleaned.DuplicateFromLists(cleaned.cis, nodes, cleaned.cos, cleaned.registers, options.GiaCofactorOptions)
	if err != nil {
		return fail(err)
	}
	return GiaDesignMiter{normalized, indices[0], indices[1], firstRegisters, secondRegisters}, nil
}
