package sbd

import (
	"fmt"
	"math/bits"
	"slices"

	"github.com/glycerine/abcgo/misc/vec"
)

func Sbd_TransposeMatrix64(a []uint64) {
	mask := uint64(0xffffffff)
	for shift := uint(32); shift != 0; {
		for k := 0; k < 64; k = (k + int(shift) + 1) &^ int(shift) {
			temp := (a[k] ^ (a[k+int(shift)] >> shift)) & mask
			a[k] ^= temp
			a[k+int(shift)] ^= temp << shift
		}
		shift >>= 1
		mask ^= mask << shift
	}
}
func Sbd_PrintMatrix64(matrix []uint64) {
	for _, row := range matrix[:64] {
		for bit := 0; bit < 64; bit++ {
			fmt.Printf("%d", row>>bit&1)
		}
		fmt.Println()
	}
	fmt.Println()
}
func Sbd_ManMatrPrint(p *Sbd_Man_t, cover []uint64, columns, rows int32) {
	for col := int32(0); col <= columns; col++ {
		fmt.Printf("%2d : ", col)
		for row := int32(0); row < rows; row++ {
			fmt.Printf("%d", cover[col]>>row&1)
		}
		fmt.Println()
	}
}
func Sbd_ManCoverReverseOrder(cover []uint64) { slices.Reverse(cover[:64]) }
func Sbd_ManAddCube1(limit int32, cover []uint64, rows int32, cube uint64) int32 {
	limit = min(limit, int32(len(cover)))
	if rows > limit {
		panic("cover rows exceed storage")
	}
	for _, row := range cover[:rows] {
		if row&cube == row {
			return rows
		}
	}
	at := int32(0)
	for _, row := range cover[:rows] {
		if row&cube != cube {
			cover[at] = row
			at++
		}
	}
	if at < limit {
		cover[at] = cube
		at++
	}
	if at < rows {
		clear(cover[at:rows])
	}
	return at
}
func Sbd_ManAddCube2(cover *[2][64]uint64, rows int32, cube [2]uint64) int32 {
	if rows > 64 {
		panic("cube rows exceed storage")
	}
	for row := int32(0); row < rows; row++ {
		if cover[0][row]&cube[0] == cover[0][row] && cover[1][row]&cube[1] == cover[1][row] {
			return rows
		}
	}
	at := int32(0)
	for row := int32(0); row < rows; row++ {
		if cover[0][row]&cube[0] != cube[0] || cover[1][row]&cube[1] != cube[1] {
			cover[0][at], cover[1][at] = cover[0][row], cover[1][row]
			at++
		}
	}
	if at < 64 {
		cover[0][at], cover[1][at] = cube[0], cube[1]
		at++
	}
	for row := at; row < rows; row++ {
		cover[0][row], cover[1][row] = 0, 0
	}
	return at
}
func sbdCandidateSearch(p *Sbd_Man_t, cover []uint64, count int32, heuristic bool) int32 {
	vec.Vec_IntClear(p.vDivSet)
	order := make([]int32, count)
	for i := range order {
		order[i] = int32(i)
	}
	target := cover[count]
	search := func(size int32, limits []int32) bool {
		var found []int32
		var visit func(int32, int32, uint64) bool
		visit = func(depth, start int32, covered uint64) bool {
			if depth == size {
				if covered == target {
					found = slices.Clone(found)
					return true
				}
				return false
			}
			for position := start; position < limits[depth]; position++ {
				index := order[position]
				found = append(found, index)
				if visit(depth+1, position+1, covered|cover[index]) {
					return true
				}
				found = found[:len(found)-1]
			}
			return false
		}
		if !visit(0, 0, 0) {
			return false
		}
		for _, index := range found {
			vec.Vec_IntPush(p.vDivSet, index)
		}
		return true
	}
	maximum := min(int32(4), p.pPars.LUTSize)
	if maximum < 1 {
		panic("invalid cover LUT size")
	}
	for size := int32(1); size <= min(int32(2), maximum); size++ {
		if search(size, []int32{count, count}) {
			return 1
		}
	}
	if heuristic {
		for i := range order {
			best := i
			for j := i + 1; j < len(order); j++ {
				if bits.OnesCount64(cover[order[best]]) < bits.OnesCount64(cover[order[j]]) {
					best = j
				}
			}
			order[i], order[best] = order[best], order[i]
		}
	}
	limits := []int32{count, count, count, count}
	if heuristic {
		limits = []int32{count/4 + 1, count/3 + 2, count/2 + 3, count}
	}
	for size := int32(3); size <= maximum; size++ {
		if search(size, limits) {
			return 1
		}
	}
	return 0
}
func Sbd_ManFindCandsSimple(p *Sbd_Man_t, cover []uint64, count int32) int32 {
	return sbdCandidateSearch(p, cover, count, false)
}
func Sbd_ManFindCands(p *Sbd_Man_t, cover []uint64, count int32) int32 {
	return sbdCandidateSearch(p, cover, count, count >= 8 && !p.pPars.Cover)
}
func sbdRefineCover(p *Sbd_Man_t, pivot int32, columns []uint64, rows int32, truth *uint64) int32 {
	divs := int32(vec.Vec_IntSize(p.vDivVars))
	free := int32(vec.Vec_IntSize(p.vWinObjs) + vec.Vec_IntSize(p.vTfo) + vec.Vec_IntSize(p.vRoots))
	pivotVar := vec.Vec_IntEntry(p.vObj2Var, int(pivot))
	for iteration := int32(0); iteration < 32 && rows < 64; iteration++ {
		if Sbd_ManFindCands(p, columns, divs) == 0 {
			return 0
		}
		for i, index := range vec.Vec_IntArray(p.vDivSet) {
			vec.Vec_IntWriteEntry(p.vDivSet, i, vec.Vec_IntEntry(p.vDivVars, int(index)))
		}
		*truth = Sbd_ManSolve(p.pSat, pivotVar, free+iteration, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits)
		switch *truth {
		case SBD_SAT_UNDEC:
			return 0
		case SBD_SAT_SAT:
			for i, value := range vec.Vec_IntArray(p.vDivValues) {
				if value == 14 || value == 13 {
					columns[i] |= uint64(1) << rows
				}
			}
			columns[divs] |= uint64(1) << rows
			rows++
		default:
			return 1
		}
	}
	return 0
}
func Sbd_ManExplore(p *Sbd_Man_t, pivot int32, truth *uint64) int32 {
	divs := int32(vec.Vec_IntSize(p.vDivVars))
	var cubes [2][2][64]uint64
	var counts [2]int32
	// Source row/column reversal plus transpose visits simulation bit63 first.
	for pattern := int32(63); pattern >= 0; pattern-- {
		var values uint64
		var care [2]uint64
		for i, variable := range vec.Vec_IntArray(p.vDivVars) {
			id := vec.Vec_IntEntry(p.vWinObjs, int(variable))
			values |= (Sbd_ObjSim0(p, id)[0] >> pattern & 1) << uint(i)
			care[0] |= (Sbd_ObjSim2(p, id)[0] >> pattern & 1) << uint(i)
			care[1] |= (Sbd_ObjSim3(p, id)[0] >> pattern & 1) << uint(i)
		}
		values |= (Sbd_ObjSim0(p, pivot)[0] >> pattern & 1) << divs
		care[0] |= (Sbd_ObjSim2(p, pivot)[0] >> pattern & 1) << divs
		care[1] |= (Sbd_ObjSim3(p, pivot)[0] >> pattern & 1) << divs
		if care[0]>>divs&1 == 0 {
			continue
		}
		index := values >> divs & 1
		for which := 0; which < 2; which++ {
			if which == 1 && care[0] == care[1] {
				continue
			}
			cube := [2]uint64{^values & care[which], values & care[which]}
			counts[index] = Sbd_ManAddCube2(&cubes[index], counts[index], cube)
		}
	}
	var cover [64]uint64
	rows := int32(0)
	for i := int32(0); i < counts[0] && rows < 32; i++ {
		for j := int32(0); j < counts[1] && rows < 32; j++ {
			cube := cubes[0][1][i]&cubes[1][0][j] | cubes[0][0][i]&cubes[1][1][j]
			if cube == 0 {
				panic("opposite care cubes have no distinguishing variable")
			}
			rows = Sbd_ManAddCube1(64, cover[:], rows, cube)
		}
	}
	var columns [64]uint64
	for row, cube := range cover[:rows] {
		for col := int32(0); col <= divs; col++ {
			columns[col] |= (cube >> col & 1) << uint(row)
		}
	}
	return sbdRefineCover(p, pivot, columns[:], rows, truth)
}
func Sbd_ManExplore2(p *Sbd_Man_t, pivot int32, truth *uint64) int32 {
	p.pSat = Sbd_ManSatSolver(p.pSat, p.pGia, p.vMirrors, pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, 0)
	if p.pSat == nil {
		*truth = 0
		vec.Vec_IntClear(p.vDivSet)
		return 1
	}
	p.pSat.SetLimits(int64(p.pPars.ConflictLimit), 0)
	var onset, offset [64]uint64
	status := Sbd_ManCollectConstantsNew(p.pSat, p.vDivVars, 4, vec.Vec_IntEntry(p.vObj2Var, int(pivot)), onset[:], offset[:])
	if status >= 0 {
		if status == 0 {
			*truth = 0
		} else {
			*truth = ^uint64(0)
		}
		vec.Vec_IntClear(p.vDivSet)
		p.nLuts[0]++
		return 1
	}
	if status == -2 {
		return 0
	}
	var rowsTable [64]uint64
	rows := int32(0)
	for _, a := range onset[:4] {
		for _, b := range offset[:4] {
			rows = Sbd_ManAddCube1(64, rowsTable[:], rows, a^b)
		}
	}
	var columns [64]uint64
	for row, cube := range rowsTable[:rows] {
		for col := 0; col <= vec.Vec_IntSize(p.vDivVars); col++ {
			columns[col] |= (cube >> uint(col) & 1) << uint(row)
		}
	}
	result := sbdRefineCover(p, pivot, columns[:], rows, truth)
	if result != 0 {
		p.nLuts[1]++
	}
	return result
}
func sbdSolveSupport(p *Sbd_Man_t, pivot, free int32) uint64 {
	if vec.Vec_IntSize(p.vDivSet) <= 6 {
		return Sbd_ManSolve(p.pSat, pivot, free, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits)
	}
	sop := sbdVector()
	values := vec.Vec_IntStart(vec.Vec_IntSize(p.vDivSet))
	if Sbd_ManSolve2(p.pSat, pivot, free, p.vDivSet, values, p.vLits, sop) != 0 {
		return 0
	}
	return SBD_SAT_SAT
}
func Sbd_ManExploreCut(p *Sbd_Man_t, pivot, count int32, leaves []int32, structureCount *int32, structures []Sbd_Str_t, free *int32) int32 {
	pivotVar := vec.Vec_IntEntry(p.vObj2Var, int(pivot))
	delay := vec.Vec_IntEntry(p.vLutLevs, int(pivot))
	size := p.pPars.LUTSize
	for attempt := int32(0); attempt < count; attempt++ {
		remove := count - 1 - attempt
		vec.Vec_IntClear(p.vDivSet)
		for i, leaf := range leaves[:count] {
			if int32(i) != remove && leaf != -1 {
				vec.Vec_IntPush(p.vDivSet, vec.Vec_IntEntry(p.vObj2Var, int(leaf)))
			}
		}
		result := sbdSolveSupport(p, pivotVar, *free)
		*free++
		if result == SBD_SAT_UNDEC {
			continue
		}
		if result == SBD_SAT_SAT {
			if vec.Vec_IntEntry(p.vLutLevs, int(leaves[remove]))-delay > -2 {
				return 0
			}
		} else {
			leaves[remove] = -1
		}
	}
	vec.Vec_IntClear(p.vDivSet)
	for _, leaf := range leaves[:count] {
		if leaf != -1 {
			vec.Vec_IntPush(p.vDivSet, leaf)
		}
	}
	if vec.Vec_IntSize(p.vDivSet) <= int(size) {
		for i, id := range vec.Vec_IntArray(p.vDivSet) {
			vec.Vec_IntWriteEntry(p.vDivSet, i, vec.Vec_IntEntry(p.vObj2Var, int(id)))
		}
		truth := Sbd_ManSolve(p.pSat, pivotVar, *free, p.vDivSet, p.vDivVars, p.vDivValues, p.vLits)
		*free++
		if truth == SBD_SAT_SAT || truth == SBD_SAT_UNDEC {
			return 0
		}
		inputs := make([]int32, vec.Vec_IntSize(p.vDivSet))
		for i := range inputs {
			inputs[i] = int32(i)
		}
		structures[0] = NewLUTStructure(inputs)
		structures[0].Res = truth
		*structureCount = 1
		p.nLuts[1]++
		return 1
	}
	var top, bottom, bottom1, bottom2 []int32
	for i, id := range vec.Vec_IntArray(p.vDivSet) {
		difference := vec.Vec_IntEntry(p.vLutLevs, int(id)) - delay
		if difference > -2 {
			return 0
		}
		if difference == -2 {
			top = append(top, int32(i))
		} else {
			bottom = append(bottom, int32(i))
			if difference == -3 {
				bottom1 = append(bottom1, int32(i))
			} else {
				bottom2 = append(bottom2, int32(i))
			}
		}
		vec.Vec_IntWriteEntry(p.vDivSet, i, vec.Vec_IntEntry(p.vObj2Var, int(id)))
	}
	if len(top) > int(size-1) {
		return 0
	}
	divs := int32(vec.Vec_IntSize(p.vDivSet))
	padded := func(fixed []int32, first int32) Sbd_Str_t {
		inputs := slices.Clone(fixed)
		for len(inputs) < int(size) {
			inputs = append(inputs, first+int32(len(inputs)-len(fixed)))
		}
		return NewLUTStructure(inputs)
	}
	solve := func(items []Sbd_Str_t, stat int) bool {
		copy(structures, items)
		*structureCount = int32(len(items))
		if Sbd_ProblemSolve(p.pGia, p.vMirrors, pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, p.vDivSet, *structureCount, structures) != 0 {
			p.nLuts[stat]++
			return true
		}
		return false
	}
	if divs <= 2*size-1 {
		t, b := slices.Clone(top), slices.Clone(bottom)
		for len(b) > int(size) {
			t = append(t, b[len(b)-1])
			b = b[:len(b)-1]
		}
		items := []Sbd_Str_t{padded(t, divs+1), NewLUTStructure(b)}
		for i := 0; i < int(size-1)-len(t); i++ {
			items = append(items, NewSelectorStructure(b))
		}
		if solve(items, 2) {
			return 1
		}
	}
	if p.pPars.LUTCount < 3 || divs < 2*size-1 {
		return 0
	}
	if len(top) <= int(size-2) {
		t, b := slices.Clone(top), slices.Clone(bottom)
		for len(b) > int(2*size) {
			t = append(t, b[len(b)-1])
			b = b[:len(b)-1]
		}
		if len(b) > int(size) {
			items := []Sbd_Str_t{padded(t, divs+1), NewLUTStructure(b[:size]), NewLUTStructure(b[len(b)-int(size):])}
			for i := 0; i < int(size-2)-len(t); i++ {
				items = append(items, NewSelectorStructure(b))
			}
			if solve(items, 3) {
				return 1
			}
		}
	}
	if len(bottom1)+len(top) <= int(2*size-2) {
		t, b1, b2 := slices.Clone(top), slices.Clone(bottom1), slices.Clone(bottom2)
		for len(b2) > int(size) {
			b1 = append(b1, b2[len(b2)-1])
			b2 = b2[:len(b2)-1]
		}
		for len(b1) > int(size-1) {
			t = append(t, b1[len(b1)-1])
			b1 = b1[:len(b1)-1]
		}
		if len(t) > int(size-1) {
			return 0
		}
		diff1, diff2 := int(size-1)-len(t), int(size-1)-len(b1)
		topInputs := append(slices.Clone(t), divs+1)
		for i := 0; i < diff1; i++ {
			topInputs = append(topInputs, divs+3+int32(i))
		}
		midInputs := append(slices.Clone(b1), divs+2)
		for i := 0; i < diff2; i++ {
			midInputs = append(midInputs, divs+3+int32(diff1+i))
		}
		items := []Sbd_Str_t{NewLUTStructure(topInputs), NewLUTStructure(midInputs), NewLUTStructure(b2)}
		for i := 0; i < diff1+diff2; i++ {
			choices := slices.Clone(b2)
			if i < diff1 {
				choices = append(choices, b1...)
			}
			items = append(items, NewSelectorStructure(choices))
		}
		if solve(items, 4) {
			return 1
		}
	}
	return 0
}
func Sbd_ManExplore3(p *Sbd_Man_t, pivot int32, count *int32, structures []Sbd_Str_t) int32 {
	p.pSat = Sbd_ManSatSolver(p.pSat, p.pGia, p.vMirrors, pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, 0)
	if p.pSat == nil {
		return 0
	}
	p.pSat.SetLimits(int64(p.pPars.ConflictLimit), 0)
	free := int32(vec.Vec_IntSize(p.vWinObjs) + vec.Vec_IntSize(p.vTfo) + vec.Vec_IntSize(p.vRoots))
	start := free
	var leaves [10]int32
	if p.pSrv != nil {
		n := Sbd_ManCutServerFirst(p.pSrv, pivot, leaves[:])
		if n < 0 {
			return 0
		}
		return Sbd_ManExploreCut(p, pivot, n, leaves[:], count, structures, &free)
	}
	for size := p.pPars.LUTSize + 1; size <= p.pPars.CutSize; size++ {
		n := Sbd_StoObjBestCut(p.pSto, pivot, size, leaves[:])
		if n >= 0 && Sbd_ManExploreCut(p, pivot, n, leaves[:], count, structures, &free) != 0 {
			return 1
		}
	}
	if free-start > SBD_FVAR_MAX {
		panic("SBD free-variable budget exhausted")
	}
	return 0
}
