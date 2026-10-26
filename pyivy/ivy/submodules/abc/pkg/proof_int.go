package abc

// Stand-alone Go port of src/proof/int (excluding the separate proof/int2).

import "slices"

type ProofIntParams struct {
	BacktrackLimit                  int
	MaxFrames                       int
	Seconds                         int
	InductionK                      int
	Rewrite                         bool
	TransitionLoop                  bool
	UsePudlak, UseOther, UseMiniSat bool
	CheckKStep, UseBias, Backward   bool
	SeparateOutputs, TwoFrames      bool
	DropSATOutputs, DumpInvariant   bool
	Verbose                         bool
	FrameReached                    int
	FileName                        string
}

func DefaultProofIntParams() ProofIntParams {
	return ProofIntParams{InductionK: 1, CheckKStep: true, FrameReached: -1}
}

type ProofIntLimits struct {
	MaxAssignments, MaxStates, MaxFrames, MaxObjects, MaxProofClauses int
	Stop                                                              func() bool
}

func normalizeProofIntLimits(l ProofIntLimits) (ProofIntLimits, error) {
	if l.MaxAssignments < 0 || l.MaxStates < 0 || l.MaxFrames < 0 || l.MaxObjects < 0 || l.MaxProofClauses < 0 {
		return l, ErrDimension
	}
	if l.MaxAssignments == 0 {
		l.MaxAssignments = 1 << 20
	}
	if l.MaxStates == 0 {
		l.MaxStates = 1 << 20
	}
	if l.MaxFrames == 0 {
		l.MaxFrames = 1000
	}
	if l.MaxObjects == 0 {
		l.MaxObjects = 1 << 24
	}
	if l.MaxProofClauses == 0 {
		l.MaxProofClauses = 1 << 24
	}
	return l, nil
}

func proofIntCancelled(l ProofIntLimits) error {
	if l.Stop != nil && l.Stop() {
		return ErrCancelled
	}
	return nil
}

func proofIntObjectLimit(graph *AigManager, limits ProofIntLimits) error {
	if graph.ObjectCount() > limits.MaxObjects {
		return ErrLimit
	}
	return nil
}

func proofIntConstraintsHold(outputs []bool, constraints int) bool {
	if constraints < 0 || constraints > len(outputs) {
		return false
	}
	for _, value := range outputs[len(outputs)-constraints:] {
		if value {
			return false
		}
	}
	return true
}

func proofIntCheckVariableMap(variableMap []int) (int, error) {
	maximum := -1
	used := make([]bool, 0)
	for _, mapped := range variableMap {
		if mapped < -2 {
			return 0, ErrDimension
		}
		if mapped < 0 {
			continue
		}
		if mapped >= len(used) {
			used = append(used, make([]bool, mapped-len(used)+1)...)
		}
		if used[mapped] {
			return 0, ErrDimension
		}
		used[mapped] = true
		if mapped > maximum {
			maximum = mapped
		}
	}
	return maximum + 1, nil
}

func proofIntCheckGraph(g *AigManager, sequential bool) error {
	if g == nil {
		return ErrDimension
	}
	if err := g.Check(); err != nil {
		return err
	}
	if sequential && g.RegisterCount() == 0 {
		return ErrDimension
	}
	return nil
}

func proofIntCopyLogic(dst, src *AigManager, mapping []AigLit, limits ProofIntLimits) error {
	for id, object := range src.objects {
		if err := proofIntCancelled(limits); err != nil {
			return err
		}
		if dst.ObjectCount() > limits.MaxObjects {
			return ErrLimit
		}
		switch object.Type {
		case AigConstant1:
			mapping[id] = AigConst1
		case AigAND:
			a, b := mapping[object.Fanin0.ID()].NotIf(object.Fanin0.Complemented()), mapping[object.Fanin1.ID()].NotIf(object.Fanin1.Complemented())
			lit, err := dst.And(a, b)
			if err != nil {
				return err
			}
			mapping[id] = lit
		case AigBuffer:
			mapping[id] = mapping[object.Fanin0.ID()].NotIf(object.Fanin0.Complemented())
		}
	}
	return nil
}

func proofIntMapped(mapping []AigLit, lit AigLit) AigLit {
	return mapping[lit.ID()].NotIf(lit.Complemented())
}

func ProofIntInitialState(registers int, limits ProofIntLimits) (*AigManager, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return nil, err
	}
	if registers <= 0 {
		return nil, ErrDimension
	}
	if err = proofIntCancelled(limits); err != nil {
		return nil, err
	}
	g := NewAigManager()
	terms := make([]AigLit, registers)
	for i := range terms {
		terms[i], err = g.CreateCI()
		if err != nil {
			return nil, err
		}
		terms[i] = terms[i].Not()
		if g.ObjectCount() > limits.MaxObjects {
			return nil, ErrLimit
		}
	}
	root := terms[0]
	for _, term := range terms[1:] {
		root, err = g.And(root, term)
		if err != nil {
			return nil, err
		}
	}
	_, err = g.CreateCO(root)
	if err != nil {
		return nil, err
	}
	if err = proofIntObjectLimit(g, limits); err != nil {
		return nil, err
	}
	return g, nil
}

