package abc

// Port of src/proof/int2/int2.h, int2Int.h, int2Core.c, int2Bmc.c,
// int2Refine.c, and int2Util.c.

import (
	"fmt"
	"math"
	"slices"
	"time"
)

// Int2Parameters contains the controls of Int2_ManPars_t.  MaxStates and
// MaxTransitions are Go-side memory/work guards; zero means unlimited.
type Int2Parameters struct {
	ConflictLimit, StartFrames, MaxFrames, SecondsLimit, InductionFrames int
	Rewrite, TransitionToInitial, ReturnInvariant, Verbose               bool
	FrameReached                                                         int
	MaxStates, MaxTransitions                                            int
	Stop                                                                 func() bool
}

func DefaultInt2Parameters() Int2Parameters {
	return Int2Parameters{StartFrames: 1, InductionFrames: 1, FrameReached: -1}
}

type Int2Status uint8

const (
	Int2Unknown Int2Status = iota
	Int2Safe
	Int2Unsafe
)

type Int2Result struct {
	Status              Int2Status
	FrameReached        int
	Counterexample      *Counterexample
	InvariantStates     [][]bool
	States, Transitions int
	Cancelled, LimitHit bool
}

// Int2Manager owns a snapshot. Stop releases it and all later operations fail.
type Int2Manager struct {
	graph *GiaManager
	pars  Int2Parameters
}

func NewInt2Manager(graph *GiaManager, pars Int2Parameters) (*Int2Manager, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	if graph.RegisterCount() == 0 || graph.PrimaryOutputCount()-graph.ConstraintCount() <= 0 || pars.StartFrames < 1 || pars.MaxFrames < 0 || pars.SecondsLimit < 0 || pars.MaxStates < 0 || pars.MaxTransitions < 0 {
		return nil, ErrDimension
	}
	return &Int2Manager{graph: graph.Clone(), pars: pars}, nil
}

func (m *Int2Manager) Stop() { m.graph = nil }

type int2TraceNode struct {
	state  []bool
	parent int
	input  []bool
}

func int2StateKey(state []bool) string {
	b := make([]byte, (len(state)+7)/8)
	for i, value := range state {
		if value {
			b[i/8] |= byte(1) << uint(i&7)
		}
	}
	return string(b)
}

func int2EnumerateInputs(count int, visit func([]bool) (bool, error)) (bool, error) {
	values := make([]bool, count)
	var walk func(int) (bool, error)
	walk = func(index int) (bool, error) {
		if index == count {
			return visit(values)
		}
		values[index] = false
		if stop, err := walk(index + 1); stop || err != nil {
			return stop, err
		}
		values[index] = true
		return walk(index + 1)
	}
	return walk(0)
}

func int2Witness(nodes []int2TraceNode, index, property int, failingInput []bool) (*Counterexample, error) {
	path := make([]int, 0)
	for index >= 0 {
		path = append(path, index)
		index = nodes[index].parent
	}
	slices.Reverse(path)
	c, err := NewCounterexample(len(nodes[path[0]].state), len(nodes[path[len(path)-1]].input), len(path))
	if err != nil {
		return nil, err
	}
	c.Property = property
	for register, value := range nodes[path[0]].state {
		_ = c.SetInitial(register, value)
	}
	for frame, nodeIndex := range path {
		inputNode := nodeIndex
		if frame+1 < len(path) {
			inputNode = path[frame+1]
		}
		values := nodes[inputNode].input
		if frame+1 == len(path) {
			values = failingInput
		}
		for input, value := range values {
			_ = c.SetInput(frame, input, value)
		}
	}
	return c, nil
}

// PerformInt2Interpolation proves the observable INT2 contract with exact
// image computation.  The checked-in C interpolation prototype is internally
// inconsistent; explicit finite-state images preserve proof and witness
// semantics without depending on that non-compiling SAT implementation.
func PerformInt2Interpolation(graph *GiaManager, pars Int2Parameters) (Int2Result, error) {
	m, err := NewInt2Manager(graph, pars)
	if err != nil {
		return Int2Result{}, err
	}
	defer m.Stop()
	return m.Perform()
}

