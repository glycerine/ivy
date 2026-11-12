package mfs

// Abc_MfsComputeRoots uses the original fanout-count, CO-fanout and level-cutoff
// rules. A dangling node is treated as its own root rather than an empty window.
func Abc_MfsComputeRoots(p *Mfs_Man_t, target, levels, limit int) []int {
	cutoff := p.levels[target] + levels
	seen := make([]bool, len(p.Network.Nodes))
	var roots []int
	var rec func(int)
	rec = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		root := p.outputDriver[id] || len(p.fanouts[id]) > limit
		for _, fo := range p.fanouts[id] {
			root = root || p.levels[fo] > cutoff
		}
		if root {
			roots = append(roots, id)
			return
		}
		for _, fo := range p.fanouts[id] {
			rec(fo)
		}
	}
	rec(target)
	if len(roots) == 0 {
		roots = append(roots, target)
	}
	return roots
}
func Abc_MfsWinMarkTfi(p *Mfs_Man_t, roots []int) []int {
	seen := make([]bool, len(p.Network.Nodes))
	var result []int
	var rec func(int)
	rec = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		for _, f := range p.Network.Nodes[id].Fanins {
			rec(f)
		}
		result = append(result, id)
	}
	for _, id := range roots {
		rec(id)
	}
	return result
}
func Abc_WinNode(p *Mfs_Man_t, target int) bool {
	Mfs_ManClean(p)
	p.Target = target
	p.Stats.NodesTried++
	p.Roots = Abc_MfsComputeRoots(p, target, p.Params.WinTfoLevs, p.Params.FanoutsMax)
	cone := Abc_MfsWinMarkTfi(p, p.Roots)
	for _, id := range cone {
		if p.Network.Nodes[id].Input {
			p.Support = append(p.Support, id)
		} else if id != 0 {
			p.Nodes = append(p.Nodes, id)
		}
	}
	if p.Params.Resub && p.Params.WinMax > 0 && len(p.Nodes) > p.Params.WinMax {
		p.Stats.WindowsSkipped++
		return false
	}
	if p.Params.Resub {
		p.Divisors = Abc_MfsComputeDivisors(p, target, p.required[target]-1)
	}
	Abc_NtkConstructAig(p, target)
	return true
}