func ProofIntDuplicateTransition(src *AigManager, limits ProofIntLimits) (*AigManager, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = proofIntCheckGraph(src, true); err != nil {
		return nil, err
	}
	dst := NewAigManager()
	dst.Name, dst.Spec = src.Name, src.Spec
	mapping := make([]AigLit, len(src.objects))
	mapping[0] = AigConst1
	for _, id := range src.cis {
		mapping[id], err = dst.CreateCI()
		if err != nil {
			return nil, err
		}
	}
	if err = proofIntCopyLogic(dst, src, mapping, limits); err != nil {
		return nil, err
	}
	poCount := src.PrimaryOutputCount()
	for i := poCount - src.constraints; i < poCount; i++ {
		if i >= 0 {
			object := src.objects[src.cos[i]]
			_, err = dst.CreateCO(proofIntMapped(mapping, object.Fanin0).Not())
			if err != nil {
				return nil, err
			}
		}
	}
	for _, co := range src.cos[poCount:] {
		object := src.objects[co]
		_, err = dst.CreateCO(proofIntMapped(mapping, object.Fanin0))
		if err != nil {
			return nil, err
		}
	}
	dst.constraints = src.constraints
	if err = dst.SetRegisterCount(src.registers); err != nil {
		return nil, err
	}
	if err = proofIntObjectLimit(dst, limits); err != nil {
		return nil, err
	}
	return dst, nil
}

func ProofIntOneOutputTransition(src *AigManager, addFirstOutput bool, limits ProofIntLimits) (*AigManager, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = proofIntCheckGraph(src, true); err != nil || src.PrimaryOutputCount() == 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	dst := NewAigManager()
	dst.Name, dst.Spec = src.Name, src.Spec
	mapping := make([]AigLit, len(src.objects))
	mapping[0] = AigConst1
	var control AigLit
	for i, id := range src.cis {
		if i == src.PrimaryInputCount() {
			control, err = dst.CreateCI()
			if err != nil {
				return nil, err
			}
		}
		mapping[id], err = dst.CreateCI()
		if err != nil {
			return nil, err
		}
	}
	if err = proofIntCopyLogic(dst, src, mapping, limits); err != nil {
		return nil, err
	}
	poCount := src.PrimaryOutputCount()
	for i := poCount - src.constraints; i < poCount; i++ {
		if i >= 0 {
			o := src.objects[src.cos[i]]
			_, err = dst.CreateCO(proofIntMapped(mapping, o.Fanin0).Not())
			if err != nil {
				return nil, err
			}
		}
	}
	if addFirstOutput {
		o := src.objects[src.cos[0]]
		_, err = dst.CreateCO(proofIntMapped(mapping, o.Fanin0))
		if err != nil {
			return nil, err
		}
		dst.constraints = src.constraints
		if err = proofIntObjectLimit(dst, limits); err != nil {
			return nil, err
		}
		return dst, nil
	}
	for i, co := range src.cos[poCount:] {
		next := proofIntMapped(mapping, src.objects[co].Fanin0)
		current := mapping[src.cis[src.PrimaryInputCount()+i]]
		mux, e := dst.Mux(control, current, next)
		if e != nil {
			return nil, e
		}
		if _, e = dst.CreateCO(mux); e != nil {
			return nil, e
		}
	}
	dst.constraints = src.constraints
	if err = dst.SetRegisterCount(src.registers); err != nil {
		return nil, err
	}
	if err = proofIntObjectLimit(dst, limits); err != nil {
		return nil, err
	}
	return dst, nil
}

