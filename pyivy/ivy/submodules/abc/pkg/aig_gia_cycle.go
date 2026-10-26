package abc

// Prefix simulation and phase recoding from src/aig/gia/giaDup.c:
// Gia_ManCycle/Gia_ManDupCycled; local RNG from giaUtil.c:Gia_ManRandom.

// GiaCycleOptions bounds simulated frames and graph/map allocation.
type GiaCycleOptions struct {
	GiaCofactorOptions
	MaxFrames int
}

// GiaCycledGraph contains the reached source state. The returned graph encodes
// newState = sourceState XOR Reached, so zero target state represents Reached.
type GiaCycledGraph struct {
	Graph   *GiaManager
	Reached []bool
}

// CycleFromZero simulates frames transitions from all-zero registers. A nil
// trace selects C's reset random input stream (one RNG value per PI per frame).
// A supplied trace must match CI state/PI dimensions; frames <= LastFrame as
// in C. Stored initial bits and failing-property metadata do not affect this
// prefix simulation. No property/constraint checks or proof claims are made.
// All raw gates, terminal/reset order and name/spec survive; constraints clear
// as in C. Source and trace scratch remain unchanged. Extended gates simulate
// their actual Boolean semantics, extending C's ordinary-AND-only simulator.
func (m *GiaManager) CycleFromZero(trace *Counterexample, frames int, options GiaCycleOptions) (GiaCycledGraph, error) {
	fail := func(e error) (GiaCycledGraph, error) { return GiaCycledGraph{}, e }
	if m == nil || frames < 0 || options.MaxFrames < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	capFrames := options.MaxFrames
	if capFrames == 0 {
		capFrames = 1 << 20
	}
	if frames > capFrames {
		return fail(ErrLimit)
	}
	capMap := options.MaxMappingEntries
	if capMap == 0 {
		capMap = 1 << 20
	}
	capObjects := options.MaxObjects
	if capObjects == 0 {
		capObjects = 1 << 20
	}
	if len(m.objects) > capMap || len(m.objects) > capObjects {
		return fail(ErrLimit)
	}
	if trace != nil && (trace.registers != m.registers || trace.inputs != m.PrimaryInputCount() || frames > trace.LastFrame()) {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	state := make([]bool, m.registers)
	inputs := make([]bool, m.PrimaryInputCount())
	z, w := uint32(3716960521), uint32(2174103536)
	random := func() uint32 { z = 36969*(z&65535) + (z >> 16); w = 18000*(w&65535) + (w >> 16); return (z << 16) + w }
	random() // Gia_ManRandom(1) both resets and advances once.
	for frame := 0; frame < frames; frame++ {
		if stopped() {
			return fail(ErrCancelled)
		}
		for i := range inputs {
			if trace == nil {
				inputs[i] = random()&1 != 0
			} else {
				inputs[i] = trace.bit(trace.registers + frame*trace.inputs + i)
			}
		}
		_, next, e := m.Step(inputs, state)
		if e != nil {
			return fail(e)
		}
		state = next
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	g, e := m.FlipRegisters(state, options.GiaCofactorOptions)
	if e != nil {
		return fail(e)
	}
	g.constraints = 0
	return GiaCycledGraph{g, state}, nil
}
