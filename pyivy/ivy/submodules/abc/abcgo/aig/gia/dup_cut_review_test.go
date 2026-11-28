package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"github.com/glycerine/abcgo/misc/util"
	"strings"
	"testing"
)

func TestReplaceCutRetainsUncutOutputLogic(t *testing.T) {
	src := Gia_ManStart(1)
	a, b, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, b)
	Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, b))
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	out := Gia_ManDupReplaceCut(src)
	if Gia_ManCiNum(out) != 3 || Gia_ManAndNum(out) != 1 {
		t.Fatal("uncut PI-only output became an independent free input")
	}
	for mask := 0; mask < 8; mask++ {
		a, b, r := mask&1 != 0, mask&2 != 0, mask&4 != 0
		checkQuantificationOutputs(t, out, []bool{a, b, r}, []bool{a, b, a && b, r})
	}
}

func TestReplaceCutPreservesConstraintRoles(t *testing.T) {
	src := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a^1)
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	if out := Gia_ManDupReplaceCut(src); out.nConstrs != 1 {
		t.Fatal("cut replacement reclassified an unchanged constraint output")
	}
}

func TestReplaceCutPreservesSimpleMode(t *testing.T) {
	src := Gia_ManStart(1)
	src.fGiaSimple = 1
	r := Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, Gia_ManAppendAnd(src, r, r^1))
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	out := Gia_ManDupReplaceCut(src)
	if out.fGiaSimple != 1 {
		t.Fatal("cut replacement lost simple construction mode")
	}
	for _, r := range []bool{false, true} {
		checkQuantificationOutputs(t, out, []bool{r}, []bool{false, r})
	}
}

func BenchmarkHighLightFlopLogic(b *testing.B) {
	src := Gia_ManStart(1)
	a, c, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	root := Gia_ManAppendAnd(src, Gia_ManAppendAnd(src, a, c), r)
	for i := 0; i < 256; i++ {
		root = Gia_ManAppendAnd(src, root, r^int32(i&1))
	}
	Gia_ManAppendCo(src, root)
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Gia_ManHighLightFlopLogic(src)
	}
}

func cutReviewGraph(signs, mode int, typed, layout bool) (*Gia_Man_t, []int32, map[int32]uint32) {
	g := Gia_ManStart(1)
	a, b, c, r0, r1 := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	phase := func(bit int) int32 { return int32((signs >> bit) & 1) }
	var x, y, z, w int32
	if typed {
		x = Gia_ManAppendXorReal(g, a^phase(0), b^phase(1))
	} else {
		x = Gia_ManAppendAnd(g, a^phase(0), b^phase(1))
	}
	y = Gia_ManAppendAnd(g, b, c^phase(2))
	if typed {
		y = Gia_ManAppendBuf(g, y)
		g.EnableMuxes()
	}
	if typed {
		z = Gia_ManAppendMuxReal(g, r0^phase(3), x, y)
		w = Gia_ManAppendXorReal(g, y, r1^phase(4))
	} else {
		z = Gia_ManAppendAnd(g, x, r0^phase(3))
		w = Gia_ManAppendAnd(g, y, r1^phase(4))
	}
	root := Gia_ManAppendAnd(g, z, w)
	o := Gia_ManAppendAnd(g, a, c)
	dangling := Gia_ManAppendAnd(g, o, r0)
	marked := map[int32]uint32{r0 >> 1: 1, r1 >> 1: 1, z >> 1: 1, w >> 1: 1, root >> 1: 1, dangling >> 1: 1}
	literals := []int32{root ^ phase(5), x, z ^ 1, 0, 1}
	if mode == 1 {
		literals = append(literals, a)
	} else if mode == 2 {
		literals = append(literals, Gia_ManAppendAnd(g, a, b))
	}
	literals = append(literals, z^phase(6), w)
	for _, lit := range literals {
		co := Gia_ManAppendCo(g, lit)
		if marked[lit>>1] != 0 {
			marked[co>>1] = 1
		}
	}
	Gia_ManSetRegNum(g, 2)
	g.SetIdentity("cut", "cut.aig")
	g.nConstrs = 1
	cuts := []int32{x >> 1, y >> 1, o >> 1}
	if layout {
		old := g
		g = Gia_ManDupOrderDfsReverse(old)
		var kept []int32
		for _, id := range cuts {
			if old.pObjs[id].Value != ^uint32(0) {
				kept = append(kept, int32(old.pObjs[id].Value>>1))
			}
		}
		cuts = kept
		next := map[int32]uint32{}
		for id := range marked {
			if old.pObjs[id].Value != ^uint32(0) {
				next[int32(old.pObjs[id].Value>>1)] = 1
			}
		}
		marked = next
	}
	g.pCexSeq = util.AbcCexSentinel
	return g, cuts, marked
}

