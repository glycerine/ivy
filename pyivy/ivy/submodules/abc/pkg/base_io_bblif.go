package abc

// ABC SOP-network conversion from src/base/io/ioWriteBblif.c.

type ABCBblifOptions struct {
	MaxObjects     int
	MaxObjectID    int
	MaxFanins      int
	MaxSOPBytes    int
	MaxBinaryBytes int
	MaxWork        int
	MaxBytes       int
	Stop           func() bool
}

func normalizeABCBblifOptions(options ABCBblifOptions) (ABCBblifOptions, error) {
	values := []*int{
		&options.MaxObjects,
		&options.MaxObjectID,
		&options.MaxFanins,
		&options.MaxSOPBytes,
		&options.MaxBinaryBytes,
		&options.MaxWork,
		&options.MaxBytes,
	}
	defaults := [...]int{1 << 20, 1 << 24, 1 << 20, 64 << 20, 256 << 20, 1 << 24, 256 << 20}
	for index, value := range values {
		if *value < 0 {
			return options, ErrDimension
		}
		if *value == 0 {
			*value = defaults[index]
		}
	}
	return options, nil
}

type abcBblifBudget struct {
	options ABCBblifOptions
	work    int
}

func (budget *abcBblifBudget) take(amount int) error {
	if amount < 0 || budget.work > budget.options.MaxWork-amount {
		return ErrLimit
	}
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	budget.work += amount
	return nil
}

func checkedABCBblifAdd(first, second int) (int, error) {
	if first < 0 || second < 0 || first > int(^uint(0)>>1)-second {
		return 0, ErrLimit
	}
	return first + second, nil
}

func checkedABCBblifMul(first, second int) (int, error) {
	if first < 0 || second < 0 || first != 0 && second > int(^uint(0)>>1)/first {
		return 0, ErrLimit
	}
	return first * second, nil
}

type abcBblifDFSFrame struct {
	id, next int
}

func abcBblifNodeOrder(network *ABCNetwork, budget *abcBblifBudget) ([]int, error) {
	state := make([]uint8, len(network.objects))
	order := make([]int, 0)
	for _, outputID := range network.cos {
		output := &network.objects[outputID]
		if len(output.Fanins) != 1 {
			return nil, ErrInvariant
		}
		start := output.Fanins[0].ID
		if network.objects[start].Type == ABCObjectPI || state[start] == 2 {
			continue
		}
		if network.objects[start].Type != ABCObjectNode {
			return nil, ErrUnsupported
		}
		stack := []abcBblifDFSFrame{{id: start}}
		for len(stack) != 0 {
			frame := &stack[len(stack)-1]
			object := &network.objects[frame.id]
			if state[frame.id] == 0 {
				if err := budget.take(1); err != nil {
					return nil, err
				}
				state[frame.id] = 1
			}
			if frame.next < len(object.Fanins) {
				child := object.Fanins[frame.next].ID
				frame.next++
				if err := budget.take(1); err != nil {
					return nil, err
				}
				if network.objects[child].Type == ABCObjectPI || state[child] == 2 {
					continue
				}
				if network.objects[child].Type != ABCObjectNode {
					return nil, ErrUnsupported
				}
				if state[child] == 1 {
					return nil, ErrInvariant
				}
				stack = append(stack, abcBblifDFSFrame{id: child})
				continue
			}
			state[frame.id] = 2
			order = append(order, frame.id)
			stack = stack[:len(stack)-1]
		}
	}
	return order, nil
}

func checkABCBblifNetwork(network *ABCNetwork, options ABCBblifOptions, budget *abcBblifBudget) error {
	if network == nil || network.Type != ABCNetworkLogic || network.Function != ABCFunctionSOP {
		return ErrUnsupported
	}
	if err := network.Validate(); err != nil {
		return err
	}
	if len(network.objects) == 0 || len(network.objects)-1 > options.MaxObjectID {
		return ErrLimit
	}
	for id := range network.objects {
		object := &network.objects[id]
		if !object.Alive {
			continue
		}
		if err := budget.take(1); err != nil {
			return err
		}
		switch object.Type {
		case ABCObjectPI:
			if len(object.Fanins) != 0 {
				return ErrInvariant
			}
		case ABCObjectPO:
			if len(object.Fanins) != 1 {
				return ErrInvariant
			}
		case ABCObjectNode:
			if len(object.Fanins) > options.MaxFanins {
				return ErrLimit
			}
		default:
			return ErrUnsupported
		}
		for _, edge := range object.Fanins {
			if err := budget.take(1); err != nil {
				return err
			}
			if edge.Complemented {
				return ErrUnsupported
			}
			faninType := network.objects[edge.ID].Type
			if faninType != ABCObjectPI && faninType != ABCObjectNode {
				return ErrUnsupported
			}
		}
	}
	return nil
}

