package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func TestTrimDualRetainsConstantMismatch(t *testing.T) {
	src := Gia_ManStart(1)
	Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, 0)
	Gia_ManAppendCo(src, 0) // Equal pair can be removed.
	Gia_ManAppendCo(src, 0)
	Gia_ManAppendCo(src, 1) // This mismatch must survive.
	if got := Vec_IntArray(Gia_ManDupTrimmedNonZero(src)); !slices.Equal(got, []int32{1}) {
		t.Errorf("retained pair indices %v, want [1]", got)
	}
	out := Gia_ManDupTrimmed(src, 1, 1, 1, -1)
	if Gia_ManPoNum(out) != 2 {
		t.Fatalf("POs%d, want one retained pair", Gia_ManPoNum(out))
	}
	checkQuantificationOutputs(t, out, []bool{false}, []bool{false, true})
}

func TestTrimmingRetainsSelectedOutputRoles(t *testing.T) {
	for _, mode := range []string{"unchanged", "constant_trim", "pi_trim"} {
		t.Run(mode, func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			x := Gia_ManAppendAnd(src, a, b)
			Gia_ManAppendCo(src, a)
			Gia_ManAppendCo(src, 0)
			Gia_ManAppendCo(src, x)
			Gia_ManAppendCo(src, x^1)
			Gia_ManAppendCo(src, r)
			Gia_ManSetRegNum(src, 1)
			src.nConstrs = 1
			src.aigerBadOutputs = 1
			var out *Gia_Man_t
			switch mode {
			case "unchanged":
				out = Gia_ManDupTrimmed(src, 0, 0, 0, -1)
			case "constant_trim":
				out = Gia_ManDupTrimmed(src, 1, 1, 0, -1)
			case "pi_trim":
				out = Gia_ManDupTrimmed2(src)
			}
			if out.nConstrs != 1 || out.aigerBadOutputs != 1 || out.nRegs != 1 {
				t.Fatalf("lost roles: constraints%d bad%d registers%d", out.nConstrs, out.aigerBadOutputs, out.nRegs)
			}
		})
	}
}

func TestCiMapRetainsBadOutputRole(t *testing.T) {
	src := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, r)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a)
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	src.aigerBadOutputs = 1
	out := Gia_ManDupDfsCiMap(src, []int32{-1, -1}, nil)
	if out.nConstrs != 1 || out.aigerBadOutputs != 1 {
		t.Fatalf("CI map roles: constraints%d bad%d", out.nConstrs, out.aigerBadOutputs)
	}
}

func TestTrimDualRetainsStatefulOutputRoles(t *testing.T) {
	src := Gia_ManStart(1)
	a, b, c, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	for _, lit := range []int32{r, a, b, c, a} {
		Gia_ManAppendCo(src, lit)
	}
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	src.aigerBadOutputs = 1
	out := Gia_ManDupTrimmed(src, 1, 1, 1, -1)
	if out.nConstrs != 1 || out.aigerBadOutputs != 1 {
		t.Fatalf("dual stateful roles: constraints%d bad%d", out.nConstrs, out.aigerBadOutputs)
	}
}

func BenchmarkTrimBufferChain(b *testing.B) {
	src := Gia_ManStart(1)
	a, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	x := Gia_ManAppendAnd(src, a, c)
	for i := 0; i < 256; i++ {
		x = Gia_ManAppendBuf(src, x)
	}
	Gia_ManAppendCo(src, x)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Gia_ManDupTrimmed(src, 1, 1, 0, -1)
	}
}

func trimReviewGraph(signs int, typed, layout bool) *Gia_Man_t {
	src := Gia_ManStart(1)
	inputs := make([]int32, 6)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(src)
	}
	a, b, control, r0, r1 := inputs[0], inputs[1], inputs[3], inputs[4], inputs[5]
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	var x, y, z int32
	if typed {
		x = Gia_ManAppendXorReal(src, a^phase(0), b^phase(1))
		src.EnableMuxes()
		y = Gia_ManAppendMuxReal(src, control^phase(3), x, r0^phase(2))
		z = Gia_ManAppendBuf(src, y^1)
	} else {
		x = Gia_ManAppendAnd(src, a^phase(0), b^phase(1))
		y = Gia_ManAppendAnd(src, x, r0^phase(2))
		z = y
	}
	w := Gia_ManAppendAnd(src, r1, a^phase(5))
	for _, lit := range []int32{0, 1, a ^ phase(0), control ^ phase(3), r0 ^ phase(2), z ^ phase(4), 0, x ^ phase(1), w ^ phase(5), 0, 1, z ^ 1, b, w ^ 1} {
		Gia_ManAppendCo(src, lit)
	}
	Gia_ManSetRegNum(src, 2)
	src.nConstrs = 3
	src.SetIdentity("trim", "trim.aig")
	if layout {
		src = Gia_ManDupOrderDfsReverse(src)
	}
	src.aigerBadOutputs = 3
	return src
}

