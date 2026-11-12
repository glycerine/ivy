package sfm

func (p *Sfm_Ntk_t) collect(roots []int, limit int) ([]int, bool) {
	seen := make([]bool, len(p.Fanins))
	var result []int
	var rec func(int) bool
	rec = func(id int) bool {
		if seen[id] {
			return true
		}
		seen[id] = true
		for _, f := range p.Fanins[id] {
			if !rec(f) {
				return false
			}
		}
		result = append(result, id)
		return limit == 0 || len(result) <= limit
	}
	for _, id := range roots {
		if !rec(id) {
			return result, false
		}
	}
	return result, true
}
func Sfm_ObjMffcSize(p *Sfm_Ntk_t, target int) int {
	if Sfm_ObjIsPi(p, target) || len(p.Fanouts[target]) != 1 {
		return 0
	}
	refs := make([]int, len(p.Fanins))
	for id := range refs {
		refs[id] = len(p.Fanouts[id])
	}
	count := 0
	var rec func(int)
	rec = func(id int) {
		if !Sfm_ObjIsNode(p, id) {
			return
		}
		count++
		for _, f := range p.Fanins[id] {
			refs[f]--
			if refs[f] == 0 {
				rec(f)
			}
		}
	}
	rec(target)
	return count
}
func (p *Sfm_Ntk_t) isRoot(id, cutoff int) bool {
	if len(p.Fanouts[id]) == 0 || len(p.Fanouts[id]) > p.Params.FanoutMax {
		return true
	}
	for _, fo := range p.Fanouts[id] {
		if Sfm_ObjIsPo(p, fo) || p.Levels[fo] > cutoff {
			return true
		}
	}
	return false
}
func Sfm_NtkCreateWindow(p *Sfm_Ntk_t, target int, _ bool) bool {
	p.pivot = target
	p.nodes = nil
	p.divs = nil
	p.roots = nil
	p.tfo = nil
	p.order = nil
	nodes, ok := p.collect([]int{target}, p.Params.WinSizeMax)
	p.nodes = nodes
	if !ok {
		p.Stats.MaxDivs++
		return false
	}
	p.nodes = nodes
	p.divs = append([]int(nil), nodes[:len(nodes)-1]...)
	initial := make([]bool, len(p.Fanins))
	for _, id := range nodes {
		initial[id] = true
	}
	visits := make([]int, len(initial))
	maxlevel := p.LevelMax - p.Reverse[target]
	for i := 0; i < len(p.divs) && len(p.divs) < p.Params.WinSizeMax; i++ {
		for pin, fo := range p.Fanouts[p.divs[i]] {
			if p.Params.FanoutMax > 0 && pin > p.Params.FanoutMax {
				break
			}
			if initial[fo] || Sfm_ObjIsPo(p, fo) || p.Levels[fo] > maxlevel {
				continue
			}
			visits[fo]++
			if visits[fo] == len(p.Fanins[fo]) {
				p.divs = append(p.divs, fo)
			}
		}
	}
	if len(p.divs) > p.Params.WinSizeMax {
		p.divs = p.divs[:p.Params.WinSizeMax]
	}
	if len(p.divs) == p.Params.WinSizeMax {
		p.Stats.MaxDivs++
	}
	exclude := make([]bool, len(initial))
	exclude[target] = true
	for _, f := range p.Fanins[target] {
		exclude[f] = true
	}
	filtered := p.divs[:0]
	for _, id := range p.divs {
		useful := !p.Fixed[id]
		for _, fo := range p.Fanouts[id] {
			useful = useful || !p.Fixed[fo]
		}
		if !exclude[id] && useful {
			filtered = append(filtered, id)
		}
	}
	p.divs = filtered
	p.Stats.TotalDivs += len(p.divs)
	cutoff := p.Levels[target] + p.Params.TfoLevMax
	if p.Params.TfoLevMax > 0 && !p.isRoot(target, cutoff) {
		seen := make([]bool, len(initial))
		var rec func(int)
		rec = func(id int) {
			if seen[id] {
				return
			}
			seen[id] = true
			if id != target {
				p.tfo = append(p.tfo, id)
			}
			if p.isRoot(id, cutoff) {
				p.roots = append(p.roots, id)
			} else {
				for _, fo := range p.Fanouts[id] {
					rec(fo)
				}
			}
		}
		rec(target)
		all := append(append(append([]int(nil), p.roots...), p.tfo...), p.divs...)
		if order, ok := p.collect(all, p.Params.WinSizeMax); ok {
			p.order = order
		} else {
			p.roots = nil
			p.tfo = nil
		}
	}
	if len(p.order) == 0 {
		p.order, _ = p.collect(append([]int{target}, p.divs...), 0)
	}
	return true
}

// Sfm_NtkDfs preserves grouped fixed boxes atomically. groupMap entries encode
// group<<1|phase as in C; phases do not affect group traversal.
func Sfm_NtkDfs(p *Sfm_Ntk_t, groups [][]int, groupMap []int, allBoxes bool) ([]int, []int) {
	seen := make([]bool, len(p.Fanins))
	var nodes, boxes []int
	var rec func(int)
	rec = func(id int) {
		if Sfm_ObjIsPi(p, id) || seen[id] || p.Deleted[id] {
			return
		}
		group := -1
		if len(groupMap) > id && groupMap[id] >= 0 {
			group = groupMap[id] >> 1
		}
		if group >= 0 {
			members := groups[group]
			for _, v := range members {
				seen[v] = true
			}
			for _, v := range members {
				for _, f := range p.Fanins[v] {
					rec(f)
				}
			}
			nodes = append(nodes, members...)
			boxes = append(boxes, group)
		} else {
			seen[id] = true
			for _, f := range p.Fanins[id] {
				rec(f)
			}
			nodes = append(nodes, id)
		}
	}
	if allBoxes {
		for _, g := range groups {
			if len(g) > 0 {
				rec(g[0])
			}
		}
	}
	for id := len(p.Fanins) - p.Pos; id < len(p.Fanins); id++ {
		rec(p.Fanins[id][0])
	}
	return nodes, boxes
}
