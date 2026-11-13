package fra

// Sequential SAT sweeping from ABC's fraInd.c.
import (
	"fmt"
	"os"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

func Fra_FraigInductionPart(g *aig.Graph, pars *Fra_Ssw_t) *aig.Graph {
	started := time.Now()
	size, verbose := pars.nPartSize, pars.fVerbose
	pars.nPartSize, pars.fVerbose = 0, 0
	defer func() { pars.nPartSize, pars.fVerbose = size, verbose }()
	var parts [][]int32
	if len(g.ClockDomains) > 0 {
		for _, domain := range g.ClockDomains {
			if size != 0 && len(domain) > int(size) {
				aig.Aig_ManPartDivide(&parts, domain, size, pars.nOverSize)
			} else {
				parts = append(parts, append([]int32(nil), domain...))
			}
		}
	} else {
		parts = aig.Aig_ManRegPartitionSimple(g, size, pars.nOverSize)
	}
	aig.Aig_ManReprStart(g, int32(len(g.Nodes())))
	for i, part := range parts {
		var pis, regs int32
		var mapping []int32
		temp := aig.Aig_ManRegCreatePart(g, part, &pis, &regs, &mapping)
		fraProjectPartitionOneHots(g, temp, verbose)
		result := Fra_FraigInduction(temp, pars)
		if result == nil {
			return nil
		}
		classes := aig.Aig_TransferMappedClasses(g, temp, mapping)
		if verbose != 0 {
			fmt.Printf("%3d : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d. It = %3d. Cl = %5d.\n", i, len(part), temp.InputCount()-len(part), pis, regs, aig.Aig_ManNodeNum(temp), pars.nIters, classes)
		}
	}
	result := aig.Aig_ManDupRepr(g, 0)
	aig.Aig_ManSeqCleanup(result)
	if verbose != 0 {
		fmt.Printf("Total time = %.2f sec\n", time.Since(started).Seconds())
	}
	return result
}

func fraProjectPartitionOneHots(source, part *aig.Graph, verbose int32) {
	if len(source.OneHotGroups) != 0 {
		part.OneHotInputGroups = aig.Aig_ManRegProjectOnehots(source, part, source.OneHotGroups, verbose)
	}
}

func Fra_FraigInduction(g *aig.Graph, params *Fra_Ssw_t) *aig.Graph {
	started := time.Now()
	if aig.Aig_ManNodeNum(g) == 0 {
		params.nIters = 0
		aig.Aig_ManReprStart(g, int32(len(g.Nodes())))
		return aig.Aig_ManDupOrdered(g)
	}
	if g.RegisterCount() <= 0 || params.nFramesK <= 0 {
		panic("fra: induction requires registers and frames")
	}
	if params.fWriteImps != 0 && params.nPartSize > 0 {
		params.nPartSize = 0
		fmt.Println("Partitioning was disabled to allow implication writing.")
	}
	if (params.nPartSize > 0 && params.nPartSize < g.RegisterCount()) || len(g.ClockDomains) > 0 {
		return Fra_FraigInductionPart(g, params)
	}
	var pars Fra_Par_t
	Fra_ParamsDefaultSeq(&pars)
	pars.NFramesP, pars.NFramesK, pars.NMaxImps, pars.NMaxLevs = params.nFramesP, params.nFramesK, params.nMaxImps, params.nMaxLevs
	pars.FVerbose, pars.FRewrite, pars.FLatchCorr = params.fVerbose, params.fRewrite, params.fLatchCorr
	pars.FUseImps, pars.FWriteImps, pars.FUse1Hot = params.fUseImps, params.fWriteImps, params.fUse1Hot
	if pars.FUse1Hot != 0 && (pars.NFramesP > 0 || pars.NFramesK > 1) {
		panic("fra: one-hot induction supports one frame without prefix")
	}
	p := Fra_ManStart(g, &pars)
	pars.NBTLimitNode = 0
	p.pSml = Fra_SmlSimulateSeq(g, pars.NFramesP, 32, 1, 1)
	p.pCla.simulation = p.pSml
	Fra_ClassesPrepare(p.pCla, pars.FLatchCorr, pars.NMaxLevs)
	if pars.FUse1Hot != 0 {
		p.vOneHots = Fra_OneHotCompute(p, p.pSml)
	}
	p.pSml = Fra_SmlStart(g, 0, pars.NFramesK+1, pars.NSimWords)
	if pars.FUseImps != 0 {
		p.pCla.vImps = Fra_ImpDerive(p, 5000000, pars.NMaxImps, pars.FLatchCorr)
	}
	timedOut := func() bool {
		return params.TimeLimit != 0 && time.Since(started) > time.Duration(float64(params.TimeLimit)*float64(time.Second))
	}
	if timedOut() {
		if params.fSilent == 0 {
			fmt.Println("Fra_FraigInduction(): Runtime limit exceeded.")
		}
		params.nIters = -1
		Fra_ManStop(p)
		return nil
	}
	Fra_BmcPerform(p, pars.NFramesP, pars.NFramesK+1)
	p.nLitsBeg = Fra_ClassesCountLits(p.pCla)
	p.nNodesBeg = aig.Aig_ManNodeNum(g)
	p.nRegsBeg = g.RegisterCount()
	p.pCla.fRefinement = 1
	iteration := int32(0)
	for ; p.pCla.fRefinement != 0; iteration++ {
		oldLits := Fra_ClassesCountLits(p.pCla)
		oldImps, oldHots := 0, int32(0)
		if p.pCla.vImps != nil {
			oldImps = vec.Vec_IntSize(p.pCla.vImps)
		}
		if p.vOneHots != nil {
			oldHots = Fra_OneHotCount(p, p.vOneHots)
		}
		if timedOut() {
			if params.fSilent == 0 {
				fmt.Println("Fra_FraigInduction(): Runtime limit exceeded.")
			}
			params.nIters = iteration
			Fra_ManStop(p)
			return nil
		}
		p.pCla.fRefinement = 0
		traversal := time.Now()
		p.pManFraig = Fra_FramesWithClasses(p)
		p.timeTrav += time.Since(traversal)
		if pars.FRewrite != 0 {
			Fra_FraigInductionRewrite(p)
		}
		var data *cnf.Cnf_Dat_t
		if pars.FUseImps != 0 {
			data = cnf.Cnf_DeriveSimple(p.pManFraig, p.pManFraig.RegisterCount())
		} else {
			data = cnf.Cnf_Derive(p.pManFraig, p.pManFraig.RegisterCount())
		}
		p.pSat = cnf.Cnf_DataWriteIntoSolver(data, 1, 0)
		p.nSatVars = int32(data.Variables)
		if p.pSat == nil {
			panic("fra: induction CNF is inconsistent")
		}
		if pars.FUseImps != 0 {
			Fra_ImpAddToSolver(p, p.pCla.vImps, data.VarNums)
			if p.pSat == nil {
				panic("fra: implication constraints contradict induction CNF")
			}
		}
		Fra_ManClean(p, int32(len(p.pManFraig.Nodes()))+aig.Aig_ManNodeNum(g))
		for _, edge := range p.pManFraig.LiveEdges() {
			number := data.VarNums[edge.NodeID()]
			if number >= 0 {
				Fra_ObjSetSatNum(p, edge, number)
				Fra_ObjSetFaninVec(p, edge, []aig.Edge{})
			}
		}
		cnf.Cnf_DataFree(data)
		if pars.FUse1Hot != 0 {
			Fra_OneHotAssume(p, p.vOneHots)
			if p.pSat == nil {
				panic("fra: one-hot constraints contradict induction CNF")
			}
		}
		p.nSatCallsRecent, p.nSatCallsSkipped = 0, 0
		if pars.FUse1Hot != 0 {
			Fra_OneHotCheck(p, p.vOneHots)
		}
		Fra_FraigSweep(p)
		p.pManFraig, p.pSat = nil, nil
		for i := range p.pMemFraig {
			p.pMemFraig[i] = aig.InvalidEdge
		}
		if p.vTimeouts != nil {
			panic("fra: SAT solver timed out during induction")
		}
		imps, hots := 0, int32(0)
		if p.pCla.vImps != nil {
			imps = vec.Vec_IntSize(p.pCla.vImps)
		}
		if p.vOneHots != nil {
			hots = Fra_OneHotCount(p, p.vOneHots)
		}
		if p.pCla.fRefinement != 0 && oldLits == Fra_ClassesCountLits(p.pCla) && oldImps == imps && oldHots == hots {
			fmt.Println("Fra_FraigInduction(): Internal error. The result may not verify.")
			// Do not return a potentially unsound reduction after the
			// refinement loop has failed to make observable progress.
			params.nIters = iteration
			Fra_ManStop(p)
			return nil
		}
	}
	traversal := time.Now()
	var result *aig.Graph
	if pars.FWriteImps != 0 && p.vOneHots != nil && Fra_OneHotCount(p, p.vOneHots) != 0 {
		name := ioa.Ioa_FileNameGenericAppend(g.Name, "_care.aig")
		fmt.Printf("Care one-hotness clauses will be written into file %q.\n", name)
		result = aig.Aig_ManDupOrdered(g)
		care := Fra_OneHotCreateExdc(p, p.vOneHots)
		if err := ioa.Ioa_WriteAiger(care, name, 0, 1); err != nil {
			panic(err)
		}
	} else {
		Fra_ClassesSelectRepr(p.pCla)
		Fra_ClassesCopyReprs(p.pCla, p.vTimeouts)
		result = aig.Aig_ManDupRepr(g, 0)
	}
	aig.Aig_ManSeqCleanup(result)
	p.timeTrav += time.Since(traversal)
	p.timeTotal = time.Since(started)
	p.nLitsEnd = Fra_ClassesCountLits(p.pCla)
	p.nNodesEnd = aig.Aig_ManNodeNum(result)
	p.nRegsEnd = result.RegisterCount()
	params.nIters = iteration
	Fra_ManStop(p)
	return result
}

func Fra_FraigInductionTest(filename string, pars *Fra_Ssw_t) int32 {
	graph := saig.Saig_ManReadBlif(filename)
	if graph == nil {
		return 0
	}
	result := Fra_FraigInduction(graph, pars)
	if result == nil {
		return 0
	}
	namedIDs, ok := graph.Data.(map[int32]int32)
	if !ok {
		panic("fra: BLIF reader omitted node IDs")
	}
	ids := make([]int32, len(graph.Nodes()))
	for i := range ids {
		ids[i] = -1
	}
	for original, current := range namedIDs {
		if current >= 0 && int(current) < len(ids) && (ids[current] < 0 || original < ids[current]) {
			ids[current] = original
		}
	}
	path := ioa.Ioa_FileNameGenericAppend(filename, ".pairs")
	file, err := os.Create(path)
	if err != nil {
		return 0
	}
	defer file.Close()
	count := 0
	for _, edge := range graph.LiveEdges() {
		repr := aig.Aig_ObjRepr(graph, edge)
		if repr == aig.InvalidEdge {
			continue
		}
		if ids[edge.NodeID()] < 0 || ids[repr.NodeID()] < 0 {
			panic("fra: missing BLIF node ID for equivalence pair")
		}
		sign := '+'
		if graph.Node(edge).Phase != graph.Node(repr).Phase {
			sign = '-'
		}
		fmt.Fprintf(file, "%d %d %c\n", ids[edge.NodeID()], ids[repr.NodeID()], sign)
		count++
	}
	if pars.fVerbose != 0 {
		fmt.Printf("Result: %d pairs of seq equiv nodes are written into file %q.\n", count, path)
	}
	return 1
}
