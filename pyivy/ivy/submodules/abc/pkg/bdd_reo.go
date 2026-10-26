package abc

// Standalone REO behavior from src/bdd/reo/reo.h, reoApi.c, reoCore.c,
// reoProfile.c, reoShuffle.c, reoSift.c, reoSwap.c, reoTransfer.c,
// reoUnits.c, reoTest.c, and module.make.

import (
	"math"
	"sort"
	"strings"
)

type REOMinimization uint8

const (
	REOMinimizeNodes REOMinimization = iota
	REOMinimizeWidth
	REOMinimizeAPL
)

type REOOptions struct {
	MaxEvaluations int
	MaxOperations  int
	MaxBytes       int
	Stop           func() bool
}

type REOProfile struct {
	Nodes, Width int
	APL          float64
	NodesByLevel []int
	WidthByLevel []int
	APLByLevel   []float64
}

type REOResult struct {
	Order, Permutation  []int
	DAG                 CuddOrderedDAG
	Initial, Final      REOProfile
	Evaluations, Passes int
	AdjacentSwaps       int
	NoninteractingSwaps int
}

type REOCoverage struct {
	Build, Complete, Partial, NativeOnly []string
}

func REOSourceCoverage() REOCoverage {
	return REOCoverage{
		Build:      []string{"reoApi.c", "reoCore.c", "reoProfile.c", "reoShuffle.c", "reoSift.c", "reoSwap.c", "reoTransfer.c", "reoUnits.c"},
		Complete:   []string{"owned unit-DAG transfer and validation", "node/width/APL profiles", "deterministic order verification"},
		Partial:    []string{"Extra_Reorder/Extra_ReorderArray", "reoReorderArray/reoReorderSift/reoReorderSwapAdjacentVars", "reoShuffle", "profile print entry points", "REO test entry points", "remap-up reconstruction is rejected"},
		NativeOnly: []string{"CUDD pointer tags and reference counts", "unit chunk allocator/free-list layout", "native hash signatures", "process output and timing"},
	}
}

func REODefinitionInventory() []string {
	return []string{
		"reoUnitsGetNextUnit", "reoUnitsRecycleUnit", "reoUnitsRecycleUnitList", "reoUnitsStopDispenser", "reoUnitsAddUnitToPlane", "reoUnitsAddToFreeUnitList",
		"reoReorderSwapAdjacentVars", "reoReorderSift", "reoShuffle", "Extra_ShuffleTest",
		"reoReorderArray", "reoResizeStructures", "reoRecursiveDeref", "reoCheckZeroRefs", "reoCheckLevels",
		"reoTransferNodesToUnits_rec", "reoTransferUnitsToNodes_rec",
		"reoProfileNodesStart", "reoProfileAplStart", "reoProfileWidthStart", "reoProfileWidthStart2", "reoProfileNodesPrint", "reoProfileAplPrint", "reoProfileWidthPrint", "reoProfileWidthVerifyLevel",
		"Extra_ReorderInit", "Extra_ReorderQuit", "Extra_ReorderSetMinimizationType", "Extra_ReorderSetRemapping", "Extra_ReorderSetIterations", "Extra_ReorderSetVerification", "Extra_ReorderSetVerbosity", "Extra_Reorder", "Extra_ReorderArray",
		"Extra_ReorderTest", "Extra_ReorderTestArray", "Extra_ReorderCudd", "Extra_bddReorderTest", "Extra_addReorderTest",
	}
}

type REOManager struct {
	maxVariables, maxNodes int
	minimization           REOMinimization
	remapUp, verify        bool
	iterations             int
	closed                 bool
}

func NewREOManager(maxVariables, maxNodes int) (*REOManager, error) {
	if maxVariables < 0 || maxVariables > 20 || maxNodes < 0 {
		return nil, ErrDimension
	}
	return &REOManager{maxVariables: maxVariables, maxNodes: maxNodes, iterations: 1}, nil
}

func (manager *REOManager) check() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed {
		return ErrInvariant
	}
	if manager.iterations < 1 || manager.minimization > REOMinimizeAPL {
		return ErrInvariant
	}
	return nil
}

