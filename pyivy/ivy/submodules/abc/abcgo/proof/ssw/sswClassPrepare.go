package ssw

import "github.com/glycerine/abcgo/aig/aig"

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
