package gia

import (
	"github.com/glycerine/abcgo/bool/kit"
	"testing"
)

func TestManagerStorageAndAppend(t *testing.T) {
	p := Gia_ManStart(1)
	constant := Gia_ManConst0(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	first := Gia_ManObj(p, a>>1)
	and := Gia_ManAppendAnd(p, a, b^1)
	Gia_ManAppendCo(p, and)
	if Gia_ManConst0(p) != constant || Gia_ManObj(p, a>>1) != first || Gia_ObjId(p, first) != a>>1 {
		t.Fatal("object pointer changed during growth")
	}
	if Gia_ObjFanin0(Gia_ManObj(p, and>>1)) != first {
		t.Fatal("backward edge lookup")
	}
	graph, err := GraphFromObjects(p.pObjs)
	if err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 4; mask++ {
		av, bv := mask&1 != 0, mask&2 != 0
		if graph.Evaluate([]bool{av, bv})[0] != (av && !bv) {
			t.Fatal(mask)
		}
	}
	Gia_ManStopP(&p)
	if p != nil {
		t.Fatal("manager pointer not cleared")
	}
}

func TestManagerBooleanAppend(t *testing.T) {
	p := Gia_ManStart(4)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	for _, l := range []int32{Gia_ManAppendOr(p, a, b), Gia_ManAppendXor(p, a, b), Gia_ManAppendMux(p, c, a, b), Gia_ManAppendMaj(p, a, b, c)} {
		Gia_ManAppendCo(p, l)
	}
	graph, err := GraphFromObjects(p.pObjs)
	if err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 8; mask++ {
		av, bv, cv := mask&1 != 0, mask&2 != 0, mask&4 != 0
		mux := bv
		if cv {
			mux = av
		}
		want := []bool{av || bv, av != bv, mux, (av && bv) || (av && cv) || (bv && cv)}
		got := graph.Evaluate([]bool{av, bv, cv})
		for i := range want {
			if got[i] != want[i] {
				t.Fatal(mask, i, got, want)
			}
		}
	}
}

func TestManagerBridgeRoundTrip(t *testing.T) {
	p := Gia_ManStart(4)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendXor(p, a, b))
	Gia_ManAppendCo(p, a)
	p.nRegs = 1
	net, err := Gia_ManToBridgeNetlist(p)
	if err != nil {
		t.Fatal(err)
	}
	copy, err := Gia_ManFromBridgeNetlist(net)
	if err != nil {
		t.Fatal(err)
	}
	if Gia_ManPiNum(copy) != 1 || Gia_ManPoNum(copy) != 1 || Gia_ManRegNum(copy) != 1 {
		t.Fatal("register partition")
	}
	g1, e1 := GraphFromObjects(p.pObjs)
	g2, e2 := GraphFromObjects(copy.pObjs)
	if e1 != nil || e2 != nil {
		t.Fatal(e1, e2)
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		x, y := g1.Evaluate(in), g2.Evaluate(in)
		for i := range x {
			if x[i] != y[i] {
				t.Fatal(mask, x, y)
			}
		}
	}
}

func TestManagerAppendAuxiliaryState(t *testing.T) {
	p := Gia_ManStart(2)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	p.pFanData = make([]int32, 5*16)
	p.nFansAlloc = 16
	p.NSimWords = 1
	p.vSims = Vec_WrdStart(int(p.nObjs))
	Vec_WrdWriteEntry(p.vSims, int(a>>1), 0xaaaaaaaaaaaaaaaa)
	Vec_WrdWriteEntry(p.vSims, int(b>>1), 0xcccccccccccccccc)
	p.fBuiltInSim = 1
	p.fSweeper = 1
	p.nSuppWords = 1
	p.vSuppWords = Vec_WrdStart(int(p.nObjs))
	Vec_WrdWriteEntry(p.vSuppWords, int(a>>1), 1)
	Vec_WrdWriteEntry(p.vSuppWords, int(b>>1), 2)
	and := Gia_ManAppendAnd(p, a, b^1)
	obj := Gia_ManObj(p, and>>1)
	if got := Vec_WrdEntry(p.vSims, int(and>>1)); got != 0x2222222222222222 {
		t.Fatalf("simulation %x", got)
	}
	if got := Vec_WrdEntry(p.vSuppWords, int(and>>1)); got != 3 {
		t.Fatal("support", got)
	}
	if p.pFanData[5*(a>>1)] != and || p.pFanData[5*(b>>1)] != and|1 {
		t.Fatal("fanin fanout edges")
	}
	if Gia_ManObj(p, a>>1).fMark0 != 1 || Gia_ManObj(p, b>>1).fMark0 != 1 {
		t.Fatal("sweeper marks")
	}
	p.fBuiltInSim = 0
	p.vSuppWords = nil
	co := Gia_ManAppendCo(p, and)
	if p.pFanData[5*obj.id] != co {
		t.Fatal("CO fanout")
	}
}

