package if_

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/sat/bsat"
)

func TestIfSatBuildsOriginalCNF(t *testing.T) {
	for _, tc := range []struct {
		name    string
		build   func(int32) any
		vars    int32
		clauses uint32
	}{{"XY", If_ManSatBuildXY, 16, 48}, {"XYZ", If_ManSatBuildXYZ, 28, 128}} {
		t.Run(tc.name, func(t *testing.T) {
			handle := tc.build(2)
			defer If_ManSatUnbuild(handle)
			solver, ok := handle.(*bsat.Solver)
			if !ok {
				t.Fatalf("builder returned %T instead of the original SAT instance", handle)
			}
			if solver.NumVars() != tc.vars || solver.Statistics().Clauses != tc.clauses {
				t.Fatalf("original CNF shape changed: vars%d clauses%d", solver.NumVars(), solver.Statistics().Clauses)
			}
		})
	}
}
func TestIfSatCheckXYRetainsSourceAssumptions(t *testing.T) {
	handle := If_ManSatBuildXY(2)
	defer If_ManSatUnbuild(handle)
	scratch := Vec_IntAlloc(8)
	var bound, free word
	if If_ManSatCheckXY(handle, 2, []word{0x8080808080808080}, 3, 5, &bound, &free, scratch) != 1 {
		t.Fatal("nested AND should be satisfiable")
	}
	want := []int32{17, 19, 21, 23, 25, 27, 29, 30}
	if got := Vec_IntArray(scratch); !slices.Equal(got, want) {
		t.Fatalf("retained assumptions%v want%v", got, want)
	}
}
func TestIfSatCheckXYSourceDegeneratePartitions(t *testing.T) {
	handle := If_ManSatBuildXY(2)
	defer If_ManSatUnbuild(handle)
	for _, tc := range []struct {
		variables int32
		set       uint32
		truth     word
	}{{0, 0, 0}, {1, 0, 0xaaaaaaaaaaaaaaaa}, {1, 1, 0xaaaaaaaaaaaaaaaa}, {1, 3, 0xaaaaaaaaaaaaaaaa}} {
		t.Run(fmt.Sprintf("vars%d/set%x", tc.variables, tc.set), func(t *testing.T) {
			var bound, free word
			if If_ManSatCheckXY(handle, 2, []word{tc.truth}, tc.variables, tc.set, &bound, &free, Vec_IntAlloc(8)) != 1 {
				t.Fatal("source-valid partition was rejected")
			}
		})
	}
}
func TestIfSatCheckXYOnlyWritesOutputPair(t *testing.T) {
	handle := If_ManSatBuildXY(2)
	defer If_ManSatUnbuild(handle)
	bound := word(0x123456789abcdef0)
	if If_ManSatCheckXY(handle, 2, []word{0x8080808080808080}, 3, 5, &bound, nil, Vec_IntAlloc(8)) != 1 {
		t.Fatal("expected SAT")
	}
	if bound != 0x123456789abcdef0 {
		t.Fatalf("single output pointer was modified: %x", bound)
	}
}

func TestIfSatCheckXYUnsatPreservesOutputs(t *testing.T) {
	handle := If_ManSatBuildXY(2)
	defer If_ManSatUnbuild(handle)
	bound, free := word(0x1234), word(0x5678)
	if If_ManSatCheckXY(handle, 2, []word{0xacacacacacacacac}, 3, 5, &bound, &free, Vec_IntAlloc(8)) != 0 {
		t.Fatal("three distinct free cofactors should be UNSAT")
	}
	if bound != 0x1234 || free != 0x5678 {
		t.Fatalf("UNSAT overwrote caller outputs: %x %x", bound, free)
	}
}
