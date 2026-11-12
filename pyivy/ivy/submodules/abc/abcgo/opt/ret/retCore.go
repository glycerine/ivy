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

// Debug retimes a duplicate backward for area and compares the two concrete
// initialized machines by exhaustive reachable-state exploration. This is a
// diagnostic routine; the ordinary retiming algorithms do not use enumeration.
func Abc_NtkRetimeDebug(n *abc.Network) int32 {
	copy := abc.Abc_NtkDup(n)
	Abc_NtkRetime(copy, 3, 0, 0, 1, 0, 1, 0)
	return retBool(!retEquivalent(n, copy))
}
func retEquivalent(a, b *abc.Network) bool {
	pi0, pi1 := a.ObjectsOfKind(abc.ObjectPI), b.ObjectsOfKind(abc.ObjectPI)
	po0, po1 := a.ObjectsOfKind(abc.ObjectPO), b.ObjectsOfKind(abc.ObjectPO)
	if len(pi0) != len(pi1) || len(po0) != len(po1) {
		return false
	}
	l0, l1 := retLatches(a), retLatches(b)
	initial := make([]bool, len(l0)+len(l1))
	for i, l := range append(append([]*abc.Object(nil), l0...), l1...) {
		initial[i] = retLatchValue(l) != 0
	}
	queue := [][]bool{initial}
	seen := make(map[string]bool)
	key := func(bits []bool) string {
		v := make([]byte, len(bits))
		for i, b := range bits {
			if b {
				v[i] = 1
			}
		}
		return string(v)
	}
	seen[key(initial)] = true
	var inputs func(int, []bool, func([]bool) bool) bool
	inputs = func(i int, in []bool, visit func([]bool) bool) bool {
		if i == len(in) {
			return visit(in)
		}
		in[i] = false
		if !inputs(i+1, in, visit) {
			return false
		}
		in[i] = true
		return inputs(i+1, in, visit)
	}
	step := func(n *abc.Network, pis, pos, latches []*abc.Object, state, in []bool) ([]bool, []bool) {
		values := make(map[*abc.Object]bool)
		for i, o := range pis {
			values[o] = in[i]
		}
		for i, l := range latches {
			values[l.Fanouts()[0]] = state[i]
		}
		outs, next := make([]bool, len(pos)), make([]bool, len(latches))
		for i, o := range pos {
			v, err := n.Evaluate(o.Signal(), values)
			if err != nil {
				panic(err)
			}
			outs[i] = v
		}
		for i, l := range latches {
			v, err := n.Evaluate(retFan0(l).Signal(), values)
			if err != nil {
				panic(err)
			}
			next[i] = v
		}
		return outs, next
	}
	for head := 0; head < len(queue); head++ {
		state := queue[head]
		ok := inputs(0, make([]bool, len(pi0)), func(in []bool) bool {
			o0, s0 := step(a, pi0, po0, l0, state[:len(l0)], in)
			o1, s1 := step(b, pi1, po1, l1, state[len(l0):], in)
			for i := range o0 {
				if o0[i] != o1[i] {
					return false
				}
			}
			next := append(s0, s1...)
			k := key(next)
			if !seen[k] {
				seen[k] = true
				queue = append(queue, next)
			}
			return true
		})
		if !ok {
			return false
		}
	}
	return true
}
