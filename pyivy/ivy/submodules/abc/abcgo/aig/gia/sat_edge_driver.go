package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

func Seg_ManConvertResult(p *Seg_Man_t) *Vec_Int_t {
	edges := vec.Vec_IntArray(p.vEdges)
	result := vec.Vec_IntAlloc(len(edges))
	for i := 0; i < len(edges); i += 2 {
		value, ok := p.pSat.Value(int32(i / 2))
		if !ok {
			panic("gia: SAT edge result unavailable")
		}
		if value {
			vec.Vec_IntPushTwo(result, edges[i], edges[i+1])
		}
	}
	return result
}

// Seg_ManComputeDelay searches for an edge assignment with minimum delay.
func Seg_ManComputeDelay(g *Gia_Man_t, initial, fanouts, two, verbose int32) {
	started := time.Now()
	p := Seg_ManAlloc(g, fanouts)
	defer Seg_ManStop(p)
	start := initial
	if start == 0 {
		start = p.DelayMax
	}
	if verbose != 0 {
		fmt.Printf("Running SatEdge with starting delay %d and edge %d (edge vars %d, total vars %d)\n", start, two+1, p.FirstVar, p.pSat.NumVars())
	}
	Seg_ManCreateCnf(p, two, verbose)
	var best *Vec_Int_t
	assumptions := make([]bsat.Literal, 0)
	for delay := p.DelayMax; delay >= 0; delay-- {
		for i := int32(0); i < Gia_ManCoNum(g); i++ {
			driver := Gia_ObjFaninId0p(g, Gia_ManCo(g, i))
			first := vec.Vec_IntEntry(p.vFirsts, int(driver))
			count := vec.Vec_IntEntry(p.vNvars, int(driver))
			if delay < count {
				assumptions = append(assumptions, bsat.Negative(first+delay))
			}
		}
		if delay > start {
			continue
		}
		before := p.pSat.Statistics().Conflicts
		status := p.pSat.Solve(assumptions...)
		if status != bsat.Satisfiable {
			if verbose != 0 {
				if status == bsat.Unsatisfiable {
					fmt.Printf("Proved UNSAT for delay %d. ", delay)
				} else {
					fmt.Printf("Resource limit reached for delay %d. ", delay)
				}
				fmt.Printf("Time = %.2f sec\n", time.Since(started).Seconds())
			}
			break
		}
		best = Seg_ManConvertResult(p)
		if verbose != 0 {
			fmt.Printf("Solution with delay %2d and %5d edges exists. Conf = %8d. Time = %.2f sec\n", delay, vec.Vec_IntSize(best)/2, p.pSat.Statistics().Conflicts-before, time.Since(started).Seconds())
		}
	}
	if best != nil {
		Gia_ManEdgeFromArray(g, best)
	}
}
