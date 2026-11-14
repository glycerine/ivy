package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

// Sle_ManExplore searches for a low-delay LUT mapping and compatible edges.
func Sle_ManExplore(g *Gia_Man_t, limit, initial, dynamic, twoEdges, verbose int32) {
	started := time.Now()
	start := initial
	if initial == 0 && Gia_ManHasMapping(g) != 0 {
		start = Gia_ManLutLevel(g, nil)
	}
	p := Sle_ManAlloc(g, start, verbose)
	defer Sle_ManStop(p)
	Sle_ManMarkupVariables(p)
	Sle_ManDeriveInit(p)
	useDynamic := dynamic
	if twoEdges != 0 {
		useDynamic = 1
	}
	Sle_ManDeriveCnf(p, limit, useDynamic)
	if verbose != 0 {
		fmt.Printf("Running SAT LUT mapping: vars %d, initial delay %d, edge limit %d.\n", p.nVarsTotal, start, 1+twoEdges)
	}
	edges := vec.Vec_IntAlloc(1000)
	mapping := vec.Vec_IntAlloc(1000)
	assumptions := make([]bsat.Literal, 0)
	for delay := p.nLevels; delay >= 0; delay-- {
		if delay < p.nLevels {
			for i := int32(0); i < Gia_ManCoNum(g); i++ {
				id := Gia_ObjFaninId0p(g, Gia_ManCo(g, i))
				if vec.Vec_BitEntry(p.vMask, int(id)) != 0 {
					assumptions = append(assumptions, bsat.Negative(vec.Vec_IntEntry(p.vDelayFirst, int(id))+delay))
				}
			}
		}
		before := p.pSat.Statistics().Conflicts
		var status bsat.Status
		for {
			p.nSatCalls++
			status = p.pSat.Solve(assumptions...)
			if status != bsat.Satisfiable || Sle_ManAddEdgeConstraints(p, 1+twoEdges) == 0 {
				break
			}
		}
		conflicts := p.pSat.Statistics().Conflicts - before
		if status != bsat.Satisfiable {
			if verbose != 0 {
				fmt.Printf("No solution for delay %d (status %d, conflicts %d). Time = %.2f sec\n", delay, status, conflicts, time.Since(started).Seconds())
			}
			break
		}
		Sle_ManDeriveResult(p, edges, mapping)
		if verbose != 0 {
			fmt.Printf("Solution with delay %d and %d edges. Conflicts = %d. Time = %.2f sec\n", delay, vec.Vec_IntSize(edges)/2, conflicts, time.Since(started).Seconds())
		}
	}
	if vec.Vec_IntSize(mapping) > 0 {
		Gia_ManEdgeFromArray(g, edges)
		g.VMapping = mapping
	}
	g.VPacking = nil
}
