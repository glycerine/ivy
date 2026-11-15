package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestBidecResynthesisPreservesLogicalPorts(t *testing.T) {
	for _, reverseInputs := range []bool{false, true} {
		for _, reverseOutputs := range []bool{false, true} {
			src := Gia_ManStart(16)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			root := Gia_ManAppendAnd(src, a, b^1)
			Gia_ManAppendCo(src, root)
			Gia_ManAppendCo(src, a^1)
			for _, ports := range []struct {
				reverse bool
				ids     *Vec_Int_t
			}{{reverseInputs, src.vCis}, {reverseOutputs, src.vCos}} {
				if ports.reverse {
					slices.Reverse(Vec_IntArray(ports.ids))
					for i, id := range Vec_IntArray(ports.ids) {
						Gia_ObjSetCioId(src.pObjs[id], int32(i))
					}
				}
			}
			areaReviewInstallMapping(src)
			out := Gia_ManPerformBidec(src, 0)
			for mask := 0; mask < 4; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0}
				want, err := src.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				got, err := out.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(got, want) {
					t.Fatalf("reverse CI%v CO%v assignment%d: got%v want%v", reverseInputs, reverseOutputs, mask, got, want)
				}
			}
		}
	}
}

func TestBidecTruthCacheGrowsWithGraph(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	old := Gia_ManAppendAnd(g, a, b)
	leaves, storage, visited := permutation(a>>1, b>>1), Vec_IntAlloc(0), Vec_IntAlloc(0)
	truth := Gia_ManConvertAigToTruth(g, Gia_ManObj(g, old>>1), leaves, storage, visited)
	if !slices.Equal(truth, []uint32{0x88888888}) {
		t.Fatalf("initial truth%08x", truth)
	}
	root := Gia_ManAppendAnd(g, a, b^1)
	truth = Gia_ManConvertAigToTruth(g, Gia_ManObj(g, root>>1), leaves, storage, visited)
	if !slices.Equal(truth, []uint32{0x22222222}) {
		t.Fatalf("new-node truth%08x, want22222222", truth)
	}
	for _, marker := range Vec_IntArray(g.vTruths) {
		if marker != -1 {
			t.Fatal("truth cache retained a live scratch index")
		}
	}
}

func BenchmarkBidecCutTruth(b *testing.B) {
	for _, n := range []int{5, 10, 15} {
		b.Run(fmt.Sprintf("width%d", n), func(b *testing.B) {
			g := bidecReviewGraph(n, 3)
			root := Gia_ObjFromLit(g, bidecReviewRoot(g, 0))
			leaves, storage, visited := permutation(Vec_IntArray(g.vCis)...), Vec_IntAlloc(0), Vec_IntAlloc(0)
			Gia_ManConvertAigToTruth(g, root, leaves, storage, visited)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Gia_ManConvertAigToTruth(g, root, leaves, storage, visited)
			}
		})
	}
}

