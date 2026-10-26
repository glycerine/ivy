package abc

// Stand-alone ZDD ordering and traversal behavior from
// src/bdd/cudd/cuddZddGroup.c, cuddZddLin.c, cuddZddReord.c,
// cuddZddSymm.c, and cuddZddUtil.c.

import (
	"sort"
	"strconv"
	"strings"
)

// CuddZDDReorderCoverage records how the source inventory maps to this
// immutable semantic layer. Partial entries preserve externally observable
// family behavior without mutating CUDD unique tables, group trees, caches,
// hooks, reference counts, or allocator state.
func CuddZDDReorderingCoverage() CuddReorderCoverage {
	return CuddReorderCoverage{
		Complete: []string{
			"Cudd_zddCoverPathToString",
		},
		Partial: []string{
			"Cudd_zddReduceHeap/Cudd_zddShuffleHeap/zddShuffle/zddSiftUp",
			"cuddZddSwapInPlace/cuddZddSwapping/zddSwapAny/cuddZddSifting/cuddZddSiftingAux/cuddZddSiftingUp/cuddZddSiftingDown/cuddZddSiftingBackward",
			"Cudd_MakeZddTreeNode/cuddZddTreeSifting/zddTreeSiftingAux/zddCountInternalMtrNodes/zddReorderChildren/zddFindNodeHiLo/zddUniqueCompareGroup/zddGroupSifting/zddGroupSiftingAux/zddGroupSiftingUp/zddGroupSiftingDown/zddGroupMove/zddGroupMoveBackward/zddGroupSiftingBackward/zddMergeGroups",
			"Cudd_zddSymmProfile/cuddZddSymmCheck/cuddZddSymmSifting/cuddZddSymmSiftingConv/cuddZddSymmSiftingAux/cuddZddSymmSiftingConvAux/cuddZddSymmSifting_up/cuddZddSymmSifting_down/cuddZddSymmSiftingBackward/zdd_group_move/zdd_group_move_backward/cuddZddSymmSummary",
			"cuddZddLinearSifting/cuddZddLinearInPlace/cuddZddLinearAux/cuddZddLinearUp/cuddZddLinearDown/cuddZddLinearBackward/cuddZddUndoMoves",
			"Cudd_zddFirstPath/Cudd_zddNextPath",
			"Cudd_zddPrintMinterm/Cudd_zddPrintCover/Cudd_zddPrintDebug/Cudd_zddDumpDot/cuddZddP/zp2/zdd_print_minterm_aux/zddPrintCoverAux",
		},
		NativeOnly: []string{
			"cuddZddAlignToBdd/cuddZddNextHigh/cuddZddNextLow/cuddZddUniqueCompare",
			"zddReorderPreprocess/zddReorderPostprocess/zddFixTree",
			"native move lists, group-tree mutation, hooks, cache flush, garbage collection, and table resizing",
		},
	}
}

// CuddZDDReorderProblem owns sorted families. It is independent of subsequent
// manager mutation and therefore makes failed reorder attempts transactional.
type CuddZDDReorderProblem struct {
	variables int
	families  [][]uint64
}

type CuddZDDOrderedNode struct {
	Variable  int
	High, Low int
}

// Node edges are one-based. Zero is the empty family and -1 is the family
// containing only the empty set.
type CuddZDDOrderedDAG struct {
	Order, Roots []int
	Nodes        []CuddZDDOrderedNode
}

