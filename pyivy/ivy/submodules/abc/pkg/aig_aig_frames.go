package abc

// Port of src/aig/aig/aigFrames.c: Aig_ManFrames.

import "slices"

// AigFramesOptions controls sequential timeframe expansion. Initialize maps
// frame-zero latch outputs to constant zero; otherwise they are fresh CIs.
// Outputs emits primary outputs, either in every frame or only the last frame
// when Enlarge is set. Registers emits one final latch-input suffix; Enlarge
// selects its drivers from frame zero instead of the last frame.
//
// Zero MaxObjects and MaxMappingEntries select conservative finite defaults.
type AigFramesOptions struct {
	Frames, MaxObjects, MaxMappingEntries int
	Initialize, Outputs, Registers        bool
	Enlarge                               bool
	Stop                                  func() bool
}

// AigFrameExpansion owns an expanded AIG and its native-layout object map.
// Map entries are source-object-major: all frames of object zero precede all
// frames of object one. Entries removed by final cleanup, and combinations the
// native routine does not populate, are invalid.
type AigFrameExpansion struct {
	graph                 *AigManager
	frames, sourceObjects int
	mapping               []AigLit
}

func (e *AigFrameExpansion) Graph() *AigManager { return e.graph }
func (e *AigFrameExpansion) Frames() int        { return e.frames }

// Mapping returns an independent copy of the complete source-object-major map.
// A mapped CO is encoded as its regular object ID in an AigLit, matching the
// native routine's possibly-complemented object-pointer representation.
func (e *AigFrameExpansion) Mapping() []AigLit { return slices.Clone(e.mapping) }

// Object maps a source object and frame. It accepts every allocated source ID,
// including tombstones, but returns ErrInvalidLiteral for an unpopulated or
// cleanup-discarded entry.
func (e *AigFrameExpansion) Object(id AigID, frame int) (AigLit, error) {
	if e == nil || frame < 0 || frame >= e.frames || uint64(id) >= uint64(e.sourceObjects) {
		return InvalidAigLit, ErrDimension
	}
	literal := e.mapping[int(id)*e.frames+frame]
	if literal == InvalidAigLit {
		return literal, ErrInvalidLiteral
	}
	return literal, nil
}

