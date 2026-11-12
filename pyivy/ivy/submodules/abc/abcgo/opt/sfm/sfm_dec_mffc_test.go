package sfm

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/map/mio"
)

func TestSfmMappedMffcArea(t *testing.T) {
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	p.Mapping[2], p.Mapping[3] = &Gate{Area: 3}, &Gate{Area: 5}
	before := append([]int(nil), p.Fanouts[2]...)
	area, cone := sfmMappedMffcArea(p, 3, nil)
	if area != 8 || !reflect.DeepEqual(cone, []int{2, 3}) {
		t.Fatalf("area %d cone %v", area, cone)
	}
	area, cone = sfmMappedMffcArea(p, 3, []int{2})
	if area != 5 || !reflect.DeepEqual(cone, []int{3}) || !reflect.DeepEqual(before, p.Fanouts[2]) {
		t.Fatalf("cut: area %d cone %v", area, cone)
	}
}

func TestSfmFindAreaNodeMatch(t *testing.T) {
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
	Sfm_LibPreprocess(lib)
	prepared := sfmLibPrepareCells(6, 0, 0, 0, 0, mio.Mio_CollectRootsNew2(lib, 6, nil, 0))
	defer Sfm_LibStop(prepared)
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	p.Mapping[2], p.Mapping[3] = &Gate{Area: 2000}, &Gate{Area: 2000}
	pars := Sfm_Par_t{TfoLevMax: 2, TfiLevMax: 3, FanoutMax: 10, MffcMax: 3, VarMax: 6, UseAndOr: true}
	gates, fanins, order, cone, ok := sfmFindAreaNodeMatch(p, 3, pars, prepared)
	if !ok || len(gates) != 1 || len(fanins) != 1 || len(fanins[0]) != 2 || !reflect.DeepEqual(cone, []int{2, 3}) || len(order) != 4 {
		t.Fatalf("match: %v %v %v %v %v", gates, fanins, order, cone, ok)
	}
}
