package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// Seg_ManCreateCnf adds LUT delay and edge compatibility constraints.
func Seg_ManCreateCnf(p *Seg_Man_t, two, verbose int32) {
	started := time.Now()
	g := p.pGia
	var timing *tim.Tim_Man_t
	if g.PManTime != nil {
		var ok bool
		timing, ok = g.PManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: invalid timing manager")
		}
	}
	add := func(lits ...int32) {
		if !p.pSat.AddClause(lits...) {
			panic("gia: inconsistent SAT edge constraints")
		}
	}
	lit := func(v int32, neg bool) int32 {
		if v < 0 {
			panic("gia: invalid SAT edge variable")
		}
		if neg {
			return 2*v + 1
		}
		return 2 * v
	}
	first := func(id int32) int32 { return vec.Vec_IntEntry(p.vFirsts, int(id)) }
	nvars := func(id int32) int32 { return vec.Vec_IntEntry(p.vNvars, int(id)) }
	start := p.pSat.Statistics().Clauses
	count := int32(0)
	for id := int32(0); id < g.nObjs; id++ {
		obj := Gia_ManObj(g, id)
		dst, nDst := first(id), nvars(id)
		if timing != nil && Gia_ObjIsCi(obj) != 0 {
			ci := Gia_ObjCioId(obj)
			box := tim.Tim_ManBoxForCi(timing, int(ci))
			if nDst > 0 && box >= 0 {
				firstCI := tim.Tim_ManBoxOutputFirst(timing, box)
				if firstCI == int(ci) {
					nInputs := tim.Tim_ManBoxInputNum(timing, box)
					firstInput := tim.Tim_ManBoxInputFirst(timing, box)
					for i := 0; i < nInputs-1; i++ {
						driver := Gia_ObjFaninId0p(g, Gia_ManCo(g, int32(firstInput+i)))
						n := nvars(driver)
						if n >= nDst {
							panic("gia: invalid SAT box levels")
						}
						for k := int32(0); k < n; k++ {
							add(lit(first(driver)+k, true), lit(dst+k+1, false))
						}
					}
				} else {
					ref := Gia_ObjId(g, Gia_ManCi(g, int32(firstCI)))
					for k := int32(0); k < nvars(ref); k++ {
						add(lit(first(ref)+k, true), lit(dst+k, false))
					}
				}
			}
			continue
		}
		if Gia_ObjIsLut(g, id) == 0 {
			continue
		}
		for _, fan := range Gia_ObjLutFanins(g, id) {
			fanObj := Gia_ManObj(g, fan)
			base, n := first(fan), nvars(fan)
			if timing != nil && Gia_ObjIsCi(fanObj) != 0 {
				if n > nDst {
					panic("gia: invalid SAT input levels")
				}
				for k := int32(0); k < n; k++ {
					add(lit(base+k, true), lit(dst+k, false))
				}
			} else if Gia_ObjIsAnd(fanObj) != 0 {
				if n == 1 || n >= nDst {
					panic("gia: invalid SAT edge levels")
				}
				if n == 0 {
					add(lit(count, true), lit(dst, false))
					add(lit(count, false), lit(dst+1, false))
				}
				for k := int32(0); k < n; k++ {
					add(lit(base+k, true), lit(count, true), lit(dst+k, false))
					add(lit(base+k, true), lit(count, false), lit(dst+k+1, false))
				}
				count++
			}
		}
	}
	if count != p.nVars {
		panic("gia: mismatched SAT edge count")
	}
	if verbose != 0 {
		fmt.Printf("Delay constraints = %d. ", p.pSat.Statistics().Clauses-start)
	}
	start = p.pSat.Statistics().Clauses
	incident := Seg_ManCollectObjEdges(p.vEdges, g.nObjs)
	for id := 0; id < vec.Vec_WecSize(incident); id++ {
		edges := vec.Vec_IntArray(vec.Vec_WecEntry(incident, id))
		if two == 0 || len(edges) > 10 {
			for i := 0; i < len(edges); i++ {
				for j := i + 1; j < len(edges); j++ {
					add(lit(edges[i], true), lit(edges[j], true))
				}
			}
		} else {
			for i := 0; i < len(edges); i++ {
				for j := i + 1; j < len(edges); j++ {
					for k := j + 1; k < len(edges); k++ {
						add(lit(edges[i], true), lit(edges[j], true), lit(edges[k], true))
					}
				}
			}
		}
	}
	for _, edge := range vec.Vec_IntArray(p.vToSkip) {
		add(lit(edge, true))
	}
	if verbose != 0 {
		fmt.Printf("Edge constraints = %d. Time = %.2f sec\n", p.pSat.Statistics().Clauses-start, time.Since(started).Seconds())
	}
}
