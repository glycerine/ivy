package gia

import (
	"context"
	"fmt"
	"math/bits"
	"os"
	"os/exec"
	"runtime/debug"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestPermutationInputGroupOverflowRejected(t *testing.T) {
	const key = "ABCGO_PERMUTATION_OVERFLOW_REVIEW_CHILD"
	if os.Getenv(key) == "1" {
		debug.SetMemoryLimit(16 << 20)
		timer := time.AfterFunc(2*time.Second, func() { panic("input-group validation did not return") })
		defer timer.Stop()
		defer func() {
			if recover() == nil {
				t.Fatal("overflowing input groups accepted")
			}
		}()
		const largest = int32(1<<31 - 1)
		Gia_ManPermuteInputs(blockReviewGraph(0, 3), largest, largest)
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestPermutationInputGroupOverflowRejected$", "-test.count=1")
	cmd.Env = append(os.Environ(), key+"=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("input-group overflow: %v (context %v): %s", err, ctx.Err(), out)
	}
}

type permutationReviewCase struct {
	mode, p1, p2 int
	order        []int32
}

func permutationReviewCases() []permutationReviewCase {
	perms := [][]int32{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	var cases []permutationReviewCase
	for i, perm := range perms {
		cases = append(cases, permutationReviewCase{0, i, 0, perm}, permutationReviewCase{1, i, 0, perm})
	}
	for pattern := 0; pattern < 32; pattern++ {
		if bits.OnesCount(uint(pattern)) != 3 {
			continue
		}
		spread := make([]int32, 5)
		for i := range spread {
			spread[i] = -1
			if pattern&(1<<uint(i)) != 0 {
				spread[i] = 99 - int32(i) // Spread treats every non--1 entry as a slot.
				if i%2 == 0 {
					spread[i] = -2 - int32(i)
				}
			}
		}
		cases = append(cases, permutationReviewCase{3, pattern, 0, spread})
		for p, perm := range perms {
			mask := slices.Clone(spread)
			index := 0
			for i := range mask {
				if mask[i] != -1 {
					mask[i] = perm[index]
					index++
				}
			}
			cases = append(cases, permutationReviewCase{2, pattern, p, mask})
		}
	}
	for ppis := 0; ppis <= 3; ppis++ {
		for extra := 0; ppis+extra <= 3; extra++ {
			var order []int32
			for i := 0; i < 3-ppis-extra; i++ {
				order = append(order, int32(i))
			}
			for i := 3 - extra; i < 3; i++ {
				order = append(order, int32(i))
			}
			for i := 3 - ppis - extra; i < 3-extra; i++ {
				order = append(order, int32(i))
			}
			cases = append(cases, permutationReviewCase{4, ppis, extra, order})
		}
	}
	return append(cases, permutationReviewCase{mode: 5})
}

func permutationReviewRecords(t *testing.T, dfs bool) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 8; signs++ {
		for _, c := range permutationReviewCases() {
			src := blockReviewGraph(signs, 3)
			if dfs {
				src = Gia_ManDupOrderDfs(src)
			}
			var out *Gia_Man_t
			switch c.mode {
			case 0:
				out = Gia_ManDupPerm(src, permutation(c.order...))
			case 1:
				out = Gia_ManDupPermFlop(src, permutation(c.order...))
			case 2:
				out = Gia_ManDupPermFlopGap(src, permutation(c.order...))
			case 3:
				out = Gia_ManDupSpreadFlop(src, permutation(c.order...))
			case 4:
				out = Gia_ManPermuteInputs(src, int32(c.p1), int32(c.p2))
			case 5:
				out = Gia_ManDupZero(src)
			}
			registers := int32(3)
			if c.mode == 2 || c.mode == 3 {
				registers = 5
			}
			if Gia_ManPiNum(out) != 3 || Gia_ManPoNum(out) != 3 || out.nRegs != registers || out.nConstrs != 0 {
				t.Fatalf("mode %d %d/%d: incorrect interface", c.mode, c.p1, c.p2)
			}
			if c.mode != 5 {
				if out.nObjs != src.nObjs+2*(registers-3) || out.nMuxes != src.nMuxes || out.nXors != src.nXors || out.nBufs != src.nBufs {
					t.Fatal("permutation changed source gate kinds or counts")
				}
			} else if out.nObjs != 13 || Gia_ManAndNum(out) != 0 {
				t.Fatal("zero duplication retained unnecessary logic")
			}
			name, spec := out.Identity()
			if name != "block" || c.mode != 5 && spec != "block.aig" || c.mode == 5 && spec != "" {
				t.Fatal("incorrect source identity convention")
			}
			fmt.Fprintf(&records, "%d %d %d %d %d %d %d", c.mode, signs, c.p1, c.p2, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
			for mask := 0; mask < 1<<uint(Gia_ManCiNum(out)); mask++ {
				inputs := make([]bool, Gia_ManCiNum(out))
				for i := range inputs {
					inputs[i] = mask&(1<<uint(i)) != 0
				}
				original := slices.Clone(inputs[:6])
				roOrder := []int32{0, 1, 2}
				switch c.mode {
				case 0, 4:
					for i, old := range c.order {
						original[old] = inputs[i]
					}
				case 1:
					roOrder = c.order
				case 2, 3:
					roOrder = slices.Clone(c.order)
					if c.mode == 3 {
						old := int32(0)
						for i := range roOrder {
							if roOrder[i] != -1 {
								roOrder[i] = old
								old++
							}
						}
					}
				}
				for i, old := range roOrder {
					if old != -1 {
						original[3+old] = inputs[3+i]
					}
				}
				values, err := src.Evaluate(original)
				if err != nil {
					t.Fatal(err)
				}
				want := slices.Clone(values[:3])
				for _, old := range roOrder {
					next := false
					if old != -1 {
						next = values[3+old]
					}
					want = append(want, next)
				}
				if c.mode == 5 {
					clear(want)
				}
				checkQuantificationOutputs(t, out, inputs, want)
				got, err := out.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				packed := 0
				for i, value := range got {
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

func TestPermutationAndRegisterGapsTypedTransitions(t *testing.T) {
	permutationReviewRecords(t, false)
}

func TestPermutationAndRegisterGapsDfsLayouts(t *testing.T) {
	permutationReviewRecords(t, true)
}

func TestPermutationAndGapEmptyInterfaces(t *testing.T) {
	src := Gia_ManStart(1)
	if Gia_ManDupPerm(src, permutation()).nObjs != 1 || Gia_ManDupPermFlop(src, permutation()).nObjs != 1 {
		t.Fatal("empty permutation added objects")
	}
	for _, out := range []*Gia_Man_t{
		Gia_ManDupSpreadFlop(src, permutation(-1, -1)),
		Gia_ManDupPermFlopGap(src, permutation(-1, -1)),
	} {
		if Gia_ManPiNum(out) != 0 || Gia_ManPoNum(out) != 0 || out.nRegs != 2 || out.nObjs != 5 {
			t.Fatal("gap-only state has wrong interface")
		}
		for mask := 0; mask < 4; mask++ {
			checkQuantificationOutputs(t, out, []bool{mask&1 != 0, mask&2 != 0}, []bool{false, false})
		}
	}
}

func TestPermutationAndGapInvalidInterfaces(t *testing.T) {
	for _, bad := range [][]int32{{0, 1}, {0, 0, 2}, {-1, 1, 2}, {0, 1, 3}} {
		for mode := 0; mode < 3; mode++ {
			t.Run(fmt.Sprintf("mode%d_%v", mode, bad), func(t *testing.T) {
				defer func() {
					if recover() == nil {
						t.Fatal("invalid permutation accepted")
					}
				}()
				src := blockReviewGraph(0, 3)
				switch mode {
				case 0:
					Gia_ManDupPerm(src, permutation(bad...))
				case 1:
					Gia_ManDupPermFlop(src, permutation(bad...))
				case 2:
					Gia_ManDupPermFlopGap(src, permutation(bad...))
				}
			})
		}
	}
	for _, groups := range [][2]int32{{-1, 0}, {0, -1}, {4, 0}, {2, 2}} {
		t.Run(fmt.Sprintf("groups%v", groups), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid input groups accepted")
				}
			}()
			Gia_ManPermuteInputs(blockReviewGraph(0, 3), groups[0], groups[1])
		})
	}
}

func TestPermutationAndDirectWireOutputs(t *testing.T) {
	for _, c := range permutationReviewCases() {
		if c.mode != 0 {
			continue
		}
		src := Gia_ManStart(1)
		for i := 0; i < 3; i++ {
			Gia_ManAppendCi(src)
		}
		for i := int32(0); i < 3; i++ {
			Gia_ManAppendCo(src, Gia_ManCiLit(src, i))
		}
		out := Gia_ManDupPerm(src, permutation(c.order...))
		for mask := 0; mask < 8; mask++ {
			inputs, want := make([]bool, 3), make([]bool, 3)
			for i, old := range c.order {
				inputs[i] = mask&(1<<uint(i)) != 0
				want[old] = inputs[i]
			}
			checkQuantificationOutputs(t, out, inputs, want)
		}
	}
}

func TestPermutationAndGapsMatchOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "permutation_gaps.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C permutation/gaps: %v: %s", err, want)
	}
	gotRows := strings.Split(strings.TrimSpace(permutationReviewRecords(t, false)), "\n")
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
