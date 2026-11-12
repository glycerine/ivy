package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestNfMatchesCutsToLibrary(t *testing.T) {
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
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	cuts := newNfCutManager(g, &pars, library)
	cuts.computeCuts()
	matcher := newNfMatcher(cuts)
	matcher.computeMapping()
	direct := matcher.objects[out>>1][0][0]
	inverted := matcher.objects[out>>1][1][0]
	if library.cells[direct.gate].Name() != "and2" || direct.cutOffset == 0 || direct.inverted {
		t.Fatalf("incorrect direct match: %+v", direct)
	}
	if inverted.delay >= nfInfinity || inverted.delay < direct.delay || inverted.flow < direct.flow {
		t.Fatalf("incorrect inverted match: %+v", inverted)
	}
	if direct.delay != library.cells[direct.gate].Delay(0) {
		t.Fatalf("unexpected delay: %+v", direct)
	}
}
