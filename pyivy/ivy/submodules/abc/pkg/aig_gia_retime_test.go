package abc

// Go-only state-transition models, exhaustive checks, and fuzz tests for
// src/aig/gia/giaRetime.c.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func giaRetimeFixture(t testing.TB, complemented bool) (*GiaManager, GiaID) {
	t.Helper()
	m := NewGiaManager()
	m.Name, m.Spec = "retime", "retime.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	if complemented {
		a, b = a.Not(), b.Not()
	}
	g, err := m.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.AppendCO(g); err != nil {
		t.Fatal(err)
	}
	if _, err = m.AppendCO(p); err != nil {
		t.Fatal(err)
	}
	if _, err = m.AppendCO(p.Not()); err != nil {
		t.Fatal(err)
	}
	if err = m.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	if err = m.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	return m, g.ID()
}

// giaCheckOneRetimeRound proves the state encoding reported by one round:
// newState[j] is the old cut value XOR its all-zero phase.
func giaCheckOneRetimeRound(t testing.TB, source *GiaManager, result GiaRetimeResult) {
	t.Helper()
	if result.Graph == nil || len(result.Rounds) != 1 {
		t.Fatal("incomplete retiming result")
	}
	round := result.Rounds[0]
	if len(round.Cut) != result.Graph.RegisterCount() || len(round.CutPhases) != len(round.Cut) {
		t.Fatal(round, result.Graph.RegisterCount())
	}
	for stateBits := 0; stateBits < 1<<source.RegisterCount(); stateBits++ {
		state := make([]bool, source.RegisterCount())
		for i := range state {
			state[i] = stateBits&(1<<i) != 0
		}
		stateValues, err := source.evaluateValues(append(make([]bool, source.PrimaryInputCount()), state...))
		if err != nil {
			t.Fatal(err)
		}
		encoded := make([]bool, len(round.Cut))
		for i, id := range round.Cut {
			encoded[i] = stateValues[id] != round.CutPhases[i]
		}
		for primaryBits := 0; primaryBits < 1<<source.PrimaryInputCount(); primaryBits++ {
			primary := make([]bool, source.PrimaryInputCount())
			for i := range primary {
				primary[i] = primaryBits&(1<<i) != 0
			}
			wantOut, oldNext, oldErr := source.Step(primary, state)
			gotOut, newNext, newErr := result.Graph.Step(primary, encoded)
			if oldErr != nil || newErr != nil || !slices.Equal(gotOut, wantOut) {
				t.Fatal(state, primary, wantOut, gotOut, oldErr, newErr)
			}
			nextValues, valueErr := source.evaluateValues(append(make([]bool, source.PrimaryInputCount()), oldNext...))
			if valueErr != nil {
				t.Fatal(valueErr)
			}
			wantNext := make([]bool, len(round.Cut))
			for i, id := range round.Cut {
				wantNext[i] = nextValues[id] != round.CutPhases[i]
			}
			if !slices.Equal(newNext, wantNext) {
				t.Fatal(state, primary, newNext, wantNext)
			}
		}
	}
}