func TestManagerStaticAndDynamicFanouts(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, c)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y)
	Gia_ManStaticFanoutStart(p)
	if Gia_ObjFanoutNumId(p, a>>1) != 2 || Gia_ObjFanoutId(p, a>>1, 0) != x>>1 || Gia_ObjFanoutId(p, a>>1, 1) != y>>1 {
		t.Fatal("static fanout order")
	}
	Gia_ManStaticFanoutStop(p)
	Gia_ManFanoutStart(p)
	first := Gia_ObjFanout0Int(p, a>>1)
	second := Gia_ObjFanoutNext(p, first)
	if first != x || second != y || Gia_ObjFanoutNext(p, second) != first {
		t.Fatal("dynamic fanout ring", first, second)
	}
	Gia_ObjRemoveFanout(p, Gia_ManObj(p, a>>1), Gia_ManObj(p, x>>1))
	if Gia_ObjFanout0Int(p, a>>1) != y || Gia_ObjFanoutNext(p, y) != y {
		t.Fatal("remove first")
	}
	Gia_ObjRemoveFanout(p, Gia_ManObj(p, a>>1), Gia_ManObj(p, y>>1))
	if Gia_ObjFanout0Int(p, a>>1) != 0 {
		t.Fatal("remove last")
	}
	Gia_ManFanoutStop(p)
}

func TestManagerObjectStoreTracksCanonicalSlice(t *testing.T) {
	p := Gia_ManStart(1)
	a := Gia_ManAppendCi(p)
	// Legacy manager algorithms can replace the backing slice while preserving
	// objects. Fanin lookup must follow the canonical manager field, not a cache.
	replacement := make([]*Gia_Obj_t, len(p.pObjs), len(p.pObjs)+2)
	copy(replacement, p.pObjs)
	p.pObjs = replacement
	obj := Gia_ManAppendObj(p)
	obj.iDiff0 = uint32(obj.id - (a >> 1))
	obj.iDiff1 = obj.iDiff0
	if Gia_ObjFanin0(obj) != Gia_ManObj(p, a>>1) {
		t.Fatal("stale object store")
	}
}

func TestSolverRepresentationRoundTrip(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendXor(p, a, b))
	Gia_ManAppendCo(p, 0)
	Gia_ManAppendCo(p, 1)
	p.nRegs = 1
	solver := Slv_ManFromGia(p)
	copy := Slv_ManToGia(solver)
	if Gia_ManObjNum(copy) != Gia_ManObjNum(p) || copy.nRegs != 1 {
		t.Fatal("manager dimensions")
	}
	x, e1 := GraphFromObjects(p.pObjs)
	y, e2 := GraphFromObjects(copy.pObjs)
	if e1 != nil || e2 != nil {
		t.Fatal(e1, e2)
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		a, b := x.Evaluate(in), y.Evaluate(in)
		for i := range a {
			if a[i] != b[i] {
				t.Fatal(mask, a, b)
			}
		}
	}
	if Slv_ObjIsCo(solver, solver.vCos[1]) != 1 {
		t.Fatal("constant-driven CO confused with CI")
	}
}

func TestReferenceReleaseKeepsAliasesAndRebuilds(t *testing.T) {
	p := Gia_ManStart(4)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	and := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, and)
	Gia_ManCreateRefs(p)
	alias := p.Refs
	if alias[a>>1] != 1 || alias[b>>1] != 1 || alias[and>>1] != 1 {
		t.Fatal(alias)
	}
	p.Refs = nil
	if alias[a>>1] != 1 {
		t.Fatal("release destroyed aliased storage")
	}
	Gia_ManCreateRefs(p)
	if Gia_ObjRefIncId(p, a>>1) != 1 || Gia_ObjRefDecId(p, a>>1) != 1 {
		t.Fatal("post/pre count semantics")
	}
	p.Refs[a>>1] = 9
	if alias[a>>1] != 1 {
		t.Fatal("rebuilt refs reused aliased storage")
	}
	Gia_ManStop(p)
	if p.Refs != nil || alias[a>>1] != 1 {
		t.Fatal("manager release lost alias semantics")
	}
}

