package abc

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func msatClauseLit(variable int, negative bool) SatLit {
	literal, err := MakeSatLit(SatVar(variable), negative)
	if err != nil {
		panic(err)
	}
	return literal
}

func msatClauseFixture(t testing.TB, variables, maximum int) (*MsatClauseContext, *MsatActivityState, *MsatVariableOrder, *MsatClauseVector) {
	t.Helper()
	order, err := NewMsatVariableOrder(variables, max(variables, 1))
	if err != nil {
		t.Fatal(err)
	}
	cone := make([]SatVar, variables)
	for index := range cone {
		cone[index] = SatVar(index)
	}
	if err = order.Clean(cone); err != nil {
		t.Fatal(err)
	}
	learned, err := NewMsatClauseVector(0, max(maximum, 1))
	if err != nil {
		t.Fatal(err)
	}
	activity, err := NewMsatActivityState(order, learned, MsatActivityConfig{
		VariableIncrement: 2,
		VariableDecay:     1.5,
		ClauseIncrement:   3,
		ClauseDecay:       2,
	})
	if err != nil {
		t.Fatal(err)
	}
	context, err := NewMsatClauseContext(variables, maximum, activity)
	if err != nil {
		t.Fatal(err)
	}
	return context, activity, order, learned
}

func TestMsatClauseCreateOriginalSemantics(t *testing.T) {
	x0, nx0 := msatClauseLit(0, false), msatClauseLit(0, true)
	x1, x2 := msatClauseLit(1, false), msatClauseLit(2, false)

	context, _, _, _ := msatClauseFixture(t, 3, 8)
	if status, clause, err := context.Create(nil, false); err != nil || status != MsatClauseConflict || clause != nil {
		t.Fatal(status, clause, err)
	}
	input := []SatLit{x1, x0, x1, x0}
	before := slices.Clone(input)
	status, clause, err := context.Create(input, false)
	if err != nil || status != MsatClauseStored || clause == nil || !slices.Equal(clause.Literals(), []SatLit{x0, x1}) || !slices.Equal(input, before) {
		t.Fatal(status, clause, err, input)
	}
	if watched, err := context.Watch(nx0); err != nil || len(watched) != 1 || watched[0] != clause {
		t.Fatal(watched, err)
	}
	if status, clause, err = context.Create([]SatLit{x0, nx0, x2}, false); err != nil || status != MsatClauseSatisfied || clause != nil {
		t.Fatal(status, clause, err)
	}

	assigned, _, _, _ := msatClauseFixture(t, 3, 8)
	if err = assigned.SetAssignment(0, &x0, 0, nil); err != nil {
		t.Fatal(err)
	}
	if status, clause, err = assigned.Create([]SatLit{nx0, x2, x2}, false); err != nil || status != MsatClauseUnit || clause != nil || assigned.Assignments()[2] != int(x2) {
		t.Fatal(status, clause, err, assigned.Assignments())
	}
	if status, clause, err = assigned.Create([]SatLit{x0, x1}, false); err != nil || status != MsatClauseSatisfied || clause != nil {
		t.Fatal(status, clause, err)
	}

	limited, _, _, _ := msatClauseFixture(t, 3, 1)
	if status, clause, err = limited.Create([]SatLit{x0, x1}, false); err != nil || status != MsatClauseStored || clause == nil {
		t.Fatal(status, clause, err)
	}
	if status, clause, err = limited.Create([]SatLit{x2}, false); err != nil || status != MsatClauseUnit || clause != nil {
		t.Fatal("unit must not consume stored-clause capacity", status, clause, err)
	}
	if status, clause, err = limited.Create([]SatLit{x0, nx0}, false); err != nil || status != MsatClauseSatisfied || clause != nil {
		t.Fatal("satisfied clauses must not consume stored-clause capacity", status, clause, err)
	}
	if status, clause, err = limited.Create([]SatLit{nx0, x1}, false); !errors.Is(err, ErrLimit) || status != MsatClauseConflict || clause != nil {
		t.Fatal(status, clause, err)
	}
}

