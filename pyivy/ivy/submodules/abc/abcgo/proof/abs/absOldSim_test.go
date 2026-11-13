package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestOldSimulationAndSensitization(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, y))
	cex := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(cex.PData, 0)
	util.Abc_InfoSetBit(cex.PData, 1)
	info := vec.Vec_PtrAllocSimInfo(len(g.Nodes()), 1)
	if got := Saig_ManSimDataInit(g, cex, info, nil); got != SAIG_ONE {
		t.Fatalf("concrete output %d", got)
	}
	unknown := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	if got := Saig_ManSimDataInit(g, cex, info, unknown); got != SAIG_UND {
		t.Fatalf("unknown-input output %d", got)
	}
	selected := Saig_ManExtendCounterExampleTest2(g, 1, cex, 0)
	if got := vec.Vec_IntArray(selected); !reflect.DeepEqual(got, []int32{1}) {
		t.Fatalf("selected pseudo-inputs %v", got)
	}
}

func TestOldSimulationPropagatesLatchAcrossFrames(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	cex := util.Abc_CexAlloc(1, 1, 2)
	cex.IFrame = 1
	util.Abc_InfoSetBit(cex.PData, 1) // input in frame zero becomes the next state
	info := vec.Vec_PtrAllocSimInfo(len(g.Nodes()), 1)
	if got := Saig_ManSimDataInit(g, cex, info, nil); got != SAIG_ONE {
		t.Fatalf("second-frame latch output %d", got)
	}
	if got := Saig_ManSimDataInit2(g, cex, info); got != SAIG_ONE_NEW {
		t.Fatalf("visited encoding second-frame output %d", got)
	}
	selected := Saig_ManExtendCounterExampleTest2(g, 0, cex, 0)
	if got := vec.Vec_IntArray(selected); !reflect.DeepEqual(got, []int32{0}) {
		t.Fatalf("selected sequential input %v", got)
	}
}

func TestOldSensitizationPreservesAllTwoInputFunctions(t *testing.T) {
	for table := 0; table < 16; table++ {
		for assignment := 0; assignment < 4; assignment++ {
			if table&(1<<assignment) == 0 {
				continue
			}
			g := aig.NewGraph()
			x, y := g.NewInput(), g.NewInput()
			root := g.Const0()
			for bits := 0; bits < 4; bits++ {
				if table&(1<<bits) == 0 {
					continue
				}
				a, b := x, y
				if bits&1 == 0 {
					a = a.Not()
				}
				if bits&2 == 0 {
					b = b.Not()
				}
				root = g.OR(root, g.AND(a, b))
			}
			g.AddOutput(root)
			cex := util.Abc_CexAlloc(0, 2, 1)
			if assignment&1 != 0 {
				util.Abc_InfoSetBit(cex.PData, 0)
			}
			if assignment&2 != 0 {
				util.Abc_InfoSetBit(cex.PData, 1)
			}
			selected := Saig_ManExtendCounterExampleTest2(g, 1, cex, 0)
			if selected == nil {
				t.Fatalf("table=%04b assignment=%02b: no result", table, assignment)
			}
			if vec.Vec_IntSize(selected) == 0 {
				other := assignment ^ 2
				if table&(1<<other) == 0 {
					t.Fatalf("table=%04b assignment=%02b omitted a required input", table, assignment)
				}
			} else if got := vec.Vec_IntArray(selected); !reflect.DeepEqual(got, []int32{1}) {
				t.Fatalf("table=%04b assignment=%02b selected %v", table, assignment, got)
			}
		}
	}
}
