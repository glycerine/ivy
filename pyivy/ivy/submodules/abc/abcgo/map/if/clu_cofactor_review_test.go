package if_

import (
	"fmt"
	"slices"
	"testing"
)

func TestCluFourCofactorsOverwriteReusedOutputs(t *testing.T) {
	// Each 2-bit row describes a function of the sole free variable. Source
	// discovery order assigns new class IDs while visiting bound assignments.
	for _, rows := range [][4]uint64{{0, 1, 2, 3}, {3, 1, 0, 2}, {1, 2, 0, 1}, {2, 0, 1, 2}} {
		t.Run(fmt.Sprint(rows), func(t *testing.T) {
			var low uint64
			for i, row := range rows {
				low |= row << uint(2*i)
			}
			truth := []uint64{low | low<<8 | low<<16 | low<<24 | low<<32 | low<<40 | low<<48 | low<<56}
			before := slices.Clone(truth)
			var storage [6][1]uint64
			cofs := make([][]uint64, 6)
			for i := range cofs {
				cofs[i] = storage[i][:]
				cofs[i][0] = ^uint64(0)
			}
			var discovered []uint64
			var expected [6]uint64
			for assignment, row := range rows {
				class := slices.Index(discovered, row)
				if class < 0 {
					class = len(discovered)
					discovered = append(discovered, row)
				}
				if class&1 != 0 {
					expected[4] |= 1 << assignment
				}
				if class&2 != 0 {
					expected[5] |= 1 << assignment
				}
			}
			copy(expected[:4], discovered)
			if len(discovered) == 3 {
				expected[3] = discovered[2]
			}
			for reuse := 0; reuse < 2; reuse++ {
				got := If_CluCountCofs4(truth, 3, 2, cofs)
				if got != int32(len(discovered)) {
					t.Fatalf("count%d want%d", got, len(discovered))
				}
				for i, want := range expected {
					if cofs[i][0] != want {
						t.Errorf("reuse%d output%d got%#x want%#x", reuse, i, cofs[i][0], want)
					}
				}
				if !slices.Equal(before, truth) {
					t.Fatal("counting modified input")
				}
			}
		})
	}
}

func TestCluFourCofactorDimensionsAndDerivation(t *testing.T) {
	for bound := 2; bound <= 6; bound++ {
		for free := 1; free <= 4; free++ {
			for count := 3; count <= 4; count++ {
				t.Run(fmt.Sprintf("bound%d/free%d/classes%d", bound, free, count), func(t *testing.T) {
					n := bound + free
					values := make([]uint64, 4)
					for x := 0; x < 1<<free; x++ {
						parity := false
						for pin := 0; pin < free; pin++ {
							parity = parity != (x>>pin&1 != 0)
						}
						values[1] |= 1 << x
						if parity {
							values[2] |= 1 << x
						} else {
							values[3] |= 1 << x
						}
					}
					truth := make([]uint64, 1<<max(0, n-6))
					var seen []uint64
					var expectedSelectors [2]uint64
					for row := 0; row < 1<<bound; row++ {
						value := values[(row^(row>>1))%count]
						class := slices.Index(seen, value)
						if class < 0 {
							class = len(seen)
							seen = append(seen, value)
						}
						for output := 0; output < 2; output++ {
							if class>>output&1 != 0 {
								expectedSelectors[output] |= 1 << row
							}
						}
						for x := 0; x < 1<<free; x++ {
							if value>>x&1 != 0 {
								at := row<<free | x
								truth[at>>6] |= 1 << uint(at&63)
							}
						}
					}
					// Canonical source truth tables stretch small functions to one word.
					if n < 6 {
						low := truth[0]
						for at := 1 << n; at < 64; at++ {
							truth[0] |= (low >> uint(at%(1<<n)) & 1) << uint(at)
						}
					}
					before := slices.Clone(truth)
					var storage [6][1]uint64
					cofs := make([][]uint64, 6)
					for i := range cofs {
						cofs[i] = storage[i][:]
						cofs[i][0] = ^uint64(0)
					}
					if got := If_CluCountCofs4(truth, int32(n), int32(bound), cofs); got != int32(count) {
						t.Fatalf("count%d want%d", got, count)
					}
					for i := 0; i < 4; i++ {
						if want := seen[min(i, count-1)]; cofs[i][0] != want {
							t.Fatalf("cofactor%d got%#x want%#x", i, cofs[i][0], want)
						}
					}
					for i, want := range expectedSelectors {
						if got := cofs[i+4][0]; got != want {
							t.Fatalf("selector%d got%#x want%#x", i, got, want)
						}
					}
					work := slices.Clone(truth)
					v2p, p2v := cluIdentity(int32(n))
					g := If_Grp_t{nVars: int32(bound), nMyu: int32(count)}
					for i := 0; i < bound; i++ {
						g.pVars[i] = int32(free + i)
					}
					var root If_Grp_t
					var first, second uint64
					If_CluDeriveDisjoint4(work, int32(n), v2p, p2v, &g, &root, &first, &second)
					if root.nVars != int32(free+2) || root.pVars[free] != int32(n) || root.pVars[free+1] != int32(n+1) {
						t.Fatalf("root%+v", root)
					}
					for pin := 0; pin < free; pin++ {
						if root.pVars[pin] != int32(pin) {
							t.Fatalf("root input%d", pin)
						}
					}
					for row := 0; row < 1<<bound; row++ {
						low := first >> row & 1
						high := second >> row & 1
						for x := 0; x < 1<<free; x++ {
							index := x | int(low)<<free | int(high)<<(free+1)
							actual := work[index>>6] >> uint(index&63) & 1
							want := before[(row<<free|x)>>6] >> uint((row<<free|x)&63) & 1
							if actual != want {
								t.Fatalf("row%d assignment%d got%d want%d", row, x, actual, want)
							}
						}
					}
					if !slices.Equal(truth, before) {
						t.Fatal("source truth changed")
					}
				})
			}
		}
	}
}
