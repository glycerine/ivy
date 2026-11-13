package fra

import (
	"reflect"
	"testing"
)

func TestClausTransposeAndPatterns(t *testing.T) {
	var matrix [32]uint32
	matrix[0] = 0x80000000
	matrix[10] = 1 << 4
	original := matrix
	clausTranspose32(&matrix)
	clausTranspose32(&matrix)
	if matrix != original {
		t.Fatal("transpose is not involutive")
	}
	leaves := [][]uint32{{0, 0xaaaaaaaa}, {0, 0xcccccccc}}
	if got := clausPatternScores(leaves, 1); !reflect.DeepEqual(got, []int32{8, 8, 8, 8}) {
		t.Fatal(got)
	}
}
func TestClausSelectCosts(t *testing.T) {
	costs := []int32{3, 8, 8, 1, 8, -1}
	threshold, n := clausSelectCosts(costs, 2, 1)
	if threshold != 8 || n != 2 || !reflect.DeepEqual(costs, []int32{-1, 8, 8, -1, -1, -1}) {
		t.Fatalf("threshold %d count %d costs %v", threshold, n, costs)
	}
}
