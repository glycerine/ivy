package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func bmcTestModel(t testing.TB, data string) *AigerModel {
	t.Helper()
	m, err := ReadAiger([]byte(data), AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func bmcTestCheck(t testing.TB, m *AigerModel, frames int) BmcResult {
	t.Helper()
	r, err := m.BoundedCheck(BmcOptions{Frames: frames})
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func TestBmcSmokeAndEarliestFailures(t *testing.T) {
	for _, test := range []struct {
		data  string
		frame int
	}{
		{"aig 2 1 1 1 0\n0\n4\n", -1},
		{"aig 2 1 1 1 0\n2\n4\n", 1},
		{"aig 2 1 1 1 0\n0\n5\n", 0},
		{"aig 1 0 1 1 0\n3\n2\n", 1},
		{"aig 1 0 1 1 0\n2 1\n2\n", 0},
		{"aig 1 0 1 1 0\n2 2\n2\n", 0},
		{"aig 0 0 0 1 0\n1\n", 0},
		{"aig 0 0 0 1 0\n0\n", -1},
	} {
		m := bmcTestModel(t, test.data)
		before := m.graph.Clone()
		for frames := 1; frames <= 4; frames++ {
			r := bmcTestCheck(t, m, frames)
			if test.frame >= 0 && test.frame < frames {
				if r.Status != BmcCounterexample || r.Frame != test.frame || r.Property != 0 || r.Counterexample == nil || r.CheckedFrames != test.frame || r.Reason != BmcWitnessFound {
					t.Fatal(test, r)
				}
				if err := m.ValidateCounterexample(r.Counterexample); err != nil {
					t.Fatal(err)
				}
				trace, err := m.MarshalABCCounterexample(r.Counterexample)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := m.ReadABCCounterexample(trace, r.Property, TraceLimits{}); err != nil {
					t.Fatal(err)
				}
			} else {
				if r.Status != BmcNoCounterexample || r.CheckedFrames != frames || r.Counterexample != nil || r.Property != -1 || r.Reason != BmcBoundReached {
					t.Fatal(test, r)
				}
			}
		}
		if !slices.Equal(m.graph.objects, before.objects) {
			t.Fatal("BMC mutated source graph")
		}
	}
}
func TestBmcConstraintPrefixes(t *testing.T) {
	// At frame zero the bug is valid. Its initial state has no admissible
	// extension past frame zero, so future constraints must not hide the bug.
	m := bmcTestModel(t, "aig 1 0 1 0 0 1 1\n1\n1\n3\n")
	r := bmcTestCheck(t, m, 4)
	if r.Status != BmcCounterexample || r.Frame != 0 {
		t.Fatal("future constraint masked an early failure", r)
	}
	// An impossible frame-zero constraint makes every finite prefix vacuous.
	impossible := bmcTestModel(t, "aig 0 0 0 0 0 1 1\n1\n0\n")
	r = bmcTestCheck(t, impossible, 4)
	if r.Status != BmcNoCounterexample || r.CheckedFrames != 4 {
		t.Fatal("inconsistent prefix", r)
	}
	// The positive assumption x excludes the only bad assignment !x.
	filtered := bmcTestModel(t, "aig 1 1 0 0 0 1 1\n3\n2\n")
	r = bmcTestCheck(t, filtered, 4)
	if r.Status != BmcNoCounterexample {
		t.Fatal("constraint polarity", r)
	}
}
func TestBmcPropertiesAndLimits(t *testing.T) {
	m := bmcTestModel(t, "aig 2 1 1 2 0\n2\n2\n4\n")
	r, err := m.BoundedCheck(BmcOptions{Frames: 3, Properties: []int{1}})
	if err != nil || r.Status != BmcCounterexample || r.Frame != 1 || r.Property != 1 {
		t.Fatal(r, err)
	}
	r, err = m.BoundedCheck(BmcOptions{Frames: 3, Properties: []int{}})
	if err != nil || r.Status != BmcNoCounterexample {
		t.Fatal("explicit empty selection", r, err)
	}
	for _, properties := range [][]int{{-1}, {2}, {0, 0}} {
		if _, err := m.BoundedCheck(BmcOptions{Frames: 1, Properties: properties}); err == nil {
			t.Fatal("invalid property selection", properties)
		}
	}
	for _, options := range []BmcOptions{{Frames: 0}, {Frames: -1}, {Frames: 1, MaxObjects: -1}, {Frames: 1, MaxMappingEntries: -1}, {Frames: 1, MaxVariables: -1}} {
		if _, err := m.BoundedCheck(options); err == nil {
			t.Fatal(options)
		}
	}
	for _, options := range []BmcOptions{{Frames: 3, MaxObjects: 1}, {Frames: 3, MaxMappingEntries: 1}, {Frames: 3, MaxVariables: 1}} {
		r, err := m.BoundedCheck(options)
		if err != nil || r.Status != BmcUnknown || r.Reason != BmcResourceLimit || r.Counterexample != nil {
			t.Fatal("resource result", r, err)
		}
	}
	simple := bmcTestModel(t, "aig 1 1 0 1 0\n2\n")
	r, err = simple.BoundedCheck(BmcOptions{Frames: 2, SAT: SatLimits{Propagations: 1}})
	if err != nil || r.Status != BmcUnknown || r.Reason != BmcSatLimit || r.SatReason != SatPropagationLimit || r.Counterexample != nil || r.CheckedFrames != 0 {
		t.Fatal("SAT budget", r, err)
	}
	for _, threshold := range []int{1, 3, 10, 20} {
		calls := 0
		r, err := m.BoundedCheck(BmcOptions{Frames: 10, SAT: SatLimits{Stop: func() bool { calls++; return calls >= threshold }}})
		if err != nil || r.Status != BmcUnknown || r.Reason != BmcCancelled || r.SatReason != SatCancelled {
			t.Fatal("cancellation", threshold, r, err)
		}
	}
	if again := bmcTestCheck(t, simple, 2); again.Status != BmcCounterexample {
		t.Fatal("interrupted run affected later calls")
	}
	both := bmcTestModel(t, "aig 0 0 0 2 0\n1\n1\n")
	r, err = both.BoundedCheck(BmcOptions{Frames: 1, Properties: []int{1, 0}})
	if err != nil || r.Property != 1 {
		t.Fatal("property tie order", r, err)
	}
	empty := bmcTestModel(t, "aig 0 0 0 0 0\n")
	r = bmcTestCheck(t, empty, 2)
	if r.Status != BmcNoCounterexample {
		t.Fatal(r)
	}
}

// bmcTruthCase defines a complete three-variable transition relation without
// invoking graph simulation or CNF: one PI, two state bits, arbitrary output,
// next-state and constraint truth masks.
type bmcTruthCase struct {
	bad, next0, next1, constraint uint8
	initial                       [2]LatchInit
}

func bmcTruthExpected(c bmcTruthCase, frames int) int {
	states := make([]bool, 4)
	for state := 0; state < 4; state++ {
		valid := true
		for r, init := range c.initial {
			value := state&(1<<r) != 0
			if (init == LatchInitZero && value) || (init == LatchInitOne && !value) {
				valid = false
			}
		}
		states[state] = valid
	}
	for frame := 0; frame < frames; frame++ {
		next := make([]bool, 4)
		for state, reachable := range states {
			if reachable {
				for input := 0; input < 2; input++ {
					index := 2*state + input
					if c.constraint&(1<<index) == 0 {
						continue
					}
					if c.bad&(1<<index) != 0 {
						return frame
					}
					target := 0
					if c.next0&(1<<index) != 0 {
						target |= 1
					}
					if c.next1&(1<<index) != 0 {
						target |= 2
					}
					next[target] = true
				}
			}
		}
		states = next
	}
	return -1
}
func bmcTruthModel(t testing.TB, c bmcTruthCase) *AigerModel {
	t.Helper()
	g := NewAigManager()
	variables := []AigLit{aigTestCI(t, g), aigTestCI(t, g), aigTestCI(t, g)}
	synthesize := func(table uint8) AigLit {
		result := AigConst0
		for assignment := 0; assignment < 8; assignment++ {
			if table&(1<<assignment) != 0 {
				cube := AigConst1
				for i, l := range variables {
					if assignment&(1<<i) == 0 {
						l = l.Not()
					}
					cube = aigTestAnd(t, g, cube, l)
				}
				var err error
				result, err = g.Or(result, cube)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
		return result
	}
	for _, table := range []uint8{c.bad, c.constraint, c.next0, c.next1} {
		aigTestCO(t, g, synthesize(table))
	}
	if err := g.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	m, err := NewAigerModel(g, 1, 0, 1, c.initial[:])
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func bmcVerifyTruthCase(t testing.TB, c bmcTruthCase, frames int) {
	t.Helper()
	model := bmcTruthModel(t, c)
	expected := bmcTruthExpected(c, frames)
	answer := bmcTestCheck(t, model, frames)
	if expected < 0 {
		if answer.Status != BmcNoCounterexample || answer.CheckedFrames != frames {
			t.Fatal("independent finite reachability", c, frames, expected, answer)
		}
	} else {
		if answer.Status != BmcCounterexample || answer.Frame != expected || answer.Counterexample == nil {
			t.Fatal("independent finite reachability", c, frames, expected, answer)
		}
		trace := answer.Counterexample
		state := 0
		for i := 0; i < 2; i++ {
			v, _ := trace.Initial(i)
			if v {
				state |= 1 << i
			}
		}
		for frame := 0; frame < trace.Frames(); frame++ {
			input, _ := trace.Input(frame, 0)
			index := 2 * state
			if input {
				index++
			}
			if c.constraint&(1<<index) == 0 {
				t.Fatal("witness violates independent constraint")
			}
			if frame == trace.LastFrame() && c.bad&(1<<index) == 0 {
				t.Fatal("witness misses independent failure")
			}
			next := 0
			if c.next0&(1<<index) != 0 {
				next |= 1
			}
			if c.next1&(1<<index) != 0 {
				next |= 2
			}
			state = next
		}
	}
}
func TestBmcIndependentSmallMachines(t *testing.T) {
	// Deterministic sequence covers 256 unrelated small truth-table machines.
	state := uint32(0x739a63c1)
	next := func() uint8 { state ^= state << 13; state ^= state >> 17; state ^= state << 5; return uint8(state) }
	for i := 0; i < 256; i++ {
		c := bmcTruthCase{bad: next(), next0: next(), next1: next(), constraint: next(), initial: [2]LatchInit{LatchInit(next() % 3), LatchInit(next() % 3)}}
		bmcVerifyTruthCase(t, c, 4)
	}
}
func TestBmcNativeFixtures(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/bmc.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Bad, Next0, Next1 uint8
			Queries           []struct{ Frames, ReturnCode, Frame int }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 64 {
		t.Fatal("native BMC fixture size")
	}
	for _, c := range fixture.Cases {
		spec := bmcTruthCase{bad: c.Bad, next0: c.Next0, next1: c.Next1, constraint: 255}
		model := bmcTruthModel(t, spec)
		for _, query := range c.Queries {
			expected := bmcTruthExpected(spec, query.Frames)
			answer := bmcTestCheck(t, model, query.Frames)
			if expected >= 0 {
				if query.ReturnCode != 0 || query.Frame != expected || answer.Status != BmcCounterexample || answer.Frame != expected {
					t.Fatal("native bounded failure disagreement", c, query, answer, expected)
				}
			} else if query.ReturnCode != -1 || query.Frame != query.Frames || answer.Status != BmcNoCounterexample {
				t.Fatal("native bounded exclusion disagreement", c, query, answer)
			}
		}
	}
}
func FuzzBmcSmallMachines(f *testing.F) {
	for _, seed := range [][]byte{{0, 0, 0, 255, 0, 0, 4}, {255, 0, 0, 255, 0, 0, 2}, {128, 170, 204, 240, 2, 1, 4}, {1, 255, 0, 1, 0, 0, 3}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 7 || len(data) > 64 {
			return
		}
		c := bmcTruthCase{bad: data[0], next0: data[1], next1: data[2], constraint: data[3], initial: [2]LatchInit{LatchInit(data[4] % 3), LatchInit(data[5] % 3)}}
		bmcVerifyTruthCase(t, c, 1+int(data[6]%5))
	})
}
