package abc

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func msatOrderCone(count int) []SatVar {
	result := make([]SatVar, count)
	for index := range result {
		result[index] = SatVar(index)
	}
	return result
}

func TestMsatVariableOrderSelectionUpdatesAndTies(t *testing.T) {
	order, err := NewMsatVariableOrder(6, 16)
	if err != nil {
		t.Fatal(err)
	}
	if err = order.Clean(msatOrderCone(6)); err != nil || order.Check() != nil {
		t.Fatal(err, order.Check(), order.Heap())
	}
	if err = order.IncreaseActivity(3, 5); err != nil {
		t.Fatal(err)
	}
	if err = order.IncreaseActivity(4, 2); err != nil || order.Check() != nil {
		t.Fatal(err, order.Check(), order.Heap())
	}
	if variable, found, err := order.Select(); err != nil || !found || variable != 3 {
		t.Fatal(variable, found, err)
	}
	if err = order.Assign(3); err != nil {
		t.Fatal(err)
	}
	if variable, found, err := order.Select(); err != nil || !found || variable != 4 {
		t.Fatal(variable, found, err)
	}
	if err = order.Unassign(3); err != nil || order.Check() != nil {
		t.Fatal(err, order.Check())
	}
	if variable, found, err := order.Select(); err != nil || !found || variable != 3 {
		t.Fatal(variable, found, err)
	}

	ties, _ := NewMsatVariableOrder(20, 32)
	ties.Clean(msatOrderCone(20))
	got := make([]SatVar, 0, 20)
	for {
		variable, found, err := ties.Select()
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			break
		}
		got = append(got, variable)
	}
	if want := []SatVar{0, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}; !slices.Equal(got, want) {
		t.Fatal(got, want)
	}
}

func TestMsatVariableOrderLazyAssignedRescaleAndBounds(t *testing.T) {
	order, _ := NewMsatVariableOrder(5, 8)
	order.Clean(msatOrderCone(5))
	order.IncreaseActivity(0, 10)
	order.IncreaseActivity(1, 8)
	order.IncreaseActivity(2, 6)
	order.Assign(0)
	order.Assign(1)
	if variable, found, err := order.Select(); err != nil || !found || variable != 2 {
		t.Fatal(variable, found, err)
	}
	if slices.Contains(order.Heap(), SatVar(0)) || slices.Contains(order.Heap(), SatVar(1)) || slices.Contains(order.Heap(), SatVar(2)) {
		t.Fatal("popped entries remain", order.Heap())
	}
	before := order.Activities()
	if err := order.RescaleActivities(1e-100); err != nil {
		t.Fatal(err)
	}
	for index, activity := range order.Activities() {
		if activity != before[index]*1e-100 {
			t.Fatal(index, activity, before[index])
		}
	}
	if err := order.SetBounds(8); err != nil || len(order.Activities()) != 8 || len(order.Heap()) != 0 {
		t.Fatal(err, len(order.Activities()), order.Heap())
	}
	if order.Activities()[0] != before[0]*1e-100 {
		t.Fatal("bounds lost prefix activity")
	}
	if err := order.SetBounds(3); err != nil || len(order.Activities()) != 3 {
		t.Fatal(err, len(order.Activities()))
	}
}

func TestMsatVariableOrderErrorsOwnershipAndLifecycle(t *testing.T) {
	for _, test := range []struct {
		variables, maximum int
		want               error
	}{
		{-1, 0, ErrDimension}, {0, -1, ErrDimension}, {9, 8, ErrLimit},
	} {
		order, err := NewMsatVariableOrder(test.variables, test.maximum)
		if order != nil || !errors.Is(err, test.want) {
			t.Fatal(test, order, err)
		}
	}
	order, _ := NewMsatVariableOrder(4, 4)
	if err := order.Clean([]SatVar{0, 1, 1}); !errors.Is(err, ErrDimension) || len(order.Heap()) != 0 {
		t.Fatal(err, order.Heap())
	}
	if err := order.Clean([]SatVar{4}); !errors.Is(err, ErrDimension) || len(order.Heap()) != 0 {
		t.Fatal(err, order.Heap())
	}
	order.Clean(msatOrderCone(4))
	heap := order.Heap()
	heap[0] = 3
	if order.Heap()[0] != 0 {
		t.Fatal("heap snapshot aliases")
	}
	activities := order.Activities()
	activities[0] = 9
	if order.Activities()[0] != 0 {
		t.Fatal("activity snapshot aliases")
	}
	for _, operation := range []func() error{
		func() error { return order.Assign(4) },
		func() error { return order.Unassign(4) },
		func() error { return order.IncreaseActivity(4, 1) },
		func() error { return order.IncreaseActivity(0, -1) },
		func() error { return order.IncreaseActivity(0, math.NaN()) },
		func() error { return order.RescaleActivities(-1) },
		func() error { return order.SetBounds(5) },
	} {
		beforeHeap, beforeActivities := order.Heap(), order.Activities()
		if err := operation(); err == nil || !slices.Equal(beforeHeap, order.Heap()) || !slices.Equal(beforeActivities, order.Activities()) {
			t.Fatal(err, beforeHeap, order.Heap(), beforeActivities, order.Activities())
		}
	}
	order.Close()
	order.Close()
	if order.Heap() != nil || order.Activities() != nil {
		t.Fatal("closed order exposes storage")
	}
	if _, _, err := order.Select(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, _, err := (*MsatVariableOrder)(nil).Select(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzMsatVariableOrder(f *testing.F) {
	for _, seed := range [][]byte{{}, {0, 0}, {0, 3, 0, 4, 3, 0}, {1, 0, 3, 0, 2, 0}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, operations []byte) {
		const variables = 16
		order, err := NewMsatVariableOrder(variables, variables)
		if err != nil {
			t.Fatal(err)
		}
		if err = order.Clean(msatOrderCone(variables)); err != nil {
			t.Fatal(err)
		}
		activities := make([]float64, variables)
		assigned := make([]bool, variables)
		for index := 0; index+1 < len(operations); index += 2 {
			op := operations[index] % 5
			variable := SatVar(operations[index+1] % variables)
			switch op {
			case 0:
				amount := float64(1 + operations[index+1]%11)
				if err = order.IncreaseActivity(variable, amount); err != nil {
					t.Fatal(err)
				}
				activities[variable] += amount
			case 1:
				if err = order.Assign(variable); err != nil {
					t.Fatal(err)
				}
				assigned[variable] = true
			case 2:
				if err = order.Unassign(variable); err != nil {
					t.Fatal(err)
				}
				assigned[variable] = false
			case 3:
				heapBefore := order.Heap()
				best := math.Inf(-1)
				for _, candidate := range heapBefore {
					if !assigned[candidate] && activities[candidate] > best {
						best = activities[candidate]
					}
				}
				selected, found, selectErr := order.Select()
				if selectErr != nil || found != !math.IsInf(best, -1) || found && (assigned[selected] || activities[selected] != best) {
					t.Fatal(selected, found, selectErr, best, heapBefore, activities, assigned)
				}
			case 4:
				if err = order.RescaleActivities(0.5); err != nil {
					t.Fatal(err)
				}
				for variable := range activities {
					activities[variable] *= 0.5
				}
			}
			if err = order.Check(); err != nil || !slices.Equal(order.Activities(), activities) {
				t.Fatal(err, order.Heap(), order.Activities(), activities)
			}
		}
	})
}
