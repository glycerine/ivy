package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

// Abc_NtkToDar copies a strashed ABC network to the native AIG manager.
// Sequential mode normalizes initial-one registers by complementing their
// corresponding input and output edges, as in ABC.
func Abc_NtkToDar(n *Network, exors, registers int32) *aig.Graph {
	if n == nil || Abc_NtkIsStrash(n) == 0 {
		panic("abc: expected strashed network")
	}
	latches := n.ObjectsOfKind(ObjectLatch)
	if registers != 0 && len(n.boxes) != len(latches) {
		panic("abc: expected latch-only boxes")
	}
	g := aig.NewGraph()
	g.Name, g.Spec = n.Name, n.Spec
	g.BarrierBuffers = int32(n.Barriers)
	for _, obj := range n.Objects() {
		obj.Temp = -1
	}
	n.constant.Temp = int(g.Const1())
	for _, ci := range n.cis {
		ci.Temp = int(g.NewInput())
	}
	if registers != 0 {
		var dc int
		for _, latch := range latches {
			if Abc_LatchIsInitDc(latch) != 0 {
				Abc_LatchSetInit0(latch)
				dc++
			}
			if Abc_LatchIsInit1(latch) != 0 {
				out := Abc_ObjFanout0(latch)
				out.Temp ^= 1
			}
		}
		if dc != 0 {
			fmt.Printf("Warning: %d registers have don't-care init values; assuming zero.\n", dc)
		}
	}
	Abc_NtkIncrementTravId(n)
	for _, co := range n.cos {
		if len(co.fanins) != 1 {
			panic("abc: invalid output")
		}
		edge := co.fanins[0]
		lit := Abc_NtkClpOneGia_rec(g, edge.Object)
		if edge.Complement {
			lit = lit.Not()
		}
		if registers != 0 && co.Kind == ObjectBI && len(co.fanouts) > 0 && Abc_LatchIsInit1(co.fanouts[0]) != 0 {
			lit = lit.Not()
		}
		g.AddOutput(lit)
	}
	aig.Aig_ManSetRegNum(g, int32(len(latches)))
	if registers != 0 {
		g.FlopNumbers = make([]int32, len(latches))
		for i := range g.FlopNumbers {
			g.FlopNumbers[i] = int32(i)
		}
	}
	return g
}
