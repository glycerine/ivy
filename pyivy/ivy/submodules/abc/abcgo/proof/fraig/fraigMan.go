package fraig

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/msat"
)

// Prove_ParamsSetDefault initializes the equivalence-checking engine settings.
func Prove_ParamsSetDefault(p *Prove_Params_t) {
	*p = Prove_Params_t{fUseFraiging: 1, fUseRewriting: 1, nItersMax: 6,
		nMiteringLimitStart: 5000, nMiteringLimitMulti: 2,
		nRewritingLimitStart: 3, nRewritingLimitMulti: 1,
		nFraigingLimitStart: 2, nFraigingLimitMulti: 8,
		nBddSizeLimit: 1000000, fBddReorder: 1}
}

// Prove_ParamsPrint prints the source engine's parameter report.
func Prove_ParamsPrint(p *Prove_Params_t) { proveParamsPrint(os.Stdout, p) }

func proveParamsPrint(w io.Writer, p *Prove_Params_t) {
	yes := func(v int32) string {
		if v != 0 {
			return "yes"
		}
		return "no"
	}
	fmt.Fprintln(w, "CEC enging parameters:")
	fmt.Fprintf(w, "Fraiging enabled: %s\n", yes(p.fUseFraiging))
	fmt.Fprintf(w, "Rewriting enabled: %s\n", yes(p.fUseRewriting))
	fmt.Fprintf(w, "BDD construction enabled: %s\n", yes(p.fUseBdds))
	fmt.Fprintf(w, "Verbose output enabled: %s\n", yes(p.fVerbose))
	fmt.Fprintf(w, "Solver iterations: %d\n", p.nItersMax)
	fmt.Fprintf(w, "Starting mitering limit: %d\n", p.nMiteringLimitStart)
	fmt.Fprintf(w, "Multiplicative coeficient for mitering: %.2f\n", p.nMiteringLimitMulti)
	fmt.Fprintf(w, "Starting number of rewriting iterations: %d\n", p.nRewritingLimitStart)
	fmt.Fprintf(w, "Multiplicative coeficient for rewriting: %.2f\n", p.nRewritingLimitMulti)
	// Preserve the source report's chosen fields, including its two fraiging labels.
	fmt.Fprintf(w, "Starting number of conflicts in fraiging: %.2f\n", p.nFraigingLimitMulti)
	fmt.Fprintf(w, "Multiplicative coeficient for fraiging: %.2f\n", p.nRewritingLimitMulti)
	fmt.Fprintf(w, "BDD size limit for bailing out: %d\n", p.nBddSizeLimit)
	fmt.Fprintf(w, "BDD reordering enabled: %s\n", yes(p.fBddReorder))
	fmt.Fprintf(w, "Last-gasp mitering limit: %d\n", p.nMiteringLimitLast)
	fmt.Fprintf(w, "Total conflict limit: %d\n", int32(p.nTotalBacktrackLimit))
	fmt.Fprintf(w, "Total inspection limit: %d\n", int32(p.nTotalInspectLimit))
	fmt.Fprintln(w, "Parameter dump complete.")
}

