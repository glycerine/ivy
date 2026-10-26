package abc

// Deterministic, bounded variable reordering from src/bdd/cudd/cuddReorder.c,
// cuddGroup.c, cuddSymmetry.c, cuddLinear.c, cuddWindow.c, cuddAnneal.c,
// cuddGenetic.c, cuddExact.c, cuddInteract.c, and cuddLevelQ.c.

import (
	"math"
	"sort"
)

type CuddReorderMethod uint8

const (
	CuddReorderSift CuddReorderMethod = iota
	CuddReorderWindow2
	CuddReorderWindow3
	CuddReorderWindow4
	CuddReorderExact
	CuddReorderAnneal
	CuddReorderGenetic
	CuddReorderSymmetry
)

type CuddReorderOptions struct {
	Lower, Upper              int
	RangeSet                  bool
	MaxEvaluations, MaxPasses int
	Seed                      uint64
	Population, Generations   int
	Stop                      func() bool
}

type CuddReorderResult struct {
	Order               []int
	InitialCost, Cost   int
	Evaluations, Passes int
	Groups              []CuddReorderGroup
}

type CuddReorderGroup struct {
	Variables []int
	Hard      bool
}

// CuddReorderCoverage inventories the native entry points represented by this
// standalone layer. Partial entries preserve mathematical behavior but not
// CUDD's process-global counters, hooks, hash-table layout, dead-node GC, or
// allocation side effects. Native-only entries require those omitted details.
type CuddReorderCoverage struct {
	Complete, Partial, NativeOnly []string
}

func CuddReorderingCoverage() CuddReorderCoverage {
	return CuddReorderCoverage{
		Complete: []string{
			"cuddInitInteract/cuddSetInteract/cuddTestInteract",
		},
		Partial: []string{
			"Cudd_ReduceHeap/Cudd_ShuffleHeap", "cuddSifting/cuddSwapping/cuddSwapInPlace",
			"Cudd_MakeTreeNode/cuddTreeSifting", "Cudd_SymmProfile/cuddSymmCheck/cuddSymmSifting",
			"cuddLinearAndSifting/cuddLinearInPlace/cuddInitLinear/cuddResizeLinear",
			"cuddWindowReorder", "cuddExact", "cuddAnnealing", "cuddGa",
			"cuddLevelQueueInit/cuddLevelQueueEnqueue/cuddLevelQueueDequeue/cuddLevelQueueQuit",
		},
		NativeOnly: []string{
			"cuddDynamicAllocNode", "ddReorderPreprocess/ddReorderPostprocess",
			"cuddNextHigh/cuddNextLow/cuddBddAlignToZdd",
			"native dead-node collection, hooks, counters, and unique-table resizing",
		},
	}
}

// CuddReorderProblem owns dense, manager-independent terminal tables. Table
// index bits are variable indices; hence every returned order denotes the same
// functions. Values are IEEE representations for ADDs and 0/1 for BDDs.
type CuddReorderProblem struct {
	variables int
	tables    [][]uint64
}

const cuddReorderMaxDenseCells = (16 << 20) / 8

func NewCuddTruthReorderProblem(variables int, tables [][]bool) (*CuddReorderProblem, error) {
	if variables < 0 || variables > 20 || len(tables) == 0 || len(tables) > 1<<16 {
		return nil, ErrDimension
	}
	size := 1 << variables
	if len(tables) > cuddReorderMaxDenseCells/size {
		return nil, ErrLimit
	}
	result := &CuddReorderProblem{variables: variables, tables: make([][]uint64, len(tables))}
	for i := range tables {
		if len(tables[i]) != size {
			return nil, ErrDimension
		}
		result.tables[i] = make([]uint64, size)
		for j, value := range tables[i] {
			if value {
				result.tables[i][j] = 1
			}
		}
	}
	return result, nil
}