// ExpandFrames implements Aig_ManFrames without borrowing source scratch data.
// The source is unchanged. Construction is atomic: errors return no partial
// expansion, and all returned graph and mapping storage is independently owned.
func (source *AigManager) ExpandFrames(options AigFramesOptions) (*AigFrameExpansion, error) {
	if source == nil || options.Frames < 1 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	entries, err := checkedSize(0, len(source.objects), options.Frames, options.MaxMappingEntries)
	if err != nil {
		return nil, err
	}
	// Aig_ManForEachNode excludes buffers. Reject them explicitly instead of
	// later consuming an absent native mapping through an unchecked pointer.
	for _, object := range source.objects {
		if object.Type == AigBuffer || object.Type == AigXOR {
			return nil, ErrUnsupported
		}
	}
	piCount := source.PrimaryInputCount()
	registerCount := source.RegisterCount()
	ciCount, err := checkedSize(0, options.Frames, piCount, options.MaxObjects)
	if err != nil {
		return nil, err
	}
	if !options.Initialize {
		ciCount, err = checkedSize(ciCount, 1, registerCount, options.MaxObjects)
		if err != nil {
			return nil, err
		}
	}
	if options.Registers && ciCount < registerCount {
		return nil, ErrDimension
	}
	if options.MaxObjects < 1 {
		return nil, ErrLimit
	}

	target := NewAigManager()
	target.Name, target.Spec = source.Name, source.Spec
	expansion := &AigFrameExpansion{
		graph: target, frames: options.Frames, sourceObjects: len(source.objects),
		mapping: make([]AigLit, entries),
	}
	for i := range expansion.mapping {
		expansion.mapping[i] = InvalidAigLit
	}
	index := func(id AigID, frame int) int { return int(id)*options.Frames + frame }
	set := func(id AigID, frame int, literal AigLit) { expansion.mapping[index(id, frame)] = literal }
	get := func(literal AigLit, frame int) (AigLit, error) {
		mapped := expansion.mapping[index(literal.ID(), frame)]
		if mapped == InvalidAigLit {
			return InvalidAigLit, ErrInvariant
		}
		return mapped.NotIf(literal.Complemented()), nil
	}
	checkStop := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		return nil
	}
	createCI := func() (AigLit, error) {
		if len(target.objects) >= options.MaxObjects {
			return InvalidAigLit, ErrLimit
		}
		return target.CreateCI()
	}
	createCO := func(driver AigLit) (AigID, error) {
		if len(target.objects) >= options.MaxObjects {
			return InvalidAigID, ErrLimit
		}
		return target.CreateCO(driver)
	}
	and := func(a, b AigLit) (AigLit, error) {
		literal, err := target.And(a, b)
		if err != nil {
			return InvalidAigLit, err
		}
		if len(target.objects) > options.MaxObjects {
			// The target is private and will be discarded, so detecting a one-node
			// overshoot here preserves atomic public behavior.
			return InvalidAigLit, ErrLimit
		}
		return literal, nil
	}

	for frame := 0; frame < options.Frames; frame++ {
		set(0, frame, AigConst1)
	}
	for frame := 0; frame < options.Frames; frame++ {
		for _, id := range source.cis[:piCount] {
			if err := checkStop(); err != nil {
				return nil, err
			}
			literal, err := createCI()
			if err != nil {
				return nil, err
			}
			set(id, frame, literal)
		}
	}
	for _, id := range source.cis[piCount:] {
		literal := AigConst0
		if !options.Initialize {
			literal, err = createCI()
			if err != nil {
				return nil, err
			}
		}
		set(id, 0, literal)
	}

	poCount := source.PrimaryOutputCount()
	for frame := 0; frame < options.Frames; frame++ {
		for id, object := range source.objects {
			if object.Type != AigAND {
				continue
			}
			if err := checkStop(); err != nil {
				return nil, err
			}
			left, err := get(object.Fanin0, frame)
			if err != nil {
				return nil, err
			}
			right, err := get(object.Fanin1, frame)
			if err != nil {
				return nil, err
			}
			literal, err := and(left, right)
			if err != nil {
				return nil, err
			}
			set(AigID(id), frame, literal)
		}
		for i, li := range source.cos[poCount:] {
			driver, err := get(source.objects[li].Fanin0, frame)
			if err != nil {
				return nil, err
			}
			if frame+1 < options.Frames {
				set(source.cis[piCount+i], frame+1, driver)
			}
		}
	}
	if options.Outputs {
		first := 0
		if options.Enlarge {
			first = options.Frames - 1
		}
		for frame := first; frame < options.Frames; frame++ {
			for _, id := range source.cos[:poCount] {
				if err := checkStop(); err != nil {
					return nil, err
				}
				driver, err := get(source.objects[id].Fanin0, frame)
				if err != nil {
					return nil, err
				}
				co, err := createCO(driver)
				if err != nil {
					return nil, err
				}
				set(id, frame, AigLit(co)<<1)
			}
		}
	}
	if options.Registers {
		driverFrame := options.Frames - 1
		if options.Enlarge {
			driverFrame = 0
		}
		for _, id := range source.cos[poCount:] {
			if err := checkStop(); err != nil {
				return nil, err
			}
			driver, err := get(source.objects[id].Fanin0, driverFrame)
			if err != nil {
				return nil, err
			}
			co, err := createCO(driver)
			if err != nil {
				return nil, err
			}
			set(id, options.Frames-1, AigLit(co)<<1)
		}
		if err := target.SetRegisterCount(registerCount); err != nil {
			return nil, err
		}
	}
	target.Cleanup()
	for i, literal := range expansion.mapping {
		if literal != InvalidAigLit && (uint64(literal.ID()) >= uint64(len(target.objects)) || target.objects[literal.ID()].Type == AigNone) {
			expansion.mapping[i] = InvalidAigLit
		}
	}
	if err := target.Check(); err != nil {
		return nil, err
	}
	return expansion, nil
}
