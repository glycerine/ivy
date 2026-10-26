package abc

// Go-only state-model and fuzz tests for src/misc/espresso/solution.c.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func TestMvcCoverSolutionLifecycleAddCloneAndChoice(t *testing.T) {
	solution, err := NewMvcCoverSolution(5)
	if err != nil {
		t.Fatal(err)
	}
	if err = solution.Add(3, nil); err != nil {
		t.Fatal(err)
	}
	if err = solution.Add(1, []int{2, 7, 3, 4, 5}); err != nil {
		t.Fatal(err)
	}
	cost, _ := solution.Cost()
	columns, _ := solution.Columns()
	if cost != 8 || !slices.Equal(columns, []int{1, 3}) {
		t.Fatal(cost, columns)
	}
	clone, err := solution.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if err = clone.Add(0, nil); err != nil {
		t.Fatal(err)
	}
	columns, _ = solution.Columns()
	if !slices.Equal(columns, []int{1, 3}) {
		t.Fatal("clone aliases source", columns)
	}
	winner, err := TakeBetterCoverSolution(solution, clone)
	if err != nil || winner != solution {
		t.Fatal(winner, err)
	}
	if _, err = clone.Cost(); !errors.Is(err, ErrInvariant) {
		t.Fatal("loser remained live", err)
	}
	if err = winner.Close(); err != nil {
		t.Fatal(err)
	}

	first, _ := NewMvcCoverSolution(1)
	second, _ := NewMvcCoverSolution(1)
	winner, err = TakeBetterCoverSolution(first, second)
	if err != nil || winner != first {
		t.Fatal(winner, err)
	}
	if _, err = second.Columns(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	first.Close()
	if winner, err = TakeBetterCoverSolution(nil, nil); winner != nil || err != nil {
		t.Fatal(winner, err)
	}
}

func TestMvcCoverSolutionAcceptRejectAndOwnership(t *testing.T) {
	rows := []uint64{0b011, 0b110, 0b100, 0b001}
	manager, cover := mvcMinimumCoverFixture(t, 3, rows)
	solution, _ := NewMvcCoverSolution(3)
	accepted, err := solution.Accept(cover, 1, []int{3, 5, 7}, MvcCoverSolutionOptions{})
	if err != nil || !slices.Equal(mvcDominanceMasks(t, accepted), []uint64{0b100, 0b001}) {
		t.Fatal(mvcDominanceMasks(t, accepted), err)
	}
	cost, _ := solution.Cost()
	columns, _ := solution.Columns()
	if cost != 5 || !slices.Equal(columns, []int{1}) {
		t.Fatal(cost, columns)
	}
	rejected, err := solution.Reject(cover, 0, MvcCoverSolutionOptions{})
	if err != nil || !slices.Equal(mvcDominanceMasks(t, rejected), []uint64{0b010, 0b110, 0b100}) {
		t.Fatal(mvcDominanceMasks(t, rejected), err)
	}
	if !slices.Equal(mvcDominanceMasks(t, cover), rows) {
		t.Fatal("source cover mutated")
	}
	rejected.cubes[0].SetBit(2, true)
	if !slices.Equal(mvcDominanceMasks(t, cover), rows) {
		t.Fatal("result aliases source")
	}
	accepted.Close()
	rejected.Close()
	solution.Close()
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func TestMvcCoverSolutionErrorsLimitsAndAtomicCancellation(t *testing.T) {
	if solution, err := NewMvcCoverSolutionWithLimit(-1, 1); solution != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(solution, err)
	}
	if solution, err := NewMvcCoverSolutionWithLimit(2, 1); solution != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(solution, err)
	}
	solution, _ := NewMvcCoverSolution(3)
	for _, test := range []struct {
		column  int
		weights []int
	}{{-1, nil}, {3, nil}, {0, []int{1}}, {0, []int{-1, 1, 1}}} {
		if err := solution.Add(test.column, test.weights); !errors.Is(err, ErrDimension) {
			t.Fatal(test, err)
		}
	}
	if err := solution.Add(0, nil); err != nil {
		t.Fatal(err)
	}
	if err := solution.Add(0, nil); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	overflow, _ := NewMvcCoverSolution(1)
	overflow.cost = math.MaxInt
	if err := overflow.Add(0, nil); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	overflow.Close()

	manager, cover := mvcMinimumCoverFixture(t, 3, []uint64{0b011, 0b110})
	beforeCost, _ := solution.Cost()
	for _, options := range []MvcCoverSolutionOptions{{MaxRows: -1}, {MaxEntries: -1}, {MaxRows: 1}, {MaxEntries: 1}, {Stop: func() bool { return true }}} {
		result, err := solution.Accept(cover, 2, nil, options)
		if result != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
		cost, _ := solution.Cost()
		if cost != beforeCost {
			t.Fatal("failed accept mutated solution", beforeCost, cost)
		}
	}
	if result, err := solution.Accept(cover, 0, nil, MvcCoverSolutionOptions{}); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	if result, err := solution.Accept(cover, 2, nil, MvcCoverSolutionOptions{}); err != nil {
		t.Fatal(err)
	} else {
		result.Close()
	}
	if result, err := solution.Reject(cover, 3, MvcCoverSolutionOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	solution.Close()
	mvcCloseMinimumFixture(t, manager, cover, nil)
}

