package gia_test

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/sat/bsat"
	"os/exec"
	"strings"
	"testing"
)

func TestMappingSATNativeXYZClauseSemantics(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_sat_xyz_review.c")
	want, err := exec.Command(executable).CombinedOutput()
	if err != nil {
		t.Fatalf("original XYZ: %v %s", err, want)
	}
	state := uint64(7654321)
	next := func() uint64 { state = state*6364136223846793005 + 1442695040888963407; return state }
	var got strings.Builder
	for k := int32(2); k <= 6; k++ {
		handle := if_.If_ManSatBuildXYZ(k)
		solver := handle.(*bsat.Solver)
		l, f := int32(1)<<uint(k), int32(1)<<uint(3*k-2)
		for trial := 0; trial < 12; trial++ {
			tables := [3]uint64{next(), next(), next()}
			for j := 0; j < 8; j++ {
				m := int32(0)
				if j == 1 {
					m = f - 1
				} else if j > 1 {
					m = int32(next() % uint64(f))
				}
				for flipped := int32(0); flipped < 2; flipped++ {
					assumptions := make([]int32, 0, 3*l+1)
					for p := int32(0); p < 3; p++ {
						for v := int32(0); v < l; v++ {
							bit := int32(tables[p] >> uint(v) & 1)
							assumptions = append(assumptions, 2*(p*l+v)+(1-bit))
						}
					}
					select0 := int32(tables[0] >> uint(m%l) & 1)
					select1 := int32(tables[1] >> uint((m>>uint(k))%l) & 1)
					pin := 4*(m>>uint(2*k)) + select0 + 2*select1
					bit := int32(tables[2]>>uint(pin)&1) ^ flipped
					assumptions = append(assumptions, 2*(3*l+m)+(1-bit))
					result := solver.SolveWithResourceLimits(assumptions, 0, 0, 0, 0)
					expected := bsat.Satisfiable
					if flipped != 0 {
						expected = bsat.Unsatisfiable
					}
					if result != expected {
						t.Fatalf("K%d trial%d minterm%d flipped%d status%v", k, trial, m, flipped, result)
					}
					nativeStatus := -1
					if result == bsat.Satisfiable {
						nativeStatus = 1
					}
					fmt.Fprintf(&got, "xyz %d %d %d %d %d\n", k, trial, m, flipped, nativeStatus)
				}
			}
		}
		if_.If_ManSatUnbuild(handle)
	}
	if got.String() != string(want) {
		t.Fatal("XYZ decisions differ from original C")
	}
	t.Log("960 reused XYZ SAT/UNSAT queries across K2..6 match source and direct LUT evaluation")
}