func NewCuddBDDReorderProblem(manager *ExtraBDDManager, roots []ExtraBDDLit, options ExtraBDDOptions) (*CuddReorderProblem, error) {
	if manager == nil || len(roots) == 0 || len(roots) > 1<<16 || manager.variables > 20 {
		return nil, ErrDimension
	}
	if err := manager.check(); err != nil {
		return nil, err
	}
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return nil, err
	}
	size := 1 << manager.variables
	if int64(size)*int64(len(roots))*8 > int64(options.MaxBytes) {
		return nil, ErrLimit
	}
	tables := make([][]uint64, len(roots))
	operations := 0
	assignment := make([]bool, manager.variables)
	for r, root := range roots {
		if !manager.valid(root) {
			return nil, ErrInvalidLiteral
		}
		tables[r] = make([]uint64, size)
		for bits := 0; bits < size; bits++ {
			operations++
			if operations > options.MaxOperations {
				return nil, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			for variable := range assignment {
				assignment[variable] = bits&(1<<variable) != 0
			}
			value, evalErr := manager.Evaluate(root, assignment)
			if evalErr != nil {
				return nil, evalErr
			}
			if value {
				tables[r][bits] = 1
			}
		}
	}
	return &CuddReorderProblem{variables: manager.variables, tables: tables}, nil
}

func NewCuddADDReorderProblem(manager *CuddADDManager, roots []CuddADDEdge, options ExtraBDDOptions) (*CuddReorderProblem, error) {
	if manager == nil || len(roots) == 0 || len(roots) > 1<<16 || manager.variables > 20 {
		return nil, ErrDimension
	}
	if err := manager.checkEdges(roots...); err != nil {
		return nil, err
	}
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return nil, err
	}
	size := 1 << manager.variables
	if int64(size)*int64(len(roots))*8 > int64(options.MaxBytes) {
		return nil, ErrLimit
	}
	tables := make([][]uint64, len(roots))
	operations := 0
	for r, root := range roots {
		tables[r] = make([]uint64, size)
		for bits := 0; bits < size; bits++ {
			operations++
			if operations > options.MaxOperations {
				return nil, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			value := manager.evaluateUnchecked(root, uint64(bits))
			tables[r][bits] = math.Float64bits(value)
		}
	}
	return &CuddReorderProblem{variables: manager.variables, tables: tables}, nil
}

func (problem *CuddReorderProblem) Variables() int {
	if problem == nil {
		return 0
	}
	return problem.variables
}

func (problem *CuddReorderProblem) valid() bool {
	if problem == nil || problem.variables < 0 || problem.variables > 20 || len(problem.tables) == 0 {
		return false
	}
	size := 1 << problem.variables
	for _, table := range problem.tables {
		if len(table) != size {
			return false
		}
	}
	return true
}

func cuddValidOrder(order []int, variables int) bool {
	if len(order) != variables {
		return false
	}
	seen := make([]bool, variables)
	for _, variable := range order {
		if variable < 0 || variable >= variables || seen[variable] {
			return false
		}
		seen[variable] = true
	}
	return true
}

type cuddReorderNode struct {
	variable, low, high int
}

type CuddOrderedNode struct {
	Variable  int
	Low, High int
}

type CuddOrderedDAG struct {
	Order, Roots []int
	Terminals    []uint64
	Nodes        []CuddOrderedNode
}

// Cost is the number of nonterminal nodes in the shared, reduced, canonical
// ordered DAG for all tables. It is independent of the source managers.
func (problem *CuddReorderProblem) Cost(order []int) (int, error) {
	dag, err := problem.canonical(order, nil)
	if err != nil {
		return 0, err
	}
	return len(dag.Nodes), nil
}

// Canonical rebuilds all roots into one immutable reduced ordered DAG. Positive
// edge IDs are one-based node IDs; negative IDs encode terminals as -index-1.
func (problem *CuddReorderProblem) Canonical(order []int) (CuddOrderedDAG, error) {
	return problem.canonical(order, nil)
}

