package exor

import "sort"

// ExorLink uses the telescoping product identity over GF(2): replace one
// factor by A XOR B, earlier factors by B and later factors by A. Permuting
// differing dimensions yields exactly d! groups of d equivalent cubes.
func ExorLink(a, b *Cube, maxDistance int) [][]*Cube {
	diff := FindDiffVars(a, b)
	if len(diff) == 0 || len(diff) > maxDistance || len(diff) > 4 {
		return nil
	}
	var groups [][]*Cube
	order := append([]int(nil), diff...)
	var visit func(int)
	visit = func(k int) {
		if k == len(order) {
			group := make([]*Cube, len(order))
			for pivot := range order {
				c := a.Clone()
				for j, d := range order {
					if d < 0 {
						if j < pivot {
							copy(c.Outputs, b.Outputs)
						} else if j == pivot {
							for w := range c.Outputs {
								c.Outputs[w] = a.Outputs[w] ^ b.Outputs[w]
							}
						}
					} else {
						value := GetVar(a, d)
						if j < pivot {
							value = GetVar(b, d)
						} else if j == pivot {
							value ^= GetVar(b, d)
						}
						c.setVar(d, value)
					}
				}
				group[pivot] = c
			}
			groups = append(groups, group)
			return
		}
		for i := k; i < len(order); i++ {
			order[k], order[i] = order[i], order[k]
			visit(k + 1)
			order[k], order[i] = order[i], order[k]
		}
	}
	visit(0)
	return groups
}
func groupCost(a []*Cube) int {
	n := 0
	for _, c := range a {
		n += literalCount(c)
	}
	return n
}
func ExorLinkCubeIteratorStart(m *Manager, a, b *Cube, d CubeDist) []*Cube {
	m.linkGroups = ExorLink(a, b, int(d)+2)
	m.linkIndex = 0
	sort.SliceStable(m.linkGroups, func(i, j int) bool { return groupCost(m.linkGroups[i]) < groupCost(m.linkGroups[j]) })
	if len(m.linkGroups) == 0 {
		return nil
	}
	return m.linkGroups[0]
}
func ExorLinkCubeIteratorNext(m *Manager) []*Cube {
	m.linkIndex++
	if m.linkIndex >= len(m.linkGroups) {
		return nil
	}
	return m.linkGroups[m.linkIndex]
}
func ExorLinkCubeIteratorPick(m *Manager, g int) []*Cube {
	if g < 0 || g >= len(m.linkGroups) {
		return nil
	}
	m.linkIndex = g
	return m.linkGroups[g]
}
func ExorLinkCubeIteratorCleanUp(m *Manager, takeLast bool) []*Cube {
	var kept []*Cube
	if takeLast && m.linkIndex < len(m.linkGroups) {
		kept = m.linkGroups[m.linkIndex]
	}
	m.linkGroups = nil
	m.linkIndex = 0
	return kept
}
