package ver

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestMappedGatePinAndTwinOutputLookup(t *testing.T) {
	source := `GATE pair 1 X=a; PIN * NONINV 1 1 1 1 1 1
GATE pair 1 Y=!a; PIN * INV 1 1 1 1 1 1`
	lib, err := mio.ReadLibrary("pair", source, nil)
	if err != nil {
		t.Fatal(err)
	}
	gate := mio.Mio_LibraryReadGateByName(lib, "pair", "")
	if gate == nil {
		t.Fatal("missing gate")
	}
	pin := mio.Mio_PinReadName(mio.Mio_GateReadPins(gate))
	out := mio.Mio_GateReadOutName(gate)
	twin := mio.Mio_GateReadTwin(gate)
	if twin == nil {
		t.Fatal("missing twin")
	}
	for _, tc := range []struct {
		name string
		want int32
	}{{pin, 0}, {out, 1}, {mio.Mio_GateReadOutName(twin), 2}, {"missing", -1}} {
		if got := Ver_FindGateInput(gate, tc.name); got != tc.want {
			t.Errorf("%s: got %d want %d", tc.name, got, tc.want)
		}
	}
}
