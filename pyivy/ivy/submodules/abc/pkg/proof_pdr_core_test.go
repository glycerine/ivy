package abc

import (
	"encoding/json"
	"os"
	"testing"
)

// Four states suffice for complete reachability: a shortest failing path
// visits at most four distinct states, including its failing state.
func pdrVerifyTruthCase(t testing.TB, c bmcTruthCase, disable bool) {
	pdrVerifyTruthOptions(t, c, PdrOptions{MaxFrames: 16, MaxQueries: 10000, DisableGeneralization: disable})
}
func pdrVerifyTruthOptions(t testing.TB, c bmcTruthCase, options PdrOptions) {
	t.Helper()
	m := bmcTruthModel(t, c)
	r, err := m.Prove(options)
	if err != nil {
		t.Fatalf("case %+v: %v", c, err)
	}
	bad := bmcTruthExpected(c, 4) >= 0
	if bad {
		if r.Status != PdrCounterexample || r.Counterexample == nil {
			t.Fatalf("reachable failure %+v: %+v", c, r)
		}
		if err := m.ValidateCounterexample(r.Counterexample); err != nil {
			t.Fatal(err)
		}
		trace := r.Counterexample
		state := 0
		for bit, init := range c.initial {
			v, err := trace.Initial(bit)
			if err != nil {
				t.Fatal(err)
			}
			if (init == LatchInitZero && v) || (init == LatchInitOne && !v) {
				t.Fatal("wrong initial state")
			}
			if v {
				state |= 1 << bit
			}
		}
		for frame := 0; frame < trace.Frames(); frame++ {
			input, err := trace.Input(frame, 0)
			if err != nil {
				t.Fatal(err)
			}
			index := 2 * state
			if input {
				index++
			}
			if c.constraint&(1<<index) == 0 {
				t.Fatal("inadmissible witness")
			}
			if frame == trace.LastFrame() && c.bad&(1<<index) == 0 {
				t.Fatal("nonfailing witness")
			}
			state = 0
			if c.next0&(1<<index) != 0 {
				state |= 1
			}
			if c.next1&(1<<index) != 0 {
				state |= 2
			}
		}
		return
	}
	if r.Status != PdrProved || r.Invariant == nil {
		t.Fatalf("safe case %+v: %+v", c, r)
	}
	// Independently enumerate initiation, induction and safety without graph
	// simulation, CNF, or SAT, so a shared encoding bug cannot validate a proof.
	for state := 0; state < 4; state++ {
		bits := []bool{state&1 != 0, state&2 != 0}
		holds, err := r.Invariant.Holds(bits)
		if err != nil {
			t.Fatal(err)
		}
		initial := true
		for bit, init := range c.initial {
			if (init == LatchInitZero && bits[bit]) || (init == LatchInitOne && !bits[bit]) {
				initial = false
			}
		}
		for input := 0; input < 2; input++ {
			index := 2*state + input
			if c.constraint&(1<<index) == 0 {
				continue
			}
			if initial && !holds {
				t.Fatal("certificate excludes admissible initial state", c, state)
			}
			if !holds {
				continue
			}
			if c.bad&(1<<index) != 0 {
				t.Fatal("certificate permits failure", c, state)
			}
			next := []bool{c.next0&(1<<index) != 0, c.next1&(1<<index) != 0}
			preserved, err := r.Invariant.Holds(next)
			if err != nil || !preserved {
				t.Fatal("certificate is not inductive", c, state, err)
			}
		}
	}
}
func TestPdrIndependentSmallMachines(t *testing.T) {
	state := uint32(0x92461bad)
	next := func() uint8 { state ^= state << 13; state ^= state >> 17; state ^= state << 5; return uint8(state) }
	for i := 0; i < 2048; i++ {
		c := bmcTruthCase{bad: next(), next0: next(), next1: next(), constraint: next(), initial: [2]LatchInit{LatchInit(next() % 3), LatchInit(next() % 3)}}
		pdrVerifyTruthCase(t, c, false)
		if i < 128 {
			pdrVerifyTruthCase(t, c, true)
		}
	}
}
func FuzzPdrReachability(f *testing.F) {
	f.Add(uint8(240), uint8(170), uint8(204), uint8(255), uint8(0), uint8(0))
	f.Add(uint8(0), uint8(173), uint8(56), uint8(255), uint8(2), uint8(2))
	f.Fuzz(func(t *testing.T, bad, n0, n1, env, i0, i1 uint8) {
		pdrVerifyTruthCase(t, bmcTruthCase{bad: bad, next0: n0, next1: n1, constraint: env, initial: [2]LatchInit{LatchInit(i0 % 3), LatchInit(i1 % 3)}}, false)
	})
}

func TestPdrReachabilityDataset(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/pdr.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Bad, Next0, Next1 uint8
			Returncode        int
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 256 {
		t.Fatal("incomplete PDR reachability dataset")
	}
	for _, c := range fixture.Cases {
		truth := bmcTruthCase{bad: c.Bad, next0: c.Next0, next1: c.Next1, constraint: 255}
		bad := bmcTruthExpected(truth, 4) >= 0
		expected := 1
		if bad {
			expected = 0
		}
		if c.Returncode != expected {
			t.Fatal("dataset result disagrees with exhaustive reachability", c)
		}
		pdrVerifyTruthCase(t, truth, false)
	}
}

func TestPdrSmokeInitializationAndCombinational(t *testing.T) {
	for _, tc := range []struct {
		data string
		bad  bool
	}{
		{"aig 2 1 1 1 0\n0\n4\n", false},
		{"aig 2 1 1 1 0\n2\n4\n", true},
		{"aig 2 1 1 1 0\n0\n5\n", true},
		{"aig 1 0 1 1 0\n3\n2\n", true},
		{"aig 1 0 1 1 0\n2 1\n2\n", true},
		{"aig 1 0 1 1 0\n2 2\n2\n", true},
		{"aig 0 0 0 1 0\n1\n", true},
		{"aig 0 0 0 1 0\n0\n", false},
		{"aig 0 0 0 0 0 1 1\n1\n0\n", false},
		{"aig 1 1 0 0 0 1 1\n3\n2\n", false},
	} {
		m := bmcTestModel(t, tc.data)
		r, err := m.Prove(PdrOptions{})
		expected := PdrProved
		if tc.bad {
			expected = PdrCounterexample
		}
		if err != nil || r.Status != expected {
			t.Fatal(tc, r, err)
		}
		if tc.bad {
			if err := m.ValidateCounterexample(r.Counterexample); err != nil {
				t.Fatal(err)
			}
		}
	}
}
func TestPdrProofQueryBudgetIncludesVerification(t *testing.T) {
	m := bmcTestModel(t, "aig 0 0 0 1 0\n0\n")
	for cap := uint64(1); cap <= 5; cap++ {
		r, err := m.Prove(PdrOptions{MaxQueries: cap})
		if err != nil || r.Stats.Queries > cap {
			t.Fatal(cap, r, err)
		}
		if cap < 5 {
			if r.Status != PdrUnknown || r.Reason != PdrQueryLimit || r.Invariant != nil {
				t.Fatal("partial verification claimed proof", cap, r)
			}
		} else if r.Status != PdrProved {
			t.Fatal(cap, r)
		}
	}
}
