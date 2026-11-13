package bmc

import "github.com/glycerine/abcgo/misc/vec"

type Div_Type_t int32

const (
	DIV_CST Div_Type_t = iota
	DIV_AND
	DIV_XOR
	DIV_MUX
	DIV_NONE
)

func Bmc_FxDivOr(a, b int32, lits []int32, phase *int32) Div_Type_t {
	if a == b {
		panic("bmc: identical divisor literals")
	}
	if a>>1 == b>>1 {
		return DIV_CST
	}
	if a > b {
		a, b = b, a
	}
	lits[0], lits[1], *phase = a^1, b^1, 1
	return DIV_AND
}
func Bmc_FxDivXor(a, b int32, lits []int32, phase *int32) Div_Type_t {
	if a == b {
		panic("bmc: identical divisor literals")
	}
	*phase ^= (a & 1) ^ (b & 1)
	lits[0], lits[1] = a&^1, b&^1
	return DIV_XOR
}
func Bmc_FxDivMux(c, cn, t, e int32, lits []int32, phase *int32) Div_Type_t {
	if c == cn || c>>1 != cn>>1 || c>>1 == t>>1 || c>>1 == e>>1 {
		panic("bmc: invalid mux divisor")
	}
	if c&1 != 0 {
		c &^= 1
		t, e = e, t
	}
	if t&1 != 0 {
		*phase ^= 1
		t ^= 1
		e ^= 1
	}
	lits[0], lits[1], lits[2] = c, t, e
	return DIV_MUX
}
func Div_FindType(a, b, lits []int32, phase *int32) Div_Type_t {
	*phase = 0
	lits[0], lits[1], lits[2] = TAB_UNUSED, TAB_UNUSED, TAB_UNUSED
	if a[0] == -1 && a[1] == -1 || b[0] == -1 && b[1] == -1 {
		return DIV_CST
	}
	if a[0] < 0 || b[0] < 0 || a[0] == b[0] || a[1] == b[1] && a[1] != -1 {
		panic("bmc: invalid divisor pair")
	}
	if a[1] == -1 && b[1] == -1 {
		return Bmc_FxDivOr(a[0], b[0], lits, phase)
	}
	if a[1] == -1 || b[1] == -1 {
		if a[1] == -1 {
			a[0], b[0] = b[0], a[0]
			a[1], b[1] = b[1], a[1]
		}
		if b[0]>>1 == a[0]>>1 {
			return Bmc_FxDivOr(b[0], a[1], lits, phase)
		}
		if b[0]>>1 == a[1]>>1 {
			return Bmc_FxDivOr(b[0], a[0], lits, phase)
		}
		return DIV_NONE
	}
	if a[0]>>1 == b[0]>>1 && a[1]>>1 == b[1]>>1 {
		return Bmc_FxDivXor(a[0], b[1], lits, phase)
	}
	if a[0]>>1 == b[0]>>1 {
		return Bmc_FxDivMux(a[0], b[0], a[1], b[1], lits, phase)
	}
	if a[0]>>1 == b[1]>>1 {
		return Bmc_FxDivMux(a[0], b[1], a[1], b[0], lits, phase)
	}
	if a[1]>>1 == b[0]>>1 {
		return Bmc_FxDivMux(a[1], b[0], a[0], b[1], lits, phase)
	}
	if a[1]>>1 == b[1]>>1 {
		return Bmc_FxDivMux(a[1], b[1], a[0], b[0], lits, phase)
	}
	return DIV_NONE
}
func Div_AddLit(lit int32, lits []int32) int32 {
	if lits[0] == -1 {
		lits[0] = lit
	} else if lits[1] == -1 {
		lits[1] = lit
	} else {
		return 1
	}
	return 0
}
func Div_FindDiv(a, b *vec.Vec_Int_t, la, lb []int32) int32 {
	x, y := vec.Vec_IntArray(a), vec.Vec_IntArray(b)
	la[0], la[1], lb[0], lb[1] = -1, -1, -1, -1
	i, j, count := 0, 0, int32(0)
	for i < len(x) && j < len(y) {
		if x[i] == y[j] {
			i++
			j++
			count++
		} else if x[i] < y[j] {
			if Div_AddLit(x[i], la) != 0 {
				return -1
			}
			i++
		} else {
			if Div_AddLit(y[j], lb) != 0 {
				return -1
			}
			j++
		}
	}
	for ; i < len(x); i++ {
		if Div_AddLit(x[i], la) != 0 {
			return -1
		}
	}
	for ; j < len(y); j++ {
		if Div_AddLit(y[j], lb) != 0 {
			return -1
		}
	}
	return count
}
