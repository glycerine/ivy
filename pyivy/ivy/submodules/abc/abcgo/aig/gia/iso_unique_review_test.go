package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestIsoUniqueCountersAgainstKnownCardinalities(t *testing.T) {
	for _, size := range []int{0, 1, 2, 255, 256, 257, 8192} {
		for pattern := 0; pattern < 3; pattern++ {
			t.Run(fmt.Sprintf("size%d/pattern%d", size, pattern), func(t *testing.T) {
				values := make([]int32, size)
				want := size
				for i := range values {
					switch pattern {
					case 0:
						values[i] = 7
					case 1:
						values[i] = int32(i % 127)
					case 2:
						values[i] = int32(uint32(i) * 2654435761) // Odd multiplication is bijective modulo 2^32.
					}
				}
				if pattern == 0 {
					want = min(size, 1)
				} else if pattern == 1 {
					want = min(size, 127)
				}
				for _, counter := range []struct {
					name string
					call func(*Vec_Int_t) int32
				}{
					{"copy/sort", Gia_Iso3Unique}, {"driver hash", iso3HashUniqueCount},
				} {
					input := vec.Vec_IntAllocArrayCopy(values, len(values))
					if got := counter.call(input); got != int32(want) {
						t.Fatalf("%s count%d want%d", counter.name, got, want)
					}
					if !slices.Equal(vec.Vec_IntArray(input), values) {
						t.Fatalf("%s changed input", counter.name)
					}
				}
			})
		}
	}
}
