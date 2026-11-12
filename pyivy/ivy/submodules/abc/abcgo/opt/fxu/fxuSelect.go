package fxu

const MAX_SIZE_LOOKAHEAD = 20

func Fxu_MatrixFindComplement(p *Fxu_Matrix, column int32) int32 { return column ^ 1 }
func Fxu_MatrixFindDouble(p *Fxu_Matrix, a, b []int32, na, nb int32) *Fxu_Double {
	first, second := sortedColumns(a[:na]), sortedColumns(b[:nb])
	if len(first) > 0 && len(second) > 0 && first[0] > second[0] {
		first, second = second, first
	}
	return p.doubleCatalog[pairKey(first, second)]
}
func Fxu_MatrixGetDoubleVars(p *Fxu_Matrix, d *Fxu_Double) ([]int32, []int32) {
	return append([]int32(nil), d.First...), append([]int32(nil), d.Second...)
}
func Fxu_MatrixFindComplementSingle(p *Fxu_Matrix, s *Fxu_Single) *Fxu_Double {
	a, b := s.Var1.Index^1, s.Var2.Index^1
	if a > b {
		a, b = b, a
	}
	return p.doubleCatalog[pairKey([]int32{a}, []int32{b})]
}
func Fxu_MatrixFindComplementDouble2(p *Fxu_Matrix, d *Fxu_Double) *Fxu_Single {
	if len(d.First) != 1 || len(d.Second) != 1 {
		return nil
	}
	a, b := d.First[0]^1, d.Second[0]^1
	if a > b {
		a, b = b, a
	}
	return p.singleCatalog[[2]int32{a, b}]
}
func Fxu_MatrixFindComplementDouble4(p *Fxu_Matrix, d *Fxu_Double) *Fxu_Double {
	if len(d.First) != 2 || len(d.Second) != 2 {
		return nil
	}
	if d.First[0]^1 != d.Second[0] || d.First[1]^1 != d.Second[1] {
		return nil
	}
	return Fxu_MatrixFindDouble(p, []int32{d.First[0], d.Second[1]}, []int32{d.Second[0], d.First[1]}, 2, 2)
}
func Fxu_Select(p *Fxu_Matrix, single **Fxu_Single, double **Fxu_Double) int32 {
	var singles []*Fxu_Single
	var doubles []*Fxu_Double
	for i := 0; i < MAX_SIZE_LOOKAHEAD; i++ {
		s := Fxu_HeapSingleGetMax(p.singleHeap)
		if s == nil {
			break
		}
		singles = append(singles, s)
	}
	for _, s := range singles {
		Fxu_HeapSingleInsert(p.singleHeap, s)
	}
	for i := 0; i < MAX_SIZE_LOOKAHEAD; i++ {
		d := Fxu_HeapDoubleGetMax(p.doubleHeap)
		if d == nil {
			break
		}
		doubles = append(doubles, d)
	}
	for _, d := range doubles {
		Fxu_HeapDoubleInsert(p.doubleHeap, d)
	}
	best := int32(-1)
	*single = nil
	*double = nil
	for _, s := range singles {
		d := Fxu_MatrixFindComplementSingle(p, s)
		weight := s.Weight
		if d != nil {
			weight += d.Weight + int32(len(d.First)+len(d.Second))
		}
		if weight > best {
			best = weight
			*single = s
			*double = d
		}
	}
	for _, d := range doubles {
		s := Fxu_MatrixFindComplementDouble2(p, d)
		weight := d.Weight
		if s != nil {
			weight += s.Weight + int32(len(d.First)+len(d.Second))
		}
		if weight > best {
			best = weight
			*single = s
			*double = d
		}
	}
	return best
}
func Fxu_SelectSCD(p *Fxu_Matrix, limit int32, a, b **Fxu_Var) int32 {
	best := int32(-10)
	for key, count := range collectSingleCounts(p) {
		weight := count - 2
		if weight > best {
			best = weight
			*a = p.Vars[key[0]]
			*b = p.Vars[key[1]]
		}
	}
	return best
}
