package gia

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/glycerine/abcgo/sat/bsat"
)

// Swp_Man_t stores the probes, path conditions and incremental SAT instance.
type Swp_Man_t struct {
	pGia                                                                *Gia_Man_t
	nConfMax, nTimeOut                                                  int32
	vProbes, vCondProbes, vCondAssump                                   *Vec_Int_t
	pSat                                                                *bsat.Solver
	vId2Lit, vFront, vFanins, vCexSwp, vCexUser                         *Vec_Int_t
	nSatVars                                                            int32
	nSatCalls, nSatCallsSat, nSatCallsUnsat, nSatCallsUndec, nSatProofs int32
	timeStart                                                           time.Time
	timeTotal, timeCnf, timeSat, timeSatSat, timeSatUnsat, timeSatUndec time.Duration
}

func Swp_ManObj2Lit(p *Swp_Man_t, id int32) int32 { return Vec_IntGetEntry(p.vId2Lit, int(id)) }
func Swp_ManLit2Lit(p *Swp_Man_t, lit int32) int32 {
	mapped := Vec_IntEntry(p.vId2Lit, int(lit>>1))
	if mapped == 0 {
		panic("gia: unmapped SAT literal")
	}
	return mapped ^ (lit & 1)
}
func Swp_ManSetObj2Lit(p *Swp_Man_t, id, lit int32) {
	if lit <= 0 {
		panic("gia: invalid SAT literal")
	}
	Vec_IntSetEntry(p.vId2Lit, int(id), lit)
}
func Swp_ManStart(graph *Gia_Man_t) *Swp_Man_t {
	if Vec_IntSize(&graph.vHTable) == 0 || graph.pData != nil {
		panic("gia: invalid sweeper start")
	}
	p := &Swp_Man_t{pGia: graph, nConfMax: 1000, vProbes: Vec_IntAlloc(100), vCondProbes: Vec_IntAlloc(100), vCondAssump: Vec_IntAlloc(100), vId2Lit: Vec_IntAlloc(10000), vFront: Vec_IntAlloc(100), vFanins: Vec_IntAlloc(100), vCexSwp: Vec_IntAlloc(100), pSat: bsat.NewSolver(), nSatVars: 2, timeStart: time.Now()}
	p.pSat.NewVariable()
	p.pSat.NewVariable()
	Swp_ManSetObj2Lit(p, 0, 2)
	if !p.pSat.AddClause(3) {
		panic("gia: cannot initialize SAT constant")
	}
	graph.pData = p
	return p
}
func Swp_ManStop(graph *Gia_Man_t) { sweeper(graph); graph.pData = nil }
func sweeper(p *Gia_Man_t) *Swp_Man_t {
	s, ok := p.pData.(*Swp_Man_t)
	if !ok || s == nil {
		panic("gia: sweeper is not running")
	}
	return s
}
func Gia_SweeperStart(p *Gia_Man_t) *Gia_Man_t {
	if p == nil {
		p = Gia_ManStart(10000)
	}
	if Vec_IntSize(&p.vHTable) == 0 {
		Gia_ManHashStart(p)
	}
	Swp_ManStart(p)
	p.fSweeper = 1
	return p
}
func Gia_SweeperStop(p *Gia_Man_t)            { p.fSweeper = 0; Swp_ManStop(p); Gia_ManHashStop(p) }
func Gia_SweeperIsRunning(p *Gia_Man_t) int32 { return giaBool(p.pData != nil) }
func Gia_SweeperMemUsage(p *Gia_Man_t) float64 {
	s := sweeper(p)
	total := int(unsafe.Sizeof(*s))
	for _, v := range []*Vec_Int_t{s.vProbes, s.vCondProbes, s.vCondAssump, s.vId2Lit, s.vFront, s.vFanins, s.vCexSwp} {
		total += Vec_IntCap(v)
	}
	return 4 * float64(total)
}
func Gia_SweeperPrintStats(p *Gia_Man_t) {
	s := sweeper(p)
	s.timeTotal = time.Since(s.timeStart)
	fmt.Printf("SAT sweeper: memory %.0f bytes, runtime %s; calls %d, SAT %d, UNSAT %d, undecided %d, proofs %d.\n", Gia_SweeperMemUsage(p), s.timeTotal, s.nSatCalls, s.nSatCallsSat, s.nSatCallsUnsat, s.nSatCallsUndec, s.nSatProofs)
}
func Gia_SweeperSetConflictLimit(p *Gia_Man_t, n int32) { sweeper(p).nConfMax = n }
func Gia_SweeperSetRuntimeLimit(p *Gia_Man_t, n int32)  { sweeper(p).nTimeOut = n }
func Gia_SweeperGetCex(p *Gia_Man_t) *Vec_Int_t {
	s := sweeper(p)
	if s.vCexUser != nil && Vec_IntSize(s.vCexUser) != int(Gia_ManPiNum(p)) {
		panic("gia: invalid sweeper counterexample")
	}
	return s.vCexUser
}
func Gia_SweeperProbeCreate(p *Gia_Man_t, lit int32) int32 {
	if lit < 0 {
		panic("gia: invalid probe literal")
	}
	s := sweeper(p)
	id := int32(Vec_IntSize(s.vProbes))
	Vec_IntPush(s.vProbes, lit)
	return id
}
func Gia_SweeperProbeLit(p *Gia_Man_t, id int32) int32 {
	lit := Vec_IntEntry(sweeper(p).vProbes, int(id))
	if lit < 0 {
		panic("gia: deleted probe")
	}
	return lit
}
func Gia_SweeperProbeDelete(p *Gia_Man_t, id int32) int32 {
	old := Gia_SweeperProbeLit(p, id)
	Vec_IntWriteEntry(sweeper(p).vProbes, int(id), -1)
	return old
}
func Gia_SweeperProbeUpdate(p *Gia_Man_t, id, lit int32) int32 {
	old := Gia_SweeperProbeLit(p, id)
	Vec_IntWriteEntry(sweeper(p).vProbes, int(id), lit)
	return old
}
func Gia_SweeperCollectValidProbeIds(p *Gia_Man_t) *Vec_Int_t {
	out := Vec_IntAlloc(1000)
	for i, lit := range Vec_IntArray(sweeper(p).vProbes) {
		if lit >= 0 {
			Vec_IntPush(out, int32(i))
		}
	}
	return out
}
func Gia_SweeperCondPush(p *Gia_Man_t, id int32)    { Vec_IntPush(sweeper(p).vCondProbes, id) }
func Gia_SweeperCondPop(p *Gia_Man_t) int32         { return Vec_IntPop(sweeper(p).vCondProbes) }
func Gia_SweeperCondVector(p *Gia_Man_t) *Vec_Int_t { return sweeper(p).vCondProbes }
