package nwk

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/misc/vec"
)

func Nwk_ManDeriveFromAig(g *Aig_Man_t) *Nwk_Man_t {
	p := Nwk_ManAlloc()
	p.Name = g.Name
	p.Spec = g.Spec
	copies := map[int]*Nwk_Obj_t{}
	constant := Nwk_ManCreateNode(p, 0, 0)
	constant.Function = hop.Hop_ManConst1(p.Hop)
	constant.Copy = g.Const1()
	copies[0] = constant
	for _, e := range g.Inputs() {
		o := Nwk_ManCreateCi(p, 0)
		o.Copy = e
		copies[e.NodeID()] = o
	}
	var convert func(aig.Edge) *Nwk_Obj_t
	convert = func(e aig.Edge) *Nwk_Obj_t {
		if n := copies[e.NodeID()]; n != nil {
			return n
		}
		n := g.Node(e)
		switch n.Kind {
		case aig.And, aig.ExclusiveOr, aig.Buffer:
		default:
			panic("nwk: invalid internal AIG object")
		}
		arity := 2
		if n.Kind == aig.Buffer {
			arity = 1
		}
		fanins := make([]*Nwk_Obj_t, arity)
		for i := range fanins {
			fanins[i] = convert(n.Fanins[i])
		}
		o := Nwk_ManCreateNode(p, int32(arity), int32(n.Refs))
		o.Copy = aig.Edge(e.NodeID() << 1)
		copies[e.NodeID()] = o
		for _, f := range fanins {
			Nwk_ObjAddFanin(o, f)
		}
		a := hop.Hop_NotCond(hop.Hop_IthVar(p.Hop, 0), bit(n.Fanins[0].Complemented()))
		o.Function = a
		if arity == 2 {
			b := hop.Hop_NotCond(hop.Hop_IthVar(p.Hop, 1), bit(n.Fanins[1].Complemented()))
			o.Function = hop.Hop_And(p.Hop, a, b)
			if n.Kind == aig.ExclusiveOr {
				o.Function = hop.Hop_Exor(p.Hop, a, b)
			}
		}
		return o
	}
	for _, e := range g.Outputs() {
		n := g.Node(e)
		driver := n.Fanins[0]
		fanin := convert(driver)
		o := Nwk_ManCreateCo(p)
		o.Copy = e
		o.Invert = driver.Complemented()
		Nwk_ObjAddFanin(o, fanin)
	}
	p.Latches = g.RegisterCount()
	p.TruePIs = int32(len(p.Inputs)) - p.Latches
	p.TruePOs = int32(len(p.Outputs)) - p.Latches
	return p
}
func Nwk_ManDeriveRetimingCut(g *Aig_Man_t, forward, verbose int32) *Vec_Ptr_t {
	p := Nwk_ManDeriveFromAig(g)
	mapping := map[*Nwk_Obj_t]aig.Edge{}
	for _, o := range p.Objects {
		if o != nil {
			mapping[o] = o.Copy.(aig.Edge)
		}
	}
	var cut *Vec_Ptr_t
	if forward != 0 {
		cut = Nwk_ManRetimeCutForward(p, g.RegisterCount(), verbose)
	} else {
		cut = Nwk_ManRetimeCutBackward(p, g.RegisterCount(), verbose)
	}
	result := vec.Vec_PtrAlloc(vec.Vec_PtrSize(cut))
	for _, o := range vectorNodes(cut) {
		vec.Vec_PtrPush(result, mapping[o])
	}
	Nwk_ManFree(p)
	return result
}

// GiaNetwork is implemented by GIA without introducing a reverse package import.
// Gate kinds match GIA's network interface: const0,input,AND,XOR,MUX,buffer,CO.
type GiaNetwork interface {
	ObjectCount() int
	InputLiterals() []int32
	OutputLiterals() []int32
	Gate(int32) (int32, int32, int32, int)
}

