package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// cexFramePriorities propagates the value and priority of each frame's
// relevant inputs through the combinational logic.
func cexFramePriorities(g *aig.Graph, cex *util.Abc_Cex_t, terms *vec.Vec_Vec_t) ([][]int32, *vec.Vec_Vec_t) {
	frames := make([][]int32, cex.IFrame+1)
	result := vec.Vec_VecAlloc(len(frames))
	vec.Vec_VecExpandInt(result, len(frames)-1)
	roots := vec.Vec_IntAlloc(0)
	for f := int32(0); f <= cex.IFrame; f++ {
		pp := make([]int32, len(g.Nodes()))
		for i := range pp {
			pp[i] = -1
		}
		pp[0] = 2*(cex.NRegs+(cex.IFrame+1)*cex.NPis) + 1
		piCount := int32(0)
		for _, id := range vec.Vec_IntArray(vec.Vec_VecEntryInt(terms, int(f))) {
			e := aig.Edge(id) << 1
			var value int32
			if saig.Saig_ObjIsPi(g, e) {
				priority := cex.NRegs + (f+1)*cex.NPis - 1 - piCount
				piCount++
				if util.Abc_InfoHasBit(cex.PData, cex.NRegs+f*cex.NPis+aig.Aig_ObjCioId(g, e)) {
					value = 1
				}
				value |= priority << 1
			} else if f == 0 {
				value = saig.Saig_ObjRegId(g, e) << 1
			} else {
				li := saig.Saig_ObjLoToLi(g, e)
				value = frames[f-1][li.NodeID()]
				if value < 0 {
					panic("bmc: missing previous frame priority")
				}
			}
			pp[id] = value
			vec.Vec_IntPush(vec.Vec_VecEntryInt(result, int(f)), value)
		}
		var eval func(aig.Edge) int32
		eval = func(e aig.Edge) int32 {
			id := e.NodeID()
			if pp[id] >= 0 {
				return pp[id] ^ int32(e&1)
			}
			n := g.Node(e)
			var x int32
			switch n.Kind {
			case aig.Output, aig.Buffer:
				x = eval(n.Fanins[0])
			case aig.And:
				a, b := eval(n.Fanins[0]), eval(n.Fanins[1])
				av, bv := a&1, b&1
				priority := a >> 1
				switch {
				case av == 1 && bv == 1:
					priority = min(a>>1, b>>1)
					x = priority<<1 | 1
				case av == 0 && bv == 1:
					x = priority << 1
				case av == 1 && bv == 0:
					x = b &^ 1
				default:
					x = max(a>>1, b>>1) << 1
				}
			case aig.ExclusiveOr:
				a, b := eval(n.Fanins[0]), eval(n.Fanins[1])
				x = min(a>>1, b>>1)<<1 | (a^b)&1
			default:
				panic("bmc: missing frame input priority")
			}
			pp[id] = x
			return x ^ int32(e&1)
		}
		var leaves *vec.Vec_Int_t
		if f < cex.IFrame {
			leaves = vec.Vec_VecEntryInt(terms, int(f+1))
		}
		Saig_ManCexMinGetCos(g, cex, leaves, roots)
		for _, id := range vec.Vec_IntArray(roots) {
			eval(aig.Edge(id) << 1)
		}
		frames[f] = pp
	}
	return frames, result
}
func Saig_ManCexMinCollectPhasePriority(g *aig.Graph, cex *util.Abc_Cex_t, terms *vec.Vec_Vec_t) *vec.Vec_Vec_t {
	_, priorities := cexFramePriorities(g, cex, terms)
	return priorities
}

// Saig_ManCexMinComputeReason selects one priority-guided justification in
// each frame, expressed as input or register literals.
func Saig_ManCexMinComputeReason(g *aig.Graph, cex *util.Abc_Cex_t, piReason int32) *vec.Vec_Vec_t {
	if g == nil || cex == nil || cex.IFrame < 0 || cex.NRegs != aig.Aig_ManRegNum(g) || cex.NPis != saig.Saig_ManPiNum(g) || cex.IPo < 0 || cex.IPo >= saig.Saig_ManPoNum(g) {
		return nil
	}
	terms := Saig_ManCexMinCollectFrameTerms(g, cex)
	frames, _ := cexFramePriorities(g, cex, terms)
	reasons := vec.Vec_VecAlloc(int(cex.IFrame + 1))
	vec.Vec_VecExpandInt(reasons, int(cex.IFrame))
	roots := vec.Vec_IntAlloc(0)
	for f := cex.IFrame; f >= 0; f-- {
		pp := frames[f]
		seen := make([]bool, len(pp))
		reason := vec.Vec_VecEntryInt(reasons, int(f))
		var visit func(aig.Edge)
		visit = func(e aig.Edge) {
			e = aig.Aig_Regular(e)
			id := e.NodeID()
			if seen[id] {
				return
			}
			seen[id] = true
			n := g.Node(e)
			switch n.Kind {
			case aig.Constant:
				return
			case aig.Input:
				var index int32
				if piReason != 0 {
					if !saig.Saig_ObjIsPi(g, e) {
						return
					}
					index = aig.Aig_ObjCioId(g, e)
				} else {
					if !saig.Saig_ObjIsLo(g, e) {
						return
					}
					index = saig.Saig_ObjRegId(g, e)
				}
				vec.Vec_IntPush(reason, index<<1|((pp[id]&1)^1))
			case aig.Output, aig.Buffer:
				visit(n.Fanins[0])
			case aig.ExclusiveOr:
				visit(n.Fanins[0])
				visit(n.Fanins[1])
			case aig.And:
				a, b := n.Fanins[0], n.Fanins[1]
				x, y := (pp[a.NodeID()]^int32(a&1))&1, (pp[b.NodeID()]^int32(b&1))&1
				if x == 1 && y == 1 {
					visit(a)
					visit(b)
				} else if x == 0 && y == 1 {
					visit(a)
				} else if x == 1 && y == 0 {
					visit(b)
				} else if pp[a.NodeID()]>>1 >= pp[b.NodeID()]>>1 {
					visit(a)
				} else {
					visit(b)
				}
			}
		}
		var leaves *vec.Vec_Int_t
		if f < cex.IFrame {
			leaves = vec.Vec_VecEntryInt(terms, int(f+1))
		}
		Saig_ManCexMinGetCos(g, cex, leaves, roots)
		for _, id := range vec.Vec_IntArray(roots) {
			visit(aig.Edge(id) << 1)
		}
	}
	return reasons
}
