package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"os"
	"time"
	"unsafe"
)

func Ga2_ManComputeTest(p *gia.Gia_Man_t) {
	start := time.Now()
	Ga2_ManMarkup(p, 5, 0)
	fmt.Printf("Time: %v\n", time.Since(start))
	var count int
	for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) != 0 && gia.Gia_ObjPhase(o) != 0 {
			count++
		}
	}
	fmt.Printf("Marked AND nodes = %6d.  Time: %v\n", count, time.Since(start))
}
func Ga2_ManDumpStats(g *gia.Gia_Man_t, pars *Abs_Par_t, solver *bsat.LabeledSolver2, frame, useN int32) {
	name := "stats_gla"
	if useN != 0 {
		name += "n"
	}
	if pars.fUseFullProof != 0 {
		name += "p"
	}
	name += ".txt"
	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
	if err != nil {
		fmt.Printf("Cannot open %s: %v\n", name, err)
		return
	}
	defer f.Close()
	_, proof := solver.Memory(true)
	fmt.Fprintf(f, "%s pi=%d ff=%d and=%d mem=%d bmc=%d", gia.Gia_ManName(g), gia.Gia_ManPiNum(g), gia.Gia_ManRegNum(g), gia.Gia_ManAndNum(g), int(1+proof/(1<<20)), frame)
	if classes := gia.Gia_ManGateClasses(g); classes != nil {
		fmt.Fprintf(f, " ff=%d and=%d", Gia_GlaCountFlops(g, classes), Gia_GlaCountNodes(g, classes))
	}
	fmt.Fprintln(f)
}
func Ga2_ManReportMemory(p *Ga2_Man_t) {
	aig := gia.Gia_ManMemory(p.pGia) + vec.Vec_IntMemory(gia.Gia_ManMapping(p.pGia))
	sat, proof := p.pSat.Memory(true)
	var mappings, other float64
	for _, v := range p.vId2Lit {
		if v != nil {
			mappings += vec.Vec_IntMemory(v)
		}
	}
	for _, v := range p.vCnfs {
		if v != nil {
			other += vec.Vec_IntMemory(v)
		}
	}
	for _, v := range []*vec.Vec_Int_t{p.vIds, p.vProofIds, p.vAbs, p.vValues, p.vLits, p.vIsopMem} {
		if v != nil {
			other += vec.Vec_IntMemory(v)
		}
	}
	other += float64(unsafe.Sizeof(*p)) + 336450 + float64(1+unsafe.Sizeof(uintptr(0)))*65536
	refine := Rnm_ManMemoryUsage(p.pRnm)
	hash := float64(len(p.pTable) * 4)
	total := aig + sat + proof + mappings + refine + hash + other
	for _, row := range []struct {
		name  string
		value float64
	}{{"AIG", aig}, {"SAT", sat}, {"Proof", proof}, {"Map", mappings}, {"Refine", refine}, {"Hash", hash}, {"Other", other}, {"TOTAL", total}} {
		percent := float64(0)
		if total != 0 {
			percent = 100 * row.value / total
		}
		fmt.Printf("Memory: %-8s %10.2f MB (%5.1f%%)\n", row.name, row.value/(1<<20), percent)
	}
}
