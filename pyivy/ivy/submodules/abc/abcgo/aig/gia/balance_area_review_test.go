package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
	"testing"
)

func TestAreaCollectorPreservesUnrelatedMark0(t *testing.T) {
	g := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b)
	co := Gia_ManAppendCo(g, root)
	for _, lit := range []int32{a, root, co} {
		Gia_ManObj(g, lit>>1).fMark0 = 1
	}
	manager := Dam_ManAlloc(g)
	Dam_ManCollectSets(manager)
	for _, lit := range []int32{a, root, co} {
		if Gia_ManObj(g, lit>>1).fMark0 != 1 {
			t.Fatalf("area collection erased caller's mark0 on node%d", lit>>1)
		}
	}
}

func TestAreaCollectorClearsOnlyItsVisitedMuxMarks(t *testing.T) {
	g := Gia_ManStart(1)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	x := Gia_ManAppendAnd(g, a, b)
	mux := Gia_ManAppendMuxReal(g, c, x, d)
	unrelated := Gia_ManAppendMuxReal(g, b, a, d)
	co := Gia_ManAppendCo(g, mux)
	for _, lit := range []int32{a, x, unrelated, co} {
		Gia_ManObj(g, lit>>1).fMark0 = 1
	}
	manager := Dam_ManAlloc(g)
	for repeat := 0; repeat < 2; repeat++ {
		Dam_ManCollectSets(manager)
		if len(manager.vVisit) != 1 || manager.vVisit[0] != mux>>1 {
			t.Fatalf("visited MUX list %v", manager.vVisit)
		}
		if Gia_ManObj(g, mux>>1).fMark0 != 0 {
			t.Fatal("temporary MUX mark leaked")
		}
		for _, lit := range []int32{a, x, unrelated, co} {
			if Gia_ManObj(g, lit>>1).fMark0 != 1 {
				t.Fatalf("caller mark erased on node%d", lit>>1)
			}
		}
	}
}

func BenchmarkAreaCollectionAndExtraction(b *testing.B) {
	b.ReportAllocs()
	for run := 0; run < b.N; run++ {
		g := areaReviewGraph(5, 0)
		p := Dam_ManAlloc(g)
		p.nLevelMax = Gia_ManSetLevels(g, nil)
		p.vNodLevR = append([]int32(nil), Vec_IntArray(Gia_ManReverseLevel(g))...)
		Dam_ManCreatePairs(p, 0)
		if vec.Vec_QueTopPriority(p.vQue) >= 2 {
			Dam_ManUpdate(p, int32(vec.Vec_QuePop(p.vQue)))
		}
	}
}

func areaReviewInstallMapping(g *Gia_Man_t) {
	g.VMapping = Vec_IntStart(int(g.nObjs))
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		Vec_IntWriteEntry(g.VMapping, int(obj.id), int32(Vec_IntSize(g.VMapping)))
		for _, value := range []int32{2, Gia_ObjFaninId0(obj, obj.id), Gia_ObjFaninId1(obj, obj.id), obj.id} {
			Vec_IntPush(g.VMapping, value)
		}
	}
}

func TestMappedAreaBalanceRetainsOutputRoles(t *testing.T) {
	for simple := int32(0); simple < 2; simple++ {
		src := areaReviewGraph(0, 0)
		areaReviewInstallMapping(src)
		if err := src.SetAigerProperties(1, 1); err != nil {
			t.Fatal(err)
		}
		out := Gia_ManAreaBalance(src, simple, 4, 0, 0)
		if out.nRegs != src.nRegs || out.nConstrs != 1 || out.aigerBadOutputs != 1 {
			t.Fatalf("mapped area result regs%d constraints%d bad%d, want2/1/1", out.nRegs, out.nConstrs, out.aigerBadOutputs)
		}
		for mask := 0; mask < 64; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0, mask&32 != 0}
			want, err := src.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			checkQuantificationOutputs(t, out, inputs, want)
		}
	}
}

func TestMappedAreaBalanceRetainsResetMetadata(t *testing.T) {
	src := Gia_ManStart(1)
	a, r0, r1 := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	root := Gia_ManAppendAnd(src, a, r0)
	for _, lit := range []int32{root, r1 ^ 1, r0, r1} {
		Gia_ManAppendCo(src, lit)
	}
	Gia_ManSetRegNum(src, 2)
	areaReviewInstallMapping(src)
	src.aigerLatchInits = []int32{1, 0}
	src.aigerComment = []byte("area-reset-model")
	out := Gia_ManAreaBalance(src, 1, 4, 0, 0)
	if !slices.Equal(out.aigerLatchInits, src.aigerLatchInits) {
		t.Fatalf("mapped area lost reset state: got%v want%v", out.aigerLatchInits, src.aigerLatchInits)
	}
	if !slices.Equal(out.aigerComment, src.aigerComment) {
		t.Fatal("mapped area lost interface comment")
	}
	checkQuantificationOutputs(t, out, []bool{true, true, false}, []bool{true, true, true, false})
	out.aigerLatchInits[0] = 0
	out.aigerComment[0] = 'X'
	if src.aigerLatchInits[0] != 1 || src.aigerComment[0] != 'a' {
		t.Fatal("mapped result aliases source interface metadata")
	}
}

func TestAreaBalancePublicInterfaceMatrix(t *testing.T) {
	cases := 0
	for family := 0; family < 6; family++ {
		for _, signs := range []int{0, 3, 5, 7} {
			for _, mapped := range []bool{false, true} {
				for simple := int32(0); simple < 2; simple++ {
					for arrival := 0; arrival < 4; arrival++ {
						for _, limit := range []int32{0, 1, 8} {
							src := Gia_ManDupNoMuxes(areaReviewGraph(family, signs), 0)
							src.SetIdentity("area-matrix", "area-matrix.aig")
							if mapped {
								areaReviewInstallMapping(src)
							}
							if err := src.SetAigerProperties(1, 1); err != nil {
								t.Fatal(err)
							}
							src.aigerLatchInits = []int32{1, 2}
							src.aigerComment = []byte("area-interface")
							balanceDepthReviewArrivals(src, arrival, 2)
							out := Gia_ManAreaBalance(src, simple, limit, 0, 0)
							if out.nRegs != 2 || out.nConstrs != 1 || out.aigerBadOutputs != 1 || !slices.Equal(out.aigerLatchInits, src.aigerLatchInits) || !slices.Equal(out.aigerComment, src.aigerComment) {
								t.Fatalf("family%d signs%d mapped%v simple%d arrival%d limit%d: lost interface regs%d constraints%d bad%d resets%v comment%q", family, signs, mapped, simple, arrival, limit, out.nRegs, out.nConstrs, out.aigerBadOutputs, out.aigerLatchInits, out.aigerComment)
							}
							if name, spec := out.Identity(); name != "area-matrix" || spec != "area-matrix.aig" {
								t.Fatalf("identity %q %q", name, spec)
							}
							restored := partitionReviewRoundTrip(t, out, 1, 1)
							for mask := 0; mask < 64; mask++ {
								inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0, mask&32 != 0}
								want, err := src.Evaluate(inputs)
								if err != nil {
									t.Fatal(err)
								}
								checkQuantificationOutputs(t, out, inputs, want)
								checkQuantificationOutputs(t, restored, inputs, want)
							}
							cases++
						}
					}
				}
			}
		}
	}
	t.Logf("%d mapped/unmapped public area rewrites and %d truth rows with serialization", cases, cases*64)
}
