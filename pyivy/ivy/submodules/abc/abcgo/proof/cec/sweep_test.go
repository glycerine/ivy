package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec/core"
	"testing"
)

func sweepFixture() (*Gia_Man_t, int32, int32) {
	p := gia.Gia_ManStart(32)
	x, y, z := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	xy := gia.Gia_ManAppendAnd(p, x, y)
	xyz := gia.Gia_ManAppendAnd(p, xy, z)
	yz := gia.Gia_ManAppendAnd(p, y, z)
	same := gia.Gia_ManAppendAnd(p, x, yz)
	gia.Gia_ManAppendCo(p, xyz)
	gia.Gia_ManAppendCo(p, same)
	return p, xyz >> 1, same >> 1
}
func TestNativeSATSuitedSweep(t *testing.T) {
	p, _, _ := sweepFixture()
	params := new(Cec_ParFra_t)
	Cec_ManFraSetDefaultParams(params)
	reduced := Cec_ManSatSweeping(p, params, 1)
	if params.Error != nil || reduced == nil {
		t.Fatalf("sweep failed: %v", params.Error)
	}
	if p.HasRepresentatives() {
		t.Fatal("source sweep left temporary equivalence classes on the input")
	}
	if gia.Gia_ManAndNum(reduced) >= gia.Gia_ManAndNum(p) {
		t.Fatal("sweep did not reduce redundant cones")
	}
	result, err := core.VerifyTwo(p, reduced, core.Options{})
	if err != nil || result.Status != core.Equivalent {
		t.Fatalf("reduction changed outputs: %+v %v", result, err)
	}
}
func TestNativeSweepDoesNotAssumeSequentialReachability(t *testing.T) {
	p := gia.Gia_ManStart(16)
	x := gia.Gia_ManAppendCi(p)
	reg := gia.Gia_ManAppendCi(p)
	both := gia.Gia_ManAppendAnd(p, x, reg)
	gia.Gia_ManAppendCo(p, both)
	gia.Gia_ManAppendCo(p, reg)
	gia.Gia_ManSetRegNum(p, 1)
	params := new(Cec_ParFra_t)
	Cec_ManFraSetDefaultParams(params)
	reduced, err := SatSweep(p, params)
	if err != nil || reduced.RegisterCount() != 1 {
		t.Fatalf("sequential interface: %v", err)
	}
	// RO=1 is legal for combinational sweeping, even if an initial trace never
	// reaches it. The graph must not be merged with constant zero.
	if gia.Gia_ObjHasRepr(p, both>>1) != 0 {
		t.Fatal("free register output treated as zero")
	}
	options := core.Options{TreatRegistersAsInputs: true}
	result, err := core.VerifyTwo(p, reduced, options)
	if err != nil || result.Status != core.Equivalent {
		t.Fatalf("sequential sweep changed function: %+v %v", result, err)
	}
}
func TestNativeSweepMiterCounterexample(t *testing.T) {
	p, _, _ := sweepFixture()
	params := new(Cec_ParFra_t)
	Cec_ManFraSetDefaultParams(params)
	params.FCheckMiter = 1
	result, err := SatSweep(p, params)
	if err != nil || result != nil || p.CombinationalCounterexample() == nil {
		t.Fatalf("miter mismatch: %v", err)
	}
}

func TestNativeSweepLevelBoundAndNoIterations(t *testing.T) {
	for _, noIterations := range []bool{false, true} {
		p, _, b := sweepFixture()
		params := new(Cec_ParFra_t)
		Cec_ManFraSetDefaultParams(params)
		if noIterations {
			params.NItersMax = 0
		} else {
			params.NLevelMax = 1
		}
		reduced, err := SatSweep(p, params)
		if err != nil || reduced == nil {
			t.Fatalf("bounded sweep: %v", err)
		}
		if gia.Gia_ObjHasRepr(p, b) != 0 {
			t.Fatal("excluded node was merged")
		}
	}
}

func TestNativeSweepComplementedClass(t *testing.T) {
	p := gia.Gia_ManStart(16)
	x, y := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	both := gia.Gia_ManAppendAnd(p, x, y)
	opposite := gia.Gia_ManAppendAnd(p, both^1, 1)
	gia.Gia_ManAppendCo(p, both)
	gia.Gia_ManAppendCo(p, opposite)
	params := new(Cec_ParFra_t)
	Cec_ManFraSetDefaultParams(params)
	reduced, err := SatSweep(p, params)
	if err != nil {
		t.Fatal(err)
	}
	if gia.Gia_ObjRepr(p, opposite>>1) != both>>1 {
		t.Fatal("complemented equivalence not identified")
	}
	result, err := core.VerifyTwo(p, reduced, core.Options{})
	if err != nil || result.Status != core.Equivalent {
		t.Fatalf("phase lost: %+v %v", result, err)
	}
}
