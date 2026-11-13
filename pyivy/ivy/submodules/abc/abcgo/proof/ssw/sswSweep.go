package ssw

// One round of SAT sweeping, ported from sswSweep.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Ssw_ManGetSatVarValue(p *Ssw_Man_t, obj aig.Edge, frame int32) int32 {
	edge := Ssw_ObjFrame(p, obj, frame)
	if edge == aig.InvalidEdge {
		panic("ssw: missing frame mapping")
	}
	variable := Ssw_ObjSatNum(p.pMSat, aig.Aig_Regular(edge))
	var value int32
	if edge.Complemented() {
		value = 1
	}
	if variable != 0 {
		bit, ok := p.pMSat.pSat.Value(variable)
		if !ok {
			panic("ssw: SAT model unavailable")
		}
		if bit {
			value ^= 1
		}
	}
	// The solver owns its encoding polarity; it may differ from the parameters.
	if p.pMSat.fPolarFlip != 0 && aig.Aig_ObjPhase(p.pFrames, aig.Aig_Regular(edge)) {
		value ^= 1
	}
	return value
}
func Ssw_CheckConstraints(p *Ssw_Man_t) {
	n := aig.Aig_ManCoNum(p.pFrames) - aig.Aig_ManRegNum(p.pAig)
	if n&1 != 0 {
		panic("ssw: unpaired constraints")
	}
	var added int32
	for i := int32(0); i < n; i += 2 {
		a, b := aig.Aig_ManCo(p.pFrames, i), aig.Aig_ManCo(p.pFrames, i+1)
		if Ssw_NodesAreEquiv(p, aig.Aig_ObjFanin0(p.pFrames, a), aig.Aig_ObjFanin0(p.pFrames, b)) != 1 {
			Ssw_NodesAreConstrained(p, aig.Aig_ObjChild0(p.pFrames, a), aig.Aig_ObjChild0(p.pFrames, b))
			added++
		}
	}
	fmt.Printf("Total constraints = %d. Added constraints = %d.\n", n/2, added)
}
func Ssw_SmlSavePatternAigPhase(p *Ssw_Man_t, frame int32) {
	clear(p.pPatWords)
	for i, obj := range p.pAig.Inputs() {
		edge := Ssw_ObjFrame(p, obj, frame)
		if edge == aig.InvalidEdge {
			panic("ssw: missing frame mapping")
		}
		if aig.Aig_ObjPhaseReal(p.pFrames, edge) {
			p.pPatWords[i/32] |= uint32(1) << uint(i%32)
		}
	}
}
func Ssw_SmlSavePatternAig(p *Ssw_Man_t, frame int32) {
	clear(p.pPatWords)
	for i, obj := range p.pAig.Inputs() {
		if Ssw_ManGetSatVarValue(p, obj, frame) != 0 {
			p.pPatWords[i/32] |= uint32(1) << uint(i%32)
		}
	}
}
func Ssw_SmlAddPatternDyn(p *Ssw_Man_t) {
	for _, item := range vec.Vec_PtrArray(p.pMSat.vUsedPis) {
		obj := item.(aig.Edge)
		if !aig.Aig_ObjIsCi(p.pFrames, obj) {
			panic("ssw: SAT variable is not a frame input")
		}
		variable := Ssw_ObjSatNum(p.pMSat, obj)
		if variable <= 0 {
			panic("ssw: missing SAT variable")
		}
		value, ok := p.pMSat.pSat.Value(variable)
		if !ok {
			panic("ssw: SAT model unavailable")
		}
		if p.pMSat.fPolarFlip != 0 && aig.Aig_ObjPhase(p.pFrames, obj) {
			value = !value
		}
		if value {
			words := vec.Vec_PtrEntry(p.vSimInfo, int(aig.Aig_ObjCioId(p.pFrames, obj))).([]uint32)
			words[p.nPatterns/32] |= uint32(1) << uint(p.nPatterns%32)
		}
	}
}
func Ssw_ManSweepNode(p *Ssw_Man_t, obj aig.Edge, frame, bmc int32, pairs *vec.Vec_Int_t) int32 {
	repr := aig.Aig_ObjRepr(p.pAig, obj)
	if repr == aig.InvalidEdge {
		return 0
	}
	nodeFrame, reprFrame := Ssw_ObjFrame(p, obj, frame), Ssw_ObjFrame(p, repr, frame)
	if nodeFrame == aig.InvalidEdge || reprFrame == aig.InvalidEdge {
		panic("ssw: missing frame mapping")
	}
	if (aig.Aig_ObjPhase(p.pAig, obj) == aig.Aig_ObjPhase(p.pAig, repr)) != (aig.Aig_ObjPhaseReal(p.pFrames, nodeFrame) == aig.Aig_ObjPhaseReal(p.pFrames, reprFrame)) {
		panic("ssw: inconsistent candidate phase")
	}
	if aig.Aig_Regular(nodeFrame) == aig.Aig_Regular(reprFrame) {
		return 0
	}
	if bmc == 0 && p.pPars.fDynamic != 0 {
		start := time.Now()
		Ssw_ManLoadSolver(p, repr, obj)
		p.nRecycleCalls++
		p.timeMarkCones += time.Since(start)
	}
	a, b := aig.Aig_Regular(reprFrame), aig.Aig_Regular(nodeFrame)
	if b == p.pFrames.Const1() {
		a, b = b, a
	}
	result := Ssw_NodesAreEquiv(p, a, b)
	if result == 1 {
		adjusted := reprFrame
		if aig.Aig_ObjPhase(p.pAig, obj) != aig.Aig_ObjPhase(p.pAig, repr) {
			adjusted = adjusted.Not()
		}
		Ssw_ObjSetFrame(p, obj, frame, adjusted)
		if p.pPars.fEquivDump2 != 0 && pairs != nil {
			vec.Vec_IntPushTwo(pairs, int32(repr.NodeID()), int32(obj.NodeID()))
		}
		return 0
	}
	if p.pPars.fEquivDump != 0 && pairs != nil {
		vec.Vec_IntPushTwo(pairs, int32(repr.NodeID()), int32(obj.NodeID()))
	}
	if result == -1 {
		Ssw_ClassesRemoveNode(p.ppClasses, obj)
		return 1
	}
	if bmc == 0 && p.pPars.fDynamic != 0 {
		Ssw_SmlAddPatternDyn(p)
		p.nPatterns++
		return 1
	}
	Ssw_SmlSavePatternAig(p, frame)
	if p.pPars.fConstrs == 0 {
		Ssw_ManResimulateWord(p, obj, repr, frame)
	} else {
		Ssw_ManResimulateBit(p, obj, repr)
	}
	if sswDeadlineExpired(p.pPars) {
		return 0
	}
	if aig.Aig_ObjRepr(p.pAig, obj) == repr {
		panic("ssw: counterexample failed to refine representative")
	}
	return 1
}
func Ssw_ManSweepBmc(p *Ssw_Man_t) int32 {
	start := time.Now()
	p.pFrames = aig.Aig_ManStart(aig.Aig_ManObjNumMax(p.pAig) * p.pPars.nFramesK)
	p.pMSat.pAig = p.pFrames
	inputs, outputs := p.pAig.Inputs(), p.pAig.Outputs()
	regs := int(aig.Aig_ManRegNum(p.pAig))
	pis, pos := len(inputs)-regs, len(outputs)-regs
	for _, lo := range inputs[pis:] {
		Ssw_ObjSetFrame(p, lo, 0, p.pFrames.Const0())
	}
	p.fRefined = 0
	for frame := int32(0); frame < p.pPars.nFramesK; frame++ {
		if sswDeadlineExpired(p.pPars) {
			return p.fRefined
		}
		Ssw_ObjSetFrame(p, p.pAig.Const1(), frame, p.pFrames.Const1())
		for _, pi := range inputs[:pis] {
			Ssw_ObjSetFrame(p, pi, frame, p.pFrames.NewInput())
		}
		for _, lo := range inputs[pis:] {
			p.fRefined |= Ssw_ManSweepNode(p, lo, frame, 1, nil)
		}
		for _, obj := range sswFrameNodes(p.pAig) {
			if sswDeadlineExpired(p.pPars) {
				return p.fRefined
			}
			Ssw_ObjSetFrame(p, obj, frame, sswMakeNode(p, p.pFrames, obj, frame))
			p.fRefined |= Ssw_ManSweepNode(p, obj, frame, 1, nil)
		}
		if frame == p.pPars.nFramesK-1 {
			break
		}
		for _, out := range outputs {
			Ssw_ObjSetFrame(p, out, frame, Ssw_ObjChild0Fra(p, out, frame))
		}
		for i, lo := range inputs[pis:] {
			edge := Ssw_ObjFrame(p, outputs[pos+i], frame)
			Ssw_ObjSetFrame(p, lo, frame+1, edge)
			Ssw_CnfNodeAddToSolver(p.pMSat, aig.Aig_Regular(edge))
		}
	}
	p.timeBmc += time.Since(start)
	return p.fRefined
}