func NewCuddZDDReorderProblem(manager *CuddZDDManager, roots []CuddZDDEdge, options CuddZDDOptions) (*CuddZDDReorderProblem, error) {
	if manager == nil || len(roots) == 0 || len(roots) > 1<<16 {
		return nil, ErrDimension
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return nil, err
	}
	for _, root := range roots {
		if !manager.valid(root) {
			return nil, ErrInvalidLiteral
		}
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		return nil, err
	}
	if len(roots) > normalized.MaxBytes/24 {
		return nil, ErrLimit
	}
	work := &cuddZDDWork{options: normalized, bytes: len(roots) * 24}
	if err := validateCuddZDDRoots(manager, roots, work); err != nil {
		return nil, err
	}
	memo := make(map[CuddZDDEdge]map[uint64]struct{})
	result := &CuddZDDReorderProblem{variables: manager.variables, families: make([][]uint64, len(roots))}
	for index, root := range roots {
		family, decodeErr := manager.decode(root, work, memo)
		if decodeErr != nil {
			return nil, decodeErr
		}
		if len(family) > (normalized.MaxBytes-work.bytes)/8 {
			return nil, ErrLimit
		}
		work.bytes += len(family) * 8
		result.families[index] = make([]uint64, 0, len(family))
		for set := range family {
			if err := work.step(); err != nil {
				return nil, err
			}
			result.families[index] = append(result.families[index], set)
		}
		sort.Slice(result.families[index], func(i, j int) bool { return result.families[index][i] < result.families[index][j] })
	}
	manager.operations += uint64(work.steps)
	return result, nil
}

// validateCuddZDDRoots checks the invariants that the core manager normally
// establishes while charging the caller's cumulative work budget. Keeping the
// check here prevents malformed internal state from turning traversal into a
// panic or unbounded recursion.
func validateCuddZDDRoots(manager *CuddZDDManager, roots []CuddZDDEdge, work *cuddZDDWork) error {
	state := make(map[CuddZDDEdge]uint8)
	var visit func(CuddZDDEdge, int) error
	visit = func(edge CuddZDDEdge, minimumVariable int) error {
		if err := work.step(); err != nil {
			return err
		}
		if !manager.valid(edge) {
			return ErrInvariant
		}
		if edge <= CuddZDDBase {
			return nil
		}
		node := manager.nodes[edge]
		variable := int(node.variable)
		if variable < minimumVariable || variable >= manager.variables || node.high == CuddZDDEmpty {
			return ErrInvariant
		}
		if state[edge] == 1 {
			return ErrInvariant
		}
		if state[edge] == 2 {
			return nil
		}
		if canonical, found := manager.unique[cuddZDDKey{variable: node.variable, high: node.high, low: node.low}]; !found || canonical != edge {
			return ErrInvariant
		}
		if work.bytes > work.options.MaxBytes-16 {
			return ErrLimit
		}
		work.bytes += 16
		state[edge] = 1
		if err := visit(node.low, variable+1); err != nil {
			return err
		}
		if err := visit(node.high, variable+1); err != nil {
			return err
		}
		state[edge] = 2
		return nil
	}
	for _, root := range roots {
		if err := visit(root, 0); err != nil {
			return err
		}
	}
	return nil
}

func (problem *CuddZDDReorderProblem) Variables() int {
	if problem == nil {
		return 0
	}
	return problem.variables
}

func (problem *CuddZDDReorderProblem) valid() bool {
	if problem == nil || problem.variables < 0 || problem.variables > 63 || len(problem.families) == 0 {
		return false
	}
	limit := uint64(1)<<problem.variables - 1
	for _, family := range problem.families {
		for i, set := range family {
			if set&^limit != 0 || i != 0 && family[i-1] >= set {
				return false
			}
		}
	}
	return true
}

type cuddZDDOrderKey struct {
	variable, high, low int
}

const cuddZDDReorderMaxBytes = 64 << 20