func TestGiaRetimeForwardCutClassesPhasesAndOwnership(t *testing.T) {
	for _, complemented := range []bool{false, true} {
		m, gate := giaRetimeFixture(t, complemented)
		before := slices.Clone(m.objects)
		result, err := m.RetimeForwardOne([]int{7, 7}, GiaRetimeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		round := result.Rounds[0]
		if round.Fixed != 0 || round.Moves != 1 || !slices.Equal(round.Cut, []GiaID{gate}) || !slices.Equal(result.Classes, []int{7}) {
			t.Fatal(complemented, round, result.Classes)
		}
		if !slices.Equal(round.CutPhases, []bool{complemented}) || result.Graph.Name != m.Name || result.Graph.Spec != m.Spec || result.Graph.ConstraintCount() != 1 || !slices.Equal(before, m.objects) {
			t.Fatal(complemented, round.CutPhases, result.Graph)
		}
		giaCheckOneRetimeRound(t, m, result)

		blocked, blockedErr := m.RetimeForwardOne([]int{3, 4}, GiaRetimeOptions{})
		if blockedErr != nil || blocked.Rounds[0].Moves != 0 || !slices.Equal(blocked.Rounds[0].Cut, m.cis[m.PrimaryInputCount():]) || !slices.Equal(blocked.Classes, []int{3, 4}) {
			t.Fatal(blocked, blockedErr)
		}
		giaCheckOneRetimeRound(t, m, blocked)
	}
}

func TestGiaRetimeForwardRoundsAndSequentialTrace(t *testing.T) {
	m, _ := giaRetimeFixture(t, false)
	result, err := m.RetimeForward(nil, GiaRetimeOptions{})
	if err != nil || len(result.Rounds) != 2 || result.Rounds[0].Moves != 1 || result.Rounds[1].Moves != 0 {
		t.Fatal(result, err)
	}
	oldState := make([]bool, m.RegisterCount())
	newState := make([]bool, result.Graph.RegisterCount())
	for cycle := 0; cycle < 32; cycle++ {
		input := []bool{cycle%3 == 1}
		oldOut, oldNext, oldErr := m.Step(input, oldState)
		newOut, newNext, newErr := result.Graph.Step(input, newState)
		if oldErr != nil || newErr != nil || !slices.Equal(oldOut, newOut) {
			t.Fatal(cycle, oldOut, newOut, oldErr, newErr)
		}
		oldState, newState = oldNext, newNext
	}
	one, err := m.RetimeForward(nil, GiaRetimeOptions{MaxIterations: 1})
	if err != nil || len(one.Rounds) != 1 || one.Rounds[0].Moves != 1 {
		t.Fatal(one, err)
	}
}

func TestGiaRetimeAutonomousCycleLimitsAndUnsupported(t *testing.T) {
	cycle := NewGiaManager()
	a, _ := cycle.AppendCI()
	b, _ := cycle.AppendCI()
	cycle.AppendCO(a)
	cycle.AppendCO(b)
	cycle.SetRegisterCount(2)
	result, err := cycle.RetimeForwardOne(nil, GiaRetimeOptions{})
	if err != nil || result.Rounds[0].Fixed != 2 || result.Rounds[0].Moves != 0 {
		t.Fatal(result, err)
	}
	giaCheckOneRetimeRound(t, cycle, result)

	m, _ := giaRetimeFixture(t, false)
	for _, options := range []GiaRetimeOptions{{MaxIterations: -1}, {MaxObjects: -1}, {MaxVisits: -1}, {MaxObjects: 1}, {MaxVisits: 1}} {
		if got, optionErr := m.RetimeForward(nil, options); optionErr == nil || got.Graph != nil {
			t.Fatal(options, got, optionErr)
		}
	}
	if got, classErr := m.RetimeForward([]int{1}, GiaRetimeOptions{}); !errors.Is(classErr, ErrDimension) || got.Graph != nil {
		t.Fatal(got, classErr)
	}
	if got, nilErr := (*GiaManager)(nil).RetimeForward(nil, GiaRetimeOptions{}); !errors.Is(nilErr, ErrDimension) || got.Graph != nil {
		t.Fatal(got, nilErr)
	}
	for threshold := 1; threshold <= 8; threshold++ {
		calls := 0
		got, cancelErr := m.RetimeForward(nil, GiaRetimeOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(cancelErr, ErrCancelled) || got.Graph != nil {
			t.Fatal(threshold, calls, got, cancelErr)
		}
	}
	for _, native := range []func(*GiaManager, GiaLit, GiaLit) error{
		func(g *GiaManager, a, b GiaLit) error { _, e := g.AppendXorReal(a, b); return e },
		func(g *GiaManager, a, b GiaLit) error { _, e := g.AppendMuxReal(a, b, GiaConst0); return e },
	} {
		g := NewGiaManager()
		a, _ := g.AppendCI()
		b, _ := g.AppendCI()
		if e := native(g, a, b); e != nil {
			t.Fatal(e)
		}
		if got, nativeErr := g.RetimeForward(nil, GiaRetimeOptions{}); !errors.Is(nativeErr, ErrUnsupported) || got.Graph != nil {
			t.Fatal(got, nativeErr)
		}
	}
}

func TestGiaRetimeDeepIterative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	chain, _ := m.AppendAnd(a, b)
	for i := 0; i < 10000; i++ {
		chain, _ = m.AppendAnd(chain, a.NotIf(i&1 != 0))
	}
	m.AppendCO(chain)
	m.AppendCO(p)
	m.AppendCO(p.Not())
	m.SetRegisterCount(2)
	result, err := m.RetimeForwardOne(nil, GiaRetimeOptions{})
	if err != nil || result.Rounds[0].Moves != 10001 || result.Graph.RegisterCount() != 1 {
		t.Fatal(result.Rounds, err)
	}
}

func FuzzGiaRetimeForwardOne(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(0x9e3779b97f4a7c15))
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		m := NewGiaManager()
		primary := 1 + random.Intn(2)
		registers := 1 + random.Intn(3)
		pis := make([]GiaLit, primary)
		states := make([]GiaLit, registers)
		for i := range pis {
			pis[i], _ = m.AppendCI()
		}
		for i := range states {
			states[i], _ = m.AppendCI()
		}
		nodes := slices.Clone(states)
		for i := 0; i < 1+random.Intn(5); i++ {
			a := nodes[random.Intn(len(nodes))]
			b := states[random.Intn(len(states))]
			if a.ID() == b.ID() {
				b = b.Not()
				// Raw ANDs require distinct regular fanins.
				if a.ID() == b.ID() {
					continue
				}
			}
			n, e := m.AppendAnd(a.NotIf(random.Intn(2) != 0), b.NotIf(random.Intn(2) != 0))
			if e == nil {
				nodes = append(nodes, n)
			}
		}
		m.AppendCO(nodes[random.Intn(len(nodes))].NotIf(random.Intn(2) != 0))
		for i := 0; i < registers; i++ {
			driver := pis[random.Intn(len(pis))].NotIf(random.Intn(2) != 0)
			m.AppendCO(driver)
		}
		m.SetRegisterCount(registers)
		result, err := m.RetimeForwardOne(nil, GiaRetimeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		giaCheckOneRetimeRound(t, m, result)
	})
}
