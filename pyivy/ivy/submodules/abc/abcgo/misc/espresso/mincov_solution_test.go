package espresso

import "testing"

func TestMinimumCoverSolutionSetCost(t *testing.T) {
	selected := solution_alloc()
	weights := []int32{3, 5}
	solution_add(selected, weights, 1)
	solution_add(selected, weights, 1)
	if selected.row.length != 1 || selected.cost != 5 {
		t.Fatalf("duplicate column changed set cost: length=%d cost=%d", selected.row.length, selected.cost)
	}
}

func TestIndependentSetWeightedDisjointRows(t *testing.T) {
	matrix := sm_alloc()
	for _, entry := range [][2]int{{0, 0}, {0, 1}, {1, 2}, {1, 3}, {2, 4}, {2, 5}} {
		sm_insert(matrix, entry[0], entry[1])
	}
	for row := matrix.first_row; row != nil; row = row.next_row {
		row.flag = 7
	}
	independent := sm_maximal_independent_set(matrix, []int32{9, 2, 8, 3, 7, 4})
	if independent.cost != 9 || independent.row.length != 3 {
		t.Fatal(independent.cost, independent.row.length)
	}
	for row := matrix.first_row; row != nil; row = row.next_row {
		if row.flag != 7 {
			t.Fatal("changed caller flags")
		}
	}
}
func TestSolutionCopiesAndAcceptReject(t *testing.T) {
	matrix := sm_alloc()
	for _, entry := range [][2]int{{0, 0}, {0, 1}, {1, 1}, {1, 2}} {
		sm_insert(matrix, entry[0], entry[1])
	}
	selected := solution_alloc()
	solution_accept(selected, matrix, nil, 1)
	if selected.cost != 1 || matrix.nrows != 0 {
		t.Fatal(selected.cost, matrix.nrows)
	}
	copy := solution_dup(selected)
	solution_add(copy, nil, 3)
	if selected.row.length != 1 || copy.row.length != 2 {
		t.Fatal("solution copy shares rows")
	}
	best := solution_choose_best(selected, copy)
	if best != selected {
		t.Fatal("wrong best solution")
	}
	matrix = sm_alloc()
	sm_insert(matrix, 0, 1)
	sm_insert(matrix, 0, 2)
	solution_reject(solution_alloc(), matrix, nil, 1)
	if sm_get_col(matrix, 1) != nil || sm_get_row(matrix, 0) == nil {
		t.Fatal("reject removed uncovered row")
	}
}
