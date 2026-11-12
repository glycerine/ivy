package cov

import (
	"strings"
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
	if got := Abc_NtkSopEsopCover(source, -1, 4, 0, 1, 0, 0); got != nil {
		t.Fatal("negative fanin limit unexpectedly succeeded")
	}
	if got := Abc_NtkSopEsopCover(source, 1, 4, 0, 1, 0, 0); got != nil {
		t.Fatal("fanin limit below two unexpectedly succeeded")
	}
	if got := Abc_NtkSopEsopCover(source, 2, 0, 0, 1, 0, 0); got != nil {
		t.Fatal("zero cube limit unexpectedly succeeded")
	}
}

func TestCoverGrowsCutBoundariesToHonorFaninLimit(t *testing.T) {
	network := abc.NewNetwork("cut_growth")
	network.Type = uint8(abc.ABC_NTK_STRASH)
	network.Function = uint8(abc.ABC_FUNC_AIG)
	inputs := make([]*abc.Object, 4)
	for i := range inputs {
		inputs[i] = network.NewObject(abc.ObjectPI)
	}
	left := network.AndEdges(inputs[0].Signal(), inputs[1].Signal())
	right := network.AndEdges(inputs[2].Signal(), inputs[3].Signal())
	root := network.AndEdges(left, right)
	output := network.NewObject(abc.ObjectPO)
	if err := network.AddFanin(output, root); err != nil {
		t.Fatal(err)
	}
	got := Abc_NtkSopEsopCover(network, 2, 4, 0, 1, 0, 0)
	if got == nil {
		t.Fatal("cover failed for a four-input cone that can be cut into two-input pieces")
	}
	for assignment := 0; assignment < 16; assignment++ {
		oldValues, newValues := map[*abc.Object]bool{}, map[*abc.Object]bool{}
		for i := range inputs {
			value := assignment&(1<<i) != 0
			oldValues[inputs[i]] = value
			newValues[got.ObjectsOfKind(abc.ObjectPI)[i]] = value
		}
		want, err := network.Evaluate(output.Signal(), oldValues)
		if err != nil {
			t.Fatal(err)
		}
		gotValue, err := got.Evaluate(got.ObjectsOfKind(abc.ObjectPO)[0].Signal(), newValues)
		if err != nil {
			t.Fatal(err)
		}
		if gotValue != want {
			t.Fatalf("assignment %04b: got %v, want %v", assignment, gotValue, want)
		}
	}
}

func TestCubeBoundIsIndependentOfFaninBound(t *testing.T) {
	network := abc.NewNetwork("three_input_parity")
	network.Type = uint8(abc.ABC_NTK_STRASH)
	network.Function = uint8(abc.ABC_FUNC_AIG)
	a, b, c := network.NewObject(abc.ObjectPI), network.NewObject(abc.ObjectPI), network.NewObject(abc.ObjectPI)
	xor := func(left, right abc.Edge) abc.Edge {
		one := network.AndEdges(left.Not(), right)
		two := network.AndEdges(left, right.Not())
		return network.AndEdges(one.Not(), two.Not()).Not()
	}
	root := xor(xor(a.Signal(), b.Signal()), c.Signal())
	output := network.NewObject(abc.ObjectPO)
	if err := network.AddFanin(output, root); err != nil {
		t.Fatal(err)
	}
	got := Abc_NtkSopEsopCover(network, 3, 4, 0, 1, 0, 0)
	if got == nil {
		t.Fatal("SOP cover with four cubes was rejected by the three-fanin limit")
	}
	for assignment := 0; assignment < 8; assignment++ {
		oldValues, newValues := map[*abc.Object]bool{}, map[*abc.Object]bool{}
		for i, input := range []*abc.Object{a, b, c} {
			value := assignment&(1<<i) != 0
			oldValues[input] = value
			newValues[got.ObjectsOfKind(abc.ObjectPI)[i]] = value
		}
		want, err := network.Evaluate(output.Signal(), oldValues)
		if err != nil {
			t.Fatal(err)
		}
		gotValue, err := got.Evaluate(got.ObjectsOfKind(abc.ObjectPO)[0].Signal(), newValues)
		if err != nil {
			t.Fatal(err)
		}
		if gotValue != want {
			t.Fatalf("assignment %03b: got %v, want %v", assignment, gotValue, want)
		}
	}
}

