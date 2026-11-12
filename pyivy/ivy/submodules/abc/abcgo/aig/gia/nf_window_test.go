package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestNfExtractWindow(t *testing.T) {
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
	library := nfPrepareLibrary(mio.Mio_CollectRootsNew2(lib, 6, nil, 0), false, true, false)
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	out := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, out)
	Gia_ManAppendCo(g, out^1)
	var pars Jf_Par_t
	nfDefaultParameters(&pars)
	cm := newNfCutManager(g, &pars, library)
	cm.computeCuts()
	matcher := newNfMatcher(cm)
	matcher.computeMapping()
	refs := matcher.mappingRefs(0)
	var roots, solution, gates Vec_Int_t
	var cuts, objCuts Vec_Wec_t
	var areas Vec_Wrd_t
	var inv word
	if got := matcher.nfExtractWindow(refs, &roots, &cuts, &objCuts, &solution, &gates, &areas, &inv, 10, 20); got != 2 {
		t.Fatalf("inputs %d", got)
	}
	if vec.Vec_IntSize(&roots) != 2 || vec.Vec_WecSize(&objCuts) != 2 || vec.Vec_WecSize(&cuts) != vec.Vec_IntSize(&gates) || vec.Vec_WecSize(&cuts) != vec.Vec_WrdSize(&areas) {
		t.Fatal("incorrect window dimensions")
	}
	if vec.Vec_IntSize(&solution) != 2 || inv != library.cells[3].AreaW {
		t.Fatal("missing mapped solution")
	}
	for _, index := range vec.Vec_IntArray(&solution) {
		if index < 0 || int(index) >= vec.Vec_WecSize(&cuts) {
			t.Fatalf("invalid solution index %d", index)
		}
	}
}
