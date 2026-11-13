package fra

// Bounded model checking of initialized frames, ported from fraBmc.c.
import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/proof/framgr"
)

// Fra_Bmc_t_ is the shared bounded-model-checking state.
type Fra_Bmc_t_ = framgr.BMC

// BMC helpers are shared with the combinational FRA manager.
func Bmc_ObjFrames(b *Fra_Bmc_t, obj aig.Edge, frame int32) aig.Edge {
	return framgr.Bmc_ObjFrames(b, obj, frame)
}
func Bmc_ObjSetFrames(b *Fra_Bmc_t, obj aig.Edge, frame int32, value aig.Edge) {
	framgr.Bmc_ObjSetFrames(b, obj, frame, value)
}
func Bmc_ObjFraig(b *Fra_Bmc_t, obj aig.Edge) aig.Edge  { return framgr.Bmc_ObjFraig(b, obj) }
func Bmc_ObjSetFraig(b *Fra_Bmc_t, obj, value aig.Edge) { framgr.Bmc_ObjSetFraig(b, obj, value) }
func Bmc_ObjChild0Frames(b *Fra_Bmc_t, obj aig.Edge, frame int32) aig.Edge {
	return framgr.Bmc_ObjChild0Frames(b, obj, frame)
}
func Bmc_ObjChild1Frames(b *Fra_Bmc_t, obj aig.Edge, frame int32) aig.Edge {
	return framgr.Bmc_ObjChild1Frames(b, obj, frame)
}
func Fra_BmcNodesAreEqual(p *Fra_Man_t, a, b aig.Edge) int32 {
	return framgr.Fra_BmcNodesAreEqual(p, a, b)
}
func Fra_BmcNodeIsConst(p *Fra_Man_t, obj aig.Edge) int32 {
	return framgr.Fra_BmcNodeIsConst(p, obj)
}
func Fra_BmcFilterImplications(p *Fra_Man_t, b *Fra_Bmc_t) {
	framgr.Fra_BmcFilterImplications(p, b)
}
func Fra_BmcStart(g *aig.Graph, prefix, depth int32) *Fra_Bmc_t {
	return framgr.Fra_BmcStart(g, prefix, depth)
}
func Fra_BmcStop(b *Fra_Bmc_t) { framgr.Fra_BmcStop(b) }
func Fra_BmcFrames(b *Fra_Bmc_t, keepOutputs int32) *aig.Graph {
	return framgr.Fra_BmcFrames(b, keepOutputs)
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
	oldImps := 0
	if p.PCla.VImps != nil {
		b.VImps = p.PCla.VImps
		oldImps = vec.Vec_IntSize(b.VImps)
	}
	if b.VImps != nil {
		b.PAigFraig, b.PObjToFraig = framgr.SweepBMCImplications(b.PAigFrames, b.PAig, b.NPref, b.NDepth, b.PObjToFrames, b.VImps, 1000000)
	} else {
		b.PAigFraig, b.PObjToFraig = Fra_FraigEquivenceMapped(b.PAigFrames, 1000000, 0, nil)
	}
	if b.PAigFraig == nil {
		panic("fra: BMC frame FRAIG failed")
	}
	if p.PPars.FVerbose != 0 {
		fmt.Printf("Original AIG = %d. Init %d frames = %d. Fraig = %d. Time = %s.\nBefore BMC: Const = %d. Class = %d. Lit = %d. Imp = %d.\n", aig.Aig_ManNodeNum(b.PAig), b.NFramesAll, aig.Aig_ManNodeNum(b.PAigFrames), aig.Aig_ManNodeNum(b.PAigFraig), time.Since(start), len(p.PCla.Constants), len(p.PCla.Classes), Fra_ClassesCountLits(p.PCla), oldImps)
	}
	prevConst, prevEqual := p.PCla.NodeIsConst, p.PCla.NodesEqual
	p.PCla.NodeIsConst = func(id int) bool { return Fra_BmcNodeIsConst(p, aig.Edge(id<<1)) != 0 }
	p.PCla.NodesEqual = func(a, b int) bool { return Fra_BmcNodesAreEqual(p, aig.Edge(a<<1), aig.Edge(b<<1)) != 0 }
	defer func() { p.PCla.NodeIsConst, p.PCla.NodesEqual = prevConst, prevEqual }()
	Fra_ClassesRefine(p.PCla)
	Fra_ClassesRefine1(p.PCla, 1, nil)
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
		// A positive input output needs a witness even when FRAIG returns
		// early for a graph without internal nodes.
		inputs := b.PAigFrames.Inputs()
		for _, out := range b.PAigFrames.Outputs() {
			child := b.PAigFrames.Node(out).Fanins[0]
			if child.Complemented() || b.PAigFrames.Node(child).Kind != aig.Input {
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
			cex = Fra_SmlCopyCounterExample(g, b.PAigFrames, model)
			if cex != nil {
				break
			}
		}
		if cex == nil {
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
