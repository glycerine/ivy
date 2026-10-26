package abc

// Derived from ABC src/proof/live/{kLiveConstraints.c,combination.c,arenaViolation.c}.

import (
	"math/bits"
	"slices"
	"strings"
)

// LiveTransformOptions bounds graph construction. Zero limits select bounded
// defaults. A failed or cancelled transform never modifies the source graph.
type LiveTransformOptions struct {
	MaxCombinations int
	MaxOperations   int
	Stop            func() bool
}

type liveTransformBudget struct {
	options LiveTransformOptions
	used    int
}

func newLiveTransformBudget(options LiveTransformOptions) (*liveTransformBudget, error) {
	if options.MaxCombinations < 0 || options.MaxOperations < 0 {
		return nil, ErrDimension
	}
	if options.MaxCombinations == 0 {
		options.MaxCombinations = 1 << 16
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 20
	}
	return &liveTransformBudget{options: options}, nil
}

func (budget *liveTransformBudget) step() error {
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	if budget.used >= budget.options.MaxOperations {
		return ErrLimit
	}
	budget.used++
	return nil
}

// LiveCombinationCount implements countCombination without the source's
// exponential recursion. Overflow is reported instead of wrapping long.
func LiveCombinationCount(n, k int) (uint64, error) {
	if n < 0 || k < 0 || k > n {
		return 0, ErrDimension
	}
	if k == 0 || k == n {
		return 1, nil
	}
	if k > n-k {
		k = n - k
	}
	result := uint64(1)
	for i := 1; i <= k; i++ {
		hi, lo := bits.Mul64(result, uint64(n-k+i))
		if hi >= uint64(i) {
			return 0, ErrLimit
		}
		quotient, remainder := bits.Div64(hi, lo, uint64(i))
		if remainder != 0 {
			return 0, ErrInvariant
		}
		result = quotient
	}
	return result, nil
}

// LiveCombinations returns the source Knuth traversal order. Entries within a
// combination are descending because listCombination prints vC[k] through
// vC[1]. Returned storage is caller-owned.
func LiveCombinations(n, k int, options LiveTransformOptions) ([][]int, error) {
	budget, err := newLiveTransformBudget(options)
	if err != nil {
		return nil, err
	}
	count, err := LiveCombinationCount(n, k)
	if err != nil {
		return nil, err
	}
	if count > uint64(budget.options.MaxCombinations) {
		return nil, ErrLimit
	}
	if k == 0 {
		if err = budget.step(); err != nil {
			return nil, err
		}
		return [][]int{{}}, nil
	}
	// c[0] is vC[1] in the source. Advancing the lowest position first
	// produces its colexicographic Knuth order, not ordinary lexicographic
	// combinations.
	current := make([]int, k+1)
	for i := range current {
		current[i] = i
	}
	current[k] = n
	result := make([][]int, 0, int(count))
	for {
		if err = budget.step(); err != nil {
			return nil, err
		}
		combination := make([]int, k)
		for i := 0; i < k; i++ {
			combination[k-1-i] = current[i]
		}
		result = append(result, combination)
		position := 0
		for position < k && current[position]+1 == current[position+1] {
			current[position] = position
			position++
		}
		if position == k {
			break
		}
		current[position]++
	}
	return result, nil
}

// LiveConstrainedResult is a duplicated sequential AIG with one additional
// primary output. OutputIndex locates antecedents-imply-consequent.
type LiveConstrainedResult struct {
	Graph       *AigManager
	OutputIndex int
}

// LiveDisjunctiveResult adds one sticky implication latch and one PO for every
// selected subset. Combinations holds descending indices into the candidate
// list in exact source traversal order.
type LiveDisjunctiveResult struct {
	Graph        *AigManager
	Combinations [][]int
	FirstOutput  int
}

// LiveArenaResult describes the extra state added by the DSC arena transform.
// Register indices address the returned graph's state vector.
type LiveArenaResult struct {
	Graph                  *AigManager
	OutputIndex            int
	BarrierRegisters       []int
	ArenaViolationRegister int
	PendingRegister        int
	MonotoneRegisters      []int
}

func liveMapLiteral(mapping []AigLit, literal AigLit) (AigLit, error) {
	if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	mapped := mapping[literal.ID()]
	if !mapped.Valid() {
		return InvalidAigLit, ErrInvalidLiteral
	}
	return mapped.NotIf(literal.Complemented()), nil
}

