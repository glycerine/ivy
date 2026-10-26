package abc

// Backend-independent local-BDD reordering for src/base/abci/abcReorder.c.

import "slices"

type ABCBDDLocalNode struct {
	Name     string
	Fanins   []int
	Function *TruthTable
}

type ABCBDDReorderOptions struct {
	Method                             CuddReorderMethod
	MaxNodes, MaxFanins, MaxTruthWords int
	MaxEvaluations, MaxPasses          int
	Stop                               func() bool
}

type ABCBDDReorderStats struct {
	FaninsBefore, FaninsAfter int
	NodesBefore, NodesAfter   int
	Evaluations, Passes       int
}

type ABCBDDReorderedNode struct {
	Name     string
	Fanins   []int
	Function *TruthTable
	Order    []int
	Stats    ABCBDDReorderStats
}

type ABCBDDNetworkReorder struct {
	Nodes                    []ABCBDDReorderedNode
	NodesBefore, NodesAfter  int
	RemovedFanins, Reordered int
}

func normalizeABCBDDReorderOptions(options ABCBDDReorderOptions) (ABCBDDReorderOptions, error) {
	if options.Method > CuddReorderSymmetry || options.MaxNodes < 0 || options.MaxFanins < 0 || options.MaxTruthWords < 0 || options.MaxEvaluations < 0 || options.MaxPasses < 0 {
		return options, ErrDimension
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 20
	}
	if options.MaxFanins == 0 {
		options.MaxFanins = 36
	}
	if options.MaxTruthWords == 0 {
		options.MaxTruthWords = 1 << 16
	}
	if options.MaxEvaluations == 0 {
		options.MaxEvaluations = 1 << 20
	}
	return options, nil
}

func abcBDDNormalizeNode(node ABCBDDLocalNode, options ABCBDDReorderOptions) (ABCBDDLocalNode, error) {
	if node.Function == nil || len(node.Fanins) > options.MaxFanins {
		if len(node.Fanins) > options.MaxFanins {
			return ABCBDDLocalNode{}, ErrLimit
		}
		return ABCBDDLocalNode{}, ErrDimension
	}
	if err := node.Function.Check(); err != nil || node.Function.Variables() != len(node.Fanins) {
		return ABCBDDLocalNode{}, ErrDimension
	}
	if len(node.Function.words) > options.MaxTruthWords {
		return ABCBDDLocalNode{}, ErrLimit
	}
	initial, err := node.Function.CompactSupport()
	if err != nil {
		return ABCBDDLocalNode{}, err
	}
	workingFanins := make([]int, len(initial.SourceVariables))
	for index, source := range initial.SourceVariables {
		workingFanins[index] = node.Fanins[source]
	}
	unique := make([]int, 0, len(workingFanins))
	positions := make([]int, len(workingFanins))
	seen := make(map[int]int, len(workingFanins))
	for index, fanin := range workingFanins {
		if fanin < 0 {
			return ABCBDDLocalNode{}, ErrDimension
		}
		position, found := seen[fanin]
		if !found {
			position = len(unique)
			seen[fanin] = position
			unique = append(unique, fanin)
		}
		positions[index] = position
	}
	if len(unique) > 20 {
		return ABCBDDLocalNode{}, ErrLimit
	}
	collapsed, err := NewConstantTruthTable(len(unique), false, options.MaxTruthWords)
	if err != nil {
		return ABCBDDLocalNode{}, err
	}
	assignments := 1 << len(unique)
	for assignment := 0; assignment < assignments; assignment++ {
		if assignment&1023 == 0 && options.Stop != nil && options.Stop() {
			return ABCBDDLocalNode{}, ErrCancelled
		}
		oldAssignment := uint64(0)
		for source, target := range positions {
			if assignment&(1<<target) != 0 {
				oldAssignment |= uint64(1) << source
			}
		}
		value, valueErr := initial.Table.Get(oldAssignment)
		if valueErr != nil {
			return ABCBDDLocalNode{}, valueErr
		}
		if value {
			if err = collapsed.Set(uint64(assignment), true); err != nil {
				return ABCBDDLocalNode{}, err
			}
		}
	}
	reduction, err := collapsed.CompactSupport()
	if err != nil {
		return ABCBDDLocalNode{}, err
	}
	minimum := make([]int, len(reduction.SourceVariables))
	for index, source := range reduction.SourceVariables {
		minimum[index] = unique[source]
	}
	return ABCBDDLocalNode{Name: node.Name, Fanins: minimum, Function: reduction.Table}, nil
}