func (manager *REOManager) SetMinimization(mode REOMinimization) error {
	if err := manager.check(); err != nil {
		return err
	}
	if mode > REOMinimizeAPL {
		return ErrDimension
	}
	manager.minimization = mode
	return nil
}

func (manager *REOManager) SetRemapping(remap bool) error {
	if err := manager.check(); err != nil {
		return err
	}
	manager.remapUp = remap
	return nil
}

func (manager *REOManager) SetIterations(iterations int) error {
	if err := manager.check(); err != nil {
		return err
	}
	if iterations < 1 || iterations > 1<<16 {
		return ErrDimension
	}
	manager.iterations = iterations
	return nil
}

func (manager *REOManager) SetVerification(verify bool) error {
	if err := manager.check(); err != nil {
		return err
	}
	manager.verify = verify
	return nil
}

func (manager *REOManager) Close() error {
	if manager == nil || manager.closed {
		return ErrInvariant
	}
	manager.closed = true
	return nil
}

func normalizeREOOptions(options REOOptions) (REOOptions, error) {
	if options.MaxEvaluations < 0 || options.MaxOperations < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxEvaluations == 0 {
		options.MaxEvaluations = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 28
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 256 << 20
	}
	return options, nil
}

type reoWork struct {
	problem     *CuddReorderProblem
	options     REOOptions
	maxNodes    int
	evaluations int
	operations  int
}