func TestMsatClauseLearnedOrderingActivityReasonAndAccessors(t *testing.T) {
	context, activity, order, learned := msatClauseFixture(t, 3, 8)
	x0 := msatClauseLit(0, false)
	nx1, nx2 := msatClauseLit(1, true), msatClauseLit(2, true)
	if err := context.SetDecisionLevel(5); err != nil {
		t.Fatal(err)
	}
	if err := context.SetAssignment(1, &nx1, 2, nil); err != nil {
		t.Fatal(err)
	}
	if err := context.SetAssignment(2, &nx2, 4, nil); err != nil {
		t.Fatal(err)
	}
	status, clause, err := context.Create([]SatLit{x0, nx1.Not(), nx2.Not()}, true)
	if err != nil || status != MsatClauseStored || !slices.Equal(clause.Literals(), []SatLit{x0, nx2.Not(), nx1.Not()}) {
		t.Fatal(status, clause, err)
	}
	if !clause.Learned() || clause.Len() != 3 || clause.Marked() || clause.TypeA() {
		t.Fatal(clause)
	}
	if number, err := clause.Number(); err != nil || number != 0 {
		t.Fatal(number, err)
	}
	if got, _ := clause.Activity(); got != 3 {
		t.Fatal(got)
	}
	if snapshot, err := activity.Snapshot(); err != nil || !slices.Equal(snapshot.VariableActivities, []float64{2, 2, 2}) {
		t.Fatal(snapshot, err)
	}
	if err = learned.Push(clause); err != nil {
		t.Fatal(err)
	}
	reason, err := clause.Reason(&x0, context.Assignments(), activity)
	if err != nil || !slices.Equal(reason, []SatLit{nx2, nx1}) {
		t.Fatal(reason, err)
	}
	if got, _ := clause.Activity(); got != 6 {
		t.Fatal(got)
	}
	context.reasons[0] = clause
	if locked, err := clause.IsLocked(context.reasons); err != nil || !locked {
		t.Fatal(locked, err)
	}
	if err = clause.SetMarked(true); err != nil {
		t.Fatal(err)
	}
	if err = clause.SetTypeA(true); err != nil {
		t.Fatal(err)
	}
	if err = clause.SetNumber(17); err != nil {
		t.Fatal(err)
	}
	if err = clause.SetActivity(1.25); err != nil || !clause.Marked() || !clause.TypeA() {
		t.Fatal(err)
	}
	if err = clause.SetActivity(float32(math.Inf(1))); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	literals := clause.Literals()
	literals[0] = nx1
	if clause.Literals()[0] != x0 {
		t.Fatal("literal snapshot aliases clause storage")
	}
	if !order.assigned[1] || !order.assigned[2] {
		t.Fatal(order.assigned)
	}
}

