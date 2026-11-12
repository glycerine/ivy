package ret

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
)

func Abc_NtkRetimeMinDelay(n, copy *abc.Network, limit, iterations, forward, verbose int32) int32 {
	var best, best2 int32
	delay := Abc_NtkRetimeMinDelayTry(copy, limit, forward, 0, iterations, &best, verbose)
	if best == 0 {
		return 1
	}
	delay2 := Abc_NtkRetimeMinDelayTry(n, limit, forward, 1, best, &best2, verbose)
	if delay != delay2 || best != best2 {
		panic("ret: replayed delay analysis differs")
	}
	return 1
}
func Abc_NtkRetimeMinDelayTry(n *abc.Network, limit, forward, initial, iterations int32, bestIteration *int32, verbose int32) int32 {
	var cone *abc.Network
	var values []int32
	if initial != 0 {
		if forward != 0 {
			Abc_NtkRetimeTranferToCopy(n)
		} else {
			values = Abc_NtkRetimeCollectLatchValues(n)
			cone = Abc_NtkRetimeBackwardInitialStart(n)
		}
	}
	best := int32(1 << 30)
	bestIter := int32(0)
	start := int32(0)
	for i := int32(0); ; i++ {
		var critical []*abc.Object
		delay := Abc_NtkRetimeTiming(n, forward, &critical)
		if i == 0 {
			start = delay
		}
		if delay < best {
			best, bestIter = delay, i
			if verbose != 0 && initial == 0 {
				fmt.Printf("Iter = %3d. Delay = %3d. Latches = %5d.\n", i, delay, len(retLatches(n)))
			}
		}
		if i == iterations || i-bestIter > 20 || limit > 0 && delay <= limit {
			break
		}
		for _, o := range critical {
			if o.Kind == abc.ObjectNode && Abc_NtkRetimeNodeIsEnabled(o, forward) != 0 {
				Abc_NtkRetimeNode(o, forward, initial)
			}
		}
		if forward == 0 {
			Abc_NtkRetimeShareLatches(n, initial)
		}
	}
	if initial != 0 {
		if forward != 0 {
			Abc_NtkRetimeTranferFromCopy(n)
		} else {
			Abc_NtkRetimeBackwardInitialFinish(n, cone, values, verbose)
		}
	}
	if verbose != 0 && initial == 0 {
		fmt.Printf("Starting delay = %3d. Final delay = %3d. IterBest = %2d.\n", start, best, bestIter)
	}
	*bestIteration = bestIter
	if iterations == 1 {
		*bestIteration = 1
	}
	return best
}
func Abc_NtkRetimeTiming(n *abc.Network, forward int32, critical *[]*abc.Object) int32 {
	abc.Abc_NtkIncrementTravId(n)
	for _, o := range n.Objects() {
		if o.Kind != abc.ObjectNode {
			o.Level = 0
			abc.Abc_NodeSetTravIdCurrent(o)
		}
	}
	maximum := int32(0)
	for _, latch := range retLatches(n) {
		if forward != 0 {
			for _, o := range latch.Fanouts() {
				maximum = max(maximum, Abc_NtkRetimeTiming_rec(o, forward))
			}
		} else {
			maximum = max(maximum, Abc_NtkRetimeTiming_rec(retFan0(latch), forward))
		}
	}
	if forward != 0 {
		for _, pi := range n.ObjectsOfKind(abc.ObjectPI) {
			for _, o := range pi.Fanouts() {
				maximum = max(maximum, Abc_NtkRetimeTiming_rec(o, forward))
			}
		}
	} else {
		for _, po := range n.ObjectsOfKind(abc.ObjectPO) {
			maximum = max(maximum, Abc_NtkRetimeTiming_rec(retFan0(po), forward))
		}
	}
	*critical = (*critical)[:0]
	seen := make(map[*abc.Object]bool)
	for _, latch := range retLatches(n) {
		var candidates []*abc.Object
		if forward != 0 {
			candidates = latch.Fanouts()
		} else {
			for _, e := range latch.Fanins() {
				candidates = append(candidates, e.Object)
			}
		}
		for _, o := range candidates {
			if !seen[o] && int32(o.Level) == maximum {
				*critical = append(*critical, o)
				seen[o] = true
			}
		}
	}
	return maximum
}
func Abc_NtkRetimeTiming_rec(o *abc.Object, forward int32) int32 {
	if abc.Abc_NodeIsTravIdCurrent(o) != 0 {
		return int32(o.Level)
	}
	abc.Abc_NodeSetTravIdCurrent(o)
	level := int32(0)
	if forward != 0 {
		for _, next := range o.Fanouts() {
			level = max(level, Abc_NtkRetimeTiming_rec(next, forward))
		}
	} else {
		for _, e := range o.Fanins() {
			level = max(level, Abc_NtkRetimeTiming_rec(e.Object, forward))
		}
	}
	o.Level = int(level + 1)
	return level + 1
}
