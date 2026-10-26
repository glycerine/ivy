package abc

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func msatActivityFixture(t testing.TB, variables, clauses int, config MsatActivityConfig) (*MsatActivityState, *MsatVariableOrder, *MsatClauseVector, []*MsatClause) {
	t.Helper()
	order, err := NewMsatVariableOrder(variables, max(variables, 1))
	if err != nil {
		t.Fatal(err)
	}
	if err = order.Clean(msatOrderCone(variables)); err != nil {
		t.Fatal(err)
	}
	learned, err := NewMsatClauseVector(0, max(clauses, 1))
	if err != nil {
		t.Fatal(err)
	}
	owned := msatSortTestClauses(make([]float32, clauses))
	for _, clause := range owned {
		if err = learned.Push(clause); err != nil {
			t.Fatal(err)
		}
	}
	state, err := NewMsatActivityState(order, learned, config)
	if err != nil {
		t.Fatal(err)
	}
	return state, order, learned, owned
}

func TestMsatActivityVariableBumpDecayStaticAndRescale(t *testing.T) {
	config := MsatActivityConfig{VariableIncrement: 2, VariableDecay: 3, ClauseIncrement: 4, ClauseDecay: 5}
	state, order, _, _ := msatActivityFixture(t, 3, 2, config)
	literal, _ := MakeSatLit(1, true)
	if err := state.VariableBump(literal); err != nil {
		t.Fatal(err)
	}
	if err := state.VariableDecay(); err != nil {
		t.Fatal(err)
	}
	if err := state.VariableBump(literal); err != nil {
		t.Fatal(err)
	}
	snapshot, err := state.Snapshot()
	if err != nil || !slices.Equal(snapshot.VariableActivities, []float64{0, 8, 0}) || snapshot.Config.VariableIncrement != 6 {
		t.Fatal(snapshot, err)
	}
	if variable, found, err := order.Select(); err != nil || !found || variable != 1 {
		t.Fatal(variable, found, err)
	}

	threshold, _, _, _ := msatActivityFixture(t, 2, 0, MsatActivityConfig{VariableIncrement: 1e100, VariableDecay: 1, ClauseIncrement: 1, ClauseDecay: 1})
	lit0, _ := MakeSatLit(0, false)
	threshold.VariableBump(lit0)
	if err = threshold.VariableBump(lit0); err != nil {
		t.Fatal(err)
	}
	snapshot, _ = threshold.Snapshot()
	if snapshot.VariableActivities[0] != 2 || snapshot.Config.VariableIncrement != 1 {
		t.Fatal(snapshot)
	}

	static, _, _, _ := msatActivityFixture(t, 2, 0, MsatActivityConfig{VariableIncrement: 7, VariableDecay: -1, ClauseIncrement: 1, ClauseDecay: 1})
	if err = static.VariableBump(lit0); err != nil {
		t.Fatal(err)
	}
	if err = static.VariableDecay(); err != nil {
		t.Fatal(err)
	}
	snapshot, _ = static.Snapshot()
	if snapshot.VariableActivities[0] != 0 || snapshot.Config.VariableIncrement != 7 {
		t.Fatal(snapshot)
	}
}

func TestMsatActivityClauseBumpDecayAndRescale(t *testing.T) {
	config := MsatActivityConfig{VariableIncrement: 1, VariableDecay: 1, ClauseIncrement: 4, ClauseDecay: 2.5}
	state, _, _, clauses := msatActivityFixture(t, 2, 3, config)
	clauses[0].activity = 1.5
	clauses[1].activity = 5
	if err := state.ClauseBump(clauses[0]); err != nil {
		t.Fatal(err)
	}
	if err := state.ClauseDecay(); err != nil {
		t.Fatal(err)
	}
	if err := state.ClauseBump(clauses[2]); err != nil {
		t.Fatal(err)
	}
	snapshot, err := state.Snapshot()
	if err != nil || !slices.Equal(snapshot.ClauseActivities, []float32{5.5, 5, 10}) || snapshot.Config.ClauseIncrement != 10 {
		t.Fatal(snapshot, err)
	}
	clauses[0].activity = float32(1e20)
	if err = state.ClauseBump(clauses[0]); err != nil {
		t.Fatal(err)
	}
	snapshot, _ = state.Snapshot()
	if snapshot.Config.ClauseIncrement != 1e-19 || snapshot.ClauseActivities[0] < 0.99 || snapshot.ClauseActivities[0] > 1.01 || snapshot.ClauseActivities[1] != 5e-20 || snapshot.ClauseActivities[2] != 1e-19 {
		t.Fatal(snapshot)
	}
}

