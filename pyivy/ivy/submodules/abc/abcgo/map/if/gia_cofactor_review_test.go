package if_

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Scalar mux definition, independent of truth-table cofactoring and minimization.
// Modes 1..4 exercise each constant-cofactor branch; mode 0 has two disjoint,
// nonconstant cofactors. Every generated function has full support.
func cofactorReviewValue(assignment, n, selector, mode int) bool {
	var left, right bool
	rank := 0
	for pin := 0; pin < n; pin++ {
		if pin == selector {
			continue
		}
		bit := assignment>>pin&1 != 0
		if mode == 0 && rank%2 != 0 {
			right = right != bit
		} else {
			left = left != bit
		}
		rank++
	}
	switch mode {
	case 1:
		right, left = left, false
	case 2:
		right, left = left, true
	case 3:
		right = false
	case 4:
		right = true
	}
	if assignment>>selector&1 != 0 {
		return right
	}
	return left
}

func TestCofactorReconstructionWideSelectorsAndPhases(t *testing.T) {
	for _, n := range []int{3, 6, 7, 9, 13} {
		for _, selector := range []int{0, n / 2, n - 1} {
			for mode := 0; mode < 5; mode++ {
				for phase := 0; phase < 8; phase++ {
					t.Run(fmt.Sprintf("inputs%d/selector%d/mode%d/phase%d", n, selector, mode, phase), func(t *testing.T) {
						p := If_ManStart(&If_Par_t{NLutSize: int32(n), NCutsMax: 4, FTruth: 1, FUseCofVars: 1})
						defer If_ManStop(p)
						truth := make([]word, 1<<max(0, n-6))
						for x := 0; x < 64*len(truth); x++ {
							if cofactorReviewValue(x, n, selector, mode) {
								truth[x>>6] |= 1 << uint(x&63)
							}
						}
						id := vec.Vec_MemHashInsert(p.vTtMem[n], truth)
						vec.Vec_StrFillExtra(p.vTtVars[n], id+1, 0)
						vec.Vec_StrWriteEntry(p.vTtVars[n], id, byte(selector))
						cut := &If_Cut_t{nLeaves: uint32(n), iCutFunc: int32(id*2 + phase&1), fCompl: uint32(phase >> 1 & 1)}
						dst := Gia_ManStart(64)
						defer Gia_ManStop(dst)
						leaves, temp, mapping := Vec_IntAlloc(n), Vec_IntAlloc(n), Vec_IntAlloc(0)
						records := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
						for _, v := range []*Vec_Int_t{leaves, temp, mapping, records} {
							defer vec.Vec_IntFree(v)
						}
						signedMask := 0
						for pin := 0; pin < n; pin++ {
							lit := Gia_ManAppendCi(dst)
							if phase&4 != 0 && pin%2 == 0 {
								lit ^= 1
								signedMask |= 1 << pin
							}
							Vec_IntPush(leaves, lit)
						}
						before := slices.Clone(Vec_IntArray(leaves))
						result := Gia_ManFromIfLogicCofVars(dst, p, cut, leaves, temp, nil, mapping, records)
						if !slices.Equal(before, Vec_IntArray(leaves)) || !slices.Equal(truth, vec.Vec_MemReadEntry(p.vTtMem[n], id)) {
							t.Fatal("reconstruction modified leaves or stored truth")
						}
						offset := vec.Vec_IntEntry(mapping, int(result>>1))
						data := Vec_IntArray(records)
						if offset <= 0 || int(offset) >= len(data) {
							t.Fatalf("missing root record: %v", data)
						}
						count := data[offset]
						expectedCount, top := int32(2), before[selector]>>1
						if mode == 0 {
							expectedCount, top = 3, result>>1
						}
						if count != expectedCount || data[offset+count+1] != -top {
							t.Fatalf("bad mux record: %v", data[offset:])
						}
						// Every record remains reachable through its LUT's mapping entry; the
						// final mux marker may name the selector rather than the root itself.
						for at := 1; at < len(data); {
							fans := int(data[at])
							if fans < 2 || at+fans+1 >= len(data) {
								t.Fatalf("broken record at%d: %v", at, data)
							}
							if at != int(offset) {
								node := data[at+fans+1]
								if node <= 0 || vec.Vec_IntEntry(mapping, int(node)) != int32(at) {
									t.Fatalf("orphaned LUT record at%d", at)
								}
							}
							at += fans + 2
						}
						gia.Gia_ManAppendCo(dst, result)
						input := make([]bool, n)
						for x := 0; x < 1<<n; x++ {
							for pin := range input {
								input[pin] = x>>pin&1 != 0
							}
							want := cofactorReviewValue(x^signedMask, n, selector, mode) != ((phase&1)^(phase>>1&1) != 0)
							got, err := dst.Evaluate(input)
							if err != nil || len(got) != 1 || got[0] != want {
								t.Fatalf("assignment%d: got%v err%v want%t", x, got, err, want)
							}
						}
					})
				}
			}
		}
	}
}

