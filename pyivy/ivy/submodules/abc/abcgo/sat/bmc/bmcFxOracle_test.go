package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

// fxEncodeTruth uses one blocking implication per input assignment, independently
// of the gates and cover-generation routines used by the implementation.
func fxEncodeTruth(s *bsat.Solver, output int32, inputs []int32, truth uint64) {
	for row := 0; row < 1<<len(inputs); row++ {
		clause := make([]bsat.Literal, 0, len(inputs)+1)
		for bit, v := range inputs {
			if row&(1<<bit) != 0 {
				clause = append(clause, bsat.Negative(v))
			} else {
				clause = append(clause, bsat.Positive(v))
			}
		}
		if truth&(uint64(1)<<row) != 0 {
			clause = append(clause, bsat.Positive(output))
		} else {
			clause = append(clause, bsat.Negative(output))
		}
		if !s.AddClause(clause...) {
			panic("invalid truth table encoding")
		}
	}
}
func fxCheckCover(t *testing.T, cubes *vec.Vec_Wec_t, truth uint64, n int) {
	t.Helper()
	for row := 0; row < 1<<n; row++ {
		covered := false
		for i := 0; i < vec.Vec_WecSize(cubes); i++ {
			match := true
			for _, lit := range vec.Vec_IntArray(vec.Vec_WecEntry(cubes, i)) {
				v := int(lit >> 1)
				if v < 0 || v >= n {
					t.Fatalf("invalid cube literal %d", lit)
				}
				if (row&(1<<v) != 0) != (lit&1 == 0) {
					match = false
					break
				}
			}
			if match {
				covered = true
				break
			}
		}
		want := truth&(uint64(1)<<row) != 0
		if covered != want {
			t.Fatalf("row %03b: covered=%v want=%v cubes=%d", row, covered, want, vec.Vec_WecSize(cubes))
		}
	}
}
func TestFxExactOnsetCovers(t *testing.T) {
	// Inputs occupy variables 3..5; outputs are 1 and 2 as expected by Bmc_FxSolve.
	cases := []struct {
		name  string
		truth uint64
	}{{"false", 0}, {"true", 0xff}, {"and", 0x88}, {"xor", 0x66}, {"mux", 0xca}, {"or", 0xee}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := bsat.NewSolver()
			for i := 0; i < 8; i++ {
				s.NewVariable()
			}
			inputs := []int32{3, 4, 5}
			fxEncodeTruth(s, 1, inputs, tc.truth)
			vars := vec.Vec_IntAllocArrayCopy(inputs, len(inputs))
			cubes := vec.Vec_WecAlloc(8)
			var count int32
			if got := Bmc_FxSolve(s, 0, 6, vars, 0, 0, &count, cubes); got != 1 {
				t.Fatalf("status=%d", got)
			}
			if int(count) != vec.Vec_WecSize(cubes) {
				t.Fatalf("count=%d cubes=%d", count, vec.Vec_WecSize(cubes))
			}
			fxCheckCover(t, cubes, tc.truth, len(inputs))
		})
	}
}
func TestFxMultipleOutputsAndIndependentGuards(t *testing.T) {
	s := bsat.NewSolver()
	for i := 0; i < 8; i++ {
		s.NewVariable()
	}
	inputs := []int32{3, 4, 5}
	fxEncodeTruth(s, 1, inputs, 0x88)
	fxEncodeTruth(s, 2, inputs, 0x66)
	vars := vec.Vec_IntAllocArrayCopy(inputs, len(inputs))
	cubes := vec.Vec_WecAlloc(8)
	for i, truth := range []uint64{0x88, 0x66} {
		if got := Bmc_FxSolve(s, int32(i), int32(6+i), vars, 0, 0, nil, cubes); got != 1 {
			t.Fatalf("output %d status=%d", i, got)
		}
		fxCheckCover(t, cubes, truth, len(inputs))
	}
}

func TestFxExhaustiveThreeInputTruthTables(t *testing.T) {
	for truth := uint64(0); truth < 256; truth++ {
		s := bsat.NewSolver()
		for i := 0; i < 6; i++ {
			s.NewVariable()
		}
		inputs := []int32{2, 3, 4}
		fxEncodeTruth(s, 1, inputs, truth)
		vars := vec.Vec_IntAllocArrayCopy(inputs, len(inputs))
		cubes := vec.Vec_WecAlloc(8)
		if got := Bmc_FxSolve(s, 0, 5, vars, 0, 0, nil, cubes); got != 1 {
			t.Fatalf("truth=%02x status=%d", truth, got)
		}
		t.Run(fmt.Sprintf("%02x", truth), func(t *testing.T) { fxCheckCover(t, cubes, truth, len(inputs)) })
	}
}
