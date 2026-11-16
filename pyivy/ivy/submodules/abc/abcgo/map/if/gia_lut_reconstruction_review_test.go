package if_

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// A signed conjunction has complete support and a known two-LUT realization.
// Encode it directly, independently of the DSD/SAT/reconstruction routines.
func lutReconstructionReviewTruth(n int) []word {
	truth := make([]word, 1<<uint(max(0, n-6)))
	for assignment := 0; assignment < len(truth)*64; assignment++ {
		yes := true
		for pin := 0; pin < n; pin++ {
			yes = yes && (assignment>>pin&1) != (pin&1)
		}
		if yes {
			truth[assignment>>6] |= 1 << uint(assignment&63)
		}
	}
	return truth
}

func lutReconstructionReviewCheckPacking(t *testing.T, dst *Gia_Man_t, root, size int32, n int, mapping, records, packing *Vec_Int_t) {
	t.Helper()
	count := 0
	if n >= 2 {
		count = 1
	}
	if n > int(size) {
		count = 2
	}
	if count == 0 {
		if !slices.Equal(vec.Vec_IntArray(packing), []int32{0}) || vec.Vec_IntSize(records) != 1 {
			t.Fatal("constant/variable result added mapping or packing records")
		}
		return
	}
	packed := vec.Vec_IntArray(packing)
	if len(packed) != count+2 || packed[0] != 1 || packed[1] != int32(count) || packed[len(packed)-1] != root>>1 {
		t.Fatalf("packing %v, want one %d-LUT group ending at%d", packed, count, root>>1)
	}
	for _, id := range packed[2:] {
		if id <= 0 || id >= gia.Gia_ManObjNum(dst) {
			t.Fatalf("invalid packed ID%d", id)
		}
		offset := vec.Vec_IntEntry(mapping, int(id))
		data := vec.Vec_IntArray(records)
		if offset <= 0 || int(offset) >= len(data) {
			t.Fatalf("missing record for packed LUT%d", id)
		}
		fans := data[offset]
		if fans < 2 || fans > size || int(offset+fans+1) >= len(data) || data[offset+fans+1] != id {
			t.Fatalf("invalid LUT%d record: %v", id, data[offset:])
		}
		for _, fanin := range data[offset+1 : offset+fans+1] {
			if fanin < 0 || fanin >= id {
				t.Fatalf("LUT%d has non-topological input%d", id, fanin)
			}
		}
	}
}

func TestLutReconstructionSizesPhasesAndPartitions(t *testing.T) {
	for size := int32(2); size <= 6; size++ {
		for n := 0; n <= int(2*size-1); n++ {
			t.Run(fmt.Sprintf("LUT%d/inputs%d", size, n), func(t *testing.T) {
				pars := &If_Par_t{NLutSize: int32(max(1, n)), NCutsMax: 4, FTruth: 1, FUseDsd: 1, PLutStruct: fmt.Sprintf("%d%d", size, size)}
				p := If_ManStart(pars)
				defer If_ManStop(p)
				ids := make([]int32, n)
				for i := range ids {
					ids[i] = If_ManCreateCi(p).Id
				}
				dsd := If_DsdManAlloc(int32(max(1, n)), size)
				defer If_DsdManFree(dsd, 0)
				p.pIfDsdMan = dsd
				truth := lutReconstructionReviewTruth(n)
				truthID := vec.Vec_MemHashInsert(p.vTtMem[n], truth)
				perm := make([]byte, n)
				literal := If_DsdManCompute(dsd, truth, int32(n), perm, "")
				vec.Vec_IntFillExtra(p.vTtDsds[n], truthID+1, -1)
				vec.Vec_IntWriteEntry(p.vTtDsds[n], truthID, literal)
				stride := max(6, n)
				vec.Vec_StrFillExtra(p.vTtPerms[n], (truthID+1)*stride, IF_BIG_CHAR)
				for i, pin := range perm {
					vec.Vec_StrWriteEntry(p.vTtPerms[n], truthID*stride+i, pin)
				}
				var sat any
				if n > int(size) {
					sat = If_ManSatBuildXY(size)
					defer If_ManSatUnbuild(sat)
				}
				maxMode := int32(1)
				if n > int(size) {
					maxMode += max(0, 2*size-1-int32(n))
				}
				for selected := int32(0); selected <= maxMode; selected++ {
					for phase := int32(0); phase <= 1; phase++ {
						for leafPhase := 0; leafPhase <= 1; leafPhase++ {
							t.Run(fmt.Sprintf("selected%d/phase%d/leaves%d", selected, phase, leafPhase), func(t *testing.T) {
								pars.FDelayOptLut = 0
								if selected != 0 {
									pars.FDelayOptLut = 1
								}
								cut := &If_Cut_t{nLeaves: uint32(n), pLeaves: ids, iCutFunc: int32(2*truthID) ^ phase}
								// In DSD coordinates, put the first K variables in the bound
								// LUT. Modes2+ share one or more of them with the root LUT.
								shared := max(0, selected-1)
								for i := int32(0); i < size; i++ {
									class := int32(1)
									if i < shared {
										class = 3
									}
									cut.uMaskFunc |= class << uint(2*i)
								}
								dst := Gia_ManStart(64)
								defer Gia_ManStop(dst)
								leaves := vec.Vec_IntAlloc(n)
								defer vec.Vec_IntFree(leaves)
								for i := 0; i < n; i++ {
									lit := Gia_ManAppendCi(dst)
									if leafPhase != 0 && i%2 == 0 {
										lit ^= 1
									}
									vec.Vec_IntPush(leaves, lit)
								}
								scratch, mapping := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
								records, packing := vec.Vec_IntAllocArrayCopy([]int32{0}, 1), vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
								defer vec.Vec_IntFree(scratch)
								defer vec.Vec_IntFree(mapping)
								defer vec.Vec_IntFree(records)
								defer vec.Vec_IntFree(packing)
								before := slices.Clone(vec.Vec_IntArray(leaves))
								root := Gia_ManFromIfLogicFindLut(p, dst, cut, sat, leaves, scratch, nil, mapping, records, packing)
								if !slices.Equal(vec.Vec_IntArray(leaves), before) {
									t.Fatal("reconstruction changed supplied leaf literals")
								}
								lutReconstructionReviewCheckPacking(t, dst, root, size, n, mapping, records, packing)
								gia.Gia_ManAppendCo(dst, root)
								for assignment := 0; assignment < 1<<uint(n); assignment++ {
									input := make([]bool, n)
									want := true
									for pin := range input {
										input[pin] = assignment>>pin&1 != 0
										logical := input[pin]
										if leafPhase != 0 && pin%2 == 0 {
											logical = !logical
										}
										want = want && (logical != (pin%2 != 0))
									}
									if phase != 0 {
										want = !want
									}
									got, err := dst.Evaluate(input)
									if err != nil || !slices.Equal(got, []bool{want}) {
										t.Fatalf("assignment%d got%v want%t err%v", assignment, got, want, err)
									}
								}
							})
						}
					}
				}
			})
		}
	}
}

