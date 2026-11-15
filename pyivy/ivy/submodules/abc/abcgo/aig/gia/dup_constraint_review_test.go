package gia

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestConstraintCollectorSharedConeTerminates(t *testing.T) {
	const key = "ABCGO_CONSTRAINT_SHARED_REVIEW_CHILD"
	if os.Getenv(key) == "1" {
		debug.SetMemoryLimit(32 << 20)
		timer := time.AfterFunc(2*time.Second, func() { panic("shared constraint cone expansion did not finish") })
		defer timer.Stop()
		src := topAndReviewSharedGraph(48)
		super := Vec_IntAlloc(3)
		Gia_ManDupWithConstrCollectAnd_rec(src, Gia_ObjFanin0(Gia_ManPo(src, 0)), super, 1)
		if !slices.Equal(Vec_IntArray(super), []int32{2, 4, 6}) {
			t.Fatalf("leaves %v", Vec_IntArray(super))
		}
		out := Gia_ManDupWithConstr(src)
		if out == nil || out.nConstrs != 1 || Gia_ManCiNum(out) != 3 || Gia_ManCoNum(out) != 2 {
			t.Fatal("constraint interface")
		}
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			values, err := out.Evaluate(inputs)
			if err != nil || (values[0] && !values[1]) != (inputs[0] && inputs[1] && inputs[2]) {
				t.Fatalf("constraint semantics %v: %v %v", inputs, values, err)
			}
		}
		parts := Gia_ManDupDemiter(src, 0)
		if Gia_ManCiNum(parts) != 3 || Gia_ManCoNum(parts) != 3 {
			t.Fatal("demiter caller lost shared-cone leaves")
		}
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			checkQuantificationOutputs(t, parts, inputs, inputs)
		}
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestConstraintCollectorSharedConeTerminates$", "-test.count=1")
	cmd.Env = append(os.Environ(), key+"=1")
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("shared constraint cone: %v (context %v): %s", err, ctx.Err(), data)
	}
}

func BenchmarkConstraintCollectorSharedCone(b *testing.B) {
	src := topAndReviewSharedGraph(12)
	root := Gia_ObjFanin0(Gia_ManPo(src, 0))
	super := Vec_IntAlloc(3)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Vec_IntClear(super)
		Gia_ManDupWithConstrCollectAnd_rec(src, root, super, 1)
	}
}

func constraintReviewGraph(signs, family int, typed, layout bool) (*Gia_Man_t, map[int32]int32) {
	g := Gia_ManStart(1)
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	levels := map[int32]int32{0: 0}
	inputs := make([]int32, 5)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(g)
		levels[inputs[i]>>1] = 0
	}
	a, b, c, d, e := inputs[0], inputs[1], inputs[2], inputs[3], inputs[4]
	and := func(a, b int32) int32 {
		lit := Gia_ManAppendAnd(g, a, b)
		levels[lit>>1] = 1 + max(levels[a>>1], levels[b>>1])
		return lit
	}
	var x, y int32
	if typed {
		x = Gia_ManAppendXorReal(g, a^phase(0), b^phase(1))
		levels[x>>1] = 2
	} else {
		x = and(a^phase(0), b^phase(1))
	}
	w := and(c^phase(2), d^phase(3))
	if typed {
		g.EnableMuxes()
		y = Gia_ManAppendMuxReal(g, c^phase(2), w, e^phase(4))
		levels[y>>1] = 3
	} else {
		y = and(w, e^phase(4))
	}
	var root int32
	switch family {
	case 0:
		root = and(x, y)
	case 1:
		root = and(x^1, y)
	case 2:
		root = and(x^1, y^1)
	case 3:
		left, right := and(x, y), and(x, e)
		root = and(left, right)
	case 4:
		root = and(x, 0)
	case 5:
		root = and(x, 1)
	case 6:
		root = and(x, a^phase(0)^1)
	default:
		panic("unknown constraint review family")
	}
	co := Gia_ManAppendCo(g, root)
	levels[co>>1] = levels[root>>1]
	g.SetIdentity("constr", "constr.aig")
	if layout {
		old := g
		g = Gia_ManDupOrderDfsReverse(old)
		mapped := map[int32]int32{}
		for id, level := range levels {
			if old.pObjs[id].Value != ^uint32(0) {
				mapped[int32(old.pObjs[id].Value>>1)] = level
			}
		}
		levels = mapped
	}
	return g, levels
}

