package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"slices"
	"strings"
	"testing"
)

func TestAbstractionRejectsInvalidInverseSentinels(t *testing.T) {
	for _, sentinel := range []int32{-2, -1 << 31} {
		t.Run(fmt.Sprint(sentinel), func(t *testing.T) {
			src := Gia_ManStart(1)
			a := Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, a)
			Gia_ManSetRegNum(src, 1)
			defer func() {
				if recover() == nil {
					t.Fatal("inverse mapping accepted a retained-register marker other than -1")
				}
			}()
			Gia_ManDupAbs(src, permutation(), permutation(sentinel))
		})
	}
}

// Enumerate every ordered subset, since promoted register order is observable
// as pseudo-primary-input order. Retained registers stay in original order.
func abstractionReviewMappings() [][]int32 {
	var all [][]int32
	var visit func([]int32)
	visit = func(prefix []int32) {
		all = append(all, slices.Clone(prefix))
		for reg := int32(0); reg < 3; reg++ {
			if !slices.Contains(prefix, reg) {
				visit(append(slices.Clone(prefix), reg))
			}
		}
	}
	visit(nil)
	return all
}

func abstractionReviewGraph(signs int32, typed bool) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b, r0, r1, r2 := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	var x, z int32
	if typed {
		x = Gia_ManAppendXorReal(g, a^(signs&1), r0^((signs>>1)&1))
	} else {
		x = Gia_ManAppendAnd(g, a^(signs&1), r0^((signs>>1)&1))
	}
	y := Gia_ManAppendAnd(g, b^((signs>>2)&1), r1)
	if typed {
		y = Gia_ManAppendBuf(g, y)
		g.EnableMuxes()
		z = Gia_ManAppendMuxReal(g, r2^((signs>>4)&1), x, y)
	} else {
		z = Gia_ManAppendAnd(g, x, y)
	}
	Gia_ManAppendCo(g, z^((signs>>3)&1))
	u := Gia_ManAppendAnd(g, r2^((signs>>4)&1), a)
	Gia_ManAppendCo(g, u^1)
	Gia_ManAppendCo(g, 1)
	Gia_ManAppendCo(g, r1^1)
	Gia_ManAppendCo(g, x)
	if typed {
		Gia_ManAppendCo(g, Gia_ManAppendXorReal(g, u, b))
	} else {
		Gia_ManAppendCo(g, Gia_ManAppendAnd(g, u, b))
	}
	Gia_ManAppendAnd(g, a, b) // Never in an output or retained next-state cone.
	Gia_ManSetRegNum(g, 3)
	g.SetIdentity("abstract", "abstract.aig")
	return g
}

func abstractionReviewTruth(signs int32, typed bool, inputs []bool) []bool {
	a, b, r0, r1, r2 := inputs[0], inputs[1], inputs[2], inputs[3], inputs[4]
	aa, rr, bb := a != (signs&1 != 0), r0 != (signs&2 != 0), b != (signs&4 != 0)
	x := aa && rr
	if typed {
		x = aa != rr
	}
	y := bb && r1
	z := x && y
	if typed {
		z = y
		if r2 != (signs&16 != 0) {
			z = x
		}
	}
	u := (r2 != (signs&16 != 0)) && a
	next := u && b
	if typed {
		next = u != b
	}
	return []bool{z != (signs&8 != 0), !u, true, !r1, x, next}
}

func abstractionReviewRecords(t *testing.T, typed, dfs bool, constraints int32) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 32; signs++ {
		for index, promoted := range abstractionReviewMappings() {
			inverse := []int32{-1, -1, -1}
			for ppi, reg := range promoted {
				inverse[reg] = int32(ppi)
			}
			inputOrder := []int32{0, 1}
			for _, reg := range promoted {
				inputOrder = append(inputOrder, 2+reg)
			}
			kept := []int32{}
			for reg, ppi := range inverse {
				if ppi < 0 {
					kept = append(kept, int32(reg))
					inputOrder = append(inputOrder, 2+int32(reg))
				}
			}
			src := abstractionReviewGraph(signs, typed)
			src.nConstrs = constraints
			if dfs {
				src = Gia_ManDupOrderDfs(src)
			}
			out := Gia_ManDupAbs(src, permutation(promoted...), permutation(inverse...))
			name, spec := out.Identity()
			gates := int32(4)
			if typed {
				gates = 5
			}
			if slices.Contains(kept, 2) {
				gates++
			}
			if Gia_ManCiNum(out) != 5 || Gia_ManPiNum(out) != 2+int32(len(promoted)) || out.nRegs != int32(len(kept)) || Gia_ManPoNum(out) != 3 || Gia_ManCoNum(out) != 3+int32(len(kept)) || Gia_ManAndNum(out) != gates || out.nConstrs != constraints || name != "abstract" || spec != "abstract.aig" || Gia_ManIsNormalized(out) == 0 {
				t.Fatal("abstraction changed input/state/output roles or kept unused logic")
			}
			for ci, sourceIndex := range inputOrder {
				if Gia_ManCi(src, sourceIndex).Value != uint32(Gia_ManCi(out, int32(ci)).id<<1) {
					t.Fatal("wrong pseudo-input or retained-register ordering")
				}
			}
			for reg, ppi := range inverse {
				mapped := Gia_ManRi(src, int32(reg)).Value
				if ppi >= 0 {
					if mapped != ^uint32(0) {
						t.Fatal("promoted register retained its next-state output")
					}
				} else if mapped != uint32(Gia_ManRi(out, int32(slices.Index(kept, int32(reg)))).id<<1) {
					t.Fatal("wrong retained next-state mapping")
				}
			}
			fmt.Fprintf(&records, "%d %d %d %d %d %d %d", signs, index, out.nObjs, Gia_ManCiNum(out), Gia_ManPiNum(out), Gia_ManCoNum(out), out.nRegs)
			for mask := 0; mask < 32; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
				original := make([]bool, 5)
				for i, index := range inputOrder {
					original[index] = inputs[i]
				}
				cos := abstractionReviewTruth(signs, typed, original)
				want := slices.Clone(cos[:3])
				for _, reg := range kept {
					want = append(want, cos[3+reg])
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

func TestAbstractionAllOrderedRegisterSubsets(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, dfs := range []bool{false, true} {
			for _, constraints := range []int32{0, 1} {
				t.Run(fmt.Sprintf("typed%v_dfs%v_constraints%d", typed, dfs, constraints), func(t *testing.T) { abstractionReviewRecords(t, typed, dfs, constraints) })
			}
		}
	}
}

func TestAbstractionMatchesOriginalSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "register_abstraction.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C abstraction: %v: %s", err, native)
	}
	got := strings.Split(strings.TrimSpace(abstractionReviewRecords(t, false, false, 0)), "\n")
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

