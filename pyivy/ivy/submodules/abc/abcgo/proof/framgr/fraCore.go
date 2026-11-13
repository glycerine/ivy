package framgr

// FRAIG sweep and combinational manager, ported from fraCore.c.
import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/fracomb"
)

func Fra_FraigMiterStatus(g *aig.Graph) int32         { return fracomb.Fra_FraigMiterStatus(g) }
func Fra_FraigMiterAssertedOutput(g *aig.Graph) int32 { return fracomb.Fra_FraigMiterAssertedOutput(g) }

// Fra_FraigVerifyCounterEx reports candidate classes disproved by a pattern.
// The original used shared object mark bits; local values avoid mutating the AIG.
func Fra_FraigVerifyCounterEx(p *Fra_Man_t, cex *vec.Vec_Int_t) {
	g := p.PManAig
	inputs := g.Inputs()
	if vec.Vec_IntSize(cex) != len(inputs) {
		panic("fra: incorrect counterexample input count")
	}
	for _, e := range g.LiveEdges() {
		if g.Node(e).MarkB {
			panic("fra: counterexample simulation marks must be clear")
		}
	}
	values := make([]bool, len(g.Nodes()))
	values[0] = true
	for i, e := range inputs {
		values[e.NodeID()] = vec.Vec_IntEntry(cex, i)&1 != 0
	}
	value := func(e aig.Edge) bool { return values[e.NodeID()] != e.Complemented() }
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			values[e.NodeID()] = value(n.Fanins[0]) && value(n.Fanins[1])
		case aig.Buffer, aig.Output:
			values[e.NodeID()] = value(n.Fanins[0])
		}
	}
	for _, id := range p.PCla.Constants {
		if g.Node(aig.Edge(id<<1)).Phase != values[id] {
			fmt.Printf("The node %d is not constant under cex!\n", id)
		}
	}
	for _, class := range p.PCla.Classes {
		for _, id := range class[1:] {
			first := class[0]
			if (g.Node(aig.Edge(first<<1)).Phase != g.Node(aig.Edge(id<<1)).Phase) != (values[first] != values[id]) {
				fmt.Printf("The nodes %d and %d are not equal under cex!\n", first, id)
			}
		}
	}
}

func Fra_FraigNode(p *Fra_Man_t, obj aig.Edge) {
	assert(!obj.Complemented(), "!Aig_IsComplement(pObj)")
	repr := Fra_ClassObjRepr(p, obj)
	if repr == aig.InvalidEdge || (p.PPars.FDoSparse == 0 && repr == p.PManAig.Const1()) {
		return
	}
	frame := p.PPars.NFramesK
	node, repNode := Fra_ObjFraig(p, obj, frame), Fra_ObjFraig(p, repr, frame)
	if node == aig.InvalidEdge || repNode == aig.InvalidEdge {
		panic("fra: missing FRAIG node mapping")
	}
	if aig.Aig_Regular(node) == aig.Aig_Regular(repNode) {
		p.NSatCallsSkipped++
		return
	}
	assert(p.PPars.NFramesK != 0 || aig.Aig_Regular(node) != p.PManFraig.Const1(), "nFramesK || Aig_Regular(pObjFraig) != pConst1")
	result := Fra_NodesAreEquiv(p, aig.Aig_Regular(repNode), aig.Aig_Regular(node))
	if result == 1 {
		if p.PManAig.Node(obj).Phase != p.PManAig.Node(repr).Phase {
			repNode = repNode.Not()
		}
		Fra_ObjSetFraig(p, obj, frame, repNode)
		return
	}
	if result == -1 {
		if p.VTimeouts == nil {
			p.VTimeouts = vec.Vec_PtrAlloc(100)
		}
		vec.Vec_PtrPush(p.VTimeouts, obj)
		if p.PPars.FSpeculate != 0 {
			panic("fra: speculation after SAT timeout is unsupported by the original assertion")
		}
		return
	}
	p.PCla.FRefinement = 1
	if p.VTimeouts != nil {
		vec.Vec_PtrPush(p.VTimeouts, obj)
	}
	Fra_SmlResimulate(p)
	if p.PManFraig.Data != nil {
		return
	}
	if p.PPars.NFramesK == 0 && Fra_ClassObjRepr(p, obj) == repr {
		panic("fra: counterexample did not refine node class")
	}
}
func Fra_FraigSweep(p *Fra_Man_t) {
	pos := int32(0)
	inputs := p.PManAig.Inputs()
	regs := int(p.PManAig.RegisterCount())
	for _, obj := range inputs[len(inputs)-regs:] {
		Fra_FraigNode(p, obj)
		if p.PPars.FUseImps != 0 {
			pos = Fra_ImpCheckForNode(p, p.PCla.VImps, obj, pos)
		}
	}
	if p.PPars.FLatchCorr != 0 {
		return
	}
	limit := p.PPars.NBTLimitNode
	for _, obj := range p.PManAig.LiveEdges() {
		n := p.PManAig.Node(obj)
		var mapped aig.Edge
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			mapped = p.PManFraig.AND(Fra_ObjChild0Fra(p, obj, p.PPars.NFramesK), Fra_ObjChild1Fra(p, obj, p.PPars.NFramesK))
		case aig.Buffer:
			mapped = Fra_ObjChild0Fra(p, obj, p.PPars.NFramesK)
		default:
			continue
		}
		Fra_ObjSetFraig(p, obj, p.PPars.NFramesK, mapped)
		if p.PManFraig.Data != nil {
			continue
		}
		if p.PPars.NLevelMax != 0 && int32(n.Level) > p.PPars.NLevelMax {
			p.PPars.NBTLimitNode = 5
		}
		Fra_FraigNode(p, obj)
		p.PPars.NBTLimitNode = limit
		if p.PPars.FUseImps != 0 {
			pos = Fra_ImpCheckForNode(p, p.PCla.VImps, obj, pos)
		}
	}
	p.NNodesMiter = aig.Aig_ManNodeNum(p.PManFraig)
	if p.PPars.FUseImps != 0 {
		Fra_ImpCompactArray(p.PCla.VImps)
	}
}

