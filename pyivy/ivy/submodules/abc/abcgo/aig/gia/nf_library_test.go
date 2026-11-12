package gia

import (
	"testing"

	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestNfLibraryEnumeratesGateMatches(t *testing.T) {
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 2 0
.end
`
	library, err := mio.ReadLibrary("nf", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	cells := mio.Mio_CollectRootsNew2(library, 6, nil, 0)
	for _, options := range [][3]bool{{false, true, false}, {true, false, false}, {false, false, true}} {
		store := nfPrepareLibrary(cells, options[0], options[1], options[2])
		id := *vec.Vec_MemHashLookup(store.truth.memory, []uint64{0x8888888888888888})
		if id < 0 || int(id) >= len(store.matches) || len(store.matches[id]) == 0 || store.truth.matchCount != len(store.matches) {
			t.Fatalf("missing AND matches: %v", options)
		}
		for _, m := range store.matches[id] {
			if Nf_Int2Cfg(Nf_Cfg2Int(m.cfg)) != m.cfg {
				t.Fatal("configuration changed")
			}
		}
	}
}
