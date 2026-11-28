package gia

import (
	"context"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

func topAndReviewSatCount(t *testing.T, g *Gia_Man_t) int {
	t.Helper()
	count := 0
	for mask := 0; mask < 1<<uint(Gia_ManCiNum(g)); mask++ {
		inputs := make([]bool, Gia_ManCiNum(g))
		for i := range inputs {
			inputs[i] = mask&(1<<uint(i)) != 0
		}
		values, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		all := true
		for _, value := range values {
			all = all && value
		}
		if all {
			count++
		}
	}
	return count
}

func TestTopAndRetainsFalseOutput(t *testing.T) {
	for _, full := range []bool{false, true} {
		t.Run(fmt.Sprint(full), func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, 0)
			Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, b))
			var out *Gia_Man_t
			if full {
				out = Gia_ManDupTopAnd(src, 0)
			} else {
				out = Gia_ManDupTopAnd_iter(src, 0)
			}
			if topAndReviewSatCount(t, src) != 0 || topAndReviewSatCount(t, out) != 0 {
				t.Fatal("AND decomposition removed a false conjunct and made UNSAT satisfiable")
			}
		})
	}
}

func TestDupClassesPreservesConstraintRoles(t *testing.T) {
	src := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a^1)
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	src.pReprsOld = make([]int32, src.nObjs)
	for i := range src.pReprsOld {
		src.pReprsOld[i] = -1
	}
	if out := Gia_ManDupDfsClasses(src); out.nConstrs != 1 {
		t.Fatal("class duplication reclassified a retained constraint")
	}
}

func topAndReviewSharedGraph(depth int) *Gia_Man_t {
	src := Gia_ManStart(1)
	a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	x, y := Gia_ManAppendAnd(src, a, b), Gia_ManAppendAnd(src, a, c)
	for i := 0; i < depth; i++ {
		old := x
		x = Gia_ManAppendAnd(src, x, y)
		y = Gia_ManAppendAnd(src, old, a)
	}
	Gia_ManAppendCo(src, x)
	return src
}

func TestTopAndSharedConeTerminates(t *testing.T) {
	const key = "ABCGO_TOPAND_SHARED_REVIEW_CHILD"
	if os.Getenv(key) == "1" {
		debug.SetMemoryLimit(32 << 20)
		timer := time.AfterFunc(2*time.Second, func() { panic("shared cone expansion did not finish") })
		defer timer.Stop()
		out := Gia_ManDupTopAnd_iter(topAndReviewSharedGraph(48), 0)
		if Gia_ManCiNum(out) != 3 || Gia_ManCoNum(out) != 0 || Gia_ManAndNum(out) != 0 {
			t.Fatal("shared conjunction was not reduced to its satisfiable CI assignments")
		}
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := testoracle.CommandContext(t, ctx, exe, "-test.run=^TestTopAndSharedConeTerminates$", "-test.count=1")
	cmd.Env = append(os.Environ(), key+"=1")
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("shared cone: %v (context %v): %s", err, ctx.Err(), data)
	}
}

func TestPrintRepresentativesFilteredNewTable(t *testing.T) {
	const key = "ABCGO_PRINT_OLD_REVIEW_CHILD"
	if mode := os.Getenv(key); mode != "" {
		src := Gia_ManStart(1)
		for i := 0; i < 3; i++ {
			Gia_ManAppendCi(src)
		}
		src.pReprsOld = []int32{-1, -1, -1, 1}
		if mode == "both" {
			src.StartRepresentatives()
			Gia_ObjSetRepr(src, 3, 2)
		}
		Gia_ManPrintRepr(src)
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"old-only", "both"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := testoracle.CommandContext(t, ctx, exe, "-test.run=^TestPrintRepresentativesFilteredNewTable$", "-test.count=1")
			cmd.Env = append(os.Environ(), key+"="+mode)
			data, err := cmd.CombinedOutput()
			if mode == "old-only" {
				if err == nil || !strings.Contains(string(data), "invalid representatives") {
					t.Fatalf("missing new table was not rejected: %v: %s", err, data)
				}
				return
			}
			if err != nil || !strings.HasPrefix(string(data), "3->2 \n") {
				t.Fatalf("old representatives: %v: %s", err, data)
			}
		})
	}
}

var topAndReviewSink *Gia_Man_t

func BenchmarkTopAndSharedFrontier(b *testing.B) {
	src := topAndReviewSharedGraph(12)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		topAndReviewSink = Gia_ManDupTopAnd_iter(src, 0)
	}
}

