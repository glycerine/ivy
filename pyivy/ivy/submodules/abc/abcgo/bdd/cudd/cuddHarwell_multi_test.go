package cudd

import (
	"io"
	"strings"
	"testing"
)

func TestHarwellReadsMatricesOneAtATime(t *testing.T) {
	first := "small matrix                 first\n" +
		"5 1 1 1 0\n" +
		"RUA 2 2 2 0\n" +
		"(3I3) (2I3) (2E10.2)\n" +
		"1 2 3\n" +
		"1 2\n" +
		"2.0 3.0\n"
	second := "small matrix                 second\n" +
		"4 1 1 1 0\n" +
		"RUA 1 1 1 0\n" +
		"(2I3) (1I3) (1E10.2)\n" +
		"1 2\n" +
		"1\n" +
		"7.5\n"
	reader := strings.NewReader(first + second)
	manager := Cudd_Init(0, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0)
	var dimensions [][2]int
	opts := testCuddOptions{multiple: 1, pr: -1}
	err := forEachHarwellMatrix(reader, manager, opts, io.Discard, func(matrix *HarwellMatrix) bool {
		dimensions = append(dimensions, [2]int{matrix.Rows, matrix.Cols})
		Cudd_RecursiveDeref(manager, matrix.Function)
		for _, variable := range append(matrix.X, matrix.Y...) {
			Cudd_RecursiveDeref(manager, variable)
		}
		for _, variable := range append(matrix.XNeg, matrix.YNeg...) {
			Cudd_RecursiveDeref(manager, variable)
		}
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(dimensions) != 2 || dimensions[0] != [2]int{2, 2} || dimensions[1] != [2]int{1, 1} {
		t.Fatalf("matrix dimensions = %v, want [[2 2] [1 1]]", dimensions)
	}
	Cudd_Quit(manager)
}
