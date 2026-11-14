package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestNfDriverMapsNetwork(t *testing.T) {
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 1 0
GATE and3 1 O=a*b*c; PIN * NONINV 1 999 5 0 5 0
.end
`
	lib, err := mio.ReadLibrary("nf", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	library := nfPrepareLibrary(mio.Mio_CollectRootsNew2(lib, 6, nil, 0), false, true, false)
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	out := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, out)
	var pars Jf_Par_t
	nfDefaultParameters(&pars)
	mapped := nfMap(g, &pars, library)
	if mapped != g || Gia_ManHasCellMapping(mapped) == 0 || Gia_ObjCellSize(mapped, out) != 2 || pars.Area != 2 {
		t.Fatalf("wrong mapping: area=%d", pars.Area)
	}
}

func TestNfDriverUsesInputAndOutputTiming(t *testing.T) {
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
	g.VInArrs = vec.Vec_FltAlloc(2)
	vec.Vec_FltPush(g.VInArrs, 2)
	vec.Vec_FltPush(g.VInArrs, 0)
	g.VOutReqs = vec.Vec_FltAlloc(1)
	vec.Vec_FltPush(g.VOutReqs, 5)
	var pars Jf_Par_t
	nfDefaultParameters(&pars)
	cuts := newNfCutManager(g, &pars, library)
	cuts.computeCuts()
	matcher := newNfMatcher(cuts)
	matcher.computeMapping()
	refs := matcher.mappingRefs(0)
	if matcher.objects[a>>1][0][0].delay != scl.Scl_Flt2Int(2) {
		t.Fatal("missing input arrival")
	}
	if matcher.required[out>>1][0] != scl.Scl_Flt2Int(5) {
		t.Fatalf("incorrect output required time: %d", matcher.required[out>>1][0])
	}
	if refs.delay != scl.Scl_Flt2Int(3) {
		t.Fatalf("incorrect mapped delay: %d", refs.delay)
	}
}
