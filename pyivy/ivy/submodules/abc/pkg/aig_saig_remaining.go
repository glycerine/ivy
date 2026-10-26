package abc

// Deterministic sequential simulation and activity analysis from
// src/aig/saig/saigSimSeq.c and saigSwitch.c. The autonomous Boolean orbit
// helper does not implement the ternary phase-abstraction flow in saigPhase.c.

import "slices"

// SaigCompiledSourceUnits returns the complete source-order inventory from
// src/aig/saig/module.make. Inventory does not imply behavioral coverage.
func SaigCompiledSourceUnits() []string {
	return []string{
		"saigCone.c", "saigConstr.c", "saigConstr2.c", "saigDual.c",
		"saigDup.c", "saigInd.c", "saigIoa.c", "saigIso.c",
		"saigIsoFast.c", "saigIsoSlow.c", "saigMiter.c", "saigOutDec.c",
		"saigPhase.c", "saigRetFwd.c", "saigRetMin.c", "saigRetStep.c",
		"saigScl.c", "saigSimFast.c", "saigSimMv.c", "saigSimSeq.c",
		"saigStrSim.c", "saigSwitch.c", "saigSynch.c", "saigTempor.c",
		"saigTrans.c", "saigWnd.c",
	}
}

// SaigSimulationOptions bounds all work and retained result storage. Zero
// limits select conservative finite defaults. Stop is polled before and during
// simulation; cancellation never returns a partial result.
type SaigSimulationOptions struct {
	MaxFrames, MaxObjects, MaxWork, MaxStoredValues, MaxBytes int
	Stop                                                      func() bool
}

func normalizeSaigSimulationOptions(options SaigSimulationOptions) (SaigSimulationOptions, error) {
	if options.MaxFrames < 0 || options.MaxObjects < 0 || options.MaxWork < 0 || options.MaxStoredValues < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 1 << 16
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 28
	}
	if options.MaxStoredValues == 0 {
		options.MaxStoredValues = 1 << 26
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 128 << 20
	}
	return options, nil
}

const (
	saigRowBytes      = 32
	saigActivityBytes = 48
	saigMapEntryBytes = 64
)

func saigAddBytes(used, count, width, limit int) (int, error) {
	return checkedSize(used, count, width, limit)
}

func saigTraceBytes(frames, objects, outputs, registers, limit int) error {
	if frames == int(^uint(0)>>1) {
		return ErrLimit
	}
	used := 128 // result object and outer slice values
	var err error
	for _, term := range [][2]int{
		{frames, saigRowBytes + outputs},
		{frames + 1, saigRowBytes + registers},
		{objects, saigActivityBytes},
		// Graph checking and frame evaluation are temporary, but bound them too.
		{objects, 5},
		{registers, 2},
		{outputs, 1},
	} {
		used, err = saigAddBytes(used, term[0], term[1], limit)
		if err != nil {
			return err
		}
	}
	return nil
}

