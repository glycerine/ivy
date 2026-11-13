package abci

import (
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func evalMiniForTest(p *miniaig.Mini_Aig_t, inputs []bool) []bool {
	values := make([]bool, miniaig.Mini_AigNodeNum(p))
	pi := 0
	var outs []bool
	eval := func(lit int) bool { return values[lit>>1] != (lit&1 != 0) }
	for i := 1; i < len(values); i++ {
		switch {
		case miniaig.Mini_AigNodeIsPi(p, i):
			values[i] = inputs[pi]
			pi++
		case miniaig.Mini_AigNodeIsAnd(p, i):
			values[i] = eval(miniaig.Mini_AigNodeFanin0(p, i)) && eval(miniaig.Mini_AigNodeFanin1(p, i))
		case miniaig.Mini_AigNodeIsPo(p, i):
			values[i] = eval(miniaig.Mini_AigNodeFanin0(p, i))
			outs = append(outs, values[i])
		}
	}
	return outs
}

func TestMiniAigFaninCopyCombinesComplements(t *testing.T) {
	p := miniaig.Mini_AigStart()
	a, b := miniaig.Mini_AigCreatePi(p), miniaig.Mini_AigCreatePi(p)
	gate := miniaig.Mini_AigAnd(p, a^1, b)
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	x, y := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	copies := vec.Vec_IntAllocArrayCopy([]int32{1, int32(x.ID<<1 | 1), int32(y.ID << 1)}, 3)
	if got := Abc_NodeFanin0Copy(n, copies, p, int32(gate>>1)); got != x.Signal() {
		t.Fatalf("fanin0 = %+v, want uncomplemented x", got)
	}
	if got := Abc_NodeFanin1Copy(n, copies, p, int32(gate>>1)); got != y.Signal() {
		t.Fatalf("fanin1 = %+v, want y", got)
	}
}
func TestMiniAigSourceConversionRoundTrip(t *testing.T) {
	p := miniaig.Mini_AigStart()
	a, b := miniaig.Mini_AigCreatePi(p), miniaig.Mini_AigCreatePi(p)
	out := miniaig.Mini_AigAnd(p, a, b^1)
	miniaig.Mini_AigCreatePo(p, out)
	miniaig.Mini_AigCreatePo(p, 1)
	n := Abc_NtkFromMiniAig(p)
	q := Abc_NtkToMiniAig(n)
	if miniaig.Mini_AigPiNum(q) != 2 || miniaig.Mini_AigPoNum(q) != 2 {
		t.Fatalf("roundtrip interface PI=%d PO=%d", miniaig.Mini_AigPiNum(q), miniaig.Mini_AigPoNum(q))
	}
	for mask := 0; mask < 4; mask++ {
		values := []bool{mask&1 != 0, mask&2 != 0}
		want := evalMiniForTest(p, values)
		got := evalMiniForTest(q, values)
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("mask %d output %d got %v want %v", mask, i, got, want)
			}
		}
	}
}
func TestMiniAigSourceSequentialInterface(t *testing.T) {
	p := miniaig.Mini_AigStart()
	input, state := miniaig.Mini_AigCreatePi(p), miniaig.Mini_AigCreatePi(p)
	miniaig.Mini_AigCreatePo(p, state)
	miniaig.Mini_AigCreatePo(p, input)
	miniaig.Mini_AigSetRegNum(p, 1)
	n := Abc_NtkFromMiniAig(p)
	if abc.Abc_NtkLatchNum(n) != 1 || len(n.Inputs()) != 2 || len(n.Outputs()) != 2 {
		t.Fatalf("interface inputs=%d outputs=%d latches=%d", len(n.Inputs()), len(n.Outputs()), abc.Abc_NtkLatchNum(n))
	}
	q := Abc_NtkToMiniAig(n)
	if miniaig.Mini_AigRegNum(q) != 1 {
		t.Fatal("register count not preserved")
	}
	for mask := 0; mask < 4; mask++ {
		values := []bool{mask&1 != 0, mask&2 != 0}
		a, b := evalMiniForTest(p, values), evalMiniForTest(q, values)
		for i := range a {
			if a[i] != b[i] {
				t.Fatalf("mask %d got %v want %v", mask, b, a)
			}
		}
	}
}