// liveCopyCombinational duplicates CIs and internal logic, then appends the
// requested extra latch outputs. COs are deliberately left to each transform
// so that its PO/RI partition is constructed in native order.
func liveCopyCombinational(source *AigManager, extraRegisters int, budget *liveTransformBudget) (*AigManager, []AigLit, []AigLit, error) {
	if source == nil || extraRegisters < 0 {
		return nil, nil, nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, nil, nil, err
	}
	target := NewAigManager()
	mapping := make([]AigLit, len(source.objects))
	for i := range mapping {
		mapping[i] = InvalidAigLit
	}
	mapping[0] = AigConst1
	for _, id := range source.cis {
		if err := budget.step(); err != nil {
			return nil, nil, nil, err
		}
		literal, err := target.CreateCI()
		if err != nil {
			return nil, nil, nil, err
		}
		mapping[id] = literal
	}
	extras := make([]AigLit, extraRegisters)
	for i := range extras {
		if err := budget.step(); err != nil {
			return nil, nil, nil, err
		}
		literal, err := target.CreateCI()
		if err != nil {
			return nil, nil, nil, err
		}
		extras[i] = literal
	}
	for id := 1; id < len(source.objects); id++ {
		object := source.objects[id]
		if object.Type != AigAND && object.Type != AigBuffer {
			continue
		}
		if err := budget.step(); err != nil {
			return nil, nil, nil, err
		}
		first, err := liveMapLiteral(mapping, object.Fanin0)
		if err != nil {
			return nil, nil, nil, err
		}
		if object.Type == AigBuffer {
			mapping[id], err = target.CreateBuffer(first)
		} else {
			second, secondErr := liveMapLiteral(mapping, object.Fanin1)
			if secondErr != nil {
				return nil, nil, nil, secondErr
			}
			mapping[id], err = target.And(first, second)
		}
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return target, mapping, extras, nil
}

func liveOutputDriver(source *AigManager, mapping []AigLit, output int) (AigLit, error) {
	if source == nil || output < 0 || output >= source.OutputCount() {
		return InvalidAigLit, ErrDimension
	}
	object := source.objects[source.cos[output]]
	if object.Type != AigCO {
		return InvalidAigLit, ErrInvariant
	}
	return liveMapLiteral(mapping, object.Fanin0)
}

func liveCreateCO(target *AigManager, driver AigLit, budget *liveTransformBudget) error {
	if err := budget.step(); err != nil {
		return err
	}
	_, err := target.CreateCO(driver)
	return err
}

func liveCopyOutputRange(source, target *AigManager, mapping []AigLit, first, end int, budget *liveTransformBudget) error {
	for index := first; index < end; index++ {
		driver, err := liveOutputDriver(source, mapping, index)
		if err != nil {
			return err
		}
		if err = liveCreateCO(target, driver, budget); err != nil {
			return err
		}
	}
	return nil
}

func liveConstrainedCone(target *AigManager, antecedents []AigLit, consequent AigLit, budget *liveTransformBudget) (AigLit, error) {
	if target == nil || !target.validSignal(consequent) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	antecedent := AigConst1
	for _, signal := range antecedents {
		if !target.validSignal(signal) {
			return InvalidAigLit, ErrInvalidLiteral
		}
		if err := budget.step(); err != nil {
			return InvalidAigLit, err
		}
		var err error
		antecedent, err = target.And(antecedent, signal)
		if err != nil {
			return InvalidAigLit, err
		}
	}
	if err := budget.step(); err != nil {
		return InvalidAigLit, err
	}
	return target.Or(antecedent.Not(), consequent)
}

// BuildConstrainedZeroLive duplicates source and appends the implication of
// the selected primary-output drivers. The original PO and RI ordering is
// unchanged and source is never used as scratch storage.
func BuildConstrainedZeroLive(source *AigManager, antecedentOutputs []int, consequentOutput int, options LiveTransformOptions) (LiveConstrainedResult, error) {
	budget, err := newLiveTransformBudget(options)
	if err != nil {
		return LiveConstrainedResult{}, err
	}
	if source == nil || consequentOutput < 0 || consequentOutput >= source.PrimaryOutputCount() {
		return LiveConstrainedResult{}, ErrDimension
	}
	for _, index := range antecedentOutputs {
		if index < 0 || index >= source.PrimaryOutputCount() {
			return LiveConstrainedResult{}, ErrDimension
		}
	}
	target, mapping, _, err := liveCopyCombinational(source, 0, budget)
	if err != nil {
		return LiveConstrainedResult{}, err
	}
	target.Name, target.Spec = source.Name+"_0Live", ""
	antecedents := make([]AigLit, len(antecedentOutputs))
	for i, index := range antecedentOutputs {
		antecedents[i], err = liveOutputDriver(source, mapping, index)
		if err != nil {
			return LiveConstrainedResult{}, err
		}
	}
	consequent, err := liveOutputDriver(source, mapping, consequentOutput)
	if err != nil {
		return LiveConstrainedResult{}, err
	}
	cone, err := liveConstrainedCone(target, antecedents, consequent, budget)
	if err != nil {
		return LiveConstrainedResult{}, err
	}
	poCount := source.PrimaryOutputCount()
	if err = liveCopyOutputRange(source, target, mapping, 0, poCount, budget); err != nil {
		return LiveConstrainedResult{}, err
	}
	if err = liveCreateCO(target, cone, budget); err != nil {
		return LiveConstrainedResult{}, err
	}
	if err = liveCopyOutputRange(source, target, mapping, poCount, source.OutputCount(), budget); err != nil {
		return LiveConstrainedResult{}, err
	}
	if err = target.SetRegisterCount(source.RegisterCount()); err != nil {
		return LiveConstrainedResult{}, err
	}
	target.Cleanup()
	if err = target.Check(); err != nil {
		return LiveConstrainedResult{}, err
	}
	return LiveConstrainedResult{Graph: target, OutputIndex: poCount}, nil
}

// CollectLiveConstraintOutputs replaces the Abc_Ntk name scan. It returns all
// csLiveConst_ outputs in order and the last csLiveTarget_ output, matching the
// source loop. Names must correspond exactly to primary outputs.
func CollectLiveConstraintOutputs(names []string) ([]int, int, error) {
	antecedents := make([]int, 0)
	consequent := -1
	for index, name := range names {
		if strings.Contains(name, "csLiveConst_") {
			antecedents = append(antecedents, index)
		} else if strings.Contains(name, "csLiveTarget_") {
			consequent = index
		}
	}
	if consequent < 0 {
		return nil, -1, ErrDimension
	}
	return antecedents, consequent, nil
}

// FindLiveOutput returns the first output whose name contains marker.
func FindLiveOutput(names []string, marker string) (int, error) {
	if marker == "" {
		return -1, ErrDimension
	}
	for index, name := range names {
		if strings.Contains(name, marker) {
			return index, nil
		}
	}
	return -1, ErrDimension
}

// GatherLiveNodes is the owned-data replacement for gatherMonotoneSignals'
// stdout-only node dump. The source routine does not perform a monotonicity
// check and always returns nil; callers can format these stable node IDs using
// AigManager.FormatObject without process-global output.
func GatherLiveNodes(source *AigManager) ([]AigID, error) {
	if source == nil {
		return nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	nodes := make([]AigID, 0, source.AndCount())
	for id, object := range source.objects {
		if object.Type == AigAND {
			nodes = append(nodes, AigID(id))
		}
	}
	return nodes, nil
}

// BuildNamedConstrainedZeroLive supplies the generateWorkingAig boundary
// without requiring an Abc_Ntk name manager.
func BuildNamedConstrainedZeroLive(source *AigManager, outputNames []string, options LiveTransformOptions) (LiveConstrainedResult, error) {
	if source == nil || len(outputNames) != source.PrimaryOutputCount() {
		return LiveConstrainedResult{}, ErrDimension
	}
	antecedents, consequent, err := CollectLiveConstraintOutputs(outputNames)
	if err != nil {
		return LiveConstrainedResult{}, err
	}
	return BuildConstrainedZeroLive(source, antecedents, consequent, options)
}

func liveAllCombinations(n, maxK int, options LiveTransformOptions) ([][]int, error) {
	if n <= 0 || maxK <= 0 || maxK > n {
		return nil, ErrDimension
	}
	total := uint64(0)
	for k := 1; k <= maxK; k++ {
		count, err := LiveCombinationCount(n, k)
		if err != nil {
			return nil, err
		}
		if ^uint64(0)-total < count {
			return nil, ErrLimit
		}
		total += count
	}
	limit := options.MaxCombinations
	if limit == 0 {
		limit = 1 << 16
	}
	if total > uint64(limit) {
		return nil, ErrLimit
	}
	result := make([][]int, 0, int(total))
	remaining := options
	remaining.MaxCombinations = limit
	for k := 1; k <= maxK; k++ {
		combinations, err := LiveCombinations(n, k, remaining)
		if err != nil {
			return nil, err
		}
		result = append(result, combinations...)
	}
	return result, nil
}

func buildLiveDisjunctiveTester(source *AigManager, roots []AigLit, maxK int, options LiveTransformOptions) (LiveDisjunctiveResult, error) {
	combinations, err := liveAllCombinations(len(roots), maxK, options)
	if err != nil {
		return LiveDisjunctiveResult{}, err
	}
	budget, err := newLiveTransformBudget(options)
	if err != nil {
		return LiveDisjunctiveResult{}, err
	}
	target, mapping, latchOutputs, err := liveCopyCombinational(source, len(combinations), budget)
	if err != nil {
		return LiveDisjunctiveResult{}, err
	}
	target.Name, target.Spec = source.Name+"_nCk", ""
	mappedRoots := make([]AigLit, len(roots))
	for i, root := range roots {
		mappedRoots[i], err = liveMapLiteral(mapping, root)
		if err != nil {
			return LiveDisjunctiveResult{}, err
		}
	}
	drivers := make([]AigLit, len(combinations))
	for i, combination := range combinations {
		disjunction := AigConst0
		for _, index := range combination {
			if err = budget.step(); err != nil {
				return LiveDisjunctiveResult{}, err
			}
			disjunction, err = target.Or(mappedRoots[index], disjunction)
			if err != nil {
				return LiveDisjunctiveResult{}, err
			}
		}
		if err = budget.step(); err != nil {
			return LiveDisjunctiveResult{}, err
		}
		drivers[i], err = target.Or(disjunction.Not(), latchOutputs[i])
		if err != nil {
			return LiveDisjunctiveResult{}, err
		}
	}
	poCount := source.PrimaryOutputCount()
	if err = liveCopyOutputRange(source, target, mapping, 0, poCount, budget); err != nil {
		return LiveDisjunctiveResult{}, err
	}
	for _, driver := range drivers {
		if err = liveCreateCO(target, driver, budget); err != nil {
			return LiveDisjunctiveResult{}, err
		}
	}
	if err = liveCopyOutputRange(source, target, mapping, poCount, source.OutputCount(), budget); err != nil {
		return LiveDisjunctiveResult{}, err
	}
	for _, driver := range drivers {
		if err = liveCreateCO(target, driver, budget); err != nil {
			return LiveDisjunctiveResult{}, err
		}
	}
	if err = target.SetRegisterCount(source.RegisterCount() + len(combinations)); err != nil {
		return LiveDisjunctiveResult{}, err
	}
	target.Cleanup()
	if err = target.Check(); err != nil {
		return LiveDisjunctiveResult{}, err
	}
	return LiveDisjunctiveResult{Graph: target, Combinations: combinations, FirstOutput: poCount}, nil
}

// BuildLiveDisjunctiveTester implements generateDisjunctiveTester with an
// explicit candidate-output list instead of an Abc_Ntk naming side channel.
func BuildLiveDisjunctiveTester(source *AigManager, candidateOutputs []int, maxK int, options LiveTransformOptions) (LiveDisjunctiveResult, error) {
	if source == nil {
		return LiveDisjunctiveResult{}, ErrDimension
	}
	roots := make([]AigLit, len(candidateOutputs))
	for i, output := range candidateOutputs {
		if output < 0 || output >= source.PrimaryOutputCount() {
			return LiveDisjunctiveResult{}, ErrDimension
		}
		object := source.objects[source.cos[output]]
		roots[i] = object.Fanin0
	}
	return buildLiveDisjunctiveTester(source, roots, maxK, options)
}

// BuildGeneralLiveDisjunctiveTester implements the exhaustive variant over
// every register output.
func BuildGeneralLiveDisjunctiveTester(source *AigManager, maxK int, options LiveTransformOptions) (LiveDisjunctiveResult, error) {
	if source == nil || source.RegisterCount() == 0 {
		return LiveDisjunctiveResult{}, ErrDimension
	}
	roots := make([]AigLit, source.RegisterCount())
	start := source.PrimaryInputCount()
	for i := range roots {
		roots[i] = AigLit(source.cis[start+i]) << 1
	}
	return buildLiveDisjunctiveTester(source, roots, maxK, options)
}

// BuildArenaConstrainedZeroLive implements the active BARRIER_MONOTONE_TEST
// configuration. Barriers contain primary-output indices whose disjunctions
// are sampled at window start and checked for changes within the window.
func BuildArenaConstrainedZeroLive(source *AigManager, outputNames []string, barriers [][]int, options LiveTransformOptions) (LiveArenaResult, error) {
	if source == nil || len(outputNames) != source.PrimaryOutputCount() || len(barriers) == 0 {
		return LiveArenaResult{}, ErrDimension
	}
	antecedentIndices, consequentIndex, err := CollectLiveConstraintOutputs(outputNames)
	if err != nil {
		return LiveArenaResult{}, err
	}
	windowIndex, err := FindLiveOutput(outputNames, "windowBegins_")
	if err != nil {
		return LiveArenaResult{}, err
	}
	withinIndex, err := FindLiveOutput(outputNames, "withinWindow_")
	if err != nil {
		return LiveArenaResult{}, err
	}
	pendingIndex, err := FindLiveOutput(outputNames, "pendingSignal")
	if err != nil {
		return LiveArenaResult{}, err
	}
	for _, barrier := range barriers {
		if len(barrier) == 0 {
			return LiveArenaResult{}, ErrDimension
		}
		for _, output := range barrier {
			if output < 0 || output >= source.PrimaryOutputCount() {
				return LiveArenaResult{}, ErrDimension
			}
		}
	}
	budget, err := newLiveTransformBudget(options)
	if err != nil {
		return LiveArenaResult{}, err
	}
	barrierCount := len(barriers)
	target, mapping, extras, err := liveCopyCombinational(source, 2*barrierCount+2, budget)
	if err != nil {
		return LiveArenaResult{}, err
	}
	target.Name, target.Spec = source.Name+"_0Live", ""
	barrierLO := extras[:barrierCount]
	violationLO := extras[barrierCount]
	pendingLO := extras[barrierCount+1]
	monotoneLO := extras[barrierCount+2:]
	window, err := liveOutputDriver(source, mapping, windowIndex)
	if err != nil {
		return LiveArenaResult{}, err
	}
	within, err := liveOutputDriver(source, mapping, withinIndex)
	if err != nil {
		return LiveArenaResult{}, err
	}
	pending, err := liveOutputDriver(source, mapping, pendingIndex)
	if err != nil {
		return LiveArenaResult{}, err
	}
	barrierSignals := make([]AigLit, barrierCount)
	barrierNext := make([]AigLit, barrierCount)
	violation := AigConst0
	for i, barrier := range barriers {
		signal := AigConst0
		for _, output := range barrier {
			driver, driverErr := liveOutputDriver(source, mapping, output)
			if driverErr != nil {
				return LiveArenaResult{}, driverErr
			}
			if err = budget.step(); err != nil {
				return LiveArenaResult{}, err
			}
			signal, err = target.Or(driver, signal)
			if err != nil {
				return LiveArenaResult{}, err
			}
		}
		barrierSignals[i] = signal
		if err = budget.step(); err != nil {
			return LiveArenaResult{}, err
		}
		atStart, gateErr := target.And(signal, window)
		if gateErr != nil {
			return LiveArenaResult{}, gateErr
		}
		if err = budget.step(); err != nil {
			return LiveArenaResult{}, err
		}
		barrierNext[i], err = target.Or(atStart, barrierLO[i])
		if err != nil {
			return LiveArenaResult{}, err
		}
		if err = budget.step(); err != nil {
			return LiveArenaResult{}, err
		}
		switched, gateErr := target.Xor(signal, barrierLO[i])
		if gateErr != nil {
			return LiveArenaResult{}, gateErr
		}
		if err = budget.step(); err != nil {
			return LiveArenaResult{}, err
		}
		changedWithin, gateErr := target.And(switched, within)
		if gateErr != nil {
			return LiveArenaResult{}, gateErr
		}
		if err = budget.step(); err != nil {
			return LiveArenaResult{}, err
		}
		violation, err = target.Or(changedWithin, violation)
		if err != nil {
			return LiveArenaResult{}, err
		}
	}
	if err = budget.step(); err != nil {
		return LiveArenaResult{}, err
	}
	violationNext, err := target.Or(violation, violationLO)
	if err != nil {
		return LiveArenaResult{}, err
	}
	if err = budget.step(); err != nil {
		return LiveArenaResult{}, err
	}
	pendingActive, err := target.And(pending, pendingLO)
	if err != nil {
		return LiveArenaResult{}, err
	}
	monotone := AigConst1
	for i, signal := range barrierSignals {
		if err = budget.step(); err != nil {
			return LiveArenaResult{}, err
		}
		guard, gateErr := target.And(pendingActive, monotoneLO[i])
		if gateErr != nil {
			return LiveArenaResult{}, gateErr
		}
		if err = budget.step(); err != nil {
			return LiveArenaResult{}, err
		}
		implied, gateErr := target.Or(guard.Not(), signal)
		if gateErr != nil {
			return LiveArenaResult{}, gateErr
		}
		if err = budget.step(); err != nil {
			return LiveArenaResult{}, err
		}
		monotone, err = target.And(monotone, implied)
		if err != nil {
			return LiveArenaResult{}, err
		}
	}
	antecedents := make([]AigLit, len(antecedentIndices))
	for i, output := range antecedentIndices {
		antecedents[i], err = liveOutputDriver(source, mapping, output)
		if err != nil {
			return LiveArenaResult{}, err
		}
	}
	consequent, err := liveOutputDriver(source, mapping, consequentIndex)
	if err != nil {
		return LiveArenaResult{}, err
	}
	constrained, err := liveConstrainedCone(target, antecedents, consequent, budget)
	if err != nil {
		return LiveArenaResult{}, err
	}
	if err = budget.step(); err != nil {
		return LiveArenaResult{}, err
	}
	withViolation, err := target.Or(constrained, violationLO)
	if err != nil {
		return LiveArenaResult{}, err
	}
	if err = budget.step(); err != nil {
		return LiveArenaResult{}, err
	}
	newOutput, err := target.And(withViolation, monotone)
	if err != nil {
		return LiveArenaResult{}, err
	}
	poCount := source.PrimaryOutputCount()
	if err = liveCopyOutputRange(source, target, mapping, 0, poCount, budget); err != nil {
		return LiveArenaResult{}, err
	}
	if err = liveCreateCO(target, newOutput, budget); err != nil {
		return LiveArenaResult{}, err
	}
	if err = liveCopyOutputRange(source, target, mapping, poCount, source.OutputCount(), budget); err != nil {
		return LiveArenaResult{}, err
	}
	for _, driver := range barrierNext {
		if err = liveCreateCO(target, driver, budget); err != nil {
			return LiveArenaResult{}, err
		}
	}
	if err = liveCreateCO(target, violationNext, budget); err != nil {
		return LiveArenaResult{}, err
	}
	if err = liveCreateCO(target, pending, budget); err != nil {
		return LiveArenaResult{}, err
	}
	for _, signal := range barrierSignals {
		if err = liveCreateCO(target, signal, budget); err != nil {
			return LiveArenaResult{}, err
		}
	}
	oldRegisters := source.RegisterCount()
	if err = target.SetRegisterCount(oldRegisters + len(extras)); err != nil {
		return LiveArenaResult{}, err
	}
	target.Cleanup()
	if err = target.Check(); err != nil {
		return LiveArenaResult{}, err
	}
	barrierRegisters := make([]int, barrierCount)
	monotoneRegisters := make([]int, barrierCount)
	for i := range barriers {
		barrierRegisters[i] = oldRegisters + i
		monotoneRegisters[i] = oldRegisters + barrierCount + 2 + i
	}
	return LiveArenaResult{
		Graph:                  target,
		OutputIndex:            poCount,
		BarrierRegisters:       barrierRegisters,
		ArenaViolationRegister: oldRegisters + barrierCount,
		PendingRegister:        oldRegisters + barrierCount + 1,
		MonotoneRegisters:      monotoneRegisters,
	}, nil
}

// CloneLiveCombinations returns independent combination storage.
func CloneLiveCombinations(combinations [][]int) [][]int {
	result := make([][]int, len(combinations))
	for i := range combinations {
		result[i] = slices.Clone(combinations[i])
	}
	return result
}
