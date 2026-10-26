package abc

// Minimal-AIG integration from src/base/abci/abcMini.c.

import (
	"strconv"
	"sync"
)

var ABCMiniAIGDefinitions = [...]string{
	"Abc_NodeFanin0Copy", "Abc_NodeFanin0Copy2", "Abc_NodeFanin1Copy",
	"Abc_NodeFanin1Copy2", "Abc_NtkFromMiniAig", "Abc_NtkInputMiniAig",
	"Abc_NtkMiniAigTest", "Abc_NtkOutputMiniAig", "Abc_NtkSetFlopNum",
	"Abc_NtkToMiniAig",
}

type ABCMiniAIGOptions struct {
	MaxObjects, MaxMappingEntries, MaxWork int
	Stop                                   func() bool
}

type abcMiniAIGBudget struct {
	options ABCMiniAIGOptions
	work    int
}

func normalizeABCMiniAIGOptions(options ABCMiniAIGOptions) (ABCMiniAIGOptions, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxWork < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 24
	}
	return options, nil
}

func newABCMiniAIGBudget(options ABCMiniAIGOptions) (*abcMiniAIGBudget, error) {
	options, err := normalizeABCMiniAIGOptions(options)
	if err != nil {
		return nil, err
	}
	return &abcMiniAIGBudget{options: options}, nil
}

func (b *abcMiniAIGBudget) tick(amount int) error {
	if amount < 0 || b.work > b.options.MaxWork-amount {
		return ErrLimit
	}
	b.work += amount
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func abcMiniToAigLiteral(mapping []AigLit, literal AIGLiteral) (AigLit, error) {
	if literal < 0 || int(literal)>>1 >= len(mapping) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	mapped := mapping[int(literal)>>1]
	if !mapped.Valid() {
		return InvalidAigLit, ErrInvariant
	}
	return mapped.NotIf(literal&1 != 0), nil
}

func abcAigToMiniLiteral(mapping []AIGLiteral, literal AigLit) (AIGLiteral, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) || mapping[literal.ID()] < 0 {
		return 0, ErrInvariant
	}
	return mapping[literal.ID()].NotIf(literal.Complemented()), nil
}

// ABCNetworkFromMiniAIG converts a checked MiniAIG into an owned strashed ABC
// network. The final register-count PIs/POs become latch outputs/inputs.
func ABCNetworkFromMiniAIG(source *MiniAIG, options ABCMiniAIGOptions) (*ABCNetwork, error) {
	budget, err := newABCMiniAIGBudget(options)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, ErrDimension
	}
	if err = source.Check(); err != nil {
		return nil, err
	}
	if source.NodeCount() > budget.options.MaxObjects || source.NodeCount() > budget.options.MaxMappingEntries {
		return nil, ErrLimit
	}
	pis, pos, _ := source.Counts()
	if source.RegisterCount() > pis || source.RegisterCount() > pos {
		return nil, ErrDimension
	}
	graph := NewAigManager()
	mapping := make([]AigLit, source.NodeCount())
	for index := range mapping {
		mapping[index] = InvalidAigLit
	}
	mapping[0] = AigConst0
	for id := 1; id < source.NodeCount(); id++ {
		if err = budget.tick(1); err != nil {
			return nil, err
		}
		nodeID := AIGNodeID(id)
		switch {
		case source.IsPI(nodeID):
			mapping[id], err = graph.CreateCI()
		case source.IsAND(nodeID):
			first, second, faninErr := source.Fanins(nodeID)
			if faninErr != nil {
				return nil, faninErr
			}
			var a, c AigLit
			a, err = abcMiniToAigLiteral(mapping, first)
			if err == nil {
				c, err = abcMiniToAigLiteral(mapping, second)
			}
			if err == nil {
				mapping[id], err = graph.And(a, c)
			}
		case source.IsPO(nodeID):
			first, _, faninErr := source.Fanins(nodeID)
			if faninErr != nil {
				return nil, faninErr
			}
			var driver AigLit
			driver, err = abcMiniToAigLiteral(mapping, first)
			if err == nil {
				var output AigID
				output, err = graph.CreateCO(driver)
				mapping[id] = AigLit(output) << 1
			}
		default:
			err = ErrInvariant
		}
		if err != nil {
			return nil, err
		}
		if graph.ObjectCount() > budget.options.MaxObjects {
			return nil, ErrLimit
		}
	}
	if err = graph.SetRegisterCount(source.RegisterCount()); err != nil {
		return nil, err
	}
	graph.Cleanup()
	if err = graph.Check(); err != nil {
		return nil, err
	}
	remaining := budget.options.MaxWork - budget.work
	if remaining == 0 {
		return nil, ErrLimit
	}
	network, err := AIGToABCNetwork(graph, ABCFunctionalOptions{MaxOperations: remaining, Stop: budget.options.Stop})
	if err != nil {
		return nil, err
	}
	if len(network.objects) > budget.options.MaxObjects {
		return nil, ErrLimit
	}
	network.Name = "MiniAig"
	for index, id := range network.pis {
		if err = network.SetObjectName(id, "pi"+strconv.Itoa(index)); err != nil {
			return nil, err
		}
	}
	for index, id := range network.pos {
		if err = network.SetObjectName(id, "po"+strconv.Itoa(index)); err != nil {
			return nil, err
		}
	}
	if err = network.Validate(); err != nil {
		return nil, err
	}
	return network, nil
}

