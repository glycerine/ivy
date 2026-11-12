package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestNfMappingReferencesBothPhases(t *testing.T) {
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
	if refs.refs[out>>1][0] == 0 || refs.refs[out>>1][1] == 0 || refs.inverters != 1 || refs.gates != 2 || refs.edges != 3 {
		t.Fatalf("wrong mapping: %+v", refs)
	}
	if refs.refs[a>>1][0] != 1 || refs.refs[b>>1][0] != 1 {
		t.Fatalf("wrong input refs: %+v", refs.refs)
	}
	if matcher.required[a>>1][0] == nfInfinity || matcher.required[b>>1][0] == nfInfinity {
		t.Fatal("missing required times")
	}
	if pars.MapAreaF != refs.area || int32(pars.Area) != refs.gates {
		t.Fatal("statistics not recorded")
	}
}
