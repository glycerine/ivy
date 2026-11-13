package ssw

import (
	"math/rand"
	"testing"
)

func TestRarityBitWordCount(t *testing.T) {
	for _, tc := range []struct{ bits, words int32 }{{0, 0}, {1, 1}, {63, 1}, {64, 1}, {65, 2}, {128, 2}, {129, 3}, {2147483647, 33554432}} {
		if got := Ssw_RarBitWordNum(tc.bits); got != tc.words {
			t.Fatalf("bits=%d got=%d want=%d", tc.bits, got, tc.words)
		}
	}
}

func TestTranspose64AgainstBitOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(6012))
	for trial := 0; trial < 100; trial++ {
		var in, got [64]uint64
		for i := range in {
			in[i] = rng.Uint64()
		}
		got = in
		transpose64(got[:])
		var slow [64]uint64
		transpose64Simple(in[:], slow[:])
		if got != slow {
			t.Fatalf("trial=%d fast and reference transposes disagree", trial)
		}
		for row := 0; row < 64; row++ {
			for col := 0; col < 64; col++ {
				want := in[63-col] >> (63 - row) & 1
				if bit := got[row] >> col & 1; bit != want {
					t.Fatalf("trial=%d row=%d col=%d got=%d want=%d", trial, row, col, bit, want)
				}
			}
		}
		transpose64(got[:])
		if got != in {
			t.Fatalf("trial=%d transpose is not involutive", trial)
		}
	}
}

func TestTranspose32AgainstBitOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(693))
	for trial := 0; trial < 100; trial++ {
		var in, got [32]uint32
		for i := range in {
			in[i] = rng.Uint32()
		}
		got = in
		transpose32(got[:])
		for row := 0; row < 32; row++ {
			for col := 0; col < 32; col++ {
				want := in[31-col] >> (31 - row) & 1
				if bit := got[row] >> col & 1; bit != want {
					t.Fatalf("trial=%d row=%d col=%d got=%d want=%d", trial, row, col, bit, want)
				}
			}
		}
	}
}
