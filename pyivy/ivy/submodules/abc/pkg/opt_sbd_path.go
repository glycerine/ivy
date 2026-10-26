package abc

// Critical-path and delay tracing from src/opt/sbd/sbdPath.c.

import "slices"

type SBDPathOptions struct {
	MaxObjects, MaxVisits int
	Stop                  func() bool
}

type SBDDelayTrace struct {
	AIGNodes, LUTs, Levels             int
	PathNodes, PathEdges, AllPathEdges int
	Path                               []bool
}

type sbdPathBudget struct {
	options SBDPathOptions
	visits  int
}

func newSBDPathBudget(graph *GiaManager, mapping *GiaLUTMapping, options SBDPathOptions) (*sbdPathBudget, error) {
	if graph == nil || mapping == nil || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 8 << 20
	}
	if len(graph.objects) > options.MaxObjects {
		return nil, ErrLimit
	}
	if err := mapping.checkGraph(graph); err != nil {
		return nil, err
	}
	budget := &sbdPathBudget{options: options}
	if err := budget.charge(1); err != nil {
		return nil, err
	}
	return budget, nil
}

func (budget *sbdPathBudget) charge(count int) error {
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	next, err := checkedSize(budget.visits, 1, count, budget.options.MaxVisits)
	if err == nil {
		budget.visits = next
	}
	return err
}

func sbdMappingFanins(mapping *GiaLUTMapping, id GiaID) ([]int, error) {
	if uint64(id) >= uint64(mapping.objects) || mapping.packed[id] == 0 {
		return nil, ErrDimension
	}
	offset := mapping.packed[id]
	count := mapping.packed[offset]
	return mapping.packed[offset+1 : offset+1+count], nil
}

func sbdUnboxedLevels(graph *GiaManager, mapping *GiaLUTMapping, budget *sbdPathBudget) ([]int, int, error) {
	levels := make([]int, len(graph.objects))
	for _, root := range mapping.Roots() {
		fanins, err := sbdMappingFanins(mapping, root)
		if err != nil {
			return nil, 0, err
		}
		if err = budget.charge(len(fanins) + 1); err != nil {
			return nil, 0, err
		}
		for _, fanin := range fanins {
			levels[root] = max(levels[root], levels[fanin]+1)
		}
	}
	maximum := 0
	for _, output := range graph.cos {
		if err := budget.charge(1); err != nil {
			return nil, 0, err
		}
		driver, err := graph.objects[output].Fanin0(output)
		if err != nil {
			return nil, 0, err
		}
		maximum = max(maximum, levels[driver.ID()])
	}
	return levels, maximum, nil
}

func sbdEnsureBoxLevel(graph *GiaManager, mapping *GiaLUTMapping, root GiaID, levels []int, ready []bool, budget *sbdPathBudget) error {
	if ready[root] {
		return nil
	}
	type frame struct {
		id   GiaID
		next int
	}
	stack := []frame{{id: root}}
	active := make([]bool, len(graph.objects))
	active[root] = true
	for len(stack) > 0 {
		current := &stack[len(stack)-1]
		if ready[current.id] {
			active[current.id] = false
			stack = stack[:len(stack)-1]
			continue
		}
		object := graph.objects[current.id]
		if object.IsCI() || object.IsConstant0() || !object.IsAnd() {
			return ErrInvariant
		}
		fanins, err := sbdMappingFanins(mapping, current.id)
		if err != nil {
			return err
		}
		if current.next < len(fanins) {
			fanin := GiaID(fanins[current.next])
			current.next++
			if err = budget.charge(1); err != nil {
				return err
			}
			if ready[fanin] {
				continue
			}
			if active[fanin] || graph.objects[fanin].IsCI() || fanin == 0 {
				return ErrInvariant
			}
			active[fanin] = true
			stack = append(stack, frame{id: fanin})
			continue
		}
		level := 0
		for _, fanin := range fanins {
			if !ready[fanin] {
				return ErrInvariant
			}
			level = max(level, levels[fanin]+1)
		}
		levels[current.id], ready[current.id] = level, true
		active[current.id] = false
		stack = stack[:len(stack)-1]
	}
	return nil
}

