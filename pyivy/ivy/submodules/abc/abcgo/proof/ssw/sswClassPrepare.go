package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ssw_ClassesPrepareSimple(g *aig.Graph, latchOnly, maxLevels int32) *Ssw_Cla_t {
	p := Ssw_ClassesStart(g)
	inputs := g.Inputs()
	nRegs := int(g.RegisterCount())
	latches := make(map[int]bool, nRegs)
	for _, obj := range inputs[len(inputs)-nRegs:] {
		latches[obj.NodeID()] = true
	}
	for _, obj := range g.LiveEdges() {
		n := g.Node(obj)
		if latchOnly != 0 {
			if !latches[obj.NodeID()] {
				continue
			}
		} else {
			if n.Kind != aig.And && n.Kind != aig.ExclusiveOr && !latches[obj.NodeID()] {
				continue
			}
			if maxLevels != 0 && int32(n.Level) > maxLevels {
				continue
			}
		}
		Ssw_ObjSetConst1Cand(g, obj)
		p.nCands1++
	}
	return p
}
func Ssw_ClassesPrepareFromReprs(g *aig.Graph) *Ssw_Cla_t {
	p := Ssw_ClassesStart(g)
	classes := make(map[int][]aig.Edge)
	for _, obj := range g.LiveEdges() {
		repr := aig.Aig_ObjRepr(g, obj)
		if repr == g.Const1() {
			p.nCands1++
			continue
		}
		if repr != aig.InvalidEdge {
			id := repr.NodeID()
			if classes[id] == nil {
				classes[id] = append(classes[id], repr)
			}
			classes[id] = append(classes[id], obj)
		}
	}
	for _, class := range classes {
		Ssw_ObjAddClass(p, class[0], class, int32(len(class)))
	}
	return p
}

func Ssw_ClassesPrepareTargets(g *aig.Graph) *Ssw_Cla_t {
	p := Ssw_ClassesStart(g)
	outputs := g.Outputs()
	numPOs := len(outputs) - int(g.RegisterCount())
	for _, obj := range outputs[:numPOs] {
		fanin := g.Node(obj).Fanins[0] &^ 1
		Ssw_ObjSetConst1Cand(g, fanin)
		p.nCands1++
	}
	return p
}
func Ssw_ClassesPreparePairs(g *aig.Graph, classes []*vec.Vec_Int_t) *Ssw_Cla_t {
	p := Ssw_ClassesStart(g)
	if len(classes) > 0 && classes[0] != nil {
		constants := classes[0]
		for i := 0; i < vec.Vec_IntSize(constants); i++ {
			id := vec.Vec_IntEntry(constants, i)
			if (i == 0) != (id == 0) {
				panic("ssw: invalid constant class")
			}
			if i == 0 {
				continue
			}
			Ssw_ObjSetConst1Cand(g, aig.Edge(id)<<1)
			p.nCands1++
		}
	}
	for i := 1; i < len(classes); i++ {
		v := classes[i]
		if v == nil {
			continue
		}
		if vec.Vec_IntSize(v) < 2 {
			panic("ssw: class must contain at least two nodes")
		}
		group := make([]aig.Edge, vec.Vec_IntSize(v))
		for i := range group {
			id := vec.Vec_IntEntry(v, i)
			if id < 0 || (i > 0 && id <= vec.Vec_IntEntry(v, i-1)) {
				panic("ssw: class must be ordered")
			}
			group[i] = aig.Edge(id) << 1
			if i > 0 {
				aig.Aig_ObjSetRepr(g, group[i], group[0])
			}
		}
		Ssw_ObjAddClass(p, group[0], group, int32(len(group)))
	}
	return p
}
func Ssw_ClassesPreparePairsSimple(g *aig.Graph, pairs *vec.Vec_Int_t) *Ssw_Cla_t {
	p := Ssw_ClassesStart(g)
	if vec.Vec_IntSize(pairs)%2 != 0 {
		panic("ssw: incomplete pair")
	}
	for i := 0; i < vec.Vec_IntSize(pairs); i += 2 {
		a, b := vec.Vec_IntEntry(pairs, i), vec.Vec_IntEntry(pairs, i+1)
		if a < 0 || a >= b {
			panic("ssw: pair must be ordered")
		}
		repr, obj := aig.Edge(a)<<1, aig.Edge(b)<<1
		aig.Aig_ObjSetRepr(g, obj, repr)
		Ssw_ObjAddClass(p, repr, []aig.Edge{repr, obj}, 2)
	}
	return p
}
