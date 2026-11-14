package abci

// ABC-network to GIA conversion from src/base/abci/abcCollapse.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
)

func Abc_NtkClpGia_rec(dst *gia.Gia_Man_t, node *abc.Object) int32 {
	if node.Temp >= 0 {
		return int32(node.Temp)
	}
	if abc.Abc_ObjIsNode(node) == 0 {
		panic("abci: GIA cone contains non-node")
	}
	edges := node.Fanins()
	a := Abc_NtkClpGia_rec(dst, edges[0].Object)
	b := Abc_NtkClpGia_rec(dst, edges[1].Object)
	if edges[0].Complement {
		a ^= 1
	}
	if edges[1].Complement {
		b ^= 1
	}
	node.Temp = int(gia.Gia_ManAppendAnd(dst, a, b))
	return int32(node.Temp)
}
func Abc_NtkClpGia(n *abc.Network) *gia.Gia_Man_t {
	if abc.Abc_NtkIsStrash(n) == 0 {
		panic("abci: GIA conversion requires strashed network")
	}
	dst := gia.Gia_ManStart(1000)
	dst.SetIdentity(n.Name, n.Spec)
	for _, o := range n.Objects() {
		o.Temp = -1
	}
	n.One().Object.Temp = 1
	for _, o := range n.Inputs() {
		o.Temp = int(gia.Gia_ManAppendCi(dst))
	}
	for _, o := range n.Outputs() {
		fanin := o.Fanins()[0]
		lit := Abc_NtkClpGia_rec(dst, fanin.Object)
		if fanin.Complement {
			lit ^= 1
		}
		gia.Gia_ManAppendCo(dst, lit)
	}
	return dst
}
