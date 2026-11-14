package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func newStrashForSenseTest() *abc.Network {
	return abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
}

func TestXsimFaninEdgePhases(t *testing.T) {
	n := newStrashForSenseTest()
	left := abc.Abc_NtkCreatePi(n)
	right := abc.Abc_NtkCreatePi(n)
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	and := abc.Abc_AigAnd(manager, left.Signal().Not(), right.Signal())
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, and.Not())

	Abc_ObjSetXsim(left, XVS0)
	Abc_ObjSetXsim(right, XVS1)
	if got := Abc_ObjGetXsimFanin0(and.Object); got != XVS1 {
		t.Fatalf("complemented first edge value = %d, want one", got)
	}
	if got := Abc_ObjGetXsimFanin1(and.Object); got != XVS1 {
		t.Fatalf("second edge value = %d, want one", got)
	}
	Abc_ObjSetXsim(and.Object, Abc_XsimAnd(Abc_ObjGetXsimFanin0(and.Object), Abc_ObjGetXsimFanin1(and.Object)))
	if got := Abc_ObjGetXsimFanin0(po); got != XVS0 {
		t.Fatalf("complemented output edge value = %d, want zero", got)
	}
}

func TestXsimUnknownInputsAndCycleInitState(t *testing.T) {
	n := newStrashForSenseTest()
	pi := abc.Abc_NtkCreatePi(n)
	bo := abc.Abc_NtkAddLatch(n, pi, abc.ABC_INIT_ZERO)
	latch := abc.Abc_ObjFanin0(bo)
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, bo)
	abc.Abc_NtkOrderCisCos(n)

	Abc_NtkXValueSimulate(n, 2, 1, 1, 0)
	if got := Abc_ObjGetXsim(bo); got != XVSX {
		t.Fatalf("latch output after unknown-state simulation = %d, want X", got)
	}
	if got := Abc_ObjGetXsim(po); got != XVSX {
		t.Fatalf("PO after unknown-state simulation = %d, want X", got)
	}

	Abc_NtkCycleInitState(n, 2, 1, 0)
	if abc.Abc_LatchIsInitDc(latch) == 0 {
		t.Fatalf("cycled latch init = %#v, want X", latch.Data)
	}
}

func TestSensitivityMiterKeepsEdgePhasesAndConstraint(t *testing.T) {
	n := newStrashForSenseTest()
	a := abc.Abc_NtkCreatePi(n)
	b := abc.Abc_NtkCreatePi(n)
	c := abc.Abc_NtkCreatePi(n)
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	function := abc.Abc_AigAnd(manager, a.Signal(), c.Signal())
	constraint := abc.Abc_AigOr(manager, a.Signal(), b.Signal())
	out := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(out, function)
	con := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(con, constraint)
	abc.Abc_NtkOrderCisCos(n)

	miter := Abc_NtkSensitivityMiter(n, 0)
	if miter == nil {
		t.Fatal("sensitivity miter construction failed")
	}
	miterOut := abc.Abc_NtkPo(miter, 0)
	root := abc.Abc_ObjChild0(miterOut)
	inputs := miter.ObjectsOfKind(abc.ObjectPI)
	for _, tc := range []struct {
		b, c, want bool
	}{
		{b: true, c: true, want: true},
		{b: false, c: true, want: false},
		{b: true, c: false, want: false},
	} {
		got, err := miter.Evaluate(root, map[*abc.Object]bool{inputs[1]: tc.b, inputs[2]: tc.c})
		if err != nil {
			t.Fatalf("evaluate miter: %v", err)
		}
		if got != tc.want {
			t.Errorf("miter(b=%v,c=%v) = %v, want %v", tc.b, tc.c, got, tc.want)
		}
	}

	constraintOnly := Abc_NtkSensitivityMiter(n, 1)
	if constraintOnly == nil {
		t.Fatal("constraint sensitivity miter construction failed")
	}
	constraintOut := abc.Abc_ObjChild0(abc.Abc_NtkPo(constraintOnly, 0))
	constraintInputs := constraintOnly.ObjectsOfKind(abc.ObjectPI)
	for _, av := range []bool{false, true} {
		for _, cv := range []bool{false, true} {
			got, err := constraintOnly.Evaluate(constraintOut, map[*abc.Object]bool{constraintInputs[0]: av, constraintInputs[2]: cv})
			if err != nil {
				t.Fatalf("evaluate constraint-only miter: %v", err)
			}
			if got {
				t.Errorf("input b was reported sensitive for a=%v,c=%v", av, cv)
			}
		}
	}
}
