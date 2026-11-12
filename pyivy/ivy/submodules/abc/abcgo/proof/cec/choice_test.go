package cec

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec/core"
	"github.com/glycerine/abcgo/proof/dch"
	"testing"
)

func TestNativeChoiceSnapshotsRetainAlternatives(t *testing.T) {
	p, _, _ := sweepFixture()
	// Snapshot outputs are interleaved: two associative variants and a repeat.
	gia.Gia_ManAppendCo(p, p.OutputLiterals()[0])
	params := dch.DefaultParams()
	result, err := ChoiceSnapshots(p, 3, &params)
	if err != nil {
		t.Fatal(err)
	}
	if result.OutputCount() != 1 || result.InputCount() != 3 {
		t.Fatal("choice interface changed")
	}
	if aig.Aig_ManChoiceNum(result) == 0 {
		t.Fatal("no actual choice alternatives generated")
	}
	before, err := core.NewModel(p)
	if err != nil {
		t.Fatal(err)
	}
	after, err := core.NewModel(gia.Gia_ManFromAig(result))
	if err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		a, _ := before.Evaluate(inputs)
		b, _ := after.Evaluate(inputs)
		if a[0] != b[0] {
			t.Fatalf("choice changed function at %d", mask)
		}
	}
}
func TestNativeChoiceSnapshotValidation(t *testing.T) {
	p, _, _ := sweepFixture()
	if _, err := ChoiceSnapshots(p, 3, nil); err == nil {
		t.Fatal("accepted nondivisible output layout")
	}
	if _, err := ChoiceSnapshots(p, 0, nil); err == nil {
		t.Fatal("accepted zero snapshots")
	}
}
func TestPublicCECChoiceDrivers(t *testing.T) {
	p, _, _ := sweepFixture()
	gia.Gia_ManAppendCo(p, p.OutputLiterals()[0])
	params := new(Cec_ParChc_t)
	Cec_ManChcSetDefaultParams(params)
	result := Cec_ManChoiceComputationVec(p, 3, params)
	if result == nil || params.Error != nil {
		t.Fatalf("choice driver: %v", params.Error)
	}
	if len(result.OutputLiterals()) != 1 {
		t.Fatal("source choice driver output stride")
	}
}

func TestPublicSingleCircuitChoiceSynthesis(t *testing.T) {
	p, _, _ := sweepFixture()
	params := new(Cec_ParChc_t)
	Cec_ManChcSetDefaultParams(params)
	result := Cec_ManChoiceComputation(p, params)
	if result == nil || params.Error != nil {
		t.Fatalf("choice synthesis failed: %v", params.Error)
	}
	proof, err := core.VerifyTwo(p, result, core.Options{})
	if err != nil || proof.Status != core.Equivalent {
		t.Fatalf("choice synthesis changed outputs %+v %v", proof, err)
	}
}
