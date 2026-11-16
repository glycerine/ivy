package if_

import (
	"fmt"
	"math/bits"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestCellConfigurationPackingBoundaries(t *testing.T) {
	for _, n := range []int{1, 2, 3, 4, 7, 8, 14, 15} {
		p := If_ManStart(&If_Par_t{NLutSize: int32(n), NCutsMax: 4})
		manager := If_DsdManAlloc(int32(n), 0)
		p.pIfDsdMan = manager
		leaves := make([]int32, n)
		for i := range leaves {
			leaves[i] = If_ManCreateCi(p).Id
		}
		p.vTtDsds[n] = vec.Vec_IntAllocArrayCopy([]int32{0, 2}, 2)
		stride := max(6, n)
		p.vTtPerms[n] = vec.Vec_StrAlloc(2 * stride)
		vec.Vec_StrFill(p.vTtPerms[n], 2*stride, 0)
		for _, truthBits := range []int{0, 1, 31, 32, 33, 63, 64, 65, 127, 128, 129} {
			for layout := 0; layout < 2; layout++ {
				for phase := 0; phase < 8; phase++ {
					t.Run(fmt.Sprintf("inputs%d/bits%d/layout%d/phase%d", n, truthBits, layout, phase), func(t *testing.T) {
						cut := &If_Cut_t{nLeaves: uint32(n), pLeaves: leaves, iCutFunc: 2, fCompl: uint32(phase >> 1 & 1)}
						vec.Vec_IntWriteEntry(p.vTtDsds[n], 1, int32(2|phase&1))
						width := bits.Len(uint(n)) + 1
						// One extra word checks padding as well as the exact boundary case.
						size := (truthBits+n*width+1+31)/32 + 1
						configs := vec.Vec_IntAllocArrayCopy([]int32{0, int32(size)}, 2)
						text := vec.Vec_StrAlloc(0)
						defer vec.Vec_IntFree(configs)
						defer vec.Vec_StrFree(text)
						manager.nTtBits = int32(truthBits)
						manager.nConfigWords = int32(1 + (truthBits+63)/64)
						vec.Vec_WrdFree(manager.vConfigs)
						manager.vConfigs = vec.Vec_WrdStart(2 * int(manager.nConfigWords))
						stored := vec.Vec_WrdArray(manager.vConfigs)[manager.nConfigWords:]
						pins := make([]int, n)
						for cellPin := range pins {
							dsdPin := (cellPin + 1) % n
							if layout != 0 {
								dsdPin = n - 1 - cellPin
							}
							stored[0] |= word(dsdPin) << uint(4*cellPin)
							cutPin := (dsdPin + n - 1) % n
							sign := (dsdPin + phase) & 1
							vec.Vec_StrWriteEntry(p.vTtPerms[n], stride+dsdPin, byte(2*cutPin+sign))
							If_CutLeaf(p, cut, int32(cutPin)).iCopy = int32(2*(cutPin+20) + (cutPin+layout)%2)
							pins[cellPin] = 2*cutPin + (sign ^ ((cutPin + layout) % 2))
						}
						var wantText strings.Builder
						wantWords := []int32{2, int32(size)}
						for record := 0; record < 2; record++ {
							binary := make([]byte, 32*size)
							for i := range binary {
								binary[i] = '0'
							}
							for bit := 0; bit < truthBits; bit++ {
								// Mix sparse and adjacent bits, including partial final words.
								on := (bit%7 == 0 || bit%11 == 3) != (record != 0)
								if on {
									stored[1+bit/64] |= 1 << uint(bit%64)
									binary[bit] = '1'
								} else {
									stored[1+bit/64] &^= 1 << uint(bit%64)
								}
							}
							for pin, lit := range pins {
								for b := 0; b < width; b++ {
									if lit>>b&1 != 0 {
										binary[truthBits+pin*width+b] = '1'
									}
								}
							}
							result := int32(200 + record*2 + (phase >> 2 & 1) ^ record)
							outputPhase := (phase & 1) ^ (phase >> 1 & 1) ^ int(result&1)
							binary[truthBits+n*width] = byte('0' + outputPhase)
							// Decode the scalar bit string into independent 32-bit chunks.
							for at := 0; at < len(binary); at += 32 {
								chunk := slices.Clone(binary[at : at+32])
								slices.Reverse(chunk)
								value, err := strconv.ParseUint(string(chunk), 2, 32)
								if err != nil {
									t.Fatal(err)
								}
								wantWords = append(wantWords, int32(value))
							}
							fmt.Fprintf(&wantText, "%d %s  ", result>>1, binary[:truthBits])
							for pin := range pins {
								field := binary[truthBits+pin*width : truthBits+(pin+1)*width]
								fmt.Fprintf(&wantText, "%c %s  ", field[0], field[1:])
							}
							fmt.Fprintf(&wantText, "%c\n", binary[truthBits+n*width])
							before := slices.Clone(stored)
							beforePerm := slices.Clone(vec.Vec_StrArray(p.vTtPerms[n]))
							Gia_ManFromIfGetConfig(configs, p, cut, result, text)
							if !slices.Equal(stored, before) || !slices.Equal(beforePerm, vec.Vec_StrArray(p.vTtPerms[n])) {
								t.Fatal("packing modified borrowed configuration/permutation")
							}
						}
						if !slices.Equal(Vec_IntArray(configs), wantWords) {
							t.Fatalf("packed words: got%v want%v", Vec_IntArray(configs), wantWords)
						}
						if got := string(vec.Vec_StrArray(text)); got != wantText.String() {
							t.Fatalf("text: got%q want%q", got, wantText.String())
						}
					})
				}
			}
		}
		If_ManStop(p)
		If_DsdManFree(manager, 0)
	}
}

func TestCellReconstructionRemainsDisabled(t *testing.T) {
	defer func() {
		if got := recover(); got != "if: programmable-cell reconstruction is disabled in ABC" {
			t.Fatalf("got panic%v, want source disabled assertion", got)
		}
	}()
	// Nil arguments ensure the source assertion precedes every read or write.
	Gia_ManFromIfLogicFindCell(nil, nil, nil, nil, nil, 0, nil, nil, nil, nil, nil, nil)
}
