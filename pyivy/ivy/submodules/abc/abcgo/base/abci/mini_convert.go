package abci

// MiniAIG/network conversions from abcMini.c.
import (
	"fmt"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func miniCopyFromLiteral(n *abc.Network, copies *vec.Vec_Int_t, literal int) abc.Edge {
	mapped := vec.Vec_IntEntry(copies, literal>>1) ^ int32(literal&1)
	edge := n.Object(int(mapped >> 1)).Signal()
	if mapped&1 != 0 {
		edge = edge.Not()
	}
	return edge
}

func Abc_NodeFanin0Copy(n *abc.Network, copies *vec.Vec_Int_t, p *miniaig.Mini_Aig_t, id int32) abc.Edge {
	return miniCopyFromLiteral(n, copies, miniaig.Mini_AigNodeFanin0(p, int(id)))
}

func Abc_NodeFanin1Copy(n *abc.Network, copies *vec.Vec_Int_t, p *miniaig.Mini_Aig_t, id int32) abc.Edge {
	return miniCopyFromLiteral(n, copies, miniaig.Mini_AigNodeFanin1(p, int(id)))
}

func miniCopiedFanin(copies []abc.Edge, literal int) abc.Edge {
	edge := copies[literal>>1]
	if literal&1 != 0 {
		edge = edge.Not()
	}
	return edge
}
func Abc_NtkFromMiniAig(p *miniaig.Mini_Aig_t) *abc.Network {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	n.Name = "MiniAig"
	copies := make([]abc.Edge, 1, miniaig.Mini_AigNodeNum(p))
	copies[0] = n.One().Not()
	for i := 1; i < miniaig.Mini_AigNodeNum(p); i++ {
		var out abc.Edge
		switch {
		case miniaig.Mini_AigNodeIsPi(p, i):
			out = abc.Abc_NtkCreatePi(n).Signal()
		case miniaig.Mini_AigNodeIsPo(p, i):
			o := abc.Abc_NtkCreatePo(n)
			abc.Abc_ObjAddFanin(o, miniCopiedFanin(copies, miniaig.Mini_AigNodeFanin0(p, i)))
			out = o.Signal()
		case miniaig.Mini_AigNodeIsAnd(p, i):
			out = abc.Abc_AigAnd(n.FunctionManager.(*abc.Abc_Aig_t), miniCopiedFanin(copies, miniaig.Mini_AigNodeFanin0(p, i)), miniCopiedFanin(copies, miniaig.Mini_AigNodeFanin1(p, i)))
		default:
			panic("abci: unknown MiniAIG object")
		}
		copies = append(copies, out)
	}
	abc.Abc_AigCleanup(n.FunctionManager.(*abc.Abc_Aig_t))
	abc.Abc_NtkAddDummyPiNames(n)
	abc.Abc_NtkAddDummyPoNames(n)
	if abc.Abc_NtkCheck(n) == 0 {
		fmt.Println("Abc_NtkFromMini(): Network check has failed.")
	}
	if miniaig.Mini_AigRegNum(p) > 0 {
		n = abc.Abc_NtkRestrashWithLatches(n, int32(miniaig.Mini_AigRegNum(p)))
	}
	return n
}
func miniFaninLiteral(node *abc.Object, index int) int {
	fanin := node.Fanins()[index]
	lit := fanin.Object.Temp
	if fanin.Complement {
		lit ^= 1
	}
	return lit
}
func Abc_NodeFanin0Copy2(node *abc.Object) int32 { return int32(miniFaninLiteral(node, 0)) }
func Abc_NodeFanin1Copy2(node *abc.Object) int32 { return int32(miniFaninLiteral(node, 1)) }
func Abc_NtkToMiniAig(n *abc.Network) *miniaig.Mini_Aig_t {
	if abc.Abc_NtkIsStrash(n) == 0 {
		panic("abci: MiniAIG conversion requires strashed network")
	}
	p := miniaig.Mini_AigStart()
	abc.Abc_NtkCleanCopy(n)
	n.One().Object.Temp = miniaig.Mini_AigLitConst1()
	for _, o := range n.Inputs() {
		o.Temp = miniaig.Mini_AigCreatePi(p)
	}
	for _, o := range n.ObjectsOfKind(abc.ObjectNode) {
		o.Temp = miniaig.Mini_AigAnd(p, miniFaninLiteral(o, 0), miniFaninLiteral(o, 1))
	}
	for _, o := range n.Outputs() {
		o.Temp = miniaig.Mini_AigCreatePo(p, miniFaninLiteral(o, 0))
	}
	miniaig.Mini_AigSetRegNum(p, int(abc.Abc_NtkLatchNum(n)))
	return p
}
