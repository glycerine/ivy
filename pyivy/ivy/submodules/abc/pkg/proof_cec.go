package abc

// Derived from ABC src/proof/cec. This file provides the owned, deterministic
// equivalence, simulation, partition, pattern, and sequential replay surface.

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

type CECStatus uint8

const (
	CECUnknown CECStatus = iota
	CECEquivalent
	CECDifferent
)

type CECOptions struct {
	MaxPatterns   int
	MaxOperations int
	MaxBytes      int
	MaxNodes      int
	MaxClauses    int
	MaxPartitions int
	MaxDepth      int
	MaxStates     int
	MaxFrames     int
	Stop          func() bool
}

type CECCounterexample struct {
	Output       int
	Frame        int
	InitialLeft  []bool
	InitialRight []bool
	Inputs       [][]bool
	LeftValue    bool
	RightValue   bool
}

type CECResult struct {
	Status       CECStatus
	Patterns     int
	States       int
	Frames       int
	Nodes        int
	Clauses      int // conservative structural CNF clause count
	Partitions   int
	Bytes        int
	Work         int
	Depth        int
	Counter      *CECCounterexample
	OutputStatus []CECStatus
}

type CECOutputClass struct {
	Representative int
	Members        []int
	Complemented   []bool
}

type cecBudget struct {
	options  CECOptions
	work     int
	bytes    int
	nodes    int
	clauses  int
	parts    int
	patterns int
	depth    int
}

func normalizeCECOptions(options CECOptions) (CECOptions, error) {
	if options.MaxPatterns < 0 || options.MaxOperations < 0 || options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxClauses < 0 || options.MaxPartitions < 0 || options.MaxDepth < 0 || options.MaxStates < 0 || options.MaxFrames < 0 {
		return CECOptions{}, ErrDimension
	}
	if options.MaxPatterns == 0 {
		options.MaxPatterns = 1 << 16
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 24
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 24
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 1 << 25
	}
	if options.MaxPartitions == 0 {
		options.MaxPartitions = 1 << 16
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 1 << 20
	}
	if options.MaxStates == 0 {
		options.MaxStates = 1 << 16
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 64
	}
	return options, nil
}

func (budget *cecBudget) result(result CECResult) CECResult {
	result.Nodes = budget.nodes
	result.Clauses = budget.clauses
	result.Partitions = budget.parts
	result.Bytes = budget.bytes
	result.Work = budget.work
	result.Depth = budget.depth
	return result
}

func (budget *cecBudget) pattern() error {
	if budget.patterns == budget.options.MaxPatterns {
		return ErrLimit
	}
	budget.patterns++
	return nil
}

func (budget *cecBudget) partition() error {
	if budget.parts == budget.options.MaxPartitions {
		return ErrLimit
	}
	budget.parts++
	return nil
}

func newCECBudget(options CECOptions) (*cecBudget, error) {
	options, err := normalizeCECOptions(options)
	if err != nil {
		return nil, err
	}
	return &cecBudget{options: options}, nil
}

