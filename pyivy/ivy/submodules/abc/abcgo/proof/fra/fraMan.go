package fra

// Manager lifecycle is implemented by the canonical shared FRA manager.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/framgr"
)

func Fra_ManStart(graph *aig.Graph, pars *Fra_Par_t) *Fra_Man_t {
	return framgr.Fra_ManStart(graph, pars)
}
func Fra_ManClean(p *Fra_Man_t, nodesMax int32)  { framgr.Fra_ManClean(p, nodesMax) }
func Fra_ManPrepareComb(p *Fra_Man_t) *aig.Graph { return framgr.Fra_ManPrepareComb(p) }
func Fra_ManFinalizeComb(p *Fra_Man_t)           { framgr.Fra_ManFinalizeComb(p) }
func Fra_ManStop(p *Fra_Man_t) []aig.Edge        { return framgr.Fra_ManStop(p) }
func Fra_ManPrint(p *Fra_Man_t)                  { framgr.Fra_ManPrint(p) }
