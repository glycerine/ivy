package gia

import (
	"context"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/glycerine/abcgo/sat/bsat"
)

func ccofReviewStateSource(regs int) *Gia_Man_t {
	g := Gia_ManStart(16)
	for i := 0; i < regs; i++ {
		Gia_ManAppendCi(g)
	}
	for i := 0; i < regs; i++ {
		Gia_ManAppendCo(g, Gia_ObjToLit(g, Gia_ManCi(g, int32(i))))
	}
	Gia_ManSetRegNum(g, int32(regs))
	return g
}
func ccofReviewFrames(regs, pis, mode, phase int) (*Gia_Man_t, int32) {
	g := Gia_ManStart(128)
	ci := make([]int32, regs+pis)
	for i := range ci {
		ci[i] = Gia_ManAppendCi(g)
	}
	Gia_ManHashStart(g)
	var state int32
	if mode == 0 || mode == 3 {
		state = 1
	}
	for i := 0; i < regs; i++ {
		lit := ci[i] ^ int32(phase>>i&1)
		if mode == 2 {
			term := Gia_ManHashAnd(g, lit, ci[regs+i%pis]^int32(phase>>(i+1)&1))
			state = Gia_ManHashOr(g, state, term)
		} else if mode == 0 || mode == 3 {
			state = Gia_ManHashAnd(g, state, lit)
		} else {
			state = Gia_ManHashOr(g, state, lit)
		}
	}
	if mode != 2 {
		lit := ci[regs] ^ int32(phase&1)
		if mode == 3 && pis > 1 {
			lit = Gia_ManHashXor(g, lit, ci[regs+1])
		}
		state = Gia_ManHashAnd(g, state, lit)
	}
	return g, state
}
func TestCircuitCofactoringImagesMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_cofactor.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native cofactoring: %v\n%s", err, native)
	}
	var got strings.Builder
	cases := 0
	for regs := 1; regs <= 3; regs++ {
		for pis := 1; pis <= 2; pis++ {
			for mode := 0; mode < 4; mode++ {
				for phase := 0; phase < 8; phase++ {
					func() {
						g, property := ccofReviewFrames(regs, pis, mode, phase)
						p := &Ccf_Man_t{pGia: ccofReviewStateSource(regs), pFrames: g, vCopies: Vec_IntAlloc(0), pSat: bsat.NewSolver()}
						defer Ccf_ManStop(p)
						Gia_ManCofExtendSolver(p)
						status := Gia_ManCofGetReachable(p, property)
						again := Gia_ManCofGetReachable(p, property)
						if status != 0 || again != 1 {
							t.Fatalf("initial/repeated enumeration statuses %d/%d", status, again)
						}
						image := 0
						for state := 0; state < 1<<regs; state++ {
							inputs := make([]bool, regs+pis)
							for i := 0; i < regs; i++ {
								inputs[i] = state>>i&1 != 0
							}
							outputs, err := g.Evaluate(inputs)
							if err != nil {
								t.Fatal(err)
							}
							any := false
							for _, output := range outputs {
								any = any || output
							}
							if any {
								image |= 1 << state
							}
							expected := false
							for piMask := 0; piMask < 1<<pis; piMask++ {
								trial := append([]bool(nil), inputs...)
								for i := 0; i < pis; i++ {
									trial[regs+i] = piMask>>i&1 != 0
								}
								values, _, err := g.evaluateObjects(trial)
								if err != nil {
									t.Fatal(err)
								}
								expected = expected || (values[property>>1] != (property&1 != 0))
							}
							if any != expected {
								t.Fatalf("state%d: image%v want%v", state, any, expected)
							}
						}
						fmt.Fprintf(&got, "case %d %d %d %d %d %d %x\n", regs, pis, mode, phase, status, again, image)
						cases++
					}()
				}
			}
		}
	}
	if got.String() != string(native) {
		t.Fatalf("source circuit-cofactoring status/image differs\nGo:\n%s\nC:\n%s", got.String(), native)
	}
	t.Logf("%d proper-image enumeration/convergence cases agree with original ABC and independent existential evaluation", cases)
}

func TestCircuitCofactoringVerboseMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_cofactor.c")
	native, err := testoracle.Command(t, exe, "diagnostic").CombinedOutput()
	if err != nil {
		t.Fatalf("native diagnostic: %v\n%s", err, native)
	}
	got := hashReviewStdout(t, func() {
		g, property := ccofReviewFrames(2, 1, 0, 0)
		p := &Ccf_Man_t{pGia: ccofReviewStateSource(2), pFrames: g, vCopies: Vec_IntAlloc(0), pSat: bsat.NewSolver(), fVerbose: 1}
		defer Ccf_ManStop(p)
		Gia_ManCofExtendSolver(p)
		fmt.Printf("status %d\n", Gia_ManCofGetReachable(p, property))
	})
	elapsed := regexp.MustCompile(`Time\s*=\s*[0-9.]+ sec`)
	if a, b := elapsed.ReplaceAllString(got, "Time=<elapsed>"), elapsed.ReplaceAllString(string(native), "Time=<elapsed>"); a != b {
		t.Fatalf("verbose cofactoring differs\nGo:%s\nC: %s", a, b)
	}
}

func TestCircuitCofactoringOriginalAcceptsLargeIncrementalBound(t *testing.T) {
	exe := compileGiaSourceOracle(t, "circuit_cofactor.c")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	output, err := testoracle.CommandContext(t, ctx, exe, "large").CombinedOutput()
	if err != nil {
		t.Fatalf("source large-bound cofactoring: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "Property holds.") || !strings.Contains(string(output), "large-result 0") {
		t.Fatalf("source did not converge on constant-safe property: %s", output)
	}
}
