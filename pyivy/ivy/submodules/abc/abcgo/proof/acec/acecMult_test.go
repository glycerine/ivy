package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
	"testing"
)

func TestMultiplierNPNCanonicalization(t *testing.T) {
	for truth := uint32(0); truth < 256; truth++ {
		small := vec.Vec_IntAlloc(16)
		got := Extra_TruthCanonNPN2(truth, 3, small)
		want := extra.Extra_TruthCanonNPN(truth, 3)
		if got != want {
			t.Fatalf("truth %02x got %02x want %02x", truth, got, want)
		}
		for _, value := range vec.Vec_IntArray(small) {
			if value&1 != 0 {
				t.Fatalf("non-normalized member %x", value)
			}
		}
		expanded := uint64(truth) * 0x0101010101010101
		members := vec.Vec_WrdAlloc(16)
		got64 := Extra_TruthCanonNPN3(expanded, 3, members)
		if got64 != uint64(want)*0x0101010101010101 {
			t.Fatalf("64-bit truth %02x got %x want %x", truth, got64, want)
		}
	}
}
func TestMultiplierInputDetection(t *testing.T) {
	g := gia.Gia_ManStart(20)
	a, b, c, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	leaves := vec.Vec_WecStart(2)
	for _, pair := range [][2]int32{{a, c}, {a, d}, {b, c}, {b, d}} {
		vec.Vec_WecPush(leaves, 1, gia.Gia_ManAppendAnd(g, pair[0], pair[1]))
	}
	got := append([]int32(nil), vec.Vec_IntArray(Acec_MultDetectInputs1(g, leaves, nil))...)
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	want := []int32{a >> 1, b >> 1, c >> 1, d >> 1}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}
