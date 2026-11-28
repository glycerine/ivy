package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"github.com/glycerine/abcgo/misc/util"
	"slices"
	"strings"
	"testing"
)

func TestCiMapPreservesConstraintClassificationForAllOutputs(t *testing.T) {
	src := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a^1)
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	out := Gia_ManDupDfsCiMap(src, []int32{-1, -1}, nil)
	if out.nConstrs != 1 {
		t.Fatal("CI substitution reclassified a retained constraint as a property")
	}
}

func TestCiMapRejectsInvalidNegativeMarkers(t *testing.T) {
	for _, marker := range []int32{-2, -1 << 31} {
		t.Run(fmt.Sprint(marker), func(t *testing.T) {
			src := Gia_ManStart(1)
			a := Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, a)
			defer func() {
				if recover() == nil {
					t.Fatal("CI map accepted a marker other than -1")
				}
			}()
			Gia_ManDupDfsCiMap(src, []int32{marker}, nil)
		})
	}
}

// Abstract entries: -1 keeps a slot, 0/1 are constants, and 2+2*i+phase
// references CI i. Convert to source object literals only when building a case.
func ciMapReviewCases() [][5]int32 {
	var cases [][5]int32
	for a := int32(-1); a <= 3; a++ {
		for b := int32(-1); b <= 5; b++ {
			cases = append(cases, [5]int32{a, b, -1, -1, -1})
		}
	}
	for phases := 0; phases < 32; phases++ {
		var mapping [5]int32
		mapping[0] = 2 + int32(phases&1)
		for i := 1; i < 5; i++ {
			mapping[i] = 2 + int32(2*(i-1)+(phases>>i)&1)
		}
		cases = append(cases, mapping)
	}
	for constants := 0; constants < 32; constants++ {
		var mapping [5]int32
		for i := range mapping {
			mapping[i] = int32((constants >> i) & 1)
		}
		cases = append(cases, mapping)
	}
	for b := int32(0); b < 2; b++ {
		for c := int32(0); c < 2; c++ {
			for d := int32(0); d < 2; d++ {
				for e := int32(0); e < 4; e++ {
					cases = append(cases, [5]int32{-1, 4 + b, 2 + c, 4 + d, 6 + e})
				}
			}
		}
	}
	return cases
}