// ABCNetworkToMiniAIG converts a frozen strashed network to an owned MiniAIG.
// Private mappings replace the source's destructive iTemp scratch.
func ABCNetworkToMiniAIG(network *ABCNetwork, options ABCMiniAIGOptions) (*MiniAIG, error) {
	budget, err := newABCMiniAIGBudget(options)
	if err != nil {
		return nil, err
	}
	if network == nil || network.Type != ABCNetworkStrash || network.Function != ABCFunctionAIG {
		return nil, ErrDimension
	}
	if err = CheckABCNetwork(network, ABCAnalysisOptions{MaxVisited: budget.options.MaxObjects, Stop: budget.options.Stop}); err != nil {
		return nil, err
	}
	if len(network.objects) > budget.options.MaxObjects || len(network.objects) > budget.options.MaxMappingEntries {
		return nil, ErrLimit
	}
	target, err := NewMiniAIG(MiniNetworkLimits{MaxNodes: budget.options.MaxObjects, Stop: budget.options.Stop})
	if err != nil {
		return nil, err
	}
	mapping := make([]AIGLiteral, len(network.objects))
	for index := range mapping {
		mapping[index] = -1
	}
	mapping[0] = 1 // Mini_AigLitConst1().
	for _, id := range network.cis {
		if err = budget.tick(1); err != nil {
			return nil, err
		}
		mapping[id], err = target.CreatePI()
		if err != nil {
			return nil, err
		}
	}
	order, err := ABCForwardDFS(network, nil, true, ABCAnalysisOptions{MaxVisited: budget.options.MaxObjects, Stop: budget.options.Stop})
	if err != nil {
		return nil, err
	}
	for _, id := range order {
		object := &network.objects[id]
		if object.Type != ABCObjectNode {
			continue
		}
		if err = budget.tick(1); err != nil {
			return nil, err
		}
		if len(object.Fanins) < 1 || len(object.Fanins) > 2 {
			return nil, ErrInvariant
		}
		read := func(edge ABCObjectEdge) (AIGLiteral, error) {
			if edge.ID < 0 || edge.ID >= len(mapping) || mapping[edge.ID] < 0 {
				return 0, ErrInvariant
			}
			return mapping[edge.ID].NotIf(edge.Complemented), nil
		}
		first, readErr := read(object.Fanins[0])
		if readErr != nil {
			return nil, readErr
		}
		if len(object.Fanins) == 1 {
			mapping[id] = first
		} else {
			second, secondErr := read(object.Fanins[1])
			if secondErr != nil {
				return nil, secondErr
			}
			mapping[id], err = target.AND(first, second)
		}
		if err != nil {
			return nil, err
		}
	}
	for _, id := range network.cos {
		if err = budget.tick(1); err != nil {
			return nil, err
		}
		object := &network.objects[id]
		if len(object.Fanins) != 1 {
			return nil, ErrInvariant
		}
		edge := object.Fanins[0]
		if edge.ID < 0 || edge.ID >= len(mapping) || mapping[edge.ID] < 0 {
			return nil, ErrInvariant
		}
		driver := mapping[edge.ID].NotIf(edge.Complemented)
		if _, err = target.CreatePO(driver); err != nil {
			return nil, err
		}
	}
	if err = target.SetRegisterCount(len(network.boxes)); err != nil {
		return nil, err
	}
	if err = target.Check(); err != nil {
		return nil, err
	}
	return target, nil
}

