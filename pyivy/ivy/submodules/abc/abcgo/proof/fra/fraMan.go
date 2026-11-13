package fra

// FRAIG manager setup and teardown, ported from ABC's fraMan.c.
import (
	"fmt"
	"os"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Fra_ManStart(graph *aig.Graph, pars *Fra_Par_t) *Fra_Man_t {
	if graph == nil || pars == nil || pars.NFramesK < 0 {
		panic("fra: invalid manager parameters")
	}
	p := &Fra_Man_t{pManAig: graph, pPars: pars, nSizeAlloc: int32(len(graph.Nodes())), nFramesAll: pars.NFramesK + 1}
	bits := (int32(graph.InputCount())-graph.RegisterCount())*p.nFramesAll + graph.RegisterCount()
	p.nPatWords = (bits + 31) / 32
	p.pPatWords = make([]uint32, p.nPatWords)
	p.vPiVars = vec.Vec_PtrAlloc(100)
	p.pCla = Fra_ClassesStart(graph)
	p.pMemFraig = make([]aig.Edge, int(p.nSizeAlloc)*int(p.nFramesAll))
	for i := range p.pMemFraig {
		p.pMemFraig[i] = aig.InvalidEdge
	}
	aig.Aig_ManRandom(1)
	return p
}

// Fra_ManClean resets the SAT fanin and variable maps for a new solver run.
func Fra_ManClean(p *Fra_Man_t, nodesMax int32) {
	if nodesMax < 0 {
		panic("fra: invalid map size")
	}
	if p.nMemAlloc < nodesMax {
		p.nMemAlloc = nodesMax + 5000
		p.pMemFanins = make([][]aig.Edge, p.nMemAlloc)
		p.pMemSatNums = make([]int32, p.nMemAlloc)
	} else {
		clear(p.pMemFanins)
		clear(p.pMemSatNums)
	}
}

func Fra_ManPrepareComb(p *Fra_Man_t) *aig.Graph {
	if p.pManFraig != nil {
		panic("fra: reduced graph already exists")
	}
	graph := aig.NewGraph()
	graph.Name, graph.Spec, graph.Assertions = p.pManAig.Name, p.pManAig.Spec, p.pManAig.Assertions
	Fra_ObjSetFraig(p, p.pManAig.Const1(), 0, graph.Const1())
	for _, input := range p.pManAig.Inputs() {
		Fra_ObjSetFraig(p, input, 0, graph.NewInput())
	}
	aig.Aig_ManSetRegNumForConstruction(graph, p.pManAig.RegisterCount())
	p.nMemAlloc = p.nSizeAlloc
	p.pMemFanins = make([][]aig.Edge, p.nMemAlloc)
	p.pMemSatNums = make([]int32, p.nMemAlloc)
	return graph
}

func Fra_ManFinalizeComb(p *Fra_Man_t) {
	for _, output := range p.pManAig.Outputs() {
		p.pManFraig.AddOutput(Fra_ObjChild0Fra(p, output, 0))
	}
	aig.Aig_ManSetRegNum(p.pManFraig, p.pManAig.RegisterCount())
	aig.Aig_ManCleanMarkB(p.pManFraig)
}

// Fra_ManStop returns the original-node-by-frame mapping formerly stored in
// Aig_Man_t.pObjCopies. Callers that use it after shutdown must retain it;
// entries are indexed by node ID * nFramesAll + frame.
func Fra_ManStop(p *Fra_Man_t) []aig.Edge {
	if p.pPars.FVerbose != 0 {
		Fra_ManPrint(p)
	}
	mapping := p.pMemFraig
	p.pMemFraig = nil
	p.pMemFanins, p.pMemSatNums, p.pPatWords = nil, nil, nil
	p.pSat, p.pCla, p.pSml = nil, nil, nil
	p.vTimeouts, p.vPiVars, p.vCex, p.vOneHots = nil, nil, nil, nil
	return mapping
}

func Fra_ManPrint(p *Fra_Man_t) {
	size := float64(len(p.pManAig.Nodes())) * float64(p.pSml.wordsTotal*4+6*8) / (1 << 20)
	denom := p.nLitsBeg
	if denom == 0 {
		denom = 1
	}
	fmt.Printf("SimWord = %d. Round = %d.  Mem = %0.2f MB.  LitBeg = %d.  LitEnd = %d. (%6.2f %%).\n", p.pPars.NSimWords, p.pSml.rounds, size, p.nLitsBeg, p.nLitsEnd, 100*float64(p.nLitsEnd)/float64(denom))
	fmt.Printf("Proof = %d. Cex = %d. Fail = %d. FailReal = %d. C-lim = %d. ImpRatio = %6.2f %%\n", p.nSatProof, p.nSatCallsSat, p.nSatFails, p.nSatFailsReal, p.pPars.NBTLimitNode, Fra_ImpComputeStateSpaceRatio(p))
	nodeDenom, regDenom := p.nNodesBeg, p.nRegsBeg
	if nodeDenom == 0 {
		nodeDenom = 1
	}
	if regDenom == 0 {
		regDenom = 1
	}
	fmt.Printf("NBeg = %d. NEnd = %d. (Gain = %6.2f %%).  RBeg = %d. REnd = %d. (Gain = %6.2f %%).\n", p.nNodesBeg, p.nNodesEnd, 100*float64(p.nNodesBeg-p.nNodesEnd)/float64(nodeDenom), p.nRegsBeg, p.nRegsEnd, 100*float64(p.nRegsBeg-p.nRegsEnd)/float64(regDenom))
	if p.pSat != nil {
		bsat.Sat_SolverPrintStats(os.Stdout, p.pSat)
	}
	if p.pPars.FUse1Hot != 0 {
		Fra_OneHotEstimateCoverage(p, p.vOneHots)
	}
	printTime := func(name string, elapsed time.Duration) { fmt.Printf("%-16s = %.2f sec\n", name, elapsed.Seconds()) }
	printTime("AIG simulation", p.timeSim)
	printTime("AIG traversal", p.timeTrav)
	if p.timeRwr != 0 {
		printTime("AIG rewriting", p.timeRwr)
	}
	printTime("SAT solving", p.timeSat)
	printTime("    Unsat", p.timeSatUnsat)
	printTime("    Sat", p.timeSatSat)
	printTime("    Fail", p.timeSatFail)
	printTime("Class refining", p.timeRef)
	printTime("TOTAL RUNTIME", p.timeTotal)
	if p.time1 != 0 {
		printTime("time1", p.time1)
	}
	if p.nSpeculs != 0 {
		fmt.Printf("Speculations = %d.\n", p.nSpeculs)
	}
}
