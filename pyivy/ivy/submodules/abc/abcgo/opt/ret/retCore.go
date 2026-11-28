package ret

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"sync/atomic"
	"time"
)

var timeRetime atomic.Int64

func retCleanupSeq(n *abc.Network) {
	seen := make(map[*abc.Object]bool)
	var visit func(*abc.Object)
	visit = func(o *abc.Object) {
		if seen[o] {
			return
		}
		seen[o] = true
		for _, e := range o.Fanins() {
			visit(e.Object)
		}
	}
	for _, o := range n.ObjectsOfKind(abc.ObjectPO) {
		visit(o)
	}
	for _, o := range n.Objects() {
		if o.Kind == abc.ObjectBlackbox || o.Kind == abc.ObjectWhitebox {
			visit(o)
			for _, out := range o.Fanouts() {
				visit(out)
			}
		}
	}
	for _, o := range n.Objects() {
		if !seen[o] && o.Kind != abc.ObjectPI && o.Kind != abc.ObjectPO && o.Kind != abc.ObjectConst1 {
			retDelete(o)
		}
	}
	abc.Abc_NtkOrderCisCos(n)
}
func Abc_NtkRetime(n *abc.Network, mode, delay, forwardOnly, backwardOnly, oneStep, useOldNames, verbose int32) int32 {
	if mode < 1 || mode > 6 || forwardOnly != 0 && backwardOnly != 0 {
		panic("ret: invalid mode or conflicting directions")
	}
	start := time.Now()
	before := len(retLatches(n))
	level := abc.Abc_NtkLevel(n)
	retCleanupSeq(n)
	var result int32
	switch mode {
	case 1:
		result = Abc_NtkRetimeIncremental(n, delay, 1, 0, 0, useOldNames, verbose)
	case 2:
		result = Abc_NtkRetimeIncremental(n, delay, 0, 0, 0, useOldNames, verbose)
	case 3:
		result = Abc_NtkRetimeMinArea(n, forwardOnly, backwardOnly, useOldNames, verbose)
	case 4:
		if backwardOnly == 0 {
			result += Abc_NtkRetimeIncremental(n, delay, 1, 1, oneStep, useOldNames, verbose)
		}
		if forwardOnly == 0 {
			result += Abc_NtkRetimeIncremental(n, delay, 0, 1, oneStep, useOldNames, verbose)
		}
	case 5:
		result = Abc_NtkRetimeMinArea(n, forwardOnly, backwardOnly, useOldNames, verbose)
		if backwardOnly == 0 {
			result += Abc_NtkRetimeIncremental(n, delay, 1, 1, 0, useOldNames, verbose)
		}
		if forwardOnly == 0 {
			result += Abc_NtkRetimeIncremental(n, delay, 0, 1, 0, useOldNames, verbose)
		}
	case 6:
		result = Abc_NtkRetimeLValue(n, 500, verbose)
	}
	elapsed := time.Since(start)
	timeRetime.Store(elapsed.Nanoseconds())
	if verbose != 0 {
		fmt.Printf("Reduction in area = %3d. Reduction in delay = %3d. Total runtime = %.2f sec\n", before-len(retLatches(n)), level-abc.Abc_NtkLevel(n), elapsed.Seconds())
	}
	return result
}
