package exor

import (
	"math/bits"
	"slices"
)

func VarWord(element int) int { return element >> 5 }
func VarBit(element int) int  { return element & 31 }
func GetVar(c *Cube, v int) VarValue {
	if v < 0 || v/16 >= len(c.Inputs) {
		panic("exor: variable outside cube")
	}
	return VarValue(c.Inputs[v/16] >> uint(2*(v%16)) & 3)
}
func ExorVar(c *Cube, v int, value VarValue) {
	if value > 3 {
		panic("exor: invalid XOR value")
	}
	c.Inputs[v/16] ^= uint32(value) << uint(2*(v%16))
}

// GetDistance saturates at five, matching the bounded ExorLink search.
func GetDistance(a, b *Cube) int {
	n := 0
	if len(a.Inputs) != len(b.Inputs) || len(a.Outputs) != len(b.Outputs) {
		panic("exor: mismatched cube dimensions")
	}
	for i, w := range a.Inputs {
		x := w ^ b.Inputs[i]
		n += bits.OnesCount32((x | x>>1) & 0x55555555)
		if n > 4 {
			return 5
		}
	}
	if !slices.Equal(a.Outputs, b.Outputs) {
		n++
	}
	return min(n, 5)
}

// FindDiffVars returns output dimension -1 first, then ascending input indices.
func FindDiffVars(a, b *Cube) []int {
	var out []int
	if !slices.Equal(a.Outputs, b.Outputs) {
		out = append(out, -1)
	}
	for w, x := range a.Inputs {
		diff := x ^ b.Inputs[w]
		for v := 0; v < 16; v++ {
			if diff>>uint(2*v)&3 != 0 {
				out = append(out, 16*w+v)
			}
		}
	}
	return out
}

type Difference struct {
	Variable        int
	Old, New, Other VarValue
}

// GetDistancePlus merges the sole differing input into a, as the source does.
func GetDistancePlus(a, b *Cube) (int, Difference) {
	d := FindDiffVars(a, b)
	info := Difference{Variable: -1}
	if len(d) == 1 && d[0] >= 0 {
		v := d[0]
		info = Difference{v, GetVar(a, v), GetVar(a, v) ^ GetVar(b, v), GetVar(b, v)}
		a.setVar(v, info.New)
	}
	return min(len(d), 5), info
}
func InsertVars(c *Cube, vars []int, values []VarValue) {
	if len(vars) != len(values) {
		panic("exor: literal insertion dimension mismatch")
	}
	for i, v := range vars {
		c.setVar(v, values[i])
	}
}
func InsertVarsWithoutClearing(c *Cube, vars []int, values []VarValue, output int) {
	if len(vars) != len(values) {
		panic("exor: literal insertion dimension mismatch")
	}
	for i, v := range vars {
		c.Inputs[v/16] |= uint32(values[i]) << uint(2*(v%16))
	}
	if output < 0 || output/32 >= len(c.Outputs) {
		panic("exor: output outside cube")
	}
	c.Outputs[output/32] |= 1 << uint(output%32)
}
func literalCount(c *Cube) int {
	n := 0
	for _, w := range c.Inputs {
		for w != 0 {
			v := w & 3
			if v == 1 || v == 2 {
				n++
			}
			w >>= 2
		}
	}
	return n
}
