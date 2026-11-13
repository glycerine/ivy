package ssw

// Dynamic bounded unrolling and counterexample extraction, ported from sswBmc.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

func Ssw_BmcUnroll_rec(p *Ssw_Frm_t, obj aig.Edge, frame int32) aig.Edge {
	if frame < 0 || obj.Complemented() {
		panic("ssw: invalid unrolling object or frame")
	}
	if result := Ssw_ObjFrame_(p, obj, frame); result != aig.InvalidEdge {
		return result
	}
	var result aig.Edge
	switch {
	case aig.Aig_ObjIsConst1(p.pAig, obj):
		result = p.pFrames.Const1()
	case saig.Saig_ObjIsPi(p.pAig, obj):
		result = p.pFrames.NewInput()
	case aig.Aig_ObjIsCo(p.pAig, obj) || aig.Aig_ObjIsBuf(p.pAig, obj):
		Ssw_BmcUnroll_rec(p, aig.Aig_ObjFanin0(p.pAig, obj), frame)
		result = Ssw_ObjChild0Fra_(p, obj, frame)
	case saig.Saig_ObjIsLo(p.pAig, obj):
		if frame == 0 {
			result = p.pFrames.Const0()
		} else {
			result = Ssw_BmcUnroll_rec(p, saig.Saig_ObjLoToLi(p.pAig, obj), frame-1)
		}
	case aig.Aig_ObjIsNode(p.pAig, obj):
		Ssw_BmcUnroll_rec(p, aig.Aig_ObjFanin0(p.pAig, obj), frame)
		Ssw_BmcUnroll_rec(p, aig.Aig_ObjFanin1(p.pAig, obj), frame)
		a, b := Ssw_ObjChild0Fra_(p, obj, frame), Ssw_ObjChild1Fra_(p, obj, frame)
		if p.pAig.Node(obj).Kind == aig.ExclusiveOr {
			result = p.pFrames.NewXor(a, b)
		} else {
			result = p.pFrames.AND(a, b)
		}
	default:
		panic("ssw: unsupported object during unrolling")
	}
	Ssw_ObjSetFrame_(p, obj, frame, result)
	return result
}
func Ssw_BmcGetCounterExample(p *Ssw_Frm_t, sat *Ssw_Sat_t, po, frame int32) *util.Abc_Cex_t {
	regs, pis := saig.Saig_ManRegNum(p.pAig), saig.Saig_ManPiNum(p.pAig)
	if regs <= 0 {
		panic("ssw: BMC requires registers")
	}
	cex := util.Abc_CexAlloc(regs, pis, frame+1)
	cex.IPo = po
	cex.IFrame = frame
	for f := int32(0); f <= frame; f++ {
		for i, obj := range p.pAig.Inputs()[:pis] {
			mapped := Ssw_ObjFrame_(p, obj, f)
			if mapped != aig.InvalidEdge && Ssw_CnfGetNodeValue(sat, mapped) != 0 {
				util.Abc_InfoSetBit(cex.PData, regs+f*pis+int32(i))
			}
		}
	}
	return cex
}
func Ssw_BmcDynamic(graph *aig.Graph, maxFrames, confLimit, verbose int32, lastFrame *int32) int32 {
	if saig.Saig_ManRegNum(graph) <= 0 {
		panic("ssw: BMC requires registers")
	}
	aig.Aig_ManSetCioIds(graph)
	sat := Ssw_SatStart(0)
	defer Ssw_SatStop(sat)
	frm := Ssw_FrmStart(graph)
	defer Ssw_FrmStop(frm)
	frm.pFrames = aig.Aig_ManStart(aig.Aig_ManObjNumMax(graph) * 3)
	sat.pAig = frm.pFrames
	if verbose != 0 {
		fmt.Printf("AIG:  PI/PO/Reg = %d/%d/%d.  Node = %6d. Lev = %5d.\n", saig.Saig_ManPiNum(graph), saig.Saig_ManPoNum(graph), saig.Saig_ManRegNum(graph), aig.Aig_ManNodeNum(graph), aig.Aig_ManLevelNum(graph))
	}
	result := int32(-1)
	for frame := int32(0); frame < maxFrames; frame++ {
		started := time.Now()
		for i, obj := range graph.Outputs()[:saig.Saig_ManPoNum(graph)] {
			mapped := Ssw_BmcUnroll_rec(frm, obj, frame)
			Ssw_CnfNodeAddToSolver(sat, aig.Aig_Regular(mapped))
			if !sat.pSat.Simplify() {
				panic("ssw: SAT formula inconsistent before BMC query")
			}
			lit := Ssw_ObjSatNum(sat, mapped) << 1
			if mapped.Complemented() {
				lit |= 1
			}
			status := sat.pSat.SolveWithSourceLimits([]bsat.Literal{lit}, int64(confLimit), 0, 0, 0)
			if status == bsat.Unsatisfiable {
				result = 1
				continue
			}
			if lastFrame != nil {
				*lastFrame = frame
			}
			if status == bsat.Satisfiable {
				cex := Ssw_BmcGetCounterExample(frm, sat, int32(i), frame)
				if err := graph.SetSequentialCounterexample(cex); err != nil {
					panic(err)
				}
				result = 0
			} else {
				result = -1
			}
			break
		}
		if verbose != 0 {
			fmt.Printf("Solved %2d outputs of frame %3d. Var = %8d. AIG = %9d. T = %s\n", saig.Saig_ManPoNum(graph), frame, sat.nSatVars, aig.Aig_ManNodeNum(frm.pFrames), time.Since(started))
		}
		if result != 1 {
			break
		}
	}
	return result
}