func (problem *CuddReorderProblem) canonical(order []int, stop func() bool) (CuddOrderedDAG, error) {
	if !problem.valid() || !cuddValidOrder(order, problem.variables) {
		return CuddOrderedDAG{}, ErrDimension
	}
	terminals := make(map[uint64]int)
	unique := make(map[cuddReorderNode]int)
	dag := CuddOrderedDAG{Order: append([]int(nil), order...), Roots: make([]int, len(problem.tables))}
	next := 1
	steps := 0
	var build func([]uint64, int, int) (int, error)
	build = func(table []uint64, depth, assignment int) (int, error) {
		steps++
		if steps&1023 == 1 && stop != nil && stop() {
			return 0, ErrCancelled
		}
		if depth == problem.variables {
			value := table[assignment]
			if id, found := terminals[value]; found {
				return id, nil
			}
			id := -len(terminals) - 1
			terminals[value] = id
			dag.Terminals = append(dag.Terminals, value)
			return id, nil
		}
		variable := order[depth]
		low, err := build(table, depth+1, assignment)
		if err != nil {
			return 0, err
		}
		high, err := build(table, depth+1, assignment|1<<variable)
		if err != nil {
			return 0, err
		}
		if low == high {
			return low, nil
		}
		key := cuddReorderNode{variable: variable, low: low, high: high}
		if id, found := unique[key]; found {
			return id, nil
		}
		id := next
		next++
		unique[key] = id
		dag.Nodes = append(dag.Nodes, CuddOrderedNode{Variable: variable, Low: low, High: high})
		return id, nil
	}
	for i, table := range problem.tables {
		root, err := build(table, 0, 0)
		if err != nil {
			return CuddOrderedDAG{}, err
		}
		dag.Roots[i] = root
	}
	return dag, nil
}

func (dag CuddOrderedDAG) Value(root int, assignment uint64) (uint64, error) {
	if root < 0 || root >= len(dag.Roots) {
		return 0, ErrDimension
	}
	edge := dag.Roots[root]
	steps := 0
	for edge > 0 {
		steps++
		if steps > len(dag.Nodes) {
			return 0, ErrInvariant
		}
		if edge > len(dag.Nodes) {
			return 0, ErrInvariant
		}
		node := dag.Nodes[edge-1]
		if node.Variable < 0 || node.Variable >= len(dag.Order) {
			return 0, ErrInvariant
		}
		if assignment&(uint64(1)<<node.Variable) != 0 {
			edge = node.High
		} else {
			edge = node.Low
		}
	}
	terminal := -edge - 1
	if terminal < 0 || terminal >= len(dag.Terminals) {
		return 0, ErrInvariant
	}
	return dag.Terminals[terminal], nil
}

type CuddInteractionMatrix struct {
	variables int
	bits      []uint64
}

func (problem *CuddReorderProblem) InteractionMatrix() (CuddInteractionMatrix, error) {
	if !problem.valid() {
		return CuddInteractionMatrix{}, ErrDimension
	}
	matrix := CuddInteractionMatrix{variables: problem.variables, bits: make([]uint64, (problem.variables*problem.variables+63)/64)}
	for _, table := range problem.tables {
		support := make([]bool, problem.variables)
		for variable := range support {
			mask := 1 << variable
			for assignment := 0; assignment < len(table); assignment++ {
				if assignment&mask == 0 && table[assignment] != table[assignment|mask] {
					support[variable] = true
					break
				}
			}
		}
		for first := 0; first < problem.variables; first++ {
			if !support[first] {
				continue
			}
			for second := first + 1; second < problem.variables; second++ {
				if support[second] {
					matrix.set(first, second)
				}
			}
		}
	}
	return matrix, nil
}

func (matrix *CuddInteractionMatrix) set(first, second int) {
	position := first*matrix.variables + second
	matrix.bits[position>>6] |= uint64(1) << (position & 63)
	position = second*matrix.variables + first
	matrix.bits[position>>6] |= uint64(1) << (position & 63)
}

func (matrix CuddInteractionMatrix) Interacts(first, second int) bool {
	if first < 0 || second < 0 || first >= matrix.variables || second >= matrix.variables || first == second {
		return false
	}
	position := first*matrix.variables + second
	return matrix.bits[position>>6]&(uint64(1)<<(position&63)) != 0
}

func (problem *CuddReorderProblem) Symmetric(first, second int) (bool, error) {
	if !problem.valid() || first < 0 || second < 0 || first >= problem.variables || second >= problem.variables {
		return false, ErrDimension
	}
	if first == second {
		return true, nil
	}
	mask := 1<<first | 1<<second
	for _, table := range problem.tables {
		for assignment := 0; assignment < len(table); assignment++ {
			if (assignment>>first)&1 != (assignment>>second)&1 && table[assignment] != table[assignment^mask] {
				return false, nil
			}
		}
	}
	return true, nil
}

