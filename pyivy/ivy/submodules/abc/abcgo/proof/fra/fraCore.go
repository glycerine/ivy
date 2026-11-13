package fra

// FRAIG sweep and combinational manager, ported from fraCore.c.
import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/fracomb"
	"github.com/glycerine/abcgo/proof/framgr"
)

func Fra_FraigMiterStatus(g *aig.Graph) int32         { return fracomb.Fra_FraigMiterStatus(g) }
func Fra_FraigMiterAssertedOutput(g *aig.Graph) int32 { return fracomb.Fra_FraigMiterAssertedOutput(g) }

// Fra_FraigVerifyCounterEx reports candidate classes disproved by a pattern.
// The original used shared object mark bits; local values avoid mutating the AIG.
func Fra_FraigVerifyCounterEx(p *Fra_Man_t, cex *vec.Vec_Int_t) {
	framgr.Fra_FraigVerifyCounterEx(p, cex)
}

func Fra_FraigNode(p *Fra_Man_t, obj aig.Edge) { framgr.Fra_FraigNode(p, obj) }
func Fra_FraigSweep(p *Fra_Man_t)              { framgr.Fra_FraigSweep(p) }

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
	if g == nil || pars == nil {
		panic("fra: nil FRAIG input")
	}
	return framgr.Sweep(g, *pars)
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
	return framgr.SweepMappedCallback(g, pars, callback)
}
func Fra_FraigEquivence(g *aig.Graph, limit, prove int32) *aig.Graph {
	result, _ := Fra_FraigEquivenceMapped(g, limit, prove, nil)
	return result
}
