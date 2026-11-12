package gia

import (
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestFrameMiniAigAndLut(t *testing.T) {
	f := abcmain.NewFrame()
	InstallFrameBridge(f)
	mini := miniaig.Mini_AigStart()
	a := miniaig.Mini_AigCreatePi(mini)
	miniaig.Mini_AigCreatePo(mini, a^1)
	if err := abcmain.Abc_FrameGiaInputMiniAig(f, mini); err != nil {
		t.Fatal(err)
	}
	original := f.Resource("pGia").(*Gia_Man_t)
	if f.Resource("pGiaMiniAig") == nil || f.Resource("vCopyMiniAig") == nil {
		t.Fatal("missing MiniAIG snapshot")
	}
	if err := abcmain.Abc_FrameGiaInputMiniAig(f, "invalid"); err == nil || f.Resource("pGia") != original {
		t.Fatal("invalid import replaced graph")
	}
	out, err := abcmain.Abc_FrameGiaOutputMiniAig(f)
	if err != nil || !miniaig.Mini_AigCheck(out.(*miniaig.Mini_Aig_t)) {
		t.Fatalf("export: %v", err)
	}
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.fCoarsen = 0
	mapped := Mf_ManPerformMapping(original, &pars)
	f.SetResource("pGia", mapped)
	v, err := abcmain.Abc_FrameGiaOutputMiniLut(f)
	if err != nil {
		t.Fatal(err)
	}
	lut := v.(*miniaig.Mini_Lut_t)
	attrs, err := abcmain.Abc_FrameGiaOutputMiniLutAttr(f, lut)
	if err != nil || len(attrs) != miniaig.Mini_LutNodeNum(lut) {
		t.Fatalf("attributes %v", err)
	}
	if err := abcmain.Abc_FrameGiaInputMiniLut(f, lut); err != nil {
		t.Fatal(err)
	}
	result := f.Resource("pGia").(*Gia_Man_t)
	for _, x := range []bool{false, true} {
		got, err := result.Evaluate([]bool{x})
		if err != nil || len(got) != 1 || got[0] == x {
			t.Fatalf("input %v: %v %v", x, got, err)
		}
	}
}
