package framgr

// Bounded model checking of initialized frames, ported from fraBmc.c.
import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

type Fra_Bmc_t_ struct {
	NPref, NDepth, NFramesAll   int32
	VImps                       *vec.Vec_Int_t
	PAig, PAigFrames, PAigFraig *aig.Graph
	PObjToFrames, PObjToFraig   []aig.Edge
}

func Bmc_ObjFrames(b *Fra_Bmc_t, obj aig.Edge, frame int32) aig.Edge {
	if frame < 0 || frame >= b.NFramesAll {
		panic("fra: BMC frame out of bounds")
	}
	value := b.PObjToFrames[obj.NodeID()*int(b.NFramesAll)+int(frame)]
	if value == aig.InvalidEdge {
		panic("fra: missing BMC frame mapping")
	}
	return value
}
func Bmc_ObjSetFrames(b *Fra_Bmc_t, obj aig.Edge, frame int32, value aig.Edge) {
	if frame < 0 || frame >= b.NFramesAll {
		panic("fra: BMC frame out of bounds")
	}
	b.PObjToFrames[obj.NodeID()*int(b.NFramesAll)+int(frame)] = value
}
func Bmc_ObjFraig(b *Fra_Bmc_t, obj aig.Edge) aig.Edge {
	value := b.PObjToFraig[obj.NodeID()]
	if value == aig.InvalidEdge {
		panic("fra: missing BMC fraig mapping")
	}
	return value
}
func Bmc_ObjSetFraig(b *Fra_Bmc_t, obj, value aig.Edge) { b.PObjToFraig[obj.NodeID()] = value }
func bmcChildFrames(b *Fra_Bmc_t, obj aig.Edge, frame int32, index int) aig.Edge {
	child := b.PAig.Node(obj).Fanins[index]
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
	m := p.PBmc
	for frame := m.NPref; frame < m.NFramesAll; frame++ {
		left, right := aig.Aig_Regular(Bmc_ObjFrames(m, a, frame)), aig.Aig_Regular(Bmc_ObjFrames(m, b, frame))
		if left != right && aig.Aig_Regular(Bmc_ObjFraig(m, left)) != aig.Aig_Regular(Bmc_ObjFraig(m, right)) {
			return 0
		}
	}
	return 1
}
func Fra_BmcNodeIsConst(p *Fra_Man_t, obj aig.Edge) int32 {
	return Fra_BmcNodesAreEqual(p, obj, p.PManAig.Const1())
}

