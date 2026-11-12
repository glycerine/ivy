package sfm

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestSfmDecSimulationCare(t *testing.T) {
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
	gate := mio.Mio_LibraryReadGateByName(lib, "and2", "")
	s := newSfmDecSimulation(4)
	s.normal[0], s.normal[1] = 0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc
	s.update(2, gate, []int{0, 1})
	s.flip(2)
	s.affected[3] = true
	s.update(3, gate, []int{2, 0})
	if got := s.care([]int{3}); got != s.normal[0] {
		t.Fatalf("care %x", got)
	}
}