func (m *Int2Manager) Perform() (Int2Result, error) {
	result := Int2Result{Status: Int2Unknown, FrameReached: -1}
	if m == nil || m.graph == nil {
		return result, ErrInvariant
	}
	g := m.graph
	deadline := time.Time{}
	if m.pars.SecondsLimit != 0 {
		deadline = time.Now().Add(time.Duration(m.pars.SecondsLimit) * time.Second)
	}
	interrupted := func() bool {
		return (m.pars.Stop != nil && m.pars.Stop()) || (!deadline.IsZero() && !time.Now().Before(deadline))
	}
	initial := make([]bool, g.RegisterCount())
	nodes := []int2TraceNode{{state: initial, parent: -1, input: make([]bool, g.PrimaryInputCount())}}
	seen := map[string]int{int2StateKey(initial): 0}
	frontier := []int{0}
	properties := g.PrimaryOutputCount() - g.ConstraintCount()
	for frame := 0; len(frontier) != 0; frame++ {
		if m.pars.MaxFrames != 0 && frame >= m.pars.MaxFrames {
			result.LimitHit = true
			return result, nil
		}
		if interrupted() {
			result.Cancelled = true
			return result, nil
		}
		nextFrontier := make([]int, 0)
		for _, nodeIndex := range frontier {
			stop, enumErr := int2EnumerateInputs(g.PrimaryInputCount(), func(input []bool) (bool, error) {
				if interrupted() {
					result.Cancelled = true
					return true, nil
				}
				if m.pars.MaxTransitions != 0 && result.Transitions >= m.pars.MaxTransitions {
					result.LimitHit = true
					return true, nil
				}
				outputs, next, stepErr := g.Step(input, nodes[nodeIndex].state)
				if stepErr != nil {
					return true, stepErr
				}
				result.Transitions++
				valid := true
				for i := properties; i < len(outputs); i++ {
					if outputs[i] { // ABC constraint outputs are violation signals.
						valid = false
						break
					}
				}
				if !valid {
					return false, nil
				}
				for property := 0; property < properties; property++ {
					if outputs[property] {
						witness, witnessErr := int2Witness(nodes, nodeIndex, property, input)
						if witnessErr != nil {
							return true, witnessErr
						}
						result.Status, result.FrameReached, result.Counterexample = Int2Unsafe, frame, witness
						return true, nil
					}
				}
				key := int2StateKey(next)
				if _, exists := seen[key]; !exists {
					if m.pars.MaxStates != 0 && len(nodes) >= m.pars.MaxStates {
						result.LimitHit = true
						return true, nil
					}
					seen[key] = len(nodes)
					nodes = append(nodes, int2TraceNode{state: slices.Clone(next), parent: nodeIndex, input: slices.Clone(input)})
					nextFrontier = append(nextFrontier, len(nodes)-1)
				}
				return false, nil
			})
			if enumErr != nil {
				return Int2Result{}, enumErr
			}
			if stop {
				result.States = len(nodes)
				return result, nil
			}
		}
		result.FrameReached, result.States = frame, len(nodes)
		if len(nextFrontier) == 0 {
			result.Status = Int2Safe
			if m.pars.ReturnInvariant {
				result.InvariantStates = make([][]bool, len(nodes))
				for i := range nodes {
					result.InvariantStates[i] = slices.Clone(nodes[i].state)
				}
			}
			return result, nil
		}
		frontier = nextFrontier
	}
	return result, nil
}

func int2CopyLogic(source, target *GiaManager, mapping []GiaLit, id GiaID) (GiaLit, error) {
	o := source.objects[id]
	a, err := o.Fanin0(id)
	if err != nil {
		return InvalidGiaLit, err
	}
	mapLit := func(l GiaLit) GiaLit { return mapping[l.ID()].NotIf(l.Complemented()) }
	a = mapLit(a)
	if o.IsBuffer() {
		return target.AppendBuffer(a)
	}
	b, err := o.Fanin1(id)
	if err != nil {
		return InvalidGiaLit, err
	}
	b = mapLit(b)
	if control, yes := source.MuxControl(id); yes {
		return target.AppendMuxReal(mapLit(control), b, a)
	}
	if o.IsXor() {
		return target.AppendXorReal(a, b)
	}
	return target.HashAnd(a, b)
}

