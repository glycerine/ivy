//go:build linux

package pdr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
	"time"
)

func TestPdrCPUDeadlineExcludesSleep(t *testing.T) {
	defer pdrPinClock()()
	p := &Pdr_Man_t{pPars: &Pdr_Par_t{}, runtimeClock: pdrCPUClock}
	p.timeToStop = pdrCPUClock() + int64(30*time.Millisecond)
	time.Sleep(70 * time.Millisecond)
	solver := bsat.NewSolver()
	if got := pdrSolve(p, solver, nil, 0); got != bsat.Satisfiable {
		t.Fatalf("sleep exhausted CPU deadline: %v", got)
	}
	for pdrCPUClock() <= p.timeToStop {
	}
	if got := pdrSolve(p, solver, nil, 0); got != bsat.Unknown {
		t.Fatalf("CPU deadline not enforced: %v", got)
	}
}

func TestPdrTimedSolveKeepsDeadlineDomainForDirectQuery(t *testing.T) {
	defer pdrPinClock()()
	g := aig.NewGraph()
	x := g.NewInput()
	q := g.NewInput()
	g.AddOutput(x)
	g.AddOutput(q)
	aig.Aig_ManSetRegNum(g, 1)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.nTimeOut, pars.fSilent = 1, 1
	p := Pdr_ManStart(g, &pars, nil)
	defer Pdr_ManStop(p)
	if got := Pdr_ManSolveInt(p); got != 0 || p.runtimeClock == nil || p.timeToStop == 0 {
		t.Fatalf("timed solve=%d clock=%v deadline=%d", got, p.runtimeClock != nil, p.timeToStop)
	}
	solver := bsat.NewSolver()
	if got := pdrSolve(p, solver, nil, 0); got != bsat.Satisfiable {
		t.Fatalf("live CPU deadline after solve returned %v", got)
	}
	p.timeToStop = p.runtimeClock() - 1
	if got := pdrSolve(p, solver, nil, 0); got != bsat.Unknown {
		t.Fatalf("expired CPU deadline after solve returned %v", got)
	}
}
