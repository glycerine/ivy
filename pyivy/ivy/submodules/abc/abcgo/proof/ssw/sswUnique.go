package ssw

// Uniqueness constraints, ported from sswUnique.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ssw_UniqueRegisterPairInfo(p *Ssw_Man_t) {
	regs := int(p.pAig.RegisterCount())
	if p.vDiffPairs == nil {
		p.vDiffPairs = vec.Vec_IntAlloc(regs)
	}
	vec.Vec_IntClear(p.vDiffPairs)
	inputs := p.pAig.Inputs()
	for _, lo := range inputs[len(inputs)-regs:] {
		a, b := Ssw_ObjFrame(p, lo, 0), Ssw_ObjFrame(p, lo, 1)
		var different int32
		if a == b {
			different = 0
		} else if a == b.Not() || aig.Aig_ObjPhaseReal(p.pFrames, a) != aig.Aig_ObjPhaseReal(p.pFrames, b) {
			different = 1
		} else if Ssw_NodesAreEquiv(p, a&^1, b&^1) != 1 {
			different = 1
		}
		vec.Vec_IntPush(p.vDiffPairs, different)
	}
}
func Ssw_ManUniqueOne(p *Ssw_Man_t, repr, obj aig.Edge, verbose int32) int32 {
	if p.pPars.nFramesK <= 1 || p.vDiffPairs == nil || vec.Vec_IntSize(p.vDiffPairs) != int(p.pAig.RegisterCount()) {
		panic("ssw: invalid uniqueness state")
	}
	var support []aig.Edge
	aig.Aig_SupportNodes(p.pAig, []aig.Edge{repr, obj}, 2, &support)
	vec.Vec_PtrClear(p.vCommon)
	feasible := false
	for _, node := range support {
		if !saig.Saig_ObjIsLo(p.pAig, node) {
			continue
		}
		vec.Vec_PtrPush(p.vCommon, node)
		index := aig.Aig_ObjCioId(p.pAig, node) - saig.Saig_ManPiNum(p.pAig)
		if vec.Vec_IntEntry(p.vDiffPairs, int(index)) != 0 {
			feasible = true
		}
	}
	if verbose != 0 {
		fmt.Printf("Node = %5d : Supp = %3d. Regs = %3d. Feasible = %t. ", obj.NodeID(), len(support), vec.Vec_PtrSize(p.vCommon), feasible)
	}
	equal := true
	for _, item := range vec.Vec_PtrArray(p.vCommon) {
		node := item.(aig.Edge)
		a, b := Ssw_ManGetSatVarValue(p, node, 0), Ssw_ManGetSatVarValue(p, node, 1)
		if a != b {
			equal = false
		}
		if verbose != 0 {
			fmt.Printf("%d", a^b)
		}
	}
	if verbose != 0 {
		fmt.Println()
	}
	if equal && feasible {
		return 1
	}
	return 0
}
func Ssw_ManUniqueAddConstraint(p *Ssw_Man_t, common *vec.Vec_Ptr_t, f1, f2 int32) int32 {
	if vec.Vec_PtrSize(common) == 0 {
		panic("ssw: empty uniqueness support")
	}
	total := p.pFrames.Const0()
	for _, item := range vec.Vec_PtrArray(common) {
		obj := item.(aig.Edge)
		if !saig.Saig_ObjIsLo(p.pAig, obj) {
			panic("ssw: uniqueness support contains non-latch")
		}
		total = p.pFrames.OR(total, p.pFrames.XOR(Ssw_ObjFrame(p, obj, f1), Ssw_ObjFrame(p, obj, f2)))
	}
	if total.NodeID() == 0 {
		return 0
	}
	Ssw_CnfNodeAddToSolver(p.pMSat, total&^1)
	if p.iOutputLit != -1 {
		panic("ssw: uniqueness constraint already active")
	}
	p.iOutputLit = Ssw_ObjSatNum(p.pMSat, total) << 1
	if total.Complemented() {
		p.iOutputLit |= 1
	}
	return 1
}
