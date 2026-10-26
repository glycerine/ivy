package abc

// SOP-network reconstruction from src/base/io/ioReadBblif.c.

func checkBblManagerForABC(manager *BblManager, options ABCBblifOptions, budget *abcBblifBudget) (int, error) {
	if manager == nil {
		return 0, ErrDimension
	}
	if err := manager.Check(); err != nil {
		return 0, err
	}
	if len(manager.objects) > options.MaxObjects {
		return 0, ErrLimit
	}
	edges := 0
	memory := len(manager.name)
	for index := range manager.objects {
		object := &manager.objects[index]
		if err := budget.take(1); err != nil {
			return 0, err
		}
		if object.id < 0 || object.id > options.MaxObjectID || object.box || len(object.fanins) > options.MaxFanins {
			return 0, ErrUnsupported
		}
		var err error
		edges, err = checkedABCBblifAdd(edges, len(object.fanins))
		if err != nil {
			return 0, err
		}
		faninBytes, mulErr := checkedABCBblifMul(len(object.fanins), 16)
		if mulErr != nil {
			return 0, mulErr
		}
		objectBytes, addErr := checkedABCBblifAdd(128, faninBytes)
		if addErr != nil {
			return 0, addErr
		}
		memory, err = checkedABCBblifAdd(memory, objectBytes)
		if err != nil {
			return 0, err
		}
		if object.kind == BblObjectNode {
			if object.function < 0 || object.function >= len(manager.functions) {
				return 0, ErrInvariant
			}
			memory, err = checkedABCBblifAdd(memory, len(manager.functions[object.function].sop))
			if err != nil {
				return 0, err
			}
		}
	}
	if memory > options.MaxBytes {
		return 0, ErrLimit
	}
	return edges, nil
}

func bblManagerNodeIndexes(manager *BblManager, options ABCBblifOptions, budget *abcBblifBudget) ([]int, error) {
	state := make([]uint8, len(manager.objects))
	order := make([]int, 0)
	for start, object := range manager.objects {
		if object.kind != BblObjectNode || state[start] == 2 {
			continue
		}
		stack := []abcBblifDFSFrame{{id: start}}
		for len(stack) != 0 {
			frame := &stack[len(stack)-1]
			current := &manager.objects[frame.id]
			if state[frame.id] == 0 {
				if err := budget.take(1); err != nil {
					return nil, err
				}
				state[frame.id] = 1
			}
			if frame.next < len(current.fanins) {
				child := current.fanins[frame.next]
				frame.next++
				if err := budget.take(1); err != nil {
					return nil, err
				}
				if child < 0 || child >= len(manager.objects) {
					return nil, ErrInvariant
				}
				switch manager.objects[child].kind {
				case BblObjectCI:
					continue
				case BblObjectNode:
				default:
					return nil, ErrInvariant
				}
				if state[child] == 1 {
					return nil, ErrInvariant
				}
				if state[child] == 0 {
					stack = append(stack, abcBblifDFSFrame{id: child})
				}
				continue
			}
			state[frame.id] = 2
			order = append(order, frame.id)
			stack = stack[:len(stack)-1]
		}
	}
	return order, nil
}

// BblManagerNodeOrder implements Bbl_ManDfs without consuming the manager's
// mark bits. It returns current BBLIF object IDs in fanin-first order.
func BblManagerNodeOrder(manager *BblManager, options ABCBblifOptions) ([]int, error) {
	options, err := normalizeABCBblifOptions(options)
	if err != nil {
		return nil, err
	}
	budget := abcBblifBudget{options: options}
	if _, err = checkBblManagerForABC(manager, options, &budget); err != nil {
		return nil, err
	}
	indexes, err := bblManagerNodeIndexes(manager, options, &budget)
	if err != nil {
		return nil, err
	}
	result := make([]int, len(indexes))
	for index, objectIndex := range indexes {
		result[index] = manager.objects[objectIndex].id
	}
	return result, nil
}