func TestBidecTruthResultOwnershipAndTypedWidths(t *testing.T) {
	cases, assignments := 0, 0
	for _, n := range []int{0, 1, 2, 3, 5, 6, 8, 9, 10, 15} {
		for phase := 0; phase < 4; phase++ {
			for _, reverse := range []bool{false, true} {
				g := Gia_ManStart(64)
				g.EnableMuxes()
				cis := make([]int32, n)
				for i := range cis {
					cis[i] = Gia_ManAppendCi(g)
				}
				root := int32(1)
				if n > 0 {
					root = Gia_ManAppendBuf(g, cis[0]^int32(phase&1))
				}
				if n > 1 {
					root = Gia_ManAppendXorReal(g, cis[0]^int32(phase&1), cis[n-1]^int32(phase>>1&1))
				}
				if n > 2 {
					root = Gia_ManAppendMuxReal(g, cis[1]^int32(phase&1), root, cis[0]^1)
					// Shared and opposite-phase fanins exercise borrowed scratch reads.
					branch := Gia_ManAppendAnd(g, root, cis[2]^int32(phase>>1&1))
					root = Gia_ManAppendXorReal(g, root^1, branch)
					root = Gia_ManAppendBuf(g, root^1)
				}
				Gia_ManAppendCo(g, root)
				leaves := permutation(Vec_IntArray(g.vCis)...)
				if reverse {
					slices.Reverse(Vec_IntArray(leaves))
				}
				storage, visited := Vec_IntAlloc(0), Vec_IntAlloc(0)
				truth := Gia_ManConvertAigToTruth(g, Gia_ObjFromLit(g, root), leaves, storage, visited)
				saved := slices.Clone(truth)
				other := Gia_ManConvertAigToTruth(g, Gia_ObjFromLit(g, root^1), leaves, storage, visited)
				for i, word := range other {
					if word != ^saved[i] {
						t.Fatal("opposite root truth is not a complement")
					}
				}
				if !slices.Equal(truth, saved) {
					t.Fatal("later query changed a caller-owned result")
				}
				if len(other) > 0 {
					other[0] ^= 1
				}
				for i := range Vec_IntArray(storage) {
					Vec_IntWriteEntry(storage, i, 0)
				}
				if !slices.Equal(truth, saved) {
					t.Fatal("scratch or another result aliases an earlier result")
				}
				for mask := 0; mask < 1<<n; mask++ {
					inputs := make([]bool, n)
					cutMask := mask
					for i := range inputs {
						inputs[i] = mask>>i&1 != 0
					}
					if reverse {
						cutMask = 0
						for i := 0; i < n; i++ {
							cutMask |= (mask >> i & 1) << (n - 1 - i)
						}
					}
					want, err := g.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					got := truth[cutMask>>5]>>uint(cutMask&31)&1 != 0
					if got != want[0] {
						t.Fatalf("width%d phase%d reversed%v assignment%d: got%v want%v", n, phase, reverse, mask, got, want[0])
					}
					assignments++
				}
				for _, marker := range Vec_IntArray(g.vTruths) {
					if marker != -1 {
						t.Fatal("typed truth conversion leaked a scratch index")
					}
				}
				cases++
			}
		}
	}
	t.Logf("%d typed/width/phase/permutation cases, %d assignments and independently owned results", cases, assignments)
}

