package llb

import "fmt"

// Llb_MtrVerifyRowsAll checks the cached sum for every partition column.
func Llb_MtrVerifyRowsAll(p *Llb_Mtr_t) {
	for col := int32(0); col < p.nCols; col++ {
		var count int32
		for row := int32(0); row < p.nRows; row++ {
			if p.pMatrix[col][row] {
				count++
			}
		}
		assert(count == p.pColSums[col], "count == p.pColSums[col]")
	}
}

// Llb_MtrVerifyColumnsAll checks the cached sum for every variable row.
func Llb_MtrVerifyColumnsAll(p *Llb_Mtr_t) {
	for row := int32(0); row < p.nRows; row++ {
		var count int32
		for col := int32(0); col < p.nCols; col++ {
			if p.pMatrix[col][row] {
				count++
			}
		}
		assert(count == p.pRowSums[row], "count == p.pRowSums[row]")
	}
}

func Llb_MtrVerifyMatrix(p *Llb_Mtr_t) {
	Llb_MtrVerifyRowsAll(p)
	Llb_MtrVerifyColumnsAll(p)
}

// Llb_MtrFindVarOrder orders variables by the last partition in which they occur.
func Llb_MtrFindVarOrder(p *Llb_Mtr_t) []int32 {
	order := make([]int32, p.nRows)
	last := make([]int32, p.nRows)
	for row := int32(0); row < p.nRows; row++ {
		order[row] = row
		for col := p.nCols - 1; col >= 0; col-- {
			if p.pMatrix[col][row] {
				last[row] = col
				break
			}
		}
	}
	// Stable adjacent ordering preserves the source algorithm's tie behavior.
	for changed := true; changed; {
		changed = false
		for i := 0; i+1 < len(order); i++ {
			if last[i] > last[i+1] {
				last[i], last[i+1] = last[i+1], last[i]
				order[i], order[i+1] = order[i+1], order[i]
				changed = true
			}
		}
	}
	return order
}

func Llb_MtrVarName(p *Llb_Mtr_t, iVar int32) string {
	switch {
	case iVar < p.nPis:
		return "pi"
	case iVar < p.nPis+p.nFfs:
		return "CS"
	case iVar >= p.nRows-p.nFfs:
		return "NS"
	default:
		return "int"
	}
}

func Llb_MtrPrint(p *Llb_Mtr_t, fOrder int32) {
	var order []int32
	if fOrder != 0 {
		order = Llb_MtrFindVarOrder(p)
	}
	for i := int32(0); i < p.nRows; i++ {
		row := i
		if order != nil {
			row = order[i]
		}
		fmt.Printf("%3d : %3d %3s ", row, p.pRowSums[row], Llb_MtrVarName(p, row))
		for col := int32(0); col < p.nCols; col++ {
			if p.pMatrix[col][row] {
				fmt.Print("*")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
}

func Llb_MtrPrintMatrixStats(p *Llb_Mtr_t) {
	first := make([]int32, p.nRows)
	last := make([]int32, p.nRows)
	var span, active, maxCut int32
	for row := int32(0); row < p.nRows; row++ {
		if p.pRowSums[row] == 0 {
			continue
		}
		first[row], last[row] = -1, -1
		for col := int32(0); col < p.nCols; col++ {
			if p.pMatrix[col][row] {
				if first[row] < 0 {
					first[row] = col
				}
				last[row] = col
			}
		}
		assert(first[row] >= 0 && first[row] <= last[row], "first[row] >= 0 && first[row] <= last[row]")
		span += last[row] - first[row]
	}
	for col := int32(0); col < p.nCols; col++ {
		for row := int32(0); row < p.nRows; row++ {
			if first[row] == col {
				active++
			}
		}
		if maxCut < active {
			maxCut = active
		}
		for row := int32(0); row < p.nRows; row++ {
			if last[row] == col {
				active--
			}
		}
	}
	avg := float64(0)
	if p.nRows != 0 {
		avg = float64(span) / float64(p.nRows)
	}
	fmt.Printf("[%4d x %4d]  Life-span =%6.2f  Max-cut =%5d\n", p.nCols, p.nRows, avg, maxCut)
}

func Llb_MtrAlloc(nPis, nFfs, nCols, nRows int32) *Llb_Mtr_t {
	if nPis < 0 || nFfs < 0 || nCols < 0 || nRows < 0 {
		panic("negative matrix dimension")
	}
	p := &Llb_Mtr_t_{
		nPis: nPis, nFfs: nFfs, nCols: nCols, nRows: nRows,
		pColSums: make([]int32, nCols),
		pColGrps: make([]*Llb_Grp_t, nCols),
		pRowSums: make([]int32, nRows),
		pMatrix: make([][]bool, nCols),
		pProdVars: make([]bool, nRows),
		pProdNums: make([]int32, nRows),
	}
	for col := range p.pMatrix {
		p.pMatrix[col] = make([]bool, nRows)
	}
	return p
}

func Llb_MtrFree(p *Llb_Mtr_t) {
	if p == nil {
		return
	}
	p.pProdVars = nil
	p.pProdNums = nil
	p.pMatrix = nil
	p.pRowSums = nil
	p.pColSums = nil
	p.pColGrps = nil
}

func Llb_MtrAddColumn(p *Llb_Mtr_t, group *Llb_Grp_t) {
	col := group.Id
	assert(col >= 0 && col < p.nCols, "col >= 0 && col < p.nCols")
	p.pColGrps[col] = group
	add := func(obj *Aig_Obj_t) {
		row := Vec_IntEntry(group.pMan.vObj2Var, Aig_ObjId(obj))
		assert(row >= 0 && row < p.nRows, "row >= 0 && row < p.nRows")
		if !p.pMatrix[col][row] {
			p.pMatrix[col][row] = true
			p.pColSums[col]++
			p.pRowSums[row]++
		}
	}
	var obj *Aig_Obj_t
	var i int32
	Vec_PtrForEachEntry(Aig_Obj_t, group.vIns, obj, i) { add(obj) }
	Vec_PtrForEachEntry(Aig_Obj_t, group.vOuts, obj, i) { add(obj) }
}

func Llb_MtrRemoveSingletonRows(p *Llb_Mtr_t) {
	for row := int32(0); row < p.nRows; row++ {
		if p.pRowSums[row] >= 2 {
			continue
		}
		for col := int32(0); col < p.nCols; col++ {
			if p.pMatrix[col][row] {
				p.pMatrix[col][row] = false
				p.pColSums[col]--
			}
		}
		p.pRowSums[row] = 0
	}
}

func Llb_MtrCreate(p *Llb_Man_t) *Llb_Mtr_t {
	matrix := Llb_MtrAlloc(Saig_ManPiNum(p.pAig), Saig_ManRegNum(p.pAig), Vec_PtrSize(p.vGroups), Vec_IntSize(p.vVar2Obj))
	var group *Llb_Grp_t
	var i int32
	Vec_PtrForEachEntry(Llb_Grp_t, p.vGroups, group, i) { Llb_MtrAddColumn(matrix, group) }
	return matrix
}
