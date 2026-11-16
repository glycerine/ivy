package util

import (
	"fmt"
	"slices"
	"testing"
)

func TestCexPIPermutationKeepsInitialStateAndFrames(t *testing.T) {
	permutations := [][]int32{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	for _, registers := range []int32{0, 1, 31, 32, 33, 67} {
		for _, frames := range []int32{3, 11} {
			for index, mapping := range permutations {
				t.Run(fmt.Sprintf("registers%d/frames%d/permutation%d", registers, frames, index), func(t *testing.T) {
					source := Abc_CexAlloc(registers, 3, frames)
					source.IPo, source.IFrame = 2, frames-1
					bits := make([]bool, source.NBits)
					for bit := range bits {
						bits[bit] = bit%3 != 1
						if bits[bit] {
							source.PData[bit/32] |= uint32(1) << uint(bit%32)
						}
					}
					before := slices.Clone(source.PData)
					want := slices.Clone(bits)
					for frame := int32(0); frame < frames; frame++ {
						for old, new := range mapping {
							want[registers+3*frame+new] = bits[registers+3*frame+int32(old)]
						}
					}
					for _, strategy := range []string{"direct", "two canonical orders"} {
						var got *Abc_Cex_t
						if strategy == "direct" {
							got = Abc_CexPermute(source, mapping)
						} else {
							// A nonidentity old order checks that the wrapper composes mappings
							// rather than treating the new order as an old-to-new map.
							old := []int32{2, 0, 1}
							next := []int32{mapping[2], mapping[0], mapping[1]}
							got = Abc_CexPermuteTwo(source, old, next)
						}
						if got == source || got.NRegs != registers || got.NPis != 3 || got.NBits != source.NBits || got.IPo != 2 || got.IFrame != frames-1 {
							t.Fatalf("%s changed witness metadata or ownership", strategy)
						}
						for bit, expected := range want {
							actual := got.PData[bit/32]&(uint32(1)<<uint(bit%32)) != 0
							if actual != expected {
								t.Fatalf("%s bit%d got%t want%t", strategy, bit, actual, expected)
							}
						}
						if !slices.Equal(source.PData, before) {
							t.Fatal("permutation mutated its source")
						}
						got.PData[0] ^= 1
						if !slices.Equal(source.PData, before) {
							t.Fatal("permutation aliases its source")
						}
					}
				})
			}
		}
	}
}
