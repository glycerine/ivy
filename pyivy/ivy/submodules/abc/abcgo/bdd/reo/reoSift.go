package reo

// Sift the variable on the most populated unsifted plane first. Visit the nearer
// end first, then the opposite direction; retain the best position and bound
// temporary cost growth by REO_REORDER_LIMIT, as in the source engine.
func reoReorderSift(p *Manager) {
	sifted := map[int]bool{}
	for len(sifted) < len(p.order) {
		selected, biggest := -1, -1
		for level, v := range p.order {
			if !sifted[v] && p.current.LevelNodes[level] > biggest {
				selected = level
				biggest = p.current.LevelNodes[level]
			}
		}
		variable := p.order[selected]
		sifted[variable] = true
		start := append([]int(nil), p.order...)
		best := append([]int(nil), start...)
		bestCost := p.cost(p.current)
		limit := float64(1 + int(REO_REORDER_LIMIT*bestCost))
		directions := []int{-1, 1}
		if selected >= len(start)/2 {
			directions = []int{1, -1}
		}
		for _, direction := range directions {
			if err := p.applyOrder(start); err != nil {
				panic(err)
			}
			for pos := selected + direction; pos >= 0 && pos < len(start); pos += direction {
				level := pos
				if direction > 0 {
					level = pos - 1
				}
				reoReorderSwapAdjacentVars(p, level, 0)
				cost := p.cost(p.current)
				if cost < bestCost-REO_COST_EPSILON {
					bestCost = cost
					best = append(best[:0], p.order...)
					candidate := float64(1 + int(REO_REORDER_LIMIT*bestCost))
					if candidate < limit {
						limit = candidate
					}
				}
				if cost >= limit {
					break
				}
			}
		}
		if err := p.applyOrder(best); err != nil {
			panic(err)
		}
	}
}