func trimReviewRecords(t *testing.T, typed, layout bool) string {
	t.Helper()
	var records strings.Builder
	for signs := 0; signs < 64; signs++ {
		for mode := 0; mode < 13; mode++ {
			src := trimReviewGraph(signs, typed, layout)
			trimCis, trimCos, value := int32((mode/6)&1), int32((mode/3)&1), int32(mode%3)-1
			inIndices := []int32{0, 1, 2, 3, 4, 5}
			if mode == 12 {
				inIndices = []int32{0, 1, 4, 5}
				if typed {
					inIndices = []int32{0, 1, 3, 4, 5}
				}
			} else if trimCis != 0 {
				inIndices = []int32{0, 1, 3, 4, 5}
			}
			var poIndices []int32
			constants := map[int32]int32{0: 0, 1: 1, 6: 0, 9: 0, 10: 1}
			for i := int32(0); i < 12; i++ {
				keep := true
				if mode == 12 {
					_, constant := constants[i]
					keep = !constant && i != 2 && i != 3
				} else if v, constant := constants[i]; trimCos != 0 && constant && (value == -1 || value == v) {
					keep = false
				}
				if keep {
					poIndices = append(poIndices, i)
				}
			}
			sourceRefs := make([]int32, src.nObjs)
			for i := range sourceRefs {
				sourceRefs[i] = int32(i + 100)
			}
			src.Refs = slices.Clone(sourceRefs)
			for id, obj := range src.pObjs {
				obj.Value = uint32(id + 999)
				obj.fMark0 = uint8(id & 1)
				obj.fMark1 = uint8((id >> 1) & 1)
			}
			var out *Gia_Man_t
			if mode == 12 {
				out = Gia_ManDupTrimmed2(src)
			} else {
				out = Gia_ManDupTrimmed(src, trimCis, trimCos, 0, value)
			}
			name, spec := out.Identity()
			wantConstr, wantBad := int32(0), int32(0)
			for _, i := range poIndices {
				if i >= 9 {
					wantConstr++
				} else if i >= 6 {
					wantBad++
				}
			}
			gates := int32(3)
			if typed {
				gates++
			}
			if Gia_ManCiNum(out) != int32(len(inIndices)) || Gia_ManPoNum(out) != int32(len(poIndices)) || out.nRegs != 2 || out.nConstrs != wantConstr || out.aigerBadOutputs != wantBad || out.nObjs != 1+int32(len(inIndices))+gates+int32(len(poIndices))+2 || name != "trim" || spec != "trim.aig" {
				t.Fatalf("typed%v layout%v signs%d mode%d: changed retained graph/interface/roles", typed, layout, signs, mode)
			}
			if !slices.Equal(src.Refs, sourceRefs) {
				t.Fatal("trim overwrote source reference cache")
			}
			for id, obj := range src.pObjs {
				if obj.fMark0 != uint8(id&1) || obj.fMark1 != uint8((id>>1)&1) {
					t.Fatal("trim changed source marks")
				}
			}
			for i := int32(0); i < 6; i++ {
				obj := Gia_ManCi(src, i)
				position := slices.Index(inIndices, i)
				if position < 0 {
					if obj.Value != ^uint32(0) {
						t.Fatal("removed CI has copy literal")
					}
				} else if obj.Value != uint32(Gia_ManCi(out, int32(position)).id<<1) {
					t.Fatal("retained PI/RO order changed")
				}
			}
			coIndices := append(slices.Clone(poIndices), 12, 13)
			for i := int32(0); i < 14; i++ {
				obj := Gia_ManCo(src, i)
				position := slices.Index(coIndices, i)
				if position < 0 {
					if obj.Value != ^uint32(0) {
						t.Fatal("removed CO has copy literal")
					}
				} else if obj.Value != uint32(Gia_ManCo(out, int32(position)).id<<1) {
					t.Fatal("retained PO/RI map changed")
				}
			}
			emit := !typed && !layout
			if emit {
				fmt.Fprintf(&records, "%d %d result %d %d %d %d values", signs, mode, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
				for _, obj := range src.pObjs {
					if Gia_ObjIsCo(obj) == 0 {
						fmt.Fprintf(&records, " %d", obj.Value)
					}
				}
				fmt.Fprint(&records, " truth")
			}
			for mask := 0; mask < 1<<uint(len(inIndices)); mask++ {
				original := make([]bool, 6)
				inputs := make([]bool, len(inIndices))
				for i, index := range inIndices {
					inputs[i] = mask&(1<<uint(i)) != 0
					original[index] = inputs[i]
				}
				phase := func(bit uint) bool { return signs&(1<<bit) != 0 }
				a, b, c, r0, r1 := original[0], original[1], original[3], original[4], original[5]
				x := (a != phase(0)) && (b != phase(1))
				if typed {
					x = (a != phase(0)) != (b != phase(1))
				}
				y := x && (r0 != phase(2))
				if typed {
					y = r0 != phase(2)
					if c != phase(3) {
						y = x
					}
				}
				z := y
				if typed {
					z = !z
				}
				w := r1 && (a != phase(5))
				all := []bool{false, true, a != phase(0), c != phase(3), r0 != phase(2), z != phase(4), false, x != phase(1), w != phase(5), false, true, !z, b, !w}
				want := make([]bool, len(coIndices))
				packed := 0
				for i, index := range coIndices {
					want[i] = all[index]
					if want[i] {
						packed |= 1 << uint(i)
					}
				}
				checkQuantificationOutputs(t, out, inputs, want)
				if emit {
					fmt.Fprintf(&records, " %d", packed)
				}
			}
			if emit {
				records.WriteByte('\n')
			}
		}
	}
	return records.String()
}

func TestTrimmingSelectionAndStateMatrix(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			trimReviewRecords(t, typed, layout)
		}
	}
}

