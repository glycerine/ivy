package abc

// Port of src/sat/bsat/satInterA.c and satInterB.c.

import "slices"

type SatAigInterpolationMode uint8

const (
	SatAigInterpolationStandard SatAigInterpolationMode = iota
	SatAigInterpolationMux
)

type SatAigInterpolationOptions struct {
	Mode            SatAigInterpolationMode
	GlobalVariables []SatVar
	Replay          SatProofCoreLimits
	MaxAigObjects   int // zero: 16M allocated IDs
}

type SatAigInterpolationStats struct {
	Variables, Roots, Clauses, Globals int
	ResolutionSteps, RootAssignments   int
	AigObjects, AigAnds, Removed       int
	Stopped                            bool
}

type SatAigInterpolation struct {
	Graph     *AigManager
	Output    AigLit
	Variables []SatVar
	Mode      SatAigInterpolationMode
}

type SatAigPartitionFormula struct {
	Graph            *AigManager
	Output           AigLit
	SharedVariables  []SatVar
	InputForVariable []int
}

type SatAigInterpolator struct {
	core          *SatProofCoreManager
	mode          SatAigInterpolationMode
	globals       []SatVar
	maxAigObjects int
	stats         SatAigInterpolationStats
	busy, stopped bool
}

type satAigInterpolationRun struct {
	graph         *AigManager
	mode          SatAigInterpolationMode
	variableTypes []int
	values        []AigLit
	inputs        []AigLit
	maximum       int
}

func checkSatAigInterpolationGlobals(variables []SatVar) error {
	for i, variable := range variables {
		if _, err := MakeSatLit(variable, false); err != nil {
			return err
		}
		if slices.Contains(variables[:i], variable) {
			return ErrInvariant
		}
	}
	return nil
}

func NewSatAigInterpolator(options SatAigInterpolationOptions) (*SatAigInterpolator, error) {
	if options.Mode != SatAigInterpolationStandard && options.Mode != SatAigInterpolationMux {
		return nil, ErrDimension
	}
	if options.MaxAigObjects < 0 {
		return nil, ErrDimension
	}
	if err := checkSatAigInterpolationGlobals(options.GlobalVariables); err != nil {
		return nil, err
	}
	core, err := NewSatProofCoreManager(options.Replay)
	if err != nil {
		return nil, err
	}
	maximum := options.MaxAigObjects
	if maximum == 0 {
		maximum = 16 << 20
	}
	return &SatAigInterpolator{core: core, mode: options.Mode, globals: slices.Clone(options.GlobalVariables), maxAigObjects: maximum}, nil
}

func (m *SatAigInterpolator) SetGlobalVariables(variables []SatVar) error {
	if m == nil || m.stopped || m.busy {
		return ErrInvariant
	}
	if err := checkSatAigInterpolationGlobals(variables); err != nil {
		return err
	}
	m.globals = slices.Clone(variables)
	return nil
}

func satAigInterpolationVariables(run *satProofCoreRun, clausesA int, explicit []SatVar) ([]SatVar, []int, error) {
	inA := make([]bool, len(run.assigns))
	inB := make([]bool, len(run.assigns))
	for clauseID := 0; clauseID < run.manager.stats.Roots; clauseID++ {
		for _, literal := range run.clauses[clauseID].literals {
			if clauseID < clausesA {
				inA[literal.Var()] = true
			} else {
				inB[literal.Var()] = true
			}
		}
	}
	globals := slices.Clone(explicit)
	if len(globals) == 0 {
		for variable := range inA {
			if inA[variable] && inB[variable] {
				globals = append(globals, SatVar(variable))
			}
		}
	}
	if len(globals) > run.manager.limits.MaxVariables {
		return nil, nil, ErrLimit
	}
	types := make([]int, len(run.assigns))
	for variable := range types {
		if inA[variable] {
			types[variable] = 1
		}
	}
	for index, variable := range globals {
		if int(variable) >= len(types) {
			return nil, nil, ErrInvalidLiteral
		}
		types[variable] = -index - 1
	}
	for variable := range types {
		if inA[variable] && inB[variable] && !slices.Contains(globals, SatVar(variable)) {
			return nil, nil, ErrInvariant
		}
	}
	return globals, types, nil
}