func abcBblifMemoryEstimate(network *ABCNetwork, nodes []int) (int, error) {
	spanBytes, err := checkedABCBblifMul(len(network.objects), 17)
	if err != nil {
		return 0, err
	}
	terminalCount, err := checkedABCBblifAdd(len(network.cis), len(network.cos))
	if err != nil {
		return 0, err
	}
	terminalBytes, err := checkedABCBblifMul(terminalCount, 32)
	if err != nil {
		return 0, err
	}
	total, err := checkedABCBblifAdd(spanBytes, terminalBytes)
	if err != nil {
		return 0, err
	}
	total, err = checkedABCBblifAdd(total, len(network.Name))
	if err != nil {
		return 0, err
	}
	for _, id := range nodes {
		object := &network.objects[id]
		amount, addErr := checkedABCBblifMul(len(object.Fanins), 16)
		if addErr != nil {
			return 0, addErr
		}
		amount, addErr = checkedABCBblifAdd(amount, len(object.Function)+64)
		if addErr != nil {
			return 0, addErr
		}
		total, addErr = checkedABCBblifAdd(total, amount)
		if addErr != nil {
			return 0, addErr
		}
	}
	return total, nil
}

// BblManagerFromABCNetwork implements Bbl_ManFromAbc for a standard SOP
// logic network. Returned manager storage is independent of the source.
func BblManagerFromABCNetwork(network *ABCNetwork, options ABCBblifOptions) (*BblManager, error) {
	options, err := normalizeABCBblifOptions(options)
	if err != nil {
		return nil, err
	}
	budget := abcBblifBudget{options: options}
	if err = checkABCBblifNetwork(network, options, &budget); err != nil {
		return nil, err
	}
	nodes, err := abcBblifNodeOrder(network, &budget)
	if err != nil {
		return nil, err
	}
	objectCount, err := checkedABCBblifAdd(len(network.cis), len(nodes))
	if err != nil {
		return nil, err
	}
	objectCount, err = checkedABCBblifAdd(objectCount, len(network.cos))
	if err != nil || objectCount > options.MaxObjects {
		return nil, ErrLimit
	}
	memory, err := abcBblifMemoryEstimate(network, nodes)
	if err != nil || memory > options.MaxBytes {
		return nil, ErrLimit
	}
	manager, err := NewBblManager(network.Name, BblLimits{
		MaxObjects:     options.MaxObjects,
		MaxObjectID:    options.MaxObjectID,
		MaxFanins:      options.MaxFanins,
		MaxSOPBytes:    options.MaxSOPBytes,
		MaxBinaryBytes: options.MaxBinaryBytes,
		Stop:           options.Stop,
	})
	if err != nil {
		return nil, err
	}
	fail := func(cause error) (*BblManager, error) {
		_ = manager.Close()
		return nil, cause
	}
	for _, id := range network.cis {
		if err = budget.take(1); err != nil {
			return fail(err)
		}
		if err = manager.AddObject(BblObjectCI, id, 0, ""); err != nil {
			return fail(err)
		}
	}
	for _, id := range nodes {
		object := &network.objects[id]
		sop, parseErr := ParseABCSOP(object.Function, ABCMetadataOptions{MaxVariables: options.MaxFanins, MaxCubes: options.MaxWork, MaxBytes: options.MaxSOPBytes, MaxOperations: options.MaxWork, Stop: options.Stop})
		if parseErr != nil {
			return fail(parseErr)
		}
		if sop.Variables != len(object.Fanins) {
			return fail(ErrUnsupported)
		}
		if err = budget.take(len(object.Function) + 1); err != nil {
			return fail(err)
		}
		if err = manager.AddObject(BblObjectNode, id, len(object.Fanins), object.Function); err != nil {
			return fail(err)
		}
	}
	for _, id := range network.cos {
		if err = budget.take(1); err != nil {
			return fail(err)
		}
		if err = manager.AddObject(BblObjectCO, id, 1, ""); err != nil {
			return fail(err)
		}
	}
	for _, id := range nodes {
		for _, edge := range network.objects[id].Fanins {
			if err = budget.take(1); err != nil {
				return fail(err)
			}
			if err = manager.AddFanin(id, edge.ID); err != nil {
				return fail(err)
			}
		}
	}
	for _, id := range network.cos {
		if err = budget.take(1); err != nil {
			return fail(err)
		}
		if err = manager.AddFanin(id, network.objects[id].Fanins[0].ID); err != nil {
			return fail(err)
		}
	}
	if err = manager.Check(); err != nil {
		return fail(err)
	}
	return manager, nil
}

// MarshalABCBinaryBLIF is the owned-byte counterpart of Io_WriteBblif.
func MarshalABCBinaryBLIF(network *ABCNetwork, options ABCBblifOptions) ([]byte, error) {
	options, err := normalizeABCBblifOptions(options)
	if err != nil {
		return nil, err
	}
	manager, err := BblManagerFromABCNetwork(network, options)
	if err != nil {
		return nil, err
	}
	defer manager.Close()
	result, err := manager.MarshalBinary()
	if err != nil {
		return nil, err
	}
	stats, err := manager.Stats()
	if err != nil {
		return nil, err
	}
	peak, err := checkedABCBblifAdd(stats.NameBytes, stats.ObjectBytes)
	if err == nil {
		peak, err = checkedABCBblifAdd(peak, stats.FunctionBytes)
	}
	if err == nil {
		peak, err = checkedABCBblifAdd(peak, len(result))
	}
	if err != nil || peak > options.MaxBytes {
		return nil, ErrLimit
	}
	return result, nil
}
