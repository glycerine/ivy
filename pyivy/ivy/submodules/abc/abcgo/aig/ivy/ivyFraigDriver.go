// FRAIG sweep, miter and resource-limited prove drivers from ivyFraig.c.
package ivy

import (
	"fmt"
	"math"
	"time"
)

func Ivy_FraigStart(source *Ivy_Man_t, params *Ivy_FraigParams_t) *Ivy_FraigMan_t {
	p := ivyFraigStartSimulation(source, params)
	Ivy_ObjSetFraig(source.pConst1, p.pManFraig.pConst1)
	for i, n := range source.vPis {
		Ivy_ObjSetFraig(n, p.pManFraig.vPis[i])
	}
	return p
}
func Ivy_FraigStartSimple(source *Ivy_Man_t, params *Ivy_FraigParams_t) *Ivy_FraigMan_t {
	if params == nil || Ivy_ManCheck(source) == 0 || Ivy_ManLatchNum(source) != 0 || Ivy_ManBufNum(source) != 0 {
		panic("ivy: invalid FRAIG graph")
	}
	return &Ivy_FraigMan_t{pParams: params, pManAig: source, pManFraig: Ivy_ManStartFrom(source)}
}
func Ivy_FraigStop(p *Ivy_FraigMan_t) {
	if p == nil {
		return
	}
	if p.pParams.FVerbose != 0 {
		Ivy_FraigPrint(p)
	}
	if p.nSimWords > 0 {
		ivyFraigStopSimulation(p)
	}
	for _, n := range p.pManFraig.vObjs {
		if n != nil {
			n.fraig = nil
		}
	}
	p.pSat = nil
	p.vPiVars = nil
}
func Ivy_FraigPrint(p *Ivy_FraigMan_t) {
	fmt.Printf("FRAIG: nodes %d -> %d, SAT calls %d, proved %d, counterexamples %d, undecided %d; simulation %s, SAT %s, total %s\n", Ivy_ManNodeNum(p.pManAig), Ivy_ManNodeNum(p.pManFraig), p.nSatCalls, p.nSatProof, p.nSatCallsSat, p.nSatFailsReal, p.timeSim, p.timeSat, p.timeTotal)
}
func Ivy_FraigAnd(p *Ivy_FraigMan_t, old *Ivy_Obj_t) *Ivy_Obj_t {
	candidate := Ivy_Oper(p.pManFraig, Ivy_ObjChild0Equiv(old), Ivy_ObjChild1Equiv(old), old.Type)
	representative := Ivy_ObjClassNodeRepr(old)
	if representative == nil || (p.pParams.FDoSparse == 0 && representative == p.pManAig.pConst1) {
		return candidate
	}
	mapped := Ivy_ObjFraig(representative)
	if mapped == nil {
		panic("ivy: unmapped simulation representative")
	}
	if Ivy_Regular(candidate) == Ivy_Regular(mapped) {
		return candidate
	}
	result := Ivy_FraigNodesAreEquiv(p, Ivy_Regular(mapped), Ivy_Regular(candidate))
	if result == 1 {
		if Ivy_ObjClassNodeNext(old) == nil {
			representative.fMarkA = 1
		}
		return Ivy_NotCond(mapped, old.fPhase^representative.fPhase)
	}
	if result == 0 {
		Ivy_FraigResimulate(p)
	}
	return candidate
}
func Ivy_FraigMiterStatus(p *Ivy_Man_t) int32 {
	if p.pData != nil {
		return 0
	}
	undecided := false
	for _, output := range p.vPos {
		edge := Ivy_ObjChild0(output)
		if edge == Ivy_ManConst0(p) {
			continue
		}
		if edge == p.pConst1 || Ivy_ObjFaninPhase(edge) != 0 {
			return 0
		}
		undecided = true
	}
	if undecided {
		return -1
	}
	return 1
}
func Ivy_FraigMiterProve(p *Ivy_FraigMan_t) {
	for _, output := range p.pManAig.vPos {
		edge := Ivy_ObjChild0Equiv(output)
		if edge == Ivy_ManConst0(p.pManFraig) {
			continue
		}
		if edge == p.pManFraig.pConst1 || Ivy_ObjFaninPhase(edge) != 0 {
			p.pManFraig.pData = make([]int32, len(p.pManFraig.vPis))
			return
		}
		result := Ivy_FraigNodeIsConst(p, Ivy_Regular(edge))
		if result == 1 {
			mapped := Ivy_NotCond(p.pManFraig.pConst1, 1^Ivy_ObjFaninC0(output))
			Ivy_ObjFanin0(output).pEquiv = mapped
			if p.nSimWords > 0 {
				Ivy_ObjSetFraig(Ivy_ObjFanin0(output), mapped)
			}
		} else if result == 0 {
			p.pManFraig.pData = Ivy_FraigCreateModel(p)
			return
		}
	}
}
func Ivy_FraigSweep(p *Ivy_FraigMan_t) {
	p.nClassesBeg = p.lClasses.nItems
	if p.lClasses.pHead == p.pManAig.pConst1 {
		p.nClassesZero = Ivy_FraigCountClassNodes(p.lClasses.pHead)
	}
	for _, n := range ivyFastMapOrder(p.pManAig) {
		var mapped *Ivy_Obj_t
		if p.pManFraig.pData != nil {
			mapped = Ivy_Oper(p.pManFraig, Ivy_ObjChild0Equiv(n), Ivy_ObjChild1Equiv(n), n.Type)
		} else {
			mapped = Ivy_FraigAnd(p, n)
		}
		n.pEquiv = mapped
		Ivy_ObjSetFraig(n, mapped)
	}
	p.nClassesEnd = p.lClasses.nItems
	p.nNodesMiter = Ivy_ManNodeNum(p.pManFraig)
	if p.pParams.FProve != 0 && p.pManFraig.pData == nil {
		Ivy_FraigMiterProve(p)
	}
	for _, n := range p.pManAig.vPos {
		Ivy_ObjCreatePo(p.pManFraig, Ivy_ObjChild0Equiv(n))
	}
	Ivy_ManCleanup(p.pManFraig)
	for class := p.lClasses.pHead; class != nil; class = Ivy_ObjEquivListNext(class) {
		class.fMarkA = 0
	}
}
func Ivy_FraigPerform_int(source *Ivy_Man_t, params *Ivy_FraigParams_t, conflicts, inspections int64, usedConflicts, usedInspections *int64) *Ivy_Man_t {
	if usedConflicts != nil {
		*usedConflicts = 0
	}
	if usedInspections != nil {
		*usedInspections = 0
	}
	if Ivy_ManNodeNum(source) == 0 {
		return Ivy_ManDup(source)
	}
	start := time.Now()
	p := Ivy_FraigStart(source, params)
	p.nBTLimitGlobal, p.nInsLimitGlobal = conflicts, inspections
	Ivy_FraigSimulate(p)
	Ivy_FraigSweep(p)
	p.timeTotal = time.Since(start)
	if p.pSat != nil {
		statistics := p.pSat.Statistics()
		if usedConflicts != nil {
			*usedConflicts = statistics.Conflicts
		}
		if usedInspections != nil {
			*usedInspections = statistics.Inspections
		}
	}
	result := p.pManFraig
	Ivy_FraigStop(p)
	if Ivy_ManCheck(result) == 0 {
		panic("ivy: FRAIG result failed validation")
	}
	return result
}
func Ivy_FraigPerform(source *Ivy_Man_t, params *Ivy_FraigParams_t) *Ivy_Man_t {
	return Ivy_FraigPerform_int(source, params, 0, 0, nil, nil)
}
func Ivy_FraigMiter(source *Ivy_Man_t, params *Ivy_FraigParams_t) *Ivy_Man_t {
	start := time.Now()
	p := Ivy_FraigStartSimple(source, params)
	for _, n := range ivyFastMapOrder(source) {
		n.pEquiv = Ivy_Oper(p.pManFraig, Ivy_ObjChild0Equiv(n), Ivy_ObjChild1Equiv(n), n.Type)
	}
	Ivy_FraigMiterProve(p)
	for _, n := range source.vPos {
		Ivy_ObjCreatePo(p.pManFraig, Ivy_ObjChild0Equiv(n))
	}
	Ivy_ManCleanup(p.pManFraig)
	p.timeTotal = time.Since(start)
	result := p.pManFraig
	Ivy_FraigStop(p)
	return result
}
func Ivy_FraigMiterPrint(p *Ivy_Man_t, label string, start time.Time, verbose int32) {
	if verbose != 0 {
		fmt.Printf("Nodes = %7d. Levels = %4d. %s %s\n", Ivy_ManNodeNum(p), Ivy_ManLevels(p), label, time.Since(start))
	}
}

