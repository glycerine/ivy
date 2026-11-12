package ret

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
)

const retInfinity = 1000000000

func Abc_NodeComputeLag(value, period int32) int32 {
	if period <= 0 {
		panic("ret: period must be positive")
	}
	return (value+(1<<16)*period)/period - (1 << 16) - retBool(value%period == 0)
}
func Abc_NodeGetLValue(o *abc.Object) int32        { return int32(o.Temp) }
func Abc_NodeSetLValue(o *abc.Object, value int32) { o.Temp = int(value) }

// As in the original active code, Pan's entry point computes lags and reports
// the best period; applying those lags is deliberately not enabled by mode 6.
func Abc_NtkRetimeLValue(n *abc.Network, iterations, verbose int32) int32 {
	Abc_NtkRetimeGetLags(n, iterations, verbose)
	return 0
}
func Abc_NtkRetimeGetLags(n *abc.Network, iterations, verbose int32) []int32 {
	upper := abc.Abc_NtkLevel(n)
	nodes := abc.Abc_NtkDfs(n, 0)
	latches := Abc_ManCollectLatches(n)
	lags := make([]int32, abc.Abc_NtkObjNumMax(n)+1)
	if upper <= 0 {
		return lags
	}
	if Abc_NtkRetimeForPeriod(n, nodes, latches, upper, iterations, verbose) == 0 {
		fmt.Println("Abc_NtkRetimeGetLags() error: The upper bound on the clock period cannot be computed.")
		return lags
	}
	best := Abc_NtkRetimeSearch_rec(n, nodes, latches, 0, upper, iterations, verbose)
	if Abc_NtkRetimeForPeriod(n, nodes, latches, best, iterations, verbose) == 0 {
		panic("ret: feasible period changed")
	}
	for _, o := range retNodes(n) {
		if o.Temp < -retInfinity/2 {
			o.Temp = 0
		}
		lags[o.ID] = int32(int8(Abc_NodeComputeLag(int32(o.Temp), best)))
	}
	fmt.Printf("The best clock period is %3d. (Currently, network is not modified.)\n", best)
	return lags
}
func Abc_NtkRetimeSearch_rec(n *abc.Network, nodes, latches []*abc.Object, low, high, iterations, verbose int32) int32 {
	if low >= high {
		panic("ret: invalid period interval")
	}
	if low+1 == high {
		return high
	}
	middle := low + (high-low)/2
	if Abc_NtkRetimeForPeriod(n, nodes, latches, middle, iterations, verbose) != 0 {
		return Abc_NtkRetimeSearch_rec(n, nodes, latches, low, middle, iterations, verbose)
	}
	return Abc_NtkRetimeSearch_rec(n, nodes, latches, middle, high, iterations, verbose)
}
func Abc_NtkRetimeForPeriod(n *abc.Network, nodes, latches []*abc.Object, period, iterations, verbose int32) int32 {
	for _, o := range n.Objects() {
		o.Temp = -retInfinity
		if len(o.Fanins()) == 0 {
			o.Temp = 0
		}
	}
	converged := false
	i := int32(1)
	for ; i <= iterations; i++ {
		if Abc_NtkRetimeUpdateLValue(n, nodes, latches, period) == 0 {
			converged = true
			break
		}
		if Abc_NtkRetimePosOverLimit(n, period) != 0 {
			break
		}
	}
	if verbose != 0 {
		fmt.Printf("Period = %3d. Iterations = %3d. Feasible = %t\n", period, i, converged)
	}
	return retBool(converged)
}
func Abc_NtkRetimeUpdateLValue(n *abc.Network, nodes, latches []*abc.Object, period int32) int32 {
	changed := false
	for _, o := range nodes {
		v := int32(-retInfinity)
		for _, e := range o.Fanins() {
			v = max(v, int32(e.Object.Temp))
		}
		v++
		if int32(o.Temp) < v {
			o.Temp = int(v)
			changed = true
		}
	}
	for _, latch := range latches {
		latch.Fanouts()[0].Temp = retFan0(retFan0(latch)).Temp - int(period)
	}
	return retBool(changed)
}
func Abc_NtkRetimePosOverLimit(n *abc.Network, period int32) int32 {
	for _, o := range n.ObjectsOfKind(abc.ObjectPO) {
		if retFan0(o).Temp > int(period) {
			return 1
		}
	}
	return 0
}
func Abc_ManCollectLatches_rec(o *abc.Object, latches *[]*abc.Object) {
	if o.Kind != abc.ObjectLatch || abc.Abc_NodeIsTravIdCurrent(o) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(o)
	driver := retFan0(retFan0(o))
	if driver.Kind == abc.ObjectBO {
		Abc_ManCollectLatches_rec(retFan0(driver), latches)
	}
	*latches = append(*latches, o)
}
func Abc_ManCollectLatches(n *abc.Network) []*abc.Object {
	abc.Abc_NtkIncrementTravId(n)
	var latches []*abc.Object
	for _, o := range retLatches(n) {
		Abc_ManCollectLatches_rec(o, &latches)
	}
	return latches
}
func Abc_NtkRetimeUsingLags(n *abc.Network, lags []int32, verbose int32) int32 {
	moves := 0
	for {
		changed := false
		for _, o := range retNodes(n) {
			lag := lags[o.ID]
			if lag == 0 {
				continue
			}
			forward := retBool(lag < 0)
			if Abc_NtkRetimeNodeIsEnabled(o, forward) != 0 {
				Abc_NtkRetimeNode(o, forward, 0)
				changed = true
				moves++
				if forward != 0 {
					lags[o.ID]++
				} else {
					lags[o.ID]--
				}
			}
		}
		if !changed {
			break
		}
	}
	if verbose != 0 {
		fmt.Printf("Total latch moves = %d.\n", moves)
	}
	remaining := 0
	for _, o := range retNodes(n) {
		if lags[o.ID] != 0 {
			remaining++
		}
	}
	if remaining != 0 {
		fmt.Printf("Warning! The number of nodes with unrealized lag = %d.\n", remaining)
	}
	return 1
}
