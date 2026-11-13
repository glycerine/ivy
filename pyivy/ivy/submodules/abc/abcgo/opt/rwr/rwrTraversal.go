package rwr

// Rwr_ManIncTravId follows the source counter/mark reset in rwrLib.c.
func Rwr_ManIncTravId(p *Rwr_Man_t) {
	old := p.nTravIds
	p.nTravIds++
	if uint32(old) < 0x8fffffff {
		return
	}
	for _, node := range p.vForest {
		node.TravId = 0
	}
	p.nTravIds = 1
}
