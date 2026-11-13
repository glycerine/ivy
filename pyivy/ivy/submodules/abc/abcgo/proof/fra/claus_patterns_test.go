package fra

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
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

// Compiled fraClaus.c with nClausesMax=0 selects nothing and returns the
// maximum possible simulation cost as its pivot.
func TestClausSelectionZeroLimit(t *testing.T) {
	p := &Clu_Man_t{nClausesMax: 0, nSimWords: 1, vCosts: vec.Vec_IntAlloc(3), vClauses: vec.Vec_IntAlloc(3)}
	for i, cost := range []int32{2, 5, 5} {
		vec.Vec_IntPush(p.vCosts, cost)
		vec.Vec_IntPush(p.vClauses, int32(i+1))
	}
	pivot := Fra_ClausSelectClauses(p)
	if pivot != 32 || p.nClauses != 0 || !reflect.DeepEqual(vec.Vec_IntArray(p.vCosts), []int32{-1, -1, -1}) {
		t.Fatalf("pivot=%d count=%d costs=%v", pivot, p.nClauses, vec.Vec_IntArray(p.vCosts))
	}
}
