package cut

import (
	"fmt"
	"os"
	"time"
)

func Cut_ManStart(params *Cut_Params_t) *Cut_Man_t {
	if params == nil || params.NVarsMax < CUT_SIZE_MIN || params.NVarsMax > CUT_SIZE_MAX || params.NKeepMax < 1 {
		panic("cut: invalid parameters")
	}
	if params.FSeq != 0 {
		params.FFilter = 1
		if params.FTruth != 0 && params.NVarsMax > 5 {
			params.FTruth = 0
		}
	}
	return &Cut_Man_t{params: params, budget: -1, newCuts: map[int32]*Cut_Cut_t{}, oldCuts: map[int32]*Cut_Cut_t{}, tempCuts: map[int32]*Cut_Cut_t{}, owned: map[*Cut_Cut_t]struct{}{}, records: map[int32]Cut_Record{}, delays: map[int32]int32{}, best: map[int32]*Cut_Cut_t{}}
}
func Cut_ManStop(p *Cut_Man_t) {
	if p == nil {
		return
	}
	for c := range p.owned {
		c.live = false
		c.owner = nil
	}
	p.owned = nil
	p.newCuts = nil
	p.oldCuts = nil
	p.tempCuts = nil
	p.records = nil
	p.best = nil
	p.fanCounts = nil
	p.attrs = nil
	p.delays = nil
	p.stopped = true
}
func Cut_ManPrintStats(p *Cut_Man_t) {
	s := p.stats
	fmt.Printf("Cut computation statistics:\nCurrent cuts = %d (trivial %d)\nPeak cuts = %d\nAllocated = %d Recycled = %d Filtered = %d\nNodes = %d Saturated = %d Mapping delay = %d\n", s.Current, s.Trivial, s.Peak, s.Allocated, s.Recycled, s.Filtered, s.Nodes, s.Limited, s.MappingDelay)
}
func Cut_ManPrintStatsToFile(p *Cut_Man_t, path string, total time.Duration) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "K=%d Cuts=%d Peak=%d Nodes=%d Delay=%d Time=%s\n", p.params.NVarsMax, p.stats.Allocated, p.stats.Peak, p.stats.Nodes, p.stats.MappingDelay, total)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Cut_ManSetFanoutCounts(p *Cut_Man_t, counts []int32) {
	p.fanCounts = append([]int32(nil), counts...)
}
func Cut_ManSetNodeAttrs(p *Cut_Man_t, attrs []int32) { p.attrs = append([]int32(nil), attrs...) }
func Cut_ManReadVarsMax(p *Cut_Man_t) int32           { return p.params.NVarsMax }
func Cut_ManReadParams(p *Cut_Man_t) *Cut_Params_t    { return p.params }
func Cut_ManReadNodeAttrs(p *Cut_Man_t) []int32       { return p.attrs }
func Cut_ManIncrementDagNodes(p *Cut_Man_t)           { p.stats.DagNodes++ }
