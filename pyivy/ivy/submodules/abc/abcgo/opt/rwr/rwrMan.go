package rwr

// Rewrite manager lifecycle and statistics from rwrMan.c.
import (
	"fmt"
	"os"
	"time"

	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/extra/npn"
)

func Rwr_ManStart(precompute int32) *Rwr_Man_t {
	started := time.Now()
	canons, phases, perms, classes := npn.Extra_Truth4VarNPN()
	p := &Rwr_Man_t{nFuncs: 1 << 16, puCanons: canons, pPhases: phases, pPerms: perms, pMap: classes, pTable: make([]*Rwr_Node_t, 1<<16)}
	p.pPractical = Rwr_ManGetPractical(p)
	for _, truth := range []uint32{0, 0xaaaa, 0xcccc, 0xf0f0, 0xff00} {
		Rwr_ManAddVar(p, truth, precompute)
	}
	p.nClasses = 5
	p.nTravIds = 1
	for _, perm := range npn.Permutations() {
		p.pPerms4 = append(p.pPerms4, append([]byte(nil), perm[:]...))
	}
	p.vLevNums = make([]int32, 0, 50)
	p.vFanins = make([]any, 0, 50)
	p.vFaninsCur = make([]any, 0, 50)
	p.vNodesTemp = make([]any, 0, 50)
	if precompute != 0 {
		Rwr_ManPrecompute(p)
		Rwr_ManWriteToArray(p)
	} else {
		Rwr_ManLoadFromArray(p, 0)
		Rwr_ManPreprocess(p)
	}
	p.timeStart = int64(time.Since(started))
	return p
}
func Rwr_ManStop(p *Rwr_Man_t) {
	if p != nil {
		*p = Rwr_Man_t{}
	}
}
func Rwr_ManPrintStats(p *Rwr_Man_t) {
	classes := 0
	for _, n := range p.nScores {
		if n > 0 {
			classes++
		}
	}
	fmt.Printf("Rewriting statistics:\nTotal cuts tries  = %8d.\nBad cuts found    = %8d.\nTotal subgraphs   = %8d.\nUsed NPN classes  = %8d.\nNodes considered  = %8d.\nNodes rewritten   = %8d.\nGain              = %8d. (%6.2f %%).\n", p.nCutsGood, p.nCutsBad, p.nSubgraphs, classes, p.nNodesConsidered, p.nNodesRewritten, p.nNodesBeg-p.nNodesEnd, 100.0*float64(p.nNodesBeg-p.nNodesEnd)/float64(p.nNodesBeg))
	for _, row := range []struct {
		name  string
		value int64
	}{{"Start", p.timeStart}, {"Cuts", p.timeCut}, {"Resynthesis", p.timeRes}, {"    Mffc", p.timeMffc}, {"    Eval", p.timeEval}, {"Update", p.timeUpdate}, {"TOTAL", p.timeTotal}} {
		fmt.Printf("%-12s: %v\n", row.name, time.Duration(row.value))
	}
	fmt.Println()
}
func Rwr_ManPrintStatsFile(p *Rwr_Man_t) {
	f, err := os.OpenFile("stats.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(f, "%d %d %d %d\n", p.nCutsGood, p.nSubgraphs, p.nNodesRewritten, p.nNodesGained)
	if err = f.Close(); err != nil {
		panic(err)
	}
}
func Rwr_ManReadDecs(p *Rwr_Man_t) *dec.Dec_Graph_t { return p.pGraph }
func Rwr_ManReadLeaves(p *Rwr_Man_t) []any          { return p.vFanins }
func Rwr_ManReadCompl(p *Rwr_Man_t) int32           { return p.fCompl }
func Rwr_ManAddTimeCuts(p *Rwr_Man_t, t int64)      { p.timeCut += t }
func Rwr_ManAddTimeUpdate(p *Rwr_Man_t, t int64)    { p.timeUpdate += t }
func Rwr_ManAddTimeTotal(p *Rwr_Man_t, t int64)     { p.timeTotal += t }

// Rwr_ManSetNodesBeg and Rwr_ManSetNodesEnd record pass boundary counts.
func Rwr_ManSetNodesBeg(p *Rwr_Man_t, count int32) { p.nNodesBeg = count }
func Rwr_ManSetNodesEnd(p *Rwr_Man_t, count int32) { p.nNodesEnd = count }
func Rwr_Precompute()                              { p := Rwr_ManStart(1); Rwr_ManStop(p) }
