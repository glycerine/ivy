package fpga

import "fmt"

// BooleanBuilder keeps cut truth construction independent of a BDD or AIG package.
// Hop managers implement this interface directly.
type BooleanBuilder interface {
	ConstantOne() any
	Complement(any) any
	Variable(int) any
	And(any, any) any
}

func Fpga_TruthsCutBdd(dd any, c *Fpga_Cut_t) any {
	b, ok := dd.(BooleanBuilder)
	if !ok {
		panic("fpga: Boolean builder required")
	}
	complement := Fpga_CutIsComplement(c) != 0
	c = Fpga_CutRegular(c)
	var build func(uint64, int) any
	build = func(truth uint64, n int) any {
		mask := truthMask(n)
		truth &= mask
		if truth == 0 {
			return b.Complement(b.ConstantOne())
		}
		if truth == mask {
			return b.ConstantOne()
		}
		half := 1 << uint(n-1)
		low := build(truth&truthMask(n-1), n-1)
		high := build(truth>>half, n-1)
		x := b.Variable(n - 1)
		return b.Complement(b.And(b.Complement(b.And(b.Complement(x), low)), b.Complement(b.And(x, high))))
	}
	result := build(c.Truth, len(c.Leaves))
	if complement {
		result = b.Complement(result)
	}
	return result
}
func Fpga_TruthsCutBdd_rec(dd any, c *Fpga_Cut_t, visited *Fpga_NodeVec_t) any {
	return Fpga_TruthsCutBdd(dd, c)
}
func Fpga_CutVolume(c *Fpga_Cut_t) int32 {
	seen := map[*Fpga_Cut_t]bool{}
	var count func(*Fpga_Cut_t) int32
	count = func(c *Fpga_Cut_t) int32 {
		c = Fpga_CutRegular(c)
		if c == nil || seen[c] || c.One == nil {
			return 0
		}
		seen[c] = true
		return 1 + count(c.One) + count(c.Two)
	}
	return count(c)
}
func Fpga_CutVolume_rec(c *Fpga_Cut_t, v *Fpga_NodeVec_t) {
	seen := map[*Fpga_Cut_t]bool{}
	var visit func(*Fpga_Cut_t)
	visit = func(c *Fpga_Cut_t) {
		c = Fpga_CutRegular(c)
		if c == nil || seen[c] || c.One == nil {
			return
		}
		seen[c] = true
		visit(c.One)
		visit(c.Two)
		if c.Root != nil {
			v.Nodes = append(v.Nodes, c.Root)
		}
	}
	visit(c)
}

// Evaluate compares the original AND graph or its selected LUT cover without altering scratch data.
func (p *Fpga_Man_t) Evaluate(inputs []bool, mapped bool) ([]bool, error) {
	if len(inputs) != len(p.Inputs) {
		return nil, fmt.Errorf("fpga: expected %d inputs", len(p.Inputs))
	}
	values := map[*Fpga_Node_t]bool{p.Const1: true}
	for i, n := range p.Inputs {
		values[n] = inputs[i]
	}
	active := map[*Fpga_Node_t]bool{}
	var eval func(*Fpga_Node_t) (bool, error)
	eval = func(edge *Fpga_Node_t) (bool, error) {
		if edge == nil {
			return false, fmt.Errorf("fpga: unset signal")
		}
		if mapped {
			edge = choiceSignal(edge)
		}
		n := Fpga_Regular(edge)
		v, found := values[n]
		if !found {
			if active[n] {
				return false, fmt.Errorf("fpga: cyclic mapping")
			}
			active[n] = true
			if mapped {
				if n.Best == nil {
					return false, fmt.Errorf("fpga: graph has not been mapped")
				}
				index := 0
				for i, leaf := range n.Best.Leaves {
					bit, e := eval(leaf)
					if e != nil {
						return false, e
					}
					if bit {
						index |= 1 << i
					}
				}
				v = n.Best.Truth>>index&1 != 0
			} else {
				a, e := eval(n.One)
				if e != nil {
					return false, e
				}
				b, e := eval(n.Two)
				if e != nil {
					return false, e
				}
				v = a && b
			}
			delete(active, n)
			values[n] = v
		}
		return v != (Fpga_IsComplement(edge) != 0), nil
	}
	out := make([]bool, len(p.Outputs))
	for i, e := range p.Outputs {
		v, err := eval(e)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}
