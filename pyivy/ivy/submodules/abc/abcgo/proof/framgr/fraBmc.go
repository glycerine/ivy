package framgr

// Bounded model checking of initialized frames, ported from fraBmc.c.
import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dar"
)

type Fra_Bmc_t_ struct {
	nPref, nDepth, nFramesAll   int32
	vImps                       *vec.Vec_Int_t
	pAig, pAigFrames, pAigFraig *aig.Graph
	pObjToFrames, pObjToFraig   []aig.Edge
}

func Bmc_ObjFrames(b *Fra_Bmc_t, obj aig.Edge, frame int32) aig.Edge {
	if frame < 0 || frame >= b.nFramesAll {
		panic("fra: BMC frame out of bounds")
	}
	value := b.pObjToFrames[obj.NodeID()*int(b.nFramesAll)+int(frame)]
	if value == aig.InvalidEdge {
		panic("fra: missing BMC frame mapping")
	}
	return value
}
func Bmc_ObjSetFrames(b *Fra_Bmc_t, obj aig.Edge, frame int32, value aig.Edge) {
	if frame < 0 || frame >= b.nFramesAll {
		panic("fra: BMC frame out of bounds")
	}
	b.pObjToFrames[obj.NodeID()*int(b.nFramesAll)+int(frame)] = value
}
func Bmc_ObjFraig(b *Fra_Bmc_t, obj aig.Edge) aig.Edge {
	value := b.pObjToFraig[obj.NodeID()]
	if value == aig.InvalidEdge {
		panic("fra: missing BMC fraig mapping")
	}
	return value
}
func Bmc_ObjSetFraig(b *Fra_Bmc_t, obj, value aig.Edge) { b.pObjToFraig[obj.NodeID()] = value }
func bmcChildFrames(b *Fra_Bmc_t, obj aig.Edge, frame int32, index int) aig.Edge {
	child := b.pAig.Node(obj).Fanins[index]
	if child == aig.InvalidEdge {
		return child
	}
	value := Bmc_ObjFrames(b, child, frame)
	if child.Complemented() {
		value = value.Not()
	}
	return value
}
func Bmc_ObjChild0Frames(b *Fra_Bmc_t, obj aig.Edge, frame int32) aig.Edge {
	return bmcChildFrames(b, obj, frame, 0)
}
func Bmc_ObjChild1Frames(b *Fra_Bmc_t, obj aig.Edge, frame int32) aig.Edge {
	return bmcChildFrames(b, obj, frame, 1)
}

func Fra_BmcNodesAreEqual(p *Fra_Man_t, a, b aig.Edge) int32 {
	m := p.pBmc
	for frame := m.nPref; frame < m.nFramesAll; frame++ {
		left, right := aig.Aig_Regular(Bmc_ObjFrames(m, a, frame)), aig.Aig_Regular(Bmc_ObjFrames(m, b, frame))
		if left != right && aig.Aig_Regular(Bmc_ObjFraig(m, left)) != aig.Aig_Regular(Bmc_ObjFraig(m, right)) {
			return 0
		}
	}
	return 1
}
func Fra_BmcNodeIsConst(p *Fra_Man_t, obj aig.Edge) int32 {
	return Fra_BmcNodesAreEqual(p, obj, p.pManAig.Const1())
}

