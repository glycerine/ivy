// Incremental wirelength estimate, ported from place_inc.c (Aaron P. Hurst).
package place

func sqHashId(id, maxID int32) int32 {
	if maxID <= 0 {
		panic("invalid hash modulus")
	}
	return int32((int64(id) * (int64(id) + 17)) % int64(maxID))
}
func fastEstimate(cell *ConcreteCell, numNets int32, nets []*ConcreteNet) float32 {
	if cell == nil {
		panic("nil estimated cell")
	}
	var length float32
	for _, net := range nets[:numNets] {
		box := getNetBBox(net)
		if net.m_numTerms == 0 {
			continue
		}
		if cell.m_x < box.x {
			length += box.x - cell.m_x
		}
		if cell.m_x > box.x+box.w {
			length += cell.m_x - box.x - box.w
		}
		if cell.m_y < box.y {
			length += box.y - cell.m_y
		}
		if cell.m_y > box.y+box.h {
			length += cell.m_y - box.y - box.h
		}
	}
	return length
}
func EstimateWirelengthIncrease(cell *ConcreteCell, nets []*ConcreteNet) float32 {
	return fastEstimate(cell, int32(len(nets)), nets)
}