func sbdBoxedLevels(graph *GiaManager, mapping *GiaLUTMapping, timing *TimManager, budget *sbdPathBudget) ([]int, int, error) {
	if timing == nil || graph.RegisterCount() != 0 || len(timing.cis) != len(graph.cis) || len(timing.cos) != len(graph.cos) {
		return nil, 0, ErrDimension
	}
	primaryInputs, primaryOutputs, err := timing.timCanonicalLayout()
	if err != nil {
		return nil, 0, err
	}
	levels := make([]int, len(graph.objects))
	ready := make([]bool, len(graph.objects))
	ready[0] = true
	for index := 0; index < primaryInputs; index++ {
		ready[graph.cis[index]] = true
	}
	for boxID, box := range timing.boxes {
		table, tableErr := timing.boxTable(boxID)
		if tableErr != nil {
			return nil, 0, tableErr
		}
		for input := 0; input < box.Inputs; input++ {
			outputID := graph.cos[box.FirstInput+input]
			driver, driverErr := graph.objects[outputID].Fanin0(outputID)
			if driverErr != nil {
				return nil, 0, driverErr
			}
			if driver.ID() != 0 {
				if driverErr = sbdEnsureBoxLevel(graph, mapping, driver.ID(), levels, ready, budget); driverErr != nil {
					return nil, 0, driverErr
				}
			}
		}
		for output := 0; output < box.Outputs; output++ {
			level := 0
			for input := 0; input < box.Inputs; input++ {
				if err = budget.charge(1); err != nil {
					return nil, 0, err
				}
				if table.Delays[output*box.Inputs+input] == TimNoArc {
					continue
				}
				outputID := graph.cos[box.FirstInput+input]
				driver, driverErr := graph.objects[outputID].Fanin0(outputID)
				if driverErr != nil {
					return nil, 0, driverErr
				}
				level = max(level, levels[driver.ID()]+1)
			}
			ci := graph.cis[box.FirstOutput+output]
			levels[ci], ready[ci] = level, true
		}
	}
	maximum := 0
	firstPrimaryOutput := len(graph.cos) - primaryOutputs
	for index := firstPrimaryOutput; index < len(graph.cos); index++ {
		outputID := graph.cos[index]
		driver, driverErr := graph.objects[outputID].Fanin0(outputID)
		if driverErr != nil {
			return nil, 0, driverErr
		}
		if driver.ID() != 0 {
			if driverErr = sbdEnsureBoxLevel(graph, mapping, driver.ID(), levels, ready, budget); driverErr != nil {
				return nil, 0, driverErr
			}
		}
		maximum = max(maximum, levels[driver.ID()])
	}
	return levels, maximum, nil
}

func sbdAddInternalToPath(graph *GiaManager, mapping *GiaLUTMapping, path []bool, budget *sbdPathBudget) ([]bool, error) {
	if len(path) != len(graph.objects) {
		return nil, ErrDimension
	}
	result := slices.Clone(path)
	for _, root := range mapping.Roots() {
		if !result[root] {
			continue
		}
		fanins, err := sbdMappingFanins(mapping, root)
		if err != nil {
			return nil, err
		}
		if err = budget.charge(len(fanins) + 1); err != nil {
			return nil, err
		}
		value := false
		for _, fanin := range fanins {
			value = value || result[fanin]
		}
		if value {
			result[root] = true
		}
	}
	return result, nil
}

// SBDAddInternalToPath is the owned form of Sbc_ManAddInternalToPath. The
// native traversal marks every LUT boundary before visiting its already-marked
// root, making the pass observably idempotent; this port preserves that detail.
func SBDAddInternalToPath(graph *GiaManager, mapping *GiaLUTMapping, path []bool, options SBDPathOptions) ([]bool, error) {
	budget, err := newSBDPathBudget(graph, mapping, options)
	if err != nil {
		return nil, err
	}
	return sbdAddInternalToPath(graph, mapping, path, budget)
}

