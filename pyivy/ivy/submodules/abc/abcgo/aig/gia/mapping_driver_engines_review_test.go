package gia_test

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestMappingDriversWithFreshOriginalC(t *testing.T) {
	exe := gia.CompileSourceOracleForTest(t, "mapping_driver_review.c")
	cases := 0
	for seed := 1; seed <= 64; seed++ {
		for mode := 0; mode < 8; mode++ {
			t.Run(fmt.Sprintf("seed%d/mode%d", seed, mode), func(t *testing.T) {
				source, err := testoracle.Command(t, exe, strconv.Itoa(seed), strconv.Itoa(mode)).CombinedOutput()
				if mode == 6 && (seed == 25 || seed == 56) {
					// These unfiltered fixtures admit an infeasible MUX cut.
					// CheckXY returns zero, then C treats all inputs as free and
					// asserts its SAT dimensions. Keep this setup boundary explicit.
					if err == nil || !strings.Contains(string(source), "nLutSize + nSSet + nFSet <= 2*nLutSize - 1") {
						t.Fatalf("original partition boundary changed: %v %s", err, source)
					}
					t.Log("unfiltered C fixture admits an infeasible cut; the Go-only matrix enables DSD filtering")
					return
				}
				if err != nil {
					t.Fatalf("native mapping: %v %s", err, source)
				}
				src := gia.NewMappingDriverReviewGraphForTest(uint32(seed))
				defer gia.Gia_ManStop(src)
				result := mappingDriverReviewRun(t, src, seed, mode)
				if result == nil {
					t.Fatal("Go mapping failed")
				}
				defer gia.Gia_ManStop(result)
				for mask := 0; mask < 16; mask++ {
					input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
					want, err := src.Evaluate(input)
					if err != nil {
						t.Fatal(err)
					}
					got, err := result.Evaluate(input)
					if err != nil {
						t.Fatal(err)
					}
					if !slices.Equal(got, want) {
						t.Fatalf("assignment%d got%v want%v", mask, got, want)
					}
				}
				// C's LUT iterator starts at object1; a hash-collapsed constant
				// may still have an unused object0 record. Characterize it explicitly.
				if seed == 16 && mode == 1 {
					raw, err := testoracle.Command(t, exe, "16", "1", "reserved").CombinedOutput()
					if err != nil {
						t.Fatalf("source reserved record: %v %s", err, raw)
					}
					first, _, ok := strings.Cut(string(raw), "\n")
					expected := fmt.Sprintf("reserved-constant-record 2 active-luts %d", gia.Gia_ManLutNum(result))
					if !ok || first != expected || gia.Vec_IntEntry(result.VMapping, 0) != 0 {
						t.Fatalf("constant-record characterization changed: %q", first)
					}
					t.Log("source retains an unused two-leaf object0 record; Go omits it, and neither LUT iterator counts it")
				}
				got := gia.DescribeMappingDriverReviewForTest(result)
				if string(source) != got {
					t.Fatalf("complete mapping result differs\nGo:\n%s\nC:\n%s", got, source)
				}
				cases++
			})
		}
	}
	t.Logf("compared %d complete native mapping/balancing graphs, metadata, mappings and truth tables", cases)
}

func TestMappingStructuredOriginalCutMinBoundary(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "mapping_driver_review.c")
	for _, mode := range []string{"8", "9"} {
		output, err := testoracle.Command(t, executable, "1", mode).CombinedOutput()
		if err == nil || !strings.Contains(string(output), "If_DsdManSuppSize(p->pIfDsdMan, If_CutDsdLit(p, pCut)) == (int)pCut->nLeaves") {
			t.Fatalf("original unminimized DSD boundary mode%s: err%v output%s", mode, err, output)
		}
	}
}