func topAndReviewGraph(kind, signs int) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	phase := func(bit int) int32 { return int32((signs >> bit) & 1) }
	u, v := Gia_ManAppendAnd(g, a^phase(0), b^phase(1)), Gia_ManAppendAnd(g, c^phase(2), d^phase(3))
	switch kind {
	case 0, 1, 6:
		root := Gia_ManAppendAnd(g, u, v^1)
		if kind == 6 {
			Gia_ManAppendCo(g, 0)
		}
		Gia_ManAppendCo(g, root)
		if kind == 1 {
			Gia_ManAppendCo(g, a^1)
		} else {
			Gia_ManAppendCo(g, v^1)
			Gia_ManAppendCo(g, 1)
		}
	case 2:
		Gia_ManAppendCo(g, u^1)
		Gia_ManAppendCo(g, v^1)
	case 3:
		Gia_ManAppendCo(g, a^phase(0))
		Gia_ManAppendCo(g, b^phase(1))
	case 4:
		x, y := u, v
		for i := 0; i < 4; i++ {
			old := x
			x = Gia_ManAppendAnd(g, x, y)
			y = Gia_ManAppendAnd(g, old, a)
		}
		Gia_ManAppendCo(g, x)
		Gia_ManAppendCo(g, x)
		Gia_ManAppendCo(g, 1)
	case 5:
		Gia_ManAppendCo(g, a)
		Gia_ManAppendCo(g, a^1)
		Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
	case 7:
		Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, 0))
		Gia_ManAppendCo(g, 1)
	case 8:
		g.EnableMuxes()
		x := Gia_ManAppendXorReal(g, a^phase(0), b^phase(1))
		mux := Gia_ManAppendMuxReal(g, c, x, d^phase(2))
		Gia_ManAppendCo(g, Gia_ManAppendAnd(g, mux, a^phase(3)))
		Gia_ManAppendCo(g, 1)
	case 9:
		buf := Gia_ManAppendBuf(g, c^phase(2))
		Gia_ManAppendCo(g, Gia_ManAppendAnd(g, buf, a^phase(0)))
		Gia_ManAppendCo(g, b^phase(1))
	}
	Gia_ManAppendCi(g)
	g.SetIdentity("top", "top.aig")
	return g
}

func topAndReviewAll(values []bool) bool {
	for _, value := range values {
		if !value {
			return false
		}
	}
	return true
}

func topAndReviewRecords(t *testing.T, extended, layout bool) string {
	t.Helper()
	var records strings.Builder
	count := 6
	if extended {
		count = 10
	}
	for kind := 0; kind < count; kind++ {
		for signs := 0; signs < 16; signs++ {
			for full := 0; full < 2; full++ {
				src := topAndReviewGraph(kind, signs)
				if layout {
					src = Gia_ManDupOrderDfsReverse(src)
				}
				var out *Gia_Man_t
				if full != 0 {
					out = Gia_ManDupTopAnd(src, 0)
				} else {
					out = Gia_ManDupTopAnd_iter(src, 0)
				}
				name, spec := out.Identity()
				if Gia_ManCiNum(out) != 5 || out.nRegs != 0 || name != "top" || spec != "top.aig" {
					t.Fatal("decomposition changed retained input interface/identity")
				}
				originalSat, resultSat := 0, 0
				fmt.Fprintf(&records, "R %d %d %d %d %d %d", kind, signs, full, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out))
				for _, obj := range src.pObjs {
					fmt.Fprintf(&records, " %d", obj.Value)
				}
				for mask := 0; mask < 32; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
					original, err := src.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					values, err := out.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					if topAndReviewAll(original) {
						originalSat++
						if !topAndReviewAll(values) {
							t.Fatal("decomposition removed a satisfying assignment")
						}
					}
					if topAndReviewAll(values) {
						resultSat++
						if full == 0 {
							lifted := make([]bool, 5)
							for i := int32(0); i < 5; i++ {
								lit := int32(Gia_ManCi(src, i).Value)
								if lit < 0 {
									t.Fatal("decomposition left an unmapped CI")
								}
								var value bool
								if lit < 2 {
									value = lit != 0
								} else {
									value = inputs[Gia_ObjCioId(Gia_ManObj(out, lit>>1))] != (lit&1 != 0)
								}
								lifted[i] = value
							}
							originalLifted, err := src.Evaluate(lifted)
							if err != nil {
								t.Fatal(err)
							}
							if !topAndReviewAll(originalLifted) {
								t.Fatal("decomposition model cannot be lifted using the forced CI assignments")
							}
						}
					}
					packed := 0
					for i, value := range values {
						if value {
							packed |= 1 << uint(i)
						}
					}
					fmt.Fprintf(&records, " %d", packed)
				}
				if (originalSat != 0) != (resultSat != 0) {
					t.Fatalf("kind%d signs%d full%d: decomposition changed conjunction satisfiability", kind, signs, full)
				}
				records.WriteByte('\n')
			}
		}
	}
	return records.String()
}

func topAndReviewClassRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 32; signs++ {
		for _, regs := range []int32{0, 2, 4} {
			src, _ := dfsReviewRepresentativeGraph(signs, regs)
			src.nConstrs = 1
			out := Gia_ManDupDfsClasses(src)
			if Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != 5 || out.nRegs != regs || out.nConstrs != 1 || Gia_ManAndNum(out) != 1 || Gia_ManIsNormalized(out) == 0 || Gia_ManHashStarted(out) != 0 {
				t.Fatal("class reduction changed interface, proof substitution or normalized result")
			}
			fmt.Fprintf(&records, "R %d %d %d", signs, regs, out.nObjs)
			for _, obj := range src.pObjs {
				fmt.Fprintf(&records, " %d", obj.Value)
			}
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				x := (inputs[0] != (signs&1 != 0)) && (inputs[1] != (signs&2 != 0))
				want := []bool{x != (signs&4 != 0), (!x) != (signs&8 != 0), (!x) != (signs&16 != 0), false, true}
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

func TestTopAndSemanticAndClassMatrices(t *testing.T) {
	topAndReviewRecords(t, true, false)
	topAndReviewRecords(t, true, true)
	topAndReviewClassRecords(t)
}

func TestTopAndMatchesWorkingCPaths(t *testing.T) {
	exe := compileGiaSourceOracle(t, "top_and.c")
	for _, mode := range []string{"top", "classes"} {
		t.Run(mode, func(t *testing.T) {
			native, err := testoracle.Command(t, exe, mode).CombinedOutput()
			if err != nil {
				t.Fatalf("C top AND: %v: %s", err, native)
			}
			var records string
			if mode == "top" {
				records = topAndReviewRecords(t, false, false)
			} else {
				records = topAndReviewClassRecords(t)
			}
			var got, want []string
			for _, line := range strings.Split(records, "\n") {
				if strings.HasPrefix(line, "R ") {
					got = append(got, line)
				}
			}
			for _, line := range strings.Split(string(native), "\n") {
				if strings.HasPrefix(line, "R ") {
					want = append(want, line)
				}
			}
			if len(got) != len(want) {
				t.Fatalf("Go cases %d, C cases %d", len(got), len(want))
			}
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("case %d:\nGo %s\nC %s", i, got[i], want[i])
				}
			}
		})
	}
	printed, err := testoracle.Command(t, exe, "printer").CombinedOutput()
	if err != nil || string(printed) != "3->2 \n" {
		t.Fatalf("C filtered new-table printer: %v: %s", err, printed)
	}

	witness, err := testoracle.Command(t, exe, "false").CombinedOutput()
	if err != nil || string(witness) != "0 4\n" {
		t.Fatalf("original C false-conjunct witness: %v: %s", err, witness)
	}
}

func TestClassDuplicationBoundaries(t *testing.T) {
	for _, constant := range []bool{false, true} {
		src := Gia_ManStart(1)
		if constant {
			Gia_ManAppendCo(src, 1)
			src.nConstrs = 1
		}
		src.pReprsOld = make([]int32, src.nObjs)
		for i := range src.pReprsOld {
			src.pReprsOld[i] = -1
		}
		out := Gia_ManDupDfsClasses(src)
		if out.nObjs != src.nObjs || out.nConstrs != src.nConstrs || Gia_ManCiNum(out) != 0 || out.nRegs != 0 {
			t.Fatal("boundary class copy changed graph/roles")
		}
		var want []bool
		if constant {
			want = []bool{true}
		}
		checkQuantificationOutputs(t, out, nil, want)
	}
	t.Run("missing old table", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("missing representatives accepted")
			}
		}()
		Gia_ManDupDfsClasses(Gia_ManStart(1))
	})
}
