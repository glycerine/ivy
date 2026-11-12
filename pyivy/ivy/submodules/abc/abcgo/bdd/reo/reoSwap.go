package reo

// Canonical rebuilds replace the C engine's pointer-tag manipulations. Swapping
// adjacent support variables still changes exactly the same variable ordering,
// while the CUDD unique table performs reduction and preserves complemented edges.
func reoReorderSwapAdjacentVars(p *Manager, level, _ int) float64 {
	if level < 0 || level+1 >= len(p.order) {
		panic("reo: invalid adjacent-swap level")
	}
	before := p.cost(p.current)
	order := append([]int(nil), p.order...)
	order[level], order[level+1] = order[level+1], order[level]
	if err := p.applyOrder(order); err != nil {
		panic(err)
	}
	p.stats.Swaps++
	return before - p.cost(p.current)
}