func TestTrimmingMatchesOriginalC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "trim.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C trimming: %v: %s", err, native)
	}
	got, want := strings.Split(strings.TrimSpace(trimReviewRecords(t, false, false)), "\n"), strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases%d C cases%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case%d\nGo %s\nC %s", i, got[i], want[i])
		}
	}
	dual, err := exec.Command(exe, "dual").CombinedOutput()
	if err != nil {
		t.Fatalf("C working dual trim: %v: %s", err, dual)
	}
	if expected := trimReviewWorkingDualRecords(t); expected != string(dual) {
		t.Fatalf("working dual trim:\nGo %s\nC %s", expected, dual)
	}
	witness, err := exec.Command(exe, "mismatch").CombinedOutput()
	if err != nil || string(witness) != "0 0 0\n" {
		t.Fatalf("inherited C constant-mismatch witness: %v: %s", err, witness)
	}
}

func trimReviewDualGraph(signs, constant int, typed, layout bool) *Gia_Man_t {
	src := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCi(src)
	r := Gia_ManAppendCi(src)
	p, q := int32(signs&1), int32((signs>>1)&1)
	var x, twin int32
	if typed {
		x = Gia_ManAppendXorReal(src, a, b^q)
		twin = Gia_ManAppendBuf(src, x)
	} else {
		x = Gia_ManAppendAnd(src, a, b^q)
		twin = Gia_ManAppendAnd(src, b^q, a)
	}
	for _, lit := range []int32{a ^ p, a ^ p, a ^ p, a ^ p ^ 1, 0, 0, q, q ^ 1, x, twin, r ^ p, a ^ q, r, 0, r ^ 1, 1, 1, 1} {
		Gia_ManAppendCo(src, lit)
	}
	next := a
	if constant != 0 {
		next = 0
	}
	Gia_ManAppendCo(src, next)
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 2
	if layout {
		src = Gia_ManDupOrderDfsReverse(src)
	}
	src.aigerBadOutputs = 4
	return src
}