func TestLutCreationMultiwordTruthAndSignedLeaves(t *testing.T) {
	for _, n := range []int{0, 1, 2, 5, 6, 7, 8, 11, 15} {
		for phase := 0; phase < 2; phase++ {
			for leafPhase := 0; leafPhase < 2; leafPhase++ {
				t.Run(fmt.Sprintf("inputs%d/phase%d/leaves%d", n, phase, leafPhase), func(t *testing.T) {
					dst := Gia_ManStart(64)
					defer Gia_ManStop(dst)
					leaves, mapping := vec.Vec_IntAlloc(n), vec.Vec_IntAlloc(0)
					records := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
					defer vec.Vec_IntFree(leaves)
					defer vec.Vec_IntFree(mapping)
					defer vec.Vec_IntFree(records)
					for pin := 0; pin < n; pin++ {
						lit := Gia_ManAppendCi(dst)
						if leafPhase != 0 && pin%2 == 0 {
							lit ^= 1
						}
						vec.Vec_IntPush(leaves, lit)
					}
					truth := lutReconstructionReviewTruth(n)
					if phase != 0 {
						for i := range truth {
							truth[i] = ^truth[i]
						}
					}
					before := slices.Clone(truth)
					root := Gia_ManFromIfLogicCreateLut(dst, truth, leaves, nil, mapping, records)
					if !slices.Equal(truth, before) {
						t.Fatal("LUT creation modified caller truth")
					}
					if n < 2 {
						if vec.Vec_IntSize(records) != 1 {
							t.Fatal("constant/variable result created LUT record")
						}
					} else {
						offset := vec.Vec_IntEntry(mapping, int(root>>1))
						data := vec.Vec_IntArray(records)
						if offset != 1 || len(data) != n+3 || data[1] != int32(n) || data[len(data)-1] != root>>1 {
							t.Fatalf("LUT record %v offset%d", data, offset)
						}
						for pin := 0; pin < n; pin++ {
							if data[pin+2] != int32(pin+1) {
								t.Fatalf("LUT input order changed: %v", data)
							}
						}
					}
					Gia_ManAppendCo(dst, root)
					for assignment := 0; assignment < 1<<uint(n); assignment++ {
						input := make([]bool, n)
						want := true
						for pin := range input {
							input[pin] = assignment>>pin&1 != 0
							logical := input[pin]
							if leafPhase != 0 && pin%2 == 0 {
								logical = !logical
							}
							want = want && (logical != (pin%2 != 0))
						}
						if phase != 0 {
							want = !want
						}
						got, err := dst.Evaluate(input)
						if err != nil || !slices.Equal(got, []bool{want}) {
							t.Fatalf("assignment%d got%v want%t err%v", assignment, got, want, err)
						}
					}
				})
			}
		}
	}
}
