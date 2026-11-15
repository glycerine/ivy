package gia

import (
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func collapseReviewBDD(dd *cudd.DdManager, n, table int) *cudd.DdNode {
	sum := dd.LogicZero
	for mask := 0; mask < 1<<n; mask++ {
		if table>>mask&1 == 0 {
			continue
		}
		cube := dd.One
		for i := 0; i < n; i++ {
			lit := cudd.Cudd_NotCond(cudd.Cudd_bddIthVar(dd, i), mask>>i&1 == 0)
			cube = cudd.Cudd_bddAnd(dd, cube, lit)
		}
		sum = cudd.Cudd_bddOr(dd, sum, cube)
	}
	return sum
}
func TestCollapseIsopCoversAndAigsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "collapse_isop.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native ISOP factoring: %v\n%s", err, native)
	}
	var got strings.Builder
	cases := 0
	for n := 2; n <= 3; n++ {
		for table := 1; table < (1<<(1<<n))-1; table++ {
			for order := 0; order < 3; order++ {
				dd := cudd.Cudd_Init(uint32(n), 0, 0, 0, 0)
				sequence := make([]int, n)
				for i := range sequence {
					switch order {
					case 0:
						sequence[i] = i
					case 1:
						sequence[i] = n - 1 - i
					case 2:
						sequence[i] = (i + 1) % n
					}
				}
				if cudd.Cudd_ShuffleHeap(dd, sequence) == 0 || cudd.Cudd_zddVarsFromBddVars(dd, 2) == 0 {
					t.Fatal("failed to configure variable order")
				}
				f := collapseReviewBDD(dd, n, table)
				cudd.Cudd_Ref(f)
				for phase := 0; phase < 1<<n; phase++ {
					g := Gia_ManStart(64)
					leaves := Vec_IntAlloc(n)
					for i := 0; i < n; i++ {
						Vec_IntPush(leaves, Gia_ManAppendCi(g)^int32(phase>>i&1))
					}
					Gia_ManHashStart(g)
					sop, cube := Vec_StrAlloc(0), Vec_StrAlloc(0)
					root := Gia_ManRebuildIsop(dd, f, g, leaves, sop, cube)
					if root < 0 {
						t.Fatalf("failed ISOP n%d table%x order%d phase%d", n, table, order, phase)
					}
					cover := Vec_StrArray(sop)
					if len(cover) == 0 || cover[len(cover)-1] != 0 {
						t.Fatal("selected SOP is not terminated")
					}
					fmt.Fprintf(&got, "case %d %x %d %d %d sop%s nodes", n, table, order, phase, root, hex.EncodeToString(cover[:len(cover)-1]))
					for _, obj := range g.pObjs {
						if Gia_ObjIsAnd(obj) != 0 {
							a, b, _, _ := g.Gate(obj.id)
							fmt.Fprintf(&got, " %d:%d", a, b)
						}
					}
					Gia_ManAppendCo(g, root)
					packed := 0
					for mask := 0; mask < 1<<n; mask++ {
						inputs := make([]bool, n)
						for i := range inputs {
							inputs[i] = mask>>i&1 != 0
						}
						out, err := g.Evaluate(inputs)
						if err != nil {
							t.Fatal(err)
						}
						expected := table>>(mask^phase)&1 != 0
						if out[0] != expected {
							t.Fatalf("n%d table%x order%d phase%d assignment%d: truth%v want%v", n, table, order, phase, mask, out[0], expected)
						}
						if out[0] {
							packed |= 1 << mask
						}
					}
					fmt.Fprintf(&got, " truth %x\n", packed)
					cases++
				}
				cudd.Cudd_RecursiveDeref(dd, f)
				cudd.Cudd_Quit(dd)
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(got.String(), "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("ISOP/factoring source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("native output length Go%d C%d", got.Len(), len(native))
	}
	t.Logf("%d exact native SOP, phase, root literal and AIG structure cases, with independent truth under all small inputs", cases)
}
