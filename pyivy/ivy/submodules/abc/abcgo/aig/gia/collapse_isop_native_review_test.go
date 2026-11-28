package gia

import (
	"encoding/hex"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
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
	native, err := testoracle.Command(t, exe).CombinedOutput()
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

func TestCollapseRebuildMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "collapse_rebuild.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native collapse/rebuild: %v\n%s", err, native)
	}
	var got strings.Builder
	for table := 0; table < 256; table++ {
		p := Gia_ManStart(128)
		Gia_ManHashStart(p)
		inputs := []int32{Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)}
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
		for _, lit := range []int32{root, root ^ 1, inputs[0], 1} {
			Gia_ManAppendCo(p, lit)
		}
		Gia_ManHashStop(p)
		clean := Gia_ManCleanup(p)
		out := Gia_ManCollapseTest(clean, 0)
		fmt.Fprintf(&got, "table %x objects %d ands", table, Gia_ManObjNum(out))
		for _, obj := range out.pObjs {
			if Gia_ObjIsAnd(obj) != 0 {
				a, b, _, _ := out.Gate(obj.id)
				fmt.Fprintf(&got, " %d:%d", a, b)
			}
		}
		got.WriteString(" outputs")
		for i := int32(0); i < Gia_ManCoNum(out); i++ {
			a, _, _, _ := out.Gate(Gia_ObjId(out, Gia_ManCo(out, i)))
			fmt.Fprintf(&got, " %d", a)
		}
		got.WriteByte('\n')
		for mask := 0; mask < 8; mask++ {
			checkQuantificationOutputs(t, out, []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}, []bool{table>>mask&1 != 0, table>>mask&1 == 0, mask&1 != 0, true})
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(got.String(), "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("collapse/rebuild source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("native output length Go%d C%d", got.Len(), len(native))
	}
	t.Log("all 256 three-input functions match exact native rebuilt graph and output literals")
}

func TestCollapseWideIsopMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "collapse_isop_wide.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native wide ISOP: %v\n%s", err, native)
	}
	var got strings.Builder
	cases := 0
	for _, n := range []int{4, 17, 33, 65} {
		for _, table := range []int{6, 0x81, 0xe8, 0xca, 0x96, 0x7e} {
			for order := 0; order < 2; order++ {
				dd := cudd.Cudd_Init(uint32(n), 0, 0, 0, 0)
				sequence := make([]int, n)
				for i := range sequence {
					sequence[i] = i
					if order != 0 {
						sequence[i] = n - 1 - i
					}
				}
				if cudd.Cudd_ShuffleHeap(dd, sequence) == 0 || cudd.Cudd_zddVarsFromBddVars(dd, 2) == 0 {
					t.Fatal("wide order failed")
				}
				positions := []int{0, n / 2, n - 1}
				f := dd.LogicZero
				for mask := 0; mask < 8; mask++ {
					if table>>mask&1 != 0 {
						cube := dd.One
						for i, index := range positions {
							cube = cudd.Cudd_bddAnd(dd, cube, cudd.Cudd_NotCond(dd.Vars[index], mask>>i&1 == 0))
						}
						f = cudd.Cudd_bddOr(dd, f, cube)
					}
				}
				cudd.Cudd_Ref(f)
				for phase := 0; phase < 8; phase++ {
					g := Gia_ManStart(int32(2*n + 64))
					Gia_ManHashStart(g)
					leaves := Vec_IntAlloc(n)
					for i := 0; i < n; i++ {
						lit := Gia_ManAppendCi(g)
						for j, index := range positions {
							if i == index {
								lit ^= int32(phase >> j & 1)
							}
						}
						Vec_IntPush(leaves, lit)
					}
					sop := Vec_StrAlloc(0)
					root := Gia_ManRebuildIsop(dd, f, g, leaves, sop, Vec_StrAlloc(0))
					if root < 0 {
						t.Fatalf("wide ISOP failed n%d table%x", n, table)
					}
					cover := Vec_StrArray(sop)
					fmt.Fprintf(&got, "case %d %x %d %d %d sop%s nodes", n, table, order, phase, root, hex.EncodeToString(cover[:len(cover)-1]))
					for _, obj := range g.pObjs {
						if Gia_ObjIsAnd(obj) != 0 {
							a, b, _, _ := g.Gate(obj.id)
							fmt.Fprintf(&got, " %d:%d", a, b)
						}
					}
					Gia_ManAppendCo(g, root)
					packed := 0
					for mask := 0; mask < 8; mask++ {
						inputs := make([]bool, n)
						for i, index := range positions {
							inputs[index] = mask>>i&1 != 0
						}
						want := table>>(mask^phase)&1 != 0
						checkQuantificationOutputs(t, g, inputs, []bool{want})
						if want {
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
				t.Fatalf("wide ISOP source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("native output length Go%d C%d", got.Len(), len(native))
	}
	t.Logf("%d exact native wide ISOP/factoring cases across 32/64-bit literal boundaries", cases)
}