func FuzzMvcCoverSolution(f *testing.F) {
	f.Add(uint8(5), []byte{3, 7, 12, 17, 24}, []byte{2, 5, 3, 1, 4}, uint8(2))
	f.Fuzz(func(t *testing.T, widthByte uint8, data, weightData []byte, columnByte uint8) {
		columns := 1 + int(widthByte%8)
		if len(data) > 12 {
			data = data[:12]
		}
		limit := byte((1 << uint(columns)) - 1)
		rows := make([]uint64, len(data))
		for index, value := range data {
			rows[index] = uint64(value & limit)
		}
		weights := make([]int, columns)
		for column := range weights {
			if len(weightData) == 0 {
				weights[column] = 1
			} else {
				weights[column] = int(weightData[column%len(weightData)] % 8)
			}
		}
		column := int(columnByte) % columns
		manager, cover := mvcMinimumCoverFixture(t, columns, rows)
		solution, _ := NewMvcCoverSolution(columns)
		present := false
		wantAccept := make([]uint64, 0, len(rows))
		wantReject := make([]uint64, 0, len(rows))
		for _, row := range rows {
			hasColumn := row>>uint(column)&1 != 0
			if hasColumn {
				present = true
			} else {
				wantAccept = append(wantAccept, row)
			}
			cleared := row &^ (uint64(1) << uint(column))
			if !hasColumn || cleared != 0 {
				wantReject = append(wantReject, cleared)
			}
		}
		rejected, err := solution.Reject(cover, column, MvcCoverSolutionOptions{MaxRows: 100, MaxEntries: 1000})
		if err != nil || !slices.Equal(mvcDominanceMasks(t, rejected), wantReject) {
			t.Fatal(mvcDominanceMasks(t, rejected), wantReject, err)
		}
		rejected.Close()
		accepted, err := solution.Accept(cover, column, weights, MvcCoverSolutionOptions{MaxRows: 100, MaxEntries: 1000})
		if !present {
			if accepted != nil || !errors.Is(err, ErrDimension) {
				t.Fatal(accepted, err)
			}
		} else {
			if err != nil || !slices.Equal(mvcDominanceMasks(t, accepted), wantAccept) {
				t.Fatal(mvcDominanceMasks(t, accepted), wantAccept, err)
			}
			cost, _ := solution.Cost()
			selected, _ := solution.Columns()
			if cost != weights[column] || !slices.Equal(selected, []int{column}) {
				t.Fatal(cost, selected, weights[column])
			}
			accepted.Close()
		}
		solution.Close()
		mvcCloseMinimumFixture(t, manager, cover, nil)
	})
}
