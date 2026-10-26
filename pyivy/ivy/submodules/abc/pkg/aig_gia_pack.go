package abc

// LUT ordering and delay-driven block packing from src/aig/gia/giaPack.c.

type GiaLUTPackOptions struct {
	MaxObjects int
	MaxVisits  int
	MaxEntries int
	Stop       func() bool
}

type GiaLUTPackResult struct {
	Packing        *GiaLUTPacking
	Order          []GiaID
	Delays, Blocks []int
	MaximumDelay   int
}

func checkGiaLUTPackOptions(options GiaLUTPackOptions) (GiaLUTPackOptions, error) {
	if options.MaxObjects < 0 || options.MaxVisits < 0 || options.MaxEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 20
	}
	return options, nil
}

func (mapping *GiaLUTMapping) checkPackGraph(manager *GiaManager, options GiaLUTPackOptions) (GiaLUTPackOptions, error) {
	options, err := checkGiaLUTPackOptions(options)
	if err != nil {
		return options, err
	}
	if manager == nil || mapping == nil {
		return options, ErrDimension
	}
	if len(manager.objects) > options.MaxObjects {
		return options, ErrLimit
	}
	if err = mapping.checkGraph(manager); err != nil {
		return options, err
	}
	if options.Stop != nil && options.Stop() {
		return options, ErrCancelled
	}
	return options, nil
}

// LUTReachableOrder preserves Gia_ManLutCollect2's output-rooted breadth-first
// discovery followed by reversal. Dangling LUTs are intentionally omitted.
func (mapping *GiaLUTMapping) LUTReachableOrder(manager *GiaManager, options GiaLUTPackOptions) ([]GiaID, error) {
	options, err := mapping.checkPackGraph(manager, options)
	if err != nil {
		return nil, err
	}
	seen := make([]bool, len(manager.objects))
	order := make([]GiaID, 0, len(mapping.Roots()))
	for _, output := range manager.cos {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		driver, driverErr := manager.objects[output].Fanin0(output)
		if driverErr != nil {
			return nil, driverErr
		}
		id := driver.ID()
		if !manager.objects[id].IsAnd() {
			continue
		}
		if _, found := mapping.Record(id); !found {
			return nil, ErrDimension
		}
		if !seen[id] {
			seen[id] = true
			order = append(order, id)
		}
	}
	visits := 0
	for index := 0; index < len(order); index++ {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		record, _ := mapping.Record(order[index])
		for _, fanin := range record.Fanins {
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			if visits == options.MaxVisits {
				return nil, ErrLimit
			}
			visits++
			if !manager.objects[fanin].IsAnd() {
				continue
			}
			if _, found := mapping.Record(fanin); !found {
				return nil, ErrDimension
			}
			if !seen[fanin] {
				seen[fanin] = true
				order = append(order, fanin)
			}
		}
	}
	for left, right := 0, len(order)-1; left < right; left, right = left+1, right-1 {
		order[left], order[right] = order[right], order[left]
	}
	return order, nil
}

func (mapping *GiaLUTMapping) lutDistanceOrder(manager *GiaManager, options GiaLUTPackOptions, visits *int) ([]GiaID, error) {
	roots := mapping.Roots()
	distance := make([]int, len(manager.objects))
	for _, output := range manager.cos {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		driver, err := manager.objects[output].Fanin0(output)
		if err != nil {
			return nil, err
		}
		distance[driver.ID()] = 1
	}
	for index := len(roots) - 1; index >= 0; index-- {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		record, _ := mapping.Record(roots[index])
		next, err := giaLUTDelayAdd(distance[record.Root], 1)
		if err != nil {
			return nil, err
		}
		for _, fanin := range record.Fanins {
			if *visits == options.MaxVisits {
				return nil, ErrLimit
			}
			(*visits)++
			distance[fanin] = max(distance[fanin], next)
		}
	}
	costs := make([]int, len(roots))
	for index, root := range roots {
		costs[index] = -distance[root]
	}
	permutation := IntCostPermutation(costs)
	order := make([]GiaID, len(roots))
	for index, position := range permutation {
		order[index] = roots[position]
	}
	return order, nil
}

