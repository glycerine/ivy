package abc

// Native simulation support for the ABC AIG port.

import "fmt"

// Evaluate computes all COs from CI values, including latch next-state
// outputs. It does not interpret COs as assertions or constraints.
func (m *AigManager) Evaluate(inputs []bool) ([]bool, error) {
	if len(inputs) != len(m.cis) {
		return nil, ErrDimension
	}
	if len(m.objects) == 0 {
		return nil, ErrInvariant
	}
	values := make([]bool, len(m.objects))
	values[0] = true
	for i, id := range m.cis {
		values[id] = inputs[i]
	}
	read := func(l AigLit) bool { return values[l.ID()] != l.Complemented() }
	for i, obj := range m.objects {
		switch obj.Type {
		case AigAND:
			values[i] = read(obj.Fanin0) && read(obj.Fanin1)
		case AigBuffer, AigCO:
			values[i] = read(obj.Fanin0)
		}
	}
	outputs := make([]bool, len(m.cos))
	for i, id := range m.cos {
		outputs[i] = values[id]
	}
	return outputs, nil
}

// Step evaluates the sequential CI/CO partition. It returns independent PO
// and next-state slices and never changes the graph or the supplied state.
func (m *AigManager) Step(primary, state []bool) (outputs, nextState []bool, err error) {
	if len(primary) != m.PrimaryInputCount() || len(state) != m.registers {
		return nil, nil, ErrDimension
	}
	inputs := make([]bool, len(m.cis))
	copy(inputs, primary)
	copy(inputs[len(primary):], state)
	all, err := m.Evaluate(inputs)
	if err != nil {
		return nil, nil, err
	}
	n := m.PrimaryOutputCount()
	return all[:n:n], all[n:], nil
}

// ReplayCounterexample independently steps a supplied trace and confirms that
// the selected PO is true at the last frame. The initial state is taken from
// the trace; the caller must separately check the model's initialization and
// any environment constraints before accepting this as a safety witness.
func (m *AigManager) ReplayCounterexample(c *Counterexample) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if c.registers != m.registers || c.inputs != m.PrimaryInputCount() || c.Property < 0 || c.Property >= m.PrimaryOutputCount() {
		return ErrDimension
	}
	state := make([]bool, c.registers)
	for i := range state {
		state[i] = c.bit(i)
	}
	primary := make([]bool, c.inputs)
	for f := 0; f < c.frames; f++ {
		for i := range primary {
			primary[i] = c.bit(c.registers + f*c.inputs + i)
		}
		outputs, next, err := m.Step(primary, state)
		if err != nil {
			return err
		}
		if f == c.LastFrame() && !outputs[c.Property] {
			return fmt.Errorf("%w: selected output is false at frame %d", ErrInvariant, f)
		}
		state = next
	}
	return nil
}
