// Port of ABC/MVSIS FXU (MVSIS Group, UC Berkeley, 2003).
// Original algorithm and licensing: src/opt/fxu; see repository copyright notice.
package fxu

func Fxu_ListMatrixAddVariable(p *Fxu_Matrix, v *Fxu_Var) { p.Vars = append(p.Vars, v) }
func Fxu_ListMatrixDelVariable(p *Fxu_Matrix, v *Fxu_Var) {
	for i, x := range p.Vars {
		if x == v {
			if len(v.Literals) != 0 {
				panic("removing used FXU column")
			}
			p.Vars[i] = nil
			return
		}
	}
}
func Fxu_ListMatrixAddCube(p *Fxu_Matrix, c *Fxu_Cube) {
	p.Cubes = append(p.Cubes, c)
	p.covers[c.Owner] = append(p.covers[c.Owner], c)
}
func Fxu_ListMatrixDelCube(p *Fxu_Matrix, c *Fxu_Cube) { deleteCube(p, c) }
func Fxu_ListMatrixAddSingle(p *Fxu_Matrix, s *Fxu_Single) {
	p.Singles = append(p.Singles, s)
	p.singleCatalog[[2]int32{s.Var1.Index, s.Var2.Index}] = s
}
func Fxu_ListMatrixDelSingle(p *Fxu_Matrix, s *Fxu_Single) {
	delete(p.singleCatalog, [2]int32{s.Var1.Index, s.Var2.Index})
	for i, x := range p.Singles {
		if x == s {
			copy(p.Singles[i:], p.Singles[i+1:])
			p.Singles[len(p.Singles)-1] = nil
			p.Singles = p.Singles[:len(p.Singles)-1]
			return
		}
	}
}
func Fxu_ListTableAddDivisor(p *Fxu_Matrix, d *Fxu_Double) {
	p.Doubles = append(p.Doubles, d)
	p.doubleCatalog[pairKey(d.First, d.Second)] = d
}
func Fxu_ListTableDelDivisor(p *Fxu_Matrix, d *Fxu_Double) {
	delete(p.doubleCatalog, pairKey(d.First, d.Second))
	for i, x := range p.Doubles {
		if x == d {
			copy(p.Doubles[i:], p.Doubles[i+1:])
			p.Doubles[len(p.Doubles)-1] = nil
			p.Doubles = p.Doubles[:len(p.Doubles)-1]
			return
		}
	}
}
func Fxu_ListCubeAddLiteral(c *Fxu_Cube, l *Fxu_Lit) { l.Cube = c; c.Literals = append(c.Literals, l) }
func Fxu_ListCubeDelLiteral(c *Fxu_Cube, l *Fxu_Lit) {
	for i, x := range c.Literals {
		if x == l {
			copy(c.Literals[i:], c.Literals[i+1:])
			c.Literals[len(c.Literals)-1] = nil
			c.Literals = c.Literals[:len(c.Literals)-1]
			return
		}
	}
}
func Fxu_ListVarAddLiteral(v *Fxu_Var, l *Fxu_Lit) { l.Var = v; v.Literals = append(v.Literals, l) }
func Fxu_ListVarDelLiteral(v *Fxu_Var, l *Fxu_Lit) {
	for i, x := range v.Literals {
		if x == l {
			copy(v.Literals[i:], v.Literals[i+1:])
			v.Literals[len(v.Literals)-1] = nil
			v.Literals = v.Literals[:len(v.Literals)-1]
			return
		}
	}
}
func Fxu_ListDoubleAddPairLast(d *Fxu_Double, pair *Fxu_Pair) {
	pair.Divisor = d
	d.Pairs = append(d.Pairs, pair)
}
func Fxu_ListDoubleAddPairFirst(d *Fxu_Double, pair *Fxu_Pair) {
	pair.Divisor = d
	d.Pairs = append([]*Fxu_Pair{pair}, d.Pairs...)
}
func Fxu_ListDoubleAddPairMiddle(d *Fxu_Double, spot, pair *Fxu_Pair) {
	pair.Divisor = d
	for i, x := range d.Pairs {
		if x == spot {
			d.Pairs = append(d.Pairs, nil)
			copy(d.Pairs[i+2:], d.Pairs[i+1:])
			d.Pairs[i+1] = pair
			return
		}
	}
	panic("FXU pair insertion point missing")
}
func Fxu_ListDoubleDelPair(d *Fxu_Double, pair *Fxu_Pair) {
	for i, x := range d.Pairs {
		if x == pair {
			copy(d.Pairs[i:], d.Pairs[i+1:])
			d.Pairs[len(d.Pairs)-1] = nil
			d.Pairs = d.Pairs[:len(d.Pairs)-1]
			pair.Divisor = nil
			return
		}
	}
}
func Fxu_ListDoubleAddPairPlace(d *Fxu_Double, pair, spot *Fxu_Pair) {
	if spot == nil {
		Fxu_ListDoubleAddPairLast(d, pair)
	} else {
		Fxu_ListDoubleAddPairMiddle(d, spot, pair)
	}
}
func Fxu_MatrixRingCubesStart(p *Fxu_Matrix) { p.cubeRing = nil }
func Fxu_MatrixRingVarsStart(p *Fxu_Matrix)  { p.varRing = nil }
func Fxu_MatrixRingCubesAdd(p *Fxu_Matrix, c *Fxu_Cube) {
	for _, x := range p.cubeRing {
		if x == c {
			return
		}
	}
	p.cubeRing = append(p.cubeRing, c)
}
func Fxu_MatrixRingVarsAdd(p *Fxu_Matrix, v *Fxu_Var) {
	for _, x := range p.varRing {
		if x == v {
			return
		}
	}
	p.varRing = append(p.varRing, v)
}
