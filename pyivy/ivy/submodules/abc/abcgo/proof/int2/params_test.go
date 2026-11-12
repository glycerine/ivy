package int2

import "testing"

func TestSetDefaultParamsResetsAllFields(t *testing.T) {
	params := Int2_ManPars_t{
		nBTLimit: 12, nFramesS: 4, nFramesMax: 99, nSecLimit: 8,
		nFramesK: 5, fRewrite: 1, fTransLoop: 1, fDropInvar: 1,
		fVerbose: 1, iFrameMax: 70, pFileName: "stale",
	}
	Int2_ManSetDefaultParams(&params)
	want := Int2_ManPars_t{nFramesS: 1, nFramesK: 1, iFrameMax: -1}
	if params != want {
		t.Fatalf("defaults = %+v, want %+v", params, want)
	}
}

func TestSetDefaultParamsRejectsNil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil parameter block to panic")
		}
	}()
	Int2_ManSetDefaultParams(nil)
}
