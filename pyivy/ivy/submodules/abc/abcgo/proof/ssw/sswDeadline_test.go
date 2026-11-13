package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
	"time"
)

func sswDeadlineGraph() *aig.Graph {
	g := aig.NewGraph()
	pi, q := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(pi, q))
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	return g
}
func TestCorrespondenceExpiredDeadlineReturnsUnknown(t *testing.T) {
	for _, latch := range []bool{false, true} {
		g := sswDeadlineGraph()
		p := new(Ssw_Pars_t)
		if latch {
			Ssw_ManSetDefaultParamsLcorr(p)
		} else {
			Ssw_ManSetDefaultParams(p)
		}
		deadline := time.Now().Add(-time.Second)
		Ssw_ParsSetDeadline(p, deadline)
		if result := Ssw_SignalCorrespondence(g, p); result != nil {
			t.Fatalf("latch=%t: expired call returned a graph", latch)
		}
		if !p.deadline.Equal(deadline) {
			t.Fatal("caller deadline changed")
		}
	}
}
func TestCorrespondenceStopsWhenDeadlineExpiresDuringRun(t *testing.T) {
	for _, latch := range []bool{false, true} {
		g := sswDeadlineGraph()
		p := new(Ssw_Pars_t)
		if latch {
			Ssw_ManSetDefaultParamsLcorr(p)
		} else {
			Ssw_ManSetDefaultParams(p)
		}
		p.TimeLimit = 10
		p.pFunc = func(any) { p.deadline = time.Now().Add(-time.Millisecond) }
		if result := Ssw_SignalCorrespondence(g, p); result != nil {
			t.Fatalf("latch=%t: timed-out call returned a graph", latch)
		}
		// Cleanup releases the fanout index even when the deadline interrupts work.
		aig.Aig_ManFanoutStart(g)
		aig.Aig_ManFanoutStop(g)
		if !p.deadline.IsZero() {
			t.Fatal("temporary TimeLimit deadline was not restored")
		}
	}
}
func TestExpiredDeadlineMakesSatQueryUnknown(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := &Ssw_Man_t{pPars: &Ssw_Pars_t{deadline: time.Now().Add(-time.Second)}, pFrames: g, pMSat: Ssw_SatStart(0), iOutputLit: -1}
	p.pMSat.pAig = g
	if Ssw_NodesAreEquiv(p, a, b) != -1 {
		t.Fatal("expired query did not return unknown")
	}
}
func TestCecDelegationWithDeadlineReturnsUnknown(t *testing.T) {
	g := sswDeadlineGraph()
	p := new(Ssw_Pars_t)
	Ssw_ManSetDefaultParams(p)
	p.fScorrGia = 1
	Ssw_ParsSetDeadline(p, time.Now().Add(time.Hour))
	if result := Ssw_SignalCorrespondence(g, p); result != nil {
		t.Fatal("deadline-unsupported CEC delegation returned a graph")
	}
}

func TestStandaloneTimeLimitExpiresDuringCallback(t *testing.T) {
	g := sswDeadlineGraph()
	p := new(Ssw_Pars_t)
	Ssw_ManSetDefaultParams(p)
	p.TimeLimit = 1
	calls := 0
	p.pFunc = func(any) { calls++; time.Sleep(1100 * time.Millisecond) }
	if result := Ssw_SignalCorrespondence(g, p); result != nil {
		t.Fatal("expired TimeLimit returned a graph")
	}
	if calls != 1 {
		t.Fatalf("callback invoked %d times after expiration", calls)
	}
}

func TestExpiredDeadlineStopsSimulationAndClassPreparation(t *testing.T) {
	g := sswDeadlineGraph()
	p := &Ssw_Pars_t{deadline: time.Now().Add(-time.Millisecond)}
	sml := Ssw_SmlStart(g, 0, 4, 1)
	Ssw_SmlInitialize(sml, 1)
	if sswSimulateOne(sml, p) || sml.nSimRounds != 0 {
		t.Fatal("expired simulation completed")
	}
	if classes := sswClassesPrepare(g, 1, 0, 0, 0, 0, 0, p); classes != nil {
		t.Fatal("expired class preparation returned classes")
	}
}

func TestExpiredFrameConstructionReturnsUnknown(t *testing.T) {
	g := sswDeadlineGraph()
	p := Ssw_ManCreate(g, &Ssw_Pars_t{nFramesK: 1, deadline: time.Now().Add(-time.Second)})
	defer Ssw_ManStop(p)
	if result := Ssw_FramesWithClasses(p); result != nil {
		t.Fatal("expired frame builder returned a graph")
	}
	for _, entry := range p.pNodeToFrames {
		if entry != aig.InvalidEdge {
			t.Fatal("expired frame builder left a mapping")
		}
	}
}

func TestExpiredConstraintSearchClearsOldAssignment(t *testing.T) {
	old := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(old, 1)
	if status := sswManSetConstrPhases(sswDeadlineGraph(), 1, &old, time.Now().Add(-time.Second)); status != -1 || old != nil {
		t.Fatalf("status %d assignment %v", status, old)
	}
}
