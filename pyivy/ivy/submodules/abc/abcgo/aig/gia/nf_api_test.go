package gia

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestNfPublicMapper(t *testing.T) {
	old := abcmain.Abc_FrameReadLibGen()
	defer abcmain.Abc_FrameSetLibGen(old)
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 1 0
.end
`
	lib, err := mio.ReadLibrary("nf", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	abcmain.Abc_FrameSetLibGen(lib)
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	out := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, out)
	var pars Jf_Par_t
	Nf_ManSetDefaultPars(&pars)
	if Nf_ManPerformMapping(g, &pars) != g || Gia_ObjCellSize(g, out) != 2 {
		t.Fatal("public mapping failed")
	}
	abcmain.Abc_FrameSetLibGen(nil)
	if Nf_ManPerformMapping(g, &pars) != nil {
		t.Fatal("missing library accepted")
	}
}