type Prove_Params_t struct {
	fUseFraiging, fUseRewriting, fUseBdds, fVerbose                                                                           int32
	nItersMax, nMiteringLimitStart, nRewritingLimitStart, nFraigingLimitStart, nBddSizeLimit, fBddReorder, nMiteringLimitLast int32
	nMiteringLimitMulti, nRewritingLimitMulti, nFraigingLimitMulti                                                            float32
	nTotalBacktrackLimit, nTotalInspectLimit, nTotalBacktracksMade, nTotalInspectsMade                                        int64
}

// Ivy_ProveParams exposes only the command-level proof settings shared with
// base/abci. The source C code passes Prove_Params_t directly across packages.
type Ivy_ProveParams struct {
	FUseFraiging, FUseRewriting, FUseBdds, FVerbose                           int32
	NItersMax, NMiteringLimitStart, NRewritingLimitStart, NFraigingLimitStart int32
	NBddSizeLimit, FBddReorder, NMiteringLimitLast                            int32
	NMiteringLimitMulti, NRewritingLimitMulti, NFraigingLimitMulti            float32
	NTotalBacktrackLimit, NTotalInspectLimit                                  int64
	NTotalBacktracksMade, NTotalInspectsMade                                  int64
}

// Ivy_FraigProveWithParams matches the cross-package C Prove_Params_t call and
// reflects mutable counters/settings back into the caller's parameter object.
func Ivy_FraigProveWithParams(graph **Ivy_Man_t, params *Ivy_ProveParams) int32 {
	if params == nil {
		panic("ivy: prove parameters required")
	}
	p := Prove_Params_t{
		fUseFraiging: params.FUseFraiging, fUseRewriting: params.FUseRewriting,
		fUseBdds: params.FUseBdds, fVerbose: params.FVerbose,
		nItersMax: params.NItersMax, nMiteringLimitStart: params.NMiteringLimitStart,
		nMiteringLimitMulti: params.NMiteringLimitMulti, nRewritingLimitStart: params.NRewritingLimitStart,
		nRewritingLimitMulti: params.NRewritingLimitMulti, nFraigingLimitStart: params.NFraigingLimitStart,
		nFraigingLimitMulti: params.NFraigingLimitMulti, nBddSizeLimit: params.NBddSizeLimit,
		fBddReorder: params.FBddReorder, nMiteringLimitLast: params.NMiteringLimitLast,
		nTotalBacktrackLimit: params.NTotalBacktrackLimit, nTotalInspectLimit: params.NTotalInspectLimit,
		nTotalBacktracksMade: params.NTotalBacktracksMade, nTotalInspectsMade: params.NTotalInspectsMade,
	}
	result := Ivy_FraigProve(graph, &p)
	params.FUseRewriting = p.fUseRewriting
	params.NTotalBacktracksMade = p.nTotalBacktracksMade
	params.NTotalInspectsMade = p.nTotalInspectsMade
	return result
}