func (problem *CuddZDDReorderProblem) canonical(order []int, stop func() bool) (CuddZDDOrderedDAG, error) {
	if !problem.valid() || !cuddValidOrder(order, problem.variables) {
		return CuddZDDOrderedDAG{}, ErrDimension
	}
	dag := CuddZDDOrderedDAG{Order: append([]int(nil), order...), Roots: make([]int, len(problem.families))}
	unique := make(map[cuddZDDOrderKey]int)
	steps := 0
	allocated := len(problem.families)*24 + len(order)*8
	var build func([]uint64, int) (int, error)
	build = func(family []uint64, level int) (int, error) {
		steps++
		if stop != nil && steps&255 == 0 && stop() {
			return 0, ErrCancelled
		}
		if len(family) == 0 {
			return 0, nil
		}
		if level == problem.variables {
			if len(family) == 1 && family[0] == 0 {
				return -1, nil
			}
			return 0, ErrInvariant
		}
		variable := order[level]
		bit := uint64(1) << variable
		if len(family) > (cuddZDDReorderMaxBytes-allocated)/8 {
			return 0, ErrLimit
		}
		allocated += len(family) * 8
		low := make([]uint64, 0, len(family))
		high := make([]uint64, 0, len(family))
		for _, set := range family {
			steps++
			if stop != nil && steps&255 == 0 && stop() {
				return 0, ErrCancelled
			}
			if set&bit == 0 {
				low = append(low, set)
			} else {
				high = append(high, set&^bit)
			}
		}
		highEdge, err := build(high, level+1)
		if err != nil {
			return 0, err
		}
		lowEdge, err := build(low, level+1)
		if err != nil {
			return 0, err
		}
		if highEdge == 0 {
			return lowEdge, nil
		}
		key := cuddZDDOrderKey{variable: variable, high: highEdge, low: lowEdge}
		if edge, found := unique[key]; found {
			return edge, nil
		}
		if len(dag.Nodes) >= 1<<20 {
			return 0, ErrLimit
		}
		if allocated > cuddZDDReorderMaxBytes-32 {
			return 0, ErrLimit
		}
		allocated += 32
		dag.Nodes = append(dag.Nodes, CuddZDDOrderedNode{Variable: variable, High: highEdge, Low: lowEdge})
		edge := len(dag.Nodes)
		unique[key] = edge
		return edge, nil
	}
	for index, family := range problem.families {
		root, err := build(family, 0)
		if err != nil {
			return CuddZDDOrderedDAG{}, err
		}
		dag.Roots[index] = root
	}
	return dag, nil
}

func (problem *CuddZDDReorderProblem) Canonical(order []int) (CuddZDDOrderedDAG, error) {
	return problem.canonical(order, nil)
}

func (problem *CuddZDDReorderProblem) Cost(order []int) (int, error) {
	dag, err := problem.canonical(order, nil)
	if err != nil {
		return 0, err
	}
	return len(dag.Nodes), nil
}

func (dag CuddZDDOrderedDAG) Families() ([][]uint64, error) {
	if !cuddValidOrder(dag.Order, len(dag.Order)) {
		return nil, ErrDimension
	}
	level := make([]int, len(dag.Order))
	for position, variable := range dag.Order {
		level[variable] = position
	}
	memo := make(map[int][]uint64)
	visiting := make(map[int]bool)
	unique := make(map[cuddZDDOrderKey]int)
	allocated := len(dag.Roots) * 24
	var decode func(int, int) ([]uint64, error)
	decode = func(edge, minimumLevel int) ([]uint64, error) {
		if edge == 0 {
			return nil, nil
		}
		if edge == -1 {
			return []uint64{0}, nil
		}
		if edge < 1 || edge > len(dag.Nodes) || visiting[edge] {
			return nil, ErrInvariant
		}
		node := dag.Nodes[edge-1]
		if node.Variable < 0 || node.Variable >= len(dag.Order) || level[node.Variable] < minimumLevel || node.High == 0 {
			return nil, ErrInvariant
		}
		key := cuddZDDOrderKey{variable: node.Variable, high: node.High, low: node.Low}
		if canonical, found := unique[key]; found && canonical != edge {
			return nil, ErrInvariant
		}
		unique[key] = edge
		if family, ok := memo[edge]; ok {
			return family, nil
		}
		visiting[edge] = true
		nextLevel := level[node.Variable] + 1
		low, err := decode(node.Low, nextLevel)
		if err != nil {
			return nil, err
		}
		high, err := decode(node.High, nextLevel)
		if err != nil {
			return nil, err
		}
		if len(low) > (cuddZDDReorderMaxBytes-allocated)/8-len(high) {
			return nil, ErrLimit
		}
		allocated += (len(low) + len(high)) * 8
		family := append([]uint64(nil), low...)
		bit := uint64(1) << node.Variable
		for _, set := range high {
			family = append(family, set|bit)
		}
		sort.Slice(family, func(i, j int) bool { return family[i] < family[j] })
		visiting[edge] = false
		memo[edge] = family
		return family, nil
	}
	result := make([][]uint64, len(dag.Roots))
	for i, root := range dag.Roots {
		family, err := decode(root, 0)
		if err != nil {
			return nil, err
		}
		if len(family) > (cuddZDDReorderMaxBytes-allocated)/8 {
			return nil, ErrLimit
		}
		allocated += len(family) * 8
		result[i] = append([]uint64(nil), family...)
	}
	return result, nil
}