func (work *reoWork) poll(extra int) error {
	if extra < 0 || work.operations > work.options.MaxOperations-extra {
		return ErrLimit
	}
	work.operations += extra
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (work *reoWork) evaluate(order []int) (CuddOrderedDAG, REOProfile, error) {
	if work.evaluations >= work.options.MaxEvaluations {
		return CuddOrderedDAG{}, REOProfile{}, ErrLimit
	}
	cells := len(work.problem.tables) * (1 << work.problem.variables)
	if err := work.poll(cells * 2); err != nil {
		return CuddOrderedDAG{}, REOProfile{}, err
	}
	work.evaluations++
	dag, err := work.problem.canonical(order, work.options.Stop)
	if err != nil {
		return CuddOrderedDAG{}, REOProfile{}, err
	}
	// Account for the canonical result and the top/weight/unique state used to
	// profile it.  The latter used to fall outside MaxBytes even though it can
	// be substantially larger than the dense input for shared multi-output
	// diagrams.
	profileBytes := int64(len(dag.Nodes))*64 + int64(len(dag.Terminals))*48 + int64(len(dag.Order)+1)*24
	if int64(len(dag.Nodes))*32+int64(len(dag.Terminals))*8+int64(cells)*8+profileBytes > int64(work.options.MaxBytes) {
		return CuddOrderedDAG{}, REOProfile{}, ErrLimit
	}
	// Width profiling may visit every intervening level for every node. Charge
	// an upper bound before allocating its scratch state so one evaluation
	// cannot silently escape the aggregate operation budget.
	profileOperations := len(dag.Roots) + (len(dag.Nodes)+len(dag.Terminals))*(len(dag.Order)+4)
	if err := work.poll(profileOperations); err != nil {
		return CuddOrderedDAG{}, REOProfile{}, err
	}
	profile, err := reoProfileWithStop(dag, work.options.Stop)
	if err == nil && work.maxNodes != 0 && profile.Nodes > work.maxNodes {
		return CuddOrderedDAG{}, REOProfile{}, ErrLimit
	}
	return dag, profile, err
}

func (work *reoWork) support() ([]bool, error) {
	support := make([]bool, work.problem.variables)
	for variable := range support {
		mask := 1 << variable
		for _, table := range work.problem.tables {
			for assignment := 0; assignment < len(table); assignment++ {
				if assignment&255 == 0 {
					if err := work.poll(min(256, len(table)-assignment)); err != nil {
						return nil, err
					}
				}
				if assignment&mask == 0 && table[assignment] != table[assignment|mask] {
					support[variable] = true
					break
				}
			}
			if support[variable] {
				break
			}
		}
	}
	return support, nil
}

func (work *reoWork) interactions() (CuddInteractionMatrix, error) {
	matrix := CuddInteractionMatrix{variables: work.problem.variables, bits: make([]uint64, (work.problem.variables*work.problem.variables+63)/64)}
	for _, table := range work.problem.tables {
		support := make([]bool, work.problem.variables)
		for variable := range support {
			mask := 1 << variable
			for assignment := 0; assignment < len(table); assignment++ {
				if assignment&255 == 0 {
					if err := work.poll(min(256, len(table)-assignment)); err != nil {
						return CuddInteractionMatrix{}, err
					}
				}
				if assignment&mask == 0 && table[assignment] != table[assignment|mask] {
					support[variable] = true
					break
				}
			}
		}
		for first := 0; first < work.problem.variables; first++ {
			if !support[first] {
				continue
			}
			for second := first + 1; second < work.problem.variables; second++ {
				if support[second] {
					matrix.set(first, second)
				}
			}
		}
	}
	return matrix, nil
}

func reoMetric(profile REOProfile, mode REOMinimization) float64 {
	switch mode {
	case REOMinimizeWidth:
		return float64(profile.Width)
	case REOMinimizeAPL:
		return profile.APL
	default:
		return float64(profile.Nodes)
	}
}

func reoBetter(candidate REOProfile, candidateOrder []int, best REOProfile, bestOrder []int, mode REOMinimization) bool {
	left, right := reoMetric(candidate, mode), reoMetric(best, mode)
	if left < right-1e-12 {
		return true
	}
	if math.Abs(left-right) > 1e-12 {
		return false
	}
	for i := range candidateOrder {
		if candidateOrder[i] != bestOrder[i] {
			return candidateOrder[i] < bestOrder[i]
		}
	}
	return false
}

func (manager *REOManager) Reorder(problem *CuddReorderProblem, initial []int, options REOOptions) (REOResult, error) {
	if err := manager.check(); err != nil {
		return REOResult{}, err
	}
	if problem == nil || !problem.valid() || problem.variables > manager.maxVariables && manager.maxVariables != 0 {
		return REOResult{}, ErrDimension
	}
	if manager.remapUp {
		return REOResult{}, ErrUnsupported
	}
	if initial == nil {
		initial = make([]int, problem.variables)
		for i := range initial {
			initial[i] = i
		}
	}
	if !cuddValidOrder(initial, problem.variables) {
		return REOResult{}, ErrDimension
	}
	var err error
	options, err = normalizeREOOptions(options)
	if err != nil {
		return REOResult{}, err
	}
	cells := int64(len(problem.tables)) * int64(1<<problem.variables)
	if cells*8 > int64(options.MaxBytes) {
		return REOResult{}, ErrLimit
	}
	work := reoWork{problem: problem, options: options, maxNodes: manager.maxNodes}
	initialDAG, initialProfile, err := work.evaluate(initial)
	if err != nil {
		return REOResult{}, err
	}
	bestOrder := append([]int(nil), initial...)
	bestDAG, bestProfile := initialDAG, initialProfile
	support, err := work.support()
	if err != nil {
		return REOResult{}, err
	}
	interactions, err := work.interactions()
	if err != nil {
		return REOResult{}, err
	}
	swaps, noninteracting, passes := 0, 0, 0
	for pass := 0; pass < manager.iterations; pass++ {
		before := reoMetric(bestProfile, manager.minimization)
		visit := make([]int, 0, problem.variables)
		for _, variable := range bestOrder {
			if support[variable] {
				visit = append(visit, variable)
			}
		}
		sort.SliceStable(visit, func(i, j int) bool {
			li, lj := reoLevelNodes(bestDAG, visit[i]), reoLevelNodes(bestDAG, visit[j])
			if li != lj {
				return li > lj
			}
			return visit[i] < visit[j]
		})
		for _, variable := range visit {
			base := append([]int(nil), bestOrder...)
			localOrder := append([]int(nil), bestOrder...)
			localDAG, localProfile := bestDAG, bestProfile
			for to := range base {
				if !support[base[to]] {
					continue
				}
				candidate, distance, crossed := reoMoveSupported(base, support, variable, to, interactions)
				dag, profile, evalErr := work.evaluate(candidate)
				if evalErr != nil {
					return REOResult{}, evalErr
				}
				swaps += distance
				noninteracting += crossed
				if reoBetter(profile, candidate, localProfile, localOrder, manager.minimization) {
					localOrder, localDAG, localProfile = candidate, dag, profile
				}
			}
			bestOrder, bestDAG, bestProfile = localOrder, localDAG, localProfile
		}
		passes++
		if reoMetric(bestProfile, manager.minimization) >= before-1e-12 {
			break
		}
	}
	permutation := make([]int, problem.variables)
	for level, variable := range bestOrder {
		permutation[variable] = level
	}
	if manager.verify {
		for root := range problem.tables {
			for assignment, expected := range problem.tables[root] {
				if assignment&255 == 0 {
					if err = work.poll(min(256, len(problem.tables[root])-assignment)); err != nil {
						return REOResult{}, err
					}
				}
				value, valueErr := bestDAG.Value(root, uint64(assignment))
				if valueErr != nil || value != expected {
					return REOResult{}, ErrInvariant
				}
			}
		}
	}
	return REOResult{Order: append([]int(nil), bestOrder...), Permutation: permutation, DAG: cloneCuddOrderedDAG(bestDAG), Initial: cloneREOProfile(initialProfile), Final: cloneREOProfile(bestProfile), Evaluations: work.evaluations, Passes: passes, AdjacentSwaps: swaps, NoninteractingSwaps: noninteracting}, nil
}

func reoMove(order []int, from, to int) []int {
	result := append([]int(nil), order...)
	value := result[from]
	if from < to {
		copy(result[from:to], result[from+1:to+1])
	} else if from > to {
		copy(result[to+1:from+1], result[to:from])
	}
	result[to] = value
	return result
}

func reoMoveSupported(order []int, support []bool, variable, targetPosition int, interactions CuddInteractionMatrix) ([]int, int, int) {
	result := append([]int(nil), order...)
	slots := make([]int, 0, len(order))
	values := make([]int, 0, len(order))
	fromRank, toRank := -1, -1
	for position, value := range order {
		if support[value] {
			if value == variable {
				fromRank = len(values)
			}
			if position == targetPosition {
				toRank = len(values)
			}
			slots = append(slots, position)
			values = append(values, value)
		}
	}
	crossed := 0
	if fromRank < toRank {
		for _, other := range values[fromRank+1 : toRank+1] {
			if !interactions.Interacts(variable, other) {
				crossed++
			}
		}
	} else if fromRank > toRank {
		for _, other := range values[toRank:fromRank] {
			if !interactions.Interacts(variable, other) {
				crossed++
			}
		}
	}
	values = reoMove(values, fromRank, toRank)
	for i, slot := range slots {
		result[slot] = values[i]
	}
	distance := fromRank - toRank
	if distance < 0 {
		distance = -distance
	}
	return result, distance, crossed
}

func reoLevelNodes(dag CuddOrderedDAG, variable int) int {
	count := 0
	for _, node := range dag.Nodes {
		if node.Variable == variable {
			count++
		}
	}
	return count
}

func cloneCuddOrderedDAG(dag CuddOrderedDAG) CuddOrderedDAG {
	return CuddOrderedDAG{Order: append([]int(nil), dag.Order...), Roots: append([]int(nil), dag.Roots...), Terminals: append([]uint64(nil), dag.Terminals...), Nodes: append([]CuddOrderedNode(nil), dag.Nodes...)}
}

func cloneREOProfile(profile REOProfile) REOProfile {
	profile.NodesByLevel = append([]int(nil), profile.NodesByLevel...)
	profile.WidthByLevel = append([]int(nil), profile.WidthByLevel...)
	profile.APLByLevel = append([]float64(nil), profile.APLByLevel...)
	return profile
}

func reoProfile(dag CuddOrderedDAG) (REOProfile, error) {
	return reoProfileWithStop(dag, nil)
}

func reoProfileWithStop(dag CuddOrderedDAG, stop func() bool) (REOProfile, error) {
	if !cuddValidOrder(dag.Order, len(dag.Order)) || len(dag.Roots) == 0 || len(dag.Terminals) == 0 {
		return REOProfile{}, ErrInvariant
	}
	active := make([]bool, len(dag.Order))
	for _, node := range dag.Nodes {
		if node.Variable < 0 || node.Variable >= len(active) {
			return REOProfile{}, ErrInvariant
		}
		active[node.Variable] = true
	}
	levels := make([]int, len(dag.Order))
	supportLevels := 0
	for _, variable := range dag.Order {
		if active[variable] {
			levels[variable] = supportLevels
			supportLevels++
		}
	}
	profile := REOProfile{NodesByLevel: make([]int, supportLevels+1), WidthByLevel: make([]int, supportLevels+1), APLByLevel: make([]float64, supportLevels+1)}
	ticks := 0
	poll := func(extra int) error {
		ticks += extra
		if ticks >= 1024 {
			ticks %= 1024
			if stop != nil && stop() {
				return ErrCancelled
			}
		}
		return nil
	}
	top := make([]int, len(dag.Nodes)+len(dag.Terminals))
	for i := range top {
		top[i] = supportLevels + 1
	}
	weights := make([]float64, len(dag.Nodes))
	unique := make(map[cuddReorderNode]struct{}, len(dag.Nodes))
	terminalValues := make(map[uint64]struct{}, len(dag.Terminals))
	for _, value := range dag.Terminals {
		if _, found := terminalValues[value]; found {
			return REOProfile{}, ErrInvariant
		}
		terminalValues[value] = struct{}{}
	}
	for _, edge := range dag.Roots {
		if err := poll(1); err != nil {
			return REOProfile{}, err
		}
		index, terminal, err := reoEdgeIndex(edge, len(dag.Nodes), len(dag.Terminals))
		if err != nil {
			return REOProfile{}, err
		}
		top[index] = 0
		if !terminal {
			weights[index]++
		}
	}
	for id := len(dag.Nodes); id >= 1; id-- {
		if err := poll(1); err != nil {
			return REOProfile{}, err
		}
		node := dag.Nodes[id-1]
		level := levels[node.Variable]
		if node.Low == node.High {
			return REOProfile{}, ErrInvariant
		}
		key := cuddReorderNode{variable: node.Variable, low: node.Low, high: node.High}
		if _, found := unique[key]; found {
			return REOProfile{}, ErrInvariant
		}
		unique[key] = struct{}{}
		profile.NodesByLevel[level]++
		profile.APLByLevel[level] += weights[id-1]
		profile.APL += weights[id-1]
		for _, child := range []int{node.Low, node.High} {
			index, terminal, err := reoEdgeIndex(child, len(dag.Nodes), len(dag.Terminals))
			if err != nil || !terminal && index >= id-1 {
				return REOProfile{}, ErrInvariant
			}
			if !terminal && levels[dag.Nodes[index].Variable] <= level {
				return REOProfile{}, ErrInvariant
			}
			if top[index] > level+1 {
				top[index] = level + 1
			}
			if !terminal {
				weights[index] += weights[id-1] * 0.5
			}
		}
	}
	for id, node := range dag.Nodes {
		if err := poll(1); err != nil {
			return REOProfile{}, err
		}
		if top[id] > supportLevels {
			return REOProfile{}, ErrInvariant
		}
		last := levels[node.Variable]
		for level := top[id]; level <= last; level++ {
			if err := poll(1); err != nil {
				return REOProfile{}, err
			}
			profile.WidthByLevel[level]++
		}
	}
	for terminal := range dag.Terminals {
		if err := poll(1); err != nil {
			return REOProfile{}, err
		}
		index := len(dag.Nodes) + terminal
		if top[index] > supportLevels {
			return REOProfile{}, ErrInvariant
		}
		for level := top[index]; level <= supportLevels; level++ {
			if err := poll(1); err != nil {
				return REOProfile{}, err
			}
			profile.WidthByLevel[level]++
		}
		profile.NodesByLevel[supportLevels]++
	}
	// REO's nNodesCur and node profile include the reachable constant units in
	// the final plane.  Keep Nodes consistent with NodesByLevel and the source
	// objective rather than reporting internal nodes only.
	for _, count := range profile.NodesByLevel {
		profile.Nodes += count
	}
	for _, width := range profile.WidthByLevel {
		profile.Width += width
	}
	return profile, nil
}

func reoEdgeIndex(edge, nodes, terminals int) (int, bool, error) {
	if edge > 0 {
		if edge > nodes {
			return 0, false, ErrInvariant
		}
		return edge - 1, false, nil
	}
	terminal := -edge - 1
	if terminal < 0 || terminal >= terminals {
		return 0, true, ErrInvariant
	}
	return nodes + terminal, true, nil
}

func (manager *REOManager) Shuffle(problem *CuddReorderProblem, order []int, options REOOptions) (CuddOrderedDAG, REOProfile, error) {
	if err := manager.check(); err != nil {
		return CuddOrderedDAG{}, REOProfile{}, err
	}
	if problem == nil || !problem.valid() || !cuddValidOrder(order, problem.variables) || problem.variables > manager.maxVariables && manager.maxVariables != 0 {
		return CuddOrderedDAG{}, REOProfile{}, ErrDimension
	}
	options, err := normalizeREOOptions(options)
	if err != nil {
		return CuddOrderedDAG{}, REOProfile{}, err
	}
	work := reoWork{problem: problem, options: options, maxNodes: manager.maxNodes}
	dag, profile, err := work.evaluate(order)
	return dag, profile, err
}

func (manager *REOManager) SwapAdjacent(problem *CuddReorderProblem, order []int, level int, options REOOptions) ([]int, REOProfile, error) {
	if level < 0 || level+1 >= len(order) {
		return nil, REOProfile{}, ErrDimension
	}
	candidate := append([]int(nil), order...)
	candidate[level], candidate[level+1] = candidate[level+1], candidate[level]
	_, profile, err := manager.Shuffle(problem, candidate, options)
	if err != nil {
		return nil, REOProfile{}, err
	}
	return candidate, profile, nil
}

type REOUnitNode struct {
	Variable  int
	Low, High int
}

type REOUnitDAG struct {
	Order, Roots []int
	Terminals    []uint64
	Units        []REOUnitNode
}

func REOTransferToUnits(dag CuddOrderedDAG) (REOUnitDAG, error) {
	if len(dag.Order) > 20 || len(dag.Roots) > 1<<16 || len(dag.Nodes) > cuddReorderMaxDenseCells || len(dag.Terminals) > cuddReorderMaxDenseCells {
		return REOUnitDAG{}, ErrLimit
	}
	if _, err := reoProfile(dag); err != nil {
		return REOUnitDAG{}, err
	}
	result := REOUnitDAG{Order: append([]int(nil), dag.Order...), Roots: append([]int(nil), dag.Roots...), Terminals: append([]uint64(nil), dag.Terminals...), Units: make([]REOUnitNode, len(dag.Nodes))}
	for i, node := range dag.Nodes {
		result.Units[i] = REOUnitNode{Variable: node.Variable, Low: node.Low, High: node.High}
	}
	return result, nil
}

func (units REOUnitDAG) TransferToDAG() (CuddOrderedDAG, error) {
	if len(units.Order) > 20 || len(units.Roots) > 1<<16 || len(units.Units) > cuddReorderMaxDenseCells || len(units.Terminals) > cuddReorderMaxDenseCells {
		return CuddOrderedDAG{}, ErrLimit
	}
	dag := CuddOrderedDAG{Order: append([]int(nil), units.Order...), Roots: append([]int(nil), units.Roots...), Terminals: append([]uint64(nil), units.Terminals...), Nodes: make([]CuddOrderedNode, len(units.Units))}
	for i, unit := range units.Units {
		dag.Nodes[i] = CuddOrderedNode{Variable: unit.Variable, Low: unit.Low, High: unit.High}
	}
	if _, err := reoProfile(dag); err != nil {
		return CuddOrderedDAG{}, err
	}
	return dag, nil
}

func (profile REOProfile) Text() string {
	var builder strings.Builder
	for level := range profile.NodesByLevel {
		builder.WriteString("level ")
		builder.WriteString(itoaSmall(level))
		builder.WriteString(" nodes ")
		builder.WriteString(itoaSmall(profile.NodesByLevel[level]))
		builder.WriteString(" width ")
		builder.WriteString(itoaSmall(profile.WidthByLevel[level]))
		builder.WriteByte('\n')
	}
	return builder.String()
}

func itoaSmall(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var buffer [24]byte
	position := len(buffer)
	for value != 0 {
		position--
		buffer[position] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		position--
		buffer[position] = '-'
	}
	return string(buffer[position:])
}
