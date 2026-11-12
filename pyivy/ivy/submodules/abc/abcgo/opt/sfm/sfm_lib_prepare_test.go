package sfm

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestSfmPrepareEmptyLibrary(t *testing.T) {
	p := sfmLibPrepareCells(6, 1, 1, 0, 0, []mio.Mio_Cell2_t{})
	defer Sfm_LibStop(p)
	if len(p.pObjs) != 0 || p.nWords != 1 {
		t.Fatal("unexpected candidates")
	}
}

func TestSfmPrepareComposedCells(t *testing.T) {
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 2 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 2 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 2 0
.end
`
	lib, err := mio.ReadLibrary("fixture", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	cells := mio.Mio_CollectRootsNew2(lib, 6, nil, 0)
	p := sfmLibPrepareCells(7, 1, 1, 0, 0, cells)
	defer Sfm_LibStop(p)
	if len(p.pObjs) == 0 {
		t.Fatal("no prepared candidates")
	}
	if area, _ := Sfm_LibFindAreaMatch(p, []uint64{0x8888888888888888, 0x8888888888888888}); area < 0 {
		t.Fatal("missing AND gate")
	}
}