func TestBiDecompositionWideMasksAndPhases(t *testing.T) {
	for _, n := range []int{4, 6, 7, 8, 11, 12} {
		for layout := 0; layout < 2; layout++ {
			for orMode := 0; orMode < 2; orMode++ {
				for phase := 0; phase < 8; phase++ {
					t.Run(fmt.Sprintf("inputs%d/layout%d/or%d/phase%d", n, layout, orMode, phase), func(t *testing.T) {
						p := If_ManStart(&If_Par_t{NLutSize: int32(2 * ((n + 1) / 2)), NCutsMax: 4, FTruth: 1, FUseAndVars: 1})
						defer If_ManStop(p)
						var masks [2]int32
						for pin := 0; pin < n; pin++ {
							side := pin % 2
							if layout != 0 {
								side = 0
								if pin >= n/2 {
									side = 1
								}
							}
							masks[side] |= 1 << pin
						}
						value := func(x int) bool {
							var parts [2]bool
							for side, mask := range masks {
								for pin := 0; pin < n; pin++ {
									if mask>>pin&1 != 0 {
										parts[side] = parts[side] != (x>>pin&1 != 0)
									}
								}
							}
							if orMode != 0 {
								return parts[0] || parts[1]
							}
							return parts[0] && parts[1]
						}
						truth := make([]word, 1<<max(0, n-6))
						for x := 0; x < 64*len(truth); x++ {
							if value(x) {
								truth[x>>6] |= 1 << uint(x&63)
							}
						}
						id := vec.Vec_MemHashInsert(p.vTtMem[n], truth)
						vec.Vec_IntFillExtra(p.vTtDecs[n], id+1, 0)
						vec.Vec_IntWriteEntry(p.vTtDecs[n], id, masks[0]|masks[1]<<16|int32(orMode)<<30)
						cut := &If_Cut_t{nLeaves: uint32(n), iCutFunc: int32(id*2 + (phase & 1)), fCompl: uint32(phase >> 1 & 1)}
						dst := Gia_ManStart(64)
						defer Gia_ManStop(dst)
						leaves, temp, mapping := Vec_IntAlloc(n), Vec_IntAlloc(n), Vec_IntAlloc(0)
						records := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
						for _, v := range []*Vec_Int_t{leaves, temp, mapping, records} {
							defer vec.Vec_IntFree(v)
						}
						signedMask := 0
						for pin := 0; pin < n; pin++ {
							lit := Gia_ManAppendCi(dst)
							if phase&4 != 0 && pin%2 == 0 {
								lit ^= 1
								signedMask |= 1 << pin
							}
							Vec_IntPush(leaves, lit)
						}
						before := slices.Clone(Vec_IntArray(leaves))
						result := Gia_ManFromIfLogicAndVars(dst, p, cut, leaves, temp, nil, mapping, records)
						if !slices.Equal(before, Vec_IntArray(leaves)) || !slices.Equal(truth, vec.Vec_MemReadEntry(p.vTtMem[n], id)) {
							t.Fatal("reconstruction modified source data")
						}
						data := Vec_IntArray(records)
						offset := int(vec.Vec_IntEntry(mapping, int(result>>1)))
						if offset <= 0 || offset+3 >= len(data) || data[offset] != 2 || data[offset+3] != -(result>>1) {
							t.Fatalf("bad join record: %v", data)
						}
						at := 1
						for side, mask := range masks {
							count := int(data[at])
							var expected []int32
							for pin, lit := range before {
								if mask>>pin&1 != 0 {
									expected = append(expected, lit>>1)
								}
							}
							if count != len(expected) || !slices.Equal(data[at+1:at+count+1], expected) {
								t.Fatalf("side%d inputs: %v want%v", side, data[at:], expected)
							}
							node := data[at+count+1]
							if node != data[offset+1+side] || vec.Vec_IntEntry(mapping, int(node)) != int32(at) {
								t.Fatalf("side%d mapping disconnected: %v", side, data)
							}
							at += count + 2
						}
						if at != offset || offset+4 != len(data) {
							t.Fatalf("unexpected extra records: %v", data)
						}
						gia.Gia_ManAppendCo(dst, result)
						input := make([]bool, n)
						for x := 0; x < 1<<n; x++ {
							for pin := range input {
								input[pin] = x>>pin&1 != 0
							}
							want := value(x^signedMask) != ((phase&1)^(phase>>1&1) != 0)
							got, err := dst.Evaluate(input)
							if err != nil || len(got) != 1 || got[0] != want {
								t.Fatalf("assignment%d: got%v err%v want%t", x, got, err, want)
							}
						}
					})
				}
			}
		}
	}
}