func Fraig_ParamsSetDefault(p *Fraig_Params_t) {
	*p = Fraig_Params_t{NPatsRand: FRAIG_PATTERNS_RANDOM, NPatsDyna: FRAIG_PATTERNS_DYNAMIC, NBTLimit: 99, NSeconds: 20, FFuncRed: 1, FFeedBack: 1, FDist1Pats: 1, FTryProve: 1}
}
func Fraig_ParamsSetDefaultFull(p *Fraig_Params_t) {
	Fraig_ParamsSetDefault(p)
	p.NBTLimit = -1
	p.FDoSparse = 1
	p.FTryProve = 0
}
func Fraig_ManCreate(params *Fraig_Params_t) *Fraig_Man_t {
	aig.Aig_ManRandom(1)
	if params == nil {
		params = &Fraig_Params_t{}
		Fraig_ParamsSetDefault(params)
	}
	if params.NPatsRand < 128 {
		params.NPatsRand = 128
	}
	if params.NPatsRand > 32768 {
		params.NPatsRand = 32768
	}
	if params.NPatsDyna < 128 {
		params.NPatsDyna = 128
	}
	if params.NPatsDyna > 32768 {
		params.NPatsDyna = 32768
	}
	if params.FFuncRed == 0 {
		params.NPatsRand = 128
		params.NPatsDyna = 128
	}
	p := &Fraig_Man_t{
		nWordsRand: (params.NPatsRand + 31) / 32,
		nWordsDyna: (params.NPatsDyna + 31) / 32,
		nBTLimit:   params.NBTLimit,
		nSeconds:   params.NSeconds,
		fFuncRed:   params.FFuncRed,
		fFeedBack:  params.FFeedBack,
		fDist1Pats: params.FDist1Pats,
		fDoSparse:  params.FDoSparse,
		fChoicing:  params.FChoicing,
		fTryProve:  params.FTryProve,
		fVerbose:   params.FVerbose,
		fVerboseP:  params.FVerboseP,
		nInspLimit: params.NInspLimit,
	}
	p.mmNodes = Fraig_MemFixedStart(int32(unsafe.Sizeof(Fraig_Node_t{})))
	p.mmSims = Fraig_MemFixedStart(4 * (p.nWordsRand + p.nWordsDyna))
	p.vInputs = Fraig_NodeVecAlloc(1000)
	p.vOutputs = Fraig_NodeVecAlloc(1000)
	p.vNodes = Fraig_NodeVecAlloc(1000)
	p.pTableS = Fraig_HashTableCreate(1000)
	p.pTableF = Fraig_HashTableCreate(1000)
	p.pTableF0 = Fraig_HashTableCreate(1000)
	p.pConst1 = Fraig_NodeCreateConst(p)
	Fraig_FeedBackInit(p)
	p.vProj = msat.Msat_IntVecAlloc(10)
	p.nTravIds = 1
	p.nTravIds2 = 1
	return p
}
func Fraig_ManFree(p *Fraig_Man_t) {
	if p.fVerbose != 0 {
		if p.fChoicing != 0 {
			Fraig_ManReportChoices(p)
		}
		Fraig_ManPrintStats(p)
	}
	for _, n := range p.vNodes.pArray {
		if n.vFanins != nil {
			Fraig_NodeVecFree(n.vFanins)
			n.vFanins = nil
		}
	}
	if p.vInputs != nil {
		Fraig_NodeVecFree(p.vInputs)
	}
	if p.vNodes != nil {
		Fraig_NodeVecFree(p.vNodes)
	}
	if p.vOutputs != nil {
		Fraig_NodeVecFree(p.vOutputs)
	}
	if p.pTableS != nil {
		Fraig_HashTableFree(p.pTableS)
	}
	if p.pTableF != nil {
		Fraig_HashTableFree(p.pTableF)
	}
	if p.pTableF0 != nil {
		Fraig_HashTableFree(p.pTableF0)
	}
	if p.pSat != nil {
		msat.Msat_SolverFree(p.pSat)
	}
	if p.vProj != nil {
		msat.Msat_IntVecFree(p.vProj)
	}
	if p.vCones != nil {
		Fraig_NodeVecFree(p.vCones)
	}
	if p.vPatsReal != nil {
		msat.Msat_IntVecFree(p.vPatsReal)
	}
	Fraig_MemFixedStop(p.mmNodes, 0)
	Fraig_MemFixedStop(p.mmSims, 0)
	*p = Fraig_Man_t{}
}
func Fraig_ManCreateSolver(p *Fraig_Man_t) {
	assert(p.pSat == nil, "p.pSat == nil")
	p.pSat = msat.Msat_SolverAlloc(500, 1, 1, 1, 1, 0)
	p.vVarsInt = msat.Msat_SolverReadConeVars(p.pSat)
	p.vAdjacents = msat.Msat_SolverReadAdjacents(p.pSat)
	p.vVarsUsed = msat.Msat_SolverReadVarsUsed(p.pSat)
}
func Fraig_ManPrintStats(p *Fraig_Man_t) {
	memory := float64(Fraig_NodeVecReadSize(p.vInputs)+Fraig_NodeVecReadSize(p.vNodes)) * float64(unsafe.Sizeof(Fraig_Node_t{})+uintptr(4*(p.nWordsRand+p.nWordsDyna))) / (1 << 20)
	fmt.Printf("Words: Random = %d. Dynamic = %d. Used = %d. Memory = %0.2f MB.\n", p.nWordsRand, p.nWordsDyna, p.iWordPerm, memory)
	fmt.Printf("Proof = %d. Counter-example = %d. Fail = %d. FailReal = %d. Zero = %d.\n", p.nSatProof, p.nSatCounter, p.nSatFails, p.nSatFailsReal, p.nSatZeros)
	fmt.Printf("Nodes: Final = %d. Total = %d. Mux = %d. (Exor = %d.) ClaVars = %d.\n", Fraig_CountNodes(p, 0), Fraig_NodeVecReadSize(p.vNodes), Fraig_ManCountMuxes(p), Fraig_ManCountExors(p), p.nVarsClauses)
	if p.pSat != nil {
		msat.Msat_SolverPrintStats(p.pSat)
	}
	for _, item := range []struct {
		name  string
		value int64
	}{{"AIG simulation", p.timeSims}, {"AIG traversal", p.timeTrav}, {"Solver feedback", p.timeFeed}, {"SAT solving", p.timeSat}, {"Network update", p.timeToNet}, {"TOTAL RUNTIME", p.timeTotal}, {"time1", p.time1}, {"time2", p.time2}, {"time3", p.time3}, {"time4", p.time4}} {
		if item.value != 0 {
			fmt.Printf("%s: %.2f sec\n", item.name, float64(item.value)/1e9)
		}
	}
}

