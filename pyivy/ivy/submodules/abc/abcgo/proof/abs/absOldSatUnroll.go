package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManUnrollWithCex builds the cone of the failing output over all frames.
// The first nInputs primary inputs are fixed to the counterexample; each
// remaining frame input is mapped to its original input and frame in mapping.
func Saig_ManUnrollWithCex(g *aig.Graph, cex *util.Abc_Cex_t, nInputs int32, mapping **vec.Vec_Int_t) *aig.Graph {
	piCount, regCount := saig.Saig_ManPiNum(g), saig.Saig_ManRegNum(g)
	if cex.NPis != piCount || cex.NRegs != regCount || cex.IPo < 0 || cex.IPo >= saig.Saig_ManPoNum(g) {
		panic("invalid unrolling counterexample")
	}
	nFrames := int(cex.IFrame) + 1
	roots, objects := make([][]aig.Edge, nFrames), make([][]aig.Edge, nFrames)
	roots[nFrames-1] = append(roots[nFrames-1], g.Outputs()[cex.IPo])
	loToLi := make(map[int]aig.Edge)
	inputIndex := make(map[int]int32)
	for i, e := range g.Inputs() {
		inputIndex[e.NodeID()] = int32(i)
	}
	for i := int32(0); i < regCount; i++ {
		loToLi[saig.Saig_ManLo(g, i).NodeID()] = saig.Saig_ManLi(g, i)
	}
	for frame := nFrames - 1; frame >= 0; frame-- {
		seen := make([]bool, len(g.Nodes()))
		var visit func(aig.Edge)
		visit = func(e aig.Edge) {
			e = aig.Aig_Regular(e)
			id := e.NodeID()
			if seen[id] {
				return
			}
			seen[id] = true
			n := g.Node(e)
			switch n.Kind {
			case aig.Output, aig.Buffer:
				visit(n.Fanins[0])
			case aig.And:
				visit(n.Fanins[0])
				visit(n.Fanins[1])
			case aig.ExclusiveOr:
				panic("unrolling expects an AND-inverter graph")
			}
			if frame > 0 {
				if li, ok := loToLi[id]; ok {
					roots[frame-1] = append(roots[frame-1], li)
				}
			}
			objects[frame] = append(objects[frame], e)
		}
		for _, root := range roots[frame] {
			visit(root)
		}
	}
	result := aig.NewGraph()
	result.Name = g.Name
	result.Spec = g.Spec
	*mapping = vec.Vec_IntAlloc(1000)
	copies := make([]aig.Edge, len(g.Nodes()))
	copies[g.Const1().NodeID()] = result.Const1()
	for i := int32(0); i < regCount; i++ {
		v := result.Const0()
		if util.Abc_InfoHasBit(cex.PData, i) {
			v = result.Const1()
		}
		copies[saig.Saig_ManLo(g, i).NodeID()] = v
	}
	copyEdge := func(e aig.Edge) aig.Edge {
		v := copies[e.NodeID()]
		if e.Complemented() {
			v = v.Not()
		}
		return v
	}
	for frame := 0; frame < nFrames; frame++ {
		for _, e := range objects[frame] {
			n := g.Node(e)
			switch n.Kind {
			case aig.And:
				copies[e.NodeID()] = result.AND(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
			case aig.Output, aig.Buffer:
				copies[e.NodeID()] = copyEdge(n.Fanins[0])
			case aig.Input:
				idx := inputIndex[e.NodeID()]
				if idx >= piCount {
					continue
				}
				if idx < nInputs {
					bit := cex.NRegs + int32(frame)*cex.NPis + idx
					copies[e.NodeID()] = result.Const0()
					if util.Abc_InfoHasBit(cex.PData, bit) {
						copies[e.NodeID()] = result.Const1()
					}
				} else {
					copies[e.NodeID()] = result.NewInput()
					vec.Vec_IntPush(*mapping, idx)
					vec.Vec_IntPush(*mapping, int32(frame))
				}
			}
		}
		if frame < nFrames-1 {
			for _, li := range roots[frame] {
				lo := saig.Saig_ObjLiToLo(g, li)
				copies[lo.NodeID()] = copies[li.NodeID()]
			}
		}
	}
	result.AddOutput(copies[g.Outputs()[cex.IPo].NodeID()].Not())
	return result
}

func Saig_RefManStart(g *aig.Graph, cex *util.Abc_Cex_t, nInputs, verbose int32) *Saig_RefMan_t {
	p := &Saig_RefMan_t{pAig: g, pCex: cex, nInputs: nInputs, fVerbose: verbose}
	p.pFrames = Saig_ManUnrollWithCex(g, cex, nInputs, &p.vMapPiF2A)
	return p
}
func Saig_RefManStop(p *Saig_RefMan_t) { p.pFrames = nil; p.vMapPiF2A = nil; p.phase = nil }