func (problem *CuddZDDReorderProblem) InteractionMatrix() (CuddInteractionMatrix, error) {
	if !problem.valid() {
		return CuddInteractionMatrix{}, ErrDimension
	}
	matrix := CuddInteractionMatrix{variables: problem.variables, bits: make([]uint64, (problem.variables*problem.variables+63)/64)}
	for _, family := range problem.families {
		var support uint64
		for _, set := range family {
			support |= set
		}
		for first := 0; first < problem.variables; first++ {
			if support&(uint64(1)<<first) == 0 {
				continue
			}
			for second := first + 1; second < problem.variables; second++ {
				if support&(uint64(1)<<second) != 0 {
					matrix.set(first, second)
				}
			}
		}
	}
	return matrix, nil
}

func (problem *CuddZDDReorderProblem) Symmetric(first, second int) (bool, error) {
	return problem.symmetric(first, second, nil)
}

func (problem *CuddZDDReorderProblem) symmetric(first, second int, stop func() bool) (bool, error) {
	if !problem.valid() || first < 0 || second < 0 || first >= problem.variables || second >= problem.variables {
		return false, ErrDimension
	}
	if first == second {
		return true, nil
	}
	mask := uint64(1)<<first | uint64(1)<<second
	steps := 0
	for _, family := range problem.families {
		for _, set := range family {
			steps++
			if stop != nil && steps&255 == 0 && stop() {
				return false, ErrCancelled
			}
			if (set>>first)&1 != (set>>second)&1 {
				target := set ^ mask
				index := sort.Search(len(family), func(i int) bool { return family[i] >= target })
				if index == len(family) || family[index] != target {
					return false, nil
				}
			}
		}
	}
	return true, nil
}

func (problem *CuddZDDReorderProblem) SymmetryGroups(order []int) ([]CuddReorderGroup, error) {
	return problem.symmetryGroups(order, nil)
}

func (problem *CuddZDDReorderProblem) symmetryGroups(order []int, stop func() bool) ([]CuddReorderGroup, error) {
	if !problem.valid() || !cuddValidOrder(order, problem.variables) {
		return nil, ErrDimension
	}
	groups := make([]CuddReorderGroup, 0)
	for start := 0; start < len(order); {
		end := start + 1
		for end < len(order) {
			symmetric, err := problem.symmetric(order[end-1], order[end], stop)
			if err != nil {
				return nil, err
			}
			if !symmetric {
				break
			}
			end++
		}
		if end-start > 1 {
			groups = append(groups, CuddReorderGroup{Variables: append([]int(nil), order[start:end]...)})
		}
		start = end
	}
	return groups, nil
}

type cuddZDDReorderWork struct {
	problem             *CuddZDDReorderProblem
	options             CuddReorderOptions
	evaluations, passes int
}

func (work *cuddZDDReorderWork) cost(order []int) (int, error) {
	if work.evaluations >= work.options.MaxEvaluations {
		return 0, ErrLimit
	}
	if work.options.Stop != nil && work.options.Stop() {
		return 0, ErrCancelled
	}
	work.evaluations++
	dag, err := work.problem.canonical(order, work.options.Stop)
	if err != nil {
		return 0, err
	}
	return len(dag.Nodes), nil
}

