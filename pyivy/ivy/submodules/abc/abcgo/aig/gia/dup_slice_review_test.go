package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func sliceReviewGraph(signs, registers int32) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a^(signs&1), b^((signs>>1)&1))
	Gia_ManAppendCo(g, x^1)
	z := Gia_ManAppendAnd(g, x, c)
	y := Gia_ManAppendAnd(g, c^((signs>>2)&1), d^((signs>>3)&1))
	r := Gia_ManAppendAnd(g, z, y)
	Gia_ManAppendAnd(g, a, c) // Copied when small, but never exposed.
	for _, lit := range []int32{r ^ 1, a, z ^ 1, 1} {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, registers)
	g.SetIdentity("slice", "original.aig")
	return g
}

func sliceReviewRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 16; signs++ {
		for _, registers := range []int32{0, 2, 4} {
			for limit := int32(-1); limit <= 5; limit++ {
				src := sliceReviewGraph(signs, registers)
				for _, obj := range src.pObjs {
					obj.fMark0, obj.fMark1 = uint8((obj.id+1)&1), 1
				}
				out := Gia_ManDupSliced(src, limit)
				copied, outputs := int32(0), int32(0)
				switch {
				case limit == 2:
					copied, outputs = 3, 2
				case limit == 3:
					copied, outputs = 4, 3
				case limit >= 4:
					copied, outputs = 5, 3
				}
				if Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != outputs || out.nRegs != 0 || out.nObjs != 5+copied+outputs {
					t.Fatalf("limit %d: wrong copied/exposed nodes or combinational interface", limit)
				}
				marks := 0
				for _, obj := range src.pObjs {
					if obj.fMark0 != 0 || obj.fMark1 != 0 {
						marks++
					}
				}
				name, spec := out.Identity()
				if marks != 0 || name != "slice" || spec != "" {
					t.Fatal("slice changed source mark/name/spec conventions")
				}
				fmt.Fprintf(&records, "%d %d %d %d %d %d %d %d %d %d", signs, registers, limit, out.nObjs,
					Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs, marks, giaBool(name == "slice"), giaBool(spec != ""))
				for mask := 0; mask < 16; mask++ {
					a, b, c, d := mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0
					x := (a != (signs&1 != 0)) && (b != (signs&2 != 0))
					z := x && c
					y := (c != (signs&4 != 0)) && (d != (signs&8 != 0))
					var want []bool
					if limit == 2 {
						want = []bool{x, y}
					} else if limit == 3 {
						want = []bool{x, z, y}
					} else if limit >= 4 {
						want = []bool{x, z, z && y}
					}
					inputs := []bool{a, b, c, d}
					checkQuantificationOutputs(t, out, inputs, want)
					values, err := out.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					packed := 0
					for i, value := range values {
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
	return records.String()
}

func TestSlicedSelectionAndMetadata(t *testing.T) {
	sliceReviewRecords(t)
}

func TestSlicedMatchesOriginalSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "slicing.c")
	want, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C slicing: %v: %s", err, want)
	}
	gotRows := strings.Split(strings.TrimSpace(sliceReviewRecords(t)), "\n")
	wantRows := strings.Split(strings.TrimSpace(string(want)), "\n")
	if len(gotRows) != len(wantRows) {
		t.Fatalf("Go cases %d, C cases %d", len(gotRows), len(wantRows))
	}
	for i := range gotRows {
		if gotRows[i] != wantRows[i] {
			t.Fatalf("case %d:\nGo %s\nC  %s", i, gotRows[i], wantRows[i])
		}
	}
}

func TestSlicedExposesRealMuxControlCone(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		src := Gia_ManStart(1)
		a, b, c, d := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
		control := Gia_ManAppendAnd(src, a^(signs&1), b)
		data := Gia_ManAppendAnd(src, c, d^((signs>>1)&1))
		src.EnableMuxes()
		Gia_ManAppendCo(src, Gia_ManAppendMuxReal(src, control^((signs>>2)&1), data, a))
		out := Gia_ManDupSliced(src, 2)
		if Gia_ManCoNum(out) != 2 {
			t.Fatal("slice omitted the control cone of an excluded MUX")
		}
		for mask := 0; mask < 16; mask++ {
			a, b, c, d := mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0
			checkQuantificationOutputs(t, out, []bool{a, b, c, d}, []bool{(a != (signs&1 != 0)) && b, c && (d != (signs&2 != 0))})
		}
	}
}

