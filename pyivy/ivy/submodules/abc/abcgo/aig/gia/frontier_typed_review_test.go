package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestFrontierTypedRoundTrip(t *testing.T) {
	for seed := uint32(0); seed < 64; seed++ {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			g, _ := falseReviewTypedGraph(seed)
			defer Gia_ManStop(g)
			if got, want := Gia_ManCrossCutSimple(g), frontierReviewLastUseCut(g); got != want {
				t.Errorf("cross-cut%d want%d", got, want)
			}
			for id, obj := range g.pObjs {
				if obj.Value != 0 {
					t.Errorf("object%d retains reference count%d", id, obj.Value)
				}
			}
			out := Gia_ManFront(g)
			defer Gia_ManStop(out)
			Gia_ManFrontTransform(out)
			cofReviewCheckRebuild(t, g, out)
			if out.nXors != g.nXors || out.nMuxes != g.nMuxes || out.nBufs != g.nBufs || !slices.Equal(out.PMuxes, g.PMuxes) {
				t.Fatal("typed manager metadata did not survive roundtrip")
			}
			for id, obj := range g.pObjs {
				got := out.pObjs[id]
				if obj.iDiff0 != got.iDiff0 || obj.iDiff1 != got.iDiff1 || obj.fCompl0 != got.fCompl0 || obj.fCompl1 != got.fCompl1 {
					t.Errorf("node%d changed physical fanins", id)
				}
			}
			for id, control := range g.PMuxes {
				if control != 0 {
					out.PMuxes[id] ^= 1
					if g.PMuxes[id] != control {
						t.Fatal("frontier MUX metadata aliases source")
					}
					break
				}
			}
		})
	}
}

func TestFrontierTypedWordConsumers(t *testing.T) {
	for seed := uint32(0); seed < 16; seed++ {
		g, _ := falseReviewTypedGraph(seed)
		for _, kind := range []string{"words", "switching", "ternary"} {
			t.Run(fmt.Sprintf("%d/%s", seed, kind), func(t *testing.T) {
				frontierReviewConsumerTruth(t, g, kind)
			})
		}
		Gia_ManStop(g)
	}
}

func frontierReviewConsumerTruth(t *testing.T, g *Gia_Man_t, kind string) {
	t.Helper()
	// Pack all sixteen Boolean CI assignments into a word. Each frame then
	// checks every output against independent scalar gate evaluation.
	inputs := []uint32{0xaaaa, 0xcccc, 0xf0f0, 0xff00}
	switch kind {
	case "words":
		pars := Gia_ParSim_t{NWords: 1}
		p := Gia_ManSimCreate(g, &pars)
		defer Gia_ManSimDelete(p)
		for i, bits := range inputs {
			Gia_SimDataCi(p, int32(i))[0] = bits
		}
		Gia_ManSimulateRound(p)
		for row := uint32(0); row < 16; row++ {
			values := cofReviewValues(g, int(row))
			for co := int32(0); co < Gia_ManCoNum(g); co++ {
				obj := Gia_ManCo(g, co)
				want := values[Gia_ObjFaninId0(obj, obj.id)] != (obj.fCompl0 != 0)
				if got := Gia_SimDataCo(p, co)[0]>>row&1 != 0; got != want {
					t.Fatalf("row%d output%d got%t want%t", row, co, got, want)
				}
			}
		}
	case "switching":
		pars := Gia_ParSwi_t{nWords: 1}
		p := Gia_ManSwiCreate(g, &pars)
		defer Gia_ManSwiDelete(p)
		for i, bits := range inputs {
			Gia_SwiDataCi(p, int32(i))[0] = bits
		}
		Gia_ManSwiSimulateRound(p, 1)
		for row := uint32(0); row < 16; row++ {
			values := cofReviewValues(g, int(row))
			for co := int32(0); co < Gia_ManCoNum(g); co++ {
				obj := Gia_ManCo(g, co)
				want := values[Gia_ObjFaninId0(obj, obj.id)] != (obj.fCompl0 != 0)
				if got := Gia_SwiDataCo(p, co)[0]>>row&1 != 0; got != want {
					t.Fatalf("row%d output%d got%t want%t", row, co, got, want)
				}
			}
		}
	case "ternary":
		p := Gia_ManTerCreate(g)
		defer Gia_ManTerDelete(p)
		for row := uint32(0); row < 16; row++ {
			for ci := int32(0); ci < Gia_ManCiNum(g); ci++ {
				Gia_ManTerSimInfoSet(p.pDataSimCis, ci, GIA_ZER+int32(row>>ci&1))
			}
			Gia_ManTerSimulateRound(p)
			values := cofReviewValues(g, int(row))
			for co := int32(0); co < Gia_ManCoNum(g); co++ {
				obj := Gia_ManCo(g, co)
				want := values[Gia_ObjFaninId0(obj, obj.id)] != (obj.fCompl0 != 0)
				if got := Gia_ManTerSimInfoGet(p.pDataSimCos, co) == GIA_ONE; got != want {
					t.Fatalf("row%d output%d got%t want%t", row, co, got, want)
				}
			}
		}
	}
}

