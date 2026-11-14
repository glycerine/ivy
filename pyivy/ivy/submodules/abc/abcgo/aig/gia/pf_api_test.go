package gia

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestPfPublicMapperReturnsDuplicate(t *testing.T) {
	old := abcmain.Abc_FrameReadLibGen()
	defer abcmain.Abc_FrameSetLibGen(old)
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 1 0
.end
`
	lib, err := mio.ReadLibrary("pf", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	abcmain.Abc_FrameSetLibGen(lib)
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	out := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, out)
	var pars Jf_Par_t
	Pf_ManSetDefaultPars(&pars)
	duplicate := Pf_ManPerformMapping(g, &pars)
	if duplicate == nil || duplicate == g || duplicate.nObjs != g.nObjs || duplicate.VCellMapping != nil {
		t.Fatal("Pf result does not match C fallback")
	}
}
