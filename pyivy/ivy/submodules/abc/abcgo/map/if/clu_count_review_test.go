package if_

import (
	"fmt"
	"slices"
	"testing"
)

func TestCluCofactorCountDimensionsAndOutputs(t *testing.T) {
	for bound := 2; bound <= 7; bound++ {
		for free := 1; free+bound <= 16; free++ {
			for requested := 1; requested <= 5; requested++ {
				count := min(requested, 1<<bound)
				if free == 1 {
					count = min(count, 4)
				}
				shifts := []int{0}
				if free < 6 {
					shifts = append(shifts, 1<<(bound+free))
				}
				for _, shift := range shifts {
					t.Run(fmt.Sprintf("bound%d/free%d/classes%d/shift%d", bound, free, count, shift), func(t *testing.T) {
						n := bound + free
						words := int(If_CluWordNum(int32(free)))
						truth := make([]uint64, (shift+(1<<n)+63)/64)
						for i := range truth {
							truth[i] = ^uint64(0)
						}
						functions := make([][]uint64, count)
						for class := range functions {
							functions[class] = make([]uint64, words)
							if free == 1 {
								functions[class][0] = uint64(class)
								continue
							}
							// Distinguish sparse cofactors at distant positions, including the
							// final word, so equal prefixes must not collapse different rows.
							for bit, position := range []int{(1 << free) - 1, (1 << (free - 1)) - 1, 0} {
								if class>>bit&1 != 0 {
									functions[class][position>>6] |= 1 << uint(position&63)
								}
							}
						}
						var seen []int
						var selector uint64
						for row := 0; row < 1<<bound; row++ {
							class := (row ^ (row >> 1)) % count
							index := slices.Index(seen, class)
							if index < 0 {
								index = len(seen)
								seen = append(seen, class)
							}
							if index != 0 && row < 64 {
								selector |= 1 << row
							}
							for x := 0; x < 1<<free; x++ {
								at := shift + (row << free) + x
								bit := functions[class][x>>6] >> uint(x&63) & 1
								truth[at>>6] = (truth[at>>6] &^ (uint64(1) << uint(at&63))) | (bit << uint(at&63))
							}
						}
						before := slices.Clone(truth)
						var outputs [][]uint64
						if bound <= 6 {
							outputs = make([][]uint64, 3)
							for i := range outputs {
								outputs[i] = make([]uint64, words+1)
								for j := range outputs[i] {
									outputs[i][j] = 0x13579bdf2468ace0
								}
							}
						}
						got := If_CluCountCofs(truth, int32(n), int32(bound), int32(shift), outputs)
						if got != int32(count) {
							t.Fatalf("count%d want%d", got, count)
						}
						if bound <= 6 {
							for i, out := range outputs {
								for w, actual := range out {
									want := uint64(0x13579bdf2468ace0)
									if count <= 2 {
										if i < 2 && w < words {
											want = functions[seen[min(i, len(seen)-1)]][w]
										}
										if i == 2 && w == 0 {
											want = selector
										}
									}
									if actual != want {
										t.Fatalf("output%d word%d got%#x want%#x", i, w, actual, want)
									}
								}
							}
						}
						if !slices.Equal(truth, before) {
							t.Fatal("counter modified input truth")
						}
					})
				}
			}
		}
	}
}

func BenchmarkCluCofactorCounting(b *testing.B) {
	for _, tc := range []struct{ n, bound int32 }{{6, 4}, {10, 4}, {16, 6}} {
		b.Run(fmt.Sprintf("inputs%d/bound%d", tc.n, tc.bound), func(b *testing.B) {
			truth := make([]uint64, If_CluWordNum(tc.n))
			free := tc.n - tc.bound
			for x := 0; x < 1<<tc.n; x++ {
				if (x>>free)&1 != 0 && (x&((1<<free)-1)) == 0 {
					truth[x>>6] |= 1 << uint(x&63)
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if got := If_CluCountCofs(truth, tc.n, tc.bound, 0, nil); got != 2 {
					b.Fatal(got)
				}
			}
		})
	}
}

// The old Go helper also accepted unaligned offsets, beyond the C caller's
// aligned layout. Keep that behavior while optimizing the aligned fast path.
func TestCluCofactorUnalignedOffsetsAndOwnedRows(t *testing.T) {
	for free := 1; free <= 7; free++ {
		for _, shift := range []int{1, 3, 63} {
			t.Run(fmt.Sprintf("free%d/shift%d", free, shift), func(t *testing.T) {
				const bound = 2
				n := free + bound
				words := int(If_CluWordNum(int32(free)))
				expected := make([]uint64, words)
				expected[0] = 1
				last := (1 << free) - 1
				expected[last>>6] |= 1 << uint(last&63)
				truth := make([]uint64, (shift+(1<<n)+63)/64)
				for row := 0; row < 4; row++ {
					if row&1 == 0 {
						continue
					}
					for x := 0; x < 1<<free; x++ {
						if expected[x>>6]>>uint(x&63)&1 != 0 {
							bit := shift + (row << free) + x
							truth[bit>>6] |= 1 << uint(bit&63)
						}
					}
				}
				outputs := [][]uint64{make([]uint64, words), make([]uint64, words), {^uint64(0)}}
				if got := If_CluCountCofs(truth, int32(n), bound, int32(shift), outputs); got != 2 {
					t.Fatalf("count%d", got)
				}
				if !slices.Equal(outputs[0], make([]uint64, words)) || !slices.Equal(outputs[1], expected) || outputs[2][0] != 0xa {
					t.Fatalf("cofactors%v", outputs)
				}
				unique, classes := cluCofactorRows(truth, int32(n), bound, int32(shift))
				if len(unique) != 2 || !slices.Equal(classes, []int{0, 1, 0, 1}) {
					t.Fatalf("classes%v", classes)
				}
				clear(truth)
				if !slices.Equal(unique[1], expected) {
					t.Fatal("unique cofactor borrowed input storage")
				}
				unique[0][0] = 123
				if !slices.Equal(unique[1], expected) {
					t.Fatal("unique cofactors share scratch storage")
				}
			})
		}
	}
}
