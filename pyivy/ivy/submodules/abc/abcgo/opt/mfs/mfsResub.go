package mfs

import "github.com/glycerine/abcgo/sat/satoko"

// Abc_NtkMfsTryResubOnce returns 1 for feasible, 0 with a divisor
// counterexample, and -1 on resource exhaustion, matching source semantics.
func Abc_NtkMfsTryResubOnce(p *Mfs_Man_t, candidates []int) int {
	assume := make([]int, len(candidates))
	for i, c := range candidates {
		if c < 0 || c >= len(p.projection) {
			panic("mfs: invalid divisor index")
		}
		assume[i] = p.projection[c] ^ 1
	}
	status := solve(p, p.Solver, assume)
	if status == satoko.SATOKO_UNSAT {
		return 1
	}
	if status == satoko.SATOKO_UNDEC {
		return -1
	}
	witness := make([]bool, len(p.projection))
	for i, l := range p.projection {
		witness[i] = p.Solver.ModelValue(l>>1) == 1
	}
	p.witnesses = append(p.witnesses, witness)
	p.Stats.Counterexamples++
	return 0
}
func (p *Mfs_Man_t) covered(indices ...int) bool {
	for _, w := range p.witnesses {
		ok := false
		for _, i := range indices {
			ok = ok || w[i]
		}
		if !ok {
			return false
		}
	}
	return true
}
func (p *Mfs_Man_t) update(target int, candidates []int, skip bool) bool {
	if skip {
		return true
	}
	truth := Abc_NtkMfsInterplate(p, candidates)
	if truth == nil {
		return false
	}
	fanins := make([]int, len(candidates))
	for i, c := range candidates {
		fanins[i] = p.Divisors[c]
	}
	Abc_NtkMfsUpdateNetwork(p, target, fanins, truth)
	return true
}
func Abc_NtkMfsSolveSatResub(p *Mfs_Man_t, target, remove int, onlyRemove, skipUpdate bool) bool {
	node := p.Network.Nodes[target]
	if remove < 0 || remove >= len(node.Fanins) {
		panic("mfs: invalid fanin")
	}
	p.Stats.TryRemoves++
	p.witnesses = nil
	var candidates []int
	offset := len(p.Divisors) - len(node.Fanins)
	for i := range node.Fanins {
		if i != remove {
			candidates = append(candidates, offset+i)
		}
	}
	status := Abc_NtkMfsTryResubOnce(p, candidates)
	if status < 0 {
		return false
	}
	if status == 1 {
		if p.update(target, candidates, skipUpdate) {
			if !skipUpdate {
				p.Stats.Removes++
			}
			return true
		}
		return false
	}
	if onlyRemove || p.Params.RrOnly {
		return false
	}
	p.Stats.TryResubs++
	for {
		next := -1
		for d := 0; d < offset; d++ {
			if p.Params.Power && p.probabilities[p.Divisors[d]] >= 0.15 {
				continue
			}
			if p.covered(d) {
				next = d
				break
			}
		}
		if next < 0 {
			return false
		}
		trial := append(append([]int(nil), candidates...), next)
		status = Abc_NtkMfsTryResubOnce(p, trial)
		if status < 0 {
			return false
		}
		if status == 1 {
			if p.update(target, trial, skipUpdate) {
				if !skipUpdate {
					p.Stats.Resubs++
				}
				return true
			}
			return false
		}
		if p.Params.WinMax > 0 && len(p.witnesses) >= p.Params.WinMax {
			return false
		}
	}
}
func Abc_NtkMfsSolveSatResub2(p *Mfs_Man_t, target, remove, remove2 int) bool {
	node := p.Network.Nodes[target]
	if remove < 0 || remove >= len(node.Fanins) || remove2 < -1 || remove2 >= len(node.Fanins) || remove == remove2 {
		panic("mfs: invalid fanin pair")
	}
	p.witnesses = nil
	offset := len(p.Divisors) - len(node.Fanins)
	var candidates []int
	for i := range node.Fanins {
		if i != remove && i != remove2 {
			candidates = append(candidates, offset+i)
		}
	}
	status := Abc_NtkMfsTryResubOnce(p, candidates)
	if status < 0 {
		return false
	}
	if status == 1 {
		if p.update(target, candidates, false) {
			p.Stats.Removes++
			return true
		}
		return false
	}
	if p.Params.RrOnly {
		return false
	}
	for {
		a, b := -1, -1
		for i := 1; i < offset && a < 0; i++ {
			if p.Params.Power && p.probabilities[p.Divisors[i]] >= 0.12 {
				continue
			}
			for j := 0; j < i; j++ {
				if p.Params.Power && p.probabilities[p.Divisors[j]] >= 0.12 {
					continue
				}
				if p.covered(j, i) {
					a, b = j, i
					break
				}
			}
		}
		if a < 0 {
			return false
		}
		trial := append(append([]int(nil), candidates...), a, b)
		status = Abc_NtkMfsTryResubOnce(p, trial)
		if status < 0 {
			return false
		}
		if status == 1 {
			if p.update(target, trial, false) {
				p.Stats.Resubs++
				return true
			}
			return false
		}
		if p.Params.WinMax > 0 && len(p.witnesses) >= p.Params.WinMax {
			return false
		}
	}
}
func Abc_NtkMfsResubNode(p *Mfs_Man_t, target int) bool {
	node := p.Network.Nodes[target]
	for i, f := range node.Fanins {
		if !p.Network.Nodes[f].Input && f != 0 && len(p.fanouts[f]) == 1 {
			if Abc_NtkMfsSolveSatResub(p, target, i, false, false) {
				return true
			}
		}
	}
	if !p.Params.Area {
		for i, f := range node.Fanins {
			if p.Network.Nodes[f].Input || len(p.fanouts[f]) != 1 {
				if Abc_NtkMfsSolveSatResub(p, target, i, true, false) {
					return true
				}
			}
		}
	}
	return false
}
func Abc_NtkMfsResubNode2(p *Mfs_Man_t, target int) bool {
	node := p.Network.Nodes[target]
	for i, f := range node.Fanins {
		if !p.Network.Nodes[f].Input && f != 0 && len(p.fanouts[f]) == 1 {
			for j := range node.Fanins {
				if i != j && Abc_NtkMfsSolveSatResub2(p, target, i, j) {
					return true
				}
			}
		}
	}
	return false
}
func Abc_NtkMfsEdgeSwapEval(p *Mfs_Man_t, target int) int {
	feasible := 0
	for i := range p.Network.Nodes[target].Fanins {
		if Abc_NtkMfsSolveSatResub(p, target, i, false, true) {
			feasible++
		}
	}
	return feasible
}
func Abc_NtkMfsEdgePower(p *Mfs_Man_t, target int) bool {
	for i, f := range p.Network.Nodes[target].Fanins {
		prob := p.probabilities[f]
		if prob >= 0.35 {
			if Abc_NtkMfsSolveSatResub(p, target, i, false, false) {
				return true
			}
		} else if prob >= 0.25 {
			if Abc_NtkMfsSolveSatResub(p, target, i, true, false) {
				return true
			}
		}
	}
	return false
}
func Abc_NtkMfsUpdateNetwork(p *Mfs_Man_t, target int, fanins []int, truth []uint32) {
	old := &p.Network.Nodes[target]
	previous := append([]int(nil), old.Fanins...)
	old.Fanins = append([]int(nil), fanins...)
	old.Truth = normalizeTruth(truth, len(fanins))
	old.local = truthGraph(old.Truth, len(fanins))
	prune(p.Network, previous)
	if err := p.refresh(); err != nil {
		panic(err)
	}
}

func prune(n *Network, roots []int) {
	refs := make([]int, len(n.Nodes))
	for _, v := range n.Nodes {
		if !v.Deleted {
			for _, f := range v.Fanins {
				refs[f]++
			}
		}
	}
	for _, e := range n.Outputs {
		refs[e.ID()]++
	}
	var rec func(int)
	rec = func(id int) {
		v := &n.Nodes[id]
		if id == 0 || v.Input || v.Deleted || refs[id] != 0 {
			return
		}
		v.Deleted = true
		for _, f := range v.Fanins {
			refs[f]--
			rec(f)
		}
	}
	for _, id := range roots {
		rec(id)
	}
}
