package if_

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"
)

func TestCluCofactorsScalarBoundaries(t *testing.T) {
	rng := rand.New(rand.NewSource(91831))
	for n := int32(1); n <= 16; n++ {
		words := int(If_CluWordNum(n))
		for pattern := 0; pattern < 4; pattern++ {
			truth := make([]uint64, words)
			for i := range truth {
				switch pattern {
				case 0:
					truth[i] = rng.Uint64()
				case 1:
					truth[i] = ^uint64(0)
				case 2: // Zero truth, with dirty output buffers below.
				case 3:
					truth[i] = uint64(1) << uint((13*i+7)%64)
				}
			}
			if n < 6 {
				low := truth[0]
				for at := 0; at < 64; at++ {
					bit := low >> uint(at%(1<<n)) & 1
					truth[0] = (truth[0] &^ (uint64(1) << uint(at))) | (bit << uint(at))
				}
			}
			before := slices.Clone(truth)
			for v := int32(0); v < n; v++ {
				t.Run(fmt.Sprintf("inputs%d/pattern%d/variable%d", n, pattern, v), func(t *testing.T) {
					zero, one := make([]uint64, words+1), make([]uint64, words+1)
					for i := range zero {
						zero[i] = 0x13579bdf2468ace0
						one[i] = 0xfedcba9876543210
					}
					If_CluCofactors(truth, n, v, zero, one)
					for x := 0; x < words*64; x++ {
						low, high := x&^(1<<v), x|(1<<v)
						want0, want1 := truth[low>>6]>>uint(low&63)&1, truth[high>>6]>>uint(high&63)&1
						if got := zero[x>>6] >> uint(x&63) & 1; got != want0 {
							t.Fatalf("cofactor0 assignment%d got%d want%d", x, got, want0)
						}
						if got := one[x>>6] >> uint(x&63) & 1; got != want1 {
							t.Fatalf("cofactor1 assignment%d got%d want%d", x, got, want1)
						}
					}
					if zero[words] != 0x13579bdf2468ace0 || one[words] != 0xfedcba9876543210 {
						t.Fatal("cofactor overwrote output suffix")
					}
					if !slices.Equal(truth, before) {
						t.Fatal("cofactor modified input truth")
					}
				})
			}
		}
	}
}

func BenchmarkCluCofactors(b *testing.B) {
	for _, tc := range []struct{ n, v int32 }{{6, 3}, {10, 7}, {16, 15}} {
		b.Run(fmt.Sprintf("inputs%d/variable%d", tc.n, tc.v), func(b *testing.B) {
			words := If_CluWordNum(tc.n)
			truth, zero, one := make([]uint64, words), make([]uint64, words), make([]uint64, words)
			for i := range truth {
				truth[i] = 0x13579bdf2468ace0 ^ uint64(i)*0x9e3779b97f4a7c15
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				If_CluCofactors(truth, tc.n, tc.v, zero, one)
			}
		})
	}
}

func TestCluWordCountDomain(t *testing.T) {
	for n := int32(0); n <= CLU_VAR_MAX; n++ {
		want := int32(1)
		if n > 6 {
			want = 1 << uint(n-6)
		}
		if got := If_CluWordNum(n); got != want {
			t.Fatalf("inputs%d words%d want%d", n, got, want)
		}
	}
	for _, n := range []int32{-1, CLU_VAR_MAX + 1} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid domain accepted")
				}
			}()
			If_CluWordNum(n)
		})
	}
}

func TestCluSixteenInputThreeLutReconstruction(t *testing.T) {
	const n = 16
	for phase := 0; phase < 4; phase++ {
		t.Run(fmt.Sprint(phase), func(t *testing.T) {
			mask := 0
			if phase&2 != 0 {
				mask = 0x5555
			}
			truth := make([]uint64, 1024)
			for x := 0; x < 1<<n; x++ {
				if (x^mask == 0xffff) != (phase&1 != 0) {
					truth[x>>6] |= 1 << uint(x&63)
				}
			}
			before := slices.Clone(truth)
			// CLU's separate 16-input domain does not require a 16-input IF cut.
			p := &If_Man_t{pPars: &If_Par_t{NLutSize: n}}
			var root, second If_Grp_t
			var f0, f1, f2 uint64
			first := If_CluCheck3(p, truth, n, 6, 6, 6, &root, &second, &f0, &f1, &f2)
			if first.nVars != 6 || second.nVars != 6 || root.nVars != 6 {
				t.Fatalf("groups first%+v second%+v root%+v", first, second, root)
			}
			expectedPins := [3][]int32{{0, 1, 2, 3, 4, 5}, {16, 6, 7, 8, 9, 10}, {17, 11, 12, 13, 14, 15}}
			if phase&2 != 0 {
				// Per-word priority cannot remove the sparse negative
				// cofactors at the high support sizes. The easy first
				// search fails, so C uses its initial high-six bound set.
				// The next stage removes four positive free pins, then its
				// high negative pin becomes detectable at seven inputs
				// because each cofactor is now a single word.
				expectedPins = [3][]int32{{10, 11, 12, 13, 14, 15}, {16, 0, 1, 2, 4, 6}, {17, 3, 5, 7, 8, 9}}
			}
			for index, want := range expectedPins {
				got := []If_Grp_t{first, second, root}[index]
				if !slices.Equal(got.pVars[:6], want) {
					t.Fatalf("group%d pins%v want%v", index, got.pVars[:6], want)
				}
			}
			evaluate := func(g *If_Grp_t, table uint64, values []bool) bool {
				index := 0
				for pin, v := range g.pVars[:g.nVars] {
					if values[v] {
						index |= 1 << pin
					}
				}
				return table>>index&1 != 0
			}
			for x := 0; x < 1<<n; x++ {
				var values [18]bool
				for pin := 0; pin < n; pin++ {
					values[pin] = x>>pin&1 != 0
				}
				values[n] = evaluate(&first, f1, values[:])
				values[n+1] = evaluate(&second, f2, values[:])
				got := evaluate(&root, f0, values[:])
				if want := (x^mask == 0xffff) != (phase&1 != 0); got != want {
					t.Fatalf("assignment%d got%t want%t", x, got, want)
				}
			}
			if !slices.Equal(truth, before) {
				t.Fatal("decomposition modified source truth")
			}
		})
	}
}
