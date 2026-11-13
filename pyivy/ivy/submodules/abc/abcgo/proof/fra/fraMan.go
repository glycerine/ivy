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
	p := &Fra_Man_t{PManAig: graph, PPars: pars, NSizeAlloc: int32(len(graph.Nodes())), NFramesAll: pars.NFramesK + 1}
	bits := (int32(graph.InputCount())-graph.RegisterCount())*p.NFramesAll + graph.RegisterCount()
	p.NPatWords = (bits + 31) / 32
	p.PPatWords = make([]uint32, p.NPatWords)
	p.VPiVars = vec.Vec_PtrAlloc(100)
	p.PCla = Fra_ClassesStart(graph)
	p.PMemFraig = make([]aig.Edge, int(p.NSizeAlloc)*int(p.NFramesAll))
	for i := range p.PMemFraig {
		p.PMemFraig[i] = aig.InvalidEdge
	}
	aig.Aig_ManRandom(1)
	return p
}

// Fra_ManClean resets the SAT fanin and variable maps for a new solver run.
func Fra_ManClean(p *Fra_Man_t, nodesMax int32) {
	if nodesMax < 0 {
		panic("fra: invalid map size")
	}
	if p.NMemAlloc < nodesMax {
		p.NMemAlloc = nodesMax + 5000
		p.PMemFanins = make([][]aig.Edge, p.NMemAlloc)
		p.PMemSatNums = make([]int32, p.NMemAlloc)
	} else {
		clear(p.PMemFanins)
		clear(p.PMemSatNums)
	}
}

func Fra_ManPrepareComb(p *Fra_Man_t) *aig.Graph {
	if p.PManFraig != nil {
		panic("fra: reduced graph already exists")
	}
	graph := aig.NewGraph()
	graph.Name, graph.Spec, graph.Assertions = p.PManAig.Name, p.PManAig.Spec, p.PManAig.Assertions
	Fra_ObjSetFraig(p, p.PManAig.Const1(), 0, graph.Const1())
	for _, input := range p.PManAig.Inputs() {
		Fra_ObjSetFraig(p, input, 0, graph.NewInput())
	}
	aig.Aig_ManSetRegNumForConstruction(graph, p.PManAig.RegisterCount())
	p.NMemAlloc = p.NSizeAlloc
	p.PMemFanins = make([][]aig.Edge, p.NMemAlloc)
	p.PMemSatNums = make([]int32, p.NMemAlloc)
	return graph
}

func Fra_ManFinalizeComb(p *Fra_Man_t) {
	for _, output := range p.PManAig.Outputs() {
		p.PManFraig.AddOutput(Fra_ObjChild0Fra(p, output, 0))
	}
	aig.Aig_ManSetRegNum(p.PManFraig, p.PManAig.RegisterCount())
	aig.Aig_ManCleanMarkB(p.PManFraig)
}

// Fra_ManStop returns the original-node-by-frame mapping formerly stored in
// Aig_Man_t.pObjCopies. Callers that use it after shutdown must retain it;
// entries are indexed by node ID * nFramesAll + frame.
func Fra_ManStop(p *Fra_Man_t) []aig.Edge {
	if p.PPars.FVerbose != 0 {
		Fra_ManPrint(p)
	}
	mapping := p.PMemFraig
	p.PMemFraig = nil
	p.PMemFanins, p.PMemSatNums, p.PPatWords = nil, nil, nil
	p.PSat, p.PCla, p.PSml = nil, nil, nil
	p.VTimeouts, p.VPiVars, p.VCex, p.VOneHots = nil, nil, nil, nil
	return mapping
}

func Fra_ManPrint(p *Fra_Man_t) {
	size := float64(len(p.PManAig.Nodes())) * float64(p.PSml.WordsTotal*4+6*8) / (1 << 20)
	denom := p.NLitsBeg
	if denom == 0 {
		denom = 1
	}
	fmt.Printf("SimWord = %d. Round = %d.  Mem = %0.2f MB.  LitBeg = %d.  LitEnd = %d. (%6.2f %%).\n", p.PPars.NSimWords, p.PSml.Rounds, size, p.NLitsBeg, p.NLitsEnd, 100*float64(p.NLitsEnd)/float64(denom))
	fmt.Printf("Proof = %d. Cex = %d. Fail = %d. FailReal = %d. C-lim = %d. ImpRatio = %6.2f %%\n", p.NSatProof, p.NSatCallsSat, p.NSatFails, p.NSatFailsReal, p.PPars.NBTLimitNode, Fra_ImpComputeStateSpaceRatio(p))
	nodeDenom, regDenom := p.NNodesBeg, p.NRegsBeg
	if nodeDenom == 0 {
		nodeDenom = 1
	}
	if regDenom == 0 {
		regDenom = 1
	}
	fmt.Printf("NBeg = %d. NEnd = %d. (Gain = %6.2f %%).  RBeg = %d. REnd = %d. (Gain = %6.2f %%).\n", p.NNodesBeg, p.NNodesEnd, 100*float64(p.NNodesBeg-p.NNodesEnd)/float64(nodeDenom), p.NRegsBeg, p.NRegsEnd, 100*float64(p.NRegsBeg-p.NRegsEnd)/float64(regDenom))
	if p.PSat != nil {
		bsat.Sat_SolverPrintStats(os.Stdout, p.PSat)
	}
	if p.PPars.FUse1Hot != 0 {
		Fra_OneHotEstimateCoverage(p, p.VOneHots)
	}
	printTime := func(name string, elapsed time.Duration) { fmt.Printf("%-16s = %.2f sec\n", name, elapsed.Seconds()) }
	printTime("AIG simulation", p.TimeSim)
	printTime("AIG traversal", p.TimeTrav)
	if p.TimeRwr != 0 {
		printTime("AIG rewriting", p.TimeRwr)
	}
	printTime("SAT solving", p.TimeSat)
	printTime("    Unsat", p.TimeSatUnsat)
	printTime("    Sat", p.TimeSatSat)
	printTime("    Fail", p.TimeSatFail)
	printTime("Class refining", p.TimeRef)
	printTime("TOTAL RUNTIME", p.TimeTotal)
	if p.Time1 != 0 {
		printTime("time1", p.Time1)
	}
	if p.NSpeculs != 0 {
		fmt.Printf("Speculations = %d.\n", p.NSpeculs)
	}
}
