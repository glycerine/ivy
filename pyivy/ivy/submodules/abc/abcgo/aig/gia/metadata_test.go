package gia

import "testing"

func TestRandomSourceSequence(t *testing.T) {
	z, w := NUMBER1, NUMBER2
	for i := 0; i < 100; i++ {
		z = 36969*(z&65535) + (z >> 16)
		w = 18000*(w&65535) + (w >> 16)
		reset := int32(0)
		if i == 0 {
			reset = 1
		}
		if got, want := Gia_ManRandom(reset), z<<16+w; got != want {
			t.Fatalf("draw %d: %x != %x", i, got, want)
		}
	}
	first := Gia_ManRandom(1)
	if got := Gia_ManRandomW(1); got != uint64(first)<<32|uint64(first) {
		t.Fatal("reset must apply to both halves")
	}
	info := Vec_PtrAlloc(2)
	a, b := []uint32{7, 7, 7}, []uint32{8, 8, 8}
	Vec_PtrPush(info, a)
	Vec_PtrPush(info, b)
	Gia_ManRandom(1)
	Gia_ManRandomInfo(info, 1, 1, 2)
	if a[1] != 7 || b[0] != 8 || b[2] != 8 || b[1] == 8 {
		t.Fatal("random fill modified outside range")
	}
}
func TestMetadataAndCounterexampleOwnership(t *testing.T) {
	p := Gia_ManStart(4)
	a := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	obj := Gia_ManObj(p, a>>1)
	Gia_ObjSetMark0(obj, 3)
	Gia_ObjSetMark1(obj, 2)
	if Gia_ObjMark0(obj) != 1 || Gia_ObjMark1(obj) != 0 {
		t.Fatal("bitfield truncation")
	}
	Gia_ManSetPhase1(p)
	if Gia_ObjPhase(Gia_ManCo(p, 0)) != 1 {
		t.Fatal("output phase")
	}
	p.StartRepresentatives()
	if !p.HasRepresentatives() || Gia_ObjRepr(p, 1) != GIA_VOID {
		t.Fatal("representative initialization")
	}
	p.SetNext(0, 1)
	if p.Next(0) != 1 {
		t.Fatal("next link")
	}
	input := []bool{true}
	if err := p.SetCombinationalCounterexample(input, 0); err != nil {
		t.Fatal(err)
	}
	model := p.CombinationalCounterexample()
	input[0] = false
	if model.PData[0]&1 == 0 {
		t.Fatal("model aliases input")
	}
	if err := p.SetCombinationalCounterexample(nil, 0); err == nil || p.CombinationalCounterexample() != model {
		t.Fatal("invalid replacement destroyed model")
	}
	if p.TakeCombinationalCounterexample() != model || p.CombinationalCounterexample() != nil || model.PData[0]&1 == 0 {
		t.Fatal("take ownership")
	}
}

func TestObjectLinksRegisterPartitionAndMux(t *testing.T) {
	p := Gia_ManStart(8)
	pi, ro := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	mux := Gia_ManAppendMux(p, pi, ro, ro^1)
	Gia_ManAppendCo(p, mux)
	Gia_ManAppendCo(p, ro)
	Gia_ManSetRegNum(p, 1)
	if Gia_ObjIsPi(p, Gia_ManCi(p, 0)) != 1 || Gia_ObjIsRo(p, Gia_ManCi(p, 1)) != 1 || Gia_ObjIsRi(p, Gia_ManCo(p, 1)) != 1 {
		t.Fatal("register partition")
	}
	if Gia_ObjRoToRi(p, Gia_ManCi(p, 1)) != Gia_ManCo(p, 1) || Gia_ObjRiToRo(p, Gia_ManCo(p, 1)) != Gia_ManCi(p, 1) {
		t.Fatal("register pairing")
	}
	obj := Gia_ManObj(p, pi>>1)
	tagged := Gia_Not(obj)
	if Gia_Obj2Lit(p, tagged) != pi^1 || Gia_Lit2Obj(p, pi^1) != tagged {
		t.Fatal("literal handle roundtrip")
	}
	if Gia_ObjFaninId2(p, obj.id) != -1 {
		t.Fatal("ordinary node has mux fanin")
	}
}

