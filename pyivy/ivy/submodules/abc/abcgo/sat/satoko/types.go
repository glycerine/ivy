// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

// Literals use ABC encoding: variable<<1 | negative-sign.
func Literal(variable int, negative bool) int {
	if variable < 0 {
		panic("negative variable")
	}
	v := variable << 1
	if negative {
		v |= 1
	}
	return v
}
func Variable(literal int) int                      { return literal >> 1 }
func Negate(literal int) int                        { return literal ^ 1 }
func var2lit(variable uint32, polarity byte) uint32 { return variable*2 + uint32(polarity&1) }
func lit2var(literal uint32) uint32                 { return literal >> 1 }
func lit_compl(literal uint32) uint32               { return literal ^ 1 }
func lit_polarity(literal uint32) byte              { return byte(literal & 1) }
