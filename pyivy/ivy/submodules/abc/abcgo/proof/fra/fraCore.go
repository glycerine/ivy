package fra

// FRAIG sweep and combinational manager, ported from fraCore.c.
import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Fra_FraigMiterStatus(g *aig.Graph) int32 {
	if g.Data != nil {
		return 0
	}
	inputs := g.Inputs()
	regs := int(g.RegisterCount())
	pis := make(map[int]bool, len(inputs)-regs)
	for _, e := range inputs[:len(inputs)-regs] {
		pis[e.NodeID()] = true
	}
	outputs := g.Outputs()
	undecided := false
	for _, out := range outputs[:len(outputs)-regs] {
		child := g.Node(out).Fanins[0]
		if child == g.Const0() {
			continue
		}
		if child == g.Const1() || pis[child.NodeID()] || g.Node(child).Phase != child.Complemented() {
			return 0
		}
		undecided = true
	}
	if undecided {
		return -1
	}
	return 1
}
func Fra_FraigMiterAssertedOutput(g *aig.Graph) int32 {
	outputs := g.Outputs()
	regs := int(g.RegisterCount())
	for i, out := range outputs[:len(outputs)-regs] {
		child := g.Node(out).Fanins[0]
		if child != g.Const0() && (child == g.Const1() || g.Node(child).Phase != child.Complemented()) {
			return int32(i)
		}
	}
	return -1
}

// Fra_FraigVerifyCounterEx reports candidate classes disproved by a pattern.
// The original used shared object mark bits; local values avoid mutating the AIG.
func Fra_FraigVerifyCounterEx(p *Fra_Man_t, cex *vec.Vec_Int_t) {
	g := p.pManAig
	inputs := g.Inputs()
	if vec.Vec_IntSize(cex) != len(inputs) {
		panic("fra: incorrect counterexample input count")
	}
	values := make([]bool, len(g.Nodes()))
	values[0] = true
	for i, e := range inputs {
		values[e.NodeID()] = vec.Vec_IntEntry(cex, i) != 0
	}
	value := func(e aig.Edge) bool { return values[e.NodeID()] != e.Complemented() }
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		switch n.Kind {
		case aig.And:
			values[e.NodeID()] = value(n.Fanins[0]) && value(n.Fanins[1])
		case aig.ExclusiveOr:
			values[e.NodeID()] = value(n.Fanins[0]) != value(n.Fanins[1])
		case aig.Buffer, aig.Output:
			values[e.NodeID()] = value(n.Fanins[0])
		}
	}
	for _, id := range p.pCla.constants {
		if g.Node(aig.Edge(id<<1)).Phase != values[id] {
			fmt.Printf("The node %d is not constant under cex!\n", id)
		}
	}
	for _, class := range p.pCla.classes {
		for _, id := range class[1:] {
			first := class[0]
			if (g.Node(aig.Edge(first<<1)).Phase != g.Node(aig.Edge(id<<1)).Phase) != (values[first] != values[id]) {
				fmt.Printf("The nodes %d and %d are not equal under cex!\n", first, id)
			}
		}
	}
}

