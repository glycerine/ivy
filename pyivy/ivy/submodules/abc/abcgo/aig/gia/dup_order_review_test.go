package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func TestOrderedCopiesPreserveSimpleMode(t *testing.T) {
	copies := []struct {
		name string
		copy func(*Gia_Man_t) *Gia_Man_t
	}{
		{"reverse", Gia_ManDupOrderDfsReverse}, {"aiger", Gia_ManDupOrderAiger},
		{"abstract", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupAbs(g, permutation(), permutation()) }},
		{"group", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupOutputGroup(g, 0, 1) }},
		{"selected", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupSelectedOutputs(g, permutation(0)) }},
		{"presence", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupOutputVec(g, permutation(1)) }},
	}
	for _, tc := range copies {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Errorf("valid copy panicked: %v", err)
				}
			}()
			src := Gia_ManStart(1)
			src.fGiaSimple = 1
			a := Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, a^1))
			out := tc.copy(src)
			if out.fGiaSimple != 1 || out.nObjs != src.nObjs {
				t.Fatal("copy changed simple AND storage")
			}
			for _, input := range []bool{false, true} {
				checkQuantificationOutputs(t, out, []bool{input}, []bool{false})
			}
		})
	}
}

func TestOrderedCopiesPreserveConstraints(t *testing.T) {
	copies := []struct {
		name string
		copy func(*Gia_Man_t) *Gia_Man_t
	}{
		{"reverse", Gia_ManDupOrderDfsReverse}, {"aiger", Gia_ManDupOrderAiger},
		{"abstract", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupAbs(g, permutation(), permutation(-1)) }},
	}
	for _, tc := range copies {
		t.Run(tc.name, func(t *testing.T) {
			src := Gia_ManStart(1)
			a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, a)
			Gia_ManAppendCo(src, a^1)
			Gia_ManAppendCo(src, r)
			Gia_ManSetRegNum(src, 1)
			src.nConstrs = 1
			out := tc.copy(src)
			if out.nConstrs != 1 {
				t.Fatal("copy reclassified unchanged constraint PO")
			}
		})
	}
}

var orderReviewCopies = []struct {
	name string
	copy func(*Gia_Man_t) *Gia_Man_t
}{{"dfs", Gia_ManDupOrderDfs}, {"reverse", Gia_ManDupOrderDfsReverse}, {"aiger", Gia_ManDupOrderAiger}}

func orderReviewRecords(t *testing.T, typed, dfs bool, constraints int32) string {
	t.Helper()
	var records strings.Builder
	for mode, method := range orderReviewCopies {
		for signs := int32(0); signs < 32; signs++ {
			for _, regs := range []int32{0, 2, 4} {
				src, _ := levelReviewGraph(signs, 0, typed)
				Gia_ManAppendCi(src) // Retain an unused input created after all outputs.
				Gia_ManSetRegNum(src, regs)
				src.nConstrs = constraints
				if dfs {
					src = Gia_ManDupOrderDfs(src)
				}
				out := method.copy(src)
				name, spec := out.Identity()
				wantNodes := src.nObjs
				if typed && !dfs && mode != 2 {
					wantNodes--
				} // Drop the dangling AND only in DFS copies.
				if out.nObjs != wantNodes || Gia_ManCiNum(out) != 5 || Gia_ManCoNum(out) != Gia_ManCoNum(src) || out.nRegs != regs || out.nConstrs != constraints || name != "levels" || spec != "levels.aig" {
					t.Fatalf("%s changed graph/interface/identity", method.name)
				}
				if mode == 2 && Gia_ManIsNormalized(out) == 0 {
					t.Fatal("AIGER copy not normalized")
				}
				for i := int32(0); i < Gia_ManCiNum(out); i++ {
					obj := Gia_ManCi(out, i)
					if Gia_ObjCioId(obj) != i || uint32(obj.id<<1) != Gia_ManCi(src, i).Value {
						t.Fatal("copy changed CI order or ordinal")
					}
				}
				for i := int32(0); i < Gia_ManCoNum(out); i++ {
					obj := Gia_ManCo(out, i)
					if Gia_ObjCioId(obj) != i || uint32(obj.id<<1) != Gia_ManCo(src, i).Value {
						t.Fatal("copy changed CO order or ordinal")
					}
				}
				fmt.Fprintf(&records, "%d %d %d %d %d %d %d", mode, signs, regs, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
				for _, obj := range out.pObjs {
					a, b, _, kind := out.Gate(obj.id)
					ordinal := int32(-1)
					switch kind {
					case 0:
						a, b = 0, 0
					case 1:
						ordinal = Gia_ObjCioId(obj)
						a, b = 0, 0
					case 5:
						b = 0
					case 6:
						ordinal = Gia_ObjCioId(obj)
						b = 0
					}
					fmt.Fprintf(&records, " %d:%d:%d:%d", kind, a, b, ordinal)
				}
				for _, obj := range src.pObjs {
					fmt.Fprintf(&records, " %d", obj.Value)
				}
				for mask := 0; mask < 32; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
					want, _ := selectionReviewTruth(signs, typed, inputs)
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
	return records.String()
}

func TestOrderedCopyLayoutsAndSemantics(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, dfs := range []bool{false, true} {
			for _, constraints := range []int32{0, 1} {
				t.Run(fmt.Sprintf("typed%v_dfs%v_constraints%d", typed, dfs, constraints), func(t *testing.T) { orderReviewRecords(t, typed, dfs, constraints) })
			}
		}
	}
}

func TestOrderedCopiesMatchOriginalSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "ordered_copies.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C ordered copies: %v: %s", err, native)
	}
	got := strings.Split(strings.TrimSpace(orderReviewRecords(t, false, false, 0)), "\n")
	want := strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases %d, C cases %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case %d:\nGo %s\nC %s", i, got[i], want[i])
		}
	}
	reprs, err := testoracle.Command(t, exe, "--representatives").CombinedOutput()
	if err != nil {
		t.Fatalf("C representatives: %v: %s", err, reprs)
	}
	if got := orderReviewRepresentativeRecords(t); got != string(reprs) {
		t.Fatalf("representatives differ:\nGo %s\nC %s", got, reprs)
	}

	witness, err := testoracle.Command(t, exe, "--constraints").CombinedOutput()
	if err != nil || string(witness) != "1 0 0 0 0\n" {
		t.Fatalf("inherited metadata witness: %v: %s", err, witness)
	}
}

func orderReviewRepresentativeRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for mode, method := range orderReviewCopies {
		for regs := int32(0); regs <= 2; regs++ {
			src := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			Gia_ManAppendCi(src)
			x := Gia_ManAppendAnd(src, a, b)
			negative := Gia_ManAppendAnd(src, a^1, b)
			zero := Gia_ManAppendAnd(src, x, negative)
			Gia_ManAppendCo(src, zero)
			y := Gia_ManAppendAnd(src, a, b)
			Gia_ManAppendCo(src, x)
			Gia_ManAppendCo(src, y)
			Gia_ManSetRegNum(src, regs)
			src.StartRepresentatives()
			Gia_ObjSetRepr(src, y>>1, x>>1)
			Gia_ObjSetRepr(src, zero>>1, 0)
			src.PNexts = Gia_ManDeriveNexts(src)
			out := method.copy(src)
			first, second := int32(src.pObjs[x>>1].Value>>1), int32(src.pObjs[y>>1].Value>>1)
			if first > second {
				first, second = second, first
			}
			constant := int32(src.pObjs[zero>>1].Value >> 1)
			if Gia_ObjRepr(out, first) != GIA_VOID || Gia_ObjRepr(out, second) != first || Gia_ObjNext(out, first) != second || Gia_ObjNext(out, second) != 0 || Gia_ObjRepr(out, constant) != 0 {
				t.Fatal("copy changed equivalence class or constant representative")
			}
			if mode == 1 && int32(src.pObjs[y>>1].Value>>1) != first {
				t.Fatal("fixture failed to reverse representative identity")
			}
			fmt.Fprintf(&records, "%d %d %d", mode, regs, out.nObjs)
			for id := int32(0); id < out.nObjs; id++ {
				fmt.Fprintf(&records, " %d:%d", Gia_ObjRepr(out, id), Gia_ObjNext(out, id))
			}
			records.WriteByte('\n')
			for mask := 0; mask < 8; mask++ {
				a, b := mask&1 != 0, mask&2 != 0
				checkQuantificationOutputs(t, out, []bool{a, b, mask&4 != 0}, []bool{false, a && b, a && b})
			}
			Gia_ObjUnsetRepr(out, second)
			if Gia_ObjRepr(src, y>>1) != x>>1 {
				t.Fatal("result representatives alias source storage")
			}
		}
	}
	return records.String()
}

func TestOrderedCopyRepresentativeRemapping(t *testing.T) {
	orderReviewRepresentativeRecords(t)
}

func TestOrderedCopyBoundaryGraphs(t *testing.T) {
	for _, method := range orderReviewCopies {
		for _, mode := range []string{"empty", "constants", "unused", "constant-buffer", "register-only"} {
			t.Run(method.name+"/"+mode, func(t *testing.T) {
				src := Gia_ManStart(1)
				var inputs, want []bool
				switch mode {
				case "constants":
					Gia_ManAppendCo(src, 0)
					Gia_ManAppendCo(src, 1)
					want = []bool{false, true}
				case "unused":
					Gia_ManAppendCi(src)
					inputs = []bool{true}
				case "constant-buffer":
					Gia_ManAppendCo(src, Gia_ManAppendBuf(src, 1))
					want = []bool{true}
				case "register-only":
					r := Gia_ManAppendCi(src)
					Gia_ManAppendCo(src, r^1)
					Gia_ManSetRegNum(src, 1)
					inputs = []bool{true}
					want = []bool{false}
				}
				out := method.copy(src)
				if out.nObjs != src.nObjs || out.nRegs != src.nRegs || Gia_ManCiNum(out) != Gia_ManCiNum(src) || Gia_ManCoNum(out) != Gia_ManCoNum(src) {
					t.Fatal("boundary copy changed nodes or interface")
				}
				checkQuantificationOutputs(t, out, inputs, want)
			})
		}
	}
}