func (problem *CuddReorderProblem) SymmetryGroups(order []int) ([]CuddReorderGroup, error) {
	if !problem.valid() || !cuddValidOrder(order, problem.variables) {
		return nil, ErrDimension
	}
	groups := make([]CuddReorderGroup, 0)
	for start := 0; start < len(order); {
		end := start + 1
		for end < len(order) {
			symmetric, err := problem.Symmetric(order[end-1], order[end])
			if err != nil || !symmetric {
				break
			}
			end++
		}
		if end-start > 1 {
			variables := append([]int(nil), order[start:end]...)
			groups = append(groups, CuddReorderGroup{Variables: variables})
		}
		start = end
	}
	return groups, nil
}

func (problem *CuddReorderProblem) AdjacentSwap(order []int, level int) ([]int, int, error) {
	if !problem.valid() || !cuddValidOrder(order, problem.variables) || level < 0 || level+1 >= len(order) {
		return nil, 0, ErrDimension
	}
	result := append([]int(nil), order...)
	result[level], result[level+1] = result[level+1], result[level]
	cost, err := problem.Cost(result)
	if err != nil {
		return nil, 0, err
	}
	return result, cost, nil
}

// ReorderGroups finds the best block order while retaining the current order
// inside each listed group. Groups must be disjoint and contiguous in order.
// The exact block search is deliberately limited to nine blocks.
func (problem *CuddReorderProblem) ReorderGroups(order []int, groups []CuddReorderGroup, options CuddReorderOptions) (CuddReorderResult, error) {
	if !problem.valid() || !cuddValidOrder(order, problem.variables) {
		return CuddReorderResult{}, ErrDimension
	}
	if options.RangeSet && (options.Lower != 0 || options.Upper != problem.variables-1) ||
		!options.RangeSet && (options.Lower != 0 || options.Upper != 0 && options.Upper != problem.variables-1) {
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
	seenGroup := make([]bool, len(groups))
	for position := 0; position < len(order); {
		groupIndex := owner[order[position]]
		if groupIndex < 0 {
			blocks = append(blocks, []int{order[position]})
			position++
			continue
		}
		if seenGroup[groupIndex] {
			return CuddReorderResult{}, ErrDimension
		}
		seenGroup[groupIndex] = true
		block := make([]int, 0, len(groups[groupIndex].Variables))
		for position < len(order) && owner[order[position]] == groupIndex {
			block = append(block, order[position])
			position++
		}
		if len(block) != len(groups[groupIndex].Variables) {
			return CuddReorderResult{}, ErrDimension
		}
		blocks = append(blocks, block)
	}
	if len(blocks) > 9 {
		return CuddReorderResult{}, ErrLimit
	}
	options, err := cuddNormalizeReorderOptions(problem.variables, options)
	if err != nil {
		return CuddReorderResult{}, err
	}
	work := cuddReorderWork{problem: problem, options: options}
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
	resultGroups := make([]CuddReorderGroup, len(groups))
	for i := range groups {
		resultGroups[i] = CuddReorderGroup{Variables: append([]int(nil), groups[i].Variables...), Hard: groups[i].Hard}
	}
	return CuddReorderResult{Order: best, InitialCost: initial, Cost: bestCost, Evaluations: work.evaluations, Passes: 1, Groups: resultGroups}, nil
}

type cuddReorderWork struct {
	problem             *CuddReorderProblem
	options             CuddReorderOptions
	evaluations, passes int
}

func cuddNormalizeReorderOptions(variables int, options CuddReorderOptions) (CuddReorderOptions, error) {
	if variables < 0 || options.Lower < 0 || options.MaxEvaluations < 0 || options.MaxPasses < 0 || options.Population < 0 || options.Generations < 0 {
		return options, ErrDimension
	}
	if !options.RangeSet && options.Lower == 0 && options.Upper == 0 {
		options.Upper = variables - 1
	}
	if variables == 0 {
		options.Lower, options.Upper = 0, -1
	} else if options.Lower > options.Upper || options.Upper >= variables {
		return options, ErrDimension
	}
	if options.MaxEvaluations == 0 {
		options.MaxEvaluations = 1 << 20
	}
	if options.MaxPasses == 0 {
		options.MaxPasses = max(1, variables*4)
	}
	if options.Population == 0 {
		options.Population = max(4, variables*2)
	}
	if options.Population < 4 {
		options.Population = 4
	}
	if options.Generations == 0 {
		options.Generations = max(4, variables*3)
	}
	if options.Seed == 0 {
		options.Seed = 0x9e3779b97f4a7c15
	}
	return options, nil
}

func (work *cuddReorderWork) cost(order []int) (int, error) {
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

func cuddOrderLess(first, second []int) bool {
	for i := range first {
		if first[i] != second[i] {
			return first[i] < second[i]
		}
	}
	return false
}

func cuddTakeCandidate(best *[]int, bestCost *int, candidate []int, cost int) bool {
	if cost < *bestCost || cost == *bestCost && cuddOrderLess(candidate, *best) {
		*best = append((*best)[:0], candidate...)
		*bestCost = cost
		return true
	}
	return false
}

func (problem *CuddReorderProblem) Reorder(order []int, method CuddReorderMethod, options CuddReorderOptions) (CuddReorderResult, error) {
	if !problem.valid() || !cuddValidOrder(order, problem.variables) {
		return CuddReorderResult{}, ErrDimension
	}
	options, err := cuddNormalizeReorderOptions(problem.variables, options)
	if err != nil {
		return CuddReorderResult{}, err
	}
	work := &cuddReorderWork{problem: problem, options: options}
	initial, err := work.cost(order)
	if err != nil {
		return CuddReorderResult{}, err
	}
	best := append([]int(nil), order...)
	bestCost := initial
	switch method {
	case CuddReorderSift:
		err = work.sift(&best, &bestCost)
	case CuddReorderWindow2:
		err = work.window(&best, &bestCost, 2)
	case CuddReorderWindow3:
		err = work.window(&best, &bestCost, 3)
	case CuddReorderWindow4:
		err = work.window(&best, &bestCost, 4)
	case CuddReorderExact:
		err = work.exact(&best, &bestCost)
	case CuddReorderAnneal:
		err = work.anneal(&best, &bestCost)
	case CuddReorderGenetic:
		err = work.genetic(&best, &bestCost)
	case CuddReorderSymmetry:
		err = work.symmetry(&best, &bestCost)
	default:
		err = ErrDimension
	}
	if err != nil {
		return CuddReorderResult{}, err
	}
	groups, err := problem.SymmetryGroups(best)
	if err != nil {
		return CuddReorderResult{}, err
	}
	return CuddReorderResult{Order: best, InitialCost: initial, Cost: bestCost, Evaluations: work.evaluations, Passes: work.passes, Groups: groups}, nil
}

func (work *cuddReorderWork) sift(best *[]int, bestCost *int) error {
	lower, upper := work.options.Lower, work.options.Upper
	for pass := 0; pass < work.options.MaxPasses; pass++ {
		improved := false
		visitOrder := append([]int(nil), (*best)[lower:upper+1]...)
		for _, variable := range visitOrder {
			local := append([]int(nil), (*best)...)
			localCost := *bestCost
			for target := lower; target <= upper; target++ {
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

func cuddPermute(values []int, visit func([]int) error) error {
	var generate func(int) error
	generate = func(index int) error {
		if index == len(values) {
			return visit(values)
		}
		for i := index; i < len(values); i++ {
			values[index], values[i] = values[i], values[index]
			if err := generate(index + 1); err != nil {
				return err
			}
			values[index], values[i] = values[i], values[index]
		}
		return nil
	}
	return generate(0)
}

func (work *cuddReorderWork) window(best *[]int, bestCost *int, width int) error {
	if width < 2 {
		return ErrDimension
	}
	for pass := 0; pass < work.options.MaxPasses; pass++ {
		improved := false
		for start := work.options.Lower; start+width-1 <= work.options.Upper; start++ {
			window := append([]int(nil), (*best)[start:start+width]...)
			local := append([]int(nil), (*best)...)
			localCost := *bestCost
			err := cuddPermute(window, func(permutation []int) error {
				candidate := append([]int(nil), (*best)...)
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

func (work *cuddReorderWork) exact(best *[]int, bestCost *int) error {
	width := work.options.Upper - work.options.Lower + 1
	if width > 10 {
		return ErrLimit
	}
	segment := append([]int(nil), (*best)[work.options.Lower:work.options.Upper+1]...)
	sort.Ints(segment)
	err := cuddPermute(segment, func(permutation []int) error {
		candidate := append([]int(nil), (*best)...)
		copy(candidate[work.options.Lower:work.options.Upper+1], permutation)
		cost, err := work.cost(candidate)
		if err != nil {
			return err
		}
		cuddTakeCandidate(best, bestCost, candidate, cost)
		return nil
	})
	if err == nil {
		work.passes++
	}
	return err
}

type cuddReorderRNG struct{ state uint64 }

func (rng *cuddReorderRNG) next() uint64 {
	x := rng.state
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	rng.state = x
	return x * 2685821657736338717
}
func (rng *cuddReorderRNG) integer(n int) int { return int(rng.next() % uint64(n)) }

func (work *cuddReorderWork) anneal(best *[]int, bestCost *int) error {
	width := work.options.Upper - work.options.Lower + 1
	if width < 2 {
		return nil
	}
	rng := cuddReorderRNG{state: work.options.Seed}
	current := append([]int(nil), (*best)...)
	currentCost := *bestCost
	temperature := float64(max(1, currentCost))
	for pass := 0; pass < work.options.MaxPasses; pass++ {
		for step := 0; step < width*width; step++ {
			first := work.options.Lower + rng.integer(width)
			second := work.options.Lower + rng.integer(width)
			if first == second {
				continue
			}
			candidate := append([]int(nil), current...)
			candidate[first], candidate[second] = candidate[second], candidate[first]
			cost, err := work.cost(candidate)
			if err != nil {
				return err
			}
			delta := cost - currentCost
			accept := delta <= 0
			if !accept {
				threshold := math.Exp(-float64(delta) / temperature)
				accept = float64(rng.next()>>11)*(1.0/(1<<53)) < threshold
			}
			if accept {
				current, currentCost = candidate, cost
			}
			cuddTakeCandidate(best, bestCost, candidate, cost)
		}
		work.passes++
		temperature *= .72
		if temperature < .05 {
			break
		}
	}
	return nil
}

type cuddReorderIndividual struct {
	order []int
	cost  int
}

func (work *cuddReorderWork) genetic(best *[]int, bestCost *int) error {
	width := work.options.Upper - work.options.Lower + 1
	if width < 2 {
		return nil
	}
	rng := cuddReorderRNG{state: work.options.Seed}
	population := make([]cuddReorderIndividual, work.options.Population)
	for i := range population {
		candidate := append([]int(nil), (*best)...)
		if i != 0 {
			for j := width - 1; j > 0; j-- {
				k := rng.integer(j + 1)
				candidate[work.options.Lower+j], candidate[work.options.Lower+k] = candidate[work.options.Lower+k], candidate[work.options.Lower+j]
			}
		}
		cost, err := work.cost(candidate)
		if err != nil {
			return err
		}
		population[i] = cuddReorderIndividual{candidate, cost}
		cuddTakeCandidate(best, bestCost, candidate, cost)
	}
	less := func(i, j int) bool {
		return population[i].cost < population[j].cost || population[i].cost == population[j].cost && cuddOrderLess(population[i].order, population[j].order)
	}
	for generation := 0; generation < work.options.Generations; generation++ {
		sort.SliceStable(population, less)
		next := make([]cuddReorderIndividual, 0, len(population))
		next = append(next, cuddReorderIndividual{append([]int(nil), population[0].order...), population[0].cost})
		for len(next) < cap(next) {
			first := population[rng.integer(max(2, len(population)/2))].order
			second := population[rng.integer(max(2, len(population)/2))].order
			child := append([]int(nil), first...)
			cut := rng.integer(width)
			used := make([]bool, work.problem.variables)
			for i := 0; i < cut; i++ {
				used[child[work.options.Lower+i]] = true
			}
			position := cut
			for i := 0; i < width; i++ {
				value := second[work.options.Lower+i]
				if !used[value] {
					child[work.options.Lower+position] = value
					used[value] = true
					position++
				}
			}
			if rng.integer(3) == 0 {
				a, b := work.options.Lower+rng.integer(width), work.options.Lower+rng.integer(width)
				child[a], child[b] = child[b], child[a]
			}
			cost, err := work.cost(child)
			if err != nil {
				return err
			}
			next = append(next, cuddReorderIndividual{child, cost})
			cuddTakeCandidate(best, bestCost, child, cost)
		}
		population = next
		work.passes++
	}
	return nil
}

func (work *cuddReorderWork) symmetry(best *[]int, bestCost *int) error {
	groups, err := work.problem.SymmetryGroups(*best)
	if err != nil {
		return err
	}
	for _, group := range groups {
		sorted := append([]int(nil), group.Variables...)
		sort.Ints(sorted)
		candidate := append([]int(nil), (*best)...)
		position := 0
		for i, variable := range candidate {
			for _, member := range group.Variables {
				if variable == member {
					candidate[i] = sorted[position]
					position++
					break
				}
			}
		}
		cost, costErr := work.cost(candidate)
		if costErr != nil {
			return costErr
		}
		cuddTakeCandidate(best, bestCost, candidate, cost)
	}
	work.passes++
	return work.sift(best, bestCost)
}

// CuddLinearMap represents an invertible GF(2) change of input basis. Rows
// give original variables as XORs of transformed variables. Reordering itself
// does not silently apply this semantic coordinate change.
type CuddLinearMap struct {
	variables int
	rows      [64]uint64
}

func NewCuddLinearMap(variables int) (CuddLinearMap, error) {
	if variables < 0 || variables > 64 {
		return CuddLinearMap{}, ErrDimension
	}
	result := CuddLinearMap{variables: variables}
	for i := 0; i < variables; i++ {
		result.rows[i] = uint64(1) << i
	}
	return result, nil
}

func (mapping *CuddLinearMap) Combine(target, source int) error {
	if mapping == nil || target < 0 || source < 0 || target >= mapping.variables || source >= mapping.variables || target == source {
		return ErrDimension
	}
	mapping.rows[target] ^= mapping.rows[source]
	return nil
}

func (mapping CuddLinearMap) Assignment(transformed uint64) (uint64, error) {
	if mapping.variables < 0 || mapping.variables > len(mapping.rows) {
		return 0, ErrDimension
	}
	var original uint64
	for variable, row := range mapping.rows[:mapping.variables] {
		if uint64(bitsOnesCount(row&transformed))&1 != 0 {
			original |= uint64(1) << variable
		}
	}
	return original, nil
}

func (mapping CuddLinearMap) Rows() []uint64 {
	if mapping.variables < 0 || mapping.variables > len(mapping.rows) {
		return nil
	}
	return append([]uint64(nil), mapping.rows[:mapping.variables]...)
}

func bitsOnesCount(value uint64) int {
	count := 0
	for value != 0 {
		value &= value - 1
		count++
	}
	return count
}

type cuddLevelQueueItem struct {
	key   uint64
	level int
}

type CuddLevelQueue struct {
	levels  int
	items   []cuddLevelQueueItem
	present map[uint64]bool
}

func NewCuddLevelQueue(levels int) (*CuddLevelQueue, error) {
	if levels <= 0 {
		return nil, ErrDimension
	}
	return &CuddLevelQueue{levels: levels, present: make(map[uint64]bool)}, nil
}

func (queue *CuddLevelQueue) Enqueue(key uint64, level int) (bool, error) {
	if queue == nil || queue.present == nil || level < 0 || level >= queue.levels {
		return false, ErrDimension
	}
	if queue.present[key] {
		return false, nil
	}
	index := sort.Search(len(queue.items), func(i int) bool { return queue.items[i].level > level })
	queue.items = append(queue.items, cuddLevelQueueItem{})
	copy(queue.items[index+1:], queue.items[index:])
	queue.items[index] = cuddLevelQueueItem{key, level}
	queue.present[key] = true
	return true, nil
}

func (queue *CuddLevelQueue) Dequeue() (uint64, int, bool) {
	if queue == nil || len(queue.items) == 0 {
		return 0, 0, false
	}
	item := queue.items[0]
	copy(queue.items, queue.items[1:])
	queue.items = queue.items[:len(queue.items)-1]
	delete(queue.present, item.key)
	return item.key, item.level, true
}

func (queue *CuddLevelQueue) Len() int {
	if queue == nil {
		return 0
	}
	return len(queue.items)
}
