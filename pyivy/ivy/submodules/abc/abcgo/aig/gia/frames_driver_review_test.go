package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func framesReviewLiveInitGraph() *Gia_Man_t {
	g := framesReviewDependencyGraph()
	ri := Gia_ManRi(g, 0)
	ri.iDiff0, ri.fCompl0 = uint32(ri.id-Gia_ManPi(g, 1).id), 0
	return g
}

func TestFramesOrOutputsAcrossFrames(t *testing.T) {
	g := framesReviewLiveInitGraph()
	defer Gia_ManStop(g)
	out := Gia_ManFrames(g, &Gia_ParFra_t{NFrames: 3, FOrPos: 1})
	defer Gia_ManStop(out)
	if out.nRegs != 3 || Gia_ManCoNum(out) != 5 {
		t.Fatalf("OR mode regs%d co%d, want3/5", out.nRegs, Gia_ManCoNum(out))
	}
}

func TestFramesInitSpecialEventPrefix(t *testing.T) {
	for _, frames := range []int32{0, 1, 4} {
		t.Run(fmt.Sprint(frames), func(t *testing.T) {
			g := Gia_ManStart(8)
			defer Gia_ManStop(g)
			r := Gia_ManAppendCi(g)
			Gia_ManAppendCo(g, r)
			Gia_ManAppendCo(g, r^1)
			Gia_ManSetRegNum(g, 1)
			var out *Gia_Man_t
			hashReviewStdout(t, func() { out = Gia_ManFramesInitSpecial(g, frames, 0) })
			if out == nil {
				t.Fatal("event-driven initialized prefix returned nil")
			}
			defer Gia_ManStop(out)
			values, err := out.Evaluate(nil)
			if err != nil || len(values) != 1 || !values[0] || out.nRegs != 0 {
				t.Fatalf("prefix must return preceding state1, got%v regs%d error%v", values, out.nRegs, err)
			}
		})
	}
}

func TestFramesNativeDrivers(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	native, err := testoracle.Command(t, exe, "--drivers-safe").CombinedOutput()
	if err != nil {
		t.Fatalf("C frame drivers: %v\n%s", err, native)
	}
	var got strings.Builder
	for frames := int32(1); frames <= 3; frames++ {
		for init := int32(0); init < 2; init++ {
			for disable := int32(0); disable < 2; disable++ {
				for or := int32(0); or < 2; or++ {
					g := framesReviewLiveInitGraph()
					p := Gia_ParFra_t{NFrames: frames, FInit: init, FDisableSt: disable, FOrPos: or}
					out := Gia_ManFrames(g, &p)
					fmt.Fprintf(&got, "driver f%d init%d disable%d or%d\n", frames, init, disable, or)
					fmt.Fprintf(&got, "result ci%d co%d regs%d objs%d\n", Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs, out.nObjs)
					for id, o := range out.pObjs {
						fmt.Fprintf(&got, "%d %d %d %d %d %d %d\n", id, o.iDiff0, o.iDiff1, o.fCompl0, o.fCompl1, o.fTerm, o.Value)
					}
					Gia_ManStop(out)
					Gia_ManStop(g)
				}
			}
		}
	}
	esopReviewCompare(t, native, got.String())
}

func TestFramesOriginalInitTraversalDefect(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	native, err := testoracle.Command(t, exe, "--init-defect").CombinedOutput()
	if err != nil {
		t.Fatalf("C initialized traversal: %v\n%s", err, native)
	}
	rows := strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(rows) != 128 {
		t.Fatalf("want128 source/corrected records, got%d", len(rows))
	}
	wrong := 0
	g := framesReviewLiveInitGraph()
	defer Gia_ManStop(g)
	for index, row := range rows {
		var mode, assignment, actual int
		if n, err := fmt.Sscanf(row, "init %d %d %d", &mode, &assignment, &actual); n != 3 || err != nil || mode != index/64 || assignment != index%64 {
			t.Fatalf("invalid native row%q", row)
		}
		state, expected := []bool{false, false, false}, 0
		for frame := 0; frame < 3; frame++ {
			in := []bool{assignment>>(frame*2)&1 != 0, assignment>>(frame*2+1)&1 != 0, state[0], state[1], state[2]}
			out, err := g.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			for po := 0; po < 2; po++ {
				if out[po] {
					expected |= 1 << (frame*2 + po)
				}
			}
			copy(state, out[2:])
		}
		if mode == 0 && actual != expected {
			wrong++
		} else if mode == 1 && actual != expected {
			t.Fatalf("corrected source row%d gives%d want%d", assignment, actual, expected)
		}
	}
	if wrong == 0 {
		t.Fatal("unmodified C did not reproduce traversal-epoch defect")
	}
}

func TestFramesInitSpecialIgnoresStaleScratch(t *testing.T) {
	g := Gia_ManStart(8)
	defer Gia_ManStop(g)
	r := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, 1)
	Gia_ManAppendCo(g, r)
	Gia_ManSetRegNum(g, 1)
	Gia_ManFillValue(g)
	var out *Gia_Man_t
	hashReviewStdout(t, func() { out = Gia_ManFramesInitSpecial(g, 1, 0) })
	if out == nil {
		t.Fatal("no prefix graph")
	}
	defer Gia_ManStop(out)
	values, err := out.Evaluate(nil)
	if err != nil || len(values) != 1 || values[0] {
		t.Fatalf("immediate event must retain initial state0, got%v error%v", values, err)
	}
}

func TestFramesOriginalSpecialStaleScratchFailure(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	cmd := testoracle.Command(t, exe, "--special-stale")
	cmd.Dir = t.TempDir()
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "Assertion") {
		t.Fatalf("expected original C stale-RI assertion, got%v\n%s", err, out)
	}
}