// ABCNetworkFromBblManager implements Bbl_ManToAbc. The returned SOP network
// owns all names, functions and adjacency storage.
func ABCNetworkFromBblManager(manager *BblManager, options ABCBblifOptions) (*ABCNetwork, error) {
	options, err := normalizeABCBblifOptions(options)
	if err != nil {
		return nil, err
	}
	budget := abcBblifBudget{options: options}
	edges, err := checkBblManagerForABC(manager, options, &budget)
	if err != nil {
		return nil, err
	}
	if _, err = bblManagerNodeIndexes(manager, options, &budget); err != nil {
		return nil, err
	}
	networkObjects, err := checkedABCBblifAdd(len(manager.objects), 1)
	networkLimit, limitErr := checkedABCBblifAdd(options.MaxObjects, 1)
	if err != nil || limitErr != nil || networkObjects > networkLimit {
		return nil, ErrLimit
	}
	nameBytes, err := checkedABCBblifAdd(len(manager.name), 16*(len(manager.objects)+1))
	if err != nil || nameBytes > options.MaxBytes {
		return nil, ErrLimit
	}
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{
		MaxObjects:   networkObjects,
		MaxEdges:     max(1, edges),
		MaxNameBytes: options.MaxBytes,
		Stop:         options.Stop,
	})
	if err != nil {
		return nil, err
	}
	network.Name = manager.name
	mapping := make([]int, len(manager.objects))
	for index, object := range manager.objects {
		if err = budget.take(1); err != nil {
			return nil, err
		}
		switch object.kind {
		case BblObjectCI:
			mapping[index], err = network.CreatePI()
		case BblObjectCO:
			mapping[index], err = network.CreatePO()
		case BblObjectNode:
			mapping[index], err = network.CreateNode()
			if err == nil {
				err = network.SetNodeFunction(mapping[index], manager.functions[object.function].sop)
			}
		default:
			err = ErrUnsupported
		}
		if err != nil {
			return nil, err
		}
	}
	for index, object := range manager.objects {
		for _, fanin := range object.fanins {
			if err = budget.take(1); err != nil {
				return nil, err
			}
			if err = network.AddFanin(mapping[index], ABCObjectEdge{ID: mapping[fanin]}); err != nil {
				return nil, err
			}
		}
	}
	if err = ABCShortNames(network); err != nil {
		return nil, err
	}
	if err = CheckABCNetwork(network, ABCAnalysisOptions{MaxVisited: options.MaxWork, Stop: options.Stop}); err != nil {
		return nil, err
	}
	return network, nil
}

func buildBblFactoredAIG(manager *AigManager, graph *DecoGraph, fanins []AigLit, options ABCBblifOptions, budget *abcBblifBudget) (AigLit, error) {
	if err := graph.check(); err != nil {
		return InvalidAigLit, err
	}
	if graph.constant {
		return AigConst1.NotIf(graph.root.Complement), nil
	}
	if len(fanins) != graph.leaves {
		return InvalidAigLit, ErrDimension
	}
	internal := len(graph.nodes) - graph.leaves
	projected, err := checkedABCBblifAdd(manager.ObjectCount(), internal)
	if err != nil || projected > options.MaxObjects {
		return InvalidAigLit, ErrLimit
	}
	projectedBytes, err := checkedABCBblifMul(projected, 32)
	if err != nil || projectedBytes > options.MaxBytes {
		return InvalidAigLit, ErrLimit
	}
	values := make([]AigLit, len(graph.nodes))
	copy(values, fanins)
	for index := graph.leaves; index < len(graph.nodes); index++ {
		if err = budget.take(1); err != nil {
			return InvalidAigLit, err
		}
		node := graph.nodes[index]
		left := values[node.Edge0.Node].NotIf(node.Edge0.Complement)
		right := values[node.Edge1.Node].NotIf(node.Edge1.Complement)
		values[index], err = manager.And(left, right)
		if err != nil {
			return InvalidAigLit, err
		}
	}
	if graph.root.Node < 0 || graph.root.Node >= len(values) {
		return InvalidAigLit, ErrInvariant
	}
	return values[graph.root.Node].NotIf(graph.root.Complement), nil
}