// Enumerate complete left-to-right paths on the small fixtures. This oracle
// deliberately unfolds shared nodes; the production collector must avoid that
// cost on the large watchdog case while returning exactly this leaf sequence.
func constraintReviewLeaves(g *Gia_Man_t, lit int32, first bool, initial []int32) []int32 {
	result := slices.Clone(initial)
	stack := []int32{lit}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		obj := g.pObjs[current>>1]
		plain := Gia_ObjIsAnd(obj) != 0 && Gia_ObjIsXor(obj) == 0 && Gia_ObjIsMux(g, obj) == 0 && Gia_ObjIsBuf(obj) == 0
		if !plain || !first && current&1 != 0 {
			if !slices.Contains(result, current) {
				result = append(result, current)
			}
		} else {
			stack = append(stack, Gia_ObjFaninLit1(obj, obj.id), Gia_ObjFaninLit0(obj, obj.id))
		}
		first = false
	}
	return result
}

func constraintReviewRecords(t *testing.T, typed bool) string {
	t.Helper()
	var records strings.Builder
	for signs := 0; signs < 32; signs++ {
		for family := 0; family < 7; family++ {
			for _, layout := range []bool{false, true} {
				src, levels := constraintReviewGraph(signs, family, typed, layout)
				root := Gia_ObjFaninLit0(Gia_ManPo(src, 0), Gia_ManPo(src, 0).id)
				for _, obj := range src.pObjs {
					obj.Value = uint32(obj.id + 77)
					obj.fMark0 = uint32(obj.id & 1)
					obj.fMark1 = uint32((obj.id >> 1) & 1)
				}
				leaves := constraintReviewLeaves(src, root, true, nil)
				super := Vec_IntAlloc(8)
				Gia_ManDupWithConstrCollectAnd_rec(src, Gia_ObjFromLit(src, root), super, 1)
				if !slices.Equal(Vec_IntArray(super), leaves) {
					t.Fatalf("typed%v signs%d family%d layout%v: leaves %v want %v", typed, signs, family, layout, Vec_IntArray(super), leaves)
				}
				fmt.Fprintf(&records, "%d %d %t leaves", signs, family, layout)
				for _, lit := range leaves {
					fmt.Fprintf(&records, " %d", lit)
				}
				// C appends source CIs in physical order, even after DFS has
				// restored their logical vector order. The native fixture reports
				// that permutation and aligns its evaluation to source CI ports.
				fmt.Fprint(&records, " ports")
				for _, obj := range src.pObjs {
					if Gia_ObjIsCi(obj) != 0 {
						fmt.Fprintf(&records, " %d", Gia_ObjCioId(obj))
					}
				}
				// Exercise stopping rules, both polarities, root override and append-to-vector
				// behavior, including COs and typed leaves supported by the Go port.
				for _, obj := range src.pObjs {
					for sign := int32(0); sign < 2; sign++ {
						for _, first := range []bool{false, true} {
							lit := obj.id<<1 | sign
							initial := []int32{Gia_ManCi(src, 0).id << 1, root ^ 1}
							Vec_IntClear(super)
							for _, v := range initial {
								Vec_IntPush(super, v)
							}
							flag := int32(0)
							if first {
								flag = 1
							}
							Gia_ManDupWithConstrCollectAnd_rec(src, Gia_ObjFromLit(src, lit), super, flag)
							want := constraintReviewLeaves(src, lit, first, initial)
							if !slices.Equal(Vec_IntArray(super), want) {
								t.Fatalf("collector stopping/append rules root%d first%v: %v want %v", lit, first, Vec_IntArray(super), want)
							}
						}
					}
				}
				for _, obj := range src.pObjs {
					if obj.Value != uint32(obj.id+77) || obj.fMark0 != uint32(obj.id&1) || obj.fMark1 != uint32((obj.id>>1)&1) {
						t.Fatal("collector changed source scratch")
					}
				}
				best := leaves[0]
				for _, lit := range leaves[1:] {
					if levels[lit>>1] > levels[best>>1] {
						best = lit
					}
				}
				out := Gia_ManDupWithConstr(src)
				if out == nil || out.nRegs != 0 || Gia_ManCiNum(out) != 5 || Gia_ManCoNum(out) != 2 || out.nConstrs != 1 {
					t.Fatal("constraint extraction lost interface/roles")
				}
				name, spec := out.Identity()
				if name != "constr" || spec != "constr.aig" {
					t.Fatal("identity lost")
				}
				if Gia_ManHashStarted(out) != 0 {
					t.Fatal("hashing left enabled")
				}
				fmt.Fprintf(&records, " result %d %d %d %d %d", out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs, out.nConstrs)
				for mask := 0; mask < 32; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
					values, original, err := src.evaluateObjects(inputs)
					if err != nil {
						t.Fatal(err)
					}
					literal := func(lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
					others := true
					for _, lit := range leaves {
						if lit != best {
							others = others && literal(lit)
						}
					}
					expected := []bool{literal(best), !others}
					checkQuantificationOutputs(t, out, inputs, expected)
					if (expected[0] && !expected[1]) != original[0] {
						t.Fatal("constraint decomposition changed accepted assignments")
					}
					packed := 0
					for i, v := range expected {
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

func TestConstraintExtractionReviewMatrix(t *testing.T) {
	constraintReviewRecords(t, false)
	constraintReviewRecords(t, true)
}

func TestConstraintExtractionMatchesOriginalC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "constraint_extract.c")
	output, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C constraint extraction: %v: %s", err, output)
	}
	got, want := strings.Split(strings.TrimSpace(constraintReviewRecords(t, false)), "\n"), strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases%d C cases%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case%d\nGo %s\nC %s", i, got[i], want[i])
		}
	}
}

func TestConstraintExtractionBoundaries(t *testing.T) {
	for _, kind := range []string{"zero", "one", "input", "inverted", "buffer", "xor"} {
		t.Run(kind, func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			var lit int32
			switch kind {
			case "zero":
				lit = 0
			case "one":
				lit = 1
			case "input":
				lit = a
			case "inverted":
				lit = Gia_ManAppendAnd(src, a, b) ^ 1
			case "buffer":
				lit = Gia_ManAppendBuf(src, a)
			case "xor":
				lit = Gia_ManAppendXorReal(src, a, b)
			}
			Gia_ManAppendCo(src, lit)
			if Gia_ManDupWithConstr(src) != nil {
				t.Fatal("nondecomposable output accepted")
			}
		})
	}
	for _, kind := range []string{"none", "multiple", "state"} {
		t.Run(kind, func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			if kind != "none" {
				Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, b))
				Gia_ManAppendCo(src, a)
			}
			if kind == "state" {
				Gia_ManSetRegNum(src, 1)
			}
			defer func() {
				if recover() == nil {
					t.Fatal("invalid constraint-extraction interface accepted")
				}
			}()
			Gia_ManDupWithConstr(src)
		})
	}
}

func TestConstraintCollectorDemiterStateCaller(t *testing.T) {
	for _, inverted := range []int32{0, 1} {
		src := Gia_ManStart(1)
		a, b, c, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
		root := Gia_ManAppendAnd(src, Gia_ManAppendAnd(src, a, b), r^1)
		Gia_ManAppendCo(src, root^inverted)
		Gia_ManAppendCo(src, c)
		Gia_ManSetRegNum(src, 1)
		out := Gia_ManDupDemiter(src, 0)
		if out.nRegs != 1 || Gia_ManPiNum(out) != 3 || Gia_ManPoNum(out) != 3 {
			t.Fatal("demiter state interface changed")
		}
		for mask := 0; mask < 16; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			got, err := out.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			value := inverted == 0
			for _, part := range got[:3] {
				if inverted == 0 {
					value = value && part
				} else {
					value = value || part
				}
			}
			original, err := src.Evaluate(inputs)
			if err != nil || value != original[0] || got[3] != original[1] {
				t.Fatalf("demiter state/caller semantics %v: %v / %v (%v)", inputs, got, original, err)
			}
		}
	}
}