// Fraig_UtilInfoAlloc returns Go-owned simulation rows, replacing the C vector
// of pointers cast to unsigned arrays. Go zero-initializes each row.
func Fraig_UtilInfoAlloc(size, words, clean int32) [][]uint32 {
	assert(size > 0 && words > 0, "positive simulation dimensions")
	data := make([]uint32, size*words)
	rows := make([][]uint32, size)
	for i := int32(0); i < size; i++ {
		rows[i] = data[i*words : (i+1)*words]
	}
	return rows
}
func Fraig_ManGetSimInfo(p *Fraig_Man_t) [][]uint32 {
	words := p.nWordsRand + p.iWordStart
	info := Fraig_UtilInfoAlloc(Fraig_NodeVecReadSize(p.vNodes), words, 0)
	for i, node := range p.vNodes.pArray {
		assert(int32(i) == node.Num, "node number matches position")
		copy(info[i], node.puSimR)
		copy(info[i][p.nWordsRand:], node.puSimD[:p.iWordStart])
	}
	return info
}
func Fraig_ManCheckClauseUsingSimInfo(p *Fraig_Man_t, a, b *Fraig_Node_t) int32 {
	invA, invB := !fraigNodeSimComplement(a), !fraigNodeSimComplement(b)
	a, b = Fraig_Regular(a), Fraig_Regular(b)
	assert(a != b, "distinct clause nodes")
	for _, pair := range []struct {
		x, y []uint32
		n    int32
	}{{a.puSimR, b.puSimR, p.nWordsRand}, {a.puSimD, b.puSimD, p.iWordStart}} {
		for i := int32(0); i < pair.n; i++ {
			x, y := pair.x[i], pair.y[i]
			if invA {
				x = ^x
			}
			if invB {
				y = ^y
			}
			if x&y != 0 {
				return 0
			}
		}
	}
	return 1
}

// Fraig_ManAddClause adds the given node literals to the existing SAT solver,
// as in the C implementation. Callers must ensure the variables are present.
func Fraig_ManAddClause(p *Fraig_Man_t, nodes []*Fraig_Node_t, n int32) {
	if p.pSat == nil {
		Fraig_ManCreateSolver(p)
	}
	msat.Msat_IntVecClear(p.vProj)
	for _, edge := range nodes[:n] {
		msat.Msat_IntVecPush(p.vProj, fraigLit(Fraig_Regular(edge), Fraig_IsComplement(edge)))
	}
	assert(msat.Msat_SolverAddClause(p.pSat, p.vProj) != 0, "clause accepted")
}
