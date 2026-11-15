package gia

import (
	"fmt"
	"math/bits"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func TestSelectedOutputsLeavesSourceOutputMappingsUnset(t *testing.T) {
	src := Gia_ManStart(1)
	a := Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a^1)
	Gia_ManDupSelectedOutputs(src, permutation(1))
	for i := int32(0); i < 2; i++ {
		if Gia_ManCo(src, i).Value != ^uint32(0) {
			t.Fatal("selected-output copy mapped a source CO where C leaves it unset")
		}
	}
}

type outputSelectionReviewCase struct {
	mode        int
	regs        int32
	first, last int32
	selected    []int32
}

func outputSelectionReviewCases(count int32) []outputSelectionReviewCase {
	var cases []outputSelectionReviewCase
	for _, regs := range []int32{0, 2, 4} {
		for first := int32(0); first <= count; first++ {
			for last := first; last <= count; last++ {
				selected := []int32{}
				for i := first; i < last; i++ {
					selected = append(selected, i)
				}
				cases = append(cases, outputSelectionReviewCase{mode: 0, regs: regs, first: first, last: last, selected: selected})
			}
		}
	}
	for mask := 0; mask < 1<<uint(count); mask++ {
		var selected []int32
		for i := int32(0); i < count; i++ {
			if mask&(1<<uint(i)) != 0 {
				selected = append(selected, i)
			}
		}
		cases = append(cases, outputSelectionReviewCase{mode: 1, selected: slices.Clone(selected)})
		slices.Reverse(selected)
		cases = append(cases, outputSelectionReviewCase{mode: 1, selected: selected})
	}
	for mask := 0; mask < 1<<uint(count); mask++ {
		var selected []int32
		for i := int32(0); i < count; i++ {
			if mask&(1<<uint(i)) != 0 {
				selected = append(selected, i)
			}
		}
		cases = append(cases, outputSelectionReviewCase{mode: 2, selected: selected})
	}
	for _, selected := range [][]int32{{count - 1, 0, 0, 3}, {3, 1, count - 1, 0}, {count - 1, count - 1, count - 1, count - 1}} {
		cases = append(cases, outputSelectionReviewCase{mode: 1, selected: selected})
	}
	return cases
}

