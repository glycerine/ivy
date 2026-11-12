package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"reflect"
	"testing"
)

func TestNfExactAreaRestoresReferences(t *testing.T) {
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
	before := append([][2]int32(nil), refs.refs...)
	selected := matcher.objects[out>>1][0][refs.selected[out>>1][0]]
	removed := matcher.adjustMatchRefs(refs, out>>1, 0, selected, false)
	if removed != library.cells[selected.gate].AreaW {
		t.Fatalf("removed area = %d", removed)
	}
	detached := append([][2]int32(nil), refs.refs...)
	if area := matcher.incrementalArea(refs, out>>1, 0, selected); area != removed {
		t.Fatalf("incremental area %d, want %d", area, removed)
	}
	if !reflect.DeepEqual(detached, refs.refs) {
		t.Fatal("area evaluation changed references")
	}
	if restored := matcher.adjustMatchRefs(refs, out>>1, 0, selected, true); restored != removed || !reflect.DeepEqual(before, refs.refs) {
		t.Fatal("failed to restore references")
	}
}

func TestNfExactMatchRespectsTiming(t *testing.T) {
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
	Mf_ManSetDefaultPars(&pars)
	cuts := newNfCutManager(g, &pars, library)
	cuts.computeCuts()
	matcher := newNfMatcher(cuts)
	matcher.computeMapping()
	refs := matcher.mappingRefs(0)
	selected := matcher.objects[out>>1][0][refs.selected[out>>1][0]]
	matcher.adjustMatchRefs(refs, out>>1, 0, selected, false)
	detached := append([][2]int32(nil), refs.refs...)
	loose, _, ok := matcher.bestExactMatch(refs, out>>1, 0, nfInfinity-1)
	if !ok || library.cells[loose.gate].Name() != "and3" {
		t.Fatalf("wrong loose-timing match: %+v", loose)
	}
	tight, _, ok := matcher.bestExactMatch(refs, out>>1, 0, selected.delay)
	if !ok || library.cells[tight.gate].Name() != "and2" {
		t.Fatalf("wrong tight-timing match: %+v", tight)
	}
	if !reflect.DeepEqual(detached, refs.refs) {
		t.Fatal("search changed reference counts")
	}
}

func TestNfExactOptimizationUpdatesMapping(t *testing.T) {
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
	Mf_ManSetDefaultPars(&pars)
	cuts := newNfCutManager(g, &pars, library)
	cuts.computeCuts()
	matcher := newNfMatcher(cuts)
	matcher.computeMapping()
	refs := matcher.mappingRefs(0)
	matcher.required[out>>1][0] = nfInfinity - 1
	if !matcher.optimizeExactMatch(refs, out>>1, 0) {
		t.Fatal("expected a smaller implementation")
	}
	if refs.refs[x>>1][0] != 0 {
		t.Fatal("intermediate gate should be unreferenced")
	}
	matcher.refreshExactStats(refs)
	if refs.gates != 1 || refs.edges != 3 {
		t.Fatalf("wrong optimized statistics: %+v", refs)
	}
	matcher.deriveMapping(refs)
	if gate := Gia_ObjCellId(g, out); library.cells[gate].Name() != "and3" {
		t.Fatalf("wrong gate %s", library.cells[gate].Name())
	}
	if Gia_ObjCellSize(g, out) != 3 {
		t.Fatal("incorrect optimized fanins")
	}
}
