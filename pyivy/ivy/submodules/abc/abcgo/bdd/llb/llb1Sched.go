package llb

import "math"

// Llb_MtrSwapColumns swaps all state associated with two partitions.
func Llb_MtrSwapColumns(p *Llb_Mtr_t, iCol1, iCol2 int32) {
	assert(iCol1 >= 0 && iCol1 < p.nCols, "valid first column")
	assert(iCol2 >= 0 && iCol2 < p.nCols, "valid second column")
	if iCol1 == iCol2 {
		return
	}
	p.pColGrps[iCol1], p.pColGrps[iCol2] = p.pColGrps[iCol2], p.pColGrps[iCol1]
	p.pMatrix[iCol1], p.pMatrix[iCol2] = p.pMatrix[iCol2], p.pMatrix[iCol1]
	p.pColSums[iCol1], p.pColSums[iCol2] = p.pColSums[iCol2], p.pColSums[iCol1]
}

func Llb_MtrFindBestColumn(p *Llb_Mtr_t, groupStart int32) int32 {
	weightBest, groupBest := int32(-math.MaxInt32), int32(-1)
	for row := int32(0); row < p.nRows-p.nFfs; row++ {
		if p.pRowSums[row] < 2 || !p.pProdVars[row] || p.pProdNums[row] != 1 {
			continue
		}
		group, occurrences := int32(-1), 0
		for col := groupStart; col < p.nCols-1; col++ {
			if p.pMatrix[col][row] {
				group, occurrences = col, occurrences+1
			}
		}
		assert(occurrences == 1, "quantifiable variable occurs in one remaining group")
		if occurrences != 1 {
			continue
		}
		var weight int32
		for k := int32(0); k < p.nRows; k++ {
			if p.pProdVars[k] && p.pMatrix[group][k] && p.pProdNums[k] == 1 {
				weight += 2
			}
			if !p.pProdVars[k] && p.pMatrix[group][k] {
				weight--
			}
		}
		if weight > 0 && weight > weightBest {
			weightBest, groupBest = weight, group
		}
	}
	if groupBest >= 0 {
		return groupBest
	}

	costBest, tieBest := int32(math.MaxInt32), int32(math.MaxInt32)
	for group := groupStart; group < p.nCols-1; group++ {
		var cost, tie int32
		for row := int32(0); row < p.nRows; row++ {
			if !p.pProdVars[row] && p.pMatrix[group][row] {
				cost++
				tie += p.pProdNums[row]
			}
		}
		if cost < costBest || cost == costBest && tie > tieBest {
			costBest, tieBest, groupBest = cost, tie, group
		}
	}
	return groupBest
}

func Llb_MtrUseSelectedColumn(p *Llb_Mtr_t, col int32) {
	assert(col >= 1 && col < p.nCols-1, "interior column")
	for row := int32(0); row < p.nRows; row++ {
		if !p.pMatrix[col][row] {
			continue
		}
		if p.pProdVars[row] && p.pProdNums[row] == 1 {
			p.pProdVars[row], p.pProdNums[row] = false, 0
			continue
		}
		if !p.pProdVars[row] {
			p.pProdVars[row], p.pProdNums[row] = true, p.pRowSums[row]
		}
		p.pProdNums[row]--
		assert(p.pProdNums[row] >= 0, "remaining partition count is nonnegative")
	}
}

func Llb_MtrVerifyColumns(p *Llb_Mtr_t, groupStart int32) {
	for row := int32(0); row < p.nRows; row++ {
		if !p.pProdVars[row] {
			continue
		}
		var count int32
		for col := groupStart; col < p.nCols; col++ {
			if p.pMatrix[col][row] {
				count++
			}
		}
		assert(count == p.pProdNums[row], "remaining count matches matrix")
	}
}

func Llb_MtrSchedule(p *Llb_Mtr_t) {
	for row := int32(0); row < p.nRows; row++ {
		isState := row >= p.nPis && row < p.nPis+p.nFfs
		if isState {
			p.pProdVars[row] = true
			p.pProdNums[row] = p.pRowSums[row] - 1
		} else {
			p.pProdVars[row] = false
			p.pProdNums[row] = p.pRowSums[row]
		}
	}
	Llb_MtrVerifyMatrix(p)
	for group := int32(1); group < p.nCols-1; group++ {
		Llb_MtrVerifyColumns(p, group)
		best := Llb_MtrFindBestColumn(p, group)
		assert(best >= group && best < p.nCols-1, "scheduler found a remaining partition")
		if best < group || best >= p.nCols-1 {
			return
		}
		Llb_MtrUseSelectedColumn(p, best)
		Llb_MtrSwapColumns(p, group, best)
	}
	Llb_MtrVerifyMatrix(p)
}