func saigOrbitBytes(frames, objects, outputs, registers, limit int) error {
	used := 128
	var err error
	for _, term := range [][2]int{
		{frames, 2*saigRowBytes + outputs + registers + saigMapEntryBytes},
		{frames, registers}, // collision-free state-key bytes
		{objects, 5},        // graph-check references and value scratch
		{registers, 2},
		{outputs, 1},
	} {
		used, err = saigAddBytes(used, term[0], term[1], limit)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaigObjectActivity is the exact population statistic accumulated after the
// requested prefix. Switching uses ABC's 2*p*(1-p) estimate, rather than the
// number of adjacent value changes. Tombstones have Live false.
type SaigObjectActivity struct {
	Live                      bool
	Samples, Ones             int
	ProbabilityOne, Switching float64
}

// SaigSimulation owns every returned slice. States contains the initial state
// followed by one next-state row per input frame.
type SaigSimulation struct {
	outputs  [][]bool
	states   [][]bool
	activity []SaigObjectActivity
}

func cloneBoolRows(rows [][]bool) [][]bool {
	result := make([][]bool, len(rows))
	for index := range rows {
		result[index] = slices.Clone(rows[index])
	}
	return result
}

func (simulation *SaigSimulation) Outputs() [][]bool {
	if simulation == nil {
		return nil
	}
	return cloneBoolRows(simulation.outputs)
}

func (simulation *SaigSimulation) States() [][]bool {
	if simulation == nil {
		return nil
	}
	return cloneBoolRows(simulation.states)
}

func (simulation *SaigSimulation) Activity() []SaigObjectActivity {
	if simulation == nil {
		return nil
	}
	return slices.Clone(simulation.activity)
}

func saigReadValue(values []bool, literal AigLit) bool {
	return values[literal.ID()] != literal.Complemented()
}

func simulateSaigFrame(graph *AigManager, primary, state []bool, values []bool, stop func() bool) ([]bool, []bool, error) {
	if len(primary) != graph.PrimaryInputCount() || len(state) != graph.RegisterCount() || len(values) != graph.ObjectCount() {
		return nil, nil, ErrDimension
	}
	clear(values)
	values[0] = true
	for index, id := range graph.cis[:graph.PrimaryInputCount()] {
		values[id] = primary[index]
	}
	for index, id := range graph.cis[graph.PrimaryInputCount():] {
		values[id] = state[index]
	}
	for index := 1; index < len(graph.objects); index++ {
		if stop != nil && index&255 == 0 && stop() {
			return nil, nil, ErrCancelled
		}
		object := graph.objects[index]
		switch object.Type {
		case AigNone, AigCI:
		case AigAND:
			values[index] = saigReadValue(values, object.Fanin0) && saigReadValue(values, object.Fanin1)
		case AigBuffer, AigCO:
			values[index] = saigReadValue(values, object.Fanin0)
		default:
			return nil, nil, ErrUnsupported
		}
	}
	outputs := make([]bool, graph.PrimaryOutputCount())
	for index, id := range graph.cos[:graph.PrimaryOutputCount()] {
		outputs[index] = values[id]
	}
	next := make([]bool, graph.RegisterCount())
	for index, id := range graph.cos[graph.PrimaryOutputCount():] {
		next[index] = values[id]
	}
	return outputs, next, nil
}

// SimulateSaigTrace performs exact Boolean replay of caller-supplied primary
// input frames. prefix frames are replayed but excluded from activity counts.
// The graph and all caller slices remain unchanged.
func (graph *AigManager) SimulateSaigTrace(primary [][]bool, initial []bool, prefix int, options SaigSimulationOptions) (*SaigSimulation, error) {
	options, err := normalizeSaigSimulationOptions(options)
	if err != nil || graph == nil || prefix < 0 || prefix > len(primary) {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if len(primary) > options.MaxFrames || len(graph.objects) > options.MaxObjects {
		return nil, ErrLimit
	}
	if len(initial) != graph.RegisterCount() {
		return nil, ErrDimension
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	for index, row := range primary {
		if len(row) != graph.PrimaryInputCount() {
			return nil, ErrDimension
		}
		if options.Stop != nil && index&255 == 255 && options.Stop() {
			return nil, ErrCancelled
		}
	}
	if err := saigTraceBytes(len(primary), len(graph.objects), graph.PrimaryOutputCount(), graph.RegisterCount(), options.MaxBytes); err != nil {
		return nil, err
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	perFrameWork, err := checkedSize(0, 2, len(graph.objects), options.MaxWork)
	if err != nil {
		return nil, err
	}
	baseWork, err := checkedSize(0, 1, len(graph.objects), options.MaxWork)
	if err != nil {
		return nil, err
	}
	baseWork, err = checkedSize(baseWork, 1, len(primary), options.MaxWork)
	if err != nil {
		return nil, err
	}
	baseWork, err = checkedSize(baseWork, 1, len(graph.objects), options.MaxWork)
	if err != nil {
		return nil, err
	}
	if _, err := checkedSize(baseWork, len(primary), perFrameWork, options.MaxWork); err != nil {
		return nil, err
	}
	stored, err := checkedSize(0, len(primary), graph.PrimaryOutputCount(), options.MaxStoredValues)
	if err != nil {
		return nil, err
	}
	stored, err = checkedSize(stored, len(primary)+1, graph.RegisterCount(), options.MaxStoredValues)
	if err != nil {
		return nil, err
	}
	// Count row descriptors as retained values too, including zero-width rows.
	stored, err = checkedSize(stored, 2, len(primary), options.MaxStoredValues)
	if err != nil {
		return nil, err
	}
	stored, err = checkedSize(stored, 1, 1, options.MaxStoredValues) // initial state row descriptor
	if err != nil {
		return nil, err
	}
	if _, err = checkedSize(stored, 1, len(graph.objects), options.MaxStoredValues); err != nil {
		return nil, err
	}

	result := &SaigSimulation{
		outputs: make([][]bool, len(primary)), states: make([][]bool, len(primary)+1),
		activity: make([]SaigObjectActivity, len(graph.objects)),
	}
	state := slices.Clone(initial)
	result.states[0] = slices.Clone(state)
	values := make([]bool, len(graph.objects))
	samples := len(primary) - prefix
	for frame, row := range primary {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		outputs, next, frameErr := simulateSaigFrame(graph, row, state, values, options.Stop)
		if frameErr != nil {
			return nil, frameErr
		}
		result.outputs[frame] = outputs
		state = next
		result.states[frame+1] = slices.Clone(state)
		if frame >= prefix {
			for id, value := range values {
				if options.Stop != nil && id&255 == 0 && options.Stop() {
					return nil, ErrCancelled
				}
				if graph.objects[id].Type != AigNone && value {
					result.activity[id].Ones++
				}
			}
		}
	}
	for id := range result.activity {
		activity := &result.activity[id]
		activity.Live = graph.objects[id].Type != AigNone
		if !activity.Live {
			continue
		}
		activity.Samples = samples
		if samples != 0 {
			activity.ProbabilityOne = float64(activity.Ones) / float64(samples)
			activity.Switching = 2 * activity.ProbabilityOne * (1 - activity.ProbabilityOne)
		}
	}
	return result, nil
}

// SaigOrbit describes an autonomous state orbit. When Complete is true,
// States ends just before the first repeated state, Prefix is its first index,
// and Period is the loop length. Outputs aligns with States. A bounded run that
// finds no repetition returns Complete false and Period zero.
type SaigOrbit struct {
	States, Outputs [][]bool
	Prefix, Period  int
	Complete        bool
}

// AnalyzeSaigAutonomousOrbit exactly explores a zero-primary-input machine
// until a repeated state or MaxFrames. It does not conflate a bounded miss with
// a proof that the machine has no cycle.
func (graph *AigManager) AnalyzeSaigAutonomousOrbit(initial []bool, options SaigSimulationOptions) (SaigOrbit, error) {
	options, err := normalizeSaigSimulationOptions(options)
	if err != nil || graph == nil {
		if err != nil {
			return SaigOrbit{}, err
		}
		return SaigOrbit{}, ErrDimension
	}
	if graph.PrimaryInputCount() != 0 || len(initial) != graph.RegisterCount() {
		return SaigOrbit{}, ErrDimension
	}
	if len(graph.objects) > options.MaxObjects {
		return SaigOrbit{}, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return SaigOrbit{}, ErrCancelled
	}
	if err := saigOrbitBytes(options.MaxFrames, len(graph.objects), graph.PrimaryOutputCount(), graph.RegisterCount(), options.MaxBytes); err != nil {
		return SaigOrbit{}, err
	}
	if err := graph.Check(); err != nil {
		return SaigOrbit{}, err
	}
	perFrameWork, err := checkedSize(4, 1, len(graph.objects), options.MaxWork)
	if err != nil {
		return SaigOrbit{}, err
	}
	perFrameWork, err = checkedSize(perFrameWork, 1, graph.RegisterCount(), options.MaxWork)
	if err != nil {
		return SaigOrbit{}, err
	}
	perFrameWork, err = checkedSize(perFrameWork, 1, graph.RegisterCount(), options.MaxWork)
	if err != nil {
		return SaigOrbit{}, err
	}
	perFrameWork, err = checkedSize(perFrameWork, 2, graph.RegisterCount(), options.MaxWork)
	if err != nil {
		return SaigOrbit{}, err
	}
	perFrameWork, err = checkedSize(perFrameWork, 1, graph.PrimaryOutputCount(), options.MaxWork)
	if err != nil {
		return SaigOrbit{}, err
	}
	if _, err := checkedSize(len(graph.objects), options.MaxFrames, perFrameWork, options.MaxWork); err != nil {
		return SaigOrbit{}, err
	}
	// Each visited frame retains a state row, output row, state key and map
	// entry in addition to the Boolean payloads.
	perFrame, err := checkedSize(3, 2, graph.RegisterCount(), options.MaxStoredValues)
	if err != nil {
		return SaigOrbit{}, err
	}
	perFrame, err = checkedSize(perFrame, 1, graph.PrimaryOutputCount(), options.MaxStoredValues)
	if err != nil {
		return SaigOrbit{}, err
	}
	stored, err := checkedSize(0, options.MaxFrames, perFrame, options.MaxStoredValues)
	if err != nil {
		return SaigOrbit{}, err
	}
	if stored > options.MaxStoredValues {
		return SaigOrbit{}, ErrLimit
	}
	result := SaigOrbit{
		States:  make([][]bool, 0, options.MaxFrames),
		Outputs: make([][]bool, 0, options.MaxFrames),
	}
	seen := make(map[string]int, options.MaxFrames)
	state := slices.Clone(initial)
	values := make([]bool, len(graph.objects))
	for frame := 0; frame < options.MaxFrames; frame++ {
		if options.Stop != nil && options.Stop() {
			return SaigOrbit{}, ErrCancelled
		}
		keyBytes := make([]byte, len(state))
		for index, value := range state {
			if value {
				keyBytes[index] = 1
			}
		}
		key := string(keyBytes)
		if first, ok := seen[key]; ok {
			result.Prefix, result.Period, result.Complete = first, frame-first, true
			return result, nil
		}
		seen[key] = frame
		result.States = append(result.States, slices.Clone(state))
		outputs, next, frameErr := simulateSaigFrame(graph, nil, state, values, options.Stop)
		if frameErr != nil {
			return SaigOrbit{}, frameErr
		}
		result.Outputs = append(result.Outputs, outputs)
		state = next
	}
	return result, nil
}
