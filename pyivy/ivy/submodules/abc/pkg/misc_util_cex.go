package abc

// Derived from ABC src/misc/util/utilCex.c and utilCex.h (Alan Mishchenko).
// Differences from C are documented in PORT_STATUS.md.

import (
	"fmt"
	"math"
	"math/bits"
	"slices"
	"strings"
)

// Counterexample stores initial register values followed by primary input
// values in frame-major order. Within each uint32, the first bit is the LSB.
// Dimensions and storage are private so successful construction guarantees
// bounds for all internal bit operations. Property is a zero-based failing
// output index, or -1 if a transformation requires property rediscovery.
type Counterexample struct {
	Property                            int
	registers, inputs, frames, bitCount int
	data                                []uint32
}

// NewCounterexample constructs a zero-filled trace containing at least one
// frame. Unlike C's intermediate allocation, its last frame is immediately
// consistent with its allocated input data. Counts fit ABC's signed int range.
func NewCounterexample(registers, inputs, frames int) (*Counterexample, error) {
	if frames < 1 {
		return nil, ErrDimension
	}
	n, err := checkedSize(registers, inputs, frames, math.MaxInt32)
	if err != nil {
		return nil, err
	}
	if registers > math.MaxInt32 || inputs > math.MaxInt32 || frames > math.MaxInt32 {
		return nil, ErrLimit
	}
	words, _ := BitWordCount(n, 32)
	return &Counterexample{registers: registers, inputs: inputs, frames: frames, bitCount: n, data: make([]uint32, words)}, nil
}

// NewFullCounterexample sets every logical bit, including initial state, to
// one. Unused tail bits are zero, unlike Abc_CexAllocFull's padding bytes.
func NewFullCounterexample(registers, inputs, frames int) (*Counterexample, error) {
	c, err := NewCounterexample(registers, inputs, frames)
	if err != nil {
		return nil, err
	}
	for i := range c.data {
		c.data[i] = math.MaxUint32
	}
	if tail := c.bitCount % 32; tail != 0 {
		c.data[len(c.data)-1] = (uint32(1) << tail) - 1
	}
	return c, nil
}
func (c *Counterexample) Registers() int { return c.registers }
func (c *Counterexample) Inputs() int    { return c.inputs }
func (c *Counterexample) Frames() int    { return c.frames }
func (c *Counterexample) LastFrame() int { return c.frames - 1 }
func (c *Counterexample) BitCount() int  { return c.bitCount }

// PackedWords returns an independent copy in ABC's 32-bit packed layout.
func (c *Counterexample) PackedWords() []uint32 { return slices.Clone(c.data) }
func (c *Counterexample) bit(index int) bool {
	return c.data[index/32]&(uint32(1)<<uint(index%32)) != 0
}
func (c *Counterexample) set(index int, value bool) {
	mask := uint32(1) << uint(index%32)
	if value {
		c.data[index/32] |= mask
	} else {
		c.data[index/32] &^= mask
	}
}
func (c *Counterexample) Bit(index int) (bool, error) {
	if index < 0 || index >= c.bitCount {
		return false, ErrDimension
	}
	return c.bit(index), nil
}
func (c *Counterexample) SetBit(index int, value bool) error {
	if index < 0 || index >= c.bitCount {
		return ErrDimension
	}
	c.set(index, value)
	return nil
}
func (c *Counterexample) Initial(register int) (bool, error) {
	if register < 0 || register >= c.registers {
		return false, ErrDimension
	}
	return c.bit(register), nil
}
func (c *Counterexample) SetInitial(register int, value bool) error {
	if register < 0 || register >= c.registers {
		return ErrDimension
	}
	c.set(register, value)
	return nil
}
func (c *Counterexample) Input(frame, input int) (bool, error) {
	if frame < 0 || frame >= c.frames || input < 0 || input >= c.inputs {
		return false, ErrDimension
	}
	return c.bit(c.registers + frame*c.inputs + input), nil
}
func (c *Counterexample) SetInput(frame, input int, value bool) error {
	if frame < 0 || frame >= c.frames || input < 0 || input >= c.inputs {
		return ErrDimension
	}
	c.set(c.registers+frame*c.inputs+input, value)
	return nil
}
func (c *Counterexample) CountOnes() int {
	count := 0
	for _, word := range c.data {
		count += bits.OnesCount32(word)
	}
	return count
}

