package espresso

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"slices"
	"testing"
)

type mincovOracleCase struct {
	Rows, Cols, Weighted, Heuristic int
	Weights                         []int32
	Masks                           []uint32
	Cost                            int32
	Cover                           []int
}

func mincovMatrix(masks []uint32, columns int) *sm_matrix {
	matrix := sm_alloc()
	for row, mask := range masks {
		for column := 0; column < columns; column++ {
			if mask>>uint(column)&1 != 0 {
				sm_insert(matrix, row, column)
			}
		}
	}
	return matrix
}
func mincovColumns(row *sm_row) []int {
	if row == nil {
		return nil
	}
	var columns []int
	for e := row.first_col; e != nil; e = e.next_col {
		columns = append(columns, e.col_num)
	}
	return columns
}
func mincovCost(row *sm_row, weights []int32) int32 {
	var cost int32
	for e := row.first_col; e != nil; e = e.next_col {
		cost += mincovWeight(weights, e.col_num)
	}
	return cost
}
func TestMinimumCoverCOracle(t *testing.T) {
	path := os.Getenv("ABC_MINCOV_ORACLE")
	if path == "" {
		path = "test_vectors/mincov_oracle.json"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cases []mincovOracleCase
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for index, tc := range cases {
		matrix := mincovMatrix(tc.Masks, tc.Cols)
		var weights []int32
		if tc.Weighted != 0 {
			weights = tc.Weights
		}
		cover := sm_minimum_cover(matrix, weights, tc.Heuristic, 0)
		if cover == nil || !verify_cover(matrix, cover) {
			t.Fatalf("case %d invalid cover", index)
		}
		columns := mincovColumns(cover)
		cost := mincovCost(cover, weights)
		if cost != tc.Cost || !slices.Equal(columns, tc.Cover) {
			t.Fatalf("case %d: cost=%d columns=%v; C cost=%d columns=%v", index, cost, columns, tc.Cost, tc.Cover)
		}
	}
}
func mincovBruteForce(masks []uint32, columns int, weights []int32) int32 {
	best := int32(math.MaxInt32)
	for chosen := uint32(0); chosen < 1<<uint(columns); chosen++ {
		valid := true
		for _, row := range masks {
			if chosen&row == 0 {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		var cost int32
		for column := 0; column < columns; column++ {
			if chosen>>uint(column)&1 != 0 {
				cost += mincovWeight(weights, column)
			}
		}
		best = min(best, cost)
	}
	return best
}
func TestMinimumCoverIndependentOptimality(t *testing.T) {
	rng := rand.New(rand.NewSource(113))
	for trial := 0; trial < 1000; trial++ {
		columns := rng.Intn(8) + 1
		rows := rng.Intn(8) + 1
		masks := make([]uint32, rows)
		for i := range masks {
			masks[i] = uint32(rng.Intn((1<<uint(columns))-1) + 1)
		}
		var weights []int32
		if trial%2 != 0 {
			weights = make([]int32, columns)
			for i := range weights {
				weights[i] = int32(rng.Intn(8))
			}
		}
		matrix := mincovMatrix(masks, columns)
		original := sm_dup(matrix)
		cover := sm_minimum_cover(matrix, weights, 0, 0)
		want := mincovBruteForce(masks, columns, weights)
		if cover == nil || !verify_cover(matrix, cover) || mincovCost(cover, weights) != want {
			t.Fatalf("trial %d masks=%v weights=%v cover=%v want=%d", trial, masks, weights, mincovColumns(cover), want)
		}
		for row := original.first_row; row != nil; row = row.next_row {
			other := sm_get_row(matrix, row.row_num)
			if other == nil || sm_row_compare(row, other) != 0 {
				t.Fatal("input matrix changed")
			}
		}
		heuristic := sm_minimum_cover(matrix, weights, 1, 0)
		if heuristic == nil || !verify_cover(matrix, heuristic) || mincovCost(heuristic, weights) < want {
			t.Fatal("invalid heuristic", trial)
		}
	}
}
func TestMinimumCoverComponentsAndSparseIDs(t *testing.T) {
	matrix := sm_alloc()
	for _, entry := range [][2]int{{7, 2}, {7, 10}, {13, 10}, {13, 27}, {99, 35}, {99, 41}, {111, 41}, {111, 50}} {
		sm_insert(matrix, entry[0], entry[1])
	}
	weights := make([]int32, 51)
	for i := range weights {
		weights[i] = 1
	}
	cover := sm_minimum_cover(matrix, weights, 0, 0)
	if !verify_cover(matrix, cover) || cover.length != 2 {
		t.Fatal(mincovColumns(cover))
	}
	empty := sm_minimum_cover(sm_alloc(), nil, 0, 0)
	if empty == nil || empty.length != 0 {
		t.Fatal("empty cover")
	}
}
func TestMinimumCoverGimpelAndIndependentBound(t *testing.T) {
	matrix := mincovMatrix([]uint32{3, 5, 10, 12}, 4)
	independent := sm_maximal_independent_set(matrix, nil)
	if independent.cost > mincovBruteForce([]uint32{3, 5, 10, 12}, 4, nil) {
		t.Fatal("inadmissible lower bound")
	}
	stats := &stats_t{}
	selected := solution_alloc()
	var reduced *solution_t
	if !gimpel_reduce(sm_dup(matrix), selected, nil, 0, 5, 0, stats, &reduced) || reduced == nil || !verify_cover(matrix, reduced.row) || stats.gimpel_count == 0 {
		t.Fatal("Gimpel reduction failed", stats)
	}
}

func TestMinimumCoverRejectsInvalidWeightCosts(t *testing.T) {
	for _, weight := range []int32{-1, math.MaxInt32} {
		t.Run(fmt.Sprint(weight), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("accepted unsupported weight bound", weight)
				}
			}()
			sm_minimum_cover(mincovMatrix([]uint32{1}, 1), []int32{weight}, 0, 0)
		})
	}
}