func abcBDDTruthValues(table *TruthTable, stop func() bool) ([]bool, error) {
	values := make([]bool, 1<<table.Variables())
	for assignment := range values {
		if assignment&1023 == 0 && stop != nil && stop() {
			return nil, ErrCancelled
		}
		value, err := table.Get(uint64(assignment))
		if err != nil {
			return nil, err
		}
		values[assignment] = value
	}
	return values, nil
}

func abcBDDIdentityOrder(order []int) bool {
	for index, variable := range order {
		if index != variable {
			return false
		}
	}
	return true
}

// ABCNodeBDDReorder removes duplicate and unused local inputs, finds a bounded
// deterministic BDD variable order, and relabels both the fanins and truth
// table so the externally observed Boolean function is unchanged.
func ABCNodeBDDReorder(node ABCBDDLocalNode, options ABCBDDReorderOptions) (ABCBDDReorderedNode, error) {
	options, err := normalizeABCBDDReorderOptions(options)
	if err != nil {
		return ABCBDDReorderedNode{}, err
	}
	if options.Stop != nil && options.Stop() {
		return ABCBDDReorderedNode{}, ErrCancelled
	}
	normalized, err := abcBDDNormalizeNode(node, options)
	if err != nil {
		return ABCBDDReorderedNode{}, err
	}
	result := ABCBDDReorderedNode{
		Name: normalized.Name, Fanins: slices.Clone(normalized.Fanins), Function: normalized.Function.Clone(),
		Stats: ABCBDDReorderStats{FaninsBefore: len(node.Fanins), FaninsAfter: len(normalized.Fanins)},
	}
	variables := len(normalized.Fanins)
	result.Order = make([]int, variables)
	for index := range result.Order {
		result.Order[index] = index
	}
	if variables < 3 {
		return result, nil
	}
	values, err := abcBDDTruthValues(normalized.Function, options.Stop)
	if err != nil {
		return ABCBDDReorderedNode{}, err
	}
	problem, err := NewCuddTruthReorderProblem(variables, [][]bool{values})
	if err != nil {
		return ABCBDDReorderedNode{}, err
	}
	reordered, err := problem.Reorder(result.Order, options.Method, CuddReorderOptions{MaxEvaluations: options.MaxEvaluations, MaxPasses: options.MaxPasses, Stop: options.Stop})
	if err != nil {
		return ABCBDDReorderedNode{}, err
	}
	inverse := make([]int, variables)
	result.Fanins = make([]int, variables)
	for target, source := range reordered.Order {
		inverse[source] = target
		result.Fanins[target] = normalized.Fanins[source]
	}
	result.Function, err = normalized.Function.PermuteVariables(inverse)
	if err != nil {
		return ABCBDDReorderedNode{}, err
	}
	result.Order = slices.Clone(reordered.Order)
	result.Stats.NodesBefore = reordered.InitialCost
	result.Stats.NodesAfter = reordered.Cost
	result.Stats.Evaluations = reordered.Evaluations
	result.Stats.Passes = reordered.Passes
	return result, nil
}

// ABCNetworkBDDReorder applies duplicate removal, minimum-base reduction and
// local reordering to every node. Inputs and truth tables remain caller-owned;
// on error no partial network result is returned.
func ABCNetworkBDDReorder(nodes []ABCBDDLocalNode, options ABCBDDReorderOptions) (ABCBDDNetworkReorder, error) {
	options, err := normalizeABCBDDReorderOptions(options)
	if err != nil {
		return ABCBDDNetworkReorder{}, err
	}
	if len(nodes) > options.MaxNodes {
		return ABCBDDNetworkReorder{}, ErrLimit
	}
	result := ABCBDDNetworkReorder{Nodes: make([]ABCBDDReorderedNode, len(nodes))}
	for index, node := range nodes {
		if options.Stop != nil && options.Stop() {
			return ABCBDDNetworkReorder{}, ErrCancelled
		}
		one, oneErr := ABCNodeBDDReorder(node, options)
		if oneErr != nil {
			return ABCBDDNetworkReorder{}, oneErr
		}
		result.Nodes[index] = one
		result.NodesBefore += one.Stats.NodesBefore
		result.NodesAfter += one.Stats.NodesAfter
		result.RemovedFanins += one.Stats.FaninsBefore - one.Stats.FaninsAfter
		if !abcBDDIdentityOrder(one.Order) {
			result.Reordered++
		}
	}
	return result, nil
}
