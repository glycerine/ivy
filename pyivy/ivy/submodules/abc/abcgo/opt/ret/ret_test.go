package ret

import (
	"github.com/glycerine/abcgo/base/abc"
	"testing"
)

func retFixture(forward bool) *abc.Network {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	n.Name = "retiming"
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	latch := func(driver *abc.Object, one bool) *abc.Object {
		in, l, out := abc.Abc_NtkCreateBi(n), abc.Abc_NtkCreateLatch(n), abc.Abc_NtkCreateBo(n)
		retAdd(in, driver.Signal())
		retAdd(l, in.Signal())
		retAdd(out, l.Signal())
		if one {
			abc.Abc_LatchSetInit1(l)
		} else {
			abc.Abc_LatchSetInit0(l)
		}
		return out
	}
	var root *abc.Object
	if forward {
		x, y := latch(a, true), latch(b, false)
		root = abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{x, y})
	} else {
		root = latch(abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{a, b}), true)
	}
	retAdd(abc.Abc_NtkCreatePo(n), root.Signal())
	abc.Abc_NtkOrderCisCos(n)
	return n
}
func TestDirectLatchMoves(t *testing.T) {
	for _, forward := range []int32{0, 1} {
		n := retFixture(forward != 0)
		original := abc.Abc_NtkDup(n)
		bindings := Abc_NtkRetimePrepareLatches(n)
		if forward != 0 {
			Abc_NtkRetimeTranferToCopy(n)
		}
		values := Abc_NtkRetimeCollectLatchValues(n)
		var cone *abc.Network
		if forward == 0 {
			cone = Abc_NtkRetimeBackwardInitialStart(n)
		}
		node := retNodes(n)[0]
		if Abc_NtkRetimeNodeIsEnabled(node, forward) == 0 {
			t.Fatal("node disabled")
		}
		Abc_NtkRetimeNode(node, forward, 1)
		if forward != 0 {
			Abc_NtkRetimeTranferFromCopy(n)
		} else {
			Abc_NtkRetimeBackwardInitialFinish(n, cone, values, 0)
		}
		Abc_NtkRetimeFinalizeLatches(n, bindings, 0, 0)
		if err := n.Check(); err != nil {
			t.Fatal(err)
		}
		if !retEquivalent(original, n) {
			t.Fatalf("forward%d changed sequential behavior", forward)
		}
		want := 2
		if forward != 0 {
			want = 1
		}
		if len(retLatches(n)) != want {
			t.Fatal("latch movement count")
		}
	}
}
func TestAllRetimingModes(t *testing.T) {
	for mode := int32(1); mode <= 6; mode++ {
		for _, forward := range []bool{false, true} {
			n := retFixture(forward)
			original := abc.Abc_NtkDup(n)
			Abc_NtkRetime(n, mode, 0, 0, 0, 0, 1, 0)
			if err := n.Check(); err != nil {
				t.Fatalf("mode%d: %v", mode, err)
			}
			if !retEquivalent(original, n) {
				t.Fatalf("mode%d changed sequential behavior", mode)
			}
		}
	}
}
func TestInitialStateSAT(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	and := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{a, b})
	retAdd(abc.Abc_NtkCreatePo(n), and.Signal())
	for _, target := range []int32{0, 1} {
		model := Abc_NtkRetimeInitialValues(n, []int32{target}, 0)
		if model == nil || Abc_NtkRetimeVerifyModel(n, []int32{target}, model) == 0 {
			t.Fatal("SAT model mismatch")
		}
	}
	retAdd(abc.Abc_NtkCreatePo(n), and.Signal())
	if Abc_NtkRetimeInitialValues(n, []int32{0, 1}, 0) != nil {
		t.Fatal("inconsistent target accepted")
	}
}
func TestSopSimulationExhaustive(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	o := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{a, b})
	for _, sop := range []string{"11 1\n", "1- 1\n-1 1\n", "01 1\n10 1\n", "11 0\n", "-- 1\n", "-- 0\n"} {
		o.Data = sop
		for x := 0; x < 4; x++ {
			a.Temp = x & 1
			b.Temp = (x >> 1) & 1
			got := Abc_ObjSopSimulate(o)
			want, err := n.Evaluate(o.Signal(), map[*abc.Object]bool{a: a.Temp != 0, b: b.Temp != 0})
			if err != nil || got != retBool(want) {
				t.Fatal("SOP simulation mismatch", sop, x)
			}
		}
	}
}
func TestLagMath(t *testing.T) {
	for period := int32(1); period < 10; period++ {
		for value := int32(-100); value <= 100; value++ {
			lag := Abc_NodeComputeLag(value, period)
			if !(lag*period < value && value <= (lag+1)*period) {
				t.Fatalf("lag%d value%d period%d", lag, value, period)
			}
		}
	}
}

func TestSourceRandomSequence(t *testing.T) {
	r := newRetRandom(1)
	for _, want := range []uint32{1804289383, 846930886, 1681692777, 1714636915, 1957747793} {
		if got := r.next(); got != want {
			t.Fatalf("random %d/%d", got, want)
		}
	}
}

func TestInitialModelRejectsMismatch(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	pi := abc.Abc_NtkCreatePi(n)
	retAdd(abc.Abc_NtkCreatePo(n), pi.Signal())
	if Abc_NtkRetimeVerifyModel(n, []int32{1}, []int32{0}) != 0 {
		t.Fatal("incorrect initial-state model accepted")
	}
}

func TestSequentialFeedbackRetiming(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	pi := abc.Abc_NtkCreatePi(n)
	in, l, out := abc.Abc_NtkCreateBi(n), abc.Abc_NtkCreateLatch(n), abc.Abc_NtkCreateBo(n)
	retAdd(l, in.Signal())
	retAdd(out, l.Signal())
	and := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{pi, out})
	retAdd(in, and.Signal())
	retAdd(abc.Abc_NtkCreatePo(n), out.Signal())
	abc.Abc_LatchSetInit1(l)
	abc.Abc_NtkOrderCisCos(n)
	original := abc.Abc_NtkDup(n)
	Abc_NtkRetime(n, 2, 0, 0, 0, 0, 1, 0)
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
	if !retEquivalent(original, n) {
		t.Fatal("feedback retiming changed behavior")
	}
}

func TestSharingRetainsLatchInputPolarity(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	pi := abc.Abc_NtkCreatePi(n)
	a, b := abc.Abc_NtkCreateLatch(n), abc.Abc_NtkCreateLatch(n)
	retAdd(a, pi.Signal())
	retAdd(b, pi.Signal().Not())
	abc.Abc_LatchSetInit0(a)
	abc.Abc_LatchSetInit0(b)
	Abc_NtkRetimeShareLatches(n, 0)
	if len(retLatches(n)) != 2 {
		t.Fatal("opposite latch inputs merged")
	}
}
