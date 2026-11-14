package gia

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestMiniAigFileAndFrameEquivalence(t *testing.T) {
	mini := miniaig.Mini_AigStart()
	a := miniaig.Mini_AigCreatePi(mini)
	miniaig.Mini_AigCreatePo(mini, a^1)
	g := Gia_ManFromMiniAig(mini, nil)
	file := filepath.Join(t.TempDir(), "sample.mini")
	Gia_ManWriteMiniAig(g, file)
	read := Gia_ManReadMiniAig(file)
	if read == nil || Gia_ManReadMiniAig(file+".missing") != nil {
		t.Fatal("incorrect MiniAIG file handling")
	}
	out, err := read.Evaluate([]bool{false})
	if err != nil || len(out) != 1 || !out[0] {
		t.Fatal(out, err)
	}
	f := abcmain.NewFrame()
	InstallFrameBridge(f)
	if err := abcmain.Abc_FrameGiaInputMiniAig(f, mini); err != nil {
		t.Fatal(err)
	}
	classes := Gia_ManDup(f.Resource("pGiaMiniAig").(*Gia_Man_t))
	classes.PReprs = make([]Gia_Rpr_t, classes.nObjs)
	for i := range classes.PReprs {
		classes.PReprs[i].iRepr = GIA_VOID
	}
	f.SetResource("pGia2", classes)
	values, err := abcmain.Abc_FrameReadMiniAigEquivClasses(f)
	if err != nil || len(values) != miniaig.Mini_AigNodeNum(mini) {
		t.Fatalf("classes %v: %v", values, err)
	}
	for _, v := range values {
		if v != -1 {
			t.Fatalf("unexpected equivalence %v", values)
		}
	}
}
