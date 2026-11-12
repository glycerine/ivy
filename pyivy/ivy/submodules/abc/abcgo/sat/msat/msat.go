// Package msat ports ABC's C MiniSAT implementation by Alan Mishchenko.
// It owns its watched clauses, floating activities and conflict-driven search.
package msat

type Msat_Type_t int

const (
	MSAT_FALSE   Msat_Type_t = -1
	MSAT_UNKNOWN Msat_Type_t = 0
	MSAT_TRUE    Msat_Type_t = 1
)
const (
	MSAT_VAR_UNASSIGNED = -1
	MSAT_LIT_UNASSIGNED = -2
	MSAT_ORDER_UNKNOWN  = -3
)

func MSAT_VAR2LIT(variable, sign int) int { return 2*variable + sign }
func MSAT_LITNOT(lit int) int             { return lit ^ 1 }
func MSAT_LITSIGN(lit int) int            { return lit & 1 }
func MSAT_LIT2VAR(lit int) int            { return lit >> 1 }
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