func ciMapReviewRecords(t *testing.T, typed, layout bool) string {
	t.Helper()
	var records strings.Builder
	for _, signs := range []int32{0, 1, 2, 4, 8, 16, 21, 31} {
		for ciCase, spec := range ciMapReviewCases() {
			for _, regs := range []int32{0, 2, 4} {
				for mode := 0; mode < 3; mode++ {
					src, _ := levelReviewGraph(signs, 0, typed)
					Gia_ManAppendCi(src)
					Gia_ManSetRegNum(src, regs)
					if layout {
						src = Gia_ManDupOrderDfsReverse(src)
					}
					src.nConstrs = 1
					src.pCexSeq = util.AbcCexSentinel
					mapping := make([]int32, 5)
					resolved := make([]int32, 5)
					for i, entry := range spec {
						switch {
						case entry < 0:
							mapping[i] = -1
							resolved[i] = int32(2 * (i + 1))
						case entry < 2:
							mapping[i] = entry
							resolved[i] = entry
						default:
							reference, phase := (entry-2)/2, (entry-2)%2
							mapping[i] = (Gia_ManCi(src, reference).id << 1) | phase
							base := int32(2 * (i + 1))
							if int(reference) != i {
								base = resolved[reference]
							}
							resolved[i] = base ^ phase
						}
					}
					before := slices.Clone(mapping)
					var selected []int32
					var args *Vec_Int_t
					if mode != 0 {
						if mode == 1 {
							first, _, _, _ := src.Gate(Gia_ManCo(src, 0).id)
							selected = []int32{first ^ 1, first, first ^ 1}
						}
						for i := int32(0); i < regs; i++ {
							lit, _, _, _ := src.Gate(Gia_ManRi(src, i).id)
							selected = append(selected, lit)
						}
						args = permutation(selected...)
					}
					out := Gia_ManDupDfsCiMap(src, mapping, args)
					expectedCo := Gia_ManCoNum(src)
					if mode != 0 {
						expectedCo = int32(len(selected))
					}
					expectedConstraints := int32(0)
					if mode == 0 {
						expectedConstraints = 1
					}
					name, specification := out.Identity()
					if Gia_ManCiNum(out) != 5 || Gia_ManCoNum(out) != expectedCo || out.nRegs != regs || out.nConstrs != expectedConstraints || name != "levels" || specification != "levels.aig" || out.pCexSeq != nil || Gia_ManHashStarted(out) != 0 || out.fGiaSimple != 0 || out.PMuxes != nil {
						t.Fatal("CI-map copy changed interface, metadata or hashing policy")
					}
					if !slices.Equal(mapping, before) || (args != nil && !slices.Equal(Vec_IntArray(args), selected)) {
						t.Fatal("substitution changed its arguments")
					}
					for i := int32(0); i < 5; i++ {
						obj := Gia_ManCi(out, i)
						if obj.id != i+1 || Gia_ObjCioId(obj) != i || Gia_ManCi(src, i).Value != uint32(resolved[i]) {
							t.Fatalf("map %d: wrong slot allocation or composed substitution", ciCase)
						}
					}
					for i := int32(0); i < Gia_ManCoNum(src); i++ {
						if Gia_ManCo(src, i).Value != ^uint32(0) {
							t.Fatal("CI-map API copied source CO object instead of its driver")
						}
					}
					fmt.Fprintf(&records, "%d %d %d %d %d %d %d %d", signs, ciCase, regs, mode, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
					for _, obj := range src.pObjs {
						fmt.Fprintf(&records, " %d", obj.Value)
					}
					for assignment := 0; assignment < 32; assignment++ {
						inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0, assignment&16 != 0}
						effective := make([]bool, 5)
						for i, entry := range spec {
							switch {
							case entry < 0:
								effective[i] = inputs[i]
							case entry < 2:
								effective[i] = entry != 0
							default:
								reference, phase := int((entry-2)/2), (entry-2)%2
								value := inputs[i]
								if reference != i {
									value = effective[reference]
								}
								effective[i] = value != (phase != 0)
							}
						}
						original, _ := selectionReviewTruth(signs, typed, effective)
						want := original
						if mode != 0 {
							want = nil
							if mode == 1 {
								want = []bool{!original[0], original[0], !original[0]}
							}
							want = append(want, original[len(original)-int(regs):]...)
						}
						checkQuantificationOutputs(t, out, inputs, want)
						packed := 0
						for i, value := range want {
							if value {
								packed |= 1 << uint(i)
							}
						}
						fmt.Fprintf(&records, " %d", packed)
					}
					records.WriteByte('\n')
				}
			}
		}
	}
	return records.String()
}

func TestCiMapChainingConstantsAndOutputModes(t *testing.T) {
	if len(ciMapReviewCases()) != 131 {
		t.Fatal("mapping fixture enumeration changed")
	}
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			t.Run(fmt.Sprintf("typed%v_layout%v", typed, layout), func(t *testing.T) { ciMapReviewRecords(t, typed, layout) })
		}
	}
}

func TestCiMapMatchesOriginalSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "ci_map.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C CI mapping: %v: %s", err, native)
	}
	got, want := strings.Split(strings.TrimSpace(ciMapReviewRecords(t, false, false)), "\n"), strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases %d, C cases %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case %d:\nGo %s\nC %s", i, got[i], want[i])
		}
	}
	witness, err := testoracle.Command(t, exe, "constraints").CombinedOutput()
	if err != nil || string(witness) != "1 0\n" {
		t.Fatalf("inherited constraint omission: %v: %s", err, witness)
	}
}

