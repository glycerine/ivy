package int

// Interpolation manager lifecycle from proof/int/intMan.c.
import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

func Inter_ManCreate(graph *Aig_Man_t, params *Inter_ManParams_t) *Inter_Man_t {
	if graph == nil || params == nil {
		panic("interpolation requires graph and parameters")
	}
	p := &Inter_Man_t{pAig: graph, nConfLimit: params.NBTLimit, fVerbose: params.FVerbose, pFileName: params.PFileName, vVarsAB: vec.Vec_IntAlloc(int(graph.RegisterCount()))}
	if params.FDropInvar != 0 {
		p.vInters = vec.Vec_PtrAlloc(100)
	}
	return p
}

func Inter_ManClean(p *Inter_Man_t) {
	if p == nil {
		return
	}
	if p.vInters != nil {
		vec.Vec_PtrClear(p.vInters)
	}
	if p.pCnfInter != nil {
		cnf.Cnf_DataFree(p.pCnfInter)
		p.pCnfInter = nil
	}
	if p.pCnfFrames != nil {
		cnf.Cnf_DataFree(p.pCnfFrames)
		p.pCnfFrames = nil
	}
	p.pInter, p.pFrames = nil, nil
}

func Inter_ManInterDump(p *Inter_Man_t, proved int32) {
	if p == nil || p.vInters == nil {
		return
	}
	graphs := make([]*aig.Graph, 0, vec.Vec_PtrSize(p.vInters))
	for _, value := range vec.Vec_PtrArray(p.vInters) {
		graph, ok := value.(*aig.Graph)
		if !ok || graph == nil {
			panic("invalid interpolant graph")
		}
		graphs = append(graphs, graph)
	}
	if len(graphs) == 0 {
		return
	}
	name := p.pFileName
	if name == "" {
		name = "invar.aig"
	}
	if err := ioa.Ioa_WriteAiger(aig.Aig_ManDupArray(graphs), name, 0, 0); err != nil {
		fmt.Printf("Could not write interpolants: %v\n", err)
		return
	}
	if proved != 0 {
		fmt.Printf("Inductive invariant is dumped into file %q.\n", name)
	} else {
		fmt.Printf("Interpolants are dumped into file %q.\n", name)
	}
}

func Inter_ManStop(p *Inter_Man_t, proved int32) {
	if p == nil {
		return
	}
	if p.fVerbose != 0 {
		p.timeOther = p.timeTotal - p.timeRwr - p.timeCnf - p.timeSat - p.timeInt - p.timeEqu
		fmt.Printf("Runtime statistics: total %s (rewriting %s, CNF %s, SAT %s, interpolation %s, containment %s, other %s)\n", time.Duration(p.timeTotal), time.Duration(p.timeRwr), time.Duration(p.timeCnf), time.Duration(p.timeSat), time.Duration(p.timeInt), time.Duration(p.timeEqu), time.Duration(p.timeOther))
	}
	if p.vInters != nil {
		Inter_ManInterDump(p, proved)
	}
	if p.pCnfAig != nil {
		cnf.Cnf_DataFree(p.pCnfAig)
	}
	p.pCnfAig, p.pAigTrans, p.pInterNew = nil, nil, nil
	Inter_ManClean(p)
	if p.vInters != nil {
		vec.Vec_PtrFree(p.vInters)
		p.vInters = nil
	}
	if p.vVarsAB != nil {
		vec.Vec_IntFree(p.vVarsAB)
		p.vVarsAB = nil
	}
}
