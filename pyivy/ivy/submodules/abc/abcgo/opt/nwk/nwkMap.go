package nwk

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/aig/saig"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

type If_Par_t = if_.If_Par_t
type If_Man_t = if_.If_Man_t
type If_Obj_t = if_.If_Obj_t

func Nwk_ManSetIfParsDefault(p *If_Par_t) {
	if_.If_ManSetDefaultPars(p)
	p.SetLUTSize(6)
	p.SetCutMin(false)
}
func Nwk_ManToIf(g *Aig_Man_t, pars *If_Par_t, mapping *Vec_Ptr_t) *If_Man_t {
	p := if_.If_ManStart(pars)
	if p == nil {
		return nil
	}
	copies := map[int]*If_Obj_t{0: if_.If_ManConst1(p)}
	for _, e := range g.Inputs() {
		copies[e.NodeID()] = if_.If_ManCreateCi(p)
	}
	var convert func(aig.Edge) *If_Obj_t
	convert = func(e aig.Edge) *If_Obj_t {
		n := copies[e.NodeID()]
		if n == nil {
			object := g.Node(e)
			switch object.Kind {
			case aig.And:
				n = if_.If_ManCreateAnd(p, convert(object.Fanins[0]), convert(object.Fanins[1]))
			case aig.ExclusiveOr:
				n = if_.If_ManCreateXor(p, convert(object.Fanins[0]), convert(object.Fanins[1]))
			case aig.Buffer:
				n = convert(object.Fanins[0])
			default:
				panic("nwk: unsupported AIG mapping object")
			}
			copies[e.NodeID()] = n
		}
		return if_.If_NotCond(n, bit(e.Complemented()))
	}
	// Materialize the entire choice graph before linking classes.
	for _, edge := range g.LiveEdges() {
		kind := g.Node(edge).Kind
		if kind == aig.And || kind == aig.ExclusiveOr || kind == aig.Buffer {
			convert(edge)
		}
	}
	for _, edge := range g.LiveEdges() {
		if equivalent := aig.Aig_ObjEquiv(g, edge); equivalent != aig.InvalidEdge {
			previous := edge
			seen := map[int]bool{edge.NodeID(): true}
			for equivalent != aig.InvalidEdge {
				if seen[equivalent.NodeID()] {
					panic("nwk: cyclic AIG choice class")
				}
				seen[equivalent.NodeID()] = true
				if_.If_ObjSetChoice(if_.If_Regular(convert(previous)), if_.If_Regular(convert(equivalent)))
				previous = equivalent
				equivalent = aig.Aig_ObjEquiv(g, equivalent)
			}
			if_.If_ManCreateChoice(p, if_.If_Regular(convert(edge)))
		}
	}
	for _, e := range g.Outputs() {
		copies[e.NodeID()] = if_.If_ManCreateCo(p, convert(g.Node(e).Fanins[0]))
	}
	vec.Vec_PtrFill(mapping, int(aig.Aig_ManObjNumMax(g)), nil)
	for id, n := range copies {
		vec.Vec_PtrWriteEntry(mapping, id, n)
	}
	if pars.Power() {
		registers := g.RegisterCount()
		aig.Aig_ManSetRegNum(g, 0)
		probabilities, err := saig.SwitchingProbabilities(g, 48, 16, false)
		aig.Aig_ManSetRegNum(g, registers)
		if err != nil {
			panic(err)
		}
		mapped := make([]float32, if_.If_ManObjNum(p))
		for _, edge := range g.LiveEdges() {
			if object := copies[edge.NodeID()]; object != nil {
				mapped[if_.If_ObjId(if_.If_Regular(object))] = probabilities[edge.NodeID()]
			}
		}
		p.SetSwitchingProbabilities(mapped)
	}
	return p
}
func localIfFunction(h *Hop_Man_t, root *If_Obj_t, leaves []int32) *Hop_Obj_t {
	memo := map[*If_Obj_t]*Hop_Obj_t{}
	active := map[*If_Obj_t]bool{}
	leafIndices := map[int32]int{}
	for i, id := range leaves {
		leafIndices[id] = i
	}
	var build func(*If_Obj_t) *Hop_Obj_t
	build = func(edge *If_Obj_t) *Hop_Obj_t {
		n := if_.If_Regular(edge)
		r, ok := memo[n]
		if !ok {
			if index, found := leafIndices[if_.If_ObjId(n)]; found {
				r = hop.Hop_IthVar(h, int32(index))
			} else if if_.If_ObjIsConst1(n) != 0 {
				r = hop.Hop_ManConst1(h)
			} else if if_.If_ObjIsCi(n) != 0 || active[n] {
				return nil
			} else {
				active[n] = true
				for alternative := n; alternative != nil; alternative = alternative.Equivalent() {
					a := build(if_.If_ObjFanin0(alternative))
					if a == nil {
						continue
					}
					b := build(if_.If_ObjFanin1(alternative))
					if b == nil {
						continue
					}
					a = hop.Hop_NotCond(a, if_.If_ObjFaninC0(alternative))
					b = hop.Hop_NotCond(b, if_.If_ObjFaninC1(alternative))
					r = hop.Hop_NotCond(hop.Hop_And(h, a, b), bit(alternative.Phase() != n.Phase()))
					break
				}
				delete(active, n)
				if r == nil {
					return nil
				}
			}
			memo[n] = r
		}
		return hop.Hop_NotCond(r, if_.If_IsComplement(edge))
	}
	result := build(root)
	if result == nil {
		panic("nwk: mapped cut cannot reconstruct local function")
	}
	return result
}
func Nwk_NodeIfToHop2_rec(h *Hop_Man_t, p *If_Man_t, o *If_Obj_t, visited *Vec_Ptr_t) *Hop_Obj_t {
	return Nwk_NodeIfToHop(h, p, o)
}
func Nwk_NodeIfToHop(h *Hop_Man_t, p *If_Man_t, o *If_Obj_t) *Hop_Obj_t {
	cut := if_.If_ObjCutBest(if_.If_Regular(o))
	leaves := if_.If_CutLeaves(cut)[:if_.If_CutLeaveNum(cut)]
	return localIfFunction(h, o, leaves)
}
func Nwk_ManFromIf(p *If_Man_t, g *Aig_Man_t, aigToIf *Vec_Ptr_t) *Nwk_Man_t {
	n := Nwk_ManAlloc()
	n.Name = g.Name
	n.Spec = g.Spec
	copies := map[int32]*Nwk_Obj_t{}
	constant := Nwk_ManCreateNode(n, 0, 0)
	constant.Function = hop.Hop_ManConst1(n.Hop)
	copies[if_.If_ObjId(if_.If_ManConst1(p))] = constant
	for i := int32(0); i < if_.If_ManCiNum(p); i++ {
		object := if_.If_ManCi(p, i)
		copies[if_.If_ObjId(object)] = Nwk_ManCreateCi(n, 0)
	}
	active := map[int32]bool{}
	var convert func(*If_Obj_t) *Nwk_Obj_t
	convert = func(o *If_Obj_t) *Nwk_Obj_t {
		o = if_.If_Regular(o)
		id := if_.If_ObjId(o)
		if copy := copies[id]; copy != nil {
			return copy
		}
		if active[id] {
			panic("nwk: cyclic mapped cut")
		}
		active[id] = true
		cut := if_.If_ObjCutBest(o)
		leaves := if_.If_CutLeaves(cut)[:if_.If_CutLeaveNum(cut)]
		fanins := make([]*Nwk_Obj_t, len(leaves))
		for i, leaf := range leaves {
			fanins[i] = convert(if_.If_ManObj(p, leaf))
		}
		copy := Nwk_ManCreateNode(n, int32(len(leaves)), 0)
		for _, fanin := range fanins {
			Nwk_ObjAddFanin(copy, fanin)
		}
		copy.Function = localIfFunction(n.Hop, o, leaves)
		copy.Arrival = if_.If_ObjArrTime(o)
		copies[id] = copy
		delete(active, id)
		return copy
	}
	for i := int32(0); i < if_.If_ManCoNum(p); i++ {
		co := if_.If_ManCo(p, i)
		driver := if_.If_ObjFanin0(co)
		fanin := convert(driver)
		o := Nwk_ManCreateCo(n)
		Nwk_ObjAddFanin(o, fanin)
		o.Invert = if_.If_ObjFaninC0(co) != 0
	}
	if p.TimingManager() != nil {
		n.Timing = tim.Tim_ManDup(p.TimingManager(), 0)
	}
	Nwk_ManMinimumBase(n, 0)
	return n
}
func Nwk_MappingIf(g *Aig_Man_t, timing *Tim_Man_t, pars *If_Par_t) *Nwk_Man_t {
	pars.SetArrivalTimes(make([]float32, g.InputCount()))
	mapping := vec.Vec_PtrAlloc(0)
	p := Nwk_ManToIf(g, pars, mapping)
	if p == nil {
		return nil
	}
	defer if_.If_ManStop(p)
	if timing != nil {
		p.SetTimingManager(tim.Tim_ManDup(timing, 0))
	}
	pars.SetCutMin(false)
	if if_.If_ManPerformMapping(p) == 0 {
		return nil
	}
	n := Nwk_ManFromIf(p, g, mapping)
	if pars.Bidec() && pars.LUTSize() <= 8 {
		Nwk_ManBidecResyn(n, 0)
	}
	return n
}
