package espresso

import (
	_ "embed"
	"encoding/json"
	"slices"
	"testing"
)

//go:embed test_vectors/sparse_oracle.json
var sparseOracleJSON []byte

type sparseOracleExpected struct {
	Entries                     [][2]int `json:"entries"`
	RowCount                    int      `json:"row_count"`
	ColCount                    int      `json:"col_count"`
	LongestRow                  int      `json:"longest_row"`
	LongestCol                  int      `json:"longest_col"`
	Subset, Intersects, Compare int
	ColSubset                   int `json:"col_subset"`
	ColIntersects               int `json:"col_intersects"`
	ColCompare                  int `json:"col_compare"`
	Hash                        int
	Intersection                []int
	RemovedRows                 int `json:"removed_rows"`
	RemovedCols                 int `json:"removed_cols"`
	Dominated                   [][2]int
	WeightedRemoved             int      `json:"weighted_removed"`
	WeightedDominated           [][2]int `json:"weighted_dominated"`
	Split                       int
	Left, Right                 [][2]int
}
type sparseOracleCase struct {
	Operations    [][3]int
	Weights       []int32
	First, Second []int
	Expected      sparseOracleExpected
}

func sparseCoordinates(matrix *sm_matrix) [][2]int {
	var result [][2]int
	for row := matrix.first_row; row != nil; row = row.next_row {
		for element := row.first_col; element != nil; element = element.next_col {
			result = append(result, [2]int{element.row_num, element.col_num})
		}
	}
	return result
}
func TestSparseOriginalCOracle(t *testing.T) {
	var cases []sparseOracleCase
	if err := json.Unmarshal(sparseOracleJSON, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 160 {
		t.Fatal("unexpected C fixture count", len(cases))
	}
	for index, test := range cases {
		matrix := sm_alloc()
		for _, op := range test.Operations {
			switch byte(op[0]) {
			case 'I':
				sm_insert(matrix, op[1], op[2])
			case 'R':
				sm_remove(matrix, op[1], op[2])
			case 'D':
				sm_delrow(matrix, op[1])
			case 'C':
				sm_delcol(matrix, op[2])
			default:
				t.Fatal("unknown fixture operation")
			}
		}
		expected := test.Expected
		if !slices.Equal(sparseCoordinates(matrix), expected.Entries) || matrix.nrows != expected.RowCount || matrix.ncols != expected.ColCount {
			t.Fatal("C matrix mismatch", index)
		}
		rowID, colID := -1, -1
		if row := sm_longest_row(matrix); row != nil {
			rowID = row.row_num
		}
		if column := sm_longest_col(matrix); column != nil {
			colID = column.col_num
		}
		if rowID != expected.LongestRow || colID != expected.LongestCol {
			t.Fatal("C longest-header mismatch", index)
		}
		a, b := sm_row_alloc(), sm_row_alloc()
		ca, cb := sm_col_alloc(), sm_col_alloc()
		for _, value := range test.First {
			sm_row_insert(a, value)
			sm_col_insert(ca, value)
		}
		for _, value := range test.Second {
			sm_row_insert(b, value)
			sm_col_insert(cb, value)
		}
		if sm_row_contains(a, b) != (expected.Subset != 0) || sm_row_intersects(a, b) != (expected.Intersects != 0) || sm_row_compare(a, b) != expected.Compare || sm_row_hash(a, 1009) != expected.Hash {
			t.Fatal("C row-vector mismatch", index)
		}
		if sm_col_contains(ca, cb) != (expected.ColSubset != 0) || sm_col_intersects(ca, cb) != (expected.ColIntersects != 0) || sm_col_compare(ca, cb) != expected.ColCompare || sm_col_hash(ca, 1009) != expected.Hash {
			t.Fatal("C column-vector mismatch", index)
		}
		if got := rowValues(sm_row_and(a, b)); !slices.Equal(got, expected.Intersection) {
			t.Fatal("C intersection mismatch", index, got)
		}
		duplicate := sm_dup(matrix)
		if got := sm_row_dominance(duplicate); got != expected.RemovedRows {
			t.Fatal("C row dominance count", index, got, expected.RemovedRows)
		}
		if got := sm_col_dominance(duplicate, nil); got != expected.RemovedCols || !slices.Equal(sparseCoordinates(duplicate), expected.Dominated) {
			t.Fatal("C column dominance mismatch", index)
		}
		weighted := sm_dup(matrix)
		if got := sm_col_dominance(weighted, test.Weights); got != expected.WeightedRemoved || !slices.Equal(sparseCoordinates(weighted), expected.WeightedDominated) {
			t.Fatal("C weighted dominance mismatch", index)
		}
		var left, right *sm_matrix
		split := sm_block_partition(matrix, &left, &right)
		if split != (expected.Split != 0) {
			t.Fatal("C partition decision", index)
		}
		if split && (!slices.Equal(sparseCoordinates(left), expected.Left) || !slices.Equal(sparseCoordinates(right), expected.Right)) {
			t.Fatal("C partition mismatch", index)
		}
		sparseCheck(t, matrix, sparseSnapshot(matrix))
	}
}
