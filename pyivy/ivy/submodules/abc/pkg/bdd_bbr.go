package abc

// Stand-alone bounded semantics from src/bdd/bbr/bbr.h, bbrCex.c,
// bbrImage.c, bbrNtbdd.c, bbrReach.c, and module.make.

import (
	"math"
	"slices"
)

// BBRBuildSources is the exact active source inventory from module.make.
var BBRBuildSources = [...]string{
	"src/bdd/bbr/bbrCex.c",
	"src/bdd/bbr/bbrImage.c",
	"src/bdd/bbr/bbrNtbdd.c",
	"src/bdd/bbr/bbrReach.c",
}

// BBRPartition is one next-state function. Values are indexed by
// (presentState << InputBits) | input.
type BBRPartition struct {
	NextBit int
	Values  []bool
}

// BBROutput is one bad-state predicate over present state and input.
type BBROutput struct {
	Values []bool
}

// BBRTransitionSystem owns an explicit, deterministic transition relation.
// Partitioning mirrors the one-function-per-latch representation used by BBR.
type BBRTransitionSystem struct {
	StateBits  int
	InputBits  int
	Initial    []uint64
	Partitions []BBRPartition
	Outputs    []BBROutput
}

type BBROptions struct {
	MaxOperations int
	MaxStates     int
	MaxBytes      int
	MaxIterations int
	Stop          func() bool
}

type BBRStatistics struct {
	Operations int
	States     int
	Iterations int
}

type BBRCounterexample struct {
	States []uint64
	Inputs []uint64
	Output int
	Frame  int
}

type BBRReachability struct {
	Reached        []uint64
	Rings          [][]uint64
	Counterexample *BBRCounterexample
	Complete       bool
	Statistics     BBRStatistics
}

type bbrBudget struct {
	options    BBROptions
	operations int
	bytes      int
}

func normalizeBBROptions(options BBROptions) (BBROptions, error) {
	if options.MaxOperations < 0 || options.MaxStates < 0 || options.MaxBytes < 0 || options.MaxIterations < 0 {
		return options, ErrDimension
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 26
	}
	if options.MaxStates == 0 {
		options.MaxStates = 1 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 1000
	}
	return options, nil
}