func ProofIntFrames(src *AigManager, frames int, addRegisterOutputs, twoFrames bool, limits ProofIntLimits) (*AigManager, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = proofIntCheckGraph(src, true); err != nil {
		return nil, err
	}
	if frames <= 0 || frames > limits.MaxFrames || src.PrimaryOutputCount()-src.constraints != 1 {
		return nil, ErrDimension
	}
	dst := NewAigManager()
	mapping := make([]AigLit, len(src.objects))
	mapping[0] = AigConst1
	piCount, poCount := src.PrimaryInputCount(), src.PrimaryOutputCount()
	for i := 0; i < src.registers; i++ {
		if addRegisterOutputs {
			mapping[src.cis[piCount+i]] = AigConst0
		} else {
			mapping[src.cis[piCount+i]], err = dst.CreateCI()
			if err != nil {
				return nil, err
			}
		}
	}
	lastProperty := InvalidAigLit
	for frame := 0; frame < frames; frame++ {
		for i := 0; i < piCount; i++ {
			mapping[src.cis[i]], err = dst.CreateCI()
			if err != nil {
				return nil, err
			}
		}
		if err = proofIntCopyLogic(dst, src, mapping, limits); err != nil {
			return nil, err
		}
		property := proofIntMapped(mapping, src.objects[src.cos[0]].Fanin0)
		for i := poCount - src.constraints; i < poCount; i++ {
			if i >= 0 {
				c := proofIntMapped(mapping, src.objects[src.cos[i]].Fanin0)
				if _, err = dst.CreateCO(c.Not()); err != nil {
					return nil, err
				}
			}
		}
		if frame == frames-2 {
			lastProperty = property
		}
		if frame+1 < frames {
			for i, co := range src.cos[poCount:] {
				mapping[src.cis[piCount+i]] = proofIntMapped(mapping, src.objects[co].Fanin0)
			}
		}
	}
	if addRegisterOutputs {
		for _, co := range src.cos[poCount:] {
			if _, err = dst.CreateCO(proofIntMapped(mapping, src.objects[co].Fanin0)); err != nil {
				return nil, err
			}
		}
	} else {
		property := proofIntMapped(mapping, src.objects[src.cos[0]].Fanin0)
		if twoFrames && lastProperty.Valid() {
			property, err = dst.Or(lastProperty, property)
			if err != nil {
				return nil, err
			}
		}
		if _, err = dst.CreateCO(property); err != nil {
			return nil, err
		}
	}
	if err = proofIntObjectLimit(dst, limits); err != nil {
		return nil, err
	}
	return dst, nil
}

// ProofIntBMCFrames is intCtrex.c's zero-initialized unrolling with only the
// property of the last frame exposed as an output.
func ProofIntBMCFrames(src *AigManager, frames int, limits ProofIntLimits) (*AigManager, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = proofIntCheckGraph(src, true); err != nil {
		return nil, err
	}
	if frames <= 0 || frames > limits.MaxFrames || src.PrimaryOutputCount() != 1 {
		return nil, ErrDimension
	}
	dst := NewAigManager()
	mapping := make([]AigLit, len(src.objects))
	mapping[0] = AigConst1
	pis, pos := src.PrimaryInputCount(), src.PrimaryOutputCount()
	for i := 0; i < src.registers; i++ {
		mapping[src.cis[pis+i]] = AigConst0
	}
	for frame := 0; frame < frames; frame++ {
		for i := 0; i < pis; i++ {
			mapping[src.cis[i]], err = dst.CreateCI()
			if err != nil {
				return nil, err
			}
		}
		if err = proofIntCopyLogic(dst, src, mapping, limits); err != nil {
			return nil, err
		}
		if frame+1 < frames {
			for i, co := range src.cos[pos:] {
				mapping[src.cis[pis+i]] = proofIntMapped(mapping, src.objects[co].Fanin0)
			}
		}
	}
	property := proofIntMapped(mapping, src.objects[src.cos[0]].Fanin0)
	if _, err = dst.CreateCO(property); err != nil {
		return nil, err
	}
	if err = proofIntObjectLimit(dst, limits); err != nil {
		return nil, err
	}
	return dst, nil
}

// ProofIntExpand adds unused inputs so that an interpolant can be compared
// against a formula over a larger support, as Inter_ManDupExpand does.
func ProofIntExpand(interpolant, other *AigManager, limits ProofIntLimits) (*AigManager, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = proofIntCheckGraph(interpolant, false); err != nil {
		return nil, err
	}
	if err = proofIntCheckGraph(other, false); err != nil {
		return nil, err
	}
	if interpolant.RegisterCount() != 0 || other.RegisterCount() != 0 || interpolant.InputCount() > other.InputCount() {
		return nil, ErrDimension
	}
	result := interpolant.Clone()
	if err = proofIntObjectLimit(result, limits); err != nil {
		return nil, err
	}
	for result.InputCount() < other.InputCount() {
		if err = proofIntCancelled(limits); err != nil {
			return nil, err
		}
		if result.ObjectCount() >= limits.MaxObjects {
			return nil, ErrLimit
		}
		if _, err = result.CreateCI(); err != nil {
			return nil, err
		}
	}
	if err = proofIntObjectLimit(result, limits); err != nil {
		return nil, err
	}
	return result, nil
}

func ProofIntContainment(newGraph, oldGraph *AigManager, limits ProofIntLimits) (bool, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return false, err
	}
	if err = proofIntCheckGraph(newGraph, false); err != nil {
		return false, err
	}
	if err = proofIntCheckGraph(oldGraph, false); err != nil {
		return false, err
	}
	if newGraph.RegisterCount() != 0 || oldGraph.RegisterCount() != 0 || newGraph.InputCount() != oldGraph.InputCount() || newGraph.OutputCount() != 1 || oldGraph.OutputCount() != 1 {
		return false, ErrDimension
	}
	n := newGraph.InputCount()
	if n >= 63 || uint64(1)<<uint(n) > uint64(limits.MaxAssignments) {
		return false, ErrLimit
	}
	values := make([]bool, n)
	for assignment := uint64(0); assignment < uint64(1)<<uint(n); assignment++ {
		if err = proofIntCancelled(limits); err != nil {
			return false, err
		}
		for i := range values {
			values[i] = assignment>>uint(i)&1 != 0
		}
		a, e := newGraph.Evaluate(values)
		if e != nil {
			return false, e
		}
		b, e := oldGraph.Evaluate(values)
		if e != nil {
			return false, e
		}
		if a[0] && !b[0] {
			return false, nil
		}
	}
	return true, nil
}

