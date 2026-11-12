package abc

import (
	"sync"

	"github.com/glycerine/abcgo/misc/vec"
)

type LatchSweep func(*Network, int32, int32, int32, int32, int32, int32, int32, int32) *Network

var propertyLatchSweep struct {
	sync.RWMutex
	fn LatchSweep
}

// RegisterLatchSweep supplies the sequential cleanup implemented by abci.
func RegisterLatchSweep(fn LatchSweep) {
	propertyLatchSweep.Lock()
	propertyLatchSweep.fn = fn
	propertyLatchSweep.Unlock()
}

// Abc_NtkCreatePropertyMonitor builds and sequentially cleans the monitor.
func Abc_NtkCreatePropertyMonitor(p *Network, ids, values *vec.Vec_Int_t) *Network {
	if p == nil || Abc_NtkIsLogic(p) == 0 || Abc_NtkLatchNum(p) == 0 || ids == nil || values == nil || vec.Vec_IntSize(ids) == 0 || vec.Vec_IntSize(ids) != vec.Vec_IntSize(values) {
		panic("abc: invalid sequential property monitor input")
	}
	propertyLatchSweep.RLock()
	sweep := propertyLatchSweep.fn
	propertyLatchSweep.RUnlock()
	if sweep == nil {
		panic("abc: sequential latch sweep is not registered")
	}
	monitor := Abc_NtkDeriveWithOnePo(p, ids, values)
	strashed := Abc_NtkStrash(monitor, 0, 1, 0)
	Abc_NtkDelete(monitor)
	if strashed == nil {
		return nil
	}
	result := sweep(strashed, 1, 1, 1, 0, 1, 512, 0, 0)
	Abc_NtkDelete(strashed)
	return result
}

// Abc_NtkCreatePropertyMonitorTest uses the example IDs from ABC's testbench.
func Abc_NtkCreatePropertyMonitorTest(p *Network) *Network {
	ids := vec.Vec_IntAllocArrayCopy([]int32{90, 80, 100}, 3)
	values := vec.Vec_IntAllocArrayCopy([]int32{1, 0, 1}, 3)
	return Abc_NtkCreatePropertyMonitor(p, ids, values)
}
