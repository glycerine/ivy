package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestPfLibraryMatches(t *testing.T) {
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
	library := pfPrepareLibrary(mio.Mio_CollectRootsNew(lib, 6, nil, 0))
	id := *vec.Vec_MemHashLookup(library.truth.memory, []uint64{0x8888888888888888})
	if id < 0 || int(id) >= len(library.matches) || len(library.matches[id]) == 0 {
		t.Fatal("missing AND matches")
	}
	for _, match := range library.matches[id] {
		if pfIntConfig(pfConfigInt(match.config)) != match.config {
			t.Fatal("configuration changed")
		}
	}
	if library.truth.matchCount != len(library.matches) {
		t.Fatal("incorrect match count")
	}
}

func TestPfEnumerateCuts(t *testing.T) {
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
	library := pfPrepareLibrary(mio.Mio_CollectRootsNew(lib, 6, nil, 0))
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	out := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, out)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	manager := library.cutManager(g, &pars)
	manager.computeCuts()
	if manager.sets[out>>1] == 0 || manager.counts[4] == 0 {
		t.Fatal("missing useful Pf cuts")
	}
}
