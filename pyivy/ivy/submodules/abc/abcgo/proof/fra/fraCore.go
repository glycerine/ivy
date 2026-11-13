package fra

// FRAIG sweep and combinational manager, ported from fraCore.c.
import (
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
