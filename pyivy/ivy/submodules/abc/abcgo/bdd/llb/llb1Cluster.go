package llb

func Llb_ManComputeCommonQuant(p *Llb_Mtr_t, col1, col2 int32) int32 {
	var weight int32
	for row := int32(0); row < p.nRows-p.nFfs; row++ {
		a, b := p.pMatrix[col1][row], p.pMatrix[col2][row]
		if a && b && p.pRowSums[row] == 2 {
			weight += 2
		} else if a != b {
			weight--
		}
	}
	return weight
}

func Llb_ManComputeBestQuant(p *Llb_Mtr_t) int32 {
	var bestWeight int32 = -100000
	best := int32(-1)
	for i := int32(1); i < p.nCols-1; i++ {
		for k := i + 1; k < p.nCols-1; k++ {
			if p.pColSums[i] == 0 || p.pColSums[i] > p.pMan.pPars.nClusterMax || p.pColSums[k] == 0 || p.pColSums[k] > p.pMan.pPars.nClusterMax {
				continue
			}
			weight := Llb_ManComputeCommonQuant(p, i, k)
			if weight > bestWeight && weight > 0 {
				bestWeight, best = weight, (i<<16)|k
			}
		}
	}
	return best
}

func Llb_ManComputeQuant(p *Llb_Mtr_t) [][]float32 {
	costs := make([][]float32, p.nCols)
	for i := range costs {
		costs[i] = make([]float32, p.nCols)
	}
	for i := int32(1); i < p.nCols-1; i++ {
		for k := i + 1; k < p.nCols-1; k++ {
			v := float32(Llb_ManComputeCommonQuant(p, i, k))
			costs[i][k], costs[k][i] = v, v
		}
	}
	return costs
}

func Llb_ManComputeCommonAttr(p *Llb_Mtr_t, col1, col2 int32) float32 {
	var common, different int32
	for row := int32(0); row < p.nRows-p.nFfs; row++ {
		a, b := p.pMatrix[col1][row], p.pMatrix[col2][row]
		if a && b {
			common++
		} else if a || b {
			different++
		}
	}
	if common+different == 0 {
		return 0
	}
	return -float32(different) / float32(common+different)
}

func Llb_ManComputeBestAttr(p *Llb_Mtr_t) int32 {
	bestWeight := float32(-100000)
	best := int32(-1)
	for i := int32(1); i < p.nCols-1; i++ {
		for k := i + 1; k < p.nCols-1; k++ {
			if p.pColSums[i] == 0 || p.pColSums[i] > p.pMan.pPars.nClusterMax || p.pColSums[k] == 0 || p.pColSums[k] > p.pMan.pPars.nClusterMax {
				continue
			}
			weight := Llb_ManComputeCommonAttr(p, i, k)
			if weight > bestWeight {
				bestWeight, best = weight, (i<<16)|k
			}
		}
	}
	return best
}

func Llb_ManComputeAttr(p *Llb_Mtr_t) [][]float32 {
	costs := make([][]float32, p.nCols)
	for i := range costs {
		costs[i] = make([]float32, p.nCols)
	}
	for i := int32(1); i < p.nCols-1; i++ {
		for k := i + 1; k < p.nCols-1; k++ {
			v := Llb_ManComputeCommonAttr(p, i, k)
			costs[i][k], costs[k][i] = v, v
		}
	}
	return costs
}

func Llb_MtrCombineSelectedColumns(p *Llb_Mtr_t, col1, col2 int32) {
	assert(col1 >= 1 && col1 < p.nCols-1, "first column is interior")
	assert(col2 >= 1 && col2 < p.nCols-1, "second column is interior")
	assert(p.pColGrps[col1] != nil && p.pColGrps[col2] != nil, "both columns have groups")
	for row := int32(0); row < p.nRows; row++ {
		if p.pMatrix[col1][row] && p.pMatrix[col2][row] {
			p.pRowSums[row]--
		}
		if !p.pMatrix[col1][row] && p.pMatrix[col2][row] {
			p.pMatrix[col1][row] = true
			p.pColSums[col1]++
		}
		p.pMatrix[col2][row] = false
	}
	p.pColSums[col2] = 0
}

func Llb_ManClusterOne(p *Llb_Mtr_t, col1, col2 int32) {
	combined := Llb_ManGroupsCombine(p.pColGrps[col1], p.pColGrps[col2])
	Llb_MtrCombineSelectedColumns(p, col1, col2)
	p.pColGrps[col1], p.pColGrps[col2] = combined, nil
}

func Llb_ManClusterCompress(p *Llb_Mtr_t) {
	write := 0
	for read := 0; read < int(p.nCols); read++ {
		if p.pColGrps[read] == nil {
			assert(p.pColSums[read] == 0, "empty group has empty column")
			continue
		}
		if write != read {
			p.pMatrix[write] = p.pMatrix[read]
			p.pColGrps[write] = p.pColGrps[read]
			p.pColSums[write] = p.pColSums[read]
		}
		write++
	}
	p.nCols = int32(write)
	p.pMatrix = p.pMatrix[:write]
	p.pColGrps = p.pColGrps[:write]
	p.pColSums = p.pColSums[:write]
}

func Llb_ManCluster(p *Llb_Mtr_t) {
	for {
		for {
			pair := Llb_ManComputeBestQuant(p)
			if pair <= 0 {
				break
			}
			Llb_ManClusterOne(p, pair>>16, pair&0xffff)
		}
		pair := Llb_ManComputeBestAttr(p)
		if pair <= 0 {
			break
		}
		Llb_ManClusterOne(p, pair>>16, pair&0xffff)
		Llb_MtrVerifyMatrix(p)
	}
	Llb_ManClusterCompress(p)
	Llb_MtrVerifyMatrix(p)
}
