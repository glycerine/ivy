package ssw

import (
	"reflect"
	"testing"
)

func TestCorrespondenceDefaultParameters(t *testing.T) {
	p := Ssw_Pars_t{fDynamic: 1, nConflicts: 5, pData: "stale"}
	Ssw_ManSetDefaultParams(&p)
	want := Ssw_Pars_t{nFramesK: 1, nFramesAddSim: 2, nBTLimit: 1000, nBTLimitGlobal: 5000000, nMinDomSize: 100, nItersStop: -1, nResimDelta: 1000, nStepsMax: -1, nSatVarMax: 1000, nRecycleCalls: 50, nSatVarMax2: 5000, nRecycleCalls2: 250}
	if !reflect.DeepEqual(p, want) {
		t.Fatalf("incorrect defaults: %+v", p)
	}
	Ssw_ManSetDefaultParamsLcorr(&p)
	want.fLatchCorrOpt, want.nBTLimit = 1, 10000
	if !reflect.DeepEqual(p, want) {
		t.Fatalf("incorrect latch defaults: %+v", p)
	}
}

func TestRarityDefaultParameters(t *testing.T) {
	p := Ssw_RarPars_t{nSolved: 9, fVerbose: 1}
	Ssw_RarSetDefaultParams(&p)
	if want := (Ssw_RarPars_t{nFrames: 20, nWords: 50, nBinSize: 8}); !reflect.DeepEqual(p, want) {
		t.Fatalf("incorrect rarity defaults: %+v", p)
	}
}