func TestPhasePatternRealMuxAndXor(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	xor := Gia_ManAppendObj(p)
	xor.iDiff0 = uint32(xor.id - (b >> 1))
	xor.iDiff1 = uint32(xor.id - (a >> 1))
	p.PMuxes = make([]uint32, p.nObjs)
	mux := Gia_ManAppendObj(p)
	mux.iDiff0 = uint32(mux.id - (a >> 1))
	mux.iDiff1 = uint32(mux.id - (b >> 1))
	p.PMuxes[mux.id] = uint32(c ^ 1)
	Gia_ManAppendCo(p, xor.id<<1)
	Gia_ManAppendCo(p, mux.id<<1|1)
	values := Vec_IntAlloc(3)
	for mask := 0; mask < 8; mask++ {
		Vec_IntClear(values)
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		for _, v := range input {
			Vec_IntPush(values, giaBool(v))
		}
		Gia_ManSetPhasePattern(p, values)
		want, err := p.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		for i, v := range want {
			if Gia_ObjPhase(Gia_ManCo(p, int32(i))) != giaBool(v) {
				t.Fatalf("phase mismatch mask=%d output=%d", mask, i)
			}
		}
	}
	Gia_ManFillValue(p)
	Gia_ManCleanMark01(p)
	Gia_ManCleanPhase(p)
	for _, obj := range p.pObjs {
		if obj.Value != ^uint32(0) || obj.fMark0 != 0 || obj.fMark1 != 0 || obj.fPhase != 0 {
			t.Fatal("scratch reset")
		}
	}
	Gia_ManCleanValue(p)
	for _, obj := range p.pObjs {
		if obj.Value != 0 {
			t.Fatal("value clear")
		}
	}
}

func TestSpecialGateConstructionOrdering(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	p.PMuxes = make([]uint32, p.nObjs)
	x := Gia_ManAppendXorReal(p, a^1, b)
	m := Gia_ManAppendMuxReal(p, c, a, b^1)
	buf := Gia_ManAppendBuf(p, m^1)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, buf)
	for mask := 0; mask < 8; mask++ {
		av, bv, cv := mask&1 != 0, mask&2 != 0, mask&4 != 0
		got, err := p.Evaluate([]bool{av, bv, cv})
		if err != nil {
			t.Fatal(err)
		}
		mux := !bv
		if cv {
			mux = av
		}
		if got[0] != (!av != bv) || got[1] != !mux {
			t.Fatalf("special gates mask=%d: %v", mask, got)
		}
	}
	if p.nXors != 1 || p.nMuxes != 1 || p.nBufs != 1 {
		t.Fatal("gate counters")
	}
}

func TestSourceGateLevels(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	p.PMuxes = make([]uint32, p.nObjs)
	x := Gia_ManAppendXorReal(p, a, b)
	m := Gia_ManAppendMuxReal(p, c, x, a)
	buf := Gia_ManAppendBuf(p, m)
	Gia_ManAppendCo(p, buf)
	Gia_ManCleanLevels(p, p.nObjs)
	for _, obj := range p.pObjs {
		if Gia_ObjIsCo(obj) != 0 {
			Gia_ObjSetCoLevel(p, obj)
		} else {
			Gia_ObjSetGateLevel(p, obj)
		}
	}
	if Gia_ObjLevelId(p, x>>1) != 2 || Gia_ObjLevelId(p, m>>1) != 4 || Gia_ObjLevelId(p, buf>>1) != 4 || Gia_ObjLevel(p, Gia_ManCo(p, 0)) != 4 {
		t.Fatal("source gate weights")
	}
	Gia_ObjUpdateLevelId(p, x>>1, 1)
	if Gia_ObjLevelId(p, x>>1) != 2 {
		t.Fatal("update decreased level")
	}
}

func TestForwardReverseLevelConventions(t *testing.T) {
	p := Gia_ManStart(5)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendXorReal(p, a, b)
	Gia_ManAppendCo(p, x)
	if Gia_ManLevelNum(p) != 2 || Gia_ManLevelAve(p) != 2 {
		t.Fatal("forward XOR gate weighting")
	}
	if Gia_ManLevelRNum(p) != 2 || Gia_ObjLevelId(p, a>>1) != 2 || Gia_ObjLevelId(p, x>>1) != 1 {
		t.Fatal("reverse unit edge weighting")
	}
}