func cutReviewRecords(t *testing.T, typed, layout bool, modes int) string {
	t.Helper()
	var records strings.Builder
	for signs := 0; signs < 128; signs++ {
		for mode := 0; mode < modes; mode++ {
			src, cuts, marked := cutReviewGraph(signs, mode, typed, layout)
			for _, obj := range src.pObjs {
				obj.fMark0 = 1
				obj.fMark1 = uint8(obj.id & 1)
			}
			Gia_ManHighLightFlopLogic(src)
			var markBits uint64
			for id, obj := range src.pObjs {
				if uint32(obj.fMark0) != marked[int32(id)] {
					t.Fatalf("typed%v layout%v object%d: wrong structural register dependence", typed, layout, id)
				}
				markBits |= uint64(obj.fMark0) << uint(id)
			}
			out := Gia_ManDupReplaceCut(src)
			freeInputs := mode // mode1 adds a; mode2 retains the original a,b output cone.
			piCount := len(cuts) + freeInputs
			gates := 3
			if mode == 2 {
				gates++
			}
			name, spec := out.Identity()
			if Gia_ManPiNum(out) != int32(piCount) || out.nRegs != 2 || Gia_ManCoNum(out) != Gia_ManCoNum(src) || Gia_ManAndNum(out) != int32(gates) || out.nConstrs != 1 || name != "cut" || spec != "cut.aig" || out.pCexSeq != nil {
				t.Fatal("cut replacement changed copied logic, interface or role metadata")
			}
			for _, obj := range src.pObjs {
				if obj.fMark0 != 0 || obj.fMark1 != uint8(obj.id&1) {
					t.Fatal("cut copy did not clean mark0 or changed independent mark1")
				}
			}
			position := func(id int32) int {
				value := src.pObjs[id].Value
				if value == ^uint32(0) || value&1 != 0 {
					t.Fatal("invalid cut/input mapping")
				}
				obj := Gia_ManObj(out, int32(value>>1))
				if Gia_ObjIsCi(obj) == 0 {
					t.Fatal("cut/input mapped to a non-CI")
				}
				return int(Gia_ObjCioId(obj))
			}
			for _, id := range cuts {
				if position(id) >= piCount {
					t.Fatal("boundary PI was placed in the register tail")
				}
			}
			for reg := int32(0); reg < 2; reg++ {
				obj := Gia_ManRo(src, reg)
				if obj.Value != uint32(Gia_ManRo(out, reg).id<<1) || position(obj.id) != piCount+int(reg) {
					t.Fatal("RO identity/order lost behind late PI discovery")
				}
			}
			for i := int32(0); i < 3; i++ {
				value := Gia_ManPi(src, i).Value
				used := mode == 1 && i == 0 || mode == 2 && i < 2
				if used {
					if value == ^uint32(0) || position(Gia_ManPi(src, i).id) >= piCount {
						t.Fatal("uncut PI output logic lost its original inputs")
					}
				} else if value != ^uint32(0) {
					t.Fatal("logic below the cut leaked original PI slots")
				}
			}
			fmt.Fprintf(&records, "%d %d %d %d %d %d %d", signs, mode, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs, markBits)
			for _, obj := range src.pObjs {
				fmt.Fprintf(&records, " %d", obj.Value)
			}
			for mask := 0; mask < 1<<uint(Gia_ManCiNum(out)); mask++ {
				inputs := make([]bool, Gia_ManCiNum(out))
				for i := range inputs {
					inputs[i] = mask&(1<<uint(i)) != 0
				}
				x, y := inputs[position(cuts[0])], inputs[position(cuts[1])]
				r0, r1 := inputs[piCount], inputs[piCount+1]
				control := r0 != (signs&8 != 0)
				z, w := x && control, y && (r1 != (signs&16 != 0))
				if typed {
					z = y
					if control {
						z = x
					}
					w = y != (r1 != (signs&16 != 0))
				}
				want := []bool{(z && w) != (signs&32 != 0), x, !z, false, true}
				if mode == 1 {
					want = append(want, inputs[position(Gia_ManPi(src, 0).id)])
				} else if mode == 2 {
					want = append(want, inputs[position(Gia_ManPi(src, 0).id)] && inputs[position(Gia_ManPi(src, 1).id)])
				}
				want = append(want, z != (signs&64 != 0), w)
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
			// Concrete PI/state assignments lift to cut signals and preserve every CO,
			// including next-state bits. Naive cuts need not preserve joint input range.
			for mask := 0; mask < 32; mask++ {
				original := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
				values, expected, err := src.evaluateObjects(original)
				if err != nil {
					t.Fatal(err)
				}
				inputs := make([]bool, Gia_ManCiNum(out))
				for _, id := range cuts {
					inputs[position(id)] = values[id]
				}
				for i := int32(0); i < 5; i++ {
					obj := Gia_ManCi(src, i)
					if obj.Value != ^uint32(0) {
						inputs[position(obj.id)] = original[i]
					}
				}
				checkQuantificationOutputs(t, out, inputs, expected)
			}
		}
	}
	return records.String()
}

func TestCutMarkerAndLiftedStateBehavior(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			t.Run(fmt.Sprintf("typed%v_layout%v", typed, layout), func(t *testing.T) { cutReviewRecords(t, typed, layout, 3) })
		}
	}
}

