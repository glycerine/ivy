package sfm

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestSfmLibCandidateDominance(t *testing.T) {
	lib := Sfm_LibStart(6, 1, 0)
	defer Sfm_LibStop(lib)
	expensive := &mio.Mio_Cell2_t{Id: 1, AreaW: 5}
	cheap := &mio.Mio_Cell2_t{Id: 2, AreaW: 3}
	truth := []uint64{0x8888888888888888}
	Sfm_LibPrepareAdd(lib, truth, nil, expensive, nil, 0)
	Sfm_LibPrepareAdd(lib, truth, nil, cheap, nil, 0)
	id := 2
	if lib.nObjRemoved != 1 || lib.pObjs[lib.lists[id]].Area != 3 {
		t.Fatal("did not replace dominated candidate")
	}
	Sfm_LibPrepareAdd(lib, truth, nil, expensive, nil, 0)
	if lib.nObjSkipped != 1 {
		t.Fatal("did not skip dominated candidate")
	}
}
