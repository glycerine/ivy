package cec

import (
	"math/bits"
	"math/rand"
	"testing"
)

func TestSimulationWordsReference(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for n := 1; n <= 20; n++ {
		for trial := 0; trial < 50; trial++ {
			a := make([]uint32, n)
			b := make([]uint32, n)
			for i := range a {
				a[i] = r.Uint32()
				b[i] = r.Uint32()
			}
			if trial%5 == 0 {
				copy(b, a)
			}
			if trial%7 == 0 {
				for i := range b {
					b[i] = ^a[i]
				}
			}
			scores := make([]int32, 32*n)
			Cec_ManSimCompareEqualScore(a, b, int32(n), scores)
			first := int32(-1)
			for bit := 0; bit < 32*n; bit++ {
				d := ((a[bit/32]^b[bit/32])>>uint(bit&31))&1 ^ (a[0]^b[0])&1
				if d != 0 {
					if first < 0 {
						first = int32(bit)
					}
					if scores[bit] != 1 {
						t.Fatal("lost score")
					}
				} else if scores[bit] != 0 {
					t.Fatal("wrong score")
				}
			}
			if Cec_ManSimCompareEqualFirstBit(a, b, int32(n)) != first {
				t.Fatal("first mismatch")
			}
			eq := int32(0)
			if first < 0 {
				eq = 1
			}
			if Cec_ManSimCompareEqual(a, b, int32(n)) != eq {
				t.Fatal("equality")
			}
			if Cec_ManSimHashKey(a, int32(n), 1009) != Cec_ManSimHashKey(complementWords(a), int32(n), 1009) {
				t.Fatal("phase-invariant hash")
			}
			var count int
			for _, v := range a {
				if a[0]&1 != 0 {
					v = ^v
				}
				count += bits.OnesCount32(v)
			}
			constScores := make([]int32, 32*n)
			Cec_ManSimCompareConstScore(a, int32(n), constScores)
			sum := int32(0)
			for _, s := range constScores {
				sum += s
			}
			if sum != int32(count) {
				t.Fatal("constant score")
			}
		}
	}
}
func complementWords(a []uint32) []uint32 {
	b := make([]uint32, len(a))
	for i, v := range a {
		b[i] = ^v
	}
	return b
}