func (budget *cecBudget) step(count int) error {
	if count < 0 || budget.work > budget.options.MaxOperations-count {
		return ErrLimit
	}
	budget.work += count
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (budget *cecBudget) retain(count int) error {
	if count < 0 || budget.bytes > budget.options.MaxBytes-count {
		return ErrLimit
	}
	budget.bytes += count
	return nil
}

func cecProduct(left, right int) (int, error) {
	if left < 0 || right < 0 || left != 0 && right > math.MaxInt/left {
		return 0, ErrLimit
	}
	return left * right, nil
}

func cecSum(counts ...int) (int, error) {
	total := 0
	for _, count := range counts {
		if count < 0 || total > math.MaxInt-count {
			return 0, ErrLimit
		}
		total += count
	}
	return total, nil
}

func cecCloneBits(values []bool) []bool {
	result := make([]bool, len(values))
	copy(result, values)
	return result
}

func cecCloneFrames(values [][]bool) [][]bool {
	result := make([][]bool, len(values))
	for i := range values {
		result[i] = cecCloneBits(values[i])
	}
	return result
}

func checkCECGraph(graph *AigManager, budget *cecBudget) error {
	if graph == nil {
		return ErrInvariant
	}
	if !utf8.ValidString(graph.Name) || !utf8.ValidString(graph.Spec) || strings.IndexByte(graph.Name, 0) >= 0 || strings.IndexByte(graph.Spec, 0) >= 0 {
		return fmt.Errorf("%w: malformed graph name or specification", ErrInvariant)
	}
	if err := budget.step(graph.ObjectCount() + graph.InputCount() + graph.OutputCount()); err != nil {
		return err
	}
	if err := graph.Check(); err != nil {
		return err
	}
	nodes := graph.ObjectCount()
	clauses := 3*graph.AndCount() + graph.OutputCount()
	if nodes > budget.options.MaxNodes-budget.nodes || clauses > budget.options.MaxClauses-budget.clauses {
		return ErrLimit
	}
	depth, err := graph.MaxOutputLevel()
	if err != nil {
		return err
	}
	if uint64(depth) > uint64(budget.options.MaxDepth) {
		return ErrLimit
	}
	budget.nodes += nodes
	budget.clauses += clauses
	if int(depth) > budget.depth {
		budget.depth = int(depth)
	}
	if err := budget.retain(graph.ObjectCount()/8 + 1); err != nil {
		return err
	}
	return nil
}

func cecPatternCount(inputs, maximum int) (int, error) {
	if inputs < 0 || inputs >= 63 {
		return 0, ErrLimit
	}
	patterns := uint64(1) << uint(inputs)
	if patterns > uint64(maximum) || patterns > uint64(math.MaxInt) {
		return 0, ErrLimit
	}
	return int(patterns), nil
}

func cecPattern(index, inputs int, target []bool) {
	for i := 0; i < inputs; i++ {
		target[i] = index&(1<<uint(i)) != 0
	}
}

func checkCECCombinationalPair(left, right *AigManager, budget *cecBudget) error {
	if err := checkCECGraph(left, budget); err != nil {
		return err
	}
	if err := checkCECGraph(right, budget); err != nil {
		return err
	}
	if left.RegisterCount() != 0 || right.RegisterCount() != 0 || left.InputCount() != right.InputCount() || left.OutputCount() != right.OutputCount() {
		return ErrDimension
	}
	return nil
}

// EvaluateCECCombinationalMiter evaluates the output-aligned XOR miter for one
// assignment. The returned slice is caller-owned; true entries distinguish the
// corresponding output pair.
func EvaluateCECCombinationalMiter(left, right *AigManager, assignment []bool, options CECOptions) ([]bool, error) {
	budget, err := newCECBudget(options)
	if err != nil {
		return nil, err
	}
	if err := checkCECCombinationalPair(left, right, budget); err != nil {
		return nil, err
	}
	if len(assignment) != left.InputCount() {
		return nil, ErrDimension
	}
	if err := budget.pattern(); err != nil {
		return nil, err
	}
	if err := budget.step(left.ObjectCount() + right.ObjectCount() + left.OutputCount()); err != nil {
		return nil, err
	}
	if err := budget.retain(left.OutputCount()); err != nil {
		return nil, err
	}
	lv, err := left.Evaluate(assignment)
	if err != nil {
		return nil, err
	}
	rv, err := right.Evaluate(assignment)
	if err != nil {
		return nil, err
	}
	miter := make([]bool, len(lv))
	for output := range miter {
		miter[output] = lv[output] != rv[output]
	}
	return miter, nil
}

// CompareAIGCombinational performs exact deterministic equivalence checking.
// The first differing assignment is returned in source input order.
func CompareAIGCombinational(left, right *AigManager, options CECOptions) (CECResult, error) {
	budget, err := newCECBudget(options)
	if err != nil {
		return CECResult{}, err
	}
	if err := checkCECCombinationalPair(left, right, budget); err != nil {
		return CECResult{}, err
	}
	patterns, err := cecPatternCount(left.InputCount(), budget.options.MaxPatterns)
	if err != nil {
		return CECResult{}, err
	}
	if err := budget.retain(left.InputCount() + left.OutputCount()); err != nil {
		return CECResult{}, err
	}
	assignment := make([]bool, left.InputCount())
	status := make([]CECStatus, left.OutputCount())
	for range status {
		if err := budget.partition(); err != nil {
			return CECResult{}, err
		}
	}
	for pattern := 0; pattern < patterns; pattern++ {
		if err := budget.pattern(); err != nil {
			return CECResult{}, err
		}
		if err := budget.step(left.ObjectCount() + right.ObjectCount() + left.OutputCount()); err != nil {
			return CECResult{}, err
		}
		cecPattern(pattern, len(assignment), assignment)
		leftValues, err := left.Evaluate(assignment)
		if err != nil {
			return CECResult{}, err
		}
		rightValues, err := right.Evaluate(assignment)
		if err != nil {
			return CECResult{}, err
		}
		for output := range leftValues {
			if leftValues[output] == rightValues[output] {
				continue
			}
			status[output] = CECDifferent
			if err := budget.step(len(assignment)); err != nil {
				return CECResult{}, err
			}
			if err := budget.retain(len(assignment)); err != nil {
				return CECResult{}, err
			}
			counter := &CECCounterexample{Output: output, Inputs: [][]bool{cecCloneBits(assignment)}, LeftValue: leftValues[output], RightValue: rightValues[output]}
			return budget.result(CECResult{Status: CECDifferent, Patterns: pattern + 1, Frames: 1, Counter: counter, OutputStatus: status}), nil
		}
	}
	for output := range status {
		status[output] = CECEquivalent
	}
	return budget.result(CECResult{Status: CECEquivalent, Patterns: patterns, Frames: 1, OutputStatus: status}), nil
}

// ReplayCECCombinational checks a returned witness independently.
func ReplayCECCombinational(left, right *AigManager, counter *CECCounterexample, options CECOptions) error {
	budget, err := newCECBudget(options)
	if err != nil {
		return err
	}
	if counter == nil || counter.Frame != 0 || len(counter.Inputs) != 1 {
		return ErrDimension
	}
	if err := checkCECCombinationalPair(left, right, budget); err != nil {
		return err
	}
	if counter.Output < 0 || counter.Output >= left.OutputCount() || len(counter.Inputs[0]) != left.InputCount() {
		return ErrDimension
	}
	if err := budget.pattern(); err != nil {
		return err
	}
	if err := budget.step(left.ObjectCount() + right.ObjectCount()); err != nil {
		return err
	}
	lv, err := left.Evaluate(counter.Inputs[0])
	if err != nil {
		return err
	}
	rv, err := right.Evaluate(counter.Inputs[0])
	if err != nil {
		return err
	}
	if lv[counter.Output] == rv[counter.Output] || lv[counter.Output] != counter.LeftValue || rv[counter.Output] != counter.RightValue {
		return fmt.Errorf("%w: counterexample does not distinguish output", ErrInvariant)
	}
	return nil
}

type cecSequentialNode struct {
	left, right []bool
	trace       [][]bool
}

func cecStateKey(left, right []bool) string {
	var builder strings.Builder
	builder.Grow(len(left) + len(right) + 1)
	for _, value := range left {
		if value {
			builder.WriteByte('1')
		} else {
			builder.WriteByte('0')
		}
	}
	builder.WriteByte('/')
	for _, value := range right {
		if value {
			builder.WriteByte('1')
		} else {
			builder.WriteByte('0')
		}
	}
	return builder.String()
}

func checkCECSequentialPair(left, right *AigManager, budget *cecBudget) error {
	if err := checkCECGraph(left, budget); err != nil {
		return err
	}
	if err := checkCECGraph(right, budget); err != nil {
		return err
	}
	if left.PrimaryInputCount() != right.PrimaryInputCount() || left.PrimaryOutputCount() != right.PrimaryOutputCount() {
		return ErrDimension
	}
	return nil
}

// CompareAIGSequential explores paired reachable states breadth first and
// returns the shortest distinguishing input trace within MaxFrames.
func CompareAIGSequential(left, right *AigManager, initialLeft, initialRight []bool, options CECOptions) (CECResult, error) {
	budget, err := newCECBudget(options)
	if err != nil {
		return CECResult{}, err
	}
	if err := checkCECSequentialPair(left, right, budget); err != nil {
		return CECResult{}, err
	}
	if len(initialLeft) != left.RegisterCount() || len(initialRight) != right.RegisterCount() {
		return CECResult{}, ErrDimension
	}
	patterns, err := cecPatternCount(left.PrimaryInputCount(), budget.options.MaxPatterns)
	if err != nil {
		return CECResult{}, err
	}
	if err := budget.retain(len(initialLeft) + len(initialRight) + 64); err != nil {
		return CECResult{}, err
	}
	queue := []cecSequentialNode{{left: cecCloneBits(initialLeft), right: cecCloneBits(initialRight)}}
	seen := map[string]struct{}{cecStateKey(initialLeft, initialRight): {}}
	assignment := make([]bool, left.PrimaryInputCount())
	explored := 0
	frames := 0
	truncated := false
	for range left.PrimaryOutputCount() {
		if err := budget.partition(); err != nil {
			return CECResult{}, err
		}
	}
	for head := 0; head < len(queue); head++ {
		node := queue[head]
		if len(node.trace) >= budget.options.MaxFrames {
			truncated = true
			continue
		}
		for pattern := 0; pattern < patterns; pattern++ {
			if err := budget.pattern(); err != nil {
				return CECResult{}, err
			}
			if err := budget.step(left.ObjectCount() + right.ObjectCount() + left.PrimaryOutputCount()); err != nil {
				return CECResult{}, err
			}
			cecPattern(pattern, len(assignment), assignment)
			lo, ln, err := left.Step(assignment, node.left)
			if err != nil {
				return CECResult{}, err
			}
			ro, rn, err := right.Step(assignment, node.right)
			if err != nil {
				return CECResult{}, err
			}
			explored++
			traceLength := len(node.trace) + 1
			if traceLength > frames {
				frames = traceLength
			}
			for output := range lo {
				if lo[output] != ro[output] {
					traceBits, err := cecProduct(traceLength, len(assignment))
					if err != nil {
						return CECResult{}, err
					}
					counterBytes, err := cecSum(len(initialLeft), len(initialRight), traceBits, 64)
					if err != nil {
						return CECResult{}, err
					}
					if err := budget.step(counterBytes); err != nil {
						return CECResult{}, err
					}
					if err := budget.retain(counterBytes); err != nil {
						return CECResult{}, err
					}
					trace := cecCloneFrames(node.trace)
					trace = append(trace, cecCloneBits(assignment))
					counter := &CECCounterexample{Output: output, Frame: len(trace) - 1, InitialLeft: cecCloneBits(initialLeft), InitialRight: cecCloneBits(initialRight), Inputs: trace, LeftValue: lo[output], RightValue: ro[output]}
					return budget.result(CECResult{Status: CECDifferent, Patterns: explored, States: len(seen), Frames: len(trace), Counter: counter}), nil
				}
			}
			key := cecStateKey(ln, rn)
			if _, found := seen[key]; found {
				continue
			}
			if len(seen) == budget.options.MaxStates {
				return CECResult{}, ErrLimit
			}
			traceBits, err := cecProduct(traceLength, len(assignment))
			if err != nil {
				return CECResult{}, err
			}
			stateBytes, err := cecSum(len(ln), len(rn), traceBits, len(key), 64)
			if err != nil {
				return CECResult{}, err
			}
			if err := budget.step(len(ln) + len(rn) + traceBits); err != nil {
				return CECResult{}, err
			}
			if err := budget.retain(stateBytes); err != nil {
				return CECResult{}, err
			}
			trace := cecCloneFrames(node.trace)
			trace = append(trace, cecCloneBits(assignment))
			seen[key] = struct{}{}
			queue = append(queue, cecSequentialNode{left: cecCloneBits(ln), right: cecCloneBits(rn), trace: trace})
		}
	}
	status := CECEquivalent
	if truncated {
		status = CECUnknown
	}
	return budget.result(CECResult{Status: status, Patterns: explored, States: len(seen), Frames: frames}), nil
}

// ReplayCECSequential independently checks a sequential witness.
func ReplayCECSequential(left, right *AigManager, counter *CECCounterexample, options CECOptions) error {
	budget, err := newCECBudget(options)
	if err != nil {
		return err
	}
	if counter == nil || counter.Frame < 0 || counter.Frame >= len(counter.Inputs) || len(counter.Inputs) != counter.Frame+1 {
		return ErrDimension
	}
	if err := checkCECSequentialPair(left, right, budget); err != nil {
		return err
	}
	if len(counter.InitialLeft) != left.RegisterCount() || len(counter.InitialRight) != right.RegisterCount() || counter.Output < 0 || counter.Output >= left.PrimaryOutputCount() {
		return ErrDimension
	}
	initialBytes, err := cecSum(len(counter.InitialLeft), len(counter.InitialRight))
	if err != nil {
		return err
	}
	if err := budget.step(initialBytes); err != nil {
		return err
	}
	if err := budget.retain(initialBytes); err != nil {
		return err
	}
	ls, rs := cecCloneBits(counter.InitialLeft), cecCloneBits(counter.InitialRight)
	for frame, inputs := range counter.Inputs {
		if len(inputs) != left.PrimaryInputCount() {
			return ErrDimension
		}
		if err := budget.pattern(); err != nil {
			return err
		}
		if err := budget.step(left.ObjectCount() + right.ObjectCount()); err != nil {
			return err
		}
		lo, ln, err := left.Step(inputs, ls)
		if err != nil {
			return err
		}
		ro, rn, err := right.Step(inputs, rs)
		if err != nil {
			return err
		}
		if frame == counter.Frame {
			if lo[counter.Output] == ro[counter.Output] || lo[counter.Output] != counter.LeftValue || ro[counter.Output] != counter.RightValue {
				return fmt.Errorf("%w: sequential witness does not distinguish output", ErrInvariant)
			}
			return nil
		}
		ls, rs = ln, rn
	}
	return ErrInvariant
}

// CECOutputClasses groups combinational outputs with identical or complemented
// truth vectors. Class and member order follows output order.
func CECOutputClasses(graph *AigManager, options CECOptions) ([]CECOutputClass, error) {
	budget, err := newCECBudget(options)
	if err != nil {
		return nil, err
	}
	if err := checkCECGraph(graph, budget); err != nil {
		return nil, err
	}
	if graph.RegisterCount() != 0 {
		return nil, ErrDimension
	}
	patterns, err := cecPatternCount(graph.InputCount(), budget.options.MaxPatterns)
	if err != nil {
		return nil, err
	}
	words := (patterns + 63) / 64
	if words != 0 && graph.OutputCount() > budget.options.MaxBytes/(words*8) {
		return nil, ErrLimit
	}
	if err := budget.retain(graph.OutputCount() * words * 8); err != nil {
		return nil, err
	}
	// Account for signature slice headers, assignment storage, the assigned
	// bitmap, and the worst-case caller-owned class/member result before those
	// allocations are made.
	resultBytes, err := cecProduct(graph.OutputCount(), 64)
	if err != nil {
		return nil, err
	}
	resultBytes, err = cecSum(resultBytes, graph.InputCount())
	if err != nil {
		return nil, err
	}
	if err := budget.retain(resultBytes); err != nil {
		return nil, err
	}
	signatures := make([][]uint64, graph.OutputCount())
	for i := range signatures {
		signatures[i] = make([]uint64, words)
	}
	assignment := make([]bool, graph.InputCount())
	for pattern := 0; pattern < patterns; pattern++ {
		if err := budget.pattern(); err != nil {
			return nil, err
		}
		if err := budget.step(graph.ObjectCount() + graph.OutputCount()); err != nil {
			return nil, err
		}
		cecPattern(pattern, len(assignment), assignment)
		values, err := graph.Evaluate(assignment)
		if err != nil {
			return nil, err
		}
		for output, value := range values {
			if value {
				signatures[output][pattern/64] |= uint64(1) << uint(pattern%64)
			}
		}
	}
	maskLast := ^uint64(0)
	if patterns%64 != 0 {
		maskLast = uint64(1)<<uint(patterns%64) - 1
	}
	assigned := make([]bool, graph.OutputCount())
	classes := make([]CECOutputClass, 0)
	for representative := 0; representative < graph.OutputCount(); representative++ {
		if assigned[representative] {
			continue
		}
		if err := budget.partition(); err != nil {
			return nil, err
		}
		class := CECOutputClass{Representative: representative, Members: []int{representative}, Complemented: []bool{false}}
		assigned[representative] = true
		for candidate := representative + 1; candidate < graph.OutputCount(); candidate++ {
			if assigned[candidate] {
				continue
			}
			if err := budget.step(1); err != nil {
				return nil, err
			}
			equal, complement := true, true
			for word := 0; word < words; word++ {
				if err := budget.step(1); err != nil {
					return nil, err
				}
				mask := ^uint64(0)
				if word == words-1 {
					mask = maskLast
				}
				if signatures[representative][word] != signatures[candidate][word] {
					equal = false
				}
				if signatures[representative][word]&mask != (^signatures[candidate][word])&mask {
					complement = false
				}
			}
			if equal || complement {
				assigned[candidate] = true
				class.Members = append(class.Members, candidate)
				class.Complemented = append(class.Complemented, complement && !equal)
			}
		}
		classes = append(classes, class)
	}
	return classes, nil
}

var CECBuildInventory = [...]string{
	"cecCec.c", "cecChoice.c", "cecClass.c", "cecCore.c", "cecCorr.c", "cecIso.c", "cecMan.c", "cecPat.c", "cecSat.c", "cecSatG.c", "cecSatG2.c", "cecSeq.c", "cecSim.c", "cecSolve.c", "cecSolveG.c", "cecSplit.c", "cecSynth.c", "cecSweep.c",
}

// CECSourceInventory records every tracked source boundary in src/proof/cec.
// CECBuildInventory is the exact module.make default build subset.
var CECSourceInventory = [...]string{
	"cec.c", "cec.h", "cecCec.c", "cecChoice.c", "cecClass.c", "cecCore.c", "cecCorr.c", "cecInt.h", "cecIso.c", "cecMan.c", "cecPat.c", "cecSat.c", "cecSatG.c", "cecSatG2.c", "cecSeq.c", "cecSim.c", "cecSimBack.c", "cecSolve.c", "cecSolveG.c", "cecSplit.c", "cecSweep.c", "cecSynth.c", "module.make",
}