// LUTDistanceOrder preserves Gia_ManLutCollect, including dangling LUTs and
// Abc_MergeSortCost's non-stable equal-cost ordering.
func (mapping *GiaLUTMapping) LUTDistanceOrder(manager *GiaManager, options GiaLUTPackOptions) ([]GiaID, error) {
	options, err := mapping.checkPackGraph(manager, options)
	if err != nil {
		return nil, err
	}
	visits := 0
	return mapping.lutDistanceOrder(manager, options, &visits)
}

func giaLUTDelayAdd(delay, increment int) (int, error) {
	if delay < 0 || increment < 0 || delay > int(^uint(0)>>1)-increment {
		return 0, ErrLimit
	}
	return delay + increment, nil
}

// Pack preserves Gia_ManLutPacking's critical-input bin choice, strict delay
// comparison, source selection-sort tie behavior, reverse group fill and
// route/direct delay recurrence. It returns owned scratch diagnostics instead
// of mutating the graph or printing the maximum delay.
func (mapping *GiaLUTMapping) Pack(manager *GiaManager, blockSize, routeDelay, directDelay int, options GiaLUTPackOptions) (GiaLUTPackResult, error) {
	fail := func(err error) (GiaLUTPackResult, error) { return GiaLUTPackResult{}, err }
	options, err := mapping.checkPackGraph(manager, options)
	if err != nil {
		return fail(err)
	}
	if blockSize < 2 || routeDelay <= 0 || directDelay <= 0 {
		return fail(ErrDimension)
	}
	visits := 0
	order, err := mapping.lutDistanceOrder(manager, options, &visits)
	if err != nil {
		return fail(err)
	}
	result := GiaLUTPackResult{Order: order, Delays: make([]int, len(manager.objects)), Blocks: make([]int, len(manager.objects))}
	blockSizes := make([]int, 0)
	for _, root := range order {
		if options.Stop != nil && options.Stop() {
			return fail(ErrCancelled)
		}
		record, _ := mapping.Record(root)
		if len(record.Fanins) < 2 || len(record.Fanins) > 32 {
			return fail(ErrDimension)
		}
		if len(record.Fanins) > options.MaxVisits-visits {
			return fail(ErrLimit)
		}
		visits += len(record.Fanins)
		fanins := append([]GiaID(nil), record.Fanins...)
		delays := make([]int, len(fanins))
		for index, fanin := range fanins {
			delays[index] = result.Delays[fanin]
		}
		for index := 0; index < len(fanins)-1; index++ {
			best := index
			for candidate := index + 1; candidate < len(fanins); candidate++ {
				if delays[candidate] > delays[best] {
					best = candidate
				}
			}
			fanins[index], fanins[best] = fanins[best], fanins[index]
			delays[index], delays[best] = delays[best], delays[index]
		}
		block := result.Blocks[fanins[0]]
		delay := 0
		if delays[0] > 0 && delays[0] > delays[1] && block >= 0 && block < len(blockSizes) && blockSizes[block] < blockSize {
			delay, err = giaLUTDelayAdd(delays[0], directDelay)
			if err != nil {
				return fail(err)
			}
			result.Blocks[root] = block
			blockSizes[block]++
		} else {
			delay, err = giaLUTDelayAdd(delays[0], routeDelay)
			if err != nil {
				return fail(err)
			}
			result.Blocks[root] = len(blockSizes)
			blockSizes = append(blockSizes, 1)
		}
		for index := 1; index < len(delays); index++ {
			candidate, addErr := giaLUTDelayAdd(delays[index], routeDelay)
			if addErr != nil {
				return fail(addErr)
			}
			delay = max(delay, candidate)
		}
		result.Delays[root] = delay
		result.MaximumDelay = max(result.MaximumDelay, delay)
	}
	groups := make([][]GiaID, len(blockSizes))
	for index, size := range blockSizes {
		groups[index] = make([]GiaID, 0, size)
	}
	for index := len(order) - 1; index >= 0; index-- {
		root := order[index]
		block := result.Blocks[root]
		groups[block] = append(groups[block], root)
	}
	result.Packing, err = NewGiaLUTPacking(mapping, groups, options.MaxEntries)
	if err != nil {
		return fail(err)
	}
	if options.Stop != nil && options.Stop() {
		return fail(ErrCancelled)
	}
	return result, nil
}