func ProofIntEquivalent(first, second *AigManager, limits ProofIntLimits) (bool, error) {
	a, err := ProofIntContainment(first, second, limits)
	if err != nil || !a {
		return a, err
	}
	return ProofIntContainment(second, first, limits)
}

func proofIntEvalState(graph *AigManager, state []bool) (bool, error) {
	if graph == nil || graph.RegisterCount() != 0 || graph.InputCount() != len(state) || graph.OutputCount() != 1 {
		return false, ErrDimension
	}
	values, err := graph.Evaluate(state)
	if err != nil {
		return false, err
	}
	return values[0], nil
}

// ProofIntInductiveContainment implements the forward/backward K-step query
// from intContain.c. A true result means no violating state/input path exists.
func ProofIntInductiveContainment(transition, interpolant *AigManager, steps int, backward bool, limits ProofIntLimits) (bool, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return false, err
	}
	if err = proofIntCheckGraph(transition, true); err != nil {
		return false, err
	}
	if steps <= 0 || steps > limits.MaxFrames || interpolant == nil || interpolant.InputCount() != transition.RegisterCount() || interpolant.OutputCount() != 1 || interpolant.RegisterCount() != 0 {
		return false, ErrDimension
	}
	regs, pis := transition.RegisterCount(), transition.PrimaryInputCount()
	if regs >= 63 || pis >= 63 || uint64(1)<<uint(regs) > uint64(limits.MaxAssignments) || uint64(1)<<uint(pis) > uint64(limits.MaxAssignments) {
		return false, ErrLimit
	}
	assignments := 0
	var visit func([]bool, int) (bool, error)
	visit = func(state []bool, frame int) (bool, error) {
		value, e := proofIntEvalState(interpolant, state)
		if e != nil {
			return false, e
		}
		if (!backward && frame < steps && !value) || (backward && frame == 0 && value) || (backward && frame > 0 && !value) {
			return false, nil
		}
		if frame == steps {
			if backward {
				return value, nil
			}
			return !value, nil
		}
		for inputCode := uint64(0); inputCode < uint64(1)<<uint(pis); inputCode++ {
			if e = proofIntCancelled(limits); e != nil {
				return false, e
			}
			assignments++
			if assignments > limits.MaxAssignments {
				return false, ErrLimit
			}
			input := make([]bool, pis)
			for i := range input {
				input[i] = inputCode>>uint(i)&1 != 0
			}
			_, next, stepErr := transition.Step(input, state)
			if stepErr != nil {
				return false, stepErr
			}
			violates, stepErr := visit(next, frame+1)
			if stepErr != nil || violates {
				return violates, stepErr
			}
		}
		return false, nil
	}
	for stateCode := uint64(0); stateCode < uint64(1)<<uint(regs); stateCode++ {
		state := make([]bool, regs)
		for i := range state {
			state[i] = stateCode>>uint(i)&1 != 0
		}
		violates, e := visit(state, 0)
		if e != nil {
			return false, e
		}
		if violates {
			return false, nil
		}
	}
	return true, nil
}

// ProofIntChecker is the owned incremental K-step checker from intCheck.c.
type ProofIntChecker struct {
	transition   *AigManager
	interpolants []*AigManager
	frames       int
	limits       ProofIntLimits
	closed       bool
}

func NewProofIntChecker(transition *AigManager, frames int, limits ProofIntLimits) (*ProofIntChecker, error) {
	var err error
	limits, err = normalizeProofIntLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = proofIntCheckGraph(transition, true); err != nil {
		return nil, err
	}
	if frames <= 0 || frames > limits.MaxFrames {
		return nil, ErrDimension
	}
	return &ProofIntChecker{transition: transition.Clone(), frames: frames, limits: limits}, nil
}

func (c *ProofIntChecker) Close() {
	if c == nil || c.closed {
		return
	}
	c.closed = true
	c.transition = nil
	c.interpolants = nil
}

