package abc

// Multi-output proving orchestration from src/sat/bmc/bmcMulti.c.

import (
	"errors"
	"math"
	"slices"
)

type MultiProveOutputStatus uint8

const (
	MultiProveUnresolved MultiProveOutputStatus = iota
	MultiProveCounterexample
	MultiProveConstantZero
)

type MultiProveStopReason uint8

const (
	MultiProveAllResolved MultiProveStopReason = iota
	MultiProveIterationLimit
	MultiProveGapLimit
	MultiProveCancelled
	MultiProveResourceLimit
)

// MultiProveOptions replaces source wall-clock timeouts with deterministic
// frame, iteration and SAT-work limits. Zero values select documented defaults.
// SimulationFrames and SimulationWords control the preliminary sequential
// simulation (20 and 5, matching Ssw_RarSimulate defaults and bmcMulti.c).
// InitialBMCFrames grows by FrameIncreasePercent after every iteration.
type MultiProveOptions struct {
	MaxIterations, GapIterations                 int
	SimulationFrames, SimulationWords            int
	InitialBMCFrames, FrameIncreasePercent       int
	MaxObjects, MaxMappingEntries, MaxSimulation int
	MaxVariables                                 int
	UseSynthesis                                 bool
	SAT                                          SatLimits
	Stop                                         func() bool
}

type MultiProveReport struct {
	Stage                                    string
	Iteration                                int
	PrimaryInputs, PrimaryOutputs, Registers int
	Nodes, Objects, Solved, TotalOutputs     int
	SolvedPercent, SizePercent               float64
}

type MultiProveOutput struct {
	Status         MultiProveOutputStatus
	Counterexample *Counterexample
}

type MultiProveResult struct {
	Outputs                    []MultiProveOutput
	Reports                    []MultiProveReport
	Reason                     MultiProveStopReason
	Iterations, ConstantZero   int
	FinalObjects, FinalOutputs int
}

