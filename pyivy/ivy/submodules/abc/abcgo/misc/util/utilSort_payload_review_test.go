package util

import (
	"fmt"
	"slices"
	"testing"
)

func TestPackedQuickSortSourcePayloadOrder(t *testing.T) {
	for _, which := range []string{"wrapper2", "wrapper3", "recursive2"} {
		for _, decreasing := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/decreasing%t", which, decreasing), func(t *testing.T) {
				data := make([]uint64, 11)
				for i := range data {
					data[i] = uint64(i+1)<<32 | uint64(2-i%2)
				}
				direction := int32(0)
				if decreasing {
					direction = 1
				}
				switch which {
				case "wrapper2":
					Abc_QuickSort2(data, 11, direction)
				case "wrapper3":
					// C's QuickSort3 wrapper deliberately invokes the two-way
					// recursive routines, not QuickSort3Inc/Dec_rec.
					Abc_QuickSort3(data, 11, direction)
				case "recursive2":
					if decreasing {
						Abc_QuickSort2Dec_rec(data, 0, 10)
					} else {
						Abc_QuickSort2Inc_rec(data, 0, 10)
					}
				}
				// Hand trace of source last-pivot partition followed by its
				// strict-comparison selection sorts below the11-word cutoff.
				want := []uint32{10, 2, 4, 8, 6, 9, 11, 5, 3, 1, 7}
				if decreasing {
					want = []uint32{9, 7, 5, 11, 3, 1, 2, 8, 6, 10, 4}
				}
				got := make([]uint32, len(data))
				for i, packed := range data {
					got[i] = uint32(packed >> 32)
				}
				if !slices.Equal(got, want) {
					t.Fatalf("source payload order%v want%v", got, want)
				}
			})
		}
	}
}

func TestPackedThreeWaySourcePayloadOrder(t *testing.T) {
	for _, decreasing := range []bool{false, true} {
		t.Run(fmt.Sprintf("decreasing%t", decreasing), func(t *testing.T) {
			data := []uint64{0xfeed, 0xbeef}
			for i := 1; i <= 11; i++ {
				data = append(data, uint64(i)<<32|1)
			}
			data = append(data, 0xcafe)
			if decreasing {
				Abc_QuickSort3Dec_rec(data, 2, 12)
			} else {
				Abc_QuickSort3Inc_rec(data, 2, 12)
			}
			// Source three-way equal-key moves use k<p and k>q. Preserve
			// those exact boundaries, even though the wrapper uses two-way.
			want := []uint32{6, 9, 8, 7, 10, 11, 4, 3, 2, 1, 5}
			got := make([]uint32, 11)
			for i, packed := range data[2:13] {
				got[i] = uint32(packed >> 32)
			}
			if !slices.Equal(got, want) || data[0] != 0xfeed || data[1] != 0xbeef || data[13] != 0xcafe {
				t.Fatalf("three-way source payload order%v want%v or range guard changed", got, want)
			}
		})
	}
}

func checkPackedSortProperties(t *testing.T, n, pattern int, which string, decreasing bool) {
	t.Helper()
	original := make([]uint64, n)
	keys := make([]uint32, n)
	state := uint32(n + 17)
	for i := range original {
		state = state*1664525 + 1013904223
		var key uint32
		switch pattern {
		case 0:
			key = 7
		case 1:
			key = uint32(i)
		case 2:
			key = uint32(n - i)
		case 3:
			key = uint32(i % 3)
		case 4:
			key = state
		default:
			key = []uint32{0, 0xffffffff, 0x80000000, 0x7fffffff}[i%4]
		}
		keys[i] = key
		original[i] = uint64(0xf0000000+uint32(i+1))<<32 | uint64(key)
	}
	slices.Sort(keys)
	if decreasing {
		slices.Reverse(keys)
	}
	data := append([]uint64{0xfeed, 0xbeef}, original...)
	data = append(data, 0xcafe, 0xface)
	direction := int32(0)
	if decreasing {
		direction = 1
	}
	switch which {
	case "wrapper2":
		Abc_QuickSort2(data[2:], int32(n), direction)
	case "wrapper3":
		Abc_QuickSort3(data[2:], int32(n), direction)
	case "recursive2":
		if decreasing {
			Abc_QuickSort2Dec_rec(data, 2, int32(n+1))
		} else {
			Abc_QuickSort2Inc_rec(data, 2, int32(n+1))
		}
	case "recursive3":
		if decreasing {
			Abc_QuickSort3Dec_rec(data, 2, int32(n+1))
		} else {
			Abc_QuickSort3Inc_rec(data, 2, int32(n+1))
		}
	}
	if data[0] != 0xfeed || data[1] != 0xbeef || data[n+2] != 0xcafe || data[n+3] != 0xface {
		t.Fatal("sort changed words outside its requested range")
	}
	seen := make([]bool, n+1)
	for i, packed := range data[2 : n+2] {
		if uint32(packed) != keys[i] {
			t.Fatalf("key%d got%08x want%08x", i, uint32(packed), keys[i])
		}
		id := uint32(packed>>32) - 0xf0000000
		if id == 0 || id > uint32(n) || seen[id] || packed != original[id-1] {
			t.Fatalf("lost/duplicated/modified payload at%d: %016x", i, packed)
		}
		seen[id] = true
	}
}

func TestPackedSortBoundaryKeysAndPayloadBijection(t *testing.T) {
	for _, n := range []int{0, 1, 2, 9, 10, 11, 12, 31, 32, 33, 129, 513} {
		for pattern := 0; pattern < 6; pattern++ {
			for _, which := range []string{"wrapper2", "wrapper3", "recursive2", "recursive3"} {
				for _, decreasing := range []bool{false, true} {
					t.Run(fmt.Sprintf("size%d/pattern%d/%s/decreasing%t", n, pattern, which, decreasing), func(t *testing.T) {
						checkPackedSortProperties(t, n, pattern, which, decreasing)
					})
				}
			}
		}
	}
}
