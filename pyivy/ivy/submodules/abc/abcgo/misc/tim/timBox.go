package tim

func Tim_ManCreateBox(p *Tim_Man_t, firstIn, nIns, firstOut, nOuts, table, black int) {
	if firstIn < 0 || nIns < 0 || firstIn+nIns > p.nCos || firstOut < 0 || nOuts < 0 || firstOut+nOuts > p.nCis {
		panic("tim: box terminal range out of bounds")
	}
	b := &Tim_Box_t{iBox: len(p.vBoxes), nInputs: nIns, nOutputs: nOuts, iDelayTable: table, fBlack: black != 0, Inouts: make([]int, nIns+nOuts)}
	p.vBoxes = append(p.vBoxes, b)
	for i := 0; i < nIns; i++ {
		b.Inouts[i] = firstIn + i
		o := &p.pCos[firstIn+i]
		o.iObj2Box = b.iBox
		o.iObj2Num = i
	}
	for i := 0; i < nOuts; i++ {
		b.Inouts[nIns+i] = firstOut + i
		o := &p.pCis[firstOut+i]
		o.iObj2Box = b.iBox
		o.iObj2Num = i
	}
}
func Tim_ManBoxForCi(p *Tim_Man_t, i int) int {
	if i < 0 || i >= p.nCis {
		return -1
	}
	return p.pCis[i].iObj2Box
}
func Tim_ManBoxForCo(p *Tim_Man_t, i int) int {
	if i < 0 || i >= p.nCos {
		return -1
	}
	return p.pCos[i].iObj2Box
}
func Tim_ManBoxInputFirst(p *Tim_Man_t, i int) int { return p.vBoxes[i].Inouts[0] }
func Tim_ManBoxInputLast(p *Tim_Man_t, i int) int {
	return Tim_ManBoxInputFirst(p, i) + p.vBoxes[i].nInputs - 1
}
func Tim_ManBoxOutputFirst(p *Tim_Man_t, i int) int { b := p.vBoxes[i]; return b.Inouts[b.nInputs] }
func Tim_ManBoxOutputLast(p *Tim_Man_t, i int) int {
	return Tim_ManBoxOutputFirst(p, i) + p.vBoxes[i].nOutputs - 1
}
func Tim_ManBoxInputNum(p *Tim_Man_t, i int) int     { return p.vBoxes[i].nInputs }
func Tim_ManBoxOutputNum(p *Tim_Man_t, i int) int    { return p.vBoxes[i].nOutputs }
func Tim_ManBoxDelayTableId(p *Tim_Man_t, i int) int { return p.vBoxes[i].iDelayTable }
func Tim_ManBoxDelayTable(p *Tim_Man_t, i int) []float32 {
	b := p.vBoxes[i]
	if b.iDelayTable < 0 {
		return nil
	}
	t := p.vDelayTables[b.iDelayTable]
	if len(t) != 3+b.nInputs*b.nOutputs || int(t[1]) != b.nInputs || int(t[2]) != b.nOutputs {
		panic("tim: incompatible delay table")
	}
	return t
}
func Tim_ManBoxIsBlack(p *Tim_Man_t, i int) bool  { return p.vBoxes[i].fBlack }
func Tim_ManBoxCopy(p *Tim_Man_t, i int) int      { return p.vBoxes[i].iCopy }
func Tim_ManBoxSetCopy(p *Tim_Man_t, i, copy int) { p.vBoxes[i].iCopy = copy }
func Tim_ManBoxFindFromCiNum(p *Tim_Man_t, n int) int {
	if n < 0 || n >= p.nCis {
		panic("tim: invalid CI")
	}
	pi := p.nCis
	for _, b := range p.vBoxes {
		pi -= b.nOutputs
	}
	if n < pi {
		return -1
	}
	for i := range p.vBoxes {
		if n < Tim_ManBoxOutputFirst(p, i) {
			return i - 1
		}
	}
	return -2
}