func Ivy_FraigProve(graph **Ivy_Man_t, params *Prove_Params_t) int32 {
	if graph == nil || *graph == nil || params == nil {
		panic("ivy: graph and prove parameters required")
	}
	var options Ivy_FraigParams_t
	Ivy_FraigParamsDefault(&options)
	options.FVerbose = params.fVerbose
	options.FProve = 1
	current := *graph
	replace := func(next *Ivy_Man_t) { Ivy_ManStop(current); current = next; *graph = current }
	divide := func(limit int32) int32 { return limit / max(int32(1), Ivy_ManPoNum(current)) }
	if params.fUseRewriting == 0 && params.fUseFraiging == 0 {
		options.NBTLimitMiter = divide(params.nMiteringLimitLast)
		replace(Ivy_FraigMiter(current, &options))
		return Ivy_FraigMiterStatus(current)
	}
	if Ivy_ManNodeNum(current) < 500 {
		options.NBTLimitMiter = divide(params.nMiteringLimitStart)
		replace(Ivy_FraigMiter(current, &options))
		if status := Ivy_FraigMiterStatus(current); status >= 0 {
			return status
		}
	}
	status := int32(-1)
	for iteration := int32(0); iteration < params.nItersMax; iteration++ {
		// Rewriting is disabled in the original active C driver as well.
		status = Ivy_FraigMiterStatus(current)
		if status >= 0 {
			break
		}
		if params.fUseFraiging != 0 {
			options.NBTLimitNode = int32(float64(params.nFraigingLimitStart) * math.Pow(float64(params.nFraigingLimitMulti), float64(iteration)))
			options.NBTLimitMiter = 1 + divide(int32(float64(params.nMiteringLimitStart)*math.Pow(float64(params.nMiteringLimitMulti), float64(iteration))))
			var conflicts, inspections int64
			replace(Ivy_FraigPerform_int(current, &options, params.nTotalBacktrackLimit, params.nTotalInspectLimit, &conflicts, &inspections))
			status = Ivy_FraigMiterStatus(current)
			if status >= 0 {
				break
			}
			params.nTotalBacktracksMade += conflicts
			params.nTotalInspectsMade += inspections
		}
		if params.nTotalBacktrackLimit > 0 && params.nTotalBacktracksMade >= params.nTotalBacktrackLimit || params.nTotalInspectLimit > 0 && params.nTotalInspectsMade >= params.nTotalInspectLimit {
			return -1
		}
	}
	if status == 0 && current.pData == nil {
		current.pData = make([]int32, len(current.vPis))
	}
	return status
}