func TestKitManagerInterfaceHashing(t *testing.T) {
	p := Gia_ManStart(4)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	existing := Gia_ManAppendAnd(p, a, b)
	before := p.nObjs
	if p.Const1() != 1 || p.InputLiteral(0) != a || p.AndLiteral(b, a, true) != existing || p.nObjs != before {
		t.Fatal("bridge manager interface")
	}
	if p.AndLiteral(a, a^1, true) != 0 || p.AndLiteral(a, 1, true) != a {
		t.Fatal("AND reductions")
	}
	raw := p.AndLiteral(a, b, false)
	if raw == existing || p.nObjs != before+1 {
		t.Fatal("raw append unexpectedly hashed")
	}
}

func TestRealKitTruthToGiaBridge(t *testing.T) {
	for _, truth := range []uint32{0, 0xffffffff, 0x88888888, 0x66666666, 0xeeeeeeee} {
		p := Gia_ManStart(8)
		Gia_ManAppendCi(p)
		Gia_ManAppendCi(p)
		var memory []uint32
		lit := kit.Kit_TruthToGia(p, []uint32{truth}, 2, &memory, nil, 1)
		Gia_ManAppendCo(p, lit)
		graph, err := GraphFromObjects(p.pObjs)
		if err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 4; mask++ {
			got := graph.Evaluate([]bool{mask&1 != 0, mask&2 != 0})[0]
			want := truth>>uint(mask)&1 != 0
			if got != want {
				t.Fatalf("truth %x assignment %d", truth, mask)
			}
		}
	}
}

func TestPackedAgiPortAndSupport(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, a)
	p.nRegs = 1
	if Gia_ManSuppSizeOne(p, Gia_ManObj(p, y>>1)) != 3 || Gia_ManSuppSize(p, []int32{x >> 1, y >> 1}, 2) != 3 {
		t.Fatal("GIA support union")
	}
	p.nTravIds = int32(^uint32(0) >> 1)
	if Gia_ManSuppSizeOne(p, Gia_ManObj(p, y>>1)) != 3 || p.nTravIds != 1 {
		t.Fatal("GIA traversal rollover")
	}
	packed := Agi_ManFromGia(p)
	if Agi_ManObjNum(packed) != p.nObjs || Agi_ManCiNum(packed) != 3 || Agi_ManCoNum(packed) != 2 || Agi_ManNodeNum(packed) != 2 {
		t.Fatal("packed dimensions")
	}
	if Agi_ManSuppSizeOne(packed, y>>1) != 3 || Agi_ManSuppSizeOne(packed, 0) != 0 {
		t.Fatal("packed support")
	}
	if Agi_ObjIsPi(packed, a>>1) != 1 || Agi_ObjIsRo(packed, c>>1) != 1 || Agi_ObjIsCo(packed, a>>1) != 0 {
		t.Fatal("PI/RO classification")
	}
	if Agi_ObjIsPo(packed, packed.vCos[0]) != 1 || Agi_ObjIsRi(packed, packed.vCos[1]) != 1 || Agi_ObjIsCo(packed, packed.vCos[1]) != 1 {
		t.Fatal("PO/RI classification")
	}
	if Agi_ObjIsMux(packed, x>>1) != 0 {
		t.Fatal("missing third edge should be sentinel")
	}
	packed.nTravIds = ^uint32(0)
	if Agi_ManSuppSizeOne(packed, y>>1) != 3 || packed.nTravIds != 1 {
		t.Fatal("AGI traversal rollover")
	}
	alias := packed.pObjs
	Agi_ManFree(packed)
	if packed.pObjs != nil || alias[0] != AGI_C0 {
		t.Fatal("packed release invalidated alias")
	}
}

func TestPackedAgiStorageGrowthAndSetters(t *testing.T) {
	p := Agi_ManAlloc(1)
	for i := 0; i < 40; i++ {
		Agi_ManAppendCi(p)
	}
	node := Agi_ManAppendAnd(p, 2, 4) >> 1
	Agi_ObjSetLit0(p, node, 3)
	Agi_ObjSetLit1(p, node, 6)
	Agi_ObjSetLit2(p, node, 8)
	if Agi_ObjLit0(p, node) != 3 || Agi_ObjLit1(p, node) != 6 || Agi_ObjLit2(p, node) != 8 || Agi_ObjIsMux(p, node) != 1 {
		t.Fatal("packed fields")
	}
	next := Agi_ManAppendAnd(p, 2, 4) >> 1
	if Agi_ObjLit2(p, next) != ^uint32(0) {
		t.Fatal("new object's third edge not initialized")
	}
}

