package abc

import (
	"errors"
	"slices"
	"testing"
)

func msatJConnect(t testing.TB, order *MsatJBoundaryOrder, adjacency [][]SatVar) {
	t.Helper()
	for variable := range adjacency {
		if err := order.SetAdjacency(SatVar(variable), adjacency[variable]); err != nil {
			t.Fatal(variable, err)
		}
	}
}

func TestMsatJBoundaryOrderTransitionsSelectionAndTies(t *testing.T) {
	order, err := NewMsatJBoundaryOrder(7, 8)
	if err != nil {
		t.Fatal(err)
	}
	adjacency := [][]SatVar{{1, 2}, {0, 3}, {0, 3, 4}, {1, 2, 5}, {2, 6}, {3}, {4}}
	msatJConnect(t, order, adjacency)
	if err = order.SetCone(msatOrderCone(6)); err != nil {
		t.Fatal(err)
	}
	if err = order.Assign(0); err != nil || !slices.Equal(order.Boundary(), []SatVar{1, 2}) {
		t.Fatal(err, order.Boundary())
	}
	if err = order.SetActivity(1, 4); err != nil {
		t.Fatal(err)
	}
	if err = order.SetActivity(2, 4); err != nil {
		t.Fatal(err)
	}
	if variable, found, selectErr := order.Select(); selectErr != nil || !found || variable != 1 {
		t.Fatal(variable, found, selectErr)
	}
	if err = order.Assign(1); err != nil || !slices.Equal(order.Boundary(), []SatVar{2, 3}) {
		t.Fatal(err, order.Boundary())
	}
	if err = order.SetActivity(3, 9); err != nil {
		t.Fatal(err)
	}
	if variable, found, selectErr := order.Select(); selectErr != nil || !found || variable != 3 {
		t.Fatal(variable, found, selectErr)
	}
	if err = order.Unassign(0); err != nil || !slices.Equal(order.Boundary(), []SatVar{3, 0}) {
		t.Fatal(err, order.Boundary())
	}
	if err = order.Unassign(1); err != nil || len(order.Boundary()) != 0 || order.Check() != nil {
		t.Fatal(err, order.Boundary(), order.Check())
	}
	if err = order.Assign(6); err != nil || len(order.Boundary()) != 0 {
		t.Fatal("outside-cone assignment entered boundary", err, order.Boundary())
	}
}

