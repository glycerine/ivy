package hop

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestTruthConstantWithoutInputs(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	for _, root := range []*Hop_Obj_t{p.pConst1, Hop_Not(p.pConst1)} {
		scratch := vec.Vec_IntAlloc(0)
		out := Hop_ManConvertAigToTruth(p, root, 9, scratch, 0)
		if len(out) != 16 || vec.Vec_IntSize(scratch) != 16 {
			t.Fatal("constant table size")
		}
		want := ^uint32(0)
		if Hop_IsComplement(root) != 0 {
			want = 0
		}
		for i, v := range out {
			if v != want || uint32(vec.Vec_IntEntry(scratch, i)) != want {
				t.Fatal("constant", i, v)
			}
		}
		six := Hop_ManComputeTruth6(p, root, 6)
		if six != uint64(want)|uint64(want)<<32 {
			t.Fatal("truth6 constant")
		}
	}
}
func TestTruthBooleanGraph(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	a, b, c := Hop_IthVar(p, 0), Hop_IthVar(p, 1), Hop_IthVar(p, 2)
	root := Hop_Mux(p, a, Hop_Exor(p, b, c), Hop_Not(b))
	for _, msb := range []int32{0, 1} {
		out := Hop_ManConvertAigToTruth(p, root, 3, nil, msb)
		for bit := uint32(0); bit < 32; bit++ {
			ai, bi, ci := uint32(0), uint32(1), uint32(2)
			if msb != 0 {
				ai, ci = ci, ai
			}
			av, bv, cv := bit>>ai&1, bit>>bi&1, bit>>ci&1
			want := bv ^ 1
			if av != 0 {
				want = bv ^ cv
			}
			if out[0]>>bit&1 != want {
				t.Fatal(msb, bit)
			}
		}
	}
	small := Hop_ManConvertAigToTruth(p, root, 3, nil, 0)[0]
	six := Hop_ManComputeTruth6(p, root, 3)
	if uint32(six) != small || uint32(six>>32) != small {
		t.Fatal("32/64 mismatch")
	}
	comp := Hop_ManConvertAigToTruth(p, Hop_Not(root), 3, nil, 0)
	if comp[0] != ^small {
		t.Fatal("complement")
	}
}
func TestTruthHighVariableAndRepeatedCalls(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	root := Hop_IthVar(p, 9)
	for _, msb := range []int32{0, 1} {
		out := Hop_ManConvertAigToTruth(p, root, 10, nil, msb)
		for bit := 0; bit < 1024; bit++ {
			v := 9
			if msb != 0 {
				v = 0
			}
			want := uint32(bit >> v & 1)
			if out[bit/32]>>uint(bit%32)&1 != want {
				t.Fatal(msb, bit)
			}
		}
	}
	first := Hop_ManConvertAigToTruth(p, root, 10, nil, 0)
	again := Hop_ManConvertAigToTruth(p, Hop_Not(root), 10, nil, 0)
	for i := range first {
		if first[i] != ^again[i] {
			t.Fatal("result changed across calls")
		}
	}
}
func TestTruthRejectsMissingSupport(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	root := Hop_IthVar(p, 2)
	Hop_ManConvertAigToTruth(p, root, 3, nil, 0)
	for _, f := range []func(){func() { Hop_ManConvertAigToTruth(p, root, 2, nil, 0) }, func() { Hop_ManComputeTruth6(p, root, 2) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("accepted omitted reachable PI")
				}
			}()
			f()
		}()
	}
}
func TestTruthWordOperations(t *testing.T) {
	a, b := []uint32{1, 2}, make([]uint32, 2)
	Hop_ManTruthCopy(b, a, 6)
	if b[1] != 2 {
		t.Fatal("copy")
	}
	Hop_ManTruthNot(b, b, 6)
	if b[0] != ^uint32(1) {
		t.Fatal("not")
	}
	Hop_ManTruthFill(b, 6)
	if b[0] != ^uint32(0) {
		t.Fatal("fill")
	}
	Hop_ManTruthClear(b, 6)
	if b[1] != 0 {
		t.Fatal("clear")
	}
}

func TestTruthSixDirectExorAndSentinels(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	a, b := Hop_IthVar(p, 0), Hop_IthVar(p, 1)
	a.iData = 0
	b.iData = 1
	node := &Hop_Obj_t{Type: AIG_EXOR, pFanin0: Hop_Not(a), pFanin1: b}
	want := ^uint64(0xaaaaaaaaaaaaaaaa) ^ uint64(0xcccccccccccccccc)
	if got := Hop_ManComputeTruth6_rec(p, node); got != want {
		t.Fatal(got, want)
	}
	a.iData = 6
	if Hop_ManComputeTruth6_rec(p, a) != 0 {
		t.Fatal("zero sentinel")
	}
	a.iData = 7
	if Hop_ManComputeTruth6_rec(p, a) != ^uint64(0) {
		t.Fatal("one sentinel")
	}
}