func Fra_FraigNode(p *Fra_Man_t, obj aig.Edge) {
	repr := Fra_ClassObjRepr(p, obj)
	if repr == aig.InvalidEdge || (p.pPars.fDoSparse == 0 && repr == p.pManAig.Const1()) {
		return
	}
	frame := p.pPars.nFramesK
	node, repNode := Fra_ObjFraig(p, obj, frame), Fra_ObjFraig(p, repr, frame)
	if node == aig.InvalidEdge || repNode == aig.InvalidEdge {
		panic("fra: missing FRAIG node mapping")
	}
	if aig.Aig_Regular(node) == aig.Aig_Regular(repNode) {
		p.nSatCallsSkipped++
		return
	}
	result := Fra_NodesAreEquiv(p, aig.Aig_Regular(repNode), aig.Aig_Regular(node))
	if result == 1 {
		if p.pManAig.Node(obj).Phase != p.pManAig.Node(repr).Phase {
			repNode = repNode.Not()
		}
		Fra_ObjSetFraig(p, obj, frame, repNode)
		return
	}
	if result == -1 {
		if p.vTimeouts == nil {
			p.vTimeouts = vec.Vec_PtrAlloc(100)
		}
		vec.Vec_PtrPush(p.vTimeouts, obj)
		if p.pPars.fSpeculate != 0 {
			panic("fra: speculation after SAT timeout is unsupported by the original assertion")
		}
		return
	}
	p.pCla.fRefinement = 1
	if p.vTimeouts != nil {
		vec.Vec_PtrPush(p.vTimeouts, obj)
	}
	Fra_SmlResimulate(p)
	if p.pManFraig.Data != nil {
		return
	}
	if p.pPars.nFramesK == 0 && Fra_ClassObjRepr(p, obj) == repr {
		panic("fra: counterexample did not refine node class")
	}
}
func Fra_FraigSweep(p *Fra_Man_t) {
	pos := int32(0)
	inputs := p.pManAig.Inputs()
	regs := int(p.pManAig.RegisterCount())
	for _, obj := range inputs[len(inputs)-regs:] {
		Fra_FraigNode(p, obj)
		if p.pPars.fUseImps != 0 {
			pos = Fra_ImpCheckForNode(p, p.pCla.vImps, obj, pos)
		}
	}
	if p.pPars.fLatchCorr != 0 {
		return
	}
	limit := p.pPars.nBTLimitNode
	for _, obj := range p.pManAig.LiveEdges() {
		n := p.pManAig.Node(obj)
		var mapped aig.Edge
		switch n.Kind {
		case aig.And:
			mapped = p.pManFraig.AND(Fra_ObjChild0Fra(p, obj, p.pPars.nFramesK), Fra_ObjChild1Fra(p, obj, p.pPars.nFramesK))
		case aig.ExclusiveOr:
			mapped = p.pManFraig.NewXor(Fra_ObjChild0Fra(p, obj, p.pPars.nFramesK), Fra_ObjChild1Fra(p, obj, p.pPars.nFramesK))
		case aig.Buffer:
			mapped = Fra_ObjChild0Fra(p, obj, p.pPars.nFramesK)
		default:
			continue
		}
		Fra_ObjSetFraig(p, obj, p.pPars.nFramesK, mapped)
		if p.pManFraig.Data != nil {
			continue
		}
		if p.pPars.nLevelMax != 0 && int32(n.Level) > p.pPars.nLevelMax {
			p.pPars.nBTLimitNode = 5
		}
		Fra_FraigNode(p, obj)
		p.pPars.nBTLimitNode = limit
		if p.pPars.fUseImps != 0 {
			pos = Fra_ImpCheckForNode(p, p.pCla.vImps, obj, pos)
		}
	}
	p.nNodesMiter = aig.Aig_ManNodeNum(p.pManFraig)
	if p.pPars.fUseImps != 0 {
		Fra_ImpCompactArray(p.pCla.vImps)
	}
}

func fraFraigPerformMapped(g *aig.Graph, pars *Fra_Par_t, callback func(*Fra_Man_t)) (*aig.Graph, []aig.Edge) {
	if g == nil || pars == nil {
		panic("fra: nil FRAIG input")
	}
	if aig.Aig_ManNodeNum(g) == 0 && callback == nil {
		result := aig.Aig_ManDupOrdered(g)
		mapping := make([]aig.Edge, len(g.Nodes()))
		for i := range mapping {
			mapping[i] = aig.Aig_ObjCopy(g, aig.Edge(i<<1))
		}
		return result, mapping
	}
	start := time.Now()
	p := Fra_ManStart(g, pars)
	p.pManFraig = Fra_ManPrepareComb(p)
	p.pSml = Fra_SmlStart(g, 0, 1, pars.nSimWords)
	Fra_SmlSimulate(p, 0)
	p.nLitsBeg = Fra_ClassesCountLits(p.pCla)
	p.nNodesBeg = aig.Aig_ManNodeNum(g)
	p.nRegsBeg = g.RegisterCount()
	if pars.fVerbose != 0 {
		Fra_ClassesPrint(p.pCla, 1)
	}
	Fra_FraigSweep(p)
	if callback != nil {
		callback(p)
	}
	Fra_ManFinalizeComb(p)
	Fra_ClassesCopyReprs(p.pCla, p.vTimeouts)
	var result *aig.Graph
	if pars.fChoicing != 0 {
		traversal := time.Now()
		result = aig.Aig_ManDupRepr(g, 1)
		aig.Aig_ManReprStart(result, int32(len(result.Nodes())))
		aig.Aig_ManTransferRepr(result, g)
		aig.Aig_ManMarkValidChoices(result)
		p.timeTrav += time.Since(traversal)
	} else {
		result = p.pManFraig
		aig.Aig_ManCleanup(result)
	}
	p.pManFraig = nil
	p.timeTotal = time.Since(start)
	p.nLitsEnd = Fra_ClassesCountLits(p.pCla)
	p.nNodesEnd = aig.Aig_ManNodeNum(result)
	p.nRegsEnd = result.RegisterCount()
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
	pars.nBTLimitNode = limit
	pars.fChoicing = 1
	pars.fDoSparse = 1
	pars.fDontShowBar = 1
	pars.nLevelMax = level
	return Fra_FraigPerform(g, &pars)
}
func Fra_FraigEquivenceMapped(g *aig.Graph, limit, prove int32, callback func(*Fra_Man_t)) (*aig.Graph, []aig.Edge) {
	var pars Fra_Par_t
	Fra_ParamsDefault(&pars)
	pars.nBTLimitNode = limit
	pars.fDoSparse = 1
	pars.fProve = prove
	pars.fDontShowBar = 1
	return fraFraigPerformMapped(g, &pars, callback)
}
func Fra_FraigEquivence(g *aig.Graph, limit, prove int32) *aig.Graph {
	result, _ := Fra_FraigEquivenceMapped(g, limit, prove, nil)
	return result
}
