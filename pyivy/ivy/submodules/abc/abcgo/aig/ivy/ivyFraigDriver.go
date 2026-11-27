// FRAIG sweep, miter and resource-limited prove drivers from ivyFraig.c.
package ivy

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
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

// Ivy_FraigPrint is ivyFraig.c's.
func Ivy_FraigPrint(p *Ivy_FraigMan_t) {
	memory := float64(Ivy_ManObjNum(p.pManAig)) * float64(p.nSimWords) * 4 / (1 << 20)
	fmt.Printf("SimWords = %d. Rounds = %d. Mem = %0.2f MB.  ", p.nSimWords, p.nSimRounds, memory)
	fmt.Printf("Classes: Beg = %d. End = %d.\n", p.nClassesBeg, p.nClassesEnd)
	fmt.Printf("Proof = %d. Counter-example = %d. Fail = %d. FailReal = %d. Zero = %d.\n",
		p.nSatProof, p.nSatCallsSat, p.nSatFails, p.nSatFailsReal, p.nClassesZero)
	fmt.Printf("Final = %d. Miter = %d. Total = %d. Mux = %d. (Exor = %d.) SatVars = %d.\n",
		Ivy_ManNodeNum(p.pManFraig), p.nNodesMiter, Ivy_ManNodeNum(p.pManAig), 0, 0, p.nSatVars)
	if p.pSat != nil {
		bsat.Sat_SolverPrintStats(os.Stdout, p.pSat)
	}
	for _, row := range []struct {
		name string
		d    time.Duration
	}{{"AIG simulation  ", p.timeSim}, {"AIG traversal   ", p.timeTrav}, {"SAT solving     ", p.timeSat}, {"    Unsat       ", p.timeSatUnsat}, {"    Sat         ", p.timeSatSat}, {"    Fail        ", p.timeSatFail}, {"Class refining  ", p.timeRef}, {"TOTAL RUNTIME   ", p.timeTotal}} {
		util.Abc_PrintTime(1, row.name, row.d.Nanoseconds()) // ABC_PRT
	}
	if p.time1 != 0 {
		util.Abc_PrintTime(1, "time1           ", p.time1.Nanoseconds())
	}
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
func Ivy_FraigMiterPrint(p *Ivy_Man_t, label string, clk int64, verbose int32) {
	if verbose != 0 {
		fmt.Printf("Nodes = %7d.  Levels = %4d.  ", Ivy_ManNodeNum(p), Ivy_ManLevels(p))
		util.Abc_PrintTime(1, label, util.Abc_Clock()-clk)
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
	// C divides by Ivy_ManPoNum directly; a miter always has outputs
	divide := func(limit int32) int32 { return limit / max(int32(1), Ivy_ManPoNum(current)) }
	if params.fVerbose != 0 {
		fmt.Printf("RESOURCE LIMITS: Iterations = %d. Rewriting = %s. Fraiging = %s.\n",
			params.nItersMax, ivyYesNo(params.fUseRewriting), ivyYesNo(params.fUseFraiging))
		fmt.Printf("Miter = %d (%3.1f).  Rwr = %d (%3.1f).  Fraig = %d (%3.1f).  Last = %d.\n",
			params.nMiteringLimitStart, params.nMiteringLimitMulti,
			params.nRewritingLimitStart, params.nRewritingLimitMulti,
			params.nFraigingLimitStart, params.nFraigingLimitMulti, params.nMiteringLimitLast)
	}
	// if SAT only, solve without iteration
	if params.fUseRewriting == 0 && params.fUseFraiging == 0 {
		clk := util.Abc_Clock()
		options.NBTLimitMiter = divide(params.nMiteringLimitLast)
		replace(Ivy_FraigMiter(current, &options))
		status := Ivy_FraigMiterStatus(current)
		Ivy_FraigMiterPrint(current, "SAT solving", clk, params.fVerbose)
		return status
	}
	if Ivy_ManNodeNum(current) < 500 {
		// run the first mitering
		clk := util.Abc_Clock()
		options.NBTLimitMiter = divide(params.nMiteringLimitStart)
		replace(Ivy_FraigMiter(current, &options))
		status := Ivy_FraigMiterStatus(current)
		Ivy_FraigMiterPrint(current, "SAT solving", clk, params.fVerbose)
		if status >= 0 {
			return status
		}
	}
	// check the current resource limits
	status := int32(-1)
	var conflicts, inspections int64
	for iteration := int32(0); iteration < params.nItersMax; iteration++ {
		if params.fVerbose != 0 {
			fmt.Printf("ITERATION %2d : Confs = %6d. FraigBTL = %3d. \n", iteration+1,
				int32(float64(params.nMiteringLimitStart)*math.Pow(float64(params.nMiteringLimitMulti), float64(iteration))),
				int32(float64(params.nFraigingLimitStart)*math.Pow(float64(params.nFraigingLimitMulti), float64(iteration))))
		}
		// rewriting is commented out in ivyFraig.c (bug in Ivy_NodeFindCutsAll)
		if status >= 0 {
			break
		}
		// catch the situation when ref pattern detects the bug
		status = Ivy_FraigMiterStatus(current)
		if status >= 0 {
			break
		}
		// try fraiging followed by mitering
		if params.fUseFraiging != 0 {
			clk := util.Abc_Clock()
			options.NBTLimitNode = int32(float64(params.nFraigingLimitStart) * math.Pow(float64(params.nFraigingLimitMulti), float64(iteration)))
			options.NBTLimitMiter = 1 + divide(int32(float64(params.nMiteringLimitStart)*math.Pow(float64(params.nMiteringLimitMulti), float64(iteration))))
			replace(Ivy_FraigPerform_int(current, &options, params.nTotalBacktrackLimit, params.nTotalInspectLimit, &conflicts, &inspections))
			status = Ivy_FraigMiterStatus(current)
			Ivy_FraigMiterPrint(current, "Fraiging   ", clk, params.fVerbose)
		}
		if status >= 0 {
			break
		}
		// add to the number of backtracks and inspects
		params.nTotalBacktracksMade += conflicts
		params.nTotalInspectsMade += inspections
		// check if global resource limit is reached
		if params.nTotalBacktrackLimit != 0 && params.nTotalBacktracksMade >= params.nTotalBacktrackLimit ||
			params.nTotalInspectLimit != 0 && params.nTotalInspectsMade >= params.nTotalInspectLimit {
			fmt.Printf("Reached global limit on conflicts/inspects. Quitting.\n")
			return -1
		}
	}
	// assign the model if it was proved by rewriting (const 1 miter)
	if status == 0 && current.pData == nil {
		current.pData = make([]int32, len(current.vPis))
	}
	return status
}

func ivyYesNo(v int32) string {
	if v != 0 {
		return "yes"
	}
	return "no"
}
