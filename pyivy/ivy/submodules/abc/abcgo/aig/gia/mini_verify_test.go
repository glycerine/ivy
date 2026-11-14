package gia

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
	abcmain "github.com/glycerine/abcgo/base/main"
)

type miniTestDispatcher struct{}

func (miniTestDispatcher) Execute(f *abcmain.Frame, line string) int32 {
	if line != "&ps; &scorr; &ps" {
		return 1
	}
	g := Gia_ManDup(f.Resource("pGiaMiniAig").(*Gia_Man_t))
	g.PReprs = make([]Gia_Rpr_t, g.nObjs)
	for i := range g.PReprs {
		g.PReprs[i].iRepr = GIA_VOID
	}
	f.SetResource("pGia2", g)
	return 0
}
func TestMiniAigVerifyWorkflow(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.mini")
	miter := filepath.Join(dir, "miter.aig")
	reduced := filepath.Join(dir, "reduced.aig")
	p := miniaig.Mini_AigStart()
	a, b := miniaig.Mini_AigCreatePi(p), miniaig.Mini_AigCreatePi(p)
	miniaig.Mini_AigAnd(p, a, b)
	x := miniaig.Mini_AigAnd(p, a, b)
	miniaig.Mini_AigCreatePo(p, x)
	miniaig.Mini_AigDump(p, input)
	frame := abcmain.NewFrame()
	InstallFrameBridge(frame)
	frame.Dispatcher = miniTestDispatcher{}
	if err := Gia_MiniAigVerifyToFiles(frame, input, miter, reduced); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{miter, reduced} {
		if info, err := os.Stat(name); err != nil || info.Size() == 0 {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}