func TestSlicedTypedGatesAndEmptyGraphs(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		for limit := int32(0); limit <= 6; limit++ {
			src := blockReviewGraph(signs, 3)
			var selected []int32
			for _, obj := range src.pObjs {
				_, _, _, kind := src.Gate(obj.id)
				support := int32(0)
				switch kind {
				case 3, 5:
					support = 2
				case 4:
					support = 4
				case 2:
					support = 5
				default:
					continue
				}
				if support <= limit {
					selected = append(selected, obj.id)
				}
			}
			out := Gia_ManDupSliced(src, limit)
			if Gia_ManCiNum(out) != 6 || Gia_ManCoNum(out) != int32(len(selected)) || out.nRegs != 0 {
				t.Fatal("typed slice changed support threshold/interface")
			}
			for mask := 0; mask < 64; mask++ {
				inputs := make([]bool, 6)
				for i := range inputs {
					inputs[i] = mask&(1<<uint(i)) != 0
				}
				values, _, err := src.evaluateObjects(inputs)
				if err != nil {
					t.Fatal(err)
				}
				var want []bool
				for _, id := range selected {
					want = append(want, values[id])
				}
				checkQuantificationOutputs(t, out, inputs, want)
			}
		}
	}
	empty := Gia_ManStart(1)
	empty.pObjs[0].fMark0, empty.pObjs[0].fMark1 = 1, 1
	out := Gia_ManDupSliced(empty, 0)
	if out.nObjs != 1 || empty.pObjs[0].fMark0 != 0 || empty.pObjs[0].fMark1 != 0 {
		t.Fatal("empty slice changed graph or retained marks")
	}
	constant := Gia_ManStart(1)
	Gia_ManAppendCo(constant, Gia_ManAppendBuf(constant, 1))
	checkQuantificationOutputs(t, Gia_ManDupSliced(constant, 0), nil, []bool{true})
	simple := Gia_ManStart(1)
	simple.fGiaSimple = 1
	a := Gia_ManAppendCi(simple)
	Gia_ManAppendCo(simple, Gia_ManAppendAnd(simple, a, a^1))
	out = Gia_ManDupSliced(simple, 1)
	if out.fGiaSimple != 1 || out.nObjs != 4 {
		t.Fatal("slice changed simple-mode AND storage")
	}
	for _, value := range []bool{false, true} {
		checkQuantificationOutputs(t, out, []bool{value}, []bool{false})
	}
}

func TestSlicedDropsOriginalSpecification(t *testing.T) {
	src := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, b))
	src.SetIdentity("slice", "original.aig")
	out := Gia_ManDupSliced(src, 2)
	if name, spec := out.Identity(); name != "slice" || spec != "" {
		t.Fatalf("slice identity %q/%q, want name only as in C", name, spec)
	}
}

func TestSlicedClearsSourceScratchMarks(t *testing.T) {
	src := Gia_ManStart(1)
	a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	x := Gia_ManAppendAnd(src, a, b)
	Gia_ManAppendCo(src, Gia_ManAppendAnd(src, x, c))
	for _, obj := range src.pObjs {
		obj.fMark0, obj.fMark1 = 1, 1
	}
	out := Gia_ManDupSliced(src, 2)
	checkQuantificationOutputs(t, out, []bool{true, true, false}, []bool{true})
	for _, obj := range src.pObjs {
		if obj.fMark0 != 0 || obj.fMark1 != 0 {
			t.Fatalf("source object %d retained slice scratch marks", obj.id)
		}
	}
}

var sliceReviewSink *Gia_Man_t

func BenchmarkSlicedScratchStorage(b *testing.B) {
	src := Gia_ManStart(1)
	a, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	var root int32
	for i := 0; i < 256; i++ {
		root = Gia_ManAppendAnd(src, a, c)
	}
	Gia_ManAppendCo(src, root)
	Gia_ManSuppSizeOne(src, Gia_ObjFromLit(src, root))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sliceReviewSink = Gia_ManDupSliced(src, 2)
	}
}