// CounterexampleStatistics is the owned replacement for Abc_CexPrintStats and
// Abc_CexPrintStatsInputs. RealInputs partitions each frame's input bits;
// AuxiliaryInputOnes counts the remaining pseudo-primary inputs.
type CounterexampleStatistics struct {
	Property, LastFrame, Registers, Inputs, Bits, Ones int
	RealInputs, RegisterOnes                           int
	RealInputOnes, AuxiliaryInputOnes                  int
}

func (c *Counterexample) Statistics(realInputs int) (CounterexampleStatistics, error) {
	if e := c.Validate(); e != nil || realInputs < 0 || realInputs > c.inputs {
		if e != nil {
			return CounterexampleStatistics{}, e
		}
		return CounterexampleStatistics{}, ErrDimension
	}
	r := CounterexampleStatistics{Property: c.Property, LastFrame: c.LastFrame(), Registers: c.registers, Inputs: c.inputs, Bits: c.bitCount, RealInputs: realInputs}
	for register := 0; register < c.registers; register++ {
		if c.bit(register) {
			r.RegisterOnes++
		}
	}
	for frame := 0; frame < c.frames; frame++ {
		for input := 0; input < c.inputs; input++ {
			if !c.bit(c.registers + frame*c.inputs + input) {
				continue
			}
			if input < realInputs {
				r.RealInputOnes++
			} else {
				r.AuxiliaryInputOnes++
			}
		}
	}
	r.Ones = r.RegisterOnes + r.RealInputOnes + r.AuxiliaryInputOnes
	return r, nil
}

func counterexamplePercent(count, total int) float64 {
	if total == 0 {
		return 0
	}
	return 100 * float64(count) / float64(total)
}

// FormatCounterexampleStatistics returns deterministic language-neutral
// diagnostics instead of writing process-global stdout.
func FormatCounterexampleStatistics(s CounterexampleStatistics) string {
	inputBits := (s.LastFrame + 1) * s.Inputs
	text := fmt.Sprintf("CEX: property=%d frame=%d registers=%d inputs=%d bits=%d ones=%d (%.2f%%)", s.Property, s.LastFrame, s.Registers, s.Inputs, s.Bits, s.Ones, counterexamplePercent(s.Ones, inputBits))
	if s.RealInputs < s.Inputs {
		realBits := (s.LastFrame + 1) * s.RealInputs
		auxiliaryBits := (s.LastFrame + 1) * (s.Inputs - s.RealInputs)
		text += fmt.Sprintf(" real-ones=%d (%.2f%%) auxiliary-ones=%d (%.2f%%)", s.RealInputOnes, counterexamplePercent(s.RealInputOnes, realBits), s.AuxiliaryInputOnes, counterexamplePercent(s.AuxiliaryInputOnes, auxiliaryBits))
	}
	return text
}

