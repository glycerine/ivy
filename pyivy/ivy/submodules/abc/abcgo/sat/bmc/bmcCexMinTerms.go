package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Saig_ManCexMinGetCos(g *aig.Graph, cex *util.Abc_Cex_t, leaves, roots *vec.Vec_Int_t) {
	if g == nil || cex == nil || roots == nil {
		panic("bmc: invalid counterexample roots")
	}
	vec.Vec_IntClear(roots)
	if leaves == nil {
		vec.Vec_IntPush(roots, int32(aig.Aig_ManCo(g, cex.IPo).NodeID()))
		return
	}
	for _, id := range vec.Vec_IntArray(leaves) {
		e := aig.Edge(id) << 1
		if saig.Saig_ObjIsLo(g, e) {
			vec.Vec_IntPush(roots, int32(saig.Saig_ObjLoToLi(g, e).NodeID()))
		}
	}
}
func Saig_ManCexMinCollectFrameTerms_rec(g *aig.Graph, e aig.Edge, terms *vec.Vec_Int_t) {
	e = aig.Aig_Regular(e)
	if aig.Aig_ObjIsTravIdCurrent(g, e) {
		return
	}
	aig.Aig_ObjSetTravIdCurrent(g, e)
	n := g.Node(e)
	switch n.Kind {
	case aig.Input:
		vec.Vec_IntPush(terms, int32(e.NodeID()))
	case aig.Output, aig.Buffer:
		Saig_ManCexMinCollectFrameTerms_rec(g, n.Fanins[0], terms)
	case aig.And, aig.ExclusiveOr:
		Saig_ManCexMinCollectFrameTerms_rec(g, n.Fanins[0], terms)
		Saig_ManCexMinCollectFrameTerms_rec(g, n.Fanins[1], terms)
	}
}
func Saig_ManCexMinCollectFrameTerms(g *aig.Graph, cex *util.Abc_Cex_t) *vec.Vec_Vec_t {
	if g == nil || cex == nil || cex.IFrame < 0 {
		return nil
	}
	frames := vec.Vec_VecAlloc(int(cex.IFrame + 1))
	vec.Vec_VecExpandInt(frames, int(cex.IFrame))
	roots := vec.Vec_IntAlloc(0)
	for f := cex.IFrame; f >= 0; f-- {
		var leaves *vec.Vec_Int_t
		if f < cex.IFrame {
			leaves = vec.Vec_VecEntryInt(frames, int(f+1))
		}
		Saig_ManCexMinGetCos(g, cex, leaves, roots)
		aig.Aig_ManIncrementTravId(g)
		for _, id := range vec.Vec_IntArray(roots) {
			Saig_ManCexMinCollectFrameTerms_rec(g, aig.Edge(id)<<1, vec.Vec_VecEntryInt(frames, int(f)))
		}
	}
	return frames
}
