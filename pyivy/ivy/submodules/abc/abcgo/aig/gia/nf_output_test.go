package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestNfFixOutputDrivers(t *testing.T) {
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 1 0
GATE nand2 2 O=!(a*b); PIN * INV 1 999 1 0 1 0
.end
`
	lib, err := mio.ReadLibrary("nf", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	library := nfPrepareLibrary(mio.Mio_CollectRootsNew2(lib, 6, nil, 0), false, true, false)
	for _, relax := range []bool{false, true} {
		g := Gia_ManStart(10)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		out := Gia_ManAppendAnd(g, a, b)
		Gia_ManAppendCo(g, out)
		Gia_ManAppendCo(g, out^1)
		var pars Jf_Par_t
		nfDefaultParameters(&pars)
		cuts := newNfCutManager(g, &pars, library)
		cuts.computeCuts()
		matcher := newNfMatcher(cuts)
		matcher.computeMapping()
		refs := matcher.mappingRefs(0)
		if refs.gates != 2 || refs.inverters != 0 {
			t.Fatalf("expected two gates: %+v", refs)
		}
		if relax {
			refs.delay += matcher.invDelay
		}
		changed := matcher.fixOutputDrivers(refs)
		if !relax && changed != 0 || relax && (changed != 1 || refs.inverters != 1 || refs.gates != 2) {
			t.Fatalf("relax %t changed %d refs %+v", relax, changed, refs)
		}
		matcher.deriveMapping(refs)
		if relax && Gia_ObjIsCellInv(g, out) == 0 {
			t.Fatal("missing output inverter")
		}
	}
}
