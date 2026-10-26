package abc

// Derived from ABC src/base/io/ioWriteAiger.c: Io_WriteAigerCex.
// Checked trace reading and model validation are native Go additions.

import (
	"bytes"
	"fmt"
	"slices"
)

// ValidateCounterexample checks initialization, every environment constraint,
// and the selected ordinary/bad PO at the last frame, on the original model.
// Constraint outputs are raw AIGER positive assumptions, not ABC's inverted
// constraint signals. Constraint outputs cannot be selected as properties.
func (m *AigerModel) ValidateCounterexample(c *Counterexample) error {
	if err := m.Validate(); err != nil {
		return err
	}
	if err := c.Validate(); err != nil {
		return err
	}
	g := m.graph
	if c.registers != len(m.initial) || c.inputs != g.PrimaryInputCount() || c.Property < 0 || c.Property >= m.outputs+m.bad {
		return ErrDimension
	}
	state := make([]bool, c.registers)
	for i, init := range m.initial {
		state[i] = c.bit(i)
		if (init == LatchInitZero && state[i]) || (init == LatchInitOne && !state[i]) {
			return fmt.Errorf("%w: initial register %d", ErrInvariant, i)
		}
	}
	primary := make([]bool, c.inputs)
	for frame := 0; frame < c.frames; frame++ {
		for i := range primary {
			primary[i] = c.bit(c.registers + frame*c.inputs + i)
		}
		outputs, next, err := g.Step(primary, state)
		if err != nil {
			return err
		}
		for i := m.outputs + m.bad; i < len(outputs); i++ {
			if !outputs[i] {
				return fmt.Errorf("%w: constraint %d is false at frame %d", ErrInvariant, i-m.outputs-m.bad, frame)
			}
		}
		if frame == c.LastFrame() && !outputs[c.Property] {
			return fmt.Errorf("%w: property %d is false at frame %d", ErrInvariant, c.Property, frame)
		}
		state = next
	}
	return nil
}

func appendTraceBits(result []byte, values []bool) []byte {
	for _, v := range values {
		if v {
			result = append(result, '1')
		} else {
			result = append(result, '0')
		}
	}
	return result
}

// MarshalABCCounterexample produces the four-vector trace format consumed
// by Ivy: result line 1, then state/input/output/next-state rows separated by
// single spaces. With zero initialization it matches Io_WriteAigerCex's format.
// Nonzero initialized states are preserved, rather than silently printed as 0.
func (m *AigerModel) MarshalABCCounterexample(c *Counterexample) ([]byte, error) {
	if err := m.ValidateCounterexample(c); err != nil {
		return nil, err
	}
	state := make([]bool, c.registers)
	primary := make([]bool, c.inputs)
	for i := range state {
		state[i] = c.bit(i)
	}
	result := []byte("1\n")
	for frame := 0; frame < c.frames; frame++ {
		for i := range primary {
			primary[i] = c.bit(c.registers + frame*c.inputs + i)
		}
		outputs, next, err := m.graph.Step(primary, state)
		if err != nil {
			return nil, err
		}
		result = appendTraceBits(result, state)
		result = append(result, ' ')
		result = appendTraceBits(result, primary)
		result = append(result, ' ')
		result = appendTraceBits(result, outputs)
		result = append(result, ' ')
		result = appendTraceBits(result, next)
		result = append(result, '\n')
		state = next
	}
	return result, nil
}

// TraceLimits bounds the untrusted four-vector witness parser.
type TraceLimits struct{ MaxBytes, MaxFrames int }

func (l TraceLimits) normalized() (TraceLimits, error) {
	if l.MaxBytes < 0 || l.MaxFrames < 0 {
		return l, ErrDimension
	}
	if l.MaxBytes == 0 {
		l.MaxBytes = 32 << 20
	}
	if l.MaxFrames == 0 {
		l.MaxFrames = 1 << 20
	}
	return l, nil
}
func parseTraceBits(data []byte, values []bool) error {
	if len(data) != len(values) {
		return ErrDimension
	}
	for i, b := range data {
		if b != '0' && b != '1' {
			return ErrFormat
		}
		values[i] = b == '1'
	}
	return nil
}

// ReadABCCounterexample verifies every vector of an Ivy/ABC four-vector trace.
// It accepts grounded 0/1 traces only. The selected property is explicit,
// since this historical ABC trace format does not serialize its property ID.
func (m *AigerModel) ReadABCCounterexample(data []byte, property int, limits TraceLimits) (*Counterexample, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	limits, err := limits.normalized()
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if property < 0 || property >= m.outputs+m.bad {
		return nil, ErrDimension
	}
	cur := aigerCursor{data: data}
	header, err := cur.line()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(header, []byte("1")) {
		return nil, cur.bad("expected counterexample status 1")
	}
	frames := bytes.Count(data[cur.pos:], []byte{'\n'})
	if frames < 1 {
		return nil, cur.bad("missing trace frames")
	}
	if frames > limits.MaxFrames {
		return nil, ErrLimit
	}
	g := m.graph
	// Each complete row has two state vectors, PIs, POs, three spaces and LF.
	minimum := 2*uint64(g.RegisterCount()) + uint64(g.PrimaryInputCount()) + uint64(g.PrimaryOutputCount()) + 4
	if minimum*uint64(frames) > uint64(len(data)-cur.pos) {
		return nil, cur.bad("trace vectors exceed remaining bytes")
	}
	c, err := NewCounterexample(g.RegisterCount(), g.PrimaryInputCount(), frames)
	if err != nil {
		return nil, err
	}
	c.Property = property
	state := make([]bool, c.registers)
	pre := make([]bool, c.registers)
	primary := make([]bool, c.inputs)
	claimedOutputs := make([]bool, g.PrimaryOutputCount())
	claimedNext := make([]bool, c.registers)
	for frame := 0; frame < frames; frame++ {
		line, err := cur.line()
		if err != nil {
			return nil, err
		}
		columns := bytes.SplitN(line, []byte{' '}, 5)
		if len(columns) != 4 {
			return nil, cur.bad("expected four trace vectors")
		}
		if err := parseTraceBits(columns[0], pre); err != nil {
			return nil, err
		}
		if frame == 0 {
			copy(state, pre)
			for i, v := range pre {
				c.set(i, v)
			}
		} else if !slices.Equal(pre, state) {
			return nil, cur.bad("trace state discontinuity")
		}
		if err := parseTraceBits(columns[1], primary); err != nil {
			return nil, err
		}
		if err := parseTraceBits(columns[2], claimedOutputs); err != nil {
			return nil, err
		}
		if err := parseTraceBits(columns[3], claimedNext); err != nil {
			return nil, err
		}
		for i, v := range primary {
			c.set(c.registers+frame*c.inputs+i, v)
		}
		outputs, next, err := g.Step(primary, state)
		if err != nil {
			return nil, err
		}
		if !slices.Equal(outputs, claimedOutputs) || !slices.Equal(next, claimedNext) {
			return nil, cur.bad("trace simulation disagrees with vectors")
		}
		state = next
	}
	if cur.pos != len(data) {
		return nil, cur.bad("trailing trace bytes")
	}
	if err := m.ValidateCounterexample(c); err != nil {
		return nil, err
	}
	return c, nil
}
