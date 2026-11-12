package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestNfDeriveCellMapping(t *testing.T) {
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 2 0 2 0
.end
`
	lib, err := mio.ReadLibrary("nf", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	library := nfPrepareLibrary(mio.Mio_CollectRootsNew2(lib, 6, nil, 0), false, true, false)
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	out := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, out)
	Gia_ManAppendCo(g, out^1)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	cuts := newNfCutManager(g, &pars, library)
	cuts.computeCuts()
	matcher := newNfMatcher(cuts)
	matcher.computeMapping()
	refs := matcher.mappingRefs(0)
	matcher.deriveMapping(refs)
	if Gia_ManHasCellMapping(g) == 0 || Gia_ObjCellSize(g, out) != 2 || Gia_ObjIsCellInv(g, out^1) == 0 {
		t.Fatal("incorrect cell mapping")
	}
	gate := Gia_ObjCellId(g, out)
	if library.cells[gate].Name() != "and2" {
		t.Fatalf("wrong gate: %d", gate)
	}
	fanins := Gia_ObjCellFanins(g, out)
	if !((fanins[0] == a && fanins[1] == b) || (fanins[0] == b && fanins[1] == a)) {
		t.Fatalf("wrong fanins: %v", fanins)
	}
}
