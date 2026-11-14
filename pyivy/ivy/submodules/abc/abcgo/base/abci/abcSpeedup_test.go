package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestDelayTraceSortPinsOrdersDecreasingArrival(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	inputs := []*abc.Object{abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)}
	root := abc.Abc_NtkCreateNode(n)
	for _, input := range inputs {
		abc.Abc_ObjAddFanin(root, input)
	}
	abc.Abc_NtkSetLutTimes(n, make([]float32, 3*abc.Abc_NtkObjNumMax(n)))
	Abc_ObjSetArrival(inputs[0], 1)
	Abc_ObjSetArrival(inputs[1], 3)
	Abc_ObjSetArrival(inputs[2], 2)
	var order [32]int32
	var delays [32]float32
	Abc_NtkDelayTraceSortPins(root, order[:], delays[:])
	if got := []int32{order[0], order[1], order[2]}; got[0] != 1 || got[1] != 2 || got[2] != 0 {
		t.Fatalf("pin order = %v, want [1 2 0]", got)
	}
}

func TestDelayTraceTimingScratchHasThreeSlotsPerObject(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a := abc.Abc_NtkCreatePi(n)
	abc.Abc_NtkSetLutTimes(n, make([]float32, 3*abc.Abc_NtkObjNumMax(n)))
	Abc_ObjSetArrival(a, 1.25)
	Abc_ObjSetRequired(a, 4.5)
	Abc_ObjSetSlack(a, 3.25)
	if Abc_ObjArrival(a) != 1.25 || Abc_ObjRequired(a) != 4.5 || Abc_ObjSlack(a) != 3.25 {
		t.Fatalf("timing tuple = (%g,%g,%g)", Abc_ObjArrival(a), Abc_ObjRequired(a), Abc_ObjSlack(a))
	}
}

func TestDelayTraceUnitDelayArrivalRequiredAndSlack(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	root := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(root, a)
	abc.Abc_ObjAddFanin(root, b)
	output := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(output, root)

	if got := Abc_NtkDelayTraceLut(n, 0); got != 1 {
		t.Fatalf("unit-delay arrival = %g, want 1", got)
	}
	if got := Abc_ObjArrival(root); got != 1 {
		t.Fatalf("root arrival = %g, want 1", got)
	}
	if got := Abc_ObjRequired(root); got != 1 {
		t.Fatalf("root required time = %g, want 1", got)
	}
	if got := Abc_ObjSlack(root); got != 0 {
		t.Fatalf("root slack = %g, want 0", got)
	}
	if got := Abc_ObjRequired(a); got != 0 {
		t.Fatalf("input required time = %g, want 0", got)
	}
}
