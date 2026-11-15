package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"os/exec"
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
	native, err := exec.Command(exe).CombinedOutput()
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