func TestAbstractionRejectsInconsistentMaps(t *testing.T) {
	cases := []struct {
		name             string
		forward, inverse []int32
	}{
		{"missing inverse", nil, nil},
		{"short inverse", nil, []int32{-1, -1}},
		{"negative flop", []int32{-1}, []int32{-1, -1, -1}},
		{"out of range flop", []int32{3}, []int32{-1, -1, -1}},
		{"duplicate flop", []int32{0, 0}, []int32{0, -1, -1}},
		{"missing inverse entry", []int32{0}, []int32{-1, -1, -1}},
		{"wrong inverse ordinal", []int32{0}, []int32{1, -1, -1}},
		{"repeated inverse ordinal", []int32{0}, []int32{0, 0, -1}},
		{"out of range inverse", []int32{0}, []int32{0, 3, -1}},
		{"inverse without forward", nil, []int32{0, -1, -1}},
		{"permutation disagreement", []int32{1, 0}, []int32{0, 1, -1}},
		{"too many promoted", []int32{0, 1, 2, 0}, []int32{0, 1, 2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := abstractionReviewGraph(0, false)
			defer func() {
				if recover() == nil {
					t.Fatal("inconsistent register maps accepted")
				}
			}()
			Gia_ManDupAbs(src, permutation(tc.forward...), permutation(tc.inverse...))
		})
	}
}

func TestAbstractionBoundaryGraphs(t *testing.T) {
	for _, mode := range []string{"empty", "constants", "unused-input", "no-primary-interface"} {
		t.Run(mode, func(t *testing.T) {
			mappings := [][]int32{nil}
			if mode == "no-primary-interface" {
				mappings = abstractionReviewMappings()
			}
			for _, promoted := range mappings {
				src := Gia_ManStart(1)
				var inverse []int32
				switch mode {
				case "constants":
					Gia_ManAppendCo(src, 0)
					Gia_ManAppendCo(src, 1)
					src.nConstrs = 1
				case "unused-input":
					Gia_ManAppendCi(src)
				case "no-primary-interface":
					r0, r1, r2 := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
					Gia_ManAppendCo(src, r1^1)
					Gia_ManAppendCo(src, r2)
					Gia_ManAppendCo(src, r0)
					Gia_ManSetRegNum(src, 3)
					inverse = []int32{-1, -1, -1}
					for i, reg := range promoted {
						inverse[reg] = int32(i)
					}
				}
				out := Gia_ManDupAbs(src, permutation(promoted...), permutation(inverse...))
				if out.nRegs != src.nRegs-int32(len(promoted)) || Gia_ManCiNum(out) != Gia_ManCiNum(src) || Gia_ManCoNum(out) != Gia_ManCoNum(src)-int32(len(promoted)) || out.nConstrs != src.nConstrs {
					t.Fatal("boundary abstraction changed interface")
				}
				inputOrder := slices.Clone(promoted)
				for reg, ppi := range inverse {
					if ppi < 0 {
						inputOrder = append(inputOrder, int32(reg))
					}
				}
				rows := 1 << uint(Gia_ManCiNum(out))
				for mask := 0; mask < rows; mask++ {
					inputs := make([]bool, Gia_ManCiNum(out))
					for i := range inputs {
						inputs[i] = mask&(1<<uint(i)) != 0
					}
					var want []bool
					switch mode {
					case "constants":
						want = []bool{false, true}
					case "no-primary-interface":
						original := make([]bool, 3)
						for i, reg := range inputOrder {
							original[reg] = inputs[i]
						}
						next := []bool{!original[1], original[2], original[0]}
						for reg, ppi := range inverse {
							if ppi < 0 {
								want = append(want, next[reg])
							}
						}
					}
					checkQuantificationOutputs(t, out, inputs, want)
				}
			}
		})
	}
}
