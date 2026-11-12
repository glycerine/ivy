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

// Abc_NtkFromDarSeqSweep rebuilds an ABC AIG after a change in register count.
func Abc_NtkFromDarSeqSweep(old *Network, g *aig.Graph) *Network {
	if old == nil || g == nil || g.Assertions != 0 || old.Barriers != 0 {
		panic("abc: invalid sequential AIG conversion")
	}
	regs := int(g.RegisterCount())
	piCount, poCount := g.InputCount()-regs, g.OutputCount()-regs
	if piCount < 0 || poCount < 0 {
		panic("abc: invalid AIG register count")
	}
	dst := Abc_NtkStartFromNoLatches(old, ABC_NTK_STRASH, ABC_FUNC_AIG)
	dst.Constraints = aig.Aig_ManConstrNum(g)
	dst.Barriers = int(g.BarrierBuffers)
	for len(dst.cis) < piCount {
		obj := Abc_NtkCreatePi(dst)
		Abc_ObjAssignName(obj, Abc_ObjName(obj), "")
	}
	Abc_NtkOrderCisCos(dst)
	if len(dst.cis) != piCount || len(dst.cos) != poCount {
		panic("abc: AIG interface mismatch")
	}
	copies := make([]Edge, len(g.Nodes()))
	valid := make([]bool, len(copies))
	copies[0], valid[0] = dst.One(), true
	inputs := g.Inputs()
	for i := 0; i < piCount; i++ {
		id := inputs[i].NodeID()
		copies[id], valid[id] = dst.cis[i].Signal(), true
	}
	for i := 0; i < regs; i++ {
		output := Abc_NtkAddLatch(dst, nil, ABC_INIT_ZERO)
		id := inputs[piCount+i].NodeID()
		copies[id], valid[id] = output.Signal(), true
	}
	Abc_NtkOrderCisCos(dst)
	manager := dst.FunctionManager.(*Abc_Aig_t)
	var build func(aig.Edge) Edge
	build = func(e aig.Edge) Edge {
		id := e.NodeID()
		if !valid[id] {
			node := g.Node(aig.Edge(id << 1))
			switch node.Kind {
			case aig.And:
				copies[id] = Abc_AigAnd(manager, build(node.Fanins[0]), build(node.Fanins[1]))
			case aig.ExclusiveOr:
				copies[id] = Abc_AigXor(manager, build(node.Fanins[0]), build(node.Fanins[1]))
			case aig.Buffer:
				copies[id] = build(node.Fanins[0])
			default:
				panic("abc: invalid AIG node")
			}
			valid[id] = true
		}
		result := copies[id]
		if e.Complemented() {
			result = result.Not()
		}
		return result
	}
	for i, output := range g.Outputs() {
		target := dst.cos[i]
		edge := build(g.Node(output).Fanins[0])
		name := Abc_ObjName(target)
		for _, ci := range dst.cis {
			if name != "" && Abc_ObjName(ci) == name {
				edge = ci.Signal()
				break
			}
		}
		Abc_ObjAddFanin(target, edge)
	}
	if len(g.FlopNumbers) == regs {
		oldLatches := old.ObjectsOfKind(ObjectLatch)
		newLatches := dst.ObjectsOfKind(ObjectLatch)
		for i, num := range g.FlopNumbers {
			if num < 0 || int(num) >= len(oldLatches) {
				continue
			}
			a, b := oldLatches[num], newLatches[i]
			if name := Abc_ObjName(a); name != "" {
				Abc_ObjAssignName(b, name, "")
			}
			if name := Abc_ObjName(Abc_ObjFanin0(a)); name != "" {
				Abc_ObjAssignName(Abc_ObjFanin0(b), name, "")
			}
			if name := Abc_ObjName(Abc_ObjFanout0(a)); name != "" {
				Abc_ObjAssignName(Abc_ObjFanout0(b), name, "")
			}
		}
	} else {
		Abc_NtkAddDummyBoxNames(dst)
	}
	return dst
}
