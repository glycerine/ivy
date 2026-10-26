package abc

// Sequential X-valued simulation from src/base/abci/abcXsim.c.

// ABCXSimDefinitions is the complete named-definition inventory of abcXsim.c.
var ABCXSimDefinitions = [...]string{
	"Abc_NtkCycleInitState", "Abc_NtkXValueSimulate", "Abc_ObjGetXsim",
	"Abc_ObjGetXsimFanin0", "Abc_ObjGetXsimFanin1", "Abc_ObjSetXsim",
	"Abc_XsimAnd", "Abc_XsimInv", "Abc_XsimPrint", "Abc_XsimRand2",
	"Abc_XsimRand3",
}

type ABCXSimOptions struct {
	Frames                      int
	UnknownInputs, UnknownState bool
	MaxFrames, MaxObjects       int
	MaxTraceEntries, MaxWork    int
	Stop                        func() bool
}

type ABCXSimFrame struct {
	Frame                    int
	Inputs, Latches, Outputs []AigTernary
}

type ABCXSimTrace struct {
	Frames []ABCXSimFrame
}

type ABCCycleInitStateOptions struct {
	Frames, MaxFrames, MaxObjects, MaxWork int
	UseUnknownInputs                       bool
	Stop                                   func() bool
}

type abcXSimBudget struct {
	maxWork int
	work    int
	stop    func() bool
}

func (b *abcXSimBudget) tick(amount int) error {
	if amount < 0 || b.work > b.maxWork-amount {
		return ErrLimit
	}
	b.work += amount
	if b.stop != nil && b.stop() {
		return ErrCancelled
	}
	return nil
}

type abcXSimRandom struct {
	z, w uint32
}

func newABCXSimRandom() abcXSimRandom {
	random := abcXSimRandom{z: 3716960521, w: 2174103536}
	random.next() // Gia_ManRandom(1) resets and advances once.
	return random
}

func (r *abcXSimRandom) next() uint32 {
	r.z = 36969*(r.z&65535) + (r.z >> 16)
	r.w = 18000*(r.w&65535) + (r.w >> 16)
	return (r.z << 16) + r.w
}

func abcXSimRandom2(random *abcXSimRandom) AigTernary {
	if random.next()&1 != 0 {
		return AigTernaryOne
	}
	return AigTernaryZero
}

func abcXSimRandom3(random *abcXSimRandom) AigTernary {
	for {
		value := AigTernary(random.next() & 3)
		if value != 0 {
			return value
		}
	}
}

func abcXSimValid(value AigTernary) bool {
	return value >= AigTernaryZero && value <= AigTernaryUnknown
}

func abcXSimText(values []AigTernary) (string, error) {
	text := make([]byte, len(values))
	for index, value := range values {
		switch value {
		case AigTernaryZero:
			text[index] = '0'
		case AigTernaryOne:
			text[index] = '1'
		case AigTernaryUnknown:
			text[index] = 'x'
		default:
			return "", ErrDimension
		}
	}
	return string(text), nil
}

// ABCXSimText renders the source's 0/1/x alphabet without writing to stdout.
func ABCXSimText(values []AigTernary) (string, error) { return abcXSimText(values) }