func TestCiMapOldRepresentativesAfterSubstitution(t *testing.T) {
	for _, signs := range []int32{0, 1, 2, 4, 8, 16, 21, 31} {
		for _, regs := range []int32{0, 2, 4} {
			for _, spec := range [][4]int32{{-1, -1, -1, -1}, {3, 2, -1, -1}, {1, 0, -1, -1}} {
				for _, selected := range []bool{false, true} {
					src, unused := dfsReviewRepresentativeGraph(signs, regs)
					src.nConstrs = 1
					mapping := make([]int32, 4)
					for i, entry := range spec {
						mapping[i] = entry
						if entry >= 2 {
							mapping[i] = (Gia_ManCi(src, (entry-2)/2).id << 1) | ((entry - 2) % 2)
						}
					}
					var args *Vec_Int_t
					if selected {
						var lits []int32
						for i := int32(0); i < Gia_ManCoNum(src); i++ {
							lit, _, _, _ := src.Gate(Gia_ManCo(src, i).id)
							lits = append(lits, lit)
						}
						args = permutation(lits...)
					}
					out := Gia_ManDupDfsCiMap(src, mapping, args)
					wantConstraints := int32(1)
					if selected {
						wantConstraints = 0
					}
					if Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != 5 || out.nRegs != regs || out.nConstrs != wantConstraints || Gia_ManAndNum(out) > 1 || src.pObjs[unused].Value != ^uint32(0) {
						t.Fatal("CI map lost representative pruning or interface")
					}
					for mask := 0; mask < 16; mask++ {
						inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
						effective := make([]bool, 4)
						for i, entry := range spec {
							switch {
							case entry < 0:
								effective[i] = inputs[i]
							case entry < 2:
								effective[i] = entry != 0
							default:
								j := int((entry - 2) / 2)
								value := inputs[i]
								if j != i {
									value = effective[j]
								}
								effective[i] = value != ((entry-2)%2 != 0)
							}
						}
						x := (effective[0] != (signs&1 != 0)) && (effective[1] != (signs&2 != 0))
						checkQuantificationOutputs(t, out, inputs, []bool{x != (signs&4 != 0), (!x) != (signs&8 != 0), (!x) != (signs&16 != 0), false, true})
					}
				}
			}
		}
	}
}

func TestCiMapBoundariesAndValidation(t *testing.T) {
	for _, selected := range []bool{false, true} {
		t.Run(fmt.Sprintf("emptySelected%v", selected), func(t *testing.T) {
			src := Gia_ManStart(1)
			var args *Vec_Int_t
			if selected {
				args = permutation()
			}
			out := Gia_ManDupDfsCiMap(src, nil, args)
			if out.nObjs != 1 || out.nRegs != 0 || Gia_ManCoNum(out) != 0 || Gia_ManHashStarted(out) != 0 {
				t.Fatal("empty mapping changed graph or retained hash table")
			}
		})
	}
	src := Gia_ManStart(1)
	src.fGiaSimple = 1
	a := Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, a^1))
	out := Gia_ManDupDfsCiMap(src, []int32{-1}, nil)
	if Gia_ManAndNum(out) != 0 || out.fGiaSimple != 0 {
		t.Fatal("hashed substitution did not normalize a simple-mode contradiction")
	}
	for _, input := range []bool{false, true} {
		checkQuantificationOutputs(t, out, []bool{input}, []bool{false})
	}
	constant := Gia_ManStart(1)
	Gia_ManAppendCo(constant, 1)
	constant.nConstrs = 1
	out = Gia_ManDupDfsCiMap(constant, nil, nil)
	checkQuantificationOutputs(t, out, nil, []bool{true})
	if out.nConstrs != 1 {
		t.Fatal("constant-only constraint lost its classification")
	}
	cases := []struct {
		name    string
		mapping []int32
		lits    []int32
		regs    int32
		useLits bool
	}{
		{"short map", []int32{-1}, nil, 0, false},
		{"long map", []int32{-1, -1, -1}, nil, 0, false},
		{"forward reference", []int32{4, -1}, nil, 0, false},
		{"internal reference", []int32{-1, 6}, nil, 0, false},
		{"out of range reference", []int32{100, -1}, nil, 0, false},
		{"negative selected literal", []int32{-1, -1}, []int32{-1}, 0, true},
		{"large selected literal", []int32{-1, -1}, []int32{1<<31 - 1}, 0, true},
		{"CO selected as logic", []int32{-1, -1}, []int32{8}, 0, true},
		{"too few next-state outputs", []int32{-1, -1}, nil, 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			x := Gia_ManAppendAnd(src, a, b)
			Gia_ManAppendCo(src, x)
			Gia_ManSetRegNum(src, tc.regs)
			var args *Vec_Int_t
			if tc.useLits {
				args = permutation(tc.lits...)
			}
			defer func() {
				if recover() == nil {
					t.Fatal("invalid CI mapping or output selection accepted")
				}
			}()
			Gia_ManDupDfsCiMap(src, tc.mapping, args)
		})
	}
}

func TestCiMapRejectsNonemptyShortStateSelection(t *testing.T) {
	src := Gia_ManStart(1)
	Gia_ManAppendCi(src)
	r0, r1 := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, r0)
	Gia_ManAppendCo(src, r1)
	Gia_ManSetRegNum(src, 2)
	defer func() {
		if recover() == nil {
			t.Fatal("one selected output cannot represent two retained registers")
		}
	}()
	Gia_ManDupDfsCiMap(src, []int32{-1, -1, -1}, permutation(r0))
}
