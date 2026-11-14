// Mapped-gate pin lookup from verCore.c.
package ver

import "github.com/glycerine/abcgo/map/mio"

func Ver_FindGateInput(gate *mio.Mio_Gate_t, name string) int32 {
	var i int32
	for pin := mio.Mio_GateReadPins(gate); pin != nil; pin = mio.Mio_PinReadNext(pin) {
		if mio.Mio_PinReadName(pin) == name {
			return i
		}
		i++
	}
	if mio.Mio_GateReadOutName(gate) == name {
		return i
	}
	if twin := mio.Mio_GateReadTwin(gate); twin != nil && mio.Mio_GateReadOutName(twin) == name {
		return i + 1
	}
	return -1
}
