// Port of ABC/MVSIS FXU (MVSIS Group, UC Berkeley, 2003).
// Original algorithm and licensing: src/opt/fxu; see repository copyright notice.
package fxu

import "sort"

func Fxu_MatrixAllocate() *Fxu_Matrix {
	return &Fxu_Matrix{covers: make(map[int32][]*Fxu_Cube), phases: make(map[int32]bool), changed: make(map[int32]bool), singleCatalog: make(map[[2]int32]*Fxu_Single), doubleCatalog: make(map[string]*Fxu_Double), singleHeap: Fxu_HeapSingleStart(), doubleHeap: Fxu_HeapDoubleStart()}
}
func Fxu_MatrixDelete(p *Fxu_Matrix) {
	if p != nil {
		*p = Fxu_Matrix{}
	}
}
func Fxu_MatrixAddVar(p *Fxu_Matrix) *Fxu_Var {
	v := &Fxu_Var{Index: int32(len(p.Vars))}
	p.Vars = append(p.Vars, v)
	return v
}
func Fxu_MatrixAddCube(p *Fxu_Matrix, v *Fxu_Var, index int32) *Fxu_Cube {
	c := &Fxu_Cube{Index: index, Owner: v.Index / 2}
	p.Cubes = append(p.Cubes, c)
	p.covers[c.Owner] = append(p.covers[c.Owner], c)
	return c
}
func Fxu_MatrixAddLiteral(p *Fxu_Matrix, c *Fxu_Cube, v *Fxu_Var) {
	at := sort.Search(len(c.Literals), func(i int) bool { return c.Literals[i].Var.Index >= v.Index })
	if at < len(c.Literals) && c.Literals[at].Var == v {
		panic("duplicate FXU literal")
	}
	l := &Fxu_Lit{Cube: c, Var: v}
	c.Literals = append(c.Literals, nil)
	copy(c.Literals[at+1:], c.Literals[at:])
	c.Literals[at] = l
	v.Literals = append(v.Literals, l)
}
func Fxu_MatrixDelLiteral(p *Fxu_Matrix, l *Fxu_Lit) {
	for i, x := range l.Cube.Literals {
		if x == l {
			copy(l.Cube.Literals[i:], l.Cube.Literals[i+1:])
			l.Cube.Literals[len(l.Cube.Literals)-1] = nil
			l.Cube.Literals = l.Cube.Literals[:len(l.Cube.Literals)-1]
			break
		}
	}
	for i, x := range l.Var.Literals {
		if x == l {
			copy(l.Var.Literals[i:], l.Var.Literals[i+1:])
			l.Var.Literals[len(l.Var.Literals)-1] = nil
			l.Var.Literals = l.Var.Literals[:len(l.Var.Literals)-1]
			break
		}
	}
}
func setCubeColumns(p *Fxu_Matrix, c *Fxu_Cube, columns []int32) {
	for len(c.Literals) > 0 {
		Fxu_MatrixDelLiteral(p, c.Literals[len(c.Literals)-1])
	}
	for _, column := range columns {
		Fxu_MatrixAddLiteral(p, c, p.Vars[column])
	}
}
func deleteCube(p *Fxu_Matrix, c *Fxu_Cube) {
	setCubeColumns(p, c, nil)
	c.deleted = true
	p.changed[c.Owner] = true
}
func Fxu_MatrixDelDivisor(p *Fxu_Matrix, d *Fxu_Double) {
	Fxu_HeapDoubleDelete(p.doubleHeap, d)
	delete(p.doubleCatalog, pairKey(d.First, d.Second))
	for i, x := range p.Doubles {
		if x == d {
			p.Doubles = append(p.Doubles[:i], p.Doubles[i+1:]...)
			break
		}
	}
}
func Fxu_MatrixAddSingle(p *Fxu_Matrix, a, b *Fxu_Var, weight int32) *Fxu_Single {
	if a.Index > b.Index {
		a, b = b, a
	}
	key := [2]int32{a.Index, b.Index}
	if old := p.singleCatalog[key]; old != nil {
		old.Weight = weight
		Fxu_HeapSingleUpdate(p.singleHeap, old)
		return old
	}
	s := &Fxu_Single{Num: p.singleSerial, Weight: weight, Var1: a, Var2: b, heapIndex: -1}
	p.singleSerial++
	p.singleCatalog[key] = s
	p.Singles = append(p.Singles, s)
	Fxu_HeapSingleInsert(p.singleHeap, s)
	return s
}
func Fxu_MatrixAddDivisor(p *Fxu_Matrix, a, b *Fxu_Cube) *Fxu_Double {
	first, second, base := pairParts(a.columns(), b.columns())
	if len(first) == 0 || len(second) == 0 {
		return nil
	}
	key := pairKey(first, second)
	d := p.doubleCatalog[key]
	if d == nil {
		d = &Fxu_Double{Num: p.doubleSerial, Weight: -int32(len(first) + len(second)), First: first, Second: second, heapIndex: -1}
		p.doubleSerial++
		p.doubleCatalog[key] = d
		p.Doubles = append(p.Doubles, d)
	}
	pair := &Fxu_Pair{Cube1: a, Cube2: b, First: first, Second: second, Base: base, Divisor: d}
	d.Pairs = append(d.Pairs, pair)
	d.Weight += int32(len(first) + len(second) + len(base) - 1)
	if d.heapIndex < 0 {
		Fxu_HeapDoubleInsert(p.doubleHeap, d)
	} else {
		Fxu_HeapDoubleUpdate(p.doubleHeap, d)
	}
	return d
}