func (c *ProofIntChecker) Perform(interpolant *AigManager) (bool, error) {
	if c == nil || c.closed || interpolant == nil || interpolant.RegisterCount() != 0 || interpolant.OutputCount() != 1 || interpolant.InputCount() != c.transition.RegisterCount() {
		return false, ErrDimension
	}
	all := append(slices.Clone(c.interpolants), interpolant)
	regs, pis := c.transition.RegisterCount(), c.transition.PrimaryInputCount()
	if regs >= 63 || pis >= 63 {
		return false, ErrLimit
	}
	assignments := 0
	var visit func([]bool, int) (bool, error)
	visit = func(state []bool, frame int) (bool, error) {
		union := false
		for _, graph := range all {
			value, err := proofIntEvalState(graph, state)
			if err != nil {
				return false, err
			}
			union = union || value
		}
		if frame < c.frames && !union {
			return false, nil
		}
		if frame == c.frames {
			return !union, nil
		}
		for code := uint64(0); code < uint64(1)<<uint(pis); code++ {
			if err := proofIntCancelled(c.limits); err != nil {
				return false, err
			}
			assignments++
			if assignments > c.limits.MaxAssignments {
				return false, ErrLimit
			}
			input := make([]bool, pis)
			for i := range input {
				input[i] = code>>uint(i)&1 != 0
			}
			_, next, err := c.transition.Step(input, state)
			if err != nil {
				return false, err
			}
			bad, err := visit(next, frame+1)
			if err != nil || bad {
				return bad, err
			}
		}
		return false, nil
	}
	passed := true
	for code := uint64(0); code < uint64(1)<<uint(regs) && passed; code++ {
		state := make([]bool, regs)
		for i := range state {
			state[i] = code>>uint(i)&1 != 0
		}
		bad, err := visit(state, 0)
		if err != nil {
			return false, err
		}
		passed = !bad
	}
	c.interpolants = append(c.interpolants, interpolant.Clone())
	return passed, nil
}

func ProofIntCheckInitial(graph *AigManager, limits ProofIntLimits) (*Counterexample, bool, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return nil, false, err
	}
	if err = proofIntCheckGraph(graph, true); err != nil || graph.PrimaryOutputCount() == 0 {
		return nil, false, ErrDimension
	}
	n := graph.PrimaryInputCount()
	if n >= 63 || uint64(1)<<uint(n) > uint64(limits.MaxAssignments) {
		return nil, false, ErrLimit
	}
	inputs, state := make([]bool, n), make([]bool, graph.registers)
	for a := uint64(0); a < uint64(1)<<uint(n); a++ {
		if err = proofIntCancelled(limits); err != nil {
			return nil, false, err
		}
		for i := range inputs {
			inputs[i] = a>>uint(i)&1 != 0
		}
		out, _, e := graph.Step(inputs, state)
		if e != nil {
			return nil, false, e
		}
		if out[0] && proofIntConstraintsHold(out[:graph.PrimaryOutputCount()], graph.constraints) {
			c, _ := NewCounterexample(graph.registers, n, 1)
			c.Property = 0
			for i, v := range inputs {
				_ = c.SetInput(0, i, v)
			}
			return c, true, nil
		}
	}
	return nil, false, nil
}

func ProofIntCheckAllStates(graph *AigManager, limits ProofIntLimits) (bool, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return false, err
	}
	if err = proofIntCheckGraph(graph, true); err != nil || graph.PrimaryOutputCount() == 0 {
		return false, ErrDimension
	}
	n := graph.InputCount()
	if n >= 63 || uint64(1)<<uint(n) > uint64(limits.MaxAssignments) {
		return false, ErrLimit
	}
	values := make([]bool, n)
	for a := uint64(0); a < uint64(1)<<uint(n); a++ {
		if err = proofIntCancelled(limits); err != nil {
			return false, err
		}
		for i := range values {
			values[i] = a>>uint(i)&1 != 0
		}
		out, e := graph.Evaluate(values)
		if e != nil {
			return false, e
		}
		if out[0] && proofIntConstraintsHold(out[:graph.PrimaryOutputCount()], graph.constraints) {
			return false, nil
		}
	}
	return true, nil
}

type proofIntPath struct {
	state  uint64
	inputs [][]bool
}