func normalizeMultiProveOptions(options MultiProveOptions) (MultiProveOptions, error) {
	if options.MaxIterations < 0 || options.GapIterations < 0 || options.SimulationFrames < 0 || options.SimulationWords < 0 || options.InitialBMCFrames < 0 || options.FrameIncreasePercent < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxSimulation < 0 || options.MaxVariables < 0 {
		return options, ErrDimension
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 1000
	}
	if options.SimulationFrames == 0 {
		options.SimulationFrames = 20
	}
	if options.SimulationWords == 0 {
		options.SimulationWords = 5
	}
	if options.InitialBMCFrames == 0 {
		options.InitialBMCFrames = 2
	}
	if options.FrameIncreasePercent == 0 {
		options.FrameIncreasePercent = 100
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if options.MaxSimulation == 0 {
		options.MaxSimulation = 1 << 24
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.SimulationWords > math.MaxInt/64 || options.SimulationWords*64 > options.MaxSimulation {
		return options, ErrLimit
	}
	return options, nil
}

// ProcessMultiProveOutputs implements Gia_ManProcessOutputs. It validates the
// complete transfer first, then moves (without cloning) found counterexamples
// into original-output slots, clears their input slots, compacts the output map
// and returns local output indexes that remain unresolved.
func ProcessMultiProveOutputs(input, output []*Counterexample, outputMap []int) (left, compacted []int, err error) {
	if len(input) != len(outputMap) {
		return nil, nil, ErrDimension
	}
	seen := make([]bool, len(output))
	for local, original := range outputMap {
		if original < 0 || original >= len(output) || seen[original] || output[original] != nil {
			return nil, nil, ErrDimension
		}
		seen[original] = true
		if input[local] != nil {
			if err := input[local].Validate(); err != nil {
				return nil, nil, err
			}
		}
	}
	left = make([]int, 0, len(input))
	compacted = make([]int, 0, len(input))
	for local, original := range outputMap {
		if input[local] == nil {
			left = append(left, local)
			compacted = append(compacted, original)
			continue
		}
		input[local].Property = original
		output[original] = input[local]
		input[local] = nil
	}
	return left, compacted, nil
}

// CountGiaConstantZeroPrimaryOutputs implements Gia_ManCountConst0PosGia.
func CountGiaConstantZeroPrimaryOutputs(graph *GiaManager) (int, error) {
	if graph == nil {
		return 0, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return 0, err
	}
	count := 0
	for _, id := range graph.cos[:graph.PrimaryOutputCount()] {
		driver, err := graph.objects[id].Fanin0(id)
		if err != nil {
			return 0, err
		}
		if driver == GiaConst0 {
			count++
		}
	}
	return count, nil
}

// CountAigConstantZeroPrimaryOutputs implements Gia_ManCountConst0Pos.
func CountAigConstantZeroPrimaryOutputs(graph *AigManager) (int, error) {
	if graph == nil {
		return 0, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return 0, err
	}
	count := 0
	for _, id := range graph.cos[:graph.PrimaryOutputCount()] {
		if graph.objects[id].Fanin0 == AigConst0 {
			count++
		}
	}
	return count, nil
}

// NewMultiProveReport is the returned-data counterpart of Gia_ManMultiReport.
func NewMultiProveReport(graph *AigManager, stage string, iteration, totalOutputs, totalObjects int) (MultiProveReport, error) {
	if graph == nil || stage == "" || iteration < 0 || totalOutputs < 0 || totalObjects < 0 {
		return MultiProveReport{}, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return MultiProveReport{}, err
	}
	r := MultiProveReport{Stage: stage, Iteration: iteration, PrimaryInputs: graph.PrimaryInputCount(), PrimaryOutputs: graph.PrimaryOutputCount(), Registers: graph.RegisterCount(), Nodes: graph.AndCount(), Objects: graph.ObjectCount(), TotalOutputs: totalOutputs}
	r.Solved = totalOutputs - r.PrimaryOutputs
	if r.Solved < 0 {
		return MultiProveReport{}, ErrDimension
	}
	r.SolvedPercent = 100 * float64(r.Solved) / float64(max(1, totalOutputs))
	r.SizePercent = 100 * float64(r.Objects) / float64(max(1, totalObjects))
	return r, nil
}

// MultiProveSynthesize is the concrete owned synthesis boundary used by the
// orchestration. The currently ported path converts through GIA, removes
// unreachable logic and converts back through structural hashing. It preserves
// terminal order and sequential behavior. Area balancing and LUT remapping from
// Gia_ManAigSyn2 remain a separate unported optimization dependency.
func (graph *AigManager) MultiProveSynthesize(options GiaCofactorOptions) (*AigManager, error) {
	if graph == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if graph.ObjectCount() > options.MaxMappingEntries || graph.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	gia, _, err := GiaFromAig(graph)
	if err != nil {
		return nil, err
	}
	if gia.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	cleaned, _, err := gia.Cleanup()
	if err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	result, _, err := cleaned.ToAig()
	if err != nil {
		return nil, err
	}
	if result.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	return result, result.Check()
}

func multiProveStop(options MultiProveOptions) bool {
	return options.Stop != nil && options.Stop() || options.SAT.Stop != nil && options.SAT.Stop()
}

// simulateMultiProve performs deterministic zero/declared-initialized random
// sequential simulation and returns original-coordinate replayable witnesses.
func simulateMultiProve(model *AigerModel, properties []int, iteration int, options MultiProveOptions) ([]*Counterexample, error) {
	result := make([]*Counterexample, len(properties))
	samples := 64 * options.SimulationWords
	// Step visits the complete object arena. Include trace generation and state
	// copying as well so MaxSimulation bounds actual logical work, not merely
	// the number of random bits produced.
	perFrame, err := checkedSize(0, model.graph.ObjectCount(), 1, options.MaxSimulation)
	if err == nil {
		perFrame, err = checkedSize(perFrame, model.graph.PrimaryInputCount(), 1, options.MaxSimulation)
	}
	if err == nil {
		perFrame, err = checkedSize(perFrame, model.graph.RegisterCount(), 1, options.MaxSimulation)
	}
	if err == nil {
		perFrame = max(1, perFrame)
		perFrame, err = checkedSize(0, options.SimulationFrames, perFrame, options.MaxSimulation)
	}
	if err != nil {
		return nil, err
	}
	work, err := checkedSize(0, samples, perFrame, options.MaxSimulation)
	if err != nil {
		return nil, err
	}
	_ = work
	var random ABCRandom
	random.Uint32(true)
	for skip := 0; skip < (iteration*17)%500; skip++ {
		random.Uint32(false)
	}
	for sample := 0; sample < samples; sample++ {
		if multiProveStop(options) {
			return nil, ErrCancelled
		}
		state := make([]bool, model.graph.RegisterCount())
		for register, initial := range model.initial {
			switch initial {
			case LatchInitOne:
				state[register] = true
			case LatchInitUnknown:
				state[register] = random.Uint32(false)&1 != 0
			}
		}
		trace, err := NewCounterexample(len(state), model.graph.PrimaryInputCount(), options.SimulationFrames)
		if err != nil {
			return nil, err
		}
		for register, value := range state {
			trace.set(register, value)
		}
		primary := make([]bool, model.graph.PrimaryInputCount())
		admissible := true
		for frame := 0; frame < options.SimulationFrames; frame++ {
			for input := range primary {
				primary[input] = random.Uint32(false)&1 != 0
				trace.set(trace.registers+frame*trace.inputs+input, primary[input])
			}
			outputs, next, err := model.graph.Step(primary, state)
			if err != nil {
				return nil, err
			}
			for constraint := model.outputs + model.bad; constraint < len(outputs); constraint++ {
				admissible = admissible && outputs[constraint]
			}
			if admissible {
				for local, property := range properties {
					if result[local] != nil || !outputs[property] {
						continue
					}
					witness, err := NewCounterexample(trace.registers, trace.inputs, frame+1)
					if err != nil {
						return nil, err
					}
					witness.Property = property
					for bit := 0; bit < witness.bitCount; bit++ {
						witness.set(bit, trace.bit(bit))
					}
					if err := model.ValidateCounterexample(witness); err != nil {
						return nil, err
					}
					result[local] = witness
				}
			}
			state = next
		}
	}
	return result, nil
}

func multiProveModel(graph *AigManager, source *AigerModel) (*AigerModel, error) {
	return NewAigerModel(graph, source.outputs, source.bad, source.constraints, source.initial)
}

// MultiProve implements the owned multi-output loop corresponding to
// Gia_ManMultiProveAig and Gia_ManMultiProve. Every SAT result is replayed on
// the original model before publication. The receiver is never mutated.
func (model *AigerModel) MultiProve(options MultiProveOptions) (MultiProveResult, error) {
	var result MultiProveResult
	if model == nil {
		return result, ErrDimension
	}
	if err := model.Validate(); err != nil {
		return result, err
	}
	if model.graph.RegisterCount() == 0 {
		return result, ErrDimension
	}
	options, err := normalizeMultiProveOptions(options)
	if err != nil {
		return result, err
	}
	propertyCount := model.outputs + model.bad
	result.Outputs = make([]MultiProveOutput, propertyCount)
	if propertyCount == 0 {
		result.Reason = MultiProveAllResolved
		return result, nil
	}
	currentGraph := model.graph.Clone()
	currentModel, err := multiProveModel(currentGraph, model)
	if err != nil {
		return MultiProveResult{}, err
	}
	totalObjects := currentGraph.ObjectCount()
	totalPhysicalOutputs := currentGraph.PrimaryOutputCount()
	outputMap := make([]int, propertyCount)
	for index := range outputMap {
		outputMap[index] = index
	}
	owned := make([]*Counterexample, propertyCount)
	localFrames, noProgress := options.InitialBMCFrames, 0
	result.Reason = MultiProveIterationLimit
	stopWith := func(cause error) (MultiProveResult, error) {
		switch {
		case errors.Is(cause, ErrCancelled):
			result.Reason = MultiProveCancelled
			return result, nil
		case errors.Is(cause, ErrLimit):
			result.Reason = MultiProveResourceLimit
			return result, nil
		default:
			return result, cause
		}
	}
	for iteration := 0; iteration < options.MaxIterations && len(outputMap) != 0; iteration++ {
		if multiProveStop(options) {
			return stopWith(ErrCancelled)
		}
		result.Iterations = iteration + 1
		before := len(outputMap)
		found, err := simulateMultiProve(currentModel, outputMap, iteration, options)
		if err != nil {
			return stopWith(err)
		}
		_, outputMap, err = ProcessMultiProveOutputs(found, owned, outputMap)
		if err != nil {
			return result, err
		}
		for index, witness := range owned {
			if witness != nil && result.Outputs[index].Counterexample == nil {
				if err := model.ValidateCounterexample(witness); err != nil {
					return result, err
				}
				result.Outputs[index] = MultiProveOutput{Status: MultiProveCounterexample, Counterexample: witness}
			}
		}
		report, err := NewMultiProveReport(currentGraph, "SIM", iteration, totalPhysicalOutputs, totalObjects)
		if err != nil {
			return result, err
		}
		report.PrimaryOutputs = len(outputMap)
		report.Solved = propertyCount - len(outputMap)
		report.SolvedPercent = 100 * float64(report.Solved) / float64(max(1, propertyCount))
		result.Reports = append(result.Reports, report)
		if len(outputMap) == 0 {
			result.Reason = MultiProveAllResolved
			break
		}

		found = make([]*Counterexample, len(outputMap))
		for local, property := range outputMap {
			bmcOptions := BmcOptions{Frames: localFrames, MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, MaxVariables: options.MaxVariables, Properties: []int{property}, SAT: options.SAT}
			callerStop := bmcOptions.SAT.Stop
			bmcOptions.SAT.Stop = func() bool { return multiProveStop(options) || callerStop != nil && callerStop() }
			answer, err := currentModel.BoundedCheck(bmcOptions)
			if err != nil {
				return result, err
			}
			if answer.Status == BmcCounterexample {
				found[local] = answer.Counterexample
			} else if answer.Reason == BmcCancelled {
				return stopWith(ErrCancelled)
			} else if answer.Reason == BmcResourceLimit {
				return stopWith(ErrLimit)
			}
		}
		_, outputMap, err = ProcessMultiProveOutputs(found, owned, outputMap)
		if err != nil {
			return result, err
		}
		for index, witness := range owned {
			if witness != nil && result.Outputs[index].Counterexample == nil {
				if err := model.ValidateCounterexample(witness); err != nil {
					return result, err
				}
				result.Outputs[index] = MultiProveOutput{Status: MultiProveCounterexample, Counterexample: witness}
			}
		}
		report, err = NewMultiProveReport(currentGraph, "BMC", iteration, totalPhysicalOutputs, totalObjects)
		if err != nil {
			return result, err
		}
		report.PrimaryOutputs = len(outputMap)
		report.Solved = propertyCount - len(outputMap)
		report.SolvedPercent = 100 * float64(report.Solved) / float64(max(1, propertyCount))
		result.Reports = append(result.Reports, report)
		if len(outputMap) == 0 {
			result.Reason = MultiProveAllResolved
			break
		}

		if len(outputMap) == before {
			noProgress++
		} else {
			noProgress = 0
		}
		if options.GapIterations != 0 && noProgress >= options.GapIterations {
			result.Reason = MultiProveGapLimit
			break
		}
		if options.UseSynthesis {
			currentGraph, err = currentGraph.MultiProveSynthesize(GiaCofactorOptions{MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxMappingEntries, Stop: options.Stop})
			if err != nil {
				return stopWith(err)
			}
			currentModel, err = multiProveModel(currentGraph, model)
			if err != nil {
				return result, err
			}
			report, err = NewMultiProveReport(currentGraph, "SYN", iteration, totalPhysicalOutputs, totalObjects)
			if err != nil {
				return result, err
			}
			report.PrimaryOutputs = len(outputMap)
			report.Solved = propertyCount - len(outputMap)
			report.SolvedPercent = 100 * float64(report.Solved) / float64(max(1, propertyCount))
			result.Reports = append(result.Reports, report)
		}
		increase := localFrames * options.FrameIncreasePercent / 100
		if increase < 1 {
			increase = 1
		}
		if localFrames > math.MaxInt-increase {
			return stopWith(ErrLimit)
		}
		localFrames += increase
	}
	for property := 0; property < propertyCount; property++ {
		id := currentGraph.cos[property]
		if currentGraph.objects[id].Fanin0 == AigConst0 {
			result.ConstantZero++
			if result.Outputs[property].Status == MultiProveUnresolved {
				result.Outputs[property].Status = MultiProveConstantZero
			}
		}
	}
	unresolved := 0
	for _, output := range result.Outputs {
		if output.Status == MultiProveUnresolved {
			unresolved++
		}
	}
	if unresolved == 0 {
		result.Reason = MultiProveAllResolved
	}
	result.FinalObjects = currentGraph.ObjectCount()
	result.FinalOutputs = unresolved
	// Ensure caller-visible slices do not alias temporary orchestration state.
	result.Reports = slices.Clone(result.Reports)
	return result, nil
}

// MultiProve supplies the zero-initialized AIG semantics used by the source
// Gia_ManMultiProve entry point.
func (manager *GiaManager) MultiProve(options MultiProveOptions) (MultiProveResult, error) {
	if manager == nil {
		return MultiProveResult{}, ErrDimension
	}
	aig, _, err := manager.ToAig()
	if err != nil {
		return MultiProveResult{}, err
	}
	properties := manager.PrimaryOutputCount() - manager.ConstraintCount()
	model, err := NewAigerModel(aig, properties, 0, manager.ConstraintCount(), nil)
	if err != nil {
		return MultiProveResult{}, err
	}
	return model.MultiProve(options)
}
