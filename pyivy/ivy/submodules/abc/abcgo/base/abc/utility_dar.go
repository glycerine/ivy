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
		in := g.NewInput()
		ci.Temp = int(in)
		if ci.Level != 0 {
			aig.Aig_ObjSetLevel(g, in, int32(ci.Level))
		}
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
			fmt.Printf("Warning: %d registers in this network have don't-care init values.\n", dc)
			fmt.Printf("The don't-care are assumed to be 0. The result may not verify.\n")
			fmt.Printf("Use command \"print_latch\" to see the init values of registers.\n")
			fmt.Printf("Use command \"zero\" to convert or \"init\" to change the values.\n")
		}
	}
	Abc_NtkIncrementTravId(n)
	// Like ABC, build every AND (in DFS order from the COs) before any CO.
	lits := make([]aig.Edge, len(n.cos))
	for i, co := range n.cos {
		if len(co.fanins) != 1 {
			panic("abc: invalid output")
		}
		edge := co.fanins[0]
		var lit aig.Edge
		if exors != 0 {
			lit = abcNtkToDarExor_rec(g, edge.Object)
		} else {
			lit = Abc_NtkClpOneGia_rec(g, edge.Object)
		}
		if edge.Complement {
			lit = lit.Not()
		}
		if registers != 0 && co.Kind == ObjectBI && len(co.fanouts) > 0 && Abc_LatchIsInit1(co.fanouts[0]) != 0 {
			lit = lit.Not()
		}
		lits[i] = lit
	}
	for _, lit := range lits {
		g.AddOutput(lit)
	}
	aig.Aig_ManSetRegNum(g, int32(len(latches)))
	aig.Aig_ManSetConstrNum(g, n.Constraints) // pMan->nConstrs (set once the COs exist)
	// remove dangling nodes: only a captured EXOR (pMan->fCatchExor) leaves
	// its two ANDs dangling, as the Go conversion visits only CO cones
	if exors != 0 && Abc_NtkGetChoiceNum(n) == 0 {
		aig.Aig_ManCleanup(g)
	}
	if registers != 0 {
		g.FlopNumbers = make([]int32, len(latches))
		for i := range g.FlopNumbers {
			g.FlopNumbers[i] = int32(i)
		}
	}
	return g
}

// Abc_NtkFromDarSeqSweep rebuilds an ABC AIG after a change in register count.
// Abc_NtkFromDar is abcDar.c's: the AIG has the old network's CIs and COs
// (in order), so the new network starts from the old one and keeps their
// names and latches. If the interface changed, the sequential-sweep
// conversion, which rebuilds the registers, is used instead.
func Abc_NtkFromDar(old *Network, g *aig.Graph) *Network {
	if old == nil || g == nil || g.Assertions != 0 {
		panic("abc: invalid AIG conversion")
	}
	if g.InputCount() != len(old.cis) || g.OutputCount() != len(old.cos) || int(g.RegisterCount()) != len(old.boxes) {
		return Abc_NtkFromDarSeqSweep(old, g)
	}
	dst := Abc_NtkStartFrom(old, ABC_NTK_STRASH, ABC_FUNC_AIG)
	dst.Constraints = aig.Aig_ManConstrNum(g)
	dst.Barriers = old.Barriers
	copies := make([]Edge, g.NodeSlots())
	copies[0] = dst.One()
	for i := 0; i < g.InputCount(); i++ {
		in := g.InputAt(i)
		ci := dst.cis[i]
		copies[in.NodeID()] = ci.Signal()
		ci.Level = int(aig.Aig_ObjLevel(g, in)) // initialize the CI levels
	}
	child := func(e aig.Edge) Edge {
		r := copies[e.NodeID()]
		if e.Complemented() {
			r = r.Not()
		}
		return r
	}
	manager := dst.FunctionManager.(*Abc_Aig_t)
	// rebuild the AIG in Aig_ManDfs order
	for _, e := range aig.Aig_ManDfsUnchecked(g, 1) {
		node := g.Node(e)
		switch node.Kind {
		case aig.Buffer:
			copies[e.NodeID()] = child(node.Fanins[0])
		case aig.And:
			copies[e.NodeID()] = Abc_AigAnd(manager, child(node.Fanins[0]), child(node.Fanins[1]))
		case aig.ExclusiveOr:
			copies[e.NodeID()] = Abc_AigXor(manager, child(node.Fanins[0]), child(node.Fanins[1]))
		}
	}
	for i := 0; i < g.OutputCount(); i++ {
		Abc_ObjAddFanin(dst.cos[i], child(g.Node(g.OutputAt(i)).Fanins[0]))
	}
	if Abc_NtkCheck(dst) == 0 {
		fmt.Printf("Abc_NtkFromDar(): Network check has failed.\n")
	}
	return dst
}

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
	copies := make([]Edge, g.NodeSlots())
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

// abcNtkToDarExor_rec builds a node's cone as Abc_NtkToDar does with
// fExors: pMan->fCatchExor makes Aig_And capture EXOR objects.
func abcNtkToDarExor_rec(g *aig.Graph, node *Object) aig.Edge {
	if Abc_NodeIsTravIdCurrent(node) != 0 || len(node.fanins) == 0 || node.IsCI() {
		return aig.Edge(node.Temp)
	}
	if node.Kind != ObjectNode || len(node.fanins) != 2 {
		panic("abc: expected strashed AND node")
	}
	Abc_NodeSetTravIdCurrent(node)
	a := abcNtkToDarExor_rec(g, node.fanins[0].Object)
	b := abcNtkToDarExor_rec(g, node.fanins[1].Object)
	if node.fanins[0].Complement {
		a = a.Not()
	}
	if node.fanins[1].Complement {
		b = b.Not()
	}
	result := aig.Aig_AndCatchExor(g, a, b)
	node.Temp = int(result)
	return result
}