func Nwk_ManColleacReached_rec(g GiaNetwork, id int32, nodes, leaves *Vec_Int_t) {
	seen := map[int32]bool{}
	for _, v := range vec.Vec_IntArray(nodes) {
		seen[v] = true
	}
	for _, v := range vec.Vec_IntArray(leaves) {
		seen[v] = true
	}
	var visit func(int32)
	visit = func(id int32) {
		if seen[id] {
			return
		}
		seen[id] = true
		a, b, c, kind := g.Gate(id)
		if kind == 1 {
			vec.Vec_IntPush(leaves, id)
			return
		}
		if kind == 0 {
			return
		}
		if kind < 2 {
			panic("nwk: invalid GIA gate")
		}
		visit(a >> 1)
		if kind == 2 || kind == 3 || kind == 4 {
			visit(b >> 1)
		}
		if kind == 4 {
			visit(c >> 1)
		}
		vec.Vec_IntPush(nodes, id)
	}
	visit(id)
}
func Nwk_ManCreateFromGia(g GiaNetwork, ppis, nodes, leaves *Vec_Int_t, inverse **Vec_Int_t) *Nwk_Man_t {
	p := Nwk_ManAlloc()
	copies := map[int32]*Nwk_Obj_t{}
	mapping := vec.Vec_IntAlloc(0)
	constant := Nwk_ManCreateNode(p, 0, 0)
	constant.Function = hop.Hop_ManConst0(p.Hop)
	copies[0] = constant
	vec.Vec_IntPush(mapping, 0)
	for _, id := range vec.Vec_IntArray(leaves) {
		o := Nwk_ManCreateCi(p, 0)
		copies[id] = o
		vec.Vec_IntPush(mapping, id)
	}
	for i := vec.Vec_IntSize(leaves); i < vec.Vec_IntSize(ppis); i++ {
		Nwk_ManCreateCi(p, 0)
		vec.Vec_IntPush(mapping, -1)
	}
	for _, id := range vec.Vec_IntArray(nodes) {
		a, b, c, kind := g.Gate(id)
		lits := []int32{a, b}
		if kind == 4 {
			lits = append(lits, c)
		} else if kind == 5 || kind == 6 {
			lits = lits[:1]
		}
		o := Nwk_ManCreateNode(p, int32(len(lits)), 0)
		args := make([]*Hop_Obj_t, len(lits))
		for i, lit := range lits {
			fanin := copies[lit>>1]
			if fanin == nil {
				panic("nwk: GIA nodes not topological")
			}
			Nwk_ObjAddFanin(o, fanin)
			args[i] = hop.Hop_NotCond(hop.Hop_IthVar(p.Hop, int32(i)), lit&1)
		}
		switch kind {
		case 2:
			o.Function = hop.Hop_And(p.Hop, args[0], args[1])
		case 3:
			o.Function = hop.Hop_Exor(p.Hop, args[0], args[1])
		case 4:
			o.Function = hop.Hop_Mux(p.Hop, args[0], args[1], args[2])
		case 5, 6:
			o.Function = args[0]
		default:
			panic("nwk: unsupported GIA node kind")
		}
		copies[id] = o
		vec.Vec_IntPush(mapping, id)
	}
	for _, id := range vec.Vec_IntArray(ppis) {
		o := Nwk_ManCreateCo(p)
		fanin := copies[id]
		if fanin == nil {
			panic(fmt.Sprintf("nwk: missing GIA output %d", id))
		}
		Nwk_ObjAddFanin(o, fanin)
		vec.Vec_IntPush(mapping, id)
	}
	for i := vec.Vec_IntSize(ppis); i < vec.Vec_IntSize(leaves); i++ {
		o := Nwk_ManCreateCo(p)
		Nwk_ObjAddFanin(o, constant)
		vec.Vec_IntPush(mapping, 0)
	}
	if inverse != nil {
		*inverse = mapping
	}
	return p
}

// GiaAbstractionNetwork supplies the abstraction PPIs and mutable gate classes
// needed by the source min-cut abstraction refinement driver.
type GiaAbstractionNetwork interface {
	GiaNetwork
	AbstractionPPIs() []int32
	SetGateClass(int32, int32)
}

func Nwk_ManDeriveMinCut(g GiaAbstractionNetwork, verbose int32) {
	leaves, nodes, ppis := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	for _, id := range g.AbstractionPPIs() {
		vec.Vec_IntPush(ppis, id)
		Nwk_ManColleacReached_rec(g, id, nodes, leaves)
	}
	var inverse *Vec_Int_t
	p := Nwk_ManCreateFromGia(g, ppis, nodes, leaves, &inverse)
	if len(p.Inputs) != len(p.Outputs) {
		panic("nwk: abstraction interface is not square")
	}
	cut := Nwk_ManRetimeCutBackward(p, int32(len(p.Inputs)), verbose)
	newNodes, newLeaves := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	for _, o := range vectorNodes(cut) {
		id := vec.Vec_IntEntry(inverse, int(o.ID))
		if id > 0 {
			Nwk_ManColleacReached_rec(g, id, newNodes, newLeaves)
		}
	}
	retained := map[int32]bool{}
	for _, id := range vec.Vec_IntArray(newNodes) {
		retained[id] = true
	}
	changed := 0
	for _, id := range vec.Vec_IntArray(nodes) {
		if !retained[id] {
			g.SetGateClass(id, 1)
			changed++
		}
	}
	if verbose != 0 {
		fmt.Fprintf(writer(p), "Min-cut: %d -> %d. Nodes %d -> %d. New gate classes %d.\n", vec.Vec_IntSize(ppis)+1, vec.Vec_PtrSize(cut), vec.Vec_IntSize(nodes), vec.Vec_IntSize(newNodes), changed)
	}
	Nwk_ManFree(p)
}
