package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestVerifySpecialCasesPrecedeSAT(t *testing.T) {
	for _, naive := range []int32{0, 1} {
		t.Run(map[int32]string{0: "default", 1: "naive"}[naive], func(t *testing.T) {
			g := gia.Gia_ManStart(8)
			a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			gia.Gia_ManAppendCo(g, a)
			gia.Gia_ManAppendCo(g, b)
			// ABC's PI-vs-PI special case sets the first input in its witness.
			// A negative conflict limit is invalid for the SAT core, so success
			// demonstrates that the special-case path did not invoke it.
			pars := Cec_ParCec_t{nBTLimit: -1, fNaive: naive, fSilent: 1, iOutFail: -1}
			if got := Cec_ManVerify(g, &pars); got != 0 || pars.Error != nil || pars.iOutFail != 0 {
				t.Fatalf("status %d, error %v, output %d", got, pars.Error, pars.iOutFail)
			}
			cex := g.CombinationalCounterexample()
			if cex == nil || cex.PData[0]&3 != 1 {
				t.Fatalf("wrong trivial witness: %+v", cex)
			}
		})
	}
}

func TestVerifyInvalidPairsStopBeforeSAT(t *testing.T) {
	g := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCo(g, 0)
	var pars Cec_ParCec_t
	if got := Cec_ManVerify(g, &pars); got != -1 || pars.Error == nil {
		t.Fatalf("status %d, error %v", got, pars.Error)
	}
}

func TestVerifyNilGraph(t *testing.T) {
	var pars Cec_ParCec_t
	if status := Cec_ManVerify(nil, &pars); status != -1 || pars.Error == nil {
		t.Fatalf("status %d error %v", status, pars.Error)
	}
}

func TestVerifyTrivialPairsWitnesses(t *testing.T) {
	for left := 0; left < 6; left++ {
		for right := 0; right < 6; right++ {
			g := gia.Gia_ManStart(8)
			a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			literals := []int32{0, 1, a, a ^ 1, b, b ^ 1}
			gia.Gia_ManAppendCo(g, literals[left])
			gia.Gia_ManAppendCo(g, literals[right])
			var pars Cec_ParCec_t
			status := Cec_ManVerify(g, &pars)
			if pars.Error != nil {
				t.Fatalf("%d %d: %v", left, right, pars.Error)
			}
			if left == right {
				if status != 1 {
					t.Fatalf("%d %d: status %d", left, right, status)
				}
				continue
			}
			if status != 0 {
				t.Fatalf("%d %d: status %d", left, right, status)
			}
			cex := g.CombinationalCounterexample()
			if cex == nil || cex.IPo != 0 {
				t.Fatalf("%d %d: witness %+v", left, right, cex)
			}
			in := []bool{cex.PData[0]&1 != 0, cex.PData[0]&2 != 0}
			out, err := g.Evaluate(in)
			if err != nil || out[0] == out[1] {
				t.Fatalf("%d %d: invalid witness %v outputs %v err %v", left, right, in, out, err)
			}
		}
	}
}

func TestVerifyPipelineForcesGenuineFraFallback(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, x)
	var pars Cec_ParCec_t
	Cec_ManCecSetDefaultParams(&pars)
	calls := 0
	forcedUnresolved := func(p *Gia_Man_t, _ *Cec_ParFra_t) (*Gia_Man_t, error) { return gia.Gia_ManDup(p), nil }
	actualFra := func(p *Gia_Man_t, verbose int32) (int32, int32, error) { calls++; return cecVerifyFra(p, verbose) }
	status := cecVerifyPipeline(g, &pars, forcedUnresolved, actualFra)
	if status != 0 || calls != 1 || pars.Error != nil || pars.iOutFail != 0 {
		t.Fatalf("status %d calls %d error %v output %d", status, calls, pars.Error, pars.iOutFail)
	}
	w := g.CombinationalCounterexample()
	if w == nil {
		t.Fatal("missing FRA witness")
	}
	out, err := g.Evaluate([]bool{w.PData[0]&1 != 0, w.PData[0]&2 != 0})
	if err != nil || out[0] == out[1] {
		t.Fatalf("witness output %v error %v", out, err)
	}
}

func TestVerifyPipelineNaiveSkipsSweep(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, x)
	var pars Cec_ParCec_t
	Cec_ManCecSetDefaultParams(&pars)
	pars.fNaive = 1
	failSweep := func(*Gia_Man_t, *Cec_ParFra_t) (*Gia_Man_t, error) {
		t.Fatal("naive path invoked sweep")
		return nil, nil
	}
	failFallback := func(*Gia_Man_t, int32) (int32, int32, error) {
		t.Fatal("naive path invoked fallback")
		return -1, -1, nil
	}
	if status := cecVerifyPipeline(g, &pars, failSweep, failFallback); status != 0 || pars.Error != nil || g.CombinationalCounterexample() != nil {
		t.Fatalf("status %d error %v", status, pars.Error)
	}
}

func TestVerifyPipelineRejectsUnsupportedNativeNodes(t *testing.T) {
	for _, kind := range []int{3, 4, 5} {
		g := gia.Gia_ManStart(16)
		g.EnableMuxes()
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		var node int32
		switch kind {
		case 3:
			node = gia.Gia_ManAppendXorReal(g, a, b)
		case 4:
			node = gia.Gia_ManAppendMuxReal(g, c, a, b)
		case 5:
			node = gia.Gia_ManAppendBuf(g, a)
		}
		gia.Gia_ManAppendCo(g, node)
		gia.Gia_ManAppendCo(g, 0)
		var p Cec_ParCec_t
		Cec_ManCecSetDefaultParams(&p)
		status := cecVerifyPipeline(g, &p, nil, nil)
		if status != -1 || p.Error == nil {
			t.Fatalf("kind %d status %d error %v", kind, status, p.Error)
		}
	}
}