// SBDCriticalPath returns the source's slack-one mapped critical-path marks.
// A nonnil timing manager enables canonical timing-box propagation.
func SBDCriticalPath(graph *GiaManager, mapping *GiaLUTMapping, timing *TimManager, options SBDPathOptions) ([]bool, error) {
	budget, err := newSBDPathBudget(graph, mapping, options)
	if err != nil {
		return nil, err
	}
	var levels []int
	var maximum int
	if timing == nil {
		levels, maximum, err = sbdUnboxedLevels(graph, mapping, budget)
	} else {
		levels, maximum, err = sbdBoxedLevels(graph, mapping, timing, budget)
	}
	if err != nil {
		return nil, err
	}
	type item struct {
		id, levelFan, slack int
	}
	path := make([]bool, len(graph.objects))
	seen := make([]bool, len(graph.objects))
	stack := make([]item, 0)
	roots := make([]item, 0, len(graph.cos))
	for outputIndex := 0; outputIndex < len(graph.cos); outputIndex++ {
		outputID := graph.cos[outputIndex]
		driver, driverErr := graph.objects[outputID].Fanin0(outputID)
		if driverErr != nil {
			return nil, driverErr
		}
		if driver.ID() != 0 && levels[driver.ID()] == maximum {
			roots = append(roots, item{id: int(driver.ID()), levelFan: levels[driver.ID()], slack: 1})
		}
	}
	for index := len(roots) - 1; index >= 0; index-- {
		stack = append(stack, roots[index])
	}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[current.id] {
			continue
		}
		if err = budget.charge(1); err != nil {
			return nil, err
		}
		seen[current.id], path[current.id] = true, true
		object := graph.objects[current.id]
		var fanins []int
		if object.IsCI() {
			if timing == nil {
				continue
			}
			boxID := timing.BoxForCI(int(object.Diff1()))
			if boxID < 0 {
				continue
			}
			box, boxErr := timing.Box(boxID)
			if boxErr != nil {
				return nil, boxErr
			}
			fanins = make([]int, 0, box.Inputs)
			for input := 0; input < box.Inputs; input++ {
				outputID := graph.cos[box.FirstInput+input]
				driver, driverErr := graph.objects[outputID].Fanin0(outputID)
				if driverErr != nil {
					return nil, driverErr
				}
				if driver.ID() != 0 {
					fanins = append(fanins, int(driver.ID()))
				}
			}
		} else if object.IsAnd() {
			fanins, err = sbdMappingFanins(mapping, GiaID(current.id))
			if err != nil {
				return nil, err
			}
		} else if current.id == 0 {
			continue
		} else {
			return nil, ErrInvariant
		}
		for index := len(fanins) - 1; index >= 0; index-- {
			fanin := fanins[index]
			threshold := int64(current.levelFan) - 1
			available := int64(levels[fanin]) + int64(current.slack)
			if available < threshold {
				continue
			}
			nextSlack := available - threshold
			if nextSlack < 0 {
				nextSlack = 0
			}
			if nextSlack > int64(^uint(0)>>1) {
				return nil, ErrLimit
			}
			stack = append(stack, item{id: fanin, levelFan: levels[fanin], slack: int(nextSlack)})
		}
	}
	return sbdAddInternalToPath(graph, mapping, path, budget)
}

// SBDTraceDelay returns Sbc_ManDelayTrace's statistics and final path marks
// instead of writing process output.
func SBDTraceDelay(graph *GiaManager, mapping *GiaLUTMapping, options SBDPathOptions) (SBDDelayTrace, error) {
	budget, err := newSBDPathBudget(graph, mapping, options)
	if err != nil {
		return SBDDelayTrace{}, err
	}
	levels, maximum, err := sbdUnboxedLevels(graph, mapping, budget)
	if err != nil {
		return SBDDelayTrace{}, err
	}
	result := SBDDelayTrace{AIGNodes: graph.AndCount(), LUTs: len(mapping.Roots()), Levels: maximum, Path: make([]bool, len(graph.objects))}
	for _, output := range graph.cos {
		driver, driverErr := graph.objects[output].Fanin0(output)
		if driverErr != nil {
			return SBDDelayTrace{}, driverErr
		}
		if levels[driver.ID()] == maximum {
			result.Path[driver.ID()] = true
		}
	}
	roots := mapping.Roots()
	for index := len(roots) - 1; index >= 0; index-- {
		root := roots[index]
		if !result.Path[root] {
			continue
		}
		result.PathNodes++
		fanins, faninErr := sbdMappingFanins(mapping, root)
		if faninErr != nil {
			return SBDDelayTrace{}, faninErr
		}
		for _, fanin := range fanins {
			if err = budget.charge(1); err != nil {
				return SBDDelayTrace{}, err
			}
			if levels[fanin]+1 < levels[root] {
				continue
			}
			if levels[fanin]+1 != levels[root] {
				return SBDDelayTrace{}, ErrInvariant
			}
			result.Path[fanin] = true
			result.PathEdges++
		}
	}
	for _, root := range roots {
		fanins, faninErr := sbdMappingFanins(mapping, root)
		if faninErr != nil {
			return SBDDelayTrace{}, faninErr
		}
		for _, fanin := range fanins {
			if err = budget.charge(1); err != nil {
				return SBDDelayTrace{}, err
			}
			if result.Path[root] && result.Path[fanin] {
				result.AllPathEdges++
			}
		}
	}
	return result, nil
}