func TestScratchFrameCopiesAndTimingOwnership(t *testing.T) {
	p := Gia_ManStart(4)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	lit := Gia_ManAppendAnd(p, a^1, b)
	obj := Gia_ManObj(p, lit>>1)
	Vec_IntFill(&p.vCopies, int(p.nObjs)*2, -1)
	Gia_ObjSetCopyF(p, 0, Gia_ManObj(p, a>>1), 10)
	Gia_ObjSetCopyF(p, 1, Gia_ManObj(p, a>>1), 20)
	if Gia_ObjFanin0CopyF(p, 0, obj) != 11 || Gia_ObjFanin0CopyF(p, 1, obj) != 21 {
		t.Fatal("frame stride or complement lost")
	}
	Gia_ManCleanCopyArray(p)
	if Gia_ObjCopyArray(p, a>>1) != -1 {
		t.Fatal("copy sentinel")
	}
	Gia_ManTimeStart(p)
	if Gia_ObjTimeArrival(p, 0) != 0 || Gia_ObjTimeRequired(p, 0) != float32(giaInfinity) || Gia_ObjTimeSlack(p, 0) != 0 {
		t.Fatal("timing defaults")
	}
	Gia_ObjSetTimeArrivalObj(p, obj, 3.5)
	Gia_ObjSetTimeRequiredObj(p, obj, 9)
	Gia_ObjSetTimeSlackObj(p, obj, 5.5)
	if Gia_ObjTimeArrivalObj(p, obj) != 3.5 || Gia_ObjTimeRequiredObj(p, obj) != 9 || Gia_ObjTimeSlackObj(p, obj) != 5.5 {
		t.Fatal("timing channels overlap")
	}
	alias := p.vTiming
	Gia_ManTimeStop(p)
	if p.vTiming != nil || Vec_FltEntry(alias, 3*int(obj.id)) != 3.5 {
		t.Fatal("timing release invalidated alias")
	}
}

func TestClassMetadataPredicatesAndVisit(t *testing.T) {
	p := Gia_ManStart(4)
	Gia_ManAppendCi(p)
	Gia_ManAppendCi(p)
	p.StartRepresentatives()
	Gia_ObjSetRepr(p, 2, 1)
	Gia_ObjSetNext(p, 1, 2)
	if Gia_ObjIsHead(p, 1) != 1 || Gia_ObjIsTail(p, 2) != 1 || Gia_ObjIsClass(p, 2) != 1 || Gia_ClassIsPair(p, 1) != 1 {
		t.Fatal("class predicates")
	}
	if Gia_ObjVisitColor(p, 1, 0) != 0 || Gia_ObjVisitColor(p, 1, 0) != 1 {
		t.Fatal("visit must return previous color")
	}
	if Gia_ObjDiffColors(p, 1, 2) != 0 || Gia_ObjDiffColors2(p, 1, 2) != 1 {
		t.Fatal("color conjunction versus disjunction")
	}
	Gia_ClassUndoPair(p, 1)
	if Gia_ObjIsNone(p, 1) != 1 || Gia_ObjIsNone(p, 2) != 1 {
		t.Fatal("undo pair")
	}
	Gia_ObjSetRepr(p, 2, 0)
	if Gia_ObjHasSameRepr(p, 0, 2) != 1 || Gia_ObjIsConst(p, 2) != 1 {
		t.Fatal("constant class")
	}
	Gia_ObjSetFailed(p, 2)
	if Gia_ObjIsFailedPair(p, 0, 2) != 1 {
		t.Fatal("constant failed pair")
	}
}

func TestClassTemporaryNegativeTail(t *testing.T) {
	p := Gia_ManStart(3)
	Gia_ManAppendCi(p)
	Gia_ManAppendCi(p)
	p.StartRepresentatives()
	Gia_ObjSetRepr(p, 2, 1)
	Gia_ObjSetNext(p, 1, 2)
	Gia_ObjSetNext(p, 2, -1)
	if Gia_ObjIsTail(p, 2) != 1 || Gia_ClassIsPair(p, 1) != 1 {
		t.Fatal("CEC -1 tail sentinel")
	}
}