func TestMsatActivityAtomicErrorsOwnershipAndLifecycle(t *testing.T) {
	order, _ := NewMsatVariableOrder(2, 2)
	learned, _ := NewMsatClauseVector(0, 2)
	valid := MsatActivityConfig{VariableIncrement: 1, VariableDecay: 1, ClauseIncrement: 1, ClauseDecay: 1}
	for _, config := range []MsatActivityConfig{
		{VariableIncrement: -1},
		{VariableIncrement: math.Inf(1)},
		{VariableDecay: math.NaN()},
		{ClauseIncrement: -1},
		{ClauseDecay: -1},
	} {
		if state, err := NewMsatActivityState(order, learned, config); state != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(config, state, err)
		}
	}
	state, err := NewMsatActivityState(order, learned, valid)
	if err != nil {
		t.Fatal(err)
	}
	if err = state.VariableBump(InvalidSatLit); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if err = state.ClauseBump(nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	bad := &MsatClause{activity: float32(math.Inf(1))}
	if err = learned.Push(bad); err != nil {
		t.Fatal(err)
	}
	before := bad.activity
	if err = state.ClauseRescale(); !errors.Is(err, ErrInvariant) || bad.activity != before {
		t.Fatal(err, bad.activity, before)
	}
	bad.activity = 0
	snapshot, err := state.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	snapshot.VariableActivities[0] = 99
	snapshot.ClauseActivities[0] = 99
	again, _ := state.Snapshot()
	if again.VariableActivities[0] == 99 || again.ClauseActivities[0] == 99 {
		t.Fatal("snapshot aliases state")
	}
	state.Close()
	state.Close()
	if _, err = state.Snapshot(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = (*MsatActivityState)(nil).Snapshot(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzMsatActivityState(f *testing.F) {
	for _, seed := range [][]byte{{}, {0, 0}, {0, 1, 1, 0, 2, 0}, {2, 1, 3, 0, 4, 0}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, operations []byte) {
		config := MsatActivityConfig{VariableIncrement: 1.25, VariableDecay: 1.1, ClauseIncrement: 0.75, ClauseDecay: 1.2}
		state, order, _, clauses := msatActivityFixture(t, 8, 4, config)
		variableActivities := make([]float64, 8)
		clauseActivities := make([]float32, 4)
		for index := 0; index+1 < len(operations); index += 2 {
			op, raw := operations[index]%6, operations[index+1]
			switch op {
			case 0:
				variable := SatVar(raw % 8)
				literal, _ := MakeSatLit(variable, raw&0x80 != 0)
				if err := state.VariableBump(literal); err != nil {
					t.Fatal(err)
				}
				variableActivities[variable] += config.VariableIncrement
			case 1:
				if err := state.VariableDecay(); err != nil {
					t.Fatal(err)
				}
				config.VariableIncrement *= config.VariableDecay
			case 2:
				clause := int(raw) % len(clauses)
				if err := state.ClauseBump(clauses[clause]); err != nil {
					t.Fatal(err)
				}
				clauseActivities[clause] += float32(config.ClauseIncrement)
			case 3:
				if err := state.ClauseDecay(); err != nil {
					t.Fatal(err)
				}
				config.ClauseIncrement *= config.ClauseDecay
			case 4:
				if err := state.VariableRescale(); err != nil {
					t.Fatal(err)
				}
				for variable := range variableActivities {
					variableActivities[variable] *= 1e-100
				}
				config.VariableIncrement *= 1e-100
			case 5:
				if err := state.ClauseRescale(); err != nil {
					t.Fatal(err)
				}
				for clause := range clauseActivities {
					clauseActivities[clause] *= float32(1e-20)
				}
				config.ClauseIncrement *= 1e-20
			}
			snapshot, err := state.Snapshot()
			if err != nil || snapshot.Config != config || !slices.Equal(snapshot.VariableActivities, variableActivities) || !slices.Equal(snapshot.ClauseActivities, clauseActivities) || order.Check() != nil {
				t.Fatal(err, snapshot, config, variableActivities, clauseActivities, order.Heap())
			}
		}
	})
}