func outputSelectionReviewRecords(t *testing.T, typed, dfs bool) string {
	t.Helper()
	var records strings.Builder
	// Fixed logical-node dependency masks, independent of the copying traversal.
	support := []uint{15, 15, 5, 8, 7, 0}
	gates := []uint{7, 56, 8, 0, 3, 0} // x/y/z/u/v/w from levelReviewGraph.
	if typed {
		support = []uint{7, 15, 8, 9, 3}
		gates = []uint{3, 255, 12, 28, 33}
	}
	for signs := int32(0); signs < 32; signs++ {
		for index, tc := range outputSelectionReviewCases(int32(len(support))) {
			src, _ := levelReviewGraph(signs, 0, typed)
			Gia_ManAppendCi(src)
			Gia_ManSetRegNum(src, tc.regs)
			src.nConstrs = 1
			if dfs {
				src = Gia_ManDupOrderDfs(src)
			}
			var out *Gia_Man_t
			switch tc.mode {
			case 0:
				out = Gia_ManDupOutputGroup(src, tc.first, tc.last)
			case 1:
				out = Gia_ManDupSelectedOutputs(src, permutation(tc.selected...))
			case 2:
				present := make([]int32, len(support))
				for _, i := range tc.selected {
					present[i] = 1
					if i%2 == 0 {
						present[i] = -7
					}
				}
				out = Gia_ManDupOutputVec(src, permutation(present...))
			}
			ciMask, gateMask := uint(0), uint(0)
			for _, co := range tc.selected {
				ciMask |= support[co]
				gateMask |= gates[co]
			}
			if tc.mode != 0 {
				ciMask = 31
			} // Explicit/mask selection retains all five CIs.
			name, spec := out.Identity()
			if Gia_ManCiNum(out) != int32(bits.OnesCount(ciMask)) || Gia_ManCoNum(out) != int32(len(tc.selected)) || Gia_ManAndNum(out) != int32(bits.OnesCount(gateMask)) || out.nRegs != 0 || out.nConstrs != 0 || name != "levels" || spec != "levels.aig" {
				t.Fatalf("case %d typed%v dfs%v: wrong cone/interface/identity", index, typed, dfs)
			}
			originalInputOrder := make([]int, Gia_ManCiNum(out))
			for i := int32(0); i < 5; i++ {
				value := Gia_ManCi(src, i).Value
				if ciMask&(1<<uint(i)) == 0 {
					if value != ^uint32(0) {
						t.Fatal("group retained unused input")
					}
					continue
				}
				if value == ^uint32(0) || value&1 != 0 {
					t.Fatal("invalid input mapping")
				}
				obj := Gia_ManObj(out, int32(value>>1))
				if Gia_ObjIsCi(obj) == 0 {
					t.Fatal("source CI mapped to a non-CI")
				}
				ordinal := Gia_ObjCioId(obj)
				if ordinal < 0 || ordinal >= Gia_ManCiNum(out) || Gia_ManCi(out, ordinal) != obj {
					t.Fatal("invalid copied input ordinal")
				}
				originalInputOrder[ordinal] = int(i)
				if tc.mode != 0 && ordinal != i {
					t.Fatal("explicit selection changed original input order")
				}
			}
			for i := int32(0); i < Gia_ManCoNum(src); i++ {
				position := slices.Index(tc.selected, i)
				want := ^uint32(0)
				if tc.mode != 1 && position >= 0 {
					want = uint32(Gia_ManCo(out, int32(position)).id << 1)
				}
				if Gia_ManCo(src, i).Value != want {
					t.Fatal("source CO mapping does not match this API's C contract")
				}
			}
			fmt.Fprintf(&records, "%d %d %d %d %d %d %d", tc.mode, signs, index, tc.regs, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out))
			for i := int32(0); i < 5; i++ {
				fmt.Fprintf(&records, " %d", Gia_ManCi(src, i).Value)
			}
			for i := int32(0); i < Gia_ManCoNum(src); i++ {
				fmt.Fprintf(&records, " %d", Gia_ManCo(src, i).Value)
			}
			for mask := 0; mask < 1<<uint(Gia_ManCiNum(out)); mask++ {
				inputs := make([]bool, Gia_ManCiNum(out))
				original := make([]bool, 5)
				for i, sourceIndex := range originalInputOrder {
					inputs[i] = mask&(1<<uint(i)) != 0
					original[sourceIndex] = inputs[i]
				}
				cos, _ := selectionReviewTruth(signs, typed, original)
				want := make([]bool, len(tc.selected))
				for i, index := range tc.selected {
					want[i] = cos[index]
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
	return records.String()
}

func TestOutputSelectionConesAndMappings(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, dfs := range []bool{false, true} {
			t.Run(fmt.Sprintf("typed%v_dfs%v", typed, dfs), func(t *testing.T) { outputSelectionReviewRecords(t, typed, dfs) })
		}
	}
}

func TestOutputSelectionsMatchOriginalSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "output_selection.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C selections: %v: %s", err, native)
	}
	got := strings.Split(strings.TrimSpace(outputSelectionReviewRecords(t, false, false)), "\n")
	want := strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases %d, C cases %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case %d:\nGo %s\nC %s", i, got[i], want[i])
		}
	}
}

func TestSelectedOutputsRejectsOversizedSelection(t *testing.T) {
	src := Gia_ManStart(1)
	Gia_ManAppendCo(src, 1)
	defer func() {
		if recover() == nil {
			t.Fatal("selected-output vector exceeded C's output-count precondition")
		}
	}()
	Gia_ManDupSelectedOutputs(src, permutation(0, 0))
}