// Format returns the counterexample summary, initial state and every input
// frame. The returned string owns all storage and always ends in a newline.
func (c *Counterexample) Format() (string, error) {
	stats, e := c.Statistics(c.inputs)
	if e != nil {
		return "", e
	}
	var b strings.Builder
	b.WriteString(FormatCounterexampleStatistics(stats))
	b.WriteString("\nstate: ")
	for register := 0; register < c.registers; register++ {
		if c.bit(register) {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}
	b.WriteByte('\n')
	for frame := 0; frame < c.frames; frame++ {
		fmt.Fprintf(&b, "frame %d: ", frame)
		for input := 0; input < c.inputs; input++ {
			if c.bit(c.registers + frame*c.inputs + input) {
				b.WriteByte('1')
			} else {
				b.WriteByte('0')
			}
		}
		b.WriteByte('\n')
	}
	return b.String(), nil
}

// Clone preserves all bits including nonzero initial register values.
func (c *Counterexample) Clone() *Counterexample {
	d := *c
	d.data = slices.Clone(c.data)
	return &d
}

// DupInputs implements Abc_CexDup: copy input frames and zero the initial state
// of a new register bank. Even when the register count is unchanged, C's dup
// discards initial register values. Use Clone to preserve them.
func (c *Counterexample) DupInputs(registers int) (*Counterexample, error) {
	if registers == -1 {
		registers = c.registers
	}
	d, err := NewCounterexample(registers, c.inputs, c.frames)
	if err != nil {
		return nil, err
	}
	d.Property = c.Property
	for i := c.registers; i < c.bitCount; i++ {
		d.set(registers+i-c.registers, c.bit(i))
	}
	return d, nil
}

// CounterexampleFromValues implements Abc_CexCreate. A nil values slice means
// an all-zero trace; skipRegisters leaves the initial register bits zero.
func CounterexampleFromValues(registers, inputs, frames, property int, values []int, skipRegisters bool) (*Counterexample, error) {
	if property < 0 {
		return nil, ErrDimension
	}
	c, err := NewCounterexample(registers, inputs, frames)
	if err != nil {
		return nil, err
	}
	c.Property = property
	offset := 0
	if skipRegisters {
		offset = registers
	}
	if values == nil {
		return c, nil
	}
	if len(values) != c.bitCount-offset {
		return nil, ErrDimension
	}
	for i, v := range values {
		c.set(offset+i, v != 0)
	}
	return c, nil
}

// TrivialCounterexample implements Abc_CexMakeTriv, decoding the combined
// frame/output number. Zero-register circuits are also accepted in Go.
func TrivialCounterexample(registers, inputs, outputs, frameOutput int) (*Counterexample, error) {
	if outputs <= 0 || frameOutput < 0 {
		return nil, ErrDimension
	}
	c, err := NewCounterexample(registers, inputs, frameOutput/outputs+1)
	if err != nil {
		return nil, err
	}
	c.Property = frameOutput % outputs
	return c, nil
}

// PermuteInputs implements Abc_CexPermute's input mapping. oldToNew must be a
// bijection. Like C it zeros the initial register bits; Clone is not implicit.
func (c *Counterexample) PermuteInputs(oldToNew []int) (*Counterexample, error) {
	if err := checkPermutation(oldToNew, c.inputs); err != nil {
		return nil, err
	}
	d, err := NewCounterexample(c.registers, c.inputs, c.frames)
	if err != nil {
		return nil, err
	}
	d.Property = c.Property
	for f := 0; f < c.frames; f++ {
		for old, newIndex := range oldToNew {
			d.set(d.registers+f*d.inputs+newIndex, c.bit(c.registers+f*c.inputs+old))
		}
	}
	return d, nil
}
func checkPermutation(p []int, n int) error {
	if len(p) != n {
		return ErrDimension
	}
	seen := make([]bool, n)
	for _, v := range p {
		if v < 0 || v >= n || seen[v] {
			return fmt.Errorf("%w: not a permutation", ErrDimension)
		}
		seen[v] = true
	}
	return nil
}

// PermuteCanonical implements Abc_CexPermuteTwo.
func (c *Counterexample) PermuteCanonical(oldOrder, newOrder []int) (*Counterexample, error) {
	if err := checkPermutation(oldOrder, c.inputs); err != nil {
		return nil, err
	}
	if err := checkPermutation(newOrder, c.inputs); err != nil {
		return nil, err
	}
	p := make([]int, c.inputs)
	for i, old := range oldOrder {
		p[old] = newOrder[i]
	}
	return c.PermuteInputs(p)
}

// TransformPhase unfolds each packed input frame into its original phases.
// The failing property must be rediscovered (-1), as in Abc_CexTransformPhase.
func (c *Counterexample) TransformPhase(oldInputs, oldOutputs, oldRegisters int) (*Counterexample, error) {
	if oldInputs <= 0 || oldOutputs <= 0 || c.inputs%oldInputs != 0 {
		return nil, ErrDimension
	}
	phases := c.inputs / oldInputs
	if phases == 0 || c.Property < 0 || c.Property/oldOutputs >= phases {
		return nil, ErrDimension
	}
	frames, err := checkedSize(0, c.frames, phases, math.MaxInt32)
	if err != nil {
		return nil, err
	}
	d, err := NewCounterexample(oldRegisters, oldInputs, frames)
	if err != nil {
		return nil, err
	}
	d.Property = -1
	for i := 0; i < c.frames*c.inputs; i++ {
		d.set(oldRegisters+i, c.bit(c.registers+i))
	}
	return d, nil
}

// TransformTemporal implements Abc_CexTransformTempor. Packed frame zero
// contains a prefix in its higher input groups; later frames use group zero.
func (c *Counterexample) TransformTemporal(oldInputs, oldRegisters int) (*Counterexample, error) {
	if oldInputs <= 0 || c.inputs%oldInputs != 0 || c.inputs < oldInputs || c.frames < 2 {
		return nil, ErrDimension
	}
	prefix := c.inputs/oldInputs - 1
	frames, err := checkedSize(c.frames, prefix, 1, math.MaxInt32)
	if err != nil {
		return nil, err
	}
	d, err := NewCounterexample(oldRegisters, oldInputs, frames)
	if err != nil {
		return nil, err
	}
	d.Property = c.Property
	for f := 0; f < prefix; f++ {
		for i := 0; i < oldInputs; i++ {
			d.set(oldRegisters+f*oldInputs+i, c.bit(c.registers+(f+1)*oldInputs+i))
		}
	}
	for f := 0; f < c.frames; f++ {
		for i := 0; i < oldInputs; i++ {
			d.set(oldRegisters+(prefix+f)*oldInputs+i, c.bit(c.registers+f*c.inputs+i))
		}
	}
	return d, nil
}

// TransformUnknownInitial implements Abc_CexTransformUndc. An x/X initial
// latch takes its value from the added inputs at the end of frame zero.
func (c *Counterexample) TransformUnknownInitial(initial string) (*Counterexample, error) {
	added := 0
	for _, ch := range []byte(initial) {
		switch ch {
		case 'x', 'X':
			added++
		case '0', '1':
		default:
			return nil, ErrDimension
		}
	}
	if added > c.inputs {
		return nil, ErrDimension
	}
	d, err := NewCounterexample(len(initial), c.inputs-added, c.frames)
	if err != nil {
		return nil, err
	}
	d.Property = c.Property
	next := 0
	for reg, ch := range []byte(initial) {
		value := ch == '1'
		if ch == 'x' || ch == 'X' {
			value = c.bit(c.registers + c.inputs - added + next)
			next++
		}
		d.set(reg, value)
	}
	for f := 0; f < c.frames; f++ {
		for i := 0; i < d.inputs; i++ {
			d.set(d.registers+f*d.inputs+i, c.bit(c.registers+f*c.inputs+i))
		}
	}
	return d, nil
}

// Merge replaces frames [begin,end) by part's frames before its last frame.
// The C operation overloads part.iPo as an offset; this checked Go operation
// requires that offset to be zero, which is necessary for its copy lengths.
func (c *Counterexample) Merge(part *Counterexample, begin, end int) (*Counterexample, error) {
	if part == nil || begin < 0 || end < begin || end >= c.frames || part.inputs != c.inputs || part.Property != 0 || part.LastFrame() > end-begin {
		return nil, ErrDimension
	}
	frames := c.frames - (end - begin) + part.LastFrame()
	d, err := NewCounterexample(c.registers, c.inputs, frames)
	if err != nil {
		return nil, err
	}
	d.Property = c.Property
	for r := 0; r < c.registers; r++ {
		d.set(r, c.bit(r))
	}
	dest := 0
	copyFrame := func(src *Counterexample, f int) {
		for i := 0; i < c.inputs; i++ {
			d.set(d.registers+dest*d.inputs+i, src.bit(src.registers+f*src.inputs+i))
		}
		dest++
	}
	for f := 0; f < begin; f++ {
		copyFrame(c, f)
	}
	for f := 0; f < part.LastFrame(); f++ {
		copyFrame(part, f)
	}
	for f := end; f < c.frames; f++ {
		copyFrame(c, f)
	}
	return d, nil
}

// Validate checks the dimensions and canonical packed layout of a trace.
// Property -1 is allowed for transformations requiring rediscovery.
func (c *Counterexample) Validate() error {
	if c == nil || c.frames < 1 || c.frames > math.MaxInt32 || c.Property < -1 {
		return ErrDimension
	}
	n, err := checkedSize(c.registers, c.inputs, c.frames, math.MaxInt32)
	if err != nil {
		return err
	}
	words, _ := BitWordCount(n, 32)
	if n != c.bitCount || words != len(c.data) {
		return ErrInvariant
	}
	if tail := n % 32; tail != 0 && c.data[len(c.data)-1]>>tail != 0 {
		return ErrInvariant
	}
	return nil
}

// CounterexampleFromCombModel implements the only C call site, which supplies
// no registers.
func CounterexampleFromCombModel(model []int, property int) (*Counterexample, error) {
	return CounterexampleFromCombModelWithRegisters(model, 0, property)
}

// CounterexampleFromCombModelWithRegisters creates a one-frame trace with a
// zero initial state and model values in the input frame. C accidentally writes
// model bit i at packed bit i, overwriting initial-state positions whenever
// registers is nonzero; Go applies the intended register offset.
func CounterexampleFromCombModelWithRegisters(model []int, registers, property int) (*Counterexample, error) {
	if property < 0 {
		return nil, ErrDimension
	}
	c, e := NewCounterexample(registers, len(model), 1)
	if e != nil {
		return nil, e
	}
	c.Property = property
	for input, value := range model {
		c.set(registers+input, value != 0)
	}
	return c, nil
}