func TestSOPDerivationCanUseNegativePolarityCover(t *testing.T) {
	network := abc.NewNetwork("negative_sop")
	network.Type = uint8(abc.ABC_NTK_STRASH)
	network.Function = uint8(abc.ABC_FUNC_AIG)
	a, b, c, d := network.NewObject(abc.ObjectPI), network.NewObject(abc.ObjectPI), network.NewObject(abc.ObjectPI), network.NewObject(abc.ObjectPI)
	ab := network.AndEdges(a.Signal(), b.Signal())
	cd := network.AndEdges(c.Signal(), d.Signal())
	root := network.AndEdges(ab.Not(), cd.Not())
	output := network.NewObject(abc.ObjectPO)
	if err := network.AddFanin(output, root); err != nil {
		t.Fatal(err)
	}
	got := Abc_NtkSopEsopCover(network, 4, 4, 0, 1, 0, 0)
	if got == nil {
		t.Fatal("negative-polarity SOP conversion failed")
	}
	negativePhase := false
	for _, object := range got.ObjectsOfKind(abc.ObjectNode) {
		if data, ok := object.Data.(string); ok && strings.Contains(data, " 0\n") {
			negativePhase = true
		}
	}
	if !negativePhase {
		t.Fatal("derivation did not use the smaller negative-polarity cover")
	}
	for assignment := 0; assignment < 16; assignment++ {
		oldValues, newValues := map[*abc.Object]bool{}, map[*abc.Object]bool{}
		for i, input := range []*abc.Object{a, b, c, d} {
			value := assignment&(1<<i) != 0
			oldValues[input] = value
			newValues[got.ObjectsOfKind(abc.ObjectPI)[i]] = value
		}
		want, err := network.Evaluate(output.Signal(), oldValues)
		if err != nil {
			t.Fatal(err)
		}
		gotValue, err := got.Evaluate(got.ObjectsOfKind(abc.ObjectPO)[0].Signal(), newValues)
		if err != nil {
			t.Fatal(err)
		}
		if gotValue != want {
			t.Fatalf("assignment %04b: got %v, want %v", assignment, gotValue, want)
		}
	}
}

func TestCoverHonorsESOPMode(t *testing.T) {
	network := abc.NewNetwork("three_input_parity")
	network.Type = uint8(abc.ABC_NTK_STRASH)
	network.Function = uint8(abc.ABC_FUNC_AIG)
	a, b, c := network.NewObject(abc.ObjectPI), network.NewObject(abc.ObjectPI), network.NewObject(abc.ObjectPI)
	xor := func(left, right abc.Edge) abc.Edge {
		one := network.AndEdges(left.Not(), right)
		two := network.AndEdges(left, right.Not())
		return network.AndEdges(one.Not(), two.Not()).Not()
	}
	root := xor(xor(a.Signal(), b.Signal()), c.Signal())
	output := network.NewObject(abc.ObjectPO)
	if err := network.AddFanin(output, root); err != nil {
		t.Fatal(err)
	}
	got := Abc_NtkSopEsopCover(network, 3, 4, 1, 1, 0, 0)
	if got == nil {
		t.Fatal("ESOP conversion failed")
	}
	foundXor := false
	for _, object := range got.ObjectsOfKind(abc.ObjectNode) {
		if data, ok := object.Data.(string); ok && strings.Contains(data, " x\n") {
			foundXor = true
		}
	}
	if !foundXor {
		t.Fatal("ESOP mode did not select the smaller XOR-special cover")
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := map[*abc.Object]bool{a: assignment&1 != 0, b: assignment&2 != 0, c: assignment&4 != 0}
		newInputs := got.ObjectsOfKind(abc.ObjectPI)
		newValues := map[*abc.Object]bool{newInputs[0]: values[a], newInputs[1]: values[b], newInputs[2]: values[c]}
		want, err := network.Evaluate(output.Signal(), values)
		if err != nil {
			t.Fatal(err)
		}
		gotValue, err := got.Evaluate(got.ObjectsOfKind(abc.ObjectPO)[0].Signal(), newValues)
		if err != nil {
			t.Fatal(err)
		}
		if gotValue != want {
			t.Fatalf("assignment %03b: got %v, want %v", assignment, gotValue, want)
		}
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

func TestManagerDropsCompletedNodeState(t *testing.T) {
	source := makeAIGForCoverTest(t, false, false)
	manager := Cov_ManAlloc(source, 2, 4)
	defer Cov_ManFree(manager)
	root := source.ObjectsOfKind(abc.ObjectNode)[0]
	Abc_ObjSetSupp(root, Vec_IntAlloc(0))
	manager.nSupps = 1
	Abc_NodeCovDropData(manager, root)
	if manager.nSupps != 0 || Abc_ObjGetSupp(root) != nil || Abc_ObjGetCover2(root) != nil {
		t.Fatal("last fanout did not release the node's cover state")
	}
}

func TestCoverSupportCountsCommonFixedVariables(t *testing.T) {
	manager := Min_ManAlloc(2)
	defer Min_ManFree(manager)
	first := makeCubePattern(manager, "-1")
	second := makeCubePattern(manager, "1-")
	first.pNext = second
	if got := Min_CoverSuppVarNum(manager, first); got != 2 {
		t.Fatalf("cover support = %d, want 2", got)
	}
}

func TestSOPAddCubeMergesDistanceOneTerms(t *testing.T) {
	manager := Min_ManAlloc(2)
	defer Min_ManFree(manager)
	Min_SopAddCube(manager, makeCubePattern(manager, "1-"))
	Min_SopAddCube(manager, makeCubePattern(manager, "0-"))
	cover := Min_CoverCollect(manager, manager.nVars)
	if got := Min_CoverCountCubes(cover); got != 1 {
		t.Fatalf("merged SOP has %d cubes, want 1", got)
	}
	if got := Min_CoverSuppVarNum(manager, cover); got != 0 {
		t.Fatalf("merged tautology support = %d, want 0", got)
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