func TestOutputSelectionValidation(t *testing.T) {
	cases := []struct {
		name string
		call func(*Gia_Man_t)
	}{
		{"negative range", func(g *Gia_Man_t) { Gia_ManDupOutputGroup(g, -1, 0) }},
		{"reversed range", func(g *Gia_Man_t) { Gia_ManDupOutputGroup(g, 1, 0) }},
		{"oversized range", func(g *Gia_Man_t) { Gia_ManDupOutputGroup(g, 0, 3) }},
		{"negative selection", func(g *Gia_Man_t) { Gia_ManDupSelectedOutputs(g, permutation(-1)) }},
		{"out of range selection", func(g *Gia_Man_t) { Gia_ManDupSelectedOutputs(g, permutation(2)) }},
		{"short presence", func(g *Gia_Man_t) { Gia_ManDupOutputVec(g, permutation(1)) }},
		{"long presence", func(g *Gia_Man_t) { Gia_ManDupOutputVec(g, permutation(1, 0, 1)) }},
		{"sequential selection", func(g *Gia_Man_t) { Gia_ManSetRegNum(g, 1); Gia_ManDupSelectedOutputs(g, permutation(0)) }},
		{"sequential presence", func(g *Gia_Man_t) { Gia_ManSetRegNum(g, 1); Gia_ManDupOutputVec(g, permutation(1, 0)) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := Gia_ManStart(1)
			a := Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, a)
			Gia_ManAppendCo(src, 1)
			defer func() {
				if recover() == nil {
					t.Fatal("invalid selection accepted")
				}
			}()
			tc.call(src)
		})
	}
}

func TestOutputSelectionBoundaryGraphs(t *testing.T) {
	for _, mode := range []string{"empty", "unused-input", "constants"} {
		t.Run(mode, func(t *testing.T) {
			for method := 0; method < 3; method++ {
				src := Gia_ManStart(1)
				if mode == "unused-input" {
					Gia_ManAppendCi(src)
				}
				if mode == "constants" {
					Gia_ManAppendCo(src, 0)
					Gia_ManAppendCo(src, 1)
				}
				for _, chooseAll := range []bool{false, true} {
					count := Gia_ManCoNum(src)
					var selected []int32
					present := make([]int32, count)
					if chooseAll {
						for i := int32(0); i < count; i++ {
							selected = append(selected, i)
							present[i] = -9
						}
					}
					var out *Gia_Man_t
					switch method {
					case 0:
						out = Gia_ManDupOutputGroup(src, 0, int32(len(selected)))
					case 1:
						out = Gia_ManDupSelectedOutputs(src, permutation(selected...))
					case 2:
						out = Gia_ManDupOutputVec(src, permutation(present...))
					}
					ciCount := Gia_ManCiNum(src)
					if method == 0 {
						ciCount = 0
					}
					if Gia_ManCiNum(out) != ciCount || Gia_ManCoNum(out) != int32(len(selected)) || out.nObjs != 1+ciCount+int32(len(selected)) {
						t.Fatal("boundary selection changed nodes/interface")
					}
					var inputs, want []bool
					if ciCount > 0 {
						inputs = []bool{true}
					}
					if chooseAll && mode == "constants" {
						want = []bool{false, true}
					}
					checkQuantificationOutputs(t, out, inputs, want)
				}
			}
		})
	}
}

func TestOutputSelectionRetainsMuxControlCone(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		for method := 0; method < 3; method++ {
			src := Gia_ManStart(1)
			a, b, c, d := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			control := Gia_ManAppendAnd(src, a^(signs&1), b)
			src.EnableMuxes()
			mux := Gia_ManAppendMuxReal(src, control, c^((signs>>1)&1), d^((signs>>2)&1))
			Gia_ManAppendCo(src, a)
			Gia_ManAppendCo(src, mux)
			Gia_ManAppendCo(src, 1)
			var out *Gia_Man_t
			switch method {
			case 0:
				out = Gia_ManDupOutputGroup(src, 1, 2)
			case 1:
				out = Gia_ManDupSelectedOutputs(src, permutation(1))
			case 2:
				out = Gia_ManDupOutputVec(src, permutation(0, -7, 0))
			}
			if Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != 1 || Gia_ManAndNum(out) != 2 {
				t.Fatal("selection lost real MUX control cone")
			}
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				original := make([]bool, 4)
				for i := int32(0); i < 4; i++ {
					ordinal := Gia_ObjCioId(Gia_ManObj(out, int32(Gia_ManCi(src, i).Value>>1)))
					original[i] = inputs[ordinal]
				}
				a, b, c, d := original[0], original[1], original[2], original[3]
				want := d != (signs&4 != 0)
				if (a != (signs&1 != 0)) && b {
					want = c != (signs&2 != 0)
				}
				checkQuantificationOutputs(t, out, inputs, []bool{want})
			}
		}
	}
}
