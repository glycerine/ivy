package ssw

import (
	"reflect"
	"testing"
)

func TestCorrespondenceDefaultParameters(t *testing.T) {
	p := Ssw_Pars_t{FDynamic: 1, nConflicts: 5, pData: "stale"}
	Ssw_ManSetDefaultParams(&p)
	want := Ssw_Pars_t{NFramesK: 1, NFramesAddSim: 2, NBTLimit: 1000, nBTLimitGlobal: 5000000, nMinDomSize: 100, NItersStop: -1, nResimDelta: 1000, nStepsMax: -1, NSatVarMax: 1000, nRecycleCalls: 50, NSatVarMax2: 5000, NRecycleCalls2: 250}
	if !reflect.DeepEqual(p, want) {
		t.Fatalf("incorrect defaults: %+v", p)
	}
	Ssw_ManSetDefaultParamsLcorr(&p)
	want.FLatchCorrOpt, want.NBTLimit = 1, 10000
	if !reflect.DeepEqual(p, want) {
		t.Fatalf("incorrect latch defaults: %+v", p)
	}
}

func TestRarityDefaultParameters(t *testing.T) {
	p := Ssw_RarPars_t{nSolved: 9, FVerbose: 1}
	Ssw_RarSetDefaultParams(&p)
	if want := (Ssw_RarPars_t{NFrames: 20, NWords: 50, NBinSize: 8}); !reflect.DeepEqual(p, want) {
		t.Fatalf("incorrect rarity defaults: %+v", p)
	}
}
