package sfm

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestSfmAreaMatch(t *testing.T) {
	lib := Sfm_LibStart(6, 0, 0)
	defer Sfm_LibStop(lib)
	cell := &mio.Mio_Cell2_t{Id: 1, AreaW: 4}
	truth := []uint64{0x8888888888888888}
	Sfm_LibPrepareAdd(lib, truth, nil, cell, nil, 0)
	area, id := Sfm_LibFindAreaMatch(lib, truth)
	if area != 4 || id != 0 {
		t.Fatalf("area %d id %d", area, id)
	}
	if area, id = Sfm_LibFindAreaMatch(lib, []uint64{0xf0f0f0f0f0f0f0f0}); area != -1 || id != -1 {
		t.Fatal("unknown match")
	}
}