func TestFrontierReorderedConsumers(t *testing.T) {
	g := Gia_ManStart(32)
	defer Gia_ManStop(g)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b^1)
	y := Gia_ManAppendAnd(g, c, d)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, y)
	Gia_ManAppendCo(g, x^1)
	Gia_ManSetRegNum(g, 1)
	for _, ids := range []*Vec_Int_t{g.vCis, g.vCos} {
		a, b := Vec_IntEntry(ids, 0), Vec_IntEntry(ids, 1)
		Vec_IntWriteEntry(ids, 0, b)
		Vec_IntWriteEntry(ids, 1, a)
		Gia_ObjSetCioId(g.pObjs[b], 0)
		Gia_ObjSetCioId(g.pObjs[a], 1)
	}
	for _, kind := range []string{"words", "switching", "ternary"} {
		t.Run(kind, func(t *testing.T) { frontierReviewConsumerTruth(t, g, kind) })
	}
}

func TestFrontierConstantControlsAndSimpleFanins(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for phases := int32(0); phases < 4; phases++ {
			t.Run(fmt.Sprintf("typed%t/phases%d", typed, phases), func(t *testing.T) {
				g := Gia_ManStart(32)
				defer Gia_ManStop(g)
				g.fGiaSimple = 1
				a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
				c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
				x := Gia_ManAppendAnd(g, a^(phases&1), a^(phases>>1))
				g.EnableMuxes()
				control := Gia_ManAppendMuxReal(g, 1, b, c)
				var y int32
				if typed {
					y = Gia_ManAppendXorReal(g, c, d)
				} else {
					y = Gia_ManAppendAnd(g, c, d)
				}
				Gia_ManAppendCo(g, x)
				Gia_ManAppendCo(g, control)
				Gia_ManAppendCo(g, y)
				out := Gia_ManFront(g)
				defer Gia_ManStop(out)
				Gia_ManFrontTransform(out)
				cofReviewCheckRebuild(t, g, out)
				if out.fGiaSimple != 1 {
					t.Fatal("lost simple-mode representation flag")
				}
				if got, want := Gia_ManCrossCutSimple(g), frontierReviewLastUseCut(g); got != want {
					t.Fatalf("cut%d want%d", got, want)
				}
				for _, obj := range g.pObjs {
					if obj.Value != 0 {
						t.Fatal("retained reference count")
					}
				}
				for _, kind := range []string{"words", "switching", "ternary"} {
					frontierReviewConsumerTruth(t, g, kind)
				}
			})
		}
	}
}

func TestFrontierEmptyAndDirectOutputs(t *testing.T) {
	for _, outputs := range []bool{false, true} {
		g := Gia_ManStart(8)
		if outputs {
			a := Gia_ManAppendCi(g)
			Gia_ManAppendCo(g, 0)
			Gia_ManAppendCo(g, 1)
			Gia_ManAppendCo(g, a)
			Gia_ManAppendCo(g, a^1)
		}
		out := Gia_ManFront(g)
		Gia_ManFrontTransform(out)
		cofReviewCheckRebuild(t, g, out)
		if text := hashReviewStdout(t, func() { Gia_ManFrontTest(g) }); text != "Verification successful.\n" {
			t.Fatalf("unexpected roundtrip report%q", text)
		}
		Gia_ManStop(out)
		Gia_ManStop(g)
	}
}