// DuplicateInt2InitialTransition adds a final control PI.  When asserted each
// register holds its current value; this creates a transition into zero from
// the zero state as intended by Int2_ManDupInit.
func DuplicateInt2InitialTransition(source *GiaManager) (*GiaManager, error) {
	if source == nil || source.RegisterCount() == 0 {
		return nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	target := NewGiaManager()
	target.Name, target.Spec = source.Name, source.Spec
	mapping := make([]GiaLit, source.ObjectCount())
	mapping[0] = GiaConst0
	for i, id := range source.cis[:source.PrimaryInputCount()] {
		l, e := target.AppendCI()
		if e != nil {
			return nil, e
		}
		mapping[id] = l
		_ = i
	}
	control, err := target.AppendCI()
	if err != nil {
		return nil, err
	}
	for _, id := range source.cis[source.PrimaryInputCount():] {
		l, e := target.AppendCI()
		if e != nil {
			return nil, e
		}
		mapping[id] = l
	}
	if err = target.HashStart(); err != nil {
		return nil, err
	}
	for id := GiaID(1); int(id) < source.ObjectCount(); id++ {
		if source.objects[id].IsAnd() {
			mapping[id], err = int2CopyLogic(source, target, mapping, id)
			if err != nil {
				return nil, err
			}
		}
	}
	for i, id := range source.cos {
		o := source.objects[id]
		d, _ := o.Fanin0(id)
		driver := mapping[d.ID()].NotIf(d.Complemented())
		if i >= source.PrimaryOutputCount() {
			ro := mapping[source.cis[source.PrimaryInputCount()+i-source.PrimaryOutputCount()]]
			driver, err = target.HashMux(control, ro, driver)
			if err != nil {
				return nil, err
			}
		}
		if _, err = target.AppendCO(driver); err != nil {
			return nil, err
		}
	}
	target.HashStop()
	if err = target.SetRegisterCount(source.RegisterCount()); err != nil {
		return nil, err
	}
	if err = target.SetConstraintCount(source.ConstraintCount()); err != nil {
		return nil, err
	}
	return target, nil
}

func Int2HasTransitionToZero(graph *GiaManager, maxTransitions int, stop func() bool) (bool, error) {
	if graph == nil || graph.RegisterCount() == 0 || maxTransitions < 0 {
		return false, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return false, err
	}
	count := 0
	found := false
	_, err := int2EnumerateInputs(graph.RegisterCount(), func(state []bool) (bool, error) {
		return int2EnumerateInputs(graph.PrimaryInputCount(), func(input []bool) (bool, error) {
			if stop != nil && stop() {
				return true, ErrCancelled
			}
			if maxTransitions != 0 && count >= maxTransitions {
				return true, ErrLimit
			}
			count++
			_, next, e := graph.Step(input, state)
			if e != nil {
				return true, e
			}
			zero := true
			for _, v := range next {
				zero = zero && !v
			}
			if zero {
				found = true
				return true, nil
			}
			return false, nil
		})
	})
	return found, err
}

// BuildInt2InitialFrames unrolls transitions from zero and exposes only the
// final register-input vector, matching Int2_ManFrameInit/Int2_ManUnroll.
func BuildInt2InitialFrames(source *GiaManager, frames int) (*GiaManager, error) {
	if source == nil || source.RegisterCount() == 0 || frames < 1 {
		return nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	t := NewGiaManager()
	t.Name, t.Spec = source.Name, source.Spec
	if err := t.HashStart(); err != nil {
		return nil, err
	}
	state := make([]GiaLit, source.RegisterCount())
	for frame := 0; frame < frames; frame++ {
		mapping := make([]GiaLit, source.ObjectCount())
		mapping[0] = GiaConst0
		for _, id := range source.cis[:source.PrimaryInputCount()] {
			l, e := t.AppendCI()
			if e != nil {
				return nil, e
			}
			mapping[id] = l
		}
		for i, id := range source.cis[source.PrimaryInputCount():] {
			mapping[id] = state[i]
		}
		for id := GiaID(1); int(id) < source.ObjectCount(); id++ {
			if source.objects[id].IsAnd() {
				var e error
				mapping[id], e = int2CopyLogic(source, t, mapping, id)
				if e != nil {
					return nil, e
				}
			}
		}
		for i := range state {
			id := source.cos[source.PrimaryOutputCount()+i]
			d, _ := source.objects[id].Fanin0(id)
			state[i] = mapping[d.ID()].NotIf(d.Complemented())
		}
	}
	for _, d := range state {
		if _, err := t.AppendCO(d); err != nil {
			return nil, err
		}
	}
	t.HashStop()
	return t, nil
}

// Int2FrameMap is the checked replacement for vMapFrames and its inline
// check/write helpers.
type Int2FrameMap struct{ frames [][]GiaLit }

func NewInt2FrameMap(frames, objects int) (*Int2FrameMap, error) {
	if frames < 0 || objects < 0 {
		return nil, ErrDimension
	}
	m := &Int2FrameMap{frames: make([][]GiaLit, frames)}
	for i := range m.frames {
		m.frames[i] = make([]GiaLit, objects)
		for j := range m.frames[i] {
			m.frames[i][j] = InvalidGiaLit
		}
	}
	return m, nil
}
func (m *Int2FrameMap) Read(frame, object int) (GiaLit, error) {
	if m == nil || frame < 0 || frame >= len(m.frames) || object < 0 || object >= len(m.frames[frame]) {
		return InvalidGiaLit, ErrDimension
	}
	l := m.frames[frame][object]
	if l == InvalidGiaLit {
		return l, ErrInvariant
	}
	return l, nil
}
func (m *Int2FrameMap) Write(frame, object int, l GiaLit) error {
	if m == nil || frame < 0 || frame >= len(m.frames) || object < 0 || object >= len(m.frames[frame]) || !l.Valid() {
		return ErrDimension
	}
	m.frames[frame][object] = l
	return nil
}

// RefineInt2Cube returns a priority-directed sufficient subset of CI indices.
func RefineInt2Cube(graph *GiaManager, assignment []bool, priority []int) ([]int, error) {
	if graph == nil || graph.RegisterCount() == 0 || graph.PrimaryOutputCount() != 1 || len(assignment) != graph.InputCount() || len(priority) != len(assignment) {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	values, err := graph.evaluateValues(assignment)
	if err != nil {
		return nil, err
	}
	outID := graph.cos[0]
	driver, _ := graph.objects[outID].Fanin0(outID)
	if values[driver.ID()] == driver.Complemented() {
		return nil, ErrInvariant
	}
	propagated := make([]int, graph.ObjectCount())
	for i := range propagated {
		propagated[i] = math.MaxInt
	}
	for _, id := range graph.cis {
		propagated[id] = priority[graph.objects[id].Diff1()]
	}
	for id := GiaID(1); int(id) < graph.ObjectCount(); id++ {
		o := graph.objects[id]
		if !o.IsAnd() {
			continue
		}
		if _, isMux := graph.MuxControl(id); o.IsXor() || isMux {
			return nil, ErrUnsupported
		}
		a, _ := o.Fanin0(id)
		b, _ := o.Fanin1(id)
		av := values[a.ID()] != a.Complemented()
		bv := values[b.ID()] != b.Complemented()
		if av && bv {
			propagated[id] = max(propagated[a.ID()], propagated[b.ID()])
		} else if !av && !bv {
			propagated[id] = min(propagated[a.ID()], propagated[b.ID()])
		} else if !av {
			propagated[id] = propagated[a.ID()]
		} else {
			propagated[id] = propagated[b.ID()]
		}
	}
	selected := make(map[int]bool)
	var justify func(GiaID) error
	justify = func(id GiaID) error {
		if id == 0 {
			return ErrInvariant
		}
		o := graph.objects[id]
		if o.IsCI() {
			selected[int(o.Diff1())] = true
			return nil
		}
		if !o.IsAnd() || o.IsXor() {
			return ErrUnsupported
		}
		a, _ := o.Fanin0(id)
		b, _ := o.Fanin1(id)
		av := values[a.ID()] != a.Complemented()
		bv := values[b.ID()] != b.Complemented()
		if values[id] {
			if err := justify(a.ID()); err != nil {
				return err
			}
			return justify(b.ID())
		}
		if !av && !bv {
			if propagated[a.ID()] <= propagated[b.ID()] {
				return justify(a.ID())
			}
			return justify(b.ID())
		}
		if !av {
			return justify(a.ID())
		}
		return justify(b.ID())
	}
	if err = justify(driver.ID()); err != nil {
		return nil, err
	}
	result := make([]int, 0, len(selected))
	for i := range assignment {
		if selected[i] {
			result = append(result, i)
		}
	}
	return result, nil
}

type Int2StateLiteral struct {
	Register int
	Value    bool
}
type Int2Cube []Int2StateLiteral

func Int2CoveredRegisterInputs(cover []Int2Cube, registers int) ([]int, error) {
	if registers < 1 {
		return nil, ErrDimension
	}
	if cover == nil {
		return []int{0}, nil
	}
	seen := make([]bool, registers)
	for _, cube := range cover {
		for _, l := range cube {
			if l.Register < 0 || l.Register >= registers {
				return nil, ErrDimension
			}
			seen[l.Register] = true
		}
	}
	r := make([]int, 0)
	for i, v := range seen {
		if v {
			r = append(r, i+1)
		}
	}
	return r, nil
}

func Int2CollectInternal(graph *GiaManager, outputs []int) ([]GiaID, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	seen := make([]bool, graph.ObjectCount())
	seen[0] = true
	for _, id := range graph.cis {
		seen[id] = true
	}
	result := make([]GiaID, 0)
	var walk func(GiaLit) error
	walk = func(l GiaLit) error {
		id := l.ID()
		if seen[id] {
			return nil
		}
		o := graph.objects[id]
		if !o.IsAnd() {
			return ErrInvariant
		}
		a, _ := o.Fanin0(id)
		b, _ := o.Fanin1(id)
		if err := walk(a); err != nil {
			return err
		}
		if err := walk(b); err != nil {
			return err
		}
		if control, isMux := graph.MuxControl(id); isMux {
			if err := walk(control); err != nil {
				return err
			}
		}
		seen[id] = true
		result = append(result, id)
		return nil
	}
	for _, index := range outputs {
		if index < 0 || index >= graph.OutputCount() {
			return nil, ErrDimension
		}
		id := graph.cos[index]
		d, _ := graph.objects[id].Fanin0(id)
		if err := walk(d); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Int2ProblemToGia converts a one-property sequential problem and a state
// cover into the combinational suffix problem used by interpolation.
func Int2ProblemToGia(source *GiaManager, cover []Int2Cube) (*GiaManager, error) {
	if source == nil || source.RegisterCount() == 0 || source.PrimaryOutputCount() != 1 {
		return nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, err
	}
	if _, err := Int2CoveredRegisterInputs(cover, source.RegisterCount()); err != nil {
		return nil, err
	}
	t := NewGiaManager()
	t.Name, t.Spec = source.Name, source.Spec
	mapping := make([]GiaLit, source.ObjectCount())
	mapping[0] = GiaConst0
	for _, id := range source.cis {
		l, e := t.AppendCI()
		if e != nil {
			return nil, e
		}
		mapping[id] = l
	}
	if err := t.HashStart(); err != nil {
		return nil, err
	}
	for id := GiaID(1); int(id) < source.ObjectCount(); id++ {
		if source.objects[id].IsAnd() {
			var e error
			mapping[id], e = int2CopyLogic(source, t, mapping, id)
			if e != nil {
				return nil, e
			}
		}
	}
	sop := GiaConst0
	if cover == nil {
		id := source.cos[0]
		d, _ := source.objects[id].Fanin0(id)
		sop = mapping[d.ID()].NotIf(d.Complemented())
	} else {
		for _, cube := range cover {
			term := GiaConst1
			for _, lit := range cube {
				id := source.cos[source.PrimaryOutputCount()+lit.Register]
				d, _ := source.objects[id].Fanin0(id)
				v := mapping[d.ID()].NotIf(d.Complemented()).NotIf(!lit.Value)
				var e error
				term, e = t.HashAnd(term, v)
				if e != nil {
					return nil, e
				}
			}
			var e error
			sop, e = t.HashOr(sop, term)
			if e != nil {
				return nil, e
			}
		}
	}
	if _, err := t.AppendCO(sop); err != nil {
		return nil, err
	}
	t.HashStop()
	return t, nil
}

func (r Int2Result) ValidateWitness(graph *GiaManager) error {
	if graph == nil || r.Status != Int2Unsafe || r.Counterexample == nil {
		return ErrDimension
	}
	if err := graph.Check(); err != nil {
		return err
	}
	c := r.Counterexample
	if c.Registers() != graph.RegisterCount() || c.Inputs() != graph.PrimaryInputCount() {
		return ErrDimension
	}
	state := make([]bool, c.Registers())
	for i := range state {
		state[i], _ = c.Initial(i)
	}
	for frame := 0; frame < c.Frames(); frame++ {
		input := make([]bool, c.Inputs())
		for i := range input {
			input[i], _ = c.Input(frame, i)
		}
		outputs, next, err := graph.Step(input, state)
		if err != nil {
			return err
		}
		properties := graph.PrimaryOutputCount() - graph.ConstraintCount()
		for i := properties; i < len(outputs); i++ {
			if outputs[i] {
				return ErrInvariant
			}
		}
		if frame == c.LastFrame() {
			if c.Property < 0 || c.Property >= properties || !outputs[c.Property] {
				return ErrInvariant
			}
			return nil
		}
		state = next
	}
	return fmt.Errorf("%w: empty witness", ErrInvariant)
}
