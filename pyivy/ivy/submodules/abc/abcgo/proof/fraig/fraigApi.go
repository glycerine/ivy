package fraig

// FRAIG manager and node accessors, adapted from fraigApi.c by Alan Mishchenko.
import (
	"fmt"

	"github.com/glycerine/abcgo/sat/msat"
)

func Fraig_ManReadVecInputs(p *Fraig_Man_t) *Fraig_NodeVec_t      { return p.vInputs }
func Fraig_ManReadVecOutputs(p *Fraig_Man_t) *Fraig_NodeVec_t     { return p.vOutputs }
func Fraig_ManReadVecNodes(p *Fraig_Man_t) *Fraig_NodeVec_t       { return p.vNodes }
func Fraig_ManReadInputs(p *Fraig_Man_t) []*Fraig_Node_t          { return p.vInputs.pArray }
func Fraig_ManReadOutputs(p *Fraig_Man_t) []*Fraig_Node_t         { return p.vOutputs.pArray }
func Fraig_ManReadNodes(p *Fraig_Man_t) []*Fraig_Node_t           { return p.vNodes.pArray }
func Fraig_ManReadInputNum(p *Fraig_Man_t) int32                  { return int32(len(p.vInputs.pArray)) }
func Fraig_ManReadOutputNum(p *Fraig_Man_t) int32                 { return int32(len(p.vOutputs.pArray)) }
func Fraig_ManReadNodeNum(p *Fraig_Man_t) int32                   { return int32(len(p.vNodes.pArray)) }
func Fraig_ManReadConst1(p *Fraig_Man_t) *Fraig_Node_t            { return p.pConst1 }
func Fraig_ManReadIthNode(p *Fraig_Man_t, i int32) *Fraig_Node_t  { return p.vNodes.pArray[i] }
func Fraig_ManReadInputNames(p *Fraig_Man_t) []string             { return p.ppInputNames }
func Fraig_ManReadOutputNames(p *Fraig_Man_t) []string            { return p.ppOutputNames }
func Fraig_ManReadVarsInt(p *Fraig_Man_t) *msat.Msat_IntVec_t     { return p.vVarsInt }
func Fraig_ManReadSat(p *Fraig_Man_t) *msat.Msat_Solver_t         { return p.pSat }
func Fraig_ManReadFuncRed(p *Fraig_Man_t) int32                   { return p.fFuncRed }
func Fraig_ManReadFeedBack(p *Fraig_Man_t) int32                  { return p.fFeedBack }
func Fraig_ManReadDoSparse(p *Fraig_Man_t) int32                  { return p.fDoSparse }
func Fraig_ManReadChoicing(p *Fraig_Man_t) int32                  { return p.fChoicing }
func Fraig_ManReadVerbose(p *Fraig_Man_t) int32                   { return p.fVerbose }
func Fraig_ManReadModel(p *Fraig_Man_t) []int32                   { return p.pModel }
func Fraig_ManReadPatternNumRandom(p *Fraig_Man_t) int32          { return p.nWordsRand * 32 }
func Fraig_ManReadPatternNumDynamic(p *Fraig_Man_t) int32         { return p.iWordStart * 32 }
func Fraig_ManReadPatternNumDynamicFiltered(p *Fraig_Man_t) int32 { return p.iPatsPerm }
func Fraig_ManReadSatFails(p *Fraig_Man_t) int32                  { return p.nSatFailsReal }
func Fraig_ManReadConflicts(p *Fraig_Man_t) int32 {
	if p.pSat == nil {
		return 0
	}
	return int32(msat.Msat_SolverReadBackTracks(p.pSat))
}
func Fraig_ManReadInspects(p *Fraig_Man_t) int32 {
	if p.pSat == nil {
		return 0
	}
	return int32(msat.Msat_SolverReadInspects(p.pSat))
}
func Fraig_ManSetFuncRed(p *Fraig_Man_t, value int32)        { p.fFuncRed = value }
func Fraig_ManSetFeedBack(p *Fraig_Man_t, value int32)       { p.fFeedBack = value }
func Fraig_ManSetDoSparse(p *Fraig_Man_t, value int32)       { p.fDoSparse = value }
func Fraig_ManSetChoicing(p *Fraig_Man_t, value int32)       { p.fChoicing = value }
func Fraig_ManSetTryProve(p *Fraig_Man_t, value int32)       { p.fTryProve = value }
func Fraig_ManSetVerbose(p *Fraig_Man_t, value int32)        { p.fVerbose = value }
func Fraig_ManSetOutputNames(p *Fraig_Man_t, names []string) { p.ppOutputNames = names }
func Fraig_ManSetInputNames(p *Fraig_Man_t, names []string)  { p.ppInputNames = names }