func ProofIntCounterexample(graph *AigManager, frames int, limits ProofIntLimits) (*Counterexample, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = proofIntCheckGraph(graph, true); err != nil {
		return nil, err
	}
	if frames <= 0 || frames > limits.MaxFrames || graph.PrimaryOutputCount()-graph.constraints != 1 || graph.registers >= 63 || graph.PrimaryInputCount() >= 63 {
		return nil, ErrDimension
	}
	pi := graph.PrimaryInputCount()
	if uint64(1)<<uint(pi) > uint64(limits.MaxAssignments) {
		return nil, ErrLimit
	}
	current := []proofIntPath{{state: 0}}
	assignments := 0
	for frame := 0; frame < frames; frame++ {
		next := make([]proofIntPath, 0)
		seenNext := make(map[uint64]bool)
		for _, path := range current {
			state := make([]bool, graph.registers)
			for i := range state {
				state[i] = path.state>>uint(i)&1 != 0
			}
			for a := uint64(0); a < uint64(1)<<uint(pi); a++ {
				if err = proofIntCancelled(limits); err != nil {
					return nil, err
				}
				assignments++
				if assignments > limits.MaxAssignments {
					return nil, ErrLimit
				}
				in := make([]bool, pi)
				for i := range in {
					in[i] = a>>uint(i)&1 != 0
				}
				out, ns, e := graph.Step(in, state)
				if e != nil {
					return nil, e
				}
				trace := append(slices.Clone(path.inputs), in)
				if !proofIntConstraintsHold(out, graph.constraints) {
					continue
				}
				if out[0] {
					c, _ := NewCounterexample(graph.registers, pi, frame+1)
					c.Property = 0
					for f, row := range trace {
						for i, v := range row {
							_ = c.SetInput(f, i, v)
						}
					}
					return c, nil
				}
				var code uint64
				for i, v := range ns {
					if v {
						code |= uint64(1) << uint(i)
					}
				}
				if seenNext[code] {
					continue
				}
				seenNext[code] = true
				next = append(next, proofIntPath{code, trace})
				if len(next) > limits.MaxStates {
					return nil, ErrLimit
				}
			}
		}
		current = next
	}
	return nil, nil
}

type ProofIntManager struct {
	graph     *AigManager
	params    ProofIntParams
	limits    ProofIntLimits
	closed    bool
	frame     int
	invariant *AigManager
}

func NewProofIntManager(graph *AigManager, params ProofIntParams, limits ProofIntLimits) (*ProofIntManager, error) {
	var err error
	limits, err = normalizeProofIntLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = proofIntCheckGraph(graph, true); err != nil {
		return nil, err
	}
	if params.MaxFrames < 0 {
		return nil, ErrDimension
	}
	if params.MaxFrames == 0 {
		params.MaxFrames = limits.MaxFrames
	}
	if params.MaxFrames > limits.MaxFrames {
		return nil, ErrLimit
	}
	return &ProofIntManager{graph: graph.Clone(), params: params, limits: limits}, nil
}
func (m *ProofIntManager) Close() {
	if m == nil || m.closed {
		return
	}
	m.closed = true
	m.graph = nil
	m.invariant = nil
}
func (m *ProofIntManager) Clean() error {
	if m == nil || m.closed {
		return ErrInvariant
	}
	m.frame = 0
	m.invariant = nil
	return nil
}
func (m *ProofIntManager) Frame() int {
	if m == nil {
		return 0
	}
	return m.frame
}
func (m *ProofIntManager) Invariant() *AigManager {
	if m == nil || m.invariant == nil {
		return nil
	}
	return m.invariant.Clone()
}

type ProofIntResult struct {
	Proved         bool
	Undecided      bool
	Frame          int
	Counterexample *Counterexample
	Invariant      *AigManager
}

func (m *ProofIntManager) Perform() (ProofIntResult, error) {
	if m == nil || m.closed {
		return ProofIntResult{}, ErrInvariant
	}
	if err := proofIntCancelled(m.limits); err != nil {
		return ProofIntResult{}, err
	}
	targetFrame := m.params.MaxFrames
	cex, err := ProofIntCounterexample(m.graph, targetFrame, m.limits)
	if err != nil {
		return ProofIntResult{}, err
	}
	if cex != nil {
		m.frame = cex.Frames() - 1
		return ProofIntResult{Frame: cex.Frames() - 1, Counterexample: cex}, nil
	}
	// Explicit reachability proves only when the complete state space was covered.
	if m.graph.registers >= 63 || uint64(1)<<uint(m.graph.registers) > uint64(m.limits.MaxStates) {
		m.frame = targetFrame
		return ProofIntResult{Undecided: true, Frame: targetFrame}, nil
	}
	inv, err := proofIntReachableInvariant(m.graph, targetFrame, m.limits)
	if err != nil {
		return ProofIntResult{}, err
	}
	if inv == nil {
		m.frame = targetFrame
		return ProofIntResult{Undecided: true, Frame: targetFrame}, nil
	}
	m.frame = targetFrame
	m.invariant = inv
	return ProofIntResult{Proved: true, Frame: targetFrame, Invariant: inv.Clone()}, nil
}

func proofIntReachableInvariant(graph *AigManager, frames int, limits ProofIntLimits) (*AigManager, error) {
	regs, pi := graph.registers, graph.PrimaryInputCount()
	seen := map[uint64]bool{0: true}
	front := []uint64{0}
	assignments := 0
	for f := 0; f < frames; f++ {
		var next []uint64
		for _, code := range front {
			state := make([]bool, regs)
			for i := range state {
				state[i] = code>>uint(i)&1 != 0
			}
			for a := uint64(0); a < uint64(1)<<uint(pi); a++ {
				if err := proofIntCancelled(limits); err != nil {
					return nil, err
				}
				assignments++
				if assignments > limits.MaxAssignments {
					return nil, ErrLimit
				}
				in := make([]bool, pi)
				for i := range in {
					in[i] = a>>uint(i)&1 != 0
				}
				outputs, ns, err := graph.Step(in, state)
				if err != nil {
					return nil, err
				}
				if !proofIntConstraintsHold(outputs, graph.constraints) {
					continue
				}
				var n uint64
				for i, v := range ns {
					if v {
						n |= uint64(1) << uint(i)
					}
				}
				if !seen[n] {
					seen[n] = true
					next = append(next, n)
					if len(seen) > limits.MaxStates {
						return nil, ErrLimit
					}
				}
			}
		}
		if len(next) == 0 {
			return proofIntStatesAIG(regs, seen, limits)
		}
		front = next
	}
	return nil, nil
}

