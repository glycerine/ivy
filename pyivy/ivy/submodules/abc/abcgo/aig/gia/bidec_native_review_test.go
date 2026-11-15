package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func bidecReviewGraph(n, signs int) *Gia_Man_t {
	g := Gia_ManStart(128)
	ci := make([]int32, n)
	for i := range ci {
		ci[i] = Gia_ManAppendCi(g)
	}
	var root int32
	if n > 0 {
		root = ci[0] ^ int32(signs&1)
	}
	Gia_ManHashStart(g)
	for i := 1; i < n; i++ {
		lit := ci[i] ^ int32(signs>>uint(i%2)&1)
		switch i % 3 {
		case 0:
			root = Gia_ManHashXor(g, root, lit)
		case 1:
			root = Gia_ManHashAnd(g, root, lit)
		case 2:
			root = Gia_ManHashMux(g, lit, root, ci[i-1]^1)
		}
	}
	Gia_ManHashStop(g)
	Gia_ManAppendCo(g, root)
	return g
}
func bidecReviewRoot(g *Gia_Man_t, selector int) int32 {
	switch selector {
	case 0:
		return Gia_ObjFaninLit0p(g, Gia_ManCo(g, 0))
	case 1:
		return Gia_ObjFaninLit0p(g, Gia_ManCo(g, 0)) ^ 1
	case 2:
		if Gia_ManCiNum(g) > 0 {
			return Gia_ObjToLit(g, Gia_ManCi(g, 0))
		}
		return 0
	case 3:
		if Gia_ManCiNum(g) > 0 {
			return Gia_ObjToLit(g, Gia_ManCi(g, Gia_ManCiNum(g)-1)) ^ 1
		}
		return 1
	case 4:
		return 0
	default:
		return 1
	}
}
func TestBidecTruthMatchesOriginalABCThrough15Inputs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "bidec_truth.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native bidec truth: %v\n%s", err, native)
	}
	var got strings.Builder
	cases := 0
	for n := 0; n <= 15; n++ {
		for signs := 0; signs < 4; signs++ {
			for reverse := 0; reverse < 2; reverse++ {
				for selector := 0; selector < 6; selector++ {
					g := bidecReviewGraph(n, signs)
					leaves := Vec_IntAlloc(n)
					for _, id := range Vec_IntArray(g.vCis) {
						Vec_IntPush(leaves, id)
					}
					if reverse != 0 {
						slices.Reverse(Vec_IntArray(leaves))
					}
					storage, visited := Vec_IntAlloc(0), Vec_IntAlloc(0)
					truth := Gia_ManConvertAigToTruth(g, Gia_ObjFromLit(g, bidecReviewRoot(g, selector)), leaves, storage, visited)
					fmt.Fprintf(&got, "case %d %d %d %d", n, signs, reverse, selector)
					for _, word := range truth {
						fmt.Fprintf(&got, " %08x", word)
					}
					got.WriteByte('\n')
					for _, marker := range Vec_IntArray(g.vTruths) {
						if marker != -1 {
							t.Fatal("native comparison left truth scratch marked")
						}
					}
					cases++
				}
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(got.String(), "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("native truth line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("native truth lengths Go%d C%d", got.Len(), len(native))
	}
	t.Logf("%d native cut-truth cases across widths0..15, phases, leaf permutations, variable and constant roots", cases)
}

func TestBidecMappedResultsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "bidec_truth.c")
	native, err := exec.Command(exe, "resynthesis").CombinedOutput()
	if err != nil {
		t.Fatalf("native resynthesis: %v\n%s", err, native)
	}
	var got strings.Builder
	cases, rows := 0, 0
	for n := 2; n <= 8; n++ {
		for signs := 0; signs < 4; signs++ {
			g := bidecReviewGraph(n, signs)
			Gia_ManAppendCo(g, bidecReviewRoot(g, 1))
			Gia_ManAppendCo(g, Gia_ObjToLit(g, Gia_ManCi(g, 0))^int32(signs&1))
			Gia_ManSetRegNum(g, 1)
			bidecReviewInstallMapping(g)
			out := Gia_ManPerformBidec(g, 0)
			fmt.Fprintf(&got, "resynthesis %d %d %d %d %d", n, signs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
			for mask := 0; mask < 1<<n; mask++ {
				inputs := make([]bool, n)
				for i := range inputs {
					inputs[i] = mask>>i&1 != 0
				}
				want, err := g.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, out, inputs, want)
				values, err := out.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				packed := 0
				for i, value := range values {
					if value {
						packed |= 1 << i
					}
				}
				fmt.Fprintf(&got, " %d", packed)
				rows++
			}
			got.WriteByte('\n')
			cases++
		}
	}
	if got.String() != string(native) {
		t.Fatalf("mapped bidec interface/truth differs\nGo:\n%s\nC:\n%s", got.String(), native)
	}
	t.Logf("%d source resynthesis cases with sequential interface, %d truth rows; backend topology is not asserted", cases, rows)
}
