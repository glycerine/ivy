package abc

import (
	"strings"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/map/mio"
)

// Abc_GateToType maps the library gate name to ABC's operator code.
func Abc_GateToType(node *Object) int32 {
	if node == nil {
		panic("abc: nil mapped gate")
	}
	gate, ok := node.Data.(*mio.Mio_Gate_t)
	if !ok || gate == nil {
		panic("abc: node has no mapped gate")
	}
	name := mio.Mio_GateReadName(gate)
	for _, entry := range []struct {
		prefix string
		code   int32
	}{
		{"buf", miniaig.ABC_OPER_BIT_BUF}, {"inv", miniaig.ABC_OPER_BIT_INV},
		{"and", miniaig.ABC_OPER_BIT_AND}, {"nand", miniaig.ABC_OPER_BIT_NAND},
		{"or", miniaig.ABC_OPER_BIT_OR}, {"nor", miniaig.ABC_OPER_BIT_NOR},
		{"xor", miniaig.ABC_OPER_BIT_XOR}, {"xnor", miniaig.ABC_OPER_BIT_NXOR},
		{"zero", miniaig.ABC_OPER_CONST_F}, {"one", miniaig.ABC_OPER_CONST_T},
	} {
		if strings.HasPrefix(name, entry.prefix) {
			return entry.code
		}
	}
	panic("abc: unrecognized mapped gate name: " + name)
}
