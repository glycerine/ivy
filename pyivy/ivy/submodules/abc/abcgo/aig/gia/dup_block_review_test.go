package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func blockReviewGraph(signs, registers int32) *Gia_Man_t {
	g := Gia_ManStart(1)
	var ci [6]int32
	for i := range ci {
		ci[i] = Gia_ManAppendCi(g)
	}
	g.EnableMuxes()
	x := Gia_ManAppendXorReal(g, ci[0]^(signs&1), ci[2])
	b := Gia_ManAppendBuf(g, x^((signs>>1)&1))
	m := Gia_ManAppendMuxReal(g, ci[4]^((signs>>2)&1), b, ci[1])
	// Interleave COs and logic to exercise the supported Go object ordering.
	Gia_ManAppendCo(g, m)
	y := Gia_ManAppendAnd(g, m, ci[3]^1)
	for _, lit := range []int32{x, b ^ 1, y, ci[5], 1} {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, registers)
	g.SetIdentity("block", "block.aig")
	return g
}

func TestBlockReductionTypedTruthAndRegisterLayout(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		for _, block := range []int32{1, 2, 3, 6} {
			for _, registers := range []int32{0, block, 6} {
				g := blockReviewGraph(signs, registers)
				out := Gia_ManDupBlock(g, block)
				if Gia_ManCiNum(out) != 6/block || Gia_ManCoNum(out) != 6/block || out.nRegs != registers/block || Gia_ManPiNum(out) != (6-registers)/block || Gia_ManPoNum(out) != (6-registers)/block {
					t.Fatalf("block %d registers %d: wrong reduced interface", block, registers)
				}
				if name, spec := out.Identity(); name != "block" || spec != "block.aig" {
					t.Fatal("lost block source identity")
				}
				for mask := 0; mask < 1<<uint(6/block); mask++ {
					inputs := make([]bool, 6/block)
					original := make([]bool, 6)
					for i := range inputs {
						inputs[i] = mask&(1<<uint(i)) != 0
						original[int32(i)*block] = inputs[i]
					}
					values, err := g.Evaluate(original)
					if err != nil {
						t.Fatal(err)
					}
					want := make([]bool, len(inputs))
					for i := range want {
						want[i] = values[int32(i)*block]
					}
					checkQuantificationOutputs(t, out, inputs, want)
				}
			}
		}
	}
}

func TestBlockReductionEmptyAndInvalidInterfaces(t *testing.T) {
	for _, block := range []int32{1, 3} {
		out := Gia_ManDupBlock(Gia_ManStart(1), block)
		if out.nObjs != 1 || Gia_ManCiNum(out) != 0 || Gia_ManCoNum(out) != 0 || out.nRegs != 0 {
			t.Fatal("empty block reduction added objects or terminals")
		}
	}
	for _, block := range []int32{0, -1, 4} {
		t.Run(fmt.Sprintf("block%d", block), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid block size accepted")
				}
			}()
			Gia_ManDupBlock(blockReviewGraph(0, 0), block)
		})
	}
	t.Run("output_count", func(t *testing.T) {
		g := Gia_ManStart(1)
		Gia_ManAppendCi(g)
		Gia_ManAppendCi(g)
		Gia_ManAppendCo(g, 0)
		defer func() {
			if recover() == nil {
				t.Fatal("indivisible output interface accepted")
			}
		}()
		Gia_ManDupBlock(g, 2)
	})
}

func TestBlockReductionMatchesOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "block_reduction.c")
	want, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C block reduction: %v: %s", err, want)
	}
	var records strings.Builder
	for signs := int32(0); signs < 8; signs++ {
		for _, block := range []int32{1, 2, 3, 6} {
			for _, registers := range []int32{0, block, 6} {
				out := Gia_ManDupBlock(blockReviewGraph(signs, registers), block)
				fmt.Fprintf(&records, "%d %d %d %d %d %d", signs, block, registers, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
				for mask := 0; mask < 1<<uint(6/block); mask++ {
					inputs := make([]bool, 6/block)
					for i := range inputs {
						inputs[i] = mask&(1<<uint(i)) != 0
					}
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
	gotRows := strings.Split(strings.TrimSpace(records.String()), "\n")
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
