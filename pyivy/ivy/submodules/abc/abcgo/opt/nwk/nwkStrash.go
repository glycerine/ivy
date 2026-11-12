package nwk

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/misc/tim"
)

func Nwk_ManStrashNode_rec(p *Aig_Man_t, o *Hop_Obj_t) {
	o = hop.Hop_Regular(o)
	if hop.Hop_ObjIsNode(o) == 0 || hop.Hop_ObjIsMarkA(o) != 0 {
		return
	}
	Nwk_ManStrashNode_rec(p, hop.Hop_ObjFanin0(o))
	Nwk_ManStrashNode_rec(p, hop.Hop_ObjFanin1(o))
	a := hop.Hop_ObjData(hop.Hop_ObjFanin0(o)).(aig.Edge)
	b := hop.Hop_ObjData(hop.Hop_ObjFanin1(o)).(aig.Edge)
	if hop.Hop_ObjFaninC0(o) != 0 {
		a = a.Not()
	}
	if hop.Hop_ObjFaninC1(o) != 0 {
		b = b.Not()
	}
	r := p.AND(a, b)
	if hop.Hop_ObjIsExor(o) != 0 {
		r = p.XOR(a, b)
	}
	hop.Hop_ObjSetData(o, r)
	hop.Hop_ObjSetMarkA(o)
}
func Nwk_ManStrashNode(p *Aig_Man_t, o *Nwk_Obj_t) Aig_Obj_t {
	if o.Function == nil {
		panic("nwk: node has no function")
	}
	bindings := map[*Hop_Obj_t]aig.Edge{hop.Hop_ManConst1(o.Manager.Hop): p.Const1()}
	for i, n := range o.Fanins {
		bindings[hop.Hop_IthVar(o.Manager.Hop, int32(i))] = n.Copy.(aig.Edge)
	}
	var build func(*Hop_Obj_t) aig.Edge
	build = func(edge *Hop_Obj_t) aig.Edge {
		n := hop.Hop_Regular(edge)
		r, ok := bindings[n]
		if !ok {
			a, b := build(hop.Hop_ObjChild0(n)), build(hop.Hop_ObjChild1(n))
			r = p.AND(a, b)
			if hop.Hop_ObjIsExor(n) != 0 {
				r = p.XOR(a, b)
			}
			bindings[n] = r
		}
		if hop.Hop_IsComplement(edge) != 0 {
			r = r.Not()
		}
		return r
	}
	return build(o.Function)
}
func Nwk_ManStrash(p *Nwk_Man_t) *Aig_Man_t {
	g := aig.NewGraph()
	g.Name = p.Name
	g.Spec = p.Spec
	if p.Timing != nil {
		g.Timing = tim.Tim_ManDup(p.Timing, 1)
	}
	for _, o := range p.Objects {
		if o != nil {
			o.Copy = nil
		}
	}
	for _, o := range vectorNodes(Nwk_ManDfs(p)) {
		switch o.Type {
		case NWK_OBJ_CI:
			o.Copy = g.NewInput()
		case NWK_OBJ_NODE:
			o.Copy = Nwk_ManStrashNode(g, o)
		case NWK_OBJ_CO:
			e := o.Fanins[0].Copy.(aig.Edge)
			if o.Invert {
				e = e.Not()
			}
			o.Copy = g.AddOutput(e)
		default:
			panic("nwk: unsupported strash object")
		}
	}
	return g
}

// SubstituteFunction rebuilds a local Hop function with supplied PI substitutions.
func SubstituteFunction(p *Hop_Man_t, root *Hop_Obj_t, replacements []*Hop_Obj_t) *Hop_Obj_t {
	memo := map[*Hop_Obj_t]*Hop_Obj_t{}
	for i, n := range replacements {
		memo[hop.Hop_IthVar(p, int32(i))] = n
	}
	var rec func(*Hop_Obj_t) *Hop_Obj_t
	rec = func(edge *Hop_Obj_t) *Hop_Obj_t {
		n := hop.Hop_Regular(edge)
		r, ok := memo[n]
		if !ok {
			if hop.Hop_ObjIsConst1(n) != 0 {
				r = hop.Hop_ManConst1(p)
			} else if hop.Hop_ObjIsPi(n) != 0 {
				panic("nwk: missing local variable substitution")
			} else {
				a, b := rec(hop.Hop_ObjChild0(n)), rec(hop.Hop_ObjChild1(n))
				r = hop.Hop_And(p, a, b)
				if hop.Hop_ObjIsExor(n) != 0 {
					r = hop.Hop_Exor(p, a, b)
				}
			}
			memo[n] = r
		}
		return hop.Hop_NotCond(r, hop.Hop_IsComplement(edge))
	}
	return rec(root)
}