func Fraig_NodeReadData0(p *Fraig_Node_t) *Fraig_Node_t { return p.pData0 }
func Fraig_NodeReadData1(p *Fraig_Node_t) *Fraig_Node_t { return p.pData1 }
func Fraig_NodeReadNum(p *Fraig_Node_t) int32           { return p.Num }
func Fraig_NodeReadOne(p *Fraig_Node_t) *Fraig_Node_t {
	if Fraig_IsComplement(p) {
		panic("fraig: complemented node in fanin accessor")
	}
	return p.p1
}
func Fraig_NodeReadTwo(p *Fraig_Node_t) *Fraig_Node_t {
	if Fraig_IsComplement(p) {
		panic("fraig: complemented node in fanin accessor")
	}
	return p.p2
}
func Fraig_NodeReadNextE(p *Fraig_Node_t) *Fraig_Node_t      { return p.pNextE }
func Fraig_NodeReadRepr(p *Fraig_Node_t) *Fraig_Node_t       { return p.pRepr }
func Fraig_NodeReadNumRefs(p *Fraig_Node_t) int32            { return p.nRefs }
func Fraig_NodeReadNumFanouts(p *Fraig_Node_t) int32         { return int32(p.nFanouts) }
func Fraig_NodeReadSimInv(p *Fraig_Node_t) int32             { return int32(p.fInv) }
func Fraig_NodeReadNumOnes(p *Fraig_Node_t) int32            { return int32(p.nOnes) }
func Fraig_NodeReadPatternsRandom(p *Fraig_Node_t) []uint32  { return p.puSimR }
func Fraig_NodeReadPatternsDynamic(p *Fraig_Node_t) []uint32 { return p.puSimD }
func Fraig_NodeSetData0(p *Fraig_Node_t, data *Fraig_Node_t) { p.pData0 = data }
func Fraig_NodeSetData1(p *Fraig_Node_t, data *Fraig_Node_t) { p.pData1 = data }

func fraigBool(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
func Fraig_NodeIsConst(p *Fraig_Node_t) int32 { return fraigBool(Fraig_Regular(p).Num == 0) }
func Fraig_NodeIsVar(p *Fraig_Node_t) int32   { return fraigBool(Fraig_Regular(p).NumPi >= 0) }
func Fraig_NodeIsAnd(p *Fraig_Node_t) int32 {
	node := Fraig_Regular(p)
	return fraigBool(node.NumPi < 0 && node.Num > 0)
}
func Fraig_NodeComparePhase(a, b *Fraig_Node_t) int32 {
	if Fraig_IsComplement(a) || Fraig_IsComplement(b) {
		panic("fraig: complemented node in phase comparison")
	}
	return int32(a.fInv ^ b.fInv)
}
func Fraig_ManReadIthVar(p *Fraig_Man_t, i int32) *Fraig_Node_t {
	if i < 0 {
		fmt.Println("Requesting a PI with a negative number")
		return nil
	}
	for int(i) >= len(p.vInputs.pArray) {
		Fraig_NodeCreatePi(p)
	}
	return p.vInputs.pArray[i]
}
func Fraig_ManSetPo(p *Fraig_Man_t, node *Fraig_Node_t) {
	Fraig_Regular(node).fNodePo = 1
	Fraig_NodeVecPush(p.vOutputs, node)
}
func Fraig_NodeAnd(p *Fraig_Man_t, a, b *Fraig_Node_t) *Fraig_Node_t {
	return Fraig_NodeAndCanon(p, a, b)
}
func Fraig_NodeOr(p *Fraig_Man_t, a, b *Fraig_Node_t) *Fraig_Node_t {
	return Fraig_Not(Fraig_NodeAndCanon(p, Fraig_Not(a), Fraig_Not(b)))
}
func Fraig_NodeExor(p *Fraig_Man_t, a, b *Fraig_Node_t) *Fraig_Node_t {
	return Fraig_NodeMux(p, a, Fraig_Not(b), b)
}
func Fraig_NodeMux(p *Fraig_Man_t, control, yes, no *Fraig_Node_t) *Fraig_Node_t {
	a := Fraig_NodeAndCanon(p, control, yes)
	b := Fraig_NodeAndCanon(p, Fraig_Not(control), no)
	return Fraig_NodeOr(p, a, b)
}
func Fraig_NodeSetChoice(_ *Fraig_Man_t, old, next *Fraig_Node_t) {
	next.pNextE = old.pNextE
	old.pNextE = next
	next.pRepr = old
}