func (problem *CuddZDDReorderProblem) Reorder(order []int, method CuddReorderMethod, options CuddReorderOptions) (CuddReorderResult, error) {
	if !problem.valid() || !cuddValidOrder(order, problem.variables) {
		return CuddReorderResult{}, ErrDimension
	}
	normalized, err := cuddNormalizeReorderOptions(problem.variables, options)
	if err != nil {
		return CuddReorderResult{}, err
	}
	work := cuddZDDReorderWork{problem: problem, options: normalized}
	initial, err := work.cost(order)
	if err != nil {
		return CuddReorderResult{}, err
	}
	best := append([]int(nil), order...)
	bestCost := initial
	switch method {
	case CuddReorderSift, CuddReorderSymmetry:
		err = work.sift(&best, &bestCost)
	case CuddReorderWindow2:
		err = work.window(&best, &bestCost, 2)
	case CuddReorderWindow3:
		err = work.window(&best, &bestCost, 3)
	case CuddReorderWindow4:
		err = work.window(&best, &bestCost, 4)
	case CuddReorderExact:
		err = work.exact(&best, &bestCost)
	default:
		err = ErrDimension
	}
	if err != nil {
		return CuddReorderResult{}, err
	}
	groups, err := problem.symmetryGroups(best, normalized.Stop)
	if err != nil {
		return CuddReorderResult{}, err
	}
	return CuddReorderResult{Order: best, InitialCost: initial, Cost: bestCost, Evaluations: work.evaluations, Passes: work.passes, Groups: groups}, nil
}

// ReorderGroups searches block permutations while preserving the order inside
// each group. Groups must be disjoint and contiguous in the supplied order.
func (problem *CuddZDDReorderProblem) ReorderGroups(order []int, groups []CuddReorderGroup, options CuddReorderOptions) (CuddReorderResult, error) {
	if !problem.valid() || !cuddValidOrder(order, problem.variables) {
		return CuddReorderResult{}, ErrDimension
	}
	owner := make([]int, problem.variables)
	for i := range owner {
		owner[i] = -1
	}
	for groupIndex, group := range groups {
		if len(group.Variables) == 0 {
			return CuddReorderResult{}, ErrDimension
		}
		for _, variable := range group.Variables {
			if variable < 0 || variable >= problem.variables || owner[variable] != -1 {
				return CuddReorderResult{}, ErrDimension
			}
			owner[variable] = groupIndex
		}
	}
	blocks := make([][]int, 0, problem.variables)
	seen := make([]bool, len(groups))
	for position := 0; position < len(order); {
		group := owner[order[position]]
		if group < 0 {
			blocks = append(blocks, []int{order[position]})
			position++
			continue
		}
		if seen[group] {
			return CuddReorderResult{}, ErrDimension
		}
		seen[group] = true
		block := make([]int, 0, len(groups[group].Variables))
		for position < len(order) && owner[order[position]] == group {
			block = append(block, order[position])
			position++
		}
		if len(block) != len(groups[group].Variables) {
			return CuddReorderResult{}, ErrDimension
		}
		blocks = append(blocks, block)
	}
	if len(blocks) > 9 {
		return CuddReorderResult{}, ErrLimit
	}
	normalized, err := cuddNormalizeReorderOptions(problem.variables, options)
	if err != nil {
		return CuddReorderResult{}, err
	}
	if normalized.Lower != 0 || normalized.Upper != problem.variables-1 {
		return CuddReorderResult{}, ErrDimension
	}
	work := cuddZDDReorderWork{problem: problem, options: normalized}
	initial, err := work.cost(order)
	if err != nil {
		return CuddReorderResult{}, err
	}
	best := append([]int(nil), order...)
	bestCost := initial
	indices := make([]int, len(blocks))
	for i := range indices {
		indices[i] = i
	}
	err = cuddPermute(indices, func(permutation []int) error {
		candidate := make([]int, 0, len(order))
		for _, block := range permutation {
			candidate = append(candidate, blocks[block]...)
		}
		cost, costErr := work.cost(candidate)
		if costErr != nil {
			return costErr
		}
		cuddTakeCandidate(&best, &bestCost, candidate, cost)
		return nil
	})
	if err != nil {
		return CuddReorderResult{}, err
	}
	ownedGroups := make([]CuddReorderGroup, len(groups))
	for i, group := range groups {
		ownedGroups[i] = CuddReorderGroup{Variables: append([]int(nil), group.Variables...), Hard: group.Hard}
	}
	return CuddReorderResult{Order: best, InitialCost: initial, Cost: bestCost, Evaluations: work.evaluations, Passes: 1, Groups: ownedGroups}, nil
}