func Fra_BmcFilterImplications(p *Fra_Man_t, b *Fra_Bmc_t) {
	if p.NFramesAll != 1 || p.PManAig != b.PAigFrames {
		panic("fra: invalid BMC implication manager")
	}
	for i := 0; i < vec.Vec_IntSize(b.VImps); i++ {
		imp := vec.Vec_IntEntry(b.VImps, i)
		if imp == 0 {
			continue
		}
		left, right := aig.Edge(Fra_ImpLeft(imp)<<1), aig.Edge(Fra_ImpRight(imp)<<1)
		for frame := b.NPref; frame < b.NFramesAll; frame++ {
			lt, rt := Bmc_ObjFrames(b, left, frame), Bmc_ObjFrames(b, right, frame)
			lf, rf := Fra_ObjFraig(p, aig.Aig_Regular(lt), 0), Fra_ObjFraig(p, aig.Aig_Regular(rt), 0)
			if lf == aig.InvalidEdge || rf == aig.InvalidEdge {
				panic("fra: missing implication mapping")
			}
			lc := b.PAig.Node(left).Phase != lt.Complemented() != lf.Complemented()
			rc := b.PAig.Node(right).Phase != rt.Complemented() != rf.Complemented()
			if aig.Aig_Regular(lf) == aig.Aig_Regular(rf) {
				if lc == rc || (lf.NodeID() == 0 && lc) {
					continue
				}
			} else if Fra_NodesAreImp(p, aig.Aig_Regular(lf), aig.Aig_Regular(rf), fraBool(lc), fraBool(rc)) == 1 {
				continue
			}
			vec.Vec_IntWriteEntry(b.VImps, i, 0)
			break
		}
	}
	Fra_ImpCompactArray(b.VImps)
}
func Fra_BmcStart(g *aig.Graph, prefix, depth int32) *Fra_Bmc_t {
	if g == nil || prefix < 0 || depth <= 0 {
		panic("fra: invalid BMC dimensions")
	}
	b := &Fra_Bmc_t{PAig: g, NPref: prefix, NDepth: depth, NFramesAll: prefix + depth}
	b.PObjToFrames = make([]aig.Edge, len(g.Nodes())*int(b.NFramesAll))
	for i := range b.PObjToFrames {
		b.PObjToFrames[i] = aig.InvalidEdge
	}
	return b
}
func Fra_BmcStop(_ *Fra_Bmc_t) {}
func Fra_BmcFrames(b *Fra_Bmc_t, keepOutputs int32) *aig.Graph {
	graph, mapping := fraUnrollInitialized(b.PAig, int(b.NFramesAll), keepOutputs != 0)
	b.PObjToFrames = mapping
	return graph
}
func Fra_BmcPerform(p *Fra_Man_t, prefix, depth int32) {
	if p.PBmc != nil {
		panic("fra: BMC already in progress")
	}
	start := time.Now()
	b := Fra_BmcStart(p.PManAig, prefix, depth)
	p.PBmc = b
	defer func() { Fra_BmcStop(b); p.PBmc = nil }()
	b.PAigFrames = Fra_BmcFrames(b, 0)
	var callback func(*Fra_Man_t)
	oldImps := 0
	if p.PCla.VImps != nil {
		b.VImps = p.PCla.VImps
		oldImps = vec.Vec_IntSize(b.VImps)
		callback = func(comb *Fra_Man_t) { Fra_BmcFilterImplications(comb, b) }
	}
	b.PAigFraig, b.PObjToFraig = Fra_FraigEquivenceMapped(b.PAigFrames, 1000000, 0, callback)
	if b.PAigFraig == nil {
		panic("fra: BMC frame FRAIG failed")
	}
	if p.PPars.FVerbose != 0 {
		fmt.Printf("Original AIG = %d. Init %d frames = %d. Fraig = %d. Time = %s.\nBefore BMC: Const = %d. Class = %d. Lit = %d. Imp = %d.\n", aig.Aig_ManNodeNum(b.PAig), b.NFramesAll, aig.Aig_ManNodeNum(b.PAigFrames), aig.Aig_ManNodeNum(b.PAigFraig), time.Since(start), len(p.PCla.Constants), len(p.PCla.Classes), Fra_ClassesCountLits(p.PCla), oldImps)
	}
	p.PCla.NodeIsConst = func(id int) bool { return Fra_BmcNodeIsConst(p, aig.Edge(id<<1)) != 0 }
	p.PCla.NodesEqual = func(a, b int) bool { return Fra_BmcNodesAreEqual(p, aig.Edge(a<<1), aig.Edge(b<<1)) != 0 }
	Fra_ClassesRefine(p.PCla)
	Fra_ClassesRefine1(p.PCla, 1, nil)
	// Nil selects the ordinary simulation callbacks, as the C driver does.
	p.PCla.NodeIsConst, p.PCla.NodesEqual = nil, nil
	if p.PPars.FVerbose != 0 {
		n := 0
		if b.VImps != nil {
			n = vec.Vec_IntSize(b.VImps)
		}
		fmt.Printf("After BMC: Const = %d. Class = %d. Lit = %d. Imp = %d.\n", len(p.PCla.Constants), len(p.PCla.Classes), Fra_ClassesCountLits(p.PCla), n)
	}
}
func Fra_BmcPerformSimple(g *aig.Graph, frames, limit, rewrite, verbose int32) {
	start := time.Now()
	b := Fra_BmcStart(g, 0, frames)
	defer Fra_BmcStop(b)
	b.PAigFrames = Fra_BmcFrames(b, 1)
	if verbose != 0 {
		fmt.Printf("AIG: PI/PO/Reg = %d/%d/%d. Node = %d.\nTime-frames (%d): PI/PO = %d/%d. Node = %d. Time = %s.\n", g.InputCount()-int(g.RegisterCount()), g.OutputCount()-int(g.RegisterCount()), g.RegisterCount(), aig.Aig_ManNodeNum(g), frames, b.PAigFrames.InputCount(), b.PAigFrames.OutputCount(), aig.Aig_ManNodeNum(b.PAigFrames), time.Since(start))
	}
	if rewrite != 0 {
		rewritten, err := dar.Dar_ManRwsat(b.PAigFrames, true, false)
		if err != nil {
			panic(err)
		}
		b.PAigFrames = rewritten
	}
	output := Fra_FraigMiterAssertedOutput(b.PAigFrames)
	var cex *util.Abc_Cex_t
	if output >= 0 {
		cex = util.Abc_CexMakeTriv(g.RegisterCount(), int32(g.InputCount())-g.RegisterCount(), int32(g.OutputCount())-g.RegisterCount(), output)
	} else {
		b.PAigFraig, _ = Fra_FraigEquivenceMapped(b.PAigFrames, limit, 1, nil)
		if b.PAigFraig == nil {
			panic("fra: BMC frame FRAIG failed")
		}
		output = Fra_FraigMiterAssertedOutput(b.PAigFraig)
		if model, ok := b.PAigFraig.Data.([]int32); ok {
			cex = Fra_SmlCopyCounterExample(g, b.PAigFrames, model)
			b.PAigFraig.Data = nil
		} else if output >= 0 {
			cex = util.Abc_CexMakeTriv(g.RegisterCount(), int32(g.InputCount())-g.RegisterCount(), int32(g.OutputCount())-g.RegisterCount(), output)
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