func TestReplaceCutMatchesWorkingCRegion(t *testing.T) {
	exe := compileGiaSourceOracle(t, "replace_cut.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C cut: %v: %s", err, native)
	}
	got, want := strings.Split(strings.TrimSpace(cutReviewRecords(t, false, false, 1)), "\n"), strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases %d, C cases %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case %d:\nGo %s\nC %s", i, got[i], want[i])
		}
	}
	witness, err := testoracle.Command(t, exe, "late").CombinedOutput()
	if err != nil || string(witness) != "1 0 2 1\n" {
		t.Fatalf("original C late PI/register witness: %v: %s", err, witness)
	}
}

func TestReplaceCutBoundaryGraphs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		src := Gia_ManStart(1)
		src.pObjs[0].fMark0 = 1
		out := Gia_ManDupReplaceCut(src)
		if out.nObjs != 1 || out.nRegs != 0 || src.pObjs[0].fMark0 != 0 {
			t.Fatal("empty cut copy changed graph/constant mark")
		}
	})
	t.Run("constant outputs", func(t *testing.T) {
		src := Gia_ManStart(1)
		Gia_ManAppendCo(src, 0)
		Gia_ManAppendCo(src, 1)
		src.nConstrs = 1
		out := Gia_ManDupReplaceCut(src)
		checkQuantificationOutputs(t, out, nil, []bool{false, true})
		if Gia_ManCiNum(out) != 0 || out.nConstrs != 1 {
			t.Fatal("constant outputs became free inputs or lost roles")
		}
	})
	t.Run("combinational output logic", func(t *testing.T) {
		src := Gia_ManStart(1)
		a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
		Gia_ManAppendCi(src)
		Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, b))
		out := Gia_ManDupReplaceCut(src)
		if Gia_ManCiNum(out) != 2 || Gia_ManAndNum(out) != 1 || out.nRegs != 0 {
			t.Fatal("a graph with no state-dependent cut lost its combinational logic")
		}
		for mask := 0; mask < 4; mask++ {
			a, b := mask&1 != 0, mask&2 != 0
			checkQuantificationOutputs(t, out, []bool{a, b}, []bool{a && b})
		}
	})
	t.Run("register-only", func(t *testing.T) {
		src := Gia_ManStart(1)
		r0, r1 := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
		Gia_ManAppendCo(src, r1^1)
		Gia_ManAppendCo(src, r0)
		Gia_ManSetRegNum(src, 2)
		out := Gia_ManDupReplaceCut(src)
		if Gia_ManPiNum(out) != 0 || out.nRegs != 2 || Gia_ManPoNum(out) != 0 {
			t.Fatal("register-only cut changed state interface")
		}
		for mask := 0; mask < 4; mask++ {
			r0, r1 := mask&1 != 0, mask&2 != 0
			checkQuantificationOutputs(t, out, []bool{r0, r1}, []bool{!r1, r0})
		}
	})
	t.Run("autonomous constant state", func(t *testing.T) {
		src := Gia_ManStart(1)
		Gia_ManAppendCi(src)
		Gia_ManAppendCo(src, 0)
		Gia_ManSetRegNum(src, 1)
		out := Gia_ManDupReplaceCut(src)
		if Gia_ManPiNum(out) != 0 || out.nRegs != 1 {
			t.Fatal("unused RO was removed")
		}
		for _, r := range []bool{false, true} {
			checkQuantificationOutputs(t, out, []bool{r}, []bool{false})
		}
	})
}

func TestReplaceCutTreatsPiOnlyMuxControlAsBoundary(t *testing.T) {
	for signs := 0; signs < 8; signs++ {
		src := Gia_ManStart(1)
		a, b, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
		control := Gia_ManAppendAnd(src, a^int32(signs&1), b)
		src.EnableMuxes()
		mux := Gia_ManAppendMuxReal(src, control^int32((signs>>1)&1), r, a^int32((signs>>2)&1))
		Gia_ManAppendCo(src, mux)
		Gia_ManAppendCo(src, r)
		Gia_ManSetRegNum(src, 1)
		out := Gia_ManDupReplaceCut(src)
		if Gia_ManPiNum(out) != 2 || out.nRegs != 1 || Gia_ManAndNum(out) != 1 || Gia_ObjCioId(Gia_ManObj(out, int32(src.pObjs[a>>1].Value>>1))) != 0 || Gia_ObjCioId(Gia_ManObj(out, int32(src.pObjs[control>>1].Value>>1))) != 1 {
			t.Fatal("pure control/data boundary order or RO tail was lost")
		}
		for mask := 0; mask < 8; mask++ {
			a, control, r := mask&1 != 0, mask&2 != 0, mask&4 != 0
			value := a != (signs&4 != 0)
			if control != (signs&2 != 0) {
				value = r
			}
			checkQuantificationOutputs(t, out, []bool{a, control, r}, []bool{value, r})
			originalControl := (a != (signs&1 != 0)) && control // Here control supplies the original b input.
			want, err := src.Evaluate([]bool{a, control, r})
			if err != nil {
				t.Fatal(err)
			}
			checkQuantificationOutputs(t, out, []bool{a, originalControl, r}, want)
		}
	}
}