func TestDualTrimmingZeroResetTraces(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			for signs := 0; signs < 4; signs++ {
				for constant := 0; constant < 2; constant++ {
					for trimCis := int32(0); trimCis < 2; trimCis++ {
						src := trimReviewDualGraph(signs, constant, typed, layout)
						kept := []int32{1, 3, 5}
						if constant == 0 {
							kept = append(kept, 6, 7)
						}
						if got := Vec_IntArray(Gia_ManDupTrimmedNonZero(src)); !slices.Equal(got, kept) {
							t.Fatalf("typed%v layout%v signs%d constant%d: retained pairs %v want %v", typed, layout, signs, constant, got, kept)
						}
						out := Gia_ManDupTrimmed(src, trimCis, 1, 1, -1)
						wantRegs := int32(1 - constant)
						wantBad := int32(4)
						if constant != 0 {
							wantBad = 0
						}
						if out.nRegs != wantRegs || Gia_ManPoNum(out) != int32(len(kept)*2) || Gia_ManPiNum(out) != 3-trimCis || out.nConstrs != 0 || out.aigerBadOutputs != wantBad {
							t.Fatal("dual trim changed surviving pair/state/role partition")
						}
						// All length-three PI sequences from zero reset. Unused source PIs vary too;
						// a removed constant register need not agree outside the reachable states.
						for sequence := 0; sequence < 512; sequence++ {
							sourceState, targetState := false, false
							for frame := 0; frame < 3; frame++ {
								code := sequence >> uint(frame*3)
								a, b, unused := code&1 != 0, code&2 != 0, code&4 != 0
								original, err := src.Evaluate([]bool{a, b, unused, sourceState})
								if err != nil {
									t.Fatal(err)
								}
								inputs := []bool{a, b}
								if trimCis == 0 {
									inputs = append(inputs, unused)
								}
								if out.nRegs > 0 {
									inputs = append(inputs, targetState)
								}
								want := make([]bool, 0, len(kept)*2+int(out.nRegs))
								for _, pair := range kept {
									want = append(want, original[2*pair], original[2*pair+1])
								}
								if out.nRegs > 0 {
									want = append(want, original[18])
								}
								checkQuantificationOutputs(t, out, inputs, want)
								sourceState = original[18]
								if out.nRegs > 0 {
									targetState = want[len(want)-1]
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestTrimmingDummyInterfacesAndSimpleMode(t *testing.T) {
	for _, kind := range []string{"empty", "state_only", "pi_outputs", "constant_outputs", "simple"} {
		for _, mode := range []string{"trim", "pi_trim", "dual"} {
			t.Run(kind+"_"+mode, func(t *testing.T) {
				if mode == "dual" && kind == "simple" {
					return
				} // Its one PO is intentionally not paired.
				src := Gia_ManStart(1)
				switch kind {
				case "state_only":
					r := Gia_ManAppendCi(src)
					Gia_ManAppendCo(src, r)
					Gia_ManSetRegNum(src, 1)
				case "pi_outputs":
					a := Gia_ManAppendCi(src)
					Gia_ManAppendCo(src, a)
					Gia_ManAppendCo(src, a)
				case "constant_outputs":
					Gia_ManAppendCi(src)
					Gia_ManAppendCo(src, 0)
					Gia_ManAppendCo(src, 1)
				case "simple":
					src.fGiaSimple = 1
					a := Gia_ManAppendCi(src)
					Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, a^1))
				}
				var out *Gia_Man_t
				switch mode {
				case "trim":
					out = Gia_ManDupTrimmed(src, 1, 1, 0, -1)
				case "pi_trim":
					out = Gia_ManDupTrimmed2(src)
				case "dual":
					out = Gia_ManDupTrimmed(src, 1, 1, 1, -1)
				}
				if Gia_ManPiNum(out) != 1 || Gia_ManPoNum(out) < 1 {
					t.Fatal("missing mandatory dummy/retained interface")
				}
				if kind == "simple" && out.fGiaSimple != 1 {
					t.Fatal("simple construction mode lost")
				}
				for mask := 0; mask < 1<<uint(Gia_ManCiNum(out)); mask++ {
					inputs := make([]bool, Gia_ManCiNum(out))
					for i := range inputs {
						inputs[i] = mask&(1<<uint(i)) != 0
					}
					want := make([]bool, Gia_ManCoNum(out))
					switch {
					case kind == "state_only" && mode != "dual":
						want[len(want)-1] = inputs[1]
					case kind == "pi_outputs" && mode == "trim":
						want[0], want[1] = inputs[0], inputs[0]
					case kind == "constant_outputs" && mode == "dual":
						want[1] = true
					}
					checkQuantificationOutputs(t, out, inputs, want)
				}
			})
		}
	}
}

func TestTrimPredicatesAndGuards(t *testing.T) {
	src := Gia_ManStart(1)
	a := Gia_ManAppendCi(src)
	for _, lit := range []int32{0, 1, a, a ^ 1} {
		Gia_ManAppendCo(src, lit)
	}
	for i := int32(0); i < 4; i++ {
		for value := int32(-1); value <= 1; value++ {
			want := int32(0)
			if i < 2 && (value == -1 || i == value) {
				want = 1
			}
			if Gia_ManPoIsToRemove(src, Gia_ManPo(src, i), value) != want {
				t.Fatal("constant selector polarity/wildcard")
			}
		}
	}
	for _, kind := range []string{"non_co", "foreign", "bad_value", "odd_pairs"} {
		t.Run(kind, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid trimming argument accepted")
				}
			}()
			switch kind {
			case "non_co":
				Gia_ManPoIsToRemove(src, Gia_ManCi(src, 0), -1)
			case "foreign":
				other := Gia_ManStart(1)
				Gia_ManAppendCo(other, 0)
				Gia_ManPoIsToRemove(src, Gia_ManPo(other, 0), -1)
			case "bad_value":
				Gia_ManPoIsToRemove(src, Gia_ManPo(src, 0), 2)
			case "odd_pairs":
				odd := Gia_ManStart(1)
				Gia_ManAppendCo(odd, 0)
				Gia_ManDupTrimmedNonZero(odd)
			}
		})
	}
}

func TestCiMapExplicitOutputsKeepDerivedRoles(t *testing.T) {
	src := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	for _, lit := range []int32{r, a, a} {
		Gia_ManAppendCo(src, lit)
	}
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	src.aigerBadOutputs = 1
	out := Gia_ManDupDfsCiMap(src, []int32{-1, -1}, permutation(r, a, a))
	if out.nConstrs != 0 || out.aigerBadOutputs != 0 {
		t.Fatal("explicit outputs inherited original role labels")
	}
}

func trimReviewWorkingDualRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for signs := 0; signs < 4; signs++ {
		for constant := 0; constant < 2; constant++ {
			for trimCis := int32(0); trimCis < 2; trimCis++ {
				src := Gia_ManStart(1)
				a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
				Gia_ManAppendCi(src)
				r := Gia_ManAppendCi(src)
				p, q := int32(signs&1), int32((signs>>1)&1)
				x, twin := Gia_ManAppendAnd(src, a, b), Gia_ManAppendAnd(src, b, a)
				for _, lit := range []int32{a, a, r ^ p, a ^ q, x, twin, r, 0} {
					Gia_ManAppendCo(src, lit)
				}
				next := a
				if constant != 0 {
					next = 0
				}
				Gia_ManAppendCo(src, next)
				Gia_ManSetRegNum(src, 1)
				kept := Gia_ManDupTrimmedNonZero(src)
				out := Gia_ManDupTrimmed(src, trimCis, 1, 1, -1)
				expectedPairs := []int32{1}
				if constant == 0 {
					expectedPairs = append(expectedPairs, 3)
				}
				if !slices.Equal(Vec_IntArray(kept), expectedPairs) || out.nRegs != int32(1-constant) || Gia_ManPiNum(out) != 3-trimCis {
					t.Fatal("working dual selection/state")
				}
				fmt.Fprintf(&records, "%d %d %d pairs", signs, constant, trimCis)
				for _, id := range Vec_IntArray(kept) {
					fmt.Fprintf(&records, " %d", id)
				}
				fmt.Fprintf(&records, " result %d %d %d %d truth", out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
				for mask := 0; mask < 1<<uint(Gia_ManCiNum(out)); mask++ {
					inputs := make([]bool, Gia_ManCiNum(out))
					for i := range inputs {
						inputs[i] = mask&(1<<uint(i)) != 0
					}
					a, r := inputs[0], false
					if out.nRegs > 0 {
						r = inputs[len(inputs)-1]
					}
					want := []bool{r != (p != 0), a != (q != 0)}
					if constant == 0 {
						want = append(want, r, false, a)
					}
					checkQuantificationOutputs(t, out, inputs, want)
					packed := 0
					for i, v := range want {
						if v {
							packed |= 1 << uint(i)
						}
					}
					fmt.Fprintf(&records, " %d", packed)
				}
				records.WriteByte('\n')
			}
		}
	}
	return records.String()
}