func (budget *bbrBudget) step(count int) error {
	if count < 0 || budget.operations > budget.options.MaxOperations-count {
		return ErrLimit
	}
	budget.operations += count
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (budget *bbrBudget) reserve(bytes int) error {
	if bytes < 0 || budget.bytes > budget.options.MaxBytes-bytes {
		return ErrLimit
	}
	budget.bytes += bytes
	return nil
}

func bbrSpace(bits int) (int, error) {
	if bits < 0 || bits > 20 {
		return 0, ErrDimension
	}
	return 1 << uint(bits), nil
}

func canonicalBBRStates(states []uint64, stateCount, maximum int) ([]uint64, error) {
	if len(states) > maximum {
		return nil, ErrLimit
	}
	result := slices.Clone(states)
	slices.Sort(result)
	result = slices.Compact(result)
	for _, state := range result {
		if state >= uint64(stateCount) {
			return nil, ErrDimension
		}
	}
	return result, nil
}

func bbrCheckedAdd(first, second int) (int, error) {
	if first < 0 || second < 0 || first > math.MaxInt-second {
		return 0, ErrLimit
	}
	return first + second, nil
}

func bbrCloneBools(values []bool, budget *bbrBudget) ([]bool, error) {
	result := make([]bool, len(values))
	const block = 4096
	for start := 0; start < len(values); start += block {
		stop := min(start+block, len(values))
		if err := budget.step(stop - start); err != nil {
			return nil, err
		}
		copy(result[start:stop], values[start:stop])
	}
	return result, nil
}

func (system BBRTransitionSystem) cloneAndValidate(options BBROptions) (BBRTransitionSystem, error) {
	stateCount, err := bbrSpace(system.StateBits)
	if err != nil {
		return BBRTransitionSystem{}, err
	}
	inputCount, err := bbrSpace(system.InputBits)
	if err != nil || system.StateBits+system.InputBits > 20 {
		return BBRTransitionSystem{}, ErrDimension
	}
	if stateCount > options.MaxStates || stateCount > math.MaxInt/inputCount {
		return BBRTransitionSystem{}, ErrLimit
	}
	rows := stateCount * inputCount
	if len(system.Partitions) != system.StateBits {
		return BBRTransitionSystem{}, ErrDimension
	}
	tables := len(system.Partitions) + len(system.Outputs)
	if tables > math.MaxInt/24 || (tables != 0 && rows > math.MaxInt/tables) {
		return BBRTransitionSystem{}, ErrLimit
	}
	bytes, err := bbrCheckedAdd(rows*tables, tables*24)
	if err != nil {
		return BBRTransitionSystem{}, err
	}
	bytes, err = bbrCheckedAdd(bytes, system.StateBits)
	if err != nil {
		return BBRTransitionSystem{}, err
	}
	if len(system.Initial) > math.MaxInt/8 {
		return BBRTransitionSystem{}, ErrLimit
	}
	bytes, err = bbrCheckedAdd(bytes, len(system.Initial)*8)
	if err != nil {
		return BBRTransitionSystem{}, err
	}
	if bytes > options.MaxBytes {
		return BBRTransitionSystem{}, ErrLimit
	}
	budget := bbrBudget{options: options}
	if err := budget.reserve(bytes); err != nil {
		return BBRTransitionSystem{}, err
	}
	if err := budget.step(len(system.Initial)); err != nil {
		return BBRTransitionSystem{}, err
	}
	seen := make([]bool, system.StateBits)
	copySystem := BBRTransitionSystem{StateBits: system.StateBits, InputBits: system.InputBits}
	copySystem.Initial, err = canonicalBBRStates(system.Initial, stateCount, options.MaxStates)
	if err != nil || len(copySystem.Initial) == 0 {
		if err != nil {
			return BBRTransitionSystem{}, err
		}
		return BBRTransitionSystem{}, ErrDimension
	}
	copySystem.Partitions = make([]BBRPartition, len(system.Partitions))
	for index, partition := range system.Partitions {
		if partition.NextBit < 0 || partition.NextBit >= system.StateBits || seen[partition.NextBit] || len(partition.Values) != rows {
			return BBRTransitionSystem{}, ErrDimension
		}
		seen[partition.NextBit] = true
		values, cloneErr := bbrCloneBools(partition.Values, &budget)
		if cloneErr != nil {
			return BBRTransitionSystem{}, cloneErr
		}
		copySystem.Partitions[index] = BBRPartition{NextBit: partition.NextBit, Values: values}
	}
	copySystem.Outputs = make([]BBROutput, len(system.Outputs))
	for index, output := range system.Outputs {
		if len(output.Values) != rows {
			return BBRTransitionSystem{}, ErrDimension
		}
		values, cloneErr := bbrCloneBools(output.Values, &budget)
		if cloneErr != nil {
			return BBRTransitionSystem{}, cloneErr
		}
		copySystem.Outputs[index].Values = values
	}
	return copySystem, nil
}

// DefaultBBRParameters preserves the public defaults set by Bbr_ManSetDefaultParams.
func DefaultBBRParameters() SaigBDDParameters {
	return SaigBDDParameters{
		MaximumBDDNodes: 50000, MaximumIterations: 1000,
		Partition: true, Reorder: true, ReorderImage: true,
		ExploredFrame: -1,
	}
}

type BBRImageTree struct {
	system BBRTransitionSystem
	last   []uint64
	closed bool
}

func NewBBRImageTree(system BBRTransitionSystem, options BBROptions) (*BBRImageTree, error) {
	normalized, err := normalizeBBROptions(options)
	if err != nil {
		return nil, err
	}
	owned, err := system.cloneAndValidate(normalized)
	if err != nil {
		return nil, err
	}
	return &BBRImageTree{system: owned}, nil
}

func (tree *BBRImageTree) check() error {
	if tree == nil {
		return ErrDimension
	}
	if tree.closed {
		return ErrInvariant
	}
	return nil
}

func (tree *BBRImageTree) next(state, input uint64, budget *bbrBudget) (uint64, error) {
	if err := budget.step(1 + len(tree.system.Partitions)); err != nil {
		return 0, err
	}
	row := int((state << uint(tree.system.InputBits)) | input)
	var next uint64
	for _, partition := range tree.system.Partitions {
		if partition.Values[row] {
			next |= uint64(1) << uint(partition.NextBit)
		}
	}
	return next, nil
}

func (tree *BBRImageTree) Image(care []uint64, options BBROptions) ([]uint64, error) {
	if err := tree.check(); err != nil {
		return nil, err
	}
	normalized, err := normalizeBBROptions(options)
	if err != nil {
		return nil, err
	}
	stateCount := 1 << uint(tree.system.StateBits)
	if stateCount > normalized.MaxStates {
		return nil, ErrLimit
	}
	states, err := canonicalBBRStates(care, stateCount, normalized.MaxStates)
	if err != nil {
		return nil, err
	}
	inputCount := 1 << uint(tree.system.InputBits)
	budget := bbrBudget{options: normalized}
	// Marker, canonical input, and worst-case owned result.
	if err := budget.reserve(stateCount*17 + len(states)*8); err != nil {
		return nil, err
	}
	marked := make([]bool, stateCount)
	for _, state := range states {
		for input := 0; input < inputCount; input++ {
			next, nextErr := tree.next(state, uint64(input), &budget)
			if nextErr != nil {
				return nil, nextErr
			}
			marked[next] = true
		}
	}
	result := make([]uint64, 0)
	for state, present := range marked {
		if present {
			if len(result) == normalized.MaxStates {
				return nil, ErrLimit
			}
			result = append(result, uint64(state))
		}
	}
	tree.last = slices.Clone(result)
	return result, nil
}

func (tree *BBRImageTree) Preimage(target []uint64, options BBROptions) ([]uint64, error) {
	if err := tree.check(); err != nil {
		return nil, err
	}
	normalized, err := normalizeBBROptions(options)
	if err != nil {
		return nil, err
	}
	stateCount := 1 << uint(tree.system.StateBits)
	if stateCount > normalized.MaxStates {
		return nil, ErrLimit
	}
	targets, err := canonicalBBRStates(target, stateCount, normalized.MaxStates)
	if err != nil {
		return nil, err
	}
	budget := bbrBudget{options: normalized}
	// Target marker, canonical target, and worst-case owned result.
	if err := budget.reserve(stateCount*9 + len(targets)*8); err != nil {
		return nil, err
	}
	wanted := make([]bool, stateCount)
	for _, state := range targets {
		wanted[state] = true
	}
	inputCount := 1 << uint(tree.system.InputBits)
	result := make([]uint64, 0)
	for state := 0; state < stateCount; state++ {
		for input := 0; input < inputCount; input++ {
			next, nextErr := tree.next(uint64(state), uint64(input), &budget)
			if nextErr != nil {
				return nil, nextErr
			}
			if wanted[next] {
				result = append(result, uint64(state))
				break
			}
		}
	}
	tree.last = slices.Clone(result)
	return result, nil
}

func (tree *BBRImageTree) Read() ([]uint64, error) {
	if err := tree.check(); err != nil {
		return nil, err
	}
	return slices.Clone(tree.last), nil
}

func (tree *BBRImageTree) Close() error {
	if err := tree.check(); err != nil {
		return err
	}
	tree.closed = true
	tree.last = nil
	tree.system = BBRTransitionSystem{}
	return nil
}

func (tree *BBRImageTree) bad(state uint64, budget *bbrBudget) (int, uint64, bool, error) {
	inputCount := 1 << uint(tree.system.InputBits)
	for input := 0; input < inputCount; input++ {
		if err := budget.step(1 + len(tree.system.Outputs)); err != nil {
			return 0, 0, false, err
		}
		row := int((state << uint(tree.system.InputBits)) | uint64(input))
		for output := range tree.system.Outputs {
			if tree.system.Outputs[output].Values[row] {
				return output, uint64(input), true, nil
			}
		}
	}
	return 0, 0, false, nil
}

func bbrTrace(state uint64, input uint64, output int, predecessor []uint64, predecessorInput []uint64, depth []int) *BBRCounterexample {
	frame := depth[state]
	states := make([]uint64, frame+1)
	inputs := make([]uint64, frame+1)
	states[frame], inputs[frame] = state, input
	for index := frame; index > 0; index-- {
		inputs[index-1] = predecessorInput[states[index]]
		states[index-1] = predecessor[states[index]]
	}
	return &BBRCounterexample{States: states, Inputs: inputs, Output: output, Frame: frame}
}

// VerifyCounterexample replays an owned trace against the transition partitions
// and selected output. It replaces the post-construction SAIG replay check.
func (tree *BBRImageTree) VerifyCounterexample(counterexample *BBRCounterexample, options BBROptions) (bool, error) {
	if err := tree.check(); err != nil {
		return false, err
	}
	if counterexample == nil || counterexample.Frame < 0 || counterexample.Frame+1 != len(counterexample.States) || len(counterexample.Inputs) != len(counterexample.States) || counterexample.Output < 0 || counterexample.Output >= len(tree.system.Outputs) {
		return false, ErrDimension
	}
	normalized, err := normalizeBBROptions(options)
	if err != nil {
		return false, err
	}
	stateCount := uint64(1) << uint(tree.system.StateBits)
	inputCount := uint64(1) << uint(tree.system.InputBits)
	if !slices.Contains(tree.system.Initial, counterexample.States[0]) {
		return false, nil
	}
	budget := bbrBudget{options: normalized}
	for frame, state := range counterexample.States {
		if state >= stateCount || counterexample.Inputs[frame] >= inputCount {
			return false, ErrDimension
		}
		if frame+1 < len(counterexample.States) {
			next, nextErr := tree.next(state, counterexample.Inputs[frame], &budget)
			if nextErr != nil {
				return false, nextErr
			}
			if next != counterexample.States[frame+1] {
				return false, nil
			}
		}
	}
	if err := budget.step(1); err != nil {
		return false, err
	}
	last := len(counterexample.States) - 1
	row := int((counterexample.States[last] << uint(tree.system.InputBits)) | counterexample.Inputs[last])
	return tree.system.Outputs[counterexample.Output].Values[row], nil
}

// Reachability computes onion rings and a deterministic shortest counterexample.
func (tree *BBRImageTree) Reachability(checkOutputs bool, options BBROptions) (BBRReachability, error) {
	if err := tree.check(); err != nil {
		return BBRReachability{}, err
	}
	normalized, err := normalizeBBROptions(options)
	if err != nil {
		return BBRReachability{}, err
	}
	stateCount := 1 << uint(tree.system.StateBits)
	if stateCount > normalized.MaxStates {
		return BBRReachability{}, ErrLimit
	}
	budget := bbrBudget{options: normalized}
	// Visited state, predecessor/input/depth tables, frontiers, rings, and result.
	if err := budget.reserve(stateCount * 64); err != nil {
		return BBRReachability{}, err
	}
	visited := make([]bool, stateCount)
	predecessor := make([]uint64, stateCount)
	predecessorInput := make([]uint64, stateCount)
	depth := make([]int, stateCount)
	frontier := slices.Clone(tree.system.Initial)
	discovered := len(frontier)
	for _, state := range frontier {
		visited[state] = true
	}
	rings := make([][]uint64, 0)
	inputCount := 1 << uint(tree.system.InputBits)
	iterations := 0
	for len(frontier) != 0 && iterations < normalized.MaxIterations {
		rings = append(rings, slices.Clone(frontier))
		if checkOutputs {
			for _, state := range frontier {
				output, input, found, badErr := tree.bad(state, &budget)
				if badErr != nil {
					return BBRReachability{}, badErr
				}
				if found {
					result := BBRReachability{Rings: cloneBBRRings(rings), Counterexample: bbrTrace(state, input, output, predecessor, predecessorInput, depth)}
					result.Reached = bbrVisited(visited)
					result.Statistics = BBRStatistics{Operations: budget.operations, States: len(result.Reached), Iterations: iterations}
					return result, nil
				}
			}
		}
		nextFrontier := make([]uint64, 0)
		for _, state := range frontier {
			for input := 0; input < inputCount; input++ {
				next, nextErr := tree.next(state, uint64(input), &budget)
				if nextErr != nil {
					return BBRReachability{}, nextErr
				}
				if !visited[next] {
					if discovered >= normalized.MaxStates {
						return BBRReachability{}, ErrLimit
					}
					visited[next] = true
					discovered++
					predecessor[next] = state
					predecessorInput[next] = uint64(input)
					depth[next] = depth[state] + 1
					nextFrontier = append(nextFrontier, next)
				}
			}
		}
		slices.Sort(nextFrontier)
		frontier = nextFrontier
		iterations++
	}
	reached := bbrVisited(visited)
	if len(frontier) != 0 {
		// Every state in Reached must be represented by exactly one returned
		// onion ring, including the first unprocessed ring at an iteration cap.
		rings = append(rings, slices.Clone(frontier))
	}
	return BBRReachability{
		Reached: reached, Rings: cloneBBRRings(rings), Complete: len(frontier) == 0,
		Statistics: BBRStatistics{Operations: budget.operations, States: len(reached), Iterations: iterations},
	}, nil
}

func bbrVisited(visited []bool) []uint64 {
	result := make([]uint64, 0)
	for state, present := range visited {
		if present {
			result = append(result, uint64(state))
		}
	}
	return result
}

func cloneBBRRings(rings [][]uint64) [][]uint64 {
	result := make([][]uint64, len(rings))
	for index := range rings {
		result[index] = slices.Clone(rings[index])
	}
	return result
}

// StateSetBDD is the owned replacement for the global-BDD storage helpers.
// It builds a canonical function over the low StateBits variables.
func (system BBRTransitionSystem) StateSetBDD(manager *ExtraBDDManager, states []uint64, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if manager == nil || manager.variables < system.StateBits {
		return 0, ErrDimension
	}
	if manager.closed {
		return 0, ErrInvariant
	}
	stateCount, err := bbrSpace(system.StateBits)
	if err != nil {
		return 0, err
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	if normalized.Stop != nil && normalized.Stop() {
		return 0, ErrCancelled
	}
	maximum := stateCount
	if len(states) > math.MaxInt/8 {
		return 0, ErrLimit
	}
	preallocation, err := bbrCheckedAdd(stateCount, len(states)*8)
	if err != nil || preallocation >= normalized.MaxBytes {
		return 0, ErrLimit
	}
	buildOperations := 2*stateCount - 1
	if buildOperations > normalized.MaxOperations {
		return 0, ErrLimit
	}
	canonical, err := canonicalBBRStates(states, stateCount, maximum)
	if err != nil {
		return 0, err
	}
	values := make([]bool, stateCount)
	for _, state := range canonical {
		values[state] = true
	}
	normalized.MaxBytes -= preallocation
	return manager.runBooleanOperation(nil, normalized, func(work *extraBDDWork) (ExtraBDDLit, error) {
		var build func(int, int) (ExtraBDDLit, error)
		build = func(variable, assignment int) (ExtraBDDLit, error) {
			if err := work.step(); err != nil {
				return 0, err
			}
			if variable == system.StateBits {
				if values[assignment] {
					return ExtraBDDTrue, nil
				}
				return ExtraBDDFalse, nil
			}
			low, lowErr := build(variable+1, assignment)
			if lowErr != nil {
				return 0, lowErr
			}
			high, highErr := build(variable+1, assignment|(1<<uint(variable)))
			if highErr != nil {
				return 0, highErr
			}
			return manager.uniqueCreate(variable, high, low, work.transaction)
		}
		return build(0, 0)
	})
}

// BBRSourceCoverage records the deliberate stand-alone boundary.
var BBRSourceCoverage = [...]string{
	"bbrImage.c: partitioned image/preimage and owned tree lifecycle are covered; native quantification scheduling and CUDD reorder interaction are adapted",
	"bbrCex.c: deterministic shortest trace reconstruction is covered",
	"bbrNtbdd.c: canonical state-set BDD construction is covered; AIG scratch-pointer storage and progress output are absent",
	"bbrReach.c: initialization, outputs, partitions, onion rings, limits, and result statuses are covered; ABC frame/network mutation is absent",
}