func TestMsatClausePropagateAndSimplify(t *testing.T) {
	x0, x1, x2 := msatClauseLit(0, false), msatClauseLit(1, false), msatClauseLit(2, false)

	clause := &MsatClause{literals: []SatLit{x0, x1, x2}}
	assignments := []int{-1, int(x1.Not()), -1}
	result, err := clause.Propagate(x1.Not(), assignments)
	if err != nil || !result.WatchMoved || result.WatchBucket != x2.Not() || !slices.Equal(clause.Literals(), []SatLit{x0, x2, x1}) {
		t.Fatal(result, clause.Literals(), err)
	}
	assignments = []int{-1, int(x1.Not()), int(x2.Not())}
	clause = &MsatClause{literals: []SatLit{x0, x1, x2}}
	result, err = clause.Propagate(x1.Not(), assignments)
	if err != nil || !result.Unit || result.Literal != x0 || result.WatchMoved {
		t.Fatal(result, err)
	}
	assignments[0] = int(x0)
	result, err = clause.Propagate(x1.Not(), assignments)
	if err != nil || !result.Satisfied {
		t.Fatal(result, err)
	}
	clause = &MsatClause{literals: []SatLit{x0, x1, x2}}
	assignments = []int{int(x0.Not()), -1, -1}
	result, err = clause.Propagate(x0.Not(), assignments)
	if err != nil || !result.WatchMoved || !slices.Equal(clause.Literals(), []SatLit{x1, x2, x0}) {
		t.Fatal(result, clause.Literals(), err)
	}

	x3 := msatClauseLit(3, false)
	clause = &MsatClause{literals: []SatLit{x0, x1, x2, x3}, activity: 9}
	satisfied, err := clause.Simplify([]int{-1, -1, int(x2.Not()), -1})
	if err != nil || satisfied || !slices.Equal(clause.Literals(), []SatLit{x0, x1, x3}) || clause.activity != 9 {
		t.Fatal(satisfied, clause.Literals(), clause.activity, err)
	}
	before := clause.Literals()
	satisfied, err = clause.Simplify([]int{-1, -1, -1, int(x3)})
	if err != nil || !satisfied || !slices.Equal(clause.Literals(), before) {
		t.Fatal(satisfied, clause.Literals(), err)
	}
	if _, err = clause.Simplify([]int{int(x0.Not()), -1, -1, -1}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestMsatClauseFormattingWatchRemovalLifecycleAndErrors(t *testing.T) {
	context, _, _, _ := msatClauseFixture(t, 3, 8)
	x0, nx1, x2 := msatClauseLit(0, false), msatClauseLit(1, true), msatClauseLit(2, false)
	status, clause, err := context.Create([]SatLit{x0, nx1, x2}, false)
	if err != nil || status != MsatClauseStored {
		t.Fatal(status, err)
	}
	if got := FormatMsatClause(clause); got != " 1 -2 3\n" {
		t.Fatal(got)
	}
	if got := FormatMsatClauseSymbols(clause); got != " 1 -2 3\n" {
		t.Fatal(got)
	}
	if got, err := clause.DIMACS(true); err != nil || got != "1 -2 3 0\n" {
		t.Fatal(got, err)
	}
	if got, err := clause.DIMACS(false); err != nil || got != "0 -1 2 \n" {
		t.Fatal(got, err)
	}
	if FormatMsatClause(nil) != "NULL pointer\n" || FormatMsatClauseSymbols(nil) != "NULL pointer\n" {
		t.Fatal("nil formatting")
	}
	if err = context.Free(clause, true); err != nil {
		t.Fatal(err)
	}
	for _, bucket := range []SatLit{x0.Not(), nx1.Not()} {
		if watched, err := context.Watch(bucket); err != nil || len(watched) != 0 {
			t.Fatal(bucket, watched, err)
		}
	}
	if _, err = clause.DIMACS(true); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	duplicateContext, _, _, _ := msatClauseFixture(t, 1, 4)
	if err = duplicateContext.SetAssignment(0, &x0, 1, nil); err != nil {
		t.Fatal(err)
	}
	status, duplicate, err := duplicateContext.Create([]SatLit{x0, x0}, true)
	if err != nil || status != MsatClauseStored {
		t.Fatal(status, duplicate, err)
	}
	if watched, err := duplicateContext.Watch(x0.Not()); err != nil || len(watched) != 2 || watched[0] != duplicate || watched[1] != duplicate {
		t.Fatal(watched, err)
	}
	if err = duplicateContext.Free(duplicate, true); err != nil {
		t.Fatal(err)
	}
	if watched, err := duplicateContext.Watch(x0.Not()); err != nil || len(watched) != 0 {
		t.Fatal(watched, err)
	}

	if err = context.SetDecisionLevel(1); err != nil {
		t.Fatal(err)
	}
	if _, _, err = context.Create([]SatLit{x0, x2}, false); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = context.Watch(InvalidSatLit); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	context.Close()
	context.Close()
	if _, _, err = context.Create([]SatLit{x0, x2}, false); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzMsatClauseOriginalModel(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 0, 2, 2, 5})
	f.Add([]byte{1, 0, 0, 0, 1, 0, 2, 4})
	f.Add([]byte{0, 0, 0, 0, 0, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		const variables = 4
		context, _, _, _ := msatClauseFixture(t, variables, 32)
		fixed := [variables]int{-1, -1, -1, -1}
		for variable := 0; variable < variables && variable < len(data); variable++ {
			switch data[variable] % 3 {
			case 1:
				literal := msatClauseLit(variable, false)
				fixed[variable] = int(literal)
				if err := context.SetAssignment(SatVar(variable), &literal, 0, nil); err != nil {
					t.Fatal(err)
				}
			case 2:
				literal := msatClauseLit(variable, true)
				fixed[variable] = int(literal)
				if err := context.SetAssignment(SatVar(variable), &literal, 0, nil); err != nil {
					t.Fatal(err)
				}
			}
		}
		literalData := data[min(len(data), variables):]
		if len(literalData) > 24 {
			literalData = literalData[:24]
		}
		input := make([]SatLit, len(literalData))
		for index, value := range literalData {
			input[index] = msatClauseLit(int(value)%variables, value&0x80 != 0)
		}
		before := slices.Clone(input)

		present := [variables][2]bool{}
		satisfied := false
		for _, literal := range input {
			variable, sign := int(literal.Var()), 0
			if literal.Negative() {
				sign = 1
			}
			present[variable][sign] = true
			if fixed[variable] == int(literal) {
				satisfied = true
			}
		}
		expected := make([]SatLit, 0, variables)
		for variable := 0; variable < variables; variable++ {
			if present[variable][0] && present[variable][1] {
				satisfied = true
			}
			for sign := 0; sign < 2; sign++ {
				if !present[variable][sign] || present[variable][1-sign] {
					continue
				}
				literal := msatClauseLit(variable, sign == 1)
				if fixed[variable] == -1 {
					expected = append(expected, literal)
				}
			}
		}
		status, clause, err := context.Create(input, false)
		if err != nil || !slices.Equal(input, before) {
			t.Fatal(status, err, input, before)
		}
		if satisfied {
			if status != MsatClauseSatisfied || clause != nil {
				t.Fatal("satisfied", status, clause, input, fixed)
			}
			return
		}
		switch len(expected) {
		case 0:
			if status != MsatClauseConflict || clause != nil {
				t.Fatal("conflict", status, clause, input, fixed)
			}
		case 1:
			if status != MsatClauseUnit || clause != nil || context.Assignments()[expected[0].Var()] != int(expected[0]) {
				t.Fatal("unit", status, clause, expected, context.Assignments())
			}
		default:
			if status != MsatClauseStored || clause == nil || !slices.Equal(clause.Literals(), expected) {
				t.Fatal("stored", status, clause, expected)
			}
			for assignment := 0; assignment < 1<<variables; assignment++ {
				consistent := true
				for variable := 0; variable < variables; variable++ {
					if fixed[variable] != -1 && ((fixed[variable]&1 == 0) != (assignment&(1<<variable) != 0)) {
						consistent = false
					}
				}
				if !consistent {
					continue
				}
				evaluate := func(literals []SatLit) bool {
					for _, literal := range literals {
						value := assignment&(1<<literal.Var()) != 0
						if value != literal.Negative() {
							return true
						}
					}
					return false
				}
				if evaluate(input) != evaluate(clause.Literals()) {
					t.Fatal("truth mismatch", input, clause.Literals(), assignment)
				}
			}
		}
	})
}

func FuzzMsatClausePropagationModel(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 0, 0})
	f.Add([]byte{1, 1, 2, 2, 0, 1})
	f.Add([]byte{0x80, 0, 2, 2, 1, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 6 {
			return
		}
		literals := make([]SatLit, 5)
		assignments := make([]int, 5)
		for variable := range literals {
			literals[variable] = msatClauseLit(variable, data[variable]&0x80 != 0)
			assignments[variable] = -1
			switch data[variable] % 3 {
			case 1:
				assignments[variable] = int(literals[variable])
			case 2:
				assignments[variable] = int(literals[variable].Not())
			}
		}
		watched := int(data[5] & 1)
		falseLiteral := literals[watched]
		trigger := falseLiteral.Not()
		assignments[falseLiteral.Var()] = int(falseLiteral.Not())

		expected := slices.Clone(literals)
		if watched == 0 {
			expected[0], expected[1] = expected[1], expected[0]
		}
		other := expected[0]
		want := MsatClausePropagation{}
		if assignments[other.Var()] == int(other) {
			want.Satisfied = true
		} else {
			move := -1
			for index := 2; index < len(expected); index++ {
				literal := expected[index]
				if assignments[literal.Var()] != int(literal.Not()) {
					move = index
					break
				}
			}
			if move >= 0 {
				expected[1], expected[move] = expected[move], expected[1]
				want.WatchMoved = true
				want.WatchBucket = expected[1].Not()
			} else {
				want.Unit = true
				want.Literal = other
			}
		}
		clause := &MsatClause{literals: slices.Clone(literals)}
		got, err := clause.Propagate(trigger, assignments)
		if err != nil || got != want || !slices.Equal(clause.Literals(), expected) {
			t.Fatal(got, want, clause.Literals(), expected, assignments, err)
		}
	})
}
