package fra

// Bounded model checking of initialized frames, ported from fraBmc.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
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
	framgr.Fra_BmcPerform(p, prefix, depth)
}
func Fra_BmcPerformSimple(g *aig.Graph, frames, limit, rewrite, verbose int32) {
	framgr.Fra_BmcPerformSimple(g, frames, limit, rewrite, verbose)
}
