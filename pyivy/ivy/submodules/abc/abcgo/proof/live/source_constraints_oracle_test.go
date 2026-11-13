package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"slices"
	"testing"
)

// Compiled ABC test_vectors/constraint_source_oracle.c prints:
// zero oldpos=2 oldconstraints=1 newpos=3 newconstraints=0 target=2
// absorber newpos=3 newconstraints=0 target=2
// combination newpos=3 newconstraints=0; dup newpos=2 newconstraints=1.
func TestSourceZeroLiveCopiesEveryPOAndDropsConstraintCount(t *testing.T) {
	p := aig.Aig_ManStart(20)
	a, b := aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, a)
	aig.Aig_ObjCreateCo(p, a.Not())
	aig.Aig_ObjCreateCo(p, b)
	aig.Aig_ManSetRegNum(p, 1)
	aig.Aig_ManSetConstrNum(p, 1)
	index := int32(-1)
	result := createNewAigWith0LivePo(p, []aig.Edge{a}, &index)
	if trueOutputs(p) != 2 || aig.Aig_ManConstrNum(p) != 1 || trueOutputs(result) != 3 || aig.Aig_ManConstrNum(result) != 0 || index != 2 {
		t.Fatalf("old=%d/%d new=%d/%d index=%d", trueOutputs(p), aig.Aig_ManConstrNum(p), trueOutputs(result), aig.Aig_ManConstrNum(result), index)
	}
	zero, absorberIndex := int32(0), int32(-1)
	absorber := introduceAbsorberLogic(p, &zero, &absorberIndex, 0)
	if trueOutputs(absorber) != 3 || aig.Aig_ManConstrNum(absorber) != 0 || absorberIndex != 2 {
		t.Fatalf("ABC absorber interface %d/%d target %d", trueOutputs(absorber), aig.Aig_ManConstrNum(absorber), absorberIndex)
	}
	combination := generateGeneralDisjunctiveTester(nil, p, 1)
	if trueOutputs(combination) != 3 || aig.Aig_ManConstrNum(combination) != 0 {
		t.Fatalf("ABC combination interface %d/%d", trueOutputs(combination), aig.Aig_ManConstrNum(combination))
	}
	duplicate := kLiveProofModelMode(p, 0, 0, -1, true)
	if trueOutputs(duplicate) != 2 || aig.Aig_ManConstrNum(duplicate) != 1 {
		t.Fatalf("ABC duplicate interface %d/%d", trueOutputs(duplicate), aig.Aig_ManConstrNum(duplicate))
	}
	for _, input := range []bool{false, true} {
		out, err := result.Evaluate([]bool{input, false})
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != input || out[1] == input || out[2] != input {
			t.Fatalf("input %v: copied PO order %v", input, out)
		}
	}
}

func TestSourceBuildersIncludeTrailingConstraintPO(t *testing.T) {
	n := constraintMonitorFixture()
	p := n.Graph
	indexes := &aigPoIndices{attrPendingSignalIndex: 0, attrSafetyInvarIndex: 3}
	var start int32
	mono := createMonotoneTester(p, indexes, &monotoneVectorsStruct{attrHintMonotone: []int32{1}}, &start)
	if start != 5 || trueOutputs(mono) != 6 || aig.Aig_ManConstrNum(mono) != 0 {
		t.Fatalf("monotone start=%d POs=%d constraints=%d", start, trueOutputs(mono), aig.Aig_ManConstrNum(mono))
	}
	disj := createDisjunctiveMonotoneTester(p, indexes, &antecedentConsequentVectorsStruct{attrConsequentCandidates: []int32{1}}, &start)
	if start != 5 || trueOutputs(disj) != 6 || aig.Aig_ManConstrNum(disj) != 0 {
		t.Fatalf("disj start=%d POs=%d constraints=%d", start, trueOutputs(disj), aig.Aig_ManConstrNum(disj))
	}
	comb := disjunctiveTester(p, []aig.Edge{p.Outputs()[1]}, 1)
	if trueOutputs(comb) != 6 || aig.Aig_ManConstrNum(comb) != 0 {
		t.Fatalf("comb POs=%d constraints=%d", trueOutputs(comb), aig.Aig_ManConstrNum(comb))
	}
	zero, target := int32(2), int32(-1)
	absorber := introduceAbsorberLogic(p, &zero, &target, 0)
	if target != 5 || trueOutputs(absorber) != 6 || aig.Aig_ManConstrNum(absorber) != 0 {
		t.Fatalf("absorber target=%d POs=%d constraints=%d", target, trueOutputs(absorber), aig.Aig_ManConstrNum(absorber))
	}
	arena := createNewAigWith0LivePoWithDSC(p, []aig.Edge{outputDriver(p, 2)}, &start, 0, 0, 0, [][]int32{{1}})
	if start != 5 || trueOutputs(arena) != 6 || aig.Aig_ManConstrNum(arena) != 0 {
		t.Fatalf("arena start=%d POs=%d constraints=%d", start, trueOutputs(arena), aig.Aig_ManConstrNum(arena))
	}
	proof := kLiveProofModelMode(p, 0, 2, -1, true)
	if trueOutputs(proof) != 5 || aig.Aig_ManConstrNum(proof) != 1 {
		t.Fatalf("source Aig_ManDupSimple POs=%d constraints=%d", trueOutputs(proof), aig.Aig_ManConstrNum(proof))
	}
}

func TestSourceSelectorsInspectTrailingConstraintPO(t *testing.T) {
	n := constraintMonitorFixture()
	if got := findHintOutputs(n); len(got) != 2 || got[0] != 1 || got[1] != 4 {
		t.Fatalf("source hints %v", got)
	}
	if got := findHintOutputsMode(n, false); len(got) != 1 || got[0] != 1 {
		t.Fatalf("native hints %v", got)
	}
}

func TestSourceAndNativeZeroLiveAgreeWithoutConstraints(t *testing.T) {
	p := aig.Aig_ManStart(10)
	a, b := aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, a)
	aig.Aig_ObjCreateCo(p, a.Not())
	aig.Aig_ObjCreateCo(p, b)
	aig.Aig_ManSetRegNum(p, 1)
	var x, y int32
	source := createNewAigWith0LivePoMode(p, []aig.Edge{a}, &x, true)
	native := createNewAigWith0LivePoMode(p, []aig.Edge{a}, &y, false)
	if x != 2 || y != 2 || trueOutputs(source) != 3 || trueOutputs(native) != 3 {
		t.Fatalf("source/native indexes %d/%d POs %d/%d", x, y, trueOutputs(source), trueOutputs(native))
	}
	for _, bit := range []bool{false, true} {
		aout, err := source.Evaluate([]bool{bit, false})
		if err != nil {
			t.Fatal(err)
		}
		bout, err := native.Evaluate([]bool{bit, false})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(aout, bout) {
			t.Fatalf("input %v source %v native %v", bit, aout, bout)
		}
	}
}