func (work *cuddZDDReorderWork) sift(best *[]int, bestCost *int) error {
	for pass := 0; pass < work.options.MaxPasses; pass++ {
		improved := false
		visit := append([]int(nil), (*best)[work.options.Lower:work.options.Upper+1]...)
		for _, variable := range visit {
			local := append([]int(nil), (*best)...)
			localCost := *bestCost
			for target := work.options.Lower; target <= work.options.Upper; target++ {
				candidate := append([]int(nil), (*best)...)
				position := 0
				for candidate[position] != variable {
					position++
				}
				if position < target {
					copy(candidate[position:target], candidate[position+1:target+1])
				} else if position > target {
					copy(candidate[target+1:position+1], candidate[target:position])
				}
				candidate[target] = variable
				cost, err := work.cost(candidate)
				if err != nil {
					return err
				}
				cuddTakeCandidate(&local, &localCost, candidate, cost)
			}
			if cuddTakeCandidate(best, bestCost, local, localCost) {
				improved = true
			}
		}
		work.passes++
		if !improved {
			break
		}
	}
	return nil
}

func (work *cuddZDDReorderWork) window(best *[]int, bestCost *int, width int) error {
	for pass := 0; pass < work.options.MaxPasses; pass++ {
		improved := false
		for start := work.options.Lower; start+width-1 <= work.options.Upper; start++ {
			base := append([]int(nil), (*best)...)
			segment := append([]int(nil), base[start:start+width]...)
			local := append([]int(nil), base...)
			localCost := *bestCost
			err := cuddPermute(segment, func(permutation []int) error {
				candidate := append([]int(nil), base...)
				copy(candidate[start:start+width], permutation)
				cost, costErr := work.cost(candidate)
				if costErr != nil {
					return costErr
				}
				cuddTakeCandidate(&local, &localCost, candidate, cost)
				return nil
			})
			if err != nil {
				return err
			}
			if cuddTakeCandidate(best, bestCost, local, localCost) {
				improved = true
			}
		}
		work.passes++
		if !improved {
			break
		}
	}
	return nil
}

func (work *cuddZDDReorderWork) exact(best *[]int, bestCost *int) error {
	width := work.options.Upper - work.options.Lower + 1
	if width > 10 {
		return ErrLimit
	}
	segment := append([]int(nil), (*best)[work.options.Lower:work.options.Upper+1]...)
	sort.Ints(segment)
	err := cuddPermute(segment, func(permutation []int) error {
		candidate := append([]int(nil), (*best)...)
		copy(candidate[work.options.Lower:work.options.Upper+1], permutation)
		cost, costErr := work.cost(candidate)
		if costErr != nil {
			return costErr
		}
		cuddTakeCandidate(best, bestCost, candidate, cost)
		return nil
	})
	if err == nil {
		work.passes++
	}
	return err
}

// Transform applies an invertible GF(2) coordinate map to each represented
// set. It is the immutable semantic counterpart of ZDD linear sifting.
func (problem *CuddZDDReorderProblem) Transform(mapping CuddLinearMap) (*CuddZDDReorderProblem, error) {
	if !problem.valid() || mapping.variables != problem.variables {
		return nil, ErrDimension
	}
	rows := mapping.Rows()
	inverse := make([]uint64, problem.variables)
	for i := range inverse {
		inverse[i] = uint64(1) << i
	}
	for column := 0; column < problem.variables; column++ {
		pivot := column
		for pivot < problem.variables && rows[pivot]&(uint64(1)<<column) == 0 {
			pivot++
		}
		if pivot == problem.variables {
			return nil, ErrInvariant
		}
		rows[column], rows[pivot] = rows[pivot], rows[column]
		inverse[column], inverse[pivot] = inverse[pivot], inverse[column]
		for row := 0; row < problem.variables; row++ {
			if row != column && rows[row]&(uint64(1)<<column) != 0 {
				rows[row] ^= rows[column]
				inverse[row] ^= inverse[column]
			}
		}
	}
	result := &CuddZDDReorderProblem{variables: problem.variables, families: make([][]uint64, len(problem.families))}
	for root, family := range problem.families {
		result.families[root] = make([]uint64, len(family))
		for i, original := range family {
			var transformed uint64
			for row := 0; row < problem.variables; row++ {
				if bitsOnesParity(inverse[row] & original) {
					transformed |= uint64(1) << row
				}
			}
			result.families[root][i] = transformed
		}
		sort.Slice(result.families[root], func(i, j int) bool { return result.families[root][i] < result.families[root][j] })
		result.families[root] = compactUint64(result.families[root])
	}
	return result, nil
}