// AigManagerFromBblManager implements Bbl_ManToAig using the native Dec_Factor
// port and source-shaped fanin-first construction.
func AigManagerFromBblManager(bbl *BblManager, options ABCBblifOptions) (*AigManager, error) {
	options, err := normalizeABCBblifOptions(options)
	if err != nil {
		return nil, err
	}
	budget := abcBblifBudget{options: options}
	if _, err = checkBblManagerForABC(bbl, options, &budget); err != nil {
		return nil, err
	}
	order, err := bblManagerNodeIndexes(bbl, options, &budget)
	if err != nil {
		return nil, err
	}
	manager := NewAigManager()
	manager.Name = bbl.name
	mapping := make([]AigLit, len(bbl.objects))
	for index := range mapping {
		mapping[index] = InvalidAigLit
	}
	for index, object := range bbl.objects {
		if object.kind != BblObjectCI {
			continue
		}
		if err = budget.take(1); err != nil {
			return nil, err
		}
		if manager.ObjectCount() >= options.MaxObjects {
			return nil, ErrLimit
		}
		mapping[index], err = manager.CreateCI()
		if err != nil {
			return nil, err
		}
	}
	factors := make([]*DecoGraph, len(bbl.functions))
	defer func() {
		for _, graph := range factors {
			if graph != nil {
				graph.Close()
			}
		}
	}()
	factorLimits := DecFactorLimits{MaxVariables: options.MaxFanins, MaxCubes: options.MaxWork, MaxCubeBytes: options.MaxSOPBytes, MaxNodes: options.MaxObjects, MaxDepth: options.MaxWork, Stop: options.Stop}
	for _, index := range order {
		object := &bbl.objects[index]
		graph := factors[object.function]
		if graph == nil {
			graph, err = FactorSOPText(bbl.functions[object.function].sop, factorLimits)
			if err != nil {
				return nil, err
			}
			factors[object.function] = graph
		}
		fanins := make([]AigLit, len(object.fanins))
		for faninIndex, fanin := range object.fanins {
			if mapping[fanin] == InvalidAigLit {
				return nil, ErrInvariant
			}
			fanins[faninIndex] = mapping[fanin]
		}
		mapping[index], err = buildBblFactoredAIG(manager, graph, fanins, options, &budget)
		if err != nil {
			return nil, err
		}
	}
	for _, object := range bbl.objects {
		if object.kind != BblObjectCO {
			continue
		}
		if len(object.fanins) != 1 || mapping[object.fanins[0]] == InvalidAigLit {
			return nil, ErrInvariant
		}
		if manager.ObjectCount() >= options.MaxObjects {
			return nil, ErrLimit
		}
		if _, err = manager.CreateCO(mapping[object.fanins[0]]); err != nil {
			return nil, err
		}
	}
	manager.Cleanup()
	if err = manager.Check(); err != nil {
		return nil, err
	}
	return manager, nil
}

// VerifyBblifABCNetworks implements the concrete result of Bbl_ManVerify.
func VerifyBblifABCNetworks(first, second *ABCNetwork, options ABCBblifOptions) (CECResult, error) {
	options, err := normalizeABCBblifOptions(options)
	if err != nil {
		return CECResult{}, err
	}
	functional := ABCFunctionalOptions{MaxOperations: options.MaxWork, MaxTruthVariables: options.MaxFanins, Stop: options.Stop}
	left, err := ABCNetworkToAIG(first, functional)
	if err != nil {
		return CECResult{}, err
	}
	right, err := ABCNetworkToAIG(second, functional)
	if err != nil {
		return CECResult{}, err
	}
	return CompareAIGCombinational(left.Graph, right.Graph, CECOptions{
		MaxPatterns: options.MaxWork, MaxOperations: options.MaxWork, MaxBytes: options.MaxBytes,
		MaxNodes: options.MaxObjects, MaxClauses: options.MaxWork, MaxPartitions: options.MaxObjects,
		MaxDepth: options.MaxWork, Stop: options.Stop,
	})
}

// ReadABCBinaryBLIF is the owned-byte counterpart of Io_ReadBblif.
func ReadABCBinaryBLIF(data []byte, check bool, options ABCBblifOptions) (*ABCNetwork, error) {
	options, err := normalizeABCBblifOptions(options)
	if err != nil {
		return nil, err
	}
	bbl, err := ParseBblBinary(data, BblLimits{
		MaxObjects: options.MaxObjects, MaxObjectID: options.MaxObjectID, MaxFanins: options.MaxFanins,
		MaxSOPBytes: options.MaxSOPBytes, MaxBinaryBytes: options.MaxBinaryBytes, Stop: options.Stop,
	})
	if err != nil {
		return nil, err
	}
	aig, err := AigManagerFromBblManager(bbl, options)
	if err != nil {
		return nil, err
	}
	network, err := AIGToABCNetwork(aig, ABCFunctionalOptions{MaxOperations: options.MaxWork, MaxTruthVariables: options.MaxFanins, Stop: options.Stop})
	if err != nil {
		return nil, err
	}
	if network.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	network.Name = bbl.name
	if err = ABCShortNames(network); err != nil {
		return nil, err
	}
	if check {
		if err = CheckABCNetwork(network, ABCAnalysisOptions{MaxVisited: options.MaxWork, Stop: options.Stop}); err != nil {
			return nil, err
		}
	}
	return network, nil
}

// RoundTripABCBinaryBLIF implements the semantic core of Bbl_ManTest without
// process files, printing or wall-clock state.
func RoundTripABCBinaryBLIF(network *ABCNetwork, options ABCBblifOptions) (*ABCNetwork, CECResult, error) {
	data, err := MarshalABCBinaryBLIF(network, options)
	if err != nil {
		return nil, CECResult{}, err
	}
	result, err := ReadABCBinaryBLIF(data, true, options)
	if err != nil {
		return nil, CECResult{}, err
	}
	verification, err := VerifyBblifABCNetworks(network, result, options)
	if err != nil {
		return nil, CECResult{}, err
	}
	return result, verification, nil
}