func newSatAigInterpolationRun(run *satProofCoreRun, clausesA int, mode SatAigInterpolationMode, maximum int, explicit []SatVar) (*satAigInterpolationRun, []SatVar, error) {
	globals, variableTypes, err := satAigInterpolationVariables(run, clausesA, explicit)
	if err != nil {
		return nil, nil, err
	}
	interpolation := &satAigInterpolationRun{
		graph:         NewAigManager(),
		mode:          mode,
		variableTypes: variableTypes,
		values:        make([]AigLit, len(run.clauses)),
		inputs:        make([]AigLit, len(globals)),
		maximum:       maximum,
	}
	for index := range interpolation.inputs {
		input, inputErr := interpolation.graph.IthVar(index)
		if inputErr != nil {
			return nil, nil, inputErr
		}
		interpolation.inputs[index] = input
	}
	if err = interpolation.checkLimit(); err != nil {
		return nil, nil, err
	}
	for clauseID := 0; clauseID < run.manager.stats.Roots; clauseID++ {
		if clauseID >= clausesA {
			interpolation.values[clauseID] = AigConst1
			continue
		}
		value := AigConst0
		for _, literal := range run.clauses[clauseID].literals {
			kind := variableTypes[literal.Var()]
			if kind >= 0 {
				continue
			}
			input := interpolation.inputs[-kind-1].NotIf(literal.Negative())
			value, err = interpolation.graph.Or(value, input)
			if err != nil {
				return nil, nil, err
			}
			if err = interpolation.checkLimit(); err != nil {
				return nil, nil, err
			}
		}
		interpolation.values[clauseID] = value
	}
	return interpolation, globals, nil
}

func (r *satAigInterpolationRun) checkLimit() error {
	if r.graph.ObjectCount() > r.maximum {
		return ErrLimit
	}
	return nil
}

func (r *satAigInterpolationRun) copy(destination, source int) {
	r.values[destination] = r.values[source]
}

func (r *satAigInterpolationRun) resolve(destination, reason, variable int, reasonLiterals []SatLit) error {
	current := r.values[destination]
	other := r.values[reason]
	var result AigLit
	var err error
	switch kind := r.variableTypes[variable]; {
	case kind == 1:
		result, err = r.graph.Or(current, other)
	case kind == 0 || r.mode == SatAigInterpolationStandard:
		result, err = r.graph.And(current, other)
	default:
		global := -kind - 1
		if global < 0 || global >= len(r.inputs) {
			return ErrInvariant
		}
		pivot := -1
		for i, literal := range reasonLiterals {
			if int(literal.Var()) == variable {
				pivot = i
				break
			}
		}
		if pivot < 0 {
			return ErrInvariant
		}
		if reasonLiterals[pivot].Negative() {
			result, err = r.graph.Mux(r.inputs[global], other, current)
		} else {
			result, err = r.graph.Mux(r.inputs[global], current, other)
		}
	}
	if err != nil {
		return err
	}
	if err = r.checkLimit(); err != nil {
		return err
	}
	r.values[destination] = result
	return nil
}

func (m *SatAigInterpolator) Interpolate(store *SatProofStore) (SatAigInterpolation, error) {
	if m == nil || m.stopped || m.busy || m.core == nil || m.core.stopped {
		return SatAigInterpolation{}, ErrInvariant
	}
	if store == nil || store.stopped || store.clausesA < 0 || store.clausesA > store.roots {
		return SatAigInterpolation{}, ErrInvariant
	}
	m.busy = true
	defer func() { m.busy, m.core.busy = false, false }()
	run, err := m.core.newRun(store)
	m.core.busy = true
	if err != nil {
		return SatAigInterpolation{}, err
	}
	interpolation, globals, err := newSatAigInterpolationRun(run, store.clausesA, m.mode, m.maxAigObjects, m.globals)
	if err != nil {
		return SatAigInterpolation{}, err
	}
	run.aigInterpolation = interpolation
	if err = run.derive(); err != nil {
		return SatAigInterpolation{}, err
	}
	output := interpolation.values[run.empty]
	if _, err = interpolation.graph.CreateCO(output); err != nil {
		return SatAigInterpolation{}, err
	}
	if err = interpolation.checkLimit(); err != nil {
		return SatAigInterpolation{}, err
	}
	removed := interpolation.graph.Cleanup()
	if err = interpolation.graph.Check(); err != nil {
		return SatAigInterpolation{}, err
	}
	coreStats := m.core.Stats()
	m.stats = SatAigInterpolationStats{
		Variables:       coreStats.Variables,
		Roots:           coreStats.Roots,
		Clauses:         coreStats.Clauses,
		Globals:         len(globals),
		ResolutionSteps: coreStats.ResolutionSteps,
		RootAssignments: coreStats.RootAssignments,
		AigObjects:      interpolation.graph.ObjectCount(),
		AigAnds:         interpolation.graph.AndCount(),
		Removed:         removed,
	}
	return SatAigInterpolation{Graph: interpolation.graph, Output: output, Variables: globals, Mode: m.mode}, nil
}

