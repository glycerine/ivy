package fsim

import "github.com/glycerine/abcgo/aig/aig"

type Fsim_Obj_t struct {
	Node, Fan0, Fan1 int
	Kind             aig.NodeKind
	Index            int
	Source           aig.Edge
}
type Fsim_Man_t struct {
	Graph                 *aig.Graph
	PIs, POs, CIs, COs    int
	CrossCut, Frontier    int
	ops, program          []Fsim_Obj_t
	encoded               []byte
	compressed            bool
	references            []int
	sourceIDs             map[int]int
	CIOrder               []int
	nextNumber, constSlot int
	words                 int
	data                  []uint32
	observer              func(int, []uint32)
}
type Manager = Fsim_Man_t

func Fsim_Var2Lit(variable, complement int) int { return variable*2 + complement }
func Fsim_Lit2Var(literal int) int              { return literal >> 1 }
func Fsim_LitIsCompl(literal int) int           { return literal & 1 }
func Fsim_LitNot(literal int) int               { return literal ^ 1 }
func Fsim_LitNotCond(literal, condition int) int {
	if condition != 0 {
		return literal ^ 1
	}
	return literal
}
func Fsim_LitRegular(literal int) int { return literal &^ 1 }
func fsimFanins(o Fsim_Obj_t) []int {
	switch o.Kind {
	case aig.And, aig.ExclusiveOr:
		return []int{o.Fan0, o.Fan1}
	case aig.Buffer, aig.Output:
		return []int{o.Fan0}
	}
	return nil
}
