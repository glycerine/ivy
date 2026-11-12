package dau

import (
	"math/bits"
	"math/rand"
	"testing"
)

func TestQuickCountsHighWidthAndFullByte(t *testing.T) {
	for n := int32(1); n <= 16; n++ {
		for trial := 0; trial < 6; trial++ {
			rng := rand.New(rand.NewSource(int64(n*10 + int32(trial))))
			table := make([]uint64, dauCountWords(n))
			for i := range table {
				table[i] = rng.Uint64()
				if trial == 0 {
					table[i] = ^uint64(0)
				}
			}
			Abc_TtNormalizeSmallTruth(table, n)
			a, b := make([]int32, n), make([]int32, n)
			want := Abc_TtCountOnesInCofsSimple(table, n, a)
			got := Abc_TtCountOnesInCofsQuick(table, n, b)
			if got != want {
				t.Fatalf("n%d total %d != %d", n, got, want)
			}
			for v := range a {
				if a[v] != b[v] {
					t.Fatalf("n%d cofactor%d %d != %d", n, v, b[v], a[v])
				}
			}
		}
	}
}
func TestCountsAndShiftedCoefficientsReference(t *testing.T) {
	for n := int32(0); n <= 10; n++ {
		rng := rand.New(rand.NewSource(int64(n)))
		for trial := 0; trial < 20; trial++ {
			table := make([]uint64, dauCountWords(n))
			for i := range table {
				table[i] = rng.Uint64()
			}
			Abc_TtNormalizeSmallTruth(table, n)
			store := make([]int32, n)
			Abc_TtSccInCofs(table, n, store)
			want := int32(0)
			cof := make([]int32, n)
			for bit := 0; bit < 64*len(table); bit++ {
				if table[bit>>6]&(1<<uint(bit&63)) == 0 {
					continue
				}
				want += 1 << uint(bits.OnesCount(uint(bit)))
				for v := int32(0); v < n; v++ {
					if bit&(1<<uint(v)) == 0 {
						compressed := (bit & ((1 << uint(v)) - 1)) | (bit>>uint(v+1))<<uint(v)
						cof[v] += 1 << uint(bits.OnesCount(uint(compressed)))
					}
				}
			}
			for v := range cof {
				if cof[v] != store[v] {
					t.Fatalf("n%d SCC cofactor%d = %d expected%d", n, v, store[v], cof[v])
				}
			}
			if got := Abc_TtScc(table, n); got != want {
				t.Fatal(n, got, want)
			}
		}
	}
}
func TestFastCountUsesActualMinterms(t *testing.T) {
	rng := rand.New(rand.NewSource(41))
	for n := int32(3); n <= 16; n++ {
		for trial := 0; trial < 10; trial++ {
			table := make([]uint64, dauCountWords(n))
			for i := range table {
				table[i] = rng.Uint64()
				if trial == 0 {
					table[i] = ^uint64(0)
				}
			}
			Abc_TtNormalizeSmallTruth(table, n)
			a, b := make([]int32, n), make([]int32, n)
			total := Abc_TtCountOnesInCofsSimple(table, n, a)
			factor := int32(1)
			if n < 6 {
				factor = 1 << uint(6-n)
			}
			got := Abc_TtCountOnesInCofsFast(table, n, b)
			if got != total/factor {
				t.Fatal(n, got, total/factor)
			}
			for v := range a {
				if b[v] != a[v]/factor {
					t.Fatal(n, v, b[v], a[v]/factor)
				}
			}
		}
	}
}
