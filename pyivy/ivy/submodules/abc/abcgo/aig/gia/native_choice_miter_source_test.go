package gia

import "testing"

// Gia_ManChoiceMiter_rec in C hashes the two stored fanins of every internal
// object as an AND, including real XOR/MUX objects.
func TestNativeChoiceMiterUsesSourceStoredFanins(t *testing.T) {
	snapshots := make([]*Gia_Man_t, 3)
	for i := range snapshots {
		g := Gia_ManStart(24)
		a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		g.EnableMuxes()
		x := Gia_ManAppendXorReal(g, a, b^1)
		m := Gia_ManAppendMuxReal(g, c, b, d^1)
		Gia_ManAppendCo(g, x^int32(i&1))
		Gia_ManAppendCo(g, m^int32((i>>1)&1))
		snapshots[i] = g
	}
	out := Gia_ManChoiceMiter(snapshots)
	if Gia_ManObjNum(out) != 13 || Gia_ManCoNum(out) != 6 {
		t.Fatalf("nodes=%d outputs=%d", Gia_ManObjNum(out), Gia_ManCoNum(out))
	}
	for mask := 0; mask < 16; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		x, m := in[0] && !in[1], in[1] && !in[3]
		want := []bool{x, !x, x, m, m, !m}
		got, err := out.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("mask %d outputs %v want %v", mask, got, want)
			}
		}
	}
}