func TestDependencyNeutralNetworkAndRealGates(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	xor := Gia_ManAppendObj(p)
	xor.iDiff0 = uint32(xor.id - (b >> 1))
	xor.iDiff1 = uint32(xor.id - (a >> 1))
	p.PMuxes = make([]uint32, p.nObjs)
	mux := Gia_ManAppendObj(p)
	mux.iDiff0 = uint32(mux.id - (a >> 1))
	mux.iDiff1 = uint32(mux.id - (b >> 1))
	p.PMuxes[mux.id] = uint32(c)
	Gia_ManAppendCo(p, xor.id<<1)
	Gia_ManAppendCo(p, mux.id<<1)
	if _, _, _, kind := p.Gate(xor.id); kind != 3 {
		t.Fatal("XOR kind")
	}
	control, then, otherwise, kind := p.Gate(mux.id)
	if kind != 4 || control != c || then != b || otherwise != a {
		t.Fatal("MUX operands", control, then, otherwise, kind)
	}
	if p.ObjectCount() != int(p.nObjs) || len(p.InputLiterals()) != 3 || len(p.OutputLiterals()) != 2 || p.RegisterCount() != 0 {
		t.Fatal("network dimensions")
	}
	for mask := 0; mask < 8; mask++ {
		av, bv, cv := mask&1 != 0, mask&2 != 0, mask&4 != 0
		out, err := p.Evaluate([]bool{av, bv, cv})
		if err != nil {
			t.Fatal(err)
		}
		expected := av
		if cv {
			expected = bv
		}
		if out[0] != (av != bv) || out[1] != expected {
			t.Fatal(mask, out)
		}
	}
	if _, err := p.Evaluate(nil); err == nil {
		t.Fatal("missing input validation")
	}
	original := mux.iDiff0
	mux.iDiff0 = 0
	if _, err := p.Evaluate([]bool{false, false, false}); err == nil {
		t.Fatal("forward edge accepted")
	}
	mux.iDiff0 = original
	ids := p.InputObjectIDs()
	ids[0] = 99
	if p.InputLiterals()[0] != a {
		t.Fatal("input IDs exposed mutable storage")
	}
}

func TestRepresentativeChainsAndCycles(t *testing.T) {
	p := Gia_ManStart(4)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	n := Gia_ManAppendAnd(p, a, b)
	p.PNexts = make([]int32, p.nObjs)
	p.PNexts[a>>1] = b >> 1
	p.PNexts[b>>1] = n >> 1
	Gia_ManDeriveReprs(p)
	if Gia_ObjRepr(p, b>>1) != a>>1 || Gia_ObjReprObj(p, n>>1) != Gia_ManObj(p, a>>1) {
		t.Fatal("class representative")
	}
	p.PReprs = nil
	p.PNexts[n>>1] = b >> 1
	defer func() {
		if recover() == nil {
			t.Fatal("cyclic choices accepted")
		}
	}()
	Gia_ManDeriveReprs(p)
}

func TestRepresentativeMergesPreserveReorderedClasses(t *testing.T) {
	p := Gia_ManStart(4)
	Gia_ManAppendCi(p)
	Gia_ManAppendCi(p)
	Gia_ManAppendCi(p)
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
	giaMergeRepr(p, 1, 3)
	giaMergeRepr(p, 2, 3)
	giaFlattenReprs(p)
	if Gia_ObjRepr(p, 2) != 1 || Gia_ObjRepr(p, 3) != 1 {
		t.Fatal("lost class member during reordered merge")
	}
	next := Gia_ManDeriveNexts(p)
	if next[1] != 2 || next[2] != 3 || next[3] != 0 {
		t.Fatal(next)
	}
}

func TestRegisterMetadataMayPrecedeConstruction(t *testing.T) {
	p := Gia_ManStart(4)
	Gia_ManSetRegNum(p, 1)
	Gia_ManAppendCi(p)
	state := Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, state)
	Gia_ManAppendCo(p, state)
	if Gia_ManPiNum(p) != 1 || Gia_ManPoNum(p) != 1 || p.RegisterCount() != 1 {
		t.Fatal("register metadata construction order")
	}
}