func normalizeABCXSimOptions(options ABCXSimOptions) (ABCXSimOptions, error) {
	if options.Frames < 0 || options.MaxFrames < 0 || options.MaxObjects < 0 || options.MaxTraceEntries < 0 || options.MaxWork < 0 {
		return options, ErrDimension
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 1 << 20
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxTraceEntries == 0 {
		options.MaxTraceEntries = 1 << 24
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 27
	}
	if options.Frames > options.MaxFrames {
		return options, ErrLimit
	}
	return options, nil
}

func abcXSimRead(values []AigTernary, edge ABCObjectEdge) (AigTernary, error) {
	if edge.ID < 0 || edge.ID >= len(values) || !abcXSimValid(values[edge.ID]) {
		return 0, ErrInvariant
	}
	value := values[edge.ID]
	if edge.Complemented {
		value = value.Not()
	}
	return value, nil
}

func abcXSimulate(network *ABCNetwork, options ABCXSimOptions, retainTrace bool) (ABCXSimTrace, []AigTernary, error) {
	var empty ABCXSimTrace
	options, err := normalizeABCXSimOptions(options)
	if err != nil {
		return empty, nil, err
	}
	if network == nil || network.Type != ABCNetworkStrash || network.Function != ABCFunctionAIG {
		return empty, nil, ErrDimension
	}
	if len(network.objects) > options.MaxObjects {
		return empty, nil, ErrLimit
	}
	if err = CheckABCNetwork(network, ABCAnalysisOptions{MaxVisited: options.MaxObjects, Stop: options.Stop}); err != nil {
		return empty, nil, err
	}
	for _, box := range network.boxes {
		if network.objects[box].Type != ABCObjectLatch {
			return empty, nil, ErrUnsupported
		}
	}
	latchInputs, latchOutputs, err := abcLatchTerminals(network)
	if err != nil {
		return empty, nil, err
	}
	width := len(network.pis) + len(network.boxes) + len(network.pos)
	if retainTrace && options.Frames > 0 && width > options.MaxTraceEntries/options.Frames {
		return empty, nil, ErrLimit
	}
	order, err := ABCForwardDFS(network, nil, true, ABCAnalysisOptions{MaxVisited: options.MaxObjects, Stop: options.Stop})
	if err != nil {
		return empty, nil, err
	}
	nodes := order[:0]
	for _, id := range order {
		if network.objects[id].Type == ABCObjectNode {
			nodes = append(nodes, id)
		}
	}
	budget := abcXSimBudget{maxWork: options.MaxWork, stop: options.Stop}
	if err = budget.tick(len(network.objects)); err != nil {
		return empty, nil, err
	}
	values := make([]AigTernary, len(network.objects))
	values[0] = AigTernaryOne
	state := make([]AigTernary, len(network.boxes))
	for index, latchID := range network.boxes {
		value := AigTernary(network.objects[latchID].Init)
		if options.UnknownState {
			value = AigTernaryUnknown
		}
		if !abcXSimValid(value) {
			return empty, nil, ErrInvariant
		}
		state[index] = value
		values[latchOutputs[index]] = value
	}
	random := newABCXSimRandom()
	assignInputs := func() error {
		for _, id := range network.pis {
			if err := budget.tick(1); err != nil {
				return err
			}
			if options.UnknownInputs {
				values[id] = AigTernaryUnknown
			} else {
				values[id] = abcXSimRandom2(&random)
			}
		}
		return nil
	}
	if err = assignInputs(); err != nil {
		return empty, nil, err
	}
	trace := ABCXSimTrace{}
	if retainTrace {
		trace.Frames = make([]ABCXSimFrame, 0, options.Frames)
	}
	for frame := 0; frame < options.Frames; frame++ {
		for _, id := range nodes {
			if err = budget.tick(1); err != nil {
				return empty, nil, err
			}
			object := &network.objects[id]
			switch len(object.Fanins) {
			case 1:
				values[id], err = abcXSimRead(values, object.Fanins[0])
			case 2:
				var first, second AigTernary
				first, err = abcXSimRead(values, object.Fanins[0])
				if err == nil {
					second, err = abcXSimRead(values, object.Fanins[1])
				}
				if err == nil {
					values[id] = aigTernaryAnd(first, second)
				}
			default:
				err = ErrInvariant
			}
			if err != nil {
				return empty, nil, err
			}
		}
		for _, id := range network.cos {
			if err = budget.tick(1); err != nil {
				return empty, nil, err
			}
			object := &network.objects[id]
			if len(object.Fanins) != 1 {
				return empty, nil, ErrInvariant
			}
			values[id], err = abcXSimRead(values, object.Fanins[0])
			if err != nil {
				return empty, nil, err
			}
		}
		if retainTrace {
			row := ABCXSimFrame{Frame: frame, Inputs: make([]AigTernary, len(network.pis)), Latches: append([]AigTernary(nil), state...), Outputs: make([]AigTernary, len(network.pos))}
			for index, id := range network.pis {
				row.Inputs[index] = values[id]
			}
			for index, id := range network.pos {
				row.Outputs[index] = values[id]
			}
			trace.Frames = append(trace.Frames, row)
		}
		for index, id := range latchInputs {
			state[index] = values[id]
			values[latchOutputs[index]] = state[index]
		}
		if frame+1 < options.Frames {
			if err = assignInputs(); err != nil {
				return empty, nil, err
			}
		}
	}
	if err = budget.tick(0); err != nil {
		return empty, nil, err
	}
	return trace, state, nil
}

// ABCXValueSimulate performs deterministic three-valued sequential simulation.
// It returns owned values instead of mutating pCopy scratch or printing rows.
func ABCXValueSimulate(network *ABCNetwork, options ABCXSimOptions) (ABCXSimTrace, error) {
	trace, _, err := abcXSimulate(network, options, true)
	return trace, err
}

// ABCCycleInitState returns an owned network whose latch initial values are the
// state reached after the requested prefix. The source network remains frozen.
func ABCCycleInitState(network *ABCNetwork, options ABCCycleInitStateOptions) (*ABCNetwork, error) {
	traceOptions := ABCXSimOptions{
		Frames: options.Frames, UnknownInputs: options.UseUnknownInputs,
		MaxFrames: options.MaxFrames, MaxObjects: options.MaxObjects,
		MaxWork: options.MaxWork, Stop: options.Stop,
	}
	_, state, err := abcXSimulate(network, traceOptions, false)
	if err != nil {
		return nil, err
	}
	result := network.Clone()
	for index, latchID := range result.boxes {
		result.objects[latchID].Init = ABCLatchInit(state[index])
	}
	if err = result.Validate(); err != nil {
		return nil, err
	}
	return result, nil
}
