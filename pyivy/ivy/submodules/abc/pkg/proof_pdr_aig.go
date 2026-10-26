package abc

// Port of src/proof/pdr/pdrInv.c:Pdr_ManDupAigWithClauses.

type PdrAigOptions struct {
	MaxObjects, MaxMappingEntries int
	Stop                          func() bool
}

// ToAig replaces the source's ordinary outputs and constraints with one
// excluded-state cube output per invariant clause, retaining all PIs, registers
// and next-state functions. Outputs are violations, not invariant clauses.
// Initialization belongs to AigerModel and is not stored in an AigManager.
// This operation does not verify the invariant and never mutates source.
func (invariant *PdrInvariant) ToAig(source *AigManager, options PdrAigOptions) (*AigManager, error) {
	if invariant == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if source.ObjectCount() > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	for _, cube := range invariant.clauses {
		if err := cube.Check(); err != nil {
			return nil, err
		}
		for _, l := range cube.state {
			if uint64(l.Index()) >= uint64(source.RegisterCount()) {
				return nil, ErrDimension
			}
		}
	}
	g := NewAigManager()
	g.Name = source.Name
	g.Spec = source.Spec
	check := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if g.ObjectCount() > options.MaxObjects {
			return ErrLimit
		}
		return nil
	}
	if err := check(); err != nil {
		return nil, err
	}
	mapping := make([]AigLit, source.ObjectCount())
	for i := range mapping {
		mapping[i] = InvalidAigLit
	}
	mapping[0] = AigConst1
	for _, id := range source.cis {
		if err := check(); err != nil {
			return nil, err
		}
		l, err := g.CreateCI()
		if err != nil {
			return nil, err
		}
		mapping[id] = l
	}
	for _, cube := range invariant.clauses {
		value := AigConst1
		for _, l := range cube.state {
			if err := check(); err != nil {
				return nil, err
			}
			signal := mapping[source.cis[source.PrimaryInputCount()+int(l.Index())]].NotIf(l.Negative())
			var err error
			value, err = g.And(value, signal)
			if err != nil {
				return nil, err
			}
		}
		if err := check(); err != nil {
			return nil, err
		}
		if _, err := g.CreateCO(value); err != nil {
			return nil, err
		}
	}
	translate := func(l AigLit) (AigLit, error) {
		if !l.Valid() || uint64(l.ID()) >= uint64(len(mapping)) || !mapping[l.ID()].Valid() {
			return InvalidAigLit, ErrInvariant
		}
		return mapping[l.ID()].NotIf(l.Complemented()), nil
	}
	for id, obj := range source.objects {
		if err := check(); err != nil {
			return nil, err
		}
		switch obj.Type {
		case AigAND:
			a, err := translate(obj.Fanin0)
			if err != nil {
				return nil, err
			}
			b, err := translate(obj.Fanin1)
			if err != nil {
				return nil, err
			}
			value, err := g.And(a, b)
			if err != nil {
				return nil, err
			}
			mapping[id] = value
		case AigBuffer:
			value, err := translate(obj.Fanin0)
			if err != nil {
				return nil, err
			}
			// Buffer identity is irrelevant to cube outputs or next-state semantics.
			mapping[id] = value
		}
	}
	for _, id := range source.cos[source.PrimaryOutputCount():] {
		if err := check(); err != nil {
			return nil, err
		}
		value, err := translate(source.objects[id].Fanin0)
		if err != nil {
			return nil, err
		}
		if _, err := g.CreateCO(value); err != nil {
			return nil, err
		}
	}
	if err := check(); err != nil {
		return nil, err
	}
	if err := g.SetRegisterCount(source.RegisterCount()); err != nil {
		return nil, err
	}
	g.Cleanup()
	if err := g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