var sswEquivDumpCounter int32

func Ssw_ManDumpEquivMiter(g *aig.Graph, pairs *vec.Vec_Int_t, num, addOutputs int32) {
	name := fmt.Sprintf("equiv%03d.aig", num)
	miter := saig.Saig_ManCreateEquivMiter(g, pairs, addOutputs)
	defer aig.Aig_ManStop(miter)
	if err := ioa.Ioa_WriteAiger(miter, name, 0, 0); err != nil {
		fmt.Printf("Cannot write %s: %v\n", name, err)
		return
	}
	fmt.Printf("AIG with %4d disproved equivs is dumped into file %q.\n", vec.Vec_IntSize(pairs)/2, name)
}
func Ssw_ManSweep(p *Ssw_Man_t) int32 {
	start := time.Now()
	p.pFrames = Ssw_FramesWithClasses(p)
	if p.pFrames == nil {
		return p.fRefined
	}
	p.pMSat.pAig = p.pFrames
	n := aig.Aig_ManCoNum(p.pFrames) - aig.Aig_ManRegNum(p.pAig)
	if n&1 != 0 {
		panic("ssw: unpaired constraints")
	}
	for i := int32(0); i < n; i += 2 {
		a, b := aig.Aig_ManCo(p.pFrames, i), aig.Aig_ManCo(p.pFrames, i+1)
		Ssw_NodesAreConstrained(p, aig.Aig_ObjChild0(p.pFrames, a), aig.Aig_ObjChild0(p.pFrames, b))
	}
	for i := int32(0); i < aig.Aig_ManRegNum(p.pAig); i++ {
		out := aig.Aig_ManCo(p.pFrames, n+i)
		Ssw_CnfNodeAddToSolver(p.pMSat, aig.Aig_ObjFanin0(p.pFrames, out))
	}
	// The C driver does not abort when simplification finds a contradiction;
	// subsequent SAT queries observe the resulting UNSAT solver state.
	p.pMSat.pSat.Simplify()
	frame := p.pPars.nFramesK
	Ssw_ObjSetFrame(p, p.pAig.Const1(), frame, p.pFrames.Const1())
	inputs := p.pAig.Inputs()
	pis := len(inputs) - int(aig.Aig_ManRegNum(p.pAig))
	for _, pi := range inputs[:pis] {
		Ssw_ObjSetFrame(p, pi, frame, p.pFrames.NewInput())
	}
	p.timeReduce += time.Since(start)
	p.fRefined = 0
	Ssw_ClassesClearRefined(p.ppClasses)
	var pairs *vec.Vec_Int_t
	if p.pPars.fEquivDump != 0 || p.pPars.fEquivDump2 != 0 {
		pairs = vec.Vec_IntAlloc(1000)
	}
	latches := make(map[int]bool)
	for _, lo := range inputs[pis:] {
		latches[lo.NodeID()] = true
	}
	for _, obj := range p.pAig.LiveEdges() {
		if sswDeadlineExpired(p.pPars) {
			return p.fRefined
		}
		if latches[obj.NodeID()] {
			p.fRefined |= Ssw_ManSweepNode(p, obj, frame, 0, pairs)
		} else if aig.Aig_ObjIsNode(p.pAig, obj) {
			Ssw_ObjSetFrame(p, obj, frame, sswMakeNode(p, p.pFrames, obj, frame))
			p.fRefined |= Ssw_ManSweepNode(p, obj, frame, 0, pairs)
		}
	}
	if p.pPars.fEquivDump != 0 {
		Ssw_ManDumpEquivMiter(p.pAig, pairs, sswEquivDumpCounter, 1)
		sswEquivDumpCounter++
	}
	if p.pPars.fEquivDump2 != 0 && p.fRefined == 0 {
		Ssw_ManDumpEquivMiter(p.pAig, pairs, 0, 0)
	}
	return p.fRefined
}
