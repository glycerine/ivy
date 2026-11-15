package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestStateCycleNativeOracle(t *testing.T) {
	exe := compileGiaSourceOracle(t, "state_cycle.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v\n%s", err, native)
	}
	var got strings.Builder
	counts := [3]int{}
	for mode := 0; mode < 3; mode++ {
		for signs := 0; signs < 8; signs++ {
			for regs := int32(0); regs <= 5; regs++ {
				initials, traces, frameCount := 1<<regs, 5, int32(4)
				if mode == 0 {
					traces, frameCount = 1, 1
				}
				if mode == 2 {
					initials = 1
				}
				for initial := 0; initial < initials; initial++ {
					for trace := 0; trace < traces; trace++ {
						for frames := int32(0); frames < frameCount; frames++ {
							src := partitionReviewGraph(signs, false, false)
							src.nRegs = regs // Select the fixture's RI/RO suffix before running the API.
							cycleReviewSetInitial(src, initial)
							cex := cycleReviewCex(src, trace)
							fmt.Fprintf(&got, "case %d %d %d %d %d %d", mode, signs, regs, initial, trace, frames)
							var out *Gia_Man_t
							switch mode {
							case 0:
								out = Gia_ManDupFlip(src, []int32{int32(initial)})
							case 1:
								Gia_ManCycle(src, cex, frames)
							case 2:
								for _, obj := range src.pObjs {
									obj.fMark0 = 1
								}
								out = Gia_ManDupCycled(src, cex, frames)
							}
							if mode != 0 {
								fmt.Fprintf(&got, " random %d", Gia_ManRandom(0))
							}
							got.WriteString(" marks")
							for _, obj := range src.pObjs {
								fmt.Fprintf(&got, " %d", obj.fMark0|obj.fMark1<<1)
							}
							if out != nil {
								stateCycleNativeResult(t, &got, src, out)
							}
							got.WriteByte('\n')
							counts[mode]++
						}
					}
				}
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(string(native), "\n"), strings.Split(got.String(), "\n")
		for i := 0; i < len(a) && i < len(b); i++ {
			if a[i] != b[i] {
				t.Fatalf("line%d\nC: %s\nGo: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("line counts C%d Go%d", len(a), len(b))
	}
	t.Logf("native matches %d flips, %d simulations and %d cycled copies", counts[0], counts[1], counts[2])
}

func stateCycleNativeResult(t *testing.T, got *strings.Builder, src, out *Gia_Man_t) {
	t.Helper()
	fmt.Fprintf(got, " result %d %d values", out.nObjs, out.nRegs)
	for _, obj := range src.pObjs {
		fmt.Fprintf(got, " %d", obj.Value)
	}
	got.WriteString(" nodes")
	for _, obj := range out.pObjs {
		a, b, _, kind := out.Gate(obj.id)
		switch kind {
		case 0, 1:
			fmt.Fprintf(got, " %d", kind)
		case 2:
			fmt.Fprintf(got, " 2:%d:%d", a, b)
		case 6:
			fmt.Fprintf(got, " 6:%d", a)
		default:
			t.Fatalf("unexpected kind %d", kind)
		}
	}
	got.WriteString(" truth")
	for mask := 0; mask < 32; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
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
		fmt.Fprintf(got, " %d", packed)
	}
}
