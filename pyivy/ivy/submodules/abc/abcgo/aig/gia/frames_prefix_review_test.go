package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// The source stops on nonzero hash literals, which may conservatively stop
// even for a semantically zero property. Independently require every completed
// frame to have zero properties, and the result to be its transition state.
func framesReviewCheckPrefix(t *testing.T, src, out *Gia_Man_t, bound int32) int {
	t.Helper()
	nPi := int(Gia_ManPiNum(src))
	if out == nil || out.nRegs != 0 || Gia_ManCoNum(out) != src.nRegs || nPi == 0 || int(Gia_ManCiNum(out))%nPi != 0 {
		t.Fatal("invalid specialized prefix interface")
	}
	attempts := int(Gia_ManCiNum(out)) / nPi
	if attempts < 1 || attempts > int(bound) {
		t.Fatalf("attempts%d outside bound%d", attempts, bound)
	}
	candidates := map[int]bool{attempts - 1: true}
	if attempts == int(bound) {
		candidates[attempts] = true
	}
	for assignment := 0; assignment < 1<<Gia_ManCiNum(out); assignment++ {
		state := make([]bool, src.nRegs)
		actual := cofReviewValues(out, assignment)
		for frame := 0; frame <= attempts; frame++ {
			if candidates[frame] {
				for i, want := range state {
					if actual[Gia_ManCo(out, int32(i)).id] != want {
						candidates[frame] = false
					}
				}
			}
			if frame == attempts {
				break
			}
			input := assignment >> (frame * nPi) & ((1 << nPi) - 1)
			for i, bit := range state {
				if bit {
					input |= 1 << (nPi + i)
				}
			}
			values := cofReviewValues(src, input)
			for po := int32(0); po < Gia_ManPoNum(src); po++ {
				if values[Gia_ManPo(src, po).id] {
					for prefix := range candidates {
						if frame < prefix {
							candidates[prefix] = false
						}
					}
				}
			}
			for i := range state {
				state[i] = values[Gia_ManRi(src, int32(i)).id]
			}
		}
	}
	for _, valid := range candidates {
		if valid {
			return 1 << Gia_ManCiNum(out)
		}
	}
	t.Fatal("prefix output is not state following a completed zero-property prefix")
	return 0
}

func TestFramesGeneratedNativeSpecialPrefix(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	native, err := exec.Command(exe, "--generated-special").CombinedOutput()
	if err != nil {
		t.Fatalf("C specialized prefix: %v\n%s", err, native)
	}
	var got strings.Builder
	rows := 0
	for seed := uint32(0); seed < 32; seed++ {
		for bound := int32(1); bound <= 3; bound++ {
			g := framesReviewGenerated(seed)
			var out *Gia_Man_t
			text := hashReviewStdout(t, func() { out = Gia_ManFramesInitSpecial(g, bound, 0) })
			got.WriteString(text)
			fmt.Fprintf(&got, "special %d f%d\n", seed, bound)
			framesReviewDump(&got, out, "result")
			rows += framesReviewCheckPrefix(t, g, out, bound)
			Gia_ManStop(out)
			Gia_ManStop(g)
		}
	}
	esopReviewCompare(t, native, got.String())
	t.Logf("96 unmodified-source prefixes and %d independent input histories", rows)
}

func TestFramesMixedSpecialPrefix(t *testing.T) {
	cases, rows := 0, 0
	for signs := 0; signs < 8; signs++ {
		for typed := 0; typed < 2; typed++ {
			for layout := 0; layout < 4; layout++ {
				for bound := int32(1); bound <= 3; bound++ {
					g := windowReviewGraph(signs, typed, layout)
					Gia_ManSetRegNum(g, 2)
					var out *Gia_Man_t
					hashReviewStdout(t, func() { out = Gia_ManFramesInitSpecial(g, bound, 0) })
					rows += framesReviewCheckPrefix(t, g, out, bound)
					Gia_ManStop(out)
					Gia_ManStop(g)
					cases++
				}
			}
		}
	}
	t.Logf("%d mixed-gate/layout prefixes and %d independent histories", cases, rows)
}
