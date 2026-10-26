package abc

// Go-only model and fuzz tests for Espresso sparse rows, columns, and matrices.

import (
	"errors"
	"math"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestEspressoSparseRowAndColumnOperations(t *testing.T) {
	row := NewEspressoSparseRow()
	row.Number, row.Flag, row.Tag = 7, 8, 9
	for _, column := range []int{5, 1, 3, 3} {
		if _, err := row.Insert(column); err != nil {
			t.Fatal(err)
		}
	}
	values, _ := row.Values()
	if !slices.Equal(values, []int{1, 3, 5}) || row.Len() != 3 || row.String() != " 1 3 5" {
		t.Fatal(values, row.Len(), row.String())
	}
	if inserted, err := row.Insert(3); err != nil || inserted {
		t.Fatal(inserted, err)
	}
	if found, err := row.Has(3); err != nil || !found {
		t.Fatal(found, err)
	}
	other := NewEspressoSparseRow()
	for _, column := range []int{3, 5} {
		_, _ = other.Insert(column)
	}
	contains, err := row.Contains(other)
	intersects, intersectErr := row.Intersects(other)
	comparison, compareErr := row.Compare(other)
	intersection, andErr := row.Intersection(other)
	intersectionValues, _ := intersection.Values()
	if err != nil || intersectErr != nil || compareErr != nil || andErr != nil || !contains || !intersects || comparison >= 0 || !slices.Equal(intersectionValues, []int{3, 5}) {
		t.Fatal(contains, intersects, comparison, intersectionValues, err, intersectErr, compareErr, andErr)
	}
	hash, err := row.Hash(97)
	if err != nil || hash != ((1*17+3)*17+5)%97 {
		t.Fatal(hash, err)
	}
	clone, err := row.Clone()
	if err != nil || clone.Number != 7 || clone.Flag != 8 || clone.Tag != 9 {
		t.Fatal(clone, err)
	}
	_, _ = clone.Remove(3)
	values, _ = row.Values()
	if !slices.Equal(values, []int{1, 3, 5}) {
		t.Fatal("clone aliases row", values)
	}
	if removed, err := row.Remove(4); err != nil || removed {
		t.Fatal(removed, err)
	}
	if _, err := row.Insert(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := row.Hash(0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	wideHashRow := NewEspressoSparseRow()
	_, _ = wideHashRow.Insert(math.MaxInt - 1)
	_, _ = wideHashRow.Insert(math.MaxInt)
	if hash, err = wideHashRow.Hash(math.MaxInt); err != nil || hash < 0 || hash >= math.MaxInt {
		t.Fatal(hash, err)
	}

	column := NewEspressoSparseColumn()
	column.Number, column.Flag, column.Tag = 11, 12, 13
	for _, rowID := range []int{9, 2, 6, 6} {
		if _, err = column.Insert(rowID); err != nil {
			t.Fatal(err)
		}
	}
	columnValues, _ := column.Values()
	if !slices.Equal(columnValues, []int{2, 6, 9}) || column.String() != " 2 6 9" {
		t.Fatal(columnValues, column.String())
	}
	columnOther := NewEspressoSparseColumn()
	_, _ = columnOther.Insert(6)
	if yes, _ := column.Contains(columnOther); !yes {
		t.Fatal("column containment")
	}
	if yes, _ := column.Intersects(columnOther); !yes {
		t.Fatal("column intersection")
	}
	columnIntersection, err := column.Intersection(columnOther)
	columnIntersectionValues, _ := columnIntersection.Values()
	if err != nil || !slices.Equal(columnIntersectionValues, []int{6}) {
		t.Fatal(columnIntersectionValues, err)
	}
	columnClone, err := column.Clone()
	if err != nil || columnClone.Number != 11 || columnClone.Flag != 12 || columnClone.Tag != 13 {
		t.Fatal(columnClone, err)
	}
}

func TestEspressoSparseMatrixLifecycleMutationAndMetadata(t *testing.T) {
	matrix := NewEspressoSparseMatrix()
	matrix.Tag = 77
	for _, cell := range [][2]int{{2, 3}, {1, 4}, {2, 1}, {5, 3}, {2, 3}} {
		if _, err := matrix.Insert(cell[0], cell[1]); err != nil {
			t.Fatal(err)
		}
	}
	if err := matrix.Validate(); err != nil {
		t.Fatal(err)
	}
	rows, columns, entries, err := matrix.Counts()
	rowIDs, _ := matrix.RowIDs()
	columnIDs, _ := matrix.ColumnIDs()
	if err != nil || rows != 3 || columns != 3 || entries != 4 || !slices.Equal(rowIDs, []int{1, 2, 5}) || !slices.Equal(columnIDs, []int{1, 3, 4}) {
		t.Fatal(rows, columns, entries, rowIDs, columnIDs, err)
	}
	row, found, err := matrix.Row(2)
	rowValues, _ := row.Values()
	column, columnFound, columnErr := matrix.Column(3)
	columnValues, _ := column.Values()
	if err != nil || columnErr != nil || !found || !columnFound || !slices.Equal(rowValues, []int{1, 3}) || !slices.Equal(columnValues, []int{2, 5}) {
		t.Fatal(rowValues, columnValues, found, columnFound, err, columnErr)
	}
	if err = matrix.SetRowMetadata(2, 10, 20); err != nil {
		t.Fatal(err)
	}
	if err = matrix.SetColumnMetadata(3, 30, 40); err != nil {
		t.Fatal(err)
	}
	if err = matrix.SetCellTag(2, 3, 50); err != nil {
		t.Fatal(err)
	}
	cellTag, tagged, err := matrix.CellTag(2, 3)
	row, _, _ = matrix.Row(2)
	column, _, _ = matrix.Column(3)
	if err != nil || !tagged || cellTag != 50 || row.Flag != 10 || row.Tag != 20 || column.Flag != 30 || column.Tag != 40 {
		t.Fatal(cellTag, tagged, row, column, err)
	}
	longRow, found, err := matrix.LongestRow()
	longColumn, columnFound, columnErr := matrix.LongestColumn()
	if err != nil || columnErr != nil || !found || !columnFound || longRow.Number != 2 || longColumn.Number != 3 {
		t.Fatal(longRow, longColumn, err, columnErr)
	}
	clone, err := matrix.Clone()
	if err != nil || clone.Tag != 77 || clone.Validate() != nil {
		t.Fatal(err)
	}
	if removed, err := matrix.Remove(2, 1); err != nil || !removed {
		t.Fatal(removed, err)
	}
	if deleted, err := matrix.DeleteRow(1); err != nil || !deleted {
		t.Fatal(deleted, err)
	}
	if deleted, err := matrix.DeleteColumn(3); err != nil || !deleted {
		t.Fatal(deleted, err)
	}
	if err = matrix.Validate(); err != nil {
		t.Fatal(err)
	}
	rows, columns, entries, _ = matrix.Counts()
	if rows != 0 || columns != 0 || entries != 0 {
		t.Fatal(rows, columns, entries)
	}
	cloneRows, cloneColumns, cloneEntries, _ := clone.Counts()
	if cloneRows != 3 || cloneColumns != 3 || cloneEntries != 4 {
		t.Fatal("clone aliases matrix", cloneRows, cloneColumns, cloneEntries)
	}
	if err = clone.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err = clone.Counts(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = matrix.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestEspressoSparseMatrixCopyTextFormatAndCompressed(t *testing.T) {
	source, err := ParseEspressoSparseMatrix([]byte("2 1\n2 105\n7 12\n"), EspressoSparseLimits{})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := source.MarshalText()
	if err != nil || string(encoded) != "2 1\n2 105\n7 12\n" {
		t.Fatal(string(encoded), err)
	}
	formatted, err := source.Format()
	wantFormat := "    001\n    010\n    125\n    ---\n  2:1.1\n  7:.1.\n"
	if err != nil || formatted != wantFormat {
		t.Fatalf("got %q want %q err=%v", formatted, wantFormat, err)
	}
	dump, err := source.Dump("matrix", 3)
	if err != nil || dump != "matrix 2 rows by 3 cols\n"+wantFormat {
		t.Fatal(dump, err)
	}
	row, _, _ := source.Row(2)
	column, _, _ := source.Column(12)
	destination := NewEspressoSparseMatrix()
	if err = destination.CopyRow(9, row); err != nil {
		t.Fatal(err)
	}
	if err = destination.CopyColumnToRow(10, column); err != nil {
		t.Fatal(err)
	}
	destinationRow, _, _ := destination.Row(10)
	destinationValues, _ := destinationRow.Values()
	if !slices.Equal(destinationValues, []int{7}) {
		t.Fatal(destinationValues)
	}

	compressed := []byte("2 35 0 5 2 0 80000000 4")
	decoded, err := ParseEspressoCompressedSparseMatrix(compressed, EspressoSparseLimits{})
	if err != nil {
		t.Fatal(err)
	}
	want, err := ParseEspressoSparseMatrix([]byte("0 0\n0 2\n0 33\n1 31\n1 34\n"), EspressoSparseLimits{})
	if err != nil {
		t.Fatal(err)
	}
	decodedText, _ := decoded.MarshalText()
	wantText, _ := want.MarshalText()
	if string(decodedText) != string(wantText) || decoded.Validate() != nil {
		t.Fatal(string(decodedText), string(wantText))
	}
	for _, matrix := range []*EspressoSparseMatrix{source, destination, decoded, want} {
		if err = matrix.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEspressoSparseMatrixErrorsAndLimits(t *testing.T) {
	if _, err := NewEspressoSparseMatrixSize(-1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	badText := [][]byte{[]byte("1"), []byte("a 2"), []byte("-1 2")}
	for _, data := range badText {
		if _, err := ParseEspressoSparseMatrix(data, EspressoSparseLimits{}); !errors.Is(err, ErrFormat) {
			t.Fatal(string(data), err)
		}
	}
	if _, err := ParseEspressoSparseMatrix([]byte("2 3"), EspressoSparseLimits{MaxRows: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := ParseEspressoSparseMatrix([]byte("0 0\n0 1"), EspressoSparseLimits{MaxEntries: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := ParseEspressoSparseMatrix([]byte("0 0"), EspressoSparseLimits{MaxBytes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	badCompressed := [][]byte{
		[]byte(""), []byte("x 2"), []byte("1 33 0 1"), []byte("1 33 0 0 2"), []byte("1 1 x 1"),
	}
	for _, data := range badCompressed {
		if _, err := ParseEspressoCompressedSparseMatrix(data, EspressoSparseLimits{}); !errors.Is(err, ErrFormat) {
			t.Fatal(string(data), err)
		}
	}
	matrix := NewEspressoSparseMatrix()
	if _, err := matrix.Insert(-1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := matrix.SetCellTag(1, 1, 1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := matrix.Dump("x", -1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := matrix.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := matrix.Insert(0, 0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzEspressoSparseMatrixState(f *testing.F) {
	f.Add([]byte{0, 2, 3, 0, 1, 4, 0, 2, 1, 1, 2, 3, 3, 2, 0})
	f.Add([]byte{0, 31, 31, 0, 0, 0, 2, 31, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		matrix := NewEspressoSparseMatrix()
		model := make(map[espressoSparseCell]bool)
		for index := 0; index+2 < len(data) && index < 300; index += 3 {
			operation := data[index] % 4
			row, column := int(data[index+1]%32), int(data[index+2]%32)
			switch operation {
			case 0:
				inserted, err := matrix.Insert(row, column)
				if err != nil || inserted == model[espressoSparseCell{row, column}] {
					t.Fatal(operation, row, column, inserted, err)
				}
				model[espressoSparseCell{row, column}] = true
			case 1:
				removed, err := matrix.Remove(row, column)
				if err != nil || removed != model[espressoSparseCell{row, column}] {
					t.Fatal(operation, row, column, removed, err)
				}
				delete(model, espressoSparseCell{row, column})
			case 2:
				_, err := matrix.DeleteRow(row)
				if err != nil {
					t.Fatal(err)
				}
				for cell := range model {
					if cell.row == row {
						delete(model, cell)
					}
				}
			case 3:
				_, err := matrix.DeleteColumn(column)
				if err != nil {
					t.Fatal(err)
				}
				for cell := range model {
					if cell.column == column {
						delete(model, cell)
					}
				}
			}
			if err := matrix.Validate(); err != nil {
				t.Fatal(index, err)
			}
		}
		encoded, err := matrix.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		want := make([]espressoSparseCell, 0, len(model))
		for cell := range model {
			want = append(want, cell)
		}
		sort.Slice(want, func(first, second int) bool {
			return want[first].row < want[second].row || want[first].row == want[second].row && want[first].column < want[second].column
		})
		lines := strings.Fields(string(encoded))
		if len(lines) != 2*len(want) {
			t.Fatal(string(encoded), want)
		}
		clone, err := matrix.Clone()
		if err != nil || clone.Validate() != nil {
			t.Fatal(err)
		}
		reparsed, err := ParseEspressoSparseMatrix(encoded, EspressoSparseLimits{})
		if err != nil || reparsed.Validate() != nil {
			t.Fatal(err)
		}
		reparsedText, _ := reparsed.MarshalText()
		cloneText, _ := clone.MarshalText()
		if string(reparsedText) != string(encoded) || string(cloneText) != string(encoded) {
			t.Fatal(string(encoded), string(reparsedText), string(cloneText))
		}
		_ = reparsed.Close()
		_ = clone.Close()
		_ = matrix.Close()
	})
}

func FuzzParseEspressoCompressedSparseMatrix(f *testing.F) {
	f.Add([]byte("2 35 0 5 2 0 80000000 4"))
	f.Add([]byte("0 0"))
	f.Fuzz(func(t *testing.T, data []byte) {
		matrix, err := ParseEspressoCompressedSparseMatrix(data, EspressoSparseLimits{MaxRows: 64, MaxColumns: 128, MaxEntries: 4096, MaxBytes: 4096})
		if err != nil {
			return
		}
		if err = matrix.Validate(); err != nil {
			t.Fatal(err)
		}
		encoded, err := matrix.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		reparsed, err := ParseEspressoSparseMatrix(encoded, EspressoSparseLimits{})
		if err != nil || reparsed.Validate() != nil {
			t.Fatal(err)
		}
		again, _ := reparsed.MarshalText()
		if string(again) != string(encoded) {
			t.Fatal("noncanonical sparse text round trip")
		}
		_ = reparsed.Close()
		_ = matrix.Close()
	})
}
