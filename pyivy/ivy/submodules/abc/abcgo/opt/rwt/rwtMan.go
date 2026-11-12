package rwt

import (
	"fmt"
	"os"
	"time"

	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
)

func Rwt_ManGlobalStart() {
	rwtGlobal.Lock()
	defer rwtGlobal.Unlock()
	if rwtGlobal.canons == nil {
		rwtGlobal.canons, rwtGlobal.phases, rwtGlobal.perms, rwtGlobal.mapping = extra.Extra_Truth4VarNPN()
	}
}

func Rwt_ManGlobalStop() {
	rwtGlobal.Lock()
	rwtGlobal.canons = nil
	rwtGlobal.phases = nil
	rwtGlobal.perms = nil
	rwtGlobal.mapping = nil
	rwtGlobal.Unlock()
}

func Rwt_ManStart(precompute int32) *Rwt_Man_t {
	start := time.Now()
	Rwt_ManGlobalStart()
	rwtGlobal.Lock()
	canons := rwtGlobal.canons
	phases := rwtGlobal.phases
	perms := rwtGlobal.perms
	mapping := rwtGlobal.mapping
	rwtGlobal.Unlock()
	p := &Rwt_Man_t{
		nFuncs:     1 << 16,
		puCanons:   canons,
		pPhases:    phases,
		pPerms:     perms,
		pMap:       mapping,
		pTable:     make([]*Rwt_Node_t, 1<<16),
		vFanins:    vec.Vec_PtrAlloc(50),
		vFaninsCur: vec.Vec_PtrAlloc(50),
		vLevNums:   vec.Vec_IntAlloc(50),
		vNodesTemp: vec.Vec_PtrAlloc(50),
		vForest:    make([]*Rwt_Node_t, 0, 1800),
		nTravIds:   1,
	}
	p.pPractical = Rwt_ManGetPractical(p)
	p.pPerms4 = extra.Extra_Permutations(4)
	Rwt_ManAddVar(p, 0x0000, precompute)
	Rwt_ManAddVar(p, 0xaaaa, precompute)
	Rwt_ManAddVar(p, 0xcccc, precompute)
	Rwt_ManAddVar(p, 0xf0f0, precompute)
	Rwt_ManAddVar(p, 0xff00, precompute)
	p.nClasses = 5
	if precompute == 0 {
		Rwt_ManLoadFromArray(p, 0)
		Rwt_ManPreprocess(p)
	}
	p.timeStart = time.Since(start)
	return p
}

func Rwt_ManStop(p *Rwt_Man_t) {
	if p == nil {
		return
	}
	for _, graph := range p.nodeGraphs {
		dec.Dec_GraphFree(graph)
	}
	p.nodeGraphs = nil
	p.vClasses = nil
	p.vNodesTemp = nil
	p.vForest = nil
	p.vLevNums = nil
	p.vFanins = nil
	p.vFaninsCur = nil
	p.pTable = nil
	p.pMapInv = nil
	p.pPractical = nil
	p.pPerms4 = nil
	p.pGraph = nil
	p.pCut = nil
}

func Rwt_ManPrintStats(p *Rwt_Man_t) {
	if p == nil {
		return
	}
	used := 0
	for _, score := range p.nScores {
		if score > 0 {
			used++
		}
	}
	fmt.Printf("Rewriting statistics:\n")
	fmt.Printf("Total cuts tries  = %8d.\n", p.nCutsGood)
	fmt.Printf("Bad cuts found    = %8d.\n", p.nCutsBad)
	fmt.Printf("Total subgraphs   = %8d.\n", p.nSubgraphs)
	fmt.Printf("Used NPN classes  = %8d.\n", used)
	fmt.Printf("Nodes considered  = %8d.\n", p.nNodesConsidered)
	fmt.Printf("Nodes rewritten   = %8d.\n", p.nNodesRewritten)
	fmt.Printf("Calculated gain   = %8d.\n", p.nNodesGained)
	for _, row := range []struct {
		name string
		d    time.Duration
	}{{"Start", p.timeStart}, {"Cuts", p.timeCut}, {"Truth", p.timeTruth}, {"Resynthesis", p.timeRes}, {"    Mffc", p.timeMffc}, {"    Eval", p.timeEval}, {"Update", p.timeUpdate}, {"TOTAL", p.timeTotal}} {
		fmt.Printf("%-12s = %9.2f sec\n", row.name, row.d.Seconds())
	}
}

func Rwt_ManPrintStatsFile(p *Rwt_Man_t) error {
	if p == nil {
		return nil
	}
	f, err := os.OpenFile("stats.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	_, writeErr := fmt.Fprintf(f, "%d %d %d %d\n", p.nCutsGood, p.nSubgraphs, p.nNodesRewritten, p.nNodesGained)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func Rwt_ManAddTimeCuts(p *Rwt_Man_t, elapsed time.Duration) {
	if p != nil {
		p.timeCut += elapsed
	}
}

func Rwt_ManAddTimeUpdate(p *Rwt_Man_t, elapsed time.Duration) {
	if p != nil {
		p.timeUpdate += elapsed
	}
}

func Rwt_ManAddTimeTotal(p *Rwt_Man_t, elapsed time.Duration) {
	if p != nil {
		p.timeTotal += elapsed
	}
}

func Rwt_Precompute() {
	p := Rwt_ManStart(1)
	Rwt_ManStop(p)
}
