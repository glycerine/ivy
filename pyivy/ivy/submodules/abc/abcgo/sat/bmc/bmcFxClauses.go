package bmc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Bmc_FxAddClauses constrains fresh variables to the selected divisors.
func Bmc_FxAddClauses(s *bsat.Solver, divs *vec.Vec_Int_t, ciStart, varStart int32) {
	entries := vec.Vec_IntArray(divs)
	if len(entries)%4 != 0 {
		panic("bmc: invalid divisor list")
	}
	n := int32(len(entries) / 4)
	for i := int32(0); i < n; i++ {
		kind := Div_Type_t(entries[4*i])
		raw := entries[4*i+1 : 4*i+4]
		a := bsat.Positive(ciStart + (raw[0] >> 1))
		b := bsat.Positive(ciStart + (raw[1] >> 1))
		z := bsat.Positive(varStart + n - 1 - i)
		var clauses [][]bsat.Literal
		switch kind {
		case DIV_AND:
			a ^= raw[0] & 1
			b ^= raw[1] & 1
			clauses = [][]bsat.Literal{{z ^ 1, a}, {z ^ 1, b}, {z, a ^ 1, b ^ 1}}
		case DIV_XOR:
			clauses = [][]bsat.Literal{{a ^ 1, b ^ 1, z ^ 1}, {a, b, z ^ 1}, {a, b ^ 1, z}, {a ^ 1, b, z}}
		case DIV_MUX:
			a ^= raw[0] & 1
			b ^= raw[1] & 1
			c := bsat.Positive(ciStart+(raw[2]>>1)) ^ (raw[2] & 1)
			clauses = [][]bsat.Literal{{a ^ 1, b ^ 1, z}, {a ^ 1, b, z ^ 1}, {a, c ^ 1, z}, {a, c, z ^ 1}}
		default:
			panic("bmc: invalid divisor function")
		}
		for _, clause := range clauses {
			if !s.AddClause(clause...) {
				return
			}
		}
	}
}
