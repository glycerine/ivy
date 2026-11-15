package gia

import (
	"slices"
	"testing"
)

func TestTrimDualRetainsConstantMismatch(t *testing.T) {
	src := Gia_ManStart(1)
	Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, 0)
	Gia_ManAppendCo(src, 0) // Equal pair can be removed.
	Gia_ManAppendCo(src, 0)
	Gia_ManAppendCo(src, 1) // This mismatch must survive.
	if got := Vec_IntArray(Gia_ManDupTrimmedNonZero(src)); !slices.Equal(got, []int32{1}) {
		t.Errorf("retained pair indices %v, want [1]", got)
	}
	out := Gia_ManDupTrimmed(src, 1, 1, 1, -1)
	if Gia_ManPoNum(out) != 2 {
		t.Fatalf("POs%d, want one retained pair", Gia_ManPoNum(out))
	}
	checkQuantificationOutputs(t, out, []bool{false}, []bool{false, true})
}

func TestTrimmingRetainsSelectedOutputRoles(t *testing.T) {
	for _, mode := range []string{"unchanged", "constant_trim", "pi_trim"} {
		t.Run(mode, func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			x := Gia_ManAppendAnd(src, a, b)
			Gia_ManAppendCo(src, a)
			Gia_ManAppendCo(src, 0)
			Gia_ManAppendCo(src, x)
			Gia_ManAppendCo(src, x^1)
			Gia_ManAppendCo(src, r)
			Gia_ManSetRegNum(src, 1)
			src.nConstrs = 1
			src.aigerBadOutputs = 1
			var out *Gia_Man_t
			switch mode {
			case "unchanged":
				out = Gia_ManDupTrimmed(src, 0, 0, 0, -1)
			case "constant_trim":
				out = Gia_ManDupTrimmed(src, 1, 1, 0, -1)
			case "pi_trim":
				out = Gia_ManDupTrimmed2(src)
			}
			if out.nConstrs != 1 || out.aigerBadOutputs != 1 || out.nRegs != 1 {
				t.Fatalf("lost roles: constraints%d bad%d registers%d", out.nConstrs, out.aigerBadOutputs, out.nRegs)
			}
		})
	}
}

func TestCiMapRetainsBadOutputRole(t *testing.T) {
	src := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, r)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a)
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	src.aigerBadOutputs = 1
	out := Gia_ManDupDfsCiMap(src, []int32{-1, -1}, nil)
	if out.nConstrs != 1 || out.aigerBadOutputs != 1 {
		t.Fatalf("CI map roles: constraints%d bad%d", out.nConstrs, out.aigerBadOutputs)
	}
}

func TestTrimDualRetainsStatefulOutputRoles(t *testing.T) {
	src := Gia_ManStart(1)
	a, b, c, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	for _, lit := range []int32{r, a, b, c, a} {
		Gia_ManAppendCo(src, lit)
	}
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	src.aigerBadOutputs = 1
	out := Gia_ManDupTrimmed(src, 1, 1, 1, -1)
	if out.nConstrs != 1 || out.aigerBadOutputs != 1 {
		t.Fatalf("dual stateful roles: constraints%d bad%d", out.nConstrs, out.aigerBadOutputs)
	}
}

func BenchmarkTrimBufferChain(b *testing.B) {
	src := Gia_ManStart(1)
	a, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	x := Gia_ManAppendAnd(src, a, c)
	for i := 0; i < 256; i++ {
		x = Gia_ManAppendBuf(src, x)
	}
	Gia_ManAppendCo(src, x)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Gia_ManDupTrimmed(src, 1, 1, 0, -1)
	}
}