func (r SatAigInterpolation) Evaluate(inputs []bool) (bool, error) {
	if r.Graph == nil || len(inputs) != len(r.Variables) || r.Graph.InputCount() != len(r.Variables) || r.Graph.OutputCount() != 1 {
		return false, ErrDimension
	}
	outputs, err := r.Graph.Evaluate(inputs)
	if err != nil {
		return false, err
	}
	return outputs[0], nil
}

func (m *SatAigInterpolator) DerivePartitionCNF(store *SatProofStore, partA bool) (SatAigPartitionFormula, error) {
	if m == nil || m.stopped || m.busy || m.core == nil || m.core.stopped || store == nil || store.stopped || store.clausesA > store.roots {
		return SatAigPartitionFormula{}, ErrInvariant
	}
	m.busy = true
	defer func() { m.busy, m.core.busy = false, false }()
	run, err := m.core.newRun(store)
	m.core.busy = true
	if err != nil {
		return SatAigPartitionFormula{}, err
	}
	globals, variableTypes, err := satAigInterpolationVariables(run, store.clausesA, m.globals)
	if err != nil {
		return SatAigPartitionFormula{}, err
	}
	graph := NewAigManager()
	inputForVariable := make([]int, len(variableTypes))
	for variable, kind := range variableTypes {
		if kind < 0 {
			inputForVariable[variable] = -kind - 1
		} else {
			inputForVariable[variable] = len(globals) + 1 + variable
		}
	}
	output := AigConst1
	first, last := 0, store.clausesA
	if !partA {
		first, last = store.clausesA, store.roots
	}
	for clauseID := first; clauseID < last; clauseID++ {
		clauseValue := AigConst0
		for _, literal := range run.clauses[clauseID].literals {
			input, inputErr := graph.IthVar(inputForVariable[literal.Var()])
			if inputErr != nil {
				return SatAigPartitionFormula{}, inputErr
			}
			clauseValue, err = graph.Or(clauseValue, input.NotIf(literal.Negative()))
			if err != nil {
				return SatAigPartitionFormula{}, err
			}
			if graph.ObjectCount() > m.maxAigObjects {
				return SatAigPartitionFormula{}, ErrLimit
			}
		}
		output, err = graph.And(output, clauseValue)
		if err != nil {
			return SatAigPartitionFormula{}, err
		}
	}
	if _, err = graph.CreateCO(output); err != nil {
		return SatAigPartitionFormula{}, err
	}
	if graph.ObjectCount() > m.maxAigObjects {
		return SatAigPartitionFormula{}, ErrLimit
	}
	if err = graph.Check(); err != nil {
		return SatAigPartitionFormula{}, err
	}
	return SatAigPartitionFormula{graph, output, globals, inputForVariable}, nil
}

func (m *SatAigInterpolator) Stats() SatAigInterpolationStats {
	if m == nil {
		return SatAigInterpolationStats{Stopped: true}
	}
	stats := m.stats
	stats.Stopped = m.stopped
	return stats
}

func (m *SatAigInterpolator) Stop() {
	if m == nil || m.stopped || m.busy {
		return
	}
	m.core.Stop()
	m.globals = nil
	m.stats = SatAigInterpolationStats{Stopped: true}
	m.stopped = true
}