func proofIntStatesAIG(regs int, states map[uint64]bool, limits ProofIntLimits) (*AigManager, error) {
	g := NewAigManager()
	vars := make([]AigLit, regs)
	var err error
	for i := range vars {
		vars[i], err = g.CreateCI()
		if err != nil {
			return nil, err
		}
	}
	sum := AigConst0
	ordered := make([]uint64, 0, len(states))
	for state := range states {
		ordered = append(ordered, state)
	}
	slices.Sort(ordered)
	for _, state := range ordered {
		term := AigConst1
		for i, v := range vars {
			term, err = g.And(term, v.NotIf(state>>uint(i)&1 == 0))
			if err != nil {
				return nil, err
			}
		}
		sum, err = g.Or(sum, term)
		if err != nil {
			return nil, err
		}
		if g.ObjectCount() > limits.MaxObjects {
			return nil, ErrLimit
		}
	}
	_, err = g.CreateCO(sum)
	if err != nil {
		return nil, err
	}
	if err = proofIntObjectLimit(g, limits); err != nil {
		return nil, err
	}
	return g, nil
}

func ProofIntInterpolateM114P(proof *M114PProofSnapshot, rootIsB []bool, variableMap []int, pudlak bool, limits ProofIntLimits) (*AigManager, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return nil, err
	}
	if proof == nil || len(rootIsB) != len(proof.Roots) {
		return nil, ErrDimension
	}
	if len(proof.Roots)+len(proof.Chains) > limits.MaxProofClauses {
		return nil, ErrLimit
	}
	sharedCount, err := proofIntCheckVariableMap(variableMap)
	if err != nil {
		return nil, err
	}
	for _, root := range proof.Roots {
		for _, literal := range root {
			if int(literal.Var()) >= len(variableMap) {
				return nil, ErrDimension
			}
		}
	}
	g := NewAigManager()
	for i := 0; i < sharedCount; i++ {
		if _, err = g.CreateCI(); err != nil {
			return nil, err
		}
	}
	values := make([]AigLit, len(proof.Roots), len(proof.Roots)+len(proof.Chains))
	clauses := make([][]SatLit, len(proof.Roots), len(proof.Roots)+len(proof.Chains))
	resolutionSteps := 0
	for i, root := range proof.Roots {
		clauses[i] = slices.Clone(root)
		if rootIsB[i] {
			values[i] = AigConst1
			continue
		}
		values[i] = AigConst0
		if !pudlak {
			for _, lit := range root {
				v := int(lit.Var())
				if v >= len(variableMap) {
					return nil, ErrDimension
				}
				if variableMap[v] >= 0 {
					x, _ := g.IthVar(variableMap[v])
					x = x.NotIf(lit.Negative())
					values[i], err = g.Or(values[i], x)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	for _, chain := range proof.Chains {
		if err = proofIntCancelled(limits); err != nil {
			return nil, err
		}
		if len(chain.ClauseIDs) < 2 || len(chain.Variables) != len(chain.ClauseIDs)-1 {
			return nil, ErrInvariant
		}
		first := chain.ClauseIDs[0]
		if first < 0 || first >= len(values) {
			return nil, ErrInvariant
		}
		value := values[first]
		res := slices.Clone(clauses[first])
		for k, id := range chain.ClauseIDs[1:] {
			resolutionSteps++
			if resolutionSteps > limits.MaxAssignments {
				return nil, ErrLimit
			}
			if id < 0 || id >= len(values) {
				return nil, ErrInvariant
			}
			pivot := int(chain.Variables[k])
			if pivot >= len(variableMap) {
				return nil, ErrDimension
			}
			next := clauses[id]
			resolved, e := proofIntResolveOn(res, next, SatVar(pivot))
			if e != nil {
				return nil, e
			}
			switch variableMap[pivot] {
			case -1:
				value, err = g.Or(value, values[id])
			case -2:
				value, err = g.And(value, values[id])
			default:
				if variableMap[pivot] < 0 {
					return nil, ErrDimension
				}
				if pudlak {
					polarity := -1
					for _, lit := range next {
						if int(lit.Var()) == pivot {
							if lit.Negative() {
								polarity = 1
							} else {
								polarity = 0
							}
							break
						}
					}
					if polarity < 0 {
						return nil, ErrInvariant
					}
					x, _ := g.IthVar(variableMap[pivot])
					x = x.NotIf(polarity != 0)
					value, err = g.Mux(x, value, values[id])
				} else {
					value, err = g.And(value, values[id])
				}
			}
			if err != nil {
				return nil, err
			}
			res = resolved
		}
		if !slices.Equal(res, chain.Clause) {
			return nil, ErrInvariant
		}
		values = append(values, value)
		clauses = append(clauses, slices.Clone(res))
		if err = proofIntObjectLimit(g, limits); err != nil {
			return nil, err
		}
	}
	if proof.FinalClause < 0 || proof.FinalClause >= len(values) || len(clauses[proof.FinalClause]) != 0 {
		return nil, ErrInvariant
	}
	_, err = g.CreateCO(values[proof.FinalClause])
	if err != nil {
		return nil, err
	}
	if err = proofIntObjectLimit(g, limits); err != nil {
		return nil, err
	}
	return g, nil
}

// ProofIntInterpolateStore is the native proof-store path used by
// Inter_ManPerformOneStep. It delegates replay to the validated SAT proof-core
// port and returns an owned AIG interpolant.
func ProofIntInterpolateStore(store *SatProofStore, pudlak bool, limits ProofIntLimits) (SatAigInterpolation, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return SatAigInterpolation{}, err
	}
	if err = proofIntCancelled(limits); err != nil {
		return SatAigInterpolation{}, err
	}
	mode := SatAigInterpolationStandard
	if pudlak {
		mode = SatAigInterpolationMux
	}
	manager, err := NewSatAigInterpolator(SatAigInterpolationOptions{
		Mode:          mode,
		MaxAigObjects: limits.MaxObjects,
		Replay: SatProofCoreLimits{
			MaxClauses:         limits.MaxProofClauses,
			MaxResolutionSteps: limits.MaxAssignments,
			Stop:               limits.Stop,
		},
	})
	if err != nil {
		return SatAigInterpolation{}, err
	}
	defer manager.Stop()
	return manager.Interpolate(store)
}

func proofIntResolveOn(a, b []SatLit, pivot SatVar) ([]SatLit, error) {
	foundA, foundB := false, false
	out := make([]SatLit, 0, len(a)+len(b))
	var pa, pb SatLit
	add := func(l SatLit) error {
		for _, x := range out {
			if x.Var() == l.Var() {
				if x != l {
					return ErrInvariant
				}
				return nil
			}
		}
		out = append(out, l)
		return nil
	}
	for _, l := range a {
		if l.Var() == pivot {
			foundA = true
			pa = l
		} else if err := add(l); err != nil {
			return nil, err
		}
	}
	for _, l := range b {
		if l.Var() == pivot {
			foundB = true
			pb = l
		} else if err := add(l); err != nil {
			return nil, err
		}
	}
	if !foundA || !foundB || pa != pb.Not() {
		return nil, ErrInvariant
	}
	slices.Sort(out)
	return out, nil
}

func ProofIntVerifyInterpolant(proof *M114PProofSnapshot, rootIsB []bool, variableMap []int, interpolant *AigManager, limits ProofIntLimits) (bool, error) {
	limits, err := normalizeProofIntLimits(limits)
	if err != nil {
		return false, err
	}
	if proof == nil || interpolant == nil || len(rootIsB) != len(proof.Roots) || interpolant.OutputCount() != 1 {
		return false, ErrDimension
	}
	sharedCount, err := proofIntCheckVariableMap(variableMap)
	if err != nil {
		return false, err
	}
	if interpolant.InputCount() != sharedCount {
		return false, ErrDimension
	}
	vars := len(variableMap)
	if vars >= 63 || uint64(1)<<uint(vars) > uint64(limits.MaxAssignments) {
		return false, ErrLimit
	}
	shared := interpolant.InputCount()
	all := make([]bool, vars)
	in := make([]bool, shared)
	for a := uint64(0); a < uint64(1)<<uint(vars); a++ {
		if err = proofIntCancelled(limits); err != nil {
			return false, err
		}
		for i := range all {
			all[i] = a>>uint(i)&1 != 0
		}
		satA, satB := true, true
		for i, c := range proof.Roots {
			ok := false
			for _, l := range c {
				if int(l.Var()) >= len(all) {
					return false, ErrDimension
				}
				ok = ok || (all[l.Var()] != l.Negative())
			}
			if rootIsB[i] {
				satB = satB && ok
			} else {
				satA = satA && ok
			}
		}
		for v, m := range variableMap {
			if m >= 0 {
				if m >= shared {
					return false, ErrDimension
				}
				in[m] = all[v]
			}
		}
		out, e := interpolant.Evaluate(in)
		if e != nil {
			return false, e
		}
		if (satA && !out[0]) || (satB && out[0]) {
			return false, nil
		}
	}
	return true, nil
}
