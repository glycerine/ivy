package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/opt/fx"
)

func TestFxNativeExtractionSelection(t *testing.T) {
	exe := compileGiaSourceOracle(t, "fx_extract_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C extractor: %v\n%s", err, native)
	}
	var got strings.Builder
	for mode := 0; mode < 3; mode++ {
		for seed := uint32(1); seed <= 64; seed++ {
			for limit := int32(0); limit <= 4; limit++ {
				for litMax := int32(0); litMax <= 6; litMax += 2 {
					for canon := 0; canon < 2; canon++ {
						state := seed
						var cubes [][]int32
						for owner := int32(6); owner < 8; owner++ {
							support := []int32{0, 1, 2, 3, 4, 5}
							if mode != 0 {
								for i := 5; i > 0; i-- {
									state = state*1664525 + 1013904223
									k := int(state % uint32(i+1))
									support[i], support[k] = support[k], support[i]
								}
							}
							width := 3
							if mode == 2 {
								width = 4
							}
							slices.Sort(support[:width])
							for row := int32(0); row < 4; row++ {
								cube := []int32{owner}
								for v := int32(0); v < int32(width); v++ {
									phase := (row ^ owner) >> v & 1
									if v >= 2 {
										state = state*1664525 + 1013904223
										phase = int32(state >> 31)
									}
									cube = append(cube, 2*support[v]+phase)
								}
								cubes = append(cubes, cube)
							}
						}
						output, _, err := fx.Extract(cubes, 8, fx.Options{MaxNewNodes: limit, LiteralLimit: litMax, CanonicalOnly: canon != 0})
						if err != nil {
							t.Fatalf("seed%d limit%d: %v", seed, limit, err)
						}
						for input := 0; input < 64; input++ {
							before := fxReviewCoverValues(t, cubes, input)
							after := fxReviewCoverValues(t, output, input)
							for owner := 6; owner < 8; owner++ {
								if before[owner] != after[owner] {
									t.Fatalf("mode%d seed%d limit%d input%d owner%d", mode, seed, limit, input, owner)
								}
							}
						}
						fmt.Fprintf(&got, "mode%d seed%d limit%d litMax%d canon%d", mode, seed, limit, litMax, canon)
						for _, cube := range output {
							got.WriteString(" [")
							for i, lit := range cube {
								if i != 0 {
									got.WriteByte(',')
								}
								fmt.Fprint(&got, lit)
							}
							got.WriteByte(']')
						}
						got.WriteByte('\n')
					}
				}
			}
		}
	}
	esopReviewCompare(t, native, got.String())
}

// Evaluate the covers directly, independently of divisor or AIG algorithms.
func fxReviewCoverValues(t *testing.T, cubes [][]int32, input int) []bool {
	t.Helper()
	n := 8
	for _, cube := range cubes {
		n = max(n, int(cube[0])+1)
		for _, lit := range cube[1:] {
			n = max(n, int(lit>>1)+1)
		}
	}
	covers := make([][][]int32, n)
	for _, cube := range cubes {
		covers[cube[0]] = append(covers[cube[0]], cube[1:])
	}
	values, done, active := make([]bool, n), make([]bool, n), make([]bool, n)
	for i := 0; i < 6; i++ {
		values[i] = input>>i&1 != 0
		done[i] = true
	}
	var eval func(int32) bool
	eval = func(id int32) bool {
		if done[id] {
			return values[id]
		}
		if active[id] {
			t.Fatalf("cyclic extracted node%d", id)
		}
		active[id] = true
		value := false
		for _, row := range covers[id] {
			term := true
			for _, lit := range row {
				if eval(lit>>1) == (lit&1 != 0) {
					term = false
				}
			}
			value = value || term
		}
		values[id], done[id], active[id] = value, true, false
		return value
	}
	eval(6)
	eval(7)
	return values
}
