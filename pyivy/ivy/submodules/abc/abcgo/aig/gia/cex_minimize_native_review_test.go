package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"github.com/glycerine/abcgo/misc/util"
	"strings"
	"testing"
)

func cexReviewThreeInputGraph(table int) *Gia_Man_t {
	p := Gia_ManStart(64)
	Gia_ManHashStart(p)
	inputs := []int32{Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)}
	Gia_ManAppendCi(p)
	root := int32(0)
	for mask := 0; mask < 8; mask++ {
		if table>>mask&1 != 0 {
			cube := int32(1)
			for i, lit := range inputs {
				cube = Gia_ManHashAnd(p, cube, lit^int32(1^(mask>>i&1)))
			}
			root = Gia_ManHashOr(p, root, cube)
		}
	}
	Gia_ManAppendCo(p, root)
	Gia_ManAppendCo(p, 0)
	Gia_ManSetRegNum(p, 1)
	Gia_ManHashStop(p)
	return Gia_ManCleanup(p)
}
func TestCounterexampleCareWorkflowMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cex_minimize.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native care minimization: %v\n%s", err, native)
	}
	var expected, got strings.Builder
	for _, line := range strings.Split(string(native), "\n") {
		if strings.HasPrefix(line, "case ") {
			expected.WriteString(line)
			expected.WriteByte('\n')
		}
	}
	cases := 0
	for table := 1; table < 255; table++ {
		for assignment := 0; assignment < 8; assignment++ {
			if table>>assignment&1 != 0 {
				for effort := int32(0); effort < 2; effort++ {
					p := cexReviewThreeInputGraph(table)
					c := util.Abc_CexAlloc(1, 3, 1)
					c.PData[0] = uint32(assignment << 1)
					care := Bmc_CexCareSatBasedMinimizeAig(p, c, effort, 0)
					fmt.Fprintf(&got, "case %x %d %d mask%x\n", table, assignment, effort, care.PData[0])
					for other := 0; other < 8; other++ {
						if (uint32(other<<1)^c.PData[0])&care.PData[0] == 0 && table>>other&1 == 0 {
							t.Fatalf("care %x does not force table%x at %d", care.PData[0], table, assignment)
						}
					}
					cases++
				}
			}
		}
	}
	if got.String() != expected.String() {
		a, b := strings.Split(got.String(), "\n"), strings.Split(expected.String(), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("care workflow source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("native output length Go%d C%d", got.Len(), expected.Len())
	}
	t.Logf("%d exact native care masks for both effort modes, independently checked for sufficiency", cases)
}

func TestCounterexampleSequentialCareMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cex_care_sequence.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native sequential care: %v\n%s", err, native)
	}
	var expected, got strings.Builder
	for _, line := range strings.Split(string(native), "\n") {
		if strings.HasPrefix(line, "case ") || strings.HasPrefix(line, "forced-true ") {
			expected.WriteString(line)
			expected.WriteByte('\n')
		}
	}
	cases, forced := 0, 0
	for transition := 0; transition < 16; transition++ {
		for _, property := range []int{8, 6, 14, 2} {
			for _, frames := range []int{2, 5, 37} {
				for seed := 0; seed < 4; seed++ {
					p := Gia_ManStart(32)
					Gia_ManHashStart(p)
					x, s := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
					Gia_ManAppendCo(p, cexReviewFunction(p, property, x, s))
					Gia_ManAppendCo(p, cexReviewFunction(p, transition, x, s))
					Gia_ManSetRegNum(p, 1)
					Gia_ManHashStop(p)
					p = Gia_ManCleanup(p)
					c := cexReviewWitness(frames, seed, 0)
					if Gia_ManVerifyCex(p, c, 0) == 0 {
						continue
					}
					preflight, _, _, negated := giaCexCareSATInstance(p, c)
					preflight.Close()
					if !negated {
						fmt.Fprintf(&got, "forced-true %d %d %d %d\n", transition, property, frames, seed)
						forced++
					}
					for effort := int32(0); effort < 2; effort++ {
						care := Bmc_CexCareSatBasedMinimizeAig(p, c, effort, 0)
						if negated {
							fmt.Fprintf(&got, "case %d %d %d %d %d mask", transition, property, frames, seed, effort)
							cexReviewWords(&got, care.PData)
							got.WriteByte('\n')
						} else if util.Abc_CexCountOnes(care) != 0 {
							t.Fatal("forced-true unrolling retained input bits")
						}

						completions := 32
						if frames < 5 {
							completions = 1 << frames
						}
						for other := 0; other < completions; other++ {
							state, failed := 0, false
							random := uint32(other + 1)
							for f := 0; f < frames; f++ {
								bit := 1 + f
								input := 0
								if frames <= 5 {
									input = other >> f & 1
								} else {
									random = random*1664525 + 1013904223
									input = int(random >> 31)
								}
								if care.PData[bit>>5]>>uint(bit&31)&1 != 0 {
									input = int(c.PData[bit>>5] >> uint(bit&31) & 1)
								}
								index := input + 2*state
								failed = property>>index&1 != 0
								state = transition >> index & 1
							}
							if !failed {
								t.Fatalf("sequential mask does not force failure: transition%d property%d frames%d", transition, property, frames)
							}
						}
						cases++
					}
				}
			}
		}
	}
	if got.String() != expected.String() {
		a, b := strings.Split(got.String(), "\n"), strings.Split(expected.String(), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("sequential care source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("native output length Go%d C%d", got.Len(), expected.Len())
	}
	t.Logf("%d sequential care checks, including %d source forced-true boundaries, with exhaustive small and sampled 37-frame completions", cases, forced)
}