func fraFraigPerformMapped(g *aig.Graph, pars *Fra_Par_t, callback func(*Fra_Man_t)) (*aig.Graph, []aig.Edge) {
	if g == nil || pars == nil {
		panic("fra: nil FRAIG input")
	}
	// ABC returns before invoking the implication callback on this path.
	if aig.Aig_ManNodeNum(g) == 0 {
		result := aig.Aig_ManDupOrdered(g)
		mapping := make([]aig.Edge, len(g.Nodes()))
		for i := range mapping {
			mapping[i] = aig.Aig_ObjCopy(g, aig.Edge(i<<1))
		}
		return result, mapping
	}
	start := time.Now()
	p := Fra_ManStart(g, pars)
	p.PManFraig = Fra_ManPrepareComb(p)
	p.PSml = Fra_SmlStart(g, 0, 1, pars.NSimWords)
	Fra_SmlSimulate(p, 0)
	p.NLitsBeg = Fra_ClassesCountLits(p.PCla)
	p.NNodesBeg = aig.Aig_ManNodeNum(g)
	p.NRegsBeg = g.RegisterCount()
	if pars.FVerbose != 0 {
		Fra_ClassesPrint(p.PCla, 1)
	}
	Fra_FraigSweep(p)
	if callback != nil {
		callback(p)
	}
	Fra_ManFinalizeComb(p)
	Fra_ClassesCopyReprs(p.PCla, p.VTimeouts)
	var result *aig.Graph
	if pars.FChoicing != 0 {
		traversal := time.Now()
		result = aig.Aig_ManDupRepr(g, 1)
		aig.Aig_ManReprStart(result, int32(len(result.Nodes())))
		aig.Aig_ManTransferRepr(result, g)
		aig.Aig_ManMarkValidChoices(result)
		p.TimeTrav += time.Since(traversal)
	} else {
		result = p.PManFraig
		aig.Aig_ManCleanup(result)
	}
	p.PManFraig = nil
	p.TimeTotal = time.Since(start)
	p.NLitsEnd = Fra_ClassesCountLits(p.PCla)
	p.NNodesEnd = aig.Aig_ManNodeNum(result)
	p.NRegsEnd = result.RegisterCount()
	mapping := Fra_ManStop(p)
	return result, mapping
}
func Fra_FraigPerform(g *aig.Graph, pars *Fra_Par_t) *aig.Graph {
	result, _ := fraFraigPerformMapped(g, pars, nil)
	return result
}
func Fra_FraigChoice(g *aig.Graph, limit, level int32) *aig.Graph {
	var pars Fra_Par_t
	Fra_ParamsDefault(&pars)
	pars.NBTLimitNode = limit
	pars.FChoicing = 1
	pars.FDoSparse = 1
	pars.FDontShowBar = 1
	pars.NLevelMax = level
	return Fra_FraigPerform(g, &pars)
}
func Fra_FraigEquivenceMapped(g *aig.Graph, limit, prove int32, callback func(*Fra_Man_t)) (*aig.Graph, []aig.Edge) {
	var pars Fra_Par_t
	Fra_ParamsDefault(&pars)
	pars.NBTLimitNode = limit
	pars.FDoSparse = 1
	pars.FProve = prove
	pars.FDontShowBar = 1
	return fraFraigPerformMapped(g, &pars, callback)
}
func Fra_FraigEquivence(g *aig.Graph, limit, prove int32) *aig.Graph {
	result, _ := Fra_FraigEquivenceMapped(g, limit, prove, nil)
	return result
}