func bidecReviewInstallMapping(g *Gia_Man_t) {
	g.VMapping = Vec_IntStart(int(g.nObjs))
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		leaves := giaObjectFanins(g, obj)
		slices.Sort(leaves)
		leaves = slices.Compact(leaves)
		leaves = slices.DeleteFunc(leaves, func(id int32) bool { return id == 0 })
		Vec_IntWriteEntry(g.VMapping, int(obj.id), int32(Vec_IntSize(g.VMapping)))
		Vec_IntPush(g.VMapping, int32(len(leaves)))
		for _, id := range leaves {
			Vec_IntPush(g.VMapping, id)
		}
		Vec_IntPush(g.VMapping, obj.id)
	}
}
func TestBidecResynthesisInterfaceAndTypedLayouts(t *testing.T) {
	cases := 0
	for signs := 0; signs < 8; signs++ {
		for typed := 0; typed < 2; typed++ {
			for layout := 0; layout < 4; layout++ {
				src := windowReviewGraph(signs, typed, layout)
				Gia_ManAppendCo(src, Gia_ManAppendBuf(src, Gia_ObjToLit(src, Gia_ManCi(src, 0))^1))
				Gia_ManAppendCo(src, 0)
				Gia_ManAppendCo(src, 1)
				Gia_ManSetRegNum(src, 1)
				if err := src.SetAigerProperties(1, 1); err != nil {
					t.Fatal(err)
				}
				src.SetIdentity("bidec-interface", "bidec-source.aig")
				src.aigerLatchInits = []int32{1}
				src.aigerComment = []byte("bidec-reset")
				src.VNamesIn = Vec_PtrAlloc(4)
				src.VNamesOut = Vec_PtrAlloc(int(Gia_ManCoNum(src)))
				for i := int32(0); i < Gia_ManCiNum(src); i++ {
					Vec_PtrPush(src.VNamesIn, fmt.Sprintf("input%d", i))
				}
				for i := int32(0); i < Gia_ManCoNum(src); i++ {
					Vec_PtrPush(src.VNamesOut, fmt.Sprintf("output%d", i))
				}
				bidecReviewInstallMapping(src)
				for _, obj := range src.pObjs {
					obj.fMark0 = uint32(obj.id & 1)
					obj.fMark1 = uint32((obj.id + 1) & 1)
				}
				out := Gia_ManPerformBidec(src, 0)
				if out.nRegs != 1 || out.nConstrs != 1 || out.aigerBadOutputs != 1 || !slices.Equal(out.aigerLatchInits, []int32{1}) || string(out.aigerComment) != "bidec-reset" {
					t.Fatal("resynthesis lost unchanged interface metadata")
				}
				if name, spec := out.Identity(); name != "bidec-interface" || spec != "bidec-source.aig" {
					t.Fatal("resynthesis lost identity")
				}
				if !slices.Equal(Vec_PtrArray(out.VNamesIn), Vec_PtrArray(src.VNamesIn)) || !slices.Equal(Vec_PtrArray(out.VNamesOut), Vec_PtrArray(src.VNamesOut)) {
					t.Fatal("resynthesis changed logical IO names")
				}
				restored := partitionReviewRoundTrip(t, out, 1, 1)
				if !slices.Equal(restored.AigerLatchInitialStates(), []int32{1}) {
					t.Fatal("serialized reset state changed")
				}
				for mask := 0; mask < 16; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
					want, err := src.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					checkQuantificationOutputs(t, out, inputs, want)
					checkQuantificationOutputs(t, restored, inputs, want)
				}
				out.aigerLatchInits[0] = 0
				out.aigerComment[0] = 'X'
				Vec_PtrWriteEntry(out.VNamesIn, 0, "changed")
				if src.aigerLatchInits[0] != 1 || string(src.aigerComment) != "bidec-reset" || Vec_PtrEntry(src.VNamesIn, 0) != "input0" {
					t.Fatal("resynthesis result aliases source interface metadata")
				}
				for _, obj := range src.pObjs {
					if obj.fMark0 != uint32(obj.id&1) || obj.fMark1 != uint32((obj.id+1)&1) {
						t.Fatal("resynthesis changed caller marks")
					}
				}
				for _, marker := range Vec_IntArray(src.vTruths) {
					if marker != -1 {
						t.Fatal("resynthesis retained live truth scratch")
					}
				}
				cases++
			}
		}
	}
	t.Logf("%d mapped typed/layout/phase cases with reset/property/name ownership and %d truth rows before and after serialization", cases, cases*16)
}

func TestBidecSupportLimitDiagnostics(t *testing.T) {
	for _, n := range []int{1, 16} {
		g := Gia_ManStart(32)
		var cis []int32
		for i := 0; i < n; i++ {
			cis = append(cis, Gia_ManAppendCi(g))
		}
		root := Gia_ManAppendBuf(g, cis[0])
		Gia_ManAppendCo(g, root)
		g.VMapping = Vec_IntStart(int(g.nObjs))
		Vec_IntWriteEntry(g.VMapping, int(root>>1), int32(Vec_IntSize(g.VMapping)))
		Vec_IntPush(g.VMapping, int32(n))
		for _, lit := range cis {
			Vec_IntPush(g.VMapping, lit>>1)
		}
		Vec_IntPush(g.VMapping, root>>1)
		for _, obj := range g.pObjs {
			obj.Value = 123
		}
		var out *Gia_Man_t
		text := hashReviewStdout(t, func() { out = Gia_ManPerformBidec(g, 0) })
		expected := "Resynthesis is not performed when nodes have less than 2 inputs.\n"
		if n > 15 {
			expected = "Resynthesis is not performed when nodes have more than 15 inputs.\n"
		}
		if out != nil || text != expected {
			t.Fatalf("support%d: output%v diagnostic%q, want nil/%q", n, out, text, expected)
		}
		if g.vTruths != nil {
			t.Fatal("rejected resynthesis initialized scratch")
		}
		for _, obj := range g.pObjs {
			if obj.Value != 123 {
				t.Fatal("rejected resynthesis changed source copy data")
			}
		}
	}
}

