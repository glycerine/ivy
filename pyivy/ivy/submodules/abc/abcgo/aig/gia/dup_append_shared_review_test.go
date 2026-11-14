package gia

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

func TestAppendSharedSameGraphReturns(t *testing.T) {
	const childKey = "ABCGO_SHARED_APPEND_REVIEW_CHILD"
	if os.Getenv(childKey) == "1" {
		// The red implementation continually allocates COs. Bound its lifetime
		// inside a child process so it cannot hang or exhaust the test runner.
		debug.SetMemoryLimit(16 << 20)
		timer := time.AfterFunc(2*time.Second, func() { panic("shared append did not return") })
		defer timer.Stop()
		g := Gia_ManStart(1)
		a := Gia_ManAppendCi(g)
		Gia_ManAppendCo(g, a^1)
		Gia_ManDupAppendShare(g, g)
		if Gia_ManCiNum(g) != 1 || Gia_ManCoNum(g) != 2 {
			t.Fatal("self append must duplicate the original outputs once")
		}
		checkQuantificationOutputs(t, g, []bool{false}, []bool{true, true})
		checkQuantificationOutputs(t, g, []bool{true}, []bool{false, false})
		for signs := int32(0); signs < 8; signs++ {
			for regs := int32(0); regs <= 2; regs++ {
				g := appendReviewGraph(signs, regs)
				var inputs, want [][]bool
				for mask := 0; mask < 1<<uint(2+regs); mask++ {
					row := make([]bool, 2+regs)
					for i := range row {
						row[i] = mask&(1<<uint(i)) != 0
					}
					values, err := g.Evaluate(row)
					if err != nil {
						t.Fatal(err)
					}
					inputs = append(inputs, row)
					want = append(want, append(values, values...))
				}
				Gia_ManDupAppendShare(g, g)
				if Gia_ManCiNum(g) != 2+regs || Gia_ManCoNum(g) != 2*(2+regs) || g.nRegs != regs {
					t.Fatal("typed self append changed interface")
				}
				for i, row := range inputs {
					checkQuantificationOutputs(t, g, row, want[i])
				}
			}
		}
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestAppendSharedSameGraphReturns$", "-test.count=1")
	cmd.Env = append(os.Environ(), childKey+"=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("shared self append: %v (context %v): %s", err, ctx.Err(), out)
	}
}

func appendSharedReviewRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for s1 := int32(0); s1 < 8; s1++ {
		for s2 := int32(0); s2 < 8; s2++ {
			for regs := int32(0); regs <= 2; regs++ {
				for mode := 0; mode < 3; mode++ {
					first, second := appendReviewGraph(s1, regs), appendReviewGraph(s2, regs)
					var out *Gia_Man_t
					wantRegs := regs
					wantInputs := 2 + regs
					if mode == 0 {
						out = Gia_ManDup2(first, second)
					} else {
						out = Gia_ManDup(first)
						if s1&1 != 0 {
							Gia_ManHashStart(out)
						}
						if mode == 1 {
							Gia_ManDupAppendShare(out, second)
						} else {
							Gia_ManDupAppend(out, second)
							wantInputs *= 2
							wantRegs = 0
							if out.nObjs != first.nObjs+second.nObjs-1 {
								t.Fatal("unhashed append changed the source object count")
							}
						}
					}
					// These entry points concatenate raw COs. Unlike AppendNew,
					// they do not merge independent transition relations.
					if Gia_ManCiNum(out) != wantInputs || Gia_ManCoNum(out) != 2*(2+regs) || out.nRegs != wantRegs {
						t.Fatalf("mode %d registers %d: wrong source interface", mode, regs)
					}
					if Gia_ManHashStarted(out) != giaBool(mode != 0) {
						t.Fatal("wrong final hash-table lifetime")
					}
					if mode != 0 {
						if name, spec := out.Identity(); name != "append" || spec != "append.aig" {
							t.Fatal("append changed destination identity")
						}
					}
					fmt.Fprintf(&records, "%d %d %d %d %d %d %d", s1, s2, regs, mode, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
					for mask := 0; mask < 1<<uint(wantInputs); mask++ {
						inputs := make([]bool, wantInputs)
						for i := range inputs {
							inputs[i] = mask&(1<<uint(i)) != 0
						}
						left := inputs[:2+regs]
						right := left
						if mode == 2 {
							right = inputs[2+regs:]
						}
						x, err := first.Evaluate(left)
						if err != nil {
							t.Fatal(err)
						}
						y, err := second.Evaluate(right)
						if err != nil {
							t.Fatal(err)
						}
						checkQuantificationOutputs(t, out, inputs, append(x, y...))
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
	}
	return records.String()
}

func TestAppendSharedTypedTruthTables(t *testing.T) {
	appendSharedReviewRecords(t)
}

func TestAppendSharedInterleavedInputsAndOutputs(t *testing.T) {
	graph := func(signs int32) *Gia_Man_t {
		g := Gia_ManStart(1)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		buffer := Gia_ManAppendBuf(g, b^((signs>>1)&1))
		Gia_ManAppendCo(g, a^1)
		c := Gia_ManAppendCi(g)
		x := Gia_ManAppendXorReal(g, a^(signs&1), b)
		g.EnableMuxes()
		Gia_ManAppendCo(g, Gia_ManAppendMuxReal(g, c^((signs>>2)&1), x, buffer))
		return g
	}
	for signs := int32(0); signs < 8; signs++ {
		for mode := 0; mode < 3; mode++ {
			first, second := graph(signs), graph(signs^7)
			var out *Gia_Man_t
			width := 3
			if mode == 0 {
				out = Gia_ManDup2(first, second)
			} else {
				out = Gia_ManDup(first)
				if mode == 1 {
					Gia_ManDupAppendShare(out, second)
				} else {
					Gia_ManDupAppend(out, second)
					width = 6
				}
			}
			for mask := 0; mask < 1<<uint(width); mask++ {
				inputs := make([]bool, width)
				for i := range inputs {
					inputs[i] = mask&(1<<uint(i)) != 0
				}
				x, err := first.Evaluate(inputs[:3])
				if err != nil {
					t.Fatal(err)
				}
				right := inputs[:3]
				if mode == 2 {
					right = inputs[3:]
				}
				y, err := second.Evaluate(right)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, out, inputs, append(x, y...))
			}
		}
	}
}

func TestAppendSharedMatchesOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "append_shared.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C shared append: %v: %s", err, want)
	}
	gotRows := strings.Split(strings.TrimSpace(appendSharedReviewRecords(t)), "\n")
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
