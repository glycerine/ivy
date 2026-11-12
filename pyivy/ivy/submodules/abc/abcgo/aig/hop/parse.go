package hop

import (
	"github.com/glycerine/abcgo/misc/parse"
	"io"
)

// EquationOperations connects the generic equation parser to this manager's
// structurally hashed nodes and complemented edges.
func EquationOperations(p *Hop_Man_t) parse.Operations[*Hop_Obj_t] {
	return parse.Operations[*Hop_Obj_t]{
		Constant: func(v bool) *Hop_Obj_t {
			if v {
				return Hop_ManConst1(p)
			}
			return Hop_ManConst0(p)
		},
		Variable: func(i int) *Hop_Obj_t { return Hop_IthVar(p, int32(i)) }, Not: Hop_Not,
		And: func(a, b *Hop_Obj_t) *Hop_Obj_t { return Hop_And(p, a, b) },
		Or:  func(a, b *Hop_Obj_t) *Hop_Obj_t { return Hop_Or(p, a, b) },
		Xor: func(a, b *Hop_Obj_t) *Hop_Obj_t { return Hop_Exor(p, a, b) },
	}
}
func CompileEquation(p *Hop_Man_t, formula string, names []string) (*Hop_Obj_t, error) {
	return parse.CompileEquation(formula, names, EquationOperations(p))
}
func ParseEquation(p *Hop_Man_t, output io.Writer, formula string, names []string) *Hop_Obj_t {
	return parse.Parse_FormulaParserEqn(output, formula, names, EquationOperations(p))
}
