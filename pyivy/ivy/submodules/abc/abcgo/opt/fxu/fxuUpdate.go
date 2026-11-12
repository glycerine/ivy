// Port of ABC/MVSIS FXU (MVSIS Group, UC Berkeley, 2003).
// Original algorithm and licensing: src/opt/fxu; see repository copyright notice.
package fxu

import "sort"

func containsColumn(columns []int32, column int32) bool {
	at := sort.Search(len(columns), func(i int) bool { return columns[i] >= column })
	return at < len(columns) && columns[at] == column
}
func Fxu_UpdateCreateNewVars(p *Fxu_Matrix, c, d **Fxu_Var, cubes int32) {
	*c = Fxu_MatrixAddVar(p)
	*d = Fxu_MatrixAddVar(p)
	p.phases[(*d).Index/2] = true
	p.changed[(*d).Index/2] = true
}
func Fxu_UpdateMatrixSingleClean(p *Fxu_Matrix, a, b, newVar *Fxu_Var) {
	for _, cube := range p.Cubes {
		if cube.deleted {
			continue
		}
		columns := cube.columns()
		if !containsColumn(columns, a.Index) || !containsColumn(columns, b.Index) {
			continue
		}
		next := make([]int32, 0, len(columns)-1)
		for _, column := range columns {
			if column != a.Index && column != b.Index {
				next = append(next, column)
			}
		}
		next = append(next, newVar.Index)
		setCubeColumns(p, cube, sortedColumns(next))
		p.changed[cube.Owner] = true
	}
}
func Fxu_UpdateMatrixDoubleClean(p *Fxu_Matrix, use, remove *Fxu_Cube) {
	_, _, base := pairParts(use.columns(), remove.columns())
	setCubeColumns(p, use, base)
	deleteCube(p, remove)
	p.changed[use.Owner] = true
}
func Fxu_UpdatePairCompare(a, b *Fxu_Pair) int32 {
	if a.Cube1.Owner < b.Cube1.Owner {
		return -1
	}
	if a.Cube1.Owner > b.Cube1.Owner {
		return 1
	}
	x, y := Fxu_PairMinCubeInt(a), Fxu_PairMinCubeInt(b)
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}
func Fxu_UpdatePairsSort(p *Fxu_Matrix, d *Fxu_Double) []*Fxu_Pair {
	pairs := append([]*Fxu_Pair(nil), d.Pairs...)
	sort.SliceStable(pairs, func(i, j int) bool { return Fxu_UpdatePairCompare(pairs[i], pairs[j]) < 0 })
	return pairs
}
func Fxu_UpdateDoublePairs(p *Fxu_Matrix, d *Fxu_Double, newVar *Fxu_Var) {
	for _, pair := range Fxu_UpdatePairsSort(p, d) {
		if pair.Cube1.deleted || pair.Cube2.deleted {
			continue
		}
		a, b, _ := pairParts(pair.Cube1.columns(), pair.Cube2.columns())
		if !sameColumns(a, d.First) || !sameColumns(b, d.Second) {
			continue
		}
		use, remove := Fxu_PairMinCube(pair), Fxu_PairMaxCube(pair)
		Fxu_UpdateMatrixDoubleClean(p, use, remove)
		Fxu_MatrixAddLiteral(p, use, newVar)
	}
}
func Fxu_UpdateMatrixDoubleCreateCubes(p *Fxu_Matrix, a, b *Fxu_Cube, d *Fxu_Double) {
	for _, column := range d.First {
		Fxu_MatrixAddLiteral(p, a, p.Vars[column])
	}
	for _, column := range d.Second {
		Fxu_MatrixAddLiteral(p, b, p.Vars[column])
	}
}
func refreshCandidates(p *Fxu_Matrix) {
	rebuildDoubles(p, false)
	rebuildSingles(p, false, p.params.SingleMax)
}
func Fxu_UpdateSingle(p *Fxu_Matrix) {
	s := Fxu_HeapSingleReadMax(p.singleHeap)
	if s == nil {
		panic("no FXU single to extract")
	}
	var c, d *Fxu_Var
	Fxu_UpdateCreateNewVars(p, &c, &d, 1)
	definition := Fxu_MatrixAddCube(p, d, 0)
	Fxu_UpdateMatrixSingleClean(p, s.Var1, s.Var2, d)
	Fxu_MatrixAddLiteral(p, definition, s.Var1)
	Fxu_MatrixAddLiteral(p, definition, s.Var2)
	p.SinglesTaken++
	refreshCandidates(p)
}
func Fxu_UpdateDouble(p *Fxu_Matrix) {
	divisor := Fxu_HeapDoubleReadMax(p.doubleHeap)
	if divisor == nil {
		panic("no FXU double to extract")
	}
	var c, d *Fxu_Var
	Fxu_UpdateCreateNewVars(p, &c, &d, 2)
	first := Fxu_MatrixAddCube(p, d, 0)
	second := Fxu_MatrixAddCube(p, d, 1)
	Fxu_UpdateMatrixDoubleCreateCubes(p, first, second, divisor)
	Fxu_UpdateDoublePairs(p, divisor, d)
	p.DoublesTaken++
	refreshCandidates(p)
}
func Fxu_Update(p *Fxu_Matrix, s *Fxu_Single, d *Fxu_Double) {
	if s == nil {
		Fxu_UpdateDouble(p)
		return
	}
	if d == nil {
		Fxu_UpdateSingle(p)
		return
	}
	var negative, positive *Fxu_Var
	Fxu_UpdateCreateNewVars(p, &negative, &positive, 1)
	definition := Fxu_MatrixAddCube(p, positive, 0)
	Fxu_UpdateMatrixSingleClean(p, s.Var1, s.Var2, positive)
	Fxu_UpdateDoublePairs(p, d, negative)
	Fxu_MatrixAddLiteral(p, definition, s.Var1)
	Fxu_MatrixAddLiteral(p, definition, s.Var2)
	p.ComplementsTaken++
	refreshCandidates(p)
}
