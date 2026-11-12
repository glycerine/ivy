package mfs

import "sort"

func Abc_MfsComputeDivisors(p *Mfs_Man_t, target, maxLevel int) []int {
	cone := Abc_MfsWinMarkTfi(p, []int{target})
	inCone := make([]bool, len(p.Network.Nodes))
	for _, id := range cone {
		inCone[id] = true
	}
	blocked := make([]bool, len(inCone))
	var block func(int)
	block = func(id int) {
		if blocked[id] || p.levels[id] > maxLevel {
			return
		}
		blocked[id] = true
		for _, fo := range p.fanouts[id] {
			block(fo)
		}
	}
	block(target)
	blocked[target] = true
	for _, f := range p.Network.Nodes[target].Fanins {
		blocked[f] = true
	}
	var divs []int
	full := func() bool { return len(divs) >= p.Params.WinMax }
	for _, id := range cone {
		if !blocked[id] && p.levels[id] <= maxLevel {
			divs = append(divs, id)
			if full() {
				break
			}
		}
	}
	for k := 0; k < len(divs) && !full(); k++ {
		for f, fo := range p.fanouts[divs[k]] {
			if p.Params.FanoutsMax > 0 && f > p.Params.FanoutsMax {
				break
			}
			v := p.Network.Nodes[fo]
			if inCone[fo] || blocked[fo] || v.Input || p.levels[fo] > maxLevel {
				continue
			}
			ok := true
			for _, fi := range v.Fanins {
				ok = ok && inCone[fi] && !blocked[fi]
			}
			if !ok {
				continue
			}
			divs = append(divs, fo)
			inCone[fo] = true
			p.Nodes = appendUnique(p.Nodes, fo)
			if full() {
				break
			}
		}
	}
	sort.SliceStable(divs, func(i, j int) bool {
		if p.levels[divs[i]] != p.levels[divs[j]] {
			return p.levels[divs[i]] < p.levels[divs[j]]
		}
		return divs[i] < divs[j]
	})
	return append(divs, p.Network.Nodes[target].Fanins...)
}
func appendUnique(a []int, v int) []int {
	for _, x := range a {
		if x == v {
			return a
		}
	}
	return append(a, v)
}

// Abc_MfsWinVisitMffc computes dereference/ref closure on private counts.
func Abc_MfsWinVisitMffc(p *Mfs_Man_t, target int) []int {
	refs := make([]int, len(p.Network.Nodes))
	for id := range refs {
		refs[id] = len(p.fanouts[id])
	}
	for _, e := range p.Network.Outputs {
		refs[e.ID()]++
	}
	var result []int
	var rec func(int)
	rec = func(id int) {
		if p.Network.Nodes[id].Input || id == 0 {
			return
		}
		result = append(result, id)
		for _, f := range p.Network.Nodes[id].Fanins {
			refs[f]--
			if refs[f] == 0 {
				rec(f)
			}
		}
	}
	rec(target)
	return result
}
