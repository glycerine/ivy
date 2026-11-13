package cec

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
	"time"
)

func TestCorrespondenceDeadlineStopsAfterCallbackWithoutMutatingSource(t *testing.T) {
	for _, latch := range []int32{0, 1} {
		g := aig.NewGraph()
		pi, a, b := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(a, b))
		g.AddOutput(pi)
		g.AddOutput(pi)
		aig.Aig_ManSetRegNum(g, 2)
		var pars Cec_ParCor_t
		Cec_ManCorSetDefaultParams(&pars)
		pars.fLatchCorr = latch
		pars.deadline = time.Now().Add(time.Hour)
		calls := 0
		pars.pFunc = func(any) { calls++; pars.deadline = time.Now().Add(-time.Second) }
		if result := cecCorrespondence(g, &pars); result != nil {
			t.Fatalf("latch=%d returned partial result", latch)
		}
		if calls == 0 {
			t.Fatalf("latch=%d did not reach callback", latch)
		}
		if aig.Aig_ManHasReprs(g) {
			t.Fatalf("latch=%d committed partial representative classes", latch)
		}
	}
}
func TestCorrespondenceDeadlinePropagatesToSAT(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	out := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, out)
	var pars Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&pars)
	pars.deadline = time.Now().Add(-time.Second)
	p := Cec_ManSatCreate(g, &pars)
	defer Cec_ManSatStop(p)
	if got := p.query(out, nil); got != -1 {
		t.Fatalf("expired SAT query returned %d", got)
	}
}

func TestDirectCorrespondenceDeadlineDiscardsSpeculativeClasses(t *testing.T) {
	for _, prefix := range []int32{0, 1} {
		for _, driver := range []string{"direct", "divide", "bridge"} {
			p := gia.Gia_ManStart(16)
			pi, a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
			gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
			gia.Gia_ManAppendCo(p, pi)
			gia.Gia_ManAppendCo(p, pi)
			gia.Gia_ManSetRegNum(p, 2)
			var pars Cec_ParCor_t
			Cec_ManCorSetDefaultParams(&pars)
			pars.nPrefix = prefix
			pars.deadline = time.Now().Add(time.Hour)
			calls := 0
			pars.pFunc = func(any) { calls++; pars.deadline = time.Now().Add(-time.Second) }
			switch driver {
			case "direct":
				if Cec_ManLSCorrespondence(p, &pars) != nil {
					t.Fatal("partial direct result")
				}
			case "divide":
				if Gia_ManScorrDivideTest(p, &pars) != nil {
					t.Fatal("partial divided result")
				}
			case "bridge":
				pars.SweepCorrespondence(p)
			}
			if calls != 1 {
				t.Fatalf("prefix=%d driver=%s callbacks %d", prefix, driver, calls)
			}
			if p.HasRepresentatives() {
				t.Fatalf("prefix=%d driver=%s left speculative representatives", prefix, driver)
			}
		}
	}
}

func TestExpiredMiterReportsEveryOutputUnknown(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	var pars Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&pars)
	pars.deadline = time.Now().Add(-time.Second)
	var status *Vec_Str_t
	Cec_ManSatSolveMiter(g, &pars, &status)
	if len(vec.Vec_StrArray(status)) != 2 || int8(vec.Vec_StrArray(status)[0]) != -1 || int8(vec.Vec_StrArray(status)[1]) != -1 {
		t.Fatalf("statuses %v", vec.Vec_StrArray(status))
	}
}

func TestDirectCorrespondenceBmcDropsCandidatesOnDeadline(t *testing.T) {
	p := gia.Gia_ManStart(8)
	x, q := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, q)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManSetRegNum(p, 1)
	p.StartRepresentatives()
	var pars Cec_ParCor_t
	Cec_ManCorSetDefaultParams(&pars)
	pars.deadline = time.Now().Add(-time.Second)
	Cec_ManLSCorrespondenceBmc(p, &pars, 0)
	if p.HasRepresentatives() {
		t.Fatal("expired BMC retained speculative classes")
	}
}