func TestSimplifiedAppendsAndPatch(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	before := p.ObjectCount()
	if Gia_ManAppendAnd2(p, a, a^1) != 0 || Gia_ManAppendAnd2(p, a, 1) != a || Gia_ManAppendXorReal2(p, a, a^1) != 1 || Gia_ManAppendXorReal2(p, a, 1) != a^1 || p.ObjectCount() != before {
		t.Fatal("constant reductions allocated objects")
	}
	majority := Gia_ManAppendMaj2(p, a, b, c)
	Gia_ManAppendCo(p, majority)
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		got, err := p.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, v := range input {
			if v {
				count++
			}
		}
		if got[0] != (count >= 2) {
			t.Fatal("majority")
		}
	}
	Gia_ManPatchCoDriver(p, 0, a^1)
	got, err := p.Evaluate([]bool{false, false, false})
	if err != nil || !got[0] {
		t.Fatal("patched CO", got, err)
	}
	p.fGiaSimple = 1
	Gia_ManAppendAnd2(p, a, a)
	last := p.pObjs[len(p.pObjs)-1]
	if Gia_ObjFaninLit0(last, last.id) != a || Gia_ObjFaninLit1(last, last.id) != a {
		t.Fatal("simple mode must preserve redundant AND")
	}
}

func TestTernaryANDAllValuesAndComplements(t *testing.T) {
	p := Gia_ManStart(6)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	set := func(obj *Gia_Obj_t, value int32) {
		switch value {
		case GIA_ZER:
			Gia_ObjTerSimSet0(obj)
		case GIA_ONE:
			Gia_ObjTerSimSet1(obj)
		case GIA_UND:
			Gia_ObjTerSimSetX(obj)
		}
	}
	for ca := int32(0); ca < 2; ca++ {
		for cb := int32(0); cb < 2; cb++ {
			and := Gia_ManAppendAnd(p, a^ca, b^cb)
			obj := Gia_ManObj(p, and>>1)
			for av := int32(GIA_ZER); av <= GIA_UND; av++ {
				for bv := int32(GIA_ZER); bv <= GIA_UND; bv++ {
					set(Gia_ManObj(p, a>>1), av)
					set(Gia_ManObj(p, b>>1), bv)
					Gia_ObjTerSimAnd(obj)
					want := Gia_XsimAndCond(av, ca, bv, cb)
					if want == GIA_ZER && Gia_ObjTerSimGet0(obj) == 0 || want == GIA_ONE && Gia_ObjTerSimGet1(obj) == 0 || want == GIA_UND && Gia_ObjTerSimGetX(obj) == 0 {
						t.Fatal("ternary AND", av, bv, ca, cb)
					}
				}
			}
		}
	}
}

func TestInputLevelsReverseRequirementsAndSlack(t *testing.T) {
	p := Gia_ManStart(5)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendXorReal(p, a, b)
	Gia_ManAppendCo(p, x)
	input := Vec_IntAlloc(2)
	Vec_IntPush(input, 3)
	Vec_IntPush(input, 1)
	if Gia_ManSetLevels(p, input) != 5 {
		t.Fatal("input level propagation")
	}
	saved := Gia_ManGetCiLevels(p)
	if Vec_IntEntry(saved, 0) != 3 || Vec_IntEntry(saved, 1) != 1 {
		t.Fatal("input extraction")
	}
	reverse := Gia_ManReverseLevel(p)
	required := Gia_ManRequiredLevel(p)
	slacks := Gia_ManComputeSlacks(p)
	if Vec_IntEntry(reverse, int(a>>1)) != 2 || Vec_IntEntry(required, int(x>>1)) != 2 {
		t.Fatal("weighted reverse and required levels")
	}
	for _, value := range Vec_IntArray(slacks) {
		if value < 0 {
			t.Fatal("negative slack in balanced graph")
		}
	}
}

func TestEnableMuxesIdempotent(t *testing.T) {
	p := Gia_ManStart(4)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	p.EnableMuxes()
	mux := Gia_ManAppendMuxReal(p, c, a, b)
	_, _, _, kind := p.Gate(mux >> 1)
	p.EnableMuxes()
	if kind != 4 || Gia_ObjFaninId2(p, mux>>1) != c>>1 {
		t.Fatal("enabling discarded MUX metadata")
	}
}