func TestMsatJBoundaryOrderBoundsOwnershipErrorsAndLifecycle(t *testing.T) {
	for _, test := range []struct {
		variables int
		maximum   int
		want      error
	}{{-1, 4, ErrDimension}, {5, 4, ErrLimit}, {0, -1, ErrDimension}} {
		order, err := NewMsatJBoundaryOrder(test.variables, test.maximum)
		if order != nil || !errors.Is(err, test.want) {
			t.Fatal(test, order, err)
		}
	}
	order, _ := NewMsatJBoundaryOrder(4, 6)
	if err := order.SetCone([]SatVar{0, 0}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := order.SetAdjacency(0, []SatVar{1, 1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := order.SetAdjacency(4, nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := order.SetActivity(0, -1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	msatJConnect(t, order, [][]SatVar{{1}, {0, 2}, {1, 3}, {2}})
	if err := order.SetCone(msatOrderCone(4)); err != nil {
		t.Fatal(err)
	}
	if err := order.Assign(0); err != nil {
		t.Fatal(err)
	}
	boundary := order.Boundary()
	boundary[0] = 3
	if !slices.Equal(order.Boundary(), []SatVar{1}) {
		t.Fatal("boundary snapshot aliases", order.Boundary())
	}
	assignments := order.Assignments()
	assignments[0] = false
	if !order.Assignments()[0] {
		t.Fatal("assignment snapshot aliases")
	}
	if err := order.Assign(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := order.Unassign(1); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := order.SetBounds(7); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err := order.SetBounds(3); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := order.SetCone(msatOrderCone(4)); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := order.SetAdjacency(0, nil); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := order.Unassign(0); err != nil || order.Check() != nil {
		t.Fatal(err, order.Check(), order.Boundary())
	}
	if err := order.SetBounds(3); err != nil || len(order.Activities()) != 3 || len(order.Boundary()) != 0 {
		t.Fatal(err, order.Activities(), order.Boundary())
	}
	order.Close()
	order.Close()
	if order.Activities() != nil || order.Assignments() != nil {
		t.Fatal("closed order exposes state")
	}
	if _, _, err := order.Select(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, _, err := (*MsatJBoundaryOrder)(nil).Select(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzMsatJBoundaryOrder(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {0, 1, 0, 2, 1, 0}, {1, 3, 0, 4, 1, 3, 0, 5}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, operations []byte) {
		const variables = 12
		order, err := NewMsatJBoundaryOrder(variables, variables)
		if err != nil {
			t.Fatal(err)
		}
		adjacency := make([][]SatVar, variables)
		for variable := 0; variable < variables; variable++ {
			left := (variable + variables - 1) % variables
			right := (variable + 1) % variables
			opposite := (variable + variables/2) % variables
			adjacency[variable] = []SatVar{SatVar(left), SatVar(right), SatVar(opposite)}
		}
		msatJConnect(t, order, adjacency)
		if err = order.SetCone(msatOrderCone(variables)); err != nil {
			t.Fatal(err)
		}
		assigned := make([]bool, variables)
		activities := make([]float64, variables)
		modelBoundary := make([]SatVar, 0, variables)
		contains := func(variable SatVar) bool { return slices.Contains(modelBoundary, variable) }
		remove := func(variable SatVar) {
			for index, current := range modelBoundary {
				if current == variable {
					modelBoundary = append(modelBoundary[:index], modelBoundary[index+1:]...)
					return
				}
			}
		}
		hasAssignedNeighbor := func(variable SatVar) bool {
			for _, other := range adjacency[variable] {
				if assigned[other] {
					return true
				}
			}
			return false
		}
		for index := 0; index+1 < len(operations); index += 2 {
			variable := SatVar(operations[index+1] % variables)
			switch operations[index] % 3 {
			case 0:
				if !assigned[variable] {
					assigned[variable] = true
					remove(variable)
					for _, other := range adjacency[variable] {
						if !assigned[other] && !contains(other) {
							modelBoundary = append(modelBoundary, other)
						}
					}
					if err = order.Assign(variable); err != nil {
						t.Fatal(err)
					}
				}
			case 1:
				if assigned[variable] {
					assigned[variable] = false
					if hasAssignedNeighbor(variable) && !contains(variable) {
						modelBoundary = append(modelBoundary, variable)
					}
					for _, other := range adjacency[variable] {
						if contains(other) && !hasAssignedNeighbor(other) {
							remove(other)
						}
					}
					if err = order.Unassign(variable); err != nil {
						t.Fatal(err)
					}
				}
			case 2:
				activity := float64(operations[index+1] % 17)
				activities[variable] = activity
				if err = order.SetActivity(variable, activity); err != nil {
					t.Fatal(err)
				}
			}
			if err = order.Check(); err != nil || !slices.Equal(order.Boundary(), modelBoundary) || !slices.Equal(order.Assignments(), assigned) || !slices.Equal(order.Activities(), activities) {
				t.Fatal(err, order.Boundary(), modelBoundary, order.Assignments(), assigned)
			}
			selected, found, selectErr := order.Select()
			if selectErr != nil || found != (len(modelBoundary) != 0) {
				t.Fatal(selected, found, selectErr, modelBoundary)
			}
			if found {
				best := modelBoundary[0]
				for _, candidate := range modelBoundary[1:] {
					if activities[candidate] > activities[best] {
						best = candidate
					}
				}
				if selected != best {
					t.Fatal(selected, best, modelBoundary, activities)
				}
			}
		}
	})
}