// ABCSetFlopNumber returns the source operation's sequential repartition as a
// new network: the last n PIs and POs become latch outputs and inputs.
func ABCSetFlopNumber(network *ABCNetwork, flops int, options ABCMiniAIGOptions) (*ABCNetwork, error) {
	if flops < 0 {
		return nil, ErrDimension
	}
	mini, err := ABCNetworkToMiniAIG(network, options)
	if err != nil {
		return nil, err
	}
	pis, pos, _ := mini.Counts()
	if flops > pis || flops > pos {
		return nil, ErrDimension
	}
	if err = mini.SetRegisterCount(flops); err != nil {
		return nil, err
	}
	return ABCNetworkFromMiniAIG(mini, options)
}

// ABCMiniAIGRoundTrip replaces the source's process-file smoke test with an
// in-memory binary round trip and returns both the checked bytes and network.
func ABCMiniAIGRoundTrip(network *ABCNetwork, options ABCMiniAIGOptions) (*ABCNetwork, []byte, error) {
	mini, err := ABCNetworkToMiniAIG(network, options)
	if err != nil {
		return nil, nil, err
	}
	data, err := mini.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	parsed, err := ParseMiniAIG(data, MiniNetworkLimits{MaxNodes: mini.limits.MaxNodes, Stop: options.Stop})
	if err != nil {
		return nil, nil, err
	}
	result, err := ABCNetworkFromMiniAIG(parsed, options)
	if err != nil {
		return nil, nil, err
	}
	return result, data, nil
}

// ABCMiniAIGFrame is the concrete owned replacement for abcMini.c's opaque
// Abc_Frame_t adapters.
type ABCMiniAIGFrame struct {
	mu      sync.RWMutex
	network *ABCNetwork
}

func NewABCMiniAIGFrame() *ABCMiniAIGFrame { return &ABCMiniAIGFrame{} }

func (frame *ABCMiniAIGFrame) Input(source *MiniAIG, options ABCMiniAIGOptions) error {
	if frame == nil {
		return ErrDimension
	}
	network, err := ABCNetworkFromMiniAIG(source, options)
	if err != nil {
		return err
	}
	frame.mu.Lock()
	frame.network = network
	frame.mu.Unlock()
	return nil
}

func (frame *ABCMiniAIGFrame) Output(options ABCMiniAIGOptions) (*MiniAIG, error) {
	if frame == nil {
		return nil, ErrDimension
	}
	frame.mu.RLock()
	if frame.network == nil {
		frame.mu.RUnlock()
		return nil, ErrInvariant
	}
	network := frame.network.Clone()
	frame.mu.RUnlock()
	return ABCNetworkToMiniAIG(network, options)
}

func (frame *ABCMiniAIGFrame) SetFlopNumber(flops int, options ABCMiniAIGOptions) error {
	if frame == nil {
		return ErrDimension
	}
	frame.mu.RLock()
	if frame.network == nil {
		frame.mu.RUnlock()
		return ErrInvariant
	}
	network := frame.network.Clone()
	frame.mu.RUnlock()
	result, err := ABCSetFlopNumber(network, flops, options)
	if err != nil {
		return err
	}
	frame.mu.Lock()
	frame.network = result
	frame.mu.Unlock()
	return nil
}

func (frame *ABCMiniAIGFrame) Network() (*ABCNetwork, error) {
	if frame == nil {
		return nil, ErrDimension
	}
	frame.mu.RLock()
	defer frame.mu.RUnlock()
	if frame.network == nil {
		return nil, ErrInvariant
	}
	return frame.network.Clone(), nil
}
