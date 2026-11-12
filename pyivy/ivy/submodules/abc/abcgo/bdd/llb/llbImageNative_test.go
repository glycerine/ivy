package llb

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestImageHelperQuantifiesAndRenamesLatchVariables(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state)
	g.AddOutput(state.Not())
	aig.Aig_ManSetRegNum(g, 1)
	dd := core.NewBDDManager(int(aig.Aig_ManObjNumMax(g)))
	defer core.Cudd_Quit(dd)
	initial := core.Cudd_Not(core.Cudd_bddIthVar(dd, state.NodeID()))
	image := Llb_ImgComputeImage(g, nil, dd, initial, nil, nil, nil, 0, 0, 0, 0)
	if image == nil {
		t.Fatal("image computation failed")
	}
	for _, test := range []struct{ value, want bool }{{false, false}, {true, true}} {
		assignment := make([]bool, len(dd.Vars))
		assignment[state.NodeID()] = test.value
		got, err := dd.Evaluate(image, assignment)
		if err != nil || got != test.want {
			t.Fatalf("image(%t)=%t, %v; want %t", test.value, got, err, test.want)
		}
	}
}

func TestImageScheduleSeparatesSingleAndSharedVariables(t *testing.T) {
	supports := vec.Vec_PtrAlloc(3)
	for _, values := range [][]int32{{1, 1, 0}, {0, 1, 1}, {0, 0, 1}} {
		v := vec.Vec_IntAlloc(len(values))
		for _, value := range values {
			vec.Vec_IntPush(v, value)
		}
		vec.Vec_PtrPush(supports, v)
	}
	var q0, q1 *Vec_Ptr_t
	Llb_ImgSchedule(supports, &q0, &q1, 0)
	if got := vec.Vec_IntArray(vec.Vec_PtrEntry(q0, 0).(*Vec_Int_t)); len(got) != 1 || got[0] != 0 {
		t.Fatalf("unique support schedule = %v", got)
	}
	if got := vec.Vec_IntArray(vec.Vec_PtrEntry(q1, 2).(*Vec_Int_t)); len(got) != 1 || got[0] != 2 {
		t.Fatalf("shared support schedule = %v", got)
	}
}

func TestPartitionBuilderHonorsVariableOrder(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(x)
	order := vec.Vec_IntStartFull(int(aig.Aig_ManObjNumMax(g)))
	for id := 0; id < vec.Vec_IntSize(order); id++ {
		vec.Vec_IntWriteEntry(order, id, int32(id))
	}
	vec.Vec_IntWriteEntry(order, int(x.NodeID()), 1)
	vec.Vec_IntWriteEntry(order, int(y.NodeID()), 0)
	dd := core.NewBDDManager(2)
	defer core.Cudd_Quit(dd)
	parts := Llb_Nonlin4FindPartitions(dd, g, order, 1)
	if vec.Vec_PtrSize(parts) != 1 {
		t.Fatalf("partition count = %d, want 1", vec.Vec_PtrSize(parts))
	}
	if got := vec.Vec_PtrEntry(parts, 0); got != core.Cudd_bddIthVar(dd, 1) {
		t.Fatal("partition output must use the requested variable index")
	}
}