func TestIdentityNamesAndRepresentativeRemap(t *testing.T) {
	src, dst := Gia_ManStart(4), Gia_ManStart(4)
	for _, p := range []*Gia_Man_t{src, dst} {
		Gia_ManAppendCi(p)
		Gia_ManAppendCi(p)
		Gia_ManAppendCo(p, 2)
	}
	src.pName, src.PSpec = "source", "source.aig"
	src.VNamesIn = Vec_PtrAlloc(2)
	Vec_PtrPush(src.VNamesIn, "a")
	Vec_PtrPush(src.VNamesIn, "b")
	src.VNamesOut = Vec_PtrAlloc(1)
	Vec_PtrPush(src.VNamesOut, "out")
	dst.CopyIdentityFrom(src)
	if err := dst.CopyIONamesFrom(src, []int32{1, -1}, nil); err != nil {
		t.Fatal(err)
	}
	if dst.pName != "source" || Vec_PtrEntry(dst.VNamesIn, 0) != "b" || Vec_PtrEntry(dst.VNamesIn, 1) != nil {
		t.Fatal("identity/name remap")
	}
	old := dst.VNamesIn
	if err := dst.CopyIONamesFrom(src, []int32{20, 1}, nil); err == nil || dst.VNamesIn != old {
		t.Fatal("invalid remap mutated names")
	}
	src.StartRepresentatives()
	Gia_ObjSetRepr(src, 2, 1)
	src.PNexts[1] = 2
	if err := dst.TransferRepresentativesFrom(src, []int32{0, 2, 1, 3}); err != nil {
		t.Fatal(err)
	}
	if Gia_ObjRepr(dst, 2) != 1 || dst.PNexts[1] != 2 {
		t.Fatal("reordered class")
	}
	src.ClearRepresentatives()
	if Gia_ObjRepr(dst, 2) != 1 {
		t.Fatal("representative storage aliases source")
	}
}

func TestMappingRecordViewsAndSimulationBlocks(t *testing.T) {
	p := Gia_ManStart(4)
	Gia_ManAppendCi(p)
	Gia_ManAppendCi(p)
	p.VMapping = Vec_IntAlloc(10)
	for _, v := range []int32{0, 3, 0, 2, 1, 2, -7} {
		Vec_IntPush(p.VMapping, v)
	}
	fanins := Gia_ObjLutFanins(p, 1)
	if len(fanins) != 2 || cap(fanins) != 2 || fanins[0] != 1 || fanins[1] != 2 || Gia_ObjLutMuxId(p, 1) != -7 || Gia_ObjLutIsMux(p, 1) != 1 {
		t.Fatal("mapping record bounds/trailer")
	}
	fanins[0] = 2
	if Gia_ObjLutFanin(p, 1, 0) != 2 {
		t.Fatal("mapping view must be borrowed")
	}
	p.VSimsPi = Vec_WrdAlloc(4)
	for _, v := range []uint64{1, 2, 3, 4} {
		Vec_WrdPush(p.VSimsPi, v)
	}
	first, second := Gia_ObjSimPi(p, 0), Gia_ObjSimPi(p, 1)
	if len(first) != 2 || cap(first) != 2 || len(second) != 2 || second[0] != 3 {
		t.Fatal("simulation strides")
	}
	first[1] = 9
	if Vec_WrdEntry(p.VSimsPi, 1) != 9 || second[0] != 3 {
		t.Fatal("simulation blocks overlap")
	}
}

func TestOppositePhaseSameFaninIsNotBuffer(t *testing.T) {
	p := Gia_ManStart(4)
	a := Gia_ManAppendCi(p)
	p.fGiaSimple = 1
	literal := Gia_ManAppendAnd(p, a, a^1)
	obj := Gia_ManObj(p, literal>>1)
	if Gia_ObjIsBuf(obj) != 0 || Gia_ObjIsAndNotBuf(obj) != 1 {
		t.Fatal("x AND !x is not a buffer")
	}
	p.VMapping = Vec_IntAlloc(8)
	Vec_IntFill(p.VMapping, int(p.nObjs), 0)
	Vec_IntWriteEntry(p.VMapping, int(obj.id), int32(Vec_IntSize(p.VMapping)))
	Vec_IntPush(p.VMapping, 1)
	Vec_IntPush(p.VMapping, a>>1)
	Vec_IntPush(p.VMapping, obj.id)
	leaves, truth, ok := p.LUT(obj.id)
	if !ok || len(leaves) != 1 || len(truth) != 1 || truth[0] != 0 {
		t.Fatal("mapped contradiction truth", leaves, truth, ok)
	}
	Gia_ManCreateRefs(p)
	if p.Refs[a>>1] != 2 {
		t.Fatal("ordinary AND has two arcs to same regular input")
	}
}
