package mpm

import "testing"

func TestSimpleLUTLibraryMatchesSource(t *testing.T) {
	for n := int32(0); n <= MPM_VAR_MAX; n++ {
		lib := Mpm_LibLutSetSimple(n)
		if lib.LutMax != n || lib.fVarPinDelays != 0 {
			t.Fatalf("size %d: %+v", n, lib)
		}
		for i := 0; i <= MPM_VAR_MAX; i++ {
			wantArea := int32(0)
			if i > 0 && int32(i) <= n {
				wantArea = MPM_UNIT_AREA
			}
			if lib.pLutAreas[i] != wantArea {
				t.Fatalf("size %d area[%d]", n, i)
			}
			for j := 0; j <= MPM_VAR_MAX; j++ {
				want := int32(0)
				if i > 0 && int32(i) <= n && j < i {
					want = MPM_UNIT_TIME
				}
				if lib.pLutDelays[i][j] != want {
					t.Fatalf("size %d delay[%d][%d]", n, i, j)
				}
			}
		}
		Mpm_LibLutFree(lib)
	}
	Mpm_LibLutFree(nil)
}
func TestMappingDefaultsResetAllFields(t *testing.T) {
	p := Mpm_Par_t{pLib: Mpm_LibLutSetSimple(4), nNumCuts: 17, DelayTarget: 42, fUseTruth: 1, fVerbose: 1}
	Mpm_ManSetParsDefault(&p)
	if p != (Mpm_Par_t{nNumCuts: 8, DelayTarget: -1}) {
		t.Fatalf("defaults %+v", p)
	}
}
