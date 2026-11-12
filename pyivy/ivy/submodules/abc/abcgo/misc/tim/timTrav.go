package tim

func Tim_ManIncrementTravId(p *Tim_Man_t) {
	if p.nTravIds >= (1<<30)-1 {
		p.nTravIds = 0
		for i := range p.pCis {
			p.pCis[i].TravId = 0
		}
		for i := range p.pCos {
			p.pCos[i].TravId = 0
		}
	}
	p.nTravIds++
}
func Tim_ManSetCurrentTravIdBoxInputs(p *Tim_Man_t, i int) {
	b := p.vBoxes[i]
	for j := 0; j < b.nInputs; j++ {
		Tim_ManBoxInput(p, b, j).TravId = p.nTravIds
	}
}
func Tim_ManSetCurrentTravIdBoxOutputs(p *Tim_Man_t, i int) {
	b := p.vBoxes[i]
	for j := 0; j < b.nOutputs; j++ {
		Tim_ManBoxOutput(p, b, j).TravId = p.nTravIds
	}
}
func Tim_ManSetPreviousTravIdBoxInputs(p *Tim_Man_t, i int) {
	b := p.vBoxes[i]
	for j := 0; j < b.nInputs; j++ {
		Tim_ManBoxInput(p, b, j).TravId = p.nTravIds - 1
	}
}
func Tim_ManSetPreviousTravIdBoxOutputs(p *Tim_Man_t, i int) {
	b := p.vBoxes[i]
	for j := 0; j < b.nOutputs; j++ {
		Tim_ManBoxOutput(p, b, j).TravId = p.nTravIds - 1
	}
}
func Tim_ManIsCiTravIdCurrent(p *Tim_Man_t, i int) bool {
	if !p.fUseTravId {
		panic("tim: traversal IDs disabled")
	}
	return p.pCis[i].TravId == p.nTravIds
}
func Tim_ManIsCoTravIdCurrent(p *Tim_Man_t, i int) bool {
	if !p.fUseTravId {
		panic("tim: traversal IDs disabled")
	}
	return p.pCos[i].TravId == p.nTravIds
}
