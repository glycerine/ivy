// Package tim manages white-box connectivity and incremental timing.
package tim

// Tim_Man_t replaces C flexible storage and pointer vectors with typed slices.
type Tim_Man_t struct {
	vBoxes       []*Tim_Box_t
	vDelayTables [][]float32
	nTravIds     int
	fUseTravId   bool
	nCis, nCos   int
	pCis, pCos   []Tim_Obj_t
}
type Tim_Box_t struct {
	iBox, TravId, nInputs, nOutputs, iDelayTable, iCopy int
	fBlack                                              bool
	Inouts                                              []int
}
type Tim_Obj_t struct {
	Id, TravId, iObj2Box, iObj2Num int
	timeArr, timeReq               float32
}

func Tim_ManCi(p *Tim_Man_t, i int) *Tim_Obj_t  { return &p.pCis[i] }
func Tim_ManCo(p *Tim_Man_t, i int) *Tim_Obj_t  { return &p.pCos[i] }
func Tim_ManBox(p *Tim_Man_t, i int) *Tim_Box_t { return p.vBoxes[i] }
func Tim_ManCiBox(p *Tim_Man_t, i int) *Tim_Box_t {
	if b := p.pCis[i].iObj2Box; b >= 0 {
		return p.vBoxes[b]
	}
	return nil
}
func Tim_ManCoBox(p *Tim_Man_t, i int) *Tim_Box_t {
	if b := p.pCos[i].iObj2Box; b >= 0 {
		return p.vBoxes[b]
	}
	return nil
}
func Tim_ManBoxInput(p *Tim_Man_t, b *Tim_Box_t, i int) *Tim_Obj_t {
	if i < 0 || i >= b.nInputs {
		panic("tim: invalid box input")
	}
	return &p.pCos[b.Inouts[i]]
}
func Tim_ManBoxOutput(p *Tim_Man_t, b *Tim_Box_t, i int) *Tim_Obj_t {
	if i < 0 || i >= b.nOutputs {
		panic("tim: invalid box output")
	}
	return &p.pCis[b.Inouts[b.nInputs+i]]
}
