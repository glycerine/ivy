package abc

// Go-only model, lifecycle, and fuzz tests for extraUtilBitMatrix.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestExtraBitMatrixRelationsFormattingAndStatistics(t *testing.T) {
	matrix, err := NewExtraBitMatrix(5, ExtraBitMatrixLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, edge := range [][2]int{{0, 1}, {1, 2}, {0, 2}, {3, 4}} {
		if err = matrix.InsertUpper(edge[1], edge[0]); err != nil {
			t.Fatal(edge, err)
		}
	}
	if err = matrix.InsertLower(0, 4); err != nil {
		t.Fatal(err)
	}
	if upper, _ := matrix.LookupUpper(2, 0); !upper {
		t.Fatal("missing upper edge")
	}
	if lower, _ := matrix.LookupLower(0, 4); !lower {
		t.Fatal("missing lower edge")
	}
	count, err := matrix.CountOnesUpper(ExtraBitMatrixLimits{})
	if err != nil || count != 4 {
		t.Fatal(count, err)
	}
	formatted, err := matrix.FormatUpper(ExtraBitMatrixLimits{})
	want := "\n 11..\n  1..\n   ..\n    1\n     \n"
	if err != nil || formatted != want {
		t.Fatal(formatted, want, err)
	}
	clique, err := matrix.IsClique(ExtraBitMatrixLimits{})
	if err != nil || !clique {
		t.Fatal(clique, err)
	}
	if err = matrix.DeleteUpper(0, 2); err != nil {
		t.Fatal(err)
	}
	clique, err = matrix.IsClique(ExtraBitMatrixLimits{})
	if err != nil || clique {
		t.Fatal(clique, err)
	}
	stats, _ := matrix.Statistics()
	if stats.Inserts != 5 || stats.Deletes != 1 || stats.Lookups == 0 {
		t.Fatal(stats)
	}
}

func TestExtraBitMatrixRowsDisjointClearOwnershipAndClose(t *testing.T) {
	first, _ := NewExtraBitMatrix(35, ExtraBitMatrixLimits{})
	second, _ := NewExtraBitMatrix(35, ExtraBitMatrixLimits{})
	words := []uint32{1 << 3, 1 << 2}
	if err := first.OrRow(1, words, ExtraBitMatrixLimits{}); err != nil {
		t.Fatal(err)
	}
	words[0] = 0
	rows, _ := first.SnapshotRows()
	if rows[1][0] != 1<<3 || ExtraBitMatrixCountWords(rows[1]) != 2 {
		t.Fatal(rows[1])
	}
	rows[1][0] = 0
	again, _ := first.SnapshotRows()
	if again[1][0] != 1<<3 {
		t.Fatal("snapshot aliases matrix")
	}
	if disjoint, err := first.Disjoint(second, ExtraBitMatrixLimits{}); err != nil || !disjoint {
		t.Fatal(disjoint, err)
	}
	if err := second.OrRow(2, []uint32{1 << 3, 0}, ExtraBitMatrixLimits{}); err != nil {
		t.Fatal(err)
	}
	if err := first.OrRows(1, 2, ExtraBitMatrixLimits{}); err != nil {
		t.Fatal(err)
	}
	rows, _ = first.SnapshotRows()
	if !slices.Equal(rows[1], rows[2]) {
		t.Fatal(rows[1], rows[2])
	}
	if disjoint, err := first.Disjoint(second, ExtraBitMatrixLimits{}); err != nil || disjoint {
		t.Fatal(disjoint, err)
	}
	if err := first.Clear(ExtraBitMatrixLimits{}); err != nil {
		t.Fatal(err)
	}
	rows, _ = first.SnapshotRows()
	for _, row := range rows {
		if ExtraBitMatrixCountWords(row) != 0 {
			t.Fatal(row)
		}
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := first.Size(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := first.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestExtraBitMatrixErrorsLimitsCancellationAndAtomicity(t *testing.T) {
	for _, size := range []int{-1, 100} {
		limit := ExtraBitMatrixLimits{}
		if size == 100 {
			limit.MaxWords = 1
		}
		if _, err := NewExtraBitMatrix(size, limit); err == nil {
			t.Fatal(size)
		}
	}
	if _, err := NewExtraBitMatrix(2, ExtraBitMatrixLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	matrix, _ := NewExtraBitMatrix(8, ExtraBitMatrixLimits{})
	if err := matrix.InsertUpper(-1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := matrix.OrRow(0, []uint32{}, ExtraBitMatrixLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := matrix.FormatUpper(ExtraBitMatrixLimits{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	before, _ := matrix.SnapshotRows()
	if _, err := matrix.IsClique(ExtraBitMatrixLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	after, _ := matrix.SnapshotRows()
	if !slices.EqualFunc(before, after, func(first, second []uint32) bool { return slices.Equal(first, second) }) {
		t.Fatal("cancelled scan mutated matrix")
	}
	other, _ := NewExtraBitMatrix(7, ExtraBitMatrixLimits{})
	if _, err := matrix.Disjoint(other, ExtraBitMatrixLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func extraBitMatrixModelClique(model [][]bool) bool {
	for first := range model {
		for second := first + 1; second < len(model); second++ {
			if !model[first][second] {
				continue
			}
			for candidate := range model {
				if candidate != first && candidate != second && model[min(candidate, first)][max(candidate, first)] != model[min(candidate, second)][max(candidate, second)] {
					return false
				}
			}
		}
	}
	return true
}

func FuzzExtraBitMatrix(f *testing.F) {
	f.Add(byte(5), []byte{0, 1, 2, 1, 2, 3, 0, 2, 4})
	f.Add(byte(1), []byte{0, 0, 3})
	f.Fuzz(func(t *testing.T, rawSize byte, operations []byte) {
		size := int(rawSize%16) + 1
		matrix, err := NewExtraBitMatrix(size, ExtraBitMatrixLimits{})
		if err != nil {
			t.Fatal(err)
		}
		model := make([][]bool, size)
		for row := range model {
			model[row] = make([]bool, size)
		}
		for index := 0; index+2 < len(operations); index += 3 {
			first, second := int(operations[index])%size, int(operations[index+1])%size
			row, column := min(first, second), max(first, second)
			switch operations[index+2] % 4 {
			case 0:
				err = matrix.InsertUpper(first, second)
				model[row][column] = true
			case 1:
				err = matrix.DeleteUpper(first, second)
				model[row][column] = false
			case 2:
				var got bool
				got, err = matrix.LookupUpper(first, second)
				if got != model[row][column] {
					t.Fatal(size, first, second, got, model[row][column])
				}
			case 3:
				var got bool
				got, err = matrix.IsClique(ExtraBitMatrixLimits{MaxOperations: 1 << 20})
				if got != extraBitMatrixModelClique(model) {
					t.Fatal(size, got, model)
				}
			}
			if err != nil {
				t.Fatal(size, first, second, err)
			}
		}
		count, err := matrix.CountOnesUpper(ExtraBitMatrixLimits{})
		want := 0
		for row := range model {
			for column := row + 1; column < size; column++ {
				if model[row][column] {
					want++
				}
			}
		}
		if err != nil || count != want {
			t.Fatal(size, count, want, err)
		}
	})
}
