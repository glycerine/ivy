package sfm

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/map/mio"
)

func TestSfmMappedAreaRunsPolarityPass(t *testing.T) {
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 3 O=a*b; PIN * NONINV 1 999 1 0 1 0
GATE nand2 1 O=!(a*b); PIN * INV 1 999 1 0 1 0
.end
`
	lib, err := mio.ReadLibrary("polarity", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2}}, 2, 1, []uint64{0, 0, 8, 0})
	p.Mapping[2] = Sfm_GateFromMio(mio.Mio_LibraryReadGateByName(lib, "and2", ""))
	want := coreTable(t, p)
	pars := Sfm_Par_t{Area: true, VarMax: 6, MffcMax: 3, NodeOne: 99, FanoutMax: 10}
	changed, err := Sfm_OptimizeMapped(p, pars, lib)
	if err != nil || changed == 0 || !reflect.DeepEqual(coreTable(t, p), want) {
		t.Fatalf("changed=%d err=%v", changed, err)
	}
	if p.Mapping[2].Name != "nand2" {
		t.Fatalf("gate %s", p.Mapping[2].Name)
	}
}