func Fra_BmcFilterImplications(p *Fra_Man_t, b *Fra_Bmc_t) {
	if p.nFramesAll != 1 || p.pManAig != b.pAigFrames {
		panic("fra: invalid BMC implication manager")
	}
	for i := 0; i < vec.Vec_IntSize(b.vImps); i++ {
		imp := vec.Vec_IntEntry(b.vImps, i)
		if imp == 0 {
			continue
		}
		left, right := aig.Edge(Fra_ImpLeft(imp)<<1), aig.Edge(Fra_ImpRight(imp)<<1)
		for frame := b.nPref; frame < b.nFramesAll; frame++ {
			lt, rt := Bmc_ObjFrames(b, left, frame), Bmc_ObjFrames(b, right, frame)
			lf, rf := Fra_ObjFraig(p, aig.Aig_Regular(lt), 0), Fra_ObjFraig(p, aig.Aig_Regular(rt), 0)
			if lf == aig.InvalidEdge || rf == aig.InvalidEdge {
				panic("fra: missing implication mapping")
			}
			lc := b.pAig.Node(left).Phase != lt.Complemented() != lf.Complemented()
			rc := b.pAig.Node(right).Phase != rt.Complemented() != rf.Complemented()
			if aig.Aig_Regular(lf) == aig.Aig_Regular(rf) {
				if lc == rc || (lf.NodeID() == 0 && lc) {
					continue
				}
			} else if Fra_NodesAreImp(p, aig.Aig_Regular(lf), aig.Aig_Regular(rf), fraBool(lc), fraBool(rc)) == 1 {
				continue
			}
			vec.Vec_IntWriteEntry(b.vImps, i, 0)
			break
		}
	}
	Fra_ImpCompactArray(b.vImps)
}
func Fra_BmcStart(g *aig.Graph, prefix, depth int32) *Fra_Bmc_t {
	if g == nil || prefix < 0 || depth <= 0 {
		panic("fra: invalid BMC dimensions")
	}
	b := &Fra_Bmc_t{pAig: g, nPref: prefix, nDepth: depth, nFramesAll: prefix + depth}
	b.pObjToFrames = make([]aig.Edge, len(g.Nodes())*int(b.nFramesAll))
	for i := range b.pObjToFrames {
		b.pObjToFrames[i] = aig.InvalidEdge
	}
	return b
}
func Fra_BmcStop(_ *Fra_Bmc_t) {}
func Fra_BmcFrames(b *Fra_Bmc_t, keepOutputs int32) *aig.Graph {
	graph, mapping := fraUnrollInitialized(b.pAig, int(b.nFramesAll), keepOutputs != 0)
	b.pObjToFrames = mapping
	return graph
}
func Fra_BmcPerform(p *Fra_Man_t, prefix, depth int32) {
	if p.pBmc != nil {
		panic("fra: BMC already in progress")
	}
	start := time.Now()
	b := Fra_BmcStart(p.pManAig, prefix, depth)
	p.pBmc = b
	defer func() { Fra_BmcStop(b); p.pBmc = nil }()
	b.pAigFrames = Fra_BmcFrames(b, 0)
	var callback func(*Fra_Man_t)
	oldImps := 0
	if p.pCla.VImps != nil {
		b.vImps = p.pCla.VImps
		oldImps = vec.Vec_IntSize(b.vImps)
		callback = func(comb *Fra_Man_t) { Fra_BmcFilterImplications(comb, b) }
	}
	b.pAigFraig, b.pObjToFraig = Fra_FraigEquivenceMapped(b.pAigFrames, 1000000, 0, callback)
	if b.pAigFraig == nil {
		panic("fra: BMC frame FRAIG failed")
	}
	if p.pPars.FVerbose != 0 {
		fmt.Printf("Original AIG = %d. Init %d frames = %d. Fraig = %d. Time = %s.\nBefore BMC: Const = %d. Class = %d. Lit = %d. Imp = %d.\n", aig.Aig_ManNodeNum(b.pAig), b.nFramesAll, aig.Aig_ManNodeNum(b.pAigFrames), aig.Aig_ManNodeNum(b.pAigFraig), time.Since(start), len(p.pCla.Constants), len(p.pCla.Classes), Fra_ClassesCountLits(p.pCla), oldImps)
	}
	prevConst, prevEqual := p.pCla.NodeIsConst, p.pCla.NodesEqual
	p.pCla.NodeIsConst = func(id int) bool { return Fra_BmcNodeIsConst(p, aig.Edge(id<<1)) != 0 }
	p.pCla.NodesEqual = func(a, b int) bool { return Fra_BmcNodesAreEqual(p, aig.Edge(a<<1), aig.Edge(b<<1)) != 0 }
	defer func() { p.pCla.NodeIsConst, p.pCla.NodesEqual = prevConst, prevEqual }()
	Fra_ClassesRefine(p.pCla)
	Fra_ClassesRefine1(p.pCla, 1, nil)
	if p.pPars.FVerbose != 0 {
		n := 0
		if b.vImps != nil {
			n = vec.Vec_IntSize(b.vImps)
		}
		fmt.Printf("After BMC: Const = %d. Class = %d. Lit = %d. Imp = %d.\n", len(p.pCla.Constants), len(p.pCla.Classes), Fra_ClassesCountLits(p.pCla), n)
	}
}
func Fra_BmcPerformSimple(g *aig.Graph, frames, limit, rewrite, verbose int32) {
	start := time.Now()
	b := Fra_BmcStart(g, 0, frames)
	defer Fra_BmcStop(b)
	b.pAigFrames = Fra_BmcFrames(b, 1)
	if verbose != 0 {
		fmt.Printf("AIG: PI/PO/Reg = %d/%d/%d. Node = %d.\nTime-frames (%d): PI/PO = %d/%d. Node = %d. Time = %s.\n", g.InputCount()-int(g.RegisterCount()), g.OutputCount()-int(g.RegisterCount()), g.RegisterCount(), aig.Aig_ManNodeNum(g), frames, b.pAigFrames.InputCount(), b.pAigFrames.OutputCount(), aig.Aig_ManNodeNum(b.pAigFrames), time.Since(start))
	}
	if rewrite != 0 {
		rewritten, err := dar.Dar_ManRwsat(b.pAigFrames, true, false)
		if err != nil {
			panic(err)
		}
		b.pAigFrames = rewritten
	}
	output := Fra_FraigMiterAssertedOutput(b.pAigFrames)
	var cex *util.Abc_Cex_t
	if output >= 0 {
		cex = util.Abc_CexMakeTriv(g.RegisterCount(), int32(g.InputCount())-g.RegisterCount(), int32(g.OutputCount())-g.RegisterCount(), output)
	} else {
		// A positive input output needs a witness even when FRAIG returns
		// early for a graph without internal nodes.
		inputs := b.pAigFrames.Inputs()
		for _, out := range b.pAigFrames.Outputs() {
			child := b.pAigFrames.Node(out).Fanins[0]
			if child.Complemented() || b.pAigFrames.Node(child).Kind != aig.Input {
				continue
			}
			model := make([]int32, len(inputs)+1)
			for i, in := range inputs {
				if in.NodeID() == child.NodeID() {
					model[i] = 1
					break
				}
			}
			model[len(inputs)] = int32(out.NodeID())
			cex = Fra_SmlCopyCounterExample(g, b.pAigFrames, model)
			if cex != nil {
				break
			}
		}
		if cex == nil {
			b.pAigFraig, _ = Fra_FraigEquivenceMapped(b.pAigFrames, limit, 1, nil)
			if b.pAigFraig == nil {
				panic("fra: BMC frame FRAIG failed")
			}
			output = Fra_FraigMiterAssertedOutput(b.pAigFraig)
			if model, ok := b.pAigFraig.Data.([]int32); ok {
				cex = Fra_SmlCopyCounterExample(g, b.pAigFrames, model)
				b.pAigFraig.Data = nil
			} else if output >= 0 {
				cex = util.Abc_CexMakeTriv(g.RegisterCount(), int32(g.InputCount())-g.RegisterCount(), int32(g.OutputCount())-g.RegisterCount(), output)
			}
		}
	}
	if cex != nil {
		if err := g.SetSequentialCounterexample(cex); err != nil {
			panic(err)
		}
	}
	if verbose != 0 {
		fmt.Printf("Fraiged init frames: Time = %s.\n", time.Since(start))
	}
}
