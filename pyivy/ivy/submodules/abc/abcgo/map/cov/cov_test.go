package cov

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func makeAIGForCoverTest(t *testing.T, complementInput, complementOutput bool) *abc.Network {
	t.Helper()
	network := abc.NewNetwork("cov_test")
	network.Type = uint8(abc.ABC_NTK_STRASH)
	network.Function = uint8(abc.ABC_FUNC_AIG)
	a := network.NewObject(abc.ObjectPI)
	a.SetName("a")
	b := network.NewObject(abc.ObjectPI)
	b.SetName("b")
	aEdge, bEdge := a.Signal(), b.Signal()
	if complementInput {
		aEdge = aEdge.Not()
	}
	root := network.AndEdges(aEdge, bEdge)
	if complementOutput {
		root = root.Not()
	}
	po := network.NewObject(abc.ObjectPO)
	po.SetName("y")
	if err := network.AddFanin(po, root); err != nil {
		t.Fatal(err)
	}
	return network
}

func TestSopEsopCoverPreservesAIGFunction(t *testing.T) {
	for _, complementInput := range []bool{false, true} {
		for _, complementOutput := range []bool{false, true} {
			source := makeAIGForCoverTest(t, complementInput, complementOutput)
			got := Abc_NtkSopEsopCover(source, 2, 4, 0, 1, 0, 0)
			if got == nil {
				t.Fatalf("conversion failed (complement input=%v output=%v)", complementInput, complementOutput)
			}
			if err := got.Check(); err != nil {
				t.Fatalf("converted graph is invalid: %v", err)
			}
			oldInputs := source.ObjectsOfKind(abc.ObjectPI)
			newInputs := got.ObjectsOfKind(abc.ObjectPI)
			oldOutput := source.ObjectsOfKind(abc.ObjectPO)[0]
			newOutput := got.ObjectsOfKind(abc.ObjectPO)[0]
			for assignment := 0; assignment < 4; assignment++ {
				oldValues := map[*abc.Object]bool{oldInputs[0]: assignment&1 != 0, oldInputs[1]: assignment&2 != 0}
				newValues := map[*abc.Object]bool{newInputs[0]: assignment&1 != 0, newInputs[1]: assignment&2 != 0}
				want, err := source.Evaluate(oldOutput.Signal(), oldValues)
				if err != nil {
					t.Fatal(err)
				}
				gotValue, err := got.Evaluate(newOutput.Signal(), newValues)
				if err != nil {
					t.Fatal(err)
				}
				if gotValue != want {
					t.Fatalf("input %02b: got %v, want %v", assignment, gotValue, want)
				}
			}
		}
	}
}

func TestCoverRejectsImpossibleLimits(t *testing.T) {
	source := makeAIGForCoverTest(t, false, false)
	if got := Abc_NtkSopEsopCover(source, 1, 4, 0, 1, 0, 0); got != nil {
		t.Fatal("fanin limit below two unexpectedly succeeded")
	}
	if got := Abc_NtkSopEsopCover(source, 2, 0, 0, 1, 0, 0); got != nil {
		t.Fatal("zero cube limit unexpectedly succeeded")
	}
}

func TestCubeCoverSerializationAndManager(t *testing.T) {
	manager := Min_ManAlloc(2)
	defer Min_ManFree(manager)
	cube := Min_CubeAlloc(manager)
	Min_CubeXorBit(cube, 0) // a positive literal
	cube.nLits = 1
	if got := Min_CubeCountLits(cube); got != 1 {
		t.Fatalf("literal count = %d, want 1", got)
	}
	coverText := Vec_StrAlloc(0)
	Min_CoverCreate(coverText, cube, '1')
	if got := string(Vec_StrArray(coverText)); got != "1- 1\n\x00" {
		t.Fatalf("cover serialization %q", got)
	}
}

func makeCubePattern(manager *Min_Man_t, pattern string) *Min_Cube_t {
	cube := Min_CubeAlloc(manager)
	for variable, value := range pattern {
		switch value {
		case '1':
			Min_CubeXorBit(cube, int32(2*variable))
			cube.nLits++
		case '0':
			Min_CubeXorBit(cube, int32(2*variable+1))
			cube.nLits++
		case '-':
		default:
			panic("invalid cube pattern")
		}
	}
	return cube
}

func TestSOPMinimizationPreservesTruthAndReducesCover(t *testing.T) {
	manager := Min_ManAlloc(3)
	defer Min_ManFree(manager)
	for minterm := 4; minterm < 8; minterm++ {
		pattern := []byte("---")
		for variable := 0; variable < 3; variable++ {
			if minterm&(1<<variable) != 0 {
				pattern[variable] = '1'
			} else {
				pattern[variable] = '0'
			}
		}
		Min_SopAddCube(manager, makeCubePattern(manager, string(pattern)))
	}
	Min_SopMinimize(manager)
	cover := Min_CoverCollect(manager, manager.nVars)
	if got := Min_CoverCountCubes(cover); got != 1 {
		t.Fatalf("minimized SOP has %d cubes, want one", got)
	}
	for assignment := uint64(0); assignment < 8; assignment++ {
		want := assignment&4 != 0
		if got := cubeCoverMatches(cover, assignment, false); got != want {
			t.Fatalf("assignment %03b: got %v, want %v", assignment, got, want)
		}
	}
}

func TestESOPMinimizationKeepsXorSemantics(t *testing.T) {
	for _, variables := range []int32{3, 5} {
		manager := Min_ManAlloc(variables)
		Min_EsopAddCube(manager, makeCubePattern(manager, "1"+repeatDash(int(variables)-1)))
		pattern := "-1" + repeatDash(int(variables)-2)
		Min_EsopAddCube(manager, makeCubePattern(manager, pattern))
		Min_EsopMinimize(manager)
		cover := Min_CoverCollect(manager, manager.nVars)
		limit := uint64(1) << variables
		for assignment := uint64(0); assignment < limit; assignment++ {
			want := (assignment&1 != 0) != (assignment&2 != 0)
			if got := cubeCoverMatches(cover, assignment, true); got != want {
				t.Fatalf("%d vars assignment %0*b: got %v, want %v", variables, variables, assignment, got, want)
			}
		}
		Min_ManFree(manager)
	}
}

func repeatDash(count int) string {
	result := make([]byte, count)
	for i := range result {
		result[i] = '-'
	}
	return string(result)
}

func cubeCoverMatches(cover *Min_Cube_t, assignment uint64, xor bool) bool {
	value := false
	for cube := cover; cube != nil; cube = cube.pNext {
		if cubeMatches(cube, assignment) {
			if xor {
				value = !value
			} else {
				return true
			}
		}
	}
	return value
}