func TestBidecInternalCutAndMaximumSupport(t *testing.T) {
	g := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	cutNode := Gia_ManAppendAnd(g, a, b)
	root := Gia_ManAppendAnd(g, cutNode, c^1)
	storage, visited := Vec_IntAlloc(0), Vec_IntAlloc(0)
	for _, cut := range []*Vec_Int_t{permutation(cutNode>>1, c>>1), permutation(c>>1, cutNode>>1)} {
		truth := Gia_ManConvertAigToTruth(g, Gia_ObjFromLit(g, root), cut, storage, visited)
		expected := uint32(0x22222222)
		if Vec_IntEntry(cut, 0) == c>>1 {
			expected = 0x44444444
		}
		if !slices.Equal(truth, []uint32{expected}) {
			t.Fatalf("internal cut: %x, want%x", truth, expected)
		}
	}
	src := Gia_ManStart(64)
	var cis []int32
	for i := 0; i < 15; i++ {
		cis = append(cis, Gia_ManAppendCi(src))
	}
	lit := cis[0]
	for _, next := range cis[1:] {
		lit = Gia_ManAppendAnd(src, lit, next)
	}
	Gia_ManAppendCo(src, lit)
	Gia_ManAppendCo(src, lit^1)
	src.VMapping = Vec_IntStart(int(src.nObjs))
	Vec_IntWriteEntry(src.VMapping, int(lit>>1), int32(Vec_IntSize(src.VMapping)))
	Vec_IntPush(src.VMapping, 15)
	for _, ci := range cis {
		Vec_IntPush(src.VMapping, ci>>1)
	}
	Vec_IntPush(src.VMapping, lit>>1)
	out := Gia_ManPerformBidec(src, 0)
	for mask := 0; mask < 1<<15; mask++ {
		inputs := make([]bool, 15)
		for i := range inputs {
			inputs[i] = mask>>i&1 != 0
		}
		expected := mask == (1<<15)-1
		checkQuantificationOutputs(t, out, inputs, []bool{expected, !expected})
	}
}

func TestBidecConstantFaninsAndCoincidentEdges(t *testing.T) {
	g := Gia_ManStart(32)
	g.fGiaSimple = 1
	g.EnableMuxes()
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	roots := []int32{Gia_ManAppendAnd(g, 0, a), Gia_ManAppendAnd(g, 1, a), Gia_ManAppendAnd(g, a, a), Gia_ManAppendAnd(g, a, a^1), Gia_ManAppendXorReal(g, 1, a), Gia_ManAppendMuxReal(g, b, 1, a), Gia_ManAppendMuxReal(g, b, 0, a), Gia_ManAppendBuf(g, 0), Gia_ManAppendBuf(g, 1)}
	leaves, storage, visited := permutation(a>>1, b>>1), Vec_IntAlloc(0), Vec_IntAlloc(0)
	for i, root := range roots {
		truth := Gia_ManConvertAigToTruth(g, Gia_ObjFromLit(g, root), leaves, storage, visited)
		for mask := 0; mask < 4; mask++ {
			av, bv := mask&1 != 0, mask&2 != 0
			expected := []bool{false, av, av, false, !av, bv || av, !bv && av, false, true}[i]
			if got := truth[0]>>mask&1 != 0; got != expected {
				t.Fatalf("constant/simple root%d assignment%d: got%v want%v", i, mask, got, expected)
			}
		}
	}
}

func TestBidecReportsDanglingNodeCleanup(t *testing.T) {
	src := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	root := Gia_ManAppendAnd(src, a, b)
	Gia_ManAppendAnd(src, a, b^1)
	Gia_ManAppendCo(src, root)
	bidecReviewInstallMapping(src)
	var out *Gia_Man_t
	text := hashReviewStdout(t, func() { out = Gia_ManPerformBidec(src, 0) })
	if want := "Gia_ManPerformBidec() node count before and after:      2 and      1.\n"; text != want {
		t.Fatalf("cleanup diagnostic%q, want%q", text, want)
	}
	if Gia_ManAndNum(out) != 1 {
		t.Fatal("dangling mapped node survived cleanup")
	}
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		checkQuantificationOutputs(t, out, inputs, []bool{inputs[0] && inputs[1]})
	}
}
