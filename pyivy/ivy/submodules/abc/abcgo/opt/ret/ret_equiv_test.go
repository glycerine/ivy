package ret

import "github.com/glycerine/abcgo/base/abc"

// retEquivalent compares two small initialized machines by exhaustive
// reachable-state exploration; the retiming tests use it as an oracle.
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
			values[l.FanoutAt(int32(0))] = state[i]
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
