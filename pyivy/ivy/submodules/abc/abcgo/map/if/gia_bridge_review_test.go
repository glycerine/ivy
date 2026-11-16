package if_

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestGiaMappingDefaultsResetAllState(t *testing.T) {
	p := If_Par_t{NLutSize: 12, NCutsMax: 99, FTruth: 1, FUseDsd: 1, PLutStruct: "44", FPower: 1, WireDelay: 7,
		pTimesArr: []float32{2}, pTimesReq: []float32{3}, pFuncWrite: func(*If_Man_t) int32 { return 0 }, pReoMan: "old", pDsdManager: &If_DsdMan_t{}, nLatchesCi: 8}
	Gia_ManSetIfParsDefault(&p)
	want := If_Par_t{NLutSize: -1, NCutsMax: 8, NFlowIters: 1, NAreaIters: 2, DelayTarget: -1, Epsilon: 0.005, FPreprocess: 1, FExpRed: 1, FEdge: 1, fUseCoAttrs: 1}
	if !reflect.DeepEqual(p, want) {
		t.Fatalf("defaults retain stale state or differ from giaIf.c:\n%+v\nwant:\n%+v", p, want)
	}
}

// Evaluate the actual IF fanin/phase records, without mapper truth machinery.
func bridgeReviewEvaluate(p *If_Man_t, inputs []bool) []bool {
	values := make([]bool, If_ManObjNum(p))
	var outputs []bool
	input := 0
	for id := int32(0); id < If_ManObjNum(p); id++ {
		o := If_ManObj(p, id)
		switch o.Type {
		case IF_CONST1:
			values[id] = true
		case IF_CI:
			values[id] = inputs[input]
			input++
		case IF_AND:
			values[id] = (values[o.pFanin0.Id] != (o.fCompl0 != 0)) && (values[o.pFanin1.Id] != (o.fCompl1 != 0))
		case IF_CO:
			values[id] = values[o.pFanin0.Id] != (o.fCompl0 != 0)
			outputs = append(outputs, values[id])
		default:
			panic("unexpected IF object type")
		}
	}
	return outputs
}

func bridgeReviewReferenceCounts(t *testing.T, p *If_Man_t) {
	t.Helper()
	refs := make([]int32, If_ManObjNum(p))
	visits := make([]int32, len(refs))
	for id := int32(0); id < If_ManObjNum(p); id++ {
		o := If_ManObj(p, id)
		if o.Type == IF_AND || o.Type == IF_CO {
			refs[o.pFanin0.Id]++
		}
		if o.Type == IF_AND {
			refs[o.pFanin1.Id]++
			visits[o.pFanin0.Id]++
			visits[o.pFanin1.Id]++
		}
	}
	for id := int32(0); id < If_ManObjNum(p); id++ {
		o := If_ManObj(p, id)
		if o.nRefs != refs[id] || o.nVisits != visits[id] || o.nVisitsCopy != visits[id] {
			t.Fatalf("node%d refs/visits/copy %d/%d/%d, want %d/%d/%d", id, o.nRefs, o.nVisits, o.nVisitsCopy, refs[id], visits[id], visits[id])
		}
	}
}

func TestGiaToIfTypedPhasesBuffersAndConstants(t *testing.T) {
	for phase := 0; phase < 64; phase++ {
		t.Run(fmt.Sprint(phase), func(t *testing.T) {
			g := gia.Gia_ManStart(32)
			defer gia.Gia_ManStop(g)
			g.EnableMuxes()
			a, b, c, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			bit := func(n uint) int32 { return int32(phase >> n & 1) }
			x := gia.Gia_ManAppendXorReal(g, a^bit(0), b^bit(1))
			mux := gia.Gia_ManAppendMuxReal(g, c^bit(2), x^bit(3), d^bit(4))
			buf := gia.Gia_ManAppendBuf(g, mux^bit(5))
			root := gia.Gia_ManAppendAnd(g, buf, a^bit(0))
			for _, lit := range []int32{root, buf, 0, 1} {
				gia.Gia_ManAppendCo(g, lit)
			}
			var pars If_Par_t
			Gia_ManSetIfParsDefault(&pars)
			pars.NLutSize = 4
			p := Gia_ManToIf(g, &pars)
			defer If_ManStop(p)
			if If_ManCiNum(p) != 4 || If_ManCoNum(p) != 4 || p.nObjs[IF_AND] != 7 || p.nLevelMax != 5 {
				t.Fatalf("typed bridge geometry CI%d CO%d AND%d level%d", If_ManCiNum(p), If_ManCoNum(p), p.nObjs[IF_AND], p.nLevelMax)
			}
			if gia.Gia_ObjValue(gia.Gia_ManObj(g, buf>>1)) != gia.Gia_ObjValue(gia.Gia_ManObj(g, mux>>1)) {
				t.Fatal("buffer did not alias its driver's IF object")
			}
			bridgeReviewReferenceCounts(t, p)
			for assignment := 0; assignment < 16; assignment++ {
				inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
				av, bv, cv, dv := inputs[0] != (bit(0) != 0), inputs[1] != (bit(1) != 0), inputs[2] != (bit(2) != 0), inputs[3] != (bit(4) != 0)
				selected := dv
				if cv {
					selected = (av != bv) != (bit(3) != 0)
				}
				buffer := selected != (bit(5) != 0)
				want := []bool{buffer && av, buffer, false, true}
				if got := bridgeReviewEvaluate(p, inputs); !slices.Equal(got, want) {
					t.Fatalf("assignment%d IF%v want%v", assignment, got, want)
				}
			}
		})
	}
}

func TestGiaToIfConstantAndDegenerateFaninCopies(t *testing.T) {
	g := gia.Gia_ManStart(24)
	defer gia.Gia_ManStop(g)
	gia.Gia_ManSetSimple(g, 1)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	for _, pair := range [][2]int32{{a, a}, {a, a ^ 1}, {0, a}, {1, b}, {0, 1}, {1, 1}} {
		root := gia.Gia_ManAppendAnd(g, pair[0], pair[1])
		gia.Gia_ManAppendCo(g, root)
		gia.Gia_ManAppendCo(g, root^1)
	}
	var pars If_Par_t
	Gia_ManSetIfParsDefault(&pars)
	pars.NLutSize = 4
	p := Gia_ManToIf(g, &pars)
	defer If_ManStop(p)
	bridgeReviewReferenceCounts(t, p)
	for assignment := 0; assignment < 4; assignment++ {
		a, b := assignment&1 != 0, assignment&2 != 0
		var want []bool
		for _, v := range []bool{a, false, false, b, false, true} {
			want = append(want, v, !v)
		}
		if got := bridgeReviewEvaluate(p, []bool{a, b}); !slices.Equal(got, want) {
			t.Fatalf("assignment%d IF%v want%v", assignment, got, want)
		}
	}
}

func TestGiaToIfEmptyInterfaces(t *testing.T) {
	for _, inputs := range []int{0, 3} {
		g := gia.Gia_ManStart(8)
		for i := 0; i < inputs; i++ {
			gia.Gia_ManAppendCi(g)
		}
		var pars If_Par_t
		Gia_ManSetIfParsDefault(&pars)
		pars.NLutSize = 4
		p := Gia_ManToIf(g, &pars)
		if If_ManObjNum(p) != int32(inputs+1) || If_ManCiNum(p) != int32(inputs) || If_ManCoNum(p) != 0 {
			t.Fatal("empty-interface geometry changed")
		}
		If_ManStop(p)
		gia.Gia_ManStop(g)
	}
}