func bitsOnesParity(value uint64) bool {
	value ^= value >> 32
	value ^= value >> 16
	value ^= value >> 8
	value ^= value >> 4
	value &= 0xf
	return (0x6996>>value)&1 != 0
}

func compactUint64(values []uint64) []uint64 {
	if len(values) < 2 {
		return values
	}
	write := 1
	for read := 1; read < len(values); read++ {
		if values[read] != values[write-1] {
			values[write] = values[read]
			write++
		}
	}
	return values[:write]
}

// Paths returns owned source-shaped paths: 0 is a low arc, 1 a high arc, and
// 2 means that the variable is absent from that path.
func (manager *CuddZDDManager) Paths(root CuddZDDEdge, options CuddZDDOptions) ([][]uint8, error) {
	if manager == nil {
		return nil, ErrDimension
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return nil, err
	}
	if !manager.valid(root) {
		return nil, ErrInvalidLiteral
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		return nil, err
	}
	baseBytes := manager.variables + 24
	if baseBytes > normalized.MaxBytes {
		return nil, ErrLimit
	}
	work := &cuddZDDWork{options: normalized, bytes: baseBytes}
	if err := validateCuddZDDRoots(manager, []CuddZDDEdge{root}, work); err != nil {
		return nil, err
	}
	path := make([]uint8, manager.variables)
	for i := range path {
		path[i] = 2
	}
	paths := make([][]uint8, 0)
	var visit func(CuddZDDEdge) error
	visit = func(edge CuddZDDEdge) error {
		if err := work.step(); err != nil {
			return err
		}
		if edge == CuddZDDEmpty {
			return nil
		}
		if edge == CuddZDDBase {
			entryBytes := len(path) + 24
			if work.bytes > normalized.MaxBytes-entryBytes {
				return ErrLimit
			}
			work.bytes += entryBytes
			paths = append(paths, append([]uint8(nil), path...))
			return nil
		}
		node := manager.nodes[edge]
		old := path[node.variable]
		path[node.variable] = 0
		if err := visit(node.low); err != nil {
			return err
		}
		path[node.variable] = 1
		if err := visit(node.high); err != nil {
			return err
		}
		path[node.variable] = old
		return nil
	}
	if err := visit(root); err != nil {
		return nil, err
	}
	manager.operations += uint64(work.steps)
	return paths, nil
}

func CuddZDDCoverPathString(path []uint8) (string, error) {
	if len(path)&1 != 0 {
		return "", ErrDimension
	}
	var result strings.Builder
	result.Grow(len(path) / 2)
	for i := 0; i < len(path); i += 2 {
		value := uint16(path[i])<<2 | uint16(path[i+1])
		switch value {
		case 0, 2, 8, 10:
			result.WriteByte('-')
		case 1, 9:
			result.WriteByte('0')
		case 4, 6:
			result.WriteByte('1')
		default:
			result.WriteByte('?')
		}
	}
	return result.String(), nil
}

func (manager *CuddZDDManager) DebugText(root CuddZDDEdge, options CuddZDDOptions) (string, error) {
	paths, err := manager.Paths(root, options)
	if err != nil {
		return "", err
	}
	nodes, err := manager.DagSize(root)
	if err != nil {
		return "", err
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		return "", err
	}
	if len(paths) > (normalized.MaxBytes-64)/max(1, manager.variables+1) {
		return "", ErrLimit
	}
	var text strings.Builder
	text.WriteString("nodes=")
	text.WriteString(strconv.Itoa(nodes))
	text.WriteString(" paths=")
	text.WriteString(strconv.Itoa(len(paths)))
	for _, path := range paths {
		text.WriteByte('\n')
		for _, value := range path {
			text.WriteByte("012"[value])
		}
	}
	if text.Len() > normalized.MaxBytes {
		return "", ErrLimit
	}
	return text.String(), nil
}
