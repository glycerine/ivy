package abc

// Timeframe construction from src/sat/bmc/bmcBmc.c.

// BmcFramesOptions controls the combinational timeframe image. Frames is the
// requested number of frames. MaxReachableANDs is the source's cumulative
// cone-size threshold: zero disables it, and a positive threshold is checked
// only after every primary output of a frame has been emitted. MaxObjects and
// MaxMappingEntries are hard Go resource limits; zero selects finite defaults.
type BmcFramesOptions struct {
	Frames, MaxReachableANDs, MaxObjects, MaxMappingEntries int
	Stop                                                    func() bool
}

// BmcFrameExpansion owns a combinational graph whose inputs are frame-major
// primary inputs and whose outputs are frame-major primary outputs. Registers
// start at zero. FramesBuilt may be smaller than requested when the reachable
// AND threshold is met.
type BmcFrameExpansion struct {
	graph                      *AigManager
	framesBuilt, reachableANDs int
}

func (result *BmcFrameExpansion) Graph() *AigManager {
	if result == nil {
		return nil
	}
	return result.graph
}

func (result *BmcFrameExpansion) FramesBuilt() int {
	if result == nil {
		return 0
	}
	return result.framesBuilt
}

func (result *BmcFrameExpansion) ReachableANDs() int {
	if result == nil {
		return 0
	}
	return result.reachableANDs
}

// countNewBmcANDs is the iterative, owned-mark equivalent of
// Saig_ManFramesCount_rec. It counts AND nodes reachable from root that have
// not been counted for an earlier output or frame.
func countNewBmcANDs(graph *AigManager, root AigLit, counted []bool, stop func() bool) (int, error) {
	if graph == nil || !graph.validSignal(root) || len(counted) != len(graph.objects) {
		return 0, ErrInvalidLiteral
	}
	stack := []AigID{root.ID()}
	total := 0
	for len(stack) != 0 {
		if stop != nil && stop() {
			return 0, ErrCancelled
		}
		last := len(stack) - 1
		id := stack[last]
		stack = stack[:last]
		if counted[id] {
			continue
		}
		object := graph.objects[id]
		if object.Type != AigAND {
			continue
		}
		counted[id] = true
		total++
		stack = append(stack, object.Fanin0.ID(), object.Fanin1.ID())
	}
	return total, nil
}

// BuildBmcFrames implements Saig_ManFramesBmc and
// Saig_ManFramesBmcLimit without borrowing source scratch fields. It retains
// the C routine's completed-frame threshold semantics while adding checked
// dimensions, hard allocation limits, cancellation and atomic failure.
func (source *AigManager) BuildBmcFrames(options BmcFramesOptions) (*BmcFrameExpansion, error) {
	if source == nil || options.Frames < 1 || options.MaxReachableANDs < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	if source.RegisterCount() == 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if len(source.objects) > options.MaxMappingEntries || options.Frames > options.MaxMappingEntries || options.MaxObjects < 1 {
		return nil, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	for id, object := range source.objects {
		if object.Type == AigBuffer || object.Type == AigXOR {
			return nil, ErrUnsupported
		}
		if id < len(source.equivs) && source.equivs[id] != InvalidAigID {
			return nil, ErrUnsupported
		}
	}

	target := NewAigManager()
	mapping := make([]AigLit, len(source.objects))
	for index := range mapping {
		mapping[index] = InvalidAigLit
	}
	mapping[0] = AigConst1
	primaryInputs := source.PrimaryInputCount()
	primaryOutputs := source.PrimaryOutputCount()
	for _, id := range source.cis[primaryInputs:] {
		mapping[id] = AigConst0
	}
	counted := []bool{false}
	result := &BmcFrameExpansion{graph: target}
	stopped := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		return nil
	}
	room := func() error {
		if len(target.objects) >= options.MaxObjects {
			return ErrLimit
		}
		return nil
	}
	lookup := func(literal AigLit) (AigLit, error) {
		if !source.validSignal(literal) {
			return InvalidAigLit, ErrInvalidLiteral
		}
		mapped := mapping[literal.ID()]
		if mapped == InvalidAigLit {
			return InvalidAigLit, ErrInvariant
		}
		return mapped.NotIf(literal.Complemented()), nil
	}

	for frame := 0; frame < options.Frames; frame++ {
		for _, id := range source.cis[:primaryInputs] {
			if err := stopped(); err != nil {
				return nil, err
			}
			if err := room(); err != nil {
				return nil, err
			}
			literal, err := target.CreateCI()
			if err != nil {
				return nil, err
			}
			counted = append(counted, false)
			mapping[id] = literal
		}
		for id, object := range source.objects {
			if object.Type != AigAND {
				continue
			}
			if err := stopped(); err != nil {
				return nil, err
			}
			left, err := lookup(object.Fanin0)
			if err != nil {
				return nil, err
			}
			right, err := lookup(object.Fanin1)
			if err != nil {
				return nil, err
			}
			before := len(target.objects)
			literal, err := target.And(left, right)
			if err != nil {
				return nil, err
			}
			if len(target.objects) > options.MaxObjects {
				return nil, ErrLimit
			}
			if len(target.objects) != before {
				counted = append(counted, false)
			}
			mapping[id] = literal
		}
		for _, id := range source.cos[:primaryOutputs] {
			if err := stopped(); err != nil {
				return nil, err
			}
			driver, err := lookup(source.objects[id].Fanin0)
			if err != nil {
				return nil, err
			}
			if err = room(); err != nil {
				return nil, err
			}
			if _, err = target.CreateCO(driver); err != nil {
				return nil, err
			}
			counted = append(counted, false)
			added, err := countNewBmcANDs(target, driver, counted, options.Stop)
			if err != nil {
				return nil, err
			}
			result.reachableANDs += added
		}
		result.framesBuilt++
		if options.MaxReachableANDs > 0 && result.reachableANDs >= options.MaxReachableANDs {
			break
		}
		if frame+1 == options.Frames {
			break
		}
		for index, id := range source.cos[primaryOutputs:] {
			if err := stopped(); err != nil {
				return nil, err
			}
			driver, err := lookup(source.objects[id].Fanin0)
			if err != nil {
				return nil, err
			}
			mapping[source.cis[primaryInputs+index]] = driver
		}
	}
	target.Cleanup()
	if err := target.Check(); err != nil {
		return nil, err
	}
	return result, nil
}
