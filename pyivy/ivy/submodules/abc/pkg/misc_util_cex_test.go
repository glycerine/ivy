package abc

import (
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func cexTestNew(t testing.TB, r, i, f int) *Counterexample {
	t.Helper()
	c, err := NewCounterexample(r, i, f)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func cexBits(t testing.TB, c *Counterexample) []bool {
	t.Helper()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	result := make([]bool, c.BitCount())
	for i := range result {
		result[i], _ = c.Bit(i)
	}
	return result
}
func TestCounterexamplePackedBoundaries(t *testing.T) {
	for _, r := range []int{0, 1, 31, 32, 33, 63, 64} {
		for _, n := range []int{0, 1, 31, 32, 33} {
			for _, f := range []int{1, 2, 3} {
				c := cexTestNew(t, r, n, f)
				if c.Registers() != r || c.Inputs() != n || c.Frames() != f || c.LastFrame() != f-1 || c.BitCount() != r+n*f {
					t.Fatal("dimensions")
				}
				expected := make([]bool, c.BitCount())
				for i := range expected {
					expected[i] = i%3 == 0
					if err := c.SetBit(i, expected[i]); err != nil {
						t.Fatal(err)
					}
				}
				if !slices.Equal(cexBits(t, c), expected) {
					t.Fatal("bits")
				}
				count := 0
				for _, b := range expected {
					if b {
						count++
					}
				}
				if c.CountOnes() != count {
					t.Fatal("count")
				}
				words := c.PackedWords()
				for i, b := range expected {
					if (words[i/32]&(1<<uint(i%32)) != 0) != b {
						t.Fatal("packing")
					}
				}
				if len(words) > 0 {
					words[0] ^= math.MaxUint32
					if !slices.Equal(cexBits(t, c), expected) {
						t.Fatal("aliased packed words")
					}
				}
				full, err := NewFullCounterexample(r, n, f)
				if err != nil || full.CountOnes() != full.BitCount() || full.Validate() != nil {
					t.Fatal("full", err)
				}
				for i := range expected {
					if err := full.SetBit(i, false); err != nil {
						t.Fatal(err)
					}
				}
				if full.CountOnes() != 0 {
					t.Fatal("clear")
				}
			}
		}
	}
}
func TestCounterexampleValidation(t *testing.T) {
	for _, v := range [][3]int{{-1, 0, 1}, {0, -1, 1}, {0, 0, 0}, {0, 1, -1}, {math.MaxInt, 1, 1}, {1, math.MaxInt, math.MaxInt}, {0, 0, math.MaxInt}} {
		if _, err := NewCounterexample(v[0], v[1], v[2]); err == nil {
			t.Fatal("accepted", v)
		}
	}
	var nilC *Counterexample
	if nilC.Validate() == nil || new(Counterexample).Validate() == nil {
		t.Fatal("invalid zero trace")
	}
	c := cexTestNew(t, 1, 2, 3)
	for _, i := range []int{-1, c.BitCount(), math.MaxInt} {
		if c.SetBit(i, true) == nil {
			t.Fatal(i)
		}
		if _, err := c.Bit(i); err == nil {
			t.Fatal(i)
		}
	}
	for _, v := range [][2]int{{-1, 0}, {0, -1}, {3, 0}, {0, 2}, {math.MaxInt, math.MaxInt}} {
		if c.SetInput(v[0], v[1], true) == nil {
			t.Fatal(v)
		}
		if _, err := c.Input(v[0], v[1]); err == nil {
			t.Fatal(v)
		}
	}
	if c.SetInitial(1, true) == nil {
		t.Fatal("register range")
	}
	c.data[0] |= 1 << 31
	if c.Validate() == nil {
		t.Fatal("dirty tail")
	}
}
func TestCounterexampleCopyAndConstructors(t *testing.T) {
	values := []int{1, 0, 2, 0, -1, 0, 1, 0}
	c, err := CounterexampleFromValues(2, 2, 3, 4, values, false)
	if err != nil {
		t.Fatal(err)
	}
	clone := c.Clone()
	clone.SetBit(0, false)
	if !c.bit(0) || clone.bit(0) {
		t.Fatal("clone aliases")
	}
	for _, r := range []int{-1, 0, 1, 2, 40} {
		d, err := c.DupInputs(r)
		if err != nil {
			t.Fatal(err)
		}
		if r == -1 {
			r = c.registers
		}
		expected := append(make([]bool, r), cexBits(t, c)[c.registers:]...)
		if !slices.Equal(cexBits(t, d), expected) || d.Property != c.Property {
			t.Fatal("dup", r)
		}
	}
	skip, err := CounterexampleFromValues(2, 2, 3, 4, values[2:], true)
	if err != nil {
		t.Fatal(err)
	}
	if skip.bit(0) || skip.bit(1) || !slices.Equal(cexBits(t, c)[2:], cexBits(t, skip)[2:]) {
		t.Fatal("skip regs")
	}
	if _, err := CounterexampleFromValues(0, 2, 1, 0, []int{1}, false); err == nil {
		t.Fatal("length mismatch")
	}
	if _, err := CounterexampleFromValues(0, 0, 1, -1, nil, false); err == nil {
		t.Fatal("negative property")
	}
	trivial, err := TrivialCounterexample(0, 3, 4, 14)
	if err != nil || trivial.Property != 2 || trivial.LastFrame() != 3 || trivial.CountOnes() != 0 {
		t.Fatal(trivial, err)
	}
	comb, err := CounterexampleFromCombModel([]int{0, 2, -1}, 1)
	if err != nil || !slices.Equal(cexBits(t, comb), []bool{false, true, true}) {
		t.Fatal(comb, err)
	}
	registered, err := CounterexampleFromCombModelWithRegisters([]int{1, 0, 1}, 2, 4)
	if err != nil || registered.Property != 4 || !slices.Equal(cexBits(t, registered), []bool{false, false, true, false, true}) {
		t.Fatal(registered, err)
	}
	if _, err = CounterexampleFromCombModelWithRegisters(nil, 0, -1); err == nil {
		t.Fatal("negative property")
	}
}

func TestCounterexampleStatisticsAndFormat(t *testing.T) {
	c, err := CounterexampleFromValues(2, 4, 2, 3, []int{1, 0, 1, 0, 1, 1, 0, 0, 1, 0}, false)
	if err != nil {
		t.Fatal(err)
	}
	stats, err := c.Statistics(2)
	want := CounterexampleStatistics{Property: 3, LastFrame: 1, Registers: 2, Inputs: 4, Bits: 10, Ones: 5, RealInputs: 2, RegisterOnes: 1, RealInputOnes: 1, AuxiliaryInputOnes: 3}
	if err != nil || stats != want {
		t.Fatal(stats, want, err)
	}
	if text := FormatCounterexampleStatistics(stats); text != "CEX: property=3 frame=1 registers=2 inputs=4 bits=10 ones=5 (62.50%) real-ones=1 (25.00%) auxiliary-ones=3 (75.00%)" {
		t.Fatal(text)
	}
	text, err := c.Format()
	if err != nil || text != "CEX: property=3 frame=1 registers=2 inputs=4 bits=10 ones=5 (62.50%)\nstate: 10\nframe 0: 1011\nframe 1: 0010\n" {
		t.Fatal(text, err)
	}
	if _, err = c.Statistics(5); err == nil {
		t.Fatal("real-input range")
	}
	if text := FormatCounterexampleStatistics(CounterexampleStatistics{}); strings.Contains(text, "NaN") || strings.Contains(text, "Inf") {
		t.Fatal(text)
	}
}
func TestCounterexamplePermutation(t *testing.T) {
	c, err := CounterexampleFromValues(1, 3, 2, 0, []int{1, 1, 0, 0, 0, 1, 0}, false)
	if err != nil {
		t.Fatal(err)
	}
	d, err := c.PermuteInputs([]int{2, 0, 1})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cexBits(t, d), []bool{false, false, false, true, true, false, false}) {
		t.Fatal(cexBits(t, d))
	}
	e, err := c.PermuteCanonical([]int{1, 2, 0}, []int{0, 1, 2})
	if err != nil || !slices.Equal(cexBits(t, d), cexBits(t, e)) {
		t.Fatal(e, err)
	}
	for _, p := range [][]int{{}, {0, 0, 2}, {0, 1, 3}, {0, 1, -1}} {
		if _, err := c.PermuteInputs(p); err == nil {
			t.Fatal(p)
		}
	}
}
func TestCounterexampleTransforms(t *testing.T) {
	c, err := CounterexampleFromValues(1, 4, 2, 3, []int{1, 1, 0, 0, 1, 0, 1, 1, 0}, false)
	if err != nil {
		t.Fatal(err)
	}
	phase, err := c.TransformPhase(2, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []bool{false, false, false, true, false, false, true, false, true, true, false}
	if !slices.Equal(cexBits(t, phase), want) || phase.Property != -1 || phase.Frames() != 4 {
		t.Fatal(cexBits(t, phase))
	}
	temporal, err := c.TransformTemporal(2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cexBits(t, temporal), []bool{false, false, false, true, true, false, false, true}) || temporal.Frames() != 3 {
		t.Fatal(cexBits(t, temporal))
	}
	unknown, err := c.TransformUnknownInitial("1xX0")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cexBits(t, unknown), []bool{true, false, true, false, true, false, false, true}) {
		t.Fatal(cexBits(t, unknown))
	}
	for _, bad := range []string{"?", "xxxxxxx"} {
		if _, err := c.TransformUnknownInitial(bad); err == nil {
			t.Fatal(bad)
		}
	}
	if _, err := c.TransformPhase(0, 2, 1); err == nil {
		t.Fatal("zero original inputs")
	}
	if _, err := c.TransformTemporal(3, 1); err == nil {
		t.Fatal("nondivisible inputs")
	}
}
func TestCounterexampleMerge(t *testing.T) {
	c, err := CounterexampleFromValues(1, 2, 5, 2, []int{1, 1, 0, 0, 1, 1, 1, 0, 0, 1, 0}, false)
	if err != nil {
		t.Fatal(err)
	}
	part, err := CounterexampleFromValues(0, 2, 2, 0, []int{0, 1, 1, 0}, false)
	if err != nil {
		t.Fatal(err)
	}
	merged, err := c.Merge(part, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cexBits(t, merged), []bool{true, true, false, false, true, false, false, true, false}) || merged.LastFrame() != 3 || merged.Property != 2 {
		t.Fatal(cexBits(t, merged))
	}
	for _, v := range [][2]int{{-1, 1}, {2, 1}, {0, 5}, {1, 1}} {
		if _, err := c.Merge(part, v[0], v[1]); err == nil {
			t.Fatal(v)
		}
	}
	part.Property = 1
	if _, err := c.Merge(part, 1, 3); err == nil {
		t.Fatal("nonzero overloaded offset")
	}
}

func FuzzCounterexampleLayout(f *testing.F) {
	f.Add([]byte{1, 3, 2, 1, 0, 1, 1, 1, 0})
	f.Add([]byte{0, 0, 0})
	f.Add([]byte{33, 31, 4, 255, 0, 255})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		regs, inputs, frames := int(data[0]%65), int(data[1]%33), 1+int(data[2]%5)
		c := cexTestNew(t, regs, inputs, frames)
		model := make([]bool, c.BitCount())
		for i := range model {
			if len(data) > 3 {
				model[i] = data[3+i%(len(data)-3)]&1 != 0
			}
			if err := c.SetBit(i, model[i]); err != nil {
				t.Fatal(err)
			}
		}
		if !slices.Equal(cexBits(t, c), model) {
			t.Fatal("packed mismatch")
		}
		d, err := c.DupInputs(regs + 1)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(cexBits(t, d), append(make([]bool, regs+1), model[regs:]...)) {
			t.Fatal("dup mismatch")
		}
		p := make([]int, inputs)
		for i := range p {
			p[i] = inputs - 1 - i
		}
		d, err = c.PermuteInputs(p)
		if err != nil {
			t.Fatal(err)
		}
		for frame := 0; frame < frames; frame++ {
			for i := 0; i < inputs; i++ {
				got, _ := d.Input(frame, inputs-1-i)
				if got != model[regs+frame*inputs+i] {
					t.Fatal("permutation mismatch")
				}
			}
		}
		clone := c.Clone()
		if !slices.Equal(cexBits(t, clone), model) {
			t.Fatal("clone mismatch")
		}
		realInputs := 0
		if inputs > 0 {
			realInputs = int(data[0]) % (inputs + 1)
		}
		stats, err := c.Statistics(realInputs)
		if err != nil || stats.Ones != c.CountOnes() || stats.RealInputOnes+stats.AuxiliaryInputOnes+stats.RegisterOnes != stats.Ones {
			t.Fatal("statistics", stats, err)
		}
		if _, err = c.Format(); err != nil {
			t.Fatal("format", err)
		}
	})
}
